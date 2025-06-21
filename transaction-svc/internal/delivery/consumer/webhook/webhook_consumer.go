package consumer

import (
	"context"
	"fmt"
	errorcode "go-saga-pattern/commoner/constant/errcode"
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/transaction-svc/internal/model"
	"go-saga-pattern/transaction-svc/internal/model/event"
	"go-saga-pattern/transaction-svc/internal/usecase/contract"
	"time"

	"github.com/bytedance/sonic"
	"go.uber.org/zap"

	"github.com/nats-io/nats.go"
)

type WebhookConsumer struct {
	transactionUseCase contract.TransactionUseCase
	js                 nats.JetStreamContext
	subject            string
	consumerName       string
	durableName        string
	logs               logs.Log
}

func NewWebhookConsumer(transactionUseCase contract.TransactionUseCase, js nats.JetStreamContext, logs logs.Log) *WebhookConsumer {
	return &WebhookConsumer{
		transactionUseCase: transactionUseCase,
		js:                 js,
		subject:            "webhook.notify",
		consumerName:       "webhook_consumer",
		durableName:        "webhook_durable",
		logs:               logs,
	}
}

// Bug potential : ctx signal from main thread
func (s *WebhookConsumer) Start(ctx context.Context) error {
	sub, err := s.js.PullSubscribe(
		s.subject,
		s.durableName,
		nats.BindStream("WEBHOOK_NOTIFY_STREAM"), // ganti dengan stream name
	)
	if err != nil {
		return fmt.Errorf("failed to create pull subscription: %w", err)
	}

	s.logs.Info("Started synchronous subscriber for", zap.String("subject", s.subject))

	func() {
		for {
			select {
			case <-ctx.Done():
				s.logs.Info("Context done, stopping webhook consumer")
				return
			default:
				msgs, err := sub.Fetch(10, nats.MaxWait(2*time.Second))
				if err != nil && err != nats.ErrTimeout {
					s.logs.Error("failed to fetch messages: %v", zap.Error(err))
					continue
				}

				for _, msg := range msgs {
					event := new(event.WebhookNotifyEvent)
					if err := sonic.ConfigFastest.Unmarshal(msg.Data, event); err != nil {
						s.logs.Error("failed to unmarshal message: %v", zap.Error(err))
						_ = msg.Nak()
						continue
					}

					s.logs.Info("Received message for order", zap.String("OrderID", event.OrderID))

					request := &model.CheckAndUpdateTransactionRequest{
						MidtransTransactionStatus: event.MidtransTransactionStatus,
						StatusCode:                event.StatusCode,
						SignatureKey:              event.SignatureKey,
						SettlementTime:            event.SettlementTime,
						OrderID:                   event.OrderID,
						GrossAmount:               event.GrossAmount,
						Body:                      event.Body,
					}

					if err := s.transactionUseCase.CheckAndUpdateTransaction(ctx, request); err != nil {
						s.logs.Warn("failed to process transaction: %v", zap.Error(err), zap.String("OrderID", event.OrderID))
						appErr, ok := err.(*helper.AppError)
						if ok && appErr.Code == errorcode.ErrInternal {
							s.logs.Error("internal server error, retrying message", zap.String("OrderID", event.OrderID))
							_ = msg.Nak()
							continue
						}
						s.logs.Warn("failed to process transaction, acknowledging message from usecase error", zap.String("OrderID", event.OrderID), zap.Error(err))
					}

					if err := msg.Ack(); err != nil {
						s.logs.Error("failed to acknowledge message: %v", zap.Error(err))
					} else {
						s.logs.Info("Message acknowledged", zap.String("OrderID", event.OrderID))
					}
				}
			}
		}
	}()

	return nil
}
