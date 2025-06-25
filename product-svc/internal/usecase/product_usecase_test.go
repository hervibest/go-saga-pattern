package usecase_test

import (
	"context"
	"errors"
	errorcode "go-saga-pattern/commoner/constant/errcode"
	"go-saga-pattern/commoner/constant/message"
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/helper/nullable"
	mockhelper "go-saga-pattern/commoner/mocks/commoner/helper"
	mocklogs "go-saga-pattern/commoner/mocks/commoner/logs"
	"go-saga-pattern/commoner/web"
	"go-saga-pattern/product-svc/internal/entity"
	mockrepository "go-saga-pattern/product-svc/internal/mocks/repository"
	mockstore "go-saga-pattern/product-svc/internal/mocks/store"
	"go-saga-pattern/product-svc/internal/model"
	"go-saga-pattern/product-svc/internal/repository/store"
	"go-saga-pattern/product-svc/internal/usecase"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestProductUseCase_OwnerCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := mockstore.NewMockDatabaseStore(ctrl)
	mockProductRepo := mockrepository.NewMockProductRepository(ctrl)
	// mockUserAdapter := mockadapter.NewMockUserAdapter(ctrl)
	mockLogs := mocklogs.NewMockLog(ctrl)
	mockValidator := mockhelper.NewMockCustomValidator(ctrl)

	uc := usecase.NewProductUseCase(
		mockProductRepo,
		mockDB,
		mockValidator,
		mockLogs,
	)

	mockLogs.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockLogs.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	t.Run("successful create product", func(t *testing.T) {
		ctx := context.Background()
		userID := uuid.New()
		description := "Test Product Pertama "
		req := &model.CreateProductRequest{
			UserID:      userID,
			Name:        "Product pertama",
			Description: &description,
			Price:       30000000,
			Quantity:    10,
		}

		// Mock validation
		mockValidator.EXPECT().ValidateUseCase(req).Return(nil)

		// Mock existence check
		mockProductRepo.EXPECT().ExistsByNameOrSlug(ctx, mockDB, req.Name, gomock.Any()).Return(false, nil)

		// Mock user creation
		expectedProduct := &entity.Product{
			UserID: req.UserID,
			Name:   req.Name,
			Slug:   gomock.Any().String(),
			Description: nullable.ToSQLString(
				req.Description,
			),
			Price:    10000,
			Quantity: 10000,
		}

		mockProductRepo.EXPECT().Insert(ctx, mockDB, gomock.Any()).DoAndReturn(
			func(ctx context.Context, db store.DatabaseStore, user *entity.Product) (*entity.Product, error) {
				// Verify password was hashed
				return expectedProduct, nil
			},
		)

		resp, err := uc.OwnerCreate(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedProduct.Name, resp.Name)
		assert.Equal(t, expectedProduct.Description.String, resp.Description)
		assert.Equal(t, expectedProduct.Quantity, resp.Quantity)
	})

	t.Run("invalid request", func(t *testing.T) {
		ctx := context.Background()
		req := &model.CreateProductRequest{}

		expectedErr := &helper.UseCaseValError{}
		mockValidator.EXPECT().ValidateUseCase(req).Return(expectedErr)

		resp, err := uc.OwnerCreate(ctx, req)
		assert.Nil(t, resp)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("product already exists", func(t *testing.T) {
		ctx := context.Background()
		userID := uuid.New()

		description := "Test Product Pertama "

		req := &model.CreateProductRequest{
			UserID:      userID,
			Name:        "Product pertama",
			Description: &description,
			Price:       30000000,
			Quantity:    10,
		}
		mockValidator.EXPECT().ValidateUseCase(req).Return(nil)
		mockProductRepo.EXPECT().ExistsByNameOrSlug(ctx, mockDB, req.Name, gomock.Any()).Return(true, nil)

		resp, err := uc.OwnerCreate(ctx, req)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrAlreadyExists, err.(*helper.AppError).Code)
	})

	t.Run("database error on existence check", func(t *testing.T) {
		ctx := context.Background()
		userID := uuid.New()

		description := "Test Product Pertama "

		req := &model.CreateProductRequest{
			UserID:      userID,
			Name:        "Product pertama",
			Description: &description,
			Price:       30000000,
			Quantity:    10,
		}
		mockValidator.EXPECT().ValidateUseCase(req).Return(nil)
		mockProductRepo.EXPECT().ExistsByNameOrSlug(ctx, mockDB, req.Name, gomock.Any()).Return(false, errors.New("internal db down"))

		resp, err := uc.OwnerCreate(ctx, req)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrInternal, err.(*helper.AppError).Code)
	})

	t.Run("database error on insert product", func(t *testing.T) {
		ctx := context.Background()
		userID := uuid.New()
		description := "Test Product Pertama "
		req := &model.CreateProductRequest{
			UserID:      userID,
			Name:        "Product pertama",
			Description: &description,
			Price:       30000000,
			Quantity:    10,
		}

		// Mock validation
		mockValidator.EXPECT().ValidateUseCase(req).Return(nil)

		// Mock existence check
		mockProductRepo.EXPECT().ExistsByNameOrSlug(ctx, mockDB, req.Name, gomock.Any()).Return(false, nil)
		mockProductRepo.EXPECT().Insert(ctx, mockDB, gomock.Any()).DoAndReturn(
			func(ctx context.Context, db store.DatabaseStore, user *entity.Product) (*entity.Product, error) {
				// Verify password was hashed
				return nil, errors.New("db down")
			},
		)

		resp, err := uc.OwnerCreate(ctx, req)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrInternal, err.(*helper.AppError).Code)
	})
}

