# PulseSight — Metrics, Logs & Runtime Observability

**PulseSight** is the fourth pillar of AirDig. It collects metrics, logs, and runtime security events to provide real-time health monitoring and security visibility for cloud infrastructure.

---

## Overview

PulseSight integrates with:
- **Prometheus** for metrics (CPU, memory, network, custom metrics)
- **Loki** for log aggregation
- **Falco** for runtime security events
- **eBPF** for deep system observability (optional)

PulseSight then:
- **Assigns health status** to resources in SkyGraph
- **Correlates metrics with drift events** (e.g., CPU spike after config change)
- **Surfaces security alerts** in the AirDig UI

---

## Features

- ✅ **Prometheus integration:** Scrape metrics from exporters
- ✅ **Loki integration:** Aggregate logs from multiple sources
- ✅ **Falco integration:** Runtime security event detection
- ✅ **Health evaluation:** Assign "healthy", "warning", "critical" status to resources
- ✅ **Alert correlation:** Link Prometheus alerts to graph nodes
- ✅ **eBPF support:** Deep network and process observability (optional)

---

## Architecture

```
┌─────────────────────────────────────────────────┐
│          Infrastructure & Applications           │
│   (Prometheus Exporters, Logs, Falco)           │
└─────────────────────────────────────────────────┘
        │              │              │
        ▼              ▼              ▼
┌────────────┐ ┌────────────┐ ┌────────────┐
│ Prometheus │ │    Loki    │ │   Falco    │
│  Scraper   │ │ Log Client │ │  Listener  │
└────────────┘ └────────────┘ └────────────┘
        │              │              │
        └──────────────┼──────────────┘
                       ▼
        ┌──────────────────────────┐
        │   PulseSight Processor   │
        │  (Health Evaluator)      │
        └──────────────────────────┘
                       │
                       ▼
        ┌──────────────────────────┐
        │  AirDig Engine (Graph)   │
        │  (Update Node Health)    │
        └──────────────────────────┘
```

---

## Data Model

### HealthStatus

```go
type HealthStatus struct {
    Status      HealthLevel       // healthy, warning, critical
    LastCheck   time.Time
    Metrics     map[string]float64 // e.g., {"cpu_usage": 85.2, "memory_usage": 72.1}
    Alerts      []Alert
    Reason      string            // Human-readable explanation
}

type HealthLevel string

const (
    Healthy  HealthLevel = "healthy"
    Warning  HealthLevel = "warning"
    Critical HealthLevel = "critical"
)
```

### RuntimeEvent

```go
type RuntimeEvent struct {
    ID          string
    Type        EventType         // falco_alert, network_flow, process_spawn
    Severity    Severity          // low, medium, high, critical
    ResourceID  string            // Linked resource (e.g., "k8s:pod:frontend-7d8f9")
    Timestamp   time.Time
    Message     string
    Metadata    map[string]any
}

type EventType string

const (
    FalcoAlert     EventType = "falco_alert"
    NetworkFlow    EventType = "network_flow"
    ProcessSpawn   EventType = "process_spawn"
    FileAccess     EventType = "file_access"
)
```

---

## Usage

### Installation

```bash
go install github.com/yourusername/airdig/pulsesight/cmd/pulsesight@latest
```

### Configuration

```yaml
# pulsesight.yaml
prometheus:
  enabled: true
  remote_write:
    url: http://mimir:9009/api/v1/push
  scrape_configs:
    - job_name: 'node'
      static_configs:
        - targets: ['localhost:9100']

    - job_name: 'kubernetes'
      kubernetes_sd_configs:
        - role: pod

loki:
  enabled: true
  url: http://loki:3100
  sources:
    - type: file
      path: /var/log/*.log
    - type: kubernetes
      namespace: production

falco:
  enabled: true
  rules:
    - /etc/falco/rules.yaml
  outputs:
    - type: webhook
      url: http://pulsesight:8080/events

health:
  evaluation_interval: 30s
  rules:
    - resource_type: ec2
      metric: cpu_usage
      thresholds:
        warning: 70
        critical: 90

    - resource_type: rds
      metric: db_connections
      thresholds:
        warning: 80
        critical: 95

storage:
  type: tidb
  dsn: "root@tcp(localhost:4000)/airdig"
```

### Run PulseSight

```bash
# Start PulseSight
pulsesight run --config pulsesight.yaml

# Query health status
pulsesight health --resource aws:ec2:i-123456

# List alerts
pulsesight alerts --severity critical --start 1h
```

---

## Metrics Collection

PulseSight uses **Prometheus** for metrics:

### Node Exporter (Infrastructure Metrics)

```bash
# Install node_exporter
docker run -d -p 9100:9100 prom/node-exporter

# PulseSight scrapes metrics every 15s
```

**Metrics:**
- CPU usage (`node_cpu_seconds_total`)
- Memory usage (`node_memory_MemAvailable_bytes`)
- Disk I/O (`node_disk_io_time_seconds_total`)
- Network traffic (`node_network_receive_bytes_total`)

### Kubernetes Metrics

```bash
# PulseSight auto-discovers K8s pods with Prometheus annotations
```

