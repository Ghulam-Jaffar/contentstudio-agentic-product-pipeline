package social

import (
	"context"
	"encoding/json"
	"time"

	models "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// MockFacebookClient is a mock implementation of FacebookAPI for testing.
type MockFacebookClient struct {
	FetchPostsFunc                     func(ctx context.Context, pageID, accessToken string) ([]kafkamodels.RawFacebookPost, error)
	FetchPostsWithLimitFunc            func(ctx context.Context, pageID, accessToken string, maxPages int) ([]kafkamodels.RawFacebookPost, error)
	FetchPostsSinceFunc                func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookPost, error)
	FetchVideosFunc                    func(ctx context.Context, pageID, accessToken string, maxPages int) ([]kafkamodels.RawFacebookVideo, error)
	FetchVideosWithLimitFunc           func(ctx context.Context, pageID, accessToken string, maxPages int) ([]kafkamodels.RawFacebookVideo, error)
	FetchVideosSinceFunc               func(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookVideo, error)
	FetchInsightsFunc                  func(ctx context.Context, pageID, accessToken string, since, until time.Time) (*kafkamodels.RawFacebookInsights, error)
	GetPostThumbnailsFunc              func(ctx context.Context, pageID, accessToken, longAccessToken, decryptionKey string, posts []clickhouse.MinimalPost) ([]clickhouse.MinimalPost, error)
	GetCompetitorPageDetailsFunc       func(ctx context.Context, pageID, accessToken string) (*models.FacebookPageDetails, *models.Picture, error)
	GetCompetitorPostsFunc             func(ctx context.Context, pageID, accessToken string, since, until time.Time, limit int) ([]*models.Post, string, error)
	GetCompetitorPostsFromURLFunc      func(ctx context.Context, nextURL, pageID, accessToken string) ([]*models.Post, string, error)
	GetCompetitorSharedPostDetailsFunc func(ctx context.Context, parentID, accessToken string) (*models.Post, error)
	GetCompetitorPagePictureFunc       func(ctx context.Context, pageID, accessToken string) (*models.Picture, error)
}

func (m *MockFacebookClient) FetchPosts(ctx context.Context, pageID, accessToken string) ([]kafkamodels.RawFacebookPost, error) {
	if m.FetchPostsFunc != nil {
		return m.FetchPostsFunc(ctx, pageID, accessToken)
	}
	return nil, nil
}

func (m *MockFacebookClient) FetchPostsWithLimit(ctx context.Context, pageID, accessToken string, maxPages int) ([]kafkamodels.RawFacebookPost, error) {
	if m.FetchPostsWithLimitFunc != nil {
		return m.FetchPostsWithLimitFunc(ctx, pageID, accessToken, maxPages)
	}
	return nil, nil
}

func (m *MockFacebookClient) FetchPostsSince(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookPost, error) {
	if m.FetchPostsSinceFunc != nil {
		return m.FetchPostsSinceFunc(ctx, pageID, accessToken, since, until)
	}
	return nil, nil
}

func (m *MockFacebookClient) FetchVideos(ctx context.Context, pageID, accessToken string, maxPages int) ([]kafkamodels.RawFacebookVideo, error) {
	if m.FetchVideosFunc != nil {
		return m.FetchVideosFunc(ctx, pageID, accessToken, maxPages)
	}
	return nil, nil
}

func (m *MockFacebookClient) FetchVideosWithLimit(ctx context.Context, pageID, accessToken string, maxPages int) ([]kafkamodels.RawFacebookVideo, error) {
	if m.FetchVideosWithLimitFunc != nil {
		return m.FetchVideosWithLimitFunc(ctx, pageID, accessToken, maxPages)
	}
	return nil, nil
}

func (m *MockFacebookClient) FetchVideosSince(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookVideo, error) {
	if m.FetchVideosSinceFunc != nil {
		return m.FetchVideosSinceFunc(ctx, pageID, accessToken, since, until)
	}
	return nil, nil
}

func (m *MockFacebookClient) FetchInsights(ctx context.Context, pageID, accessToken string, since, until time.Time) (*kafkamodels.RawFacebookInsights, error) {
	if m.FetchInsightsFunc != nil {
		return m.FetchInsightsFunc(ctx, pageID, accessToken, since, until)
	}
	return nil, nil
}

func (m *MockFacebookClient) GetPostThumbnails(ctx context.Context, pageID, accessToken, longAccessToken, decryptionKey string, posts []clickhouse.MinimalPost) ([]clickhouse.MinimalPost, error) {
	if m.GetPostThumbnailsFunc != nil {
		return m.GetPostThumbnailsFunc(ctx, pageID, accessToken, longAccessToken, decryptionKey, posts)
	}
	return posts, nil
}

