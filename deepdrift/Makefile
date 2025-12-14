.PHONY: build test clean install run-detect run-impact run-watch

# Build the DeepDrift CLI
build:
	@echo "Building DeepDrift..."
	@mkdir -p bin
	@go build -o bin/deepdrift ./cmd/deepdrift
	@echo "Build complete: bin/deepdrift"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Install DeepDrift to $GOPATH/bin
install:
	@echo "Installing DeepDrift..."
	@go install ./cmd/deepdrift
	@echo "Install complete"

# Run drift detection
run-detect:
	@./bin/deepdrift --command detect --state terraform.tfstate

# Run impact analysis
run-impact:
	@./bin/deepdrift --command impact --state terraform.tfstate --graph graph.json

# Run continuous monitoring
run-watch:
	@./bin/deepdrift --command watch --state terraform.tfstate --interval 5m

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run || echo "Install golangci-lint for linting: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

# Show help
help:
	@echo "DeepDrift Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build          Build the DeepDrift CLI"
	@echo "  make test           Run tests"
	@echo "  make test-coverage  Run tests with coverage report"
	@echo "  make clean          Clean build artifacts"
	@echo "  make install        Install DeepDrift to \$$GOPATH/bin"
	@echo "  make run-detect     Run drift detection"
	@echo "  make run-impact     Run impact analysis"
	@echo "  make run-watch      Run continuous monitoring"
	@echo "  make fmt            Format code"
	@echo "  make lint           Run linter"
