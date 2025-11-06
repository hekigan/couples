# Test Database Setup Guide

**Last Updated**: November 2025

This guide provides multiple approaches for setting up a test database for running the full integration test suite.

---

## 📋 Overview

Your application uses **Supabase** (PostgreSQL + additional features), so you have several options:

1. ✅ **RECOMMENDED**: Supabase CLI with Docker (Local Supabase instance)
2. ⚡ **ALTERNATIVE**: Separate Supabase Test Project (Cloud-based)
3. 🔧 **ADVANCED**: Plain PostgreSQL with Docker (Manual setup)

---

## Option 1: Supabase CLI with Docker (RECOMMENDED) ⭐

This is the **best approach** because:
- ✅ Matches production environment exactly
- ✅ Isolated test environment
- ✅ Fast and reproducible
- ✅ No external dependencies
- ✅ Free and local

### Prerequisites

```bash
# Install Supabase CLI
brew install supabase/tap/supabase  # macOS
# OR
npm install -g supabase              # All platforms

# Install Docker Desktop
# Download from: https://www.docker.com/products/docker-desktop
```

### Step 1: Initialize Supabase Locally

```bash
# Navigate to project root
cd /Users/blanes.laurent/Documents/dev/tests/couples

# Initialize Supabase
supabase init

# This creates a 'supabase' directory with configuration
```

### Step 2: Start Local Supabase

```bash
# Start all Supabase services (PostgreSQL, Studio, Auth, etc.)
supabase start

# This will output connection details:
# - API URL: http://localhost:54321
# - Studio URL: http://localhost:54323
# - DB URL: postgresql://postgres:postgres@localhost:54322/postgres
# - Service Role Key: eyJhbGci...
```

**Note**: Supabase CLI automatically:
- Creates PostgreSQL container
- Sets up Auth service
- Provides Studio UI for database management
- Runs migrations automatically

### Step 3: Apply Database Schema

```bash
# Copy your schema to supabase/migrations
cp sql/schema.sql supabase/migrations/00000000000000_init_schema.sql

# Apply migration
supabase db reset

# OR apply manually via Supabase Studio:
# 1. Open http://localhost:54323
# 2. Go to SQL Editor
# 3. Paste contents of sql/schema.sql
# 4. Click "Run"
```

### Step 4: Seed Test Data

```bash
# Copy seed file
cp sql/seed.sql supabase/seed.sql

# Seed will run automatically on db reset, or run manually:
supabase db reset --seed-only

# OR via Studio SQL Editor:
# Paste contents of sql/seed.sql and run
```

### Step 5: Create Test Environment File

```bash
# Create .env.test file
cat > .env.test << 'EOF'
# Test Database Configuration
TEST_SUPABASE_URL=http://localhost:54321
TEST_SUPABASE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6ImFub24iLCJleHAiOjE5ODM4MTI5OTZ9.CRXP1A7WOeoJeXxjNni43kdQwgnWNReilDMblYTn_I0
TEST_SUPABASE_SERVICE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImV4cCI6MTk4MzgxMjk5Nn0.EGIM96RAZx35lJzdJsyH-qQwv8Hdp7fsn3W0YpN81IU

# These are default Supabase local keys - safe to commit
EOF
```

### Step 6: Update Test Files to Use Test Database

Create a test helper:

```go
// internal/services/test_helpers.go
package services

import (
    "os"
    "testing"

    "github.com/joho/godotenv"
    "github.com/supabase-community/supabase-go"
)

// SetupTestDatabase initializes a test database connection
func SetupTestDatabase(t *testing.T) *supabase.Client {
    // Load test environment
    godotenv.Load("../../.env.test")

    testURL := os.Getenv("TEST_SUPABASE_URL")
    testKey := os.Getenv("TEST_SUPABASE_KEY")

    if testURL == "" || testKey == "" {
        t.Skip("Test database not configured. Run 'supabase start' first.")
    }

    client, err := supabase.NewClient(testURL, testKey, nil)
    if err != nil {
        t.Fatalf("Failed to create test client: %v", err)
    }

    return client
}

// CleanupTestData cleans up test data after tests
func CleanupTestData(t *testing.T, client *supabase.Client) {
    // Clean up in reverse dependency order
    tables := []string{
        "question_history",
        "answers",
        "room_join_requests",
        "room_invitations",
        "notifications",
        "friends",
        "rooms",
        "users",
    }

    for _, table := range tables {
        _, _, err := client.From(table).Delete("", "").Execute()
        if err != nil {
            t.Logf("Warning: Failed to clean %s: %v", table, err)
        }
    }
}
```

