package event

type WebhookNotifyEvent struct {
	MidtransTransactionStatus string `json:"transaction_status"`
	StatusCode                string `json:"status_code"`
	SignatureKey              string `json:"signature_key"`
	SettlementTime            string `json:"settlement_time"`
	OrderID                   string `json:"order_id"`
	GrossAmount               string `json:"gross_amount"`
	Body                      []byte `json:"body"`
}
