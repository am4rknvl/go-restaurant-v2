package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"restaurant-system/internal/models"

	"github.com/google/uuid"
)

type OrderSQLService struct {
	db *sql.DB
}

func NewOrderSQLService(db *sql.DB) *OrderSQLService { return &OrderSQLService{db: db} }

type CreateOrderItemReq struct {
	MenuItemID          string `json:"menu_item_id"`
	Quantity            int    `json:"quantity"`
	SpecialInstructions string `json:"special_instructions,omitempty"`
}

// CreateOrder accepts optional sessionID by passing it as last parameter
func (s *OrderSQLService) CreateOrder(ctx context.Context, customerID string, items []CreateOrderItemReq, sessionID ...string) (*models.Order, error) {
	if len(items) == 0 {
		return nil, errors.New("items required")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	orderID := uuid.New().String()
	now := time.Now()

	var total float64
	var orderItems []models.OrderItem
	for _, it := range items {
		var mi models.MenuItem
		e := tx.QueryRowContext(ctx, "SELECT id, name, price FROM menu_items WHERE id=$1 AND available=TRUE", it.MenuItemID).
			Scan(&mi.ID, &mi.Name, &mi.Price)
		if e != nil {
			err = e
			return nil, err
		}
		lineTotal := mi.Price * float64(it.Quantity)
		total += lineTotal
		orderItems = append(orderItems, models.OrderItem{
			ID: uuid.New().String(), OrderID: orderID, MenuItemID: mi.ID,
			Name: mi.Name, Price: mi.Price, Quantity: it.Quantity, TotalPrice: lineTotal,
			SpecialInstructions: it.SpecialInstructions,
		})
	}

	if len(sessionID) > 0 && sessionID[0] != "" {
		_, err = tx.ExecContext(ctx, "INSERT INTO orders (id, customer_id, session_id, total_amount, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)", orderID, customerID, sessionID[0], total, string(models.OrderStatusPending), now, now)
	} else {
		_, err = tx.ExecContext(ctx, "INSERT INTO orders (id, customer_id, total_amount, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6)", orderID, customerID, total, string(models.OrderStatusPending), now, now)
	}
	if err != nil {
		return nil, err
	}

	for _, oi := range orderItems {
		_, err = tx.ExecContext(ctx, "INSERT INTO order_items (id, order_id, menu_item_id, name, price, quantity, total_price, special_instructions) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)", oi.ID, oi.OrderID, oi.MenuItemID, oi.Name, oi.Price, oi.Quantity, oi.TotalPrice, oi.SpecialInstructions)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	ord := &models.Order{ID: orderID, CustomerID: customerID, Items: orderItems, TotalAmount: total, Status: models.OrderStatusPending, CreatedAt: now, UpdatedAt: now}
	if len(sessionID) > 0 {
		ord.SessionID = sessionID[0]
	}
	return ord, nil
}

func (s *OrderSQLService) GetOrder(ctx context.Context, id string) (*models.Order, error) {
	var ord models.Order
	err := s.db.QueryRowContext(ctx, "SELECT id, customer_id, session_id, total_amount, status, created_at, updated_at FROM orders WHERE id=$1", id).
		Scan(&ord.ID, &ord.CustomerID, &ord.SessionID, &ord.TotalAmount, &ord.Status, &ord.CreatedAt, &ord.UpdatedAt)
	if err != nil {
		return nil, err
	}
	items, err := s.GetOrderItems(ctx, id)
	if err != nil {
		return nil, err
	}
	ord.Items = items
	return &ord, nil
}

func (s *OrderSQLService) GetOrderItems(ctx context.Context, orderID string) ([]models.OrderItem, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, order_id, menu_item_id, name, price, quantity, total_price, special_instructions FROM order_items WHERE order_id=$1", orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.OrderItem
	for rows.Next() {
		var it models.OrderItem
		if err := rows.Scan(&it.ID, &it.OrderID, &it.MenuItemID, &it.Name, &it.Price, &it.Quantity, &it.TotalPrice, &it.SpecialInstructions); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, nil
}

// List orders by customer id
func (s *OrderSQLService) ListOrdersByCustomer(ctx context.Context, customerID string) ([]*models.Order, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, customer_id, session_id, total_amount, status, created_at, updated_at FROM orders WHERE customer_id=$1 ORDER BY created_at DESC", customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.CustomerID, &o.SessionID, &o.TotalAmount, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		res = append(res, &o)
	}
	return res, nil
}

// Reorder creates a new order from a previous order id (quick reorder)
func (s *OrderSQLService) Reorder(ctx context.Context, orderID string) (*models.Order, error) {
	ord, err := s.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	// build create items
	var items []CreateOrderItemReq
	for _, it := range ord.Items {
		items = append(items, CreateOrderItemReq{MenuItemID: it.MenuItemID, Quantity: it.Quantity, SpecialInstructions: it.SpecialInstructions})
	}
	return s.CreateOrder(ctx, ord.CustomerID, items, ord.SessionID)
}

func (s *OrderSQLService) ListOrders(ctx context.Context) ([]*models.Order, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, customer_id, session_id, total_amount, status, created_at, updated_at FROM orders ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.CustomerID, &o.SessionID, &o.TotalAmount, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		res = append(res, &o)
	}
	return res, nil
}

func (s *OrderSQLService) UpdateOrderStatus(ctx context.Context, id string, status models.OrderStatus) error {
	_, err := s.db.ExecContext(ctx, "UPDATE orders SET status=$1, updated_at=$2 WHERE id=$3", status, time.Now(), id)
	return err
}

// SetOrderETA updates the estimated ready time for an order
func (s *OrderSQLService) SetOrderETA(ctx context.Context, id string, eta time.Time) error {
	_, err := s.db.ExecContext(ctx, "UPDATE orders SET estimated_ready_at=$1, updated_at=now() WHERE id=$2", eta, id)
	return err
}

// SyncOrders accepts a batch of orders (for offline sync) and creates them in a transaction.
func (s *OrderSQLService) SyncOrders(ctx context.Context, customerID string, itemsBatch [][]CreateOrderItemReq, sessionID string) ([]*models.Order, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var created []*models.Order
	for _, items := range itemsBatch {
		// use same logic as CreateOrder but within tx
		orderID := uuid.New().String()
		now := time.Now()
		var total float64
		var orderItems []models.OrderItem
		for _, it := range items {
			var mi models.MenuItem
			e := tx.QueryRowContext(ctx, "SELECT id, name, price FROM menu_items WHERE id=$1 AND available=TRUE", it.MenuItemID).
				Scan(&mi.ID, &mi.Name, &mi.Price)
			if e != nil {
				err = e
				return nil, err
			}
			lineTotal := mi.Price * float64(it.Quantity)
			total += lineTotal
			orderItems = append(orderItems, models.OrderItem{ID: uuid.New().String(), OrderID: orderID, MenuItemID: mi.ID, Name: mi.Name, Price: mi.Price, Quantity: it.Quantity, TotalPrice: lineTotal, SpecialInstructions: it.SpecialInstructions})
		}

		if sessionID != "" {
			_, err = tx.ExecContext(ctx, "INSERT INTO orders (id, customer_id, session_id, total_amount, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)", orderID, customerID, sessionID, total, string(models.OrderStatusPending), now, now)
		} else {
			_, err = tx.ExecContext(ctx, "INSERT INTO orders (id, customer_id, total_amount, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6)", orderID, customerID, total, string(models.OrderStatusPending), now, now)
		}
		if err != nil {
			return nil, err
		}
		for _, oi := range orderItems {
			_, err = tx.ExecContext(ctx, "INSERT INTO order_items (id, order_id, menu_item_id, name, price, quantity, total_price, special_instructions) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)", oi.ID, oi.OrderID, oi.MenuItemID, oi.Name, oi.Price, oi.Quantity, oi.TotalPrice, oi.SpecialInstructions)
			if err != nil {
				return nil, err
			}
		}
		ord := &models.Order{ID: orderID, CustomerID: customerID, Items: orderItems, TotalAmount: total, Status: models.OrderStatusPending, CreatedAt: now, UpdatedAt: now}
		if sessionID != "" {
			ord.SessionID = sessionID
		}
		created = append(created, ord)
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return created, nil
}
