package enum

type TransactionStatus string

var (
	TransactionStatusPending   TransactionStatus = "PENDING"
	TransactionStatusExpired   TransactionStatus = "EXPIRED"
	TransactionStatusSuccess   TransactionStatus = "SUCCESS"
	TransactionStatusCancelled TransactionStatus = "CANCELED"
	TransactionStatusRefunding TransactionStatus = "REFUNDING"
	TransactionStatusRefunded  TransactionStatus = "REFUNDED"
	TransactionStatusFailed    TransactionStatus = "FAILED"
)
