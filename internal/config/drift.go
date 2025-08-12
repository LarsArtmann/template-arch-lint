package config

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// DriftDetector monitors configuration drift and sends alerts
type DriftDetector struct {
	name            string
	reloadableConfig *ReloadableConfig
	baseline        *ConfigSnapshot
	checkInterval   time.Duration
	alertThreshold  time.Duration
	alerters        []Alerter
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	running         bool
	lastCheck       time.Time
	driftHistory    []DriftEvent
	maxHistorySize  int
}

// ConfigSnapshot represents a snapshot of configuration at a point in time
type ConfigSnapshot struct {
	Timestamp    time.Time              `json:"timestamp"`
	Environment  string                 `json:"environment"`
	Version      string                 `json:"version"`
	Hash         string                 `json:"hash"`
	Config       *Config                `json:"config"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// DriftEvent represents a configuration drift event
type DriftEvent struct {
	ID           string        `json:"id"`
	Timestamp    time.Time     `json:"timestamp"`
	Type         string        `json:"type"` // "drift_detected", "drift_resolved", "check_failed"
	Severity     string        `json:"severity"` // "low", "medium", "high", "critical"
	Description  string        `json:"description"`
	Changes      []ConfigDiff  `json:"changes"`
	OldSnapshot  *ConfigSnapshot `json:"old_snapshot"`
	NewSnapshot  *ConfigSnapshot `json:"new_snapshot"`
	AlertSent    bool          `json:"alert_sent"`
}

// DriftAlert represents an alert about configuration drift
type DriftAlert struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"` // "info", "warning", "error", "critical"
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details"`
	Event       *DriftEvent            `json:"event"`
	Recipients  []string               `json:"recipients,omitempty"`
}

// Alerter defines the interface for sending alerts
type Alerter interface {
	SendAlert(ctx context.Context, alert DriftAlert) error
	IsAvailable(ctx context.Context) bool
	GetName() string
	Close() error
}

// NewDriftDetector creates a new configuration drift detector
func NewDriftDetector(name string, reloadableConfig *ReloadableConfig, options ...DriftDetectorOption) *DriftDetector {
	ctx, cancel := context.WithCancel(context.Background())
	
	dd := &DriftDetector{
		name:            name,
		reloadableConfig: reloadableConfig,
		checkInterval:   30 * time.Second,
		alertThreshold:  5 * time.Minute,
		alerters:        make([]Alerter, 0),
		ctx:             ctx,
		cancel:          cancel,
		driftHistory:    make([]DriftEvent, 0),
		maxHistorySize:  100,
	}

	// Apply options
	for _, option := range options {
		option(dd)
	}

	// Create initial baseline
	dd.createBaseline()

	return dd
}

// DriftDetectorOption defines options for configuring the drift detector
type DriftDetectorOption func(*DriftDetector)

// WithCheckInterval sets the drift check interval
func WithCheckInterval(interval time.Duration) DriftDetectorOption {
	return func(dd *DriftDetector) {
		dd.checkInterval = interval
	}
}

// WithAlertThreshold sets the threshold for sending alerts
func WithAlertThreshold(threshold time.Duration) DriftDetectorOption {
	return func(dd *DriftDetector) {
		dd.alertThreshold = threshold
	}
}

// WithAlerter adds an alerter to the drift detector
func WithAlerter(alerter Alerter) DriftDetectorOption {
	return func(dd *DriftDetector) {
		dd.alerters = append(dd.alerters, alerter)
	}
}

// WithMaxHistorySize sets the maximum size of drift history
func WithMaxHistorySize(size int) DriftDetectorOption {
	return func(dd *DriftDetector) {
		dd.maxHistorySize = size
	}
}

