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

	hrv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/hr/v1"
	pkgmessaging "github.com/HuynhHoangPhuc/myrmex/pkg/messaging"
	msgnats "github.com/HuynhHoangPhuc/myrmex/pkg/messaging/nats"
	msgpubsub "github.com/HuynhHoangPhuc/myrmex/pkg/messaging/pubsub"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/application/query"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/infrastructure/messaging"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/infrastructure/persistence"
	"github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/infrastructure/persistence/sqlc"
	grpcif "github.com/HuynhHoangPhuc/myrmex/services/module-hr/internal/interface/grpc"
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

	// 4. Repositories (sqlc)
	queries := sqlc.New(pool)
	teacherRepo := persistence.NewTeacherRepository(queries)
	deptRepo := persistence.NewDepartmentRepository(queries)

	// 5. Messaging publisher — select backend via MESSAGING_BACKEND env var
	pub := newPublisher(ctx, v, zapLog, pkgmessaging.StreamConfig{
		Name: "HR_EVENTS", Subjects: []string{"hr.>"},
	})
	defer pub.Close() //nolint:errcheck
	publisher := command.EventPublisher(messaging.NewEventPublisher(pub))

	// 6. Command handlers
	createTeacherHandler := command.NewCreateTeacherHandler(teacherRepo, publisher)
	updateTeacherHandler := command.NewUpdateTeacherHandler(teacherRepo, publisher)
	deleteTeacherHandler := command.NewDeleteTeacherHandler(teacherRepo, publisher)
	updateAvailabilityHandler := command.NewUpdateAvailabilityHandler(teacherRepo, publisher)
	createDepartmentHandler := command.NewCreateDepartmentHandler(deptRepo, publisher)

	// 7. Query handlers
	getTeacherHandler := query.NewGetTeacherHandler(teacherRepo)
	listTeachersHandler := query.NewListTeachersHandler(teacherRepo)
	getAvailabilityHandler := query.NewGetAvailabilityHandler(teacherRepo)
	listDepartmentsHandler := query.NewListDepartmentsHandler(deptRepo)

	// 8. gRPC servers
	teacherServer := grpcif.NewTeacherServer(
		createTeacherHandler,
		updateTeacherHandler,
		deleteTeacherHandler,
		updateAvailabilityHandler,
		getTeacherHandler,
		listTeachersHandler,
		getAvailabilityHandler,
	)
	departmentServer := grpcif.NewDepartmentServer(
		createDepartmentHandler,
		listDepartmentsHandler,
		deptRepo,
	)

	// 9. Start gRPC server
	grpcServer := grpc.NewServer()
	hrv1.RegisterTeacherServiceServer(grpcServer, teacherServer)
	hrv1.RegisterDepartmentServiceServer(grpcServer, departmentServer)

	grpcPort := v.GetInt("server.grpc_port")
	grpcLis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		zapLog.Fatal("listen grpc", zap.Error(err))
	}

	go func() {
		zapLog.Info("starting HR gRPC server", zap.Int("port", grpcPort))
		if err := grpcServer.Serve(grpcLis); err != nil {
			zapLog.Error("grpc server error", zap.Error(err))
		}
	}()

	// 10. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zapLog.Info("shutting down HR module...")

	grpcServer.GracefulStop()
	zapLog.Info("HR module stopped")
}

// newPublisher creates a messaging.Publisher based on MESSAGING_BACKEND config.
// Ensures the given stream/topic exists. Falls back to NoopPublisher on error.
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
