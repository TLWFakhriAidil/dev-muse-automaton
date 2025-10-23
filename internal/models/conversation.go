package models

import "time"

// AIWhatsapp represents a WhatsApp conversation with a prospect
type AIWhatsapp struct {
	IDProspect      string     `json:"id_prospect"`
	ProspectNum     string     `json:"prospect_num"`
	IDDevice        string     `json:"id_device"`
	Stage           *string    `json:"stage,omitempty"`
	Niche           *string    `json:"niche,omitempty"`
	ConvLast        *string    `json:"conv_last,omitempty"` // Stores "User: message\nBot: reply"
	IsActive        bool       `json:"is_active"`
	LastInteraction *time.Time `json:"last_interaction,omitempty"`
	FlowID          *string    `json:"flow_id,omitempty"`
	CurrentNode     *string    `json:"current_node,omitempty"`
	SessionData     map[string]interface{} `json:"session_data,omitempty"` // JSONB for flow execution state
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	Status          string     `json:"status"` // active, completed, abandoned
}

// CreateConversationRequest is the request body for creating a conversation
type CreateConversationRequest struct {
	ProspectNum string  `json:"prospect_num" validate:"required"`
	IDDevice    string  `json:"id_device" validate:"required"`
	Stage       *string `json:"stage,omitempty"`
	Niche       *string `json:"niche,omitempty"`
	FlowID      *string `json:"flow_id,omitempty"`
}

// UpdateConversationRequest is the request body for updating a conversation
type UpdateConversationRequest struct {
	Stage               *string                 `json:"stage,omitempty"`
	Niche               *string                 `json:"niche,omitempty"`
	ConversationHistory *map[string]interface{} `json:"conversation_history,omitempty"`
	IsActive            *bool                   `json:"is_active,omitempty"`
	FlowID              *string                 `json:"flow_id,omitempty"`
	CurrentNode         *string                 `json:"current_node,omitempty"`
	SessionData         *map[string]interface{} `json:"session_data,omitempty"`
	Status              *string                 `json:"status,omitempty"` // active, completed, abandoned
}

// AddMessageRequest is the request body for adding a message to conversation history
type AddMessageRequest struct {
	Role    string `json:"role" validate:"required,oneof=user assistant system"`
	Content string `json:"content" validate:"required"`
}

// ConversationResponse is the response for conversation operations
type ConversationResponse struct {
	Success      bool           `json:"success"`
	Message      string         `json:"message"`
	Conversation *AIWhatsapp    `json:"conversation,omitempty"`
	Conversations []AIWhatsapp  `json:"conversations,omitempty"`
}

// ConversationStats represents conversation statistics
type ConversationStats struct {
	TotalConversations     int            `json:"total_conversations"`
	ActiveConversations    int            `json:"active_conversations"`
	CompletedConversations int            `json:"completed_conversations"`
	AbandonedConversations int            `json:"abandoned_conversations"`
	ByStage                map[string]int `json:"by_stage"`
	ByNiche                map[string]int `json:"by_niche"`
	ByDevice               map[string]int `json:"by_device"`
}
