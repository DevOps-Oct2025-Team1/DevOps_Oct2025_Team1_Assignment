package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
		
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func TestValidateHandler(t *testing.T) {
	tests := []struct {
		name               string
		method             string
		body               string
		expectedStatus     int
		checkResponse      bool
		expectedContentType string
	}{
		{
			name:           "GET method not allowed",
			method:         http.MethodGet,
			body:           `{"token":"test"}`,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "PUT method not allowed",
			method:         http.MethodPut,
			body:           `{"token":"test"}`,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "DELETE method not allowed",
			method:         http.MethodDelete,
			body:           `{"token":"test"}`,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Invalid JSON syntax",
			method:         http.MethodPost,
			body:           `{"token": "invalid`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty body",
			method:         http.MethodPost,
			body:           "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Not JSON",
			method:         http.MethodPost,
			body:           "not json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:                "Valid POST request with token",
			method:              http.MethodPost,
			body:                `{"token":"some-token"}`,
			expectedStatus:      http.StatusOK,
			checkResponse:       true,
			expectedContentType: "application/json",
		},
		{
			name:                "Valid POST request with empty token",
			method:              http.MethodPost,
			body:                `{"token":""}`,
			expectedStatus:      http.StatusOK,
			checkResponse:       true,
			expectedContentType: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/validate", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			validateHandler(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.checkResponse {
				if contentType := rec.Header().Get("Content-Type"); contentType != tt.expectedContentType {
					t.Errorf("expected Content-Type '%s', got '%s'", tt.expectedContentType, contentType)
				}

				var resp ValidateResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				t.Logf("Response: Valid=%v, Role=%s", resp.Valid, resp.Role)
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	tests := []struct {
		name           string
		password       string
		testPassword   string
		expectedResult bool
	}{
		{
			name:           "Correct password",
			password:       "mySecurePassword123",
			testPassword:   "mySecurePassword123",
			expectedResult: true,
		},
		{
			name:           "Incorrect password",
			password:       "mySecurePassword123",
			testPassword:   "wrongPassword",
			expectedResult: false,
		},
		{
			name:           "Empty password",
			password:       "mySecurePassword123",
			testPassword:   "",
			expectedResult: false,
		},
		{
			name:           "Case sensitive check",
			password:       "Password123",
			testPassword:   "password123",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := bcrypt.GenerateFromPassword([]byte(tt.password), bcrypt.DefaultCost)
			if err != nil {
				t.Fatalf("Failed to generate hash: %v", err)
			}

			result := checkPassword(string(hash), tt.testPassword)

			if result != tt.expectedResult {
				t.Errorf("checkPassword() = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

func TestGenerateJWT(t *testing.T) {
	tests := []struct {
		name     string
		username string
		role     string
	}{
		{
			name:     "Admin user",
			username: "admin",
			role:     "admin",
		},
		{
			name:     "Regular user",
			username: "john_doe",
			role:     "user",
		},
		{
			name:     "Empty role",
			username: "test_user",
			role:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString, err := generateJWT(tt.username, tt.role)
			if err != nil {
				t.Fatalf("generateJWT() error = %v", err)
			}

			if tokenString == "" {
				t.Error("generateJWT() returned empty token")
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return jwtSecret, nil
			})

			if err != nil {
				t.Fatalf("Failed to parse generated token: %v", err)
			}

			if !token.Valid {
				t.Error("Generated token is not valid")
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if username, exists := claims["username"]; !exists || username != tt.username {
					t.Errorf("Username claim = %v, want %v", username, tt.username)
				}

				if role, exists := claims["role"]; !exists || role != tt.role {
					t.Errorf("Role claim = %v, want %v", role, tt.role)
				}
			} else {
				t.Error("Failed to parse token claims")
			}
		})
	}
}

func TestValidateJWT(t *testing.T) {
	tests := []struct {
		name          string
		setupToken    func() string
		expectedValid bool
		expectedRole  string
	}{
		{
			name: "Valid token with admin role",
			setupToken: func() string {
				token, _ := generateJWT("admin_user", "admin")
				return token
			},
			expectedValid: true,
			expectedRole:  "admin",
		},
		{
			name: "Valid token with user role",
			setupToken: func() string {
				token, _ := generateJWT("regular_user", "user")
				return token
			},
			expectedValid: true,
			expectedRole:  "user",
		},
		{
			name: "Invalid token - malformed",
			setupToken: func() string {
				return "hashire hashire umamusume"
			},
			expectedValid: false,
			expectedRole:  "",
		},
		{
			name: "Invalid token - empty string",
			setupToken: func() string {return ""},
			expectedValid: false,
			expectedRole:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString := tt.setupToken()
			valid, role := validateJWT(tokenString)

			if valid != tt.expectedValid {
				t.Errorf("validateJWT() valid = %v, want %v", valid, tt.expectedValid)
			}

			if role != tt.expectedRole {
				t.Errorf("validateJWT() role = %v, want %v", role, tt.expectedRole)
			}
		})
	}
}

func TestGenerateAndValidateJWT_Integration(t *testing.T) {
	username := "integration_test_user"
	role := "admin"

	token, err := generateJWT(username, role)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	valid, returnedRole := validateJWT(token)

	if !valid {
		t.Error("Generated token should be valid")
	}

	if returnedRole != role {
		t.Errorf("Role = %v, want %v", returnedRole, role)
	}
}

