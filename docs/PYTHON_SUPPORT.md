# Python/FastAPI Support

**bffgen** now supports generating Backend-for-Frontend services using **Python** and **FastAPI**! This guide covers everything you need to know about using bffgen with Python.

---

## 📋 Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Features](#features)
- [Project Structure](#project-structure)
- [Configuration](#configuration)
- [Code Generation](#code-generation)
- [Advanced Features](#advanced-features)
- [Examples](#examples)
- [Comparison with Other Runtimes](#comparison-with-other-runtimes)
- [Testing](#testing)
- [Deployment](#deployment)

---

## Overview

bffgen's Python support provides:

- ✅ **FastAPI framework** - Modern, fast (high-performance) web framework
- ✅ **Async/await** - Fully asynchronous endpoints by default
- ✅ **Type safety** - Pydantic models for request/response validation
- ✅ **Auto-generated docs** - Interactive OpenAPI/Swagger documentation
- ✅ **Production-ready** - Circuit breakers, caching, middleware
- ✅ **Test infrastructure** - pytest setup with async support

---

## Quick Start

### Initialize a Python Project

```bash
# Basic initialization
bffgen init my-bff --lang python-fastapi

# With options
bffgen init my-bff \
  --lang python-fastapi \
  --async=true \
  --pkg-manager pip

cd my-bff
```

### Setup and Run

```bash
# Install dependencies
./setup.sh

# Start the server
source venv/bin/activate
uvicorn main:app --reload
```

### Access the Application

- **API Docs**: http://localhost:8000/docs
- **Alternative Docs**: http://localhost:8000/redoc
- **Health Check**: http://localhost:8000/health

---

## Features

### Core Features

| Feature | Status | Description |
|---------|--------|-------------|
| FastAPI Framework | ✅ | Modern async web framework |
| Async/Await | ✅ | Generated routes and the HTTP client are async |
| Type Hints | ✅ | Complete type annotations |
| Pydantic Models | ✅ | Data validation and serialization |
| OpenAPI Docs | ✅ | Auto-generated interactive docs |
| CORS Support | ✅ | Configurable CORS middleware |
| JWT Auth | ✅ | JWT authentication middleware |
| Rate Limiting | ✅ | Request rate limiting (slowapi) |
| Logging | ✅ | Structured logging middleware |

### Advanced Features

| Feature | Status | Description |
|---------|--------|-------------|
| Proxy Routes | ✅ | Path params, query strings, auth, 4xx passthrough |
| Timeouts & Retries | ✅ | Per backend; retries only for idempotent methods |
| Circuit Breaker | ✅ | One per backend, used by generated routes |
| Caching | ✅ | Redis-based caching (optional) |
| Request ID | ✅ | Request tracking |
| Error Handling | ✅ | 502/503/504 with a consistent JSON body |
| Health Checks | ✅ | Service health endpoints |

### Package Managers

- **pip** (default) - requirements.txt
- **Poetry** - pyproject.toml (coming soon)

---

## Project Structure

```
my-bff/
├── main.py                    # FastAPI application entry point
├── config.py                  # Pydantic Settings configuration
├── dependencies.py            # Dependency injection
├── logger.py                  # Logging setup
├── requirements.txt           # Python dependencies
├── setup.sh                   # Setup script
├── .env                       # Environment variables
├── .gitignore                 # Python-specific gitignore
│
├── routers/                   # API route handlers
│   ├── __init__.py
│   ├── users_router.py        # Generated router
│   └── products_router.py     # Generated router
│
├── services/                  # Backend service clients
│   ├── __init__.py
│   ├── users_service.py       # HTTP client for users API
│   └── products_service.py    # HTTP client for products API
│
├── models/                    # Pydantic models (optional)
│   └── __init__.py
│
├── middleware/                # Custom middleware
│   ├── __init__.py
│   ├── auth_middleware.py     # JWT authentication
│   └── logging_middleware.py  # Request/response logging
│
├── utils/                     # Utility modules
│   ├── __init__.py
│   ├── cache_manager.py       # Redis caching
│   └── circuit_breaker.py     # Circuit breaker pattern
│
├── tests/                     # Test suite
│   ├── __init__.py
│   ├── conftest.py            # Pytest fixtures
│   ├── pytest.ini             # Pytest configuration
│   └── test_*.py              # Test files
│
└── bffgen.config.py.json      # bffgen configuration
```

---

## Configuration

### bffgen.config.py.json

Each backend lists the routes the BFF exposes (`path`) and where they go on the backend
(`upstreamPath`, default: the same path). Path parameters can be written `{id}` or `:id`.

```json
{
  "project": { "name": "my-bff", "framework": "fastapi", "async": true },
  "backends": [
    {
      "name": "users",
      "baseUrl": "http://localhost:3001",
      "timeout": 5000,
      "retries": 2,
      "circuitBreaker": { "enabled": true, "failureThreshold": 5, "resetTimeout": 30000 },
      "endpoints": [
        { "name": "list_users", "method": "GET", "path": "/api/users", "upstreamPath": "/users" },
        { "name": "get_user", "method": "GET", "path": "/api/users/{id}", "upstreamPath": "/users/{id}", "requiresAuth": true },
        { "name": "create_user", "method": "POST", "path": "/api/users", "upstreamPath": "/users", "requiresAuth": true }
      ]
    }
  ]
}
```

| Field | Meaning | Default |
|---|---|---|
| `baseUrl` (or `url`) | Backend base URL | required |
| `timeout` | Per-attempt timeout in **ms**, or a duration string such as `"5s"` | `10000` |
| `retries` | Extra attempts for `GET`, `HEAD` and `OPTIONS` requests | `2` |
| `circuitBreaker.enabled` | Use a circuit breaker for this backend | `true` |
| `circuitBreaker.failureThreshold` | Consecutive failures that open the circuit | `5` |
| `circuitBreaker.resetTimeout` | Time in **ms** the circuit stays open before a trial request (`recoveryTimeout` is accepted too) | `30000` |
| `endpoints[].requiresAuth` | Require a valid JWT (`Authorization: Bearer …`); otherwise 401 | `false` |

Times are in milliseconds, as in Node.js projects. Earlier Python examples used seconds
(`"timeout": 30`), so `bffgen generate` rejects times under 100 ms and suggests the intended value.

Node.js-style endpoints (`exposeAs` for the exposed path, `path` for the backend path) are
accepted too.

### config.py and .env

`config.py` defines the app settings (port, CORS, JWT, rate limiting, Redis) with Pydantic
Settings, read from environment variables and `.env`. Unknown keys in `.env` are ignored, so
`.env` can also hold the backend overrides below.

### Backend overrides

Each backend's settings can be overridden with environment variables or `.env` entries. The
prefix is the backend name in upper case with other characters replaced by `_`
(`user-service` → `USER_SERVICE`).

| Variable | Overrides |
|---|---|
| `USERS_URL` (or `USERS_SERVICE_URL`) | `baseUrl` |
| `USERS_TIMEOUT` | `timeout` (ms) |
| `USERS_RETRIES` | `retries` |
| `USERS_CIRCUIT_FAILURE_THRESHOLD` | `circuitBreaker.failureThreshold` |
| `USERS_CIRCUIT_RESET_TIMEOUT` | `circuitBreaker.resetTimeout` (ms) |

---

## Code Generation

### Generate Routers and Services

```bash
bffgen generate
```

This writes, for each backend with endpoints:

- `routers/<backend>_router.py`: one route per endpoint that proxies to the backend
- `services/<backend>_service.py`: typed async methods for calling the backend from your own code
- `routers/__init__.py`: the list of routers, which `main.py` registers with `include_routers(app)`

These files are overwritten on every run; keep your own code in other files.

`bffgen generate` also creates or upgrades the runtime files the routers use:
`utils/upstream.py`, `services/http_client.py` and `utils/circuit_breaker.py`. A file is
replaced only if its `bffgen:runtime` marker is older than the current one, and your previous
copy is kept as `<file>.bak`. `--dry-run` and `--check` only report what would change.

### Generated Router Example

```python
router = APIRouter(tags=["users"])
upstream = create_upstream(
    "users",
    env_key="USERS",
    base_url="http://localhost:3001",
    timeout=5000,
    retries=2,
    circuit_breaker={"enabled": True, "failureThreshold": 5, "resetTimeout": 30000},
)


# GET /api/users/{id} -> GET /users/{id}
@router.api_route(
    "/api/users/{id}",
    methods=["GET", "HEAD"],
    dependencies=[Depends(require_authentication)],
)
async def get_user(request: Request, id: str) -> Response:
    return await upstream.proxy(request, "GET", "/users/{id}")
```

### What generated routes do

They behave the same as generated Express and Fastify routes
(see [Generated Node.js Routes](NODEJS_ROUTES.md) for details):

- **Forwarding:** path parameters are substituted and URL-encoded, and the query string is passed
  on as received. `authorization`, `accept`, `accept-language`, conditional headers and
  `x-request-id` are forwarded; cookies are not. Request bodies are forwarded for every method
  except `GET` and `HEAD`.
- **Backend responses below 500** are passed back unchanged, with their `content-type`,
  `cache-control`, `etag`, `last-modified` and `location` headers.
- **Timeouts and retries:** each attempt has its own timeout. Only `GET`, `HEAD` and `OPTIONS`
  are retried (after timeouts, connection failures and 5xx responses), so a `POST` is never sent
  twice.
- **Circuit breaker:** one per backend, shared by its routes. After `failureThreshold`
  consecutive failures, requests are rejected immediately until `resetTimeout` has passed; then
  one trial request decides whether it closes again. 4xx responses don't count as failures.
- **Errors:**

  | Status | `code` | When |
  |---|---|---|
  | `504 Gateway Timeout` | `UPSTREAM_TIMEOUT` | No response within the timeout, on every attempt |
  | `503 Service Unavailable` | `UPSTREAM_UNAVAILABLE` | The backend could not be reached |
  | `503 Service Unavailable` | `CIRCUIT_OPEN` | The circuit is open; `Retry-After` gives the seconds left |
  | `502 Bad Gateway` | `UPSTREAM_ERROR` | The backend answered with a 5xx status (its body is not passed on) |

  The body is `{"error", "message", "code", "service"}`.

One difference from Express and Fastify: Starlette decodes the URL path before routing, so a
path parameter can't contain an encoded slash (`%2F`); such requests get 404.

### Generated Service Example

```python
from services.users_service import UsersService

users = UsersService()
user = await users.get_user(id="42", query={"fields": "name"})
```

Service methods return the parsed body and raise `HttpError` for 4xx responses, or
`UpstreamError` / `CircuitOpenError` for failures. Those errors carry a `status_code` of
502, 503 or 504.

---

## Advanced Features

### Circuit Breaker

Generated routes already use one breaker per backend. For your own calls:

```python
from utils.circuit_breaker import CircuitBreaker, CircuitOpenError

breaker = CircuitBreaker(failure_threshold=5, timeout=30, name="inventory")  # timeout in seconds

try:
    result = await breaker.call(some_async_function, *args, **kwargs)
except CircuitOpenError as error:
    ...  # error.retry_after: seconds until the next trial request
```

### Caching (Redis)

```python
from utils.cache_manager import cache

# In service method
async def get_popular_movies(self, page: int = 1):
    cache_key = f"popular:movies:page{page}"
    
    # Try cache first
    cached = await cache.get(cache_key)
    if cached:
        return cached
    
    # Fetch from API
    result = await self._fetch_from_api(page)
    
    # Cache it
    await cache.set(cache_key, result, ttl=300)
    
    return result
```

### Custom Middleware

```python
# middleware/timing_middleware.py
import time
from starlette.middleware.base import BaseHTTPMiddleware

class TimingMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request, call_next):
        start = time.time()
        response = await call_next(request)
        duration = time.time() - start
        response.headers["X-Process-Time"] = f"{duration:.3f}"
        return response

# main.py
app.add_middleware(TimingMiddleware)
```

### Dependency Injection

```python
# dependencies.py
from fastapi import Header, HTTPException, Depends
import jwt  # PyJWT
from config import settings

async def get_current_user(authorization: str = Header(None)) -> str:
    """Extract user from JWT token"""
    if not authorization or not authorization.startswith("Bearer "):
        raise HTTPException(status_code=401, detail="Missing token")
    
    token = authorization.split(" ")[1]
    payload = jwt.decode(token, settings.JWT_SECRET, algorithms=[settings.JWT_ALGORITHM])
    return payload.get("sub")

# Use in routers
@router.get("/profile")
async def get_profile(user_id: str = Depends(get_current_user)):
    # user_id is automatically injected
    return {"user_id": user_id}
```

---

## Examples

### Example 1: Simple API Gateway

```bash
# Initialize
bffgen init api-gateway --lang python-fastapi

# Configure backends (edit bffgen.config.py.json)
# Generate code
cd api-gateway
bffgen generate

# Run
./setup.sh
source venv/bin/activate
uvicorn main:app --reload
```

### Example 2: Aggregation Pattern

See the full example project: [Movie Dashboard BFF](../examples/projects/python/movie-dashboard-bff/)

This demonstrates:
- Parallel requests to multiple backends
- Data enrichment (combining TMDB + user data)
- Circuit breakers and caching
- Custom aggregation routers

---

## Comparison with Other Runtimes

| Feature | Go | Node.js | Python |
|---------|----|---------| -------|
| **Performance** | Excellent | Very Good | Good |
| **Async Model** | Goroutines | Event Loop | Async/Await |
| **Type Safety** | Strong | TypeScript | Type Hints |
| **Ecosystem** | Large | Massive | Massive |
| **Learning Curve** | Moderate | Easy | Easy |
| **Use Cases** | High-performance, microservices | Full-stack, real-time | Data science, ML, API |

### When to Use Python/FastAPI

**✅ Good for:**
- API gateways with ML/data science backends
- Teams already using Python
- Rapid prototyping
- Integration with Python ML models
- Data transformation and enrichment

**⚠️ Consider alternatives if:**
- Maximum performance is critical
- Very high concurrency (>10k connections)
- Team prefers Go or Node.js

---

## Testing

### Run Tests

```bash
# All tests
pytest

# With coverage
pytest --cov=. --cov-report=html

# Specific tests
pytest tests/test_routers.py

# Verbose
pytest -v
```

### Writing Tests

`tests/conftest.py` provides a `client` fixture and an `auth_headers` fixture with a valid JWT
signed with your `JWT_SECRET`:

```python
# tests/test_users_router.py
from fastapi.testclient import TestClient


def test_get_user_requires_auth(client: TestClient, auth_headers: dict):
    assert client.get("/api/users/42").status_code == 401
    assert client.get("/api/users/42", headers=auth_headers).status_code != 401
```

Generated routes call backends over HTTP. To test responses end to end, run the backends or
point `<NAME>_URL` at a test server before the app is imported.

---

## Deployment

### Docker

```dockerfile
FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

EXPOSE 8000

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
```

### Docker Compose

```yaml
version: '3.8'

services:
  bff:
    build: .
    ports:
      - "8000:8000"
    environment:
      - PORT=8000
      - REDIS_URL=redis://redis:6379
    depends_on:
      - redis
  
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
```

### Production Checklist

- [ ] Set strong `JWT_SECRET`
- [ ] Enable HTTPS
- [ ] Configure proper CORS origins
- [ ] Enable caching (Redis)
- [ ] Set up logging aggregation
- [ ] Configure rate limiting
- [ ] Add health checks
- [ ] Set up monitoring (Prometheus/Grafana)
- [ ] Use environment-specific configs

---

## CLI Reference

### Init Command

```bash
bffgen init <project-name> [flags]

Flags:
  --lang string              Language (python-fastapi)
  --async bool              Generate async endpoints (default: true)
  --pkg-manager string      Package manager: pip or poetry (default: pip)
  --skip-tests              Skip test file generation
  --skip-docs               Skip API documentation
  --include-ci              Generate GitHub Actions workflow
  --include-docker          Generate Dockerfile
  --include-all-infra       Generate all infrastructure files
```

### Generate Command

```bash
bffgen generate

# Generates routers and services from bffgen.config.py.json
```

### Add Template Command

```bash
bffgen add-template <template-name>

# Available templates:
# - auth: Authentication endpoints
# - ecommerce: E-commerce endpoints
# - content: Content management endpoints
```

---

## Troubleshooting

### Common Issues

**Issue: `ModuleNotFoundError`**
```bash
# Solution: Activate virtual environment
source venv/bin/activate
```

**Issue: Port already in use**
```bash
# Solution: Kill process or use different port
uvicorn main:app --port 8001
```

**Issue: Pydantic validation errors**
```bash
# Solution: Check environment variables match config.py
# Make sure .env file is present
```

**Issue: CORS errors**
```bash
# Solution: Add frontend URL to CORS_ORIGINS in .env
CORS_ORIGINS=http://localhost:3000,http://localhost:8080
```

---

## Resources

- **FastAPI Documentation**: https://fastapi.tiangolo.com
- **Pydantic Documentation**: https://docs.pydantic.dev
- **httpx Documentation**: https://www.python-httpx.org
- **pytest-asyncio**: https://pytest-asyncio.readthedocs.io
- **Example Project**: [Movie Dashboard BFF](../examples/projects/python/movie-dashboard-bff/)

---

## Next Steps

1. ✅ Initialize your first Python/FastAPI project
2. ✅ Configure backends in `bffgen.config.py.json`
3. ✅ Generate routers and services with `bffgen generate`
4. ✅ Customize with middleware and utilities
5. ✅ Write tests with pytest
6. ✅ Deploy to production

**Happy coding with bffgen + Python! 🐍🚀**

