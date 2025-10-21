package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"chatbot-automation/internal/database"
	"chatbot-automation/internal/models"
	"chatbot-automation/internal/utils"

	"github.com/sirupsen/logrus"
)

// aiWhatsappRepositorySupabase implements AIWhatsappRepository using Supabase SDK
type aiWhatsappRepositorySupabase struct {
	supabase *database.SupabaseSDK
}

// NewAIWhatsappRepositorySupabase creates a Supabase-based AI WhatsApp repository
func NewAIWhatsappRepositorySupabase(supabase *database.SupabaseSDK) AIWhatsappRepository {
	return &aiWhatsappRepositorySupabase{supabase: supabase}
}

// GetDB returns nil since Supabase SDK doesn't use sql.DB
// This maintains interface compatibility but transactions should use RPC
func (r *aiWhatsappRepositorySupabase) GetDB() *sql.DB {
	return nil
}

// CreateAIWhatsapp creates a new AI WhatsApp conversation record via Supabase REST API
func (r *aiWhatsappRepositorySupabase) CreateAIWhatsapp(ai *models.AIWhatsapp) error {
	ctx := context.Background()
	ai.CreatedAt = time.Now()
	ai.UpdatedAt = time.Now()

	// Handle ConvLast as sql.NullString
	var convLastValue interface{}
	if ai.ConvLast.Valid {
		convLastValue = ai.ConvLast.String
	} else {
		convLastValue = nil
	}

	data := map[string]interface{}{
		"id_device":        ai.IDDevice,
		"prospect_num":     ai.ProspectNum,
		"prospect_name":    ai.ProspectName,
		"stage":            ai.Stage,
		"date_order":       ai.DateOrder,
		"conv_last":        convLastValue,
		"conv_current":     ai.ConvCurrent,
		"human":            ai.Human,
		"niche":            ai.Niche,
		"intro":            ai.Intro,
		"flow_id":          ai.FlowID,
		"current_node_id":  ai.CurrentNodeID,
		"last_node_id":     ai.LastNodeID,
		"waiting_for_reply": ai.WaitingForReply,
		"execution_status": ai.ExecutionStatus,
		"execution_id":     ai.ExecutionID,
		"created_at":       ai.CreatedAt,
		"updated_at":       ai.UpdatedAt,
	}

	var result models.AIWhatsapp
	err := r.supabase.From("ai_whatsapp").Insert(ctx, data, &result)
	if err != nil {
		logrus.WithError(err).WithField("prospect_num", ai.ProspectNum).Error("Failed to create AI WhatsApp record via Supabase")
		return fmt.Errorf("failed to create AI WhatsApp record: %w", err)
	}

	ai.IDProspect = result.IDProspect
	logrus.WithFields(logrus.Fields{
		"id_prospect":  ai.IDProspect,
		"prospect_num": ai.ProspectNum,
		"id_device":    ai.IDDevice,
	}).Info("AI WhatsApp record created successfully via Supabase")

	return nil
}

// GetAIWhatsappByProspectNum retrieves an AI WhatsApp record by prospect number
func (r *aiWhatsappRepositorySupabase) GetAIWhatsappByProspectNum(prospectNum string) (*models.AIWhatsapp, error) {
	ctx := context.Background()

	var results []models.AIWhatsapp
	err := r.supabase.From("ai_whatsapp").
		Select("*").
		Eq("prospect_num", prospectNum).
		Limit(1).
		Execute(ctx, &results)

	if err != nil {
		logrus.WithError(err).Error("Failed to get AI WhatsApp by prospect number via Supabase")
		return nil, fmt.Errorf("failed to get AI WhatsApp record: %w", err)
	}

	if len(results) == 0 {
		return nil, nil
	}

	return &results[0], nil
}

