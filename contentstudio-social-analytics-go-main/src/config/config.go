package config

import (
	"strings"

	"github.com/joho/godotenv"  // To load .env files
	"github.com/rs/zerolog/log" // Using global logger for config loading issues
	"github.com/spf13/viper"
)

const dotEnvPath = ".env" // Path relative to the project root where the binary runs

// MongoConfig holds MongoDB specific configuration.
// Note: `mapstructure` tags are used by Viper to unmarshal values into the struct.
// `env` tags could be used with a library like `caarlos0/env` if direct env var mapping is preferred
// over Viper's automatic env var binding.

// MongoConfig holds MongoDB specific configuration.
type MongoConfig struct {
	URI      string `mapstructure:"uri"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	// Add other MongoDB options like Timeout, PoolSize etc. if needed
}

// SASLConfig holds Kafka SASL authentication configuration.
type SASLConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	Mechanism string `mapstructure:"mechanism"` // e.g., "SCRAM-SHA-256", "SCRAM-SHA-512", "PLAIN"
}

// SentryConfig holds Sentry telemetry configuration.
type SentryConfig struct {
	DSN              string  `mapstructure:"dsn"`
	Environment      string  `mapstructure:"environment"`
	Release          string  `mapstructure:"release"`
	Debug            bool    `mapstructure:"debug"`
	EnableTracing    bool    `mapstructure:"enable_tracing"`
	TracesSampleRate float64 `mapstructure:"traces_sample_rate"`
}

// KafkaConfig holds Kafka specific configuration.
type KafkaConfig struct {
	Brokers     []string   `mapstructure:"brokers"` // Comma-separated list of brokers from env, e.g., "host1:9092,host2:9092"
	TopicPrefix string     `mapstructure:"topic_prefix"`
	SASL        SASLConfig `mapstructure:"sasl"`
}

// FacebookConfig holds Facebook specific configuration.
type FacebookConfig struct {
	AppID                 string  `mapstructure:"app_id"`
	AppSecret             string  `mapstructure:"app_secret"`
	AppToken              string  `mapstructure:"app_token"`
	PerTokenRPS           float64 `mapstructure:"per_token_rps"`
	PerTokenBurst         int     `mapstructure:"per_token_burst"`
	GlobalRPS             float64 `mapstructure:"global_rps"`
	GlobalBurst           int     `mapstructure:"global_burst"`
	PerAccountConcurrency float64 `mapstructure:"per_account_concurrency"`
}

// ClickHouseConfig holds ClickHouse specific configuration.
type ClickHouseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Database     string `mapstructure:"database"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Secure       bool   `mapstructure:"secure"`
	Compression  bool   `mapstructure:"compression"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	// MaxExecutionTimeInSec configures ClickHouse max_execution_time query setting (seconds).
	// When zero or negative we fall back to the historical default of 60 seconds.
	MaxExecutionTimeInSec int `mapstructure:"max_execution_time_in_sec"`
}

// RedisConfig holds Redis specific configuration.
type RedisConfig struct {
	Addr       string `mapstructure:"addr"`
	Password   string `mapstructure:"password"`
	DB         int    `mapstructure:"db"`
	MaxRetries int    `mapstructure:"max_retries"`
	PoolSize   int    `mapstructure:"pool_size"`
}

// EmailConfig holds email notification configuration.
type EmailConfig struct {
	SMTPHost     string `mapstructure:"smtp_host"`
	SMTPPort     int    `mapstructure:"smtp_port"`
	SMTPUsername string `mapstructure:"smtp_username"`
	SMTPPassword string `mapstructure:"smtp_password"`
	FromEmail    string `mapstructure:"from_email"`
	BackendURL   string `mapstructure:"backend_url"`
}

// PusherConfig holds Pusher real-time notification configuration.
type PusherConfig struct {
	AppID   string `mapstructure:"app_id"`
	Key     string `mapstructure:"key"`
	Secret  string `mapstructure:"secret"`
	Cluster string `mapstructure:"cluster"`
}

// TikTokConfig holds TikTok specific configuration.
type TikTokConfig struct {
	ClientKey    string `mapstructure:"client_key"`
	ClientSecret string `mapstructure:"client_secret"`
}

// YouTubeConfig holds YouTube specific configuration.
type YouTubeConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}

// TwitterConfig holds Twitter/X specific configuration.
type TwitterConfig struct {
	ConsumerKey    string `mapstructure:"consumer_key"`
	ConsumerSecret string `mapstructure:"consumer_secret"`
}

// GMBConfig holds Google My Business specific configuration.
type GMBConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}

// JWTConfig holds JWT authentication configuration.
type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	AdminSecret string `mapstructure:"admin_secret_key"`
	Algorithm   string `mapstructure:"algorithm"`
	Issuer      string `mapstructure:"issuer"`
}

// S3Config holds AWS S3 configuration for file storage.
type S3Config struct {
	Region          string `mapstructure:"region"`
	Bucket          string `mapstructure:"bucket"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	Endpoint        string `mapstructure:"endpoint"`
	CDNBaseURL      string `mapstructure:"cdn_base_url"`
}

