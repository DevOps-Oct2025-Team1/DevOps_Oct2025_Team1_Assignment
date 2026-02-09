package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// setupHandlersTestDB initializes a test database for handlers
func setupHandlersTestDB(t *testing.T) (sqlmock.Sqlmock, func()) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}

	originalDB := db
	db = mockDB

	cleanup := func() {
		mockDB.Close()
		db = originalDB
	}

	return mock, cleanup
}

// createAuthenticatedRequest creates a request with user context
func createAuthenticatedRequest(method, path string, body []byte, userID int, username string) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewBuffer(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	ctx := context.WithValue(req.Context(), userIDKey, userID)
	ctx = context.WithValue(ctx, usernameKey, username)
	return req.WithContext(ctx)
}

// Health Handler Tests

// TestUnitHealthHandler_Success tests the health handler
func TestUnitHealthHandler_Success(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	healthHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp HealthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", resp.Status)
	}

	if resp.Service != "prompt-manager" {
		t.Errorf("Expected service 'prompt-manager', got '%s'", resp.Service)
	}
}

// TestUnitHealthHandler_WrongMethod tests the health handler with wrong method
func TestUnitHealthHandler_WrongMethod(t *testing.T) {
	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/health", nil)
			rr := httptest.NewRecorder()

			healthHandler(rr, req)

			if rr.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
			}
		})
	}
}

// Models Handler Tests

// TestUnitModelsHandler_Success tests the models handler
func TestUnitModelsHandler_Success(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/models", nil)
	rr := httptest.NewRecorder()

	modelsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("Expected data to be a map")
	}

	models, ok := data["models"].([]any)
	if !ok {
		t.Fatalf("Expected models to be an array")
	}

	if len(models) == 0 {
		t.Errorf("Expected models, got none")
	}
}

// TestUnitModelsHandler_WrongMethod tests the models handler with wrong method
func TestUnitModelsHandler_WrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/models", nil)
	rr := httptest.NewRecorder()

	modelsHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

// CreateChat Handler Tests

// TestUnitCreateChatHandler_Success tests creating a chat successfully
func TestUnitCreateChatHandler_Success(t *testing.T) {
	mock, cleanup := setupHandlersTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`INSERT INTO chats \(user_id, model, title\) VALUES \(\$1, \$2, \$3\) RETURNING id, created_at, updated_at`).
		WithArgs(1, "gemma3", "New Chat").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("1", time.Now(), time.Now()))

	body := CreateChatRequest{Model: "gemma3"}
	bodyBytes, _ := json.Marshal(body)

	req := createAuthenticatedRequest(http.MethodPost, "/chats", bodyBytes, 1, "testuser")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	createChatHandler(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, rr.Code)
	}
}

// TestUnitCreateChatHandler_DefaultModel tests creating a chat with default model
func TestUnitCreateChatHandler_DefaultModel(t *testing.T) {
	mock, cleanup := setupHandlersTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`INSERT INTO chats \(user_id, model, title\) VALUES \(\$1, \$2, \$3\) RETURNING id, created_at, updated_at`).
		WithArgs(1, "gemma3", "New Chat").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("1", time.Now(), time.Now()))

	body := CreateChatRequest{} // No model specified
	bodyBytes, _ := json.Marshal(body)

	req := createAuthenticatedRequest(http.MethodPost, "/chats", bodyBytes, 1, "testuser")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	createChatHandler(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, rr.Code)
	}
}

