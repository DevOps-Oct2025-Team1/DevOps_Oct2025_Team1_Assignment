package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestUnitExtractUserInfoFromToken_Valid tests extracting user info from a valid JWT token
func TestUnitExtractUserInfoFromToken_Valid(t *testing.T) {
	// Create a mock JWT payload
	payload := map[string]interface{}{
		"user_id":  float64(123),
		"username": "testuser",
	}
	payloadBytes, _ := json.Marshal(payload)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)

	// Create a mock JWT token (header.payload.signature)
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." + encodedPayload + ".signature"

	userID, username := extractUserInfoFromToken(token)

	if userID != 123 {
		t.Errorf("Expected user_id 123, got %d", userID)
	}

	if username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", username)
	}
}

// TestUnitExtractUserInfoFromToken_InvalidFormat tests extracting from invalid token format
func TestUnitExtractUserInfoFromToken_InvalidFormat(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"single part", "onlyonepart"},
		{"two parts", "header.payload"},
		{"four parts", "header.payload.signature.extra"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, username := extractUserInfoFromToken(tt.token)
			if userID != 0 {
				t.Errorf("Expected user_id 0, got %d", userID)
			}
			if username != "" {
				t.Errorf("Expected empty username, got '%s'", username)
			}
		})
	}
}

// TestUnitExtractUserInfoFromToken_InvalidBase64 tests extracting from token with invalid base64
func TestUnitExtractUserInfoFromToken_InvalidBase64(t *testing.T) {
	token := "header.!!!invalid-base64!!!.signature"

	userID, username := extractUserInfoFromToken(token)

	if userID != 0 {
		t.Errorf("Expected user_id 0, got %d", userID)
	}
	if username != "" {
		t.Errorf("Expected empty username, got '%s'", username)
	}
}

// TestUnitExtractUserInfoFromToken_InvalidJSON tests extracting from token with invalid JSON payload
func TestUnitExtractUserInfoFromToken_InvalidJSON(t *testing.T) {
	invalidJSON := base64.RawURLEncoding.EncodeToString([]byte("not valid json"))
	token := "header." + invalidJSON + ".signature"

	userID, username := extractUserInfoFromToken(token)

	if userID != 0 {
		t.Errorf("Expected user_id 0, got %d", userID)
	}
	if username != "" {
		t.Errorf("Expected empty username, got '%s'", username)
	}
}

// TestUnitExtractUserInfoFromToken_MissingUserID tests extracting from token without user_id
func TestUnitExtractUserInfoFromToken_MissingUserID(t *testing.T) {
	payload := map[string]interface{}{
		"username": "testuser",
	}
	payloadBytes, _ := json.Marshal(payload)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	token := "header." + encodedPayload + ".signature"

	userID, username := extractUserInfoFromToken(token)

	if userID != 0 {
		t.Errorf("Expected user_id 0, got %d", userID)
	}
	if username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", username)
	}
}

// TestUnitExtractUserInfoFromToken_MissingUsername tests extracting from token without username
func TestUnitExtractUserInfoFromToken_MissingUsername(t *testing.T) {
	payload := map[string]interface{}{
		"user_id": float64(456),
	}
	payloadBytes, _ := json.Marshal(payload)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	token := "header." + encodedPayload + ".signature"

	userID, username := extractUserInfoFromToken(token)

	if userID != 456 {
		t.Errorf("Expected user_id 456, got %d", userID)
	}
	if username != "" {
		t.Errorf("Expected empty username, got '%s'", username)
	}
}

