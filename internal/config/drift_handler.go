package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// DriftHandler provides HTTP endpoints for drift detection management
type DriftHandler struct {
	driftDetector *DriftDetector
}

// NewDriftHandler creates a new drift detection handler
func NewDriftHandler(driftDetector *DriftDetector) *DriftHandler {
	return &DriftHandler{
		driftDetector: driftDetector,
	}
}

// HandleGetStatus returns the current drift detection status
func (dh *DriftHandler) HandleGetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := dh.driftDetector.GetStats()
	baseline := dh.driftDetector.GetBaseline()

	response := map[string]interface{}{
		"status":    "active",
		"stats":     stats,
		"baseline":  baseline,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// HandleGetHistory returns the drift detection history
func (dh *DriftHandler) HandleGetHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	severityFilter := r.URL.Query().Get("severity")
	typeFilter := r.URL.Query().Get("type")

	limit := 50 // default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // default offset
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	history := dh.driftDetector.GetHistory()
	
	// Apply filters
	var filteredHistory []DriftEvent
	for _, event := range history {
		// Apply severity filter
		if severityFilter != "" && event.Severity != severityFilter {
			continue
		}
		
		// Apply type filter
		if typeFilter != "" && event.Type != typeFilter {
			continue
		}
		
		filteredHistory = append(filteredHistory, event)
	}

	// Apply pagination
	total := len(filteredHistory)
	
	start := offset
	if start > total {
		start = total
	}
	
	end := start + limit
	if end > total {
		end = total
	}
	
	paginatedHistory := filteredHistory[start:end]

	response := map[string]interface{}{
		"events":    paginatedHistory,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// HandleGetBaseline returns the current configuration baseline
func (dh *DriftHandler) HandleGetBaseline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	baseline := dh.driftDetector.GetBaseline()
	if baseline == nil {
		http.Error(w, "No baseline available", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(baseline); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode baseline: %v", err), http.StatusInternalServerError)
		return
	}
}

// HandleUpdateBaseline manually updates the configuration baseline
func (dh *DriftHandler) HandleUpdateBaseline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dh.driftDetector.UpdateBaseline()

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":    "success",
		"message":   "Configuration baseline updated successfully",
		"timestamp": time.Now(),
	}
	json.NewEncoder(w).Encode(response)
}

// HandleTriggerCheck manually triggers a drift detection check
func (dh *DriftHandler) HandleTriggerCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// This would trigger a manual drift check
	// For now, we'll simulate this by performing the check
	go dh.driftDetector.performDriftCheck()

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":    "success",
		"message":   "Drift detection check triggered",
		"timestamp": time.Now(),
	}
	json.NewEncoder(w).Encode(response)
}

// HandleGetStats returns detailed statistics about drift detection
func (dh *DriftHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := dh.driftDetector.GetStats()
	history := dh.driftDetector.GetHistory()

	// Calculate additional statistics
	recentEvents := 0
	alertsSent := 0
	cutoff := time.Now().Add(-24 * time.Hour)

	for _, event := range history {
		if event.Timestamp.After(cutoff) {
			recentEvents++
		}
		if event.AlertSent {
			alertsSent++
		}
	}

	enhancedStats := map[string]interface{}{
		"basic_stats":    stats,
		"recent_events":  recentEvents,
		"alerts_sent":    alertsSent,
		"alert_rate":     float64(alertsSent) / float64(len(history)) * 100,
		"timestamp":      time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(enhancedStats); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode stats: %v", err), http.StatusInternalServerError)
		return
	}
}

// HandleGetAlerts returns recent alerts
func (dh *DriftHandler) HandleGetAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	history := dh.driftDetector.GetHistory()
	
	// Filter for events that had alerts sent
	var alerts []DriftEvent
	for _, event := range history {
		if event.AlertSent {
			alerts = append(alerts, event)
		}
	}

	// Reverse to show most recent first
	for i, j := 0, len(alerts)-1; i < j; i, j = i+1, j-1 {
		alerts[i], alerts[j] = alerts[j], alerts[i]
	}

	response := map[string]interface{}{
		"alerts":    alerts,
		"count":     len(alerts),
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode alerts: %v", err), http.StatusInternalServerError)
		return
	}
}

// HandleGetHealth returns the health status of the drift detection system
func (dh *DriftHandler) HandleGetHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := dh.driftDetector.GetStats()
	
	// Determine health based on various factors
	isHealthy := true
	issues := make([]string, 0)
	
	// Check if detector is running
	if running, ok := stats["running"].(bool); !ok || !running {
		isHealthy = false
		issues = append(issues, "drift detector is not running")
	}
	
	// Check if baseline exists
	if stats["baseline_hash"] == nil {
		isHealthy = false
		issues = append(issues, "no configuration baseline exists")
	}
	
	// Check last check time
	if lastCheck, ok := stats["last_check"].(time.Time); ok {
		if time.Since(lastCheck) > 2*time.Hour {
			isHealthy = false
			issues = append(issues, "drift detection checks are stale")
		}
	}

	status := "healthy"
	statusCode := http.StatusOK
	
	if !isHealthy {
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := map[string]interface{}{
		"status":     status,
		"healthy":    isHealthy,
		"issues":     issues,
		"stats":      stats,
		"timestamp":  time.Now(),
	}
	
	json.NewEncoder(w).Encode(response)
}

