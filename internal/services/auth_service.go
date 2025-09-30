package services

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"restaurant-system/internal/auth"
	"restaurant-system/internal/config"
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

	// Create session and generate access + refresh tokens
	// refresh token stored in sessions table
	sessionID := uuid.New().String()
	accessToken, refreshToken, err := auth.GenerateAccessAndRefresh(account.ID, "customer")
	if err != nil {
		return "", err
	}

	refreshExpires := time.Now().Add(24 * time.Hour)
	if h := config.Auth().RefreshTokenHours; h > 0 {
		refreshExpires = time.Now().Add(time.Duration(h) * time.Hour)
	}

	_, err = s.db.Conn().Exec(
		"INSERT INTO sessions (id, account_id, token, device_id, refresh_token, expires_at, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		sessionID, account.ID, accessToken, req.DeviceID, refreshToken, refreshExpires, time.Now(),
	)
	if err != nil {
		return "", err
	}

	// return a JSON-style compound token string (client should store both separately)
	// but our handler will return structured JSON instead of this string; keep compatibility
	return accessToken + "::" + refreshToken, nil
}

// RefreshTokens validates a refresh token, issues new access+refresh and stores rotated refresh token
func (s *AuthService) RefreshTokens(refreshToken string, deviceID string) (newAccess string, newRefresh string, err error) {
	// Find session with refresh token
	var sessionID string
	var accountID string
	err = s.db.Conn().QueryRow("SELECT id, account_id FROM sessions WHERE refresh_token = $1 AND device_id = $2 AND expires_at > NOW()", refreshToken, deviceID).Scan(&sessionID, &accountID)
	if err != nil {
		return "", "", err
	}

	newAccess, newRefresh, err = auth.GenerateAccessAndRefresh(accountID, "customer")
	if err != nil {
		return "", "", err
	}

	refreshExpires := time.Now().Add(24 * time.Hour)
	if h := config.Auth().RefreshTokenHours; h > 0 {
		refreshExpires = time.Now().Add(time.Duration(h) * time.Hour)
	}

	_, err = s.db.Conn().Exec("UPDATE sessions SET token = $1, refresh_token = $2, expires_at = $3, updated_at = $4 WHERE id = $5", newAccess, newRefresh, refreshExpires, time.Now(), sessionID)
	if err != nil {
		return "", "", err
	}
	return newAccess, newRefresh, nil
}

// Logout invalidates a refresh token (delete session)
func (s *AuthService) Logout(refreshToken string, deviceID string) error {
	_, err := s.db.Conn().Exec("DELETE FROM sessions WHERE refresh_token = $1 AND device_id = $2", refreshToken, deviceID)
	return err
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
