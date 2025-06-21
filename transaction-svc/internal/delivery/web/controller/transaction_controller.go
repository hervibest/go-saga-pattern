package controller

import (
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/commoner/web"
	"go-saga-pattern/transaction-svc/internal/delivery/web/middleware"
	"go-saga-pattern/transaction-svc/internal/model"
	"go-saga-pattern/transaction-svc/internal/usecase"

	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TransactionController interface {
	CreateTransaction(ctx *fiber.Ctx) error
}

type transactionController struct {
	transactionUseCase usecase.TransactionUseCase
	logs               logs.Log
}

func NewTransactionController(transactionUseCase usecase.TransactionUseCase, logs logs.Log) TransactionController {
	return &transactionController{transactionUseCase: transactionUseCase, logs: logs}
}

func (c *transactionController) CreateTransaction(ctx *fiber.Ctx) error {
	request := new(model.CreateTransactionRequest)
	if err := ctx.BodyParser(request); err != nil {
		return helper.ErrBodyParserResponseJSON(ctx, err)
	}

	user := middleware.GetUser(ctx)
	request.UserID = uuid.MustParse(user.ID)
	response, err := c.transactionUseCase.CreateTransaction(ctx.UserContext(), request)
	if err != nil {
		return helper.ErrUseCaseResponseJSON(ctx, "Create transaction error : ", err, c.logs)
	}

	return ctx.Status(http.StatusCreated).JSON(web.WebResponse[*model.CreateTransactionResponse]{
		Success: true,
		Data:    response,
	})
}
