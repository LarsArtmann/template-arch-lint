// Package container provides dependency injection for the application.
package container

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http/pprof"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3" // Import sqlite3 driver for database/sql
	"github.com/samber/do"

	"github.com/LarsArtmann/template-arch-lint/internal/application/handlers"
	"github.com/LarsArtmann/template-arch-lint/internal/application/middleware"
	"github.com/LarsArtmann/template-arch-lint/internal/config"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	"github.com/LarsArtmann/template-arch-lint/internal/infrastructure/persistence"
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
	c.registerConfig()

	// Register logger
	c.registerLogger()

	// Register database
	c.registerDatabase()

	// Register repositories
	c.registerRepositories()

	// Register services
	c.registerServices()

	// Register handlers
	c.registerHandlers()

	// Register HTTP server
	c.registerHTTPServer()

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
func (c *Container) registerConfig() {
	do.Provide(c.injector, func(_ *do.Injector) (*config.Config, error) {
		configPath := os.Getenv("CONFIG_PATH")
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
		return cfg, nil
	})
}

// registerLogger registers the logger
func (c *Container) registerLogger() {
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
}

// registerDatabase registers the database connection
func (c *Container) registerDatabase() {
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
}

// registerRepositories registers all repositories
func (c *Container) registerRepositories() {
	do.Provide(c.injector, func(i *do.Injector) (repositories.UserRepository, error) {
		db := do.MustInvoke[*sql.DB](i)
		logger := do.MustInvoke[*slog.Logger](i)

		// Use SQL-based repository implementation
		repo := persistence.NewSQLUserRepository(db, logger)
		return repo, nil
	})
}

// registerServices registers all domain services
func (c *Container) registerServices() {
	do.Provide(c.injector, func(i *do.Injector) (*services.UserService, error) {
		userRepo := do.MustInvoke[repositories.UserRepository](i)
		service := services.NewUserService(userRepo)
		return service, nil
	})
}

// registerHandlers registers all HTTP handlers
func (c *Container) registerHandlers() {
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

	// Register health handler
	do.Provide(c.injector, func(i *do.Injector) (*handlers.HealthHandler, error) {
		db := do.MustInvoke[*sql.DB](i)
		logger := do.MustInvoke[*slog.Logger](i)

		handler := handlers.NewHealthHandler(db, logger)
		return handler, nil
	})

	// Register performance handler
	do.Provide(c.injector, func(i *do.Injector) (*handlers.PerformanceHandler, error) {
		logger := do.MustInvoke[*slog.Logger](i)

		handler := handlers.NewPerformanceHandler(logger)
		return handler, nil
	})
}

// registerHTTPServer registers the HTTP server and router
func (c *Container) registerHTTPServer() {
	do.Provide(c.injector, func(i *do.Injector) (*gin.Engine, error) {
		cfg := do.MustInvoke[*config.Config](i)
		logger := do.MustInvoke[*slog.Logger](i)
		userHandler := do.MustInvoke[*handlers.UserHandler](i)
		templHandler := do.MustInvoke[*handlers.TemplateHandler](i)
		healthHandler := do.MustInvoke[*handlers.HealthHandler](i)
		perfHandler := do.MustInvoke[*handlers.PerformanceHandler](i)

		router := c.createGinRouter(cfg, logger)
		c.setupCoreRoutes(router, cfg, healthHandler, perfHandler)
		c.setupTemplateRoutes(router, templHandler)
		c.setupAPIRoutes(router, userHandler)

		logger.Info("HTTP router configured successfully with template, API, health, and performance routes")
		return router, nil
	})
}

// createGinRouter creates and configures gin router with middleware
func (c *Container) createGinRouter(cfg *config.Config, logger *slog.Logger) *gin.Engine {
	c.setGinMode(cfg.App.Environment)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.WithCorrelationID(logger))
	router.Use(middleware.WithStructuredLogging(logger))

	return router
}

// setGinMode sets gin mode based on environment
func (c *Container) setGinMode(environment string) {
	switch environment {
	case "production":
		gin.SetMode(gin.ReleaseMode)
	case "development":
		gin.SetMode(gin.DebugMode)
	default:
		gin.SetMode(gin.TestMode)
	}
}

// setupCoreRoutes sets up core application routes
func (c *Container) setupCoreRoutes(router *gin.Engine, cfg *config.Config, healthHandler *handlers.HealthHandler, perfHandler *handlers.PerformanceHandler) {
	// Register comprehensive health check endpoints
	handlers.RegisterHealthRoutes(router, healthHandler)

	// Setup performance profiling endpoints (development/debug only)
	c.setupProfilingRoutes(router, cfg)

	// Setup performance monitoring endpoints
	c.setupPerformanceRoutes(router, cfg, perfHandler)

	// Root redirect to users page
	router.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/users")
	})
}

