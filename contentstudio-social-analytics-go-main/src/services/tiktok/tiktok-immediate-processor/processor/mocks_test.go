package processor

import (
	"context"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ================== Shared Mock Aliases ==================
// These aliases allow tests to use the shared mocks from their respective packages.
// Actual implementations are in:
//   - db/mongodb/mock.go
//   - models/db/clickhouse/conversions/mock.go
//   - kafka/mock.go

// mockClickHouseClient is an alias for the shared ClickHouse mock
type mockClickHouseClient = conversions.MockClickHouseClient

// mockMongoRepo is an alias for the shared MongoDB mock
type mockMongoRepo = mongodb.MockUnifiedSocialRepository

// mockKafkaProducer is an alias for the shared Kafka producer mock
type mockKafkaProducer = kafka.MockProducer

// ================== Mock TikTok Client ==================

type mockTikTokClient struct {
	fetchVideosSinceFunc func(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]kafkamodels.RawTikTokPost, error)
	fetchInsightsFunc    func(ctx context.Context, accountID, accessToken string, since, until time.Time) (*kafkamodels.ParsedTikTokInsights, error)
}

func (m *mockTikTokClient) FetchVideosSince(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]kafkamodels.RawTikTokPost, error) {
	if m.fetchVideosSinceFunc != nil {
		return m.fetchVideosSinceFunc(ctx, accountID, accessToken, since, until)
	}
	return nil, nil
}

func (m *mockTikTokClient) FetchInsights(ctx context.Context, accountID, accessToken string, since, until time.Time) (*kafkamodels.ParsedTikTokInsights, error) {
	if m.fetchInsightsFunc != nil {
		return m.fetchInsightsFunc(ctx, accountID, accessToken, since, until)
	}
	return nil, nil
}

var _ TikTokClientInterface = (*mockTikTokClient)(nil)

// ================== Mock ClickHouse Sink ==================

type mockClickHouseSink struct {
	convertPostFunc    func(post *kafkamodels.ParsedTikTokPost) *clickhousemodels.TikTokPosts
	convertInsightFunc func(insight *kafkamodels.ParsedTikTokInsights) *clickhousemodels.TikTokInsights
}

func (m *mockClickHouseSink) ConvertTikTokPost(post *kafkamodels.ParsedTikTokPost) *clickhousemodels.TikTokPosts {
	if m.convertPostFunc != nil {
		return m.convertPostFunc(post)
	}
	return &clickhousemodels.TikTokPosts{
		TikTokID: post.ID,
	}
}

func (m *mockClickHouseSink) ConvertTikTokInsight(insight *kafkamodels.ParsedTikTokInsights) *clickhousemodels.TikTokInsights {
	if m.convertInsightFunc != nil {
		return m.convertInsightFunc(insight)
	}
	return &clickhousemodels.TikTokInsights{
		TikTokID: insight.TikTokID,
		RecordID: insight.RecordID,
	}
}

var _ ClickHouseSinkInterface = (*mockClickHouseSink)(nil)

// ================== Mock Pusher Client ==================

type mockPusherClient struct {
	triggerFunc func(channel string, event string, data interface{}) error
}

func (m *mockPusherClient) Trigger(channel string, event string, data interface{}) error {
	if m.triggerFunc != nil {
		return m.triggerFunc(channel, event, data)
	}
	return nil
}

var _ PusherClientInterface = (*mockPusherClient)(nil)

// ================== Mock Notifier ==================

type mockNotifier struct {
	sendAnalyticsNotificationFunc func(userID, workspaceID, platform, accountID, accountName string, isCompetitor bool) error
}

func (m *mockNotifier) SendAnalyticsNotification(userID, workspaceID, platform, accountID, accountName string, isCompetitor bool) error {
	if m.sendAnalyticsNotificationFunc != nil {
		return m.sendAnalyticsNotificationFunc(userID, workspaceID, platform, accountID, accountName, isCompetitor)
	}
	return nil
}

var _ NotifierInterface = (*mockNotifier)(nil)
