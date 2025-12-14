# DeepDrift

**Intelligent Terraform Drift Detection with Impact Analysis**

DeepDrift is a powerful drift detection and impact analysis tool that integrates with [TFDrift-Falco](https://github.com/higakikeita/tfdrift-falco) and [SkyGraph](../skygraph) to provide comprehensive infrastructure drift intelligence.

## Features

- **🔍 Real-time Drift Detection**: Integrates with TFDrift-Falco for CloudTrail-based drift detection
- **📊 Impact Analysis**: Uses SkyGraph to analyze the blast radius of infrastructure changes
- **🎯 Root Cause Analysis**: Traces drift back to CloudTrail events and IAM principals
- **🔐 Security-Aware**: Automatically escalates severity for security-related resource changes
- **⚡ Continuous Monitoring**: Watch mode for ongoing drift detection
- **📈 Graph-based Analysis**: BFS traversal to find affected resources (up to 3 hops)

## Architecture

```
┌─────────────────┐
│  Terraform      │
│  State File     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐      ┌──────────────┐
│  TFDrift-Falco  │◄─────│  CloudTrail  │
│  (External)     │      │  Events      │
└────────┬────────┘      └──────────────┘
         │
         │ JSON Output
         ▼
┌─────────────────┐
│  DeepDrift      │
│  Adapter        │
└────────┬────────┘
         │
         │ DriftEvent
         ▼
┌─────────────────┐      ┌──────────────┐
│  Impact         │◄─────│  SkyGraph    │
│  Analyzer       │      │  Graph Data  │
└────────┬────────┘      └──────────────┘
         │
         │ ImpactAnalysisResult
         ▼
┌─────────────────┐
│  CLI Output     │
│  / JSON Export  │
└─────────────────┘
```

## Installation

### Prerequisites

1. **TFDrift-Falco** must be installed and configured:
   ```bash
   # Default location: ~/tfdrift-falco/bin/tfdrift
   # Or specify custom path with --tfdrift flag
   ```

2. **Go 1.21+** for building from source

### Build

```bash
# Clone the repository
git clone https://github.com/higakikeita/airdig.git
cd airdig/deepdrift

# Build
make build

# Or install to $GOPATH/bin
make install
```

## Usage

### 1. Detect Drift

Detect infrastructure drift using TFDrift-Falco:

```bash
deepdrift --command detect --state terraform.tfstate

# With custom TFDrift path
deepdrift --command detect \
  --state terraform.tfstate \
  --tfdrift /path/to/tfdrift

# Save output to JSON
deepdrift --command detect \
  --state terraform.tfstate \
  --output drift-events.json
```

**Example Output:**
```
==============================================
  DeepDrift - Drift Detection & Impact Analysis
  Version: 0.1.0 (Alpha)
==============================================

Running drift detection...
State file: terraform.tfstate

Found 3 drift events

1. [modified] aws:ec2:i-123456 (ec2)
   Severity: medium
   User: alice@example.com
   Event: ModifyInstanceAttribute

2. [deleted] aws:sg:sg-789012 (security_group)
   Severity: critical
   User: bob@example.com
   Event: DeleteSecurityGroup

3. [created] aws:s3:my-bucket (s3)
   Severity: low
   User: charlie@example.com
   Event: CreateBucket
```

### 2. Impact Analysis

Analyze the impact of drift using SkyGraph:

```bash
# Generate SkyGraph first
cd ../skygraph
./bin/skygraph scan --output graph.json

# Run impact analysis
cd ../deepdrift
deepdrift --command impact \
  --state terraform.tfstate \
  --graph ../skygraph/graph.json

# Save results
deepdrift --command impact \
  --state terraform.tfstate \
  --graph ../skygraph/graph.json \
  --output impact-results.json
```

**Example Output:**
```
Running impact analysis...
Graph file: graph.json

Loaded graph: 42 nodes, 67 edges

Analyzed 3 drift events

1. [modified] aws:ec2:i-123456 (ec2)
   Severity: medium
   Affected resources: 5
   Blast radius: 2 hops
   Recommendations:
     • Verify configuration changes against security policies
     • Apply changes to Terraform code or revert via terraform apply

2. [deleted] aws:sg:sg-789012 (security_group)
   Severity: critical
   Affected resources: 12
   Blast radius: 3 hops
   Recommendations:
     • Review if this resource deletion was intentional
     • Check 12 dependent resources for potential issues
     • Update Terraform state if deletion is permanent
```

### 3. Continuous Monitoring

Run continuous drift monitoring:

```bash
deepdrift --command watch \
  --state terraform.tfstate \
  --interval 5m

# Monitor every 30 seconds
deepdrift --command watch \
  --state terraform.tfstate \
  --interval 30s
```

**Example Output:**
```
Starting continuous drift monitoring...
Interval: 5m0s

[2025-12-14T18:00:00+09:00] Detected 0 drift events

[2025-12-14T18:05:00+09:00] Detected 1 drift events
  - [modified] aws:ec2:i-123456 (ec2)

[2025-12-14T18:10:00+09:00] Detected 0 drift events
```

## Configuration

### Command-Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--command` | Command to run: detect, impact, watch | `detect` |
| `--state` | Terraform state file path | `terraform.tfstate` |
| `--graph` | SkyGraph JSON file path (required for impact) | - |
| `--tfdrift` | TFDrift binary path | `~/tfdrift-falco/bin/tfdrift` |
| `--config` | TFDrift config file path | - |
| `--output` | Output file path | stdout |
| `--interval` | Watch interval for continuous monitoring | `5m` |

### TFDrift Configuration

DeepDrift uses TFDrift-Falco's configuration. Create a config file:

```yaml
# tfdrift-config.yaml
cloudtrail:
  region: us-west-2
  lookback_minutes: 60

filters:
  resource_types:
    - ec2
    - security_group
    - vpc
    - rds
    - s3
    - lambda
```

Then run:
```bash
deepdrift --command detect \
  --state terraform.tfstate \
  --config tfdrift-config.yaml
```

## Data Model

### DriftEvent

```go
type DriftEvent struct {
    ID                string                 // Unique event ID
    ResourceID        string                 // e.g., "aws:ec2:i-123456"
    ResourceType      string                 // e.g., "ec2", "security_group"
    Type              DriftType              // created, modified, deleted
    Timestamp         time.Time              // When drift was detected
    Before            map[string]interface{} // State before change
    After             map[string]interface{} // State after change
    Diff              map[string]interface{} // Detailed diff
    RootCause         *RootCause             // CloudTrail event info
    ImpactedResources []string               // List of affected resources
    Severity          Severity               // low, medium, high, critical
}
```

### ImpactAnalysisResult

```go
type ImpactAnalysisResult struct {
    DriftEventID          string             // Original drift event ID
    AffectedResourceCount int                // Number of affected resources
    AffectedResources     []AffectedResource // Detailed impact information
    BlastRadius           int                // Maximum graph distance (hops)
    Recommendations       []string           // Suggested actions
    Severity              Severity           // Overall severity
}
```

## Impact Analysis Algorithm

DeepDrift uses **Breadth-First Search (BFS)** to traverse the SkyGraph and find affected resources:

1. **Start Node**: The resource with drift
2. **Traversal**: BFS up to 3 hops from the start node
3. **Relationship Types**:
   - `network`: Network connectivity (VPC, Subnet, Security Group)
   - `dependency`: Service dependencies (EC2 → RDS, Lambda → DynamoDB)
   - `ownership`: Parent-child relationships (VPC → Subnet)

4. **Severity Calculation**:
   - Base severity from drift type and resource type
   - Escalated if >10 resources affected
   - Escalated if security resources affected (IAM, KMS, Security Groups)

## Examples

### Example 1: Detect and Analyze Deleted Security Group

```bash
# 1. Detect drift
deepdrift --command detect --state terraform.tfstate --output drift.json

# 2. Analyze impact
deepdrift --command impact \
  --state terraform.tfstate \
  --graph graph.json \
  --output impact.json

# 3. Review results
cat impact.json | jq '.[] | select(.severity == "critical")'
```

### Example 2: Continuous Monitoring with Alerts

```bash
# Monitor every minute and save to log
deepdrift --command watch \
  --state terraform.tfstate \
  --interval 1m \
  >> drift-monitor.log 2>&1 &

# Monitor the log
tail -f drift-monitor.log
```

## Integration with CI/CD

### GitHub Actions

```yaml
name: Drift Detection
on:
  schedule:
    - cron: '0 */6 * * *'  # Every 6 hours

jobs:
  detect-drift:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install TFDrift
        run: |
          # Install TFDrift-Falco
          wget https://github.com/higakikeita/tfdrift-falco/releases/download/v1.0.0/tfdrift-linux-amd64
          chmod +x tfdrift-linux-amd64
          sudo mv tfdrift-linux-amd64 /usr/local/bin/tfdrift

      - name: Install DeepDrift
        run: |
          cd deepdrift
          make install

      - name: Run Drift Detection
        run: |
          deepdrift --command detect \
            --state terraform.tfstate \
            --output drift-report.json

      - name: Upload Results
        uses: actions/upload-artifact@v3
        with:
          name: drift-report
          path: drift-report.json
```

## Development

### Project Structure

```
deepdrift/
├── cmd/
│   └── deepdrift/          # CLI entry point
│       └── main.go
├── pkg/
│   ├── types/              # Core data types
│   │   └── drift.go
│   ├── tfdrift/            # TFDrift adapter
│   │   └── adapter.go
│   └── impact/             # Impact analysis engine
│       └── analyzer.go
├── go.mod
├── Makefile
└── README.md
```

### Run Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Format code
make fmt

# Run linter
make lint
```

### Add New Features

1. **Custom Severity Rules**: Extend `calculateSeverity()` in `pkg/tfdrift/adapter.go`
2. **New Recommendations**: Modify `generateRecommendations()` in `pkg/impact/analyzer.go`
3. **Custom Graph Traversal**: Adjust `findAffectedResources()` in `pkg/impact/analyzer.go`

## Roadmap

### v0.1.0 (Current - Alpha)
- ✅ TFDrift-Falco adapter
- ✅ SkyGraph integration for impact analysis
- ✅ CLI with detect/impact/watch commands
- ✅ CloudTrail root cause analysis
- ✅ BFS-based impact traversal

### v0.2.0 (Q1 2025)
- [ ] Support for Azure and GCP
- [ ] Persistent storage (TiDB/ClickHouse)
- [ ] Web UI for visualization
- [ ] Slack/Teams notifications
- [ ] Custom policy engine

### v0.3.0 (Q2 2025)
- [ ] Machine learning-based anomaly detection
- [ ] Drift prediction
- [ ] Auto-remediation workflows
- [ ] Integration with Terraform Cloud

## Integration with TFDrift

DeepDrift is built on top of the existing **TFDrift-Falco** project. It extends TFDrift with:

1. **CloudTrail correlation**: Link drift to audit logs (via TFDrift)
2. **Graph integration**: Use SkyGraph for impact analysis
3. **Severity calculation**: Security-aware risk assessment
4. **Real-time monitoring**: Continuous drift detection

**TFDrift-Falco** handles the core drift detection logic, while **DeepDrift** adds the intelligence layer on top.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](../CONTRIBUTING.md) for details.

## License

Apache License 2.0 - See [LICENSE](../LICENSE) for details.

## Related Projects

- **[SkyGraph](../skygraph)**: Cloud resource graph builder
- **[TFDrift-Falco](https://github.com/higakikeita/tfdrift-falco)**: CloudTrail-based drift detection
- **[Airdig](https://github.com/higakikeita/airdig)**: Next-generation cloud observability platform

## Support

- GitHub Issues: https://github.com/higakikeita/airdig/issues
- Documentation: https://github.com/higakikeita/airdig/wiki
- Twitter: @higakikeita
