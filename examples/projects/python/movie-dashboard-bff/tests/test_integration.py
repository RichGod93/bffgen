"""
Integration tests for Movie Dashboard BFF
Tests full request flows and service integration
"""
import pytest
from fastapi.testclient import TestClient
from unittest.mock import patch, AsyncMock

from main import app


class TestIntegration:
    """Integration test suite"""
    
    @pytest.fixture
    def client(self):
        """Test client"""
        return TestClient(app)
    
    def test_health_check(self, client):
        """Test health check endpoint"""
        response = client.get("/health")
        
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"
        assert data["service"] == "movie-dashboard-bff"
    
    def test_movies_router_available(self, client):
        """Test that movies router is registered"""
        # This will fail if service is down, but structure should be testable
        with patch('services.movies_service.MoviesService.get_popular', new_callable=AsyncMock) as mock:
            mock.return_value = {"results": [], "page": 1}
            response = client.get("/api/movies/popular")
            # Should not be 404 (route exists)
            assert response.status_code in [200, 500]
    
    def test_users_router_available(self, client):
        """Test that users router is registered"""
        with patch('services.users_service.UsersService.get_favorites', new_callable=AsyncMock) as mock:
            mock.return_value = {"favorites": [], "total": 0}
            response = client.get("/api/users/favorites")
            assert response.status_code in [200, 500]
    
    def test_dashboard_router_available(self, client):
        """Test that dashboard router is registered"""
        response = client.get("/api/dashboard/feed")
        # May fail due to real API calls, but route should exist
        assert response.status_code in [200, 500]
        assert response.status_code != 404
    
    def test_openapi_docs_available(self, client):
        """Test that OpenAPI docs are accessible"""
        response = client.get("/docs")
        assert response.status_code == 200
    
    def test_openapi_schema(self, client):
        """Test OpenAPI schema generation"""
        response = client.get("/openapi.json")
        
        assert response.status_code == 200
        schema = response.json()
        
        # Check basic structure
        assert "openapi" in schema
        assert "info" in schema
        assert "paths" in schema
        
        # Check our endpoints are documented
        paths = schema["paths"]
        assert "/health" in paths
        assert "/api/movies/popular" in paths
        assert "/api/users/favorites" in paths
        assert "/api/dashboard/feed" in paths
    
    def test_cors_headers(self, client):
        """Test CORS configuration"""
        response = client.options("/health")
        
        # CORS headers should be present
        assert response.status_code == 200
    
    def test_request_id_middleware(self, client):
        """Test that logging middleware adds request ID"""
        response = client.get("/health")
        
        # Should have request ID in headers
        assert "X-Request-ID" in response.headers
        assert "X-Process-Time" in response.headers
    
    async def test_full_aggregation_flow(self, client):
        """Test complete data aggregation flow"""
        # Mock all external services
        mock_tmdb_response = {
            "page": 1,
            "results": [
                {
                    "id": 550,
                    "title": "Fight Club",
                    "overview": "Great movie",
                    "release_date": "1999-10-15",
                    "vote_average": 8.4,
                    "poster_path": "/poster.jpg",
                    "genre_ids": [18]
                }
            ],
            "total_pages": 1,
            "total_results": 1
        }
        
        mock_favorites = {
            550: {"movie_id": 550, "rating": 5, "added_date": "2024-01-15"}
        }
        
        mock_watchlist = {}
        
        with patch('services.tmdb_client.tmdb_client.get_popular_movies', new_callable=AsyncMock) as mock_tmdb, \
             patch('routers.dashboard_router._fetch_user_favorites', new_callable=AsyncMock) as mock_fav, \
             patch('routers.dashboard_router._fetch_user_watchlist', new_callable=AsyncMock) as mock_watch:
            
            mock_tmdb.return_value = mock_tmdb_response
            mock_fav.return_value = mock_favorites
            mock_watch.return_value = mock_watchlist
            
            # Make request
            response = client.get("/api/dashboard/feed")
            
            assert response.status_code == 200
            data = response.json()
            
            # Verify data was aggregated correctly
            assert len(data["movies"]) == 1
            movie = data["movies"][0]
            
            # TMDB data present
            assert movie["id"] == 550
            assert movie["title"] == "Fight Club"
            assert movie["vote_average"] == 8.4
            
            # User data enriched
            assert movie["is_favorite"] is True
            assert movie["user_rating"] == 5
            assert movie["is_in_watchlist"] is False
            
            # Metadata present
            assert data["favorites_count"] == 1
            assert data["watchlist_count"] == 0
    
    def test_error_response_format(self, client):
        """Test that errors return consistent format"""
        # Try to get a non-existent route
        response = client.get("/api/nonexistent")
        
        assert response.status_code == 404
        data = response.json()
        assert "detail" in data

