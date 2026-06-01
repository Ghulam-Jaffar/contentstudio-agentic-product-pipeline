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

// ClickHouseClientInterface is an alias to the shared interface in conversions package.
type ClickHouseClientInterface = conversions.ClickHouseClientInterface

// GMBAPI is an alias to the shared GMB API interface.
type GMBAPI = social.GMBAPI

// GMBClientInterface defines the subset of GMB API operations needed by this processor.
type GMBClientInterface interface {
	RefreshToken(ctx context.Context, refreshToken string) (*social.RefreshTokenResponse, error)
	FetchVoiceOfMerchant(ctx context.Context, locationID, accessToken string) (*social.VoiceOfMerchantResponse, error)
	FetchPerformanceMetrics(ctx context.Context, locationID, accessToken string, startDate, endDate time.Time) (*social.GMBPerformanceResponse, error)
	FetchSearchKeywords(ctx context.Context, locationID, accessToken string, startMonth, endMonth time.Time) (*social.GMBSearchKeywordsResponse, error)
	FetchLocalPosts(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*social.GMBLocalPostsResponse, error)
	FetchReviews(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*social.GMBReviewsResponse, error)
	FetchMediaAssets(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*social.GMBMediaAssetsResponse, error)
}

// PusherClientInterface abstracts the Pusher client for testing
type PusherClientInterface interface {
	Trigger(channel string, event string, data interface{}) error
}

// NotifierInterface abstracts the notification service for testing
type NotifierInterface interface {
	SendAnalyticsNotification(userID, workspaceID, platform, accountID, accountName string, isCompetitor bool) error
}

// ClickHouseSinkInterface defines the interface for GMB ClickHouse data operations
type ClickHouseSinkInterface interface {
	BulkInsertGMBDailyMetrics(ctx context.Context, metrics []*clickhousemodels.GMBDailyMetrics) error
	BulkInsertGMBMediaAssets(ctx context.Context, assets []*clickhousemodels.GMBMediaAssets) error
	BulkInsertGMBSearchKeywordsMonthly(ctx context.Context, keywords []*clickhousemodels.GMBSearchKeywordsMonthly) error
	BulkInsertGMBLocalPosts(ctx context.Context, posts []*clickhousemodels.GMBLocalPosts) error
	BulkInsertGMBReviews(ctx context.Context, reviews []*clickhousemodels.GMBReviews) error
}

// ProcessorInterface defines the interface for GMB account processing
type ProcessorInterface interface {
	ProcessAccount(wo ImmediateWorkOrder) error
}

// ================== Shared Aliases ==================

// UnifiedSocialRepository is an alias for the shared MongoDB repository interface.
type UnifiedSocialRepository = mongodb.UnifiedSocialRepository

// KafkaConsumer is an alias to the shared interface in kafka package.
type KafkaConsumer = kafka.Consumer
