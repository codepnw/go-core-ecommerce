package orderservice

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/codepnw/go-starter-kit/internal/config"
	"github.com/codepnw/go-starter-kit/internal/errs"
	"github.com/codepnw/go-starter-kit/internal/features/cart"
	cartrepository "github.com/codepnw/go-starter-kit/internal/features/cart/repository"
	"github.com/codepnw/go-starter-kit/internal/features/order"
	orderrepository "github.com/codepnw/go-starter-kit/internal/features/order/repository"
	productrepository "github.com/codepnw/go-starter-kit/internal/features/product/repository"
	"github.com/codepnw/go-starter-kit/pkg/database"
)

type OrderService interface {
	CreateOrder(ctx context.Context, userID, address string) (string, error)
	GetOrderDetails(ctx context.Context, orderID int64) (*order.OrderDetailResponse, error)
	MyOrders(ctx context.Context, userID string, page, limit int) (*order.OrderListResponse, error)
	UpdateStatus(ctx context.Context, orderID int64, status order.OrderStatus) error
}

type orderService struct {
	tx        database.TxManager
	orderRepo orderrepository.OrderRepository
	prodRepo  productrepository.ProductRepository
	prodRedis productrepository.ProductRedisRepository
	cartRepo  cartrepository.CartRepository
}

type OrderServiceDeps struct {
	Tx        database.TxManager
	OrderRepo orderrepository.OrderRepository
	ProdRepo  productrepository.ProductRepository
	ProdRedis productrepository.ProductRedisRepository
	CartRepo  cartrepository.CartRepository
}

