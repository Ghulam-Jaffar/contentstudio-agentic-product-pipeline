package main

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

// ================== NewProfileProcessor Tests ==================

func TestNewProfileProcessor(t *testing.T) {
	client := &MockLinkedInClient{}
	producer := &kafka.MockProducer{}
	log := logger.New("error")

	p := NewProfileProcessor(client, producer, log)

	if p == nil {
		t.Fatal("NewProfileProcessor returned nil")
	}
	if p.client == nil {
		t.Error("client is nil")
	}
	if p.producer == nil {
		t.Error("producer is nil")
	}
}

// ================== ProcessProfileWorkOrderTestable Tests ==================

func TestProfileProcessor_ProcessProfileWorkOrderTestable_Success(t *testing.T) {
	var produceCalls int32

	client := &MockLinkedInClient{
		FetchMemberCreatorPostAnalyticsRawFunc: func(ctx context.Context, accessToken string, queryType string, startDate, endDate *time.Time) ([]byte, error) {
			return []byte(`{"analytics":"data"}`), nil
		},
		FetchMemberFollowersCountRawFunc: func(ctx context.Context, accessToken string, startDate, endDate *time.Time) ([]byte, error) {
			return []byte(`{"followers":500}`), nil
		},
	}

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddInt32(&produceCalls, 1)
			return nil
		},
	}
	log := logger.New("error")

	p := NewProfileProcessor(client, producer, log)

	order := LinkedInAccountWorkOrder{
		ID:          "acc123",
		LinkedinID:  "li456",
		WorkspaceID: "ws789",
		SyncType:    "incremental",
	}

	timestampChan := make(chan TimestampUpdateRequest, 10)

	err := p.ProcessProfileWorkOrderTestable(context.Background(), order, "token123", timestampChan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have received timestamp update
	select {
	case req := <-timestampChan:
		if req.AccountID != "acc123" {
			t.Errorf("AccountID = %q, want %q", req.AccountID, "acc123")
		}
	case <-time.After(time.Second):
		t.Error("expected timestamp update")
	}

	// Should have produced messages
	if atomic.LoadInt32(&produceCalls) < 1 {
		t.Errorf("produce calls = %d, want >= 1", produceCalls)
	}
}

func TestProfileProcessor_ProcessProfileWorkOrderTestable_TokenError(t *testing.T) {
	client := &MockLinkedInClient{
		FetchMemberCreatorPostAnalyticsRawFunc: func(ctx context.Context, accessToken string, queryType string, startDate, endDate *time.Time) ([]byte, error) {
			return nil, &testTokenError{err: context.DeadlineExceeded}
		},
		FetchMemberFollowersCountRawFunc: func(ctx context.Context, accessToken string, startDate, endDate *time.Time) ([]byte, error) {
			return nil, &testTokenError{err: context.DeadlineExceeded}
		},
	}

	producer := &kafka.MockProducer{}
	log := logger.New("error")

	p := NewProfileProcessor(client, producer, log)

	order := LinkedInAccountWorkOrder{
		ID:         "acc123",
		LinkedinID: "li456",
		SyncType:   "incremental",
	}

	timestampChan := make(chan TimestampUpdateRequest, 10)

	err := p.ProcessProfileWorkOrderTestable(context.Background(), order, "token123", timestampChan)
	// Token errors should return nil (handled gracefully)
	if err != nil {
		t.Errorf("expected nil error for token error, got %v", err)
	}

	// Should NOT have received timestamp update
	select {
	case <-timestampChan:
		t.Error("should not receive timestamp update on token error")
	default:
		// Expected
	}
}

// ================== fetchProfileAnalyticsTestable Tests ==================

