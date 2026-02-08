package cartrepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/codepnw/go-starter-kit/internal/errs"
	"github.com/codepnw/go-starter-kit/internal/features/cart"
)

//go:generate mockgen -source=cart_repository.go -destination=cart_repository_mock.go -package=cartrepository
type CartRepository interface {
	FindCartID(ctx context.Context, userID string) (int64, error)
	AddItem(ctx context.Context, cartID, productID int64, quantity int) error
	GetCartItems(ctx context.Context, userID string) ([]*cart.CartItemResult, error)
	RemoveItem(ctx context.Context, cartID, productID int64) error
}

type cartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) FindCartID(ctx context.Context, userID string) (int64, error) {
	var cartID int64
	query := `
		INSERT INTO carts (user_id) VALUES ($1)
		ON CONFLICT (user_id)
			DO UPDATE SET updated_at = NOW()
		RETURNING id
	`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&cartID)
	if err != nil {
		return 0, err
	}
	return cartID, nil
}

func (r *cartRepository) AddItem(ctx context.Context, cartID, productID int64, quantity int) error {
	query := `
		INSERT INTO cart_items (cart_id, product_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT ON CONSTRAINT cart_items_unique
		DO UPDATE SET
			quantity = cart_items.quantity + EXCLUDED.quantity,
			updated_at = NOW()
	`
	res, err := r.db.ExecContext(ctx, query, cartID, productID, quantity)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errs.ErrProductNotFound
		}
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errs.ErrProductNotFound
	}
	return nil
}

func (r *cartRepository) GetCartItems(ctx context.Context, userID string) ([]*cart.CartItemResult, error) {
	query := `
		SELECT
	 		ci.id,
			ci.product_id,
			ci.quantity,
			p.name AS product_name,
			p.price,
			p.stock
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.cart_id = (SELECT id FROM carts WHERE user_id = $1)
		ORDER BY ci.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}
		return nil, err
	}

	var items []*cart.CartItemResult

	for rows.Next() {
		item := new(cart.CartItemResult)
		if err := rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.Quantity,
			&item.ProductName,
			&item.Price,
			&item.Stock,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *cartRepository) RemoveItem(ctx context.Context, cartID, productID int64) error {
	query := `
		DELETE FROM cart_items 
		WHERE cart_id = $1 AND product_id = $2
	`
	res, err := r.db.ExecContext(ctx, query, cartID, productID)
	if err != nil {
		return err
	}
	
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errs.ErrProductNotFound
	}
	return nil
}