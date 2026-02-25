package entity

import (
	"testing"
)

func TestSubject_Validate(t *testing.T) {
	tests := []struct {
		name    string
		subject Subject
		wantErr bool
	}{
		{
			name:    "valid subject",
			subject: Subject{Code: "CS101", Name: "Intro to CS", Credits: 3, WeeklyHours: 4},
			wantErr: false,
		},
		{
			name:    "missing code",
			subject: Subject{Code: "", Name: "Intro to CS", Credits: 3},
			wantErr: true,
		},
		{
			name:    "missing name",
			subject: Subject{Code: "CS101", Name: "", Credits: 3},
			wantErr: true,
		},
		{
			name:    "negative credits",
			subject: Subject{Code: "CS101", Name: "Intro", Credits: -1},
			wantErr: true,
		},
		{
			name:    "negative weekly hours",
			subject: Subject{Code: "CS101", Name: "Intro", Credits: 3, WeeklyHours: -1},
			wantErr: true,
		},
		{
			name:    "zero credits allowed",
			subject: Subject{Code: "CS101", Name: "Intro", Credits: 0, WeeklyHours: 0},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.subject.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSubject_Deactivate(t *testing.T) {
	s := &Subject{Code: "CS101", Name: "Intro", Credits: 3, IsActive: true}
	if !s.IsActive {
		t.Fatal("expected IsActive=true before Deactivate")
	}
	s.Deactivate()
	if s.IsActive {
		t.Fatal("expected IsActive=false after Deactivate")
	}
	// Deactivate is idempotent
	s.Deactivate()
	if s.IsActive {
		t.Fatal("expected IsActive=false after second Deactivate")
	}
}
