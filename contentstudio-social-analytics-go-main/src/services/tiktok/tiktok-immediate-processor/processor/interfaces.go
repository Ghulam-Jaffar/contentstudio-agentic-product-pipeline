package processor

import (
	"context"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ClickHouseClientInterface is an alias to the shared interface in conversions package.
// This allows services to use the common interface for ClickHouse operations.
type ClickHouseClientInterface = conversions.ClickHouseClientInterface

// TikTokAPI is an alias to the shared interface in clients/social package.
// This allows services to use the common interface for TikTok API operations.
type TikTokAPI = social.TikTokAPI

// TikTokClientInterface defines the subset of TikTok API operations needed by this processor.
// This is a smaller interface than the full TikTokAPI for dependency minimization.
type TikTokClientInterface interface {
	FetchVideosSince(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]kafkamodels.RawTikTokPost, error)
	FetchInsights(ctx context.Context, accountID, accessToken string, since, until time.Time) (*kafkamodels.ParsedTikTokInsights, error)
}

// PusherClientInterface abstracts the Pusher client for testing
type PusherClientInterface interface {
	Trigger(channel string, event string, data interface{}) error
}

// NotifierInterface abstracts the notification service for testing
type NotifierInterface interface {
	SendAnalyticsNotification(userID, workspaceID, platform, accountID, accountName string, isCompetitor bool) error
}

// ClickHouseSinkInterface defines the interface for ClickHouse data conversion operations
type ClickHouseSinkInterface interface {
	ConvertTikTokPost(post *kafkamodels.ParsedTikTokPost) *clickhousemodels.TikTokPosts
	ConvertTikTokInsight(insight *kafkamodels.ParsedTikTokInsights) *clickhousemodels.TikTokInsights
}

// ProcessorInterface defines the interface for TikTok account processing
type ProcessorInterface interface {
	ProcessAccount(wo ImmediateWorkOrder) error
}

// sinkCreator is a function type for creating ClickHouse sinks (allows mocking in tests)
type sinkCreator func() ClickHouseSinkInterface