func TestProductUseCase_GetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := mockstore.NewMockDatabaseStore(ctrl)
	mockProductRepo := mockrepository.NewMockProductRepository(ctrl)
	// mockUserAdapter := mockadapter.NewMockUserAdapter(ctrl)
	mockLogs := mocklogs.NewMockLog(ctrl)
	mockValidator := mockhelper.NewMockCustomValidator(ctrl)

	uc := usecase.NewProductUseCase(
		mockProductRepo,
		mockDB,
		mockValidator,
		mockLogs,
	)

	mockLogs.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockLogs.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	t.Run("product not found", func(t *testing.T) {
		ctx := context.Background()
		productID := uuid.New()

		// Mock existence check
		mockProductRepo.EXPECT().FindByID(ctx, mockDB, productID).Return(nil, pgx.ErrNoRows)
		resp, err := uc.GetByID(ctx, productID)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrResourceNotFound, err.(*helper.AppError).Code)
	})

	t.Run("db down", func(t *testing.T) {
		ctx := context.Background()
		productID := uuid.New()

		// Mock existence check
		mockProductRepo.EXPECT().FindByID(ctx, mockDB, productID).Return(nil, errors.New("db down"))
		resp, err := uc.GetByID(ctx, productID)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrInternal, err.(*helper.AppError).Code)
	})

	t.Run("success get by id", func(t *testing.T) {
		ctx := context.Background()
		productID := uuid.New()
		description := "Description"
		// Mock user creation
		expectedProduct := &entity.Product{
			UserID: uuid.New(),
			Name:   "Barang pertama",
			Slug:   gomock.Any().String(),
			Description: nullable.ToSQLString(
				&description,
			),
			Price:    10000,
			Quantity: 10000,
		}
		// Mock existence check
		mockProductRepo.EXPECT().FindByID(ctx, mockDB, productID).Return(expectedProduct, nil)
		resp, err := uc.GetByID(ctx, productID)
		assert.NoError(t, err)
		assert.Equal(t, expectedProduct.Name, resp.Name)
		assert.Equal(t, expectedProduct.Description.String, resp.Description)
		assert.Equal(t, expectedProduct.Quantity, resp.Quantity)
	})
}

