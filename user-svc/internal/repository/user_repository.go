package repository

import (
	"context"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/user-svc/internal/entity"
	"go-saga-pattern/user-svc/internal/repository/store"

	"github.com/georgysavva/scany/v2/pgxscan"
	"go.uber.org/zap"
)

type UserRepository interface {
	DeleteByEmail(ctx context.Context, db store.Querier, email string) error
	DeleteByID(ctx context.Context, db store.Querier, id string) error
	ExistsByUsernameOrEmail(ctx context.Context, db store.Querier, username string, email string) (bool, error)
	FindByEmail(ctx context.Context, db store.Querier, email string) (*entity.User, error)
	FindByID(ctx context.Context, db store.Querier, id string) (*entity.User, error)
	Insert(ctx context.Context, db store.Querier, user *entity.User) (*entity.User, error)
}
type userRepositoryImpl struct {
	log logs.Log
}

func NewUserRepository(log logs.Log) UserRepository {
	return &userRepositoryImpl{log: log}
}

func (r *userRepositoryImpl) Insert(ctx context.Context, db store.Querier, user *entity.User) (*entity.User, error) {
	query := `
	INSERT INTO users
		(username, email, password)
	VALUES
		($1, $2, $3)
	RETURNING 
		id, created_at, updated_at
	`

	if err := pgxscan.Get(ctx, db, user, query, user.Username, user.Email, user.Password); err != nil {
		r.log.Error("failed to exec insert query", zap.String("query", query), zap.Error(err))
		return nil, err
	}
	return user, nil
}

func (r *userRepositoryImpl) FindByID(ctx context.Context, db store.Querier, id string) (*entity.User, error) {
	user := new(entity.User)
	query := `SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL`
	if err := pgxscan.Get(ctx, db, user, query, id); err != nil {
		r.log.Error("failed to get query", zap.String("query", query), zap.Error(err))
		return nil, err
	}
	return user, nil
}

func (r *userRepositoryImpl) FindByEmail(ctx context.Context, db store.Querier, email string) (*entity.User, error) {
	user := new(entity.User)
	query := `SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL`
	if err := pgxscan.Get(ctx, db, user, query, email); err != nil {
		r.log.Error("failed to get query", zap.String("query", query), zap.Error(err))
		return nil, err
	}
	return user, nil
}

func (r *userRepositoryImpl) DeleteByID(ctx context.Context, db store.Querier, id string) error {
	query := `UPDATE users SET deleted_at = now() WHERE id = $1 AND deleted_at IS NOT NULL`
	_, err := db.Exec(ctx, query, id)
	if err != nil {
		r.log.Error("failed to exec delete query", zap.String("query", query), zap.Error(err))
		return err
	}
	return nil
}

func (r *userRepositoryImpl) DeleteByEmail(ctx context.Context, db store.Querier, email string) error {
	query := `UPDATE users SET deleted_at = now() WHERE email = $1 AND deleted_at IS NOT NULL`
	_, err := db.Exec(ctx, query, email)
	if err != nil {
		r.log.Error("failed to exec delete query", zap.String("query", query), zap.Error(err))
		return err

	}
	return nil
}

func (r *userRepositoryImpl) ExistsByUsernameOrEmail(ctx context.Context, db store.Querier, username, email string) (bool, error) {
	var total int
	query := `SELECT COUNT(*) FROM users WHERE username = $1 OR email = $2 AND deleted_at IS NULL`
	if err := pgxscan.Get(ctx, db, &total, query, username, email); err != nil {
		r.log.Error("failed to get query", zap.String("query", query), zap.Error(err))
		return false, err
	}

	if total > 0 {
		return true, nil
	}

	return false, nil
}
