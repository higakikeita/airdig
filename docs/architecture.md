# AirDig Architecture

## Overview

AirDig is a cloud observability platform built on a unified graph-based data model. It integrates four core pillars—**SkyGraph**, **DeepDrift**, **TraceCore**, and **PulseSight**—to provide comprehensive visibility from the cloud topology layer down to runtime execution.

---

## System Architecture

> 📊 **Interactive Diagrams:** See [detailed architecture diagrams](./diagrams/) for Mermaid-based visualizations.

### High-Level Architecture

For a complete interactive diagram, see [System Architecture Diagram](./diagrams/system-architecture.md).

```
┌─────────────────────────────────────────────────────────────┐
│                      AirDig UI (Next.js)                     │
│  ┌──────────┬──────────┬──────────┬──────────┬──────────┐  │
│  │ SkyGraph │DeepDrift │TraceCore │PulseSight│ Unified  │  │
│  │   View   │   View   │   View   │   View   │   View   │  │
│  └──────────┴──────────┴──────────┴──────────┴──────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                  AirDig Engine (GraphQL API)                 │
│                    Unified Graph Model                       │
└─────────────────────────────────────────────────────────────┘
        │              │              │              │
        ▼              ▼              ▼              ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│  SkyGraph   │ │  DeepDrift  │ │ TraceCore   │ │ PulseSight  │
│             │ │             │ │             │ │             │
│ Cloud Topo  │ │ Drift Detect│ │  APM/Trace  │ │Metrics/Logs │
│   Engine    │ │   Engine    │ │   Engine    │ │   Engine    │
└─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘
        │              │              │              │
        ▼              ▼              ▼              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Data Platform Layer                       │
│  ┌──────────┬──────────┬──────────┬──────────┬──────────┐  │
│  │  TiDB/   │  Tempo   │  Mimir   │   Loki   │  Redis   │  │
│  │ClickHouse│  (S3)    │ (Metrics)│  (Logs)  │ (Cache)  │  │
│  └──────────┴──────────┴──────────┴──────────┴──────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Data Collection Layer                      │
│  ┌──────────┬──────────┬──────────┬──────────┬──────────┐  │
│  │Cloud APIs│Terraform │   OTel   │Prometheus│  Falco   │  │
│  │  (AWS/   │  State   │Collector │ Exporters│  eBPF    │  │
│  │ GCP/k8s) │  Parser  │          │          │          │  │
│  └──────────┴──────────┴──────────┴──────────┴──────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Architecture Diagrams

- 📊 [System Architecture](./diagrams/system-architecture.md) - Complete system overview with all components
- 🔄 [Data Flow](./diagrams/data-flow.md) - Event-driven architecture and data pipelines
- 🟦 [SkyGraph Architecture](./diagrams/skygraph-architecture.md) - Cloud scanning and graph building internals
- 🚀 [Deployment Architecture](./diagrams/deployment-architecture.md) - Kubernetes and Docker Compose deployments

---

## The Four Pillars

### 🟦 1. SkyGraph — Cloud Topology & Dependency Graph

**Purpose:** Build a complete, real-time graph of cloud infrastructure resources and their dependencies.

**Components:**
- **Cloud Scanners:** AWS, GCP, Azure, Kubernetes API clients
- **IaC Parsers:** Terraform state/plan parser, CloudFormation, CDK
- **Graph Builder:** Constructs nodes (resources) and edges (dependencies)
- **Storage:** TiDB or ClickHouse for graph data

**Data Model:**
```go
type ResourceNode {
    ID         string
    Type       string  // ec2, vpc, rds, lambda, k8s_pod, etc.
    Provider   string  // aws, gcp, azure, kubernetes
    Region     string
    Metadata   map[string]any
    Health     HealthStatus  // from PulseSight
    DriftState DriftStatus   // from DeepDrift
}

