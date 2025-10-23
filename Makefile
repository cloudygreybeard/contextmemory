# ContextMemory MCP - Development Automation
# Model Context Protocol server for conversation context management

.PHONY: help setup build test clean install uninstall dev.iterate dev.clean-iterate info.status version.sync version.validate version.bump-patch version.bump-minor version.bump-major

# Default target
help:
	@echo "ContextMemory MCP - Development Commands"
	@echo "========================================"
	@echo ""
	@echo "Setup & Build:"
	@echo "  setup           Initialize development environment"
	@echo "  build           Build all components"
	@echo "  build.server    Build MCP server for current platform"
	@echo "  build.memctl    Build memctl CLI for current platform"
	@echo "  build.all       Build all architectures (requires GoReleaser)"
	@echo "  clean           Clean build artifacts"
	@echo ""
	@echo "Development:"
	@echo "  dev.iterate     Full build and install cycle"
	@echo "  dev.clean-iterate  Clean uninstall -> build -> install cycle"
	@echo ""
	@echo "Testing:"
	@echo "  test            Run all tests"
	@echo "  test.unit       Run unit tests"
	@echo "  test.integration Run integration tests"
	@echo "  test.coverage   Run tests with coverage"
	@echo ""
	@echo "Installation:"
	@echo "  install         Install MCP server and memctl locally"
	@echo "  install.server  Install MCP server binary only"
	@echo "  install.memctl  Install memctl binary only"
	@echo "  uninstall       Remove all binaries"
	@echo ""
	@echo "Release:"
	@echo "  release         Create release with GoReleaser"
	@echo "  release.snapshot Create snapshot release"
	@echo "  release.validate Run release validation"
	@echo ""
	@echo "Version Management:"
	@echo "  version.sync    Sync all versions from git tag"
	@echo "  version.validate Check version consistency across components"
	@echo "  version.bump-patch Bump patch version with git tag"
	@echo "  version.bump-minor Bump minor version with git tag"
	@echo "  version.bump-major Bump major version with git tag"
	@echo ""
	@echo "Information:"
	@echo "  info.status     Show project status"

# Get version from git tag for build ldflags
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Setup development environment
setup:
	@echo "[SETUP] Initializing ContextMemory MCP development environment..."
	@mkdir -p build/bin hack docs
	@echo "[SETUP] Created directory structure"
	@if [ ! -f .gitignore ]; then \
		echo "# ContextMemory MCP" > .gitignore; \
		echo "build/" >> .gitignore; \
		echo "*.log" >> .gitignore; \
		echo "*.tmp" >> .gitignore; \
		echo "*.temp" >> .gitignore; \
		echo "coverage.out" >> .gitignore; \
		echo "coverage.html" >> .gitignore; \
		echo ".DS_Store" >> .gitignore; \
		echo "[SETUP] Created .gitignore"; \
	fi
	@echo "[SUCCESS] Development environment ready"

# Build all components
build: build.server build.memctl

# Build MCP server
build.server:
	@echo "[BUILD] Building MCP server..."
	@mkdir -p build/bin
	@cd cmd/mcp-server-sdk && go build $(LDFLAGS) -o ../../build/bin/contextmemory .
	@echo "[SUCCESS] MCP server built: build/bin/contextmemory"

# Build memctl CLI
build.memctl:
	@echo "[BUILD] Building memctl CLI..."
	@mkdir -p build/bin
	@cd cmd/memctl && go build $(LDFLAGS) -o ../../build/bin/memctl .
	@echo "[SUCCESS] memctl CLI built: build/bin/memctl"

# Build all architectures (requires GoReleaser)
build.all:
	@echo "[BUILD] Building for all architectures..."
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser build --snapshot --clean; \
	else \
		echo "[ERROR] GoReleaser not installed. Install with: go install github.com/goreleaser/goreleaser@latest"; \
		exit 1; \
	fi
	@echo "[SUCCESS] Multi-architecture builds completed"

# Development iteration cycle
dev.iterate: build install
	@echo "[DEV] Development iteration complete!"

# Clean development iteration cycle (uninstall -> build -> install)
dev.clean-iterate: uninstall build install
	@echo "[DEV] Clean development iteration complete!"

# Install MCP server and memctl locally
install: install.server install.memctl

install.server:
	@echo "[INSTALL] Installing ContextMemory MCP server..."
	@if [ -f build/bin/contextmemory ]; then \
		echo "[INSTALL] Installing MCP server to /usr/local/bin/contextmemory (requires sudo)..."; \
		sudo cp build/bin/contextmemory /usr/local/bin/contextmemory && echo "[SUCCESS] MCP server installed to /usr/local/bin/contextmemory"; \
	else \
		echo "[INFO] MCP server not built - run 'make build.server' first"; \
	fi

