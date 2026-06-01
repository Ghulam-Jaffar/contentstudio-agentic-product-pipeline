package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func mustMarshal(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("mustMarshal: %v", err)
	}
	return b
}

// ================== FetcherConfig Tests ==================

func TestDefaultFetcherConfig(t *testing.T) {
	cfg := DefaultFetcherConfig()

	if cfg.MaxMediaWorkers != maxMediaWorkers {
		t.Errorf("MaxMediaWorkers = %d, want %d", cfg.MaxMediaWorkers, maxMediaWorkers)
	}
	if cfg.MaxInsightsWorkers != maxInsightsWorkers {
		t.Errorf("MaxInsightsWorkers = %d, want %d", cfg.MaxInsightsWorkers, maxInsightsWorkers)
	}
	if cfg.MediaQueueSize != mediaQueueSize {
		t.Errorf("MediaQueueSize = %d, want %d", cfg.MediaQueueSize, mediaQueueSize)
	}
	if cfg.InsightsQueueSize != insightsQueueSize {
		t.Errorf("InsightsQueueSize = %d, want %d", cfg.InsightsQueueSize, insightsQueueSize)
	}
	if cfg.MaxConcurrentAccounts != maxConcurrentAccounts {
		t.Errorf("MaxConcurrentAccounts = %d, want %d", cfg.MaxConcurrentAccounts, maxConcurrentAccounts)
	}
	if cfg.TimestampUpdateChanSize != timestampUpdateChanSize {
		t.Errorf("TimestampUpdateChanSize = %d, want %d", cfg.TimestampUpdateChanSize, timestampUpdateChanSize)
	}
}

// ================== NewFetcherService Tests ==================

func TestNewFetcherService(t *testing.T) {
	cfg := FetcherConfig{
		MaxMediaWorkers:    5,
		MaxInsightsWorkers: 3,
		MediaQueueSize:     100,
		InsightsQueueSize:  50,
	}

	log := logger.New("error")
	deps := FetcherDependencies{Log: log}

	svc := NewFetcherService(cfg, deps)

	if svc == nil {
		t.Fatal("NewFetcherService returned nil")
	}
	if svc.config.MaxMediaWorkers != 5 {
		t.Errorf("MaxMediaWorkers = %d, want 5", svc.config.MaxMediaWorkers)
	}
	if svc.config.MaxInsightsWorkers != 3 {
		t.Errorf("MaxInsightsWorkers = %d, want 3", svc.config.MaxInsightsWorkers)
	}
}

// ================== FetcherService.Run Tests ==================

func TestFetcherService_Run_ContextCancel(t *testing.T) {
	cfg := FetcherConfig{
		MaxMediaWorkers:         1,
		MaxInsightsWorkers:      1,
		MediaQueueSize:          10,
		InsightsQueueSize:       10,
		TimestampUpdateChanSize: 10,
	}

	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}
	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}

	deps := FetcherDependencies{
		Producer: mockProducer,
		Consumer: mockConsumer,
		Log:      log,
	}

	svc := NewFetcherService(cfg, deps)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- svc.Run(ctx)
	}()

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not exit after context cancel")
	}
}

func TestFetcherService_Run_ProcessesAccounts(t *testing.T) {
	cfg := FetcherConfig{
		MaxMediaWorkers:         1,
		MaxInsightsWorkers:      1,
		MediaQueueSize:          10,
		InsightsQueueSize:       10,
		TimestampUpdateChanSize: 10,
		DecryptionKey:           "test_key",
		AppSecret:               "test_secret",
	}

	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}
	mockRepo := &mongodb.MockUnifiedSocialRepository{}

	testAccountID := primitive.NewObjectID()
	testWorkspaceID := primitive.NewObjectID()

	batch := kafkamodels.InstagramBatchWorkOrder{
		BatchID:  "test-batch-1",
		SyncType: "incremental",
		Accounts: []kafkamodels.InstagramAccountWorkOrder{
			{
				ID:          testAccountID.Hex(),
				InstagramID: "ig123",
				WorkspaceID: testWorkspaceID.Hex(),
				AccessToken: "EAAxxxxxxxx",
				SyncType:    "incremental",
			},
		},
	}

	mockConsumer := &kafka.MockConsumerWithMessages{
		Messages: []kafka.MockMessage{
			{
				Topic: "work-order-instagram",
				Key:   []byte("test-batch-1"),
				Value: mustMarshal(t, batch),
			},
		},
	}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": instagramID}, nil
		},
		FetchMediaSinceFunc: func(ctx context.Context, instagramID, accessToken string, since time.Time) ([]kafkamodels.RawInstagramMedia, error) {
			return nil, nil
		},
		FetchMediaFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return nil, nil
		},
		FetchStoriesFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return nil, nil
		},
		FetchAccountDemographicsFunc: func(ctx context.Context, instagramID, accessToken string) (*kafkamodels.RawInstagramDemographics, error) {
			return nil, nil
		},
		FetchInsightsDailyFunc: func(ctx context.Context, instagramID, accessToken string, days, concurrency int) ([]social.DailyInsight, error) {
			return nil, nil
		},
	}

	deps := FetcherDependencies{
		Producer: mockProducer,
		Consumer: mockConsumer,
		MongoRepo: mockRepo,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	}

	svc := NewFetcherService(cfg, deps)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := svc.Run(ctx)
	if err != nil {
		t.Errorf("Run returned error: %v", err)
	}

	metrics := svc.GetMetrics()
	if metrics.AccountsProcessed < 1 {
		t.Errorf("AccountsProcessed = %d, want >= 1", metrics.AccountsProcessed)
	}
}