// GetAIWhatsappByID retrieves an AI WhatsApp record by ID
func (r *aiWhatsappRepositorySupabase) GetAIWhatsappByID(id int) (*models.AIWhatsapp, error) {
	ctx := context.Background()

	var results []models.AIWhatsapp
	err := r.supabase.From("ai_whatsapp").
		Select("*").
		Eq("id_prospect", id).
		Limit(1).
		Execute(ctx, &results)

	if err != nil {
		logrus.WithError(err).Error("Failed to get AI WhatsApp by ID via Supabase")
		return nil, fmt.Errorf("failed to get AI WhatsApp record: %w", err)
	}

	if len(results) == 0 {
		return nil, nil
	}

	return &results[0], nil
}

// GetAIWhatsappByDevice retrieves all AI WhatsApp records for a specific device
func (r *aiWhatsappRepositorySupabase) GetAIWhatsappByDevice(idDevice string) ([]models.AIWhatsapp, error) {
	ctx := context.Background()

	var results []models.AIWhatsapp
	err := r.supabase.From("ai_whatsapp").
		Select("*").
		Eq("id_device", idDevice).
		Order("created_at", false). // descending
		Execute(ctx, &results)

	if err != nil {
		logrus.WithError(err).Error("Failed to get AI WhatsApp by device via Supabase")
		return nil, fmt.Errorf("failed to get AI WhatsApp records: %w", err)
	}

	return results, nil
}

// GetAIWhatsappByNiche retrieves AI WhatsApp records by niche
func (r *aiWhatsappRepositorySupabase) GetAIWhatsappByNiche(niche string) ([]models.AIWhatsapp, error) {
	ctx := context.Background()

	var results []models.AIWhatsapp
	err := r.supabase.From("ai_whatsapp").
		Select("*").
		Eq("niche", niche).
		Order("created_at", false).
		Execute(ctx, &results)

	if err != nil {
		logrus.WithError(err).Error("Failed to get AI WhatsApp by niche via Supabase")
		return nil, fmt.Errorf("failed to get AI WhatsApp records: %w", err)
	}

	return results, nil
}

// GetActiveAIConversations retrieves all active AI conversations
func (r *aiWhatsappRepositorySupabase) GetActiveAIConversations() ([]models.AIWhatsapp, error) {
	ctx := context.Background()

	var results []models.AIWhatsapp
	err := r.supabase.From("ai_whatsapp").
		Select("*").
		Eq("execution_status", "active").
		Order("updated_at", false).
		Execute(ctx, &results)

	if err != nil {
		logrus.WithError(err).Error("Failed to get active AI conversations via Supabase")
		return nil, fmt.Errorf("failed to get active AI conversations: %w", err)
	}

	return results, nil
}

// GetConversationHistory retrieves conversation history for a prospect
// Note: This now queries conv_last field from ai_whatsapp table
func (r *aiWhatsappRepositorySupabase) GetConversationHistory(prospectNum string, limit int) ([]models.ConversationLog, error) {
	ctx := context.Background()

	var results []models.AIWhatsapp
	err := r.supabase.From("ai_whatsapp").
		Select("prospect_num, conv_last, updated_at").
		Eq("prospect_num", prospectNum).
		Order("updated_at", false).
		Limit(limit).
		Execute(ctx, &results)

	if err != nil {
		logrus.WithError(err).Error("Failed to get conversation history via Supabase")
		return nil, fmt.Errorf("failed to get conversation history: %w", err)
	}

	// Convert to ConversationLog format
	var logs []models.ConversationLog
	for _, r := range results {
		if r.ConvLast.Valid {
			logs = append(logs, models.ConversationLog{
				ProspectNum: r.ProspectNum,
				Message:     r.ConvLast.String,
				Timestamp:   r.UpdatedAt,
			})
		}
	}

	return logs, nil
}

