package route

import (
	"go-saga-pattern/transaction-svc/internal/delivery/web/controller"

	"github.com/gofiber/fiber/v2"
)

type ListenerRoute struct {
	app                *fiber.App
	listenerController controller.ListenerController
}

func NewListenerRoute(app *fiber.App, listenerController controller.ListenerController) *ListenerRoute {
	return &ListenerRoute{
		app:                app,
		listenerController: listenerController,
	}
}

func (r *ListenerRoute) RegisterRoutes() {
	listenerRoute := r.app.Group("/api/v1/transaction")
	listenerRoute.Post("/webhook/notify", r.listenerController.NotifyTransaction)
}
