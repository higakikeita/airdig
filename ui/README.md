# AirDig UI — Unified Graph Visualization

**AirDig UI** is the web-based user interface that visualizes the unified graph from all four pillars in a beautiful, interactive experience.

---

## Overview

The AirDig UI provides:
- **Interactive graph visualization** (Cytoscape.js)
- **Tab-based views** for each pillar
- **Unified view** combining all data
- **Real-time updates** via WebSocket
- **Node/edge inspector** for detailed information

---

## Status

**Coming in v0.4.0**

The UI will be developed after the AirDig Engine is complete.

---

## Planned Features

- Interactive graph with zoom, pan, and search
- Tab-based navigation:
  - SkyGraph (cloud topology)
  - DeepDrift (change timeline)
  - TraceCore (service map)
  - PulseSight (health dashboard)
  - Unified View (all layers combined)
- Node coloring based on health status
- Edge highlighting for drift/traces
- Timeline view for historical changes
- Dark mode
- Keyboard shortcuts

---

## Tech Stack

- **Framework:** Next.js 14 (App Router)
- **Graph Visualization:** Cytoscape.js + D3.js
- **UI Library:** Tailwind CSS + shadcn/ui
- **API Client:** GraphQL (Apollo Client)
- **Real-time:** WebSocket subscriptions
- **State Management:** Zustand
- **Charts:** Recharts

---

## Architecture (Planned)

```
┌─────────────────────────────────────────────────┐
│              AirDig UI (Next.js)                 │
│                                                  │
│  ┌──────────────────────────────────────────┐  │
│  │       Tab Navigation                     │  │
│  │  [SkyGraph] [DeepDrift] [TraceCore]      │  │
│  │  [PulseSight] [Unified View]             │  │
│  └──────────────────────────────────────────┘  │
│                                                  │
│  ┌─────────────────┬──────────────────────────┐│
│  │  Graph Canvas   │   Inspector Panel        ││
│  │  (Cytoscape.js) │   - Node details         ││
│  │                 │   - Drift events         ││
│  │                 │   - Traces               ││
│  │                 │   - Metrics              ││
│  └─────────────────┴──────────────────────────┘│
│                                                  │
│  ┌──────────────────────────────────────────┐  │
│  │       Timeline / Search Bar              │  │
│  └──────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
                       │
                       ▼ (GraphQL + WebSocket)
         ┌──────────────────────────┐
         │   AirDig Engine API      │
         └──────────────────────────┘
```

---

## Views

### 1. SkyGraph View

**Cloud topology visualization**

- Shows all cloud resources as nodes
- Displays dependencies as edges
- Color-coded by resource type (EC2 = blue, RDS = green, etc.)
- Pan/zoom/search functionality

### 2. DeepDrift View

**Change timeline and drift detection**

- Timeline of infrastructure changes
- Drift events highlighted
- Before/after comparison
- CloudTrail correlation

### 3. TraceCore View

**Service map and APM**

- Service dependency graph
- Latency and error rate on edges
- Click to see sample traces
- Integration with Tempo UI

### 4. PulseSight View

**Health dashboard**

- Metrics dashboard (Grafana-style)
- Alert list
- Falco events timeline
- Log viewer

### 5. Unified View

**All pillars combined**

- Single graph with all data
- Node color = health status (green/yellow/red)
- Edge types:
  - Network (solid line)
  - Drift (dashed line)
  - Call (dotted line)
- Filters to toggle layers

---

## Example: Node Inspector

Click on a node (e.g., EC2 instance) to see:

```
┌─────────────────────────────────────────┐
│  EC2 Instance: i-123456                 │
├─────────────────────────────────────────┤
│  Type: t3.large                         │
│  Region: us-east-1                      │
│  Status: ● Running                      │
│  Health: 🟡 Warning                     │
├─────────────────────────────────────────┤
│  Drift Events (2):                      │
│  • Security group modified (5m ago)     │
│  • Instance type changed (2h ago)       │
├─────────────────────────────────────────┤
│  Services:                              │
│  • api (TraceCore)                      │
│  • worker (TraceCore)                   │
├─────────────────────────────────────────┤
│  Metrics:                               │
│  • CPU: 75% (Warning)                   │
│  • Memory: 60%                          │
│  • Network: 120 Mbps                    │
├─────────────────────────────────────────┤
│  Recent Logs:                           │
│  [ERROR] Connection timeout to DB       │
│  [WARN] High CPU usage detected         │
└─────────────────────────────────────────┘
```

---

## Development

### Setup (Coming in v0.4.0)

```bash
cd ui
npm install
npm run dev
```

### Environment Variables

```env
NEXT_PUBLIC_AIRDIG_API_URL=http://localhost:8080/graphql
NEXT_PUBLIC_WS_URL=ws://localhost:8080/subscriptions
```

---

## Design Mockups

**Coming soon**

---

**AirDig UI** is part of the [AirDig](https://github.com/yourusername/airdig) project.
