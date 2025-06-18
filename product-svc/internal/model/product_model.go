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
	Id string `json:"id" validate:"required,uuid"`
}

type OwnerSearchProductsRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required,uuid"`
	Page   int       `json:"page" validate:"required,min=1"`
	Limit  int       `json:"limit" validate:"required,min=1,max=100"`
}

type PublicSearchProductsRequest struct {
	Page  int `json:"page" validate:"required,min=1"`
	Limit int `json:"limit" validate:"required,min=1,max=100"`
}

type UpdateProductRequest struct {
	Id          uuid.UUID `json:"id" validate:"required,uuid"`
	UserID      uuid.UUID `json:"user_id" validate:"required,uuid"`
	Name        string    `json:"name" validate:"required,omitempty"`
	Description *string   `json:"description" validate:"omitempty"`
	Price       float64   `json:"price" validate:"required,omitempty,gt=0"`
	Quantity    int       `json:"quantity" validate:"required,omitempty,gt=0"`
}

type DeleteProductRequest struct {
	Id     uuid.UUID `json:"id" validate:"required,uuid"`
	UserID uuid.UUID `json:"user_id" validate:"required,uuid"`
}

type ProductResponse struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	CreatedAt   string  `json:"created_at,omitempty"`
	UpdatedAt   string  `json:"updated_at,omitempty"`
	DeletedAt   string  `json:"deleted_at,omitempty"`
}
