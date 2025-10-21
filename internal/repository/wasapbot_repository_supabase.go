package repository

import (
	"context"
	"fmt"
	"time"

	"nodepath-chat/internal/database"
	"nodepath-chat/internal/models"
	"nodepath-chat/internal/utils"

	"github.com/sirupsen/logrus"
)

// wasapBotRepositorySupabase implements WasapBotRepository using Supabase SDK
type wasapBotRepositorySupabase struct {
	supabase *database.SupabaseSDK
}

// NewWasapBotRepositorySupabase creates a Supabase-based wasapBot repository
func NewWasapBotRepositorySupabase(supabase *database.SupabaseSDK) WasapBotRepository {
	return &wasapBotRepositorySupabase{supabase: supabase}
}

// GetByProspectAndDevice retrieves a wasapBot record by prospect number and device ID
func (r *wasapBotRepositorySupabase) GetByProspectAndDevice(prospectNum, deviceID string) (*models.WasapBot, error) {
	ctx := context.Background()

	var results []models.WasapBot
	err := r.supabase.From("wasapBot").
		Select("*").
		Eq("prospect_num", prospectNum).
		Eq("id_device", deviceID).
		Limit(1).
		Execute(ctx, &results)

	if err != nil {
		return nil, fmt.Errorf("failed to get wasapBot by prospect and device via Supabase: %w", err)
	}

	if len(results) == 0 {
		return nil, nil
	}

	return &results[0], nil
}

// GetActiveExecution retrieves an active execution for a prospect and device ID
func (r *wasapBotRepositorySupabase) GetActiveExecution(prospectNum, deviceID string) (*models.WasapBot, error) {
	ctx := context.Background()

	var results []models.WasapBot
	err := r.supabase.From("wasapBot").
		Select("*").
		Eq("prospect_num", prospectNum).
		Eq("id_device", deviceID).
		Eq("execution_status", "active").
		Limit(1).
		Execute(ctx, &results)

	if err != nil {
		return nil, fmt.Errorf("failed to get active execution via Supabase: %w", err)
	}

	if len(results) == 0 {
		return nil, nil
	}

	return &results[0], nil
}

// GetByExecutionID retrieves a wasapBot record by execution ID
func (r *wasapBotRepositorySupabase) GetByExecutionID(executionID string) (*models.WasapBot, error) {
	ctx := context.Background()

	var results []models.WasapBot
	err := r.supabase.From("wasapBot").
		Select("*").
		Eq("execution_id", executionID).
		Limit(1).
		Execute(ctx, &results)

	if err != nil {
		return nil, fmt.Errorf("failed to get wasapBot by execution ID via Supabase: %w", err)
	}

	if len(results) == 0 {
		return nil, nil
	}

	return &results[0], nil
}

// Create creates a new wasapBot record
func (r *wasapBotRepositorySupabase) Create(wasapBot *models.WasapBot) error {
	ctx := context.Background()

	var result models.WasapBot
	err := r.supabase.From("wasapBot").Insert(ctx, wasapBot, &result)
	if err != nil {
		return fmt.Errorf("failed to create wasapBot record via Supabase: %w", err)
	}

	wasapBot.IDProspect = result.IDProspect
	return nil
}

// Update updates an existing wasapBot record
func (r *wasapBotRepositorySupabase) Update(wasapBot *models.WasapBot) error {
	ctx := context.Background()

	err := r.supabase.From("wasapBot").
		Eq("id_prospect", wasapBot.IDProspect).
		Update(ctx, wasapBot)

	if err != nil {
		return fmt.Errorf("failed to update wasapBot record via Supabase: %w", err)
	}

	return nil
}

// UpdateExecutionStatus updates the execution status
func (r *wasapBotRepositorySupabase) UpdateExecutionStatus(executionID, status string) error {
	ctx := context.Background()

	updateData := map[string]interface{}{
		"execution_status": status,
	}

	err := r.supabase.From("wasapBot").
		Eq("execution_id", executionID).
		Update(ctx, updateData)

	if err != nil {
		return fmt.Errorf("failed to update execution status via Supabase: %w", err)
	}

	return nil
}

// UpdateCurrentNode updates the current node ID
func (r *wasapBotRepositorySupabase) UpdateCurrentNode(executionID, nodeID string) error {
	ctx := context.Background()

	updateData := map[string]interface{}{
		"current_node_id": nodeID,
	}

	err := r.supabase.From("wasapBot").
		Eq("execution_id", executionID).
		Update(ctx, updateData)

	if err != nil {
		return fmt.Errorf("failed to update current node via Supabase: %w", err)
	}

	return nil
}

// UpdateWaitingStatus updates the waiting status for an execution
func (r *wasapBotRepositorySupabase) UpdateWaitingStatus(executionID string, waitingValue int) error {
	ctx := context.Background()

	updateData := map[string]interface{}{
		"waiting_for_reply": waitingValue,
		"date_last":         time.Now().Format("2006-01-02 15:04:05"),
	}

	err := r.supabase.From("wasapBot").
		Eq("execution_id", executionID).
		Update(ctx, updateData)

	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"execution_id":  executionID,
			"waiting_value": waitingValue,
		}).Error("Failed to update waiting status in wasapBot via Supabase")
		return fmt.Errorf("failed to update waiting status: %w", err)
	}

	return nil
}

// SaveConversationHistory saves conversation history to conv_last field
func (r *wasapBotRepositorySupabase) SaveConversationHistory(prospectNum, deviceID, userMessage, botResponse, stage, nama string) error {
	ctx := context.Background()

	// Check if record exists
	var existing []models.WasapBot
	err := r.supabase.From("wasapBot").
		Select("id_prospect, conv_last").
		Eq("prospect_num", prospectNum).
		Eq("id_device", deviceID).
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

	now := time.Now().Format("2006-01-02 15:04:05")

	if len(existing) > 0 {
		// Update existing record
		updateData := map[string]interface{}{
			"conv_last":  convLastValue,
			"stage":      stage,
			"nama":       nama,
			"date_last":  now,
		}

		err = r.supabase.From("wasapBot").
			Eq("prospect_num", prospectNum).
			Eq("id_device", deviceID).
			Update(ctx, updateData)

		if err != nil {
			return fmt.Errorf("failed to update conversation history via Supabase: %w", err)
		}

		logrus.WithFields(logrus.Fields{
			"prospect_num": prospectNum,
		}).Info("WasapBot conversation history updated successfully via Supabase")
	} else {
		// Create new record
		insertData := map[string]interface{}{
			"prospect_num":      prospectNum,
			"id_device":         deviceID,
			"stage":             stage,
			"conv_last":         convLastValue,
			"nama":              nama,
			"date_start":        now,
			"date_last":         now,
			"status":            "Prospek",
			"waiting_for_reply": 0,
		}

		var result models.WasapBot
		err = r.supabase.From("wasapBot").Insert(ctx, insertData, &result)
		if err != nil {
			return fmt.Errorf("failed to create new conversation record via Supabase: %w", err)
		}

		logrus.WithFields(logrus.Fields{
			"prospect_num": prospectNum,
			"id_device":    deviceID,
		}).Info("New WasapBot conversation record created successfully via Supabase")
	}

	return nil
}

// TryAcquireSession attempts to create a session lock for a prospect/device pair
func (r *wasapBotRepositorySupabase) TryAcquireSession(prospectNum, deviceID string) (bool, error) {
	ctx := context.Background()

	data := map[string]interface{}{
		"id_prospect": prospectNum,
		"id_device":   deviceID,
		"timestamp":   time.Now().Format(time.RFC3339Nano),
	}

	var result map[string]interface{}
	err := r.supabase.From("wasapBot_session").Insert(ctx, data, &result)
	if err != nil {
		// Check if it's a duplicate key error
		if fmt.Sprintf("%v", err) == "duplicate key value violates unique constraint" {
			return false, nil
		}
		return false, fmt.Errorf("failed to acquire WasapBot session lock via Supabase: %w", err)
	}

	return true, nil
}

// ReleaseSession removes the session lock for a prospect/device pair
func (r *wasapBotRepositorySupabase) ReleaseSession(prospectNum, deviceID string) error {
	ctx := context.Background()

	err := r.supabase.From("wasapBot_session").
		Eq("id_prospect", prospectNum).
		Eq("id_device", deviceID).
		Delete(ctx)

	if err != nil {
		return fmt.Errorf("failed to release WasapBot session lock via Supabase: %w", err)
	}

	return nil
}

