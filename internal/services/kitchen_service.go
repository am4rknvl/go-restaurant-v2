package services

import (
	"fmt"
	"restaurant-system/internal/database"
	"restaurant-system/internal/models"
	"time"
)

type KitchenService struct {
	db *database.DB
}

func NewKitchenService(db *database.DB) *KitchenService {
	return &KitchenService{db: db}
}

func (s *KitchenService) GetPendingOrders() ([]*models.Order, error) {
	// Get orders that are confirmed or preparing
	rows, err := s.db.Conn().Query(
		"SELECT id, customer_id, total_amount, status, created_at, updated_at FROM orders WHERE status IN ('confirmed', 'preparing') ORDER BY created_at ASC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(&order.ID, &order.CustomerID, &order.TotalAmount, &order.Status, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			return nil, err
		}

		// Get order items
		order.Items, err = s.db.GetOrderItems(order.ID)
		if err != nil {
			return nil, err
		}

		orders = append(orders, &order)
	}

	return orders, nil
}

func (s *KitchenService) UpdateOrderStatus(orderID string, status models.OrderStatus) error {
	// Validate status transition
	currentStatus, err := s.getOrderStatus(orderID)
	if err != nil {
		return err
	}

	if !s.isValidStatusTransition(currentStatus, status) {
		return fmt.Errorf("invalid status transition from %s to %s", currentStatus, status)
	}

	_, err = s.db.Conn().Exec(
		"UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3",
		status, time.Now(), orderID,
	)

	return err
}

func (s *KitchenService) getOrderStatus(orderID string) (models.OrderStatus, error) {
	var status models.OrderStatus
	err := s.db.Conn().QueryRow(
		"SELECT status FROM orders WHERE id = $1",
		orderID,
	).Scan(&status)

	return status, err
}

func (s *KitchenService) isValidStatusTransition(current, new models.OrderStatus) bool {
	// Define valid status transitions
	validTransitions := map[models.OrderStatus][]models.OrderStatus{
		models.OrderStatusPending:   {models.OrderStatusConfirmed, models.OrderStatusCancelled},
		models.OrderStatusConfirmed: {models.OrderStatusPreparing, models.OrderStatusCancelled},
		models.OrderStatusPreparing: {models.OrderStatusReady, models.OrderStatusCancelled},
		models.OrderStatusReady:     {models.OrderStatusCompleted},
		models.OrderStatusCompleted: {},
		models.OrderStatusCancelled: {},
	}

	allowedStatuses, exists := validTransitions[current]
	if !exists {
		return false
	}

	for _, allowed := range allowedStatuses {
		if allowed == new {
			return true
		}
	}

	return false
}

func (s *KitchenService) GetOrderDetails(orderID string) (*models.Order, error) {
	var order models.Order
	err := s.db.Conn().QueryRow(
		"SELECT id, customer_id, total_amount, status, created_at, updated_at FROM orders WHERE id = $1",
		orderID,
	).Scan(&order.ID, &order.CustomerID, &order.TotalAmount, &order.Status, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		return nil, err
	}

	// Get order items
	order.Items, err = s.db.GetOrderItems(orderID)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (s *KitchenService) GetOrdersByStatus(status models.OrderStatus) ([]*models.Order, error) {
	rows, err := s.db.Conn().Query(
		"SELECT id, customer_id, total_amount, status, created_at, updated_at FROM orders WHERE status = $1 ORDER BY created_at ASC",
		status,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(&order.ID, &order.CustomerID, &order.TotalAmount, &order.Status, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			return nil, err
		}

		// Get order items
		order.Items, err = s.db.GetOrderItems(order.ID)
		if err != nil {
			return nil, err
		}

		orders = append(orders, &order)
	}

	return orders, nil
}
