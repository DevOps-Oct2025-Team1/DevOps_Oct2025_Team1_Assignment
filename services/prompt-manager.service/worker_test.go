package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestUnitBuildPromptFromHistory tests the buildPromptFromHistory function
func TestUnitBuildPromptFromHistory_Empty(t *testing.T) {
	messages := []Message{}
	result := buildPromptFromHistory(messages)

	expected := "Assistant: "
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestUnitBuildPromptFromHistory_SingleUserMessage tests with a single user message
func TestUnitBuildPromptFromHistory_SingleUserMessage(t *testing.T) {
	messages := []Message{
		{Role: RoleUser, Content: "Hello", Status: StatusCompleted},
	}
	result := buildPromptFromHistory(messages)

	expected := "User: Hello\nAssistant: "
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestUnitBuildPromptFromHistory_Conversation tests with a full conversation
func TestUnitBuildPromptFromHistory_Conversation(t *testing.T) {
	messages := []Message{
		{Role: RoleUser, Content: "Hello", Status: StatusCompleted},
		{Role: RoleAssistant, Content: "Hi there!", Status: StatusCompleted},
		{Role: RoleUser, Content: "How are you?", Status: StatusCompleted},
	}
	result := buildPromptFromHistory(messages)

	expected := "User: Hello\nAssistant: Hi there!\nUser: How are you?\nAssistant: "
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestUnitBuildPromptFromHistory_SkipsPendingAssistant tests that pending assistant messages are skipped
func TestUnitBuildPromptFromHistory_SkipsPendingAssistant(t *testing.T) {
	messages := []Message{
		{Role: RoleUser, Content: "Hello", Status: StatusCompleted},
		{Role: RoleAssistant, Content: "", Status: StatusPending},
	}
	result := buildPromptFromHistory(messages)

	expected := "User: Hello\nAssistant: "
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestUnitBuildPromptFromHistory_WithSystemMessage tests with a system message
func TestUnitBuildPromptFromHistory_WithSystemMessage(t *testing.T) {
	messages := []Message{
		{Role: RoleSystem, Content: "You are a helpful assistant.", Status: StatusCompleted},
		{Role: RoleUser, Content: "Hello", Status: StatusCompleted},
	}
	result := buildPromptFromHistory(messages)

	expected := "System: You are a helpful assistant.\nUser: Hello\nAssistant: "
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestUnitBuildPromptFromHistory_LongConversation tests with multiple exchanges
func TestUnitBuildPromptFromHistory_LongConversation(t *testing.T) {
	messages := []Message{
		{Role: RoleSystem, Content: "Be concise.", Status: StatusCompleted},
		{Role: RoleUser, Content: "What is 2+2?", Status: StatusCompleted},
		{Role: RoleAssistant, Content: "4", Status: StatusCompleted},
		{Role: RoleUser, Content: "What is 3+3?", Status: StatusCompleted},
		{Role: RoleAssistant, Content: "6", Status: StatusCompleted},
		{Role: RoleUser, Content: "What is 4+4?", Status: StatusCompleted},
		{Role: RoleAssistant, Content: "", Status: StatusPending}, // Current pending response
	}
	result := buildPromptFromHistory(messages)

	expected := "System: Be concise.\nUser: What is 2+2?\nAssistant: 4\nUser: What is 3+3?\nAssistant: 6\nUser: What is 4+4?\nAssistant: "
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestUnitBuildPromptFromHistory_MultilineContent tests with multiline message content
func TestUnitBuildPromptFromHistory_MultilineContent(t *testing.T) {
	messages := []Message{
		{Role: RoleUser, Content: "Line 1\nLine 2\nLine 3", Status: StatusCompleted},
	}
	result := buildPromptFromHistory(messages)

	expected := "User: Line 1\nLine 2\nLine 3\nAssistant: "
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestUnitCallLLMService_Success tests a successful LLM service call
func TestUnitCallLLMService_Success(t *testing.T) {
	// Create a mock LLM server
	mockLLMServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if r.URL.Path != "/completion" {
			t.Errorf("Expected path /completion, got %s", r.URL.Path)
		}

		var req LlamaCppRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if req.Stream != false {
			t.Errorf("Expected stream=false, got %v", req.Stream)
		}

		resp := LlamaCppResponse{
			Content:          "Hello, I am an assistant!",
			TokensEvaluated:  10,
			TokensPredicted:  15,
			GenerationTimeMs: 1000,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockLLMServer.Close()

	content, tokens, err := callLLMService(mockLLMServer.URL, "User: Hello\nAssistant: ")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if content != "Hello, I am an assistant!" {
		t.Errorf("Expected content 'Hello, I am an assistant!', got '%s'", content)
	}

	if tokens != 25 { // 10 + 15
		t.Errorf("Expected tokens 25, got %d", tokens)
	}
}

// TestUnitCallLLMService_ServerError tests LLM service returning an error
func TestUnitCallLLMService_ServerError(t *testing.T) {
	mockLLMServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer mockLLMServer.Close()

	content, tokens, err := callLLMService(mockLLMServer.URL, "User: Hello\nAssistant: ")
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}

	if content != "" {
		t.Errorf("Expected empty content, got '%s'", content)
	}

	if tokens != 0 {
		t.Errorf("Expected 0 tokens, got %d", tokens)
	}
}

// TestUnitCallLLMService_InvalidResponse tests LLM service returning invalid JSON
func TestUnitCallLLMService_InvalidResponse(t *testing.T) {
	mockLLMServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not valid json"))
	}))
	defer mockLLMServer.Close()

	content, tokens, err := callLLMService(mockLLMServer.URL, "User: Hello\nAssistant: ")
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}

	if content != "" {
		t.Errorf("Expected empty content, got '%s'", content)
	}

	if tokens != 0 {
		t.Errorf("Expected 0 tokens, got %d", tokens)
	}
}

// TestUnitCallLLMService_ConnectionError tests LLM service connection error
func TestUnitCallLLMService_ConnectionError(t *testing.T) {
	// Create a server and immediately close it to get a port that refuses connections
	closedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	closedServerURL := closedServer.URL
	closedServer.Close()

	content, tokens, err := callLLMService(closedServerURL, "User: Hello\nAssistant: ")
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}

	if content != "" {
		t.Errorf("Expected empty content, got '%s'", content)
	}

	if tokens != 0 {
		t.Errorf("Expected 0 tokens, got %d", tokens)
	}
}

// TestUnitCallLLMService_EmptyResponse tests LLM service returning empty content
func TestUnitCallLLMService_EmptyResponse(t *testing.T) {
	mockLLMServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := LlamaCppResponse{
			Content:          "",
			TokensEvaluated:  5,
			TokensPredicted:  0,
			GenerationTimeMs: 100,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockLLMServer.Close()

	content, tokens, err := callLLMService(mockLLMServer.URL, "User: Hello\nAssistant: ")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if content != "" {
		t.Errorf("Expected empty content, got '%s'", content)
	}

	if tokens != 5 {
		t.Errorf("Expected tokens 5, got %d", tokens)
	}
}

// TestUnitCallLLMService_LongResponse tests LLM service returning a long response
func TestUnitCallLLMService_LongResponse(t *testing.T) {
	longContent := "This is a very long response that contains multiple sentences. " +
		"It simulates a typical LLM output that might span several paragraphs. " +
		"The response includes various types of content and information."

	mockLLMServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := LlamaCppResponse{
			Content:          longContent,
			TokensEvaluated:  50,
			TokensPredicted:  100,
			GenerationTimeMs: 5000,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockLLMServer.Close()

	content, tokens, err := callLLMService(mockLLMServer.URL, "User: Write a paragraph\nAssistant: ")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if content != longContent {
		t.Errorf("Expected long content, got '%s'", content)
	}

	if tokens != 150 { // 50 + 100
		t.Errorf("Expected tokens 150, got %d", tokens)
	}
}

// TestUnitCallLLMService_RequestFormat tests that the request format is correct
func TestUnitCallLLMService_RequestFormat(t *testing.T) {
	mockLLMServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify content type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}

		// Verify request body
		var req LlamaCppRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if req.NPredict != 512 {
			t.Errorf("Expected n_predict=512, got %d", req.NPredict)
		}

		if req.Stream != false {
			t.Errorf("Expected stream=false, got %v", req.Stream)
		}

		if req.MaxTokens != 2000 {
			t.Errorf("Expected max_tokens=2000, got %d", req.MaxTokens)
		}

		expectedPrompt := "Test prompt"
		if req.Prompt != expectedPrompt {
			t.Errorf("Expected prompt '%s', got '%s'", expectedPrompt, req.Prompt)
		}

		resp := LlamaCppResponse{
			Content:          "Response",
			TokensEvaluated:  5,
			TokensPredicted:  5,
			GenerationTimeMs: 100,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockLLMServer.Close()

	_, _, err := callLLMService(mockLLMServer.URL, "Test prompt")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}
