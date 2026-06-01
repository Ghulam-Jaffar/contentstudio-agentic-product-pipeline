package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ================== Service Tests ==================

func TestNewService(t *testing.T) {
	log := logger.New("error")
	consumer := NewMockKafkaConsumerWithMessages(nil)
	producer := NewMockKafkaProducer()
	mongoRepo := NewMockUnifiedSocialRepository()
	pinterestClient := NewMockPinterestClient()

	svc := NewService(pinterestClient, producer, consumer, mongoRepo, log, "test-key")

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.PinterestClient == nil {
		t.Error("PinterestClient should not be nil")
	}
	if svc.Producer == nil {
		t.Error("Producer should not be nil")
	}
	if svc.Consumer == nil {
		t.Error("Consumer should not be nil")
	}
	if svc.MongoRepo == nil {
		t.Error("MongoRepo should not be nil")
	}
	if svc.Logger == nil {
		t.Error("Logger should not be nil")
	}
	if svc.DecryptionKey != "test-key" {
		t.Errorf("DecryptionKey = %q, want %q", svc.DecryptionKey, "test-key")
	}
	if svc.MaxWorkers != maxWorkers {
		t.Errorf("MaxWorkers = %d, want %d", svc.MaxWorkers, maxWorkers)
	}
	if svc.MaxConcurrentAccounts != maxConcurrentAccounts {
		t.Errorf("MaxConcurrentAccounts = %d, want %d", svc.MaxConcurrentAccounts, maxConcurrentAccounts)
	}
	if svc.IdleTimeout != idleTimeout {
		t.Errorf("IdleTimeout = %v, want %v", svc.IdleTimeout, idleTimeout)
	}
}

func TestService_Run_EmptyConsumer(t *testing.T) {
	log := logger.New("error")
	consumer := NewMockKafkaConsumerWithMessages([]MockMessage{})
	producer := NewMockKafkaProducer()
	mongoRepo := NewMockUnifiedSocialRepository()
	pinterestClient := NewMockPinterestClient()

	svc := NewService(pinterestClient, producer, consumer, mongoRepo, log, "test-key")
	svc.IdleTimeout = 100 * time.Millisecond
	svc.IdleCheckPeriod = 20 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := svc.Run(ctx)

	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("Run should complete: %v", err)
	}
}

func TestService_Run_WithBatch(t *testing.T) {
	log := logger.New("error")

	batch := &kafkamodels.PinterestBatchWorkOrder{
		BatchID: "test-batch",
		Accounts: []kafkamodels.PinterestAccountWorkOrder{
			{
				ID:          "acc_1",
				AccountID:   "pinterest_user_1",
				AccessToken: "token_1",
				AccountType: kafkamodels.PinterestAccountTypeProfile,
				SyncType:    kafkamodels.PinterestSyncTypeIncremental,
				WorkspaceID: "workspace_1",
			},
		},
	}
	batchJSON, _ := json.Marshal(batch)

	consumer := NewMockKafkaConsumerWithMessages([]MockMessage{
		{Topic: topicWorkOrderBatch, Key: []byte("batch_1"), Value: batchJSON},
	})

	var producedMessages []struct {
		Topic string
		Key   []byte
		Value []byte
	}
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			producedMessages = append(producedMessages, struct {
				Topic string
				Key   []byte
				Value []byte
			}{Topic: topic, Key: key, Value: value})
			return nil
		},
	}

	mongoRepo := NewMockUnifiedSocialRepository()

	pinterestClient := &social.MockPinterestClient{
		GetUserAccountFunc: func(ctx context.Context, accessToken string) (*social.PinterestUserAccount, error) {
			return &social.PinterestUserAccount{
				ID:            "pinterest_user_1",
				Username:      "testuser",
				FollowerCount: 1000,
				BoardCount:    5,
				PinCount:      100,
			}, nil
		},
		GetUserAccountAnalyticsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.PinterestUserAnalyticsResponse, error) {
			return &social.PinterestUserAnalyticsResponse{}, nil
		},
		GetBoardsFunc: func(ctx context.Context, accessToken string) (*social.PinterestBoardsResponse, error) {
			return &social.PinterestBoardsResponse{
				Items: []social.PinterestBoard{},
			}, nil
		},
	}

	svc := NewService(pinterestClient, producer, consumer, mongoRepo, log, "test-key")
	svc.IdleTimeout = 100 * time.Millisecond
	svc.IdleCheckPeriod = 20 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := svc.Run(ctx)

	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("Run failed: %v", err)
	}

	// Allow some time for message processing
	time.Sleep(100 * time.Millisecond)
}

