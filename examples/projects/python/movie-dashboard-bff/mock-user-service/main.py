"""
Mock User Service
Provides in-memory storage for user favorites and watchlist
For development and testing purposes
"""
from fastapi import FastAPI, HTTPException, Body
from typing import Dict, Any, List
import logging
from datetime import datetime

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(
    title="Mock User Service",
    description="In-memory user favorites and watchlist service",
    version="1.0.0"
)

# In-memory storage
favorites_store: Dict[int, Dict[str, Any]] = {}
watchlist_store: Dict[int, Dict[str, Any]] = {}

# Sample data for testing
SAMPLE_FAVORITES = [
    {"movie_id": 550, "title": "Fight Club", "rating": 5, "added_date": "2024-01-15"},
    {"movie_id": 155, "title": "The Dark Knight", "rating": 5, "added_date": "2024-01-10"},
    {"movie_id": 13, "title": "Forrest Gump", "rating": 4, "added_date": "2024-01-05"},
]

SAMPLE_WATCHLIST = [
    {"movie_id": 680, "title": "Pulp Fiction", "added_date": "2024-02-01"},
    {"movie_id": 27205, "title": "Inception", "added_date": "2024-01-28"},
]

# Initialize with sample data
for fav in SAMPLE_FAVORITES:
    favorites_store[fav["movie_id"]] = fav

for item in SAMPLE_WATCHLIST:
    watchlist_store[item["movie_id"]] = item


@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "service": "mock-user-service",
        "favorites_count": len(favorites_store),
        "watchlist_count": len(watchlist_store)
    }


# === Favorites Endpoints ===

@app.get("/favorites")
async def get_favorites():
    """
    Get all favorite movies
    
    Returns:
        List of favorite movies with ratings
    """
    logger.info(f"Fetching favorites (count: {len(favorites_store)})")
    return {
        "favorites": list(favorites_store.values()),
        "total": len(favorites_store)
    }


@app.post("/favorites")
async def add_favorite(payload: Dict[str, Any] = Body(...)):
    """
    Add a movie to favorites
    
    Request body:
        {
            "movie_id": int,
            "title": str,
            "rating": int (1-5, optional),
            "poster_path": str (optional)
        }
    
    Returns:
        Success message and updated favorite
    """
    movie_id = payload.get("movie_id")
    
    if not movie_id:
        raise HTTPException(status_code=400, detail="movie_id is required")
    
    if movie_id in favorites_store:
        raise HTTPException(status_code=409, detail="Movie already in favorites")
    
    favorite = {
        "movie_id": movie_id,
        "title": payload.get("title", f"Movie {movie_id}"),
        "rating": payload.get("rating"),
        "poster_path": payload.get("poster_path"),
        "added_date": datetime.now().isoformat()
    }
    
    favorites_store[movie_id] = favorite
    logger.info(f"Added movie {movie_id} to favorites")
    
    return {
        "message": "Movie added to favorites",
        "favorite": favorite
    }


@app.delete("/favorites/{movie_id}")
async def delete_favorite(movie_id: int):
    """
    Remove a movie from favorites
    
    Args:
        movie_id: ID of the movie to remove
        
    Returns:
        Success message
    """
    if movie_id not in favorites_store:
        raise HTTPException(status_code=404, detail="Movie not found in favorites")
    
    removed = favorites_store.pop(movie_id)
    logger.info(f"Removed movie {movie_id} from favorites")
    
    return {
        "message": "Movie removed from favorites",
        "removed": removed
    }


@app.get("/favorites/{movie_id}")
async def get_favorite(movie_id: int):
    """
    Get a specific favorite
    
    Args:
        movie_id: ID of the movie
        
    Returns:
        Favorite details or 404
    """
    if movie_id not in favorites_store:
        raise HTTPException(status_code=404, detail="Movie not found in favorites")
    
    return favorites_store[movie_id]


# === Watchlist Endpoints ===

@app.get("/watchlist")
async def get_watchlist():
    """
    Get all movies in watchlist
    
    Returns:
        List of movies in watchlist
    """
    logger.info(f"Fetching watchlist (count: {len(watchlist_store)})")
    return {
        "watchlist": list(watchlist_store.values()),
        "total": len(watchlist_store)
    }


@app.post("/watchlist")
async def add_to_watchlist(payload: Dict[str, Any] = Body(...)):
    """
    Add a movie to watchlist
    
    Request body:
        {
            "movie_id": int,
            "title": str,
            "poster_path": str (optional),
            "release_date": str (optional)
        }
    
    Returns:
        Success message and watchlist item
    """
    movie_id = payload.get("movie_id")
    
    if not movie_id:
        raise HTTPException(status_code=400, detail="movie_id is required")
    
    if movie_id in watchlist_store:
        raise HTTPException(status_code=409, detail="Movie already in watchlist")
    
    watchlist_item = {
        "movie_id": movie_id,
        "title": payload.get("title", f"Movie {movie_id}"),
        "poster_path": payload.get("poster_path"),
        "release_date": payload.get("release_date"),
        "added_date": datetime.now().isoformat()
    }
    
    watchlist_store[movie_id] = watchlist_item
    logger.info(f"Added movie {movie_id} to watchlist")
    
    return {
        "message": "Movie added to watchlist",
        "watchlist_item": watchlist_item
    }


@app.delete("/watchlist/{movie_id}")
async def delete_from_watchlist(movie_id: int):
    """
    Remove a movie from watchlist
    
    Args:
        movie_id: ID of the movie to remove
        
    Returns:
        Success message
    """
    if movie_id not in watchlist_store:
        raise HTTPException(status_code=404, detail="Movie not found in watchlist")
    
    removed = watchlist_store.pop(movie_id)
    logger.info(f"Removed movie {movie_id} from watchlist")
    
    return {
        "message": "Movie removed from watchlist",
        "removed": removed
    }


@app.get("/watchlist/{movie_id}")
async def get_watchlist_item(movie_id: int):
    """
    Get a specific watchlist item
    
    Args:
        movie_id: ID of the movie
        
    Returns:
        Watchlist item details or 404
    """
    if movie_id not in watchlist_store:
        raise HTTPException(status_code=404, detail="Movie not found in watchlist")
    
    return watchlist_store[movie_id]


# === Utility Endpoints ===

@app.post("/reset")
async def reset_data():
    """Reset to sample data (for testing)"""
    favorites_store.clear()
    watchlist_store.clear()
    
    for fav in SAMPLE_FAVORITES:
        favorites_store[fav["movie_id"]] = fav
    
    for item in SAMPLE_WATCHLIST:
        watchlist_store[item["movie_id"]] = item
    
    logger.info("Data reset to sample data")
    return {
        "message": "Data reset successfully",
        "favorites_count": len(favorites_store),
        "watchlist_count": len(watchlist_store)
    }


@app.delete("/clear")
async def clear_all_data():
    """Clear all data (for testing)"""
    favorites_store.clear()
    watchlist_store.clear()
    
    logger.info("All data cleared")
    return {
        "message": "All data cleared",
        "favorites_count": 0,
        "watchlist_count": 0
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host="0.0.0.0", port=3001, reload=True)

