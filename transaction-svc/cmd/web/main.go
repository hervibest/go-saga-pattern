package main

import (
	"context"
	"fmt"
	"strconv"

	"go-saga-pattern/commoner/discovery"
	"go-saga-pattern/commoner/discovery/consul"
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/logs"

	"go-saga-pattern/transaction-svc/internal/adapter"
	"go-saga-pattern/transaction-svc/internal/config"
	"go-saga-pattern/transaction-svc/internal/delivery/web/controller"
	"go-saga-pattern/transaction-svc/internal/delivery/web/middleware"
	"go-saga-pattern/transaction-svc/internal/delivery/web/route"
	"go-saga-pattern/transaction-svc/internal/gateway/task"
	"go-saga-pattern/transaction-svc/internal/repository"
	"go-saga-pattern/transaction-svc/internal/repository/store"
	"go-saga-pattern/transaction-svc/internal/usecase"

	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

var (
	app *fiber.App
)

func webServer(ctx context.Context) error {
	serverConfig := config.NewServerConfig()
	app = config.NewApp()

	logger, _ := logs.NewLogger()
	db := config.NewPostgresDatabase()
	defer db.Close()
	js := config.NewJetStream(logger)
	redis := config.NewRedisClient(logger)
	asyncClient := config.NewAsyncConfig()
	midtransClient := config.NewMidtransClient()

	databaseStore := store.NewDatabaseStore(db)
	messagingAdapter := adapter.NewMessagingAdapter(js)
	cacheAdapter := adapter.NewCacheAdapter(redis)
	paymentAdapter := adapter.NewPaymentAdapter(midtransClient, cacheAdapter, logger)

	customValidator := helper.NewCustomValidator()
	timeParserHelper := helper.NewTimeParserHelper(logger)

	registry, err := consul.NewRegistry(serverConfig.ConsulAddr, serverConfig.TransactionSvcName)
	if err != nil {
		logger.Error("Failed to create consul registry for service" + err.Error())
	}

	HTTPServiceID := discovery.GenerateServiceID(serverConfig.TransactionSvcName + "-http")
	httpPortInt, _ := strconv.Atoi(serverConfig.TransactionHTTPPort)

	userAdapter, err := adapter.NewUserAdapter(ctx, registry, logger)
	if err != nil {
		logger.Error("Failed to create user adapter", zap.Error(err))
		return err
	}

	productAdapter, err := adapter.NewProductAdapter(ctx, registry, logger)
	if err != nil {
		logger.Error("Failed to create product adapter", zap.Error(err))
		return err
	}

	err = registry.RegisterService(ctx, serverConfig.TransactionSvcName+"-http", HTTPServiceID, serverConfig.TransactionHTTPAddr, httpPortInt, []string{"http"})
	if err != nil {
		logger.Error("Failed to register transaction service to consul", zap.Error(err))
	}

	go func() {
		<-ctx.Done()
		logger.Info("Context canceled. Deregistering services...")
		registry.DeregisterService(context.Background(), HTTPServiceID)

		logger.Info("Shutting down servers...")
		if err := app.Shutdown(); err != nil {
			logger.Error("Failed to shutdown app server", zap.Error(err))
		}

		logger.Info("Successfully shutdown...")
	}()

	go consul.StartHealthCheckLoop(ctx, registry, HTTPServiceID, serverConfig.TransactionSvcName+"-http", logger)

	transactionRepo := repository.NewTransactionRepository()
	transactionTransactionRepo := repository.NewTransactionDetailRepository()

	transactionTask := task.NewTransactionTask(asyncClient)

	transactionUC := usecase.NewTransactionUseCase(transactionRepo, transactionTransactionRepo, databaseStore, productAdapter, messagingAdapter,
		paymentAdapter, cacheAdapter, transactionTask, timeParserHelper, customValidator, logger)

	transactionController := controller.NewTransactionController(transactionUC, logger)

	userMiddleware := middleware.NewUserAuth(userAdapter, logger)

	TransactionRoute := route.NewTransactionRoute(app, transactionController, userMiddleware)
	TransactionRoute.RegisterRoutes()

	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- app.Listen(fmt.Sprintf("%s:%s", serverConfig.TransactionHTTPAddr, serverConfig.TransactionHTTPPort))
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

	if err := webServer(ctx); err != nil {
		fmt.Printf("Error starting web server: %v\n", err)
		stop()
		return
	}
}
