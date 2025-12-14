package impact

import (
	"fmt"
	"testing"
	"time"

	"github.com/higakikeita/airdig/deepdrift/pkg/types"
	"github.com/higakikeita/airdig/skygraph/pkg/graph"
)

// createTestGraph creates a test graph with typical AWS resources
func createTestGraph() *graph.Graph {
	g := graph.NewGraph()

	// VPC
	vpc := graph.ResourceNode{
		ID:       "aws:vpc:vpc-123",
		Type:     "vpc",
		Provider: "aws",
		Region:   "us-west-2",
		Name:     "main-vpc",
		Metadata: map[string]interface{}{
			"cidr": "10.0.0.0/16",
		},
	}
	g.AddNode(vpc)

	// Subnets
	subnet1 := graph.ResourceNode{
		ID:       "aws:subnet:subnet-111",
		Type:     "subnet",
		Provider: "aws",
		Region:   "us-west-2",
		Name:     "public-subnet-1",
		Metadata: map[string]interface{}{
			"cidr": "10.0.1.0/24",
		},
	}
	g.AddNode(subnet1)

	subnet2 := graph.ResourceNode{
		ID:       "aws:subnet:subnet-222",
		Type:     "subnet",
		Provider: "aws",
		Region:   "us-west-2",
		Name:     "private-subnet-1",
		Metadata: map[string]interface{}{
			"cidr": "10.0.2.0/24",
		},
	}
	g.AddNode(subnet2)

	// Security Group
	sg := graph.ResourceNode{
		ID:       "aws:sg:sg-789",
		Type:     "security_group",
		Provider: "aws",
		Region:   "us-west-2",
		Name:     "web-sg",
		Metadata: map[string]interface{}{
			"vpc_id": "vpc-123",
		},
	}
	g.AddNode(sg)

	// EC2 Instances
	ec2_1 := graph.ResourceNode{
		ID:       "aws:ec2:i-111",
		Type:     "ec2",
		Provider: "aws",
		Region:   "us-west-2",
		Name:     "web-server-1",
		Metadata: map[string]interface{}{
			"instance_type": "t3.micro",
		},
	}
	g.AddNode(ec2_1)

	ec2_2 := graph.ResourceNode{
		ID:       "aws:ec2:i-222",
		Type:     "ec2",
		Provider: "aws",
		Region:   "us-west-2",
		Name:     "web-server-2",
		Metadata: map[string]interface{}{
			"instance_type": "t3.micro",
		},
	}
	g.AddNode(ec2_2)

	// RDS
	rds := graph.ResourceNode{
		ID:       "aws:rds:db-123",
		Type:     "rds",
		Provider: "aws",
		Region:   "us-west-2",
		Name:     "mysql-db",
		Metadata: map[string]interface{}{
			"engine": "mysql",
		},
	}
	g.AddNode(rds)

	// Lambda
	lambda := graph.ResourceNode{
		ID:       "aws:lambda:func-456",
		Type:     "lambda",
		Provider: "aws",
		Region:   "us-west-2",
		Name:     "api-handler",
		Metadata: map[string]interface{}{
			"runtime": "go1.x",
		},
	}
	g.AddNode(lambda)

	// Add edges (relationships)
	// VPC -> Subnets (ownership)
	g.AddEdge(graph.Edge{From: vpc.ID, To: subnet1.ID, Type: "ownership"})
	g.AddEdge(graph.Edge{From: vpc.ID, To: subnet2.ID, Type: "ownership"})

	// VPC -> Security Group (ownership)
	g.AddEdge(graph.Edge{From: vpc.ID, To: sg.ID, Type: "ownership"})

	// EC2 -> Subnet (network)
	g.AddEdge(graph.Edge{From: ec2_1.ID, To: subnet1.ID, Type: "network"})
	g.AddEdge(graph.Edge{From: ec2_2.ID, To: subnet1.ID, Type: "network"})

	// EC2 -> Security Group (network)
	g.AddEdge(graph.Edge{From: ec2_1.ID, To: sg.ID, Type: "network"})
	g.AddEdge(graph.Edge{From: ec2_2.ID, To: sg.ID, Type: "network"})

	// EC2 -> RDS (dependency)
	g.AddEdge(graph.Edge{From: ec2_1.ID, To: rds.ID, Type: "dependency"})
	g.AddEdge(graph.Edge{From: ec2_2.ID, To: rds.ID, Type: "dependency"})

	// RDS -> Subnet (network)
	g.AddEdge(graph.Edge{From: rds.ID, To: subnet2.ID, Type: "network"})

	// Lambda -> Subnet (network)
	g.AddEdge(graph.Edge{From: lambda.ID, To: subnet2.ID, Type: "network"})

	// Lambda -> RDS (dependency)
	g.AddEdge(graph.Edge{From: lambda.ID, To: rds.ID, Type: "dependency"})

	return g
}

