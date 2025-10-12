# Infrastructure Scaffolding Guide

## Overview

bffgen provides opt-in infrastructure scaffolding to help you deploy production-ready BFF services quickly. These features generate CI/CD pipelines, Docker configurations, health checks, graceful shutdown handlers, and development environments.

**Key Benefits:**

- Save 2-3 hours of infrastructure setup time
- Production-ready configurations out of the box
- Consistent patterns across Go and Node.js projects
- Easy customization for your specific needs

---

## Features

### 1. CI/CD Pipeline (GitHub Actions)

Generates a complete GitHub Actions workflow with:

- Automated testing on pull requests and pushes
- Matrix strategy for multiple language versions
- Caching for faster builds
- Linting and code quality checks
- Optional Docker image building and pushing to registry
- Code coverage reporting with Codecov

**Flag:** `--include-ci`

**Generated Files:**

- `.github/workflows/ci.yml`

**Supports:**

- Go: Tests with Go 1.21 and 1.22
- Node.js: Tests with Node 18.x and 20.x

### 2. Production Dockerfile

Generates optimized, multi-stage Dockerfiles with:

- Minimal runtime images (Alpine Linux)
- Non-root user for security
- Health check directives
- Layer caching optimization
- Production-ready environment

**Flag:** `--include-docker`

**Generated Files:**

- `Dockerfile`
- `.dockerignore`

**Features:**

- **Go:** Multi-stage build, static binary compilation
- **Node.js:** Production dependencies only, optimized layers

### 3. Enhanced Health Checks

Generates comprehensive health check endpoints with:

- Liveness probe (basic availability check)
- Readiness probe (dependency validation)
- Parallel dependency checking
- Backend service health verification
- Redis connectivity check
- Kubernetes-ready format

**Flag:** `--include-health`

**Generated Files:**

- **Go:** `internal/health/health.go`, `internal/shutdown/graceful.go`
- **Node.js:** `src/utils/health.js`, `src/utils/graceful-shutdown.js`

**Endpoints:**

- `GET /healthz` - Liveness probe (always returns 200 OK)
- `GET /health` - Readiness probe (checks all dependencies)

### 4. Graceful Shutdown

Generates signal handlers for clean shutdown with:

- SIGTERM and SIGINT handling
- Connection draining (waits for in-flight requests)
- Configurable timeout
- Cleanup hooks for resources
- Zero-downtime deployments

**Flag:** Included automatically with `--include-health`

**Features:**

- Stops accepting new connections
- Waits for existing requests to complete
- Runs cleanup tasks (close DB connections, flush logs)
- Exits with proper status codes

### 5. Development Docker Compose

Generates a complete local development environment with:

- BFF service container
- Redis container with health checks
- Redis Commander UI (optional, debug profile)
- Volume mounting for hot reload
- Network configuration
- Environment variable templates

**Flag:** `--include-compose`

**Generated Files:**

- `docker-compose.yml`

**Features:**

- One command to start entire stack: `docker-compose up`
- Redis data persistence with volumes
- Automatic service dependency management
- Debug tools available with profiles

---

## Usage

### Basic Usage

```bash
# Generate all infrastructure files
bffgen init my-bff --lang go --include-all-infra

# Generate specific features
bffgen init my-bff --lang nodejs-express --include-ci --include-docker

# Generate CI and health checks only
bffgen init my-bff --lang nodejs-fastify --include-ci --include-health
```

### Flags Reference

| Flag                  | Description                                           |
| --------------------- | ----------------------------------------------------- |
| `--include-ci`        | Generate GitHub Actions CI/CD workflow                |
| `--include-docker`    | Generate production Dockerfile and .dockerignore      |
| `--include-health`    | Generate health checks and graceful shutdown handlers |
| `--include-compose`   | Generate development docker-compose.yml               |
| `--include-all-infra` | Enable all infrastructure features (shortcut)         |

---

## Examples

### Example 1: Full Production Setup (Go)

```bash
# Create Go BFF with all infrastructure
bffgen init my-go-bff --lang go --framework chi --include-all-infra

# Directory structure
my-go-bff/
├── .github/workflows/ci.yml          # CI/CD pipeline
├── Dockerfile                         # Production container
├── .dockerignore                      # Docker ignore rules
├── docker-compose.yml                 # Development environment
├── internal/
│   ├── health/health.go              # Health check handlers
│   └── shutdown/graceful.go          # Graceful shutdown
└── ... (other BFF files)
```

**What you get:**

- Automated testing on every push
- Production-ready Docker image
- Health endpoints for Kubernetes
- Local development environment
- Graceful shutdown handling

### Example 2: Node.js Express with CI/CD

