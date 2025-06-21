package usecase

import (
	"context"
	errorcode "go-saga-pattern/commoner/constant/errcode"
	"go-saga-pattern/commoner/constant/message"
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/helper/nullable"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/commoner/web"
	"go-saga-pattern/product-svc/internal/entity"
	"go-saga-pattern/product-svc/internal/model"
	"go-saga-pattern/product-svc/internal/model/converter"
	"go-saga-pattern/product-svc/internal/repository"
	"go-saga-pattern/product-svc/internal/repository/store"
	"strings"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type ProductUseCase interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.ProductResponse, error)
	GetBySlug(ctx context.Context, slug string) (*model.ProductResponse, error)
	OwnerCreate(ctx context.Context, request *model.CreateProductRequest) (*model.ProductResponse, error)
	OwnerDelete(ctx context.Context, request *model.DeleteProductRequest) error
	OwnerSearch(ctx context.Context, request *model.OwnerSearchProductsRequest) ([]*model.ProductResponse, *web.PageMetadata, error)
	OwnerUpdate(ctx context.Context, request *model.UpdateProductRequest) (*model.ProductResponse, error)
	PublicSearch(ctx context.Context, request *model.PublicSearchProductsRequest) ([]*model.ProductResponse, *web.PageMetadata, error)
}

type productUseCase struct {
	productRepository repository.ProductRepository
	databaseStore     store.DatabaseStore
	validator         helper.CustomValidator
	log               logs.Log
}

func NewProductUseCase(productRepository repository.ProductRepository,
	databaseStore store.DatabaseStore, validator helper.CustomValidator,
	log logs.Log,
) ProductUseCase {
	return &productUseCase{
		productRepository: productRepository,
		databaseStore:     databaseStore,
		validator:         validator,
		log:               log,
	}
}

func (uc *productUseCase) OwnerCreate(ctx context.Context, request *model.CreateProductRequest) (*model.ProductResponse, error) {
	if validatonErrs := uc.validator.ValidateUseCase(request); validatonErrs != nil {
		return nil, validatonErrs
	}

	slug := slug.Make(request.Name)

	isExistsByNameOrSlug, err := uc.productRepository.ExistsByNameOrSlug(ctx, uc.databaseStore, request.Name, slug)
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

	createdProduct, err := uc.productRepository.Insert(ctx, uc.databaseStore, product)
	if err != nil {
		return nil, helper.WrapInternalServerError(uc.log, "failed to insert product", err)
	}

	return converter.ProductToResponse(createdProduct), nil
}

func (uc *productUseCase) GetByID(ctx context.Context, id uuid.UUID) (*model.ProductResponse, error) {
	product, err := uc.productRepository.FindByID(ctx, uc.databaseStore, id)
	if err != nil {
		if strings.Contains(err.Error(), pgx.ErrNoRows.Error()) {
			return nil, helper.NewUseCaseError(errorcode.ErrResourceNotFound, message.ProductNotFound)
		}
		return nil, helper.WrapInternalServerError(uc.log, "failed to find product by id", err)
	}

	return converter.ProductToResponse(product), nil
}

func (uc *productUseCase) GetBySlug(ctx context.Context, slug string) (*model.ProductResponse, error) {
	product, err := uc.productRepository.FindBySlug(ctx, uc.databaseStore, slug)
	if err != nil {
		if strings.Contains(err.Error(), pgx.ErrNoRows.Error()) {
			return nil, helper.NewUseCaseError(errorcode.ErrResourceNotFound, message.ProductNotFound)
		}
		return nil, helper.WrapInternalServerError(uc.log, "failed to find product by id", err)
	}

	return converter.ProductToResponse(product), nil
}

