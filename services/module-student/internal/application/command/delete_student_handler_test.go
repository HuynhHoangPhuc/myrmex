package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestDeleteStudentHandler(t *testing.T) {
	studentID := uuid.New()

	tests := []struct {
		name    string
		cmd     DeleteStudentCommand
		repoErr error
		wantErr bool
	}{
		{name: "success", cmd: DeleteStudentCommand{ID: studentID}},
		{name: "repo error", cmd: DeleteStudentCommand{ID: studentID}, repoErr: errors.New("db"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDeleteStudentHandler(&mockStudentRepo{deleteErr: tt.repoErr}, NewNoopPublisher())
			err := handler.Handle(context.Background(), tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
