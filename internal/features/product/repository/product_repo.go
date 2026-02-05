package productrepository

import (
	"context"
	"database/sql"

	"github.com/codepnw/go-starter-kit/internal/features/product"
)

type ProductRepository interface {
	InsertProduct(ctx context.Context, input *product.Product) error
	FindProduct(ctx context.Context, productID string) (*product.Product, error)
	ListProducts(ctx context.Context, limit, offset int) ([]*product.Product, error)
	UpdateProduct(ctx context.Context, productID string, input *product.Product) error
	DeleteProduct(ctx context.Context, productID string) error
}

type productRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) InsertProduct(ctx context.Context, input *product.Product) error {
	panic("unimplemented")
}

func (r *productRepository) FindProduct(ctx context.Context, productID string) (*product.Product, error) {
	panic("unimplemented")
}

func (r *productRepository) ListProducts(ctx context.Context, limit, offset int) ([]*product.Product, error) {
	panic("unimplemented")
}

func (r *productRepository) UpdateProduct(ctx context.Context, productID string, input *product.Product) error {
	panic("unimplemented")
}

func (r *productRepository) DeleteProduct(ctx context.Context, productID string) error {
	panic("unimplemented")
}
