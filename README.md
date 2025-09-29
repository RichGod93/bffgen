# bffgen

**Backend-for-Frontend (BFF) generator** - Scaffold secure, production-ready BFF services in Go with enhanced backend architecture support, JWT auth, rate limiting, and comprehensive logging.

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Latest Release](https://img.shields.io/badge/Latest-v1.0.1-brightgreen.svg)](https://github.com/RichGod93/bffgen/releases/v1.0.1)

---

## ⚡ Quick Start

```bash
# Install latest version with enhanced backend architecture support
go install github.com/RichGod93/bffgen/cmd/bffgen@v1.0.1

# Create BFF with your preferred backend architecture
bffgen init my-bff
# Choose: 1) Microservices, 2) Monolithic, 3) Hybrid
cd my-bff

# Start backend services (configurable URLs from init)
# Then run the BFF server
go run main.go
```

**Example Output:**

```
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

| Command        | Description                          |
| -------------- | ------------------------------------ |
| `init`         | Scaffold new BFF project             |
| `add-route`    | Add backend endpoint interactively   |
| `add-template` | Add auth/ecommerce/content templates |
| `generate`     | Generate Go code from config         |
| `postman`      | Create Postman collection            |
| `dev`          | Run development server               |
| `config`       | Manage global configuration          |

---

## ✨ Features

### 🏗️ **Enhanced Backend Architecture Support**

- **Microservices**: Different ports/URLs for each service
- **Monolithic**: Single port/URL for all services
- **Hybrid**: Services on same port with different paths
- **Smart Configuration**: Auto-generates tailored bff.config.yaml
- **Intelligent Defaults**: Smart port numbering and URL suggestions

### 🔒 Security Features

- **JWT Authentication** - Token validation with user context injection
- **Rate Limiting** - Fiber built-in, Chi/Echo with Redis
- **Security Headers** - XSS, CSRF, Content-Type protection
- **CORS Configuration** - Restrictive origins, credentials support
- **Request Validation** - Size limits, content-type validation

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

```
my-bff/
├── main.go                 # Generated server with routes
├── bff.config.yaml         # Service configuration
├── go.mod                  # Dependencies
├── README.md               # Project docs
└── internal/
    ├── routes/             # Route definitions
    ├── aggregators/        # Data aggregation
    └── templates/          # Template files
```

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
