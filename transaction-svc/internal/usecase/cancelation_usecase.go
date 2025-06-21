package usecase

import (
	"context"
	"database/sql"
	"go-saga-pattern/commoner/constant/enum"
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/transaction-svc/internal/adapter"
	"go-saga-pattern/transaction-svc/internal/model/event"
	"go-saga-pattern/transaction-svc/internal/repository"
	"go-saga-pattern/transaction-svc/internal/repository/store"
	"go-saga-pattern/transaction-svc/internal/usecase/contract"
	"strings"
	"time"

	"go.uber.org/zap"
)

type cancelationUseCase struct {
	db               store.DatabaseStore
	transactionRepo  repository.TransactionRepository
	messagingAdapter adapter.MessagingAdapter
	logs             logs.Log
}

func NewCancelationUseCase(db store.DatabaseStore, transactionRepo repository.TransactionRepository, messagingAdapter adapter.MessagingAdapter,
	logs logs.Log) contract.CancelationUseCase {
	return &cancelationUseCase{
		db:               db,
		transactionRepo:  transactionRepo,
		messagingAdapter: messagingAdapter,
		logs:             logs,
	}
}

func (uc *cancelationUseCase) ExpirePendingTransaction(ctx context.Context, transactionId string) error {
	if err := store.BeginTransaction(ctx, uc.logs, uc.db, func(tx store.Transaction) error {
		if err := uc.updateTransactionStatusIfPending(ctx, tx, transactionId, enum.TrxInternalStatusExpired); err != nil {
			return helper.WrapInternalServerError(uc.logs, "expire failed", err)
		}
		return nil
	}); err != nil {
		uc.logs.Error("failed to expired pending transaction", zap.String("transactionId", transactionId), zap.Error(err))
	}

	uc.logs.Info("success expired pending transaction", zap.String("transactionId", transactionId))

	return nil
}

func (uc *cancelationUseCase) ExpireFinalTransaction(ctx context.Context, transactionId string) error {
	if err := store.BeginTransaction(ctx, uc.logs, uc.db, func(tx store.Transaction) error {
		if err := uc.updateTransactionStatusIfPending(ctx, tx, transactionId, enum.TrxInternalStatusExpiredCheckedInvalid); err != nil {
			return helper.WrapInternalServerError(uc.logs, "expire failed", err)
		}
		return nil
	}); err != nil {
		uc.logs.Error("failed to expired final transaction", zap.String("transactionId", transactionId), zap.Error(err))
	}

	uc.logs.Info("success expired final transaction", zap.String("transactionId", transactionId))
	event := &event.TransactionEvent{
		TransactionID: transactionId,
		Status:        enum.TransactionEventExpired,
	}

	if err := uc.messagingAdapter.Publish(ctx, "transaction.expired", event); err != nil {
		return helper.WrapInternalServerError(uc.logs, "failed to publish transaction expired event", err)
	}

	return nil
}

func (uc *cancelationUseCase) CancelPendingTransaction(ctx context.Context, transactionId string) error {
	if err := store.BeginTransaction(ctx, uc.logs, uc.db, func(tx store.Transaction) error {
		if err := uc.updateTransactionStatusIfPending(ctx, tx, transactionId, enum.TrxInternalStatusCancelledByUser); err != nil {
			return helper.WrapInternalServerError(uc.logs, "cancel failed", err)
		}
		return nil
	}); err != nil {
		uc.logs.Error("failed to cancel pending transaction", zap.String("transactionId", transactionId), zap.Error(err))
		return helper.WrapInternalServerError(uc.logs, "cancel pending transaction failed", err)
	}

	uc.logs.Info("success cancelled pending transaction", zap.String("transactionId", transactionId))
	return nil
}

func (uc *cancelationUseCase) updateTransactionStatusIfPending(ctx context.Context, tx store.Querier, transactionId string, status enum.TrxInternalStatus) error {
	transaction, err := uc.transactionRepo.FindByID(ctx, tx, transactionId, true)
	if err != nil {
		return err
	}

	//Make sure that only pending or token ready or expired transactions can be cancelled
	if transaction.InternalStatus != enum.TrxInternalStatusPending &&
		transaction.InternalStatus != enum.TrxInternalStatusTokenReady &&
		transaction.InternalStatus != enum.TrxInternalStatusExpired {
		uc.logs.Warn("transaction is not pending or token ready", zap.String("transactionId", transactionId), zap.String("internalStatus", string(transaction.InternalStatus)))
		return nil
	}

	now := time.Now()
	if status == enum.TrxInternalStatusExpired {
		transaction.TransactionStatus = enum.TransactionStatusExpired
		transaction.InternalStatus = enum.TrxInternalStatusExpired
	} else if status == enum.TrxInternalStatusExpiredCheckedInvalid {
		transaction.TransactionStatus = enum.TransactionStatusExpired
		transaction.InternalStatus = enum.TrxInternalStatusExpiredCheckedInvalid
	} else {
		transaction.TransactionStatus = enum.TransactionStatusCancelled
		transaction.InternalStatus = enum.TrxInternalStatusCancelledByUser
	}
	transaction.SnapToken = sql.NullString{Valid: true, String: ""}
	transaction.UpdatedAt = &now

	if err := uc.transactionRepo.UpdateStatus(ctx, tx, transaction); err != nil {
		if strings.Contains(err.Error(), "no rows affected") {
			uc.logs.Error("[LOGIC UPDATE FAILED] no rows affected when updating transaction status", zap.String("transactionId", transactionId), zap.Error(err))
			return nil
		}
		uc.logs.Error("failed to update transaction status", zap.String("transactionId", transactionId), zap.Error(err))

		return err
	}

	return nil
}
