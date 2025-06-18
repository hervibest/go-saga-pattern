package controller

import (
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/commoner/web"
	"go-saga-pattern/product-svc/internal/model"
	"go-saga-pattern/product-svc/internal/usecase"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
	logs           logs.Log
}

func NewProductController(productUseCase usecase.ProductUseCase, logs logs.Log) ProductController {
	return &productController{productUseCase: productUseCase, logs: logs}
}

// TODO implement user id
func (c *productController) OwnerCreate(ctx *fiber.Ctx) error {
	request := new(model.CreateProductRequest)

	if err := ctx.BodyParser(request); err != nil {
		return helper.ErrBodyParserResponseJSON(ctx, err)
	}
	response, err := c.productUseCase.OwnerCreate(ctx.UserContext(), request)
	if err != nil {
		if validatonErrs, ok := err.(*helper.UseCaseValError); ok {
			return helper.ErrValidationResponseJSON(ctx, validatonErrs)
		}
		return helper.ErrUseCaseResponseJSON(ctx, "Create product error : ", err, c.logs)
	}

	return ctx.Status(http.StatusCreated).JSON(web.WebResponse[*model.ProductResponse]{
		Success: true,
		Data:    response,
	})
}

// TODO implement user id
func (c *productController) OwnerDelete(ctx *fiber.Ctx) error {
	productId := ctx.Params("id")
	if productId == "" {
		return helper.ErrCustomResponseJSON(ctx, http.StatusBadRequest, "Product ID is required")
	}

	parsedId, err := uuid.Parse(productId)
	if err != nil {
		return helper.ErrCustomResponseJSON(ctx, http.StatusBadRequest, "Invalid Product ID format")
	}

	request := new(model.DeleteProductRequest)
	request.Id = parsedId

	err = c.productUseCase.OwnerDelete(ctx.UserContext(), request)
	if err != nil {
		if validatonErrs, ok := err.(*helper.UseCaseValError); ok {
			return helper.ErrValidationResponseJSON(ctx, validatonErrs)
		}
		return helper.ErrUseCaseResponseJSON(ctx, "Delete product error : ", err, c.logs)
	}

	return ctx.Status(http.StatusOK).JSON(web.WebResponse[any]{
		Success: true,
	})
}

// TODO implement user id
func (c *productController) OwnerSearch(ctx *fiber.Ctx) error {

}

// TODO implement user id
func (c *productController) OwnerUpdate(ctx *fiber.Ctx) error {

}

func (c *productController) PublicSearch(ctx *fiber.Ctx) error {

}

// TODO implement user id
func (c *productController) GetById(ctx *fiber.Ctx) error {

}

func (c *productController) GetBySlug(ctx *fiber.Ctx) error {

}
