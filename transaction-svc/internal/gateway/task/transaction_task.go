package task

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type transactionTask struct {
	asynqClient *asynq.Client
}

// A list of task types.
const (
	TypeTransactionExpire      = "transaction:expire"
	TypeTransactionExpireFinal = "transaction:expire:final"
)

type TransactionExpirePayload struct {
	TransactionID uuid.UUID `json:"transaction_id"`
}

func NewTransactionExpireTask(transactionID uuid.UUID) (*asynq.Task, error) {
	payload, err := json.Marshal(TransactionExpirePayload{TransactionID: transactionID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeTransactionExpire, payload), nil
}

func NewTransactionExpireFinalTask(transactionID uuid.UUID) (*asynq.Task, error) {
	payload, err := json.Marshal(TransactionExpirePayload{TransactionID: transactionID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeTransactionExpireFinal, payload), nil
}

type TransactionTask interface {
	EnqueueTransactionExpire(transactionID uuid.UUID) error
	EnqueueTransactionExpireFinal(transactionID uuid.UUID) error
}

func NewTransactionTask(asynqClient *asynq.Client) TransactionTask {
	return &transactionTask{asynqClient: asynqClient}
}

func (t *transactionTask) EnqueueTransactionExpire(transactionID uuid.UUID) error {
	task, err := NewTransactionExpireTask(transactionID)
	if err != nil {
		return err
	}
	_, err = t.asynqClient.Enqueue(task, asynq.ProcessIn(10*time.Second))
	if err != nil {
		return err
	}
	return nil
}

func (t *transactionTask) EnqueueTransactionExpireFinal(transactionID uuid.UUID) error {
	task, err := NewTransactionExpireFinalTask(transactionID)
	if err != nil {
		return err
	}
	_, err = t.asynqClient.Enqueue(task, asynq.ProcessIn(30*time.Second))
	if err != nil {
		return err
	}
	return nil
}
