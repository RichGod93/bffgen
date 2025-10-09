# Testing Checklist for v1.2.0 Release

## Pre-Release Testing Checklist

### ✅ Build & Compilation Tests

- [ ] `go build ./...` - All packages compile
- [ ] `go test ./...` - All Go tests pass
- [ ] `go vet ./...` - No linting errors
- [ ] `make build` - Binary builds successfully
- [ ] Binary runs: `./bffgen version`

### ✅ Express Project Tests

#### Project Creation

- [ ] Create Express project: `bffgen init test-express --lang nodejs-express`
- [ ] Verify directory structure created
- [ ] Verify package.json created with correct dependencies
- [ ] Check Redis in optionalDependencies
- [ ] Check new npm scripts (test:aggregation, cache:clear)

#### Utility Files

- [ ] `src/utils/aggregator.js` exists
- [ ] `src/utils/cache-manager.js` exists
- [ ] `src/utils/circuit-breaker.js` exists
- [ ] `src/utils/response-transformer.js` exists
- [ ] `src/utils/request-batcher.js` exists
- [ ] `src/utils/field-selector.js` exists
- [ ] `src/utils/logger.js` exists

#### Example Files

- [ ] `src/examples/user-dashboard.controller.js` exists
- [ ] `src/examples/ecommerce-catalog.controller.js` exists
- [ ] `src/examples/README.md` exists

#### Redis Setup

- [ ] `docker-compose.yml` exists
- [ ] `scripts/clear-cache.js` exists and is executable

#### Code Generation

- [ ] Add route: `bffgen add-route`
- [ ] Run generate: `bffgen generate`
- [ ] Controller generated in `src/controllers/`
- [ ] Service generated in `src/services/`
- [ ] Route generated in `src/routes/`

#### Controller Content Verification

- [ ] Controller imports ParallelAggregator
- [ ] Controller imports CacheManager
- [ ] Controller imports CircuitBreaker
- [ ] Controller imports ResponseTransformer
- [ ] Controller has cache-first strategy
- [ ] Controller has circuit breaker logic
- [ ] No template syntax errors

#### JavaScript Syntax

- [ ] All utility files: `node --check src/utils/*.js`
- [ ] All examples: `node --check src/examples/*.js`
- [ ] Generated controller: `node --check src/controllers/*.js`

---

### ✅ Fastify Project Tests

#### Project Creation

- [ ] Create Fastify project: `bffgen init test-fastify --lang nodejs-fastify`
- [ ] Verify directory structure created
- [ ] Verify package.json with Fastify dependencies

#### Utility Files

- [ ] All 6 utility files created in `src/utils/`
- [ ] Fastify-specific adaptations correct

#### Example Files

- [ ] Both examples created
- [ ] Examples use Fastify request/reply pattern

#### Code Generation

- [ ] Add route: `bffgen add-route`
- [ ] Run generate: `bffgen generate`
- [ ] Fastify controller generated
- [ ] Fastify service generated

---

### ✅ Go Project Regression Tests

#### Project Creation

- [ ] Create Go project: `bffgen init test-go --lang go --framework chi`
- [ ] Verify Go structure (no `src/` directory)
- [ ] Verify go.mod created
- [ ] Verify main.go created

#### No Node.js Contamination

- [ ] Go project has no `src/` directory
- [ ] Go project has no `package.json`
- [ ] Go project has no Node.js utilities
- [ ] Go template copy works (if using template option)

---

### ✅ Doctor Command Tests

- [ ] Run in Express project: `bffgen doctor`
- [ ] Run in Fastify project: `bffgen doctor`
- [ ] Run in Go project: `bffgen doctor`
- [ ] All checks pass

---

### ✅ Documentation Tests

- [ ] `docs/NODEJS_AGGREGATION.md` exists
- [ ] All code examples in documentation are syntactically correct
- [ ] Links in documentation work
- [ ] Examples can be copy-pasted and run

---

### ✅ Backward Compatibility Tests

#### Existing Projects (v1.1.0)

- [ ] Old Express projects still work
- [ ] Old Fastify projects still work
- [ ] Old Go projects still work
- [ ] `bffgen generate` works on old config files
- [ ] No breaking changes to existing commands

---

### ✅ Edge Cases

- [ ] Project with no endpoints (controllers dir empty - expected)
- [ ] Project without Redis (falls back to memory cache - expected)
- [ ] Large number of endpoints (20+)
- [ ] Special characters in service names
- [ ] Very long project paths

---

### ✅ Performance Tests

- [ ] Init command completes in < 10 seconds
- [ ] Generate command completes in < 5 seconds
- [ ] Binary size reasonable (< 50MB)
- [ ] Memory usage acceptable during generation

---

### ✅ Integration Tests

#### Redis Integration

- [ ] Start Redis: `docker-compose up redis`
- [ ] Cache manager connects to Redis
- [ ] Cache operations work (set/get/delete)
- [ ] Fallback to memory when Redis down
- [ ] Clear cache script works: `npm run cache:clear`

#### npm Scripts

- [ ] `npm run dev` works
- [ ] `npm test` works (if tests not skipped)
- [ ] `npm run lint` works
- [ ] `npm run cache:clear` works

---

## Manual Test Commands

### Quick Smoke Test

```bash
# Build
make clean && make build

# Test Express
./bffgen init test-express --lang nodejs-express
cd test-express
ls src/utils/  # Should show 7 files
ls src/examples/  # Should show 2 files
cat package.json | grep redis  # Should find redis
cd ..

# Test Fastify
./bffgen init test-fastify --lang nodejs-fastify
cd test-fastify
ls src/utils/  # Should show 7 files
cd ..

# Test Go (regression)
./bffgen init test-go --lang go
cd test-go
ls  # Should NOT have src/ directory
cd ..

# Cleanup
rm -rf test-express test-fastify test-go
```

### Full Generate Test

```bash
# Create project
./bffgen init full-test --lang nodejs-express

cd full-test

# Add a route interactively
./bffgen add-route
# Service: users
# Method: GET
# Path: /users
# ExposeAs: /api/users
# Auth: no

# Generate code
./bffgen generate

# Verify
ls src/controllers/  # Should have users.controller.js
ls src/services/     # Should have users.service.js
ls src/routes/       # Should have users.js

# Check controller content
grep "ParallelAggregator" src/controllers/users.controller.js
grep "CacheManager" src/controllers/users.controller.js
grep "CircuitBreaker" src/controllers/users.controller.js
```

---

## Pre-Release Checklist

### Code Quality

- [ ] All Go tests pass
- [ ] No linting errors
- [ ] No compiler warnings
- [ ] JavaScript syntax valid

### Documentation

- [ ] README updated
- [ ] RELEASE_NOTES complete
- [ ] CHANGELOG updated
- [ ] Migration guide provided

### Git

- [ ] All changes committed
- [ ] Branch up to date with main
- [ ] No uncommitted changes
- [ ] Version numbers updated

### Release Artifacts

- [ ] Tag created: `git tag v1.2.0`
- [ ] Release notes prepared
- [ ] Binary builds for all platforms
- [ ] Checksums generated

---

## Test Results

**Date:** ******\_******  
**Tester:** ******\_******  
**Total Tests:** **_ / _**  
**Pass Rate:** \_\_\_%

**Critical Failures:** (list any)

- None expected

**Non-Critical Issues:** (list any)

-

**Ready for Release:** [ ] YES [ ] NO

**Notes:**
