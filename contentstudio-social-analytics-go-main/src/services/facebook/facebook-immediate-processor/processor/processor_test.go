package processor

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/parsing"
)

func TestWorkOrder_Struct(t *testing.T) {
	wo := WorkOrder{
		ID:              "account123",
		AccountID:       "fb_123456",
		Type:            "Page",
		AccessToken:     "token_abc",
		WorkspaceID:     "workspace_789",
		LongAccessToken: "long_token_xyz",
		SyncType:        "full",
	}

	if wo.ID != "account123" {
		t.Fatalf("expected ID 'account123', got '%s'", wo.ID)
	}
	if wo.AccountID != "fb_123456" {
		t.Fatalf("expected AccountID 'fb_123456', got '%s'", wo.AccountID)
	}
	if wo.Type != "Page" {
		t.Fatalf("expected Type 'Page', got '%s'", wo.Type)
	}
	if wo.SyncType != "full" {
		t.Fatalf("expected SyncType 'full', got '%s'", wo.SyncType)
	}
}

func TestParsedData_Struct(t *testing.T) {
	parsed := &ParsedData{
		Posts:         []kafkamodels.ParsedFacebookPost{},
		MediaAssets:   []kafkamodels.ParsedFacebookMediaAsset{},
		VideoInsights: []kafkamodels.ParsedFacebookVideoInsights{},
		ReelsInsights: []kafkamodels.ParsedFacebookReelsInsights{},
		Insights:      []*kafkamodels.ParsedFacebookInsights{},
	}

	if parsed.Posts == nil {
		t.Fatal("expected Posts to be initialized")
	}
	if parsed.MediaAssets == nil {
		t.Fatal("expected MediaAssets to be initialized")
	}
	if parsed.VideoInsights == nil {
		t.Fatal("expected VideoInsights to be initialized")
	}
	if parsed.ReelsInsights == nil {
		t.Fatal("expected ReelsInsights to be initialized")
	}
	if parsed.Insights == nil {
		t.Fatal("expected Insights to be initialized")
	}
}

func TestParsedData_WithData(t *testing.T) {
	parsed := &ParsedData{
		Posts: []kafkamodels.ParsedFacebookPost{
			{PostID: "post1", PageID: "page1"},
			{PostID: "post2", PageID: "page1"},
		},
		MediaAssets: []kafkamodels.ParsedFacebookMediaAsset{
			{MediaID: "media1", PostID: "post1"},
		},
		VideoInsights: []kafkamodels.ParsedFacebookVideoInsights{
			{PostID: "video1", PageID: "page1"},
		},
		ReelsInsights: []kafkamodels.ParsedFacebookReelsInsights{
			{PostID: "reel1", PageID: "page1"},
		},
		Insights: []*kafkamodels.ParsedFacebookInsights{
			{PageID: "page1"},
		},
	}

	if len(parsed.Posts) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(parsed.Posts))
	}
	if len(parsed.MediaAssets) != 1 {
		t.Fatalf("expected 1 media asset, got %d", len(parsed.MediaAssets))
	}
	if len(parsed.VideoInsights) != 1 {
		t.Fatalf("expected 1 video insight, got %d", len(parsed.VideoInsights))
	}
	if len(parsed.ReelsInsights) != 1 {
		t.Fatalf("expected 1 reel insight, got %d", len(parsed.ReelsInsights))
	}
	if len(parsed.Insights) != 1 {
		t.Fatalf("expected 1 page insight, got %d", len(parsed.Insights))
	}
}

func TestProcessor_Struct(t *testing.T) {
	p := &Processor{
		MongoRepo:      nil,
		FacebookClient: nil,
		Parser:         nil,
		Sink:           nil,
		Notifier:       nil,
		PusherClient:   nil,
		Producer:       nil,
		Logger:         nil,
		Config:         nil,
	}

	if p.MongoRepo != nil {
		t.Fatal("expected nil MongoRepo")
	}
	if p.FacebookClient != nil {
		t.Fatal("expected nil FacebookClient")
	}
}

func TestGetStringFromExtraData(t *testing.T) {
	cases := []struct {
		name      string
		extraData map[string]interface{}
		key       string
		expected  string
	}{
		{
			name:      "nil map",
			extraData: nil,
			key:       "key",
			expected:  "",
		},
		{
			name:      "empty map",
			extraData: map[string]interface{}{},
			key:       "key",
			expected:  "",
		},
		{
			name: "key exists with string value",
			extraData: map[string]interface{}{
				"name": "Test Page",
			},
			key:      "name",
			expected: "Test Page",
		},
		{
			name: "key exists with non-string value",
			extraData: map[string]interface{}{
				"count": 123,
			},
			key:      "count",
			expected: "",
		},
		{
			name: "key does not exist",
			extraData: map[string]interface{}{
				"other": "value",
			},
			key:      "missing",
			expected: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := getStringFromExtraData(tc.extraData, tc.key)
			if result != tc.expected {
				t.Fatalf("expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestWorkOrder_EmptyFields(t *testing.T) {
	wo := WorkOrder{}

	if wo.ID != "" {
		t.Fatal("expected empty ID")
	}
	if wo.AccountID != "" {
		t.Fatal("expected empty AccountID")
	}
	if wo.Type != "" {
		t.Fatal("expected empty Type")
	}
}

func TestParsedData_NilFields(t *testing.T) {
	parsed := &ParsedData{}

	if parsed.Posts != nil {
		t.Fatal("expected nil Posts")
	}
	if parsed.MediaAssets != nil {
		t.Fatal("expected nil MediaAssets")
	}
	if parsed.VideoInsights != nil {
		t.Fatal("expected nil VideoInsights")
	}
}

// Mock ClickHouse sink tests
func TestMockClickHouseSink_BulkInsertPosts_Success(t *testing.T) {
	called := false
	mock := &mockClickHouseSink{
		BulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			called = true
			if len(posts) != 2 {
				t.Fatalf("expected 2 posts, got %d", len(posts))
			}
			return nil
		},
	}

	posts := []*clickhousemodels.FacebookPosts{
		{PageID: "page1", PostID: "post1"},
		{PageID: "page1", PostID: "post2"},
	}

	err := mock.BulkInsertPosts(context.Background(), posts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("BulkInsertPosts was not called")
	}
}

func TestMockClickHouseSink_BulkInsertPosts_Error(t *testing.T) {
	expectedErr := errors.New("insert failed")
	mock := &mockClickHouseSink{
		BulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			return expectedErr
		},
	}

	err := mock.BulkInsertPosts(context.Background(), nil)
	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestMockClickHouseSink_BulkInsertMediaAssets_Success(t *testing.T) {
	called := false
	mock := &mockClickHouseSink{
		BulkInsertMediaAssetsFunc: func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
			called = true
			return nil
		},
	}

	assets := []*clickhousemodels.FacebookMediaAssets{
		{MediaID: "media1", PostID: "post1"},
	}

	err := mock.BulkInsertMediaAssets(context.Background(), assets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("BulkInsertMediaAssets was not called")
	}
}

func TestMockClickHouseSink_BulkInsertVideoInsights_Success(t *testing.T) {
	called := false
	mock := &mockClickHouseSink{
		BulkInsertVideoInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
			called = true
			return nil
		},
	}

	insights := []*clickhousemodels.FacebookVideoInsights{
		{PageID: "page1", PostID: "video1"},
	}

	err := mock.BulkInsertVideoInsights(context.Background(), insights)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("BulkInsertVideoInsights was not called")
	}
}

func TestMockClickHouseSink_BulkInsertReelsInsights_Success(t *testing.T) {
	called := false
	mock := &mockClickHouseSink{
		BulkInsertReelsInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
			called = true
			return nil
		},
	}

	insights := []*clickhousemodels.FacebookReelsInsights{
		{PageID: "page1", PostID: "reel1"},
	}

	err := mock.BulkInsertReelsInsights(context.Background(), insights)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("BulkInsertReelsInsights was not called")
	}
}

func TestMockClickHouseSink_BulkInsertInsights_Success(t *testing.T) {
	called := false
	mock := &mockClickHouseSink{
		BulkInsertInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
			called = true
			return nil
		},
	}

	insights := []*clickhousemodels.FacebookInsights{
		{PageID: "page1"},
	}

	err := mock.BulkInsertInsights(context.Background(), insights)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("BulkInsertInsights was not called")
	}
}