// TestUnitCreateChatHandler_InvalidModel tests creating a chat with invalid model
func TestUnitCreateChatHandler_InvalidModel(t *testing.T) {
	body := CreateChatRequest{Model: "invalid-model"}
	bodyBytes, _ := json.Marshal(body)

	req := createAuthenticatedRequest(http.MethodPost, "/chats", bodyBytes, 1, "testuser")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	createChatHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

// TestUnitCreateChatHandler_Unauthorized tests creating a chat without auth
func TestUnitCreateChatHandler_Unauthorized(t *testing.T) {
	body := CreateChatRequest{Model: "gemma3"}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/chats", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	createChatHandler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

// TestUnitCreateChatHandler_InvalidBody tests creating a chat with invalid JSON body
func TestUnitCreateChatHandler_InvalidBody(t *testing.T) {
	req := createAuthenticatedRequest(http.MethodPost, "/chats", []byte("invalid json"), 1, "testuser")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	createChatHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

// TestUnitCreateChatHandler_WrongMethod tests creating a chat with wrong method
func TestUnitCreateChatHandler_WrongMethod(t *testing.T) {
	body := CreateChatRequest{Model: "gemma3"}
	bodyBytes, _ := json.Marshal(body)

	req := createAuthenticatedRequest(http.MethodGet, "/chats", bodyBytes, 1, "testuser")
	rr := httptest.NewRecorder()

	createChatHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

// TestUnitCreateChatHandler_DBError tests creating a chat with database error
func TestUnitCreateChatHandler_DBError(t *testing.T) {
	mock, cleanup := setupHandlersTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`INSERT INTO chats \(user_id, model, title\) VALUES \(\$1, \$2, \$3\) RETURNING id, created_at, updated_at`).
		WithArgs(1, "gemma3", "New Chat").
		WillReturnError(sql.ErrConnDone)

	body := CreateChatRequest{Model: "gemma3"}
	bodyBytes, _ := json.Marshal(body)

	req := createAuthenticatedRequest(http.MethodPost, "/chats", bodyBytes, 1, "testuser")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	createChatHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

// ListChats Handler Tests

// TestUnitListChatsHandler_Success tests listing chats successfully
func TestUnitListChatsHandler_Success(t *testing.T) {
	mock, cleanup := setupHandlersTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM chats WHERE user_id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery(`SELECT id, user_id, title, model, created_at, updated_at FROM chats WHERE user_id = \$1 ORDER BY updated_at DESC LIMIT \$2 OFFSET \$3`).
		WithArgs(1, 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "model", "created_at", "updated_at"}).
			AddRow("1", 1, "Chat 1", "gemma3", time.Now(), time.Now()).
			AddRow("2", 1, "Chat 2", "qwen3", time.Now(), time.Now()))

	req := createAuthenticatedRequest(http.MethodGet, "/chats", nil, 1, "testuser")
	rr := httptest.NewRecorder()

	listChatsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp ChatListResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(resp.Chats) != 2 {
		t.Errorf("Expected 2 chats, got %d", len(resp.Chats))
	}

	if resp.TotalCount != 2 {
		t.Errorf("Expected total count 2, got %d", resp.TotalCount)
	}
}

// TestUnitListChatsHandler_Pagination tests listing chats with pagination params
func TestUnitListChatsHandler_Pagination(t *testing.T) {
	mock, cleanup := setupHandlersTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM chats WHERE user_id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(50))

	mock.ExpectQuery(`SELECT id, user_id, title, model, created_at, updated_at FROM chats WHERE user_id = \$1 ORDER BY updated_at DESC LIMIT \$2 OFFSET \$3`).
		WithArgs(1, 10, 20). // page 3, pageSize 10 -> offset 20
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "model", "created_at", "updated_at"}).
			AddRow("21", 1, "Chat 21", "gemma3", time.Now(), time.Now()))

	req := createAuthenticatedRequest(http.MethodGet, "/chats?page=3&page_size=10", nil, 1, "testuser")
	rr := httptest.NewRecorder()

	listChatsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp ChatListResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Page != 3 {
		t.Errorf("Expected page 3, got %d", resp.Page)
	}

	if resp.PageSize != 10 {
		t.Errorf("Expected page size 10, got %d", resp.PageSize)
	}
}

// TestUnitListChatsHandler_Unauthorized tests listing chats without auth
func TestUnitListChatsHandler_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/chats", nil)
	rr := httptest.NewRecorder()

	listChatsHandler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

// TestUnitListChatsHandler_WrongMethod tests listing chats with wrong method
func TestUnitListChatsHandler_WrongMethod(t *testing.T) {
	req := createAuthenticatedRequest(http.MethodPost, "/chats", nil, 1, "testuser")
	rr := httptest.NewRecorder()

	listChatsHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

// TestUnitListChatsHandler_DBError tests listing chats with database error
func TestUnitListChatsHandler_DBError(t *testing.T) {
	mock, cleanup := setupHandlersTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM chats WHERE user_id = \$1`).
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	req := createAuthenticatedRequest(http.MethodGet, "/chats", nil, 1, "testuser")
	rr := httptest.NewRecorder()

	listChatsHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

// GetChat Handler Tests

// TestUnitGetChatHandler_Success tests getting a chat successfully
func TestUnitGetChatHandler_Success(t *testing.T) {
	mock, cleanup := setupHandlersTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, user_id, title, model, created_at, updated_at FROM chats WHERE id = \$1 AND user_id = \$2`).
		WithArgs("123", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "model", "created_at", "updated_at"}).
			AddRow("123", 1, "Test Chat", "gemma3", time.Now(), time.Now()))

	mock.ExpectQuery(`SELECT id, chat_id, role, content, status, error_message, tokens_used, created_at FROM messages WHERE chat_id = \$1 ORDER BY created_at ASC`).
		WithArgs("123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "chat_id", "role", "content", "status", "error_message", "tokens_used", "created_at"}).
			AddRow("1", "123", RoleUser, "Hello", StatusCompleted, nil, nil, time.Now()))

	req := createAuthenticatedRequest(http.MethodGet, "/chats/123", nil, 1, "testuser")
	rr := httptest.NewRecorder()

	getChatHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp ChatWithMessages
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Chat.ID != "123" {
		t.Errorf("Expected chat ID '123', got '%s'", resp.Chat.ID)
	}

	if len(resp.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(resp.Messages))
	}
}

