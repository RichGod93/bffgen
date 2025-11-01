"""
Logging middleware for request/response tracking
"""
import time
import logging
from fastapi import Request
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.responses import Response

logger = logging.getLogger(__name__)


class LoggingMiddleware(BaseHTTPMiddleware):
    """Middleware to log all requests and responses"""
    
    async def dispatch(self, request: Request, call_next):
        # Log request
        start_time = time.time()
        request_id = request.headers.get("X-Request-ID", "no-id")
        
        logger.info(
            f"Request started: {request.method} {request.url.path} "
            f"[ID: {request_id}]"
        )
        
        # Process request
        response = await call_next(request)
        
        # Log response
        duration = time.time() - start_time
        logger.info(
            f"Request completed: {request.method} {request.url.path} "
            f"[ID: {request_id}] - Status: {response.status_code} - "
            f"Duration: {duration:.3f}s"
        )
        
        # Add response headers
        response.headers["X-Request-ID"] = request_id
        response.headers["X-Process-Time"] = f"{duration:.3f}"
        
        return response

