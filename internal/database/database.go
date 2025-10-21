package database

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"nodepath-chat/internal/config"
	_ "github.com/lib/pq" // PostgreSQL driver for Supabase
	"github.com/sirupsen/logrus"
)

// resolveIPv4 forcefully resolves a hostname to its IPv4 address to avoid IPv6 issues in Railway
func resolveIPv4(hostname string) (string, error) {
	// First try using context with explicit IPv4 network type
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Force IPv4-only resolution using net.Dialer
	dialer := &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	
	// Try to dial to get the actual IPv4 address used
	conn, err := dialer.DialContext(ctx, "tcp4", hostname+":5432")
	if err == nil {
		defer conn.Close()
		if tcpAddr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
			if ipv4 := tcpAddr.IP.To4(); ipv4 != nil {
				logrus.WithFields(logrus.Fields{
					"hostname": hostname,
					"ipv4":     ipv4.String(),
					"method":   "tcp4_dial",
				}).Info("Successfully resolved hostname to IPv4 via TCP4 dial")
				return ipv4.String(), nil
			}
		}
	}
	
	// Fallback: try explicit A record lookup
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return "", fmt.Errorf("failed to lookup IP for %s: %w", hostname, err)
	}
	
	// Find the first IPv4 address
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			logrus.WithFields(logrus.Fields{
				"hostname": hostname,
				"ipv4":     ipv4.String(),
				"method":   "lookup_fallback",
			}).Info("Resolved hostname to IPv4 via lookup fallback")
			return ipv4.String(), nil
		}
	}
	
	// Final fallback: use known Supabase IPv4 ranges or external DNS
	if strings.Contains(hostname, "supabase.co") {
		// Try using Google DNS to get IPv4 address
		r := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Second * 5,
				}
				return d.DialContext(ctx, "udp4", "8.8.8.8:53")
			},
		}
		
		ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel2()
		
		addrs, err := r.LookupHost(ctx2, hostname)
		if err == nil {
			for _, addr := range addrs {
				if ip := net.ParseIP(addr); ip != nil && ip.To4() != nil {
					logrus.WithFields(logrus.Fields{
						"hostname": hostname,
						"ipv4":     addr,
						"method":   "google_dns",
					}).Info("Resolved hostname to IPv4 via Google DNS")
					return addr, nil
				}
			}
		}
	}
	
	return "", fmt.Errorf("no IPv4 address found for hostname: %s after trying multiple methods", hostname)
}

