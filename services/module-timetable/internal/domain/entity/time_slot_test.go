package entity

import "testing"

func TestTimeSlotValidate(t *testing.T) {
	tests := []struct {
		name    string
		slot    *TimeSlot
		wantErr bool
	}{
		{
			name:    "invalid day low",
			slot:    &TimeSlot{DayOfWeek: -1, StartPeriod: 1, EndPeriod: 2},
			wantErr: true,
		},
		{
			name:    "invalid day high",
			slot:    &TimeSlot{DayOfWeek: 7, StartPeriod: 1, EndPeriod: 2},
			wantErr: true,
		},
		{
			name:    "start period too small",
			slot:    &TimeSlot{DayOfWeek: 1, StartPeriod: 0, EndPeriod: 2},
			wantErr: true,
		},
		{
			name:    "end before start",
			slot:    &TimeSlot{DayOfWeek: 1, StartPeriod: 3, EndPeriod: 3},
			wantErr: true,
		},
		{
			name:    "valid slot",
			slot:    &TimeSlot{DayOfWeek: 1, StartPeriod: 2, EndPeriod: 4},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.slot.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got=%v", tt.wantErr, err)
			}
		})
	}
}

func TestTimeSlotOverlapsWith(t *testing.T) {
	tests := []struct {
		name string
		a    *TimeSlot
		b    *TimeSlot
		want bool
	}{
		{
			name: "same day overlap",
			a:    &TimeSlot{DayOfWeek: 1, StartPeriod: 1, EndPeriod: 3},
			b:    &TimeSlot{DayOfWeek: 1, StartPeriod: 2, EndPeriod: 4},
			want: true,
		},
		{
			name: "same day adjacent",
			a:    &TimeSlot{DayOfWeek: 1, StartPeriod: 1, EndPeriod: 3},
			b:    &TimeSlot{DayOfWeek: 1, StartPeriod: 3, EndPeriod: 5},
			want: false,
		},
		{
			name: "different day",
			a:    &TimeSlot{DayOfWeek: 1, StartPeriod: 1, EndPeriod: 3},
			b:    &TimeSlot{DayOfWeek: 2, StartPeriod: 1, EndPeriod: 3},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.OverlapsWith(tt.b); got != tt.want {
				t.Fatalf("expected overlap=%v, got=%v", tt.want, got)
			}
		})
	}
}
