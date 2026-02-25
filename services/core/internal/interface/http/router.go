package http

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/auth"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/interface/middleware"
	"go.uber.org/zap"
)

type RouterConfig struct {
	AuthHandler      *AuthHandler
	UserHandler      *UserHandler
	ModuleHandler    *ModuleHandler
	GatewayProxy     *GatewayProxy
	ChatHandler      *ChatHandler
	HRHandler        *HRHandler
	SubjectHandler   *SubjectHandler
	TimetableHandler *TimetableHandler
	JWTService       *auth.JWTService
	Logger           *zap.Logger
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

		// HR module routes
		if cfg.HRHandler != nil {
			hr := protected.Group("/hr")
			{
				hr.GET("/teachers", cfg.HRHandler.ListTeachers)
				hr.POST("/teachers", cfg.HRHandler.CreateTeacher)
				hr.GET("/teachers/:id", cfg.HRHandler.GetTeacher)
				hr.PATCH("/teachers/:id", cfg.HRHandler.UpdateTeacher)
				hr.DELETE("/teachers/:id", cfg.HRHandler.DeleteTeacher)
				hr.GET("/teachers/:id/availability", cfg.HRHandler.GetTeacherAvailability)
				hr.PUT("/teachers/:id/availability", cfg.HRHandler.UpdateTeacherAvailability)
				hr.GET("/departments", cfg.HRHandler.ListDepartments)
				hr.POST("/departments", cfg.HRHandler.CreateDepartment)
			}
		}

		// Subject module routes
		if cfg.SubjectHandler != nil {
			subjects := protected.Group("/subjects")
			{
				// Static routes before parameterized to avoid conflicts
				subjects.GET("/dag/validate", cfg.SubjectHandler.ValidateDAG)
				subjects.GET("/dag/topological-sort", cfg.SubjectHandler.TopologicalSort)
				subjects.GET("", cfg.SubjectHandler.ListSubjects)
				subjects.POST("", cfg.SubjectHandler.CreateSubject)
				subjects.GET("/:id", cfg.SubjectHandler.GetSubject)
				subjects.PATCH("/:id", cfg.SubjectHandler.UpdateSubject)
				subjects.DELETE("/:id", cfg.SubjectHandler.DeleteSubject)
				subjects.GET("/:id/prerequisites", cfg.SubjectHandler.ListPrerequisites)
				subjects.POST("/:id/prerequisites", cfg.SubjectHandler.AddPrerequisite)
				subjects.DELETE("/:id/prerequisites/:prereqId", cfg.SubjectHandler.RemovePrerequisite)
			}
		}

		// Timetable module routes
		if cfg.TimetableHandler != nil {
			tt := protected.Group("/timetable")
			{
				tt.GET("/semesters", cfg.TimetableHandler.ListSemesters)
				tt.POST("/semesters", cfg.TimetableHandler.CreateSemester)
				tt.GET("/semesters/:id", cfg.TimetableHandler.GetSemester)
				tt.POST("/semesters/:id/offered-subjects", cfg.TimetableHandler.AddOfferedSubject)
				tt.DELETE("/semesters/:id/offered-subjects/:subjectId", cfg.TimetableHandler.RemoveOfferedSubject)
				tt.POST("/semesters/:id/generate", cfg.TimetableHandler.GenerateSchedule)
				tt.GET("/schedules", cfg.TimetableHandler.ListSchedules)
				tt.GET("/schedules/:id", cfg.TimetableHandler.GetSchedule)
				tt.PUT("/schedules/:id/entries/:entryId", cfg.TimetableHandler.ManualAssign)
				tt.GET("/suggest-teachers", cfg.TimetableHandler.SuggestTeachers)
				tt.GET("/schedules/:id/stream", cfg.TimetableHandler.StreamScheduleStatus)
			}
		}
	}

	return r
}
