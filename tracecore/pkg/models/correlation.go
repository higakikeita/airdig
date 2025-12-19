package models

import (
	"time"
)

// DriftEvent represents a drift event from DeepDrift
type DriftEvent struct {
	ID           string                 `json:"id"`
	ResourceID   string                 `json:"resource_id"`
	ResourceType string                 `json:"resource_type"`
	DriftType    string                 `json:"drift_type"`
	Severity     string                 `json:"severity"`
	DetectedAt   time.Time              `json:"detected_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// MetricComparison represents before/after metrics comparison
type MetricComparison struct {
	Before float64 `json:"before"`
	After  float64 `json:"after"`
	Change float64 `json:"change"`
	PctChange float64 `json:"pct_change"`
}

// ServiceImpact represents the impact on a service from a drift event
type ServiceImpact struct {
	ServiceName   string                      `json:"service_name"`
	ResourceID    string                      `json:"resource_id"`
	Latency       MetricComparison            `json:"latency"`
	ErrorRate     MetricComparison            `json:"error_rate"`
	RequestRate   MetricComparison            `json:"request_rate"`
	SampleTraces  []string                    `json:"sample_traces"`
	Metadata      map[string]interface{}      `json:"metadata,omitempty"`
}

// CorrelationReport represents the complete correlation analysis between drift and traces
type CorrelationReport struct {
	DriftEvent       DriftEvent       `json:"drift_event"`
	AnalysisWindow   time.Duration    `json:"analysis_window"`
	WindowBefore     TimeWindow       `json:"window_before"`
	WindowAfter      TimeWindow       `json:"window_after"`
	AffectedServices []ServiceImpact  `json:"affected_services"`
	TotalImpact      MetricComparison `json:"total_impact"`
	Summary          string           `json:"summary"`
	GeneratedAt      time.Time        `json:"generated_at"`
}

// TimeWindow represents a time range
type TimeWindow struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ResourceServiceMapping represents a mapping between infrastructure and services
type ResourceServiceMapping struct {
	ResourceID    string    `json:"resource_id"`
	ResourceType  string    `json:"resource_type"`
	ServiceName   string    `json:"service_name"`
	Confidence    float64   `json:"confidence"`
	FirstSeen     time.Time `json:"first_seen"`
	LastSeen      time.Time `json:"last_seen"`
	TraceCount    int       `json:"trace_count"`
}

// Calculate calculates the change and percentage change
func (mc *MetricComparison) Calculate() {
	mc.Change = mc.After - mc.Before
	if mc.Before != 0 {
		mc.PctChange = (mc.Change / mc.Before) * 100
	}
}
