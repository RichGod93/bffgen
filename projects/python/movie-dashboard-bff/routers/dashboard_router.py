"""
Dashboard Router - Data Aggregation & Enrichment
Demonstrates BFF pattern by combining TMDB API with user preferences
"""
from fastapi import APIRouter, HTTPException, Depends, Query
from typing import List, Dict, Any, Optional
import asyncio
import logging

from models.movie_models import (
    EnrichedMovie,
    EnrichedMovieDetails,
    PersonalizedFeed,
    DashboardFeed,
    DashboardStats,
    UserMovieData,
    Movie
)
from services.tmdb_client import tmdb_client
from services.users_service import UsersService
from dependencies import get_current_user

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/api/dashboard", tags=["Dashboard"])

# Initialize services
users_service = UsersService()


async def _fetch_user_favorites() -> Dict[int, Dict[str, Any]]:
    """
    Fetch user's favorites from user service
    Returns dict mapping movie_id -> favorite data
    """
    try:
        response = await users_service.get_favorites()
        favorites = response.get("favorites", [])
        return {fav["movie_id"]: fav for fav in favorites}
    except Exception as e:
        logger.error(f"Failed to fetch favorites: {e}")
        return {}


async def _fetch_user_watchlist() -> Dict[int, Dict[str, Any]]:
    """
    Fetch user's watchlist from user service
    Returns dict mapping movie_id -> watchlist data
    """
    try:
        response = await users_service.get_watchlist()
        watchlist = response.get("watchlist", [])
        return {item["movie_id"]: item for item in watchlist}
    except Exception as e:
        logger.error(f"Failed to fetch watchlist: {e}")
        return {}


def _enrich_movie(
    movie: Dict[str, Any],
    favorites: Dict[int, Dict[str, Any]],
    watchlist: Dict[int, Dict[str, Any]]
) -> EnrichedMovie:
    """
    Enrich a movie with user data
    
    Args:
        movie: Raw movie data from TMDB
        favorites: User's favorites mapping
        watchlist: User's watchlist mapping
        
    Returns:
        EnrichedMovie with user context
    """
    movie_id = movie.get("id")
    
    return EnrichedMovie(
        **movie,
        is_favorite=movie_id in favorites,
        is_in_watchlist=movie_id in watchlist,
        user_rating=favorites.get(movie_id, {}).get("rating")
    )


