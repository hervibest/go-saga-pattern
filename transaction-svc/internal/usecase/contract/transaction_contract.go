package contract

import (
	"context"
	"go-saga-pattern/transaction-svc/internal/model"
)

type TransactionUseCase interface {
	CreateTransaction(ctx context.Context, request *model.CreateTransactionRequest) (*model.CreateTransactionResponse, error)
	CheckAndUpdateTransaction(ctx context.Context, request *model.CheckAndUpdateTransactionRequest) error
	// GetTransactionDetail(ctx context.Context, request *model.GetTransactionRequest) (*model.TransactionResponse, error)
}
