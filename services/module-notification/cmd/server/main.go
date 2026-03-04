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

	"github.com/jackc/pgx/v5/pgxpool"
	natsgo "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/HuynhHoangPhuc/myrmex/services/module-notification/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/module-notification/internal/infrastructure/email"
	"github.com/HuynhHoangPhuc/myrmex/services/module-notification/internal/infrastructure/messaging"
	"github.com/HuynhHoangPhuc/myrmex/services/module-notification/internal/infrastructure/persistence"
	httpif "github.com/HuynhHoangPhuc/myrmex/services/module-notification/internal/interface/http"
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

	pool, err := pgxpool.New(ctx, v.GetString("database.url"))
	if err != nil {
		zapLog.Fatal("connect to database", zap.Error(err))
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		zapLog.Fatal("ping database", zap.Error(err))
	}
	zapLog.Info("connected to database")

	// 4. Connect to NATS (optional — gracefully degrade when unavailable)
	var nc *natsgo.Conn
	var js jetstream.JetStream
	if natsURL := v.GetString("nats.url"); natsURL != "" {
		nc, err = natsgo.Connect(natsURL)
		if err != nil {
			zapLog.Warn("NATS unavailable, event consumer disabled", zap.Error(err))
		} else {
			defer nc.Close()
			js, err = jetstream.New(nc)
			if err != nil {
				zapLog.Warn("JetStream init failed, event consumer disabled", zap.Error(err))
				js = nil
			} else {
				zapLog.Info("connected to NATS JetStream")
			}
		}
	} else {
		zapLog.Warn("NATS URL not configured, skipping connection")
	}

	// 5. Repositories
	notifRepo := persistence.NewNotificationRepository(pool)
	emailQueueRepo := persistence.NewEmailQueueRepository(pool)
	prefRepo := persistence.NewPreferenceRepository(pool)

	// 6. NATS publisher (WS push relay)
	publisher := messaging.NewNATSPublisher(nc, zapLog)

	// 7. Email infrastructure
	renderer, err := email.NewTemplateRenderer()
	if err != nil {
		zapLog.Fatal("init email template renderer", zap.Error(err))
	}

	smtpSvc := email.NewSMTPService(email.SMTPConfig{
		Host:      v.GetString("smtp.host"),
		Port:      v.GetInt("smtp.port"),
		Username:  v.GetString("smtp.username"),
		Password:  v.GetString("smtp.password"),
		FromEmail: v.GetString("smtp.from_email"),
		FromName:  v.GetString("smtp.from_name"),
	}, zapLog)

	queueWorker := email.NewQueueWorker(emailQueueRepo, smtpSvc, zapLog)
	go queueWorker.Start(ctx)

	// 8. Dispatch command (preference-filtered notification dispatcher)
	dispatchCmd := command.NewDispatchNotificationCommand(notifRepo, prefRepo, publisher, zapLog)

	// 9. Event consumer (JetStream — optional, skipped when NATS unavailable)
	if js != nil {
		resolver := messaging.NewRecipientResolver(pool)
		consumer := messaging.NewEventConsumer(js, dispatchCmd, renderer, emailQueueRepo, resolver, zapLog)
		if err := consumer.Start(ctx); err != nil {
			zapLog.Warn("event consumer start failed", zap.Error(err))
		}
	}

	// 10. HTTP handlers + router
	notifHandler := httpif.NewNotificationHandler(notifRepo, zapLog)
	prefHandler := httpif.NewPreferenceHandler(prefRepo, zapLog)
	announceHandler := httpif.NewAnnouncementHandler(nc, zapLog)
	mux := httpif.NewRouter(notifHandler, prefHandler, announceHandler)

	// 11. Start HTTP server
	httpPort := v.GetInt("server.http_port")
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", httpPort),
		Handler: mux,
	}

	go func() {
		zapLog.Info("starting Notification HTTP server", zap.Int("port", httpPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLog.Error("http server error", zap.Error(err))
			os.Exit(1)
		}
	}()

	// 12. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zapLog.Info("shutting down Notification module...")

	cancel()
	if err := httpServer.Shutdown(context.Background()); err != nil {
		zapLog.Error("http shutdown error", zap.Error(err))
	}
	zapLog.Info("Notification module stopped")
}
