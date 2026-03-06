package main

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	pkgnats "github.com/HuynhHoangPhuc/myrmex/pkg/nats"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// initConfig loads config from file + env into a new Viper instance.
func initConfig() (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigName("local")
	v.AddConfigPath("config")
	v.AddConfigPath(".")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	return v, nil
}

// initLogger creates a zap development logger.
func initLogger() (*zap.Logger, error) {
	return zap.NewDevelopment()
}

// initDatabase connects to Postgres using the URL from config and returns a connection pool.
// Pool is tuned for core's HTTP/WS traffic (max 200 total DB connections across services).
func initDatabase(ctx context.Context, v *viper.Viper, log *zap.Logger) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(v.GetString("database.url"))
	if err != nil {
		return nil, err
	}
	poolCfg.MaxConns = 30
	poolCfg.MinConns = 3
	poolCfg.MaxConnLifetime = 30 * time.Minute
	poolCfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	log.Info("connected to database")
	return pool, nil
}

// natsResult groups the outputs of a successful NATS connection.
type natsResult struct {
	NC        *nats.Conn
	JS        jetstream.JetStream
	Publisher *pkgnats.Publisher
}

// initNATS connects to NATS JetStream and ensures required streams exist.
// Returns nil when NATS is unavailable — callers must nil-check before use.
func initNATS(ctx context.Context, v *viper.Viper, log *zap.Logger) *natsResult {
	nc, js, err := pkgnats.Connect(v.GetString("nats.url"))
	if err != nil {
		log.Warn("nats connection failed, continuing without events", zap.Error(err))
		return nil
	}

	publisher := pkgnats.NewPublisher(js)

	if _, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     "TIMETABLE",
		Subjects: []string{"timetable.schedule.>"},
	}); err != nil {
		log.Warn("create TIMETABLE stream failed", zap.Error(err))
	}
	if _, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     "CORE_EVENTS",
		Subjects: []string{"core.>"},
	}); err != nil {
		log.Warn("create CORE_EVENTS stream failed", zap.Error(err))
	}
	log.Info("connected to NATS")

	return &natsResult{NC: nc, JS: js, Publisher: publisher}
}

// toNATSPublisher safely extracts the Publisher from a natsResult (nil-safe).
func toNATSPublisher(n *natsResult) *pkgnats.Publisher {
	if n == nil {
		return nil
	}
	return n.Publisher
}

// toNATSJS safely extracts the JetStream from a natsResult (nil-safe).
func toNATSJS(n *natsResult) jetstream.JetStream {
	if n == nil {
		return nil
	}
	return n.JS
}

// initRedis connects to Redis and returns the client.
// Returns nil when Redis is unavailable or not configured.
func initRedis(ctx context.Context, v *viper.Viper, log *zap.Logger) *redis.Client {
	redisAddr := strings.TrimSpace(v.GetString("redis.addr"))
	if redisAddr == "" {
		return nil
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: v.GetString("redis.password"),
		DB:       v.GetInt("redis.db"),
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Warn("Redis unavailable, WS push relay disabled", zap.Error(err))
		return nil
	}
	log.Info("connected to Redis for WS push relay", zap.String("addr", redisAddr))
	return rdb
}
