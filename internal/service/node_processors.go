package service

import (
	"chatbot-automation/internal/models"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// StartNodeProcessor processes start nodes
type StartNodeProcessor struct{}

func (p *StartNodeProcessor) ProcessNode(ctx *models.ExecutionContext, node *models.FlowNode, edges []models.FlowEdge) (*models.ExecutionResult, error) {
	// Find next node
	nextNodeID := ""
	for _, edge := range edges {
		if edge.Source == node.ID {
			nextNodeID = edge.Target
			break
		}
	}

	if nextNodeID == "" {
		return &models.ExecutionResult{
			Success:       false,
			Message:       "No outgoing edge from start node",
			CompletedFlow: true,
		}, nil
	}

	return &models.ExecutionResult{
		Success:     true,
		Message:     "Flow started",
		NextNodeID:  nextNodeID,
		ShouldReply: false,
		Variables:   ctx.Variables,
	}, nil
}

func (p *StartNodeProcessor) GetNodeType() models.NodeType {
	return models.NodeTypeStart
}

// MessageNodeProcessor processes message nodes
type MessageNodeProcessor struct{}

func (p *MessageNodeProcessor) ProcessNode(ctx *models.ExecutionContext, node *models.FlowNode, edges []models.FlowEdge) (*models.ExecutionResult, error) {
	// Get message from node data
	message, ok := node.Data["message"].(string)
	if !ok || message == "" {
		message = "Hello!"
	}

	// Replace variables in message
	message = p.replaceVariables(message, ctx.Variables)

	// Find next node
	nextNodeID := ""
	for _, edge := range edges {
		if edge.Source == node.ID {
			nextNodeID = edge.Target
			break
		}
	}

	return &models.ExecutionResult{
		Success:       true,
		Message:       "Message sent",
		NextNodeID:    nextNodeID,
		Response:      message,
		ShouldReply:   true,
		Variables:     ctx.Variables,
		CompletedFlow: nextNodeID == "",
	}, nil
}

func (p *MessageNodeProcessor) GetNodeType() models.NodeType {
	return models.NodeTypeMessage
}

func (p *MessageNodeProcessor) replaceVariables(message string, variables map[string]interface{}) string {
	if variables == nil {
		return message
	}

	result := message
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		valueStr := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, valueStr)
	}

	return result
}

// AINodeProcessor processes AI nodes
type AINodeProcessor struct {
	aiService *AIService
}

func (p *AINodeProcessor) ProcessNode(ctx *models.ExecutionContext, node *models.FlowNode, edges []models.FlowEdge) (*models.ExecutionResult, error) {
	// Get AI configuration from node data
	provider, _ := node.Data["provider"].(string)
	if provider == "" {
		provider = "openai"
	}

	model, _ := node.Data["model"].(string)
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	systemPrompt, _ := node.Data["systemPrompt"].(string)
	prompt, _ := node.Data["prompt"].(string)

	// Build messages
	messages := make([]models.AIMessage, 0)

	// Add context if available
	if prompt != "" {
		messages = append(messages, models.AIMessage{
			Role:    "system",
			Content: prompt,
		})
	}

	// Add user message
	if ctx.UserMessage != "" {
		messages = append(messages, models.AIMessage{
			Role:    "user",
			Content: ctx.UserMessage,
		})
	}

	// Generate AI response
	aiReq := &models.AICompletionRequest{
		Provider:  models.AIProvider(provider),
		Model:     models.AIModel(model),
		Messages:  messages,
		DeviceID:  ctx.DeviceID,
	}

	if systemPrompt != "" {
		aiReq.SystemPrompt = &systemPrompt
	}

	// Note: This would need a userID - for now we'll skip actual AI call
	// In production, this should be called with proper authentication
	response := "AI response placeholder"

	// Store AI response in variables
	if ctx.Variables == nil {
		ctx.Variables = make(map[string]interface{})
	}
	ctx.Variables["ai_response"] = response
	ctx.Variables["last_user_message"] = ctx.UserMessage

	// Find next node
	nextNodeID := ""
	for _, edge := range edges {
		if edge.Source == node.ID {
			nextNodeID = edge.Target
			break
		}
	}

	return &models.ExecutionResult{
		Success:       true,
		Message:       "AI response generated",
		NextNodeID:    nextNodeID,
		Response:      response,
		ShouldReply:   true,
		Variables:     ctx.Variables,
		CompletedFlow: nextNodeID == "",
	}, nil
}

func (p *AINodeProcessor) GetNodeType() models.NodeType {
	return models.NodeTypeAI
}

// ConditionNodeProcessor processes condition nodes
type ConditionNodeProcessor struct{}

