package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	// development , staging , production
	Environment string

	ServiceName string
	Version     string

	// Server configuration
	Server ServerConfig
	//Database
	Database DatabaseConfig

	Auth AuthConfig

	RateLimit RateLimitConfig

	Redis RedisConfig

	Loggin LoggingConfig

	Metrics MetricsConfig

	Tracing TracingConfig

	Features FeatureConfig

	Logging LoggingConfig

	Webhook WebhookConfig
}

// WebhookConfig holds webhook delivery configuration.
type WebhookConfig struct {
	// Delivery timeout
	DeliveryTimeout time.Duration

	// Retry settings
	MaxRetries    int
	RetryInterval time.Duration

	// Signature settings
	SignatureHeader    string
	SignatureAlgorithm string

	// Failure threshold before disabling
	FailureThreshold int
}

// =============================================================================
// Redis Configuration
// =============================================================================

// RedisConfig holds Redis configuration.
type RedisConfig struct {
	// Connection string (takes precedence if set)
	URL string

	// Individual connection parameters
	Host     string
	Port     int
	Password string
	DB       int

	// Connection pool settings
	PoolSize     int
	MinIdleConns int

	// Timeouts
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	// TLS
	TLSEnabled bool

	// Cluster mode
	ClusterEnabled bool
	ClusterNodes   []string
}

// Addr returns the Redis address in host:port format.
func (c RedisConfig) Addr() string {
	if c.URL != "" {
		return c.URL
	}
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
	// Enable rate limiting
	Enabled bool

	// Default limits (requests per second)
	DefaultRPS   int
	DefaultBurst int

	// Plan-specific limits
	Plans map[string]PlanRateLimit

	// Redis key prefix
	KeyPrefix string

	// Window duration for sliding window
	WindowDuration time.Duration
}

// PlanRateLimit holds rate limits for a specific plan.
type PlanRateLimit struct {
	RPS   int
	Burst int
}

type ServerConfig struct {
	// Host to bind to
	Host string

	// Port to listen on
	Port int

	// Read timeout for requests
	ReadTimeout time.Duration

	// Write timeout for responses
	WriteTimeout time.Duration

	// Idle timeout for keep-alive connections
	IdleTimeout time.Duration

	// Shutdown timeout for graceful shutdown
	ShutdownTimeout time.Duration

	// Maximum request body size
	MaxBodySize int64

	// Enable CORS
	CORSEnabled bool

	// Allowed CORS origins
	CORSOrigins []string

	// Enable request logging
	RequestLogging bool

	// Enable request ID middleware
	RequestID bool

	// TLS configuration
	TLSEnabled  bool
	TLSCertFile string
	TLSKeyFile  string
}

