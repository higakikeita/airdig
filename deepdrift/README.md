# DeepDrift — Infrastructure Drift Detection & Change Intelligence

**DeepDrift** is the second pillar of AirDig. It detects drift between Infrastructure-as-Code (IaC) desired state and actual cloud state, correlates changes with audit logs, and provides impact analysis.

---

## Overview

DeepDrift extends the **TFDrift** engine to provide:
- **Drift detection:** Compare Terraform state vs actual AWS/GCP/Azure resources
- **Change correlation:** Link drift events to CloudTrail/audit logs
- **Impact analysis:** Use the SkyGraph dependency graph to predict blast radius
- **Real-time monitoring:** Continuously watch for configuration drift

---

## Features

- ✅ **Terraform integration:** Parse `terraform.tfstate` and compare with cloud APIs
- ✅ **Drift detection:** Identify created, modified, deleted resources
- ✅ **CloudTrail correlation:** Find who changed what, and when
- ✅ **Graph-based impact analysis:** Predict which resources are affected
- ✅ **Event storage:** Store drift events with full before/after state
- ✅ **CLI tool:** `deepdrift detect` for one-time scans or CI/CD integration

---

## Architecture

```
┌─────────────────────────────────────────────────┐
│            DeepDrift Engine                      │
└─────────────────────────────────────────────────┘
        │                          │
        ▼                          ▼
┌──────────────────┐    ┌──────────────────┐
│ Terraform State  │    │  Cloud API       │
│    Parser        │    │  State Fetcher   │
└──────────────────┘    └──────────────────┘
        │                          │
        └──────────┬───────────────┘
                   ▼
        ┌──────────────────────────┐
        │     Diff Engine          │
        │  (Compare & Generate     │
        │   Drift Events)          │
        └──────────────────────────┘
                   │
                   ▼
        ┌──────────────────────────┐
        │  CloudTrail Correlator   │
        └──────────────────────────┘
                   │
                   ▼
        ┌──────────────────────────┐
        │   Impact Analyzer        │
        │  (Uses SkyGraph)         │
        └──────────────────────────┘
                   │
                   ▼
        ┌──────────────────────────┐
        │   TiDB / ClickHouse      │
        │   (Drift Event Store)    │
        └──────────────────────────┘
```

---

## Data Model

### DriftEvent

```go
type DriftEvent struct {
    ID           string            // Unique event ID
    ResourceID   string            // Resource identifier (e.g., "aws:ec2:i-123456")
    Type         DriftType         // created, modified, deleted
    Timestamp    time.Time
    Before       map[string]any    // State before change
    After        map[string]any    // State after change
    Diff         map[string]any    // Detailed diff
    RootCause    string            // CloudTrail event ID or user
    ImpactedEdges []string         // List of affected graph edges
}

type DriftType string

const (
    DriftCreated  DriftType = "created"
    DriftModified DriftType = "modified"
    DriftDeleted  DriftType = "deleted"
)
```

---

## Usage

### Installation

```bash
go install github.com/yourusername/airdig/deepdrift/cmd/deepdrift@latest
```

### Detect Drift

```bash
# Compare Terraform state with AWS
deepdrift detect --state terraform.tfstate --provider aws

# Specify region
deepdrift detect --state terraform.tfstate --provider aws --region us-east-1

# Output to JSON
deepdrift detect --state terraform.tfstate --provider aws --output drift.json

# Continuous monitoring (daemon mode)
deepdrift watch --state terraform.tfstate --provider aws --interval 5m
```

### CloudTrail Correlation

```bash
# Correlate drift with CloudTrail events
deepdrift correlate --drift-file drift.json --cloudtrail-log events.json
```

### Impact Analysis

```bash
# Analyze impact of drift (requires SkyGraph)
deepdrift impact --drift-file drift.json --graph skygraph.json
```

---

## Configuration