// GetConversationLogsByStage retrieves conversation logs by stage
func (r *aiWhatsappRepositorySupabase) GetConversationLogsByStage(stage string) ([]models.ConversationLog, error) {
	ctx := context.Background()

	var results []models.AIWhatsapp
	err := r.supabase.From("ai_whatsapp").
		Select("prospect_num, conv_last, updated_at, stage").
		Eq("stage", stage).
		Order("updated_at", false).
		Execute(ctx, &results)

	if err != nil {
		logrus.WithError(err).Error("Failed to get conversation logs by stage via Supabase")
		return nil, fmt.Errorf("failed to get conversation logs: %w", err)
	}

	// Convert to ConversationLog format
	var logs []models.ConversationLog
	for _, r := range results {
		if r.ConvLast.Valid {
			logs = append(logs, models.ConversationLog{
				ProspectNum: r.ProspectNum,
				Message:     r.ConvLast.String,
				Stage:       r.Stage,
				Timestamp:   r.UpdatedAt,
			})
		}
	}

	return logs, nil
}

// GetAIWhatsappByProspectAndDevice retrieves an AI WhatsApp record by prospect and device
func (r *aiWhatsappRepositorySupabase) GetAIWhatsappByProspectAndDevice(prospectNum, idDevice string) (*models.AIWhatsapp, error) {
	ctx := context.Background()

	var results []models.AIWhatsapp
	err := r.supabase.From("ai_whatsapp").
		Select("*").
		Eq("prospect_num", prospectNum).
		Eq("id_device", idDevice).
		Limit(1).
		Execute(ctx, &results)

	if err != nil {
		logrus.WithError(err).Error("Failed to get AI WhatsApp by prospect and device via Supabase")
		return nil, fmt.Errorf("failed to get AI WhatsApp record: %w", err)
	}

	if len(results) == 0 {
		return nil, nil
	}

	return &results[0], nil
}

// UpdateAIWhatsapp updates an existing AI WhatsApp record
func (r *aiWhatsappRepositorySupabase) UpdateAIWhatsapp(ai *models.AIWhatsapp) error {
	ctx := context.Background()
	ai.UpdatedAt = time.Now()

	// Handle ConvLast as sql.NullString
	var convLastValue interface{}
	if ai.ConvLast.Valid {
		convLastValue = ai.ConvLast.String
	} else {
		convLastValue = nil
	}

	updateData := map[string]interface{}{
		"prospect_name":     ai.ProspectName,
		"stage":             ai.Stage,
		"date_order":        ai.DateOrder,
		"conv_last":         convLastValue,
		"conv_current":      ai.ConvCurrent,
		"human":             ai.Human,
		"niche":             ai.Niche,
		"intro":             ai.Intro,
		"flow_id":           ai.FlowID,
		"current_node_id":   ai.CurrentNodeID,
		"last_node_id":      ai.LastNodeID,
		"waiting_for_reply": ai.WaitingForReply,
		"execution_status":  ai.ExecutionStatus,
		"execution_id":      ai.ExecutionID,
		"updated_at":        ai.UpdatedAt,
	}

	err := r.supabase.From("ai_whatsapp").
		Eq("id_prospect", ai.IDProspect).
		Update(ctx, updateData)

	if err != nil {
		logrus.WithError(err).WithField("id_prospect", ai.IDProspect).Error("Failed to update AI WhatsApp record via Supabase")
		return fmt.Errorf("failed to update AI WhatsApp record: %w", err)
	}

	logrus.WithField("id_prospect", ai.IDProspect).Info("AI WhatsApp record updated successfully via Supabase")
	return nil
}

// UpdateFlowTrackingFields updates flow tracking fields for a prospect
func (r *aiWhatsappRepositorySupabase) UpdateFlowTrackingFields(prospectNum, idDevice string, flowID, currentNodeID, lastNodeID string, waitingForReply int, executionStatus, executionID string) error {
	ctx := context.Background()

	updateData := map[string]interface{}{
		"flow_id":           flowID,
		"current_node_id":   currentNodeID,
		"last_node_id":      lastNodeID,
		"waiting_for_reply": waitingForReply,
		"execution_status":  executionStatus,
		"execution_id":      executionID,
		"updated_at":        time.Now(),
	}

	err := r.supabase.From("ai_whatsapp").
		Eq("prospect_num", prospectNum).
		Eq("id_device", idDevice).
		Update(ctx, updateData)

	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"prospect_num": prospectNum,
			"id_device":    idDevice,
		}).Error("Failed to update flow tracking fields via Supabase")
		return fmt.Errorf("failed to update flow tracking fields: %w", err)
	}

	return nil
}

