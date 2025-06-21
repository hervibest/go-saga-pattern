package usecase

import (
	"context"
	"errors"
	"fmt"
	errorcode "go-saga-pattern/commoner/constant/errcode"
	"go-saga-pattern/commoner/constant/message"
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/user-svc/internal/adapter"
	"go-saga-pattern/user-svc/internal/entity"
	"go-saga-pattern/user-svc/internal/model"
	"go-saga-pattern/user-svc/internal/model/converter"
	"go-saga-pattern/user-svc/internal/repository"
	"go-saga-pattern/user-svc/internal/repository/store"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type UserUseCase interface {
	CurrentUser(ctx context.Context, email string) (*model.UserResponse, error)
	LoginUser(ctx context.Context, request *model.LoginUserRequest) (*model.UserResponse, *model.TokenResponse, error)
	RegisterUser(ctx context.Context, request *model.RegisterUserRequest) (*model.UserResponse, error)
	VerifyUser(ctx context.Context, token string) (*model.AuthResponse, error)
	Logout(ctx context.Context, request *model.LogoutUserRequest) error
}

type userUseCase struct {
	db              store.DB
	userRepository  repository.UserRepository
	jwtAdapter      adapter.JWTAdapter
	cacheAdapter    adapter.CacheAdapter
	customValidator helper.CustomValidator
	logs            logs.Log
}

func NewUserUseCase(db store.DB, userRepository repository.UserRepository, jwtAdapter adapter.JWTAdapter,
	cacheAdapter adapter.CacheAdapter, customValidator helper.CustomValidator, logs logs.Log) UserUseCase {
	return &userUseCase{
		db:              db,
		userRepository:  userRepository,
		jwtAdapter:      jwtAdapter,
		cacheAdapter:    cacheAdapter,
		customValidator: customValidator,
		logs:            logs,
	}
}

func (uc *userUseCase) RegisterUser(ctx context.Context, request *model.RegisterUserRequest) (*model.UserResponse, error) {
	if validatonErrs := uc.customValidator.ValidateUseCase(request); validatonErrs != nil {
		return nil, validatonErrs
	}

	isExists, err := uc.userRepository.ExistsByUsernameOrEmail(ctx, uc.db, request.Username, request.Email)
	if err != nil {
		return nil, helper.WrapInternalServerError(uc.logs, "failed to check user exist in register user", err)
	}

	if isExists {
		return nil, helper.NewUseCaseError(errorcode.ErrInvalidArgument, message.ClientUserAlreadyExist)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, helper.WrapInternalServerError(uc.logs, "failed to generate hashed password bcrypt", err)
	}

	user := &entity.User{
		Username: request.Username,
		Email:    request.Email,
		Password: string(hashedPassword),
	}

	createdUser, err := uc.userRepository.Insert(ctx, uc.db, user)
	if err != nil {
		return nil, helper.WrapInternalServerError(uc.logs, "failed to insert new user", err)
	}

	return converter.UserToResponse(createdUser), nil
}

func (uc *userUseCase) LoginUser(ctx context.Context, request *model.LoginUserRequest) (*model.UserResponse, *model.TokenResponse, error) {
	if validatonErrs := uc.customValidator.ValidateUseCase(request); validatonErrs != nil {
		return nil, nil, validatonErrs
	}

	user, err := uc.userRepository.FindByEmail(ctx, uc.db, request.Email)
	if err != nil {
		if strings.Contains(err.Error(), pgx.ErrNoRows.Error()) {
			return nil, nil, helper.NewUseCaseError(errorcode.ErrInvalidArgument, message.ClientInvalidEmailOrPassword)
		}
		return nil, nil, helper.WrapInternalServerError(uc.logs, "failed to find user by username", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)); err != nil {
		return nil, nil, helper.NewUseCaseError(errorcode.ErrInvalidArgument, message.ClientInvalidEmailOrPassword)
	}

	accessTokenDetail, err := uc.jwtAdapter.GenerateUserAccessToken(user.ID)
	if err != nil {
		return nil, nil, helper.WrapInternalServerError(uc.logs, "failed to generate access token", err)
	}

	auth := &entity.Auth{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}

	if err := uc.saveUserToCache(ctx, auth, accessTokenDetail.ExpiresAt); err != nil {
		return nil, nil, helper.WrapInternalServerError(uc.logs, "failed to save user cache", err)
	}

	token := &model.TokenResponse{
		AccessToken: accessTokenDetail.Token,
	}

	return converter.UserToResponse(user), token, nil
}

func (uc *userUseCase) saveUserToCache(ctx context.Context, auth *entity.Auth, expiresAt time.Time) error {
	jsonValue, err := sonic.ConfigFastest.Marshal(auth)
	if err != nil {
		return fmt.Errorf("marshal user : %+v", err)
	}

	if err := uc.cacheAdapter.Set(ctx, auth.ID.String(), jsonValue, time.Until(expiresAt)); err != nil {
		return fmt.Errorf("save user body into cache : %+v", err)
	}

	return nil
}

func (uc *userUseCase) CurrentUser(ctx context.Context, email string) (*model.UserResponse, error) {
	admin, err := uc.userRepository.FindByEmail(ctx, uc.db, email)
	if err != nil {
		if strings.Contains(err.Error(), pgx.ErrNoRows.Error()) {
			return nil, helper.NewUseCaseError(errorcode.ErrUserNotFound, "Invalid token")
		}
		return nil, helper.WrapInternalServerError(uc.logs, "failed to find by email not google", err)
	}

	return converter.UserToResponse(admin), nil
}

func (uc *userUseCase) VerifyUser(ctx context.Context, token string) (*model.AuthResponse, error) {
	accessTokenDetail, err := uc.jwtAdapter.VerifyUserAccessToken(token)
	if err != nil {
		return nil, helper.NewUseCaseError(errorcode.ErrUnauthorized, "Invalid access token")
	}

	userId, err := uc.cacheAdapter.Get(ctx, token)
	if userId != "" {
		return nil, helper.NewUseCaseError(errorcode.ErrUnauthorized, "You have already signed out")
	}

	cachedUserStr, err := uc.cacheAdapter.Get(ctx, accessTokenDetail.UserID)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, helper.WrapInternalServerError(uc.logs, "failed to get cached user", err)
	}

	auth := new(entity.Auth)
	//If redis stale, get from db
	if errors.Is(err, redis.Nil) {
		user, err := uc.userRepository.FindByID(ctx, uc.db, accessTokenDetail.UserID)
		if err != nil {
			if strings.Contains(err.Error(), pgx.ErrNoRows.Error()) {
				uc.logs.Debug("Err no rows for user id in user repo find by id (access token user id)")
				return nil, helper.NewUseCaseError(errorcode.ErrUnauthorized, "invalid access token")
			}
			return nil, helper.WrapInternalServerError(uc.logs, "failed to find user by id", err)
		}

		auth.ID = user.ID
		auth.Username = user.Username
		auth.Email = user.Email

		if err = uc.saveUserToCache(ctx, auth, accessTokenDetail.ExpiresAt); err != nil {
			return nil, err
		}
	} else {
		if err := sonic.ConfigFastest.Unmarshal([]byte(cachedUserStr), &auth); err != nil {
			return nil, helper.WrapInternalServerError(uc.logs, "failed to unmarshal user body from cached", err)
		}
	}

	authResponse := &model.AuthResponse{
		ID:        auth.ID.String(),
		Username:  auth.Username,
		Email:     auth.Email,
		Token:     accessTokenDetail.Token,
		ExpiresAt: accessTokenDetail.ExpiresAt,
	}

	return authResponse, nil
}

func (u *userUseCase) Logout(ctx context.Context, request *model.LogoutUserRequest) error {
	if err := u.cacheAdapter.Set(ctx, request.AccessToken, "revoked", time.Until(request.ExpiresAt)); err != nil {
		return helper.WrapInternalServerError(u.logs, "failed to save access token to cache for logout : ", err)
	}
	return nil
}
