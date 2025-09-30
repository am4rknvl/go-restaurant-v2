package config

import (
	"log"
	"os"
	"strconv"
)

type PaymentsConfig struct {
	MerchantAppID string
	FabricAppID   string
	ShortCode     string
	AppSecret     string
	PrivateKeyPEM string
}

var paymentsConfig PaymentsConfig

type AuthConfig struct {
	JWTSecret          string
	AccessTokenMinutes int
	RefreshTokenHours  int
}

var authConfig AuthConfig

func Load() {
	paymentsConfig = PaymentsConfig{
		MerchantAppID: os.Getenv("TELEBIRR_MERCHANT_APP_ID"),
		FabricAppID:   os.Getenv("TELEBIRR_FABRIC_APP_ID"),
		ShortCode:     os.Getenv("TELEBIRR_SHORT_CODE"),
		AppSecret:     os.Getenv("TELEBIRR_APP_SECRET"),
		PrivateKeyPEM: os.Getenv("TELEBIRR_PRIVATE_KEY_PEM"),
	}

	if paymentsConfig.MerchantAppID == "" || paymentsConfig.FabricAppID == "" || paymentsConfig.ShortCode == "" || paymentsConfig.AppSecret == "" || paymentsConfig.PrivateKeyPEM == "" {
		log.Println("warning: missing Telebirr env vars; payment features may not work")
	}

	// load auth config
	authConfig = AuthConfig{
		JWTSecret:          os.Getenv("JWT_SECRET"),
		AccessTokenMinutes: 15,
		RefreshTokenHours:  24,
	}

	if v := os.Getenv("ACCESS_TOKEN_MINUTES"); v != "" {
		// simple atoi
		if n, err := strconv.Atoi(v); err == nil {
			authConfig.AccessTokenMinutes = n
		}
	}
	if v := os.Getenv("REFRESH_TOKEN_HOURS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			authConfig.RefreshTokenHours = n
		}
	}
	if authConfig.JWTSecret == "" {
		log.Println("warning: JWT_SECRET is not set, using default (not secure) - set JWT_SECRET in production")
	}
}

func Payments() PaymentsConfig {
	return paymentsConfig
}

// Auth returns the loaded auth configuration
func Auth() AuthConfig {
	return authConfig
}

// extend Load to populate authConfig
func init() {
	// noop: actual population happens in Load when env is loaded in main
}
