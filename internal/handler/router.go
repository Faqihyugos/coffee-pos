package handler

import (
	"database/sql"

	"github.com/faqihyugos/coffee-pos/config"
	"github.com/faqihyugos/coffee-pos/internal/entity"
	"github.com/faqihyugos/coffee-pos/internal/middleware"
	"github.com/faqihyugos/coffee-pos/internal/repository"
	"github.com/faqihyugos/coffee-pos/internal/service"
	"github.com/faqihyugos/coffee-pos/pkg/response"
	"github.com/faqihyugos/coffee-pos/pkg/validator"
	"github.com/gin-gonic/gin"
)

// NewRouter initialises a Gin engine with global middleware and all routes.
func NewRouter(db *sql.DB, cfg *config.Config, v *validator.Validator) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// 1. Repositories
	userRepo := repository.NewUserRepository(db)

	// 2. Services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpireHours)

	// 3. Handlers
	authHandler := NewAuthHandler(authService, v)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", healthCheck)

		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Route group untuk owner — semua endpoint di sini butuh login + role owner
		ownerGroup := v1.Group("/owner")
		ownerGroup.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		ownerGroup.Use(middleware.RoleMiddleware(entity.RoleOwner))
		_ = ownerGroup

		// Route group untuk cashier — semua endpoint di sini butuh login + role cashier
		cashierGroup := v1.Group("/cashier")
		cashierGroup.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		cashierGroup.Use(middleware.RoleMiddleware(entity.RoleCashier))
		_ = cashierGroup
	}

	return r
}

func healthCheck(c *gin.Context) {
	response.OK(c, "server is running", nil)
}

