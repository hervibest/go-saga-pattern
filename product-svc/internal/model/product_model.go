package model

import "github.com/google/uuid"

type CreateProductRequest struct {
	UserID      uuid.UUID `json:"user_id" validate:"required,uuid"`
	Name        string    `json:"name" validate:"required"`
	Description *string   `json:"description"`
	Price       float64   `json:"price" validate:"required,gt=0"`
	Quantity    int       `json:"quantity" validate:"required,gt=0"`
}

type GetProductRequest struct {
	ID string `json:"id" validate:"required,uuid"`
}

type OwnerSearchProductsRequest struct {
	UserID uuid.UUID `validate:"required,uuid"`
	Page   int       `validate:"required,min=1"`
	Limit  int       `validate:"required,min=1,max=100"`
}

type OwnerGetProductRequest struct {
	UserID    uuid.UUID `validate:"required,uuid"`
	ProductID uuid.UUID `validate:"required,uuid"`
}

type PublicSearchProductsRequest struct {
	Page  int `json:"page" validate:"required,min=1"`
	Limit int `json:"limit" validate:"required,min=1,max=100"`
}

type UpdateProductRequest struct {
	ID          uuid.UUID `json:"-" validate:"required,uuid"`
	UserID      uuid.UUID `json:"user_id" validate:"required,uuid"`
	Name        string    `json:"name" validate:"required"`
	Description *string   `json:"description"`
	Price       float64   `json:"price" validate:"gt=0"`
	Quantity    int       `json:"quantity" validate:"gte=0"`
}

type DeleteProductRequest struct {
	ID     uuid.UUID `json:"id" validate:"required,uuid"`
	UserID uuid.UUID `json:"user_id" validate:"required,uuid"`
}

type ProductResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	CreatedAt   string  `json:"created_at,omitempty"`
	UpdatedAt   string  `json:"updated_at,omitempty"`
	DeletedAt   string  `json:"deleted_at,omitempty"`
}

type CheckProductQuantity struct {
	ProductID uuid.UUID
	Quantity  int
	Price     float64
}

type CheckProductsQuantityRequest struct {
	TransactionID uuid.UUID
	Products      []*CheckProductQuantity
}

type CancelProductTransactionsRequest struct {
	TransactionID uuid.UUID
}

type ExpireProductTransactionsRequest struct {
	TransactionID uuid.UUID
}

type CommitProductTransactionsRequest struct {
	TransactionID uuid.UUID
}

type SettleProductTransactionRequest struct {
	TransactionID uuid.UUID
}

type CheckProductsQuantityRequestResponse struct {
	TransactionID uuid.UUID          `json:"transaction_id"`
	Products      []*ProductResponse `json:"products"`
}