func TestMockClickHouseSink_NilFunctions(t *testing.T) {
	mock := &mockClickHouseSink{}

	// All BulkInsert methods should return nil when function is not set
	if err := mock.BulkInsertPosts(context.Background(), nil); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if err := mock.BulkInsertMediaAssets(context.Background(), nil); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if err := mock.BulkInsertVideoInsights(context.Background(), nil); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if err := mock.BulkInsertReelsInsights(context.Background(), nil); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if err := mock.BulkInsertInsights(context.Background(), nil); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestClickHouseSinkInterface_Implementation(t *testing.T) {
	// This test verifies the interface is properly defined
	var sink ClickHouseSinkInterface = &mockClickHouseSink{}
	if sink == nil {
		t.Fatal("expected non-nil interface")
	}
}

// ================== Processor Tests with Mocks ==================

func TestProcessor_New(t *testing.T) {
	mockRepo := &mockMongoRepo{}
	mockSink := &mockClickHouseSink{}
	mockProducer := &mockKafkaProducer{}

	p := &Processor{
		MongoRepo: mockRepo,
		Sink:      mockSink,
		Producer:  mockProducer,
	}

	if p.MongoRepo == nil {
		t.Fatal("expected MongoRepo to be set")
	}
	if p.Sink == nil {
		t.Fatal("expected Sink to be set")
	}
	if p.Producer == nil {
		t.Fatal("expected Producer to be set")
	}
}

func TestMockMongoRepo_FindByID_Success(t *testing.T) {
	expectedID := primitive.NewObjectID()
	expectedAccount := &mongomodels.SocialIntegration{
		ID:                 expectedID,
		PlatformIdentifier: "fb_123456",
		PlatformName:       "Test Page",
		State:              "Added",
		ExtraData: map[string]interface{}{
			"workspace_id": "ws123",
			"name":         "Test Page Name",
		},
	}

	mock := &mockMongoRepo{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			if id != expectedID {
				t.Fatalf("expected ID %v, got %v", expectedID, id)
			}
			return expectedAccount, nil
		},
	}

	account, err := mock.FindByID(context.Background(), expectedID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if account == nil {
		t.Fatal("expected non-nil account")
	}
	if account.PlatformIdentifier != "fb_123456" {
		t.Fatalf("expected PlatformIdentifier 'fb_123456', got '%s'", account.PlatformIdentifier)
	}
}

func TestMockMongoRepo_FindByID_NotFound(t *testing.T) {
	mock := &mockMongoRepo{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return nil, errors.New("account not found")
		},
	}

	account, err := mock.FindByID(context.Background(), primitive.NewObjectID())
	if err == nil {
		t.Fatal("expected error")
	}
	if account != nil {
		t.Fatal("expected nil account")
	}
}

func TestMockMongoRepo_UpdateState(t *testing.T) {
	stateUpdated := ""
	mock := &mockMongoRepo{
		UpdateStateFunc: func(ctx context.Context, id primitive.ObjectID, state string) error {
			stateUpdated = state
			return nil
		},
	}

	err := mock.UpdateState(context.Background(), primitive.NewObjectID(), "Processing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stateUpdated != "Processing" {
		t.Fatalf("expected state 'Processing', got '%s'", stateUpdated)
	}
}

func TestMockMongoRepo_UpdateAnalyticsTimestamp(t *testing.T) {
	fieldUpdated := ""
	mock := &mockMongoRepo{
		UpdateAnalyticsTimestampFunc: func(ctx context.Context, id primitive.ObjectID, field string, timestamp time.Time) error {
			fieldUpdated = field
			return nil
		},
	}

	err := mock.UpdateAnalyticsTimestamp(context.Background(), primitive.NewObjectID(), "analytics", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fieldUpdated != "analytics" {
		t.Fatalf("expected field 'analytics', got '%s'", fieldUpdated)
	}
}

func TestMockKafkaProducer_Produce_Success(t *testing.T) {
	topicReceived := ""
	mock := &mockKafkaProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			topicReceived = topic
			return nil
		},
	}

	err := mock.Produce(context.Background(), "test-topic", []byte("key"), []byte("value"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if topicReceived != "test-topic" {
		t.Fatalf("expected topic 'test-topic', got '%s'", topicReceived)
	}
}

func TestMockKafkaProducer_Produce_Error(t *testing.T) {
	expectedErr := errors.New("kafka unavailable")
	mock := &mockKafkaProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return expectedErr
		},
	}

	err := mock.Produce(context.Background(), "test-topic", nil, nil)
	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestMockKafkaProducer_Close(t *testing.T) {
	mock := &mockKafkaProducer{}
	err := mock.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProcessor_WithMockedDependencies(t *testing.T) {
	// Test creating a processor with all mocked dependencies
	accountID := primitive.NewObjectID()

	mockRepo := &mockMongoRepo{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:                 accountID,
				PlatformIdentifier: "fb_page_123",
				AccessToken:        "encrypted_token",
				State:              "Added",
				ExtraData: map[string]interface{}{
					"workspace_id": "workspace_456",
					"name":         "My Facebook Page",
				},
			}, nil
		},
		UpdateStateFunc: func(ctx context.Context, id primitive.ObjectID, state string) error {
			return nil
		},
		UpdateAnalyticsTimestampFunc: func(ctx context.Context, id primitive.ObjectID, field string, timestamp time.Time) error {
			return nil
		},
	}

	postsInserted := 0
	mockCH := &mockClickHouseSink{
		BulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			postsInserted = len(posts)
			return nil
		},
		BulkInsertMediaAssetsFunc: func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
			return nil
		},
		BulkInsertVideoInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
			return nil
		},
		BulkInsertReelsInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
			return nil
		},
		BulkInsertInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
			return nil
		},
	}

	messagesProduced := 0
	mockProducer := &mockKafkaProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			messagesProduced++
			return nil
		},
	}

	p := &Processor{
		MongoRepo: mockRepo,
		Sink:      mockCH,
		Producer:  mockProducer,
	}

	// Verify processor is properly configured
	if p.MongoRepo == nil {
		t.Fatal("MongoRepo should not be nil")
	}
	if p.Sink == nil {
		t.Fatal("Sink should not be nil")
	}
	if p.Producer == nil {
		t.Fatal("Producer should not be nil")
	}

	// Test that mocks work correctly
	account, err := p.MongoRepo.FindByID(context.Background(), accountID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if account.PlatformIdentifier != "fb_page_123" {
		t.Fatalf("unexpected PlatformIdentifier: %s", account.PlatformIdentifier)
	}

	// Test ClickHouse insert
	err = p.Sink.BulkInsertPosts(context.Background(), []*clickhousemodels.FacebookPosts{
		{PageID: "page1", PostID: "post1"},
		{PageID: "page1", PostID: "post2"},
	})
	if err != nil {
		t.Fatalf("BulkInsertPosts failed: %v", err)
	}
	if postsInserted != 2 {
		t.Fatalf("expected 2 posts inserted, got %d", postsInserted)
	}

	// Test Kafka producer
	err = p.Producer.Produce(context.Background(), "test-topic", []byte("key"), []byte("value"))
	if err != nil {
		t.Fatalf("Produce failed: %v", err)
	}
	if messagesProduced != 1 {
		t.Fatalf("expected 1 message produced, got %d", messagesProduced)
	}
}

