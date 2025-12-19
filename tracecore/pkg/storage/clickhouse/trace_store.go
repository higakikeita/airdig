package clickhouse

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/higakikeita/airdig/tracecore/pkg/models"
)

// TraceStore handles trace storage operations
type TraceStore struct {
	client *Client
}

// NewTraceStore creates a new trace store
func NewTraceStore(client *Client) *TraceStore {
	return &TraceStore{client: client}
}

// SaveSpans saves a batch of spans to ClickHouse
func (s *TraceStore) SaveSpans(ctx context.Context, spans []models.Span) error {
	if len(spans) == 0 {
		return nil
	}

	batch, err := s.client.PrepareBatch(ctx, `
		INSERT INTO traces (
			trace_id, span_id, parent_span_id, service_name, operation_name,
			start_time, end_time, duration_ns, status_code, resource_id, attributes
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, span := range spans {
		// Serialize attributes to JSON
		attrsJSON, err := json.Marshal(span.Attributes)
		if err != nil {
			attrsJSON = []byte("{}")
		}

		// Extract resource ID
		resourceID := span.GetResourceID()

		err = batch.Append(
			span.TraceID,
			span.SpanID,
			span.ParentSpanID,
			span.GetService(),
			span.OperationName,
			span.StartTime,
			span.EndTime,
			span.Duration.Nanoseconds(),
			string(span.StatusCode),
			resourceID,
			string(attrsJSON),
		)
		if err != nil {
			return fmt.Errorf("failed to append span: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	return nil
}

// GetTraceByID retrieves all spans for a given trace ID
func (s *TraceStore) GetTraceByID(ctx context.Context, traceID string) (*models.Trace, error) {
	query := `
		SELECT
			trace_id, span_id, parent_span_id, service_name, operation_name,
			start_time, end_time, duration_ns, status_code, resource_id, attributes
		FROM traces
		WHERE trace_id = ?
		ORDER BY start_time ASC
	`

	rows, err := s.client.Query(ctx, query, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query trace: %w", err)
	}
	defer rows.Close()

	var spans []models.Span
	for rows.Next() {
		var span models.Span
		var durationNs int64
		var statusCode string
		var attrsJSON string

		err := rows.Scan(
			&span.TraceID,
			&span.SpanID,
			&span.ParentSpanID,
			&span.ServiceName,
			&span.OperationName,
			&span.StartTime,
			&span.EndTime,
			&durationNs,
			&statusCode,
			&span.ResourceAttrs,
			&attrsJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan span: %w", err)
		}

		span.Duration = time.Duration(durationNs)
		span.StatusCode = models.SpanStatus(statusCode)

		// Deserialize attributes
		if attrsJSON != "" {
			if err := json.Unmarshal([]byte(attrsJSON), &span.Attributes); err != nil {
				span.Attributes = make(map[string]interface{})
			}
		}

		spans = append(spans, span)
	}

	if len(spans) == 0 {
		return nil, fmt.Errorf("trace not found: %s", traceID)
	}

	// Build trace
	trace := &models.Trace{
		TraceID:  traceID,
		Spans:    spans,
		SpanCount: len(spans),
	}

	// Calculate trace metrics
	if len(spans) > 0 {
		trace.StartTime = spans[0].StartTime
		trace.EndTime = spans[0].EndTime
		
		servicesMap := make(map[string]bool)
		hasError := false

		for _, span := range spans {
			if span.StartTime.Before(trace.StartTime) {
				trace.StartTime = span.StartTime
			}
			if span.EndTime.After(trace.EndTime) {
				trace.EndTime = span.EndTime
			}
			servicesMap[span.GetService()] = true
			if span.IsError() {
				hasError = true
			}
		}

		trace.Duration = trace.EndTime.Sub(trace.StartTime)
		trace.HasError = hasError
		
		for svc := range servicesMap {
			trace.Services = append(trace.Services, svc)
		}
	}

	return trace, nil
}

// ListTraces retrieves traces with filters
func (s *TraceStore) ListTraces(ctx context.Context, filters TraceFilters) ([]models.TraceSummary, error) {
	query := `
		SELECT
			trace_id,
			any(service_name) as root_service,
			any(operation_name) as root_operation,
			min(start_time) as start_time,
			max(end_time) - min(start_time) as duration,
			count() as span_count,
			countIf(status_code = 'error') as error_count,
			groupUniqArray(service_name) as services
		FROM traces
		WHERE 1=1
	`

	args := []interface{}{}

	if !filters.StartTime.IsZero() {
		query += " AND start_time >= ?"
		args = append(args, filters.StartTime)
	}

	if !filters.EndTime.IsZero() {
		query += " AND start_time <= ?"
		args = append(args, filters.EndTime)
	}

	if filters.ServiceName != "" {
		query += " AND service_name = ?"
		args = append(args, filters.ServiceName)
	}

	if filters.MinDuration > 0 {
		query += " AND duration_ns >= ?"
		args = append(args, filters.MinDuration.Nanoseconds())
	}

	if filters.HasError {
		query += " AND status_code = 'error'"
	}

	query += `
		GROUP BY trace_id
		ORDER BY start_time DESC
	`

	if filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	} else {
		query += " LIMIT 100"
	}

	rows, err := s.client.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query traces: %w", err)
	}
	defer rows.Close()

	var traces []models.TraceSummary
	for rows.Next() {
		var trace models.TraceSummary
		var durationNs int64
		
		err := rows.Scan(
			&trace.TraceID,
			&trace.RootService,
			&trace.RootOp,
			&trace.StartTime,
			&durationNs,
			&trace.SpanCount,
			&trace.ErrorCount,
			&trace.Services,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trace: %w", err)
		}

		trace.Duration = time.Duration(durationNs)
		traces = append(traces, trace)
	}

	return traces, nil
}

// GetServiceStats retrieves statistics for a service
func (s *TraceStore) GetServiceStats(ctx context.Context, serviceName string, days int) (*ServiceStats, error) {
	query := `
		SELECT
			count() as request_count,
			countIf(status_code = 'error') as error_count,
			avg(duration_ns) as avg_duration,
			quantile(0.95)(duration_ns) as p95_duration,
			quantile(0.99)(duration_ns) as p99_duration
		FROM traces
		WHERE service_name = ?
		  AND date >= today() - INTERVAL ? DAY
	`

	var stats ServiceStats
	var avgDurationNs, p95DurationNs, p99DurationNs float64

	err := s.client.QueryRow(ctx, query, serviceName, days).Scan(
		&stats.RequestCount,
		&stats.ErrorCount,
		&avgDurationNs,
		&p95DurationNs,
		&p99DurationNs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query service stats: %w", err)
	}

	stats.ServiceName = serviceName
	stats.AvgDuration = time.Duration(avgDurationNs)
	stats.P95Duration = time.Duration(p95DurationNs)
	stats.P99Duration = time.Duration(p99DurationNs)
	
	if stats.RequestCount > 0 {
		stats.ErrorRate = float64(stats.ErrorCount) / float64(stats.RequestCount) * 100
	}

	return &stats, nil
}

// TraceFilters holds filters for trace queries
type TraceFilters struct {
	StartTime   time.Time
	EndTime     time.Time
	ServiceName string
	MinDuration time.Duration
	HasError    bool
	Limit       int
}

// ServiceStats holds statistics for a service
type ServiceStats struct {
	ServiceName  string
	RequestCount uint64
	ErrorCount   uint64
	ErrorRate    float64
	AvgDuration  time.Duration
	P95Duration  time.Duration
	P99Duration  time.Duration
}
