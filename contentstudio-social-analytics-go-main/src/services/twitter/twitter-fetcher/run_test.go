package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ================== RunService Tests ==================

func TestRunService_BasicFlow(t *testing.T) {
	log := logger.New("error")

	batch := &kafkamodels.TwitterBatchWorkOrder{
		BatchID: "batch_1",
		Accounts: []kafkamodels.TwitterAccountWorkOrder{
			{
				TwitterID:        "twitter_user_1",
				OAuthToken:       "token_1",
				OAuthTokenSecret: "secret_1",
			},
		},
	}
	batchJSON, _ := json.Marshal(batch)

	consumer := NewMockKafkaConsumerWithMessages([]MockMessage{
		{Topic: "work-order-twitter-batch", Key: []byte("batch_1"), Value: batchJSON},
	})
	producer := NewMockKafkaProducer()
	mongoRepo := NewMockUnifiedSocialRepository()
	twitterClient := NewMockTwitterClient()

	deps := &FetcherDependencies{
		Consumer:      consumer,
		Producer:      producer,
		MongoRepo:     mongoRepo,
		TwitterClient: twitterClient,
		Log:           log,
	}

	cfg := FetcherConfig{
		MaxWorkers:               1,
		WorkQueueSize:            10,
		AccountSemaphoreCapacity: 1,
		DecryptionKey:            "test-key",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := RunService(ctx, deps, cfg)

	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("RunService failed: %v", err)
	}
}

func TestRunService_EmptyConsumer(t *testing.T) {
	log := logger.New("error")

	consumer := NewMockKafkaConsumerWithMessages([]MockMessage{})
	producer := NewMockKafkaProducer()
	mongoRepo := NewMockUnifiedSocialRepository()
	twitterClient := NewMockTwitterClient()

	deps := &FetcherDependencies{
		Consumer:      consumer,
		Producer:      producer,
		MongoRepo:     mongoRepo,
		TwitterClient: twitterClient,
		Log:           log,
	}

	cfg := DefaultFetcherConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := RunService(ctx, deps, cfg)

	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("RunService should complete: %v", err)
	}
}

func TestRunService_MultipleBatches(t *testing.T) {
	log := logger.New("error")

	batch1 := &kafkamodels.TwitterBatchWorkOrder{
		BatchID: "batch_1",
		Accounts: []kafkamodels.TwitterAccountWorkOrder{
			{TwitterID: "user_1", OAuthToken: "token_1", OAuthTokenSecret: "secret_1"},
		},
	}
	batch2 := &kafkamodels.TwitterBatchWorkOrder{
		BatchID: "batch_2",
		Accounts: []kafkamodels.TwitterAccountWorkOrder{
			{TwitterID: "user_2", OAuthToken: "token_2", OAuthTokenSecret: "secret_2"},
		},
	}

	batch1JSON, _ := json.Marshal(batch1)
	batch2JSON, _ := json.Marshal(batch2)

	consumer := NewMockKafkaConsumerWithMessages([]MockMessage{
		{Topic: "work-order-twitter-batch", Key: []byte("batch_1"), Value: batch1JSON},
		{Topic: "work-order-twitter-batch", Key: []byte("batch_2"), Value: batch2JSON},
	})
	producer := NewMockKafkaProducer()
	mongoRepo := NewMockUnifiedSocialRepository()
	twitterClient := NewMockTwitterClient()

	deps := &FetcherDependencies{
		Consumer:      consumer,
		Producer:      producer,
		MongoRepo:     mongoRepo,
		TwitterClient: twitterClient,
		Log:           log,
	}

	cfg := FetcherConfig{
		MaxWorkers:               2,
		WorkQueueSize:            10,
		AccountSemaphoreCapacity: 1,
		DecryptionKey:            "test-key",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	err := RunService(ctx, deps, cfg)

	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("RunService failed: %v", err)
	}
}

// ================== FetcherConfig Tests ==================

