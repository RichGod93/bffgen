"""
Configuration settings for Movie Dashboard BFF
Uses Pydantic Settings for environment variable management
"""
from pydantic_settings import BaseSettings
from pydantic import field_validator
from typing import List, Union
import os


class Settings(BaseSettings):
    """Application settings"""
    
    # Application Settings
    PROJECT_NAME: str = "movie-dashboard-bff"
    VERSION: str = "1.0.0"
    PORT: int = 8000
    DEBUG: bool = True
    
    # CORS Settings
    CORS_ORIGINS: Union[List[str], str] = ["http://localhost:3000", "http://localhost:8080"]
    
    @field_validator('CORS_ORIGINS', mode='before')
    @classmethod
    def parse_cors_origins(cls, v):
        """Parse CORS_ORIGINS from comma-separated string or list"""
        if isinstance(v, str):
            return [origin.strip() for origin in v.split(',')]
        return v
    
    # JWT Settings
    JWT_SECRET: str = "your-secret-key-change-in-production"
    JWT_ALGORITHM: str = "HS256"
    JWT_EXPIRATION_MINUTES: int = 30
    
    # Rate Limiting
    RATE_LIMIT_ENABLED: bool = True
    RATE_LIMIT_PER_MINUTE: int = 60
    
    # Redis/Cache Settings
    REDIS_URL: str = "redis://localhost:6379"
    CACHE_TTL: int = 300  # 5 minutes
    CACHE_ENABLED: bool = False
    
    # Circuit Breaker Settings
    CIRCUIT_BREAKER_ENABLED: bool = True
    CIRCUIT_BREAKER_FAILURE_THRESHOLD: int = 5
    CIRCUIT_BREAKER_TIMEOUT_SECONDS: int = 60
    
    # Logging
    LOG_LEVEL: str = "INFO"
    
    # TMDB Configuration (READ FROM ENVIRONMENT)
    TMDB_API_KEY: str = os.getenv("TMDB_API_KEY", "")
    TMDB_READ_TOKEN: str = os.getenv("TMDB_READ_TOKEN", "")
    TMDB_BASE_URL: str = "https://api.themoviedb.org/3"
    
    class Config:
        env_file = ".env"
        case_sensitive = True


settings = Settings()

