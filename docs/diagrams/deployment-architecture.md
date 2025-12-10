# AirDig デプロイメントアーキテクチャ

## Kubernetes デプロイメント

```mermaid
graph TB
    subgraph "Kubernetes Cluster"
        subgraph "Ingress Layer"
            Ingress[Ingress<br/>NGINX/Traefik]
        end

        subgraph "UI Tier"
            UI1[AirDig UI<br/>Pod 1]
            UI2[AirDig UI<br/>Pod 2]
            UISvc[UI Service<br/>ClusterIP]
        end

        subgraph "API Tier"
            Engine1[AirDig Engine<br/>Pod 1]
            Engine2[AirDig Engine<br/>Pod 2]
            Engine3[AirDig Engine<br/>Pod 3]
            EngineSvc[Engine Service<br/>ClusterIP]
        end

        subgraph "Pillar Services"
            Sky[SkyGraph<br/>CronJob]
            Drift[DeepDrift<br/>Deployment]
            Trace[TraceCore<br/>Deployment]
            Pulse[PulseSight<br/>Deployment]
        end

        subgraph "Data Platform"
            subgraph "TiDB Cluster"
                TiDB1[TiDB Server 1]
                TiDB2[TiDB Server 2]
                PD1[PD Server 1]
                PD2[PD Server 2]
                TiKV1[TiKV 1]
                TiKV2[TiKV 2]
                TiKV3[TiKV 3]
            end

            Tempo[Tempo<br/>StatefulSet]
            Mimir[Mimir<br/>StatefulSet]
            Loki[Loki<br/>StatefulSet]
            Redis[Redis<br/>StatefulSet]
            Kafka[Kafka<br/>StatefulSet]
        end

        subgraph "Storage"
            S3[S3/MinIO<br/>Trace Storage]
            PV1[PersistentVolume<br/>TiKV Data]
            PV2[PersistentVolume<br/>Metrics]
            PV3[PersistentVolume<br/>Logs]
        end
    end

    subgraph "External Access"
        User[User Browser]
        API[External API Client]
    end

    subgraph "External Services"
        AWS[AWS API]
        K8sAPI[Kubernetes API]
        CloudTrail[CloudTrail]
    end

    User --> Ingress
    API --> Ingress

    Ingress --> UISvc
    Ingress --> EngineSvc

    UISvc --> UI1
    UISvc --> UI2

    EngineSvc --> Engine1
    EngineSvc --> Engine2
    EngineSvc --> Engine3

    Engine1 --> TiDB1
    Engine2 --> TiDB2
    Engine3 --> TiDB1

    Engine1 --> Redis
    Engine2 --> Redis
    Engine3 --> Redis

    Engine1 --> Kafka
    Engine2 --> Kafka
    Engine3 --> Kafka

    Sky --> Kafka
    Drift --> Kafka
    Trace --> Kafka
    Pulse --> Kafka

    Kafka --> Engine1
    Kafka --> Engine2
    Kafka --> Engine3

    Sky -.-> AWS
    Sky -.-> K8sAPI
    Drift -.-> AWS
    Drift -.-> CloudTrail

    Trace --> Tempo
    Tempo --> S3

    Pulse --> Mimir
    Pulse --> Loki

    TiDB1 --> PD1
    TiDB2 --> PD2
    PD1 --> TiKV1
    PD1 --> TiKV2
    PD1 --> TiKV3

    TiKV1 --> PV1
    TiKV2 --> PV1
    TiKV3 --> PV1

    Mimir --> PV2
    Loki --> PV3

    style UI1 fill:#e1f5ff
    style UI2 fill:#e1f5ff
    style Engine1 fill:#fff3e0
    style Engine2 fill:#fff3e0
    style Engine3 fill:#fff3e0
    style Sky fill:#e8f5e9
    style Drift fill:#f3e5f5
    style Trace fill:#fce4ec
    style Pulse fill:#fff9c4
```

## ネットワークフロー

```mermaid
graph LR
    subgraph "External"
        Internet[Internet]
    end

    subgraph "Load Balancer"
        LB[Load Balancer<br/>AWS ALB/ELB]
    end

    subgraph "Kubernetes"
        subgraph "Ingress"
            Ingress[Ingress Controller]
        end

        subgraph "Services"
            UISvc[UI Service<br/>:3000]
            EngineSvc[Engine Service<br/>:8080]
        end

        subgraph "Pods"
            UI[UI Pods]
            Engine[Engine Pods]
        end
    end

    subgraph "Data Layer"
        TiDB[(TiDB<br/>:4000)]
        Redis[(Redis<br/>:6379)]
        Kafka[(Kafka<br/>:9092)]
    end

    Internet -->|HTTPS| LB
    LB -->|HTTP| Ingress

    Ingress -->|/| UISvc
    Ingress -->|/api/*| EngineSvc
    Ingress -->|/graphql| EngineSvc

    UISvc --> UI
    EngineSvc --> Engine

    Engine --> TiDB
    Engine --> Redis
    Engine --> Kafka

    UI -.GraphQL.-> Engine

    style Internet fill:#ffebee
    style LB fill:#e3f2fd
    style Ingress fill:#f3e5f5
    style Engine fill:#fff3e0
    style TiDB fill:#e0f7fa
```

