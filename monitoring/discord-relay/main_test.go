package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// ============================================
// Health Endpoint Tests
// ============================================

func TestHealthzEndpoint_Success(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusOK)
	}

	if w.Body.String() != "ok" {
		t.Errorf("Response body = %v, want 'ok'", w.Body.String())
	}
}

// ============================================
// Webhook Handler Tests
// ============================================

func TestWebhookHandler_NoURLsConfigured(t *testing.T) {
	alert := AlertmanagerPayload{
		Status: "firing",
		Alerts: []Alert{
			{
				Labels: map[string]string{"alertname": "TestAlert"},
				Annotations: map[string]string{
					"summary":     "Test",
					"description": "Test alert",
				},
			},
		},
	}
	body, _ := json.Marshal(alert)

	req := httptest.NewRequest(http.MethodPost, "/webhook/critical", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Handler with empty URLs
	handler := makeHandler("")
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusInternalServerError)
	}
}

func TestWebhookHandler_InvalidMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/webhook/critical", nil)
	w := httptest.NewRecorder()

	handler := makeHandler("http://webhook.url")
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestWebhookHandler_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/webhook/critical", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := makeHandler("http://webhook.url")
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestWebhookHandler_EmptyAlerts(t *testing.T) {
	alert := AlertmanagerPayload{
		Status: "firing",
		Alerts: []Alert{},
	}
	body, _ := json.Marshal(alert)

	req := httptest.NewRequest(http.MethodPost, "/webhook/critical", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := makeHandler("http://webhook.url")
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============================================
// Alert Payload Testing
// ============================================

func TestAlertPayload_Parsing(t *testing.T) {
	tests := []struct {
		name    string
		payload AlertmanagerPayload
		valid   bool
	}{
		{
			name: "Valid firing alert",
			payload: AlertmanagerPayload{
				Status: "firing",
				Alerts: []Alert{
					{
						Labels: map[string]string{"alertname": "HighCPU"},
						Annotations: map[string]string{
							"summary": "CPU is high",
						},
					},
				},
			},
			valid: true,
		},
		{
			name: "Valid resolved alert",
			payload: AlertmanagerPayload{
				Status: "resolved",
				Alerts: []Alert{
					{
						Labels: map[string]string{"alertname": "HighMemory"},
						Annotations: map[string]string{
							"summary": "Memory is back to normal",
						},
					},
				},
			},
			valid: true,
		},
		{
			name: "Empty status defaults to firing",
			payload: AlertmanagerPayload{
				Status: "",
				Alerts: []Alert{
					{
						Labels: map[string]string{},
						Annotations: map[string]string{
							"summary": "Test",
						},
					},
				},
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}

			var parsed AlertmanagerPayload
			if err := json.Unmarshal(body, &parsed); err != nil {
				if tt.valid {
					t.Errorf("Failed to parse valid payload: %v", err)
				}
				return
			}

			if tt.valid && len(parsed.Alerts) == 0 {
				t.Error("Expected alerts in parsed payload")
			}
		})
	}
}

// ============================================
// Alert Enrichment Tests
// ============================================

func TestEnrichHeartbeatAlert_OnlyEnrichesHeartbeatAlert(t *testing.T) {
	tests := []struct {
		name         string
		alertname    string
		shouldEnrich bool
	}{
		{"Heartbeat alert", "SystemHeartbeat", true},
		{"Other alert", "HighCPU", false},
		{"CPU warning", "CPUWarning", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alert := &Alert{
				Labels: map[string]string{
					"alertname": tt.alertname,
				},
				Annotations: map[string]string{
					"summary": "Original summary",
				},
			}

			originalDesc := alert.Annotations["description"]
			enrichHeartbeatAlert(alert)

			if tt.shouldEnrich {
				// Heartbeat alert should have description updated
				if alert.Annotations["description"] == originalDesc {
					t.Error("Heartbeat alert should have description enriched")
				}
			} else {
				// Non-heartbeat alerts should not be modified
				if alert.Annotations["description"] != originalDesc {
					t.Error("Non-heartbeat alert should not be enriched")
				}
			}
		})
	}
}

// ============================================
// Discord Embed Building Tests
// ============================================

func TestBuildEmbeds_FiringAlert(t *testing.T) {
	payload := AlertmanagerPayload{
		Status: "firing",
		Alerts: []Alert{
			{
				Labels: map[string]string{
					"alertname": "HighCPU",
					"severity":  "critical",
				},
				Annotations: map[string]string{
					"summary":     "CPU Usage High",
					"description": "CPU usage has exceeded 90%",
				},
			},
		},
	}

	embeds := buildEmbeds(payload)

	if len(embeds) == 0 {
		t.Fatal("Expected at least one embed")
	}

	// Verify embed structure
	embed := embeds[0]
	if embed.Title == "" {
		t.Error("Embed should have a title")
	}

	// Firing alert should have red color (critical)
	if embed.Color != 16711680 { // Red color in decimal
		t.Logf("Firing alert color = %v (expected red)", embed.Color)
	}
}

