package productservice

import (
	"context"

	"github.com/codepnw/go-starter-kit/internal/features/product"
	productrepository "github.com/codepnw/go-starter-kit/internal/features/product/repository"
)

type ProductService interface {
	CreateProduct(ctx context.Context, input *product.Product) error
	GetProduct(ctx context.Context, productID string) (*product.Product, error)
	GetProducts(ctx context.Context, limit, offset int) ([]*product.Product, error)
	UpdateProduct(ctx context.Context, input *product.Product) error
	DeleteProduct(ctx context.Context, productID string) error
}

type productService struct {
	repo productrepository.ProductRepository
}

func NewProductService(repo productrepository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) CreateProduct(ctx context.Context, input *product.Product) error {
	panic("unimplemented")
}

func (s *productService) GetProduct(ctx context.Context, productID string) (*product.Product, error) {
	panic("unimplemented")
}

func (s *productService) GetProducts(ctx context.Context, limit int, offset int) ([]*product.Product, error) {
	panic("unimplemented")
}

func (s *productService) UpdateProduct(ctx context.Context, input *product.Product) error {
	panic("unimplemented")
}

func (s *productService) DeleteProduct(ctx context.Context, productID string) error {
	panic("unimplemented")
}
