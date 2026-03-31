.PHONY: build install run logs uninstall clean test dev help

# Variables
BINARY_NAME=backup-installer
BUILD_DIR=./bin
CMD_DIR=./cmd/installer
VERSION=1.0.0

# Default target
all: build

## build: Compile the binary
build:
	@echo "🔨 Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags="-s -w -X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "✅ Binary created: $(BUILD_DIR)/$(BINARY_NAME)"

## install: Run the installer wizard
install: build
	@echo "🚀 Running installer..."
	@sudo $(BUILD_DIR)/$(BINARY_NAME)

## run-now: Execute backup pipeline manually
run-now:
	@echo "⚡ Running backup pipeline..."
	@./scripts/pipeline.sh

## logs: View backup logs
logs:
	@echo "📋 Tailing backup logs..."
	@tail -f ./logs/pipeline-$$(date +%Y-%m-%d).log

## uninstall: Remove crontab and generated files
uninstall:
	@echo "🗑️  Uninstalling backup installer..."
	@crontab -r 2>/dev/null || true
	@rm -f ./scripts/pipeline.sh
	@echo "✅ Uninstalled. Config preserved in ./config/"

## clean: Remove build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "✅ Clean complete"

## test: Run tests
test:
	@echo "🧪 Running tests..."
	@go test ./... -v

## audit: Run security audit before commit
audit:
	@./scripts/audit.sh

## dev: Run in development mode
dev:
	@echo "👨‍💻 Running in development mode..."
	@go run $(CMD_DIR)

## help: Show this help message
help:
	@echo "📦 INVITSM Backup Installer v$(VERSION)"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