// Delete deletes a WasapBot record by ID
func (r *wasapBotRepositorySupabase) Delete(idProspect int) error {
	ctx := context.Background()

	err := r.supabase.From("wasapBot").
		Eq("id_prospect", idProspect).
		Delete(ctx)

	if err != nil {
		logrus.WithError(err).WithField("id_prospect", idProspect).Error("Failed to delete WasapBot record via Supabase")
		return fmt.Errorf("failed to delete WasapBot record: %w", err)
	}

	logrus.WithField("id_prospect", idProspect).Info("WasapBot record deleted successfully via Supabase")
	return nil
}

// GetAllWasapBotData retrieves all WasapBot data with filters
func (r *wasapBotRepositorySupabase) GetAllWasapBotData(limit, offset int, deviceFilter, stageFilter, statusFilter, search string, userID string) ([]map[string]interface{}, int, error) {
	ctx := context.Background()

	logrus.WithFields(logrus.Fields{
		"limit":        limit,
		"offset":       offset,
		"deviceFilter": deviceFilter,
		"stageFilter":  stageFilter,
		"statusFilter": statusFilter,
		"search":       search,
		"userID":       userID,
	}).Info("GetAllWasapBotData called via Supabase")

	// Build query - start with all records
	query := r.supabase.From("wasapBot").Select("id_prospect, prospect_num, nama, stage, date_last, id_device, niche, status, alamat, pakej, cara_bayaran, tarikh_gaji, current_node_id, no_fon")

	// Apply device filter for multiple devices
	if deviceFilter != "" && deviceFilter != "all" {
		devices := utils.SplitAndTrim(deviceFilter, ",")
		if len(devices) > 0 {
			// For Supabase, we need to convert to interface{} slice
			deviceInterfaces := make([]interface{}, len(devices))
			for i, d := range devices {
				deviceInterfaces[i] = d
			}
			query = query.In("id_device", deviceInterfaces)
			logrus.WithField("device_filter_applied", devices).Info("Applying device filter via Supabase")
		}
	}

	// Apply stage filter
	if stageFilter != "" && stageFilter != "all" {
		query = query.Eq("stage", stageFilter)
	}

	// Apply status filter
	if statusFilter != "" && statusFilter != "all" {
		query = query.Eq("status", statusFilter)
	}

	// Apply search filter
	if search != "" {
		// Note: Supabase doesn't support OR queries easily, so we'll need to fetch and filter
		// For now, we'll just search by prospect_num (can be enhanced later)
		query = query.Ilike("prospect_num", "%"+search+"%")
	}

	// Get count first (fetch all matching records to count them)
	var countResults []models.WasapBot
	countQuery := r.supabase.From("wasapBot").Select("id_prospect")

	if deviceFilter != "" && deviceFilter != "all" {
		devices := utils.SplitAndTrim(deviceFilter, ",")
		if len(devices) > 0 {
			deviceInterfaces := make([]interface{}, len(devices))
			for i, d := range devices {
				deviceInterfaces[i] = d
			}
			countQuery = countQuery.In("id_device", deviceInterfaces)
		}
	}
	if stageFilter != "" && stageFilter != "all" {
		countQuery = countQuery.Eq("stage", stageFilter)
	}
	if statusFilter != "" && statusFilter != "all" {
		countQuery = countQuery.Eq("status", statusFilter)
	}
	if search != "" {
		countQuery = countQuery.Ilike("prospect_num", "%"+search+"%")
	}

	err := countQuery.Execute(ctx, &countResults)
	if err != nil {
		logrus.WithError(err).Error("Failed to get count via Supabase")
		return nil, 0, fmt.Errorf("failed to get count: %w", err)
	}

	total := len(countResults)
	logrus.WithField("total_count", total).Info("Total records found via Supabase")

	// Add order and pagination
	query = query.Order("date_last", false).Limit(limit).Offset(offset)

	// Execute main query
	var results []map[string]interface{}
	err = query.Execute(ctx, &results)
	if err != nil {
		logrus.WithError(err).Error("Failed to query wasapBot data via Supabase")
		return nil, 0, fmt.Errorf("failed to query wasapBot data: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"results_count": len(results),
	}).Info("Query completed via Supabase")

	return results, total, nil
}

