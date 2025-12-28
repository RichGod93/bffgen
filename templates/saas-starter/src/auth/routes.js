const express = require('express');
const bcrypt = require('bcryptjs');
const { generateToken, generateRefreshToken } = require('./jwt');

const router = express.Router();

/**
 * Authentication Routes
 * Project: {{PROJECT_NAME}}
 */

// In-memory user store (replace with database in production)
const users = new Map();

// Register new user
router.post('/register', async (req, res) => {
  try {
    const { email, password, name } = req.body;

    // Validation
    if (!email || !password) {
      return res.status(400).json({ error: 'Email and password are required' });
    }

    // Check if user exists
    if (users.has(email)) {
      return res.status(409).json({ error: 'User already exists' });
    }

    // Hash password
    const hashedPassword = await bcrypt.hash(password, 10);

    // Create user
    const user = {
      id: Date.now().toString(),
      email,
      name: name || email.split('@')[0],
      password: hashedPassword,
      role: 'user',
      createdAt: new Date().toISOString()
    };

    users.set(email, user);

    // Generate tokens
    const token = generateToken({ id: user.id, email: user.email, role: user.role });
    const refreshToken = generateRefreshToken({ id: user.id });

    res.status(201).json({
      user: {
        id: user.id,
        email: user.email,
        name: user.name,
        role: user.role
      },
      token,
      refreshToken
    });
  } catch (error) {
    console.error('Registration error:', error);
    res.status(500).json({ error: 'Registration failed' });
  }
});

// Login
router.post('/login', async (req, res) => {
  try {
    const { email, password } = req.body;

    // Validation
    if (!email || !password) {
      return res.status(400).json({ error: 'Email and password are required' });
    }

    // Find user
    const user = users.get(email);
    if (!user) {
      return res.status(401).json({ error: 'Invalid credentials' });
    }

    // Verify password
    const validPassword = await bcrypt.compare(password, user.password);
    if (!validPassword) {
      return res.status(401).json({ error: 'Invalid credentials' });
    }

    // Generate tokens
    const token = generateToken({ id: user.id, email: user.email, role: user.role });
    const refreshToken = generateRefreshToken({ id: user.id });

    res.json({
      user: {
        id: user.id,
        email: user.email,
        name: user.name,
        role: user.role
      },
      token,
      refreshToken
    });
  } catch (error) {
    console.error('Login error:', error);
    res.status(500).json({ error: 'Login failed' });
  }
});

// Refresh token
router.post('/refresh', async (req, res) => {
  try {
    const { refreshToken } = req.body;

    if (!refreshToken) {
      return res.status(400).json({ error: 'Refresh token required' });
    }

    // Verify refresh token (simplified - use proper verification)
    // TODO: Implement proper refresh token validation with database storage
    
    const token = generateToken({ id: 'user-id', email: 'user@example.com', role: 'user' });
    const newRefreshToken = generateRefreshToken({ id: 'user-id' });

    res.json({ token, refreshToken: newRefreshToken });
  } catch (error) {
    console.error('Token refresh error:', error);
    res.status(401).json({ error: 'Invalid refresh token' });
  }
});

// Logout
router.post('/logout', (req, res) => {
  // TODO: Implement token blacklisting
  res.json({ message: 'Logged out successfully' });
});

module.exports = router;