### Step 7: Update Test Files

Example for `question_service_test.go`:

```go
func TestGetRandomQuestion_CategoryFiltering(t *testing.T) {
    client := SetupTestDatabase(t)
    defer CleanupTestData(t, client)

    service := NewQuestionService(client)

    // Now run actual test logic
    question, err := service.GetRandomQuestion(
        context.Background(),
        uuid.New(),
        "en",
        []uuid.UUID{},
    )

    if err != nil {
        t.Errorf("Expected no error, got: %v", err)
    }

    if question == nil {
        t.Error("Expected question, got nil")
    }
}
```

### Step 8: Run Tests

```bash
# Load test environment
export $(cat .env.test | xargs)

# Run full test suite (without -short)
go test -v ./internal/services/...

# Run with coverage
go test -coverprofile=coverage.out ./internal/services/...

# View coverage
go tool cover -html=coverage.out
```

### Useful Commands

```bash
# Start Supabase
supabase start

# Stop Supabase
supabase stop

# Reset database (wipes all data and re-runs migrations)
supabase db reset

# View database URL
supabase status

# Open Supabase Studio
open http://localhost:54323

# View logs
supabase logs

# Stop and remove all data
supabase stop --no-backup
```

---

## Option 2: Separate Supabase Test Project (ALTERNATIVE) ⚡

Use Supabase's free tier to create a dedicated test project.

### Pros
- ✅ Easy to set up
- ✅ Accessible from anywhere
- ✅ No local Docker required
- ✅ Free tier available

### Cons
- ⚠️ Requires internet connection
- ⚠️ Slower than local
- ⚠️ Shares data across test runs (need cleanup)

### Setup Steps

1. **Create Test Project**
   - Go to https://supabase.com/dashboard
   - Click "New Project"
   - Name it "couples-game-test"
   - Choose a region
   - Set strong database password

2. **Apply Schema**
   ```bash
   # Copy your schema
   # Open Supabase Dashboard → SQL Editor
   # Paste contents of sql/schema.sql
   # Click "Run"
   ```

3. **Seed Data**
   ```bash
   # In SQL Editor, paste sql/seed.sql
   # Click "Run"
   ```

4. **Get Credentials**
   - Go to Settings → API
   - Copy URL and Anon Key

5. **Create Test Environment**
   ```bash
   cat > .env.test << 'EOF'
   TEST_SUPABASE_URL=https://your-test-project.supabase.co
   TEST_SUPABASE_KEY=your-test-anon-key
   TEST_SUPABASE_SERVICE_KEY=your-test-service-role-key
   EOF
   ```

6. **Add Cleanup Between Tests**
   - Important: Clean up after each test run
   - Use `CleanupTestData()` helper from Option 1

---

## Option 3: Plain PostgreSQL with Docker (ADVANCED) 🔧

Use a plain PostgreSQL container if you don't need Supabase features.

### docker-compose.test.yml

```yaml
version: '3.8'

services:
  test-db:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
      POSTGRES_DB: couples_test
    ports:
      - "5433:5432"  # Use different port to avoid conflicts
    volumes:
      - ./sql/schema.sql:/docker-entrypoint-initdb.d/01-schema.sql
      - ./sql/seed.sql:/docker-entrypoint-initdb.d/02-seed.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U test"]
      interval: 5s
      timeout: 5s
      retries: 5

  test-runner:
    build: .
    depends_on:
      test-db:
        condition: service_healthy
    environment:
      - TEST_DB_URL=postgresql://test:test@test-db:5432/couples_test
    command: go test -v ./internal/services/...
```

### Setup Steps