// Initialize creates and returns a Supabase PostgreSQL database connection
func Initialize(cfg *config.Config) (*sql.DB, error) {
	if cfg.SupabaseURL == "" || cfg.SupabaseDBPassword == "" {
		return nil, fmt.Errorf("Failed to initialize Supabase database - check your connection settings")
	}

	// RAILWAY FIX: Use CGO resolver for better IPv4 compatibility
	os.Setenv("GODEBUG", "netdns=cgo")

	logrus.Info("🚀 Initializing Supabase PostgreSQL database connection (Railway IPv4 mode)")

	// Build PostgreSQL connection string from Supabase URL
	projectRef := extractProjectRef(cfg.SupabaseURL)
	logrus.WithField("project_ref", projectRef).Debug("Extracted project reference")

	// RAILWAY FIX: Try connection pooler first (better IPv4 support)
	var connStr string
	hostname := fmt.Sprintf("db.%s.supabase.co", projectRef)
	
	// STRATEGY 1: Try Connection Pooler (better IPv4/IPv6 compatibility)
	// Supabase uses Supavisor pooler with format: postgres.[PROJECT_REF]
	// Transaction mode (6543) is recommended for serverless/Railway
	poolerStrategies := []struct {
		name string
		connStr string
	}{
		{
			name: "POOLER-TRANSACTION-6543",
			connStr: fmt.Sprintf("postgres://postgres.%s:%s@aws-0-us-west-1.pooler.supabase.com:6543/postgres",
				projectRef, cfg.SupabaseDBPassword),
		},
		{
			name: "POOLER-SESSION-5432",
			connStr: fmt.Sprintf("postgres://postgres.%s:%s@aws-0-us-west-1.pooler.supabase.com:5432/postgres",
				projectRef, cfg.SupabaseDBPassword),
		},
	}

	for _, strategy := range poolerStrategies {
		logrus.WithField("strategy", strategy.name).Info("🔄 Attempting Supabase pooler connection")

		testDB, err := sql.Open("postgres", strategy.connStr)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			err = testDB.PingContext(ctx)
			cancel()
			testDB.Close()

			if err == nil {
				logrus.WithField("strategy", strategy.name).Info("✅ Pooler connection successful!")
				connStr = strategy.connStr
				break
			}
		}
	}

	// STRATEGY 2: Try direct connection with sslmode variations
	if connStr == "" {
		logrus.Info("🔄 Pooler failed, trying direct connection strategies")
		directStrategies := []struct {
			name string
			connStr string
		}{
			{
				name: "DIRECT-SSL-DISABLE",
				connStr: fmt.Sprintf("postgresql://postgres:%s@%s:5432/postgres?sslmode=disable&connect_timeout=10",
					cfg.SupabaseDBPassword, hostname),
			},
			{
				name: "DIRECT-SSL-REQUIRE",
				connStr: fmt.Sprintf("host=%s port=5432 user=postgres dbname=postgres password=%s sslmode=require connect_timeout=10",
					hostname, cfg.SupabaseDBPassword),
			},
			{
				name: "DIRECT-IPV4-FORCE",
				connStr: fmt.Sprintf("host=%s port=5432 user=postgres dbname=postgres password=%s sslmode=disable connect_timeout=15 target_session_attrs=read-write",
					hostname, cfg.SupabaseDBPassword),
			},
		}

		for _, strategy := range directStrategies {
			logrus.WithField("strategy", strategy.name).Info("🔄 Attempting direct connection")

			testDB, err := sql.Open("postgres", strategy.connStr)
			if err == nil {
				ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
				err = testDB.PingContext(ctx)
				cancel()
				testDB.Close()

				if err == nil {
					logrus.WithField("strategy", strategy.name).Info("✅ Direct connection successful!")
					connStr = strategy.connStr
					break
				}
			}
		}
	}

	// STRATEGY 3: IPv4 resolution fallback
	if connStr == "" {
		logrus.Info("🔄 All strategies failed, attempting IPv4 resolution")
		ipv4Address, err := resolveIPv4(hostname)
		if err == nil {
			logrus.WithField("ipv4", ipv4Address).Info("Resolved IPv4, using hostaddr")
			connStr = fmt.Sprintf("hostaddr=%s host=%s port=5432 user=postgres dbname=postgres password=%s sslmode=require connect_timeout=15",
				ipv4Address, hostname, cfg.SupabaseDBPassword)
		} else {
			logrus.Error("🚨 IPv4 resolution failed - using last resort hostname connection")
			connStr = fmt.Sprintf("host=%s port=5432 user=postgres dbname=postgres password=%s sslmode=require connect_timeout=20",
				hostname, cfg.SupabaseDBPassword)
		}
	}
	
	logrus.WithField("connection_string", strings.ReplaceAll(connStr, cfg.SupabaseDBPassword, "***")).Debug("Using Railway-optimized connection string")
	
	// Open PostgreSQL connection
	logrus.WithField("hostname", hostname).Info("Connecting to Supabase with IPv4-only configuration")
	
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logrus.WithError(err).WithField("connection_string", strings.ReplaceAll(connStr, cfg.SupabaseDBPassword, "[REDACTED]")).Error("Failed to open database connection")
		return nil, fmt.Errorf("failed to open Supabase PostgreSQL connection: %w", err)
	}
	
	logrus.Info("Database connection opened, attempting to ping")

	// Configure connection pool for high concurrency (3000+ users)
	// Optimized settings for handling 3000+ concurrent users with real-time messaging
	db.SetMaxOpenConns(500)   // Increased significantly for 3000+ concurrent users
	db.SetMaxIdleConns(100)   // Higher idle connections to reduce connection overhead
	db.SetConnMaxLifetime(60 * time.Minute) // Longer lifetime to reduce connection churn (in minutes)
	db.SetConnMaxIdleTime(15 * time.Minute) // Balanced idle time for resource efficiency (in minutes)

	// RAILWAY FIX: Reduced retry logic for faster startup
	var pingErr error
	maxRetries := 5  // Reduced for Railway quick startup
	retryDelay := 2 * time.Second  // Shorter delay for Railway
	
	// Prepare fallback connection strings
	localhostConnStr := fmt.Sprintf("host=localhost port=5432 user=postgres dbname=postgres password=%s sslmode=disable connect_timeout=60", 
		cfg.SupabaseDBPassword)
	
	for i := 0; i < maxRetries; i++ {
		pingErr = db.Ping()
		if pingErr == nil {
			break
		}
		
		// Try localhost connection at attempt 5
		if i == 5 {
			logrus.Info("Attempting localhost connection as fallback (attempt 5)")
			db.Close()
			logrus.WithField("connection_string", strings.ReplaceAll(localhostConnStr, cfg.SupabaseDBPassword, "[REDACTED]")).
				Info("Using localhost connection string")
			
			var dbLocal *sql.DB
			var errLocal error
			dbLocal, errLocal = sql.Open("postgres", localhostConnStr)
			if errLocal != nil {
				logrus.WithError(errLocal).Error("Failed to open localhost connection")
				// Reopen with original connection string
				logrus.Info("Reopening with original connection string")
				db, _ = sql.Open("postgres", connStr)
			} else {
				pingErr := dbLocal.Ping()
				if pingErr == nil {
					logrus.Info("Successfully connected to database via localhost")
					return dbLocal, nil
				}
				logrus.WithError(pingErr).Error("Failed to ping localhost connection, reverting to original connection")
				dbLocal.Close()
				db, _ = sql.Open("postgres", connStr)
			}
		}
		
		// Special handling for IPv6 errors - try to force IPv4 as last resort
		if strings.Contains(pingErr.Error(), "network is unreachable") || 
		   strings.Contains(pingErr.Error(), "IPv6") {
			logrus.Info("Detected IPv6 network issue, attempting IPv4-only connection")
			
			// Try to get a fresh IPv4 address with our improved resolution
			freshIPv4, resolveErr := resolveIPv4(hostname)
			if resolveErr != nil {
				logrus.WithError(resolveErr).Warn("Failed to resolve fresh IPv4 address")
				
				// Last resort: try a hardcoded approach for Railway
				logrus.Info("Attempting last resort IPv4 connection using external DNS")
				fallbackConnStr := fmt.Sprintf("host=%s port=5432 user=postgres dbname=postgres password=%s sslmode=require connect_timeout=30 application_name=railway-ipv4-fallback", 
					hostname, cfg.SupabaseDBPassword)
				
				db.Close()
				db, err = sql.Open("postgres", fallbackConnStr)
				if err != nil {
					logrus.WithError(err).Error("Failed to create fallback connection")
					// Continue with original connection
					db, _ = sql.Open("postgres", connStr)
				}
			} else {
				// Use the fresh IPv4 address with hostaddr to force IPv4
				ipv4Conn := fmt.Sprintf("hostaddr=%s port=5432 user=postgres dbname=postgres password=%s sslmode=require connect_timeout=30 application_name=railway-ipv4", 
					freshIPv4, cfg.SupabaseDBPassword)
				
				logrus.WithField("ipv4", freshIPv4).Info("Using fresh IPv4 address for Railway compatibility")
				
				// Try to close and reopen with IPv4-only connection
				db.Close()
				db, err = sql.Open("postgres", ipv4Conn)
				if err != nil {
					logrus.WithError(err).Error("Failed to create IPv4-only connection")
					// Continue with original connection
					db, _ = sql.Open("postgres", connStr)
				}
			}
		}
		
		logrus.WithFields(logrus.Fields{
			"attempt": i + 1,
			"max_retries": maxRetries,
			"error": pingErr.Error(),
			"environment": "production",
		}).Warn("Failed to ping Supabase database, retrying...")
		
		if i < maxRetries-1 {
			logrus.WithField("delay_seconds", retryDelay.Seconds()).Info("Waiting before retry...")
			time.Sleep(retryDelay)
			// Increase delay for next retry (exponential backoff with higher factor for production)
			retryDelay = time.Duration(float64(retryDelay) * 2.0)
		}
	}
	
	if pingErr != nil {
		// Check if the error contains IPv6 connection issue or connection refused
		if strings.Contains(pingErr.Error(), "network is unreachable") || 
		   strings.Contains(pingErr.Error(), "connection refused") {
			logrus.Error("Connection issue detected - forcing direct connection to Supabase")
			
			// Force direct connection to Supabase as last resort
			directConnStr := fmt.Sprintf("host=%s port=5432 user=postgres dbname=postgres password=%s sslmode=disable", 
				hostname, cfg.SupabaseDBPassword)
			
			// Add IPv4 hints
			directConnStr += " target_session_attrs=read-write connect_timeout=300"
			
			// Try one more time with direct connection
			db.Close()
			db, err = sql.Open("postgres", directConnStr)
			if err != nil {
				return nil, fmt.Errorf("failed to open direct Supabase connection: %w", err)
			}
			
			// Final attempt
			if err = db.Ping(); err != nil {
				// Last resort - try with a completely different approach
				logrus.Error("Final attempt failed - trying with explicit IPv4 connection")
				finalConnStr := fmt.Sprintf("postgresql://postgres:%s@%s:5432/postgres?sslmode=disable", 
					cfg.SupabaseDBPassword, hostname)
				
				db.Close()
				db, err = sql.Open("postgres", finalConnStr)
				if err != nil || db.Ping() != nil {
					return nil, fmt.Errorf("all connection attempts to Supabase failed: %w", err)
				}
			}
			
			logrus.Info("✅ Supabase connection established with direct connection")
			return db, nil
		}
		
		return nil, fmt.Errorf("failed to ping Supabase PostgreSQL database after %d attempts: %w", maxRetries, pingErr)
	}

	logrus.Info("✅ Supabase PostgreSQL database connection established successfully")
	return db, nil
}

