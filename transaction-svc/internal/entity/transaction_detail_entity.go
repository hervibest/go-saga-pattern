package entity

import (
	"time"

	"github.com/google/uuid"
)

type TransactionDetail struct {
	ID            uuid.UUID  `db:"id"`
	TransactionID uuid.UUID  `db:"transaction_id"`
	ProductID     uuid.UUID  `db:"product_id"`
	Quantity      int        `db:"quantity"`
	Price         float64    `db:"price"`
	CreatedAt     *time.Time `db:"created_at"`
}
