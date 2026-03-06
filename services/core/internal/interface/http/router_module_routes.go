package http

import (
	"github.com/gin-gonic/gin"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/interface/middleware"
)

// registerModuleRoutes registers HR, Subject, Timetable, Student (admin + portal),
// Analytics, and Import routes onto the given authenticated router group.
func registerModuleRoutes(protected *gin.RouterGroup, cfg RouterConfig) {
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

	// Student module routes (admin) + portal (student self-service)
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
