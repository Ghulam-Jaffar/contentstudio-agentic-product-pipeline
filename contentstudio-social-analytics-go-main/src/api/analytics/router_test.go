package analytics

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	campaignLabelAPI "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/campaign_label"
	facebookAPI "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/facebook"
	fbCompetitorAPI "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/fb_competitor"
	igCompetitorAPI "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/ig_competitor"
	instagramAPI "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/instagram"
	lookerStudioAPI "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/looker_studio"
	overviewAPI "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/overview"
	pinterestAPI "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/pinterest"
	youtubeAPI "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/youtube"
	"github.com/rs/zerolog"

	gmbAPI "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/gmb"
	linkedin "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/linkedin"
	tiktokAPI "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/tiktok"
	twitterAPI "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/twitter"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	campaignLabelTypes "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/campaign_label"
	facebookTypes "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/facebook"
	fbCompetitorTypes "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/fb_competitor"
	gmbTypes "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/gmb"
	igCompetitorTypes "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/ig_competitor"
	instagramTypes "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/instagram"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/linkedin"
	overviewTypes "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/overview"
	pinterestTypes "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/pinterest"
	tiktokTypes "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/tiktok"
	twitterTypes "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/twitter"
	youtubeTypes "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/youtube"
	mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	campaignLabelSvc "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/campaign_label"
	facebookSvc "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/facebook"
	fbCompetitorSvc "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/fb_competitor"
	gmbSvc "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/gmb"
	igCompetitorSvc "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/ig_competitor"
	instagramSvc "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/instagram"
	service "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/linkedin"
	overviewSvc "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/overview"
	pinterestSvc "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/pinterest"
	tiktokSvc "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/tiktok"
	twitterSvc "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/twitter"
	youtubeSvc "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/youtube"
)

// --- Mock service ---

type mockService struct{}

var _ service.Service = (*mockService)(nil)

func (m *mockService) GetSummary(_ context.Context, _ *types.LinkedInRequest) (*types.SummaryResponse, error) {
	return &types.SummaryResponse{Status: true, Overview: map[string]*types.SummaryMetrics{"current": {}, "previous": {}}}, nil
}
func (m *mockService) GetAudienceGrowth(_ context.Context, _ *types.LinkedInRequest) (*types.AudienceGrowthResponse, error) {
	return &types.AudienceGrowthResponse{Status: true, AudienceGrowth: &types.AudienceGrowthData{}, AudienceGrowthRollup: map[string]*types.AudienceGrowthRollup{"current": {}, "previous": {}}}, nil
}
func (m *mockService) GetPageViews(_ context.Context, _ *types.LinkedInRequest) (*types.PageViewsResponse, error) {
	return &types.PageViewsResponse{Status: true, PageViews: &types.PageViewsData{}, PageViewsRollup: map[string]*types.PageViewsRollup{"current": {}, "previous": {}}}, nil
}
func (m *mockService) GetPublishingBehaviour(_ context.Context, _ *types.PublishingBehaviourRequest) (*types.PublishingBehaviourResponse, error) {
	return &types.PublishingBehaviourResponse{Status: true, PublishingBehaviour: &types.PublishingBehaviourData{}, PublishingBehaviourRollup: map[string][]types.PublishingBehaviourMediaType{"current": {}, "previous": {}}}, nil
}
func (m *mockService) GetTopPosts(_ context.Context, _ *types.TopPostsRequest) (*types.TopPostsResponse, error) {
	return &types.TopPostsResponse{Status: true, TopPosts: []types.TopPost{}}, nil
}
func (m *mockService) GetPostsPerDay(_ context.Context, _ *types.LinkedInRequest) (*types.PostsPerDayResponse, error) {
	return &types.PostsPerDayResponse{Status: true, PostsPerDays: &types.PostsPerDayData{Data: types.PostsPerDayInner{Days: map[string]int32{}}}}, nil
}
func (m *mockService) GetHashtags(_ context.Context, _ *types.LinkedInRequest) (*types.HashtagsResponse, error) {
	return &types.HashtagsResponse{Status: true, TopHashtags: &types.HashtagsData{}, TopHashtagsRollup: map[string]*types.HashtagsRollup{"current": {}, "previous": {}}}, nil
}
func (m *mockService) GetFollowersDemographics(_ context.Context, _ *types.LinkedInRequest) (*types.DemographicsResponse, error) {
	return &types.DemographicsResponse{Status: true, FollowerDemographics: map[string]*types.DemographicCategory{}}, nil
}

func newTestHandler() *linkedin.LinkedInHandler {
	return linkedin.NewLinkedInHandler(&mockService{}, zerolog.New(io.Discard))
}

// --- GMB Mock service ---

type mockGMBService struct{}

var _ gmbSvc.Service = (*mockGMBService)(nil)

func (m *mockGMBService) GetSummary(_ context.Context, _ *gmbTypes.GMBRequest) (*gmbTypes.SummaryResponse, error) {
	return &gmbTypes.SummaryResponse{Status: true, Overview: map[string]*gmbTypes.SummaryMetrics{"current": {}, "previous": {}}}, nil
}
func (m *mockGMBService) GetImpressions(_ context.Context, _ *gmbTypes.GMBRequest) (*gmbTypes.ImpressionsResponse, error) {
	return &gmbTypes.ImpressionsResponse{Status: true, Impressions: &gmbTypes.ImpressionsData{}}, nil
}
func (m *mockGMBService) GetActions(_ context.Context, _ *gmbTypes.GMBRequest) (*gmbTypes.ActionsResponse, error) {
	return &gmbTypes.ActionsResponse{Status: true, Actions: &gmbTypes.ActionsData{}}, nil
}
func (m *mockGMBService) GetSearchKeywords(_ context.Context, _ *gmbTypes.SearchKeywordsRequest) (*gmbTypes.SearchKeywordsResponse, error) {
	return &gmbTypes.SearchKeywordsResponse{Status: true, Keywords: []gmbTypes.SearchKeyword{}}, nil
}
func (m *mockGMBService) GetTopPosts(_ context.Context, _ *gmbTypes.TopPostsRequest) (*gmbTypes.TopPostsResponse, error) {
	return &gmbTypes.TopPostsResponse{Status: true, Posts: []gmbTypes.TopPost{}}, nil
}
func (m *mockGMBService) GetPublishingBehavior(_ context.Context, _ *gmbTypes.GMBRequest) (*gmbTypes.PublishingBehaviorResponse, error) {
	return &gmbTypes.PublishingBehaviorResponse{Status: true, PublishingBehaviour: &gmbTypes.PublishingBehaviorData{}}, nil
}
func (m *mockGMBService) GetReviews(_ context.Context, _ *gmbTypes.GMBRequest) (*gmbTypes.ReviewsResponse, error) {
	return &gmbTypes.ReviewsResponse{Status: true, Reviews: &gmbTypes.ReviewsData{}}, nil
}
func (m *mockGMBService) GetMediaActivity(_ context.Context, _ *gmbTypes.GMBRequest) (*gmbTypes.MediaActivityResponse, error) {
	return &gmbTypes.MediaActivityResponse{Status: true, MediaActivity: &gmbTypes.MediaActivityData{}}, nil
}

