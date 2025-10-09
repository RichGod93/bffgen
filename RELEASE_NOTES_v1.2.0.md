# Release Notes - bffgen v1.2.0

## 🎉 Major Feature Release: Node.js Aggregation Utilities

**Release Date:** October 2025  
**Type:** Minor Version (Backward Compatible)  
**Previous Version:** v1.1.0

---

## 🚀 What's New

### Node.js BFF Aggregation Suite

This release brings Node.js BFF projects to feature parity with Go implementations by adding comprehensive aggregation utilities and production-ready patterns.

#### 6 New Core Utilities

All Node.js projects now include production-ready utilities in `src/utils/`:

1. **`aggregator.js`** - Parallel & Waterfall Service Execution

   - Execute multiple service calls simultaneously
   - Configurable timeouts
   - Graceful error handling
   - Fail-fast or continue-on-error modes

2. **`cache-manager.js`** - Redis-Backed Caching

   - Redis support with automatic in-memory fallback
   - TTL-based expiration
   - Key prefixing for organization
   - Graceful degradation when Redis unavailable

3. **`circuit-breaker.js`** - Fault Tolerance

   - Three states: CLOSED, OPEN, HALF_OPEN
   - Configurable failure threshold
   - Automatic recovery testing
   - Fallback function support

4. **`response-transformer.js`** - Data Transformation

   - Field selection (pick/omit)
   - Object flattening and merging
   - Computed fields
   - Sensitive data sanitization

5. **`request-batcher.js`** - N+1 Query Prevention

   - Automatic request batching
   - Configurable batch window
   - Promise-based resolution
   - Prevents database/API hammering

6. **`field-selector.js`** - GraphQL-like Field Selection
   - Query parameter-based field filtering
   - Nested field support
   - Express middleware & Fastify decorator
   - Reduces payload sizes

---

### Enhanced Controller Templates

**Before:**

```javascript
// TODO: Add custom business logic, data transformation, or aggregation
// TODO: Add caching logic if needed
// TODO: Transform or aggregate data here
```

**After:**

```javascript
const aggregator = new ParallelAggregator({ timeout: 5000 });
const cache = new CacheManager({ ttl: 300, prefix: "service:" });
const breaker = new CircuitBreaker({ failureThreshold: 5 });

// Actual working implementations with:
// - Cache-first strategy
// - Circuit breaker protection
// - Response transformation
// - Parallel service calls
```

---

### Complete Working Examples

Two fully-functional example controllers in `src/examples/`:

1. **User Dashboard Controller** (`user-dashboard.controller.js`)

   - Aggregates profile + orders + preferences
   - Demonstrates parallel fetching
   - Shows cache strategy
   - Graceful degradation example

2. **E-commerce Catalog Controller** (`ecommerce-catalog.controller.js`)
   - Products + inventory + ratings aggregation
   - Circuit breaker for inventory service
   - Request batching for ratings (prevents N+1)
   - Advanced caching patterns

---

### Redis & Docker Support

#### Docker Compose Configuration

Every Node.js project now includes `docker-compose.yml`:

- Redis 7-alpine with persistence
- Optional Redis Commander UI
- Health checks
- Production-ready configuration

#### Cache Management Script

New `scripts/clear-cache.js`:

- Clear Redis cache via CLI
- Check Redis availability
- Provides setup instructions

---

### Updated package.json

#### New Dependencies

```json
"optionalDependencies": {
  "redis": "^4.6.11"
}
```

#### New Scripts

```json
"test:aggregation": "jest --testPathPattern=aggregation",
"test:utils": "jest --testPathPattern=utils",
"cache:clear": "node scripts/clear-cache.js"
```

---

## 📚 Documentation

### New Documentation

- **`docs/NODEJS_AGGREGATION.md`** - 900+ lines comprehensive guide
  - Complete API reference for all utilities
  - 150+ code examples
  - Best practices
  - Performance optimization tips
  - Troubleshooting guide

### Updated Documentation