// TestUnitGetChatHandler_NotFound tests getting a non-existent chat
func TestUnitGetChatHandler_NotFound(t *testing.T) {
	mock, cleanup := setupHandlersTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, user_id, title, model, created_at, updated_at FROM chats WHERE id = \$1 AND user_id = \$2`).
		WithArgs("999", 1).
		WillReturnError(sql.ErrNoRows)

	req := createAuthenticatedRequest(http.MethodGet, "/chats/999", nil, 1, "testuser")
	rr := httptest.NewRecorder()

	getChatHandler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

// TestUnitGetChatHandler_Unauthorized tests getting a chat without auth
func TestUnitGetChatHandler_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/chats/123", nil)
	rr := httptest.NewRecorder()

	getChatHandler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

// TestUnitGetChatHandler_WrongMethod tests getting a chat with wrong method
func TestUnitGetChatHandler_WrongMethod(t *testing.T) {
	req := createAuthenticatedRequest(http.MethodPost, "/chats/123", nil, 1, "testuser")
	rr := httptest.NewRecorder()

	getChatHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

// TestUnitGetChatHandler_ChatError tests getting a chat where the chats has a database error
func TestUnitGetChatHandler_ChatError(t *testing.T) {
	mock, cleanup := setupHandlersTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, user_id, title, model, created_at, updated_at FROM chats WHERE id = \$1 AND user_id = \$2`).
		WithArgs("123", 1).
		WillReturnError(sql.ErrConnDone)
	mock.ExpectQuery(`SELECT id, chat_id, role, content, status, error_message, tokens_used, created_at FROM messages WHERE chat_id = \$1 ORDER BY created_at ASC`).
		WithArgs("123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "chat_id", "role", "content", "status", "error_message", "tokens_used", "created_at"}).
			AddRow("1", "123", RoleUser, "Hello", StatusCompleted, nil, nil, time.Now()))

	req := createAuthenticatedRequest(http.MethodGet, "/chats/123", nil, 1, "testuser")
	rr := httptest.NewRecorder()

	getChatHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

