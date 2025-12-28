const express = require('express');
const { authenticateToken } = require('../auth/jwt');

const router = express.Router();

/**
 * Admin Dashboard Routes  
 * Project: {{PROJECT_NAME}}
 */

// Dashboard stats
router.get('/stats', authenticateToken, async (req, res) => {
  // TODO: Implement admin check
  // TODO: Fetch real stats from database
  
  res.json({
    users: { total: 150, active: 120, new: 12 },
    revenue: { total: 45000, monthly: 15000, growth: 0.15 },
    subscriptions: { active: 85, trial: 30, cancelled: 5 }
  });
});

// Recent activity
router.get('/activity', authenticateToken, async (req, res) => {
  res.json({
    activities: [
      { type: 'user_signup', user: 'user@example.com', timestamp: new Date() },
      { type: 'subscription_created', user: 'customer@example.com', timestamp: new Date() }
    ]
  });
});

module.exports = router;
