package service

import (
	"chatbot-automation/internal/models"
	"chatbot-automation/internal/repository"
	"context"
	"fmt"
)

// StageService handles stage value business logic
type StageService struct {
	stageRepo *repository.StageRepository
}

// NewStageService creates a new stage service
func NewStageService(stageRepo *repository.StageRepository) *StageService {
	return &StageService{
		stageRepo: stageRepo,
	}
}

// CreateStageValue creates a new stage value
func (s *StageService) CreateStageValue(ctx context.Context, req *models.CreateStageValueRequest) (*models.StageValueResponse, error) {
	stage := &models.StageValue{
		IDDevice:      req.IDDevice,
		Stage:         req.Stage,
		TypeInputData: req.TypeInputData,
		ColumnsData:   req.ColumnsData,
		InputHardCode: req.InputHardCode,
	}

	if err := s.stageRepo.CreateStageValue(ctx, stage); err != nil {
		return nil, fmt.Errorf("failed to create stage value: %w", err)
	}

	return &models.StageValueResponse{
		Success:    true,
		Message:    "Stage value created successfully",
		StageValue: stage,
	}, nil
}

// GetStageValue retrieves a stage value by ID
func (s *StageService) GetStageValue(ctx context.Context, stageID int) (*models.StageValueResponse, error) {
	stage, err := s.stageRepo.GetStageValueByID(ctx, stageID)
	if err != nil {
		return &models.StageValueResponse{
			Success: false,
			Message: "Stage value not found",
		}, nil
	}

	return &models.StageValueResponse{
		Success:    true,
		StageValue: stage,
	}, nil
}

// GetAllStageValues retrieves all stage values
func (s *StageService) GetAllStageValues(ctx context.Context) (*models.StageValueResponse, error) {
	stages, err := s.stageRepo.GetAllStageValues(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get stage values: %w", err)
	}

	return &models.StageValueResponse{
		Success:     true,
		Message:     fmt.Sprintf("Found %d stage values", len(stages)),
		StageValues: stages,
	}, nil
}

// UpdateStageValue updates a stage value
func (s *StageService) UpdateStageValue(ctx context.Context, stageID int, req *models.UpdateStageValueRequest) (*models.StageValueResponse, error) {
	// Get stage to check if it exists
	_, err := s.stageRepo.GetStageValueByID(ctx, stageID)
	if err != nil {
		return &models.StageValueResponse{
			Success: false,
			Message: "Stage value not found",
		}, nil
	}

	// Build update map
	updates := make(map[string]interface{})

	if req.IDDevice != nil {
		updates["id_device"] = *req.IDDevice
	}
	if req.Stage != nil {
		updates["stage"] = *req.Stage
	}
	if req.TypeInputData != nil {
		updates["type_inputdata"] = *req.TypeInputData
	}
	if req.ColumnsData != nil {
		updates["columnsdata"] = *req.ColumnsData
	}
	if req.InputHardCode != nil {
		updates["inputhardcode"] = *req.InputHardCode
	}

	if len(updates) == 0 {
		return &models.StageValueResponse{
			Success: false,
			Message: "No fields to update",
		}, nil
	}

	if err := s.stageRepo.UpdateStageValue(ctx, stageID, updates); err != nil {
		return nil, fmt.Errorf("failed to update stage value: %w", err)
	}

	// Get updated stage
	updatedStage, _ := s.stageRepo.GetStageValueByID(ctx, stageID)

	return &models.StageValueResponse{
		Success:    true,
		Message:    "Stage value updated successfully",
		StageValue: updatedStage,
	}, nil
}

// DeleteStageValue deletes a stage value
func (s *StageService) DeleteStageValue(ctx context.Context, stageID int) (*models.StageValueResponse, error) {
	// Get stage to check if it exists
	_, err := s.stageRepo.GetStageValueByID(ctx, stageID)
	if err != nil {
		return &models.StageValueResponse{
			Success: false,
			Message: "Stage value not found",
		}, nil
	}

	if err := s.stageRepo.DeleteStageValue(ctx, stageID); err != nil {
		return nil, fmt.Errorf("failed to delete stage value: %w", err)
	}

	return &models.StageValueResponse{
		Success: true,
		Message: "Stage value deleted successfully",
	}, nil
}
