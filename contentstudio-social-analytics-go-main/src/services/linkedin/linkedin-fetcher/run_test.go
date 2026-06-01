package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ================== FetcherConfig Tests ==================

func TestDefaultFetcherConfig(t *testing.T) {
	cfg := DefaultFetcherConfig()

	if cfg.MaxPageWorkers != maxPageWorkers {
		t.Errorf("MaxPageWorkers = %d, want %d", cfg.MaxPageWorkers, maxPageWorkers)
	}
	if cfg.MaxProfileWorkers != maxProfileWorkers {
		t.Errorf("MaxProfileWorkers = %d, want %d", cfg.MaxProfileWorkers, maxProfileWorkers)
	}
	if cfg.WorkOrderChanSize != workOrderChanSize {
		t.Errorf("WorkOrderChanSize = %d, want %d", cfg.WorkOrderChanSize, workOrderChanSize)
	}
	if cfg.TimestampUpdateChanSize != timestampUpdateChanSize {
		t.Errorf("TimestampUpdateChanSize = %d, want %d", cfg.TimestampUpdateChanSize, timestampUpdateChanSize)
	}
	if cfg.MaxConcurrentAccounts != maxConcurrentAccounts {
		t.Errorf("MaxConcurrentAccounts = %d, want %d", cfg.MaxConcurrentAccounts, maxConcurrentAccounts)
	}
}

// ================== NewFetcherService Tests ==================

func TestNewFetcherService(t *testing.T) {
	cfg := FetcherConfig{
		MaxPageWorkers:    5,
		MaxProfileWorkers: 5,
		WorkOrderChanSize: 100,
	}

	log := logger.New("error")
	deps := FetcherDependencies{Log: log}

	svc := NewFetcherService(cfg, deps)

	if svc == nil {
		t.Fatal("NewFetcherService returned nil")
	}
	if svc.config.MaxPageWorkers != 5 {
		t.Errorf("MaxPageWorkers = %d, want 5", svc.config.MaxPageWorkers)
	}
}

// ================== FetcherService.Run Tests ==================

func TestFetcherService_Run_ContextCancel(t *testing.T) {
	cfg := FetcherConfig{
		MaxPageWorkers:          1,
		MaxProfileWorkers:       1,
		WorkOrderChanSize:       10,
		TimestampUpdateChanSize: 10,
	}

	log := logger.New("error")

	pageConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}
	profileConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}
	producer := &kafka.MockProducer{}

	deps := FetcherDependencies{
		PageConsumer:    pageConsumer,
		ProfileConsumer: profileConsumer,
		Producer:        producer,
		Log:             log,
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

func TestFetcherService_Run_ProcessesPageBatches(t *testing.T) {
	cfg := FetcherConfig{
		MaxPageWorkers:          1,
		MaxProfileWorkers:       1,
		WorkOrderChanSize:       10,
		TimestampUpdateChanSize: 10,
	}

	log := logger.New("error")

	batch := LinkedInBatchWorkOrder{
		BatchID:  "batch123",
		SyncType: "incremental",
		Accounts: []kafkamodels.LinkedinAccountWorkOrder{
			{ID: "acc1", LinkedinID: "li1"},
			{ID: "acc2", LinkedinID: "li2"},
		},
	}
	batchJSON, _ := json.Marshal(batch)

	pageConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, topics[0], []byte("key"), batchJSON)
			<-ctx.Done()
			return ctx.Err()
		},
	}
	profileConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}
	producer := &kafka.MockProducer{}

	deps := FetcherDependencies{
		PageConsumer:    pageConsumer,
		ProfileConsumer: profileConsumer,
		Producer:        producer,
		Log:             log,
	}

	svc := NewFetcherService(cfg, deps)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- svc.Run(ctx)
	}()

	time.Sleep(200 * time.Millisecond)

	metrics := svc.GetMetrics()
	if metrics.PageBatchesReceived < 1 {
		t.Errorf("PageBatchesReceived = %d, want >= 1", metrics.PageBatchesReceived)
	}
	if metrics.PageAccountsProcessed < 2 {
		t.Errorf("PageAccountsProcessed = %d, want >= 2", metrics.PageAccountsProcessed)
	}

	cancel()
	<-done
}