func TestFetcherService_Run_NoAccounts(t *testing.T) {
	cfg := FetcherConfig{
		MaxMediaWorkers:         1,
		MaxInsightsWorkers:      1,
		MediaQueueSize:          10,
		InsightsQueueSize:       10,
		TimestampUpdateChanSize: 10,
	}

	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}
	mockConsumer := &kafka.MockConsumer{} // ConsumeFunc is nil — returns immediately

	deps := FetcherDependencies{
		Producer: mockProducer,
		Consumer: mockConsumer,
		Log:      log,
	}

	svc := NewFetcherService(cfg, deps)

	err := svc.Run(context.Background())
	if err != nil {
		t.Errorf("Run returned error: %v", err)
	}

	metrics := svc.GetMetrics()
	if metrics.AccountsProcessed != 0 {
		t.Errorf("AccountsProcessed = %d, want 0", metrics.AccountsProcessed)
	}
}

func TestFetcherService_Run_WithMongoRepo(t *testing.T) {
	cfg := FetcherConfig{
		MaxMediaWorkers:         1,
		MaxInsightsWorkers:      1,
		MediaQueueSize:          10,
		InsightsQueueSize:       10,
		TimestampUpdateChanSize: 10,
	}

	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}
	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	mockConsumer := &kafka.MockConsumer{} // ConsumeFunc is nil — returns immediately

	deps := FetcherDependencies{
		Producer:  mockProducer,
		Consumer:  mockConsumer,
		MongoRepo: mockRepo,
		Log:       log,
	}

	svc := NewFetcherService(cfg, deps)

	err := svc.Run(context.Background())
	if err != nil {
		t.Errorf("Run returned error: %v", err)
	}
}

// ================== GetMetrics Tests ==================

func TestFetcherService_GetMetrics_Initial(t *testing.T) {
	cfg := FetcherConfig{}
	log := logger.New("error")
	deps := FetcherDependencies{Log: log}

	svc := NewFetcherService(cfg, deps)

	metrics := svc.GetMetrics()
	if metrics.BatchesReceived != 0 {
		t.Errorf("BatchesReceived = %d, want 0", metrics.BatchesReceived)
	}
	if metrics.AccountsProcessed != 0 {
		t.Errorf("AccountsProcessed = %d, want 0", metrics.AccountsProcessed)
	}
	if metrics.MediaJobsCreated != 0 {
		t.Errorf("MediaJobsCreated = %d, want 0", metrics.MediaJobsCreated)
	}
	if metrics.InsightsJobsCreated != 0 {
		t.Errorf("InsightsJobsCreated = %d, want 0", metrics.InsightsJobsCreated)
	}
}

// ================== FetcherMetrics Tests ==================

func TestFetcherMetrics_Struct(t *testing.T) {
	m := FetcherMetrics{
		BatchesReceived:     100,
		AccountsProcessed:   500,
		MediaJobsCreated:    400,
		InsightsJobsCreated: 450,
		TimestampUpdates:    300,
	}

	if m.BatchesReceived != 100 {
		t.Errorf("BatchesReceived = %d, want 100", m.BatchesReceived)
	}
	if m.AccountsProcessed != 500 {
		t.Errorf("AccountsProcessed = %d, want 500", m.AccountsProcessed)
	}
	if m.MediaJobsCreated != 400 {
		t.Errorf("MediaJobsCreated = %d, want 400", m.MediaJobsCreated)
	}
	if m.InsightsJobsCreated != 450 {
		t.Errorf("InsightsJobsCreated = %d, want 450", m.InsightsJobsCreated)
	}
	if m.TimestampUpdates != 300 {
		t.Errorf("TimestampUpdates = %d, want 300", m.TimestampUpdates)
	}
}

// ================== Worker Loop Tests ==================

func TestFetcherService_MediaWorkerLoop_ContextCancel(t *testing.T) {
	cfg := FetcherConfig{}
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}

	deps := FetcherDependencies{
		Producer: mockProducer,
		Log:      log,
	}

	svc := NewFetcherService(cfg, deps)

	jobs := make(chan MediaJob, 10)
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		svc.mediaWorkerLoop(ctx, 0, jobs, timestampUpdateChan)
		close(done)
	}()

	// Workers now exit via channel close (not context cancel)
	close(jobs)

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("mediaWorkerLoop did not exit after channel close")
	}
}

func TestFetcherService_MediaWorkerLoop_ChannelClose(t *testing.T) {
	cfg := FetcherConfig{}
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}

	deps := FetcherDependencies{
		Producer: mockProducer,
		Log:      log,
	}

	svc := NewFetcherService(cfg, deps)

	jobs := make(chan MediaJob, 10)
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		svc.mediaWorkerLoop(ctx, 0, jobs, timestampUpdateChan)
		close(done)
	}()

	close(jobs)

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("mediaWorkerLoop did not exit after channel close")
	}
}

func TestFetcherService_InsightsWorkerLoop_ContextCancel(t *testing.T) {
	cfg := FetcherConfig{}
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}

	deps := FetcherDependencies{
		Producer: mockProducer,
		Log:      log,
	}

	svc := NewFetcherService(cfg, deps)

	jobs := make(chan InsightsJob, 10)

	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		svc.insightsWorkerLoop(ctx, 0, jobs)
		close(done)
	}()

	// Workers now exit via channel close (not context cancel)
	close(jobs)

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("insightsWorkerLoop did not exit after channel close")
	}
}

func TestFetcherService_InsightsWorkerLoop_ChannelClose(t *testing.T) {
	cfg := FetcherConfig{}
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}

	deps := FetcherDependencies{
		Producer: mockProducer,
		Log:      log,
	}

	svc := NewFetcherService(cfg, deps)

	jobs := make(chan InsightsJob, 10)

	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		svc.insightsWorkerLoop(ctx, 0, jobs)
		close(done)
	}()

	close(jobs)

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("insightsWorkerLoop did not exit after channel close")
	}
}

// ================== processMediaJobWithClient Tests ==================