func TestProductUseCase_GetBySlug(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := mockstore.NewMockDatabaseStore(ctrl)
	mockProductRepo := mockrepository.NewMockProductRepository(ctrl)
	// mockUserAdapter := mockadapter.NewMockUserAdapter(ctrl)
	mockLogs := mocklogs.NewMockLog(ctrl)
	mockValidator := mockhelper.NewMockCustomValidator(ctrl)

	uc := usecase.NewProductUseCase(
		mockProductRepo,
		mockDB,
		mockValidator,
		mockLogs,
	)

	mockLogs.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockLogs.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	t.Run("product not found", func(t *testing.T) {
		ctx := context.Background()
		productSlug := "product-abc"

		// Mock existence check
		mockProductRepo.EXPECT().FindBySlug(ctx, mockDB, productSlug).Return(nil, pgx.ErrNoRows)
		resp, err := uc.GetBySlug(ctx, productSlug)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrResourceNotFound, err.(*helper.AppError).Code)
	})

	t.Run("db down", func(t *testing.T) {
		ctx := context.Background()
		productSlug := "product-abc"
		// Mock existence check
		mockProductRepo.EXPECT().FindBySlug(ctx, mockDB, productSlug).Return(nil, errors.New("db down"))
		resp, err := uc.GetBySlug(ctx, productSlug)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrInternal, err.(*helper.AppError).Code)
	})

	t.Run("success get by slug", func(t *testing.T) {
		ctx := context.Background()
		productSlug := "product-abc"
		description := "Description"
		// Mock user creation
		expectedProduct := &entity.Product{
			UserID: uuid.New(),
			Name:   "Barang pertama",
			Slug:   gomock.Any().String(),
			Description: nullable.ToSQLString(
				&description,
			),
			Price:    10000,
			Quantity: 10000,
		}
		// Mock existence check
		mockProductRepo.EXPECT().FindBySlug(ctx, mockDB, productSlug).Return(expectedProduct, nil)
		resp, err := uc.GetBySlug(ctx, productSlug)
		assert.NoError(t, err)
		assert.Equal(t, expectedProduct.Name, resp.Name)
		assert.Equal(t, expectedProduct.Description.String, resp.Description)
		assert.Equal(t, expectedProduct.Quantity, resp.Quantity)
	})
}

