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

	if addr := v.GetString("hr.grpc_addr"); addr != "" {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Warn("hr grpc client failed", zap.Error(err))
		} else {
			h.conns = append(h.conns, conn)
			h.HR = httpif.NewHRHandler(
				hrv1.NewTeacherServiceClient(conn),
				hrv1.NewDepartmentServiceClient(conn),
			)
			log.Info("hr handler ready", zap.String("addr", addr))
		}
	}

	if addr := v.GetString("subject.grpc_addr"); addr != "" {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Warn("subject grpc client failed", zap.Error(err))
		} else {
			h.conns = append(h.conns, conn)
			h.Subject = httpif.NewSubjectHandler(
				subjectv1.NewSubjectServiceClient(conn),
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
			h.Timetable = httpif.NewTimetableHandler(
				timetablev1.NewTimetableServiceClient(conn),
				timetablev1.NewSemesterServiceClient(conn),
				js,
			)
			log.Info("timetable handler ready", zap.String("addr", addr))
		}
	}

	return h
}
