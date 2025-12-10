package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/yourusername/airdig/skygraph/pkg/graph"
)

func main() {
	fmt.Println("SkyGraph - Cloud Topology Scanner")
	fmt.Println("Version: 0.1.0 (MVP)")
	fmt.Println()

	// Create a demo graph
	g := createDemoGraph()

	// Print stats
	fmt.Printf("Graph Stats:\n")
	fmt.Printf("  Nodes: %d\n", g.NodeCount())
	fmt.Printf("  Edges: %d\n", g.EdgeCount())
	fmt.Println()

	// Export to JSON
	if err := exportJSON(g, "graph.json"); err != nil {
		fmt.Fprintf(os.Stderr, "Error exporting graph: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Graph exported to graph.json")
}

// createDemoGraph creates a demo AWS infrastructure graph
func createDemoGraph() *graph.Graph {
	g := graph.NewGraph()

	// VPC
	vpc := graph.ResourceNode{
		ID:       "aws:vpc:vpc-123456",
		Type:     "vpc",
		Provider: "aws",
		Region:   "us-east-1",
		Name:     "production-vpc",
		Metadata: map[string]interface{}{
			"cidr": "10.0.0.0/16",
		},
		Tags: map[string]string{
			"Environment": "production",
		},
	}
	g.AddNode(vpc)

	// Subnet
	subnet := graph.ResourceNode{
		ID:       "aws:subnet:subnet-123456",
		Type:     "subnet",
		Provider: "aws",
		Region:   "us-east-1",
		Name:     "production-subnet-1a",
		Metadata: map[string]interface{}{
			"cidr":             "10.0.1.0/24",
			"availability_zone": "us-east-1a",
		},
		Tags: map[string]string{
			"Environment": "production",
		},
	}
	g.AddNode(subnet)

	// VPC -> Subnet edge
	g.AddEdge(graph.Edge{
		From: vpc.ID,
		To:   subnet.ID,
		Type: "ownership",
	})

	// Security Group
	sg := graph.ResourceNode{
		ID:       "aws:sg:sg-123456",
		Type:     "security_group",
		Provider: "aws",
		Region:   "us-east-1",
		Name:     "web-sg",
		Metadata: map[string]interface{}{
			"ingress_rules": []map[string]interface{}{
				{"port": 80, "protocol": "tcp", "cidr": "0.0.0.0/0"},
				{"port": 443, "protocol": "tcp", "cidr": "0.0.0.0/0"},
			},
		},
		Tags: map[string]string{
			"Environment": "production",
		},
	}
	g.AddNode(sg)

	// EC2 Instance
	ec2 := graph.ResourceNode{
		ID:       "aws:ec2:i-123456",
		Type:     "ec2",
		Provider: "aws",
		Region:   "us-east-1",
		Name:     "web-server-1",
		Metadata: map[string]interface{}{
			"instance_type": "t3.large",
			"state":         "running",
			"private_ip":    "10.0.1.10",
			"public_ip":     "54.123.45.67",
		},
		Tags: map[string]string{
			"Environment": "production",
			"Role":        "web",
		},
	}
	g.AddNode(ec2)

	// Subnet -> EC2 edge
	g.AddEdge(graph.Edge{
		From: subnet.ID,
		To:   ec2.ID,
		Type: "network",
	})

	// Security Group -> EC2 edge
	g.AddEdge(graph.Edge{
		From: sg.ID,
		To:   ec2.ID,
		Type: "network",
	})

	// RDS Instance
	rds := graph.ResourceNode{
		ID:       "aws:rds:database-1",
		Type:     "rds",
		Provider: "aws",
		Region:   "us-east-1",
		Name:     "production-db",
		Metadata: map[string]interface{}{
			"engine":         "postgres",
			"engine_version": "14.7",
			"instance_class": "db.t3.medium",
			"storage":        100,
		},
		Tags: map[string]string{
			"Environment": "production",
		},
	}
	g.AddNode(rds)

	// Subnet -> RDS edge
	g.AddEdge(graph.Edge{
		From: subnet.ID,
		To:   rds.ID,
		Type: "network",
	})

	// EC2 -> RDS dependency edge
	g.AddEdge(graph.Edge{
		From: ec2.ID,
		To:   rds.ID,
		Type: "dependency",
		Metadata: map[string]interface{}{
			"connection_type": "database",
		},
	})

	return g
}

// exportJSON exports the graph to a JSON file
func exportJSON(g *graph.Graph, filename string) error {
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