func TestGetStringFromExtraData_AllCases(t *testing.T) {
	cases := []struct {
		name      string
		extraData map[string]interface{}
		key       string
		expected  string
	}{
		{"nil map returns empty", nil, "key", ""},
		{"empty map returns empty", map[string]interface{}{}, "key", ""},
		{"key exists string", map[string]interface{}{"name": "Test"}, "name", "Test"},
		{"key exists int returns empty", map[string]interface{}{"count": 123}, "count", ""},
		{"key exists bool returns empty", map[string]interface{}{"active": true}, "active", ""},
		{"key missing returns empty", map[string]interface{}{"other": "val"}, "missing", ""},
		{"workspace_id", map[string]interface{}{"workspace_id": "ws123"}, "workspace_id", "ws123"},
		{"access_token", map[string]interface{}{"access_token": "token123"}, "access_token", "token123"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := getStringFromExtraData(tc.extraData, tc.key)
			if result != tc.expected {
				t.Fatalf("expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// ================== Facebook Client Mock Tests ==================

func TestMockFacebookClient_FetchPostsSince_Success(t *testing.T) {
	expectedPosts := []kafkamodels.RawFacebookPost{
		{ID: "post1", Message: "Hello World", CreatedTime: kafkamodels.FacebookTime{Time: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)}},
		{ID: "post2", Message: "Another post", CreatedTime: kafkamodels.FacebookTime{Time: time.Date(2024, 1, 14, 10, 0, 0, 0, time.UTC)}},
	}

	mock := &mockFacebookClient{
		fetchPostsSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookPost, error) {
			if pageID != "page123" {
				t.Fatalf("expected pageID 'page123', got '%s'", pageID)
			}
			if accessToken != "token_abc" {
				t.Fatalf("expected accessToken 'token_abc', got '%s'", accessToken)
			}
			return expectedPosts, nil
		},
	}

	posts, err := mock.FetchPostsSince(context.Background(), "page123", "token_abc", time.Now().Add(-24*time.Hour), time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(posts) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(posts))
	}
	if posts[0].ID != "post1" {
		t.Fatalf("expected first post ID 'post1', got '%s'", posts[0].ID)
	}
}

func TestMockFacebookClient_FetchPostsSince_Error(t *testing.T) {
	expectedErr := errors.New("API rate limit exceeded")
	mock := &mockFacebookClient{
		fetchPostsSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookPost, error) {
			return nil, expectedErr
		},
	}

	posts, err := mock.FetchPostsSince(context.Background(), "page123", "token_abc", time.Now(), time.Now())
	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	if posts != nil {
		t.Fatal("expected nil posts on error")
	}
}

func TestMockFacebookClient_FetchVideosSince_Success(t *testing.T) {
	expectedVideos := []kafkamodels.RawFacebookVideo{
		{ID: "video1", Description: "My video"},
	}

	mock := &mockFacebookClient{
		fetchVideosSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookVideo, error) {
			return expectedVideos, nil
		},
	}

	videos, err := mock.FetchVideosSince(context.Background(), "page123", "token_abc", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(videos) != 1 {
		t.Fatalf("expected 1 video, got %d", len(videos))
	}
}

func TestMockFacebookClient_FetchInsights_Success(t *testing.T) {
	expectedInsights := &kafkamodels.RawFacebookInsights{
		PageID: "page123",
	}

	mock := &mockFacebookClient{
		fetchInsightsFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) (*kafkamodels.RawFacebookInsights, error) {
			return expectedInsights, nil
		},
	}

	insights, err := mock.FetchInsights(context.Background(), "page123", "token_abc", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if insights == nil {
		t.Fatal("expected non-nil insights")
	}
	if insights.PageID != "page123" {
		t.Fatalf("expected PageID 'page123', got '%s'", insights.PageID)
	}
}

func TestMockFacebookClient_NilFunctions(t *testing.T) {
	mock := &mockFacebookClient{}

	posts, err := mock.FetchPostsSince(context.Background(), "", "", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if posts != nil {
		t.Fatal("expected nil posts")
	}

	videos, err := mock.FetchVideosSince(context.Background(), "", "", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if videos != nil {
		t.Fatal("expected nil videos")
	}

	insights, err := mock.FetchInsights(context.Background(), "", "", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if insights != nil {
		t.Fatal("expected nil insights")
	}
}

// ================== Integration-style Tests with All Mocks ==================

func TestProcessor_FullWorkflow_WithMocks(t *testing.T) {
	accountID := primitive.NewObjectID()

	// Mock MongoDB
	mockRepo := &mockMongoRepo{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:                 accountID,
				PlatformIdentifier: "fb_page_123",
				AccessToken:        "plain_token_for_testing",
				State:              "Added",
				ExtraData: map[string]interface{}{
					"workspace_id": "workspace_456",
					"name":         "My Facebook Page",
					"access_token": "fallback_token",
				},
			}, nil
		},
		UpdateStateFunc: func(ctx context.Context, id primitive.ObjectID, state string) error {
			return nil
		},
		UpdateAnalyticsTimestampFunc: func(ctx context.Context, id primitive.ObjectID, field string, timestamp time.Time) error {
			return nil
		},
	}

	// Mock Facebook Client
	mockFB := &mockFacebookClient{
		fetchPostsSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookPost, error) {
			return []kafkamodels.RawFacebookPost{
				{ID: "post1", Message: "Test post 1", CreatedTime: kafkamodels.FacebookTime{Time: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)}},
				{ID: "post2", Message: "Test post 2", CreatedTime: kafkamodels.FacebookTime{Time: time.Date(2024, 1, 14, 10, 0, 0, 0, time.UTC)}},
			}, nil
		},
		fetchVideosSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookVideo, error) {
			return []kafkamodels.RawFacebookVideo{
				{ID: "video1", Description: "Test video"},
			}, nil
		},
		fetchInsightsFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) (*kafkamodels.RawFacebookInsights, error) {
			return &kafkamodels.RawFacebookInsights{
				PageID: pageID,
			}, nil
		},
	}

	// Mock ClickHouse
	postsInserted := 0
	mockCH := &mockClickHouseSink{
		BulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			postsInserted = len(posts)
			return nil
		},
		BulkInsertMediaAssetsFunc: func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
			return nil
		},
		BulkInsertVideoInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
			return nil
		},
		BulkInsertReelsInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
			return nil
		},
		BulkInsertInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
			return nil
		},
	}

	// Mock Kafka Producer
	mockProducer := &mockKafkaProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return nil
		},
	}

	// Create Processor with all mocks
	p := &Processor{
		MongoRepo:      mockRepo,
		FacebookClient: mockFB,
		Sink:           mockCH,
		Producer:       mockProducer,
	}

	// Verify all dependencies are set
	if p.MongoRepo == nil {
		t.Fatal("MongoRepo should not be nil")
	}
	if p.FacebookClient == nil {
		t.Fatal("FacebookClient should not be nil")
	}
	if p.Sink == nil {
		t.Fatal("Sink should not be nil")
	}

	// Test fetching data through the mocks
	ctx := context.Background()

	// Test MongoDB FindByID
	account, err := p.MongoRepo.FindByID(ctx, accountID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if account.PlatformIdentifier != "fb_page_123" {
		t.Fatalf("unexpected PlatformIdentifier: %s", account.PlatformIdentifier)
	}

	// Test Facebook API calls
	posts, err := p.FacebookClient.FetchPostsSince(ctx, "fb_page_123", "token", time.Now().Add(-24*time.Hour), time.Now())
	if err != nil {
		t.Fatalf("FetchPostsSince failed: %v", err)
	}
	if len(posts) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(posts))
	}

	videos, err := p.FacebookClient.FetchVideosSince(ctx, "fb_page_123", "token", time.Now().Add(-24*time.Hour), time.Now())
	if err != nil {
		t.Fatalf("FetchVideosSince failed: %v", err)
	}
	if len(videos) != 1 {
		t.Fatalf("expected 1 video, got %d", len(videos))
	}

	insights, err := p.FacebookClient.FetchInsights(ctx, "fb_page_123", "token", time.Now().Add(-24*time.Hour), time.Now())
	if err != nil {
		t.Fatalf("FetchInsights failed: %v", err)
	}
	if insights == nil {
		t.Fatal("expected non-nil insights")
	}

	// Test ClickHouse insert
	err = p.Sink.BulkInsertPosts(ctx, []*clickhousemodels.FacebookPosts{
		{PageID: "page1", PostID: "post1"},
		{PageID: "page1", PostID: "post2"},
	})
	if err != nil {
		t.Fatalf("BulkInsertPosts failed: %v", err)
	}
	if postsInserted != 2 {
		t.Fatalf("expected 2 posts inserted, got %d", postsInserted)
	}

	// Test state updates
	err = p.MongoRepo.UpdateState(ctx, accountID, "Processed")
	if err != nil {
		t.Fatalf("UpdateState failed: %v", err)
	}
}

func TestProcessor_ErrorHandling_MongoFindByIDError(t *testing.T) {
	expectedErr := errors.New("MongoDB connection failed")
	mockRepo := &mockMongoRepo{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return nil, expectedErr
		},
	}

	p := &Processor{
		MongoRepo: mockRepo,
	}

	_, err := p.MongoRepo.FindByID(context.Background(), primitive.NewObjectID())
	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestProcessor_ErrorHandling_FacebookAPIError(t *testing.T) {
	expectedErr := errors.New("Facebook API rate limit exceeded")
	mockFB := &mockFacebookClient{
		fetchPostsSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookPost, error) {
			return nil, expectedErr
		},
	}

	p := &Processor{
		FacebookClient: mockFB,
	}

	_, err := p.FacebookClient.FetchPostsSince(context.Background(), "page", "token", time.Now(), time.Now())
	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestProcessor_ErrorHandling_ClickHouseInsertError(t *testing.T) {
	expectedErr := errors.New("ClickHouse insert failed")
	mockCH := &mockClickHouseSink{
		BulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			return expectedErr
		},
	}

	p := &Processor{
		Sink: mockCH,
	}

	err := p.Sink.BulkInsertPosts(context.Background(), nil)
	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestFacebookClientInterface_ImplementedByMock(t *testing.T) {
	var client FacebookClientInterface = &mockFacebookClient{}
	if client == nil {
		t.Fatal("expected non-nil interface")
	}
}

// ================== ProcessAccount Tests ==================
// Note: Full integration tests for ProcessAccount require ClickHouse connection
// because storeInClickHouse() creates a new sink internally.
// These tests focus on testing error paths before the ClickHouse connection is attempted.

func TestProcessor_ProcessAccount_InvalidAccountID(t *testing.T) {
	p := &Processor{}
	workOrder := WorkOrder{
		ID: "invalid-objectid",
	}

	err := p.ProcessAccount(context.Background(), workOrder)
	if err == nil {
		t.Fatal("expected error for invalid account ID")
	}
	if !strings.Contains(err.Error(), "invalid account ID") {
		t.Fatalf("expected 'invalid account ID' error, got: %v", err)
	}
}

func TestProcessor_ProcessAccount_AccountNotFound(t *testing.T) {
	mockRepo := &mockMongoRepo{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return nil, nil
		},
	}

	p := &Processor{
		MongoRepo: mockRepo,
	}

	accountID := primitive.NewObjectID()
	workOrder := WorkOrder{
		ID: accountID.Hex(),
	}

	err := p.ProcessAccount(context.Background(), workOrder)
	if err == nil {
		t.Fatal("expected error for account not found")
	}
	if !strings.Contains(err.Error(), "account not found") {
		t.Fatalf("expected 'account not found' error, got: %v", err)
	}
}

func TestProcessor_ProcessAccount_MongoDBError(t *testing.T) {
	expectedErr := errors.New("MongoDB connection failed")
	mockRepo := &mockMongoRepo{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return nil, expectedErr
		},
	}

	p := &Processor{
		MongoRepo: mockRepo,
	}

	accountID := primitive.NewObjectID()
	workOrder := WorkOrder{
		ID: accountID.Hex(),
	}

	err := p.ProcessAccount(context.Background(), workOrder)
	if err == nil {
		t.Fatal("expected error for MongoDB failure")
	}
	if !strings.Contains(err.Error(), "failed to fetch account from MongoDB") {
		t.Fatalf("expected 'failed to fetch account' error, got: %v", err)
	}
}

func TestProcessor_ProcessAccount_NoAccessToken(t *testing.T) {
	accountID := primitive.NewObjectID()
	mockRepo := &mockMongoRepo{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:          accountID,
				AccessToken: "",
				ExtraData:   map[string]interface{}{},
			}, nil
		},
		UpdateStateFunc: func(ctx context.Context, id primitive.ObjectID, state string) error {
			return nil
		},
	}

	p := &Processor{
		MongoRepo: mockRepo,
		Config:    &config.Config{DecryptionKey: "test-key"},
	}

	workOrder := WorkOrder{
		ID: accountID.Hex(),
	}

	err := p.ProcessAccount(context.Background(), workOrder)
	if err == nil {
		t.Fatal("expected error for no access token")
	}
	if !strings.Contains(err.Error(), "no valid access token") {
		t.Fatalf("expected 'no valid access token' error, got: %v", err)
	}
}

// ================== Full ProcessAccount Integration Tests ==================

