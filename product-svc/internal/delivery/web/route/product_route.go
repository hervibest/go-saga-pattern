package route

import (
	"go-saga-pattern/product-svc/internal/delivery/web/controller"

	"github.com/gofiber/fiber/v2"
)

type ProductRoute struct {
	app               *fiber.App
	productController controller.ProductController
	userMiddleware    fiber.Handler
}

func NewProductRoute(app *fiber.App, productController controller.ProductController, userMiddleware fiber.Handler) *ProductRoute {
	return &ProductRoute{
		app:               app,
		productController: productController,
		userMiddleware:    userMiddleware,
	}
}

func (r *ProductRoute) RegisterRoutes() {
	userRoutes := r.app.Group("/api/v1/products", r.userMiddleware)
	userRoutes.Post("/", r.productController.OwnerCreate)
	userRoutes.Get("/", r.productController.OwnerSearch)
	userRoutes.Put("/:id", r.productController.OwnerUpdate)
	userRoutes.Delete("/delete/:id", r.productController.OwnerDelete)
}
