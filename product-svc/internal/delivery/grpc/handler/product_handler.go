package handler

import (
	"context"
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/product-svc/internal/model"
	"go-saga-pattern/product-svc/internal/usecase"
	"go-saga-pattern/proto/productpb"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductHandler struct {
	productTranscationUC usecase.ProductTransactionUseCase
	productpb.UnimplementedProductServiceServer
}

func NewProductHandler(server *grpc.Server, productTranscationUC usecase.ProductTransactionUseCase) {
	handler := &ProductHandler{
		productTranscationUC: productTranscationUC,
	}
	productpb.RegisterProductServiceServer(server, handler)
}

func (h *ProductHandler) CheckProductAndReserve(ctx context.Context, pbReq *productpb.CheckProductAndReserveRequest,
) (*productpb.CheckProductQuantityResponse, error) {
	if pbReq == nil {
	}

	parsedTransactionID, err := uuid.Parse(pbReq.GetTransactionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid transaction ID format")
	}

	products := make([]*model.CheckProductQuantity, 0, len(pbReq.GetProducts()))
	for _, productPb := range pbReq.GetProducts() {
		if productPb.GetQuantity() <= 0 {
			return nil, status.Error(codes.InvalidArgument, "Product quantity must be greater than zero")
		}

		productID, err := uuid.Parse(productPb.GetProductId())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "Invalid product ID format")
		}
		products = append(products, &model.CheckProductQuantity{
			ProductID: productID,
			Quantity:  int(productPb.GetQuantity()),
			Price:     float64(productPb.GetPrice()),
		})

	}

	request := &model.CheckProductsQuantityRequest{
		TransactionID: parsedTransactionID,
		Products:      products,
	}

	response, err := h.productTranscationUC.CheckProductsAndReserve(ctx, request)
	if err != nil {
		return nil, helper.ErrGRPC(err)
	}

	productResponsePb := make([]*productpb.Product, 0, len(response.Products))
	for _, product := range response.Products {
		productResponsePb = append(productResponsePb, &productpb.Product{
			Id:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       float32(product.Price),
			Quantity:    int32(product.Quantity),
		})
	}

	pbResponse := &productpb.CheckProductQuantityResponse{
		Products: productResponsePb,
	}

	return pbResponse, nil
}
