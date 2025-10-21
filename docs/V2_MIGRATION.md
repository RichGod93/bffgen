# bffgen v2.0 Migration Guide

## Overview

bffgen v2.0 introduces significant stability and developer experience improvements. This guide helps you migrate from v1.x to v2.0.

**Release Date:** October 2025  
**Version:** 2.0.0

---

## What's New in v2.0

### Core Stability Improvements

1. **Idempotent Generation**

   - Running `bffgen generate` multiple times is now safe
   - Uses markers to track generated sections
   - Add `--force` flag to force regeneration

2. **Real Proxy Implementation**

   - Go projects now use `httputil.ReverseProxy` (not "not implemented" stubs)
   - Proper error handling and request forwarding
   - Production-ready proxying

3. **Working Authentication**

   - JWT auth is fully implemented (no more TODO comments)
   - Both Express and Fastify have working middleware
   - Token validation, cookie handling, profile endpoints all functional

4. **Config Validation**

   - New `bffgen config validate` command
   - Validates syntax, structure, and business rules
   - Catches duplicate endpoints, invalid URLs, missing fields

5. **Transaction Rollback**

   - Failed operations automatically rollback
   - Backups stored in `.bffgen/backup/`
   - No more half-completed generations

6. **Node.js add-route Support**
   - `bffgen add-route` now works for both Go and Node.js
   - Updates `bffgen.config.json` correctly
   - Full validation and duplicate detection

### Developer Experience Improvements

7. **Progress Indicators**

   - Visual feedback during multi-step operations
   - Spinners, progress bars, elapsed time

8. **Colorized Diffs**

   - `--dry-run` shows git-style colorized diffs
   - Preview exactly what will change before applying

9. **Post-Init Guidance**

   - Detects missing tools (node, npm, go, docker)
   - Shows install instructions
   - Optionally runs `npm install` automatically
   - Personalized next steps

10. **Auto-Route Registration**

    - Routes automatically imported in `src/index.js`
    - No manual import statements needed
    - Uses `bffgen:begin:routes` markers

11. **Runtime Override**
    - Global `--runtime` flag to override detection
    - Stores runtime in `.bffgen/metadata.json`
    - Warns if override conflicts with detected type

### Feature Parity

12. **Go Aggregation Library**

    - New `internal/aggregation/` package
    - ParallelAggregator, CacheManager, CircuitBreaker
    - ResponseTransformer, RequestBatcher
    - Matches Node.js utility capabilities

13. **Enhanced Testing**

    - Test fixtures and integration templates
    - `.env.test` for test environments
    - Jest configuration with 70% coverage thresholds
    - Mock JWT tokens and authenticated requests

14. **Real HTTP Clients**
    - Aggregators use actual HTTP calls
    - Fallback to mock data if backend unavailable
    - Environment variable configuration

### New Commands

15. **Config Converter**

    ```bash
    bffgen convert config --from yaml --to json
    ```

16. **Add Infrastructure**

    ```bash
    bffgen add-infra --ci --docker --compose --health
    ```

---

## Breaking Changes

### 1. Config Validation is Strict

**v1.x:** Invalid configs might generate broken code  
**v2.0:** `bffgen generate` validates config first and fails fast

**Migration:** Run `bffgen config validate` to find and fix errors

### 2. Generation is Idempotent

**v1.x:** Re-running `generate` could duplicate routes  
**v2.0:** Re-running updates existing code in-place

**Migration:** Use `--force` flag to force full regeneration

### 3. File Structure Changes

**New directory:** `.bffgen/`

- `state.json` - Generation state tracking
- `metadata.json` - Project runtime information
- `backup/` - Transaction backups

**Migration:** No action needed, created automatically

### 4. Auth Code is Implemented

**v1.x:** Auth endpoints had TODO comments  
**v2.0:** Working JWT implementation out of the box

**Migration:** If you implemented custom auth, review generated code

### 5. Auto-Route Registration

**v1.x:** Routes needed manual import in index.js  
**v2.0:** Routes auto-imported using markers

**Migration:**

- New projects: Automatic
- Existing projects: Add markers manually or regenerate index.js

---

## Migration Steps

### For New Projects

No migration needed! Just use bffgen v2.0:

```bash
bffgen init my-new-bff --lang nodejs-express
```

### For Existing v1.x Projects

#### Option 1: Fresh Regeneration (Recommended)

1. **Backup your project:**

   ```bash
   cp -r my-project my-project-backup
   ```

2. **Run config validation:**

   ```bash
   cd my-project
   bffgen config validate
   ```