func TestFetcherService_ProcessMediaJobWithClient_Success(t *testing.T) {
	cfg := FetcherConfig{}
	log := logger.New("error")

	var produceCalled int
	mockProducer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			produceCalled++
			return nil
		},
	}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": instagramID, "followers_count": 1000}, nil
		},
		FetchMediaFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return []kafkamodels.RawInstagramMedia{
				{ID: "media1", MediaType: "IMAGE"},
				{ID: "media2", MediaType: "VIDEO"},
			}, nil
		},
		FetchStoriesFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return []kafkamodels.RawInstagramMedia{
				{ID: "story1", MediaType: "STORY"},
			}, nil
		},
		FetchMediaInsightsFunc: func(ctx context.Context, mediaID, accessToken, mediaType, mediaProductType string) (*kafkamodels.RawInstagramMediaInsights, error) {
			return &kafkamodels.RawInstagramMediaInsights{}, nil
		},
	}

	deps := FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	}

	svc := NewFetcherService(cfg, deps)

	job := MediaJob{
		Order: ResolvedOrder{
			InstagramID:          "ig123",
			AccessTokenPlaintext: "token",
		},
		SyncType: "full_sync",
	}

	timestampChan := make(chan TimestampUpdateRequest, 10)
	ctx := context.Background()

	svc.processMediaJobWithClient(ctx, log, job, timestampChan)

	if produceCalled != 1 {
		t.Errorf("Produce called %d times, want 1", produceCalled)
	}

	if len(timestampChan) != 1 {
		t.Errorf("timestampChan has %d messages, want 1", len(timestampChan))
	}
}

func TestFetcherService_ProcessMediaJobWithClient_UserInfoError(t *testing.T) {
	cfg := FetcherConfig{}
	log := logger.New("error")

	var produceCalled int
	mockProducer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			produceCalled++
			return nil
		},
	}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return nil, fmt.Errorf("user info error")
		},
	}

	deps := FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	}

	svc := NewFetcherService(cfg, deps)

	job := MediaJob{
		Order: ResolvedOrder{
			InstagramID:          "ig123",
			AccessTokenPlaintext: "token",
		},
	}

	timestampChan := make(chan TimestampUpdateRequest, 10)
	ctx := context.Background()

	svc.processMediaJobWithClient(ctx, log, job, timestampChan)

	if produceCalled != 0 {
		t.Errorf("Produce should not be called on user info error, called %d times", produceCalled)
	}
}

func TestFetcherService_ProcessMediaJobWithClient_MediaSince(t *testing.T) {
	cfg := FetcherConfig{}
	log := logger.New("error")

	var fetchMediaSinceCalled bool
	mockProducer := &kafka.MockProducer{}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": instagramID}, nil
		},
		FetchMediaSinceFunc: func(ctx context.Context, instagramID, accessToken string, since time.Time) ([]kafkamodels.RawInstagramMedia, error) {
			fetchMediaSinceCalled = true
			return []kafkamodels.RawInstagramMedia{}, nil
		},
		FetchStoriesFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return nil, nil
		},
	}

	deps := FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	}

	svc := NewFetcherService(cfg, deps)

	since := time.Now().Add(-24 * time.Hour)
	job := MediaJob{
		Order: ResolvedOrder{
			InstagramID:          "ig123",
			AccessTokenPlaintext: "token",
		},
		SyncType: "incremental",
		Since:    &since,
	}

	timestampChan := make(chan TimestampUpdateRequest, 10)
	ctx := context.Background()

	svc.processMediaJobWithClient(ctx, log, job, timestampChan)

	if !fetchMediaSinceCalled {
		t.Error("FetchMediaSince should be called for incremental sync")
	}
}

// ================== processInsightsJobWithClient Tests ==================

func TestFetcherService_ProcessInsightsJobWithClient_Success(t *testing.T) {
	cfg := FetcherConfig{}
	log := logger.New("error")

	var produceCalled int
	mockProducer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			produceCalled++
			return nil
		},
	}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": instagramID}, nil
		},
		FetchAccountDemographicsFunc: func(ctx context.Context, instagramID, accessToken string) (*kafkamodels.RawInstagramDemographics, error) {
			return &kafkamodels.RawInstagramDemographics{}, nil
		},
		FetchInsightsDailyFunc: func(ctx context.Context, instagramID, accessToken string, days, concurrency int) ([]social.DailyInsight, error) {
			return []social.DailyInsight{{Date: time.Now()}}, nil
		},
	}

	deps := FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	}

	svc := NewFetcherService(cfg, deps)

	job := InsightsJob{
		Order: ResolvedOrder{
			InstagramID:          "ig123",
			AccessTokenPlaintext: "token",
		},
		SyncType: "incremental",
		Since:    time.Now().Add(-14 * 24 * time.Hour),
		Until:    time.Now(),
	}

	ctx := context.Background()

	svc.processInsightsJobWithClient(ctx, log, job)

	if produceCalled != 1 {
		t.Errorf("Produce called %d times, want 1", produceCalled)
	}
}

func TestFetcherService_ProcessInsightsJobWithClient_NoUserInfo(t *testing.T) {
	cfg := FetcherConfig{}
	log := logger.New("error")

	var produceCalled int
	mockProducer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			produceCalled++
			return nil
		},
	}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return nil, nil // No user info
		},
		FetchAccountDemographicsFunc: func(ctx context.Context, instagramID, accessToken string) (*kafkamodels.RawInstagramDemographics, error) {
			return &kafkamodels.RawInstagramDemographics{}, nil
		},
		FetchInsightsDailyFunc: func(ctx context.Context, instagramID, accessToken string, days, concurrency int) ([]social.DailyInsight, error) {
			return []social.DailyInsight{{Date: time.Now()}}, nil
		},
	}

	deps := FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	}

	svc := NewFetcherService(cfg, deps)

	job := InsightsJob{
		Order: ResolvedOrder{
			InstagramID:          "ig123",
			AccessTokenPlaintext: "token",
		},
		Since: time.Now().Add(-14 * 24 * time.Hour),
		Until: time.Now(),
	}

	ctx := context.Background()

	svc.processInsightsJobWithClient(ctx, log, job)

	if produceCalled != 0 {
		t.Errorf("Produce should not be called without user info, called %d times", produceCalled)
	}
}

