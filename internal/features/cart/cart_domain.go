package cart

import "time"

type Cart struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CartItem struct {
	ID        int64     `json:"id" db:"id"`
	CartID    int64     `json:"cart_id" db:"cart_id"`
	ProductID int64     `json:"product_id" db:"product_id"`
	Quantity  int       `json:"quantity" db:"quantity"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CartItemResult from DB
type CartItemResult struct {
	ID          int64  `db:"id"`
	ProductID   int64  `db:"product_id"`
	Quantity    int    `db:"quantity"`
	ProductName string `db:"product_name"`
	Price       int    `db:"price"`
	Stock       int    `db:"stock"`
}

type CartResponse struct {
	Items       []CartItemData `json:"items"`
	TotolAmount int            `json:"total_amount"`
	TotalQty    int            `json:"total_qty"`
}

type CartItemData struct {
	ID          int64  `json:"id"`
	ProductID   int64  `json:"product_id"`
	ProductName string `json:"product_name"`
	Price       int    `json:"price"`
	Quantity    int    `json:"quantity"`
	Total       int    `json:"total"`
	IsStockOK   bool   `json:"is_stock_ok"`
}