// UpdateConversationStage updates the conversation stage for a prospect
func (r *aiWhatsappRepositorySupabase) UpdateConversationStage(prospectNum string, stage string) error {
	ctx := context.Background()

	updateData := map[string]interface{}{
		"stage":      stage,
		"updated_at": time.Now(),
	}

	err := r.supabase.From("ai_whatsapp").
		Eq("prospect_num", prospectNum).
		Update(ctx, updateData)

	if err != nil {
		return fmt.Errorf("failed to update conversation stage via Supabase: %w", err)
	}

	return nil
}

// UpdateProspectName updates the prospect name for a conversation
func (r *aiWhatsappRepositorySupabase) UpdateProspectName(prospectNum, idDevice, prospectName string) error {
	ctx := context.Background()

	updateData := map[string]interface{}{
		"prospect_name": prospectName,
		"updated_at":    time.Now(),
	}

	err := r.supabase.From("ai_whatsapp").
		Eq("prospect_num", prospectNum).
		Eq("id_device", idDevice).
		Update(ctx, updateData)

	if err != nil {
		return fmt.Errorf("failed to update prospect name via Supabase: %w", err)
	}

	return nil
}

// UpdateHumanTakeover updates the human takeover status
func (r *aiWhatsappRepositorySupabase) UpdateHumanTakeover(prospectNum string, human int) error {
	ctx := context.Background()

	updateData := map[string]interface{}{
		"human":      human,
		"updated_at": time.Now(),
	}

	err := r.supabase.From("ai_whatsapp").
		Eq("prospect_num", prospectNum).
		Update(ctx, updateData)

	if err != nil {
		logrus.WithError(err).WithField("prospect_num", prospectNum).Error("Failed to update human takeover via Supabase")
		return fmt.Errorf("failed to update human takeover: %w", err)
	}

	return nil
}

// UpdateHumanStatus updates the human status by ID
func (r *aiWhatsappRepositorySupabase) UpdateHumanStatus(idProspect string, human int) error {
	ctx := context.Background()

	updateData := map[string]interface{}{
		"human":      human,
		"updated_at": time.Now(),
	}

	err := r.supabase.From("ai_whatsapp").
		Eq("id_prospect", idProspect).
		Update(ctx, updateData)

	if err != nil {
		logrus.WithError(err).WithField("id_prospect", idProspect).Error("Failed to update human status via Supabase")
		return fmt.Errorf("failed to update human status: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"id_prospect": idProspect,
		"human":       human,
	}).Info("Human status updated successfully via Supabase")

	return nil
}

// UpdateConvCurrent updates the current conversation context
func (r *aiWhatsappRepositorySupabase) UpdateConvCurrent(prospectNum string, convCurrent string) error {
	ctx := context.Background()

	updateData := map[string]interface{}{
		"conv_current": convCurrent,
		"updated_at":   time.Now(),
	}

	err := r.supabase.From("ai_whatsapp").
		Eq("prospect_num", prospectNum).
		Update(ctx, updateData)

	if err != nil {
		logrus.WithError(err).WithField("prospect_num", prospectNum).Error("Failed to update conv_current via Supabase")
		return fmt.Errorf("failed to update conv_current: %w", err)
	}

	logrus.WithField("prospect_num", prospectNum).Debug("conv_current updated successfully via Supabase")
	return nil
}

// UpdateConvLast updates the last conversation field
func (r *aiWhatsappRepositorySupabase) UpdateConvLast(prospectNum string, convLast interface{}) error {
	ctx := context.Background()

	updateData := map[string]interface{}{
		"conv_last":  convLast,
		"updated_at": time.Now(),
	}

	err := r.supabase.From("ai_whatsapp").
		Eq("prospect_num", prospectNum).
		Update(ctx, updateData)

	if err != nil {
		logrus.WithError(err).WithField("prospect_num", prospectNum).Error("Failed to update conv_last via Supabase")
		return fmt.Errorf("failed to update conv_last: %w", err)
	}

	return nil
}

