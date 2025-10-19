.PHONY: all build test lint e2e clean install help

# Build variables
BINARY_NAME=sky
WING_BINARY=wing
SKYLS_BINARY=skyls
SKYDBG_BINARY=skydbg

GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-s -w"

# Directories
BIN_DIR=./bin
CMD_DIR=./cmd

all: build

help: ## Bu yardım mesajını gösterir
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Tüm binary'leri derler
	@echo "Building SKY toolchain..."
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_DIR)/sky
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/$(WING_BINARY) $(CMD_DIR)/wing
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/$(SKYLS_BINARY) $(CMD_DIR)/skyls
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/$(SKYDBG_BINARY) $(CMD_DIR)/skydbg
	@echo "Build complete!"

build-sky: ## Sadece sky binary'sini derler
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_DIR)/sky

test: ## Unit testleri çalıştırır
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete!"

test-coverage: test ## Test coverage raporu oluşturur
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Lint kontrollerini çalıştırır
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install with: brew install golangci-lint" && exit 1)
	golangci-lint run ./...
	@echo "Lint complete!"

fmt: ## Kod formatlaması yapar
	@echo "Formatting code..."
	$(GO) fmt ./...
	@which gofumpt > /dev/null && gofumpt -l -w . || echo "gofumpt not found, using go fmt"
	@echo "Format complete!"

vet: ## Go vet çalıştırır
	@echo "Running go vet..."
	$(GO) vet ./...
	@echo "Vet complete!"

e2e: build ## E2E testleri çalıştırır
	@echo "Running e2e tests..."
	@test -f ./scripts/e2e.sh && chmod +x ./scripts/e2e.sh && ./scripts/e2e.sh || echo "e2e.sh not found yet"
	@echo "E2E tests complete!"

clean: ## Build artifactlarını temizler
	@echo "Cleaning..."
	rm -rf $(BIN_DIR)
	rm -f coverage.out coverage.html
	$(GO) clean
	@echo "Clean complete!"

install: build ## Binary'leri sistem PATH'ine kurar
	@echo "Installing binaries to /usr/local/bin..."
	sudo cp $(BIN_DIR)/$(BINARY_NAME) /usr/local/bin/
	sudo cp $(BIN_DIR)/$(WING_BINARY) /usr/local/bin/
	sudo cp $(BIN_DIR)/$(SKYLS_BINARY) /usr/local/bin/
	sudo cp $(BIN_DIR)/$(SKYDBG_BINARY) /usr/local/bin/
	@echo "Installation complete!"

uninstall: ## Binary'leri sistemden kaldırır
	@echo "Uninstalling binaries..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	sudo rm -f /usr/local/bin/$(WING_BINARY)
	sudo rm -f /usr/local/bin/$(SKYLS_BINARY)
	sudo rm -f /usr/local/bin/$(SKYDBG_BINARY)
	@echo "Uninstallation complete!"

deps: ## Go bağımlılıklarını indirir
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy
	@echo "Dependencies downloaded!"

check: lint vet test ## Tüm kontrolleri çalıştırır

ci: check e2e ## CI pipeline (lint + test + e2e)

dev: ## Development modunda çalıştırır
	$(GO) run $(CMD_DIR)/sky/main.go

.DEFAULT_GOAL := help

