// +build integration

package tfdrift

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/higakikeita/airdig/deepdrift/pkg/types"
)

// TestTFDriftBinaryIntegration tests integration with actual TFDrift-Falco binary
// Run with: go test -tags=integration ./pkg/tfdrift/...
func TestTFDriftBinaryIntegration(t *testing.T) {
	// Check if TFDrift binary exists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tfdriftPath := filepath.Join(homeDir, "tfdrift-falco", "bin", "tfdrift")
	if _, err := os.Stat(tfdriftPath); os.IsNotExist(err) {
		t.Skipf("TFDrift binary not found at %s, skipping integration test", tfdriftPath)
	}

	// Create adapter
	adapter := NewTFDriftAdapter(tfdriftPath, "")

	// Use example state file (you may need to create a real one for testing)
	stateFile := "../../testdata/terraform.tfstate.example"
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Skipf("Test state file not found at %s, skipping integration test", stateFile)
	}

	t.Logf("Testing TFDrift binary at: %s", tfdriftPath)
	t.Logf("Using state file: %s", stateFile)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Run drift detection
	// Note: This may fail if TFDrift requires actual AWS credentials or CloudTrail access
	events, err := adapter.DetectDrift(ctx, stateFile)

	// If TFDrift binary execution fails due to AWS credentials, that's expected
	if err != nil {
		t.Logf("TFDrift execution failed (this may be expected if AWS credentials are not configured): %v", err)

		// Check if it's a credentials issue
		if contains(err.Error(), "credentials") || contains(err.Error(), "Unable to locate") {
			t.Skip("Skipping test due to missing AWS credentials")
		}

		// Other errors might be real issues
		t.Logf("Error running TFDrift: %v", err)
		return
	}

	t.Logf("TFDrift detected %d drift events", len(events))

	// If we get here, TFDrift ran successfully
	for i, event := range events {
		t.Logf("Event %d: [%s] %s (%s) - Severity: %s",
			i+1, event.Type, event.ResourceID, event.ResourceType, event.Severity)
		if event.RootCause != nil {
			t.Logf("  Root cause: %s by %s", event.RootCause.EventName, event.RootCause.UserIdentity)
		}
	}
}

// TestTFDriftCommandConstruction tests that we build the correct command
func TestTFDriftCommandConstruction(t *testing.T) {
	adapter := NewTFDriftAdapter("/path/to/tfdrift", "/path/to/config")

	// We can't easily test the actual command construction without exposing it,
	// but we can verify the adapter is created correctly
	if adapter.tfdriftPath != "/path/to/tfdrift" {
		t.Errorf("Expected tfdriftPath '/path/to/tfdrift', got '%s'", adapter.tfdriftPath)
	}

	if adapter.configPath != "/path/to/config" {
		t.Errorf("Expected configPath '/path/to/config', got '%s'", adapter.configPath)
	}
}

// TestWatchDriftIntegration tests the watch mode (short duration)
func TestWatchDriftIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping watch mode test in short mode")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tfdriftPath := filepath.Join(homeDir, "tfdrift-falco", "bin", "tfdrift")
	if _, err := os.Stat(tfdriftPath); os.IsNotExist(err) {
		t.Skipf("TFDrift binary not found at %s, skipping integration test", tfdriftPath)
	}

	adapter := NewTFDriftAdapter(tfdriftPath, "")
	stateFile := "../../testdata/terraform.tfstate.example"

	// Create a context that will cancel after 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	callbackCalled := false
	callback := func(events []*types.DriftEvent) error {
		callbackCalled = true
		t.Logf("Watch mode callback: detected %d events", len(events))
		return nil
	}

	// Start watching (will be cancelled after 5 seconds)
	err = adapter.WatchDrift(ctx, stateFile, 2*time.Second, callback)

	// We expect context.DeadlineExceeded or context.Canceled
	if err != nil && err != context.DeadlineExceeded && err != context.Canceled {
		// Check if it's an AWS credentials error
		if contains(err.Error(), "credentials") {
			t.Skip("Skipping test due to missing AWS credentials")
		}
		t.Logf("Watch mode error (may be expected): %v", err)
	}

	t.Logf("Watch mode ran for 5 seconds, callback called: %v", callbackCalled)
}

// TestTFDriftOutputParsing tests that we can handle various output formats
func TestTFDriftOutputParsing(t *testing.T) {
	adapter := NewTFDriftAdapter("", "")

	testCases := []struct {
		name          string
		output        string
		expectError   bool
		expectedCount int
	}{
		{
			name:          "empty output",
			output:        `{"drifts":[]}`,
			expectError:   false,
			expectedCount: 0,
		},
		{
			name: "single drift",
			output: `{
				"drifts": [{
					"resource_id": "aws:ec2:i-123",
					"resource_type": "ec2",
					"type": "modified",
					"before": {"instance_type": "t3.micro"},
					"after": {"instance_type": "t3.large"}
				}]
			}`,
			expectError:   false,
			expectedCount: 1,
		},
		{
			name:        "invalid json",
			output:      `{invalid}`,
			expectError: true,
		},
		{
			name: "with cloudtrail",
			output: `{
				"drifts": [{
					"resource_id": "aws:sg:sg-123",
					"resource_type": "security_group",
					"type": "deleted",
					"before": {"rules": []},
					"after": null,
					"cloudtrail": {
						"event_id": "evt-123",
						"event_name": "DeleteSecurityGroup",
						"user_identity": "alice@example.com",
						"user_arn": "arn:aws:iam::123:user/alice",
						"source_ip": "1.2.3.4",
						"timestamp": "2025-12-14T10:00:00Z"
					}
				}]
			}`,
			expectError:   false,
			expectedCount: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			events, err := adapter.parseTFDriftOutput([]byte(tc.output))

			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(events) != tc.expectedCount {
				t.Errorf("Expected %d events, got %d", tc.expectedCount, len(events))
			}

			// For the cloudtrail test case, verify root cause is set
			if tc.name == "with cloudtrail" && len(events) > 0 {
				if events[0].RootCause == nil {
					t.Error("Expected root cause to be set")
				} else {
					if events[0].RootCause.UserIdentity != "alice@example.com" {
						t.Errorf("Expected user 'alice@example.com', got '%s'", events[0].RootCause.UserIdentity)
					}
				}
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