func TestProfileProcessor_fetchProfileAnalyticsTestable_Success(t *testing.T) {
	client := &MockLinkedInClient{
		FetchMemberCreatorPostAnalyticsRawFunc: func(ctx context.Context, accessToken string, queryType string, startDate, endDate *time.Time) ([]byte, error) {
			return []byte(`{"` + queryType + `":"data"}`), nil
		},
		FetchMemberFollowersCountRawFunc: func(ctx context.Context, accessToken string, startDate, endDate *time.Time) ([]byte, error) {
			if startDate == nil && endDate == nil {
				return []byte(`{"totalFollowers":1000}`), nil
			}
			return []byte(`{"dailyFollowers":[]}`), nil
		},
	}

	log := logger.New("error")
	p := NewProfileProcessor(client, nil, log)

	results, err := p.fetchProfileAnalyticsTestable(context.Background(), "li123", "token123", time.Now().AddDate(0, -1, 0), time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results.ImpressionData == nil {
		t.Error("expected ImpressionData")
	}
	if results.MembersReachedData == nil {
		t.Error("expected MembersReachedData")
	}
	if results.ReshareData == nil {
		t.Error("expected ReshareData")
	}
	if results.ReactionData == nil {
		t.Error("expected ReactionData")
	}
	if results.CommentData == nil {
		t.Error("expected CommentData")
	}
	if results.FollowerData == nil {
		t.Error("expected FollowerData")
	}
	if results.TotalFollowerData == nil {
		t.Error("expected TotalFollowerData")
	}
}

func TestProfileProcessor_fetchProfileAnalyticsTestable_TokenError(t *testing.T) {
	client := &MockLinkedInClient{
		FetchMemberCreatorPostAnalyticsRawFunc: func(ctx context.Context, accessToken string, queryType string, startDate, endDate *time.Time) ([]byte, error) {
			return nil, &testTokenError{err: context.DeadlineExceeded}
		},
		FetchMemberFollowersCountRawFunc: func(ctx context.Context, accessToken string, startDate, endDate *time.Time) ([]byte, error) {
			return nil, nil
		},
	}

	log := logger.New("error")
	p := NewProfileProcessor(client, nil, log)

	_, err := p.fetchProfileAnalyticsTestable(context.Background(), "li123", "token123", time.Now().AddDate(0, -1, 0), time.Now())
	if err == nil {
		t.Error("expected error for token error")
	}
}

func TestProfileProcessor_fetchProfileAnalyticsTestable_PartialSuccess(t *testing.T) {
	client := &MockLinkedInClient{
		FetchMemberCreatorPostAnalyticsRawFunc: func(ctx context.Context, accessToken string, queryType string, startDate, endDate *time.Time) ([]byte, error) {
			// Only return data for IMPRESSION
			if queryType == "IMPRESSION" {
				return []byte(`{"impressions":100}`), nil
			}
			return nil, context.DeadlineExceeded // Non-token error
		},
		FetchMemberFollowersCountRawFunc: func(ctx context.Context, accessToken string, startDate, endDate *time.Time) ([]byte, error) {
			return []byte(`{"followers":500}`), nil
		},
	}

	log := logger.New("error")
	p := NewProfileProcessor(client, nil, log)

	results, err := p.fetchProfileAnalyticsTestable(context.Background(), "li123", "token123", time.Now().AddDate(0, -1, 0), time.Now())
	// Non-token errors don't fail the whole batch
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have partial results
	if results.ImpressionData == nil {
		t.Error("expected ImpressionData")
	}
	if results.FollowerData == nil {
		t.Error("expected FollowerData")
	}
}

// ================== publishProfileInsightsTestable Tests ==================

func TestProfileProcessor_publishProfileInsightsTestable_Success(t *testing.T) {
	var produceCalls int32
	var lastTopic string

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddInt32(&produceCalls, 1)
			lastTopic = topic
			return nil
		},
	}
	log := logger.New("error")
	p := NewProfileProcessor(nil, producer, log)

	results := &profileAnalyticsResults{
		ImpressionData:     []byte(`{"impressions":100}`),
		MembersReachedData: []byte(`{"reached":50}`),
		ReshareData:        []byte(`{"reshares":10}`),
		ReactionData:       []byte(`{"reactions":25}`),
		CommentData:        []byte(`{"comments":5}`),
		FollowerData:       []byte(`{"dailyFollowers":[]}`),
		TotalFollowerData:  []byte(`{"totalFollowers":1000}`),
	}

	err := p.publishProfileInsightsTestable(context.Background(), "li123", results)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if atomic.LoadInt32(&produceCalls) != 1 {
		t.Errorf("produce calls = %d, want 1", produceCalls)
	}
	if lastTopic != profileInsightsTopic {
		t.Errorf("topic = %q, want %q", lastTopic, profileInsightsTopic)
	}
}

