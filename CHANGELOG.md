# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [3.4.0] - 2026-10-15

Generated Express and Fastify routes are now real proxies: they forward path parameters and
query strings, time out, retry safely, stop calling failing backends, and tell clients what
went wrong. See [Generated Node.js Routes](docs/NODEJS_ROUTES.md).

### Added
- **Timeouts, retries and a circuit breaker per backend** in generated routes. Configure them
  per backend with `timeout`, `retries` and `circuitBreaker` in `bffgen.config.json` (defaults
  10 s, 2 retries, breaker opening after 5 consecutive failures for 30 s), or override them with
  `<NAME>_TIMEOUT`, `<NAME>_RETRIES`, `<NAME>_CIRCUIT_FAILURE_THRESHOLD` and
  `<NAME>_CIRCUIT_RESET_TIMEOUT`. Only `GET`, `HEAD` and `OPTIONS` requests are retried.
- **Gateway status codes for backend failures**: `504` when a backend times out, `503` when it
  is unreachable or its circuit is open (with `Retry-After`), `502` when it returns a 5xx. The
  body is `{ error, message, code, service }`; a backend's 5xx body is no longer passed on.
- `src/utils/upstream.js`, the runtime that generated routes use. `bffgen generate` adds it to
  existing projects and upgrades outdated `httpClient.js` and `circuit-breaker.js` files,
  keeping your previous copies as `.bak`.

### Changed
- In `httpClient.js`, `retries` now means extra attempts (it was total attempts), and the
  default timeout is 10 s (was 30 s). Failures throw `UpstreamError`, and 4xx responses from
  `request()` throw `HttpError` with the backend's status and body.
- In `circuit-breaker.js`, the half-open state admits a single trial request (it admitted every
  request), 4xx errors no longer count as failures, and the default `resetTimeout` is 30 s
  (was 60 s). Open-circuit rejections throw `CircuitOpenError`.
- Generated services report the backend's 4xx status, or 502/503/504, instead of 500.
- Backend environment variable prefixes replace characters other than letters and digits with
  `_` (`user-service` → `USER_SERVICE_URL`). `<NAME>_SERVICE_URL`, which `init` writes to
  `.env.example`, is now honored too.

### Fixed
- **Path parameters were not substituted**: `/api/users/42` called the backend at
  `/users/:id`. Parameters are now substituted and URL-encoded; `{id}` style is accepted too.
- **Query strings were dropped.** They are forwarded unchanged.
- **Backend 4xx responses** such as `404` reached clients with a bffgen error body; they are now
  passed through as is, along with `content-type`, `cache-control`, `etag`, `last-modified` and
  `location` headers.
- Non-JSON and empty (`204`) backend responses made generated routes fail with 500.
- `generate` now reports endpoints whose backend `path` uses a parameter missing from
  `exposeAs`, or whose paths contain characters that would break the generated code.

### Documentation
- README: the version badge now tracks npm, and the `go install` instructions, which installed an
  old v1.x release or failed, are replaced with a note explaining why and how to install instead.
- `docs/REGENERATION_SAFE_SCAFFOLDING.md` now states which Node.js files `generate` overwrites.
- Removed unverified time-savings figures from the docs, and fixed broken documentation links.

## [3.3.0] - 2026-10-11

Generated Express and Fastify projects now pass `npm run lint` and `npm test` out of the box.

### Added
- **ESLint configuration for generated Node.js projects** - `bffgen init` now writes an
  `eslint.config.js` (ESLint 9 flat config, `@eslint/js` recommended rules, Node and Jest globals).
  Previously `npm run lint` failed because no config file existed. `lint` now checks the whole
  project (`eslint .`), including tests.

### Changed
- Generated Express/Fastify projects use **ESLint 9** (was ESLint 8, which is end-of-life).
- Generated Fastify `src/index.js` exports `buildApp()`, which returns a configured Fastify instance;
  the server only starts listening when the file is run directly. Express `src/index.js` likewise only
  calls `listen()` when run directly. This lets tests import the app without binding a port.
- Generated controller/service handler names are now camelCase identifiers derived from the route,
  e.g. `GET /api/users/:id` → `getApiUsersById` (was `getapiusers:id`).

### Fixed
- **Generated dependency versions were empty** - `bffgen init` for Express and Fastify wrote every
  dependency in `package.json` as `""`, so npm installed the latest major versions (e.g. Express 5,
  ESLint 10) instead of the versions in `internal/registry/versions.yaml`. Versions are now loaded
  from the manifest.
- **Python `requirements.txt` was not rendered** - FastAPI projects contained literal
  `{{GetVersion "..."}}` lines, which broke `pip install -r requirements.txt`. It is now rendered
  with the manifest versions.
- `GetVersion` in templates now fails generation when a package is missing from the manifest,
  instead of silently emitting an empty version.
- **Syntax errors in generated controllers and services** - any endpoint with a path parameter
  (e.g. `/users/:id`) produced an invalid method name, so the files could not be loaded.
- **Generated GET/HEAD routes failed with Express 4** - routes forwarded `req.body` (`{}` under
  `express.json()`) on GET requests, and `fetch` rejects GET/HEAD requests with a body. Bodies are now
  only forwarded for methods that accept one.
