# Postman Collection Generation

## Overview

The `bffgen postman` command generates Postman collections from your BFF configuration, supporting both Go (`bff.config.yaml`) and Node.js (`bffgen.config.json`) projects.

---

## Usage

```bash
# Navigate to your BFF project directory
cd my-bff-project

# Generate Postman collection
bffgen postman
```

**Output:** `bff-postman-collection.json` in the current directory

---

## Runtime Detection

The postman command automatically detects your project type:

1. **Go Projects:** Reads from `bff.config.yaml`
2. **Node.js Projects:** Reads from `bffgen.config.json`

Detection is based on:

- Presence of `package.json` (Node.js)
- Presence of `go.mod` (Go)
- Existing config files

You can override detection:

```bash
bffgen --runtime nodejs-express postman
bffgen --runtime go postman
```

---

## What Gets Generated

The Postman collection includes:

### 1. Collection Metadata

- Name: "BFF API Collection"
- Postman schema v2.1.0
- Timestamp of generation

### 2. Variables

```json
{
  "key": "baseUrl",
  "value": "http://localhost:8080",
  "type": "string"
}
```

### 3. Service Groups

Each service becomes a folder in Postman:

- Service name as folder name
- All endpoints grouped under the service

### 4. Endpoints

For each endpoint:

- HTTP method (GET, POST, PUT, DELETE, etc.)
- URL with path parameters
- Headers (Content-Type, Accept)
- Description (shows backend mapping)

### 5. Health Check

Always includes a health check endpoint:

- Method: GET
- Path: `/health`
- Purpose: Verify BFF server status

---

## Example Generated Collection

```json
{
  "info": {
    "name": "BFF API Collection",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "users",
      "item": [
        {
          "name": "Get Users",
          "request": {
            "method": "GET",
            "url": "{{baseUrl}}/users",
            "description": "Proxy to http://localhost:3000/api/users"
          }
        }
      ]
    },
    {
      "name": "Health Check",
      "item": [
        {
          "name": "Health Check",
          "request": {
            "method": "GET",
            "url": "{{baseUrl}}/health"
          }
        }
      ]
    }
  ],
  "variable": [
    {
      "key": "baseUrl",
      "value": "http://localhost:8080"
    }
  ]
}
```

---

## Configuration Requirements

### For Go Projects (`bff.config.yaml`)

```yaml
services:
  users:
    baseUrl: "http://localhost:3000"
    endpoints:
      - name: "Get Users"
        path: "/api/users"
        method: "GET"
        exposeAs: "/users"

settings:
  port: 8080
```

**Required fields:**

- `services` - Map of service definitions
- `services.<name>.baseUrl` - Backend service URL
- `services.<name>.endpoints` - Array of endpoints
- `endpoints[].name` - Endpoint display name
- `endpoints[].path` - Backend path
- `endpoints[].method` - HTTP method
- `endpoints[].exposeAs` - BFF exposed path
- `settings.port` - BFF server port

### For Node.js Projects (`bffgen.config.json`)

```json
{
  "server": {
    "port": 8080
  },
  "backends": [
    {
      "name": "users",
      "baseUrl": "http://localhost:3000/api",
      "endpoints": [
        {
          "name": "Get Users",
          "path": "/users",
          "method": "GET",
          "exposeAs": "/api/users"
        }
      ]
    }
  ]
}
```

**Required fields:**

- `server.port` - BFF server port
- `backends` - Array of backend definitions
- `backends[].name` - Service name
- `backends[].baseUrl` - Backend service URL
- `backends[].endpoints` - Array of endpoints
- `endpoints[].name` - Endpoint display name
- `endpoints[].path` - Backend path
- `endpoints[].method` - HTTP method
- `endpoints[].exposeAs` - BFF exposed path

---

## Validation

The command validates:

1. **Config file exists**

   - Go: `bff.config.yaml`
   - Node.js: `bffgen.config.json`

2. **Services defined**

   - At least one service must exist
   - Each service must have a `baseUrl`

3. **Endpoints structure**

   - `name` - Cannot be empty
   - `path` - Cannot be empty
   - `method` - Must be valid HTTP method
   - `exposeAs` - Cannot be empty

4. **HTTP methods**

   - Valid: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS
   - Invalid methods are rejected

5. **Character validation**
   - Service names: No whitespace or control characters
   - Endpoint names: No tabs, newlines, or carriage returns

---

## Import into Postman

### Method 1: Postman App

1. Open Postman
2. Click "Import" button
3. Select "bff-postman-collection.json"
4. Collection appears in sidebar

### Method 2: Postman CLI

```bash
# Import collection
postman collection import bff-postman-collection.json

# Run collection
postman collection run bff-postman-collection.json
```

---

## Using the Collection

### 1. Start Your BFF Server

