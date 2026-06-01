package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	redisdb "github.com/d4interactive/contentstudio-social-analytics-go/src/db/redis"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/listening/scheduler"
)

const defaultSchedulerInterval = 15 * time.Minute

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	log := logger.New(cfg.LogLevel)
	log.Info().Msg("Starting Listening Scheduler")

	mongoClient, err := connectMongo(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer mongoClient.Disconnect(context.Background())

	producer, err := kafka.NewProducer(cfg.Kafka, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka producer")
	}
	defer producer.Close()

	redisClient := initRedis(cfg, log)
	defer redisClient.Close()

	repo := mongodb.NewListeningRepository(mongoClient.Database(cfg.Mongo.Database), log)
	workspaceRepo := mongodb.NewListeningWorkspaceRepository(mongoClient.Database(cfg.Mongo.Database), log)
	lock := redisdb.NewDistributedLock(redisClient, log.Logger)
	lockTTL := time.Duration(cfg.Listening.LockTTLMin) * time.Minute
	interval := schedulerInterval(cfg.Listening.SchedulerIntervalSec)
	if lockTTL <= 0 || lockTTL < interval {
		lockTTL = interval
	}

	sched := scheduler.NewRecurringScheduler(repo, producer, lock, log, lockTTL).
		WithSuperAdminResolver(workspaceRepo).
		WithOwnerQuotaChecker(workspaceRepo)

	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Info().Msg("Shutdown signal received")
		cancel()
	}()

	runPass(runCtx, sched, log)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Info().
		Dur("interval", interval).
		Dur("lock_ttl", lockTTL).
		Msg("Recurring listening scheduler loop started")

	for {
		select {
		case <-runCtx.Done():
			log.Info().Msg("Listening Scheduler stopped")
			return
		case <-ticker.C:
			runPass(runCtx, sched, log)
		}
	}
}

func runPass(ctx context.Context, sched *scheduler.RecurringScheduler, log *logger.Logger) {
	stats, err := sched.RunOnce(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Recurring listening scheduler pass failed")
		return
	}

	log.Debug().
		Int("total", stats.Total).
		Int("produced", stats.Produced).
		Int("failed", stats.Failed).
		Msg("Recurring listening scheduler pass finished")
}

func schedulerInterval(seconds int) time.Duration {
	if seconds <= 0 {
		return defaultSchedulerInterval
	}
	return time.Duration(seconds) * time.Second
}

func connectMongo(cfg *config.Config) (*mongo.Client, error) {
	return mongo.Connect(context.Background(), mongoClientOptions(cfg))
}

func mongoClientOptions(cfg *config.Config) *options.ClientOptions {
	clientOpts := options.Client().ApplyURI(cfg.Mongo.URI)
	if cfg.Mongo.Username != "" {
		clientOpts.SetAuth(options.Credential{
			Username:   cfg.Mongo.Username,
			Password:   cfg.Mongo.Password,
			AuthSource: cfg.Mongo.Database,
		})
	}

	return clientOpts
}

// schedulerRedis is the set of Redis methods required by the listening-scheduler
// (distributed lock + graceful shutdown).
type schedulerRedis interface {
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	CompareAndDelete(ctx context.Context, key, expected string) (bool, error)
	Close() error
}

func initRedis(cfg *config.Config, log *logger.Logger) schedulerRedis {
	if cfg.Redis.Addr == "" {
		log.Warn().Msg("APP_REDIS_ADDR not set; using in-memory mock Redis for scheduler lock")
		return &redisdb.MockRedisClient{}
	}

	client, err := redisdb.NewClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Redis connection failed; cannot start listening scheduler")
	}

	return client
}
