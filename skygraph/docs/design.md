# SkyGraph MVP 設計書

## 概要

SkyGraph は AirDig の基盤となるクラウド構成グラフエンジンです。
AWS API からリソース情報を取得し、依存関係を解析してグラフとして可視化します。

---

## MVP（v0.1.0）の目標

### 機能スコープ

**実装する機能：**
- ✅ AWS リソーススキャン（EC2, VPC, Subnet, Security Group, RDS）
- ✅ グラフモデル構築（ノード + エッジ）
- ✅ 依存関係の自動検出
- ✅ JSON 形式でのエクスポート
- ✅ CLI ツール（`skygraph scan`）

**将来のバージョンで実装：**
- ⏳ Kubernetes スキャン（v0.2.0）
- ⏳ GCP/Azure スキャン（v0.3.0）
- ⏳ TiDB ストレージ（v0.2.0）
- ⏳ リアルタイム更新（v0.3.0）
- ⏳ GraphQL API（v0.3.0）

---

## アーキテクチャ

### システム構成

```
┌─────────────────────────────────────────────┐
│         SkyGraph CLI                        │
│  (skygraph scan --provider aws)             │
└─────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────┐
│         Scanner Orchestrator                │
│  - 認証情報管理                               │
│  - スキャン順序制御                           │
│  - 並列実行                                  │
└─────────────────────────────────────────────┘
        │           │           │
        ▼           ▼           ▼
┌──────────┐ ┌──────────┐ ┌──────────┐
│   EC2    │ │   VPC    │ │   RDS    │
│ Scanner  │ │ Scanner  │ │ Scanner  │
└──────────┘ └──────────┘ └──────────┘
        │           │           │
        └───────────┼───────────┘
                    ▼
        ┌───────────────────────┐
        │   Graph Builder       │
        │  - ノード生成           │
        │  - エッジ推論           │
        │  - 重複排除             │
        └───────────────────────┘
                    │
                    ▼
        ┌───────────────────────┐
        │   Storage Layer       │
        │  - JSON Export        │
        │  - (TiDB: 将来)       │
        └───────────────────────┘
```

---

## データフロー

### 1. スキャンフェーズ

```
AWS API → Scanner → RawResource → ResourceNode
```

**例：EC2 インスタンスのスキャン**

```go
// AWS API レスポンス
ec2Instance := &ec2.Instance{
    InstanceId: "i-123456",
    InstanceType: "t3.large",
    VpcId: "vpc-789012",
    SubnetId: "subnet-345678",
    SecurityGroups: []*ec2.GroupIdentifier{
        {GroupId: "sg-901234"},
    },
}

// ResourceNode に変換
node := ResourceNode{
    ID: "aws:ec2:i-123456",
    Type: "ec2",
    Provider: "aws",
    Region: "us-east-1",
    Metadata: map[string]interface{}{
        "instance_type": "t3.large",
        "state": "running",
        "vpc_id": "vpc-789012",
        "subnet_id": "subnet-345678",
    },
}
```

### 2. エッジ推論フェーズ

グラフビルダーは以下のルールでエッジを生成します：

| 関係 | From | To | Type | 理由 |
|------|------|-----|------|------|
| VPC → Subnet | VPC | Subnet | ownership | Subnet は VPC に所属 |
| Subnet → EC2 | Subnet | EC2 | network | EC2 は Subnet 内に配置 |
| SG → EC2 | SecurityGroup | EC2 | network | EC2 に SG が適用 |
| EC2 → RDS | EC2 | RDS | dependency | アプリが DB に接続 |

**エッジ推論例：**

```go
// EC2 の metadata に subnet_id がある場合
if subnetID, ok := ec2Node.Metadata["subnet_id"].(string); ok {
    // Subnet → EC2 のエッジを生成
    edge := Edge{
        From: "aws:subnet:" + subnetID,
        To: ec2Node.ID,
        Type: "network",
    }
}
```

---

## コンポーネント設計

### 1. Scanner インターフェース

全てのスキャナーは共通インターフェースを実装：

```go
type Scanner interface {
    // スキャンを実行してリソースノードを返す
    Scan(ctx context.Context) ([]ResourceNode, error)

    // スキャナーの名前を返す
    Name() string
}
```

### 2. AWS Scanner 実装

**ディレクトリ構成：**

```
pkg/
  aws/
    scanner.go          # AWS スキャナー本体
    ec2.go              # EC2 スキャナー
    vpc.go              # VPC スキャナー
    rds.go              # RDS スキャナー
    security_group.go   # SG スキャナー
```

**実装例：EC2 Scanner**

```go
type EC2Scanner struct {
    client *ec2.Client
    region string
}

func (s *EC2Scanner) Scan(ctx context.Context) ([]ResourceNode, error) {
    // DescribeInstances を呼び出し
    result, err := s.client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
    if err != nil {
        return nil, err
    }

    nodes := []ResourceNode{}
    for _, reservation := range result.Reservations {
        for _, instance := range reservation.Instances {
            node := s.convertToNode(instance)
            nodes = append(nodes, node)
        }
    }

    return nodes, nil
}
```

### 3. Graph Builder

**役割：**
- スキャナーから取得したノードを統合
- エッジを推論して生成
- 重複を排除

**実装：**

