package main

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

// ================== NewPageProcessor Tests ==================

func TestNewPageProcessor(t *testing.T) {
	client := &MockLinkedInClient{}
	geoResolver := &MockGeoResolver{}
	producer := &kafka.MockProducer{}
	log := logger.New("error")

	p := NewPageProcessor(client, geoResolver, producer, log)

	if p == nil {
		t.Fatal("NewPageProcessor returned nil")
	}
	if p.client == nil {
		t.Error("client is nil")
	}
	if p.geoResolver == nil {
		t.Error("geoResolver is nil")
	}
	if p.producer == nil {
		t.Error("producer is nil")
	}
}

// ================== ProcessPageWorkOrderTestable Tests ==================

func TestPageProcessor_ProcessPageWorkOrderTestable_Success(t *testing.T) {
	var produceCalls int32

	client := &MockLinkedInClient{
		FetchPostsPaginatedFunc: func(ctx context.Context, linkedinID string, entityType string, accessToken string, cutoffTime time.Time) ([]json.RawMessage, error) {
			return []json.RawMessage{
				json.RawMessage(`{"id":"urn:li:ugcPost:123","content":{}}`),
			}, nil
		},
		FetchStatsRawFunc: func(ctx context.Context, linkedinID string, ugcPosts []string, shares []string, accessToken string) ([]byte, error) {
			return []byte(`{"elements":[]}`), nil
		},
		FetchFollowerStatsWithGeoIDsFunc: func(ctx context.Context, linkedinID string, accessToken string) (*social.FollowerStatsWithGeoIDs, error) {
			return &social.FollowerStatsWithGeoIDs{
				RawStats: []byte(`{"elements":[]}`),
				GeoIDs:   nil,
			}, nil
		},
		BuildFollowerDataWithGeoNamesFunc: func(stats *social.FollowerStatsWithGeoIDs, geoNames map[string]string) ([]byte, error) {
			return []byte(`{"followers":1000}`), nil
		},
		FetchPageStatisticsRawFunc: func(ctx context.Context, linkedinID string, accessToken string, startMs, endMs int64) ([]byte, error) {
			return []byte(`{"pageViews":500}`), nil
		},
		FetchShareStatisticsRawFunc: func(ctx context.Context, linkedinID string, accessToken string, startMs, endMs int64) ([]byte, error) {
			return []byte(`{"shares":100}`), nil
		},
		FetchOrganizationDetailsRawFunc: func(ctx context.Context, linkedinID string, accessToken string) ([]byte, error) {
			return []byte(`{"name":"Test Org"}`), nil
		},
	}

	geoResolver := &MockGeoResolver{}
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddInt32(&produceCalls, 1)
			return nil
		},
	}
	log := logger.New("error")

	p := NewPageProcessor(client, geoResolver, producer, log)

	order := LinkedInAccountWorkOrder{
		ID:          "acc123",
		LinkedinID:  "li456",
		WorkspaceID: "ws789",
		SyncType:    "incremental",
	}

	timestampChan := make(chan TimestampUpdateRequest, 10)

	err := p.ProcessPageWorkOrderTestable(context.Background(), order, "token123", timestampChan)
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

