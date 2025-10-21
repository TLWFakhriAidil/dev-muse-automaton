package service

import (
	"chatbot-automation/internal/models"
	"chatbot-automation/internal/repository"
	"context"
	"fmt"
)

// FlowService handles flow business logic
type FlowService struct {
	flowRepo   *repository.FlowRepository
	deviceRepo *repository.DeviceRepository
}

// NewFlowService creates a new flow service
func NewFlowService(flowRepo *repository.FlowRepository, deviceRepo *repository.DeviceRepository) *FlowService {
	return &FlowService{
		flowRepo:   flowRepo,
		deviceRepo: deviceRepo,
	}
}

// CreateFlow creates a new flow for a device
func (s *FlowService) CreateFlow(ctx context.Context, userID string, req *models.CreateFlowRequest) (*models.FlowResponse, error) {
	// Try to find device by device_id field first, then by UUID id
	device, err := s.deviceRepo.GetDeviceByDeviceID(ctx, req.IDDevice)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup device: %w", err)
	}

	// If not found by device_id, try by UUID id
	if device == nil {
		device, err = s.deviceRepo.GetDeviceByID(ctx, req.IDDevice)
		if err != nil {
			return &models.FlowResponse{
				Success: false,
				Message: "Device not found",
			}, nil
		}
	}

	// Verify ownership
	if device.UserID == nil || *device.UserID != userID {
		return &models.FlowResponse{
			Success: false,
			Message: "Access denied - device does not belong to you",
		}, nil
	}

	// Create flow using the user-friendly device identifier
	// Try IDDevice first, fallback to DeviceID, then to ID as last resort
	deviceIdentifier := req.IDDevice // Use what user provided
	if device.IDDevice != nil && *device.IDDevice != "" {
		deviceIdentifier = *device.IDDevice
	} else if device.DeviceID != nil && *device.DeviceID != "" {
		deviceIdentifier = *device.DeviceID
	}

	flow := &models.ChatbotFlow{
		IDDevice: deviceIdentifier, // Use the user-friendly identifier
		Name:     req.Name,
		Niche:    req.Niche,
		Nodes:    req.Nodes,
		Edges:    req.Edges,
	}

	if err := s.flowRepo.CreateFlow(ctx, flow); err != nil {
		return nil, fmt.Errorf("failed to create flow: %w", err)
	}

	return &models.FlowResponse{
		Success: true,
		Message: "Flow created successfully",
		Flow:    flow,
	}, nil
}

// GetFlow retrieves a flow by ID or device identifier
// If flowID looks like a device identifier, gets the first flow for that device
func (s *FlowService) GetFlow(ctx context.Context, userID, flowID string) (*models.FlowResponse, error) {
	// Try to get flow by UUID first
	flow, err := s.flowRepo.GetFlowByID(ctx, flowID)

	// If not found by UUID, try as device identifier
	if err != nil {
		// Look up device to verify ownership
		device, err := s.deviceRepo.GetDeviceByDeviceID(ctx, flowID)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup device: %w", err)
		}

		if device == nil {
			device, err = s.deviceRepo.GetDeviceByID(ctx, flowID)
			if err != nil {
				return &models.FlowResponse{
					Success: false,
					Message: "Flow not found",
				}, nil
			}
		}

		// Verify ownership
		if device.UserID == nil || *device.UserID != userID {
			return &models.FlowResponse{
				Success: false,
				Message: "Access denied",
			}, nil
		}

		// Get flows for this device using the device identifier, not UUID
		flows, err := s.flowRepo.GetFlowsByDeviceID(ctx, flowID)
		if err != nil || len(flows) == 0 {
			return &models.FlowResponse{
				Success: false,
				Message: "Flow not found",
			}, nil
		}

		// Return first flow
		return &models.FlowResponse{
			Success: true,
			Flow:    &flows[0],
		}, nil
	}

	// Flow found by UUID, verify device ownership
	device, err := s.deviceRepo.GetDeviceByDeviceID(ctx, flow.IDDevice)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup device: %w", err)
	}

	if device == nil {
		device, err = s.deviceRepo.GetDeviceByID(ctx, flow.IDDevice)
		if err != nil || device.UserID == nil || *device.UserID != userID {
			return &models.FlowResponse{
				Success: false,
				Message: "Access denied",
			}, nil
		}
	} else if device.UserID == nil || *device.UserID != userID {
		return &models.FlowResponse{
			Success: false,
			Message: "Access denied",
		}, nil
	}

	return &models.FlowResponse{
		Success: true,
		Flow:    flow,
	}, nil
}

