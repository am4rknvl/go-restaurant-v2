# Error Handling Documentation

## Overview

This project implements comprehensive error handling across all API endpoints with standardized error responses, proper HTTP status codes, and detailed error messages.

## Error Response Structure

All errors follow a consistent JSON structure:

```json
{
  "code": "error_code",
  "message": "Human-readable error message",
  "details": "Optional detailed error information"
}
```

### Validation Errors

For validation errors with multiple fields:

```json
{
  "code": "validation_error",
  "message": "Validation failed",
  "errors": [
    {
      "field": "email",
      "message": "Invalid email format"
    },
    {
      "field": "password",
      "message": "password must be at least 8 characters"
    }
  ]
}
```

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `bad_request` | 400 | Invalid request parameters |
| `unauthorized` | 401 | Authentication required or failed |
| `forbidden` | 403 | Insufficient permissions |
| `not_found` | 404 | Resource not found |
| `conflict` | 409 | Resource already exists |
| `validation_error` | 400 | Input validation failed |
| `unprocessable_entity` | 422 | Request understood but cannot be processed |
| `internal_error` | 500 | Internal server error |
| `database_error` | 500 | Database operation failed |

## Usage in Handlers

### Basic Error Handling

```go
import "restaurant-system/internal/errors"

// Handle binding errors
if err := c.ShouldBindJSON(&body); err != nil {
    errors.HandleBindError(c, err)
    return
}

// Handle validation errors
if body.Name == "" {
    errors.RespondError(c, errors.NewValidation("name is required"))
    return
}

// Handle database errors
if err := service.Create(ctx, &item); err != nil {
    errors.HandleDatabaseError(c, err, "create item")
    return
}

// Success responses
errors.RespondCreated(c, item)  // 201 Created
errors.RespondOK(c, data)       // 200 OK
errors.RespondNoContent(c)      // 204 No Content
```

### Custom Error Messages

```go
// Create custom errors
err := errors.NewBadRequest("Invalid order status")
err = errors.NewUnauthorized("Invalid credentials")
err = errors.NewNotFound("Order not found")
err = errors.NewConflict("Email already exists")

// Add details to errors
err := errors.NewInternal("Failed to process payment").WithDetails(originalErr.Error())
errors.RespondError(c, err)
```

### Validation Helpers

```go
// Validate required fields
if !errors.ValidateRequired(c, "email", body.Email) {
    return
}

// Validate UUID format
if !errors.ValidateUUID(c, "order_id", orderID) {
    return
}

// Multiple field validation
validationErrors := []errors.ValidationError{
    {Field: "email", Message: "Invalid email format"},
    {Field: "age", Message: "Must be at least 18"},
}
errors.RespondValidationErrors(c, validationErrors)
```

## Middleware

### Error Handler Middleware

Automatically catches panics and handles errors:

```go
router.Use(middleware.ErrorHandler())
```

Features:
- Panic recovery with stack trace logging
- Automatic error response formatting
- Context error handling

### Request Logger Middleware

Logs all incoming requests and responses:

```go
router.Use(middleware.RequestLogger())
```

Logs:
- Request method, path, and client IP
- Response status code
- Request/response timing

## Examples by Endpoint Type

### Menu Endpoints

```go
// Create menu item (admin only)
POST /api/v1/menu/item
{
  "name": "Pizza",
  "category": "main",
  "price": 15.99
}

// Success: 201 Created
// Error: 400 Bad Request
{
  "code": "validation_error",
  "message": "name is required"
}
```

### Order Endpoints

```go
// Create order
POST /api/v1/orders
{
  "customer_id": "123",
  "items": [
    {"menu_item_id": "456", "quantity": 2}
  ]
}

// Success: 201 Created
// Error: 400 Bad Request
{
  "code": "validation_error",
  "message": "at least one item is required"
}
```

### Authentication Endpoints

```go
// Verify OTP
POST /api/v1/auth/verify-otp
{
  "phone_number": "+1234567890",
  "code": "123456"
}

// Success: 200 OK
{
  "access_token": "...",
  "refresh_token": "..."
}

// Error: 401 Unauthorized
{
  "code": "unauthorized",
  "message": "Invalid OTP code",
  "details": "OTP expired or incorrect"
}
```

### Payment Endpoints

```go
// Create payment
POST /api/v1/payments
{
  "order_id": 123,
  "amount_cents": 1599,
  "provider": "telebirr"
}

// Success: 201 Created
// Error: 400 Bad Request
{
  "code": "validation_error",
  "message": "invalid payment provider"
}
```

## Security Considerations

### Menu Endpoints Security

- **Read operations** (GET): Public access - customers can view menus
- **Write operations** (POST, PUT, DELETE): Admin only
- Categories: Admin only for write, public for read
- QR Menu: Public access for customer scanning

### Error Message Security

- Never expose sensitive information in error messages
- Use generic messages for authentication failures
- Log detailed errors server-side, return sanitized messages to clients
- Don't reveal database schema or internal implementation details

## Best Practices

1. **Always validate input** before processing
2. **Use appropriate HTTP status codes**
3. **Provide clear, actionable error messages**
4. **Log errors server-side** for debugging
5. **Handle database errors gracefully**
6. **Validate business logic** (e.g., payment amounts > 0)
7. **Use middleware** for consistent error handling
8. **Test error scenarios** in addition to happy paths

## Testing Error Handling

### Example Test Cases

```bash
# Test validation error
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id": ""}'

# Expected: 400 Bad Request
# {"code":"validation_error","message":"customer_id is required"}

# Test unauthorized access
curl -X POST http://localhost:8080/api/v1/menu/item \
  -H "Content-Type: application/json" \
  -d '{"name":"Pizza"}'

# Expected: 401 Unauthorized
# {"code":"unauthorized","message":"Authentication required"}

# Test not found
curl http://localhost:8080/api/v1/payments/99999

# Expected: 404 Not Found
# {"code":"not_found","message":"Payment not found"}
```

## Migration Guide

### Before (Old Style)

```go
if err := c.ShouldBindJSON(&body); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
}
```

### After (New Style)

```go
if err := c.ShouldBindJSON(&body); err != nil {
    errors.HandleBindError(c, err)
    return
}
```

## Future Improvements

- [ ] Add error tracking/monitoring integration (e.g., Sentry)
- [ ] Implement rate limiting with proper error responses
- [ ] Add request ID tracking for error correlation
- [ ] Implement circuit breaker pattern for external services
- [ ] Add error metrics and monitoring dashboards
- [ ] Internationalization (i18n) for error messages
- [ ] Add error recovery strategies for specific error types
