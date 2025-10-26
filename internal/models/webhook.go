package models

// WhacenterWebhookData represents incoming webhook data from Whacenter
type WhacenterWebhookData struct {
	IsGroup  bool   `json:"isGroup"`
	Message  string `json:"message"`
	From     string `json:"from"`
	Phone    string `json:"phone"`
	PushName string `json:"pushName"`
}

// WahaWebhookData represents incoming webhook data from Waha
type WahaWebhookData struct {
	Payload WahaPayload `json:"payload"`
}

// WahaPayload contains the actual message data from Waha
type WahaPayload struct {
	Body  string        `json:"body"`
	From  string        `json:"from"`
	Data  WahaDataInfo  `json:"_data"`
}

// WahaDataInfo contains additional info from Waha
type WahaDataInfo struct {
	Info WahaInfo `json:"Info"`
}

// WahaInfo contains sender/recipient information
type WahaInfo struct {
	PushName     string `json:"PushName"`
	SenderAlt    string `json:"SenderAlt"`
	RecipientAlt string `json:"RecipientAlt"`
}

// ExtractedMessage represents the normalized message data
type ExtractedMessage struct {
	PhoneNumber string
	Message     string
	Name        string
	Provider    string
	DeviceID    string
}

// WasapBot represents a record in wasapbot table for WhatsApp Bot flows
type WasapBot struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	DeviceID     string                 `json:"id_device"` // Database column: id_device
	ProspectNum  string                 `json:"prospect_num"`
	Niche        string                 `json:"niche"`
	Stage        string                 `json:"stage"`
	Data         map[string]interface{} `json:"data"`
	CreatedAt    string                 `json:"created_at"`
	UpdatedAt    string                 `json:"updated_at"`
}

// AIWhatsApp represents a record in ai_whatsapp table for Chatbot AI flows
type AIWhatsApp struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	DeviceID     string                 `json:"id_device"` // Database column: id_device
	ProspectNum  string                 `json:"prospect_num"`
	Niche        string                 `json:"niche"`
	Stage        string                 `json:"stage"`
	Data         map[string]interface{} `json:"data"`
	CreatedAt    string                 `json:"created_at"`
	UpdatedAt    string                 `json:"updated_at"`
}
