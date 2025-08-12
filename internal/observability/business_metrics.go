// Package observability provides basic business metrics tracking
package observability

import (
	"context"
	"log/slog"

	"github.com/LarsArtmann/template-arch-lint/internal/config"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
)

// BusinessMetrics provides basic metrics collection
type BusinessMetrics struct {
	config *config.ObservabilityConfig
	logger *slog.Logger
}

// NewBusinessMetrics creates a new business metrics collector
func NewBusinessMetrics(cfg *config.ObservabilityConfig, logger *slog.Logger) (*BusinessMetrics, error) {
	return &BusinessMetrics{
		config: cfg,
		logger: logger,
	}, nil
}

// TraceBusinessOperation provides basic operation tracing
func (bm *BusinessMetrics) TraceBusinessOperation(ctx context.Context, operation string) (context.Context, func(error)) {
	// Simplified implementation for template
	return ctx, func(error) {}
}

// RecordUserCreated records user creation metrics
func (bm *BusinessMetrics) RecordUserCreated(ctx context.Context, user *entities.User) {
	if bm.logger != nil && bm.config != nil && bm.config.Enabled {
		bm.logger.Info("User created", "user_id", user.ID, "email", user.Email)
	}
}

// RecordUserValidation records user validation metrics
func (bm *BusinessMetrics) RecordUserValidation(ctx context.Context, valid bool, field string) {
	if bm.logger != nil && bm.config != nil && bm.config.Enabled {
		bm.logger.Debug("User validation", "valid", valid, "field", field)
	}
}

// GetUserMetricsSummary returns a summary of user metrics
func (bm *BusinessMetrics) GetUserMetricsSummary(ctx context.Context) map[string]interface{} {
	return map[string]interface{}{
		"status": "observability_enabled",
		"metrics_collected": bm.config != nil && bm.config.Enabled,
	}
}

// RecordLoOperation records a functional programming operation
func RecordLoOperation(ctx context.Context, operation string, count int) {
	// Simplified implementation for template
	if logger := slog.Default(); logger != nil {
		logger.Debug("Lo operation", "operation", operation, "count", count)
	}
}