// TestUnitGetChatHandler_MessagesError tests getting a chat where the messages have a database error
func TestUnitGetChatHandler_MessagesError(t *testing.T) {
	mock, cleanup := setupHandlersTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, user_id, title, model, created_at, updated_at FROM chats WHERE id = \$1 AND user_id = \$2`).
		WithArgs("123", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "model", "created_at", "updated_at"}).
			AddRow("123", 1, "Test Chat", "gemma3", time.Now(), time.Now()))
	mock.ExpectQuery(`SELECT id, chat_id, role, content, status, error_message, tokens_used, created_at FROM messages WHERE chat_id = \$1 ORDER BY created_at ASC`).
		WithArgs("123").
		WillReturnError(sql.ErrConnDone)

	req := createAuthenticatedRequest(http.MethodGet, "/chats/123", nil, 1, "testuser")
	rr := httptest.NewRecorder()

	getChatHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

// DeleteChat Handler Tests

// TestUnitUnitDeleteChatHandler_Success tests deleting a chat successfully
func TestUnitUnitDeleteChatHandler_Success(t *testing.T) {
	mock, cleanup := setupHandlersTestDB(t)
	defer cleanup()

	mock.ExpectExec(`DELETE FROM chats WHERE id = \$1 AND user_id = \$2`).
		WithArgs("123", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := createAuthenticatedRequest(http.MethodDelete, "/chats/123", nil, 1, "testuser")
	rr := httptest.NewRecorder()

	deleteChatHandler(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, rr.Code)
	}
}

// TestUnitDeleteChatHandler_NotFound tests deleting a non-existent chat
func TestUnitDeleteChatHandler_NotFound(t *testing.T) {
	mock, cleanup := setupHandlersTestDB(t)
	defer cleanup()

	mock.ExpectExec(`DELETE FROM chats WHERE id = \$1 AND user_id = \$2`).
		WithArgs("999", 1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	req := createAuthenticatedRequest(http.MethodDelete, "/chats/999", nil, 1, "testuser")
	rr := httptest.NewRecorder()

	deleteChatHandler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

// TestUnitDeleteChatHandler_Unauthorized tests deleting a chat without auth
func TestUnitDeleteChatHandler_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/chats/123", nil)
	rr := httptest.NewRecorder()

	deleteChatHandler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

// TestUnitDeleteChatHandler_WrongMethod tests deleting a chat with wrong method
func TestUnitDeleteChatHandler_WrongMethod(t *testing.T) {
	req := createAuthenticatedRequest(http.MethodGet, "/chats/123", nil, 1, "testuser")
	rr := httptest.NewRecorder()

	deleteChatHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

// TestUnitDeleteChatHandler_DBError tests deleting a chat with database error
func TestUnitDeleteChatHandler_DBError(t *testing.T) {
	mock, cleanup := setupHandlersTestDB(t)
	defer cleanup()

	mock.ExpectExec(`DELETE FROM chats WHERE id = \$1 AND user_id = \$2`).
		WithArgs("123", 1).
		WillReturnError(sql.ErrConnDone)

	req := createAuthenticatedRequest(http.MethodDelete, "/chats/123", nil, 1, "testuser")
	rr := httptest.NewRecorder()

	deleteChatHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

// SendMessage Handler Tests

// TestUnitSendMessageHandler_Success tests sending a message successfully
func TestUnitSendMessageHandler_Success(t *testing.T) {
	mock, cleanup := setupHandlersTestDB(t)
	defer cleanup()

	// Store original pubsub vars and set to nil to skip publishing
	originalPubsubClient := pubsubClient
	originalTopic := topic
	pubsubClient = nil
	topic = nil
	defer func() {
		pubsubClient = originalPubsubClient
		topic = originalTopic
	}()

	// Expect GetChat query
	mock.ExpectQuery(`SELECT id, user_id, title, model, created_at, updated_at FROM chats WHERE id = \$1 AND user_id = \$2`).
		WithArgs("123", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "model", "created_at", "updated_at"}).
			AddRow("123", 1, "Test Chat", "gemma3", time.Now(), time.Now()))

	// Expect CreateMessage query
	mock.ExpectQuery(`INSERT INTO messages \(chat_id, role, content, status\) VALUES \(\$1, \$2, \$3, \$4\) RETURNING id, created_at`).
		WithArgs("123", RoleUser, "Hello there", StatusCompleted).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow("1", time.Now()))

	// Expect chat update timestamp
	mock.ExpectExec(`UPDATE chats SET updated_at = CURRENT_TIMESTAMP WHERE id = \$1`).
		WithArgs("123").
		WillReturnResult(sqlmock.NewResult(0, 1))

	body := SendMessageRequest{Content: "Hello there"}
	bodyBytes, _ := json.Marshal(body)

	req := createAuthenticatedRequest(http.MethodPost, "/chats/123/messages", bodyBytes, 1, "testuser")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	sendMessageHandler(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}
}

// TestUnitSendMessageHandler_ChatNotFound tests sending a message to non-existent chat
func TestUnitSendMessageHandler_ChatNotFound(t *testing.T) {
	mock, cleanup := setupHandlersTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, user_id, title, model, created_at, updated_at FROM chats WHERE id = \$1 AND user_id = \$2`).
		WithArgs("999", 1).
		WillReturnError(sql.ErrNoRows)

	body := SendMessageRequest{Content: "Hello"}
	bodyBytes, _ := json.Marshal(body)

	req := createAuthenticatedRequest(http.MethodPost, "/chats/999/messages", bodyBytes, 1, "testuser")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	sendMessageHandler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

