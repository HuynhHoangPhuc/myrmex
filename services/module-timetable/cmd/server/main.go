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
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	timetablev1 "github.com/HuynhHoangPhuc/myrmex/gen/go/timetable/v1"
	pkgmessaging "github.com/HuynhHoangPhuc/myrmex/pkg/messaging"
	msgnats "github.com/HuynhHoangPhuc/myrmex/pkg/messaging/nats"
	msgpubsub "github.com/HuynhHoangPhuc/myrmex/pkg/messaging/pubsub"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/application/query"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/domain/service"
	infragrpc "github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/infrastructure/grpc"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/infrastructure/messaging"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/infrastructure/persistence"
	"github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/infrastructure/persistence/sqlc"
	grpcif "github.com/HuynhHoangPhuc/myrmex/services/module-timetable/internal/interface/grpc"
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
	poolCfg, err := pgxpool.ParseConfig(v.GetString("database.url"))
	if err != nil {
		zapLog.Fatal("parse database config", zap.Error(err))
	}
	poolCfg.MaxConns = 15
	poolCfg.MinConns = 3
	poolCfg.MaxConnLifetime = 30 * time.Minute
	poolCfg.MaxConnIdleTime = 5 * time.Minute
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
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

	// 5. Infrastructure gRPC clients
	hrClient, err := infragrpc.NewHRClient(v.GetString("hr.grpc_addr"))
	if err != nil {
		zapLog.Fatal("connect to HR service", zap.Error(err))
	}

	subjectClient, err := infragrpc.NewSubjectClient(v.GetString("subject.grpc_addr"))
	if err != nil {
		zapLog.Fatal("connect to Subject service", zap.Error(err))
	}

	// 6. Messaging publisher — select backend via MESSAGING_BACKEND env var
	pub := newPublisher(ctx, v, zapLog, pkgmessaging.StreamConfig{
		Name: "TIMETABLE", Subjects: []string{"timetable.>"},
	})
	defer pub.Close() //nolint:errcheck
	publisher := command.EventPublisher(messaging.NewEventPublisher(pub))

	// 7. Domain services
	emptyChecker := service.NewConstraintChecker(nil, nil, nil, nil)
	ranker := service.NewTeacherRanker(emptyChecker)

	// 8. Command handlers
	createSemesterHandler := command.NewCreateSemesterHandler(semesterRepo, publisher)
	createRoomHandler := command.NewCreateRoomHandler(roomRepo)
	generateScheduleHandler := command.NewGenerateScheduleHandler(
		semesterRepo, scheduleRepo, roomRepo, hrClient, subjectClient, publisher,
	)
	manualAssignHandler := command.NewManualAssignHandler(scheduleRepo, publisher)
	publishScheduleHandler := command.NewPublishScheduleHandler(scheduleRepo)
	_ = createRoomHandler
	_ = publishScheduleHandler

	// 9. Query handlers
	getScheduleHandler := query.NewGetScheduleHandler(scheduleRepo)
	listSchedulesHandler := query.NewListSchedulesHandler(scheduleRepo)
	suggestTeachersHandler := query.NewSuggestTeachersHandler(hrClient, ranker)

	// 10. gRPC servers
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

	// 11. Start gRPC
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

	// 12. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zapLog.Info("shutting down Timetable module...")
	grpcServer.GracefulStop()
	zapLog.Info("Timetable module stopped")
}

func newPublisher(ctx context.Context, v *viper.Viper, log *zap.Logger, stream pkgmessaging.StreamConfig) pkgmessaging.Publisher {
	switch v.GetString("messaging.backend") {
	case "pubsub":
		pub, err := msgpubsub.NewPublisher(ctx, v.GetString("gcp.project_id"), log)
		if err != nil {
			log.Warn("Pub/Sub publisher unavailable", zap.Error(err))
			return &pkgmessaging.NoopPublisher{}
		}
		if err := pub.EnsureStream(ctx, stream); err != nil {
			log.Warn("ensure pubsub topic failed", zap.Error(err))
		}
		log.Info("messaging backend: Cloud Pub/Sub", zap.String("stream", stream.Name))
		return pub
	default: // "nats" or ""
		natsURL := v.GetString("nats.url")
		if natsURL == "" {
			return &pkgmessaging.NoopPublisher{}
		}
		p, err := msgnats.NewPublisher(natsURL)
		if err != nil {
			log.Warn("NATS publisher unavailable", zap.Error(err))
			return &pkgmessaging.NoopPublisher{}
		}
		if err := p.EnsureStream(ctx, stream); err != nil {
			log.Warn("ensure stream failed", zap.Error(err))
		}
		log.Info("messaging backend: NATS", zap.String("stream", stream.Name))
		return p
	}
}
