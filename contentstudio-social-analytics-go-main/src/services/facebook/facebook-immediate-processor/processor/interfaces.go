package processor

import (
	"context"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// FacebookAPI is an alias to the shared interface in clients/social package.
// This allows services to use the common interface for Facebook API operations.
type FacebookAPI = social.FacebookAPI

// FacebookClientInterface defines the subset of Facebook API operations needed by this processor.
// This is a smaller interface than the full FacebookAPI for dependency minimization.
type FacebookClientInterface interface {
	FetchPostsSince(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookPost, error)
	FetchVideosSince(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookVideo, error)
	FetchInsights(ctx context.Context, pageID, accessToken string, since, until time.Time) (*kafkamodels.RawFacebookInsights, error)
}

// PusherClientInterface abstracts the Pusher client for testing
type PusherClientInterface interface {
	Trigger(channel string, event string, data interface{}) error
}

// NotifierInterface abstracts the notification service for testing
type NotifierInterface interface {
	SendAnalyticsNotification(userID, workspaceID, platform, accountID, accountName string, isCompetitor bool) error
}

// ClickHouseSinkInterface defines the interface for ClickHouse data conversion and storage operations.
// The ClickHouseSink implements both conversion and bulk-insert methods, so a single interface
// replaces the previous split between ClickHouseClientInterface (inserts) and a separate sink (conversions).
type ClickHouseSinkInterface interface {
	// Conversion methods
	ConvertFacebookPost(post *kafkamodels.ParsedFacebookPost) *clickhousemodels.FacebookPosts
	ConvertFacebookMediaAssets(asset *kafkamodels.ParsedFacebookMediaAsset) *clickhousemodels.FacebookMediaAssets
	ConvertFacebookVideoInsights(insights *kafkamodels.ParsedFacebookVideoInsights) *clickhousemodels.FacebookVideoInsights
	ConvertFacebookReelsInsights(insights *kafkamodels.ParsedFacebookReelsInsights) *clickhousemodels.FacebookReelsInsights
	ConvertFacebookInsights(insights *kafkamodels.ParsedFacebookInsights) *clickhousemodels.FacebookInsights
	// Bulk insert methods
	BulkInsertPosts(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error
	BulkInsertMediaAssets(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error
	BulkInsertVideoInsights(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error
	BulkInsertReelsInsights(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error
	BulkInsertInsights(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error
}
