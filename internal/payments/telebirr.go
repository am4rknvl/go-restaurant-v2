package payments

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"restaurant-system/internal/config"
)

type InitiateRequest struct {
	OutTradeNo  string  `json:"outTradeNo"`
	Subject     string  `json:"subject"`
	TotalAmount float64 `json:"totalAmount"`
	ReturnUrl   string  `json:"returnUrl"`
	NotifyUrl   string  `json:"notifyUrl"`
	PhoneNumber string  `json:"msisdn,omitempty"`
}

type InitiateResponse struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	CheckoutURL string `json:"checkoutUrl"`
	TradeNo     string `json:"tradeNo,omitempty"`
}

// InitiatePayment creates a Telebirr payment and returns a checkout URL for the user.
func InitiatePayment(req *InitiateRequest) (*InitiateResponse, error) {
	cfg := config.Payments()
	apiBase := os.Getenv("TELEBIRR_API_BASE")
	if apiBase == "" {
		apiBase = "https://api.telebirr.com"
	}

	// Compose params
	params := map[string]string{
		"appId":       cfg.MerchantAppID,
		"outTradeNo":  req.OutTradeNo,
		"subject":     req.Subject,
		"totalAmount": strconv.FormatFloat(req.TotalAmount, 'f', 2, 64),
		"shortCode":   cfg.ShortCode,
		"nonceStr":    strconv.FormatInt(time.Now().UnixNano(), 10),
		"timestamp":   strconv.FormatInt(time.Now().Unix(), 10),
		"returnUrl":   req.ReturnUrl,
		"notifyUrl":   req.NotifyUrl,
	}
	if req.PhoneNumber != "" {
		params["msisdn"] = req.PhoneNumber
	}

	params["sign"] = signParams(params, cfg.AppSecret)

	bodyBytes, _ := json.Marshal(params)
	httpReq, _ := http.NewRequest(http.MethodPost, apiBase+"/payment/v1/merchantPay", bytes.NewReader(bodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var gatewayResp struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			PayUrl  string `json:"toPayUrl"`
			TradeNo string `json:"tradeNo"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&gatewayResp); err != nil {
		return nil, err
	}
	if gatewayResp.Code != "SUCCESS" && gatewayResp.Code != "200" {
		return nil, errors.New(gatewayResp.Msg)
	}
	return &InitiateResponse{
		Code:        gatewayResp.Code,
		Message:     gatewayResp.Msg,
		CheckoutURL: gatewayResp.Data.PayUrl,
		TradeNo:     gatewayResp.Data.TradeNo,
	}, nil
}

// VerifyCallback verifies Telebirr callback signature.
func VerifyCallback(m map[string]string) bool {
	cfg := config.Payments()

	values := map[string]string{}
	for k, v := range m {
		if k == "sign" {
			continue
		}
		values[k] = v
	}
	expected := signParams(values, cfg.AppSecret)
	return hmac.Equal([]byte(expected), []byte(m["sign"]))
}

func signParams(params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var buf bytes.Buffer
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(params[k])
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(buf.Bytes())
	return hex.EncodeToString(mac.Sum(nil))
}
