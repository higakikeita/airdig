# AirDig Engine — Unified Graph Integration Layer

**AirDig Engine** is the integration layer that unifies all four pillars (SkyGraph, DeepDrift, TraceCore, PulseSight) into a single, queryable graph.

---

## Overview

The AirDig Engine:
- **Merges data** from all four pillars into a unified graph model
- **Provides a GraphQL API** for querying the graph
- **Handles real-time updates** via event-driven architecture
- **Stores the graph** in TiDB or ClickHouse

---

## Status

**Coming in v0.3.0**

The Engine will be developed after the four pillars reach MVP status.

---

## Planned Features

- Unified graph data model
- GraphQL API
- Event stream processing (Kafka/NATS)
- Real-time graph updates
- Data correlation engine
- WebSocket for live UI updates

---

## Architecture (Planned)

```
┌─────────────────────────────────────────────────┐
│              AirDig Engine                       │
│                                                  │
│  ┌──────────────────────────────────────────┐  │
│  │       GraphQL API Server                 │  │
│  └──────────────────────────────────────────┘  │
│                                                  │
│  ┌──────────────────────────────────────────┐  │
│  │       Event Stream Processor             │  │
│  │  (Kafka/NATS Consumer)                   │  │
│  └──────────────────────────────────────────┘  │
│                                                  │
│  ┌──────────────────────────────────────────┐  │
│  │       Graph Merge Engine                 │  │
│  │  (Combine SkyGraph + Drift + Trace +     │  │
│  │   PulseSight)                            │  │
│  └──────────────────────────────────────────┘  │
│                                                  │
│  ┌──────────────────────────────────────────┐  │
│  │       Data Store (TiDB/ClickHouse)       │  │
│  └──────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
```

---

## Data Model

The unified graph will combine:

```go
type UnifiedNode struct {
    // From SkyGraph
    ResourceNode

    // From PulseSight
    Health HealthStatus

    // From DeepDrift
    DriftStatus DriftStatus
    LastDrift   *DriftEvent

    // From TraceCore
    Services []string  // Services running on this resource
}

type UnifiedEdge struct {
    From   string
    To     string
    Type   EdgeType  // network, dependency, call, drift, change
    Weight float64

    // From TraceCore (if Type = call)
    Latency   *time.Duration
    ErrorRate *float64

    // From DeepDrift (if Type = drift)
    DriftEvent *DriftEvent
}
```

---

## API (Planned)

### GraphQL Schema

```graphql
type Query {
  # Get resource by ID
  resource(id: ID!): Resource

  # Get all resources by type
  resources(type: String, provider: String): [Resource!]!

  # Get drift events
  driftEvents(resourceId: ID, start: Time, end: Time): [DriftEvent!]!

  # Get service map
  serviceMap: ServiceMap!

  # Search resources
  search(query: String!): [Resource!]!
}

type Resource {
  id: ID!
  type: String!
  provider: String!
  name: String!
  metadata: JSON!
  health: HealthStatus!
  drift: DriftStatus!
  services: [Service!]!
  dependencies: [Resource!]!
}

type Subscription {
  # Real-time updates
  resourceUpdated(id: ID): Resource!
  driftDetected: DriftEvent!
  alertFired: Alert!
}
```

---

## Development

**Coming soon in v0.3.0**

---

**AirDig Engine** is part of the [AirDig](https://github.com/yourusername/airdig) project.
