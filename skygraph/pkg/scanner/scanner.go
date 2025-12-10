package scanner

import (
	"context"

	"github.com/yourusername/airdig/skygraph/pkg/graph"
)

// Scanner はクラウドリソースをスキャンするためのインターフェース
type Scanner interface {
	// Scan はリソースをスキャンして ResourceNode のスライスを返す
	Scan(ctx context.Context) ([]graph.ResourceNode, error)

	// Name はスキャナーの名前を返す（例: "ec2", "vpc"）
	Name() string
}

// Config はスキャナーの設定
type Config struct {
	// Provider はクラウドプロバイダー名（aws, gcp, azure, kubernetes）
	Provider string

	// Region はリージョン名（AWS/GCPの場合）
	Region string

	// Profile は AWS プロファイル名（オプション）
	Profile string

	// Resources はスキャン対象リソースタイプ（空の場合は全て）
	Resources []string
}

// Result はスキャン結果
type Result struct {
	// Nodes はスキャンで取得したリソースノード
	Nodes []graph.ResourceNode

	// Errors は各スキャナーで発生したエラー（部分的失敗を許容）
	Errors map[string]error
}
