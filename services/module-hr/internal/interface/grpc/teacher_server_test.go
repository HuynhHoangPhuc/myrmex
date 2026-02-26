package grpc

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	hrv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/hr/v1"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/application/query"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/domain/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const hrBufSize = 1024 * 1024

func startHRTestServer(t *testing.T, register func(*grpc.Server)) *grpc.ClientConn {
	t.Helper()
	lis := bufconn.Listen(hrBufSize)
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

type mockTeacherRepository struct {
	createTeacher *entity.Teacher
	createErr     error
	getTeacher    *entity.Teacher
	getErr        error
	listSpecs     []string
	listSpecsErr  error
}

var _ repository.TeacherRepository = (*mockTeacherRepository)(nil)

func (m *mockTeacherRepository) Create(_ context.Context, teacher *entity.Teacher) (*entity.Teacher, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	if m.createTeacher != nil {
		return m.createTeacher, nil
	}
	return teacher, nil
}

func (m *mockTeacherRepository) GetByID(_ context.Context, _ uuid.UUID) (*entity.Teacher, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.getTeacher, nil
}

func (m *mockTeacherRepository) List(_ context.Context, _, _ int32) ([]*entity.Teacher, error) {
	return nil, nil
}

func (m *mockTeacherRepository) ListByDepartment(_ context.Context, _ uuid.UUID, _, _ int32) ([]*entity.Teacher, error) {
	return nil, nil
}

func (m *mockTeacherRepository) Count(_ context.Context) (int64, error) {
	return 0, nil
}

func (m *mockTeacherRepository) Update(_ context.Context, teacher *entity.Teacher) (*entity.Teacher, error) {
	return teacher, nil
}

func (m *mockTeacherRepository) Delete(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockTeacherRepository) SearchByName(_ context.Context, _ string, _, _ int32) ([]*entity.Teacher, error) {
	return nil, nil
}

func (m *mockTeacherRepository) SearchBySpecialization(_ context.Context, _ string) ([]*entity.Teacher, error) {
	return nil, nil
}

func (m *mockTeacherRepository) AddSpecialization(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}

func (m *mockTeacherRepository) RemoveSpecialization(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}

func (m *mockTeacherRepository) ListSpecializations(_ context.Context, _ uuid.UUID) ([]string, error) {
	if m.listSpecsErr != nil {
		return nil, m.listSpecsErr
	}
	return m.listSpecs, nil
}

func (m *mockTeacherRepository) UpsertAvailability(_ context.Context, _ *entity.Availability) (*entity.Availability, error) {
	return nil, nil
}

func (m *mockTeacherRepository) DeleteAvailability(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockTeacherRepository) ListAvailability(_ context.Context, _ uuid.UUID) ([]*entity.Availability, error) {
	return nil, nil
}

func TestTeacherServer_CreateTeacher_Success(t *testing.T) {
	teacherID := uuid.New()
	now := time.Now()
	repo := &mockTeacherRepository{createTeacher: &entity.Teacher{
		ID:        teacherID,
		FullName:  "Ada Lovelace",
		Email:     "ada@example.com",
		Title:     "Professor",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}}
	createHandler := command.NewCreateTeacherHandler(repo, command.NewNoopPublisher())
	conn := startHRTestServer(t, func(server *grpc.Server) {
		hrv1.RegisterTeacherServiceServer(server, NewTeacherServer(createHandler, nil, nil, nil, nil, nil, nil))
	})

	client := hrv1.NewTeacherServiceClient(conn)
	resp, err := client.CreateTeacher(context.Background(), &hrv1.CreateTeacherRequest{
		FullName: "Ada Lovelace",
		Email:    "ada@example.com",
		Title:    "Professor",
	})
	if err != nil {
		t.Fatalf("CreateTeacher error: %v", err)
	}
	if resp.Teacher.GetId() != teacherID.String() {
		t.Fatalf("unexpected teacher id: %s", resp.Teacher.GetId())
	}
	if resp.Teacher.GetFullName() != "Ada Lovelace" {
		t.Fatalf("unexpected teacher name: %s", resp.Teacher.GetFullName())
	}
}

func TestTeacherServer_CreateTeacher_InvalidArgument(t *testing.T) {
	conn := startHRTestServer(t, func(server *grpc.Server) {
		hrv1.RegisterTeacherServiceServer(server, NewTeacherServer(nil, nil, nil, nil, nil, nil, nil))
	})

	client := hrv1.NewTeacherServiceClient(conn)
	_, err := client.CreateTeacher(context.Background(), &hrv1.CreateTeacherRequest{Email: "bad@example.com"})
	statusErr, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected status error, got %v", err)
	}
	if statusErr.Code() != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %s", statusErr.Code())
	}
}

func TestTeacherServer_GetTeacher_Success(t *testing.T) {
	teacherID := uuid.New()
	now := time.Now()
	repo := &mockTeacherRepository{
		getTeacher: &entity.Teacher{
			ID:        teacherID,
			FullName:  "Grace Hopper",
			Email:     "grace@example.com",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		listSpecs: []string{"compiler"},
	}
	getHandler := query.NewGetTeacherHandler(repo)
	conn := startHRTestServer(t, func(server *grpc.Server) {
		hrv1.RegisterTeacherServiceServer(server, NewTeacherServer(nil, nil, nil, nil, getHandler, nil, nil))
	})

	client := hrv1.NewTeacherServiceClient(conn)
	resp, err := client.GetTeacher(context.Background(), &hrv1.GetTeacherRequest{Id: teacherID.String()})
	if err != nil {
		t.Fatalf("GetTeacher error: %v", err)
	}
	if resp.Teacher.GetId() != teacherID.String() {
		t.Fatalf("unexpected teacher id: %s", resp.Teacher.GetId())
	}
}

func TestTeacherServer_GetTeacher_NotFound(t *testing.T) {
	repo := &mockTeacherRepository{getErr: errors.New("teacher not found")}
	getHandler := query.NewGetTeacherHandler(repo)
	conn := startHRTestServer(t, func(server *grpc.Server) {
		hrv1.RegisterTeacherServiceServer(server, NewTeacherServer(nil, nil, nil, nil, getHandler, nil, nil))
	})

	client := hrv1.NewTeacherServiceClient(conn)
	_, err := client.GetTeacher(context.Background(), &hrv1.GetTeacherRequest{Id: uuid.New().String()})
	statusErr, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected status error, got %v", err)
	}
	if statusErr.Code() != codes.NotFound {
		t.Fatalf("expected NotFound, got %s", statusErr.Code())
	}
}
