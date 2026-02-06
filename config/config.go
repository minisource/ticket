package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the ticket service
type Config struct {
	Server   ServerConfig
	MongoDB  MongoDBConfig
	Redis    RedisConfig
	Auth     AuthConfig
	Notifier NotifierConfig
	SLA      SLAConfig
	Ticket   TicketConfig
	Logging  LoggingConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port            int
	Host            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// MongoDBConfig holds MongoDB configuration
type MongoDBConfig struct {
	URI             string
	Database        string
	MaxPoolSize     uint64
	MinPoolSize     uint64
	MaxConnIdleTime time.Duration
}

// RedisConfig holds Redis configuration for caching
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// AuthConfig holds auth service configuration
type AuthConfig struct {
	ServiceURL        string
	IntrospectionPath string
	ClientID          string
	ClientSecret      string
	CacheSeconds      int
	SkipPaths         []string
}

// NotifierConfig holds notifier service configuration
type NotifierConfig struct {
	ServiceURL   string
	ClientID     string
	ClientSecret string
	Enabled      bool
}

// SLAConfig holds SLA configuration
type SLAConfig struct {
	Enabled              bool
	DefaultResponseHours int
	DefaultResolveHours  int
	EscalationEnabled    bool
	BusinessHoursStart   int
	BusinessHoursEnd     int
	WorkDays             []int
}

// TicketConfig holds ticket-specific configuration
type TicketConfig struct {
	MaxTitleLength          int
	MaxDescriptionLength    int
	MaxAttachmentsPerTicket int
	MaxAttachmentSizeMB     int
	AllowedFileTypes        []string
	AutoAssignEnabled       bool
	RequireDepartment       bool
	AllowCustomerClose      bool
	RateLimitPerMinute      int
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	_ = godotenv.Load()

	return &Config{
		Server: ServerConfig{
			Port:            getEnvAsInt("SERVER_PORT", 5011),
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:     getDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout:    getDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			ShutdownTimeout: getDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		MongoDB: MongoDBConfig{
			URI:             getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			Database:        getEnv("MONGODB_DATABASE", "minisource_tickets"),
			MaxPoolSize:     uint64(getEnvAsInt("MONGODB_MAX_POOL_SIZE", 100)),
			MinPoolSize:     uint64(getEnvAsInt("MONGODB_MIN_POOL_SIZE", 10)),
			MaxConnIdleTime: getDuration("MONGODB_MAX_CONN_IDLE_TIME", 30*time.Minute),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 3),
		},
		Auth: AuthConfig{
			ServiceURL:        getEnv("AUTH_SERVICE_URL", "http://localhost:5001"),
			IntrospectionPath: getEnv("AUTH_INTROSPECTION_PATH", "/api/v1/oauth/introspect"),
			ClientID:          getEnv("AUTH_CLIENT_ID", "ticket-service"),
			ClientSecret:      getEnv("AUTH_CLIENT_SECRET", "ticket-service-secret-key"),
			CacheSeconds:      getEnvAsInt("AUTH_CACHE_SECONDS", 300),
			SkipPaths:         getEnvAsSlice("AUTH_SKIP_PATHS", []string{"/health", "/ready", "/metrics"}),
		},
		Notifier: NotifierConfig{
			ServiceURL:   getEnv("NOTIFIER_SERVICE_URL", "http://localhost:5003"),
			ClientID:     getEnv("NOTIFIER_CLIENT_ID", "ticket-service"),
			ClientSecret: getEnv("NOTIFIER_CLIENT_SECRET", "ticket-service-secret-key"),
			Enabled:      getEnvAsBool("NOTIFIER_ENABLED", true),
		},
		SLA: SLAConfig{
			Enabled:              getEnvAsBool("SLA_ENABLED", true),
			DefaultResponseHours: getEnvAsInt("SLA_DEFAULT_RESPONSE_HOURS", 24),
			DefaultResolveHours:  getEnvAsInt("SLA_DEFAULT_RESOLVE_HOURS", 72),
			EscalationEnabled:    getEnvAsBool("SLA_ESCALATION_ENABLED", true),
			BusinessHoursStart:   getEnvAsInt("SLA_BUSINESS_HOURS_START", 9),
			BusinessHoursEnd:     getEnvAsInt("SLA_BUSINESS_HOURS_END", 17),
			WorkDays:             getEnvAsIntSlice("SLA_WORK_DAYS", []int{1, 2, 3, 4, 5}),
		},
		Ticket: TicketConfig{
			MaxTitleLength:          getEnvAsInt("TICKET_MAX_TITLE_LENGTH", 200),
			MaxDescriptionLength:    getEnvAsInt("TICKET_MAX_DESCRIPTION_LENGTH", 10000),
			MaxAttachmentsPerTicket: getEnvAsInt("TICKET_MAX_ATTACHMENTS", 10),
			MaxAttachmentSizeMB:     getEnvAsInt("TICKET_MAX_ATTACHMENT_SIZE_MB", 10),
			AllowedFileTypes:        getEnvAsSlice("TICKET_ALLOWED_FILE_TYPES", []string{"jpg", "jpeg", "png", "gif", "pdf", "doc", "docx", "txt", "zip"}),
			AutoAssignEnabled:       getEnvAsBool("TICKET_AUTO_ASSIGN_ENABLED", true),
			RequireDepartment:       getEnvAsBool("TICKET_REQUIRE_DEPARTMENT", true),
			AllowCustomerClose:      getEnvAsBool("TICKET_ALLOW_CUSTOMER_CLOSE", true),
			RateLimitPerMinute:      getEnvAsInt("TICKET_RATE_LIMIT_PER_MINUTE", 10),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}, nil
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

func getEnvAsIntSlice(key string, defaultValue []int) []int {
	if value := os.Getenv(key); value != "" {
		parts := strings.Split(value, ",")
		result := make([]int, 0, len(parts))
		for _, part := range parts {
			if v, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
				result = append(result, v)
			}
		}
		return result
	}
	return defaultValue
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
