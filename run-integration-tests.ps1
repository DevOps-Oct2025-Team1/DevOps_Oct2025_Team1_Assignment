# PowerShell script to run integration tests for backend services

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Backend Integration Tests Runner" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if Docker is running
Write-Host "Checking Docker..." -ForegroundColor Yellow
$dockerRunning = docker ps 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Docker is not running. Please start Docker Desktop." -ForegroundColor Red
    exit 1
}
Write-Host "[OK] Docker is running" -ForegroundColor Green
Write-Host ""

# Load environment variables from .env file
if (Test-Path .env) {
    Write-Host "Loading environment variables from .env..." -ForegroundColor Yellow
    Get-Content .env | ForEach-Object {
        if ($_ -match '^([^=]+)=(.*)$' -and $_ -notmatch '^#') {
            [Environment]::SetEnvironmentVariable($matches[1], $matches[2], "Process")
        }
    }
    Write-Host "[OK] Environment variables loaded" -ForegroundColor Green
} else {
    Write-Host "WARNING: .env file not found. Using test defaults." -ForegroundColor Yellow
    $env:DB_HOST = "localhost"
    $env:DB_PORT = "5432"
    $env:DB_USER = "postgres"
    $env:DB_PASSWORD = "postgres"
    $env:DB_NAME = "devops_db"
    $env:JWT_SECRET = "test-secret-key"
}
Write-Host ""

# Start database if not running
Write-Host "Starting PostgreSQL database..." -ForegroundColor Yellow
docker-compose -f docker-compose.yml -f docker-compose.test.yml up -d db
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to start database" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] Database started" -ForegroundColor Green
Write-Host ""

# Wait for database to be ready
Write-Host "Waiting for database to be ready..." -ForegroundColor Yellow
$maxAttempts = 30
$attempt = 0
$dbReady = $false

while ($attempt -lt $maxAttempts -and -not $dbReady) {
    $attempt++
    # Use pg_isready inside the db container for robust health check
    docker-compose -f docker-compose.yml -f docker-compose.test.yml exec -T db pg_isready -U $env:POSTGRES_USER -d $env:POSTGRES_DB 2>$null
    if ($LASTEXITCODE -eq 0) {
        $dbReady = $true
    } else {
        Start-Sleep -Seconds 1
        Write-Host "." -NoNewline
    }
}

Write-Host ""
if (-not $dbReady) {
    Write-Host "ERROR: Database did not become ready in time" -ForegroundColor Red
    docker-compose -f docker-compose.yml -f docker-compose.test.yml logs db
    exit 1
}
Write-Host "[OK] Database is ready" -ForegroundColor Green
Write-Host ""

# Override DB_HOST for local testing
$env:DB_HOST = "localhost"

# Run Auth Service Tests
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Running Auth Service Integration Tests" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$authServicePath = Join-Path "services" "auth.service"
Push-Location $authServicePath
go test -v
$authTestResult = $LASTEXITCODE
Pop-Location

Write-Host ""

# Run API Gateway Tests
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Running API Gateway Integration Tests" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$gatewayServicePath = Join-Path "services" "api-gateway.service"
Push-Location $gatewayServicePath
go test -v
$gatewayTestResult = $LASTEXITCODE
Pop-Location

Write-Host ""

# Summary
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Test Results Summary" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

if ($authTestResult -eq 0) {
    Write-Host "[PASS] Auth Service Tests: PASSED" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Auth Service Tests: FAILED" -ForegroundColor Red
}

if ($gatewayTestResult -eq 0) {
    Write-Host "[PASS] API Gateway Tests: PASSED" -ForegroundColor Green
} else {
    Write-Host "[FAIL] API Gateway Tests: FAILED" -ForegroundColor Red
}

Write-Host ""

# Clean up option
Write-Host "Do you want to stop the database? (y/N): " -NoNewline -ForegroundColor Yellow
$response = Read-Host
if ($response -eq 'y' -or $response -eq 'Y') {
    Write-Host "Stopping database..." -ForegroundColor Yellow
    docker-compose -f docker-compose.yml -f docker-compose.test.yml down
    Write-Host "[OK] Database stopped" -ForegroundColor Green
}

# Exit with appropriate code
if ($authTestResult -ne 0 -or $gatewayTestResult -ne 0) {
    exit 1
}

exit 0