// AIAgentsConfig holds configuration for the external AI agents microservice.
type AIAgentsConfig struct {
	BaseURL string `mapstructure:"base_url"`
	APIKey  string `mapstructure:"api_key"`
	Timeout int    `mapstructure:"timeout"`
}

// LookerStudioConfig holds configuration for the Looker Studio community connector.
type LookerStudioConfig struct {
	ConnectorID string `mapstructure:"connector_id"`
	Env         string `mapstructure:"env"`
}

// Data365Config holds Data365 API configuration.
type Data365Config struct {
	BaseURL      string  `mapstructure:"base_url"`
	AccessToken  string  `mapstructure:"access_token"`
	RPS          float64 `mapstructure:"rps"`
	Burst        int     `mapstructure:"burst"`
	PollInterval int     `mapstructure:"poll_interval"`
	PollTimeout  int     `mapstructure:"poll_timeout"`
}

// URLRefresherConfig holds rate-limit overrides for the url-refresher job.
// When non-zero these take precedence over the shared FacebookConfig values,
// so the URL refresher can be tuned independently without affecting fetchers.
type URLRefresherConfig struct {
	GlobalRPS     float64 `mapstructure:"global_rps"`
	GlobalBurst   int     `mapstructure:"global_burst"`
	PerTokenRPS   float64 `mapstructure:"per_token_rps"`
	PerTokenBurst int     `mapstructure:"per_token_burst"`
}

// ListeningConfig holds listening pipeline configuration.
//
// BatchSize and BatchSizeInitial are per-keyword caps on max_posts the fetcher
// requests from Data365 per cycle. The fetcher chooses BatchSizeInitial for
// work orders with SyncType="initial" (first crawl of a topic) and BatchSize
// otherwise. Both are caps on a single fetch cycle, NOT a hard ceiling on
// total mentions a topic can collect — multiple cycles + cursor resumption
// accumulate over time.
type ListeningConfig struct {
	BatchSize                       int `mapstructure:"batch_size"`
	BatchSizeInitial                int `mapstructure:"batch_size_initial"`
	DedupTTLHours                   int `mapstructure:"dedup_ttl_hours"`
	MaxRetries                      int `mapstructure:"max_retries"`
	LockTTLMin                      int `mapstructure:"lock_ttl_min"`
	SchedulerIntervalSec            int `mapstructure:"scheduler_interval_sec"`
	EnrichmentBackfillIntervalSec   int `mapstructure:"enrichment_backfill_interval_sec"`
	EnrichmentBackfillLookbackHours int `mapstructure:"enrichment_backfill_lookback_hours"`
}

// Config holds all application configuration.
type Config struct {
	Environment            string             `mapstructure:"environment"`
	LogLevel               string             `mapstructure:"log_level"`
	TwitterJobLogsURL      string             `mapstructure:"twitter_job_logs_url"`
	Mongo                  MongoConfig        `mapstructure:"mongo"`
	Kafka                  KafkaConfig        `mapstructure:"kafka"`
	Sentry                 SentryConfig       `mapstructure:"sentry"`
	Facebook               FacebookConfig     `mapstructure:"facebook"`
	ClickHouse             ClickHouseConfig   `mapstructure:"clickhouse"`
	Redis                  RedisConfig        `mapstructure:"redis"`
	Email                  EmailConfig        `mapstructure:"email"`
	Pusher                 PusherConfig       `mapstructure:"pusher"`
	TikTok                 TikTokConfig       `mapstructure:"tiktok"`
	YouTube                YouTubeConfig      `mapstructure:"youtube"`
	Twitter                TwitterConfig      `mapstructure:"twitter"`
	GMB                    GMBConfig          `mapstructure:"gmb"`
	S3                     S3Config           `mapstructure:"s3"`
	JWT                    JWTConfig          `mapstructure:"jwt"`
	AIAgents               AIAgentsConfig     `mapstructure:"ai_agents"`
	LookerStudio           LookerStudioConfig `mapstructure:"looker_studio"`
	Data365                Data365Config      `mapstructure:"data365"`
	Listening              ListeningConfig    `mapstructure:"listening"`
	URLRefresher           URLRefresherConfig `mapstructure:"url_refresher"`
	APIKey                 string             `mapstructure:"api_key"`
	DecryptionKey          string             `mapstructure:"decryption_key"`
	SinkIdleTimeoutMinutes int                `mapstructure:"sink_idle_timeout_minutes"`
	// Add other configurations like Redis, Server ports etc. here
}

