#!/bin/bash

# Promptlands dev runner - starts backend and frontend together

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Load environment variables from .env
if [ -f "backend/.env" ]; then
    export $(grep -v '^#' backend/.env | xargs)
    echo -e "${GREEN}Loaded environment from backend/.env${NC}"
fi

# Trap to kill background processes on exit
cleanup() {
    echo -e "\n${BLUE}Shutting down...${NC}"
    kill $BACKEND_PID 2>/dev/null || true
    kill $FRONTEND_PID 2>/dev/null || true
    exit 0
}
trap cleanup SIGINT SIGTERM

# Check if dependencies are installed
if [ ! -d "frontend/node_modules" ]; then
    echo -e "${BLUE}Installing frontend dependencies...${NC}"
    cd frontend && npm install && cd ..
fi

echo -e "${GREEN}Starting Promptlands...${NC}"
echo -e "${BLUE}Backend:${NC}  http://localhost:8080"
echo -e "${BLUE}Frontend:${NC} http://localhost:5173"
echo -e "${BLUE}Press Ctrl+C to stop${NC}\n"

# Start backend
cd backend
if [ -n "$GEMINI_API_KEY" ]; then
    echo -e "${GREEN}Using Gemini API (real LLM)${NC}"
    go run ./cmd/server --no-db 2>&1 | sed 's/^/[backend] /' &
else
    echo -e "${BLUE}No API key - using mock LLM${NC}"
    go run ./cmd/server --dev 2>&1 | sed 's/^/[backend] /' &
fi
BACKEND_PID=$!
cd ..

# Wait a moment for backend to start
sleep 1

# Start frontend
cd frontend
npm run dev 2>&1 | sed 's/^/[frontend] /' &
FRONTEND_PID=$!
cd ..

# Wait for both processes
wait $BACKEND_PID $FRONTEND_PID
