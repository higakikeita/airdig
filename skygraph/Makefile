.PHONY: build test clean demo install

# ビルド
build:
	@echo "Building SkyGraph..."
	@mkdir -p bin
	@go build -o bin/skygraph ./cmd/skygraph-full

# デモ版のビルド
build-demo:
	@echo "Building SkyGraph demo..."
	@mkdir -p bin
	@go build -o bin/skygraph-demo ./cmd/skygraph

# テスト
test:
	@echo "Running tests..."
	@go test -v ./...

# カバレッジ
coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# クリーン
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f graph.json
	@rm -f coverage.out coverage.html

# デモ実行
demo: build-demo
	@echo "Running demo..."
	@./bin/skygraph-demo

# AWS スキャン実行（us-east-1）
scan: build
	@echo "Scanning AWS (us-east-1)..."
	@./bin/skygraph --provider aws --region us-east-1 --output graph.json

# AWS スキャン実行（verbose）
scan-verbose: build
	@echo "Scanning AWS (verbose)..."
	@./bin/skygraph --provider aws --region us-east-1 --output graph.json --verbose

# 依存関係のダウンロード
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# フォーマット
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint
lint:
	@echo "Running linter..."
	@golangci-lint run

# インストール
install: build
	@echo "Installing to /usr/local/bin..."
	@cp bin/skygraph /usr/local/bin/
	@echo "Installed successfully!"

# ヘルプ
help:
	@echo "SkyGraph Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build          - Build SkyGraph CLI"
	@echo "  make build-demo     - Build demo version"
	@echo "  make test           - Run tests"
	@echo "  make coverage       - Generate coverage report"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make demo           - Run demo"
	@echo "  make scan           - Scan AWS (requires credentials)"
	@echo "  make scan-verbose   - Scan AWS with verbose output"
	@echo "  make deps           - Download dependencies"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Run linter"
	@echo "  make install        - Install to /usr/local/bin"
