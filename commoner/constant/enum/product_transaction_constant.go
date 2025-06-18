package enum

type ProductTransactionStatusEnum string

const (
	ProductTransactionStatusReserved ProductTransactionStatusEnum = "RESERVED"
	ProductTransactionStatusCanceled ProductTransactionStatusEnum = "CANCELED"
	ProductTransactionStatusComitted ProductTransactionStatusEnum = "COMMITED"
	ProductTransactionStatusExpired  ProductTransactionStatusEnum = "EXPIRED"
	ProductTransactionStatusSettled  ProductTransactionStatusEnum = "SETTLED"
)
