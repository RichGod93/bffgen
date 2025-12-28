const stripe = require('stripe')(process.env.STRIPE_SECRET_KEY);

/**
 * Stripe Billing Integration
 * Project: {{PROJECT_NAME}}
 */

class BillingService {
  // Create a new customer
  async createCustomer(email, name) {
    const customer = await stripe.customers.create({
      email,
      name,
      metadata: {
        project: '{{PROJECT_NAME}}'
      }
    });
    return customer;
  }

  // Create a subscription
  async createSubscription(customerId, priceId) {
    const subscription = await stripe.subscriptions.create({
      customer: customerId,
      items: [{ price: priceId }],
      payment_behavior: 'default_incomplete',
      expand: ['latest_invoice.payment_intent'],
    });
    return subscription;
  }

  // Cancel a subscription
  async cancelSubscription(subscriptionId) {
    const subscription = await stripe.subscriptions.del(subscriptionId);
    return subscription;
  }

  // Get subscription details
  async getSubscription(subscriptionId) {
    const subscription = await stripe.subscriptions.retrieve(subscriptionId);
    return subscription;
  }
}

module.exports = new BillingService();
