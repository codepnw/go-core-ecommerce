package carthandler

import (
	"net/http"
	"strconv"

	"github.com/codepnw/go-starter-kit/internal/auth"
	"github.com/codepnw/go-starter-kit/internal/errs"
	cartservice "github.com/codepnw/go-starter-kit/internal/features/cart/service"
	producthandler "github.com/codepnw/go-starter-kit/internal/features/product/handler"
	"github.com/codepnw/go-starter-kit/pkg/utils/response"
	"github.com/gin-gonic/gin"
)

type cartHandler struct {
	service cartservice.CartService
}

func NewCartHandler(service cartservice.CartService) *cartHandler {
	return &cartHandler{service: service}
}

func (h *cartHandler) AddItem(c *gin.Context) {
	req := new(AddToCartReq)
	if err := c.ShouldBindJSON(req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	// Get UserID Context
	userID, err := auth.GetUserIDFromContext(c.Request.Context())
	if err != nil {
		response.ResponseError(c, http.StatusUnauthorized, err)
		return
	}

	// AddItem Service
	err = h.service.AddItem(c.Request.Context(), userID, req.ProductID, req.Quantity)
	if err != nil {
		switch err {
		case errs.ErrUnauthorized:
			response.ResponseError(c, http.StatusUnauthorized, err)
		case errs.ErrStockNotEnough:
			response.ResponseError(c, http.StatusBadRequest, err)
		case errs.ErrProductNotFound:
			response.ResponseError(c, http.StatusNotFound, err)
		default:
			response.ResponseError(c, http.StatusInternalServerError, err)
		}
		return
	}

	response.ResponseSuccess(c, http.StatusOK, "added product to cart")
}

func (h *cartHandler) GetCart(c *gin.Context) {
	userID, err := auth.GetUserIDFromContext(c.Request.Context())
	if err != nil {
		response.ResponseError(c, http.StatusUnauthorized, err)
		return
	}

	resp, err := h.service.GetCart(c.Request.Context(), userID)
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

func (h *cartHandler) RemoveItme(c *gin.Context) {
	pIDStr := c.Param(producthandler.ParamProductID)
	productID, err := strconv.ParseInt(pIDStr, 10, 64)
	if err != nil {
		response.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	userID, err := auth.GetUserIDFromContext(c.Request.Context())
	if err != nil {
		response.ResponseError(c, http.StatusUnauthorized, err)
		return
	}

	err = h.service.RemoveItem(c.Request.Context(), userID, productID)
	if err != nil {
		switch err {
		case errs.ErrProductNotFound:
			response.ResponseError(c, http.StatusNotFound, err)
		default:
			response.ResponseError(c, http.StatusInternalServerError, err)
		}
		return
	}
	
	response.ResponseSuccess(c, http.StatusNoContent, nil)
}
