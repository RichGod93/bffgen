# Gateway BFF - API Gateway for Microservices

A production-ready Backend-for-Frontend (BFF) gateway that proxies requests to multiple microservices with authentication, rate limiting, circuit breakers, and response aggregation.

## 🏗️ Architecture

```
┌─────────────┐
│   Frontend  │
│ (Port 3000) │
└──────┬──────┘
       │
       ▼
┌──────────────────────────────────┐
│     Gateway BFF (Port 8080)      │
│  - Authentication (JWT)          │
│  - Rate Limiting                 │
│  - CORS Handling                 │
│  - Circuit Breakers              │
│  - Response Aggregation          │
└──────┬───────────┬───────────┬───┘
       │           │           │
       ▼           ▼           ▼
┌──────────┐ ┌────────────┐ ┌─────────────────┐
│  Users   │ │ Analytics  │ │ Notifications   │
│   4000   │ │    4001    │ │      4002       │
└──────────┘ └────────────┘ └─────────────────┘
```

## 📋 Services

### Users Service (Port 4000)

Manages user data and profiles.

**Endpoints via Gateway:**

- `GET /api/users` - List all users
- `POST /api/users` - Create user
- `GET /api/users/{id}` - Get user by ID
- `PUT /api/users/{id}` - Update user
- `DELETE /api/users/{id}` - Delete user

### Analytics Service (Port 4001)

Provides metrics and event tracking.

**Endpoints via Gateway:**

- `GET /api/analytics/metrics` - Get system metrics (auth required)
- `GET /api/analytics/events` - List events (auth required)
- `POST /api/analytics/events` - Create event (auth required)

### Notifications Service (Port 4002)

Handles user notifications.

**Endpoints via Gateway:**

- `GET /api/notifications` - List notifications (auth required)
- `GET /api/notifications/{id}` - Get notification by ID (auth required)
- `POST /api/notifications/{id}/read` - Mark as read (auth required)
- `POST /api/notifications/read-all` - Mark all as read (auth required)

### Aggregated Endpoints

#### Dashboard Endpoint

`GET /api/dashboard` (auth required)

Aggregates data from multiple services into a single response:

- User profile (from Users service)
- System metrics (from Analytics service)
- User notifications (from Notifications service)

Response format:

```json
{
  "user": { "id": "1", "name": "Alice", ... },
  "metrics": { "active_users": 1234, ... },
  "notifications": { "notifications": [...], "unread": 3 }
}
```

## 🚀 Quick Start

### 1. Start Mock Backend Services

```bash
cd mock-services

# Option A: Use helper script
./run-all-services.sh

# Option B: Run individually
go run users-service.go &
go run analytics-service.go &
go run notifications-service.go &
```

### 2. Set Environment Variables

```bash
export JWT_SECRET=your-super-secure-secret-key-change-in-production
export ENCRYPTION_KEY=your-base64-encoded-32-byte-key
export REDIS_URL=redis://localhost:6379  # Optional for rate limiting
```

### 3. Start the Gateway BFF

```bash
# From project root
go run main.go

# Or use bffgen dev command
bffgen dev
```

### 4. Test the Gateway

```bash
# Health check
curl http://localhost:8080/health

# List users (no auth required)
curl http://localhost:8080/api/users

# Get metrics (requires auth)
curl -H "Authorization: Bearer <your-jwt-token>" \
  http://localhost:8080/api/analytics/metrics

# Aggregated dashboard (requires auth)
curl -H "Authorization: Bearer <your-jwt-token>" \
  http://localhost:8080/api/dashboard
```

## 🔐 Authentication

The gateway uses JWT authentication with encryption:

### Generate a JWT Token

```go
import "github.com/golang-jwt/jwt/v5"

claims := jwt.MapClaims{
    "user_id": "1",
    "email": "user@example.com",
    "exp": time.Now().Add(time.Hour * 24).Unix(),
}

token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
```

### Using the Token

```bash
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  http://localhost:8080/api/users/1
```

## 📦 Project Structure

```
gateway-bff/
├── main.go                         # Main server entry point
├── bff.config.yaml                 # Gateway configuration
├── go.mod                          # Go dependencies
├── bff-postman-collection.json     # Postman API collection
├── PROJECT_OVERVIEW.md             # This file
├── README.md                       # Generated docs
│
├── .github/workflows/
│   └── ci.yml                      # CI/CD pipeline
│
├── Dockerfile                      # Production container
├── .dockerignore
├── docker-compose.yml              # Dev environment
│
├── internal/
│   ├── auth/                       # JWT authentication
│   ├── routes/                     # Route handlers
│   ├── aggregators/                # Response aggregation
│   ├── health/                     # Health checks
│   └── templates/                  # Route templates
│
└── mock-services/
    ├── users-service.go            # Mock users API
    ├── analytics-service.go        # Mock analytics API
    ├── notifications-service.go    # Mock notifications API
    ├── run-all-services.sh         # Helper script
    └── README.md                   # Mock services docs
```

