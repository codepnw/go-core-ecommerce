package cartservice

import (
	"context"

	"github.com/codepnw/go-starter-kit/internal/config"
	"github.com/codepnw/go-starter-kit/internal/errs"
	"github.com/codepnw/go-starter-kit/internal/features/cart"
	cartrepository "github.com/codepnw/go-starter-kit/internal/features/cart/repository"
	productservice "github.com/codepnw/go-starter-kit/internal/features/product/service"
)

type CartService interface {
	AddItem(ctx context.Context, userID string, productID int64, quantity int) error
	GetCart(ctx context.Context, userID string) (*cart.CartResponse, error)
	RemoveItem(ctx context.Context, userID string, productID int64) error
}

type cartService struct {
	repo    cartrepository.CartRepository
	prodSrv productservice.ProductService
}

func NewCartService(repo cartrepository.CartRepository, prodSrv productservice.ProductService) CartService {
	return &cartService{
		repo:    repo,
		prodSrv: prodSrv,
	}
}

func (s *cartService) AddItem(ctx context.Context, userID string, productID int64, quantity int) error {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	// Check Product Stock
	prodData, err := s.prodSrv.GetProduct(ctx, productID)
	if err != nil {
		return err
	}
	if prodData.Stock < quantity {
		return errs.ErrStockNotEnough
	}

	// Get CartID
	cartID, err := s.repo.FindCartID(ctx, userID)
	if err != nil {
		return err
	}

	// Create Cart Item
	if err := s.repo.AddItem(ctx, cartID, productID, quantity); err != nil {
		return err
	}
	return nil
}

func (s *cartService) GetCart(ctx context.Context, userID string) (*cart.CartResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	// Get Cart Items
	items, err := s.repo.GetCartItems(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Prepare Response
	resp := &cart.CartResponse{
		Items:       make([]cart.CartItemData, 0),
		TotolAmount: 0,
		TotalQty:    0,
	}

	for _, item := range items {
		itemTotal := item.Price * item.Quantity

		isStockOK := item.Stock >= item.Quantity

		// Map Data
		data := cart.CartItemData{
			ID:          item.ID,
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
			Total:       itemTotal,
			IsStockOK:   isStockOK,
		}

		resp.Items = append(resp.Items, data)
		resp.TotolAmount += itemTotal
		resp.TotalQty += item.Quantity
	}
	return resp, nil
}

func (s *cartService) RemoveItem(ctx context.Context, userID string, productID int64) error {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()
	
	cartID, err := s.repo.FindCartID(ctx, userID)
	if err != nil {
		return err
	}
	
	return s.repo.RemoveItem(ctx, cartID, productID)
}