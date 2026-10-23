# bffgen

**Backend-for-Frontend (BFF) generator** - Scaffold secure, production-ready BFF services in **Go**, **Node.js (Express/Fastify)**, or **Python (FastAPI)** with JWT auth, rate limiting, CORS, and comprehensive logging.

[![CI](https://github.com/RichGod93/bffgen/actions/workflows/ci.yml/badge.svg)](https://github.com/RichGod93/bffgen/actions/workflows/ci.yml)
[![npm](https://img.shields.io/npm/v/bffgen.svg)](https://www.npmjs.com/package/bffgen)
[![Release](https://img.shields.io/github/v/release/RichGod93/bffgen)](https://github.com/RichGod93/bffgen/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/RichGod93/bffgen)](https://go.dev/)
[![Node Version](https://img.shields.io/badge/Node-18+-green.svg)](https://nodejs.org/)
[![Python Version](https://img.shields.io/badge/Python-3.10+-orange.svg)](https://python.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Python Support](https://img.shields.io/badge/Python-FastAPI-blue.svg)](https://github.com/RichGod93/bffgen)

---

## ⚡ Quick Start

### Go BFF

```bash
# Install
npm install -g bffgen

# Create Go BFF (Chi, Echo, or Fiber)
bffgen init my-go-bff --lang go --framework chi
cd my-go-bff && go run main.go
```

### Node.js BFF (Express)

```bash
# Create Express BFF
bffgen init my-express-bff --lang nodejs-express
cd my-express-bff
npm install && npm run dev
```

### Node.js BFF (Fastify)

```bash
# Create Fastify BFF
bffgen init my-fastify-bff --lang nodejs-fastify
cd my-fastify-bff
npm install && npm run dev
```

### Python BFF (FastAPI)

```bash
# Create FastAPI BFF
bffgen init my-python-bff --lang python-fastapi
cd my-python-bff
./setup.sh && source venv/bin/activate
uvicorn main:app --reload
```

**Example Output:**

```text
✅ BFF project 'my-bff' initialized successfully!

📋 Backend Configuration Summary:
   Architecture: Monolithic
   - Backend: http://localhost:3000/api
   - Services: users, products, orders, cart, auth

🔧 Setup Instructions:
   1. Start your monolithic backend: http://localhost:3000/api
   2. Run the BFF server: cd my-bff && go run main.go
   3. Test endpoints: curl http://localhost:8080/health

📁 Navigate to the project: cd my-bff
🚀 Start development server: go run main.go

🔐 Secure Authentication Setup:
   1. Set encryption key: export ENCRYPTION_KEY=<key>
   2. Set JWT secret: export JWT_SECRET=<key>
   3. Features: Encrypted JWT tokens, secure sessions, CSRF protection
   4. Auth endpoints: /api/auth/login, /api/auth/refresh, /api/auth/logout
```

---

## 🛠️ Commands

| Command           | Description                                                     | Go  | Node.js | Python |
| ----------------- | --------------------------------------------------------------- | --- | ------- | ------ |
| `init`            | Scaffold new BFF project (Go/Express/Fastify/FastAPI)           | ✅  | ✅      | ✅     |
| `add-route`       | Add backend endpoint interactively                              | ✅  | ✅      | ✅     |
| `add-template`    | Add auth/ecommerce/content templates                            | ✅  | ✅      | ✅     |
| `generate`        | Generate routes, controllers, and services from config          | ✅  | ✅      | ✅     |
| `generate-docs`   | Generate OpenAPI/Swagger documentation                          | -   | ✅      | -      |
| `postman`         | Create Postman collection                                       | ✅  | ✅      | ✅     |
| `dev`             | Run development server (Go/Python, use `npm run dev` for Node.js) | ✅  | -       | ✅     |
| `config`          | Manage global configuration and validation                      | ✅  | ✅      | ✅     |
| `config validate` | Validate project configuration (NEW v2.0)                       | ✅  | ✅      | ✅     |
| `convert`         | Convert config between YAML and JSON (NEW v2.0)                 | ✅  | ✅      | ✅     |
| `add-infra`       | Add infrastructure to existing project (NEW v2.0)               | ✅  | ✅      | ✅     |
| `doctor`          | Check project health and configuration                          | ✅  | ✅      | ✅     |

---

## ✨ Features

### 🌐 **Multi-Runtime Support**

- **Go (Chi/Echo/Fiber)** - High-performance, compiled servers
- **Node.js Express** - Popular, flexible web framework
- **Node.js Fastify** - Fast, schema-based framework
- **Node.js Apollo GraphQL** - GraphQL server with Apollo
- **Python FastAPI** - Modern, async Python web framework with automatic API docs
- **Go gqlgen GraphQL** - Type-safe GraphQL for Go
- **Template-Based Generation** - Embedded templates for consistency
- **Auto-Detection** - Commands detect project type automatically
- **Language-Specific Optimization** - Tailored code generation for each runtime

### 🏗️ **Enhanced Backend Architecture Support**

- **Microservices**: Different ports/URLs for each service
- **Monolithic**: Single port/URL for all services
- **Hybrid**: Services on same port with different paths
- **Smart Configuration**: Auto-generates bff.config.yaml or bffgen.config.json
- **Intelligent Defaults**: Smart port numbering and URL suggestions

### 🔒 Security Features

- **JWT Authentication** - Token validation with user context injection
- **Role-Based Access Control** - `"roles": ["admin"]` per endpoint: 403 without one of the roles (Express, Fastify, FastAPI; [details](docs/NODEJS_ROUTES.md#authentication-and-roles-rbac))
- **Rate Limiting** - Built-in for all runtimes (Redis optional)
- **Security Headers** - Helmet, CSP, HSTS, XSS protection
- **CORS Configuration** - Restrictive origins, credentials support
- **Request Validation** - Size limits, content-type validation

### 🎨 **Developer Experience**

- **Interactive CLI** - Guided project setup with prompts
- **Template System** - Pre-built templates (auth, ecommerce, content)
- **Code Generation** - Auto-generate routes, controllers, and services
- **Hot Reload** - Development mode with auto-restart (`node --watch` for Node.js)
- **Professional Structure** - `src/` directory, tests, middleware
- **Comprehensive Scripts** - `npm run dev`, `npm test`, `npm run lint`
- **Regeneration-Safe** - Re-run `generate` any time: files you edited are kept, the same config gives byte-identical output
- **Config Validation** - `bffgen config validate` catches errors before generation
- **Smart Tool Detection** - Auto-detects missing tools with install instructions
- **Auto npm install** - Optionally runs dependency installation after init
- **Dry Run and Check** - `generate --dry-run` lists changes without writing; `--check` fails CI when generated code is stale
- **Progress Indicators** - Visual feedback during multi-step operations
- **Runtime Override** - `--runtime` flag to override automatic detection
- **Auto-Route Registration** - Routes automatically imported in index files
- **Transaction Rollback** - Failed operations rollback automatically
- **Config Converter** - Convert between YAML (Go) and JSON (Node.js) formats
- **Add Infrastructure** - Add CI/CD, Docker, health checks to existing projects

### ✨ **Enhanced Node.js Scaffolding**

- **Controllers & Services** - Auto-generated with separation of concerns
- **Configurable Middleware** - Choose validation, logging, request ID tracking
- **Test Infrastructure** - Jest setup with sample tests and 70% coverage goals
- **API Documentation** - Swagger UI at `/api-docs` with OpenAPI 3.0 spec
- **Structured Logging** - Winston (Express) or Pino (Fastify) with file rotation
- **Resilient Routes** - Generated routes (Express, Fastify and FastAPI) forward path params and query strings, with per-backend timeouts, retries for idempotent requests, a circuit breaker, and 502/503/504 on backend failure ([Node.js](docs/NODEJS_ROUTES.md), [Python](docs/PYTHON_SUPPORT.md#what-generated-routes-do))
- **Aggregation Routes** - `"aggregates"` in the config combine several backends in one `GET` endpoint, called in parallel; an optional backend that fails leaves its data `null` (partial response) instead of failing the request (Express, Fastify and FastAPI; [details](docs/NODEJS_ROUTES.md#aggregates))
- **CLI Flags** - Non-interactive mode with full customization

### 🚀 **Infrastructure Scaffolding (NEW)**

- **CI/CD Pipelines** - GitHub Actions workflows with tests, linting, and Docker builds
- **Production Dockerfile** - Multi-stage, optimized, security-hardened containers
- **Health Checks** - Kubernetes-ready liveness/readiness probes with dependency checking
- **Graceful Shutdown** - Signal handling for zero-downtime deployments
- **Development Environment** - Docker Compose with Redis and all services
- **Opt-In Features** - Choose what you need with simple CLI flags
- **Cross-Runtime** - Works for both Go and Node.js projects

---

## 📦 Installation

### npm (Recommended for Node.js & Python Developers)

```bash
# Install globally
npm install -g bffgen

# Or use npx (no installation needed)
npx bffgen init my-project --lang nodejs-express

# Create FastAPI BFF
npx bffgen init my-python-bff --lang python-fastapi
```

### Go Install

```bash
go install github.com/RichGod93/bffgen/v3/cmd/bffgen@latest
```

Requires Go 1.24+. Prebuilt binaries (with `checksums.txt`) are on
[GitHub Releases](https://github.com/RichGod93/bffgen/releases/latest).

### From Source

```bash
git clone https://github.com/RichGod93/bffgen
cd bffgen && make build
sudo mv bffgen /usr/local/bin/
```

---

## 🚀 Usage Examples

### Initialize Project with Backend Architecture

```bash
bffgen init my-bff
✔ Which framework? (chi/echo/fiber) [chi]: fiber
✔ Frontend URLs (comma-separated) [localhost:3000,localhost:3001]: localhost:5173
✔ What's your backend architecture?
  1) Microservices (different ports/URLs)
  2) Monolithic (single port/URL)
  3) Hybrid (some services on same port)
✔ Select option (1-3) [1]: 2
✔ Backend base URL (e.g., 'http://localhost:3000/api'): http://localhost:3000/api
✔ Configure routes now or later?
  1) Define manually
  2) Use a template
  3) Skip for now
✔ Select option (1-3) [3]: 3
```

### Example: Microservices Architecture

```bash
bffgen init my-microservices-bff
✔ Which framework? (chi/echo/fiber) [chi]: chi
✔ Frontend URLs (comma-separated) [localhost:3000,localhost:3001]: localhost:5173
✔ What's your backend architecture?
  1) Microservices (different ports/URLs)
  2) Monolithic (single port/URL)
  3) Hybrid (some services on same port)
✔ Select option (1-3) [1]: 1
🔧 Configuring Microservices Backend
✔ Service name (e.g., 'users', 'products', 'orders'): users
✔ Base URL for users (e.g., 'http://localhost:4000/api'): http://localhost:4000/api
✅ Added users service on http://localhost:4000/api
✔ Service name (e.g., 'users', 'products', 'orders'): products
✔ Base URL for products (e.g., 'http://localhost:4000/api'): http://localhost:5000/api
✅ Added products service on http://localhost:5000/api
✔ Service name (e.g., 'users', 'products', 'orders'):
✔ Configure routes now or later?
  1) Define manually
  2) Use a template
  3) Skip for now
✔ Select option (1-3): 3

✅ BFF project 'my-microservices-bff' initialized successfully!

📋 Backend Configuration Summary:
   Architecture: Microservices
   - users: http://localhost:4000/api
   - products: http://localhost:5000/api

🔧 Setup Instructions:
   1. Start your microservices on the configured ports:
      - users: http://localhost:4000/api
      - products: http://localhost:5000/api
   2. Run the BFF server:
      cd my-microservices-bff
      go run main.go
   3. Test the endpoints:
      curl http://localhost:8080/health

📁 Navigate to the project: cd my-microservices-bff
🚀 Start development server: go run main.go
```

### Working with Templates (Optional Routes)

```bash
# Add routes using templates
bffgen add-template auth
# 📁 Template added: internal/templates/auth.yaml

# Add manual routes
bffgen add-route
# ✔ Service name: payments
# ✔ Endpoint path: /api/payments
# ✔ HTTP method [GET]:

# Generate routes in main.go when ready
bffgen generate
# ✅ Code generation completed!
# 📁 Updated: main.go (with proxy routes)
```

### Run Development Server

```bash
# Your project structure is ready! No generation needed.
ls -la my-microservices-bff/
# bff.config.yaml  go.mod  main.go  README.md  internal/

# Start your backend services first
# Users API: http://localhost:4000/api
# Products API: http://localhost:5000/api

# Then run the BFF server
cd my-microservices-bff
go run main.go
# 🚀 BFF server starting on :8080
```

### Create Postman Collection

```bash
bffgen postman
# 📮 Generating Postman collection from bff.config.yaml
# ✅ Postman collection generated successfully!
# 📁 Created file: bff-postman-collection.json
```

### Infrastructure Scaffolding

Generate production-ready infrastructure with opt-in flags:

```bash
# Generate everything (CI/CD, Docker, health checks, docker-compose)
bffgen init my-prod-bff --lang go --include-all-infra

# Selective generation
bffgen init my-api --lang nodejs-express --include-ci --include-docker

# Just health checks and graceful shutdown
bffgen init my-service --lang nodejs-fastify --include-health

# Add infrastructure to existing project (NEW v2.0)
cd existing-project
bffgen add-infra --ci --docker --compose
```

**Available Flags:**

- `--include-ci` - GitHub Actions CI/CD workflow
- `--include-docker` - Production Dockerfile and .dockerignore
- `--include-health` - Health check endpoints and graceful shutdown
- `--include-compose` - Development docker-compose.yml
- `--include-all-infra` - All infrastructure features

**Generated Files:**

```text
my-prod-bff/
├── .github/workflows/ci.yml      # CI/CD pipeline
├── Dockerfile                     # Production container
├── .dockerignore                  # Docker ignore rules
├── docker-compose.yml             # Dev environment
├── internal/                      # Go
│   ├── health/health.go          # Health checks
│   └── shutdown/graceful.go      # Graceful shutdown
└── src/utils/                     # Node.js
    ├── health.js                 # Health checks
    └── graceful-shutdown.js      # Graceful shutdown
```

**Quick Start:**

```bash
# Start development environment
docker-compose up

# Build production image
docker build -t my-bff:latest .

# Run with health checks
docker run -p 8080:8080 my-bff:latest
curl http://localhost:8080/healthz  # Liveness probe
curl http://localhost:8080/health   # Readiness probe
```

📖 **Detailed Guide:** [Infrastructure Documentation](docs/INFRASTRUCTURE.md)

---

## ⚙️ Configuration

bffgen saves your preferences in `~/.bffgen/bffgen.yaml` for re-runs:

### View Configuration

```bash
bffgen config show
```

### Set Defaults

```bash
bffgen config set framework fiber
bffgen config set cors_origins localhost:5173,myapp.com
bffgen config set jwt_secret my-super-secret-key
bffgen config set redis_url redis://localhost:6379
bffgen config set port 3000
bffgen config set route_option 2
```

### Reset Configuration

```bash
bffgen config reset
```

**Configuration File Location:** `~/.bffgen/bffgen.yaml`

---

## 🔴 Redis Setup (Chi/Echo Only)

```bash
# macOS
brew install redis && brew services start redis

# Ubuntu
sudo apt install redis-server && sudo systemctl start redis-server

# Docker
docker run -d -p 6379:6379 redis:alpine

# Verify
redis-cli ping  # Should return: PONG
```

**Note:** Fiber includes built-in rate limiting, no Redis needed.

---

## 🔐 JWT Authentication

### Environment Setup

```bash
export JWT_SECRET=your-super-secure-secret-key-change-in-production
```

### Token Generation

```go
import "github.com/golang-jwt/jwt/v5"

claims := jwt.MapClaims{
    "user_id": "123",
    "email": "user@example.com",
    "exp": time.Now().Add(time.Hour * 24).Unix(),
}

token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
```

### Usage

```bash
curl -H "Authorization: Bearer <your-jwt-token>" http://localhost:8080/api/protected
```

---

## 📂 Project Structure

### Go Project

```text
my-go-bff/
├── main.go                 # Generated server with routes
├── bff.config.yaml         # Service configuration
├── go.mod                  # Dependencies
├── README.md               # Project docs
└── internal/
    ├── routes/             # Route definitions
    ├── aggregators/        # Data aggregation
    └── templates/          # Template files
```

### Node.js Project (Express/Fastify)

```text
my-node-bff/
├── src/
│   ├── index.js            # Main server file
│   ├── controllers/        # 🆕 Auto-generated business logic
│   ├── services/           # 🆕 Auto-generated HTTP clients
│   ├── middleware/         # 🆕 Configurable middleware
│   ├── routes/             # Generated route files
│   ├── config/             # 🆕 Swagger configuration
│   └── utils/              # 🆕 Logger utility
├── tests/                  # 🆕 Jest test infrastructure
│   ├── unit/               # Unit tests
│   ├── integration/        # Integration tests
│   └── setup.js            # Test helpers
├── docs/                   # 🆕 API documentation
│   └── openapi.yaml
├── jest.config.js          # 🆕 Jest configuration
├── bffgen.config.json      # BFF configuration
├── package.json            # Dependencies & scripts
├── .env.example            # Environment template
├── .gitignore              # Git ignore
└── README.md               # Project docs
```

---

## 🚀 Node.js Usage Guide

### Express Workflow (Enhanced)

```bash
# 1. Initialize Express project with full features
bffgen init my-express-bff --lang nodejs-express --middleware all

# 2. Navigate to project
cd my-express-bff

# 3. Add a template (auth, ecommerce, or content)
bffgen add-template ecommerce

# 4. Generate routes, controllers, and services
bffgen generate

# 5. Generate API documentation
bffgen generate-docs

# 6. Install dependencies
npm install

# 7. Start development server
npm run dev

# 8. Test endpoints
curl http://localhost:8080/health
curl http://localhost:8080/api/products

# 9. View API documentation
open http://localhost:8080/api-docs

# 10. Run tests
npm test
```

### Fastify Workflow

```bash
# 1. Initialize Fastify project
bffgen init my-fastify-bff --lang nodejs-fastify

# 2. Navigate to project
cd my-fastify-bff

# 3. Add auth template
bffgen add-template auth

# 4. Generate routes
bffgen generate

# 5. Install dependencies
npm install

# 6. Start development server
npm run dev

# 7. Test auth endpoints (the demo login accepts any credentials and is off in production)
curl http://localhost:8080/health
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

### Available npm Scripts

| Script               | Description                        |
| -------------------- | ---------------------------------- |
| `npm start`          | Start production server            |
| `npm run dev`        | Start development server (`node --watch`) |
| `npm run dev:watch`  | Watch mode with file watching      |
| `npm test`           | Run tests with coverage            |
| `npm run test:watch` | Watch mode for TDD                 |
| `npm run lint`       | Check code with ESLint             |
| `npm run lint:fix`   | Auto-fix linting issues            |
| `npm run format`     | Format code with Prettier          |
| `npm run validate`   | Run lint + format + tests          |

### Templates for Node.js

All templates work with both Go and Node.js projects:

| Template    | Services               | Endpoints | Description                               |
| ----------- | ---------------------- | --------- | ----------------------------------------- |
| `auth`      | auth                   | 5         | Login, register, refresh, logout, profile |
| `ecommerce` | products, orders, cart | 13        | Full e-commerce backend                   |
| `content`   | posts, comments, likes | 12        | Content management system                 |

**Usage:**

```bash
bffgen add-template auth       # Adds auth endpoints to config
bffgen generate                # Generates route files
```

### Environment Configuration

Node.js projects use `.env.example` as a template:

```bash
# Copy and configure
cp .env.example .env

# Edit with your values
vim .env
```

**Key Variables:**

- `NODE_ENV` - Environment (development/production)
- `PORT` - Server port (default: 8080)
- `JWT_SECRET` - HS256 secret for JWT tokens; with `NODE_ENV=production` the server won't start without a random one of 32+ characters ([details](docs/NODEJS_ROUTES.md#tokens-and-jwt_secret))
- `AUTH_DEMO_LOGIN` - Serve the demo `/api/auth/login` (accepts any credentials); default: everywhere except production
- `CORS_ORIGINS` - Allowed CORS origins
- `*_URL` - Backend service URLs
- `*_TIMEOUT`, `*_RETRIES`, `*_CIRCUIT_*` - Per-backend resilience overrides ([details](docs/NODEJS_ROUTES.md#environment-overrides))

---

## 🎯 Enhanced Features Deep Dive

### Controllers & Services Architecture

```javascript
// Thin route layer
router.get('/api/users', authenticate, controller.getUsers);

// Controller focuses on business logic
async getUsers(req, res, next) {
  const users = await this.service.getUsers();
  const enriched = this.enrichUserData(users);
  res.json(enriched);
}

// Service handles HTTP communication
async getUsers() {
  return this.client.get('/users');
}
```

### CLI Flags Reference

```bash
# Full-featured setup
bffgen init my-bff \
  --lang nodejs-express \
  --middleware all \
  --controller-type both

# Minimal setup
bffgen init my-bff \
  --lang nodejs-fastify \
  --middleware none \
  --skip-tests \
  --skip-docs

# Custom middleware
bffgen init my-bff \
  --lang nodejs-express \
  --middleware validation,logger
```

### Generated Files Summary

**During `init`:**

- Main server (`src/index.js`)
- HTTP client (`src/services/httpClient.js`)
- Logger utility (`src/utils/logger.js`)
- Swagger config (`src/config/swagger-*.js`)
- Jest config (`jest.config.js`, `tests/setup.js`)
- Sample test (`tests/integration/health.test.js`)

**During `generate`:**

- Routes (`src/routes/{service}.js`), see [Generated Node.js Routes](docs/NODEJS_ROUTES.md)
- Controllers (`src/controllers/{service}.controller.js`)
- Services (`src/services/{service}.service.js`)

**During `generate-docs`:**

- OpenAPI spec (`docs/openapi.yaml`)

### v2.0 Workflow Enhancements (NEW)

```bash
# Validate configuration before generation
bffgen config validate

# Preview changes before applying
bffgen generate --dry-run

# Also regenerate files you edited (your versions are kept as .bak)
bffgen generate --force

# Convert config formats
bffgen convert config --from yaml --to json

# Override runtime detection
bffgen --runtime nodejs-express generate

# Add infrastructure to existing project
bffgen add-infra --all
```

**Key Improvements:**

- Routes auto-register in `src/index.js` (no manual imports needed)
- JWT auth is fully implemented (no TODO comments)
- Generation is idempotent (safe to run multiple times)
- Failed operations automatically rollback
- Real `httputil.ReverseProxy` for Go (no more "not implemented" stubs)

### Quick Links

- 📖 [Enhanced Scaffolding Guide](docs/ENHANCED_SCAFFOLDING.md)
- 🚀 [Infrastructure Scaffolding](docs/INFRASTRUCTURE.md)
- 📋 [Quick Reference](docs/QUICK_REFERENCE.md)
- 🛡️ [Generated Node.js Routes](docs/NODEJS_ROUTES.md)
- 🧩 [Node.js Aggregation](docs/NODEJS_AGGREGATION.md)
- 🆕 [v2.0 Features & Migration](docs/V2_MIGRATION.md)

---

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

---

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- [Chi Router](https://github.com/go-chi/chi) - Lightweight HTTP router
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [JWT](https://github.com/golang-jwt/jwt) - JSON Web Tokens
- [Winston](https://github.com/winstonjs/winston) - Express logging
- [Pino](https://github.com/pinojs/pino) - Fastify logging
- [Jest](https://jestjs.io/) - Testing framework
- [Swagger](https://swagger.io/) - API documentation
- Inspired by [Backend-for-Frontend pattern](https://martinfowler.com/articles/bff.html) by Martin Fowler