```go
type GraphBuilder struct {
    graph *Graph
}

func (b *GraphBuilder) AddNodes(nodes []ResourceNode) {
    for _, node := range nodes {
        // 重複チェック
        if b.graph.FindNode(node.ID) == nil {
            b.graph.AddNode(node)
        }
    }
}

func (b *GraphBuilder) InferEdges() {
    // 全ノードを走査してエッジを推論
    for _, node := range b.graph.Nodes {
        edges := b.inferEdgesForNode(node)
        for _, edge := range edges {
            b.graph.AddEdge(edge)
        }
    }
}

func (b *GraphBuilder) inferEdgesForNode(node ResourceNode) []Edge {
    edges := []Edge{}

    switch node.Type {
    case "ec2":
        // subnet_id があれば Subnet → EC2 エッジ
        if subnetID, ok := node.Metadata["subnet_id"].(string); ok {
            edges = append(edges, Edge{
                From: "aws:subnet:" + subnetID,
                To: node.ID,
                Type: "network",
            })
        }

        // security_groups があれば SG → EC2 エッジ
        if sgs, ok := node.Metadata["security_groups"].([]string); ok {
            for _, sg := range sgs {
                edges = append(edges, Edge{
                    From: "aws:sg:" + sg,
                    To: node.ID,
                    Type: "network",
                })
            }
        }

    case "subnet":
        // vpc_id があれば VPC → Subnet エッジ
        if vpcID, ok := node.Metadata["vpc_id"].(string); ok {
            edges = append(edges, Edge{
                From: "aws:vpc:" + vpcID,
                To: node.ID,
                Type: "ownership",
            })
        }
    }

    return edges
}
```

---

## CLI 設計

### コマンド構成

```bash
# 基本スキャン
skygraph scan --provider aws --region us-east-1

# 複数リージョン
skygraph scan --provider aws --region us-east-1,us-west-2

# 出力先指定
skygraph scan --provider aws --output graph.json

# 特定リソースタイプのみ
skygraph scan --provider aws --resources ec2,vpc

# 設定ファイル使用
skygraph scan --config skygraph.yaml
```

### 設定ファイル形式

```yaml
# skygraph.yaml
provider: aws

aws:
  regions:
    - us-east-1
    - us-west-2

  # スキャン対象リソース
  resources:
    - ec2
    - vpc
    - subnet
    - security_group
    - rds

  # AWS 認証情報（省略時は環境変数/~/.aws/credentials）
  profile: default

output:
  format: json
  path: ./output/graph.json
```

---

## エラーハンドリング

### 1. 認証エラー

```go
// AWS 認証失敗時
if err := scanner.Authenticate(); err != nil {
    return fmt.Errorf("AWS authentication failed: %w", err)
}
```

### 2. 部分的スキャン失敗

```go
// 一部のリソースがスキャンできなくても続行
for _, scanner := range scanners {
    nodes, err := scanner.Scan(ctx)
    if err != nil {
        log.Warnf("Scanner %s failed: %v", scanner.Name(), err)
        continue
    }
    builder.AddNodes(nodes)
}
```

### 3. API レート制限

```go
// AWS API レート制限対策（指数バックオフ）
retrier := retry.NewExponential(retry.WithMaxRetries(3))
err := retrier.Do(func() error {
    return client.DescribeInstances(ctx, input)
})
```

---

## テスト戦略

### 1. ユニットテスト

```go
// pkg/graph/model_test.go
func TestGraph_AddNode(t *testing.T) {
    g := NewGraph()
    node := ResourceNode{ID: "test-1", Type: "ec2"}

    g.AddNode(node)

    assert.Equal(t, 1, g.NodeCount())
    assert.NotNil(t, g.FindNode("test-1"))
}
```

### 2. インテグレーションテスト（モック使用）

```go
// pkg/aws/ec2_test.go
func TestEC2Scanner_Scan(t *testing.T) {
    // AWS SDK のモック
    mockClient := &mockEC2Client{
        instances: []*ec2.Instance{
            {InstanceId: aws.String("i-123")},
        },
    }

    scanner := &EC2Scanner{client: mockClient}
    nodes, err := scanner.Scan(context.Background())

    require.NoError(t, err)
    assert.Len(t, nodes, 1)
    assert.Equal(t, "aws:ec2:i-123", nodes[0].ID)
}
```

---

## パフォーマンス

### 並列スキャン

```go
// 複数のスキャナーを並列実行
var wg sync.WaitGroup
nodeChan := make(chan []ResourceNode, len(scanners))

for _, scanner := range scanners {
    wg.Add(1)
    go func(s Scanner) {
        defer wg.Done()
        nodes, err := s.Scan(ctx)
        if err == nil {
            nodeChan <- nodes
        }
    }(scanner)
}

wg.Wait()
close(nodeChan)

// 結果を集約
for nodes := range nodeChan {
    builder.AddNodes(nodes)
}
```

### メモリ使用量

- **想定リソース数：** 1,000〜10,000 ノード
- **メモリ使用量：** 約 10MB〜100MB（JSON）
- **大規模環境：** ストリーミング処理で対応（将来実装）

---

## セキュリティ

### 1. AWS 認証情報

```go
// 環境変数、~/.aws/credentials、IAM ロールの順で試行
cfg, err := config.LoadDefaultConfig(ctx,
    config.WithRegion(region),
    config.WithSharedConfigProfile(profile),
)
```

### 2. 読み取り専用権限

**必要な IAM ポリシー：**

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:Describe*",
        "rds:Describe*",
        "vpc:Describe*"
      ],
      "Resource": "*"
    }
  ]
}
```

---

## 今後の拡張

### v0.2.0

- Kubernetes スキャン
- TiDB ストレージバックエンド
- 増分スキャン（差分のみ取得）

### v0.3.0

- GCP/Azure スキャン
- リアルタイム更新（CloudWatch Events）
- GraphQL API

---

## まとめ

SkyGraph MVP は以下を実現します：

✅ AWS の主要リソースをスキャン
✅ グラフモデルで依存関係を表現
✅ JSON でエクスポート
✅ CLI ツールで簡単実行

これにより、DeepDrift / TraceCore / PulseSight が依存する**グラフ基盤**が完成します。
