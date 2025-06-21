package adapter

import (
	"context"
	"fmt"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/transaction-svc/internal/config"
	"go-saga-pattern/transaction-svc/internal/model"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/sony/gobreaker"
)

type PaymentAdapter interface {
	CreateSnapshot(ctx context.Context, request *model.PaymentSnapshotRequest) (*snap.Response, error)
	CheckTransactionStatus(ctx context.Context, transactionId string) (*coreapi.TransactionStatusResponse, error)
	GetPaymentServerKey() string
}

type paymentAdapter struct {
	midtransClient *config.MidtransClient
	circuitBreaker *gobreaker.CircuitBreaker
	cacheAdapter   CacheAdapter
	logs           logs.Log
}

func NewPaymentAdapter(midtransClient *config.MidtransClient, cacheAdapter CacheAdapter, logs logs.Log) PaymentAdapter {
	cbSettings := gobreaker.Settings{
		Name:        "MidtransSnapshot",
		MaxRequests: 3,
		Interval:    60 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.Requests >= 5 && float64(counts.TotalFailures)/float64(counts.Requests) >= 0.6
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logs.Error(fmt.Sprintf("[CircuitBreaker] %s: %s -> %s", name, from.String(), to.String()))

			if from == gobreaker.StateOpen && (to == gobreaker.StateClosed || to == gobreaker.StateHalfOpen) {
				// publish recovery signal ke Redis Stream
				err := cacheAdapter.XAdd(context.Background(), &redis.XAddArgs{
					Stream: "midtrans:recovery",
					Values: map[string]interface{}{
						"circuit": name,
						"state":   to.String(),
						"time":    time.Now().Unix(),
					},
				})
				if err != nil {
					logs.Error(fmt.Sprintf("failed to publish circuit recovery event: %v", err))
				}
			}
		},
	}

	cb := gobreaker.NewCircuitBreaker(cbSettings)

	return &paymentAdapter{
		midtransClient: midtransClient,
		circuitBreaker: cb,
		logs:           logs,
		cacheAdapter:   cacheAdapter,
	}
}

func (a *paymentAdapter) CreateSnapshot(ctx context.Context, request *model.PaymentSnapshotRequest) (*snap.Response, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  request.OrderID,
			GrossAmt: request.GrossAmount,
		},
		CustomerDetail: &midtrans.CustomerDetails{
			Email: request.Email,
		},
	}

	res, err := a.circuitBreaker.Execute(func() (interface{}, error) {
		resultChan := make(chan *snap.Response, 1)
		errChan := make(chan error, 1)

		go func() {
			resp, err := a.midtransClient.Snap.CreateTransaction(snapReq)
			if err != nil {
				errChan <- err
				return
			}
			resultChan <- resp
		}()

		select {
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("midtrans request timeout: %w", timeoutCtx.Err())
		case err := <-errChan:
			return nil, err
		case resp := <-resultChan:
			return resp, nil
		}
	})

	if err != nil {
		a.logs.Error("[PaymentAdapter] CreateSnapshot error:", zap.Error(err))
		return nil, fmt.Errorf("midtrans create snapshot error: %w", err)
	}

	return res.(*snap.Response), nil
}

func (a *paymentAdapter) CheckTransactionStatus(ctx context.Context, transactionId string) (*coreapi.TransactionStatusResponse, error) {
	a.logs.Info("[PaymentAdapter] CheckTransactionStatus called", zap.String("transactionId", transactionId))
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := a.circuitBreaker.Execute(func() (interface{}, error) {
		resultChan := make(chan *coreapi.TransactionStatusResponse, 1)
		errChan := make(chan error, 1)

		go func() {
			resp, err := a.midtransClient.CoreApi.CheckTransaction(transactionId)
			if err != nil {
				errChan <- err
				return
			}
			resultChan <- resp
		}()

		select {
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("midtrans request timeout: %w", timeoutCtx.Err())
		case err := <-errChan:
			return nil, err
		case resp := <-resultChan:
			return resp, nil
		}
	})

	if err != nil {
		a.logs.Error("[PaymentAdapter] CheckTransactionStatus error:", zap.Error(err))
		return nil, fmt.Errorf("midtrans check status transaction error: %w", err)
	}

	return res.(*coreapi.TransactionStatusResponse), nil
}

func (a *paymentAdapter) GetPaymentServerKey() string {
	return a.midtransClient.Snap.ServerKey
}
