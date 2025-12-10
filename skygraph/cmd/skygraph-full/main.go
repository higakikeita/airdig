package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/yourusername/airdig/skygraph/pkg/aws"
	"github.com/yourusername/airdig/skygraph/pkg/builder"
)

var (
	provider = flag.String("provider", "aws", "Cloud provider (aws, gcp, azure, kubernetes)")
	region   = flag.String("region", "us-east-1", "AWS region")
	profile  = flag.String("profile", "default", "AWS profile")
	output   = flag.String("output", "graph.json", "Output file path")
	verbose  = flag.Bool("verbose", false, "Verbose output")
)

func main() {
	flag.Parse()

	fmt.Println("==============================================")
	fmt.Println("  SkyGraph - Cloud Topology Scanner")
	fmt.Println("  Version: 0.1.0 MVP")
	fmt.Println("==============================================")
	fmt.Println()

	if *provider != "aws" {
		fmt.Fprintf(os.Stderr, "Error: Only 'aws' provider is supported in v0.1.0\n")
		os.Exit(1)
	}

	// コンテキスト作成（タイムアウト 5分）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Printf("Provider: %s\n", *provider)
	fmt.Printf("Region: %s\n", *region)
	fmt.Printf("Profile: %s\n", *profile)
	fmt.Println()

	// AWS スキャナーを作成
	fmt.Println("Initializing AWS scanner...")
	scanner, err := aws.NewAWSScanner(ctx, *region, *profile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create AWS scanner: %v\n", err)
		os.Exit(1)
	}

	// スキャン実行
	fmt.Println("Scanning AWS resources...")
	fmt.Println("  - VPC")
	fmt.Println("  - Subnet")
	fmt.Println("  - Security Group")
	fmt.Println("  - EC2 Instances")
	fmt.Println("  - RDS Instances")
	fmt.Println()

	startTime := time.Now()
	result, err := scanner.ScanAll(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Scan failed: %v\n", err)
		os.Exit(1)
	}
	scanDuration := time.Since(startTime)

	// エラーレポート
	if len(result.Errors) > 0 {
		fmt.Println("⚠ Some scanners failed:")
		for name, err := range result.Errors {
			fmt.Printf("  - %s: %v\n", name, err)
		}
		fmt.Println()
	}

	// グラフ構築
	fmt.Println("Building graph...")
	graphBuilder := builder.NewGraphBuilder()
	graphBuilder.AddNodes(result.Nodes)

	if err := graphBuilder.InferEdges(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to infer edges: %v\n", err)
		os.Exit(1)
	}

	graph := graphBuilder.Build()

	// 統計情報
	fmt.Println()
	fmt.Println("==============================================")
	fmt.Println("  Scan Results")
	fmt.Println("==============================================")
	fmt.Printf("Duration: %.2f seconds\n", scanDuration.Seconds())
	fmt.Printf("Nodes: %d\n", graph.NodeCount())
	fmt.Printf("Edges: %d\n", graph.EdgeCount())
	fmt.Println()

	// ノードタイプ別の統計
	nodeCounts := make(map[string]int)
	for _, node := range graph.Nodes {
		nodeCounts[node.Type]++
	}

	fmt.Println("Resources by type:")
	for nodeType, count := range nodeCounts {
		fmt.Printf("  - %s: %d\n", nodeType, count)
	}
	fmt.Println()

	// エッジタイプ別の統計
	if *verbose {
		edgeCounts := make(map[string]int)
		for _, edge := range graph.Edges {
			edgeCounts[edge.Type]++
		}

		fmt.Println("Edges by type:")
		for edgeType, count := range edgeCounts {
			fmt.Printf("  - %s: %d\n", edgeType, count)
		}
		fmt.Println()
	}

	// JSON エクスポート
	fmt.Printf("Exporting to %s...\n", *output)
	if err := exportJSON(graph, *output); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to export graph: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✅ Done!")
	fmt.Printf("Graph saved to: %s\n", *output)
}

// exportJSON はグラフを JSON ファイルにエクスポート
func exportJSON(g interface{}, filename string) error {
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