func (m *MockFacebookClient) GetCompetitorPageDetails(ctx context.Context, pageID, accessToken string) (*models.FacebookPageDetails, *models.Picture, error) {
	if m.GetCompetitorPageDetailsFunc != nil {
		return m.GetCompetitorPageDetailsFunc(ctx, pageID, accessToken)
	}
	return nil, nil, nil
}

func (m *MockFacebookClient) GetCompetitorPosts(ctx context.Context, pageID, accessToken string, since, until time.Time, limit int) ([]*models.Post, string, error) {
	if m.GetCompetitorPostsFunc != nil {
		return m.GetCompetitorPostsFunc(ctx, pageID, accessToken, since, until, limit)
	}
	return nil, "", nil
}

func (m *MockFacebookClient) GetCompetitorPostsFromURL(ctx context.Context, nextURL, pageID, accessToken string) ([]*models.Post, string, error) {
	if m.GetCompetitorPostsFromURLFunc != nil {
		return m.GetCompetitorPostsFromURLFunc(ctx, nextURL, pageID, accessToken)
	}
	return nil, "", nil
}

func (m *MockFacebookClient) GetCompetitorSharedPostDetails(ctx context.Context, parentID, accessToken string) (*models.Post, error) {
	if m.GetCompetitorSharedPostDetailsFunc != nil {
		return m.GetCompetitorSharedPostDetailsFunc(ctx, parentID, accessToken)
	}
	return nil, nil
}

func (m *MockFacebookClient) GetCompetitorPagePicture(ctx context.Context, pageID, accessToken string) (*models.Picture, error) {
	if m.GetCompetitorPagePictureFunc != nil {
		return m.GetCompetitorPagePictureFunc(ctx, pageID, accessToken)
	}
	return nil, nil
}

// Verify MockFacebookClient implements FacebookAPI
var _ FacebookAPI = (*MockFacebookClient)(nil)

// MockInstagramClient is a mock implementation of InstagramAPI for testing.
type MockInstagramClient struct {
	FetchMediaFunc               func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error)
	FetchMediaSinceFunc          func(ctx context.Context, instagramID, accessToken string, since time.Time) ([]kafkamodels.RawInstagramMedia, error)
	FetchMediaWithLimitFunc      func(ctx context.Context, instagramID, accessToken string, maxPages int) ([]kafkamodels.RawInstagramMedia, error)
	FetchAllMediaFunc            func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error)
	FetchAccountMediaFunc        func(ctx context.Context, instagramID, accessToken string, limit int) (*kafkamodels.RawInstagramAccountResponse, error)
	FetchStoriesFunc             func(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error)
	FetchInsightsFunc            func(ctx context.Context, instagramID, accessToken string, since, until time.Time) (*kafkamodels.RawInstagramInsightsResponse, error)
	FetchInsightsDailyFunc       func(ctx context.Context, instagramID, accessToken string, days, concurrency int) ([]DailyInsight, error)
	FetchMediaInsightsFunc       func(ctx context.Context, mediaID, accessToken, mediaType, mediaProductType string) (*kafkamodels.RawInstagramMediaInsights, error)
	FetchAccountDemographicsFunc func(ctx context.Context, instagramID, accessToken string) (*kafkamodels.RawInstagramDemographics, error)
	FetchUserInfoFunc            func(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error)
	WithBaseURLFunc              func(url string) *InstagramClient
}

func (m *MockInstagramClient) FetchMedia(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
	if m.FetchMediaFunc != nil {
		return m.FetchMediaFunc(ctx, instagramID, accessToken)
	}
	return nil, nil
}

func (m *MockInstagramClient) FetchMediaSince(ctx context.Context, instagramID, accessToken string, since time.Time) ([]kafkamodels.RawInstagramMedia, error) {
	if m.FetchMediaSinceFunc != nil {
		return m.FetchMediaSinceFunc(ctx, instagramID, accessToken, since)
	}
	return nil, nil
}

func (m *MockInstagramClient) FetchMediaWithLimit(ctx context.Context, instagramID, accessToken string, maxPages int) ([]kafkamodels.RawInstagramMedia, error) {
	if m.FetchMediaWithLimitFunc != nil {
		return m.FetchMediaWithLimitFunc(ctx, instagramID, accessToken, maxPages)
	}
	return nil, nil
}

func (m *MockInstagramClient) FetchAllMedia(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
	if m.FetchAllMediaFunc != nil {
		return m.FetchAllMediaFunc(ctx, instagramID, accessToken)
	}
	return nil, nil
}

func (m *MockInstagramClient) FetchAccountMedia(ctx context.Context, instagramID, accessToken string, limit int) (*kafkamodels.RawInstagramAccountResponse, error) {
	if m.FetchAccountMediaFunc != nil {
		return m.FetchAccountMediaFunc(ctx, instagramID, accessToken, limit)
	}
	return nil, nil
}