// LoadConfig loads configuration from environment variables and/or a config file.
// It uses Viper for flexibility.
func LoadConfig() (*Config, error) {
	// Attempt to load .env file.
	// If it's not found, godotenv.Load doesn't return an error by default,
	// allowing the application to proceed with actual environment variables or defaults.
	err := godotenv.Load(dotEnvPath)
	if err != nil {
		// Log a warning if .env file was expected but not found or failed to load for other reasons.
		// Don't make this fatal, as env vars might be set in the environment directly.
		log.Warn().Err(err).Str("path", dotEnvPath).Msg("Error loading .env file, will rely on environment variables or defaults")
	} else {
		log.Info().Str("path", dotEnvPath).Msg(".env file loaded successfully")
	}

	var cfg Config

	//// Set default values
	//viper.SetDefault("ENVIRONMENT", "development")
	//viper.SetDefault("LOG_LEVEL", "info")
	//viper.SetDefault("MONGO.URI", "mongodb://localhost:27017")
	//viper.SetDefault("MONGO.DATABASE", "social_analytics")
	//viper.SetDefault("KAFKA.BROKERS", "localhost:9092") // Default single broker, env var can be comma-separated for multiple
	//viper.SetDefault("KAFKA.TOPIC_PREFIX", "")
	//viper.SetDefault("KAFKA.SASL.ENABLED", false)
	//viper.SetDefault("KAFKA.SASL.USERNAME", "")
	//viper.SetDefault("KAFKA.SASL.PASSWORD", "")
	//viper.SetDefault("KAFKA.SASL.MECHANISM", "SCRAM-SHA-512")
	//viper.SetDefault("FACEBOOK.APP_ID", "")
	//viper.SetDefault("FACEBOOK.APP_SECRET", "")
	//viper.SetDefault("CLICKHOUSE.HOST", "")
	//viper.SetDefault("CLICKHOUSE.PORT", 9000)
	//viper.SetDefault("CLICKHOUSE.DATABASE", "")
	//viper.SetDefault("CLICKHOUSE.USERNAME", "")
	//viper.SetDefault("CLICKHOUSE.PASSWORD", "")
	//viper.SetDefault("CLICKHOUSE.SECURE", false)
	//viper.SetDefault("CLICKHOUSE.COMPRESSION", false)
	//viper.SetDefault("CLICKHOUSE.MAX_OPEN_CONNS", 10)
	//viper.SetDefault("CLICKHOUSE.MAX_IDLE_CONNS", 5)
	//viper.SetDefault("REDIS.ADDR", "")
	//viper.SetDefault("REDIS.PASSWORD", "")
	//viper.SetDefault("REDIS.DB", 0)
	//viper.SetDefault("REDIS.MAX_RETRIES", 3)
	//viper.SetDefault("REDIS.POOL_SIZE", 10)
	//viper.SetDefault("EMAIL.SMTP_HOST", "")
	//viper.SetDefault("EMAIL.SMTP_PORT", 587)
	//viper.SetDefault("EMAIL.SMTP_USERNAME", "")
	//viper.SetDefault("EMAIL.SMTP_PASSWORD", "")
	//viper.SetDefault("EMAIL.FROM_EMAIL", "")
	//viper.SetDefault("PUSHER.APP_ID", "")
	//viper.SetDefault("PUSHER.KEY", "")
	//viper.SetDefault("PUSHER.SECRET", "")
	//viper.SetDefault("PUSHER.CLUSTER", "")
	//viper.SetDefault("TIKTOK.CLIENT_KEY", "")
	//viper.SetDefault("TIKTOK.CLIENT_SECRET", "")
	//viper.SetDefault("DECRYPTION_KEY", "")

	// Configure Viper to read environment variables
	// It can automatically override config file settings with env vars.
	// Example: MONGO.URI can be overridden by APP_MONGO_URI (if SetEnvPrefix is APP)
	// or MONGO_URI (if SetEnvKeyReplacer is used to replace . with _)
	//viper.SetEnvPrefix("APP") // Optional: if you want all env vars to be like APP_MONGO_URI
	//viper.AutomaticEnv()      // Read all environment variables

	// Replace . with _ in environment variable names (e.g., MONGO.URI becomes MONGO_URI)
	// This is often more conventional for environment variables.

	// Re-apply defaults and settings to this new Viper instance
	v := viper.New()
	//v.SetDefault("ENVIRONMENT", "development")
	//v.SetDefault("LOG_LEVEL", "info")
	//v.SetDefault("MONGO.URI", "mongodb://mongodb:27017")
	//v.SetDefault("MONGO.DATABASE", "social_analytics")
	//v.SetDefault("KAFKA.BROKERS", "kafka:9092")
	//v.SetDefault("KAFKA.TOPIC_PREFIX", "")
	//v.SetDefault("KAFKA.SASL.ENABLED", false)
	//v.SetDefault("KAFKA.SASL.USERNAME", "")
	//v.SetDefault("KAFKA.SASL.PASSWORD", "")
	//v.SetDefault("KAFKA.SASL.MECHANISM", "SCRAM-SHA-512")
	//v.SetDefault("FACEBOOK.APP_ID", "")
	//v.SetDefault("FACEBOOK.APP_SECRET", "")
	//v.SetDefault("CLICKHOUSE.HOST", "schema")
	//v.SetDefault("CLICKHOUSE.PORT", 9000)
	//v.SetDefault("CLICKHOUSE.DATABASE", "")
	//v.SetDefault("CLICKHOUSE.USERNAME", "")
	//v.SetDefault("CLICKHOUSE.PASSWORD", "")
	//v.SetDefault("CLICKHOUSE.SECURE", false)
	//v.SetDefault("CLICKHOUSE.COMPRESSION", false)
	//v.SetDefault("CLICKHOUSE.MAX_OPEN_CONNS", 10)
	//v.SetDefault("CLICKHOUSE.MAX_IDLE_CONNS", 5)
	//v.SetDefault("REDIS.ADDR", "")
	//v.SetDefault("REDIS.PASSWORD", "")
	//v.SetDefault("REDIS.DB", 0)
	//v.SetDefault("REDIS.MAX_RETRIES", 3)
	//v.SetDefault("REDIS.POOL_SIZE", 10)
	//v.SetDefault("EMAIL.SMTP_HOST", "")
	//v.SetDefault("EMAIL.SMTP_PORT", 587)
	//v.SetDefault("EMAIL.SMTP_USERNAME", "")
	//v.SetDefault("EMAIL.SMTP_PASSWORD", "")
	//v.SetDefault("EMAIL.FROM_EMAIL", "")
	//v.SetDefault("PUSHER.APP_ID", "")
	//v.SetDefault("PUSHER.KEY", "")
	//v.SetDefault("PUSHER.SECRET", "")
	//v.SetDefault("PUSHER.CLUSTER", "")
	//v.SetDefault("TIKTOK.CLIENT_KEY", "")
	//v.SetDefault("TIKTOK.CLIENT_SECRET", "")
	//v.SetDefault("DECRYPTION_KEY", "")

	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	BindEnvs(v)

	// You can also add config file reading here if needed:
	// viper.SetConfigName("config") // name of config file (without extension)
	// viper.SetConfigType("yaml")   // or json, toml, etc.
	// viper.AddConfigPath(".")      // look for config in the working directory
	// viper.AddConfigPath("/etc/appname/") // path to look for the config file in
	// if err := viper.ReadInConfig(); err != nil {
	// 	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
	// 		// Config file not found; ignore error if desired or load from env only
	// 		log.Warn().Msg("Config file not found, loading from environment variables and defaults.")
	// 	} else {
	// 		// Config file was found but another error was produced
	// 		return nil, fmt.Errorf("failed to read config file: %w", err)
	// 	}
	// }

	// Unmarshal the configuration into the Config struct
	if err := v.Unmarshal(&cfg); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal configuration")
		return nil, err
	}

	// Post-process Kafka brokers: Viper reads the env var as a single string,
	// so we need to split it into a slice.
	kafkaBrokersStr := v.GetString("KAFKA.BROKERS") // Viper uses . for keys, auto-handles prefix APP_ and _ replacer

	if kafkaBrokersStr != "" {
		cfg.Kafka.Brokers = strings.Split(kafkaBrokersStr, ",")

	} else {
		// If KAFKA.BROKERS was empty or not set, and default was also empty (though we set a default)
		// ensure it's an empty slice, not nil, for consistency, though our default prevents this.
		cfg.Kafka.Brokers = []string{}
	}

	log.Info().Interface("loaded_config", cfg).Msg("Configuration loaded successfully")
	return &cfg, nil
}

