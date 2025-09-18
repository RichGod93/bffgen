# bffgen

**Backend-for-Frontend (BFF) generator** - Scaffold secure, production-ready BFF services in Go with JWT auth, rate limiting, and comprehensive logging.

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## ⚡ Quick Start

```bash
# Install
go install github.com/RichGod93/bffgen/cmd/bffgen@latest

# Create BFF
bffgen init my-bff
cd my-bff

# Add routes & generate code
bffgen add-template auth
bffgen generate

# Run server
go run main.go
```

**Output:**

```
✅ BFF project 'my-bff' initialized successfully!
📁 Navigate to the project: cd my-bff
🚀 Start development server: bffgen dev

🔴 Redis Setup Required for Rate Limiting (Chi/Echo only):
   1. Install Redis: brew install redis (macOS) or apt install redis (Ubuntu)
   2. Start Redis: redis-server
   3. Set environment: export REDIS_URL=redis://localhost:6379
   Note: Fiber includes built-in rate limiting, no Redis needed

🔐 JWT Authentication Setup:
   1. Set JWT secret: export JWT_SECRET=your-secure-secret-key
   2. Generate tokens in your auth service
   3. Include 'Authorization: Bearer <token>' header in requests
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

---

## 🔒 Security Features

- **JWT Authentication** - Token validation with user context injection
- **Rate Limiting** - Fiber built-in, Chi/Echo with Redis
- **Security Headers** - XSS, CSRF, Content-Type protection
- **CORS Configuration** - Restrictive origins, credentials support
- **Request Validation** - Size limits, content-type validation

---

## 📦 Installation

**Quick Install:**

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

### Initialize Project

```bash
bffgen init my-bff
✔ Which framework? (chi/echo/fiber) [chi]:
✔ Frontend URLs (comma-separated) [localhost:3000,localhost:3001]: localhost:5173
✔ Configure routes now or later?
   1) Define manually
   2) Use a template
   3) Skip for now
✔ Select option (1-3) [3]: 2
```

### Add Authentication Template

```bash
bffgen add-template auth
```

### Generate Code

```bash
bffgen generate
# ✅ Code generation completed!
# 📁 Updated files:
#    - main.go (with proxy routes)
#    - cmd/server/main.go (server entry point)
```

### Create Postman Collection

```bash
bffgen postman
# 📮 Generating Postman collection from bff.config.yaml
# ✅ Postman collection generated successfully!
# 📁 Created file: bff-postman-collection.json
```

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