// createBaseline creates an initial configuration baseline
func (dd *DriftDetector) createBaseline() {
	config := dd.reloadableConfig.GetConfig()
	snapshot := dd.createSnapshot(config)
	
	dd.mu.Lock()
	dd.baseline = snapshot
	dd.mu.Unlock()
	
	fmt.Printf("Created configuration baseline for '%s' (hash: %s)\\n", dd.name, snapshot.Hash)
}

// createSnapshot creates a configuration snapshot
func (dd *DriftDetector) createSnapshot(config *Config) *ConfigSnapshot {
	// Create a deep copy for the snapshot
	configBytes, _ := json.Marshal(config)
	var configCopy Config
	json.Unmarshal(configBytes, &configCopy)
	
	// Calculate hash of the configuration
	hash := dd.calculateConfigHash(&configCopy)
	
	return &ConfigSnapshot{
		Timestamp:   time.Now(),
		Environment: config.App.Environment,
		Version:     config.App.Version,
		Hash:        hash,
		Config:      &configCopy,
		Metadata: map[string]interface{}{
			"detector": dd.name,
			"source":   "drift_detection",
		},
	}
}

// calculateConfigHash calculates a hash of the configuration for drift detection
func (dd *DriftDetector) calculateConfigHash(config *Config) string {
	// Serialize config to JSON for hashing (excluding dynamic fields)
	configForHashing := struct {
		Server        ServerConfig        `json:"server"`
		Database      DatabaseConfig      `json:"database"`
		Logging       LoggingConfig       `json:"logging"`
		App           AppConfig           `json:"app"`
		Observability ObservabilityConfig `json:"observability"`
		Features      FeaturesConfig      `json:"features"`
		Security      SecurityConfig      `json:"security"`
		Health        HealthConfig        `json:"health"`
		Cache         CacheConfig         `json:"cache"`
		External      ExternalConfig      `json:"external"`
		Backup        BackupConfig        `json:"backup"`
		Resources     ResourcesConfig     `json:"resources"`
	}{
		Server:        config.Server,
		Database:      config.Database,
		Logging:       config.Logging,
		App:           config.App,
		Observability: config.Observability,
		Features:      config.Features,
		Security:      config.Security,
		Health:        config.Health,
		Cache:         config.Cache,
		External:      config.External,
		Backup:        config.Backup,
		Resources:     config.Resources,
	}

	configBytes, _ := json.Marshal(configForHashing)
	hash := md5.Sum(configBytes)
	return hex.EncodeToString(hash[:])
}

// Start starts the drift detection process
func (dd *DriftDetector) Start() error {
	dd.mu.Lock()
	if dd.running {
		dd.mu.Unlock()
		return fmt.Errorf("drift detector is already running")
	}
	dd.running = true
	dd.mu.Unlock()

	// Start the monitoring goroutine
	go dd.monitorLoop()

	fmt.Printf("Started drift detection for '%s' (check interval: %v)\\n", dd.name, dd.checkInterval)
	return nil
}

// Stop stops the drift detection process
func (dd *DriftDetector) Stop() error {
	dd.mu.Lock()
	if !dd.running {
		dd.mu.Unlock()
		return fmt.Errorf("drift detector is not running")
	}
	dd.running = false
	dd.mu.Unlock()

	if dd.cancel != nil {
		dd.cancel()
	}

	// Close all alerters
	for _, alerter := range dd.alerters {
		alerter.Close()
	}

	fmt.Printf("Stopped drift detection for '%s'\\n", dd.name)
	return nil
}

// monitorLoop runs the continuous monitoring loop
func (dd *DriftDetector) monitorLoop() {
	ticker := time.NewTicker(dd.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			dd.performDriftCheck()
		case <-dd.ctx.Done():
			return
		}
	}
}

// performDriftCheck performs a single drift check
func (dd *DriftDetector) performDriftCheck() {
	dd.mu.Lock()
	dd.lastCheck = time.Now()
	dd.mu.Unlock()

	config := dd.reloadableConfig.GetConfig()
	currentSnapshot := dd.createSnapshot(config)

	dd.mu.RLock()
	baseline := dd.baseline
	dd.mu.RUnlock()

	// Check for drift
	if currentSnapshot.Hash != baseline.Hash {
		dd.handleDrift(baseline, currentSnapshot)
	}
}