func TestService_Run_ContextCancelled(t *testing.T) {
	log := logger.New("error")
	consumer := NewMockKafkaConsumerWithMessages([]MockMessage{})
	producer := NewMockKafkaProducer()
	mongoRepo := NewMockUnifiedSocialRepository()
	pinterestClient := NewMockPinterestClient()

	svc := NewService(pinterestClient, producer, consumer, mongoRepo, log, "test-key")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := svc.Run(ctx)

	if err != nil && err != context.Canceled {
		t.Fatalf("Run should handle cancelled context: %v", err)
	}
}

// ================== Constants Tests ==================

func TestConstants(t *testing.T) {
	if maxWorkers <= 0 {
		t.Errorf("maxWorkers should be > 0, got %d", maxWorkers)
	}
	if maxConcurrentAccounts <= 0 {
		t.Errorf("maxConcurrentAccounts should be > 0, got %d", maxConcurrentAccounts)
	}
	if workOrderChanSize <= 0 {
		t.Errorf("workOrderChanSize should be > 0, got %d", workOrderChanSize)
	}
	if timestampChanSize <= 0 {
		t.Errorf("timestampChanSize should be > 0, got %d", timestampChanSize)
	}
	if consumerGroup == "" {
		t.Error("consumerGroup should not be empty")
	}
	if topicWorkOrderBatch != "work-order-pinterest" {
		t.Errorf("topicWorkOrderBatch = %q, want %q", topicWorkOrderBatch, "work-order-pinterest")
	}
	if topicRawUsers != "raw-pinterest-users" {
		t.Errorf("topicRawUsers = %q, want %q", topicRawUsers, "raw-pinterest-users")
	}
	if topicRawBoards != "raw-pinterest-boards" {
		t.Errorf("topicRawBoards = %q, want %q", topicRawBoards, "raw-pinterest-boards")
	}
	if topicRawPins != "raw-pinterest-pins" {
		t.Errorf("topicRawPins = %q, want %q", topicRawPins, "raw-pinterest-pins")
	}
	if topicRawPinInsights != "raw-pinterest-pin-insights" {
		t.Errorf("topicRawPinInsights = %q, want %q", topicRawPinInsights, "raw-pinterest-pin-insights")
	}
	if topicRawUserInsights != "raw-pinterest-user-insights" {
		t.Errorf("topicRawUserInsights = %q, want %q", topicRawUserInsights, "raw-pinterest-user-insights")
	}
}

func TestDateRangeConstants(t *testing.T) {
	if fullSyncDays <= 0 {
		t.Errorf("fullSyncDays should be > 0, got %d", fullSyncDays)
	}
	if incrementalSyncDays <= 0 {
		t.Errorf("incrementalSyncDays should be > 0, got %d", incrementalSyncDays)
	}
	if immediateSyncDays <= 0 {
		t.Errorf("immediateSyncDays should be > 0, got %d", immediateSyncDays)
	}
	if analyticsEndDateOffset <= 0 {
		t.Errorf("analyticsEndDateOffset should be > 0, got %d", analyticsEndDateOffset)
	}
	if fullPageSize <= 0 {
		t.Errorf("fullPageSize should be > 0, got %d", fullPageSize)
	}
	if incrementalPageSize <= 0 {
		t.Errorf("incrementalPageSize should be > 0, got %d", incrementalPageSize)
	}
	if immediatePageSize <= 0 {
		t.Errorf("immediatePageSize should be > 0, got %d", immediatePageSize)
	}
}

