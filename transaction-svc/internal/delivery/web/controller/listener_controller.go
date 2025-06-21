package controller

import (
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/commoner/web"
	"go-saga-pattern/transaction-svc/internal/model"
	"go-saga-pattern/transaction-svc/internal/usecase"

	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ListenerController interface {
	NotifyTransaction(ctx *fiber.Ctx) error
}

type listenerController struct {
	listernerUseCase usecase.ListenerUseCase
	logs             logs.Log
}

func NewListenerController(listernerUseCase usecase.ListenerUseCase,
	logs logs.Log) ListenerController {
	return &listenerController{listernerUseCase: listernerUseCase, logs: logs}
}

func (c *listenerController) NotifyTransaction(ctx *fiber.Ctx) error {
	webhookRequest := new(model.WebhookNotifyRequest)
	if err := ctx.BodyParser(webhookRequest); err != nil {
		return helper.ErrBodyParserResponseJSON(ctx, err)
	}
	if _, err := uuid.Parse(webhookRequest.OrderID); err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid order id")
	}

	webhookRequest.Body = ctx.Body()

	c.logs.Info("Notify transaction", zap.String("order_id", webhookRequest.OrderID), zap.String("extermal status", webhookRequest.MidtransTransactionStatus))
	if err := c.listernerUseCase.ConsumeAndProduceWebhook(ctx.Context(), webhookRequest); err != nil {
		return helper.ErrUseCaseResponseJSON(ctx, "Notify webhook : ", err, c.logs)
	}

	return ctx.Status(http.StatusOK).JSON(web.WebResponse[any]{
		Success: true,
	})

}
