package entity

import (
	"database/sql"
	"go-saga-pattern/commoner/constant/enum"
	"time"

	"github.com/google/uuid"
)

type ProductTransaction struct {
	TransactionID uuid.UUID                         `db:"id"`
	ProductID     uuid.UUID                         `db:"product_id"`
	Status        enum.ProductTransactionStatusEnum `db:"status"`
	Quantity      int                               `db:"quantity"`
	TotalPrice    float64                           `db:"total_price"`
	ReservedAt    *time.Time                        `db:"reserved_at"`
	CanceledAt    sql.NullTime                      `db:"canceled_at"`
	CommittedAt   sql.NullTime                      `db:"committed_at"`
	ExpiredAt     sql.NullTime                      `db:"expired_at"`
	SettledAt     sql.NullTime                      `db:"settled_at"`
}
