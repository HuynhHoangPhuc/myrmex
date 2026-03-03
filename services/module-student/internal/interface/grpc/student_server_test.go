package grpc

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	studentv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/student/v1"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/application/query"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const studentBufSize = 1024 * 1024

func startStudentTestServer(t *testing.T, register func(*grpc.Server)) *grpc.ClientConn {
	t.Helper()
	lis := bufconn.Listen(studentBufSize)
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

type mockStudentRepository struct {
	createStudent *entity.Student
	createErr     error
	getStudent    *entity.Student
	getErr        error
	listStudents  []*entity.Student
	listErr       error
	totalCount    int64
	countErr      error
	deleteErr     error
}

var _ repository.StudentRepository = (*mockStudentRepository)(nil)

func (m *mockStudentRepository) Create(_ context.Context, student *entity.Student) (*entity.Student, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	if m.createStudent != nil {
		return m.createStudent, nil
	}
	return student, nil
}

func (m *mockStudentRepository) GetByID(_ context.Context, _ uuid.UUID) (*entity.Student, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.getStudent, nil
}

func (m *mockStudentRepository) List(_ context.Context, _ *uuid.UUID, _ *string, _, _ int32) ([]*entity.Student, error) {
	return m.listStudents, m.listErr
}

func (m *mockStudentRepository) Count(_ context.Context, _ *uuid.UUID, _ *string) (int64, error) {
	return m.totalCount, m.countErr
}

func (m *mockStudentRepository) GetByUserID(_ context.Context, _ uuid.UUID) (*entity.Student, error) {
	panic("not implemented")
}

func (m *mockStudentRepository) Update(_ context.Context, student *entity.Student) (*entity.Student, error) {
	return student, nil
}

func (m *mockStudentRepository) LinkUser(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*entity.Student, error) {
	panic("not implemented")
}

func (m *mockStudentRepository) Delete(_ context.Context, _ uuid.UUID) error {
	return m.deleteErr
}

func newTestStudentServer(
	createStudent *command.CreateStudentHandler,
	updateStudent *command.UpdateStudentHandler,
	deleteStudent *command.DeleteStudentHandler,
	getStudent *query.GetStudentHandler,
	listStudents *query.ListStudentsHandler,
) *StudentServer {
	return NewStudentServer(
		createStudent,
		updateStudent,
		deleteStudent,
		nil, // linkUserToStudent
		getStudent,
		nil, // getStudentByUserID
		listStudents,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil, // createInviteCode
		nil, // validateInviteCode
		nil, // redeemInviteCode
	)
}

func TestStudentServer_CreateStudent_Success(t *testing.T) {
	studentID := uuid.New()
	departmentID := uuid.New()
	now := time.Now()
	repo := &mockStudentRepository{createStudent: &entity.Student{
		ID:             studentID,
		StudentCode:    "ST001",
		FullName:       "Ada Lovelace",
		Email:          "ada@example.com",
		DepartmentID:   departmentID,
		EnrollmentYear: 2026,
		Status:         entity.StudentStatusActive,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}}
	createHandler := command.NewCreateStudentHandler(repo, command.NewNoopPublisher())
	conn := startStudentTestServer(t, func(server *grpc.Server) {
		studentv1.RegisterStudentServiceServer(server, newTestStudentServer(createHandler, nil, nil, nil, nil))
	})

	client := studentv1.NewStudentServiceClient(conn)
	resp, err := client.CreateStudent(context.Background(), &studentv1.CreateStudentRequest{
		StudentCode:    "ST001",
		FullName:       "Ada Lovelace",
		Email:          "ada@example.com",
		DepartmentId:   departmentID.String(),
		EnrollmentYear: 2026,
	})
	if err != nil {
		t.Fatalf("CreateStudent error: %v", err)
	}
	if resp.Student.GetId() != studentID.String() {
		t.Fatalf("unexpected student id: %s", resp.Student.GetId())
	}
	if resp.Student.GetFullName() != "Ada Lovelace" {
		t.Fatalf("unexpected student name: %s", resp.Student.GetFullName())
	}
}

func TestStudentServer_CreateStudent_InvalidArgument(t *testing.T) {
	conn := startStudentTestServer(t, func(server *grpc.Server) {
		studentv1.RegisterStudentServiceServer(server, newTestStudentServer(nil, nil, nil, nil, nil))
	})

	client := studentv1.NewStudentServiceClient(conn)
	_, err := client.CreateStudent(context.Background(), &studentv1.CreateStudentRequest{StudentCode: "ST001"})
	statusErr, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected status error, got %v", err)
	}
	if statusErr.Code() != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %s", statusErr.Code())
	}
}

func TestStudentServer_GetStudent_NotFound(t *testing.T) {
	repo := &mockStudentRepository{getErr: pgx.ErrNoRows}
	getHandler := query.NewGetStudentHandler(repo)
	conn := startStudentTestServer(t, func(server *grpc.Server) {
		studentv1.RegisterStudentServiceServer(server, newTestStudentServer(nil, nil, nil, getHandler, nil))
	})

	client := studentv1.NewStudentServiceClient(conn)
	_, err := client.GetStudent(context.Background(), &studentv1.GetStudentRequest{Id: uuid.New().String()})
	statusErr, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected status error, got %v", err)
	}
	if statusErr.Code() != codes.NotFound {
		t.Fatalf("expected NotFound, got %s", statusErr.Code())
	}
}

func TestStudentServer_GetStudent_InternalError(t *testing.T) {
	repo := &mockStudentRepository{getErr: errors.New("db unavailable")}
	getHandler := query.NewGetStudentHandler(repo)
	conn := startStudentTestServer(t, func(server *grpc.Server) {
		studentv1.RegisterStudentServiceServer(server, newTestStudentServer(nil, nil, nil, getHandler, nil))
	})

	client := studentv1.NewStudentServiceClient(conn)
	_, err := client.GetStudent(context.Background(), &studentv1.GetStudentRequest{Id: uuid.New().String()})
	statusErr, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected status error, got %v", err)
	}
	if statusErr.Code() != codes.Internal {
		t.Fatalf("expected Internal, got %s", statusErr.Code())
	}
}

func TestStudentServer_DeleteStudent_NotFound(t *testing.T) {
	repo := &mockStudentRepository{deleteErr: pgx.ErrNoRows}
	deleteHandler := command.NewDeleteStudentHandler(repo, command.NewNoopPublisher())
	conn := startStudentTestServer(t, func(server *grpc.Server) {
		studentv1.RegisterStudentServiceServer(server, newTestStudentServer(nil, nil, deleteHandler, nil, nil))
	})

	client := studentv1.NewStudentServiceClient(conn)
	_, err := client.DeleteStudent(context.Background(), &studentv1.DeleteStudentRequest{Id: uuid.New().String()})
	statusErr, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected status error, got %v", err)
	}
	if statusErr.Code() != codes.NotFound {
		t.Fatalf("expected NotFound, got %s", statusErr.Code())
	}
}
