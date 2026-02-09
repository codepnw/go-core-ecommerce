package productrepository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/codepnw/go-starter-kit/internal/errs"
	"github.com/codepnw/go-starter-kit/internal/features/product"
)

//go:generate mockgen -source=product_repository.go -destination=product_repository_mock.go -package=productrepository
type ProductRepository interface {
	InsertProduct(ctx context.Context, input *product.Product) error
	FindProduct(ctx context.Context, productID int64) (*product.Product, error)
	ListProducts(ctx context.Context, limit, offset int) ([]*product.Product, error)
	UpdateProduct(ctx context.Context, input *product.Product) error
	DeleteProduct(ctx context.Context, productID int64) error
	IncreaseStock(ctx context.Context, productID int64, qty int) error
	
	// Transaction
	DecreaseStockTx(ctx context.Context, tx *sql.Tx, productID int64, qty int) error
}

type productRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) InsertProduct(ctx context.Context, input *product.Product) error {
	query := `
		INSERT INTO products (name, price, stock, sku, version)
		VALUES ($1, $2, $3, $4, 1) RETURNING id, version
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		&input.Name,
		&input.Price,
		&input.Stock,
		&input.SKU,
	).Scan(
		&input.ID,
		&input.Version,
	)
	if err != nil {
		if strings.Contains(err.Error(), "products_sku_unique") {
			return errs.ErrProductSKUExists
		}
		return err
	}
	return nil
}

func (r *productRepository) FindProduct(ctx context.Context, productID int64) (*product.Product, error) {
	var p product.Product

	query := `
		SELECT id, name, price, stock, sku, version
		FROM products WHERE id = $1
	`
	err := r.db.QueryRowContext(ctx, query, productID).Scan(
		&p.ID,
		&p.Name,
		&p.Price,
		&p.Stock,
		&p.SKU,
		&p.Version,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrProductNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *productRepository) ListProducts(ctx context.Context, limit, offset int) ([]*product.Product, error) {
	query := `
		SELECT id, name, price, stock, sku, version
		FROM products LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}

	var products []*product.Product

	for rows.Next() {
		var p product.Product
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Price,
			&p.Stock,
			&p.SKU,
			&p.Version,
		); err != nil {
			return nil, err
		}
		products = append(products, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *productRepository) UpdateProduct(ctx context.Context, input *product.Product) error {
	query := `
		UPDATE products SET name = $1, price = $2, sku = $3
		WHERE id = $4
	`
	_, err := r.db.ExecContext(ctx, query, input.Name, input.Price, input.SKU, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return errs.ErrProductNotFound
		case strings.Contains(err.Error(), "products_sku_unique"):
			return errs.ErrProductSKUExists
		default:
			return err
		}
	}
	return nil
}

func (r *productRepository) DeleteProduct(ctx context.Context, productID int64) error {
	query := `DELETE FROM products WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, productID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errs.ErrProductNotFound
	}
	return nil
}

func (r *productRepository) IncreaseStock(ctx context.Context, productID int64, qty int) error {
	query := `
		UPDATE products SET stock = stock + $1, version = version + 1
		WHERE id = $2
	`
	res, err := r.db.ExecContext(ctx, query, qty, productID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errs.ErrProductNotFound
	}
	return nil
}

func (r *productRepository) DecreaseStockTx(ctx context.Context, tx *sql.Tx, productID int64, qty int) error {
	query := `
		UPDATE products SET stock = stock - $1, version = version + 1
		WHERE id = $2 AND stock >= $1
	`
	res, err := tx.ExecContext(ctx, query, qty, productID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errs.ErrStockNotEnough
	}
	return nil
}
