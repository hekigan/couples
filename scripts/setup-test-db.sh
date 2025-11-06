#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Test Database Setup Script${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if Supabase CLI is installed
if ! command -v supabase &> /dev/null; then
    echo -e "${RED}❌ Supabase CLI not found${NC}"
    echo -e "${YELLOW}Installing Supabase CLI...${NC}"

    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        brew install supabase/tap/supabase
    else
        # Linux or other
        npm install -g supabase
    fi

    echo -e "${GREEN}✅ Supabase CLI installed${NC}"
else
    echo -e "${GREEN}✅ Supabase CLI found${NC}"
fi

# Check if Docker is running
if ! docker info &> /dev/null; then
    echo -e "${RED}❌ Docker is not running${NC}"
    echo -e "${YELLOW}Please start Docker Desktop and run this script again${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Docker is running${NC}"

# Initialize Supabase if not already done
if [ ! -d "supabase" ]; then
    echo -e "${YELLOW}Initializing Supabase...${NC}"
    supabase init
    echo -e "${GREEN}✅ Supabase initialized${NC}"
else
    echo -e "${GREEN}✅ Supabase already initialized${NC}"
fi

# Start Supabase
echo -e "${YELLOW}Starting Supabase services...${NC}"
supabase start

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Supabase started successfully${NC}"
else
    echo -e "${RED}❌ Failed to start Supabase${NC}"
    exit 1
fi

# Copy migration files
echo -e "${YELLOW}Setting up database schema...${NC}"
if [ -f "sql/schema.sql" ]; then
    cp sql/schema.sql supabase/migrations/00000000000000_init_schema.sql
    echo -e "${GREEN}✅ Schema file copied${NC}"
fi

if [ -f "sql/seed.sql" ]; then
    cp sql/seed.sql supabase/seed.sql
    echo -e "${GREEN}✅ Seed file copied${NC}"
fi

# Reset database to apply migrations
echo -e "${YELLOW}Applying migrations and seed data...${NC}"
supabase db reset

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Database reset and seeded${NC}"
else
    echo -e "${RED}❌ Failed to reset database${NC}"
    exit 1
fi

# Create .env.test file
if [ ! -f ".env.test" ]; then
    echo -e "${YELLOW}Creating .env.test file...${NC}"
    cp .env.test.example .env.test
    echo -e "${GREEN}✅ .env.test created${NC}"
else
    echo -e "${GREEN}✅ .env.test already exists${NC}"
fi

# Display connection info
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Connection Information${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}Supabase Studio:${NC} http://localhost:54323"
echo -e "${GREEN}API URL:${NC}         http://localhost:54321"
echo -e "${GREEN}DB URL:${NC}          postgresql://postgres:postgres@localhost:54322/postgres"
echo ""

# Display test commands
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Ready to Run Tests!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}Run all tests:${NC}"
echo -e "  go test -v ./internal/services/..."
echo ""
echo -e "${GREEN}Run with coverage:${NC}"
echo -e "  go test -coverprofile=coverage.out ./internal/services/..."
echo -e "  go tool cover -html=coverage.out"
echo ""
echo -e "${GREEN}Open Supabase Studio:${NC}"
echo -e "  open http://localhost:54323"
echo ""
echo -e "${GREEN}View Supabase status:${NC}"
echo -e "  supabase status"
echo ""
echo -e "${GREEN}Stop Supabase:${NC}"
echo -e "  supabase stop"
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}✨ Setup complete! Happy testing! 🚀${NC}"
echo -e "${BLUE}========================================${NC}"
