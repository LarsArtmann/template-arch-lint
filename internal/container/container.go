// Simple dependency injection container for HTMX template implementation
package container

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/samber/do"

	"github.com/LarsArtmann/template-arch-lint/internal/application/handlers"
	"github.com/LarsArtmann/template-arch-lint/internal/application/middleware"
	"github.com/LarsArtmann/template-arch-lint/internal/config"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	"github.com/LarsArtmann/template-arch-lint/internal/infrastructure/persistence"
	"github.com/LarsArtmann/template-arch-lint/internal/observability"
)

// Container represents the DI container
type Container struct {
	injector *do.Injector
}

// New creates a new DI container
func New() *Container {
	injector := do.New()
	return &Container{
		injector: injector,
	}
}

// RegisterAll registers all dependencies in the container
func (c *Container) RegisterAll() error {
	// Register configuration
	if err := c.registerConfig(); err != nil {
		return fmt.Errorf("failed to register config: %w", err)
	}

	// Register logger
	if err := c.registerLogger(); err != nil {
		return fmt.Errorf("failed to register logger: %w", err)
	}

	// Register database
	if err := c.registerDatabase(); err != nil {
		return fmt.Errorf("failed to register database: %w", err)
	}

	// Register repositories
	if err := c.registerRepositories(); err != nil {
		return fmt.Errorf("failed to register repositories: %w", err)
	}

	// Register database performance monitor
	if err := c.registerDBPerformanceMonitor(); err != nil {
		return fmt.Errorf("failed to register DB performance monitor: %w", err)
	}

	// Register services
	if err := c.registerServices(); err != nil {
		return fmt.Errorf("failed to register services: %w", err)
	}

	// Register handlers
	if err := c.registerHandlers(); err != nil {
		return fmt.Errorf("failed to register handlers: %w", err)
	}

	// Register performance metrics
	if err := c.registerPerformanceMetrics(); err != nil {
		return fmt.Errorf("failed to register performance metrics: %w", err)
	}

	// Register resource manager
	if err := c.registerResourceManager(); err != nil {
		return fmt.Errorf("failed to register resource manager: %w", err)
	}

	// Register cache manager
	if err := c.registerCacheManager(); err != nil {
		return fmt.Errorf("failed to register cache manager: %w", err)
	}

	// Register Prometheus metrics
	if err := c.registerPrometheusMetrics(); err != nil {
		return fmt.Errorf("failed to register Prometheus metrics: %w", err)
	}

	// Register SLA tracker
	if err := c.registerSLATracker(); err != nil {
		return fmt.Errorf("failed to register SLA tracker: %w", err)
	}

	// Register HTTP server
	if err := c.registerHTTPServer(); err != nil {
		return fmt.Errorf("failed to register HTTP server: %w", err)
	}

	return nil
}

// GetInjector returns the underlying do.Injector
func (c *Container) GetInjector() *do.Injector {
	return c.injector
}

// Shutdown gracefully shuts down the container
func (c *Container) Shutdown() error {
	return c.injector.Shutdown()
}

// registerConfig registers the configuration
func (c *Container) registerConfig() error {
	do.Provide(c.injector, func(i *do.Injector) (*config.Config, error) {
		configPath := os.Getenv("CONFIG_PATH")
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
		return cfg, nil
	})
	return nil
}

// registerLogger registers the logger
func (c *Container) registerLogger() error {
	do.Provide(c.injector, func(i *do.Injector) (*slog.Logger, error) {
		cfg := do.MustInvoke[*config.Config](i)

		var level slog.Level
		switch cfg.Logging.Level {
		case "debug":
			level = slog.LevelDebug
		case "info":
			level = slog.LevelInfo
		case "warn":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		default:
			level = slog.LevelInfo
		}

		var handler slog.Handler
		opts := &slog.HandlerOptions{Level: level}

		if cfg.Logging.Format == "json" {
			handler = slog.NewJSONHandler(os.Stdout, opts)
		} else {
			handler = slog.NewTextHandler(os.Stdout, opts)
		}

		logger := slog.New(handler)
		slog.SetDefault(logger)

		return logger, nil
	})
	return nil
}