// TestUnitGetUserID tests the getUserID context helper
func TestUnitGetUserID(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *http.Request
		expected int
	}{
		{
			name: "valid user ID in context",
			setup: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				ctx := context.WithValue(req.Context(), userIDKey, 123)
				return req.WithContext(ctx)
			},
			expected: 123,
		},
		{
			name: "no user ID in context",
			setup: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/", nil)
			},
			expected: 0,
		},
		{
			name: "wrong type in context",
			setup: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				ctx := context.WithValue(req.Context(), userIDKey, "not-an-int")
				return req.WithContext(ctx)
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setup()
			result := getUserID(req)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// TestUnitGetUsername tests the getUsername context helper
func TestUnitGetUsername(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *http.Request
		expected string
	}{
		{
			name: "valid username in context",
			setup: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				ctx := context.WithValue(req.Context(), usernameKey, "testuser")
				return req.WithContext(ctx)
			},
			expected: "testuser",
		},
		{
			name: "no username in context",
			setup: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/", nil)
			},
			expected: "",
		},
		{
			name: "wrong type in context",
			setup: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				ctx := context.WithValue(req.Context(), usernameKey, 123)
				return req.WithContext(ctx)
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setup()
			result := getUsername(req)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestUnitGetUserRole tests the getUserRole context helper
func TestUnitGetUserRole(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *http.Request
		expected string
	}{
		{
			name: "valid role in context",
			setup: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				ctx := context.WithValue(req.Context(), roleKey, "admin")
				return req.WithContext(ctx)
			},
			expected: "admin",
		},
		{
			name: "no role in context",
			setup: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/", nil)
			},
			expected: "",
		},
		{
			name: "wrong type in context",
			setup: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				ctx := context.WithValue(req.Context(), roleKey, 123)
				return req.WithContext(ctx)
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setup()
			result := getUserRole(req)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestUnitAuthMiddleware_MissingHeader tests authMiddleware with missing Authorization header
func TestUnitAuthMiddleware_MissingHeader(t *testing.T) {
	handler := authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

// TestUnitAuthMiddleware_InvalidFormat tests authMiddleware with invalid Authorization format
func TestUnitAuthMiddleware_InvalidFormat(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{"no bearer prefix", "token123"},
		{"wrong prefix", "Basic token123"},
		{"only bearer", "Bearer"},
		{"extra parts", "Bearer token extra"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := authMiddleware(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", tt.header)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
			}
		})
	}
}

// TestUnitAuthMiddleware_AuthServiceUnavailable tests authMiddleware when auth service is unavailable
func TestUnitAuthMiddleware_AuthServiceUnavailable(t *testing.T) {
	// Create a server and immediately close it to get a port that refuses connections
	closedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	closedServerURL := closedServer.URL
	closedServer.Close()

	originalURL := authServiceURL
	authServiceURL = closedServerURL
	defer func() { authServiceURL = originalURL }()

	handler := authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer sometoken")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, rr.Code)
	}
}

// TestUnitAuthMiddleware_InvalidToken tests authMiddleware when auth service returns invalid token
func TestUnitAuthMiddleware_InvalidToken(t *testing.T) {
	// Create a mock auth server that returns invalid token
	mockAuthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer mockAuthServer.Close()

	originalURL := authServiceURL
	authServiceURL = mockAuthServer.URL
	defer func() { authServiceURL = originalURL }()

	handler := authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalidtoken")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

// TestUnitAuthMiddleware_ValidNotTrue tests authMiddleware when auth service returns valid=false
func TestUnitAuthMiddleware_ValidNotTrue(t *testing.T) {
	// Create a mock auth server that returns valid=false
	mockAuthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := AuthValidateResponse{Valid: false}
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockAuthServer.Close()

	originalURL := authServiceURL
	authServiceURL = mockAuthServer.URL
	defer func() { authServiceURL = originalURL }()

	handler := authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer sometoken")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

// TestUnitAuthMiddleware_Success tests authMiddleware with valid token
func TestUnitAuthMiddleware_Success(t *testing.T) {
	// Create a mock auth server that returns valid token
	mockAuthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := AuthValidateResponse{
			Valid: true,
			Role:  "user",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockAuthServer.Close()

	originalURL := authServiceURL
	authServiceURL = mockAuthServer.URL
	defer func() { authServiceURL = originalURL }()

	// Create a valid JWT token with user info
	payload := map[string]interface{}{
		"user_id":  float64(123),
		"username": "testuser",
	}
	payloadBytes, _ := json.Marshal(payload)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." + encodedPayload + ".signature"

	var capturedUserID int
	var capturedUsername string
	var capturedRole string

	handler := authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = getUserID(r)
		capturedUsername = getUsername(r)
		capturedRole = getUserRole(r)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if capturedUserID != 123 {
		t.Errorf("Expected user_id 123, got %d", capturedUserID)
	}

	if capturedUsername != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", capturedUsername)
	}

	if capturedRole != "user" {
		t.Errorf("Expected role 'user', got '%s'", capturedRole)
	}
}

// TestUnitAuthMiddleware_TokenWithoutUserID tests authMiddleware with token missing user_id
func TestUnitAuthMiddleware_TokenWithoutUserID(t *testing.T) {
	// Create a mock auth server that returns valid token
	mockAuthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := AuthValidateResponse{
			Valid: true,
			Role:  "user",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockAuthServer.Close()

	originalURL := authServiceURL
	authServiceURL = mockAuthServer.URL
	defer func() { authServiceURL = originalURL }()

	// Create a JWT token without user_id
	payload := map[string]interface{}{
		"username": "testuser",
	}
	payloadBytes, _ := json.Marshal(payload)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." + encodedPayload + ".signature"

	handler := authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}
