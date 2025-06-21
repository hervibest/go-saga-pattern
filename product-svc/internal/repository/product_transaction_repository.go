package repository

import (
	"context"
	"fmt"
	"go-saga-pattern/commoner/constant/enum"
	"go-saga-pattern/product-svc/internal/entity"
	"go-saga-pattern/product-svc/internal/repository/store"
	"strings"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
)

type ProductTransactionRepository interface {
	FindManyByTrxID(ctx context.Context, db store.Querier, transactionID uuid.UUID, forUpdate bool) ([]*entity.ProductTransaction, error)
	Insert(ctx context.Context, db store.Querier, productTransaction *entity.ProductTransaction) (*entity.ProductTransaction, error)
	UpdateStatus(ctx context.Context, db store.Querier, transactionID uuid.UUID, status enum.ProductTransactionStatusEnum) error
	InsertMany(ctx context.Context, db store.Querier, productTransactions []*entity.ProductTransaction) ([]*entity.ProductTransaction, error)
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

func (r *productTransactionRepository) UpdateStatus(ctx context.Context, db store.Querier, transactionID uuid.UUID,
	status enum.ProductTransactionStatusEnum) error {

	// var returningStatusTime string
	query := "UPDATE product_transactions SET status = $1, "
	switch status {
	case enum.ProductTransactionStatusCanceled:
		query += "canceled_at = now()"
		// returningStatusTime = " RETURNING canceled_at"
	case enum.ProductTransactionStatusExpired:
		query += "expired_at = now()"
		// returningStatusTime = " RETURNING expired_at"
	case enum.ProductTransactionStatusComitted:
		query += "committed_at = now()"
		// returningStatusTime = " RETURNING committed_at"
	case enum.ProductTransactionStatusSettled:
		query += "settled_at = now()"
		// returningStatusTime = " RETURNING settled_at"
	default:
		return fmt.Errorf("invalid status: %s", status)
	}

	query += ", updated_at = now() WHERE transaction_id = $2"
	// query += returningStatusTime

	_, err := db.Exec(ctx, query, status, transactionID)
	if err != nil {
		return err
	}

	return nil
}

func (r *productTransactionRepository) FindManyByTrxID(ctx context.Context, db store.Querier,
	transactionID uuid.UUID, forUpdate bool) ([]*entity.ProductTransaction, error) {

	var productTransactions []*entity.ProductTransaction
	query := "SELECT * FROM product_transactions WHERE transaction_id = $1 "
	if forUpdate {
		query += "FOR UPDATE"
	}

	if err := pgxscan.Select(ctx, db, &productTransactions, query, transactionID); err != nil {
		return nil, err
	}

	return productTransactions, nil
}

func (r *productTransactionRepository) InsertMany(ctx context.Context, db store.Querier,
	productTransactions []*entity.ProductTransaction) ([]*entity.ProductTransaction, error) {
	query := `
    INSERT INTO product_transactions 
    (transaction_id, product_id, status, quantity, total_price, reserved_at) 
    VALUES `

	var args []interface{}
	var valueStrings []string
	argPos := 1

	for _, pt := range productTransactions {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, now())",
			argPos, argPos+1, argPos+2, argPos+3, argPos+4))

		args = append(args, pt.TransactionID, pt.ProductID, pt.Status, pt.Quantity, pt.TotalPrice)
		argPos += 5
	}

	query += strings.Join(valueStrings, ",")
	query += " RETURNING transaction_id, product_id, status, quantity, total_price, reserved_at"

	if err := pgxscan.Select(ctx, db, &productTransactions, query, args...); err != nil {
		return nil, err
	}

	return productTransactions, nil
}
