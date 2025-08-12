package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LarsArtmann/template-arch-lint/internal/config"
)

// Example application demonstrating comprehensive configuration management
func main() {
	fmt.Println("Starting Configuration Management Demo")
	
	// 1. Load Environment-Specific Configuration
	environment := os.Getenv("APP_ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}
	
	fmt.Printf("Loading configuration for environment: %s\n", environment)
	
	reloadableConfig, err := config.NewReloadableConfig(fmt.Sprintf("configs/%s.yaml", environment))
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}
	defer reloadableConfig.Close()
	
	cfg := reloadableConfig.GetConfig()
	fmt.Printf("✓ Configuration loaded: %s v%s (%s)\n", cfg.App.Name, cfg.App.Version, cfg.App.Environment)
	
	// 2. Initialize Secrets Management
	fmt.Println("\nInitializing secrets management...")
	
	secretConfig := config.SecretConfig{
		Provider: "env", // Use environment variables as primary source
		CacheTTL: 5 * time.Minute,
		FileConfig: config.FileConfig{
			Path:   "examples/config-integration/secrets.json",
			Format: "json",
		},
	}
	
	secretsManager, err := config.NewSecretsManager(secretConfig)
	if err != nil {
		log.Fatal("Failed to initialize secrets manager:", err)
	}
	defer secretsManager.Close()
	
	// Expand secrets in configuration
	if err := config.ExpandSecrets(cfg, secretsManager); err != nil {
		fmt.Printf("Warning: Could not expand all secrets: %v\n", err)
	} else {
		fmt.Println("✓ Secrets expanded in configuration")
	}
	
	// 3. Initialize Feature Flags
	fmt.Println("\nInitializing feature flags...")
	
	featureManager := reloadableConfig.GetFeatureManager()
	if featureManager != nil {
		stats := featureManager.GetStats()
		fmt.Printf("✓ Feature flags initialized: %d total flags, %d enabled\n", 
			stats["total_flags"], stats["enabled_flags"])
		
		// Demonstrate feature flag usage
		demonstrateFeatureFlags(featureManager)
	}
	
	// 4. Initialize Configuration Drift Detection
	fmt.Println("\nInitializing drift detection...")
	
	driftDetector := config.NewDriftDetector(
		"demo-app",
		reloadableConfig,
		config.WithCheckInterval(15*time.Second),
		config.WithAlertThreshold(2*time.Minute),
		config.WithAlerter(config.NewLogAlerter()),
	)
	
	if err := driftDetector.Start(); err != nil {
		log.Fatal("Failed to start drift detector:", err)
	}
	defer driftDetector.Stop()
	
	fmt.Println("✓ Drift detection started")
	
	// 5. Setup Configuration Change Monitoring
	setupConfigurationMonitoring(reloadableConfig, driftDetector)
	
	// 6. Setup HTTP Server with Configuration APIs
	mux := setupHTTPServer(reloadableConfig, secretsManager, driftDetector)
	
	// 7. Start HTTP Server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}
	
	fmt.Printf("\n🚀 Server starting on %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	fmt.Println("\nAvailable endpoints:")
	fmt.Println("  • Configuration: http://localhost:8080/api/config/")
	fmt.Println("  • Feature Flags: http://localhost:8080/api/config/flags")
	fmt.Println("  • Secrets:       http://localhost:8080/api/secrets/health")
	fmt.Println("  • Drift Status:  http://localhost:8080/api/drift/status")
	fmt.Println("  • Health Check:  http://localhost:8080/health")
	fmt.Println("\nPress Ctrl+C to stop...")
	
	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()
	
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	fmt.Println("\n🛑 Shutting down server...")
	
	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.GracefulShutdownTimeout)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	
	fmt.Println("✓ Server shutdown complete")
}

// demonstrateFeatureFlags shows how to use feature flags
func demonstrateFeatureFlags(featureManager *config.FeatureManager) {
	// Basic feature flag check
	if featureManager.IsEnabled("debug_endpoints") {
		fmt.Println("  ✓ Debug endpoints are enabled")
	} else {
		fmt.Println("  ✗ Debug endpoints are disabled")
	}
	
	// Feature flag with context
	ctx := config.FeatureContext{
		UserID:      "demo-user-123",
		Environment: "development",
		Timestamp:   time.Now(),
		Properties: map[string]string{
			"region": "us-east-1",
			"tier":   "premium",
		},
	}
	
	if featureManager.IsEnabledForContext("beta_features", ctx) {
		fmt.Println("  ✓ Beta features enabled for demo user")
	} else {
		fmt.Println("  ✗ Beta features disabled for demo user")
	}
	
	// Create a dynamic feature flag
	dynamicFlag := &config.FeatureFlag{
		Name:        "dynamic_demo",
		Enabled:     true,
		Description: "Demonstration of dynamic feature flag",
		Conditions: []config.FeatureCondition{
			{
				Type:     "percentage",
				Operator: "percentage",
				Values:   []interface{}{75.0}, // 75% rollout
			},
		},
		Metadata: map[string]interface{}{
			"created_by": "demo",
			"team":       "platform",
		},
	}
	
	if err := featureManager.UpdateFlag(dynamicFlag); err != nil {
		fmt.Printf("  ✗ Failed to create dynamic flag: %v\n", err)
	} else {
		fmt.Println("  ✓ Created dynamic feature flag with 75% rollout")
	}
}

