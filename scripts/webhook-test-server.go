// Test webhook server for AlertManager integration testing
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// AlertWebhook represents an AlertManager webhook payload
type AlertWebhook struct {
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	Status            string            `json:"status"`
	Receiver          string            `json:"receiver"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Alerts            []Alert           `json:"alerts"`
}

// Alert represents a single alert in the webhook
type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
}

func main() {
	// Start webhook test server
	http.HandleFunc("/webhook", handleWebhook)
	http.HandleFunc("/slack", handleSlack)
	http.HandleFunc("/pagerduty", handlePagerDuty)
	http.HandleFunc("/health", handleHealth)

	log.Println("Starting webhook test server on :9999")
	log.Println("Available endpoints:")
	log.Println("  - POST /webhook    - Generic AlertManager webhook")
	log.Println("  - POST /slack      - Slack webhook simulation")
	log.Println("  - POST /pagerduty  - PagerDuty webhook simulation")
	log.Println("  - GET  /health     - Health check")

	if err := http.ListenAndServe(":9999", nil); err != nil {
		log.Fatal("Failed to start webhook test server:", err)
	}
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	var webhook AlertWebhook
	if err := json.Unmarshal(body, &webhook); err != nil {
		log.Printf("Failed to parse webhook payload: %v", err)
		log.Printf("Raw payload: %s", string(body))
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	log.Printf("=== ALERT WEBHOOK RECEIVED ===")
	log.Printf("Status: %s", webhook.Status)
	log.Printf("Receiver: %s", webhook.Receiver)
	log.Printf("Group Key: %s", webhook.GroupKey)
	log.Printf("Group Labels: %+v", webhook.GroupLabels)
	log.Printf("Common Labels: %+v", webhook.CommonLabels)
	log.Printf("Common Annotations: %+v", webhook.CommonAnnotations)

	for i, alert := range webhook.Alerts {
		log.Printf("--- Alert %d ---", i+1)
		log.Printf("  Status: %s", alert.Status)
		log.Printf("  Labels: %+v", alert.Labels)
		log.Printf("  Annotations: %+v", alert.Annotations)
		log.Printf("  Starts At: %s", alert.StartsAt.Format(time.RFC3339))
		if !alert.EndsAt.IsZero() {
			log.Printf("  Ends At: %s", alert.EndsAt.Format(time.RFC3339))
		}
		log.Printf("  Generator URL: %s", alert.GeneratorURL)
	}
	log.Printf("===============================")

	// Respond with success
	response := map[string]string{
		"status":  "received",
		"message": fmt.Sprintf("Processed %d alerts", len(webhook.Alerts)),
		"time":    time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleSlack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	log.Printf("=== SLACK WEBHOOK RECEIVED ===")
	log.Printf("Payload: %s", string(body))
	log.Printf("==============================")

	// Simulate Slack response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func handlePagerDuty(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	log.Printf("=== PAGERDUTY WEBHOOK RECEIVED ===")
	log.Printf("Payload: %s", string(body))
	log.Printf("==================================")

	// Simulate PagerDuty response
	response := map[string]interface{}{
		"status":   "success",
		"message":  "Event processed",
		"dedup_key": fmt.Sprintf("test-%d", time.Now().Unix()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    time.Since(time.Now()).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}