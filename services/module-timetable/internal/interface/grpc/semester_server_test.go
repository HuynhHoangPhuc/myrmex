package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	hrv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/hr/v1"
	subjectv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/subject/v1"
	timetablev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/timetable/v1"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/application/query"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/entity"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/repository"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/service"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/valueobject"
	infragrpc "github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/infrastructure/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const timetableBufSize = 1024 * 1024

func startTimetableTestServer(t *testing.T, register func(*grpc.Server)) *grpc.ClientConn {
	t.Helper()
	lis := bufconn.Listen(timetableBufSize)
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

type mockSemesterRepository struct {
	created   *entity.Semester
	createErr error
	getByID   map[uuid.UUID]*entity.Semester
	getErr    error
	slots     []*entity.TimeSlot
}

var _ repository.SemesterRepository = (*mockSemesterRepository)(nil)

func (m *mockSemesterRepository) Create(_ context.Context, sem *entity.Semester) (*entity.Semester, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	if m.created != nil {
		return m.created, nil
	}
	return sem, nil
}

func (m *mockSemesterRepository) GetByID(_ context.Context, id uuid.UUID) (*entity.Semester, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if m.getByID != nil {
		if sem, ok := m.getByID[id]; ok {
			return sem, nil
		}
	}
	return nil, errors.New("semester not found")
}

func (m *mockSemesterRepository) List(_ context.Context, _, _ int32) ([]*entity.Semester, error) {
	return nil, nil
}

func (m *mockSemesterRepository) Count(_ context.Context) (int64, error) {
	return 0, nil
}

func (m *mockSemesterRepository) AddOfferedSubject(_ context.Context, _, _ uuid.UUID) (*entity.Semester, error) {
	return nil, nil
}

func (m *mockSemesterRepository) RemoveOfferedSubject(_ context.Context, _, _ uuid.UUID) (*entity.Semester, error) {
	return nil, nil
}

func (m *mockSemesterRepository) CreateTimeSlot(_ context.Context, _ *entity.TimeSlot) (*entity.TimeSlot, error) {
	return nil, nil
}

func (m *mockSemesterRepository) ListTimeSlots(_ context.Context, _ uuid.UUID) ([]*entity.TimeSlot, error) {
	if m.slots != nil {
		return m.slots, nil
	}
	return nil, nil
}

type mockScheduleRepository struct {
	byID            map[uuid.UUID]*entity.Schedule
	entries         map[uuid.UUID][]*entity.ScheduleEntry
	created         *entity.Schedule
	createdID       uuid.UUID
	createCallCount int
}

var _ repository.ScheduleRepository = (*mockScheduleRepository)(nil)

func (m *mockScheduleRepository) Create(_ context.Context, schedule *entity.Schedule) (*entity.Schedule, error) {
	m.createCallCount++
	if m.created != nil {
		return m.created, nil
	}
	if m.createdID != uuid.Nil {
		schedule.ID = m.createdID
	}
	return schedule, nil
}

func (m *mockScheduleRepository) GetByID(_ context.Context, id uuid.UUID) (*entity.Schedule, error) {
	if schedule, ok := m.byID[id]; ok {
		return schedule, nil
	}
	return nil, errors.New("schedule not found")
}

func (m *mockScheduleRepository) ListBySemester(_ context.Context, _ uuid.UUID) ([]*entity.Schedule, error) {
	return nil, nil
}

func (m *mockScheduleRepository) ListPaged(_ context.Context, _ uuid.UUID, _, _ int32) ([]*entity.Schedule, error) {
	return nil, nil
}

func (m *mockScheduleRepository) CountSchedules(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}

func (m *mockScheduleRepository) UpdateResult(_ context.Context, _ uuid.UUID, _ float64, _ int, _ float64) (*entity.Schedule, error) {
	return nil, nil
}

func (m *mockScheduleRepository) UpdateStatus(_ context.Context, _ uuid.UUID, _ valueobject.ScheduleStatus) (*entity.Schedule, error) {
	return nil, nil
}

func (m *mockScheduleRepository) CreateEntry(_ context.Context, entry *entity.ScheduleEntry) (*entity.ScheduleEntry, error) {
	return entry, nil
}

func (m *mockScheduleRepository) GetEntry(_ context.Context, _ uuid.UUID) (*entity.ScheduleEntry, error) {
	return nil, nil
}

func (m *mockScheduleRepository) ListEntries(_ context.Context, scheduleID uuid.UUID) ([]*entity.ScheduleEntry, error) {
	if m.entries != nil {
		return m.entries[scheduleID], nil
	}
	return nil, nil
}

func (m *mockScheduleRepository) UpdateEntry(_ context.Context, entry *entity.ScheduleEntry) (*entity.ScheduleEntry, error) {
	return entry, nil
}

func (m *mockScheduleRepository) DeleteEntry(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockScheduleRepository) DeleteAllEntries(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockScheduleRepository) AppendEvent(_ context.Context, _ uuid.UUID, _, _ string, _ json.RawMessage) error {
	return nil
}

func TestSemesterServer_CreateSemester_Success(t *testing.T) {
	now := time.Now()
	semesterID := uuid.New()
	repo := &mockSemesterRepository{created: &entity.Semester{
		ID:        semesterID,
		Name:      "Fall 2026",
		Year:      2026,
		Term:      1,
		StartDate: now,
		EndDate:   now.Add(24 * time.Hour),
		CreatedAt: now,
	}}
	createHandler := command.NewCreateSemesterHandler(repo)
	conn := startTimetableTestServer(t, func(server *grpc.Server) {
		timetablev1.RegisterSemesterServiceServer(server, NewSemesterServer(createHandler, nil, repo))
	})

	client := timetablev1.NewSemesterServiceClient(conn)
	resp, err := client.CreateSemester(context.Background(), &timetablev1.CreateSemesterRequest{
		Name:      "Fall 2026",
		Year:      2026,
		Term:      1,
		StartDate: timestamppbFromTime(now),
		EndDate:   timestamppbFromTime(now.Add(24 * time.Hour)),
	})
	if err != nil {
		t.Fatalf("CreateSemester error: %v", err)
	}
	if resp.Semester.GetId() != semesterID.String() {
		t.Fatalf("unexpected semester id: %s", resp.Semester.GetId())
	}
}

func TestSemesterServer_CreateSemester_InvalidArgument(t *testing.T) {
	createHandler := command.NewCreateSemesterHandler(&mockSemesterRepository{})
	conn := startTimetableTestServer(t, func(server *grpc.Server) {
		timetablev1.RegisterSemesterServiceServer(server, NewSemesterServer(createHandler, nil, &mockSemesterRepository{}))
	})

	client := timetablev1.NewSemesterServiceClient(conn)
	_, err := client.CreateSemester(context.Background(), &timetablev1.CreateSemesterRequest{})
	statusErr, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected status error, got %v", err)
	}
	if statusErr.Code() != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %s", statusErr.Code())
	}
}