// handleDrift handles detected configuration drift
func (dd *DriftDetector) handleDrift(oldSnapshot, newSnapshot *ConfigSnapshot) {
	// Calculate differences
	differences := CompareConfigs(oldSnapshot.Config, newSnapshot.Config)
	
	// Determine severity
	severity := dd.calculateSeverity(differences)
	
	// Create drift event
	event := DriftEvent{
		ID:          fmt.Sprintf("drift_%d", time.Now().UnixNano()),
		Timestamp:   time.Now(),
		Type:        "drift_detected",
		Severity:    severity,
		Description: fmt.Sprintf("Configuration drift detected in '%s' with %d changes", dd.name, len(differences)),
		Changes:     differences,
		OldSnapshot: oldSnapshot,
		NewSnapshot: newSnapshot,
		AlertSent:   false,
	}

	// Add to history
	dd.addToHistory(event)

	// Send alerts if necessary
	if dd.shouldSendAlert(event) {
		dd.sendAlert(event)
		event.AlertSent = true
	}

	// Update baseline if this represents a legitimate change
	if dd.shouldUpdateBaseline(event) {
		dd.mu.Lock()
		dd.baseline = newSnapshot
		dd.mu.Unlock()
		fmt.Printf("Updated configuration baseline for '%s' (new hash: %s)\\n", dd.name, newSnapshot.Hash)
	}

	fmt.Printf("Configuration drift detected in '%s': %d changes (severity: %s)\\n", 
		dd.name, len(differences), severity)
}

// calculateSeverity calculates the severity of drift based on changes
func (dd *DriftDetector) calculateSeverity(differences []ConfigDiff) string {
	criticalFields := []string{
		"database.dsn", "database.driver",
		"app.environment",
		"security.tls",
		"observability.exporters",
	}
	
	highImpactFields := []string{
		"server.port", "server.host",
		"logging.level",
		"features",
		"security",
	}

	criticalChanges := 0
	highImpactChanges := 0

	for _, diff := range differences {
		for _, criticalField := range criticalFields {
			if strings.HasPrefix(diff.Field, criticalField) {
				criticalChanges++
				break
			}
		}
		
		for _, highField := range highImpactFields {
			if strings.HasPrefix(diff.Field, highField) {
				highImpactChanges++
				break
			}
		}
	}

	if criticalChanges > 0 {
		return "critical"
	}
	if highImpactChanges > 2 {
		return "high"
	}
	if len(differences) > 5 {
		return "medium"
	}
	return "low"
}

// shouldSendAlert determines if an alert should be sent for the drift event
func (dd *DriftDetector) shouldSendAlert(event DriftEvent) bool {
	// Always alert on critical changes
	if event.Severity == "critical" {
		return true
	}
	
	// Alert on high severity if multiple changes
	if event.Severity == "high" && len(event.Changes) > 1 {
		return true
	}
	
	// Rate limit alerts for lower severity events
	dd.mu.RLock()
	defer dd.mu.RUnlock()
	
	recentAlerts := 0
	cutoff := time.Now().Add(-dd.alertThreshold)
	
	for i := len(dd.driftHistory) - 1; i >= 0; i-- {
		historyEvent := dd.driftHistory[i]
		if historyEvent.Timestamp.Before(cutoff) {
			break
		}
		if historyEvent.AlertSent {
			recentAlerts++
		}
	}
	
	// Limit to 1 alert per threshold period for non-critical events
	return recentAlerts == 0
}

// shouldUpdateBaseline determines if the baseline should be updated
func (dd *DriftDetector) shouldUpdateBaseline(event DriftEvent) bool {
	// Update baseline for configuration reloads (legitimate changes)
	// Don't update for unexpected drift
	
	// Check if this change came from a legitimate configuration reload
	// This is simplified - in reality you'd track configuration sources
	return event.Severity != "critical"
}