// GetAllWasapBotDataWithDates retrieves all WasapBot data with filters including date range
func (r *wasapBotRepositorySupabase) GetAllWasapBotDataWithDates(limit, offset int, deviceFilter, stageFilter, statusFilter, search, dateFrom, dateTo string, userID string) ([]map[string]interface{}, int, error) {
	ctx := context.Background()

	logrus.WithFields(logrus.Fields{
		"limit":        limit,
		"offset":       offset,
		"deviceFilter": deviceFilter,
		"stageFilter":  stageFilter,
		"statusFilter": statusFilter,
		"search":       search,
		"dateFrom":     dateFrom,
		"dateTo":       dateTo,
		"userID":       userID,
	}).Info("GetAllWasapBotDataWithDates called via Supabase")

	// Build query
	query := r.supabase.From("wasapBot").Select("id_prospect, prospect_num, nama, stage, date_last, date_start, id_device, niche, status, alamat, pakej, cara_bayaran, tarikh_gaji, current_node_id, no_fon")

	// Apply date filters
	if dateFrom != "" {
		query = query.Gte("date_start", dateFrom)
		logrus.WithField("date_from_applied", dateFrom).Info("Applying date from filter via Supabase")
	}
	if dateTo != "" {
		query = query.Lte("date_start", dateTo)
		logrus.WithField("date_to_applied", dateTo).Info("Applying date to filter via Supabase")
	}

	// Apply other filters (same as GetAllWasapBotData)
	if deviceFilter != "" && deviceFilter != "all" {
		devices := utils.SplitAndTrim(deviceFilter, ",")
		if len(devices) > 0 {
			deviceInterfaces := make([]interface{}, len(devices))
			for i, d := range devices {
				deviceInterfaces[i] = d
			}
			query = query.In("id_device", deviceInterfaces)
		}
	}

	if stageFilter != "" && stageFilter != "all" {
		if stageFilter == "No Stage" {
			query = query.IsNull("stage")
		} else {
			query = query.Eq("stage", stageFilter)
		}
	}

	if statusFilter != "" && statusFilter != "all" {
		query = query.Eq("status", statusFilter)
	}

	if search != "" {
		query = query.Ilike("prospect_num", "%"+search+"%")
	}

	// Get count
	countQuery := r.supabase.From("wasapBot").Select("id_prospect")
	if dateFrom != "" {
		countQuery = countQuery.Gte("date_start", dateFrom)
	}
	if dateTo != "" {
		countQuery = countQuery.Lte("date_start", dateTo)
	}
	if deviceFilter != "" && deviceFilter != "all" {
		devices := utils.SplitAndTrim(deviceFilter, ",")
		if len(devices) > 0 {
			deviceInterfaces := make([]interface{}, len(devices))
			for i, d := range devices {
				deviceInterfaces[i] = d
			}
			countQuery = countQuery.In("id_device", deviceInterfaces)
		}
	}
	if stageFilter != "" && stageFilter != "all" {
		if stageFilter == "No Stage" {
			countQuery = countQuery.IsNull("stage")
		} else {
			countQuery = countQuery.Eq("stage", stageFilter)
		}
	}
	if statusFilter != "" && statusFilter != "all" {
		countQuery = countQuery.Eq("status", statusFilter)
	}
	if search != "" {
		countQuery = countQuery.Ilike("prospect_num", "%"+search+"%")
	}

	var countResults []models.WasapBot
	err := countQuery.Execute(ctx, &countResults)
	if err != nil {
		logrus.WithError(err).Error("Failed to get count via Supabase")
		return nil, 0, fmt.Errorf("failed to get count: %w", err)
	}

	total := len(countResults)
	logrus.WithField("total_count", total).Info("Total records found via Supabase")

	// Add order and pagination
	query = query.Order("date_last", false).Limit(limit).Offset(offset)

	// Execute
	var results []map[string]interface{}
	err = query.Execute(ctx, &results)
	if err != nil {
		logrus.WithError(err).Error("Failed to query wasapBot data via Supabase")
		return nil, 0, fmt.Errorf("failed to query wasapBot data: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"results_count": len(results),
	}).Info("Query completed via Supabase")

	return results, total, nil
}

