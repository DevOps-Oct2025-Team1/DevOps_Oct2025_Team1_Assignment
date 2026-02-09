package main

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// setupChatsTestDB initializes a test database for chats
func setupChatsTestDB(t *testing.T) (sqlmock.Sqlmock, func()) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}

	// Replace global db with mock
	originalDB := db
	db = mockDB

	cleanup := func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet sqlmock expectations: %v", err)
		}
		mockDB.Close()
		db = originalDB
	}

	return mock, cleanup
}

// TestUnitCreateChat tests the CreateChat function
func TestUnitCreateChat(t *testing.T) {
	mock, cleanup := setupChatsTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`INSERT INTO chats \(user_id, model, title\) VALUES \(\$1, \$2, \$3\) RETURNING id, created_at, updated_at`).
		WithArgs(1, "qwen3", "New Chat").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("1", time.Now(), time.Now()))

	chat, err := CreateChat(1, "qwen3")
	if err != nil {
		t.Fatalf("Failed to create chat: %v", err)
	}

	if chat.Title != "New Chat" {
		t.Errorf("Expected chat title 'New Chat', got '%s'", chat.Title)
	}
}

// Get Chats Handler Tests

// TestUnitGetChat_Success tests the GetChat function with a valid chat ID and user ID
func TestUnitGetChat_Success(t *testing.T) {
	mock, cleanup := setupChatsTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, user_id, title, model, created_at, updated_at FROM chats WHERE id = \$1 AND user_id = \$2`).
		WithArgs("1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "model", "created_at", "updated_at"}).
			AddRow("1", 1, "Test Chat", "qwen3", time.Now(), time.Now()))

	chat, err := GetChat("1", 1)
	if err != nil {
		t.Fatalf("Failed to get chat: %v", err)
	}

	if chat.Title != "Test Chat" {
		t.Errorf("Expected chat title 'Test Chat', got '%s'", chat.Title)
	}
}

// TestUnitGetChat_NotFound tests the GetChat function with a non-existent chat ID
func TestUnitGetChat_NotFound(t *testing.T) {
	mock, cleanup := setupChatsTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, user_id, title, model, created_at, updated_at FROM chats WHERE id = \$1 AND user_id = \$2`).
		WithArgs("1", 1).
		WillReturnError(sql.ErrNoRows)

	chat, err := GetChat("1", 1)
	if err != nil {
		t.Fatalf("Failed to get chat: %v", err)
	}

	if chat != nil {
		t.Errorf("Expected chat to be nil, got '%v'", chat)
	}
}

// TestUnitGetChat_Error tests the GetChat function with a database error
func TestUnitGetChat_Error(t *testing.T) {
	mock, cleanup := setupChatsTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, user_id, title, model, created_at, updated_at FROM chats WHERE id = \$1 AND user_id = \$2`).
		WithArgs("1", 1).
		WillReturnError(sql.ErrConnDone)

	chat, err := GetChat("1", 1)
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}

	if chat != nil {
		t.Errorf("Expected chat to be nil, got '%v'", chat)
	}
}

// ListChats Tests

// TestUnitListChats_Success tests the ListChats function with successful retrieval
func TestUnitListChats_Success(t *testing.T) {
	mock, cleanup := setupChatsTestDB(t)
	defer cleanup()

	// Expect count query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM chats WHERE user_id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Expect list query
	mock.ExpectQuery(`SELECT id, user_id, title, model, created_at, updated_at FROM chats WHERE user_id = \$1 ORDER BY updated_at DESC LIMIT \$2 OFFSET \$3`).
		WithArgs(1, 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "model", "created_at", "updated_at"}).
			AddRow("1", 1, "Chat 1", "gemma3", time.Now(), time.Now()).
			AddRow("2", 1, "Chat 2", "qwen3", time.Now(), time.Now()))

	chats, totalCount, err := ListChats(1, 1, 20)
	if err != nil {
		t.Fatalf("Failed to list chats: %v", err)
	}

	if len(chats) != 2 {
		t.Errorf("Expected 2 chats, got %d", len(chats))
	}

	if totalCount != 2 {
		t.Errorf("Expected total count of 2, got %d", totalCount)
	}
}

// TestUnitListChats_Empty tests the ListChats function with no chats
func TestUnitListChats_Empty(t *testing.T) {
	mock, cleanup := setupChatsTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM chats WHERE user_id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT id, user_id, title, model, created_at, updated_at FROM chats WHERE user_id = \$1 ORDER BY updated_at DESC LIMIT \$2 OFFSET \$3`).
		WithArgs(1, 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "model", "created_at", "updated_at"}))

	chats, totalCount, err := ListChats(1, 1, 20)
	if err != nil {
		t.Fatalf("Failed to list chats: %v", err)
	}

	if len(chats) != 0 {
		t.Errorf("Expected empty chats slice, got %d", len(chats))
	}

	if totalCount != 0 {
		t.Errorf("Expected total count of 0, got %d", totalCount)
	}
}

