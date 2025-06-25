package main

import (
	"context"
	"fmt"

	"go-saga-pattern/commoner/discovery"
	"go-saga-pattern/commoner/discovery/consul"
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/product-svc/internal/adapter"
	"go-saga-pattern/product-svc/internal/config"
	consumer "go-saga-pattern/product-svc/internal/delivery/consumer/transaction"
	grpcHandler "go-saga-pattern/product-svc/internal/delivery/grpc/handler"
	"go-saga-pattern/product-svc/internal/delivery/web/controller"
	"go-saga-pattern/product-svc/internal/delivery/web/middleware"
	"go-saga-pattern/product-svc/internal/delivery/web/route"
	"go-saga-pattern/product-svc/internal/repository/store"

	"go-saga-pattern/product-svc/internal/repository"
	"go-saga-pattern/product-svc/internal/usecase"
	"net"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	grpcServer *grpc.Server
	app        *fiber.App
)

func webServer(ctx context.Context) error {
	serverConfig := config.NewServerConfig()
	app = config.NewApp()

	logger, _ := logs.NewLogger()
	db := config.NewPostgresDatabase()
	defer db.Close()

	databaseStore := store.NewDatabaseStore(db)
	jetStreamConfig := config.NewJetStream(logger)
	config.DeleteTransactionStream(jetStreamConfig, logger)
	config.InitTransactionStream(jetStreamConfig, logger)

	customValidator := helper.NewCustomValidator()

	registry, err := consul.NewRegistry(serverConfig.ConsulAddr, serverConfig.ProductSvcName)
	if err != nil {
		logger.Error("Failed to create consul registry for service" + err.Error())
	}

	GRPCserviceID := discovery.GenerateServiceID(serverConfig.ProductSvcName + "-grpc")
	grpcPortInt, _ := strconv.Atoi(serverConfig.ProductGRPCPort)

	userAdapter, err := adapter.NewUserAdapter(ctx, registry, logger)
	if err != nil {
		logger.Error("Failed to create user adapter", zap.Error(err))
		return err
	}

	err = registry.RegisterService(ctx, serverConfig.ProductSvcName+"-grpc", GRPCserviceID, serverConfig.ProductGRPCAddr, grpcPortInt, []string{"grpc"})
	if err != nil {
		logger.Error("Failed to register user service to consul", zap.Error(err))
	}

	go func() {
		<-ctx.Done()
		logger.Info("Context canceled. Deregistering services...")
		registry.DeregisterService(context.Background(), GRPCserviceID)

		logger.Info("Shutting down servers...")
		if err := app.Shutdown(); err != nil {
		}
		if grpcServer != nil {
			grpcServer.GracefulStop()
		}
		logger.Info("Successfully shutdown...")
	}()

	go consul.StartHealthCheckLoop(ctx, registry, GRPCserviceID, serverConfig.ProductSvcName+"-grpc", logger)

	productRepo := repository.NewProductRepository()
	productTransactionRepo := repository.NewProductTransactionRepository()

	productUC := usecase.NewProductUseCase(productRepo, databaseStore, customValidator, logger)
	productTransactionUC := usecase.NewProductTransactionUseCase(productRepo, productTransactionRepo, databaseStore, customValidator, logger)

	productController := controller.NewProductController(productUC, logger)

	go func() {
		grpcServer = grpc.NewServer()
		reflection.Register(grpcServer)
		l, err := net.Listen("tcp", fmt.Sprintf("%s:%s", serverConfig.ProductGRPCAddr, serverConfig.ProductGRPCPort))
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to listen: %v", err))
			return
		}

		defer l.Close()

		grpcHandler.NewProductHandler(grpcServer, productTransactionUC)

		if err := grpcServer.Serve(l); err != nil {
			logger.Error(fmt.Sprintf("Failed to start gRPC server: %v", err))
		}
	}()

	userMiddleware := middleware.NewUserAuth(userAdapter, logger)

	userRoute := route.NewProductRoute(app, productController, userMiddleware)
	userRoute.RegisterRoutes()

	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- app.Listen(fmt.Sprintf("%s:%s", serverConfig.ProductHTTPAddr, serverConfig.ProductHTTPPort))
	}()

	transactionConsumer := consumer.NewTransactionConsumer(productTransactionUC, jetStreamConfig, logger)
	if err := transactionConsumer.ConsumeAllEvents(ctx); err != nil {
		logger.Error("Failed to consume transaction events", zap.Error(err))
	}

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
