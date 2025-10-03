# bffgen

**Backend-for-Frontend (BFF) generator** - Scaffold secure, production-ready BFF services in **Go**, **Node.js (Express)**, or **Node.js (Fastify)** with JWT auth, rate limiting, CORS, and comprehensive logging.

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![Node Version](https://img.shields.io/badge/Node-18+-green.svg)](https://nodejs.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Latest Release](https://img.shields.io/badge/Latest-v1.0.1-brightgreen.svg)](https://github.com/RichGod93/bffgen/releases/v1.0.1)

---

## ⚡ Quick Start

### Go BFF

```bash
# Install
go install github.com/RichGod93/bffgen/cmd/bffgen@v1.0.1

# Create Go BFF (Chi, Echo, or Fiber)
bffgen init my-go-bff --lang go --framework chi
cd my-go-bff && go run main.go
```

### Node.js BFF (Express)

```bash
# Create Express BFF
bffgen init my-express-bff --lang nodejs-express
cd my-express-bff
npm install && npm run dev
```

### Node.js BFF (Fastify)

```bash
# Create Fastify BFF
bffgen init my-fastify-bff --lang nodejs-fastify
cd my-fastify-bff
npm install && npm run dev
```

**Example Output:**

```text
✅ BFF project 'my-bff' initialized successfully!

📋 Backend Configuration Summary:
   Architecture: Monolithic
   - Backend: http://localhost:3000/api
   - Services: users, products, orders, cart, auth

🔧 Setup Instructions:
   1. Start your monolithic backend: http://localhost:3000/api
   2. Run the BFF server: cd my-bff && go run main.go
   3. Test endpoints: curl http://localhost:8080/health

📁 Navigate to the project: cd my-bff
🚀 Start development server: go run main.go

🔐 Secure Authentication Setup:
   1. Set encryption key: export ENCRYPTION_KEY=<key>
   2. Set JWT secret: export JWT_SECRET=<key>
   3. Features: Encrypted JWT tokens, secure sessions, CSRF protection
   4. Auth endpoints: /api/auth/login, /api/auth/refresh, /api/auth/logout
```

---

## 🛠️ Commands

| Command        | Description                                                     | Go  | Node.js |
| -------------- | --------------------------------------------------------------- | --- | ------- |
| `init`         | Scaffold new BFF project (Go/Express/Fastify)                   | ✅  | ✅      |
| `add-route`    | Add backend endpoint interactively                              | ✅  | ✅      |
| `add-template` | Add auth/ecommerce/content templates                            | ✅  | ✅      |
| `generate`     | Generate routes from config                                     | ✅  | ✅      |
| `postman`      | Create Postman collection                                       | ✅  | ✅      |
| `dev`          | Run development server (Go only, use `npm run dev` for Node.js) | ✅  | -       |
| `config`       | Manage global configuration                                     | ✅  | ✅      |

---

## ✨ Features

### 🌐 **Multi-Runtime Support**

- **Go (Chi/Echo/Fiber)** - High-performance, compiled servers
- **Node.js Express** - Popular, flexible web framework
- **Node.js Fastify** - Fast, schema-based framework
- **Template-Based Generation** - Embedded templates for consistency
- **Auto-Detection** - Commands detect project type automatically

### 🏗️ **Enhanced Backend Architecture Support**

- **Microservices**: Different ports/URLs for each service
- **Monolithic**: Single port/URL for all services
- **Hybrid**: Services on same port with different paths
- **Smart Configuration**: Auto-generates bff.config.yaml or bffgen.config.json
- **Intelligent Defaults**: Smart port numbering and URL suggestions

### 🔒 Security Features

- **JWT Authentication** - Token validation with user context injection
- **Rate Limiting** - Built-in for all runtimes (Redis optional)
- **Security Headers** - Helmet, CSP, HSTS, XSS protection
- **CORS Configuration** - Restrictive origins, credentials support
- **Request Validation** - Size limits, content-type validation

### 🎨 **Developer Experience**

- **Interactive CLI** - Guided project setup with prompts
- **Template System** - Pre-built templates (auth, ecommerce, content)
- **Code Generation** - Auto-generate routes from config
- **Hot Reload** - Development mode with auto-restart (nodemon for Node.js)
- **Professional Structure** - `src/` directory, tests, middleware
- **Comprehensive Scripts** - `npm run dev`, `npm test`, `npm run lint`

---

## 📦 Installation

**Quick Install (Latest v1.0.1):**

```bash
go install github.com/RichGod93/bffgen/cmd/bffgen@v1.0.1
```

**Latest Stable:**

```bash
go install github.com/RichGod93/bffgen/cmd/bffgen@latest
```

**From Source:**

```bash
git clone https://github.com/RichGod93/bffgen
cd bffgen && go build -o bffgen ./cmd/bffgen
sudo mv bffgen /usr/local/bin/
```

---

## 🚀 Usage Examples

### Initialize Project with Backend Architecture

```bash
bffgen init my-bff
✔ Which framework? (chi/echo/fiber) [chi]: fiber
✔ Frontend URLs (comma-separated) [localhost:3000,localhost:3001]: localhost:5173
✔ What's your backend architecture?
  1) Microservices (different ports/URLs)
  2) Monolithic (single port/URL)
  3) Hybrid (some services on same port)
✔ Select option (1-3) [1]: 2
✔ Backend base URL (e.g., 'http://localhost:3000/api'): http://localhost:3000/api
✔ Configure routes now or later?
  1) Define manually
  2) Use a template
  3) Skip for now
✔ Select option (1-3) [3]: 3
```

### Example: Microservices Architecture

```bash
bffgen init my-microservices-bff
✔ Which framework? (chi/echo/fiber) [chi]: chi
✔ Frontend URLs (comma-separated) [localhost:3000,localhost:3001]: localhost:5173
✔ What's your backend architecture?
  1) Microservices (different ports/URLs)
  2) Monolithic (single port/URL)
  3) Hybrid (some services on same port)
✔ Select option (1-3) [1]: 1
🔧 Configuring Microservices Backend
✔ Service name (e.g., 'users', 'products', 'orders'): users
✔ Base URL for users (e.g., 'http://localhost:4000/api'): http://localhost:4000/api
✅ Added users service on http://localhost:4000/api
✔ Service name (e.g., 'users', 'products', 'orders'): products
✔ Base URL for products (e.g., 'http://localhost:4000/api'): http://localhost:5000/api
✅ Added products service on http://localhost:5000/api
✔ Service name (e.g., 'users', 'products', 'orders'):
✔ Configure routes now or later?
  1) Define manually
  2) Use a template
  3) Skip for now
✔ Select option (1-3): 3

✅ BFF project 'my-microservices-bff' initialized successfully!

📋 Backend Configuration Summary:
   Architecture: Microservices
   - users: http://localhost:4000/api
   - products: http://localhost:5000/api

🔧 Setup Instructions:
   1. Start your microservices on the configured ports:
      - users: http://localhost:4000/api
      - products: http://localhost:5000/api
   2. Run the BFF server:
      cd my-microservices-bff
      go run main.go
   3. Test the endpoints:
      curl http://localhost:8080/health

📁 Navigate to the project: cd my-microservices-bff
🚀 Start development server: go run main.go
```

### Working with Templates (Optional Routes)

```bash
# Add routes using templates
bffgen add-template auth
# 📁 Template added: internal/templates/auth.yaml

# Add manual routes
bffgen add-route
# ✔ Service name: payments
# ✔ Endpoint path: /api/payments
# ✔ HTTP method [GET]:

# Generate routes in main.go when ready
bffgen generate
# ✅ Code generation completed!
# 📁 Updated: main.go (with proxy routes)
```

### Run Development Server

```bash
# Your project structure is ready! No generation needed.
ls -la my-microservices-bff/
# bff.config.yaml  go.mod  main.go  README.md  internal/

# Start your backend services first
# Users API: http://localhost:4000/api
# Products API: http://localhost:5000/api

# Then run the BFF server
cd my-microservices-bff
go run main.go
# 🚀 BFF server starting on :8080
```

### Create Postman Collection

```bash
bffgen postman
# 📮 Generating Postman collection from bff.config.yaml
# ✅ Postman collection generated successfully!
# 📁 Created file: bff-postman-collection.json
```

---

## ⚙️ Configuration

bffgen saves your preferences in `~/.bffgen/bffgen.yaml` for re-runs:

### View Configuration

```bash
bffgen config show
```

### Set Defaults

```bash
bffgen config set framework fiber
bffgen config set cors_origins localhost:5173,myapp.com
bffgen config set jwt_secret my-super-secret-key
bffgen config set redis_url redis://localhost:6379
bffgen config set port 3000
bffgen config set route_option 2
```

### Reset Configuration

```bash
bffgen config reset
```

**Configuration File Location:** `~/.bffgen/bffgen.yaml`

---

## 🔴 Redis Setup (Chi/Echo Only)

```bash
# macOS
brew install redis && brew services start redis

# Ubuntu
sudo apt install redis-server && sudo systemctl start redis-server

# Docker
docker run -d -p 6379:6379 redis:alpine

# Verify
redis-cli ping  # Should return: PONG
```

**Note:** Fiber includes built-in rate limiting, no Redis needed.

---

## 🔐 JWT Authentication

### Environment Setup

```bash
export JWT_SECRET=your-super-secure-secret-key-change-in-production
```

### Token Generation

```go
import "github.com/golang-jwt/jwt/v5"

claims := jwt.MapClaims{
    "user_id": "123",
    "email": "user@example.com",
    "exp": time.Now().Add(time.Hour * 24).Unix(),
}

token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
```

### Usage

```bash
curl -H "Authorization: Bearer <your-jwt-token>" http://localhost:8080/api/protected
```

---

## 📂 Project Structure

### Go Project

```text
my-go-bff/
├── main.go                 # Generated server with routes
├── bff.config.yaml         # Service configuration
├── go.mod                  # Dependencies
├── README.md               # Project docs
└── internal/
    ├── routes/             # Route definitions
    ├── aggregators/        # Data aggregation
    └── templates/          # Template files
```

### Node.js Project (Express/Fastify)

```text
my-node-bff/
├── src/
│   ├── index.js            # Main server file
│   ├── middleware/         # Auth, error handling
│   ├── routes/             # Generated route files
│   ├── controllers/        # Business logic
│   ├── config/             # Configuration
│   └── utils/              # Utilities
├── tests/
│   ├── unit/               # Unit tests
│   └── integration/        # Integration tests
├── bffgen.config.json      # BFF configuration
├── package.json            # Dependencies & scripts
├── .env.example            # Environment template
├── .gitignore              # Git ignore
└── README.md               # Project docs
```

---

## 🚀 Node.js Usage Guide

### Express Workflow

```bash
# 1. Initialize Express project
bffgen init my-express-bff --lang nodejs-express

# 2. Navigate to project
cd my-express-bff

# 3. Add a template (auth, ecommerce, or content)
bffgen add-template ecommerce

# 4. Generate route files
bffgen generate

# 5. Install dependencies
npm install

# 6. Start development server
npm run dev:watch

# 7. Test endpoints
curl http://localhost:8080/health
curl http://localhost:8080/api/products
```

### Fastify Workflow

```bash
# 1. Initialize Fastify project
bffgen init my-fastify-bff --lang nodejs-fastify

# 2. Navigate to project
cd my-fastify-bff

# 3. Add auth template
bffgen add-template auth

# 4. Generate routes
bffgen generate

# 5. Install dependencies
npm install

# 6. Start development server
npm run dev

# 7. Test auth endpoints
curl http://localhost:8080/health
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

### Available npm Scripts

| Script               | Description                        |
| -------------------- | ---------------------------------- |
| `npm start`          | Start production server            |
| `npm run dev`        | Start development server (nodemon) |
| `npm run dev:watch`  | Watch mode with file watching      |
| `npm test`           | Run tests with coverage            |
| `npm run test:watch` | Watch mode for TDD                 |
| `npm run lint`       | Check code with ESLint             |
| `npm run lint:fix`   | Auto-fix linting issues            |
| `npm run format`     | Format code with Prettier          |
| `npm run validate`   | Run lint + format + tests          |

### Templates for Node.js

All templates work with both Go and Node.js projects:

| Template    | Services               | Endpoints | Description                               |
| ----------- | ---------------------- | --------- | ----------------------------------------- |
| `auth`      | auth                   | 5         | Login, register, refresh, logout, profile |
| `ecommerce` | products, orders, cart | 13        | Full e-commerce backend                   |
| `content`   | posts, comments, likes | 12        | Content management system                 |

**Usage:**

```bash
bffgen add-template auth       # Adds auth endpoints to config
bffgen generate                # Generates route files
```

### Environment Configuration

Node.js projects use `.env.example` as a template:

```bash
# Copy and configure
cp .env.example .env

# Edit with your values
vim .env
```

**Key Variables:**

- `NODE_ENV` - Environment (development/production)
- `PORT` - Server port (default: 8080)
- `JWT_SECRET` - Secret for JWT tokens
- `CORS_ORIGINS` - Allowed CORS origins
- `*_URL` - Backend service URLs

---

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

---

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- [Chi Router](https://github.com/go-chi/chi) - Lightweight HTTP router
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [JWT](https://github.com/golang-jwt/jwt) - JSON Web Tokens
- Inspired by [Backend-for-Frontend pattern](https://martinfowler.com/articles/bff.html) by Martin Fowler
