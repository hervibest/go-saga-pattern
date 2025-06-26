package usecase

import (
	"context"
	"go-saga-pattern/commoner/constant/enum"
	errorcode "go-saga-pattern/commoner/constant/errcode"
	"go-saga-pattern/commoner/constant/message"
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/product-svc/internal/entity"
	"go-saga-pattern/product-svc/internal/model"
	"go-saga-pattern/product-svc/internal/model/converter"
	"go-saga-pattern/product-svc/internal/repository"
	"go-saga-pattern/product-svc/internal/repository/store"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ProductTransactionUseCase interface {
	CancelProductTransactions(ctx context.Context, request *model.CancelProductTransactionsRequest) error
	CheckProductsAndReserve(ctx context.Context, request *model.CheckProductsQuantityRequest) (*model.CheckProductsQuantityRequestResponse, error)
	CommitProductTransactionsRequest(ctx context.Context, request *model.CommitProductTransactionsRequest) error
	ExpireProductTransactions(ctx context.Context, request *model.ExpireProductTransactionsRequest) error
	SettleProducts(ctx context.Context, request *model.SettleProductTransactionRequest) error
	updateAndRestoreProductTransactions(ctx context.Context, transactionID uuid.UUID, status enum.ProductTransactionStatusEnum) error
}

type productTransactionUseCase struct {
	productRepository      repository.ProductRepository
	productTransactionRepo repository.ProductTransactionRepository
	databaseStore          store.DatabaseStore
	validator              helper.CustomValidator
	log                    logs.Log
}

func NewProductTransactionUseCase(productRepository repository.ProductRepository,
	productTransactionRepo repository.ProductTransactionRepository, databaseStore store.DatabaseStore,
	validator helper.CustomValidator, log logs.Log,
) ProductTransactionUseCase {
	return &productTransactionUseCase{
		productRepository:      productRepository,
		productTransactionRepo: productTransactionRepo,
		databaseStore:          databaseStore,
		validator:              validator,
		log:                    log,
	}
}

// TODO : Cannot deleted product when product transaction exists and status != canceled or expired
func (uc *productTransactionUseCase) CancelProductTransactions(ctx context.Context, request *model.CancelProductTransactionsRequest) error {
	if err := uc.updateAndRestoreProductTransactions(ctx, request.TransactionID, enum.ProductTransactionStatusCanceled); err != nil {
		return err
	}

	return nil
}

// TODO : Cannot deleted product when product transaction exists and status != canceled or expired
func (uc *productTransactionUseCase) ExpireProductTransactions(ctx context.Context, request *model.ExpireProductTransactionsRequest) error {
	if err := uc.updateAndRestoreProductTransactions(ctx, request.TransactionID, enum.ProductTransactionStatusExpired); err != nil {
		return err
	}

	return nil
}

// TODO : Cannot deleted product when product transaction exists and status != canceled or expired
func (uc *productTransactionUseCase) updateAndRestoreProductTransactions(ctx context.Context, transactionID uuid.UUID,
	status enum.ProductTransactionStatusEnum) error {
	if err := store.BeginTransaction(ctx, uc.log, uc.databaseStore, func(tx store.Transaction) error {
		productTransactions, err := uc.productTransactionRepo.FindManyByTrxID(ctx, uc.databaseStore, transactionID, true)
		if err != nil {
			return helper.WrapInternalServerError(uc.log, "failed to find products by user id", err)
		}

		if len(productTransactions) == 0 {
			return helper.NewUseCaseError(errorcode.ErrInvalidArgument, message.ProductTranscationNotFound)
		}

		productIDs := make([]uuid.UUID, 0, len(productTransactions))
		for _, productReq := range productTransactions {
			productIDs = append(productIDs, productReq.ProductID)
		}

		products, err := uc.productRepository.FindManyByIDs(ctx, uc.databaseStore, productIDs, enum.LockTypeUpdateEnum)
		if err != nil {
			return helper.WrapInternalServerError(uc.log, "failed to find products by user id", err)
		}

		// CASE WHEN PRODUCT ALREADY DELETED (HAVE TO HANDLE PRODUCT CONSISTENCY WELL)
		if len(productIDs) != len(products) {
			return helper.NewUseCaseError(errorcode.ErrInvalidArgument, "product product")
		}

		// NO NEED TO VALIDATE QUANTITY OR PRICE
		for _, productTransaction := range productTransactions {
			if err := uc.productRepository.RestoreQuantity(ctx, uc.databaseStore, productTransaction.ProductID,
				productTransaction.Quantity); err != nil {
				if strings.Contains(err.Error(), message.InternalNoRowsAffected) {
					return helper.NewUseCaseError(errorcode.ErrResourceNotFound, message.ProductNotFoundOrAlreadyDeleted)
				}
				return err
			}
		}

		err = uc.productTransactionRepo.UpdateStatus(ctx, uc.databaseStore, transactionID, status)
		if err != nil {
			return helper.WrapInternalServerError(uc.log, "failed to update many product transactions", err)
		}
		return nil
	}); err != nil {
		uc.log.Error("failed to check products and reserve", zap.Error(err))
		return err
	}

	return nil
}

func (uc *productTransactionUseCase) CheckProductsAndReserve(ctx context.Context, request *model.CheckProductsQuantityRequest) (*model.CheckProductsQuantityRequestResponse, error) {
	uc.log.Info("[CheckProductsAndReserve] Starting to check products and reserve quantities", zap.String("TransactionID", request.TransactionID.String()))

	if request == nil {
		uc.log.Warn("[CheckProductsAndReserve] Received nil request")
		return nil, helper.NewUseCaseError(errorcode.ErrInvalidArgument, message.InvalidRequestError)
	}

	uc.log.Info("[CheckProductsAndReserve] Number of products in request", zap.Int("ProductCount", len(request.Products)))

	productReqMap := make(map[uuid.UUID]*model.CheckProductQuantity)
	var productIDs []uuid.UUID

	for i, productReq := range request.Products {
		uc.log.Info("[CheckProductsAndReserve] Processing product", zap.Int("Index", i), zap.String("ProductID", productReq.ProductID.String()))

		if productReq == nil {
			uc.log.Warn("[CheckProductsAndReserve] Received nil product request at position", zap.Int("Index", i))
			continue
		}

		productReqMap[productReq.ProductID] = productReq
		productIDs = append(productIDs, productReq.ProductID)

		uc.log.Info("[CheckProductsAndReserve] Successfully processed product")
	}

	uc.log.Info("[CheckProductsAndReserve] All products processed successfully", zap.Int("TotalProducts", len(productIDs)))
	var products []*entity.Product

	if err := store.BeginTransaction(ctx, uc.log, uc.databaseStore, func(tx store.Transaction) error {
		var err error
		products, err = uc.productRepository.FindManyByIDs(ctx, tx, productIDs, enum.LockTypeUpdateEnum)
		if err != nil {
			return helper.WrapInternalServerError(uc.log, "failed to find products by user id", err)
		}

		if len(productIDs) != len(products) {
			return helper.NewUseCaseError(errorcode.ErrInvalidArgument, message.ProductNotFound)
		}

		productTransactions := make([]*entity.ProductTransaction, 0, len(products))
		for _, product := range products {
			productReq := productReqMap[product.ID]
			if product.Quantity == 0 {
				return helper.NewUseCaseError(errorcode.ErrInvalidArgument, message.ProductOutOfStock)
			}

			if productReq.Quantity > product.Quantity {
				return helper.NewUseCaseError(errorcode.ErrInvalidArgument, message.RequestedProductMoreThanAvailable)
			}

			if productReq.Price != product.Price {
				return helper.NewUseCaseError(errorcode.ErrInvalidArgument, message.PriceChanged)
			}

			if err := uc.productRepository.ReduceQuantity(ctx, tx, productReq.ProductID, productReq.Quantity); err != nil {
				if strings.Contains(err.Error(), message.InternalNoRowsAffected) {
					return helper.NewUseCaseError(errorcode.ErrResourceNotFound, message.ProductNotFoundOrAlreadyDeleted)
				}
				return helper.WrapInternalServerError(uc.log, "failed to reduce product quantity", err)
			}

			//TODO finalize
			productTransaction := &entity.ProductTransaction{
				TransactionID: request.TransactionID,
				ProductID:     product.ID,
				Status:        enum.ProductTransactionStatusComitted,
				Quantity:      productReq.Quantity,
				TotalPrice:    product.Price,
			}

			productTransactions = append(productTransactions, productTransaction)
		}

		productTransactions, err = uc.productTransactionRepo.InsertMany(ctx, tx, productTransactions)
		if err != nil {
			return helper.WrapInternalServerError(uc.log, "failed to insert many product transactions", err)
		}
		return nil
	}); err != nil {
		uc.log.Error("failed to check products and reserve", zap.Error(err))
		return nil, err
	}

	return converter.ProductsToCheckQuantityResponse(request.TransactionID, products), nil
}

// TODO : Cannot deleted product when product transaction exists and status != canceled or expired
func (uc *productTransactionUseCase) CommitProductTransactionsRequest(ctx context.Context, request *model.CommitProductTransactionsRequest) error {
	if err := uc.updateProductTransactionsStatus(ctx, request.TransactionID, enum.ProductTransactionStatusComitted); err != nil {
		return err
	}

	return nil
}

// TODO : Cannot deleted product when product transaction exists and status != canceled or expired
func (uc *productTransactionUseCase) SettleProducts(ctx context.Context, request *model.SettleProductTransactionRequest) error {
	if err := uc.updateProductTransactionsStatus(ctx, request.TransactionID, enum.ProductTransactionStatusSettled); err != nil {
		return err
	}

	return nil
}

// TODO : Cannot deleted product when product transaction exists and status != canceled or expired
func (uc *productTransactionUseCase) updateProductTransactionsStatus(ctx context.Context, transactionID uuid.UUID,
	status enum.ProductTransactionStatusEnum) error {
	if err := store.BeginTransaction(ctx, uc.log, uc.databaseStore, func(tx store.Transaction) error {
		productTransactions, err := uc.productTransactionRepo.FindManyByTrxID(ctx, uc.databaseStore, transactionID, true)
		if err != nil {
			return helper.WrapInternalServerError(uc.log, "failed to find products by user id", err)
		}

		if len(productTransactions) == 0 {
			return helper.NewUseCaseError(errorcode.ErrInvalidArgument, message.ProductTranscationNotFound)
		}

		productIDs := make([]uuid.UUID, 0, len(productTransactions))
		for _, productReq := range productTransactions {
			productIDs = append(productIDs, productReq.ProductID)
		}

		products, err := uc.productRepository.FindManyByIDs(ctx, uc.databaseStore, productIDs, enum.LockTypeShareEnum)
		if err != nil {
			return helper.WrapInternalServerError(uc.log, "failed to find products by user id", err)
		}

		// CASE WHEN PRODUCT ALREADY DELETED (HAVE TO HANDLE PRODUCT CONSISTENCY WELL) (THIS SHOULD BE IMPOSSIBLE CASE)
		if len(productIDs) != len(products) {
			return helper.NewUseCaseError(errorcode.ErrInvalidArgument, message.ProductNotFoundOrAlreadyDeleted)
		}

		err = uc.productTransactionRepo.UpdateStatus(ctx, uc.databaseStore, transactionID, status)
		if err != nil {
			uc.log.Error("failed to update many product transactions", zap.Error(err))
			return helper.WrapInternalServerError(uc.log, "failed to insert many product transactions", err)
		}
		return nil
	}); err != nil {
		uc.log.Error("failed to check products and reserve", zap.Error(err))
		return err
	}

	return nil
}
