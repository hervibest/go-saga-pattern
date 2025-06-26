package model

import (
	"encoding/json"
	"go-saga-pattern/commoner/constant/enum"

	"github.com/google/uuid"
)

type CreateTransactionRequest struct {
	UserID   uuid.UUID            `json:"user_id" validate:"required,uuid"`
	Products []TransactionProduct `json:"products" validate:"required"`
}

type GetTransactionRequest struct {
	UserID        uuid.UUID `json:"user_id" validate:"required,uuid"`
	TransacitonID uuid.UUID `json:"transaction_id" validate:"required,uuid"`
}

type UserSearchTransactionRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required,uuid"`
	Page   int       `validate:"required,min=1"`
	Limit  int       `validate:"required,min=1,max=100"`
}

type OwnerSearchTransactionRequest struct {
	UserID    uuid.UUID `json:"user_id" validate:"required,uuid"`
	ProductID uuid.UUID `json:"product_id" validate:"required,uuid"`
	Page      int       `validate:"required,min=1"`
	Limit     int       `validate:"required,min=1,max=100"`
}

type TransactionProduct struct {
	ProductID uuid.UUID `json:"product_id" validate:"required,uuid"`
	Price     float64   `json:"price" validate:"required,gt=0"`
	Quantity  int       `json:"quantity" validate:"required,gt=0"`
}

type CreateTransactionResponse struct {
	TransactionId string `json:"transaction_id"`
	SnapToken     string `json:"snap_token,omitempty"`
	RedirectURL   string `json:"redirect_url,omitempty"`
}

type TransactionResponse struct {
	ID                 string                       `json:"id"`
	UserID             string                       `json:"user_id"`
	TotalPrice         float64                      `json:"total_price,omitempty"`
	TransactionStatus  enum.TransactionStatus       `json:"transaction_status"`
	CheckoutAt         string                       `json:"checkout_at,omitempty"`
	PaymentAt          string                       `json:"payment_at,omitempty"`
	UpdatedAt          string                       `json:"update_at,omitempty"`
	TransactionDetails []*TransactionDetailResponse `json:"transaction_details,omitempty"`
}

type TransactionDetailResponse struct {
	ID        string  `json:"id"`
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	CreatedAt string  `json:"created_at,omitempty"`
}

type CheckAndUpdateTransactionRequest struct {
	MidtransTransactionStatus string `json:"transaction_status" validate:"required"`
	StatusCode                string `json:"status_code" validate:"required"`
	SignatureKey              string `json:"signature_key" validate:"required"`
	SettlementTime            string `json:"settlement_time" validate:"required"`
	OrderID                   string `json:"order_id" validate:"required"`
	GrossAmount               string `json:"gross_amount" validate:"required"`
	Body                      []byte `json:"-"`
}

type WebhookNotifyRequest struct {
	MidtransTransactionType   string          `json:"transaction_type"`
	MidtransTransactionTime   string          `json:"transaction_time"`
	MidtransTransactionStatus string          `json:"transaction_status" validate:"required"`
	MidtransTransactionID     string          `json:"transaction_id"`
	StatusMessage             string          `json:"status_message"`
	StatusCode                string          `json:"status_code" validate:"required"`
	SignatureKey              string          `json:"signature_key" validate:"required"`
	SettlementTime            string          `json:"settlement_time" validate:"required"`
	ReferenceID               string          `json:"reference_id"`
	PaymentType               string          `json:"payment_type"`
	OrderID                   string          `json:"order_id" validate:"required"`
	Metadata                  json.RawMessage `json:"metadata"`
	MerchantID                string          `json:"merchant_id"`
	GrossAmount               string          `json:"gross_amount" validate:"required"`
	FraudStatus               string          `json:"fraud_status"`
	ExpiryTime                string          `json:"expiry_time"`
	Currency                  string          `json:"currency"`
	Acquirer                  string          `json:"acquirer"`
	Body                      []byte          `json:"-"`
}
