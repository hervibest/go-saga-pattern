package entity

import (
	"database/sql"
	"encoding/json"
	"go-saga-pattern/commoner/constant/enum"
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID                       uuid.UUID              `db:"id"`
	UserID                   uuid.UUID              `db:"user_id"`
	TotalPrice               float64                `db:"total_price"`
	TransactionStatus        enum.TransactionStatus `db:"transaction_status"`
	InternalStatus           enum.TrxInternalStatus `db:"internal_status"`
	ExternalStatus           sql.NullString         `db:"external_status"`
	ExternalSettlementAt     sql.NullTime           `db:"external_settlement_at"`
	ExternalCallbackResponse *json.RawMessage       `db:"external_callback_response"`
	SnapToken                sql.NullString         `db:"snap_token"`
	CheckoutAt               *time.Time             `db:"checkout_at"`
	PaymentAt                sql.NullTime           `db:"payment_at"`
	UpdatedAt                *time.Time             `db:"updated_at"`
}

type TransactionWithTotal struct {
	ID                       uuid.UUID              `db:"id"`
	UserID                   uuid.UUID              `db:"user_id"`
	TotalPrice               float64                `db:"total_price"`
	TransactionStatus        enum.TransactionStatus `db:"transaction_status"`
	InternalStatus           enum.TrxInternalStatus `db:"internal_status"`
	ExternalStatus           sql.NullString         `db:"external_status"`
	ExternalSettlementAt     sql.NullTime           `db:"external_settlement_at"`
	ExternalCallbackResponse *json.RawMessage       `db:"external_callback_response"`
	SnapToken                sql.NullString         `db:"snap_token"`
	CheckoutAt               *time.Time             `db:"checkout_at"`
	PaymentAt                sql.NullTime           `db:"payment_at"`
	UpdatedAt                *time.Time             `db:"updated_at"`
	Total                    int                    `db:"total"`
}

type TransactionWithDetail struct {
	TransactionID                       uuid.UUID                  `db:"transaction_id"`
	TransactionUserID                   uuid.UUID                  `db:"transaction_user_id"`
	TransactionTotalPrice               float64                    `db:"transaction_total_price"`
	TransactionStatus                   enum.TransactionStatus     `db:"transaction_transaction_status"`
	TransactionInternalStatus           enum.TrxInternalStatus     `db:"transaction_internal_status"`
	TransactionExternalStatus           enum.MidtransPaymentStatus `db:"transaction_external_status"`
	TransactionExternalSettlementAt     *time.Time                 `db:"transaction_external_settlement_at"`
	TransactionExternalCallbackResponse json.RawMessage            `db:"transaction_external_callback_response"`
	TransactionCheckoutAt               *time.Time                 `db:"transaction_checkout_at"`
	TransactionPaymentAt                *time.Time                 `db:"transaction_payment_at"`
	TransactionUpdatedAt                *time.Time                 `db:"transaction_updated_at"`
	TransactionDetailID                 uuid.UUID                  `db:"transaction_detail_id"`
	TransactionDetailProductID          uuid.UUID                  `db:"transaction_detail_product_id"`
	TransactionDetailTransactionID      uuid.UUID                  `db:"transaction_detail_transaction_id"`
	TransactionDetailQuantity           int                        `db:"transaction_detail_quantity"`
	TransactionDetailPrice              float64                    `db:"transaction_detail_price"`
	TransactionDetailCreatedAt          *time.Time                 `db:"transaction_detail_created_at"`
}

type TransactionWithDetailAndTotal struct {
	TransactionID                       uuid.UUID              `db:"transaction_id"`
	TransactionUserID                   uuid.UUID              `db:"transaction_user_id"`
	TransactionTotalPrice               float64                `db:"transaction_total_price"`
	TransactionStatus                   enum.TransactionStatus `db:"transaction_transaction_status"`
	TransactionInternalStatus           enum.TrxInternalStatus `db:"transaction_internal_status"`
	TransactionExternalStatus           *string                `db:"transaction_external_status"`
	TransactionExternalSettlementAt     *time.Time             `db:"transaction_external_settlement_at"`
	TransactionExternalCallbackResponse json.RawMessage        `db:"transaction_external_callback_response"`
	TransactionCheckoutAt               *time.Time             `db:"transaction_checkout_at"`
	TransactionPaymentAt                *time.Time             `db:"transaction_payment_at"`
	TransactionUpdatedAt                *time.Time             `db:"transaction_updated_at"`
	TransactionDetailID                 uuid.UUID              `db:"transaction_detail_id"`
	TransactionDetailTransactionID      uuid.UUID              `db:"transaction_detail_transaction_id"`
	TransactionDetailProductID          uuid.UUID              `db:"transaction_detail_product_id"`
	TransactionDetailQuantity           int                    `db:"transaction_detail_quantity"`
	TransactionDetailPrice              float64                `db:"transaction_detail_price"`
	TransactionDetailCreatedAt          *time.Time             `db:"transaction_detail_created_at"`
	Total                               int                    `db:"total"`
}
