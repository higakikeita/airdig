import React, { useEffect, useState } from 'react';
import { Activity, AlertTriangle, Shield, TrendingUp } from 'lucide-react';
import ResourceGraph from './ResourceGraph';
import DriftTable from './DriftTable';
import TimeSeriesChart from './TimeSeriesChart';
import apiClient from '../api/client';
import type { DriftStats, ImpactStats } from '../api/client';

const CloudscapeDashboard: React.FC = () => {
  const [activeView, setActiveView] = useState<'graph' | 'table' | 'analytics'>('graph');
  const [driftStats, setDriftStats] = useState<DriftStats | null>(null);
  const [impactStats, setImpactStats] = useState<ImpactStats | null>(null);

  useEffect(() => {
    loadData();
    const interval = setInterval(() => {
      loadData();
    }, 30000); // Update every 30 seconds
    return () => clearInterval(interval);
  }, []);

  const loadData = async () => {
    try {
      const [driftsRes, impactRes] = await Promise.all([
        apiClient.getDriftStats(7),
        apiClient.getImpactStats(7),
      ]);

      setDriftStats(driftsRes.stats);
      setImpactStats(impactRes.stats);
    } catch (err) {
      console.error('Failed to load data:', err);
    }
  };

  return (
    <div style={{ display: 'flex', height: '100%', overflow: 'hidden' }}>
        {/* Sidebar Navigation - Always compact */}
        <div style={{
          width: '180px',
          background: 'var(--color-background-container-content)',
          borderRight: '1px solid var(--color-border-container-divider)',
          overflowY: 'auto',
          display: 'flex',
          flexDirection: 'column',
        }}>
          {/* Navigation Header - Minimal */}
          <div style={{
            padding: '12px 16px',
            borderBottom: '1px solid var(--color-border-container-divider)',
          }}>
            <div style={{
              fontSize: '13px',
              fontWeight: '700',
              color: 'white',
              marginBottom: '2px',
            }}>
              Dashboard
            </div>
            <div style={{
              fontSize: '10px',
              color: 'var(--color-text-body-secondary)',
            }}>
              Real-time monitoring
            </div>
          </div>

          {/* Navigation Sections */}
          <div style={{ padding: '12px 12px' }}>
            <div style={{
              fontSize: '10px',
              fontWeight: '700',
              color: 'var(--color-text-label)',
              textTransform: 'uppercase',
              letterSpacing: '0.5px',
              marginBottom: '6px',
              paddingLeft: '4px',
            }}>
              Views
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
              <NavItem
                label="Graph"
                active={activeView === 'graph'}
                onClick={() => setActiveView('graph')}
              />
              <NavItem
                label="Table"
                active={activeView === 'table'}
                onClick={() => setActiveView('table')}
              />
              <NavItem
                label="Analytics"
                active={activeView === 'analytics'}
                onClick={() => setActiveView('analytics')}
              />
            </div>
          </div>

          {/* Stats Section - Compact */}
          <div style={{ padding: '12px 12px', flex: 1 }}>
            <div style={{
              fontSize: '10px',
              fontWeight: '700',
              color: 'var(--color-text-label)',
              textTransform: 'uppercase',
              letterSpacing: '0.5px',
              marginBottom: '8px',
              paddingLeft: '4px',
            }}>
              Summary (7d)
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
              <StatItem
                icon={<Activity size={14} />}
                label="Drifts"
                value={driftStats?.total_count || 0}
              />
              <StatItem
                icon={<AlertTriangle size={14} />}
                label="Critical"
                value={driftStats?.by_severity?.critical || 0}
                statusColor="#ef4444"
              />
              <StatItem
                icon={<TrendingUp size={14} />}
                label="Blast Radius"
                value={impactStats?.avg_blast_radius?.toFixed(1) || '0.0'}
                statusColor="#f59e0b"
              />
              <StatItem
                icon={<Shield size={14} />}
                label="Affected"
                value={impactStats?.avg_affected_resources?.toFixed(0) || '0'}
                statusColor="#06b6d4"
              />
            </div>
          </div>
        </div>

        {/* Main Content */}
        <div style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
          {/* Content Header - Compact for all views */}
          {activeView !== 'graph' && (
            <div style={{
              background: 'var(--color-background-container-header)',
              padding: '12px 20px',
              borderBottom: '1px solid var(--color-border-container-divider)',
            }}>
              <h1 style={{ fontSize: '16px', fontWeight: '700', color: 'white', margin: 0 }}>
                {activeView === 'table' ? 'Drift Events' : 'Analytics'}
              </h1>
              <p style={{
                fontSize: '11px',
                color: 'var(--color-text-body-secondary)',
                margin: '2px 0 0 0',
              }}>
                {activeView === 'analytics' ? 'Time series analysis & trends' : 'Recent drift detection results'}
              </p>
            </div>
          )}

          {/* Content Area - Minimal padding */}
          <div style={{
            flex: 1,
            background: 'var(--color-background-layout-main)',
            padding: activeView === 'graph' ? 0 : '16px',
            overflow: activeView === 'graph' ? 'hidden' : 'auto'
          }}>
            {activeView === 'graph' && (
              <div style={{ height: '100%', width: '100%' }}>
                <ResourceGraph />
              </div>
            )}
            {activeView === 'table' && (
              <div style={{ height: '100%' }}>
                <DriftTable />
              </div>
            )}
            {activeView === 'analytics' && (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-l)' }}>
                <TimeSeriesChart />
              </div>
            )}
          </div>
        </div>
      </div>
  );
};

// Helper Components
const NavItem: React.FC<{ label: string; active: boolean; onClick: () => void }> = ({ label, active, onClick }) => (
  <div
    onClick={onClick}
    style={{
      padding: '6px 8px',
      borderRadius: '4px',
      cursor: 'pointer',
      background: active ? 'var(--color-background-container-header)' : 'transparent',
      color: active ? 'var(--aws-smile-orange)' : 'var(--color-text-body-default)',
      fontWeight: active ? '600' : '400',
      fontSize: '12px',
      transition: 'all 0.15s ease-in-out',
      borderLeft: active ? '2px solid var(--aws-smile-orange)' : '2px solid transparent',
    }}
  >
    {label}
  </div>
);

const StatItem: React.FC<{ icon: React.ReactNode; label: string; value: string | number; statusColor?: string }> = ({
  icon,
  label,
  value,
  statusColor,
}) => (
  <div style={{
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
    padding: '6px 8px',
    background: 'var(--color-background-container-header)',
    borderRadius: '4px',
    border: `1px solid ${statusColor || 'transparent'}`,
  }}>
    <div style={{ color: statusColor || 'var(--color-text-body-secondary)', flexShrink: 0 }}>
      {icon}
    </div>
    <div style={{ flex: 1, minWidth: 0 }}>
      <div style={{ fontSize: '10px', color: 'var(--color-text-label)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
        {label}
      </div>
      <div style={{ fontSize: '16px', fontWeight: '700', color: statusColor || 'white', lineHeight: '1.2' }}>
        {value}
      </div>
    </div>
  </div>
);

export default CloudscapeDashboard;