func TestProductUseCase_OwnerUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := mockstore.NewMockDatabaseStore(ctrl)
	mockProductRepo := mockrepository.NewMockProductRepository(ctrl)
	// mockUserAdapter := mockadapter.NewMockUserAdapter(ctrl)
	mockLogs := mocklogs.NewMockLog(ctrl)
	mockValidator := mockhelper.NewMockCustomValidator(ctrl)

	uc := usecase.NewProductUseCase(
		mockProductRepo,
		mockDB,
		mockValidator,
		mockLogs,
	)

	mockLogs.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockLogs.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	t.Run("invalid request", func(t *testing.T) {
		ctx := context.Background()
		req := &model.UpdateProductRequest{}

		expectedErr := &helper.UseCaseValError{}
		mockValidator.EXPECT().ValidateUseCase(req).Return(expectedErr)

		resp, err := uc.OwnerUpdate(ctx, req)
		assert.Nil(t, resp)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("product not found", func(t *testing.T) {
		ctx := context.Background()
		userID := uuid.New()
		productID := uuid.New()

		description := "Test Product Pertama "

		req := &model.UpdateProductRequest{
			ID:          productID,
			UserID:      userID,
			Name:        "Product pertama",
			Description: &description,
			Price:       30000000,
			Quantity:    10,
		}
		mockValidator.EXPECT().ValidateUseCase(req).Return(nil)
		mockProductRepo.EXPECT().FindByIDAndUserID(ctx, mockDB, req.ID, req.UserID).Return(nil, pgx.ErrNoRows)

		resp, err := uc.OwnerUpdate(ctx, req)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrResourceNotFound, err.(*helper.AppError).Code)
	})
	t.Run("failed to get product by id and user id", func(t *testing.T) {
		ctx := context.Background()
		userID := uuid.New()
		productID := uuid.New()

		description := "Test Product Pertama "

		req := &model.UpdateProductRequest{
			ID:          productID,
			UserID:      userID,
			Name:        "Product pertama",
			Description: &description,
			Price:       30000000,
			Quantity:    10,
		}
		mockValidator.EXPECT().ValidateUseCase(req).Return(nil)
		mockProductRepo.EXPECT().FindByIDAndUserID(ctx, mockDB, req.ID, req.UserID).Return(nil, errors.New("db down"))

		resp, err := uc.OwnerUpdate(ctx, req)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrInternal, err.(*helper.AppError).Code)
	})
	t.Run("name or slug already used", func(t *testing.T) {
		ctx := context.Background()
		userID := uuid.New()
		productID := uuid.New()

		description := "Test Product Pertama "

		req := &model.UpdateProductRequest{
			ID:          productID,
			UserID:      userID,
			Name:        "Product pertama",
			Description: &description,
			Price:       30000000,
			Quantity:    10,
		}
		mockValidator.EXPECT().ValidateUseCase(req).Return(nil)

		expectedProduct := &entity.Product{
			ID:     productID,
			UserID: userID,
			Name:   "product A",
			Slug:   gomock.Any().String(),
			Description: nullable.ToSQLString(
				&description,
			),
			Price:    30000,
			Quantity: 10,
		}

		mockProductRepo.EXPECT().FindByIDAndUserID(ctx, mockDB, req.ID, req.UserID).Return(expectedProduct, nil)
		mockProductRepo.EXPECT().ExistByNameOrSlugExceptHerself(ctx, mockDB, req.Name, gomock.Any(), req.ID).Return(true, nil)

		resp, err := uc.OwnerUpdate(ctx, req)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrAlreadyExists, err.(*helper.AppError).Code)
	})

	t.Run("failed to check exist by name or slug execpt herself", func(t *testing.T) {
		ctx := context.Background()
		userID := uuid.New()
		productID := uuid.New()

		description := "Test Product Pertama "

		req := &model.UpdateProductRequest{
			ID:          productID,
			UserID:      userID,
			Name:        "Product pertama",
			Description: &description,
			Price:       30000000,
			Quantity:    10,
		}
		mockValidator.EXPECT().ValidateUseCase(req).Return(nil)

		expectedProduct := &entity.Product{
			ID:     productID,
			UserID: userID,
			Name:   "product A",
			Slug:   gomock.Any().String(),
			Description: nullable.ToSQLString(
				&description,
			),
			Price:    30000,
			Quantity: 10,
		}

		mockProductRepo.EXPECT().FindByIDAndUserID(ctx, mockDB, req.ID, req.UserID).Return(expectedProduct, nil)
		mockProductRepo.EXPECT().ExistByNameOrSlugExceptHerself(ctx, mockDB, req.Name, gomock.Any(), req.ID).Return(false, errors.New("db down"))

		resp, err := uc.OwnerUpdate(ctx, req)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrInternal, err.(*helper.AppError).Code)
	})

	t.Run("failed to update product by ID", func(t *testing.T) {
		ctx := context.Background()
		userID := uuid.New()
		productID := uuid.New()

		description := "Test Product Pertama "

		req := &model.UpdateProductRequest{
			ID:          productID,
			UserID:      userID,
			Name:        "Product pertama",
			Description: &description,
			Price:       30000000,
			Quantity:    10,
		}
		mockValidator.EXPECT().ValidateUseCase(req).Return(nil)

		expectedProduct := &entity.Product{
			ID:     productID,
			UserID: userID,
			Name:   "product A",
			Slug:   gomock.Any().String(),
			Description: nullable.ToSQLString(
				&description,
			),
			Price:    30000,
			Quantity: 10,
		}

		mockProductRepo.EXPECT().FindByIDAndUserID(ctx, mockDB, req.ID, req.UserID).Return(expectedProduct, nil)
		mockProductRepo.EXPECT().ExistByNameOrSlugExceptHerself(ctx, mockDB, req.Name, gomock.Any(), req.ID).Return(false, nil)
		mockProductRepo.EXPECT().UpdateByID(ctx, mockDB, expectedProduct).Return(nil, errors.New("db downs"))

		resp, err := uc.OwnerUpdate(ctx, req)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrInternal, err.(*helper.AppError).Code)
	})

	t.Run("success update product", func(t *testing.T) {
		ctx := context.Background()
		userID := uuid.New()
		productID := uuid.New()

		description := "Test Product Pertama "

		req := &model.UpdateProductRequest{
			ID:          productID,
			UserID:      userID,
			Name:        "Product pertama",
			Description: &description,
			Price:       30000000,
			Quantity:    10,
		}
		mockValidator.EXPECT().ValidateUseCase(req).Return(nil)

		expectedProduct := &entity.Product{
			ID:     productID,
			UserID: userID,
			Name:   "product A",
			Slug:   gomock.Any().String(),
			Description: nullable.ToSQLString(
				&description,
			),
			Price:    30000,
			Quantity: 10,
		}

		mockProductRepo.EXPECT().FindByIDAndUserID(ctx, mockDB, req.ID, req.UserID).Return(expectedProduct, nil)
		mockProductRepo.EXPECT().ExistByNameOrSlugExceptHerself(ctx, mockDB, req.Name, gomock.Any(), req.ID).Return(false, nil)
		mockProductRepo.EXPECT().UpdateByID(ctx, mockDB, expectedProduct).Return(expectedProduct, nil)

		resp, err := uc.OwnerUpdate(ctx, req)
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, expectedProduct.Name, resp.Name)
		assert.Equal(t, expectedProduct.Description.String, resp.Description)
		assert.Equal(t, expectedProduct.Quantity, resp.Quantity)

	})
}

