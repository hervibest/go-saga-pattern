package usecase_test

import (
	"context"
	errorcode "go-saga-pattern/commoner/constant/errcode"
	"go-saga-pattern/commoner/helper"
	mockhelper "go-saga-pattern/commoner/mocks/commoner/helper"
	mocklogs "go-saga-pattern/commoner/mocks/commoner/logs"
	mockrepository "go-saga-pattern/product-svc/internal/mocks/repository"
	mockstore "go-saga-pattern/product-svc/internal/mocks/store"
	"go-saga-pattern/product-svc/internal/model"
	"go-saga-pattern/product-svc/internal/usecase"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestProductTransactionUseCase_CheckProductsAndReserve(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := mockstore.NewMockDatabaseStore(ctrl)
	mockProductRepo := mockrepository.NewMockProductRepository(ctrl)
	mockProductTransRepo := mockrepository.NewMockProductTransactionRepository(ctrl)
	// mockUserAdapter := mockadapter.NewMockUserAdapter(ctrl)
	mockLogs := mocklogs.NewMockLog(ctrl)
	mockValidator := mockhelper.NewMockCustomValidator(ctrl)

	uc := usecase.NewProductTransactionUseCase(
		mockProductRepo,
		mockProductTransRepo,
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
		req := &model.CheckProductsQuantityRequest{}

		resp, err := uc.CheckProductsAndReserve(ctx, req)
		assert.Nil(t, resp)
		assert.Equal(t, errorcode.ErrInvalidArgument, err.(*helper.AppError).Code)
		tx := mockstore.NewMockTransaction(ctrl)
		mockDB.EXPECT().Begin(ctx).Return(tx, nil)
	})
}
