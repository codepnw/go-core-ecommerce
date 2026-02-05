package server

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/codepnw/go-starter-kit/internal/config"
	userhandler "github.com/codepnw/go-starter-kit/internal/features/user/handler"
	userrepository "github.com/codepnw/go-starter-kit/internal/features/user/repository"
	userservice "github.com/codepnw/go-starter-kit/internal/features/user/service"
	"github.com/codepnw/go-starter-kit/internal/middleware"
	"github.com/codepnw/go-starter-kit/pkg/database"
	jwttoken "github.com/codepnw/go-starter-kit/pkg/jwttoken"
	"github.com/codepnw/go-starter-kit/pkg/utils/response"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	db     *sql.DB
	router *gin.Engine
	token  jwttoken.JWTToken
	mid    *middleware.Middleware
	tx     database.TxManager
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
	
	// Prefix Default: /api/v1
	prefix := s.router.Group(cfg.APP.Prefix)
	
	// Register Routes
	s.registerHealthRoutes(prefix)
	s.registerUserRoutes(prefix)

	return s, nil
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) registerHealthRoutes(r *gin.RouterGroup) {
	r.GET("/health", func(c *gin.Context) {
		response.ResponseSuccess(c, http.StatusOK, "Go Starter Kit Running...")
	})
}

func (s *Server) registerUserRoutes(r *gin.RouterGroup) {
	repo := userrepository.NewUserRepository(s.db)
	service := userservice.NewUserService(s.tx, s.token, repo)
	handler := userhandler.NewUserHandler(service)

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
