BINARY_NAME=gommits
BUILD_DIR=bin
MAIN_PATH=./cmd/app
GO_FILES=$(shell find . -name "*.go" -type f)

.DEFAULT_GOAL := help

.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: run
run:
	@echo "Running $(BINARY_NAME)..."
	go run $(MAIN_PATH)

.PHONY: build-run
build-run: build
	@echo "Running built binary..."
	./$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	go clean

.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

.PHONY: lint
lint: ## requires golangci-lint
	@echo "Linting code..."
	golangci-lint run

.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(MAIN_PATH)

# Build for multiple platforms
.PHONY: build-all
build-all: ## Build for multiple platforms
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Cross-platform builds complete"

# with file watching - requires air
.PHONY: dev
dev:
	@echo "Starting development mode..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not found. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Falling back to regular run..."; \
		$(MAKE) run; \
	fi
