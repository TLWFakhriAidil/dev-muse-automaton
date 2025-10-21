package database

import (
	"context"
	"fmt"

	"nodepath-chat/internal/config"
	"github.com/sirupsen/logrus"
)

// SupabaseSDK mimics the @supabase/supabase-js SDK pattern
// This provides a familiar interface similar to TypeScript Supabase client
type SupabaseSDK struct {
	restClient *SupabaseRestClient
	config     *config.Config
}

// NewSupabaseSDK creates a new Supabase SDK instance (like createClient in JS)
func NewSupabaseSDK(cfg *config.Config) (*SupabaseSDK, error) {
	restClient, err := NewSupabaseRestClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Supabase SDK: %w", err)
	}

	logrus.Info("✅ Supabase SDK initialized (JavaScript-like pattern)")

	return &SupabaseSDK{
		restClient: restClient,
		config:     cfg,
	}, nil
}

// From returns a table query builder (mimics supabase.from('table_name'))
func (s *SupabaseSDK) From(table string) *TableQuery {
	return &TableQuery{
		table:      table,
		restClient: s.restClient,
		filters:    make(map[string]string),
	}
}

// TableQuery mimics the Supabase JavaScript query builder
type TableQuery struct {
	table      string
	restClient *SupabaseRestClient
	filters    map[string]string
	selectCols string
}

// Select specifies columns to return (mimics .select('col1, col2'))
func (tq *TableQuery) Select(columns string) *TableQuery {
	tq.selectCols = columns
	return tq
}

// Eq adds an equality filter (mimics .eq('column', value))
func (tq *TableQuery) Eq(column string, value interface{}) *TableQuery {
	tq.filters[column] = fmt.Sprintf("%v", value)
	return tq
}

// Single executes the query and returns a single row (mimics .single())
func (tq *TableQuery) Single(ctx context.Context, result interface{}) error {
	err := tq.restClient.Query(ctx, tq.table, tq.filters, result)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	return nil
}

// Execute runs the query and returns multiple rows
func (tq *TableQuery) Execute(ctx context.Context, result interface{}) error {
	return tq.restClient.Query(ctx, tq.table, tq.filters, result)
}

// Insert adds a new record (mimics .insert({data}))
func (tq *TableQuery) Insert(ctx context.Context, data interface{}, result interface{}) error {
	return tq.restClient.Insert(ctx, tq.table, data, result)
}

// Update modifies records (mimics .update({data}))
func (tq *TableQuery) Update(ctx context.Context, data interface{}) error {
	return tq.restClient.Update(ctx, tq.table, tq.filters, data)
}

// Delete removes records (mimics .delete())
func (tq *TableQuery) Delete(ctx context.Context) error {
	return tq.restClient.Delete(ctx, tq.table, tq.filters)
}

// Example usage (mimics JavaScript pattern):
//
// // JavaScript:
// const { data, error } = await supabase
//   .from('users')
//   .select('id, username')
//   .eq('id', userId)
//   .single();
//
// // Go equivalent:
// var user User
// err := supabase.From("users").
//   Select("id, username").
//   Eq("id", userId).
//   Single(ctx, &user)
