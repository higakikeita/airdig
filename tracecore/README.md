# TraceCore — Distributed Tracing & APM

**TraceCore** is the third pillar of AirDig. It provides distributed tracing and application performance monitoring (APM) by integrating with OpenTelemetry and correlating application behavior with infrastructure changes.

---

## Overview

TraceCore ingests distributed traces from applications instrumented with OpenTelemetry and:
- **Generates service maps:** Visualize service dependencies and call graphs
- **Correlates with infrastructure:** Link traces to cloud resources (from SkyGraph)
- **Detects anomalies:** Identify latency spikes correlated with drift events
- **Stores traces:** Integrates with Tempo or ClickHouse for long-term storage

---

## Features

- ✅ **OpenTelemetry-native:** Built on OTel standards (OTLP protocol)
- ✅ **Distributed tracing:** Full support for traces and spans
- ✅ **Service map generation:** Automatic call graph visualization
- ✅ **Infrastructure correlation:** Link traces to EC2, pods, RDS, etc.
- ✅ **Drift correlation:** Detect when infrastructure changes affect performance
- ✅ **Tempo integration:** Export to Grafana Tempo for storage and querying

---

## Architecture

```
┌─────────────────────────────────────────────────┐
│      Application (OTel Instrumented)             │
└─────────────────────────────────────────────────┘
                       │
                       ▼ (OTLP)
┌─────────────────────────────────────────────────┐
│     OpenTelemetry Collector                      │
│       (TraceCore Processor)                      │
└─────────────────────────────────────────────────┘
                       │
        ┌──────────────┼──────────────┐
        ▼              ▼              ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│ Service Map │ │  Resource   │ │   Tempo     │
│  Generator  │ │  Correlator │ │  Exporter   │
└─────────────┘ └─────────────┘ └─────────────┘
        │              │              │
        ▼              ▼              ▼
┌─────────────────────────────────────────────────┐
│         AirDig Engine (Graph)                    │
└─────────────────────────────────────────────────┘
```

---

## Data Model

### TraceEdge

```go
type TraceEdge struct {
    From        string        // Source service name
    To          string        // Target service name
    Latency     time.Duration // Average latency
    ErrorRate   float64       // Error rate (0.0 - 1.0)
    RequestRate float64       // Requests per second
    TraceIDs    []string      // Sample trace IDs
    Timestamp   time.Time
}
```

### ServiceNode

```go
type ServiceNode struct {
    Name         string            // Service name
    Type         string            // web, api, database, etc.
    ResourceID   string            // Linked cloud resource (e.g., "aws:ec2:i-123")
    Endpoints    []string          // HTTP endpoints
    Metadata     map[string]any
}
```

---

## Usage

### Prerequisites

1. **Instrument your application with OpenTelemetry:**

```go
// Example: Go application
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func main() {
    exporter, _ := otlptracegrpc.New(context.Background(),
        otlptracegrpc.WithEndpoint("localhost:4317"),
        otlptracegrpc.WithInsecure(),
    )

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
    )

    otel.SetTracerProvider(tp)

    // Your application code...
}
```

2. **Deploy TraceCore Collector:**

```bash
docker run -p 4317:4317 -p 4318:4318 \
  -v $(pwd)/tracecore.yaml:/etc/tracecore.yaml \
  airdig/tracecore:latest
```

### Configuration

```yaml
# tracecore.yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 10s
    send_batch_size: 1024

  airdig:
    # TraceCore custom processor
    resource_correlation:
      enabled: true
      skygraph_api: http://skygraph:8080

    service_map:
      enabled: true
      window: 5m

exporters:
  tempo:
    endpoint: tempo:4317
    insecure: true

  airdig_engine:
    endpoint: http://airdig-engine:8080/traces
    api_key: ${AIRDIG_API_KEY}

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch, airdig]
      exporters: [tempo, airdig_engine]
```

### Querying Traces

```bash
# List services
tracecore services

# Get service map
tracecore servicemap --output graph.json

# Query traces by service
tracecore query --service frontend --start 1h

# Find traces with errors
tracecore query --error --start 24h

# Correlate with drift
tracecore correlate --drift-event sg-123456 --window 1h
```

