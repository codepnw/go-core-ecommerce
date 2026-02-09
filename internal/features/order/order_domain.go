package order

import "time"

type OrderStatus string

const (
	StatusPending   OrderStatus = "PENDING"
	StatusPaid      OrderStatus = "PAID"
	StatusShipped   OrderStatus = "SHIPPED"
	StatusCompleted OrderStatus = "COMPLETED"
	StatusCancelled OrderStatus = "CANCELLED"
)

type Order struct {
	ID          int64       `json:"id" db:"id"`
	UserID      string      `json:"user_id" db:"user_id"`
	TotalAmount int         `json:"total_amount" db:"total_amount"`
	Status      OrderStatus `json:"status" db:"status"`
	Address     string      `json:"address" db:"address"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

type OrderItem struct {
	ID        int64 `json:"id" db:"id"`
	OrderID   int64 `json:"order_id" db:"order_id"`
	ProductID int64 `json:"product_id" db:"product_id"`
	Quantity  int   `json:"quantity" db:"quantity"`
	Price     int   `json:"price" db:"price"`
}

type OrderItemReq struct {
	OrderID   int64 `json:"order_id"`
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
	Price     int   `json:"price"`
}
