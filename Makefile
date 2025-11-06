.PHONY: help build run test clean docker-build docker-run sass dev

# Default target
help:
	@echo "Couple Card Game - Makefile Commands"
	@echo ""
	@echo "Available targets:"
	@echo "  make build       - Build the Go binary"
	@echo "  make run         - Run the server"
	@echo "  make test        - Run tests"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make sass        - Compile SASS to CSS"
	@echo "  make sass-watch  - Watch and compile SASS"
	@echo "  make dev         - Run in development mode (with sass watch)"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run  - Run Docker container"
	@echo ""

# Build the Go binary
build:
	@echo "Building Go binary..."
	@go build -o server ./cmd/server/main.go
	@echo "Build complete: ./server"

# Run the server
run: sass
	@echo "Starting server..."
	@./server

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f server
	@rm -f couple-game
	@rm -f main
	@rm -rf static/css/main.css*
	@echo "Clean complete"

# Compile SASS to CSS
sass:
	@echo "Compiling SASS..."
	@npx sass sass/main.scss static/css/main.css
	@echo "SASS compilation complete"

# Watch and compile SASS
sass-watch:
	@echo "Watching SASS files..."
	@npx sass --watch sass/main.scss static/css/main.css

# Development mode
dev:
	@echo "Starting development mode..."
	@make sass
	@go run ./cmd/server/main.go

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t couple-card-game:latest .
	@echo "Docker image built"

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	@docker-compose up -d
	@echo "Container started"

# Stop Docker container
docker-stop:
	@docker-compose down

# View Docker logs
docker-logs:
	@docker-compose logs -f

# Install dependencies
deps:
	@echo "Installing Go dependencies..."
	@go mod download
	@echo "Installing Node dependencies..."
	@npm install -g sass
	@echo "Dependencies installed"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatted"

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run
	@echo "Linting complete"

# Database setup
db-setup:
	@echo "Setting up database..."
	@echo "Please run sql/schema.sql and sql/seed.sql in your Supabase dashboard"

# Full setup
setup: deps sass db-setup
	@echo "Setup complete!"
	@echo "1. Copy .env.example to .env"
	@echo "2. Update .env with your Supabase credentials"
	@echo "3. Run 'make build' to build the server"
	@echo "4. Run 'make run' to start the server"