// GetWasapBotStats retrieves WasapBot statistics
func (r *wasapBotRepositorySupabase) GetWasapBotStats(deviceFilter string, userID string) (map[string]interface{}, error) {
	ctx := context.Background()

	stats := map[string]interface{}{
		"totalProspects":      0,
		"activeExecutions":    0,
		"completedExecutions": 0,
		"uniqueSchools":       0,
		"uniquePackages":      0,
		"totalWithPhone":      0,
	}

	// Build base query with device filter
	baseQuery := r.supabase.From("wasapBot").Select("prospect_num")
	if deviceFilter != "" && deviceFilter != "all" {
		baseQuery = baseQuery.Eq("instance", deviceFilter)
	}

	// Total prospects (distinct count - fetch all and count unique)
	var allProspects []models.WasapBot
	err := baseQuery.Execute(ctx, &allProspects)
	if err == nil {
		// Count unique prospect_num
		uniqueMap := make(map[string]bool)
		for _, p := range allProspects {
			uniqueMap[p.ProspectNum.String] = true
		}
		stats["totalProspects"] = len(uniqueMap)
	}

	// Active executions
	activeQuery := r.supabase.From("wasapBot").Select("id_prospect")
	if deviceFilter != "" && deviceFilter != "all" {
		activeQuery = activeQuery.Eq("instance", deviceFilter)
	}
	activeQuery = activeQuery.Eq("execution_status", "active")

	var activeResults []models.WasapBot
	err = activeQuery.Execute(ctx, &activeResults)
	if err == nil {
		stats["activeExecutions"] = len(activeResults)
	}

	// Completed executions
	completedQuery := r.supabase.From("wasapBot").Select("id_prospect")
	if deviceFilter != "" && deviceFilter != "all" {
		completedQuery = completedQuery.Eq("instance", deviceFilter)
	}
	completedQuery = completedQuery.Eq("status", "Customer")

	var completedResults []models.WasapBot
	err = completedQuery.Execute(ctx, &completedResults)
	if err == nil {
		stats["completedExecutions"] = len(completedResults)
	}

	// Total with phone
	phoneQuery := r.supabase.From("wasapBot").Select("id_prospect, no_fon")
	if deviceFilter != "" && deviceFilter != "all" {
		phoneQuery = phoneQuery.Eq("instance", deviceFilter)
	}
	phoneQuery = phoneQuery.IsNotNull("no_fon")

	var phoneResults []models.WasapBot
	err = phoneQuery.Execute(ctx, &phoneResults)
	if err == nil {
		count := 0
		for _, p := range phoneResults {
			if p.NoFon.Valid && p.NoFon.String != "" {
				count++
			}
		}
		stats["totalWithPhone"] = count
	}

	return stats, nil
}

