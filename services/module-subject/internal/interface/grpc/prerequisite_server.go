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

	addHandler        *command.AddPrerequisiteHandler
	removeHandler     *command.RemovePrerequisiteHandler
	listHandler       *query.ListPrerequisitesHandler
	validateHandler   *query.ValidateDAGHandler
	topoHandler       *query.TopologicalSortHandler
	fullDAGHandler    *query.GetFullDAGHandler
	conflictsHandler  *query.CheckConflictsHandler
}

// NewPrerequisiteServer constructs a PrerequisiteServer with all required handlers.
func NewPrerequisiteServer(
	addHandler *command.AddPrerequisiteHandler,
	removeHandler *command.RemovePrerequisiteHandler,
	listHandler *query.ListPrerequisitesHandler,
	validateHandler *query.ValidateDAGHandler,
	topoHandler *query.TopologicalSortHandler,
	fullDAGHandler *query.GetFullDAGHandler,
	conflictsHandler *query.CheckConflictsHandler,
) *PrerequisiteServer {
	return &PrerequisiteServer{
		addHandler:       addHandler,
		removeHandler:    removeHandler,
		listHandler:      listHandler,
		validateHandler:  validateHandler,
		topoHandler:      topoHandler,
		fullDAGHandler:   fullDAGHandler,
		conflictsHandler: conflictsHandler,
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

// GetFullDAG returns all subjects and edges in one response.
func (s *PrerequisiteServer) GetFullDAG(ctx context.Context, _ *subjectv1.GetFullDAGRequest) (*subjectv1.GetFullDAGResponse, error) {
	result, err := s.fullDAGHandler.Handle(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get full dag: %v", err)
	}

	nodes := make([]*subjectv1.DAGNode, len(result.Subjects))
	for i, s := range result.Subjects {
		nodes[i] = subjectToDAGNode(s)
	}

	edges := make([]*subjectv1.DAGEdge, len(result.Prerequisites))
	for i, p := range result.Prerequisites {
		edges[i] = &subjectv1.DAGEdge{
			SourceId: p.PrerequisiteID.String(),
			TargetId: p.SubjectID.String(),
			Type:     p.Type.String(),
			Priority: p.Priority,
		}
	}

	return &subjectv1.GetFullDAGResponse{Nodes: nodes, Edges: edges}, nil
}

// CheckPrerequisiteConflicts identifies subjects with missing hard prerequisites.
func (s *PrerequisiteServer) CheckPrerequisiteConflicts(ctx context.Context, req *subjectv1.CheckConflictsRequest) (*subjectv1.CheckConflictsResponse, error) {
	ids := make([]uuid.UUID, 0, len(req.GetSubjectIds()))
	for _, raw := range req.GetSubjectIds() {
		id, err := uuid.Parse(raw)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid subject_id %q: %v", raw, err)
		}
		ids = append(ids, id)
	}

	conflicts, err := s.conflictsHandler.Handle(ctx, query.CheckConflictsQuery{SubjectIDs: ids})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "check conflicts: %v", err)
	}

	details := make([]*subjectv1.ConflictDetail, len(conflicts))
	for i, c := range conflicts {
		missing := make([]*subjectv1.MissingPrerequisite, len(c.Missing))
		for j, m := range c.Missing {
			missing[j] = &subjectv1.MissingPrerequisite{
				Id:   m.ID.String(),
				Name: m.Name,
				Code: m.Code,
				Type: "hard",
			}
		}
		details[i] = &subjectv1.ConflictDetail{
			SubjectId:   c.Subject.ID.String(),
			SubjectName: c.Subject.Name,
			Missing:     missing,
		}
	}
	return &subjectv1.CheckConflictsResponse{Conflicts: details}, nil
}

// prereqToProto maps a domain Prerequisite entity to its protobuf representation.
func prereqToProto(p *entity.Prerequisite) *subjectv1.Prerequisite {
	return &subjectv1.Prerequisite{
		SubjectId:      p.SubjectID.String(),
		PrerequisiteId: p.PrerequisiteID.String(),
		Type:           p.Type.String(),
		Priority:       p.Priority,
	}
}

// subjectToDAGNode maps a domain Subject entity to a DAGNode proto.
func subjectToDAGNode(s *entity.Subject) *subjectv1.DAGNode {
	return &subjectv1.DAGNode{
		Id:           s.ID.String(),
		Code:         s.Code,
		Name:         s.Name,
		Credits:      s.Credits,
		DepartmentId: s.DepartmentID,
		WeeklyHours:  s.WeeklyHours,
		IsActive:     s.IsActive,
	}
}
