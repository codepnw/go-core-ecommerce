package orderhandler

type CreateOrderReq struct {
	Address string `json:"address" binding:"required"`
}
