package main

import (
	"github.com/nats-io/nats.go/jetstream"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	hrv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/hr/v1"
	subjectv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/subject/v1"
	timetablev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/timetable/v1"
	httpif "github.com/HuynhHoangPhuc/myrmex/services/core/internal/interface/http"
)

// moduleHandlers holds gateway handlers and gRPC connections for lifecycle management.
type moduleHandlers struct {
	HR        *httpif.HRHandler
	Subject   *httpif.SubjectHandler
	Timetable *httpif.TimetableHandler
	Dashboard *httpif.DashboardHandler
	conns     []*grpc.ClientConn
}

// Close releases all gRPC client connections.
func (m *moduleHandlers) Close() {
	for _, c := range m.conns {
		c.Close()
	}
}

// buildModuleHandlers creates gRPC client connections and returns typed handlers.
// Nil-safe: returns nil handler if connection fails or addr not configured.
// Caller must defer Close() to release connections on shutdown.
func buildModuleHandlers(v *viper.Viper, js jetstream.JetStream, log *zap.Logger) moduleHandlers {
	var h moduleHandlers

	// Track individual gRPC clients for DashboardHandler aggregation.
	var (
		teacherClient    hrv1.TeacherServiceClient
		departmentClient hrv1.DepartmentServiceClient
		subjectClient    subjectv1.SubjectServiceClient
		semesterClient   timetablev1.SemesterServiceClient
	)

	if addr := v.GetString("hr.grpc_addr"); addr != "" {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Warn("hr grpc client failed", zap.Error(err))
		} else {
			h.conns = append(h.conns, conn)
			teacherClient = hrv1.NewTeacherServiceClient(conn)
			departmentClient = hrv1.NewDepartmentServiceClient(conn)
			h.HR = httpif.NewHRHandler(teacherClient, departmentClient)
			log.Info("hr handler ready", zap.String("addr", addr))
		}
	}

	if addr := v.GetString("subject.grpc_addr"); addr != "" {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Warn("subject grpc client failed", zap.Error(err))
		} else {
			h.conns = append(h.conns, conn)
			subjectClient = subjectv1.NewSubjectServiceClient(conn)
			h.Subject = httpif.NewSubjectHandler(
				subjectClient,
				subjectv1.NewPrerequisiteServiceClient(conn),
			)
			log.Info("subject handler ready", zap.String("addr", addr))
		}
	}

	if addr := v.GetString("timetable.grpc_addr"); addr != "" {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Warn("timetable grpc client failed", zap.Error(err))
		} else {
			h.conns = append(h.conns, conn)
			semesterClient = timetablev1.NewSemesterServiceClient(conn)
			h.Timetable = httpif.NewTimetableHandler(
				timetablev1.NewTimetableServiceClient(conn),
				semesterClient,
				js,
			)
			log.Info("timetable handler ready", zap.String("addr", addr))
		}
	}

	// Dashboard aggregates counts from all modules; nil clients are handled gracefully.
	h.Dashboard = httpif.NewDashboardHandler(teacherClient, departmentClient, subjectClient, semesterClient)

	return h
}
