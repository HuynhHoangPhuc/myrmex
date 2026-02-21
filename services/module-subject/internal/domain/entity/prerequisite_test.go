package entity

import (
	"testing"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/valueobject"
)

func TestPrerequisite_Validate(t *testing.T) {
	id := uuid.New()
	other := uuid.New()

	tests := []struct {
		name    string
		p       Prerequisite
		wantErr bool
	}{
		{
			name: "valid",
			p: Prerequisite{
				SubjectID:      id,
				PrerequisiteID: other,
				Type:           valueobject.PrerequisiteTypeHard,
				Priority:       1,
			},
			wantErr: false,
		},
		{
			name: "self-loop",
			p: Prerequisite{
				SubjectID:      id,
				PrerequisiteID: id,
				Type:           valueobject.PrerequisiteTypeHard,
				Priority:       1,
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			p: Prerequisite{
				SubjectID:      id,
				PrerequisiteID: other,
				Type:           valueobject.PrerequisiteType("invalid"),
				Priority:       1,
			},
			wantErr: true,
		},
		{
			name: "priority zero",
			p: Prerequisite{
				SubjectID:      id,
				PrerequisiteID: other,
				Type:           valueobject.PrerequisiteTypeHard,
				Priority:       0,
			},
			wantErr: true,
		},
		{
			name: "priority six",
			p: Prerequisite{
				SubjectID:      id,
				PrerequisiteID: other,
				Type:           valueobject.PrerequisiteTypeHard,
				Priority:       6,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.p.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
