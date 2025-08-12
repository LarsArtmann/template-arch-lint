package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server       ServerConfig       `mapstructure:"server" validate:"required"`
	Database     DatabaseConfig     `mapstructure:"database" validate:"required"`
	Logging      LoggingConfig      `mapstructure:"logging" validate:"required"`
	App          AppConfig          `mapstructure:"app" validate:"required"`
	Observability ObservabilityConfig `mapstructure:"observability" validate:"required"`
	Features     FeaturesConfig     `mapstructure:"features"`
	Security     SecurityConfig     `mapstructure:"security"`
	Health       HealthConfig       `mapstructure:"health"`
	Cache        CacheConfig        `mapstructure:"cache"`
	External     ExternalConfig     `mapstructure:"external"`
	Backup       BackupConfig       `mapstructure:"backup"`
	Resources    ResourcesConfig    `mapstructure:"resources"`
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	Host                    string        `mapstructure:"host" validate:"required"`
	Port                    int           `mapstructure:"port" validate:"required,min=1,max=65535"`
	ReadTimeout             time.Duration `mapstructure:"read_timeout"`
	WriteTimeout            time.Duration `mapstructure:"write_timeout"`
	IdleTimeout             time.Duration `mapstructure:"idle_timeout"`
	GracefulShutdownTimeout time.Duration `mapstructure:"graceful_shutdown_timeout"`
}

// DatabaseConfig contains database configuration
type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver" validate:"required,oneof=sqlite3 postgres mysql"`
	DSN             string        `mapstructure:"dsn" validate:"required"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level" validate:"required,valid_log_level"`
	Format string `mapstructure:"format" validate:"required,oneof=json text"`
	Output string `mapstructure:"output" validate:"required"`
}

// AppConfig contains application-specific configuration
type AppConfig struct {
	Name        string `mapstructure:"name" validate:"required"`
	Version     string `mapstructure:"version" validate:"required"`
	Environment string `mapstructure:"environment" validate:"required,valid_environment"`
	Debug       bool   `mapstructure:"debug"`
}

// ObservabilityConfig contains OpenTelemetry configuration
type ObservabilityConfig struct {
	Enabled           bool                     `mapstructure:"enabled"`
	ServiceName       string                   `mapstructure:"service_name" validate:"required"`
	ServiceVersion    string                   `mapstructure:"service_version" validate:"required"`
	Environment       string                   `mapstructure:"environment" validate:"required"`
	Tracing           TracingConfig            `mapstructure:"tracing"`
	Metrics           MetricsConfig            `mapstructure:"metrics"`
	Exporters         ExportersConfig          `mapstructure:"exporters"`
	SamplingRate      float64                  `mapstructure:"sampling_rate" validate:"min=0,max=1"`
}

// TracingConfig contains tracing-specific configuration
type TracingConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Endpoint    string `mapstructure:"endpoint"`
	Headers     map[string]string `mapstructure:"headers"`
	HTTPDetails bool   `mapstructure:"http_details"`
	DBQueries   bool   `mapstructure:"db_queries"`
}

// MetricsConfig contains metrics-specific configuration
type MetricsConfig struct {
	Enabled         bool              `mapstructure:"enabled"`
	Endpoint        string            `mapstructure:"endpoint"`
	Headers         map[string]string `mapstructure:"headers"`
	PushInterval    time.Duration     `mapstructure:"push_interval"`
	BusinessMetrics bool              `mapstructure:"business_metrics"`
}

// ExportersConfig contains configuration for different exporters
type ExportersConfig struct {
	Prometheus PrometheusConfig `mapstructure:"prometheus"`
	OTLP       OTLPConfig       `mapstructure:"otlp"`
	Jaeger     JaegerConfig     `mapstructure:"jaeger"`
}

// PrometheusConfig contains Prometheus exporter configuration
type PrometheusConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port" validate:"min=1,max=65535"`
	Path    string `mapstructure:"path"`
}

