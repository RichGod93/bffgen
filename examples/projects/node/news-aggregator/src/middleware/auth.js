/**
 * Authentication Middleware
 * Validates JWT tokens and adds user context to requests
 */

const jwt = require('jsonwebtoken');

/**
 * Middleware to verify JWT token
 * @param {Object} req - Express request object
 * @param {Object} res - Express response object
 * @param {Function} next - Express next function
 */
function authenticate(req, res, next) {
  try {
    const authHeader = req.headers.authorization;
    
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return res.status(401).json({
        error: 'Unauthorized',
        message: 'No token provided'
      });
    }
    
    const token = authHeader.substring(7); // Remove 'Bearer ' prefix
    const jwtSecret = process.env.JWT_SECRET || 'your-secret-key-change-in-production';
    
    // Verify token
    const decoded = jwt.verify(token, jwtSecret);
    
    // Add user info to request
    req.user = decoded;
    req.userId = decoded.userId || decoded.sub;
    
    next();
  } catch (error) {
    if (error.name === 'TokenExpiredError') {
      return res.status(401).json({
        error: 'Unauthorized',
        message: 'Token expired'
      });
    }
    
    if (error.name === 'JsonWebTokenError') {
      return res.status(401).json({
        error: 'Unauthorized',
        message: 'Invalid token'
      });
    }
    
    return res.status(500).json({
      error: 'Internal Server Error',
      message: 'Authentication error'
    });
  }
}

/**
 * Optional authentication - doesn't fail if no token
 * Useful for endpoints that work differently for authenticated users
 */
function optionalAuth(req, res, next) {
  try {
    const authHeader = req.headers.authorization;
    
    if (authHeader && authHeader.startsWith('Bearer ')) {
      const token = authHeader.substring(7);
      const jwtSecret = process.env.JWT_SECRET || 'your-secret-key-change-in-production';
      
      const decoded = jwt.verify(token, jwtSecret);
      req.user = decoded;
      req.userId = decoded.userId || decoded.sub;
    }
    
    next();
  } catch (error) {
    // Ignore errors, continue without authentication
    next();
  }
}

module.exports = {
  authenticate,
  optionalAuth
};

