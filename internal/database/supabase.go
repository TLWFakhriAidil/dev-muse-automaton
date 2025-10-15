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

	// Validate Supabase URL format
	if err := validateSupabaseURL(cfg.SupabaseURL); err != nil {
		return nil, fmt.Errorf("invalid SUPABASE_URL: %w", err)
	}

	// Log which database is being used
	logrus.WithField("url", cfg.SupabaseURL).Info("Connecting to Supabase (PostgreSQL) database")

	// Initialize Supabase client
	client, err := supa.NewClient(cfg.SupabaseURL, cfg.SupabaseServiceKey, &supa.ClientOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create Supabase client: %w", err)
	}

	// Build PostgreSQL connection string
	// Supabase uses PostgreSQL, so we extract connection details from the Supabase URL
	// Connection format uses service key for authentication
	postgresURI, err := buildPostgresURI(cfg.SupabaseURL, cfg.SupabaseServiceKey)
	if err != nil {
		return nil, fmt.Errorf("failed to build PostgreSQL connection string: %w", err)
	}

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

// validateSupabaseURL validates the format of the Supabase URL
func validateSupabaseURL(supabaseURL string) error {
	if supabaseURL == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	// Check for duplicate https: prefix (common Railway mistake)
	// Example: "https:https://..." should be "https://..."
	if len(supabaseURL) > 6 && supabaseURL[:6] == "https:" && supabaseURL[6:13] == "https://" {
		return fmt.Errorf("duplicate 'https:' prefix detected - got: %s\n\n"+
			"❌ Your URL: %s\n"+
			"✅ Should be: %s\n\n"+
			"Fix: In Railway, edit SUPABASE_URL and remove the duplicate 'https:' prefix",
			supabaseURL,
			supabaseURL,
			supabaseURL[6:])
	}

	// Check for https:// prefix
	if len(supabaseURL) < 8 || supabaseURL[:8] != "https://" {
		return fmt.Errorf("URL must start with 'https://' - got: %s", supabaseURL)
	}

	// Check for .supabase.co suffix
	url := supabaseURL
	if url[len(url)-1] == '/' {
		url = url[:len(url)-1]
	}
	if len(url) < 20 || url[len(url)-12:] != ".supabase.co" {
		return fmt.Errorf("URL must end with '.supabase.co' - got: %s", supabaseURL)
	}

	// Extract project reference
	start := 8 // len("https://")
	end := len(url) - 12 // len(".supabase.co")
	projectRef := url[start:end]

	if projectRef == "" || len(projectRef) < 10 {
		return fmt.Errorf("invalid project reference extracted from URL (too short): %s", projectRef)
	}

	// Check for common mistakes
	if projectRef == "your-project-ref" || projectRef == "your-project" || projectRef == "xxxxx" {
		return fmt.Errorf("please replace the placeholder URL with your actual Supabase project URL from https://app.supabase.com")
	}

	logrus.WithField("project_ref", projectRef).Info("Supabase URL validation passed")
	return nil
}

// buildPostgresURI constructs a PostgreSQL connection URI from Supabase configuration
func buildPostgresURI(supabaseURL, serviceToken string) (string, error) {
	// Supabase provides a REST API URL like: https://xxxxx.supabase.co
	// We need to convert this to PostgreSQL connection string using the service role token
	
	// Extract project reference from Supabase URL
	// URL format: https://xxxxx.supabase.co or https://xxxxx.supabase.co/
	url := supabaseURL
	if url[len(url)-1] == '/' {
		url = url[:len(url)-1]
	}
	
	// Extract between https:// and .supabase.co
	start := 8 // len("https://")
	end := len(url) - 12 // len(".supabase.co")
	
	if end <= start {
		return "", fmt.Errorf("invalid URL format: cannot extract project reference from %s", supabaseURL)
	}
	
	projectRef := url[start:end]
	
	if projectRef == "" {
		return "", fmt.Errorf("empty project reference extracted from URL: %s", supabaseURL)
	}

	// Build PostgreSQL connection URI using service role token for auth
	// Use connection pooler (port 6543) instead of direct connection (port 5432)
	// This provides better IPv4 support and connection pooling for Railway deployments
	host := fmt.Sprintf("db.%s.supabase.co", projectRef)
	
	// Use port 6543 (connection pooler) with transaction mode for better compatibility
	// Connection pooler supports IPv4 and provides better stability on Railway
	uri := fmt.Sprintf("postgres://postgres.%s:%s@aws-0-us-east-1.pooler.supabase.com:6543/postgres?sslmode=require", projectRef, serviceToken)
	
	logrus.WithFields(logrus.Fields{
		"project_ref": projectRef,
		"host": host,
	}).Debug("Built PostgreSQL connection URI")
	
	return uri, nil
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
