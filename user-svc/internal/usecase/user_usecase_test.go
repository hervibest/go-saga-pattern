package usecase_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	errorcode "go-saga-pattern/commoner/constant/errcode"
	"go-saga-pattern/commoner/helper"
	mockhelper "go-saga-pattern/commoner/mocks/commoner/helper"
	mocklogs "go-saga-pattern/commoner/mocks/commoner/logs"
	"go-saga-pattern/user-svc/internal/entity"
	mockadapter "go-saga-pattern/user-svc/internal/mocks/adapter"
	mockrepository "go-saga-pattern/user-svc/internal/mocks/repository"
	mockstore "go-saga-pattern/user-svc/internal/mocks/store"
	"go-saga-pattern/user-svc/internal/model"
	"go-saga-pattern/user-svc/internal/repository/store"
	"go-saga-pattern/user-svc/internal/usecase"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestUserUseCase_RegisterUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := mockstore.NewMockDB(ctrl)
	mockUserRepo := mockrepository.NewMockUserRepository(ctrl)
	mockJWT := mockadapter.NewMockJWTAdapter(ctrl)
	mockCache := mockadapter.NewMockCacheAdapter(ctrl)
	mockValidator := mockhelper.NewMockCustomValidator(ctrl)
	mockLogs := mocklogs.NewMockLog(ctrl)

	uc := usecase.NewUserUseCase(
		mockDB,
		mockUserRepo,
		mockJWT,
		mockCache,
		mockValidator,
		mockLogs,
	)

	mockLogs.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockLogs.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	t.Run("successful registration", func(t *testing.T) {
		ctx := context.Background()
		req := &model.RegisterUserRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		// Mock validation
		mockValidator.EXPECT().ValidateUseCase(req).Return(nil)

		// Mock existence check
		mockUserRepo.EXPECT().ExistsByUsernameOrEmail(ctx, mockDB, req.Username, req.Email).Return(false, nil)

		// Mock user creation
		expectedUser := &entity.User{
			Username: req.Username,
			Email:    req.Email,
			Password: "hashedpassword",
		}
		mockUserRepo.EXPECT().Insert(ctx, mockDB, gomock.Any()).DoAndReturn(
			func(ctx context.Context, db store.DB, user *entity.User) (*entity.User, error) {
				// Verify password was hashed
				err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
				assert.NoError(t, err)
				return expectedUser, nil
			},
		)

		resp, err := uc.RegisterUser(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser.Username, resp.Username)
		assert.Equal(t, expectedUser.Email, resp.Email)
	})

	t.Run("validation error", func(t *testing.T) {
		ctx := context.Background()
		req := &model.RegisterUserRequest{}

		expectedErr := &helper.UseCaseValError{}
		mockValidator.EXPECT().ValidateUseCase(req).Return(expectedErr)

		resp, err := uc.RegisterUser(ctx, req)
		assert.Nil(t, resp)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("user already exists", func(t *testing.T) {
		ctx := context.Background()
		req := &model.RegisterUserRequest{
			Username: "existinguser",
			Email:    "existing@example.com",
			Password: "password123",
		}

		mockValidator.EXPECT().ValidateUseCase(req).Return(nil)
		mockUserRepo.EXPECT().ExistsByUsernameOrEmail(ctx, mockDB, req.Username, req.Email).Return(true, nil)

		resp, err := uc.RegisterUser(ctx, req)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrInvalidArgument, err.(*helper.AppError).Code)
	})

	t.Run("database error on existence check", func(t *testing.T) {
		ctx := context.Background()
		req := &model.RegisterUserRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		mockValidator.EXPECT().ValidateUseCase(req).Return(nil)
		mockUserRepo.EXPECT().ExistsByUsernameOrEmail(ctx, mockDB, req.Username, req.Email).Return(false, errors.New("db error"))
		mockLogs.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

		resp, err := uc.RegisterUser(ctx, req)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrInternal, err.(*helper.AppError).Code)
	})

	t.Run("password hashing error", func(t *testing.T) {
		ctx := context.Background()
		req := &model.RegisterUserRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: string(make([]byte, 73)), // bcrypt max length is 72 bytes
		}

		mockValidator.EXPECT().ValidateUseCase(req).Return(nil)
		mockUserRepo.EXPECT().ExistsByUsernameOrEmail(ctx, mockDB, req.Username, req.Email).Return(false, nil)
		mockLogs.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())

		resp, err := uc.RegisterUser(ctx, req)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrInternal, err.(*helper.AppError).Code)
	})

	t.Run("database error on insert", func(t *testing.T) {
		ctx := context.Background()
		req := &model.RegisterUserRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		mockValidator.EXPECT().ValidateUseCase(req).Return(nil)
		mockUserRepo.EXPECT().ExistsByUsernameOrEmail(ctx, mockDB, req.Username, req.Email).Return(false, nil)
		mockUserRepo.EXPECT().Insert(ctx, mockDB, gomock.Any()).Return(nil, errors.New("db error"))
		mockLogs.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())

		resp, err := uc.RegisterUser(ctx, req)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrInternal, err.(*helper.AppError).Code)
	})
}

