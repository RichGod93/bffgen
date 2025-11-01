# Testing Gateway BFF

Complete guide to testing the API Gateway with mock services.

## Prerequisites

- Go 1.21+ installed
- Terminals available (3 recommended)
- curl or Postman

## Quick Test (5 minutes)

### 1. Start Mock Services

**Terminal 1:**

```bash
cd gateway-bff/mock-services
./run-all-services.sh
```

Expected output:

```
✅ All services started!
Service Status:
  - Users Service:         http://localhost:4000 (PID: xxxx)
  - Analytics Service:     http://localhost:4001 (PID: xxxx)
  - Notifications Service: http://localhost:4002 (PID: xxxx)

Health Checks:
  - Users:         ✓ Healthy
  - Analytics:     ✓ Healthy
  - Notifications: ✓ Healthy
```

### 2. Start Gateway

**Terminal 2:**

```bash
cd gateway-bff
go run cmd/server/main.go
```

Expected output:

```
🚀 BFF server starting on :8080
```

### 3. Test Endpoints

**Terminal 3:**

```bash
# Health check
curl http://localhost:8080/health

# List users (public endpoint)
curl http://localhost:8080/api/users

# Get specific user
curl http://localhost:8080/api/users/1
```

## Detailed Testing

### Test All User Endpoints

```bash
# 1. List all users
curl http://localhost:8080/api/users
# Expected: {"users": [...], "total": 4}

# 2. Get user by ID
curl http://localhost:8080/api/users/1
# Expected: {"id": "1", "name": "Alice Johnson", ...}

# 3. Create user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"New User","email":"new@example.com","role":"user"}'
# Expected: 201 Created

# 4. Update user
curl -X PUT http://localhost:8080/api/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Updated Alice","email":"alice@example.com","role":"admin"}'
# Expected: {"message": "User updated successfully", ...}

# 5. Delete user
curl -X DELETE http://localhost:8080/api/users/4
# Expected: {"message": "User deleted successfully"}
```

### Test Analytics Endpoints

```bash
# 1. Get system metrics
curl http://localhost:8080/api/analytics/metrics
# Expected: {"metrics": [...], "total": 6}

# 2. Get events
curl http://localhost:8080/api/analytics/events
# Expected: {"events": [...], "total": 3}

# 3. Create event
curl -X POST http://localhost:8080/api/analytics/events \
  -H "Content-Type: application/json" \
  -d '{"type":"page_view","user_id":"1","data":{"page":"/dashboard"}}'
# Expected: 201 Created
```

### Test Notifications Endpoints

```bash
# 1. Get all notifications
curl http://localhost:8080/api/notifications
# Expected: {"notifications": [...], "total": 4, "unread": 3}

# 2. Get notifications for specific user
curl "http://localhost:8080/api/notifications?user_id=1"
# Expected: Filtered notifications for user 1

# 3. Get specific notification
curl http://localhost:8080/api/notifications/notif-1
# Expected: {"id": "notif-1", "title": "Welcome!", ...}

# 4. Mark notification as read
curl -X POST http://localhost:8080/api/notifications/notif-1/read
# Expected: {"message": "Notification marked as read", ...}

# 5. Mark all as read
curl -X POST http://localhost:8080/api/notifications/read-all
# Expected: {"message": "X notifications marked as read", ...}
```

## Testing with Postman

### Import Collection

1. Open Postman
2. Click **Import**
3. Select `bff-postman-collection.json`
4. Collection will load with all 12 endpoints

### Set Variables

1. Click on the collection name
2. Go to **Variables** tab
3. Set:
   - `baseUrl`: `http://localhost:8080`
   - `token`: (leave empty for public endpoints)

### Run Collection

1. Click collection **Run** button
2. Select all requests
3. Click **Run Gateway BFF**
4. View results

## Automated Testing Script

Save as `test-gateway.sh`:

```bash
#!/bin/bash

BASE_URL="http://localhost:8080"
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

test_endpoint() {
    local method=$1
    local endpoint=$2
    local expected_code=$3
    local description=$4

    response=$(curl -s -o /dev/null -w "%{http_code}" -X $method "$BASE_URL$endpoint")

    if [ "$response" -eq "$expected_code" ]; then
        echo -e "${GREEN}✓${NC} $description (HTTP $response)"
    else
        echo -e "${RED}✗${NC} $description (Expected $expected_code, got $response)"
    fi
}

echo "Testing Gateway BFF..."
echo ""

# Health check
test_endpoint "GET" "/health" 200 "Health check"

# Users
test_endpoint "GET" "/api/users" 200 "List users"
test_endpoint "GET" "/api/users/1" 200 "Get user by ID"

# Analytics
test_endpoint "GET" "/api/analytics/metrics" 200 "Get metrics"
test_endpoint "GET" "/api/analytics/events" 200 "Get events"

# Notifications
test_endpoint "GET" "/api/notifications" 200 "Get notifications"
test_endpoint "GET" "/api/notifications/notif-1" 200 "Get specific notification"

echo ""
echo "Testing complete!"
```

