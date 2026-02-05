package product

type Product struct {
	ID      int64  `json:"id" db:"id"`
	Name    string `json:"name" db:"name"`
	Price   int    `json:"price" db:"price"`
	Stock   int    `json:"stock" db:"stock"`
	SKU     string `json:"sku" db:"sku"`
	Version int    `json:"version" db:"version"`
}