```bash
# Create Express BFF with CI and Docker
bffgen init my-express-bff --lang nodejs-express --include-ci --include-docker

# Directory structure
my-express-bff/
├── .github/workflows/ci.yml          # CI/CD pipeline
├── Dockerfile                         # Production container
├── .dockerignore                      # Docker ignore rules
├── src/
│   └── index.js                      # Main server (with health endpoints if --include-health)
└── ... (other BFF files)
```

### Example 3: Minimal Setup with Health Checks

```bash
# Just add health checks and graceful shutdown
bffgen init my-minimal-bff --lang nodejs-fastify --include-health

# Directory structure
my-minimal-bff/
├── src/
│   └── utils/
│       ├── health.js                 # Health check utility
│       └── graceful-shutdown.js      # Graceful shutdown handler
└── ... (other BFF files)
```

---

## Generated File Details

### GitHub Actions Workflow (.github/workflows/ci.yml)

**Go Workflow:**

```yaml
name: CI
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  test:
    strategy:
      matrix:
        go-version: ["1.21", "1.22"]
    steps:
      - Checkout code
      - Setup Go with caching
      - Install dependencies
      - Run tests with coverage
      - Run linter (golangci-lint)
      - Build binary
      - Upload coverage to Codecov
```

**Node.js Workflow:**

```yaml
jobs:
  test:
    strategy:
      matrix:
        node-version: ["18.x", "20.x"]
    steps:
      - Checkout code
      - Setup Node.js with npm caching
      - Install dependencies (npm ci)
      - Run linter
      - Run tests with coverage
      - Upload coverage to Codecov
```

**Optional Docker Job** (if `--include-docker` also specified):

- Builds Docker image after tests pass
- Pushes to Docker Hub registry
- Only runs on main branch pushes
- Uses layer caching for speed

**Setup Requirements:**

1. No setup needed for basic CI (tests, linting)
2. For Docker push: Add `DOCKER_USERNAME` and `DOCKER_PASSWORD` secrets in GitHub repo settings
3. For Codecov: Add `CODECOV_TOKEN` secret (optional, works without it)

### Production Dockerfile

**Go Multi-Stage Dockerfile:**

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder
- Install ca-certificates and tzdata
- Copy go.mod and go.sum
- Download dependencies
- Copy source code
- Build static binary with size optimization

# Runtime stage
FROM alpine:latest
- Install runtime dependencies
- Create non-root user (app:1000)
- Copy binary from builder
- Expose port
- Add health check
- Run as non-root user
```

**Node.js Multi-Stage Dockerfile:**

```dockerfile
# Build stage
FROM node:20-alpine AS builder
- Copy package files
- Install production dependencies only
- Copy source code

# Runtime stage
FROM node:20-alpine
- Install curl for health checks
- Create non-root user (app:1000)
- Copy dependencies and source
- Set NODE_ENV=production
- Expose port
- Add health check
- Run as non-root user
```

**Security Features:**

- Non-root user execution
- Minimal attack surface (Alpine Linux)
- No unnecessary packages
- Static binary (Go) or production deps only (Node.js)

**Build Commands:**

```bash
# Build image
docker build -t my-bff:latest .

# Run container
docker run -p 8080:8080 -e JWT_SECRET=secret my-bff:latest

# Test health check
curl http://localhost:8080/healthz
```

### Health Check Endpoints

**Go Implementation:**

```go
// Liveness probe - always returns OK
GET /healthz
Response: {"status":"ok","timestamp":"2025-10-12T10:30:00Z"}

// Readiness probe - checks dependencies
GET /health
Response: {
  "status": "ok",  // or "degraded"
  "version": "1.0.0",
  "timestamp": "2025-10-12T10:30:00Z",
  "dependencies": {
    "http://backend1:3000": true,
    "http://backend2:3000": false  // if unhealthy
  }
}
```

**Node.js Implementation:**

```javascript
// Usage in your code
const HealthChecker = require("./utils/health");

const healthChecker = new HealthChecker({
  version: "1.0.0",
  redisClient: redisClient, // optional
  backendServices: {
    "users-service": "http://users-api:3000",
    "products-service": "http://products-api:3000",
  },
});

// Express
app.get("/healthz", (req, res) => res.json(healthChecker.liveness()));
app.get("/health", async (req, res) => {
  const status = await healthChecker.readiness();
  res.status(status.status === "ok" ? 200 : 503).json(status);
});
```

**Kubernetes Integration:**

```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
  failureThreshold: 3
```

### Graceful Shutdown

**Go Implementation:**

```go
import "github.com/<your-project>/internal/shutdown"

