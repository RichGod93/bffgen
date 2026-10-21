# bffgen

**Backend-for-Frontend (BFF) generator.** Scaffold a BFF service in **Node.js (Express or Fastify)**,
**Python (FastAPI)** or **Go (Chi, Echo, Fiber)**, then generate its routes from a config file
listing your backend services.

[![npm version](https://img.shields.io/npm/v/bffgen.svg)](https://www.npmjs.com/package/bffgen)
[![Monthly downloads](https://img.shields.io/npm/dm/bffgen.svg)](https://npm-stat.com/charts.html?package=bffgen)
[![Total downloads](https://img.shields.io/npm/dt/bffgen.svg?label=total%20downloads)](https://npm-stat.com/charts.html?package=bffgen)
[![CI](https://github.com/RichGod93/bffgen/actions/workflows/ci.yml/badge.svg)](https://github.com/RichGod93/bffgen/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](https://github.com/RichGod93/bffgen/blob/master/LICENSE)

## Quick start

```bash
# Node.js (Express or Fastify)
npx bffgen init my-bff --lang nodejs-express      # or --lang nodejs-fastify
cd my-bff && npm install && npm run dev

# Python (FastAPI, Python 3.10+)
npx bffgen init my-bff --lang python-fastapi
cd my-bff && ./setup.sh && source venv/bin/activate && uvicorn main:app --reload

# Go (Chi, Echo or Fiber)
npx bffgen init my-bff --lang go --framework chi
cd my-bff && go run main.go
```

Or install it globally: `npm install -g bffgen`, then run `bffgen init …`.

## Generating routes

List your backends in the project config (`bffgen.config.json` for Node.js,
`bffgen.config.py.json` for Python, `bff.config.yaml` for Go) and run `bffgen generate`:

```json
{
  "backends": [
    {
      "name": "users",
      "baseUrl": "http://localhost:4001/api",
      "timeout": 5000,
      "retries": 2,
      "endpoints": [
        { "name": "getUser", "method": "GET", "path": "/users/:id", "exposeAs": "/api/users/:id", "requiresAuth": true }
      ]
    }
  ]
}
```

Generated Express, Fastify and FastAPI routes:

- forward path parameters, query strings and safe headers to the backend
- use a per-backend timeout, retries for `GET`/`HEAD`/`OPTIONS`, and a circuit breaker
- require a valid JWT (`"requiresAuth": true`, else **401**) and, with `"roles": ["admin"]`, one of
  the listed roles (else **403**); roles come from the token's `roles` claim or one you name in
  `JWT_ROLES_CLAIM`
- answer **504** when a backend times out, **503** when it is unreachable or its circuit is open
  (with `Retry-After` while the circuit is open), and **502** when it returns a 5xx; backend 4xx
  responses are passed through

Re-running `bffgen generate` is safe: files you edited are kept (`--force` regenerates them and
saves your version as `<file>.bak`), the same config always gives the same output, `--dry-run`
writes nothing, and `--check` exits 1 in CI when generated code is out of date.
See [Generated routes](https://github.com/RichGod93/bffgen/blob/master/docs/NODEJS_ROUTES.md).

## What a new Node.js project includes

- JWT authentication middleware, rate limiting, CORS and Helmet security headers
- `/health` endpoint and graceful shutdown
- Structured logging (Winston for Express, Pino for Fastify)
- Swagger UI at `/api-docs`
- ESLint 9 and Jest set up: a fresh project passes `npm run lint` and `npm test`
- Dependencies with no known production vulnerabilities at release time (`npm audit --omit=dev`)
- A parallel aggregation helper (`src/utils/aggregator.js`) and an example aggregator controller
  for combining several backends in one endpoint

## Commands

| Command | Description |
|---|---|
| `bffgen init <name>` | Create a project (`--lang nodejs-express \| nodejs-fastify \| python-fastapi \| go`) |
| `bffgen generate` | Generate routes, controllers and services from the config (`--dry-run`, `--check`, `--force`) |
| `bffgen add-route` | Add a backend endpoint interactively |
| `bffgen add-template` | Add a predefined template (auth, ecommerce, content, mobile-api, gateway) |
| `bffgen add-infra` | Add CI, Docker, docker-compose or health-check files to a project |
| `bffgen config validate` | Check the project config |
| `bffgen convert` | Convert a config between YAML and JSON |
| `bffgen generate-docs` | Generate OpenAPI documentation from the config |
| `bffgen postman` | Generate a Postman collection |
| `bffgen doctor` | Diagnose common project problems |
| `bffgen version` | Print the version |

Run `bffgen <command> --help` for all options.

## Programmatic API

```javascript
const bffgen = require("bffgen");

await bffgen.init({ name: "my-bff", lang: "nodejs-express" });
await bffgen.generate({ dryRun: true }); // or { check: true }
console.log(bffgen.getVersion());
```

## Installation details

The package downloads the prebuilt `bffgen` binary for your platform from the matching
[GitHub release](https://github.com/RichGod93/bffgen/releases) and verifies its SHA-256 checksum.
Supported: macOS, Linux and Windows, each on x64 and arm64. Node.js 18+ is required.

If the download fails (for example behind a proxy), download the binary from the release page, or
install with Go: `go install github.com/RichGod93/bffgen/v3/cmd/bffgen@latest`.

## Links

- [Documentation and full README](https://github.com/RichGod93/bffgen#readme)
- [Changelog](https://github.com/RichGod93/bffgen/blob/master/CHANGELOG.md)
- [Python (FastAPI) guide](https://github.com/RichGod93/bffgen/blob/master/docs/PYTHON_SUPPORT.md)
- [Aggregation guide](https://github.com/RichGod93/bffgen/blob/master/docs/NODEJS_AGGREGATION.md)
- [Issues](https://github.com/RichGod93/bffgen/issues)

MIT License
