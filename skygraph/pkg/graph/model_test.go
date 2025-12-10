package graph

import (
	"testing"
	"time"
)

func TestNewGraph(t *testing.T) {
	g := NewGraph()

	if g == nil {
		t.Fatal("NewGraph() returned nil")
	}

	if g.NodeCount() != 0 {
		t.Errorf("Expected 0 nodes, got %d", g.NodeCount())
	}

	if g.EdgeCount() != 0 {
		t.Errorf("Expected 0 edges, got %d", g.EdgeCount())
	}
}

func TestGraph_AddNode(t *testing.T) {
	g := NewGraph()

	node := ResourceNode{
		ID:       "test-node-1",
		Type:     "ec2",
		Provider: "aws",
		Region:   "us-east-1",
		Name:     "test-instance",
		Metadata: map[string]interface{}{
			"instance_type": "t3.micro",
		},
		Tags: map[string]string{
			"Environment": "test",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	g.AddNode(node)

	if g.NodeCount() != 1 {
		t.Errorf("Expected 1 node, got %d", g.NodeCount())
	}

	found := g.FindNode("test-node-1")
	if found == nil {
		t.Error("Node not found after adding")
	}

	if found.Type != "ec2" {
		t.Errorf("Expected type 'ec2', got '%s'", found.Type)
	}
}

func TestGraph_AddEdge(t *testing.T) {
	g := NewGraph()

	edge := Edge{
		From: "node-1",
		To:   "node-2",
		Type: "network",
	}

	g.AddEdge(edge)

	if g.EdgeCount() != 1 {
		t.Errorf("Expected 1 edge, got %d", g.EdgeCount())
	}
}

func TestGraph_FindNode(t *testing.T) {
	g := NewGraph()

	node1 := ResourceNode{ID: "node-1", Type: "vpc"}
	node2 := ResourceNode{ID: "node-2", Type: "ec2"}

	g.AddNode(node1)
	g.AddNode(node2)

	// 存在するノード
	found := g.FindNode("node-1")
	if found == nil {
		t.Error("Expected to find node-1")
	}
	if found.Type != "vpc" {
		t.Errorf("Expected type 'vpc', got '%s'", found.Type)
	}

	// 存在しないノード
	notFound := g.FindNode("node-999")
	if notFound != nil {
		t.Error("Expected nil for non-existent node")
	}
}

func TestGraph_FindEdges(t *testing.T) {
	g := NewGraph()

	g.AddEdge(Edge{From: "node-1", To: "node-2", Type: "network"})
	g.AddEdge(Edge{From: "node-2", To: "node-3", Type: "dependency"})
	g.AddEdge(Edge{From: "node-3", To: "node-1", Type: "ownership"})

	// node-1 に関連するエッジ（From または To）
	edges := g.FindEdges("node-1")
	if len(edges) != 2 {
		t.Errorf("Expected 2 edges for node-1, got %d", len(edges))
	}

	// node-2 に関連するエッジ
	edges = g.FindEdges("node-2")
	if len(edges) != 2 {
		t.Errorf("Expected 2 edges for node-2, got %d", len(edges))
	}

	// 存在しないノード
	edges = g.FindEdges("node-999")
	if len(edges) != 0 {
		t.Errorf("Expected 0 edges for non-existent node, got %d", len(edges))
	}
}
