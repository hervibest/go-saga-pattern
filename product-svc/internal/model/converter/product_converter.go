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
	responses := make([]*model.ProductResponse, 0, len(products))
	for _, product := range products {
		responses = append(responses, ProductToResponse(product))
	}
	return responses
}

func ProductsWithTotalToResponses(productsWithTotal []*entity.ProductWithTotal) []*model.ProductResponse {
	responses := make([]*model.ProductResponse, 0, len(productsWithTotal))
	for _, productWithTotal := range productsWithTotal {
		product := &entity.Product{
			ID:          productWithTotal.ID,
			UserID:      productWithTotal.UserID,
			Name:        productWithTotal.Name,
			Slug:        productWithTotal.Slug,
			Description: productWithTotal.Description,
			Price:       productWithTotal.Price,
			Quantity:    productWithTotal.Quantity,
		}

		responses = append(responses, ProductToResponse(product))
	}
	return responses
}

func ProductsToCheckQuantityResponse(transactionID uuid.UUID, products []*entity.Product) *model.CheckProductsQuantityRequestResponse {
	responses := make([]*model.ProductResponse, 0, len(products))
	for _, product := range products {
		responses = append(responses, &model.ProductResponse{
			ID:       product.ID.String(),
			Quantity: product.Quantity,
			Price:    product.Price},
		)
	}
	return &model.CheckProductsQuantityRequestResponse{TransactionID: transactionID, Products: responses}
}
