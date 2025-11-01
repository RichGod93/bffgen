# Quick Start Guide - Gateway BFF

Get the Gateway BFF up and running in 5 minutes.

## Prerequisites

- Go 1.21+ installed
- Basic understanding of REST APIs
- Terminal access

## Step-by-Step Setup

### Step 1: Review the Project Structure

```bash
cd gateway-bff
ls -la

# You should see:
# - main.go                      (Gateway server)
# - bff.config.yaml              (Configuration)
# - mock-services/               (Test backends)
# - bff-postman-collection.json  (API testing)
```

### Step 2: Start Mock Backend Services

Open a new terminal window:

```bash
cd gateway-bff/mock-services
chmod +x run-all-services.sh
./run-all-services.sh
```

You should see:

```
✅ All services started!
Service Status:
  - Users Service:         http://localhost:4000
  - Analytics Service:     http://localhost:4001
  - Notifications Service: http://localhost:4002
```

**Keep this terminal running!**

### Step 3: Set Environment Variables

In a new terminal (in the `gateway-bff` directory):

```bash
export JWT_SECRET=my-super-secret-jwt-key-for-testing
export ENCRYPTION_KEY=my-encryption-key-for-testing-only
```

### Step 4: Start the Gateway BFF

```bash
go run cmd/server/main.go
```

Expected output:

```
🚀 BFF server starting on :8080
```

**Keep this terminal running!**

### Step 5: Test the Gateway

Open a third terminal and test the endpoints:

```bash
# 1. Health check
curl http://localhost:8080/health
# Response: BFF server is running!

# 2. List users (no auth required)
curl http://localhost:8080/api/users
# Response: {"users": [...], "total": 4}

# 3. Get specific user
curl http://localhost:8080/api/users/1
# Response: {"id": "1", "name": "Alice Johnson", ...}

# 4. Get analytics metrics (requires auth - see below)
curl http://localhost:8080/api/analytics/metrics
# Response: 401 Unauthorized (expected without JWT token)
```

## Testing Authenticated Endpoints

For endpoints that require authentication, you need a JWT token.

### Option 1: Use Postman Collection

1. Open Postman
2. Import `bff-postman-collection.json`
3. Set environment variable:
   - `baseUrl`: `http://localhost:8080`
   - `token`: Your JWT token (see Option 2 for generation)
4. Test all endpoints with one click

### Option 2: Generate JWT Token Manually

Create a file `generate-token.go`:

```go
package main

import (
	"fmt"
	"os"
	"time"
	"github.com/golang-jwt/jwt/v5"
)

func main() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "my-super-secret-jwt-key-for-testing"
	}

	claims := jwt.MapClaims{
		"user_id": "1",
		"email":   "alice@example.com",
		"role":    "admin",
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))

	fmt.Println("JWT Token:")
	fmt.Println(tokenString)
}
```

Generate token:

```bash
go run generate-token.go
```

Use the token:

```bash
TOKEN="<paste-token-here>"

# Test authenticated endpoint
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/analytics/metrics

# Test aggregated dashboard
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/dashboard
```

## Testing All Services

### Users Service

```bash
# List all users
curl http://localhost:8080/api/users

# Get user by ID
curl http://localhost:8080/api/users/1

# Create user (requires auth)
curl -X POST http://localhost:8080/api/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"New User","email":"new@example.com","role":"user"}'

# Update user (requires auth)
curl -X PUT http://localhost:8080/api/users/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Updated Name","email":"updated@example.com","role":"admin"}'

# Delete user (requires auth)
curl -X DELETE http://localhost:8080/api/users/1 \
  -H "Authorization: Bearer $TOKEN"
```

### Analytics Service

```bash
# Get system metrics (requires auth)
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/analytics/metrics

# List events (requires auth)
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/analytics/events

# Create event (requires auth)
curl -X POST http://localhost:8080/api/analytics/events \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"type":"page_view","user_id":"1","data":{"page":"/dashboard"}}'
```

### Notifications Service

```bash
# List notifications (requires auth)
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/notifications

# Get specific notification (requires auth)
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/notifications/notif-1

# Mark as read (requires auth)
curl -X POST http://localhost:8080/api/notifications/notif-1/read \
  -H "Authorization: Bearer $TOKEN"

# Mark all as read (requires auth)
curl -X POST http://localhost:8080/api/notifications/read-all \
  -H "Authorization: Bearer $TOKEN"
```

### Aggregated Dashboard

```bash
# Get combined user dashboard (requires auth)
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/dashboard
```

This endpoint combines data from:

- Users service (user profile)
- Analytics service (metrics)
- Notifications service (notifications)

## Troubleshooting

### Port Already in Use

```bash
# Kill processes on ports
lsof -ti :8080 | xargs kill -9  # Gateway
lsof -ti :4000 | xargs kill -9  # Users service
lsof -ti :4001 | xargs kill -9  # Analytics service
lsof -ti :4002 | xargs kill -9  # Notifications service
```

### Mock Services Not Responding

```bash
# Check if services are running
curl http://localhost:4000/health
curl http://localhost:4001/health
curl http://localhost:4002/health

# Restart services
cd mock-services
./run-all-services.sh
```

### 401 Unauthorized

- Ensure you're using a valid JWT token
- Check that JWT_SECRET matches between token generation and gateway
- Verify token hasn't expired (tokens expire after 24 hours by default)

### Connection Refused

- Ensure mock services are running first
- Check that ports 4000, 4001, 4002 are not blocked
- Verify no firewall is blocking connections

## Next Steps

1. ✅ Test all endpoints with curl or Postman
2. 🔄 Modify `bff.config.yaml` to add custom endpoints
3. 🔄 Replace mock services with real backends
4. 🔄 Deploy using Docker: `docker-compose up`
5. 🔄 Set up CI/CD pipeline (already configured in `.github/workflows/ci.yml`)

## Quick Reference

**Gateway URL:** http://localhost:8080  
**Users Service:** http://localhost:4000  
**Analytics Service:** http://localhost:4001  
**Notifications Service:** http://localhost:4002

**Configuration:** `bff.config.yaml`  
**Postman Collection:** `bff-postman-collection.json`  
**Project Overview:** `PROJECT_OVERVIEW.md`

## Getting Help

- Review `PROJECT_OVERVIEW.md` for architecture details
- Check `mock-services/README.md` for backend service docs
- Run `bffgen doctor` to check project health
- Visit [bffgen documentation](https://github.com/RichGod93/bffgen)

---

**Happy coding! 🚀**
