# 🎬 Movie Dashboard BFF

A **Backend-for-Frontend (BFF)** service built with **FastAPI** that aggregates movie data from **TMDB API** with user preferences, demonstrating modern API gateway patterns.

> **Generated with [bffgen](https://github.com/RichGod93/bffgen)** - A CLI tool for scaffolding BFF services

---

## ✨ Features

- **🎥 Real TMDB Integration** - Live movie data from The Movie Database API
- **⚡ Async/Await** - Fully asynchronous for optimal performance  
- **🔄 Data Aggregation** - Combines multiple backend services (TMDB + User Service)
- **👤 Personalized Responses** - Enriches movie data with user favorites and ratings
- **🛡️ Resilience Patterns** - Circuit breakers, caching, error handling
- **📊 Auto-Generated Docs** - Interactive OpenAPI/Swagger documentation
- **✅ Type Safety** - Pydantic models for request/response validation
- **🧪 Comprehensive Tests** - Unit, integration, and e2e test suites

---

## 🏗️ Architecture

```
┌─────────────────┐
│                 │
│   Frontend      │
│   (React/Vue)   │
│                 │
└────────┬────────┘
         │
         ├─ GET /api/dashboard/feed
         ├─ GET /api/movies/popular
         └─ POST /api/users/favorites
         │
┌────────▼────────────────────────────┐
│                                     │
│   Movie Dashboard BFF (FastAPI)    │
│   ─────────────────────────────    │
│                                     │
│   ┌──────────────────────────┐    │
│   │  Dashboard Router        │    │
│   │  (Aggregation Logic)     │    │
│   └────┬──────────────┬──────┘    │
│        │              │            │
│   ┌────▼────┐    ┌───▼─────┐     │
│   │ TMDB    │    │  User   │     │
│   │ Client  │    │ Service │     │
│   └────┬────┘    └───┬─────┘     │
└────────┼──────────────┼───────────┘
         │              │
    ┌────▼─────┐  ┌────▼──────────┐
    │  TMDB    │  │ Mock User     │
    │  API     │  │ Service       │
    │ (External│  │ (Port 3001)   │
    └──────────┘  └───────────────┘
```

### Key Patterns Demonstrated

- **BFF Pattern**: Single API optimized for frontend needs
- **Data Aggregation**: Parallel requests to multiple backends
- **Data Enrichment**: Combining data from different sources
- **Circuit Breaker**: Preventing cascading failures
- **Graceful Degradation**: Continues working if user service fails

---

## 🚀 Quick Start

### Prerequisites

- **Python 3.8+** 
- **pip** or **Poetry**
- **TMDB API Key** (included in config)

### 1. Install Dependencies

```bash
# Option A: Using setup script (recommended)
./setup.sh

# Option B: Manual installation
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

### 2. Start Services

**Option A: Start all services together**
```bash
./start-all.sh
```

**Option B: Start services separately**

Terminal 1 - Mock User Service:
```bash
cd mock-user-service
python main.py
```

Terminal 2 - Main BFF:
```bash
source venv/bin/activate
uvicorn main:app --reload
```

### 3. Access the Application

- **📖 API Documentation**: http://localhost:8000/docs
- **📊 Alternative Docs**: http://localhost:8000/redoc  
- **❤️ Health Check**: http://localhost:8000/health
- **👤 Mock User Service**: http://localhost:3001/docs

---

## 📚 API Endpoints

### Dashboard (Aggregated)

These endpoints demonstrate the BFF pattern by combining data from multiple sources:

- `GET /api/dashboard/feed` - Personalized movie feed (popular + user data)
- `GET /api/dashboard/movie/{id}/enriched` - Movie details with user context
- `GET /api/dashboard/complete` - Complete dashboard (popular, trending, stats)
- `GET /api/dashboard/search/enriched` - Search with user context

### Movies (TMDB Proxy)

Direct access to TMDB API through the BFF:

- `GET /api/movies/popular` - Popular movies
- `GET /api/movies/{id}` - Movie details
- `GET /api/movies/search` - Search movies

### Users (User Service Proxy)

User preferences management:

- `GET /api/users/favorites` - Get user's favorites
- `POST /api/users/favorites` - Add to favorites
- `DELETE /api/users/favorites/{id}` - Remove from favorites
- `GET /api/users/watchlist` - Get watchlist
- `POST /api/users/watchlist` - Add to watchlist

---

## 🧪 Example Requests

### Get Personalized Feed

```bash
curl http://localhost:8000/api/dashboard/feed?page=1
```

**Response:**
```json
{
  "movies": [
    {
      "id": 550,
      "title": "Fight Club",
      "vote_average": 8.4,
      "is_favorite": true,
      "is_in_watchlist": false,
      "user_rating": 5
    }
  ],
  "total_pages": 100,
  "current_page": 1,
  "favorites_count": 5,
  "watchlist_count": 3
}
```

### Get Enriched Movie Details

```bash
curl http://localhost:8000/api/dashboard/movie/550/enriched
```

### Search with User Context

```bash
curl "http://localhost:8000/api/dashboard/search/enriched?query=inception"
```

### Add to Favorites

```bash
curl -X POST http://localhost:3001/favorites \
  -H "Content-Type: application/json" \
  -d '{
    "movie_id": 550,
    "title": "Fight Club",
    "rating": 5
  }'
