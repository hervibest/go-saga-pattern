package helper

import (
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ErrGRPC(err error) error {
	if appErr, ok := err.(*AppError); ok {
		if appErr.Err != nil {
			log.Printf("[gRPC Auth Error] %s: %v", appErr.Code, appErr.Err)
		}
		return appErr.GRPCErrorCode()
	}

	log.Printf("[gRPC Unhandled Error] %v", err)
	return status.Error(codes.Internal, "Unexpected error")
}
