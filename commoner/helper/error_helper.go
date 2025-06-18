package helper

import (
	"fmt"
	errorcode "go-saga-pattern/commoner/constant/errcode"
	"go-saga-pattern/commoner/logs"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AppError struct {
	Code        string
	Message     string
	GRPCCode    codes.Code
	Err         error
	InternalErr error
}

func NewUseCaseWithInternalError(code, message string, errs ...error) *AppError {
	return &AppError{Code: code, Message: message, InternalErr: errs[0]}
}

func NewUseCaseError(code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func NewAppError(code string, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func NewAppGRPCError(code string, grpcCode codes.Code, message string) *AppError {
	return &AppError{
		Code:     code,
		GRPCCode: grpcCode,
		Message:  message,
	}
}

func NewAppGRPCInternalError(code string, grpcCode codes.Code, message string, err error) *AppError {
	return &AppError{
		Code:     code,
		GRPCCode: grpcCode,
		Message:  message,
		Err:      err,
	}
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code string, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func (e *AppError) HTTPStatus() int {
	switch e.Code {
	case errorcode.ErrUnauthorized, errorcode.ErrUserSignedOut:
		return 401
	case errorcode.ErrForbidden: // ✅ Tambahan di sini
		return 403
	case errorcode.ErrValidationFailed, errorcode.ErrInvalidArgument:
		return 422
	case errorcode.ErrAlreadyExists:
		return 409
	case errorcode.ErrUserNotFound, errorcode.ErrResourceNotFound:
		return 404
	case errorcode.ErrTooManyRequests:
		return 429
	case errorcode.ErrExternal:
		return 503
	default:
		return 500
	}
}

func (e *AppError) GRPCErrorCode() error {
	switch e.Code {
	case errorcode.ErrUnauthorized, errorcode.ErrUserSignedOut:
		return status.Error(codes.Unauthenticated, e.Message)
	case errorcode.ErrValidationFailed, errorcode.ErrInvalidArgument:
		return status.Error(codes.InvalidArgument, e.Message)
	case errorcode.ErrAlreadyExists:
		return status.Error(codes.AlreadyExists, e.Message)
	case errorcode.ErrUserNotFound, errorcode.ErrResourceNotFound:
		return status.Error(codes.NotFound, e.Message)
	case errorcode.ErrTooManyRequests:
		return status.Error(codes.ResourceExhausted, e.Message)
	default:
		return status.Error(codes.Internal, "Internal server error")
	}
}

func WrapInternalServerError(logs logs.Log, internalMsg string, err error) error {
	logs.Error(internalMsg, zap.Error(err))
	return NewAppError(errorcode.ErrInternal, "Something went wrong. Please try again later", err)
}

func WrapExternalServiceUnavailable(logs logs.Log, internalMsg string, err error) error {
	logs.Error(internalMsg, zap.Error(err))
	return NewAppError(errorcode.ErrExternal, "Service unavailable. Please try again later.", err)
}

func FromGRPCError(err error) *AppError {
	st, ok := status.FromError(err)
	if !ok {
		return NewAppGRPCInternalError(errorcode.ErrInternal, codes.Unknown, "non-gRPC error", err)
	}

	switch st.Code() {
	case codes.Unauthenticated:
		return NewAppGRPCError(errorcode.ErrUnauthorized, codes.Unauthenticated, st.Message())
	case codes.InvalidArgument:
		return NewAppGRPCError(errorcode.ErrInvalidArgument, codes.InvalidArgument, st.Message())
	case codes.NotFound:
		return NewAppGRPCError(errorcode.ErrUserNotFound, codes.NotFound, st.Message())
	case codes.ResourceExhausted:
		return NewAppGRPCError(errorcode.ErrTooManyRequests, codes.ResourceExhausted, st.Message())
	default:
		return NewAppGRPCInternalError(errorcode.ErrInternal, st.Code(), st.Message(), err)
	}
}