func TestFetcherService_ProcessInsightsJobWithClient_NoData(t *testing.T) {
	cfg := FetcherConfig{}
	log := logger.New("error")

	var produceCalled int
	mockProducer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			produceCalled++
			return nil
		},
	}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": instagramID}, nil
		},
		FetchAccountDemographicsFunc: func(ctx context.Context, instagramID, accessToken string) (*kafkamodels.RawInstagramDemographics, error) {
			return nil, nil // No demographics
		},
		FetchInsightsDailyFunc: func(ctx context.Context, instagramID, accessToken string, days, concurrency int) ([]social.DailyInsight, error) {
			return []social.DailyInsight{}, nil // No insights
		},
	}

	deps := FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	}

	svc := NewFetcherService(cfg, deps)

	job := InsightsJob{
		Order: ResolvedOrder{
			InstagramID:          "ig123",
			AccessTokenPlaintext: "token",
		},
		Since: time.Now().Add(-14 * 24 * time.Hour),
		Until: time.Now(),
	}

	ctx := context.Background()

	svc.processInsightsJobWithClient(ctx, log, job)

	if produceCalled != 0 {
		t.Errorf("Produce should not be called with no data, called %d times", produceCalled)
	}
}

func TestFetcherService_ProcessInsightsJobWithClient_DaysCalculation(t *testing.T) {
	cfg := FetcherConfig{}
	log := logger.New("error")

	var capturedDays int
	mockProducer := &kafka.MockProducer{}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": instagramID}, nil
		},
		FetchAccountDemographicsFunc: func(ctx context.Context, instagramID, accessToken string) (*kafkamodels.RawInstagramDemographics, error) {
			return nil, nil
		},
		FetchInsightsDailyFunc: func(ctx context.Context, instagramID, accessToken string, days, concurrency int) ([]social.DailyInsight, error) {
			capturedDays = days
			return []social.DailyInsight{{Date: time.Now()}}, nil
		},
	}

	deps := FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	}

	svc := NewFetcherService(cfg, deps)

	// Test with 100 days (should be capped to 89)
	job := InsightsJob{
		Order: ResolvedOrder{
			InstagramID:          "ig123",
			AccessTokenPlaintext: "token",
		},
		Since: time.Now().Add(-100 * 24 * time.Hour),
		Until: time.Now(),
	}

	ctx := context.Background()
	svc.processInsightsJobWithClient(ctx, log, job)

	if capturedDays != 89 {
		t.Errorf("days should be capped at 89, got %d", capturedDays)
	}
}

// ================== DefaultInstagramClientFactory Tests ==================

func TestDefaultInstagramClientFactory(t *testing.T) {
	client := DefaultInstagramClientFactory("secret", false)
	if client == nil {
		t.Fatal("DefaultInstagramClientFactory returned nil")
	}
}

func TestDefaultInstagramClientFactory_ConnectedViaInstagram(t *testing.T) {
	client := DefaultInstagramClientFactory("secret", true)
	if client == nil {
		t.Fatal("DefaultInstagramClientFactory returned nil")
	}
}

// ================== NewFetcherService_DefaultClientFactory Tests ==================

func TestNewFetcherService_DefaultClientFactory(t *testing.T) {
	log := logger.New("error")
	svc := NewFetcherService(FetcherConfig{}, FetcherDependencies{Log: log})
	if svc.deps.ClientFactory == nil {
		t.Fatal("ClientFactory should default to DefaultInstagramClientFactory")
	}
}

// ================== processMediaJobWithClient Error Path Tests ==================

func TestFetcherService_ProcessMediaJobWithClient_AuthErrorOnUserInfo(t *testing.T) {
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return nil, fmt.Errorf("OAuthException (#190) token expired")
		},
	}

	svc := NewFetcherService(FetcherConfig{}, FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	})

	job := MediaJob{
		Order: ResolvedOrder{InstagramID: "ig123", AccessTokenPlaintext: "token"},
	}

	timestampChan := make(chan TimestampUpdateRequest, 10)
	svc.processMediaJobWithClient(context.Background(), log, job, timestampChan)

	if len(timestampChan) != 0 {
		t.Error("should not send timestamp update on auth error")
	}
}

func TestFetcherService_ProcessMediaJobWithClient_MediaFetchError(t *testing.T) {
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": instagramID}, nil
		},
		FetchMediaFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return nil, fmt.Errorf("API rate limit")
		},
	}

	svc := NewFetcherService(FetcherConfig{}, FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	})

	job := MediaJob{
		Order:    ResolvedOrder{InstagramID: "ig123", AccessTokenPlaintext: "token"},
		SyncType: "full_sync",
	}

	timestampChan := make(chan TimestampUpdateRequest, 10)
	svc.processMediaJobWithClient(context.Background(), log, job, timestampChan)

	// Should still send timestamp on media error (continues processing)
}

func TestFetcherService_ProcessMediaJobWithClient_MediaFetchExpectedError(t *testing.T) {
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": instagramID}, nil
		},
		FetchMediaFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return nil, fmt.Errorf("OAuthException (#190) token expired")
		},
	}

	svc := NewFetcherService(FetcherConfig{}, FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	})

	job := MediaJob{
		Order:    ResolvedOrder{InstagramID: "ig123", AccessTokenPlaintext: "token"},
		SyncType: "full_sync",
	}

	timestampChan := make(chan TimestampUpdateRequest, 10)
	svc.processMediaJobWithClient(context.Background(), log, job, timestampChan)
}

