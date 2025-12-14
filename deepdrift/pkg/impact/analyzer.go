package impact

import (
	"fmt"

	"github.com/higakikeita/airdig/deepdrift/pkg/types"
	"github.com/higakikeita/airdig/skygraph/pkg/graph"
)

// Analyzer は drift のインパクトを分析する
type Analyzer struct {
	graph *graph.Graph
}

// NewAnalyzer は新しい Analyzer を作成
func NewAnalyzer(g *graph.Graph) *Analyzer {
	return &Analyzer{
		graph: g,
	}
}

// AnalyzeImpact は drift イベントのインパクトを分析
func (a *Analyzer) AnalyzeImpact(event *types.DriftEvent) (*types.ImpactAnalysisResult, error) {
	// グラフからリソースノードを検索
	node := a.graph.FindNode(event.ResourceID)
	if node == nil {
		// ノードが見つからない場合は影響なし
		return &types.ImpactAnalysisResult{
			DriftEventID:          event.ID,
			AffectedResourceCount: 0,
			AffectedResources:     []types.AffectedResource{},
			BlastRadius:           0,
			Recommendations:       []string{"Resource not found in graph. May be a new resource."},
			Severity:              types.SeverityLow,
		}, nil
	}

	// 影響を受けるリソースを探索
	affectedResources := a.findAffectedResources(node, event.Type)

	// 推奨アクションを生成
	recommendations := a.generateRecommendations(event, affectedResources)

	// 全体の深刻度を計算
	severity := a.calculateOverallSeverity(event, affectedResources)

	return &types.ImpactAnalysisResult{
		DriftEventID:          event.ID,
		AffectedResourceCount: len(affectedResources),
		AffectedResources:     affectedResources,
		BlastRadius:           a.calculateBlastRadius(affectedResources),
		Recommendations:       recommendations,
		Severity:              severity,
	}, nil
}

// findAffectedResources は影響を受けるリソースを BFS で探索
func (a *Analyzer) findAffectedResources(startNode *graph.ResourceNode, driftType types.DriftType) []types.AffectedResource {
	affected := []types.AffectedResource{}
	visited := make(map[string]bool)
	queue := []struct {
		node     *graph.ResourceNode
		distance int
		relType  string
	}{
		{startNode, 0, "self"},
	}

	// BFS で探索（最大3ホップまで）
	maxDistance := 3
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current.node.ID] {
			continue
		}
		visited[current.node.ID] = true

		// 自分自身は除外
		if current.distance > 0 {
			impact := types.AffectedResource{
				ResourceID:        current.node.ID,
				ResourceType:      current.node.Type,
				RelationType:      current.relType,
				Distance:          current.distance,
				ImpactDescription: a.generateImpactDescription(current.node, driftType, current.relType),
			}
			affected = append(affected, impact)
		}

		// 最大距離に達したら探索を停止
		if current.distance >= maxDistance {
			continue
		}

		// 隣接ノードを探索
		edges := a.graph.FindEdges(current.node.ID)
		for _, edge := range edges {
			var nextNodeID string
			var relType string

			if edge.From == current.node.ID {
				nextNodeID = edge.To
				relType = edge.Type
			} else {
				nextNodeID = edge.From
				relType = edge.Type
			}

			nextNode := a.graph.FindNode(nextNodeID)
			if nextNode != nil && !visited[nextNodeID] {
				queue = append(queue, struct {
					node     *graph.ResourceNode
					distance int
					relType  string
				}{nextNode, current.distance + 1, relType})
			}
		}
	}

	return affected
}

// generateImpactDescription は影響の説明を生成
func (a *Analyzer) generateImpactDescription(node *graph.ResourceNode, driftType types.DriftType, relType string) string {
	switch relType {
	case "network":
		return fmt.Sprintf("Network connectivity may be affected")
	case "dependency":
		return fmt.Sprintf("Dependent resource may experience issues")
	case "ownership":
		return fmt.Sprintf("Parent-child relationship affected")
	default:
		return fmt.Sprintf("May be impacted by drift")
	}
}

// generateRecommendations は推奨アクションを生成
func (a *Analyzer) generateRecommendations(event *types.DriftEvent, affected []types.AffectedResource) []string {
	recommendations := []string{}

	switch event.Type {
	case types.DriftDeleted:
		recommendations = append(recommendations, "Review if this resource deletion was intentional")
		if len(affected) > 0 {
			recommendations = append(recommendations, fmt.Sprintf("Check %d dependent resources for potential issues", len(affected)))
		}
		recommendations = append(recommendations, "Update Terraform state if deletion is permanent")

	case types.DriftModified:
		recommendations = append(recommendations, "Verify configuration changes against security policies")
		if event.ResourceType == "security_group" {
			recommendations = append(recommendations, "Review security group rules for potential vulnerabilities")
		}
		recommendations = append(recommendations, "Apply changes to Terraform code or revert via terraform apply")

	case types.DriftCreated:
		recommendations = append(recommendations, "Import resource into Terraform state if needed")
		recommendations = append(recommendations, "Document the reason for manual resource creation")
	}

	return recommendations
}

// calculateBlastRadius は影響範囲（最大距離）を計算
func (a *Analyzer) calculateBlastRadius(affected []types.AffectedResource) int {
	maxDistance := 0
	for _, resource := range affected {
		if resource.Distance > maxDistance {
			maxDistance = resource.Distance
		}
	}
	return maxDistance
}

// calculateOverallSeverity は全体の深刻度を計算
func (a *Analyzer) calculateOverallSeverity(event *types.DriftEvent, affected []types.AffectedResource) types.Severity {
	// イベント自体の深刻度から開始
	severity := event.Severity

	// 影響を受けるリソース数に応じて深刻度を調整
	affectedCount := len(affected)
	if affectedCount > 10 {
		// 多くのリソースが影響を受ける場合は深刻度を上げる
		if severity == types.SeverityLow {
			severity = types.SeverityMedium
		} else if severity == types.SeverityMedium {
			severity = types.SeverityHigh
		} else if severity == types.SeverityHigh {
			severity = types.SeverityCritical
		}
	}

	// セキュリティ関連リソースが影響を受ける場合は深刻度を上げる
	for _, resource := range affected {
		if isSecurityResource(resource.ResourceType) {
			if severity == types.SeverityLow {
				severity = types.SeverityMedium
			} else if severity == types.SeverityMedium {
				severity = types.SeverityHigh
			}
			break
		}
	}

	return severity
}

// isSecurityResource はセキュリティ関連リソースかを判定
func isSecurityResource(resourceType string) bool {
	securityResources := map[string]bool{
		"security_group": true,
		"iam_role":       true,
		"iam_policy":     true,
		"kms_key":        true,
	}
	return securityResources[resourceType]
}

// AnalyzeBatch は複数の drift イベントのインパクトを一括分析
func (a *Analyzer) AnalyzeBatch(events []*types.DriftEvent) ([]*types.ImpactAnalysisResult, error) {
	results := make([]*types.ImpactAnalysisResult, 0, len(events))

	for _, event := range events {
		result, err := a.AnalyzeImpact(event)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze event %s: %w", event.ID, err)
		}
		results = append(results, result)
	}

	return results, nil
}
