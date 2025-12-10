# AirDig Development Roadmap

## Overview

This roadmap outlines the development plan for AirDig, organized by milestones and features. Each pillar (SkyGraph, DeepDrift, TraceCore, PulseSight) will be developed in parallel, then integrated into the unified AirDig Engine.

---

## Milestone Timeline

```
v0.1.0 → v0.2.0 → v0.3.0 → v0.4.0 → v1.0.0
  MVP     Pillars  Integration Engine   Release
```

---

## v0.1.0 — Foundation (Current Phase)

**Goal:** Establish project structure and core data models

### Completed
- [x] Project structure and GitHub setup
- [x] Documentation (architecture, vision, roadmap)
- [x] License and contribution guidelines

### In Progress
- [ ] SkyGraph MVP (AWS resource scanning)
- [ ] DeepDrift core engine (TFDrift integration)

### Tasks
- [ ] Define unified graph data model
- [ ] Set up development environment (Docker Compose)
- [ ] Create example configurations
- [ ] Set up CI/CD pipeline (GitHub Actions)

**Target:** 2024 Q1

---

## v0.2.0 — Pillar Development

**Goal:** Build MVP for each of the four pillars independently

### 🟦 SkyGraph
- [ ] AWS resource scanner (EC2, VPC, RDS, Lambda, S3)
- [ ] Kubernetes resource scanner (pods, services, deployments)
- [ ] Graph builder (nodes + edges)
- [ ] TiDB/ClickHouse storage backend
- [ ] JSON export functionality
- [ ] Basic CLI tool (`skygraph scan`)

### 🟢 DeepDrift
- [ ] Terraform state parser
- [ ] AWS API state fetcher
- [ ] Diff engine (compare desired vs actual)
- [ ] CloudTrail correlation
- [ ] Drift event storage
- [ ] CLI tool (`deepdrift detect`)

### 🟣 TraceCore
- [ ] OpenTelemetry Collector setup
- [ ] Custom processor (map traces to resources)
- [ ] Tempo integration (trace storage)
- [ ] Service map generator
- [ ] Basic query API

### 🟡 PulseSight
- [ ] Prometheus metrics ingestion
- [ ] Loki log collection
- [ ] Falco event listener (optional)
- [ ] Health status evaluator
- [ ] Alert integration

**Target:** 2024 Q2

---

## v0.3.0 — Integration Layer

**Goal:** Integrate all four pillars into AirDig Engine

### AirDig Engine
- [ ] Unified graph model
- [ ] Event-driven architecture (Kafka/NATS)
- [ ] GraphQL API
- [ ] Real-time graph updates
- [ ] Data correlation (drift ↔ traces ↔ metrics)
- [ ] Query interface

### Data Platform
- [ ] TiDB cluster setup
- [ ] Tempo deployment (S3 backend)
- [ ] Mimir deployment (Prometheus storage)
- [ ] Loki deployment
- [ ] Redis cache layer

### Integrations
- [ ] CloudTrail ingestion
- [ ] Kubernetes watch events
- [ ] Terraform webhook integration

**Target:** 2024 Q3

---

## v0.4.0 — User Interface

**Goal:** Build the unified UI for AirDig

### UI Components
- [ ] Next.js project setup
- [ ] Cytoscape.js graph visualization
- [ ] Tab-based views (SkyGraph, DeepDrift, TraceCore, PulseSight)
- [ ] Unified view (all pillars combined)
- [ ] Node/edge inspector panel
- [ ] Timeline view (changes over time)
- [ ] Search and filter
- [ ] Real-time updates (WebSocket)

### User Experience
- [ ] Interactive graph (zoom, pan, click)
- [ ] Node coloring (health status)
- [ ] Edge highlighting (drift, traces)
- [ ] Tooltips and context menus
- [ ] Keyboard shortcuts
- [ ] Dark mode

**Target:** 2024 Q4

---

## v1.0.0 — Public Release

**Goal:** Production-ready, fully documented, OSS release

### Core Features
- [x] SkyGraph: Multi-cloud scanning (AWS, GCP, Azure, K8s)
- [x] DeepDrift: Full drift detection + CloudTrail correlation
- [x] TraceCore: OpenTelemetry-native APM
- [x] PulseSight: Metrics, logs, runtime security
- [x] AirDig Engine: Unified graph + GraphQL API
- [x] UI: Complete web interface

### Documentation
- [ ] Installation guide
- [ ] User guide
- [ ] API reference
- [ ] Architecture deep-dive
- [ ] Contribution guide
- [ ] Tutorial videos

### Deployment
- [ ] Docker Compose quickstart
- [ ] Kubernetes Helm chart
- [ ] Terraform modules (self-hosting)
- [ ] Cloud marketplace (AWS/GCP)

### Community
- [ ] Website (airdig.io)
- [ ] Blog posts
- [ ] Conference talks
- [ ] GitHub Discussions
- [ ] Slack/Discord community

**Target:** 2025 Q1

---

## Post-1.0 Features

### Intelligence & Automation
- [ ] AI-driven root cause analysis
- [ ] Predictive impact analysis (before changes)
- [ ] Anomaly detection on graph patterns
- [ ] Auto-remediation suggestions
- [ ] Cost analysis and optimization

### Enterprise Features
- [ ] Multi-tenancy
- [ ] SSO/RBAC
- [ ] Compliance reporting (SOC2, ISO27001)
- [ ] Audit logging
- [ ] SLA monitoring

### Extensibility
- [ ] Plugin system
- [ ] Custom scanner SDK
- [ ] Webhook integrations
- [ ] Marketplace

### Additional Clouds
- [ ] Oracle Cloud
- [ ] IBM Cloud
- [ ] Alibaba Cloud
- [ ] DigitalOcean, Linode, etc.

---

## Development Principles

### Parallel Development
- All four pillars are developed independently
- Each pillar has its own repository/module
- Integration happens at the Engine layer

### MVP-First
- Ship minimal viable features quickly
- Iterate based on feedback
- Avoid over-engineering

### Community-Driven
- Open development (GitHub)
- Public roadmap (GitHub Projects)
- RFCs for major features
- Community contributions welcome

### Quality Standards
- Unit tests (80%+ coverage)
- Integration tests
- Documentation for all features
- Performance benchmarks

---

## How to Contribute

See specific roadmap items you want to work on?

1. **Check GitHub Issues:** Look for issues tagged with the milestone
2. **Join Discussions:** Share your ideas in [Discussions](https://github.com/yourusername/airdig/discussions)
3. **Submit RFCs:** For major features, submit an RFC (Request for Comments)
4. **Open PRs:** Start small, iterate fast

---

## Tracking Progress

- **GitHub Projects:** [AirDig Project Board](https://github.com/yourusername/airdig/projects)
- **Milestones:** [GitHub Milestones](https://github.com/yourusername/airdig/milestones)
- **Releases:** [GitHub Releases](https://github.com/yourusername/airdig/releases)

---

## Contact

- **Maintainers:** [MAINTAINERS.md](../MAINTAINERS.md)
- **Discussions:** [GitHub Discussions](https://github.com/yourusername/airdig/discussions)
- **Email:** airdig@example.com

---

**Last Updated:** 2024-01-01