// TestUnitSendMessageHandler_EmptyContent tests sending a message with empty content
func TestUnitSendMessageHandler_EmptyContent(t *testing.T) {
	mock, cleanup := setupHandlersTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, user_id, title, model, created_at, updated_at FROM chats WHERE id = \$1 AND user_id = \$2`).
		WithArgs("123", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "model", "created_at", "updated_at"}).
			AddRow("123", 1, "Test Chat", "gemma3", time.Now(), time.Now()))

	body := SendMessageRequest{Content: ""}
	bodyBytes, _ := json.Marshal(body)

	req := createAuthenticatedRequest(http.MethodPost, "/chats/123/messages", bodyBytes, 1, "testuser")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	sendMessageHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

// TestUnitSendMessageHandler_Unauthorized tests sending a message without auth
func TestUnitSendMessageHandler_Unauthorized(t *testing.T) {
	body := SendMessageRequest{Content: "Hello"}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/chats/123/messages", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	sendMessageHandler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

// TestUnitSendMessageHandler_WrongMethod tests sending a message with wrong method
func TestUnitSendMessageHandler_WrongMethod(t *testing.T) {
	req := createAuthenticatedRequest(http.MethodGet, "/chats/123/messages", nil, 1, "testuser")
	rr := httptest.NewRecorder()

	sendMessageHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

// TestUnitSendMessageHandler_InvalidBody tests sending a message with invalid JSON body
func TestUnitSendMessageHandler_InvalidBody(t *testing.T) {
	mock, cleanup := setupHandlersTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, user_id, title, model, created_at, updated_at FROM chats WHERE id = \$1 AND user_id = \$2`).
		WithArgs("123", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "model", "created_at", "updated_at"}).
			AddRow("123", 1, "Test Chat", "gemma3", time.Now(), time.Now()))

	req := createAuthenticatedRequest(http.MethodPost, "/chats/123/messages", []byte("invalid json"), 1, "testuser")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	sendMessageHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

// TestUnitSendMessageHandler_ChatError tests sending a message where the chat has a database error
func TestUnitSendMessageHandler_ChatError(t *testing.T) {
	mock, cleanup := setupHandlersTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, user_id, title, model, created_at, updated_at FROM chats WHERE id = \$1 AND user_id = \$2`).
		WithArgs("123", 1).
		WillReturnError(sql.ErrConnDone)

	body := SendMessageRequest{Content: "Hello there"}
	bodyBytes, _ := json.Marshal(body)

	req := createAuthenticatedRequest(http.MethodPost, "/chats/123/messages", bodyBytes, 1, "testuser")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	sendMessageHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

// TestUnitSendMessageHandler_MessagesError tests sending a message where the messages have a database error
func TestUnitSendMessageHandler_MessagesError(t *testing.T) {
	mock, cleanup := setupHandlersTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, user_id, title, model, created_at, updated_at FROM chats WHERE id = \$1 AND user_id = \$2`).
		WithArgs("123", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "model", "created_at", "updated_at"}).
			AddRow("123", 1, "Test Chat", "gemma3", time.Now(), time.Now()))
	mock.ExpectQuery(`SELECT id, chat_id, role, content, status, error_message, tokens_used, created_at FROM messages WHERE chat_id = \$1 ORDER BY created_at ASC`).
		WithArgs("123").
		WillReturnError(sql.ErrConnDone)

	body := SendMessageRequest{Content: "Hello there"}
	bodyBytes, _ := json.Marshal(body)

	req := createAuthenticatedRequest(http.MethodPost, "/chats/123/messages", bodyBytes, 1, "testuser")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	sendMessageHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

