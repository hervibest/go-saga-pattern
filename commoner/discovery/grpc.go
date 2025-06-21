package discovery

import (
	"context"
	"fmt"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/commoner/utils"
	"math/rand"
	"strconv"
	"time"

	consul "github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Registry interface {
	RegisterService(ctx context.Context, serviceName, serviceID, serviceAddress string, servicePort int, tags []string) error
	DeregisterService(ctx context.Context, serviceID string) error
	GetService(ctx context.Context, serviceName string) ([]*consul.ServiceEntry, error)
	HealthCheck(serviceID, serviceName string) error
}

func GenerateServiceID(serviceName string) string {
	return fmt.Sprintf("%s-%d", serviceName, rand.New(rand.NewSource(time.Now().UnixNano())).Int())
}

func ServiceConnection(ctx context.Context, serviceName string, registry Registry, logs logs.Log) (*grpc.ClientConn, error) {
	logs.Info(fmt.Sprintf("attempting to connect to service: %s", serviceName))
	retryTime, _ := strconv.Atoi(utils.GetEnv("SERVICE_DISCOVERY_RETRY_TIME"))
	maxRetries, _ := strconv.Atoi(utils.GetEnv("SERVICE_DISCOVERY_MAX_RETRIES"))
	retryDelay := time.Duration(retryTime) * time.Second

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("service connection cancelled by context: %w", ctx.Err())
		default:
			logs.Error(fmt.Sprintf("trying to connect service: %s with attempt: %d and max retries: %d", serviceName, attempt, maxRetries))
		}

		service, err := registry.GetService(ctx, serviceName)
		if err != nil {
			lastErr = fmt.Errorf("failed to get service: %w", err)
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("service connection cancelled by context: %w", ctx.Err())
			case <-time.After(retryDelay):
				continue
			}
		}

		if len(service) == 0 {
			lastErr = fmt.Errorf("service %s not found", serviceName)
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("service connection cancelled by context: %w", ctx.Err())
			case <-time.After(retryDelay):
				continue
			}
		}

		serviceEntry := service[rand.Intn(len(service))]

		conn, err := grpc.NewClient(
			fmt.Sprintf("%s:%d", serviceEntry.Service.Address, serviceEntry.Service.Port),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			lastErr = fmt.Errorf("failed to connect: %w", err)
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("service connection cancelled by context: %w", ctx.Err())
			case <-time.After(retryDelay):
				continue
			}
		}

		return conn, nil
	}

	return nil, fmt.Errorf("service connection failed after %d retries: %w", maxRetries, lastErr)
}
