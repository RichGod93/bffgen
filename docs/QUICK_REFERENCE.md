# bffgen Quick Reference Guide

## Commands at a Glance

```bash
# Initialize new project
bffgen init <name> --lang nodejs-express
bffgen init <name> --lang nodejs-fastify
bffgen init <name> --lang go --framework chi

# Generate code (routes, controllers, services)
bffgen generate

# Generate API documentation
bffgen generate-docs

# Add templates
bffgen add-template auth
bffgen add-template ecommerce
bffgen add-template content

# Add custom route
bffgen add-route

# Create Postman collection
bffgen postman

# Run development server (Go only)
bffgen dev

# Check project health
bffgen doctor
```

## Init Flags

```bash
# Language/Runtime
--lang, -l          go, nodejs-express, nodejs-fastify
--runtime, -r       (alias for --lang)

# Framework
--framework, -f     chi, echo, fiber (Go)
                    express, fastify (Node.js)

# Node.js Specific
--middleware        validation,logger,requestId,all,none
--controller-type   basic, aggregator, both [default: both]
--skip-tests        Don't generate test files
--skip-docs         Don't generate API docs
```

## Quick Start Recipes

### Express with Full Features

```bash
bffgen init my-bff --lang nodejs-express --middleware all
cd my-bff
bffgen add-template auth
bffgen generate
npm install && npm run dev
```

### Fastify Minimal

```bash
bffgen init my-bff \
  --lang nodejs-fastify \
  --middleware none \
  --skip-tests \
  --skip-docs
cd my-bff
npm install && npm start
```

### Go Chi Server

```bash
bffgen init my-bff --lang go --framework chi
cd my-bff
bffgen add-template ecommerce
bffgen generate
go run main.go
```

## Generated File Structure

### Express/Fastify

```
my-bff/
├── src/
│   ├── index.js           # Main server
│   ├── controllers/       # Business logic
│   ├── services/          # HTTP communication
│   ├── middleware/        # Request processing
│   ├── routes/            # Route handlers
│   ├── utils/             # Utilities (logger)
│   └── config/            # Configuration (swagger)
├── tests/                 # Jest tests
├── docs/                  # API documentation
├── jest.config.js
├── package.json
└── bffgen.config.json
```

### Go

```
my-bff/
├── main.go                # Main server
├── cmd/server/main.go     # Server entry
├── internal/
│   ├── routes/            # Route handlers
│   ├── aggregators/       # Data aggregation
│   ├── auth/              # Authentication
│   └── templates/         # YAML templates
├── go.mod
└── bff.config.yaml
```

## Common Workflows

### Add New Service Endpoint

```bash
# 1. Add to config manually (bffgen.config.json)
# or use interactive:
bffgen add-route

# 2. Generate code
bffgen generate

# 3. Files created:
# - src/routes/{service}.js
# - src/controllers/{service}.controller.js
# - src/services/{service}.service.js
```

### Generate API Documentation

```bash
# After adding routes
bffgen generate-docs

# Access at http://localhost:8080/api-docs
# Or export: docs/openapi.yaml
```

### Add Tests

```bash
# Tests auto-generated with:
bffgen init my-bff --lang nodejs-express

# Run tests:
cd my-bff
npm test

# Watch mode:
npm run test:watch

# Coverage:
npm run test:coverage
```

## Environment Variables

### Express/Fastify

```env
# Server
NODE_ENV=development
PORT=8080
HOST=0.0.0.0

# CORS
CORS_ORIGINS=http://localhost:3000,http://localhost:5173

# Security
JWT_SECRET=your-secret-key
COOKIE_SECRET=your-cookie-secret

# Logging
LOG_LEVEL=debug

# Backend Services
{SERVICE_NAME}_URL=http://localhost:5000/api
{SERVICE_NAME}_TIMEOUT=30000
{SERVICE_NAME}_RETRIES=3

# Rate Limiting
RATE_LIMIT=100
REDIS_URL=redis://localhost:6379
```

### Go

```env
JWT_SECRET=your-secret-key
ENCRYPTION_KEY=your-encryption-key
REDIS_URL=redis://localhost:6379
```

## NPM Scripts (Node.js)

```bash
npm start              # Production server
npm run dev            # Development (nodemon)
npm run dev:watch      # Watch mode
npm test               # Run tests
npm run test:watch     # Test watch mode
npm run test:coverage  # Coverage report
npm run lint           # ESLint
npm run lint:fix       # Auto-fix lint errors
npm run format         # Prettier format
npm run validate       # Lint + format + test
```

## Accessing Generated Features

### Swagger UI

```
http://localhost:8080/api-docs
```

### OpenAPI Spec

```
http://localhost:8080/api-docs.json
```

### Health Checks

```
http://localhost:8080/health      # Health status
http://localhost:8080/ready       # Readiness check
http://localhost:8080/live        # Liveness (Fastify)
```

## Customization

### Modify Controller Logic

Edit `src/controllers/{service}.controller.js` and add:

- Data transformation
- Multi-service aggregation
- Caching logic
- Business rules

### Modify Service Layer

Edit `src/services/{service}.service.js` and customize:

- Request/response interceptors
- Error handling
- Retry logic
- Timeout settings

### Add Custom Middleware

Create `src/middleware/{custom}.js` and register in `src/index.js`

### Extend Tests

Add test files in `tests/unit/` or `tests/integration/`

## Tips & Best Practices

### Controllers

- Use **basic** for simple pass-through
- Use **aggregator** when combining multiple services
- Implement caching in aggregator controllers

### Services

- Configure timeouts per service via env vars
- Use retry logic for transient failures
- Transform backend errors to standard format

### Middleware

- Order matters: logger → auth → validation → routes
- Use `asyncHandler` to catch async errors (Express)
- Use Fastify hooks for async operations

### Testing

- Mock HTTP calls with `nock`
- Use global test helpers from `tests/setup.js`
- Separate unit and integration tests

### Logging

- Use structured logging (JSON) in production
- Include request IDs for correlation
- Log errors with stack traces in development

## Troubleshooting

| Issue               | Solution                                    |
| ------------------- | ------------------------------------------- |
| Build fails         | Run `go build ./cmd/bffgen` to check errors |
| npm install fails   | Check Node.js version (>=18.0.0)            |
| Tests fail          | Verify all dependencies installed           |
| Swagger not loading | Check config files in `src/config/`         |
| Logger not working  | Check `logs/` directory permissions         |
| Controller errors   | Ensure service files exist                  |
| Service timeout     | Increase timeout in .env                    |

## Performance

### Build Time

- bffgen binary: ~2 seconds
- Project init: < 1 second
- Code generation: < 500ms
- npm install: ~30 seconds

### Generated Code

- Controllers: ~50 lines each
- Services: ~100 lines each
- Tests: ~80 lines each
- Total: Production-ready in seconds

## Support

- **Documentation**: `docs/` directory
- **Examples**: `examples/` directory
- **GitHub**: Issues and PRs welcome
- **Tests**: Run comprehensive test suite
