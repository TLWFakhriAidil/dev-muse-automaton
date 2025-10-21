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
	orderBy    string
	limitVal   int
	offsetVal  int
}

// Select specifies columns to return (mimics .select('col1, col2'))
func (tq *TableQuery) Select(columns string) *TableQuery {
	tq.selectCols = columns
	return tq
}

// Eq adds an equality filter (mimics .eq('column', value))
func (tq *TableQuery) Eq(column string, value interface{}) *TableQuery {
	tq.filters[column] = fmt.Sprintf("eq.%v", value)
	return tq
}

// Neq adds a not-equal filter (mimics .neq('column', value))
func (tq *TableQuery) Neq(column string, value interface{}) *TableQuery {
	tq.filters[column] = fmt.Sprintf("neq.%v", value)
	return tq
}

// Gt adds a greater-than filter (mimics .gt('column', value))
func (tq *TableQuery) Gt(column string, value interface{}) *TableQuery {
	tq.filters[column] = fmt.Sprintf("gt.%v", value)
	return tq
}

// Gte adds a greater-than-or-equal filter (mimics .gte('column', value))
func (tq *TableQuery) Gte(column string, value interface{}) *TableQuery {
	tq.filters[column] = fmt.Sprintf("gte.%v", value)
	return tq
}

// Lt adds a less-than filter (mimics .lt('column', value))
func (tq *TableQuery) Lt(column string, value interface{}) *TableQuery {
	tq.filters[column] = fmt.Sprintf("lt.%v", value)
	return tq
}

// Lte adds a less-than-or-equal filter (mimics .lte('column', value))
func (tq *TableQuery) Lte(column string, value interface{}) *TableQuery {
	tq.filters[column] = fmt.Sprintf("lte.%v", value)
	return tq
}

// Like adds a LIKE filter (mimics .like('column', pattern))
func (tq *TableQuery) Like(column string, pattern string) *TableQuery {
	tq.filters[column] = fmt.Sprintf("like.%s", pattern)
	return tq
}

// Ilike adds a case-insensitive LIKE filter (mimics .ilike('column', pattern))
func (tq *TableQuery) Ilike(column string, pattern string) *TableQuery {
	tq.filters[column] = fmt.Sprintf("ilike.%s", pattern)
	return tq
}

// In adds an IN filter (mimics .in('column', [values]))
func (tq *TableQuery) In(column string, values []interface{}) *TableQuery {
	var stringVals []string
	for _, v := range values {
		stringVals = append(stringVals, fmt.Sprintf("%v", v))
	}
	// PostgREST syntax: column=in.(val1,val2,val3)
	tq.filters[column] = fmt.Sprintf("in.(%s)", join(stringVals, ","))
	return tq
}

// Is adds an IS filter for NULL checks (mimics .is('column', null))
func (tq *TableQuery) Is(column string, value interface{}) *TableQuery {
	tq.filters[column] = fmt.Sprintf("is.%v", value)
	return tq
}

// IsNull adds an IS NULL filter (mimics .is('column', null))
func (tq *TableQuery) IsNull(column string) *TableQuery {
	tq.filters[column] = "is.null"
	return tq
}

// IsNotNull adds an IS NOT NULL filter (mimics .not('column', 'is', null))
func (tq *TableQuery) IsNotNull(column string) *TableQuery {
	tq.filters[column] = "not.is.null"
	return tq
}

// Not adds a NOT filter (mimics .not('column', 'eq', value))
func (tq *TableQuery) Not(column string, operator string, value interface{}) *TableQuery {
	tq.filters[column] = fmt.Sprintf("not.%s.%v", operator, value)
	return tq
}

// Or adds an OR filter for multiple conditions on same column
// Note: For complex OR across different columns, use OrFilter with raw PostgREST syntax
func (tq *TableQuery) Or(conditions string) *TableQuery {
	tq.filters["or"] = fmt.Sprintf("(%s)", conditions)
	return tq
}

// And adds additional AND conditions (default behavior, but explicit for clarity)
func (tq *TableQuery) And(column string, operator string, value interface{}) *TableQuery {
	tq.filters[column] = fmt.Sprintf("%s.%v", operator, value)
	return tq
}

// Order specifies the order (mimics .order('column', { ascending: false }))
func (tq *TableQuery) Order(column string, ascending bool) *TableQuery {
	if ascending {
		tq.orderBy = column + ".asc"
	} else {
		tq.orderBy = column + ".desc"
	}
	return tq
}

// Limit specifies the maximum number of rows (mimics .limit(n))
func (tq *TableQuery) Limit(limit int) *TableQuery {
	tq.limitVal = limit
	return tq
}

// Offset specifies the number of rows to skip (mimics .offset(n))
func (tq *TableQuery) Offset(offset int) *TableQuery {
	tq.offsetVal = offset
	return tq
}

// Range specifies the range of rows (mimics .range(from, to))
func (tq *TableQuery) Range(from, to int) *TableQuery {
	tq.offsetVal = from
	tq.limitVal = to - from + 1
	return tq
}

// Single executes the query and returns a single row (mimics .single())
func (tq *TableQuery) Single(ctx context.Context, result interface{}) error {
	// Add limit=1 for single queries
	originalLimit := tq.limitVal
	tq.limitVal = 1
	defer func() { tq.limitVal = originalLimit }()

	err := tq.executeQuery(ctx, result)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	return nil
}

// Execute runs the query and returns multiple rows
func (tq *TableQuery) Execute(ctx context.Context, result interface{}) error {
	return tq.executeQuery(ctx, result)
}

// executeQuery is the internal method that builds and executes the query
func (tq *TableQuery) executeQuery(ctx context.Context, result interface{}) error {
	// Build the complete filter map including order, limit, offset
	filters := make(map[string]string)
	for k, v := range tq.filters {
		filters[k] = v
	}

	// Add order, limit, offset if specified
	if tq.orderBy != "" {
		filters["order"] = tq.orderBy
	}
	if tq.limitVal > 0 {
		filters["limit"] = fmt.Sprintf("%d", tq.limitVal)
	}
	if tq.offsetVal > 0 {
		filters["offset"] = fmt.Sprintf("%d", tq.offsetVal)
	}
	if tq.selectCols != "" {
		filters["select"] = tq.selectCols
	}

	return tq.restClient.Query(ctx, tq.table, filters, result)
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

// Upsert performs an INSERT or UPDATE (mimics .upsert({data}))
// Uses the Prefer: resolution=merge-duplicates header
func (tq *TableQuery) Upsert(ctx context.Context, data interface{}, result interface{}) error {
	// For now, use Insert - can be enhanced to use proper upsert with headers
	return tq.restClient.Insert(ctx, tq.table, data, result)
}

// RPC executes a stored procedure/function (mimics supabase.rpc('function_name', params))
func (s *SupabaseSDK) RPC(ctx context.Context, functionName string, params map[string]interface{}, result interface{}) error {
	return s.restClient.ExecuteRPC(ctx, functionName, params, result)
}

// Helper function to join strings
func join(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
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
