# AirDig — クラウドを掘る。すべてを視る。

![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)
![Status](https://img.shields.io/badge/status-alpha-orange.svg)

[English](./README.md) | 日本語

**AirDig** は、クラウドインフラストラクチャ全体を空から深層まで可視化する次世代のクラウド観測・ドリフトインテリジェンスプラットフォームです。

**Sysdig** の深い可視性と **Falco** のランタイムセキュリティからインスピレーションを得て、AirDig は OS レイヤーを超えて、クラウド、アプリケーション、インフラストラクチャレイヤーまで観測性を拡張します。

---

## 🌟 AirDig とは？

AirDig は4つの重要な観測性の柱を、単一のグラフベースプラットフォームに統合します：

- **SkyGraph** — クラウドトポロジーとインフラ依存関係の可視化
- **DeepDrift** — リアルタイムドリフト検出と変更インテリジェンス
- **TraceCore** — 分散トレーシングとアプリケーションパフォーマンス監視
- **PulseSight** — メトリクス、ログ、ランタイムセキュリティイベント

これらの柱が一体となって、以下を組み合わせた **クラウド環境の360°ビュー** を提供します：
- クラウドリソース構成（AWS、GCP、Azure、Kubernetes）
- Infrastructure-as-Code（Terraform、CloudFormation）の状態とドリフト
- アプリケーショントレース（OpenTelemetry）
- ランタイムメトリクス（Prometheus）とセキュリティイベント（Falco、eBPF）

---

## 🧱 4つの柱

### 🟦 1. SkyGraph
**クラウドトポロジーと依存関係グラフの可視化**

- クラウドAPI（AWS、GCP、Azure、Kubernetes）をスキャン
- 統合されたリソースグラフを構築
- 依存関係とネットワークトポロジーを可視化
- IaC ツール（Terraform、CDK）と統合

### 🟢 2. DeepDrift
**インフラストラクチャドリフト検出と変更インテリジェンス**

- IaC の理想状態と実際のクラウド状態の差分を検出
- CloudTrail/監査ログと変更を相関付け
- 変更の影響分析を提供
- TFDrift エンジンと統合

### 🟣 3. TraceCore
**分散トレーシングと APM**

- OpenTelemetry トレースを取り込み
- サービスマップを生成
- アプリケーション動作とインフラ変更を相関付け
- Tempo/Jaeger にエクスポート

### 🟡 4. PulseSight
**メトリクス、ログ、ランタイム観測性**

- Prometheus メトリクスを取り込み
- ログを収集（Loki）
- ランタイムセキュリティイベント（Falco、eBPF）
- リソースの健全性ステータスを追跡

---

## 🎯 なぜ AirDig？

| 機能 | Datadog | Wiz | Sysdig | AirDig |
|------|---------|-----|--------|--------|
| クラウド構成グラフ | ❌ | ✅ | ❌ | ✅ |
| ドリフト検出 | ❌ | ❌ | ❌ | ✅ |
| APM / トレーシング | ✅ | ❌ | ✅ | ✅ |
| ランタイムセキュリティ | ⚠️ | ✅ | ✅ | ✅ |
| 統合グラフビュー | ❌ | ⚠️ | ❌ | ✅ |

AirDig は、クラウドトポロジー、ドリフトインテリジェンス、APM、ランタイム観測性を単一のグラフベースビューで統合する**唯一のプラットフォーム**です。

---

## 🚀 クイックスタート

### 前提条件

- Go 1.21+
- Docker & Docker Compose
- AWS/GCP/Azure 認証情報（クラウドスキャン用）
- Terraform（オプション、ドリフト検出用）

### インストール

```bash
# リポジトリをクローン
git clone https://github.com/higakikeita/airdig.git
cd airdig

# デモを実行
docker-compose up -d

# UI にアクセス
open http://localhost:3000
```

---

## 📚 ドキュメント

- [アーキテクチャ](./docs/architecture.md) — システム設計とデータフロー
- [ビジョン](./docs/vision.md) — プロジェクトの哲学と目標
- [ロードマップ](./docs/roadmap.md) — 開発計画とマイルストーン

### コンポーネントドキュメント

- [SkyGraph](./skygraph/README.md) — クラウドグラフエンジン
- [DeepDrift](./deepdrift/README.md) — ドリフト検出エンジン
- [TraceCore](./tracecore/README.md) — APM とトレーシング
- [PulseSight](./pulsesight/README.md) — メトリクスとランタイム観測性

---

## 🛠️ 開発ステータス

| コンポーネント | ステータス | バージョン |
|-----------|--------|---------|
| SkyGraph | 🟡 アルファ | v0.1.0 |
| DeepDrift | 🟢 ベータ | v0.5.0 |
| TraceCore | 🔴 計画中 | - |
| PulseSight | 🔴 計画中 | - |
| AirDig Engine | 🔴 計画中 | - |
| UI | 🔴 計画中 | - |

---

## 🤝 コントリビューション

貢献を歓迎します！詳細は [CONTRIBUTING.md](./CONTRIBUTING.md) をご覧ください。

---

## 📝 ライセンス

AirDig は [Apache License 2.0](./LICENSE) の下でライセンスされています。

---

## 🙏 謝辞

AirDig は巨人の肩の上に立っています：

- **Sysdig** — 深いシステム可視性の先駆者
- **Falco** — ランタイムセキュリティのイノベーション
- **Stratoshark** — クラウド API 観測性
- **OpenTelemetry** — 分散トレーシング標準
- **Terraform** — Infrastructure-as-Code

---

## 🔗 リンク

- [ドキュメント](./docs/)
- [GitHub Issues](https://github.com/higakikeita/airdig/issues)
- [Discussions](https://github.com/higakikeita/airdig/discussions)

---

## 📖 使用例

### SkyGraph でクラウドをスキャン

```bash
cd skygraph

# デモを実行（サンプルグラフを生成）
make demo

# AWS の実際のインフラをスキャン
make build
./bin/skygraph --provider aws --region us-east-1
```

**出力例：**

```
==============================================
  SkyGraph - Cloud Topology Scanner
  Version: 0.1.0 MVP
==============================================

Provider: aws
Region: us-east-1

Scanning AWS resources...
  - VPC
  - Subnet
  - Security Group
  - EC2 Instances
  - RDS Instances

Building graph...

==============================================
  Scan Results
==============================================
Duration: 3.45 seconds
Nodes: 23
Edges: 31

Resources by type:
  - vpc: 2
  - subnet: 4
  - security_group: 5
  - ec2: 8
  - rds: 4

✅ Done!
Graph saved to: graph.json
```

詳細は [SkyGraph の使用例](./skygraph/EXAMPLE.md) をご覧ください。

---

## 🎨 アーキテクチャ概要

```
┌─────────────────────────────────────────────────────────────┐
│                      AirDig UI (Next.js)                     │
│  ┌──────────┬──────────┬──────────┬──────────┬──────────┐  │
│  │ SkyGraph │DeepDrift │TraceCore │PulseSight│ Unified  │  │
│  │   View   │   View   │   View   │   View   │   View   │  │
│  └──────────┴──────────┴──────────┴──────────┴──────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                  AirDig Engine (GraphQL API)                 │
│                    Unified Graph Model                       │
└─────────────────────────────────────────────────────────────┘
        │              │              │              │
        ▼              ▼              ▼              ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│  SkyGraph   │ │  DeepDrift  │ │ TraceCore   │ │ PulseSight  │
│             │ │             │ │             │ │             │
│ Cloud Topo  │ │ Drift Detect│ │  APM/Trace  │ │Metrics/Logs │
│   Engine    │ │   Engine    │ │   Engine    │ │   Engine    │
└─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘
```

詳細は[アーキテクチャドキュメント](./docs/architecture.md)をご覧ください。

---

## 🌍 なぜ「AirDig」？

**Sysdig** がシステムコールとコンテナ動作を深く掘り下げるように、**AirDig** はクラウドインフラストラクチャとアプリケーション動作を深く掘り下げます。

- **Air**（空）— 高高度から全体を俯瞰
- **Dig**（掘る）— 深層まで詳細に調査

**空から深層まで、すべてを視る。**

---

## 🔮 将来の機能

### フェーズ2：インテリジェンス
- AI 駆動のルートコーズ分析
- 予測的影響分析（変更適用前）
- グラフパターンの異常検出
- 自動修復の提案

### フェーズ3：プラットフォーム
- マルチテナンシー（チーム単位のグラフ）
- GitOps 統合（IaC リポジトリと同期）
- カスタムプラグインマーケットプレイス
- コスト最適化（支出をグラフにマッピング）

### フェーズ4：エコシステム
- SaaS オファリング
- エンタープライズ機能（SSO、監査ログ、コンプライアンス）
- コミュニティ駆動のスキャナーライブラリ
- 統合マーケットプレイス

---

## 💡 ユースケース

### 「なぜアプリが遅い？」

**AirDig を使う前：**
1. Datadog APM でレイテンシスパイクを確認
2. AWS コンソールで何百ものリソースをスクロール
3. Terraform state を手動で diff
4. CloudTrail で JSON ログを精査
5. 最終的に発見：10分前に誰かが RDS インスタンスタイプを変更

**AirDig を使うと：**
1. AirDig UI を開く
2. RDS インスタンスの赤いノードを確認
3. クリック → ドリフト注釈を表示：「RDS インスタンスタイプが変更されました」
4. 相関トレースでクエリレイテンシの増加を表示
5. CloudTrail イベントを表示：「user@example.com がインスタンスタイプを変更」
6. **合計時間：30秒**

---

**AirDig — 空から深層まで。クラウド全体を可視化。**
