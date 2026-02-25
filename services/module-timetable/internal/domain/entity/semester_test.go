package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSemester_Validate(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		sem     Semester
		wantErr bool
	}{
		{
			name: "valid",
			sem: Semester{
				Name:      "Fall 2026",
				Year:      2026,
				Term:      1,
				StartDate: now,
				EndDate:   now.Add(24 * time.Hour),
			},
			wantErr: false,
		},
		{
			name: "missing name",
			sem: Semester{
				Name:      "",
				Year:      2026,
				Term:      1,
				StartDate: now,
				EndDate:   now.Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "year too low",
			sem: Semester{
				Name:      "Spring",
				Year:      1999,
				Term:      1,
				StartDate: now,
				EndDate:   now.Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "year too high",
			sem: Semester{
				Name:      "Spring",
				Year:      2101,
				Term:      1,
				StartDate: now,
				EndDate:   now.Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "term 0 invalid",
			sem: Semester{
				Name:      "Spring",
				Year:      2026,
				Term:      0,
				StartDate: now,
				EndDate:   now.Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "term 4 invalid",
			sem: Semester{
				Name:      "Spring",
				Year:      2026,
				Term:      4,
				StartDate: now,
				EndDate:   now.Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "end before start",
			sem: Semester{
				Name:      "Spring",
				Year:      2026,
				Term:      2,
				StartDate: now,
				EndDate:   now.Add(-1 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "end equals start",
			sem: Semester{
				Name:      "Spring",
				Year:      2026,
				Term:      2,
				StartDate: now,
				EndDate:   now,
			},
			wantErr: true,
		},
		{
			name: "term 3 valid",
			sem: Semester{
				Name:      "Summer",
				Year:      2026,
				Term:      3,
				StartDate: now,
				EndDate:   now.Add(48 * time.Hour),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sem.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSemester_AddOfferedSubject(t *testing.T) {
	s := &Semester{}
	id1 := uuid.New()
	id2 := uuid.New()

	s.AddOfferedSubject(id1)
	if len(s.OfferedSubjectIDs) != 1 {
		t.Fatalf("expected 1, got %d", len(s.OfferedSubjectIDs))
	}

	// Adding same id again should not duplicate
	s.AddOfferedSubject(id1)
	if len(s.OfferedSubjectIDs) != 1 {
		t.Fatalf("expected 1 after duplicate add, got %d", len(s.OfferedSubjectIDs))
	}

	s.AddOfferedSubject(id2)
	if len(s.OfferedSubjectIDs) != 2 {
		t.Fatalf("expected 2, got %d", len(s.OfferedSubjectIDs))
	}
}

func TestSemester_RemoveOfferedSubject(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	s := &Semester{OfferedSubjectIDs: []uuid.UUID{id1, id2}}

	s.RemoveOfferedSubject(id1)
	if len(s.OfferedSubjectIDs) != 1 {
		t.Fatalf("expected 1 after remove, got %d", len(s.OfferedSubjectIDs))
	}
	if s.OfferedSubjectIDs[0] != id2 {
		t.Fatal("expected id2 to remain")
	}

	// Removing non-existent id is a no-op
	s.RemoveOfferedSubject(uuid.New())
	if len(s.OfferedSubjectIDs) != 1 {
		t.Fatalf("expected 1 after no-op remove, got %d", len(s.OfferedSubjectIDs))
	}
}