// extractProjectRef extracts the project reference from Supabase URL
// Example: https://abcdefghijklmnop.supabase.co -> abcdefghijklmnop
func extractProjectRef(supabaseURL string) string {
	// Remove protocol
	url := strings.TrimPrefix(supabaseURL, "https://")
	url = strings.TrimPrefix(url, "http://")
	
	// Extract project reference (everything before .supabase.co)
	parts := strings.Split(url, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	
	return url
}

// RunMigrations runs all database migrations
func RunMigrations(db *sql.DB) error {
	logrus.Info("Running database migrations")

	// Test chat executions table removed from migrations
	migrations := []string{
		createFlowsTable,
		createDeviceSettingsTable,
		createUsersTable,
		createUserSessionsTable,
		createAIWhatsappTable,
		createWasapBotTable,
		createConversationLogTable,
		createOrdersTable,
		createAIWhatsappSessionTable,
		createWasapBotSessionTable,
	}

	for i, migration := range migrations {
		logrus.WithField("migration", i+1).Debug("Running migration")
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to run migration %d: %w", i+1, err)
		}
	}

	// Remove deprecated columns
	if err := removeDeprecatedColumnsFromFlowsTable(db); err != nil {
		logrus.WithError(err).Warn("Some deprecated columns may not exist, continuing...")
	}

	// Add missing columns with error handling
	if err := addMissingColumnsToFlowsTable(db); err != nil {
		logrus.WithError(err).Warn("Some columns may already exist, continuing...")
	}

	if err := addMissingColumnsToDeviceSettingsTable(db); err != nil {
		logrus.WithError(err).Warn("Some device settings columns may already exist, continuing...")
	}

	// Update provider values from 'rvsb_wasap' to 'waha'
	if err := updateProviderRvsbWasapToWaha(db); err != nil {
		logrus.WithError(err).Warn("Failed to update provider values, continuing...")
	}

	// Update provider ENUM to include 'waha' (critical for WAHA provider support)
	if err := updateProviderEnum(db); err != nil {
		logrus.WithError(err).Warn("Failed to update provider ENUM, continuing...")
	}

	// Drop deprecated billing columns from orders
	if err := dropDeprecatedBillingColumns(db); err != nil {
		logrus.WithError(err).Warn("Failed to drop deprecated billing columns, continuing...")
	}

	// Convert user_id column from INT to CHAR(36) in orders
	if err := convertUserIDToChar36(db); err != nil {
		logrus.WithError(err).Warn("Failed to convert user_id column in orders, continuing...")
	}

	// Convert user_id column from INT to CHAR(36) in device_setting
	if err := convertDeviceUserIDToChar36(db); err != nil {
		logrus.WithError(err).Warn("Failed to convert user_id column in device_setting, continuing...")
	}

	// Make device_id nullable for manual device creation
	if err := makeDeviceIDNullable(db); err != nil {
		logrus.WithError(err).Warn("Failed to make device_id nullable in device_setting, continuing...")
	}

	logrus.Info("Database migrations completed successfully")
	return nil
}

