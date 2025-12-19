package receiver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/higakikeita/airdig/tracecore/pkg/models"
	"github.com/higakikeita/airdig/tracecore/pkg/storage/clickhouse"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// Config holds OTLP receiver configuration
type Config struct {
	GRPCPort int
	HTTPPort int
}

// DefaultConfig returns default receiver configuration
func DefaultConfig() *Config {
	return &Config{
		GRPCPort: 4317,
		HTTPPort: 4318,
	}
}

// OTLPReceiver handles OTLP trace ingestion
type OTLPReceiver struct {
	config     *Config
	traceStore *clickhouse.TraceStore
	httpServer *http.Server
	logger     Logger
}

// Logger interface for logging
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

// NewOTLPReceiver creates a new OTLP receiver
func NewOTLPReceiver(config *Config, traceStore *clickhouse.TraceStore, logger Logger) *OTLPReceiver {
	if config == nil {
		config = DefaultConfig()
	}

	return &OTLPReceiver{
		config:     config,
		traceStore: traceStore,
		logger:     logger,
	}
}

// Start starts the OTLP receiver (HTTP only for Phase 1)
func (r *OTLPReceiver) Start(ctx context.Context) error {
	// Start HTTP server
	go func() {
		if err := r.startHTTP(ctx); err != nil {
			r.logger.Error("HTTP server error", "error", err)
		}
	}()

	r.logger.Info("OTLP receiver started", 
		"http_port", r.config.HTTPPort)

	return nil
}

// Stop stops the OTLP receiver
func (r *OTLPReceiver) Stop(ctx context.Context) error {
	if r.httpServer != nil {
		return r.httpServer.Shutdown(ctx)
	}

	return nil
}

// startHTTP starts the HTTP server
func (r *OTLPReceiver) startHTTP(ctx context.Context) error {
	mux := http.NewServeMux()

	// Register OTLP HTTP endpoint
	mux.HandleFunc("/v1/traces", r.handleHTTPTraces)

	r.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", r.config.HTTPPort),
		Handler: mux,
	}

	r.logger.Info("Starting HTTP server", "port", r.config.HTTPPort)

	if err := r.httpServer.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server error: %w", err)
	}

	return nil
}

// handleHTTPTraces handles HTTP trace requests
func (r *OTLPReceiver) handleHTTPTraces(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.logger.Error("Failed to read request body", "error", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Parse OTLP JSON (for Phase 1 simplicity, we'll support JSON format)
	unmarshaler := ptrace.JSONUnmarshaler{}
	traces, err := unmarshaler.UnmarshalTraces(body)
	if err != nil {
		r.logger.Error("Failed to unmarshal traces", "error", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Process traces
	if err := r.consumeTraces(req.Context(), traces); err != nil {
		r.logger.Error("Failed to consume traces", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
	})
}

// consumeTraces processes incoming traces
func (r *OTLPReceiver) consumeTraces(ctx context.Context, traces ptrace.Traces) error {
	// Convert OTLP traces to internal model
	spans, err := r.convertOTLPToSpans(traces)
	if err != nil {
		r.logger.Error("Failed to convert traces", "error", err)
		return err
	}

	// Store spans if store is available
	if r.traceStore != nil {
		if err := r.traceStore.SaveSpans(ctx, spans); err != nil {
			r.logger.Error("Failed to save spans", "error", err, "count", len(spans))
			return err
		}

		r.logger.Debug("Stored spans", "count", len(spans))
	} else {
		r.logger.Debug("Received spans (no storage configured)", "count", len(spans))
	}

	return nil
}

// convertOTLPToSpans converts OTLP traces to internal span model
func (r *OTLPReceiver) convertOTLPToSpans(traces ptrace.Traces) ([]models.Span, error) {
	var spans []models.Span

	resourceSpans := traces.ResourceSpans()
	for i := 0; i < resourceSpans.Len(); i++ {
		rs := resourceSpans.At(i)
		resource := rs.Resource()
		
		// Extract resource attributes
		resourceAttrs := make(map[string]string)
		resource.Attributes().Range(func(k string, v pcommon.Value) bool {
			resourceAttrs[k] = v.AsString()
			return true
		})

		scopeSpans := rs.ScopeSpans()
		for j := 0; j < scopeSpans.Len(); j++ {
			ss := scopeSpans.At(j)
			
			otlpSpans := ss.Spans()
			for k := 0; k < otlpSpans.Len(); k++ {
				otlpSpan := otlpSpans.At(k)
				
				// Convert to internal model
				span := models.Span{
					TraceID:       otlpSpan.TraceID().String(),
					SpanID:        otlpSpan.SpanID().String(),
					ParentSpanID:  otlpSpan.ParentSpanID().String(),
					OperationName: otlpSpan.Name(),
					StartTime:     otlpSpan.StartTimestamp().AsTime(),
					EndTime:       otlpSpan.EndTimestamp().AsTime(),
					Duration:      otlpSpan.EndTimestamp().AsTime().Sub(otlpSpan.StartTimestamp().AsTime()),
					ResourceAttrs: resourceAttrs,
					Attributes:    make(map[string]interface{}),
				}

				// Extract service name from resource attributes
				if svc, ok := resourceAttrs["service.name"]; ok {
					span.ServiceName = svc
				}

				// Convert status
				switch otlpSpan.Status().Code() {
				case ptrace.StatusCodeOk:
					span.StatusCode = models.SpanStatusOK
				case ptrace.StatusCodeError:
					span.StatusCode = models.SpanStatusError
				default:
					span.StatusCode = models.SpanStatusUnset
				}

				// Extract span attributes
				otlpSpan.Attributes().Range(func(k string, v pcommon.Value) bool {
					span.Attributes[k] = v.AsRaw()
					return true
				})

				spans = append(spans, span)
			}
		}
	}

	return spans, nil
}
