# SkyGraph 使用例

## 1. デモの実行

デモ版では、サンプルのAWSインフラストラクチャグラフを生成します。

```bash
# ビルド
make build-demo

# 実行
make demo

# または
./bin/skygraph-demo
```

**出力例：**

```
SkyGraph - Cloud Topology Scanner
Version: 0.1.0 (MVP)

Graph Stats:
  Nodes: 5
  Edges: 5

Graph exported to graph.json
```

**生成される graph.json の構造：**

```json
{
  "nodes": [
    {
      "id": "aws:vpc:vpc-123456",
      "type": "vpc",
      "provider": "aws",
      "region": "us-east-1",
      "name": "production-vpc",
      "metadata": {
        "cidr": "10.0.0.0/16"
      },
      "tags": {
        "Environment": "production"
      }
    },
    {
      "id": "aws:ec2:i-123456",
      "type": "ec2",
      "provider": "aws",
      "region": "us-east-1",
      "name": "web-server-1",
      "metadata": {
        "instance_type": "t3.large",
        "state": "running"
      }
    }
  ],
  "edges": [
    {
      "from": "aws:vpc:vpc-123456",
      "to": "aws:subnet:subnet-123456",
      "type": "ownership"
    },
    {
      "from": "aws:subnet:subnet-123456",
      "to": "aws:ec2:i-123456",
      "type": "network"
    }
  ]
}
```

---

## 2. 実際のAWSスキャン

### 前提条件

1. **AWS認証情報の設定**

```bash
# 方法1: 環境変数
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_REGION="us-east-1"

# 方法2: ~/.aws/credentials
aws configure
```

2. **IAM権限**

最低限必要な権限：

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeInstances",
        "ec2:DescribeVpcs",
        "ec2:DescribeSubnets",
        "ec2:DescribeSecurityGroups",
        "rds:DescribeDBInstances"
      ],
      "Resource": "*"
    }
  ]
}
```

### 基本的なスキャン

```bash
# ビルド
make build

# スキャン実行（us-east-1）
./bin/skygraph --provider aws --region us-east-1

# または Makefile 経由
make scan
```

**出力例：**

```
==============================================
  SkyGraph - Cloud Topology Scanner
  Version: 0.1.0 MVP
==============================================

Provider: aws
Region: us-east-1
Profile: default

Initializing AWS scanner...
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

### 詳細出力（verbose）

```bash
./bin/skygraph --provider aws --region us-east-1 --verbose
```

追加で以下が表示されます：

```
Edges by type:
  - ownership: 11
  - network: 15
  - dependency: 5
```

### 複数リージョンのスキャン

現在はサポートしていませんが、複数回実行して結果をマージできます：

```bash
# us-east-1
./bin/skygraph --provider aws --region us-east-1 --output graph-east.json

# us-west-2
./bin/skygraph --provider aws --region us-west-2 --output graph-west.json

# 後でマージ（将来のバージョンで実装予定）
```

---

## 3. 出力結果の確認

### graph.json の構造

```json
{
  "nodes": [
    {
      "id": "aws:vpc:vpc-0a1b2c3d",
      "type": "vpc",
      "provider": "aws",
      "region": "us-east-1",
      "name": "production-vpc",
      "metadata": {
        "vpc_id": "vpc-0a1b2c3d",
        "cidr_block": "10.0.0.0/16",
        "state": "available",
        "is_default": false
      },
      "tags": {
        "Name": "production-vpc",
        "Environment": "production"
      },
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T14:30:00Z"
    }
  ],
  "edges": [
    {
      "from": "aws:vpc:vpc-0a1b2c3d",
      "to": "aws:subnet:subnet-1a2b3c4d",
      "type": "ownership"
    }
  ]
}
```

### jq を使った分析

```bash
# ノード数
jq '.nodes | length' graph.json

# 特定タイプのリソース数
jq '.nodes | map(select(.type == "ec2")) | length' graph.json

# EC2インスタンスの一覧
jq '.nodes[] | select(.type == "ec2") | {id, name, instance_type: .metadata.instance_type}' graph.json

# 特定のVPCに属するリソース
jq '.nodes[] | select(.metadata.vpc_id == "vpc-0a1b2c3d") | {id, type, name}' graph.json

# エッジのタイプ別集計
jq '.edges | group_by(.type) | map({type: .[0].type, count: length})' graph.json
```

---

## 4. トラブルシューティング

### AWS認証エラー

```
Error: Failed to create AWS scanner: operation error: failed to resolve credentials
```

**解決策：**

```bash
# 認証情報を確認
aws sts get-caller-identity

# プロファイルを指定
./bin/skygraph --provider aws --profile your-profile
```

### 権限エラー

```
⚠ Some scanners failed:
  - ec2: AccessDenied: User is not authorized to perform: ec2:DescribeInstances
```

**解決策：**

必要なIAM権限を付与してください。

### タイムアウトエラー

```
Error: Scan failed: context deadline exceeded
```

**解決策：**

リージョン内のリソースが多すぎる場合、タイムアウトが発生する可能性があります。
将来のバージョンでタイムアウト設定をサポート予定です。

---

## 5. 次のステップ

### TiDBへの保存（v0.2.0で実装予定）

```bash
./bin/skygraph --provider aws --store tidb --dsn "root@tcp(localhost:4000)/airdig"
```

### 可視化（v0.4.0で実装予定）

AirDig UIでグラフを可視化：

```bash
# AirDig Engine 起動
airdig-engine --graph graph.json

# UI 起動
cd ui && npm run dev
```

---

## 6. 参考

- [README](./README.md) - SkyGraph の概要
- [設計書](./docs/design.md) - 詳細設計
- [AirDig メインドキュメント](../README.md)