func main() {
    server := &http.Server{Addr: ":8080", Handler: router}

    // Start server in goroutine
    go func() {
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()

    // Wait for shutdown signal (blocks)
    shutdown.GracefulShutdown(server, 30*time.Second)
}
```

**Node.js Implementation:**

```javascript
const GracefulShutdown = require("./utils/graceful-shutdown");

const server = app.listen(PORT, () => {
  console.log(`Server running on port ${PORT}`);
});

// Setup graceful shutdown
const gracefulShutdown = new GracefulShutdown(server, {
  timeout: 30000, // 30 seconds
  logger: console,
  cleanup: async () => {
    // Close database connections
    await db.close();
    // Close Redis connection
    await redis.quit();
    // Flush logs
    logger.flush();
  },
});

gracefulShutdown.handle();
```

**What happens on SIGTERM:**

1. Log shutdown signal received
2. Stop accepting new connections
3. Wait for in-flight requests to complete (up to timeout)
4. Run cleanup tasks (close DB, Redis, etc.)
5. Exit with code 0 (success) or 1 (error)

### Docker Compose Development Environment

**Generated docker-compose.yml:**

```yaml
services:
  my-bff-bff:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - REDIS_URL=redis://redis:6379
      - USERS_SERVICE_URL=http://users-api:3000
      - JWT_SECRET=dev-secret
    depends_on:
      redis:
        condition: service_healthy
    volumes:
      - .:/app # Hot reload

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]

  redis-commander:
    image: rediscommander/redis-commander:latest
    ports:
      - "8081:8081"
    profiles:
      - debug # Only starts with: docker-compose --profile debug up
```

**Usage:**

```bash
# Start all services
docker-compose up

# Start with Redis UI for debugging
docker-compose --profile debug up

# Start in background
docker-compose up -d

# View logs
docker-compose logs -f my-bff-bff

# Stop all services
docker-compose down

# Clean up volumes too
docker-compose down -v
```

---

## Customization

### Modifying CI Workflow

Edit `.github/workflows/ci.yml`:

```yaml
# Add custom test commands
- name: Run integration tests
  run: npm run test:integration

# Add deployment step
- name: Deploy to staging
  if: github.ref == 'refs/heads/develop'
  run: ./deploy-staging.sh

# Change Go versions
strategy:
  matrix:
    go-version: ['1.21', '1.22', '1.23']
```

### Customizing Dockerfile

**Add environment-specific configurations:**

```dockerfile
# Add build args
ARG NODE_ENV=production
ENV NODE_ENV=$NODE_ENV

# Add custom dependencies
RUN apk add --no-cache postgresql-client

# Change exposed port
EXPOSE 3000

# Add custom healthcheck interval
HEALTHCHECK --interval=10s --timeout=2s CMD curl -f http://localhost:3000/health || exit 1
```

### Extending Health Checks

**Add custom dependency checks (Node.js):**

```javascript
const HealthChecker = require("./utils/health");

class CustomHealthChecker extends HealthChecker {
  async readiness() {
    const baseStatus = await super.readiness();

    // Add database check
    try {
      await db.ping();
      baseStatus.dependencies.database = true;
    } catch (err) {
      baseStatus.dependencies.database = false;
      baseStatus.status = "degraded";
    }

    return baseStatus;
  }
}
```

### Extending Docker Compose

**Add database service:**

```yaml
services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: bff_db
      POSTGRES_USER: bff_user
      POSTGRES_PASSWORD: dev_password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

**Add backend mock service:**

```yaml
services:
  mock-backend:
    image: mockoon/cli:latest
    volumes:
      - ./mocks/backend.json:/data/backend.json:ro
    command: ["--data", "/data/backend.json", "--port", "3000"]
    ports:
      - "3000:3000"
```

---

## Best Practices

### CI/CD

1. **Keep workflows fast**

   - Use caching for dependencies
   - Run tests in parallel where possible
   - Only build Docker on main branch

2. **Use matrix strategy**

   - Test against multiple language versions
   - Catch compatibility issues early

3. **Secure secrets**
   - Never commit secrets to repository
   - Use GitHub Secrets for sensitive data
   - Rotate credentials regularly

### Docker

1. **Security**

   - Always run as non-root user
   - Use minimal base images (Alpine)
   - Keep images updated
   - Scan for vulnerabilities

2. **Optimization**

   - Use multi-stage builds
   - Leverage layer caching
   - Copy dependencies before source code
   - Use .dockerignore liberally

3. **Health checks**
   - Keep health check endpoints lightweight
   - Set appropriate timeouts
   - Don't fail on slow dependencies in liveness probe

### Health Checks

1. **Liveness vs Readiness**
   - **Liveness:** Just check if process is alive
   - **Readiness:** Check if ready to serve traffic
2. **Dependency checking**
   - Set reasonable timeouts (2-3 seconds)
   - Don't cascade failures
   - Log failed dependency checks
