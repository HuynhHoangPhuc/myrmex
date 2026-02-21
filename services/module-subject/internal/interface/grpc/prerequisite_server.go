package grpc

import (
	"context"
	"strings"

	"github.com/google/uuid"
	subjectv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/subject/v1"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/application/query"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PrerequisiteServer implements subjectv1.PrerequisiteServiceServer.
type PrerequisiteServer struct {
	subjectv1.UnimplementedPrerequisiteServiceServer

	addHandler      *command.AddPrerequisiteHandler
	removeHandler   *command.RemovePrerequisiteHandler
	listHandler     *query.ListPrerequisitesHandler
	validateHandler *query.ValidateDAGHandler
	topoHandler     *query.TopologicalSortHandler
}

// NewPrerequisiteServer constructs a PrerequisiteServer with all required handlers.
func NewPrerequisiteServer(
	addHandler *command.AddPrerequisiteHandler,
	removeHandler *command.RemovePrerequisiteHandler,
	listHandler *query.ListPrerequisitesHandler,
	validateHandler *query.ValidateDAGHandler,
	topoHandler *query.TopologicalSortHandler,
) *PrerequisiteServer {
	return &PrerequisiteServer{
		addHandler:      addHandler,
		removeHandler:   removeHandler,
		listHandler:     listHandler,
		validateHandler: validateHandler,
		topoHandler:     topoHandler,
	}
}

// AddPrerequisite adds a prerequisite edge with DAG cycle validation.
func (s *PrerequisiteServer) AddPrerequisite(ctx context.Context, req *subjectv1.AddPrerequisiteRequest) (*subjectv1.AddPrerequisiteResponse, error) {
	subjectID, err := uuid.Parse(req.GetSubjectId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid subject_id: %v", err)
	}
	prereqID, err := uuid.Parse(req.GetPrerequisiteId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid prerequisite_id: %v", err)
	}

	prereq, err := s.addHandler.Handle(ctx, command.AddPrerequisiteCommand{
		SubjectID:      subjectID,
		PrerequisiteID: prereqID,
		Type:           "hard",
		Priority:       1,
	})
	if err != nil {
		// Surface cycle detection as FailedPrecondition.
		if strings.Contains(err.Error(), "cycle") {
			return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "add prerequisite: %v", err)
	}

	return &subjectv1.AddPrerequisiteResponse{Prerequisite: prereqToProto(prereq)}, nil
}

// RemovePrerequisite removes a prerequisite edge.
func (s *PrerequisiteServer) RemovePrerequisite(ctx context.Context, req *subjectv1.RemovePrerequisiteRequest) (*subjectv1.RemovePrerequisiteResponse, error) {
	subjectID, err := uuid.Parse(req.GetSubjectId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid subject_id: %v", err)
	}
	prereqID, err := uuid.Parse(req.GetPrerequisiteId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid prerequisite_id: %v", err)
	}

	if err := s.removeHandler.Handle(ctx, command.RemovePrerequisiteCommand{
		SubjectID:      subjectID,
		PrerequisiteID: prereqID,
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "remove prerequisite: %v", err)
	}
	return &subjectv1.RemovePrerequisiteResponse{}, nil
}

// ListPrerequisites returns all prerequisites for a given subject.
func (s *PrerequisiteServer) ListPrerequisites(ctx context.Context, req *subjectv1.ListPrerequisitesRequest) (*subjectv1.ListPrerequisitesResponse, error) {
	subjectID, err := uuid.Parse(req.GetSubjectId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid subject_id: %v", err)
	}

	prereqs, err := s.listHandler.Handle(ctx, query.ListPrerequisitesQuery{SubjectID: subjectID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list prerequisites: %v", err)
	}

	protos := make([]*subjectv1.Prerequisite, len(prereqs))
	for i, p := range prereqs {
		protos[i] = prereqToProto(p)
	}
	return &subjectv1.ListPrerequisitesResponse{Prerequisites: protos}, nil
}

// ValidateDAG runs full graph cycle detection on the prerequisite DAG.
func (s *PrerequisiteServer) ValidateDAG(ctx context.Context, _ *subjectv1.ValidateDAGRequest) (*subjectv1.ValidateDAGResponse, error) {
	result, err := s.validateHandler.Handle(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "validate dag: %v", err)
	}
	return &subjectv1.ValidateDAGResponse{
		IsValid:   result.IsValid,
		CyclePath: result.CyclePath,
	}, nil
}

// TopologicalSort returns subject IDs in prerequisite-first order.
func (s *PrerequisiteServer) TopologicalSort(ctx context.Context, _ *subjectv1.TopologicalSortRequest) (*subjectv1.TopologicalSortResponse, error) {
	ids, err := s.topoHandler.Handle(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "cycle") {
			return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "topological sort: %v", err)
	}

	strIDs := make([]string, len(ids))
	for i, id := range ids {
		strIDs[i] = id.String()
	}
	return &subjectv1.TopologicalSortResponse{SubjectIds: strIDs}, nil
}

// prereqToProto maps a domain Prerequisite entity to its protobuf representation.
func prereqToProto(p *entity.Prerequisite) *subjectv1.Prerequisite {
	return &subjectv1.Prerequisite{
		SubjectId:      p.SubjectID.String(),
		PrerequisiteId: p.PrerequisiteID.String(),
	}
}