// GetFlowsByDevice retrieves all flows for a specific device
func (s *FlowService) GetFlowsByDevice(ctx context.Context, userID, deviceID string) (*models.FlowResponse, error) {
	// Try to find device by device_id field first, then by UUID id
	device, err := s.deviceRepo.GetDeviceByDeviceID(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup device: %w", err)
	}

	// If not found by device_id, try by UUID id
	if device == nil {
		device, err = s.deviceRepo.GetDeviceByID(ctx, deviceID)
		if err != nil {
			return &models.FlowResponse{
				Success: false,
				Message: "Device not found",
			}, nil
		}
	}

	// Verify ownership
	if device.UserID == nil || *device.UserID != userID {
		return &models.FlowResponse{
			Success: false,
			Message: "Access denied",
		}, nil
	}

	// Get flows using the device's UUID id
	flows, err := s.flowRepo.GetFlowsByDeviceID(ctx, device.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get flows: %w", err)
	}

	return &models.FlowResponse{
		Success: true,
		Message: fmt.Sprintf("Found %d flows", len(flows)),
		Flows:   flows,
	}, nil
}

// GetAllUserFlows retrieves all flows for all user devices
func (s *FlowService) GetAllUserFlows(ctx context.Context, userID string) (*models.FlowResponse, error) {
	// Get all user devices
	devices, err := s.deviceRepo.GetDevicesByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user devices: %w", err)
	}

	if len(devices) == 0 {
		return &models.FlowResponse{
			Success: true,
			Message: "No devices found",
			Flows:   []models.ChatbotFlow{},
		}, nil
	}

	// Extract device IDs
	deviceIDs := make([]string, len(devices))
	for i, device := range devices {
		if device.IDDevice != nil {
			deviceIDs[i] = *device.IDDevice
		}
	}

	// Get all flows for user devices
	flows, err := s.flowRepo.GetAllFlowsByUserDevices(ctx, deviceIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get flows: %w", err)
	}

	return &models.FlowResponse{
		Success: true,
		Message: fmt.Sprintf("Found %d flows across %d devices", len(flows), len(devices)),
		Flows:   flows,
	}, nil
}

