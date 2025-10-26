package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"chatbot-automation/internal/models"
	"chatbot-automation/internal/repository"
)

// FlowExecutionEngine handles the execution of flow nodes
type FlowExecutionEngine struct {
	deviceRepo *repository.DeviceRepository
	convRepo   *repository.ConversationRepository
}

// FlowNode represents a node in the flow
type FlowNode struct {
	ID     string                 `json:"id"`
	Type   string                 `json:"type"`
	Label  string                 `json:"label"`
	Config map[string]interface{} `json:"config"`
	X      float64                `json:"x"`
	Y      float64                `json:"y"`
}

// FlowEdge represents a connection between nodes
type FlowEdge struct {
	From          string `json:"from"`
	To            string `json:"to"`
	ConditionType string `json:"conditionType,omitempty"`
	ConditionValue string `json:"conditionValue,omitempty"`
}

// FlowData represents the complete flow structure
type FlowData struct {
	Nodes       []FlowNode `json:"nodes"`
	Connections []FlowEdge `json:"connections"`
}

// ExecuteFlow processes the flow starting from a specific node
func (s *FlowProcessorService) ExecuteFlow(
	ctx context.Context,
	flow *models.ChatbotFlow,
	conversationID string,
	userMessage string,
	currentStage string,
) error {
	log.Printf("🚀 Starting flow execution for conversation: %s", conversationID)

	// Check if NodesData is empty
	if flow.NodesData == "" {
		log.Printf("⚠️  Flow NodesData is empty - flow not configured yet")
		return fmt.Errorf("flow has no nodes configured")
	}

	// Parse flow data
	var flowData FlowData
	if err := json.Unmarshal([]byte(flow.NodesData), &flowData); err != nil {
		log.Printf("❌ Failed to parse flow data: %v", err)
		log.Printf("📝 NodesData content: %s", flow.NodesData)
		return fmt.Errorf("failed to parse flow data: %w", err)
	}

	log.Printf("📊 Flow has %d nodes and %d connections", len(flowData.Nodes), len(flowData.Connections))

	// Find starting node
	startNode := s.findStartingNode(flowData, currentStage)
	if startNode == nil {
		log.Printf("⚠️  No starting node found for stage: %s", currentStage)
		return fmt.Errorf("no starting node found")
	}

	log.Printf("🎯 Starting from node: %s (Type: %s)", startNode.ID, startNode.Type)

	// Execute flow from starting node
	return s.executeFromNode(ctx, flow, &flowData, startNode, conversationID, userMessage, currentStage)
}

// findStartingNode finds the node to start execution from
func (s *FlowProcessorService) findStartingNode(flowData FlowData, currentStage string) *FlowNode {
	// If no current stage, find the first node (after start node if exists)
	if currentStage == "" || currentStage == "start" {
		// Look for a node that has no incoming connections (or is connected from start)
		for i := range flowData.Nodes {
			node := &flowData.Nodes[i]
			// Skip if this is a start-type node
			if strings.Contains(strings.ToLower(node.Type), "start") {
				continue
			}

			// Check if this node has incoming connections
			hasIncoming := false
			for _, edge := range flowData.Connections {
				if edge.To == node.ID {
					hasIncoming = true
					break
				}
			}

			// If no incoming connections, this could be the first node
			if !hasIncoming {
				return node
			}
		}

		// If all nodes have incoming connections, get the first node
		if len(flowData.Nodes) > 0 {
			return &flowData.Nodes[0]
		}
	}

	// Otherwise, try to find node by ID matching the stage
	for i := range flowData.Nodes {
		if flowData.Nodes[i].ID == currentStage {
			return &flowData.Nodes[i]
		}
	}

	// Default to first node
	if len(flowData.Nodes) > 0 {
		return &flowData.Nodes[0]
	}

	return nil
}

// executeFromNode executes the flow starting from a specific node
func (s *FlowProcessorService) executeFromNode(
	ctx context.Context,
	flow *models.ChatbotFlow,
	flowData *FlowData,
	node *FlowNode,
	conversationID string,
	userMessage string,
	currentStage string,
) error {
	log.Printf("🔄 Executing node: %s (Type: %s)", node.ID, node.Type)

	// Execute the current node
	continueFlow, err := s.executeNode(ctx, flow, node, conversationID, userMessage)
	if err != nil {
		return fmt.Errorf("failed to execute node %s: %w", node.ID, err)
	}

	// If node says to stop flow (e.g., waiting_reply), stop here
	if !continueFlow {
		log.Printf("⏸️  Flow paused at node: %s", node.ID)
		// Update current node in conversation
		return s.updateConversationNode(ctx, conversationID, node.ID)
	}

	// Find next node
	nextNode := s.findNextNode(flowData, node, userMessage)
	if nextNode == nil {
		log.Printf("✅ Flow completed - no more nodes")
		return s.updateConversationNode(ctx, conversationID, "completed")
	}

	// Continue to next node
	return s.executeFromNode(ctx, flow, flowData, nextNode, conversationID, userMessage, currentStage)
}

