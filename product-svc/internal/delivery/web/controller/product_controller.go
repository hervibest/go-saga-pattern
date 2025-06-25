package controller

import (
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/commoner/web"
	"go-saga-pattern/product-svc/internal/delivery/web/middleware"
	"go-saga-pattern/product-svc/internal/model"
	"go-saga-pattern/product-svc/internal/usecase"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ProductController interface {
	OwnerCreate(ctx *fiber.Ctx) error
	OwnerDelete(ctx *fiber.Ctx) error
	OwnerSearch(ctx *fiber.Ctx) error
	OwnerUpdate(ctx *fiber.Ctx) error
	PublicSearch(ctx *fiber.Ctx) error
	GetByID(ctx *fiber.Ctx) error
	GetBySlug(ctx *fiber.Ctx) error
}
type productController struct {
	productUseCase usecase.ProductUseCase
	logs           logs.Log
}

func NewProductController(productUseCase usecase.ProductUseCase, logs logs.Log) ProductController {
	return &productController{productUseCase: productUseCase, logs: logs}
}

func (c *productController) OwnerCreate(ctx *fiber.Ctx) error {
	request := new(model.CreateProductRequest)
	if err := ctx.BodyParser(request); err != nil {
		return helper.ErrBodyParserResponseJSON(ctx, err)
	}

	user := middleware.GetUser(ctx)
	request.UserID = uuid.MustParse(user.ID)
	response, err := c.productUseCase.OwnerCreate(ctx.UserContext(), request)
	if err != nil {
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
	request.ID = parsedId

	user := middleware.GetUser(ctx)
	request.UserID = uuid.MustParse(user.ID)

	err = c.productUseCase.OwnerDelete(ctx.UserContext(), request)
	if err != nil {
		return helper.ErrUseCaseResponseJSON(ctx, "Delete product error : ", err, c.logs)
	}

	return ctx.Status(http.StatusOK).JSON(web.WebResponse[any]{
		Success: true,
	})
}

// TODO implement user id
func (c *productController) OwnerSearch(ctx *fiber.Ctx) error {
	request := new(model.OwnerSearchProductsRequest)
	request.Limit = ctx.QueryInt("limit", 10)
	request.Page = ctx.QueryInt("page", 1)

	user := middleware.GetUser(ctx)
	request.UserID = uuid.MustParse(user.ID)

	products, pageMetadata, err := c.productUseCase.OwnerSearch(ctx.UserContext(), request)
	if err != nil {
		return helper.ErrUseCaseResponseJSON(ctx, "Search products error : ", err, c.logs)
	}

	return ctx.Status(http.StatusOK).JSON(web.WebResponse[[]*model.ProductResponse]{
		Success:      true,
		Data:         products,
		PageMetadata: pageMetadata,
	})

}

// TODO implement user id
func (c *productController) OwnerUpdate(ctx *fiber.Ctx) error {
	productId := ctx.Params("id")
	if productId == "" {
		return helper.ErrCustomResponseJSON(ctx, http.StatusBadRequest, "Product ID is required")
	}

	parsedId, err := uuid.Parse(productId)
	if err != nil {
		return helper.ErrCustomResponseJSON(ctx, http.StatusBadRequest, "Invalid Product ID format")
	}

	request := new(model.UpdateProductRequest)
	request.ID = parsedId

	if err := ctx.BodyParser(request); err != nil {
		return helper.ErrBodyParserResponseJSON(ctx, err)
	}

	user := middleware.GetUser(ctx)
	request.UserID = uuid.MustParse(user.ID)
	c.logs.Info("Update product request", zap.Any("request", request))

	product, err := c.productUseCase.OwnerUpdate(ctx.UserContext(), request)
	if err != nil {
		return helper.ErrUseCaseResponseJSON(ctx, "Update product error : ", err, c.logs)
	}

	return ctx.Status(http.StatusOK).JSON(web.WebResponse[*model.ProductResponse]{
		Success: true,
		Data:    product,
	})
}

func (c *productController) PublicSearch(ctx *fiber.Ctx) error {
	request := new(model.PublicSearchProductsRequest)
	request.Limit = ctx.QueryInt("limit", 10)
	request.Page = ctx.QueryInt("page", 1)

	products, pageMetadata, err := c.productUseCase.PublicSearch(ctx.UserContext(), request)
	if err != nil {
		return helper.ErrUseCaseResponseJSON(ctx, "Search products error : ", err, c.logs)
	}

	return ctx.Status(http.StatusOK).JSON(web.WebResponse[[]*model.ProductResponse]{
		Success:      true,
		Data:         products,
		PageMetadata: pageMetadata,
	})
}

// TODO implement user id
func (c *productController) GetByID(ctx *fiber.Ctx) error {
	productId := ctx.Params("id")
	if productId == "" {
		return helper.ErrCustomResponseJSON(ctx, http.StatusBadRequest, "Product ID is required")
	}

	parsedId, err := uuid.Parse(productId)
	if err != nil {
		return helper.ErrCustomResponseJSON(ctx, http.StatusBadRequest, "Invalid Product ID format")
	}

	product, err := c.productUseCase.GetByID(ctx.UserContext(), parsedId)
	if err != nil {
		return helper.ErrUseCaseResponseJSON(ctx, "Get product by ID error : ", err, c.logs)
	}

	return ctx.Status(http.StatusOK).JSON(web.WebResponse[*model.ProductResponse]{
		Success: true,
		Data:    product,
	})
}

func (c *productController) GetBySlug(ctx *fiber.Ctx) error {
	slug := ctx.Params("slug")
	if slug == "" {
		return helper.ErrCustomResponseJSON(ctx, http.StatusBadRequest, "Product slug is required")
	}

	product, err := c.productUseCase.GetBySlug(ctx.UserContext(), slug)
	if err != nil {

		return helper.ErrUseCaseResponseJSON(ctx, "Get product by slug error : ", err, c.logs)
	}

	return ctx.Status(http.StatusOK).JSON(web.WebResponse[*model.ProductResponse]{
		Success: true,
		Data:    product,
	})
}