// OTLPConfig contains OTLP exporter configuration
type OTLPConfig struct {
	Enabled   bool              `mapstructure:"enabled"`
	Endpoint  string            `mapstructure:"endpoint"`
	Headers   map[string]string `mapstructure:"headers"`
	Insecure  bool              `mapstructure:"insecure"`
}

// JaegerConfig contains Jaeger exporter configuration
type JaegerConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Endpoint string `mapstructure:"endpoint"`
}

// FeaturesConfig contains feature flag configuration
type FeaturesConfig struct {
	EnableDebugEndpoints  bool `mapstructure:"enable_debug_endpoints"`
	EnableProfiling       bool `mapstructure:"enable_profiling"`
	EnableHotReload       bool `mapstructure:"enable_hot_reload"`
	EnableCORS            bool `mapstructure:"enable_cors"`
	EnableRequestLogging  bool `mapstructure:"enable_request_logging"`
	EnableQueryLogging    bool `mapstructure:"enable_query_logging"`
	EnableMetricsDebug    bool `mapstructure:"enable_metrics_debug"`
	EnableBetaFeatures    bool `mapstructure:"enable_beta_features"`
	EnableLoadTesting     bool `mapstructure:"enable_load_testing"`
}

// SecurityConfig contains security-related configuration
type SecurityConfig struct {
	EnableAuth        bool             `mapstructure:"enable_auth"`
	EnableRateLimit   bool             `mapstructure:"enable_rate_limiting"`
	CORS              CORSConfig       `mapstructure:"cors"`
	RateLimit         RateLimitConfig  `mapstructure:"rate_limit"`
	TLS               TLSConfig        `mapstructure:"tls"`
	APIKeys           APIKeysConfig    `mapstructure:"api_keys"`
	RequestSizeLimit  string           `mapstructure:"request_size_limit"`
}

// CORSConfig contains CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins" validate:"required"`
	AllowedMethods   []string `mapstructure:"allowed_methods" validate:"required"`
	AllowedHeaders   []string `mapstructure:"allowed_headers" validate:"required"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
}

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int `mapstructure:"requests_per_minute" validate:"min=1"`
	Burst             int `mapstructure:"burst" validate:"min=1"`
}

// TLSConfig contains TLS configuration
type TLSConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	CertFile   string `mapstructure:"cert_file"`
	KeyFile    string `mapstructure:"key_file"`
	MinVersion string `mapstructure:"min_version" validate:"omitempty,oneof=1.0 1.1 1.2 1.3"`
}

// APIKeysConfig contains API key configuration
type APIKeysConfig struct {
	Enabled        bool          `mapstructure:"enabled"`
	HeaderName     string        `mapstructure:"header_name" validate:"required_if=Enabled true"`
	RotateInterval time.Duration `mapstructure:"rotate_interval"`
}

// HealthConfig contains health check configuration
type HealthConfig struct {
	Enabled                bool          `mapstructure:"enabled"`
	Endpoint               string        `mapstructure:"endpoint" validate:"required_if=Enabled true"`
	Timeout                time.Duration `mapstructure:"timeout"`
	CheckDatabase          bool          `mapstructure:"check_database"`
	CheckExternalServices  bool          `mapstructure:"check_external_services"`
}

// CacheConfig contains cache configuration
type CacheConfig struct {
	Enabled     bool          `mapstructure:"enabled"`
	RedisURL    string        `mapstructure:"redis_url" validate:"required_if=Enabled true"`
	DefaultTTL  time.Duration `mapstructure:"default_ttl"`
	MaxMemory   string        `mapstructure:"max_memory"`
	ClusterMode bool          `mapstructure:"cluster_mode"`
}

// ExternalConfig contains external service configuration
type ExternalConfig struct {
	APITimeout      time.Duration        `mapstructure:"api_timeout"`
	RetryAttempts   int                  `mapstructure:"retry_attempts" validate:"min=0,max=10"`
	CircuitBreaker  CircuitBreakerConfig `mapstructure:"circuit_breaker"`
}

// CircuitBreakerConfig contains circuit breaker configuration
type CircuitBreakerConfig struct {
	Enabled   bool          `mapstructure:"enabled"`
	Threshold int           `mapstructure:"threshold" validate:"min=1"`
	Timeout   time.Duration `mapstructure:"timeout"`
}

// BackupConfig contains backup configuration
type BackupConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	Schedule      string `mapstructure:"schedule" validate:"required_if=Enabled true"`
	RetentionDays int    `mapstructure:"retention_days" validate:"min=1"`
	S3Bucket      string `mapstructure:"s3_bucket" validate:"required_if=Enabled true"`
}

// ResourcesConfig contains resource limits configuration
type ResourcesConfig struct {
	MaxMemory       string        `mapstructure:"max_memory"`
	MaxCPUCores     int           `mapstructure:"max_cpu_cores" validate:"min=1"`
	MaxConnections  int           `mapstructure:"max_connections" validate:"min=1"`
	RequestTimeout  time.Duration `mapstructure:"request_timeout"`
	MaxRequestSize  string        `mapstructure:"max_request_size"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	// Set default values
	setDefaults()

	// Setup configuration paths and sources
	if err := setupConfigPaths(configPath); err != nil {
		return nil, err
	}

	// Setup environment variable bindings
	if err := setupEnvironmentBindings(); err != nil {
		return nil, err
	}

	// Read and parse configuration
	config, err := readAndParseConfig()
	if err != nil {
		return nil, err
	}

	return config, nil
}

