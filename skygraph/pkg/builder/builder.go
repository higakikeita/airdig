package builder

import (
	"fmt"

	"github.com/yourusername/airdig/skygraph/pkg/graph"
)

// GraphBuilder はスキャン結果からグラフを構築
type GraphBuilder struct {
	graph *graph.Graph
}

// NewGraphBuilder は新しい GraphBuilder を作成
func NewGraphBuilder() *GraphBuilder {
	return &GraphBuilder{
		graph: graph.NewGraph(),
	}
}

// AddNodes はノードをグラフに追加（重複排除）
func (b *GraphBuilder) AddNodes(nodes []graph.ResourceNode) {
	for _, node := range nodes {
		// 既存チェック
		if b.graph.FindNode(node.ID) == nil {
			b.graph.AddNode(node)
		}
	}
}

// InferEdges は全ノードからエッジを推論
func (b *GraphBuilder) InferEdges() error {
	for _, node := range b.graph.Nodes {
		edges, err := b.inferEdgesForNode(node)
		if err != nil {
			return fmt.Errorf("failed to infer edges for %s: %w", node.ID, err)
		}

		for _, edge := range edges {
			// エッジの to ノードが存在するか確認
			if b.graph.FindNode(edge.To) != nil {
				b.graph.AddEdge(edge)
			}
		}
	}

	return nil
}

// inferEdgesForNode は1つのノードに対してエッジを推論
func (b *GraphBuilder) inferEdgesForNode(node graph.ResourceNode) ([]graph.Edge, error) {
	edges := make([]graph.Edge, 0)

	switch node.Type {
	case "vpc":
		// VPC に対するエッジは他のノードから生成されるため、ここでは何もしない

	case "subnet":
		// Subnet → VPC (ownership)
		if vpcID, ok := node.Metadata["vpc_id"].(string); ok && vpcID != "" {
			edges = append(edges, graph.Edge{
				From: fmt.Sprintf("aws:vpc:%s", vpcID),
				To:   node.ID,
				Type: "ownership",
			})
		}

	case "security_group":
		// Security Group → VPC (ownership)
		if vpcID, ok := node.Metadata["vpc_id"].(string); ok && vpcID != "" {
			edges = append(edges, graph.Edge{
				From: fmt.Sprintf("aws:vpc:%s", vpcID),
				To:   node.ID,
				Type: "ownership",
			})
		}

	case "ec2":
		// EC2 → Subnet (network)
		if subnetID, ok := node.Metadata["subnet_id"].(string); ok && subnetID != "" {
			edges = append(edges, graph.Edge{
				From: fmt.Sprintf("aws:subnet:%s", subnetID),
				To:   node.ID,
				Type: "network",
			})
		}

		// EC2 → Security Groups (network)
		if sgIDs, ok := node.Metadata["security_groups"].([]string); ok {
			for _, sgID := range sgIDs {
				edges = append(edges, graph.Edge{
					From: fmt.Sprintf("aws:sg:%s", sgID),
					To:   node.ID,
					Type: "network",
				})
			}
		}

	case "rds":
		// RDS → Subnet (network)
		if subnetIDs, ok := node.Metadata["subnet_ids"].([]string); ok {
			for _, subnetID := range subnetIDs {
				edges = append(edges, graph.Edge{
					From: fmt.Sprintf("aws:subnet:%s", subnetID),
					To:   node.ID,
					Type: "network",
				})
			}
		}

		// RDS → Security Groups (network)
		if sgIDs, ok := node.Metadata["security_groups"].([]string); ok {
			for _, sgID := range sgIDs {
				edges = append(edges, graph.Edge{
					From: fmt.Sprintf("aws:sg:%s", sgID),
					To:   node.ID,
					Type: "network",
				})
			}
		}

		// EC2 → RDS (dependency) の推論
		// 同じ VPC 内の EC2 は RDS に依存している可能性がある
		if vpcID, ok := node.Metadata["vpc_id"].(string); ok && vpcID != "" {
			for _, n := range b.graph.Nodes {
				if n.Type == "ec2" {
					if ec2VpcID, ok := n.Metadata["vpc_id"].(string); ok && ec2VpcID == vpcID {
						edges = append(edges, graph.Edge{
							From: n.ID,
							To:   node.ID,
							Type: "dependency",
							Metadata: map[string]interface{}{
								"inferred": true,
								"reason":   "same VPC",
							},
						})
					}
				}
			}
		}
	}

	return edges, nil
}

// Build はグラフを完成させて返す
func (b *GraphBuilder) Build() *graph.Graph {
	return b.graph
}

// GetGraph はグラフを返す
func (b *GraphBuilder) GetGraph() *graph.Graph {
	return b.graph
}
