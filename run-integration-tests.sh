#!/bin/bash
# Bash script to run integration tests for backend services
# Compatible with Linux/macOS and GitHub Actions

# Color codes for output
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${CYAN}========================================"
echo -e "Backend Integration Tests Runner"
echo -e "========================================${NC}"
echo ""

# Check if Docker is running
echo -e "${YELLOW}Checking Docker...${NC}"
if ! docker ps > /dev/null 2>&1; then
    echo -e "${RED}ERROR: Docker is not running. Please start Docker.${NC}"
    exit 1
fi
echo -e "${GREEN}[OK] Docker is running${NC}"
echo ""

# Load environment variables from .env file
if [ -f .env ]; then
    echo -e "${YELLOW}Loading environment variables from .env...${NC}"
    set -a
    source .env
    set +a
    echo -e "${GREEN}[OK] Environment variables loaded${NC}"
else
    echo -e "${YELLOW}WARNING: .env file not found. Using test defaults.${NC}"
    export POSTGRES_HOST="db"
    export POSTGRES_PORT="5432"
    export POSTGRES_USER="postgres"
    export POSTGRES_PASSWORD="postgres"
    export POSTGRES_DB="devops_db"
    export JWT_SECRET="test-secret-key"
fi
echo ""

# Start database if not running
echo -e "${YELLOW}Starting PostgreSQL database...${NC}"
if ! docker-compose -f docker-compose.yml -f docker-compose.test.yml up -d db; then
    echo -e "${RED}ERROR: Failed to start database${NC}"
    exit 1
fi
echo -e "${GREEN}[OK] Database started${NC}"
echo ""

# Wait for database to be ready
echo -e "${YELLOW}Waiting for database to be ready...${NC}"
max_attempts=30
attempt=0
db_ready=false

while [ $attempt -lt $max_attempts ] && [ "$db_ready" = false ]; do
    attempt=$((attempt + 1))
    # Use pg_isready inside the db container for robust health check
    if docker-compose -f docker-compose.yml -f docker-compose.test.yml exec -T db pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB" > /dev/null 2>&1; then
        db_ready=true
    else
        sleep 1
        echo -n "."
    fi
done

echo ""
if [ "$db_ready" = false ]; then
    echo -e "${RED}ERROR: Database did not become ready in time${NC}"
    docker-compose -f docker-compose.yml -f docker-compose.test.yml logs db
    exit 1
fi
echo -e "${GREEN}[OK] Database is ready${NC}"
echo ""

# Override DB_HOST for local testing
export DB_HOST="localhost"
export DB_USER="$POSTGRES_USER"
export DB_PASSWORD="$POSTGRES_PASSWORD"
export DB_NAME="$POSTGRES_DB"
export DB_PORT="$POSTGRES_PORT"

# Run Auth Service Tests
echo -e "${CYAN}========================================"
echo -e "Running Auth Service Integration Tests"
echo -e "========================================${NC}"
echo ""

cd services/auth.service || exit 1
go test -v
auth_test_result=$?
cd ../.. || exit 1

echo ""

# Run API Gateway Tests
echo -e "${CYAN}========================================"
echo -e "Running API Gateway Integration Tests"
echo -e "========================================${NC}"
echo ""

cd services/api-gateway.service || exit 1
go test -v
gateway_test_result=$?
cd ../.. || exit 1

echo ""

# Summary
echo -e "${CYAN}========================================"
echo -e "Test Results Summary"
echo -e "========================================${NC}"

if [ $auth_test_result -eq 0 ]; then
    echo -e "${GREEN}[PASS] Auth Service Tests: PASSED${NC}"
else
    echo -e "${RED}[FAIL] Auth Service Tests: FAILED${NC}"
fi

if [ $gateway_test_result -eq 0 ]; then
    echo -e "${GREEN}[PASS] API Gateway Tests: PASSED${NC}"
else
    echo -e "${RED}[FAIL] API Gateway Tests: FAILED${NC}"
fi

echo ""

# Clean up option
echo -en "${YELLOW}Do you want to stop the database? (y/N): ${NC}"
read -r response
if [[ "$response" =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Stopping database...${NC}"
    docker-compose -f docker-compose.yml -f docker-compose.test.yml down
    echo -e "${GREEN}[OK] Database stopped${NC}"
fi

# Exit with appropriate code
if [ $auth_test_result -ne 0 ] || [ $gateway_test_result -ne 0 ]; then
    exit 1
fi

exit 0
