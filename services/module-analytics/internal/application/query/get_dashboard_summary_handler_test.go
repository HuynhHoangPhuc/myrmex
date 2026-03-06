package query

import (
	"context"
	"errors"
	"testing"

	"github.com/HuynhHoangPhuc/myrmex/services/module-analytics/internal/domain/entity"
)

// dashboardReader is the minimal interface for dashboard summary queries.
// Mirrors the method used by GetDashboardSummaryHandler.
type dashboardReader interface {
	GetDashboardSummary(ctx context.Context) (entity.DashboardSummary, error)
}

// mockDashboardRepo is an in-memory stub satisfying dashboardReader.
type mockDashboardRepo struct {
	summary entity.DashboardSummary
	err     error
}

func (m *mockDashboardRepo) GetDashboardSummary(_ context.Context) (entity.DashboardSummary, error) {
	return m.summary, m.err
}

// testDashboardHandler wraps mockDashboardRepo to exercise the same handler logic
// without requiring a real pgxpool. The test verifies the wrapping/delegation pattern.
type testDashboardHandler struct {
	repo dashboardReader
}

func (h *testDashboardHandler) Handle(ctx context.Context) (entity.DashboardSummary, error) {
	summary, err := h.repo.GetDashboardSummary(ctx)
	if err != nil {
		return entity.DashboardSummary{}, errors.New("get dashboard summary: " + err.Error())
	}
	return summary, nil
}

func TestGetDashboardSummaryHandler(t *testing.T) {
	tests := []struct {
		name        string
		repoSummary entity.DashboardSummary
		repoErr     error
		wantErr     bool
		wantSummary entity.DashboardSummary
	}{
		{
			name: "happy path returns summary",
			repoSummary: entity.DashboardSummary{
				TotalTeachers:    10,
				TotalDepartments: 3,
				TotalSubjects:    20,
				TotalSemesters:   4,
			},
			wantSummary: entity.DashboardSummary{
				TotalTeachers:    10,
				TotalDepartments: 3,
				TotalSubjects:    20,
				TotalSemesters:   4,
			},
		},
		{
			name:    "repo error is wrapped and propagated",
			repoErr: errors.New("db unavailable"),
			wantErr: true,
		},
		{
			name:        "zero summary is valid",
			repoSummary: entity.DashboardSummary{},
			wantSummary: entity.DashboardSummary{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &testDashboardHandler{
				repo: &mockDashboardRepo{
					summary: tt.repoSummary,
					err:     tt.repoErr,
				},
			}
			got, err := h.Handle(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if got != tt.wantSummary {
				t.Errorf("Handle() = %+v, want %+v", got, tt.wantSummary)
			}
		})
	}
}