func TestUserUseCase_LoginUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := mockstore.NewMockDB(ctrl)
	mockUserRepo := mockrepository.NewMockUserRepository(ctrl)
	mockJWT := mockadapter.NewMockJWTAdapter(ctrl)
	mockCache := mockadapter.NewMockCacheAdapter(ctrl)
	mockValidator := mockhelper.NewMockCustomValidator(ctrl)
	mockLogs := mocklogs.NewMockLog(ctrl)

	mockLogs.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockLogs.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	uc := usecase.NewUserUseCase(
		mockDB,
		mockUserRepo,
		mockJWT,
		mockCache,
		mockValidator,
		mockLogs,
	)

	t.Run("successful login", func(t *testing.T) {
		ctx := context.Background()
		req := &model.LoginUserRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		const testUserID = "46e3081a-6405-4434-b34a-34dc0fab5e8d" // Fixed test UUID
		user := &entity.User{
			ID:       uuid.MustParse(testUserID),
			Username: "testuser",
			Email:    req.Email,
			Password: string(hashedPassword),
		}

		assert.Equal(t, testUserID, user.ID.String())

		// Mock validation
		mockValidator.EXPECT().ValidateUseCase(req).Return(nil)

		// Mock user lookup
		mockUserRepo.EXPECT().FindByEmail(ctx, mockDB, req.Email).Return(user, nil)

		// Mock token generation
		tokenDetail := &entity.UserAccessToken{
			Token:     "access_token",
			ExpiresAt: time.Now().Add(time.Hour),
		}

		assert.Equal(t, testUserID, user.ID.String())

		mockJWT.EXPECT().GenerateUserAccessToken(user.ID).Return(tokenDetail, nil)

		// Mock cache save
		expectedAuth := &entity.Auth{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		}
		fmt.Println("testUserID           :", testUserID)
		fmt.Println("user.ID              :", user.ID.String())

		jsonData, _ := json.Marshal(expectedAuth)

		assert.Equal(t, testUserID, expectedAuth.ID.String())

		mockCache.EXPECT().Set(
			gomock.Any(),
			user.ID.String(),
			jsonData,
			gomock.Any(),
		).Return(nil)

		userResp, tokenResp, err := uc.LoginUser(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, user.Username, userResp.Username)
		assert.Equal(t, tokenDetail.Token, tokenResp.AccessToken)
	})

	t.Run("validation error", func(t *testing.T) {
		ctx := context.Background()
		req := &model.LoginUserRequest{}

		expectedErr := &helper.UseCaseValError{}
		mockValidator.EXPECT().ValidateUseCase(req).Return(expectedErr)

		userResp, tokenResp, err := uc.LoginUser(ctx, req)
		assert.Nil(t, userResp)
		assert.Nil(t, tokenResp)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("user not found", func(t *testing.T) {
		ctx := context.Background()
		req := &model.LoginUserRequest{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}

		mockValidator.EXPECT().ValidateUseCase(req).Return(nil)
		mockUserRepo.EXPECT().FindByEmail(ctx, mockDB, req.Email).Return(nil, pgx.ErrNoRows)

		userResp, tokenResp, err := uc.LoginUser(ctx, req)
		assert.Nil(t, userResp)
		assert.Nil(t, tokenResp)
		assert.Equal(t, errorcode.ErrInvalidArgument, err.(*helper.AppError).Code)
	})

	t.Run("database error on find user", func(t *testing.T) {
		ctx := context.Background()
		req := &model.LoginUserRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		mockValidator.EXPECT().ValidateUseCase(req).Return(nil)
		mockUserRepo.EXPECT().FindByEmail(ctx, mockDB, req.Email).Return(nil, errors.New("db error"))

		userResp, tokenResp, err := uc.LoginUser(ctx, req)
		assert.Nil(t, userResp)
		assert.Nil(t, tokenResp)
		assert.Equal(t, errorcode.ErrInternal, err.(*helper.AppError).Code)
	})

	t.Run("invalid password", func(t *testing.T) {
		ctx := context.Background()
		req := &model.LoginUserRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
		user := &entity.User{
			ID:       uuid.New(),
			Username: "testuser",
			Email:    req.Email,
			Password: string(hashedPassword),
		}

		mockValidator.EXPECT().ValidateUseCase(req).Return(nil)
		mockUserRepo.EXPECT().FindByEmail(ctx, mockDB, req.Email).Return(user, nil)

		userResp, tokenResp, err := uc.LoginUser(ctx, req)
		assert.Nil(t, userResp)
		assert.Nil(t, tokenResp)
		assert.Equal(t, errorcode.ErrInvalidArgument, err.(*helper.AppError).Code)
	})

	t.Run("token generation error", func(t *testing.T) {
		ctx := context.Background()
		req := &model.LoginUserRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		user := &entity.User{
			ID:       uuid.New(),
			Username: "testuser",
			Email:    req.Email,
			Password: string(hashedPassword),
		}

		mockValidator.EXPECT().ValidateUseCase(req).Return(nil)
		mockUserRepo.EXPECT().FindByEmail(ctx, mockDB, req.Email).Return(user, nil)
		mockJWT.EXPECT().GenerateUserAccessToken(user.ID).Return(nil, errors.New("token error"))

		userResp, tokenResp, err := uc.LoginUser(ctx, req)
		assert.Nil(t, userResp)
		assert.Nil(t, tokenResp)
		assert.Equal(t, errorcode.ErrInternal, err.(*helper.AppError).Code)
	})

	t.Run("cache save error", func(t *testing.T) {
		ctx := context.Background()
		req := &model.LoginUserRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		user := &entity.User{
			ID:       uuid.New(),
			Username: "testuser",
			Email:    req.Email,
			Password: string(hashedPassword),
		}

		tokenDetail := &entity.UserAccessToken{
			Token:     "access_token",
			ExpiresAt: time.Now().Add(time.Hour),
		}

		mockValidator.EXPECT().ValidateUseCase(req).Return(nil)
		mockUserRepo.EXPECT().FindByEmail(ctx, mockDB, req.Email).Return(user, nil)
		mockJWT.EXPECT().GenerateUserAccessToken(user.ID).Return(tokenDetail, nil)
		mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("cache error"))

		userResp, tokenResp, err := uc.LoginUser(ctx, req)
		assert.Nil(t, userResp)
		assert.Nil(t, tokenResp)
		assert.Equal(t, errorcode.ErrInternal, err.(*helper.AppError).Code)
	})
}