// Addr returns the server address in host:port format.
func (c ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// =============================================================================
// Database Configuration
// =============================================================================

// DatabaseConfig holds PostgreSQL configuration.
type DatabaseConfig struct {
	// Connection string (takes precedence if set)
	URL string

	// Individual connection parameters
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string

	// Connection pool settings
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration

	// Query timeout
	QueryTimeout time.Duration

	// Enable query logging
	LogQueries bool

	// Migration settings
	MigrationsPath string
	AutoMigrate    bool
}

// DSN returns the database connection string.
func (c DatabaseConfig) DSN() string {
	if c.URL != "" {
		return c.URL
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	// JWT settings
	JWTSecret          string
	JWTIssuer          string
	JWTAccessTokenTTL  time.Duration
	JWTRefreshTokenTTL time.Duration
	TokenCodeTTL       time.Duration

	// API key settings
	APIKeyPrefix string
	APIKeyLength int

	// OAuth settings
	OAuthEnabled bool

	// Session settings
	SessionSecret   string
	SessionTTL      time.Duration
	SessionSecure   bool
	SessionHTTPOnly bool
	SessionSameSite string

	// Password requirements
	PasswordMinLength      int
	PasswordRequireUpper   bool
	PasswordRequireLower   bool
	PasswordRequireDigit   bool
	PasswordRequireSpecial bool

	// Bcrypt cost
	BcryptCost int
}

// =============================================================================
// Observability Configuration
// =============================================================================

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	// Log level (debug, info, warn, error)
	Level string

	// Log format (json, text)
	Format string

	// Include caller info
	IncludeCaller bool

	// Include stack trace on errors
	IncludeStackTrace bool

	// Output (stdout, stderr, file)
	Output string

	// File path (if output is file)
	FilePath string
}

// TracingConfig holds distributed tracing configuration.
type TracingConfig struct {
	// Enable tracing
	Enabled bool

	// Provider (jaeger, zipkin, otlp)
	Provider string

	// Endpoint
	Endpoint string

	// Service name
	ServiceName string

	// Sample rate (0.0 to 1.0)
	SampleRate float64
}

// MetricsConfig holds metrics configuration.
type MetricsConfig struct {
	// Enable metrics
	Enabled bool

	// Metrics endpoint path
	Path string

	// Include runtime metrics
	IncludeRuntime bool

	// Include database metrics
	IncludeDatabase bool

	// Include HTTP metrics
	IncludeHTTP bool
}

// =============================================================================
// Feature Flags
// =============================================================================

// FeatureConfig holds feature flag configuration.
type FeatureConfig struct {
	// Enable GraphQL playground
	GraphQLPlayground bool

	// Enable API documentation
	APIDocumentation bool

	// Enable new checkout flow
	NewCheckoutFlow bool

	// Enable real-time inventory
	RealTimeInventory bool

	// Enable multi-currency
	MultiCurrency bool

	// Enable theme editor v2
	ThemeEditorV2 bool
}

// =============================================================================
// Loading Functions
// =============================================================================

// Load loads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		ServiceName: getEnv("SERVICE_NAME", "yourplatform"),
		Version:     getEnv("VERSION", "dev"),

		Server: ServerConfig{
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			Port:            getEnvInt("SERVER_PORT", 8080),
			ReadTimeout:     getEnvDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getEnvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:     getEnvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: getEnvDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
			MaxBodySize:     getEnvInt64("SERVER_MAX_BODY_SIZE", 10*1024*1024), // 10MB
			CORSEnabled:     getEnvBool("CORS_ENABLED", true),
			CORSOrigins:     getEnvSlice("CORS_ORIGINS", []string{"*"}),
			RequestLogging:  getEnvBool("REQUEST_LOGGING", true),
			RequestID:       getEnvBool("REQUEST_ID", true),
			TLSEnabled:      getEnvBool("TLS_ENABLED", false),
			TLSCertFile:     getEnv("TLS_CERT_FILE", ""),
			TLSKeyFile:      getEnv("TLS_KEY_FILE", ""),
		},

		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", ""),
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", ""),
			Password:        getEnv("DB_PASSWORD", ""),
			Database:        getEnv("DB_NAME", "platform_dev"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxConns:        int32(getEnvInt("DB_MAX_CONNS", 25)),
			MinConns:        int32(getEnvInt("DB_MIN_CONNS", 5)),
			MaxConnLifetime: getEnvDuration("DB_MAX_CONN_LIFETIME", time.Hour),
			MaxConnIdleTime: getEnvDuration("DB_MAX_CONN_IDLE_TIME", 30*time.Minute),
			QueryTimeout:    getEnvDuration("DB_QUERY_TIMEOUT", 30*time.Second),
			LogQueries:      getEnvBool("DB_LOG_QUERIES", false),
			MigrationsPath:  getEnv("DB_MIGRATIONS_PATH", "db/migrations"),
			AutoMigrate:     getEnvBool("DB_AUTO_MIGRATE", false),
		},

		Redis: RedisConfig{
			URL:            getEnv("REDIS_URL", ""),
			Host:           getEnv("REDIS_HOST", "localhost"),
			Port:           getEnvInt("REDIS_PORT", 6379),
			Password:       getEnv("REDIS_PASSWORD", ""),
			DB:             getEnvInt("REDIS_DB", 0),
			PoolSize:       getEnvInt("REDIS_POOL_SIZE", 10),
			MinIdleConns:   getEnvInt("REDIS_MIN_IDLE_CONNS", 5),
			DialTimeout:    getEnvDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:    getEnvDuration("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout:   getEnvDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
			TLSEnabled:     getEnvBool("REDIS_TLS_ENABLED", false),
			ClusterEnabled: getEnvBool("REDIS_CLUSTER_ENABLED", false),
			ClusterNodes:   getEnvSlice("REDIS_CLUSTER_NODES", nil),
		},

		// Storage: StorageConfig{
		// 	Provider:        getEnv("STORAGE_PROVIDER", "s3"),
		// 	Bucket:          getEnv("S3_BUCKET", "yourplatform-dev"),
		// 	Region:          getEnv("S3_REGION", "us-east-1"),
		// 	Endpoint:        getEnv("S3_ENDPOINT", ""),
		// 	AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
		// 	SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		// 	ForcePathStyle:  getEnvBool("S3_FORCE_PATH_STYLE", false),
		// 	CDNBaseURL:      getEnv("CDN_BASE_URL", ""),
		// 	MaxUploadSize:   getEnvInt64("MAX_UPLOAD_SIZE", 50*1024*1024), // 50MB
		// 	AllowedMimeTypes: getEnvSlice("ALLOWED_MIME_TYPES", []string{
		// 		"image/jpeg", "image/png", "image/gif", "image/webp",
		// 		"application/pdf", "text/css", "application/javascript",
		// 	}),
		// 	PresignedURLTTL: getEnvDuration("PRESIGNED_URL_TTL", 15*time.Minute),
		// 	LocalPath:       getEnv("STORAGE_LOCAL_PATH", "./storage"),
		// },

		// Kafka: KafkaConfig{
		// 	Brokers:                getEnvSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
		// 	ConsumerGroup:          getEnv("KAFKA_CONSUMER_GROUP", "yourplatform"),
		// 	TopicPrefix:            getEnv("KAFKA_TOPIC_PREFIX", ""),
		// 	SecurityProtocol:       getEnv("KAFKA_SECURITY_PROTOCOL", ""),
		// 	SASLMechanism:          getEnv("KAFKA_SASL_MECHANISM", ""),
		// 	SASLUsername:           getEnv("KAFKA_SASL_USERNAME", ""),
		// 	SASLPassword:           getEnv("KAFKA_SASL_PASSWORD", ""),
		// 	TLSEnabled:             getEnvBool("KAFKA_TLS_ENABLED", false),
		// 	ProducerTimeout:        getEnvDuration("KAFKA_PRODUCER_TIMEOUT", 10*time.Second),
		// 	ProducerRetries:        getEnvInt("KAFKA_PRODUCER_RETRIES", 3),
		// 	ProducerBatchSize:      getEnvInt("KAFKA_PRODUCER_BATCH_SIZE", 100),
		// 	ConsumerTimeout:        getEnvDuration("KAFKA_CONSUMER_TIMEOUT", 10*time.Second),
		// 	ConsumerMaxPollRecords: getEnvInt("KAFKA_CONSUMER_MAX_POLL_RECORDS", 500),
		// },

		Auth: AuthConfig{
			JWTSecret:              getEnv("JWT_SECRET", ""),
			JWTIssuer:              getEnv("JWT_ISSUER", "yourplatform"),
			TokenCodeTTL:           getEnvDuration("TokenCodeTTL", 5*time.Minute),
			JWTAccessTokenTTL:      getEnvDuration("JWT_ACCESS_TOKEN_TTL", 15*time.Minute),
			JWTRefreshTokenTTL:     getEnvDuration("JWT_REFRESH_TOKEN_TTL", 7*24*time.Hour),
			APIKeyPrefix:           getEnv("API_KEY_PREFIX", "sk_"),
			APIKeyLength:           getEnvInt("API_KEY_LENGTH", 32),
			OAuthEnabled:           getEnvBool("OAUTH_ENABLED", true),
			SessionSecret:          getEnv("SESSION_SECRET", ""),
			SessionTTL:             getEnvDuration("SESSION_TTL", 24*time.Hour),
			SessionSecure:          getEnvBool("SESSION_SECURE", true),
			SessionHTTPOnly:        getEnvBool("SESSION_HTTP_ONLY", true),
			SessionSameSite:        getEnv("SESSION_SAME_SITE", "lax"),
			PasswordMinLength:      getEnvInt("PASSWORD_MIN_LENGTH", 8),
			PasswordRequireUpper:   getEnvBool("PASSWORD_REQUIRE_UPPER", true),
			PasswordRequireLower:   getEnvBool("PASSWORD_REQUIRE_LOWER", true),
			PasswordRequireDigit:   getEnvBool("PASSWORD_REQUIRE_DIGIT", true),
			PasswordRequireSpecial: getEnvBool("PASSWORD_REQUIRE_SPECIAL", false),
			BcryptCost:             getEnvInt("BCRYPT_COST", 12),
		},

		RateLimit: RateLimitConfig{
			Enabled:        getEnvBool("RATE_LIMIT_ENABLED", true),
			DefaultRPS:     getEnvInt("RATE_LIMIT_DEFAULT_RPS", 10),
			DefaultBurst:   getEnvInt("RATE_LIMIT_DEFAULT_BURST", 50),
			KeyPrefix:      getEnv("RATE_LIMIT_KEY_PREFIX", "ratelimit:"),
			WindowDuration: getEnvDuration("RATE_LIMIT_WINDOW", time.Second),
			Plans: map[string]PlanRateLimit{
				"free":       {RPS: 2, Burst: 10},
				"starter":    {RPS: 4, Burst: 20},
				"pro":        {RPS: 10, Burst: 50},
				"enterprise": {RPS: 100, Burst: 500},
			},
		},

		// Stripe: StripeConfig{
		// 	SecretKey:      getEnv("STRIPE_SECRET_KEY", ""),
		// 	PublishableKey: getEnv("STRIPE_PUBLISHABLE_KEY", ""),
		// 	WebhookSecret:  getEnv("STRIPE_WEBHOOK_SECRET", ""),
		// 	APIVersion:     getEnv("STRIPE_API_VERSION", "2024-12-18.acacia"),
		// },

		// Email: EmailConfig{
		// 	Provider:       getEnv("EMAIL_PROVIDER", "smtp"),
		// 	SendGridAPIKey: getEnv("SENDGRID_API_KEY", ""),
		// 	SESRegion:      getEnv("SES_REGION", "us-east-1"),
		// 	SMTPHost:       getEnv("SMTP_HOST", "localhost"),
		// 	SMTPPort:       getEnvInt("SMTP_PORT", 1025),
		// 	SMTPUsername:   getEnv("SMTP_USERNAME", ""),
		// 	SMTPPassword:   getEnv("SMTP_PASSWORD", ""),
		// 	SMTPTLS:        getEnvBool("SMTP_TLS", false),
		// 	FromAddress:    getEnv("EMAIL_FROM_ADDRESS", "noreply@yourplatform.com"),
		// 	FromName:       getEnv("EMAIL_FROM_NAME", "YourPlatform"),
		// 	TemplatesPath:  getEnv("EMAIL_TEMPLATES_PATH", "templates/email"),
		// },

		// Webhook: WebhookConfig{
		// 	DeliveryTimeout:    getEnvDuration("WEBHOOK_DELIVERY_TIMEOUT", 30*time.Second),
		// 	MaxRetries:         getEnvInt("WEBHOOK_MAX_RETRIES", 5),
		// 	RetryInterval:      getEnvDuration("WEBHOOK_RETRY_INTERVAL", 5*time.Minute),
		// 	SignatureHeader:    getEnv("WEBHOOK_SIGNATURE_HEADER", "X-Webhook-Signature"),
		// 	SignatureAlgorithm: getEnv("WEBHOOK_SIGNATURE_ALGORITHM", "sha256"),
		// 	FailureThreshold:   getEnvInt("WEBHOOK_FAILURE_THRESHOLD", 10),
		// },

		Logging: LoggingConfig{
			Level:             getEnv("LOG_LEVEL", "info"),
			Format:            getEnv("LOG_FORMAT", "json"),
			IncludeCaller:     getEnvBool("LOG_INCLUDE_CALLER", false),
			IncludeStackTrace: getEnvBool("LOG_INCLUDE_STACK_TRACE", true),
			Output:            getEnv("LOG_OUTPUT", "stdout"),
			FilePath:          getEnv("LOG_FILE_PATH", ""),
		},

		Tracing: TracingConfig{
			Enabled:     getEnvBool("TRACING_ENABLED", false),
			Provider:    getEnv("TRACING_PROVIDER", "otlp"),
			Endpoint:    getEnv("TRACING_ENDPOINT", "localhost:4317"),
			ServiceName: getEnv("TRACING_SERVICE_NAME", "yourplatform"),
			SampleRate:  getEnvFloat("TRACING_SAMPLE_RATE", 0.1),
		},

		Metrics: MetricsConfig{
			Enabled:         getEnvBool("METRICS_ENABLED", true),
			Path:            getEnv("METRICS_PATH", "/metrics"),
			IncludeRuntime:  getEnvBool("METRICS_INCLUDE_RUNTIME", true),
			IncludeDatabase: getEnvBool("METRICS_INCLUDE_DATABASE", true),
			IncludeHTTP:     getEnvBool("METRICS_INCLUDE_HTTP", true),
		},

		Features: FeatureConfig{
			GraphQLPlayground: getEnvBool("FEATURE_GRAPHQL_PLAYGROUND", true),
			APIDocumentation:  getEnvBool("FEATURE_API_DOCUMENTATION", true),
			NewCheckoutFlow:   getEnvBool("FEATURE_NEW_CHECKOUT_FLOW", false),
			RealTimeInventory: getEnvBool("FEATURE_REALTIME_INVENTORY", false),
			MultiCurrency:     getEnvBool("FEATURE_MULTI_CURRENCY", false),
			ThemeEditorV2:     getEnvBool("FEATURE_THEME_EDITOR_V2", false),
		},
	}

	// Validate required configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// MustLoad loads configuration or panics on error.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}

// =============================================================================
// Validation
// =============================================================================

// Validate validates the configuration.
func (c *Config) Validate() error {
	var errs []string

	// Database validation
	if c.Database.URL == "" && c.Database.Host == "" {
		errs = append(errs, "DATABASE_URL or DB_HOST is required")
	}

	// Auth validation (only in non-development)
	if c.Environment != "development" {
		if c.Auth.JWTSecret == "" {
			errs = append(errs, "JWT_SECRET is required")
		}
		if len(c.Auth.JWTSecret) < 32 {
			errs = append(errs, "JWT_SECRET must be at least 32 characters")
		}
		if c.Auth.SessionSecret == "" {
			errs = append(errs, "SESSION_SECRET is required")
		}
	}

	// // Stripe validation (if payments enabled)
	// if c.Stripe.SecretKey != "" && !strings.HasPrefix(c.Stripe.SecretKey, "sk_") {
	// 	errs = append(errs, "STRIPE_SECRET_KEY must start with 'sk_'")
	// }

	if len(errs) > 0 {
		return fmt.Errorf("configuration errors:\n  - %s", strings.Join(errs, "\n  - "))
	}

	return nil
}

// =============================================================================
// Environment Helpers
// =============================================================================

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsStaging returns true if running in staging mode.
func (c *Config) IsStaging() bool {
	return c.Environment == "staging"
}

// IsProduction returns true if running in production mode.
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// =============================================================================
// Helper Functions
// =============================================================================

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		parts := strings.Split(value, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}
