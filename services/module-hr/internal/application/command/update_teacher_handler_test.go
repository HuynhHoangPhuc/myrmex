package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestUpdateTeacherHandler(t *testing.T) {
	teacherID := uuid.New()

	tests := []struct {
		name    string
		cmd     UpdateTeacherCommand
		repoErr error
		wantErr bool
	}{
		{
			name:    "success",
			cmd:     UpdateTeacherCommand{ID: teacherID, FullName: "Updated", Email: "updated@example.com", MaxHoursPerWeek: 15, IsActive: true},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "repo error",
			cmd:     UpdateTeacherCommand{ID: teacherID, FullName: "Updated", Email: "updated@example.com", MaxHoursPerWeek: 15, IsActive: true},
			repoErr: errors.New("db"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewUpdateTeacherHandler(&mockTeacherRepo{updateErr: tt.repoErr}, NewNoopPublisher())
			_, err := handler.Handle(context.Background(), tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
