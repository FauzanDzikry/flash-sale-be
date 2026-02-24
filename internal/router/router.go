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
	DB  *gorm.DB
	Cfg *config.Config
}

func New(deps Deps) *gin.Engine {
	userRepo := repository.NewUserRepository(deps.DB)
	authSvc := service.NewAuthService(userRepo, deps.Cfg)
	tokenBlacklist := store.NewMemoryBlacklist()
	authHandler := handler.NewAuthHandler(authSvc, tokenBlacklist)

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
	}

	return r
}
