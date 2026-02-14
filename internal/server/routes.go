package server

import (
	"fmt"
	"net/http"

	"github.com/codepnw/go-starter-kit/pkg/utils/response"
	"github.com/gin-gonic/gin"

	orderhandler "github.com/codepnw/go-starter-kit/internal/features/order/handler"
	producthandler "github.com/codepnw/go-starter-kit/internal/features/product/handler"
)

// -------------------- HEALTH Routes -----------------------
func (s *Server) registerHealthRoutes(r *gin.RouterGroup) {
	r.GET("/health", func(c *gin.Context) {
		response.ResponseSuccess(c, http.StatusOK, "Core Ecommerce API Running...")
	})
}

// -------------------- USER Routes -----------------------
func (s *Server) registerUserRoutes(r *gin.RouterGroup) {
	handler := s.handlerUser

	// Auth Routes
	auth := r.Group("/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
		auth.POST("/refresh-token", handler.RefreshToken)

		// Authorized
		auth.POST("/logout", handler.Logout, s.mid.Authorized())
	}

	// Users Routes
	users := r.Group("/users", s.mid.Authorized())
	{
		users.GET("/profile", handler.GetProfile)
	}
}

// -------------------- PRODUCT Routes -----------------------
func (s *Server) registerProductRoutes(r *gin.RouterGroup) {
	handler := s.handlerProduct
	paramID := fmt.Sprintf("/:%s", producthandler.ParamProductID)

	// Public Routes
	public := r.Group("/products")
	{
		public.GET("/", handler.GetProducts)
		public.GET(paramID, handler.GetProduct)
	}

	// Authorized Routes
	authorized := r.Group("/products", s.mid.Authorized())
	{
		authorized.POST("/", handler.CreateProduct)
		authorized.PATCH(paramID, handler.UpdateProduct)
		authorized.DELETE(paramID, handler.DeleteProduct)
		authorized.POST(paramID+"/stock", handler.IncreaseStock)
	}
}

// -------------------- CART Routes -----------------------
func (s *Server) registerCartRoutes(r *gin.RouterGroup) {
	handler := s.handlerCart

	carts := r.Group("/cart", s.mid.Authorized())
	{
		carts.GET("/", handler.GetCart)
		carts.POST("/items", handler.AddItem)
		carts.DELETE(fmt.Sprintf("/items/:%s", producthandler.ParamProductID), handler.RemoveItme)
	}
}

// -------------------- ORDER Routes -----------------------
func (s *Server) registerOrderRoutes(r *gin.RouterGroup) {
	handler := s.handlerOrder
	paramID := fmt.Sprintf("/:%s", orderhandler.ParamOrderID)

	orders := r.Group("/orders", s.mid.Authorized())
	{
		orders.GET("/", handler.MyOrders)
		orders.POST("/checkout", handler.CreateOrder)
		orders.GET(paramID, handler.GetOrderDetails)
	}
}
