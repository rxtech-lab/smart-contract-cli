.PHONY: e2e-network e2e-test e2e-test-stop lint

build:
	@echo "Building smart-contract-cli..."
	@go build -o smart-contract-cli ./main.go

e2e-network:
	@echo "Starting Anvil network..."
	pkill -f "anvil" || true
	anvil &
	@echo "Waiting for Anvil to be ready..."
	@sleep 2

test: e2e-network
	@echo "Running tests..."
	@go test ./...
	@$(MAKE) e2e-test-stop

e2e-test-stop:
	@echo "Stopping Anvil network..."
	pkill -f "anvil" || true

fmt:
	@echo "Formatting code..."
	@go fmt ./...

lint:
	@echo "Running golangci-lint..."
	@if command -v golangci-lint > /dev/null 2>&1; then \
		golangci-lint run ./...; \
	elif [ -f ~/go/bin/golangci-lint ]; then \
		~/go/bin/golangci-lint run ./...; \
	else \
		echo "golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi