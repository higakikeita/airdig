package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/yourusername/airdig/skygraph/pkg/graph"
)

// SubnetScanner は Subnet をスキャン
type SubnetScanner struct {
	client *ec2.Client
	region string
}

// NewSubnetScanner は新しい Subnet スキャナーを作成
func NewSubnetScanner(client *ec2.Client, region string) *SubnetScanner {
	return &SubnetScanner{
		client: client,
		region: region,
	}
}

// Name はスキャナー名を返す
func (s *SubnetScanner) Name() string {
	return "subnet"
}

// Scan は Subnet をスキャン
func (s *SubnetScanner) Scan(ctx context.Context) ([]graph.ResourceNode, error) {
	result, err := s.client.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe subnets: %w", err)
	}

	nodes := make([]graph.ResourceNode, 0, len(result.Subnets))

	for _, subnet := range result.Subnets {
		node := graph.ResourceNode{
			ID:       fmt.Sprintf("aws:subnet:%s", *subnet.SubnetId),
			Type:     "subnet",
			Provider: "aws",
			Region:   s.region,
			Name:     getNameTag(subnet.Tags),
			Metadata: map[string]interface{}{
				"subnet_id":         *subnet.SubnetId,
				"vpc_id":            *subnet.VpcId,
				"cidr_block":        *subnet.CidrBlock,
				"availability_zone": *subnet.AvailabilityZone,
				"state":             string(subnet.State),
				"available_ips":     *subnet.AvailableIpAddressCount,
			},
			Tags:      convertTags(subnet.Tags),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}
