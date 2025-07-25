package repository

import (
	"context"
	"fmt"
	"go-saga-pattern/commoner/constant/enum"
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/web"
	"go-saga-pattern/transaction-svc/internal/entity"
	"go-saga-pattern/transaction-svc/internal/model"
	"go-saga-pattern/transaction-svc/internal/repository/store"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/lib/pq"
)

type TransactionRepository interface {
	Insert(ctx context.Context, db store.Querier, transaction *entity.Transaction) (*entity.Transaction, error)
	FindByID(ctx context.Context, db store.Querier, id string, forUpdate bool) (*entity.Transaction, error)
	FindByUserID(ctx context.Context, db store.Querier, userID string) ([]*entity.Transaction, error)
	FindDetailByID(ctx context.Context, db store.Querier, userID string) ([]*entity.TransactionWithDetail, error)
	FindManyCheckable(ctx context.Context, tx store.Querier) ([]*entity.Transaction, error)
	FindManyByUserID(ctx context.Context, db store.Querier, request *model.UserSearchTransactionRequest) ([]*entity.TransactionWithTotal, *web.PageMetadata, error)
	FindManyWithDetailByUserID(ctx context.Context, db store.Querier, request *model.UserSearchTransactionRequest) ([]*entity.TransactionWithDetailAndTotal, *web.PageMetadata, error)
	FindManyWithDetailByProductID(ctx context.Context, db store.Querier, request *model.OwnerSearchTransactionRequest) ([]*entity.TransactionWithDetailAndTotal, *web.PageMetadata, error)
	UpdateCallback(ctx context.Context, db store.Querier, transaction *entity.Transaction) error
	UpdateToken(ctx context.Context, tx store.Querier, transaction *entity.Transaction) error
	UpdateStatus(ctx context.Context, tx store.Querier, transaction *entity.Transaction) error
	// FindByProductID(ctx context.Context, db store.Querier, productID string) ([]*entity.Transaction, error)
	// UpdateByID(ctx context.Context, db store.Querier, transaction *entity.Transaction) (*entity.Transaction, error)
	// DeleteByID(ctx context.Context, db store.Querier, id string) error
	// ExistsByID(ctx context.Context, db store.Querier, id string) (bool, error)
}
type transactionRepository struct {
}

func NewTransactionRepository() TransactionRepository {
	return &transactionRepository{}
}

func (r *transactionRepository) Insert(ctx context.Context, db store.Querier, transaction *entity.Transaction) (*entity.Transaction, error) {
	query := `
	INSERT INTO transactions
		(id, user_id, total_price, transaction_status, internal_status)
	VALUES
		($1, $2, $3, $4, $5)
	RETURNING
		checkout_at, updated_at
	`
	if err := pgxscan.Get(ctx, db, transaction, query, transaction.ID, transaction.UserID, transaction.TotalPrice,
		transaction.TransactionStatus, transaction.InternalStatus); err != nil {
		return nil, err
	}

	return transaction, nil
}

func (r *transactionRepository) FindByID(ctx context.Context, db store.Querier, id string, forUpdate bool) (*entity.Transaction, error) {
	query := `
	SELECT
		id, user_id, total_price, transaction_status, internal_status,
		external_status, external_settlement_at, external_callback_response,
		checkout_at, payment_at, updated_at
	FROM
		transactions
	WHERE
		id = $1
	`
	if forUpdate {
		query += " FOR UPDATE"
	}

	transaction := new(entity.Transaction)
	if err := pgxscan.Get(ctx, db, transaction, query, id); err != nil {
		return nil, err
	}

	return transaction, nil
}

func (r *transactionRepository) FindByUserID(ctx context.Context, db store.Querier, userID string) ([]*entity.Transaction, error) {
	query := `
	SELECT
		id, user_id, total_price, transaction_status, internal_status,
		external_status, external_settlement_at, external_callback_response,
		checkout_at, payment_at, updated_at
	FROM
		transactions
	WHERE
		user_id = $1
	`
	var transactions []*entity.Transaction
	if err := pgxscan.Select(ctx, db, &transactions, query, userID); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *transactionRepository) FindDetailByID(ctx context.Context, db store.Querier, userID string) ([]*entity.TransactionWithDetail, error) {
	var transactionWithDetail []*entity.TransactionWithDetail
	query := `
	SELECT
		t.id AS transaction_id
		t.user_id AS transaction_user_id
		t.total_price AS transaction_total_price
		t.transaction_status AS transaction_transaction_status
		t.internal_status AS transaction_internal_status
		t.external_status AS transaction_external_status
		t.external_settlement_at AS transaction_external_settlement_at
		t.external_callback_response AS transaction_external_callback_response
		t.checkout_at AS transaction_checkout_at
		t.payment_at AS transaction_payment_at
		t.updated_at AS transaction_updated_at
		td.id AS transaction_detail_id
		td.transaction_id AS transaction_detail_transaction_id
		td.product_id AS transaction_detail_product_id
		td.price AS transaction_detail_price
		td.quantity AS transaction_detail_quantity
		td.created_at AS transaction_detail_created_at
	FROM 
		transactions AS t
	LEFT JOIN
		transaction_details AS td
	ON
		t.id = td.transaction_id
	WHERE 
		t.id = $1
	`
	if err := pgxscan.Select(ctx, db, &transactionWithDetail, query, userID); err != nil {
		return nil, err
	}

	return transactionWithDetail, nil
}