func TestFetcherService_Run_ProcessesProfileBatches(t *testing.T) {
	cfg := FetcherConfig{
		MaxPageWorkers:          1,
		MaxProfileWorkers:       1,
		WorkOrderChanSize:       10,
		TimestampUpdateChanSize: 10,
	}

	log := logger.New("error")

	batch := LinkedInBatchWorkOrder{
		BatchID:  "batch456",
		SyncType: "full",
		Accounts: []kafkamodels.LinkedinAccountWorkOrder{
			{ID: "acc3", LinkedinID: "li3"},
		},
	}
	batchJSON, _ := json.Marshal(batch)

	pageConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}
	profileConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, topics[0], []byte("key"), batchJSON)
			<-ctx.Done()
			return ctx.Err()
		},
	}
	producer := &kafka.MockProducer{}

	deps := FetcherDependencies{
		PageConsumer:    pageConsumer,
		ProfileConsumer: profileConsumer,
		Producer:        producer,
		Log:             log,
	}

	svc := NewFetcherService(cfg, deps)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- svc.Run(ctx)
	}()

	time.Sleep(200 * time.Millisecond)

	metrics := svc.GetMetrics()
	if metrics.ProfileBatchesReceived < 1 {
		t.Errorf("ProfileBatchesReceived = %d, want >= 1", metrics.ProfileBatchesReceived)
	}
	if metrics.ProfileAccountsProcessed < 1 {
		t.Errorf("ProfileAccountsProcessed = %d, want >= 1", metrics.ProfileAccountsProcessed)
	}

	cancel()
	<-done
}

func TestFetcherService_Run_WithMongoRepo(t *testing.T) {
	cfg := FetcherConfig{
		MaxPageWorkers:          1,
		MaxProfileWorkers:       1,
		WorkOrderChanSize:       10,
		TimestampUpdateChanSize: 10,
	}

	log := logger.New("error")

	pageConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}
	profileConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}
	producer := &kafka.MockProducer{}
	mockRepo := &mongodb.MockUnifiedSocialRepository{}

	deps := FetcherDependencies{
		PageConsumer:    pageConsumer,
		ProfileConsumer: profileConsumer,
		Producer:        producer,
		MongoRepo:       mockRepo,
		Log:             log,
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
		t.Fatal("Run did not exit")
	}
}

// ================== GetMetrics Tests ==================

func TestFetcherService_GetMetrics_Initial(t *testing.T) {
	cfg := FetcherConfig{}
	log := logger.New("error")
	deps := FetcherDependencies{Log: log}

	svc := NewFetcherService(cfg, deps)

	metrics := svc.GetMetrics()
	if metrics.PageBatchesReceived != 0 {
		t.Errorf("PageBatchesReceived = %d, want 0", metrics.PageBatchesReceived)
	}
	if metrics.ProfileBatchesReceived != 0 {
		t.Errorf("ProfileBatchesReceived = %d, want 0", metrics.ProfileBatchesReceived)
	}
	if metrics.PageAccountsProcessed != 0 {
		t.Errorf("PageAccountsProcessed = %d, want 0", metrics.PageAccountsProcessed)
	}
	if metrics.ProfileAccountsProcessed != 0 {
		t.Errorf("ProfileAccountsProcessed = %d, want 0", metrics.ProfileAccountsProcessed)
	}
}

// ================== FetcherMetrics Tests ==================

func TestFetcherMetrics_Struct(t *testing.T) {
	m := FetcherMetrics{
		PageBatchesReceived:      100,
		ProfileBatchesReceived:   50,
		PageAccountsProcessed:    500,
		ProfileAccountsProcessed: 200,
		TimestampUpdates:         300,
	}

	if m.PageBatchesReceived != 100 {
		t.Errorf("PageBatchesReceived = %d, want 100", m.PageBatchesReceived)
	}
	if m.ProfileBatchesReceived != 50 {
		t.Errorf("ProfileBatchesReceived = %d, want 50", m.ProfileBatchesReceived)
	}
	if m.PageAccountsProcessed != 500 {
		t.Errorf("PageAccountsProcessed = %d, want 500", m.PageAccountsProcessed)
	}
	if m.ProfileAccountsProcessed != 200 {
		t.Errorf("ProfileAccountsProcessed = %d, want 200", m.ProfileAccountsProcessed)
	}
}

// ================== Worker Loop Tests ==================

func TestFetcherService_PageWorkerLoop_ContextCancel(t *testing.T) {
	cfg := FetcherConfig{}
	log := logger.New("error")
	deps := FetcherDependencies{Log: log}

	svc := NewFetcherService(cfg, deps)

	jobs := make(chan WorkOrderMessage, 10)
	timestampChan := make(chan TimestampUpdateRequest, 10)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		svc.pageWorkerLoop(ctx, 0, jobs, timestampChan)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("pageWorkerLoop did not exit after context cancel")
	}
}

