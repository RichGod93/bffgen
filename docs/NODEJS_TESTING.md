# Node.js Runtime Support - Testing Guide

## ✅ Implementation Complete!

The bffgen CLI now successfully supports generating BFF projects for:

- **Go** (chi, echo, fiber)
- **Node.js with Express**
- **Node.js with Fastify**

## Test Results

### ✅ Express.js Test - PASSED

```bash
./bffgen init test-express --lang nodejs-express
cd test-express
npm install
node index.js
```

**Result:**

- ✅ `package.json` generated with correct dependencies
- ✅ `index.js` generated with full Express.js server
- ✅ Server starts successfully on port 8080
- ✅ Health check endpoint responds: `{"status":"healthy","timestamp":"..."}`
- ✅ Includes: CORS, helmet, rate-limiting, security headers, error handling

### ✅ Fastify Test - PASSED

```bash
./bffgen init test-fastify --lang nodejs-fastify
cd test-fastify
npm install
node index.js
```

**Result:**

- ✅ `package.json` generated with Fastify v4 dependencies
- ✅ `index.js` generated with full Fastify server
- ✅ Server starts successfully on port 8080
- ✅ Health check endpoint responds: `{"status":"healthy","timestamp":"..."}`
- ✅ Includes: CORS, helmet, rate-limiting, security headers, error handling, structured logging

## Quick Start Examples

### 1. Generate Express BFF

```bash
./bffgen init my-express-bff --lang nodejs-express
```

### 2. Generate Fastify BFF

```bash
./bffgen init my-fastify-bff --lang nodejs-fastify
```

### 3. Generate Go BFF (still works!)

```bash
./bffgen init my-go-bff --lang go --framework chi
```

## Features Implemented

### Express.js Template Includes:

- ✅ Express v4.18.2
- ✅ CORS configuration
- ✅ Helmet security headers
- ✅ Express rate limiting
- ✅ JWT authentication placeholders
- ✅ Cookie parser
- ✅ Request validation
- ✅ Error handling middleware
- ✅ 404 handler
- ✅ Health check endpoint
- ✅ TODO comments for adding routes

### Fastify Template Includes:

- ✅ Fastify v4.28.1
- ✅ @fastify/cors
- ✅ @fastify/helmet
- ✅ @fastify/rate-limit
- ✅ @fastify/jwt
- ✅ @fastify/cookie
- ✅ Structured logging (pino)
- ✅ Async/await plugin registration
- ✅ Error handling
- ✅ 404 handler
- ✅ Health check endpoint
- ✅ TODO comments for adding routes

## Generated Project Structure

### Express Project

```
my-express-bff/
├── package.json          # Dependencies & scripts
├── index.js              # Main Express server (90+ lines)
├── bff.config.yaml       # BFF configuration
├── README.md             # Project documentation
├── controllers/          # API controllers (empty)
├── middleware/           # Custom middleware (empty)
├── routes/               # Route handlers (empty)
└── utils/                # Utility functions (empty)
```

### Fastify Project

```
my-fastify-bff/
├── package.json          # Dependencies & scripts
├── index.js              # Main Fastify server (90+ lines)
├── bff.config.yaml       # BFF configuration
├── README.md             # Project documentation
├── controllers/          # API controllers (empty)
├── middleware/           # Custom middleware (empty)
├── routes/               # Route handlers (empty)
└── utils/                # Utility functions (empty)
```

## Running the Servers

### Express

```bash
cd my-express-bff
npm install
npm start          # Production
# or
npm run dev        # Development with nodemon
```

### Fastify

```bash
cd my-fastify-bff
npm install
npm start          # Production
# or
npm run dev        # Development with nodemon
```

## Testing the Health Endpoint

```bash
curl http://localhost:8080/health
```

**Expected Response:**

```json
{ "status": "healthy", "timestamp": "2025-10-03T01:06:22.606Z" }
```

## Configuration Flags

The CLI now supports:

```bash
--lang, -l          # Specify language/runtime
--runtime, -r       # Alias for --lang
--framework, -f     # Specify framework (overrides default)
```

### Valid Language Options:

- `go` (default framework: chi)
- `nodejs-express`
- `nodejs-fastify`

### Valid Framework Options:

- For Go: `chi`, `echo`, `fiber`
- For Node.js: `express`, `fastify`

## Example Commands

```bash
# Express with flags (non-interactive)
./bffgen init my-bff --lang nodejs-express

# Fastify with flags
./bffgen init my-bff --lang nodejs-fastify

# Go with specific framework
./bffgen init my-bff --lang go --framework fiber

# Using runtime alias
./bffgen init my-bff --runtime nodejs-express
```

## Dependencies

### Express Dependencies

```json
{
  "express": "^4.18.2",
  "cors": "^2.8.5",
  "helmet": "^7.1.0",
  "express-rate-limit": "^7.1.5",
  "jsonwebtoken": "^9.0.2",
  "cookie-parser": "^1.4.6",
  "express-validator": "^7.0.1"
}
```

### Fastify Dependencies

```json
{
  "@fastify/cors": "^8.5.0",
  "@fastify/helmet": "^11.1.1",
  "@fastify/rate-limit": "^9.1.0",
  "fastify": "^4.28.1",
  "@fastify/jwt": "^7.2.4",
  "@fastify/cookie": "^9.3.1"
}
```

## Code Quality

- ✅ All builds pass: `go build ./cmd/bffgen`
- ✅ All tests pass: `go test ./cmd/bffgen/commands`
- ✅ Express server starts and responds
- ✅ Fastify server starts and responds
- ✅ Generated code follows best practices
- ✅ Security headers included
- ✅ Rate limiting configured
- ✅ Error handling implemented

## Architecture Improvements

The refactoring reduced `init.go` from **1856 lines** to **90 lines**:

- `init.go` (90 lines) - CLI command definition
- `init_backend.go` (454 lines) - Backend service configuration
- `init_helpers.go` (419 lines) - File generation & Node.js templates
- **Total: 963 lines (48% reduction)**

## Next Steps

To extend the Node.js support:

1. **Add route generation** - Implement `bffgen add-route` for Node.js
2. **Add template support** - Port YAML templates to Express/Fastify
3. **Add middleware** - Generate authentication middleware
4. **Add testing** - Generate Jest test files
5. **Add Docker** - Generate Dockerfiles for Node.js

## Cleanup

To remove test projects:

```bash
rm -rf test-*-bff test-simple-express test-fastify
```

---

**Status: ✅ FULLY FUNCTIONAL**

Both Express.js and Fastify BFF generation work end-to-end, including:

- Project scaffolding
- Dependency management
- Server generation
- Health check endpoints
- Security middleware
- Error handling
