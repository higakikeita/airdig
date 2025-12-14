package tfdrift

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/higakikeita/airdig/deepdrift/pkg/types"
)

// TFDriftAdapter は TFDrift-Falco との連携を提供する
type TFDriftAdapter struct {
	tfdriftPath string
	configPath  string
}

// NewTFDriftAdapter は新しい TFDriftAdapter を作成
func NewTFDriftAdapter(tfdriftPath, configPath string) *TFDriftAdapter {
	return &TFDriftAdapter{
		tfdriftPath: tfdriftPath,
		configPath:  configPath,
	}
}

// DetectDrift は TFDrift を実行して drift イベントを検出
func (a *TFDriftAdapter) DetectDrift(ctx context.Context, stateFile string) ([]*types.DriftEvent, error) {
	// TFDrift を実行
	output, err := a.runTFDrift(ctx, stateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to run tfdrift: %w", err)
	}

	// TFDrift の出力を DeepDrift の DriftEvent に変換
	events, err := a.parseTFDriftOutput(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tfdrift output: %w", err)
	}

	return events, nil
}

// runTFDrift は TFDrift コマンドを実行
func (a *TFDriftAdapter) runTFDrift(ctx context.Context, stateFile string) ([]byte, error) {
	// TFDrift のパスを確認
	tfdriftBin := a.tfdriftPath
	if tfdriftBin == "" {
		// デフォルトは tfdrift-falco プロジェクトの bin/tfdrift
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		tfdriftBin = filepath.Join(homeDir, "tfdrift-falco", "bin", "tfdrift")
	}

	// TFDrift コマンドを構築
	args := []string{"detect"}
	if stateFile != "" {
		args = append(args, "--state", stateFile)
	}
	if a.configPath != "" {
		args = append(args, "--config", a.configPath)
	}
	args = append(args, "--output", "json")

	// コマンドを実行
	cmd := exec.CommandContext(ctx, tfdriftBin, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("tfdrift command failed: %w\nOutput: %s", err, string(output))
	}

	return output, nil
}

// parseTFDriftOutput は TFDrift の JSON 出力を DeepDrift の DriftEvent に変換
func (a *TFDriftAdapter) parseTFDriftOutput(output []byte) ([]*types.DriftEvent, error) {
	// TFDrift の出力形式（仮定）
	type TFDriftOutput struct {
		Drifts []struct {
			ResourceID   string                 `json:"resource_id"`
			ResourceType string                 `json:"resource_type"`
			Type         string                 `json:"type"` // "created", "modified", "deleted"
			Before       map[string]interface{} `json:"before"`
			After        map[string]interface{} `json:"after"`
			CloudTrail   *struct {
				EventID      string    `json:"event_id"`
				EventName    string    `json:"event_name"`
				UserIdentity string    `json:"user_identity"`
				UserARN      string    `json:"user_arn"`
				SourceIP     string    `json:"source_ip"`
				Timestamp    time.Time `json:"timestamp"`
			} `json:"cloudtrail,omitempty"`
		} `json:"drifts"`
	}

	var tfdriftOutput TFDriftOutput
	if err := json.Unmarshal(output, &tfdriftOutput); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tfdrift output: %w", err)
	}

	// DeepDrift の DriftEvent に変換
	events := make([]*types.DriftEvent, 0, len(tfdriftOutput.Drifts))
	for _, drift := range tfdriftOutput.Drifts {
		event := &types.DriftEvent{
			ID:           fmt.Sprintf("drift-%d", time.Now().UnixNano()),
			ResourceID:   drift.ResourceID,
			ResourceType: drift.ResourceType,
			Type:         types.DriftType(drift.Type),
			Timestamp:    time.Now(),
			Before:       drift.Before,
			After:        drift.After,
			Diff:         calculateDiff(drift.Before, drift.After),
			Severity:     calculateSeverity(drift.Type, drift.ResourceType),
		}

		// CloudTrail 情報がある場合は RootCause を設定
		if drift.CloudTrail != nil {
			event.RootCause = &types.RootCause{
				CloudTrailEventID: drift.CloudTrail.EventID,
				EventName:         drift.CloudTrail.EventName,
				UserIdentity:      drift.CloudTrail.UserIdentity,
				UserARN:           drift.CloudTrail.UserARN,
				SourceIP:          drift.CloudTrail.SourceIP,
				Timestamp:         drift.CloudTrail.Timestamp,
			}
		}

		events = append(events, event)
	}

	return events, nil
}

// calculateDiff は Before と After の差分を計算
func calculateDiff(before, after map[string]interface{}) map[string]interface{} {
	diff := make(map[string]interface{})

	// Before にあって After にないキー（削除）
	for key, beforeVal := range before {
		if afterVal, exists := after[key]; !exists {
			diff[key] = map[string]interface{}{
				"type":   "deleted",
				"before": beforeVal,
			}
		} else if !equal(beforeVal, afterVal) {
			diff[key] = map[string]interface{}{
				"type":   "modified",
				"before": beforeVal,
				"after":  afterVal,
			}
		}
	}

	// After にあって Before にないキー（追加）
	for key, afterVal := range after {
		if _, exists := before[key]; !exists {
			diff[key] = map[string]interface{}{
				"type":  "added",
				"after": afterVal,
			}
		}
	}

	return diff
}

// equal は2つの値が等しいかを判定（簡易版）
func equal(a, b interface{}) bool {
	// TODO: より詳細な比較ロジックを実装
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	return string(aJSON) == string(bJSON)
}

// calculateSeverity は drift の深刻度を計算
func calculateSeverity(driftType, resourceType string) types.Severity {
	// セキュリティ関連リソースの変更は高い深刻度
	securityResources := map[string]bool{
		"security_group": true,
		"iam_role":       true,
		"iam_policy":     true,
		"kms_key":        true,
	}

	if securityResources[resourceType] {
		if driftType == string(types.DriftDeleted) {
			return types.SeverityCritical
		}
		return types.SeverityHigh
	}

	// 削除は一般的に高い深刻度
	if driftType == string(types.DriftDeleted) {
		return types.SeverityHigh
	}

	// 変更は中程度
	if driftType == string(types.DriftModified) {
		return types.SeverityMedium
	}

	// 作成は低い深刻度
	return types.SeverityLow
}

// WatchDrift は継続的に drift を監視（デーモンモード）
func (a *TFDriftAdapter) WatchDrift(ctx context.Context, stateFile string, interval time.Duration, callback func([]*types.DriftEvent) error) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			events, err := a.DetectDrift(ctx, stateFile)
			if err != nil {
				// エラーをログに記録して継続
				fmt.Fprintf(os.Stderr, "drift detection failed: %v\n", err)
				continue
			}

			if len(events) > 0 {
				if err := callback(events); err != nil {
					fmt.Fprintf(os.Stderr, "callback failed: %v\n", err)
				}
			}
		}
	}
}
