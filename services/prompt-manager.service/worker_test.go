package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestUnitBuildPromptFromHistory_Empty tests the buildPromptFromHistory function with empty messages
func TestUnitBuildPromptFromHistory_Empty(t *testing.T) {
	messages := []Message{}
	result := buildPromptFromHistory(messages, "gemma3")

	// Should still have the system prompt preamble and model response trigger
	if !strings.Contains(result, systemPrompt) {
		t.Errorf("Expected result to contain system prompt, got '%s'", result)
	}
	if !strings.HasSuffix(result, "<start_of_turn>model\n") {
		t.Errorf("Expected result to end with model turn, got '%s'", result)
	}
}

// TestUnitBuildPromptFromHistory_SingleUserMessage tests with a single user message
func TestUnitBuildPromptFromHistory_SingleUserMessage(t *testing.T) {
	messages := []Message{
		{Role: RoleUser, Content: "Hello", Status: StatusCompleted},
	}
	result := buildPromptFromHistory(messages, "gemma3")

	if !strings.Contains(result, "<start_of_turn>user\nHello<end_of_turn>") {
		t.Errorf("Expected result to contain user message, got '%s'", result)
	}
	if !strings.HasSuffix(result, "<start_of_turn>model\n") {
		t.Errorf("Expected result to end with model turn, got '%s'", result)
	}
}

// TestUnitBuildPromptFromHistory_Conversation tests with a full conversation
func TestUnitBuildPromptFromHistory_Conversation(t *testing.T) {
	messages := []Message{
		{Role: RoleUser, Content: "Hello", Status: StatusCompleted},
		{Role: RoleAssistant, Content: "Hi there!", Status: StatusCompleted},
		{Role: RoleUser, Content: "How are you?", Status: StatusCompleted},
	}
	result := buildPromptFromHistory(messages, "gemma3")

	if !strings.Contains(result, "<start_of_turn>user\nHello<end_of_turn>") {
		t.Errorf("Expected result to contain first user message, got '%s'", result)
	}
	if !strings.Contains(result, "<start_of_turn>model\nHi there!<end_of_turn>") {
		t.Errorf("Expected result to contain assistant message, got '%s'", result)
	}
	if !strings.Contains(result, "<start_of_turn>user\nHow are you?<end_of_turn>") {
		t.Errorf("Expected result to contain second user message, got '%s'", result)
	}
}

// TestUnitBuildPromptFromHistory_SkipsPendingAssistant tests that pending assistant messages are skipped
func TestUnitBuildPromptFromHistory_SkipsPendingAssistant(t *testing.T) {
	messages := []Message{
		{Role: RoleUser, Content: "Hello", Status: StatusCompleted},
		{Role: RoleAssistant, Content: "", Status: StatusPending},
	}
	result := buildPromptFromHistory(messages, "gemma3")

	if !strings.Contains(result, "<start_of_turn>user\nHello<end_of_turn>") {
		t.Errorf("Expected result to contain user message, got '%s'", result)
	}
	// Should NOT contain an empty assistant message
	if strings.Contains(result, "<start_of_turn>model\n<end_of_turn>") {
		t.Errorf("Should not contain empty pending assistant message, got '%s'", result)
	}
}

// TestUnitBuildPromptFromHistory_Qwen3Format tests Qwen3 model-specific format
func TestUnitBuildPromptFromHistory_Qwen3Format(t *testing.T) {
	messages := []Message{
		{Role: RoleUser, Content: "Hello", Status: StatusCompleted},
	}
	result := buildPromptFromHistory(messages, "qwen3")

	// Qwen3 uses ChatML format
	if !strings.Contains(result, "<|im_start|>system\n") {
		t.Errorf("Expected Qwen3 system tag, got '%s'", result)
	}
	// Should contain /no_think to disable reasoning output
	if !strings.Contains(result, "/no_think") {
		t.Errorf("Expected Qwen3 to have /no_think directive, got '%s'", result)
	}
	if !strings.Contains(result, "<|im_start|>user\nHello<|im_end|>") {
		t.Errorf("Expected Qwen3 user format, got '%s'", result)
	}
	if !strings.HasSuffix(result, "<|im_start|>assistant\n") {
		t.Errorf("Expected result to end with assistant turn, got '%s'", result)
	}
}

// TestUnitBuildPromptFromHistory_Gemma3Format tests Gemma3 model-specific format
func TestUnitBuildPromptFromHistory_Gemma3Format(t *testing.T) {
	messages := []Message{
		{Role: RoleUser, Content: "What is 2+2?", Status: StatusCompleted},
		{Role: RoleAssistant, Content: "4", Status: StatusCompleted},
		{Role: RoleUser, Content: "What is 3+3?", Status: StatusCompleted},
	}
	result := buildPromptFromHistory(messages, "gemma3")

	// Gemma3 uses start_of_turn/end_of_turn format
	if !strings.Contains(result, "<start_of_turn>user\n") {
		t.Errorf("Expected Gemma3 user tag, got '%s'", result)
	}
	if !strings.Contains(result, "<start_of_turn>model\n4<end_of_turn>") {
		t.Errorf("Expected Gemma3 model response, got '%s'", result)
	}
	if !strings.HasSuffix(result, "<start_of_turn>model\n") {
		t.Errorf("Expected result to end with model turn, got '%s'", result)
	}
}

// TestUnitBuildPromptFromHistory_DefaultModel tests default model falls back to Gemma3
func TestUnitBuildPromptFromHistory_DefaultModel(t *testing.T) {
	messages := []Message{
		{Role: RoleUser, Content: "Hello", Status: StatusCompleted},
	}
	result := buildPromptFromHistory(messages, "unknown_model")

	// Should use Gemma3 format by default
	if !strings.Contains(result, "<start_of_turn>") {
		t.Errorf("Expected Gemma3 format for unknown model, got '%s'", result)
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
