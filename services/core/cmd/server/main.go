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

	// 7. Repositories (sqlc)
	queries := sqlc.New(pool)
	userRepo := persistence.NewUserRepository(queries)
	moduleRepo := persistence.NewModuleRepository(queries)

	// 8. Command handlers
	registerUserHandler := command.NewRegisterUserHandler(userRepo, es, publisher, passwordSvc)
	updateUserHandler := command.NewUpdateUserHandler(userRepo)
	deleteUserHandler := command.NewDeleteUserHandler(userRepo)
	registerModuleHandler := command.NewRegisterModuleHandler(moduleRepo)

	// 9. Query handlers
	getUserHandler := query.NewGetUserHandler(userRepo)
	listUsersHandler := query.NewListUsersHandler(userRepo)
	loginHandler := query.NewLoginHandler(userRepo, jwtSvc, passwordSvc)
	listModulesHandler := query.NewListModulesHandler(moduleRepo)

	// 10. HTTP handlers
	authHandler := httpif.NewAuthHandler(registerUserHandler, loginHandler, jwtSvc)
	userHandler := httpif.NewUserHandler(getUserHandler, listUsersHandler, updateUserHandler, deleteUserHandler)
	gatewayProxy := httpif.NewGatewayProxy(zapLog)
	defer gatewayProxy.Close()
	moduleHandler := httpif.NewModuleHandler(registerModuleHandler, listModulesHandler, gatewayProxy)

	// 10a. Module gRPC client connections (graceful — nil if unavailable)
	modHandlers := buildModuleHandlers(v, js, zapLog)
	defer modHandlers.Close()

	// 10b. AI Chat: LLM provider + tool registry + executor + handler
	llmProvider := buildLLMProvider(v)
	toolRegistry := agent.NewToolRegistry()
	toolRegistry.Register(agent.DefaultTools)
	toolExecutor := agent.NewToolExecutor(toolRegistry, gatewayProxy, nil, zapLog)
	chatMsgHandler := command.NewChatMessageHandler(llmProvider, toolRegistry, toolExecutor, zapLog)
	chatHandler := httpif.NewChatHandler(chatMsgHandler, jwtSvc, zapLog)

	// 11. Router
	router := httpif.NewRouter(httpif.RouterConfig{
		AuthHandler:      authHandler,
		UserHandler:      userHandler,
		ModuleHandler:    moduleHandler,
		GatewayProxy:     gatewayProxy,
		ChatHandler:      chatHandler,
		HRHandler:        modHandlers.HR,
		SubjectHandler:   modHandlers.Subject,
		TimetableHandler: modHandlers.Timetable,
		JWTService:       jwtSvc,
		Logger:           zapLog,
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
//   llm.provider  = "openai" | "claude"
//   llm.api_key   = API key
//   llm.model     = model name
//   llm.base_url  = base URL (optional, for OpenAI-compat endpoints like Ollama)
func buildLLMProvider(v *viper.Viper) llm.LLMProvider {
	provider := v.GetString("llm.provider")
	apiKey := v.GetString("llm.api_key")
	model := v.GetString("llm.model")
	baseURL := v.GetString("llm.base_url")

	switch provider {
	case "claude":
		if model == "" {
			model = "claude-haiku-4-5"
		}
		return llm.NewClaudeProvider(apiKey, model)
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
