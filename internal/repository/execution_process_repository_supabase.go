package repository

import (
	"context"
	"fmt"
	"time"

	"nodepath-chat/internal/database"
	"nodepath-chat/internal/models"

	"github.com/sirupsen/logrus"
)

// executionProcessRepositorySupabase implements ExecutionProcessRepository using Supabase SDK
type executionProcessRepositorySupabase struct {
	supabase *database.SupabaseSDK
}

// NewExecutionProcessRepositorySupabase creates a Supabase-based execution process repository
func NewExecutionProcessRepositorySupabase(supabase *database.SupabaseSDK) ExecutionProcessRepository {
	return &executionProcessRepositorySupabase{supabase: supabase}
}

// CreateExecution creates a new execution record and returns its ID
func (r *executionProcessRepositorySupabase) CreateExecution(idDevice, idProspect string) (int, error) {
	ctx := context.Background()

	data := map[string]interface{}{
		"id_device":  idDevice,
		"id_prospect": idProspect,
		"times":      time.Now(),
	}

	var result models.ExecutionProcess
	err := r.supabase.From("execution_process").Insert(ctx, data, &result)
	if err != nil {
		logrus.WithError(err).Error("Failed to create execution record via Supabase")
		return 0, fmt.Errorf("failed to create execution record: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"id_device":    idDevice,
		"id_prospect":  idProspect,
		"id_execution": result.IDChatInput,
	}).Info("✅ Created execution record via Supabase")

	return result.IDChatInput, nil
}

// GetOldestExecution gets the oldest execution record for a device and prospect
func (r *executionProcessRepositorySupabase) GetOldestExecution(idDevice, idProspect string) (*models.ExecutionProcess, error) {
	ctx := context.Background()

	var executions []models.ExecutionProcess
	err := r.supabase.From("execution_process").
		Select("*").
		Eq("id_device", idDevice).
		Eq("id_prospect", idProspect).
		Order("id_chatInput", true). // ascending = oldest first
		Limit(1).
		Execute(ctx, &executions)

	if err != nil {
		logrus.WithError(err).Error("Failed to get oldest execution record via Supabase")
		return nil, fmt.Errorf("failed to get oldest execution record: %w", err)
	}

	if len(executions) == 0 {
		return nil, nil
	}

	return &executions[0], nil
}

// DeleteExecutions deletes all execution records for a device and prospect
func (r *executionProcessRepositorySupabase) DeleteExecutions(idDevice, idProspect string) error {
	ctx := context.Background()

	err := r.supabase.From("execution_process").
		Eq("id_device", idDevice).
		Eq("id_prospect", idProspect).
		Delete(ctx)

	if err != nil {
		logrus.WithError(err).Error("Failed to delete execution records via Supabase")
		return fmt.Errorf("failed to delete execution records: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"id_device":   idDevice,
		"id_prospect": idProspect,
	}).Info("🧹 Cleaned up execution records via Supabase")

	return nil
}
