package aws

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/yourusername/airdig/skygraph/pkg/graph"
	"github.com/yourusername/airdig/skygraph/pkg/scanner"
)

// AWSScanner は AWS リソースをスキャンする
type AWSScanner struct {
	region  string
	profile string

	ec2Client *ec2.Client
	rdsClient *rds.Client
}

// NewAWSScanner は新しい AWS スキャナーを作成
func NewAWSScanner(ctx context.Context, region, profile string) (*AWSScanner, error) {
	// AWS 設定をロード
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &AWSScanner{
		region:    region,
		profile:   profile,
		ec2Client: ec2.NewFromConfig(cfg),
		rdsClient: rds.NewFromConfig(cfg),
	}, nil
}

// ScanAll は全ての AWS リソースをスキャン
func (s *AWSScanner) ScanAll(ctx context.Context) (*scanner.Result, error) {
	result := &scanner.Result{
		Nodes:  make([]graph.ResourceNode, 0),
		Errors: make(map[string]error),
	}

	// 各リソーススキャナーを並列実行
	scanners := []scanner.Scanner{
		NewVPCScanner(s.ec2Client, s.region),
		NewSubnetScanner(s.ec2Client, s.region),
		NewSecurityGroupScanner(s.ec2Client, s.region),
		NewEC2Scanner(s.ec2Client, s.region),
		NewRDSScanner(s.rdsClient, s.region),
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, sc := range scanners {
		wg.Add(1)
		go func(scanner scanner.Scanner) {
			defer wg.Done()

			nodes, err := scanner.Scan(ctx)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				result.Errors[scanner.Name()] = err
			} else {
				result.Nodes = append(result.Nodes, nodes...)
			}
		}(sc)
	}

	wg.Wait()

	return result, nil
}
