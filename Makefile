.PHONY: dev dev-backend dev-frontend install build clean db-up db-down

# Run the full stack in dev mode (no database required)
dev:
	@echo "Starting Promptlands in dev mode..."
	@make -j2 dev-backend dev-frontend

dev-backend:
	@cd backend && go run ./cmd/server --dev

dev-frontend:
	@cd frontend && npm run dev

# Install dependencies
install:
	@echo "Installing backend dependencies..."
	@cd backend && go mod tidy
	@echo "Installing frontend dependencies..."
	@cd frontend && npm install

# Build for production
build:
	@echo "Building backend..."
	@cd backend && go build -o bin/server ./cmd/server
	@echo "Building frontend..."
	@cd frontend && npm run build

# Start databases
db-up:
	docker-compose up -d

db-down:
	docker-compose down

# Run with database (production-like)
run: db-up
	@sleep 2
	@make -j2 run-backend dev-frontend

run-backend:
	@cd backend && go run ./cmd/server

# Clean build artifacts
clean:
	@rm -rf backend/bin
	@rm -rf frontend/.svelte-kit
	@rm -rf frontend/build