// HandleExportBaseline exports the current baseline configuration
func (dh *DriftHandler) HandleExportBaseline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	baseline := dh.driftDetector.GetBaseline()
	if baseline == nil {
		http.Error(w, "No baseline available", http.StatusNotFound)
		return
	}

	// Set headers for file download
	filename := fmt.Sprintf("config-baseline-%s.json", baseline.Hash[:8])
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	if err := json.NewEncoder(w).Encode(baseline); err != nil {
		http.Error(w, fmt.Sprintf("Failed to export baseline: %v", err), http.StatusInternalServerError)
		return
	}
}

// HandleCompareSnapshots compares two configuration snapshots
func (dh *DriftHandler) HandleCompareSnapshots(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		OldSnapshot *ConfigSnapshot `json:"old_snapshot"`
		NewSnapshot *ConfigSnapshot `json:"new_snapshot"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if request.OldSnapshot == nil || request.NewSnapshot == nil {
		http.Error(w, "Both old_snapshot and new_snapshot are required", http.StatusBadRequest)
		return
	}

	differences := CompareConfigs(request.OldSnapshot.Config, request.NewSnapshot.Config)

	response := map[string]interface{}{
		"differences": differences,
		"count":       len(differences),
		"old_hash":    request.OldSnapshot.Hash,
		"new_hash":    request.NewSnapshot.Hash,
		"timestamp":   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode comparison: %v", err), http.StatusInternalServerError)
		return
	}
}

// RegisterRoutes registers all drift detection routes
func (dh *DriftHandler) RegisterRoutes(mux *http.ServeMux, pathPrefix string) {
	if pathPrefix == "" {
		pathPrefix = "/api/drift"
	}

	mux.HandleFunc(pathPrefix+"/status", dh.HandleGetStatus)
	mux.HandleFunc(pathPrefix+"/history", dh.HandleGetHistory)
	mux.HandleFunc(pathPrefix+"/baseline", dh.HandleGetBaseline)
	mux.HandleFunc(pathPrefix+"/baseline/update", dh.HandleUpdateBaseline)
	mux.HandleFunc(pathPrefix+"/baseline/export", dh.HandleExportBaseline)
	mux.HandleFunc(pathPrefix+"/check", dh.HandleTriggerCheck)
	mux.HandleFunc(pathPrefix+"/stats", dh.HandleGetStats)
	mux.HandleFunc(pathPrefix+"/alerts", dh.HandleGetAlerts)
	mux.HandleFunc(pathPrefix+"/health", dh.HandleGetHealth)
	mux.HandleFunc(pathPrefix+"/compare", dh.HandleCompareSnapshots)
}

// DriftMiddleware provides drift-aware middleware
type DriftMiddleware struct {
	driftDetector *DriftDetector
}

// NewDriftMiddleware creates a new drift detection middleware
func NewDriftMiddleware(driftDetector *DriftDetector) *DriftMiddleware {
	return &DriftMiddleware{
		driftDetector: driftDetector,
	}
}

// AlertOnDrift creates middleware that adds drift detection headers
func (dm *DriftMiddleware) AlertOnDrift() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stats := dm.driftDetector.GetStats()
			baseline := dm.driftDetector.GetBaseline()

			if baseline != nil {
				w.Header().Set("X-Config-Baseline-Hash", baseline.Hash)
				w.Header().Set("X-Config-Baseline-Timestamp", baseline.Timestamp.Format(time.RFC3339))
			}

			if lastCheck, ok := stats["last_check"].(time.Time); ok {
				w.Header().Set("X-Config-Last-Check", lastCheck.Format(time.RFC3339))
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireStableConfig creates middleware that rejects requests during configuration drift
func (dm *DriftMiddleware) RequireStableConfig() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			history := dm.driftDetector.GetHistory()
			
			// Check for recent critical drift events
			cutoff := time.Now().Add(-5 * time.Minute)
			for i := len(history) - 1; i >= 0; i-- {
				event := history[i]
				if event.Timestamp.Before(cutoff) {
					break
				}
				
				if event.Type == "drift_detected" && event.Severity == "critical" {
					http.Error(w, "Service temporarily unavailable due to configuration drift", http.StatusServiceUnavailable)
					return
				}
			}
			
			next.ServeHTTP(w, r)
		})
	}
}