func TestPageProcessor_ProcessPageWorkOrderTestable_TokenError(t *testing.T) {
	client := &MockLinkedInClient{
		FetchPostsPaginatedFunc: func(ctx context.Context, linkedinID string, entityType string, accessToken string, cutoffTime time.Time) ([]json.RawMessage, error) {
			return nil, &testTokenError{err: context.DeadlineExceeded}
		},
		FetchFollowerStatsWithGeoIDsFunc: func(ctx context.Context, linkedinID string, accessToken string) (*social.FollowerStatsWithGeoIDs, error) {
			return nil, &testTokenError{err: context.DeadlineExceeded}
		},
		FetchOrganizationDetailsRawFunc: func(ctx context.Context, linkedinID string, accessToken string) ([]byte, error) {
			return nil, &testTokenError{err: context.DeadlineExceeded}
		},
	}

	geoResolver := &MockGeoResolver{}
	producer := &kafka.MockProducer{}
	log := logger.New("error")

	p := NewPageProcessor(client, geoResolver, producer, log)

	order := LinkedInAccountWorkOrder{
		ID:         "acc123",
		LinkedinID: "li456",
		SyncType:   "incremental",
	}

	timestampChan := make(chan TimestampUpdateRequest, 10)

	err := p.ProcessPageWorkOrderTestable(context.Background(), order, "token123", timestampChan)
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

// ================== fetchAndPublishPagePostsTestable Tests ==================

func TestPageProcessor_fetchAndPublishPagePostsTestable_Success(t *testing.T) {
	var produceCalls int32

	client := &MockLinkedInClient{
		FetchPostsPaginatedFunc: func(ctx context.Context, linkedinID string, entityType string, accessToken string, cutoffTime time.Time) ([]json.RawMessage, error) {
			return []json.RawMessage{
				json.RawMessage(`{"id":"urn:li:ugcPost:123","content":{}}`),
				json.RawMessage(`{"id":"urn:li:share:456","content":{}}`),
			}, nil
		},
		FetchStatsRawFunc: func(ctx context.Context, linkedinID string, ugcPosts []string, shares []string, accessToken string) ([]byte, error) {
			return []byte(`{"elements":[{"share":"urn:li:share:456","totalShareStatistics":{"likeCount":10}}]}`), nil
		},
		FetchImagesRawFunc: func(ctx context.Context, ids []string, accessToken string) ([]byte, error) {
			return nil, nil
		},
		FetchVideosRawFunc: func(ctx context.Context, ids []string, accessToken string) ([]byte, error) {
			return nil, nil
		},
		FetchDocumentsRawFunc: func(ctx context.Context, ids []string, accessToken string) ([]byte, error) {
			return nil, nil
		},
	}

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddInt32(&produceCalls, 1)
			return nil
		},
	}
	log := logger.New("error")

	p := NewPageProcessor(client, nil, producer, log)

	order := &LinkedInAccountWorkOrder{
		ID:         "acc123",
		LinkedinID: "li456",
	}

	err := p.fetchAndPublishPagePostsTestable(context.Background(), order, "token123", time.Now().AddDate(0, -1, 0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if atomic.LoadInt32(&produceCalls) != 2 {
		t.Errorf("produce calls = %d, want 2", produceCalls)
	}
}

func TestPageProcessor_fetchAndPublishPagePostsTestable_EmptyPosts(t *testing.T) {
	client := &MockLinkedInClient{
		FetchPostsPaginatedFunc: func(ctx context.Context, linkedinID string, entityType string, accessToken string, cutoffTime time.Time) ([]json.RawMessage, error) {
			return nil, nil
		},
	}

	producer := &kafka.MockProducer{}
	log := logger.New("error")

	p := NewPageProcessor(client, nil, producer, log)

	order := &LinkedInAccountWorkOrder{
		ID:         "acc123",
		LinkedinID: "li456",
	}

	err := p.fetchAndPublishPagePostsTestable(context.Background(), order, "token123", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ================== fetchPageAnalyticsTestable Tests ==================

func TestPageProcessor_fetchPageAnalyticsTestable_Success(t *testing.T) {
	client := &MockLinkedInClient{
		FetchFollowerStatsWithGeoIDsFunc: func(ctx context.Context, linkedinID string, accessToken string) (*social.FollowerStatsWithGeoIDs, error) {
			return &social.FollowerStatsWithGeoIDs{
				RawStats: []byte(`{}`),
				GeoIDs:   []social.GeoIDWithType{{ID: "123", Type: "country"}},
			}, nil
		},
		BuildFollowerDataWithGeoNamesFunc: func(stats *social.FollowerStatsWithGeoIDs, geoNames map[string]string) ([]byte, error) {
			return []byte(`{"followers":1000}`), nil
		},
		FetchPageStatisticsRawFunc: func(ctx context.Context, linkedinID string, accessToken string, startMs, endMs int64) ([]byte, error) {
			return []byte(`{"pageViews":500}`), nil
		},
		FetchShareStatisticsRawFunc: func(ctx context.Context, linkedinID string, accessToken string, startMs, endMs int64) ([]byte, error) {
			return []byte(`{"shares":100}`), nil
		},
	}

	geoResolver := &MockGeoResolver{
		ResolveGeoIDsWithTypeFunc: func(ctx context.Context, geoIDs []social.GeoIDWithType, accessToken string) (map[string]string, error) {
			return map[string]string{"123": "United States"}, nil
		},
	}

	log := logger.New("error")
	p := NewPageProcessor(client, geoResolver, nil, log)

	results, err := p.fetchPageAnalyticsTestable(context.Background(), "li123", "token123", time.Now().AddDate(0, -1, 0), time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results.FollowerData == nil {
		t.Error("expected FollowerData")
	}
	if results.PageStats == nil {
		t.Error("expected PageStats")
	}
	if results.ShareStats == nil {
		t.Error("expected ShareStats")
	}
}

func TestPageProcessor_fetchPageAnalyticsTestable_TokenError(t *testing.T) {
	client := &MockLinkedInClient{
		FetchFollowerStatsWithGeoIDsFunc: func(ctx context.Context, linkedinID string, accessToken string) (*social.FollowerStatsWithGeoIDs, error) {
			return nil, &testTokenError{err: context.DeadlineExceeded}
		},
		FetchPageStatisticsRawFunc: func(ctx context.Context, linkedinID string, accessToken string, startMs, endMs int64) ([]byte, error) {
			return nil, nil
		},
		FetchShareStatisticsRawFunc: func(ctx context.Context, linkedinID string, accessToken string, startMs, endMs int64) ([]byte, error) {
			return nil, nil
		},
	}

	geoResolver := &MockGeoResolver{}
	log := logger.New("error")
	p := NewPageProcessor(client, geoResolver, nil, log)

	_, err := p.fetchPageAnalyticsTestable(context.Background(), "li123", "token123", time.Now().AddDate(0, -1, 0), time.Now())
	if err == nil {
		t.Error("expected error for token error")
	}
}

// ================== publishPageInsightsTestable Tests ==================

func TestPageProcessor_publishPageInsightsTestable_Success(t *testing.T) {
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
	p := NewPageProcessor(nil, nil, producer, log)

	results := &pageAnalyticsResults{
		FollowerData: []byte(`{"followers":1000}`),
		PageStats:    []byte(`{"pageViews":500}`),
		ShareStats:   []byte(`{"shares":100}`),
	}

	err := p.publishPageInsightsTestable(context.Background(), "li123", results)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if atomic.LoadInt32(&produceCalls) != 1 {
		t.Errorf("produce calls = %d, want 1", produceCalls)
	}
	if lastTopic != pageInsightsTopic {
		t.Errorf("topic = %q, want %q", lastTopic, pageInsightsTopic)
	}
}

func TestPageProcessor_publishPageInsightsTestable_NoData(t *testing.T) {
	var produceCalls int32

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddInt32(&produceCalls, 1)
			return nil
		},
	}
	log := logger.New("error")
	p := NewPageProcessor(nil, nil, producer, log)

	results := &pageAnalyticsResults{}

	err := p.publishPageInsightsTestable(context.Background(), "li123", results)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if atomic.LoadInt32(&produceCalls) != 0 {
		t.Errorf("produce calls = %d, want 0 (no data)", produceCalls)
	}
}

// ================== fetchAndPublishOrgDetailsTestable Tests ==================

func TestPageProcessor_fetchAndPublishOrgDetailsTestable_Success(t *testing.T) {
	var produceCalls int32

	client := &MockLinkedInClient{
		FetchOrganizationDetailsRawFunc: func(ctx context.Context, linkedinID string, accessToken string) ([]byte, error) {
			return []byte(`{"name":"Test Org"}`), nil
		},
	}

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddInt32(&produceCalls, 1)
			return nil
		},
	}
	log := logger.New("error")
	p := NewPageProcessor(client, nil, producer, log)

	err := p.fetchAndPublishOrgDetailsTestable(context.Background(), "li123", "token123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if atomic.LoadInt32(&produceCalls) != 1 {
		t.Errorf("produce calls = %d, want 1", produceCalls)
	}
}