// TestUnitListChats_Pagination tests the ListChats function with pagination
func TestUnitListChats_Pagination(t *testing.T) {
	mock, cleanup := setupChatsTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM chats WHERE user_id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(25))

	mock.ExpectQuery(`SELECT id, user_id, title, model, created_at, updated_at FROM chats WHERE user_id = \$1 ORDER BY updated_at DESC LIMIT \$2 OFFSET \$3`).
		WithArgs(1, 10, 10). // page 2, pageSize 10 -> offset 10
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "model", "created_at", "updated_at"}).
			AddRow("11", 1, "Chat 11", "gemma3", time.Now(), time.Now()))

	chats, totalCount, err := ListChats(1, 2, 10)
	if err != nil {
		t.Fatalf("Failed to list chats: %v", err)
	}

	if len(chats) != 1 {
		t.Errorf("Expected 1 chat, got %d", len(chats))
	}

	if totalCount != 25 {
		t.Errorf("Expected total count of 25, got %d", totalCount)
	}
}

// TestUnitListChats_CountError tests the ListChats function with count query error
func TestUnitListChats_CountError(t *testing.T) {
	mock, cleanup := setupChatsTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM chats WHERE user_id = \$1`).
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	chats, totalCount, err := ListChats(1, 1, 20)
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}

	if chats != nil {
		t.Errorf("Expected chats to be nil, got %v", chats)
	}

	if totalCount != 0 {
		t.Errorf("Expected total count of 0, got %d", totalCount)
	}
}

// TestUnitListChats_QueryError tests the ListChats function with query error
func TestUnitListChats_QueryError(t *testing.T) {
	mock, cleanup := setupChatsTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM chats WHERE user_id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	mock.ExpectQuery(`SELECT id, user_id, title, model, created_at, updated_at FROM chats WHERE user_id = \$1 ORDER BY updated_at DESC LIMIT \$2 OFFSET \$3`).
		WithArgs(1, 20, 0).
		WillReturnError(sql.ErrConnDone)

	chats, totalCount, err := ListChats(1, 1, 20)
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}

	if chats != nil {
		t.Errorf("Expected chats to be nil, got %v", chats)
	}

	if totalCount != 0 {
		t.Errorf("Expected total count of 0, got %d", totalCount)
	}
}

// DeleteChat Tests

// TestUnitDeleteChat_Success tests the DeleteChat function with successful deletion
func TestUnitDeleteChat_Success(t *testing.T) {
	mock, cleanup := setupChatsTestDB(t)
	defer cleanup()

	mock.ExpectExec(`DELETE FROM chats WHERE id = \$1 AND user_id = \$2`).
		WithArgs("1", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := DeleteChat("1", 1)
	if err != nil {
		t.Fatalf("Failed to delete chat: %v", err)
	}
}

// TestUnitDeleteChat_NotFound tests the DeleteChat function when chat is not found
func TestUnitDeleteChat_NotFound(t *testing.T) {
	mock, cleanup := setupChatsTestDB(t)
	defer cleanup()

	mock.ExpectExec(`DELETE FROM chats WHERE id = \$1 AND user_id = \$2`).
		WithArgs("999", 1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := DeleteChat("999", 1)
	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}
}

// TestUnitDeleteChat_Error tests the DeleteChat function with a database error
func TestUnitDeleteChat_Error(t *testing.T) {
	mock, cleanup := setupChatsTestDB(t)
	defer cleanup()

	mock.ExpectExec(`DELETE FROM chats WHERE id = \$1 AND user_id = \$2`).
		WithArgs("1", 1).
		WillReturnError(sql.ErrConnDone)

	err := DeleteChat("1", 1)
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}
}

// CreateMessage Tests

// TestUnitCreateMessage_UserMessage tests creating a user message
func TestUnitCreateMessage_UserMessage(t *testing.T) {
	mock, cleanup := setupChatsTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`INSERT INTO messages \(chat_id, role, content, status\) VALUES \(\$1, \$2, \$3, \$4\) RETURNING id, created_at`).
		WithArgs("1", RoleUser, "Hello", StatusCompleted).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow("1", time.Now()))

	mock.ExpectExec(`UPDATE chats SET updated_at = CURRENT_TIMESTAMP WHERE id = \$1`).
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	msg, err := CreateMessage("1", RoleUser, "Hello")
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	if msg.Role != RoleUser {
		t.Errorf("Expected role 'user', got '%s'", msg.Role)
	}

	if msg.Status != StatusCompleted {
		t.Errorf("Expected status 'completed', got '%s'", msg.Status)
	}

	if msg.Content != "Hello" {
		t.Errorf("Expected content 'Hello', got '%s'", msg.Content)
	}
}

// TestUnitCreateMessage_AssistantMessage tests creating an assistant message (should be pending)
func TestUnitCreateMessage_AssistantMessage(t *testing.T) {
	mock, cleanup := setupChatsTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`INSERT INTO messages \(chat_id, role, content, status\) VALUES \(\$1, \$2, \$3, \$4\) RETURNING id, created_at`).
		WithArgs("1", RoleAssistant, "", StatusPending).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow("2", time.Now()))

	mock.ExpectExec(`UPDATE chats SET updated_at = CURRENT_TIMESTAMP WHERE id = \$1`).
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	msg, err := CreateMessage("1", RoleAssistant, "")
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	if msg.Role != RoleAssistant {
		t.Errorf("Expected role 'assistant', got '%s'", msg.Role)
	}

	if msg.Status != StatusPending {
		t.Errorf("Expected status 'pending', got '%s'", msg.Status)
	}
}

// TestUnitCreateMessage_Error tests creating a message with a database error
func TestUnitCreateMessage_Error(t *testing.T) {
	mock, cleanup := setupChatsTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`INSERT INTO messages \(chat_id, role, content, status\) VALUES \(\$1, \$2, \$3, \$4\) RETURNING id, created_at`).
		WithArgs("1", RoleUser, "Hello", StatusCompleted).
		WillReturnError(sql.ErrConnDone)

	msg, err := CreateMessage("1", RoleUser, "Hello")
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}

	if msg != nil {
		t.Errorf("Expected message to be nil, got %v", msg)
	}
}

// GetMessages Tests

// TestUnitGetMessages_Success tests the GetMessages function with successful retrieval
func TestUnitGetMessages_Success(t *testing.T) {
	mock, cleanup := setupChatsTestDB(t)
	defer cleanup()

	tokens := 100
	mock.ExpectQuery(`SELECT id, chat_id, role, content, status, error_message, tokens_used, created_at FROM messages WHERE chat_id = \$1 ORDER BY created_at ASC`).
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "chat_id", "role", "content", "status", "error_message", "tokens_used", "created_at"}).
			AddRow("1", "1", RoleUser, "Hello", StatusCompleted, nil, nil, time.Now()).
			AddRow("2", "1", RoleAssistant, "Hi there!", StatusCompleted, nil, &tokens, time.Now()))

	messages, err := GetMessages("1")
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	if messages[0].Role != RoleUser {
		t.Errorf("Expected first message role 'user', got '%s'", messages[0].Role)
	}

	if messages[0].Content != "Hello" {
		t.Errorf("Expected first message content 'Hello', got '%s'", messages[0].Content)
	}

	if messages[1].Role != RoleAssistant {
		t.Errorf("Expected second message role 'assistant', got '%s'", messages[1].Role)
	}

	if messages[1].Content != "Hi there!" {
		t.Errorf("Expected second message content 'Hi there!', got '%s'", messages[1].Content)
	}
}

// TestUnitGetMessages_Empty tests the GetMessages function with no messages
func TestUnitGetMessages_Empty(t *testing.T) {
	mock, cleanup := setupChatsTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, chat_id, role, content, status, error_message, tokens_used, created_at FROM messages WHERE chat_id = \$1 ORDER BY created_at ASC`).
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "chat_id", "role", "content", "status", "error_message", "tokens_used", "created_at"}))

	messages, err := GetMessages("1")
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	if len(messages) != 0 {
		t.Errorf("Expected empty messages slice, got %d", len(messages))
	}
}