- **`NODEJS_BFF_ENHANCEMENT_SUMMARY.md`** - Implementation summary
- **`README.md`** - Updated feature list
- **`src/examples/README.md`** - Usage guide for examples

---

## 🔧 Technical Changes

### Files Added (13)

- `internal/templates/node/common/aggregator.js.tmpl`
- `internal/templates/node/common/cache-manager.js.tmpl`
- `internal/templates/node/common/circuit-breaker.js.tmpl`
- `internal/templates/node/common/response-transformer.js.tmpl`
- `internal/templates/node/common/request-batcher.js.tmpl`
- `internal/templates/node/common/field-selector.js.tmpl`
- `internal/templates/node/common/examples/user-dashboard.controller.js.tmpl`
- `internal/templates/node/common/examples/ecommerce-catalog.controller.js.tmpl`
- `internal/templates/node/common/examples/README.md`
- `internal/templates/node/common/docker-compose.yml.tmpl`
- `internal/templates/node/common/scripts/clear-cache.js.tmpl`
- `docs/NODEJS_AGGREGATION.md`
- `scripts/test_nodejs_aggregation.sh`

### Files Modified (8)

- `internal/templates/node/express/controller-aggregator.js.tmpl`
- `internal/templates/node/fastify/controller-aggregator.js.tmpl`
- `internal/templates/node/express/package.json.tmpl`
- `internal/templates/node/fastify/package.json.tmpl`
- `internal/templates/loader.go`
- `internal/templates/embedded.go`
- `cmd/bffgen/commands/init_helpers.go`
- `cmd/bffgen/commands/init_backend.go`

---

## ✅ Feature Comparison

| Feature                 | v1.1.0 (Before) | v1.2.0 (After)         | Status       |
| ----------------------- | --------------- | ---------------------- | ------------ |
| Parallel Service Calls  | ❌ Manual       | ✅ ParallelAggregator  | **NEW**      |
| Production Caching      | ⚠️ Basic Map    | ✅ Redis + Fallback    | **ENHANCED** |
| Circuit Breakers        | ❌ None         | ✅ Full Implementation | **NEW**      |
| Request Batching        | ❌ None         | ✅ RequestBatcher      | **NEW**      |
| Response Transformation | ⚠️ TODOs        | ✅ Full Library        | **ENHANCED** |
| Field Selection         | ❌ None         | ✅ FieldSelector       | **NEW**      |
| Working Examples        | ❌ None         | ✅ 2 Complete Examples | **NEW**      |
| Redis Setup             | ❌ Manual       | ✅ docker-compose.yml  | **NEW**      |

---

## 🎯 Benefits

### For Developers

- ✅ Production-ready patterns out of the box
- ✅ No need to implement aggregation from scratch
- ✅ Working examples to learn from
- ✅ Comprehensive documentation
- ✅ Faster development

### For Projects

- ✅ Better performance (parallel calls, caching)
- ✅ Higher reliability (circuit breakers, graceful degradation)
- ✅ Reduced backend load (caching, batching)
- ✅ Optimized responses (field selection, transformation)
- ✅ Production-ready from day one

---

## 📦 Installation

### Upgrade from v1.1.0

```bash
go install github.com/RichGod93/bffgen/cmd/bffgen@v1.2.0
```

### Fresh Install

```bash
go install github.com/RichGod93/bffgen/cmd/bffgen@latest
```

### From Source

```bash
git clone https://github.com/RichGod93/bffgen
cd bffgen
git checkout v1.2.0
make install VERSION=v1.2.0
```

---

## 🚦 Migration Guide

### For Existing Projects (v1.1.0 → v1.2.0)

**No breaking changes!** All existing projects continue to work.

To add new utilities to existing projects:

1. **Copy utility files manually:**

```bash
# From a newly generated project
cp new-project/src/utils/*.js existing-project/src/utils/
cp -r new-project/src/examples existing-project/src/
```

2. **Update package.json:**

