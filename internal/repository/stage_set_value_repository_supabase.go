package repository

import (
	"context"
	"fmt"
	"time"

	"nodepath-chat/internal/database"
	"nodepath-chat/internal/models"
)

// StageSetValueRepositorySupabase implements stage set value repository using Supabase SDK
type StageSetValueRepositorySupabase struct {
	supabase *database.SupabaseSDK
}

// NewStageSetValueRepositorySupabase creates a Supabase-based stage set value repository
func NewStageSetValueRepositorySupabase(supabase *database.SupabaseSDK) *StageSetValueRepositorySupabase {
	return &StageSetValueRepositorySupabase{
		supabase: supabase,
	}
}

// GetAll retrieves all stage set values
func (r *StageSetValueRepositorySupabase) GetAll() ([]*models.StageSetValue, error) {
	ctx := context.Background()

	var values []models.StageSetValue
	err := r.supabase.From("stageSetValue").
		Select("*").
		Order("stageSetValue_id", false). // descending
		Execute(ctx, &values)

	if err != nil {
		return nil, fmt.Errorf("failed to query stage set values via Supabase: %w", err)
	}

	// Convert []models.StageSetValue to []*models.StageSetValue
	result := make([]*models.StageSetValue, len(values))
	for i := range values {
		result[i] = &values[i]
	}

	return result, nil
}

// GetByDeviceID retrieves all stage set values for a specific device
func (r *StageSetValueRepositorySupabase) GetByDeviceID(deviceID string) ([]*models.StageSetValue, error) {
	ctx := context.Background()

	var values []models.StageSetValue
	err := r.supabase.From("stageSetValue").
		Select("*").
		Eq("id_device", deviceID).
		Order("stage", true). // ascending
		Execute(ctx, &values)

	if err != nil {
		return nil, fmt.Errorf("failed to query stage set values by device via Supabase: %w", err)
	}

	// Convert []models.StageSetValue to []*models.StageSetValue
	result := make([]*models.StageSetValue, len(values))
	for i := range values {
		result[i] = &values[i]
	}

	return result, nil
}

// GetByStage retrieves a stage set value by device and stage number
func (r *StageSetValueRepositorySupabase) GetByStage(deviceID string, stage int) (*models.StageSetValue, error) {
	ctx := context.Background()

	var values []models.StageSetValue
	err := r.supabase.From("stageSetValue").
		Select("*").
		Eq("id_device", deviceID).
		Eq("stage", stage).
		Limit(1).
		Execute(ctx, &values)

	if err != nil {
		return nil, fmt.Errorf("failed to get stage set value via Supabase: %w", err)
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("no rows found") // Mimic sql.ErrNoRows behavior
	}

	return &values[0], nil
}

// Create inserts a new stage set value
func (r *StageSetValueRepositorySupabase) Create(value *models.StageSetValue) error {
	ctx := context.Background()

	now := time.Now()
	data := map[string]interface{}{
		"id_device":      value.IDDevice,
		"stage":          value.Stage,
		"type_inputData": value.TypeInputData,
		"columnsData":    value.ColumnsData,
		"inputHardCode":  value.InputHardCode,
		"created_at":     now,
		"updated_at":     now,
	}

	var result models.StageSetValue
	err := r.supabase.From("stageSetValue").Insert(ctx, data, &result)
	if err != nil {
		return fmt.Errorf("failed to create stage set value via Supabase: %w", err)
	}

	value.StageSetValueID = result.StageSetValueID
	value.CreatedAt = result.CreatedAt
	value.UpdatedAt = result.UpdatedAt

	return nil
}

// Update updates an existing stage set value
func (r *StageSetValueRepositorySupabase) Update(value *models.StageSetValue) error {
	ctx := context.Background()

	updateData := map[string]interface{}{
		"stage":          value.Stage,
		"type_inputData": value.TypeInputData,
		"columnsData":    value.ColumnsData,
		"inputHardCode":  value.InputHardCode,
		"updated_at":     time.Now(),
	}

	err := r.supabase.From("stageSetValue").
		Eq("stageSetValue_id", value.StageSetValueID).
		Update(ctx, updateData)

	if err != nil {
		return fmt.Errorf("failed to update stage set value via Supabase: %w", err)
	}

	return nil
}

// Delete removes a stage set value by ID
func (r *StageSetValueRepositorySupabase) Delete(id int) error {
	ctx := context.Background()

	err := r.supabase.From("stageSetValue").
		Eq("stageSetValue_id", id).
		Delete(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete stage set value via Supabase: %w", err)
	}

	return nil
}
