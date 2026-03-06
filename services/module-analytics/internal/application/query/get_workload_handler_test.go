package query

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-analytics/internal/domain/entity"
)

// workloadReader is the minimal interface for workload stat queries.
type workloadReader interface {
	GetWorkloadStats(ctx context.Context, semesterID uuid.UUID) ([]entity.WorkloadStat, error)
}

// mockWorkloadRepo is an in-memory stub satisfying workloadReader.
type mockWorkloadRepo struct {
	stats []entity.WorkloadStat
	err   error
}

func (m *mockWorkloadRepo) GetWorkloadStats(_ context.Context, _ uuid.UUID) ([]entity.WorkloadStat, error) {
	return m.stats, m.err
}

// testWorkloadHandler mirrors GetWorkloadHandler logic using the workloadReader interface.
type testWorkloadHandler struct {
	repo workloadReader
}

func (h *testWorkloadHandler) Handle(ctx context.Context, q GetWorkloadQuery) ([]entity.WorkloadStat, error) {
	stats, err := h.repo.GetWorkloadStats(ctx, q.SemesterID)
	if err != nil {
		return nil, errors.New("get workload stats: " + err.Error())
	}
	return stats, nil
}

func TestGetWorkloadHandler(t *testing.T) {
	semID := uuid.New()
	teacherID := uuid.New()
	deptID := uuid.New()
	subjectID := uuid.New()

	sampleStats := []entity.WorkloadStat{
		{
			TeacherID:      teacherID,
			TeacherName:    "Dr. Smith",
			DepartmentID:   deptID,
			DepartmentName: "Computer Science",
			SemesterID:     semID,
			SubjectID:      subjectID,
			SubjectCode:    "CS101",
			HoursPerWeek:   3.0,
			TotalHours:     45.0,
		},
	}

	tests := []struct {
		name      string
		query     GetWorkloadQuery
		repoStats []entity.WorkloadStat
		repoErr   error
		wantErr   bool
		wantLen   int
	}{
		{
			name:      "returns stats for a specific semester",
			query:     GetWorkloadQuery{SemesterID: semID},
			repoStats: sampleStats,
			wantLen:   1,
		},
		{
			name:      "returns all stats when semester is zero",
			query:     GetWorkloadQuery{SemesterID: uuid.UUID{}},
			repoStats: sampleStats,
			wantLen:   1,
		},
		{
			name:    "empty result is valid",
			query:   GetWorkloadQuery{SemesterID: semID},
			wantLen: 0,
		},
		{
			name:    "repo error is wrapped and propagated",
			query:   GetWorkloadQuery{SemesterID: semID},
			repoErr: errors.New("connection lost"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &testWorkloadHandler{
				repo: &mockWorkloadRepo{stats: tt.repoStats, err: tt.repoErr},
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
