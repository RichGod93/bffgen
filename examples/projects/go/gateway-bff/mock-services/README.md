# Mock Backend Services

This directory contains mock backend microservices for testing the Gateway BFF.

## Services

### 1. Users Service (Port 4000)

**Base URL:** `http://localhost:4000/api`

Endpoints:

- `GET /api/users` - List all users
- `POST /api/users` - Create a new user
- `GET /api/users/{id}` - Get user by ID
- `PUT /api/users/{id}` - Update user
- `DELETE /api/users/{id}` - Delete user
- `GET /health` - Health check

### 2. Analytics Service (Port 4001)

**Base URL:** `http://localhost:4001/api`

Endpoints:

- `GET /api/metrics` - Get system metrics
- `GET /api/events` - List all events
- `POST /api/events` - Create event
- `GET /health` - Health check

### 3. Notifications Service (Port 4002)

**Base URL:** `http://localhost:4002/api`

Endpoints:

- `GET /api/notifications` - List notifications
- `GET /api/notifications/{id}` - Get notification by ID
- `POST /api/notifications/{id}/read` - Mark notification as read
- `POST /api/notifications/read-all` - Mark all notifications as read
- `GET /health` - Health check

## Running the Services

### Option 1: Run Individually

```bash
# Terminal 1 - Users Service
cd users && go run main.go

# Terminal 2 - Analytics Service
cd analytics && go run main.go

# Terminal 3 - Notifications Service
cd notifications && go run main.go
```

### Option 2: Use the Helper Script

```bash
# Make the script executable
chmod +x run-all-services.sh

# Run all services
./run-all-services.sh
```

### Option 3: Run in Background

```bash
(cd users && go run main.go) &
(cd analytics && go run main.go) &
(cd notifications && go run main.go) &

# To stop all services
pkill -f "users/main\|analytics/main\|notifications/main"
```

## Testing the Services

### Health Checks

```bash
curl http://localhost:4000/health
curl http://localhost:4001/health
curl http://localhost:4002/health
```

### Users Service

```bash
# List users
curl http://localhost:4000/api/users

# Get specific user
curl http://localhost:4000/api/users/1

# Create user
curl -X POST http://localhost:4000/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"New User","email":"newuser@example.com","role":"user"}'

# Update user
curl -X PUT http://localhost:4000/api/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Updated Name","email":"updated@example.com","role":"admin"}'

# Delete user
curl -X DELETE http://localhost:4000/api/users/1
```

### Analytics Service

```bash
# Get metrics
curl http://localhost:4001/api/metrics

# List events
curl http://localhost:4001/api/events

# Create event
curl -X POST http://localhost:4001/api/events \
  -H "Content-Type: application/json" \
  -d '{"type":"page_view","user_id":"1","data":{"page":"/home"}}'
```

### Notifications Service

```bash
# List all notifications
curl http://localhost:4002/api/notifications

# List notifications for specific user
curl "http://localhost:4002/api/notifications?user_id=1"

# Get specific notification
curl http://localhost:4002/api/notifications/notif-1

# Mark as read
curl -X POST http://localhost:4002/api/notifications/notif-1/read

# Mark all as read for user
curl -X POST "http://localhost:4002/api/notifications/read-all?user_id=1"
```

## Testing via Gateway BFF

Once the Gateway BFF is running on port 8080, you can access all services through it:

```bash
# Via Gateway (port 8080)
curl http://localhost:8080/api/users
curl http://localhost:8080/api/analytics/metrics
curl http://localhost:8080/api/notifications

# Dashboard aggregator endpoint
curl http://localhost:8080/api/dashboard
```

## Notes

- All services support CORS for frontend integration
- Services use in-memory storage (data resets on restart)
- No authentication required for mock services (Gateway handles auth)
- All responses are JSON formatted
