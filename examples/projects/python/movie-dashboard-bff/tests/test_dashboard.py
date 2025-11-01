"""
Unit tests for dashboard router
Tests data aggregation and enrichment logic
"""
import pytest
from unittest.mock import AsyncMock, patch
from fastapi.testclient import TestClient

from main import app


class TestDashboardRouter:
    """Test suite for Dashboard endpoints"""
    
    @pytest.fixture
    def client(self):
        """Test client"""
        return TestClient(app)
    
    @pytest.fixture
    def mock_tmdb_movies(self):
        """Mock TMDB movies response"""
        return {
            "page": 1,
            "results": [
                {
                    "id": 550,
                    "title": "Fight Club",
                    "overview": "A ticking-time-bomb insomniac...",
                    "release_date": "1999-10-15",
                    "vote_average": 8.4,
                    "poster_path": "/poster.jpg",
                    "genre_ids": [18]
                },
                {
                    "id": 155,
                    "title": "The Dark Knight",
                    "overview": "Batman raises the stakes...",
                    "release_date": "2008-07-18",
                    "vote_average": 9.0,
                    "poster_path": "/poster2.jpg",
                    "genre_ids": [28, 18]
                }
            ],
            "total_pages": 100,
            "total_results": 2000
        }
    
    @pytest.fixture
    def mock_favorites(self):
        """Mock user favorites"""
        return {
            "favorites": [
                {"movie_id": 550, "title": "Fight Club", "rating": 5, "added_date": "2024-01-15"}
            ],
            "total": 1
        }
    
    @pytest.fixture
    def mock_watchlist(self):
        """Mock user watchlist"""
        return {
            "watchlist": [
                {"movie_id": 155, "title": "The Dark Knight", "added_date": "2024-01-20"}
            ],
            "total": 1
        }
    
    async def test_get_personalized_feed(self, client, mock_tmdb_movies, mock_favorites, mock_watchlist):
        """Test personalized feed endpoint"""
        with patch('services.tmdb_client.tmdb_client.get_popular_movies', new_callable=AsyncMock) as mock_tmdb, \
             patch('routers.dashboard_router._fetch_user_favorites', new_callable=AsyncMock) as mock_fav, \
             patch('routers.dashboard_router._fetch_user_watchlist', new_callable=AsyncMock) as mock_watch:
            
            mock_tmdb.return_value = mock_tmdb_movies
            mock_fav.return_value = {550: mock_favorites["favorites"][0]}
            mock_watch.return_value = {155: mock_watchlist["watchlist"][0]}
            
            response = client.get("/api/dashboard/feed")
            
            assert response.status_code == 200
            data = response.json()
            
            # Check structure
            assert "movies" in data
            assert "total_pages" in data
            assert "favorites_count" in data
            assert "watchlist_count" in data
            
            # Check enrichment
            assert len(data["movies"]) == 2
            assert data["favorites_count"] == 1
            assert data["watchlist_count"] == 1
            
            # Verify movies are enriched with user data
            fight_club = next(m for m in data["movies"] if m["id"] == 550)
            assert fight_club["is_favorite"] is True
            assert fight_club["user_rating"] == 5
            
            dark_knight = next(m for m in data["movies"] if m["id"] == 155)
            assert dark_knight["is_in_watchlist"] is True
    
    async def test_get_personalized_feed_pagination(self, client, mock_tmdb_movies):
        """Test feed pagination"""
        with patch('services.tmdb_client.tmdb_client.get_popular_movies', new_callable=AsyncMock) as mock_tmdb, \
             patch('routers.dashboard_router._fetch_user_favorites', new_callable=AsyncMock) as mock_fav, \
             patch('routers.dashboard_router._fetch_user_watchlist', new_callable=AsyncMock) as mock_watch:
            
            mock_tmdb.return_value = mock_tmdb_movies
            mock_fav.return_value = {}
            mock_watch.return_value = {}
            
            response = client.get("/api/dashboard/feed?page=2")
            
            assert response.status_code == 200
            data = response.json()
            assert data["current_page"] == 2
            
            # Verify TMDB was called with correct page
            mock_tmdb.assert_called_once_with(page=2)
    
    async def test_get_enriched_movie_details(self, client):
        """Test enriched movie details endpoint"""
        mock_movie_details = {
            "id": 550,
            "title": "Fight Club",
            "overview": "A ticking-time-bomb insomniac...",
            "release_date": "1999-10-15",
            "runtime": 139,
            "vote_average": 8.4,
            "credits": {
                "cast": [
                    {"id": 287, "name": "Brad Pitt", "character": "Tyler Durden"}
                ],
                "crew": [
                    {"id": 7467, "name": "David Fincher", "job": "Director"}
                ]
            },
            "reviews": {"total_results": 10},
            "similar": {"results": []},
            "genres": [{"id": 18, "name": "Drama"}]
        }
        
        with patch('services.tmdb_client.tmdb_client.get_movie_details', new_callable=AsyncMock) as mock_tmdb, \
             patch('routers.dashboard_router._fetch_user_favorites', new_callable=AsyncMock) as mock_fav, \
             patch('routers.dashboard_router._fetch_user_watchlist', new_callable=AsyncMock) as mock_watch:
            
            mock_tmdb.return_value = mock_movie_details
            mock_fav.return_value = {
                550: {"movie_id": 550, "rating": 5, "added_date": "2024-01-15"}
            }
            mock_watch.return_value = {}
            
            response = client.get("/api/dashboard/movie/550/enriched")
            
            assert response.status_code == 200
            data = response.json()
            
            # Check basic info
            assert data["id"] == 550
            assert data["title"] == "Fight Club"
            
            # Check user data
            assert "user_data" in data
            assert data["user_data"]["is_favorite"] is True
            assert data["user_data"]["user_rating"] == 5
            
            # Check aggregated data
            assert "cast" in data
            assert "reviews_count" in data
            assert data["reviews_count"] == 10
    
    async def test_get_complete_dashboard(self, client, mock_tmdb_movies):
        """Test complete dashboard endpoint"""
        mock_trending = {
            "results": [
                {"id": 680, "title": "Pulp Fiction", "vote_average": 8.9}
            ]
        }
        
        with patch('services.tmdb_client.tmdb_client.get_popular_movies', new_callable=AsyncMock) as mock_popular, \
             patch('services.tmdb_client.tmdb_client.get_trending_movies', new_callable=AsyncMock) as mock_trend, \
             patch('routers.dashboard_router._fetch_user_favorites', new_callable=AsyncMock) as mock_fav, \
             patch('routers.dashboard_router._fetch_user_watchlist', new_callable=AsyncMock) as mock_watch:
            
            mock_popular.return_value = mock_tmdb_movies
            mock_trend.return_value = mock_trending
            mock_fav.return_value = {
                550: {"movie_id": 550, "rating": 5},
                155: {"movie_id": 155, "rating": 4}
            }
            mock_watch.return_value = {}
            
            response = client.get("/api/dashboard/complete")
            
            assert response.status_code == 200
            data = response.json()
            
            # Check all sections present
            assert "popular_movies" in data
            assert "trending_movies" in data
            assert "stats" in data
            
            # Check stats calculation
            stats = data["stats"]
            assert stats["total_favorites"] == 2
            assert stats["avg_rating"] == 4.5  # (5 + 4) / 2
    
    async def test_search_movies_enriched(self, client):
        """Test enriched search endpoint"""
        mock_search_results = {
            "page": 1,
            "results": [
                {"id": 550, "title": "Fight Club", "vote_average": 8.4}
            ],
            "total_pages": 1,
            "total_results": 1
        }
        
        with patch('services.tmdb_client.tmdb_client.search_movies', new_callable=AsyncMock) as mock_search, \
             patch('routers.dashboard_router._fetch_user_favorites', new_callable=AsyncMock) as mock_fav, \
             patch('routers.dashboard_router._fetch_user_watchlist', new_callable=AsyncMock) as mock_watch:
            
            mock_search.return_value = mock_search_results
            mock_fav.return_value = {550: {"movie_id": 550, "rating": 5}}
            mock_watch.return_value = {}
            
            response = client.get("/api/dashboard/search/enriched?query=fight+club")
            
            assert response.status_code == 200
            data = response.json()
            
            assert data["query"] == "fight club"
            assert len(data["results"]) == 1
            assert data["results"][0]["is_favorite"] is True
    
    async def test_feed_handles_tmdb_error(self, client):
        """Test graceful handling of TMDB API errors"""
        with patch('services.tmdb_client.tmdb_client.get_popular_movies', new_callable=AsyncMock) as mock_tmdb:
            mock_tmdb.side_effect = Exception("TMDB API error")
            
            response = client.get("/api/dashboard/feed")
            
            # Should return error
            assert response.status_code == 500
    
    async def test_feed_handles_user_service_error(self, client, mock_tmdb_movies):
        """Test graceful handling when user service fails"""
        with patch('services.tmdb_client.tmdb_client.get_popular_movies', new_callable=AsyncMock) as mock_tmdb, \
             patch('routers.dashboard_router._fetch_user_favorites', new_callable=AsyncMock) as mock_fav, \
             patch('routers.dashboard_router._fetch_user_watchlist', new_callable=AsyncMock) as mock_watch:
            
            mock_tmdb.return_value = mock_tmdb_movies
            mock_fav.side_effect = Exception("User service down")
            mock_watch.side_effect = Exception("User service down")
            
            response = client.get("/api/dashboard/feed")
            
            # Should still work, but without user data
            assert response.status_code == 200
            data = response.json()
            assert data["favorites_count"] == 0
            assert data["watchlist_count"] == 0
            assert len(data["movies"]) == 2
            
            # Movies should have default user data
            for movie in data["movies"]:
                assert movie["is_favorite"] is False
                assert movie["is_in_watchlist"] is False

