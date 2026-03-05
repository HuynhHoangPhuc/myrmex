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
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	pkgmessaging "github.com/HuynhHoangPhuc/myrmex/pkg/messaging"
	msgnats "github.com/HuynhHoangPhuc/myrmex/pkg/messaging/nats"
	msgpubsub "github.com/HuynhHoangPhuc/myrmex/pkg/messaging/pubsub"
	msgredis "github.com/HuynhHoangPhuc/myrmex/pkg/messaging/redis"
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

	poolCfg, err := pgxpool.ParseConfig(v.GetString("database.url"))
	if err != nil {
		zapLog.Fatal("parse database config", zap.Error(err))
	}
	poolCfg.MaxConns = 20
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

	// 4. Repositories
	notifRepo := persistence.NewNotificationRepository(pool)
	emailQueueRepo := persistence.NewEmailQueueRepository(pool)
	prefRepo := persistence.NewPreferenceRepository(pool)

	// 5. Redis client (optional — used for WS push relay)
	var rdb *redis.Client
	if redisAddr := strings.TrimSpace(v.GetString("redis.addr")); redisAddr != "" {
		rdb = redis.NewClient(&redis.Options{
			Addr:     redisAddr,
			Password: v.GetString("redis.password"),
			DB:       v.GetInt("redis.db"),
		})
		if err := rdb.Ping(ctx).Err(); err != nil {
			zapLog.Warn("Redis unavailable, WS push relay disabled", zap.Error(err))
			rdb = nil
		} else {
			defer rdb.Close()
			zapLog.Info("connected to Redis for WS push relay", zap.String("addr", redisAddr))
		}
	} else {
		zapLog.Warn("redis.addr not configured, WS push relay disabled")
	}

	// 6. Push publisher (Redis → core WSBroker → WebSocket clients)
	pushPub := msgredis.NewPushPublisher(rdb)
	pushPublisher := messaging.NewRedisPushPublisher(pushPub, zapLog)

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
	dispatchCmd := command.NewDispatchNotificationCommand(notifRepo, prefRepo, pushPublisher, zapLog)

	// 9. Messaging (optional — announcement publisher + event consumer)
	announcePub := newPublisher(ctx, v, zapLog, pkgmessaging.StreamConfig{
		Name: "NOTIFICATION_EVENTS", Subjects: []string{"notification.>"},
	})
	defer announcePub.Close() //nolint:errcheck

	msgConsumer := newConsumer(ctx, v, zapLog)
	if msgConsumer != nil {
		defer msgConsumer.Close() //nolint:errcheck
		resolver := messaging.NewRecipientResolver(pool)
		consumer := messaging.NewEventConsumer(msgConsumer, dispatchCmd, renderer, emailQueueRepo, resolver, zapLog)
		if err := consumer.Start(ctx); err != nil {
			zapLog.Warn("event consumer start failed", zap.Error(err))
		} else {
			zapLog.Info("notification event consumer started")
		}
	} else {
		zapLog.Warn("messaging consumer not configured, event consumer disabled")
	}

	// 10. HTTP handlers + router
	notifHandler := httpif.NewNotificationHandler(notifRepo, zapLog)
	prefHandler := httpif.NewPreferenceHandler(prefRepo, zapLog)
	announceHandler := httpif.NewAnnouncementHandler(announcePub, zapLog)
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
