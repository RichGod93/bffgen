# Memory Safety in bffgen

## Overview

Memory safety is critical for preventing crashes, data corruption, and security vulnerabilities. This document covers memory safety practices in both **bffgen itself** (the code generator) and **generated projects** (Go and Node.js BFFs).

---

## Memory Safety in bffgen (Go)

### Built-in Go Safety Features

bffgen benefits from Go's inherent memory safety:

1. **Automatic Garbage Collection**

   - No manual memory management
   - Automatic deallocation of unused objects
   - No dangling pointers

2. **Bounds Checking**

   - All array/slice accesses are bounds-checked at runtime
   - Prevents buffer overflows

3. **Type Safety**

   - Strong static typing
   - No unsafe type conversions without explicit casting

4. **No NULL Pointer Arithmetic**
   - Pointers can't be manipulated arithmetically
   - Prevents pointer-based exploits

### bffgen-Specific Safety Practices

#### 1. Safe String Operations

```go
// ✅ SAFE: Using strings.Builder for concatenation
var builder strings.Builder
for _, service := range services {
    builder.WriteString(service.Name)
}
result := builder.String()

// ❌ UNSAFE: Repeated string concatenation (inefficient, but still safe)
result := ""
for _, service := range services {
    result += service.Name  // Creates new string each time
}
```

#### 2. Safe File Operations

```go
// ✅ SAFE: Proper error handling and resource cleanup
func writeFile(path string, content []byte) error {
    file, err := os.Create(path)
    if err != nil {
        return fmt.Errorf("failed to create file: %w", err)
    }
    defer file.Close()  // Always close file

    if _, err := file.Write(content); err != nil {
        return fmt.Errorf("failed to write: %w", err)
    }
    return nil
}
```

#### 3. Safe Map Access

```go
// ✅ SAFE: Check if key exists before access
if service, exists := config.Services["users"]; exists {
    processService(service)
} else {
    return fmt.Errorf("service not found")
}

// ❌ UNSAFE: Direct access can panic if key doesn't exist
// service := config.Services["users"]  // Panic if key missing
```

#### 4. Safe Slice Operations

```go
// ✅ SAFE: Check bounds before access
if len(endpoints) > 0 {
    firstEndpoint := endpoints[0]
}

// ✅ SAFE: Use range to avoid bounds issues
for i, endpoint := range endpoints {
    fmt.Printf("Endpoint %d: %s\n", i, endpoint.Name)
}

// ❌ UNSAFE: Direct index access without bounds check
// firstEndpoint := endpoints[0]  // Panic if empty
```

#### 5. Safe Concurrency (Transaction System)

```go
// ✅ SAFE: Use mutex for concurrent access
type Transaction struct {
    operations []FileOperation
    mutex      sync.Mutex
    active     bool
}

func (t *Transaction) AddOperation(op FileOperation) {
    t.mutex.Lock()
    defer t.mutex.Unlock()
    t.operations = append(t.operations, op)
}
```

---

## Memory Safety in Generated Go Projects

### Generated Code Safety Features

#### 1. Safe HTTP Handlers

```go
// Generated handler with proper error handling
func handleUsers(w http.ResponseWriter, r *http.Request) {
    // Input validation prevents buffer overflows
    if r.ContentLength > MaxRequestSize {
        http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
        return
    }

    // Safe body reading with size limit
    body, err := io.ReadAll(io.LimitReader(r.Body, MaxRequestSize))
    if err != nil {
        http.Error(w, "Failed to read body", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    // Process safely...
}
```

#### 2. Safe Proxy Implementation

```go
// Generated proxy with memory-safe operations
func createProxyHandler(backendURL, backendPath string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // URL parsing with error handling
        target, err := url.Parse(backendURL)
        if err != nil {
            http.Error(w, "Invalid backend URL", http.StatusInternalServerError)
            return
        }

        // Create proxy with timeout to prevent memory leaks
        proxy := httputil.NewSingleHostReverseProxy(target)

        // Set timeout context
        ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
        defer cancel()

        r = r.WithContext(ctx)
        proxy.ServeHTTP(w, r)
    }
}
```

#### 3. Safe JWT Validation

