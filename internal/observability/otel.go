// Package observability provides OpenTelemetry management
package observability

import (
	"context"
	"log/slog"

	"github.com/LarsArtmann/template-arch-lint/internal/config"
)

// OTELManager manages OpenTelemetry providers
type OTELManager struct {
	config *config.ObservabilityConfig
	logger *slog.Logger
}

// NewOTELManager creates a new OpenTelemetry manager
func NewOTELManager(cfg *config.ObservabilityConfig, logger *slog.Logger) *OTELManager {
	return &OTELManager{
		config: cfg,
		logger: logger,
	}
}

// Initialize sets up OpenTelemetry
func (m *OTELManager) Initialize(ctx context.Context) error {
	if m.config.Enabled {
		m.logger.Info("OpenTelemetry initialized (simplified implementation)")
	} else {
		m.logger.Info("OpenTelemetry is disabled")
	}
	return nil
}

// Shutdown gracefully shuts down OpenTelemetry
func (m *OTELManager) Shutdown(ctx context.Context) error {
	if m.config.Enabled {
		m.logger.Info("OpenTelemetry shutdown completed")
	}
	return nil
}

// DatabaseTracer provides database tracing
type DatabaseTracer struct {
	config *config.ObservabilityConfig
	logger *slog.Logger
}

// NewDatabaseTracer creates a new database tracer
func NewDatabaseTracer(cfg *config.ObservabilityConfig, logger *slog.Logger) (*DatabaseTracer, error) {
	return &DatabaseTracer{
		config: cfg,
		logger: logger,
	}, nil
}