```yaml
# deepdrift.yaml
terraform:
  state_file: ./terraform.tfstate
  backend: s3  # or local, remote

provider:
  name: aws
  region: us-east-1

cloudtrail:
  enabled: true
  bucket: my-cloudtrail-bucket
  prefix: AWSLogs/

storage:
  type: tidb
  dsn: "root@tcp(localhost:4000)/airdig"

monitoring:
  interval: 5m  # Check every 5 minutes
  alerts:
    slack: https://hooks.slack.com/...
```

Run with config:

```bash
deepdrift detect --config deepdrift.yaml
```

---

## Integration with TFDrift

DeepDrift is built on top of the existing **TFDrift** project. It extends TFDrift with:

1. **CloudTrail correlation:** Link drift to audit logs
2. **Graph integration:** Use SkyGraph for impact analysis
3. **Event storage:** Persist drift events in TiDB
4. **Real-time monitoring:** Continuous drift detection

**Migration from TFDrift:**

If you're already using TFDrift, DeepDrift is a drop-in replacement with additional features. Simply replace `tfdrift` with `deepdrift` in your commands.

---

## Example: Drift Detection Flow

### 1. Detect Drift

```bash
$ deepdrift detect --state terraform.tfstate --provider aws
```

**Output:**

```
Found 3 drift events:

1. [MODIFIED] aws_security_group.web (sg-123456)
   - ingress rule added: 0.0.0.0/0:22
   - Detected at: 2024-01-15 14:32:18 UTC

2. [DELETED] aws_instance.worker (i-789012)
   - Instance terminated outside Terraform
   - Detected at: 2024-01-15 14:35:42 UTC

3. [CREATED] aws_s3_bucket.logs (my-logs-bucket)
   - Bucket created manually
   - Detected at: 2024-01-15 14:40:11 UTC
```

### 2. Correlate with CloudTrail

```bash
$ deepdrift correlate --drift-file drift.json
```

**Output:**

```
Correlation results:

1. sg-123456 (security group modified)
   - CloudTrail Event: AuthorizeSecurityGroupIngress
   - User: john.doe@example.com
   - Time: 2024-01-15 14:32:17 UTC
   - Source IP: 203.0.113.5

2. i-789012 (instance terminated)
   - CloudTrail Event: TerminateInstances
   - User: automation-role
   - Time: 2024-01-15 14:35:41 UTC

3. my-logs-bucket (bucket created)
   - CloudTrail Event: CreateBucket
   - User: admin@example.com
   - Time: 2024-01-15 14:40:10 UTC
```

### 3. Impact Analysis

```bash
$ deepdrift impact --drift-file drift.json --graph skygraph.json
```

**Output:**

```
Impact analysis:

1. sg-123456 (security group modified)
   - Affects: 5 EC2 instances
   - Risk: HIGH (port 22 opened to 0.0.0.0/0)
   - Recommendation: Restrict SSH access to specific IPs

2. i-789012 (instance terminated)
   - Affects: 1 load balancer (target group)
   - Risk: MEDIUM (reduced capacity)

3. my-logs-bucket (bucket created)
   - No dependencies found
   - Risk: LOW
```

---

## Development

### Prerequisites

- Go 1.21+
- Terraform
- AWS credentials configured

### Build

```bash
cd deepdrift
go build -o bin/deepdrift ./cmd/deepdrift
```

### Test

```bash
go test ./...
```

---

## Roadmap

### v0.1.0 (Current)
- [ ] Terraform state parser
- [ ] AWS drift detection
- [ ] Diff engine
- [ ] JSON export
- [ ] CLI tool

### v0.2.0
- [ ] CloudTrail correlation
- [ ] Impact analysis (SkyGraph integration)
- [ ] Event storage (TiDB)
- [ ] Continuous monitoring mode

### v0.3.0
- [ ] GCP/Azure support
- [ ] Slack/PagerDuty alerting
- [ ] GraphQL query API
- [ ] Web UI integration

---

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for development guidelines.

---

## License

Apache License 2.0 — see [LICENSE](../LICENSE)

---

**DeepDrift** is part of the [AirDig](https://github.com/yourusername/airdig) project.
