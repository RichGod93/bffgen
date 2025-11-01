"""
Enhanced TMDB API Client with comprehensive features
Provides additional functionality beyond basic service methods
"""
import httpx
import logging
from typing import Dict, Any, List, Optional
from config import settings
from utils.circuit_breaker import CircuitBreaker
from utils.cache_manager import cache

logger = logging.getLogger(__name__)


class TMDBClient:
    """
    Enhanced TMDB API client with:
    - Bearer token authentication
    - Circuit breaker pattern
    - Caching support
    - Comprehensive error handling
    - Extended endpoint coverage
    """
    
    def __init__(self):
        self.base_url = settings.TMDB_BASE_URL
        self.api_key = settings.TMDB_API_KEY
        self.read_token = settings.TMDB_READ_TOKEN
        self.timeout = 30.0
        self.client: Optional[httpx.AsyncClient] = None
        self.circuit_breaker = CircuitBreaker("tmdb-api")
    
    async def _get_client(self) -> httpx.AsyncClient:
        """Get or create HTTP client with Bearer authentication"""
        if self.client is None:
            headers = {
                "Authorization": f"Bearer {self.read_token}",
                "Content-Type": "application/json",
                "Accept": "application/json"
            }
            self.client = httpx.AsyncClient(
                timeout=self.timeout,
                headers=headers,
                follow_redirects=True
            )
        return self.client
    
    async def _make_request(
        self,
        method: str,
        endpoint: str,
        params: Optional[Dict] = None,
        cache_key: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Make HTTP request with circuit breaker and caching
        
        Args:
            method: HTTP method (GET, POST, etc.)
            endpoint: API endpoint path
            params: Query parameters
            cache_key: Optional cache key for GET requests
            
        Returns:
            API response as dictionary
        """
        # Check cache for GET requests
        if method == "GET" and cache_key:
            cached = await cache.get(cache_key)
            if cached:
                logger.info(f"Cache hit for {cache_key}")
                return cached
        
        # Make request with circuit breaker
        async def _request():
            client = await self._get_client()
            url = f"{self.base_url}{endpoint}"
            logger.info(f"{method} {url}")
            
            response = await client.request(method, url, params=params)
            response.raise_for_status()
            return response.json()
        
        result = await self.circuit_breaker.call(_request)
        
        # Cache successful GET requests
        if method == "GET" and cache_key:
            await cache.set(cache_key, result)
        
        return result
    
    # === Popular & Trending ===
    
    async def get_popular_movies(self, page: int = 1, region: Optional[str] = None) -> Dict[str, Any]:
        """
        Get popular movies
        
        Args:
            page: Page number (1-1000)
            region: ISO 3166-1 code to filter results
            
        Returns:
            Popular movies with pagination
        """
        params = {"page": page}
        if region:
            params["region"] = region
        
        cache_key = f"popular:movies:page{page}:region{region}"
        return await self._make_request("GET", "/movie/popular", params, cache_key)
    
    async def get_trending_movies(self, time_window: str = "day") -> Dict[str, Any]:
        """
        Get trending movies
        
        Args:
            time_window: 'day' or 'week'
            
        Returns:
            Trending movies list
        """
        cache_key = f"trending:movies:{time_window}"
        return await self._make_request("GET", f"/trending/movie/{time_window}", cache_key=cache_key)
    
    async def get_now_playing(self, page: int = 1, region: Optional[str] = None) -> Dict[str, Any]:
        """Get movies currently in theaters"""
        params = {"page": page}
        if region:
            params["region"] = region
        
        cache_key = f"now_playing:page{page}:region{region}"
        return await self._make_request("GET", "/movie/now_playing", params, cache_key)
    
    async def get_upcoming_movies(self, page: int = 1, region: Optional[str] = None) -> Dict[str, Any]:
        """Get upcoming movies"""
        params = {"page": page}
        if region:
            params["region"] = region
        
        cache_key = f"upcoming:page{page}:region{region}"
        return await self._make_request("GET", "/movie/upcoming", params, cache_key)
    
    async def get_top_rated(self, page: int = 1, region: Optional[str] = None) -> Dict[str, Any]:
        """Get top rated movies"""
        params = {"page": page}
        if region:
            params["region"] = region
        
        cache_key = f"top_rated:page{page}:region{region}"
        return await self._make_request("GET", "/movie/top_rated", params, cache_key)
    
    # === Movie Details ===
    
    async def get_movie_details(
        self,
        movie_id: int,
        append_to_response: Optional[List[str]] = None
    ) -> Dict[str, Any]:
        """
        Get detailed movie information
        
        Args:
            movie_id: TMDB movie ID
            append_to_response: Additional data to include (credits, reviews, similar, videos, etc.)
            
        Returns:
            Comprehensive movie details
        """
        params = {}
        if append_to_response:
            params["append_to_response"] = ",".join(append_to_response)
        else:
            # Default: include most useful data
            params["append_to_response"] = "credits,reviews,similar,videos,images"
        
        cache_key = f"movie:{movie_id}:full"
        return await self._make_request("GET", f"/movie/{movie_id}", params, cache_key)
    
    async def get_movie_credits(self, movie_id: int) -> Dict[str, Any]:
        """Get cast and crew for a movie"""
        cache_key = f"movie:{movie_id}:credits"
        return await self._make_request("GET", f"/movie/{movie_id}/credits", cache_key=cache_key)
    
    async def get_movie_reviews(self, movie_id: int, page: int = 1) -> Dict[str, Any]:
        """Get user reviews for a movie"""
        cache_key = f"movie:{movie_id}:reviews:page{page}"
        return await self._make_request("GET", f"/movie/{movie_id}/reviews", {"page": page}, cache_key)
    
    async def get_movie_recommendations(self, movie_id: int, page: int = 1) -> Dict[str, Any]:
        """Get recommended movies based on a movie"""
        cache_key = f"movie:{movie_id}:recommendations:page{page}"
        return await self._make_request("GET", f"/movie/{movie_id}/recommendations", {"page": page}, cache_key)
    
    async def get_similar_movies(self, movie_id: int, page: int = 1) -> Dict[str, Any]:
        """Get similar movies"""
        cache_key = f"movie:{movie_id}:similar:page{page}"
        return await self._make_request("GET", f"/movie/{movie_id}/similar", {"page": page}, cache_key)
    
    # === Search ===
    
    async def search_movies(
        self,
        query: str,
        page: int = 1,
        year: Optional[int] = None,
        include_adult: bool = False
    ) -> Dict[str, Any]:
        """
        Search for movies
        
        Args:
            query: Search query string
            page: Page number
            year: Filter by release year
            include_adult: Include adult content
            
        Returns:
            Search results with pagination
        """
        params = {
            "query": query,
            "page": page,
            "include_adult": include_adult
        }
        if year:
            params["year"] = year
        
        # Don't cache search results as they're user-specific
        return await self._make_request("GET", "/search/movie", params)
    
    async def search_multi(self, query: str, page: int = 1) -> Dict[str, Any]:
        """
        Multi-search (movies, TV, people)
        
        Args:
            query: Search query
            page: Page number
            
        Returns:
            Multi-search results
        """
        params = {"query": query, "page": page}
        return await self._make_request("GET", "/search/multi", params)
    
    # === Discover ===
    
    async def discover_movies(
        self,
        page: int = 1,
        sort_by: str = "popularity.desc",
        with_genres: Optional[List[int]] = None,
        year: Optional[int] = None,
        vote_average_gte: Optional[float] = None
    ) -> Dict[str, Any]:
        """
        Discover movies with filters
        
        Args:
            page: Page number
            sort_by: Sort criteria (popularity.desc, vote_average.desc, etc.)
            with_genres: List of genre IDs
            year: Filter by year
            vote_average_gte: Minimum vote average
            
        Returns:
            Filtered movie results
        """
        params = {"page": page, "sort_by": sort_by}
        
        if with_genres:
            params["with_genres"] = ",".join(map(str, with_genres))
        if year:
            params["year"] = year
        if vote_average_gte:
            params["vote_average.gte"] = vote_average_gte
        
        return await self._make_request("GET", "/discover/movie", params)
    
    # === Genres ===
    
    async def get_genres(self) -> Dict[str, Any]:
        """Get list of movie genres"""
        cache_key = "genres:movies"
        return await self._make_request("GET", "/genre/movie/list", cache_key=cache_key)
    
    # === Person ===
    
    async def get_person_details(self, person_id: int) -> Dict[str, Any]:
        """Get details about a person (actor, director, etc.)"""
        cache_key = f"person:{person_id}"
        return await self._make_request("GET", f"/person/{person_id}", cache_key=cache_key)
    
    async def get_person_movie_credits(self, person_id: int) -> Dict[str, Any]:
        """Get movie credits for a person"""
        cache_key = f"person:{person_id}:movie_credits"
        return await self._make_request("GET", f"/person/{person_id}/movie_credits", cache_key=cache_key)
    
    # === Utility ===
    
    async def close(self):
        """Close HTTP client and cleanup resources"""
        if self.client:
            await self.client.aclose()
            self.client = None
        logger.info("TMDB client closed")


# Global TMDB client instance
tmdb_client = TMDBClient()