Run with:

```bash
chmod +x test-gateway.sh
./test-gateway.sh
```

## Testing Scenarios

### Scenario 1: Complete User Workflow

```bash
# 1. List all users
curl http://localhost:8080/api/users | jq

# 2. Create new user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"test@example.com","role":"user"}' | jq

# 3. Update the user (use ID from step 2)
curl -X PUT http://localhost:8080/api/users/5 \
  -H "Content-Type: application/json" \
  -d '{"name":"Updated Test User","email":"test@example.com","role":"admin"}' | jq

# 4. Verify update
curl http://localhost:8080/api/users/5 | jq

# 5. Delete the user
curl -X DELETE http://localhost:8080/api/users/5 | jq
```

### Scenario 2: Analytics Tracking

```bash
# 1. Record page view
curl -X POST http://localhost:8080/api/analytics/events \
  -H "Content-Type: application/json" \
  -d '{"type":"page_view","user_id":"1","data":{"page":"/home"}}' | jq

# 2. Record button click
curl -X POST http://localhost:8080/api/analytics/events \
  -H "Content-Type: application/json" \
  -d '{"type":"button_click","user_id":"1","data":{"button":"subscribe"}}' | jq

# 3. Check all events
curl http://localhost:8080/api/analytics/events | jq

# 4. View metrics
curl http://localhost:8080/api/analytics/metrics | jq
```

### Scenario 3: Notification Management

```bash
# 1. Get unread notifications for user
curl "http://localhost:8080/api/notifications?user_id=1" | jq

# 2. Mark first notification as read
curl -X POST http://localhost:8080/api/notifications/notif-1/read | jq

# 3. Verify it's marked as read
curl http://localhost:8080/api/notifications/notif-1 | jq

# 4. Mark all remaining as read
curl -X POST "http://localhost:8080/api/notifications/read-all?user_id=1" | jq
```

## Load Testing

### Using Apache Bench

```bash
# Install (macOS)
brew install apache2

# Test health endpoint
ab -n 1000 -c 10 http://localhost:8080/health

# Test users endpoint
ab -n 1000 -c 10 http://localhost:8080/api/users
```

### Using wrk

```bash
# Install
brew install wrk

# Test with 10 connections for 30 seconds
wrk -t10 -c10 -d30s http://localhost:8080/api/users

# Test with POST request
wrk -t10 -c10 -d30s -s post.lua http://localhost:8080/api/analytics/events
```

## Troubleshooting Tests

### Gateway not starting

```bash
# Check if port 8080 is in use
lsof -i :8080

# Kill existing process
lsof -ti :8080 | xargs kill -9

# Restart gateway
go run cmd/server/main.go
```

### Mock services not responding

```bash
# Check service health
curl http://localhost:4000/health  # Users
curl http://localhost:4001/health  # Analytics
curl http://localhost:4002/health  # Notifications

# Restart services
cd mock-services
./run-all-services.sh
```

### Connection refused errors

```bash
# Ensure services are running in order:
# 1. Mock services (ports 4000-4002)
# 2. Gateway (port 8080)

# Check what's listening
lsof -i :4000
lsof -i :4001
lsof -i :4002
lsof -i :8080
```

## Testing Checklist

- [ ] All mock services start successfully
- [ ] Gateway starts on port 8080
- [ ] Health check returns 200 OK
- [ ] Can list all users
- [ ] Can get specific user
- [ ] Can create new user
- [ ] Can update user
- [ ] Can delete user
- [ ] Can get analytics metrics
- [ ] Can get analytics events
- [ ] Can create analytics events
- [ ] Can list notifications
- [ ] Can mark notification as read
- [ ] Can mark all notifications as read

## Next Steps

After successful testing:

1. **Add authentication** - Use `internal/auth/secure_auth.go`
2. **Add aggregator endpoint** - Use `internal/aggregators/dashboard.go`
3. **Deploy with Docker** - Use `docker-compose up`
4. **Set up monitoring** - Add metrics collection
5. **Replace mock services** - Connect to real backends

## Useful Commands

```bash
# View gateway logs with filtering
go run cmd/server/main.go | grep "Proxying"

# Test with verbose output
curl -v http://localhost:8080/api/users

# Test with headers
curl -H "Content-Type: application/json" \
     -H "Accept: application/json" \
     http://localhost:8080/api/users

# Save response to file
curl http://localhost:8080/api/users > response.json

# Test multiple endpoints in parallel
parallel curl ::: \
  http://localhost:8080/api/users \
  http://localhost:8080/api/analytics/metrics \
  http://localhost:8080/api/notifications
```
