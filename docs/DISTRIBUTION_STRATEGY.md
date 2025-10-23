# ContextMemory Distribution Strategy

## Overview

This document outlines distribution options for the ContextMemory MCP server, focusing on binary distribution and automated release processes.

## Distribution Options Analysis

### Option 1: Binary Distribution

```ascii
Binary Distribution Architecture:

┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Developer     │    │   Local Binary  │    │   Local Storage │
│   Machine       │    │  /usr/local/bin │    │ ~/.contextmemory│
│                 │    │                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │   Cursor    │─┼────┼▶│ mcp-server  │─┼────┼▶│ memories/   │ │
│ │    IDE      │ │    │ │   binary    │ │    │ │ threads/    │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ │ metadata/   │ │
│                 │    │                 │    │ └─────────────┘ │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │                 │
│ │   VS Code   │─┼────┼▶│   Config    │ │    │                 │
│ │    IDE      │ │    │ │ mcp.json    │ │    │                 │
│ └─────────────┘ │    │ └─────────────┘ │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

**Pros:**
- **Zero Infrastructure**: No server maintenance required
- **Privacy**: All data stays local 
- **Low Latency**: Direct filesystem access
- **Offline Capable**: Works without internet
- **Simple Installation**: Single binary + config

**Cons:**
- **No Cross-Device Sync**: Memories tied to specific machine
- **No Collaboration**: Single-user only
- **Manual Updates**: User must update binary
- **Limited Analytics**: No usage insights


### Option 2: Automated Release Process

Focus on improving the binary distribution with automated build and release processes:

```ascii
Release Automation Pipeline:

┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐
│   Git Push       │    │ GitHub Actions   │    │   GoReleaser     │
│   (main branch)  │───▶│ Workflow Trigger │───▶│  Cross-Platform  │
│                  │    │                  │    │     Builds       │
└──────────────────┘    └──────────────────┘    └──────────────────┘
                                                           │
                                                           ▼
┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐
│Package Managers  │    │ GitHub Releases  │    │   Binary Assets  │
│(brew, apt, etc.) │◀───│   Automated      │◀───│ (darwin, linux,  │
│                  │    │   Publishing     │    │   windows)       │
└──────────────────┘    └──────────────────┘    └──────────────────┘
```

**Benefits:**
- Consistent cross-platform builds
- Automated version management and changelog generation
- Package manager integration (Homebrew, APT, etc.)
- Reduced manual release overhead
- Faster iteration and deployment cycles

## Implementation Plans

### Phase 1: Binary Distribution

#### Installation Process

```bash
# 1. Download binary
curl -L https://github.com/contextmemory/releases/latest/download/contextmemory-darwin-arm64.tar.gz | tar xz

# 2. Install to system path  
sudo mv contextmemory /usr/local/bin/
sudo chmod +x /usr/local/bin/contextmemory

# 3. Initialize configuration
contextmemory init

# 4. Configure IDE
contextmemory config cursor
```

#### IDE Configuration

**Cursor Configuration (`~/.cursor/mcp.json`):**
```json
{
  "mcpServers": {
    "contextmemory": {
      "command": "/usr/local/bin/contextmemory",
      "args": ["mcp-server"],
      "env": {
        "CONTEXTMEMORY_HOME": "~/.contextmemory",
        "CONTEXTMEMORY_LOG_LEVEL": "info"
      }
    }
  }
}
```

**VS Code Configuration (`.vscode/mcp.json`):**
```json
{
  "mcpServers": {
    "contextmemory": {
      "command": "/usr/local/bin/contextmemory", 
      "args": ["mcp-server", "--ide=vscode"],
      "cwd": "${workspaceFolder}"
    }
  }
}
```

#### Package Managers (Planned)

Package manager support is planned for future releases:

```bash
# Homebrew (macOS) - Planned
brew install contextmemory/tap/contextmemory

# APT (Ubuntu/Debian) - Planned
sudo apt install contextmemory

# Chocolatey (Windows) - Planned
choco install contextmemory

# Snap (Linux) - Planned
sudo snap install contextmemory
```

### Phase 2: Release Automation Implementation

#### GoReleaser Configuration

```yaml
# .goreleaser.yaml
version: 2

before:
  hooks:
    - go mod tidy
    - go test ./...

builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - windows
      - darwin
    goarch:
      - amd64
      - arm64
    main: ./cmd/mcp-server-sdk
    binary: contextmemory

archives:
  - format: tar.gz
    name_template: >-
      {{ .ProjectName }}_
      {{- .Version }}_
      {{- title .Os }}_
      {{- if eq .Arch "amd64" }}x86_64
      {{- else if eq .Arch "386" }}i386
      {{- else }}{{ .Arch }}{{ end }}

checksum:
  name_template: 'checksums.txt'

release:
  github:
    owner: contextmemory
    name: contextmemory

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
```

#### GitHub Actions Workflow

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v4
        with:
          go-version: stable
      - uses: goreleaser/goreleaser-action@v5
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```


## Distribution Mechanics

### Binary Distribution Details

#### Build System

```yaml
# GitHub Actions workflow
name: Release
on:
  push:
    tags: ['v*']

jobs:
  build:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        arch: [amd64, arm64]
    
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
      
      - name: Build binary
        run: |
          GOOS=${{ matrix.os }} GOARCH=${{ matrix.arch }} \
          go build -ldflags="-s -w" -o contextmemory-${{ matrix.os }}-${{ matrix.arch }}
      
      - name: Create release
        uses: goreleaser/goreleaser-action@v3
```

#### Installation Methods

```ascii
Installation Flow Options:

Option A: Direct Binary
├── Download from GitHub releases
├── Copy to /usr/local/bin/
├── Make executable (chmod +x)
└── Run contextmemory init

Option B: Package Manager
├── Add repository to system
├── Install via package manager
├── Auto-configure system paths
└── Service management included

Option C: Build from Source
├── Clone repository
├── Build with Go toolchain
├── Install to system path
├── Configure IDE manually
└── Verify installation
```

## Recommended Approach

### Phase 1: Binary Distribution
1. Focus on binary distribution for market validation
2. Simple installation process via curl script and package managers
3. IDE configuration templates for Cursor, VS Code, and others
4. Local storage optimization for performance and privacy

### Phase 2: Enhanced Binary  
1. Auto-update mechanism for seamless upgrades
2. Configuration management tools and improved CLI
3. Export/import capabilities for data portability
4. Multiple IDE support with universal MCP client

### Phase 3: Extended Distribution
1. Package manager integrations (Homebrew, APT, Scoop, etc.)
2. Installer packages for major operating systems
3. Verification and signing for security
4. Usage analytics and crash reporting (opt-in)

This approach focuses on perfecting the binary distribution model while maintaining simplicity and user control.

## Implementation Next Steps

1. **Set up GoReleaser configuration** (`.goreleaser.yaml` created)
2. **Configure GitHub Actions workflow** (`.github/workflows/release.yml` created) 
3. **Test release process** with a development tag
4. **Set up package manager repositories** (Homebrew tap, APT repository)
5. **Create installation documentation** and verification scripts

The focus remains on binary distribution excellence while avoiding complexity from cloud infrastructure.