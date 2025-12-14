package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/higakikeita/airdig/deepdrift/pkg/storage/clickhouse"
	"github.com/higakikeita/airdig/deepdrift/pkg/types"
)

// handleHealth handles health check requests
func (s *Server) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		ctx := r.Context()
		status, err := s.chClient.HealthCheck(ctx)
		if err != nil {
			respondError(w, http.StatusServiceUnavailable, "ClickHouse unhealthy: "+err.Error())
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"status":     "healthy",
			"clickhouse": status,
			"timestamp":  time.Now().Unix(),
		})
	}
}

// handleDrifts handles drift events list requests
func (s *Server) handleDrifts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		ctx := r.Context()

		// Parse query parameters
		limit := parseQueryInt(r, "limit", 100)
		resourceType := parseQueryString(r, "resource_type", "")
		severity := parseQueryString(r, "severity", "")
		driftType := parseQueryString(r, "drift_type", "")
		userIdentity := parseQueryString(r, "user_identity", "")

		startTime, _ := parseQueryTime(r, "start_time")
		endTime, _ := parseQueryTime(r, "end_time")

		// Build filter
		filter := &clickhouse.DriftEventFilter{
			StartTime:    startTime,
			EndTime:      endTime,
			ResourceType: resourceType,
			UserIdentity: userIdentity,
			Limit:        limit,
		}

		if severity != "" {
			filter.Severity = types.Severity(severity)
		}
		if driftType != "" {
			filter.DriftType = types.DriftType(driftType)
		}

		// Query drifts
		drifts, err := s.driftStore.ListDriftEvents(ctx, filter)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to query drifts: "+err.Error())
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"drifts": drifts,
			"count":  len(drifts),
		})
	}
}

// handleDriftByID handles single drift event requests
func (s *Server) handleDriftByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Extract ID from path
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/drifts/")
		if path == "" || strings.Contains(path, "/") {
			respondError(w, http.StatusBadRequest, "Invalid drift ID")
			return
		}

		ctx := r.Context()
		drift, err := s.driftStore.GetDriftEvent(ctx, path)
		if err != nil {
			respondError(w, http.StatusNotFound, "Drift not found: "+err.Error())
			return
		}

		respondJSON(w, http.StatusOK, drift)
	}
}

// handleDriftStats handles drift statistics requests
func (s *Server) handleDriftStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		ctx := r.Context()
		days := parseQueryInt(r, "days", 7)

		stats, err := s.driftStore.GetDriftStats(ctx, days)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to get stats: "+err.Error())
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"stats": stats,
			"days":  days,
		})
	}
}

// handleImpactAnalysis handles impact analysis list requests
func (s *Server) handleImpactAnalysis() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		ctx := r.Context()

		// Parse query parameters
		limit := parseQueryInt(r, "limit", 100)
		severity := parseQueryString(r, "severity", "")
		minBlastRadius := parseQueryInt(r, "min_blast_radius", 0)
		minAffectedResources := parseQueryInt(r, "min_affected_resources", 0)

		startTime, _ := parseQueryTime(r, "start_time")
		endTime, _ := parseQueryTime(r, "end_time")

		// Build filter
		filter := &clickhouse.ImpactAnalysisFilter{
			StartTime:            startTime,
			EndTime:              endTime,
			MinBlastRadius:       minBlastRadius,
			MinAffectedResources: minAffectedResources,
			Limit:                limit,
		}

		if severity != "" {
			filter.Severity = types.Severity(severity)
		}

		// Query impact analysis
		results, err := s.impactStore.ListImpactAnalysis(ctx, filter)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to query impact analysis: "+err.Error())
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"results": results,
			"count":   len(results),
		})
	}
}

// handleImpactByDriftID handles impact analysis by drift ID
func (s *Server) handleImpactByDriftID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Extract drift ID from path
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/impact/")
		if path == "" || strings.Contains(path, "/") {
			respondError(w, http.StatusBadRequest, "Invalid drift ID")
			return
		}

		ctx := r.Context()
		result, err := s.impactStore.GetImpactAnalysis(ctx, path)
		if err != nil {
			respondError(w, http.StatusNotFound, "Impact analysis not found: "+err.Error())
			return
		}

		respondJSON(w, http.StatusOK, result)
	}
}

// handleImpactStats handles impact analysis statistics
func (s *Server) handleImpactStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		ctx := r.Context()
		days := parseQueryInt(r, "days", 7)

		stats, err := s.impactStore.GetImpactStats(ctx, days)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to get stats: "+err.Error())
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"stats": stats,
			"days":  days,
		})
	}
}

