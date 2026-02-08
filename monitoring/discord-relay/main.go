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
	Content  string         `json:"content,omitempty"`
	Embeds   []DiscordEmbed `json:"embeds"`
	Username string         `json:"username,omitempty"`
}

type DiscordEmbed struct {
	Title       string              `json:"title"`
	Description string              `json:"description,omitempty"`
	Color       int                 `json:"color"`
	Fields      []DiscordEmbedField `json:"fields,omitempty"`
	Footer      *DiscordEmbedFooter `json:"footer,omitempty"`
	Timestamp   string              `json:"timestamp,omitempty"`
}

type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type DiscordEmbedFooter struct {
	Text string `json:"text"`
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

		if len(payload.Alerts) == 0 {
			http.Error(w, "no alerts in payload", http.StatusBadRequest)
			return
		}

		alertJSON, err := json.Marshal(payload)
		if err != nil {
			http.Error(w, "marshal alerts: "+err.Error(), http.StatusInternalServerError)
			return
		}

		for _, url := range urls {
			if err := postDiscord(url, string(alertJSON)); err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

func buildEmbeds(payload AlertmanagerPayload) []DiscordEmbed {
	var embeds []DiscordEmbed
	status := firstNonEmpty(payload.Status, "firing")
	isResolved := status == "resolved"

	maxAlerts := 10
	alertCount := len(payload.Alerts)
	if alertCount > maxAlerts {
		alertCount = maxAlerts
	}

	for i := 0; i < alertCount; i++ {
		alert := payload.Alerts[i]
		name := firstNonEmpty(alert.Labels["alertname"], "unknown")
		severity := firstNonEmpty(alert.Labels["severity"], "warning")
		summary := firstNonEmpty(alert.Annotations["summary"], alert.Annotations["description"])
		description := firstNonEmpty(alert.Annotations["description"])
		instance := firstNonEmpty(alert.Labels["instance"], alert.Labels["pod"], alert.Labels["service"])
		category := firstNonEmpty(alert.Labels["category"], "general")

		color := severityToColor(severity, isResolved)
		emoji := getEmoji(severity)

		embed := DiscordEmbed{
			Title:     fmt.Sprintf("%s %s [%s]", emoji, name, strings.ToUpper(severity)),
			Color:     color,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		if summary != "" {
			embed.Description = summary
		}

		var fields []DiscordEmbedField

		if category != "" && category != "general" {
			fields = append(fields, DiscordEmbedField{
				Name:   "Category",
				Value:  strings.Title(category),
				Inline: true,
			})
		}

		if severity != "" {
			fields = append(fields, DiscordEmbedField{
				Name:   "Severity",
				Value:  strings.ToUpper(severity),
				Inline: true,
			})
		}

		if description != "" && description != summary {
			fields = append(fields, DiscordEmbedField{
				Name:   "Details",
				Value:  description,
				Inline: false,
			})
		}

		if instance != "" {
			fields = append(fields, DiscordEmbedField{
				Name:   "Instance",
				Value:  instance,
				Inline: true,
			})
		}

		embed.Fields = fields
		embed.Footer = &DiscordEmbedFooter{
			Text: "Alertmanager - Monitoring System",
		}

		embeds = append(embeds, embed)
	}

	if len(payload.Alerts) > maxAlerts {
		embed := DiscordEmbed{
			Title:       "⚠️ More Alerts",
			Color:       0xFFA500,
			Description: fmt.Sprintf("Showing %d of %d alerts. %d more not displayed.", maxAlerts, len(payload.Alerts), len(payload.Alerts)-maxAlerts),
		}
		embeds = append(embeds, embed)
	}

	return embeds
}

func severityToColor(severity string, isResolved bool) int {
	if isResolved {
		return 0x27AE60
	}
	switch strings.ToLower(severity) {
	case "critical":
		return 0xD32F2F
	case "warning":
		return 0xFFA500
	case "info":
		return 0x0099FF
	default:
		return 0x808080
	}
}

func getEmoji(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "🔴"
	case "warning":
		return "🟡"
	case "info":
		return "ℹ️"
	default:
		return "⚪"
	}
}

func postDiscord(url, message string) error {
	var alerts AlertmanagerPayload
	err := json.Unmarshal([]byte(message), &alerts)
	if err != nil {
		return fmt.Errorf("unmarshal alerts: %w", err)
	}

	embeds := buildEmbeds(alerts)
	payload := DiscordMessage{
		Embeds:   embeds,
		Username: "Monitoring Alert",
	}
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
