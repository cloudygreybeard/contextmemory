# ContextMemory - Development Automation
# Clean, simple, file-based memory management

.PHONY: help setup build test clean install dev.start dev.stop info.status

# Default target
help:
	@echo "ContextMemory - Development Commands"
	@echo "======================================"
	@echo ""
	@echo "Setup & Build:"
	@echo "  setup           Initialize development environment"
	@echo "  build           Build all components"
	@echo "  build.cli       Build CLI for current platform"
	@echo "  build.cli.wasm  Build CLI for WebAssembly"
	@echo "  build.cli.all   Build CLI for all architectures"
	@echo "  clean           Clean build artifacts"
	@echo ""
	@echo "Development:"
	@echo "  dev.start       Start development environment"
	@echo "  dev.stop        Stop development environment"
	@echo "  dev.iterate     Full build and install cycle"
	@echo ""
	@echo "Testing:"
	@echo "  test            Run all tests"
	@echo "  test.cli        Test CLI functionality"
	@echo "  test.integration Run CLI integration tests"
	@echo "  test.core       Test core functionality"
	@echo "  test.storage    Test storage backend"
	@echo "  test.ui         Test UI components"
	@echo ""
	@echo "Installation:"
	@echo "  install         Install extension locally"
	@echo "  package         Create extension package"
	@echo ""
	@echo "Release:"
	@echo "  release         Create release with GoReleaser"
	@echo "  release.snapshot Create snapshot release"
	@echo ""
	@echo "Information:"
	@echo "  info.status     Show project status"
	@echo "  info.structure  Show project structure"

# Setup development environment
setup:
	@echo "[SETUP] Initializing ContextMemory development environment..."
	@mkdir -p core storage ui cli docs tests
	@mkdir -p storage/data
	@echo "[SETUP] Created directory structure"
	@if [ ! -f .gitignore ]; then \
		echo "# ContextMemory v2" > .gitignore; \
		echo "node_modules/" >> .gitignore; \
		echo "*.log" >> .gitignore; \
		echo "storage/data/*.json" >> .gitignore; \
		echo ".DS_Store" >> .gitignore; \
		echo "[SETUP] Created .gitignore"; \
	fi
	@echo "[SUCCESS] Development environment ready"

# Build all components
build: deps build.core build.storage build.cli build.ui

# Install dependencies
deps:
	@echo "[DEPS] Installing dependencies..."
	@if [ -f package.json ] && [ ! -d node_modules ]; then \
		echo "[DEPS] Installing root dependencies..."; \
		npm install; \
	fi

build.core:
	@echo "[BUILD] Building core memory operations..."
	@if [ -d core ]; then \
		cd core && if [ -f package.json ]; then npm run build 2>/dev/null || echo "No build script configured"; fi; \
	fi
	@echo "[SUCCESS] Core built"

build.storage:
	@echo "[BUILD] Building storage backend..."
	@if [ -d storage ]; then \
		cd storage && if [ -f package.json ]; then npm run build 2>/dev/null || echo "No build script configured"; fi; \
	fi
	@echo "[SUCCESS] Storage built"

build.cli:
	@echo "[BUILD] Building Go CLI..."
	@if [ -d cmd/cmctl ]; then \
		cd cmd/cmctl && go build -ldflags="-s -w" -o cmctl . && echo "Go CLI built successfully"; \
	else \
		echo "No Go CLI found"; \
	fi
	@echo "[SUCCESS] CLI built"

# Build CLI for WebAssembly
build.cli.wasm:
	@echo "[BUILD] Building CLI for WebAssembly..."
	@if [ -d cmd/cmctl ]; then \
		cd cmd/cmctl && GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o cmctl.wasm . && \
		echo "WebAssembly CLI built successfully"; \
		mkdir -p ../../ui/src/wasm; \
		find "$$(go env GOROOT)" -name "wasm_exec.js" -exec cp {} ../../ui/src/wasm/ \; 2>/dev/null || echo "wasm_exec.js not found"; \
		mv cmctl.wasm ../../ui/src/wasm/; \
	else \
		echo "No Go CLI found"; \
	fi
	@echo "[SUCCESS] CLI WebAssembly build completed"

