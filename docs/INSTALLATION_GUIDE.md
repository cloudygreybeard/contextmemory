# ContextMemory Installation Guide

## Quick Start

The fastest way to get ContextMemory running with Cursor IDE:

```bash
# Download and install (macOS/Linux)
curl -sSL https://install.contextmemory.dev | bash

# Configure Cursor automatically  
contextmemory config cursor

# Test the installation
contextmemory test
```

## Manual Installation

### Step 1: Download Binary

#### macOS (Apple Silicon)
```bash
curl -L https://github.com/contextmemory/contextmemory/releases/latest/download/contextmemory-darwin-arm64.tar.gz | tar xz
```

#### macOS (Intel)
```bash
curl -L https://github.com/contextmemory/contextmemory/releases/latest/download/contextmemory-darwin-amd64.tar.gz | tar xz
```

#### Linux (x64)
```bash
curl -L https://github.com/contextmemory/contextmemory/releases/latest/download/contextmemory-linux-amd64.tar.gz | tar xz
```

#### Windows (x64)
```powershell
Invoke-WebRequest -Uri "https://github.com/contextmemory/contextmemory/releases/latest/download/contextmemory-windows-amd64.zip" -OutFile "contextmemory.zip"
Expand-Archive -Path "contextmemory.zip" -DestinationPath "."
```

### Step 2: Install to System Path

#### macOS/Linux
```bash
# Make executable
chmod +x contextmemory

# Install to system path
sudo mv contextmemory /usr/local/bin/

# Verify installation
contextmemory --version
```

#### Windows
```powershell
# Create program directory
New-Item -ItemType Directory -Force -Path "C:\Program Files\ContextMemory"

# Move binary
Move-Item -Path "contextmemory.exe" -Destination "C:\Program Files\ContextMemory\"

# Add to PATH (requires admin)
$env:PATH += ";C:\Program Files\ContextMemory"
```

### Step 3: Initialize Configuration

```bash
# Initialize ContextMemory with default settings
contextmemory init

# This creates:
# ~/.contextmemory/config.json
# ~/.contextmemory/storage/
# ~/.contextmemory/logs/
```

## IDE Configuration

### Cursor IDE

#### Automatic Configuration
```bash
contextmemory config cursor
```

This automatically creates `~/.cursor/mcp.json`:

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

#### Manual Configuration

1. **Create or edit** `~/.cursor/mcp.json`
2. **Add the ContextMemory server configuration**:

```json
{
  "mcpServers": {
    "contextmemory": {
      "command": "/usr/local/bin/contextmemory",
      "args": ["mcp-server"],
      "env": {
        "CONTEXTMEMORY_HOME": "~/.contextmemory"
      }
    }
  }
}
```

3. **Restart Cursor IDE**
4. **Verify** by typing `@contextmemory` in a chat

### VS Code

#### Automatic Configuration
```bash
contextmemory config vscode
```

#### Manual Configuration

Create `.vscode/mcp.json` in your workspace:

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

### Other IDEs

ContextMemory can work with any IDE that supports the MCP protocol:

```json
{
  "mcpServers": {
    "contextmemory": {
      "command": "/usr/local/bin/contextmemory",
      "args": ["mcp-server", "--ide=generic"],
      "stdio": true
    }
  }
}
```

## Package Manager Installation

### Homebrew (macOS)

```bash
# Add tap
brew tap contextmemory/tap

# Install
brew install contextmemory

# Configure Cursor
contextmemory config cursor
```

### APT (Ubuntu/Debian)

```bash
# Add repository
curl -fsSL https://pkg.contextmemory.dev/gpg | sudo apt-key add -
echo "deb https://pkg.contextmemory.dev/apt stable main" | sudo tee /etc/apt/sources.list.d/contextmemory.list

# Install
sudo apt update
sudo apt install contextmemory

# Configure
contextmemory config cursor
```

### YUM/DNF (RHEL/Fedora)

```bash
# Add repository
sudo rpm --import https://pkg.contextmemory.dev/gpg
sudo tee /etc/yum.repos.d/contextmemory.repo <<EOF
[contextmemory]
name=ContextMemory Repository
baseurl=https://pkg.contextmemory.dev/rpm
enabled=1
gpgcheck=1
gpgkey=https://pkg.contextmemory.dev/gpg
EOF

# Install
sudo dnf install contextmemory  # or sudo yum install contextmemory

# Configure
contextmemory config cursor
```

### Chocolatey (Windows)

```powershell
# Install Chocolatey if not already installed
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

# Install ContextMemory
choco install contextmemory

# Configure
contextmemory config cursor
```

### Snap (Linux)

```bash
# Install
sudo snap install contextmemory

# Configure
contextmemory config cursor
```

## Verification & Testing

### Test Installation

```bash
# Check version
contextmemory --version

# Test MCP server
contextmemory test

# Check configuration
contextmemory config show
```

### Test IDE Integration

1. **Restart your IDE** after configuration
2. **Open a chat/conversation**
3. **Type** `@contextmemory` and verify tools appear:
   - `@contextmemory get`
   - `@contextmemory create` 
   - `@contextmemory checkpoint`
   - `@contextmemory monitor`
   - `@contextmemory snapshot`
   - `@contextmemory thread`

