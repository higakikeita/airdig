package clickhouse

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/higakikeita/airdig/deepdrift/pkg/types"
)

// ImpactStore handles impact analysis storage operations
type ImpactStore struct {
	client *Client
}

// NewImpactStore creates a new impact store
func NewImpactStore(client *Client) *ImpactStore {
	return &ImpactStore{
		client: client,
	}
}

// SaveImpactAnalysis saves an impact analysis result
func (s *ImpactStore) SaveImpactAnalysis(ctx context.Context, result *types.ImpactAnalysisResult) error {
	// Save to impact_analysis table
	affectedResourcesJSON, _ := json.Marshal(result.AffectedResources)
	recommendationsJSON, _ := json.Marshal(result.Recommendations)

	query := `
		INSERT INTO impact_analysis (
			drift_event_id, affected_resource_count, blast_radius, severity,
			affected_resources, recommendations, analyzed_at, date
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	analyzedAt := time.Now()
	if err := s.client.Exec(ctx, query,
		result.DriftEventID,
		result.AffectedResourceCount,
		result.BlastRadius,
		string(result.Severity),
		string(affectedResourcesJSON),
		string(recommendationsJSON),
		analyzedAt,
		analyzedAt.Truncate(24*time.Hour),
	); err != nil {
		return fmt.Errorf("failed to save impact analysis: %w", err)
	}

	// Save individual affected resources to normalized table
	if len(result.AffectedResources) > 0 {
		if err := s.saveAffectedResources(ctx, result.DriftEventID, result.AffectedResources); err != nil {
			return fmt.Errorf("failed to save affected resources: %w", err)
		}
	}

	return nil
}

// SaveImpactAnalysisBatch saves multiple impact analysis results
func (s *ImpactStore) SaveImpactAnalysisBatch(ctx context.Context, results []*types.ImpactAnalysisResult) error {
	if len(results) == 0 {
		return nil
	}

	batch, err := s.client.PrepareBatch(ctx, `
		INSERT INTO impact_analysis (
			drift_event_id, affected_resource_count, blast_radius, severity,
			affected_resources, recommendations, analyzed_at, date
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	analyzedAt := time.Now()
	for _, result := range results {
		affectedResourcesJSON, _ := json.Marshal(result.AffectedResources)
		recommendationsJSON, _ := json.Marshal(result.Recommendations)

		if err := batch.Append(
			result.DriftEventID,
			result.AffectedResourceCount,
			result.BlastRadius,
			string(result.Severity),
			string(affectedResourcesJSON),
			string(recommendationsJSON),
			analyzedAt,
			analyzedAt.Truncate(24*time.Hour),
		); err != nil {
			return fmt.Errorf("failed to append to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	// Save affected resources
	for _, result := range results {
		if len(result.AffectedResources) > 0 {
			if err := s.saveAffectedResources(ctx, result.DriftEventID, result.AffectedResources); err != nil {
				return fmt.Errorf("failed to save affected resources: %w", err)
			}
		}
	}

	return nil
}

// saveAffectedResources saves affected resources to normalized table
func (s *ImpactStore) saveAffectedResources(ctx context.Context, driftEventID string, resources []types.AffectedResource) error {
	if len(resources) == 0 {
		return nil
	}

	batch, err := s.client.PrepareBatch(ctx, `
		INSERT INTO affected_resources (
			drift_event_id, resource_id, resource_type, relation_type,
			distance, impact_description, date
		)
	`)
	if err != nil {
		return err
	}

	date := time.Now().Truncate(24 * time.Hour)
	for _, resource := range resources {
		if err := batch.Append(
			driftEventID,
			resource.ResourceID,
			resource.ResourceType,
			resource.RelationType,
			resource.Distance,
			resource.ImpactDescription,
			date,
		); err != nil {
			return err
		}
	}

	return batch.Send()
}

// GetImpactAnalysis retrieves impact analysis for a drift event
func (s *ImpactStore) GetImpactAnalysis(ctx context.Context, driftEventID string) (*types.ImpactAnalysisResult, error) {
	query := `
		SELECT
			drift_event_id, affected_resource_count, blast_radius, severity,
			affected_resources, recommendations, analyzed_at
		FROM impact_analysis
		WHERE drift_event_id = ?
		ORDER BY analyzed_at DESC
		LIMIT 1
	`

	var result types.ImpactAnalysisResult
	var severity string
	var affectedResourcesJSON, recommendationsJSON string
	var analyzedAt time.Time

	row := s.client.QueryRow(ctx, query, driftEventID)
	if err := row.Scan(
		&result.DriftEventID,
		&result.AffectedResourceCount,
		&result.BlastRadius,
		&severity,
		&affectedResourcesJSON,
		&recommendationsJSON,
		&analyzedAt,
	); err != nil {
		return nil, fmt.Errorf("failed to get impact analysis: %w", err)
	}

	result.Severity = types.Severity(severity)

	// Parse JSON
	json.Unmarshal([]byte(affectedResourcesJSON), &result.AffectedResources)
	json.Unmarshal([]byte(recommendationsJSON), &result.Recommendations)

	return &result, nil
}

// ListImpactAnalysis lists impact analysis results with filters
func (s *ImpactStore) ListImpactAnalysis(ctx context.Context, filter *ImpactAnalysisFilter) ([]*types.ImpactAnalysisResult, error) {
	query := `
		SELECT
			drift_event_id, affected_resource_count, blast_radius, severity,
			affected_resources, recommendations, analyzed_at
		FROM impact_analysis
		WHERE 1=1
	`

	args := []interface{}{}

	if filter != nil {
		if !filter.StartTime.IsZero() {
			query += " AND analyzed_at >= ?"
			args = append(args, filter.StartTime)
		}
		if !filter.EndTime.IsZero() {
			query += " AND analyzed_at <= ?"
			args = append(args, filter.EndTime)
		}
		if filter.Severity != "" {
			query += " AND severity = ?"
			args = append(args, string(filter.Severity))
		}
		if filter.MinBlastRadius > 0 {
			query += " AND blast_radius >= ?"
			args = append(args, filter.MinBlastRadius)
		}
		if filter.MinAffectedResources > 0 {
			query += " AND affected_resource_count >= ?"
			args = append(args, filter.MinAffectedResources)
		}
	}

	query += " ORDER BY analyzed_at DESC"

	if filter != nil && filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	rows, err := s.client.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query impact analysis: %w", err)
	}
	defer rows.Close()

	results := []*types.ImpactAnalysisResult{}
	for rows.Next() {
		var result types.ImpactAnalysisResult
		var severity string
		var affectedResourcesJSON, recommendationsJSON string
		var analyzedAt time.Time

		if err := rows.Scan(
			&result.DriftEventID,
			&result.AffectedResourceCount,
			&result.BlastRadius,
			&severity,
			&affectedResourcesJSON,
			&recommendationsJSON,
			&analyzedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		result.Severity = types.Severity(severity)

		json.Unmarshal([]byte(affectedResourcesJSON), &result.AffectedResources)
		json.Unmarshal([]byte(recommendationsJSON), &result.Recommendations)

		results = append(results, &result)
	}

	return results, nil
}

// GetAffectedResources retrieves affected resources for a drift event
func (s *ImpactStore) GetAffectedResources(ctx context.Context, driftEventID string) ([]types.AffectedResource, error) {
	query := `
		SELECT
			resource_id, resource_type, relation_type, distance, impact_description
		FROM affected_resources
		WHERE drift_event_id = ?
		ORDER BY distance, resource_type
	`

	rows, err := s.client.Query(ctx, query, driftEventID)
	if err != nil {
		return nil, fmt.Errorf("failed to query affected resources: %w", err)
	}
	defer rows.Close()

	resources := []types.AffectedResource{}
	for rows.Next() {
		var resource types.AffectedResource
		if err := rows.Scan(
			&resource.ResourceID,
			&resource.ResourceType,
			&resource.RelationType,
			&resource.Distance,
			&resource.ImpactDescription,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		resources = append(resources, resource)
	}

	return resources, nil
}

// GetImpactStats returns impact analysis statistics
func (s *ImpactStore) GetImpactStats(ctx context.Context, days int) (*ImpactStats, error) {
	stats := &ImpactStats{}

	// Total analyzed
	query := `
		SELECT count()
		FROM impact_analysis
		WHERE date >= today() - INTERVAL ? DAY
	`
	if err := s.client.QueryRow(ctx, query, days).Scan(&stats.TotalAnalyzed); err != nil {
		return nil, err
	}

	// Average blast radius
	query = `
		SELECT avg(blast_radius)
		FROM impact_analysis
		WHERE date >= today() - INTERVAL ? DAY
	`
	if err := s.client.QueryRow(ctx, query, days).Scan(&stats.AvgBlastRadius); err != nil {
		return nil, err
	}

	// Average affected resources
	query = `
		SELECT avg(affected_resource_count)
		FROM impact_analysis
		WHERE date >= today() - INTERVAL ? DAY
	`
	if err := s.client.QueryRow(ctx, query, days).Scan(&stats.AvgAffectedResources); err != nil {
		return nil, err
	}

	// Top affected resource types
	query = `
		SELECT resource_type, count() as cnt
		FROM affected_resources
		WHERE date >= today() - INTERVAL ? DAY
		GROUP BY resource_type
		ORDER BY cnt DESC
		LIMIT 10
	`
	rows, err := s.client.Query(ctx, query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats.TopAffectedResourceTypes = make(map[string]uint64)
	for rows.Next() {
		var resourceType string
		var count uint64
		if err := rows.Scan(&resourceType, &count); err != nil {
			return nil, err
		}
		stats.TopAffectedResourceTypes[resourceType] = count
	}

	return stats, nil
}

// GetHighImpactDrifts returns drifts with high impact
func (s *ImpactStore) GetHighImpactDrifts(ctx context.Context, days int, limit int) ([]*HighImpactDrift, error) {
	query := `
		SELECT
			de.id,
			de.resource_id,
			de.resource_type,
			de.drift_type,
			de.timestamp,
			ia.affected_resource_count,
			ia.blast_radius,
			ia.severity
		FROM drift_events de
		JOIN impact_analysis ia ON de.id = ia.drift_event_id
		WHERE de.date >= today() - INTERVAL ? DAY
		ORDER BY ia.blast_radius DESC, ia.affected_resource_count DESC
		LIMIT ?
	`

	rows, err := s.client.Query(ctx, query, days, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query high impact drifts: %w", err)
	}
	defer rows.Close()

	drifts := []*HighImpactDrift{}
	for rows.Next() {
		var drift HighImpactDrift
		var driftType, severity string
		if err := rows.Scan(
			&drift.ID,
			&drift.ResourceID,
			&drift.ResourceType,
			&driftType,
			&drift.Timestamp,
			&drift.AffectedResourceCount,
			&drift.BlastRadius,
			&severity,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		drift.DriftType = types.DriftType(driftType)
		drift.Severity = types.Severity(severity)
		drifts = append(drifts, &drift)
	}

	return drifts, nil
}

// ImpactAnalysisFilter defines filters for querying impact analysis
type ImpactAnalysisFilter struct {
	StartTime            time.Time
	EndTime              time.Time
	Severity             types.Severity
	MinBlastRadius       int
	MinAffectedResources int
	Limit                int
}

// ImpactStats contains impact analysis statistics
type ImpactStats struct {
	TotalAnalyzed            uint64            `json:"total_analyzed"`
	AvgBlastRadius           float64           `json:"avg_blast_radius"`
	AvgAffectedResources     float64           `json:"avg_affected_resources"`
	TopAffectedResourceTypes map[string]uint64 `json:"top_affected_resource_types"`
}

// HighImpactDrift represents a drift with high impact
type HighImpactDrift struct {
	ID                   string           `json:"id"`
	ResourceID           string           `json:"resource_id"`
	ResourceType         string           `json:"resource_type"`
	DriftType            types.DriftType  `json:"drift_type"`
	Timestamp            time.Time        `json:"timestamp"`
	AffectedResourceCount int             `json:"affected_resource_count"`
	BlastRadius          int              `json:"blast_radius"`
	Severity             types.Severity   `json:"severity"`
}
