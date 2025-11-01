# Mock User Service

A lightweight FastAPI service that provides in-memory storage for user favorites and watchlist functionality.

## Purpose

This service simulates a user data backend for the Movie Dashboard BFF, allowing you to:
- Store and retrieve user's favorite movies
- Manage a user's watchlist
- Test the BFF's data aggregation features without a full backend

## Quick Start

### 1. Install Dependencies

```bash
pip install -r requirements.txt
```

Or from the project root:

```bash
cd mock-user-service
pip install -r requirements.txt
```

### 2. Run the Service

```bash
python main.py
```

Or with uvicorn directly:

```bash
uvicorn main:app --port 3001 --reload
```

### 3. Access the API

- **API Docs**: http://localhost:3001/docs
- **Health Check**: http://localhost:3001/health

## API Endpoints

### Favorites

- `GET /favorites` - Get all favorites
- `POST /favorites` - Add to favorites
- `DELETE /favorites/{movie_id}` - Remove from favorites
- `GET /favorites/{movie_id}` - Get specific favorite

### Watchlist

- `GET /watchlist` - Get watchlist
- `POST /watchlist` - Add to watchlist
- `DELETE /watchlist/{movie_id}` - Remove from watchlist
- `GET /watchlist/{movie_id}` - Get specific watchlist item

### Utility

- `POST /reset` - Reset to sample data
- `DELETE /clear` - Clear all data

## Sample Data

The service includes sample favorites and watchlist items for testing:

**Favorites:**
- Fight Club (ID: 550)
- The Dark Knight (ID: 155)
- Forrest Gump (ID: 13)

**Watchlist:**
- Pulp Fiction (ID: 680)
- Inception (ID: 27205)

## Example Requests

### Add to Favorites

```bash
curl -X POST http://localhost:3001/favorites \
  -H "Content-Type: application/json" \
  -d '{
    "movie_id": 278,
    "title": "The Shawshank Redemption",
    "rating": 5
  }'
```

### Get Favorites

```bash
curl http://localhost:3001/favorites
```

### Remove from Favorites

```bash
curl -X DELETE http://localhost:3001/favorites/278
```

## Integration with BFF

The main Movie Dashboard BFF is configured to call this service at `http://localhost:3001`. Make sure this service is running before starting the BFF.

## Notes

- **In-Memory Storage**: All data is stored in memory and will be lost when the service restarts
- **No Authentication**: This is a mock service - no auth required
- **Single User**: Simulates a single user's data (no multi-user support)
- **Development Only**: Not intended for production use

## Running Both Services

Use the `start-all.sh` script from the project root to start both the mock service and the main BFF together:

```bash
cd ..
./start-all.sh
```