func newTestGMBHandler() *gmbAPI.GMBHandler {
	return gmbAPI.NewGMBHandler(&mockGMBService{}, zerolog.New(io.Discard))
}

type mockFacebookService struct{}

var _ facebookSvc.Service = (*mockFacebookService)(nil)

func (m *mockFacebookService) GetSummary(_ context.Context, _ *facebookTypes.FacebookRequest) (*facebookTypes.SummaryResponse, error) {
	return &facebookTypes.SummaryResponse{Status: true, Overview: map[string]*facebookTypes.SummaryMetrics{"current": {}, "previous": {}}}, nil
}
func (m *mockFacebookService) GetAudienceGrowth(_ context.Context, _ *facebookTypes.FacebookRequest) (*facebookTypes.AudienceGrowthResponse, error) {
	return &facebookTypes.AudienceGrowthResponse{Status: true, AudienceGrowth: &facebookTypes.AudienceGrowthData{}, AudienceGrowthRollup: map[string]*facebookTypes.AudienceGrowthRollup{"current": {}, "previous": {}}}, nil
}
func (m *mockFacebookService) GetPublishingBehaviour(_ context.Context, _ *facebookTypes.FacebookRequest) (*facebookTypes.PublishingBehaviourResponse, error) {
	return &facebookTypes.PublishingBehaviourResponse{Status: true, PublishingBehaviour: &facebookTypes.PublishingBehaviourData{}, PublishingBehaviourRollup: map[string]*facebookTypes.PublishingRollup{"current": {}, "previous": {}}}, nil
}
func (m *mockFacebookService) GetTopPosts(_ context.Context, _ *facebookTypes.FacebookRequest) (*facebookTypes.TopPostsResponse, error) {
	return &facebookTypes.TopPostsResponse{Status: true, TopPosts: []facebookTypes.TopPost{}}, nil
}
func (m *mockFacebookService) GetActiveUsers(_ context.Context, _ *facebookTypes.FacebookRequest) (*facebookTypes.ActiveUsersResponse, error) {
	return &facebookTypes.ActiveUsersResponse{Status: true, ActiveUsers: &facebookTypes.ActiveUsersData{}}, nil
}
func (m *mockFacebookService) GetImpressions(_ context.Context, _ *facebookTypes.FacebookRequest) (*facebookTypes.ImpressionsResponse, error) {
	return &facebookTypes.ImpressionsResponse{Status: true, Impressions: &facebookTypes.ImpressionsData{}, ImpressionsRollup: map[string]*facebookTypes.ImpressionsRollup{"current": {}, "previous": {}}}, nil
}
func (m *mockFacebookService) GetEngagement(_ context.Context, _ *facebookTypes.FacebookRequest) (*facebookTypes.EngagementResponse, error) {
	return &facebookTypes.EngagementResponse{Status: true, Engagement: &facebookTypes.EngagementContainer{Engagement: &facebookTypes.EngagementData{}, EngagementRollup: map[string]*facebookTypes.EngagementRollup{"current": {}, "previous": {}}}}, nil
}
func (m *mockFacebookService) GetReelsAnalytics(_ context.Context, _ *facebookTypes.FacebookRequest) (*facebookTypes.ReelsAnalyticsResponse, error) {
	return &facebookTypes.ReelsAnalyticsResponse{Status: true, Reels: &facebookTypes.ReelsData{}, ReelsRollup: map[string]*facebookTypes.ReelsRollup{"current": {}, "previous": {}}}, nil
}
func (m *mockFacebookService) GetVideoInsights(_ context.Context, _ *facebookTypes.FacebookRequest) (*facebookTypes.VideoInsightsResponse, error) {
	return &facebookTypes.VideoInsightsResponse{Status: true, VideoInsights: &facebookTypes.VideoInsightsData{}, VideoRollup: map[string]*facebookTypes.VideoRollup{"current": {}, "previous": {}}}, nil
}
func (m *mockFacebookService) GetDemographics(_ context.Context, _ *facebookTypes.FacebookRequest) (*facebookTypes.DemographicsResponse, error) {
	return &facebookTypes.DemographicsResponse{Status: true}, nil
}
func (m *mockFacebookService) GetOverviewDemographics(_ context.Context, _ *facebookTypes.FacebookRequest) (*facebookTypes.DemographicsResponse, error) {
	return &facebookTypes.DemographicsResponse{Status: true}, nil
}
func (m *mockFacebookService) GetAudienceLocation(_ context.Context, _ *facebookTypes.FacebookRequest) (*facebookTypes.DemographicsResponse, error) {
	return &facebookTypes.DemographicsResponse{Status: true}, nil
}

func newTestFacebookHandler() *facebookAPI.Handler {
	return facebookAPI.NewHandler(&mockFacebookService{}, zerolog.New(io.Discard))
}

// --- Instagram Mock service ---

type mockInstagramService struct{}

var _ instagramSvc.Service = (*mockInstagramService)(nil)

