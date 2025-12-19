import React, { useEffect, useRef, useState } from 'react';
import cytoscape from 'cytoscape';
import type { Core, LayoutOptions } from 'cytoscape';
import type { GraphNode, GraphEdge } from '../graph/graphUtils';
import { getCytoscapeStylesheet } from './cytoscapeStyles';
import { convertGraphToCytoscape, addSyntheticEdges } from './cytoscapeAdapters';
import { getLayerLayout } from './cytoscapeLayerLayout';

export type LayoutMode = 'cose' | 'dagre' | 'grid' | 'circle' | 'concentric' | 'layers';

interface CytoscapeGraphProps {
  nodes: GraphNode[];
  edges: GraphEdge[];
  layoutMode?: LayoutMode;
  onNodeClick?: (node: GraphNode) => void;
}

export const CytoscapeGraph: React.FC<CytoscapeGraphProps> = ({
  nodes,
  edges,
  layoutMode = 'cose',
  onNodeClick,
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const cyRef = useRef<Core | null>(null);
  const [selectedNode, setSelectedNode] = useState<any>(null);

  // Initialize Cytoscape
  useEffect(() => {
    if (!containerRef.current) return;

    // Convert data
    const { elements } = convertGraphToCytoscape(nodes, edges);

    // Add synthetic edges based on relationships
    addSyntheticEdges(nodes, elements);

    // Add internet node if there are NAT gateways
    const hasNatGateway = nodes.some(n => n.type === 'nat_gateway');
    if (hasNatGateway) {
      elements.unshift({
        data: {
          id: 'internet-global',
          label: 'インターネット',
          type: 'internet',
        },
      });
    }

    // Initialize Cytoscape
    const cy = cytoscape({
      container: containerRef.current,
      elements,
      style: getCytoscapeStylesheet(),
      layout: getLayoutOptions(layoutMode, nodes),
      minZoom: 0.2,
      maxZoom: 2.5,
      wheelSensitivity: 0.2,
      // Ensure proper rendering and panning
      pixelRatio: 'auto',
      motionBlur: false,
      autoungrabify: false,
      autounselectify: false,
    });

    cyRef.current = cy;

    // Center the graph after layout completes
    cy.on('layoutstop', () => {
      // Wait a bit for rendering to complete
      setTimeout(() => {
        const allNodes = cy.nodes();

        if (allNodes.length > 0) {
          // Fit all nodes with generous padding
          cy.fit(allNodes, 100);

          // Force center the viewport
          cy.center(allNodes);

          // Ensure reasonable zoom level
          const zoom = cy.zoom();
          if (zoom < 0.3) {
            cy.zoom(0.3);
            cy.center(allNodes);
          }
          if (zoom > 2) {
            cy.zoom(2);
            cy.center(allNodes);
          }
        }
      }, 100);
    });

    // Add window resize handler
    const handleResize = () => {
      if (cyRef.current) {
        cyRef.current.resize();
        cyRef.current.fit(cyRef.current.nodes(), 100);
        cyRef.current.center();
      }
    };

    window.addEventListener('resize', handleResize);

    // Trigger initial resize to ensure proper sizing
    setTimeout(() => {
      if (cyRef.current) {
        cyRef.current.resize();
        cyRef.current.fit(cyRef.current.nodes(), 100);
        cyRef.current.center();
      }
    }, 200);

    // Event handlers
    cy.on('tap', 'node', (event) => {
      const node = event.target;
      const data = node.data();
      setSelectedNode(data);

      if (onNodeClick && data._original) {
        onNodeClick(data._original);
      }
    });

    cy.on('tap', (event) => {
      if (event.target === cy) {
        setSelectedNode(null);
      }
    });

    // Cleanup
    return () => {
      window.removeEventListener('resize', handleResize);
      cy.destroy();
      cyRef.current = null;
    };
  }, [nodes, edges, layoutMode, onNodeClick]);

  // Update layout when layout mode changes
  useEffect(() => {
    if (cyRef.current) {
      // Extract original nodes for layer layout
      const originalNodes: GraphNode[] = cyRef.current.nodes().map((ele: any) => ele.data('_original')).filter(Boolean);

      const layout = cyRef.current.layout(getLayoutOptions(layoutMode, originalNodes));
      layout.run();

      // Center after layout completes
      layout.on('layoutstop', () => {
        setTimeout(() => {
          if (cyRef.current) {
            const allNodes = cyRef.current.nodes();

            if (allNodes.length > 0) {
              cyRef.current.fit(allNodes, 100);
              cyRef.current.center(allNodes);

              const zoom = cyRef.current.zoom();
              if (zoom < 0.3) {
                cyRef.current.zoom(0.3);
                cyRef.current.center(allNodes);
              }
              if (zoom > 2) {
                cyRef.current.zoom(2);
                cyRef.current.center(allNodes);
              }
            }
          }
        }, 100);
      });
    }
  }, [layoutMode]);

  return (
    <div style={{ position: 'relative', width: '100%', height: '100%' }}>
      {/* Cytoscape Container */}
      <div
        ref={containerRef}
        style={{
          width: '100%',
          height: '100%',
          background: '#0f1b2a',
          position: 'relative',
          overflow: 'hidden',
        }}
      />

      {/* Node Details Panel */}
      {selectedNode && (
        <div
          style={{
            position: 'absolute',
            top: 16,
            right: 16,
            width: 320,
            maxHeight: 'calc(100% - 32px)',
            background: '#16191f',
            border: '1px solid #414d5c',
            borderRadius: '8px',
            padding: '16px',
            overflow: 'auto',
            boxShadow: '0 4px 6px rgba(0, 0, 0, 0.3)',
            zIndex: 1000,
          }}
        >
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'start',
              marginBottom: '12px',
            }}
          >
            <div>
              <div style={{ fontSize: '16px', fontWeight: 600, color: 'white' }}>
                {selectedNode.label}
              </div>
              <div style={{ fontSize: '12px', color: '#aab7b8', marginTop: '4px' }}>
                {selectedNode.type}
              </div>
            </div>
            <button
              onClick={() => setSelectedNode(null)}
              style={{
                background: 'transparent',
                border: 'none',
                color: '#aab7b8',
                cursor: 'pointer',
                fontSize: '20px',
                padding: '0 4px',
              }}
            >
              ✕
            </button>
          </div>

          {/* Health Status */}
          {selectedNode.health && (
            <div style={{ marginBottom: '12px' }}>
              <div style={{ fontSize: '11px', color: '#aab7b8', marginBottom: '4px' }}>
                健康状態
              </div>
              <div
                style={{
                  display: 'inline-block',
                  padding: '4px 8px',
                  borderRadius: '4px',
                  fontSize: '11px',
                  fontWeight: 600,
                  background:
                    selectedNode.health === 'healthy'
                      ? '#1d810220'
                      : selectedNode.health === 'warning'
                      ? '#f59e0b20'
                      : '#d1321220',
                  color:
                    selectedNode.health === 'healthy'
                      ? '#1d8102'
                      : selectedNode.health === 'warning'
                      ? '#f59e0b'
                      : '#d13212',
                }}
              >
                {selectedNode.health === 'healthy'
                  ? '✓ 正常'
                  : selectedNode.health === 'warning'
                  ? '⚠ 警告'
                  : '✗ 異常'}
              </div>
            </div>
          )}

          {/* Public Access Badge */}
          {selectedNode.public && (
            <div style={{ marginBottom: '12px' }}>
              <div
                style={{
                  display: 'inline-block',
                  padding: '4px 8px',
                  borderRadius: '4px',
                  fontSize: '11px',
                  fontWeight: 600,
                  background: '#ef444420',
                  color: '#ef4444',
                }}
              >
                🌐 パブリックアクセス
              </div>
            </div>
          )}

          {/* Details */}
          {selectedNode.details && Object.keys(selectedNode.details).length > 0 && (
            <div>
              <div style={{ fontSize: '11px', color: '#aab7b8', marginBottom: '8px' }}>
                詳細情報
              </div>
              <div style={{ fontSize: '12px', color: 'white' }}>
                {Object.entries(selectedNode.details).map(([key, value]) => (
                  <div
                    key={key}
                    style={{
                      marginBottom: '6px',
                      paddingBottom: '6px',
                      borderBottom: '1px solid #414d5c20',
                    }}
                  >
                    <div style={{ color: '#aab7b8', fontSize: '10px', marginBottom: '2px' }}>
                      {key}
                    </div>
                    <div style={{ color: 'white', wordBreak: 'break-all' }}>
                      {typeof value === 'boolean'
                        ? value
                          ? 'はい'
                          : 'いいえ'
                        : value?.toString() || 'N/A'}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Region/Provider */}
          {selectedNode.region && (
            <div style={{ marginTop: '12px', fontSize: '10px', color: '#64748b' }}>
              {selectedNode.provider}:{selectedNode.region}
            </div>
          )}
        </div>
      )}
    </div>
  );
};

// Get layout options based on mode
const getLayoutOptions = (mode: LayoutMode, nodes: GraphNode[] = []): LayoutOptions => {
  switch (mode) {
    case 'cose':
      return {
        name: 'cose',
        nodeRepulsion: 8000,
        idealEdgeLength: 100,
        nodeOverlap: 20,
        refresh: 20,
        fit: true,
        padding: 30,
        randomize: false,
        componentSpacing: 100,
        animate: true,
        animationDuration: 500,
      };

    case 'dagre':
      return {
        name: 'preset', // We'll use preset for now, dagre needs separate plugin
        fit: true,
        padding: 30,
      };

    case 'grid':
      return {
        name: 'grid',
        fit: true,
        padding: 30,
        avoidOverlap: true,
        condense: false,
        rows: undefined,
        cols: undefined,
      };

    case 'circle':
      return {
        name: 'circle',
        fit: true,
        padding: 30,
        avoidOverlap: true,
        radius: undefined,
      };

    case 'concentric':
      return {
        name: 'concentric',
        fit: true,
        padding: 30,
        avoidOverlap: true,
        concentric: (node: any) => node.degree(),
        levelWidth: () => 2,
      };

    case 'layers':
      return getLayerLayout(nodes);

    default:
      return {
        name: 'cose',
        fit: true,
        padding: 30,
      };
  }
};

export default CytoscapeGraph;