type Edge {
    From   string
    To     string
    Type   string  // network, dependency, call, drift, change
    Weight float64
}
```

**Key Features:**
- Real-time resource discovery
- Multi-cloud support (AWS, GCP, Azure, K8s)
- Dependency mapping (VPC → Subnet → EC2 → RDS)
- Integration with IaC tools

---

### 🟢 2. DeepDrift — Drift Detection & Change Intelligence

**Purpose:** Detect and analyze drift between IaC desired state and actual cloud state, with root cause analysis.

**Components:**
- **TFDrift Core:** Compares Terraform state vs. actual cloud state
- **Diff Engine:** Generates DAG-based diff graphs
- **Event Correlator:** Links drift to CloudTrail/audit logs
- **Impact Analyzer:** Predicts impact on other resources (using graph)

**Data Model:**
```go
type DriftEvent {
    ResourceID   string
    Type         string  // created, modified, deleted
    Timestamp    time.Time
    Before       map[string]any
    After        map[string]any
    RootCause    string  // from CloudTrail
    ImpactedEdges []Edge
}
```

**Key Features:**
- Real-time drift detection
- CloudTrail/Stratoshark correlation
- Impact analysis (who changed what, and what broke)
- Graph-based change propagation tracking

---

### 🟣 3. TraceCore — Distributed Tracing & APM

**Purpose:** Capture application execution traces and correlate them with infrastructure changes.

**Components:**
- **OTel Collector:** Receives traces (OTLP protocol)
- **Service Map Generator:** Builds call graphs from spans
- **Trace Storage:** Tempo (S3-backed) or ClickHouse
- **Graph Integrator:** Maps services to infrastructure resources

**Data Model:**
```go
type TraceEdge {
    From        string  // service name
    To          string  // service name
    Latency     time.Duration
    ErrorRate   float64
    TraceIDs    []string
    Timestamp   time.Time
}
```

**Key Features:**
- OpenTelemetry-native
- Service dependency mapping
- Latency and error rate tracking
- **Unique:** Correlate traces with drift events (e.g., "RDS change → latency spike")

---

### 🟡 4. PulseSight — Metrics, Logs & Runtime Observability

**Purpose:** Collect system health metrics, logs, and runtime security events.

**Components:**
- **Prometheus Scraper:** Node exporter, K8s metrics, custom metrics
- **Log Collector:** Loki-compatible, OTel logs
- **Runtime Security:** Falco rules, eBPF events
- **Health Evaluator:** Assigns health status to resources

**Data Model:**
```go
type HealthStatus {
    Status      string  // healthy, warning, critical
    LastCheck   time.Time
    Metrics     map[string]float64
    Alerts      []Alert
}

type RuntimeEvent {
    Type        string  // falco_alert, network_flow, process_spawn
    Severity    string
    ResourceID  string
    Timestamp   time.Time
    Metadata    map[string]any
}
```

**Key Features:**
- Real-time metrics ingestion (Prometheus)
- Log aggregation (Loki)
- Runtime security (Falco/eBPF)
- **Unique:** Maps metrics/events to graph nodes (e.g., "EC2 CPU spike → red node in UI")

---

## AirDig Engine — Unified Graph Layer

The **AirDig Engine** is the integration layer that:
1. Merges data from all four pillars into a unified graph
2. Provides a GraphQL API for querying
3. Handles real-time updates (event-driven)
4. Stores the graph in TiDB/ClickHouse

**Core Responsibilities:**
- **Graph Unification:** Combine SkyGraph nodes + DeepDrift edges + TraceCore edges + PulseSight attributes
- **Real-time Updates:** React to CloudTrail events, drift events, trace ingestion, metrics
- **Query Interface:** GraphQL API for UI and external integrations
- **Event Stream:** Kafka/NATS for pub-sub between pillars

---

## Data Flow

### Example: Drift Detection → Impact Analysis → UI Update

1. **DeepDrift** detects a security group change (AWS API)
2. **CloudTrail event** arrives → correlated with drift
3. **SkyGraph** updates the graph edge (EC2 ↔ SG)
4. **PulseSight** detects connection failures (Prometheus alert)
5. **TraceCore** shows increased error rate in traces
6. **AirDig Engine** merges all data → marks node as "unhealthy"
7. **UI** displays red node with drift annotation + trace link

---

## Technology Stack

| Layer | Technology |
|-------|------------|
| **UI** | Next.js, Cytoscape.js, D3.js, TailwindCSS |
| **API** | Go, GraphQL (gqlgen), gRPC |
| **Graph Database** | TiDB (OLTP+OLAP) or ClickHouse |
| **Trace Storage** | Grafana Tempo (S3-backed) |
| **Metrics** | Prometheus, Grafana Mimir |
| **Logs** | Grafana Loki |
| **Cache** | Redis |
| **Event Bus** | Kafka or NATS |
| **Collector** | OpenTelemetry Collector (custom processors) |
| **Runtime** | eBPF (Cilium, Pixie), Falco |
| **Cloud APIs** | AWS SDK, GCP SDK, Azure SDK, K8s client-go |

---

## Deployment Architecture

AirDig is designed to run on Kubernetes but can also run via Docker Compose for development.

### Kubernetes Deployment

```
┌─────────────────────────────────────────────────────┐
│                   AirDig UI (Pod)                    │
└─────────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────┐
│              AirDig Engine (Deployment)              │
│             (GraphQL API + Graph Processor)          │
└─────────────────────────────────────────────────────┘
        │              │              │              │
        ▼              ▼              ▼              ▼
┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────┐
│ SkyGraph   │ │ DeepDrift  │ │ TraceCore  │ │PulseSight  │
│  Service   │ │  Service   │ │  Service   │ │  Service   │
└────────────┘ └────────────┘ └────────────┘ └────────────┘
        │              │              │              │
        ▼              ▼              ▼              ▼
┌─────────────────────────────────────────────────────┐
│              Data Platform (StatefulSets)            │
│  TiDB | Tempo (S3) | Mimir | Loki | Redis | Kafka  │
└─────────────────────────────────────────────────────┘
```

---

## Security Considerations

- **IAM/RBAC:** Read-only cloud API access for scanners
- **Secrets Management:** Vault or K8s secrets for credentials
- **Data Encryption:** At-rest (TiDB/S3) and in-transit (TLS)
- **Audit Logging:** All API calls logged via OpenTelemetry

---

## Scalability

- **Horizontal Scaling:** All services (SkyGraph, DeepDrift, TraceCore, PulseSight) are stateless
- **Graph Storage:** TiDB scales to petabytes with TiKV distributed storage
- **Trace Storage:** Tempo uses S3 (unlimited scale)
- **Event Bus:** Kafka/NATS for high-throughput event streaming

---

## Comparison with Existing Tools

| Feature | Datadog | Wiz | Sysdig | CloudGraph | AirDig |
|---------|---------|-----|--------|------------|--------|
| Cloud Config Graph | ❌ | ✅ | ❌ | ✅ | ✅ |
| Drift Detection | ❌ | ❌ | ❌ | ❌ | ✅ |
| APM / Distributed Tracing | ✅ | ❌ | ✅ | ❌ | ✅ |
| Runtime Security | ⚠️ | ✅ | ✅ | ❌ | ✅ |
| Unified Graph View | ❌ | ⚠️ | ❌ | ⚠️ | ✅ |
| IaC Integration | ❌ | ⚠️ | ❌ | ❌ | ✅ |
| Open Source | ❌ | ❌ | ⚠️ | ✅ | ✅ |

**AirDig is the only platform that unifies all layers in a single graph.**

---

## Future Enhancements

- **AI-Driven Root Cause Analysis:** Use ML to predict impact of changes
- **Multi-Tenancy:** Isolate graphs per team/environment
- **Custom Plugins:** Allow users to write custom scanners/analyzers
- **GitOps Integration:** Sync with Git for IaC repos
- **Cost Analysis:** Map resource costs to the graph

---

## Conclusion

AirDig extends the philosophy of Sysdig and Falco from the OS/container layer to the **entire cloud stack**. By unifying topology, drift, tracing, and runtime events in a single graph, AirDig provides unprecedented visibility and intelligence for modern cloud operations.

---

**Next Steps:**
- [Vision Document](./vision.md)
- [Development Roadmap](./roadmap.md)
- [SkyGraph Design](../skygraph/README.md)
