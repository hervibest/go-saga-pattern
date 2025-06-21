package entity

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID      `db:"id"`
	UserID      uuid.UUID      `db:"user_id"`
	Name        string         `db:"name"`
	Slug        string         `db:"slug"`
	Description sql.NullString `db:"description"`
	Price       float64        `db:"price"`
	Quantity    int            `db:"quantity"`
	CreatedAt   *time.Time     `db:"created_at"`
	UpdatedAt   *time.Time     `db:"updated_at"`
	DeletedAt   sql.NullTime   `db:"deleted_at"`
}

type ProductWithTotal struct {
	ID          uuid.UUID      `db:"id"`
	UserID      uuid.UUID      `db:"user_id"`
	Name        string         `db:"name"`
	Slug        string         `db:"slug"`
	Description sql.NullString `db:"description"`
	Price       float64        `db:"price"`
	Quantity    int            `db:"quantity"`
	CreatedAt   *time.Time     `db:"created_at"`
	UpdatedAt   *time.Time     `db:"updated_at"`
	DeletedAt   sql.NullTime   `db:"deleted_at"`
	TotalData   int            `db:"total_data"`
}
