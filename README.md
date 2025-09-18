# bffgen

`bffgen` is a Go-based CLI tool that helps developers quickly scaffold **Backend-for-Frontend (BFF)** services.  
It enables teams to aggregate backend endpoints and expose them in a frontend-friendly way, with minimal setup.

---

## 🚀 Features (Current)

- **`init`** → scaffold a new BFF project with chi router and config file
- **`add-route`** → interactively add backend endpoints to your BFF
- **`add-template`** → use predefined templates (auth, ecommerce, content)
- **`add-aggregator`** → create data aggregation endpoints
- **`generate`** → generate Go code for routes from config
- **`postman`** → generate Postman collection for API testing
- **`dev`** → run a local BFF server with proxying

---

## 📦 Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/richgodusen/bffgen
cd bffgen

# Build the binary
go build -o bffgen ./cmd/bffgen

# Install globally (optional)
sudo mv bffgen /usr/local/bin/
```

### With Go Install

```bash
# Install directly from GitHub
go install github.com/richgodusen/bffgen/cmd/bffgen@latest
```

---

## 🔒 Security Features

bffgen generates BFFs with secure defaults:

**✅ Security Headers:**

- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Permissions-Policy: geolocation=(), microphone=(), camera=()`

**✅ CORS Configuration:**

- Restrictive origins (localhost:3000, localhost:3001 by default)
- Credentials enabled for authenticated requests
- Specific allowed headers and methods

**✅ Request Validation:**

- 10MB request size limit
- Content-Type validation for POST/PUT requests
- Request timeout (60 seconds)

**✅ Authentication & Security:**

- JWT token validation with HMAC signing
- User context injection (user_id, email)
- Public endpoint exclusions (health, auth routes)
- Environment-based JWT secret configuration

**✅ Comprehensive Logging:**

- Structured request/response logging
- Request duration tracking
- Request ID correlation
- Error logging with context

**✅ Rate Limiting:**

- Fiber: Built-in in-memory rate limiting (100 req/min)
- Chi/Echo: Redis-based rate limiting (requires Redis setup)
- Configurable via environment variables

**⚠️ Production Considerations:**

- Set JWT_SECRET environment variable
- Configure Redis for Chi/Echo rate limiting
- Set up proper logging aggregation
- Configure monitoring and alerting

---

## 🛠 Usage

### Initialize a Project

```bash
bffgen init my-bff
cd my-bff
```

This creates:

- `main.go` - Chi router server
- `bff.config.yaml` - Configuration file
- `go.mod` - Go module file
- `README.md` - Project documentation
- Directory structure for routes and templates

### Configure Backend Services

Edit `bff.config.yaml`:

```yaml
services:
  users:
    baseUrl: "http://localhost:4000/api"
    endpoints:
      - name: "getUser"
        path: "/users/:id"
        method: "GET"
        exposeAs: "/api/users/:id"
      - name: "createUser"
        path: "/users"
        method: "POST"
        exposeAs: "/api/users"
  orders:
    baseUrl: "http://localhost:5000/api"
    endpoints:
      - name: "getOrders"
        path: "/orders"
        method: "GET"
        exposeAs: "/api/orders"

settings:
  port: 8080
  timeout: 30s
  retries: 3
```

### Generate Code and Test

```bash
# Generate Go code from configuration
bffgen generate

# Generate Postman collection for testing
bffgen postman

# Run the BFF server
bffgen dev
```

Output:

```
🔧 Generating Go code from bff.config.yaml
✅ Code generation completed!
📁 Updated files:
   - main.go (with proxy routes)
   - cmd/server/main.go (server entry point)

🚀 Run 'go run main.go' to start your BFF server

📮 Generate Postman collection: bffgen postman
   This creates a ready-to-import collection for testing your BFF endpoints
```

Postman Collection Generation:

```
📮 Generating Postman collection from bff.config.yaml

🔍 Step 1: Checking for BFF configuration...
✅ Found bff.config.yaml
🔍 Step 2: Loading and validating configuration...
✅ Configuration loaded successfully
🔍 Step 3: Validating service configurations...
✅ All service configurations are valid
🔍 Step 4: Generating Postman collection...
✅ Postman collection generated successfully!
📁 Created file: bff-postman-collection.json

📋 Collection Summary:
   • auth service: 5 endpoints
   • Total: 5 endpoints across 1 services
   • BFF server port: 8080

🚀 Next Steps:
   1. Import 'bff-postman-collection.json' into Postman
   2. Start your BFF server: go run main.go
   3. Test your endpoints using the collection

💡 Pro Tips:
   • Use the 'baseUrl' variable to switch between environments
   • The collection includes a health check endpoint
   • All endpoints are pre-configured with proper headers
```