// ISSUE product doesnt populated
func (uc *productUseCase) OwnerUpdate(ctx context.Context, request *model.UpdateProductRequest) (*model.ProductResponse, error) {
	if validatonErrs := uc.validator.ValidateUseCase(request); validatonErrs != nil {
		return nil, validatonErrs
	}

	product, err := uc.productRepository.FindByIDAndUserID(ctx, uc.databaseStore, request.ID, request.UserID)
	if err != nil {
		if strings.Contains(err.Error(), pgx.ErrNoRows.Error()) {
			return nil, helper.NewUseCaseError(errorcode.ErrResourceNotFound, message.ProductNotFound)
		}
		uc.log.Error("failed to find product by id", zap.Error(err), zap.String("product_id", request.ID.String()))
		return nil, helper.WrapInternalServerError(uc.log, "failed to find product by id", err)
	}

	if request.Name != "" && request.Name != product.Name {
		slug := slug.Make(request.Name)

		isExistsByNameOrSlug, err := uc.productRepository.ExistByNameOrSlugExceptHerself(ctx, uc.databaseStore, request.Name, slug, request.ID)
		if err != nil {
			uc.log.Error("failed to check product exists by name or slug", zap.Error(err), zap.String("product_id", request.ID.String()))
			return nil, helper.WrapInternalServerError(uc.log, "failed to check product exists by name or slug", err)
		}

		if isExistsByNameOrSlug {
			return nil, helper.NewUseCaseError(errorcode.ErrAlreadyExists, message.ProductIsExistsByNameOrSlug)
		}

		product.Name = request.Name
		product.Slug = slug
	}

	product.ID = request.ID
	product.UserID = request.UserID
	product.Description = nullable.ToSQLString(request.Description)
	product.Price = request.Price
	product.Quantity = request.Quantity

	product, err = uc.productRepository.UpdateByID(ctx, uc.databaseStore, product)
	if err != nil {
		uc.log.Error("failed to update product", zap.Error(err), zap.String("product_id", request.ID.String()))
		return nil, helper.WrapInternalServerError(uc.log, "failed to update product", err)
	}

	return converter.ProductToResponse(product), nil
}

func (uc *productUseCase) OwnerDelete(ctx context.Context, request *model.DeleteProductRequest) error {
	if validatonErrs := uc.validator.ValidateUseCase(request); validatonErrs != nil {
		return validatonErrs
	}

	product, err := uc.productRepository.FindByIDAndUserID(ctx, uc.databaseStore, request.ID, request.UserID)
	if err != nil {
		if strings.Contains(err.Error(), pgx.ErrNoRows.Error()) {
			return helper.NewUseCaseError(errorcode.ErrResourceNotFound, message.ProductNotFound)
		}
		return helper.WrapInternalServerError(uc.log, "failed to find product by id", err)
	}

	if err := uc.productRepository.DeleteByIDAndUserID(ctx, uc.databaseStore, product.ID, product.UserID); err != nil {
		if strings.Contains(err.Error(), message.InternalNoRowsAffected) {
			return helper.NewUseCaseError(errorcode.ErrResourceNotFound, message.ProductNotFoundOrAlreadyDeleted)
		}
		return helper.WrapInternalServerError(uc.log, "failed to delete product", err)
	}

	uc.log.Info("Product deleted successfully", zap.String("product_id", product.ID.String()))

	return nil
}

func (uc *productUseCase) OwnerSearch(ctx context.Context, request *model.OwnerSearchProductsRequest) ([]*model.ProductResponse, *web.PageMetadata, error) {
	products, metadata, err := uc.productRepository.OwnerFindAll(ctx, uc.databaseStore, request)
	if err != nil {
		return nil, nil, helper.WrapInternalServerError(uc.log, "failed to find owner products by user id", err)
	}

	if products == nil {
		return nil, metadata, nil
	}

	return converter.ProductsWithTotalToResponses(products), metadata, nil
}

func (uc *productUseCase) PublicSearch(ctx context.Context, request *model.PublicSearchProductsRequest) ([]*model.ProductResponse, *web.PageMetadata, error) {
	products, metadata, err := uc.productRepository.PublicFindAll(ctx, uc.databaseStore, request.Limit, request.Page)
	if err != nil {
		return nil, nil, helper.WrapInternalServerError(uc.log, "failed to find public products", err)
	}

	if products == nil {
		return nil, metadata, nil
	}

	return converter.ProductsWithTotalToResponses(products), metadata, nil
}
