package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/yourusername/airdig/skygraph/pkg/graph"
)

// SecurityGroupScanner は Security Group をスキャン
type SecurityGroupScanner struct {
	client *ec2.Client
	region string
}

// NewSecurityGroupScanner は新しい Security Group スキャナーを作成
func NewSecurityGroupScanner(client *ec2.Client, region string) *SecurityGroupScanner {
	return &SecurityGroupScanner{
		client: client,
		region: region,
	}
}

// Name はスキャナー名を返す
func (s *SecurityGroupScanner) Name() string {
	return "security_group"
}

// Scan は Security Group をスキャン
func (s *SecurityGroupScanner) Scan(ctx context.Context) ([]graph.ResourceNode, error) {
	result, err := s.client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe security groups: %w", err)
	}

	nodes := make([]graph.ResourceNode, 0, len(result.SecurityGroups))

	for _, sg := range result.SecurityGroups {
		// Ingress rules を整形
		ingressRules := make([]map[string]interface{}, 0, len(sg.IpPermissions))
		for _, perm := range sg.IpPermissions {
			rule := map[string]interface{}{
				"protocol": getStringPtr(perm.IpProtocol),
			}
			if perm.FromPort != nil {
				rule["from_port"] = *perm.FromPort
			}
			if perm.ToPort != nil {
				rule["to_port"] = *perm.ToPort
			}

			// CIDR blocks
			cidrs := make([]string, 0, len(perm.IpRanges))
			for _, ipRange := range perm.IpRanges {
				if ipRange.CidrIp != nil {
					cidrs = append(cidrs, *ipRange.CidrIp)
				}
			}
			if len(cidrs) > 0 {
				rule["cidr_blocks"] = cidrs
			}

			ingressRules = append(ingressRules, rule)
		}

		node := graph.ResourceNode{
			ID:       fmt.Sprintf("aws:sg:%s", *sg.GroupId),
			Type:     "security_group",
			Provider: "aws",
			Region:   s.region,
			Name:     getStringPtr(sg.GroupName),
			Metadata: map[string]interface{}{
				"group_id":      *sg.GroupId,
				"group_name":    *sg.GroupName,
				"description":   getStringPtr(sg.Description),
				"vpc_id":        getStringPtr(sg.VpcId),
				"ingress_rules": ingressRules,
				"egress_count":  len(sg.IpPermissionsEgress),
			},
			Tags:      convertTags(sg.Tags),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}
