package adapter

import (
	"context"
	"go-saga-pattern/commoner/discovery"
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/commoner/utils"
	"go-saga-pattern/proto/productpb"
	"go-saga-pattern/transaction-svc/internal/model"
	"log"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ProductAdapter interface {
	CheckProductAndReserve(ctx context.Context, transationID uuid.UUID, request []*model.CheckProductQuantity) ([]*model.ProductResponse, error)
	OwnerGetProduct(ctx context.Context, userID, productID uuid.UUID) (*model.ProductResponse, error)
}

type productAdapter struct {
	client productpb.ProductServiceClient
}

func NewProductAdapter(ctx context.Context, registry discovery.Registry, logs logs.Log) (ProductAdapter, error) {
	productServiecName := utils.GetEnv("PRODUCT_SVC_NAME") + "-grpc"
	logs.Info("Connecting to user service", zap.String("service_name", productServiecName))
	conn, err := discovery.ServiceConnection(ctx, productServiecName, registry, logs)
	if err != nil {
		return nil, err
	}

	log.Print("successfuly connected to user-svc-grpc")
	client := productpb.NewProductServiceClient(conn)

	return &productAdapter{
		client: client,
	}, nil
}

func (a *productAdapter) CheckProductAndReserve(ctx context.Context, transationID uuid.UUID, request []*model.CheckProductQuantity) ([]*model.ProductResponse, error) {
	requestPb := make([]*productpb.CheckProductQuantity, 0, len(request))
	for _, product := range request {
		requestPb = append(requestPb, &productpb.CheckProductQuantity{
			ProductId: product.ProductID.String(),
			Quantity:  int32(product.Quantity),
			Price:     float32(product.Price),
		})
	}

	processPhotoRequest := &productpb.CheckProductAndReserveRequest{
		TransactionId: transationID.String(),
		Products:      requestPb,
	}

	response, err := a.client.CheckProductAndReserve(ctx, processPhotoRequest)
	if err != nil {
		return nil, helper.FromGRPCError(err)
	}

	products := make([]*model.ProductResponse, 0, len(response.Products))
	for _, product := range response.Products {
		products = append(products, &model.ProductResponse{
			ID:          product.Id,
			Quantity:    int(product.Quantity),
			Price:       float64(product.Price),
			Name:        product.Name,
			Description: product.Description,
		})
	}

	return products, nil
}

func (a *productAdapter) OwnerGetProduct(ctx context.Context, userID, productID uuid.UUID) (*model.ProductResponse, error) {
	ownerGetProductRequest := &productpb.OwnerGetProductRequest{
		ProductId: productID.String(),
		UserId:    userID.String(),
	}

	response, err := a.client.OwnerGetProduct(ctx, ownerGetProductRequest)
	if err != nil {
		return nil, helper.FromGRPCError(err)
	}

	return &model.ProductResponse{
		ID:          response.Product.Id,
		Quantity:    int(response.Product.Quantity),
		Price:       float64(response.Product.Price),
		Name:        response.Product.Name,
		Description: response.Product.Description,
	}, nil

}