## 🔧 Configuration

The gateway is configured via `bff.config.yaml`. Key sections:

### Server Settings

```yaml
server:
  port: 8080
  timeout:
    read: "30s"
    write: "30s"
```

### Service Configuration

```yaml
services:
  users:
    baseUrl: "http://localhost:4000/api"
    timeout: "30s"
    retries: 3
    circuit_breaker:
      enabled: true
      failure_threshold: 5
```

### CORS

```yaml
cors:
  enabled: true
  origins:
    - "http://localhost:3000"
    - "http://localhost:5173"
```

## 🧪 Testing with Postman

1. Import `bff-postman-collection.json` into Postman
2. Set the `baseUrl` variable to `http://localhost:8080`
3. For authenticated endpoints, set the `token` variable with a valid JWT
4. Run the collection to test all endpoints

## 🐳 Docker Deployment

### Build and Run

```bash
# Build image
docker build -t gateway-bff:latest .

# Run container
docker run -p 8080:8080 \
  -e JWT_SECRET=your-secret \
  -e ENCRYPTION_KEY=your-key \
  gateway-bff:latest
```

### Docker Compose

```bash
# Start all services (Gateway + Mock backends)
docker-compose up

# Stop all services
docker-compose down
```

## 📊 Monitoring

### Health Checks

```bash
# Liveness probe
curl http://localhost:8080/healthz

# Readiness probe (includes dependency checks)
curl http://localhost:8080/health
```

### Prometheus Metrics

```bash
# Metrics endpoint
curl http://localhost:8080/metrics
```

## 🔒 Security Features

- **JWT Authentication** - Token-based auth with encryption
- **Rate Limiting** - Configurable per-endpoint rate limits
- **Circuit Breakers** - Automatic failure recovery
- **CORS Protection** - Restrictive origin policies
- **Security Headers** - CSP, XSS protection, frame options
- **Request Validation** - Body size limits, content-type checks
- **CSRF Protection** - Token-based CSRF prevention

## 🛠️ Development Commands

```bash
# Add new routes
bffgen add-route

# Add templates (auth, ecommerce, content)
bffgen add-template auth

# Generate code from config
bffgen generate

# Validate configuration
bffgen config validate

# Check project health
bffgen doctor
```

## 📚 Key Features

✅ **Microservices Architecture** - Route to multiple backend services  
✅ **Response Aggregation** - Combine data from multiple services  
✅ **Circuit Breakers** - Prevent cascade failures  
✅ **JWT Authentication** - Secure token-based auth  
✅ **Rate Limiting** - Per-service rate limits  
✅ **CORS Support** - Frontend integration ready  
✅ **Health Checks** - Kubernetes-compatible probes  
✅ **Prometheus Metrics** - Production observability  
✅ **Docker Support** - Production-ready containers  
✅ **CI/CD Pipeline** - GitHub Actions workflow

## 🤔 Common Tasks

### Adding a New Service

1. Update `bff.config.yaml`:

```yaml
services:
  new_service:
    baseUrl: "http://localhost:4003/api"
    endpoints:
      - name: "get_data"
        path: "/data"
        method: "GET"
        exposeAs: "/api/data"
        authRequired: false
```

2. Regenerate code:

```bash
bffgen generate
```

### Adding an Aggregator

1. Edit `bff.config.yaml`:

```yaml
aggregators:
  - name: "my_aggregator"
    endpoint: "/api/combined"
    method: "GET"
    authRequired: true
    services:
      - service: "users"
        endpoint: "get_user"
      - service: "analytics"
        endpoint: "get_metrics"
```

2. Regenerate and restart.

## 📖 Additional Resources

- [bffgen Documentation](https://github.com/RichGod93/bffgen)
- [Mock Services Guide](mock-services/README.md)
- [BFF Pattern Guide](https://martinfowler.com/articles/bff.html)

## 🎯 Next Steps

1. ✅ Start mock services
2. ✅ Start the gateway
3. ✅ Test with Postman collection
4. 🔄 Replace mock services with real backends
5. 🔄 Deploy to production with Docker
6. 🔄 Set up monitoring and alerting
