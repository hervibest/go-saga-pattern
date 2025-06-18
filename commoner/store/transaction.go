package store

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Transaction interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
