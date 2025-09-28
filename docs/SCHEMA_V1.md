# bffgen Configuration Schema v1

This document describes the bffgen configuration schema version 1.0, which provides a comprehensive, validated configuration format for generating Backend-for-Frontend (BFF) services.

## Overview

The bffgen v1 schema defines a structured configuration format that covers all aspects of BFF service generation, including:

- Project metadata and framework selection
- Server configuration and timeouts
- Authentication and security settings
- CORS and middleware configuration
- Service definitions and endpoint mapping
- Aggregator patterns for data composition
- Logging, monitoring, and observability
- Build and deployment settings

## Schema Validation

The configuration is validated against a JSON Schema (Draft 7) to ensure:

- Required fields are present
- Data types are correct
- Values are within acceptable ranges
- Enums are respected
- Patterns are matched

## Configuration Structure

### Root Level

```yaml
version: "1.0" # Required: Schema version
project:# Required:
  Project configuration
  # ... project settings
server:# Optional:
  Server configuration
  # ... server settings
auth:# Optional:
  Authentication configuration
  # ... auth settings
# ... other sections
```

### Project Configuration

```yaml
project:
  name: "my-bff" # Required: Project name (alphanumeric, hyphens, underscores)
  description: "My BFF service" # Optional: Project description
  version: "1.0.0" # Optional: Project version (semantic versioning)
  language: "go" # Optional: Target language (default: "go")
  framework: "chi" # Required: HTTP framework ("chi", "echo", "fiber")
  output: # Optional: Output configuration
    directory: "." # Output directory (default: ".")
    package: "mybff" # Go package name
    module: "github.com/example/my-bff" # Go module name
```

### Server Configuration

```yaml
server:
  port: 8080 # Server port (1-65535)
  host: "0.0.0.0" # Server host
  timeout: # Timeout configuration
    read: "30s" # Read timeout
    write: "30s" # Write timeout
    idle: "120s" # Idle timeout
  graceful_shutdown: # Graceful shutdown
    enabled: true # Enable graceful shutdown
    timeout: "30s" # Shutdown timeout
```

### Authentication Configuration

```yaml
auth:
  mode: "jwt" # Auth mode: "none", "jwt", "session", "oauth2"
  jwt: # JWT configuration
    secret: "your-secret-key" # JWT signing secret
    expiration: "15m" # Token expiration
    refresh_expiration: "24h" # Refresh token expiration
    encryption: # JWT encryption (JWE)
      enabled: true # Enable encryption
      algorithm: "AES-GCM" # Encryption algorithm
  session: # Session configuration
    store: "memory" # Session store: "memory", "redis"
    expiration: "24h" # Session expiration
    secure: true # Use secure cookies
  csrf: # CSRF protection
    enabled: true # Enable CSRF protection
    header: "X-CSRF-Token" # CSRF header name
```

### CORS Configuration

```yaml
cors:
  enabled: true # Enable CORS
  origins: # Allowed origins
    - "http://localhost:3000"
    - "https://myapp.com"
  methods: # Allowed HTTP methods
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "OPTIONS"
  headers: # Allowed headers
    - "Accept"
    - "Authorization"
    - "Content-Type"
    - "X-CSRF-Token"
  credentials: true # Allow credentials
  max_age: 86400 # Preflight cache duration (seconds)
```

### Security Configuration

```yaml
security:
  headers: # Security headers
    enabled: true # Enable security headers
    content_type_options: "nosniff" # X-Content-Type-Options
    frame_options: "DENY" # X-Frame-Options
    xss_protection: "1; mode=block" # X-XSS-Protection
    referrer_policy: "strict-origin-when-cross-origin" # Referrer-Policy
    permissions_policy: "geolocation=(), microphone=(), camera=()" # Permissions-Policy
  rate_limiting: # Rate limiting
    enabled: true # Enable rate limiting
    requests_per_minute: 100 # Requests per minute per IP
    burst: 10 # Burst limit
    store: "memory" # Rate limiting store: "memory", "redis"
  request_validation: # Request validation
    max_body_size: "10mb" # Maximum request body size
    allowed_content_types: # Allowed content types
      - "application/json"
      - "application/x-www-form-urlencoded"
      - "multipart/form-data"
```

### Service Configuration

```yaml
services:
  auth: # Service name
    base_url: "http://localhost:3000/api" # Required: Service base URL
    timeout: "30s" # Service timeout
    retries: 3 # Retry attempts (0-5)
    circuit_breaker: # Circuit breaker
      enabled: true # Enable circuit breaker
      failure_threshold: 5 # Failure threshold
      recovery_timeout: "60s" # Recovery timeout
    endpoints: # Service endpoints
      - name: "login" # Endpoint name
        path: "/auth/login" # Backend path
        method: "POST" # HTTP method
        expose_as: "/api/auth/login" # Frontend-facing path
        auth_required: false # Require authentication
        cache: # Response caching
          enabled: true # Enable caching
          ttl: "5m" # Cache TTL
          key_template: "auth:{user_id}" # Cache key template
        transform: # Request/response transformation
          request:# Request transformation rules
            # ... transformation config
          response:# Response transformation rules
            # ... transformation config
```

### Aggregator Configuration

