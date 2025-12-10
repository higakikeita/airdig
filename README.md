# AirDig — Dig the Cloud. See Everything.

![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)
![Status](https://img.shields.io/badge/status-alpha-orange.svg)

English | [日本語](./README.ja.md)

**AirDig** is a next-generation cloud observability and drift intelligence platform that visualizes your entire cloud infrastructure from the sky to the core.

Inspired by the deep visibility of **Sysdig** and the runtime security of **Falco**, AirDig extends observability beyond the OS layer into the cloud, application, and infrastructure layers.

---

## 🌟 What is AirDig?

AirDig unifies four critical observability pillars into a single, graph-based platform:

- **SkyGraph** — Cloud topology and infrastructure dependency visualization
- **DeepDrift** — Real-time drift detection and change intelligence
- **TraceCore** — Distributed tracing and application performance monitoring
- **PulseSight** — Metrics, logs, and runtime security events

Together, these pillars provide a **360° view of your cloud environment**, combining:
- Cloud resource configuration (AWS, GCP, Azure, Kubernetes)
- Infrastructure-as-Code (Terraform, CloudFormation) state and drift
- Application traces (OpenTelemetry)
- Runtime metrics (Prometheus) and security events (Falco, eBPF)

---

## 🧱 The Four Pillars

### 🟦 1. SkyGraph
**Cloud topology and dependency graph visualization**

- Scans cloud APIs (AWS, GCP, Azure, Kubernetes)
- Builds a unified resource graph
- Visualizes dependencies and network topology
- Integrates with IaC tools (Terraform, CDK)

### 🟢 2. DeepDrift
**Infrastructure drift detection and change intelligence**

- Detects drift between IaC desired state and actual cloud state
- Correlates changes with CloudTrail/audit logs
- Provides change impact analysis
- Integrates with TFDrift engine

### 🟣 3. TraceCore
**Distributed tracing and APM**

- Ingests OpenTelemetry traces
- Generates service maps
- Correlates application behavior with infrastructure changes
- Exports to Tempo/Jaeger

### 🟡 4. PulseSight
**Metrics, logs, and runtime observability**

- Ingests Prometheus metrics
- Collects logs (Loki)
- Runtime security events (Falco, eBPF)
- Resource health status tracking

---

## 🎯 Why AirDig?

| Feature | Datadog | Wiz | Sysdig | AirDig |
|---------|---------|-----|--------|--------|
| Cloud Config Graph | ❌ | ✅ | ❌ | ✅ |
| Drift Detection | ❌ | ❌ | ❌ | ✅ |
| APM / Tracing | ✅ | ❌ | ✅ | ✅ |
| Runtime Security | ⚠️ | ✅ | ✅ | ✅ |
| Unified Graph View | ❌ | ⚠️ | ❌ | ✅ |

AirDig is the **only platform** that unifies cloud topology, drift intelligence, APM, and runtime observability in a single graph-based view.

---

## 🚀 Quickstart

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- AWS/GCP/Azure credentials (for cloud scanning)
- Terraform (optional, for drift detection)

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/airdig.git
cd airdig

# Run the demo
docker-compose up -d

# Access the UI
open http://localhost:3000
```

---

## 📚 Documentation

- [Architecture](./docs/architecture.md) — System design and data flow
- [Vision](./docs/vision.md) — Project philosophy and goals
- [Roadmap](./docs/roadmap.md) — Development plan and milestones

### Component Documentation

- [SkyGraph](./skygraph/README.md) — Cloud graph engine
- [DeepDrift](./deepdrift/README.md) — Drift detection engine
- [TraceCore](./tracecore/README.md) — APM and tracing
- [PulseSight](./pulsesight/README.md) — Metrics and runtime observability

---

## 🛠️ Development Status

| Component | Status | Version |
|-----------|--------|---------|
| SkyGraph | 🟡 Alpha | v0.1.0 |
| DeepDrift | 🟢 Beta | v0.5.0 |
| TraceCore | 🔴 Planned | - |
| PulseSight | 🔴 Planned | - |
| AirDig Engine | 🔴 Planned | - |
| UI | 🔴 Planned | - |

---

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for details.

---

## 📝 License

AirDig is licensed under the [Apache License 2.0](./LICENSE).

---

## 🙏 Acknowledgments

AirDig stands on the shoulders of giants:

- **Sysdig** — For pioneering deep system visibility
- **Falco** — For runtime security innovation
- **Stratoshark** — For cloud API observability
- **OpenTelemetry** — For distributed tracing standards
- **Terraform** — For infrastructure-as-code

---

## 🔗 Links

- [Documentation](./docs/)
- [GitHub Issues](https://github.com/yourusername/airdig/issues)
- [Discussions](https://github.com/yourusername/airdig/discussions)

---

**AirDig — From Sky to Core. Your Entire Cloud, Visualized.**
