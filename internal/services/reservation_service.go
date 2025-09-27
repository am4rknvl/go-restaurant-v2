package services

import (
	"context"
	"database/sql"
	"errors"

	"restaurant-system/internal/models"

	"github.com/google/uuid"
)

type ReservationService struct {
	db *sql.DB
}

func NewReservationService(db *sql.DB) *ReservationService { return &ReservationService{db: db} }

func (s *ReservationService) CreateReservation(ctx context.Context, r *models.Reservation) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	if r.AccountID == "" || r.TableID == "" || r.ReservedFor.IsZero() {
		return errors.New("invalid reservation fields")
	}
	_, err := s.db.ExecContext(ctx, "INSERT INTO reservations (id, account_id, table_id, party_size, reserved_for, status, notes, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,now(),now())",
		r.ID, r.AccountID, r.TableID, r.PartySize, r.ReservedFor, r.Status, r.Notes)
	return err
}

func (s *ReservationService) ListUpcoming(ctx context.Context) ([]*models.Reservation, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, account_id, table_id, party_size, reserved_for, status, notes, created_at, updated_at FROM reservations WHERE reserved_for >= now() ORDER BY reserved_for ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*models.Reservation
	for rows.Next() {
		var r models.Reservation
		if err := rows.Scan(&r.ID, &r.AccountID, &r.TableID, &r.PartySize, &r.ReservedFor, &r.Status, &r.Notes, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		res = append(res, &r)
	}
	return res, nil
}

func (s *ReservationService) GetReservation(ctx context.Context, id string) (*models.Reservation, error) {
	var r models.Reservation
	err := s.db.QueryRowContext(ctx, "SELECT id, account_id, table_id, party_size, reserved_for, status, notes, created_at, updated_at FROM reservations WHERE id=$1", id).
		Scan(&r.ID, &r.AccountID, &r.TableID, &r.PartySize, &r.ReservedFor, &r.Status, &r.Notes, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *ReservationService) UpdateReservation(ctx context.Context, r *models.Reservation) error {
	if r.ID == "" {
		return errors.New("id required")
	}
	_, err := s.db.ExecContext(ctx, "UPDATE reservations SET table_id=$1, party_size=$2, reserved_for=$3, status=$4, notes=$5, updated_at=now() WHERE id=$6",
		r.TableID, r.PartySize, r.ReservedFor, r.Status, r.Notes, r.ID)
	return err
}

func (s *ReservationService) CancelReservation(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "UPDATE reservations SET status='cancelled', updated_at=now() WHERE id=$1", id)
	return err
}
