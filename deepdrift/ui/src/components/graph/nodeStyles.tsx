import React from 'react';

// AWS Official Icons from aws-icons-for-plantuml
const AWS_ICONS_BASE = 'https://raw.githubusercontent.com/awslabs/aws-icons-for-plantuml/main/dist';

export const AWS_ICONS: Record<string, string> = {
  vpc: `${AWS_ICONS_BASE}/Networking/VPC.png`,
  subnet: `${AWS_ICONS_BASE}/Networking/VPCSubnet.png`,
  security_group: `${AWS_ICONS_BASE}/SecurityIdentityCompliance/VPCSecurityGroup.png`,
  ec2: `${AWS_ICONS_BASE}/Compute/EC2.png`,
  rds: `${AWS_ICONS_BASE}/Database/RDS.png`,
  elb: `${AWS_ICONS_BASE}/NetworkingContentDelivery/ElasticLoadBalancing.png`,
  alb: `${AWS_ICONS_BASE}/NetworkingContentDelivery/ElasticLoadBalancing.png`,
  lambda: `${AWS_ICONS_BASE}/Compute/Lambda.png`,
  s3: `${AWS_ICONS_BASE}/Storage/SimpleStorageService.png`,
  cloudwatch: `${AWS_ICONS_BASE}/ManagementGovernance/CloudWatch.png`,
  iam: `${AWS_ICONS_BASE}/SecurityIdentityCompliance/IAM.png`,
  eks_cluster: `${AWS_ICONS_BASE}/Compute/ElasticKubernetesService.png`,
  eks_node_group: `${AWS_ICONS_BASE}/Compute/EC2.png`,
  nat_gateway: `${AWS_ICONS_BASE}/Networking/VPCNATGateway.png`,
  waf_web_acl: `${AWS_ICONS_BASE}/SecurityIdentityCompliance/WAF.png`,
  waf_ip_set: `${AWS_ICONS_BASE}/SecurityIdentityCompliance/WAF.png`,
};

export const getNodeIcon = (type: string) => {
  const iconUrl = AWS_ICONS[type] || AWS_ICONS.ec2;
  return (
    <img
      src={iconUrl}
      alt={type}
      style={{ width: '32px', height: '32px', objectFit: 'contain' }}
    />
  );
};

export const getNodeColor = (type: string): string => {
  const colors: Record<string, string> = {
    vpc: '#527FFF',
    subnet: '#7AA116',
    security_group: '#DD344C',
    ec2: '#FF9900',
    rds: '#527FFF',
    elb: '#8C4FFF',
    alb: '#8C4FFF',
    lambda: '#FF9900',
    s3: '#569A31',
    eks_cluster: '#FF9900',
    eks_node_group: '#FF9900',
    nat_gateway: '#7AA116',
    waf_web_acl: '#DD344C',
    waf_ip_set: '#DD344C',
    default: '#232F3E',
  };
  return colors[type] || colors.default;
};

export const CustomNode: React.FC<{ data: any }> = ({ data }) => {
  const color = getNodeColor(data.type);
  const isPublic = data.metadata?.public_ip || data.metadata?.is_public;
  const hasPublicAccess = data.metadata?.ingress_rules?.some((rule: any) =>
    rule.cidr_blocks?.includes('0.0.0.0/0')
  );

  // Special styling for Internet node
  if (data.type === 'internet') {
    return (
      <div style={{
        padding: '16px 20px',
        borderRadius: '50%',
        border: '3px solid #10b981',
        backgroundColor: '#0f172a',
        minWidth: '80px',
        minHeight: '80px',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        boxShadow: '0 0 20px rgba(16, 185, 129, 0.4)',
      }}>
        <div style={{ fontSize: '32px', marginBottom: '4px' }}>🌐</div>
        <div style={{ fontSize: '11px', fontWeight: 600, color: '#10b981' }}>Internet</div>
      </div>
    );
  }

  return (
    <div
      style={{
        padding: '12px 16px',
        borderRadius: '8px',
        border: `2px solid ${color}`,
        backgroundColor: '#1e293b',
        minWidth: '180px',
        boxShadow: hasPublicAccess ? '0 0 12px rgba(239, 68, 68, 0.5)' : '0 4px 6px rgba(0, 0, 0, 0.3)',
        position: 'relative',
      }}
    >
      {(isPublic || hasPublicAccess) && (
        <div style={{
          position: 'absolute',
          top: '-8px',
          right: '-8px',
          backgroundColor: hasPublicAccess ? '#ef4444' : '#f59e0b',
          color: 'white',
          fontSize: '9px',
          fontWeight: 600,
          padding: '2px 6px',
          borderRadius: '4px',
          boxShadow: '0 2px 4px rgba(0,0,0,0.3)',
        }}>
          {hasPublicAccess ? '🌐 PUBLIC' : '🔓 External'}
        </div>
      )}

      <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '4px' }}>
        <div style={{ color }}>{getNodeIcon(data.type)}</div>
        <div style={{ fontWeight: 600, fontSize: '14px', color: 'white' }}>{data.label}</div>
      </div>
      <div style={{ fontSize: '11px', color: '#94a3b8' }}>{data.type}</div>

      {data.metadata?.public_ip && (
        <div style={{ fontSize: '10px', color: '#fbbf24', marginTop: '4px', display: 'flex', alignItems: 'center', gap: '4px' }}>
          <span>🌐</span>
          <span>{data.metadata.public_ip}</span>
        </div>
      )}
      {data.metadata?.private_ip && (
        <div style={{ fontSize: '10px', color: '#94a3b8', marginTop: '2px', display: 'flex', alignItems: 'center', gap: '4px' }}>
          <span>🔒</span>
          <span>{data.metadata.private_ip}</span>
        </div>
      )}

      {data.type === 'security_group' && data.metadata?.ingress_rules && (
        <div style={{ fontSize: '10px', color: '#94a3b8', marginTop: '4px', borderTop: '1px solid #334155', paddingTop: '4px' }}>
          <div style={{ fontWeight: 600, marginBottom: '2px' }}>開放ポート:</div>
          {data.metadata.ingress_rules.slice(0, 3).map((rule: any, idx: number) => (
            <div key={idx} style={{ color: rule.cidr_blocks?.includes('0.0.0.0/0') ? '#ef4444' : '#94a3b8' }}>
              • {rule.protocol === '-1' ? 'All' : `${rule.protocol}:${rule.from_port || 'all'}`}
              {rule.cidr_blocks?.includes('0.0.0.0/0') && ' (全公開)'}
            </div>
          ))}
        </div>
      )}

      {data.region && (
        <div style={{ fontSize: '10px', color: '#64748b', marginTop: '4px' }}>
          {data.provider}:{data.region}
        </div>
      )}
    </div>
  );
};
