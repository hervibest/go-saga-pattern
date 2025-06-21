package consumer

import (
	"context"
	"fmt"
	errorcode "go-saga-pattern/commoner/constant/errcode"
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/product-svc/internal/model"
	"go-saga-pattern/product-svc/internal/model/event"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func (s *TransactionConsumer) setupConsumer(subject string) error {
	consumerConfig := &nats.ConsumerConfig{
		Durable:       s.durableNames[subject],
		AckPolicy:     nats.AckExplicitPolicy,
		MaxDeliver:    5,
		BackOff:       []time.Duration{1 * time.Second, 5 * time.Second, 10 * time.Second},
		DeliverPolicy: nats.DeliverAllPolicy,
		AckWait:       30 * time.Second,
		FilterSubject: subject,
	}

	_, err := s.js.AddConsumer("TRANSACTION_STREAM", consumerConfig)
	return err
}

func (s *TransactionConsumer) handleMessage(ctx context.Context, msg *nats.Msg) {
	event := new(event.TransactionEvent)
	if err := sonic.ConfigFastest.Unmarshal(msg.Data, event); err != nil {
		s.logs.Error("failed to unmarshal message", zap.Error(err))
		_ = msg.Nak()
		return
	}

	var err error
	switch msg.Subject {
	case "transaction.committed":
		request := &model.CommitProductTransactionsRequest{
			TransactionID: uuid.MustParse(event.TransactionID),
		}
		err = s.transactionUseCase.CommitProductTransactionsRequest(ctx, request)

	case "transaction.settled":
		request := &model.SettleProductTransactionRequest{
			TransactionID: uuid.MustParse(event.TransactionID),
		}
		err = s.transactionUseCase.SettleProducts(ctx, request)

	case "transaction.canceled":
		request := &model.CancelProductTransactionsRequest{
			TransactionID: uuid.MustParse(event.TransactionID),
		}
		err = s.transactionUseCase.CancelProductTransactions(ctx, request)

	case "transaction.expired":
		request := &model.ExpireProductTransactionsRequest{
			TransactionID: uuid.MustParse(event.TransactionID),
		}
		err = s.transactionUseCase.ExpireProductTransactions(ctx, request)

	default:
		err = fmt.Errorf("unknown subject: %s", msg.Subject)
	}

	if err != nil {
		s.handleError(msg, err, event.TransactionID)
		return
	}

	if err := msg.Ack(); err != nil {
		s.logs.Error("failed to ACK message", zap.Error(err))
	}
}

func (s *TransactionConsumer) handleError(msg *nats.Msg, err error, transactionID string) {
	s.logs.Error("failed to process transaction",
		zap.Error(err),
		zap.String("TransactionID", transactionID))

	appErr, ok := err.(*helper.AppError)
	if !ok {
		appErr = &helper.AppError{Code: errorcode.ErrInternal}
	}

	switch appErr.Code {
	case errorcode.ErrInvalidArgument:
		s.logs.Warn("Invalid argument, acknowledging", zap.String("TransactionID", transactionID))
		_ = msg.Ack()
	default:
		delay := 10 * time.Second
		if err := msg.NakWithDelay(delay); err != nil {
			s.logs.Error("failed to NAK message", zap.Error(err))
		}
	}
}
