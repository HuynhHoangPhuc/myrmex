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

	subjectv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/subject/v1"
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
	pool, err := pgxpool.New(ctx, v.GetString("database.url"))
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

	// 5a. NATS publisher (optional — continue without if unavailable)
	var publisher command.EventPublisher
	natsURL := v.GetString("nats.url")
	if natsURL != "" {
		np, err := messaging.NewNATSPublisher(natsURL)
		if err != nil {
			zapLog.Warn("NATS unavailable, events will not be published", zap.Error(err))
			publisher = command.NewNoopPublisher()
		} else {
			publisher = np
			zapLog.Info("connected to NATS for event publishing")
		}
	} else {
		publisher = command.NewNoopPublisher()
	}

	// 6. Command handlers
	createSubjectHandler := command.NewCreateSubjectHandler(subjectRepo, publisher)
	updateSubjectHandler := command.NewUpdateSubjectHandler(subjectRepo, publisher)
	deleteSubjectHandler := command.NewDeleteSubjectHandler(subjectRepo, publisher)
	addPrereqHandler := command.NewAddPrerequisiteHandler(prereqRepo, subjectRepo, dagService, publisher)
	removePrereqHandler := command.NewRemovePrerequisiteHandler(prereqRepo, publisher)

	// 7. Query handlers
	getSubjectHandler := query.NewGetSubjectHandler(subjectRepo)
	listSubjectsHandler := query.NewListSubjectsHandler(subjectRepo)
	listPrereqsHandler := query.NewListPrerequisitesHandler(prereqRepo)
	validateDAGHandler := query.NewValidateDAGHandler(dagService)
	topoSortHandler := query.NewTopologicalSortHandler(dagService)
	fullDAGHandler := query.NewGetFullDAGHandler(subjectRepo, prereqRepo)
	checkConflictsHandler := query.NewCheckConflictsHandler(dagService, subjectRepo)

	// 8. gRPC servers (interface layer)
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

	// 9. gRPC server setup
	grpcServer := grpc.NewServer()
	subjectv1.RegisterSubjectServiceServer(grpcServer, subjectServer)
	subjectv1.RegisterPrerequisiteServiceServer(grpcServer, prereqServer)

	grpcPort := v.GetInt("server.grpc_port")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		zapLog.Fatal("listen grpc", zap.Error(err))
	}

	// 10. Start gRPC server in background
	go func() {
		zapLog.Info("starting gRPC server", zap.Int("port", grpcPort))
		if err := grpcServer.Serve(lis); err != nil {
			zapLog.Error("grpc server error", zap.Error(err))
			os.Exit(1)
		}
	}()

	// 11. Graceful shutdown on SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLog.Info("shutting down subject module...")
	grpcServer.GracefulStop()
	zapLog.Info("subject module stopped")
}
