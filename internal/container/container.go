// Package container provides dependency injection for the application.
package container

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
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
}

// registerHTTPServer registers the HTTP server and router
func (c *Container) registerHTTPServer() {
	do.Provide(c.injector, func(i *do.Injector) (*gin.Engine, error) {
		cfg := do.MustInvoke[*config.Config](i)
		logger := do.MustInvoke[*slog.Logger](i)
		userHandler := do.MustInvoke[*handlers.UserHandler](i)
		templHandler := do.MustInvoke[*handlers.TemplateHandler](i)

		router := c.createGinRouter(cfg, logger)
		c.setupCoreRoutes(router, cfg)
		c.setupTemplateRoutes(router, templHandler)
		c.setupAPIRoutes(router, userHandler)

		logger.Info("HTTP router configured successfully with template and API routes")
		return router, nil
	})
}

// createGinRouter creates and configures gin router with middleware
func (c *Container) createGinRouter(cfg *config.Config, logger *slog.Logger) *gin.Engine {
	c.setGinMode(cfg.App.Environment)
	
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLoggingMiddleware(logger))
	
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
func (c *Container) setupCoreRoutes(router *gin.Engine, cfg *config.Config) {
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