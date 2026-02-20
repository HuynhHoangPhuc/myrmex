package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	hrv1 "github.com/myrmex-erp/myrmex/gen/go/hr/v1"
	"github.com/myrmex-erp/myrmex/services/module-hr/internal/application/command"
	"github.com/myrmex-erp/myrmex/services/module-hr/internal/application/query"
	"github.com/myrmex-erp/myrmex/services/module-hr/internal/infrastructure/persistence"
	"github.com/myrmex-erp/myrmex/services/module-hr/internal/infrastructure/persistence/sqlc"
	grpcif "github.com/myrmex-erp/myrmex/services/module-hr/internal/interface/grpc"
)

func main() {
	// 1. Load config
	v := viper.New()
	v.SetConfigName("local")
	v.AddConfigPath("config")
	v.AddConfigPath(".")
	v.AutomaticEnv()
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("read config: %v", err)
	}

	// 2. Init logger
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer zapLog.Sync()

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

	// 4. Repositories (sqlc)
	queries := sqlc.New(pool)
	teacherRepo := persistence.NewTeacherRepository(queries)
	deptRepo := persistence.NewDepartmentRepository(queries)

	// 5. Command handlers
	createTeacherHandler := command.NewCreateTeacherHandler(teacherRepo)
	updateTeacherHandler := command.NewUpdateTeacherHandler(teacherRepo)
	deleteTeacherHandler := command.NewDeleteTeacherHandler(teacherRepo)
	updateAvailabilityHandler := command.NewUpdateAvailabilityHandler(teacherRepo)
	createDepartmentHandler := command.NewCreateDepartmentHandler(deptRepo)

	// 6. Query handlers
	getTeacherHandler := query.NewGetTeacherHandler(teacherRepo)
	listTeachersHandler := query.NewListTeachersHandler(teacherRepo)
	getAvailabilityHandler := query.NewGetAvailabilityHandler(teacherRepo)
	listDepartmentsHandler := query.NewListDepartmentsHandler(deptRepo)

	// 7. gRPC servers
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

	// 8. Start gRPC server
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

	// 9. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zapLog.Info("shutting down HR module...")

	grpcServer.GracefulStop()
	zapLog.Info("HR module stopped")
}
