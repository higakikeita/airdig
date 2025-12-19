import type { Node, Edge } from 'reactflow';
import type { GraphNode } from './graphUtils';
import { groupNodesByVPC, createFlowNode, createEdge } from './graphUtils';

export const buildArchitectureLayout = (filteredNodes: GraphNode[]): { nodes: Node[]; edges: Edge[] } => {
  const flowNodes: Node[] = [];
  const flowEdges: Edge[] = [];

  const { vpcGroups, standaloneNodes } = groupNodesByVPC(filteredNodes);
  const internetGateways = filteredNodes.filter(n => n.type === 'internet_gateway' || n.type === 'igw');
  const natGateways = filteredNodes.filter(n => n.type === 'nat_gateway');

  let yOffset = 150;

  // Layout VPCs and their resources
  vpcGroups.forEach((resources, vpcId) => {
    const vpc = filteredNodes.find(n => n.id === vpcId);
    if (!vpc) return;

    flowNodes.push(createFlowNode(vpc, { x: 100, y: yOffset }));

    const subnets = resources.filter(r => r.type === 'subnet');
    const instances = resources.filter(r => r.type === 'ec2');
    const securityGroups = resources.filter(r => r.type === 'security_group');

    let xOffset = 400;

    // Layout subnets
    subnets.forEach((subnet, idx) => {
      const subnetY = yOffset + idx * 180;
      flowNodes.push(createFlowNode(subnet, { x: xOffset, y: subnetY }));

      flowEdges.push(createEdge(vpc.id, subnet.id, {
        style: { stroke: '#7AA116', strokeWidth: 2 },
        markerColor: '#7AA116',
      }));

      // Layout instances in subnet
      const subnetInstances = instances.filter(inst =>
        inst.metadata?.subnet_id === subnet.id.split(':').pop()
      );

      subnetInstances.forEach((inst, instIdx) => {
        const instX = xOffset + 250 + instIdx * 200;
        flowNodes.push(createFlowNode(inst, { x: instX, y: subnetY }));

        flowEdges.push(createEdge(subnet.id, inst.id, {
          style: { stroke: '#FF9900', strokeWidth: 2 },
          markerColor: '#FF9900',
        }));
      });
    });

    // Layout security groups
    securityGroups.forEach((sg, idx) => {
      const sgY = yOffset + 50;
      const sgX = xOffset + 600 + idx * 180;
      flowNodes.push(createFlowNode(sg, { x: sgX, y: sgY }));
    });

    yOffset += subnets.length * 180 + 100;
  });

  // Layout standalone nodes
  standaloneNodes.forEach((node, idx) => {
    const col = idx % 5;
    const row = Math.floor(idx / 5);
    flowNodes.push(createFlowNode(node, {
      x: 100 + col * 250,
      y: yOffset + row * 150
    }));
  });

  // Layout NAT gateways
  natGateways.forEach((nat, idx) => {
    flowNodes.push(createFlowNode(nat, {
      x: 100 + idx * 250,
      y: 50
    }));
  });

  // Layout internet gateways
  internetGateways.forEach((igw, idx) => {
    flowNodes.push(createFlowNode(igw, {
      x: 600 + idx * 250,
      y: 50
    }));
  });

  return { nodes: flowNodes, edges: flowEdges };
};