func (m *MockInstagramClient) FetchStories(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
	if m.FetchStoriesFunc != nil {
		return m.FetchStoriesFunc(ctx, instagramID, accessToken)
	}
	return nil, nil
}

func (m *MockInstagramClient) FetchInsights(ctx context.Context, instagramID, accessToken string, since, until time.Time) (*kafkamodels.RawInstagramInsightsResponse, error) {
	if m.FetchInsightsFunc != nil {
		return m.FetchInsightsFunc(ctx, instagramID, accessToken, since, until)
	}
	return nil, nil
}

func (m *MockInstagramClient) FetchInsightsDaily(ctx context.Context, instagramID, accessToken string, days, concurrency int) ([]DailyInsight, error) {
	if m.FetchInsightsDailyFunc != nil {
		return m.FetchInsightsDailyFunc(ctx, instagramID, accessToken, days, concurrency)
	}
	return nil, nil
}

func (m *MockInstagramClient) FetchMediaInsights(ctx context.Context, mediaID, accessToken, mediaType, mediaProductType string) (*kafkamodels.RawInstagramMediaInsights, error) {
	if m.FetchMediaInsightsFunc != nil {
		return m.FetchMediaInsightsFunc(ctx, mediaID, accessToken, mediaType, mediaProductType)
	}
	return nil, nil
}

func (m *MockInstagramClient) FetchAccountDemographics(ctx context.Context, instagramID, accessToken string) (*kafkamodels.RawInstagramDemographics, error) {
	if m.FetchAccountDemographicsFunc != nil {
		return m.FetchAccountDemographicsFunc(ctx, instagramID, accessToken)
	}
	return nil, nil
}

func (m *MockInstagramClient) FetchUserInfo(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
	if m.FetchUserInfoFunc != nil {
		return m.FetchUserInfoFunc(ctx, instagramID, accessToken)
	}
	return nil, nil
}

func (m *MockInstagramClient) WithBaseURL(url string) *InstagramClient {
	if m.WithBaseURLFunc != nil {
		return m.WithBaseURLFunc(url)
	}
	return nil
}

// Verify MockInstagramClient implements InstagramAPI
var _ InstagramAPI = (*MockInstagramClient)(nil)

// MockLinkedInClient is a mock implementation of LinkedInAPI for testing.
type MockLinkedInClient struct {
	FetchSharesFunc                           func(ctx context.Context, organisationID, accessToken string) ([]json.RawMessage, error)
	FetchPostsPaginatedFunc                   func(ctx context.Context, linkedinID string, entityType string, accessToken string, cutoffTime time.Time) ([]json.RawMessage, error)
	FetchImagesRawFunc                        func(ctx context.Context, ids []string, accessToken string) ([]byte, error)
	FetchVideosRawFunc                        func(ctx context.Context, ids []string, accessToken string) ([]byte, error)
	FetchDocumentsRawFunc                     func(ctx context.Context, ids []string, accessToken string) ([]byte, error)
	FetchStatsRawFunc                         func(ctx context.Context, linkedinID string, ugcPosts []string, shares []string, accessToken string) ([]byte, error)
	FetchFollowerDataFunc                     func(ctx context.Context, linkedinID string, accessToken string) ([]byte, error)
	FetchFollowerDataWithGeoNamesFunc         func(ctx context.Context, linkedinID string, accessToken string, geoNames map[string]string) ([]byte, error)
	FetchFollowerStatsWithGeoIDsFunc          func(ctx context.Context, linkedinID string, accessToken string) (*FollowerStatsWithGeoIDs, error)
	BuildFollowerDataWithGeoNamesFunc         func(stats *FollowerStatsWithGeoIDs, geoNames map[string]string) ([]byte, error)
	GetGeoIDsFromFollowerStatsRawFunc         func(ctx context.Context, linkedinID string, accessToken string) ([]string, error)
	GetGeoIDsWithTypeFromFollowerStatsRawFunc func(ctx context.Context, linkedinID string, accessToken string) ([]GeoIDWithType, error)
	ResolveGeoIDsFunc                         func(ctx context.Context, geoIDs []string, accessToken string) (map[string]string, error)
	FetchOrganizationDetailsRawFunc           func(ctx context.Context, linkedinID string, accessToken string) ([]byte, error)
	FetchPageStatisticsRawFunc                func(ctx context.Context, linkedinID string, accessToken string, startMs, endMs int64) ([]byte, error)
	FetchShareStatisticsRawFunc               func(ctx context.Context, linkedinID string, accessToken string, startMs, endMs int64) ([]byte, error)
	FetchMemberCreatorPostAnalyticsRawFunc    func(ctx context.Context, accessToken string, queryType string, startDate, endDate *time.Time) ([]byte, error)
	FetchMemberFollowersCountRawFunc          func(ctx context.Context, accessToken string, startDate, endDate *time.Time) ([]byte, error)
}