// executeNode executes a single node's action
func (s *FlowProcessorService) executeNode(
	ctx context.Context,
	flow *models.ChatbotFlow,
	node *FlowNode,
	conversationID string,
	userMessage string,
) (bool, error) {
	log.Printf("⚙️  Executing node type: %s", node.Type)

	switch node.Type {
	case "send_message":
		return s.executeSendMessage(ctx, flow, node, conversationID)

	case "delay":
		return s.executeDelay(ctx, node)

	case "waiting_reply":
		return s.executeWaitingReply(ctx, conversationID, node)

	case "waiting_times":
		return s.executeWaitingTimes(ctx, conversationID, node)

	case "ai_prompt":
		return s.executeAIPrompt(ctx, flow, node, conversationID, userMessage)

	case "stage":
		return s.executeStage(ctx, conversationID, node)

	case "send_image", "send_audio", "send_video":
		return s.executeSendMedia(ctx, flow, node, conversationID)

	case "conditions":
		return s.executeConditions(ctx, node, userMessage)

	default:
		log.Printf("⚠️  Unknown node type: %s, skipping", node.Type)
		return true, nil
	}
}

// executeSendMessage sends a WhatsApp message
func (s *FlowProcessorService) executeSendMessage(
	ctx context.Context,
	flow *models.ChatbotFlow,
	node *FlowNode,
	conversationID string,
) (bool, error) {
	// Get message text from config
	text, ok := node.Config["text"].(string)
	if !ok || text == "" {
		log.Printf("⚠️  No text configured for send_message node")
		return true, nil
	}

	log.Printf("📤 Sending message: %s", text)

	// Get conversation to get phone number
	conversation, err := s.convRepo.GetConversationByID(ctx, conversationID)
	if err != nil || conversation == nil {
		log.Printf("❌ Failed to get conversation for sending: %v", err)
		return true, fmt.Errorf("failed to get conversation: %w", err)
	}

	// Send WhatsApp message
	err = s.whatsappService.SendMessage(ctx, flow.IDDevice, conversation.ProspectNum, text, "", "")
	if err != nil {
		log.Printf("❌ Failed to send WhatsApp message: %v", err)
		return true, fmt.Errorf("failed to send message: %w", err)
	}

	log.Printf("✅ Message sent successfully to %s", conversation.ProspectNum)

	// Update conv_last with bot reply
	return true, s.updateConvLast(ctx, conversationID, "Bot", text)
}

// executeDelay pauses execution for specified seconds
func (s *FlowProcessorService) executeDelay(ctx context.Context, node *FlowNode) (bool, error) {
	// Get delay from config (should be 3 seconds by default)
	delay := 3
	if delayVal, ok := node.Config["delay"].(float64); ok {
		delay = int(delayVal)
	}

	log.Printf("⏱️  Delaying for %d seconds", delay)
	time.Sleep(time.Duration(delay) * time.Second)
	log.Printf("✅ Delay completed")

	return true, nil
}

// executeWaitingReply pauses flow until user replies (no timeout)
func (s *FlowProcessorService) executeWaitingReply(
	ctx context.Context,
	conversationID string,
	node *FlowNode,
) (bool, error) {
	log.Printf("⏸️  Waiting for user reply (no timeout)")

	// Update conversation to waiting state
	// Flow will resume when next webhook message arrives
	return false, nil // false = stop flow execution
}

// executeWaitingTimes pauses and waits for user reply with timeout
func (s *FlowProcessorService) executeWaitingTimes(
	ctx context.Context,
	conversationID string,
	node *FlowNode,
) (bool, error) {
	// Get timeout from config (should be 8 seconds by default)
	timeout := 8
	if delayVal, ok := node.Config["delay"].(float64); ok {
		timeout = int(delayVal)
	}

	log.Printf("⏳ Waiting for user reply with %d second timeout", timeout)

	// TODO: Implement timeout logic
	// For now, just continue after timeout
	time.Sleep(time.Duration(timeout) * time.Second)

	log.Printf("⏱️  Timeout reached, continuing flow")
	return true, nil
}

// executeAIPrompt processes AI prompt node
func (s *FlowProcessorService) executeAIPrompt(
	ctx context.Context,
	flow *models.ChatbotFlow,
	node *FlowNode,
	conversationID string,
	userMessage string,
) (bool, error) {
	// Get AI prompt from config
	prompt, ok := node.Config["text"].(string)
	if !ok || prompt == "" {
		log.Printf("⚠️  No prompt configured for AI node")
		return true, nil
	}

	log.Printf("✨ Processing AI prompt: %s", prompt)

	// TODO: Implement AI API call (OpenAI, etc.)
	// This will send prompt + user message to AI and get response

	aiResponse := "AI response placeholder"
	log.Printf("🤖 AI Response: %s", aiResponse)

	// Send AI response back to user
	// TODO: Implement WhatsApp API call

	// Update conv_last with bot reply
	return true, s.updateConvLast(ctx, conversationID, "Bot", aiResponse)
}

