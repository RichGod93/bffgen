#!/bin/bash

BASE_URL="http://localhost:8080"
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     Testing Gateway BFF Endpoints         ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
echo ""

test_endpoint() {
    local method=$1
    local endpoint=$2
    local expected_code=$3
    local description=$4
    local data=$5
    
    if [ -z "$data" ]; then
        response=$(curl -s -o /dev/null -w "%{http_code}" -X $method "$BASE_URL$endpoint")
    else
        response=$(curl -s -o /dev/null -w "%{http_code}" -X $method "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi
    
    if [ "$response" -eq "$expected_code" ]; then
        echo -e "${GREEN}✓${NC} $description (HTTP $response)"
        return 0
    else
        echo -e "${RED}✗${NC} $description (Expected $expected_code, got $response)"
        return 1
    fi
}

# Track results
passed=0
failed=0

# Health check
echo -e "${BLUE}[Health Check]${NC}"
if test_endpoint "GET" "/health" 200 "Health check"; then
    ((passed++))
else
    ((failed++))
fi
echo ""

# Users endpoints
echo -e "${BLUE}[Users Service - 5 endpoints]${NC}"
if test_endpoint "GET" "/api/users" 200 "List all users"; then ((passed++)); else ((failed++)); fi
if test_endpoint "GET" "/api/users/1" 200 "Get user by ID"; then ((passed++)); else ((failed++)); fi
if test_endpoint "POST" "/api/users" 201 "Create user" '{"name":"Test User","email":"test@example.com","role":"user"}'; then ((passed++)); else ((failed++)); fi
if test_endpoint "PUT" "/api/users/1" 200 "Update user" '{"name":"Updated User","email":"updated@example.com","role":"admin"}'; then ((passed++)); else ((failed++)); fi
if test_endpoint "DELETE" "/api/users/5" 200 "Delete user"; then ((passed++)); else ((failed++)); fi
echo ""

# Analytics endpoints
echo -e "${BLUE}[Analytics Service - 3 endpoints]${NC}"
if test_endpoint "GET" "/api/analytics/metrics" 200 "Get system metrics"; then ((passed++)); else ((failed++)); fi
if test_endpoint "GET" "/api/analytics/events" 200 "Get analytics events"; then ((passed++)); else ((failed++)); fi
if test_endpoint "POST" "/api/analytics/events" 201 "Create analytics event" '{"type":"test_event","user_id":"1","data":{"test":true}}'; then ((passed++)); else ((failed++)); fi
echo ""

# Notifications endpoints
echo -e "${BLUE}[Notifications Service - 4 endpoints]${NC}"
if test_endpoint "GET" "/api/notifications" 200 "Get all notifications"; then ((passed++)); else ((failed++)); fi
if test_endpoint "GET" "/api/notifications/notif-1" 200 "Get specific notification"; then ((passed++)); else ((failed++)); fi
if test_endpoint "POST" "/api/notifications/notif-1/read" 200 "Mark notification as read"; then ((passed++)); else ((failed++)); fi
if test_endpoint "POST" "/api/notifications/read-all" 200 "Mark all as read"; then ((passed++)); else ((failed++)); fi
echo ""

# Summary
echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║              Test Summary                  ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
echo -e "${GREEN}Passed:${NC} $passed"
echo -e "${RED}Failed:${NC} $failed"
echo -e "Total:  $((passed + failed))"
echo ""

if [ $failed -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed${NC}"
    exit 1
fi