---

## Service Map

TraceCore automatically generates a service map from observed traces:

```
┌─────────────┐
│   frontend  │
└─────────────┘
       │
       ├──────────────┐
       ▼              ▼
┌─────────────┐ ┌─────────────┐
│     api     │ │   cache     │
└─────────────┘ └─────────────┘
       │
       ├──────────────┐
       ▼              ▼
┌─────────────┐ ┌─────────────┐
│  database   │ │  queue      │
└─────────────┘ └─────────────┘
```

Each edge shows:
- **Latency** (p50, p95, p99)
- **Error rate**
- **Request rate**

---

## Infrastructure Correlation

TraceCore links services to cloud resources using resource attributes:

```go
// In your application, add resource attributes:
resource.New(context.Background(),
    resource.WithAttributes(
        semconv.ServiceName("frontend"),
        attribute.String("cloud.provider", "aws"),
        attribute.String("cloud.resource.id", "i-123456"),
        attribute.String("k8s.pod.name", "frontend-7d8f9"),
    ),
)
```

TraceCore then:
1. Extracts `cloud.resource.id` from spans
2. Queries SkyGraph for the resource node
3. Creates a link: `ServiceNode → ResourceNode`

**In AirDig UI:**
- Click on an EC2 instance → see which services run on it
- Click on a service → see which resources it depends on

---

## Drift Correlation

When a drift event occurs (e.g., security group change), TraceCore can:

1. **Query traces in the time window** around the drift event
2. **Compare latency before/after** the change
3. **Generate a correlation report:**

```bash
$ tracecore correlate --drift-event sg-123456 --window 1h

Correlation Report:
- Drift Event: sg-123456 (security group modified)
- Time: 2024-01-15 14:32:18 UTC
- Affected Service: api

Performance Impact:
- Before change: p95 latency = 120ms
- After change: p95 latency = 450ms
- Error rate increase: 0.1% → 2.3%

Sample Traces:
- trace-abc123 (500ms, error: connection timeout)
- trace-def456 (620ms, error: connection refused)
```

---

## Example: End-to-End Trace

```
Trace ID: abc123def456
Duration: 450ms
Status: ERROR

Spans:
1. frontend.handleRequest (200ms)
   └─> 2. api.processOrder (400ms)
           └─> 3. database.query (380ms) ← SLOW
                   Error: connection timeout

Resource: database → RDS instance (rds-prod-01)
Drift Event: Security group sg-123456 modified at 14:32
Impact: Database connection blocked by new SG rule
```

**Root Cause:** Security group change blocked database connections.

---

## Development

### Prerequisites

- Go 1.21+
- OpenTelemetry Collector Contrib (for base)

### Build Custom Collector

```bash
cd tracecore
go build -o bin/tracecore ./cmd/tracecore
```

### Run Locally

```bash
# Start Tempo (for trace storage)
docker run -p 3200:3200 -p 4317:4317 grafana/tempo:latest

# Start TraceCore
./bin/tracecore --config tracecore.yaml
```

### Send Test Traces

```bash
# Using telemetrygen
go install github.com/open-telemetry/opentelemetry-collector-contrib/cmd/telemetrygen@latest

telemetrygen traces --otlp-insecure --rate 10 --duration 60s
```

---

## Roadmap

### v0.1.0 (Current)
- [ ] OpenTelemetry Collector setup
- [ ] Custom processor for resource correlation
- [ ] Service map generator
- [ ] Tempo exporter
- [ ] CLI tool

### v0.2.0
- [ ] SkyGraph integration (resource linking)
- [ ] Drift correlation engine
- [ ] GraphQL query API
- [ ] Web UI (service map view)

### v0.3.0
- [ ] Anomaly detection (latency spikes)
- [ ] Auto-instrumentation support
- [ ] Metrics correlation (RED method)
- [ ] Cost attribution (trace → resource → cost)

---

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for development guidelines.

---

## License

Apache License 2.0 — see [LICENSE](../LICENSE)

---

**TraceCore** is part of the [AirDig](https://github.com/yourusername/airdig) project.