func TestUserUseCase_CurrentUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := mockstore.NewMockDB(ctrl)
	mockUserRepo := mockrepository.NewMockUserRepository(ctrl)
	mockJWT := mockadapter.NewMockJWTAdapter(ctrl)
	mockCache := mockadapter.NewMockCacheAdapter(ctrl)
	mockValidator := mockhelper.NewMockCustomValidator(ctrl)
	mockLogs := mocklogs.NewMockLog(ctrl)

	mockLogs.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockLogs.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	uc := usecase.NewUserUseCase(
		mockDB,
		mockUserRepo,
		mockJWT,
		mockCache,
		mockValidator,
		mockLogs,
	)

	t.Run("failed to find user by email", func(t *testing.T) {
		ctx := context.Background()
		email := "testuser@gmail.com"

		const testUserID = "46e3081a-6405-4434-b34a-34dc0fab5e8d" // Fixed test UUID
		user := &entity.User{
			ID:       uuid.MustParse(testUserID),
			Username: "testuser",
			Email:    email,
		}

		assert.Equal(t, testUserID, user.ID.String())

		// Mock validation
		// Mock user lookup
		mockUserRepo.EXPECT().FindByEmail(ctx, mockDB, email).Return(nil, pgx.ErrNoRows)

		_, err := uc.CurrentUser(ctx, email)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrUserNotFound, err.(*helper.AppError).Code)
	})

	t.Run("internal failed to find user by email", func(t *testing.T) {
		ctx := context.Background()
		email := "testuser@gmail.com"

		const testUserID = "46e3081a-6405-4434-b34a-34dc0fab5e8d" // Fixed test UUID
		user := &entity.User{
			ID:       uuid.MustParse(testUserID),
			Username: "testuser",
			Email:    email,
		}

		assert.Equal(t, testUserID, user.ID.String())

		// Mock validation
		// Mock user lookup
		mockUserRepo.EXPECT().FindByEmail(ctx, mockDB, email).Return(nil, errors.New("internal error"))

		_, err := uc.CurrentUser(ctx, email)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrInternal, err.(*helper.AppError).Code)
	})

	t.Run("successful get current user", func(t *testing.T) {
		ctx := context.Background()
		email := "testuser@gmail.com"

		const testUserID = "46e3081a-6405-4434-b34a-34dc0fab5e8d" // Fixed test UUID
		user := &entity.User{
			ID:       uuid.MustParse(testUserID),
			Username: "testuser",
			Email:    email,
		}

		assert.Equal(t, testUserID, user.ID.String())

		// Mock validation
		// Mock user lookup
		mockUserRepo.EXPECT().FindByEmail(ctx, mockDB, email).Return(user, nil)

		userResp, err := uc.CurrentUser(ctx, email)
		assert.NoError(t, err)
		assert.Equal(t, user.Username, userResp.Username)
	})
}