// UpdateWaitingStatus updates the waiting status for an execution
func (r *aiWhatsappRepositorySupabase) UpdateWaitingStatus(executionID string, waitingValue int32) error {
	ctx := context.Background()

	updateData := map[string]interface{}{
		"waiting_for_reply": waitingValue,
		"updated_at":        time.Now(),
	}

	err := r.supabase.From("ai_whatsapp").
		Eq("execution_id", executionID).
		Update(ctx, updateData)

	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"execution_id":  executionID,
			"waiting_value": waitingValue,
		}).Error("Failed to update waiting status via Supabase")
		return fmt.Errorf("failed to update waiting status: %w", err)
	}

	return nil
}

// SaveConversationHistory saves conversation history to the conv_last field
func (r *aiWhatsappRepositorySupabase) SaveConversationHistory(prospectNum, idDevice, userMessage, botResponse, stage, prospectName string) error {
	ctx := context.Background()

	// Check if record exists
	var existing []models.AIWhatsapp
	err := r.supabase.From("ai_whatsapp").
		Select("id_prospect, conv_last").
		Eq("prospect_num", prospectNum).
		Eq("id_device", idDevice).
		Limit(1).
		Execute(ctx, &existing)

	if err != nil {
		return fmt.Errorf("failed to check existing record via Supabase: %w", err)
	}

	// Build conversation history
	var convHistory string
	if len(existing) > 0 && existing[0].ConvLast.Valid {
		convHistory = existing[0].ConvLast.String
	}

	// Add new conversation entries
	if userMessage != "" {
		if convHistory != "" {
			convHistory += "\n"
		}
		convHistory += "USER:" + userMessage
	}
	if botResponse != "" {
		if convHistory != "" {
			convHistory += "\n"
		}
		convHistory += "BOT:" + botResponse
	}

	// Determine conv_last value
	var convLastValue interface{}
	if convHistory == "" {
		convLastValue = nil
	} else {
		convLastValue = convHistory
	}

	now := time.Now()

	if len(existing) > 0 {
		// Update existing record
		updateData := map[string]interface{}{
			"conv_last":     convLastValue,
			"stage":         stage,
			"prospect_name": prospectName,
			"updated_at":    now,
		}

		err = r.supabase.From("ai_whatsapp").
			Eq("prospect_num", prospectNum).
			Eq("id_device", idDevice).
			Update(ctx, updateData)

		if err != nil {
			return fmt.Errorf("failed to update conversation history via Supabase: %w", err)
		}

		logrus.WithField("prospect_num", prospectNum).Info("Conversation history updated successfully via Supabase")
	} else {
		// Create new record
		insertData := map[string]interface{}{
			"prospect_num":  prospectNum,
			"id_device":     idDevice,
			"stage":         stage,
			"conv_last":     convLastValue,
			"prospect_name": prospectName,
			"created_at":    now,
			"updated_at":    now,
		}

		var result models.AIWhatsapp
		err = r.supabase.From("ai_whatsapp").Insert(ctx, insertData, &result)
		if err != nil {
			return fmt.Errorf("failed to create new conversation record via Supabase: %w", err)
		}

		logrus.WithFields(logrus.Fields{
			"prospect_num": prospectNum,
			"id_device":    idDevice,
		}).Info("New conversation record created successfully via Supabase")
	}

	return nil
}

// DeleteAIWhatsapp deletes an AI WhatsApp record by ID
func (r *aiWhatsappRepositorySupabase) DeleteAIWhatsapp(id int) error {
	ctx := context.Background()

	err := r.supabase.From("ai_whatsapp").
		Eq("id_prospect", id).
		Delete(ctx)

	if err != nil {
		logrus.WithError(err).WithField("id", id).Error("Failed to delete AI WhatsApp record via Supabase")
		return fmt.Errorf("failed to delete AI WhatsApp record: %w", err)
	}

	logrus.WithField("id", id).Info("AI WhatsApp record deleted successfully via Supabase")
	return nil
}