// registerDatabase registers the database connection
func (c *Container) registerDatabase() error {
	do.Provide(c.injector, func(i *do.Injector) (*sql.DB, error) {
		cfg := do.MustInvoke[*config.Config](i)
		logger := do.MustInvoke[*slog.Logger](i)

		db, err := sql.Open(cfg.Database.Driver, cfg.Database.DSN)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}

		// Configure connection pool
		db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
		db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
		db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
		db.SetConnMaxIdleTime(cfg.Database.ConnMaxIdleTime)

		// Test connection
		ctx := context.Background()
		if err := db.PingContext(ctx); err != nil {
			return nil, fmt.Errorf("failed to ping database: %w", err)
		}

		logger.Info("Database connected successfully",
			"driver", cfg.Database.Driver,
			"max_open_conns", cfg.Database.MaxOpenConns,
			"max_idle_conns", cfg.Database.MaxIdleConns,
		)

		return db, nil
	})
	return nil
}

// registerRepositories registers all repositories
func (c *Container) registerRepositories() error {
	do.Provide(c.injector, func(i *do.Injector) (repositories.UserRepository, error) {
		db := do.MustInvoke[*sql.DB](i)
		logger := do.MustInvoke[*slog.Logger](i)

		// Use SQL-based repository implementation
		repo := persistence.NewSQLUserRepository(db, logger)
		return repo, nil
	})
	return nil
}

// registerResourceManager registers the resource manager
func (c *Container) registerResourceManager() error {
	do.Provide(c.injector, func(i *do.Injector) (*observability.ResourceManager, error) {
		logger := do.MustInvoke[*slog.Logger](i)
		performanceMetrics := do.MustInvoke[*observability.PerformanceMetrics](i)
		cfg := do.MustInvoke[*config.Config](i)
		
		// Create resource config based on application configuration
		resourceConfig := &observability.ResourceConfig{
			MaxMemoryBytes:       parseMemoryString(cfg.Resources.MaxMemory),
			GCTarget:            100.0,
			GCInterval:          5 * time.Minute,
			MaxGoroutines:        cfg.Resources.MaxConnections,
			GoroutineThreshold:   cfg.Resources.MaxConnections / 2,
			MonitoringInterval:   30 * time.Second,
			OptimizationEnabled:  true,
			OptimizationCooldown: 5 * time.Minute,
		}
		
		resourceManager := observability.NewResourceManager(logger, performanceMetrics, resourceConfig)
		
		// Start monitoring in background
		go resourceManager.StartMonitoring(context.Background())
		
		// Optimize for environment
		if err := resourceManager.OptimizeForEnvironment(cfg.App.Environment); err != nil {
			logger.Warn("Failed to optimize for environment", "environment", cfg.App.Environment, "error", err)
		}
		
		return resourceManager, nil
	})
	return nil
}

// parseMemoryString parses memory strings like "512MB" into bytes
func parseMemoryString(memStr string) int64 {
	// Simple parsing for common formats
	if memStr == "" {
		return 512 * 1024 * 1024 // Default 512MB
	}
	
	// This is a simplified parser - in production you'd want more robust parsing
	switch {
	case len(memStr) >= 2 && memStr[len(memStr)-2:] == "MB":
		if val := parseNumber(memStr[:len(memStr)-2]); val > 0 {
			return val * 1024 * 1024
		}
	case len(memStr) >= 2 && memStr[len(memStr)-2:] == "GB":
		if val := parseNumber(memStr[:len(memStr)-2]); val > 0 {
			return val * 1024 * 1024 * 1024
		}
	}
	
	return 512 * 1024 * 1024 // Default fallback
}

// parseNumber parses a number string
func parseNumber(str string) int64 {
	// Simple number parsing
	var result int64
	for _, char := range str {
		if char >= '0' && char <= '9' {
			result = result*10 + int64(char-'0')
		}
	}
	return result
}

// registerCacheManager registers the cache manager
func (c *Container) registerCacheManager() error {
	do.Provide(c.injector, func(i *do.Injector) (*observability.CacheManager, error) {
		logger := do.MustInvoke[*slog.Logger](i)
		performanceMetrics := do.MustInvoke[*observability.PerformanceMetrics](i)
		cfg := do.MustInvoke[*config.Config](i)
		
		// Create cache config based on application configuration
		cacheConfig := &observability.CacheConfig{
			L1Enabled:           true,
			L1MaxSize:           parseMemoryString(cfg.Cache.MaxMemory) / 2, // Half for L1 cache
			L1TTL:               cfg.Cache.DefaultTTL,
			L1EvictionPolicy:    "LRU",
			L2Enabled:           cfg.Cache.Enabled,
			L2RedisURL:          cfg.Cache.RedisURL,
			L2TTL:               cfg.Cache.DefaultTTL * 2,
			L2ClusterMode:       cfg.Cache.ClusterMode,
			MonitoringInterval:  30 * time.Second,
			HitRatioThreshold:   0.8,
			EvictionThreshold:   0.9,
			WarmupEnabled:       true,
			HotPathDetection:    true,
			HotPathThreshold:    100,
			HotPathCacheTTL:     1 * time.Hour,
		}
		
		cacheManager := observability.NewCacheManager(logger, performanceMetrics, cacheConfig)
		
		// Start monitoring in background
		go cacheManager.StartMonitoring(context.Background())
		
		return cacheManager, nil
	})
	return nil
}

