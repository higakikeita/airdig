# SkyGraph — Cloud Topology & Dependency Graph

**SkyGraph** is the first pillar of AirDig. It builds a complete, real-time graph of cloud infrastructure resources and their dependencies across AWS, GCP, Azure, and Kubernetes.

---

## Overview

SkyGraph scans cloud APIs and constructs a unified graph model where:
- **Nodes** represent cloud resources (EC2, VPC, RDS, K8s pods, etc.)
- **Edges** represent relationships (network connections, dependencies, ownership)

This graph becomes the **foundation** for all other AirDig pillars:
- **DeepDrift** adds drift edges
- **TraceCore** adds call edges
- **PulseSight** adds health attributes

---

## Features

- ✅ **Multi-cloud scanning:** AWS, GCP, Azure, Kubernetes
- ✅ **Resource discovery:** Automatic detection of all resources in scope
- ✅ **Dependency mapping:** VPC → Subnet → EC2 → RDS relationships
- ✅ **IaC integration:** Parse Terraform state to understand desired state
- ✅ **Graph storage:** TiDB or ClickHouse backend
- ✅ **JSON export:** Export graph for analysis or integration

---

## Architecture

```
┌─────────────────────────────────────────────────┐
│              SkyGraph Scanner                    │
└─────────────────────────────────────────────────┘
        │              │              │
        ▼              ▼              ▼
┌────────────┐ ┌────────────┐ ┌────────────┐
│  AWS API   │ │  GCP API   │ │  K8s API   │
│  Scanner   │ │  Scanner   │ │  Scanner   │
└────────────┘ └────────────┘ └────────────┘
        │              │              │
        └──────────────┼──────────────┘
                       ▼
        ┌──────────────────────────┐
        │    Graph Builder         │
        │  (Nodes + Edges)         │
        └──────────────────────────┘
                       │
                       ▼
        ┌──────────────────────────┐
        │   TiDB / ClickHouse      │
        └──────────────────────────┘
```

---

## Data Model

### ResourceNode

```go
type ResourceNode struct {
    ID         string            // Unique identifier (e.g., "aws:ec2:i-123456")
    Type       string            // Resource type (e.g., "ec2", "vpc", "rds")
    Provider   string            // Cloud provider (e.g., "aws", "gcp", "kubernetes")
    Region     string            // Region/zone
    Name       string            // Human-readable name
    Metadata   map[string]any    // Provider-specific attributes
    Tags       map[string]string // Resource tags
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

### Edge

```go
type Edge struct {
    From   string  // Source node ID
    To     string  // Target node ID
    Type   string  // Relationship type (e.g., "network", "dependency", "ownership")
    Weight float64 // Optional weight (e.g., bandwidth, cost)
}
```

### Graph

```go
type Graph struct {
    Nodes []ResourceNode
    Edges []Edge
}
```

---

## Supported Resources

### AWS (v0.1.0 MVP)
- [x] EC2 instances
- [x] VPC
- [x] Subnets
- [x] Security Groups
- [x] RDS instances
- [ ] Lambda functions
- [ ] S3 buckets
- [ ] ELB/ALB

### Kubernetes (v0.2.0)
- [ ] Pods
- [ ] Services
- [ ] Deployments
- [ ] ConfigMaps
- [ ] Secrets
- [ ] Ingress

### GCP (v0.3.0)
- [ ] Compute instances
- [ ] VPC networks
- [ ] Cloud SQL
- [ ] GKE clusters

### Azure (v0.4.0)
- [ ] Virtual machines
- [ ] Virtual networks
- [ ] Azure SQL

---

## Usage

### Installation

```bash
go install github.com/yourusername/airdig/skygraph/cmd/skygraph@latest
```

### Scan AWS

```bash
# Scan all resources in default region
skygraph scan --provider aws

# Scan specific region
skygraph scan --provider aws --region us-east-1

# Export to JSON
skygraph scan --provider aws --output graph.json

# Store in TiDB
skygraph scan --provider aws --store tidb --dsn "root@tcp(localhost:4000)/airdig"
```

### Scan Kubernetes

```bash
# Use current kubeconfig context
skygraph scan --provider kubernetes

# Specific namespace
skygraph scan --provider kubernetes --namespace production
```

---

## Configuration

SkyGraph uses a YAML configuration file:

```yaml
# skygraph.yaml
providers:
  - name: aws
    enabled: true
    regions:
      - us-east-1
      - us-west-2
    resources:
      - ec2
      - vpc
      - rds
      - lambda

  - name: kubernetes
    enabled: true
    kubeconfig: ~/.kube/config
    namespaces:
      - default
      - production

storage:
  type: tidb
  dsn: "root@tcp(localhost:4000)/airdig"

export:
  format: json
  path: ./output/graph.json
```

Run with config:

```bash
skygraph scan --config skygraph.yaml
```

---

## Development

### Prerequisites

- Go 1.21+
- AWS credentials configured (`~/.aws/credentials` or env vars)
- (Optional) TiDB or ClickHouse for storage

### Build

```bash
cd skygraph
go build -o bin/skygraph ./cmd/skygraph
```

### Run

```bash
./bin/skygraph scan --provider aws
```

### Test

```bash
go test ./...
```

---

## Roadmap

### v0.1.0 (Current)
- [x] Project structure
- [x] Data model definition
- [ ] AWS scanner (EC2, VPC, RDS, SG)
- [ ] Graph builder
- [ ] JSON export
- [ ] CLI tool

### v0.2.0
- [ ] Kubernetes scanner
- [ ] TiDB storage backend
- [ ] Incremental scanning (delta updates)
- [ ] Terraform state parser integration

### v0.3.0
- [ ] GCP scanner
- [ ] ClickHouse storage backend
- [ ] GraphQL query API
- [ ] Real-time updates (event-driven)

---

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for development guidelines.

---

## License

Apache License 2.0 — see [LICENSE](../LICENSE)

---

**SkyGraph** is part of the [AirDig](https://github.com/yourusername/airdig) project.
