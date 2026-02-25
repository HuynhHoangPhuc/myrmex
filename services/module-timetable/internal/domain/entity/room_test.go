package entity

import (
	"testing"
)

func TestRoom_Validate(t *testing.T) {
	tests := []struct {
		name     string
		room     Room
		wantErr  bool
		wantType string
	}{
		{
			name:     "valid with type",
			room:     Room{Name: "Lab 101", Capacity: 30, Type: "lab"},
			wantErr:  false,
			wantType: "lab",
		},
		{
			name:     "valid without type defaults to classroom",
			room:     Room{Name: "Room A", Capacity: 40, Type: ""},
			wantErr:  false,
			wantType: "classroom",
		},
		{
			name:    "missing name",
			room:    Room{Name: "", Capacity: 30, Type: "classroom"},
			wantErr: true,
		},
		{
			name:    "zero capacity",
			room:    Room{Name: "Room B", Capacity: 0, Type: "classroom"},
			wantErr: true,
		},
		{
			name:    "negative capacity",
			room:    Room{Name: "Room C", Capacity: -5, Type: "classroom"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.room.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.room.Type != tt.wantType {
				t.Fatalf("Type = %q, want %q", tt.room.Type, tt.wantType)
			}
		})
	}
}
