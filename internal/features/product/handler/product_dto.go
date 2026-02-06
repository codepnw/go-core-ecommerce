package producthandler

const ParamProductID = "product_id"

type ProductCreateReq struct {
	Name  string `json:"name" binding:"required,min=3"`
	Price int    `json:"price" binding:"required,gt=0"`
	Stock int    `json:"stock" binding:"required,gt=0"`
	SKU   string `json:"sku" binding:"required,min=3"`
}

type ProductUpdateReq struct {
	Name  *string `json:"name" binding:"omitempty,min=3"`
	Price *int    `json:"price" binding:"omitempty,gt=0"`
	SKU   *string `json:"sku" binding:"omitempty,min=3"`
}

type IncreaseStockReq struct {
	Qty int `json:"qty" binding:"required,gt=0"`
}