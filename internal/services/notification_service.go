package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"

	"restaurant-system/internal/models"

	"github.com/google/uuid"
)

type NotificationService struct {
	db *sql.DB
}

func NewNotificationService(db *sql.DB) *NotificationService { return &NotificationService{db: db} }

func (s *NotificationService) Subscribe(ctx context.Context, sub *models.Subscription) error {
	if sub.ID == "" {
		sub.ID = uuid.New().String()
	}
	if sub.AccountID == "" || sub.Kind == "" || sub.Endpoint == "" {
		return errors.New("invalid subscription")
	}
	meta := sub.Metadata
	// store metadata as JSON string
	if meta == "" {
		meta = "{}"
	}
	_, err := s.db.ExecContext(ctx, "INSERT INTO subscriptions (id, account_id, kind, endpoint, metadata, created_at) VALUES ($1,$2,$3,$4,$5,now())", sub.ID, sub.AccountID, sub.Kind, sub.Endpoint, meta)
	return err
}

func (s *NotificationService) Unsubscribe(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM subscriptions WHERE id=$1", id)
	return err
}

func (s *NotificationService) ListForAccount(ctx context.Context, accountID string) ([]models.Subscription, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, account_id, kind, endpoint, metadata, created_at FROM subscriptions WHERE account_id=$1", accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []models.Subscription
	for rows.Next() {
		var sub models.Subscription
		var meta sql.NullString
		if err := rows.Scan(&sub.ID, &sub.AccountID, &sub.Kind, &sub.Endpoint, &meta, &sub.CreatedAt); err != nil {
			return nil, err
		}
		if meta.Valid {
			sub.Metadata = meta.String
		}
		res = append(res, sub)
	}
	return res, nil
}

// SendNotification sends text payload to a subscription; in production implement web-push & SMS gateway
func (s *NotificationService) SendNotification(ctx context.Context, sub models.Subscription, payload interface{}) error {
	b, _ := json.Marshal(payload)
	switch sub.Kind {
	case "push":
		// TODO: implement web-push using VAPID keys
		log.Printf("[push] to %s: %s\n", sub.Endpoint, string(b))
	case "sms":
		// TODO: integrate SMS provider like Twilio
		log.Printf("[sms] to %s: %s\n", sub.Endpoint, string(b))
	default:
		return errors.New("unknown subscription kind")
	}
	return nil
}