func TestProcessor_ProcessAccount_FullWorkflow_Success(t *testing.T) {
	accountID := primitive.NewObjectID()
	stateUpdates := []string{}
	postsInserted := 0
	var mu sync.Mutex
	var postsSinceStart, postsSinceEnd time.Time
	var videosSinceStart, videosSinceEnd time.Time
	var insightsSinceStart, insightsSinceEnd time.Time

	mockRepo := &mockMongoRepo{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:                 accountID,
				AccessToken:        "",
				PlatformIdentifier: "fb_page_123",
				State:              "Added",
				ExtraData: map[string]interface{}{
					"access_token": "test_token_value",
					"workspace_id": "workspace_123",
					"name":         "Test Page",
				},
			}, nil
		},
		UpdateStateFunc: func(ctx context.Context, id primitive.ObjectID, state string) error {
			stateUpdates = append(stateUpdates, state)
			return nil
		},
		UpdateAnalyticsTimestampFunc: func(ctx context.Context, id primitive.ObjectID, field string, timestamp time.Time) error {
			return nil
		},
	}

	mockFB := &mockFacebookClient{
		fetchPostsSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookPost, error) {
			mu.Lock()
			postsSinceStart, postsSinceEnd = since, until
			mu.Unlock()
			return []kafkamodels.RawFacebookPost{
				{ID: "123_456", Message: "Test post", CreatedTime: kafkamodels.FacebookTime{Time: time.Now()}},
			}, nil
		},
		fetchVideosSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookVideo, error) {
			mu.Lock()
			videosSinceStart, videosSinceEnd = since, until
			mu.Unlock()
			return []kafkamodels.RawFacebookVideo{}, nil
		},
		fetchInsightsFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) (*kafkamodels.RawFacebookInsights, error) {
			mu.Lock()
			insightsSinceStart, insightsSinceEnd = since, until
			mu.Unlock()
			return nil, nil
		},
	}

	mockCH := &mockClickHouseSink{
		BulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			postsInserted = len(posts)
			return nil
		},
		BulkInsertMediaAssetsFunc: func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
			return nil
		},
		BulkInsertVideoInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
			return nil
		},
		BulkInsertReelsInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
			return nil
		},
		BulkInsertInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
			return nil
		},
	}

	p := &Processor{
		MongoRepo:      mockRepo,
		FacebookClient: mockFB,
		Sink:           mockCH,
		Parser:         parsing.NewFacebookParser(),
		Logger:         &logger.Logger{},
		Config:         &config.Config{DecryptionKey: "test-key"},
	}

	workOrder := WorkOrder{
		ID:          accountID.Hex(),
		AccountID:   "fb_page_123",
		WorkspaceID: "workspace_123",
		StartDate:   "2025-01-01",
		EndDate:     "2025-01-31",
	}

	err := p.ProcessAccount(context.Background(), workOrder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stateUpdates) < 2 {
		t.Fatalf("expected at least 2 state updates, got %d", len(stateUpdates))
	}
	if stateUpdates[0] != "Processing" {
		t.Fatalf("expected first state 'Processing', got '%s'", stateUpdates[0])
	}
	if stateUpdates[len(stateUpdates)-1] != "Processed" {
		t.Fatalf("expected last state 'Processed', got '%s'", stateUpdates[len(stateUpdates)-1])
	}
	if postsInserted != 1 {
		t.Fatalf("expected 1 post inserted, got %d", postsInserted)
	}

	expectedStart := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC)

	mu.Lock()
	defer mu.Unlock()
	if !postsSinceStart.Equal(expectedStart) || !postsSinceEnd.Equal(expectedEnd) {
		t.Fatalf("expected post range %v - %v, got %v - %v", expectedStart, expectedEnd, postsSinceStart, postsSinceEnd)
	}
	if !videosSinceStart.Equal(expectedStart) || !videosSinceEnd.Equal(expectedEnd) {
		t.Fatalf("expected video range %v - %v, got %v - %v", expectedStart, expectedEnd, videosSinceStart, videosSinceEnd)
	}
	if !insightsSinceStart.Equal(expectedStart) || !insightsSinceEnd.Equal(expectedEnd) {
		t.Fatalf("expected insights range %v - %v, got %v - %v", expectedStart, expectedEnd, insightsSinceStart, insightsSinceEnd)
	}
}

func TestProcessor_ProcessAccount_ClickHouseInsertError(t *testing.T) {
	accountID := primitive.NewObjectID()

	mockRepo := &mockMongoRepo{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:                 accountID,
				AccessToken:        "",
				PlatformIdentifier: "fb_page_123",
				State:              "Processing",
				ExtraData: map[string]interface{}{
					"access_token": "test_token",
					"workspace_id": "workspace_123",
					"name":         "Test Page",
				},
			}, nil
		},
		UpdateStateFunc: func(ctx context.Context, id primitive.ObjectID, state string) error {
			return nil
		},
	}

	mockFB := &mockFacebookClient{
		fetchPostsSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookPost, error) {
			return []kafkamodels.RawFacebookPost{
				{ID: "123_456", Message: "Test post", CreatedTime: kafkamodels.FacebookTime{Time: time.Now()}},
			}, nil
		},
		fetchVideosSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookVideo, error) {
			return []kafkamodels.RawFacebookVideo{}, nil
		},
		fetchInsightsFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) (*kafkamodels.RawFacebookInsights, error) {
			return nil, nil
		},
	}

	expectedErr := errors.New("ClickHouse connection failed")
	mockCH := &mockClickHouseSink{
		BulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			return expectedErr
		},
	}

	p := &Processor{
		MongoRepo:      mockRepo,
		FacebookClient: mockFB,
		Sink:           mockCH,
		Parser:         parsing.NewFacebookParser(),
		Logger:         &logger.Logger{},
		Config:         &config.Config{DecryptionKey: "test-key"},
	}

	workOrder := WorkOrder{
		ID:          accountID.Hex(),
		AccountID:   "fb_page_123",
		WorkspaceID: "workspace_123",
	}

	err := p.ProcessAccount(context.Background(), workOrder)
	if err == nil {
		t.Fatal("expected error for ClickHouse failure")
	}
	if !strings.Contains(err.Error(), "failed to store data") {
		t.Fatalf("expected 'failed to store data' error, got: %v", err)
	}
}

func TestProcessor_ProcessAccount_WithVideosAndInsights(t *testing.T) {
	accountID := primitive.NewObjectID()

	mockRepo := &mockMongoRepo{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:                 accountID,
				AccessToken:        "",
				PlatformIdentifier: "fb_page_123",
				FacebookID:         "fb_page_123",
				State:              "Processing",
				ExtraData: map[string]interface{}{
					"access_token": "test_token",
					"workspace_id": "workspace_123",
					"name":         "Test Page",
				},
			}, nil
		},
		UpdateStateFunc: func(ctx context.Context, id primitive.ObjectID, state string) error {
			return nil
		},
		UpdateAnalyticsTimestampFunc: func(ctx context.Context, id primitive.ObjectID, field string, timestamp time.Time) error {
			return nil
		},
	}

	mockFB := &mockFacebookClient{
		fetchPostsSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookPost, error) {
			return []kafkamodels.RawFacebookPost{}, nil
		},
		fetchVideosSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookVideo, error) {
			return []kafkamodels.RawFacebookVideo{
				{ID: "video_1", Description: "Test video"},
			}, nil
		},
		fetchInsightsFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) (*kafkamodels.RawFacebookInsights, error) {
			return &kafkamodels.RawFacebookInsights{
				PageID: pageID,
			}, nil
		},
	}

	mockCH := &mockClickHouseSink{
		BulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			return nil
		},
		BulkInsertMediaAssetsFunc: func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
			return nil
		},
		BulkInsertVideoInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
			return nil
		},
		BulkInsertReelsInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
			return nil
		},
		BulkInsertInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
			return nil
		},
	}

	p := &Processor{
		MongoRepo:      mockRepo,
		FacebookClient: mockFB,
		Sink:           mockCH,
		Parser:         parsing.NewFacebookParser(),
		Logger:         &logger.Logger{},
		Config:         &config.Config{DecryptionKey: "test-key"},
	}

	workOrder := WorkOrder{
		ID:          accountID.Hex(),
		AccountID:   "fb_page_123",
		WorkspaceID: "workspace_123",
	}

	err := p.ProcessAccount(context.Background(), workOrder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProcessor_ProcessAccount_FetchDataError(t *testing.T) {
	accountID := primitive.NewObjectID()
	stateFailed := false

	mockRepo := &mockMongoRepo{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:                 accountID,
				AccessToken:        "",
				PlatformIdentifier: "fb_page_123",
				State:              "Processing",
				ExtraData: map[string]interface{}{
					"access_token": "test_token",
					"workspace_id": "workspace_123",
					"name":         "Test Page",
				},
			}, nil
		},
		UpdateStateFunc: func(ctx context.Context, id primitive.ObjectID, state string) error {
			if state == "Failed" {
				stateFailed = true
			}
			return nil
		},
	}

	// Facebook client returns errors for all fetch operations
	mockFB := &mockFacebookClient{
		fetchPostsSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookPost, error) {
			return nil, errors.New("API error")
		},
		fetchVideosSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookVideo, error) {
			return nil, errors.New("API error")
		},
		fetchInsightsFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) (*kafkamodels.RawFacebookInsights, error) {
			return nil, errors.New("API error")
		},
	}

	mockCH := &mockClickHouseSink{}

	p := &Processor{
		MongoRepo:      mockRepo,
		FacebookClient: mockFB,
		Sink:           mockCH,
		Parser:         parsing.NewFacebookParser(),
		Logger:         &logger.Logger{},
		Config:         &config.Config{DecryptionKey: "test-key"},
	}

	workOrder := WorkOrder{
		ID:          accountID.Hex(),
		AccountID:   "fb_page_123",
		WorkspaceID: "workspace_123",
	}

	// The processor logs errors but doesn't fail the whole process if fetch fails
	err := p.ProcessAccount(context.Background(), workOrder)
	// Fetch errors are logged but don't fail the process, so we expect success
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = stateFailed // State might be updated to Failed or Processed depending on implementation
}

