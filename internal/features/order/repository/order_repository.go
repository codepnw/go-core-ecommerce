package orderrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/codepnw/go-starter-kit/internal/errs"
	"github.com/codepnw/go-starter-kit/internal/features/order"
)

//go:generate mockgen -source=order_repository.go -destination=order_repository_mock.go -package=orderrepository
type OrderRepository interface {
	FindOrderDetails(ctx context.Context, orderID int64) (*order.Order, error)
	FindMyOrders(ctx context.Context, userID string, limit, offset int) ([]*order.Order, int64, error)

	// Transaction
	InsertOrderTx(ctx context.Context, tx *sql.Tx, userID string, totalAmount int64, address string) (int64, time.Time, error)
	InsertOrderItemTx(ctx context.Context, tx *sql.Tx, item order.OrderItemReq) error
	FindOrderItemsTx(ctx context.Context, tx *sql.Tx, orderID int64) ([]*order.OrderItem, error)
	UpdateStatusTx(ctx context.Context, tx *sql.Tx, orderID int64, status order.OrderStatus) error
}

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) FindOrderDetails(ctx context.Context, orderID int64) (*order.Order, error) {
	// Find orders table
	queryOrder := `
		SELECT id, user_id, address, total_amount, status, created_at, updated_at
		FROM orders WHERE id = $1
	`
	ord := new(order.Order)

	err := r.db.QueryRowContext(ctx, queryOrder, orderID).Scan(
		&ord.ID,
		&ord.UserID,
		&ord.Address,
		&ord.TotalAmount,
		&ord.Status,
		&ord.CreatedAt,
		&ord.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrOrderNotFound
		}
		return nil, fmt.Errorf("get order failed: %w", err)
	}

	// Find order_items table
	queryItems := `
		SELECT oi.id, oi.product_id, p.name, oi.quantity, oi.price
		FROM order_items oi
		JOIN products p ON oi.product_id = p.id
		WHERE oi.order_id = $1
	`
	rows, err := r.db.QueryContext(ctx, queryItems, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []order.OrderItem

	for rows.Next() {
		var item order.OrderItem
		if err := rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.ProductName,
			&item.Quantity,
			&item.Price,
		); err != nil {
			return nil, fmt.Errorf("scan item failed: %w", err)
		}
		items = append(items, item)
	}
	// Add items to order
	ord.Items = items

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return ord, nil
}

func (r *orderRepository) InsertOrderTx(ctx context.Context, tx *sql.Tx, userID string, totalAmount int64, address string) (int64, time.Time, error) {
	var orderID int64
	var createdAt time.Time

	query := `
		INSERT INTO orders (user_id, total_amount, status, address)
		VALUES ($1, $2, 'PENDING', $3) RETURNING id, created_at
	`
	err := tx.QueryRowContext(ctx, query, userID, totalAmount, address).Scan(&orderID, &createdAt)
	if err != nil {
		return 0, time.Time{}, err
	}
	return orderID, createdAt, nil
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

func (r *orderRepository) FindMyOrders(ctx context.Context, userID string, limit, offset int) ([]*order.Order, int64, error) {
	query := `
		SELECT id, created_at, status, total_amount
		FROM orders
		WHERE user_id = $1 ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, errs.ErrUserNotFound
		}
		return nil, 0, err
	}
	defer rows.Close()

	var orders []*order.Order

	for rows.Next() {
		o := new(order.Order)
		if err := rows.Scan(
			&o.ID,
			&o.CreatedAt,
			&o.Status,
			&o.TotalAmount,
		); err != nil {
			return nil, 0, err
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// --- Count Orders
	var total int64
	queryCount := `SELECT COUNT(*) FROM orders WHERE user_id = $1`

	err = r.db.QueryRowContext(ctx, queryCount, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

func (r *orderRepository) FindOrderItemsTx(ctx context.Context, tx *sql.Tx, orderID int64) ([]*order.OrderItem, error) {
	query := `
		SELECT product_id, quantity
		FROM order_items
		WHERE order_id = $1
	`
	rows, err := tx.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*order.OrderItem

	for rows.Next() {
		item := new(order.OrderItem)
		if err := rows.Scan(
			&item.ProductID,
			&item.Quantity,
		); err != nil {
			return nil, err
		}
		orders = append(orders, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *orderRepository) UpdateStatusTx(ctx context.Context, tx *sql.Tx, orderID int64, status order.OrderStatus) error {
	query := `UPDATE orders SET status = $1 WHERE id = $2`
	res, err := tx.ExecContext(ctx, query, status, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errs.ErrOrderNotFound
		}
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errs.ErrOrderNotFound
	}
	return nil
}
