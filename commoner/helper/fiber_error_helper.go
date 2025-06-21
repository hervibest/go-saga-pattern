package helper

import (
	"fmt"
	errorcode "go-saga-pattern/commoner/constant/errcode"
	"go-saga-pattern/commoner/logs"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type BodyParseErrorResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

type ValidationErrorResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ValidationErrorField struct {
	Field   string `json:"field"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

func ErrCustomResponseJSON(ctx *fiber.Ctx, status int, message string) error {
	return ctx.Status(status).JSON(ErrorResponse{
		Success: false,
		Message: message,
	})
}

func ErrBodyParserResponseJSON(ctx *fiber.Ctx, err error) error {
	return ctx.Status(http.StatusBadRequest).JSON(BodyParseErrorResponse{
		Success: false,
		Message: "Invalid request. Please check the submitted data.",
	})
}

func ErrValidationResponseJSON(ctx *fiber.Ctx, validatonErrs *UseCaseValError) error {
	return ctx.Status(http.StatusUnprocessableEntity).JSON(ValidationErrorResponse{
		Success: false,
		Message: "Validation error",
		Errors:  validatonErrs.GetValidationErrors(),
	})
}

func ErrUseCaseResponseJSON(ctx *fiber.Ctx, msg string, err error, logs logs.Log) error {
	if validatonErrs, ok := err.(*UseCaseValError); ok {
		return ErrValidationResponseJSON(ctx, validatonErrs)
	}

	if appErr, ok := err.(*AppError); ok {
		logs.Info(fmt.Sprintf("UseCase error in controller : %s [%s]: %v", msg, appErr.Code, appErr.Err))
		if appErr.Err != nil && appErr.Err.Error() == "" && appErr.Code == errorcode.ErrInternal {
			logs.Error(fmt.Sprintf("Internal error in controller : %s [%s]: %v", msg, appErr.Code, appErr.Err.Error()))
		} else {
			logs.Info(fmt.Sprintf("Client error in controller : %s [%s]: %v", msg, appErr.Code, appErr.Message))
		}

		return ctx.Status(appErr.HTTPStatus()).JSON(ErrorResponse{
			Success: false,
			Message: appErr.Message,
		})
	}

	return fiber.NewError(fiber.StatusInternalServerError, "Something went wrong. Please try again later")
}
