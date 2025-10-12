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

## Using These Examples

### Copy to New Project

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

### Reference for Configuration

Use these examples as reference when:

- Setting up microservices architecture
- Configuring multiple backend services
- Understanding endpoint definitions
- Setting up health checks
- Configuring authentication

## Configuration Schema Differences

### Go (YAML) vs Node.js (JSON)

**Go Projects:** `bff.config.yaml`

- YAML format
- Go-specific settings (go_version, build_tags)
- Chi/Echo/Fiber framework options

**Node.js Projects:** `bffgen.config.json`

- JSON format
- Node.js-specific settings (runtime, npm scripts)
- Express/Fastify framework options

## Key Configuration Sections

### Project

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

```bash
# Validate config JSON
cat examples/express-example.json | jq .

# Use as starting point
bffgen init my-project --lang nodejs-express
cp examples/express-example.json my-project/bffgen.config.json
cd my-project
bffgen generate
```

## Customization Tips

1. **Modify URLs**: Update `baseUrl` to match your backend services
2. **Add Endpoints**: Add more endpoints to any backend service
3. **Change Ports**: Update `server.port` and backend URLs
4. **Toggle Features**: Enable/disable authentication, logging, docs
5. **Environment**: Use environment variables in generated `.env.example`

## Additional Resources

- [Quick Reference Guide](../docs/QUICK_REFERENCE.md)
- [Enhanced Scaffolding Guide](../docs/ENHANCED_SCAFFOLDING.md)
- [Migration Guide](../docs/MIGRATION_GUIDE.md)
- [Schema Documentation](../schemas/bffgen-v1.json)
