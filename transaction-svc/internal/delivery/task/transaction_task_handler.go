package taskhandler

import (
	"context"
	"fmt"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/transaction-svc/internal/gateway/task"
	"go-saga-pattern/transaction-svc/internal/usecase/contract"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type TransactionTaskHandler interface {
	HandleExpire(ctx context.Context, t *asynq.Task) error
	HandleFinalExpire(ctx context.Context, t *asynq.Task) error
}

type transactionTaskHandler struct {
	cancelationUseCase contract.CancelationUseCase
	log                logs.Log
}

func NewTransactionTaskHandler(cancelationUseCase contract.CancelationUseCase, log logs.Log) TransactionTaskHandler {
	return &transactionTaskHandler{
		cancelationUseCase: cancelationUseCase,
		log:                log,
	}
}

func (th *transactionTaskHandler) HandleExpire(ctx context.Context, t *asynq.Task) error {
	var p task.TransactionExpirePayload
	if err := sonic.ConfigFastest.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	if p.TransactionID == uuid.Nil {
		return fmt.Errorf("transaction_id is required: %w", asynq.SkipRetry)
	}

	if err := th.cancelationUseCase.ExpirePendingTransaction(ctx, p.TransactionID.String()); err != nil {
		th.log.Warn("failed to expire pending transaction", zap.Error(err), zap.String("transaction_id", p.TransactionID.String()))
		return fmt.Errorf("failed to expire pending transaction: %w", err)
	}

	return nil
}

func (th *transactionTaskHandler) HandleFinalExpire(ctx context.Context, t *asynq.Task) error {
	var p task.TransactionExpirePayload
	if err := sonic.ConfigFastest.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	if p.TransactionID == uuid.Nil {
		return fmt.Errorf("transaction_id is required: %w", asynq.SkipRetry)
	}

	if err := th.cancelationUseCase.ExpireFinalTransaction(ctx, p.TransactionID.String()); err != nil {
		th.log.Warn("failed to expire final transaction", zap.Error(err), zap.String("transaction_id", p.TransactionID.String()))
		return fmt.Errorf("failed to expire final transaction: %w", err)
	}

	return nil
}