func (m *mockInstagramService) GetSummary(_ context.Context, _ *instagramTypes.InstagramRequest) (*instagramTypes.SummaryResponse, error) {
	return &instagramTypes.SummaryResponse{Status: true}, nil
}
func (m *mockInstagramService) GetAudienceGrowth(_ context.Context, _ *instagramTypes.InstagramRequest) (*instagramTypes.AudienceGrowthResponse, error) {
	return &instagramTypes.AudienceGrowthResponse{Status: true}, nil
}
func (m *mockInstagramService) GetPublishingBehaviour(_ context.Context, _ *instagramTypes.PublishingBehaviourRequest) (*instagramTypes.PublishingBehaviourResponse, error) {
	return &instagramTypes.PublishingBehaviourResponse{Status: true}, nil
}
func (m *mockInstagramService) GetTopPosts(_ context.Context, _ *instagramTypes.TopPostsRequest) (*instagramTypes.TopPostsResponse, error) {
	return &instagramTypes.TopPostsResponse{Status: true}, nil
}
func (m *mockInstagramService) GetActiveUsers(_ context.Context, _ *instagramTypes.InstagramRequest) (*instagramTypes.ActiveUsersResponse, error) {
	return &instagramTypes.ActiveUsersResponse{Status: true}, nil
}
func (m *mockInstagramService) GetImpressions(_ context.Context, _ *instagramTypes.InstagramRequest) (*instagramTypes.ImpressionsResponse, error) {
	return &instagramTypes.ImpressionsResponse{Status: true}, nil
}
func (m *mockInstagramService) GetEngagement(_ context.Context, _ *instagramTypes.InstagramRequest) (*instagramTypes.EngagementResponse, error) {
	return &instagramTypes.EngagementResponse{Status: true}, nil
}
func (m *mockInstagramService) GetHashtags(_ context.Context, _ *instagramTypes.InstagramRequest) (*instagramTypes.HashtagsResponse, error) {
	return &instagramTypes.HashtagsResponse{Status: true}, nil
}
func (m *mockInstagramService) GetStoriesPerformance(_ context.Context, _ *instagramTypes.InstagramRequest) (*instagramTypes.StoriesPerformanceResponse, error) {
	return &instagramTypes.StoriesPerformanceResponse{Status: true}, nil
}
func (m *mockInstagramService) GetReelsPerformance(_ context.Context, _ *instagramTypes.InstagramRequest) (*instagramTypes.ReelsPerformanceResponse, error) {
	return &instagramTypes.ReelsPerformanceResponse{Status: true}, nil
}
func (m *mockInstagramService) GetDemographicsAge(_ context.Context, _ *instagramTypes.InstagramRequest) (*instagramTypes.DemographicsAgeResponse, error) {
	return &instagramTypes.DemographicsAgeResponse{}, nil
}
func (m *mockInstagramService) GetCountryCity(_ context.Context, _ *instagramTypes.InstagramRequest) (*instagramTypes.CountryCityResponse, error) {
	return &instagramTypes.CountryCityResponse{}, nil
}

func newTestInstagramHandler() *instagramAPI.InstagramHandler {
	return instagramAPI.NewInstagramHandler(&mockInstagramService{}, zerolog.New(io.Discard))
}

// --- YouTube Mock service ---

type mockYoutubeService struct{}

var _ youtubeSvc.Service = (*mockYoutubeService)(nil)

func (m *mockYoutubeService) GetSummary(_ context.Context, _ *youtubeTypes.YoutubeRequest) (*youtubeTypes.SummaryResponse, error) {
	return &youtubeTypes.SummaryResponse{Status: true}, nil
}
func (m *mockYoutubeService) GetSubscriberTrend(_ context.Context, _ *youtubeTypes.YoutubeRequest) (*youtubeTypes.SubscriberTrendResponse, error) {
	return &youtubeTypes.SubscriberTrendResponse{Status: true}, nil
}
func (m *mockYoutubeService) GetDynamicSubscriberTrend(_ context.Context, _ *youtubeTypes.YoutubeRequest) (*youtubeTypes.SubscriberTrendResponse, error) {
	return &youtubeTypes.SubscriberTrendResponse{Status: true}, nil
}
func (m *mockYoutubeService) GetEngagementTrend(_ context.Context, _ *youtubeTypes.YoutubeRequest) (*youtubeTypes.EngagementTrendResponse, error) {
	return &youtubeTypes.EngagementTrendResponse{Status: true}, nil
}
func (m *mockYoutubeService) GetDynamicEngagementTrend(_ context.Context, _ *youtubeTypes.YoutubeRequest) (*youtubeTypes.EngagementTrendResponse, error) {
	return &youtubeTypes.EngagementTrendResponse{Status: true}, nil
}
func (m *mockYoutubeService) GetViewsTrend(_ context.Context, _ *youtubeTypes.YoutubeRequest) (*youtubeTypes.ViewsTrendResponse, error) {
	return &youtubeTypes.ViewsTrendResponse{Status: true}, nil
}
func (m *mockYoutubeService) GetDynamicViewsTrend(_ context.Context, _ *youtubeTypes.YoutubeRequest) (*youtubeTypes.ViewsTrendResponse, error) {
	return &youtubeTypes.ViewsTrendResponse{Status: true}, nil
}
func (m *mockYoutubeService) GetWatchTimeTrend(_ context.Context, _ *youtubeTypes.YoutubeRequest) (*youtubeTypes.WatchTimeTrendResponse, error) {
	return &youtubeTypes.WatchTimeTrendResponse{Status: true}, nil
}
func (m *mockYoutubeService) GetDynamicWatchTimeTrend(_ context.Context, _ *youtubeTypes.YoutubeRequest) (*youtubeTypes.WatchTimeTrendResponse, error) {
	return &youtubeTypes.WatchTimeTrendResponse{Status: true}, nil
}
func (m *mockYoutubeService) GetFindVideo(_ context.Context, _ *youtubeTypes.YoutubeRequest) (*youtubeTypes.FindVideoResponse, error) {
	return &youtubeTypes.FindVideoResponse{Status: true}, nil
}
func (m *mockYoutubeService) GetVideoSharing(_ context.Context, _ *youtubeTypes.YoutubeRequest) (*youtubeTypes.VideoSharingResponse, error) {
	return &youtubeTypes.VideoSharingResponse{Status: true}, nil
}
func (m *mockYoutubeService) GetTopVideos(_ context.Context, _ *youtubeTypes.YoutubeRequest) (*youtubeTypes.TopVideosResponse, error) {
	return &youtubeTypes.TopVideosResponse{Status: true}, nil
}
func (m *mockYoutubeService) GetLeastVideos(_ context.Context, _ *youtubeTypes.YoutubeRequest) (*youtubeTypes.LeastVideosResponse, error) {
	return &youtubeTypes.LeastVideosResponse{Status: true}, nil
}
func (m *mockYoutubeService) GetSortedTopVideos(_ context.Context, _ *youtubeTypes.TopVideosRequest) (*youtubeTypes.SortedTopVideosResponse, error) {
	return &youtubeTypes.SortedTopVideosResponse{Status: true}, nil
}
func (m *mockYoutubeService) GetPerformanceAndSchedule(_ context.Context, _ *youtubeTypes.YoutubeRequest) (*youtubeTypes.PerformanceScheduleResponse, error) {
	return &youtubeTypes.PerformanceScheduleResponse{Status: true}, nil
}

