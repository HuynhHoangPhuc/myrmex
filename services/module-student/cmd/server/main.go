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

	pkgcache "github.com/HuynhHoangPhuc/myrmex/pkg/cache"
	subjectv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/subject/v1"
	studentv1 "github.com/HuynhHoangPhuc/myrmex/gen/go/student/v1"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/application/command"
	appservice "github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/application/service"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/application/query"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/infrastructure/messaging"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/infrastructure/persistence"
	"github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/infrastructure/persistence/sqlc"
	grpcif "github.com/HuynhHoangPhuc/myrmex/services/module-student/internal/interface/grpc"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type subscription interface {
	Unsubscribe() error
}

func main() {
	v := viper.New()
	v.SetConfigName("local")
	v.AddConfigPath("config")
	v.AddConfigPath(".")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("read config: %v", err)
	}

	zapLog, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer zapLog.Sync()

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

	queries := sqlc.New(pool)
	studentRepo := persistence.NewStudentRepository(queries)
	enrollmentRepo := persistence.NewEnrollmentRepository(queries)

	var publisher command.EventPublisher
	var natsPublisher *messaging.NATSPublisher
	natsURL := v.GetString("nats.url")
	if natsURL != "" {
		np, err := messaging.NewNATSPublisher(natsURL)
		if err != nil {
			zapLog.Warn("NATS unavailable, events will not be published", zap.Error(err))
			publisher = command.NewNoopPublisher()
		} else {
			natsPublisher = np
			publisher = np
			zapLog.Info("connected to NATS for event publishing")
		}
	} else {
		publisher = command.NewNoopPublisher()
	}

	// Redis cache — shared by prerequisite checker and grade/transcript handlers
	var cacheStore pkgcache.Cache
	redisAddr := strings.TrimSpace(v.GetString("redis.addr"))
	if redisAddr != "" {
		cacheStore = pkgcache.NewRedisCache(redisAddr, v.GetString("redis.password"), v.GetInt("redis.db"))
		zapLog.Info("configured redis cache", zap.String("addr", redisAddr), zap.Int("db", v.GetInt("redis.db")))
	}

	var prerequisiteChecker command.PrerequisiteChecker
	var prerequisiteInvalidationSubs []subscription
	var subjectClient query.SubjectListClient
	subjectGRPCAddr := strings.TrimSpace(v.GetString("subject.grpc_addr"))
	if subjectGRPCAddr != "" {
		dialCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		subjectConn, err := grpc.DialContext(
			dialCtx,
			subjectGRPCAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)
		cancel()
		if err != nil {
			zapLog.Fatal("connect to subject service", zap.String("addr", subjectGRPCAddr), zap.Error(err))
		}
		defer subjectConn.Close()

		subjectClient = subjectv1.NewSubjectServiceClient(subjectConn)
		prerequisiteChecker = appservice.NewPrerequisiteChecker(
			enrollmentRepo,
			subjectv1.NewSubjectServiceClient(subjectConn),
			subjectv1.NewPrerequisiteServiceClient(subjectConn),
			cacheStore,
			time.Hour,
		)
		if cacheStore != nil && natsPublisher != nil {
			subs, err := messaging.NewPrerequisiteCacheInvalidator(cacheStore).Start(ctx, natsPublisher)
			if err != nil {
				zapLog.Warn("failed to start prerequisite cache invalidator", zap.Error(err))
			} else {
				prerequisiteInvalidationSubs = make([]subscription, len(subs))
				for i, sub := range subs {
					prerequisiteInvalidationSubs[i] = sub
				}
				zapLog.Info("started prerequisite cache invalidator", zap.Int("subjects", len(subs)))
			}
		}
		zapLog.Info("connected to subject service for prerequisite checks", zap.String("addr", subjectGRPCAddr))
	} else {
		zapLog.Warn("subject.grpc_addr not configured; prerequisite checks disabled")
	}

	gradeRepo := persistence.NewGradeRepository(queries)

	createStudentHandler := command.NewCreateStudentHandler(studentRepo, publisher)
	updateStudentHandler := command.NewUpdateStudentHandler(studentRepo, publisher)
	deleteStudentHandler := command.NewDeleteStudentHandler(studentRepo, publisher)
	requestEnrollmentHandler := command.NewRequestEnrollmentHandler(studentRepo, enrollmentRepo, prerequisiteChecker, publisher)
	reviewEnrollmentHandler := command.NewReviewEnrollmentHandler(enrollmentRepo, publisher)
	assignGradeHandler := command.NewAssignGradeHandler(enrollmentRepo, gradeRepo, cacheStore, publisher)
	updateGradeHandler := command.NewUpdateGradeHandler(enrollmentRepo, gradeRepo, cacheStore)
	getStudentHandler := query.NewGetStudentHandler(studentRepo)
	listStudentsHandler := query.NewListStudentsHandler(studentRepo)
	listEnrollmentRequestsHandler := query.NewListEnrollmentRequestsHandler(enrollmentRepo)
	getStudentEnrollmentsHandler := query.NewGetStudentEnrollmentsHandler(enrollmentRepo)
	getTranscriptHandler := query.NewGetStudentTranscriptHandler(studentRepo, gradeRepo, subjectClient, cacheStore)

	studentServer := grpcif.NewStudentServer(
		createStudentHandler,
		updateStudentHandler,
		deleteStudentHandler,
		getStudentHandler,
		listStudentsHandler,
		requestEnrollmentHandler,
		reviewEnrollmentHandler,
		listEnrollmentRequestsHandler,
		getStudentEnrollmentsHandler,
		prerequisiteChecker,
		assignGradeHandler,
		updateGradeHandler,
		getTranscriptHandler,
	)

	grpcServer := grpc.NewServer()
	studentv1.RegisterStudentServiceServer(grpcServer, studentServer)

	grpcPort := v.GetInt("server.grpc_port")
	grpcLis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		zapLog.Fatal("listen grpc", zap.Error(err))
	}

	go func() {
		zapLog.Info("starting student gRPC server", zap.Int("port", grpcPort))
		if err := grpcServer.Serve(grpcLis); err != nil {
			zapLog.Error("grpc server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zapLog.Info("shutting down student module...")

	for _, sub := range prerequisiteInvalidationSubs {
		_ = sub.Unsubscribe()
	}
	grpcServer.GracefulStop()
	zapLog.Info("student module stopped")
}
