package processor

import (
	"context"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
)

// mockClickHouseClient is an alias for the shared ClickHouse mock
type mockClickHouseClient = conversions.MockClickHouseClient

// mockMongoRepo is an alias for the shared MongoDB mock
type mockMongoRepo = mongodb.MockUnifiedSocialRepository

// mockKafkaProducer is an alias for the shared Kafka producer mock
type mockKafkaProducer = kafka.MockProducer

// ================== Mock GMB Client ==================

type mockGMBClient struct {
	refreshTokenFunc            func(ctx context.Context, refreshToken string) (*social.RefreshTokenResponse, error)
	fetchVoiceOfMerchantFunc    func(ctx context.Context, locationID, accessToken string) (*social.VoiceOfMerchantResponse, error)
	fetchPerformanceMetricsFunc func(ctx context.Context, locationID, accessToken string, startDate, endDate time.Time) (*social.GMBPerformanceResponse, error)
	fetchSearchKeywordsFunc     func(ctx context.Context, locationID, accessToken string, startMonth, endMonth time.Time) (*social.GMBSearchKeywordsResponse, error)
	fetchLocalPostsFunc         func(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*social.GMBLocalPostsResponse, error)
	fetchReviewsFunc            func(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*social.GMBReviewsResponse, error)
	fetchMediaAssetsFunc        func(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*social.GMBMediaAssetsResponse, error)
}

var _ GMBClientInterface = (*mockGMBClient)(nil)

func (m *mockGMBClient) RefreshToken(ctx context.Context, refreshToken string) (*social.RefreshTokenResponse, error) {
	if m.refreshTokenFunc != nil {
		return m.refreshTokenFunc(ctx, refreshToken)
	}
	return &social.RefreshTokenResponse{AccessToken: "mock-access-token"}, nil
}

func (m *mockGMBClient) FetchVoiceOfMerchant(ctx context.Context, locationID, accessToken string) (*social.VoiceOfMerchantResponse, error) {
	if m.fetchVoiceOfMerchantFunc != nil {
		return m.fetchVoiceOfMerchantFunc(ctx, locationID, accessToken)
	}
	return &social.VoiceOfMerchantResponse{HasVoiceOfMerchant: true}, nil
}

func (m *mockGMBClient) FetchPerformanceMetrics(ctx context.Context, locationID, accessToken string, startDate, endDate time.Time) (*social.GMBPerformanceResponse, error) {
	if m.fetchPerformanceMetricsFunc != nil {
		return m.fetchPerformanceMetricsFunc(ctx, locationID, accessToken, startDate, endDate)
	}
	return &social.GMBPerformanceResponse{}, nil
}

func (m *mockGMBClient) FetchSearchKeywords(ctx context.Context, locationID, accessToken string, startMonth, endMonth time.Time) (*social.GMBSearchKeywordsResponse, error) {
	if m.fetchSearchKeywordsFunc != nil {
		return m.fetchSearchKeywordsFunc(ctx, locationID, accessToken, startMonth, endMonth)
	}
	return &social.GMBSearchKeywordsResponse{}, nil
}

func (m *mockGMBClient) FetchLocalPosts(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*social.GMBLocalPostsResponse, error) {
	if m.fetchLocalPostsFunc != nil {
		return m.fetchLocalPostsFunc(ctx, accountID, locationID, accessToken, pageToken)
	}
	return &social.GMBLocalPostsResponse{}, nil
}

func (m *mockGMBClient) FetchReviews(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*social.GMBReviewsResponse, error) {
	if m.fetchReviewsFunc != nil {
		return m.fetchReviewsFunc(ctx, accountID, locationID, accessToken, pageToken)
	}
	return &social.GMBReviewsResponse{}, nil
}

func (m *mockGMBClient) FetchMediaAssets(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*social.GMBMediaAssetsResponse, error) {
	if m.fetchMediaAssetsFunc != nil {
		return m.fetchMediaAssetsFunc(ctx, accountID, locationID, accessToken, pageToken)
	}
	return &social.GMBMediaAssetsResponse{}, nil
}

// ================== Mock ClickHouse Sink ==================

type mockClickHouseSink struct {
	bulkInsertDailyMetricsFunc   func(ctx context.Context, metrics []*clickhousemodels.GMBDailyMetrics) error
	bulkInsertMediaAssetsFunc    func(ctx context.Context, assets []*clickhousemodels.GMBMediaAssets) error
	bulkInsertSearchKeywordsFunc func(ctx context.Context, keywords []*clickhousemodels.GMBSearchKeywordsMonthly) error
	bulkInsertLocalPostsFunc     func(ctx context.Context, posts []*clickhousemodels.GMBLocalPosts) error
	bulkInsertReviewsFunc        func(ctx context.Context, reviews []*clickhousemodels.GMBReviews) error
}

var _ ClickHouseSinkInterface = (*mockClickHouseSink)(nil)

func (m *mockClickHouseSink) BulkInsertGMBDailyMetrics(ctx context.Context, metrics []*clickhousemodels.GMBDailyMetrics) error {
	if m.bulkInsertDailyMetricsFunc != nil {
		return m.bulkInsertDailyMetricsFunc(ctx, metrics)
	}
	return nil
}

func (m *mockClickHouseSink) BulkInsertGMBMediaAssets(ctx context.Context, assets []*clickhousemodels.GMBMediaAssets) error {
	if m.bulkInsertMediaAssetsFunc != nil {
		return m.bulkInsertMediaAssetsFunc(ctx, assets)
	}
	return nil
}

func (m *mockClickHouseSink) BulkInsertGMBSearchKeywordsMonthly(ctx context.Context, keywords []*clickhousemodels.GMBSearchKeywordsMonthly) error {
	if m.bulkInsertSearchKeywordsFunc != nil {
		return m.bulkInsertSearchKeywordsFunc(ctx, keywords)
	}
	return nil
}

func (m *mockClickHouseSink) BulkInsertGMBLocalPosts(ctx context.Context, posts []*clickhousemodels.GMBLocalPosts) error {
	if m.bulkInsertLocalPostsFunc != nil {
		return m.bulkInsertLocalPostsFunc(ctx, posts)
	}
	return nil
}

func (m *mockClickHouseSink) BulkInsertGMBReviews(ctx context.Context, reviews []*clickhousemodels.GMBReviews) error {
	if m.bulkInsertReviewsFunc != nil {
		return m.bulkInsertReviewsFunc(ctx, reviews)
	}
	return nil
}

// ================== Mock Pusher Client ==================

type mockPusherClient struct {
	triggerFunc func(channel string, event string, data interface{}) error
}

var _ PusherClientInterface = (*mockPusherClient)(nil)

func (m *mockPusherClient) Trigger(channel string, event string, data interface{}) error {
	if m.triggerFunc != nil {
		return m.triggerFunc(channel, event, data)
	}
	return nil
}

// ================== Mock Notifier ==================

type mockNotifier struct {
	sendAnalyticsNotificationFunc func(userID, workspaceID, platform, accountID, accountName string, isCompetitor bool) error
}

var _ NotifierInterface = (*mockNotifier)(nil)

func (m *mockNotifier) SendAnalyticsNotification(userID, workspaceID, platform, accountID, accountName string, isCompetitor bool) error {
	if m.sendAnalyticsNotificationFunc != nil {
		return m.sendAnalyticsNotificationFunc(userID, workspaceID, platform, accountID, accountName, isCompetitor)
	}
	return nil
}
