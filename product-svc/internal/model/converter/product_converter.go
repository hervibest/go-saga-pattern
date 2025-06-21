package converter

import (
	"go-saga-pattern/commoner/helper/nullable"
	"go-saga-pattern/product-svc/internal/entity"
	"go-saga-pattern/product-svc/internal/model"

	"github.com/google/uuid"
)

func ProductToResponse(product *entity.Product) *model.ProductResponse {
	return &model.ProductResponse{
		ID:          product.ID.String(),
		Name:        product.Name,
		Description: nullable.SQLtoString(product.Description),
		Price:       product.Price,
		Quantity:    product.Quantity,
	}
}

func ProductsToResponses(products []*entity.Product) []*model.ProductResponse {
	responses := make([]*model.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = ProductToResponse(product)
	}
	return responses
}

func ProductsWithTotalToResponses(productsWithTotal []*entity.ProductWithTotal) []*model.ProductResponse {
	responses := make([]*model.ProductResponse, len(productsWithTotal))
	for i, productWithTotal := range productsWithTotal {
		product := &entity.Product{
			ID:          productWithTotal.ID,
			UserID:      productWithTotal.UserID,
			Name:        productWithTotal.Name,
			Slug:        productWithTotal.Slug,
			Description: productWithTotal.Description,
			Price:       productWithTotal.Price,
			Quantity:    productWithTotal.Quantity,
		}

		responses[i] = ProductToResponse(product)
	}
	return responses
}

func ProductsToCheckQuantityResponse(transactionID uuid.UUID, products []*entity.Product) *model.CheckProductsQuantityRequestResponse {
	responses := make([]*model.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = &model.ProductResponse{
			ID:       product.ID.String(),
			Quantity: product.Quantity,
			Price:    product.Price,
		}
	}
	return &model.CheckProductsQuantityRequestResponse{TransactionID: transactionID, Products: responses}
}
