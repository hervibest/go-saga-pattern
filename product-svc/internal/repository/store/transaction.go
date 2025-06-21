package store

import (
	"context"
	errorcode "go-saga-pattern/commoner/constant/errcode"
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/logs"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Transaction interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

func BeginTransaction(ctx context.Context, logs logs.Log, db DatabaseStore, fn func(tx Transaction) error) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		// logs.Error(fmt.Sprintf("failed begin transaction %s", err.Error()), &logger.Options{
		// 	IsPrintStack: true,
		// })
		return helper.NewAppError(errorcode.ErrInternal, "Something went wrong. Please try again later", err)
	}

	rolledBack := false
	defer func() {
		if err != nil && !rolledBack {
			_ = tx.Rollback(ctx)
		}
	}()

	if err = fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		rolledBack = true
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		// logs.Error(fmt.Sprintf("failed to commit transaction %s", err.Error()), &logger.Options{
		// 	IsPrintStack: true,
		// })
		return helper.NewAppError(errorcode.ErrInternal, "Something went wrong. Please try again later", err)
	}
	return nil
}

type DatabaseStore interface {
	Begin(ctx context.Context) (Transaction, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (Transaction, error)
	Close()
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type databaseAdapter struct {
	db *pgxpool.Pool
}

func NewDatabaseStore(db *pgxpool.Pool) DatabaseStore {
	return &databaseAdapter{
		db: db,
	}
}

func (r *databaseAdapter) Begin(ctx context.Context) (Transaction, error) {
	return r.db.Begin(ctx)
}

func (r *databaseAdapter) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (Transaction, error) {
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
