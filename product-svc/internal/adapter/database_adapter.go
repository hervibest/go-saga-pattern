package adapter

import (
	"context"
	"go-saga-pattern/commoner/store"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DatabaseAdapter interface {
	Begin(ctx context.Context) (store.Transaction, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (store.Transaction, error)
	Close()
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type databaseAdapter struct {
	db store.DB
}

func NewDatabaseAdapter(db store.DB) DatabaseAdapter {
	return &databaseAdapter{
		db: db,
	}
}

func (r *databaseAdapter) Begin(ctx context.Context) (store.Transaction, error) {
	return r.db.Begin(ctx)
}

func (r *databaseAdapter) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (store.Transaction, error) {
	return r.db.BeginTx(ctx, txOptions)
}

func (r *databaseAdapter) Close() {
	if r.db != nil {
		r.db.Close()
	}
}

func (r *databaseAdapter) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return r.db.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

func (r *databaseAdapter) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return r.db.Exec(ctx, sql, arguments...)
}

func (r *databaseAdapter) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return r.db.Query(ctx, sql, args...)
}

func (r *databaseAdapter) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return r.db.QueryRow(ctx, sql, args...)
}
