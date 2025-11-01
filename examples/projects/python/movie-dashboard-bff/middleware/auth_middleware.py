"""
Authentication middleware for JWT token validation
"""
import logging
from fastapi import Request
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.responses import JSONResponse
from jose import JWTError, jwt
from config import settings

logger = logging.getLogger(__name__)


class AuthMiddleware(BaseHTTPMiddleware):
    """Middleware to validate JWT tokens on protected routes"""
    
    # Routes that don't require authentication
    PUBLIC_PATHS = ["/health", "/docs", "/redoc", "/openapi.json"]
    
    async def dispatch(self, request: Request, call_next):
        # Skip authentication for public paths
        if any(request.url.path.startswith(path) for path in self.PUBLIC_PATHS):
            return await call_next(request)
        
        # Get authorization header
        auth_header = request.headers.get("Authorization")
        
        if not auth_header or not auth_header.startswith("Bearer "):
            return JSONResponse(
                status_code=401,
                content={"detail": "Missing or invalid authorization header"}
            )
        
        # Extract and validate token
        token = auth_header.split(" ")[1]
        
        try:
            payload = jwt.decode(
                token,
                settings.JWT_SECRET,
                algorithms=[settings.JWT_ALGORITHM]
            )
            # Add user info to request state
            request.state.user_id = payload.get("sub")
            request.state.user = payload
            
        except JWTError as e:
            logger.warning(f"JWT validation failed: {e}")
            return JSONResponse(
                status_code=401,
                content={"detail": "Invalid authentication credentials"}
            )
        
        return await call_next(request)

