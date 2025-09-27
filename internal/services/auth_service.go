package services

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"restaurant-system/internal/database"
	"restaurant-system/internal/models"

	"github.com/google/uuid"
)

type AuthService struct {
	db *database.DB
}

func NewAuthService(db *database.DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) RequestOTP(req *models.RequestOTPRequest) error {
	// Check authorized device
	var exists int
	err := s.db.Conn().QueryRow("SELECT 1 FROM authorized_devices WHERE device_id = $1", req.DeviceID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("unauthorized device")
	}

	code, err := generateOTPCode(6)
	if err != nil {
		return err
	}

	otp := &models.OTP{
		ID:          uuid.New().String(),
		PhoneNumber: req.PhoneNumber,
		Code:        code,
		ExpiresAt:   time.Now().Add(5 * time.Minute),
		CreatedAt:   time.Now(),
	}

	_, err = s.db.Conn().Exec(
		"INSERT INTO otps (id, phone_number, code, expires_at, created_at) VALUES ($1, $2, $3, $4, $5)",
		otp.ID, otp.PhoneNumber, otp.Code, otp.ExpiresAt, otp.CreatedAt,
	)
	if err != nil {
		return err
	}

	// TODO: Integrate SMS provider; for now we just simulate success
	return nil
}

func (s *AuthService) VerifyOTP(req *models.VerifyOTPRequest) (string, error) {
	// Check authorized device
	var exists int
	err := s.db.Conn().QueryRow("SELECT 1 FROM authorized_devices WHERE device_id = $1", req.DeviceID).Scan(&exists)
	if err != nil {
		return "", fmt.Errorf("unauthorized device")
	}

	// Validate OTP
	var otpID string
	err = s.db.Conn().QueryRow(
		"SELECT id FROM otps WHERE phone_number = $1 AND code = $2 AND expires_at > NOW() ORDER BY created_at DESC LIMIT 1",
		req.PhoneNumber, req.Code,
	).Scan(&otpID)
	if err != nil {
		return "", fmt.Errorf("invalid or expired code")
	}

	// Ensure account exists
	accountService := NewAccountService(s.db)
	account, err := accountService.GetAccountByPhoneNumber(req.PhoneNumber)
	if err != nil || account == nil || account.ID == "" {
		account, err = accountService.CreateAccount(&models.CreateAccountRequest{PhoneNumber: req.PhoneNumber})
		if err != nil {
			return "", err
		}
	}

	// Create session
	sessionID := uuid.New().String()
	token := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	_, err = s.db.Conn().Exec(
		"INSERT INTO sessions (id, account_id, token, device_id, expires_at, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		sessionID, account.ID, token, req.DeviceID, expiresAt, time.Now(),
	)
	if err != nil {
		return "", err
	}

	return token, nil
}

func generateOTPCode(length int) (string, error) {
	digits := "0123456789"
	code := make([]byte, length)
	for i := 0; i < length; i++ {
		nBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		code[i] = digits[nBig.Int64()]
	}
	return string(code), nil
}
