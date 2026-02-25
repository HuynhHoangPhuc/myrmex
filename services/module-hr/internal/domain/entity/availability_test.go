package entity

import (
	"testing"
)

func TestAvailability_Validate(t *testing.T) {
	tests := []struct {
		name    string
		slot    Availability
		wantErr bool
	}{
		{
			name:    "valid slot monday",
			slot:    Availability{DayOfWeek: 0, StartPeriod: 1, EndPeriod: 3},
			wantErr: false,
		},
		{
			name:    "valid slot sunday",
			slot:    Availability{DayOfWeek: 6, StartPeriod: 5, EndPeriod: 8},
			wantErr: false,
		},
		{
			name:    "day too low",
			slot:    Availability{DayOfWeek: -1, StartPeriod: 1, EndPeriod: 2},
			wantErr: true,
		},
		{
			name:    "day too high",
			slot:    Availability{DayOfWeek: 7, StartPeriod: 1, EndPeriod: 2},
			wantErr: true,
		},
		{
			name:    "start equals end",
			slot:    Availability{DayOfWeek: 1, StartPeriod: 3, EndPeriod: 3},
			wantErr: true,
		},
		{
			name:    "start after end",
			slot:    Availability{DayOfWeek: 2, StartPeriod: 5, EndPeriod: 2},
			wantErr: true,
		},
		{
			name:    "zero start zero end",
			slot:    Availability{DayOfWeek: 0, StartPeriod: 0, EndPeriod: 0},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.slot.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
