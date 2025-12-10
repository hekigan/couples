.PHONY: help build run test test-short test-full test-coverage test-coverage-html clean docker-build docker-run sass sass-watch dev
.PHONY: js-build js-build-dev js-watch js-clean
.PHONY: test-db-setup test-db-start test-db-stop test-db-reset test-db-status test-db-studio
.PHONY: test-e2e test-e2e-ui test-e2e-headed test-e2e-debug test-e2e-report test-e2e-setup
.PHONY: templ-install templ-generate templ-watch templ-clean

# Default target
help:
	@echo "Couple Card Game - Makefile Commands"
	@echo ""
	@echo "Available targets:"
	@echo ""
	@echo "Build & Run:"
	@echo "  make build         - Build the Go binary"
	@echo "  make run           - Run the server"
	@echo "  make dev           - Run in development mode with Air hot-reload"
	@echo ""
	@echo "Testing:"
	@echo "  make test          - Run short tests (unit tests only)"
	@echo "  make test-full     - Run full test suite (requires test DB)"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make test-coverage-html - Open coverage report in browser"
	@echo ""
	@echo "E2E Testing:"
	@echo "  make test-e2e        - Run E2E tests with Playwright"
	@echo "  make test-e2e-ui     - Run E2E tests in Playwright UI mode"
	@echo "  make test-e2e-headed - Run E2E tests in headed browser mode"
	@echo "  make test-e2e-debug  - Run E2E tests in debug mode"
	@echo "  make test-e2e-report - Open Playwright test report"
	@echo "  make test-e2e-setup  - Setup E2E testing (one-time)"
	@echo ""
	@echo "Test Database:"
	@echo "  make test-db-setup  - Setup test database (one-time)"
	@echo "  make test-db-start  - Start test database"
	@echo "  make test-db-stop   - Stop test database"
	@echo "  make test-db-reset  - Reset test database (clean slate)"
	@echo "  make test-db-status - Show test database status"
	@echo "  make test-db-studio - Open Supabase Studio UI"
	@echo ""
	@echo "Styling:"
	@echo "  make sass          - Compile SASS to CSS"
	@echo "  make sass-watch    - Watch and compile SASS"
	@echo ""
	@echo "JavaScript:"
	@echo "  make js-build      - Build JS bundles (production)"
	@echo "  make js-build-dev  - Build JS bundles (development)"
	@echo "  make js-watch      - Watch and rebuild JS bundles"
	@echo "  make js-clean      - Clean JS bundles"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-run    - Run Docker container"
	@echo "  make docker-stop   - Stop Docker container"
	@echo ""
	@echo "Utilities:"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make install       - Install dependencies (one-time setup)"
	@echo "  make fmt           - Format code"
	@echo "  make lint          - Lint code"
	@echo ""

# Build the Go binary
build: templ-generate sass js-build
	@echo "Building Go binary..."
	@go build -o server ./cmd/server
	@echo "Build complete: ./server"

# Run the server
run: build
	@echo "Starting server..."
	@ENV=production ./server

# Run short tests (unit tests only, no database required)
test:
	@echo "Running short tests (unit tests only)..."
	@go test -v -short ./...

# Alias for test
test-short: test

# Run full test suite (requires test database)
test-full:
	@echo "Running full test suite..."
	@echo "Make sure test database is running: make test-db-status"
	@go test -v ./internal/services/...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -short -coverprofile=coverage.out ./internal/services/...
	@go tool cover -func=coverage.out | grep total
	@echo ""
	@echo "Coverage report saved to: coverage.out"
	@echo "View HTML report with: make test-coverage-html"

# Open coverage report in browser
test-coverage-html: test-coverage
	@echo "Opening coverage report in browser..."
	@go tool cover -html=coverage.out

# Clean build artifacts
clean: js-clean templ-clean
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
	@npx sass sass/admin.scss static/css/admin.css
	@echo "SASS compilation complete"

# Watch and compile SASS
sass-watch:
	@echo "Watching SASS files..."
	@npx sass --watch sass/main.scss:static/css/main.css sass/admin.scss:static/css/admin.css

# ============================================
# JavaScript Bundling (esbuild)
# ============================================

# Build JavaScript bundles (production mode)
js-build:
	@echo "Building JavaScript bundles (production)..."
	@ENV=production go run ./cmd/esbuild/main.go build
	@echo "✅ JavaScript bundles built"

# Build JavaScript bundles (development mode)
js-build-dev:
	@echo "Building JavaScript bundles (development)..."
	@ENV=development go run ./cmd/esbuild/main.go build
	@echo "✅ JavaScript bundles built (dev mode)"

