.PHONY: build run clean test help

# Build configuration
BINARY_NAME=tracecore
CMD_PATH=./cmd/tracecore
BUILD_DIR=./build

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build the binary
build:
	@echo "Building TraceCore..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the application
run: build
	@echo "Starting TraceCore..."
	$(BUILD_DIR)/$(BINARY_NAME)

# Run without ClickHouse
run-no-db: build
	@echo "Starting TraceCore without ClickHouse..."
	CLICKHOUSE_ENABLED=false $(BUILD_DIR)/$(BINARY_NAME)

# Run tests
test:
	$(GOTEST) -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Install dependencies
deps:
	@echo "Fetching dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies fetched"

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Display help
help:
	@echo "TraceCore Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build       - Build the TraceCore binary"
	@echo "  run         - Build and run TraceCore"
	@echo "  run-no-db   - Build and run without ClickHouse"
	@echo "  test        - Run tests"
	@echo "  clean       - Remove build artifacts"
	@echo "  deps        - Install dependencies"
	@echo "  fmt         - Format code"
	@echo "  lint        - Run linter"
	@echo "  help        - Display this help message"
	@echo ""
	@echo "Environment variables:"
	@echo "  CLICKHOUSE_HOST     - ClickHouse server host (default: localhost)"
	@echo "  CLICKHOUSE_PORT     - ClickHouse server port (default: 9000)"
	@echo "  CLICKHOUSE_DATABASE - ClickHouse database name (default: tracecore)"
	@echo "  OTLP_GRPC_PORT      - OTLP gRPC receiver port (default: 4317)"
	@echo "  OTLP_HTTP_PORT      - OTLP HTTP receiver port (default: 4318)"
	@echo "  API_PORT            - HTTP API server port (default: 8082)"
