package database

import (
	"context"
	"database/sql"

	"github.com/sirupsen/logrus"
)

// RESTAdapter wraps SupabaseRestClient to provide sql.DB-like interface
// This allows gradual migration from SQL to REST API
type RESTAdapter struct {
	RestClient *SupabaseRestClient
	// Keep a dummy sql.DB for compatibility (won't be used)
	DummyDB *sql.DB
}

// NewRESTAdapter creates an adapter that uses REST API but provides sql.DB interface
func NewRESTAdapter(restClient *SupabaseRestClient) *RESTAdapter {
	logrus.Info("🔄 Creating REST API adapter for backward compatibility")
	return &RESTAdapter{
		RestClient: restClient,
		DummyDB:    nil, // Will be nil to force REST API usage
	}
}

// Query executes a query that returns rows (SELECT)
func (a *RESTAdapter) Query(table string, filters map[string]string, result interface{}) error {
	ctx := context.Background()
	return a.RestClient.Query(ctx, table, filters, result)
}

// Insert adds a new record
func (a *RESTAdapter) Insert(table string, data interface{}, result interface{}) error {
	ctx := context.Background()
	return a.RestClient.Insert(ctx, table, data, result)
}

// Update modifies existing records
func (a *RESTAdapter) Update(table string, filters map[string]string, data interface{}) error {
	ctx := context.Background()
	return a.RestClient.Update(ctx, table, filters, data)
}

// Delete removes records
func (a *RESTAdapter) Delete(table string, filters map[string]string) error {
	ctx := context.Background()
	return a.RestClient.Delete(ctx, table, filters)
}

// Ping checks if the connection is alive
func (a *RESTAdapter) Ping() error {
	ctx := context.Background()
	return a.RestClient.Ping(ctx)
}

// GetDB returns the dummy sql.DB (returns nil to force REST API usage)
func (a *RESTAdapter) GetDB() *sql.DB {
	if a.DummyDB != nil {
		logrus.Warn("⚠️ Attempting to use direct SQL - this will not work with REST API")
		logrus.Warn("💡 Please migrate this code to use REST API methods")
	}
	return a.DummyDB
}

// Helper function to execute raw SQL queries via Supabase RPC
// This is for complex queries that can't be done via simple REST API
func (a *RESTAdapter) ExecuteRPC(functionName string, params map[string]interface{}, result interface{}) error {
	ctx := context.Background()

	// Call Supabase stored procedure via REST API
	// Use the Insert method to call RPC (POST request)
	return a.RestClient.Insert(ctx, "rpc/"+functionName, params, result)
}