func TestProductUseCase_OwnerDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := mockstore.NewMockDatabaseStore(ctrl)
	mockProductRepo := mockrepository.NewMockProductRepository(ctrl)
	// mockUserAdapter := mockadapter.NewMockUserAdapter(ctrl)
	mockLogs := mocklogs.NewMockLog(ctrl)
	mockValidator := mockhelper.NewMockCustomValidator(ctrl)

	uc := usecase.NewProductUseCase(
		mockProductRepo,
		mockDB,
		mockValidator,
		mockLogs,
	)

	mockLogs.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockLogs.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	t.Run("invalid request", func(t *testing.T) {
		ctx := context.Background()
		req := &model.DeleteProductRequest{}

		expectedErr := &helper.UseCaseValError{}
		mockValidator.EXPECT().ValidateUseCase(req).Return(expectedErr)

		err := uc.OwnerDelete(ctx, req)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("product not found", func(t *testing.T) {
		ctx := context.Background()
		userID := uuid.New()
		productID := uuid.New()

		req := &model.DeleteProductRequest{
			ID:     productID,
			UserID: userID,
		}
		mockValidator.EXPECT().ValidateUseCase(req).Return(nil)
		mockProductRepo.EXPECT().FindByIDAndUserID(ctx, mockDB, req.ID, req.UserID).Return(nil, pgx.ErrNoRows)

		err := uc.OwnerDelete(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrResourceNotFound, err.(*helper.AppError).Code)
	})
	t.Run("failed to get product by id and user id", func(t *testing.T) {
		ctx := context.Background()
		userID := uuid.New()
		productID := uuid.New()

		req := &model.DeleteProductRequest{
			ID:     productID,
			UserID: userID,
		}
		mockValidator.EXPECT().ValidateUseCase(req).Return(nil)
		mockProductRepo.EXPECT().FindByIDAndUserID(ctx, mockDB, req.ID, req.UserID).Return(nil, errors.New("db down"))

		err := uc.OwnerDelete(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrInternal, err.(*helper.AppError).Code)
	})

	t.Run("product has already deleted", func(t *testing.T) {
		ctx := context.Background()
		userID := uuid.New()
		productID := uuid.New()

		description := "Test Product Pertama "

		req := &model.DeleteProductRequest{
			ID:     productID,
			UserID: userID,
		}
		mockValidator.EXPECT().ValidateUseCase(req).Return(nil)

		expectedProduct := &entity.Product{
			ID:     productID,
			UserID: userID,
			Name:   "product A",
			Slug:   gomock.Any().String(),
			Description: nullable.ToSQLString(
				&description,
			),
			Price:    30000,
			Quantity: 10,
		}

		mockProductRepo.EXPECT().FindByIDAndUserID(ctx, mockDB, req.ID, req.UserID).Return(expectedProduct, nil)
		mockProductRepo.EXPECT().DeleteByIDAndUserID(ctx, mockDB, req.ID, req.UserID).Return(errors.New(message.InternalNoRowsAffected))

		err := uc.OwnerDelete(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrResourceNotFound, err.(*helper.AppError).Code)
	})

	t.Run("failed to delete product by ID", func(t *testing.T) {
		ctx := context.Background()
		userID := uuid.New()
		productID := uuid.New()

		description := "Test Product Pertama "

		req := &model.DeleteProductRequest{
			ID:     productID,
			UserID: userID,
		}
		mockValidator.EXPECT().ValidateUseCase(req).Return(nil)

		expectedProduct := &entity.Product{
			ID:     productID,
			UserID: userID,
			Name:   "product A",
			Slug:   gomock.Any().String(),
			Description: nullable.ToSQLString(
				&description,
			),
			Price:    30000,
			Quantity: 10,
		}

		mockProductRepo.EXPECT().FindByIDAndUserID(ctx, mockDB, req.ID, req.UserID).Return(expectedProduct, nil)
		mockProductRepo.EXPECT().DeleteByIDAndUserID(ctx, mockDB, req.ID, req.UserID).Return(errors.New("db downs"))

		err := uc.OwnerDelete(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrInternal, err.(*helper.AppError).Code)
	})

	t.Run("success delete product", func(t *testing.T) {
		ctx := context.Background()
		userID := uuid.New()
		productID := uuid.New()

		description := "Test Product Pertama "

		req := &model.DeleteProductRequest{
			ID:     productID,
			UserID: userID,
		}
		mockValidator.EXPECT().ValidateUseCase(req).Return(nil)

		expectedProduct := &entity.Product{
			ID:     productID,
			UserID: userID,
			Name:   "product A",
			Slug:   gomock.Any().String(),
			Description: nullable.ToSQLString(
				&description,
			),
			Price:    30000,
			Quantity: 10,
		}

		mockProductRepo.EXPECT().FindByIDAndUserID(ctx, mockDB, req.ID, req.UserID).Return(expectedProduct, nil)
		mockProductRepo.EXPECT().DeleteByIDAndUserID(ctx, mockDB, req.ID, req.UserID).Return(nil)

		err := uc.OwnerDelete(ctx, req)
		assert.NoError(t, err)

	})

}

func TestProductUseCase_OwnerSearch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := mockstore.NewMockDatabaseStore(ctrl)
	mockProductRepo := mockrepository.NewMockProductRepository(ctrl)
	// mockUserAdapter := mockadapter.NewMockUserAdapter(ctrl)
	mockLogs := mocklogs.NewMockLog(ctrl)
	mockValidator := mockhelper.NewMockCustomValidator(ctrl)

	uc := usecase.NewProductUseCase(
		mockProductRepo,
		mockDB,
		mockValidator,
		mockLogs,
	)

	mockLogs.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockLogs.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	t.Run("internal db down", func(t *testing.T) {
		ctx := context.Background()
		req := &model.OwnerSearchProductsRequest{}

		mockProductRepo.EXPECT().OwnerFindAll(ctx, mockDB, req).Return(nil, nil, errors.New("db down"))

		product, metadata, err := uc.OwnerSearch(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrInternal, err.(*helper.AppError).Code)
		assert.Nil(t, product)
		assert.Nil(t, metadata)
	})

	t.Run("success but no product available", func(t *testing.T) {
		ctx := context.Background()
		req := &model.OwnerSearchProductsRequest{}
		metadata := &web.PageMetadata{}
		mockProductRepo.EXPECT().OwnerFindAll(ctx, mockDB, req).Return(nil, metadata, nil)

		product, respMetadata, err := uc.OwnerSearch(ctx, req)
		assert.Nil(t, err)
		assert.Equal(t, metadata, respMetadata)
		assert.Nil(t, product)
	})

	t.Run("success with available product", func(t *testing.T) {
		ctx := context.Background()
		req := &model.OwnerSearchProductsRequest{}
		metadata := &web.PageMetadata{}

		description := "desc"
		products := []*entity.ProductWithTotal{
			{
				ID:     uuid.New(),
				UserID: uuid.New(),
				Name:   "product A",
				Slug:   gomock.Any().String(),
				Description: nullable.ToSQLString(
					&description,
				),
				Price:    30000,
				Quantity: 10,
			},
		}

		mockProductRepo.EXPECT().OwnerFindAll(ctx, mockDB, req).Return(products, metadata, nil)

		respProduct, respMetadata, err := uc.OwnerSearch(ctx, req)
		assert.Nil(t, err)
		assert.Equal(t, metadata, respMetadata)
		assert.Equal(t, products[0].ID.String(), respProduct[0].ID)
		assert.Equal(t, products[0].Name, respProduct[0].Name)
	})
}