func TestFetcherService_PageWorkerLoop_ChannelClose(t *testing.T) {
	cfg := FetcherConfig{}
	log := logger.New("error")
	deps := FetcherDependencies{Log: log}

	svc := NewFetcherService(cfg, deps)

	jobs := make(chan WorkOrderMessage, 10)
	timestampChan := make(chan TimestampUpdateRequest, 10)

	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		svc.pageWorkerLoop(ctx, 0, jobs, timestampChan)
		close(done)
	}()

	close(jobs)

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("pageWorkerLoop did not exit after channel close")
	}
}

func TestFetcherService_ProfileWorkerLoop_ContextCancel(t *testing.T) {
	cfg := FetcherConfig{}
	log := logger.New("error")
	deps := FetcherDependencies{Log: log}

	svc := NewFetcherService(cfg, deps)

	jobs := make(chan WorkOrderMessage, 10)
	timestampChan := make(chan TimestampUpdateRequest, 10)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		svc.profileWorkerLoop(ctx, 0, jobs, timestampChan)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("profileWorkerLoop did not exit after context cancel")
	}
}

func TestFetcherService_ProfileWorkerLoop_ChannelClose(t *testing.T) {
	cfg := FetcherConfig{}
	log := logger.New("error")
	deps := FetcherDependencies{Log: log}

	svc := NewFetcherService(cfg, deps)

	jobs := make(chan WorkOrderMessage, 10)
	timestampChan := make(chan TimestampUpdateRequest, 10)

	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		svc.profileWorkerLoop(ctx, 0, jobs, timestampChan)
		close(done)
	}()

	close(jobs)

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("profileWorkerLoop did not exit after channel close")
	}
}

// ================== FetchPostsWithClient Tests ==================

