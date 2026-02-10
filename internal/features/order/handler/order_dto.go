package orderhandler

const ParamOrderID = "order_id"

type CreateOrderReq struct {
	Address string `json:"address" binding:"required"`
}
