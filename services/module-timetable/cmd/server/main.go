package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	timetablev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/timetable/v1"

	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/application/query"
	infragrpc "github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/infrastructure/grpc"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/infrastructure/messaging"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/infrastructure/persistence"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/infrastructure/persistence/sqlc"
	grpcif "github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/interface/grpc"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/service"
)

func main() {
	// 1. Load config
	v := viper.New()
	v.SetConfigName("local")
	v.AddConfigPath("config")
	v.AddConfigPath(".")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("read config: %v", err)
	}

	// 2. Init logger
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer zapLog.Sync() //nolint:errcheck

	// 3. Connect to database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, v.GetString("database.url"))
	if err != nil {
		zapLog.Fatal("connect to database", zap.Error(err))
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		zapLog.Fatal("ping database", zap.Error(err))
	}
	zapLog.Info("connected to database")

	// 4. Repositories
	queries := sqlc.New(pool)
	semesterRepo := persistence.NewSemesterRepository(queries)
	roomRepo := persistence.NewRoomRepository(queries)
	scheduleRepo := persistence.NewScheduleRepository(queries)

	// 5. Infrastructure clients
	hrClient, err := infragrpc.NewHRClient(v.GetString("hr.grpc_addr"))
	if err != nil {
		zapLog.Fatal("connect to HR service", zap.Error(err))
	}

	subjectClient, err := infragrpc.NewSubjectClient(v.GetString("subject.grpc_addr"))
	if err != nil {
		zapLog.Fatal("connect to Subject service", zap.Error(err))
	}

	natsPublisher, err := messaging.NewNATSPublisher(v.GetString("nats.url"))
	if err != nil {
		zapLog.Warn("connect to NATS (non-fatal, events disabled)", zap.Error(err))
		// natsPublisher will be nil — use no-op publisher below
	}

	var publisher command.EventPublisher
	if natsPublisher != nil {
		publisher = natsPublisher
	} else {
		publisher = &noopPublisher{}
	}

	// 6. Domain services
	// ConstraintChecker is initialised with empty maps here; the solver populates
	// its own instance at generation time with fresh data from HR/Subject.
	emptyChecker := service.NewConstraintChecker(nil, nil, nil, nil)
	ranker := service.NewTeacherRanker(emptyChecker)

	// 7. Command handlers
	createSemesterHandler := command.NewCreateSemesterHandler(semesterRepo, publisher)
	createRoomHandler := command.NewCreateRoomHandler(roomRepo)
	generateScheduleHandler := command.NewGenerateScheduleHandler(
		semesterRepo, scheduleRepo, roomRepo, hrClient, subjectClient, publisher,
	)
	manualAssignHandler := command.NewManualAssignHandler(scheduleRepo, publisher)
	publishScheduleHandler := command.NewPublishScheduleHandler(scheduleRepo)
	_ = createRoomHandler
	_ = publishScheduleHandler

	// 8. Query handlers
	getScheduleHandler := query.NewGetScheduleHandler(scheduleRepo)
	listSchedulesHandler := query.NewListSchedulesHandler(scheduleRepo)
	suggestTeachersHandler := query.NewSuggestTeachersHandler(hrClient, ranker)

	// 9. gRPC servers
	timetableServer := grpcif.NewTimetableServer(
		generateScheduleHandler,
		manualAssignHandler,
		getScheduleHandler,
		listSchedulesHandler,
		suggestTeachersHandler,
		roomRepo,
	)
	semesterServer := grpcif.NewSemesterServer(
		createSemesterHandler,
		listSchedulesHandler,
		semesterRepo,
	)

	// 10. Start gRPC
	grpcServer := grpc.NewServer()
	timetablev1.RegisterTimetableServiceServer(grpcServer, timetableServer)
	timetablev1.RegisterSemesterServiceServer(grpcServer, semesterServer)

	grpcPort := v.GetInt("server.grpc_port")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		zapLog.Fatal("listen grpc", zap.Error(err))
	}

	go func() {
		zapLog.Info("starting Timetable gRPC server", zap.Int("port", grpcPort))
		if err := grpcServer.Serve(lis); err != nil {
			zapLog.Error("grpc server error", zap.Error(err))
		}
	}()

	// 11. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zapLog.Info("shutting down Timetable module...")
	grpcServer.GracefulStop()
	zapLog.Info("Timetable module stopped")
}

// noopPublisher silently discards events when NATS is unavailable.
type noopPublisher struct{}

func (n *noopPublisher) Publish(_ context.Context, _ string, _ any) error { return nil }