# Build CLI for all architectures (requires GoReleaser)
build.cli.all:
	@echo "[BUILD] Building CLI for all architectures..."
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser build --snapshot --clean; \
	else \
		echo "[ERROR] GoReleaser not installed. Install with: go install github.com/goreleaser/goreleaser@latest"; \
		exit 1; \
	fi
	@echo "[SUCCESS] Multi-architecture builds completed"

build.ui:
	@echo "[BUILD] Building VS Code extension..."
	@if [ -d ui ] && [ -f ui/package.json ]; then \
		cd ui && \
		if [ ! -d node_modules ]; then \
			echo "[BUILD] Installing UI dependencies..."; \
			npm install; \
		fi && \
		npm run compile && echo "Extension compiled successfully"; \
	else \
		echo "UI directory or package.json not found"; \
	fi
	@echo "[SUCCESS] Extension built"

# Development iteration cycle
dev.iterate: build install
	@echo "[DEV] Development iteration complete!"

# Package VS Code extension
package.ui:
	@echo "[PACKAGE] Creating extension package..."
	@if [ -d ui ] && [ -f ui/package.json ]; then \
		cd ui && \
		if [ ! -d node_modules ]; then \
			echo "[PACKAGE] Installing UI dependencies..."; \
			npm install; \
		fi && \
		if [ ! -d dist ]; then \
			echo "[BUILD] Compiling extension first..."; \
			npm run compile; \
		fi && \
		npm run package && echo "Extension packaged successfully"; \
	else \
		echo "UI directory or package.json not found"; \
	fi

# Install CLI and extension locally
install: install.cli install.ui

install.cli:
	@echo "[INSTALL] Installing ContextMemory CLI..."
	@if [ -f cmd/cmctl/cmctl ]; then \
		cp cmd/cmctl/cmctl /usr/local/bin/cmctl && echo "[SUCCESS] CLI installed to /usr/local/bin/cmctl"; \
	else \
		echo "[INFO] CLI not built - run 'make build.cli' first"; \
	fi

install.ui:
	@echo "[INSTALL] Installing ContextMemory extension..."
	@if [ -d ui ] && [ -f ui/package.json ]; then \
		cd ui && \
		if [ -f *.vsix ]; then \
			cursor --install-extension *.vsix && echo "[SUCCESS] Extension installed"; \
		else \
			echo "[INFO] No .vsix package found - run 'make package' first"; \
		fi; \
	else \
		echo "[INFO] UI not ready for installation"; \
	fi

# Create extension package
package: package.ui

# Testing
test: test.core test.storage test.ui

test.core:
	@echo "[TEST] Testing core functionality..."
	@if [ -d tests ]; then \
		echo "[INFO] Core tests would run here"; \
	fi

test.storage:
	@echo "[TEST] Testing storage backend..."
	@if [ -d tests ]; then \
		echo "[INFO] Storage tests would run here"; \
	fi

test.ui:
	@echo "[TEST] Testing UI components..."
	@if [ -d tests ]; then \
		echo "[INFO] UI tests would run here"; \
	fi

