# bffgen

**Backend-for-Frontend (BFF) generator** - Scaffold secure, production-ready BFF services in **Go**, **Node.js (Express)**, or **Node.js (Fastify)** with JWT auth, rate limiting, CORS, and comprehensive logging.

[![npm version](https://img.shields.io/npm/v/bffgen.svg)](https://www.npmjs.com/package/bffgen)
[![Downloads](https://img.shields.io/npm/dm/bffgen.svg)](https://www.npmjs.com/package/bffgen)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## ⚡ Quick Start

### Using npx (No Installation)

```bash
# Create Express BFF
npx bffgen init my-express-bff --lang nodejs-express
cd my-express-bff
npm install && npm run dev

# Create Fastify BFF
npx bffgen init my-fastify-bff --lang nodejs-fastify

# Create Go BFF
npx bffgen init my-go-bff --lang go --framework chi
```

### Global Installation

```bash
# Install globally
npm install -g bffgen

# Use anywhere
bffgen init my-project --lang nodejs-express
cd my-project && npm run dev
```

---

## ✨ Features

### 🌐 **Multi-Runtime Support**

- **Node.js Express** - Popular, flexible web framework
- **Node.js Fastify** - Fast, schema-based framework
- **Go (Chi/Echo/Fiber)** - High-performance, compiled servers

### 🚀 **Production-Ready Aggregation**

- **Parallel Service Calls** - Fetch from multiple backends simultaneously
- **Redis Caching** - Built-in caching with automatic fallback
- **Circuit Breakers** - Prevent cascade failures
- **Request Batching** - Avoid N+1 queries
- **Response Transformation** - Filter and optimize API responses
- **Field Selection** - GraphQL-like field filtering for REST
- **Go & Node.js Parity** - Same utilities available in both runtimes (v2.0+)

### 🔒 **Security Features**

- **JWT Authentication** - Token validation with user context
- **Rate Limiting** - Built-in for all runtimes
- **Security Headers** - Helmet, CSP, HSTS, XSS protection
- **CORS Configuration** - Restrictive origins, credentials support

### 🎨 **Developer Experience**

- **Interactive CLI** - Guided project setup
- **Template System** - Pre-built templates (auth, ecommerce, content)
- **Code Generation** - Auto-generate routes, controllers, services
- **Hot Reload** - Development mode with auto-restart
- **Comprehensive Tests** - Jest setup with sample tests

### ⚡ **v2.0 Enhancements** (NEW)

- **Idempotent Generation** - Safe to run `generate` multiple times
- **Config Validation** - `bffgen config validate` catches errors pre-generation
- **Colorized Diffs** - Preview changes with `--dry-run`
- **Progress Indicators** - Visual feedback during operations
- **Auto-Route Registration** - Routes automatically imported
- **Runtime Override** - `--runtime` flag for explicit control
- **Transaction Rollback** - Failed operations automatically rollback
- **Smart Tool Detection** - Auto-detects missing dependencies
- **Memory Safety CI** - Automated leak detection and security scanning

---

## 🛠️ Commands

```bash
# Initialize new BFF project
bffgen init my-bff --lang nodejs-express

# Add route interactively
bffgen add-route

# Add template (auth, ecommerce, content)
bffgen add-template auth

# Generate routes, controllers, and services
bffgen generate

# Generate with preview (v2.0)
bffgen generate --dry-run

# Force regeneration (v2.0)
bffgen generate --force

# Validate configuration (v2.0)
bffgen config validate

# Convert config formats (v2.0)
bffgen convert --from yaml --to json

# Add infrastructure (v2.0)
bffgen add-infra --ci --docker --compose

# Generate API documentation
bffgen generate-docs

# Create Postman collection
bffgen postman

# Health check
bffgen doctor

# Run development server (Go only)
bffgen dev

# Show version
bffgen version
```

---

## 📚 Examples

### Node.js Express Example

```bash
# Create project
npx bffgen init my-express-bff --lang nodejs-express

# Project structure:
my-express-bff/
├── src/
│   ├── index.js              # Express server
│   ├── routes/               # Route handlers
│   ├── controllers/          # Business logic with aggregation
│   ├── services/             # HTTP clients
│   ├── middleware/           # Auth, validation, logging
│   ├── utils/                # Aggregation utilities
│   │   ├── aggregator.js     # Parallel requests
│   │   ├── cache-manager.js  # Redis caching
│   │   ├── circuit-breaker.js # Fault tolerance
│   │   ├── response-transformer.js # Data transformation (v2.0)
│   │   ├── request-batcher.js     # Request batching (v2.0)
│   │   └── ...
│   └── examples/             # Working aggregation examples
├── tests/                    # Jest tests
├── docker-compose.yml        # Redis setup
├── package.json
└── bffgen.config.json        # BFF configuration
```

### Aggregation Example (v2.0)

```javascript
const ParallelAggregator = require("./utils/aggregator");
const CacheManager = require("./utils/cache-manager");
const CircuitBreaker = require("./utils/circuit-breaker");
const ResponseTransformer = require("./utils/response-transformer");

const aggregator = new ParallelAggregator({ timeout: 5000 });
const cache = new CacheManager({ ttl: 300 });
const breaker = new CircuitBreaker({ failureThreshold: 5 });
const transformer = new ResponseTransformer();

// Fetch from multiple services in parallel with circuit breaker
const results = await aggregator.fetchParallel([
  {
    name: "user",
    fetch: () => breaker.execute(() => UserService.getProfile(userId)),
  },
  { name: "orders", fetch: () => OrdersService.getOrders(userId) },
  { name: "preferences", fetch: () => PreferencesService.get(userId) },
]);

// Transform and sanitize responses
const dashboard = {
  user: transformer.sanitize(
    results.find((r) => r.service === "user" && r.success)?.data
  ),
  orders: results.find((r) => r.service === "orders" && r.success)?.data || [],
  preferences:
    results.find((r) => r.service === "preferences" && r.success)?.data || {},
};
```

---

## 🔧 Programmatic API

```javascript
const bffgen = require("bffgen");

// Initialize project programmatically
await bffgen.init({
  name: "my-project",
  lang: "nodejs-express",
  framework: "express",
  skipTests: false,
});

// Generate code
await bffgen.generate();

// Get version
const version = bffgen.getVersion();
```

---

## 📖 Documentation

- [Full Documentation](https://github.com/RichGod93/bffgen)
- [Node.js Aggregation Guide](https://github.com/RichGod93/bffgen/blob/main/docs/NODEJS_AGGREGATION.md)
- [Quick Reference](https://github.com/RichGod93/bffgen/blob/main/docs/QUICK_REFERENCE.md)
- [Examples](https://github.com/RichGod93/bffgen/tree/main/examples)

---

## 🌍 Platform Support

Supported platforms:

- ✅ macOS (Intel & Apple Silicon)
- ✅ Linux (x64 & ARM64)
- ✅ Windows (x64)

The appropriate binary for your platform is automatically downloaded during installation.

---

## 🐛 Troubleshooting

### Installation Issues

If installation fails:

1. **Check your internet connection**
2. **Clear npm cache:**

   ```bash
   npm cache clean --force
   npm install -g bffgen
   ```

3. **Manual installation:**
   Download from [GitHub Releases](https://github.com/RichGod93/bffgen/releases)

### Platform Not Supported

If your platform isn't supported, you can:

- Install via Go: `go install github.com/RichGod93/bffgen/cmd/bffgen@latest`
- Build from source: Clone the repo and run `make build`

---

## 🤝 Contributing

Contributions are welcome! Please see the [Contributing Guide](https://github.com/RichGod93/bffgen/blob/main/CONTRIBUTING.md).

---

## 📝 License

MIT License - see [LICENSE](LICENSE) file for details.

---

## 🔗 Links

- [GitHub Repository](https://github.com/RichGod93/bffgen)
- [Documentation](https://github.com/RichGod93/bffgen#readme)
- [Issue Tracker](https://github.com/RichGod93/bffgen/issues)
- [npm Package](https://www.npmjs.com/package/bffgen)

---

**Made with ❤️ for the Backend-for-Frontend pattern**