```yaml
aggregators:
  - name: "user_dashboard" # Aggregator name
    endpoint: "/api/dashboard" # Frontend endpoint
    method: "GET" # HTTP method
    auth_required: true # Require authentication
    services: # Service calls
      - service: "user" # Service name
        endpoint: "get_user" # Endpoint name
        required: true # Required for aggregator
        timeout: "10s" # Override service timeout
      - service: "content" # Another service
        endpoint: "get_posts" # Endpoint name
        required: false # Optional for aggregator
    response: # Response configuration
      merge_strategy: "merge" # Merge strategy: "merge", "array", "custom"
      template: | # Custom response template
        {
          "user": {{.user}},
          "recent_posts": {{.content.posts}}
        }
```

### Logging Configuration

```yaml
logging:
  level: "info" # Log level: "debug", "info", "warn", "error"
  format: "json" # Log format: "json", "text"
  output: "stdout" # Output: "stdout", "stderr", "file"
  file: # File logging
    path: "logs/app.log" # Log file path
    max_size: "100mb" # Maximum file size
    max_backups: 3 # Maximum backup files
    max_age: 28 # Maximum age in days
  request_logging: # Request logging
    enabled: true # Enable request logging
    log_body: false # Log request/response bodies
    log_headers: true # Log request headers
```

### Monitoring Configuration

```yaml
monitoring:
  health_check: # Health check
    enabled: true # Enable health check
    path: "/health" # Health check path
    checks: # Health checks
      - "database" # Database check
      - "redis" # Redis check
      - "external_services" # External services check
  metrics: # Metrics
    enabled: true # Enable metrics
    path: "/metrics" # Metrics path
    prometheus: true # Enable Prometheus metrics
  tracing: # Distributed tracing
    enabled: false # Enable tracing
    jaeger: # Jaeger configuration
      endpoint: "http://localhost:14268/api/traces" # Jaeger endpoint
      service_name: "my-bff" # Service name
```

### Middleware Configuration

```yaml
middleware:
  - name: "request_logger" # Middleware name
    type: "logging" # Middleware type
    enabled: true # Enable middleware
    order: 1 # Execution order (lower = earlier)
    config: # Middleware-specific config
      format: "json"
      level: "info"
  - name: "rate_limiter" # Another middleware
    type: "rate_limit"
    enabled: true
    order: 2
    config:
      requests_per_minute: 100
      burst: 10
```

### Environment Configuration

```yaml
environment:
  variables: # Environment variables
    JWT_SECRET: "your-secret-key" # JWT secret
    REDIS_URL: "redis://localhost:6379" # Redis URL
    LOG_LEVEL: "info" # Log level
  files: # Environment files
    - ".env" # Main environment file
    - ".env.local" # Local overrides
```

### Build Configuration

```yaml
build:
  go_version: "1.21" # Minimum Go version
  build_tags: [] # Build tags
  ldflags: [] # Linker flags
  cgo_enabled: false # Enable CGO
```

### Deployment Configuration

```yaml
deployment:
  docker: # Docker configuration
    enabled: true # Generate Dockerfile
    base_image: "golang:1.21-alpine" # Base image
    final_image: "alpine:latest" # Final runtime image
    multi_stage: true # Use multi-stage build
  kubernetes: # Kubernetes configuration
    enabled: false # Generate K8s manifests
    namespace: "default" # K8s namespace
    replicas: 3 # Number of replicas
    resources: # Resource limits
      requests: # Resource requests
        cpu: "100m" # CPU request
        memory: "128Mi" # Memory request
      limits: # Resource limits
        cpu: "500m" # CPU limit
        memory: "512Mi" # Memory limit
```

## Default Values

The schema provides sensible defaults for all optional fields:

- **Framework**: `chi`
- **Port**: `8080`
- **Auth Mode**: `jwt`
- **CORS**: Enabled with localhost origins
- **Security**: All security features enabled
- **Logging**: JSON format, info level
- **Rate Limiting**: 100 requests/minute, 10 burst
- **Timeouts**: 30s read/write, 120s idle

## Validation

The configuration is validated against the JSON Schema to ensure:

1. **Required fields** are present
2. **Data types** are correct
3. **Enums** are respected
4. **Patterns** are matched (e.g., version strings, URLs)
5. **Ranges** are valid (e.g., port numbers, timeouts)

## Usage

### Command Line

```bash
# Validate a configuration file
bffgen validate config.yaml

# Generate code from validated configuration
bffgen generate --config config.yaml
```

### Programmatic

```go
import (
    "github.com/RichGod93/bffgen/internal/validation"
    "github.com/RichGod93/bffgen/internal/types"
)

// Create validator
validator, err := validation.NewSchemaValidator()
if err != nil {
    log.Fatal(err)
}

// Validate configuration
config, err := validator.ValidateYAMLFile("bffgen.yaml")
if err != nil {
    log.Fatal(err)
}

// Use validated configuration
fmt.Printf("Project: %s\n", config.Project.Name)
```

## Migration from Legacy Format

The v1 schema is backward compatible with the legacy `bff.config.yaml` format. Migration tools are provided to convert existing configurations:

```bash
# Migrate legacy configuration
bffgen migrate --from bff.config.yaml --to bffgen.yaml
```

## Examples

See the `examples/` directory for complete configuration examples:

- `bffgen-v1.yaml` - Comprehensive example with all features
- `minimal.yaml` - Minimal configuration
- `auth-focused.yaml` - Authentication-focused configuration
- `microservices.yaml` - Multi-service configuration

## Schema Reference

The complete JSON Schema is available at `schemas/bffgen-v1.json` and can be used for:

- IDE validation and autocomplete
- Custom validation tools
- Documentation generation
- Configuration testing