func TestAnalyzer_AnalyzeImpact_DeletedSecurityGroup(t *testing.T) {
	g := createTestGraph()
	analyzer := NewAnalyzer(g)

	event := &types.DriftEvent{
		ID:           "drift-001",
		ResourceID:   "aws:sg:sg-789",
		ResourceType: "security_group",
		Type:         types.DriftDeleted,
		Timestamp:    time.Now(),
		Severity:     types.SeverityCritical,
	}

	result, err := analyzer.AnalyzeImpact(event)
	if err != nil {
		t.Fatalf("AnalyzeImpact failed: %v", err)
	}

	// セキュリティグループが削除された場合、VPC、EC2、Subnet が影響を受ける
	if result.AffectedResourceCount < 3 {
		t.Errorf("Expected at least 3 affected resources, got %d", result.AffectedResourceCount)
	}

	// セキュリティグループの削除は critical
	if result.Severity != types.SeverityCritical {
		t.Errorf("Expected severity critical, got %s", result.Severity)
	}

	// 推奨アクションが含まれているか確認
	if len(result.Recommendations) == 0 {
		t.Error("Expected recommendations, got none")
	}

	t.Logf("Affected resources: %d", result.AffectedResourceCount)
	t.Logf("Blast radius: %d hops", result.BlastRadius)
	t.Logf("Severity: %s", result.Severity)
	for _, rec := range result.Recommendations {
		t.Logf("  - %s", rec)
	}
}

func TestAnalyzer_AnalyzeImpact_ModifiedEC2(t *testing.T) {
	g := createTestGraph()
	analyzer := NewAnalyzer(g)

	event := &types.DriftEvent{
		ID:           "drift-002",
		ResourceID:   "aws:ec2:i-111",
		ResourceType: "ec2",
		Type:         types.DriftModified,
		Timestamp:    time.Now(),
		Severity:     types.SeverityMedium,
		Before: map[string]interface{}{
			"instance_type": "t3.micro",
		},
		After: map[string]interface{}{
			"instance_type": "t3.large",
		},
	}

	result, err := analyzer.AnalyzeImpact(event)
	if err != nil {
		t.Fatalf("AnalyzeImpact failed: %v", err)
	}

	// EC2 の変更は Subnet、SG、RDS に影響
	if result.AffectedResourceCount < 3 {
		t.Errorf("Expected at least 3 affected resources, got %d", result.AffectedResourceCount)
	}

	// EC2 の変更は medium または high (影響を受けるリソースに security_group が含まれる場合は high)
	if result.Severity != types.SeverityMedium && result.Severity != types.SeverityHigh {
		t.Errorf("Expected severity medium or high, got %s", result.Severity)
	}

	t.Logf("Affected resources: %d", result.AffectedResourceCount)
	t.Logf("Blast radius: %d hops", result.BlastRadius)
}

func TestAnalyzer_AnalyzeImpact_DeletedVPC(t *testing.T) {
	g := createTestGraph()
	analyzer := NewAnalyzer(g)

	event := &types.DriftEvent{
		ID:           "drift-003",
		ResourceID:   "aws:vpc:vpc-123",
		ResourceType: "vpc",
		Type:         types.DriftDeleted,
		Timestamp:    time.Now(),
		Severity:     types.SeverityHigh,
	}

	result, err := analyzer.AnalyzeImpact(event)
	if err != nil {
		t.Fatalf("AnalyzeImpact failed: %v", err)
	}

	// VPC が削除されると全てのリソースが影響を受ける
	// Subnet x2, SG, EC2 x2, RDS, Lambda = 7
	if result.AffectedResourceCount < 5 {
		t.Errorf("Expected at least 5 affected resources, got %d", result.AffectedResourceCount)
	}

	// VPC 削除は high severity (多くのリソースに影響)
	if result.Severity != types.SeverityHigh && result.Severity != types.SeverityCritical {
		t.Errorf("Expected severity high or critical, got %s", result.Severity)
	}

	// Blast radius は大きいはず
	if result.BlastRadius < 2 {
		t.Errorf("Expected blast radius >= 2, got %d", result.BlastRadius)
	}

	t.Logf("Affected resources: %d", result.AffectedResourceCount)
	t.Logf("Blast radius: %d hops", result.BlastRadius)
	t.Logf("Severity: %s", result.Severity)
}

