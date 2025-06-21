package model

type PaymentSnapshotRequest struct {
	OrderID     string `json:"order_id,omitempty"`
	GrossAmount int64  `json:"gross_amount,omitempty"`
	Email       string `json:"email,omitempty"`
}
