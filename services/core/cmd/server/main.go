package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	pkgmessaging "github.com/HuynhHoangPhuc/myrmex/pkg/messaging"
	msgnats "github.com/HuynhHoangPhuc/myrmex/pkg/messaging/nats"
	msgpubsub "github.com/HuynhHoangPhuc/myrmex/pkg/messaging/pubsub"
	msgredis "github.com/HuynhHoangPhuc/myrmex/pkg/messaging/redis"
	"github.com/HuynhHoangPhuc/myrmex/pkg/eventstore"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/application/query"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/agent"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/messaging"
	notifinfra "github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/notification"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/persistence"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/persistence/sqlc"
	httpif "github.com/HuynhHoangPhuc/myrmex/services/core/internal/interface/http"
)

func main() {
	// 1. Load config
	v, err := initConfig()
	if err != nil {
		log.Fatalf("read config: %v", err)
	}

	// 2. Init logger
	zapLog, err := initLogger()
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer zapLog.Sync()

	// 2b. Init Sentry (optional — only when DSN is configured)
	if sentryDSN := v.GetString("sentry.dsn"); sentryDSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              sentryDSN,
			Environment:      v.GetString("environment"),
			TracesSampleRate: 0.1,
		}); err != nil {
			zapLog.Warn("sentry init failed, continuing without error tracking", zap.Error(err))
		} else {
			defer sentry.Flush(2 * time.Second)
			zapLog.Info("sentry initialized")
		}
	}

	// 3. Connect to database
	ctx := context.Background()
	pool, err := initDatabase(ctx, v, zapLog)
	if err != nil {
		zapLog.Fatal("connect to database", zap.Error(err))
	}
	defer pool.Close()

	// 4. Event store
	es := eventstore.NewPostgresEventStore(pool, "core")

	// 5. NATS connection (optional, continue without if unavailable)
	natsConn := initNATS(ctx, v, zapLog)
	if natsConn != nil {
		defer natsConn.NC.Close()
	}

	// 5b. Redis client (optional — used for WS push relay)
	rdb := initRedis(ctx, v, zapLog)
	if rdb != nil {
		defer rdb.Close()
	}

	// 6. Auth services
	authSvcs, err := initAuthServices(ctx, v, zapLog)
	if err != nil {
		zapLog.Fatal("init auth services", zap.Error(err))
	}
	jwtSvc := authSvcs.JWT
	passwordSvc := authSvcs.Password
	oauthSvc := authSvcs.OAuth

	// 6a. Generate internal service JWT for tool executor HTTP dispatch
	internalJWT, err := jwtSvc.GenerateInternalToken()
	if err != nil {
		zapLog.Fatal("failed to generate internal service token", zap.Error(err))
	}
	selfURL := buildSelfURL(v)

	// 7. Repositories (sqlc)
	queries := sqlc.New(pool)
	userRepo := persistence.NewUserRepository(queries)
	moduleRepo := persistence.NewModuleRepository(queries)
	auditRepo := persistence.NewAuditLogRepository(pool)

	// 8. Command handlers
	registerUserHandler := command.NewRegisterUserHandler(userRepo, es, toNATSPublisher(natsConn), passwordSvc)
	updateUserHandler := command.NewUpdateUserHandler(userRepo)
	deleteUserHandler := command.NewDeleteUserHandler(userRepo)
	updateUserRoleHandler := command.NewUpdateUserRoleHandler(userRepo, toNATSPublisher(natsConn))
	registerModuleHandler := command.NewRegisterModuleHandler(moduleRepo)

	// 9. Query handlers
	getUserHandler := query.NewGetUserHandler(userRepo)
	listUsersHandler := query.NewListUsersHandler(userRepo)
	loginHandler := query.NewLoginHandler(userRepo, jwtSvc, passwordSvc)
	listModulesHandler := query.NewListModulesHandler(moduleRepo)

	// 10. HTTP handlers
	userHandler := httpif.NewUserHandler(getUserHandler, listUsersHandler, updateUserHandler, deleteUserHandler)
	gatewayProxy := httpif.NewGatewayProxy(zapLog)
	defer gatewayProxy.Close()
	moduleHandler := httpif.NewModuleHandler(registerModuleHandler, listModulesHandler, gatewayProxy)

	// 10a. Module gRPC client connections (graceful — nil if unavailable)
	modHandlers := buildModuleHandlers(v, toNATSJS(natsConn), zapLog)
	defer modHandlers.Close()

	importHandler := httpif.NewImportHandler(registerUserHandler, modHandlers.TeacherClient, modHandlers.StudentClient)

	authHandler := httpif.NewAuthHandler(registerUserHandler, loginHandler, jwtSvc,
		userRepo, updateUserHandler, deleteUserHandler, modHandlers.StudentClient)
	adminRoleHandler := httpif.NewAdminRoleHandler(updateUserRoleHandler)

	var oauthHandler *httpif.OAuthHandler
	if oauthSvc != nil {
		oauthHandler = httpif.NewOAuthHandler(oauthSvc, jwtSvc, userRepo)
	}

	// Audit consumer + publisher: backend-agnostic selection
	auditHandler := httpif.NewAuditHandler(auditRepo)
	var auditPub pkgmessaging.Publisher = &pkgmessaging.NoopPublisher{}
	if v.GetString("messaging.backend") == "pubsub" {
		if pub, err := msgpubsub.NewPublisher(ctx, v.GetString("gcp.project_id"), zapLog); err != nil {
			zapLog.Warn("Pub/Sub audit publisher unavailable", zap.Error(err))
		} else {
			auditPub = pub
			zapLog.Info("audit publisher: Cloud Pub/Sub")
		}
		if c, err := msgpubsub.NewConsumer(ctx, v.GetString("gcp.project_id"), zapLog); err != nil {
			zapLog.Warn("Pub/Sub audit consumer unavailable", zap.Error(err))
		} else {
			auditConsumer := messaging.NewAuditConsumer(c, auditRepo, zapLog)
			if err := auditConsumer.Start(ctx); err != nil {
				zapLog.Warn("audit consumer failed to start", zap.Error(err))
			}
		}
	} else if natsConn != nil {
		auditConsumer := messaging.NewAuditConsumer(msgnats.NewConsumerFromJS(natsConn.NC, natsConn.JS), auditRepo, zapLog)
		if err := auditConsumer.Start(ctx); err != nil {
			zapLog.Warn("audit consumer failed to start", zap.Error(err))
		}
		auditPub = msgnats.NewPublisherFromJS(natsConn.NC, natsConn.JS)
	}

	// Notification infrastructure: WS broker relays Redis push events from module-notification
	wsBroker := notifinfra.NewWSBroker(zapLog)
	redisSub := msgredis.NewPushSubscriber(rdb)
	wsBroker.StartRedisRelay(ctx, redisSub)
	notifHandler := httpif.NewNotificationHandler(wsBroker, jwtSvc, zapLog)

	// 10b. AI Chat: LLM provider + tool registry + executor + handler
	llmProvider := buildLLMProvider(v)
	toolRegistry := agent.NewToolRegistry()
	toolRegistry.Register(agent.DefaultTools)
	toolExecutor := agent.NewToolExecutor(toolRegistry, gatewayProxy, selfURL, internalJWT, zapLog)
	chatMsgHandler := command.NewChatMessageHandler(llmProvider, toolRegistry, toolExecutor, zapLog)
	chatHandler := httpif.NewChatHandler(chatMsgHandler, jwtSvc, zapLog)

	// 11. Router
	router := httpif.NewRouter(httpif.RouterConfig{
		AuthHandler:          authHandler,
		OAuthHandler:         oauthHandler,
		AdminRoleHandler:     adminRoleHandler,
		AuditHandler:         auditHandler,
		AuditPublisher:       auditPub,
		UserHandler:          userHandler,
		ModuleHandler:        moduleHandler,
		GatewayProxy:         gatewayProxy,
		ChatHandler:          chatHandler,
		NotificationHandler:  notifHandler,
		HRHandler:            modHandlers.HR,
		SubjectHandler:       modHandlers.Subject,
		TimetableHandler:     modHandlers.Timetable,
		StudentHandler:       modHandlers.Student,
		ImportHandler:        importHandler,
		DashboardHandler:     modHandlers.Dashboard,
		AnalyticsHTTPAddr:    v.GetString("analytics.http_addr"),
		NotificationHTTPAddr: v.GetString("notification.http_addr"),
		JWTService:           jwtSvc,
		Logger:               zapLog,
		DB:                   pool,
		Rdb:                  rdb,
	})

	// 12. gRPC server (placeholder for module registry)
	grpcServer := grpc.NewServer()
	grpcLis, err := net.Listen("tcp", fmt.Sprintf(":%d", v.GetInt("server.grpc_port")))
	if err != nil {
		zapLog.Fatal("listen grpc", zap.Error(err))
	}

	// 13. Start servers
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", v.GetInt("server.http_port")),
		Handler: router,
	}

	go func() {
		zapLog.Info("starting gRPC server", zap.Int("port", v.GetInt("server.grpc_port")))
		if err := grpcServer.Serve(grpcLis); err != nil {
			zapLog.Error("grpc server error", zap.Error(err))
		}
	}()

	go func() {
		zapLog.Info("starting HTTP server", zap.Int("port", v.GetInt("server.http_port")))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLog.Fatal("http server error", zap.Error(err))
		}
	}()

	// 14. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zapLog.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	grpcServer.GracefulStop()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		zapLog.Error("http shutdown error", zap.Error(err))
	}
	zapLog.Info("server stopped")
}
