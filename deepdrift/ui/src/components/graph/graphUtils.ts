import { MarkerType } from 'reactflow';
import type { Node, Edge } from 'reactflow';

// Types
export interface GraphNode {
  id: string;
  type: string;
  provider: string;
  region: string;
  name: string;
  metadata?: Record<string, any>;
  tags?: Record<string, string>;
}

export interface GraphEdge {
  from: string;
  to: string;
  type: string;
  metadata?: Record<string, any>;
}

export interface ResourceGraphData {
  nodes: GraphNode[];
  edges: GraphEdge[];
}

// Layer configuration
export interface LayerConfig {
  id: string;
  label: string;
  labelEn: string;
  y: number;
  height: number;
  color: string;
  services: string[];
}

export const SERVICE_LAYERS: Record<string, LayerConfig> = {
  INTERNET: {
    id: 'internet',
    label: 'インターネット',
    labelEn: 'Internet',
    y: 50,
    height: 100,
    color: '#10b981',
    services: ['internet'],
  },
  EDGE: {
    id: 'edge',
    label: 'エッジ層',
    labelEn: 'Edge Layer',
    y: 200,
    height: 150,
    color: '#8b5cf6',
    services: ['cloudfront', 'waf_web_acl', 'waf_ip_set', 'route53'],
  },
  NETWORK: {
    id: 'network',
    label: 'ネットワーク層',
    labelEn: 'Network Layer',
    y: 400,
    height: 180,
    color: '#06b6d4',
    services: ['alb', 'nlb', 'elb', 'nat_gateway', 'internet_gateway', 'igw', 'vpc', 'subnet', 'security_group'],
  },
  COMPUTE: {
    id: 'compute',
    label: 'コンピュート層',
    labelEn: 'Compute Layer',
    y: 630,
    height: 200,
    color: '#f59e0b',
    services: ['ec2', 'ecs', 'eks_cluster', 'eks_node_group', 'lambda'],
  },
  DATA: {
    id: 'data',
    label: 'データ層',
    labelEn: 'Data Layer',
    y: 880,
    height: 180,
    color: '#6366f1',
    services: ['rds', 's3', 'dynamodb', 'elasticache', 'cloudwatch_logs'],
  },
};

// Get layer for a given resource type
export const getResourceLayer = (type: string): LayerConfig | null => {
  for (const layer of Object.values(SERVICE_LAYERS)) {
    if (layer.services.includes(type)) {
      return layer;
    }
  }
  return null;
};

// Japanese label mapping
export const JAPANESE_LABELS: Record<string, string> = {
  vpc: 'VPC',
  subnet: 'サブネット',
  security_group: 'セキュリティグループ',
  ec2: 'EC2インスタンス',
  rds: 'RDSデータベース',
  elb: 'ロードバランサー',
  alb: 'ALB',
  lambda: 'Lambda関数',
  s3: 'S3バケット',
  eks_cluster: 'EKSクラスター',
  eks_node_group: 'EKSノードグループ',
  nat_gateway: 'NATゲートウェイ',
  waf_web_acl: 'WAF Web ACL',
  waf_ip_set: 'WAF IPセット',
  internet_gateway: 'インターネットゲートウェイ',
  dynamodb: 'DynamoDB',
  cloudwatch_logs: 'CloudWatch Logs',
  iam_role: 'IAMロール',
  iam_user: 'IAMユーザー',
};

// Get node label with fallback
export const getNodeLabel = (node: GraphNode): string => {
  return JAPANESE_LABELS[node.type] || node.name || node.id.split(':').pop() || node.id;
};

// Create a flow node from a graph node
export const createFlowNode = (
  node: GraphNode,
  position: { x: number; y: number },
  nodeType: string = 'custom'
): Node => {
  return {
    id: node.id,
    type: nodeType,
    position,
    data: {
      label: getNodeLabel(node),
      type: node.type,
      provider: node.provider,
      region: node.region,
      metadata: node.metadata,
      tags: node.tags,
    },
  };
};

// Create a layer background node
export const createLayerBackground = (layer: LayerConfig, width: number = 1400): Node => {
  return {
    id: `layer-bg-${layer.id}`,
    type: 'default',
    position: { x: 50, y: layer.y },
    data: { label: '' },
    style: {
      width,
      height: layer.height,
      background: `${layer.color}15`,
      border: `2px solid ${layer.color}40`,
      borderRadius: '12px',
      zIndex: -1,
      pointerEvents: 'none',
    },
    draggable: false,
    selectable: false,
  };
};

// Create a layer label node
export const createLayerLabel = (layer: LayerConfig, xOffset: number = 60): Node => {
  return {
    id: `layer-label-${layer.id}`,
    type: 'default',
    position: { x: xOffset, y: layer.y + 10 },
    data: { label: layer.label },
    style: {
      background: 'transparent',
      border: 'none',
      fontSize: '14px',
      fontWeight: '700',
      color: layer.color,
      pointerEvents: 'none',
    },
    draggable: false,
    selectable: false,
  };
};

// Create an edge between two nodes
export const createEdge = (
  sourceId: string,
  targetId: string,
  options: {
    type?: string;
    animated?: boolean;
    style?: Record<string, any>;
    label?: string;
    markerColor?: string;
  } = {}
): Edge => {
  const {
    type = 'smoothstep',
    animated = false,
    style = {},
    label,
    markerColor = '#6b7280',
  } = options;

  return {
    id: `edge-${sourceId}-${targetId}`,
    source: sourceId,
    target: targetId,
    type,
    animated,
    style,
    label,
    markerEnd: { type: MarkerType.ArrowClosed, color: markerColor },
  };
};

// Group nodes by VPC
export const groupNodesByVPC = (nodes: GraphNode[]): {
  vpcGroups: Map<string, GraphNode[]>;
  standaloneNodes: GraphNode[];
} => {
  const vpcGroups = new Map<string, GraphNode[]>();
  const standaloneNodes: GraphNode[] = [];

  // Initialize VPC groups
  nodes.forEach(node => {
    if (node.type === 'vpc') {
      vpcGroups.set(node.id, []);
    }
  });

  // Assign nodes to VPC groups
  nodes.forEach(node => {
    if (node.type === 'vpc') return;

    const vpcId = node.metadata?.vpc_id;
    if (vpcId) {
      const vpcKey = `aws:vpc:${vpcId}`;
      if (vpcGroups.has(vpcKey)) {
        vpcGroups.get(vpcKey)!.push(node);
      } else {
        standaloneNodes.push(node);
      }
    } else {
      standaloneNodes.push(node);
    }
  });

  return { vpcGroups, standaloneNodes };
};

// Filter important resource types
export const filterImportantResources = (nodes: GraphNode[]): GraphNode[] => {
  const importantTypes = [
    'vpc', 'subnet', 'ec2', 'rds', 'elb', 'alb', 's3', 'security_group',
    'eks_cluster', 'eks_node_group', 'nat_gateway', 'waf_web_acl', 'waf_ip_set',
    'lambda', 'dynamodb', 'internet_gateway', 'igw'
  ];
  return nodes.filter(node => importantTypes.includes(node.type));
};
