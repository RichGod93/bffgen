"""
Cache manager for Redis-based caching
"""
import redis.asyncio as redis
import json
import logging
from typing import Optional, Any
from config import settings

logger = logging.getLogger(__name__)


class CacheManager:
    """Async Redis cache manager"""
    
    def __init__(self):
        self.redis_client: Optional[redis.Redis] = None
        self.enabled = settings.CACHE_ENABLED
    
    async def connect(self):
        """Connect to Redis"""
        if not self.enabled:
            logger.info("Cache is disabled")
            return
        
        try:
            self.redis_client = redis.from_url(
                settings.REDIS_URL,
                encoding="utf-8",
                decode_responses=True
            )
            await self.redis_client.ping()
            logger.info("Connected to Redis cache")
        except Exception as e:
            logger.error(f"Failed to connect to Redis: {e}")
            self.enabled = False
    
    async def get(self, key: str) -> Optional[Any]:
        """Get value from cache"""
        if not self.enabled or not self.redis_client:
            return None
        
        try:
            value = await self.redis_client.get(key)
            if value:
                return json.loads(value)
            return None
        except Exception as e:
            logger.error(f"Cache get error for key {key}: {e}")
            return None
    
    async def set(self, key: str, value: Any, ttl: Optional[int] = None):
        """Set value in cache with TTL"""
        if not self.enabled or not self.redis_client:
            return
        
        try:
            ttl = ttl or settings.CACHE_TTL
            await self.redis_client.setex(
                key,
                ttl,
                json.dumps(value)
            )
        except Exception as e:
            logger.error(f"Cache set error for key {key}: {e}")
    
    async def delete(self, key: str):
        """Delete key from cache"""
        if not self.enabled or not self.redis_client:
            return
        
        try:
            await self.redis_client.delete(key)
        except Exception as e:
            logger.error(f"Cache delete error for key {key}: {e}")
    
    async def close(self):
        """Close Redis connection"""
        if self.redis_client:
            await self.redis_client.close()


# Global cache instance
cache = CacheManager()

