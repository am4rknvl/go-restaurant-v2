package services

import (
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

type TelebirrC2BService struct {
	db     *gorm.DB
	config models.TelebirrC2BConfig
	client *http.Client
}

func NewTelebirrC2BService(db *gorm.DB, config models.TelebirrC2BConfig) *TelebirrC2BService {
	return &TelebirrC2BService{
		db:     db,
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

type H5PayResponse struct {
	Code       string `json:"code"`
	Msg        string `json:"msg"`
	SubCode    string `json:"sub_code"`
	SubMsg     string `json:"sub_msg"`
	H5PayURL   string `json:"h5_pay_url"`
	OutTradeNo string `json:"out_trade_no"`
	TradeNo    string `json:"trade_no"`
}

func (s *TelebirrC2BService) CreateH5Payment(orderID string, amount float64, subject, body string) (*models.TelebirrC2BOrder, error) {
	outTradeNo := fmt.Sprintf("REST_C2B_%s_%d", orderID, time.Now().Unix())
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Create business content
	bizContent := models.TelebirrC2BBizContent{
		OutTradeNo:     outTradeNo,
		Subject:        subject,
		Body:           body,
		TotalAmount:    fmt.Sprintf("%.2f", amount),
		TimeoutExpress: "30m",
		PassbackParams: fmt.Sprintf("order_id=%s", orderID),
	}

	bizContentJSON, err := json.Marshal(bizContent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal biz_content: %v", err)
	}

	// Create order request
	orderReq := models.TelebirrC2BOrderRequest{
		AppID:      s.config.AppID,
		Method:     "telebirr.payment.h5pay",
		Format:     "JSON",
		Charset:    "utf-8",
		SignType:   "RSA2",
		Timestamp:  timestamp,
		Version:    "1.0",
		NotifyURL:  s.config.NotifyURL,
		BizContent: string(bizContentJSON),
	}

	// Generate signature
	sign, err := s.generateSignForOrder(orderReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signature: %v", err)
	}
	orderReq.Sign = sign

	// Make API request to create H5 payment
	h5PayURL, tradeNo, err := s.callH5PayAPI(orderReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create H5 payment: %v", err)
	}

	// Save order to database
	c2bOrder := models.TelebirrC2BOrder{
		ID:             uuid.New().String(),
		OrderID:        orderID,
		OutTradeNo:     outTradeNo,
		Subject:        subject,
		Body:           body,
		TotalAmount:    amount,
		NotifyURL:      s.config.NotifyURL,
		ReturnURL:      s.config.ReturnURL,
		TimeoutExpress: "30m",
		PassbackParams: fmt.Sprintf("order_id=%s", orderID),
		Status:         "pending",
		H5PayURL:       h5PayURL,
		TradeNo:        tradeNo,
	}

	if err := s.db.Create(&c2bOrder).Error; err != nil {
		return nil, fmt.Errorf("failed to save order: %v", err)
	}

	return &c2bOrder, nil
}

func (s *TelebirrC2BService) callH5PayAPI(orderReq models.TelebirrC2BOrderRequest) (string, string, error) {
	// Convert struct to form data
	formData := url.Values{}
	formData.Set("appid", orderReq.AppID)
	formData.Set("method", orderReq.Method)
	formData.Set("format", orderReq.Format)
	formData.Set("charset", orderReq.Charset)
	formData.Set("sign_type", orderReq.SignType)
	formData.Set("sign", orderReq.Sign)
	formData.Set("timestamp", orderReq.Timestamp)
	formData.Set("version", orderReq.Version)
	formData.Set("notify_url", orderReq.NotifyURL)
	formData.Set("biz_content", orderReq.BizContent)

	req, err := http.NewRequest("POST", s.config.UnifiedOrderURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return "", "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	var h5PayResp H5PayResponse
	if err := json.Unmarshal(body, &h5PayResp); err != nil {
		return "", "", fmt.Errorf("failed to parse response: %v", err)
	}

	if h5PayResp.Code != "10000" {
		return "", "", fmt.Errorf("telebirr API error: %s - %s", h5PayResp.Code, h5PayResp.Msg)
	}

	return h5PayResp.H5PayURL, h5PayResp.TradeNo, nil
}

func (s *TelebirrC2BService) ProcessC2BNotification(notification map[string]string) error {
	outTradeNo := notification["out_trade_no"]
	if outTradeNo == "" {
		return fmt.Errorf("missing out_trade_no in notification")
	}

	// Verify signature
	if !s.verifyC2BNotificationSign(notification) {
		return fmt.Errorf("invalid notification signature")
	}

	// Find existing order
	var c2bOrder models.TelebirrC2BOrder
	if err := s.db.Where("out_trade_no = ?", outTradeNo).First(&c2bOrder).Error; err != nil {
		return fmt.Errorf("order not found: %v", err)
	}

	// Parse total amount
	totalAmount := 0.0
	if amountStr, ok := notification["total_amount"]; ok {
		fmt.Sscanf(amountStr, "%f", &totalAmount)
	}

	// Create notification record
	notif := models.TelebirrC2BNotification{
		ID:             uuid.New().String(),
		OutTradeNo:     outTradeNo,
		TradeNo:        notification["trade_no"],
		TradeStatus:    notification["trade_status"],
		TotalAmount:    totalAmount,
		Currency:       notification["currency"],
		PassbackParams: notification["passback_params"],
		Sign:           notification["sign"],
		SignType:       notification["sign_type"],
	}

	if gmtPaymentStr, ok := notification["gmt_payment"]; ok {
		if gmtPayment, err := time.Parse("2006-01-02 15:04:05", gmtPaymentStr); err == nil {
			notif.GmtPayment = gmtPayment
		}
	}

	// Update order status in transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(&notif).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create notification record: %v", err)
	}

	// Update order status based on trade status
	switch notification["trade_status"] {
	case "TRADE_SUCCESS":
		c2bOrder.Status = "completed"
	case "TRADE_CLOSED":
		c2bOrder.Status = "failed"
	case "WAIT_BUYER_PAY":
		c2bOrder.Status = "pending"
	default:
		c2bOrder.Status = "unknown"
	}

	if err := tx.Save(&c2bOrder).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update order status: %v", err)
	}

	tx.Commit()
	return nil
}

func (s *TelebirrC2BService) generateSignForOrder(req models.TelebirrC2BOrderRequest) (string, error) {
	// Create parameter map for signing (exclude sign field)
	params := map[string]string{
		"appid":       req.AppID,
		"method":      req.Method,
		"format":      req.Format,
		"charset":     req.Charset,
		"sign_type":   req.SignType,
		"timestamp":   req.Timestamp,
		"version":     req.Version,
		"notify_url":  req.NotifyURL,
		"biz_content": req.BizContent,
	}

	return s.generateSignFromMap(params)
}

func (s *TelebirrC2BService) generateSignFromMap(params map[string]string) (string, error) {
	// Sort parameters by key
	var keys []string
	for k := range params {
		if k != "sign" && params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// Build sign string
	var signStr strings.Builder
	for i, k := range keys {
		if i > 0 {
			signStr.WriteString("&")
		}
		signStr.WriteString(fmt.Sprintf("%s=%s", k, params[k]))
	}

	return s.rsaSign(signStr.String())
}

func (s *TelebirrC2BService) rsaSign(data string) (string, error) {
	// Parse private key
	block, _ := pem.Decode([]byte(s.config.PrivateKey))
	if block == nil {
		return "", fmt.Errorf("failed to parse private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %v", err)
	}

	// Create signature
	hash := sha256.Sum256([]byte(data))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign: %v", err)
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

func (s *TelebirrC2BService) verifyC2BNotificationSign(params map[string]string) bool {
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

func (s *TelebirrC2BService) rsaVerify(data, sign string) bool {
	// Parse public key
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

	// Verify signature
	signature, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false
	}

	hash := sha256.Sum256([]byte(data))
	err = rsa.VerifyPKCS1v15(rsaPublicKey, crypto.SHA256, hash[:], signature)
	return err == nil
}

// DB returns the database instance for handlers to access
func (s *TelebirrC2BService) DB() *gorm.DB {
	return s.db
}

func (s *TelebirrC2BService) GetOrderByOutTradeNo(outTradeNo string) (*models.TelebirrC2BOrder, error) {
	var order models.TelebirrC2BOrder
	if err := s.db.Where("out_trade_no = ?", outTradeNo).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (s *TelebirrC2BService) GetOrdersByOrderID(orderID string) ([]models.TelebirrC2BOrder, error) {
	var orders []models.TelebirrC2BOrder
	if err := s.db.Where("order_id = ?", orderID).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}