func TestFetcherService_ProcessMediaJobWithClient_StoriesError(t *testing.T) {
	log := logger.New("error")

	var produceCalled int
	mockProducer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			produceCalled++
			return nil
		},
	}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": instagramID}, nil
		},
		FetchMediaFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return []kafkamodels.RawInstagramMedia{{ID: "m1", MediaType: "IMAGE"}}, nil
		},
		FetchStoriesFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return nil, fmt.Errorf("stories API error")
		},
		FetchMediaInsightsFunc: func(ctx context.Context, mediaID, accessToken, mediaType, mediaProductType string) (*kafkamodels.RawInstagramMediaInsights, error) {
			return nil, nil
		},
	}

	svc := NewFetcherService(FetcherConfig{}, FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	})

	job := MediaJob{
		Order: ResolvedOrder{InstagramID: "ig123", AccessTokenPlaintext: "token"},
	}

	timestampChan := make(chan TimestampUpdateRequest, 10)
	svc.processMediaJobWithClient(context.Background(), log, job, timestampChan)

	// Should still publish media even if stories fail
	if produceCalled != 1 {
		t.Errorf("Produce called %d times, want 1 (media should still publish)", produceCalled)
	}
}

func TestFetcherService_ProcessMediaJobWithClient_StoriesExpectedError(t *testing.T) {
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": instagramID}, nil
		},
		FetchMediaFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return []kafkamodels.RawInstagramMedia{{ID: "m1", MediaType: "IMAGE"}}, nil
		},
		FetchStoriesFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return nil, fmt.Errorf("not enough viewers")
		},
		FetchMediaInsightsFunc: func(ctx context.Context, mediaID, accessToken, mediaType, mediaProductType string) (*kafkamodels.RawInstagramMediaInsights, error) {
			return nil, nil
		},
	}

	svc := NewFetcherService(FetcherConfig{}, FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	})

	job := MediaJob{
		Order: ResolvedOrder{InstagramID: "ig123", AccessTokenPlaintext: "token"},
	}

	timestampChan := make(chan TimestampUpdateRequest, 10)
	svc.processMediaJobWithClient(context.Background(), log, job, timestampChan)
}

func TestFetcherService_ProcessMediaJobWithClient_MediaInsightsAuthError(t *testing.T) {
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": instagramID}, nil
		},
		FetchMediaFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return []kafkamodels.RawInstagramMedia{{ID: "m1", MediaType: "IMAGE"}}, nil
		},
		FetchStoriesFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return nil, nil
		},
		FetchMediaInsightsFunc: func(ctx context.Context, mediaID, accessToken, mediaType, mediaProductType string) (*kafkamodels.RawInstagramMediaInsights, error) {
			return nil, fmt.Errorf("OAuthException (#190) token expired")
		},
	}

	svc := NewFetcherService(FetcherConfig{}, FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	})

	job := MediaJob{
		Order: ResolvedOrder{InstagramID: "ig123", AccessTokenPlaintext: "token"},
	}

	timestampChan := make(chan TimestampUpdateRequest, 10)
	svc.processMediaJobWithClient(context.Background(), log, job, timestampChan)
}

func TestFetcherService_ProcessMediaJobWithClient_MediaInsightsExpectedError(t *testing.T) {
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": instagramID}, nil
		},
		FetchMediaFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return []kafkamodels.RawInstagramMedia{{ID: "m1", MediaType: "IMAGE"}}, nil
		},
		FetchStoriesFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return nil, nil
		},
		FetchMediaInsightsFunc: func(ctx context.Context, mediaID, accessToken, mediaType, mediaProductType string) (*kafkamodels.RawInstagramMediaInsights, error) {
			return nil, fmt.Errorf("not enough viewers")
		},
	}

	svc := NewFetcherService(FetcherConfig{}, FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	})

	job := MediaJob{
		Order: ResolvedOrder{InstagramID: "ig123", AccessTokenPlaintext: "token"},
	}

	timestampChan := make(chan TimestampUpdateRequest, 10)
	svc.processMediaJobWithClient(context.Background(), log, job, timestampChan)
}

func TestFetcherService_ProcessMediaJobWithClient_MediaInsightsUnexpectedError(t *testing.T) {
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": instagramID}, nil
		},
		FetchMediaFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return []kafkamodels.RawInstagramMedia{{ID: "m1", MediaType: "IMAGE"}}, nil
		},
		FetchStoriesFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return nil, nil
		},
		FetchMediaInsightsFunc: func(ctx context.Context, mediaID, accessToken, mediaType, mediaProductType string) (*kafkamodels.RawInstagramMediaInsights, error) {
			return nil, fmt.Errorf("unexpected server error")
		},
	}

	svc := NewFetcherService(FetcherConfig{}, FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	})

	job := MediaJob{
		Order: ResolvedOrder{InstagramID: "ig123", AccessTokenPlaintext: "token"},
	}

	timestampChan := make(chan TimestampUpdateRequest, 10)
	svc.processMediaJobWithClient(context.Background(), log, job, timestampChan)
}

func TestFetcherService_ProcessMediaJobWithClient_StoryInsightsAuthError(t *testing.T) {
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": instagramID}, nil
		},
		FetchMediaFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return nil, nil
		},
		FetchStoriesFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return []kafkamodels.RawInstagramMedia{{ID: "s1", MediaType: "STORY"}}, nil
		},
		FetchMediaInsightsFunc: func(ctx context.Context, mediaID, accessToken, mediaType, mediaProductType string) (*kafkamodels.RawInstagramMediaInsights, error) {
			return nil, fmt.Errorf("OAuthException (#190) token expired")
		},
	}

	svc := NewFetcherService(FetcherConfig{}, FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	})

	job := MediaJob{
		Order: ResolvedOrder{InstagramID: "ig123", AccessTokenPlaintext: "token"},
	}

	timestampChan := make(chan TimestampUpdateRequest, 10)
	svc.processMediaJobWithClient(context.Background(), log, job, timestampChan)
}

