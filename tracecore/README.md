# TraceCore

Distributed tracing & APM component for the AirDig platform.

## Architecture

```
Application (OTel)
    ↓ OTLP/gRPC:4317 or HTTP:4318
TraceCore OTLP Receiver
    ↓
Trace Processor
    ├→ Store Traces (ClickHouse)
    ├→ Service Map Generator (5-min windows)
    └→ Resource Correlator (link to SkyGraph)
         ↓
API Server :8082
    ↓ HTTP/JSON
React UI (Cytoscape.js)
```

## Features

- **OTLP Trace Ingestion**: Receive traces via gRPC (port 4317) and HTTP (port 4318)
- **Service Map Generation**: Auto-generate service dependency graphs with latency/error metrics
- **Infrastructure Correlation**: Link services to AWS resources (EC2, Lambda, EKS) via SkyGraph
- **Drift Correlation**: Analyze trace metrics before/after infrastructure changes (DeepDrift integration)
- **Interactive UI**: Cytoscape.js-based service map visualization
- **Tempo Export**: Export traces to Grafana Tempo for long-term storage

## Technology Stack

### Backend
- Go 1.23
- OpenTelemetry Collector libraries
- ClickHouse (shared with DeepDrift)
- gRPC + HTTP/REST

### Frontend
- React 19 + TypeScript
- Vite
- Cytoscape.js (graph visualization)

## API Endpoints

```
GET  /health
GET  /api/v1/status
GET  /api/v1/servicemap?start={time}&end={time}&service={name}
GET  /api/v1/services
GET  /api/v1/traces?start={time}&end={time}&service={name}&limit={n}
GET  /api/v1/traces/{trace_id}
GET  /api/v1/correlation/drift/{drift_event_id}?window={duration}
```

## Quick Start

```bash
# Start TraceCore server
go run cmd/tracecore/main.go

# Start UI development server
cd ui && npm run dev
```

## Configuration

TraceCore uses environment variables for configuration:

- `TRACECORE_OTLP_GRPC_PORT`: OTLP gRPC receiver port (default: 4317)
- `TRACECORE_OTLP_HTTP_PORT`: OTLP HTTP receiver port (default: 4318)
- `TRACECORE_API_PORT`: HTTP API server port (default: 8082)
- `CLICKHOUSE_ADDR`: ClickHouse server address (default: localhost:9000)
- `SKYGRAPH_URL`: SkyGraph API URL (default: http://localhost:8001)
- `DEEPDRIFT_URL`: DeepDrift API URL (default: http://localhost:8080)

## Development Status

**Phase 1 (In Progress)**: Core Infrastructure
- [x] Project initialization
- [ ] Data models
- [ ] ClickHouse client & schema
- [ ] Trace storage
- [ ] OTLP receiver
- [ ] Health check endpoint

## License

MIT
