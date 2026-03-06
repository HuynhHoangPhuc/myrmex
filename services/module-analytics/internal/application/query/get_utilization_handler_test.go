package query

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-analytics/internal/domain/entity"
)

// utilizationReader is the minimal interface for utilization stat queries.
type utilizationReader interface {
	GetUtilizationStats(ctx context.Context, semesterID uuid.UUID) ([]entity.UtilizationStat, error)
}

// mockUtilizationRepo is an in-memory stub satisfying utilizationReader.
type mockUtilizationRepo struct {
	stats []entity.UtilizationStat
	err   error
}

func (m *mockUtilizationRepo) GetUtilizationStats(_ context.Context, _ uuid.UUID) ([]entity.UtilizationStat, error) {
	return m.stats, m.err
}

// testUtilizationHandler mirrors GetUtilizationHandler logic using the utilizationReader interface.
type testUtilizationHandler struct {
	repo utilizationReader
}

func (h *testUtilizationHandler) Handle(ctx context.Context, q GetUtilizationQuery) ([]entity.UtilizationStat, error) {
	stats, err := h.repo.GetUtilizationStats(ctx, q.SemesterID)
	if err != nil {
		return nil, errors.New("get utilization stats: " + err.Error())
	}
	return stats, nil
}

func TestGetUtilizationHandler(t *testing.T) {
	semID := uuid.New()
	deptID := uuid.New()

	sampleStats := []entity.UtilizationStat{
		{
			DepartmentID:   deptID,
			DepartmentName: "Computer Science",
			SemesterID:     semID,
			AssignedSlots:  80,
			TotalSlots:     100,
			UtilizationPct: 80.0,
		},
	}

	tests := []struct {
		name      string
		query     GetUtilizationQuery
		repoStats []entity.UtilizationStat
		repoErr   error
		wantErr   bool
		wantLen   int
	}{
		{
			name:      "returns stats for a specific semester",
			query:     GetUtilizationQuery{SemesterID: semID},
			repoStats: sampleStats,
			wantLen:   1,
		},
		{
			name:      "returns all stats when semester is zero",
			query:     GetUtilizationQuery{SemesterID: uuid.UUID{}},
			repoStats: sampleStats,
			wantLen:   1,
		},
		{
			name:    "empty result is valid",
			query:   GetUtilizationQuery{SemesterID: semID},
			wantLen: 0,
		},
		{
			name:    "repo error is wrapped and propagated",
			query:   GetUtilizationQuery{SemesterID: semID},
			repoErr: errors.New("query timeout"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &testUtilizationHandler{
				repo: &mockUtilizationRepo{stats: tt.repoStats, err: tt.repoErr},
			}
			got, err := h.Handle(context.Background(), tt.query)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("Handle() returned %d stats, want %d", len(got), tt.wantLen)
			}
		})
	}
}
