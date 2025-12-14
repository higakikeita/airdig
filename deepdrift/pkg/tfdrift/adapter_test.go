package tfdrift

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/higakikeita/airdig/deepdrift/pkg/types"
)

func TestParseTFDriftOutput(t *testing.T) {
	adapter := NewTFDriftAdapter("", "")

	// Mock TFDrift JSON output
	mockOutput := `{
		"drifts": [
			{
				"resource_id": "aws:ec2:i-123456",
				"resource_type": "ec2",
				"type": "modified",
				"before": {
					"instance_type": "t3.micro",
					"tags": {"Name": "web-server"}
				},
				"after": {
					"instance_type": "t3.large",
					"tags": {"Name": "web-server"}
				},
				"cloudtrail": {
					"event_id": "evt-12345",
					"event_name": "ModifyInstanceAttribute",
					"user_identity": "alice@example.com",
					"user_arn": "arn:aws:iam::123456789012:user/alice",
					"source_ip": "203.0.113.5",
					"timestamp": "2025-12-14T10:00:00Z"
				}
			},
			{
				"resource_id": "aws:sg:sg-789012",
				"resource_type": "security_group",
				"type": "deleted",
				"before": {
					"vpc_id": "vpc-123",
					"ingress": [{"protocol": "tcp", "port": 80}]
				},
				"after": null,
				"cloudtrail": {
					"event_id": "evt-67890",
					"event_name": "DeleteSecurityGroup",
					"user_identity": "bob@example.com",
					"user_arn": "arn:aws:iam::123456789012:user/bob",
					"source_ip": "198.51.100.42",
					"timestamp": "2025-12-14T10:05:00Z"
				}
			},
			{
				"resource_id": "aws:s3:my-new-bucket",
				"resource_type": "s3",
				"type": "created",
				"before": null,
				"after": {
					"versioning": true,
					"encryption": "AES256"
				}
			}
		]
	}`

	events, err := adapter.parseTFDriftOutput([]byte(mockOutput))
	if err != nil {
		t.Fatalf("parseTFDriftOutput failed: %v", err)
	}

	if len(events) != 3 {
		t.Fatalf("Expected 3 events, got %d", len(events))
	}

	// Test first event (modified EC2)
	e1 := events[0]
	if e1.ResourceID != "aws:ec2:i-123456" {
		t.Errorf("Expected resource ID 'aws:ec2:i-123456', got '%s'", e1.ResourceID)
	}
	if e1.ResourceType != "ec2" {
		t.Errorf("Expected resource type 'ec2', got '%s'", e1.ResourceType)
	}
	if e1.Type != types.DriftModified {
		t.Errorf("Expected type 'modified', got '%s'", e1.Type)
	}
	if e1.Severity != types.SeverityMedium {
		t.Errorf("Expected severity 'medium', got '%s'", e1.Severity)
	}
	if e1.RootCause == nil {
		t.Error("Expected RootCause to be set")
	} else {
		if e1.RootCause.EventName != "ModifyInstanceAttribute" {
			t.Errorf("Expected event name 'ModifyInstanceAttribute', got '%s'", e1.RootCause.EventName)
		}
		if e1.RootCause.UserIdentity != "alice@example.com" {
			t.Errorf("Expected user 'alice@example.com', got '%s'", e1.RootCause.UserIdentity)
		}
	}

	// Test second event (deleted security group)
	e2 := events[1]
	if e2.ResourceID != "aws:sg:sg-789012" {
		t.Errorf("Expected resource ID 'aws:sg:sg-789012', got '%s'", e2.ResourceID)
	}
	if e2.Type != types.DriftDeleted {
		t.Errorf("Expected type 'deleted', got '%s'", e2.Type)
	}
	if e2.Severity != types.SeverityCritical {
		t.Errorf("Expected severity 'critical' for deleted security group, got '%s'", e2.Severity)
	}

	// Test third event (created S3)
	e3 := events[2]
	if e3.ResourceID != "aws:s3:my-new-bucket" {
		t.Errorf("Expected resource ID 'aws:s3:my-new-bucket', got '%s'", e3.ResourceID)
	}
	if e3.Type != types.DriftCreated {
		t.Errorf("Expected type 'created', got '%s'", e3.Type)
	}
	if e3.Severity != types.SeverityLow {
		t.Errorf("Expected severity 'low' for created resource, got '%s'", e3.Severity)
	}
	if e3.RootCause != nil {
		t.Error("Expected RootCause to be nil for resource without CloudTrail")
	}

	t.Logf("Successfully parsed %d drift events", len(events))
}

func TestCalculateSeverity(t *testing.T) {
	tests := []struct {
		name         string
		driftType    string
		resourceType string
		expected     types.Severity
	}{
		{
			name:         "deleted security group",
			driftType:    string(types.DriftDeleted),
			resourceType: "security_group",
			expected:     types.SeverityCritical,
		},
		{
			name:         "modified security group",
			driftType:    string(types.DriftModified),
			resourceType: "security_group",
			expected:     types.SeverityHigh,
		},
		{
			name:         "deleted IAM role",
			driftType:    string(types.DriftDeleted),
			resourceType: "iam_role",
			expected:     types.SeverityCritical,
		},
		{
			name:         "modified IAM policy",
			driftType:    string(types.DriftModified),
			resourceType: "iam_policy",
			expected:     types.SeverityHigh,
		},
		{
			name:         "deleted EC2",
			driftType:    string(types.DriftDeleted),
			resourceType: "ec2",
			expected:     types.SeverityHigh,
		},
		{
			name:         "modified EC2",
			driftType:    string(types.DriftModified),
			resourceType: "ec2",
			expected:     types.SeverityMedium,
		},
		{
			name:         "created S3 bucket",
			driftType:    string(types.DriftCreated),
			resourceType: "s3",
			expected:     types.SeverityLow,
		},
		{
			name:         "created VPC",
			driftType:    string(types.DriftCreated),
			resourceType: "vpc",
			expected:     types.SeverityLow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateSeverity(tt.driftType, tt.resourceType)
			if result != tt.expected {
				t.Errorf("calculateSeverity(%s, %s) = %s; want %s",
					tt.driftType, tt.resourceType, result, tt.expected)
			}
		})
	}
}