// Migration SQL statements - PostgreSQL format for Supabase
const createFlowsTable = `
CREATE TABLE IF NOT EXISTS chatbot_flows (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    niche TEXT,
    id_device VARCHAR(255),
    nodes JSONB,
    edges JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_chatbot_flows_updated_at 
    BEFORE UPDATE ON chatbot_flows 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
`

// Test chat executions table schema removed

const createDeviceSettingsTable = `
CREATE TABLE IF NOT EXISTS device_setting (
    id VARCHAR(255) PRIMARY KEY,
    device_id VARCHAR(255),
    api_key_option VARCHAR(100) DEFAULT 'openai/gpt-4.1' CHECK (api_key_option IN ('openai/gpt-5-chat', 'openai/gpt-5-mini', 'openai/chatgpt-4o-latest', 'openai/gpt-4.1', 'google/gemini-2.5-pro', 'google/gemini-pro-1.5')),
    webhook_id VARCHAR(500),
    provider VARCHAR(20) DEFAULT 'wablas' CHECK (provider IN ('whacenter', 'wablas', 'waha')),
    phone_number VARCHAR(20),
    api_key TEXT,
    id_device VARCHAR(255),
    id_erp VARCHAR(255),
    id_admin VARCHAR(255),
    user_id CHAR(36),
    instance TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_device_setting_user_id ON device_setting(user_id);
CREATE INDEX IF NOT EXISTS idx_device_setting_id_device ON device_setting(id_device);

CREATE TRIGGER update_device_setting_updated_at 
    BEFORE UPDATE ON device_setting 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
`

// Create users table for authentication (matches existing schema)
const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
	id CHAR(36) NOT NULL PRIMARY KEY,
	email VARCHAR(255) NOT NULL,
	full_name VARCHAR(255) NOT NULL,
	password_hash VARCHAR(255) NOT NULL,
	is_active BOOLEAN DEFAULT true,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	last_login TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