func TestDefaultFetcherConfig(t *testing.T) {
	cfg := DefaultFetcherConfig()

	if cfg.MaxWorkers == 0 {
		t.Error("MaxWorkers should not be zero")
	}
	if cfg.WorkQueueSize == 0 {
		t.Error("WorkQueueSize should not be zero")
	}
	if cfg.AccountSemaphoreCapacity == 0 {
		t.Error("AccountSemaphoreCapacity should not be zero")
	}
	if cfg.MaxConcurrentAccounts == 0 {
		t.Error("MaxConcurrentAccounts should not be zero")
	}
}

func TestFetcherConfig_CustomValues(t *testing.T) {
	cfg := FetcherConfig{
		MaxWorkers:               5,
		WorkQueueSize:            50,
		AccountSemaphoreCapacity: 2,
	}

	if cfg.MaxWorkers != 5 {
		t.Errorf("Expected MaxWorkers=5, got %d", cfg.MaxWorkers)
	}
	if cfg.WorkQueueSize != 50 {
		t.Errorf("Expected WorkQueueSize=50, got %d", cfg.WorkQueueSize)
	}
	if cfg.AccountSemaphoreCapacity != 2 {
		t.Errorf("Expected AccountSemaphoreCapacity=2, got %d", cfg.AccountSemaphoreCapacity)
	}
}

func TestHandleWorkOrder_LogsTwitterJobMetadata(t *testing.T) {
	log := logger.New("error")
	producer := NewMockKafkaProducer()
	mongoRepo := NewMockUnifiedSocialRepository()
	var payloads []mongodb.TwitterJobMetadataPayload
	mongoRepo.InsertTwitterJobMetadataFunc = func(ctx context.Context, payload mongodb.TwitterJobMetadataPayload) error {
		payloads = append(payloads, payload)
		return nil
	}

	twitterClient := NewMockTwitterClient()
	twitterClient.FetchUserInfoFunc = func(ctx context.Context, userIDs []string, oauthToken, oauthTokenSecret string) (*social.TwitterUserResponse, error) {
		return &social.TwitterUserResponse{
			Data: []social.TwitterUser{
				{
					ID:       "tw_user_1",
					Name:     "User 1",
					Username: "user1",
				},
			},
		}, nil
	}
	twitterClient.FetchUserTweetsFunc = func(ctx context.Context, userID, oauthToken, oauthTokenSecret string, maxResults int, paginationToken string) (*social.TwitterTweetsResponse, error) {
		return &social.TwitterTweetsResponse{
			Data: []social.TwitterTweet{},
			Meta: &social.TwitterMeta{},
		}, nil
	}

	deps := &FetcherDependencies{
		Consumer:      NewMockKafkaConsumerWithMessages(nil),
		Producer:      producer,
		MongoRepo:     mongoRepo,
		TwitterClient: twitterClient,
		Log:           log,
	}
	cfg := DefaultFetcherConfig()
	cfg.DecryptionKey = "test-key"

	order := kafkamodels.TwitterAccountWorkOrder{
		ID:               "",
		WorkspaceID:      "ws_1",
		TwitterID:        "tw_user_1",
		OAuthToken:       "token",
		OAuthTokenSecret: "secret",
		NTweets:          10,
		AppName:          "demo-app",
		AppID:            "6647030deef711b09d005f02",
		ExecutedBy:       "internal",
		SyncType:         "incremental",
	}
	orderBytes, _ := json.Marshal(order)

	err := handleWorkOrder(context.Background(), []byte(order.TwitterID), orderBytes, deps, cfg, &FetcherMetrics{}, log)
	if err != nil {
		t.Fatalf("handleWorkOrder() error = %v", err)
	}

	if len(payloads) != 1 {
		t.Fatalf("expected 1 job log payload, got %d", len(payloads))
	}
	got := payloads[0]
	if got.PlatformID != "tw_user_1" {
		t.Fatalf("platform_id = %s, want tw_user_1", got.PlatformID)
	}
	// len(tweet response)=0, plus 1 for user info
	if got.CreditsUsed != 1 {
		t.Fatalf("credits_used = %d, want 1", got.CreditsUsed)
	}
}