func TestFetcherService_ProcessMediaJobWithClient_StoryInsightsExpectedError(t *testing.T) {
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": instagramID}, nil
		},
		FetchMediaFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return nil, nil
		},
		FetchStoriesFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return []kafkamodels.RawInstagramMedia{{ID: "s1", MediaType: "STORY"}}, nil
		},
		FetchMediaInsightsFunc: func(ctx context.Context, mediaID, accessToken, mediaType, mediaProductType string) (*kafkamodels.RawInstagramMediaInsights, error) {
			return nil, fmt.Errorf("not enough viewers")
		},
	}

	svc := NewFetcherService(FetcherConfig{}, FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	})

	job := MediaJob{
		Order: ResolvedOrder{InstagramID: "ig123", AccessTokenPlaintext: "token"},
	}

	timestampChan := make(chan TimestampUpdateRequest, 10)
	svc.processMediaJobWithClient(context.Background(), log, job, timestampChan)
}

func TestFetcherService_ProcessMediaJobWithClient_StoryInsightsUnexpectedError(t *testing.T) {
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": instagramID}, nil
		},
		FetchMediaFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return nil, nil
		},
		FetchStoriesFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return []kafkamodels.RawInstagramMedia{{ID: "s1", MediaType: "STORY"}}, nil
		},
		FetchMediaInsightsFunc: func(ctx context.Context, mediaID, accessToken, mediaType, mediaProductType string) (*kafkamodels.RawInstagramMediaInsights, error) {
			return nil, fmt.Errorf("unexpected server error")
		},
	}

	svc := NewFetcherService(FetcherConfig{}, FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	})

	job := MediaJob{
		Order: ResolvedOrder{InstagramID: "ig123", AccessTokenPlaintext: "token"},
	}

	timestampChan := make(chan TimestampUpdateRequest, 10)
	svc.processMediaJobWithClient(context.Background(), log, job, timestampChan)
}

func TestFetcherService_ProcessMediaJobWithClient_ConnectedViaInstagram(t *testing.T) {
	log := logger.New("error")

	var capturedConnectedViaIG bool
	mockProducer := &kafka.MockProducer{}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": instagramID}, nil
		},
		FetchMediaFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return nil, nil
		},
		FetchStoriesFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return nil, nil
		},
	}

	svc := NewFetcherService(FetcherConfig{}, FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			capturedConnectedViaIG = connectedViaInstagram
			return mockClient
		},
		Log: log,
	})

	job := MediaJob{
		Order: ResolvedOrder{
			InstagramID:           "ig123",
			AccessTokenPlaintext:  "token",
			ConnectedViaInstagram: true,
		},
	}

	timestampChan := make(chan TimestampUpdateRequest, 10)
	svc.processMediaJobWithClient(context.Background(), log, job, timestampChan)

	if !capturedConnectedViaIG {
		t.Error("ClientFactory should receive connectedViaInstagram=true")
	}
}

// ================== processInsightsJobWithClient Error Path Tests ==================

func TestFetcherService_ProcessInsightsJobWithClient_AuthErrorOnUserInfo(t *testing.T) {
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return nil, fmt.Errorf("OAuthException (#190) token expired")
		},
		FetchAccountDemographicsFunc: func(ctx context.Context, instagramID, accessToken string) (*kafkamodels.RawInstagramDemographics, error) {
			return nil, nil
		},
		FetchInsightsDailyFunc: func(ctx context.Context, instagramID, accessToken string, days, concurrency int) ([]social.DailyInsight, error) {
			return nil, nil
		},
	}

	svc := NewFetcherService(FetcherConfig{}, FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	})

	job := InsightsJob{
		Order:    ResolvedOrder{InstagramID: "ig123", AccessTokenPlaintext: "token"},
		Since:    time.Now().Add(-14 * 24 * time.Hour),
		Until:    time.Now(),
		SyncType: "incremental",
	}

	svc.processInsightsJobWithClient(context.Background(), log, job)
	// Auth error on user info is logged but does not abort (continues to fetch demographics)
}

func TestFetcherService_ProcessInsightsJobWithClient_UnexpectedUserInfoError(t *testing.T) {
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return nil, fmt.Errorf("network timeout")
		},
		FetchAccountDemographicsFunc: func(ctx context.Context, instagramID, accessToken string) (*kafkamodels.RawInstagramDemographics, error) {
			return nil, nil
		},
		FetchInsightsDailyFunc: func(ctx context.Context, instagramID, accessToken string, days, concurrency int) ([]social.DailyInsight, error) {
			return nil, nil
		},
	}

	svc := NewFetcherService(FetcherConfig{}, FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	})

	job := InsightsJob{
		Order:    ResolvedOrder{InstagramID: "ig123", AccessTokenPlaintext: "token"},
		Since:    time.Now().Add(-14 * 24 * time.Hour),
		Until:    time.Now(),
		SyncType: "incremental",
	}

	svc.processInsightsJobWithClient(context.Background(), log, job)
}

func TestFetcherService_ProcessInsightsJobWithClient_DemographicsError(t *testing.T) {
	log := logger.New("error")

	var produceCalled int
	mockProducer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			produceCalled++
			return nil
		},
	}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": instagramID}, nil
		},
		FetchAccountDemographicsFunc: func(ctx context.Context, instagramID, accessToken string) (*kafkamodels.RawInstagramDemographics, error) {
			return nil, fmt.Errorf("demographics API error")
		},
		FetchInsightsDailyFunc: func(ctx context.Context, instagramID, accessToken string, days, concurrency int) ([]social.DailyInsight, error) {
			return []social.DailyInsight{{Date: time.Now()}}, nil
		},
	}

	svc := NewFetcherService(FetcherConfig{}, FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	})

	job := InsightsJob{
		Order:    ResolvedOrder{InstagramID: "ig123", AccessTokenPlaintext: "token"},
		Since:    time.Now().Add(-14 * 24 * time.Hour),
		Until:    time.Now(),
		SyncType: "incremental",
	}

	svc.processInsightsJobWithClient(context.Background(), log, job)

	if produceCalled != 1 {
		t.Errorf("Produce called %d times, want 1 (should still publish with daily insights)", produceCalled)
	}
}

