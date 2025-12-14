import React, { useEffect, useState } from 'react';
import apiClient from '../api/client';
import type { DriftStats, ImpactStats, DriftEvent, HighImpactDrift } from '../api/client';

const Dashboard: React.FC = () => {
  const [driftStats, setDriftStats] = useState<DriftStats | null>(null);
  const [impactStats, setImpactStats] = useState<ImpactStats | null>(null);
  const [recentDrifts, setRecentDrifts] = useState<DriftEvent[]>([]);
  const [highImpactDrifts, setHighImpactDrifts] = useState<HighImpactDrift[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadData();
    const interval = setInterval(loadData, 30000); // Refresh every 30 seconds
    return () => clearInterval(interval);
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
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
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load data');
    } finally {
      setLoading(false);
    }
  };

  const getSeverityColor = (severity: string): string => {
    switch (severity) {
      case 'critical': return '#dc2626';
      case 'high': return '#ea580c';
      case 'medium': return '#ca8a04';
      case 'low': return '#65a30d';
      default: return '#6b7280';
    }
  };

  if (loading && !driftStats) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <div>Loading...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ padding: '20px', color: '#dc2626' }}>
        <h2>Error</h2>
        <p>{error}</p>
        <button onClick={loadData}>Retry</button>
      </div>
    );
  }

  return (
    <div style={{ padding: '20px', fontFamily: 'sans-serif' }}>
      <h1>DeepDrift Dashboard</h1>
      <p style={{ color: '#6b7280' }}>Real-time Terraform Drift Detection & Impact Analysis</p>

      {/* KPI Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '20px', marginTop: '30px' }}>
        <div style={{ padding: '20px', border: '1px solid #e5e7eb', borderRadius: '8px', backgroundColor: '#f9fafb' }}>
          <h3 style={{ margin: 0, fontSize: '14px', color: '#6b7280' }}>Total Drifts (7d)</h3>
          <p style={{ margin: '10px 0 0 0', fontSize: '32px', fontWeight: 'bold' }}>{driftStats?.total_count || 0}</p>
        </div>

        <div style={{ padding: '20px', border: '1px solid #e5e7eb', borderRadius: '8px', backgroundColor: '#fef2f2' }}>
          <h3 style={{ margin: 0, fontSize: '14px', color: '#6b7280' }}>Critical</h3>
          <p style={{ margin: '10px 0 0 0', fontSize: '32px', fontWeight: 'bold', color: '#dc2626' }}>
            {driftStats?.by_severity?.critical || 0}
          </p>
        </div>

        <div style={{ padding: '20px', border: '1px solid #e5e7eb', borderRadius: '8px', backgroundColor: '#f9fafb' }}>
          <h3 style={{ margin: 0, fontSize: '14px', color: '#6b7280' }}>Avg Blast Radius</h3>
          <p style={{ margin: '10px 0 0 0', fontSize: '32px', fontWeight: 'bold' }}>
            {impactStats?.avg_blast_radius?.toFixed(1) || '0.0'}
          </p>
        </div>

        <div style={{ padding: '20px', border: '1px solid #e5e7eb', borderRadius: '8px', backgroundColor: '#f9fafb' }}>
          <h3 style={{ margin: 0, fontSize: '14px', color: '#6b7280' }}>Avg Affected Resources</h3>
          <p style={{ margin: '10px 0 0 0', fontSize: '32px', fontWeight: 'bold' }}>
            {impactStats?.avg_affected_resources?.toFixed(1) || '0.0'}
          </p>
        </div>
      </div>

      {/* Recent Drifts */}
      <div style={{ marginTop: '40px' }}>
        <h2>Recent Drift Events</h2>
        <div style={{ border: '1px solid #e5e7eb', borderRadius: '8px', overflow: 'hidden' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ backgroundColor: '#f9fafb' }}>
                <th style={{ padding: '12px', textAlign: 'left', borderBottom: '1px solid #e5e7eb' }}>Time</th>
                <th style={{ padding: '12px', textAlign: 'left', borderBottom: '1px solid #e5e7eb' }}>Resource</th>
                <th style={{ padding: '12px', textAlign: 'left', borderBottom: '1px solid #e5e7eb' }}>Type</th>
                <th style={{ padding: '12px', textAlign: 'left', borderBottom: '1px solid #e5e7eb' }}>Drift</th>
                <th style={{ padding: '12px', textAlign: 'left', borderBottom: '1px solid #e5e7eb' }}>Severity</th>
                <th style={{ padding: '12px', textAlign: 'left', borderBottom: '1px solid #e5e7eb' }}>User</th>
              </tr>
            </thead>
            <tbody>
              {recentDrifts.length === 0 ? (
                <tr>
                  <td colSpan={6} style={{ padding: '20px', textAlign: 'center', color: '#6b7280' }}>
                    No drift events found
                  </td>
                </tr>
              ) : (
                recentDrifts.map((drift) => (
                  <tr key={drift.id} style={{ borderBottom: '1px solid #f3f4f6' }}>
                    <td style={{ padding: '12px' }}>{new Date(drift.timestamp).toLocaleString()}</td>
                    <td style={{ padding: '12px', fontFamily: 'monospace', fontSize: '12px' }}>{drift.resource_id}</td>
                    <td style={{ padding: '12px' }}>{drift.resource_type}</td>
                    <td style={{ padding: '12px' }}>
                      <span style={{
                        padding: '4px 8px',
                        borderRadius: '4px',
                        fontSize: '12px',
                        backgroundColor: drift.type === 'deleted' ? '#fef2f2' : drift.type === 'created' ? '#f0fdf4' : '#fef3c7',
                        color: drift.type === 'deleted' ? '#dc2626' : drift.type === 'created' ? '#16a34a' : '#ca8a04'
                      }}>
                        {drift.type}
                      </span>
                    </td>
                    <td style={{ padding: '12px' }}>
                      <span style={{
                        padding: '4px 8px',
                        borderRadius: '4px',
                        fontSize: '12px',
                        backgroundColor: getSeverityColor(drift.severity) + '20',
                        color: getSeverityColor(drift.severity)
                      }}>
                        {drift.severity}
                      </span>
                    </td>
                    <td style={{ padding: '12px', fontSize: '12px' }}>
                      {drift.root_cause?.user_identity || '-'}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* High Impact Drifts */}
      <div style={{ marginTop: '40px' }}>
        <h2>High Impact Drifts</h2>
        <div style={{ border: '1px solid #e5e7eb', borderRadius: '8px', overflow: 'hidden' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ backgroundColor: '#f9fafb' }}>
                <th style={{ padding: '12px', textAlign: 'left', borderBottom: '1px solid #e5e7eb' }}>Resource</th>
                <th style={{ padding: '12px', textAlign: 'left', borderBottom: '1px solid #e5e7eb' }}>Type</th>
                <th style={{ padding: '12px', textAlign: 'left', borderBottom: '1px solid #e5e7eb' }}>Affected</th>
                <th style={{ padding: '12px', textAlign: 'left', borderBottom: '1px solid #e5e7eb' }}>Blast Radius</th>
                <th style={{ padding: '12px', textAlign: 'left', borderBottom: '1px solid #e5e7eb' }}>Severity</th>
              </tr>
            </thead>
            <tbody>
              {highImpactDrifts.length === 0 ? (
                <tr>
                  <td colSpan={5} style={{ padding: '20px', textAlign: 'center', color: '#6b7280' }}>
                    No high impact drifts found
                  </td>
                </tr>
              ) : (
                highImpactDrifts.map((drift) => (
                  <tr key={drift.id} style={{ borderBottom: '1px solid #f3f4f6' }}>
                    <td style={{ padding: '12px', fontFamily: 'monospace', fontSize: '12px' }}>{drift.resource_id}</td>
                    <td style={{ padding: '12px' }}>{drift.resource_type}</td>
                    <td style={{ padding: '12px', fontWeight: 'bold' }}>{drift.affected_resource_count}</td>
                    <td style={{ padding: '12px', fontWeight: 'bold' }}>{drift.blast_radius} hops</td>
                    <td style={{ padding: '12px' }}>
                      <span style={{
                        padding: '4px 8px',
                        borderRadius: '4px',
                        fontSize: '12px',
                        backgroundColor: getSeverityColor(drift.severity) + '20',
                        color: getSeverityColor(drift.severity)
                      }}>
                        {drift.severity}
                      </span>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
