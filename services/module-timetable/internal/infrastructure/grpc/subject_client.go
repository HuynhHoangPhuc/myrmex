package grpc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	subjectv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/subject/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// SubjectInfo holds minimal subject data needed by the CSP solver.
type SubjectInfo struct {
	ID                      uuid.UUID
	Code                    string
	Name                    string
	DepartmentID            uuid.UUID
	Credits                 int
	RequiredSpecializations []string // sourced from department mapping (empty if none)
}

// SubjectClient wraps the Subject module gRPC connection.
type SubjectClient struct {
	subject      subjectv1.SubjectServiceClient
	prerequisite subjectv1.PrerequisiteServiceClient
}

// NewSubjectClient dials the Subject gRPC server.
func NewSubjectClient(addr string) (*SubjectClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial subject service %s: %w", addr, err)
	}
	return &SubjectClient{
		subject:      subjectv1.NewSubjectServiceClient(conn),
		prerequisite: subjectv1.NewPrerequisiteServiceClient(conn),
	}, nil
}

// NewSubjectClientWithServices constructs a client from existing gRPC service clients.
func NewSubjectClientWithServices(
	subject subjectv1.SubjectServiceClient,
	prerequisite subjectv1.PrerequisiteServiceClient,
) *SubjectClient {
	return &SubjectClient{subject: subject, prerequisite: prerequisite}
}

// ListSubjectsByIDs fetches subject details for a set of subject IDs.
func (c *SubjectClient) ListSubjectsByIDs(ctx context.Context, ids []uuid.UUID) ([]SubjectInfo, error) {
	// Subject service ListSubjects returns all; we filter to the offered set.
	resp, err := c.subject.ListSubjects(ctx, &subjectv1.ListSubjectsRequest{})
	if err != nil {
		return nil, fmt.Errorf("list subjects: %w", err)
	}

	wanted := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		wanted[id] = true
	}

	var result []SubjectInfo
	for _, s := range resp.Subjects {
		id, err := uuid.Parse(s.Id)
		if err != nil {
			return nil, fmt.Errorf("parse subject id %q: %w", s.Id, err)
		}
		if !wanted[id] {
			continue
		}
		deptID, _ := uuid.Parse(s.DepartmentId)
		result = append(result, SubjectInfo{
			ID:           id,
			Code:         s.Code,
			Name:         s.Name,
			DepartmentID: deptID,
			Credits:      int(s.Credits),
		})
	}
	return result, nil
}

// TopologicalSort returns subject IDs ordered by prerequisite DAG (dependencies first).
func (c *SubjectClient) TopologicalSort(ctx context.Context) ([]uuid.UUID, error) {
	resp, err := c.prerequisite.TopologicalSort(ctx, &subjectv1.TopologicalSortRequest{})
	if err != nil {
		return nil, fmt.Errorf("topological sort: %w", err)
	}
	ids := make([]uuid.UUID, 0, len(resp.SubjectIds))
	for _, idStr := range resp.SubjectIds {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}
