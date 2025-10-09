# Node.js Aggregation Guide

## Overview

This guide covers the advanced aggregation utilities available in bffgen-generated Node.js BFF projects. These utilities enable you to create production-ready Backend-for-Frontend services that aggregate data from multiple microservices efficiently and reliably.

## Table of Contents

1. [Core Utilities](#core-utilities)
2. [Quick Start](#quick-start)
3. [Parallel Aggregator](#parallel-aggregator)
4. [Cache Manager](#cache-manager)
5. [Circuit Breaker](#circuit-breaker)
6. [Response Transformer](#response-transformer)
7. [Request Batcher](#request-batcher)
8. [Field Selector](#field-selector)
9. [Complete Examples](#complete-examples)
10. [Best Practices](#best-practices)
11. [Performance Optimization](#performance-optimization)

---

## Core Utilities

All generated Node.js projects include the following utilities in `src/utils/`:

- **aggregator.js** - Parallel and waterfall service execution
- **cache-manager.js** - Redis-backed caching with in-memory fallback
- **circuit-breaker.js** - Fault tolerance and graceful degradation
- **response-transformer.js** - Data transformation and sanitization
- **request-batcher.js** - Request batching to prevent N+1 queries
- **field-selector.js** - GraphQL-like field selection for REST

---

## Quick Start

### Basic Parallel Aggregation

```javascript
const ParallelAggregator = require("../utils/aggregator");
const aggregator = new ParallelAggregator({ timeout: 5000 });

// Fetch from multiple services simultaneously
const results = await aggregator.fetchParallel([
  { name: "user", fetch: () => UserService.getProfile(userId) },
  { name: "orders", fetch: () => OrdersService.getOrders(userId) },
  {
    name: "preferences",
    fetch: () => PreferencesService.getPreferences(userId),
  },
]);

// Access results with graceful degradation
const user = results.find((r) => r.service === "user" && r.success)?.data;
const orders =
  results.find((r) => r.service === "orders" && r.success)?.data || [];
```

### With Caching

```javascript
const CacheManager = require("../utils/cache-manager");
const cache = new CacheManager({ ttl: 300, prefix: "dashboard:" });

// Check cache first
const cached = await cache.get(`user:${userId}`);
if (cached) {
  return cached;
}

// Fetch and cache result
const data = await fetchUserData(userId);
await cache.set(`user:${userId}`, data, 300);
```

### With Circuit Breaker

```javascript
const CircuitBreaker = require("../utils/circuit-breaker");
const breaker = new CircuitBreaker({ failureThreshold: 5 });

// Execute with automatic fallback
const result = await breaker.execute(
  () => InventoryService.getStock(productId),
  () => ({ inStock: false, message: "Inventory service unavailable" })
);
```

---

## Parallel Aggregator

### Purpose

Execute multiple service calls simultaneously to reduce total response time.

### Configuration

```javascript
const aggregator = new ParallelAggregator({
  timeout: 30000, // Maximum time for each service call (ms)
  failFast: false, // If true, fail immediately on first error
});
```

### Parallel Execution

```javascript
const results = await aggregator.fetchParallel([
  {
    name: "service1",
    fetch: async () => {
      // Your service call here
      return await fetch("http://api.example.com/data");
    },
  },
  {
    name: "service2",
    fetch: async () => {
      return await fetch("http://api2.example.com/data");
    },
  },
]);

// Result format:
// [
//   { service: 'service1', data: {...}, error: null, success: true },
//   { service: 'service2', data: {...}, error: null, success: true }
// ]
```

### Waterfall Execution

For dependent calls where each service needs results from the previous:

```javascript
const results = await aggregator.fetchWaterfall([
  {
    name: "user",
    fetch: async () => await UserService.getUser(userId),
  },
  {
    name: "permissions",
    fetch: async (previousResults) => {
      const user = previousResults[0].data;
      return await PermissionsService.getPermissions(user.roleId);
    },
  },
]);
```

### Error Handling

```javascript
const results = await aggregator.fetchParallel(services);

// Check for failures
const failed = results.filter((r) => !r.success);
if (failed.length > 0) {
  console.warn(
    "Some services failed:",
    failed.map((f) => f.service)
  );
}

// Use partial data
const successfulData = results
  .filter((r) => r.success)
  .reduce((acc, r) => {
    acc[r.service] = r.data;
    return acc;
  }, {});
```

---

## Cache Manager

### Purpose

Reduce backend load and improve response times with Redis-backed caching.

### Configuration

```javascript
const cache = new CacheManager({
  ttl: 300, // Default time-to-live in seconds
  prefix: "myapp:", // Key prefix for organization
});

// Requires REDIS_URL environment variable
// Falls back to in-memory cache if Redis unavailable
```

### Basic Operations

```javascript
// Set a value (with custom TTL)
await cache.set("user:123", userData, 60);

// Get a value
const data = await cache.get("user:123");

// Delete a value
await cache.delete("user:123");

// Clear all keys with prefix
await cache.clear();

// Check if using Redis
if (cache.isUsingRedis()) {
  console.log("Using Redis cache");
}
```

### Caching Strategy

```javascript
async function getUserData(userId) {
  const cacheKey = `user:${userId}`;

  // 1. Try cache first
  const cached = await cache.get(cacheKey);
  if (cached) {
    return cached;
  }

  // 2. Fetch from source
  const data = await UserService.getProfile(userId);

  // 3. Cache for future requests
  await cache.set(cacheKey, data, 300);

  return data;
}
```

### Cache Invalidation

```javascript
// Invalidate on update
async function updateUser(userId, updates) {
  await UserService.update(userId, updates);
  await cache.delete(`user:${userId}`);
}

// Invalidate multiple related keys
async function updateOrder(orderId, updates) {
  const order = await OrderService.update(orderId, updates);

  // Invalidate order cache
  await cache.delete(`order:${orderId}`);

  // Invalidate user's orders list
  await cache.delete(`user:${order.userId}:orders`);
}
```

### Redis Setup

```bash
# Using Docker
docker-compose up redis

# Or install locally
brew install redis  # macOS
sudo apt install redis  # Ubuntu

# Set environment variable
export REDIS_URL=redis://localhost:6379

# Clear cache manually
npm run cache:clear
```

---

## Circuit Breaker

### Purpose

Prevent cascade failures by automatically stopping requests to failing services.

### States

- **CLOSED**: Normal operation, requests pass through
- **OPEN**: Service is failing, requests blocked (use fallback)
- **HALF_OPEN**: Testing if service recovered

### Configuration

```javascript
const breaker = new CircuitBreaker({
  failureThreshold: 5, // Open after 5 consecutive failures
  resetTimeout: 60000, // Try again after 60 seconds
  monitoringPeriod: 10000, // 10 second monitoring window
});
```

### Usage

```javascript
// With fallback
const data = await breaker.execute(
  () => RiskService.checkFraud(transactionId),
  () => ({ riskScore: 0, checked: false }) // Fallback when circuit open
);

// Without fallback (will throw error when open)
try {
  const data = await breaker.execute(() =>
    RecommendationService.getRecommendations(userId)
  );
} catch (error) {
  // Handle circuit open or service error
}
```

### Monitoring

```javascript
// Get circuit state
const state = breaker.getState();
console.log(state);
// {
//   state: 'CLOSED',
//   failures: 0,
//   stats: { total: 100, failed: 5, succeeded: 95 },
//   nextAttempt: null
// }

// Manual control
breaker.reset(); // Reset to CLOSED
breaker.forceOpen(); // Force OPEN state
```

### Health Check Endpoint

```javascript
app.get("/health/circuits", (req, res) => {
  res.json({
    inventory: inventoryBreaker.getState(),
    recommendations: recommendationBreaker.getState(),
    fraud: fraudBreaker.getState(),
  });
});
```

---

## Response Transformer

### Purpose

Transform, filter, and sanitize API responses for frontend consumption.

### Field Selection

```javascript
const ResponseTransformer = require("../utils/response-transformer");

// Pick specific fields
const user = ResponseTransformer.pick(fullUser, ["id", "name", "email"]);

// Omit sensitive fields
const safe = ResponseTransformer.omit(user, ["password", "ssn"]);

// Sanitize (removes common sensitive fields)
const sanitized = ResponseTransformer.sanitize(user);
```

### Object Transformation

```javascript
// Rename fields
const renamed = ResponseTransformer.rename(user, {
  first_name: "firstName",
  last_name: "lastName",
});

// Flatten nested objects
const flat = ResponseTransformer.flatten({
  user: { name: "John", address: { city: "NYC" } },
});
// Result: { 'user.name': 'John', 'user.address.city': 'NYC' }

// Merge multiple objects
const combined = ResponseTransformer.merge(
  { id: 1, name: "John" },
  { email: "john@example.com" },
  { role: "admin" }
);
```

### Computed Fields

```javascript
const enriched = ResponseTransformer.addComputedFields(user, {
  fullName: (u) => `${u.firstName} ${u.lastName}`,
  isAdmin: (u) => u.role === "admin",
  avatarUrl: (u) => `https://cdn.example.com/avatars/${u.id}.jpg`,
});
```

### Array Transformations

```javascript
// Transform array items
const simplified = ResponseTransformer.transformArray(products, (product) =>
  ResponseTransformer.pick(product, ["id", "name", "price"])
);

// Map keys/values
const upperCaseKeys = ResponseTransformer.mapKeys(obj, (key) =>
  key.toUpperCase()
);
const doubled = ResponseTransformer.mapValues(obj, (value) => value * 2);
```

---

## Request Batcher

### Purpose

Batch multiple requests to the same service to prevent N+1 query problems.

### Configuration

```javascript
const batcher = new RequestBatcher({
  batchWindow: 10, // Wait 10ms to collect requests
  maxBatchSize: 50, // Execute when 50 requests collected
});
```

### Usage

```javascript
// Problem: N+1 queries
for (const product of products) {
  const rating = await RatingService.getRating(product.id); // 100 calls!
  product.rating = rating;
}

// Solution: Batching
const enriched = await Promise.all(
  products.map(async (product) => {
    const rating = await batcher.batch(
      "ratings",
      (productIds) => RatingService.getBulkRatings(productIds),
      product.id
    );
    return { ...product, rating };
  })
);
// Only 1-2 calls (depending on batch window)!
```

### Batch Function Requirements

The batch function receives an array of IDs and must return an array of objects with `id` field:

```javascript
async function getBulkRatings(productIds) {
  const response = await fetch("/api/ratings/bulk", {
    method: "POST",
    body: JSON.stringify({ productIds }),
  });
  const data = await response.json();

  // Must return array with id field
  return data.map((rating) => ({
    id: rating.productId, // Required!
    score: rating.score,
    count: rating.count,
  }));
}
```

### Monitoring

```javascript
// Get statistics
const stats = batcher.getStats();
console.log(stats);
// {
//   activeBatches: 2,
//   totalPendingRequests: 45,
//   batches: [
//     { key: 'ratings', pending: 23 },
//     { key: 'inventory', pending: 22 }
//   ]
// }

// Clear specific batch
batcher.clear("ratings");

// Clear all batches
batcher.clearAll();
```

---

## Field Selector

### Purpose

Allow clients to specify which fields they need (like GraphQL for REST).

### Basic Usage

```javascript
const FieldSelector = require("../utils/field-selector");

// GET /api/users/123?fields=id,name,email
const selected = FieldSelector.selectFields(user, req.query.fields);
```

### Express Middleware

```javascript
const FieldSelector = require("../utils/field-selector");

// Apply to all routes
app.use(FieldSelector.middleware());

// Or specific routes
router.get(
  "/users/:id",
  FieldSelector.middleware({ defaultFields: ["id", "name"] }),
  async (req, res) => {
    const user = await UserService.getUser(req.params.id);
    res.json(user); // Automatically filtered
  }
);
```

### Fastify Decorator

```javascript
const FieldSelector = require("../utils/field-selector");

// Register decorator
fastify.register(FieldSelector.fastifyDecorator());

// Use in route
fastify.get("/users/:id", async (request, reply) => {
  const user = await UserService.getUser(request.params.id);
  return reply.selectFields(user, request.query.fields);
});
```

### Nested Fields

```javascript
// GET /api/users/123?fields=id,name,address.city,address.country
const selected = FieldSelector.selectFields(
  user,
  "id,name,address.city,address.country",
  {
    includeNested: true,
  }
);
```

### With Relationships

```javascript
// GET /api/users/123?fields=id,name,orders(id,total,status)
const parsed = FieldSelector.parseWithRelationships(req.query.fields);
// {
//   fields: ['id', 'name'],
//   relationships: {
//     orders: ['id', 'total', 'status']
//   }
// }
```

---

## Complete Examples

### Example 1: User Dashboard Aggregator

Located at `src/examples/user-dashboard.controller.js`:

```javascript
const ParallelAggregator = require("../utils/aggregator");
const CacheManager = require("../utils/cache-manager");
const ResponseTransformer = require("../utils/response-transformer");

class UserDashboardController {
  constructor() {
    this.aggregator = new ParallelAggregator({ timeout: 5000 });
    this.cache = new CacheManager({ ttl: 60, prefix: "dashboard:" });
  }

  async getDashboard(req, res, next) {
    try {
      const userId = req.params.userId;
      const cacheKey = `user:${userId}`;

      // Check cache
      const cached = await this.cache.get(cacheKey);
      if (cached) return res.json(cached);

      // Fetch from multiple services
      const results = await this.aggregator.fetchParallel([
        { name: "profile", fetch: () => UserService.getProfile(userId) },
        {
          name: "orders",
          fetch: () => OrdersService.getRecent(userId, { limit: 5 }),
        },
        { name: "preferences", fetch: () => PreferencesService.get(userId) },
      ]);

      // Transform and combine
      const dashboard = {
        user: ResponseTransformer.pick(
          results.find((r) => r.service === "profile")?.data,
          ["id", "name", "email", "avatar"]
        ),
        recentOrders: results.find((r) => r.service === "orders")?.data || [],
        settings: results.find((r) => r.service === "preferences")?.data || {},
      };

      // Cache result
      await this.cache.set(cacheKey, dashboard, 60);
      res.json(dashboard);
    } catch (error) {
      next(error);
    }
  }
}
```

### Example 2: E-commerce Catalog with Batching

Located at `src/examples/ecommerce-catalog.controller.js`:

```javascript
const ParallelAggregator = require("../utils/aggregator");
const CacheManager = require("../utils/cache-manager");
const CircuitBreaker = require("../utils/circuit-breaker");
const RequestBatcher = require("../utils/request-batcher");
const ResponseTransformer = require("../utils/response-transformer");

class CatalogController {
  constructor() {
    this.aggregator = new ParallelAggregator({ timeout: 3000 });
    this.cache = new CacheManager({ ttl: 300, prefix: "catalog:" });
    this.batcher = new RequestBatcher({ batchWindow: 10, maxBatchSize: 50 });
    this.inventoryBreaker = new CircuitBreaker({ failureThreshold: 5 });
  }

  async getCatalog(req, res, next) {
    try {
      const { category, page = 1, limit = 20 } = req.query;
      const cacheKey = `catalog:${category}:${page}:${limit}`;

      // Check cache
      const cached = await this.cache.get(cacheKey);
      if (cached) return res.json(cached);

      // Fetch products and inventory in parallel
      const results = await this.aggregator.fetchParallel([
        {
          name: "products",
          fetch: () => ProductsService.getByCategory(category, { page, limit }),
        },
        {
          name: "inventory",
          fetch: () =>
            this.inventoryBreaker.execute(
              () => InventoryService.getBulk(category),
              () => ({ items: [] }) // Fallback if service down
            ),
        },
      ]);

      const products =
        results.find((r) => r.service === "products")?.data?.items || [];
      const inventory =
        results.find((r) => r.service === "inventory")?.data?.items || [];
      const inventoryMap = new Map(inventory.map((i) => [i.productId, i]));

      // Enrich products with batched ratings
      const enriched = await Promise.all(
        products.map(async (product) => {
          const inv = inventoryMap.get(product.id);
          const rating = await this.batcher
            .batch(
              "ratings",
              (ids) => ReviewsService.getBulkRatings(ids),
              product.id
            )
            .catch(() => null);

          return {
            ...ResponseTransformer.pick(product, [
              "id",
              "name",
              "price",
              "images",
            ]),
            inStock: inv?.quantity > 0 || false,
            rating: rating
              ? { score: rating.rating, count: rating.count }
              : null,
          };
        })
      );

      const catalog = { products: enriched, pagination: { page, limit } };
      await this.cache.set(cacheKey, catalog, 300);
      res.json(catalog);
    } catch (error) {
      next(error);
    }
  }
}
```

---

## Best Practices

### 1. Caching Strategy

```javascript
// ✅ Good: Cache expensive operations
const dashboard = (await cache.get(key)) || (await fetchAndCache(userId));

// ❌ Bad: Caching cheap operations
await cache.set("timestamp", Date.now()); // Unnecessary overhead
```

### 2. Circuit Breaker Placement

```javascript
// ✅ Good: Wrap unreliable external services
const recommendations = await breaker.execute(
  () => ThirdPartyAPI.getRecommendations(userId),
  () => [] // Graceful degradation
);

// ❌ Bad: Wrapping reliable database calls
await breaker.execute(() => db.query("SELECT ...")); // Unnecessary
```

### 3. Graceful Degradation

```javascript
// ✅ Good: Continue with partial data
const results = await aggregator.fetchParallel(services);
const user = results.find((r) => r.service === "user")?.data;
const orders = results.find((r) => r.service === "orders")?.data || []; // Default to empty

// ❌ Bad: Fail if any service fails
if (results.some((r) => !r.success)) {
  throw new Error("Service unavailable");
}
```

### 4. Response Transformation

```javascript
// ✅ Good: Only send needed fields
const publicUser = ResponseTransformer.pick(user, ["id", "name", "avatar"]);

// ❌ Bad: Sending everything including sensitive data
res.json(user); // Includes password hash, email, etc.
```

### 5. Request Batching

```javascript
// ✅ Good: Batch when fetching related data in loops
const enriched = await Promise.all(
  items.map((item) => batcher.batch("details", fetchDetails, item.id))
);

// ❌ Bad: Not batching N+1 queries
for (const item of items) {
  item.details = await fetchDetails(item.id); // N separate requests
}
```

---

## Performance Optimization

### Parallel vs Waterfall

```javascript
// Parallel: When services are independent (faster)
const [user, orders, preferences] = await Promise.all([
  UserService.get(id),
  OrdersService.get(id),
  PreferencesService.get(id),
]);

// Waterfall: When services depend on each other
const user = await UserService.get(id);
const permissions = await PermissionsService.get(user.roleId);
const resources = await ResourcesService.get(permissions.accessLevel);
```

### Timeout Configuration

```javascript
// Critical path: Short timeout
const aggregator = new ParallelAggregator({ timeout: 1000 });

// Background tasks: Longer timeout
const backgroundAggregator = new ParallelAggregator({ timeout: 30000 });
```

### Cache TTL Guidelines

```javascript
// User data: Medium TTL (5 minutes)
cache.set(key, userData, 300);

// Product catalog: Long TTL (15 minutes)
cache.set(key, products, 900);

// Real-time data: Short TTL (30 seconds)
cache.set(key, liveData, 30);

// Static data: Very long TTL (1 hour)
cache.set(key, staticContent, 3600);
```

### Memory Management

```javascript
// ✅ Good: Clear cache periodically
setInterval(() => batcher.clearAll(), 60000);

// ✅ Good: Limit batch size
const batcher = new RequestBatcher({ maxBatchSize: 50 });

// ❌ Bad: Unbounded cache growth
// Always set TTL on cached items
```

---

## Troubleshooting

### Redis Connection Issues

```javascript
// Check if Redis is available
if (cache.isUsingRedis()) {
  console.log('Using Redis cache');
} else {
  console.warn('Falling back to in-memory cache');
}

// Manual Redis setup
docker-compose up redis
export REDIS_URL=redis://localhost:6379
```

### Circuit Breaker Always Open

```javascript
// Check state
console.log(breaker.getState());

// Adjust thresholds
const breaker = new CircuitBreaker({
  failureThreshold: 10, // Increase tolerance
  resetTimeout: 30000, // Retry sooner
});

// Manual reset
breaker.reset();
```

### Slow Aggregation

```javascript
// Check which service is slow
const results = await aggregator.fetchParallel(services);
results.forEach((r) => {
  console.log(`${r.service}: ${r.duration}ms`);
});

// Reduce timeout
const fastAggregator = new ParallelAggregator({ timeout: 2000 });
```

---

## Additional Resources

- [Example Controllers](../internal/templates/node/common/examples/)
- [Template Source Code](../internal/templates/node/common/)
- [Testing Guide](./NODEJS_TESTING.md)
- [API Documentation](./API_REFERENCE.md)

---

## Summary

The bffgen Node.js aggregation utilities provide:

✅ **Parallel aggregation** for faster response times
✅ **Production-ready caching** with Redis
✅ **Circuit breakers** for fault tolerance
✅ **Response transformation** for clean APIs
✅ **Request batching** to prevent N+1 queries
✅ **Field selection** for flexible responses

Use these utilities to build robust, performant Backend-for-Frontend services that aggregate data from multiple microservices while maintaining reliability and performance.
