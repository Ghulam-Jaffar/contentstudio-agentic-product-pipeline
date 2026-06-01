package processor

import (
	"context"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ClickHouseClientInterface is an alias to the shared interface in conversions package.
type ClickHouseClientInterface = conversions.ClickHouseClientInterface

// TwitterAPI is an alias to the shared interface in clients/social package.
type TwitterAPI = social.TwitterAPI

// TwitterTweetFetcher defines the subset of Twitter API operations needed by this processor.
type TwitterTweetFetcher interface {
	FetchUserTweets(ctx context.Context, userID, oauthToken, oauthTokenSecret string, maxResults int, paginationToken string) (*social.TwitterTweetsResponse, error)
	FetchUserInfo(ctx context.Context, userIDs []string, oauthToken, oauthTokenSecret string) (*social.TwitterUserResponse, error)
}

// TwitterPostSink is the interface for storing Twitter posts and insights in ClickHouse.
type TwitterPostSink interface {
	BulkInsertTwitterPosts(ctx context.Context, posts []*chmodels.TwitterPosts) error
	BulkInsertTwitterInsights(ctx context.Context, insights []*chmodels.TwitterInsights) error
}

// SocialRepository is the interface for accessing and updating social account data from MongoDB.
type SocialRepository interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error)
	Update(ctx context.Context, id primitive.ObjectID, updates bson.M) error
	InsertTwitterJobMetadata(ctx context.Context, payload mongodb.TwitterJobMetadataPayload) error
	RecordProcessingError(ctx context.Context, id primitive.ObjectID, errorMessage string) error
	ClearProcessingError(ctx context.Context, id primitive.ObjectID) error
}

// PusherClientInterface abstracts the Pusher client for testing.
type PusherClientInterface interface {
	Trigger(channel string, event string, data interface{}) error
}

// NotifierInterface abstracts the notification service for testing.
type NotifierInterface interface {
	SendAnalyticsNotification(userID, workspaceID, platform, accountID, accountName string, isCompetitor bool) error
}

// ProcessorInterface defines the interface for Twitter account processing.
type ProcessorInterface interface {
	ProcessAccount(ctx context.Context, wo ImmediateWorkOrder) error
}