4. **Test a simple command**:
   ```
   @contextmemory get --output-format table --limit 5
   ```

### Test Automation Features

```bash
# Test monitoring
@contextmemory monitor --conversation-id "test_conversation" --message-count 10 --estimated-tokens 5000

# Expected output:
# Status: healthy
# Token Utilization: 5%
# Recommendation: Continue conversation - context window healthy

# Test snapshot creation
@contextmemory snapshot --conversation-id "test_conversation" --recent-content "This is a test snapshot" --token-count 5000 --messages-since 5

# Test checkpoint creation  
@contextmemory checkpoint --conversation-id "test_conversation" --title "Test Checkpoint" --content "This is a test checkpoint"
```

## Configuration Options

### Global Configuration (`~/.contextmemory/config.json`)

```json
{
  "version": "0.7.0",
  "storage": {
    "type": "file",
    "path": "~/.contextmemory/storage",
    "backup_enabled": true,
    "backup_interval": "24h"
  },
  "monitoring": {
    "max_tokens": 100000,
    "warning_threshold": 0.8,
    "critical_threshold": 0.95,
    "snapshot_interval": 5,
    "checkpoint_interval": 20
  },
  "logging": {
    "level": "info",
    "file": "~/.contextmemory/logs/contextmemory.log",
    "max_size": "10MB",
    "max_files": 5
  },
  "server": {
    "timeout": "30s",
    "buffer_size": 1024
  }
}
```

### Project-Specific Configuration (`.contextmemory/config.json`)

```json
{
  "project": "example-project",
  "namespace": "project-memories",
  "monitoring": {
    "snapshot_interval": 3,
    "checkpoint_interval": 15,
    "warning_threshold": 0.7
  },
  "labels": {
    "project": "example-project",
    "team": "engineering",
    "environment": "development"
  }
}
```

## Troubleshooting

### Common Issues

#### 1. "contextmemory command not found"

**Solution:**
```bash
# Check if binary exists
ls -la /usr/local/bin/contextmemory

# Check PATH
echo $PATH

# Re-add to PATH if needed
export PATH="/usr/local/bin:$PATH"
echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bashrc
```

#### 2. "Permission denied" when running contextmemory

**Solution:**
```bash
# Make executable
chmod +x /usr/local/bin/contextmemory

# Check ownership
ls -la /usr/local/bin/contextmemory

# Fix ownership if needed
sudo chown $(whoami):$(whoami) /usr/local/bin/contextmemory
```

#### 3. IDE doesn't recognize @contextmemory commands

**Solution:**
1. **Verify MCP configuration**:
   ```bash
   contextmemory config show
   ```

2. **Check IDE configuration file**:
   ```bash
   # Cursor
   cat ~/.cursor/mcp.json
   
   # VS Code  
   cat .vscode/mcp.json
   ```

3. **Restart IDE completely**

4. **Check logs**:
   ```bash
   contextmemory logs
   ```

#### 4. "Failed to connect to MCP server"

**Solution:**
1. **Test server directly**:
   ```bash
   contextmemory mcp-server --test
   ```

2. **Check server logs**:
   ```bash
   tail -f ~/.contextmemory/logs/contextmemory.log
   ```

3. **Verify configuration**:
   ```bash
   contextmemory config validate
   ```

### Getting Help

#### Debug Mode
```bash
# Enable debug logging
export CONTEXTMEMORY_LOG_LEVEL=debug
contextmemory mcp-server

# Or modify config.json
{
  "logging": {
    "level": "debug"
  }
}
```

#### Support Channels
- **Documentation**: https://docs.contextmemory.dev
- **GitHub Issues**: https://github.com/contextmemory/contextmemory/issues
- **Discord Community**: https://discord.gg/contextmemory
- **Email Support**: support@contextmemory.dev

## Updating

### Automatic Updates (Recommended)
```bash
contextmemory update
```

### Manual Updates
```bash
# Download latest version
curl -sSL https://install.contextmemory.dev | bash

# Or with package managers
brew upgrade contextmemory       # Homebrew
sudo apt update && sudo apt upgrade contextmemory  # APT
choco upgrade contextmemory      # Chocolatey
```

### Backup Before Update
```bash
# Backup configuration and data
contextmemory backup --output ~/contextmemory-backup-$(date +%Y%m%d).tar.gz

# Restore if needed
contextmemory restore ~/contextmemory-backup-20251229.tar.gz
```

## Next Steps

After successful installation:

1. Read the [Automation System Guide](AUTOMATION_SYSTEM.md) to understand automation features
2. Try the [Getting Started Tutorial](GETTING_STARTED.md) with real conversations  
3. Customize your [Configuration](CONFIGURATION.md) for your workflow
4. Join the [Community](https://discord.gg/contextmemory) for tips and support

## Development Installation

### Build from Source
```bash
# Clone repository
git clone https://github.com/contextmemory/contextmemory.git
cd contextmemory

# Build MCP server
go build -o contextmemory ./cmd/mcp-server-sdk/

# Install locally
sudo mv contextmemory /usr/local/bin/
```

### Testing Your Build
```bash
# Verify installation
contextmemory version

# Test configuration
contextmemory config cursor

# Run integration tests
contextmemory test
```

This installation guide covers all major scenarios for getting ContextMemory up and running reliably.