func TestTimetableServer_GenerateSchedule_Success(t *testing.T) {
	scheduleID := uuid.New()
	semesterID := uuid.New()

	teacherID := uuid.New()
	subjectID := uuid.New()
	roomID := uuid.New()
	slotID := uuid.New()
	semesterRepo := &mockSemesterRepository{
		getByID: map[uuid.UUID]*entity.Semester{
			semesterID: {
				ID:                semesterID,
				Name:              "Fall 2026",
				Year:              2026,
				Term:              1,
				StartDate:         time.Now(),
				EndDate:           time.Now().Add(24 * time.Hour),
				OfferedSubjectIDs: []uuid.UUID{subjectID},
			},
		},
		slots: []*entity.TimeSlot{{
			ID:          slotID,
			SemesterID:  semesterID,
			DayOfWeek:   1,
			StartPeriod: 1,
			EndPeriod:   2,
		}},
	}
	roomRepo := &mockRoomRepository{rooms: []*entity.Room{{ID: roomID, Name: "R1", Capacity: 30, Type: "classroom", IsActive: true}}}

	scheduleRepo := &mockScheduleRepository{
		createdID: scheduleID,
		byID: map[uuid.UUID]*entity.Schedule{
			scheduleID: {ID: scheduleID, SemesterID: semesterID, Score: 100},
		},
	}

	hrClient := &mockHRClient{
		teachers: []service.TeacherInfo{{ID: teacherID, FullName: "Teacher", Specializations: []string{"general"}, MaxHoursPerWeek: 10}},
		availability: map[uuid.UUID][]*entity.TimeSlot{
			teacherID: {{ID: slotID, DayOfWeek: 1, StartPeriod: 1, EndPeriod: 2}},
		},
	}
	subjectClient := &mockSubjectClient{subjects: []infragrpc.SubjectInfo{{ID: subjectID, Code: "CS101", Credits: 1}}}
	publisher := &mockEventPublisher{}
	teacherServiceClient := &mockTeacherServiceClient{teachers: hrClient.teachers, availability: hrClient.availability}
	subjectServiceClient := &mockSubjectServiceClient{subjects: subjectClient.subjects}

	generatorHandler := command.NewGenerateScheduleHandler(
		semesterRepo,
		scheduleRepo,
		roomRepo,
		infragrpc.NewHRClientWithTeacherClient(teacherServiceClient),
		infragrpc.NewSubjectClientWithServices(subjectServiceClient, &mockPrerequisiteServiceClient{}),
		publisher,
	)
	getHandler := query.NewGetScheduleHandler(scheduleRepo)
	conn := startTimetableTestServer(t, func(server *grpc.Server) {
		timetablev1.RegisterTimetableServiceServer(server, NewTimetableServer(generatorHandler, nil, getHandler, nil, nil))
	})

	client := timetablev1.NewTimetableServiceClient(conn)
	resp, err := client.GenerateSchedule(context.Background(), &timetablev1.GenerateScheduleRequest{SemesterId: semesterID.String()})
	if err != nil {
		t.Fatalf("GenerateSchedule error: %v", err)
	}
	if resp.GetSchedule() == nil {
		t.Fatal("expected schedule in response")
	}
	if resp.GetSchedule().GetId() != scheduleID.String() {
		t.Fatalf("unexpected schedule id: got %s want %s", resp.GetSchedule().GetId(), scheduleID)
	}
	if scheduleRepo.createCallCount == 0 {
		t.Fatal("expected schedule repository create to be called")
	}
}

