# Migration Guide: Enhanced Scaffolding

## Overview

This guide helps existing bffgen users migrate to the enhanced scaffolding system with controllers, services, tests, and documentation.

## What Changed

### Before (v1.0.1)

```
my-bff/
├── src/
│   ├── index.js
│   ├── routes/          # Only routes generated
│   ├── middleware/      # Empty
│   ├── controllers/     # Empty
│   └── utils/           # Empty
└── package.json
```

### After (v1.1.0+)

```
my-bff/
├── src/
│   ├── index.js
│   ├── routes/          # Generated routes
│   ├── controllers/     # ✨ Auto-generated controllers
│   ├── services/        # ✨ Auto-generated services + HTTP client
│   ├── middleware/      # ✨ Configurable middleware
│   ├── utils/           # ✨ Logger utility
│   └── config/          # ✨ Swagger configuration
├── tests/               # ✨ Jest infrastructure
├── docs/                # ✨ OpenAPI documentation
└── jest.config.js       # ✨ Test configuration
```

## Migration Strategies

### Strategy 1: Fresh Start (Recommended)

Best for small projects or proof-of-concepts.

```bash
# 1. Backup your current project
cp -r my-bff my-bff-backup

# 2. Create new project with enhanced features
bffgen init my-bff-new --lang nodejs-express --middleware all

# 3. Copy your bffgen.config.json
cp my-bff/bffgen.config.json my-bff-new/

# 4. Generate everything
cd my-bff-new
bffgen generate
bffgen generate-docs

# 5. Copy any custom code
# - Custom business logic from old controllers
# - Custom middleware
# - Environment variables

# 6. Test thoroughly
npm install
npm test
npm run dev
```

### Strategy 2: Incremental Migration

Best for production projects with custom code.

#### Step 1: Add New Files

```bash
# In your existing project
cd my-existing-bff

# Generate new infrastructure
mkdir -p src/services src/utils src/config tests/integration tests/unit docs
```

#### Step 2: Add HTTP Client

Create `src/services/httpClient.js`:

```bash
# Copy from a fresh bffgen project or manually create
bffgen init temp-bff --lang nodejs-express
cp temp-bff/src/services/httpClient.js src/services/
rm -rf temp-bff
```

#### Step 3: Add Logger

Create `src/utils/logger.js`:

```bash
# Copy from fresh project
bffgen init temp-bff --lang nodejs-express
cp temp-bff/src/utils/logger.js src/utils/
rm -rf temp-bff
```

#### Step 4: Update package.json

Add new dependencies:

```json
{
  "dependencies": {
    "winston": "^3.11.0",
    "morgan": "^1.10.0",
    "swagger-ui-express": "^5.0.0",
    "swagger-jsdoc": "^6.2.8"
  },
  "devDependencies": {
    "jest": "^29.7.0",
    "supertest": "^6.3.3",
    "nock": "^13.5.0"
  }
}
```

Run `npm install`

#### Step 5: Generate Services

```bash
# This will create service files for existing backends
bffgen generate
```

#### Step 6: Refactor Controllers

Update your existing controllers to use the new service layer:

**Before:**

```javascript
async getUsers(req, res) {
  const response = await fetch('http://backend/users');
  const data = await response.json();
  res.json(data);
}
```

**After:**

```javascript
const usersService = require('../services/users.service');

async getUsers(req, res, next) {
  try {
    const data = await usersService.getAll();
    res.json(data);
  } catch (error) {
    next(error);
  }
}
```

#### Step 7: Add Tests

```bash
# Copy test infrastructure
bffgen init temp-bff --lang nodejs-express
cp temp-bff/jest.config.js .
cp temp-bff/tests/setup.js tests/
cp -r temp-bff/tests/integration/* tests/integration/
rm -rf temp-bff

# Run tests
npm test
```

#### Step 8: Add Swagger

```bash
# Copy Swagger config
bffgen init temp-bff --lang nodejs-express
cp -r temp-bff/src/config/* src/config/
rm -rf temp-bff

# Generate OpenAPI spec
bffgen generate-docs

# Update src/index.js to include Swagger
# Add at top:
const { setupSwagger } = require('./config/swagger-setup');

// After middleware setup, before routes:
setupSwagger(app);
```

## Breaking Changes

### None! 🎉

The enhanced scaffolding is **100% backward compatible**:

- Existing projects continue to work
- Old commands still function
- No required migrations
- New features are opt-in

## New Command Examples

### Initialize with All Features

```bash
bffgen init my-bff \
  --lang nodejs-express \
  --middleware all \
  --controller-type both
```

### Generate Everything

```bash
cd my-bff
bffgen add-template auth
bffgen generate          # Routes + Controllers + Services
bffgen generate-docs     # OpenAPI spec
```

### Minimal Setup

```bash
bffgen init my-minimal \
  --lang nodejs-fastify \
  --middleware none \
  --skip-tests \
  --skip-docs
```

## Testing Your Migration

### Checklist

- [ ] All dependencies installed (`npm install`)
- [ ] Server starts without errors (`npm run dev`)
- [ ] Health endpoint responds (`curl http://localhost:8080/health`)
- [ ] Tests pass (`npm test`)
- [ ] Swagger UI loads (`http://localhost:8080/api-docs`)
- [ ] Routes work as before
- [ ] Controllers properly use services
- [ ] Logging works (check `logs/` directory)

### Common Issues

**Issue**: `Cannot find module './services/httpClient'`
**Solution**: Ensure `src/services/httpClient.js` exists. Copy from fresh project or run `bffgen generate`.

**Issue**: Tests fail with module not found
**Solution**: Run `npm install jest supertest nock`

**Issue**: Swagger UI shows 404
**Solution**: Ensure `src/config/swagger-setup.js` exists and is imported in `src/index.js`

**Issue**: Logger creates errors
**Solution**: Ensure `logs/` directory exists and is writable: `mkdir -p logs`

## Rollback Plan

If you encounter issues:

```bash
# 1. Stop the server
# 2. Restore from backup
mv my-bff my-bff-new
mv my-bff-backup my-bff

# 3. Or use git
git checkout -- .
git clean -fd
```

## Benefits of Migrating

- **98% faster setup** for new features
- **Separation of concerns** - easier testing and maintenance
- **Automatic retries** - more resilient to backend failures
- **Structured logging** - easier debugging in production
- **API documentation** - self-documenting APIs
- **Test infrastructure** - higher code quality

## Support

- **Issues**: [GitHub Issues](https://github.com/RichGod93/bffgen/issues)
- **Docs**: `docs/ENHANCED_SCAFFOLDING.md`
- **Examples**: `examples/` directory
- **Tests**: Run `./scripts/test_enhanced_scaffolding.sh`

## Timeline

Recommended migration timeline:

- **Week 1**: Create test project with new features
- **Week 2**: Migrate one service incrementally
- **Week 3**: Migrate remaining services
- **Week 4**: Add tests and documentation