// registerDBPerformanceMonitor registers the database performance monitor
func (c *Container) registerDBPerformanceMonitor() error {
	do.Provide(c.injector, func(i *do.Injector) (*persistence.DBPerformanceMonitor, error) {
		db := do.MustInvoke[*sql.DB](i)
		logger := do.MustInvoke[*slog.Logger](i)
		performanceMetrics := do.MustInvoke[*observability.PerformanceMetrics](i)

		monitor := persistence.NewDBPerformanceMonitor(db, logger, performanceMetrics)
		
		// Set optimal database pragmas on startup
		ctx := context.Background()
		if err := monitor.SetOptimalPragmas(ctx); err != nil {
			logger.Error("Failed to set optimal database pragmas", "error", err)
		}

		return monitor, nil
	})
	return nil
}

// registerServices registers all domain services
func (c *Container) registerServices() error {
	do.Provide(c.injector, func(i *do.Injector) (*services.UserService, error) {
		userRepo := do.MustInvoke[repositories.UserRepository](i)
		service := services.NewUserService(userRepo)
		return service, nil
	})
	return nil
}

// registerHandlers registers all HTTP handlers
func (c *Container) registerHandlers() error {
	// Register API handler
	do.Provide(c.injector, func(i *do.Injector) (*handlers.UserHandler, error) {
		userService := do.MustInvoke[*services.UserService](i)
		logger := do.MustInvoke[*slog.Logger](i)

		handler := handlers.NewUserHandler(userService, logger)
		return handler, nil
	})

	// Register template handler
	do.Provide(c.injector, func(i *do.Injector) (*handlers.TemplateHandler, error) {
		userService := do.MustInvoke[*services.UserService](i)
		logger := do.MustInvoke[*slog.Logger](i)

		handler := handlers.NewTemplateHandler(userService, logger)
		return handler, nil
	})

	// Register monitoring handler
	do.Provide(c.injector, func(i *do.Injector) (*handlers.MonitoringHandler, error) {
		cfg := do.MustInvoke[*config.Config](i)
		logger := do.MustInvoke[*slog.Logger](i)
		prometheusMetrics := do.MustInvoke[*observability.PrometheusMetrics](i)
		slaTracker := do.MustInvoke[*observability.SLATracker](i)

		handler := handlers.NewMonitoringHandler(cfg, logger, prometheusMetrics, slaTracker)
		return handler, nil
	})

	return nil
}

// registerPerformanceMetrics registers the performance metrics collector
func (c *Container) registerPerformanceMetrics() error {
	do.Provide(c.injector, func(i *do.Injector) (*observability.PerformanceMetrics, error) {
		logger := do.MustInvoke[*slog.Logger](i)
		
		performanceMetrics, err := observability.NewPerformanceMetrics(logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create performance metrics: %w", err)
		}

		// Start runtime metrics collection
		go performanceMetrics.StartRuntimeMetricsCollection(context.Background())
		
		return performanceMetrics, nil
	})
	return nil
}

// registerPrometheusMetrics registers the Prometheus metrics collector
func (c *Container) registerPrometheusMetrics() error {
	do.Provide(c.injector, func(i *do.Injector) (*observability.PrometheusMetrics, error) {
		cfg := do.MustInvoke[*config.Config](i)
		logger := do.MustInvoke[*slog.Logger](i)
		
		prometheusMetrics := observability.NewPrometheusMetrics(cfg, logger)
		
		// Start the Prometheus metrics server
		ctx := context.Background()
		if err := prometheusMetrics.Start(ctx); err != nil {
			return nil, fmt.Errorf("failed to start Prometheus metrics server: %w", err)
		}
		
		return prometheusMetrics, nil
	})
	return nil
}

// registerSLATracker registers the SLA tracking service
func (c *Container) registerSLATracker() error {
	do.Provide(c.injector, func(i *do.Injector) (*observability.SLATracker, error) {
		cfg := do.MustInvoke[*config.Config](i)
		logger := do.MustInvoke[*slog.Logger](i)
		prometheusMetrics := do.MustInvoke[*observability.PrometheusMetrics](i)
		
		slaTracker := observability.NewSLATracker(cfg, logger, prometheusMetrics)
		
		// Start the SLA tracker
		ctx := context.Background()
		if err := slaTracker.Start(ctx); err != nil {
			return nil, fmt.Errorf("failed to start SLA tracker: %w", err)
		}
		
		return slaTracker, nil
	})
	return nil
}

