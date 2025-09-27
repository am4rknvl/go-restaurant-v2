package services

import (
	"context"
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

type AuthBasicService struct{ db *sql.DB }

func NewAuthBasicService(db *sql.DB) *AuthBasicService { return &AuthBasicService{db: db} }

func (s *AuthBasicService) Signup(ctx context.Context, phone string, password string, role string) error {
	if phone == "" || password == "" {
		return sql.ErrNoRows
	}
	if role == "" {
		role = "customer"
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO accounts (id, phone_number, password_hash, role) VALUES (gen_random_uuid()::text, $1, $2, $3) ON CONFLICT (phone_number) DO UPDATE SET password_hash = EXCLUDED.password_hash, role = EXCLUDED.role",
		phone, string(hash), role,
	)
	return err
}

// Signin returns account id and role
func (s *AuthBasicService) Signin(ctx context.Context, phone string, password string) (string, string, error) {
	var id, hash, role string
	if err := s.db.QueryRowContext(ctx, "SELECT id, password_hash, COALESCE(role,'customer') FROM accounts WHERE phone_number=$1", phone).Scan(&id, &hash, &role); err != nil {
		return "", "", err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return "", "", sql.ErrNoRows
	}
	return id, role, nil
}
