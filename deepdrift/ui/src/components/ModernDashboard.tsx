import React, { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import {
  AlertTriangle,
  Activity,
  TrendingUp,
  Shield,
  Zap,
  Clock,
  Server
} from 'lucide-react';
import {
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer
} from 'recharts';
import apiClient from '../api/client';
import type { DriftStats, ImpactStats, DriftEvent, HighImpactDrift } from '../api/client';
import ResourceGraph from './ResourceGraph';

const COLORS = {
  critical: '#ef4444',
  high: '#f97316',
  medium: '#eab308',
  low: '#22c55e',
  gradient1: '#8b5cf6',
  gradient2: '#ec4899',
  background: '#0f172a',
  card: '#1e293b',
  cardHover: '#334155',
};

const ModernDashboard: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'overview' | 'graph'>('overview');
  const [driftStats, setDriftStats] = useState<DriftStats | null>(null);
  const [impactStats, setImpactStats] = useState<ImpactStats | null>(null);
  const [recentDrifts, setRecentDrifts] = useState<DriftEvent[]>([]);
  const [highImpactDrifts, setHighImpactDrifts] = useState<HighImpactDrift[]>([]);
  const [loading, setLoading] = useState(true);
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date());

  useEffect(() => {
    loadData();
    const interval = setInterval(() => {
      loadData();
      setLastUpdate(new Date());
    }, 10000); // Update every 10 seconds for demo
    return () => clearInterval(interval);
  }, []);

  const loadData = async () => {
    try {
      const [driftsRes, impactRes, recentRes, highImpactRes] = await Promise.all([
        apiClient.getDriftStats(7),
        apiClient.getImpactStats(7),
        apiClient.listDrifts({ limit: 10 }),
        apiClient.getHighImpactDrifts(7, 10),
      ]);

      setDriftStats(driftsRes.stats);
      setImpactStats(impactRes.stats);
      setRecentDrifts(recentRes.drifts);
      setHighImpactDrifts(highImpactRes.drifts);
      setLoading(false);
    } catch (err) {
      console.error('Failed to load data:', err);
      setLoading(false);
    }
  };

  // Prepare chart data
  const severityData = driftStats ? [
    { name: 'Critical', value: driftStats.by_severity?.critical || 0, color: COLORS.critical },
    { name: 'High', value: driftStats.by_severity?.high || 0, color: COLORS.high },
    { name: 'Medium', value: driftStats.by_severity?.medium || 0, color: COLORS.medium },
    { name: 'Low', value: driftStats.by_severity?.low || 0, color: COLORS.low },
  ] : [];

  const typeData = driftStats ? [
    { name: 'Created', value: driftStats.by_type?.created || 0 },
    { name: 'Modified', value: driftStats.by_type?.modified || 0 },
    { name: 'Deleted', value: driftStats.by_type?.deleted || 0 },
  ] : [];

  const resourceTypeData = driftStats?.by_resource_type ?
    Object.entries(driftStats.by_resource_type).map(([name, value]) => ({
      name,
      value
    })).slice(0, 8) : [];

  if (loading) {
    return (
      <div style={{
        height: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: `linear-gradient(135deg, ${COLORS.gradient1}, ${COLORS.gradient2})`,
      }}>
        <motion.div
          animate={{ rotate: 360 }}
          transition={{ duration: 1, repeat: Infinity, ease: 'linear' }}
        >
          <Activity size={48} color="white" />
        </motion.div>
      </div>
    );
  }

  return (
    <div style={{
      minHeight: '100vh',
      backgroundColor: COLORS.background,
      color: 'white',
      fontFamily: "'Inter', sans-serif",
      padding: '20px',
    }}>
      {/* Header */}
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        style={{
          background: `linear-gradient(135deg, ${COLORS.gradient1}, ${COLORS.gradient2})`,
          padding: '40px',
          borderRadius: '16px',
          marginBottom: '30px',
          position: 'relative',
          overflow: 'hidden',
        }}
      >
        <div style={{
          position: 'absolute',
          top: 0,
          right: 0,
          width: '300px',
          height: '300px',
          background: 'rgba(255, 255, 255, 0.1)',
          borderRadius: '50%',
          filter: 'blur(60px)',
        }} />
        <h1 style={{ margin: 0, fontSize: '42px', fontWeight: 'bold', position: 'relative' }}>
          <Zap size={40} style={{ display: 'inline', marginRight: '15px' }} />
          DeepDrift Analytics
        </h1>
        <p style={{ margin: '10px 0 0 0', fontSize: '18px', opacity: 0.9, position: 'relative' }}>
          Real-time Terraform Drift Detection & Impact Analysis
        </p>
        <div style={{
          position: 'absolute',
          top: '20px',
          right: '20px',
          display: 'flex',
          alignItems: 'center',
          gap: '8px',
          background: 'rgba(255, 255, 255, 0.2)',
          padding: '8px 16px',
          borderRadius: '20px',
          backdropFilter: 'blur(10px)',
        }}>
          <Clock size={16} />
          <span style={{ fontSize: '14px' }}>Last update: {lastUpdate.toLocaleTimeString()}</span>
        </div>
      </motion.div>

      {/* Tab Navigation */}
      <div style={{
        display: 'flex',
        gap: '10px',
        marginBottom: '30px',
      }}>
        <button
          onClick={() => setActiveTab('overview')}
          style={{
            padding: '12px 24px',
            borderRadius: '8px',
            border: 'none',
            backgroundColor: activeTab === 'overview' ? COLORS.gradient1 : COLORS.card,
            color: 'white',
            fontSize: '16px',
            fontWeight: activeTab === 'overview' ? '600' : '400',
            cursor: 'pointer',
            transition: 'all 0.3s ease',
          }}
        >
          Overview
        </button>
        <button
          onClick={() => setActiveTab('graph')}
          style={{
            padding: '12px 24px',
            borderRadius: '8px',
            border: 'none',
            backgroundColor: activeTab === 'graph' ? COLORS.gradient1 : COLORS.card,
            color: 'white',
            fontSize: '16px',
            fontWeight: activeTab === 'graph' ? '600' : '400',
            cursor: 'pointer',
            transition: 'all 0.3s ease',
          }}
        >
          Resource Graph
        </button>
      </div>

      {activeTab === 'overview' && (
        <>
      {/* KPI Cards */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
        gap: '20px',
        marginBottom: '30px',
      }}>
        <KPICard
          icon={<Activity size={32} />}
          title="Total Drifts (7d)"
          value={driftStats?.total_count || 0}
          trend="+12%"
          color={COLORS.gradient1}
          delay={0}
        />
        <KPICard
          icon={<AlertTriangle size={32} />}
          title="Critical Events"
          value={driftStats?.by_severity?.critical || 0}
          trend="+5"
          color={COLORS.critical}
          delay={0.1}
          pulse
        />
        <KPICard
          icon={<TrendingUp size={32} />}
          title="Avg Blast Radius"
          value={impactStats?.avg_blast_radius?.toFixed(1) || '0.0'}
          suffix=" hops"
          color={COLORS.medium}
          delay={0.2}
        />
        <KPICard
          icon={<Shield size={32} />}
          title="Avg Affected Resources"
          value={impactStats?.avg_affected_resources?.toFixed(1) || '0.0'}
          color={COLORS.low}
          delay={0.3}
        />
      </div>

      {/* Charts Row */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fit, minmax(400px, 1fr))',
        gap: '20px',
        marginBottom: '30px',
      }}>
        <ChartCard title="Drift by Severity" delay={0.4}>
          <ResponsiveContainer width="100%" height={300}>
            <PieChart>
              <Pie
                data={severityData}
                cx="50%"
                cy="50%"
                innerRadius={60}
                outerRadius={100}
                paddingAngle={5}
                dataKey="value"
              >
                {severityData.map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={entry.color} />
                ))}
              </Pie>
              <Tooltip
                contentStyle={{
                  backgroundColor: COLORS.card,
                  border: 'none',
                  borderRadius: '8px',
                  color: 'white',
                }}
              />
              <Legend />
            </PieChart>
          </ResponsiveContainer>
        </ChartCard>

        <ChartCard title="Drift by Type" delay={0.5}>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={typeData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="name" stroke="#94a3b8" />
              <YAxis stroke="#94a3b8" />
              <Tooltip
                contentStyle={{
                  backgroundColor: COLORS.card,
                  border: 'none',
                  borderRadius: '8px',
                  color: 'white',
                }}
              />
              <Bar dataKey="value" fill={COLORS.gradient1} radius={[8, 8, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </ChartCard>

        <ChartCard title="Top Resource Types" delay={0.6}>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={resourceTypeData} layout="vertical">
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis type="number" stroke="#94a3b8" />
              <YAxis dataKey="name" type="category" stroke="#94a3b8" width={100} />
              <Tooltip
                contentStyle={{
                  backgroundColor: COLORS.card,
                  border: 'none',
                  borderRadius: '8px',
                  color: 'white',
                }}
              />
              <Bar dataKey="value" fill={COLORS.gradient2} radius={[0, 8, 8, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </ChartCard>
      </div>

      {/* Recent Drifts Table */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.7 }}
        style={{
          backgroundColor: COLORS.card,
          borderRadius: '16px',
          padding: '30px',
          marginBottom: '30px',
          boxShadow: '0 4px 6px rgba(0, 0, 0, 0.3)',
        }}
      >
        <h2 style={{
          margin: '0 0 20px 0',
          display: 'flex',
          alignItems: 'center',
          gap: '10px',
          fontSize: '24px',
        }}>
          <Server size={28} />
          Recent Drift Events
        </h2>
        <div style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'separate', borderSpacing: '0 8px' }}>
            <thead>
              <tr>
                <th style={tableHeaderStyle}>Time</th>
                <th style={tableHeaderStyle}>Resource</th>
                <th style={tableHeaderStyle}>Type</th>
                <th style={tableHeaderStyle}>Drift</th>
                <th style={tableHeaderStyle}>Severity</th>
                <th style={tableHeaderStyle}>User</th>
              </tr>
            </thead>
            <tbody>
              {recentDrifts.map((drift, index) => (
                <motion.tr
                  key={drift.id}
                  initial={{ opacity: 0, x: -20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: 0.8 + index * 0.05 }}
                  style={{
                    backgroundColor: COLORS.cardHover,
                    transition: 'all 0.3s ease',
                  }}
                  whileHover={{ scale: 1.02, backgroundColor: '#475569' }}
                >
                  <td style={tableCellStyle}>
                    {new Date(drift.timestamp).toLocaleString()}
                  </td>
                  <td style={{ ...tableCellStyle, fontFamily: 'monospace', fontSize: '13px' }}>
                    {drift.resource_id}
                  </td>
                  <td style={tableCellStyle}>
                    <Badge color={COLORS.gradient1}>{drift.resource_type}</Badge>
                  </td>
                  <td style={tableCellStyle}>
                    <Badge
                      color={
                        drift.type === 'deleted' ? COLORS.critical :
                        drift.type === 'created' ? COLORS.low :
                        COLORS.medium
                      }
                    >
                      {drift.type}
                    </Badge>
                  </td>
                  <td style={tableCellStyle}>
                    <SeverityBadge severity={drift.severity} />
                  </td>
                  <td style={{ ...tableCellStyle, fontSize: '13px' }}>
                    {drift.root_cause?.user_identity || '-'}
                  </td>
                </motion.tr>
              ))}
            </tbody>
          </table>
        </div>
      </motion.div>

      {/* High Impact Drifts */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.9 }}
        style={{
          backgroundColor: COLORS.card,
          borderRadius: '16px',
          padding: '30px',
          boxShadow: '0 4px 6px rgba(0, 0, 0, 0.3)',
        }}
      >
        <h2 style={{
          margin: '0 0 20px 0',
          display: 'flex',
          alignItems: 'center',
          gap: '10px',
          fontSize: '24px',
        }}>
          <AlertTriangle size={28} />
          High Impact Drifts
        </h2>
        <div style={{ display: 'grid', gap: '16px' }}>
          {highImpactDrifts.map((drift, index) => (
            <motion.div
              key={drift.id}
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: 1 + index * 0.05 }}
              whileHover={{ scale: 1.02, x: 5 }}
              style={{
                backgroundColor: COLORS.cardHover,
                borderRadius: '12px',
                padding: '20px',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                borderLeft: `4px solid ${getSeverityColor(drift.severity)}`,
                transition: 'all 0.3s ease',
              }}
            >
              <div style={{ flex: 1 }}>
                <div style={{ fontFamily: 'monospace', fontSize: '14px', marginBottom: '8px' }}>
                  {drift.resource_id}
                </div>
                <div style={{ fontSize: '13px', color: '#94a3b8' }}>
                  {drift.resource_type}
                </div>
              </div>
              <div style={{ display: 'flex', gap: '20px', alignItems: 'center' }}>
                <div style={{ textAlign: 'center' }}>
                  <div style={{ fontSize: '24px', fontWeight: 'bold', color: COLORS.medium }}>
                    {drift.affected_resource_count}
                  </div>
                  <div style={{ fontSize: '12px', color: '#94a3b8' }}>Affected</div>
                </div>
                <div style={{ textAlign: 'center' }}>
                  <div style={{ fontSize: '24px', fontWeight: 'bold', color: COLORS.high }}>
                    {drift.blast_radius}
                  </div>
                  <div style={{ fontSize: '12px', color: '#94a3b8' }}>Hops</div>
                </div>
                <SeverityBadge severity={drift.severity} large />
              </div>
            </motion.div>
          ))}
        </div>
      </motion.div>
        </>
      )}

      {activeTab === 'graph' && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.5 }}
        >
          <h2 style={{ fontSize: '24px', fontWeight: 'bold', marginBottom: '20px' }}>
            Resource Dependency Graph
          </h2>
          <ResourceGraph />
        </motion.div>
      )}
    </div>
  );
};

