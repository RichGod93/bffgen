# bffgen

`bffgen` is a Go-based CLI tool that helps developers quickly scaffold **Backend-for-Frontend (BFF)** services.  
It enables teams to aggregate backend endpoints and expose them in a frontend-friendly way, with minimal setup.

---

## 🚀 Features (MVP)

- **`init`** → scaffold a new BFF project with chi router and config file.
- **`add-route`** → interactively add backend endpoints to your BFF.
- **`add-template`** → use predefined templates (auth, ecommerce, content).
- **`generate`** → generate Go code for routes from config.
- **`dev`** → run a local BFF server with proxying.

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

### Run the BFF Server

```bash
bffgen dev
```

Output:

```
🚀 BFF server starting on :8080
📋 Aggregated routes:
   GET  /api/users/:id  → http://localhost:4000/api/users/:id
   POST /api/users      → http://localhost:4000/api/users
   GET  /api/orders     → http://localhost:5000/api/orders

🌐 Server running at http://localhost:8080
💡 Health check: http://localhost:8080/health
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

## 🛤 Roadmap

### Phase 1 (Current)

- ✅ Basic CLI scaffolding
- ✅ YAML configuration
- ✅ HTTP proxy functionality
- ✅ Chi router integration

### Phase 2 (Planned)

- 🔄 Interactive route addition
- 🔄 Automatic code generation
- 🔄 Template system improvements
- 🔄 Authentication middleware

### Phase 3 (Future)

- 🔮 GraphQL support (schema stitching)
- 🔮 Rate limiting / caching (Redis integration)
- 🔮 Plugin system for extensibility
- 🔮 Docker integration (`bffgen dockerize`)
- 🔮 SDK generation for frontend frameworks

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
