package consul

import (
	"context"
	"fmt"
	"go-saga-pattern/commoner/logs"
	"time"

	consul "github.com/hashicorp/consul/api"
)

type Registry struct {
	client *consul.Client
}

func NewRegistry(address, serviceName string) (*Registry, error) {
	config := consul.DefaultConfig()
	config.Address = address

	client, err := consul.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	return &Registry{
		client: client,
	}, nil
}

func (r *Registry) RegisterService(ctx context.Context, serviceName, serviceID, serviceAddress string, servicePort int, tags []string) error {
	checkID := "service:" + serviceID

	registration := &consul.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Address: serviceAddress,
		Port:    servicePort,
		Tags:    tags,
		Check: &consul.AgentServiceCheck{
			TTL:                            "10s", // Menentukan waktu interval untuk TTL check
			Status:                         consul.HealthPassing,
			CheckID:                        checkID, // Pastikan check ID sesuai dengan yang Anda gunakan di UpdateTTL
			DeregisterCriticalServiceAfter: "1m",    // Deregistrasi setelah gagal dalam 1 menit
		},
	}

	err := r.client.Agent().ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	return nil
}

func (r *Registry) DeregisterService(ctx context.Context, serviceID string) error {
	err := r.client.Agent().ServiceDeregister(serviceID)
	if err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	return nil
}

func (r *Registry) GetService(ctx context.Context, serviceName string) ([]*consul.ServiceEntry, error) {
	services, _, err := r.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	return services, nil
}

func (r *Registry) HealthCheck(serviceID, serviceName string) error {

	checkID := "service:" + serviceID

	return r.client.Agent().UpdateTTL(checkID, "online", consul.HealthPassing)
}

func StartHealthCheckLoop(ctx context.Context, registry *Registry, serviceID, serviceName string, logs logs.Log) {
	failureCount := 0
	const maxFailures = 5
	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := registry.HealthCheck(serviceID, serviceName)
			if err != nil {
				logs.Error(fmt.Sprintf("Failed to perform health check for %s: %v", serviceName, err))
				failureCount++
				if failureCount >= maxFailures {
					logs.Error(fmt.Sprintf("Max health check failures reached for %s. Exiting loop.", serviceName))
					return
				}
			} else {
				failureCount = 0
			}
			time.Sleep(2 * time.Second)
		}
	}
}