**Go:**

```bash
go run main.go
```

**Node.js:**

```bash
npm run dev
# or
npm start
```

### 2. Test Health Check

1. Open "Health Check" folder
2. Click "Health Check" request
3. Click "Send"
4. Expected: `{"status": "ok"}`

### 3. Test Endpoints

1. Navigate to service folder (e.g., "users")
2. Select an endpoint
3. Modify body/parameters if needed
4. Click "Send"

### 4. Switch Environments

The collection uses `{{baseUrl}}` variable:

**Create environments:**

Development:

```json
{
  "baseUrl": "http://localhost:8080"
}
```

Production:

```json
{
  "baseUrl": "https://api.example.com"
}
```

**Switch:**

1. Click environment dropdown
2. Select environment
3. All requests use new base URL

---

## Troubleshooting

### Error: Config file not found

**Cause:** Not in a BFF project directory

**Solution:**

```bash
cd path/to/your/bff-project
bffgen postman
```

### Error: No services configured

**Cause:** Empty or invalid config

**Solution:**

```bash
# Add a template
bffgen add-template auth

# Or add a custom route
bffgen add-route
```

### Error: Invalid HTTP method

**Cause:** Unsupported method in config

**Solution:** Use only: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS

### Warning: Service has no endpoints

**Cause:** Service defined but no endpoints array

**Solution:** Either:

1. Add endpoints to the service
2. Remove the empty service

### Generated collection is empty

**Cause:** Config has services but no endpoint arrays

**Solution:** Check config structure matches examples above

---

## Advanced Usage

### Custom Port

The collection uses the port from your config:

**Go:**

```yaml
settings:
  port: 3000 # Collection uses http://localhost:3000
```

**Node.js:**

```json
{
  "server": {
    "port": 3000
  }
}
```

### Path Parameters

Path parameters are automatically handled:

**Config:**

```yaml
endpoints:
  - name: "Get User by ID"
    path: "/users/{id}"
    method: "GET"
    exposeAs: "/users/{id}"
```

**Generated URL:**

```
{{baseUrl}}/users/:id
```

**In Postman:** Replace `:id` with actual value

### Authentication

For endpoints requiring authentication:

1. Add Authorization header
2. Or use Pre-request Script:

```javascript
// In collection settings > Pre-request Script
pm.request.headers.add({
  key: "Authorization",
  value: "Bearer " + pm.environment.get("token"),
});
```

---

## Integration with CI/CD

### Newman (Postman CLI)

```bash
# Install Newman
npm install -g newman

# Run collection
newman run bff-postman-collection.json \
  --environment production.json \
  --reporters cli,json \
  --reporter-json-export results.json
```

### GitHub Actions

```yaml
name: API Tests
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Generate Postman collection
        run: bffgen postman

      - name: Start BFF server
        run: |
          go run main.go &
          sleep 5

      - name: Run tests
        run: newman run bff-postman-collection.json
```

---

## Comparison: Go vs Node.js

| Feature          | Go                | Node.js                |
| ---------------- | ----------------- | ---------------------- |
| Config File      | `bff.config.yaml` | `bffgen.config.json`   |
| Format           | YAML              | JSON                   |
| Service Key      | `services.<name>` | `backends[].name`      |
| Port Location    | `settings.port`   | `server.port`          |
| Detection        | `go.mod` present  | `package.json` present |
| Generated Output | Identical         | Identical              |

**The generated Postman collection is identical regardless of runtime.**

---

## Best Practices

1. **Regenerate after changes**

   ```bash
   bffgen add-route
   bffgen postman  # Update collection
   ```

2. **Version control**

   - Commit `bff-postman-collection.json`
   - Team members can import same collection

3. **Environment variables**

   - Don't hardcode tokens
   - Use Postman environments

4. **Naming conventions**

   - Use descriptive endpoint names
   - Group related endpoints in services

5. **Keep in sync**
   - Regenerate when config changes
   - Test collection after regeneration

---

## Future Enhancements

Potential additions (not yet implemented):

- [ ] Request body examples
- [ ] Response examples
- [ ] Authentication configurations
- [ ] Test scripts generation
- [ ] Multiple environment files
- [ ] OpenAPI → Postman conversion

---

## Related Commands

```bash
# Initialize project
bffgen init my-bff --lang nodejs-express

# Add endpoints
bffgen add-route

# Generate code
bffgen generate

# Generate Postman
bffgen postman

# Validate config
bffgen config validate
```

---

## Conclusion

The `postman` command provides:

✅ Automatic runtime detection  
✅ Validates configuration  
✅ Generates Postman v2.1.0 collections  
✅ Includes all endpoints + health check  
✅ Uses variables for flexibility  
✅ Works with both Go and Node.js

**Generated collections are ready to import and use immediately!**
