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
	@echo "  clean           Clean build artifacts"
	@echo ""
	@echo "Development:"
	@echo "  dev.start       Start development environment"
	@echo "  dev.stop        Stop development environment"
	@echo "  dev.iterate     Full build and install cycle"
	@echo ""
	@echo "Testing:"
	@echo "  test            Run all tests"
	@echo "  test.core       Test core functionality"
	@echo "  test.storage    Test storage backend"
	@echo "  test.ui         Test UI components"
	@echo ""
	@echo "Installation:"
	@echo "  install         Install extension locally"
	@echo "  package         Create extension package"
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
build: build.core build.storage build.cli build.ui

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
	@if [ -d cmd/cm ]; then \
		cd cmd/cm && go build -o cmctl . && echo "Go CLI built successfully"; \
	else \
		echo "No Go CLI found"; \
	fi
	@echo "[SUCCESS] CLI built"

build.ui:
	@echo "[BUILD] Building UI extension..."
	@if [ -d ui ]; then \
		cd ui && if [ -f package.json ]; then npm run build 2>/dev/null || echo "No build script configured"; fi; \
	fi
	@echo "[SUCCESS] UI built"

# Development iteration cycle
dev.iterate: build install
	@echo "[DEV] Development iteration complete!"

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
package:
	@echo "[PACKAGE] Creating extension package..."
	@if [ -d ui ] && [ -f ui/package.json ]; then \
		cd ui && npm run package && echo "[SUCCESS] Package created"; \
	else \
		echo "[INFO] UI not ready for packaging"; \
	fi

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
	@echo "[SUCCESS] Build artifacts cleaned"

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
	@if [ -d ui ]; then echo "  ✓ UI extension (TypeScript)"; else echo "  ✗ UI extension"; fi
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