// GetWasapBotStatsWithDates retrieves WasapBot statistics with date filtering
func (r *wasapBotRepositorySupabase) GetWasapBotStatsWithDates(deviceFilter, dateFrom, dateTo string, userID string) (map[string]interface{}, error) {
	ctx := context.Background()

	stats := map[string]interface{}{
		"totalProspects":      0,
		"activeExecutions":    0,
		"completedExecutions": 0,
		"uniqueSchools":       0,
		"uniquePackages":      0,
		"totalWithPhone":      0,
		"stageBreakdown":      make(map[string]int),
		"daily_data":          []map[string]interface{}{},
	}

	// Build base query with filters
	baseQuery := r.supabase.From("wasapBot").Select("prospect_num")
	if dateFrom != "" {
		baseQuery = baseQuery.Gte("date_start", dateFrom)
	}
	if dateTo != "" {
		baseQuery = baseQuery.Lte("date_start", dateTo)
	}
	if deviceFilter != "" && deviceFilter != "all" {
		devices := utils.SplitAndTrim(deviceFilter, ",")
		if len(devices) > 0 {
			deviceInterfaces := make([]interface{}, len(devices))
			for i, d := range devices {
				deviceInterfaces[i] = d
			}
			baseQuery = baseQuery.In("id_device", deviceInterfaces)
		}
	}

	// Total prospects
	var allProspects []models.WasapBot
	err := baseQuery.Execute(ctx, &allProspects)
	if err == nil {
		uniqueMap := make(map[string]bool)
		for _, p := range allProspects {
			uniqueMap[p.ProspectNum.String] = true
		}
		stats["totalProspects"] = len(uniqueMap)
	}

	// Active executions (current_node_id is NOT 'end')
	activeQuery := r.supabase.From("wasapBot").Select("id_prospect")
	if dateFrom != "" {
		activeQuery = activeQuery.Gte("date_start", dateFrom)
	}
	if dateTo != "" {
		activeQuery = activeQuery.Lte("date_start", dateTo)
	}
	if deviceFilter != "" && deviceFilter != "all" {
		devices := utils.SplitAndTrim(deviceFilter, ",")
		if len(devices) > 0 {
			deviceInterfaces := make([]interface{}, len(devices))
			for i, d := range devices {
				deviceInterfaces[i] = d
			}
			activeQuery = activeQuery.In("id_device", deviceInterfaces)
		}
	}
	activeQuery = activeQuery.Neq("current_node_id", "end")

	var activeResults []models.WasapBot
	err = activeQuery.Execute(ctx, &activeResults)
	if err == nil {
		stats["activeExecutions"] = len(activeResults)
	}

	// Completed executions
	completedQuery := r.supabase.From("wasapBot").Select("id_prospect")
	if dateFrom != "" {
		completedQuery = completedQuery.Gte("date_start", dateFrom)
	}
	if dateTo != "" {
		completedQuery = completedQuery.Lte("date_start", dateTo)
	}
	if deviceFilter != "" && deviceFilter != "all" {
		devices := utils.SplitAndTrim(deviceFilter, ",")
		if len(devices) > 0 {
			deviceInterfaces := make([]interface{}, len(devices))
			for i, d := range devices {
				deviceInterfaces[i] = d
			}
			completedQuery = completedQuery.In("id_device", deviceInterfaces)
		}
	}
	completedQuery = completedQuery.Eq("current_node_id", "end")

	var completedResults []models.WasapBot
	err = completedQuery.Execute(ctx, &completedResults)
	if err == nil {
		stats["completedExecutions"] = len(completedResults)
	}

	// Total with phone
	phoneQuery := r.supabase.From("wasapBot").Select("id_prospect, no_fon")
	if dateFrom != "" {
		phoneQuery = phoneQuery.Gte("date_start", dateFrom)
	}
	if dateTo != "" {
		phoneQuery = phoneQuery.Lte("date_start", dateTo)
	}
	if deviceFilter != "" && deviceFilter != "all" {
		devices := utils.SplitAndTrim(deviceFilter, ",")
		if len(devices) > 0 {
			deviceInterfaces := make([]interface{}, len(devices))
			for i, d := range devices {
				deviceInterfaces[i] = d
			}
			phoneQuery = phoneQuery.In("id_device", deviceInterfaces)
		}
	}
	phoneQuery = phoneQuery.IsNotNull("no_fon")

	var phoneResults []models.WasapBot
	err = phoneQuery.Execute(ctx, &phoneResults)
	if err == nil {
		count := 0
		for _, p := range phoneResults {
			if p.NoFon.Valid && p.NoFon.String != "" {
				count++
			}
		}
		stats["totalWithPhone"] = count
	}

	// Stage breakdown
	stageQuery := r.supabase.From("wasapBot").Select("stage")
	if dateFrom != "" {
		stageQuery = stageQuery.Gte("date_start", dateFrom)
	}
	if dateTo != "" {
		stageQuery = stageQuery.Lte("date_start", dateTo)
	}
	if deviceFilter != "" && deviceFilter != "all" {
		devices := utils.SplitAndTrim(deviceFilter, ",")
		if len(devices) > 0 {
			deviceInterfaces := make([]interface{}, len(devices))
			for i, d := range devices {
				deviceInterfaces[i] = d
			}
			stageQuery = stageQuery.In("id_device", deviceInterfaces)
		}
	}

	var stageResults []models.WasapBot
	err = stageQuery.Execute(ctx, &stageResults)
	if err == nil {
		stageBreakdown := make(map[string]int)
		for _, r := range stageResults {
			stageName := "No Stage"
			if r.Stage.Valid && r.Stage.String != "" {
				stageName = r.Stage.String
			}
			stageBreakdown[stageName]++
		}
		stats["stageBreakdown"] = stageBreakdown
	}

	return stats, nil
}
