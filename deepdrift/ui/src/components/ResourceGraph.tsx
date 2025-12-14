import React, { useEffect, useState } from 'react';
import ReactFlow, {
  Controls,
  Background,
  useNodesState,
  useEdgesState,
  MarkerType,
} from 'reactflow';
import type { Node, Edge } from 'reactflow';
import 'reactflow/dist/style.css';

interface GraphNode {
  id: string;
  type: string;
  provider: string;
  region: string;
  name: string;
  metadata?: Record<string, any>;
  tags?: Record<string, string>;
}

interface GraphEdge {
  from: string;
  to: string;
  type: string;
  metadata?: Record<string, any>;
}

interface ResourceGraphData {
  nodes: GraphNode[];
  edges: GraphEdge[];
}

// AWS Official Icons from aws-icons-for-plantuml
const AWS_ICONS_BASE = 'https://raw.githubusercontent.com/awslabs/aws-icons-for-plantuml/main/dist';

const AWS_ICONS: Record<string, string> = {
  vpc: `${AWS_ICONS_BASE}/Networking/VPC.png`,
  subnet: `${AWS_ICONS_BASE}/Networking/VPCSubnet.png`,
  security_group: `${AWS_ICONS_BASE}/SecurityIdentityCompliance/VPCSecurityGroup.png`,
  ec2: `${AWS_ICONS_BASE}/Compute/EC2.png`,
  rds: `${AWS_ICONS_BASE}/Database/RDS.png`,
  elb: `${AWS_ICONS_BASE}/NetworkingContentDelivery/ElasticLoadBalancing.png`,
  lambda: `${AWS_ICONS_BASE}/Compute/Lambda.png`,
  s3: `${AWS_ICONS_BASE}/Storage/SimpleStorageService.png`,
  cloudwatch: `${AWS_ICONS_BASE}/ManagementGovernance/CloudWatch.png`,
  iam: `${AWS_ICONS_BASE}/SecurityIdentityCompliance/IAM.png`,
};

const getNodeIcon = (type: string) => {
  const iconUrl = AWS_ICONS[type] || AWS_ICONS.ec2;
  return (
    <img
      src={iconUrl}
      alt={type}
      style={{ width: '32px', height: '32px', objectFit: 'contain' }}
    />
  );
};

const getNodeColor = (type: string) => {
  // AWS brand colors and service-specific colors
  const colors: Record<string, string> = {
    vpc: '#527FFF', // AWS VPC Blue
    subnet: '#7AA116', // AWS Networking Green
    security_group: '#DD344C', // AWS Security Red
    ec2: '#FF9900', // AWS Orange
    rds: '#527FFF', // AWS Database Blue
    elb: '#8C4FFF', // AWS Network Purple
    lambda: '#FF9900', // AWS Compute Orange
    s3: '#569A31', // AWS Storage Green
    default: '#232F3E', // AWS Dark
  };
  return colors[type] || colors.default;
};

const CustomNode: React.FC<{ data: any }> = ({ data }) => {
  const color = getNodeColor(data.type);

  return (
    <div
      style={{
        padding: '12px 16px',
        borderRadius: '8px',
        border: `2px solid ${color}`,
        backgroundColor: '#1e293b',
        minWidth: '180px',
        boxShadow: '0 4px 6px rgba(0, 0, 0, 0.3)',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '4px' }}>
        <div style={{ color }}>{getNodeIcon(data.type)}</div>
        <div style={{ fontWeight: 600, fontSize: '14px', color: 'white' }}>{data.label}</div>
      </div>
      <div style={{ fontSize: '11px', color: '#94a3b8' }}>{data.type}</div>
      {data.region && (
        <div style={{ fontSize: '10px', color: '#64748b', marginTop: '4px' }}>
          {data.provider}:{data.region}
        </div>
      )}
    </div>
  );
};

const nodeTypes = {
  custom: CustomNode,
};

export const ResourceGraph: React.FC = () => {
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadGraphData();
  }, []);

  const loadGraphData = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/v1/graph');
      const data: ResourceGraphData = await response.json();

      // Convert to React Flow format
      const flowNodes: Node[] = data.nodes.map((node, index) => ({
        id: node.id,
        type: 'custom',
        position: calculatePosition(index, data.nodes.length),
        data: {
          label: node.name,
          type: node.type,
          provider: node.provider,
          region: node.region,
          metadata: node.metadata,
          tags: node.tags,
        },
      }));

      const flowEdges: Edge[] = data.edges.map((edge, index) => ({
        id: `e-${index}`,
        source: edge.from,
        target: edge.to,
        type: 'smoothstep',
        animated: edge.type === 'dependency',
        label: edge.type,
        labelStyle: { fill: '#94a3b8', fontSize: 10 },
        labelBgStyle: { fill: '#0f172a' },
        style: { stroke: getEdgeColor(edge.type), strokeWidth: 2 },
        markerEnd: {
          type: MarkerType.ArrowClosed,
          color: getEdgeColor(edge.type),
        },
      }));

      setNodes(flowNodes);
      setEdges(flowEdges);
      setLoading(false);
    } catch (error) {
      console.error('Failed to load graph data:', error);
      setLoading(false);
    }
  };

  const calculatePosition = (index: number, total: number) => {
    // Simple circular layout
    const radius = 250;
    const angle = (2 * Math.PI * index) / total;
    return {
      x: 400 + radius * Math.cos(angle),
      y: 300 + radius * Math.sin(angle),
    };
  };

  const getEdgeColor = (type: string) => {
    const colors: Record<string, string> = {
      ownership: '#8b5cf6',
      network: '#06b6d4',
      dependency: '#f59e0b',
      default: '#475569',
    };
    return colors[type] || colors.default;
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '600px', color: 'white' }}>
        Loading graph...
      </div>
    );
  }

  return (
    <div style={{ height: '600px', backgroundColor: '#0f172a', borderRadius: '12px', overflow: 'hidden' }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        nodeTypes={nodeTypes}
        fitView
        attributionPosition="bottom-left"
        style={{ background: '#0f172a' }}
      >
        <Background color="#1e293b" gap={16} />
        <Controls />
      </ReactFlow>
    </div>
  );
};

export default ResourceGraph;
