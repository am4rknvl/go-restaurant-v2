package services

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"restaurant-system/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TelebirrService struct {
	db     *gorm.DB
	config models.TelebirrConfig
	client *http.Client
}

func NewTelebirrService(db *gorm.DB, config models.TelebirrConfig) *TelebirrService {
	return &TelebirrService{
		db:     db,
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type OrderRequest struct {
	AppID          string `json:"appid"`
	MerchOrderID   string `json:"merch_order_id"`
	TotalAmount    string `json:"total_amount"`
	Subject        string `json:"subject"`
	Body           string `json:"body"`
	NotifyURL      string `json:"notify_url"`
	ReturnURL      string `json:"return_url"`
	TimeoutExpress string `json:"timeout_express"`
	Nonce          string `json:"nonce"`
	Timestamp      string `json:"timestamp"`
	Sign           string `json:"sign"`
	SignType       string `json:"sign_type"`
}

type OrderResponse struct {
	PrepayID string `json:"prepay_id"`
	Code     string `json:"code"`
	Msg      string `json:"msg"`
}

func (s *TelebirrService) GetValidToken() (*models.TelebirrToken, error) {
	var token models.TelebirrToken
	err := s.db.Where("expires_at > ?", time.Now()).Order("created_at desc").First(&token).Error

	if err == gorm.ErrRecordNotFound {
		return s.acquireNewToken()
	}
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (s *TelebirrService) acquireNewToken() (*models.TelebirrToken, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", s.config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(s.config.AppID, s.config.PrivateKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	token := models.TelebirrToken{
		ID:          uuid.New().String(),
		AccessToken: tokenResp.AccessToken,
		TokenType:   tokenResp.TokenType,
		ExpiresIn:   tokenResp.ExpiresIn,
		ExpiresAt:   time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}

	if err := s.db.Create(&token).Error; err != nil {
		return nil, err
	}

	return &token, nil
}

func (s *TelebirrService) CreatePrepaidOrder(orderID string, amount float64, subject, body string) (*models.TelebirrOrder, error) {
	token, err := s.GetValidToken()
	if err != nil {
		return nil, err
	}

	merchOrderID := fmt.Sprintf("REST_%s_%d", orderID, time.Now().Unix())
	nonce := uuid.New().String()
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	orderReq := OrderRequest{
		AppID:          s.config.AppID,
		MerchOrderID:   merchOrderID,
		TotalAmount:    fmt.Sprintf("%.2f", amount),
		Subject:        subject,
		Body:           body,
		NotifyURL:      s.config.NotifyURL,
		ReturnURL:      s.config.ReturnURL,
		TimeoutExpress: "30m",
		Nonce:          nonce,
		Timestamp:      timestamp,
		SignType:       "RSA2",
	}

	sign, err := s.generateSign(orderReq)
	if err != nil {
		return nil, err
	}
	orderReq.Sign = sign

	jsonData, err := json.Marshal(orderReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", s.config.OrderURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var orderResp OrderResponse
	if err := json.Unmarshal(respBody, &orderResp); err != nil {
		return nil, err
	}

	if orderResp.Code != "0" {
		return nil, fmt.Errorf("telebirr order creation failed: %s", orderResp.Msg)
	}

	telebirrOrder := models.TelebirrOrder{
		ID:             uuid.New().String(),
		OrderID:        orderID,
		PrepayID:       orderResp.PrepayID,
		MerchOrderID:   merchOrderID,
		Amount:         amount,
		Subject:        subject,
		Body:           body,
		NotifyURL:      s.config.NotifyURL,
		ReturnURL:      s.config.ReturnURL,
		TimeoutExpress: "30m",
		Status:         "pending",
	}

	if err := s.db.Create(&telebirrOrder).Error; err != nil {
		return nil, err
	}

	return &telebirrOrder, nil
}

func (s *TelebirrService) GeneratePaymentURL(prepayID string) (string, error) {
	var telebirrOrder models.TelebirrOrder
	if err := s.db.Where("prepay_id = ?", prepayID).First(&telebirrOrder).Error; err != nil {
		return "", err
	}

	rawRequest := map[string]string{
		"prepay_id":      prepayID,
		"merch_order_id": telebirrOrder.MerchOrderID,
		"appid":          s.config.AppID,
		"nonce":          uuid.New().String(),
		"timestamp":      fmt.Sprintf("%d", time.Now().Unix()),
	}

	sign, err := s.generateSignFromMap(rawRequest)
	if err != nil {
		return "", err
	}
	rawRequest["sign"] = sign
	rawRequest["sign_type"] = "RSA2"

	params := url.Values{}
	for k, v := range rawRequest {
		params.Set(k, v)
	}

	paymentURL := fmt.Sprintf("%s?%s", s.config.WebCheckoutURL, params.Encode())

	telebirrOrder.PaymentURL = paymentURL
	s.db.Save(&telebirrOrder)

	return paymentURL, nil
}

func (s *TelebirrService) ProcessNotification(notification map[string]string) error {
	prepayID := notification["prepay_id"]
	if prepayID == "" {
		return fmt.Errorf("missing prepay_id in notification")
	}

	if !s.verifyNotificationSign(notification) {
		return fmt.Errorf("invalid notification signature")
	}

	var telebirrOrder models.TelebirrOrder
	if err := s.db.Where("prepay_id = ?", prepayID).First(&telebirrOrder).Error; err != nil {
		return err
	}

	totalAmount := 0.0
	if amountStr, ok := notification["total_amount"]; ok {
		fmt.Sscanf(amountStr, "%f", &totalAmount)
	}

	notif := models.TelebirrNotification{
		ID:           uuid.New().String(),
		PrepayID:     prepayID,
		MerchOrderID: notification["merch_order_id"],
		TradeNo:      notification["trade_no"],
		TradeStatus:  notification["trade_status"],
		TotalAmount:  totalAmount,
		Currency:     notification["currency"],
		Sign:         notification["sign"],
		SignType:     notification["sign_type"],
	}

	if gmtPaymentStr, ok := notification["gmt_payment"]; ok {
		if gmtPayment, err := time.Parse("2006-01-02 15:04:05", gmtPaymentStr); err == nil {
			notif.GmtPayment = gmtPayment
		}
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(&notif).Error; err != nil {
		tx.Rollback()
		return err
	}

	if notification["trade_status"] == "TRADE_SUCCESS" {
		telebirrOrder.Status = "completed"
	} else if notification["trade_status"] == "TRADE_CLOSED" {
		telebirrOrder.Status = "failed"
	}

	if err := tx.Save(&telebirrOrder).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (s *TelebirrService) generateSign(req OrderRequest) (string, error) {
	params := map[string]string{
		"appid":           req.AppID,
		"merch_order_id":  req.MerchOrderID,
		"total_amount":    req.TotalAmount,
		"subject":         req.Subject,
		"body":            req.Body,
		"notify_url":      req.NotifyURL,
		"return_url":      req.ReturnURL,
		"timeout_express": req.TimeoutExpress,
		"nonce":           req.Nonce,
		"timestamp":       req.Timestamp,
	}

	return s.generateSignFromMap(params)
}

func (s *TelebirrService) generateSignFromMap(params map[string]string) (string, error) {
	var keys []string
	for k := range params {
		if k != "sign" && k != "sign_type" && params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var signStr strings.Builder
	for i, k := range keys {
		if i > 0 {
			signStr.WriteString("&")
		}
		signStr.WriteString(fmt.Sprintf("%s=%s", k, params[k]))
	}

	return s.rsaSign(signStr.String())
}

func (s *TelebirrService) rsaSign(data string) (string, error) {
	block, _ := pem.Decode([]byte(s.config.PrivateKey))
	if block == nil {
		return "", fmt.Errorf("failed to parse private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256([]byte(data))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

func (s *TelebirrService) verifyNotificationSign(params map[string]string) bool {
	sign := params["sign"]
	if sign == "" {
		return false
	}

	signStr, err := s.generateSignFromMap(params)
	if err != nil {
		return false
	}

	return s.rsaVerify(signStr, sign)
}

func (s *TelebirrService) rsaVerify(data, sign string) bool {
	block, _ := pem.Decode([]byte(s.config.PublicKey))
	if block == nil {
		return false
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return false
	}

	signature, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false
	}

	hash := sha256.Sum256([]byte(data))
	err = rsa.VerifyPKCS1v15(rsaPublicKey, crypto.SHA256, hash[:], signature)
	return err == nil
}
