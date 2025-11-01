#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Starting Mock Backend Microservices      ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
echo ""

# Function to check if a port is in use
check_port() {
    lsof -i :$1 > /dev/null 2>&1
    return $?
}

# Function to kill process on port
kill_port() {
    if check_port $1; then
        echo -e "${RED}⚠️  Port $1 is in use. Killing existing process...${NC}"
        lsof -ti :$1 | xargs kill -9 2>/dev/null
        sleep 1
    fi
}

# Clean up any existing services
kill_port 4000
kill_port 4001
kill_port 4002

echo -e "${GREEN}Starting services...${NC}"
echo ""

# Start Users Service
echo -e "${BLUE}🚀 Starting Users Service (Port 4000)...${NC}"
(cd users && go run main.go) > /dev/null 2>&1 &
USERS_PID=$!
sleep 1

# Start Analytics Service
echo -e "${BLUE}📊 Starting Analytics Service (Port 4001)...${NC}"
(cd analytics && go run main.go) > /dev/null 2>&1 &
ANALYTICS_PID=$!
sleep 1

# Start Notifications Service
echo -e "${BLUE}🔔 Starting Notifications Service (Port 4002)...${NC}"
(cd notifications && go run main.go) > /dev/null 2>&1 &
NOTIFICATIONS_PID=$!
sleep 1

echo ""
echo -e "${GREEN}✅ All services started!${NC}"
echo ""
echo -e "${BLUE}Service Status:${NC}"
echo "  - Users Service:         http://localhost:4000 (PID: $USERS_PID)"
echo "  - Analytics Service:     http://localhost:4001 (PID: $ANALYTICS_PID)"
echo "  - Notifications Service: http://localhost:4002 (PID: $NOTIFICATIONS_PID)"
echo ""
echo -e "${BLUE}Health Checks:${NC}"

# Test health endpoints
sleep 2
if curl -s http://localhost:4000/health > /dev/null; then
    echo -e "  - Users:         ${GREEN}✓ Healthy${NC}"
else
    echo -e "  - Users:         ${RED}✗ Not responding${NC}"
fi

if curl -s http://localhost:4001/health > /dev/null; then
    echo -e "  - Analytics:     ${GREEN}✓ Healthy${NC}"
else
    echo -e "  - Analytics:     ${RED}✗ Not responding${NC}"
fi

if curl -s http://localhost:4002/health > /dev/null; then
    echo -e "  - Notifications: ${GREEN}✓ Healthy${NC}"
else
    echo -e "  - Notifications: ${RED}✗ Not responding${NC}"
fi

echo ""
echo -e "${BLUE}Press Ctrl+C to stop all services${NC}"
echo ""

# Trap Ctrl+C to cleanly shut down services
trap "echo ''; echo -e '${RED}Stopping all services...${NC}'; kill $USERS_PID $ANALYTICS_PID $NOTIFICATIONS_PID 2>/dev/null; exit" INT TERM

# Wait for all background processes
wait