func TestProductUseCase_PublicSearch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := mockstore.NewMockDatabaseStore(ctrl)
	mockProductRepo := mockrepository.NewMockProductRepository(ctrl)
	// mockUserAdapter := mockadapter.NewMockUserAdapter(ctrl)
	mockLogs := mocklogs.NewMockLog(ctrl)
	mockValidator := mockhelper.NewMockCustomValidator(ctrl)

	uc := usecase.NewProductUseCase(
		mockProductRepo,
		mockDB,
		mockValidator,
		mockLogs,
	)

	mockLogs.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockLogs.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	t.Run("internal db down", func(t *testing.T) {
		ctx := context.Background()
		req := &model.PublicSearchProductsRequest{
			Page:  1,
			Limit: 1,
		}

		mockProductRepo.EXPECT().PublicFindAll(ctx, mockDB, req.Page, req.Limit).Return(nil, nil, errors.New("db down"))

		product, metadata, err := uc.PublicSearch(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrInternal, err.(*helper.AppError).Code)
		assert.Nil(t, product)
		assert.Nil(t, metadata)
	})

	t.Run("success but no product available", func(t *testing.T) {
		ctx := context.Background()
		req := &model.PublicSearchProductsRequest{
			Page:  1,
			Limit: 1,
		}
		metadata := &web.PageMetadata{}
		mockProductRepo.EXPECT().PublicFindAll(ctx, mockDB, req.Page, req.Limit).Return(nil, metadata, nil)

		product, respMetadata, err := uc.PublicSearch(ctx, req)
		assert.Nil(t, err)
		assert.Equal(t, metadata, respMetadata)
		assert.Nil(t, product)
	})

	t.Run("success with available product", func(t *testing.T) {
		ctx := context.Background()
		req := &model.PublicSearchProductsRequest{
			Page:  1,
			Limit: 1,
		}

		metadata := &web.PageMetadata{}

		description := "desc"
		products := []*entity.ProductWithTotal{
			{
				ID:     uuid.New(),
				UserID: uuid.New(),
				Name:   "product A",
				Slug:   gomock.Any().String(),
				Description: nullable.ToSQLString(
					&description,
				),
				Price:    30000,
				Quantity: 10,
			},
		}

		mockProductRepo.EXPECT().PublicFindAll(ctx, mockDB, req.Page, req.Limit).Return(products, metadata, nil)

		respProduct, respMetadata, err := uc.PublicSearch(ctx, req)
		assert.Nil(t, err)
		assert.Equal(t, metadata, respMetadata)
		assert.Equal(t, products[0].ID.String(), respProduct[0].ID)
		assert.Equal(t, products[0].Name, respProduct[0].Name)
	})
}
