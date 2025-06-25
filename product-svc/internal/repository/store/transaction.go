package store

import (
	"context"
	errorcode "go-saga-pattern/commoner/constant/errcode"
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/logs"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Transaction interface {
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
