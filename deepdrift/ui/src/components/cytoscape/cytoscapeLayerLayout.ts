import type { LayoutOptions } from 'cytoscape';
import { getResourceLayer } from '../graph/graphUtils';

// Custom preset layout for layer-based architecture view
export const getLayerLayout = (nodes: any[]): LayoutOptions => {
  // Define layer positions (vertical stacking)
  const layerY: Record<string, number> = {
    internet: 100,
    edge: 250,
    network: 400,
    compute: 550,
    data: 700,
  };

  // Group nodes by layer
  const nodesByLayer: Record<string, any[]> = {
    internet: [],
    edge: [],
    network: [],
    compute: [],
    data: [],
    unknown: [],
  };

  nodes.forEach(node => {
    const layer = getResourceLayer(node.type);
    if (layer) {
      nodesByLayer[layer.id].push(node);
    } else {
      nodesByLayer.unknown.push(node);
    }
  });

  // Calculate positions for each node
  const positions: Record<string, { x: number; y: number }> = {};
  const containerWidth = 1200;
  const startX = 100;

  Object.entries(nodesByLayer).forEach(([layerId, layerNodes]) => {
    if (layerNodes.length === 0) return;

    const y = layerY[layerId] || 850;
    const spacing = Math.min(150, (containerWidth - 200) / Math.max(layerNodes.length - 1, 1));

    layerNodes.forEach((node, index) => {
      const x = startX + index * spacing + (containerWidth - layerNodes.length * spacing) / 2;
      positions[node.id] = { x, y };
    });
  });

  return {
    name: 'preset',
    positions: (node: any) => {
      const pos = positions[node.id()];
      return pos || { x: 600, y: 850 };
    },
    fit: true,
    padding: 80,
    animate: true,
    animationDuration: 500,
  };
};

// Layer background configuration for rendering
export interface LayerBackground {
  id: string;
  label: string;
  y: number;
  height: number;
  color: string;
}

export const LAYER_BACKGROUNDS: LayerBackground[] = [
  { id: 'internet', label: 'インターネット層', y: 50, height: 100, color: '#10b981' },
  { id: 'edge', label: 'エッジ層', y: 200, height: 100, color: '#8b5cf6' },
  { id: 'network', label: 'ネットワーク層', y: 350, height: 100, color: '#06b6d4' },
  { id: 'compute', label: 'コンピュート層', y: 500, height: 100, color: '#f59e0b' },
  { id: 'data', label: 'データ層', y: 650, height: 100, color: '#6366f1' },
];