func (m *MockLinkedInClient) FetchShares(ctx context.Context, organisationID, accessToken string) ([]json.RawMessage, error) {
	if m.FetchSharesFunc != nil {
		return m.FetchSharesFunc(ctx, organisationID, accessToken)
	}
	return nil, nil
}

func (m *MockLinkedInClient) FetchPostsPaginated(ctx context.Context, linkedinID string, entityType string, accessToken string, cutoffTime time.Time) ([]json.RawMessage, error) {
	if m.FetchPostsPaginatedFunc != nil {
		return m.FetchPostsPaginatedFunc(ctx, linkedinID, entityType, accessToken, cutoffTime)
	}
	return nil, nil
}

func (m *MockLinkedInClient) FetchImagesRaw(ctx context.Context, ids []string, accessToken string) ([]byte, error) {
	if m.FetchImagesRawFunc != nil {
		return m.FetchImagesRawFunc(ctx, ids, accessToken)
	}
	return nil, nil
}

func (m *MockLinkedInClient) FetchVideosRaw(ctx context.Context, ids []string, accessToken string) ([]byte, error) {
	if m.FetchVideosRawFunc != nil {
		return m.FetchVideosRawFunc(ctx, ids, accessToken)
	}
	return nil, nil
}

func (m *MockLinkedInClient) FetchDocumentsRaw(ctx context.Context, ids []string, accessToken string) ([]byte, error) {
	if m.FetchDocumentsRawFunc != nil {
		return m.FetchDocumentsRawFunc(ctx, ids, accessToken)
	}
	return nil, nil
}

func (m *MockLinkedInClient) FetchStatsRaw(ctx context.Context, linkedinID string, ugcPosts []string, shares []string, accessToken string) ([]byte, error) {
	if m.FetchStatsRawFunc != nil {
		return m.FetchStatsRawFunc(ctx, linkedinID, ugcPosts, shares, accessToken)
	}
	return nil, nil
}

func (m *MockLinkedInClient) FetchFollowerData(ctx context.Context, linkedinID string, accessToken string) ([]byte, error) {
	if m.FetchFollowerDataFunc != nil {
		return m.FetchFollowerDataFunc(ctx, linkedinID, accessToken)
	}
	return nil, nil
}

func (m *MockLinkedInClient) FetchFollowerDataWithGeoNames(ctx context.Context, linkedinID string, accessToken string, geoNames map[string]string) ([]byte, error) {
	if m.FetchFollowerDataWithGeoNamesFunc != nil {
		return m.FetchFollowerDataWithGeoNamesFunc(ctx, linkedinID, accessToken, geoNames)
	}
	return nil, nil
}

func (m *MockLinkedInClient) FetchFollowerStatsWithGeoIDs(ctx context.Context, linkedinID string, accessToken string) (*FollowerStatsWithGeoIDs, error) {
	if m.FetchFollowerStatsWithGeoIDsFunc != nil {
		return m.FetchFollowerStatsWithGeoIDsFunc(ctx, linkedinID, accessToken)
	}
	return nil, nil
}

func (m *MockLinkedInClient) BuildFollowerDataWithGeoNames(stats *FollowerStatsWithGeoIDs, geoNames map[string]string) ([]byte, error) {
	if m.BuildFollowerDataWithGeoNamesFunc != nil {
		return m.BuildFollowerDataWithGeoNamesFunc(stats, geoNames)
	}
	return nil, nil
}

func (m *MockLinkedInClient) GetGeoIDsFromFollowerStatsRaw(ctx context.Context, linkedinID string, accessToken string) ([]string, error) {
	if m.GetGeoIDsFromFollowerStatsRawFunc != nil {
		return m.GetGeoIDsFromFollowerStatsRawFunc(ctx, linkedinID, accessToken)
	}
	return nil, nil
}

func (m *MockLinkedInClient) GetGeoIDsWithTypeFromFollowerStatsRaw(ctx context.Context, linkedinID string, accessToken string) ([]GeoIDWithType, error) {
	if m.GetGeoIDsWithTypeFromFollowerStatsRawFunc != nil {
		return m.GetGeoIDsWithTypeFromFollowerStatsRawFunc(ctx, linkedinID, accessToken)
	}
	return nil, nil
}

func (m *MockLinkedInClient) ResolveGeoIDs(ctx context.Context, geoIDs []string, accessToken string) (map[string]string, error) {
	if m.ResolveGeoIDsFunc != nil {
		return m.ResolveGeoIDsFunc(ctx, geoIDs, accessToken)
	}
	return nil, nil
}

func (m *MockLinkedInClient) FetchOrganizationDetailsRaw(ctx context.Context, linkedinID string, accessToken string) ([]byte, error) {
	if m.FetchOrganizationDetailsRawFunc != nil {
		return m.FetchOrganizationDetailsRawFunc(ctx, linkedinID, accessToken)
	}
	return nil, nil
}

