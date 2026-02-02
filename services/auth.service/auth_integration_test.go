package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// Setup test database connection
func setupTestDB(t *testing.T) *sql.DB {
	// Use test database or main database for integration tests
	connStr := "host=" + getEnv("DB_HOST", "localhost") +
		" port=" + getEnv("DB_PORT", "5432") +
		" user=" + getEnv("DB_USER", "postgres") +
		" password=" + getEnv("DB_PASSWORD", "postgres") +
		" dbname=" + getEnv("DB_NAME", "devops_db") +
		" sslmode=disable"

	testDB, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := testDB.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	return testDB
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// Clean up test data
func cleanupTestData(t *testing.T, testDB *sql.DB, username string) {
	_, err := testDB.Exec("DELETE FROM users WHERE username = $1", username)
	if err != nil {
		t.Logf("Warning: Failed to cleanup test user %s: %v", username, err)
	}
}

// Test: User Login - Success
func TestLoginHandler_Success(t *testing.T) {
	// Setup
	testDB := setupTestDB(t)
	defer testDB.Close()

	db = testDB
	jwtSecret = []byte(getEnv("JWT_SECRET", "test-secret-key"))

	// Create test user
	testUsername := "test_login_user"
	testPassword := "testpassword123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)

	_, err := testDB.Exec(
		"INSERT INTO users (username, password_hash, role) VALUES ($1, $2, $3) ON CONFLICT (username) DO NOTHING",
		testUsername, string(hashedPassword), "user",
	)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	defer cleanupTestData(t, testDB, testUsername)

	// Create request
	loginReq := LoginRequest{
		Username: testUsername,
		Password: testPassword,
	}
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	rr := httptest.NewRecorder()
	loginHandler(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var loginResp LoginResponse
	err = json.NewDecoder(rr.Body).Decode(&loginResp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if loginResp.Token == "" {
		t.Error("Expected token in response, got empty string")
	}
}

// Test: User Login - Invalid Credentials
func TestLoginHandler_InvalidCredentials(t *testing.T) {
	// Setup
	testDB := setupTestDB(t)
	defer testDB.Close()

	db = testDB
	jwtSecret = []byte(getEnv("JWT_SECRET", "test-secret-key"))

	// Create request with wrong password
	loginReq := LoginRequest{
		Username: "nonexistent_user",
		Password: "wrongpassword",
	}
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	rr := httptest.NewRecorder()
	loginHandler(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

// Test: JWT Validation - Valid Token
func TestValidateHandler_ValidToken(t *testing.T) {
	// Setup
	jwtSecret = []byte(getEnv("JWT_SECRET", "test-secret-key"))

	// Generate valid token
	token, err := generateJWT("testuser", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Create request
	validateReq := ValidateRequest{Token: token}
	body, _ := json.Marshal(validateReq)
	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	rr := httptest.NewRecorder()
	validateHandler(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var validateResp ValidateResponse
	err = json.NewDecoder(rr.Body).Decode(&validateResp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !validateResp.Valid {
		t.Error("Expected valid token, got invalid")
	}

	if validateResp.Role != "admin" {
		t.Errorf("Expected role 'admin', got '%s'", validateResp.Role)
	}
}

// Test: JWT Validation - Invalid Token
func TestValidateHandler_InvalidToken(t *testing.T) {
	// Setup
	jwtSecret = []byte(getEnv("JWT_SECRET", "test-secret-key"))

	// Create request with invalid token
	validateReq := ValidateRequest{Token: "invalid.token.here"}
	body, _ := json.Marshal(validateReq)
	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	rr := httptest.NewRecorder()
	validateHandler(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var validateResp ValidateResponse
	err := json.NewDecoder(rr.Body).Decode(&validateResp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if validateResp.Valid {
		t.Error("Expected invalid token, got valid")
	}
}

// Test: Get All Users - Success (Admin)
func TestGetUsersHandler_Success(t *testing.T) {
	// Setup
	testDB := setupTestDB(t)
	defer testDB.Close()

	db = testDB
	jwtSecret = []byte(getEnv("JWT_SECRET", "test-secret-key"))

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/users", nil)

	// Execute
	rr := httptest.NewRecorder()
	getUsersHandler(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var users []User
	err := json.NewDecoder(rr.Body).Decode(&users)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Success - users list returned (may be empty)
	t.Logf("Retrieved %d users from database", len(users))
}

// Test: Create User - Success
func TestCreateUserHandler_Success(t *testing.T) {
	// Setup
	testDB := setupTestDB(t)
	defer testDB.Close()

	db = testDB

	testUsername := "test_create_user_" + time.Now().Format("20060102150405")
	defer cleanupTestData(t, testDB, testUsername)

	// Create request
	createReq := CreateUserRequest{
		Username: testUsername,
		Password: "password123",
		Role:     "user",
	}
	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	rr := httptest.NewRecorder()
	createUserHandler(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v, body: %s",
			status, http.StatusCreated, rr.Body.String())
	}

	var user User
	err := json.NewDecoder(rr.Body).Decode(&user)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if user.Username != testUsername {
		t.Errorf("Expected username '%s', got '%s'", testUsername, user.Username)
	}

	if user.Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", user.Role)
	}
}

// Test: Create User - Duplicate Username
func TestCreateUserHandler_DuplicateUsername(t *testing.T) {
	// Setup
	testDB := setupTestDB(t)
	defer testDB.Close()

	db = testDB

	testUsername := "test_duplicate_user"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	// Create initial user
	_, err := testDB.Exec(
		"INSERT INTO users (username, password_hash, role) VALUES ($1, $2, $3) ON CONFLICT (username) DO NOTHING",
		testUsername, string(hashedPassword), "user",
	)
	if err != nil {
		t.Fatalf("Failed to create initial test user: %v", err)
	}
	defer cleanupTestData(t, testDB, testUsername)

	// Try to create duplicate
	createReq := CreateUserRequest{
		Username: testUsername,
		Password: "password123",
		Role:     "user",
	}
	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	rr := httptest.NewRecorder()
	createUserHandler(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusConflict {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusConflict)
	}
}

// Test: Update User - Success
func TestEditUserHandler_Success(t *testing.T) {
	// Setup
	testDB := setupTestDB(t)
	defer testDB.Close()

	db = testDB

	// Create test user
	testUsername := "test_edit_user"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	var userID int
	err := testDB.QueryRow(
		"INSERT INTO users (username, password_hash, role) VALUES ($1, $2, $3) ON CONFLICT (username) DO UPDATE SET role = $3 RETURNING id",
		testUsername, string(hashedPassword), "user",
	).Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	defer cleanupTestData(t, testDB, testUsername)

	// Create update request
	updateReq := UpdateUserRequest{Role: "premium"}
	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPut, "/users/"+strconv.Itoa(userID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	rr := httptest.NewRecorder()
	editUserHandler(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v, body: %s",
			status, http.StatusOK, rr.Body.String())
	}

	// Verify in database
	var role string
	err = testDB.QueryRow("SELECT role FROM users WHERE id = $1", userID).Scan(&role)
	if err != nil {
		t.Fatalf("Failed to query updated user: %v", err)
	}

	if role != "premium" {
		t.Errorf("Expected role 'premium', got '%s'", role)
	}
}

// Test: Delete User - Success
func TestDeleteUserHandler_Success(t *testing.T) {
	// Setup
	testDB := setupTestDB(t)
	defer testDB.Close()

	db = testDB

	// Create test user
	testUsername := "test_delete_user_" + time.Now().Format("20060102150405")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	var userID int
	err := testDB.QueryRow(
		"INSERT INTO users (username, password_hash, role) VALUES ($1, $2, $3) RETURNING id",
		testUsername, string(hashedPassword), "user",
	).Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create delete request
	req := httptest.NewRequest(http.MethodDelete, "/users/"+strconv.Itoa(userID), nil)

	// Execute
	rr := httptest.NewRecorder()
	deleteUserHandler(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v, body: %s",
			status, http.StatusOK, rr.Body.String())
	}

	// Verify deletion in database
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM users WHERE id = $1", userID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to verify deletion: %v", err)
	}

	if count != 0 {
		t.Error("User was not deleted from database")
	}
}

// Test: Database Connection
func TestDatabaseConnection(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	err := testDB.Ping()
	if err != nil {
		t.Fatalf("Database connection failed: %v", err)
	}
}