// Helper Components
const KPICard: React.FC<{
  icon: React.ReactNode;
  title: string;
  value: number | string;
  suffix?: string;
  trend?: string;
  color: string;
  delay: number;
  pulse?: boolean;
}> = ({ icon, title, value, suffix = '', trend, color, delay, pulse }) => (
  <motion.div
    initial={{ opacity: 0, scale: 0.9 }}
    animate={{ opacity: 1, scale: 1 }}
    transition={{ delay }}
    whileHover={{ scale: 1.05 }}
    style={{
      backgroundColor: COLORS.card,
      borderRadius: '16px',
      padding: '25px',
      position: 'relative',
      overflow: 'hidden',
      boxShadow: '0 4px 6px rgba(0, 0, 0, 0.3)',
      cursor: 'pointer',
    }}
  >
    <div style={{
      position: 'absolute',
      top: -20,
      right: -20,
      width: '120px',
      height: '120px',
      background: `radial-gradient(circle, ${color}40, transparent)`,
      borderRadius: '50%',
      filter: 'blur(20px)',
    }} />
    <div style={{ position: 'relative', zIndex: 1 }}>
      <div style={{ color, marginBottom: '15px', opacity: 0.9 }}>
        {icon}
      </div>
      <div style={{ fontSize: '14px', color: '#94a3b8', marginBottom: '10px' }}>
        {title}
      </div>
      <div style={{ fontSize: '36px', fontWeight: 'bold', marginBottom: '5px' }}>
        {value}{suffix}
      </div>
      {trend && (
        <div style={{
          fontSize: '14px',
          color: COLORS.low,
          display: 'flex',
          alignItems: 'center',
          gap: '5px',
        }}>
          <TrendingUp size={16} /> {trend}
        </div>
      )}
    </div>
    {pulse && (
      <motion.div
        animate={{ scale: [1, 1.2, 1], opacity: [1, 0, 1] }}
        transition={{ duration: 2, repeat: Infinity }}
        style={{
          position: 'absolute',
          top: 25,
          left: 25,
          width: '32px',
          height: '32px',
          borderRadius: '50%',
          border: `2px solid ${color}`,
        }}
      />
    )}
  </motion.div>
);

