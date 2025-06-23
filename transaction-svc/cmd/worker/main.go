package main

import (
	"context"
	"fmt"

	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/logs"

	"go-saga-pattern/transaction-svc/internal/adapter"
	"go-saga-pattern/transaction-svc/internal/config"
	consumer "go-saga-pattern/transaction-svc/internal/delivery/consumer/webhook"
	"go-saga-pattern/transaction-svc/internal/delivery/scheduler"
	taskhandler "go-saga-pattern/transaction-svc/internal/delivery/task"
	"go-saga-pattern/transaction-svc/internal/gateway/task"
	"go-saga-pattern/transaction-svc/internal/repository"
	"go-saga-pattern/transaction-svc/internal/repository/store"
	"go-saga-pattern/transaction-svc/internal/usecase"

	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
)

// ISSUE : nil in some usecase (i think it was not a good idea and best practice)
func worker(ctx context.Context) error {
	logger, _ := logs.NewLogger()
	db := config.NewPostgresDatabase()
	defer db.Close()
	js := config.NewJetStream(logger)
	redis := config.NewRedisClient(logger)
	asyncClient := config.NewAsyncConfig()
	midtransClient := config.NewMidtransClient()
	goCronConfig := config.NewGocron(logger)

	config.DeleteWebhookStream(js, logger)
	config.InitWebhookStream(js, logger)

	databaseStore := store.NewDatabaseStore(db)
	messagingAdapter := adapter.NewMessagingAdapter(js)
	cacheAdapter := adapter.NewCacheAdapter(redis)
	paymentAdapter := adapter.NewPaymentAdapter(midtransClient, cacheAdapter, logger)

	customValidator := helper.NewCustomValidator()
	timeParserHelper := helper.NewTimeParserHelper(logger)

	transactionRepo := repository.NewTransactionRepository()
	transactionTransactionRepo := repository.NewTransactionDetailRepository()

	transactionTask := task.NewTransactionTask(asyncClient)

	transactionUC := usecase.NewTransactionUseCase(transactionRepo, transactionTransactionRepo, databaseStore, nil, messagingAdapter,
		paymentAdapter, cacheAdapter, transactionTask, timeParserHelper, customValidator, logger)
	cancelationUC := usecase.NewCancelationUseCase(databaseStore, transactionRepo, messagingAdapter, logger)
	schedulerUC := usecase.NewSchedulerUseCase(databaseStore, transactionRepo, transactionUC, cancelationUC, paymentAdapter, logger)

	transactionConsumer := consumer.NewWebhookConsumer(transactionUC, js, logger)
	go transactionConsumer.Start(ctx)
	serverErrors := make(chan error, 1)

	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr: "127.0.0.1:6379",
			// Password:     "",               // jika pakai password
			// DB:           0,                // optional: default DB
			// TLSConfig:    &tls.Config{},    // jika pakai Redis TLS (bukan HTTPS ya!)
			// DialTimeout:  10 * time.Second, // increased timeout
			// ReadTimeout:  10 * time.Second,
			// WriteTimeout: 10 * time.Second,
		},
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: 2,
			// Optionally specify multiple queues with different priority.
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			// See the godoc for other configuration options
		},
	)

	schedulerRunner := scheduler.NewSchedulerRunner(goCronConfig, schedulerUC, logger)
	go schedulerRunner.Start()

	expireTaskHandler := taskhandler.NewTransactionTaskHandler(cancelationUC, logger)

	mux := asynq.NewServeMux()
	mux.HandleFunc(task.TypeTransactionExpire, expireTaskHandler.HandleExpire)
	mux.HandleFunc(task.TypeTransactionExpireFinal, expireTaskHandler.HandleFinalExpire)
	go func() {
		serverErrors <- srv.Run(mux)
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-serverErrors:
		return err
	}

}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := worker(ctx); err != nil {
		fmt.Printf("Error starting web server: %v\n", err)
		stop()
		return
	}
}