// handleHighImpactDrifts handles high impact drifts requests
func (s *Server) handleHighImpactDrifts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		ctx := r.Context()
		days := parseQueryInt(r, "days", 7)
		limit := parseQueryInt(r, "limit", 50)

		drifts, err := s.impactStore.GetHighImpactDrifts(ctx, days, limit)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to get high impact drifts: "+err.Error())
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"drifts": drifts,
			"count":  len(drifts),
			"days":   days,
		})
	}
}

// handleCreateDrift handles drift event creation (for testing/manual input)
func (s *Server) handleCreateDrift() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var event types.DriftEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		ctx := r.Context()
		if err := s.driftStore.SaveDriftEvent(ctx, &event); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to save drift event: "+err.Error())
			return
		}

		respondJSON(w, http.StatusCreated, map[string]interface{}{
			"message": "Drift event created",
			"id":      event.ID,
		})
	}
}

// handleCreateImpact handles impact analysis creation (for testing/manual input)
func (s *Server) handleCreateImpact() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var result types.ImpactAnalysisResult
		if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		ctx := r.Context()
		if err := s.impactStore.SaveImpactAnalysis(ctx, &result); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to save impact analysis: "+err.Error())
			return
		}

		respondJSON(w, http.StatusCreated, map[string]interface{}{
			"message":       "Impact analysis created",
			"drift_event_id": result.DriftEventID,
		})
	}
}

// GraphNode represents a node in the resource graph
type GraphNode struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Provider  string                 `json:"provider"`
	Region    string                 `json:"region"`
	Name      string                 `json:"name"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Tags      map[string]string      `json:"tags,omitempty"`
	CreatedAt string                 `json:"created_at,omitempty"`
	UpdatedAt string                 `json:"updated_at,omitempty"`
}

// GraphEdge represents an edge in the resource graph
type GraphEdge struct {
	From     string                 `json:"from"`
	To       string                 `json:"to"`
	Type     string                 `json:"type"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ResourceGraph represents the complete resource graph
type ResourceGraph struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

// handleGraph returns the resource graph data
func (s *Server) handleGraph() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// For now, return sample graph data
		// TODO: Load actual graph from file or database
		graph := ResourceGraph{
			Nodes: []GraphNode{
				{
					ID:       "aws:vpc:vpc-123456",
					Type:     "vpc",
					Provider: "aws",
					Region:   "us-east-1",
					Name:     "production-vpc",
					Metadata: map[string]interface{}{"cidr": "10.0.0.0/16"},
					Tags:     map[string]string{"Environment": "production"},
				},
				{
					ID:       "aws:subnet:subnet-123456",
					Type:     "subnet",
					Provider: "aws",
					Region:   "us-east-1",
					Name:     "production-subnet-1a",
					Metadata: map[string]interface{}{
						"availability_zone": "us-east-1a",
						"cidr":              "10.0.1.0/24",
					},
					Tags: map[string]string{"Environment": "production"},
				},
				{
					ID:       "aws:sg:sg-123456",
					Type:     "security_group",
					Provider: "aws",
					Region:   "us-east-1",
					Name:     "web-sg",
					Tags:     map[string]string{"Environment": "production"},
				},
				{
					ID:       "aws:ec2:i-123456",
					Type:     "ec2",
					Provider: "aws",
					Region:   "us-east-1",
					Name:     "web-server-1",
					Metadata: map[string]interface{}{
						"instance_type": "t3.large",
						"private_ip":    "10.0.1.10",
						"public_ip":     "54.123.45.67",
						"state":         "running",
					},
					Tags: map[string]string{
						"Environment": "production",
						"Role":        "web",
					},
				},
				{
					ID:       "aws:rds:database-1",
					Type:     "rds",
					Provider: "aws",
					Region:   "us-east-1",
					Name:     "production-db",
					Metadata: map[string]interface{}{
						"engine":         "postgres",
						"engine_version": "14.7",
						"instance_class": "db.t3.medium",
						"storage":        100,
					},
					Tags: map[string]string{"Environment": "production"},
				},
			},
			Edges: []GraphEdge{
				{From: "aws:vpc:vpc-123456", To: "aws:subnet:subnet-123456", Type: "ownership"},
				{From: "aws:subnet:subnet-123456", To: "aws:ec2:i-123456", Type: "network"},
				{From: "aws:sg:sg-123456", To: "aws:ec2:i-123456", Type: "network"},
				{From: "aws:subnet:subnet-123456", To: "aws:rds:database-1", Type: "network"},
				{
					From: "aws:ec2:i-123456",
					To:   "aws:rds:database-1",
					Type: "dependency",
					Metadata: map[string]interface{}{
						"connection_type": "database",
					},
				},
			},
		}

		respondJSON(w, http.StatusOK, graph)
	}
}
