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
| Async/Await | ✅ | Fully asynchronous by default |
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
| Circuit Breaker | ✅ | Prevent cascading failures |
| Caching | ✅ | Redis-based caching (optional) |
| Request ID | ✅ | Request tracking |
| Error Handling | ✅ | Standardized error responses |
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

Python projects use a JSON configuration file:

```json
{
  "project": {
    "name": "my-bff",
    "framework": "fastapi",
    "async": true
  },
  "backends": [
    {
      "name": "users",
      "baseUrl": "http://localhost:3001",
      "timeout": 30,
      "endpoints": [
        {
          "name": "get_users",
          "method": "GET",
          "path": "/api/users",
          "upstreamPath": "/users",
          "description": "Get all users"
        },
        {
          "name": "create_user",
          "method": "POST",
          "path": "/api/users",
          "upstreamPath": "/users",
          "description": "Create a new user"
        }
      ]
    }
  ]
}
```

### config.py (Pydantic Settings)

```python
from pydantic_settings import BaseSettings
from typing import List, Union

class Settings(BaseSettings):
    # Application
    PROJECT_NAME: str = "my-bff"
    PORT: int = 8000
    DEBUG: bool = True
    
    # CORS
    CORS_ORIGINS: Union[List[str], str] = ["http://localhost:3000"]
    
    # JWT
    JWT_SECRET: str = "your-secret-key"
    JWT_ALGORITHM: str = "HS256"
    
    # Rate Limiting
    RATE_LIMIT_ENABLED: bool = True
    RATE_LIMIT_PER_MINUTE: int = 60
    
    # Caching (optional)
    REDIS_URL: str = "redis://localhost:6379"
    CACHE_ENABLED: bool = False
    CACHE_TTL: int = 300
    
    # Circuit Breaker
    CIRCUIT_BREAKER_ENABLED: bool = True
    CIRCUIT_BREAKER_FAILURE_THRESHOLD: int = 5
    CIRCUIT_BREAKER_TIMEOUT_SECONDS: int = 60
    
    class Config:
        env_file = ".env"

settings = Settings()
```

### Environment Variables (.env)

```bash
PROJECT_NAME=my-bff
PORT=8000
DEBUG=true

CORS_ORIGINS=http://localhost:3000,http://localhost:8080

JWT_SECRET=your-secret-key-change-in-production
JWT_ALGORITHM=HS256

RATE_LIMIT_ENABLED=true
RATE_LIMIT_PER_MINUTE=60

REDIS_URL=redis://localhost:6379
CACHE_ENABLED=false

CIRCUIT_BREAKER_ENABLED=true
```

---

## Code Generation

### Generate Routers and Services

```bash
# Generate from config
bffgen generate

# What gets generated:
# - routers/{backend}_router.py
# - services/{backend}_service.py
# - Updates to main.py (router registration)
```

### Generated Router Example

```python
"""
Users Router - Generated by bffgen
"""
from fastapi import APIRouter, HTTPException, Body
from typing import Dict, Any
import logging

from services.users_service import UsersService

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/api/users", tags=["Users"])
users_service = UsersService()


@router.get("")
async def get_users():
    """Get all users"""
    try:
        result = await users_service.get_users()
        return result
    except Exception as e:
        logger.error(f"Error fetching users: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.post("")
async def create_user(payload: Dict[str, Any] = Body(...)):
    """Create a new user"""
    try:
        result = await users_service.create_user(payload=payload)
        return result
    except Exception as e:
        logger.error(f"Error creating user: {e}")
        raise HTTPException(status_code=500, detail=str(e))
```

### Generated Service Example

```python
"""
Users Service - Generated by bffgen
HTTP client for Users API backend
"""
import httpx
import logging
from typing import Dict, Any

logger = logging.getLogger(__name__)


class UsersService:
    """Service class for Users API operations"""
    
    def __init__(self):
        self.base_url = "http://localhost:3001"
        self.timeout = 30.0
        self.client = None
    
    async def _get_client(self) -> httpx.AsyncClient:
        """Get or create HTTP client"""
        if self.client is None:
            self.client = httpx.AsyncClient(timeout=self.timeout)
        return self.client
    
    async def get_users(self) -> Dict[str, Any]:
        """Get all users"""
        client = await self._get_client()
        url = f"{self.base_url}/users"
        
        logger.info(f"Fetching users from {url}")
        response = await client.get(url)
        response.raise_for_status()
        
        return response.json()
    
    async def create_user(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        """Create a new user"""
        client = await self._get_client()
        url = f"{self.base_url}/users"
        
        logger.info(f"Creating user at {url}")
        response = await client.post(url, json=payload)
        response.raise_for_status()
        
        return response.json()
    
    async def close(self):
        """Close HTTP client"""
        if self.client:
            await self.client.aclose()
```

---

## Advanced Features

### Circuit Breaker

Prevents cascading failures when backend services are down:

```python
from utils.circuit_breaker import CircuitBreaker

# Create circuit breaker
breaker = CircuitBreaker("my-service", failure_threshold=5, timeout=60)

# Use it
async def call_backend():
    return await breaker.call(some_async_function, *args, **kwargs)
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
from jose import jwt
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

See the full example project: [Movie Dashboard BFF](../projects/python/movie-dashboard-bff/)

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

```python
# tests/test_users_router.py
import pytest
from fastapi.testclient import TestClient
from unittest.mock import AsyncMock, patch

from main import app

client = TestClient(app)

async def test_get_users():
    """Test getting users"""
    with patch('services.users_service.UsersService.get_users', new_callable=AsyncMock) as mock:
        mock.return_value = {"users": [{"id": 1, "name": "Alice"}]}
        
        response = client.get("/api/users")
        
        assert response.status_code == 200
        assert len(response.json()["users"]) == 1
```

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
- **Example Project**: [Movie Dashboard BFF](../projects/python/movie-dashboard-bff/)

---

## Next Steps

1. ✅ Initialize your first Python/FastAPI project
2. ✅ Configure backends in `bffgen.config.py.json`
3. ✅ Generate routers and services with `bffgen generate`
4. ✅ Customize with middleware and utilities
5. ✅ Write tests with pytest
6. ✅ Deploy to production

**Happy coding with bffgen + Python! 🐍🚀**

