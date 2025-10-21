package models

import "time"

// User represents a user in the system
type User struct {
	ID         string     `json:"id"`
	Email      string     `json:"email"`
	FullName   string     `json:"full_name"`
	Password   string     `json:"password,omitempty"` // Omit in responses
	Gmail      *string    `json:"gmail,omitempty"`
	Phone      *string    `json:"phone,omitempty"`
	Status     string     `json:"status"`
	Expired    *string    `json:"expired,omitempty"`
	IsActive   bool       `json:"is_active"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	LastLogin  *time.Time `json:"last_login,omitempty"`
}

// UserSession represents a user session
type UserSession struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// RegisterRequest is the request body for user registration
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	FullName string `json:"full_name" validate:"required,min=2"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginRequest is the request body for user login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse is the response for successful authentication
type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
	User    *User  `json:"user,omitempty"`
}
