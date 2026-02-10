package orderservice

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/codepnw/go-starter-kit/internal/config"
	"github.com/codepnw/go-starter-kit/internal/errs"
	cartrepository "github.com/codepnw/go-starter-kit/internal/features/cart/repository"
	"github.com/codepnw/go-starter-kit/internal/features/order"
	orderrepository "github.com/codepnw/go-starter-kit/internal/features/order/repository"
	productrepository "github.com/codepnw/go-starter-kit/internal/features/product/repository"
	"github.com/codepnw/go-starter-kit/pkg/database"
)

type OrderService interface {
	CreateOrder(ctx context.Context, userID, address string) (string, error)
	GetOrderDetails(ctx context.Context, orderID int64) (*order.OrderDetailResponse, error)
}

type orderService struct {
	tx        database.TxManager
	orderRepo orderrepository.OrderRepository
	prodRepo  productrepository.ProductRepository
	cartRepo  cartrepository.CartRepository
}

func NewOrderService(
	tx database.TxManager,
	orderRepo orderrepository.OrderRepository,
	prodRepo productrepository.ProductRepository,
	cartRepo cartrepository.CartRepository,
) OrderService {
	return &orderService{
		tx:        tx,
		orderRepo: orderRepo,
		prodRepo:  prodRepo,
		cartRepo:  cartRepo,
	}
}

// GetOrderDetails implements OrderService.
func (s *orderService) GetOrderDetails(ctx context.Context, orderID int64) (*order.OrderDetailResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	ordData, err := s.orderRepo.FindOrderDetails(ctx, orderID)
	if err != nil {
		return nil, err
	}
	
	// Details Response
	resp := &order.OrderDetailResponse{
		OrderNo:   generateOrderNo(ordData.ID, ordData.CreatedAt),
		OrderDate: ordData.CreatedAt.Format(time.DateTime),
		Status:    ordData.Status,
		Address:   ordData.Address,
		Amount:    int64(ordData.TotalAmount),
		Items:     make([]order.OrderItemResponse, 0),
	}
	
	// Add Items Response
	for _, item := range ordData.Items {
		ordItem := order.OrderItemResponse{
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       int64(item.Price),
			Total:       int64(item.Price) * int64(item.Quantity),
		}
		resp.Items = append(resp.Items, ordItem)
	}
	return resp, nil
}

// CreateOrder implements OrderService.
func (s *orderService) CreateOrder(ctx context.Context, userID, address string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	// 1. Find Cart Items
	cartItems, err := s.cartRepo.GetCartItems(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("get cart items failed: %w", err)
	}
	if len(cartItems) == 0 {
		return "", errs.ErrCartEmpty
	}
	// Calculate Total Amount
	var totalAmount int64 = 0
	for _, item := range cartItems {
		totalAmount += int64(item.Price) * int64(item.Quantity)
	}

	var orderID int64
	var orderCreatedAt time.Time

	// Transaction
	err = s.tx.WithTx(ctx, func(tx *sql.Tx) error {
		// 2. Create Order
		id, createdAt, err := s.orderRepo.InsertOrderTx(ctx, tx, userID, totalAmount, address)
		if err != nil {
			return fmt.Errorf("insert order failed: %w", err)
		}
		orderID = id
		orderCreatedAt = createdAt

		// 3. Loop Items
		for _, item := range cartItems {
			// 3.1 Product Decrease Stock
			if err := s.prodRepo.DecreaseStockTx(ctx, tx, item.ProductID, item.Quantity); err != nil {
				return fmt.Errorf("product %s out of stock: %w", item.ProductName, err)
			}

			// 3.2 Create Order Items
			err := s.orderRepo.InsertOrderItemTx(ctx, tx, order.OrderItemReq{
				OrderID:   orderID,
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				Price:     item.Price, // Snapshot! current price
			})
			if err != nil {
				return fmt.Errorf("insert order items failed: %w", err)
			}
		}

		// 4. Clear Cart
		if err := s.cartRepo.ClearCartTx(ctx, tx, userID); err != nil {
			return fmt.Errorf("clear cart failed: %w", err)
		}

		return nil // Commit Transaction
	})
	if err != nil {
		return "", err
	}

	return generateOrderNo(orderID, orderCreatedAt), nil
}

// -------- HELPER ------------

func generateOrderNo(orderID int64, createdAt time.Time) string {
	now := createdAt.Format("20060201")
	return fmt.Sprintf("ORD-%s-%06d", now, orderID)
}