const ChartCard: React.FC<{
  title: string;
  children: React.ReactNode;
  delay: number;
}> = ({ title, children, delay }) => (
  <motion.div
    initial={{ opacity: 0, y: 20 }}
    animate={{ opacity: 1, y: 0 }}
    transition={{ delay }}
    style={{
      backgroundColor: COLORS.card,
      borderRadius: '16px',
      padding: '25px',
      boxShadow: '0 4px 6px rgba(0, 0, 0, 0.3)',
    }}
  >
    <h3 style={{ margin: '0 0 20px 0', fontSize: '18px', fontWeight: '600' }}>{title}</h3>
    {children}
  </motion.div>
);

const Badge: React.FC<{ children: React.ReactNode; color: string }> = ({ children, color }) => (
  <span style={{
    display: 'inline-block',
    padding: '4px 12px',
    borderRadius: '12px',
    fontSize: '12px',
    fontWeight: '600',
    backgroundColor: `${color}30`,
    color,
    border: `1px solid ${color}60`,
  }}>
    {children}
  </span>
);

const SeverityBadge: React.FC<{ severity: string; large?: boolean }> = ({ severity, large }) => {
  const color = getSeverityColor(severity);
  return (
    <Badge color={color}>
      <span style={{ fontSize: large ? '14px' : '12px' }}>
        {severity.toUpperCase()}
      </span>
    </Badge>
  );
};

const getSeverityColor = (severity: string): string => {
  switch (severity.toLowerCase()) {
    case 'critical': return COLORS.critical;
    case 'high': return COLORS.high;
    case 'medium': return COLORS.medium;
    case 'low': return COLORS.low;
    default: return '#6b7280';
  }
};

const tableHeaderStyle: React.CSSProperties = {
  padding: '12px 16px',
  textAlign: 'left',
  fontSize: '13px',
  fontWeight: '600',
  color: '#94a3b8',
  textTransform: 'uppercase',
  letterSpacing: '0.5px',
};

const tableCellStyle: React.CSSProperties = {
  padding: '16px',
  borderRadius: '8px',
};

export default ModernDashboard;