```

**More examples**: See `examples/curl-examples.sh` or `examples/requests.http`

---

## 🧩 Project Structure

```
movie-dashboard-bff/
├── main.py                 # FastAPI application entry point
├── config.py               # Settings and configuration
├── dependencies.py         # Dependency injection
├── requirements.txt        # Python dependencies
├── setup.sh               # Dependency installation script
├── start-all.sh           # Start all services
│
├── routers/               # API route handlers
│   ├── movies_router.py   # TMDB movie endpoints (generated)
│   ├── users_router.py    # User favorites/watchlist (generated)
│   └── dashboard_router.py # Aggregation endpoints (custom)
│
├── services/              # Backend service clients
│   ├── tmdb_client.py     # Enhanced TMDB API client
│   ├── movies_service.py  # Basic TMDB service (generated)
│   └── users_service.py   # User service client (generated)
│
├── models/                # Pydantic data models
│   └── movie_models.py    # Movie, EnrichedMovie, PersonalizedFeed, etc.
│
├── middleware/            # Custom middleware
│   ├── auth_middleware.py # JWT authentication
│   └── logging_middleware.py # Request logging
│
├── utils/                 # Utility modules
│   ├── cache_manager.py   # Redis caching
│   └── circuit_breaker.py # Circuit breaker pattern
│
├── tests/                 # Test suites
│   ├── conftest.py        # Pytest configuration
│   ├── test_tmdb_client.py
│   ├── test_dashboard.py
│   └── test_integration.py
│
├── mock-user-service/     # Mock backend service
│   ├── main.py            # FastAPI app (port 3001)
│   ├── requirements.txt
│   └── tests/
│
├── examples/              # Example requests
│   ├── requests.http      # VS Code REST Client
│   └── curl-examples.sh   # Shell script with curl commands
│
└── bffgen.config.py.json  # bffgen configuration
```

---

## ⚙️ Configuration

Configuration is managed through `config.py` using Pydantic Settings. Values can be overridden via environment variables or `.env` file.

### Key Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `PORT` | `8000` | Server port |
| `TMDB_API_KEY` | (provided) | TMDB API key |
| `TMDB_READ_TOKEN` | (provided) | TMDB Bearer token |
| `CACHE_ENABLED` | `false` | Enable Redis caching |
| `CIRCUIT_BREAKER_ENABLED` | `true` | Enable circuit breaker |
| `RATE_LIMIT_ENABLED` | `true` | Enable rate limiting |

### Environment Variables

Create a `.env` file (already included):

```bash
PORT=8000
DEBUG=true
TMDB_API_KEY=567c6b5ff1e4d84bb8f9adf76b084c4b
CACHE_ENABLED=false
```

---

## 🧪 Testing

### Run All Tests

```bash
# Activate venv
source venv/bin/activate

