package types

import "time"

// DriftType は drift の種類を表す
type DriftType string

const (
	DriftCreated  DriftType = "created"
	DriftModified DriftType = "modified"
	DriftDeleted  DriftType = "deleted"
)

// DriftEvent は drift イベントを表す
type DriftEvent struct {
	// ID はイベントの一意識別子
	ID string `json:"id"`

	// ResourceID はリソースの識別子 (e.g., "aws:ec2:i-123456")
	ResourceID string `json:"resource_id"`

	// ResourceType はリソースタイプ (e.g., "ec2", "vpc", "security_group")
	ResourceType string `json:"resource_type"`

	// Type は drift のタイプ (created, modified, deleted)
	Type DriftType `json:"type"`

	// Timestamp はイベント発生時刻
	Timestamp time.Time `json:"timestamp"`

	// Before は変更前の状態
	Before map[string]interface{} `json:"before,omitempty"`

	// After は変更後の状態
	After map[string]interface{} `json:"after,omitempty"`

	// Diff は詳細な差分
	Diff map[string]interface{} `json:"diff,omitempty"`

	// RootCause は原因 (CloudTrail event ID, IAM user, etc.)
	RootCause *RootCause `json:"root_cause,omitempty"`

	// ImpactedResources はこの drift により影響を受けるリソースのリスト
	ImpactedResources []string `json:"impacted_resources,omitempty"`

	// Severity は drift の深刻度
	Severity Severity `json:"severity"`
}

// RootCause は drift の根本原因を表す
type RootCause struct {
	// CloudTrailEventID は CloudTrail イベント ID
	CloudTrailEventID string `json:"cloudtrail_event_id,omitempty"`

	// EventName は CloudTrail イベント名
	EventName string `json:"event_name,omitempty"`

	// UserIdentity は変更を行ったユーザー
	UserIdentity string `json:"user_identity,omitempty"`

	// UserARN はユーザーの ARN
	UserARN string `json:"user_arn,omitempty"`

	// SourceIP は変更元の IP アドレス
	SourceIP string `json:"source_ip,omitempty"`

	// Timestamp はイベント発生時刻
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// Severity は drift の深刻度を表す
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// ImpactAnalysisResult はインパクト分析の結果を表す
type ImpactAnalysisResult struct {
	// DriftEventID は対象の drift イベント ID
	DriftEventID string `json:"drift_event_id"`

	// AffectedResourceCount は影響を受けるリソース数
	AffectedResourceCount int `json:"affected_resource_count"`

	// AffectedResources は影響を受けるリソースの詳細
	AffectedResources []AffectedResource `json:"affected_resources"`

	// BlastRadius は影響範囲 (グラフ内のホップ数)
	BlastRadius int `json:"blast_radius"`

	// Recommendations は推奨アクション
	Recommendations []string `json:"recommendations"`

	// Severity は全体の深刻度
	Severity Severity `json:"severity"`
}

// AffectedResource は影響を受けるリソースを表す
type AffectedResource struct {
	// ResourceID はリソース ID
	ResourceID string `json:"resource_id"`

	// ResourceType はリソースタイプ
	ResourceType string `json:"resource_type"`

	// RelationType は関係タイプ (e.g., "network", "dependency")
	RelationType string `json:"relation_type"`

	// Distance はグラフ内の距離 (ホップ数)
	Distance int `json:"distance"`

	// ImpactDescription は影響の説明
	ImpactDescription string `json:"impact_description"`
}