func TestFetcherService_ProcessInsightsJobWithClient_DemographicsExpectedError(t *testing.T) {
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": instagramID}, nil
		},
		FetchAccountDemographicsFunc: func(ctx context.Context, instagramID, accessToken string) (*kafkamodels.RawInstagramDemographics, error) {
			return nil, fmt.Errorf("not enough viewers")
		},
		FetchInsightsDailyFunc: func(ctx context.Context, instagramID, accessToken string, days, concurrency int) ([]social.DailyInsight, error) {
			return []social.DailyInsight{{Date: time.Now()}}, nil
		},
	}

	svc := NewFetcherService(FetcherConfig{}, FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	})

	job := InsightsJob{
		Order:    ResolvedOrder{InstagramID: "ig123", AccessTokenPlaintext: "token"},
		Since:    time.Now().Add(-14 * 24 * time.Hour),
		Until:    time.Now(),
		SyncType: "incremental",
	}

	svc.processInsightsJobWithClient(context.Background(), log, job)
}

func TestFetcherService_ProcessInsightsJobWithClient_InsightsDailyError(t *testing.T) {
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": instagramID}, nil
		},
		FetchAccountDemographicsFunc: func(ctx context.Context, instagramID, accessToken string) (*kafkamodels.RawInstagramDemographics, error) {
			return &kafkamodels.RawInstagramDemographics{}, nil
		},
		FetchInsightsDailyFunc: func(ctx context.Context, instagramID, accessToken string, days, concurrency int) ([]social.DailyInsight, error) {
			return nil, fmt.Errorf("insights daily error")
		},
	}

	svc := NewFetcherService(FetcherConfig{}, FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	})

	job := InsightsJob{
		Order:    ResolvedOrder{InstagramID: "ig123", AccessTokenPlaintext: "token"},
		Since:    time.Now().Add(-14 * 24 * time.Hour),
		Until:    time.Now(),
		SyncType: "incremental",
	}

	svc.processInsightsJobWithClient(context.Background(), log, job)
}

func TestFetcherService_ProcessInsightsJobWithClient_InsightsDailyExpectedError(t *testing.T) {
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": instagramID}, nil
		},
		FetchAccountDemographicsFunc: func(ctx context.Context, instagramID, accessToken string) (*kafkamodels.RawInstagramDemographics, error) {
			return &kafkamodels.RawInstagramDemographics{}, nil
		},
		FetchInsightsDailyFunc: func(ctx context.Context, instagramID, accessToken string, days, concurrency int) ([]social.DailyInsight, error) {
			return nil, fmt.Errorf("not enough viewers")
		},
	}

	svc := NewFetcherService(FetcherConfig{}, FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	})

	job := InsightsJob{
		Order:    ResolvedOrder{InstagramID: "ig123", AccessTokenPlaintext: "token"},
		Since:    time.Now().Add(-14 * 24 * time.Hour),
		Until:    time.Now(),
		SyncType: "incremental",
	}

	svc.processInsightsJobWithClient(context.Background(), log, job)
}

func TestFetcherService_ProcessInsightsJobWithClient_ProduceError(t *testing.T) {
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return fmt.Errorf("kafka unavailable")
		},
	}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": instagramID}, nil
		},
		FetchAccountDemographicsFunc: func(ctx context.Context, instagramID, accessToken string) (*kafkamodels.RawInstagramDemographics, error) {
			return &kafkamodels.RawInstagramDemographics{}, nil
		},
		FetchInsightsDailyFunc: func(ctx context.Context, instagramID, accessToken string, days, concurrency int) ([]social.DailyInsight, error) {
			return []social.DailyInsight{{Date: time.Now()}}, nil
		},
	}

	svc := NewFetcherService(FetcherConfig{}, FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	})

	job := InsightsJob{
		Order:    ResolvedOrder{InstagramID: "ig123", AccessTokenPlaintext: "token"},
		Since:    time.Now().Add(-14 * 24 * time.Hour),
		Until:    time.Now(),
		SyncType: "incremental",
	}

	svc.processInsightsJobWithClient(context.Background(), log, job)
	// Should handle error gracefully
}

func TestFetcherService_ProcessInsightsJobWithClient_DaysMinimum(t *testing.T) {
	log := logger.New("error")

	var capturedDays int
	mockProducer := &kafka.MockProducer{}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": instagramID}, nil
		},
		FetchAccountDemographicsFunc: func(ctx context.Context, instagramID, accessToken string) (*kafkamodels.RawInstagramDemographics, error) {
			return nil, nil
		},
		FetchInsightsDailyFunc: func(ctx context.Context, instagramID, accessToken string, days, concurrency int) ([]social.DailyInsight, error) {
			capturedDays = days
			return []social.DailyInsight{{Date: time.Now()}}, nil
		},
	}

	svc := NewFetcherService(FetcherConfig{}, FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	})

	// Since and Until are the same (0 hours apart) → days should be capped to minimum 1
	now := time.Now()
	job := InsightsJob{
		Order: ResolvedOrder{InstagramID: "ig123", AccessTokenPlaintext: "token"},
		Since: now,
		Until: now,
	}

	svc.processInsightsJobWithClient(context.Background(), log, job)

	if capturedDays != 1 {
		t.Errorf("days should be minimum 1, got %d", capturedDays)
	}
}

// ================== Run with no MongoRepo Tests ==================

