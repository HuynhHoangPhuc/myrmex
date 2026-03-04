package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/HuynhHoangPhuc/myrmex/pkg/eventstore"
	pkgnats "github.com/HuynhHoangPhuc/myrmex/pkg/nats"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/application/query"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/agent"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/auth"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/llm"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/messaging"
	notifinfra "github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/notification"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/persistence"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/persistence/sqlc"
	httpif "github.com/HuynhHoangPhuc/myrmex/services/core/internal/interface/http"
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

	// 4. Event store
	es := eventstore.NewPostgresEventStore(pool, "core")

	// 5. NATS connection (optional, continue without if unavailable)
	var publisher *pkgnats.Publisher
	nc, js, err := pkgnats.Connect(v.GetString("nats.url"))
	if err != nil {
		zapLog.Warn("nats connection failed, continuing without events", zap.Error(err))
	} else {
		defer nc.Close()
		publisher = pkgnats.NewPublisher(js)
		// Ensure TIMETABLE stream exists for SSE subscriptions
		if _, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
			Name:     "TIMETABLE",
			Subjects: []string{"timetable.schedule.>"},
		}); err != nil {
			zapLog.Warn("create TIMETABLE stream failed", zap.Error(err))
		}
		zapLog.Info("connected to NATS")
	}

	// 6. Auth services
	accessExpiry, _ := time.ParseDuration(v.GetString("jwt.access_expiry"))
	refreshExpiry, _ := time.ParseDuration(v.GetString("jwt.refresh_expiry"))
	jwtSvc := auth.NewJWTService(v.GetString("jwt.secret"), accessExpiry, refreshExpiry)
	passwordSvc := auth.NewPasswordService()

	// 6b. OAuth service (optional — only initialized when client IDs are configured)
	var oauthSvc *auth.OAuthService
	if gClientID := v.GetString("oauth.google.client_id"); gClientID != "" {
		oauthCfg := auth.OAuthConfig{
			GoogleClientID:        gClientID,
			GoogleClientSecret:    v.GetString("oauth.google.client_secret"),
			GoogleRedirectURL:     v.GetString("oauth.google.redirect_url"),
			MicrosoftClientID:     v.GetString("oauth.microsoft.client_id"),
			MicrosoftClientSecret: v.GetString("oauth.microsoft.client_secret"),
			MicrosoftTenantID:     v.GetString("oauth.microsoft.tenant_id"),
			MicrosoftRedirectURL:  v.GetString("oauth.microsoft.redirect_url"),
			FrontendCallbackURL:   v.GetString("oauth.frontend_callback_url"),
		}
		svc, err := auth.NewOAuthService(ctx, oauthCfg)
		if err != nil {
			zapLog.Warn("oauth service init failed, continuing without OAuth", zap.Error(err))
		} else {
			oauthSvc = svc
			zapLog.Info("oauth service initialized")
		}
	} else {
		zapLog.Info("oauth not configured (set oauth.google.client_id to enable)")
	}

	// 6a. Generate internal service JWT for tool executor HTTP dispatch
	internalJWT, err := jwtSvc.GenerateInternalToken()
	if err != nil {
		zapLog.Fatal("failed to generate internal service token", zap.Error(err))
	}
	selfURL := v.GetString("server.self_url")
	if selfURL == "" {
		selfURL = fmt.Sprintf("http://localhost:%d", v.GetInt("server.http_port"))
	}

	// 7. Repositories (sqlc)
	queries := sqlc.New(pool)
	userRepo := persistence.NewUserRepository(queries)
	moduleRepo := persistence.NewModuleRepository(queries)
	auditRepo := persistence.NewAuditLogRepository(pool)
	notifRepo := persistence.NewNotificationRepository(pool)

	// 8. Command handlers
	registerUserHandler := command.NewRegisterUserHandler(userRepo, es, publisher, passwordSvc)
	updateUserHandler := command.NewUpdateUserHandler(userRepo)
	deleteUserHandler := command.NewDeleteUserHandler(userRepo)
	updateUserRoleHandler := command.NewUpdateUserRoleHandler(userRepo)
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
	modHandlers := buildModuleHandlers(v, js, zapLog)
	defer modHandlers.Close()

	authHandler := httpif.NewAuthHandler(registerUserHandler, loginHandler, jwtSvc,
		userRepo, updateUserHandler, deleteUserHandler, modHandlers.StudentClient)
	adminRoleHandler := httpif.NewAdminRoleHandler(updateUserRoleHandler)

	var oauthHandler *httpif.OAuthHandler
	if oauthSvc != nil {
		oauthHandler = httpif.NewOAuthHandler(oauthSvc, jwtSvc, userRepo)
	}

	// Audit consumer: only start when NATS is available
	auditHandler := httpif.NewAuditHandler(auditRepo)
	if js != nil {
		auditConsumer := messaging.NewAuditConsumer(js, auditRepo, zapLog)
		if err := auditConsumer.Start(ctx); err != nil {
			zapLog.Warn("audit consumer failed to start", zap.Error(err))
		} else {
			defer auditConsumer.Stop()
		}
	}

	// Notification infrastructure: WS broker + email service + NATS consumer
	wsBroker := notifinfra.NewWSBroker(zapLog)
	emailSvc := notifinfra.NewEmailService(
		v.GetString("resend.api_key"),
		v.GetString("resend.from_addr"),
		zapLog,
	)
	// Wire user email lookup for the email service
	notifinfra.GetUserEmail = func(userID string) (string, error) {
		uid, err := uuid.Parse(userID)
		if err != nil {
			return "", err
		}
		u, err := userRepo.GetByID(ctx, uid)
		if err != nil {
			return "", err
		}
		return u.Email, nil
	}
	notifHandler := httpif.NewNotificationHandler(notifRepo, wsBroker, jwtSvc, zapLog)
	if js != nil {
		notifConsumer := messaging.NewNotificationConsumer(js, notifRepo, wsBroker, emailSvc, zapLog)
		if err := notifConsumer.Start(ctx); err != nil {
			zapLog.Warn("notification consumer failed to start", zap.Error(err))
		} else {
			defer notifConsumer.Stop()
		}
	}

	// 10b. AI Chat: LLM provider + tool registry + executor + handler
	llmProvider := buildLLMProvider(v)
	toolRegistry := agent.NewToolRegistry()
	toolRegistry.Register(agent.DefaultTools)
	toolExecutor := agent.NewToolExecutor(toolRegistry, gatewayProxy, selfURL, internalJWT, zapLog)
	chatMsgHandler := command.NewChatMessageHandler(llmProvider, toolRegistry, toolExecutor, zapLog)
	chatHandler := httpif.NewChatHandler(chatMsgHandler, jwtSvc, zapLog)

	// 11. Router
	router := httpif.NewRouter(httpif.RouterConfig{
		AuthHandler:         authHandler,
		OAuthHandler:        oauthHandler,
		AdminRoleHandler:    adminRoleHandler,
		AuditHandler:        auditHandler,
		NATSConn:            nc, // nil when NATS unavailable → audit middleware no-ops
		UserHandler:         userHandler,
		ModuleHandler:       moduleHandler,
		GatewayProxy:        gatewayProxy,
		ChatHandler:         chatHandler,
		NotificationHandler: notifHandler,
		HRHandler:           modHandlers.HR,
		SubjectHandler:      modHandlers.Subject,
		TimetableHandler:    modHandlers.Timetable,
		StudentHandler:      modHandlers.Student,
		DashboardHandler:    modHandlers.Dashboard,
		AnalyticsHTTPAddr:   v.GetString("analytics.http_addr"),
		JWTService:          jwtSvc,
		Logger:              zapLog,
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

// buildLLMProvider reads config and returns the configured LLM provider.
// Defaults to OpenAI-compatible if llm.provider is not set.
// Config keys:
//
//	llm.provider  = "openai" | "claude" | "gemini"
//	llm.api_key   = API key
//	llm.model     = model name
//	llm.base_url  = base URL (optional, for OpenAI-compat endpoints like Ollama)
func buildLLMProvider(v *viper.Viper) llm.LLMProvider {
	provider := v.GetString("llm.provider")
	apiKey := v.GetString("llm.api_key")
	model := v.GetString("llm.model")
	baseURL := v.GetString("llm.base_url")

	switch provider {
	case "mock":
		return llm.NewMockProvider()
	case "claude":
		if model == "" {
			model = "claude-haiku-4-5"
		}
		return llm.NewClaudeProvider(apiKey, model)
	case "gemini":
		if model == "" {
			model = "gemini-2.0-flash"
		}
		return llm.NewGeminiProvider(apiKey, model)
	default: // "openai" or any OpenAI-compatible
		if model == "" {
			model = "gpt-4o-mini"
		}
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}
		return llm.NewOpenAIProvider(apiKey, model, baseURL)
	}
}