func (m *MockLinkedInClient) FetchPageStatisticsRaw(ctx context.Context, linkedinID string, accessToken string, startMs, endMs int64) ([]byte, error) {
	if m.FetchPageStatisticsRawFunc != nil {
		return m.FetchPageStatisticsRawFunc(ctx, linkedinID, accessToken, startMs, endMs)
	}
	return nil, nil
}

func (m *MockLinkedInClient) FetchShareStatisticsRaw(ctx context.Context, linkedinID string, accessToken string, startMs, endMs int64) ([]byte, error) {
	if m.FetchShareStatisticsRawFunc != nil {
		return m.FetchShareStatisticsRawFunc(ctx, linkedinID, accessToken, startMs, endMs)
	}
	return nil, nil
}

func (m *MockLinkedInClient) FetchMemberCreatorPostAnalyticsRaw(ctx context.Context, accessToken string, queryType string, startDate, endDate *time.Time) ([]byte, error) {
	if m.FetchMemberCreatorPostAnalyticsRawFunc != nil {
		return m.FetchMemberCreatorPostAnalyticsRawFunc(ctx, accessToken, queryType, startDate, endDate)
	}
	return nil, nil
}

func (m *MockLinkedInClient) FetchMemberFollowersCountRaw(ctx context.Context, accessToken string, startDate, endDate *time.Time) ([]byte, error) {
	if m.FetchMemberFollowersCountRawFunc != nil {
		return m.FetchMemberFollowersCountRawFunc(ctx, accessToken, startDate, endDate)
	}
	return nil, nil
}

// Verify MockLinkedInClient implements LinkedInAPI
var _ LinkedInAPI = (*MockLinkedInClient)(nil)

// MockTikTokClient is a mock implementation of TikTokAPI for testing.
type MockTikTokClient struct {
	FetchUserVideosFunc func(ctx context.Context, userID, accessToken string, cursor, maxCount int) (json.RawMessage, int64, error)
	FetchUserInfoFunc   func(ctx context.Context, accessToken string) (json.RawMessage, error)
	FetchVideoListFunc  func(ctx context.Context, accessToken string, cursor int64, maxCount int) (json.RawMessage, int64, bool, error)
	RefreshTokenFunc    func(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error)
}

func (m *MockTikTokClient) FetchUserVideos(ctx context.Context, userID, accessToken string, cursor, maxCount int) (json.RawMessage, int64, error) {
	if m.FetchUserVideosFunc != nil {
		return m.FetchUserVideosFunc(ctx, userID, accessToken, cursor, maxCount)
	}
	return nil, 0, nil
}

func (m *MockTikTokClient) FetchUserInfo(ctx context.Context, accessToken string) (json.RawMessage, error) {
	if m.FetchUserInfoFunc != nil {
		return m.FetchUserInfoFunc(ctx, accessToken)
	}
	return nil, nil
}

func (m *MockTikTokClient) FetchVideoList(ctx context.Context, accessToken string, cursor int64, maxCount int) (json.RawMessage, int64, bool, error) {
	if m.FetchVideoListFunc != nil {
		return m.FetchVideoListFunc(ctx, accessToken, cursor, maxCount)
	}
	return nil, 0, false, nil
}

func (m *MockTikTokClient) RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error) {
	if m.RefreshTokenFunc != nil {
		return m.RefreshTokenFunc(ctx, refreshToken)
	}
	return nil, nil
}

// Verify MockTikTokClient implements TikTokAPI
var _ TikTokAPI = (*MockTikTokClient)(nil)

// MockGMBClient is a mock implementation of GMBAPI for testing.
type MockGMBClient struct {
	RefreshTokenFunc            func(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error)
	FetchVoiceOfMerchantFunc    func(ctx context.Context, locationID, accessToken string) (*VoiceOfMerchantResponse, error)
	FetchPerformanceMetricsFunc func(ctx context.Context, locationID, accessToken string, startDate, endDate time.Time) (*GMBPerformanceResponse, error)
	FetchSearchKeywordsFunc     func(ctx context.Context, locationID, accessToken string, startMonth, endMonth time.Time) (*GMBSearchKeywordsResponse, error)
	FetchLocalPostsFunc         func(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*GMBLocalPostsResponse, error)
	FetchReviewsFunc            func(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*GMBReviewsResponse, error)
	FetchMediaAssetsFunc        func(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*GMBMediaAssetsResponse, error)
}

func (m *MockGMBClient) RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error) {
	if m.RefreshTokenFunc != nil {
		return m.RefreshTokenFunc(ctx, refreshToken)
	}
	return &RefreshTokenResponse{AccessToken: "mock_access_token"}, nil
}

