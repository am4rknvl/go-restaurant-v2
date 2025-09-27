package services

import (
	"fmt"
	"restaurant-system/internal/database"
	"restaurant-system/internal/models"
	"time"

	"github.com/google/uuid"
)

type AccountService struct {
	db *database.DB
}

func NewAccountService(db *database.DB) *AccountService {
	return &AccountService{db: db}
}

func (s *AccountService) CreateAccount(req *models.CreateAccountRequest) (*models.Account, error) {
	// Check if account already exists
	var existingID string
	err := s.db.Conn().QueryRow(
		"SELECT id FROM accounts WHERE phone_number = $1",
		req.PhoneNumber,
	).Scan(&existingID)

	if err == nil {
		// Account exists, return it
		return s.GetAccount(existingID)
	}

	// Create new account
	accountID := uuid.New().String()
	account := &models.Account{
		ID:          accountID,
		PhoneNumber: req.PhoneNumber,
		Balance:     0.0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err = s.db.Conn().Exec(
		"INSERT INTO accounts (id, phone_number, balance, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)",
		account.ID, account.PhoneNumber, account.Balance, account.CreatedAt, account.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (s *AccountService) GetAccount(accountID string) (*models.Account, error) {
	var account models.Account
	err := s.db.Conn().QueryRow(
		"SELECT id, phone_number, balance, created_at, updated_at FROM accounts WHERE id = $1",
		accountID,
	).Scan(&account.ID, &account.PhoneNumber, &account.Balance, &account.CreatedAt, &account.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (s *AccountService) GetAccountByPhoneNumber(phoneNumber string) (*models.Account, error) {
	var account models.Account
	err := s.db.Conn().QueryRow(
		"SELECT id, phone_number, balance, created_at, updated_at FROM accounts WHERE phone_number = $1",
		phoneNumber,
	).Scan(&account.ID, &account.PhoneNumber, &account.Balance, &account.CreatedAt, &account.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (s *AccountService) GetAccountBalance(accountID string) (*models.AccountBalanceResponse, error) {
	var balance float64
	var phoneNumber string
	err := s.db.Conn().QueryRow(
		"SELECT balance, phone_number FROM accounts WHERE id = $1",
		accountID,
	).Scan(&balance, &phoneNumber)

	if err != nil {
		return nil, fmt.Errorf("account not found")
	}

	return &models.AccountBalanceResponse{
		AccountID: accountID,
		Balance:   balance,
		Currency:  "USD",
	}, nil
}

func (s *AccountService) UpdateAccountBalance(accountID string, newBalance float64) error {
	_, err := s.db.Conn().Exec(
		"UPDATE accounts SET balance = $1, updated_at = $2 WHERE id = $3",
		newBalance, time.Now(), accountID,
	)
	return err
}

func (s *AccountService) AddToAccountBalance(accountID string, amount float64) error {
	// Get current balance
	var currentBalance float64
	err := s.db.Conn().QueryRow(
		"SELECT balance FROM accounts WHERE id = $1",
		accountID,
	).Scan(&currentBalance)

	if err != nil {
		return fmt.Errorf("account not found")
	}

	// Update with new balance
	newBalance := currentBalance + amount
	return s.UpdateAccountBalance(accountID, newBalance)
}
