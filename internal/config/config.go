package config

import (
	"log"
	"os"
)

type PaymentsConfig struct {
	MerchantAppID string
	FabricAppID   string
	ShortCode     string
	AppSecret     string
	PrivateKeyPEM string
}

var paymentsConfig PaymentsConfig

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
}

func Payments() PaymentsConfig {
	return paymentsConfig
}
