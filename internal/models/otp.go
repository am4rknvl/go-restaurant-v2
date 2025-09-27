package models

import "time"

type OTP struct {
    ID          string    `json:"id"`
    PhoneNumber string    `json:"phone_number"`
    Code        string    `json:"code"`
    ExpiresAt   time.Time `json:"expires_at"`
    CreatedAt   time.Time `json:"created_at"`
}

type RequestOTPRequest struct {
    PhoneNumber string `json:"phone_number" binding:"required"`
    DeviceID    string `json:"device_id" binding:"required"`
}

type VerifyOTPRequest struct {
    PhoneNumber string `json:"phone_number" binding:"required"`
    Code        string `json:"code" binding:"required"`
    DeviceID    string `json:"device_id" binding:"required"`
}
