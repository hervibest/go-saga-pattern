package controller

import (
	"go-saga-pattern/product-svc/internal/model"
	"go-saga-pattern/product-svc/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type ProductController interface {
	OwnerCreate(ctx *fiber.Ctx) error
	OwnerDelete(ctx *fiber.Ctx) error
	OwnerSearch(ctx *fiber.Ctx) error
	OwnerUpdate(ctx *fiber.Ctx) error
	PublicSearch(ctx *fiber.Ctx) error
	GetById(ctx *fiber.Ctx) error
	GetBySlug(ctx *fiber.Ctx) error
}
type productController struct {
	productUseCase usecase.ProductUseCase
}

func NewProductController() ProductController {
	return &productController{}
}

func (c *productController) OwnerCreate(ctx *fiber.Ctx) error {
	request := new(model.CreateProductRequest)
	if err := ctx.BodyParser(request); err != nil {

	}
}

func (c *productController) OwnerDelete(ctx *fiber.Ctx) error {

}

func (c *productController) OwnerSearch(ctx *fiber.Ctx) error {

}

func (c *productController) OwnerUpdate(ctx *fiber.Ctx) error {

}

func (c *productController) PublicSearch(ctx *fiber.Ctx) error {

}

func (c *productController) GetById(ctx *fiber.Ctx) error {

}

func (c *productController) GetBySlug(ctx *fiber.Ctx) error {

}
