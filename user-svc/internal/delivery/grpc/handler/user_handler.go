package grpc

import (
	"context"
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/proto/userpb"
	"go-saga-pattern/user-svc/internal/usecase"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type UserGRPCHandler struct {
	userUC usecase.UserUseCase
	userpb.UnimplementedUserServiceServer
}

func NewUserGRPCHandler(server *grpc.Server, userUC usecase.UserUseCase) {
	handler := &UserGRPCHandler{
		userUC: userUC,
	}
	userpb.RegisterUserServiceServer(server, handler)
}

func (h *UserGRPCHandler) AuthenticateUser(ctx context.Context, req *userpb.AuthenticateRequest) (*userpb.AuthenticateResponse, error) {
	response, err := h.userUC.VerifyUser(ctx, req.GetToken())
	if err != nil {
		appErr, ok := err.(*helper.AppError)
		if ok {
			return nil, appErr.GRPCErrorCode()
		}
	}

	user := &userpb.User{
		Id:       response.ID,
		Username: response.Username,
		Email:    response.Email,
	}

	return &userpb.AuthenticateResponse{
		Status: int64(codes.OK),
		User:   user}, nil
}
