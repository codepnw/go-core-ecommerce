package productservice

import (
	"context"

	"github.com/codepnw/go-starter-kit/internal/config"
	"github.com/codepnw/go-starter-kit/internal/features/product"
	productrepository "github.com/codepnw/go-starter-kit/internal/features/product/repository"
)

//go:generate mockgen -source=product_service.go -destination=product_service_mock.go -package=productservice
type ProductService interface {
	CreateProduct(ctx context.Context, input *product.Product) error
	GetProduct(ctx context.Context, productID int64) (*product.Product, error)
	GetProducts(ctx context.Context, limit, offset int) ([]*product.Product, error)
	IncreaseStock(ctx context.Context, productID int64, qty int) error
	UpdateProduct(ctx context.Context, input UpdateProductInput) error
	DeleteProduct(ctx context.Context, productID int64) error
}

type productService struct {
	repo productrepository.ProductRepository
}

func NewProductService(repo productrepository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) CreateProduct(ctx context.Context, input *product.Product) error {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	if err := s.repo.InsertProduct(ctx, input); err != nil {
		return err
	}
	return nil
}

func (s *productService) GetProduct(ctx context.Context, productID int64) (*product.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	productData, err := s.repo.FindProduct(ctx, productID)
	if err != nil {
		return nil, err
	}
	return productData, nil
}

func (s *productService) GetProducts(ctx context.Context, limit int, offset int) ([]*product.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()
	
	if limit == 0 {
		limit = 10
	}

	products, err := s.repo.ListProducts(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (s *productService) IncreaseStock(ctx context.Context, productID int64, qty int) error {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()
	
	if err := s.repo.IncreaseStock(ctx, productID, qty); err != nil {
		return err
	}
	return nil
}

type UpdateProductInput struct {
	ID    int64
	Name  *string
	Price *int
	SKU   *string
}

func (s *productService) UpdateProduct(ctx context.Context, input UpdateProductInput) error {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()
	
	exists, err := s.repo.FindProduct(ctx, input.ID)
	if err != nil {
		return err
	}
	
	if input.Name != nil {
		exists.Name = *input.Name
	}
	if input.Price != nil {
		exists.Price = *input.Price
	}
	if input.SKU != nil {
		exists.SKU = *input.SKU
	}

	if err := s.repo.UpdateProduct(ctx, exists); err != nil {
		return err
	}
	return nil
}

func (s *productService) DeleteProduct(ctx context.Context, productID int64) error {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	if err := s.repo.DeleteProduct(ctx, productID); err != nil {
		return err
	}
	return nil
}
