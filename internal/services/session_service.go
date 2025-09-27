package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"restaurant-system/internal/models"
)

type SessionService struct {
	db *sql.DB
}

func NewSessionService(db *sql.DB) *SessionService { return &SessionService{db: db} }

func (s *SessionService) StartSession(ctx context.Context, sess *models.Session) (*models.Session, error) {
	if sess.TableID == "" {
		return nil, errors.New("table_id required")
	}
	now := time.Now()
	_, err := s.db.ExecContext(ctx, "INSERT INTO sessions (id, table_id, customer, status, created_at) VALUES ($1,$2,$3,$4,$5)", sess.ID, sess.TableID, sess.Customer, string(models.SessionStatusActive), now)
	if err != nil {
		return nil, err
	}
	sess.Status = models.SessionStatusActive
	sess.CreatedAt = now
	return sess, nil
}

func (s *SessionService) GetSession(ctx context.Context, id string) (*models.Session, error) {
	var sess models.Session
	if err := s.db.QueryRowContext(ctx, "SELECT id, table_id, customer, status, created_at, closed_at FROM sessions WHERE id=$1", id).
		Scan(&sess.ID, &sess.TableID, &sess.Customer, &sess.Status, &sess.CreatedAt, &sess.ClosedAt); err != nil {
		return nil, err
	}
	// fetch orders associated with session: orders.table_session_id assumed
	rows, err := s.db.QueryContext(ctx, "SELECT id, customer_id, total_amount, status, created_at, updated_at FROM orders WHERE session_id=$1 ORDER BY created_at ASC", id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var o models.Order
			if err := rows.Scan(&o.ID, &o.CustomerID, &o.TotalAmount, &o.Status, &o.CreatedAt, &o.UpdatedAt); err == nil {
				sess.Orders = append(sess.Orders, o)
			}
		}
	}
	return &sess, nil
}

func (s *SessionService) CloseSession(ctx context.Context, id string) error {
	now := time.Now()
	_, err := s.db.ExecContext(ctx, "UPDATE sessions SET status=$1, closed_at=$2 WHERE id=$3", string(models.SessionStatusClosed), now, id)
	return err
}
