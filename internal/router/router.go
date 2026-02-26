package router

import (
	"flash-sale-be/internal/config"
	"flash-sale-be/internal/handler"
	"flash-sale-be/internal/middleware"
	"flash-sale-be/internal/repository"
	"flash-sale-be/internal/service"
	"flash-sale-be/internal/store"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Deps struct {
	DB               *gorm.DB
	Cfg              *config.Config
	CheckoutService  service.CheckoutService
}

func New(deps Deps) *gin.Engine {
	// Auth
	userRepo := repository.NewUserRepository(deps.DB)
	authSvc := service.NewAuthService(userRepo, deps.Cfg)
	tokenBlacklist := store.NewMemoryBlacklist()
	authHandler := handler.NewAuthHandler(authSvc, tokenBlacklist)

	// Products
	productsRepo := repository.NewProductsRepository(deps.DB)
	productsService := service.NewProductsService(productsRepo)
	productsHandler := handler.NewProductsHandler(productsService)

	// Checkout (requires Deps.CheckoutService from main)
	checkoutHandler := handler.NewCheckoutHandler(deps.CheckoutService)

	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", handler.Ping)

		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", middleware.Jwt(deps.Cfg, tokenBlacklist), authHandler.Logout)
			auth.GET("/me", middleware.Jwt(deps.Cfg, tokenBlacklist), authHandler.Me)
		}
		products := v1.Group("/products")
		{
			products.POST("/", middleware.Jwt(deps.Cfg, tokenBlacklist), productsHandler.CreateProduct)
			products.PUT("/:id", middleware.Jwt(deps.Cfg, tokenBlacklist), productsHandler.UpdateProduct)
			products.GET("/all", middleware.Jwt(deps.Cfg, tokenBlacklist), productsHandler.GetAllProducts)
			products.GET("/:id", middleware.Jwt(deps.Cfg, tokenBlacklist), productsHandler.GetProductById)
			products.GET("/", middleware.Jwt(deps.Cfg, tokenBlacklist), productsHandler.GetAllProductsByUser)
			products.DELETE("/:id", middleware.Jwt(deps.Cfg, tokenBlacklist), productsHandler.DeleteProduct)
		}
		checkouts := v1.Group("/checkouts")
		checkouts.Use(middleware.Jwt(deps.Cfg, tokenBlacklist))
		{
			checkouts.POST("/", checkoutHandler.Checkout)
		}
	}

	return r
}