func newTestYoutubeHandler() *youtubeAPI.Handler {
	return youtubeAPI.NewHandler(&mockYoutubeService{}, zerolog.New(io.Discard))
}

// --- Pinterest Mock service ---

type mockPinterestService struct{}

var _ pinterestSvc.Service = (*mockPinterestService)(nil)

func (m *mockPinterestService) GetSummary(_ context.Context, _ *pinterestTypes.PinterestRequest) (*pinterestTypes.SummaryResponse, error) {
	return &pinterestTypes.SummaryResponse{Status: true}, nil
}
func (m *mockPinterestService) GetFollowerTrend(_ context.Context, _ *pinterestTypes.PinterestRequest) (*pinterestTypes.FollowerTrendResponse, error) {
	return &pinterestTypes.FollowerTrendResponse{Status: true}, nil
}
func (m *mockPinterestService) GetDynamicFollowerTrend(_ context.Context, _ *pinterestTypes.PinterestRequest) (*pinterestTypes.FollowerTrendResponse, error) {
	return &pinterestTypes.FollowerTrendResponse{Status: true}, nil
}
func (m *mockPinterestService) GetImpressionsTrend(_ context.Context, _ *pinterestTypes.PinterestRequest) (*pinterestTypes.ImpressionsTrendResponse, error) {
	return &pinterestTypes.ImpressionsTrendResponse{Status: true}, nil
}
func (m *mockPinterestService) GetDynamicImpressionsTrend(_ context.Context, _ *pinterestTypes.PinterestRequest) (*pinterestTypes.ImpressionsTrendResponse, error) {
	return &pinterestTypes.ImpressionsTrendResponse{Status: true}, nil
}
func (m *mockPinterestService) GetEngagementTrend(_ context.Context, _ *pinterestTypes.PinterestRequest) (*pinterestTypes.EngagementTrendResponse, error) {
	return &pinterestTypes.EngagementTrendResponse{Status: true}, nil
}
func (m *mockPinterestService) GetDynamicEngagementTrend(_ context.Context, _ *pinterestTypes.PinterestRequest) (*pinterestTypes.EngagementTrendResponse, error) {
	return &pinterestTypes.EngagementTrendResponse{Status: true}, nil
}
func (m *mockPinterestService) GetPinPosting(_ context.Context, _ *pinterestTypes.FilteredPinRequest) (*pinterestTypes.PinPostingResponse, error) {
	return &pinterestTypes.PinPostingResponse{Status: true}, nil
}
func (m *mockPinterestService) GetDynamicPinPosting(_ context.Context, _ *pinterestTypes.FilteredPinRequest) (*pinterestTypes.PinPostingResponse, error) {
	return &pinterestTypes.PinPostingResponse{Status: true}, nil
}
func (m *mockPinterestService) GetPinRollup(_ context.Context, _ *pinterestTypes.PinterestRequest) (*pinterestTypes.PinRollupResponse, error) {
	return &pinterestTypes.PinRollupResponse{Status: true}, nil
}
func (m *mockPinterestService) GetTopPins(_ context.Context, _ *pinterestTypes.TopPinsRequest) (*pinterestTypes.TopPinsResponse, error) {
	return &pinterestTypes.TopPinsResponse{Status: true}, nil
}
func (m *mockPinterestService) GetPinPerformance(_ context.Context, _ *pinterestTypes.PinterestRequest) (*pinterestTypes.PinPerformanceResponse, error) {
	return &pinterestTypes.PinPerformanceResponse{Status: true}, nil
}

func newTestPinterestHandler() *pinterestAPI.Handler {
	return pinterestAPI.NewHandler(&mockPinterestService{}, zerolog.New(io.Discard))
}

type mockTwitterService struct{}

var _ twitterSvc.Service = (*mockTwitterService)(nil)

func (m *mockTwitterService) GetPageAndPostsInsights(_ context.Context, _ *twitterTypes.TwitterRequest) (*twitterTypes.MetricsResponse, error) {
	return &twitterTypes.MetricsResponse{Data: map[string]interface{}{}}, nil
}
func (m *mockTwitterService) GetEngagementImpressionData(_ context.Context, _ *twitterTypes.TwitterRequest) (*twitterTypes.EngagementImpressionResponse, error) {
	return &twitterTypes.EngagementImpressionResponse{}, nil
}
func (m *mockTwitterService) GetFollowersTrendData(_ context.Context, _ *twitterTypes.TwitterRequest) (*twitterTypes.FollowersTrendResponse, error) {
	return &twitterTypes.FollowersTrendResponse{}, nil
}
func (m *mockTwitterService) GetTopTweets(_ context.Context, _ *twitterTypes.TweetsRequest) (*twitterTypes.TopTweetsResponse, error) {
	return &twitterTypes.TopTweetsResponse{TopTweets: []twitterTypes.Tweet{}}, nil
}
func (m *mockTwitterService) GetLeastTweets(_ context.Context, _ *twitterTypes.TweetsRequest) (*twitterTypes.LeastTweetsResponse, error) {
	return &twitterTypes.LeastTweetsResponse{LeastTweets: []twitterTypes.Tweet{}}, nil
}
func (m *mockTwitterService) GetCreditsUsedCount(_ context.Context, _ *twitterTypes.TwitterRequest) (*twitterTypes.CreditsUsedResponse, error) {
	return &twitterTypes.CreditsUsedResponse{}, nil
}

func newTestTwitterHandler() *twitterAPI.Handler {
	return twitterAPI.NewHandler(&mockTwitterService{}, zerolog.New(io.Discard))
}

type mockTiktokService struct{}

var _ tiktokSvc.Service = (*mockTiktokService)(nil)