// setupConfigPaths configures viper to read from the specified config file or default locations
func setupConfigPaths(configPath string) error {
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		// Look for config in current directory and /etc/
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/template-arch-lint/")
		viper.AddConfigPath("$HOME/.template-arch-lint/")
	}

	// Enable reading from environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP") // Environment variables will be prefixed with APP_
	return nil
}

// setupEnvironmentBindings configures all environment variable bindings
func setupEnvironmentBindings() error {
	if err := bindServerEnvVars(); err != nil {
		return err
	}
	if err := bindDatabaseEnvVars(); err != nil {
		return err
	}
	if err := bindLoggingEnvVars(); err != nil {
		return err
	}
	if err := bindAppEnvVars(); err != nil {
		return err
	}
	if err := bindObservabilityEnvVars(); err != nil {
		return err
	}
	if err := bindFeaturesEnvVars(); err != nil {
		return err
	}
	if err := bindSecurityEnvVars(); err != nil {
		return err
	}
	if err := bindHealthEnvVars(); err != nil {
		return err
	}
	if err := bindCacheEnvVars(); err != nil {
		return err
	}
	if err := bindExternalEnvVars(); err != nil {
		return err
	}
	if err := bindBackupEnvVars(); err != nil {
		return err
	}
	return bindResourcesEnvVars()
}

// bindServerEnvVars binds server-related environment variables
func bindServerEnvVars() error {
	serverBindings := map[string]string{
		"server.host":                        "APP_SERVER_HOST",
		"server.port":                        "APP_SERVER_PORT",
		"server.read_timeout":                "APP_SERVER_READ_TIMEOUT",
		"server.write_timeout":               "APP_SERVER_WRITE_TIMEOUT",
		"server.idle_timeout":                "APP_SERVER_IDLE_TIMEOUT",
		"server.graceful_shutdown_timeout":   "APP_SERVER_GRACEFUL_SHUTDOWN_TIMEOUT",
	}
	return bindEnvVars(serverBindings, "server")
}

// bindDatabaseEnvVars binds database-related environment variables
func bindDatabaseEnvVars() error {
	databaseBindings := map[string]string{
		"database.driver":              "APP_DATABASE_DRIVER",
		"database.dsn":                 "APP_DATABASE_DSN",
		"database.max_open_conns":      "APP_DATABASE_MAX_OPEN_CONNS",
		"database.max_idle_conns":      "APP_DATABASE_MAX_IDLE_CONNS",
		"database.conn_max_lifetime":   "APP_DATABASE_CONN_MAX_LIFETIME",
		"database.conn_max_idle_time":  "APP_DATABASE_CONN_MAX_IDLE_TIME",
	}
	return bindEnvVars(databaseBindings, "database")
}