// UpdateFlow updates a flow by UUID or device identifier
func (s *FlowService) UpdateFlow(ctx context.Context, userID, flowID string, req *models.UpdateFlowRequest) (*models.FlowResponse, error) {
	// Try to get flow by UUID first
	flow, err := s.flowRepo.GetFlowByID(ctx, flowID)

	// If not found by UUID, try as device identifier
	if err != nil {
		// Look up device
		device, err := s.deviceRepo.GetDeviceByDeviceID(ctx, flowID)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup device: %w", err)
		}

		if device == nil {
			device, err = s.deviceRepo.GetDeviceByID(ctx, flowID)
			if err != nil {
				return &models.FlowResponse{
					Success: false,
					Message: "Flow not found",
				}, nil
			}
		}

		// Verify ownership
		if device.UserID == nil || *device.UserID != userID {
			return &models.FlowResponse{
				Success: false,
				Message: "Access denied",
			}, nil
		}

		// Get first flow for this device using the device identifier, not UUID
		flows, err := s.flowRepo.GetFlowsByDeviceID(ctx, flowID)
		if err != nil || len(flows) == 0 {
			return &models.FlowResponse{
				Success: false,
				Message: "Flow not found",
			}, nil
		}

		flow = &flows[0]
	} else {
		// Flow found by UUID, verify device ownership
		device, err := s.deviceRepo.GetDeviceByDeviceID(ctx, flow.IDDevice)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup device: %w", err)
		}

		if device == nil {
			device, err = s.deviceRepo.GetDeviceByID(ctx, flow.IDDevice)
			if err != nil || device.UserID == nil || *device.UserID != userID {
				return &models.FlowResponse{
					Success: false,
					Message: "Access denied",
				}, nil
			}
		} else if device.UserID == nil || *device.UserID != userID {
			return &models.FlowResponse{
				Success: false,
				Message: "Access denied",
			}, nil
		}
	}

	// Build update map
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Niche != nil {
		updates["niche"] = *req.Niche
	}
	if req.Nodes != nil {
		updates["nodes"] = req.Nodes
	}
	if req.Edges != nil {
		updates["edges"] = req.Edges
	}

	if len(updates) == 0 {
		return &models.FlowResponse{
			Success: false,
			Message: "No fields to update",
		}, nil
	}

	// Update using the flow's actual UUID
	if err := s.flowRepo.UpdateFlow(ctx, flow.ID, updates); err != nil {
		return nil, fmt.Errorf("failed to update flow: %w", err)
	}

	// Get updated flow
	updatedFlow, _ := s.flowRepo.GetFlowByID(ctx, flow.ID)

	return &models.FlowResponse{
		Success: true,
		Message: "Flow updated successfully",
		Flow:    updatedFlow,
	}, nil
}

// DeleteFlow deletes a flow by UUID or device identifier
func (s *FlowService) DeleteFlow(ctx context.Context, userID, flowID string) (*models.FlowResponse, error) {
	// Try to get flow by UUID first
	flow, err := s.flowRepo.GetFlowByID(ctx, flowID)

	// If not found by UUID, try as device identifier
	if err != nil {
		// Look up device
		device, err := s.deviceRepo.GetDeviceByDeviceID(ctx, flowID)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup device: %w", err)
		}

		if device == nil {
			device, err = s.deviceRepo.GetDeviceByID(ctx, flowID)
			if err != nil {
				return &models.FlowResponse{
					Success: false,
					Message: "Flow not found",
				}, nil
			}
		}

		// Verify ownership
		if device.UserID == nil || *device.UserID != userID {
			return &models.FlowResponse{
				Success: false,
				Message: "Access denied",
			}, nil
		}

		// Get first flow for this device using the device identifier, not UUID
		flows, err := s.flowRepo.GetFlowsByDeviceID(ctx, flowID)
		if err != nil || len(flows) == 0 {
			return &models.FlowResponse{
				Success: false,
				Message: "Flow not found",
			}, nil
		}

		flow = &flows[0]
	} else {
		// Flow found by UUID, verify device ownership
		device, err := s.deviceRepo.GetDeviceByDeviceID(ctx, flow.IDDevice)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup device: %w", err)
		}

		if device == nil {
			device, err = s.deviceRepo.GetDeviceByID(ctx, flow.IDDevice)
			if err != nil || device.UserID == nil || *device.UserID != userID {
				return &models.FlowResponse{
					Success: false,
					Message: "Access denied",
				}, nil
			}
		} else if device.UserID == nil || *device.UserID != userID {
			return &models.FlowResponse{
				Success: false,
				Message: "Access denied",
			}, nil
		}
	}

	// Delete using the flow's actual UUID
	if err := s.flowRepo.DeleteFlow(ctx, flow.ID); err != nil {
		return nil, fmt.Errorf("failed to delete flow: %w", err)
	}

	return &models.FlowResponse{
		Success: true,
		Message: "Flow deleted successfully",
	}, nil
}