func TestProcessor_StoreInClickHouse_NilData(t *testing.T) {
	p := &Processor{}

	err := p.storeInClickHouse(context.Background(), WorkOrder{}, nil)
	if err == nil {
		t.Fatal("expected error for nil data")
	}
	if !strings.Contains(err.Error(), "parsed data is nil") {
		t.Fatalf("expected 'parsed data is nil' error, got: %v", err)
	}
}

func TestMockClickHouseSink_AllMethods(t *testing.T) {
	mock := &mockClickHouseSink{}

	// Test ConvertFacebookPost
	post := &kafkamodels.ParsedFacebookPost{PageID: "page1", PostID: "post1"}
	converted := mock.ConvertFacebookPost(post)
	if converted.PageID != "page1" {
		t.Fatalf("expected PageID 'page1', got '%s'", converted.PageID)
	}

	// Test ConvertFacebookMediaAssets
	asset := &kafkamodels.ParsedFacebookMediaAsset{PostID: "post1"}
	convertedAsset := mock.ConvertFacebookMediaAssets(asset)
	if convertedAsset.PostID != "post1" {
		t.Fatalf("expected PostID 'post1', got '%s'", convertedAsset.PostID)
	}

	// Test ConvertFacebookVideoInsights
	video := &kafkamodels.ParsedFacebookVideoInsights{PostID: "video1"}
	convertedVideo := mock.ConvertFacebookVideoInsights(video)
	if convertedVideo.PostID != "video1" {
		t.Fatalf("expected PostID 'video1', got '%s'", convertedVideo.PostID)
	}

	// Test ConvertFacebookReelsInsights
	reel := &kafkamodels.ParsedFacebookReelsInsights{PostID: "reel1"}
	convertedReel := mock.ConvertFacebookReelsInsights(reel)
	if convertedReel.PostID != "reel1" {
		t.Fatalf("expected PostID 'reel1', got '%s'", convertedReel.PostID)
	}

	// Test ConvertFacebookInsights
	insights := &kafkamodels.ParsedFacebookInsights{PageID: "page1"}
	convertedInsights := mock.ConvertFacebookInsights(insights)
	if convertedInsights.PageID != "page1" {
		t.Fatalf("expected PageID 'page1', got '%s'", convertedInsights.PageID)
	}
}

// ================== storeInClickHouse Tests ==================

func TestProcessor_StoreInClickHouse_WithAllDataTypes(t *testing.T) {
	postsInserted := 0
	mediaInserted := 0
	videosInserted := 0
	reelsInserted := 0
	insightsInserted := 0

	mockCH := &mockClickHouseSink{
		BulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			postsInserted = len(posts)
			return nil
		},
		BulkInsertMediaAssetsFunc: func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
			mediaInserted = len(assets)
			return nil
		},
		BulkInsertVideoInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
			videosInserted = len(insights)
			return nil
		},
		BulkInsertReelsInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
			reelsInserted = len(insights)
			return nil
		},
		BulkInsertInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
			insightsInserted = len(insights)
			return nil
		},
	}

	p := &Processor{
		Sink: mockCH,
	}

	data := &ParsedData{
		Posts: []kafkamodels.ParsedFacebookPost{
			{PageID: "page1", PostID: "post1"},
			{PageID: "page1", PostID: "post2"},
		},
		MediaAssets: []kafkamodels.ParsedFacebookMediaAsset{
			{PostID: "post1", Link: "http://example.com/image1.jpg"},
		},
		VideoInsights: []kafkamodels.ParsedFacebookVideoInsights{
			{PostID: "video1"},
		},
		ReelsInsights: []kafkamodels.ParsedFacebookReelsInsights{
			{PostID: "reel1"},
		},
		Insights: []*kafkamodels.ParsedFacebookInsights{
			{PageID: "page1"},
			{PageID: "page1"},
		},
	}

	err := p.storeInClickHouse(context.Background(), WorkOrder{}, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if postsInserted != 2 {
		t.Fatalf("expected 2 posts inserted, got %d", postsInserted)
	}
	if mediaInserted != 1 {
		t.Fatalf("expected 1 media asset inserted, got %d", mediaInserted)
	}
	if videosInserted != 1 {
		t.Fatalf("expected 1 video insight inserted, got %d", videosInserted)
	}
	if reelsInserted != 1 {
		t.Fatalf("expected 1 reel insight inserted, got %d", reelsInserted)
	}
	if insightsInserted != 2 {
		t.Fatalf("expected 2 page insights inserted, got %d", insightsInserted)
	}
}

func TestProcessor_StoreInClickHouse_MediaAssetsError(t *testing.T) {
	expectedErr := errors.New("media assets insert failed")
	mockCH := &mockClickHouseSink{
		BulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			return nil
		},
		BulkInsertMediaAssetsFunc: func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
			return expectedErr
		},
	}

	p := &Processor{
		Sink: mockCH,
	}

	data := &ParsedData{
		Posts: []kafkamodels.ParsedFacebookPost{
			{PageID: "page1", PostID: "post1"},
		},
		MediaAssets: []kafkamodels.ParsedFacebookMediaAsset{
			{PostID: "post1"},
		},
	}

	err := p.storeInClickHouse(context.Background(), WorkOrder{}, data)
	if err == nil {
		t.Fatal("expected error for media assets insert failure")
	}
	if !strings.Contains(err.Error(), "failed to insert media assets") {
		t.Fatalf("expected 'failed to insert media assets' error, got: %v", err)
	}
}

func TestProcessor_StoreInClickHouse_VideoInsightsError(t *testing.T) {
	expectedErr := errors.New("video insights insert failed")
	mockCH := &mockClickHouseSink{
		BulkInsertVideoInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
			return expectedErr
		},
	}

	p := &Processor{
		Sink: mockCH,
	}

	data := &ParsedData{
		VideoInsights: []kafkamodels.ParsedFacebookVideoInsights{
			{PostID: "video1"},
		},
	}

	err := p.storeInClickHouse(context.Background(), WorkOrder{}, data)
	if err == nil {
		t.Fatal("expected error for video insights insert failure")
	}
	if !strings.Contains(err.Error(), "failed to insert video insights") {
		t.Fatalf("expected 'failed to insert video insights' error, got: %v", err)
	}
}

func TestProcessor_StoreInClickHouse_ReelsInsightsError(t *testing.T) {
	expectedErr := errors.New("reels insights insert failed")
	mockCH := &mockClickHouseSink{
		BulkInsertReelsInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
			return expectedErr
		},
	}

	p := &Processor{
		Sink: mockCH,
	}

	data := &ParsedData{
		ReelsInsights: []kafkamodels.ParsedFacebookReelsInsights{
			{PostID: "reel1"},
		},
	}

	err := p.storeInClickHouse(context.Background(), WorkOrder{}, data)
	if err == nil {
		t.Fatal("expected error for reels insights insert failure")
	}
	if !strings.Contains(err.Error(), "failed to insert reels insights") {
		t.Fatalf("expected 'failed to insert reels insights' error, got: %v", err)
	}
}

func TestProcessor_StoreInClickHouse_PageInsightsError(t *testing.T) {
	expectedErr := errors.New("page insights insert failed")
	mockCH := &mockClickHouseSink{
		BulkInsertInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
			return expectedErr
		},
	}

	p := &Processor{
		Sink: mockCH,
	}

	data := &ParsedData{
		Insights: []*kafkamodels.ParsedFacebookInsights{
			{PageID: "page1"},
		},
	}

	err := p.storeInClickHouse(context.Background(), WorkOrder{}, data)
	if err == nil {
		t.Fatal("expected error for page insights insert failure")
	}
	if !strings.Contains(err.Error(), "failed to insert page insights") {
		t.Fatalf("expected 'failed to insert page insights' error, got: %v", err)
	}
}

func TestProcessor_StoreInClickHouse_EmptyData(t *testing.T) {
	mockCH := &mockClickHouseSink{}

	p := &Processor{
		Sink: mockCH,
	}

	data := &ParsedData{}

	err := p.storeInClickHouse(context.Background(), WorkOrder{}, data)
	if err != nil {
		t.Fatalf("unexpected error for empty data: %v", err)
	}
}

func TestProcessor_StoreInClickHouse_NilConversions(t *testing.T) {
	postsInserted := 0

	// Mock sink that returns nil for conversions but tracks bulk inserts
	mockCH := &mockClickHouseSink{
		BulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			postsInserted = len(posts)
			return nil
		},
		convertPostFunc: func(post *kafkamodels.ParsedFacebookPost) *clickhousemodels.FacebookPosts {
			return nil
		},
	}

	p := &Processor{
		Sink: mockCH,
	}

	data := &ParsedData{
		Posts: []kafkamodels.ParsedFacebookPost{
			{PageID: "page1", PostID: "post1"},
		},
	}

	err := p.storeInClickHouse(context.Background(), WorkOrder{}, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No posts should be inserted because conversion returned nil
	if postsInserted != 0 {
		t.Fatalf("expected 0 posts inserted (nil conversion), got %d", postsInserted)
	}
}

// ================== Notification Tests ==================