func TestCalculateDiff(t *testing.T) {
	before := map[string]interface{}{
		"instance_type": "t3.micro",
		"monitoring":    false,
		"tags": map[string]interface{}{
			"Name": "web-server",
			"Env":  "dev",
		},
	}

	after := map[string]interface{}{
		"instance_type": "t3.large",
		"monitoring":    true,
		"tags": map[string]interface{}{
			"Name": "web-server",
			"Env":  "prod",
		},
		"backup": true,
	}

	diff := calculateDiff(before, after)

	// instance_type が変更されたことを確認
	if _, exists := diff["instance_type"]; !exists {
		t.Error("Expected 'instance_type' in diff")
	} else {
		diffEntry := diff["instance_type"].(map[string]interface{})
		if diffEntry["type"] != "modified" {
			t.Errorf("Expected type 'modified', got '%v'", diffEntry["type"])
		}
	}

	// monitoring が変更されたことを確認
	if _, exists := diff["monitoring"]; !exists {
		t.Error("Expected 'monitoring' in diff")
	}

	// tags が変更されたことを確認
	if _, exists := diff["tags"]; !exists {
		t.Error("Expected 'tags' in diff")
	}

	// backup が追加されたことを確認
	if _, exists := diff["backup"]; !exists {
		t.Error("Expected 'backup' in diff")
	} else {
		diffEntry := diff["backup"].(map[string]interface{})
		if diffEntry["type"] != "added" {
			t.Errorf("Expected type 'added', got '%v'", diffEntry["type"])
		}
	}

	t.Logf("Diff calculated: %d changes", len(diff))
}

func TestEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected bool
	}{
		{
			name:     "equal strings",
			a:        "hello",
			b:        "hello",
			expected: true,
		},
		{
			name:     "different strings",
			a:        "hello",
			b:        "world",
			expected: false,
		},
		{
			name:     "equal numbers",
			a:        42,
			b:        42,
			expected: true,
		},
		{
			name:     "different numbers",
			a:        42,
			b:        43,
			expected: false,
		},
		{
			name:     "equal booleans",
			a:        true,
			b:        true,
			expected: true,
		},
		{
			name: "equal maps",
			a: map[string]interface{}{
				"key": "value",
			},
			b: map[string]interface{}{
				"key": "value",
			},
			expected: true,
		},
		{
			name: "different maps",
			a: map[string]interface{}{
				"key": "value1",
			},
			b: map[string]interface{}{
				"key": "value2",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := equal(tt.a, tt.b)
			if result != tt.expected {
				aJSON, _ := json.Marshal(tt.a)
				bJSON, _ := json.Marshal(tt.b)
				t.Errorf("equal(%s, %s) = %v; want %v",
					string(aJSON), string(bJSON), result, tt.expected)
			}
		})
	}
}

func TestNewTFDriftAdapter(t *testing.T) {
	adapter := NewTFDriftAdapter("/path/to/tfdrift", "/path/to/config")

	if adapter.tfdriftPath != "/path/to/tfdrift" {
		t.Errorf("Expected tfdriftPath '/path/to/tfdrift', got '%s'", adapter.tfdriftPath)
	}

	if adapter.configPath != "/path/to/config" {
		t.Errorf("Expected configPath '/path/to/config', got '%s'", adapter.configPath)
	}
}

func TestParseTFDriftOutput_EmptyDrifts(t *testing.T) {
	adapter := NewTFDriftAdapter("", "")

	mockOutput := `{"drifts": []}`

	events, err := adapter.parseTFDriftOutput([]byte(mockOutput))
	if err != nil {
		t.Fatalf("parseTFDriftOutput failed: %v", err)
	}

	if len(events) != 0 {
		t.Errorf("Expected 0 events for empty drifts, got %d", len(events))
	}
}

func TestParseTFDriftOutput_InvalidJSON(t *testing.T) {
	adapter := NewTFDriftAdapter("", "")

	mockOutput := `{invalid json`

	_, err := adapter.parseTFDriftOutput([]byte(mockOutput))
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestParseTFDriftOutput_CloudTrailTimestamp(t *testing.T) {
	adapter := NewTFDriftAdapter("", "")

	mockOutput := `{
		"drifts": [
			{
				"resource_id": "aws:ec2:i-123",
				"resource_type": "ec2",
				"type": "modified",
				"before": {},
				"after": {},
				"cloudtrail": {
					"event_id": "evt-123",
					"event_name": "ModifyInstance",
					"user_identity": "test@example.com",
					"user_arn": "arn:aws:iam::123:user/test",
					"source_ip": "1.2.3.4",
					"timestamp": "2025-12-14T10:30:45Z"
				}
			}
		]
	}`

	events, err := adapter.parseTFDriftOutput([]byte(mockOutput))
	if err != nil {
		t.Fatalf("parseTFDriftOutput failed: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}

	if events[0].RootCause == nil {
		t.Fatal("Expected RootCause to be set")
	}

	expectedTime, _ := time.Parse(time.RFC3339, "2025-12-14T10:30:45Z")
	if !events[0].RootCause.Timestamp.Equal(expectedTime) {
		t.Errorf("Expected timestamp %v, got %v",
			expectedTime, events[0].RootCause.Timestamp)
	}
}