```bash
# Start test database
docker-compose -f docker-compose.test.yml up test-db -d

# Wait for healthy
docker-compose -f docker-compose.test.yml ps

# Run tests
docker-compose -f docker-compose.test.yml up test-runner

# Stop
docker-compose -f docker-compose.test.yml down -v
```

### Note
This requires modifying your services to support plain PostgreSQL connections without Supabase client library.

---

## 🎯 Recommended Workflow

For development and testing, use **Option 1 (Supabase CLI)**:

```bash
# 1. One-time setup
supabase init
supabase start

# 2. Daily workflow
supabase db reset  # Clean slate
go test -v ./internal/services/...

# 3. View data in Studio
open http://localhost:54323

# 4. When done
supabase stop
```

---

## 🔄 CI/CD Integration

For GitHub Actions or other CI:

```yaml
# .github/workflows/test.yml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.22'

      - name: Setup Supabase CLI
        uses: supabase/setup-cli@v1

      - name: Start Supabase
        run: supabase start

      - name: Run tests
        run: |
          export TEST_SUPABASE_URL=http://localhost:54321
          export TEST_SUPABASE_KEY=${{ secrets.SUPABASE_ANON_KEY }}
          go test -v -coverprofile=coverage.out ./internal/services/...

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out

      - name: Stop Supabase
        run: supabase stop
```

---

## 📊 Comparison Table

| Feature | Supabase CLI | Supabase Cloud | Plain PostgreSQL |
|---------|--------------|----------------|------------------|
| Setup Time | 5 min | 10 min | 15 min |
| Speed | ⚡⚡⚡ Fast | ⚡ Moderate | ⚡⚡⚡ Fast |
| Isolation | ✅ Full | ⚠️ Shared | ✅ Full |
| Auth Features | ✅ Yes | ✅ Yes | ❌ No |
| Studio UI | ✅ Yes | ✅ Yes | ❌ No |
| Internet Required | ❌ No | ✅ Yes | ❌ No |
| Production Parity | ✅ 100% | ✅ 100% | ⚠️ 80% |
| Cost | Free | Free tier | Free |
| Recommended | ⭐⭐⭐ | ⭐⭐ | ⭐ |

---

## 🐛 Troubleshooting

### Supabase CLI Issues

**Problem**: `supabase start` fails
```bash
# Solution: Clean up old containers
docker ps -a | grep supabase
docker rm -f $(docker ps -a -q --filter="name=supabase")
supabase stop --no-backup
supabase start
```

**Problem**: Port already in use
```bash
# Solution: Stop conflicting services
lsof -ti:54321 | xargs kill -9  # API port
lsof -ti:54323 | xargs kill -9  # Studio port
lsof -ti:54322 | xargs kill -9  # DB port
```

**Problem**: Migrations fail
```bash
# Solution: Reset and reapply
supabase db reset
# Or manually via Studio
```

### Test Connection Issues

**Problem**: Tests can't connect to database
```bash
# Check Supabase is running
supabase status

# Verify connection
psql postgresql://postgres:postgres@localhost:54322/postgres

# Check environment variables
echo $TEST_SUPABASE_URL
echo $TEST_SUPABASE_KEY
```

---

## 📝 Next Steps

After setup:

1. ✅ Verify connection: `supabase status`
2. ✅ Run one test: `go test -v -run TestJoinStrings ./internal/services/...`
3. ✅ Run full suite: `go test -v ./internal/services/...`
4. ✅ Generate coverage: `go test -coverprofile=coverage.out ./internal/services/...`
5. ✅ View in browser: `go tool cover -html=coverage.out`
6. ✅ Aim for 80%+ coverage

---

## 🎉 Summary

**Recommended Setup**: Use Supabase CLI with Docker

```bash
# Quick start (5 minutes)
brew install supabase/tap/supabase
cd your-project
supabase init
supabase start
supabase db reset
go test -v ./internal/services/...
```

This gives you:
- ✅ Full Supabase environment locally
- ✅ Fast, isolated tests
- ✅ Production parity
- ✅ Beautiful Studio UI
- ✅ Zero cost

**Happy Testing!** 🚀
