# AirDig Vision

## The Problem

Modern cloud infrastructure has become incredibly complex:

- **Multiple layers:** Cloud resources, Kubernetes, containers, applications
- **Constant change:** Infrastructure drifts from IaC, configurations change without notice
- **Blind spots:** No single tool shows the full picture from cloud topology to application behavior
- **Fragmented tools:** Security teams use Wiz, ops teams use Datadog, platform teams use Terraform—no integration

**The result?** When something breaks, teams spend hours correlating data from 5+ different tools to understand what changed, where, and why.

---

## The Vision

**AirDig is the unified observability platform that sees your entire cloud—from the sky to the core.**

Inspired by:
- **Sysdig** — Deep system visibility at the OS/container level
- **Falco** — Runtime security monitoring
- **Stratoshark** — Cloud API observability

AirDig extends this philosophy **upward** into the cloud, application, and infrastructure layers.

### Core Principles

1. **Graph-First:** Everything is a node (resource, service, process) and everything has edges (dependencies, calls, changes)
2. **Unified View:** One platform, one graph, one UI—from cloud topology to distributed traces
3. **Change-Aware:** Not just "what exists" but "what changed, when, why, and what broke"
4. **Real-Time:** Live updates, not batch jobs
5. **Open Source:** Built on open standards (OpenTelemetry, Prometheus, Falco)

---

## What Makes AirDig Different?

### 1. Unified Graph Model

Every observability tool today operates in a silo:
- **Wiz** shows cloud config but not application behavior
- **Datadog** shows APM but not IaC drift
- **Sysdig** shows runtime but not cloud topology

**AirDig unifies all of this in a single graph.**

A single EC2 instance in AirDig is:
- A **node** in the cloud topology (from SkyGraph)
- Connected to **drift events** (from DeepDrift)
- Emitting **traces** (from TraceCore)
- Reporting **metrics** and **security events** (from PulseSight)

### 2. Drift Intelligence

AirDig doesn't just detect drift—it understands it:
- **What changed?** (Security group rule added)
- **Who changed it?** (CloudTrail correlation)
- **What broke?** (Increased latency in APM)
- **What's the impact?** (Graph-based blast radius analysis)

This is **change intelligence**, not just change detection.

### 3. The "Digital Twin" of Your Cloud

AirDig maintains a real-time digital twin of your entire cloud infrastructure:
- Every resource, every connection, every change
- Updated in real-time via CloudTrail, K8s watches, OTel streams
- Visualized in a beautiful, interactive graph UI

This is the **single source of truth** for your infrastructure.

---

## Use Cases

### 1. "Why is my app slow?"

**Before AirDig:**
1. Check Datadog APM → see latency spike
2. Check AWS Console → scroll through hundreds of resources
3. Check Terraform state → manually diff
4. Check CloudTrail → sift through JSON logs
5. Finally find: someone changed the RDS instance type 10 minutes ago

**With AirDig:**
1. Open AirDig UI
2. See red node on RDS instance
3. Click → see drift annotation: "RDS instance type changed"
4. See correlated trace showing increased query latency
5. See CloudTrail event: "user@example.com changed instance type"
6. Total time: **30 seconds**

### 2. "Who changed this security group?"

**Before AirDig:**
1. Check AWS Console → find security group
2. Open CloudTrail → search by resource ID
3. Parse JSON logs → find user
4. Check what other resources are affected → manual process

**With AirDig:**
1. Open AirDig → navigate to security group node
2. See drift event with CloudTrail link
3. Click "Impact Analysis" → graph highlights all affected EC2 instances
4. Total time: **15 seconds**

### 3. "Is my infrastructure in sync with Terraform?"

**Before AirDig:**
1. Run `terraform plan` → see 47 changes
2. Manually review each change
3. Realize someone made manual changes in AWS Console
4. Spend hours reconciling state

**With AirDig:**
1. Open DeepDrift view
2. See visual diff: 47 resources highlighted
3. Filter by "manual changes" → 5 resources
4. Click each → see CloudTrail event + who did it
5. Auto-generate Terraform import commands
6. Total time: **5 minutes**

---

## The AirDig Philosophy

### "Dig the Cloud"

Just as **Sysdig** digs deep into system calls and container behavior, **AirDig** digs deep into cloud infrastructure and application behavior.

### "See Everything"

Observability is not about having 10 dashboards. It's about having **one unified view** that shows:
- Structure (topology)
- Change (drift)
- Behavior (traces)
- Health (metrics/logs)

### "From Sky to Core"

AirDig sees your cloud from multiple perspectives:
- **Sky** (high-level topology, multi-cloud view)
- **Layers** (VPCs, subnets, pods, containers)
- **Core** (processes, system calls, network flows via eBPF)

---

## Inspiration: Standing on the Shoulders of Giants

AirDig inherits the DNA of:

### Sysdig
- Deep visibility philosophy
- System call tracing
- Container-native design

### Falco
- Runtime security monitoring
- Rule-based alerting
- eBPF-powered observability

### Stratoshark
- Cloud API observability
- CloudTrail intelligence
- High-altitude infrastructure view

### OpenTelemetry
- Vendor-neutral standards
- Distributed tracing
- Unified telemetry model

### CloudGraph
- Graph-based resource modeling
- Multi-cloud scanning
- Dependency mapping

**AirDig combines the best ideas from all of these into a single, unified platform.**

---

## Design Goals

1. **Open Source First:** Built in public, MIT/Apache licensed
2. **Vendor Neutral:** Works with AWS, GCP, Azure, K8s
3. **Standards-Based:** OpenTelemetry, Prometheus, Falco
4. **Beautiful UX:** Engineers should *want* to use it
5. **Scalable:** From startups to enterprises
6. **Extensible:** Plugin architecture for custom scanners/analyzers

---

## Success Metrics

AirDig will be successful when:

1. **Time to root cause** drops from hours to minutes
2. Engineers have **one tool** instead of five
3. Drift becomes **visible** and **actionable**
4. Infrastructure changes are **predictable** (impact analysis before apply)
5. Security and operations teams share **one graph**

---

## The Future

### Phase 1: Foundation (Current)
- SkyGraph: Cloud topology mapping
- DeepDrift: Drift detection
- TraceCore: APM integration
- PulseSight: Metrics/logs/runtime

### Phase 2: Intelligence
- AI-driven root cause analysis
- Predictive impact analysis (before changes are applied)
- Anomaly detection on graph patterns
- Auto-remediation suggestions

### Phase 3: Platform
- Multi-tenancy (per-team graphs)
- GitOps integration (sync with IaC repos)
- Custom plugin marketplace
- Cost optimization (map spending to graph)

### Phase 4: Ecosystem
- SaaS offering
- Enterprise features (SSO, audit logs, compliance)
- Community-driven scanner library
- Integration marketplace

---

## Join Us

AirDig is an open-source project built by engineers, for engineers.

**We believe:**
- Observability should be unified, not fragmented
- Change should be visible and understood
- Cloud infrastructure should have a digital twin
- Tools should be beautiful and delightful to use

**If you believe this too, join us.**

- Contribute code: [GitHub](https://github.com/yourusername/airdig)
- Share ideas: [Discussions](https://github.com/yourusername/airdig/discussions)
- Report issues: [Issues](https://github.com/yourusername/airdig/issues)

---

**AirDig — From Sky to Core. Your Entire Cloud, Visualized.**