func TestProfileProcessor_publishProfileInsightsTestable_NoData(t *testing.T) {
	var produceCalls int32

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddInt32(&produceCalls, 1)
			return nil
		},
	}
	log := logger.New("error")
	p := NewProfileProcessor(nil, producer, log)

	results := &profileAnalyticsResults{}

	err := p.publishProfileInsightsTestable(context.Background(), "li123", results)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if atomic.LoadInt32(&produceCalls) != 0 {
		t.Errorf("produce calls = %d, want 0 (no data)", produceCalls)
	}
}

func TestProfileProcessor_publishProfileInsightsTestable_PartialData(t *testing.T) {
	var produceCalls int32

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddInt32(&produceCalls, 1)
			return nil
		},
	}
	log := logger.New("error")
	p := NewProfileProcessor(nil, producer, log)

	// Only some fields populated
	results := &profileAnalyticsResults{
		ImpressionData: []byte(`{"impressions":100}`),
		FollowerData:   []byte(`{"dailyFollowers":[]}`),
	}

	err := p.publishProfileInsightsTestable(context.Background(), "li123", results)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if atomic.LoadInt32(&produceCalls) != 1 {
		t.Errorf("produce calls = %d, want 1", produceCalls)
	}
}

// ================== fetchAndPublishProfileInsightsTestable Tests ==================

func TestProfileProcessor_fetchAndPublishProfileInsightsTestable_Success(t *testing.T) {
	var produceCalls int32

	client := &MockLinkedInClient{
		FetchMemberCreatorPostAnalyticsRawFunc: func(ctx context.Context, accessToken string, queryType string, startDate, endDate *time.Time) ([]byte, error) {
			return []byte(`{"data":"value"}`), nil
		},
		FetchMemberFollowersCountRawFunc: func(ctx context.Context, accessToken string, startDate, endDate *time.Time) ([]byte, error) {
			return []byte(`{"followers":500}`), nil
		},
	}

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddInt32(&produceCalls, 1)
			return nil
		},
	}
	log := logger.New("error")
	p := NewProfileProcessor(client, producer, log)

	err := p.fetchAndPublishProfileInsightsTestable(context.Background(), "li123", "token123", time.Now().AddDate(0, -1, 0), time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if atomic.LoadInt32(&produceCalls) != 1 {
		t.Errorf("produce calls = %d, want 1", produceCalls)
	}
}

func TestProfileProcessor_fetchAndPublishProfileInsightsTestable_TokenError(t *testing.T) {
	client := &MockLinkedInClient{
		FetchMemberCreatorPostAnalyticsRawFunc: func(ctx context.Context, accessToken string, queryType string, startDate, endDate *time.Time) ([]byte, error) {
			return nil, &testTokenError{err: context.DeadlineExceeded}
		},
		FetchMemberFollowersCountRawFunc: func(ctx context.Context, accessToken string, startDate, endDate *time.Time) ([]byte, error) {
			return nil, nil
		},
	}

	log := logger.New("error")
	p := NewProfileProcessor(client, nil, log)

	err := p.fetchAndPublishProfileInsightsTestable(context.Background(), "li123", "token123", time.Now().AddDate(0, -1, 0), time.Now())
	if !isTokenError(err) {
		t.Errorf("expected token error, got %v", err)
	}
}