func NewOrderService(deps *OrderServiceDeps) OrderService {
	return &orderService{
		tx:        deps.Tx,
		orderRepo: deps.OrderRepo,
		prodRepo:  deps.ProdRepo,
		prodRedis: deps.ProdRedis,
		cartRepo:  deps.CartRepo,
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

	// Get Cart Items
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

	// ============================
	// Redis Gatekeeper (Fast-Fail)
	decreasedItems := make([]*cart.CartItemResult, 0, len(cartItems))
	for _, item := range cartItems {
		// Check Product Promotion
		isPromo, _ := s.prodRedis.CheckStockExists(ctx, item.ProductID)
		
		if isPromo {
			// Redis Decrease Product Stock
			_, err := s.prodRedis.DecreaseStock(ctx, item.ProductID, item.Quantity)
			if err != nil {
				// Rollback: Redis Increase Product Stock
				for _, dItem := range decreasedItems {
					_ = s.prodRedis.IncreaseStock(ctx, dItem.ProductID, dItem.Quantity)
				}
				return "", errs.ErrStockNotEnough
			}
			decreasedItems = append(decreasedItems, item)
		}
	}

	// Sort ID protect Deadlock!
	sort.Slice(cartItems, func(i, j int) bool {
		return cartItems[i].ProductID < cartItems[j].ProductID
	})

	var orderID int64
	var orderCreatedAt time.Time

	// Transaction
	err = s.tx.WithTx(ctx, func(tx *sql.Tx) error {
		// Create Order
		id, createdAt, err := s.orderRepo.InsertOrderTx(ctx, tx, userID, totalAmount, address)
		if err != nil {
			return fmt.Errorf("insert order failed: %w", err)
		}
		orderID = id
		orderCreatedAt = createdAt

		for _, item := range cartItems {
			// Decrease Product Stock
			if err := s.prodRepo.DecreaseStockTx(ctx, tx, item.ProductID, item.Quantity); err != nil {
				return fmt.Errorf("product %s out of stock: %w", item.ProductName, err)
			}

			// Create Order Items
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

		// Clear Cart
		if err := s.cartRepo.ClearCartTx(ctx, tx, userID); err != nil {
			return fmt.Errorf("clear cart failed: %w", err)
		}
		return nil // Commit Transaction
	})
	if err != nil {
		// Transaction Failed: Rollback Redis Product Stock
		for _, item := range cartItems {
			_ = s.prodRedis.IncreaseStock(ctx, item.ProductID, item.Quantity)
		}
		return "", err
	}

	return generateOrderNo(orderID, orderCreatedAt), nil
}

// MyOrders implements OrderService.
func (s *orderService) MyOrders(ctx context.Context, userID string, page, limit int) (*order.OrderListResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Find My Orders
	orders, total, err := s.orderRepo.FindMyOrders(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	totalPage := int(math.Ceil(float64(total) / float64(limit)))

	// Response
	resp := &order.OrderListResponse{
		Orders:      make([]*order.OrderResponse, 0, len(orders)),
		TotalOrders: total,
		Page:        page,
		Limit:       limit,
		TotalPage:   totalPage,
		HasNextPage: page < totalPage,
		HasPrevPage: page > 1,
	}

	for _, item := range orders {
		o := &order.OrderResponse{
			OrderNo:     generateOrderNo(item.ID, item.CreatedAt),
			TotalAmount: int64(item.TotalAmount),
			Status:      item.Status,
			CreatedAt:   item.CreatedAt.Format(time.DateTime),
		}
		resp.Orders = append(resp.Orders, o)
	}
	return resp, nil
}

// UpdateStatus implements OrderService.
func (s *orderService) UpdateStatus(ctx context.Context, orderID int64, newStatus order.OrderStatus) error {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	// Order Details
	orderData, err := s.orderRepo.FindOrderDetails(ctx, orderID)
	if err != nil {
		return err
	}
	currentStatus := orderData.Status

	// Check Status Transition
	if !isValidStatus(currentStatus, newStatus) {
		return errs.ErrInvalidStatusTransition
	}

	// Cancel Order Method
	if newStatus == order.StatusCancelled {
		return s.cancelOrder(ctx, orderID)
	}

	return s.tx.WithTx(ctx, func(tx *sql.Tx) error {
		// Update New Status
		if err := s.orderRepo.UpdateStatusTx(ctx, tx, orderID, newStatus); err != nil {
			return err
		}
		return nil
	})
}

// -------- Private Method ------------

func (s *orderService) cancelOrder(ctx context.Context, orderID int64) error {
	var orderItems []*order.OrderItem

	err := s.tx.WithTx(ctx, func(tx *sql.Tx) error {
		items, err := s.orderRepo.FindOrderItemsTx(ctx, tx, orderID)
		if err != nil {
			return err
		}
		orderItems = items

		for _, item := range items {
			if err := s.prodRepo.IncreaseStockTx(ctx, tx, item.ProductID, item.Quantity); err != nil {
				return err
			}
		}

		if err := s.orderRepo.UpdateStatusTx(ctx, tx, orderID, order.StatusCancelled); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Transaction Success: Redis Rollback Product Stock
	for _, item := range orderItems {
		_ = s.prodRedis.IncreaseStock(ctx, item.ProductID, item.Quantity)
	}
	return nil
}

// -------- HELPER Function ------------

func isValidStatus(oldStatus, newStatus order.OrderStatus) bool {
	// Pending 		-> (Paid, Cancelled)
	// Paid 		-> (Shipped, Cancelled)
	// Shipped 		-> Completed
	// Completed  	-> End Process
	// Cancelled	-> End Process
	switch oldStatus {
	case order.StatusPending:
		return newStatus == order.StatusPaid || newStatus == order.StatusCancelled
	case order.StatusPaid:
		return newStatus == order.StatusShipped || newStatus == order.StatusCancelled
	case order.StatusShipped:
		return newStatus == order.StatusCompleted
	case order.StatusCompleted, order.StatusCancelled:
		return false
	default:
		return false
	}
}

func generateOrderNo(orderID int64, createdAt time.Time) string {
	now := createdAt.Format("20060201")
	return fmt.Sprintf("ORD-%s-%06d", now, orderID)
}
