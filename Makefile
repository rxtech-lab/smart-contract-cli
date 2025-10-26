.PHONY: e2e-network e2e-test e2e-test-stop

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