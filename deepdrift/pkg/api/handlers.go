package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/higakikeita/airdig/deepdrift/pkg/drift"
	"github.com/higakikeita/airdig/deepdrift/pkg/storage/clickhouse"
	"github.com/higakikeita/airdig/deepdrift/pkg/terraform"
	"github.com/higakikeita/airdig/deepdrift/pkg/types"
)

// handleHealth handles health check requests
func (s *Server) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		response := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
		}

		if s.chClient != nil {
			ctx := r.Context()
			status, err := s.chClient.HealthCheck(ctx)
			if err == nil {
				response["clickhouse"] = status
			}
		} else {
			response["mode"] = "in-memory"
		}

		respondJSON(w, http.StatusOK, response)
	}
}

// detectDrifts performs real-time drift detection by comparing Terraform state with AWS resources
func (s *Server) detectDrifts(ctx context.Context) ([]*types.DriftEvent, error) {
	// 1. Load Terraform state
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	statePath := filepath.Join(homeDir, "tfdrift-falco", "examples", "terraform", "terraform.tfstate")

	// Check if comprehensive test state exists
	comprehensivePath := filepath.Join(homeDir, "tfdrift-falco", "examples", "comprehensive-test", "terraform.tfstate")
	if _, err := os.Stat(comprehensivePath); err == nil {
		statePath = comprehensivePath
	}

	tfReader := terraform.NewStateReader(statePath)
	tfState, err := tfReader.Load()
	if err != nil {
		log.Printf("Warning: Failed to load Terraform state from %s: %v", statePath, err)
		return []*types.DriftEvent{}, nil
	}

	log.Printf("Loaded Terraform state with %d resources", len(tfState.Resources))

	// 2. Get AWS EC2 instances from SkyGraph
	awsInstances, err := s.getAWSEC2Instances(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS resources: %w", err)
	}

	log.Printf("Retrieved %d AWS EC2 instances from SkyGraph", len(awsInstances))

	// 3. Compare Terraform resources with AWS resources
	comparator := drift.NewComparator()
	drifts := []*types.DriftEvent{}

	tfEC2Instances := tfState.GetEC2Instances()
	log.Printf("Found %d EC2 instances in Terraform state", len(tfEC2Instances))

	for _, tfResource := range tfEC2Instances {
		if len(tfResource.Instances) == 0 {
			continue
		}

		tfID, ok := tfResource.Instances[0].Attributes["id"].(string)
		if !ok {
			continue
		}

		// Find matching AWS resource
		for _, awsResource := range awsInstances {
			awsID, _ := awsResource["instance_id"].(string)
			if awsID == tfID {
				if driftEvent := comparator.CompareEC2(tfResource, awsResource); driftEvent != nil {
					log.Printf("Drift detected for %s: %s", tfID, driftEvent.Diff)
					drifts = append(drifts, driftEvent)
				}
				break
			}
		}
	}

	// 4. Check Security Groups
	awsSGs, err := s.getAWSSecurityGroups(ctx)
	if err == nil {
		tfSGs := tfState.GetSecurityGroups()
		log.Printf("Found %d Security Groups in Terraform state", len(tfSGs))

		for _, tfResource := range tfSGs {
			if len(tfResource.Instances) == 0 {
				continue
			}

			tfID, ok := tfResource.Instances[0].Attributes["id"].(string)
			if !ok {
				continue
			}

			for _, awsResource := range awsSGs {
				awsID, _ := awsResource["security_group_id"].(string)
				if awsID == tfID {
					if driftEvent := comparator.CompareSecurityGroup(tfResource, awsResource); driftEvent != nil {
						log.Printf("Security Group drift detected for %s", tfID)
						drifts = append(drifts, driftEvent)
					}
					break
				}
			}
		}
	}

	// 5. Check S3 Buckets
	awsS3Buckets, err := s.getAWSS3Buckets(ctx)
	if err == nil {
		tfS3Buckets := tfState.GetS3Buckets()
		log.Printf("Found %d S3 buckets in Terraform state", len(tfS3Buckets))

		for _, tfResource := range tfS3Buckets {
			if len(tfResource.Instances) == 0 {
				continue
			}

			tfBucket, ok := tfResource.Instances[0].Attributes["bucket"].(string)
			if !ok {
				continue
			}

			for _, awsResource := range awsS3Buckets {
				awsBucket, _ := awsResource["bucket_name"].(string)
				if awsBucket == tfBucket {
					if driftEvent := comparator.CompareS3(tfResource, awsResource); driftEvent != nil {
						log.Printf("S3 drift detected for %s", tfBucket)
						drifts = append(drifts, driftEvent)
					}
					break
				}
			}
		}
	}

	log.Printf("Detected %d drifts", len(drifts))
	return drifts, nil
}

// getAWSEC2Instances retrieves EC2 instances from SkyGraph
func (s *Server) getAWSEC2Instances(ctx context.Context) ([]map[string]interface{}, error) {
	skyGraphURL := "http://localhost:8001/api/v1/graph"

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(skyGraphURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SkyGraph: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("SkyGraph returned status %d: %s", resp.StatusCode, string(body))
	}

	var graphResp struct {
		Nodes []struct {
			ID       string                 `json:"id"`
			Type     string                 `json:"type"`
			Metadata map[string]interface{} `json:"metadata"`
			Tags     map[string]string      `json:"tags"`
		} `json:"nodes"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&graphResp); err != nil {
		return nil, fmt.Errorf("failed to decode SkyGraph response: %w", err)
	}

	// Extract EC2 instances
	instances := []map[string]interface{}{}
	for _, node := range graphResp.Nodes {
		if node.Type == "ec2" {
			instance := make(map[string]interface{})
			for k, v := range node.Metadata {
				instance[k] = v
			}
			instance["tags"] = node.Tags
			instances = append(instances, instance)
		}
	}

	return instances, nil
}

// getAWSSecurityGroups retrieves Security Groups from SkyGraph
func (s *Server) getAWSSecurityGroups(ctx context.Context) ([]map[string]interface{}, error) {
	skyGraphURL := "http://localhost:8001/api/v1/graph"

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(skyGraphURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SkyGraph: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("SkyGraph returned status %d: %s", resp.StatusCode, string(body))
	}

	var graphResp struct {
		Nodes []struct {
			ID       string                 `json:"id"`
			Type     string                 `json:"type"`
			Metadata map[string]interface{} `json:"metadata"`
			Tags     map[string]string      `json:"tags"`
		} `json:"nodes"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&graphResp); err != nil {
		return nil, fmt.Errorf("failed to decode SkyGraph response: %w", err)
	}

	// Extract Security Groups
	sgs := []map[string]interface{}{}
	for _, node := range graphResp.Nodes {
		if node.Type == "security_group" {
			sg := make(map[string]interface{})
			for k, v := range node.Metadata {
				sg[k] = v
			}
			sg["tags"] = node.Tags
			sgs = append(sgs, sg)
		}
	}

	return sgs, nil
}

// getAWSS3Buckets retrieves S3 buckets from SkyGraph
func (s *Server) getAWSS3Buckets(ctx context.Context) ([]map[string]interface{}, error) {
	skyGraphURL := "http://localhost:8001/api/v1/graph"

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(skyGraphURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SkyGraph: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("SkyGraph returned status %d: %s", resp.StatusCode, string(body))
	}

	var graphResp struct {
		Nodes []struct {
			ID       string                 `json:"id"`
			Type     string                 `json:"type"`
			Metadata map[string]interface{} `json:"metadata"`
			Tags     map[string]string      `json:"tags"`
		} `json:"nodes"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&graphResp); err != nil {
		return nil, fmt.Errorf("failed to decode SkyGraph response: %w", err)
	}

	// Extract S3 buckets
	buckets := []map[string]interface{}{}
	for _, node := range graphResp.Nodes {
		if node.Type == "s3" {
			bucket := make(map[string]interface{})
			for k, v := range node.Metadata {
				bucket[k] = v
			}
			bucket["tags"] = node.Tags
			buckets = append(buckets, bucket)
		}
	}

	return buckets, nil
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
		var drifts []*types.DriftEvent
		var err error

		if s.chClient != nil {
			drifts, err = s.driftStore.ListDriftEvents(ctx, filter)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to query drifts: "+err.Error())
				return
			}
		} else {
			// Perform real-time drift detection when ClickHouse is disabled
			drifts, err = s.detectDrifts(ctx)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to detect drifts: "+err.Error())
				return
			}
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

		// If ClickHouse is not configured, return empty stats
		if s.driftStore == nil {
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"stats": map[string]interface{}{
					"total_count": 0,
					"by_severity": map[string]int{
						"critical": 0,
						"high":     0,
						"medium":   0,
						"low":      0,
					},
					"by_type": map[string]int{},
				},
				"days": days,
			})
			return
		}

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

		// If ClickHouse is not configured, return empty stats
		if s.impactStore == nil {
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"stats": map[string]interface{}{
					"total_impact_events":     0,
					"avg_blast_radius":        0.0,
					"avg_affected_resources":  0.0,
					"max_blast_radius":        0.0,
				},
				"days": days,
			})
			return
		}

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

		// Fetch real graph data from SkyGraph
		graph, err := s.getGraphFromSkyGraph(r.Context())
		if err != nil {
			log.Printf("Failed to get graph from SkyGraph: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to load resource graph: "+err.Error())
			return
		}

		respondJSON(w, http.StatusOK, graph)
	}
}

// getGraphFromSkyGraph fetches the resource graph from SkyGraph service
func (s *Server) getGraphFromSkyGraph(ctx context.Context) (*ResourceGraph, error) {
	skyGraphURL := "http://localhost:8001/api/v1/graph"

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(skyGraphURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SkyGraph: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("SkyGraph returned status %d: %s", resp.StatusCode, string(body))
	}

	var skyGraphResp struct {
		Nodes []struct {
			ID       string                 `json:"id"`
			Type     string                 `json:"type"`
			Provider string                 `json:"provider"`
			Region   string                 `json:"region"`
			Name     string                 `json:"name"`
			Metadata map[string]interface{} `json:"metadata"`
			Tags     map[string]string      `json:"tags"`
		} `json:"nodes"`
		Edges []struct {
			From     string                 `json:"from"`
			To       string                 `json:"to"`
			Type     string                 `json:"type"`
			Metadata map[string]interface{} `json:"metadata,omitempty"`
		} `json:"edges"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&skyGraphResp); err != nil {
		return nil, fmt.Errorf("failed to decode SkyGraph response: %w", err)
	}

	// Convert SkyGraph response to ResourceGraph format
	graph := &ResourceGraph{
		Nodes: make([]GraphNode, 0, len(skyGraphResp.Nodes)),
		Edges: make([]GraphEdge, 0, len(skyGraphResp.Edges)),
	}

	for _, node := range skyGraphResp.Nodes {
		graph.Nodes = append(graph.Nodes, GraphNode{
			ID:       node.ID,
			Type:     node.Type,
			Provider: node.Provider,
			Region:   node.Region,
			Name:     node.Name,
			Metadata: node.Metadata,
			Tags:     node.Tags,
		})
	}

	for _, edge := range skyGraphResp.Edges {
		graph.Edges = append(graph.Edges, GraphEdge{
			From:     edge.From,
			To:       edge.To,
			Type:     edge.Type,
			Metadata: edge.Metadata,
		})
	}

	log.Printf("Loaded graph with %d nodes and %d edges from SkyGraph", len(graph.Nodes), len(graph.Edges))
	return graph, nil
}

