import type { ElementDefinition } from 'cytoscape';
import type { GraphNode, GraphEdge } from '../graph/graphUtils';
import { JAPANESE_LABELS } from '../graph/graphUtils';

export interface CytoscapeData {
  elements: ElementDefinition[];
}

// Determine health status based on node metadata
const determineHealth = (node: GraphNode): 'healthy' | 'warning' | 'critical' | undefined => {
  // Public access without proper security = warning
  const hasPublicAccess = node.metadata?.ingress_rules?.some((rule: any) =>
    rule.cidr_blocks?.includes('0.0.0.0/0')
  );

  if (hasPublicAccess) {
    // Critical if sensitive service (database, etc.)
    if (node.type === 'rds' || node.type === 'dynamodb') {
      return 'critical';
    }
    return 'warning';
  }

  // Default to healthy
  return 'healthy';
};

// Extract details from node metadata
const extractDetails = (node: GraphNode): Record<string, any> => {
  const details: Record<string, any> = {};

  if (node.metadata) {
    // EC2
    if (node.type === 'ec2') {
      details.instanceType = node.metadata.instance_type;
      details.instanceState = node.metadata.state;
      details.privateIp = node.metadata.private_ip;
      details.publicIp = node.metadata.public_ip;
    }

    // RDS
    if (node.type === 'rds') {
      details.engine = node.metadata.engine;
      details.engineVersion = node.metadata.engine_version;
      details.instanceClass = node.metadata.instance_class;
      details.allocatedStorage = node.metadata.allocated_storage;
      details.multiAZ = node.metadata.multi_az;
    }

    // EKS
    if (node.type === 'eks_cluster') {
      details.version = node.metadata.version;
      details.status = node.metadata.status;
      details.endpoint = node.metadata.endpoint;
    }

    // Security Group
    if (node.type === 'security_group') {
      details.ingressRules = node.metadata.ingress_rules?.length || 0;
      details.egressRules = node.metadata.egress_rules?.length || 0;
    }

    // S3
    if (node.type === 's3') {
      details.versioning = node.metadata.versioning;
      details.encryption = node.metadata.encryption;
    }
  }

  return details;
};

// Check if node has public access
const isPublicNode = (node: GraphNode): boolean => {
  return !!(
    node.metadata?.public_ip ||
    node.metadata?.is_public ||
    node.metadata?.ingress_rules?.some((rule: any) =>
      rule.cidr_blocks?.includes('0.0.0.0/0')
    )
  );
};

// Generate a readable label for a node
const generateNodeLabel = (node: GraphNode): string => {
  // Try to get a human-readable name from tags
  const tagName = node.tags?.Name || node.tags?.name;

  // Get Japanese type label
  const typeLabel = JAPANESE_LABELS[node.type];

  // Get short ID (last part after the last colon or slash)
  const shortId = node.id.split(':').pop()?.split('/').pop() || node.id;

  // Priority order:
  // 1. If there's a tag name, use "TypeLabel (TagName)"
  // 2. If there's a type label, use "TypeLabel"
  // 3. Fall back to short ID
  if (tagName && typeLabel) {
    return `${typeLabel} (${tagName})`;
  } else if (typeLabel) {
    return typeLabel;
  } else if (tagName) {
    return tagName;
  } else {
    return shortId;
  }
};

// Convert GraphNode to Cytoscape element
export const graphNodeToCytoscapeNode = (node: GraphNode): ElementDefinition => {
  const label = generateNodeLabel(node);

  return {
    data: {
      id: node.id,
      label,
      type: node.type,
      health: determineHealth(node),
      public: isPublicNode(node),
      details: extractDetails(node),
      provider: node.provider,
      region: node.region,
      // Store original node for reference
      _original: node,
    },
  };
};

// Convert GraphEdge to Cytoscape edge
export const graphEdgeToCytoscapeEdge = (edge: GraphEdge): ElementDefinition => {
  return {
    data: {
      id: `${edge.from}-${edge.to}`,
      source: edge.from,
      target: edge.to,
      type: edge.type || 'default',
    },
  };
};

// Convert full graph data to Cytoscape format
export const convertGraphToCytoscape = (
  nodes: GraphNode[],
  edges: GraphEdge[]
): CytoscapeData => {
  const elements: ElementDefinition[] = [
    ...nodes.map(graphNodeToCytoscapeNode),
    ...edges.map(graphEdgeToCytoscapeEdge),
  ];

  return { elements };
};

// Add synthetic edges based on relationships
export const addSyntheticEdges = (
  nodes: GraphNode[],
  elements: ElementDefinition[]
): void => {
  const nodeMap = new Map(nodes.map(n => [n.id, n]));

  nodes.forEach(node => {
    // VPC → Subnet
    if (node.type === 'subnet' && node.metadata?.vpc_id) {
      const vpcId = `aws:vpc:${node.metadata.vpc_id}`;
      if (nodeMap.has(vpcId)) {
        elements.push({
          data: {
            id: `${vpcId}-${node.id}`,
            source: vpcId,
            target: node.id,
            type: 'ownership',
          },
        });
      }
    }

    // Subnet → EC2
    if (node.type === 'ec2' && node.metadata?.subnet_id) {
      const subnetId = `aws:subnet:${node.metadata.subnet_id}`;
      if (nodeMap.has(subnetId)) {
        elements.push({
          data: {
            id: `${subnetId}-${node.id}`,
            source: subnetId,
            target: node.id,
            type: 'ownership',
          },
        });
      }
    }

    // ALB → EC2 (network connection)
    if (node.type === 'alb' || node.type === 'elb') {
      const vpcId = node.metadata?.vpc_id;
      if (vpcId) {
        nodes.forEach(targetNode => {
          if (targetNode.type === 'ec2' && targetNode.metadata?.vpc_id === vpcId) {
            elements.push({
              data: {
                id: `${node.id}-${targetNode.id}`,
                source: node.id,
                target: targetNode.id,
                type: 'network',
              },
            });
          }
        });
      }
    }

    // EC2 → RDS (dependency)
    if (node.type === 'ec2') {
      const vpcId = node.metadata?.vpc_id;
      if (vpcId) {
        nodes.forEach(targetNode => {
          if (targetNode.type === 'rds' && targetNode.metadata?.vpc_id === vpcId) {
            elements.push({
              data: {
                id: `${node.id}-${targetNode.id}`,
                source: node.id,
                target: targetNode.id,
                type: 'dependency',
              },
            });
          }
        });
      }
    }

    // NAT Gateway → Internet
    if (node.type === 'nat_gateway') {
      // Check if internet node exists, if not we'll add it in the component
      const internetNodeId = 'internet-global';
      elements.push({
        data: {
          id: `${node.id}-${internetNodeId}`,
          source: node.id,
          target: internetNodeId,
          type: 'network',
        },
      });
    }
  });
};