// TestUnitGetMessages_Error tests the GetMessages function with a database error
func TestUnitGetMessages_Error(t *testing.T) {
	mock, cleanup := setupChatsTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, chat_id, role, content, status, error_message, tokens_used, created_at FROM messages WHERE chat_id = \$1 ORDER BY created_at ASC`).
		WithArgs("1").
		WillReturnError(sql.ErrConnDone)

	messages, err := GetMessages("1")
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}

	if messages != nil {
		t.Errorf("Expected messages to be nil, got %v", messages)
	}
}

// UpdateMessageStatus Tests

// TestUnitUpdateMessageStatus_StatusOnly tests updating only the status
func TestUnitUpdateMessageStatus_StatusOnly(t *testing.T) {
	mock, cleanup := setupChatsTestDB(t)
	defer cleanup()

	mock.ExpectExec(`UPDATE messages SET status = \$1, updated_at = CURRENT_TIMESTAMP WHERE id = \$2`).
		WithArgs(StatusCompleted, "1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := UpdateMessageStatus("1", StatusCompleted, nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to update message status: %v", err)
	}
}

// TestUnitUpdateMessageStatus_WithContent tests updating status with content
func TestUnitUpdateMessageStatus_WithContent(t *testing.T) {
	mock, cleanup := setupChatsTestDB(t)
	defer cleanup()

	content := "Hello, I am an assistant"
	mock.ExpectExec(`UPDATE messages SET status = \$1, updated_at = CURRENT_TIMESTAMP, content = \$2 WHERE id = \$3`).
		WithArgs(StatusCompleted, content, "1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := UpdateMessageStatus("1", StatusCompleted, &content, nil, nil)
	if err != nil {
		t.Fatalf("Failed to update message status: %v", err)
	}
}

// TestUnitUpdateMessageStatus_WithTokens tests updating status with tokens
func TestUnitUpdateMessageStatus_WithTokens(t *testing.T) {
	mock, cleanup := setupChatsTestDB(t)
	defer cleanup()

	tokens := 150
	mock.ExpectExec(`UPDATE messages SET status = \$1, updated_at = CURRENT_TIMESTAMP, tokens_used = \$2 WHERE id = \$3`).
		WithArgs(StatusCompleted, tokens, "1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := UpdateMessageStatus("1", StatusCompleted, nil, &tokens, nil)
	if err != nil {
		t.Fatalf("Failed to update message status: %v", err)
	}
}

// TestUnitUpdateMessageStatus_WithError tests updating status with error message
func TestUnitUpdateMessageStatus_WithError(t *testing.T) {
	mock, cleanup := setupChatsTestDB(t)
	defer cleanup()

	errorMsg := "LLM service unavailable"
	mock.ExpectExec(`UPDATE messages SET status = \$1, updated_at = CURRENT_TIMESTAMP, error_message = \$2 WHERE id = \$3`).
		WithArgs(StatusFailed, errorMsg, "1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := UpdateMessageStatus("1", StatusFailed, nil, nil, &errorMsg)
	if err != nil {
		t.Fatalf("Failed to update message status: %v", err)
	}
}

// TestUnitUpdateMessageStatus_AllFields tests updating status with all optional fields
func TestUnitUpdateMessageStatus_AllFields(t *testing.T) {
	mock, cleanup := setupChatsTestDB(t)
	defer cleanup()

	content := "Response content"
	tokens := 200
	errorMsg := ""
	mock.ExpectExec(`UPDATE messages SET status = \$1, updated_at = CURRENT_TIMESTAMP, content = \$2, tokens_used = \$3, error_message = \$4 WHERE id = \$5`).
		WithArgs(StatusCompleted, content, tokens, errorMsg, "1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := UpdateMessageStatus("1", StatusCompleted, &content, &tokens, &errorMsg)
	if err != nil {
		t.Fatalf("Failed to update message status: %v", err)
	}
}

// TestUnitUpdateMessageStatus_Error tests updating message status with a database error
func TestUnitUpdateMessageStatus_Error(t *testing.T) {
	mock, cleanup := setupChatsTestDB(t)
	defer cleanup()

	mock.ExpectExec(`UPDATE messages SET status = \$1, updated_at = CURRENT_TIMESTAMP WHERE id = \$2`).
		WithArgs(StatusCompleted, "1").
		WillReturnError(sql.ErrConnDone)

	err := UpdateMessageStatus("1", StatusCompleted, nil, nil, nil)
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}
}

// UpdateChatTitle Tests

// TestUnitUpdateChatTitle_Success tests successful title update
func TestUnitUpdateChatTitle_Success(t *testing.T) {
	mock, cleanup := setupChatsTestDB(t)
	defer cleanup()

	mock.ExpectExec(`UPDATE chats SET title = \$1 WHERE id = \$2`).
		WithArgs("New Title", "1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := UpdateChatTitle("1", "New Title")
	if err != nil {
		t.Fatalf("Failed to update chat title: %v", err)
	}
}

// TestUnitUpdateChatTitle_Error tests updating chat title with a database error
func TestUnitUpdateChatTitle_Error(t *testing.T) {
	mock, cleanup := setupChatsTestDB(t)
	defer cleanup()

	mock.ExpectExec(`UPDATE chats SET title = \$1 WHERE id = \$2`).
		WithArgs("New Title", "1").
		WillReturnError(sql.ErrConnDone)

	err := UpdateChatTitle("1", "New Title")
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}
}
