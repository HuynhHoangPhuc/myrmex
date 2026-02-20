package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

type ModuleRegistration struct {
	ID              uuid.UUID
	Name            string
	Version         string
	GRPCAddress     string
	HealthStatus    HealthStatus
	RegisteredAt    time.Time
	LastHealthCheck *time.Time
}

func (m *ModuleRegistration) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("module name is required")
	}
	if m.Version == "" {
		return fmt.Errorf("module version is required")
	}
	if m.GRPCAddress == "" {
		return fmt.Errorf("gRPC address is required")
	}
	return nil
}