func (m *mockTiktokService) GetPageAndPostsInsights(_ context.Context, _ *tiktokTypes.TiktokRequest) (map[string]interface{}, error) {
	return map[string]interface{}{"data": map[string]interface{}{}}, nil
}
func (m *mockTiktokService) GetPageFollowersAndViews(_ context.Context, _ *tiktokTypes.TiktokRequest) (map[string]interface{}, error) {
	return map[string]interface{}{"data": []map[string]interface{}{}}, nil
}
func (m *mockTiktokService) GetDynamicPageFollowersAndViews(_ context.Context, _ *tiktokTypes.TiktokRequest) (map[string]interface{}, error) {
	return map[string]interface{}{"data": []map[string]interface{}{}}, nil
}
func (m *mockTiktokService) GetPostsAndEngagements(_ context.Context, _ *tiktokTypes.TiktokRequest) (map[string]interface{}, error) {
	return map[string]interface{}{"data": []map[string]interface{}{}}, nil
}
func (m *mockTiktokService) GetDailyEngagementsData(_ context.Context, _ *tiktokTypes.TiktokRequest) (map[string]interface{}, error) {
	return map[string]interface{}{"data": []map[string]interface{}{}}, nil
}
func (m *mockTiktokService) GetDynamicDailyEngagementsData(_ context.Context, _ *tiktokTypes.TiktokRequest) (map[string]interface{}, error) {
	return map[string]interface{}{"data": []map[string]interface{}{}}, nil
}
func (m *mockTiktokService) GetTopAndLeastPerformingPosts(_ context.Context, _ *tiktokTypes.TiktokRequest) (map[string]interface{}, error) {
	return map[string]interface{}{"data": map[string]interface{}{"top_posts": []map[string]interface{}{}, "least_posts": []map[string]interface{}{}}}, nil
}
func (m *mockTiktokService) GetPostsData(_ context.Context, _ *tiktokTypes.PostsRequest) (map[string]interface{}, error) {
	return map[string]interface{}{"data": []map[string]interface{}{}}, nil
}

func newTestTiktokHandler() *tiktokAPI.Handler {
	return tiktokAPI.NewHandler(&mockTiktokService{}, zerolog.New(io.Discard))
}

// --- Mock overview service ---

type mockOverviewService struct{}

var _ overviewSvc.Service = (*mockOverviewService)(nil)

func (m *mockOverviewService) GetSummary(_ context.Context, _ *overviewTypes.OverviewRequest) (*overviewTypes.SummaryResponse, error) {
	return &overviewTypes.SummaryResponse{Summary: &overviewTypes.SummaryData{}}, nil
}
func (m *mockOverviewService) GetTopPerformingGraph(_ context.Context, _ *overviewTypes.OverviewRequest) (*overviewTypes.TopPerformingGraphResponse, error) {
	return &overviewTypes.TopPerformingGraphResponse{}, nil
}
func (m *mockOverviewService) GetPlatformDataGrouped(_ context.Context, _ *overviewTypes.OverviewRequest) ([]*overviewTypes.PlatformDataRow, error) {
	return nil, nil
}
func (m *mockOverviewService) GetPlatformDataIndividual(_ context.Context, _ *overviewTypes.OverviewRequest) ([]*overviewTypes.AccountDataRow, error) {
	return nil, nil
}
func (m *mockOverviewService) GetPlatformDataDetailed(_ context.Context, _ *overviewTypes.OverviewRequest) ([]*overviewTypes.AccountDataDetailedRow, error) {
	return nil, nil
}
func (m *mockOverviewService) GetPlatformDataGraphs(_ context.Context, _ *overviewTypes.OverviewRequest) ([]*overviewTypes.AccountDataGraphsRow, error) {
	return nil, nil
}
func (m *mockOverviewService) GetTopPosts(_ context.Context, _ *overviewTypes.TopPostsRequest) ([]*overviewTypes.TopPostRow, error) {
	return nil, nil
}

func newTestOverviewHandler() *overviewAPI.Handler {
	return overviewAPI.NewHandler(&mockOverviewService{}, zerolog.New(io.Discard))
}

// --- Campaign & Label Mock service ---

type mockCampaignLabelService struct{}

var _ campaignLabelSvc.Service = (*mockCampaignLabelService)(nil)

func (m *mockCampaignLabelService) SetPostIds(_ context.Context, _ *campaignLabelTypes.CampaignLabelRequest) (*campaignLabelTypes.SetPostIdsResponse, error) {
	return &campaignLabelTypes.SetPostIdsResponse{MatchedPostedIds: []string{"post1"}}, nil
}
func (m *mockCampaignLabelService) GetSummaryAnalytics(_ context.Context, _ *campaignLabelTypes.CampaignLabelRequest) (*campaignLabelTypes.SummaryResponse, error) {
	return &campaignLabelTypes.SummaryResponse{
		Current:    map[string]interface{}{"total_posts": 10},
		Previous:   map[string]interface{}{"total_posts": 5},
		Difference: map[string]interface{}{"total_posts": 5},
		Percentage: map[string]interface{}{"total_posts": 100.0},
	}, nil
}
func (m *mockCampaignLabelService) GetBreakdownData(_ context.Context, _ *campaignLabelTypes.CampaignLabelRequest) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}
func (m *mockCampaignLabelService) GetInsightsBreakdown(_ context.Context, _ *campaignLabelTypes.CampaignLabelRequest) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}
func (m *mockCampaignLabelService) GetPlannerAnalytics(_ context.Context, _ *campaignLabelTypes.PlannerAnalyticsRequest) (map[string]interface{}, error) {
	return map[string]interface{}{"engagement": 100}, nil
}

func newTestCampaignLabelHandler() *campaignLabelAPI.Handler {
	return campaignLabelAPI.NewHandler(&mockCampaignLabelService{}, zerolog.New(io.Discard))
}

// --- FB Competitor Mock service ---

type mockFBCompetitorService struct{}

var _ fbCompetitorSvc.Service = (*mockFBCompetitorService)(nil)

func (m *mockFBCompetitorService) GetDataTableMetrics(_ context.Context, _ *fbCompetitorTypes.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockFBCompetitorService) GetPostingActivityGraphByTypes(_ context.Context, _ *fbCompetitorTypes.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockFBCompetitorService) GetPostingActivityBySpecificType(_ context.Context, _ *fbCompetitorTypes.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockFBCompetitorService) GetTopAndLeastPerformingPosts(_ context.Context, _ *fbCompetitorTypes.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockFBCompetitorService) GetTopHashtags(_ context.Context, _ *fbCompetitorTypes.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockFBCompetitorService) GetIndividualHashtagData(_ context.Context, _ *fbCompetitorTypes.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockFBCompetitorService) GetBiographyData(_ context.Context, _ *fbCompetitorTypes.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockFBCompetitorService) GetFollowersGrowthComparison(_ context.Context, _ *fbCompetitorTypes.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockFBCompetitorService) GetPostReactDistribution(_ context.Context, _ *fbCompetitorTypes.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockFBCompetitorService) GetPostReactDistributionByCompany(_ context.Context, _ *fbCompetitorTypes.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockFBCompetitorService) GetPostTypeDistribution(_ context.Context, _ *fbCompetitorTypes.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockFBCompetitorService) GetPostEngagementOverTime(_ context.Context, _ *fbCompetitorTypes.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockFBCompetitorService) GetPostEngagementByCompetitor(_ context.Context, _ *fbCompetitorTypes.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}

