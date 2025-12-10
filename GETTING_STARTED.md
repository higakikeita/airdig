# Getting Started with AirDig

Welcome to **AirDig** — the unified cloud observability platform!

---

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/airdig.git
cd airdig
```

### 2. Try the SkyGraph Demo

```bash
# Build SkyGraph
cd skygraph
go build -o bin/skygraph ./cmd/skygraph

# Run demo (generates a sample graph)
./bin/skygraph
```

This will create a `graph.json` file with a demo AWS infrastructure graph.

---

## Project Structure

```
airdig/
├── README.md              # Main project overview
├── LICENSE                # Apache 2.0 license
├── .gitignore
├── docs/                  # Documentation
│   ├── architecture.md    # System architecture
│   ├── vision.md          # Project vision and philosophy
│   └── roadmap.md         # Development roadmap
├── skygraph/              # Pillar 1: Cloud topology
│   ├── README.md
│   ├── cmd/skygraph/      # CLI tool
│   └── pkg/graph/         # Graph data model
├── deepdrift/             # Pillar 2: Drift detection
│   └── README.md
├── tracecore/             # Pillar 3: APM & tracing
│   └── README.md
├── pulsesight/            # Pillar 4: Metrics & logs
│   └── README.md
├── engine/                # AirDig Engine (v0.3.0)
│   └── README.md
└── ui/                    # Web UI (v0.4.0)
    └── README.md
```

---

## The Four Pillars

AirDig is built on four independent but integrated pillars:

### 🟦 1. [SkyGraph](./skygraph/README.md)
**Cloud topology and dependency graph**
- Scans AWS, GCP, Azure, Kubernetes
- Builds a unified resource graph
- Foundation for all other pillars

**Status:** 🟡 In Development (v0.1.0)

### 🟢 2. [DeepDrift](./deepdrift/README.md)
**Infrastructure drift detection**
- Compares Terraform state vs actual cloud state
- Correlates with CloudTrail events
- Impact analysis using SkyGraph

**Status:** 🟢 Existing (TFDrift integration)

### 🟣 3. [TraceCore](./tracecore/README.md)
**Distributed tracing & APM**
- OpenTelemetry-native
- Service map generation
- Correlates traces with infrastructure

**Status:** 🔴 Planned (v0.2.0)

### 🟡 4. [PulseSight](./pulsesight/README.md)
**Metrics, logs, and runtime security**
- Prometheus metrics
- Loki logs
- Falco runtime events
- Health status for resources

**Status:** 🔴 Planned (v0.2.0)

---

## Development Roadmap

- **v0.1.0** (Current): Foundation — Project structure, SkyGraph MVP
- **v0.2.0**: Pillar Development — All four pillars reach MVP
- **v0.3.0**: Integration — AirDig Engine unifies all pillars
- **v0.4.0**: UI — Web interface with graph visualization
- **v1.0.0**: Public Release — Production-ready OSS release

See [docs/roadmap.md](./docs/roadmap.md) for details.

---

## Documentation

- **[Architecture](./docs/architecture.md)** — Technical deep-dive
- **[Vision](./docs/vision.md)** — Philosophy and goals
- **[Roadmap](./docs/roadmap.md)** — Development plan

---

## Contributing

We welcome contributions! Here's how to get started:

1. **Pick a pillar** — Choose SkyGraph, DeepDrift, TraceCore, or PulseSight
2. **Check the issues** — Look for open issues in that pillar
3. **Join discussions** — Share your ideas
4. **Submit PRs** — Start small, iterate fast

See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

---

## Community

- **GitHub Issues:** [Report bugs or request features](https://github.com/yourusername/airdig/issues)
- **Discussions:** [Ask questions and share ideas](https://github.com/yourusername/airdig/discussions)
- **Slack:** Coming soon

---

## License

AirDig is licensed under the [Apache License 2.0](./LICENSE).

---

## Acknowledgments

AirDig is inspired by:
- **Sysdig** — Deep system visibility
- **Falco** — Runtime security
- **Stratoshark** — Cloud API observability
- **OpenTelemetry** — Distributed tracing
- **CloudGraph** — Graph-based cloud modeling

---

**Ready to dig into the cloud? Let's build AirDig together!**