**Metrics:**
- Pod CPU/memory (`container_cpu_usage_seconds_total`)
- Pod restarts (`kube_pod_container_status_restarts_total`)

### Custom Application Metrics

```go
// Expose custom metrics in your application
import "github.com/prometheus/client_golang/prometheus/promhttp"

http.Handle("/metrics", promhttp.Handler())
http.ListenAndServe(":8080", nil)
```

---

## Log Collection

PulseSight uses **Loki** for log aggregation:

### File Logs

```yaml
loki:
  sources:
    - type: file
      path: /var/log/app/*.log
      labels:
        app: myapp
        env: production
```

### Kubernetes Logs

```yaml
loki:
  sources:
    - type: kubernetes
      namespace: production
      pod_selector:
        app: frontend
```

### Query Logs

```bash
# Query logs via PulseSight CLI
pulsesight logs --resource k8s:pod:frontend-7d8f9 --start 1h

# Filter by log level
pulsesight logs --level error --start 24h
```

---

## Runtime Security (Falco)

PulseSight integrates with **Falco** for runtime security monitoring:

### Falco Rules

```yaml
# /etc/falco/rules.yaml
- rule: Unauthorized Process Spawned
  desc: Detect unexpected processes in containers
  condition: spawned_process and container and not proc.name in (allowed_processes)
  output: Unauthorized process spawned (user=%user.name command=%proc.cmdline)
  priority: WARNING
```

### PulseSight Integration

When Falco triggers a rule, PulseSight:
1. Receives the webhook event
2. Extracts the resource ID (pod, container, host)
3. Creates a `RuntimeEvent` in AirDig
4. Updates the resource's health status to "warning" or "critical"

**In AirDig UI:**
- Nodes with security events are highlighted in red
- Click to see Falco rule details

---

## Health Evaluation

PulseSight continuously evaluates resource health based on metrics and alerts:

### Example: EC2 Instance Health

```go
// Pseudo-code
func EvaluateEC2Health(metrics map[string]float64, alerts []Alert) HealthStatus {
    cpuUsage := metrics["cpu_usage"]
    memUsage := metrics["memory_usage"]

    if cpuUsage > 90 || memUsage > 90 {
        return HealthStatus{Status: Critical, Reason: "High resource usage"}
    }

    if cpuUsage > 70 || memUsage > 70 {
        return HealthStatus{Status: Warning, Reason: "Elevated resource usage"}
    }

    if len(alerts) > 0 {
        return HealthStatus{Status: Warning, Reason: "Active alerts"}
    }

    return HealthStatus{Status: Healthy}
}
```

**Result:**
- Green node in AirDig UI = Healthy
- Yellow node = Warning
- Red node = Critical

---

## eBPF Integration (Optional)

For deep observability, PulseSight can use **eBPF** (via Cilium, Pixie, or custom probes):

### Network Flow Tracking

```bash
# PulseSight captures network flows using eBPF
pulsesight ebpf --mode network --interface eth0
```

**Use Cases:**
- Detect unusual network connections
- Track inter-service communication at the kernel level
- Correlate network issues with drift events

### Process Monitoring

```bash
# Track process spawns and file access
pulsesight ebpf --mode process
```

---

## Alerting

PulseSight forwards alerts to external systems:

### Slack Integration

```yaml
alerts:
  slack:
    webhook_url: https://hooks.slack.com/services/XXX
    channel: "#ops-alerts"
```

### PagerDuty Integration

```yaml
alerts:
  pagerduty:
    integration_key: ${PAGERDUTY_KEY}
```

---

## Example: Drift + Metrics Correlation

**Scenario:**
1. DeepDrift detects: Security group `sg-123456` modified (port 22 opened)
2. PulseSight detects: 5 minutes later, CPU usage on `i-123456` spikes to 95%
3. PulseSight creates a correlation event:
   - "Security group change may have caused increased load"
4. AirDig UI highlights both events on the graph

**Root Cause:**
Opening SSH allowed a brute-force attack, spiking CPU.

---

## Development

### Prerequisites

- Go 1.21+
- Prometheus (for metrics)
- Loki (for logs)
- Falco (optional, for runtime security)

### Build

```bash
cd pulsesight
go build -o bin/pulsesight ./cmd/pulsesight
```

### Run Locally

```bash
# Start Prometheus
docker run -p 9090:9090 prom/prometheus

# Start Loki
docker run -p 3100:3100 grafana/loki

# Start PulseSight
./bin/pulsesight run --config pulsesight.yaml
```

---

## Roadmap

### v0.1.0 (Current)
- [ ] Prometheus metrics scraper
- [ ] Loki log client
- [ ] Health evaluator
- [ ] Alert integration
- [ ] CLI tool

### v0.2.0
- [ ] Falco integration
- [ ] eBPF network flow tracking
- [ ] SkyGraph integration (update node health)
- [ ] GraphQL query API

### v0.3.0
- [ ] Anomaly detection (ML-based)
- [ ] Auto-scaling recommendations
- [ ] Cost analysis (metrics → cost)
- [ ] Custom health rules DSL

---

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for development guidelines.

---

## License

Apache License 2.0 — see [LICENSE](../LICENSE)

---

**PulseSight** is part of the [AirDig](https://github.com/yourusername/airdig) project.
