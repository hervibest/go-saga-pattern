package contract

import (
	"context"
	"go-saga-pattern/commoner/web"
	"go-saga-pattern/transaction-svc/internal/model"
)

type TransactionUseCase interface {
	CreateTransaction(ctx context.Context, request *model.CreateTransactionRequest) (*model.CreateTransactionResponse, error)
	CheckAndUpdateTransaction(ctx context.Context, request *model.CheckAndUpdateTransactionRequest) error
	UserSearch(ctx context.Context, request *model.UserSearchTransactionRequest) ([]*model.TransactionResponse, *web.PageMetadata, error)
	UserSearchWithDetail(ctx context.Context, request *model.UserSearchTransactionRequest) ([]*model.TransactionResponse, *web.PageMetadata, error)
	OwnerSearchWithDetail(ctx context.Context, request *model.OwnerSearchTransactionRequest) ([]*model.TransactionResponse, *web.PageMetadata, error)
}