func (m *MockGMBClient) FetchVoiceOfMerchant(ctx context.Context, locationID, accessToken string) (*VoiceOfMerchantResponse, error) {
	if m.FetchVoiceOfMerchantFunc != nil {
		return m.FetchVoiceOfMerchantFunc(ctx, locationID, accessToken)
	}
	return &VoiceOfMerchantResponse{HasVoiceOfMerchant: true, HasBusinessAuthority: true}, nil
}

func (m *MockGMBClient) FetchPerformanceMetrics(ctx context.Context, locationID, accessToken string, startDate, endDate time.Time) (*GMBPerformanceResponse, error) {
	if m.FetchPerformanceMetricsFunc != nil {
		return m.FetchPerformanceMetricsFunc(ctx, locationID, accessToken, startDate, endDate)
	}
	return &GMBPerformanceResponse{}, nil
}

func (m *MockGMBClient) FetchSearchKeywords(ctx context.Context, locationID, accessToken string, startMonth, endMonth time.Time) (*GMBSearchKeywordsResponse, error) {
	if m.FetchSearchKeywordsFunc != nil {
		return m.FetchSearchKeywordsFunc(ctx, locationID, accessToken, startMonth, endMonth)
	}
	return &GMBSearchKeywordsResponse{}, nil
}

func (m *MockGMBClient) FetchLocalPosts(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*GMBLocalPostsResponse, error) {
	if m.FetchLocalPostsFunc != nil {
		return m.FetchLocalPostsFunc(ctx, accountID, locationID, accessToken, pageToken)
	}
	return &GMBLocalPostsResponse{}, nil
}

func (m *MockGMBClient) FetchReviews(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*GMBReviewsResponse, error) {
	if m.FetchReviewsFunc != nil {
		return m.FetchReviewsFunc(ctx, accountID, locationID, accessToken, pageToken)
	}
	return &GMBReviewsResponse{}, nil
}

func (m *MockGMBClient) FetchMediaAssets(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*GMBMediaAssetsResponse, error) {
	if m.FetchMediaAssetsFunc != nil {
		return m.FetchMediaAssetsFunc(ctx, accountID, locationID, accessToken, pageToken)
	}
	return &GMBMediaAssetsResponse{}, nil
}

// Verify MockGMBClient implements GMBAPI
var _ GMBAPI = (*MockGMBClient)(nil)

// MockTwitterClient is a mock implementation of TwitterAPI for testing.
type MockTwitterClient struct {
	FetchUserTweetsFunc func(ctx context.Context, userID, oauthToken, oauthTokenSecret string, maxResults int, paginationToken string) (*TwitterTweetsResponse, error)
	FetchUserInfoFunc   func(ctx context.Context, userIDs []string, oauthToken, oauthTokenSecret string) (*TwitterUserResponse, error)
}

func (m *MockTwitterClient) FetchUserTweets(ctx context.Context, userID, oauthToken, oauthTokenSecret string, maxResults int, paginationToken string) (*TwitterTweetsResponse, error) {
	if m.FetchUserTweetsFunc != nil {
		return m.FetchUserTweetsFunc(ctx, userID, oauthToken, oauthTokenSecret, maxResults, paginationToken)
	}
	return &TwitterTweetsResponse{}, nil
}

func (m *MockTwitterClient) FetchUserInfo(ctx context.Context, userIDs []string, oauthToken, oauthTokenSecret string) (*TwitterUserResponse, error) {
	if m.FetchUserInfoFunc != nil {
		return m.FetchUserInfoFunc(ctx, userIDs, oauthToken, oauthTokenSecret)
	}
	return &TwitterUserResponse{}, nil
}

// Verify MockTwitterClient implements TwitterAPI
var _ TwitterAPI = (*MockTwitterClient)(nil)

// MockGeoResolver is a mock implementation of GeoResolverAPI for testing.
type MockGeoResolver struct {
	ResolveGeoIDsFunc         func(ctx context.Context, geoIDs []string, accessToken string) (map[string]string, error)
	ResolveGeoIDsWithTypeFunc func(ctx context.Context, geoIDsWithType []GeoIDWithType, accessToken string) (map[string]string, error)
}

func (m *MockGeoResolver) ResolveGeoIDs(ctx context.Context, geoIDs []string, accessToken string) (map[string]string, error) {
	if m.ResolveGeoIDsFunc != nil {
		return m.ResolveGeoIDsFunc(ctx, geoIDs, accessToken)
	}
	return map[string]string{}, nil
}

func (m *MockGeoResolver) ResolveGeoIDsWithType(ctx context.Context, geoIDsWithType []GeoIDWithType, accessToken string) (map[string]string, error) {
	if m.ResolveGeoIDsWithTypeFunc != nil {
		return m.ResolveGeoIDsWithTypeFunc(ctx, geoIDsWithType, accessToken)
	}
	return map[string]string{}, nil
}