// bindLoggingEnvVars binds logging-related environment variables
func bindLoggingEnvVars() error {
	loggingBindings := map[string]string{
		"logging.level":  "APP_LOGGING_LEVEL",
		"logging.format": "APP_LOGGING_FORMAT",
		"logging.output": "APP_LOGGING_OUTPUT",
	}
	return bindEnvVars(loggingBindings, "logging")
}

// bindAppEnvVars binds app-related environment variables
func bindAppEnvVars() error {
	appBindings := map[string]string{
		"app.name":        "APP_APP_NAME",
		"app.version":     "APP_APP_VERSION",
		"app.environment": "APP_APP_ENVIRONMENT",
		"app.debug":       "APP_APP_DEBUG",
	}
	return bindEnvVars(appBindings, "app")
}

// bindObservabilityEnvVars binds observability-related environment variables
func bindObservabilityEnvVars() error {
	observabilityBindings := map[string]string{
		"observability.enabled":                      "APP_OBSERVABILITY_ENABLED",
		"observability.service_name":                 "APP_OBSERVABILITY_SERVICE_NAME",
		"observability.service_version":              "APP_OBSERVABILITY_SERVICE_VERSION",
		"observability.environment":                  "APP_OBSERVABILITY_ENVIRONMENT",
		"observability.sampling_rate":                "APP_OBSERVABILITY_SAMPLING_RATE",
		"observability.tracing.enabled":              "APP_OBSERVABILITY_TRACING_ENABLED",
		"observability.tracing.endpoint":             "APP_OBSERVABILITY_TRACING_ENDPOINT",
		"observability.tracing.http_details":         "APP_OBSERVABILITY_TRACING_HTTP_DETAILS",
		"observability.tracing.db_queries":           "APP_OBSERVABILITY_TRACING_DB_QUERIES",
		"observability.metrics.enabled":              "APP_OBSERVABILITY_METRICS_ENABLED",
		"observability.metrics.endpoint":             "APP_OBSERVABILITY_METRICS_ENDPOINT",
		"observability.metrics.push_interval":        "APP_OBSERVABILITY_METRICS_PUSH_INTERVAL",
		"observability.metrics.business_metrics":     "APP_OBSERVABILITY_METRICS_BUSINESS_METRICS",
		"observability.exporters.prometheus.enabled": "APP_OBSERVABILITY_EXPORTERS_PROMETHEUS_ENABLED",
		"observability.exporters.prometheus.port":    "APP_OBSERVABILITY_EXPORTERS_PROMETHEUS_PORT",
		"observability.exporters.prometheus.path":    "APP_OBSERVABILITY_EXPORTERS_PROMETHEUS_PATH",
		"observability.exporters.otlp.enabled":       "APP_OBSERVABILITY_EXPORTERS_OTLP_ENABLED",
		"observability.exporters.otlp.endpoint":      "APP_OBSERVABILITY_EXPORTERS_OTLP_ENDPOINT",
		"observability.exporters.otlp.insecure":      "APP_OBSERVABILITY_EXPORTERS_OTLP_INSECURE",
		"observability.exporters.jaeger.enabled":     "APP_OBSERVABILITY_EXPORTERS_JAEGER_ENABLED",
		"observability.exporters.jaeger.endpoint":    "APP_OBSERVABILITY_EXPORTERS_JAEGER_ENDPOINT",
	}
	return bindEnvVars(observabilityBindings, "observability")
}

// bindEnvVars is a helper function that binds a map of configuration keys to environment variables
func bindEnvVars(bindings map[string]string, section string) error {
	for configKey, envVar := range bindings {
		if err := viper.BindEnv(configKey, envVar); err != nil {
			return fmt.Errorf("failed to bind %s env var: %w", configKey, err)
		}
	}
	return nil
}

