package task

import (
	"encoding/json"
	"go-saga-pattern/commoner/utils"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type transactionTask struct {
	transactionTTL      time.Duration
	transactionFinalTTL time.Duration
	asynqClient         *asynq.Client
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
	ttlStr := utils.GetEnv("TRANSACTION_EXPIRATION_TTL")
	ttlInt, err := strconv.Atoi(ttlStr)
	if err != nil || ttlInt <= 0 {
		log.Printf("Invalid TRANSACTION_EXPIRATION_TTL value: %q, defaulting to 60 seconds", ttlStr)
		ttlInt = 60
	}

	ttlFinalStr := utils.GetEnv("TRANSACTION_EXPIRATION_FINAL_TTL")
	ttlFinalInt, err := strconv.Atoi(ttlFinalStr)
	if err != nil || ttlInt <= 0 {
		log.Printf("Invalid TRANSACTION_EXPIRATION_TTL value: %q, defaulting to 120 seconds", ttlStr)
		ttlInt = 120
	}

	return &transactionTask{
		transactionTTL:      time.Duration(ttlInt) * time.Second,
		transactionFinalTTL: time.Duration(ttlFinalInt) * time.Second,
		asynqClient:         asynqClient,
	}
}

func (t *transactionTask) EnqueueTransactionExpire(transactionID uuid.UUID) error {
	task, err := NewTransactionExpireTask(transactionID)
	if err != nil {
		return err
	}
	_, err = t.asynqClient.Enqueue(task, asynq.ProcessIn(t.transactionTTL))
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
	_, err = t.asynqClient.Enqueue(task, asynq.ProcessIn(t.transactionFinalTTL))
	if err != nil {
		return err
	}
	return nil
}
