package persistence

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/infrastructure/persistence/sqlc"
)

func TestStudentRowToEntity(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	deptID := uuid.New()
	now := time.Now()

	student := studentRowToEntity(sqlc.StudentStudent{
		ID:             pgtype.UUID{Bytes: id, Valid: true},
		StudentCode:    "ST001",
		UserID:         pgtype.UUID{Bytes: userID, Valid: true},
		FullName:       "Ada Lovelace",
		Email:          "ada@example.com",
		DepartmentID:   pgtype.UUID{Bytes: deptID, Valid: true},
		EnrollmentYear: 2026,
		Status:         "active",
		IsActive:       true,
		CreatedAt:      pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:      pgtype.Timestamptz{Time: now, Valid: true},
	})

	if student.ID != id {
		t.Fatalf("ID = %s, want %s", student.ID, id)
	}
	if student.UserID == nil || *student.UserID != userID {
		t.Fatalf("unexpected user id")
	}
	if student.DepartmentID != deptID {
		t.Fatalf("DepartmentID = %s, want %s", student.DepartmentID, deptID)
	}
}