// readAndParseConfig reads the config file and parses it into a Config struct
func readAndParseConfig() (*Config, error) {
	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is OK, we'll use defaults and env vars
	}

	// Unmarshal config into struct
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "5s")
	viper.SetDefault("server.write_timeout", "10s")
	viper.SetDefault("server.idle_timeout", "60s")
	viper.SetDefault("server.graceful_shutdown_timeout", "30s")

	// Database defaults
	viper.SetDefault("database.driver", "sqlite3")
	viper.SetDefault("database.dsn", "./app.db")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", "5m")
	viper.SetDefault("database.conn_max_idle_time", "5m")

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")

	// App defaults
	viper.SetDefault("app.name", "template-arch-lint")
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.debug", false)

	// Observability defaults
	viper.SetDefault("observability.enabled", true)
	viper.SetDefault("observability.service_name", "template-arch-lint")
	viper.SetDefault("observability.service_version", "1.0.0")
	viper.SetDefault("observability.environment", "development")
	viper.SetDefault("observability.sampling_rate", 1.0)
	
	// Tracing defaults
	viper.SetDefault("observability.tracing.enabled", true)
	viper.SetDefault("observability.tracing.endpoint", "http://localhost:4318/v1/traces")
	viper.SetDefault("observability.tracing.http_details", true)
	viper.SetDefault("observability.tracing.db_queries", true)
	
	// Metrics defaults
	viper.SetDefault("observability.metrics.enabled", true)
	viper.SetDefault("observability.metrics.endpoint", "http://localhost:4318/v1/metrics")
	viper.SetDefault("observability.metrics.push_interval", "15s")
	viper.SetDefault("observability.metrics.business_metrics", true)
	
	// Prometheus exporter defaults
	viper.SetDefault("observability.exporters.prometheus.enabled", true)
	viper.SetDefault("observability.exporters.prometheus.port", 2112)
	viper.SetDefault("observability.exporters.prometheus.path", "/metrics")
	
	// OTLP exporter defaults
	viper.SetDefault("observability.exporters.otlp.enabled", true)
	viper.SetDefault("observability.exporters.otlp.endpoint", "http://localhost:4318")
	viper.SetDefault("observability.exporters.otlp.insecure", true)
	
	// Jaeger exporter defaults
	viper.SetDefault("observability.exporters.jaeger.enabled", false)
	viper.SetDefault("observability.exporters.jaeger.endpoint", "http://localhost:14268/api/traces")

	// Features defaults
	viper.SetDefault("features.enable_debug_endpoints", false)
	viper.SetDefault("features.enable_profiling", false)
	viper.SetDefault("features.enable_hot_reload", false)
	viper.SetDefault("features.enable_cors", true)
	viper.SetDefault("features.enable_request_logging", false)
	viper.SetDefault("features.enable_query_logging", false)
	viper.SetDefault("features.enable_metrics_debug", false)
	viper.SetDefault("features.enable_beta_features", false)
	viper.SetDefault("features.enable_load_testing", false)

	// Security defaults
	viper.SetDefault("security.enable_auth", false)
	viper.SetDefault("security.enable_rate_limiting", false)
	viper.SetDefault("security.cors.allowed_origins", []string{"*"})
	viper.SetDefault("security.cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("security.cors.allowed_headers", []string{"*"})
	viper.SetDefault("security.cors.allow_credentials", false)
	viper.SetDefault("security.rate_limit.requests_per_minute", 60)
	viper.SetDefault("security.rate_limit.burst", 10)
	viper.SetDefault("security.tls.enabled", false)
	viper.SetDefault("security.tls.min_version", "1.2")
	viper.SetDefault("security.api_keys.enabled", false)
	viper.SetDefault("security.api_keys.header_name", "X-API-Key")
	viper.SetDefault("security.api_keys.rotate_interval", "24h")
	viper.SetDefault("security.request_size_limit", "10MB")

	// Health defaults
	viper.SetDefault("health.enabled", true)
	viper.SetDefault("health.endpoint", "/health")
	viper.SetDefault("health.timeout", "5s")
	viper.SetDefault("health.check_database", true)
	viper.SetDefault("health.check_external_services", false)

	// Cache defaults
	viper.SetDefault("cache.enabled", false)
	viper.SetDefault("cache.default_ttl", "1h")
	viper.SetDefault("cache.max_memory", "256mb")
	viper.SetDefault("cache.cluster_mode", false)

	// External defaults
	viper.SetDefault("external.api_timeout", "30s")
	viper.SetDefault("external.retry_attempts", 3)
	viper.SetDefault("external.circuit_breaker.enabled", true)
	viper.SetDefault("external.circuit_breaker.threshold", 5)
	viper.SetDefault("external.circuit_breaker.timeout", "60s")

	// Backup defaults
	viper.SetDefault("backup.enabled", false)
	viper.SetDefault("backup.retention_days", 30)

	// Resources defaults
	viper.SetDefault("resources.max_memory", "512MB")
	viper.SetDefault("resources.max_cpu_cores", 1)
	viper.SetDefault("resources.max_connections", 1000)
	viper.SetDefault("resources.request_timeout", "30s")
	viper.SetDefault("resources.max_request_size", "10MB")
}