```go
// Generated JWT middleware with safe operations
func authenticateJWT(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")

        // Safe string operations
        if authHeader == "" || len(authHeader) < 7 {
            http.Error(w, "Missing token", http.StatusUnauthorized)
            return
        }

        // Bounds-checked substring
        tokenString := authHeader[7:]  // Safe: already checked length

        // Parse with validation
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            // Validate signing method
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method")
            }
            return []byte(os.Getenv("JWT_SECRET")), nil
        })

        if err != nil || !token.Valid {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

---

## Memory Safety in Generated Node.js Projects

### Node.js Safety Considerations

Node.js is **memory-safe** (no manual memory management) but has different concerns:

1. **No Buffer Overflows** (in user code)
2. **Garbage Collection** (automatic)
3. **V8 Engine Protection** (bounds checking, type safety)

**However, vulnerabilities can still occur:**

- Memory leaks from event listeners
- Prototype pollution
- ReDoS (Regular Expression Denial of Service)
- Unvalidated input causing crashes

### Generated Node.js Code Safety Features

#### 1. Input Validation

```javascript
// Generated Express route with validation
const { body, validationResult } = require("express-validator");

router.post(
  "/api/users",
  [
    // Validation prevents injection and crashes
    body("email").isEmail().normalizeEmail(),
    body("password").isLength({ min: 6, max: 128 }),
    body("name").trim().isLength({ min: 1, max: 255 }),
  ],
  async (req, res) => {
    const errors = validationResult(req);
    if (!errors.isEmpty()) {
      return res.status(400).json({ errors: errors.array() });
    }

    // Safe to use validated input
    const user = await createUser(req.body);
    res.json(user);
  }
);
```

#### 2. Request Size Limits

```javascript
// Generated index.js with size limits
const express = require("express");
const app = express();

// Prevent memory exhaustion from large payloads
app.use(express.json({ limit: "10mb" }));
app.use(express.urlencoded({ extended: true, limit: "10mb" }));
```

#### 3. Safe JWT Handling

```javascript
// Generated JWT middleware
const jwt = require("jsonwebtoken");

const authenticateJWT = (req, res, next) => {
  const token =
    req.cookies.access_token ||
    (req.headers.authorization && req.headers.authorization.split(" ")[1]);

  if (!token) {
    return res.status(401).json({ error: "Authentication required" });
  }

  try {
    // Safe parsing with verification
    const decoded = jwt.verify(token, process.env.JWT_SECRET, {
      algorithms: ["HS256"], // Prevent algorithm confusion
      maxAge: "24h", // Prevent token reuse
    });

    req.user = decoded;
    next();
  } catch (error) {
    // Catch all JWT errors safely
    return res.status(403).json({ error: "Invalid or expired token" });
  }
};
```

#### 4. Safe HTTP Client

```javascript
// Generated HTTP client with safety features
const axios = require("axios");

class HTTPClient {
  constructor(config) {
    this.client = axios.create({
      timeout: config.timeout || 30000, // Prevent hanging
      maxContentLength: 50 * 1024 * 1024, // 50MB limit
      maxBodyLength: 50 * 1024 * 1024,
      validateStatus: (status) => status < 500, // Don't throw on 4xx
    });

    // Request interceptor for safety
    this.client.interceptors.request.use(
      (config) => {
        // Validate URL to prevent SSRF
        const url = new URL(config.url);
        if (url.protocol !== "http:" && url.protocol !== "https:") {
          throw new Error("Invalid protocol");
        }
        return config;
      },
      (error) => Promise.reject(error)
    );
  }

  async get(url, options = {}) {
    try {
      const response = await this.client.get(url, options);
      return response.data;
    } catch (error) {
      // Safe error handling
      if (error.code === "ECONNABORTED") {
        throw new Error("Request timeout");
      }
      throw error;
    }
  }
}
```

#### 5. Memory Leak Prevention

```javascript
// Generated graceful shutdown to prevent leaks
const gracefulShutdown = () => {
  console.log("Received shutdown signal, closing server gracefully...");

  server.close(() => {
    console.log("HTTP server closed");

    // Close database connections
    if (db) db.close();

    // Clear all timers
    clearInterval(healthCheckInterval);

    // Exit cleanly
    process.exit(0);
  });

  // Force shutdown after timeout
  setTimeout(() => {
    console.error("Forced shutdown after timeout");
    process.exit(1);
  }, 10000);
};