// DeleteConversationLogs deletes conversation logs for a prospect
// Note: Since we're using conv_last field, this clears the conversation history
func (r *aiWhatsappRepositorySupabase) DeleteConversationLogs(prospectNum string) error {
	ctx := context.Background()

	updateData := map[string]interface{}{
		"conv_last":  nil,
		"updated_at": time.Now(),
	}

	err := r.supabase.From("ai_whatsapp").
		Eq("prospect_num", prospectNum).
		Update(ctx, updateData)

	if err != nil {
		logrus.WithError(err).WithField("prospect_num", prospectNum).Error("Failed to delete conversation logs via Supabase")
		return fmt.Errorf("failed to delete conversation logs: %w", err)
	}

	logrus.WithField("prospect_num", prospectNum).Info("Conversation logs deleted successfully via Supabase")
	return nil
}

// GetConversationStats retrieves conversation statistics for a device
func (r *aiWhatsappRepositorySupabase) GetConversationStats(idDevice string) (map[string]int, error) {
	ctx := context.Background()

	stats := make(map[string]int)

	// Get all records for the device
	query := r.supabase.From("ai_whatsapp").Select("stage, execution_status")
	if idDevice != "" && idDevice != "all" {
		query = query.Eq("id_device", idDevice)
	}

	var results []models.AIWhatsapp
	err := query.Execute(ctx, &results)
	if err != nil {
		logrus.WithError(err).Error("Failed to get conversation stats via Supabase")
		return nil, fmt.Errorf("failed to get conversation stats: %w", err)
	}

	// Count by stage
	stageCount := make(map[string]int)
	activeCount := 0

	for _, r := range results {
		if r.Stage.Valid && r.Stage.String != "" {
			stageCount[r.Stage.String]++
		}
		if r.ExecutionStatus.Valid && r.ExecutionStatus.String == "active" {
			activeCount++
		}
	}

	stats["total"] = len(results)
	stats["active"] = activeCount
	for stage, count := range stageCount {
		stats["stage_"+stage] = count
	}

	return stats, nil
}

// GetActiveConversationCount returns the count of active conversations
func (r *aiWhatsappRepositorySupabase) GetActiveConversationCount() (int, error) {
	ctx := context.Background()

	var results []models.AIWhatsapp
	err := r.supabase.From("ai_whatsapp").
		Select("id_prospect").
		Eq("execution_status", "active").
		Execute(ctx, &results)

	if err != nil {
		logrus.WithError(err).Error("Failed to get active conversation count via Supabase")
		return 0, fmt.Errorf("failed to get active conversation count: %w", err)
	}

	return len(results), nil
}

// GetConversationsByDateRange retrieves conversations within a date range
func (r *aiWhatsappRepositorySupabase) GetConversationsByDateRange(startDate, endDate time.Time) ([]models.AIWhatsapp, error) {
	ctx := context.Background()

	var results []models.AIWhatsapp
	err := r.supabase.From("ai_whatsapp").
		Select("*").
		Gte("created_at", startDate.Format("2006-01-02")).
		Lte("created_at", endDate.Format("2006-01-02")).
		Order("created_at", false).
		Execute(ctx, &results)

	if err != nil {
		logrus.WithError(err).Error("Failed to get conversations by date range via Supabase")
		return nil, fmt.Errorf("failed to get conversations by date range: %w", err)
	}

	return results, nil
}

// TryAcquireSession attempts to create a session lock for a prospect/device pair
func (r *aiWhatsappRepositorySupabase) TryAcquireSession(phoneNumber, deviceID string) (bool, error) {
	ctx := context.Background()

	data := map[string]interface{}{
		"phone_number": phoneNumber,
		"device_id":    deviceID,
		"timestamp":    time.Now().Format(time.RFC3339Nano),
	}

	var result map[string]interface{}
	err := r.supabase.From("ai_session_locks").Insert(ctx, data, &result)
	if err != nil {
		// Check if it's a duplicate key error
		if strings.Contains(fmt.Sprintf("%v", err), "duplicate key") || strings.Contains(fmt.Sprintf("%v", err), "unique constraint") {
			return false, nil
		}
		return false, fmt.Errorf("failed to acquire session lock via Supabase: %w", err)
	}

	return true, nil
}

