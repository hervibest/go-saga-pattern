package main

import (
	"context"
	"fmt"
	"strconv"

	"go-saga-pattern/commoner/discovery"
	"go-saga-pattern/commoner/discovery/consul"
	"go-saga-pattern/commoner/logs"

	"go-saga-pattern/transaction-svc/internal/adapter"
	"go-saga-pattern/transaction-svc/internal/config"
	"go-saga-pattern/transaction-svc/internal/delivery/web/controller"
	"go-saga-pattern/transaction-svc/internal/delivery/web/route"
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

	messagingAdapter := adapter.NewMessagingAdapter(js)

	registry, err := consul.NewRegistry(serverConfig.ConsulAddr, serverConfig.ListenerSvcName)
	if err != nil {
		logger.Error("Failed to create consul registry for service" + err.Error())
	}

	HTTPServiceID := discovery.GenerateServiceID(serverConfig.ListenerSvcName + "-http")
	httpPortInt, _ := strconv.Atoi(serverConfig.ListenerHTTPPort)

	err = registry.RegisterService(ctx, serverConfig.ListenerSvcName+"-http", HTTPServiceID, serverConfig.ListenerHTTPAddr, httpPortInt, []string{"http"})
	if err != nil {
		logger.Error("Failed to register listener service to consul", zap.Error(err))
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

	go consul.StartHealthCheckLoop(ctx, registry, HTTPServiceID, serverConfig.ListenerSvcName+"-http", logger)

	listenerUC := usecase.NewListenerUseCase(messagingAdapter, logger)
	listenerController := controller.NewListenerController(listenerUC, logger)

	productRoute := route.NewListenerRoute(app, listenerController)
	productRoute.RegisterRoutes()

	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- app.Listen(fmt.Sprintf("%s:%s", serverConfig.ListenerHTTPAddr, serverConfig.ListenerHTTPPort))
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