### Add Templates

```bash
# Show authentication template
bffgen add-template auth

# Show e-commerce template
bffgen add-template ecommerce

# Show content management template
bffgen add-template content
```

---

## 📂 Project Structure

```
my-bff/
├── main.go                 # Chi router server
├── bff.config.yaml         # BFF configuration
├── go.mod                  # Go module
├── README.md               # Project documentation
└── internal/
    ├── routes/             # Custom route handlers
    ├── aggregators/        # Data aggregation logic
    └── templates/          # Template definitions
```

---

## 🔧 Configuration Reference

### Service Configuration

```yaml
services:
  service-name:
    baseUrl: "http://backend-service:port/api"
    endpoints:
      - name: "endpoint-name" # Internal name
        path: "/backend/path/:id" # Backend endpoint path
        method: "GET" # HTTP method
        exposeAs: "/api/frontend" # Frontend-facing path
```

### Supported HTTP Methods

- `GET`
- `POST`
- `PUT`
- `DELETE`
- `PATCH`

### Global Settings

```yaml
settings:
  port: 8080 # BFF server port
  timeout: 30s # Request timeout
  retries: 3 # Retry attempts
```

---

## 🛤 Development Status

### ✅ Completed Features

- **Core CLI Framework** - Cobra-based command structure
- **Project Initialization** - `bffgen init` with chi router setup
- **Configuration Management** - YAML-based service configuration
- **Template System** - Pre-built templates (auth, ecommerce, content)
- **Route Management** - Interactive route addition and validation
- **Code Generation** - Automatic Go code generation from config
- **Postman Integration** - Collection generation for API testing
- **Development Server** - Local BFF server with proxy functionality
- **Error Handling** - Comprehensive validation and user-friendly error messages

### 🔄 Current Development Stage

**Phase 1 Complete** - Core BFF functionality is production-ready

- ✅ Basic CLI scaffolding
- ✅ YAML configuration with validation
- ✅ HTTP proxy functionality
- ✅ Chi router integration
- ✅ Interactive route addition
- ✅ Automatic code generation
- ✅ Template system with 3 built-in templates
- ✅ Postman collection generation
- ✅ Comprehensive error handling and user guidance

### 🚧 Next Phase (Planned)

**Phase 2 - Enhanced Features**

- 🔄 Real proxy implementation (currently placeholder)
- 🔄 Authentication middleware integration
- 🔄 Request/response transformation
- 🔄 Environment-specific configurations
- 🔄 Advanced aggregation patterns

### 🔮 Future Roadmap

**Phase 3 - Advanced Capabilities**

- 🔮 GraphQL support (schema stitching)
- 🔮 Rate limiting / caching (Redis integration)
- 🔮 Plugin system for extensibility
- 🔮 Docker integration (`bffgen dockerize`)
- 🔮 SDK generation for frontend frameworks
- 🔮 Monitoring and observability
- 🔮 Multi-environment deployment

---

## 🔴 Redis Setup (Chi/Echo Only)

For Chi and Echo frameworks, Redis is required for rate limiting:

### Installation

**macOS:**
```bash
brew install redis
brew services start redis
```

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install redis-server
sudo systemctl start redis-server
```

**Docker:**
```bash
docker run -d -p 6379:6379 redis:alpine
```

### Configuration

Set the Redis URL environment variable:
```bash
export REDIS_URL=redis://localhost:6379
```

### Verification

Test Redis connection:
```bash
redis-cli ping
# Should return: PONG
```

**Note:** Fiber framework includes built-in rate limiting and doesn't require Redis.

---

## 🔐 JWT Authentication Setup

### Environment Configuration

Set your JWT secret:
```bash
export JWT_SECRET=your-super-secure-secret-key-change-in-production
```

### Token Generation

Generate JWT tokens in your authentication service:
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

Include the token in requests:
```bash
curl -H "Authorization: Bearer <your-jwt-token>" http://localhost:8080/api/protected
```

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Setup

```bash
# Clone and setup
git clone https://github.com/richgodusen/bffgen
cd bffgen

# Install dependencies
go mod tidy

# Build and test
go build -o bffgen ./cmd/bffgen
./bffgen --help
```

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- [Chi Router](https://github.com/go-chi/chi) - Lightweight HTTP router
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [YAML](https://github.com/go-yaml/yaml) - YAML parsing
- Inspired by the Backend-for-Frontend pattern by [Martin Fowler](https://martinfowler.com/articles/bff.html)
