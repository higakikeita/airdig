package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/yourusername/airdig/skygraph/pkg/graph"
)

// EC2Scanner は EC2 インスタンスをスキャン
type EC2Scanner struct {
	client *ec2.Client
	region string
}

// NewEC2Scanner は新しい EC2 スキャナーを作成
func NewEC2Scanner(client *ec2.Client, region string) *EC2Scanner {
	return &EC2Scanner{
		client: client,
		region: region,
	}
}

// Name はスキャナー名を返す
func (s *EC2Scanner) Name() string {
	return "ec2"
}

// Scan は EC2 インスタンスをスキャン
func (s *EC2Scanner) Scan(ctx context.Context) ([]graph.ResourceNode, error) {
	result, err := s.client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe instances: %w", err)
	}

	nodes := make([]graph.ResourceNode, 0)

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			// Security Group IDs を抽出
			sgIDs := make([]string, 0, len(instance.SecurityGroups))
			for _, sg := range instance.SecurityGroups {
				if sg.GroupId != nil {
					sgIDs = append(sgIDs, *sg.GroupId)
				}
			}

			node := graph.ResourceNode{
				ID:       fmt.Sprintf("aws:ec2:%s", *instance.InstanceId),
				Type:     "ec2",
				Provider: "aws",
				Region:   s.region,
				Name:     getNameTag(instance.Tags),
				Metadata: map[string]interface{}{
					"instance_id":      *instance.InstanceId,
					"instance_type":    string(instance.InstanceType),
					"state":            string(instance.State.Name),
					"vpc_id":           getStringPtr(instance.VpcId),
					"subnet_id":        getStringPtr(instance.SubnetId),
					"private_ip":       getStringPtr(instance.PrivateIpAddress),
					"public_ip":        getStringPtr(instance.PublicIpAddress),
					"availability_zone": getStringPtr(instance.Placement.AvailabilityZone),
					"security_groups":  sgIDs,
					"ami_id":           getStringPtr(instance.ImageId),
				},
				Tags:      convertTags(instance.Tags),
				CreatedAt: getTimePtr(instance.LaunchTime),
				UpdatedAt: time.Now(),
			}

			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}
