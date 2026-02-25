package entity

import (
	"testing"
)

func TestModuleRegistration_Validate(t *testing.T) {
	tests := []struct {
		name    string
		mod     ModuleRegistration
		wantErr bool
	}{
		{
			name:    "valid",
			mod:     ModuleRegistration{Name: "hr", Version: "v1", GRPCAddress: "localhost:50051"},
			wantErr: false,
		},
		{
			name:    "missing name",
			mod:     ModuleRegistration{Name: "", Version: "v1", GRPCAddress: "localhost:50051"},
			wantErr: true,
		},
		{
			name:    "missing version",
			mod:     ModuleRegistration{Name: "hr", Version: "", GRPCAddress: "localhost:50051"},
			wantErr: true,
		},
		{
			name:    "missing grpc address",
			mod:     ModuleRegistration{Name: "hr", Version: "v1", GRPCAddress: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mod.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHealthStatus_Values(t *testing.T) {
	// Ensure the constants are the expected string values.
	if HealthStatusHealthy != "healthy" {
		t.Fatalf("unexpected HealthStatusHealthy: %s", HealthStatusHealthy)
	}
	if HealthStatusUnhealthy != "unhealthy" {
		t.Fatalf("unexpected HealthStatusUnhealthy: %s", HealthStatusUnhealthy)
	}
	if HealthStatusUnknown != "unknown" {
		t.Fatalf("unexpected HealthStatusUnknown: %s", HealthStatusUnknown)
	}
}
