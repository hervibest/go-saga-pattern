package usecase

import (
	"context"
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/transaction-svc/internal/adapter"
	"go-saga-pattern/transaction-svc/internal/model"
	"go-saga-pattern/transaction-svc/internal/model/event"
)

type ListenerUseCase interface {
	ConsumeAndProduceWebhook(ctx context.Context, request *model.WebhookNotifyRequest) error
}
type listenerUseCase struct {
	messagingAdapter adapter.MessagingAdapter
	log              logs.Log
}

func NewListenerUseCase(messagingAdapter adapter.MessagingAdapter, log logs.Log) ListenerUseCase {
	return &listenerUseCase{messagingAdapter: messagingAdapter, log: log}
}

func (uc *listenerUseCase) ConsumeAndProduceWebhook(ctx context.Context, request *model.WebhookNotifyRequest) error {
	event := &event.WebhookNotifyEvent{
		MidtransTransactionStatus: request.MidtransTransactionStatus,
		StatusCode:                request.StatusCode,
		SignatureKey:              request.SignatureKey,
		SettlementTime:            request.SettlementTime,
		OrderID:                   request.OrderID,
		GrossAmount:               request.GrossAmount,
		Body:                      request.Body,
	}

	if err := uc.messagingAdapter.Publish(ctx, "webhook.notify", event); err != nil {
		return helper.WrapInternalServerError(uc.log, "failed to publish webhook notify event", err)
	}

	return nil
}
