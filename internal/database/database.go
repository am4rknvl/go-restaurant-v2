package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"restaurant-system/internal/models"

	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
}

func (db *DB) Conn() *sql.DB { return db.conn }

func Initialize() (*DB, error) {
	pgURL := os.Getenv("PG_URL")
	if pgURL == "" {
		host := getenvDefault("PG_HOST", "localhost")
		port := getenvDefault("PG_PORT", "5432")
		user := getenvDefault("PG_USER", "postgres")
		password := getenvDefault("PG_PASSWORD", "postgres")
		dbname := getenvDefault("PG_DATABASE", "restaurant")
		sslmode := getenvDefault("PG_SSLMODE", "disable")
		pgURL = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslmode)
	}

	conn, err := sql.Open("postgres", pgURL)
	if err != nil {
		return nil, err
	}

	db := &DB{conn: conn}

	if err := db.createTables(); err != nil {
		return nil, err
	}

	if err := db.seedMenuItems(); err != nil {
		return nil, err
	}

	log.Println("Database initialized successfully")
	return db, nil
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS menu_items (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			price REAL NOT NULL,
			category TEXT NOT NULL,
			available BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS accounts (
			id TEXT PRIMARY KEY,
			phone_number TEXT UNIQUE NOT NULL,
			balance REAL DEFAULT 0,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
			id TEXT PRIMARY KEY,
			customer_id TEXT NOT NULL,
			total_amount REAL NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			FOREIGN KEY (customer_id) REFERENCES accounts(id)
		)`,
		`CREATE TABLE IF NOT EXISTS order_items (
			id TEXT PRIMARY KEY,
			order_id TEXT NOT NULL,
			menu_item_id TEXT NOT NULL,
			name TEXT NOT NULL,
			price REAL NOT NULL,
			quantity INTEGER NOT NULL,
			total_price REAL NOT NULL,
			FOREIGN KEY (order_id) REFERENCES orders(id),
			FOREIGN KEY (menu_item_id) REFERENCES menu_items(id)
		)`,
		`CREATE TABLE IF NOT EXISTS payments (
			id TEXT PRIMARY KEY,
			order_id TEXT NOT NULL,
			amount REAL NOT NULL,
			method TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			transaction_id TEXT,
			phone_number TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			FOREIGN KEY (order_id) REFERENCES orders(id)
		)`,
		`CREATE TABLE IF NOT EXISTS otps (
			id TEXT PRIMARY KEY,
			phone_number TEXT NOT NULL,
			code TEXT NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			account_id TEXT NOT NULL,
			token TEXT NOT NULL,
			device_id TEXT NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			FOREIGN KEY (account_id) REFERENCES accounts(id)
		)`,
		`CREATE TABLE IF NOT EXISTS authorized_devices (
			device_id TEXT PRIMARY KEY,
			name TEXT,
			registered_at TIMESTAMPTZ DEFAULT NOW()
		)`,
	}

	for _, query := range queries {
		if _, err := db.conn.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) seedMenuItems() error {
	// Check if menu items already exist
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM menu_items").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil // Already seeded
	}

	menuItems := []struct {
		id          string
		name        string
		description string
		price       float64
		category    string
	}{
		{"item-1", "Burger Deluxe", "Juicy beef patty with fresh vegetables", 15.99, "Main Course"},
		{"item-2", "Chicken Wings", "Spicy buffalo wings with ranch dip", 12.99, "Appetizer"},
		{"item-3", "Caesar Salad", "Fresh romaine lettuce with caesar dressing", 8.99, "Salad"},
		{"item-4", "Pizza Margherita", "Classic pizza with tomato and mozzarella", 18.99, "Main Course"},
		{"item-5", "Fish & Chips", "Beer-battered fish with crispy fries", 16.99, "Main Course"},
		{"item-6", "Chocolate Cake", "Rich chocolate cake with vanilla ice cream", 6.99, "Dessert"},
		{"item-7", "Fresh Juice", "Orange, apple, or mixed fruit juice", 4.99, "Beverage"},
		{"item-8", "Coffee", "Freshly brewed coffee", 3.99, "Beverage"},
	}

	for _, item := range menuItems {
		_, err := db.conn.Exec(
			"INSERT INTO menu_items (id, name, description, price, category) VALUES ($1, $2, $3, $4, $5)",
			item.id, item.name, item.description, item.price, item.category,
		)
		if err != nil {
			return err
		}
	}

	log.Println("Menu items seeded successfully")
	return nil
}

// Helper to insert order items
func (db *DB) StoreOrderItems(orderID string, items []models.OrderItem) error {
	for _, item := range items {
		_, err := db.conn.Exec(
			"INSERT INTO order_items (id, order_id, menu_item_id, name, price, quantity, total_price) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			item.ID, item.OrderID, item.MenuItemID, item.Name, item.Price, item.Quantity, item.TotalPrice,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// Helper to retrieve order items
func (db *DB) GetOrderItems(orderID string) ([]models.OrderItem, error) {
	rows, err := db.conn.Query(
		"SELECT id, order_id, menu_item_id, name, price, quantity, total_price FROM order_items WHERE order_id = $1",
		orderID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		err := rows.Scan(&item.ID, &item.OrderID, &item.MenuItemID, &item.Name, &item.Price, &item.Quantity, &item.TotalPrice)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}