func TestFetchPostsWithClient_Success(t *testing.T) {
	client := &MockLinkedInClient{
		FetchPostsPaginatedFunc: func(ctx context.Context, linkedinID string, entityType string, accessToken string, cutoffTime time.Time) ([]json.RawMessage, error) {
			return []json.RawMessage{
				json.RawMessage(`{"id":"post1"}`),
				json.RawMessage(`{"id":"post2"}`),
			}, nil
		},
	}

	posts, err := FetchPostsWithClient(context.Background(), client, "li123", "organization", "token123", time.Time{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(posts) != 2 {
		t.Errorf("len(posts) = %d, want 2", len(posts))
	}
}

func TestFetchPostsWithClient_Empty(t *testing.T) {
	client := &MockLinkedInClient{
		FetchPostsPaginatedFunc: func(ctx context.Context, linkedinID string, entityType string, accessToken string, cutoffTime time.Time) ([]json.RawMessage, error) {
			return nil, nil
		},
	}

	posts, err := FetchPostsWithClient(context.Background(), client, "li123", "organization", "token123", time.Time{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if posts != nil {
		t.Errorf("expected nil posts, got %v", posts)
	}
}

// ================== FetchInsightsWithClient Tests ==================

func TestFetchInsightsWithClient_Success(t *testing.T) {
	client := &MockLinkedInClient{
		FetchFollowerDataFunc: func(ctx context.Context, linkedinID string, accessToken string) ([]byte, error) {
			return []byte(`{"followers":1000}`), nil
		},
	}

	data, err := FetchInsightsWithClient(context.Background(), client, "li123", "token123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty data")
	}
}

// ================== FetchOrganizationWithClient Tests ==================

func TestFetchOrganizationWithClient_Success(t *testing.T) {
	client := &MockLinkedInClient{
		FetchOrganizationDetailsRawFunc: func(ctx context.Context, linkedinID string, accessToken string) ([]byte, error) {
			return []byte(`{"name":"Test Org"}`), nil
		},
	}

	data, err := FetchOrganizationWithClient(context.Background(), client, "li123", "token123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty data")
	}
}

// ================== FetchPageStatisticsWithClient Tests ==================

func TestFetchPageStatisticsWithClient_Success(t *testing.T) {
	client := &MockLinkedInClient{
		FetchPageStatisticsRawFunc: func(ctx context.Context, linkedinID string, accessToken string, startMs, endMs int64) ([]byte, error) {
			return []byte(`{"pageViews":500}`), nil
		},
	}

	data, err := FetchPageStatisticsWithClient(context.Background(), client, "li123", "token123", 0, 1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty data")
	}
}

// ================== FetchShareStatisticsWithClient Tests ==================

func TestFetchShareStatisticsWithClient_Success(t *testing.T) {
	client := &MockLinkedInClient{
		FetchShareStatisticsRawFunc: func(ctx context.Context, linkedinID string, accessToken string, startMs, endMs int64) ([]byte, error) {
			return []byte(`{"shareCount":100}`), nil
		},
	}

	data, err := FetchShareStatisticsWithClient(context.Background(), client, "li123", "token123", 0, 1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty data")
	}
}

// ================== ProcessPageWorkOrderWithClient Tests ==================

func TestProcessPageWorkOrderWithClient_Success(t *testing.T) {
	client := &MockLinkedInClient{
		FetchPostsPaginatedFunc: func(ctx context.Context, linkedinID string, entityType string, accessToken string, cutoffTime time.Time) ([]json.RawMessage, error) {
			return []json.RawMessage{
				json.RawMessage(`{"id":"post1"}`),
				json.RawMessage(`{"id":"post2"}`),
			}, nil
		},
		FetchFollowerDataFunc: func(ctx context.Context, linkedinID string, accessToken string) ([]byte, error) {
			return []byte(`{"followers":1000}`), nil
		},
		FetchOrganizationDetailsRawFunc: func(ctx context.Context, linkedinID string, accessToken string) ([]byte, error) {
			return []byte(`{"name":"Test Org"}`), nil
		},
	}

	geoResolver := &MockGeoResolver{}
	producer := &kafka.MockProducer{}

	order := LinkedInAccountWorkOrder{
		ID:          "acc123",
		LinkedinID:  "li456",
		WorkspaceID: "ws789",
	}

	result, err := ProcessPageWorkOrderWithClient(context.Background(), client, geoResolver, producer, order, "token123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PostsCount != 2 {
		t.Errorf("PostsCount = %d, want 2", result.PostsCount)
	}
	if !result.HasInsights {
		t.Error("expected HasInsights to be true")
	}
	if !result.HasOrgDetails {
		t.Error("expected HasOrgDetails to be true")
	}
}

func TestProcessPageWorkOrderWithClient_FetchPostsError(t *testing.T) {
	client := &MockLinkedInClient{
		FetchPostsPaginatedFunc: func(ctx context.Context, linkedinID string, entityType string, accessToken string, cutoffTime time.Time) ([]json.RawMessage, error) {
			return nil, context.DeadlineExceeded
		},
	}

	geoResolver := &MockGeoResolver{}
	producer := &kafka.MockProducer{}

	order := LinkedInAccountWorkOrder{
		ID:         "acc123",
		LinkedinID: "li456",
	}

	_, err := ProcessPageWorkOrderWithClient(context.Background(), client, geoResolver, producer, order, "token123")
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded error, got %v", err)
	}
}

func TestProcessPageWorkOrderWithClient_FetchInsightsError(t *testing.T) {
	client := &MockLinkedInClient{
		FetchPostsPaginatedFunc: func(ctx context.Context, linkedinID string, entityType string, accessToken string, cutoffTime time.Time) ([]json.RawMessage, error) {
			return []json.RawMessage{}, nil
		},
		FetchFollowerDataFunc: func(ctx context.Context, linkedinID string, accessToken string) ([]byte, error) {
			return nil, context.Canceled
		},
	}

	geoResolver := &MockGeoResolver{}
	producer := &kafka.MockProducer{}

	order := LinkedInAccountWorkOrder{
		ID:         "acc123",
		LinkedinID: "li456",
	}

	_, err := ProcessPageWorkOrderWithClient(context.Background(), client, geoResolver, producer, order, "token123")
	if err != context.Canceled {
		t.Errorf("expected Canceled error, got %v", err)
	}
}

// ================== ProcessProfileWorkOrderWithClient Tests ==================

func TestProcessProfileWorkOrderWithClient_Success(t *testing.T) {
	client := &MockLinkedInClient{
		FetchMemberCreatorPostAnalyticsRawFunc: func(ctx context.Context, accessToken string, queryType string, startDate, endDate *time.Time) ([]byte, error) {
			return []byte(`{"analytics":"data"}`), nil
		},
		FetchMemberFollowersCountRawFunc: func(ctx context.Context, accessToken string, startDate, endDate *time.Time) ([]byte, error) {
			return []byte(`{"followers":500}`), nil
		},
	}

	producer := &kafka.MockProducer{}

	order := LinkedInAccountWorkOrder{
		ID:          "acc123",
		LinkedinID:  "li456",
		WorkspaceID: "ws789",
	}

	result, err := ProcessProfileWorkOrderWithClient(context.Background(), client, producer, order, "token123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AnalyticsCount != 2 {
		t.Errorf("AnalyticsCount = %d, want 2", result.AnalyticsCount)
	}
	if !result.HasFollowerData {
		t.Error("expected HasFollowerData to be true")
	}
}

func TestProcessProfileWorkOrderWithClient_AnalyticsError(t *testing.T) {
	client := &MockLinkedInClient{
		FetchMemberCreatorPostAnalyticsRawFunc: func(ctx context.Context, accessToken string, queryType string, startDate, endDate *time.Time) ([]byte, error) {
			return nil, context.DeadlineExceeded
		},
	}

	producer := &kafka.MockProducer{}

	order := LinkedInAccountWorkOrder{
		ID:         "acc123",
		LinkedinID: "li456",
	}

	_, err := ProcessProfileWorkOrderWithClient(context.Background(), client, producer, order, "token123")
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded error, got %v", err)
	}
}