// Verify MockGeoResolver implements GeoResolverAPI
var _ GeoResolverAPI = (*MockGeoResolver)(nil)

// MockYouTubeClient is a mock implementation of YouTubeClient for testing.
type MockYouTubeClient struct {
	RefreshTokenFunc            func(ctx context.Context, refreshToken string) (*YouTubeTokenResponse, error)
	FetchChannelsFunc           func(ctx context.Context, accessToken string) (*YouTubeChannelResponse, error)
	FetchVideosFunc             func(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]YouTubeActivityItem, error)
	FetchVideoDetailsFunc       func(ctx context.Context, accessToken string, videoIDs []string) ([]YouTubeVideoItem, error)
	FetchActivityInsightsFunc   func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*YouTubeAnalyticsResponse, error)
	FetchTrafficInsightsFunc    func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*YouTubeAnalyticsResponse, error)
	FetchSharedInsightsFunc     func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*YouTubeAnalyticsResponse, error)
	FetchVideoInsightsFunc      func(ctx context.Context, accessToken, videoID string, startDate, endDate time.Time) (*YouTubeAnalyticsResponse, error)
	FetchAllVideosAnalyticsFunc func(ctx context.Context, accessToken string, startDate, endDate time.Time) (map[string]*VideoAnalytics, error)
	DetectMediaTypesFunc        func(ctx context.Context, videos []YouTubeVideoItem) map[string]string
	IsShortFunc                 func(ctx context.Context, videoID string) bool
}

func (m *MockYouTubeClient) RefreshToken(ctx context.Context, refreshToken string) (*YouTubeTokenResponse, error) {
	if m.RefreshTokenFunc != nil {
		return m.RefreshTokenFunc(ctx, refreshToken)
	}
	return &YouTubeTokenResponse{AccessToken: "new_access_token"}, nil
}

func (m *MockYouTubeClient) FetchChannels(ctx context.Context, accessToken string) (*YouTubeChannelResponse, error) {
	if m.FetchChannelsFunc != nil {
		return m.FetchChannelsFunc(ctx, accessToken)
	}
	return &YouTubeChannelResponse{}, nil
}

func (m *MockYouTubeClient) FetchVideos(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]YouTubeActivityItem, error) {
	if m.FetchVideosFunc != nil {
		return m.FetchVideosFunc(ctx, accessToken, uploadsPlaylistID, since)
	}
	return nil, nil
}

func (m *MockYouTubeClient) FetchVideoDetails(ctx context.Context, accessToken string, videoIDs []string) ([]YouTubeVideoItem, error) {
	if m.FetchVideoDetailsFunc != nil {
		return m.FetchVideoDetailsFunc(ctx, accessToken, videoIDs)
	}
	return nil, nil
}

func (m *MockYouTubeClient) FetchActivityInsights(ctx context.Context, accessToken string, startDate, endDate time.Time) (*YouTubeAnalyticsResponse, error) {
	if m.FetchActivityInsightsFunc != nil {
		return m.FetchActivityInsightsFunc(ctx, accessToken, startDate, endDate)
	}
	return &YouTubeAnalyticsResponse{}, nil
}

func (m *MockYouTubeClient) FetchTrafficInsights(ctx context.Context, accessToken string, startDate, endDate time.Time) (*YouTubeAnalyticsResponse, error) {
	if m.FetchTrafficInsightsFunc != nil {
		return m.FetchTrafficInsightsFunc(ctx, accessToken, startDate, endDate)
	}
	return &YouTubeAnalyticsResponse{}, nil
}

func (m *MockYouTubeClient) FetchSharedInsights(ctx context.Context, accessToken string, startDate, endDate time.Time) (*YouTubeAnalyticsResponse, error) {
	if m.FetchSharedInsightsFunc != nil {
		return m.FetchSharedInsightsFunc(ctx, accessToken, startDate, endDate)
	}
	return &YouTubeAnalyticsResponse{}, nil
}

func (m *MockYouTubeClient) FetchVideoInsights(ctx context.Context, accessToken, videoID string, startDate, endDate time.Time) (*YouTubeAnalyticsResponse, error) {
	if m.FetchVideoInsightsFunc != nil {
		return m.FetchVideoInsightsFunc(ctx, accessToken, videoID, startDate, endDate)
	}
	return &YouTubeAnalyticsResponse{}, nil
}

func (m *MockYouTubeClient) FetchAllVideosAnalytics(ctx context.Context, accessToken string, startDate, endDate time.Time) (map[string]*VideoAnalytics, error) {
	if m.FetchAllVideosAnalyticsFunc != nil {
		return m.FetchAllVideosAnalyticsFunc(ctx, accessToken, startDate, endDate)
	}
	return map[string]*VideoAnalytics{}, nil
}