// ReleaseSession removes the session lock for a prospect/device pair
func (r *aiWhatsappRepositorySupabase) ReleaseSession(phoneNumber, deviceID string) error {
	ctx := context.Background()

	err := r.supabase.From("ai_session_locks").
		Eq("phone_number", phoneNumber).
		Eq("device_id", deviceID).
		Delete(ctx)

	if err != nil {
		return fmt.Errorf("failed to release session lock via Supabase: %w", err)
	}

	return nil
}
// GetAllAIWhatsappData and GetAnalyticsData methods

// GetAllAIWhatsappData retrieves all AI WhatsApp conversation records with pagination and filtering via Supabase REST API
func (r *aiWhatsappRepositorySupabase) GetAllAIWhatsappData(limit, offset int, deviceFilter, stageFilter, search string, userID string, startDate, endDate *time.Time) ([]models.AIWhatsapp, int, error) {
	ctx := context.Background()

	logrus.WithFields(logrus.Fields{
		"limit":        limit,
		"offset":       offset,
		"deviceFilter": deviceFilter,
		"stageFilter":  stageFilter,
		"search":       search,
		"userID":       userID,
	}).Info("GetAllAIWhatsappData called via Supabase")

	// Build base query
	query := r.supabase.From("ai_whatsapp").Select("*")

	// Apply device filter (supports single or comma-separated multiple devices)
	if deviceFilter != "" {
		if strings.Contains(deviceFilter, ",") {
			deviceIDs := utils.SplitAndTrim(deviceFilter, ",")
			deviceInterfaces := make([]interface{}, len(deviceIDs))
			for i, d := range deviceIDs {
				deviceInterfaces[i] = d
			}
			query = query.In("id_device", deviceInterfaces)
			logrus.WithField("deviceIDs", deviceIDs).Info("Added multiple device filter via Supabase")
		} else {
			query = query.Eq("id_device", deviceFilter)
		}
	}

	// Apply stage filter
	if stageFilter != "" {
		query = query.Eq("stage", stageFilter)
	}

	// Apply search filter
	if search != "" {
		query = query.Ilike("prospect_num", "%"+search+"%")
	}

	// Apply date range filter
	if startDate != nil && endDate != nil {
		query = query.Gte("created_at", startDate.Format("2006-01-02")).
			Lte("created_at", endDate.Format("2006-01-02"))
		logrus.WithFields(logrus.Fields{
			"startDate": startDate.Format("2006-01-02"),
			"endDate":   endDate.Format("2006-01-02"),
		}).Info("Added date range filter via Supabase")
	}

	// Get count first
	countQuery := r.supabase.From("ai_whatsapp").Select("id_prospect")
	if deviceFilter != "" {
		if strings.Contains(deviceFilter, ",") {
			deviceIDs := utils.SplitAndTrim(deviceFilter, ",")
			deviceInterfaces := make([]interface{}, len(deviceIDs))
			for i, d := range deviceIDs {
				deviceInterfaces[i] = d
			}
			countQuery = countQuery.In("id_device", deviceInterfaces)
		} else {
			countQuery = countQuery.Eq("id_device", deviceFilter)
		}
	}
	if stageFilter != "" {
		countQuery = countQuery.Eq("stage", stageFilter)
	}
	if search != "" {
		countQuery = countQuery.Ilike("prospect_num", "%"+search+"%")
	}
	if startDate != nil && endDate != nil {
		countQuery = countQuery.Gte("created_at", startDate.Format("2006-01-02")).
			Lte("created_at", endDate.Format("2006-01-02"))
	}

	var countResults []models.AIWhatsapp
	err := countQuery.Execute(ctx, &countResults)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get count via Supabase - returning empty result")
		return []models.AIWhatsapp{}, 0, nil
	}

	total := len(countResults)

	if total == 0 {
		logrus.Info("No AI WhatsApp data found for given filters via Supabase - returning empty result")
		return []models.AIWhatsapp{}, 0, nil
	}

	// Add ORDER BY and pagination
	query = query.Order("updated_at", false).Limit(limit).Offset(offset)

	// Execute main query
	var conversations []models.AIWhatsapp
	err = query.Execute(ctx, &conversations)
	if err != nil {
		logrus.WithError(err).Warn("Failed to execute main query via Supabase - returning empty result")
		return []models.AIWhatsapp{}, 0, nil
	}

	logrus.WithFields(logrus.Fields{
		"total_records":    total,
		"returned_records": len(conversations),
		"limit":            limit,
		"offset":           offset,
	}).Info("Successfully retrieved AI WhatsApp data via Supabase")

	return conversations, total, nil
}

