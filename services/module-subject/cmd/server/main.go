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

	subjectv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/subject/v1"
	pkgmessaging "github.com/HuynhHoangPhuc/myrmex/pkg/messaging"
	msgnats "github.com/HuynhHoangPhuc/myrmex/pkg/messaging/nats"
	msgpubsub "github.com/HuynhHoangPhuc/myrmex/pkg/messaging/pubsub"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/application/query"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/domain/service"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/infrastructure/messaging"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/infrastructure/persistence"
	"github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/infrastructure/persistence/sqlc"
	grpcif "github.com/HuynhHoangPhuc/myrmex/services/module-subject/internal/interface/grpc"
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

	// 4. Infrastructure: sqlc queries + repositories
	queries := sqlc.New(pool)
	subjectRepo := persistence.NewSubjectRepository(queries)
	prereqRepo := persistence.NewPrerequisiteRepository(queries)

	// 5. Domain services
	dagService := service.NewDAGService(prereqRepo)

	// 6. Messaging publisher — select backend via MESSAGING_BACKEND env var
	pub := newPublisher(ctx, v, zapLog, pkgmessaging.StreamConfig{
		Name: "SUBJECT_EVENTS", Subjects: []string{"subject.>"},
	})
	defer pub.Close() //nolint:errcheck
	publisher := command.EventPublisher(messaging.NewEventPublisher(pub))

	// 7. Command handlers
	createSubjectHandler := command.NewCreateSubjectHandler(subjectRepo, publisher)
	updateSubjectHandler := command.NewUpdateSubjectHandler(subjectRepo, publisher)
	deleteSubjectHandler := command.NewDeleteSubjectHandler(subjectRepo, publisher)
	addPrereqHandler := command.NewAddPrerequisiteHandler(prereqRepo, subjectRepo, dagService, publisher)
	removePrereqHandler := command.NewRemovePrerequisiteHandler(prereqRepo, publisher)

	// 8. Query handlers
	getSubjectHandler := query.NewGetSubjectHandler(subjectRepo)
	listSubjectsHandler := query.NewListSubjectsHandler(subjectRepo)
	listPrereqsHandler := query.NewListPrerequisitesHandler(prereqRepo)
	validateDAGHandler := query.NewValidateDAGHandler(dagService)
	topoSortHandler := query.NewTopologicalSortHandler(dagService)
	fullDAGHandler := query.NewGetFullDAGHandler(subjectRepo, prereqRepo)
	checkConflictsHandler := query.NewCheckConflictsHandler(dagService, subjectRepo)

	// 9. gRPC servers
	subjectServer := grpcif.NewSubjectServer(
		createSubjectHandler,
		updateSubjectHandler,
		deleteSubjectHandler,
		getSubjectHandler,
		listSubjectsHandler,
	)
	prereqServer := grpcif.NewPrerequisiteServer(
		addPrereqHandler,
		removePrereqHandler,
		listPrereqsHandler,
		validateDAGHandler,
		topoSortHandler,
		fullDAGHandler,
		checkConflictsHandler,
	)

	// 10. gRPC server setup
	grpcServer := grpc.NewServer()
	subjectv1.RegisterSubjectServiceServer(grpcServer, subjectServer)
	subjectv1.RegisterPrerequisiteServiceServer(grpcServer, prereqServer)

	grpcPort := v.GetInt("server.grpc_port")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		zapLog.Fatal("listen grpc", zap.Error(err))
	}

	go func() {
		zapLog.Info("starting gRPC server", zap.Int("port", grpcPort))
		if err := grpcServer.Serve(lis); err != nil {
			zapLog.Error("grpc server error", zap.Error(err))
			os.Exit(1)
		}
	}()

	// 11. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLog.Info("shutting down subject module...")
	grpcServer.GracefulStop()
	zapLog.Info("subject module stopped")
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
