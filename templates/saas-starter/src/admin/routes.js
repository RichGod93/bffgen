const express = require('express');
const { authenticateToken, requireAdmin } = require('../auth/jwt');

const router = express.Router();

/**
 * Admin Routes
 * Project: {{PROJECT_NAME}}
 */

// Get dashboard statistics
router.get('/stats', authenticateToken, requireAdmin, async (req, res) => {
  // TODO: Fetch real statistics from database
  res.json({
    stats: {
      totalUsers: 156,
      activeUsers: 89,
      totalRevenue: 12450.00,
      monthlyRevenue: 3200.00,
      newSignups: 23,
      churnRate: 2.5
    }
  });
});

// Get recent activity
router.get('/activity', authenticateToken, requireAdmin, async (req, res) => {
  // TODO: Fetch real activity from database
  res.json({
    activities: [
      {
        id: 1,
        type: 'user_signup',
        user: 'john@example.com',
        timestamp: new Date(Date.now() - 1000 * 60 * 15).toISOString(),
        details: 'New user registered'
      },
      {
        id: 2,
        type: 'payment_successful',
        user: 'jane@example.com',
        amount: 49.99,
        timestamp: new Date(Date.now() - 1000 * 60 * 30).toISOString(),
        details: 'Pro plan subscription'
      },
      {
        id: 3,
        type: 'user_login',
        user: 'admin@example.com',
        timestamp: new Date(Date.now() - 1000 * 60 * 45).toISOString(),
        details: 'Admin login'
      }
    ]
  });
});

// Get system settings
router.get('/settings', authenticateToken, requireAdmin, async (req, res) => {
  // TODO: Fetch settings from database
  res.json({
    settings: {
      maintenanceMode: false,
      allowSignups: true,
      requireEmailVerification: true,
      maxUsersPerAccount: 10
    }
  });
});

// Update system settings
router.put('/settings', authenticateToken, requireAdmin, async (req, res) => {
  const updates = req.body;
  
  // TODO: Update settings in database
  res.json({
    settings: updates,
    message: 'Settings updated successfully'
  });
});

module.exports = router;
