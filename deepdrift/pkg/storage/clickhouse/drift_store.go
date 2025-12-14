package clickhouse

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/higakikeita/airdig/deepdrift/pkg/types"
)

// DriftStore handles drift event storage operations
type DriftStore struct {
	client *Client
}

// NewDriftStore creates a new drift store
func NewDriftStore(client *Client) *DriftStore {
	return &DriftStore{
		client: client,
	}
}

// SaveDriftEvent saves a single drift event to ClickHouse
func (s *DriftStore) SaveDriftEvent(ctx context.Context, event *types.DriftEvent) error {
	query := `
		INSERT INTO drift_events (
			id, resource_id, resource_type, drift_type, severity,
			timestamp, state_before, state_after, diff,
			cloudtrail_event_id, event_name, user_identity, user_arn,
			source_ip, root_cause_timestamp, date
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?
		)
	`

	// Convert maps to JSON strings
	stateBefore, _ := json.Marshal(event.Before)
	stateAfter, _ := json.Marshal(event.After)
	diff, _ := json.Marshal(event.Diff)

	// Extract root cause fields
	var cloudtrailEventID, eventName, userIdentity, userARN, sourceIP string
	var rootCauseTimestamp time.Time
	if event.RootCause != nil {
		cloudtrailEventID = event.RootCause.CloudTrailEventID
		eventName = event.RootCause.EventName
		userIdentity = event.RootCause.UserIdentity
		userARN = event.RootCause.UserARN
		sourceIP = event.RootCause.SourceIP
		rootCauseTimestamp = event.RootCause.Timestamp
	}

	return s.client.Exec(ctx, query,
		event.ID,
		event.ResourceID,
		event.ResourceType,
		string(event.Type),
		string(event.Severity),
		event.Timestamp,
		string(stateBefore),
		string(stateAfter),
		string(diff),
		cloudtrailEventID,
		eventName,
		userIdentity,
		userARN,
		sourceIP,
		rootCauseTimestamp,
		event.Timestamp.Truncate(24*time.Hour), // date
	)
}

