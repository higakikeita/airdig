package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/higakikeita/airdig/deepdrift/pkg/api"
	"github.com/higakikeita/airdig/deepdrift/pkg/impact"
	"github.com/higakikeita/airdig/deepdrift/pkg/storage/clickhouse"
	"github.com/higakikeita/airdig/deepdrift/pkg/tfdrift"
	"github.com/higakikeita/airdig/deepdrift/pkg/types"
	"github.com/higakikeita/airdig/skygraph/pkg/graph"
)

var (
	command       = flag.String("command", "detect", "Command to run: detect, impact, watch, server")
	stateFile     = flag.String("state", "terraform.tfstate", "Terraform state file path")
	graphFile     = flag.String("graph", "", "SkyGraph JSON file path (required for impact analysis)")
	tfdriftPath   = flag.String("tfdrift", "", "TFDrift binary path (default: ~/tfdrift-falco/bin/tfdrift)")
	configPath    = flag.String("config", "", "TFDrift config file path")
	output        = flag.String("output", "", "Output file path (default: stdout)")
	watchInterval = flag.Duration("interval", 5*time.Minute, "Watch interval for continuous monitoring")

	// Server flags
	serverPort      = flag.Int("port", 8080, "API server port")
	serverHost      = flag.String("host", "0.0.0.0", "API server host")
	clickhouseHost  = flag.String("clickhouse-host", "localhost", "ClickHouse host")
	clickhousePort  = flag.Int("clickhouse-port", 9000, "ClickHouse port")
	clickhouseDB    = flag.String("clickhouse-db", "deepdrift", "ClickHouse database name")
)