func TestProcessProfileWorkOrderWithClient_FollowerError(t *testing.T) {
	client := &MockLinkedInClient{
		FetchMemberCreatorPostAnalyticsRawFunc: func(ctx context.Context, accessToken string, queryType string, startDate, endDate *time.Time) ([]byte, error) {
			return []byte(`{}`), nil
		},
		FetchMemberFollowersCountRawFunc: func(ctx context.Context, accessToken string, startDate, endDate *time.Time) ([]byte, error) {
			return nil, context.Canceled
		},
	}

	producer := &kafka.MockProducer{}

	order := LinkedInAccountWorkOrder{
		ID:         "acc123",
		LinkedinID: "li456",
	}

	_, err := ProcessProfileWorkOrderWithClient(context.Background(), client, producer, order, "token123")
	if err != context.Canceled {
		t.Errorf("expected Canceled error, got %v", err)
	}
}

// ================== BuildRawPagePosts Tests ==================

func TestBuildRawPagePosts(t *testing.T) {
	posts := []json.RawMessage{
		json.RawMessage(`{"id":"post1"}`),
	}
	statsData := []byte(`{"stats":"data"}`)
	imagesData := []byte(`{"images":"data"}`)
	videosData := []byte(`{"videos":"data"}`)
	documentsData := []byte(`{"documents":"data"}`)

	result := BuildRawPagePosts("acc123", "ws456", "li789", posts, statsData, imagesData, videosData, documentsData)

	if result.AccountID != "acc123" {
		t.Errorf("AccountID = %q, want %q", result.AccountID, "acc123")
	}
	if result.WorkspaceID != "ws456" {
		t.Errorf("WorkspaceID = %q, want %q", result.WorkspaceID, "ws456")
	}
	if result.LinkedinID != "li789" {
		t.Errorf("LinkedinID = %q, want %q", result.LinkedinID, "li789")
	}
	if len(result.Posts) != 1 {
		t.Errorf("len(Posts) = %d, want 1", len(result.Posts))
	}
}

// ================== BuildRawPageInsights Tests ==================

func TestBuildRawPageInsights(t *testing.T) {
	followerData := []byte(`{"followers":1000}`)
	pageStatsData := []byte(`{"pageViews":500}`)
	shareStatsData := []byte(`{"shares":100}`)
	orgDetailsData := []byte(`{"name":"Org"}`)

	result := BuildRawPageInsights("acc123", "ws456", "li789", followerData, pageStatsData, shareStatsData, orgDetailsData)

	if result.AccountID != "acc123" {
		t.Errorf("AccountID = %q, want %q", result.AccountID, "acc123")
	}
	if result.WorkspaceID != "ws456" {
		t.Errorf("WorkspaceID = %q, want %q", result.WorkspaceID, "ws456")
	}
	if result.LinkedinID != "li789" {
		t.Errorf("LinkedinID = %q, want %q", result.LinkedinID, "li789")
	}
	if string(result.FollowerData) != `{"followers":1000}` {
		t.Errorf("FollowerData = %q, want %q", string(result.FollowerData), `{"followers":1000}`)
	}
}

// ================== PageFetchResult Tests ==================

func TestPageFetchResult_Struct(t *testing.T) {
	result := PageFetchResult{
		PostsCount:    10,
		HasInsights:   true,
		HasOrgDetails: true,
	}

	if result.PostsCount != 10 {
		t.Errorf("PostsCount = %d, want 10", result.PostsCount)
	}
	if !result.HasInsights {
		t.Error("expected HasInsights to be true")
	}
	if !result.HasOrgDetails {
		t.Error("expected HasOrgDetails to be true")
	}
}

// ================== ProfileFetchResult Tests ==================

func TestProfileFetchResult_Struct(t *testing.T) {
	result := ProfileFetchResult{
		AnalyticsCount:  5,
		HasFollowerData: true,
	}

	if result.AnalyticsCount != 5 {
		t.Errorf("AnalyticsCount = %d, want 5", result.AnalyticsCount)
	}
	if !result.HasFollowerData {
		t.Error("expected HasFollowerData to be true")
	}
}
