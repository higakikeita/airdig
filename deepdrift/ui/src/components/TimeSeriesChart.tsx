import React, { useEffect, useState } from 'react';
import { LineChart, Line, AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { TrendingUp } from 'lucide-react';
import apiClient from '../api/client';
import type { DriftEvent } from '../api/client';

interface TimeSeriesData {
  date: string;
  total: number;
  critical: number;
  high: number;
  medium: number;
  low: number;
  created: number;
  modified: number;
  deleted: number;
}

const TimeSeriesChart: React.FC = () => {
  const [data, setData] = useState<TimeSeriesData[]>([]);
  const [loading, setLoading] = useState(true);
  const [chartType, setChartType] = useState<'line' | 'area'>('area');
  const [dataView, setDataView] = useState<'severity' | 'type'>('severity');

  useEffect(() => {
    loadData();
    const interval = setInterval(() => {
      loadData();
    }, 60000); // Update every minute
    return () => clearInterval(interval);
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const response = await apiClient.listDrifts({ limit: 1000 });
      const drifts = response.drifts || [];

      // Aggregate data by date
      const aggregated = aggregateByDate(drifts);
      setData(aggregated);
    } catch (err) {
      console.error('Failed to load time series data:', err);
    } finally {
      setLoading(false);
    }
  };

  const aggregateByDate = (drifts: DriftEvent[]): TimeSeriesData[] => {
    const dateMap = new Map<string, TimeSeriesData>();

    // Get last 30 days
    const today = new Date();
    for (let i = 29; i >= 0; i--) {
      const date = new Date(today);
      date.setDate(date.getDate() - i);
      const dateStr = date.toISOString().split('T')[0];
      dateMap.set(dateStr, {
        date: dateStr,
        total: 0,
        critical: 0,
        high: 0,
        medium: 0,
        low: 0,
        created: 0,
        modified: 0,
        deleted: 0,
      });
    }

    // Aggregate drifts
    drifts.forEach(drift => {
      const dateStr = drift.timestamp.split('T')[0];
      const existing = dateMap.get(dateStr);
      if (existing) {
        existing.total++;
        existing[drift.severity]++;
        existing[drift.type]++;
      }
    });

    return Array.from(dateMap.values());
  };

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return `${date.getMonth() + 1}/${date.getDate()}`;
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '300px', color: 'white' }}>
        Loading chart...
      </div>
    );
  }

  const CustomTooltip = ({ active, payload, label }: any) => {
    if (active && payload && payload.length) {
      return (
        <div style={{
          background: 'var(--color-background-container-content)',
          border: '1px solid var(--color-border-container-divider)',
          borderRadius: '4px',
          padding: 'var(--space-s)',
        }}>
          <div style={{ fontSize: 'var(--font-size-body-s)', fontWeight: '700', color: 'white', marginBottom: 'var(--space-xs)' }}>
            {formatDate(label)}
          </div>
          {payload.map((entry: any, index: number) => (
            <div key={index} style={{ fontSize: 'var(--font-size-body-s)', color: entry.color }}>
              {entry.name}: {entry.value}
            </div>
          ))}
        </div>
      );
    }
    return null;
  };

  const Chart = chartType === 'area' ? AreaChart : LineChart;
  const ChartComponent = chartType === 'area' ? Area : Line;

  return (
    <div style={{
      background: 'var(--color-background-container-content)',
      border: '1px solid var(--color-border-container-divider)',
      borderRadius: '8px',
      padding: 'var(--space-l)',
    }}>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--space-l)' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-s)' }}>
          <TrendingUp size={20} style={{ color: 'var(--aws-smile-orange)' }} />
          <h3 style={{ fontSize: 'var(--font-size-heading-s)', fontWeight: '700', color: 'white', margin: 0 }}>
            Drift Trends (Last 30 Days)
          </h3>
        </div>

        <div style={{ display: 'flex', gap: 'var(--space-s)' }}>
          {/* View Toggle */}
          <select
            value={dataView}
            onChange={(e) => setDataView(e.target.value as 'severity' | 'type')}
            style={{
              background: 'var(--color-background-container-header)',
              border: '1px solid var(--color-border-container-divider)',
              borderRadius: '4px',
              padding: 'var(--space-xs) var(--space-s)',
              color: 'white',
              fontSize: 'var(--font-size-body-s)',
            }}
          >
            <option value="severity">By Severity</option>
            <option value="type">By Type</option>
          </select>

          {/* Chart Type Toggle */}
          <select
            value={chartType}
            onChange={(e) => setChartType(e.target.value as 'line' | 'area')}
            style={{
              background: 'var(--color-background-container-header)',
              border: '1px solid var(--color-border-container-divider)',
              borderRadius: '4px',
              padding: 'var(--space-xs) var(--space-s)',
              color: 'white',
              fontSize: 'var(--font-size-body-s)',
            }}
          >
            <option value="area">Area Chart</option>
            <option value="line">Line Chart</option>
          </select>
        </div>
      </div>

      {/* Chart */}
      <ResponsiveContainer width="100%" height={300}>
        <Chart data={data} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
          <defs>
            <linearGradient id="colorCritical" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#d13212" stopOpacity={0.8} />
              <stop offset="95%" stopColor="#d13212" stopOpacity={0} />
            </linearGradient>
            <linearGradient id="colorHigh" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#f89256" stopOpacity={0.8} />
              <stop offset="95%" stopColor="#f89256" stopOpacity={0} />
            </linearGradient>
            <linearGradient id="colorMedium" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#f9b959" stopOpacity={0.8} />
              <stop offset="95%" stopColor="#f9b959" stopOpacity={0} />
            </linearGradient>
            <linearGradient id="colorLow" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#64748b" stopOpacity={0.8} />
              <stop offset="95%" stopColor="#64748b" stopOpacity={0} />
            </linearGradient>
            <linearGradient id="colorCreated" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#1d8102" stopOpacity={0.8} />
              <stop offset="95%" stopColor="#1d8102" stopOpacity={0} />
            </linearGradient>
            <linearGradient id="colorModified" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#f9b959" stopOpacity={0.8} />
              <stop offset="95%" stopColor="#f9b959" stopOpacity={0} />
            </linearGradient>
            <linearGradient id="colorDeleted" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#d13212" stopOpacity={0.8} />
              <stop offset="95%" stopColor="#d13212" stopOpacity={0} />
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" />
          <XAxis
            dataKey="date"
            tickFormatter={formatDate}
            stroke="#64748b"
            style={{ fontSize: '11px' }}
          />
          <YAxis stroke="#64748b" style={{ fontSize: '11px' }} />
          <Tooltip content={<CustomTooltip />} />
          <Legend
            wrapperStyle={{ fontSize: '12px', color: 'white' }}
            iconType="line"
          />

          {dataView === 'severity' ? (
            <>
              <ChartComponent
                type="monotone"
                dataKey="critical"
                name="Critical"
                stroke="#d13212"
                fill={chartType === 'area' ? 'url(#colorCritical)' : undefined}
                strokeWidth={2}
              />
              <ChartComponent
                type="monotone"
                dataKey="high"
                name="High"
                stroke="#f89256"
                fill={chartType === 'area' ? 'url(#colorHigh)' : undefined}
                strokeWidth={2}
              />
              <ChartComponent
                type="monotone"
                dataKey="medium"
                name="Medium"
                stroke="#f9b959"
                fill={chartType === 'area' ? 'url(#colorMedium)' : undefined}
                strokeWidth={2}
              />
              <ChartComponent
                type="monotone"
                dataKey="low"
                name="Low"
                stroke="#64748b"
                fill={chartType === 'area' ? 'url(#colorLow)' : undefined}
                strokeWidth={2}
              />
            </>
          ) : (
            <>
              <ChartComponent
                type="monotone"
                dataKey="created"
                name="Created"
                stroke="#1d8102"
                fill={chartType === 'area' ? 'url(#colorCreated)' : undefined}
                strokeWidth={2}
              />
              <ChartComponent
                type="monotone"
                dataKey="modified"
                name="Modified"
                stroke="#f9b959"
                fill={chartType === 'area' ? 'url(#colorModified)' : undefined}
                strokeWidth={2}
              />
              <ChartComponent
                type="monotone"
                dataKey="deleted"
                name="Deleted"
                stroke="#d13212"
                fill={chartType === 'area' ? 'url(#colorDeleted)' : undefined}
                strokeWidth={2}
              />
            </>
          )}
        </Chart>
      </ResponsiveContainer>
    </div>
  );
};

export default TimeSeriesChart;