// executeStage updates the conversation stage
func (s *FlowProcessorService) executeStage(
	ctx context.Context,
	conversationID string,
	node *FlowNode,
) (bool, error) {
	// Get stage name from config
	stageName, ok := node.Config["value"].(string)
	if !ok || stageName == "" {
		log.Printf("⚠️  No stage value configured")
		return true, nil
	}

	log.Printf("🎯 Updating stage to: %s", stageName)

	// Update conversation stage
	updates := map[string]interface{}{
		"stage": stageName,
	}

	err := s.convRepo.UpdateConversation(ctx, conversationID, updates)
	if err != nil {
		return true, fmt.Errorf("failed to update stage: %w", err)
	}

	log.Printf("✅ Stage updated successfully")
	return true, nil
}

// executeSendMedia sends media (image/audio/video)
func (s *FlowProcessorService) executeSendMedia(
	ctx context.Context,
	flow *models.ChatbotFlow,
	node *FlowNode,
	conversationID string,
) (bool, error) {
	// Get media URL from config
	url, ok := node.Config["url"].(string)
	if !ok || url == "" {
		log.Printf("⚠️  No URL configured for media node")
		return true, nil
	}

	log.Printf("📤 Sending %s: %s", node.Type, url)

	// Get conversation to get phone number
	conversation, err := s.convRepo.GetConversationByID(ctx, conversationID)
	if err != nil || conversation == nil {
		log.Printf("❌ Failed to get conversation for sending media: %v", err)
		return true, fmt.Errorf("failed to get conversation: %w", err)
	}

	// Map node type to media type
	mediaType := ""
	switch node.Type {
	case "send_image":
		mediaType = "image"
	case "send_audio":
		mediaType = "audio"
	case "send_video":
		mediaType = "video"
	}

	// Send WhatsApp media
	err = s.whatsappService.SendMessage(ctx, flow.IDDevice, conversation.ProspectNum, "", mediaType, url)
	if err != nil {
		log.Printf("❌ Failed to send WhatsApp media: %v", err)
		return true, fmt.Errorf("failed to send media: %w", err)
	}

	log.Printf("✅ Media sent successfully to %s", conversation.ProspectNum)

	// Update conv_last with bot media send (just the URL)
	return true, s.updateConvLast(ctx, conversationID, "Bot", url)
}

// executeConditions evaluates conditions
func (s *FlowProcessorService) executeConditions(
	ctx context.Context,
	node *FlowNode,
	userMessage string,
) (bool, error) {
	log.Printf("🔀 Evaluating conditions")

	// Conditions are handled in findNextNode
	// This node just passes through
	return true, nil
}

// findNextNode finds the next node to execute based on edges
func (s *FlowProcessorService) findNextNode(
	flowData *FlowData,
	currentNode *FlowNode,
	userMessage string,
) *FlowNode {
	// Find all outgoing edges from current node
	var outgoingEdges []FlowEdge
	for _, edge := range flowData.Connections {
		if edge.From == currentNode.ID {
			outgoingEdges = append(outgoingEdges, edge)
		}
	}

	if len(outgoingEdges) == 0 {
		log.Printf("ℹ️  No outgoing edges from node: %s", currentNode.ID)
		return nil
	}

	// If only one edge, follow it
	if len(outgoingEdges) == 1 {
		return s.findNodeByID(flowData, outgoingEdges[0].To)
	}

	// Multiple edges - check conditions
	// TODO: Implement condition matching
	// For now, just follow the first edge
	log.Printf("⚠️  Multiple edges found, following first one")
	return s.findNodeByID(flowData, outgoingEdges[0].To)
}

// findNodeByID finds a node by its ID
func (s *FlowProcessorService) findNodeByID(flowData *FlowData, nodeID string) *FlowNode {
	for i := range flowData.Nodes {
		if flowData.Nodes[i].ID == nodeID {
			return &flowData.Nodes[i]
		}
	}
	return nil
}

// updateConversationNode updates the current node in conversation
func (s *FlowProcessorService) updateConversationNode(
	ctx context.Context,
	conversationID string,
	nodeID string,
) error {
	updates := map[string]interface{}{
		"current_node_id": nodeID,
	}

	return s.convRepo.UpdateConversation(ctx, conversationID, updates)
}

// updateConvLast updates the conversation history
func (s *FlowProcessorService) updateConvLast(
	ctx context.Context,
	conversationID string,
	role string,
	message string,
) error {
	// Get current conversation
	conv, err := s.convRepo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return err
	}

	// Append to conv_last
	convLast := ""
	if conv.ConvLast != nil {
		convLast = *conv.ConvLast
	}

	newLine := fmt.Sprintf("%s: %s", role, message)
	if convLast != "" {
		convLast += "\n" + newLine
	} else {
		convLast = newLine
	}

	updates := map[string]interface{}{
		"conv_last": convLast,
	}

	return s.convRepo.UpdateConversation(ctx, conversationID, updates)
}
