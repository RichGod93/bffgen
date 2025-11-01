"""
Tests for Mock User Service
"""
import pytest
from fastapi.testclient import TestClient
import sys
import os

# Add parent directory to path
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from main import app, favorites_store, watchlist_store


class TestMockUserService:
    """Test suite for mock user service"""
    
    @pytest.fixture
    def client(self):
        """Test client"""
        return TestClient(app)
    
    @pytest.fixture(autouse=True)
    def reset_stores(self):
        """Reset stores before each test"""
        favorites_store.clear()
        watchlist_store.clear()
        yield
        favorites_store.clear()
        watchlist_store.clear()
    
    def test_health_check(self, client):
        """Test health check endpoint"""
        response = client.get("/health")
        
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"
        assert "favorites_count" in data
        assert "watchlist_count" in data
    
    # === Favorites Tests ===
    
    def test_get_favorites_empty(self, client):
        """Test getting empty favorites"""
        response = client.get("/favorites")
        
        assert response.status_code == 200
        data = response.json()
        assert data["favorites"] == []
        assert data["total"] == 0
    
    def test_add_favorite(self, client):
        """Test adding a movie to favorites"""
        payload = {
            "movie_id": 550,
            "title": "Fight Club",
            "rating": 5
        }
        
        response = client.post("/favorites", json=payload)
        
        assert response.status_code == 200
        data = response.json()
        assert data["message"] == "Movie added to favorites"
        assert data["favorite"]["movie_id"] == 550
        assert data["favorite"]["rating"] == 5
    
    def test_add_duplicate_favorite(self, client):
        """Test adding duplicate favorite returns error"""
        payload = {"movie_id": 550, "title": "Fight Club"}
        
        # Add first time
        client.post("/favorites", json=payload)
        
        # Try to add again
        response = client.post("/favorites", json=payload)
        
        assert response.status_code == 409
        assert "already in favorites" in response.json()["detail"].lower()
    
    def test_get_favorites_after_add(self, client):
        """Test retrieving favorites after adding"""
        client.post("/favorites", json={"movie_id": 550, "title": "Fight Club", "rating": 5})
        client.post("/favorites", json={"movie_id": 155, "title": "The Dark Knight", "rating": 4})
        
        response = client.get("/favorites")
        
        assert response.status_code == 200
        data = response.json()
        assert data["total"] == 2
        assert len(data["favorites"]) == 2
    
    def test_get_specific_favorite(self, client):
        """Test getting a specific favorite"""
        client.post("/favorites", json={"movie_id": 550, "title": "Fight Club"})
        
        response = client.get("/favorites/550")
        
        assert response.status_code == 200
        data = response.json()
        assert data["movie_id"] == 550
        assert data["title"] == "Fight Club"
    
    def test_get_nonexistent_favorite(self, client):
        """Test getting a favorite that doesn't exist"""
        response = client.get("/favorites/999")
        
        assert response.status_code == 404
    
    def test_delete_favorite(self, client):
        """Test removing a movie from favorites"""
        client.post("/favorites", json={"movie_id": 550, "title": "Fight Club"})
        
        response = client.delete("/favorites/550")
        
        assert response.status_code == 200
        data = response.json()
        assert data["message"] == "Movie removed from favorites"
        
        # Verify it's gone
        response = client.get("/favorites")
        assert response.json()["total"] == 0
    
    def test_delete_nonexistent_favorite(self, client):
        """Test deleting a favorite that doesn't exist"""
        response = client.delete("/favorites/999")
        
        assert response.status_code == 404
    
    # === Watchlist Tests ===
    
    def test_get_watchlist_empty(self, client):
        """Test getting empty watchlist"""
        response = client.get("/watchlist")
        
        assert response.status_code == 200
        data = response.json()
        assert data["watchlist"] == []
        assert data["total"] == 0
    
    def test_add_to_watchlist(self, client):
        """Test adding a movie to watchlist"""
        payload = {
            "movie_id": 680,
            "title": "Pulp Fiction",
            "release_date": "1994-10-14"
        }
        
        response = client.post("/watchlist", json=payload)
        
        assert response.status_code == 200
        data = response.json()
        assert data["message"] == "Movie added to watchlist"
        assert data["watchlist_item"]["movie_id"] == 680
    
    def test_add_duplicate_to_watchlist(self, client):
        """Test adding duplicate to watchlist returns error"""
        payload = {"movie_id": 680, "title": "Pulp Fiction"}
        
        client.post("/watchlist", json=payload)
        response = client.post("/watchlist", json=payload)
        
        assert response.status_code == 409
    
    def test_get_watchlist_after_add(self, client):
        """Test retrieving watchlist after adding"""
        client.post("/watchlist", json={"movie_id": 680, "title": "Pulp Fiction"})
        client.post("/watchlist", json={"movie_id": 27205, "title": "Inception"})
        
        response = client.get("/watchlist")
        
        assert response.status_code == 200
        data = response.json()
        assert data["total"] == 2
    
    def test_delete_from_watchlist(self, client):
        """Test removing a movie from watchlist"""
        client.post("/watchlist", json={"movie_id": 680, "title": "Pulp Fiction"})
        
        response = client.delete("/watchlist/680")
        
        assert response.status_code == 200
        
        # Verify it's gone
        response = client.get("/watchlist")
        assert response.json()["total"] == 0
    
    # === Utility Endpoints Tests ===
    
    def test_reset_data(self, client):
        """Test resetting to sample data"""
        # Add some data
        client.post("/favorites", json={"movie_id": 999, "title": "Test Movie"})
        
        # Reset
        response = client.post("/reset")
        
        assert response.status_code == 200
        data = response.json()
        assert data["message"] == "Data reset successfully"
        assert data["favorites_count"] > 0  # Sample data restored
        
        # Verify sample data is present
        response = client.get("/favorites")
        favorites = response.json()["favorites"]
        assert any(fav["movie_id"] == 550 for fav in favorites)  # Fight Club
    
    def test_clear_all_data(self, client):
        """Test clearing all data"""
        # Add some data
        client.post("/favorites", json={"movie_id": 550, "title": "Fight Club"})
        client.post("/watchlist", json={"movie_id": 680, "title": "Pulp Fiction"})
        
        # Clear all
        response = client.delete("/clear")
        
        assert response.status_code == 200
        data = response.json()
        assert data["favorites_count"] == 0
        assert data["watchlist_count"] == 0
        
        # Verify empty
        assert client.get("/favorites").json()["total"] == 0
        assert client.get("/watchlist").json()["total"] == 0
    
    def test_add_favorite_without_movie_id(self, client):
        """Test validation: movie_id is required"""
        response = client.post("/favorites", json={"title": "Test"})
        
        assert response.status_code == 400
    
    def test_add_watchlist_without_movie_id(self, client):
        """Test validation: movie_id is required for watchlist"""
        response = client.post("/watchlist", json={"title": "Test"})
        
        assert response.status_code == 400