```json
{
  "optionalDependencies": {
    "redis": "^4.6.11"
  },
  "scripts": {
    "test:aggregation": "jest --testPathPattern=aggregation",
    "cache:clear": "node scripts/clear-cache.js"
  }
}
```

3. **Add Redis setup (optional):**

```bash
cp new-project/docker-compose.yml existing-project/
cp -r new-project/scripts existing-project/
```

4. **Regenerate controllers:**

```bash
cd existing-project
bffgen generate
```

---

## 🧪 Testing

### Automated Tests

Run the comprehensive test suite:

```bash
./scripts/test_nodejs_aggregation.sh
```

### Manual Testing

```bash
# Create new Express project
bffgen init my-express-bff --lang nodejs-express

# Verify utilities created
ls my-express-bff/src/utils/
# aggregator.js cache-manager.js circuit-breaker.js ...

# Verify examples
ls my-express-bff/src/examples/
# user-dashboard.controller.js ecommerce-catalog.controller.js

# Add routes and generate controllers
cd my-express-bff
bffgen add-route
bffgen generate

# Check controllers
ls src/controllers/
```

---

## 🐛 Bug Fixes

### Template Loading

- **Fixed:** Examples and scripts now properly embedded in binary
- **Fixed:** Go template files no longer copied to Node.js projects
- **Fixed:** `HandlerNamePascal` template variable error in controllers

### Embed Directive

- **Changed:** `//go:embed node/**/*.tmpl` → `//go:embed node/**/*`
- **Impact:** All files in template directories now embedded correctly

---

## ⚠️ Known Limitations

1. **Redis is optional** - Falls back to in-memory cache
2. **Examples use mock data** - Replace with actual service calls
3. **Streaming aggregation** not yet implemented (planned for v1.3.0)
4. **TypeScript support** postponed (may come in v1.3.0 or v2.0.0)

---

## 📖 Documentation

### New Guides

- [Node.js Aggregation Guide](docs/NODEJS_AGGREGATION.md) - Complete reference
- [Example Controllers Guide](src/examples/README.md) - How to use examples

### Updated Guides

- [Architecture Overview](docs/ARCHITECTURE.md)
- [Quick Reference](docs/QUICK_REFERENCE.md)

---

## 🔮 Future Roadmap

### Planned for v1.3.0

- Streaming aggregation for large datasets
- GraphQL integration
- Metrics collection (Prometheus/StatsD)
- Advanced retry strategies
- Load shedding & adaptive concurrency

### Considering for v2.0.0

- Full TypeScript support
- Breaking changes for better API design
- Plugin system for custom aggregators
- Multi-region support

---

## 👥 Contributors

Thanks to all contributors who made this release possible!

---

## 📝 Changelog

### Added

- Parallel request aggregator utility
- Redis cache manager with fallback
- Circuit breaker for fault tolerance
- Response transformation utilities
- Request batcher for N+1 prevention
- Field selector for GraphQL-like filtering
- User dashboard example controller
- E-commerce catalog example controller
- Docker Compose configuration for Redis
- Cache clear utility script
- Comprehensive aggregation documentation

### Changed

- Enhanced Express aggregator controller template
- Enhanced Fastify aggregator controller template
- Updated package.json templates with Redis
- Improved template loading logic

### Fixed

- Template embedding for examples and scripts
- Go templates incorrectly copied to Node.js projects
- Controller template variable scope issues
- Embed directive now includes all template files

---

## 🙏 Acknowledgments

Special thanks to the BFF pattern community and feedback from early adopters.

---

**Full Changelog:** [v1.1.0...v1.2.0](https://github.com/RichGod93/bffgen/compare/v1.1.0...v1.2.0)  
**Documentation:** [docs/](docs/)  
**Examples:** [examples/](examples/)

---

## Questions or Issues?

- 📖 Read the [Node.js Aggregation Guide](docs/NODEJS_AGGREGATION.md)
- 🐛 [Report an Issue](https://github.com/RichGod93/bffgen/issues)
- 💬 [Discussions](https://github.com/RichGod93/bffgen/discussions)
