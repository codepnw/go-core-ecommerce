package server

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/codepnw/go-starter-kit/internal/config"
	carthandler "github.com/codepnw/go-starter-kit/internal/features/cart/handler"
	cartrepository "github.com/codepnw/go-starter-kit/internal/features/cart/repository"
	cartservice "github.com/codepnw/go-starter-kit/internal/features/cart/service"
	orderhandler "github.com/codepnw/go-starter-kit/internal/features/order/handler"
	orderrepository "github.com/codepnw/go-starter-kit/internal/features/order/repository"
	orderservice "github.com/codepnw/go-starter-kit/internal/features/order/service"
	producthandler "github.com/codepnw/go-starter-kit/internal/features/product/handler"
	productrepository "github.com/codepnw/go-starter-kit/internal/features/product/repository"
	productservice "github.com/codepnw/go-starter-kit/internal/features/product/service"
	userhandler "github.com/codepnw/go-starter-kit/internal/features/user/handler"
	userrepository "github.com/codepnw/go-starter-kit/internal/features/user/repository"
	userservice "github.com/codepnw/go-starter-kit/internal/features/user/service"
	"github.com/codepnw/go-starter-kit/internal/middleware"
	"github.com/codepnw/go-starter-kit/pkg/database"
	jwttoken "github.com/codepnw/go-starter-kit/pkg/jwttoken"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	db     *sql.DB
	router *gin.Engine
	token  jwttoken.JWTToken
	mid    *middleware.Middleware
	tx     database.TxManager
	// Handler Domain
	handlerUser    *userhandler.UserHandler
	handlerProduct *producthandler.ProductHandler
	handlerCart    *carthandler.CartHandler
	handlerOrder   *orderhandler.OrderHandler
}

func NewServer(cfg *config.EnvConfig, db *sql.DB) (*Server, error) {
	r := gin.New()

	// JWT Token
	token, err := jwttoken.NewJWTToken(cfg.JWT.AppName, cfg.JWT.SecretKey, cfg.JWT.RefreshKey)
	if err != nil {
		return nil, err
	}

	// Middleware
	mid := middleware.InitMiddleware(token)

	// DB Transaction
	tx := database.NewDBTransaction(db)

	// Denpendency Injection
	s := &Server{
		db:     db,
		router: r,
		token:  token,
		mid:    mid,
		tx:     tx,
	}

	// Gin Middleware
	s.ginMiddleware(r)
	
	// Setup Domain Handler
	s.setupHandler()

	// Prefix Default: /api/v1
	prefix := s.router.Group(cfg.APP.Prefix)

	// Register Routes
	s.registerHealthRoutes(prefix)
	s.registerUserRoutes(prefix)
	s.registerProductRoutes(prefix)
	s.registerCartRoutes(prefix)
	s.registerOrderRoutes(prefix)

	return s, nil
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) ginMiddleware(r *gin.Engine) {
	r.Use(gin.Recovery())
	r.Use(s.mid.Logger())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
}

func (s *Server) setupHandler() {
	// User Handler Setup
	userRepo := userrepository.NewUserRepository(s.db)
	userService := userservice.NewUserService(s.tx, s.token, userRepo)
	s.handlerUser = userhandler.NewUserHandler(userService)

	// Product Handler Setup
	prodRepo := productrepository.NewProductRepository(s.db)
	prodService := productservice.NewProductService(prodRepo)
	s.handlerProduct = producthandler.NewProductHandler(prodService)

	// Cart Handler Setup
	cartRepo := cartrepository.NewCartRepository(s.db)
	cartSrv := cartservice.NewCartService(cartRepo, prodService)
	s.handlerCart = carthandler.NewCartHandler(cartSrv)

	// Order Handler Setup
	ordRepo := orderrepository.NewOrderRepository(s.db)
	ordService := orderservice.NewOrderService(s.tx, ordRepo, prodRepo, cartRepo)
	s.handlerOrder = orderhandler.NewOrderHandler(ordService)
}