// registerHTTPServer registers the HTTP server and router
func (c *Container) registerHTTPServer() error {
	do.Provide(c.injector, func(i *do.Injector) (*gin.Engine, error) {
		cfg := do.MustInvoke[*config.Config](i)
		logger := do.MustInvoke[*slog.Logger](i)
		userHandler := do.MustInvoke[*handlers.UserHandler](i)
		templHandler := do.MustInvoke[*handlers.TemplateHandler](i)
		monitoringHandler := do.MustInvoke[*handlers.MonitoringHandler](i)
		prometheusMetrics := do.MustInvoke[*observability.PrometheusMetrics](i)
		slaTracker := do.MustInvoke[*observability.SLATracker](i)

		// Set Gin mode based on environment
		if cfg.App.Environment == "production" {
			gin.SetMode(gin.ReleaseMode)
		} else if cfg.App.Environment == "development" {
			gin.SetMode(gin.DebugMode)
		} else {
			gin.SetMode(gin.TestMode)
		}

		// Create router
		router := gin.New()

		// Add middleware
		router.Use(gin.Recovery())
		router.Use(middleware.RequestLoggingMiddleware(logger))
		router.Use(prometheusMetrics.HTTPMiddleware())
		router.Use(observability.SLAMiddleware(slaTracker, logger))

		// Setup profiling endpoints if enabled
		middleware.SetupProfilingRoutes(
			router,
			cfg.Features.EnableProfiling,
			os.Getenv("PPROF_USERNAME"),
			os.Getenv("PPROF_PASSWORD"),
			[]string{"127.0.0.1", "::1"}, // Only allow localhost by default
		)

		// Health check endpoint
		router.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"version": cfg.App.Version,
				"name":    cfg.App.Name,
			})
		})

		// Root redirect to users page
		router.GET("/", func(c *gin.Context) {
			c.Redirect(302, "/users")
		})

		// Template-based web routes (HTMX)
		{
			// Main pages
			router.GET("/users", templHandler.UsersPage)
			router.GET("/users/new", templHandler.CreateUserPage)
			router.GET("/users/:id/edit", templHandler.EditUserPage)

			// HTMX endpoints
			router.GET("/users/search", templHandler.SearchUsers)
			router.GET("/users/:id/edit-inline", templHandler.EditUserInline)
			router.GET("/users/:id/cancel-edit", templHandler.CancelUserEdit)
			router.GET("/users/stats", templHandler.UserStatsPartial)

			// HTMX form submissions
			router.POST("/users", templHandler.CreateUser)
			router.PUT("/users/:id", templHandler.UpdateUser)
			router.DELETE("/users/:id", templHandler.DeleteUser)
		}

		// API routes (JSON)
		api := router.Group("/api/v1")
		{
			users := api.Group("/users")
			{
				users.POST("", userHandler.CreateUser)
				users.GET("/:id", userHandler.GetUser)
				users.PUT("/:id", userHandler.UpdateUser)
				users.DELETE("/:id", userHandler.DeleteUser)
				users.GET("", userHandler.ListUsers)

				// Functional programming endpoints
				users.GET("/stats", userHandler.GetUserStats)
				users.GET("/active", userHandler.GetActiveUsers)
				users.GET("/emails", userHandler.GetUserEmails)
				users.GET("/filtered", userHandler.GetUsersFiltered)
				users.GET("/by-domains", userHandler.GetUsersByDomains)
				users.POST("/functional", userHandler.CreateUserFunctional)
				users.POST("/validate-batch", userHandler.ValidateUsersBatch)
			}

			// Monitoring and SLA endpoints
			monitoring := api.Group("/monitoring")
			{
				monitoring.GET("/sla", monitoringHandler.GetSLAStatus)
				monitoring.GET("/sla/:tier", monitoringHandler.GetSLAStatusForTier)
				monitoring.GET("/health", monitoringHandler.GetHealthDetails)
				
				// Test endpoints for alert generation
				testing := monitoring.Group("/test")
				{
					testing.POST("/slow", monitoringHandler.SimulateSlowResponse)
					testing.POST("/error", monitoringHandler.SimulateError)
					testing.POST("/alert/:type", monitoringHandler.TriggerAlert)
				}
			}
		}

		logger.Info("HTTP router configured successfully with template and API routes")
		return router, nil
	})
	return nil
}