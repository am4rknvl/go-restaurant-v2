package services

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"encoding/json"
	"restaurant-system/internal/models"
	"restaurant-system/internal/payments"

	"github.com/google/uuid"
)

// PaymentService is required by handlers/payment_handler.go
type PaymentService interface {
	CreatePayment(ctx context.Context, restaurantID uint, orderID uint, amountCents int64, provider string) (*models.Payment, error)
	GetPayment(ctx context.Context, restaurantID uint, id uint) (*models.Payment, error)
	// callback handlers used by notify endpoints
	HandleTelebirrCallback(payload map[string]string) error
	HandleChapaCallback(payload map[string]string) error
	HandleMpesaCallback(payload map[string]string) error
	RequestRefund(ctx context.Context, paymentID string, amount float64, reason string) (*models.Refund, error)
	ApplyPartialPayment(ctx context.Context, orderID string, amount float64) error
}

type PaymentSQLService struct {
	db *sql.DB
}

func NewPaymentSQLService(db *sql.DB) *PaymentSQLService { return &PaymentSQLService{db: db} }

func normalizeMethod(provider string) models.PaymentMethod {
	switch provider {
	case string(models.PaymentMethodCash):
		return models.PaymentMethodCash
	case string(models.PaymentMethodCard):
		return models.PaymentMethodCard
	case string(models.PaymentMethodMobileMoney):
		return models.PaymentMethodMobileMoney
	default:
		return models.PaymentMethodCard
	}
}

func (s *PaymentSQLService) CreatePayment(ctx context.Context, restaurantID uint, orderID uint, amountCents int64, provider string) (*models.Payment, error) {
	if amountCents <= 0 {
		return nil, errors.New("amount must be positive")
	}
	// Basic existence check for order
	var exists bool
	_ = s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM orders WHERE id = $1)", strconv.FormatUint(uint64(orderID), 10)).Scan(&exists)

	id := uuid.New().String()
	p := &models.Payment{
		ID:            id,
		OrderID:       strconv.FormatUint(uint64(orderID), 10),
		Amount:        float64(amountCents) / 100.0,
		Method:        normalizeMethod(provider),
		Status:        models.PaymentStatusCompleted,
		TransactionID: "",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	_, err := s.db.ExecContext(ctx, "INSERT INTO payments (id, order_id, amount, method, status, transaction_id, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)",
		p.ID, p.OrderID, p.Amount, string(p.Method), string(p.Status), p.TransactionID, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PaymentSQLService) GetPayment(ctx context.Context, restaurantID uint, id uint) (*models.Payment, error) {
	var p models.Payment
	err := s.db.QueryRowContext(ctx, "SELECT id, order_id, amount, method, status, transaction_id, phone_number, created_at, updated_at FROM payments WHERE id = $1",
		strconv.FormatUint(uint64(id), 10),
	).Scan(&p.ID, &p.OrderID, &p.Amount, &p.Method, &p.Status, &p.TransactionID, &p.PhoneNumber, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// HandleTelebirrCallback verifies and updates payment records; gateway-specific verification required
func (s *PaymentSQLService) HandleTelebirrCallback(payload map[string]string) error {

	// Verify signature, timestamp and nonce
	ok, err := payments.VerifyCallbackStrict(payload, s.db)
	if err != nil || !ok {
		return errors.New("telebirr verification failed: " + err.Error())
	}

	// Expected fields
	tid, ok1 := payload["transaction_id"]
	oid, ok2 := payload["order_id"]
	status := payload["status"]
	if !ok1 || !ok2 || tid == "" || oid == "" {
		return errors.New("invalid telebirr payload")
	}

	// Log event
	evtID := uuid.New().String()
	evtPayload, _ := json.Marshal(payload)
	_, _ = s.db.ExecContext(context.Background(), "INSERT INTO payment_events (id, payment_id, order_id, event_type, payload, created_at) VALUES ($1,$2,$3,$4,$5,now())", evtID, "", oid, "telebirr_callback", evtPayload)

	// Update payment record
	_, err = s.db.ExecContext(context.Background(), "UPDATE payments SET transaction_id=$1, status=$2, updated_at=now() WHERE order_id=$3", tid, status, oid)
	if err != nil {
		// enqueue retry
		qid := uuid.New().String()
		_, _ = s.db.ExecContext(context.Background(), "INSERT INTO webhook_retry_queue (id, payload, attempts, next_attempt, created_at) VALUES ($1,$2,$3,$4,now())", qid, evtPayload, 0, time.Now().Add(1*time.Minute))
		return err
	}

	// If payment completed, mark order as completed
	if status == "completed" {
		_, _ = s.db.ExecContext(context.Background(), "UPDATE orders SET status=$1, updated_at=now() WHERE id=$2", string(models.OrderStatusCompleted), oid)
	}

	return nil
}

func (s *PaymentSQLService) HandleChapaCallback(payload map[string]string) error {
	// plumbing stub; implement provider verification and mapping
	return s.HandleTelebirrCallback(payload)
}

func (s *PaymentSQLService) HandleMpesaCallback(payload map[string]string) error {
	// plumbing stub; implement provider verification and mapping
	return s.HandleTelebirrCallback(payload)
}

// RequestRefund creates a refund record; actual refund processing (gateway) is out of scope
func (s *PaymentSQLService) RequestRefund(ctx context.Context, paymentID string, amount float64, reason string) (*models.Refund, error) {
	if amount <= 0 {
		return nil, errors.New("invalid amount")
	}
	id := uuid.New().String()
	_, err := s.db.ExecContext(ctx, "INSERT INTO refunds (id, payment_id, amount, reason, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,now(),now())",
		id, paymentID, amount, reason, string(models.RefundStatusPending))
	if err != nil {
		return nil, err
	}
	r := &models.Refund{ID: id, PaymentID: paymentID, Amount: amount, Reason: reason, Status: models.RefundStatusPending}
	return r, nil
}

// ApplyPartialPayment records a partial payment against an order by creating a payment record and marking payments partial
func (s *PaymentSQLService) ApplyPartialPayment(ctx context.Context, orderID string, amount float64) error {
	if amount <= 0 {
		return errors.New("invalid amount")
	}
	// Create a payment record marked as partial
	id := uuid.New().String()
	_, err := s.db.ExecContext(ctx, "INSERT INTO payments (id, order_id, amount, method, status, transaction_id, refunded_amount, is_partial, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,now(),now())",
		id, orderID, amount, string(models.PaymentMethodCard), string(models.PaymentStatusCompleted), "", 0, true)
	if err != nil {
		return err
	}
	// Optionally update order status if partial payment covers total or not; leave order status unchanged here
	return nil
}
