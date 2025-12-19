package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/higakikeita/airdig/deepdrift/pkg/storage/clickhouse"
)

// Server represents the REST API server
type Server struct {
	addr        string
	driftStore  *clickhouse.DriftStore
	impactStore *clickhouse.ImpactStore
	chClient    *clickhouse.Client
	mux         *http.ServeMux
	server      *http.Server
}

// Config holds server configuration
type Config struct {
	Host            string
	Port            int
	ClickHouseAddr  string
	ClickHouseDB    string
	EnableCORS      bool
	AllowedOrigins  []string
}

// DefaultConfig returns default server configuration
func DefaultConfig() *Config {
	return &Config{
		Host:           "0.0.0.0",
		Port:           8080,
		ClickHouseAddr: "localhost:9000",
		ClickHouseDB:   "deepdrift",
		EnableCORS:     true,
		AllowedOrigins: []string{"*"},
	}
}

// NewServer creates a new API server
func NewServer(config *Config, chClient *clickhouse.Client) *Server {
	if config == nil {
		config = DefaultConfig()
	}

	// Only create stores if ClickHouse client is provided
	var driftStore *clickhouse.DriftStore
	var impactStore *clickhouse.ImpactStore
	if chClient != nil {
		driftStore = clickhouse.NewDriftStore(chClient)
		impactStore = clickhouse.NewImpactStore(chClient)
	}

	s := &Server{
		addr:        fmt.Sprintf("%s:%d", config.Host, config.Port),
		driftStore:  driftStore,
		impactStore: impactStore,
		chClient:    chClient,
		mux:         http.NewServeMux(),
	}

	// Register routes
	s.registerRoutes(config)

	s.server = &http.Server{
		Addr:         s.addr,
		Handler:      s.mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// registerRoutes registers all API routes
func (s *Server) registerRoutes(config *Config) {
	// Health check
	s.mux.HandleFunc("/health", s.handleHealth())
	s.mux.HandleFunc("/api/v1/health", s.handleHealth())

	// Drift events
	s.mux.HandleFunc("/api/v1/drifts", s.handleDrifts())
	s.mux.HandleFunc("/api/v1/drifts/", s.handleDriftByID())
	s.mux.HandleFunc("/api/v1/drifts/stats", s.handleDriftStats())

	// Impact analysis
	s.mux.HandleFunc("/api/v1/impact", s.handleImpactAnalysis())
	s.mux.HandleFunc("/api/v1/impact/", s.handleImpactByDriftID())
	s.mux.HandleFunc("/api/v1/impact/stats", s.handleImpactStats())
	s.mux.HandleFunc("/api/v1/impact/high", s.handleHighImpactDrifts())

	// Resource graph
	s.mux.HandleFunc("/api/v1/graph", s.handleGraph())
	s.mux.HandleFunc("/api/v1/graph/intended", s.handleIntendedGraph())

	// Resources
	s.mux.HandleFunc("/api/v1/resources", s.handleResources())

	// Static files (for React UI)
	fs := http.FileServer(http.Dir("./ui/dist"))
	s.mux.Handle("/ui/", http.StripPrefix("/ui/", fs))

	// Apply CORS middleware if enabled
	if config.EnableCORS {
		s.applyCORSMiddleware(config.AllowedOrigins)
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Printf("Starting DeepDrift API server on %s", s.addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down DeepDrift API server...")
	return s.server.Shutdown(ctx)
}

// applyCORSMiddleware applies CORS middleware to all routes
func (s *Server) applyCORSMiddleware(allowedOrigins []string) {
	originalMux := s.mux
	s.mux = http.NewServeMux()

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			for _, allowed := range allowedOrigins {
				if allowed == "*" || allowed == origin {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
					w.Header().Set("Access-Control-Max-Age", "3600")
					break
				}
			}
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		originalMux.ServeHTTP(w, r)
	})
}

// Response helpers

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// respondError sends an error response
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]interface{}{
		"error":   message,
		"status":  status,
		"timestamp": time.Now().Unix(),
	})
}

// parseQueryInt parses an integer query parameter
func parseQueryInt(r *http.Request, key string, defaultValue int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultValue
	}
	var result int
	fmt.Sscanf(val, "%d", &result)
	if result <= 0 {
		return defaultValue
	}
	return result
}

// parseQueryString parses a string query parameter
func parseQueryString(r *http.Request, key string, defaultValue string) string {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultValue
	}
	return val
}

// parseQueryTime parses a time query parameter (RFC3339 format)
func parseQueryTime(r *http.Request, key string) (time.Time, error) {
	val := r.URL.Query().Get(key)
	if val == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, val)
}