# Watch and rebuild JavaScript bundles
js-watch:
	@echo "Watching JavaScript files..."
	@ENV=development go run ./cmd/esbuild/main.go watch

# Clean JavaScript bundles
js-clean:
	@rm -rf static/dist/*.js static/dist/*.js.map
	@echo "✅ JavaScript bundles cleaned"

# ============================================
# Templ Generation Commands
# ============================================

# Install templ CLI (one-time setup)
templ-install:
	@echo "Installing templ CLI..."
	@go install github.com/a-h/templ/cmd/templ@latest
	@echo "✅ Templ CLI installed"

# Generate templ components
templ-generate:
	@echo "Generating templ components..."
	@$(shell go env GOPATH)/bin/templ generate
	@echo "✅ Templ components generated"

# Watch and regenerate templ components
templ-watch:
	@echo "Watching templ files..."
	@$(shell go env GOPATH)/bin/templ generate --watch

# Clean generated templ files
templ-clean:
	@echo "Cleaning generated templ files..."
	@find internal/views -name "*_templ.go" -type f -delete 2>/dev/null || true
	@echo "✅ Generated templ files cleaned"

# ============================================
# Development & Building
# ============================================

# Development mode with Air hot-reload
dev: templ-generate sass js-build-dev
	@echo "🚀 Starting development mode with Air hot-reload..."
	@echo "⚠️  For full hot-reload, run in separate terminals:"
	@echo "   Terminal 2: make sass-watch"
	@echo "   Terminal 3: make js-watch"
	@echo "   Terminal 4: make templ-watch"
	@ENV=development $(shell go env GOPATH)/bin/air

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

# Install dependencies (one-time setup)
install:
	@echo "Installing Go dependencies..."
	@go mod download
	@echo "Installing Node dependencies..."
	@npm install -g sass
	@echo "Installing Air hot-reload tool..."
	@go install github.com/air-verse/air@latest
	@echo "Installing Templ CLI..."
	@go install github.com/a-h/templ/cmd/templ@latest
	@echo "Dependencies installed"

# Alias for backward compatibility
deps: install

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
setup: install sass db-setup
	@echo "Setup complete!"
	@echo "1. Copy .env.example to .env"
	@echo "2. Update .env with your Supabase credentials"
	@echo "3. Run 'make build' to build the server"
	@echo "4. Run 'make run' to start the server"

# ============================================
# Test Database Commands
# ============================================

# Setup test database (one-time setup)
test-db-setup:
	@echo "Setting up test database..."
	@chmod +x scripts/setup-test-db.sh
	@./scripts/setup-test-db.sh

# Start test database
test-db-start:
	@echo "Starting test database..."
	@supabase start
	@echo ""
	@echo "✅ Test database started!"
	@echo "   Studio UI: http://localhost:54323"
	@echo "   API URL:   http://localhost:54321"
	@echo ""
	@echo "Run tests with: make test-full"

# Stop test database
test-db-stop:
	@echo "Stopping test database..."
	@supabase stop
	@echo "✅ Test database stopped"

# Reset test database (clean slate with fresh schema and seed data)
test-db-reset:
	@echo "Resetting test database..."
	@supabase db reset
	@echo "✅ Test database reset complete"
	@echo "All data wiped and schema/seed reapplied"

# Show test database status
test-db-status:
	@echo "Test database status:"
	@echo ""
	@supabase status || (echo "❌ Test database not running" && echo "Start with: make test-db-start" && exit 1)

# Open Supabase Studio UI in browser
test-db-studio:
	@echo "Opening Supabase Studio..."
	@open http://localhost:54323 2>/dev/null || xdg-open http://localhost:54323 2>/dev/null || echo "Please open http://localhost:54323 in your browser"

# ============================================
# E2E Testing Commands (Playwright)
# ============================================

# Setup E2E testing (one-time setup)
test-e2e-setup:
	@echo "Setting up E2E testing with Playwright..."
	@npm install
	@npx playwright install chromium
	@echo "✅ E2E testing setup complete"
	@echo ""
	@echo "Run tests with: make test-e2e"

# Run E2E tests
test-e2e:
	@echo "Running E2E tests with Playwright..."
	@npx playwright test

# Run E2E tests in UI mode (interactive)
test-e2e-ui:
	@echo "Opening Playwright UI..."
	@npx playwright test --ui

# Run E2E tests in headed mode (visible browser)
test-e2e-headed:
	@echo "Running E2E tests in headed mode..."
	@npx playwright test --headed

# Run E2E tests in debug mode
test-e2e-debug:
	@echo "Running E2E tests in debug mode..."
	@npx playwright test --debug

# Open Playwright test report
test-e2e-report:
	@echo "Opening Playwright test report..."
	@npx playwright show-report



