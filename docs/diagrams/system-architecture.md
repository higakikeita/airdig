# AirDig システムアーキテクチャ図

## 全体システムアーキテクチャ

```mermaid
graph TB
    subgraph "AirDig UI Layer"
        UI[AirDig UI<br/>Next.js + Cytoscape.js]
        UI_Sky[SkyGraph View]
        UI_Drift[DeepDrift View]
        UI_Trace[TraceCore View]
        UI_Pulse[PulseSight View]
        UI_Unified[Unified View]

        UI --> UI_Sky
        UI --> UI_Drift
        UI --> UI_Trace
        UI --> UI_Pulse
        UI --> UI_Unified
    end

    subgraph "AirDig Engine"
        Engine[AirDig Engine<br/>GraphQL API]
        GraphDB[(Unified Graph<br/>TiDB/ClickHouse)]
        EventBus[Event Bus<br/>Kafka/NATS]

        Engine --> GraphDB
        Engine --> EventBus
    end

    subgraph "Four Pillars"
        subgraph "SkyGraph"
            Sky[SkyGraph Scanner]
            SkyAWS[AWS Scanner]
            SkyK8s[K8s Scanner]
            SkyGCP[GCP Scanner]

            Sky --> SkyAWS
            Sky --> SkyK8s
            Sky --> SkyGCP
        end

        subgraph "DeepDrift"
            Drift[DeepDrift Engine]
            DriftTF[Terraform Parser]
            DriftCT[CloudTrail Correlator]

            Drift --> DriftTF
            Drift --> DriftCT
        end

        subgraph "TraceCore"
            Trace[TraceCore Processor]
            TraceOTel[OTel Collector]
            TraceTempo[(Tempo<br/>S3)]

            Trace --> TraceOTel
            Trace --> TraceTempo
        end

        subgraph "PulseSight"
            Pulse[PulseSight Processor]
            PulseProm[Prometheus]
            PulseLoki[Loki]
            PulseFalco[Falco]

            Pulse --> PulseProm
            Pulse --> PulseLoki
            Pulse --> PulseFalco
        end
    end

    subgraph "Data Sources"
        AWS[AWS API]
        K8s[Kubernetes API]
        TF[Terraform State]
        CT[CloudTrail]
        Apps[Applications<br/>OpenTelemetry]
        Metrics[Metrics Exporters]
        Logs[Log Sources]
        Runtime[Runtime Events<br/>eBPF/Falco]
    end

    UI -.GraphQL/WebSocket.-> Engine

    Sky --> Engine
    Drift --> Engine
    Trace --> Engine
    Pulse --> Engine

    EventBus -.-> Sky
    EventBus -.-> Drift
    EventBus -.-> Trace
    EventBus -.-> Pulse

    SkyAWS -.-> AWS
    SkyK8s -.-> K8s
    DriftTF -.-> TF
    DriftCT -.-> CT
    TraceOTel -.-> Apps
    PulseProm -.-> Metrics
    PulseLoki -.-> Logs
    PulseFalco -.-> Runtime

    style UI fill:#e1f5ff
    style Engine fill:#fff3e0
    style Sky fill:#e8f5e9
    style Drift fill:#f3e5f5
    style Trace fill:#fce4ec
    style Pulse fill:#fff9c4
```

## コンポーネント説明

### UI Layer
- **AirDig UI**: Next.js ベースの Web UI
- **5つのビュー**: 各 Pillar ごとのビュー + 統合ビュー

### AirDig Engine
- **GraphQL API**: クライアントからのクエリを処理
- **Unified Graph DB**: 全データを統合したグラフデータベース
- **Event Bus**: 非同期イベント配信

### Four Pillars
- **SkyGraph**: クラウドリソース構成をスキャン
- **DeepDrift**: IaC と実態の差分を検出
- **TraceCore**: アプリケーショントレースを収集
- **PulseSight**: メトリクス、ログ、ランタイムイベントを収集

### Data Sources
- クラウド API、Kubernetes、Terraform、CloudTrail
- OpenTelemetry instrumented アプリ
- Prometheus exporters、Loki logs、Falco events