# Run tests with coverage
pytest

# With detailed output
pytest -v

# With coverage report
pytest --cov=. --cov-report=html
```

### Test Suites

- **test_tmdb_client.py** - TMDB API client tests (13 tests)
- **test_dashboard.py** - Dashboard aggregation tests (10 tests)
- **test_integration.py** - Integration tests (10 tests)
- **mock-user-service/tests/** - User service tests (15 tests)

**Total: 48 comprehensive tests**

---

## 🛠️ Development

### Code Quality

```bash
# Format code
black .

# Lint
ruff check .

# Type checking
mypy .
```

### Adding New Endpoints

1. **Update `bffgen.config.py.json`** - Add backend endpoint definition
2. **Run `bffgen generate`** - Regenerate routers/services (if using bffgen CLI)
3. **Create custom logic** - Add aggregation in `routers/dashboard_router.py`
4. **Write tests** - Add test cases
5. **Update docs** - Document the new endpoint

---

## 🎯 BFF Pattern Benefits

This project demonstrates several BFF advantages:

### 1. **Data Aggregation**
```python
# Single frontend request
GET /api/dashboard/feed

# BFF makes parallel backend calls
→ TMDB API (popular movies)
→ User Service (favorites)
→ User Service (watchlist)

# Returns combined, enriched data
```

### 2. **Data Transformation**
```python
# TMDB returns: {"id": 550, "title": "Fight Club", "vote_average": 8.4}
# User Service returns: {"movie_id": 550, "rating": 5}

# BFF combines into:
{
  "id": 550,
  "title": "Fight Club",
  "vote_average": 8.4,
  "is_favorite": true,
  "user_rating": 5
}
```

### 3. **Resilience**
- Circuit breaker prevents cascading failures
- Graceful degradation if user service is down
- Caching reduces load on TMDB API

---

## 🚧 Limitations & Future Enhancements

### Current Limitations

- **Single User**: Mock service doesn't support multiple users
- **No Persistence**: User data stored in memory only
- **No Authentication**: Authentication middleware is disabled by default
- **No Real Recommendations**: Uses TMDB similar/popular instead of ML

### Potential Enhancements

- [ ] Add Redis for persistent caching
- [ ] Implement real user authentication (OAuth2)
- [ ] Add PostgreSQL for user data persistence
- [ ] Build recommendation engine based on user preferences
- [ ] Add WebSocket support for real-time updates
- [ ] Implement GraphQL endpoint
- [ ] Add Docker Compose for easy deployment
- [ ] CI/CD pipeline with GitHub Actions

---

## 📖 Generated with bffgen

This project was scaffolded using [bffgen](https://github.com/RichGod93/bffgen), a CLI tool for creating Backend-for-Frontend services.

### Commands Used

```bash
# Initialize project
bffgen init movie-dashboard-bff --lang python-fastapi --async=true

# Generate routers and services
bffgen generate

# Add predefined templates (optional)
bffgen add-template auth
```

### bffgen Features Used

- ✅ FastAPI project scaffolding
- ✅ Async endpoint generation
- ✅ Router/service code generation
- ✅ Middleware setup (logging, auth)
- ✅ Test infrastructure
- ✅ Configuration management

---

## 📄 License

This project is a demo application. Feel free to use it as a reference for your own BFF implementations.

---

## 🤝 Contributing

This is a demo project, but suggestions and improvements are welcome!

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

---

## 📞 Support

- **TMDB API**: https://www.themoviedb.org/documentation/api
- **FastAPI Docs**: https://fastapi.tiangolo.com
- **bffgen**: https://github.com/RichGod93/bffgen

---

## 🎓 Learning Resources

This project demonstrates:

- BFF architectural pattern
- FastAPI async programming
- API aggregation patterns
- Circuit breaker implementation
- Pydantic data validation
- OpenAPI documentation
- Pytest async testing

Perfect for learning modern Python API development! 🚀

