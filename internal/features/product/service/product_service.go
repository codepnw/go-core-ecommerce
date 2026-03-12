package productservice

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/codepnw/go-starter-kit/internal/config"
	"github.com/codepnw/go-starter-kit/internal/errs"
	"github.com/codepnw/go-starter-kit/internal/features/product"
	productrepository "github.com/codepnw/go-starter-kit/internal/features/product/repository"
	"github.com/codepnw/go-starter-kit/internal/helper"
	"github.com/codepnw/go-starter-kit/pkg/database"
)

//go:generate mockgen -source=product_service.go -destination=product_service_mock.go -package=productservice
type ProductService interface {
	CreateProduct(ctx context.Context, input *product.Product) error
	CreateProductPromotion(ctx context.Context, productID int64, stock int, discountPercent int) error
	GetProduct(ctx context.Context, productID int64) (*product.Product, error)
	GetProducts(ctx context.Context, page, limit int) ([]*product.Product, error)
	IncreaseStock(ctx context.Context, productID int64, qty int) error
	UpdateProduct(ctx context.Context, input UpdateProductInput) error
	DeleteProduct(ctx context.Context, productID int64) error
}

type productService struct {
	db      database.DBTX
	tx      database.TxManager
	repo    productrepository.ProductRepository
	redisDB productrepository.ProductRedisRepository
}

type ProductServiceDeps struct {
	DB        database.DBTX
	Tx        database.TxManager
	ProdRepo  productrepository.ProductRepository
	ProdRedis productrepository.ProductRedisRepository
}

func NewProductService(deps *ProductServiceDeps) ProductService {
	return &productService{
		db:      deps.DB,
		tx:      deps.Tx,
		repo:    deps.ProdRepo,
		redisDB: deps.ProdRedis,
	}
}

func (s *productService) CreateProduct(ctx context.Context, input *product.Product) error {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	// Insert Product Repository
	if err := s.repo.InsertProduct(ctx, s.db, input); err != nil {
		return err
	}
	return nil
}

func (s *productService) CreateProductPromotion(ctx context.Context, productID int64, stock int, discountPercent int) error {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	// Get Product
	prodData, err := s.repo.FindProduct(ctx, productID)
	if err != nil {
		return fmt.Errorf("find product failed: %w", err)
	}

	if prodData.Stock < stock {
		return errs.ErrStockNotEnough
	}

	var promoProductID int64
	// Transaction
	err = s.tx.WithTx(ctx, func(tx *sql.Tx) error {
		// Decrease Product stock
		if err := s.repo.DecreaseStockTx(ctx, tx, prodData.ID, stock); err != nil {
			return fmt.Errorf("decrease stock failed: %w", err)
		}

		// Insert Product Promotion
		prodInput := &product.Product{
			Name:  prodData.Name + "(Flash Sale)",
			Price: prodData.Price * (100 - discountPercent) / 100,
			Stock: stock,
			SKU:   prodData.SKU + "-PROMO",
		}
		if err := s.repo.InsertProduct(ctx, tx, prodInput); err != nil {
			return fmt.Errorf("insert product (flash sale) failed: %w", err)
		}

		promoProductID = prodInput.ID
		return nil
	})
	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	// Set Redis
	if err := s.redisDB.SetStock(ctx, promoProductID, stock); err != nil {
		return fmt.Errorf("redis set stock failed: %w", err)
	}
	return nil
}

func (s *productService) GetProduct(ctx context.Context, productID int64) (*product.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	productCache := new(product.Product)
	redisKey := helper.RedisProductItemKey(productID)

	// Get Product Cache
	err := s.redisDB.GetJSONCache(ctx, redisKey, productCache)
	if err == nil {
		return productCache, nil
	}

	// Get Product Repository
	productData, err := s.repo.FindProduct(ctx, productID)
	if err != nil {
		return nil, err
	}

	// Set Product Cache
	_ = s.redisDB.SetJSONCache(ctx, redisKey, productData, time.Minute*2)

	return productData, nil
}

func (s *productService) GetProducts(ctx context.Context, page int, limit int) ([]*product.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 10
	}
	offset := (page - 1) * limit

	productsCache := make([]*product.Product, 0, limit)
	redisKey := helper.RedisProductListKey(page, limit)

	// Get Products Cache
	err := s.redisDB.GetJSONCache(ctx, redisKey, &productsCache)
	if err == nil {
		return productsCache, nil
	}

	// Get Products Repository
	products, err := s.repo.ListProducts(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	// Set Products Cache
	_ = s.redisDB.SetJSONCache(ctx, redisKey, products, time.Minute*2)

	return products, nil
}

func (s *productService) IncreaseStock(ctx context.Context, productID int64, qty int) error {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	if err := s.repo.IncreaseStock(ctx, productID, qty); err != nil {
		return err
	}

	// Delete Product Cache
	redisKey := helper.RedisProductItemKey(productID)
	_ = s.redisDB.DeleteCache(ctx, redisKey)
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

	// Update Product Repository
	if err := s.repo.UpdateProduct(ctx, exists); err != nil {
		return err
	}

	// Delete Product Cache
	redisKey := helper.RedisProductItemKey(exists.ID)
	_ = s.redisDB.DeleteCache(ctx, redisKey)
	return nil
}

func (s *productService) DeleteProduct(ctx context.Context, productID int64) error {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	// Delete Product Repository
	if err := s.repo.DeleteProduct(ctx, productID); err != nil {
		return err
	}

	// Delete Product Cache
	redisKey := helper.RedisProductItemKey(productID)
	_ = s.redisDB.DeleteCache(ctx, redisKey)
	return nil
}
