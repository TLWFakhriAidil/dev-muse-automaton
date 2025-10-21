package repository

import (
	"context"
	"fmt"
	"time"

	"nodepath-chat/internal/database"
	"nodepath-chat/internal/models"

	"github.com/sirupsen/logrus"
)

// deviceSettingsRepositorySupabase implements DeviceSettingsRepository using Supabase SDK
// This is the NEW implementation that works with Railway (no IPv6 issues)
type deviceSettingsRepositorySupabase struct {
	supabase *database.SupabaseSDK
}

// NewDeviceSettingsRepositorySupabase creates a Supabase-based device settings repository
func NewDeviceSettingsRepositorySupabase(supabase *database.SupabaseSDK) DeviceSettingsRepository {
	return &deviceSettingsRepositorySupabase{
		supabase: supabase,
	}
}

// CreateDeviceSettings creates a new device settings record using Supabase REST API
func (r *deviceSettingsRepositorySupabase) CreateDeviceSettings(settings *models.DeviceSettings) error {
	ctx := context.Background()
	settings.CreatedAt = time.Now()
	settings.UpdatedAt = time.Now()

	var result models.DeviceSettings
	err := r.supabase.From("device_setting").Insert(ctx, settings, &result)
	if err != nil {
		logrus.WithError(err).Error("Failed to create device settings")
		return fmt.Errorf("failed to create device settings: %w", err)
	}

	logrus.WithField("device_id", settings.DeviceID).Info("Device settings created successfully via Supabase")
	return nil
}

// GetDeviceSettingsByID retrieves device settings by device_id
func (r *deviceSettingsRepositorySupabase) GetDeviceSettingsByID(deviceID string) (*models.DeviceSettings, error) {
	ctx := context.Background()

	var settings []models.DeviceSettings
	err := r.supabase.From("device_setting").
		Select("*").
		Eq("device_id", deviceID).
		Limit(1).
		Execute(ctx, &settings)

	if err != nil {
		logrus.WithError(err).Error("Failed to get device settings by ID")
		return nil, fmt.Errorf("failed to get device settings: %w", err)
	}

	if len(settings) == 0 {
		return nil, nil
	}

	return &settings[0], nil
}

// GetDeviceSettingsByDevice retrieves device settings by id_device
func (r *deviceSettingsRepositorySupabase) GetDeviceSettingsByDevice(idDevice string) (*models.DeviceSettings, error) {
	ctx := context.Background()

	var settings []models.DeviceSettings
	err := r.supabase.From("device_setting").
		Select("*").
		Eq("id_device", idDevice).
		Limit(1).
		Execute(ctx, &settings)

	if err != nil {
		logrus.WithError(err).Error("Failed to get device settings by device")
		return nil, fmt.Errorf("failed to get device settings: %w", err)
	}

	if len(settings) == 0 {
		return nil, nil
	}

	return &settings[0], nil
}

// GetAllDeviceSettings retrieves all device settings
func (r *deviceSettingsRepositorySupabase) GetAllDeviceSettings() ([]models.DeviceSettings, error) {
	ctx := context.Background()

	var settings []models.DeviceSettings
	err := r.supabase.From("device_setting").
		Select("*").
		Order("created_at", false). // descending
		Execute(ctx, &settings)

	if err != nil {
		logrus.WithError(err).Error("Failed to get all device settings")
		return nil, fmt.Errorf("failed to get all device settings: %w", err)
	}

	return settings, nil
}

// GetDeviceSettingsByProvider retrieves device settings by provider
func (r *deviceSettingsRepositorySupabase) GetDeviceSettingsByProvider(provider string) ([]models.DeviceSettings, error) {
	ctx := context.Background()

	var settings []models.DeviceSettings
	err := r.supabase.From("device_setting").
		Select("*").
		Eq("provider", provider).
		Order("created_at", false).
		Execute(ctx, &settings)

	if err != nil {
		logrus.WithError(err).Error("Failed to get device settings by provider")
		return nil, fmt.Errorf("failed to get device settings by provider: %w", err)
	}

	return settings, nil
}

// GetAPIKeyByDevice retrieves API key for a specific device
func (r *deviceSettingsRepositorySupabase) GetAPIKeyByDevice(idDevice string) (string, error) {
	ctx := context.Background()

	var settings []models.DeviceSettings
	err := r.supabase.From("device_setting").
		Select("api_key").
		Eq("id_device", idDevice).
		Limit(1).
		Execute(ctx, &settings)

	if err != nil {
		logrus.WithError(err).Error("Failed to get API key by device")
		return "", fmt.Errorf("failed to get API key: %w", err)
	}

	if len(settings) == 0 {
		return "", nil
	}

	if settings[0].APIKey.Valid {
		return settings[0].APIKey.String, nil
	}

	return "", nil
}

// GetProviderByDevice retrieves provider for a specific device
func (r *deviceSettingsRepositorySupabase) GetProviderByDevice(idDevice string) (string, error) {
	ctx := context.Background()

	var settings []models.DeviceSettings
	err := r.supabase.From("device_setting").
		Select("provider").
		Eq("id_device", idDevice).
		Limit(1).
		Execute(ctx, &settings)

	if err != nil {
		logrus.WithError(err).Error("Failed to get provider by device")
		return "", fmt.Errorf("failed to get provider: %w", err)
	}

	if len(settings) == 0 {
		return "", nil
	}

	return settings[0].Provider, nil
}

