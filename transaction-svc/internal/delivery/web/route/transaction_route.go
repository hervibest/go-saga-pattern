package route

import (
	"go-saga-pattern/transaction-svc/internal/delivery/web/controller"

	"github.com/gofiber/fiber/v2"
)

type TransactionRoute struct {
	app                   *fiber.App
	transactionController controller.TransactionController
	userMiddleware        fiber.Handler
}

func NewTransactionRoute(app *fiber.App, transactionController controller.TransactionController, userMiddleware fiber.Handler) *TransactionRoute {
	return &TransactionRoute{
		app:                   app,
		transactionController: transactionController,
		userMiddleware:        userMiddleware,
	}
}

func (r *TransactionRoute) RegisterRoutes() {
	userRoutes := r.app.Group("/api/v1/transaction", r.userMiddleware)
	userRoutes.Post("/buy", r.transactionController.CreateTransaction)
	userRoutes.Get("/", r.transactionController.UserSearch)
	userRoutes.Get("/detail", r.transactionController.UserSearchWithDetail)
	userRoutes.Get("/owner/detail", r.transactionController.OwnerSearchWithDetail)
}
