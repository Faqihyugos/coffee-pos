package handler

import (
	"database/sql"

	"github.com/faqihyugos/coffee-pos/config"
	"github.com/faqihyugos/coffee-pos/internal/entity"
	"github.com/faqihyugos/coffee-pos/internal/middleware"
	"github.com/faqihyugos/coffee-pos/internal/repository"
	"github.com/faqihyugos/coffee-pos/internal/service"
	"github.com/faqihyugos/coffee-pos/pkg/response"
	"github.com/faqihyugos/coffee-pos/pkg/txmanager"
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
	categoryRepo := repository.NewCategoryRepository(db)
	productRepo := repository.NewProductRepository(db)
	stockRepo := repository.NewStockRepository(db)
	tableRepo := repository.NewTableRepository(db)
	txMgr := txmanager.New(db)

	// 2. Services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpireHours)
	userService := service.NewUserService(userRepo, authService)
	categoryService := service.NewCategoryService(categoryRepo)
	productService := service.NewProductService(productRepo, categoryRepo)
	stockService := service.NewStockService(stockRepo, productRepo, txMgr)
	tableService := service.NewTableService(tableRepo)

	// 3. Handlers
	authHandler := NewAuthHandler(authService, v)
	userHandler := NewUserHandler(userService, v)
	categoryHandler := NewCategoryHandler(categoryService, v)
	productHandler := NewProductHandler(productService, v)
	stockHandler := NewStockHandler(stockService, v)
	tableHandler := NewTableHandler(tableService, v)

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
		ownerGroup.GET("/categories", categoryHandler.FindAll)
		ownerGroup.GET("/categories/:id", categoryHandler.FindByID)
		ownerGroup.POST("/categories", categoryHandler.Create)
		ownerGroup.PUT("/categories/:id", categoryHandler.Update)
		ownerGroup.DELETE("/categories/:id", categoryHandler.Delete)
		ownerGroup.GET("/products", productHandler.FindAll)
		ownerGroup.GET("/products/:id", productHandler.FindByID)
		ownerGroup.POST("/products", productHandler.Create)
		ownerGroup.PUT("/products/:id", productHandler.Update)
		ownerGroup.DELETE("/products/:id", productHandler.Delete)
		ownerGroup.GET("/products/:id/stock", stockHandler.GetStock)
		ownerGroup.POST("/products/:id/stock/adjustment", stockHandler.Adjust)
		ownerGroup.GET("/products/:id/stock/movements", stockHandler.GetMovements)
		ownerGroup.GET("/tables", tableHandler.FindAll)
		ownerGroup.POST("/tables", tableHandler.Create)
		ownerGroup.PUT("/tables/:id", tableHandler.Update)
		ownerGroup.DELETE("/tables/:id", tableHandler.Delete)
		ownerGroup.GET("/cashiers", userHandler.FindAll)
		ownerGroup.POST("/cashiers", userHandler.Create)
		ownerGroup.PUT("/cashiers/:id", userHandler.Update)
		ownerGroup.PATCH("/cashiers/:id/toggle-status", userHandler.ToggleStatus)

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
