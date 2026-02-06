# Integration Testing Guide

This document explains how to run integration tests for the backend Go services.

## Overview

Integration tests have been created for:
- **Auth Service** (`auth.service/auth_integration_test.go`)
- **API Gateway** (`api-gateway.service/gateway_integration_test.go`)

## Prerequisites

1. **Docker and Docker Compose** installed
2. **Go 1.25+** installed
3. **PostgreSQL database** running (via Docker Compose)

## Running Integration Tests

### Step 1: Start the Database

First, start the PostgreSQL database using Docker Compose:

```bash
docker-compose up -d db
```

Wait for the database to be healthy (about 10-15 seconds):

```bash
docker-compose ps
```

### Step 2: Set Environment Variables

For Auth Service tests, set the required environment variables.

**Option A: Create test.env file from template**

First, copy the example file and update the placeholder values:

```bash
# Copy the template
cp services/auth.service/test.env.example services/auth.service/test.env

# Edit test.env and replace the placeholder values:
# - DB_USER (change from "test_user_CHANGE_ME")
# - DB_PASSWORD (change from "CHANGE_ME_test_password_min_16_chars")
# - JWT_SECRET (change from "CHANGE_ME_random_secret_min_32_chars_for_testing_only")
```

Then source the file:

```bash
# Linux/Mac
source services/auth.service/test.env

# Windows (PowerShell)
Get-Content services\auth.service\test.env | ForEach-Object {
    if ($_ -match '^([^=]+)=(.*)$') {
        [Environment]::SetEnvironmentVariable($matches[1], $matches[2])
    }
}
```

**Option B: Set environment variables manually**

**Windows PowerShell:**
```powershell
$env:DB_HOST="localhost"
$env:DB_PORT="5432"
$env:DB_USER="postgres"
$env:DB_PASSWORD="your_password_here"
$env:DB_NAME="devops_db"
$env:JWT_SECRET="your_jwt_secret_here"
```

**Linux/Mac:**
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=your_password_here
export DB_NAME=devops_db
export JWT_SECRET=your_jwt_secret_here
```

### Step 3: Run Auth Service Tests

Navigate to the auth service directory and run tests:

```bash
cd services/auth.service
go test -v
```

For verbose output with detailed logs:

```bash
go test -v -run TestLoginHandler_Success
```

To run specific tests:

```bash
go test -v -run TestCreateUserHandler_Success
go test -v -run TestLoginHandler
go test -v -run TestValidateHandler
```

### Step 4: Run API Gateway Tests

Navigate to the API Gateway directory:

```bash
cd services/api-gateway.service
go test -v
```

## Test Coverage

### Auth Service Tests

1. **Authentication Tests:**
   - ✅ User login with valid credentials
   - ✅ User login with invalid credentials
   - ✅ JWT validation (valid token)
   - ✅ JWT validation (invalid token)

2. **Admin Operations Tests:**
   - ✅ Get all users
   - ✅ Create new user
   - ✅ Create user with duplicate username (conflict)
   - ✅ Update user role
   - ✅ Delete user

3. **Database Tests:**
   - ✅ Database connection
   - ✅ Data persistence

### API Gateway Tests

1. **CORS Tests:**
   - ✅ CORS headers on regular requests
   - ✅ CORS preflight (OPTIONS) requests

2. **Routing Tests:**
   - ✅ Reverse proxy to auth service
   - ✅ Admin routes proxying
   - ✅ Invalid route handling

3. **Integration Tests:**
   - ✅ Full flow with CORS
   - ✅ Environment configuration

## Running Tests with Docker Compose

To run tests in a containerized environment:

1. Build and start all services:
```bash
docker-compose up -d
```

2. Execute tests inside the container:
```bash
# Auth service tests
docker-compose exec auth go test -v

# API Gateway tests (if needed)
docker-compose exec api-gateway go test -v
```

## Test Database

Tests use the same database as the development environment. Test data is cleaned up after each test using the `cleanupTestData` function.

### Database Schema

Ensure the database has the following schema:

```sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Troubleshooting

### Database Connection Issues

If tests fail with database connection errors:

1. Verify the database is running:
```bash
docker-compose ps db
```

2. Check database logs:
```bash
docker-compose logs db
```

3. Verify environment variables are set correctly:
```bash
# Windows
echo $env:DB_HOST
echo $env:DB_PORT

# Linux/Mac
echo $DB_HOST
echo $DB_PORT
```

### Port Conflicts

If port 5432 is already in use:

1. Check for existing PostgreSQL instances
2. Stop them or change the port in `docker-compose.yml`
3. Update environment variables accordingly

### JWT Secret Mismatch

Ensure the `JWT_SECRET` environment variable matches between:
- Your `.env` file
- Test environment variables
- Docker Compose environment

## CI/CD Integration

To integrate these tests into a CI/CD pipeline:

1. **GitHub Actions Example:**

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_DB: devops_db
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: postgres
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      
      - name: Run Auth Service Tests
        env:
          DB_HOST: localhost
          DB_PORT: 5432
          DB_USER: postgres
          DB_PASSWORD: postgres
          DB_NAME: devops_db
          JWT_SECRET: test-secret-key
        run: |
          cd services/auth.service
          go test -v
      
      - name: Run API Gateway Tests
        run: |
          cd services/api-gateway.service
          go test -v
```

## Coverage Reports

Generate test coverage reports:

```bash
# Auth service
cd services/auth.service
go test -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# API Gateway
cd services/api-gateway.service
go test -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## Next Steps

After running integration tests successfully:

1. ✅ Verify all tests pass
2. ✅ Review coverage reports
3. ✅ Add more test cases as needed
4. ✅ Integrate into CI/CD pipeline
5. ✅ Document any edge cases
6. ✅ Set up automated testing on pull requests