// GetAnalyticsData retrieves analytics data from ai_whatsapp table with date filtering via Supabase REST API
func (r *aiWhatsappRepositorySupabase) GetAnalyticsData(startDate, endDate time.Time, idDevice string, userID string) (map[string]interface{}, error) {
	ctx := context.Background()

	logrus.WithFields(logrus.Fields{
		"startDate": startDate.Format("2006-01-02"),
		"endDate":   endDate.Format("2006-01-02"),
		"idDevice":  idDevice,
		"userID":    userID,
	}).Info("GetAnalyticsData called via Supabase")

	analytics := map[string]interface{}{
		"total_conversations":      0,
		"active_conversations":     0,
		"completed_conversations":  0,
		"stage_breakdown":          make(map[string]int),
		"niche_breakdown":          make(map[string]int),
		"daily_conversations":      []map[string]interface{}{},
	}

	// Build base query with date range
	query := r.supabase.From("ai_whatsapp").Select("*").
		Gte("created_at", startDate.Format("2006-01-02")).
		Lte("created_at", endDate.Format("2006-01-02"))

	// Apply device filter if specified
	if idDevice != "" && idDevice != "all" {
		query = query.Eq("id_device", idDevice)
	}

	// Fetch all matching records
	var results []models.AIWhatsapp
	err := query.Execute(ctx, &results)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch analytics data via Supabase")
		return nil, fmt.Errorf("failed to fetch analytics data: %w", err)
	}

	// Calculate statistics
	analytics["total_conversations"] = len(results)

	stageBreakdown := make(map[string]int)
	nicheBreakdown := make(map[string]int)
	dailyCount := make(map[string]int)
	activeCount := 0
	completedCount := 0

	for _, record := range results {
		// Count by stage
		if record.Stage.Valid && record.Stage.String != "" {
			stageBreakdown[record.Stage.String]++
		}

		// Count by niche
		if record.Niche != "" {
			nicheBreakdown[record.Niche]++
		}

		// Count by date
		dateKey := record.CreatedAt.Format("2006-01-02")
		dailyCount[dateKey]++

		// Count active vs completed
		if record.ExecutionStatus.Valid && record.ExecutionStatus.String == "active" {
			activeCount++
		} else if record.ExecutionStatus.Valid && record.ExecutionStatus.String == "completed" {
			completedCount++
		}
	}

	analytics["active_conversations"] = activeCount
	analytics["completed_conversations"] = completedCount
	analytics["stage_breakdown"] = stageBreakdown
	analytics["niche_breakdown"] = nicheBreakdown

	// Convert daily counts to array format
	dailyData := []map[string]interface{}{}
	for date, count := range dailyCount {
		dailyData = append(dailyData, map[string]interface{}{
			"date":  date,
			"count": count,
		})
	}
	analytics["daily_conversations"] = dailyData

	logrus.WithFields(logrus.Fields{
		"total":     len(results),
		"active":    activeCount,
		"completed": completedCount,
		"stages":    len(stageBreakdown),
	}).Info("Analytics data retrieved successfully via Supabase")

	return analytics, nil
}
