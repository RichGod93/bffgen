"""
Circuit breaker implementation for resilient API calls
"""
import time
import logging
from enum import Enum
from typing import Callable, Any
from config import settings

logger = logging.getLogger(__name__)


class CircuitState(Enum):
    """Circuit breaker states"""
    CLOSED = "closed"  # Normal operation
    OPEN = "open"      # Failing, reject requests
    HALF_OPEN = "half_open"  # Testing if service recovered


class CircuitBreaker:
    """
    Circuit breaker pattern implementation
    Prevents cascading failures by failing fast when a service is down
    """
    
    def __init__(
        self,
        name: str,
        failure_threshold: int = None,
        timeout: int = None
    ):
        self.name = name
        self.failure_threshold = failure_threshold or settings.CIRCUIT_BREAKER_FAILURE_THRESHOLD
        self.timeout = timeout or settings.CIRCUIT_BREAKER_TIMEOUT_SECONDS
        self.failure_count = 0
        self.last_failure_time = None
        self.state = CircuitState.CLOSED
    
    async def call(self, func: Callable, *args, **kwargs) -> Any:
        """Execute function with circuit breaker protection"""
        if not settings.CIRCUIT_BREAKER_ENABLED:
            return await func(*args, **kwargs)
        
        # Check if circuit should transition to HALF_OPEN
        if self.state == CircuitState.OPEN:
            if time.time() - self.last_failure_time >= self.timeout:
                logger.info(f"Circuit {self.name}: Transitioning to HALF_OPEN")
                self.state = CircuitState.HALF_OPEN
            else:
                raise Exception(f"Circuit {self.name} is OPEN - failing fast")
        
        try:
            result = await func(*args, **kwargs)
            
            # Success - reset or close circuit
            if self.state == CircuitState.HALF_OPEN:
                logger.info(f"Circuit {self.name}: Closing after successful call")
                self.state = CircuitState.CLOSED
                self.failure_count = 0
            
            return result
            
        except Exception as e:
            self.failure_count += 1
            self.last_failure_time = time.time()
            
            logger.warning(
                f"Circuit {self.name}: Failure {self.failure_count}/{self.failure_threshold} - {str(e)}"
            )
            
            # Open circuit if threshold reached
            if self.failure_count >= self.failure_threshold:
                logger.error(f"Circuit {self.name}: OPENING circuit")
                self.state = CircuitState.OPEN
            
            raise
    
    def reset(self):
        """Manually reset the circuit breaker"""
        self.state = CircuitState.CLOSED
        self.failure_count = 0
        self.last_failure_time = None
        logger.info(f"Circuit {self.name}: Manually reset")