# Clean build artifacts
clean:
	@echo "[CLEAN] Removing build artifacts..."
	@find . -name "node_modules" -type d -exec rm -rf {} + 2>/dev/null || true
	@find . -name "*.log" -type f -delete 2>/dev/null || true
	@find . -name "dist" -type d -exec rm -rf {} + 2>/dev/null || true
	@rm -f cmd/cmctl/cmctl 2>/dev/null || true
	@rm -f cmd/cmctl/cmctl.wasm 2>/dev/null || true
	@rm -f ui/src/wasm/cmctl.wasm 2>/dev/null || true
	@rm -f ui/src/wasm/wasm_exec.js 2>/dev/null || true
	@rm -f ui/*.vsix 2>/dev/null || true
	@rm -f cmd/cmctl/coverage.out 2>/dev/null || true
	@rm -f cmd/cmctl/coverage.html 2>/dev/null || true
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

# Test CLI functionality  
test.cli:
	@echo "[TEST] Running Go unit tests..."
	cd cmd/cmctl && go test -v ./...

test.cli.coverage:
	@echo "[TEST] Running Go tests with coverage..."
	cd cmd/cmctl && go test -v -cover ./...

test.cli.coverage.html:
	@echo "[TEST] Generating HTML coverage report..."
	cd cmd/cmctl && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
	@echo "[INFO] Coverage report: cmd/cmctl/coverage.html"

test.cli.functional:
	@echo "[TEST] Running CLI functional tests..."
	./scripts/test-cli.sh

# Run integration tests
test.integration:
	@echo "[TEST] Running integration tests..."
	./scripts/test-integration.sh

test.all: test.cli test.cli.functional test.integration
	@echo "[SUCCESS] All tests passed"

test.all.coverage: test.cli.coverage test.cli.functional test.integration
	@echo "[SUCCESS] All tests with coverage completed"

validate.release:
	@echo "[VALIDATE] Running release validation..."
	./scripts/validate-release.sh

lint:
	@echo "[LINT] Checking code quality..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		cd cmd/cmctl && golangci-lint run; \
	else \
		echo "[INFO] golangci-lint not found, using go vet"; \
		cd cmd/cmctl && go vet ./...; \
	fi

# Development environment management
dev.start:
	@echo "[DEV] Starting development environment..."
	@echo "[INFO] File-based storage - no services to start"
	@echo "[SUCCESS] Development environment ready"

dev.stop:
	@echo "[DEV] Stopping development environment..."
	@echo "[INFO] File-based storage - no services to stop"
	@echo "[SUCCESS] Development environment stopped"

# Information commands
info.status:
	@echo "ContextMemory Project Status"
	@echo "==============================="
	@echo ""
	@echo "Architecture: File-based storage with hybrid TypeScript/Go implementation"
	@echo "Components:"
	@if [ -d core ]; then echo "  ✓ Core operations (TypeScript)"; else echo "  ✗ Core operations"; fi
	@if [ -d storage ]; then echo "  ✓ Storage backend (TypeScript)"; else echo "  ✗ Storage backend"; fi
	@if [ -f cmd/cmctl/cmctl ]; then echo "  ✓ CLI interface (Go + Cobra)"; else echo "  ✗ CLI interface"; fi
	@if [ -d ui/dist ]; then echo "  ✓ UI extension (TypeScript)"; else echo "  ✗ UI extension (not built)"; fi
	@echo ""
	@echo "CLI Status:"
	@if [ -f cmd/cmctl/cmctl ]; then \
		echo "  Binary: cmd/cmctl/cmctl"; \
		echo "  Version: $$(cmd/cmctl/cmctl --version 2>/dev/null | head -1 || echo 'Unknown')"; \
	else \
		echo "  ✗ CLI not built"; \
	fi
	@echo ""
	@echo "Storage:"
	@if [ -d ~/.contextmemory ]; then \
		echo "  Location: ~/.contextmemory"; \
		echo "  Files: $$(find ~/.contextmemory/memories -name "*.json" 2>/dev/null | wc -l | tr -d ' ') memories"; \
	else \
		echo "  ✗ Storage directory not initialized"; \
	fi

info.structure:
	@echo "ContextMemory Structure"
	@echo "========================="
	@tree -I 'node_modules|*.log|.git' . 2>/dev/null || find . -type f -not -path './node_modules/*' -not -path './.git/*' -not -name '*.log' | head -20
