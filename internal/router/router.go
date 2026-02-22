package router

import (
	"flash-sale-be/internal/config"
	"flash-sale-be/internal/handler"
	"flash-sale-be/internal/middleware"
	"flash-sale-be/internal/repository"
	"flash-sale-be/internal/service"

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
	authHandler := handler.NewAuthHandler(authSvc)

	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/me", middleware.Jwt(deps.Cfg), authHandler.Me)
		}
	}

	return r
}
