package database

import (
	"database/sql"
	"fmt"

	"nodepath-chat/internal/config"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/sirupsen/logrus"
	supa "github.com/supabase-community/supabase-go"
)

// SupabaseClient wraps the Supabase client and SQL connection
type SupabaseClient struct {
	Client *supa.Client
	DB     *sql.DB
}

// InitializeSupabase creates and returns a Supabase client connection
func InitializeSupabase(cfg *config.Config) (*SupabaseClient, error) {
	// Check if Supabase credentials are provided
	if cfg.SupabaseURL == "" || cfg.SupabaseServiceKey == "" {
		return nil, fmt.Errorf("SUPABASE_URL and SUPABASE_SERVICE_KEY environment variables are required")
	}

	// Log which database is being used
	logrus.Info("Connecting to Supabase (PostgreSQL) database")

	// Initialize Supabase client
	client, err := supa.NewClient(cfg.SupabaseURL, cfg.SupabaseServiceKey, &supa.ClientOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create Supabase client: %w", err)
	}

	// Build PostgreSQL connection string
	// Supabase uses PostgreSQL, so we extract connection details from the Supabase URL
	// Connection format uses service key for authentication
	postgresURI := buildPostgresURI(cfg.SupabaseURL, cfg.SupabaseServiceKey)

	// Open PostgreSQL connection for standard database/sql operations
	db, err := sql.Open("postgres", postgresURI)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Configure connection pool for high concurrency (3000+ users)
	// Optimized settings for handling 3000+ concurrent users with real-time messaging
	db.SetMaxOpenConns(500)   // Increased significantly for 3000+ concurrent users
	db.SetMaxIdleConns(100)   // Higher idle connections to reduce connection overhead
	db.SetConnMaxLifetime(60) // Longer lifetime to reduce connection churn (in minutes)
	db.SetConnMaxIdleTime(15) // Balanced idle time for resource efficiency (in minutes)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping Supabase database: %w", err)
	}

	logrus.Info("Supabase database connection established successfully")

	return &SupabaseClient{
		Client: client,
		DB:     db,
	}, nil
}

// buildPostgresURI constructs a PostgreSQL connection URI from Supabase configuration
func buildPostgresURI(supabaseURL, serviceToken string) string {
	// Supabase provides a REST API URL like: https://xxxxx.supabase.co
	// We need to convert this to PostgreSQL connection string using the service role token
	
	// Extract project reference from Supabase URL
	// URL format: https://xxxxx.supabase.co or https://xxxxx.supabase.co/
	projectRef := ""
	if len(supabaseURL) > 8 { // https://
		url := supabaseURL
		if url[len(url)-1] == '/' {
			url = url[:len(url)-1]
		}
		// Extract between https:// and .supabase.co
		start := 8 // len("https://")
		if idx := len(url) - len(".supabase.co"); idx > start {
			projectRef = url[start:idx]
		}
	}

	// Build PostgreSQL connection URI using service role token for auth
	host := fmt.Sprintf("db.%s.supabase.co", projectRef)
	uri := fmt.Sprintf("postgres://postgres:%s@%s:5432/postgres?sslmode=require", serviceToken, host)
	
	return uri
}

// RunSupabaseMigrations runs all database migrations for Supabase (PostgreSQL)
// Note: This function will need to be updated with PostgreSQL-compatible SQL
func RunSupabaseMigrations(db *sql.DB) error {
	logrus.Info("Running Supabase (PostgreSQL) database migrations")

	// For now, we'll log that migrations need to be converted
	// The actual migration SQL needs to be converted from MySQL to PostgreSQL syntax
	logrus.Warn("Database migrations need to be converted from MySQL to PostgreSQL syntax")
	logrus.Warn("Please run Supabase migrations using the Supabase CLI or Dashboard")

	// TODO: Convert MySQL migrations to PostgreSQL
	// Key differences:
	// 1. AUTO_INCREMENT -> SERIAL or IDENTITY
	// 2. VARCHAR(255) -> VARCHAR(255) or TEXT
	// 3. TINYINT(1) -> BOOLEAN
	// 4. ENUM -> CHECK constraint or custom type
	// 5. JSON -> JSONB (better performance)
	// 6. TIMESTAMP -> TIMESTAMP WITH TIME ZONE
	// 7. INDEX syntax differences
	// 8. COLLATION differences

	return nil
}