// handleResources handles resource list requests by querying SkyGraph
func (s *Server) handleResources() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Parse query parameters
		resourceType := parseQueryString(r, "type", "")
		provider := parseQueryString(r, "provider", "")
		region := parseQueryString(r, "region", "")
		limit := parseQueryInt(r, "limit", 100)

		// Query SkyGraph for resources
		skyGraphURL := "http://localhost:8001/api/v1/graph"
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", skyGraphURL, nil)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to create request: "+err.Error())
			return
		}

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to query SkyGraph: "+err.Error())
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("SkyGraph returned status %d: %s", resp.StatusCode, string(body)))
			return
		}

		// Parse SkyGraph response
		type SkyGraphNode struct {
			ID        string                 `json:"id"`
			Type      string                 `json:"type"`
			Provider  string                 `json:"provider"`
			Region    string                 `json:"region"`
			Name      string                 `json:"name"`
			Metadata  map[string]interface{} `json:"metadata"`
			Tags      map[string]string      `json:"tags"`
			CreatedAt string                 `json:"created_at"`
			UpdatedAt string                 `json:"updated_at"`
		}

		type SkyGraphResponse struct {
			Nodes []SkyGraphNode `json:"nodes"`
		}

		var skyGraphResp SkyGraphResponse
		if err := json.NewDecoder(resp.Body).Decode(&skyGraphResp); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to parse SkyGraph response: "+err.Error())
			return
		}

		// Filter resources based on query parameters
		filtered := make([]SkyGraphNode, 0)
		for _, node := range skyGraphResp.Nodes {
			if resourceType != "" && node.Type != resourceType {
				continue
			}
			if provider != "" && node.Provider != provider {
				continue
			}
			if region != "" && node.Region != region {
				continue
			}
			filtered = append(filtered, node)
			if len(filtered) >= limit {
				break
			}
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"resources": filtered,
			"count":     len(filtered),
			"total":     len(skyGraphResp.Nodes),
		})
	}
}

// handleIntendedGraph returns the intended architecture diagram from Terraform state
func (s *Server) handleIntendedGraph() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Load Terraform state
		homeDir, err := os.UserHomeDir()
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to get home directory")
			return
		}

		statePath := filepath.Join(homeDir, "tfdrift-falco", "examples", "terraform", "terraform.tfstate")
		
		// Check if comprehensive test state exists
		comprehensivePath := filepath.Join(homeDir, "tfdrift-falco", "examples", "comprehensive-test", "terraform.tfstate")
		if _, err := os.Stat(comprehensivePath); err == nil {
			statePath = comprehensivePath
		}

		tfReader := terraform.NewStateReader(statePath)
		tfState, err := tfReader.Load()
		if err != nil {
			log.Printf("Failed to load Terraform state from %s: %v", statePath, err)
			respondError(w, http.StatusInternalServerError, "Failed to load Terraform state: "+err.Error())
			return
		}

		// Generate diagram
		diagram := tfState.GenerateDiagram()

		log.Printf("Generated intended diagram with %d nodes and %d edges from Terraform state", 
			len(diagram.Nodes), len(diagram.Edges))

		respondJSON(w, http.StatusOK, diagram)
	}
}