func newTestFBCompetitorHandler() *fbCompetitorAPI.Handler {
	return fbCompetitorAPI.NewHandler(&mockFBCompetitorService{}, zerolog.New(io.Discard))
}

// --- IG Competitor Mock service ---

type mockIGCompetitorService struct{}

var _ igCompetitorSvc.Service = (*mockIGCompetitorService)(nil)

func (m *mockIGCompetitorService) GetDataTableMetrics(_ context.Context, _ *igCompetitorTypes.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockIGCompetitorService) GetPostingActivityGraphByTypes(_ context.Context, _ *igCompetitorTypes.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockIGCompetitorService) GetPostingActivityBySpecificType(_ context.Context, _ *igCompetitorTypes.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockIGCompetitorService) GetPostingActivityTableByType(_ context.Context, _ *igCompetitorTypes.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockIGCompetitorService) GetFollowersGrowthComparison(_ context.Context, _ *igCompetitorTypes.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockIGCompetitorService) GetTopAndLeastPerformingPosts(_ context.Context, _ *igCompetitorTypes.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockIGCompetitorService) GetTopHashtags(_ context.Context, _ *igCompetitorTypes.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockIGCompetitorService) GetIndividualHashtagData(_ context.Context, _ *igCompetitorTypes.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockIGCompetitorService) GetBiographyData(_ context.Context, _ *igCompetitorTypes.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}

func newTestIGCompetitorHandler() *igCompetitorAPI.Handler {
	return igCompetitorAPI.NewHandler(&mockIGCompetitorService{}, zerolog.New(io.Discard))
}

type mockLookerAPIKeyFinder struct{ err error }

func (m *mockLookerAPIKeyFinder) FindActiveByUserID(_ context.Context, _ string) (*mongoModels.ApiKey, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &mongoModels.ApiKey{Key: "test-api-key"}, nil
}

func newTestLookerStudioHandler() *lookerStudioAPI.Handler {
	return lookerStudioAPI.NewHandler(
		config.LookerStudioConfig{ConnectorID: "test-connector-id"},
		&mockLookerAPIKeyFinder{},
		zerolog.New(io.Discard),
	)
}

const validQueryStr = "workspace_id=ws1&linkedin_id=li_123&start_date=2025-01-01&end_date=2025-01-31&timezone=UTC"

