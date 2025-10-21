package service

import (
	"chatbot-automation/internal/models"
	"chatbot-automation/internal/repository"
	"chatbot-automation/internal/whatsapp"
	"context"
	"fmt"
)

// WhatsAppService handles WhatsApp message sending
type WhatsAppService struct {
	deviceRepo *repository.DeviceRepository
	providers  map[string]whatsapp.Provider
}

// NewWhatsAppService creates a new WhatsApp service
func NewWhatsAppService(deviceRepo *repository.DeviceRepository) *WhatsAppService {
	return &WhatsAppService{
		deviceRepo: deviceRepo,
		providers:  make(map[string]whatsapp.Provider),
	}
}

// SendMessage sends a WhatsApp message using the appropriate provider
func (s *WhatsAppService) SendMessage(ctx context.Context, deviceID string, to string, message string, mediaType string, mediaURL string) error {
	// Get device
	device, err := s.deviceRepo.GetDeviceByDeviceID(ctx, deviceID)
	if err != nil {
		device, err = s.deviceRepo.GetDeviceByID(ctx, deviceID)
		if err != nil {
			return fmt.Errorf("device not found: %w", err)
		}
	}

	if device == nil {
		return fmt.Errorf("device not found")
	}

	// Get provider configuration from device
	provider := device.Provider
	if provider == "" {
		provider = "waha" // Default
	}

	apiKey := ""
	if device.APIKey != nil {
		apiKey = *device.APIKey
	}

	// Base URL - you may want to add this field to DeviceSetting model or use a config
	baseURL := "https://api.waha.pro" // Default Waha URL
	// TODO: Add APIURL field to DeviceSetting model for custom base URLs

	instance := deviceID
	if device.Instance != nil && *device.Instance != "" {
		instance = *device.Instance
	} else if device.IDDevice != nil && *device.IDDevice != "" {
		instance = *device.IDDevice
	}

	// Get or create provider
	whatsappProvider, err := s.getProvider(provider, baseURL, apiKey, instance)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	// Build message request
	req := &models.SendMessageRequest{
		To:   to,
		Body: message,
		Type: "text",
	}

	// Set media type and URL if provided
	if mediaType != "" && mediaURL != "" {
		req.Type = mediaType
		req.MediaURL = mediaURL
	}

	// Send message
	_, err = whatsappProvider.SendMessage(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// getProvider gets or creates a WhatsApp provider instance
func (s *WhatsAppService) getProvider(providerName string, baseURL string, apiKey string, instance string) (whatsapp.Provider, error) {
	// Check cache
	cacheKey := fmt.Sprintf("%s:%s", providerName, instance)
	if provider, ok := s.providers[cacheKey]; ok {
		return provider, nil
	}

	// Create new provider
	config := &whatsapp.ProviderConfig{
		BaseURL:  baseURL,
		APIKey:   apiKey,
		Instance: instance,
	}

	var provider whatsapp.Provider
	switch providerName {
	case "waha":
		provider = whatsapp.NewWahaProvider(config)
	case "wablas":
		provider = whatsapp.NewWablasProvider(config)
	case "whacenter":
		provider = whatsapp.NewWhacenterProvider(config)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}

	// Cache provider
	s.providers[cacheKey] = provider

	return provider, nil
}
