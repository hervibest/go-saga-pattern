package usecase

import (
	"context"
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/transaction-svc/internal/adapter"
	"go-saga-pattern/transaction-svc/internal/entity"
	"go-saga-pattern/transaction-svc/internal/model/converter"
	"go-saga-pattern/transaction-svc/internal/repository"
	"go-saga-pattern/transaction-svc/internal/repository/store"
	"go-saga-pattern/transaction-svc/internal/usecase/contract"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/midtrans/midtrans-go/coreapi"
	"go.uber.org/zap"
)

type SchedulerUseCase interface {
	CheckTransactionStatus(ctx context.Context) error
}

type schedulerUseCase struct {
	db                    store.DatabaseStore
	transactionRepository repository.TransactionRepository
	paymentAdapter        adapter.PaymentAdapter
	transactionUseCase    contract.TransactionUseCase
	cancelationUseCase    contract.CancelationUseCase
	logs                  logs.Log
}

func NewSchedulerUseCase(
	db store.DatabaseStore,
	transactionRepository repository.TransactionRepository,
	transactionUseCase contract.TransactionUseCase,
	cancelationUseCase contract.CancelationUseCase,
	paymentAdapter adapter.PaymentAdapter,
	logs logs.Log,
) SchedulerUseCase {
	return &schedulerUseCase{db: db,
		transactionRepository: transactionRepository,
		transactionUseCase:    transactionUseCase,
		cancelationUseCase:    cancelationUseCase,
		paymentAdapter:        paymentAdapter,
		logs:                  logs}
}

type Job struct {
	response    *coreapi.TransactionStatusResponse
	transaction *entity.Transaction
}

func (uc *schedulerUseCase) CheckTransactionStatus(ctx context.Context) error {
	transactions, err := uc.transactionRepository.FindManyCheckable(ctx, uc.db)
	if err != nil {
		return helper.WrapInternalServerError(uc.logs, "failed to find many checkable transaction", err)
	}

	if len(transactions) == 0 {
		uc.logs.Info("There are no checkable transaction, ignoring check transaction process")
		return nil
	}

	checkJobs := make(chan *entity.Transaction, 10)
	updateJobs := make(chan *Job, 10)
	var wgCheck, wgUpdate sync.WaitGroup

	// Stage 1: Worker pool untuk CheckTransactionStatus
	for i := 0; i < 10; i++ {
		wgCheck.Add(1)
		go func() {
			defer wgCheck.Done()
			for tx := range checkJobs {
				// if tx.CheckoutAt != nil && tx.Status != enum.TransactionStatusExpired {
				// 	if time.Since(*tx.CreatedAt) > 15*time.Minute {
				// 		// Mark transaksi sebagai expired
				// 		err := uc.cancelationUseCase.ExpirePendingTransaction(ctx, tx.Id)
				// 		if err != nil {
				// 			log.Printf("Gagal meng-expire transaksi %s: %v", tx.Id, err)
				// 		}
				// 		continue // skip pengecekan ke payment gateway
				// 	}
				// }
				resp, err := uc.paymentAdapter.CheckTransactionStatus(context.Background(), tx.ID.String())
				if err != nil {
					uc.logs.Error("[SchedulerUseCase] CheckTransactionStatus error:", zap.Error(err), zap.String("transactionId", tx.ID.String()))
					continue
				}

				updateJobs <- &Job{
					response:    resp,
					transaction: tx,
				}
			}
		}()
	}

	// Stage 2: Worker pool untuk CheckAndUpdateTransaction
	for i := 0; i < 5; i++ {
		wgUpdate.Add(1)
		go func() {
			defer wgUpdate.Done()
			for job := range updateJobs {
				jsonValue, err := sonic.ConfigFastest.Marshal(job.response)
				if err != nil {
					uc.logs.Error("[SchedulerUseCase] Failed to marshal response", zap.Error(err), zap.String("transactionId", job.transaction.ID.String()))
					continue
				}

				request := converter.SchedulerReqToCheckAndUpdate(job.response, jsonValue)
				_ = uc.transactionUseCase.CheckAndUpdateTransaction(
					ctx, request,
				)
			}
		}()
	}

	// Kirim job transaksi
	go func() {
		for _, tx := range transactions {
			checkJobs <- tx
		}
		close(checkJobs)
	}()

	// Tutup updateJobs setelah semua check selesai
	go func() {
		wgCheck.Wait()
		close(updateJobs)
	}()

	wgUpdate.Wait()
	return nil
}