func TestProcessor_SendPusherNotification_NilClient(t *testing.T) {
	p := &Processor{
		PusherClient: nil,
	}

	account := &mongomodels.SocialIntegration{
		FacebookID: "fb123",
	}

	// Should not panic with nil client
	p.sendPusherNotification(account, "workspace123", mongomodels.StateAdded)
}

func TestProcessor_SendEmailNotification_NilNotifier(t *testing.T) {
	p := &Processor{
		Notifier: nil,
	}

	// Should not panic with nil notifier
	p.sendEmailNotification("user123", "workspace123", "fb123", "Test Page")
}

func TestProcessor_SendToDeadLetterQueue(t *testing.T) {
	p := &Processor{}

	workOrder := WorkOrder{
		ID:        "test123",
		AccountID: "fb123",
	}

	// Should not panic - currently a no-op
	p.SendToDeadLetterQueue(workOrder, errors.New("test error"))
}

// ================== ProcessAccount Edge Cases ==================

func TestProcessor_ProcessAccount_WithPusherAndNotifier(t *testing.T) {
	accountID := primitive.NewObjectID()

	mockRepo := &mockMongoRepo{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:                 accountID,
				AccessToken:        "",
				PlatformIdentifier: "fb_page_123",
				FacebookID:         "fb_page_123",
				State:              "Added", // Trigger email notification
				ExtraData: map[string]interface{}{
					"access_token": "test_token",
					"workspace_id": "workspace_123",
					"name":         "Test Page",
				},
			}, nil
		},
		UpdateStateFunc: func(ctx context.Context, id primitive.ObjectID, state string) error {
			return nil
		},
		UpdateAnalyticsTimestampFunc: func(ctx context.Context, id primitive.ObjectID, field string, timestamp time.Time) error {
			return nil
		},
	}

	mockFB := &mockFacebookClient{
		fetchPostsSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookPost, error) {
			return []kafkamodels.RawFacebookPost{
				{
					ID:           "post_1",
					Message:      "test post",
					StatusType:   "mobile_status_update",
					PermalinkURL: "https://example.com/post_1",
					CreatedTime:   kafkamodels.FacebookTime{Time: time.Date(2026, 4, 22, 0, 0, 0, 0, time.UTC)},
					UpdatedTime:   kafkamodels.FacebookTime{Time: time.Date(2026, 4, 22, 0, 0, 0, 0, time.UTC)},
				},
			}, nil
		},
		fetchVideosSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookVideo, error) {
			return []kafkamodels.RawFacebookVideo{
				{
					ID:          "video_1",
					PostID:      "video_1",
					PermalinkURL: "https://example.com/video_1",
					CreatedTime:  kafkamodels.FacebookTime{Time: time.Date(2026, 4, 22, 0, 0, 0, 0, time.UTC)},
					UpdatedTime:  kafkamodels.FacebookTime{Time: time.Date(2026, 4, 22, 0, 0, 0, 0, time.UTC)},
				},
			}, nil
		},
		fetchInsightsFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) (*kafkamodels.RawFacebookInsights, error) {
			return &kafkamodels.RawFacebookInsights{
				PageID: pageID,
			}, nil
		},
	}

	mockCH := &mockClickHouseSink{
		BulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			return nil
		},
		BulkInsertMediaAssetsFunc: func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
			return nil
		},
		BulkInsertVideoInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
			return nil
		},
		BulkInsertReelsInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
			return nil
		},
		BulkInsertInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
			return nil
		},
	}

	p := &Processor{
		MongoRepo:      mockRepo,
		FacebookClient: mockFB,
		Sink:           mockCH,
		Parser:         parsing.NewFacebookParser(),
		Logger:         &logger.Logger{},
		Config:         &config.Config{DecryptionKey: "test-key"},
		PusherClient:   nil, // Will trigger early return in sendPusherNotification
		Notifier:       nil, // Will trigger early return in sendEmailNotification
	}

	workOrder := WorkOrder{
		ID:          accountID.Hex(),
		AccountID:   "fb_page_123",
		WorkspaceID: "workspace_123",
	}

	err := p.ProcessAccount(context.Background(), workOrder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProcessor_ProcessAccount_WithMultiplePosts(t *testing.T) {
	accountID := primitive.NewObjectID()
	postsInserted := 0

	mockRepo := &mockMongoRepo{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:                 accountID,
				AccessToken:        "",
				PlatformIdentifier: "fb_page_123",
				State:              "Processing",
				ExtraData: map[string]interface{}{
					"access_token": "test_token",
					"workspace_id": "workspace_123",
					"name":         "Test Page",
				},
			}, nil
		},
		UpdateStateFunc: func(ctx context.Context, id primitive.ObjectID, state string) error {
			return nil
		},
		UpdateAnalyticsTimestampFunc: func(ctx context.Context, id primitive.ObjectID, field string, timestamp time.Time) error {
			return nil
		},
	}

	mockFB := &mockFacebookClient{
		fetchPostsSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookPost, error) {
			return []kafkamodels.RawFacebookPost{
				{ID: "123_1", Message: "Post 1", CreatedTime: kafkamodels.FacebookTime{Time: time.Now()}},
				{ID: "123_2", Message: "Post 2", CreatedTime: kafkamodels.FacebookTime{Time: time.Now()}},
				{ID: "123_3", Message: "Post 3", CreatedTime: kafkamodels.FacebookTime{Time: time.Now()}},
			}, nil
		},
		fetchVideosSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookVideo, error) {
			return []kafkamodels.RawFacebookVideo{}, nil
		},
		fetchInsightsFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) (*kafkamodels.RawFacebookInsights, error) {
			return nil, nil
		},
	}

	mockCH := &mockClickHouseSink{
		BulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			postsInserted = len(posts)
			return nil
		},
		BulkInsertMediaAssetsFunc: func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
			return nil
		},
		BulkInsertVideoInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
			return nil
		},
		BulkInsertReelsInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
			return nil
		},
		BulkInsertInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
			return nil
		},
	}

	p := &Processor{
		MongoRepo:      mockRepo,
		FacebookClient: mockFB,
		Sink:           mockCH,
		Parser:         parsing.NewFacebookParser(),
		Logger:         &logger.Logger{},
		Config:         &config.Config{DecryptionKey: "test-key"},
	}

	workOrder := WorkOrder{
		ID:          accountID.Hex(),
		AccountID:   "fb_page_123",
		WorkspaceID: "workspace_123",
	}

	err := p.ProcessAccount(context.Background(), workOrder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if postsInserted != 3 {
		t.Fatalf("expected 3 posts inserted, got %d", postsInserted)
	}
}

// ================== Parse All Data Tests ==================

func TestProcessor_ParseAllData_WithVideos(t *testing.T) {
	p := &Processor{
		Parser: parsing.NewFacebookParser(),
		Logger: &logger.Logger{},
	}

	account := &mongomodels.SocialIntegration{
		FacebookID: "page123",
		ExtraData: map[string]interface{}{
			"name":         "Test Page",
			"workspace_id": "workspace123",
		},
	}

	workOrder := WorkOrder{
		WorkspaceID: "workspace123",
	}

	posts := []kafkamodels.RawFacebookPost{
		{ID: "page123_post1", Message: "Test", CreatedTime: kafkamodels.FacebookTime{Time: time.Now()}},
	}
	videos := []kafkamodels.RawFacebookVideo{
		{ID: "video1", PostID: "post_video1"},
	}

	capture := func(string, error, map[string]interface{}) {}
	parsed, err := p.parseAllData(workOrder, account, posts, videos, nil, capture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed == nil {
		t.Fatal("expected parsed data, got nil")
	}
	if len(parsed.Posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(parsed.Posts))
	}
}

func TestProcessor_ParseAllData_WithInsights(t *testing.T) {
	p := &Processor{
		Parser: parsing.NewFacebookParser(),
		Logger: &logger.Logger{},
	}

	account := &mongomodels.SocialIntegration{
		FacebookID: "page123",
		ExtraData: map[string]interface{}{
			"name":         "Test Page",
			"workspace_id": "workspace123",
		},
	}

	workOrder := WorkOrder{
		WorkspaceID: "workspace123",
	}

	posts := []kafkamodels.RawFacebookPost{
		{ID: "page123_post1", Message: "Test", CreatedTime: kafkamodels.FacebookTime{Time: time.Now()}},
	}
	insights := &kafkamodels.RawFacebookInsights{
		PageID: "page123",
	}

	capture := func(string, error, map[string]interface{}) {}
	parsed, err := p.parseAllData(workOrder, account, posts, nil, insights, capture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed == nil {
		t.Fatal("expected parsed data, got nil")
	}
	if len(parsed.Posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(parsed.Posts))
	}
}

