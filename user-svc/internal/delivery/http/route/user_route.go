package route

import (
	"go-saga-pattern/user-svc/internal/delivery/http/controller"

	"github.com/gofiber/fiber/v2"
)

type UserRoute struct {
	app            *fiber.App
	userHandler    controller.UserControler
	userMiddleware fiber.Handler
}

func NewUserRoute(app *fiber.App, userHandler controller.UserControler, userMiddleware fiber.Handler) *UserRoute {
	return &UserRoute{
		app:            app,
		userHandler:    userHandler,
		userMiddleware: userMiddleware,
	}
}
func (r *UserRoute) RegisterRoutes() {
	r.app.Post("/api/v1/user/login", r.userHandler.LoginUser)
	r.app.Post("/api/v1/user/register", r.userHandler.RegisterUser)

	userRoutes := r.app.Group("/api/v1/users", r.userMiddleware)
	userRoutes.Get("/current", r.userHandler.CurrentUser)
	userRoutes.Post("/logout", r.userHandler.UserLogout)
}
