package adapter

import (
	"context"
	"go-saga-pattern/commoner/discovery"
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/commoner/utils"
	"go-saga-pattern/proto/userpb"

	"go.uber.org/zap"
)

type UserAdapter interface {
	AuthenticateUser(ctx context.Context, token string) (*userpb.AuthenticateResponse, error)
}

type userAdapter struct {
	client userpb.UserServiceClient
}

func NewUserAdapter(ctx context.Context, registry discovery.Registry, logs logs.Log) (UserAdapter, error) {
	userServiceName := utils.GetEnv("USER_SVC_NAME") + "-grpc"
	logs.Info("Connecting to user service", zap.String("service_name", userServiceName))
	conn, err := discovery.ServiceConnection(ctx, userServiceName, registry, logs)
	if err != nil {
		return nil, err
	}

	logs.Info("successfuly connected", zap.String("service_name", userServiceName))
	client := userpb.NewUserServiceClient(conn)

	return &userAdapter{
		client: client,
	}, nil
}

func (a *userAdapter) AuthenticateUser(ctx context.Context, token string) (*userpb.AuthenticateResponse, error) {
	processPhotoRequest := &userpb.AuthenticateRequest{
		Token: token,
	}

	response, err := a.client.AuthenticateUser(ctx, processPhotoRequest)
	if err != nil {
		return nil, helper.FromGRPCError(err)
	}

	return response, nil
}
