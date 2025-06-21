package enum

type TransactionEvent string

const (
	TransactionEventCommited  = "COMMITED"
	TransactionEventCancelled = "CANCELLED"
	TransactionEventSettled   = "SETTLED"
	TransactionEventExpired   = "EXPIRED"
)
