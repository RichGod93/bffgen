#!/bin/bash

# Start All Services Script
# Starts both the mock user service and the main BFF service

echo "🎬 Starting Movie Dashboard BFF Services..."
echo ""

# Check if virtual environment exists
if [ ! -d "venv" ]; then
    echo "❌ Virtual environment not found. Run ./setup.sh first."
    exit 1
fi

# Activate virtual environment
source venv/bin/activate

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "🛑 Stopping services..."
    kill $USER_SERVICE_PID $BFF_PID 2>/dev/null
    exit 0
}

trap cleanup SIGINT SIGTERM

# Start mock user service in background
echo "🚀 Starting Mock User Service (port 3001)..."
cd mock-user-service
python main.py > ../logs/user-service.log 2>&1 &
USER_SERVICE_PID=$!
cd ..

# Wait a moment for user service to start
sleep 2

# Check if user service started successfully
if ! ps -p $USER_SERVICE_PID > /dev/null; then
    echo "❌ Failed to start Mock User Service"
    cat logs/user-service.log
    exit 1
fi

echo "✅ Mock User Service started (PID: $USER_SERVICE_PID)"

# Start main BFF service in background
echo "🚀 Starting Movie Dashboard BFF (port 8000)..."
uvicorn main:app --reload > logs/bff.log 2>&1 &
BFF_PID=$!

# Wait a moment for BFF to start
sleep 3

# Check if BFF started successfully
if ! ps -p $BFF_PID > /dev/null; then
    echo "❌ Failed to start BFF"
    cat logs/bff.log
    kill $USER_SERVICE_PID 2>/dev/null
    exit 1
fi

echo "✅ Movie Dashboard BFF started (PID: $BFF_PID)"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🎉 All services are running!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📖 API Documentation:"
echo "   Main BFF:        http://localhost:8000/docs"
echo "   User Service:    http://localhost:3001/docs"
echo ""
echo "🔗 Health Checks:"
echo "   Main BFF:        http://localhost:8000/health"
echo "   User Service:    http://localhost:3001/health"
echo ""
echo "📊 Example Endpoints:"
echo "   Personalized Feed:   http://localhost:8000/api/dashboard/feed"
echo "   Popular Movies:      http://localhost:8000/api/movies/popular"
echo "   User Favorites:      http://localhost:3001/favorites"
echo ""
echo "📝 Logs:"
echo "   BFF:             tail -f logs/bff.log"
echo "   User Service:    tail -f logs/user-service.log"
echo ""
echo "Press Ctrl+C to stop all services"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Keep script running and show combined logs
tail -f logs/bff.log logs/user-service.log &
TAIL_PID=$!

# Wait for interrupt
wait $BFF_PID