// TestUnitSendMessageHandler_InvalidPath tests sending a message with invalid path
func TestUnitSendMessageHandler_InvalidPath(t *testing.T) {
	_, cleanup := setupHandlersTestDB(t)
	defer cleanup()

	body := SendMessageRequest{Content: "Hello there"}
	bodyBytes, _ := json.Marshal(body)

	req := createAuthenticatedRequest(http.MethodPost, "/chats/123/messages/invalid", bodyBytes, 1, "testuser")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	sendMessageHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

// ChatsHandler Router Tests

// TestUnitChatsHandler_ListChats tests the chats router for listing chats
func TestUnitChatsHandler_ListChats(t *testing.T) {
	mock, cleanup := setupHandlersTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM chats WHERE user_id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT id, user_id, title, model, created_at, updated_at FROM chats WHERE user_id = \$1 ORDER BY updated_at DESC LIMIT \$2 OFFSET \$3`).
		WithArgs(1, 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "model", "created_at", "updated_at"}))

	req := createAuthenticatedRequest(http.MethodGet, "/chats", nil, 1, "testuser")
	rr := httptest.NewRecorder()

	chatsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

// TestUnitChatsHandler_CreateChat tests the chats router for creating chats
func TestUnitChatsHandler_CreateChat(t *testing.T) {
	mock, cleanup := setupHandlersTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`INSERT INTO chats \(user_id, model, title\) VALUES \(\$1, \$2, \$3\) RETURNING id, created_at, updated_at`).
		WithArgs(1, "gemma3", "New Chat").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("1", time.Now(), time.Now()))

	body := CreateChatRequest{Model: "gemma3"}
	bodyBytes, _ := json.Marshal(body)

	req := createAuthenticatedRequest(http.MethodPost, "/chats", bodyBytes, 1, "testuser")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	chatsHandler(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, rr.Code)
	}
}

// TestUnitChatsHandler_GetChat tests the chats router for getting a specific chat
func TestUnitChatsHandler_GetChat(t *testing.T) {
	mock, cleanup := setupHandlersTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, user_id, title, model, created_at, updated_at FROM chats WHERE id = \$1 AND user_id = \$2`).
		WithArgs("123", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "model", "created_at", "updated_at"}).
			AddRow("123", 1, "Test Chat", "gemma3", time.Now(), time.Now()))

	mock.ExpectQuery(`SELECT id, chat_id, role, content, status, error_message, tokens_used, created_at FROM messages WHERE chat_id = \$1 ORDER BY created_at ASC`).
		WithArgs("123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "chat_id", "role", "content", "status", "error_message", "tokens_used", "created_at"}))

	req := createAuthenticatedRequest(http.MethodGet, "/chats/123", nil, 1, "testuser")
	rr := httptest.NewRecorder()

	chatsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

// TestUnitChatsHandler_DeleteChat tests the chats router for deleting a chat
func TestUnitChatsHandler_DeleteChat(t *testing.T) {
	mock, cleanup := setupHandlersTestDB(t)
	defer cleanup()

	mock.ExpectExec(`DELETE FROM chats WHERE id = \$1 AND user_id = \$2`).
		WithArgs("123", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := createAuthenticatedRequest(http.MethodDelete, "/chats/123", nil, 1, "testuser")
	rr := httptest.NewRecorder()

	chatsHandler(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, rr.Code)
	}
}

// TestUnitChatsHandler_SendMessage tests the chats router for sending a message
func TestUnitChatsHandler_SendMessage(t *testing.T) {
	mock, cleanup := setupHandlersTestDB(t)
	defer cleanup()

	// Store original pubsub vars and set to nil to skip publishing
	originalPubsubClient := pubsubClient
	originalTopic := topic
	pubsubClient = nil
	topic = nil
	defer func() {
		pubsubClient = originalPubsubClient
		topic = originalTopic
	}()

	mock.ExpectQuery(`SELECT id, user_id, title, model, created_at, updated_at FROM chats WHERE id = \$1 AND user_id = \$2`).
		WithArgs("123", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "model", "created_at", "updated_at"}).
			AddRow("123", 1, "Test Chat", "gemma3", time.Now(), time.Now()))

	mock.ExpectQuery(`INSERT INTO messages \(chat_id, role, content, status\) VALUES \(\$1, \$2, \$3, \$4\) RETURNING id, created_at`).
		WithArgs("123", RoleUser, "Hello", StatusCompleted).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow("1", time.Now()))

	mock.ExpectExec(`UPDATE chats SET updated_at = CURRENT_TIMESTAMP WHERE id = \$1`).
		WithArgs("123").
		WillReturnResult(sqlmock.NewResult(0, 1))

	body := SendMessageRequest{Content: "Hello"}
	bodyBytes, _ := json.Marshal(body)

	req := createAuthenticatedRequest(http.MethodPost, "/chats/123/messages", bodyBytes, 1, "testuser")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	chatsHandler(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, rr.Code)
	}
}

// TestUnitChatsHandler_NotFound tests the chats router for an unknown path
func TestUnitChatsHandler_NotFound(t *testing.T) {
	req := createAuthenticatedRequest(http.MethodGet, "/unknown", nil, 1, "testuser")
	rr := httptest.NewRecorder()

	chatsHandler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

// TestUnitChatsHandler_MethodNotAllowed tests the chats router for unsupported method
func TestUnitChatsHandler_MethodNotAllowed(t *testing.T) {
	req := createAuthenticatedRequest(http.MethodPatch, "/chats", nil, 1, "testuser")
	rr := httptest.NewRecorder()

	chatsHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}
