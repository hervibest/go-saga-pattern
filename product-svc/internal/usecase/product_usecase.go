package usecase

import (
	"context"
	errorcode "go-saga-pattern/commoner/constant/errcode"
	"go-saga-pattern/commoner/constant/message"
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/helper/nullable"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/commoner/web"
	"go-saga-pattern/product-svc/internal/adapter"
	"go-saga-pattern/product-svc/internal/entity"
	"go-saga-pattern/product-svc/internal/model"
	"go-saga-pattern/product-svc/internal/model/converter"
	"go-saga-pattern/product-svc/internal/repository"
	"strings"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type ProductUseCase interface {
	GetById(ctx context.Context, id uuid.UUID) (*model.ProductResponse, error)
	GetBySlug(ctx context.Context, slug string) (*model.ProductResponse, error)
	OwnerCreate(ctx context.Context, request *model.CreateProductRequest) (*model.ProductResponse, error)
	OwnerDelete(ctx context.Context, request *model.DeleteProductRequest) error
	OwnerSearch(ctx context.Context, request *model.OwnerSearchProductsRequest) ([]*model.ProductResponse, *web.PageMetadata, error)
	OwnerUpdate(ctx context.Context, request *model.UpdateProductRequest) (*model.ProductResponse, error)
	PublicSearch(ctx context.Context, request *model.PublicSearchProductsRequest) ([]*model.ProductResponse, *web.PageMetadata, error)
}

type productUseCase struct {
	productRepository repository.ProductRepository
	databaseAdapter   adapter.DatabaseAdapter
	validator         helper.CustomValidator
	log               logs.Log
}

func NewProductUseCase(productRepository repository.ProductRepository,
	databaseAdapter adapter.DatabaseAdapter, validator helper.CustomValidator,
	log logs.Log,
) ProductUseCase {
	return &productUseCase{
		productRepository: productRepository,
		databaseAdapter:   databaseAdapter,
		validator:         validator,
		log:               log,
	}
}

func (uc *productUseCase) OwnerCreate(ctx context.Context, request *model.CreateProductRequest) (*model.ProductResponse, error) {
	if validatonErrs := uc.validator.ValidateUseCase(request); validatonErrs != nil {
		return nil, validatonErrs
	}

	slug := slug.Make(request.Name)

	isExistsByNameOrSlug, err := uc.productRepository.ExistsByNameOrSlug(ctx, uc.databaseAdapter, request.Name, slug)
	if err != nil {
		return nil, helper.WrapInternalServerError(uc.log, "failed to check product exists by name or slug", err)
	}

	if isExistsByNameOrSlug {
		return nil, helper.NewUseCaseError(errorcode.ErrAlreadyExists, message.ProductIsExistsByNameOrSlug)
	}

	product := &entity.Product{
		UserID: request.UserID,
		Name:   request.Name,
		Slug:   slug,
		Description: nullable.ToSQLString(
			request.Description,
		),
		Price:    request.Price,
		Quantity: request.Quantity,
	}

	createdProduct, err := uc.productRepository.Insert(ctx, uc.databaseAdapter, product)
	if err != nil {
		return nil, helper.WrapInternalServerError(uc.log, "failed to insert product", err)
	}

	return converter.ProductToResponse(createdProduct), nil
}

func (uc *productUseCase) GetById(ctx context.Context, id uuid.UUID) (*model.ProductResponse, error) {
	product, err := uc.productRepository.FindByID(ctx, uc.databaseAdapter, id)
	if err != nil {
		if strings.Contains(err.Error(), pgx.ErrNoRows.Error()) {
			return nil, helper.NewUseCaseError(errorcode.ErrResourceNotFound, message.ProductNotFound)
		}
		return nil, helper.WrapInternalServerError(uc.log, "failed to find product by id", err)
	}

	return converter.ProductToResponse(product), nil
}

func (uc *productUseCase) GetBySlug(ctx context.Context, slug string) (*model.ProductResponse, error) {
	product, err := uc.productRepository.FindBySlug(ctx, uc.databaseAdapter, slug)
	if err != nil {
		if strings.Contains(err.Error(), pgx.ErrNoRows.Error()) {
			return nil, helper.NewUseCaseError(errorcode.ErrResourceNotFound, message.ProductNotFound)
		}
		return nil, helper.WrapInternalServerError(uc.log, "failed to find product by id", err)
	}

	return converter.ProductToResponse(product), nil
}

func (uc *productUseCase) OwnerUpdate(ctx context.Context, request *model.UpdateProductRequest) (*model.ProductResponse, error) {
	if validatonErrs := uc.validator.ValidateUseCase(request); validatonErrs != nil {
		return nil, validatonErrs
	}

	product, err := uc.productRepository.FindByIDAndUserID(ctx, uc.databaseAdapter, request.Id, request.UserID)
	if err != nil {
		if strings.Contains(err.Error(), pgx.ErrNoRows.Error()) {
			return nil, helper.NewUseCaseError(errorcode.ErrResourceNotFound, message.ProductNotFound)
		}
		return nil, helper.WrapInternalServerError(uc.log, "failed to find product by id", err)
	}

	if request.Name != "" && request.Name != product.Name {
		slug := slug.Make(request.Name)

		isExistsByNameOrSlug, err := uc.productRepository.ExistByNameOrSlugExceptHerself(ctx, uc.databaseAdapter, request.Name, slug, request.Id)
		if err != nil {
			return nil, helper.WrapInternalServerError(uc.log, "failed to check product exists by name or slug", err)
		}

		if isExistsByNameOrSlug {
			return nil, helper.NewUseCaseError(errorcode.ErrAlreadyExists, message.ProductIsExistsByNameOrSlug)
		}

		product.Name = request.Name
		product.Slug = slug
	}

	product.Description = nullable.ToSQLString(request.Description)
	product.Price = request.Price
	product.Quantity = request.Quantity

	product, err = uc.productRepository.UpdateById(ctx, uc.databaseAdapter, product)
	if err != nil {
		return nil, helper.WrapInternalServerError(uc.log, "failed to update product", err)
	}

	return converter.ProductToResponse(product), nil
}

func (uc *productUseCase) OwnerDelete(ctx context.Context, request *model.DeleteProductRequest) error {
	if validatonErrs := uc.validator.ValidateUseCase(request); validatonErrs != nil {
		return validatonErrs
	}

	product, err := uc.productRepository.FindByIDAndUserID(ctx, uc.databaseAdapter, request.Id, request.UserID)
	if err != nil {
		if strings.Contains(err.Error(), pgx.ErrNoRows.Error()) {
			return helper.NewUseCaseError(errorcode.ErrResourceNotFound, message.ProductNotFound)
		}
		return helper.WrapInternalServerError(uc.log, "failed to find product by id", err)
	}

	if err := uc.productRepository.DeleteByIDAndUserID(ctx, uc.databaseAdapter, product.Id, product.UserID); err != nil {
		if strings.Contains(err.Error(), "product not found or already deleted") {
			return helper.NewUseCaseError(errorcode.ErrResourceNotFound, message.ProductNotFoundOrAlreadyDeleted)
		}
		return helper.WrapInternalServerError(uc.log, "failed to delete product", err)
	}

	uc.log.Info("Product deleted successfully", zap.String("product_id", product.Id.String()))

	return nil
}

func (uc *productUseCase) OwnerSearch(ctx context.Context, request *model.OwnerSearchProductsRequest) ([]*model.ProductResponse, *web.PageMetadata, error) {
	products, metadata, err := uc.productRepository.OwnerFindAll(ctx, uc.databaseAdapter, request.UserID, request.Limit, request.Page)
	if err != nil {
		return nil, nil, helper.WrapInternalServerError(uc.log, "failed to find products by user id", err)
	}

	if products == nil {
		return nil, metadata, nil
	}

	return converter.ProductsWithTotalToResponses(products), metadata, nil
}

func (uc *productUseCase) PublicSearch(ctx context.Context, request *model.PublicSearchProductsRequest) ([]*model.ProductResponse, *web.PageMetadata, error) {
	products, metadata, err := uc.productRepository.PublicFindAll(ctx, uc.databaseAdapter, request.Limit, request.Page)
	if err != nil {
		return nil, nil, helper.WrapInternalServerError(uc.log, "failed to find products by user id", err)
	}

	if products == nil {
		return nil, metadata, nil
	}

	return converter.ProductsWithTotalToResponses(products), metadata, nil
}

//TODO Transaction Wrapper

func (uc *productUseCase) CheckProductsAndReserve(ctx context.Context, request *model.CheckProductsQuantity) error {
	productIDs := make([]uuid.UUID, 0, len(request.Products))
	for _, productReq := range request.Products {
		productIDs = append(productIDs, productReq.ProductID)
	}

	products, err := uc.productRepository.FindManyByIDs(ctx, uc.databaseAdapter, productIDs, true)
	if err != nil {
		return helper.WrapInternalServerError(uc.log, "failed to find products by user id", err)
	}

	if len(productIDs) != len(products) {
		return helper.NewUseCaseError(errorcode.ErrInvalidArgument, "product not found")
	}

	productTransactions := make([]*entity.ProductTransaction, 0, len(products))
	for idx, product := range products {
		productReq := request.Products[idx]
		if productReq.Quantity > product.Quantity {
			return helper.NewUseCaseError(errorcode.ErrInvalidArgument, "product not found")
		}

		if productReq.Price != product.Price {
			return helper.NewUseCaseError(errorcode.ErrInvalidArgument, "product not found")
		}

		if err := uc.productRepository.ReduceQuantity(ctx, uc.databaseAdapter, productReq.ProductID, productReq.Quantity); err != nil {

		}

		//TODO finalize
		productTransaction := &entity.ProductTransaction{
			TransactionID: request.TransactionID,
			ProductID:     product.Id,
		}

		productTransactions = append(productTransactions, productTransaction)
	}

}