func BindEnvs(v *viper.Viper) {
	bind := func(key string) {
		if err := v.BindEnv(key); err != nil {
			log.Warn().Err(err).Msgf("Failed to bind env for key: %s", key)
		}
	}

	// Top-level
	bind("environment")
	bind("log_level")
	bind("decryption_key")
	bind("twitter_job_logs_url")

	// Mongo
	bind("mongo.uri")
	bind("mongo.database")
	bind("mongo.username")
	bind("mongo.password")

	// Kafka
	bind("kafka.brokers")
	bind("kafka.topic_prefix")

	// Kafka SASL
	bind("kafka.sasl.enabled")
	bind("kafka.sasl.username")
	bind("kafka.sasl.password")
	bind("kafka.sasl.mechanism")

	// Sentry
	bind("sentry.dsn")
	bind("sentry.environment")
	bind("sentry.release")
	bind("sentry.debug")
	bind("sentry.enable_tracing")
	bind("sentry.traces_sample_rate")

	// Facebook
	bind("facebook.app_id")
	bind("facebook.app_secret")
	bind("facebook.app_token")

	// ClickHouse
	bind("clickhouse.host")
	bind("clickhouse.port")
	bind("clickhouse.username")
	bind("clickhouse.password")
	bind("clickhouse.database")
	bind("clickhouse.secure")
	bind("clickhouse.compression")
	bind("clickhouse.max_open_conns")
	bind("clickhouse.max_idle_conns")
	bind("clickhouse.max_execution_time_in_sec")

	// Redis
	bind("redis.addr")
	bind("redis.password")
	bind("redis.db")
	bind("redis.max_retries")
	bind("redis.pool_size")

	// Email
	bind("email.smtp_host")
	bind("email.smtp_port")
	bind("email.smtp_username")
	bind("email.smtp_password")
	bind("email.from_email")
	bind("email.backend_url")

	// Pusher
	bind("pusher.app_id")
	bind("pusher.key")
	bind("pusher.secret")
	bind("pusher.cluster")

	// TikTok
	bind("tiktok.client_key")
	bind("tiktok.client_secret")

	// YouTube
	bind("youtube.client_id")
	bind("youtube.client_secret")

	// Twitter
	bind("twitter.consumer_key")
	bind("twitter.consumer_secret")

	// GMB (Google My Business)
	bind("gmb.client_id")
	bind("gmb.client_secret")

	// S3
	bind("s3.region")
	bind("s3.bucket")
	bind("s3.access_key_id")
	bind("s3.secret_access_key")
	bind("s3.endpoint")
	bind("s3.cdn_base_url")

	// JWT
	bind("jwt.secret")
	bind("jwt.algorithm")
	bind("jwt.issuer")
	if err := v.BindEnv("jwt.admin_secret_key", "JWT_ADMIN_SECRET_KEY"); err != nil {
		log.Warn().Err(err).Msg("Failed to bind env for key: jwt.admin_secret_key")
	}

	// AI Agents
	bind("ai_agents.base_url")
	bind("ai_agents.api_key")
	bind("ai_agents.timeout")

	// Data365
	bind("data365.base_url")
	bind("data365.access_token")
	bind("data365.rps")
	bind("data365.burst")
	bind("data365.poll_interval")
	bind("data365.poll_timeout")

	// Listening
	bind("listening.batch_size")
	bind("listening.batch_size_initial")
	bind("listening.dedup_ttl_hours")
	bind("listening.max_retries")
	bind("listening.lock_ttl_min")
	bind("listening.scheduler_interval_sec")
	bind("listening.enrichment_backfill_interval_sec")
	bind("listening.enrichment_backfill_lookback_hours")

	// URL Refresher (dedicated rate limits — override Facebook.* for the url-refresher job only)
	bind("url_refresher.global_rps")
	bind("url_refresher.global_burst")
	bind("url_refresher.per_token_rps")
	bind("url_refresher.per_token_burst")

	// API Key
	bind("api_key")

	// Sink
	bind("sink_idle_timeout_minutes")

	// Looker Studio
	bind("looker_studio.connector_id")
	bind("looker_studio.env")
}