3. **Kubernetes**
   - Use both liveness and readiness probes
   - Set appropriate initialDelaySeconds
   - Configure failureThreshold for stability

### Graceful Shutdown

1. **Timeout values**

   - Set based on longest request duration
   - Typical: 30 seconds
   - For long-running requests: 60+ seconds

2. **Cleanup tasks**

   - Close database connections
   - Flush logs and metrics
   - Cancel background jobs
   - Clear caches if needed

3. **Testing**
   - Send SIGTERM to test locally
   - Monitor logs during shutdown
   - Verify no dropped connections

---

## Troubleshooting

### CI Workflow Not Running

**Problem:** Workflow doesn't trigger on push

**Solution:**

1. Check workflow is in `.github/workflows/` directory
2. Verify YAML syntax is valid
3. Check branch names match in workflow
4. Ensure workflow file is committed and pushed

### Docker Build Fails

**Problem:** `COPY` command fails

**Solution:**

1. Check files exist before COPY
2. Verify .dockerignore isn't excluding needed files
3. Use `docker build --progress=plain` for detailed logs
4. Check file permissions

**Problem:** Image is too large

**Solution:**

1. Use Alpine base images
2. Clean up in same layer: `RUN install && cleanup`
3. Use multi-stage builds
4. Check what's in image: `docker history my-image`

### Health Check Always Fails

**Problem:** `/health` returns 503

**Solution:**

1. Check backend services are actually running
2. Verify network connectivity to backends
3. Check timeout values aren't too low
4. Add logging to see which dependency fails
5. Test individual services: `curl http://backend/health`

**Problem:** Health check times out

**Solution:**

1. Reduce number of dependencies checked
2. Increase health check timeout
3. Check for slow network or DNS issues
4. Consider parallel checking (already implemented)

### Graceful Shutdown Not Working

**Problem:** Process exits immediately on SIGTERM

**Solution:**

1. Verify signal handlers are registered
2. Check you're not calling `process.exit()` elsewhere
3. Ensure server.close() is being called
4. Add logging to track shutdown flow

**Problem:** Timeout always exceeded

**Solution:**

1. Check for hanging connections
2. Increase timeout value
3. Force close after timeout
4. Review cleanup tasks for blocking operations

### Docker Compose Issues

**Problem:** Services can't communicate

**Solution:**

1. Check all services in same network
2. Use service names for hostnames (not localhost)
3. Verify ports are exposed correctly
4. Check environment variables are set

**Problem:** Redis connection refused

**Solution:**

1. Wait for Redis health check to pass
2. Use `depends_on` with condition
3. Check REDIS_URL format: `redis://redis:6379`
4. Verify Redis container is actually running

---

## Migration Guide

### Adding Infrastructure to Existing Project

If you have an existing bffgen project without infrastructure:

1. **Generate new project with infrastructure**

   ```bash
   bffgen init temp-project --lang your-lang --include-all-infra
   ```

2. **Copy infrastructure files**

   ```bash
   cp -r temp-project/.github your-project/
   cp temp-project/Dockerfile your-project/
   cp temp-project/.dockerignore your-project/
   cp temp-project/docker-compose.yml your-project/
   ```

3. **Copy health and shutdown utilities**

   **Go:**

   ```bash
   cp -r temp-project/internal/health your-project/internal/
   cp -r temp-project/internal/shutdown your-project/internal/
   ```

   **Node.js:**

   ```bash
   cp temp-project/src/utils/health.js your-project/src/utils/
   cp temp-project/src/utils/graceful-shutdown.js your-project/src/utils/
   ```

4. **Integrate into your server code** (see examples above)

5. **Test locally**

   ```bash
   docker build -t my-bff .
   docker run -p 8080:8080 my-bff
   curl http://localhost:8080/healthz
   ```

6. **Clean up**
   ```bash
   rm -rf temp-project
   ```

---

## Summary

Infrastructure scaffolding in bffgen provides:

✅ **Time Savings:** 2-3 hours of setup automated  
✅ **Production Ready:** Security, optimization, monitoring built-in  
✅ **Consistent:** Same patterns for Go and Node.js  
✅ **Flexible:** Opt-in features, easy to customize  
✅ **Best Practices:** Industry-standard configurations

**Quick Start:**

```bash
# Full production setup in one command
bffgen init my-bff --lang go --include-all-infra

# Or pick what you need
bffgen init my-bff --lang nodejs-express --include-ci --include-docker --include-health
```

For more information, see:

- [Main README](../README.md)
- [Enhanced Scaffolding Guide](./ENHANCED_SCAFFOLDING.md)
- [Node.js Aggregation Guide](./NODEJS_AGGREGATION.md)
