// Dependency injection container using samber/do
package container

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	_ "github.com/mattn/go-sqlite3"

	"github.com/LarsArtmann/template-arch-lint/internal/application/handlers"
	"github.com/LarsArtmann/template-arch-lint/internal/config"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	inmemrepos "github.com/LarsArtmann/template-arch-lint/internal/infrastructure/repositories"
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

	// Register services
	if err := c.registerServices(); err != nil {
		return fmt.Errorf("failed to register services: %w", err)
	}

	// Register handlers
	if err := c.registerHandlers(); err != nil {
		return fmt.Errorf("failed to register handlers: %w", err)
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
		
		repo := persistence.NewSQLUserRepository(db, logger)
		return repo, nil
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
	do.Provide(c.injector, func(i *do.Injector) (*handlers.UserHandler, error) {
		userService := do.MustInvoke[*services.UserService](i)
		logger := do.MustInvoke[*slog.Logger](i)
		
		handler := handlers.NewUserHandler(userService, logger)
		return handler, nil
	})
	return nil
}

// registerHTTPServer registers the HTTP server and router
func (c *Container) registerHTTPServer() error {
	do.Provide(c.injector, func(i *do.Injector) (*gin.Engine, error) {
		cfg := do.MustInvoke[*config.Config](i)
		logger := do.MustInvoke[*slog.Logger](i)
		userHandler := do.MustInvoke[*handlers.UserHandler](i)

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
		router.Use(gin.Logger())
		router.Use(gin.Recovery())
		
		// Add custom middleware for logging
		router.Use(func(c *gin.Context) {
			logger.Info("Request received",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"user_agent", c.GetHeader("User-Agent"),
				"remote_addr", c.ClientIP(),
			)
			c.Next()
		})

		// Health check endpoint
		router.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"version": cfg.App.Version,
				"name":    cfg.App.Name,
			})
		})

		// API routes
		api := router.Group("/api/v1")
		{
			users := api.Group("/users")
			{
				users.POST("", userHandler.CreateUser)
				users.GET("/:id", userHandler.GetUser)
				users.PUT("/:id", userHandler.UpdateUser)
				users.DELETE("/:id", userHandler.DeleteUser)
				users.GET("", userHandler.ListUsers)
			}
		}

		logger.Info("HTTP router configured successfully")
		return router, nil
	})
	return nil
}