func timestamppbFromTime(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

type mockRoomRepository struct {
	rooms []*entity.Room
}

var _ repository.RoomRepository = (*mockRoomRepository)(nil)

func (m *mockRoomRepository) Create(_ context.Context, room *entity.Room) (*entity.Room, error) {
	return room, nil
}

func (m *mockRoomRepository) GetByID(_ context.Context, _ uuid.UUID) (*entity.Room, error) {
	return nil, context.Canceled
}

func (m *mockRoomRepository) List(_ context.Context, _, _ int32) ([]*entity.Room, error) {
	if m.rooms != nil {
		return m.rooms, nil
	}
	return nil, nil
}

func (m *mockRoomRepository) Count(_ context.Context) (int64, error) {
	return int64(len(m.rooms)), nil
}

type mockHRClient struct {
	teachers     []service.TeacherInfo
	availability map[uuid.UUID][]*entity.TimeSlot
}

type mockSubjectClient struct {
	subjects []infragrpc.SubjectInfo
}

type mockEventPublisher struct{}

type mockTeacherServiceClient struct {
	teachers     []service.TeacherInfo
	availability map[uuid.UUID][]*entity.TimeSlot
}

type mockSubjectServiceClient struct {
	subjects []infragrpc.SubjectInfo
}

type mockPrerequisiteServiceClient struct{}

var _ command.EventPublisher = (*mockEventPublisher)(nil)
var _ hrv1.TeacherServiceClient = (*mockTeacherServiceClient)(nil)
var _ subjectv1.SubjectServiceClient = (*mockSubjectServiceClient)(nil)
var _ subjectv1.PrerequisiteServiceClient = (*mockPrerequisiteServiceClient)(nil)

func (m *mockHRClient) ListTeachers(_ context.Context) ([]service.TeacherInfo, error) {
	if m.teachers != nil {
		return m.teachers, nil
	}
	return nil, nil
}

func (m *mockHRClient) GetTeacherAvailability(_ context.Context, teacherID uuid.UUID) ([]*entity.TimeSlot, error) {
	if m.availability != nil {
		if slots, ok := m.availability[teacherID]; ok {
			return slots, nil
		}
	}
	return nil, nil
}

func (m *mockTeacherServiceClient) CreateTeacher(context.Context, *hrv1.CreateTeacherRequest, ...grpc.CallOption) (*hrv1.CreateTeacherResponse, error) {
	return &hrv1.CreateTeacherResponse{}, nil
}

func (m *mockTeacherServiceClient) GetTeacher(context.Context, *hrv1.GetTeacherRequest, ...grpc.CallOption) (*hrv1.GetTeacherResponse, error) {
	return &hrv1.GetTeacherResponse{}, nil
}

func (m *mockTeacherServiceClient) ListTeachers(ctx context.Context, _ *hrv1.ListTeachersRequest, _ ...grpc.CallOption) (*hrv1.ListTeachersResponse, error) {
	teachers := make([]*hrv1.Teacher, len(m.teachers))
	for i, t := range m.teachers {
		teachers[i] = &hrv1.Teacher{Id: t.ID.String(), FullName: t.FullName, IsActive: true}
	}
	return &hrv1.ListTeachersResponse{Teachers: teachers}, nil
}

func (m *mockTeacherServiceClient) UpdateTeacher(context.Context, *hrv1.UpdateTeacherRequest, ...grpc.CallOption) (*hrv1.UpdateTeacherResponse, error) {
	return &hrv1.UpdateTeacherResponse{}, nil
}

func (m *mockTeacherServiceClient) DeleteTeacher(context.Context, *hrv1.DeleteTeacherRequest, ...grpc.CallOption) (*hrv1.DeleteTeacherResponse, error) {
	return &hrv1.DeleteTeacherResponse{}, nil
}

func (m *mockTeacherServiceClient) ListTeacherAvailability(ctx context.Context, req *hrv1.ListTeacherAvailabilityRequest, _ ...grpc.CallOption) (*hrv1.ListTeacherAvailabilityResponse, error) {
	id, err := uuid.Parse(req.GetTeacherId())
	if err != nil {
		return &hrv1.ListTeacherAvailabilityResponse{}, nil
	}
	slots := m.availability[id]
	resp := &hrv1.ListTeacherAvailabilityResponse{TeacherId: req.GetTeacherId()}
	for _, s := range slots {
		resp.AvailableSlots = append(resp.AvailableSlots, &hrv1.TimeSlot{
			DayOfWeek:   int32(s.DayOfWeek),
			StartPeriod: int32(s.StartPeriod),
			EndPeriod:   int32(s.EndPeriod),
		})
	}
	return resp, nil
}

func (m *mockTeacherServiceClient) UpdateTeacherAvailability(context.Context, *hrv1.UpdateTeacherAvailabilityRequest, ...grpc.CallOption) (*hrv1.UpdateTeacherAvailabilityResponse, error) {
	return &hrv1.UpdateTeacherAvailabilityResponse{}, nil
}

func (m *mockSubjectClient) ListSubjectsByIDs(_ context.Context, _ []uuid.UUID) ([]infragrpc.SubjectInfo, error) {
	if m.subjects != nil {
		return m.subjects, nil
	}
	return nil, nil
}

func (m *mockSubjectClient) TopologicalSort(_ context.Context) ([]uuid.UUID, error) {
	return nil, nil
}

func (m *mockSubjectServiceClient) CreateSubject(context.Context, *subjectv1.CreateSubjectRequest, ...grpc.CallOption) (*subjectv1.CreateSubjectResponse, error) {
	return &subjectv1.CreateSubjectResponse{}, nil
}

func (m *mockSubjectServiceClient) GetSubject(context.Context, *subjectv1.GetSubjectRequest, ...grpc.CallOption) (*subjectv1.GetSubjectResponse, error) {
	return &subjectv1.GetSubjectResponse{}, nil
}

func (m *mockSubjectServiceClient) ListSubjects(ctx context.Context, _ *subjectv1.ListSubjectsRequest, _ ...grpc.CallOption) (*subjectv1.ListSubjectsResponse, error) {
	subjects := make([]*subjectv1.Subject, len(m.subjects))
	for i, s := range m.subjects {
		subjects[i] = &subjectv1.Subject{Id: s.ID.String(), Code: s.Code, Credits: int32(s.Credits)}
	}
	return &subjectv1.ListSubjectsResponse{Subjects: subjects}, nil
}

func (m *mockSubjectServiceClient) UpdateSubject(context.Context, *subjectv1.UpdateSubjectRequest, ...grpc.CallOption) (*subjectv1.UpdateSubjectResponse, error) {
	return &subjectv1.UpdateSubjectResponse{}, nil
}

func (m *mockSubjectServiceClient) DeleteSubject(context.Context, *subjectv1.DeleteSubjectRequest, ...grpc.CallOption) (*subjectv1.DeleteSubjectResponse, error) {
	return &subjectv1.DeleteSubjectResponse{}, nil
}

func (m *mockPrerequisiteServiceClient) AddPrerequisite(context.Context, *subjectv1.AddPrerequisiteRequest, ...grpc.CallOption) (*subjectv1.AddPrerequisiteResponse, error) {
	return &subjectv1.AddPrerequisiteResponse{}, nil
}

func (m *mockPrerequisiteServiceClient) RemovePrerequisite(context.Context, *subjectv1.RemovePrerequisiteRequest, ...grpc.CallOption) (*subjectv1.RemovePrerequisiteResponse, error) {
	return &subjectv1.RemovePrerequisiteResponse{}, nil
}

func (m *mockPrerequisiteServiceClient) ListPrerequisites(context.Context, *subjectv1.ListPrerequisitesRequest, ...grpc.CallOption) (*subjectv1.ListPrerequisitesResponse, error) {
	return &subjectv1.ListPrerequisitesResponse{}, nil
}

func (m *mockPrerequisiteServiceClient) ValidateDAG(context.Context, *subjectv1.ValidateDAGRequest, ...grpc.CallOption) (*subjectv1.ValidateDAGResponse, error) {
	return &subjectv1.ValidateDAGResponse{}, nil
}

func (m *mockPrerequisiteServiceClient) TopologicalSort(context.Context, *subjectv1.TopologicalSortRequest, ...grpc.CallOption) (*subjectv1.TopologicalSortResponse, error) {
	return &subjectv1.TopologicalSortResponse{}, nil
}

func (m *mockEventPublisher) Publish(_ context.Context, _ string, _ any) error {
	return nil
}