func TestFetcherService_Run_NoMongoRepo(t *testing.T) {
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}

	svc := NewFetcherService(
		FetcherConfig{
			MaxMediaWorkers:         1,
			MaxInsightsWorkers:      1,
			MediaQueueSize:          10,
			InsightsQueueSize:       10,
			TimestampUpdateChanSize: 10,
		},
		FetcherDependencies{
			Producer: mockProducer,
			Log:      log,
		},
	)

	err := svc.Run(context.Background())
	if err != nil {
		t.Errorf("Run returned error: %v", err)
	}
}

// ================== MediaWorkerLoop with job processing Tests ==================

func TestFetcherService_MediaWorkerLoop_ProcessesJob(t *testing.T) {
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": instagramID}, nil
		},
		FetchMediaFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return nil, nil
		},
		FetchStoriesFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return nil, nil
		},
	}

	svc := NewFetcherService(FetcherConfig{}, FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	})

	jobs := make(chan MediaJob, 10)
	timestampChan := make(chan TimestampUpdateRequest, 10)

	done := make(chan struct{})
	go func() {
		svc.mediaWorkerLoop(context.Background(), 0, jobs, timestampChan)
		close(done)
	}()

	jobs <- MediaJob{
		Order: ResolvedOrder{InstagramID: "ig123", AccessTokenPlaintext: "token"},
	}

	close(jobs)
	<-done

	metrics := svc.GetMetrics()
	if metrics.MediaJobsCreated != 1 {
		t.Errorf("MediaJobsCreated = %d, want 1", metrics.MediaJobsCreated)
	}
}

func TestFetcherService_InsightsWorkerLoop_ProcessesJob(t *testing.T) {
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}

	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return nil, nil // No user info
		},
		FetchAccountDemographicsFunc: func(ctx context.Context, instagramID, accessToken string) (*kafkamodels.RawInstagramDemographics, error) {
			return nil, nil
		},
		FetchInsightsDailyFunc: func(ctx context.Context, instagramID, accessToken string, days, concurrency int) ([]social.DailyInsight, error) {
			return nil, nil
		},
	}

	svc := NewFetcherService(FetcherConfig{}, FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log: log,
	})

	jobs := make(chan InsightsJob, 10)

	done := make(chan struct{})
	go func() {
		svc.insightsWorkerLoop(context.Background(), 0, jobs)
		close(done)
	}()

	jobs <- InsightsJob{
		Order: ResolvedOrder{InstagramID: "ig123", AccessTokenPlaintext: "token"},
		Since: time.Now().Add(-14 * 24 * time.Hour),
		Until: time.Now(),
	}

	close(jobs)
	<-done

	metrics := svc.GetMetrics()
	if metrics.InsightsJobsCreated != 1 {
		t.Errorf("InsightsJobsCreated = %d, want 1", metrics.InsightsJobsCreated)
	}
}

func TestFetcherService_MediaWorkerLoop_SemAcquireBlocks(t *testing.T) {
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}
	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": instagramID}, nil
		},
		FetchMediaFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return nil, nil
		},
		FetchStoriesFunc: func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
			return nil, nil
		},
	}

	svc := NewFetcherService(FetcherConfig{}, FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log:      log,
	})

	jobs := make(chan MediaJob, 10)
	tsChan := make(chan TimestampUpdateRequest, 10)

	// Pre-acquire the per-account semaphore so the worker's Acquire blocks
	sem := semForAccount("sem_test_media_v2", perAccountConcurrency)
	sem.Acquire(context.Background(), 1)

	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		svc.mediaWorkerLoop(ctx, 0, jobs, tsChan)
		close(done)
	}()

	// Send a job for the same account — Acquire will block
	jobs <- MediaJob{
		Order: ResolvedOrder{InstagramID: "sem_test_media_v2", AccessTokenPlaintext: "token"},
	}

	// Release the semaphore and close the channel so the worker can finish
	time.Sleep(20 * time.Millisecond)
	sem.Release(1)
	close(jobs)

	select {
	case <-done:
		// Expected — worker processed the job and exited
	case <-time.After(5 * time.Second):
		t.Fatal("mediaWorkerLoop did not exit after semaphore release and channel close")
	}
}

func TestFetcherService_InsightsWorkerLoop_SemAcquireBlocks(t *testing.T) {
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}
	mockClient := &MockInstagramClient{
		FetchUserInfoFunc: func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": instagramID}, nil
		},
		FetchAccountDemographicsFunc: func(ctx context.Context, instagramID, accessToken string) (*kafkamodels.RawInstagramDemographics, error) {
			return nil, nil
		},
		FetchInsightsDailyFunc: func(ctx context.Context, instagramID, accessToken string, days, concurrency int) ([]social.DailyInsight, error) {
			return nil, nil
		},
	}

	svc := NewFetcherService(FetcherConfig{}, FetcherDependencies{
		Producer: mockProducer,
		ClientFactory: func(appSecret string, connectedViaInstagram bool) InstagramAPI {
			return mockClient
		},
		Log:      log,
	})

	jobs := make(chan InsightsJob, 10)

	// Pre-acquire the per-account semaphore so the worker's Acquire blocks
	sem := semForAccount("sem_test_insights_v2", perAccountConcurrency)
	sem.Acquire(context.Background(), 1)

	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		svc.insightsWorkerLoop(ctx, 0, jobs)
		close(done)
	}()

	// Send a job for the same account — Acquire will block
	jobs <- InsightsJob{
		Order: ResolvedOrder{InstagramID: "sem_test_insights_v2", AccessTokenPlaintext: "token"},
	}

	// Release the semaphore and close the channel so the worker can finish
	time.Sleep(20 * time.Millisecond)
	sem.Release(1)
	close(jobs)

	select {
	case <-done:
		// Expected — worker processed the job and exited
	case <-time.After(5 * time.Second):
		t.Fatal("insightsWorkerLoop did not exit after semaphore release and channel close")
	}

	if svc.GetMetrics().InsightsJobsCreated != 1 {
		t.Errorf("InsightsJobsCreated = %d, want 1", svc.GetMetrics().InsightsJobsCreated)
	}
}