func TestRegisterRoutes(t *testing.T) {
	mux := http.NewServeMux()
	h := newTestHandler()
	gh := newTestGMBHandler()
	fh := newTestFacebookHandler()
	th := newTestTwitterHandler()
	tkh := newTestTiktokHandler()
	ih := newTestInstagramHandler()
	yh := newTestYoutubeHandler()
	ph := newTestPinterestHandler()
	ovh := newTestOverviewHandler()
	clh := newTestCampaignLabelHandler()
	fbch := newTestFBCompetitorHandler()
	igch := newTestIGCompetitorHandler()
	RegisterRoutes(mux, h, gh, fh, th, tkh, ih, yh, ph, ovh, clh, nil, fbch, igch, nil)

	validOverviewBody, _ := json.Marshal(map[string]interface{}{
		"workspace_id": "ws1",
		"start_date":   "2025-01-01",
		"end_date":     "2025-01-31",
		"timezone":     "UTC",
	})
	validOverviewBodyWithType, _ := json.Marshal(map[string]interface{}{
		"workspace_id": "ws1",
		"start_date":   "2025-01-01",
		"end_date":     "2025-01-31",
		"timezone":     "UTC",
		"type":         "grouped",
	})
	validOverviewBodyTopPosts, _ := json.Marshal(map[string]interface{}{
		"workspace_id": "ws1",
		"start_date":   "2025-01-01",
		"end_date":     "2025-01-31",
		"timezone":     "UTC",
		"type":         "total_engagement",
		"limit":        10,
	})
	validCampaignLabelBody, _ := json.Marshal(map[string]interface{}{
		"workspace_id":      "ws1",
		"start_date":        "2025-01-01",
		"end_date":          "2025-01-31",
		"campaigns":         []string{"camp1"},
		"labels":            []string{"label1"},
		"facebook_accounts": []string{"fb_123"},
	})
	validPlannerBody, _ := json.Marshal(map[string]interface{}{
		"workspace_id": "ws1",
		"id":           "plan1",
		"all_post_ids": []string{"post1"},
		"platforms":    "facebook",
	})

	tests := []struct {
		name           string
		method         string
		path           string
		body           []byte
		expectedStatus int
	}{
		{name: "summary", method: "GET", path: "/analytics/overview/linkedin/summary", expectedStatus: http.StatusOK},
		{name: "audienceGrowth", method: "GET", path: "/analytics/overview/linkedin/audienceGrowth", expectedStatus: http.StatusOK},
		{name: "pageViews", method: "GET", path: "/analytics/overview/linkedin/pageViews", expectedStatus: http.StatusOK},
		{name: "publishingBehaviour", method: "GET", path: "/analytics/overview/linkedin/publishingBehaviour", expectedStatus: http.StatusOK},
		{name: "topPosts", method: "GET", path: "/analytics/overview/linkedin/topPosts", expectedStatus: http.StatusOK},
		{name: "postsPerDays", method: "GET", path: "/analytics/overview/linkedin/postsPerDays", expectedStatus: http.StatusOK},
		{name: "hashtags", method: "GET", path: "/analytics/overview/linkedin/hashtags", expectedStatus: http.StatusOK},
		{name: "getTopPosts", method: "GET", path: "/analytics/overview/linkedin/getTopPosts", expectedStatus: http.StatusOK},
		{name: "followersDemographics", method: "GET", path: "/analytics/overview/linkedin/followersDemographics", expectedStatus: http.StatusOK},
		{name: "facebook summary get", method: "GET", path: "/analytics/overview/facebook/summary", expectedStatus: http.StatusOK},
		{name: "facebook demographics get", method: "GET", path: "/analytics/overview/facebook/demographics", expectedStatus: http.StatusOK},
		{name: "facebook ai insights route", method: "GET", path: "/analytics/overview/facebook/ai_insights", expectedStatus: http.StatusServiceUnavailable},
		{name: "linkedin ai insights route", method: "GET", path: "/analytics/overview/linkedin/ai_insights", expectedStatus: http.StatusServiceUnavailable},
		{name: "gmb ai insights route", method: "GET", path: "/analytics/overview/gmb/ai_insights", expectedStatus: http.StatusServiceUnavailable},
		{name: "twitter page and post insights route", method: "GET", path: "/analytics/overview/twitter/getPageAndPostsInsights", expectedStatus: http.StatusOK},
		{name: "twitter engagement and impression route", method: "GET", path: "/analytics/overview/twitter/getEngagementImpressionData", expectedStatus: http.StatusOK},
		{name: "twitter followers trend route", method: "GET", path: "/analytics/overview/twitter/getFollowersTrendData", expectedStatus: http.StatusOK},
		{name: "twitter top tweets route", method: "GET", path: "/analytics/overview/twitter/getTopTweets", expectedStatus: http.StatusOK},
		{name: "twitter least tweets route", method: "GET", path: "/analytics/overview/twitter/getLeastTweets", expectedStatus: http.StatusOK},
		{name: "twitter credits used route", method: "GET", path: "/analytics/overview/twitter/getCreditsUsedCount", expectedStatus: http.StatusOK},
		{name: "tiktok summary route", method: "GET", path: "/analytics/overview/tiktok/getPageAndPostsInsights", expectedStatus: http.StatusOK},
		{name: "tiktok followers route", method: "GET", path: "/analytics/overview/tiktok/getPageFollowersAndViews", expectedStatus: http.StatusOK},
		{name: "tiktok posts engagements route", method: "GET", path: "/analytics/overview/tiktok/getPostsAndEngagements", expectedStatus: http.StatusOK},
		{name: "tiktok daily engagements route", method: "GET", path: "/analytics/overview/tiktok/getDailyEngagementsData", expectedStatus: http.StatusOK},
		{name: "tiktok top and least posts route", method: "GET", path: "/analytics/overview/tiktok/getTopAndLeastPerformingPosts", expectedStatus: http.StatusOK},
		{name: "tiktok posts data route", method: "GET", path: "/analytics/overview/tiktok/getPostsData", expectedStatus: http.StatusOK},
		{name: "tiktok ai insights route", method: "GET", path: "/analytics/overview/tiktok/ai_insights", expectedStatus: http.StatusServiceUnavailable},
		{name: "unknown path returns 404", method: "GET", path: "/analytics/overview/linkedin/nonexistent", expectedStatus: http.StatusNotFound},
		{name: "POST on GET-only route", method: "POST", path: "/analytics/overview/linkedin/summary", expectedStatus: http.StatusMethodNotAllowed},
		{name: "POST on facebook summary route", method: "POST", path: "/analytics/overview/facebook/summary", expectedStatus: http.StatusMethodNotAllowed},
		{name: "POST on facebook ai insights route", method: "POST", path: "/analytics/overview/facebook/ai_insights", expectedStatus: http.StatusMethodNotAllowed},
		{name: "POST on linkedin ai insights route", method: "POST", path: "/analytics/overview/linkedin/ai_insights", expectedStatus: http.StatusMethodNotAllowed},
		{name: "POST on gmb ai insights route", method: "POST", path: "/analytics/overview/gmb/ai_insights", expectedStatus: http.StatusMethodNotAllowed},
		{name: "POST on tiktok ai insights route", method: "POST", path: "/analytics/overview/tiktok/ai_insights", expectedStatus: http.StatusMethodNotAllowed},
		// Overview V2 POST routes
		{name: "overviewV2 getSummary POST", method: "POST", path: "/analytics/overview/overviewV2/getSummary", body: validOverviewBody, expectedStatus: http.StatusOK},
		{name: "overviewV2 getTopPerformingGraph POST", method: "POST", path: "/analytics/overview/overviewV2/getTopPerformingGraph", body: validOverviewBody, expectedStatus: http.StatusOK},
		{name: "overviewV2 getPlatformData grouped POST", method: "POST", path: "/analytics/overview/overviewV2/getPlatformData", body: validOverviewBodyWithType, expectedStatus: http.StatusOK},
		{name: "overviewV2 getPlatformDataDetailed POST", method: "POST", path: "/analytics/overview/overviewV2/getPlatformDataDetailed", body: validOverviewBody, expectedStatus: http.StatusOK},
		{name: "overviewV2 getPlatformDataGraphs POST", method: "POST", path: "/analytics/overview/overviewV2/getPlatformDataGraphs", body: validOverviewBody, expectedStatus: http.StatusOK},
		{name: "overviewV2 getTopPosts POST", method: "POST", path: "/analytics/overview/overviewV2/getTopPosts", body: validOverviewBodyTopPosts, expectedStatus: http.StatusOK},
		// Overview V2 GET returns 405 (routes are POST-only)
		{name: "overviewV2 getSummary GET returns 405", method: "GET", path: "/analytics/overview/overviewV2/getSummary", expectedStatus: http.StatusMethodNotAllowed},
		{name: "overviewV2 getTopPerformingGraph GET returns 405", method: "GET", path: "/analytics/overview/overviewV2/getTopPerformingGraph", expectedStatus: http.StatusMethodNotAllowed},
		{name: "overviewV2 getPlatformData GET returns 405", method: "GET", path: "/analytics/overview/overviewV2/getPlatformData", expectedStatus: http.StatusMethodNotAllowed},
		{name: "overviewV2 getPlatformDataDetailed GET returns 405", method: "GET", path: "/analytics/overview/overviewV2/getPlatformDataDetailed", expectedStatus: http.StatusMethodNotAllowed},
		{name: "overviewV2 getPlatformDataGraphs GET returns 405", method: "GET", path: "/analytics/overview/overviewV2/getPlatformDataGraphs", expectedStatus: http.StatusMethodNotAllowed},
		{name: "overviewV2 getTopPosts GET returns 405", method: "GET", path: "/analytics/overview/overviewV2/getTopPosts", expectedStatus: http.StatusMethodNotAllowed},
		// Campaign & Label POST routes
		{name: "campaignLabel setPostIds POST", method: "POST", path: "/analytics/campaignLabelAnalytics/setPostIdsForCampaignsAndLabels", body: validCampaignLabelBody, expectedStatus: http.StatusOK},
		{name: "campaignLabel getSummary POST", method: "POST", path: "/analytics/campaignLabelAnalytics/getSummaryAnalytics", body: validCampaignLabelBody, expectedStatus: http.StatusOK},
		{name: "campaignLabel getBreakdown POST", method: "POST", path: "/analytics/campaignLabelAnalytics/getCampaignLabelBreakdownData", body: validCampaignLabelBody, expectedStatus: http.StatusOK},
		{name: "campaignLabel getInsights POST", method: "POST", path: "/analytics/campaignLabelAnalytics/getCampaignLabelInsightsBreakdown", body: validCampaignLabelBody, expectedStatus: http.StatusOK},
		{name: "campaignLabel getPlannerAnalytics POST", method: "POST", path: "/analytics/campaignLabelAnalytics/getPlannerAnalytics", body: validPlannerBody, expectedStatus: http.StatusOK},
		// Campaign & Label GET returns 405 (routes are POST-only)
		{name: "campaignLabel setPostIds GET returns 405", method: "GET", path: "/analytics/campaignLabelAnalytics/setPostIdsForCampaignsAndLabels", expectedStatus: http.StatusMethodNotAllowed},
		{name: "campaignLabel getSummary GET returns 405", method: "GET", path: "/analytics/campaignLabelAnalytics/getSummaryAnalytics", expectedStatus: http.StatusMethodNotAllowed},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			url := tc.path
			if tc.expectedStatus == http.StatusOK && tc.body == nil {
				switch {
				case tc.path == "/analytics/overview/facebook/ai_insights":
					url += "?workspace_id=ws1&facebook_id=fb_123&date=2025-01-01+-+2025-01-31&type=page_impressions&limit=15"
				case tc.path == "/analytics/overview/linkedin/ai_insights":
					url += "?workspace_id=ws1&linkedin_id=li_123&date=2025-01-01+-+2025-01-31&type=page_views&limit=15"
				case tc.path == "/analytics/overview/gmb/ai_insights":
					url += "?workspace_id=ws1&gmb_id=loc_123&type=impressions_overview"
				case tc.path == "/analytics/overview/tiktok/ai_insights":
					url += "?workspace_id=ws1&tiktok_id=tt_123&date=2025-01-01+-+2025-01-31&type=insights_summary&limit=5"
				case tc.path == "/analytics/overview/facebook/demographics":
					url += "?workspace_id=ws1&facebook_id=fb_123&start_date=2025-01-01&end_date=2025-01-31&timezone=UTC"
				case tc.path == "/analytics/overview/facebook/summary":
					url += "?workspace_id=ws1&facebook_id=fb_123&start_date=2025-01-01&end_date=2025-01-31&timezone=UTC"
				case strings.HasPrefix(tc.path, "/analytics/overview/twitter/"):
					url += "?workspace_id=ws1&twitter_id=tw_123&start_date=2025-01-01&end_date=2025-01-31&timezone=UTC"
				case strings.HasPrefix(tc.path, "/analytics/overview/tiktok/"):
					url += "?workspace_id=ws1&tiktok_id=tt_123&start_date=2025-01-01&end_date=2025-01-31&timezone=UTC"
				default:
					url += "?" + validQueryStr
				}
			}
			var reqBody io.Reader
			if tc.body != nil {
				reqBody = bytes.NewReader(tc.body)
			}
			req := httptest.NewRequest(tc.method, url, reqBody)
			if tc.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Fatalf("expected %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestLookerStudioRouteRegistration(t *testing.T) {
	t.Run("GET route is registered when handler is non-nil", func(t *testing.T) {
		mux := http.NewServeMux()
		RegisterRoutes(mux,
			newTestHandler(), newTestGMBHandler(), newTestFacebookHandler(),
			newTestTwitterHandler(), newTestTiktokHandler(), newTestInstagramHandler(),
			newTestYoutubeHandler(), newTestPinterestHandler(), newTestOverviewHandler(),
			newTestCampaignLabelHandler(), newTestLookerStudioHandler(),
			newTestFBCompetitorHandler(), newTestIGCompetitorHandler(), nil,
		)
		req := httptest.NewRequest("GET", "/analytics/looker-studio/connect?platform=facebook&workspace_id=ws1&account_id=acc1", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		// Route is registered — must not be 404
		if w.Code == http.StatusNotFound {
			t.Fatal("expected looker-studio/connect to be registered, got 404")
		}
	})

	t.Run("GET route not registered when handler is nil", func(t *testing.T) {
		mux := http.NewServeMux()
		RegisterRoutes(mux,
			newTestHandler(), newTestGMBHandler(), newTestFacebookHandler(),
			newTestTwitterHandler(), newTestTiktokHandler(), newTestInstagramHandler(),
			newTestYoutubeHandler(), newTestPinterestHandler(), newTestOverviewHandler(),
			newTestCampaignLabelHandler(), nil,
			newTestFBCompetitorHandler(), newTestIGCompetitorHandler(), nil,
		)
		req := httptest.NewRequest("GET", "/analytics/looker-studio/connect", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 when handler is nil, got %d", w.Code)
		}
	})

	t.Run("POST route is registered", func(t *testing.T) {
		mux := http.NewServeMux()
		RegisterRoutes(mux,
			newTestHandler(), newTestGMBHandler(), newTestFacebookHandler(),
			newTestTwitterHandler(), newTestTiktokHandler(), newTestInstagramHandler(),
			newTestYoutubeHandler(), newTestPinterestHandler(), newTestOverviewHandler(),
			newTestCampaignLabelHandler(), newTestLookerStudioHandler(),
			newTestFBCompetitorHandler(), newTestIGCompetitorHandler(), nil,
		)
		body, _ := json.Marshal(map[string]string{"workspace_id": "ws1"})
		req := httptest.NewRequest("POST", "/analytics/looker-studio/connect", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		// Route is registered — must not be 404 or 405
		if w.Code == http.StatusNotFound || w.Code == http.StatusMethodNotAllowed {
			t.Fatalf("expected POST to be registered, got %d", w.Code)
		}
	})
}
