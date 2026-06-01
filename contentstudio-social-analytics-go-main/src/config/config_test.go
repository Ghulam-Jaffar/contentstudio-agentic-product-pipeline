package config

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func Test_LoadConfig_Table(t *testing.T) {
	cases := []struct {
		name       string
		envVars    map[string]string
		expectErr  bool
		checkField func(*Config) bool
	}{
		{
			name:      "loads with no env vars",
			envVars:   map[string]string{},
			expectErr: false,
		},
		{
			name: "loads environment variable",
			envVars: map[string]string{
				"APP_ENVIRONMENT": "test",
			},
			expectErr: false,
			checkField: func(c *Config) bool {
				return c.Environment == "test"
			},
		},
		{
			name: "loads mongo config",
			envVars: map[string]string{
				"APP_MONGO_URI":      "mongodb://localhost:27017",
				"APP_MONGO_DATABASE": "test_db",
			},
			expectErr: false,
			checkField: func(c *Config) bool {
				return c.Mongo.URI == "mongodb://localhost:27017" && c.Mongo.Database == "test_db"
			},
		},
		{
			name: "loads kafka brokers as string",
			envVars: map[string]string{
				"APP_KAFKA_BROKERS": "broker1:9092,broker2:9092",
			},
			expectErr: false,
			checkField: func(c *Config) bool {
				return len(c.Kafka.Brokers) == 2 &&
					c.Kafka.Brokers[0] == "broker1:9092" &&
					c.Kafka.Brokers[1] == "broker2:9092"
			},
		},
		{
			name: "empty kafka brokers results in empty slice",
			envVars: map[string]string{
				"APP_KAFKA_BROKERS": "",
			},
			expectErr: false,
			checkField: func(c *Config) bool {
				return len(c.Kafka.Brokers) == 0
			},
		},
		{
			name: "loads clickhouse config",
			envVars: map[string]string{
				"APP_CLICKHOUSE_HOST":     "localhost",
				"APP_CLICKHOUSE_PORT":     "9000",
				"APP_CLICKHOUSE_DATABASE": "analytics",
			},
			expectErr: false,
			checkField: func(c *Config) bool {
				return c.ClickHouse.Host == "localhost" &&
					c.ClickHouse.Port == 9000 &&
					c.ClickHouse.Database == "analytics"
			},
		},
		{
			name: "loads sentry config",
			envVars: map[string]string{
				"APP_SENTRY_DSN":         "https://key@sentry.io/123",
				"APP_SENTRY_ENVIRONMENT": "production",
			},
			expectErr: false,
			checkField: func(c *Config) bool {
				return c.Sentry.DSN == "https://key@sentry.io/123" &&
					c.Sentry.Environment == "production"
			},
		},
		{
			name: "loads redis config",
			envVars: map[string]string{
				"APP_REDIS_ADDR":     "localhost:6379",
				"APP_REDIS_PASSWORD": "secret",
			},
			expectErr: false,
			checkField: func(c *Config) bool {
				return c.Redis.Addr == "localhost:6379" &&
					c.Redis.Password == "secret"
			},
		},
		{
			name: "loads decryption key",
			envVars: map[string]string{
				"APP_DECRYPTION_KEY": "my-secret-key",
			},
			expectErr: false,
			checkField: func(c *Config) bool {
				return c.DecryptionKey == "my-secret-key"
			},
		},
		{
			name: "loads twitter job logs url",
			envVars: map[string]string{
				"APP_TWITTER_JOB_LOGS_URL": "https://example.com/twitter-job-logs",
			},
			expectErr: false,
			checkField: func(c *Config) bool {
				return c.TwitterJobLogsURL == "https://example.com/twitter-job-logs"
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envVars {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tc.envVars {
					os.Unsetenv(k)
				}
			}()

			cfg, err := LoadConfig()

			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cfg == nil {
				t.Fatal("expected config, got nil")
			}

			if tc.checkField != nil && !tc.checkField(cfg) {
				t.Fatal("field check failed")
			}
		})
	}
}

func Test_BindEnvs_Table(t *testing.T) {
	cases := []struct {
		name   string
		envKey string
		envVal string
		cfgKey string
	}{
		{
			name:   "binds environment",
			envKey: "APP_ENVIRONMENT",
			envVal: "test_env",
			cfgKey: "environment",
		},
		{
			name:   "binds log_level",
			envKey: "APP_LOG_LEVEL",
			envVal: "debug",
			cfgKey: "log_level",
		},
		{
			name:   "binds twitter_job_logs_url",
			envKey: "APP_TWITTER_JOB_LOGS_URL",
			envVal: "https://example.com/twitter-job-logs",
			cfgKey: "twitter_job_logs_url",
		},
		{
			name:   "binds mongo.uri",
			envKey: "APP_MONGO_URI",
			envVal: "mongodb://test:27017",
			cfgKey: "mongo.uri",
		},
		{
			name:   "binds kafka.brokers",
			envKey: "APP_KAFKA_BROKERS",
			envVal: "kafka:9092",
			cfgKey: "kafka.brokers",
		},
		{
			name:   "binds clickhouse.host",
			envKey: "APP_CLICKHOUSE_HOST",
			envVal: "clickhouse",
			cfgKey: "clickhouse.host",
		},
		{
			name:   "binds sentry.dsn",
			envKey: "APP_SENTRY_DSN",
			envVal: "https://sentry.io",
			cfgKey: "sentry.dsn",
		},
		{
			name:   "binds facebook.app_id",
			envKey: "APP_FACEBOOK_APP_ID",
			envVal: "12345",
			cfgKey: "facebook.app_id",
		},
		{
			name:   "binds redis.addr",
			envKey: "APP_REDIS_ADDR",
			envVal: "redis:6379",
			cfgKey: "redis.addr",
		},
		{
			name:   "binds email.smtp_host",
			envKey: "APP_EMAIL_SMTP_HOST",
			envVal: "smtp.test.com",
			cfgKey: "email.smtp_host",
		},
		{
			name:   "binds pusher.app_id",
			envKey: "APP_PUSHER_APP_ID",
			envVal: "pusher123",
			cfgKey: "pusher.app_id",
		},
		{
			name:   "binds tiktok.client_key",
			envKey: "APP_TIKTOK_CLIENT_KEY",
			envVal: "tiktok_key",
			cfgKey: "tiktok.client_key",
		},
		{
			name:   "binds s3.bucket",
			envKey: "APP_S3_BUCKET",
			envVal: "my-bucket",
			cfgKey: "s3.bucket",
		},
		{
			name:   "binds listening.batch_size_initial",
			envKey: "APP_LISTENING_BATCH_SIZE_INITIAL",
			envVal: "5000",
			cfgKey: "listening.batch_size_initial",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv(tc.envKey, tc.envVal)
			defer os.Unsetenv(tc.envKey)

			v := viper.New()
			v.SetEnvPrefix("APP")
			v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

			BindEnvs(v)

			got := v.GetString(tc.cfgKey)
			if got != tc.envVal {
				t.Fatalf("expected %s, got %s", tc.envVal, got)
			}
		})
	}
}

