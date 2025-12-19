import React, { useEffect, useState } from 'react';
import apiClient from '../api/client';
import type { GraphNode, ResourceGraphData } from './graph/graphUtils';
import { filterImportantResources } from './graph/graphUtils';
import CytoscapeGraph from './cytoscape/CytoscapeGraph';
import type { LayoutMode } from './cytoscape/CytoscapeGraph';

type ViewMode = 'force' | 'hierarchical' | 'grid' | 'circle' | 'layers';
type DataSource = 'actual' | 'intended';

const viewModeToLayout: Record<ViewMode, LayoutMode> = {
  force: 'cose',
  hierarchical: 'concentric',
  grid: 'grid',
  circle: 'circle',
  layers: 'layers',
};

export const ResourceGraph: React.FC = () => {
  const [loading, setLoading] = useState(true);
  const [stats, setStats] = useState({ total: 0, filtered: 0 });
  const [viewMode, setViewMode] = useState<ViewMode>('force');
  const [dataSource, setDataSource] = useState<DataSource>('actual');
  const [rawData, setRawData] = useState<ResourceGraphData>({ nodes: [], edges: [] });
  const [intendedData, setIntendedData] = useState<ResourceGraphData>({ nodes: [], edges: [] });
  const [displayNodes, setDisplayNodes] = useState<GraphNode[]>([]);

  useEffect(() => {
    fetchGraphData();
  }, []);

  useEffect(() => {
    const currentData = dataSource === 'actual' ? rawData : intendedData;
    const filteredNodes = filterImportantResources(currentData.nodes);
    setStats({ total: currentData.nodes.length, filtered: filteredNodes.length });
    setDisplayNodes(filteredNodes);
  }, [rawData, intendedData, dataSource]);

  const fetchGraphData = async () => {
    try {
      setLoading(true);
      const [actualGraph, intendedGraph] = await Promise.all([
        apiClient.getGraph(),
        apiClient.getIntendedGraph(),
      ]);
      setRawData(actualGraph);
      setIntendedData(intendedGraph);
    } catch (error) {
      console.error('Failed to fetch graph:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleNodeClick = (node: GraphNode) => {
    console.log('Node clicked:', node);
  };

  if (loading) {
    return (
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          height: '100%',
          fontSize: '16px',
          color: '#94a3b8',
          background: '#0f1b2a',
        }}
      >
        読み込み中...
      </div>
    );
  }

  return (
    <div style={{ width: '100%', height: '100%', position: 'relative', background: '#0f1b2a' }}>
      {/* Compact Header */}
      <div
        style={{
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          height: '50px',
          background: 'linear-gradient(135deg, #1e293b 0%, #0f172a 100%)',
          borderBottom: '2px solid #ff9900',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '0 20px',
          zIndex: 10,
          backdropFilter: 'blur(10px)',
        }}
      >
        {/* Left: Title & Stats */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '20px' }}>
          <div style={{ fontSize: '16px', fontWeight: 700, color: 'white', letterSpacing: '0.5px' }}>
            ⚡ AirDig <span style={{ color: '#ff9900' }}>DeepDrift</span>
          </div>
          <div style={{ display: 'flex', gap: '16px', fontSize: '12px', color: '#94a3b8' }}>
            <div>
              全体: <span style={{ color: 'white', fontWeight: 600 }}>{stats.total}</span>
            </div>
            <div>
              表示: <span style={{ color: '#10b981', fontWeight: 600 }}>{stats.filtered}</span>
            </div>
          </div>
        </div>

        {/* Center: Data Source Toggle */}
        <div
          style={{
            display: 'flex',
            gap: '4px',
            backgroundColor: '#1e293b',
            borderRadius: '6px',
            padding: '3px',
          }}
        >
          {[
            { source: 'actual' as DataSource, label: 'AWS実環境', icon: '☁️' },
            { source: 'intended' as DataSource, label: 'Terraform', icon: '📝' },
          ].map((tab) => (
            <button
              key={tab.source}
              onClick={() => setDataSource(tab.source)}
              style={{
                padding: '6px 14px',
                fontSize: '12px',
                fontWeight: 500,
                color: dataSource === tab.source ? 'white' : '#94a3b8',
                backgroundColor: dataSource === tab.source ? '#0ea5e9' : 'transparent',
                border: 'none',
                borderRadius: '4px',
                cursor: 'pointer',
                transition: 'all 0.2s',
              }}
            >
              {tab.icon} {tab.label}
            </button>
          ))}
        </div>

        {/* Right: View Mode Tabs */}
        <div
          style={{
            display: 'flex',
            gap: '4px',
            backgroundColor: '#1e293b',
            borderRadius: '6px',
            padding: '3px',
          }}
        >
          {[
            { mode: 'force' as ViewMode, label: '力学', icon: '🌀' },
            { mode: 'hierarchical' as ViewMode, label: '階層', icon: '🏗️' },
            { mode: 'layers' as ViewMode, label: 'レイヤー', icon: '📊' },
            { mode: 'grid' as ViewMode, label: 'グリッド', icon: '⊞' },
            { mode: 'circle' as ViewMode, label: '円形', icon: '○' },
          ].map((tab) => (
            <button
              key={tab.mode}
              onClick={() => setViewMode(tab.mode)}
              style={{
                padding: '6px 12px',
                fontSize: '12px',
                fontWeight: 500,
                color: viewMode === tab.mode ? 'white' : '#94a3b8',
                backgroundColor: viewMode === tab.mode ? '#f59e0b' : 'transparent',
                border: 'none',
                borderRadius: '4px',
                cursor: 'pointer',
                transition: 'all 0.2s',
              }}
            >
              {tab.icon} {tab.label}
            </button>
          ))}
        </div>
      </div>

      {/* Graph Canvas */}
      <div style={{ width: '100%', height: '100%', paddingTop: '50px' }}>
        <CytoscapeGraph
          nodes={displayNodes}
          edges={[]}
          layoutMode={viewModeToLayout[viewMode]}
          onNodeClick={handleNodeClick}
        />
      </div>
    </div>
  );
};

export default ResourceGraph;
