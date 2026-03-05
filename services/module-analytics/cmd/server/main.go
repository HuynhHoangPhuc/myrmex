package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	pkgmessaging "github.com/HuynhHoangPhuc/myrmex/pkg/messaging"
	msgnats "github.com/HuynhHoangPhuc/myrmex/pkg/messaging/nats"
	msgpubsub "github.com/HuynhHoangPhuc/myrmex/pkg/messaging/pubsub"
	"github.com/HuynhHoangPhuc/myrmex/services/module-analytics/internal/application/export"
	"github.com/HuynhHoangPhuc/myrmex/services/module-analytics/internal/application/query"
	"github.com/HuynhHoangPhuc/myrmex/services/module-analytics/internal/infrastructure/messaging"
	"github.com/HuynhHoangPhuc/myrmex/services/module-analytics/internal/infrastructure/persistence"
	httpif "github.com/HuynhHoangPhuc/myrmex/services/module-analytics/internal/interface/grpc"
	httpexport "github.com/HuynhHoangPhuc/myrmex/services/module-analytics/internal/interface/http"
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	// 4. Repository
	repo := persistence.NewAnalyticsRepository(pool)

	// 5. Start event consumers (optional — analytics is best-effort)
	msgConsumer := newConsumer(ctx, v, zapLog)
	if msgConsumer != nil {
		consumer := messaging.NewConsumer(msgConsumer, repo, zapLog)
		if err := consumer.Start(ctx); err != nil {
			zapLog.Warn("analytics consumer start failed", zap.Error(err))
		} else {
			defer msgConsumer.Close() //nolint:errcheck
			zapLog.Info("analytics consumers started")
		}
	} else {
		zapLog.Warn("messaging consumer not configured, analytics consumers disabled")
	}

	// 6. Query handlers
	workloadHandler := query.NewGetWorkloadHandler(repo)
	utilizationHandler := query.NewGetUtilizationHandler(repo)
	dashboardHandler := query.NewGetDashboardSummaryHandler(repo)

	// 7. HTTP routes
	server := httpif.NewAnalyticsHTTPServer(workloadHandler, utilizationHandler, dashboardHandler, repo, zapLog)
	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	pdfGen := export.NewPDFGenerator(repo)
	excelGen := export.NewExcelGenerator(repo)
	exportHandler := httpexport.NewExportHandler(pdfGen, excelGen)
	mux.HandleFunc("GET /export", exportHandler.HandleExport)

	// 8. Start HTTP server
	httpPort := v.GetInt("server.http_port")
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", httpPort),
		Handler: mux,
	}

	go func() {
		zapLog.Info("starting Analytics HTTP server", zap.Int("port", httpPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLog.Error("http server error", zap.Error(err))
			os.Exit(1)
		}
	}()

	// 9. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zapLog.Info("shutting down Analytics module...")

	cancel()
	if err := httpServer.Shutdown(context.Background()); err != nil {
		zapLog.Error("http shutdown error", zap.Error(err))
	}
	zapLog.Info("Analytics module stopped")
}

func newConsumer(ctx context.Context, v *viper.Viper, log *zap.Logger) pkgmessaging.Consumer {
	switch v.GetString("messaging.backend") {
	case "pubsub":
		c, err := msgpubsub.NewConsumer(ctx, v.GetString("gcp.project_id"), log)
		if err != nil {
			log.Warn("Pub/Sub consumer unavailable", zap.Error(err))
			return nil
		}
		log.Info("messaging consumer: Cloud Pub/Sub")
		return c
	default: // "nats" or ""
		natsURL := v.GetString("nats.url")
		if natsURL == "" {
			return nil
		}
		c, err := msgnats.NewConsumer(natsURL)
		if err != nil {
			log.Warn("NATS consumer unavailable", zap.Error(err))
			return nil
		}
		log.Info("messaging consumer: NATS")
		return c
	}
}