func TestPageProcessor_fetchAndPublishOrgDetailsTestable_TokenError(t *testing.T) {
	client := &MockLinkedInClient{
		FetchOrganizationDetailsRawFunc: func(ctx context.Context, linkedinID string, accessToken string) ([]byte, error) {
			return nil, &testTokenError{err: context.DeadlineExceeded}
		},
	}

	log := logger.New("error")
	p := NewPageProcessor(client, nil, nil, log)

	err := p.fetchAndPublishOrgDetailsTestable(context.Background(), "li123", "token123")
	if !isTokenError(err) {
		t.Errorf("expected token error, got %v", err)
	}
}

// ================== Helper Tests ==================

func TestPageProcessor_fetchPostAssetsTestable(t *testing.T) {
	client := &MockLinkedInClient{
		FetchStatsRawFunc: func(ctx context.Context, linkedinID string, ugcPosts []string, shares []string, accessToken string) ([]byte, error) {
			return []byte(`{"elements":[{"share":"urn:li:share:456","totalShareStatistics":{"likeCount":10}}]}`), nil
		},
		FetchImagesRawFunc: func(ctx context.Context, ids []string, accessToken string) ([]byte, error) {
			// The parseAssetBatch function expects results with "id" field
			return []byte(`{"results":{"urn:li:image:img1":{"id":"urn:li:image:img1","downloadUrl":"http://example.com/img1.jpg"}}}`), nil
		},
		FetchVideosRawFunc: func(ctx context.Context, ids []string, accessToken string) ([]byte, error) {
			return nil, nil
		},
		FetchDocumentsRawFunc: func(ctx context.Context, ids []string, accessToken string) ([]byte, error) {
			return nil, nil
		},
	}

	log := logger.New("error")
	p := NewPageProcessor(client, nil, nil, log)

	statsMap, imgMap, _, _ := p.fetchPostAssetsTestable(
		context.Background(),
		"li123",
		"token123",
		[]string{"urn:li:ugcPost:123"},
		[]string{"urn:li:share:456"},
		[]string{"urn:li:image:img1"},
		nil,
		nil,
	)

	if len(statsMap) == 0 {
		t.Error("expected stats")
	}
	if len(imgMap) == 0 {
		t.Error("expected images")
	}
}

// testTokenError helper for testing - matches isTokenError patterns
type testTokenError struct {
	err error
}

func (e *testTokenError) Error() string {
	return "401 Unauthorized: " + e.err.Error()
}
