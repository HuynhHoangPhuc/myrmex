package grpc

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/HuynhHoangPhuc/myrmex/services/module-analytics/internal/application/query"
	"github.com/HuynhHoangPhuc/myrmex/services/module-analytics/internal/infrastructure/persistence"
)

// AnalyticsHTTPServer exposes analytics query endpoints over HTTP.
// Named "grpc" package for structural consistency with other modules;
// actual transport is HTTP since no proto is defined yet.
type AnalyticsHTTPServer struct {
	workload    *query.GetWorkloadHandler
	utilization *query.GetUtilizationHandler
	dashboard   *query.GetDashboardSummaryHandler
	repo        *persistence.AnalyticsRepository
	log         *zap.Logger
}

// NewAnalyticsHTTPServer creates an AnalyticsHTTPServer wiring all query handlers.
func NewAnalyticsHTTPServer(
	workload *query.GetWorkloadHandler,
	utilization *query.GetUtilizationHandler,
	dashboard *query.GetDashboardSummaryHandler,
	repo *persistence.AnalyticsRepository,
	log *zap.Logger,
) *AnalyticsHTTPServer {
	return &AnalyticsHTTPServer{
		workload:    workload,
		utilization: utilization,
		dashboard:   dashboard,
		repo:        repo,
		log:         log,
	}
}

// RegisterRoutes attaches all HTTP handlers to the given mux.
func (s *AnalyticsHTTPServer) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /workload", s.handleWorkload)
	mux.HandleFunc("GET /utilization", s.handleUtilization)
	mux.HandleFunc("GET /dashboard", s.handleDashboard)
	mux.HandleFunc("GET /department-metrics", s.handleDepartmentMetrics)
	mux.HandleFunc("GET /schedule-metrics", s.handleScheduleMetrics)
	mux.HandleFunc("GET /schedule-heatmap", s.handleScheduleHeatmap)
}

func (s *AnalyticsHTTPServer) handleWorkload(w http.ResponseWriter, r *http.Request) {
	semesterID := parseSemesterID(r)
	stats, err := s.workload.Handle(r.Context(), query.GetWorkloadQuery{SemesterID: semesterID})
	if err != nil {
		s.log.Error("workload handler", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, stats)
}

func (s *AnalyticsHTTPServer) handleUtilization(w http.ResponseWriter, r *http.Request) {
	semesterID := parseSemesterID(r)
	stats, err := s.utilization.Handle(r.Context(), query.GetUtilizationQuery{SemesterID: semesterID})
	if err != nil {
		s.log.Error("utilization handler", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, stats)
}

func (s *AnalyticsHTTPServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	summary, err := s.dashboard.Handle(r.Context())
	if err != nil {
		s.log.Error("dashboard handler", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, summary)
}

func (s *AnalyticsHTTPServer) handleDepartmentMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := s.repo.GetDepartmentMetrics(r.Context())
	if err != nil {
		s.log.Error("department metrics handler", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, metrics)
}

func (s *AnalyticsHTTPServer) handleScheduleMetrics(w http.ResponseWriter, r *http.Request) {
	semesterID := parseSemesterID(r)
	metrics, err := s.repo.GetScheduleMetrics(r.Context(), semesterID)
	if err != nil {
		s.log.Error("schedule metrics handler", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, metrics)
}

func (s *AnalyticsHTTPServer) handleScheduleHeatmap(w http.ResponseWriter, r *http.Request) {
	semesterID := parseSemesterID(r)
	cells, err := s.repo.GetScheduleHeatmap(r.Context(), semesterID)
	if err != nil {
		s.log.Error("schedule heatmap handler", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, cells)
}

// parseSemesterID extracts and parses the semester_id query param; returns zero UUID on missing/invalid.
func parseSemesterID(r *http.Request) uuid.UUID {
	raw := r.URL.Query().Get("semester_id")
	if raw == "" {
		return uuid.Nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil
	}
	return id
}

// writeJSON serialises v to JSON and writes it to w with Content-Type application/json.
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "encode error", http.StatusInternalServerError)
	}
}
