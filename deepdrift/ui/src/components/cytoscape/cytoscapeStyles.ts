import type cytoscape from 'cytoscape';

// AWS Official Icons
const AWS_ICONS_BASE = 'https://raw.githubusercontent.com/awslabs/aws-icons-for-plantuml/main/dist';

export const AWS_ICONS = {
  vpc: `${AWS_ICONS_BASE}/Networking/VPC.png`,
  subnet: `${AWS_ICONS_BASE}/Networking/VPCSubnet.png`,
  security_group: `${AWS_ICONS_BASE}/SecurityIdentityCompliance/VPCSecurityGroup.png`,
  ec2: `${AWS_ICONS_BASE}/Compute/EC2.png`,
  rds: `${AWS_ICONS_BASE}/Database/RDS.png`,
  alb: `${AWS_ICONS_BASE}/NetworkingContentDelivery/ElasticLoadBalancing.png`,
  elb: `${AWS_ICONS_BASE}/NetworkingContentDelivery/ElasticLoadBalancing.png`,
  lambda: `${AWS_ICONS_BASE}/Compute/Lambda.png`,
  s3: `${AWS_ICONS_BASE}/Storage/SimpleStorageService.png`,
  eks_cluster: `${AWS_ICONS_BASE}/Compute/ElasticKubernetesService.png`,
  eks_node_group: `${AWS_ICONS_BASE}/Compute/EC2.png`,
  nat_gateway: `${AWS_ICONS_BASE}/Networking/VPCNATGateway.png`,
  waf_web_acl: `${AWS_ICONS_BASE}/SecurityIdentityCompliance/WAF.png`,
  waf_ip_set: `${AWS_ICONS_BASE}/SecurityIdentityCompliance/WAF.png`,
  internet_gateway: `${AWS_ICONS_BASE}/Networking/VPCInternetGateway.png`,
  dynamodb: `${AWS_ICONS_BASE}/Database/DynamoDB.png`,
  cloudwatch_logs: `${AWS_ICONS_BASE}/ManagementGovernance/CloudWatchLogs.png`,
  iam_role: `${AWS_ICONS_BASE}/SecurityIdentityCompliance/IAMRole.png`,
  iam_user: `${AWS_ICONS_BASE}/SecurityIdentityCompliance/IAMIdentityCenter.png`,
};