`

// Create user_sessions table for authentication sessions (matches existing schema)
const createUserSessionsTable = `
CREATE TABLE IF NOT EXISTS user_sessions (
	id CHAR(36) NOT NULL PRIMARY KEY,
	user_id CHAR(36) NOT NULL,
	token VARCHAR(255) NOT NULL,
	expires_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
`

// AI WhatsApp conversation table for managing AI-powered WhatsApp conversations
const createAIWhatsappTable = `
CREATE TABLE IF NOT EXISTS ai_whatsapp (
    id_prospect SERIAL PRIMARY KEY,
    flow_reference VARCHAR(255) DEFAULT NULL,
    execution_id VARCHAR(255) DEFAULT NULL,
    date_order TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    id_device VARCHAR(255) DEFAULT NULL,
    niche VARCHAR(255) DEFAULT NULL,
    prospect_name VARCHAR(255) DEFAULT NULL,
    prospect_num VARCHAR(255) DEFAULT NULL,
    intro VARCHAR(255) DEFAULT NULL,
    stage VARCHAR(255) DEFAULT NULL,
    conv_last TEXT,
    conv_current TEXT,
    execution_status VARCHAR(20) DEFAULT NULL CHECK (execution_status IN ('active','completed','failed')),
    flow_id VARCHAR(255) DEFAULT NULL,
    current_node_id VARCHAR(255) DEFAULT NULL,
    last_node_id VARCHAR(255) DEFAULT NULL,
    waiting_for_reply BOOLEAN DEFAULT false,
    balas VARCHAR(255) DEFAULT NULL,
    human INTEGER DEFAULT 0,
    keywordiklan VARCHAR(255) DEFAULT NULL,
    marketer VARCHAR(255) DEFAULT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    update_today TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_execution_id ON ai_whatsapp(execution_id);
CREATE INDEX IF NOT EXISTS idx_flow_id ON ai_whatsapp(flow_id);
CREATE INDEX IF NOT EXISTS idx_current_node_id ON ai_whatsapp(current_node_id);
CREATE INDEX IF NOT EXISTS idx_id_device ON ai_whatsapp(id_device);
CREATE INDEX IF NOT EXISTS idx_prospect_num ON ai_whatsapp(prospect_num);

CREATE TRIGGER update_ai_whatsapp_updated_at 
    BEFORE UPDATE ON ai_whatsapp 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
`

// WasapBot table for WasapBot Exama flow process
const createWasapBotTable = `
CREATE TABLE IF NOT EXISTS wasapBot (
  id_prospect SERIAL PRIMARY KEY,
  flow_reference VARCHAR(255) DEFAULT NULL,
  execution_id VARCHAR(255) DEFAULT NULL,
  execution_status VARCHAR(20) DEFAULT NULL CHECK (execution_status IN ('active','completed','failed')),
  flow_id VARCHAR(255) DEFAULT NULL,
  current_node_id VARCHAR(255) DEFAULT NULL,
  last_node_id VARCHAR(255) DEFAULT NULL,
  waiting_for_reply BOOLEAN DEFAULT false,
  marketer_id VARCHAR(100) DEFAULT NULL,
  prospect_num VARCHAR(100) DEFAULT NULL,
  niche VARCHAR(300) DEFAULT NULL,
  instance VARCHAR(255) DEFAULT NULL,
  peringkat_sekolah VARCHAR(100) DEFAULT NULL,
  alamat VARCHAR(100) DEFAULT NULL,
  nama VARCHAR(100) DEFAULT NULL,
  pakej VARCHAR(100) DEFAULT NULL,
  no_fon VARCHAR(20) DEFAULT NULL,
  cara_bayaran VARCHAR(100) DEFAULT NULL,
  tarikh_gaji VARCHAR(20) DEFAULT NULL,
  stage VARCHAR(200) DEFAULT NULL,
  temp_stage VARCHAR(200) DEFAULT NULL,
  conv_start VARCHAR(200) DEFAULT NULL,
  conv_last TEXT,
  date_start VARCHAR(50) DEFAULT NULL,
  date_last VARCHAR(50) DEFAULT NULL,
  status VARCHAR(200) DEFAULT 'Prospek',
  staff_cls VARCHAR(200) DEFAULT NULL,
  umur VARCHAR(200) DEFAULT NULL,
  kerja VARCHAR(200) DEFAULT NULL,
  sijil VARCHAR(200) DEFAULT NULL,
  user_input TEXT,
  alasan VARCHAR(200) DEFAULT NULL,
  nota VARCHAR(200) DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_wasapbot_prospect_num ON wasapBot(prospect_num);
CREATE INDEX IF NOT EXISTS idx_wasapbot_flow_id ON wasapBot(flow_id);
CREATE INDEX IF NOT EXISTS idx_wasapbot_execution_id ON wasapBot(execution_id);
CREATE INDEX IF NOT EXISTS idx_wasapbot_instance ON wasapBot(instance);
`

// Conversation log table for storing all AI conversation history
const createConversationLogTable = `
CREATE TABLE IF NOT EXISTS conversation_log (
    id VARCHAR(255) PRIMARY KEY,
    prospect_num VARCHAR(20) NOT NULL,
    sender VARCHAR(10) NOT NULL CHECK (sender IN ('user', 'bot', 'staff')),
    message TEXT NOT NULL,
    message_type VARCHAR(10) DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'document', 'audio', 'video')),
    stage VARCHAR(255),
    ai_response JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_conversation_log_prospect_num ON conversation_log(prospect_num);
CREATE INDEX IF NOT EXISTS idx_conversation_log_sender ON conversation_log(sender);
CREATE INDEX IF NOT EXISTS idx_conversation_log_created_at ON conversation_log(created_at);
CREATE INDEX IF NOT EXISTS idx_conversation_log_stage ON conversation_log(stage);
`

// Orders table for Billplz payment integration
const createOrdersTable = `
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    amount DECIMAL(10,2) NOT NULL,
    collection_id VARCHAR(255),
    status VARCHAR(20) DEFAULT 'Pending' CHECK (status IN ('Pending', 'Processing', 'Success', 'Failed')),
    bill_id VARCHAR(255),
    url TEXT,
    product VARCHAR(255) NOT NULL,
    method VARCHAR(50) DEFAULT 'billplz',
    user_id CHAR(36),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_orders_bill_id ON orders(bill_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);

CREATE TRIGGER update_orders_updated_at 
    BEFORE UPDATE ON orders 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
`

const createAIWhatsappSessionTable = `
CREATE TABLE IF NOT EXISTS ai_whatsapp_session (
    id_sessionX SERIAL PRIMARY KEY,
    phone_number VARCHAR(255) NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    locked_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    timestamp VARCHAR(255) NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_ai_whatsapp_session ON ai_whatsapp_session(phone_number, device_id);
CREATE INDEX IF NOT EXISTS idx_ai_whatsapp_session_device ON ai_whatsapp_session(device_id);
CREATE INDEX IF NOT EXISTS idx_ai_session_locked ON ai_whatsapp_session(locked_at);
CREATE INDEX IF NOT EXISTS idx_ai_session_activity ON ai_whatsapp_session(last_activity);
`

const createWasapBotSessionTable = `
CREATE TABLE IF NOT EXISTS wasapBot_session (
    id_sessionY SERIAL PRIMARY KEY,
    id_prospect VARCHAR(255) NOT NULL,
    id_device VARCHAR(255) NOT NULL,
    timestamp VARCHAR(255) NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_wasapbot_session ON wasapBot_session(id_prospect, id_device);
CREATE INDEX IF NOT EXISTS idx_wasapbot_session_device ON wasapBot_session(id_device);
`

// addMissingColumnsToFlowsTable adds missing columns to the flows table
func addMissingColumnsToFlowsTable(db *sql.DB) error {
	columns := []struct {
		name       string
		definition string
	}{
		{"niche", "TEXT"},
		{"id_device", "VARCHAR(255)"},
	}

	for _, col := range columns {
		// Check if column exists (PostgreSQL syntax)
		var count int
		err := db.QueryRow(`
			SELECT COUNT(*) 
			FROM information_schema.columns 
			WHERE table_schema = 'public' 
			AND table_name = 'chatbot_flows' 
			AND column_name = $1
		`, col.name).Scan(&count)

		if err != nil {
			return fmt.Errorf("failed to check column %s: %w", col.name, err)
		}

		if count == 0 {
			// Column doesn't exist, add it
			query := fmt.Sprintf("ALTER TABLE chatbot_flows ADD COLUMN %s %s", col.name, col.definition)
			if _, err := db.Exec(query); err != nil {
				return fmt.Errorf("failed to add column %s: %w", col.name, err)
			}
			logrus.WithField("column", col.name).Info("Added missing column")
		} else {
			logrus.WithField("column", col.name).Debug("Column already exists")
		}
	}
	return nil
}

// updateProviderRvsbWasapToWaha updates provider values from 'rvsb_wasap' to 'waha'
func updateProviderRvsbWasapToWaha(db *sql.DB) error {
	// Update existing records that have 'rvsb_wasap' provider to 'waha'
	result, err := db.Exec("UPDATE device_setting SET provider = 'waha' WHERE provider = 'rvsb_wasap'")
	if err != nil {
		return fmt.Errorf("failed to update provider values: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected > 0 {
		logrus.WithField("rows_updated", rowsAffected).Info("Updated provider values from 'rvsb_wasap' to 'waha'")
	} else {
		logrus.Debug("No records found with 'rvsb_wasap' provider to update")
	}

	return nil
}

// updateProviderEnum updates the provider constraint to include 'waha' and remove 'rvsb_wasap'
func updateProviderEnum(db *sql.DB) error {
	logrus.Info("🔧 Checking provider constraint in device_setting table")

	// Check if table exists first (PostgreSQL syntax)
	var tableExists int
	err := db.QueryRow(`
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name = 'device_setting'
	`).Scan(&tableExists)

	if err != nil {
		logrus.WithError(err).Warn("Failed to check if device_setting table exists, skipping provider constraint update")
		return nil
	}

	if tableExists == 0 {
		logrus.Info("device_setting table doesn't exist yet, skipping provider constraint update")
		return nil
	}

	// For PostgreSQL, we need to drop and recreate the constraint
	logrus.Info("🔧 Updating provider constraint to include 'waha' and remove 'rvsb_wasap'")
	
	// Drop existing constraint if it exists
	_, err = db.Exec("ALTER TABLE device_setting DROP CONSTRAINT IF EXISTS device_setting_provider_check")
	if err != nil {
		logrus.WithError(err).Warn("Failed to drop existing provider constraint")
	}

	// Add new constraint
	_, err = db.Exec("ALTER TABLE device_setting ADD CONSTRAINT device_setting_provider_check CHECK (provider IN ('whacenter', 'wablas', 'waha'))")
	if err != nil {
		logrus.WithError(err).Error("❌ Failed to update provider constraint - this will cause WAHA provider issues")
		return fmt.Errorf("failed to update provider constraint: %w", err)
	}

	logrus.Info("✅ Successfully updated provider constraint to support WAHA provider")
	return nil
}

// removeDeprecatedColumnsFromFlowsTable removes deprecated columns from the flows table
func removeDeprecatedColumnsFromFlowsTable(db *sql.DB) error {
	columns := []string{
		"global_instance",
		"global_open_router_key",
	}

	for _, col := range columns {
		// Check if column exists (PostgreSQL syntax)
		var count int
		err := db.QueryRow(`
			SELECT COUNT(*) 
			FROM information_schema.columns 
			WHERE table_schema = 'public' 
			AND table_name = 'chatbot_flows' 
			AND column_name = $1
		`, col).Scan(&count)

		if err != nil {
			return fmt.Errorf("failed to check column %s: %w", col, err)
		}

		if count > 0 {
			// Column exists, drop it
			query := fmt.Sprintf("ALTER TABLE chatbot_flows DROP COLUMN %s", col)
			if _, err := db.Exec(query); err != nil {
				return fmt.Errorf("failed to drop column %s: %w", col, err)
			}
			logrus.WithField("column", col).Info("Removed deprecated column")
		} else {
			logrus.WithField("column", col).Debug("Deprecated column does not exist")
		}
	}
	return nil
}

// addMissingColumnsToDeviceSettingsTable adds missing columns to the device settings table
func addMissingColumnsToDeviceSettingsTable(db *sql.DB) error {
	// Define columns that should exist
	columns := []struct {
		name       string
		definition string
	}{
		{"phone_number", "VARCHAR(20)"},
		{"instance", "TEXT"},
		{"user_id", "CHAR(36)"},
	}

	for _, col := range columns {
		// Check if column exists (PostgreSQL syntax)
		var count int
		err := db.QueryRow(`
			SELECT COUNT(*) 
			FROM information_schema.columns 
			WHERE table_schema = 'public' 
			AND table_name = 'device_setting' 
			AND column_name = $1
		`, col.name).Scan(&count)

		if err != nil {
			return fmt.Errorf("failed to check column %s: %w", col.name, err)
		}

		if count == 0 {
			// Column doesn't exist, add it
			query := fmt.Sprintf("ALTER TABLE device_setting ADD COLUMN %s %s", col.name, col.definition)
			if _, err := db.Exec(query); err != nil {
				return fmt.Errorf("failed to add column %s: %w", col.name, err)
			}
			logrus.WithField("column", col.name).Info("Added missing column")
		} else {
			logrus.WithField("column", col.name).Debug("Column already exists")
		}
	}
	return nil
}

// dropDeprecatedBillingColumns drops deprecated billing columns from orders table
func dropDeprecatedBillingColumns(db *sql.DB) error {
	columns := []string{
		"customer_email",
		"customer_name",
		"billing_phone",
		"billing_address",
		"billing_city",
		"billing_state",
		"billing_postcode",
	}

	for _, col := range columns {
		// Check if column exists
		var count int
		err := db.QueryRow(`
			SELECT COUNT(*) 
			FROM INFORMATION_SCHEMA.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() 
			AND TABLE_NAME = 'orders' 
			AND COLUMN_NAME = ?
		`, col).Scan(&count)

		if err != nil {
			return fmt.Errorf("failed to check column %s: %w", col, err)
		}

		if count > 0 {
			// Column exists, drop it
			query := fmt.Sprintf("ALTER TABLE orders DROP COLUMN %s", col)
			if _, err := db.Exec(query); err != nil {
				return fmt.Errorf("failed to drop column %s: %w", col, err)
			}
			logrus.WithField("column", col).Info("Dropped deprecated billing column from orders")
		} else {
			logrus.WithField("column", col).Debug("Deprecated billing column does not exist in orders")
		}
	}
	return nil
}

// convertUserIDToChar36 converts user_id column from INT to CHAR(36) in orders table
func convertUserIDToChar36(db *sql.DB) error {
	// Check current data type of user_id column
	var dataType string
	err := db.QueryRow(`
		SELECT DATA_TYPE 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'orders' 
		AND COLUMN_NAME = 'user_id'
	`).Scan(&dataType)

	if err != nil {
		return fmt.Errorf("failed to check user_id column type: %w", err)
	}

	// If already CHAR, skip
	if dataType == "char" {
		logrus.Info("user_id column is already CHAR(36) in orders")
		return nil
	}

	// Convert INT to CHAR(36) - first set existing data to NULL or empty since INT can't convert to UUID
	_, err = db.Exec("UPDATE orders SET user_id = NULL WHERE user_id IS NOT NULL")
	if err != nil {
		logrus.WithError(err).Warn("Failed to clear user_id values before conversion")
	}

	// Alter column type
	_, err = db.Exec("ALTER TABLE orders MODIFY COLUMN user_id CHAR(36)")
	if err != nil {
		return fmt.Errorf("failed to convert user_id to CHAR(36): %w", err)
	}

	logrus.Info("Successfully converted user_id column from INT to CHAR(36) in orders")
	return nil
}

// convertDeviceUserIDToChar36 converts user_id column from INT to CHAR(36) in device_setting table
func convertDeviceUserIDToChar36(db *sql.DB) error {
	// Check current data type of user_id column
	var dataType string
	err := db.QueryRow(`
		SELECT DATA_TYPE 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'device_setting' 
		AND COLUMN_NAME = 'user_id'
	`).Scan(&dataType)

	if err != nil {
		return fmt.Errorf("failed to check user_id column type: %w", err)
	}

	// If already CHAR, skip
	if dataType == "char" {
		logrus.Info("user_id column is already CHAR(36) in device_setting")
		return nil
	}

	logrus.Warn("⚠️  CRITICAL: user_id column needs conversion from INT to CHAR(36)")
	logrus.Warn("⚠️  This will set all user_id values to NULL!")
	logrus.Warn("⚠️  You MUST run the fix script: scripts/fix_device_user_links.sql")
	logrus.Warn("⚠️  Or manually re-link devices to users after migration")

	// Convert INT to CHAR(36) - WARNING: This sets existing data to NULL since INT can't convert to UUID
	// Users will need to run fix script or manually re-link devices
	_, err = db.Exec("UPDATE device_setting SET user_id = NULL WHERE user_id IS NOT NULL")
	if err != nil {
		logrus.WithError(err).Warn("Failed to clear user_id values before conversion in device_setting")
	}

	// Alter column type
	_, err = db.Exec("ALTER TABLE device_setting MODIFY COLUMN user_id CHAR(36)")
	if err != nil {
		return fmt.Errorf("failed to convert user_id to CHAR(36) in device_setting: %w", err)
	}

	// Add index if it doesn't exist
	_, err = db.Exec("ALTER TABLE device_setting ADD INDEX IF NOT EXISTS idx_user_id (user_id)")
	if err != nil {
		logrus.WithError(err).Warn("Failed to add index to user_id column")
	}

	logrus.Info("✅ Converted user_id column from INT to CHAR(36) in device_setting")
	logrus.Warn("⚠️  ACTION REQUIRED: Run scripts/fix_device_user_links.sql to restore device-user links")
	return nil
}

// makeDeviceIDNullable makes device_id column nullable to allow manual device creation
func makeDeviceIDNullable(db *sql.DB) error {
	logrus.Info("🔧 Checking device_id column nullability in device_setting table")

	// Check if table exists first
	var tableExists int
	err := db.QueryRow(`
		SELECT COUNT(*) 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'device_setting'
	`).Scan(&tableExists)

	if err != nil {
		logrus.WithError(err).Warn("Failed to check if device_setting table exists, skipping migration")
		return nil // Don't fail the entire migration for this
	}

	if tableExists == 0 {
		logrus.Info("device_setting table doesn't exist yet, skipping device_id nullable migration")
		return nil
	}

	// Check if device_id column allows NULL
	var isNullable string
	err = db.QueryRow(`
		SELECT IS_NULLABLE 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'device_setting' 
		AND COLUMN_NAME = 'device_id'
	`).Scan(&isNullable)

	if err != nil {
		logrus.WithError(err).Warn("Failed to check device_id column nullability, attempting to alter anyway")
		// Continue and try to alter, might be a column that doesn't exist yet
	} else {
		// If already nullable, skip
		if isNullable == "YES" {
			logrus.Info("✅ device_id column is already nullable in device_setting")
			return nil
		}
		logrus.WithField("is_nullable", isNullable).Info("device_id column current nullability status")
	}

	// Make column nullable - this is the critical fix
	logrus.Info("🔧 Altering device_id column to allow NULL values")
	_, err = db.Exec("ALTER TABLE device_setting MODIFY COLUMN device_id VARCHAR(255) NULL")
	if err != nil {
		logrus.WithError(err).Error("❌ Failed to make device_id nullable - this will cause WAHA provider issues")
		return fmt.Errorf("failed to make device_id nullable: %w", err)
	}

	logrus.Info("✅ Successfully made device_id column nullable in device_setting for manual device creation")

	// Verify the change was applied
	err = db.QueryRow(`
		SELECT IS_NULLABLE 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'device_setting' 
		AND COLUMN_NAME = 'device_id'
	`).Scan(&isNullable)

	if err == nil {
		logrus.WithField("is_nullable_after", isNullable).Info("✅ Verified device_id column nullability after migration")
	}

	return nil
}