// bindFeaturesEnvVars binds features-related environment variables
func bindFeaturesEnvVars() error {
	featuresBindings := map[string]string{
		"features.enable_debug_endpoints":  "APP_FEATURES_ENABLE_DEBUG_ENDPOINTS",
		"features.enable_profiling":        "APP_FEATURES_ENABLE_PROFILING",
		"features.enable_hot_reload":       "APP_FEATURES_ENABLE_HOT_RELOAD",
		"features.enable_cors":             "APP_FEATURES_ENABLE_CORS",
		"features.enable_request_logging":  "APP_FEATURES_ENABLE_REQUEST_LOGGING",
		"features.enable_query_logging":    "APP_FEATURES_ENABLE_QUERY_LOGGING",
		"features.enable_metrics_debug":    "APP_FEATURES_ENABLE_METRICS_DEBUG",
		"features.enable_beta_features":    "APP_FEATURES_ENABLE_BETA_FEATURES",
		"features.enable_load_testing":     "APP_FEATURES_ENABLE_LOAD_TESTING",
	}
	return bindEnvVars(featuresBindings, "features")
}

// bindSecurityEnvVars binds security-related environment variables
func bindSecurityEnvVars() error {
	securityBindings := map[string]string{
		"security.enable_auth":                      "APP_SECURITY_ENABLE_AUTH",
		"security.enable_rate_limiting":             "APP_SECURITY_ENABLE_RATE_LIMITING",
		"security.cors.allowed_origins":             "APP_SECURITY_CORS_ALLOWED_ORIGINS",
		"security.cors.allowed_methods":             "APP_SECURITY_CORS_ALLOWED_METHODS",
		"security.cors.allowed_headers":             "APP_SECURITY_CORS_ALLOWED_HEADERS",
		"security.cors.allow_credentials":           "APP_SECURITY_CORS_ALLOW_CREDENTIALS",
		"security.rate_limit.requests_per_minute":   "APP_SECURITY_RATE_LIMIT_REQUESTS_PER_MINUTE",
		"security.rate_limit.burst":                 "APP_SECURITY_RATE_LIMIT_BURST",
		"security.tls.enabled":                      "APP_SECURITY_TLS_ENABLED",
		"security.tls.cert_file":                    "APP_SECURITY_TLS_CERT_FILE",
		"security.tls.key_file":                     "APP_SECURITY_TLS_KEY_FILE",
		"security.tls.min_version":                  "APP_SECURITY_TLS_MIN_VERSION",
		"security.api_keys.enabled":                 "APP_SECURITY_API_KEYS_ENABLED",
		"security.api_keys.header_name":             "APP_SECURITY_API_KEYS_HEADER_NAME",
		"security.api_keys.rotate_interval":         "APP_SECURITY_API_KEYS_ROTATE_INTERVAL",
		"security.request_size_limit":               "APP_SECURITY_REQUEST_SIZE_LIMIT",
	}
	return bindEnvVars(securityBindings, "security")
}

