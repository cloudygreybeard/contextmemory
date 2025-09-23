# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.6.3] - 2025-09-22

### Fixed
- Chat title extraction now preserves actual user-set titles from Cursor AI pane
- Improved correlation between composer data and aiService prompts for accurate title matching
- Title preservation is now verbatim without unwanted capitalization changes

### Changed
- Enhanced parseAIServicePromptsWithTitles function to use composer title mapping
- Updated GetChatData to pre-extract composer titles for correlation

## [0.6.2] - 2025-09-22

### Fixed
- Fixed Cursor chat capture in VS Code extension to directly access AI pane data
- Extension no longer requires markdown file to be open for chat capture
- Added proper progress indicators and error handling for chat import

### Enhanced
- Improved user experience with better error messages and retry options
- Added direct integration with `cmctl import-cursor-chat` command

## [0.6.1] - 2025-09-22

### Added
- **Cursor AI Chat Import**: Direct integration with Cursor's AI pane conversations
  - `cmctl import-cursor-chat` command for importing chat sessions
  - `cmctl list-cursor-chats` command for browsing available conversations
  - Support for multiple Cursor data formats (`aiService.prompts`, `composer.composerData`)
  - Intelligent conversation analysis with technology detection and activity classification
- **Pure Go SQLite Integration**: GORM with `github.com/glebarez/sqlite` for database access
  - No CGO dependencies for improved cross-platform compatibility
  - WebAssembly compilation support for embedded extension usage
- **Enhanced AI Analysis**: Improved conversation parsing and labeling
  - Programming language detection from chat content
  - Activity classification (debugging, implementation, learning, etc.)
  - Smart naming based on conversation topics
- **WebAssembly CLI**: Optional embedding of Go CLI as WebAssembly in VS Code extension
  - 13MB `cmctl.wasm` with full CLI functionality
  - Direct function calls without subprocess overhead
- **Multi-Architecture CI/CD**: GoReleaser configuration for automated builds
  - Linux, macOS, Windows support (x86_64, aarch64)
  - WebAssembly target included in build matrix
  - GitHub Actions workflows for testing and releases

### Changed
- Simplified VS Code extension UI to focus on chat capture workflow
- Updated CLI commands to use unified `get` command (deprecating separate `list`)
- Enhanced output formatting with better error handling
- Improved version compatibility checking between CLI and extension

### Technical
- Database access uses GORM with structured models for type safety
- Read-only access to Cursor's workspace storage at `~/Library/Application Support/Cursor/User/workspaceStorage/`
- JSON-based memory storage with automated indexing
- Cross-platform build pipeline supporting traditional platforms and WebAssembly

### Documentation
- Added `TECHNICAL_INNOVATIONS.md` documenting architecture choices
- Updated installation and usage instructions
- Improved CLI help text and examples

## [0.6.0] - 2025-09-21

### Added
- Initial release with file-based session context management
- CLI tool (`cmctl`) with create, get, list, search, delete operations
- VS Code extension with tree view and commands
- Label-based organization system
- Multiple output formats (JSON, YAML, JSONPath, Go templates)
- Version compatibility checking between CLI and extension

### Features
- File-based JSON storage
- Advanced search capabilities
- Flexible labeling system
- Cross-platform CLI distribution
- VS Code integration with tree view

[0.6.1]: https://github.com/cloudygreybeard/contextmemory/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/cloudygreybeard/contextmemory/releases/tag/v0.6.0