@router.get("/feed", response_model=PersonalizedFeed)
async def get_personalized_feed(
    page: int = Query(1, ge=1, le=1000, description="Page number"),
    user_id: str = Depends(get_current_user)
):
    """
    Get personalized movie feed
    
    Aggregates:
    - Popular movies from TMDB
    - User's favorites from user service
    - Enriches movies with user context
    
    This demonstrates the BFF pattern: combining multiple backend services
    and transforming data to match frontend needs
    """
    logger.info(f"Fetching personalized feed for user {user_id}, page {page}")
    
    try:
        # Parallel requests to TMDB and user service
        popular_movies_task = tmdb_client.get_popular_movies(page=page)
        favorites_task = _fetch_user_favorites()
        watchlist_task = _fetch_user_watchlist()
        
        # Wait for all requests to complete
        popular_response, favorites, watchlist = await asyncio.gather(
            popular_movies_task,
            favorites_task,
            watchlist_task,
            return_exceptions=True
        )
        
        # Handle errors gracefully
        if isinstance(popular_response, Exception):
            logger.error(f"Failed to fetch popular movies: {popular_response}")
            raise HTTPException(status_code=500, detail="Failed to fetch movies from TMDB")
        
        if isinstance(favorites, Exception):
            logger.warning(f"Failed to fetch favorites, continuing without: {favorites}")
            favorites = {}
        
        if isinstance(watchlist, Exception):
            logger.warning(f"Failed to fetch watchlist, continuing without: {watchlist}")
            watchlist = {}
        
        # Enrich movies with user data
        movies_data = popular_response.get("results", [])
        enriched_movies = [
            _enrich_movie(movie, favorites, watchlist)
            for movie in movies_data
        ]
        
        # Build response
        return PersonalizedFeed(
            movies=enriched_movies,
            total_pages=popular_response.get("total_pages", 1),
            current_page=page,
            total_results=popular_response.get("total_results", 0),
            favorites_count=len(favorites),
            watchlist_count=len(watchlist)
        )
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error building personalized feed: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/movie/{movie_id}/enriched", response_model=EnrichedMovieDetails)
async def get_enriched_movie_details(
    movie_id: int,
    user_id: str = Depends(get_current_user)
):
    """
    Get movie details enriched with user data
    
    Aggregates:
    - Movie details from TMDB (with credits, reviews, similar)
    - User's data for this movie (favorite, watchlist, rating)
    
    Returns comprehensive view combining multiple data sources
    """
    logger.info(f"Fetching enriched details for movie {movie_id}, user {user_id}")
    
    try:
        # Parallel requests
        movie_task = tmdb_client.get_movie_details(
            movie_id,
            append_to_response=["credits", "reviews", "similar", "recommendations"]
        )
        favorites_task = _fetch_user_favorites()
        watchlist_task = _fetch_user_watchlist()
        
        movie_details, favorites, watchlist = await asyncio.gather(
            movie_task,
            favorites_task,
            watchlist_task,
            return_exceptions=True
        )
        
        # Handle errors
        if isinstance(movie_details, Exception):
            logger.error(f"Failed to fetch movie details: {movie_details}")
            raise HTTPException(status_code=500, detail="Failed to fetch movie details from TMDB")
        
        if isinstance(favorites, Exception):
            favorites = {}
        
        if isinstance(watchlist, Exception):
            watchlist = {}
        
        # Build user data
        user_data = UserMovieData(
            is_favorite=movie_id in favorites,
            is_in_watchlist=movie_id in watchlist,
            user_rating=favorites.get(movie_id, {}).get("rating"),
            added_to_favorites_date=favorites.get(movie_id, {}).get("added_date"),
            added_to_watchlist_date=watchlist.get(movie_id, {}).get("added_date")
        )
        
        # Extract credits
        credits = movie_details.get("credits", {})
        cast = credits.get("cast", [])[:10]  # Top 10 cast members
        crew = credits.get("crew", [])[:10]  # Top 10 crew members
        
        # Extract similar movies
        similar = movie_details.get("similar", {}).get("results", [])[:6]
        
        # Build enriched response
        return EnrichedMovieDetails(
            **movie_details,
            user_data=user_data,
            cast=cast,
            crew=crew,
            reviews_count=movie_details.get("reviews", {}).get("total_results", 0),
            similar_movies=similar
        )
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error fetching enriched movie details: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/complete", response_model=DashboardFeed)
async def get_complete_dashboard(
    user_id: str = Depends(get_current_user)
):
    """
    Get complete dashboard with multiple movie lists
    
    Aggregates:
    - Popular movies
    - Trending movies (today)
    - User's favorites
    - User's watchlist
    - User statistics
    
    This endpoint demonstrates advanced BFF aggregation:
    multiple parallel requests to different backends
    """
    logger.info(f"Fetching complete dashboard for user {user_id}")
    
    try:
        # Parallel requests to multiple endpoints
        popular_task = tmdb_client.get_popular_movies(page=1)
        trending_task = tmdb_client.get_trending_movies(time_window="day")
        favorites_task = _fetch_user_favorites()
        watchlist_task = _fetch_user_watchlist()
        
        popular_response, trending_response, favorites, watchlist = await asyncio.gather(
            popular_task,
            trending_task,
            favorites_task,
            watchlist_task,
            return_exceptions=True
        )
        
        # Handle errors with fallbacks
        popular_movies = []
        if not isinstance(popular_response, Exception):
            popular_movies = popular_response.get("results", [])[:12]
        
        trending_movies = []
        if not isinstance(trending_response, Exception):
            trending_movies = trending_response.get("results", [])[:12]
        
        if isinstance(favorites, Exception):
            favorites = {}
        
        if isinstance(watchlist, Exception):
            watchlist = {}
        
        # Enrich all movies
        enriched_popular = [
            _enrich_movie(movie, favorites, watchlist)
            for movie in popular_movies
        ]
        
        enriched_trending = [
            _enrich_movie(movie, favorites, watchlist)
            for movie in trending_movies
        ]
        
        # Calculate user stats
        favorite_ratings = [
            fav.get("rating") 
            for fav in favorites.values() 
            if fav.get("rating") is not None
        ]
        
        avg_rating = sum(favorite_ratings) / len(favorite_ratings) if favorite_ratings else None
        
        stats = DashboardStats(
            total_favorites=len(favorites),
            total_watchlist=len(watchlist),
            movies_watched=len(favorites),  # Assuming favorites are watched
            avg_rating=avg_rating,
            favorite_genres=[]  # Could be computed from favorites
        )
        
        return DashboardFeed(
            popular_movies=enriched_popular,
            trending_movies=enriched_trending,
            recommended_for_you=[],  # Could implement recommendation algorithm
            continue_watching=[],     # Would need viewing history
            stats=stats,
            page=1,
            total_pages=1
        )
        
    except Exception as e:
        logger.error(f"Error building complete dashboard: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/search/enriched")
async def search_movies_enriched(
    query: str = Query(..., min_length=1, description="Search query"),
    page: int = Query(1, ge=1, le=1000),
    user_id: str = Depends(get_current_user)
):
    """
    Search movies with user context
    
    Searches TMDB and enriches results with user data
    """
    logger.info(f"Searching movies: '{query}' for user {user_id}")
    
    try:
        # Parallel requests
        search_task = tmdb_client.search_movies(query=query, page=page)
        favorites_task = _fetch_user_favorites()
        watchlist_task = _fetch_user_watchlist()
        
        search_response, favorites, watchlist = await asyncio.gather(
            search_task,
            favorites_task,
            watchlist_task,
            return_exceptions=True
        )
        
        if isinstance(search_response, Exception):
            raise HTTPException(status_code=500, detail="Search failed")
        
        if isinstance(favorites, Exception):
            favorites = {}
        
        if isinstance(watchlist, Exception):
            watchlist = {}
        
        # Enrich search results
        movies = search_response.get("results", [])
        enriched_movies = [
            _enrich_movie(movie, favorites, watchlist)
            for movie in movies
        ]
        
        return {
            "results": enriched_movies,
            "query": query,
            "page": page,
            "total_pages": search_response.get("total_pages", 1),
            "total_results": search_response.get("total_results", 0)
        }
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error in enriched search: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=str(e))