// setupConfigurationMonitoring sets up monitoring for configuration changes
func setupConfigurationMonitoring(reloadableConfig *config.ReloadableConfig, driftDetector *config.DriftDetector) {
	// Subscribe to configuration changes
	configChanges := make(chan config.ConfigChange, 10)
	reloadableConfig.Subscribe(configChanges)
	
	// Subscribe to feature flag updates
	featureUpdates := make(chan config.FeatureUpdate, 10)
	featureManager := reloadableConfig.GetFeatureManager()
	if featureManager != nil {
		featureManager.Subscribe(featureUpdates)
	}
	
	// Monitor configuration changes
	go func() {
		for {
			select {
			case change := <-configChanges:
				fmt.Printf("\n📝 Configuration Change: %s\n", change.Type)
				if len(change.Differences) > 0 {
					fmt.Printf("   %d changes detected:\n", len(change.Differences))
					for i, diff := range change.Differences {
						if i >= 3 { // Limit output
							fmt.Printf("   ... and %d more changes\n", len(change.Differences)-3)
							break
						}
						fmt.Printf("   • %s: %v → %v (%s)\n", 
							diff.Field, diff.OldValue, diff.NewValue, diff.Action)
					}
				}
				
			case update := <-featureUpdates:
				fmt.Printf("\n🎯 Feature Flag Update: %s %s\n", update.Flag.Name, update.Action)
				fmt.Printf("   Enabled: %t, Description: %s\n", update.Flag.Enabled, update.Flag.Description)
			}
		}
	}()
	
	fmt.Println("✓ Configuration change monitoring active")
}

// setupHTTPServer sets up the HTTP server with all configuration APIs
func setupHTTPServer(reloadableConfig *config.ReloadableConfig, secretsManager *config.SecretsManager, driftDetector *config.DriftDetector) *http.ServeMux {
	mux := http.NewServeMux()
	
	// Register configuration management routes
	configHandler := config.NewRuntimeConfigHandler(reloadableConfig)
	configHandler.RegisterRoutes(mux, "/api/config")
	
	// Register secrets management routes
	secretsHandler := config.NewSecretsHandler(secretsManager)
	secretsHandler.RegisterRoutes(mux, "/api/secrets")
	
	// Register drift detection routes
	driftHandler := config.NewDriftHandler(driftDetector)
	driftHandler.RegisterRoutes(mux, "/api/drift")
	
	// Setup middleware
	configMiddleware := config.NewConfigMiddleware(reloadableConfig)
	_ = config.NewSecretsMiddleware(secretsManager) // Available for use
	driftMiddleware := config.NewDriftMiddleware(driftDetector)
	
	// Add a health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		cfg := reloadableConfig.GetConfig()
		_ = reloadableConfig.GetStats()  // Available for extended health check
		_ = driftDetector.GetStats()     // Available for drift monitoring
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","app":{"name":"%s","version":"%s","environment":"%s"},"timestamp":"%s"}`, 
			cfg.App.Name, cfg.App.Version, cfg.App.Environment, time.Now().Format(time.RFC3339))
	})
	
	// Add demo endpoints to show middleware usage
	
	// Feature-gated endpoint
	demoHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message":"This endpoint is protected by a feature flag","timestamp":"%s"}`, 
			time.Now().Format(time.RFC3339))
	})
	mux.Handle("/demo/feature-gated", configMiddleware.FeatureGate("beta_features")(demoHandler))
	
	// Drift-aware endpoint
	stableHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message":"This endpoint requires stable configuration","timestamp":"%s"}`, 
			time.Now().Format(time.RFC3339))
	})
	mux.Handle("/demo/stable", driftMiddleware.RequireStableConfig()(stableHandler))
	
	// Configuration headers middleware for all routes
	finalHandler := configMiddleware.ConfigHeaders()(mux)
	
	// Wrap with drift detection headers
	finalHandler = driftMiddleware.AlertOnDrift()(finalHandler)
	
	return finalHandler.(*http.ServeMux)
}