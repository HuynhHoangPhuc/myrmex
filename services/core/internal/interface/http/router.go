package http

import (
	"fmt"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/HuynhHoangPhuc/myrmex/pkg/messaging"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/auth"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/interface/middleware"
	"go.uber.org/zap"
)

type RouterConfig struct {
	AuthHandler             *AuthHandler
	OAuthHandler            *OAuthHandler
	AdminRoleHandler        *AdminRoleHandler
	AuditHandler            *AuditHandler
	AuditPublisher          messaging.Publisher // nil → audit middleware disabled
	UserHandler             *UserHandler
	ModuleHandler           *ModuleHandler
	GatewayProxy            *GatewayProxy
	ChatHandler             *ChatHandler
	NotificationHandler     *NotificationHandler
	HRHandler               *HRHandler
	SubjectHandler          *SubjectHandler
	TimetableHandler        *TimetableHandler
	StudentHandler          *StudentHandler
	StudentPortalHandler    *StudentPortalHandler
	DashboardHandler        *DashboardHandler
	ImportHandler           *ImportHandler
	AnalyticsHTTPAddr       string // e.g. "http://module-analytics:8055"
	NotificationHTTPAddr    string // e.g. "http://module-notification:8056"
	JWTService              *auth.JWTService
	Logger                  *zap.Logger
	DB                      *pgxpool.Pool  // for health check
	Rdb                     *redis.Client  // for health check
}

func NewRouter(cfg RouterConfig) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
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
			// Role management: admin/super_admin only
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

		// HR module routes — dept_head/teacher require dept scope
		if cfg.HRHandler != nil {
			hr := protected.Group("/hr")
			hr.Use(middleware.RequireRole("admin", "super_admin", "dean", "dept_head", "manager"))
			hr.Use(middleware.RequireDeptScope())
			{
				hr.GET("/teachers", cfg.HRHandler.ListTeachers)
				hr.POST("/teachers", middleware.RequireRole("admin", "super_admin", "dept_head"), cfg.HRHandler.CreateTeacher)
				hr.GET("/teachers/:id", cfg.HRHandler.GetTeacher)
				hr.PATCH("/teachers/:id", middleware.RequireRole("admin", "super_admin", "dept_head"), cfg.HRHandler.UpdateTeacher)
				hr.DELETE("/teachers/:id", middleware.RequireRole("admin", "super_admin"), cfg.HRHandler.DeleteTeacher)
				hr.GET("/teachers/:id/availability", cfg.HRHandler.GetTeacherAvailability)
				hr.PUT("/teachers/:id/availability", middleware.RequireRole("admin", "super_admin", "dept_head"), cfg.HRHandler.UpdateTeacherAvailability)
				hr.GET("/departments", cfg.HRHandler.ListDepartments)
				hr.POST("/departments", middleware.RequireRole("admin", "super_admin"), cfg.HRHandler.CreateDepartment)
			}
		}

		// Subject module routes
		if cfg.SubjectHandler != nil {
			subjects := protected.Group("/subjects")
			{
				// Static /dag/* routes MUST come before /:id to avoid param capture
				subjects.GET("/dag/validate", cfg.SubjectHandler.ValidateDAG)
				subjects.GET("/dag/topological-sort", cfg.SubjectHandler.TopologicalSort)
				subjects.GET("/dag/full", cfg.SubjectHandler.FullDAG)
				subjects.POST("/dag/check-conflicts", cfg.SubjectHandler.CheckConflicts)
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
				tt.GET("/rooms", cfg.TimetableHandler.ListRooms)
				tt.GET("/semesters", cfg.TimetableHandler.ListSemesters)
				tt.POST("/semesters", cfg.TimetableHandler.CreateSemester)
				tt.GET("/semesters/:id", cfg.TimetableHandler.GetSemester)
				tt.PUT("/semesters/:id/rooms", cfg.TimetableHandler.SetSemesterRooms)
				tt.POST("/semesters/:id/slots", cfg.TimetableHandler.CreateTimeSlot)
				tt.DELETE("/semesters/:id/slots/:slotId", cfg.TimetableHandler.DeleteTimeSlot)
				tt.POST("/semesters/:id/slots/preset", cfg.TimetableHandler.ApplyTimeSlotPreset)
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

		// Student module routes
		if cfg.StudentHandler != nil {
			students := protected.Group("/students")
			students.Use(middleware.RequireRole("admin"))
			{
				students.GET("", cfg.StudentHandler.ListStudents)
				students.POST("", cfg.StudentHandler.CreateStudent)
				students.GET("/:id", cfg.StudentHandler.GetStudent)
				students.PATCH("/:id", cfg.StudentHandler.UpdateStudent)
				students.DELETE("/:id", cfg.StudentHandler.DeleteStudent)
				students.GET("/:id/transcript", cfg.StudentHandler.GetStudentTranscript)
				students.POST("/:id/invite-code", cfg.StudentHandler.GenerateInviteCode)
			}

			enrollments := protected.Group("/enrollments")
			enrollments.Use(middleware.RequireRole("admin"))
			{
				enrollments.GET("", cfg.StudentHandler.ListEnrollments)
				enrollments.PATCH("/:id/review", cfg.StudentHandler.ReviewEnrollment)
			}

			grades := protected.Group("/grades")
			grades.Use(middleware.RequireRole("admin", "teacher"))
			{
				grades.POST("", cfg.StudentHandler.AssignGrade)
				grades.PATCH("/:id", cfg.StudentHandler.UpdateGrade)
			}

			// Student self-service portal routes (Phase 2)
			if cfg.StudentPortalHandler != nil {
				portal := protected.Group("/student")
				portal.Use(middleware.RequireRole("student"))
				portal.Use(cfg.StudentPortalHandler.ResolveStudentMiddleware())
				{
					portal.GET("/me", cfg.StudentPortalHandler.GetMyProfile)
					portal.GET("/enrollments", cfg.StudentPortalHandler.ListMyEnrollments)
					portal.POST("/enrollments", cfg.StudentPortalHandler.RequestMyEnrollment)
					portal.GET("/enrollments/check-prerequisites", cfg.StudentPortalHandler.CheckMyPrerequisites)
					portal.GET("/transcript", cfg.StudentPortalHandler.GetMyTranscript)
					portal.GET("/transcript/export", cfg.StudentPortalHandler.ExportTranscript)
				}
			}
		}

		// Analytics module routes (reverse proxy to module-analytics HTTP)
		if cfg.AnalyticsHTTPAddr != "" {
			analytics := protected.Group("/analytics")
			analyticsProxy := newAnalyticsProxy(cfg.AnalyticsHTTPAddr)
			{
				analytics.GET("/workload", analyticsProxy)
				analytics.GET("/utilization", analyticsProxy)
				analytics.GET("/dashboard", analyticsProxy)
				analytics.GET("/department-metrics", analyticsProxy)
				analytics.GET("/schedule-metrics", analyticsProxy)
				analytics.GET("/schedule-heatmap", analyticsProxy)
				analytics.GET("/export", analyticsProxy)
			}
		}

		// Bulk import routes (admin/super_admin only)
		if cfg.ImportHandler != nil {
			imports := protected.Group("/admin/import")
			imports.Use(middleware.RequireRole("admin", "super_admin"))
			{
				imports.POST("/teachers", cfg.ImportHandler.ImportTeachers)
				imports.POST("/students", cfg.ImportHandler.ImportStudents)
				imports.GET("/template/teachers", cfg.ImportHandler.TeacherTemplate)
				imports.GET("/template/students", cfg.ImportHandler.StudentTemplate)
			}
		}
	}

	return r
}