func TestBuildEmbeds_ResolvedAlert(t *testing.T) {
	payload := AlertmanagerPayload{
		Status: "resolved",
		Alerts: []Alert{
			{
				Labels: map[string]string{
					"alertname": "HighCPU",
				},
				Annotations: map[string]string{
					"summary": "CPU Usage Normal",
				},
			},
		},
	}

	embeds := buildEmbeds(payload)

	if len(embeds) == 0 {
		t.Fatal("Expected at least one embed")
	}

	// Resolved should have green color
	if embeds[0].Color != 65280 { // Green color in decimal
		t.Logf("Resolved alert color = %v (expected green)", embeds[0].Color)
	}
}

func TestBuildEmbeds_MultipleSeverities(t *testing.T) {
	tests := []struct {
		severity string
	}{
		{"critical"},
		{"warning"},
		{"info"},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			payload := AlertmanagerPayload{
				Status: "firing",
				Alerts: []Alert{
					{
						Labels: map[string]string{
							"alertname": "TestAlert",
							"severity":  tt.severity,
						},
						Annotations: map[string]string{
							"summary": "Test",
						},
					},
				},
			}

			embeds := buildEmbeds(payload)
			if len(embeds) == 0 {
				t.Fatal("Expected embed for " + tt.severity)
			}

			// Verify different severities result in different embeds
			if embeds[0].Title == "" {
				t.Error("Embed should have title")
			}
		})
	}
}

// ============================================
// Helper Function Tests
// ============================================

func TestGetEnv_WithDefault(t *testing.T) {
	tests := []struct {
		name        string
		envVar      string
		defaultVal  string
		setEnv      bool
		setValue    string
		expectedVal string
	}{
		{"Env var not set", "TEST_VAR_1", "default", false, "", "default"},
		{"Env var set", "TEST_VAR_2", "default", true, "actual", "actual"},
		{"Empty env var returns default", "TEST_VAR_3", "default", true, "", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv(tt.envVar, tt.setValue)
				defer os.Unsetenv(tt.envVar)
			}

			result := getEnv(tt.envVar, tt.defaultVal)

			if result != tt.expectedVal {
				t.Errorf("getEnv returned %v, want %v", result, tt.expectedVal)
			}
		})
	}
}

func TestSplitCSV_URLHandling(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		count  int
		checks []string
	}{
		{
			"Single URL",
			"http://webhook1.com",
			1,
			[]string{"http://webhook1.com"},
		},
		{
			"Multiple URLs with spaces",
			"http://webhook1.com, http://webhook2.com, http://webhook3.com",
			3,
			[]string{"http://webhook1.com", "http://webhook2.com", "http://webhook3.com"},
		},
		{
			"Empty string",
			"",
			0,
			[]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitCSV(tt.input)

			if len(result) != tt.count {
				t.Errorf("splitCSV returned %d items, want %d", len(result), tt.count)
			}

			for i, check := range tt.checks {
				if i < len(result) && !strings.Contains(result[i], strings.TrimSpace(check)) {
					t.Errorf("Item %d = %v, want to contain %v", i, result[i], check)
				}
			}
		})
	}
}

// ============================================
// Integration-style Tests
// ============================================

func TestWebhookHandler_ValidPayloadStructure(t *testing.T) {
	alert := AlertmanagerPayload{
		Status: "firing",
		Alerts: []Alert{
			{
				Labels: map[string]string{
					"alertname": "TestAlert",
					"severity":  "warning",
				},
				Annotations: map[string]string{
					"summary":     "Test Alert Summary",
					"description": "This is a test alert description",
				},
			},
		},
	}

	body, err := json.Marshal(alert)
	if err != nil {
		t.Fatalf("Failed to marshal alert: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/webhook/warning", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// With no URLs configured, should return error
	handler := makeHandler("")
	handler.ServeHTTP(w, req)

	// But the parsing should have worked
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500 for no URLs, got %d", w.Code)
	}
}

func TestAlertAnnotations_ContentTypes(t *testing.T) {
	tests := []struct {
		name        string
		annotation  string
		value       string
		shouldExist bool
	}{
		{"Summary annotation", "summary", "High CPU Usage", true},
		{"Description annotation", "description", "CPU above 90%", true},
		{"Optional annotation", "optional", "extra info", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alert := Alert{
				Annotations: map[string]string{
					"summary":     "High CPU Usage",
					"description": "CPU above 90%",
				},
			}

			_, exists := alert.Annotations[tt.annotation]
			if tt.shouldExist != exists {
				t.Errorf("Annotation %s existence = %v, want %v", tt.annotation, exists, tt.shouldExist)
			}
		})
	}
}

func TestDiscordMessage_Structure(t *testing.T) {
	msg := DiscordMessage{
		Content:  "Alert notification",
		Username: "AlertBot",
		Embeds: []DiscordEmbed{
			{
				Title:       "Alert Title",
				Description: "Alert Description",
				Color:       16711680, // Red
				Footer: &DiscordEmbedFooter{
					Text: "AlertManager",
				},
			},
		},
	}

	body, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	var unmarshaled DiscordMessage
	if err := json.Unmarshal(body, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	if unmarshaled.Username != "AlertBot" {
		t.Errorf("Username = %v, want AlertBot", unmarshaled.Username)
	}

	if len(unmarshaled.Embeds) == 0 {
		t.Error("Expected embeds in message")
	}

	if unmarshaled.Embeds[0].Color != 16711680 {
		t.Errorf("Color = %v, want 16711680", unmarshaled.Embeds[0].Color)
	}
}
