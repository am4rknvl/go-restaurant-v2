package models

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error string `json:"error" example:"invalid_credentials"`
}

// MessageResponse represents a standard message response
type MessageResponse struct {
	Message string `json:"message" example:"operation_successful"`
}

// TokenResponse represents authentication token response
type TokenResponse struct {
	Token     string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	AccountID string `json:"account_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Role      string `json:"role" example:"customer"`
}

// TokenPairResponse represents access and refresh token pair
type TokenPairResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}
