package models

import (
	"time"
)

// ServiceNode represents a service in the service map
type ServiceNode struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	ResourceID string                 `json:"resource_id,omitempty"`
	Endpoints  []string               `json:"endpoints"`
	FirstSeen  time.Time              `json:"first_seen"`
	LastSeen   time.Time              `json:"last_seen"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ServiceEdge represents a dependency between two services with metrics
type ServiceEdge struct {
	From         string        `json:"from"`
	To           string        `json:"to"`
	WindowStart  time.Time     `json:"window_start"`
	WindowEnd    time.Time     `json:"window_end"`
	LatencyP50   time.Duration `json:"latency_p50"`
	LatencyP95   time.Duration `json:"latency_p95"`
	LatencyP99   time.Duration `json:"latency_p99"`
	RequestCount uint64        `json:"request_count"`
	ErrorCount   uint64        `json:"error_count"`
	ErrorRate    float64       `json:"error_rate"`
	RequestRate  float64       `json:"request_rate"`
	SampleTraces []string      `json:"sample_traces,omitempty"`
}

// ServiceMap represents the complete service dependency graph
type ServiceMap struct {
	Nodes     []ServiceNode `json:"nodes"`
	Edges     []ServiceEdge `json:"edges"`
	Timestamp time.Time     `json:"timestamp"`
}

// ServiceStats represents aggregated statistics for a service
type ServiceStats struct {
	ServiceName  string        `json:"service_name"`
	RequestCount uint64        `json:"request_count"`
	ErrorCount   uint64        `json:"error_count"`
	ErrorRate    float64       `json:"error_rate"`
	AvgLatency   time.Duration `json:"avg_latency"`
	P95Latency   time.Duration `json:"p95_latency"`
	P99Latency   time.Duration `json:"p99_latency"`
	Period       string        `json:"period"`
}

// ServiceOperation represents an operation within a service
type ServiceOperation struct {
	ServiceName  string        `json:"service_name"`
	OperationName string       `json:"operation_name"`
	Count        uint64        `json:"count"`
	ErrorCount   uint64        `json:"error_count"`
	ErrorRate    float64       `json:"error_rate"`
	AvgDuration  time.Duration `json:"avg_duration"`
	P95Duration  time.Duration `json:"p95_duration"`
}