func TestProcessor_ParseAllData_EmptyData(t *testing.T) {
	p := &Processor{
		Parser: parsing.NewFacebookParser(),
		Logger: &logger.Logger{},
	}

	account := &mongomodels.SocialIntegration{
		FacebookID: "page123",
		ExtraData: map[string]interface{}{
			"name":         "Test Page",
			"workspace_id": "workspace123",
		},
	}

	workOrder := WorkOrder{
		WorkspaceID: "workspace123",
	}

	capture := func(string, error, map[string]interface{}) {}
	parsed, err := p.parseAllData(workOrder, account, nil, nil, nil, capture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed == nil {
		t.Fatal("expected parsed data, got nil")
	}
	if len(parsed.Posts) != 0 {
		t.Fatalf("expected 0 posts, got %d", len(parsed.Posts))
	}
}

func TestProcessor_ParseAllData_WithRegularVideo(t *testing.T) {
	p := &Processor{
		Parser: parsing.NewFacebookParser(),
		Logger: &logger.Logger{},
	}

	account := &mongomodels.SocialIntegration{
		FacebookID: "page123",
		ExtraData: map[string]interface{}{
			"name":         "Test Page",
			"workspace_id": "workspace123",
		},
	}

	workOrder := WorkOrder{
		WorkspaceID: "workspace123",
	}

	posts := []kafkamodels.RawFacebookPost{
		{
			ID:         "page123_post_video1",
			StatusType: "added_video",
		},
	}
	videos := []kafkamodels.RawFacebookVideo{
		{
			ID:     "video1",
			PostID: "post_video1",
		},
	}

	capture := func(string, error, map[string]interface{}) {}
	parsed, err := p.parseAllData(workOrder, account, posts, videos, nil, capture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed == nil {
		t.Fatal("expected parsed data, got nil")
	}
	// Video without BlueReelsPlayCount should be added to VideoInsights
	if len(parsed.VideoInsights) != 1 {
		t.Fatalf("expected 1 video insight, got %d", len(parsed.VideoInsights))
	}
}

func TestProcessor_ParseAllData_SkipsVideoMappedToImagePost(t *testing.T) {
	p := &Processor{
		Parser: parsing.NewFacebookParser(),
		Logger: &logger.Logger{},
	}

	account := &mongomodels.SocialIntegration{
		FacebookID: "page123",
		ExtraData: map[string]interface{}{
			"name":         "Test Page",
			"workspace_id": "workspace123",
		},
	}

	workOrder := WorkOrder{
		ID:          "account123",
		WorkspaceID: "workspace123",
	}

	posts := []kafkamodels.RawFacebookPost{
		{
			ID:         "page123_post_image1",
			StatusType: "added_photos",
		},
	}
	videos := []kafkamodels.RawFacebookVideo{
		{
			ID:     "video1",
			PostID: "post_image1",
		},
	}

	capture := func(string, error, map[string]interface{}) {}
	parsed, err := p.parseAllData(workOrder, account, posts, videos, nil, capture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed == nil {
		t.Fatal("expected parsed data, got nil")
	}
	if len(parsed.VideoInsights) != 0 {
		t.Fatalf("expected invalid video insight to be skipped, got %d", len(parsed.VideoInsights))
	}
}

// ================== Notification Tests with Mocks ==================

func TestProcessor_SendPusherNotification_WithMock(t *testing.T) {
	mockPusher := &mockPusherClient{}

	p := &Processor{
		PusherClient: mockPusher,
	}

	account := &mongomodels.SocialIntegration{
		FacebookID: "fb123",
	}

	p.sendPusherNotification(account, "workspace123", mongomodels.StateAdded)

	if !mockPusher.triggerCalled {
		t.Fatal("expected Trigger to be called")
	}
	if mockPusher.lastChannel != "fb-analytics-channel-workspace123-fb123" {
		t.Fatalf("unexpected channel: %s", mockPusher.lastChannel)
	}
	if mockPusher.lastEvent != "syncing-workspace123-fb123" {
		t.Fatalf("unexpected event: %s", mockPusher.lastEvent)
	}
}

func TestProcessor_SendEmailNotification_WithMock(t *testing.T) {
	mockNotif := &mockNotifier{}

	p := &Processor{
		Notifier: mockNotif,
	}

	p.sendEmailNotification("user123", "workspace123", "fb123", "Test Page")

	if !mockNotif.sendCalled {
		t.Fatal("expected SendAnalyticsNotification to be called")
	}
	if mockNotif.lastUserID != "user123" {
		t.Fatalf("unexpected userID: %s", mockNotif.lastUserID)
	}
	if mockNotif.lastWorkspaceID != "workspace123" {
		t.Fatalf("unexpected workspaceID: %s", mockNotif.lastWorkspaceID)
	}
	if mockNotif.lastPlatform != "facebook" {
		t.Fatalf("unexpected platform: %s", mockNotif.lastPlatform)
	}
	if mockNotif.lastAccountID != "fb123" {
		t.Fatalf("unexpected accountID: %s", mockNotif.lastAccountID)
	}
	if mockNotif.lastAccountName != "Test Page" {
		t.Fatalf("unexpected accountName: %s", mockNotif.lastAccountName)
	}
}

// ================== ProcessAccount Full Flow Tests ==================

func TestProcessor_ProcessAccount_WithNotifications(t *testing.T) {
	accountID := primitive.NewObjectID()
	mockPusher := &mockPusherClient{}
	mockNotifier := &mockNotifier{}

	mockRepo := &mockMongoRepo{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:                 accountID,
				AccessToken:        "",
				PlatformIdentifier: "fb_page_123",
				FacebookID:         "fb_page_123",
				State:              "Added",
				ExtraData: map[string]interface{}{
					"access_token": "test_token",
					"workspace_id": "workspace_123",
					"name":         "Test Page",
					"user_id":      "user_123",
				},
			}, nil
		},
		UpdateStateFunc: func(ctx context.Context, id primitive.ObjectID, state string) error {
			return nil
		},
		UpdateAnalyticsTimestampFunc: func(ctx context.Context, id primitive.ObjectID, field string, timestamp time.Time) error {
			return nil
		},
	}

	mockFB := &mockFacebookClient{
		fetchPostsSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookPost, error) {
			return []kafkamodels.RawFacebookPost{}, nil
		},
		fetchVideosSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookVideo, error) {
			return []kafkamodels.RawFacebookVideo{}, nil
		},
		fetchInsightsFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) (*kafkamodels.RawFacebookInsights, error) {
			return nil, nil
		},
	}

	mockCH := &mockClickHouseSink{}

	p := &Processor{
		MongoRepo:      mockRepo,
		FacebookClient: mockFB,
		Sink:           mockCH,
		Parser:         parsing.NewFacebookParser(),
		Logger:         &logger.Logger{},
		Config:         &config.Config{DecryptionKey: "test-key"},
		PusherClient:   mockPusher,
		Notifier:       mockNotifier,
	}

	workOrder := WorkOrder{
		ID:          accountID.Hex(),
		AccountID:   "fb_page_123",
		WorkspaceID: "workspace_123",
	}

	err := p.ProcessAccount(context.Background(), workOrder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProcessor_ProcessAccount_SkipsNotificationsWhenNoDataFetched(t *testing.T) {
	accountID := primitive.NewObjectID()
	mockPusher := &mockPusherClient{}
	mockNotifier := &mockNotifier{}

	mockRepo := &mockMongoRepo{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:                 accountID,
				AccessToken:        "",
				PlatformIdentifier: "fb_page_123",
				FacebookID:         "fb_page_123",
				State:              "Added",
				ExtraData: map[string]interface{}{
					"access_token": "test_token",
					"workspace_id": "workspace_123",
					"name":         "Test Page",
					"user_id":      "user_123",
				},
			}, nil
		},
		UpdateStateFunc: func(ctx context.Context, id primitive.ObjectID, state string) error {
			return nil
		},
		UpdateAnalyticsTimestampFunc: func(ctx context.Context, id primitive.ObjectID, field string, timestamp time.Time) error {
			return nil
		},
	}

	mockFB := &mockFacebookClient{
		fetchPostsSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookPost, error) {
			return []kafkamodels.RawFacebookPost{}, nil
		},
		fetchVideosSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookVideo, error) {
			return []kafkamodels.RawFacebookVideo{}, nil
		},
		fetchInsightsFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) (*kafkamodels.RawFacebookInsights, error) {
			return nil, nil
		},
	}

	mockCH := &mockClickHouseSink{}

	p := &Processor{
		MongoRepo:      mockRepo,
		FacebookClient: mockFB,
		Sink:           mockCH,
		Parser:         parsing.NewFacebookParser(),
		Logger:         &logger.Logger{},
		Config:         &config.Config{DecryptionKey: "test-key"},
		PusherClient:   mockPusher,
		Notifier:       mockNotifier,
	}

	workOrder := WorkOrder{
		ID:          accountID.Hex(),
		AccountID:   "fb_page_123",
		WorkspaceID: "workspace_123",
	}

	err := p.ProcessAccount(context.Background(), workOrder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mockPusher.triggerCalled {
		t.Fatal("expected Pusher notification to be skipped when no Facebook data was fetched")
	}
	if mockNotifier.sendCalled {
		t.Fatal("expected email notification to be skipped when no Facebook data was fetched")
	}
}

