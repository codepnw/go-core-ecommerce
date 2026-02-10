package orderhandler

import (
	"net/http"
	"strconv"

	"github.com/codepnw/go-starter-kit/internal/auth"
	"github.com/codepnw/go-starter-kit/internal/errs"
	orderservice "github.com/codepnw/go-starter-kit/internal/features/order/service"
	"github.com/codepnw/go-starter-kit/pkg/utils/response"
	"github.com/gin-gonic/gin"
)

type orderHandler struct {
	service orderservice.OrderService
}

func NewOrderHandler(service orderservice.OrderService) *orderHandler {
	return &orderHandler{service: service}
}

func (h *orderHandler) CreateOrder(c *gin.Context) {
	userID, err := auth.GetUserIDFromContext(c.Request.Context())
	if err != nil {
		response.ResponseError(c, http.StatusUnauthorized, err)
		return
	}

	req := new(CreateOrderReq)
	if err := c.ShouldBindJSON(req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	orderNo, err := h.service.CreateOrder(c.Request.Context(), userID, req.Address)
	if err != nil {
		switch err {
		case errs.ErrCartEmpty:
			response.ResponseError(c, http.StatusBadRequest, err)
		default:
			response.ResponseError(c, http.StatusInternalServerError, err)
		}
		return
	}

	response.ResponseSuccess(c, http.StatusOK, gin.H{
		"message":  "order created successfully",
		"order_no": orderNo,
	})
}

func (h *orderHandler) GetOrderDetails(c *gin.Context) {
	orderID, _ := strconv.ParseInt(c.Param(ParamOrderID), 10, 64)

	resp, err := h.service.GetOrderDetails(c.Request.Context(), orderID)
	if err != nil {
		switch err {
		case errs.ErrOrderNotFound:
			response.ResponseError(c, http.StatusNotFound, err)
		default:
			response.ResponseError(c, http.StatusInternalServerError, err)
		}
		return
	}

	response.ResponseSuccess(c, http.StatusOK, resp)
}