func TestAnalyzer_AnalyzeImpact_CreatedResource(t *testing.T) {
	g := createTestGraph()
	analyzer := NewAnalyzer(g)

	event := &types.DriftEvent{
		ID:           "drift-004",
		ResourceID:   "aws:s3:my-new-bucket",
		ResourceType: "s3",
		Type:         types.DriftCreated,
		Timestamp:    time.Now(),
		Severity:     types.SeverityLow,
	}

	result, err := analyzer.AnalyzeImpact(event)
	if err != nil {
		t.Fatalf("AnalyzeImpact failed: %v", err)
	}

	// グラフに存在しないリソースなので影響なし
	if result.AffectedResourceCount != 0 {
		t.Errorf("Expected 0 affected resources for new resource, got %d", result.AffectedResourceCount)
	}

	// 新規作成は low severity
	if result.Severity != types.SeverityLow {
		t.Errorf("Expected severity low, got %s", result.Severity)
	}

	// 推奨アクションが含まれているか確認
	if len(result.Recommendations) == 0 {
		t.Error("Expected recommendations, got none")
	}

	t.Logf("Recommendations: %v", result.Recommendations)
}

func TestAnalyzer_AnalyzeBatch(t *testing.T) {
	g := createTestGraph()
	analyzer := NewAnalyzer(g)

	events := []*types.DriftEvent{
		{
			ID:           "drift-001",
			ResourceID:   "aws:sg:sg-789",
			ResourceType: "security_group",
			Type:         types.DriftDeleted,
			Timestamp:    time.Now(),
			Severity:     types.SeverityCritical,
		},
		{
			ID:           "drift-002",
			ResourceID:   "aws:ec2:i-111",
			ResourceType: "ec2",
			Type:         types.DriftModified,
			Timestamp:    time.Now(),
			Severity:     types.SeverityMedium,
		},
	}

	results, err := analyzer.AnalyzeBatch(events)
	if err != nil {
		t.Fatalf("AnalyzeBatch failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// 最初のイベント (SG 削除) は critical
	if results[0].Severity != types.SeverityCritical {
		t.Errorf("Expected first result severity critical, got %s", results[0].Severity)
	}

	// 2番目のイベント (EC2 変更) は medium または high
	if results[1].Severity != types.SeverityMedium && results[1].Severity != types.SeverityHigh {
		t.Errorf("Expected second result severity medium or high, got %s", results[1].Severity)
	}

	t.Logf("Batch analysis complete: %d events processed", len(results))
}

func TestAnalyzer_SeverityEscalation(t *testing.T) {
	g := createTestGraph()

	// 多くのリソースを追加してネットワークを拡大
	for i := 3; i <= 15; i++ {
		ec2 := graph.ResourceNode{
			ID:       fmt.Sprintf("aws:ec2:i-%d", i+100),
			Type:     "ec2",
			Provider: "aws",
			Region:   "us-west-2",
		}
		g.AddNode(ec2)

		// Security Group に接続
		g.AddEdge(graph.Edge{
			From: ec2.ID,
			To:   "aws:sg:sg-789",
			Type: "network",
		})
	}

	analyzer := NewAnalyzer(g)

	event := &types.DriftEvent{
		ID:           "drift-005",
		ResourceID:   "aws:sg:sg-789",
		ResourceType: "security_group",
		Type:         types.DriftModified,
		Timestamp:    time.Now(),
		Severity:     types.SeverityMedium,
	}

	result, err := analyzer.AnalyzeImpact(event)
	if err != nil {
		t.Fatalf("AnalyzeImpact failed: %v", err)
	}

	// 10個以上のリソースが影響を受けるので severity が high に上がるはず
	if result.AffectedResourceCount <= 10 {
		t.Logf("Warning: Expected >10 affected resources for severity escalation test, got %d", result.AffectedResourceCount)
	}

	// Severity が元の medium から high または critical にエスカレートしているか確認
	if result.Severity == types.SeverityMedium || result.Severity == types.SeverityLow {
		t.Errorf("Expected severity to be escalated to high or critical due to >10 affected resources, got %s (affected: %d)",
			result.Severity, result.AffectedResourceCount)
	}

	t.Logf("Severity escalated from medium to %s due to %d affected resources",
		result.Severity, result.AffectedResourceCount)
}

func TestAnalyzer_BlastRadiusCalculation(t *testing.T) {
	g := createTestGraph()
	analyzer := NewAnalyzer(g)

	event := &types.DriftEvent{
		ID:           "drift-006",
		ResourceID:   "aws:lambda:func-456",
		ResourceType: "lambda",
		Type:         types.DriftModified,
		Timestamp:    time.Now(),
		Severity:     types.SeverityLow,
	}

	result, err := analyzer.AnalyzeImpact(event)
	if err != nil {
		t.Fatalf("AnalyzeImpact failed: %v", err)
	}

	// Lambda -> Subnet (1 hop), RDS (1 hop), VPC (2 hops), etc.
	if result.BlastRadius < 1 {
		t.Errorf("Expected blast radius >= 1, got %d", result.BlastRadius)
	}

	// Blast radius は最大 3 まで
	if result.BlastRadius > 3 {
		t.Errorf("Expected blast radius <= 3, got %d", result.BlastRadius)
	}

	t.Logf("Blast radius: %d hops", result.BlastRadius)
	t.Logf("Affected resources: %d", result.AffectedResourceCount)
}
