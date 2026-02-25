package entity

import (
	"testing"
)

func TestDepartment_Validate(t *testing.T) {
	tests := []struct {
		name    string
		dept    Department
		wantErr bool
	}{
		{
			name:    "valid",
			dept:    Department{Name: "Computer Science", Code: "CS"},
			wantErr: false,
		},
		{
			name:    "missing name",
			dept:    Department{Name: "", Code: "CS"},
			wantErr: true,
		},
		{
			name:    "missing code",
			dept:    Department{Name: "Computer Science", Code: ""},
			wantErr: true,
		},
		{
			name:    "both empty",
			dept:    Department{Name: "", Code: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dept.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
