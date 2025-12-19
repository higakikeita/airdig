import type { Node, Edge } from 'reactflow';
import type { GraphNode } from './graphUtils';
import {
  SERVICE_LAYERS,
  getResourceLayer,
  createFlowNode,
  createLayerBackground,
  createLayerLabel,
  createEdge,
} from './graphUtils';

export const buildLayersLayout = (filteredNodes: GraphNode[]): { nodes: Node[]; edges: Edge[] } => {
  const flowNodes: Node[] = [];
  const flowEdges: Edge[] = [];

  // Group resources by layer
  const layerGroups = new Map<string, GraphNode[]>();

  // Initialize layer groups
  Object.values(SERVICE_LAYERS).forEach(layer => {
    layerGroups.set(layer.id, []);
  });

  // Classify each resource into a layer
  filteredNodes.forEach(node => {
    const layer = getResourceLayer(node.type);
    if (layer) {
      layerGroups.get(layer.id)!.push(node);
    }
  });

  // Add layer background and labels
  Object.values(SERVICE_LAYERS).forEach(layer => {
    const nodesInLayer = layerGroups.get(layer.id)!;
    if (nodesInLayer.length > 0) {
      flowNodes.push(createLayerBackground(layer));
      flowNodes.push(createLayerLabel(layer));
    }
  });

  // Layout resources within each layer
  Object.values(SERVICE_LAYERS).forEach(layer => {
    const nodesInLayer = layerGroups.get(layer.id)!;
    if (nodesInLayer.length === 0) return;

    const baseY = layer.y + 60;
    const spacing = 200;
    let xOffset = 200;

    // Group by VPC within layer for network-related resources
    if (layer.id === 'network' || layer.id === 'compute') {
      const vpcGroups = new Map<string, GraphNode[]>();
      const nonVpcNodes: GraphNode[] = [];

      nodesInLayer.forEach(node => {
        if (node.type === 'vpc' || node.type === 'subnet') {
          const vpcId = node.type === 'vpc' ? node.id : node.metadata?.vpc_id;
          if (vpcId) {
            const vpcKey = node.type === 'vpc' ? vpcId : `aws:vpc:${vpcId}`;
            if (!vpcGroups.has(vpcKey)) {
              vpcGroups.set(vpcKey, []);
            }
            vpcGroups.get(vpcKey)!.push(node);
          } else {
            nonVpcNodes.push(node);
          }
        } else {
          nonVpcNodes.push(node);
        }
      });

      // Layout VPC groups
      vpcGroups.forEach((vpcNodes) => {
        const vpc = vpcNodes.find(n => n.type === 'vpc');
        const subnets = vpcNodes.filter(n => n.type === 'subnet');

        if (vpc) {
          flowNodes.push(createFlowNode(vpc, { x: xOffset, y: baseY }));
          xOffset += spacing;
        }

        subnets.forEach(subnet => {
          flowNodes.push(createFlowNode(subnet, { x: xOffset, y: baseY }));
          xOffset += spacing;
        });
      });

      // Layout non-VPC nodes
      nonVpcNodes.forEach((node, idx) => {
        const row = Math.floor(idx / 6);
        const col = idx % 6;
        flowNodes.push(createFlowNode(node, {
          x: xOffset + col * spacing,
          y: baseY + row * 100
        }));
      });
    } else {
      // Simple grid layout for other layers
      nodesInLayer.forEach((node, idx) => {
        const row = Math.floor(idx / 6);
        const col = idx % 6;
        flowNodes.push(createFlowNode(node, {
          x: 200 + col * spacing,
          y: baseY + row * 100
        }));
      });
    }
  });

  // Create connections between layers
  addLayerConnections(filteredNodes, flowNodes, flowEdges);

  return { nodes: flowNodes, edges: flowEdges };
};

const addLayerConnections = (
  filteredNodes: GraphNode[],
  flowNodes: Node[],
  flowEdges: Edge[]
): void => {
  filteredNodes.forEach(node => {
    // Connect ALB to EC2 instances
    if (node.type === 'alb' || node.type === 'elb') {
      filteredNodes.forEach(targetNode => {
        if (targetNode.type === 'ec2') {
          flowEdges.push(createEdge(node.id, targetNode.id, {
            style: { stroke: '#06b6d4', strokeWidth: 2 },
            markerColor: '#06b6d4',
          }));
        }
      });
    }

    // Connect EC2 to RDS/S3
    if (node.type === 'ec2') {
      const vpcId = node.metadata?.vpc_id;
      filteredNodes.forEach(targetNode => {
        if ((targetNode.type === 'rds' || targetNode.type === 's3') &&
            targetNode.metadata?.vpc_id === vpcId) {
          flowEdges.push(createEdge(node.id, targetNode.id, {
            style: { stroke: '#f59e0b', strokeWidth: 2, strokeDasharray: '5,5' },
            markerColor: '#f59e0b',
          }));
        }
      });
    }

    // Connect NAT Gateway to Internet
    if (node.type === 'nat_gateway') {
      let internetNode = flowNodes.find(n => n.data.type === 'internet');
      if (!internetNode) {
        internetNode = {
          id: 'internet-global',
          type: 'custom',
          position: { x: 700, y: 80 },
          data: {
            label: 'インターネット',
            type: 'internet',
            provider: 'external',
            region: '',
            metadata: {},
          },
        };
        flowNodes.push(internetNode);
      }

      flowEdges.push(createEdge(node.id, 'internet-global', {
        animated: true,
        style: { stroke: '#10b981', strokeWidth: 3 },
        markerColor: '#10b981',
        label: '外部通信',
      }));
    }
  });
};
