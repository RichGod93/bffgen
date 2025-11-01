"""
Pytest configuration and fixtures
"""
import pytest
from fastapi.testclient import TestClient
from main import app


@pytest.fixture
def client():
    """Test client fixture"""
    return TestClient(app)


@pytest.fixture
def mock_user_id():
    """Mock user ID for testing"""
    return "test-user-123"


@pytest.fixture
def auth_headers(mock_user_id):
    """Mock authentication headers"""
    return {
        "Authorization": f"Bearer mock-token-{mock_user_id}"
    }

