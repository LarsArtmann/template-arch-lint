package values

import "slices"

// EnvVar represents a typed environment variable name for type-safe configuration
type EnvVar string

// Server configuration environment variables
const (
	EnvServerPort         EnvVar = "APP_SERVER_PORT"
	EnvServerHost         EnvVar = "APP_SERVER_HOST"
	EnvServerReadTimeout  EnvVar = "APP_SERVER_READ_TIMEOUT"
	EnvServerWriteTimeout EnvVar = "APP_SERVER_WRITE_TIMEOUT"
	EnvServerIdleTimeout  EnvVar = "APP_SERVER_IDLE_TIMEOUT"
)

// Database configuration environment variables
const (
	EnvDatabaseDriver          EnvVar = "APP_DATABASE_DRIVER"
	EnvDatabaseDSN             EnvVar = "APP_DATABASE_DSN"
	EnvDatabaseMaxOpenConns    EnvVar = "APP_DATABASE_MAX_OPEN_CONNS"
	EnvDatabaseMaxIdleConns    EnvVar = "APP_DATABASE_MAX_IDLE_CONNS"
	EnvDatabaseConnMaxLifetime EnvVar = "APP_DATABASE_CONN_MAX_LIFETIME"
)

// Logging configuration environment variables
const (
	EnvLoggingLevel  EnvVar = "APP_LOGGING_LEVEL"
	EnvLoggingFormat EnvVar = "APP_LOGGING_FORMAT"
	EnvLoggingOutput EnvVar = "APP_LOGGING_OUTPUT"
)

// Application configuration environment variables
const (
	EnvApplicationName        EnvVar = "APP_NAME"
	EnvApplicationVersion     EnvVar = "APP_VERSION"
	EnvApplicationEnvironment EnvVar = "APP_ENVIRONMENT"
)

// JWT configuration environment variables
const (
	EnvJWTSecretKey EnvVar = "APP_JWT_SECRET_KEY" // #nosec G101 -- This is just the environment variable name, not a hardcoded secret
	EnvJWTIssuer    EnvVar = "APP_JWT_ISSUER"
	EnvJWTAlgorithm EnvVar = "APP_JWT_ALGORITHM"
)

// String returns the environment variable name as a string
func (e EnvVar) String() string {
	return string(e)
}

// AllEnvVars returns a slice of all defined environment variables
func AllEnvVars() []EnvVar {
	return []EnvVar{
		// Server
		EnvServerPort,
		EnvServerHost,
		EnvServerReadTimeout,
		EnvServerWriteTimeout,
		EnvServerIdleTimeout,

		// Database
		EnvDatabaseDriver,
		EnvDatabaseDSN,
		EnvDatabaseMaxOpenConns,
		EnvDatabaseMaxIdleConns,
		EnvDatabaseConnMaxLifetime,

		// Logging
		EnvLoggingLevel,
		EnvLoggingFormat,
		EnvLoggingOutput,

		// Application
		EnvApplicationName,
		EnvApplicationVersion,
		EnvApplicationEnvironment,

		// JWT
		EnvJWTSecretKey,
		EnvJWTIssuer,
		EnvJWTAlgorithm,
	}
}

// Category returns the configuration category for this environment variable
func (e EnvVar) Category() string {
	switch e {
	case EnvServerPort, EnvServerHost, EnvServerReadTimeout, EnvServerWriteTimeout, EnvServerIdleTimeout:
		return "server"
	case EnvDatabaseDriver, EnvDatabaseDSN, EnvDatabaseMaxOpenConns, EnvDatabaseMaxIdleConns, EnvDatabaseConnMaxLifetime:
		return "database"
	case EnvLoggingLevel, EnvLoggingFormat, EnvLoggingOutput:
		return "logging"
	case EnvApplicationName, EnvApplicationVersion, EnvApplicationEnvironment:
		return "application"
	case EnvJWTSecretKey, EnvJWTIssuer, EnvJWTAlgorithm:
		return "jwt"
	default:
		return "unknown"
	}
}

// IsRequired returns true if this environment variable is required for the application to run
func (e EnvVar) IsRequired() bool {
	requiredVars := []EnvVar{
		EnvApplicationName,
		EnvApplicationVersion,
		EnvApplicationEnvironment,
		EnvJWTSecretKey,
		EnvJWTIssuer,
		EnvJWTAlgorithm,
	}

	return slices.Contains(requiredVars, e)
}

// IsSensitive returns true if this environment variable contains sensitive information
func (e EnvVar) IsSensitive() bool {
	sensitiveVars := []EnvVar{
		EnvJWTSecretKey,
		EnvDatabaseDSN, // May contain passwords
	}

	return slices.Contains(sensitiveVars, e)
}