// ================== Helper Function Tests ==================

func TestIsUnauthorizedError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"status 401", ErrUnauthorized, true},
		{"contains unauthorized", &testError{msg: "unauthorized access"}, true},
		{"contains status 401", &testError{msg: "failed with status 401"}, true},
		{"other error", &testError{msg: "network timeout"}, false},
		{"500 error", &testError{msg: "status 500 internal server error"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUnauthorizedError(tt.err)
			if got != tt.expected {
				t.Errorf("isUnauthorizedError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParsePinterestDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
		wantZero bool
	}{
		{"empty string", "", time.Time{}, true},
		{"RFC3339", "2024-01-15T10:30:00Z", time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC), false},
		{"RFC3339 with millis", "2024-01-15T10:30:00.123Z", time.Date(2024, 1, 15, 10, 30, 0, 123000000, time.UTC), false},
		{"date only", "2024-01-15", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), false},
		{"invalid format", "not-a-date", time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePinterestDate(tt.input)
			if tt.wantZero {
				if !got.IsZero() {
					t.Errorf("parsePinterestDate(%q) = %v, want zero time", tt.input, got)
				}
			} else {
				if got.IsZero() {
					t.Errorf("parsePinterestDate(%q) returned zero time, want %v", tt.input, tt.expected)
				}
			}
		})
	}
}

func TestSemForAccount(t *testing.T) {
	sem1 := semForAccount("account_1", 1)
	if sem1 == nil {
		t.Fatal("expected non-nil semaphore")
	}

	sem1Again := semForAccount("account_1", 1)
	if sem1 != sem1Again {
		t.Error("expected same semaphore for same account")
	}

	sem2 := semForAccount("account_2", 1)
	if sem1 == sem2 {
		t.Error("expected different semaphore for different account")
	}
}

// ================== WorkOrderMessage Tests ==================

func TestWorkOrderMessage_Fields(t *testing.T) {
	msg := WorkOrderMessage{
		AccountID:   "acc_123",
		Value:       []byte(`{"key": "value"}`),
		AccessToken: "token_xyz",
	}

	if msg.AccountID != "acc_123" {
		t.Errorf("AccountID = %q, want %q", msg.AccountID, "acc_123")
	}
	if string(msg.Value) != `{"key": "value"}` {
		t.Errorf("Value = %q, want %q", string(msg.Value), `{"key": "value"}`)
	}
	if msg.AccessToken != "token_xyz" {
		t.Errorf("AccessToken = %q, want %q", msg.AccessToken, "token_xyz")
	}
}

// ================== TimestampUpdateRequest Tests ==================

func TestTimestampUpdateRequest_Fields(t *testing.T) {
	req := TimestampUpdateRequest{
		AccountID: "acc_123",
		UserID:    "user_456",
	}

	if req.AccountID != "acc_123" {
		t.Errorf("AccountID = %q, want %q", req.AccountID, "acc_123")
	}
	if req.UserID != "user_456" {
		t.Errorf("UserID = %q, want %q", req.UserID, "user_456")
	}
}

// ================== Invalid JSON Tests ==================

func TestService_InvalidJSON(t *testing.T) {
	log := logger.New("error")

	consumer := &kafka.MockConsumerWithMessages{
		Messages: []kafka.MockMessage{
			{Topic: topicWorkOrderBatch, Key: []byte("batch_1"), Value: []byte("invalid-json")},
		},
	}
	producer := NewMockKafkaProducer()
	mongoRepo := NewMockUnifiedSocialRepository()
	pinterestClient := NewMockPinterestClient()

	svc := NewService(pinterestClient, producer, consumer, mongoRepo, log, "test-key")
	svc.IdleTimeout = 100 * time.Millisecond
	svc.IdleCheckPeriod = 20 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := svc.Run(ctx)

	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("Run should handle invalid JSON gracefully: %v", err)
	}
}

// ================== Test Helper Types ==================

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
