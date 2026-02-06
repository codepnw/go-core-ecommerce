package producthandler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/codepnw/go-starter-kit/internal/errs"
	"github.com/codepnw/go-starter-kit/internal/features/product"
	productservice "github.com/codepnw/go-starter-kit/internal/features/product/service"
	"github.com/codepnw/go-starter-kit/pkg/utils/response"
	"github.com/gin-gonic/gin"
)

type productHandler struct {
	service productservice.ProductService
}

func NewProductHandler(service productservice.ProductService) *productHandler {
	return &productHandler{service: service}
}

func (h *productHandler) CreateProduct(c *gin.Context) {
	req := new(ProductCreateReq)

	if err := c.ShouldBindJSON(req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	input := &product.Product{
		Name:  req.Name,
		Price: req.Price,
		Stock: req.Stock,
		SKU:   req.SKU,
	}

	if err := h.service.CreateProduct(c.Request.Context(), input); err != nil {
		switch err {
		case errs.ErrProductSKUExists:
			response.ResponseError(c, http.StatusBadRequest, err)
		default:
			response.ResponseError(c, http.StatusInternalServerError, err)
		}
		return
	}

	response.ResponseSuccess(c, http.StatusCreated, input)
}

func (h *productHandler) GetProduct(c *gin.Context) {
	productID, _ := strconv.Atoi(c.Param(ParamProductID))

	resp, err := h.service.GetProduct(c.Request.Context(), int64(productID))
	if err != nil {
		switch err {
		case errs.ErrProductNotFound:
			response.ResponseError(c, http.StatusNotFound, err)
		default:
			response.ResponseError(c, http.StatusInternalServerError, err)
		}
		return
	}

	response.ResponseSuccess(c, http.StatusOK, resp)
}

func (h *productHandler) GetProducts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	resp, err := h.service.GetProducts(c.Request.Context(), limit, offset)
	if err != nil {
		response.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	response.ResponseSuccess(c, http.StatusOK, resp)
}

func (h *productHandler) IncreaseStock(c *gin.Context) {
	id, err := h.getProductID(c)
	if err != nil {
		response.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	req := new(IncreaseStockReq)

	if err := c.ShouldBindJSON(req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	if err := h.service.IncreaseStock(c.Request.Context(), id, req.Qty); err != nil {
		switch err {
		case errs.ErrProductNotFound:
			response.ResponseError(c, http.StatusNotFound, err)
		default:
			response.ResponseError(c, http.StatusInternalServerError, err)
		}
		return
	}

	msg := fmt.Sprintf("product id %d increase stock %d", id, req.Qty)
	response.ResponseSuccess(c, http.StatusOK, msg)
}

func (h *productHandler) UpdateProduct(c *gin.Context) {
	id, err := h.getProductID(c)
	if err != nil {
		response.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	req := new(ProductUpdateReq)

	if err := c.ShouldBindJSON(req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	input := productservice.UpdateProductInput{
		ID:    id,
		Name:  req.Name,
		Price: req.Price,
		SKU:   req.SKU,
	}

	if err := h.service.UpdateProduct(c.Request.Context(), input); err != nil {
		response.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	msg := fmt.Sprintf("product id %d updated", id)
	response.ResponseSuccess(c, http.StatusOK, msg)
}

func (h *productHandler) DeleteProduct(c *gin.Context) {
	id, err := h.getProductID(c)
	if err != nil {
		response.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	if err := h.service.DeleteProduct(c.Request.Context(), id); err != nil {
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

func (h *productHandler) getProductID(c *gin.Context) (int64, error) {
	id, err := strconv.Atoi(c.Param(ParamProductID))
	if err != nil {
		return 0, err
	}
	return int64(id), nil
}