// GetAPIKeyOptionByDevice retrieves API key option for a specific device
func (r *deviceSettingsRepositorySupabase) GetAPIKeyOptionByDevice(idDevice string) (string, error) {
	ctx := context.Background()

	var settings []models.DeviceSettings
	err := r.supabase.From("device_setting").
		Select("api_key_option").
		Eq("id_device", idDevice).
		Limit(1).
		Execute(ctx, &settings)

	if err != nil {
		logrus.WithError(err).Error("Failed to get API key option by device")
		return "", fmt.Errorf("failed to get API key option: %w", err)
	}

	if len(settings) == 0 {
		return "", nil
	}

	return settings[0].APIKeyOption, nil
}

// UpdateDeviceSettings updates an existing device settings record
func (r *deviceSettingsRepositorySupabase) UpdateDeviceSettings(settings *models.DeviceSettings) error {
	ctx := context.Background()
	settings.UpdatedAt = time.Now()

	updateData := map[string]interface{}{
		"api_key_option": settings.APIKeyOption,
		"webhook_id":     settings.WebhookID,
		"provider":       settings.Provider,
		"api_key":        settings.APIKey,
		"id_device":      settings.IDDevice,
		"updated_at":     settings.UpdatedAt,
	}

	err := r.supabase.From("device_setting").
		Eq("device_id", settings.DeviceID).
		Update(ctx, updateData)

	if err != nil {
		logrus.WithError(err).Error("Failed to update device settings")
		return fmt.Errorf("failed to update device settings: %w", err)
	}

	logrus.WithField("device_id", settings.DeviceID).Info("Device settings updated successfully via Supabase")
	return nil
}

// UpdateAPIKey updates the API key for a specific device
func (r *deviceSettingsRepositorySupabase) UpdateAPIKey(deviceID string, apiKey string) error {
	ctx := context.Background()

	updateData := map[string]interface{}{
		"api_key":    apiKey,
		"updated_at": time.Now(),
	}

	err := r.supabase.From("device_setting").
		Eq("device_id", deviceID).
		Update(ctx, updateData)

	if err != nil {
		logrus.WithError(err).Error("Failed to update API key")
		return fmt.Errorf("failed to update API key: %w", err)
	}

	logrus.WithField("device_id", deviceID).Info("API key updated successfully via Supabase")
	return nil
}

// UpdateProvider updates the provider for a specific device
func (r *deviceSettingsRepositorySupabase) UpdateProvider(deviceID string, provider string) error {
	ctx := context.Background()

	updateData := map[string]interface{}{
		"provider":   provider,
		"updated_at": time.Now(),
	}

	err := r.supabase.From("device_setting").
		Eq("device_id", deviceID).
		Update(ctx, updateData)

	if err != nil {
		logrus.WithError(err).Error("Failed to update provider")
		return fmt.Errorf("failed to update provider: %w", err)
	}

	logrus.WithField("device_id", deviceID).Info("Provider updated successfully via Supabase")
	return nil
}

// UpdateAPIKeyOption updates the API key option for a specific device
func (r *deviceSettingsRepositorySupabase) UpdateAPIKeyOption(deviceID string, apiKeyOption string) error {
	ctx := context.Background()

	updateData := map[string]interface{}{
		"api_key_option": apiKeyOption,
		"updated_at":     time.Now(),
	}

	err := r.supabase.From("device_setting").
		Eq("device_id", deviceID).
		Update(ctx, updateData)

	if err != nil {
		logrus.WithError(err).Error("Failed to update API key option")
		return fmt.Errorf("failed to update API key option: %w", err)
	}

	logrus.WithField("device_id", deviceID).Info("API key option updated successfully via Supabase")
	return nil
}

// UpdateWebhookID updates the webhook ID for a specific device
func (r *deviceSettingsRepositorySupabase) UpdateWebhookID(deviceID string, webhookID string) error {
	ctx := context.Background()

	updateData := map[string]interface{}{
		"webhook_id": webhookID,
		"updated_at": time.Now(),
	}

	err := r.supabase.From("device_setting").
		Eq("device_id", deviceID).
		Update(ctx, updateData)

	if err != nil {
		logrus.WithError(err).Error("Failed to update webhook ID")
		return fmt.Errorf("failed to update webhook ID: %w", err)
	}

	logrus.WithField("device_id", deviceID).Info("Webhook ID updated successfully via Supabase")
	return nil
}

// DeleteDeviceSettings deletes device settings by device_id
func (r *deviceSettingsRepositorySupabase) DeleteDeviceSettings(deviceID string) error {
	ctx := context.Background()

	err := r.supabase.From("device_setting").
		Eq("device_id", deviceID).
		Delete(ctx)

	if err != nil {
		logrus.WithError(err).Error("Failed to delete device settings")
		return fmt.Errorf("failed to delete device settings: %w", err)
	}

	logrus.WithField("device_id", deviceID).Info("Device settings deleted successfully via Supabase")
	return nil
}

// DeviceExists checks if a device exists in the settings
func (r *deviceSettingsRepositorySupabase) DeviceExists(idDevice string) (bool, error) {
	ctx := context.Background()

	var settings []models.DeviceSettings
	err := r.supabase.From("device_setting").
		Select("device_id").
		Eq("id_device", idDevice).
		Limit(1).
		Execute(ctx, &settings)

	if err != nil {
		logrus.WithError(err).Error("Failed to check if device exists")
		return false, fmt.Errorf("failed to check device existence: %w", err)
	}

	return len(settings) > 0, nil
}

// GetDeviceCount returns the total number of configured devices
func (r *deviceSettingsRepositorySupabase) GetDeviceCount() (int, error) {
	ctx := context.Background()

	var settings []models.DeviceSettings
	err := r.supabase.From("device_setting").
		Select("device_id").
		Execute(ctx, &settings)

	if err != nil {
		logrus.WithError(err).Error("Failed to get device count")
		return 0, fmt.Errorf("failed to get device count: %w", err)
	}

	return len(settings), nil
}