func main() {
	flag.Parse()

	fmt.Println("==============================================")
	fmt.Println("  DeepDrift - Drift Detection & Impact Analysis")
	fmt.Println("  Version: 0.1.0 (Alpha)")
	fmt.Println("==============================================")
	fmt.Println()

	ctx := context.Background()

	switch *command {
	case "detect":
		if err := runDetect(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "impact":
		if err := runImpact(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "watch":
		if err := runWatch(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "server":
		if err := runServer(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", *command)
		fmt.Fprintf(os.Stderr, "Available commands: detect, impact, watch, server\n")
		os.Exit(1)
	}
}

func runDetect(ctx context.Context) error {
	fmt.Println("Running drift detection...")
	fmt.Printf("State file: %s\n", *stateFile)
	fmt.Println()

	// TFDrift adapter を作成
	adapter := tfdrift.NewTFDriftAdapter(*tfdriftPath, *configPath)

	// Drift detection を実行
	events, err := adapter.DetectDrift(ctx, *stateFile)
	if err != nil {
		return fmt.Errorf("drift detection failed: %w", err)
	}

	// 結果を表示
	fmt.Printf("Found %d drift events\n\n", len(events))

	if len(events) == 0 {
		fmt.Println("✅ No drift detected")
		return nil
	}

	for i, event := range events {
		fmt.Printf("%d. [%s] %s (%s)\n", i+1, event.Type, event.ResourceID, event.ResourceType)
		fmt.Printf("   Severity: %s\n", event.Severity)
		if event.RootCause != nil {
			fmt.Printf("   User: %s\n", event.RootCause.UserIdentity)
			fmt.Printf("   Event: %s\n", event.RootCause.EventName)
		}
		fmt.Println()
	}

	// 出力ファイルに保存
	if *output != "" {
		return saveJSON(events, *output)
	}

	return nil
}

func runImpact(ctx context.Context) error {
	if *graphFile == "" {
		return fmt.Errorf("--graph is required for impact analysis")
	}

	fmt.Println("Running impact analysis...")
	fmt.Printf("Graph file: %s\n", *graphFile)
	fmt.Println()

	// SkyGraph を読み込み
	g, err := loadGraph(*graphFile)
	if err != nil {
		return fmt.Errorf("failed to load graph: %w", err)
	}

	fmt.Printf("Loaded graph: %d nodes, %d edges\n\n", g.NodeCount(), g.EdgeCount())

	// Drift detection を実行
	adapter := tfdrift.NewTFDriftAdapter(*tfdriftPath, *configPath)
	events, err := adapter.DetectDrift(ctx, *stateFile)
	if err != nil {
		return fmt.Errorf("drift detection failed: %w", err)
	}

	if len(events) == 0 {
		fmt.Println("✅ No drift detected")
		return nil
	}

	// Impact analysis を実行
	analyzer := impact.NewAnalyzer(g)
	results, err := analyzer.AnalyzeBatch(events)
	if err != nil {
		return fmt.Errorf("impact analysis failed: %w", err)
	}

	// 結果を表示
	fmt.Printf("Analyzed %d drift events\n\n", len(results))

	for i, result := range results {
		event := events[i]
		fmt.Printf("%d. [%s] %s (%s)\n", i+1, event.Type, event.ResourceID, event.ResourceType)
		fmt.Printf("   Severity: %s\n", result.Severity)
		fmt.Printf("   Affected resources: %d\n", result.AffectedResourceCount)
		fmt.Printf("   Blast radius: %d hops\n", result.BlastRadius)

		if len(result.Recommendations) > 0 {
			fmt.Println("   Recommendations:")
			for _, rec := range result.Recommendations {
				fmt.Printf("     • %s\n", rec)
			}
		}
		fmt.Println()
	}

	// 出力ファイルに保存
	if *output != "" {
		return saveJSON(results, *output)
	}

	return nil
}

func runWatch(ctx context.Context) error {
	fmt.Println("Starting continuous drift monitoring...")
	fmt.Printf("Interval: %s\n", *watchInterval)
	fmt.Println()

	adapter := tfdrift.NewTFDriftAdapter(*tfdriftPath, *configPath)

	// コールバック関数
	callback := func(events []*types.DriftEvent) error {
		fmt.Printf("[%s] Detected %d drift events\n", time.Now().Format(time.RFC3339), len(events))

		for _, event := range events {
			fmt.Printf("  - [%s] %s (%s)\n", event.Type, event.ResourceID, event.ResourceType)
		}
		fmt.Println()

		return nil
	}

	// 継続的監視を開始
	return adapter.WatchDrift(ctx, *stateFile, *watchInterval, callback)
}

func loadGraph(filename string) (*graph.Graph, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var g graph.Graph
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, err
	}

	return &g, nil
}

func saveJSON(data interface{}, filename string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return err
	}

	fmt.Printf("✅ Saved to: %s\n", filename)
	return nil
}

func runServer(ctx context.Context) error {
	fmt.Println("Starting API server...")
	fmt.Printf("Server: %s:%d\n", *serverHost, *serverPort)
	fmt.Printf("ClickHouse: %s:%d/%s\n", *clickhouseHost, *clickhousePort, *clickhouseDB)
	fmt.Println()

	// Connect to ClickHouse
	chConfig := &clickhouse.Config{
		Host:     *clickhouseHost,
		Port:     *clickhousePort,
		Database: *clickhouseDB,
		Username: "default",
		Password: "",
		Debug:    false,
	}

	chClient, err := clickhouse.NewClient(chConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}
	defer chClient.Close()

	fmt.Println("✅ Connected to ClickHouse")

	// Create API server
	apiConfig := &api.Config{
		Host:           *serverHost,
		Port:           *serverPort,
		ClickHouseAddr: fmt.Sprintf("%s:%d", *clickhouseHost, *clickhousePort),
		ClickHouseDB:   *clickhouseDB,
		EnableCORS:     true,
		AllowedOrigins: []string{"*"},
	}

	server := api.NewServer(apiConfig, chClient)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 1)

	// Start server in goroutine
	go func() {
		if err := server.Start(); err != nil {
			errChan <- err
		}
	}()

	fmt.Printf("✅ Server started on http://%s:%d\n", *serverHost, *serverPort)
	fmt.Println()
	fmt.Println("API Endpoints:")
	fmt.Println("  GET  /health                    - Health check")
	fmt.Println("  GET  /api/v1/drifts             - List drift events")
	fmt.Println("  GET  /api/v1/drifts/{id}        - Get drift event by ID")
	fmt.Println("  GET  /api/v1/drifts/stats       - Get drift statistics")
	fmt.Println("  GET  /api/v1/impact             - List impact analysis")
	fmt.Println("  GET  /api/v1/impact/{id}        - Get impact by drift ID")
	fmt.Println("  GET  /api/v1/impact/stats       - Get impact statistics")
	fmt.Println("  GET  /api/v1/impact/high        - Get high impact drifts")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop the server...")

	// Wait for shutdown signal or error
	select {
	case <-sigChan:
		fmt.Println("\nReceived shutdown signal, gracefully shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	}
}
