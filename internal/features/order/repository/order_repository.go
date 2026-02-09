package orderrepository

import (
	"context"
	"database/sql"

	"github.com/codepnw/go-starter-kit/internal/features/order"
)

//go:generate mockgen -source=order_repository.go -destination=order_repository_mock.go -package=orderrepository
type OrderRepository interface {
	// Transaction
	InsertOrderTx(ctx context.Context, tx *sql.Tx, userID string, totalAmount int64, address string) (int64, error)
	InsertOrderItemTx(ctx context.Context, tx *sql.Tx, item order.OrderItemReq) error
}

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) InsertOrderTx(ctx context.Context, tx *sql.Tx, userID string, totalAmount int64, address string) (int64, error) {
	var orderID int64
	query := `
		INSERT INTO orders (user_id, total_amount, status, address)
		VALUES ($1, $2, 'PENDING', $3) RETURNING id
	`
	err := tx.QueryRowContext(ctx, query, userID, totalAmount, address).Scan(&orderID)
	if err != nil {
		return 0, err
	}
	return orderID, nil
}

func (r *orderRepository) InsertOrderItemTx(ctx context.Context, tx *sql.Tx, item order.OrderItemReq) error {
	query := `
		INSERT INTO order_items (order_id, product_id, quantity, price)
		VALUES ($1, $2, $3, $4)
	`
	_, err := tx.ExecContext(ctx, query, item.OrderID, item.ProductID, item.Quantity, item.Price)
	if err != nil {
		return err
	}
	return nil
}
