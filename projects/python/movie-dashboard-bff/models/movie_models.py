"""
Pydantic models for movie data structures
Provides type safety and validation for API responses
"""
from pydantic import BaseModel, Field
from typing import Optional, List, Dict, Any
from datetime import datetime


# === Basic TMDB Models ===

class Genre(BaseModel):
    """Movie genre"""
    id: int
    name: str


class ProductionCompany(BaseModel):
    """Production company"""
    id: int
    name: str
    logo_path: Optional[str] = None
    origin_country: Optional[str] = None


class CastMember(BaseModel):
    """Cast member"""
    id: int
    name: str
    character: str
    profile_path: Optional[str] = None
    order: Optional[int] = None


class CrewMember(BaseModel):
    """Crew member"""
    id: int
    name: str
    job: str
    department: str
    profile_path: Optional[str] = None


class Movie(BaseModel):
    """
    Basic movie information from TMDB
    Matches TMDB API response structure
    """
    id: int
    title: str
    original_title: Optional[str] = None
    overview: Optional[str] = None
    release_date: Optional[str] = None
    poster_path: Optional[str] = None
    backdrop_path: Optional[str] = None
    vote_average: Optional[float] = None
    vote_count: Optional[int] = None
    popularity: Optional[float] = None
    adult: Optional[bool] = False
    video: Optional[bool] = False
    genre_ids: Optional[List[int]] = []
    original_language: Optional[str] = None


class MovieDetails(BaseModel):
    """
    Detailed movie information
    Extended version with runtime, budget, revenue, etc.
    """
    id: int
    title: str
    original_title: Optional[str] = None
    tagline: Optional[str] = None
    overview: Optional[str] = None
    release_date: Optional[str] = None
    runtime: Optional[int] = None
    
    # Media
    poster_path: Optional[str] = None
    backdrop_path: Optional[str] = None
    
    # Ratings
    vote_average: Optional[float] = None
    vote_count: Optional[int] = None
    popularity: Optional[float] = None
    
    # Financial
    budget: Optional[int] = None
    revenue: Optional[int] = None
    
    # Classification
    adult: Optional[bool] = False
    video: Optional[bool] = False
    status: Optional[str] = None
    
    # Related data
    genres: Optional[List[Genre]] = []
    production_companies: Optional[List[ProductionCompany]] = []
    production_countries: Optional[List[Dict[str, Any]]] = []
    spoken_languages: Optional[List[Dict[str, Any]]] = []
    
    # URLs
    homepage: Optional[str] = None
    imdb_id: Optional[str] = None


# === User Data Models ===

class UserMovieData(BaseModel):
    """User's data for a specific movie"""
    is_favorite: bool = False
    is_in_watchlist: bool = False
    user_rating: Optional[int] = Field(None, ge=1, le=5, description="User rating 1-5")
    watched_date: Optional[str] = None
    added_to_favorites_date: Optional[str] = None
    added_to_watchlist_date: Optional[str] = None


class FavoriteMovie(BaseModel):
    """Movie in user's favorites"""
    movie_id: int
    title: str
    poster_path: Optional[str] = None
    rating: Optional[int] = Field(None, ge=1, le=5)
    added_date: str


class WatchlistMovie(BaseModel):
    """Movie in user's watchlist"""
    movie_id: int
    title: str
    poster_path: Optional[str] = None
    release_date: Optional[str] = None
    added_date: str


# === Enriched Models (TMDB + User Data) ===

class EnrichedMovie(Movie):
    """
    Movie with user-specific data
    Combines TMDB movie data with user preferences
    """
    # User data
    is_favorite: bool = False
    is_in_watchlist: bool = False
    user_rating: Optional[int] = None
    
    # Additional context
    in_theaters: Optional[bool] = None
    streaming_available: Optional[bool] = None


class EnrichedMovieDetails(MovieDetails):
    """
    Detailed movie with user data and additional context
    """
    # User data
    user_data: Optional[UserMovieData] = None
    
    # Credits (if included)
    cast: Optional[List[CastMember]] = []
    crew: Optional[List[CrewMember]] = []
    
    # Reviews count
    reviews_count: Optional[int] = None
    
    # Similar/Recommended
    similar_movies: Optional[List[Movie]] = []
    recommended_movies: Optional[List[Movie]] = []


# === Feed/List Response Models ===

class PersonalizedFeed(BaseModel):
    """
    Personalized movie feed response
    Combines popular movies with user preferences
    """
    movies: List[EnrichedMovie]
    total_pages: int = Field(..., description="Total pages available")
    current_page: int = Field(..., description="Current page number")
    total_results: int = Field(..., description="Total number of results")
    favorites_count: int = Field(0, description="Number of user's favorites")
    watchlist_count: int = Field(0, description="Number of items in watchlist")


class MoviesListResponse(BaseModel):
    """Generic movies list response"""
    movies: List[Movie]
    total_pages: int
    current_page: int
    total_results: int


class SearchResults(BaseModel):
    """Search results response"""
    results: List[Movie]
    total_pages: int
    page: int
    total_results: int
    query: str


# === Dashboard Aggregation Models ===

class DashboardStats(BaseModel):
    """User dashboard statistics"""
    total_favorites: int = 0
    total_watchlist: int = 0
    movies_watched: int = 0
    avg_rating: Optional[float] = None
    favorite_genres: List[str] = []


class DashboardFeed(BaseModel):
    """
    Complete dashboard feed
    Includes personalized movies and user stats
    """
    popular_movies: List[EnrichedMovie]
    trending_movies: List[EnrichedMovie]
    recommended_for_you: List[EnrichedMovie]
    continue_watching: List[EnrichedMovie]
    stats: DashboardStats
    page: int = 1
    total_pages: int = 1


# === Error Models ===

class ErrorResponse(BaseModel):
    """Standard error response"""
    detail: str
    error_code: Optional[str] = None
    timestamp: str = Field(default_factory=lambda: datetime.utcnow().isoformat())


# === Request Models ===

class AddToFavoritesRequest(BaseModel):
    """Request to add movie to favorites"""
    movie_id: int
    title: str
    poster_path: Optional[str] = None
    rating: Optional[int] = Field(None, ge=1, le=5)


class AddToWatchlistRequest(BaseModel):
    """Request to add movie to watchlist"""
    movie_id: int
    title: str
    poster_path: Optional[str] = None
    release_date: Optional[str] = None


class RateMovieRequest(BaseModel):
    """Request to rate a movie"""
    movie_id: int
    rating: int = Field(..., ge=1, le=5, description="Rating from 1 to 5")