// healthHandler checks database and Redis connectivity and returns aggregate health status.
func healthHandler(db *pgxpool.Pool, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		checks := gin.H{}
		allOk := true

		// DB check
		if db != nil {
			if err := db.Ping(c.Request.Context()); err != nil {
				checks["database"] = "error: " + err.Error()
				allOk = false
			} else {
				checks["database"] = "ok"
			}
		} else {
			checks["database"] = "unconfigured"
		}

		// Redis check
		if rdb != nil {
			if err := rdb.Ping(c.Request.Context()).Err(); err != nil {
				checks["redis"] = "error: " + err.Error()
				allOk = false
			} else {
				checks["redis"] = "ok"
			}
		} else {
			checks["redis"] = "unconfigured"
		}

		status := "ok"
		code := 200
		if !allOk {
			status = "degraded"
			code = 503
		}
		c.JSON(code, gin.H{"status": status, "checks": checks})
	}
}

// newNotificationProxy creates a gin handler that reverse-proxies to module-notification.
// It injects X-User-ID and X-User-Role headers from the validated JWT claims in the gin context.
func newNotificationProxy(targetAddr string) gin.HandlerFunc {
	target, err := url.Parse(targetAddr)
	if err != nil {
		panic(fmt.Sprintf("invalid notification target URL %q: %v", targetAddr, err))
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	return func(c *gin.Context) {
		// Inject user identity headers so module-notification can identify the caller
		c.Request.Header.Set("X-User-ID", c.GetString("user_id"))
		c.Request.Header.Set("X-User-Role", c.GetString("user_role"))
		c.Request.Header.Set("X-Department-ID", c.GetString("department_id"))

		// Rewrite path: /api/notifications/... → /notifications/...
		c.Request.URL.Path = "/notifications" + c.Param("path")
		c.Request.Host = target.Host
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// newAnalyticsProxy creates a gin handler that reverse-proxies to the analytics HTTP service.
func newAnalyticsProxy(targetAddr string) gin.HandlerFunc {
	target, err := url.Parse(targetAddr)
	if err != nil {
		panic(fmt.Sprintf("invalid analytics target URL %q: %v", targetAddr, err))
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	return func(c *gin.Context) {
		// Rewrite path: /api/analytics/workload → /workload
		c.Request.URL.Path = c.Request.URL.Path[len("/api/analytics"):]
		if c.Request.URL.Path == "" {
			c.Request.URL.Path = "/"
		}
		c.Request.Host = target.Host
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
