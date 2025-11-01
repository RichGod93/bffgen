# bffgen Configuration Examples

This directory contains example configuration files demonstrating bffgen's features across different runtimes and frameworks.

## Available Examples

### Go Projects

**`bffgen-v1.yaml`** - Comprehensive Go BFF example

- Framework: Chi router
- Features: Auth, CORS, rate limiting, circuit breakers
- Services: auth, user, content
- Aggregators: Dashboard, auth flow
- Middleware: Request logging, rate limiting, auth, CORS

### Node.js Express Projects

**`express-example.json`** - Express BFF example

- Framework: Express.js
- Features: JWT auth, Swagger UI, structured logging
- Services: auth, users, products
- Demonstrates: Microservices architecture, health checks
- Controllers: Auto-generated with separation of concerns

### Node.js Fastify Projects

**`fastify-example.json`** - Fastify BFF example

- Framework: Fastify
- Features: High-performance logging (Pino), Swagger, schema validation
- Services: auth, orders, cart
- Demonstrates: E-commerce BFF pattern, health/ready/live endpoints
- Controllers: With aggregator pattern support

### Python FastAPI Projects

**`python-example.json`** - FastAPI BFF example

- Framework: FastAPI
- Features: Async endpoints, Pydantic validation, OpenAPI docs, JWT auth
- Services: auth, users, products
- Demonstrates: Async/await patterns, circuit breakers, Redis caching
- Endpoints: Auto-generated routers with type hints and validation

## Reference Projects

The `projects/` directory contains complete, production-ready BFF implementations:

### Python Projects

**`projects/python/movie-dashboard-bff/`** - Movie Dashboard BFF

- Real-world example integrating TMDB API
- Features: Data aggregation, circuit breakers, caching, comprehensive testing
- Backends: TMDB API integration, mock user service
- Demonstrates: External API integration, dashboard aggregation pattern
- Documentation: Complete setup guide, API examples, testing strategy

## Using These Examples

### Copy to New Project

**Node.js Express:**

```bash
# Initialize with specific runtime
bffgen init my-bff --lang nodejs-express

# Copy example config
cp examples/express-example.json my-bff/bffgen.config.json

# Generate code
cd my-bff
bffgen generate
bffgen generate-docs

# Install and run
npm install
npm run dev
```

**Python FastAPI:**

```bash
# Initialize with specific runtime
bffgen init my-bff --lang python-fastapi

# Copy example config
cp examples/python-example.json my-bff/bffgen.config.py.json

# Generate code
cd my-bff
bffgen generate

# Install and run
./setup.sh
uvicorn main:app --reload --port 8000
```

### Reference for Configuration

Use these examples as reference when:

- Setting up microservices architecture
- Configuring multiple backend services
- Understanding endpoint definitions
- Setting up health checks
- Configuring authentication

## Configuration Schema Differences

### Go (YAML) vs Node.js (JSON) vs Python (JSON)

**Go Projects:** `bff.config.yaml`

- YAML format
- Go-specific settings (go_version, build_tags)
- Chi/Echo/Fiber framework options

**Node.js Projects:** `bffgen.config.json`

- JSON format
- Node.js-specific settings (runtime, npm scripts)
- Express/Fastify framework options

**Python Projects:** `bffgen.config.py.json`

- JSON format
- Python-specific settings (async, packageManager)
- FastAPI framework with async/await support
- Pydantic models for validation

## Key Configuration Sections

### Project

**Node.js:**

```json
{
  "project": {
    "name": "my-bff",
    "version": "1.0.0",
    "runtime": "nodejs-express",
    "framework": "express"
  }
}
```

**Python:**

```json
{
  "project": {
    "name": "my-bff",
    "version": "1.0.0",
    "framework": "fastapi",
    "async": true,
    "packageManager": "pip"
  }
}
```

### Server

```json
{
  "server": {
    "port": 8080,
    "cors": { "origins": [...] },
    "rateLimit": { "max": 100 },
    "security": { "helmet": true }
  }
}
```

### Backends

**Node.js (using exposeAs):**

```json
{
  "backends": [
    {
      "name": "auth",
      "baseUrl": "http://localhost:5000/api",
      "endpoints": [
        {
          "path": "/auth/login",
          "method": "POST",
          "exposeAs": "/api/auth/login",
          "requiresAuth": false
        }
      ]
    }
  ]
}
```

**Python (using upstreamPath):**

```json
{
  "backends": [
    {
      "name": "auth",
      "baseUrl": "http://localhost:5000",
      "endpoints": [
        {
          "name": "login",
          "method": "POST",
          "path": "/api/auth/login",
          "upstreamPath": "/auth/login",
          "requiresAuth": false,
          "description": "User authentication endpoint"
        }
      ]
    }
  ]
}
```

### Features

```json
{
  "features": {
    "authentication": { "enabled": true, "type": "jwt" },
    "logging": { "level": "info" },
    "documentation": { "swagger": { "enabled": true } }
  }
}
```

## Testing Examples

**Node.js:**

```bash
# Validate config JSON
cat examples/express-example.json | jq .

# Use as starting point
bffgen init my-project --lang nodejs-express
cp examples/express-example.json my-project/bffgen.config.json
cd my-project
bffgen generate
npm install
npm test
```

**Python:**

```bash
# Validate config JSON
cat examples/python-example.json | jq .

# Use as starting point
bffgen init my-project --lang python-fastapi
cp examples/python-example.json my-project/bffgen.config.py.json
cd my-project
bffgen generate
./setup.sh
pytest
```

## Customization Tips

1. **Modify URLs**: Update `baseUrl` to match your backend services
2. **Add Endpoints**: Add more endpoints to any backend service
3. **Change Ports**: Update `server.port` and backend URLs
4. **Toggle Features**: Enable/disable authentication, logging, docs
5. **Environment**: Use environment variables in generated `.env` or `.env.example`
6. **Python Async**: Toggle `async: true/false` for sync vs async endpoints
7. **Package Manager**: Choose `pip` or `poetry` for Python dependency management

## Additional Resources

### Documentation

- [Quick Reference Guide](../docs/QUICK_REFERENCE.md)
- [Python Support Guide](../docs/PYTHON_SUPPORT.md)
- [Enhanced Scaffolding Guide](../docs/ENHANCED_SCAFFOLDING.md)
- [Migration Guide](../docs/MIGRATION_GUIDE.md)
- [Schema Documentation](../schemas/bffgen-v1.json)

### Example Projects

- [Movie Dashboard BFF](./projects/python/movie-dashboard-bff/) - Full Python/FastAPI example
