# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [3.9.0] - 2026-10-21

Config-driven aggregation routes for Express, Fastify and FastAPI. See
[Aggregates](docs/NODEJS_ROUTES.md#aggregates).

### Added
- **`"aggregates"`**: each entry is a `GET` route that calls several backends in parallel and
  answers `{ "data": { <as>: … }, "sources": { <as>: "ok" | <error code> } }`. Sources use their
  backend's timeout, retries, circuit breaker and environment overrides, and may carry a query
  string (`/orders?userId=:id`). An optional source that fails leaves its data `null` (partial
  response, `Cache-Control: no-store`); a failed `"required": true` source, or every source
  failing, fails the request like a proxied route (502/503/504, or the source's 4xx passed
  through). `requiresAuth` and `roles` work as on endpoints. Routes are generated into
  `src/routes/aggregates.js` (Python: `routers/aggregates_router.py`), which falls under the same
  regeneration rules as other generated files.
- `config validate`, `generate-docs` (OpenAPI) and the JSON Schema cover aggregates. A backend used
  only by aggregates no longer needs `endpoints`. `convert` to Go warns that aggregates are not
  converted.

### Changed
- The runtime file `src/utils/upstream.js` (Python: `utils/upstream.py`) gains `aggregate()`
  (`bffgen:runtime 2`). `bffgen generate` upgrades it in existing projects and keeps the previous
  copy as `upstream.js.bak`. Runtime files are now versioned one by one, so the other runtime
  files are left alone.

### Fixed
- `ParallelAggregator` (`src/utils/aggregator.js`, new projects): `failFast: true` now rejects with
  the first error (it used to resolve like `failFast: false`), and per-call timeout timers are
  cleared when a call finishes (they used to keep running, holding the process open and
  accumulating under load).

## [3.8.0] - 2026-10-21

Role-based access control for generated routes (Express, Fastify and FastAPI). See
[Authentication and roles](docs/NODEJS_ROUTES.md#authentication-and-roles-rbac).

### Added
- **`"roles"` on endpoints**: `"roles": ["admin", "support"]` requires a valid JWT with at least one
  of the roles; other users get `403 Forbidden` (`INSUFFICIENT_ROLE`; the required roles aren't
  revealed). `roles` implies `requiresAuth`. Roles are read from the `roles` claim (or `role`),
  or from the claim named by `JWT_ROLES_CLAIM`, which also accepts dotted paths such as
  `realm_access.roles`. Generated routes use the new runtime file `src/middleware/authorize.js`
  (Python: `utils/authorize.py`, `require_roles(...)`), which `bffgen generate` adds to existing
  projects. If you already have a file of your own at that path, it is kept; `generate` then
  refuses to generate endpoints with `roles` until you rename it.
- `bffgen config validate` checks `roles`; `generate-docs` adds a 403 response to endpoints with
  roles; `convert` to Go warns that endpoints with `requiresAuth` or `roles` are unprotected in
  the Go project (Go configs have neither).
- **`schemas/bffgen-config.json`**: a JSON Schema for `bffgen.config.json` and
  `bffgen.config.py.json`. New configs reference it in `$schema`, so editors validate and
  autocomplete them.

### Changed
- **Invalid or expired tokens get `401` instead of `403`** in Fastify routes (and in Express's
  `/api/auth/profile`), as RFC 9110 defines them: 401 = not authenticated, 403 = authenticated but
  not allowed. New projects are generated that way. In existing projects that use `roles`,
  `bffgen generate` changes that one line of `src/index.js`, saves the previous copy as
  `src/index.js.bak` and reports it (`--check` lists it as pending); projects without `roles`
  are left alone.
- 401 responses in new Node.js projects include `WWW-Authenticate: Bearer`, which RFC 9110
  requires (FastAPI projects already sent it). `authorize.js` sends it in existing projects too;
  for the other 401s, add `reply.header('WWW-Authenticate', 'Bearer')` (Express:
  `res.set(...)`) before them in `src/index.js` and `src/middleware/auth.js`.

### Fixed
- **Express returned 500 for a token that isn't valid yet** (`nbf` in the future); it now
  returns 401, like any other invalid token. Existing projects: in `src/middleware/auth.js`,
  replace the two `error.name === …` checks with `error instanceof jwt.JsonWebTokenError`.
- **`bffgen config validate` didn't work in Python projects**: it validated them as Go projects
  and failed looking for `bff.config.yaml`. It now checks `bffgen.config.py.json` with the same
  rules as `bffgen generate`.
- New Node.js configs referenced the schema `https://bffgen.dev/schema.json`, a domain that
  doesn't exist; they now reference `schemas/bffgen-config.json` in this repository.

## [3.7.0] - 2026-10-21

Regenerating is now safe: files you edit are kept, `--dry-run` writes nothing, and the same
config always produces the same output. See [Regenerating](docs/NODEJS_ROUTES.md#regenerating).

### Changed
- **`bffgen generate` no longer overwrites files you edited** (Node.js and Python). Generated
  route, controller and service files (Python: routers, services, `routers/__init__.py`) are
  regenerated only if they still have the content bffgen last wrote; edited files are kept and
  reported; emptying a file counts as an edit. `--force` regenerates them too and saves your
  version as `<file>.bak` (`<file>.bak.1`, `.bak.2`, … if earlier backups exist, so none is
  overwritten; runtime-file upgrades do the same). Line endings don't count as edits, so Windows
  checkouts with `core.autocrlf=true` regenerate normally. What bffgen wrote is recorded in
  `.bffgen/generated.json`; commit it. The generated `.gitignore` now ignores `.bffgen/*` except
  that file, and `bffgen generate` fixes the `.bffgen/` line in Python projects made before 3.7.
- **An invalid config writes nothing.** Node.js used to skip the backend with a warning and write
  the rest; it now stops with an error, as Python already did. Every write is planned first, so
  other errors (such as a corrupt `.bffgen/generated.json`) write nothing either. Backend names
  must contain only letters, digits, `-` and `_` (they become file names and `require()` paths)
  and must be unique, also when they differ only in case or in `-` vs `_` (`users`/`Users`,
  `user-service`/`user_service`).
- `bffgen generate --check` exits with status 1 when generated code is out of date, so CI can
  catch a config that wasn't regenerated. That includes an edited file the config now generates
  differently, and (Python) a `main.py` that doesn't register the routers yet.
- Files for a backend removed from the config are reported, not deleted.

### Fixed
- **`--dry-run` and `--check` wrote route, controller and service files** in Node.js projects
  (overwriting custom code). They now write nothing.
- `--force` was ignored in Node.js projects.
- Two backends with the same name registered the same routes twice in `src/index.js`, so Fastify
  projects failed to start.
- "Auto-registered routes in src/index.js" was printed even when `src/index.js` has no
  `bffgen:begin:routes` markers and nothing was registered.

## [3.6.0] - 2026-10-21

Security update. Newly generated projects no longer install dependencies with known
vulnerabilities, and `go install` now gets the current release.

### Security
- **Fastify projects shipped `@fastify/jwt` 9, whose `fast-jwt` dependency has critical
  advisories** (including authentication bypasses). Generated projects now use `@fastify/jwt`
  ^10.2.5 and `@fastify/swagger-ui` ^6.1.2. `npm audit --omit=dev` is clean for fresh Express and
  Fastify projects. Existing projects: `npm install @fastify/jwt@^10.2.5 @fastify/swagger-ui@^6.1.2`.
- **FastAPI projects used `python-jose`**, which is unmaintained and has unfixed advisories (as
  does its `ecdsa` dependency). Generated projects now verify JWTs with **PyJWT**; `passlib`
  (unused) is gone. FastAPI is now `>=0.150` with Starlette `>=1.3.1`, which fixes the Starlette
  advisories. `pip-audit` is clean for fresh projects, dev tools included. Existing projects:
  replace `python-jose[cryptography]` with `pyjwt` in `requirements.txt`, and in
  `middleware/auth.py` and `tests/conftest.py` change `from jose import jwt, JWTError` to
  `import jwt` and `except JWTError` to `except jwt.InvalidTokenError`.
- Generated Go projects pin chi v5.5.0 (chi < 5.3.0 has IP-spoofing advisories in
  `middleware.RealIP`), echo v4.16.0, fiber v2.52.15 and golang-jwt v5.3.1, with `go 1.25.0`.
  bffgen itself now builds with chi v5.3.0.
- Release binaries are built with the newest stable Go instead of Go 1.24, which no longer
  receives security fixes.

### Changed
- **Generated FastAPI projects need Python 3.10+** (3.9 is end-of-life, and patched Starlette needs
  3.10). Dependency ranges were refreshed (pytest 9, pytest-asyncio 1, ruff, black). `pyproject.toml`
  (Poetry) now uses the same versions as `requirements.txt`; it had its own, older ones.
- Generated FastAPI projects get a `ruff.toml` with a fixed rule set, so `ruff check` gives the same
  result on every Ruff release.
- Generated Express and Fastify projects use Jest 30 and Node's built-in `node --watch` for
  `npm run dev` instead of nodemon (one dependency fewer, no advisories). Test-filter scripts use
  Jest 30's `--testPathPatterns`.
- The Go module path is now `github.com/RichGod93/bffgen/v3`, so
  `go install github.com/RichGod93/bffgen/v3/cmd/bffgen@latest` installs the current release (it
  used to install v1.3.0). `bffgen --version` reports the module version for `go install` builds
  instead of `dev`.

### Fixed
- The generated FastAPI `setup.sh` checked for Python 3.9 while saying 3.10 was required.
- `init` no longer writes a throwaway `package.json` with outdated versions before rendering the real one.

### Removed
- Unused Python GraphQL templates (never generated, pinned outdated packages).

### Project
- Releases are now cut by merging a `release/vX.Y.Z` branch: CI checks the branch name and
  version, and the merge tags the release and publishes the binaries (and, once npm trusted
  publishing is enabled, the npm package), then smoke-tests them. See [Releasing](docs/RELEASING.md).
- CI generates Express, Fastify and FastAPI projects on every pull request and runs their lint,
  tests, a dependency audit and the end-to-end checks in `test/e2e/`. It also runs govulncheck.

## [3.5.0] - 2026-10-15

Python (FastAPI) projects now work, and their generated routes match the Express and Fastify
ones: path parameters, query strings, auth, timeouts, retries, a circuit breaker per backend,
and 504/503/502 on backend failure. See [Python/FastAPI Support](docs/PYTHON_SUPPORT.md#what-generated-routes-do).

### Fixed
- **Fresh FastAPI projects crashed at startup**: `.env` contained `DEBUG=true`, which the
  generated settings rejected as an unknown key. `config.py` now ignores unknown `.env` keys.
  Existing projects: add `extra = "ignore"` to the `Config` class in `config.py`
  (`bffgen generate` prints this hint).
- **Generated FastAPI routes could not reach any backend**: the configured `baseUrl` never
  reached the generated service, and `<NAME>_URL` was ignored, so every route returned 500.
- `{id}` in a backend path was sent as Python's built-in `id` function; path parameters are now
  substituted and URL-encoded (`:id` works too). Query strings are forwarded.
- `requiresAuth` was ignored; such routes now require a valid JWT and return 401 otherwise.
- Backend 4xx responses became 500 with the exception text; they are now passed through.
- Regenerating appended the router imports to `main.py` again on every run. Routers are now
  listed in a generated `routers/__init__.py`; older `main.py` files are fixed up in place.
- `main.py` in `async: false` projects called the middleware's `call_next` without `await`.
- `add-template` wrote template services in random order; it now sorts them (Node.js and Python).

### Added
- Per-backend `timeout`, `retries` and `circuitBreaker` settings for Python projects (same
  format as Node.js: times in ms), with `<NAME>_TIMEOUT`, `<NAME>_RETRIES`,
  `<NAME>_CIRCUIT_FAILURE_THRESHOLD` and `<NAME>_CIRCUIT_RESET_TIMEOUT` overrides that can
  also live in `.env`.
- Python runtime files `utils/upstream.py` and `services/http_client.py`; `generate` adds them
  to existing projects and upgrades `utils/circuit_breaker.py`, keeping the old copy as `.bak`.
- Generated service classes with typed methods per endpoint, sharing the routes' client and
  breaker.
- `tests/conftest.py` provides an `auth_headers` fixture with a valid JWT.

### Changed
- Generated FastAPI routes use the full `path` from the config instead of an `/api/<backend>`
  router prefix, and are always async (`project.async` no longer affects them).
- Python configs give times in milliseconds, like Node.js. `generate` rejects `timeout` or
  `resetTimeout` values under 100 ms, which were meant as seconds (`"30s"` also works).
  `circuitBreaker.recoveryTimeout` is accepted as an alias of `resetTimeout`.
- `dependencies.py`: `get_current_user` now returns the JWT payload (it always returned `None`),
  and `require_auth` uses the same JWT check as generated routes.
- `CircuitBreaker` (Python) admits a single trial request when half-open, ignores 4xx errors,
  defaults to a 30 s `timeout`, and raises `CircuitOpenError` (a `CircuitBreakerError`) with
  `retry_after`.

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
