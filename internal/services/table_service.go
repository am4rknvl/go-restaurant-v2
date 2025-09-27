package services

import (
	"context"
	"database/sql"
	"errors"

	"restaurant-system/internal/models"
)

type TableService struct {
	db *sql.DB
}

func NewTableService(db *sql.DB) *TableService { return &TableService{db: db} }

func (s *TableService) CreateTable(ctx context.Context, t *models.Table) error {
	if t.Name == "" {
		return errors.New("name required")
	}
	_, err := s.db.ExecContext(ctx, "INSERT INTO tables (id, name, qr_code, created_at) VALUES ($1,$2,$3,now())", t.ID, t.Name, t.QRCode)
	return err
}

func (s *TableService) ListTables(ctx context.Context) ([]models.Table, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, name, qr_code, created_at FROM tables ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []models.Table
	for rows.Next() {
		var t models.Table
		if err := rows.Scan(&t.ID, &t.Name, &t.QRCode, &t.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, t)
	}
	return res, nil
}

func (s *TableService) GetTable(ctx context.Context, id string) (*models.Table, error) {
	var t models.Table
	if err := s.db.QueryRowContext(ctx, "SELECT id, name, qr_code, created_at FROM tables WHERE id=$1", id).Scan(&t.ID, &t.Name, &t.QRCode, &t.CreatedAt); err != nil {
		return nil, err
	}
	return &t, nil
}
