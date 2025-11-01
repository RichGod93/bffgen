"""
Unit tests for TMDB client
Tests TMDB API integration with mocked responses
"""
import pytest
from unittest.mock import AsyncMock, patch, MagicMock
import httpx

from services.tmdb_client import TMDBClient, tmdb_client


class TestTMDBClient:
    """Test suite for TMDBClient"""
    
    @pytest.fixture
    def client(self):
        """Create fresh TMDB client for each test"""
        return TMDBClient()
    
    @pytest.fixture
    def mock_movie_response(self):
        """Mock movie data"""
        return {
            "id": 550,
            "title": "Fight Club",
            "overview": "A ticking-time-bomb insomniac...",
            "release_date": "1999-10-15",
            "vote_average": 8.4,
            "poster_path": "/pB8BM7pdSp6B6Ih7QZ4DrQ3PmJK.jpg"
        }
    
    @pytest.fixture
    def mock_popular_response(self):
        """Mock popular movies response"""
        return {
            "page": 1,
            "results": [
                {"id": 550, "title": "Fight Club"},
                {"id": 155, "title": "The Dark Knight"}
            ],
            "total_pages": 100,
            "total_results": 2000
        }
    
    async def test_get_popular_movies(self, client, mock_popular_response):
        """Test fetching popular movies"""
        with patch.object(client, '_make_request', new_callable=AsyncMock) as mock_request:
            mock_request.return_value = mock_popular_response
            
            result = await client.get_popular_movies(page=1)
            
            assert result == mock_popular_response
            mock_request.assert_called_once_with(
                "GET",
                "/movie/popular",
                {"page": 1},
                "popular:movies:page1:regionNone"
            )
    
    async def test_get_movie_details(self, client, mock_movie_response):
        """Test fetching movie details"""
        with patch.object(client, '_make_request', new_callable=AsyncMock) as mock_request:
            mock_request.return_value = mock_movie_response
            
            result = await client.get_movie_details(550)
            
            assert result["id"] == 550
            assert result["title"] == "Fight Club"
            assert "credits" in mock_request.call_args[0][1] or \
                   "append_to_response" in str(mock_request.call_args)
    
    async def test_search_movies(self, client):
        """Test movie search"""
        mock_response = {
            "page": 1,
            "results": [{"id": 550, "title": "Fight Club"}],
            "total_pages": 1,
            "total_results": 1
        }
        
        with patch.object(client, '_make_request', new_callable=AsyncMock) as mock_request:
            mock_request.return_value = mock_response
            
            result = await client.search_movies("fight club")
            
            assert len(result["results"]) == 1
            assert result["results"][0]["title"] == "Fight Club"
    
    async def test_get_trending_movies(self, client):
        """Test fetching trending movies"""
        mock_response = {
            "page": 1,
            "results": [{"id": 155, "title": "The Dark Knight"}]
        }
        
        with patch.object(client, '_make_request', new_callable=AsyncMock) as mock_request:
            mock_request.return_value = mock_response
            
            result = await client.get_trending_movies(time_window="day")
            
            assert len(result["results"]) == 1
            mock_request.assert_called_once_with(
                "GET",
                "/trending/movie/day",
                cache_key="trending:movies:day"
            )
    
    async def test_get_genres(self, client):
        """Test fetching genres list"""
        mock_response = {
            "genres": [
                {"id": 28, "name": "Action"},
                {"id": 18, "name": "Drama"}
            ]
        }
        
        with patch.object(client, '_make_request', new_callable=AsyncMock) as mock_request:
            mock_request.return_value = mock_response
            
            result = await client.get_genres()
            
            assert len(result["genres"]) == 2
            assert result["genres"][0]["name"] == "Action"
    
    async def test_authentication_header(self, client):
        """Test that Bearer token is included in requests"""
        mock_response = httpx.Response(200, json={"id": 550})
        
        with patch('httpx.AsyncClient.get', new_callable=AsyncMock) as mock_get:
            mock_get.return_value = mock_response
            
            await client._get_client()
            
            # Client should be created with Bearer token
            assert client.client is not None
            assert "Authorization" in client.client.headers
            assert client.client.headers["Authorization"].startswith("Bearer ")
    
    async def test_circuit_breaker_integration(self, client):
        """Test circuit breaker is used for requests"""
        with patch.object(client.circuit_breaker, 'call', new_callable=AsyncMock) as mock_circuit:
            mock_circuit.return_value = {"id": 550}
            
            with patch.object(client, '_get_client', new_callable=AsyncMock):
                await client._make_request("GET", "/test", {})
            
            # Circuit breaker should be called
            mock_circuit.assert_called_once()
    
    async def test_error_handling(self, client):
        """Test error handling for failed requests"""
        with patch.object(client, '_get_client', new_callable=AsyncMock) as mock_get_client:
            mock_client = AsyncMock()
            mock_response = httpx.Response(404, json={"error": "Not found"})
            mock_response.raise_for_status = MagicMock(side_effect=httpx.HTTPStatusError(
                "404", request=httpx.Request("GET", "http://test"), response=mock_response
            ))
            mock_client.request = AsyncMock(return_value=mock_response)
            mock_get_client.return_value = mock_client
            
            with pytest.raises(httpx.HTTPStatusError):
                await client._make_request("GET", "/movie/999999")
    
    async def test_discover_movies_with_filters(self, client):
        """Test discover endpoint with multiple filters"""
        mock_response = {
            "page": 1,
            "results": [{"id": 550, "title": "Fight Club"}]
        }
        
        with patch.object(client, '_make_request', new_callable=AsyncMock) as mock_request:
            mock_request.return_value = mock_response
            
            result = await client.discover_movies(
                page=1,
                sort_by="vote_average.desc",
                with_genres=[18, 28],
                year=1999,
                vote_average_gte=8.0
            )
            
            # Verify filters are passed correctly
            call_args = mock_request.call_args
            params = call_args[0][1]
            
            assert params["sort_by"] == "vote_average.desc"
            assert "with_genres" in params
            assert params["year"] == 1999
            assert params["vote_average.gte"] == 8.0
    
    async def test_close_client(self, client):
        """Test client cleanup"""
        # Create a client
        await client._get_client()
        assert client.client is not None
        
        # Close it
        with patch.object(client.client, 'aclose', new_callable=AsyncMock) as mock_close:
            await client.close()
            mock_close.assert_called_once()