func TestProcessor_ProcessAccount_DecryptionKeyUsed(t *testing.T) {
	accountID := primitive.NewObjectID()

	mockRepo := &mockMongoRepo{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:                 accountID,
				AccessToken:        "encrypted_token",
				PlatformIdentifier: "fb_page_123",
				State:              "Processing",
				ExtraData: map[string]interface{}{
					"workspace_id": "workspace_123",
					"name":         "Test Page",
				},
			}, nil
		},
		UpdateStateFunc: func(ctx context.Context, id primitive.ObjectID, state string) error {
			return nil
		},
		UpdateAnalyticsTimestampFunc: func(ctx context.Context, id primitive.ObjectID, field string, timestamp time.Time) error {
			return nil
		},
	}

	mockFB := &mockFacebookClient{
		fetchPostsSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookPost, error) {
			return []kafkamodels.RawFacebookPost{}, nil
		},
		fetchVideosSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookVideo, error) {
			return []kafkamodels.RawFacebookVideo{}, nil
		},
		fetchInsightsFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) (*kafkamodels.RawFacebookInsights, error) {
			return nil, nil
		},
	}

	mockCH := &mockClickHouseSink{}

	p := &Processor{
		MongoRepo:      mockRepo,
		FacebookClient: mockFB,
		Sink:           mockCH,
		Parser:         parsing.NewFacebookParser(),
		Logger:         &logger.Logger{},
		Config:         &config.Config{DecryptionKey: "test-decryption-key"},
	}

	workOrder := WorkOrder{
		ID:          accountID.Hex(),
		AccountID:   "fb_page_123",
		WorkspaceID: "workspace_123",
	}

	// This should not error even if decryption fails (it will use the token as-is)
	err := p.ProcessAccount(context.Background(), workOrder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ================== isExpectedFacebookError Tests ==================

func TestIsExpectedFacebookError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"OAuthException 190", errors.New("Error validating access token - token expired"), true},
		{"OAuthException 100", errors.New("OAuthException/100"), true},
		{"OAuthException 2", errors.New("OAuthException/2"), true},
		{"Error validating access token", errors.New("Error validating access token: Session has expired"), true},
		{"GraphMethodException", errors.New("GraphMethodException/100"), true},
		{"Tried accessing nonexisting field", errors.New("Tried accessing nonexisting field (OAuthException)"), true},
		{"status 401", errors.New("facebook api error (status 401): unauthorized"), true},
		{"status 403", errors.New("facebook api error (status 403): forbidden"), true},
		{"permission denied", errors.New("does not exist, cannot be loaded due to missing permissions"), true},
		{"network error", errors.New("connection timeout"), false},
		{"parse error", errors.New("json parse failed"), false},
		{"status 500", errors.New("internal server error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isExpectedFacebookError(tt.err)
			if got != tt.expected {
				t.Errorf("isExpectedFacebookError() = %v, want %v for error: %v", got, tt.expected, tt.err)
			}
		})
	}
}

// ==================== Logging Contract Tests ====================

func TestLoggingContract_Facebook_ExpectedError_WarnOnly(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()
	captureRecords, captureCleanup := logger.InstallCaptureSpy()
	defer captureCleanup()
	_, hookCleanup := logger.InstallHookSpy()
	defer hookCleanup()

	accountID := primitive.NewObjectID()
	mockRepo := &mockMongoRepo{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:                 accountID,
				PlatformIdentifier: "fb_page_123",
				AccessToken:        "",
				ExtraData: map[string]interface{}{
					"access_token": "test_token",
					"workspace_id": "ws123",
					"name":         "Test Page",
				},
			}, nil
		},
		UpdateStateFunc: func(ctx context.Context, id primitive.ObjectID, state string) error {
			return nil
		},
		UpdateAnalyticsTimestampFunc: func(ctx context.Context, id primitive.ObjectID, field string, timestamp time.Time) error {
			return nil
		},
	}

	expectedAPIErr := errors.New("Error validating access token: Session has expired")
	mockFB := &mockFacebookClient{
		fetchPostsSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookPost, error) {
			return nil, expectedAPIErr
		},
		fetchVideosSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookVideo, error) {
			return nil, nil
		},
		fetchInsightsFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) (*kafkamodels.RawFacebookInsights, error) {
			return nil, nil
		},
	}

	mockCH := &mockClickHouseSink{}

	p := &Processor{
		MongoRepo:      mockRepo,
		FacebookClient: mockFB,
		Sink:           mockCH,
		Parser:         parsing.NewFacebookParser(),
		Logger:         log,
		Config:         &config.Config{DecryptionKey: "test-key"},
	}

	wo := WorkOrder{
		ID:          accountID.Hex(),
		AccountID:   "fb_page_123",
		WorkspaceID: "ws123",
	}

	err := p.ProcessAccount(context.Background(), wo)

	// Expected error should be swallowed (returns nil)
	if err != nil {
		t.Fatalf("expected nil error for expected API error, got: %v", err)
	}

	output := buf.String()

	// Should have WRN level
	if !strings.Contains(output, "WRN") {
		t.Error("expected WRN-level log entries")
	}

	// Should NOT have ERR level
	if strings.Contains(output, "ERR") {
		t.Error("unexpected ERR-level log entries; processors should not log at Error level")
	}

	// CaptureException should NOT have been called for the expected error
	for _, rec := range *captureRecords {
		if rec.Err != nil && strings.Contains(rec.Err.Error(), "Error validating access token") {
			t.Error("CaptureException should NOT be called for expected Facebook API errors")
		}
	}
}

func TestLoggingContract_Facebook_UnexpectedReturnedError_NoCaptureException(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()
	captureRecords, captureCleanup := logger.InstallCaptureSpy()
	defer captureCleanup()

	mockRepo := &mockMongoRepo{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return nil, errors.New("database connection lost")
		},
	}

	p := &Processor{
		MongoRepo: mockRepo,
		Logger:    log,
		Config:    &config.Config{DecryptionKey: "test-key"},
	}

	wo := WorkOrder{
		ID:          primitive.NewObjectID().Hex(),
		AccountID:   "fb_123",
		WorkspaceID: "ws_123",
	}

	err := p.ProcessAccount(context.Background(), wo)

	// Error IS returned to caller
	if err == nil {
		t.Fatal("expected error to be returned")
	}

	// CaptureException should NOT be called (main module handles returned errors)
	if len(*captureRecords) != 0 {
		t.Errorf("CaptureException should NOT be called when error is returned to caller; got %d calls", len(*captureRecords))
	}

	output := buf.String()

	// Should NOT have ERR level
	if strings.Contains(output, "ERR") {
		t.Error("unexpected ERR-level log entries in processor")
	}
}

func TestLoggingContract_Facebook_SwallowedError_UsesCaptureException(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()
	captureRecords, captureCleanup := logger.InstallCaptureSpy()
	defer captureCleanup()

	accountID := primitive.NewObjectID()
	mockRepo := &mockMongoRepo{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:                 accountID,
				PlatformIdentifier: "fb_page_123",
				AccessToken:        "",
				ExtraData: map[string]interface{}{
					"access_token": "test_token",
					"workspace_id": "ws123",
					"name":         "Test Page",
				},
			}, nil
		},
		UpdateStateFunc: func(ctx context.Context, id primitive.ObjectID, state string) error {
			return nil
		},
		UpdateAnalyticsTimestampFunc: func(ctx context.Context, id primitive.ObjectID, field string, timestamp time.Time) error {
			return nil
		},
	}

	unexpectedErr := errors.New("random network timeout")
	mockFB := &mockFacebookClient{
		fetchPostsSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookPost, error) {
			return nil, unexpectedErr
		},
		fetchVideosSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookVideo, error) {
			return nil, nil
		},
		fetchInsightsFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) (*kafkamodels.RawFacebookInsights, error) {
			return nil, nil
		},
	}

	mockCH := &mockClickHouseSink{}

	p := &Processor{
		MongoRepo:      mockRepo,
		FacebookClient: mockFB,
		Sink:           mockCH,
		Parser:         parsing.NewFacebookParser(),
		Logger:         log,
		Config:         &config.Config{DecryptionKey: "test-key"},
	}

	wo := WorkOrder{
		ID:          accountID.Hex(),
		AccountID:   "fb_page_123",
		WorkspaceID: "ws123",
	}

	err := p.ProcessAccount(context.Background(), wo)

	// Error should be swallowed (returns nil)
	if err != nil {
		t.Fatalf("expected nil error (swallowed), got: %v", err)
	}

	// CaptureException SHOULD have been called for the swallowed unexpected error
	found := false
	for _, rec := range *captureRecords {
		if rec.Err == unexpectedErr {
			found = true
			break
		}
	}
	if !found {
		t.Error("CaptureException should be called for unexpected swallowed errors")
	}

	output := buf.String()

	if !strings.Contains(output, "WRN") {
		t.Error("expected WRN-level log for swallowed error")
	}
	if strings.Contains(output, "ERR") {
		t.Error("unexpected ERR-level log; processors should not log at Error level")
	}
}

func TestLoggingContract_Facebook_NoErrorLogDuplication(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()
	_, hookCleanup := logger.InstallHookSpy()
	defer hookCleanup()

	accountID := primitive.NewObjectID()
	mockRepo := &mockMongoRepo{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:                 accountID,
				PlatformIdentifier: "fb_page_123",
				AccessToken:        "",
				ExtraData: map[string]interface{}{
					"access_token": "test_token",
					"workspace_id": "ws123",
					"name":         "Test Page",
				},
			}, nil
		},
		UpdateStateFunc: func(ctx context.Context, id primitive.ObjectID, state string) error {
			return nil
		},
		UpdateAnalyticsTimestampFunc: func(ctx context.Context, id primitive.ObjectID, field string, timestamp time.Time) error {
			return nil
		},
	}

	mockFB := &mockFacebookClient{
		fetchPostsSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookPost, error) {
			return nil, errors.New("posts fetch failed")
		},
		fetchVideosSinceFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookVideo, error) {
			return nil, errors.New("videos fetch failed")
		},
		fetchInsightsFunc: func(ctx context.Context, pageID, accessToken string, since, until time.Time) (*kafkamodels.RawFacebookInsights, error) {
			return nil, errors.New("insights fetch failed")
		},
	}

	mockCH := &mockClickHouseSink{}

	p := &Processor{
		MongoRepo:      mockRepo,
		FacebookClient: mockFB,
		Sink:           mockCH,
		Parser:         parsing.NewFacebookParser(),
		Logger:         log,
		Config:         &config.Config{DecryptionKey: "test-key"},
	}

	wo := WorkOrder{
		ID:          accountID.Hex(),
		AccountID:   "fb_page_123",
		WorkspaceID: "ws123",
	}

	_ = p.ProcessAccount(context.Background(), wo)

	output := buf.String()
	errCount := strings.Count(output, "ERR")
	if errCount > 0 {
		t.Errorf("expected 0 ERR-level entries, got %d; processors should never log at Error level", errCount)
	}
}
