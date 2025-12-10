package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/yourusername/airdig/skygraph/pkg/graph"
)

// RDSScanner は RDS インスタンスをスキャン
type RDSScanner struct {
	client *rds.Client
	region string
}

// NewRDSScanner は新しい RDS スキャナーを作成
func NewRDSScanner(client *rds.Client, region string) *RDSScanner {
	return &RDSScanner{
		client: client,
		region: region,
	}
}

// Name はスキャナー名を返す
func (s *RDSScanner) Name() string {
	return "rds"
}

// Scan は RDS インスタンスをスキャン
func (s *RDSScanner) Scan(ctx context.Context) ([]graph.ResourceNode, error) {
	result, err := s.client.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe RDS instances: %w", err)
	}

	nodes := make([]graph.ResourceNode, 0, len(result.DBInstances))

	for _, db := range result.DBInstances {
		// Subnet Group から VPC ID と Subnet IDs を取得
		var vpcID string
		subnetIDs := make([]string, 0)

		if db.DBSubnetGroup != nil {
			vpcID = getStringPtr(db.DBSubnetGroup.VpcId)
			for _, subnet := range db.DBSubnetGroup.Subnets {
				if subnet.SubnetIdentifier != nil {
					subnetIDs = append(subnetIDs, *subnet.SubnetIdentifier)
				}
			}
		}

		// Security Groups
		sgIDs := make([]string, 0, len(db.VpcSecurityGroups))
		for _, sg := range db.VpcSecurityGroups {
			if sg.VpcSecurityGroupId != nil {
				sgIDs = append(sgIDs, *sg.VpcSecurityGroupId)
			}
		}

		node := graph.ResourceNode{
			ID:       fmt.Sprintf("aws:rds:%s", *db.DBInstanceIdentifier),
			Type:     "rds",
			Provider: "aws",
			Region:   s.region,
			Name:     *db.DBInstanceIdentifier,
			Metadata: map[string]interface{}{
				"db_instance_id":   *db.DBInstanceIdentifier,
				"engine":           *db.Engine,
				"engine_version":   *db.EngineVersion,
				"instance_class":   *db.DBInstanceClass,
				"storage":          *db.AllocatedStorage,
				"storage_type":     getStringPtr(db.StorageType),
				"status":           *db.DBInstanceStatus,
				"endpoint":         getStringPtr(db.Endpoint.Address),
				"port":             getInt32Ptr(db.Endpoint.Port),
				"vpc_id":           vpcID,
				"subnet_ids":       subnetIDs,
				"security_groups":  sgIDs,
				"multi_az":         *db.MultiAZ,
				"publicly_accessible": *db.PubliclyAccessible,
			},
			Tags:      convertRDSTags(db.TagList),
			CreatedAt: getTimePtr(db.InstanceCreateTime),
			UpdatedAt: time.Now(),
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}
