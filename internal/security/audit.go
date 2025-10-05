package security

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditEvent represents a security audit event
type AuditEvent struct {
	ID          string    `json:"id" db:"id"`
	Timestamp   time.Time `json:"timestamp" db:"timestamp"`
	UserID      string    `json:"user_id" db:"user_id"`
	Action      string    `json:"action" db:"action"`
	Resource    string    `json:"resource" db:"resource"`
	ResourceID  string    `json:"resource_id" db:"resource_id"`
	IPAddress   string    `json:"ip_address" db:"ip_address"`
	UserAgent   string    `json:"user_agent" db:"user_agent"`
	Status      string    `json:"status" db:"status"` // success, failure
	Details     string    `json:"details" db:"details"`
	RequestID   string    `json:"request_id" db:"request_id"`
}

// AuditLogger handles security audit logging
type AuditLogger struct {
	db *sql.DB
}

func NewAuditLogger(db *sql.DB) *AuditLogger {
	return &AuditLogger{db: db}
}

// Log records an audit event
func (al *AuditLogger) Log(ctx context.Context, event AuditEvent) error {
	if event.ID == "" {
		event.ID = generateID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	query := `
		INSERT INTO audit_logs (id, timestamp, user_id, action, resource, resource_id, 
			ip_address, user_agent, status, details, request_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := al.db.ExecContext(ctx, query,
		event.ID, event.Timestamp, event.UserID, event.Action, event.Resource,
		event.ResourceID, event.IPAddress, event.UserAgent, event.Status,
		event.Details, event.RequestID,
	)
	return err
}

// AuditMiddleware logs all requests to sensitive endpoints
func AuditMiddleware(logger *AuditLogger, sensitiveActions []string) gin.HandlerFunc {
	actionMap := make(map[string]bool)
	for _, action := range sensitiveActions {
		actionMap[action] = true
	}

	return func(c *gin.Context) {
		// Check if this is a sensitive action
		path := c.Request.URL.Path
		method := c.Request.Method
		action := method + " " + path

		if !actionMap[action] && !isSensitivePath(path) {
			c.Next()
			return
		}

		// Capture request details
		userID, _ := c.Get("account_id")
		requestID, _ := c.Get("request_id")

		// Process request
		c.Next()

		// Log after request completes
		status := "success"
		if c.Writer.Status() >= 400 {
			status = "failure"
		}

		details := map[string]interface{}{
			"method":      method,
			"path":        path,
			"status_code": c.Writer.Status(),
			"query":       c.Request.URL.Query(),
		}
		detailsJSON, _ := json.Marshal(details)

		event := AuditEvent{
			UserID:     getString(userID),
			Action:     action,
			Resource:   extractResource(path),
			ResourceID: c.Param("id"),
			IPAddress:  c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			Status:     status,
			Details:    string(detailsJSON),
			RequestID:  getString(requestID),
		}

		go logger.Log(context.Background(), event)
	}
}

// QueryAuditLogs retrieves audit logs with filters
func (al *AuditLogger) QueryAuditLogs(ctx context.Context, filters map[string]interface{}, limit int) ([]AuditEvent, error) {
	query := `
		SELECT id, timestamp, user_id, action, resource, resource_id, 
			ip_address, user_agent, status, details, request_id
		FROM audit_logs
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	if userID, ok := filters["user_id"].(string); ok && userID != "" {
		query += " AND user_id = $" + string(rune(argCount))
		args = append(args, userID)
		argCount++
	}

	if action, ok := filters["action"].(string); ok && action != "" {
		query += " AND action = $" + string(rune(argCount))
		args = append(args, action)
		argCount++
	}

	if resource, ok := filters["resource"].(string); ok && resource != "" {
		query += " AND resource = $" + string(rune(argCount))
		args = append(args, resource)
		argCount++
	}

	query += " ORDER BY timestamp DESC LIMIT $" + string(rune(argCount))
	args = append(args, limit)

	rows, err := al.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []AuditEvent
	for rows.Next() {
		var event AuditEvent
		err := rows.Scan(
			&event.ID, &event.Timestamp, &event.UserID, &event.Action,
			&event.Resource, &event.ResourceID, &event.IPAddress,
			&event.UserAgent, &event.Status, &event.Details, &event.RequestID,
		)
		if err != nil {
			continue
		}
		events = append(events, event)
	}

	return events, nil
}

func isSensitivePath(path string) bool {
	sensitivePaths := []string{
		"/api/v1/auth",
		"/api/v1/accounts",
		"/api/v1/payments",
		"/api/v1/inventory",
		"/api/v1/reports",
		"/api/v1/discounts",
	}

	for _, sp := range sensitivePaths {
		if len(path) >= len(sp) && path[:len(sp)] == sp {
			return true
		}
	}
	return false
}

func extractResource(path string) string {
	parts := splitPath(path)
	if len(parts) >= 4 {
		return parts[3] // e.g., /api/v1/orders -> "orders"
	}
	return "unknown"
}

func splitPath(path string) []string {
	var parts []string
	current := ""
	for _, char := range path {
		if char == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func getString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func generateID() string {
	return "audit_" + time.Now().Format("20060102150405") + "_" + randomString(8)
}
