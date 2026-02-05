package producthandler

import (
	productservice "github.com/codepnw/go-starter-kit/internal/features/product/service"
	"github.com/gin-gonic/gin"
)

type productHandler struct {
	service productservice.ProductService
}

func NewProductHandler(service productservice.ProductService) *productHandler {
	return &productHandler{service: service}
}

func (h *productHandler) CreateProduct(c *gin.Context) {
}

func (h *productHandler) GetProduct(c *gin.Context) {
}

func (h *productHandler) GetProducts(c *gin.Context) {
}

func (h *productHandler) UpdateProduct(c *gin.Context) {
}

func (h *productHandler) DeleteProduct(c *gin.Context) {
}
