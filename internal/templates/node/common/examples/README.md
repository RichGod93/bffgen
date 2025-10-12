# BFFGen Aggregation Examples

This directory contains complete working examples demonstrating advanced BFF aggregation patterns.

## Examples

### 1. User Dashboard Controller (`user-dashboard.controller.js`)

**Demonstrates:**

- Parallel service aggregation
- Cache management
- Response transformation
- Graceful degradation

**Use Case:** Aggregate user profile, recent orders, and preferences into a single dashboard endpoint.

**Key Features:**

- Fetches from 3 services simultaneously
- Caches results for 60 seconds
- Continues with partial data if services fail
- Filters response to only needed fields

**Endpoint:** `GET /api/dashboard/:userId`

**Response:**

```json
{
  "user": {
    "id": "123",
    "name": "John Doe",
    "email": "john@example.com",
    "avatar": "https://..."
  },
  "orders": [{ "id": "order-1", "total": 99.99, "status": "delivered" }],
  "settings": {
    "theme": "dark",
    "notifications": true
  },
  "hasErrors": false
}
```

---

### 2. E-commerce Catalog Controller (`ecommerce-catalog.controller.js`)

**Demonstrates:**

- Parallel service aggregation
- Circuit breaker for fault tolerance
- Request batching to prevent N+1 queries
- Advanced caching strategies

**Use Case:** Build product catalog with inventory and ratings from multiple services.

**Key Features:**

- Fetches products and inventory in parallel
- Uses circuit breaker for unreliable inventory service
- Batches rating requests (100 products → 2-3 API calls)
- Caches catalog for 5 minutes

**Endpoint:** `GET /api/catalog?category=electronics&page=1&limit=20`

**Response:**

```json
{
  "products": [
    {
      "id": "prod-1",
      "name": "Product Name",
      "price": 29.99,
      "images": ["img1.jpg"],
      "inStock": true,
      "quantity": 50,
      "rating": {
        "score": "4.5",
        "count": 128
      }
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100
  },
  "metadata": {
    "category": "electronics",
    "inventoryAvailable": true,
    "timestamp": "2024-01-20T10:30:00Z"
  }
}
```

---

## How to Use These Examples

### Option 1: Copy and Adapt

Copy the example controller to your `src/controllers/` directory and modify for your needs:

```bash
cp src/examples/user-dashboard.controller.js src/controllers/
```

### Option 2: Use as Reference

Study the examples to understand patterns, then implement your own controllers.

### Option 3: Import Directly

Import and use the example controllers in your routes:

```javascript
// Express
const dashboard = require("./examples/user-dashboard.controller");
router.get("/dashboard/:userId", dashboard.getDashboard.bind(dashboard));

// Fastify
const dashboard = require("./examples/user-dashboard.controller");
fastify.get("/dashboard/:userId", dashboard.getDashboard.bind(dashboard));
```

---

## Customization

### Replace Mock Data

The examples include mock data for demonstration. Replace with your actual service calls:

```javascript
// Replace this:
return { id: userId, name: "Mock Data" };

// With this:
return await UserService.getProfile(userId, headers);
```

### Adjust Timeouts and TTLs

Modify based on your requirements:

```javascript
// Aggressive caching for catalog
this.cache = new CacheManager({ ttl: 900, prefix: "catalog:" });

// Short caching for user data
this.cache = new CacheManager({ ttl: 60, prefix: "user:" });
```

### Add More Services

Extend the parallel aggregation:

```javascript
const results = await this.aggregator.fetchParallel([
  { name: "user", fetch: () => UserService.get(id) },
  { name: "orders", fetch: () => OrdersService.get(id) },
  { name: "preferences", fetch: () => PreferencesService.get(id) },
  // Add more services
  { name: "recommendations", fetch: () => RecommendationService.get(id) },
  { name: "notifications", fetch: () => NotificationService.getUnread(id) },
]);
```

---

## Testing

Run tests for the examples:

```bash
npm run test:aggregation
```

## Performance Monitoring

Add monitoring to track aggregation performance:

```javascript
async getDashboard(req, res, next) {
  const startTime = Date.now();

  try {
    // ... your aggregation logic

    const duration = Date.now() - startTime;
    req.log.info(`Dashboard aggregation took ${duration}ms`);

  } catch (error) {
    req.log.error('Dashboard aggregation failed', error);
    next(error);
  }
}
```

---

## Related Documentation

- [Node.js Aggregation Guide](../../../../docs/NODEJS_AGGREGATION.md)
- [Utility API Reference](../utils/)
- [Testing Guide](../../../../docs/NODEJS_TESTING.md)

---

## Support

For questions or issues:

1. Check the [Aggregation Guide](../../../../docs/NODEJS_AGGREGATION.md)
2. Review utility source code in `src/utils/`
3. Open an issue on GitHub
