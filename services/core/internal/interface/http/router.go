package http

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/auth"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/interface/middleware"
	"go.uber.org/zap"
)

type RouterConfig struct {
	AuthHandler    *AuthHandler
	UserHandler    *UserHandler
	ModuleHandler  *ModuleHandler
	GatewayProxy   *GatewayProxy
	ChatHandler    *ChatHandler
	JWTService     *auth.JWTService
	Logger         *zap.Logger
}

func NewRouter(cfg RouterConfig) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.LoggingMiddleware(cfg.Logger))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// WebSocket chat endpoint — auth via ?token= query param
	if cfg.ChatHandler != nil {
		r.GET("/ws/chat", cfg.ChatHandler.HandleWebSocket)
	}

	api := r.Group("/api")

	// Auth routes (public, rate-limited)
	authGroup := api.Group("/auth")
	authGroup.Use(middleware.RateLimitMiddleware(20, time.Minute))
	{
		authGroup.POST("/register", cfg.AuthHandler.Register)
		authGroup.POST("/login", cfg.AuthHandler.Login)
		authGroup.POST("/refresh", cfg.AuthHandler.Refresh)
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(cfg.JWTService))
	{
		// User routes
		users := protected.Group("/users")
		{
			users.GET("", cfg.UserHandler.ListUsers)
			users.GET("/:id", cfg.UserHandler.GetUser)
			users.PUT("/:id", cfg.UserHandler.UpdateUser)
			users.DELETE("/:id", middleware.RequireRole("admin"), cfg.UserHandler.DeleteUser)
		}

		// Module routes (admin only)
		if cfg.ModuleHandler != nil {
			modules := protected.Group("/modules")
			modules.Use(middleware.RequireRole("admin"))
			{
				modules.GET("", cfg.ModuleHandler.ListModules)
				modules.POST("", cfg.ModuleHandler.RegisterModule)
				modules.DELETE("/:name", cfg.ModuleHandler.UnregisterModule)
			}
		}

		// Gateway proxy for module APIs
		if cfg.GatewayProxy != nil {
			protected.Any("/hr/*path", cfg.GatewayProxy.ProxyHandler())
			protected.Any("/subjects/*path", cfg.GatewayProxy.ProxyHandler())
			protected.Any("/timetable/*path", cfg.GatewayProxy.ProxyHandler())
		}
	}

	return r
}