- **Fastify JWT authentication returned 500** - `src/index.js` requires `jsonwebtoken`, but it was
  missing from the generated `package.json`.
- **Generated Jest setup did not run** - `jest.config.js` referenced the uninstalled
  `jest-watch-typeahead` plugin and misspelled `coverageThreshold`; the Express health test imported
  the app from the wrong path; Fastify's ESM-only `cookie@2` dependency could not be loaded by Jest
  (now transformed via `@babel/plugin-transform-modules-commonjs`).
- **Fastify health tests asserted nothing** - every assertion was commented out. They now exercise
  `/health`, `/ready`, and the 404 handler via `app.inject()`.
- Lint errors in generated utilities (`no-prototype-builtins`, unused import, useless try/catch).

### Security
- `bffgen template install` now passes `--` before the repository URL to `git clone`, so a URL
  beginning with `-` cannot be interpreted as a git option.

### Internal
- CI: upgrade golangci-lint to v1.64.8 (v1.63.4 could not lint a Go 1.24 module, so CI had been
  failing since v3.0.0), pin gosec to v2.29.0, exclude non-compilable `templates/` from gosec, and fix the pre-existing
  lint findings (unchecked errors, dead code, ineffectual assignments).

## [3.2.0] - 2026-01-10

### Added
- **Dynamic Dependency Management System** - Major architectural improvement
  - Central version manifest (`internal/registry/versions.yaml`) - single source of truth for 50+ dependencies
  - NPM registry client for fetching latest package versions from npmjs.org
  - PyPI registry client for Python package versions from pypi.org
  - 24-hour disk cache for fetched versions (`~/.bffgen/cache/versions/`)
  - New `bffgen update-deps` CLI command to update dependency versions
    - `--dry-run` flag to preview updates
    - `--runtime` flag to update specific runtime only
  - Template function `GetVersion` for dynamic version substitution
  - Concurrent package fetching for improved performance

### Changed
- Converted Express templates to use dynamic versions from manifest
- Converted Fastify templates to use dynamic versions from manifest
- Converted Python FastAPI templates to use dynamic versions from manifest
- Template rendering now supports custom function maps
- Dependency versions no longer hardcoded in templates

### Improved
- Eliminated need to rebuild binary for dependency updates
- Reduced maintenance burden - one YAML file instead of multiple templates
- Foundation for automated CI/CD dependency updates
- Faster iteration on dependency changes

### Technical Details
- New packages: `internal/registry/` with manifest loader, NPM client, PyPI client, and caching
- New file: `internal/templates/versions.go` with template helper functions
- Modified: `internal/templates/loader.go` to support `GetVersion` function map
- ~850 lines of new code

## [3.1.0] - 2026-01-10

### Fixed
- **CI/CD Version Mismatch** - Updated release workflow to use Go 1.24 (matching go.mod)
- **Template Path Errors** - Corrected all Node.js runtime template paths from `nodejs-*` to `node/*`
- **Incomplete GraphQL Support** - Removed nodejs-yoga runtime (incomplete implementation)

### Added
- **Windows ARM64 Support** - Added to both CI/CD builds and NPM package
  - Surface Pro X and other Windows on ARM devices now supported
  - Binary releases include windows-arm64 target
  - NPM installation works on Windows ARM64

###Changed
- **Updated Node.js Version Requirement** - Minimum version now 18.0.0 (from 14.0.0)
  - Aligns with modern LTS versions
  - Drops EOL Node.js 14 support
- **Updated Dependencies to Latest Stable Versions**:
  - **Express**: 4.19.2, Helmet 8.0.0, Winston 3.15.0, dotenv 16.4.5
  - **Fastify**: 5.1.0 (major upgrade), @fastify/* plugins updated to compatible versions
  - **Python**: FastAPI 0.115.0+, Uvicorn 0.30.0+, Pydantic 2.9.0+
- Deprecated hardcoded dependency generation functions in `language_generators.go`
- Updated README to accurately reflect supported runtimes (removed nodejs-yoga mentions)

### Removed
- nodejs-yoga GraphQL runtime support (incomplete, misleading users)
- GRAPHQL_SUPPORT.md (moved to docs/ for clarity)

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

[3.2.0]: https://github.com/RichGod93/bffgen/compare/v3.1.0...v3.2.0
[3.1.0]: https://github.com/RichGod93/bffgen/compare/v3.0.0...v3.1.0
[3.0.0]: https://github.com/RichGod93/bffgen/compare/v2.2.1...v3.0.0
[2.2.1]: https://github.com/RichGod93/bffgen/compare/v2.2.0...v2.2.1
[2.2.0]: https://github.com/RichGod93/bffgen/compare/v2.1.0...v2.2.0
[2.1.0]: https://github.com/RichGod93/bffgen/compare/v2.0.1...v2.1.0
[2.0.1]: https://github.com/RichGod93/bffgen/compare/v2.0.0...v2.0.1
[2.0.0]: https://github.com/RichGod93/bffgen/releases/tag/v2.0.0
