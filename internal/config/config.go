package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for the application with high-performance settings
type Config struct {
	// Server configuration
	Port   int
	AppEnv string

	// Database configuration - Supabase (PostgreSQL) ONLY
	SupabaseURL        string // Supabase project URL
	SupabaseAnonKey    string // Supabase anonymous key
	SupabaseServiceKey string // Supabase service role key (for backend operations)
	SupabaseDBPassword string // Supabase database password (for direct PostgreSQL connections)

	// Redis configuration
	RedisURL          string
	RedisClusterAddrs []string // Support for Redis clustering

	// WhatsApp configuration
	WhatsAppStoragePath string
	WhatsAppSessionDir  string
	WhatsAppMaxDevices  int // Support for multiple devices

	// OpenRouter configuration
	OpenRouterDefaultKey string
	OpenRouterTimeout    int // Configurable timeout
	OpenRouterMaxRetries int // Max retries for AI requests

	// Security configuration
	JWTSecret     string
	SessionSecret string

	// Performance configuration
	MaxConcurrentUsers int    // Maximum concurrent users
	WebSocketEnabled   bool   // Enable WebSocket support
	CDNEnabled         bool   // Enable CDN for media files
	CDNBaseURL         string // CDN base URL
}

// Load loads configuration from environment variables with performance optimizations
func Load() *Config {
	cfg := &Config{
		// Server configuration - Railway sets PORT at runtime
		Port:   getEnvAsInt("PORT", 8080),
		AppEnv: getEnv("APP_ENV", "development"),

		// Supabase configuration (REQUIRED)
		SupabaseURL:        getEnv("SUPABASE_URL", ""),
		SupabaseAnonKey:    getEnv("SUPABASE_ANON_KEY", ""),
		SupabaseServiceKey: getEnv("SUPABASE_SERVICE_KEY", ""),
		SupabaseDBPassword: getEnv("SUPABASE_DB_PASSWORD", ""),

		// Redis configuration with clustering support
		RedisURL:          getEnv("REDIS_URL", ""),
		RedisClusterAddrs: getEnvAsSlice("REDIS_CLUSTER_ADDRS", ","),

		// WhatsApp configuration with multi-device support
		WhatsAppStoragePath: getEnv("WHATSAPP_STORAGE_PATH", "./whatsapp_sessions"),
		WhatsAppSessionDir:  getEnv("WHATSAPP_SESSION_DIR", "./whatsapp_sessions"),
		WhatsAppMaxDevices:  getEnvAsInt("WHATSAPP_MAX_DEVICES", 10),

		// OpenRouter configuration with performance settings
		OpenRouterDefaultKey: getEnv("OPENROUTER_DEFAULT_KEY", ""),
		OpenRouterTimeout:    getEnvAsInt("OPENROUTER_TIMEOUT", 15), // Reduced from 30s
		OpenRouterMaxRetries: getEnvAsInt("OPENROUTER_MAX_RETRIES", 2),

		// Security configuration
		JWTSecret:     getEnv("JWT_SECRET", "your-jwt-secret-key"),
		SessionSecret: getEnv("SESSION_SECRET", "your-session-secret-key"),

		// Performance configuration for 3000+ concurrent users
		MaxConcurrentUsers: getEnvAsInt("MAX_CONCURRENT_USERS", 5000),
		WebSocketEnabled:   getEnvAsBool("WEBSOCKET_ENABLED", true),
		CDNEnabled:         getEnvAsBool("CDN_ENABLED", false),
		CDNBaseURL:         getEnv("CDN_BASE_URL", ""),
	}

	return cfg
}

// IsProduction returns true if the app is running in production
func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

// IsDevelopment returns true if the app is running in development
func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development"
}

// getEnv gets an environment variable with a fallback value
// Trims whitespace to handle Railway environment variable formatting
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}

// getEnvAsInt gets an environment variable as an integer with a fallback value
func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

// getEnvAsBool gets an environment variable as a boolean with a fallback value
func getEnvAsBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return fallback
}

// getEnvAsSlice gets an environment variable as a slice with a separator
func getEnvAsSlice(key, separator string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, separator)
	}
	return []string{}
}
