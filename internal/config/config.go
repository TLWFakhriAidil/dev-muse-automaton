package config

import "os"

type Config struct {
	Port                   int
	SupabaseURL            string
	SupabaseAnonKey        string
	SupabaseServiceRoleKey string
}

func Load() *Config {
	port := 8080
	if p := os.Getenv("PORT"); p != "" {
		// Parse port if needed
	}

	return &Config{
		Port:                   port,
		SupabaseURL:            getEnv("SUPABASE_URL", "https://bjnjucwpwdzgsnqmpmff.supabase.co"),
		SupabaseAnonKey:        getEnv("SUPABASE_ANON_KEY", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImJqbmp1Y3dwd2R6Z3NucW1wbWZmIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjA0OTk1MzksImV4cCI6MjA3NjA3NTUzOX0.vw1rOUqYWFkPNDwTdEgIfsCO9pyvTsFKaXHq3RcRTNU"),
		SupabaseServiceRoleKey: os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
