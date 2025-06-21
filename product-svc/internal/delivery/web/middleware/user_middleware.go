package middleware

import (
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/product-svc/internal/adapter"
	"go-saga-pattern/product-svc/internal/model"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// TODO SHOULD TOKEN VALIDATED ?
func NewUserAuth(userAdapter adapter.UserAdapter, logs logs.Log) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		token := strings.TrimPrefix(ctx.Get("Authorization", ""), "Bearer ")
		if token == "" || token == "NOT_FOUND" {
			return fiber.NewError(fiber.ErrUnauthorized.Code, "Unauthorized access")
		}

		authResponse, err := userAdapter.AuthenticateUser(ctx.UserContext(), token)
		if err != nil {
			return helper.ErrUseCaseResponseJSON(ctx, "Authenticate user : ", err, logs)
		}

		auth := &model.AuthResponse{
			ID:       authResponse.GetUser().GetId(),
			Username: authResponse.GetUser().GetUsername(),
			Email:    authResponse.GetUser().GetEmail()}

		ctx.Locals("user", auth)
		return ctx.Next()
	}
}

func GetUser(ctx *fiber.Ctx) *model.AuthResponse {
	return ctx.Locals("user").(*model.AuthResponse)
}
