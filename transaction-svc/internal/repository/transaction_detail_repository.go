package repository

import (
	"context"
	"fmt"
	"go-saga-pattern/transaction-svc/internal/entity"
	"go-saga-pattern/transaction-svc/internal/repository/store"
	"strings"

	"github.com/georgysavva/scany/v2/pgxscan"
)

type TransactionDetailRepository interface {
	InsertMany(ctx context.Context, db store.Querier, transactionDetails []*entity.TransactionDetail) ([]*entity.TransactionDetail, error)
	// FindByID(ctx context.Context, db store.Querier, id string) (*entity.TransactionDetail, error)
	// FindByUserID(ctx context.Context, db store.Querier, userID string) ([]*entity.TransactionDetail, error)
}
type transactionDetailRepository struct {
}

func NewTransactionDetailRepository() TransactionDetailRepository {
	return &transactionDetailRepository{}
}

// ISSUE #1
func (transactionDetailRepository) InsertMany(ctx context.Context, db store.Querier, transactionDetails []*entity.TransactionDetail) ([]*entity.TransactionDetail, error) {
	query := `
	INSERT INTO 
		transaction_details
		(transaction_id, product_id, quantity, price)
	VALUES `

	var args []interface{}
	var valueStrings []string
	argPos := 1

	for _, td := range transactionDetails {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)",
			argPos, argPos+1, argPos+2, argPos+3))

		args = append(args, td.TransactionID, td.ProductID, td.Quantity, td.Price)
		argPos += 4
	}

	query += strings.Join(valueStrings, ",")
	query += " RETURNING id, transaction_id, product_id, quantity, price, created_at"

	if err := pgxscan.Select(ctx, db, &transactionDetails, query, args...); err != nil {
		return nil, err
	}

	return transactionDetails, nil
}

// func (transactionDetailRepository) FindByID(ctx context.Context, db store.Querier, id string) (*entity.TransactionDetail, error) {

// }

// func (transactionDetailRepository) FindByUserID(ctx context.Context, db store.Querier, userID string) ([]*entity.TransactionDetail, error) {

// }