## Docker Compose 構成（開発環境）

```mermaid
graph TB
    subgraph "Docker Compose"
        subgraph "Application Services"
            UI[airdig-ui<br/>:3000]
            Engine[airdig-engine<br/>:8080]
            Sky[skygraph<br/>CronJob]
            Drift[deepdrift<br/>Daemon]
        end

        subgraph "Data Services"
            TiDB[tidb<br/>:4000]
            PD[pd<br/>:2379]
            TiKV[tikv<br/>:20160]
            Tempo[tempo<br/>:3200]
            Prometheus[prometheus<br/>:9090]
            Loki[loki<br/>:3100]
            Grafana[grafana<br/>:3001]
        end

        subgraph "Message Queue"
            Kafka[kafka<br/>:9092]
            Zookeeper[zookeeper<br/>:2181]
        end

        subgraph "Cache & Storage"
            Redis[redis<br/>:6379]
            MinIO[minio<br/>:9000]
        end
    end

    subgraph "External Access"
        Browser[Browser]
    end

    Browser -->|http://localhost:3000| UI
    Browser -->|http://localhost:3001| Grafana

    UI --> Engine
    Engine --> TiDB
    Engine --> Redis
    Engine --> Kafka

    Sky --> Kafka
    Drift --> Kafka

    Tempo --> MinIO
    Prometheus --> MinIO
    Loki --> MinIO

    TiDB --> PD
    PD --> TiKV

    Kafka --> Zookeeper

    Grafana --> Prometheus
    Grafana --> Loki
    Grafana --> Tempo

    style UI fill:#e1f5ff
    style Engine fill:#fff3e0
    style Sky fill:#e8f5e9
    style Drift fill:#f3e5f5
```

## スケーリング戦略

```mermaid
graph TD
    subgraph "Stateless Services (水平スケール)"
        UI[AirDig UI<br/>HPA: 2-10 replicas]
        Engine[AirDig Engine<br/>HPA: 3-20 replicas]
        Sky[SkyGraph<br/>Parallelism: 1-5]
    end

    subgraph "Stateful Services (垂直スケール + シャーディング)"
        TiDB[TiDB Cluster<br/>3-10 nodes]
        Kafka[Kafka Cluster<br/>3-7 brokers]
        Redis[Redis<br/>1 primary + 2 replicas]
    end

    subgraph "Storage Services"
        Tempo[Tempo<br/>StatefulSet + S3]
        Mimir[Mimir<br/>StatefulSet + S3]
        Loki[Loki<br/>StatefulSet + S3]
    end

    subgraph "Auto-Scaling Triggers"
        CPU[CPU > 70%]
        Memory[Memory > 80%]
        Queue[Kafka Lag > 1000]
    end

    CPU -.-> UI
    CPU -.-> Engine
    Memory -.-> Engine
    Queue -.-> Engine

    UI --> LB1[Load Balancer]
    Engine --> LB2[Service Mesh<br/>Istio/Linkerd]

    LB2 --> TiDB
    LB2 --> Kafka
    LB2 --> Redis

    style UI fill:#e1f5ff
    style Engine fill:#fff3e0
    style Sky fill:#e8f5e9
```

## 高可用性構成

```mermaid
graph TB
    subgraph "Region A (Primary)"
        subgraph "AZ-1a"
            UI1a[UI Pod]
            Engine1a[Engine Pod]
            TiDB1a[TiDB Node]
        end

        subgraph "AZ-1b"
            UI1b[UI Pod]
            Engine1b[Engine Pod]
            TiDB1b[TiDB Node]
        end

        subgraph "AZ-1c"
            UI1c[UI Pod]
            Engine1c[Engine Pod]
            TiDB1c[TiDB Node]
        end

        LB1[Load Balancer<br/>Multi-AZ]
    end

    subgraph "Region B (DR)"
        subgraph "AZ-2a"
            UI2a[UI Pod<br/>Standby]
            Engine2a[Engine Pod<br/>Standby]
            TiDB2a[TiDB Node<br/>Replica]
        end
    end

    subgraph "Global Services"
        DNS[Route53<br/>Health Check]
        S3Global[S3 Cross-Region<br/>Replication]
    end

    DNS --> LB1
    DNS -.Failover.-> Engine2a

    LB1 --> UI1a
    LB1 --> UI1b
    LB1 --> UI1c

    UI1a --> Engine1a
    UI1b --> Engine1b
    UI1c --> Engine1c

    Engine1a --> TiDB1a
    Engine1b --> TiDB1b
    Engine1c --> TiDB1c

    TiDB1a -.Replication.-> TiDB2a
    TiDB1b -.Replication.-> TiDB2a
    TiDB1c -.Replication.-> TiDB2a

    Engine1a --> S3Global
    Engine2a -.Standby.-> S3Global

    style UI1a fill:#90ee90
    style UI1b fill:#90ee90
    style UI1c fill:#90ee90
    style UI2a fill:#ffcccb
    style Engine2a fill:#ffcccb
    style TiDB2a fill:#ffcccb
```
