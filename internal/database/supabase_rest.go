package database

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"nodepath-chat/internal/config"
	"github.com/sirupsen/logrus"
)

// SupabaseRestClient provides HTTP REST API access to Supabase
// This bypasses PostgreSQL connection and uses PostgREST instead
type SupabaseRestClient struct {
	BaseURL    string
	APIKey     string
	ServiceKey string
	HTTPClient *http.Client
}

// NewSupabaseRestClient creates a new REST API client for Supabase
func NewSupabaseRestClient(cfg *config.Config) (*SupabaseRestClient, error) {
	if cfg.SupabaseURL == "" || cfg.SupabaseAnonKey == "" {
		return nil, fmt.Errorf("SUPABASE_URL and SUPABASE_ANON_KEY are required for REST API")
	}

	logrus.Info("🚀 Initializing Supabase REST API client (HTTP - no IPv6 issues)")

	return &SupabaseRestClient{
		BaseURL:    cfg.SupabaseURL + "/rest/v1",
		APIKey:     cfg.SupabaseAnonKey,
		ServiceKey: cfg.SupabaseServiceKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Query executes a SELECT query via REST API
func (c *SupabaseRestClient) Query(ctx context.Context, table string, filters map[string]string, result interface{}) error {
	url := fmt.Sprintf("%s/%s", c.BaseURL, table)

	// Add query parameters
	if len(filters) > 0 {
		url += "?"
		first := true
		for key, value := range filters {
			if !first {
				url += "&"
			}
			url += fmt.Sprintf("%s=eq.%s", key, value)
			first = false
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add required headers
	req.Header.Set("apikey", c.APIKey)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("query failed with status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// Insert adds a new record via REST API
func (c *SupabaseRestClient) Insert(ctx context.Context, table string, data interface{}, result interface{}) error {
	url := fmt.Sprintf("%s/%s", c.BaseURL, table)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("apikey", c.APIKey)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("insert failed with status %d: %s", resp.StatusCode, string(body))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// Update modifies existing records via REST API
func (c *SupabaseRestClient) Update(ctx context.Context, table string, filters map[string]string, data interface{}) error {
	url := fmt.Sprintf("%s/%s", c.BaseURL, table)

	// Add query parameters
	if len(filters) > 0 {
		url += "?"
		first := true
		for key, value := range filters {
			if !first {
				url += "&"
			}
			url += fmt.Sprintf("%s=eq.%s", key, value)
			first = false
		}
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("apikey", c.APIKey)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Delete removes records via REST API
func (c *SupabaseRestClient) Delete(ctx context.Context, table string, filters map[string]string) error {
	url := fmt.Sprintf("%s/%s", c.BaseURL, table)

	// Add query parameters
	if len(filters) > 0 {
		url += "?"
		first := true
		for key, value := range filters {
			if !first {
				url += "&"
			}
			url += fmt.Sprintf("%s=eq.%s", key, value)
			first = false
		}
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("apikey", c.APIKey)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Ping tests the REST API connection
func (c *SupabaseRestClient) Ping(ctx context.Context) error {
	// Try to query a system table
	url := fmt.Sprintf("%s/chatbot_flows?limit=1", c.BaseURL)

	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create ping request: %w", err)
	}

	req.Header.Set("apikey", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("ping failed with status %d", resp.StatusCode)
	}

	logrus.Info("✅ Supabase REST API connection successful")
	return nil
}

// ExecuteRPC executes a Supabase stored procedure/function via RPC
func (c *SupabaseRestClient) ExecuteRPC(ctx context.Context, functionName string, params map[string]interface{}, result interface{}) error {
	url := fmt.Sprintf("%s/rpc/%s", c.BaseURL, functionName)

	jsonData, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal RPC params: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create RPC request: %w", err)
	}

	// Add required headers
	req.Header.Set("apikey", c.APIKey)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("RPC request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("RPC failed with status %d: %s", resp.StatusCode, string(body))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode RPC response: %w", err)
		}
	}

	return nil
}