func Test_Config_Structs(t *testing.T) {
	t.Run("MongoConfig fields", func(t *testing.T) {
		cfg := MongoConfig{
			URI:      "mongodb://localhost",
			Database: "test",
			Username: "user",
			Password: "pass",
		}
		if cfg.URI != "mongodb://localhost" {
			t.Fatal("URI mismatch")
		}
	})

	t.Run("KafkaConfig with SASL", func(t *testing.T) {
		cfg := KafkaConfig{
			Brokers:     []string{"broker1", "broker2"},
			TopicPrefix: "prefix_",
			SASL: SASLConfig{
				Enabled:   true,
				Username:  "user",
				Password:  "pass",
				Mechanism: "SCRAM-SHA-512",
			},
		}
		if !cfg.SASL.Enabled {
			t.Fatal("SASL should be enabled")
		}
	})

	t.Run("ClickHouseConfig fields", func(t *testing.T) {
		cfg := ClickHouseConfig{
			Host:                  "localhost",
			Port:                  9000,
			Database:              "test",
			Username:              "default",
			Password:              "",
			Secure:                true,
			Compression:           true,
			MaxOpenConns:          10,
			MaxIdleConns:          5,
			MaxExecutionTimeInSec: 60,
		}
		if cfg.Port != 9000 {
			t.Fatal("Port mismatch")
		}
	})

	t.Run("SentryConfig fields", func(t *testing.T) {
		cfg := SentryConfig{
			DSN:              "https://sentry.io",
			Environment:      "prod",
			Release:          "v1.0",
			Debug:            false,
			EnableTracing:    true,
			TracesSampleRate: 0.5,
		}
		if cfg.TracesSampleRate != 0.5 {
			t.Fatal("TracesSampleRate mismatch")
		}
	})

	t.Run("RedisConfig fields", func(t *testing.T) {
		cfg := RedisConfig{
			Addr:       "localhost:6379",
			Password:   "secret",
			DB:         0,
			MaxRetries: 3,
			PoolSize:   10,
		}
		if cfg.PoolSize != 10 {
			t.Fatal("PoolSize mismatch")
		}
	})

	t.Run("EmailConfig fields", func(t *testing.T) {
		cfg := EmailConfig{
			SMTPHost:     "smtp.example.com",
			SMTPPort:     587,
			SMTPUsername: "user",
			SMTPPassword: "pass",
			FromEmail:    "noreply@example.com",
			BackendURL:   "https://api.example.com",
		}
		if cfg.SMTPPort != 587 {
			t.Fatal("SMTPPort mismatch")
		}
	})

	t.Run("PusherConfig fields", func(t *testing.T) {
		cfg := PusherConfig{
			AppID:   "123",
			Key:     "key",
			Secret:  "secret",
			Cluster: "us2",
		}
		if cfg.Cluster != "us2" {
			t.Fatal("Cluster mismatch")
		}
	})

	t.Run("TikTokConfig fields", func(t *testing.T) {
		cfg := TikTokConfig{
			ClientKey:    "key",
			ClientSecret: "secret",
		}
		if cfg.ClientKey != "key" {
			t.Fatal("ClientKey mismatch")
		}
	})

	t.Run("S3Config fields", func(t *testing.T) {
		cfg := S3Config{
			Region:          "us-east-1",
			Bucket:          "my-bucket",
			AccessKeyID:     "access",
			SecretAccessKey: "secret",
			Endpoint:        "https://s3.amazonaws.com",
			CDNBaseURL:      "https://cdn.example.com",
		}
		if cfg.Region != "us-east-1" {
			t.Fatal("Region mismatch")
		}
	})

	t.Run("FacebookConfig fields", func(t *testing.T) {
		cfg := FacebookConfig{
			AppID:                 "123",
			AppSecret:             "secret",
			AppToken:              "token",
			PerTokenRPS:           5.0,
			PerTokenBurst:         10,
			GlobalRPS:             20.0,
			GlobalBurst:           40,
			PerAccountConcurrency: 2.0,
		}
		if cfg.PerTokenRPS != 5.0 {
			t.Fatal("PerTokenRPS mismatch")
		}
	})
}
