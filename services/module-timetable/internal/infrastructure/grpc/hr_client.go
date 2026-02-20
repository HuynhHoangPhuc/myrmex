package grpc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	hrv1 "github.com/myrmex-erp/myrmex/gen/go/hr/v1"
	"github.com/myrmex-erp/myrmex/services/module-timetable/internal/domain/entity"
	"github.com/myrmex-erp/myrmex/services/module-timetable/internal/domain/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// HRClient wraps the HR module gRPC connection and exposes domain-typed methods.
type HRClient struct {
	teacher hrv1.TeacherServiceClient
}

// NewHRClient dials the HR gRPC server and returns a ready client.
func NewHRClient(addr string) (*HRClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial hr service %s: %w", addr, err)
	}
	return &HRClient{teacher: hrv1.NewTeacherServiceClient(conn)}, nil
}

// ListTeachers fetches all active teachers from the HR module.
func (c *HRClient) ListTeachers(ctx context.Context) ([]service.TeacherInfo, error) {
	resp, err := c.teacher.ListTeachers(ctx, &hrv1.ListTeachersRequest{})
	if err != nil {
		return nil, fmt.Errorf("list teachers from HR: %w", err)
	}
	result := make([]service.TeacherInfo, 0, len(resp.Teachers))
	for _, t := range resp.Teachers {
		if !t.IsActive {
			continue
		}
		id, _ := uuid.Parse(t.Id)
		result = append(result, service.TeacherInfo{
			ID:              id,
			FullName:        t.FullName,
			MaxHoursPerWeek: 20, // default; specializations fetched separately
		})
	}
	return result, nil
}

// GetTeacherAvailability fetches availability slots for a single teacher.
func (c *HRClient) GetTeacherAvailability(ctx context.Context, teacherID uuid.UUID) ([]*entity.TimeSlot, error) {
	resp, err := c.teacher.ListTeacherAvailability(ctx, &hrv1.ListTeacherAvailabilityRequest{
		TeacherId: teacherID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("get availability for teacher %s: %w", teacherID, err)
	}
	slots := make([]*entity.TimeSlot, 0, len(resp.AvailableSlots))
	for _, s := range resp.AvailableSlots {
		slots = append(slots, &entity.TimeSlot{
			DayOfWeek:   int(s.DayOfWeek),
			StartPeriod: int(s.StartPeriod),
			EndPeriod:   int(s.EndPeriod),
		})
	}
	return slots, nil
}
