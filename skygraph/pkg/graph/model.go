package graph

import "time"

// ResourceNode represents a cloud resource in the graph
type ResourceNode struct {
	// Unique identifier (e.g., "aws:ec2:i-123456", "k8s:pod:frontend-7d8f9")
	ID string `json:"id"`

	// Resource type (e.g., "ec2", "vpc", "rds", "k8s_pod")
	Type string `json:"type"`

	// Cloud provider (e.g., "aws", "gcp", "azure", "kubernetes")
	Provider string `json:"provider"`

	// Region or zone
	Region string `json:"region,omitempty"`

	// Human-readable name
	Name string `json:"name,omitempty"`

	// Provider-specific attributes
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Resource tags/labels
	Tags map[string]string `json:"tags,omitempty"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Edge represents a relationship between two resources
type Edge struct {
	// Source node ID
	From string `json:"from"`

	// Target node ID
	To string `json:"to"`

	// Relationship type:
	// - "network": Network connection (e.g., EC2 in VPC)
	// - "dependency": Logical dependency (e.g., EC2 depends on RDS)
	// - "ownership": Parent-child relationship (e.g., VPC owns Subnet)
	// - "call": Service call (from TraceCore)
	// - "drift": Configuration drift (from DeepDrift)
	Type string `json:"type"`

	// Optional weight (e.g., bandwidth, cost, priority)
	Weight float64 `json:"weight,omitempty"`

	// Edge metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Graph represents the complete resource graph
type Graph struct {
	Nodes []ResourceNode `json:"nodes"`
	Edges []Edge         `json:"edges"`
}

// NewGraph creates a new empty graph
func NewGraph() *Graph {
	return &Graph{
		Nodes: make([]ResourceNode, 0),
		Edges: make([]Edge, 0),
	}
}

// AddNode adds a node to the graph
func (g *Graph) AddNode(node ResourceNode) {
	g.Nodes = append(g.Nodes, node)
}

// AddEdge adds an edge to the graph
func (g *Graph) AddEdge(edge Edge) {
	g.Edges = append(g.Edges, edge)
}

// FindNode finds a node by ID
func (g *Graph) FindNode(id string) *ResourceNode {
	for i := range g.Nodes {
		if g.Nodes[i].ID == id {
			return &g.Nodes[i]
		}
	}
	return nil
}

// FindEdges finds all edges for a given node
func (g *Graph) FindEdges(nodeID string) []Edge {
	edges := make([]Edge, 0)
	for _, edge := range g.Edges {
		if edge.From == nodeID || edge.To == nodeID {
			edges = append(edges, edge)
		}
	}
	return edges
}

// NodeCount returns the number of nodes in the graph
func (g *Graph) NodeCount() int {
	return len(g.Nodes)
}

// EdgeCount returns the number of edges in the graph
func (g *Graph) EdgeCount() int {
	return len(g.Edges)
}