func (p *ConditionNodeProcessor) ProcessNode(ctx *models.ExecutionContext, node *models.FlowNode, edges []models.FlowEdge) (*models.ExecutionResult, error) {
	// Get condition from node data
	conditionType, _ := node.Data["conditionType"].(string)
	variableName, _ := node.Data["variable"].(string)
	operator, _ := node.Data["operator"].(string)
	compareValue, _ := node.Data["value"].(string)

	// Evaluate condition
	conditionMet := false

	if conditionType == "message_contains" {
		// Check if user message contains a keyword
		keyword, _ := node.Data["keyword"].(string)
		conditionMet = strings.Contains(strings.ToLower(ctx.UserMessage), strings.ToLower(keyword))
	} else if conditionType == "variable_check" {
		// Check variable value
		if ctx.Variables != nil {
			variableValue := ctx.Variables[variableName]
			conditionMet = p.evaluateCondition(variableValue, operator, compareValue)
		}
	}

	// Find next node based on condition
	nextNodeID := ""
	for _, edge := range edges {
		if edge.Source != node.ID {
			continue
		}

		// Check edge label/handle for "true" or "false"
		if conditionMet {
			if edge.SourceHandle == "true" || edge.Label == "true" || edge.Label == "Yes" {
				nextNodeID = edge.Target
				break
			}
		} else {
			if edge.SourceHandle == "false" || edge.Label == "false" || edge.Label == "No" {
				nextNodeID = edge.Target
				break
			}
		}
	}

	// If no specific edge found, take first available
	if nextNodeID == "" {
		for _, edge := range edges {
			if edge.Source == node.ID {
				nextNodeID = edge.Target
				break
			}
		}
	}

	return &models.ExecutionResult{
		Success:       true,
		Message:       fmt.Sprintf("Condition evaluated: %v", conditionMet),
		NextNodeID:    nextNodeID,
		ShouldReply:   false,
		Variables:     ctx.Variables,
		CompletedFlow: nextNodeID == "",
	}, nil
}

func (p *ConditionNodeProcessor) GetNodeType() models.NodeType {
	return models.NodeTypeCondition
}

func (p *ConditionNodeProcessor) evaluateCondition(value interface{}, operator string, compareValue string) bool {
	valueStr := fmt.Sprintf("%v", value)

	switch operator {
	case "equals", "==":
		return valueStr == compareValue
	case "not_equals", "!=":
		return valueStr != compareValue
	case "contains":
		return strings.Contains(strings.ToLower(valueStr), strings.ToLower(compareValue))
	case "greater_than", ">":
		valNum, err1 := strconv.ParseFloat(valueStr, 64)
		compareNum, err2 := strconv.ParseFloat(compareValue, 64)
		if err1 == nil && err2 == nil {
			return valNum > compareNum
		}
	case "less_than", "<":
		valNum, err1 := strconv.ParseFloat(valueStr, 64)
		compareNum, err2 := strconv.ParseFloat(compareValue, 64)
		if err1 == nil && err2 == nil {
			return valNum < compareNum
		}
	}

	return false
}

// DelayNodeProcessor processes delay nodes
type DelayNodeProcessor struct{}

func (p *DelayNodeProcessor) ProcessNode(ctx *models.ExecutionContext, node *models.FlowNode, edges []models.FlowEdge) (*models.ExecutionResult, error) {
	// Get delay duration from node data
	delaySeconds := 0
	if delayVal, ok := node.Data["delay"].(float64); ok {
		delaySeconds = int(delayVal)
	} else if delayStr, ok := node.Data["delay"].(string); ok {
		if val, err := strconv.Atoi(delayStr); err == nil {
			delaySeconds = val
		}
	}

	if delaySeconds <= 0 {
		delaySeconds = 1 // Default 1 second
	}

	// In a real implementation, this would schedule the next node execution
	// For now, we'll just pause
	time.Sleep(time.Duration(delaySeconds) * time.Second)

	// Find next node
	nextNodeID := ""
	for _, edge := range edges {
		if edge.Source == node.ID {
			nextNodeID = edge.Target
			break
		}
	}

	return &models.ExecutionResult{
		Success:       true,
		Message:       fmt.Sprintf("Delayed for %d seconds", delaySeconds),
		NextNodeID:    nextNodeID,
		ShouldReply:   false,
		Variables:     ctx.Variables,
		DelaySeconds:  delaySeconds,
		CompletedFlow: nextNodeID == "",
	}, nil
}

func (p *DelayNodeProcessor) GetNodeType() models.NodeType {
	return models.NodeTypeDelay
}

// EndNodeProcessor processes end nodes
type EndNodeProcessor struct{}

func (p *EndNodeProcessor) ProcessNode(ctx *models.ExecutionContext, node *models.FlowNode, edges []models.FlowEdge) (*models.ExecutionResult, error) {
	// Get final message if any
	message, _ := node.Data["message"].(string)
	if message == "" {
		message = "Thank you for your time!"
	}

	return &models.ExecutionResult{
		Success:       true,
		Message:       "Flow completed",
		Response:      message,
		ShouldReply:   true,
		Variables:     ctx.Variables,
		CompletedFlow: true,
	}, nil
}

func (p *EndNodeProcessor) GetNodeType() models.NodeType {
	return models.NodeTypeEnd
}
