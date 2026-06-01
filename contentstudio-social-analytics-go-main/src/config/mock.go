package config

// NewTestConfig creates a Config instance with default test values.
// This can be used in tests to avoid loading from environment.
func NewTestConfig() *Config {
	return &Config{
		Environment: "test",
		LogLevel:    "error",
		Mongo: MongoConfig{
			URI:      "mongodb://localhost:27017",
			Database: "test_db",
			Username: "test",
			Password: "test",
		},
		Kafka: KafkaConfig{
			Brokers:     []string{"localhost:9092"},
			TopicPrefix: "test-",
			SASL: SASLConfig{
				Enabled:   false,
				Username:  "",
				Password:  "",
				Mechanism: "PLAIN",
			},
		},
		Facebook: FacebookConfig{
			AppID:                 "test_app_id",
			AppSecret:             "test_app_secret",
			AppToken:              "test_token",
			PerTokenRPS:           4.0,
			PerTokenBurst:         4,
			GlobalRPS:             12.0,
			GlobalBurst:           12,
			PerAccountConcurrency: 1,
		},
		ClickHouse: ClickHouseConfig{
			Host:                  "localhost",
			Port:                  9000,
			Database:              "test_db",
			Username:              "default",
			Password:              "",
			Secure:                false,
			Compression:           false,
			MaxOpenConns:          5,
			MaxIdleConns:          2,
			MaxExecutionTimeInSec: 60,
		},
		Redis: RedisConfig{
			Addr:       "localhost:6379",
			Password:   "",
			DB:         0,
			MaxRetries: 3,
			PoolSize:   10,
		},
		Email: EmailConfig{
			SMTPHost:     "localhost",
			SMTPPort:     587,
			SMTPUsername: "test",
			SMTPPassword: "test",
			FromEmail:    "test@test.com",
			BackendURL:   "http://localhost:8080",
		},
		Pusher: PusherConfig{
			AppID:   "test_app",
			Key:     "test_key",
			Secret:  "test_secret",
			Cluster: "us2",
		},
		DecryptionKey: "test_decryption_key_32bytes!!!!",
	}
}