export const getCytoscapeStylesheet = (): cytoscape.StylesheetStyle[] => [
  // Base node style
  {
    selector: 'node',
    style: {
      'label': 'data(label)',
      'text-valign': 'bottom',
      'text-halign': 'center',
      'text-margin-y': 5,
      'color': '#ffffff',
      'font-size': 11,
      'font-family': 'Amazon Ember, -apple-system, sans-serif',
      'font-weight': 500,
      'background-color': '#232f3e',
      'border-width': 3,
      'border-color': '#545b64',
      'width': 60,
      'height': 60,
    } as any,
  },

  // VPC
  {
    selector: 'node[type="vpc"]',
    style: {
      'width': 90,
      'height': 90,
      'background-image': AWS_ICONS.vpc,
      'background-fit': 'contain',
      'background-clip': 'none',
      'border-color': '#0972d3',
      'border-width': 4,
    } as any,
  },

  // Subnet
  {
    selector: 'node[type="subnet"]',
    style: {
      'width': 70,
      'height': 70,
      'background-image': AWS_ICONS.subnet,
      'background-fit': 'contain',
      'background-clip': 'none',
      'border-color': '#7AA116',
    } as any,
  },

  // Security Group
  {
    selector: 'node[type="security_group"]',
    style: {
      'width': 65,
      'height': 65,
      'background-image': AWS_ICONS.security_group,
      'background-fit': 'contain',
      'background-clip': 'none',
      'border-color': '#DD344C',
    } as any,
  },

  // EC2
  {
    selector: 'node[type="ec2"]',
    style: {
      'width': 60,
      'height': 60,
      'background-image': AWS_ICONS.ec2,
      'background-fit': 'contain',
      'background-clip': 'none',
      'border-color': '#FF9900',
    } as any,
  },

  // EKS Cluster
  {
    selector: 'node[type="eks_cluster"]',
    style: {
      'width': 80,
      'height': 80,
      'background-image': AWS_ICONS.eks_cluster,
      'background-fit': 'contain',
      'background-clip': 'none',
      'border-color': '#ff9900',
      'border-width': 4,
    } as any,
  },

  // EKS Node Group
  {
    selector: 'node[type="eks_node_group"]',
    style: {
      'width': 65,
      'height': 65,
      'background-image': AWS_ICONS.eks_node_group,
      'background-fit': 'contain',
      'background-clip': 'none',
      'border-color': '#00acc1',
    } as any,
  },

  // RDS
  {
    selector: 'node[type="rds"]',
    style: {
      'width': 70,
      'height': 70,
      'background-image': AWS_ICONS.rds,
      'background-fit': 'contain',
      'background-clip': 'none',
      'border-color': '#527fff',
    } as any,
  },

  // ALB/ELB
  {
    selector: 'node[type="alb"], node[type="elb"]',
    style: {
      'width': 70,
      'height': 70,
      'background-image': AWS_ICONS.alb,
      'background-fit': 'contain',
      'background-clip': 'none',
      'border-color': '#8C4FFF',
    } as any,
  },

  // S3
  {
    selector: 'node[type="s3"]',
    style: {
      'width': 65,
      'height': 65,
      'background-image': AWS_ICONS.s3,
      'background-fit': 'contain',
      'background-clip': 'none',
      'border-color': '#569A31',
    } as any,
  },

  // NAT Gateway
  {
    selector: 'node[type="nat_gateway"]',
    style: {
      'width': 65,
      'height': 65,
      'background-image': AWS_ICONS.nat_gateway,
      'background-fit': 'contain',
      'background-clip': 'none',
      'border-color': '#7AA116',
    } as any,
  },

  // Internet Gateway
  {
    selector: 'node[type="internet_gateway"], node[type="igw"]',
    style: {
      'width': 70,
      'height': 70,
      'background-image': AWS_ICONS.internet_gateway,
      'background-fit': 'contain',
      'background-clip': 'none',
      'border-color': '#10b981',
    } as any,
  },

  // DynamoDB
  {
    selector: 'node[type="dynamodb"]',
    style: {
      'width': 65,
      'height': 65,
      'background-image': AWS_ICONS.dynamodb,
      'background-fit': 'contain',
      'background-clip': 'none',
      'border-color': '#527fff',
    } as any,
  },

  // Lambda
  {
    selector: 'node[type="lambda"]',
    style: {
      'width': 60,
      'height': 60,
      'background-image': AWS_ICONS.lambda,
      'background-fit': 'contain',
      'background-clip': 'none',
      'border-color': '#FF9900',
    } as any,
  },

  // WAF
  {
    selector: 'node[type="waf_web_acl"], node[type="waf_ip_set"]',
    style: {
      'width': 65,
      'height': 65,
      'background-image': AWS_ICONS.waf_web_acl,
      'background-fit': 'contain',
      'background-clip': 'none',
      'border-color': '#DD344C',
    } as any,
  },

  // CloudWatch Logs
  {
    selector: 'node[type="cloudwatch_logs"]',
    style: {
      'width': 55,
      'height': 55,
      'background-image': AWS_ICONS.cloudwatch_logs,
      'background-fit': 'contain',
      'background-clip': 'none',
      'border-color': '#FF9900',
    } as any,
  },

  // IAM Role/User
  {
    selector: 'node[type="iam_role"], node[type="iam_user"]',
    style: {
      'width': 50,
      'height': 50,
      'background-image': AWS_ICONS.iam_role,
      'background-fit': 'contain',
      'background-clip': 'none',
      'border-color': '#DD344C',
    } as any,
  },

  // Internet node (special)
  {
    selector: 'node[type="internet"]',
    style: {
      'width': 100,
      'height': 100,
      'label': '🌐 Internet',
      'text-valign': 'center',
      'text-halign': 'center',
      'font-size': 14,
      'font-weight': 600,
      'background-color': '#0f172a',
      'border-color': '#10b981',
      'border-width': 4,
      'shape': 'ellipse',
    } as any,
  },

  // Health status - healthy
  {
    selector: 'node[health="healthy"]',
    style: {
      'border-color': '#1d8102',
      'border-width': 4,
    } as any,
  },

  // Health status - warning
  {
    selector: 'node[health="warning"]',
    style: {
      'border-color': '#f59e0b',
      'border-width': 4,
    } as any,
  },

  // Health status - critical
  {
    selector: 'node[health="critical"]',
    style: {
      'border-color': '#d13212',
      'border-width': 4,
    } as any,
  },

  // Public access indicator
  {
    selector: 'node[public="true"]',
    style: {
      'border-style': 'double',
      'border-width': 5,
    } as any,
  },

  // Base edge style
  {
    selector: 'edge',
    style: {
      'width': 2,
      'line-color': '#545b64',
      'target-arrow-color': '#545b64',
      'target-arrow-shape': 'triangle',
      'curve-style': 'bezier',
      'arrow-scale': 1.5,
    } as any,
  },

  // Network connection
  {
    selector: 'edge[type="network"]',
    style: {
      'line-color': '#06b6d4',
      'target-arrow-color': '#06b6d4',
      'width': 3,
    } as any,
  },

  // Ownership/containment
  {
    selector: 'edge[type="ownership"]',
    style: {
      'line-color': '#0972d3',
      'target-arrow-color': '#0972d3',
      'width': 2,
      'line-style': 'dashed',
    } as any,
  },

  // Dependency
  {
    selector: 'edge[type="dependency"]',
    style: {
      'line-color': '#f59e0b',
      'target-arrow-color': '#f59e0b',
      'width': 2,
      'line-style': 'dotted',
    } as any,
  },

  // Selected state
  {
    selector: ':selected',
    style: {
      'border-width': 6,
      'border-color': '#ff9900',
      'overlay-color': '#ff9900',
      'overlay-padding': 10,
      'overlay-opacity': 0.3,
    } as any,
  },
];