// bindHealthEnvVars binds health-related environment variables
func bindHealthEnvVars() error {
	healthBindings := map[string]string{
		"health.enabled":                    "APP_HEALTH_ENABLED",
		"health.endpoint":                   "APP_HEALTH_ENDPOINT",
		"health.timeout":                    "APP_HEALTH_TIMEOUT",
		"health.check_database":             "APP_HEALTH_CHECK_DATABASE",
		"health.check_external_services":    "APP_HEALTH_CHECK_EXTERNAL_SERVICES",
	}
	return bindEnvVars(healthBindings, "health")
}

// bindCacheEnvVars binds cache-related environment variables
func bindCacheEnvVars() error {
	cacheBindings := map[string]string{
		"cache.enabled":       "APP_CACHE_ENABLED",
		"cache.redis_url":     "APP_CACHE_REDIS_URL",
		"cache.default_ttl":   "APP_CACHE_DEFAULT_TTL",
		"cache.max_memory":    "APP_CACHE_MAX_MEMORY",
		"cache.cluster_mode":  "APP_CACHE_CLUSTER_MODE",
	}
	return bindEnvVars(cacheBindings, "cache")
}

// bindExternalEnvVars binds external services-related environment variables
func bindExternalEnvVars() error {
	externalBindings := map[string]string{
		"external.api_timeout":                  "APP_EXTERNAL_API_TIMEOUT",
		"external.retry_attempts":              "APP_EXTERNAL_RETRY_ATTEMPTS",
		"external.circuit_breaker.enabled":     "APP_EXTERNAL_CIRCUIT_BREAKER_ENABLED",
		"external.circuit_breaker.threshold":   "APP_EXTERNAL_CIRCUIT_BREAKER_THRESHOLD",
		"external.circuit_breaker.timeout":     "APP_EXTERNAL_CIRCUIT_BREAKER_TIMEOUT",
	}
	return bindEnvVars(externalBindings, "external")
}

// bindBackupEnvVars binds backup-related environment variables
func bindBackupEnvVars() error {
	backupBindings := map[string]string{
		"backup.enabled":        "APP_BACKUP_ENABLED",
		"backup.schedule":       "APP_BACKUP_SCHEDULE",
		"backup.retention_days": "APP_BACKUP_RETENTION_DAYS",
		"backup.s3_bucket":      "APP_BACKUP_S3_BUCKET",
	}
	return bindEnvVars(backupBindings, "backup")
}

// bindResourcesEnvVars binds resources-related environment variables
func bindResourcesEnvVars() error {
	resourcesBindings := map[string]string{
		"resources.max_memory":       "APP_RESOURCES_MAX_MEMORY",
		"resources.max_cpu_cores":    "APP_RESOURCES_MAX_CPU_CORES",
		"resources.max_connections":  "APP_RESOURCES_MAX_CONNECTIONS",
		"resources.request_timeout":  "APP_RESOURCES_REQUEST_TIMEOUT",
		"resources.max_request_size": "APP_RESOURCES_MAX_REQUEST_SIZE",
	}
	return bindEnvVars(resourcesBindings, "resources")
}

// validateConfig validates the configuration using struct tags
func validateConfig(config *Config) error {
	validate := validator.New()
	
	// Add custom validation for environment-specific rules
	validate.RegisterValidation("valid_environment", validateEnvironment)
	validate.RegisterValidation("valid_log_level", validateLogLevel)
	
	if err := validate.Struct(config); err != nil {
		return fmt.Errorf("validation errors: %w", err)
	}
	
	// Additional business logic validation
	if err := validateBusinessLogic(config); err != nil {
		return fmt.Errorf("business logic validation failed: %w", err)
	}
	
	return nil
}

// validateEnvironment validates environment values
func validateEnvironment(fl validator.FieldLevel) bool {
	env := fl.Field().String()
	validEnvs := []string{"development", "staging", "production", "testing", "local"}
	for _, validEnv := range validEnvs {
		if env == validEnv {
			return true
		}
	}
	return false
}