// SaveDriftEvents saves multiple drift events in a batch
func (s *DriftStore) SaveDriftEvents(ctx context.Context, events []*types.DriftEvent) error {
	if len(events) == 0 {
		return nil
	}

	batch, err := s.client.PrepareBatch(ctx, `
		INSERT INTO drift_events (
			id, resource_id, resource_type, drift_type, severity,
			timestamp, state_before, state_after, diff,
			cloudtrail_event_id, event_name, user_identity, user_arn,
			source_ip, root_cause_timestamp, date
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, event := range events {
		stateBefore, _ := json.Marshal(event.Before)
		stateAfter, _ := json.Marshal(event.After)
		diff, _ := json.Marshal(event.Diff)

		var cloudtrailEventID, eventName, userIdentity, userARN, sourceIP string
		var rootCauseTimestamp time.Time
		if event.RootCause != nil {
			cloudtrailEventID = event.RootCause.CloudTrailEventID
			eventName = event.RootCause.EventName
			userIdentity = event.RootCause.UserIdentity
			userARN = event.RootCause.UserARN
			sourceIP = event.RootCause.SourceIP
			rootCauseTimestamp = event.RootCause.Timestamp
		}

		if err := batch.Append(
			event.ID,
			event.ResourceID,
			event.ResourceType,
			string(event.Type),
			string(event.Severity),
			event.Timestamp,
			string(stateBefore),
			string(stateAfter),
			string(diff),
			cloudtrailEventID,
			eventName,
			userIdentity,
			userARN,
			sourceIP,
			rootCauseTimestamp,
			event.Timestamp.Truncate(24*time.Hour),
		); err != nil {
			return fmt.Errorf("failed to append to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	return nil
}

// GetDriftEvent retrieves a single drift event by ID
func (s *DriftStore) GetDriftEvent(ctx context.Context, id string) (*types.DriftEvent, error) {
	query := `
		SELECT
			id, resource_id, resource_type, drift_type, severity,
			timestamp, state_before, state_after, diff,
			cloudtrail_event_id, event_name, user_identity, user_arn,
			source_ip, root_cause_timestamp
		FROM drift_events
		WHERE id = ?
		LIMIT 1
	`

	var event types.DriftEvent
	var driftType, severity string
	var stateBefore, stateAfter, diff string
	var cloudtrailEventID, eventName, userIdentity, userARN, sourceIP string
	var rootCauseTimestamp time.Time

	row := s.client.QueryRow(ctx, query, id)
	err := row.Scan(
		&event.ID,
		&event.ResourceID,
		&event.ResourceType,
		&driftType,
		&severity,
		&event.Timestamp,
		&stateBefore,
		&stateAfter,
		&diff,
		&cloudtrailEventID,
		&eventName,
		&userIdentity,
		&userARN,
		&sourceIP,
		&rootCauseTimestamp,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get drift event: %w", err)
	}

	event.Type = types.DriftType(driftType)
	event.Severity = types.Severity(severity)

	// Parse JSON strings
	json.Unmarshal([]byte(stateBefore), &event.Before)
	json.Unmarshal([]byte(stateAfter), &event.After)
	json.Unmarshal([]byte(diff), &event.Diff)

	// Set root cause if present
	if cloudtrailEventID != "" {
		event.RootCause = &types.RootCause{
			CloudTrailEventID: cloudtrailEventID,
			EventName:         eventName,
			UserIdentity:      userIdentity,
			UserARN:           userARN,
			SourceIP:          sourceIP,
			Timestamp:         rootCauseTimestamp,
		}
	}

	return &event, nil
}

// ListDriftEvents lists drift events with filters
func (s *DriftStore) ListDriftEvents(ctx context.Context, filter *DriftEventFilter) ([]*types.DriftEvent, error) {
	query := `
		SELECT
			id, resource_id, resource_type, drift_type, severity,
			timestamp, state_before, state_after, diff,
			cloudtrail_event_id, event_name, user_identity, user_arn,
			source_ip, root_cause_timestamp
		FROM drift_events
		WHERE 1=1
	`

	args := []interface{}{}

	if filter != nil {
		if !filter.StartTime.IsZero() {
			query += " AND timestamp >= ?"
			args = append(args, filter.StartTime)
		}
		if !filter.EndTime.IsZero() {
			query += " AND timestamp <= ?"
			args = append(args, filter.EndTime)
		}
		if filter.ResourceType != "" {
			query += " AND resource_type = ?"
			args = append(args, filter.ResourceType)
		}
		if filter.DriftType != "" {
			query += " AND drift_type = ?"
			args = append(args, string(filter.DriftType))
		}
		if filter.Severity != "" {
			query += " AND severity = ?"
			args = append(args, string(filter.Severity))
		}
		if filter.UserIdentity != "" {
			query += " AND user_identity = ?"
			args = append(args, filter.UserIdentity)
		}
	}

	query += " ORDER BY timestamp DESC"

	if filter != nil && filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	rows, err := s.client.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query drift events: %w", err)
	}
	defer rows.Close()

	events := []*types.DriftEvent{}
	for rows.Next() {
		var event types.DriftEvent
		var driftType, severity string
		var stateBefore, stateAfter, diff string
		var cloudtrailEventID, eventName, userIdentity, userARN, sourceIP string
		var rootCauseTimestamp time.Time

		if err := rows.Scan(
			&event.ID,
			&event.ResourceID,
			&event.ResourceType,
			&driftType,
			&severity,
			&event.Timestamp,
			&stateBefore,
			&stateAfter,
			&diff,
			&cloudtrailEventID,
			&eventName,
			&userIdentity,
			&userARN,
			&sourceIP,
			&rootCauseTimestamp,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		event.Type = types.DriftType(driftType)
		event.Severity = types.Severity(severity)

		json.Unmarshal([]byte(stateBefore), &event.Before)
		json.Unmarshal([]byte(stateAfter), &event.After)
		json.Unmarshal([]byte(diff), &event.Diff)

		if cloudtrailEventID != "" {
			event.RootCause = &types.RootCause{
				CloudTrailEventID: cloudtrailEventID,
				EventName:         eventName,
				UserIdentity:      userIdentity,
				UserARN:           userARN,
				SourceIP:          sourceIP,
				Timestamp:         rootCauseTimestamp,
			}
		}

		events = append(events, &event)
	}

	return events, nil
}

// GetDriftStats returns drift statistics
func (s *DriftStore) GetDriftStats(ctx context.Context, days int) (*DriftStats, error) {
	stats := &DriftStats{}

	// Total count
	query := `
		SELECT count()
		FROM drift_events
		WHERE date >= today() - INTERVAL ? DAY
	`
	if err := s.client.QueryRow(ctx, query, days).Scan(&stats.TotalCount); err != nil {
		return nil, err
	}

	// Count by severity
	query = `
		SELECT severity, count()
		FROM drift_events
		WHERE date >= today() - INTERVAL ? DAY
		GROUP BY severity
	`
	rows, err := s.client.Query(ctx, query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats.BySeverity = make(map[string]uint64)
	for rows.Next() {
		var severity string
		var count uint64
		if err := rows.Scan(&severity, &count); err != nil {
			return nil, err
		}
		stats.BySeverity[severity] = count
	}

	// Count by type
	query = `
		SELECT drift_type, count()
		FROM drift_events
		WHERE date >= today() - INTERVAL ? DAY
		GROUP BY drift_type
	`
	rows, err = s.client.Query(ctx, query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats.ByType = make(map[string]uint64)
	for rows.Next() {
		var driftType string
		var count uint64
		if err := rows.Scan(&driftType, &count); err != nil {
			return nil, err
		}
		stats.ByType[driftType] = count
	}

	// Count by resource type
	query = `
		SELECT resource_type, count()
		FROM drift_events
		WHERE date >= today() - INTERVAL ? DAY
		GROUP BY resource_type
		ORDER BY count() DESC
		LIMIT 10
	`
	rows, err = s.client.Query(ctx, query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats.ByResourceType = make(map[string]uint64)
	for rows.Next() {
		var resourceType string
		var count uint64
		if err := rows.Scan(&resourceType, &count); err != nil {
			return nil, err
		}
		stats.ByResourceType[resourceType] = count
	}

	return stats, nil
}

// DriftEventFilter defines filters for querying drift events
type DriftEventFilter struct {
	StartTime    time.Time
	EndTime      time.Time
	ResourceType string
	DriftType    types.DriftType
	Severity     types.Severity
	UserIdentity string
	Limit        int
}

// DriftStats contains drift statistics
type DriftStats struct {
	TotalCount       uint64            `json:"total_count"`
	BySeverity       map[string]uint64 `json:"by_severity"`
	ByType           map[string]uint64 `json:"by_type"`
	ByResourceType   map[string]uint64 `json:"by_resource_type"`
}

// DeleteOldDriftEvents deletes drift events older than specified days
func (s *DriftStore) DeleteOldDriftEvents(ctx context.Context, days int) (uint64, error) {
	query := `
		ALTER TABLE drift_events
		DELETE WHERE date < today() - INTERVAL ? DAY
	`

	if err := s.client.Exec(ctx, query, days); err != nil {
		return 0, fmt.Errorf("failed to delete old drift events: %w", err)
	}

	// Get affected rows (this is an approximation)
	var count uint64
	countQuery := `
		SELECT count()
		FROM drift_events
		WHERE date < today() - INTERVAL ? DAY
	`
	if err := s.client.QueryRow(ctx, countQuery, days).Scan(&count); err != nil {
		// If query fails, return 0 but don't fail the operation
		count = 0
	}

	return count, nil
}
