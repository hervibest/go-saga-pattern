package repository

import (
	"context"
	"go-saga-pattern/commoner/constant/enum"
	"go-saga-pattern/commoner/store"
	"go-saga-pattern/product-svc/internal/entity"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
)

type ProductTransactionRepository interface {
	FindManyByTrxID(ctx context.Context, db store.Querier, transactionID uuid.UUID, forUpdate bool) ([]*entity.ProductTransaction, error)
	Insert(ctx context.Context, db store.Querier, productTransaction *entity.ProductTransaction) (*entity.ProductTransaction, error)
	UpdateStatus(ctx context.Context, db store.Querier, productTransaction *entity.ProductTransaction,
		status enum.ProductTransactionStatusEnum) (*entity.ProductTransaction, error)
}

type productTransactionRepository struct {
}

func NewProductTransactionRepository() ProductTransactionRepository {
	return &productTransactionRepository{}
}

func (r *productTransactionRepository) Insert(ctx context.Context, db store.Querier,
	productTransaction *entity.ProductTransaction) (*entity.ProductTransaction, error) {
	query := `
	INSERT INTO 
	product_transactions 
	(transaction_id, product_id, status, quantity, total_price, reserved_at) 
	VALUES
	($1, now(), $3, $4, $5, now())
	RETURNING id
	`
	if err := pgxscan.Get(ctx, db, productTransaction, query, productTransaction.TransactionID,
		productTransaction.ProductID, productTransaction.Status, productTransaction.Quantity,
		productTransaction.TotalPrice); err != nil {
		return nil, err
	}

	return productTransaction, nil
}

func (r *productTransactionRepository) UpdateStatus(ctx context.Context, db store.Querier,
	productTransaction *entity.ProductTransaction, status enum.ProductTransactionStatusEnum) (*entity.ProductTransaction, error) {

	var returningStatusTime string
	query := "UPDATE product_transactions SET status = $1, "
	switch status {
	case enum.ProductTransactionStatusCanceled:
		query += "canceled_at = now()"
		returningStatusTime = " RETURNING canceled_at"
	case enum.ProductTransactionStatusExpired:
		query += "expired_at = now()"
		returningStatusTime = " RETURNING expired_at"
	case enum.ProductTransactionStatusComitted:
		query += "commited_at = now()"
		returningStatusTime = " RETURNING commited_at"
	case enum.ProductTransactionStatusSettled:
		query += "settled_at = now()"
		returningStatusTime = " RETURNING settled_at"
	}

	query += " WHERE transaction_id = $2 AND product_id = $3"
	query += returningStatusTime

	if err := pgxscan.Get(ctx, db, productTransaction, query, status, productTransaction.TransactionID,
		productTransaction.ProductID); err != nil {
		return nil, err
	}

	return productTransaction, nil
}

func (r *productTransactionRepository) FindManyByTrxID(ctx context.Context, db store.Querier,
	transactionID uuid.UUID, forUpdate bool) ([]*entity.ProductTransaction, error) {

	var productTransactions []*entity.ProductTransaction
	query := "SELECT * FROM product_transations WHERE transaction_id = $1 "
	if forUpdate {
		query += "FOR UPDATE"
	}

	if err := pgxscan.Select(ctx, db, &productTransactions, query, transactionID); err != nil {
		return nil, err
	}

	return productTransactions, nil
}
