package models

import (
	"time"
)

// SpanStatus represents the status of a span
type SpanStatus string

const (
	SpanStatusUnset SpanStatus = "unset"
	SpanStatusOK    SpanStatus = "ok"
	SpanStatusError SpanStatus = "error"
)

// Span represents a single span in a distributed trace
type Span struct {
	TraceID       string                 `json:"trace_id"`
	SpanID        string                 `json:"span_id"`
	ParentSpanID  string                 `json:"parent_span_id,omitempty"`
	ServiceName   string                 `json:"service_name"`
	OperationName string                 `json:"operation_name"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       time.Time              `json:"end_time"`
	Duration      time.Duration          `json:"duration_ns"`
	StatusCode    SpanStatus             `json:"status_code"`
	ResourceAttrs map[string]string      `json:"resource_attrs,omitempty"`
	Attributes    map[string]interface{} `json:"attributes,omitempty"`
}

// Trace represents a complete trace with all its spans
type Trace struct {
	TraceID    string    `json:"trace_id"`
	Spans      []Span    `json:"spans"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Duration   time.Duration `json:"duration"`
	SpanCount  int       `json:"span_count"`
	Services   []string  `json:"services"`
	HasError   bool      `json:"has_error"`
}

// TraceSummary represents a lightweight view of a trace
type TraceSummary struct {
	TraceID     string        `json:"trace_id"`
	RootService string        `json:"root_service"`
	RootOp      string        `json:"root_operation"`
	StartTime   time.Time     `json:"start_time"`
	Duration    time.Duration `json:"duration"`
	SpanCount   int           `json:"span_count"`
	ErrorCount  int           `json:"error_count"`
	Services    []string      `json:"services"`
}

// GetRootSpan returns the root span (span without parent) of the trace
func (t *Trace) GetRootSpan() *Span {
	for i := range t.Spans {
		if t.Spans[i].ParentSpanID == "" {
			return &t.Spans[i]
		}
	}
	return nil
}

// GetService extracts the service name from span resource attributes
func (s *Span) GetService() string {
	if s.ServiceName != "" {
		return s.ServiceName
	}
	if svc, ok := s.ResourceAttrs["service.name"]; ok {
		return svc
	}
	return "unknown"
}

// GetResourceID extracts cloud resource ID from span attributes
func (s *Span) GetResourceID() string {
	// Try cloud.resource.id first
	if id, ok := s.ResourceAttrs["cloud.resource.id"]; ok {
		return id
	}
	// Try AWS-specific attributes
	if id, ok := s.ResourceAttrs["aws.ecs.container.arn"]; ok {
		return id
	}
	if id, ok := s.ResourceAttrs["aws.ecs.task.arn"]; ok {
		return id
	}
	// Try K8s attributes
	if id, ok := s.ResourceAttrs["k8s.pod.name"]; ok {
		return id
	}
	// Fallback to host.name
	if id, ok := s.ResourceAttrs["host.name"]; ok {
		return id
	}
	return ""
}

// IsError returns true if the span has an error status
func (s *Span) IsError() bool {
	return s.StatusCode == SpanStatusError
}
