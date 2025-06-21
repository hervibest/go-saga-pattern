package enum

type MidtransPaymentStatus string

const (
	PaymentStatusCapture    MidtransPaymentStatus = "capture"
	PaymentStatusSettlement MidtransPaymentStatus = "settlement"
	PaymentStatusPending    MidtransPaymentStatus = "pending"
	PaymentStatusDeny       MidtransPaymentStatus = "deny"
	PaymentStatusCancel     MidtransPaymentStatus = "cancel"
	PaymentStatusExpire     MidtransPaymentStatus = "expire"
	PaymentStatusFailure    MidtransPaymentStatus = "failure"
)