// sendAlert sends alerts to all configured alerters
func (dd *DriftDetector) sendAlert(event DriftEvent) {
	alert := DriftAlert{
		ID:        fmt.Sprintf("alert_%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Level:     dd.mapSeverityToLevel(event.Severity),
		Title:     fmt.Sprintf("Configuration Drift Detected: %s", dd.name),
		Message:   event.Description,
		Details: map[string]interface{}{
			"detector":      dd.name,
			"changes_count": len(event.Changes),
			"environment":   event.NewSnapshot.Environment,
			"old_hash":      event.OldSnapshot.Hash,
			"new_hash":      event.NewSnapshot.Hash,
		},
		Event: &event,
	}

	for _, alerter := range dd.alerters {
		if alerter.IsAvailable(dd.ctx) {
			go func(a Alerter) {
				if err := a.SendAlert(dd.ctx, alert); err != nil {
					fmt.Printf("Failed to send alert via %s: %v\\n", a.GetName(), err)
				}
			}(alerter)
		}
	}
}

// mapSeverityToLevel maps drift severity to alert level
func (dd *DriftDetector) mapSeverityToLevel(severity string) string {
	switch severity {
	case "critical":
		return "critical"
	case "high":
		return "error"
	case "medium":
		return "warning"
	case "low":
		return "info"
	default:
		return "info"
	}
}

// addToHistory adds a drift event to the history
func (dd *DriftDetector) addToHistory(event DriftEvent) {
	dd.mu.Lock()
	defer dd.mu.Unlock()

	dd.driftHistory = append(dd.driftHistory, event)

	// Trim history if it exceeds maximum size
	if len(dd.driftHistory) > dd.maxHistorySize {
		dd.driftHistory = dd.driftHistory[1:]
	}
}

// GetHistory returns the drift detection history
func (dd *DriftDetector) GetHistory() []DriftEvent {
	dd.mu.RLock()
	defer dd.mu.RUnlock()

	// Return a copy
	history := make([]DriftEvent, len(dd.driftHistory))
	copy(history, dd.driftHistory)
	return history
}

// GetBaseline returns the current baseline configuration
func (dd *DriftDetector) GetBaseline() *ConfigSnapshot {
	dd.mu.RLock()
	defer dd.mu.RUnlock()

	if dd.baseline == nil {
		return nil
	}

	// Return a copy
	baselineCopy := *dd.baseline
	return &baselineCopy
}

// UpdateBaseline manually updates the baseline configuration
func (dd *DriftDetector) UpdateBaseline() {
	config := dd.reloadableConfig.GetConfig()
	newBaseline := dd.createSnapshot(config)

	dd.mu.Lock()
	oldBaseline := dd.baseline
	dd.baseline = newBaseline
	dd.mu.Unlock()

	// Create an event for the baseline update
	event := DriftEvent{
		ID:          fmt.Sprintf("baseline_update_%d", time.Now().UnixNano()),
		Timestamp:   time.Now(),
		Type:        "baseline_updated",
		Severity:    "info",
		Description: fmt.Sprintf("Configuration baseline updated for '%s'", dd.name),
		Changes:     CompareConfigs(oldBaseline.Config, newBaseline.Config),
		OldSnapshot: oldBaseline,
		NewSnapshot: newBaseline,
		AlertSent:   false,
	}

	dd.addToHistory(event)
	fmt.Printf("Updated configuration baseline for '%s' (hash: %s -> %s)\\n", 
		dd.name, oldBaseline.Hash, newBaseline.Hash)
}

// GetStats returns statistics about the drift detector
func (dd *DriftDetector) GetStats() map[string]interface{} {
	dd.mu.RLock()
	defer dd.mu.RUnlock()

	stats := map[string]interface{}{
		"name":             dd.name,
		"running":          dd.running,
		"check_interval":   dd.checkInterval,
		"last_check":       dd.lastCheck,
		"history_size":     len(dd.driftHistory),
		"max_history_size": dd.maxHistorySize,
		"alerters_count":   len(dd.alerters),
	}

	if dd.baseline != nil {
		stats["baseline_hash"] = dd.baseline.Hash
		stats["baseline_timestamp"] = dd.baseline.Timestamp
		stats["baseline_environment"] = dd.baseline.Environment
	}

	// Count events by type and severity
	eventCounts := make(map[string]int)
	severityCounts := make(map[string]int)

	for _, event := range dd.driftHistory {
		eventCounts[event.Type]++
		severityCounts[event.Severity]++
	}

	stats["event_counts"] = eventCounts
	stats["severity_counts"] = severityCounts

	return stats
}

// LogAlerter implements the Alerter interface for logging alerts
type LogAlerter struct {
	name string
}

// NewLogAlerter creates a new log-based alerter
func NewLogAlerter() *LogAlerter {
	return &LogAlerter{name: "log"}
}

func (la *LogAlerter) SendAlert(ctx context.Context, alert DriftAlert) error {
	fmt.Printf("[%s] DRIFT ALERT: %s - %s\\n", alert.Level, alert.Title, alert.Message)
	if alert.Event != nil {
		fmt.Printf("  Changes: %d, Severity: %s\\n", len(alert.Event.Changes), alert.Event.Severity)
		for _, change := range alert.Event.Changes {
			fmt.Printf("    %s: %v -> %v (%s)\\n", 
				change.Field, change.OldValue, change.NewValue, change.Action)
		}
	}
	return nil
}

func (la *LogAlerter) IsAvailable(ctx context.Context) bool {
	return true
}

func (la *LogAlerter) GetName() string {
	return la.name
}

func (la *LogAlerter) Close() error {
	return nil
}

// EmailAlerter implements the Alerter interface for email alerts (mock implementation)
type EmailAlerter struct {
	name       string
	smtpServer string
	from       string
	to         []string
}

// NewEmailAlerter creates a new email-based alerter
func NewEmailAlerter(smtpServer, from string, to []string) *EmailAlerter {
	return &EmailAlerter{
		name:       "email",
		smtpServer: smtpServer,
		from:       from,
		to:         to,
	}
}

func (ea *EmailAlerter) SendAlert(ctx context.Context, alert DriftAlert) error {
	// Mock implementation - in reality would send actual email
	fmt.Printf("EMAIL ALERT to %v: %s\\n", ea.to, alert.Title)
	return nil
}

func (ea *EmailAlerter) IsAvailable(ctx context.Context) bool {
	// In reality, would check SMTP server connectivity
	return ea.smtpServer != ""
}

func (ea *EmailAlerter) GetName() string {
	return ea.name
}

func (ea *EmailAlerter) Close() error {
	return nil
}

// SlackAlerter implements the Alerter interface for Slack alerts (mock implementation)
type SlackAlerter struct {
	name    string
	webhook string
	channel string
}

// NewSlackAlerter creates a new Slack-based alerter
func NewSlackAlerter(webhook, channel string) *SlackAlerter {
	return &SlackAlerter{
		name:    "slack",
		webhook: webhook,
		channel: channel,
	}
}

func (sa *SlackAlerter) SendAlert(ctx context.Context, alert DriftAlert) error {
	// Mock implementation - in reality would send to Slack API
	fmt.Printf("SLACK ALERT to %s: %s\\n", sa.channel, alert.Title)
	return nil
}

func (sa *SlackAlerter) IsAvailable(ctx context.Context) bool {
	// In reality, would check Slack API connectivity
	return sa.webhook != ""
}

func (sa *SlackAlerter) GetName() string {
	return sa.name
}

func (sa *SlackAlerter) Close() error {
	return nil
}