func (r *transactionRepository) UpdateCallback(ctx context.Context, db store.Querier, transaction *entity.Transaction) error {
	query := `
	UPDATE transactions
	SET
		internal_status = $1,
		transaction_status = $2,
		external_status = $3,
		external_settlement_at = $4,
		external_callback_response = $5,
		payment_at = $6,
		updated_at = now()
	WHERE
		id = $7
	RETURNING
		checkout_at, updated_at
	`
	if err := pgxscan.Get(ctx, db, transaction, query, transaction.InternalStatus, transaction.TransactionStatus, transaction.ExternalStatus,
		transaction.ExternalSettlementAt, transaction.ExternalCallbackResponse, transaction.PaymentAt, transaction.ID); err != nil {
		return err
	}

	return nil
}

func (r *transactionRepository) UpdateToken(ctx context.Context, tx store.Querier, transaction *entity.Transaction) error {
	query := `UPDATE transactions SET internal_status = $1, snap_token = $2, updated_at = now() WHERE id = $3`

	row, err := tx.Exec(ctx, query, transaction.InternalStatus, transaction.SnapToken, transaction.ID)
	affected := row.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}

func (r *transactionRepository) UpdateStatus(ctx context.Context, tx store.Querier, transaction *entity.Transaction) error {
	query := `UPDATE transactions SET transaction_status = $1, internal_status = $2, snap_token = COALESCE($3, snap_token), updated_at = $4 WHERE id = $5`

	row, err := tx.Exec(ctx, query, transaction.TransactionStatus, transaction.InternalStatus, transaction.SnapToken, transaction.UpdatedAt, transaction.ID)
	if err != nil {
		return err
	}

	if row.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected for transaction ID: %s", transaction.ID)
	}

	return nil
}

func (r *transactionRepository) FindManyCheckable(ctx context.Context, tx store.Querier) ([]*entity.Transaction, error) {
	transactions := make([]*entity.Transaction, 0)
	statuses := []enum.TrxInternalStatus{
		enum.TrxInternalStatusExpired,
		enum.TrxInternalStatusPending,
		enum.TrxInternalStatusTokenReady,
	}

	query := `
	SELECT
		id, user_id, transaction_status, checkout_at, payment_at, total_price, internal_status
	FROM
		transactions
	WHERE
		internal_status = ANY($1)
	`

	if err := pgxscan.Select(ctx, tx, &transactions, query, pq.Array(statuses)); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *transactionRepository) FindManyByUserID(ctx context.Context, db store.Querier, request *model.UserSearchTransactionRequest) ([]*entity.TransactionWithTotal, *web.PageMetadata, error) {
	transactions := make([]*entity.TransactionWithTotal, 0)
	query := `
	SELECT 
		id,
		user_id,
		total_price,
		transaction_status,
		internal_status,
		external_status,
		external_settlement_at,
		external_callback_response,
		snap_token,
		checkout_at,
		payment_at,
		updated_at,
		COUNT (*) OVER () as total
	FROM 
		transactions
	WHERE 
		user_id = $1
	LIMIT $2 
	OFFSET $3
	
	`
	if err := pgxscan.Select(ctx, db, &transactions, query, request.UserID, request.Limit, (request.Page-1)*request.Limit); err != nil {
		return nil, nil, err
	}

	if len(transactions) == 0 {
		return nil, &web.PageMetadata{}, nil
	}

	totalItems := transactions[0].Total

	pageMetadata := helper.CalculatePagination(int64(totalItems), request.Page, request.Limit)

	return transactions, pageMetadata, nil

}

