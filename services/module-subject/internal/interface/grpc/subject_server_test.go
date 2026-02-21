package grpc

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	subjectv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/subject/v1"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/repository"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const subjectBufSize = 1024 * 1024

func startSubjectTestServer(t *testing.T, register func(*grpc.Server)) *grpc.ClientConn {
	t.Helper()
	lis := bufconn.Listen(subjectBufSize)
	server := grpc.NewServer()
	register(server)

	go func() {
		_ = server.Serve(lis)
	}()

	t.Cleanup(server.Stop)

	conn, err := grpc.NewClient("passthrough:///bufconn",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial bufconn: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

type mockSubjectRepository struct {
	created *entity.Subject
	createErr error
	getByID map[uuid.UUID]*entity.Subject
	getErr  error
}

var _ repository.SubjectRepository = (*mockSubjectRepository)(nil)

func (m *mockSubjectRepository) Create(_ context.Context, subject *entity.Subject) (*entity.Subject, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	if m.created != nil {
		return m.created, nil
	}
	return subject, nil
}

func (m *mockSubjectRepository) GetByID(_ context.Context, id uuid.UUID) (*entity.Subject, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if m.getByID != nil {
		if subject, ok := m.getByID[id]; ok {
			return subject, nil
		}
	}
	return nil, errors.New("subject not found")
}

func (m *mockSubjectRepository) GetByCode(_ context.Context, _ string) (*entity.Subject, error) {
	return nil, errors.New("subject not found")
}

func (m *mockSubjectRepository) List(_ context.Context, _, _ int32) ([]*entity.Subject, error) {
	return nil, nil
}

func (m *mockSubjectRepository) ListByDepartment(_ context.Context, _ string, _, _ int32) ([]*entity.Subject, error) {
	return nil, nil
}

func (m *mockSubjectRepository) Count(_ context.Context) (int64, error) {
	return 0, nil
}

func (m *mockSubjectRepository) CountByDepartment(_ context.Context, _ string) (int64, error) {
	return 0, nil
}

func (m *mockSubjectRepository) Update(_ context.Context, subject *entity.Subject) (*entity.Subject, error) {
	return subject, nil
}

func (m *mockSubjectRepository) Delete(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockSubjectRepository) ListAllIDs(_ context.Context) ([]uuid.UUID, error) {
	return nil, nil
}

type mockPrerequisiteRepository struct {
	added *entity.Prerequisite
	addErr error
	edges []*entity.Prerequisite
}

var _ repository.PrerequisiteRepository = (*mockPrerequisiteRepository)(nil)

func (m *mockPrerequisiteRepository) Add(_ context.Context, prereq *entity.Prerequisite) (*entity.Prerequisite, error) {
	if m.addErr != nil {
		return nil, m.addErr
	}
	if m.added != nil {
		return m.added, nil
	}
	return prereq, nil
}

func (m *mockPrerequisiteRepository) Remove(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (m *mockPrerequisiteRepository) ListBySubject(_ context.Context, _ uuid.UUID) ([]*entity.Prerequisite, error) {
	return nil, nil
}

func (m *mockPrerequisiteRepository) ListAll(_ context.Context) ([]*entity.Prerequisite, error) {
	return m.edges, nil
}

func (m *mockPrerequisiteRepository) Get(_ context.Context, _, _ uuid.UUID) (*entity.Prerequisite, error) {
	return nil, nil
}

func TestSubjectServer_CreateSubject_Success(t *testing.T) {
	now := time.Now()
	subjectID := uuid.New()
	repo := &mockSubjectRepository{created: &entity.Subject{
		ID:        subjectID,
		Code:      "CS101",
		Name:      "Intro to CS",
		Credits:   3,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}}
	createHandler := command.NewCreateSubjectHandler(repo)
	conn := startSubjectTestServer(t, func(server *grpc.Server) {
		subjectv1.RegisterSubjectServiceServer(server, NewSubjectServer(createHandler, nil, nil, nil, nil))
	})

	client := subjectv1.NewSubjectServiceClient(conn)
	resp, err := client.CreateSubject(context.Background(), &subjectv1.CreateSubjectRequest{
		Code:    "CS101",
		Name:    "Intro to CS",
		Credits: 3,
	})
	if err != nil {
		t.Fatalf("CreateSubject error: %v", err)
	}
	if resp.Subject.GetId() != subjectID.String() {
		t.Fatalf("unexpected subject id: %s", resp.Subject.GetId())
	}
}

func TestPrerequisiteServer_AddPrerequisite_Success(t *testing.T) {
	subjectID := uuid.New()
	prereqID := uuid.New()
	subjectRepo := &mockSubjectRepository{getByID: map[uuid.UUID]*entity.Subject{
		subjectID: {ID: subjectID, Code: "CS201", Name: "Algorithms"},
		prereqID:  {ID: prereqID, Code: "CS101", Name: "Intro to CS"},
	}}
	prereqRepo := &mockPrerequisiteRepository{}
	addHandler := command.NewAddPrerequisiteHandler(prereqRepo, subjectRepo, service.NewDAGService(prereqRepo))

	conn := startSubjectTestServer(t, func(server *grpc.Server) {
		subjectv1.RegisterPrerequisiteServiceServer(server, NewPrerequisiteServer(addHandler, nil, nil, nil, nil))
	})

	client := subjectv1.NewPrerequisiteServiceClient(conn)
	resp, err := client.AddPrerequisite(context.Background(), &subjectv1.AddPrerequisiteRequest{
		SubjectId:      subjectID.String(),
		PrerequisiteId: prereqID.String(),
	})
	if err != nil {
		t.Fatalf("AddPrerequisite error: %v", err)
	}
	if resp.Prerequisite.GetSubjectId() != subjectID.String() {
		t.Fatalf("unexpected subject id: %s", resp.Prerequisite.GetSubjectId())
	}
}

func TestPrerequisiteServer_AddPrerequisite_Cycle(t *testing.T) {
	first := uuid.New()
	second := uuid.New()
	subjectRepo := &mockSubjectRepository{getByID: map[uuid.UUID]*entity.Subject{
		first:  {ID: first, Code: "CS201", Name: "Algorithms"},
		second: {ID: second, Code: "CS101", Name: "Intro to CS"},
	}}
	prereqRepo := &mockPrerequisiteRepository{edges: []*entity.Prerequisite{{
		SubjectID:      first,
		PrerequisiteID: second,
	}}}
	addHandler := command.NewAddPrerequisiteHandler(prereqRepo, subjectRepo, service.NewDAGService(prereqRepo))

	conn := startSubjectTestServer(t, func(server *grpc.Server) {
		subjectv1.RegisterPrerequisiteServiceServer(server, NewPrerequisiteServer(addHandler, nil, nil, nil, nil))
	})

	client := subjectv1.NewPrerequisiteServiceClient(conn)
	_, err := client.AddPrerequisite(context.Background(), &subjectv1.AddPrerequisiteRequest{
		SubjectId:      second.String(),
		PrerequisiteId: first.String(),
	})
	statusErr, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected status error, got %v", err)
	}
	if statusErr.Code() != codes.FailedPrecondition {
		t.Fatalf("expected FailedPrecondition, got %s", statusErr.Code())
	}
}
