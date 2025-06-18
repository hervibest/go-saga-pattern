package repository

import (
	"context"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/commoner/store"
	"go-saga-pattern/user-svc/internal/entity"

	"github.com/georgysavva/scany/v2/pgxscan"
	"go.uber.org/zap"
)

type UserRepository interface {
	ExistsByUsername(ctx context.Context, db store.Querier, username string) (bool, error)
	FindAdminByUsername(ctx context.Context, db store.Querier, username string) (*entity.User, error)
	FindAdminByID(ctx context.Context, db store.Querier, id string) (*entity.User, error)
	FindEmployeeByUsername(ctx context.Context, db store.Querier, username string) (*entity.User, error)
	FindEmployeeByID(ctx context.Context, db store.Querier, id string) (*entity.User, error)
	FindManyEmployees(ctx context.Context, db store.Querier) (*[]*entity.User, error)
	Insert(ctx context.Context, db store.Querier, user *entity.User) (*entity.User, error)
}
type userRepositoryImpl struct {
	log logs.Log
}

func NewUserRepository(log logs.Log) UserRepository {
	return &userRepositoryImpl{log: log}
}

func (r *userRepositoryImpl) ExistsByUsername(ctx context.Context, db store.Querier, username string) (bool, error) {
	var total int
	query := `SELECT COUNT(*) FROM users WHERE username = $1`
	if err := pgxscan.Get(ctx, db, &total, query, username); err != nil {
		r.log.Error("failed to get query", zap.String("query", query), zap.Error(err))
		return false, err
	}

	if total > 0 {
		return true, nil
	}

	return false, nil
}

func (r *userRepositoryImpl) Insert(ctx context.Context, db store.Querier, user *entity.User) (*entity.User, error) {
	query := `
	INSERT INTO users 
	(username, password, role) 
	VALUES ($1, $2, $3) RETURNING id
	`
	err := pgxscan.Get(ctx, db, user, query, user.Username, user.Password, user.Role)
	if err != nil {
		r.log.Error("failed to exec insert query", zap.String("query", query), zap.Error(err))
		return nil, err
	}
	return user, nil
}

func (r *userRepositoryImpl) FindEmployeeByID(ctx context.Context, db store.Querier, id string) (*entity.User, error) {
	user := new(entity.User)
	query := `SELECT * FROM users WHERE id = $1 AND role = 'EMPLOYEE'`
	if err := pgxscan.Get(ctx, db, user, query, id); err != nil {
		r.log.Error("failed to get query", zap.String("query", query), zap.Error(err))
		return nil, err
	}
	return user, nil
}

func (r *userRepositoryImpl) FindEmployeeByUsername(ctx context.Context, db store.Querier, username string) (*entity.User, error) {
	user := new(entity.User)
	query := `SELECT * FROM users WHERE username = $1 AND role = 'EMPLOYEE'`
	if err := pgxscan.Get(ctx, db, user, query, username); err != nil {
		r.log.Error("failed to get query", zap.String("query", query), zap.Error(err))
		return nil, err
	}
	return user, nil
}

func (r *userRepositoryImpl) FindManyEmployees(ctx context.Context, db store.Querier) (*[]*entity.User, error) {
	users := make([]*entity.User, 0)
	query := `SELECT * FROM users WHERE role = 'EMPLOYEE'`
	if err := pgxscan.Select(ctx, db, &users, query); err != nil {
		r.log.Error("failed to get query", zap.String("query", query), zap.Error(err))
		return nil, err
	}
	return &users, nil
}

func (r *userRepositoryImpl) FindAdminByID(ctx context.Context, db store.Querier, id string) (*entity.User, error) {
	user := new(entity.User)
	query := `SELECT * FROM users WHERE id = $1 AND role = 'ADMIN'`
	if err := pgxscan.Get(ctx, db, user, query, id); err != nil {
		r.log.Error("failed to get query", zap.String("query", query), zap.Error(err))
		return nil, err
	}
	return user, nil
}

func (r *userRepositoryImpl) FindAdminByUsername(ctx context.Context, db store.Querier, username string) (*entity.User, error) {
	user := new(entity.User)
	query := `SELECT * FROM users WHERE username = $1 AND role = 'ADMIN'`
	if err := pgxscan.Get(ctx, db, user, query, username); err != nil {
		r.log.Error("failed to get query", zap.String("query", query), zap.Error(err))
		return nil, err
	}
	return user, nil
}