// setupProfilingRoutes configures pprof endpoints for performance analysis
func (c *Container) setupProfilingRoutes(router *gin.Engine, cfg *config.Config) {
	// Only enable profiling in development or debug environments
	if cfg.App.Environment == "development" || cfg.App.Environment == "debug" {
		// Create a route group for debug endpoints
		debug := router.Group("/debug")
		{
			// Standard pprof endpoints
			debug.GET("/pprof/", gin.WrapF(pprof.Index))
			debug.GET("/pprof/allocs", gin.WrapF(pprof.Handler("allocs").ServeHTTP))
			debug.GET("/pprof/block", gin.WrapF(pprof.Handler("block").ServeHTTP))
			debug.GET("/pprof/cmdline", gin.WrapF(pprof.Cmdline))
			debug.GET("/pprof/goroutine", gin.WrapF(pprof.Handler("goroutine").ServeHTTP))
			debug.GET("/pprof/heap", gin.WrapF(pprof.Handler("heap").ServeHTTP))
			debug.GET("/pprof/mutex", gin.WrapF(pprof.Handler("mutex").ServeHTTP))
			debug.GET("/pprof/profile", gin.WrapF(pprof.Profile))
			debug.GET("/pprof/symbol", gin.WrapF(pprof.Symbol))
			debug.GET("/pprof/trace", gin.WrapF(pprof.Trace))
			debug.GET("/pprof/threadcreate", gin.WrapF(pprof.Handler("threadcreate").ServeHTTP))

			// Custom performance metrics endpoint
			debug.GET("/metrics", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message":         "Performance metrics endpoint",
					"pprof_available": true,
					"endpoints": gin.H{
						"cpu_profile":  "/debug/pprof/profile?seconds=30",
						"heap_profile": "/debug/pprof/heap",
						"goroutines":   "/debug/pprof/goroutine",
						"allocs":       "/debug/pprof/allocs",
						"trace":        "/debug/pprof/trace?seconds=10",
					},
					"usage": gin.H{
						"cpu_profiling":  "curl http://localhost:8080/debug/pprof/profile?seconds=30 -o cpu.prof",
						"heap_analysis":  "curl http://localhost:8080/debug/pprof/heap -o heap.prof",
						"goroutine_dump": "curl http://localhost:8080/debug/pprof/goroutine -o goroutine.prof",
						"view_profile":   "go tool pprof cpu.prof",
						"trace_analysis": "curl http://localhost:8080/debug/pprof/trace?seconds=10 -o trace.out && go tool trace trace.out",
					},
				})
			})
		}
	}
}

// setupPerformanceRoutes configures performance monitoring endpoints
func (c *Container) setupPerformanceRoutes(router *gin.Engine, cfg *config.Config, perfHandler *handlers.PerformanceHandler) {
	// Performance monitoring is available in all environments but with different access levels
	perf := router.Group("/performance")
	{
		// Basic runtime stats (always available)
		perf.GET("/stats", perfHandler.RuntimeStats)
		perf.GET("/health", perfHandler.HealthMetrics)
		perf.GET("/info", perfHandler.DebugInfo)

		// Development/debug only endpoints
		if cfg.App.Environment == "development" || cfg.App.Environment == "debug" {
			perf.POST("/gc", perfHandler.ForceGC)
			perf.GET("/memory", perfHandler.MemoryDump)
		}
	}
}

// setupTemplateRoutes configures template-based routes for HTMX
func (c *Container) setupTemplateRoutes(router *gin.Engine, templHandler *handlers.TemplateHandler) {
	// Main pages
	router.GET("/users", templHandler.UsersPage)
	router.GET("/users/new", templHandler.CreateUserPage)
	router.GET("/users/:id/edit", templHandler.EditUserPage)

	// HTMX endpoints
	router.GET("/users/search", templHandler.SearchUsers)
	router.GET("/users/:id/edit-inline", templHandler.EditUserInline)
	router.GET("/users/:id/cancel-edit", templHandler.CancelUserEdit)

	// HTMX form submissions
	router.POST("/users", templHandler.CreateUser)
	router.PUT("/users/:id", templHandler.UpdateUser)
	router.DELETE("/users/:id", templHandler.DeleteUser)
}

// setupAPIRoutes configures JSON API routes
func (c *Container) setupAPIRoutes(router *gin.Engine, userHandler *handlers.UserHandler) {
	api := router.Group("/api/v1")
	users := api.Group("/users")

	// Standard CRUD operations
	users.POST("", userHandler.CreateUser)
	users.GET("/:id", userHandler.GetUser)
	users.PUT("/:id", userHandler.UpdateUser)
	users.DELETE("/:id", userHandler.DeleteUser)
	users.GET("", userHandler.ListUsers)

	// Functional programming endpoints
	users.GET("/active", userHandler.GetActiveUsers)
	users.POST("/functional", userHandler.CreateUserFunctional)
}