func (m *MockYouTubeClient) DetectMediaTypes(ctx context.Context, videos []YouTubeVideoItem) map[string]string {
	if m.DetectMediaTypesFunc != nil {
		return m.DetectMediaTypesFunc(ctx, videos)
	}
	result := make(map[string]string)
	for _, v := range videos {
		result[v.ID] = "video"
	}
	return result
}

func (m *MockYouTubeClient) IsYouTubeShort(ctx context.Context, videoID string) bool {
	if m.IsShortFunc != nil {
		return m.IsShortFunc(ctx, videoID)
	}
	return false
}

// Verify MockYouTubeClient implements YouTubeAPI
var _ YouTubeAPI = (*MockYouTubeClient)(nil)

// MockPinterestClient is a mock implementation of PinterestAPI for testing.
type MockPinterestClient struct {
	GetUserAccountFunc          func(ctx context.Context, accessToken string) (*PinterestUserAccount, error)
	GetUserAccountAnalyticsFunc func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*PinterestUserAnalyticsResponse, error)
	GetBoardsFunc               func(ctx context.Context, accessToken string) (*PinterestBoardsResponse, error)
	GetBoardFunc                func(ctx context.Context, accessToken, boardID string) (*PinterestBoard, error)
	GetBoardPinsFunc            func(ctx context.Context, accessToken, boardID string, pageSize int, bookmark string) (*PinterestPinsResponse, error)
	GetUserPinsFunc             func(ctx context.Context, accessToken string, pageSize int, bookmark string) (*PinterestPinsResponse, error)
	GetPinAnalyticsFunc         func(ctx context.Context, accessToken, pinID string, startDate, endDate time.Time) (*PinterestPinAnalyticsResponse, error)
	GetMultiPinAnalyticsFunc    func(ctx context.Context, accessToken string, pinIDs []string, startDate, endDate time.Time) (map[string]*PinterestPinAnalyticsResponse, error)
}

func (m *MockPinterestClient) GetUserAccount(ctx context.Context, accessToken string) (*PinterestUserAccount, error) {
	if m.GetUserAccountFunc != nil {
		return m.GetUserAccountFunc(ctx, accessToken)
	}
	return &PinterestUserAccount{}, nil
}

func (m *MockPinterestClient) GetUserAccountAnalytics(ctx context.Context, accessToken string, startDate, endDate time.Time) (*PinterestUserAnalyticsResponse, error) {
	if m.GetUserAccountAnalyticsFunc != nil {
		return m.GetUserAccountAnalyticsFunc(ctx, accessToken, startDate, endDate)
	}
	return &PinterestUserAnalyticsResponse{}, nil
}

func (m *MockPinterestClient) GetBoards(ctx context.Context, accessToken string) (*PinterestBoardsResponse, error) {
	if m.GetBoardsFunc != nil {
		return m.GetBoardsFunc(ctx, accessToken)
	}
	return &PinterestBoardsResponse{}, nil
}

func (m *MockPinterestClient) GetBoard(ctx context.Context, accessToken, boardID string) (*PinterestBoard, error) {
	if m.GetBoardFunc != nil {
		return m.GetBoardFunc(ctx, accessToken, boardID)
	}
	return &PinterestBoard{}, nil
}

func (m *MockPinterestClient) GetBoardPins(ctx context.Context, accessToken, boardID string, pageSize int, bookmark string) (*PinterestPinsResponse, error) {
	if m.GetBoardPinsFunc != nil {
		return m.GetBoardPinsFunc(ctx, accessToken, boardID, pageSize, bookmark)
	}
	return &PinterestPinsResponse{}, nil
}

func (m *MockPinterestClient) GetUserPins(ctx context.Context, accessToken string, pageSize int, bookmark string) (*PinterestPinsResponse, error) {
	if m.GetUserPinsFunc != nil {
		return m.GetUserPinsFunc(ctx, accessToken, pageSize, bookmark)
	}
	return &PinterestPinsResponse{}, nil
}

func (m *MockPinterestClient) GetPinAnalytics(ctx context.Context, accessToken, pinID string, startDate, endDate time.Time) (*PinterestPinAnalyticsResponse, error) {
	if m.GetPinAnalyticsFunc != nil {
		return m.GetPinAnalyticsFunc(ctx, accessToken, pinID, startDate, endDate)
	}
	return &PinterestPinAnalyticsResponse{}, nil
}

func (m *MockPinterestClient) GetMultiPinAnalytics(ctx context.Context, accessToken string, pinIDs []string, startDate, endDate time.Time) (map[string]*PinterestPinAnalyticsResponse, error) {
	if m.GetMultiPinAnalyticsFunc != nil {
		return m.GetMultiPinAnalyticsFunc(ctx, accessToken, pinIDs, startDate, endDate)
	}
	return map[string]*PinterestPinAnalyticsResponse{}, nil
}

// Verify MockPinterestClient implements PinterestAPI
var _ PinterestAPI = (*MockPinterestClient)(nil)