func TestUserUseCase_VerifyUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := mockstore.NewMockDB(ctrl)
	mockUserRepo := mockrepository.NewMockUserRepository(ctrl)
	mockJWT := mockadapter.NewMockJWTAdapter(ctrl)
	mockCache := mockadapter.NewMockCacheAdapter(ctrl)
	mockValidator := mockhelper.NewMockCustomValidator(ctrl)
	mockLogs := mocklogs.NewMockLog(ctrl)

	mockLogs.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockLogs.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogs.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	uc := usecase.NewUserUseCase(
		mockDB,
		mockUserRepo,
		mockJWT,
		mockCache,
		mockValidator,
		mockLogs,
	)

	t.Run("invalid access token", func(t *testing.T) {
		ctx := context.Background()
		token := "invalid_token"
		mockJWT.EXPECT().VerifyUserAccessToken(token).Return(nil, errors.New("invalid token"))

		_, err := uc.VerifyUser(ctx, token)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrUnauthorized, err.(*helper.AppError).Code)
	})

	t.Run("get cache internal error", func(t *testing.T) {
		ctx := context.Background()
		token := "valid_token"

		accessTokenDetail := &entity.UserAccessToken{
			Token:     token,
			ExpiresAt: time.Now().Add(time.Hour),
		}

		mockJWT.EXPECT().VerifyUserAccessToken(token).Return(accessTokenDetail, nil)

		mockCache.EXPECT().Get(ctx, token).Return("", errors.New("cache error"))

		_, err := uc.VerifyUser(ctx, token)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrInternal, err.(*helper.AppError).Code)
	})

	t.Run("user already signed out", func(t *testing.T) {
		ctx := context.Background()
		token := "valid_token"
		const userId = "46e3081a-6405-4434-b34a-34dc0fab5e8d" // Fixed test UUID

		accessTokenDetail := &entity.UserAccessToken{
			UserID:    userId,
			Token:     token,
			ExpiresAt: time.Now().Add(time.Hour),
		}

		mockJWT.EXPECT().VerifyUserAccessToken(token).Return(accessTokenDetail, nil)
		mockCache.EXPECT().Get(ctx, token).Return(userId, nil)

		_, err := uc.VerifyUser(ctx, token)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrUnauthorized, err.(*helper.AppError).Code)
	})

	t.Run("failed to get cached user data", func(t *testing.T) {
		ctx := context.Background()
		token := "valid_token"
		const userId = "46e3081a-6405-4434-b34a-34dc0fab5e8d" // Fixed test UUID

		accessTokenDetail := &entity.UserAccessToken{
			UserID:    userId,
			Token:     token,
			ExpiresAt: time.Now().Add(time.Hour),
		}

		mockJWT.EXPECT().VerifyUserAccessToken(token).Return(accessTokenDetail, nil)
		mockCache.EXPECT().Get(ctx, token).Return("", nil)
		mockCache.EXPECT().Get(ctx, accessTokenDetail.UserID).Return("", errors.New("cache error"))

		_, err := uc.VerifyUser(ctx, token)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrInternal, err.(*helper.AppError).Code)
	})

	t.Run("invalid user id when cache is nil", func(t *testing.T) {
		ctx := context.Background()
		token := "valid_token"
		const userId = "46e3081a-6405-4434-b34a-34dc0fab5e8d" // Fixed test UUID

		accessTokenDetail := &entity.UserAccessToken{
			UserID:    userId,
			Token:     token,
			ExpiresAt: time.Now().Add(time.Hour),
		}

		mockJWT.EXPECT().VerifyUserAccessToken(token).Return(accessTokenDetail, nil)
		mockCache.EXPECT().Get(ctx, token).Return("", nil)
		mockCache.EXPECT().Get(ctx, accessTokenDetail.UserID).Return("", redis.Nil)
		mockUserRepo.EXPECT().FindByID(ctx, mockDB, accessTokenDetail.UserID).Return(nil, pgx.ErrNoRows)

		_, err := uc.VerifyUser(ctx, token)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrUnauthorized, err.(*helper.AppError).Code)
	})

	t.Run("internal find by id error", func(t *testing.T) {
		ctx := context.Background()
		token := "valid_token"
		const userId = "46e3081a-6405-4434-b34a-34dc0fab5e8d" // Fixed test UUID

		accessTokenDetail := &entity.UserAccessToken{
			UserID:    userId,
			Token:     token,
			ExpiresAt: time.Now().Add(time.Hour),
		}

		mockJWT.EXPECT().VerifyUserAccessToken(token).Return(accessTokenDetail, nil)
		mockCache.EXPECT().Get(ctx, token).Return("", nil)
		mockCache.EXPECT().Get(ctx, accessTokenDetail.UserID).Return("", redis.Nil)
		mockUserRepo.EXPECT().FindByID(ctx, mockDB, accessTokenDetail.UserID).Return(nil, errors.New("internal error"))

		_, err := uc.VerifyUser(ctx, token)
		assert.Error(t, err)
		assert.Equal(t, errorcode.ErrInternal, err.(*helper.AppError).Code)
	})

	t.Run("success fallback to db when redis nil", func(t *testing.T) {
		ctx := context.Background()
		token := "valid_token"
		const userId = "46e3081a-6405-4434-b34a-34dc0fab5e8d" // Fixed test UUID

		accessTokenDetail := &entity.UserAccessToken{
			UserID:    userId,
			Token:     token,
			ExpiresAt: time.Now().Add(time.Hour),
		}

		mockJWT.EXPECT().VerifyUserAccessToken(token).Return(accessTokenDetail, nil)
		mockCache.EXPECT().Get(ctx, token).Return("", nil)
		mockCache.EXPECT().Get(ctx, accessTokenDetail.UserID).Return("", redis.Nil)
		user := &entity.User{
			ID:       uuid.MustParse(userId),
			Username: "testuser",
			Email:    "testemail@.com",
		}

		mockUserRepo.EXPECT().FindByID(ctx, mockDB, accessTokenDetail.UserID).Return(user, nil)
		expectedAuth := &entity.Auth{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		}

		jsonData, _ := json.Marshal(expectedAuth)

		assert.Equal(t, userId, expectedAuth.ID.String())

		mockCache.EXPECT().Set(
			gomock.Any(),
			user.ID.String(),
			jsonData,
			gomock.Any(),
		).Return(nil)

		res, err := uc.VerifyUser(ctx, token)

		assert.NoError(t, err)
		assert.Equal(t, user.Username, res.Username)
		assert.Equal(t, user.Email, res.Email)
		assert.Equal(t, user.ID.String(), res.ID)
	})

	//TODO success get from cache

}