process.on("SIGTERM", gracefulShutdown);
process.on("SIGINT", gracefulShutdown);
```

---

## Common Vulnerabilities Prevented

### 1. Buffer Overflow

**Go:** Prevented by bounds checking  
**Node.js:** Prevented by V8 engine + size limits

### 2. Use-After-Free

**Go:** Prevented by garbage collection  
**Node.js:** Prevented by garbage collection

### 3. NULL Pointer Dereference

**Go:** Prevented by checking before access  
**Node.js:** Prevented by checking `null`/`undefined`

### 4. Integer Overflow

**Go:** Checked in critical code paths  
**Node.js:** Number type prevents most issues

### 5. Memory Leaks

**Go:** Prevented by proper context cancellation and defer  
**Node.js:** Prevented by proper event listener cleanup

---

## Best Practices for Users

### When Modifying Generated Code

#### Go Projects

```go
// ✅ DO: Check errors and use defer
func loadData(path string) (*Data, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()  // Always cleanup

    // ...
}

// ✅ DO: Use context for timeouts
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// ✅ DO: Validate all inputs
if len(input) > MaxSize {
    return fmt.Errorf("input too large")
}
```

#### Node.js Projects

```javascript
// ✅ DO: Validate all inputs
app.post(
  "/api/data",
  [body("field").isString().isLength({ max: 1000 })],
  handler
);

// ✅ DO: Use async/await with try-catch
async function fetchData() {
  try {
    const data = await api.get("/endpoint");
    return data;
  } catch (error) {
    logger.error("Fetch failed:", error);
    throw error;
  }
}

// ✅ DO: Clean up event listeners
process.on("SIGTERM", () => {
  server.close();
  // Remove all listeners
  emitter.removeAllListeners();
});
```

---

## Testing for Memory Safety

### Go Projects

```bash
# Run with race detector
go test -race ./...

# Profile memory usage
go test -memprofile=mem.prof
go tool pprof mem.prof

# Check for memory leaks
go test -run=TestLongRunning -memprofilerate=1
```

### Node.js Projects

```bash
# Run with heap profiling
node --inspect index.js

# Check for memory leaks in tests
npm test -- --detectLeaks

# Profile with clinic.js
npm install -g clinic
clinic doctor -- node index.js
```

---

## Security Considerations

### Input Validation (Both Runtimes)

```javascript
// Always validate external input
function sanitizeInput(input) {
  // Length check
  if (input.length > MAX_LENGTH) {
    throw new Error("Input too large");
  }

  // Type check
  if (typeof input !== "string") {
    throw new Error("Invalid type");
  }

  // Pattern check
  if (!/^[a-zA-Z0-9_-]+$/.test(input)) {
    throw new Error("Invalid characters");
  }

  return input;
}
```

### Environment Variables

```javascript
// ✅ Validate required env vars at startup
const requiredEnvVars = ["JWT_SECRET", "DATABASE_URL"];
for (const varName of requiredEnvVars) {
  if (!process.env[varName]) {
    throw new Error(`Missing required env var: ${varName}`);
  }
}
```

---

## Monitoring and Detection

### Metrics to Track

1. **Memory Usage Trends**

   - Heap size over time
   - Detect gradual increases (leaks)

2. **Request Processing Time**

   - Detect hangs or slowdowns

3. **Error Rates**

   - Unexpected crashes or errors

4. **Resource Limits**
   - File descriptors
   - Database connections
   - Open sockets

### Alerting Rules

```yaml
# Example Prometheus alert
- alert: MemoryLeak
  expr: process_resident_memory_bytes > 1e9
  for: 5m
  annotations:
    summary: "Possible memory leak detected"
```

---

## Conclusion

**bffgen's Approach to Memory Safety:**

1. ✅ Uses memory-safe languages (Go, Node.js)
2. ✅ Generates defensive code with validation
3. ✅ Implements proper error handling
4. ✅ Uses established libraries (no custom parsers)
5. ✅ Includes resource limits and timeouts
6. ✅ Provides cleanup and graceful shutdown

**Your Responsibility:**

- Follow best practices when modifying generated code
- Keep dependencies updated
- Monitor production systems
- Test thoroughly, including edge cases

Memory safety in bffgen is achieved through **layered defenses**: language safety, defensive coding, validation, limits, and monitoring.
