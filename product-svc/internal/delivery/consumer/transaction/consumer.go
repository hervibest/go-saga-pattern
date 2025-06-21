package consumer

import (
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/product-svc/internal/usecase"

	"github.com/nats-io/nats.go"
)

// consumer.go
type TransactionConsumer struct {
	transactionUseCase usecase.ProductTransactionUseCase
	js                 nats.JetStreamContext
	logs               logs.Log
	subjects           []string
	durableNames       map[string]string
}

func NewTransactionConsumer(
	transactionUseCase usecase.ProductTransactionUseCase,
	js nats.JetStreamContext,
	logs logs.Log,
) *TransactionConsumer {
	return &TransactionConsumer{
		transactionUseCase: transactionUseCase,
		js:                 js,
		logs:               logs,
		subjects: []string{
			"transaction.committed",
			"transaction.settled",
			"transaction.canceled",
			"transaction.expired",
		},
		durableNames: map[string]string{
			"transaction.committed": "transaction_committed_consumer",
			"transaction.settled":   "transaction_settled_consumer",
			"transaction.canceled":  "transaction_canceled_consumer",
			"transaction.expired":   "transaction_expired_consumer",
		},
	}
}
