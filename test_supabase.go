package main

import (
	"fmt"
	"os"
	"net"
	"strings"
	"database/sql"
	
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// resolveIPv4 resolves a hostname to its IPv4 address to avoid IPv6 issues
func resolveIPv4(hostname string) (string, error) {
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return "", err
	}
	
	// Find the first IPv4 address
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			logrus.WithFields(logrus.Fields{
				"hostname": hostname,
				"ipv4":     ipv4.String(),
			}).Info("Resolved hostname to IPv4")
			return ipv4.String(), nil
		}
	}
	
	return "", fmt.Errorf("no IPv4 address found for hostname: %s", hostname)
}

func main() {
	// Configure logging
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	
	// Get Supabase credentials from environment variables
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseDBPassword := os.Getenv("SUPABASE_DB_PASSWORD")
	
	if supabaseURL == "" || supabaseDBPassword == "" {
		logrus.Fatal("SUPABASE_URL and SUPABASE_DB_PASSWORD are required")
	}
	
	logrus.Info("🚀 Testing Supabase PostgreSQL connection")
	
	// Extract project reference from Supabase URL
	url := strings.TrimPrefix(supabaseURL, "https://")
	url = strings.TrimPrefix(url, "http://")
	
	parts := strings.Split(url, ".")
	if len(parts) == 0 {
		logrus.Fatal("Invalid Supabase URL format")
	}
	
	projectRef := parts[0]
	logrus.WithField("project_ref", projectRef).Info("Extracted project reference")
	
	// Resolve hostname to IPv4 to avoid IPv6 connection issues
	hostname := fmt.Sprintf("db.%s.supabase.co", projectRef)
	ipv4Address, err := resolveIPv4(hostname)
	
	var connStr string
	if err != nil {
		// Fallback to hostname if IPv4 resolution fails
		logrus.WithError(err).Warn("Failed to resolve IPv4, using hostname")
		connStr = fmt.Sprintf("host=%s port=5432 user=postgres dbname=postgres sslmode=require connect_timeout=30 password=%s",
			hostname, supabaseDBPassword)
	} else {
		// Use IPv4 address directly to force IPv4 connection
		logrus.WithField("ipv4", ipv4Address).Info("Using IPv4 address for Railway compatibility")
		connStr = fmt.Sprintf("host=%s port=5432 user=postgres dbname=postgres sslmode=require connect_timeout=30 password=%s",
			ipv4Address, supabaseDBPassword)
	}
	
	logrus.WithField("connection_string", strings.ReplaceAll(connStr, supabaseDBPassword, "***")).Debug("Using connection string")
	
	// Open PostgreSQL connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to open Supabase PostgreSQL connection")
	}
	defer db.Close()
	
	// Test the connection
	if err := db.Ping(); err != nil {
		logrus.WithError(err).Fatal("Failed to ping Supabase PostgreSQL database")
	}
	
	logrus.Info("✅ Supabase PostgreSQL database connection established successfully")
	
	// Try a simple query
	var result int
	err = db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to execute test query")
	}
	
	logrus.WithField("result", result).Info("✅ Test query executed successfully")
}