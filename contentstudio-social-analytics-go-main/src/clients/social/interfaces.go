package social

import (
	"context"
	"encoding/json"
	"time"

	models "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// RefreshTokenResponse represents the response from a token refresh operation
// It includes both the new tokens and their expiration information
type RefreshTokenResponse struct {
	AccessToken      string      `json:"access_token"`
	RefreshToken     string      `json:"refresh_token"`
	ExpiresIn        interface{} `json:"expires_in"`         // Can be int (seconds) or object with date/timezone
	RefreshExpiresIn interface{} `json:"refresh_expires_in"` // Can be int (seconds) or object with date/timezone
	Scope            string      `json:"scope,omitempty"`
}

// FacebookAPI defines the interface for Facebook API operations.
// This interface is used for dependency injection and testing.
type FacebookAPI interface {
	// Post fetching
	FetchPosts(ctx context.Context, pageID, accessToken string) ([]kafkamodels.RawFacebookPost, error)
	FetchPostsWithLimit(ctx context.Context, pageID, accessToken string, maxPages int) ([]kafkamodels.RawFacebookPost, error)
	FetchPostsSince(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookPost, error)

	// Video fetching
	FetchVideos(ctx context.Context, pageID, accessToken string, maxPages int) ([]kafkamodels.RawFacebookVideo, error)
	FetchVideosWithLimit(ctx context.Context, pageID, accessToken string, maxPages int) ([]kafkamodels.RawFacebookVideo, error)
	FetchVideosSince(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookVideo, error)

	// Insights
	FetchInsights(ctx context.Context, pageID, accessToken string, since, until time.Time) (*kafkamodels.RawFacebookInsights, error)

	// Thumbnails
	GetPostThumbnails(ctx context.Context, pageID, accessToken, longAccessToken, decryptionKey string, posts []clickhouse.MinimalPost) ([]clickhouse.MinimalPost, error)

	// Competitor methods
	GetCompetitorPageDetails(ctx context.Context, pageID, accessToken string) (*models.FacebookPageDetails, *models.Picture, error)
	GetCompetitorPosts(ctx context.Context, pageID, accessToken string, since, until time.Time, limit int) ([]*models.Post, string, error)
	GetCompetitorPostsFromURL(ctx context.Context, nextURL, pageID, accessToken string) ([]*models.Post, string, error)
	GetCompetitorSharedPostDetails(ctx context.Context, parentID, accessToken string) (*models.Post, error)
	GetCompetitorPagePicture(ctx context.Context, pageID, accessToken string) (*models.Picture, error)
}

// InstagramAPI defines the interface for Instagram API operations.
// This interface is used for dependency injection and testing.
type InstagramAPI interface {
	// Media fetching
	FetchMedia(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error)
	FetchMediaSince(ctx context.Context, instagramID, accessToken string, since time.Time) ([]kafkamodels.RawInstagramMedia, error)
	FetchMediaWithLimit(ctx context.Context, instagramID, accessToken string, maxPages int) ([]kafkamodels.RawInstagramMedia, error)
	FetchAllMedia(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error)
	FetchAccountMedia(ctx context.Context, instagramID, accessToken string, limit int) (*kafkamodels.RawInstagramAccountResponse, error)

	// Stories
	FetchStories(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error)

	// Insights
	FetchInsights(ctx context.Context, instagramID, accessToken string, since, until time.Time) (*kafkamodels.RawInstagramInsightsResponse, error)
	FetchInsightsDaily(ctx context.Context, instagramID, accessToken string, days, concurrency int) ([]DailyInsight, error)
	FetchMediaInsights(ctx context.Context, mediaID, accessToken, mediaType, mediaProductType string) (*kafkamodels.RawInstagramMediaInsights, error)

	// Account info
	FetchAccountDemographics(ctx context.Context, instagramID, accessToken string) (*kafkamodels.RawInstagramDemographics, error)
	FetchUserInfo(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error)

	// Configuration
	WithBaseURL(url string) *InstagramClient
}

// LinkedInAPI defines the interface for LinkedIn API operations.
// This interface is used for dependency injection and testing.
type LinkedInAPI interface {
	// Posts fetching
	FetchShares(ctx context.Context, organisationID, accessToken string) ([]json.RawMessage, error)
	FetchPostsPaginated(ctx context.Context, linkedinID string, entityType string, accessToken string, cutoffTime time.Time) ([]json.RawMessage, error)

	// Media fetching
	FetchImagesRaw(ctx context.Context, ids []string, accessToken string) ([]byte, error)
	FetchVideosRaw(ctx context.Context, ids []string, accessToken string) ([]byte, error)
	FetchDocumentsRaw(ctx context.Context, ids []string, accessToken string) ([]byte, error)

	// Stats
	FetchStatsRaw(ctx context.Context, linkedinID string, ugcPosts []string, shares []string, accessToken string) ([]byte, error)

	// Follower data
	FetchFollowerData(ctx context.Context, linkedinID string, accessToken string) ([]byte, error)
	FetchFollowerDataWithGeoNames(ctx context.Context, linkedinID string, accessToken string, geoNames map[string]string) ([]byte, error)
	FetchFollowerStatsWithGeoIDs(ctx context.Context, linkedinID string, accessToken string) (*FollowerStatsWithGeoIDs, error)
	BuildFollowerDataWithGeoNames(stats *FollowerStatsWithGeoIDs, geoNames map[string]string) ([]byte, error)

	// Geo resolution
	GetGeoIDsFromFollowerStatsRaw(ctx context.Context, linkedinID string, accessToken string) ([]string, error)
	GetGeoIDsWithTypeFromFollowerStatsRaw(ctx context.Context, linkedinID string, accessToken string) ([]GeoIDWithType, error)
	ResolveGeoIDs(ctx context.Context, geoIDs []string, accessToken string) (map[string]string, error)

	// Organization info
	FetchOrganizationDetailsRaw(ctx context.Context, linkedinID string, accessToken string) ([]byte, error)
	FetchPageStatisticsRaw(ctx context.Context, linkedinID string, accessToken string, startMs, endMs int64) ([]byte, error)
	FetchShareStatisticsRaw(ctx context.Context, linkedinID string, accessToken string, startMs, endMs int64) ([]byte, error)

	// Member/Creator analytics
	FetchMemberCreatorPostAnalyticsRaw(ctx context.Context, accessToken string, queryType string, startDate, endDate *time.Time) ([]byte, error)
	FetchMemberFollowersCountRaw(ctx context.Context, accessToken string, startDate, endDate *time.Time) ([]byte, error)
}

// TikTokAPI defines the interface for TikTok API operations.
// This interface is used for dependency injection and testing.
type TikTokAPI interface {
	FetchUserVideos(ctx context.Context, userID, accessToken string, cursor, maxCount int) (json.RawMessage, int64, error)
	FetchUserInfo(ctx context.Context, accessToken string) (json.RawMessage, error)
	FetchVideoList(ctx context.Context, accessToken string, cursor int64, maxCount int) (json.RawMessage, int64, bool, error)
	RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error)
}

