package enum

type TrxInternalStatus string

var (
	TrxInternalStatusPending               TrxInternalStatus = "PENDING"
	TrxInternalStatusTokenReady            TrxInternalStatus = "TOKEN_READY"
	TrxInternalStatusExpired               TrxInternalStatus = "EXPIRED"
	TrxInternalStatusExpiredCheckedInvalid TrxInternalStatus = "EXPIRED_CHECKED_INVALID"
	TrxInternalStatusExpiredCheckedValid   TrxInternalStatus = "EXPIRED_CHECKED_VALID"
	TrxInternalStatusLateSettlement        TrxInternalStatus = "LATE_SETTLEMENT"
	TrxInternalStatusSettled               TrxInternalStatus = "SETTLED"
	TrxInternalStatusCancelledBySystem     TrxInternalStatus = "CANCELED_BY_SYSTEM"
	TrxInternalStatusCancelledByUser       TrxInternalStatus = "CANCELED_BY_USER"
	TrxInternalStatusRefunded              TrxInternalStatus = "REFUNDED"
	TrxInternalStatusFailed                TrxInternalStatus = "FAILED"
)