func (r *transactionRepository) FindManyWithDetailByUserID(
	ctx context.Context,
	db store.Querier,
	request *model.UserSearchTransactionRequest,
) ([]*entity.TransactionWithDetailAndTotal, *web.PageMetadata, error) {

	// Langkah 1: Hitung total transaksi (bukan total baris hasil join)
	var totalItems int
	countQuery := `
		SELECT COUNT(*)
		FROM transactions
		WHERE user_id = $1
	`
	if err := db.QueryRow(ctx, countQuery, request.UserID).Scan(&totalItems); err != nil {
		return nil, nil, err
	}

	if totalItems == 0 {
		return []*entity.TransactionWithDetailAndTotal{}, helper.CalculatePagination(0, request.Page, request.Limit), nil
	}

	query := `
		SELECT
			t.id AS transaction_id,
			t.user_id AS transaction_user_id,
			t.total_price AS transaction_total_price,
			t.transaction_status AS transaction_transaction_status,
			t.internal_status AS transaction_internal_status,
			t.external_status AS transaction_external_status,
			t.external_settlement_at AS transaction_external_settlement_at,
			t.external_callback_response AS transaction_external_callback_response,
			t.checkout_at AS transaction_checkout_at,
			t.payment_at AS transaction_payment_at,
			t.updated_at AS transaction_updated_at,
			td.id AS transaction_detail_id,
			td.transaction_id AS transaction_detail_transaction_id,
			td.product_id AS transaction_detail_product_id,
			td.quantity AS transaction_detail_quantity,
			td.price AS transaction_detail_price,
			td.created_at AS transaction_detail_created_at
		FROM (
			SELECT *
			FROM transactions
			WHERE user_id = $1
			ORDER BY checkout_at DESC
			LIMIT $2 OFFSET $3
		) AS t
	LEFT JOIN
		transaction_details AS td
	ON
		t.id = td.transaction_id
	ORDER BY t.checkout_at DESC, td.created_at ASC
	`

	transactions := make([]*entity.TransactionWithDetailAndTotal, 0)
	if err := pgxscan.Select(ctx, db, &transactions, query, request.UserID, request.Limit, (request.Page-1)*request.Limit); err != nil {
		return nil, nil, err
	}

	if len(transactions) == 0 {
		return nil, &web.PageMetadata{}, nil
	}

	pageMetadata := helper.CalculatePagination(int64(totalItems), request.Page, request.Limit)

	return transactions, pageMetadata, nil
}

func (r *transactionRepository) FindManyWithDetailByProductID(ctx context.Context, db store.Querier, request *model.OwnerSearchTransactionRequest) ([]*entity.TransactionWithDetailAndTotal, *web.PageMetadata, error) {
	query := `
		SELECT
			COUNT(*) OVER() AS total,
			t.id AS transaction_id,
			t.user_id AS transaction_user_id,
			t.total_price AS transaction_total_price,
			t.transaction_status AS transaction_transaction_status,
			t.internal_status AS transaction_internal_status,
			t.external_status AS transaction_external_status,
			t.external_settlement_at AS transaction_external_settlement_at,
			t.external_callback_response AS transaction_external_callback_response,
			t.checkout_at AS transaction_checkout_at,
			t.payment_at AS transaction_payment_at,
			t.updated_at AS transaction_updated_at,
			td.id AS transaction_detail_id,
			td.transaction_id AS transaction_detail_transaction_id,
			td.product_id AS transaction_detail_product_id,
			td.quantity AS transaction_detail_quantity,
			td.price AS transaction_detail_price,
			td.created_at AS transaction_detail_created_at
		FROM transactions AS t
	JOIN
		transaction_details AS td
	ON
		t.id = td.transaction_id
	WHERE 
		td.product_id = $1
	ORDER BY t.checkout_at DESC, td.created_at ASC
	LIMIT $2 OFFSET $3
	`

	transactions := make([]*entity.TransactionWithDetailAndTotal, 0)
	if err := pgxscan.Select(ctx, db, &transactions, query, request.ProductID, request.Limit, (request.Page-1)*request.Limit); err != nil {
		return nil, nil, err
	}

	if len(transactions) == 0 {
		return nil, &web.PageMetadata{}, nil
	}

	pageMetadata := helper.CalculatePagination(int64(transactions[0].Total), request.Page, request.Limit)

	return transactions, pageMetadata, nil

}

// func (r *transactionRepository) DeleteByID(ctx context.Context, db store.Querier, id string) error {
// 	return
// }

// func (r *transactionRepository) ExistsByID(ctx context.Context, db store.Querier, id string) (bool, error) {
// 	return
// }