// validateLogLevel validates log level values
func validateLogLevel(fl validator.FieldLevel) bool {
	level := fl.Field().String()
	validLevels := []string{"debug", "info", "warn", "error"}
	for _, validLevel := range validLevels {
		if level == validLevel {
			return true
		}
	}
	return false
}

// validateCronExpression validates cron expressions (basic validation)
func validateCronExpression(fl validator.FieldLevel) bool {
	cron := fl.Field().String()
	if cron == "" {
		return true // Allow empty for optional fields
	}
	// Basic cron validation - should have 5 parts
	parts := strings.Fields(cron)
	return len(parts) == 5
}

// validateBusinessLogic performs additional business logic validation
func validateBusinessLogic(config *Config) error {
	// Validate that TLS is required in production
	if config.App.Environment == "production" && !config.Security.TLS.Enabled {
		return fmt.Errorf("TLS must be enabled in production environment")
	}
	
	// Validate that authentication is enabled in production
	if config.App.Environment == "production" && !config.Security.EnableAuth {
		return fmt.Errorf("authentication must be enabled in production environment")
	}
	
	// Validate that debug features are disabled in production
	if config.App.Environment == "production" {
		if config.Features.EnableDebugEndpoints {
			return fmt.Errorf("debug endpoints must be disabled in production")
		}
		if config.App.Debug {
			return fmt.Errorf("debug mode must be disabled in production")
		}
	}
	
	// Validate database configuration
	if config.Database.Driver == "postgres" && !strings.Contains(config.Database.DSN, "sslmode") {
		return fmt.Errorf("PostgreSQL connections should specify SSL mode")
	}
	
	// Validate observability in production
	if config.App.Environment == "production" && config.Observability.SamplingRate > 0.2 {
		return fmt.Errorf("sampling rate should be <= 0.2 in production for performance")
	}
	
	return nil
}

// validateSizeFormat validates size format strings like "10MB", "1GB"
func validateSizeFormat(fl validator.FieldLevel) bool {
	size := fl.Field().String()
	if size == "" {
		return true // Allow empty for optional fields
	}
	// Simple validation for size format
	validSuffixes := []string{"B", "KB", "MB", "GB", "TB"}
	for _, suffix := range validSuffixes {
		if strings.HasSuffix(strings.ToUpper(size), suffix) {
			return true
		}
	}
	return false
}

// ConfigDiff represents differences between two configurations
type ConfigDiff struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
	Action   string      `json:"action"` // "added", "removed", "changed"
}

// CompareConfigs compares two configurations and returns differences
func CompareConfigs(oldConfig, newConfig *Config) []ConfigDiff {
	// This is a simplified comparison - in a real implementation
	// you would use reflection or a dedicated library for deep comparison
	var diffs []ConfigDiff

	// Compare server configuration
	if oldConfig.Server.Host != newConfig.Server.Host {
		diffs = append(diffs, ConfigDiff{
			Field:    "server.host",
			OldValue: oldConfig.Server.Host,
			NewValue: newConfig.Server.Host,
			Action:   "changed",
		})
	}

	if oldConfig.Server.Port != newConfig.Server.Port {
		diffs = append(diffs, ConfigDiff{
			Field:    "server.port",
			OldValue: oldConfig.Server.Port,
			NewValue: newConfig.Server.Port,
			Action:   "changed",
		})
	}

	// Compare database configuration
	if oldConfig.Database.Driver != newConfig.Database.Driver {
		diffs = append(diffs, ConfigDiff{
			Field:    "database.driver",
			OldValue: oldConfig.Database.Driver,
			NewValue: newConfig.Database.Driver,
			Action:   "changed",
		})
	}

	// Compare app environment
	if oldConfig.App.Environment != newConfig.App.Environment {
		diffs = append(diffs, ConfigDiff{
			Field:    "app.environment",
			OldValue: oldConfig.App.Environment,
			NewValue: newConfig.App.Environment,
			Action:   "changed",
		})
	}

	return diffs
}