// GMBAPI defines the interface for Google My Business / Google Business Profile API operations.
// This interface is used for dependency injection and testing.
type GMBAPI interface {
	RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error)
	FetchVoiceOfMerchant(ctx context.Context, locationID, accessToken string) (*VoiceOfMerchantResponse, error)
	FetchPerformanceMetrics(ctx context.Context, locationID, accessToken string, startDate, endDate time.Time) (*GMBPerformanceResponse, error)
	FetchSearchKeywords(ctx context.Context, locationID, accessToken string, startMonth, endMonth time.Time) (*GMBSearchKeywordsResponse, error)
	FetchLocalPosts(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*GMBLocalPostsResponse, error)
	FetchReviews(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*GMBReviewsResponse, error)
	FetchMediaAssets(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*GMBMediaAssetsResponse, error)
}

// TwitterAPI defines the interface for Twitter API v2 operations.
// This interface is used for dependency injection and testing.
type TwitterAPI interface {
	FetchUserTweets(ctx context.Context, userID, oauthToken, oauthTokenSecret string, maxResults int, paginationToken string) (*TwitterTweetsResponse, error)
	FetchUserInfo(ctx context.Context, userIDs []string, oauthToken, oauthTokenSecret string) (*TwitterUserResponse, error)
}

// GeoResolverAPI defines the interface for LinkedIn geo resolution operations.
// This interface is used for dependency injection and testing.
type GeoResolverAPI interface {
	// ResolveGeoIDs resolves LinkedIn geo IDs to human-readable names
	ResolveGeoIDs(ctx context.Context, geoIDs []string, accessToken string) (map[string]string, error)

	// ResolveGeoIDsWithType resolves geo IDs with type information
	ResolveGeoIDsWithType(ctx context.Context, geoIDsWithType []GeoIDWithType, accessToken string) (map[string]string, error)
}

// Verify interfaces are implemented at compile time
var (
	_ FacebookAPI    = (*FacebookClient)(nil)
	_ InstagramAPI   = (*InstagramClient)(nil)
	_ LinkedInAPI    = (*LinkedInClient)(nil)
	_ TikTokAPI      = (*TikTokClient)(nil)
	_ GMBAPI         = (*GMBClient)(nil)
	_ TwitterAPI     = (*TwitterClient)(nil)
	_ GeoResolverAPI = (*GeoResolver)(nil)
)
