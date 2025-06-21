package repository

import (
	"context"
	"errors"
	"go-saga-pattern/commoner/constant/enum"
	"go-saga-pattern/commoner/helper"
	"go-saga-pattern/commoner/web"
	"go-saga-pattern/product-svc/internal/entity"
	"go-saga-pattern/product-svc/internal/model"
	"go-saga-pattern/product-svc/internal/repository/store"
	"log"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ProductRepository interface {
	DeleteByIDAndUserID(ctx context.Context, db store.Querier, id uuid.UUID, userID uuid.UUID) error
	ExistByNameOrSlugExceptHerself(ctx context.Context, db store.Querier, name string, slug string, id uuid.UUID) (bool, error)
	ExistsByNameOrSlug(ctx context.Context, db store.Querier, name string, slug string) (bool, error)
	FindByID(ctx context.Context, db store.Querier, id uuid.UUID) (*entity.Product, error)
	FindByIDAndUserID(ctx context.Context, db store.Querier, id uuid.UUID, userID uuid.UUID) (*entity.Product, error)
	FindBySlug(ctx context.Context, db store.Querier, slug string) (*entity.Product, error)
	FindManyByIDs(ctx context.Context, db store.Querier, ids []uuid.UUID, lockType enum.LockTypeEnum) ([]*entity.Product, error)
	Insert(ctx context.Context, db store.Querier, product *entity.Product) (*entity.Product, error)
	OwnerFindAll(ctx context.Context, db store.Querier, request *model.OwnerSearchProductsRequest) ([]*entity.ProductWithTotal, *web.PageMetadata, error)
	PublicFindAll(ctx context.Context, db store.Querier, page int, limit int) ([]*entity.ProductWithTotal, *web.PageMetadata, error)
	UpdateByID(ctx context.Context, db store.Querier, product *entity.Product) (*entity.Product, error)
	// UpdateQuantityByID(ctx context.Context, db store.Querier, id uuid.UUID, quantity int) (*entity.Product, error)
	ReduceQuantity(ctx context.Context, db store.Querier, id uuid.UUID, quantity int) error
	RestoreQuantity(ctx context.Context, db store.Querier, id uuid.UUID, quantity int) error
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
		id, user_id, name, slug, description, price, quantity, created_at, updated_at, deleted_at
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

func (r *productRepository) FindManyByIDs(ctx context.Context, db store.Querier, ids []uuid.UUID, lockType enum.LockTypeEnum) ([]*entity.Product, error) {
	var products []*entity.Product
	query := `
	SELECT 
		id, name, slug, description, price, quantity, created_at, updated_at, deleted_at 
	FROM 
		products 
	WHERE 
		id = ANY($1) AND deleted_at IS NULL
	`
	if lockType == enum.LockTypeUpdateEnum {
		query += " FOR UPDATE"
	} else if lockType == enum.LockTypeShareEnum {
		query += " FOR SHARE"
	}

	if err := pgxscan.Select(ctx, db, &products, query, pq.Array(ids)); err != nil {
		return nil, err
	}

	return products, nil
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

func (r *productRepository) UpdateByID(ctx context.Context, db store.Querier, product *entity.Product) (*entity.Product, error) {
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
	log.Default().Printf("Update Product Query: %s with Product: %+v", query, product)
	if err := pgxscan.Get(ctx, db, product, query, product.Name, product.Slug, product.Description,
		product.Price, product.Quantity, product.ID, product.UserID); err != nil {
		return nil, err
	}
	return product, nil
}

func (r *productRepository) DeleteByIDAndUserID(ctx context.Context, db store.Querier, id, userID uuid.UUID) error {
	query := `
	UPDATE
		products
	SET
		deleted_at = NOW()
	WHERE
		id = $1 AND user_id = $2 AND deleted_at IS NULL
	`
	row, err := db.Exec(ctx, query, id, userID)
	if err != nil {
		return err
	}

	log.Printf("Delete Product Query: %s with ID: %s and UserID: %s", query, id, userID)

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

func (r *productRepository) OwnerFindAll(ctx context.Context, db store.Querier, request *model.OwnerSearchProductsRequest) ([]*entity.ProductWithTotal, *web.PageMetadata, error) {
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
	if err := pgxscan.Select(ctx, db, &products, query, request.UserID, request.Limit, (request.Page-1)*request.Limit); err != nil {
		return nil, &web.PageMetadata{}, err
	}

	if len(products) == 0 {
		return nil, nil, nil
	}

	totalItems = products[0].TotalData

	pageMetadata := helper.CalculatePagination(int64(totalItems), request.Page, request.Limit)
	return products, pageMetadata, nil
}

func (r *productRepository) ReduceQuantity(ctx context.Context, db store.Querier, id uuid.UUID, quantity int) error {
	query := "UPDATE products SET quantity = quantity - $1 WHERE id = $2"
	row, err := db.Exec(ctx, query, quantity, id)
	if err != nil {
		return err
	}
	if row.RowsAffected() == 0 {
		return errors.New("invalid product id")
	}

	return nil
}

func (r *productRepository) RestoreQuantity(ctx context.Context, db store.Querier, id uuid.UUID, quantity int) error {
	query := "UPDATE products SET quantity = quantity + $1 WHERE id = $2"
	row, err := db.Exec(ctx, query, quantity, id)
	if err != nil {
		return err
	}
	if row.RowsAffected() == 0 {
		return errors.New("invalid product id")
	}

	return nil
}
