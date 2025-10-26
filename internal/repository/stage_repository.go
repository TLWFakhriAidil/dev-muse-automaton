package repository

import (
	"chatbot-automation/internal/database"
	"chatbot-automation/internal/models"
	"context"
	"encoding/json"
	"fmt"
)

// StageRepository handles stage value data operations
type StageRepository struct {
	supabase *database.SupabaseClient
}

// NewStageRepository creates a new stage repository
func NewStageRepository(supabase *database.SupabaseClient) *StageRepository {
	return &StageRepository{
		supabase: supabase,
	}
}

// CreateStageValue creates a new stage value
func (r *StageRepository) CreateStageValue(ctx context.Context, stage *models.StageValue) error {
	// Log the data being sent
	stageJSON, _ := json.Marshal(stage)
	fmt.Printf("🔍 Creating stage value: %s\n", string(stageJSON))

	data, err := r.supabase.InsertAsAdmin("stagesetvalue", stage)
	if err != nil {
		fmt.Printf("❌ Database error: %v\n", err)
		return fmt.Errorf("failed to create stage value: %w", err)
	}

	fmt.Printf("✅ Database response: %s\n", string(data))

	var stages []models.StageValue
	if err := json.Unmarshal(data, &stages); err != nil {
		fmt.Printf("❌ JSON parse error: %v\n", err)
		return fmt.Errorf("failed to parse created stage value: %w", err)
	}

	if len(stages) > 0 {
		*stage = stages[0]
		fmt.Printf("✅ Stage value created with ID: %d\n", stage.ID)
	} else {
		fmt.Printf("⚠️  No stage returned from database\n")
	}

	return nil
}

// GetStageValueByID retrieves a stage value by ID
func (r *StageRepository) GetStageValueByID(ctx context.Context, stageID int) (*models.StageValue, error) {
	data, err := r.supabase.QueryAsAdmin("stagesetvalue", map[string]string{
		"select":            "*",
		"stagesetvalue_id": fmt.Sprintf("eq.%d", stageID),
		"limit":             "1",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get stage value: %w", err)
	}

	var stages []models.StageValue
	if err := json.Unmarshal(data, &stages); err != nil {
		return nil, fmt.Errorf("failed to parse stage value: %w", err)
	}

	if len(stages) == 0 {
		return nil, fmt.Errorf("stage value not found")
	}

	return &stages[0], nil
}

// GetAllStageValues retrieves all stage values
func (r *StageRepository) GetAllStageValues(ctx context.Context) ([]models.StageValue, error) {
	data, err := r.supabase.QueryAsAdmin("stagesetvalue", map[string]string{
		"select": "*",
		"order":  "stagesetvalue_id.desc",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get stage values: %w", err)
	}

	var stages []models.StageValue
	if err := json.Unmarshal(data, &stages); err != nil {
		return nil, fmt.Errorf("failed to parse stage values: %w", err)
	}

	return stages, nil
}

// UpdateStageValue updates a stage value
func (r *StageRepository) UpdateStageValue(ctx context.Context, stageID int, updates map[string]interface{}) error {
	_, err := r.supabase.UpdateAsAdmin("stagesetvalue", map[string]string{
		"stagesetvalue_id": fmt.Sprintf("%d", stageID),
	}, updates)

	if err != nil {
		return fmt.Errorf("failed to update stage value: %w", err)
	}

	return nil
}

// DeleteStageValue deletes a stage value
func (r *StageRepository) DeleteStageValue(ctx context.Context, stageID int) error {
	err := r.supabase.DeleteAsAdmin("stagesetvalue", map[string]string{
		"stagesetvalue_id": fmt.Sprintf("%d", stageID),
	})

	if err != nil {
		return fmt.Errorf("failed to delete stage value: %w", err)
	}

	return nil
}