install.memctl:
	@echo "[INSTALL] Installing memctl CLI..."
	@if [ -f build/bin/memctl ]; then \
		echo "[INSTALL] Installing memctl to /usr/local/bin/memctl (requires sudo)..."; \
		sudo cp build/bin/memctl /usr/local/bin/memctl && echo "[SUCCESS] memctl installed to /usr/local/bin/memctl"; \
	else \
		echo "[INFO] memctl not built - run 'make build.memctl' first"; \
	fi

# Uninstall binaries
uninstall:
	@echo "[UNINSTALL] Removing ContextMemory MCP binaries..."
	@if [ -f /usr/local/bin/contextmemory ]; then \
		echo "[UNINSTALL] Removing MCP server from /usr/local/bin/contextmemory (requires sudo)..."; \
		sudo rm -f /usr/local/bin/contextmemory && echo "[SUCCESS] MCP server removed"; \
	else \
		echo "[INFO] MCP server not installed at /usr/local/bin/contextmemory"; \
	fi
	@if [ -f /usr/local/bin/memctl ]; then \
		echo "[UNINSTALL] Removing memctl from /usr/local/bin/memctl (requires sudo)..."; \
		sudo rm -f /usr/local/bin/memctl && echo "[SUCCESS] memctl removed"; \
	else \
		echo "[INFO] memctl not installed at /usr/local/bin/memctl"; \
	fi

# Testing
test: test.unit test.integration

test.unit:
	@echo "[TEST] Running unit tests..."
	@go test -v ./...

test.integration:
	@echo "[TEST] Running integration tests..."
	@if [ -f hack/test-integration.sh ]; then ./hack/test-integration.sh; else echo "[INFO] No integration tests configured"; fi

test.coverage:
	@echo "[TEST] Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "[INFO] Coverage report: coverage.html"

# Clean build artifacts
clean:
	@echo "[CLEAN] Removing build artifacts..."
	@rm -rf build/ 2>/dev/null || true
	@rm -f coverage.out coverage.html 2>/dev/null || true
	@echo "[SUCCESS] Build artifacts cleaned"

# Release with GoReleaser
release:
	@echo "[RELEASE] Creating release..."
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --clean; \
	else \
		echo "[ERROR] GoReleaser not installed. Install with: go install github.com/goreleaser/goreleaser@latest"; \
		exit 1; \
	fi
	@echo "[SUCCESS] Release completed"

# Release snapshot for testing
release.snapshot:
	@echo "[RELEASE] Creating snapshot release..."
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --snapshot --clean; \
	else \
		echo "[ERROR] GoReleaser not installed. Install with: go install github.com/goreleaser/goreleaser@latest"; \
		exit 1; \
	fi
	@echo "[SUCCESS] Snapshot release completed"

release.validate:
	@echo "[VALIDATE] Running release validation..."
	@if [ -f hack/validate-release.sh ]; then ./hack/validate-release.sh; else echo "[INFO] No release validation configured"; fi

# Information commands
info.status:
	@echo "ContextMemory MCP Project Status"
	@echo "================================"
	@echo ""
	@echo "Architecture: Model Context Protocol server with file-based storage"
	@echo "Components:"
	@if [ -f build/bin/contextmemory ]; then echo "  ✓ MCP server (built)"; else echo "  ✗ MCP server (not built)"; fi
	@if [ -f build/bin/memctl ]; then echo "  ✓ memctl CLI (built)"; else echo "  ✗ memctl CLI (not built)"; fi
	@echo ""
	@echo "Installation Status:"
	@if [ -f /usr/local/bin/contextmemory ]; then \
		echo "  ✓ MCP server: /usr/local/bin/contextmemory"; \
		echo "  Version: $$(contextmemory --version 2>/dev/null | head -1 || echo 'Unknown')"; \
	else \
		echo "  ✗ MCP server not installed"; \
	fi
	@if [ -f /usr/local/bin/memctl ]; then \
		echo "  ✓ memctl CLI: /usr/local/bin/memctl"; \
		echo "  Version: $$(memctl version --output text 2>/dev/null | grep 'Client Version:' || echo 'Unknown')"; \
	else \
		echo "  ✗ memctl CLI not installed"; \
	fi
	@echo ""
	@echo "Storage:"
	@if [ -d ~/.contextmemory ]; then \
		echo "  Location: ~/.contextmemory"; \
		echo "  Memories: $$(find ~/.contextmemory -name "*.json" 2>/dev/null | wc -l | tr -d ' ') files"; \
	else \
		echo "  ✗ Storage directory not initialized"; \
	fi

# Git-based version management
version.sync:
	@./hack/version sync

version.validate:
	@./hack/version validate

version.bump-patch:
	@./hack/version bump patch

version.bump-minor:
	@./hack/version bump minor

version.bump-major:
	@./hack/version bump major




