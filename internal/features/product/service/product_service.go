package productservice

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/codepnw/go-starter-kit/internal/config"
	"github.com/codepnw/go-starter-kit/internal/features/product"
	productrepository "github.com/codepnw/go-starter-kit/internal/features/product/repository"
	"github.com/redis/go-redis/v9"
)

//go:generate mockgen -source=product_service.go -destination=product_service_mock.go -package=productservice
type ProductService interface {
	CreateProduct(ctx context.Context, input *product.Product) error
	GetProduct(ctx context.Context, productID int64) (*product.Product, error)
	GetProducts(ctx context.Context, page, limit int) ([]*product.Product, error)
	IncreaseStock(ctx context.Context, productID int64, qty int) error
	UpdateProduct(ctx context.Context, input UpdateProductInput) error
	DeleteProduct(ctx context.Context, productID int64) error
}

type productService struct {
	repo    productrepository.ProductRepository
	redisDB *redis.Client
}

func NewProductService(repo productrepository.ProductRepository, redisDB *redis.Client) ProductService {
	return &productService{
		repo:    repo,
		redisDB: redisDB,
	}
}

func (s *productService) CreateProduct(ctx context.Context, input *product.Product) error {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()
	
	// Get Product Repository
	if err := s.repo.InsertProduct(ctx, input); err != nil {
		return err
	}

	// Set Redis DB
	if data, err := json.Marshal(input); err == nil {
		s.redisDB.Set(ctx, generateRedisKey(input.ID), data, config.RedisProductDuration)
	}

	return nil
}

func (s *productService) GetProduct(ctx context.Context, productID int64) (*product.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	redisKey := generateRedisKey(productID)
	// Get Redis DB
	val, err := s.redisDB.Get(ctx, redisKey).Result()
	if err == nil {
		p := new(product.Product)
		if err := json.Unmarshal([]byte(val), p); err == nil {
			// Return Found Product
			return p, nil
		}
	}

	// Get Product Repository
	productData, err := s.repo.FindProduct(ctx, productID)
	if err != nil {
		return nil, err
	}

	// Set Redis DB
	if data, err := json.Marshal(productData); err == nil {
		s.redisDB.Set(ctx, redisKey, data, config.RedisProductDuration)
	}

	return productData, nil
}

func (s *productService) GetProducts(ctx context.Context, page int, limit int) ([]*product.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	// Get Redis DB
	redisKey := fmt.Sprintf("products:page:%d:limit:%d", page, limit)
	val, err := s.redisDB.Get(ctx, redisKey).Result()
	if err == nil {
		var products []*product.Product
		if err := json.Unmarshal([]byte(val), &products); err == nil {
			return products, nil
		}
	}

	if limit == 0 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Get Product Repository
	products, err := s.repo.ListProducts(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	// Set Redis DB
	if data, err := json.Marshal(products); err == nil {
		s.redisDB.Set(ctx, redisKey, data, 1*time.Minute)
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

	// Delete Redis DB
	_ = s.redisDB.Del(ctx, generateRedisKey(input.ID))
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

// ----------- HELPER ---------------

func generateRedisKey(productID int64) string {
	return fmt.Sprintf("product:%d", productID)
}
