package repository

import (
	"context"
	"errors"
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/store"
	"go-saga-pattern/commoner/web"
	"go-saga-pattern/product-svc/internal/entity"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
)

type ProductRepository interface {
	DeleteByIDAndUserID(ctx context.Context, db store.Querier, id uuid.UUID, userId uuid.UUID) error
	ExistByNameOrSlugExceptHerself(ctx context.Context, db store.Querier, name string, slug string, id uuid.UUID) (bool, error)
	ExistsByNameOrSlug(ctx context.Context, db store.Querier, name string, slug string) (bool, error)
	FindByID(ctx context.Context, db store.Querier, id uuid.UUID) (*entity.Product, error)
	FindByIDAndUserID(ctx context.Context, db store.Querier, id uuid.UUID, userID uuid.UUID) (*entity.Product, error)
	FindBySlug(ctx context.Context, db store.Querier, slug string) (*entity.Product, error)
	Insert(ctx context.Context, db store.Querier, product *entity.Product) (*entity.Product, error)
	OwnerFindAll(ctx context.Context, db store.Querier, userID uuid.UUID, page int, limit int) ([]*entity.ProductWithTotal, *web.PageMetadata, error)
	PublicFindAll(ctx context.Context, db store.Querier, page int, limit int) ([]*entity.ProductWithTotal, *web.PageMetadata, error)
	UpdateById(ctx context.Context, db store.Querier, product *entity.Product) (*entity.Product, error)
}

type productRepository struct{}

func NewProductRepository() ProductRepository {
	return &productRepository{}
}

func (r *productRepository) Insert(ctx context.Context, db store.Querier, product *entity.Product) (*entity.Product, error) {
	query := `
	INSERT INTO products
		(user_id, name, slug, description, price, quantity)
	VALUES
		($1, $2, $3, $4, $5, $6)
	RETURNING
		id, created_at, updated_at
	`
	if err := pgxscan.Get(ctx, db, product, query, product.UserID,
		product.Name, product.Slug, product.Description, product.Price, product.Quantity); err != nil {
		return nil, err
	}
	return product, nil
}

func (r *productRepository) FindByIDAndUserID(ctx context.Context, db store.Querier, id, userID uuid.UUID) (*entity.Product, error) {
	query := `
	SELECT
		id, name, slug, description, price, quantity, created_at, updated_at, deleted_at
	FROM
		products
	WHERE
		id = $1 AND user_id = $2 AND deleted_at IS NULL
	`
	product := new(entity.Product)
	if err := pgxscan.Get(ctx, db, product, query, id, userID); err != nil {
		return nil, err
	}
	return product, nil
}

func (r *productRepository) FindByID(ctx context.Context, db store.Querier, id uuid.UUID) (*entity.Product, error) {
	query := `
	SELECT
		id, name, slug, description, price, quantity, created_at, updated_at, deleted_at
	FROM
		products
	WHERE
		id = $1 AND deleted_at IS NULL
	`
	product := new(entity.Product)
	if err := pgxscan.Get(ctx, db, product, query, id); err != nil {
		return nil, err
	}
	return product, nil
}

func (r *productRepository) FindBySlug(ctx context.Context, db store.Querier, slug string) (*entity.Product, error) {
	query := `
	SELECT
		id, name, slug, description, price, quantity, created_at, updated_at, deleted_at
	FROM
		products
	WHERE
		slug = $1 AND deleted_at IS NULL
	`
	product := new(entity.Product)
	if err := pgxscan.Get(ctx, db, product, query, slug); err != nil {
		return nil, err
	}
	return product, nil
}

func (r *productRepository) ExistsByNameOrSlug(ctx context.Context, db store.Querier, name, slug string) (bool, error) {
	query := `
	SELECT
		COUNT(*)
	FROM
		products
	WHERE
		(name = $1 OR slug = $2) AND deleted_at IS NULL
	`
	var count int64
	if err := db.QueryRow(ctx, query, name, slug).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *productRepository) ExistByNameOrSlugExceptHerself(ctx context.Context, db store.Querier, name, slug string, id uuid.UUID) (bool, error) {
	query := `
	SELECT
		COUNT(*)
	FROM
		products
	WHERE
		(name = $1 OR slug = $2) AND deleted_at IS NULL AND id != $3
	`
	var count int64
	if err := db.QueryRow(ctx, query, name, slug, id).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *productRepository) UpdateById(ctx context.Context, db store.Querier, product *entity.Product) (*entity.Product, error) {
	query := `
	UPDATE
		products
	SET
		name = COALESCE($1, name),
		slug = COALESCE($2, slug),
		description = COALESCE($3, description),
		price = COALESCE($4, price),
		quantity = COALESCE($5, quantity),
		updated_at = NOW()
	WHERE
		id = $6 AND user_id = $7 AND deleted_at IS NULL
	RETURNING
		created_at, updated_at
	`
	if err := pgxscan.Get(ctx, db, product, query, product.Name, product.Slug, product.Description,
		product.Price, product.Quantity, product.Id, product.UserID); err != nil {
		return nil, err
	}
	return product, nil
}

func (r *productRepository) DeleteByIDAndUserID(ctx context.Context, db store.Querier, id, userId uuid.UUID) error {
	query := `
	UPDATE
		products
	SET
		deleted_at = NOW()
	WHERE
		id = $1 AND user_id = $2 AND deleted_at IS NULL
	`
	row, err := db.Exec(ctx, query, id, userId)
	if err != nil {
		return err
	}

	if row.RowsAffected() == 0 {
		return errors.New("product not found or already deleted")
	}

	if row.RowsAffected() > 1 {
		return errors.New("multiple products deleted, which is unexpected")
	}

	return nil

}

func (r *productRepository) PublicFindAll(ctx context.Context, db store.Querier, page,
	limit int) ([]*entity.ProductWithTotal, *web.PageMetadata, error) {
	var totalItems int
	query := `
	SELECT
		COUNT(*) OVER () AS total_data,
		id, name, slug, description, price, quantity, created_at, updated_at, deleted_at
	FROM
		products
	WHERE
		deleted_at IS NULL
	ORDER BY
		created_at DESC
	LIMIT $1 OFFSET $2
	`

	var products []*entity.ProductWithTotal
	if err := pgxscan.Select(ctx, db, &products, query, limit, (page-1)*limit); err != nil {
		return nil, nil, err
	}

	if len(products) == 0 {
		return nil, nil, nil
	}

	totalItems = products[0].TotalData

	pageMetadata := helper.CalculatePagination(int64(totalItems), page, limit)
	return products, pageMetadata, nil
}

func (r *productRepository) OwnerFindAll(ctx context.Context, db store.Querier, userID uuid.UUID, page,
	limit int) ([]*entity.ProductWithTotal, *web.PageMetadata, error) {
	var totalItems int
	query := `
	SELECT
		COUNT(*) OVER () AS total_data,
		id, name, slug, description, price, quantity, created_at, updated_at, deleted_at
	FROM
		products
	WHERE
		user_id = $1 AND deleted_at IS NULL
	ORDER BY
		created_at DESC
	LIMIT $2 OFFSET $3
	`

	var products []*entity.ProductWithTotal
	if err := pgxscan.Select(ctx, db, &products, query, userID, limit, (page-1)*limit); err != nil {
		return nil, &web.PageMetadata{}, err
	}

	if len(products) == 0 {
		return nil, nil, nil
	}

	totalItems = products[0].TotalData

	pageMetadata := helper.CalculatePagination(int64(totalItems), page, limit)
	return products, pageMetadata, nil
}
