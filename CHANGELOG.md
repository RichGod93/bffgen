# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [3.0.0] - 2025-12-28

### Added
- **Python/FastAPI Template** - Complete built-in template with async REST aggregation
  - FastAPI framework support with async/await patterns
  - Pydantic models for type-safe request/response validation
  - Parallel request aggregation for optimal performance
  - Auto-generated API documentation (Swagger/ReDoc)
  - Docker containerization support
  - Health check endpoints for Kubernetes
- **Template Management System**
  - Template installer for GitHub-based templates (`bffgen template install <url>`)
  - Template manager with caching and registry support
  - Template update and removal commands
  - Comprehensive template validation
  - Support for community templates via registry
- **Enhanced TUI (Terminal UI)**
  - Interactive template selector with fuzzy search
  - Test type selector with descriptions
  - Progress visualization with animated gradients
  - Enhanced error and message display
- **Watch Mode**
  - Auto-regeneration on configuration changes
  - Smart file watching with debouncing
  - Diff preview before applying changes
  - Automatic rollback on errors
- **Test Generation Improvements**
  - Comprehensive unit test generation
  - Integration test scaffolding
  - Mock generators for all supported languages
  - Test fixtures and helpers

### Changed
- Reorganized template directory structure (flattened `built-in/` to root level)
- Updated template loading logic to support bundled and community templates
- Enhanced scaffolding engine with better variable substitution
- Improved error messages and user feedback

### Fixed
- Template variable syntax consistency across all templates
- Go template variable format ({{PROJECT_NAME}} → consistent usage)
- Scaffolder test compilation errors
- Template loading from nested directories

### Tests
- Added comprehensive test suite for template installer (7 test cases)
- Added comprehensive test suite for scaffolder (7 test cases)
- Total test coverage: 31 tests for template system
- All tests passing ✅

### Documentation
- Added `CONTRIBUTING_TEMPLATES.md` for template creation guide
- Added `GRAPHQL_SUPPORT.md` for GraphQL implementation details
- Added `TEMPLATE_SYSTEM_DESIGN.md` for template system architecture
- Updated `.gitignore` to exclude internal documentation

## [2.2.1] - 2024-12-XX

### Fixed
- Linter errors and warnings
- Code formatting issues
- Pointer dereference warnings in tests

## [2.2.0] - 2024-XX-XX

### Added
- Python FastAPI runtime support (initial)
- Enhanced scaffolding capabilities

## [2.1.0] - 2024-XX-XX

### Added
- Additional runtime support
- Infrastructure improvements

## [2.0.1] - 2024-XX-XX

### Fixed
- Bug fixes and stability improvements

## [2.0.0] - 2024-XX-XX

### Added
- Major version release
- Complete rewrite of core functionality

---

[3.0.0]: https://github.com/RichGod93/bffgen/compare/v2.2.1...v3.0.0
[2.2.1]: https://github.com/RichGod93/bffgen/compare/v2.2.0...v2.2.1
[2.2.0]: https://github.com/RichGod93/bffgen/compare/v2.1.0...v2.2.0
[2.1.0]: https://github.com/RichGod93/bffgen/compare/v2.0.1...v2.1.0
[2.0.1]: https://github.com/RichGod93/bffgen/compare/v2.0.0...v2.0.1
[2.0.0]: https://github.com/RichGod93/bffgen/releases/tag/v2.0.0
