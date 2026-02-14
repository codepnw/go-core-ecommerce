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

type OrderHandler struct {
	service orderservice.OrderService
}

func NewOrderHandler(service orderservice.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
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

func (h *OrderHandler) GetOrderDetails(c *gin.Context) {
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

func (h *OrderHandler) MyOrders(c *gin.Context) {
	userID, err := auth.GetUserIDFromContext(c.Request.Context())
	if err != nil {
		response.ResponseError(c, http.StatusUnauthorized, err)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	resp, err := h.service.MyOrders(c.Request.Context(), userID, page, limit)
	if err != nil {
		switch err {
		case errs.ErrUserNotFound:
			response.ResponseError(c, http.StatusNotFound, err)
		default:
			response.ResponseError(c, http.StatusInternalServerError, err)
		}
		return
	}

	response.ResponseSuccess(c, http.StatusOK, resp)
}
