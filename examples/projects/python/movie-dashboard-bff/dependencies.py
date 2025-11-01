"""
Dependency injection for FastAPI routes
"""
from fastapi import Header, HTTPException, Depends
from jose import JWTError, jwt
from config import settings
from typing import Optional


async def get_current_user(authorization: Optional[str] = Header(None)) -> str:
    """
    Dependency to extract and validate JWT token from Authorization header
    
    For this demo, we'll use a simplified version that returns a mock user_id
    In production, implement proper JWT validation
    """
    # For demo purposes, return a mock user ID
    # In production, validate the JWT token properly
    if authorization and authorization.startswith("Bearer "):
        token = authorization.split(" ")[1]
        try:
            # Decode JWT (simplified for demo)
            payload = jwt.decode(token, settings.JWT_SECRET, algorithms=[settings.JWT_ALGORITHM])
            user_id = payload.get("sub")
            if user_id is None:
                raise HTTPException(status_code=401, detail="Invalid authentication credentials")
            return user_id
        except JWTError:
            raise HTTPException(status_code=401, detail="Invalid authentication credentials")
    
    # For demo/testing, return a default user ID
    return "demo-user-123"


async def verify_api_key(x_api_key: Optional[str] = Header(None)) -> bool:
    """
    Dependency to verify API key from header
    Optional - can be used for additional security
    """
    # For demo purposes, always return True
    # In production, validate against a secure store
    return True

