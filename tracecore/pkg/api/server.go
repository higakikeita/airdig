package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/higakikeita/airdig/tracecore/pkg/storage/clickhouse"
)

// Config holds API server configuration
type Config struct {
	Host string
	Port int
}

// DefaultConfig returns default API server configuration
func DefaultConfig() *Config {
	return &Config{
		Host: "localhost",
		Port: 8082,
	}
}

// Server is the HTTP API server
type Server struct {
	addr        string
	mux         *http.ServeMux
	server      *http.Server
	traceStore  *clickhouse.TraceStore
	chClient    *clickhouse.Client
	logger      Logger
}

// Logger interface
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

// NewServer creates a new API server
func NewServer(config *Config, traceStore *clickhouse.TraceStore, chClient *clickhouse.Client, logger Logger) *Server {
	if config == nil {
		config = DefaultConfig()
	}

	s := &Server{
		addr:       fmt.Sprintf("%s:%d", config.Host, config.Port),
		mux:        http.NewServeMux(),
		traceStore: traceStore,
		chClient:   chClient,
		logger:     logger,
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures HTTP routes
func (s *Server) setupRoutes() {
	// Health and status
	s.mux.HandleFunc("/health", s.handleHealth())
	s.mux.HandleFunc("/api/v1/status", s.handleStatus())

	// Trace endpoints
	s.mux.HandleFunc("/api/v1/traces", s.corsMiddleware(s.handleTraces()))
	s.mux.HandleFunc("/api/v1/traces/", s.corsMiddleware(s.handleTraceByID()))

	// Service endpoints
	s.mux.HandleFunc("/api/v1/services", s.corsMiddleware(s.handleServices()))
	s.mux.HandleFunc("/api/v1/services/", s.corsMiddleware(s.handleServiceStats()))

	// Service map endpoints (placeholder for Phase 2)
	s.mux.HandleFunc("/api/v1/servicemap", s.corsMiddleware(s.handleServiceMap()))
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	s.server = &http.Server{
		Addr:    s.addr,
		Handler: s.mux,
	}

	s.logger.Info("Starting API server", "addr", s.addr)

	go func() {
		if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
			s.logger.Error("Server error", "error", err)
		}
	}()

	return nil
}

// Stop stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// Health check handler
func (s *Server) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		health := map[string]interface{}{
			"status": "ok",
			"timestamp": time.Now().Unix(),
		}

		// Check ClickHouse if available
		if s.chClient != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()

			if status, err := s.chClient.HealthCheck(ctx); err == nil {
				health["clickhouse"] = status
			} else {
				health["clickhouse"] = map[string]interface{}{
					"connected": false,
					"error": err.Error(),
				}
			}
		}

		respondJSON(w, http.StatusOK, health)
	}
}

// Status handler
func (s *Server) handleStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := map[string]interface{}{
			"service": "tracecore",
			"version": "0.1.0",
			"timestamp": time.Now().Unix(),
		}

		respondJSON(w, http.StatusOK, status)
	}
}

// Traces list handler
func (s *Server) handleTraces() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		
		// Parse query parameters
		filters := clickhouse.TraceFilters{
			ServiceName: r.URL.Query().Get("service"),
			Limit:       parseQueryInt(r, "limit", 100),
		}

		if start := r.URL.Query().Get("start"); start != "" {
			if t, err := time.Parse(time.RFC3339, start); err == nil {
				filters.StartTime = t
			}
		}

		if end := r.URL.Query().Get("end"); end != "" {
			if t, err := time.Parse(time.RFC3339, end); err == nil {
				filters.EndTime = t
			}
		}

		traces, err := s.traceStore.ListTraces(ctx, filters)
		if err != nil {
			s.logger.Error("Failed to list traces", "error", err)
			respondJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Failed to list traces",
			})
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"traces": traces,
			"count":  len(traces),
		})
	}
}

// Trace by ID handler
func (s *Server) handleTraceByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract trace ID from path
		traceID := r.URL.Path[len("/api/v1/traces/"):]
		if traceID == "" {
			http.Error(w, "Trace ID required", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		trace, err := s.traceStore.GetTraceByID(ctx, traceID)
		if err != nil {
			s.logger.Error("Failed to get trace", "trace_id", traceID, "error", err)
			respondJSON(w, http.StatusNotFound, map[string]string{
				"error": "Trace not found",
			})
			return
		}

		respondJSON(w, http.StatusOK, trace)
	}
}

// Services list handler (placeholder)
func (s *Server) handleServices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// TODO: Implement service list query
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"services": []string{},
			"message": "Service list will be implemented in Phase 2",
		})
	}
}

// Service stats handler
func (s *Server) handleServiceStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		serviceName := r.URL.Path[len("/api/v1/services/"):]
		if serviceName == "" {
			http.Error(w, "Service name required", http.StatusBadRequest)
			return
		}

		days := parseQueryInt(r, "days", 7)

		ctx := r.Context()
		stats, err := s.traceStore.GetServiceStats(ctx, serviceName, days)
		if err != nil {
			s.logger.Error("Failed to get service stats", "service", serviceName, "error", err)
			respondJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Failed to get service stats",
			})
			return
		}

		respondJSON(w, http.StatusOK, stats)
	}
}

// Service map handler (placeholder for Phase 2)
func (s *Server) handleServiceMap() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// TODO: Implement service map generation in Phase 2
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"nodes": []interface{}{},
			"edges": []interface{}{},
			"message": "Service map will be implemented in Phase 2",
		})
	}
}

// CORS middleware
func (s *Server) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// Helper functions

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func parseQueryInt(r *http.Request, key string, defaultValue int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultValue
	}

	var result int
	if _, err := fmt.Sscanf(val, "%d", &result); err != nil {
		return defaultValue
	}

	return result
}
