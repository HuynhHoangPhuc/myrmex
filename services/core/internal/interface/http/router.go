package http

import (
	"time"

	"github.com/gin-gonic/gin"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/HuynhHoangPhuc/myrmex/pkg/messaging"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/auth"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/interface/middleware"
	"go.uber.org/zap"
)

type RouterConfig struct {
	AuthHandler          *AuthHandler
	OAuthHandler         *OAuthHandler
	AdminRoleHandler     *AdminRoleHandler
	AuditHandler         *AuditHandler
	AuditPublisher       messaging.Publisher // nil → audit middleware disabled
	UserHandler          *UserHandler
	ModuleHandler        *ModuleHandler
	GatewayProxy         *GatewayProxy
	ChatHandler          *ChatHandler
	NotificationHandler  *NotificationHandler
	HRHandler            *HRHandler
	SubjectHandler       *SubjectHandler
	TimetableHandler     *TimetableHandler
	StudentHandler       *StudentHandler
	StudentPortalHandler *StudentPortalHandler
	DashboardHandler     *DashboardHandler
	ImportHandler        *ImportHandler
	AnalyticsHTTPAddr    string // e.g. "http://module-analytics:8055"
	NotificationHTTPAddr string // e.g. "http://module-notification:8056"
	JWTService           *auth.JWTService
	Logger               *zap.Logger
	DB                   *pgxpool.Pool // for health check
	Rdb                  *redis.Client // for health check
}

func NewRouter(cfg RouterConfig) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	// Sentry middleware first so it captures panics from all subsequent handlers
	r.Use(sentrygin.New(sentrygin.Options{Repanic: true}))
	r.Use(gin.Recovery())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.LoggingMiddleware(cfg.Logger))

	// Health check
	r.GET("/health", healthHandler(cfg.DB, cfg.Rdb))

	// WebSocket endpoints — auth via ?token= query param
	if cfg.ChatHandler != nil {
		r.GET("/ws/chat", cfg.ChatHandler.HandleWebSocket)
	}
	if cfg.NotificationHandler != nil {
		r.GET("/ws/notifications", cfg.NotificationHandler.HandleWebSocket)
	}

	api := r.Group("/api")

	// Auth routes (public, rate-limited)
	authGroup := api.Group("/auth")
	authGroup.Use(middleware.RateLimitMiddleware(10, time.Minute))
	{
		authGroup.POST("/register", cfg.AuthHandler.Register)
		authGroup.POST("/register-student", cfg.AuthHandler.RegisterStudent)
		authGroup.POST("/login", cfg.AuthHandler.Login)
		authGroup.POST("/refresh", cfg.AuthHandler.Refresh)

		// OAuth/OIDC routes — only registered when OAuthHandler is configured
		if cfg.OAuthHandler != nil {
			oauth := authGroup.Group("/oauth")
			oauth.GET("/google/login", cfg.OAuthHandler.GoogleLogin)
			oauth.GET("/google/callback", cfg.OAuthHandler.GoogleCallback)
			oauth.GET("/microsoft/login", cfg.OAuthHandler.MicrosoftLogin)
			oauth.GET("/microsoft/callback", cfg.OAuthHandler.MicrosoftCallback)
			oauth.POST("/exchange", cfg.OAuthHandler.ExchangeAuthCode)
		}
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(cfg.JWTService))
	protected.Use(middleware.AuditMiddleware(cfg.AuditPublisher))
	{
		// Current user profile
		protected.GET("/auth/me", cfg.UserHandler.Me)

		// Dashboard stats
		if cfg.DashboardHandler != nil {
			protected.GET("/dashboard/stats", cfg.DashboardHandler.Stats)
		}

		// User routes
		users := protected.Group("/users")
		{
			users.GET("", cfg.UserHandler.ListUsers)
			users.GET("/:id", cfg.UserHandler.GetUser)
			users.PUT("/:id", cfg.UserHandler.UpdateUser)
			users.DELETE("/:id", middleware.RequireRole("admin", "super_admin"), cfg.UserHandler.DeleteUser)
			if cfg.AdminRoleHandler != nil {
				users.PATCH("/:id/role", middleware.RequireRole("admin", "super_admin"), cfg.AdminRoleHandler.UpdateUserRole)
			}
		}

		// Audit logs: admin/super_admin read-only
		if cfg.AuditHandler != nil {
			auditGroup := protected.Group("/audit-logs")
			auditGroup.Use(middleware.RequireRole("admin", "super_admin"))
			auditGroup.GET("", cfg.AuditHandler.ListAuditLogs)
		}

		// Notification REST endpoints — proxied to module-notification
		if cfg.NotificationHTTPAddr != "" {
			notifProxy := newNotificationProxy(cfg.NotificationHTTPAddr)
			notifs := protected.Group("/notifications")
			notifs.Any("", notifProxy)
			notifs.Any("/*path", notifProxy)
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

		// Module-specific routes: HR, Subject, Timetable, Student, Analytics, Import
		registerModuleRoutes(protected, cfg)
	}

	return r
}
