package controller

import (
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/user-svc/internal/delivery/http/middleware"
	"go-saga-pattern/user-svc/internal/model"
	"go-saga-pattern/user-svc/internal/usecase"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserControler interface {
	CurrentUser(ctx *fiber.Ctx) error
	LoginUser(ctx *fiber.Ctx) error
	RegisterUser(ctx *fiber.Ctx) error
	UserLogout(ctx *fiber.Ctx) error
}
type userControlerImpl struct {
	userUC usecase.UserUseCase
	logs   logs.Log
}

func NewUserController(userUC usecase.UserUseCase, logs logs.Log) UserControler {
	return &userControlerImpl{
		userUC: userUC,
		logs:   logs,
	}
}

func (c *userControlerImpl) LoginUser(ctx *fiber.Ctx) error {
	request := new(model.LoginUserRequest)
	if err := ctx.BodyParser(request); err != nil {
		return helper.ErrBodyParserResponseJSON(ctx, err)
	}

	user, token, err := c.userUC.LoginUser(ctx.UserContext(), request)
	if err != nil {

		return helper.ErrUseCaseResponseJSON(ctx, "Login user error : ", err, c.logs)
	}

	response := map[string]interface{}{
		"user":  user,
		"token": token,
	}

	return ctx.Status(http.StatusOK).JSON(model.WebResponse[any]{
		Success: true,
		Data:    response,
	})
}

// TODO STILL RANDOM UUID CREATEDBY
func (c *userControlerImpl) RegisterUser(ctx *fiber.Ctx) error {
	request := new(model.RegisterUserRequest)
	if err := ctx.BodyParser(request); err != nil {
		return helper.ErrBodyParserResponseJSON(ctx, err)
	}

	if err := helper.SetBaseModel(ctx, request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid IP address")
	}

	request.CreatedBy = uuid.Must(uuid.NewRandom())
	request.UpdatedBy = request.CreatedBy

	user, err := c.userUC.RegisterUser(ctx.UserContext(), request)
	if err != nil {

		return helper.ErrUseCaseResponseJSON(ctx, "Register user error : ", err, c.logs)
	}

	response := map[string]interface{}{
		"user": user,
	}

	return ctx.Status(http.StatusOK).JSON(model.WebResponse[any]{
		Success: true,
		Data:    response,
	})
}

func (c *userControlerImpl) CurrentUser(ctx *fiber.Ctx) error {
	user := middleware.GetUser(ctx)

	userResponse, err := c.userUC.CurrentUser(ctx.Context(), user.Email)
	if err != nil {
		return helper.ErrUseCaseResponseJSON(ctx, "Current error : ", err, c.logs)
	}

	return ctx.Status(http.StatusOK).JSON(model.WebResponse[*model.UserResponse]{
		Success: true,
		Data:    userResponse,
	})
}

func (c *userControlerImpl) UserLogout(ctx *fiber.Ctx) error {
	user := middleware.GetUser(ctx)

	request := new(model.LogoutUserRequest)
	request.UserId = user.ID
	request.AccessToken = user.Token
	request.ExpiresAt = user.ExpiresAt

	if err := c.userUC.Logout(ctx.Context(), request); err != nil {
		return helper.ErrUseCaseResponseJSON(ctx, "Logout error : ", err, c.logs)
	}

	return ctx.Status(http.StatusOK).JSON(model.WebResponse[any]{
		Success: true,
	})
}
