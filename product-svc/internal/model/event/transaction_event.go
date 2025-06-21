package event

type TransactionEvent struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"` // e.g., "committed", "settled", "cancelled", "expired"
}
