package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/yourusername/airdig/skygraph/pkg/graph"
)

// VPCScanner は VPC をスキャン
type VPCScanner struct {
	client *ec2.Client
	region string
}

// NewVPCScanner は新しい VPC スキャナーを作成
func NewVPCScanner(client *ec2.Client, region string) *VPCScanner {
	return &VPCScanner{
		client: client,
		region: region,
	}
}

// Name はスキャナー名を返す
func (s *VPCScanner) Name() string {
	return "vpc"
}

// Scan は VPC をスキャン
func (s *VPCScanner) Scan(ctx context.Context) ([]graph.ResourceNode, error) {
	result, err := s.client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe VPCs: %w", err)
	}

	nodes := make([]graph.ResourceNode, 0, len(result.Vpcs))

	for _, vpc := range result.Vpcs {
		node := graph.ResourceNode{
			ID:       fmt.Sprintf("aws:vpc:%s", *vpc.VpcId),
			Type:     "vpc",
			Provider: "aws",
			Region:   s.region,
			Name:     getNameTag(vpc.Tags),
			Metadata: map[string]interface{}{
				"vpc_id":       *vpc.VpcId,
				"cidr_block":   *vpc.CidrBlock,
				"state":        string(vpc.State),
				"is_default":   *vpc.IsDefault,
				"dhcp_options": *vpc.DhcpOptionsId,
			},
			Tags:      convertTags(vpc.Tags),
			CreatedAt: time.Now(), // AWS doesn't provide creation time for VPC
			UpdatedAt: time.Now(),
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}
