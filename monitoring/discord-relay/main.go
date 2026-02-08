package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type AlertmanagerPayload struct {
	Status string  `json:"status"`
	Alerts []Alert `json:"alerts"`
}

type Alert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

type DiscordMessage struct {
	Content string `json:"content"`
}

func main() {
	port := getEnv("PORT", "8080")
	critical := getEnv("DISCORD_CRITICAL_URLS", "")
	warning := getEnv("DISCORD_WARNING_URLS", "")
	main := getEnv("DISCORD_MAIN_URLS", "")

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/webhook/critical", makeHandler(critical))
	mux.HandleFunc("/webhook/warning", makeHandler(warning))
	mux.HandleFunc("/webhook/main", makeHandler(main))

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("discord-relay listening on :%s", port)
	log.Fatal(server.ListenAndServe())
}

func makeHandler(urlsEnv string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		urls := splitCSV(urlsEnv)
		if len(urls) == 0 {
			http.Error(w, "no discord webhook urls configured", http.StatusInternalServerError)
			return
		}

		var payload AlertmanagerPayload
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&payload); err != nil {
			http.Error(w, "invalid json payload", http.StatusBadRequest)
			return
		}

		message := buildMessage(payload)
		if message == "" {
			http.Error(w, "empty message", http.StatusBadRequest)
			return
		}

		for _, url := range urls {
			if err := postDiscord(url, message); err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

func buildMessage(payload AlertmanagerPayload) string {
	status := payload.Status
	if status == "" {
		status = "firing"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Alertmanager %s (%d alerts)\n", status, len(payload.Alerts)))

	maxAlerts := 15
	for i, alert := range payload.Alerts {
		if i >= maxAlerts {
			sb.WriteString(fmt.Sprintf("- ...and %d more\n", len(payload.Alerts)-maxAlerts))
			break
		}
		name := firstNonEmpty(alert.Labels["alertname"], "unknown")
		severity := firstNonEmpty(alert.Labels["severity"], "unknown")
		summary := firstNonEmpty(alert.Annotations["summary"], alert.Annotations["description"], name)
		sb.WriteString(fmt.Sprintf("- %s (%s): %s\n", name, severity, summary))
	}

	return truncate(sb.String(), 1900)
}

func postDiscord(url, message string) error {
	payload := DiscordMessage{Content: message}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal discord message: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("post to discord: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord webhook returned %s", resp.Status)
	}

	return nil
}

func splitCSV(value string) []string {
	var out []string
	for _, item := range strings.Split(value, ",") {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max-1] + "\u2026"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
