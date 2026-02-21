package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestDeleteTeacherHandler(t *testing.T) {
	teacherID := uuid.New()

	tests := []struct {
		name    string
		cmd     DeleteTeacherCommand
		repoErr error
		wantErr bool
	}{
		{
			name:    "success",
			cmd:     DeleteTeacherCommand{ID: teacherID},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "repo error",
			cmd:     DeleteTeacherCommand{ID: teacherID},
			repoErr: errors.New("db"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDeleteTeacherHandler(&mockTeacherRepo{deleteErr: tt.repoErr})
			err := handler.Handle(context.Background(), tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
