package controller

import (
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/commoner/web"
	"go-saga-pattern/transaction-svc/internal/delivery/web/middleware"
	"go-saga-pattern/transaction-svc/internal/model"
	"go-saga-pattern/transaction-svc/internal/usecase/contract"

	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TransactionController interface {
	CreateTransaction(ctx *fiber.Ctx) error
	UserSearch(ctx *fiber.Ctx) error
	UserSearchWithDetail(ctx *fiber.Ctx) error
	OwnerSearchWithDetail(ctx *fiber.Ctx) error
}

type transactionController struct {
	transactionUseCase contract.TransactionUseCase
	logs               logs.Log
}

func NewTransactionController(transactionUseCase contract.TransactionUseCase, logs logs.Log) TransactionController {
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

func (c *transactionController) UserSearch(ctx *fiber.Ctx) error {
	request := new(model.UserSearchTransactionRequest)
	request.Page = ctx.QueryInt("page", 1)
	request.Limit = ctx.QueryInt("limit", 10)
	user := middleware.GetUser(ctx)
	request.UserID = uuid.MustParse(user.ID)
	response, pageMetadata, err := c.transactionUseCase.UserSearch(ctx.UserContext(), request)
	if err != nil {
		return helper.ErrUseCaseResponseJSON(ctx, "User search transaction error : ", err, c.logs)
	}

	baseURL := ctx.BaseURL() + ctx.Path()
	helper.GeneratePageURLs(baseURL, pageMetadata)

	return ctx.Status(http.StatusCreated).JSON(web.WebResponse[[]*model.TransactionResponse]{
		Success:      true,
		Data:         response,
		PageMetadata: pageMetadata,
	})
}

func (c *transactionController) UserSearchWithDetail(ctx *fiber.Ctx) error {
	request := new(model.UserSearchTransactionRequest)
	request.Page = ctx.QueryInt("page", 1)
	request.Limit = ctx.QueryInt("limit", 10)
	user := middleware.GetUser(ctx)
	request.UserID = uuid.MustParse(user.ID)
	response, pageMetadata, err := c.transactionUseCase.UserSearchWithDetail(ctx.UserContext(), request)
	if err != nil {
		return helper.ErrUseCaseResponseJSON(ctx, "User search transaction error : ", err, c.logs)
	}

	c.logs.Info("User search with detail accessed")

	baseURL := ctx.BaseURL() + ctx.Path()
	helper.GeneratePageURLs(baseURL, pageMetadata)

	return ctx.Status(http.StatusOK).JSON(web.WebResponse[[]*model.TransactionResponse]{
		Success:      true,
		Data:         response,
		PageMetadata: pageMetadata,
	})
}

func (c *transactionController) OwnerSearchWithDetail(ctx *fiber.Ctx) error {
	request := new(model.OwnerSearchTransactionRequest)
	parsedProductID, err := uuid.Parse(ctx.Query("product_id"))
	if err != nil {
		return helper.ErrCustomResponseJSON(ctx, http.StatusBadRequest, "Invalid product id")
	}

	request.ProductID = parsedProductID
	request.Page = ctx.QueryInt("page", 1)
	request.Limit = ctx.QueryInt("limit", 10)
	user := middleware.GetUser(ctx)
	request.UserID = uuid.MustParse(user.ID)
	response, pageMetadata, err := c.transactionUseCase.OwnerSearchWithDetail(ctx.UserContext(), request)
	if err != nil {
		return helper.ErrUseCaseResponseJSON(ctx, "User search transaction error : ", err, c.logs)
	}

	c.logs.Info("Owner search with detail accessed")

	baseURL := ctx.BaseURL() + ctx.Path()
	helper.GeneratePageURLs(baseURL, pageMetadata)

	return ctx.Status(http.StatusOK).JSON(web.WebResponse[[]*model.TransactionResponse]{
		Success:      true,
		Data:         response,
		PageMetadata: pageMetadata,
	})
}