3. **Fix any validation errors**

4. **Regenerate with force:**

   ```bash
   bffgen generate --dry-run  # Preview changes
   bffgen generate --force    # Apply changes
   ```

5. **Review auth endpoints** (now have real implementations)

6. **Test thoroughly**

#### Option 2: Incremental Upgrade

1. **Update bffgen binary:**

   ```bash
   go install github.com/RichGod93/bffgen/cmd/bffgen@v2.0.0
   ```

2. **Add markers to index.js** (for auto-registration):

   In `src/index.js`, wrap your route imports:

   ```javascript
   // bffgen:begin:routes
   // Your route imports here
   app.use(require("./routes/users"));
   // bffgen:end:routes
   ```

3. **Run validation:**

   ```bash
   bffgen config validate
   ```

4. **Generate new routes:**

   ```bash
   bffgen generate  # Now idempotent
   ```

#### Option 3: Stay on v1.x

If you have heavily customized generated code:

```bash
# Pin to v1.2.0
go install github.com/RichGod93/bffgen/cmd/bffgen@v1.2.0
```

---

## New Global Flags

All commands now support:

```bash
--runtime string     Override runtime detection (go, nodejs-express, nodejs-fastify)
--verbose           Enable verbose output
--no-color          Disable colored output
```

## New Generate Flags

```bash
bffgen generate --check     # Check mode (show what would change)
bffgen generate --dry-run   # Dry run with diff preview
bffgen generate --force     # Force overwrite existing files
bffgen generate --verbose   # Verbose output
```

---

## Testing Your Migration

### 1. Validate Configuration

```bash
bffgen config validate
```

**Expected:** No errors

### 2. Preview Generation

```bash
bffgen generate --dry-run
```

**Expected:** See colorized diff of changes

### 3. Generate Code

```bash
bffgen generate
```

**Expected:**

- Routes auto-registered
- No "not implemented" stubs
- No TODO comments in auth

### 4. Test Server

For Go:

```bash
go run main.go
```

For Node.js:

```bash
npm install
npm run dev
```

**Expected:** Server starts without errors

### 5. Test Endpoints

```bash
# Health check
curl http://localhost:8080/health

# Auth (should work, not return "implement authentication")
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```

---

## Troubleshooting

### Issue: "No project configuration found"

**Cause:** Runtime detection failing  
**Fix:** Use `--runtime` flag:

```bash
bffgen --runtime nodejs-express generate
```

### Issue: "Duplicate endpoint error"

**Cause:** Strict validation catching duplicates  
**Fix:** Review config, remove duplicates:

```bash
bffgen config validate  # Shows exact duplicate
```

### Issue: "Routes not auto-registered"

**Cause:** Missing markers in index.js  
**Fix:** Add markers or use convert:

```bash
# Backup first
cp src/index.js src/index.js.backup

# Regenerate index.js from template
bffgen init temp --lang nodejs-express
cp temp/src/index.js src/index.js
rm -rf temp

# Then generate
bffgen generate
```

### Issue: ".bffgen directory appearing"

**Cause:** This is normal in v2.0  
**Action:** Add to `.gitignore`:

```
.bffgen/
```

---

## Rollback to v1.x

If you encounter issues:

```bash
# Restore backup
rm -rf my-project
cp -r my-project-backup my-project

# Reinstall v1.2.0
go install github.com/RichGod93/bffgen/cmd/bffgen@v1.2.0
```

---

## Getting Help

- **Documentation:** [docs/](https://github.com/RichGod93/bffgen/tree/master/docs)
- **Issues:** [GitHub Issues](https://github.com/RichGod93/bffgen/issues)
- **Discussions:** [GitHub Discussions](https://github.com/RichGod93/bffgen/discussions)

---

## Changelog Summary

**Added:**

- Config validation command
- Config converter tool
- Add-infra command
- Go aggregation library (5 utilities)
- Transaction rollback system
- Progress indicators
- Colorized diffs
- Runtime override flag
- Auto-route registration
- Test fixtures and templates

**Changed:**

- Generation is now idempotent
- Auth code is fully implemented
- Proxy implementation uses real httputil.ReverseProxy
- add-route works for both Go and Node.js
- Post-init guidance with tool detection

**Fixed:**

- No more "not implemented" stubs
- No more TODO comments in generated code
- No duplicate route generation
- Proper error handling and rollback

**Full Changelog:** [RELEASE_NOTES_v2.0.0.md](../RELEASE_NOTES_v2.0.0.md)
