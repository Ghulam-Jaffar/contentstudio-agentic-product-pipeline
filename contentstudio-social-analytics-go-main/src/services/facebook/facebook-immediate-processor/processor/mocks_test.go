package processor

import (
	"context"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ================== Shared Mock Aliases ==================
// These aliases allow tests to use the shared mocks from their respective packages.
// Actual implementations are in:
//   - db/mongodb/mock.go
//   - kafka/mock.go

// mockMongoRepo is an alias for the shared MongoDB mock
type mockMongoRepo = mongodb.MockUnifiedSocialRepository

// mockKafkaProducer is an alias for the shared Kafka producer mock
type mockKafkaProducer = kafka.MockProducer

// ================== Mock Facebook Client ==================

type mockFacebookClient struct {
	fetchPostsSinceFunc  func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookPost, error)
	fetchVideosSinceFunc func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookVideo, error)
	fetchInsightsFunc    func(ctx context.Context, pageID, accessToken string, since, until time.Time) (*kafkamodels.RawFacebookInsights, error)
}

func (m *mockFacebookClient) FetchPostsSince(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookPost, error) {
	if m.fetchPostsSinceFunc != nil {
		return m.fetchPostsSinceFunc(ctx, pageID, accessToken, since, until)
	}
	return nil, nil
}

func (m *mockFacebookClient) FetchVideosSince(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookVideo, error) {
	if m.fetchVideosSinceFunc != nil {
		return m.fetchVideosSinceFunc(ctx, pageID, accessToken, since, until)
	}
	return nil, nil
}

func (m *mockFacebookClient) FetchInsights(ctx context.Context, pageID, accessToken string, since, until time.Time) (*kafkamodels.RawFacebookInsights, error) {
	if m.fetchInsightsFunc != nil {
		return m.fetchInsightsFunc(ctx, pageID, accessToken, since, until)
	}
	return nil, nil
}

var _ FacebookClientInterface = (*mockFacebookClient)(nil)

// ================== Mock ClickHouse Sink ==================

type mockClickHouseSink struct {
	// Conversion functions
	convertPostFunc          func(post *kafkamodels.ParsedFacebookPost) *clickhousemodels.FacebookPosts
	convertMediaAssetsFunc   func(asset *kafkamodels.ParsedFacebookMediaAsset) *clickhousemodels.FacebookMediaAssets
	convertVideoInsightsFunc func(insights *kafkamodels.ParsedFacebookVideoInsights) *clickhousemodels.FacebookVideoInsights
	convertReelsInsightsFunc func(insights *kafkamodels.ParsedFacebookReelsInsights) *clickhousemodels.FacebookReelsInsights
	convertInsightsFunc      func(insights *kafkamodels.ParsedFacebookInsights) *clickhousemodels.FacebookInsights
	// Bulk insert functions
	BulkInsertPostsFunc         func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error
	BulkInsertMediaAssetsFunc   func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error
	BulkInsertVideoInsightsFunc func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error
	BulkInsertReelsInsightsFunc func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error
	BulkInsertInsightsFunc      func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error
}

func (m *mockClickHouseSink) ConvertFacebookPost(post *kafkamodels.ParsedFacebookPost) *clickhousemodels.FacebookPosts {
	if m.convertPostFunc != nil {
		return m.convertPostFunc(post)
	}
	return &clickhousemodels.FacebookPosts{
		PageID: post.PageID,
		PostID: post.PostID,
	}
}

func (m *mockClickHouseSink) ConvertFacebookMediaAssets(asset *kafkamodels.ParsedFacebookMediaAsset) *clickhousemodels.FacebookMediaAssets {
	if m.convertMediaAssetsFunc != nil {
		return m.convertMediaAssetsFunc(asset)
	}
	return &clickhousemodels.FacebookMediaAssets{
		PostID: asset.PostID,
	}
}

func (m *mockClickHouseSink) ConvertFacebookVideoInsights(insights *kafkamodels.ParsedFacebookVideoInsights) *clickhousemodels.FacebookVideoInsights {
	if m.convertVideoInsightsFunc != nil {
		return m.convertVideoInsightsFunc(insights)
	}
	return &clickhousemodels.FacebookVideoInsights{
		PostID: insights.PostID,
	}
}

func (m *mockClickHouseSink) ConvertFacebookReelsInsights(insights *kafkamodels.ParsedFacebookReelsInsights) *clickhousemodels.FacebookReelsInsights {
	if m.convertReelsInsightsFunc != nil {
		return m.convertReelsInsightsFunc(insights)
	}
	return &clickhousemodels.FacebookReelsInsights{
		PostID: insights.PostID,
	}
}

func (m *mockClickHouseSink) ConvertFacebookInsights(insights *kafkamodels.ParsedFacebookInsights) *clickhousemodels.FacebookInsights {
	if m.convertInsightsFunc != nil {
		return m.convertInsightsFunc(insights)
	}
	return &clickhousemodels.FacebookInsights{
		PageID: insights.PageID,
	}
}

func (m *mockClickHouseSink) BulkInsertPosts(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
	if m.BulkInsertPostsFunc != nil {
		return m.BulkInsertPostsFunc(ctx, posts)
	}
	return nil
}

func (m *mockClickHouseSink) BulkInsertMediaAssets(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
	if m.BulkInsertMediaAssetsFunc != nil {
		return m.BulkInsertMediaAssetsFunc(ctx, assets)
	}
	return nil
}

func (m *mockClickHouseSink) BulkInsertVideoInsights(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
	if m.BulkInsertVideoInsightsFunc != nil {
		return m.BulkInsertVideoInsightsFunc(ctx, insights)
	}
	return nil
}

func (m *mockClickHouseSink) BulkInsertReelsInsights(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
	if m.BulkInsertReelsInsightsFunc != nil {
		return m.BulkInsertReelsInsightsFunc(ctx, insights)
	}
	return nil
}

func (m *mockClickHouseSink) BulkInsertInsights(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
	if m.BulkInsertInsightsFunc != nil {
		return m.BulkInsertInsightsFunc(ctx, insights)
	}
	return nil
}

var _ ClickHouseSinkInterface = (*mockClickHouseSink)(nil)

// ================== Mock Pusher Client ==================

type mockPusherClient struct {
	triggerCalled bool
	lastChannel   string
	lastEvent     string
	lastData      interface{}
}

func (m *mockPusherClient) Trigger(channel string, event string, data interface{}) error {
	m.triggerCalled = true
	m.lastChannel = channel
	m.lastEvent = event
	m.lastData = data
	return nil
}

var _ PusherClientInterface = (*mockPusherClient)(nil)

// ================== Mock Notifier ==================

type mockNotifier struct {
	sendCalled      bool
	lastUserID      string
	lastWorkspaceID string
	lastPlatform    string
	lastAccountID   string
	lastAccountName string
}

func (m *mockNotifier) SendAnalyticsNotification(userID, workspaceID, platform, accountID, accountName string, isCompetitor bool) error {
	m.sendCalled = true
	m.lastUserID = userID
	m.lastWorkspaceID = workspaceID
	m.lastPlatform = platform
	m.lastAccountID = accountID
	m.lastAccountName = accountName
	return nil
}

var _ NotifierInterface = (*mockNotifier)(nil)
