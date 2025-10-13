package models

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error string `json:"error" example:"invalid_credentials"`
}

// AppErrorResponse represents detailed error response with code
type AppErrorResponse struct {
	Code    string `json:"code" example:"validation_error"`
	Message string `json:"message" example:"Invalid input provided"`
	Details string `json:"details,omitempty" example:"name is required"`
}

// ValidationErrorResponse represents validation errors
type ValidationErrorResponse struct {
	Code    string             `json:"code" example:"validation_error"`
	Message string             `json:"message" example:"Validation failed"`
	Errors  []ValidationDetail `json:"errors"`
}

// ValidationDetail represents a single field validation error
type ValidationDetail struct {
	Field   string `json:"field" example:"email"`
	Message string `json:"message" example:"Invalid email format"`
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

// CategoriesListResponse represents list of categories
type CategoriesListResponse struct {
	Categories []Category `json:"categories"`
}

// MenuItemsListResponse represents list of menu items
type MenuItemsListResponse struct {
	Items []MenuItem `json:"items"`
}

// OrdersListResponse represents list of orders
type OrdersListResponse struct {
	Orders []Order `json:"orders"`
}

// TablesListResponse represents list of tables
type TablesListResponse struct {
	Tables []Table `json:"tables"`
}

// ReservationsListResponse represents list of reservations
type ReservationsListResponse struct {
	Reservations []Reservation `json:"reservations"`
}

// SubscriptionsListResponse represents list of notification subscriptions
type SubscriptionsListResponse struct {
	Subscriptions []Subscription `json:"subscriptions"`
}
