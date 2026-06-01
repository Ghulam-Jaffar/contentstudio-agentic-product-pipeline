// Package analytics registers all analytics API routes on an http.ServeMux.
// Routes are organized by platform under /analytics/overview/{platform}/.
// Uses Go 1.22+ method-based routing patterns (e.g. "GET /path").
package analytics

import (
	"net/http"

	campaignLabel "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/campaign_label"
	facebook "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/facebook"
	fbCompetitor "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/fb_competitor"
	gmb "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/gmb"
	igCompetitor "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/ig_competitor"
	instagram "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/instagram"
	linkedin "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/linkedin"
	lookerStudio "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/looker_studio"
	metaAds "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/meta_ads"
	overviewHandler "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/overview"
	pinterest "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/pinterest"
	tiktok "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/tiktok"
	twitter "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/twitter"
	youtube "github.com/d4interactive/contentstudio-social-analytics-go/src/api/analytics/youtube"
)

// RegisterRoutes registers all analytics endpoints for all platforms.
// Authentication is handled at a higher level by AuthMiddleware wrapping the mux.
func RegisterRoutes(
	mux *http.ServeMux,
	linkedinHandler *linkedin.LinkedInHandler,
	gmbHandler *gmb.GMBHandler,
	facebookHandler *facebook.Handler,
	twitterHandler *twitter.Handler,
	tiktokHandler *tiktok.Handler,
	instagramHandler *instagram.InstagramHandler,
	youtubeHandler *youtube.Handler,
	pinterestHandler *pinterest.Handler,
	overviewH *overviewHandler.Handler,
	campaignLabelHandler *campaignLabel.Handler,
	lookerStudioHandler *lookerStudio.Handler,
	fbCompetitorHandler *fbCompetitor.Handler,
	igCompetitorHandler *igCompetitor.Handler,
	metaAdsHandler *metaAds.Handler,
) {
	// Looker Studio connector
	if lookerStudioHandler != nil {
		mux.HandleFunc("GET /analytics/looker-studio/connect", lookerStudioHandler.HandleConnect)
		mux.HandleFunc("POST /analytics/looker-studio/connect", lookerStudioHandler.HandleConnect)
	}
	// LinkedIn routes
	mux.HandleFunc("GET /analytics/overview/linkedin/summary", linkedinHandler.HandleSummary)
	mux.HandleFunc("GET /analytics/overview/linkedin/audienceGrowth", linkedinHandler.HandleAudienceGrowth)
	mux.HandleFunc("GET /analytics/overview/linkedin/pageViews", linkedinHandler.HandlePageViews)
	mux.HandleFunc("GET /analytics/overview/linkedin/publishingBehaviour", linkedinHandler.HandlePublishingBehaviour)
	mux.HandleFunc("GET /analytics/overview/linkedin/topPosts", linkedinHandler.HandleTopPosts)
	mux.HandleFunc("GET /analytics/overview/linkedin/postsPerDays", linkedinHandler.HandlePostsPerDay)
	mux.HandleFunc("GET /analytics/overview/linkedin/hashtags", linkedinHandler.HandleHashtags)
	mux.HandleFunc("GET /analytics/overview/linkedin/getTopPosts", linkedinHandler.HandleGetTopPosts)
	mux.HandleFunc("GET /analytics/overview/linkedin/followersDemographics", linkedinHandler.HandleFollowersDemographics)
	mux.HandleFunc("GET /analytics/overview/linkedin/ai_insights", linkedinHandler.HandleAIInsights)

	// GMB routes
	mux.HandleFunc("GET /analytics/overview/gmb/summary", gmbHandler.HandleSummary)
	mux.HandleFunc("GET /analytics/overview/gmb/impressions", gmbHandler.HandleImpressions)
	mux.HandleFunc("GET /analytics/overview/gmb/actions", gmbHandler.HandleActions)
	mux.HandleFunc("GET /analytics/overview/gmb/searchKeywords", gmbHandler.HandleSearchKeywords)
	mux.HandleFunc("GET /analytics/overview/gmb/topPosts", gmbHandler.HandleTopPosts)
	mux.HandleFunc("GET /analytics/overview/gmb/publishingBehavior", gmbHandler.HandlePublishingBehavior)
	mux.HandleFunc("GET /analytics/overview/gmb/reviews", gmbHandler.HandleReviews)
	mux.HandleFunc("GET /analytics/overview/gmb/mediaActivity", gmbHandler.HandleMediaActivity)
	mux.HandleFunc("GET /analytics/overview/gmb/ai_insights", gmbHandler.HandleAIInsights)

	// Facebook routes
	mux.HandleFunc("GET /analytics/overview/facebook/summary", facebookHandler.HandleSummary)
	mux.HandleFunc("GET /analytics/overview/facebook/overviewAudienceGrowth", facebookHandler.HandleAudienceGrowth)
	mux.HandleFunc("GET /analytics/overview/facebook/overviewPublishingBehaviour", facebookHandler.HandlePublishingBehaviour)
	mux.HandleFunc("GET /analytics/overview/facebook/overviewTopPosts", facebookHandler.HandleOverviewTopPosts)
	mux.HandleFunc("GET /analytics/overview/facebook/getTopPosts", facebookHandler.HandleGetTopPosts)
	mux.HandleFunc("GET /analytics/overview/facebook/overviewActiveUsers", facebookHandler.HandleActiveUsers)
	mux.HandleFunc("GET /analytics/overview/facebook/overviewImpressions", facebookHandler.HandleImpressions)
	mux.HandleFunc("GET /analytics/overview/facebook/overviewEngagement", facebookHandler.HandleEngagement)
	mux.HandleFunc("GET /analytics/overview/facebook/overviewReelsAnalytics", facebookHandler.HandleReelsAnalytics)
	mux.HandleFunc("GET /analytics/overview/facebook/overviewVideoInsights", facebookHandler.HandleVideoInsights)
	mux.HandleFunc("GET /analytics/overview/facebook/demographics", facebookHandler.HandleDemographics)
	mux.HandleFunc("GET /analytics/overview/facebook/overviewDemographics", facebookHandler.HandleOverviewDemographics)
	mux.HandleFunc("GET /analytics/overview/facebook/overviewAudienceLocation", facebookHandler.HandleAudienceLocation)
	mux.HandleFunc("GET /analytics/overview/facebook/ai_insights", facebookHandler.HandleAIInsights)

	// Twitter routes
	mux.HandleFunc("GET /analytics/overview/twitter/getPageAndPostsInsights", twitterHandler.HandlePageAndPostsInsights)
	mux.HandleFunc("GET /analytics/overview/twitter/getEngagementImpressionData", twitterHandler.HandleEngagementImpressionData)
	mux.HandleFunc("GET /analytics/overview/twitter/getFollowersTrendData", twitterHandler.HandleFollowersTrendData)
	mux.HandleFunc("GET /analytics/overview/twitter/getTopTweets", twitterHandler.HandleTopTweets)
	mux.HandleFunc("GET /analytics/overview/twitter/getLeastTweets", twitterHandler.HandleLeastTweets)
	mux.HandleFunc("GET /analytics/overview/twitter/getCreditsUsedCount", twitterHandler.HandleCreditsUsedCount)

	// TikTok routes
	mux.HandleFunc("GET /analytics/overview/tiktok/getPageAndPostsInsights", tiktokHandler.HandlePageAndPostsInsights)
	mux.HandleFunc("GET /analytics/overview/tiktok/getPageFollowersAndViews", tiktokHandler.HandlePageFollowersAndViews)
	mux.HandleFunc("GET /analytics/overview/tiktok/getPostsAndEngagements", tiktokHandler.HandlePostsAndEngagements)
	mux.HandleFunc("GET /analytics/overview/tiktok/getDailyEngagementsData", tiktokHandler.HandleDailyEngagementsData)
	mux.HandleFunc("GET /analytics/overview/tiktok/getTopAndLeastPerformingPosts", tiktokHandler.HandleTopAndLeastPerformingPosts)
	mux.HandleFunc("GET /analytics/overview/tiktok/getPostsData", tiktokHandler.HandlePostsData)
	mux.HandleFunc("GET /analytics/overview/tiktok/ai_insights", tiktokHandler.HandleAIInsights)

	// Instagram routes
	mux.HandleFunc("GET /analytics/overview/instagram/summary", instagramHandler.HandleSummary)
	mux.HandleFunc("GET /analytics/overview/instagram/audienceGrowth", instagramHandler.HandleAudienceGrowth)
	mux.HandleFunc("GET /analytics/overview/instagram/publishingBehaviour", instagramHandler.HandlePublishingBehaviour)
	mux.HandleFunc("GET /analytics/overview/instagram/topPosts", instagramHandler.HandleTopPosts)
	mux.HandleFunc("GET /analytics/overview/instagram/getTopPosts", instagramHandler.HandleGetTopPosts)
	mux.HandleFunc("GET /analytics/overview/instagram/activeUsers", instagramHandler.HandleActiveUsers)
	mux.HandleFunc("GET /analytics/overview/instagram/impressions", instagramHandler.HandleImpressions)
	mux.HandleFunc("GET /analytics/overview/instagram/engagement", instagramHandler.HandleEngagement)
	mux.HandleFunc("GET /analytics/overview/instagram/hashtags", instagramHandler.HandleHashtags)
	mux.HandleFunc("GET /analytics/overview/instagram/storiesPerformance", instagramHandler.HandleStoriesPerformance)
	mux.HandleFunc("GET /analytics/overview/instagram/reelsPerformance", instagramHandler.HandleReelsPerformance)
	mux.HandleFunc("GET /analytics/overview/instagram/demographicsAge", instagramHandler.HandleDemographicsAge)
	mux.HandleFunc("GET /analytics/overview/instagram/countryCity", instagramHandler.HandleCountryCity)
	mux.HandleFunc("GET /analytics/overview/instagram/ai_insights", instagramHandler.HandleAIInsights)

	// YouTube routes
	mux.HandleFunc("GET /analytics/overview/youtube/overviewSummary", youtubeHandler.HandleSummary)
	mux.HandleFunc("GET /analytics/overview/youtube/overviewSubscriberTrend", youtubeHandler.HandleSubscriberTrend)
	mux.HandleFunc("GET /analytics/overview/youtube/overviewDynamicSubscriberTrend", youtubeHandler.HandleDynamicSubscriberTrend)
	mux.HandleFunc("GET /analytics/overview/youtube/overviewEngagementTrend", youtubeHandler.HandleEngagementTrend)
	mux.HandleFunc("GET /analytics/overview/youtube/overviewDynamicEngagementTrend", youtubeHandler.HandleDynamicEngagementTrend)
	mux.HandleFunc("GET /analytics/overview/youtube/overviewViewsTrend", youtubeHandler.HandleViewsTrend)
	mux.HandleFunc("GET /analytics/overview/youtube/overviewDynamicViewsTrend", youtubeHandler.HandleDynamicViewsTrend)
	mux.HandleFunc("GET /analytics/overview/youtube/overviewWatchTimeTrend", youtubeHandler.HandleWatchTimeTrend)
	mux.HandleFunc("GET /analytics/overview/youtube/overviewDynamicWatchTimeTrend", youtubeHandler.HandleDynamicWatchTimeTrend)
	mux.HandleFunc("GET /analytics/overview/youtube/overviewFindVideo", youtubeHandler.HandleFindVideo)
	mux.HandleFunc("GET /analytics/overview/youtube/overviewVideoSharing", youtubeHandler.HandleVideoSharing)
	mux.HandleFunc("GET /analytics/overview/youtube/overviewTopPosts", youtubeHandler.HandleTopPosts)
	mux.HandleFunc("GET /analytics/overview/youtube/overviewLeastPosts", youtubeHandler.HandleLeastPosts)
	mux.HandleFunc("GET /analytics/overview/youtube/getSortedTopPosts", youtubeHandler.HandleGetTopPosts)
	mux.HandleFunc("GET /analytics/overview/youtube/overviewPerformanceAndVideoPostingSchedule", youtubeHandler.HandlePerformanceAndSchedule)
	mux.HandleFunc("GET /analytics/overview/youtube/ai_insights", youtubeHandler.HandleAIInsights)

	// Pinterest routes
	mux.HandleFunc("GET /analytics/overview/pinterest/overviewSummary", pinterestHandler.HandleSummary)
	mux.HandleFunc("GET /analytics/overview/pinterest/overviewFollowers", pinterestHandler.HandleFollowerTrend)
	mux.HandleFunc("GET /analytics/overview/pinterest/overviewDynamicFollowers", pinterestHandler.HandleDynamicFollowerTrend)
	mux.HandleFunc("GET /analytics/overview/pinterest/overviewImpressions", pinterestHandler.HandleImpressionsTrend)
	mux.HandleFunc("GET /analytics/overview/pinterest/overviewDynamicImpressions", pinterestHandler.HandleDynamicImpressionsTrend)
	mux.HandleFunc("GET /analytics/overview/pinterest/overviewEngagement", pinterestHandler.HandleEngagementTrend)
	mux.HandleFunc("GET /analytics/overview/pinterest/overviewDynamicEngagement", pinterestHandler.HandleDynamicEngagementTrend)
	mux.HandleFunc("GET /analytics/overview/pinterest/overviewPinPostingPerDay", pinterestHandler.HandlePinPosting)
	mux.HandleFunc("GET /analytics/overview/pinterest/overviewDynamicPinPostingPerDay", pinterestHandler.HandleDynamicPinPosting)
	mux.HandleFunc("GET /analytics/overview/pinterest/overviewPinPostingRollup", pinterestHandler.HandlePinRollup)
	mux.HandleFunc("GET /analytics/overview/pinterest/overviewTopPins", pinterestHandler.HandleTopPins)
	mux.HandleFunc("GET /analytics/overview/pinterest/overviewPinPostingPerformance", pinterestHandler.HandlePinPerformance)
	mux.HandleFunc("GET /analytics/overview/pinterest/ai_insights", pinterestHandler.HandleAIInsights)

	// Cross-platform Overview V2 routes
	mux.HandleFunc("POST /analytics/overview/overviewV2/getSummary", overviewH.HandleSummary)
	mux.HandleFunc("POST /analytics/overview/overviewV2/getTopPerformingGraph", overviewH.HandleTopPerformingGraph)
	mux.HandleFunc("POST /analytics/overview/overviewV2/getPlatformData", overviewH.HandlePlatformData)
	mux.HandleFunc("POST /analytics/overview/overviewV2/getPlatformDataDetailed", overviewH.HandlePlatformDataDetailed)
	mux.HandleFunc("POST /analytics/overview/overviewV2/getPlatformDataGraphs", overviewH.HandlePlatformDataGraphs)
	mux.HandleFunc("POST /analytics/overview/overviewV2/getTopPosts", overviewH.HandleTopPosts)
	mux.HandleFunc("POST /analytics/overview/overviewV2/ai_insights", overviewH.HandleAIInsights)

	// Campaign & Label Analytics routes (POST — payloads contain variable-length arrays
	// of campaigns, labels, and per-platform account IDs)
	mux.HandleFunc("POST /analytics/campaignLabelAnalytics/setPostIdsForCampaignsAndLabels", campaignLabelHandler.HandleSetPostIds)
	mux.HandleFunc("POST /analytics/campaignLabelAnalytics/getSummaryAnalytics", campaignLabelHandler.HandleSummaryAnalytics)
	mux.HandleFunc("POST /analytics/campaignLabelAnalytics/getCampaignLabelBreakdownData", campaignLabelHandler.HandleBreakdownData)
	mux.HandleFunc("POST /analytics/campaignLabelAnalytics/getCampaignLabelInsightsBreakdown", campaignLabelHandler.HandleInsightsBreakdown)
	mux.HandleFunc("POST /analytics/campaignLabelAnalytics/getPlannerAnalytics", campaignLabelHandler.HandlePlannerAnalytics)

	// Facebook Competitor routes (analytics — GET, CRUD — POST)
	mux.HandleFunc("GET /analytics/overview/facebook/competitor/dataTableMetrics", fbCompetitorHandler.HandleDataTableMetrics)
	mux.HandleFunc("GET /analytics/overview/facebook/competitor/postingActivityGraphByTypes", fbCompetitorHandler.HandlePostingActivityGraphByTypes)
	mux.HandleFunc("GET /analytics/overview/facebook/competitor/postingActivityBySpecificType", fbCompetitorHandler.HandlePostingActivityBySpecificType)
	mux.HandleFunc("GET /analytics/overview/facebook/competitor/postReactDistribution", fbCompetitorHandler.HandlePostReactDistribution)
	mux.HandleFunc("GET /analytics/overview/facebook/competitor/postReactDistributionByCompany", fbCompetitorHandler.HandlePostReactDistributionByCompany)
	mux.HandleFunc("GET /analytics/overview/facebook/competitor/postTypeDistribution", fbCompetitorHandler.HandlePostTypeDistribution)
	mux.HandleFunc("GET /analytics/overview/facebook/competitor/topAndLeastPerformingPosts", fbCompetitorHandler.HandleTopAndLeastPerformingPosts)
	mux.HandleFunc("GET /analytics/overview/facebook/competitor/topHashtags", fbCompetitorHandler.HandleTopHashtags)
	mux.HandleFunc("GET /analytics/overview/facebook/competitor/individualHashtagData", fbCompetitorHandler.HandleIndividualHashtagData)
	mux.HandleFunc("GET /analytics/overview/facebook/competitor/biographyData", fbCompetitorHandler.HandleBiographyData)
	mux.HandleFunc("GET /analytics/overview/facebook/competitor/followersGrowthComparison", fbCompetitorHandler.HandleFollowersGrowthComparison)
	mux.HandleFunc("GET /analytics/overview/facebook/competitor/postEngagementOverTime", fbCompetitorHandler.HandlePostEngagementOverTime)
	mux.HandleFunc("GET /analytics/overview/facebook/competitor/postEngagementByCompetitor", fbCompetitorHandler.HandlePostEngagementByCompetitor)

	// Meta Ads Analytics routes
	mux.HandleFunc("GET /analytics/overview/meta-ads/summary", metaAdsHandler.HandleSummary)
	mux.HandleFunc("GET /analytics/overview/meta-ads/resultsByObjective", metaAdsHandler.HandleResultsByObjective)
	mux.HandleFunc("GET /analytics/overview/meta-ads/impressionsVsSpend", metaAdsHandler.HandleImpressionsVsSpend)
	mux.HandleFunc("GET /analytics/overview/meta-ads/clicksVsCtr", metaAdsHandler.HandleClicksVsCTR)
	mux.HandleFunc("GET /analytics/overview/meta-ads/topCampaigns", metaAdsHandler.HandleTopCampaigns)
	mux.HandleFunc("GET /analytics/overview/meta-ads/performanceTrend", metaAdsHandler.HandlePerformanceTrend)
	mux.HandleFunc("GET /analytics/overview/meta-ads/performanceByLevel", metaAdsHandler.HandlePerformanceByLevel)
	mux.HandleFunc("GET /analytics/overview/meta-ads/performanceByPlatform", metaAdsHandler.HandlePerformanceByPlatform)
	mux.HandleFunc("GET /analytics/overview/meta-ads/campaignsList", metaAdsHandler.HandleCampaignsList)
	mux.HandleFunc("GET /analytics/overview/meta-ads/adSetsList", metaAdsHandler.HandleAdSetsList)
	mux.HandleFunc("GET /analytics/overview/meta-ads/adsList", metaAdsHandler.HandleAdsList)
	mux.HandleFunc("GET /analytics/overview/meta-ads/demographicsAgeGender", metaAdsHandler.HandleDemographicsAgeGender)
	mux.HandleFunc("GET /analytics/overview/meta-ads/demographicsRegionCountry", metaAdsHandler.HandleDemographicsRegionCountry)
	// Meta Ads AI Insights routes
	mux.HandleFunc("GET /analytics/overview/meta-ads/aiInsightsSummary", metaAdsHandler.HandleAIInsightsSummary)
	mux.HandleFunc("GET /analytics/overview/meta-ads/aiInsightsDetailed", metaAdsHandler.HandleAIInsightsDetailed)

	// Instagram Competitor routes (analytics — GET, CRUD — POST)
	mux.HandleFunc("GET /analytics/overview/instagram/competitor/dataTableMetrics", igCompetitorHandler.HandleDataTableMetrics)
	mux.HandleFunc("GET /analytics/overview/instagram/competitor/postingActivityGraphByTypes", igCompetitorHandler.HandlePostingActivityGraphByTypes)
	mux.HandleFunc("GET /analytics/overview/instagram/competitor/postingActivityBySpecificType", igCompetitorHandler.HandlePostingActivityBySpecificType)
	mux.HandleFunc("GET /analytics/overview/instagram/competitor/postingActivityTableByType", igCompetitorHandler.HandlePostingActivityTableByType)
	mux.HandleFunc("GET /analytics/overview/instagram/competitor/followersGrowthComparison", igCompetitorHandler.HandleFollowersGrowthComparison)
	mux.HandleFunc("GET /analytics/overview/instagram/competitor/topAndLeastPerformingPosts", igCompetitorHandler.HandleTopAndLeastPerformingPosts)
	mux.HandleFunc("GET /analytics/overview/instagram/competitor/topHashtags", igCompetitorHandler.HandleTopHashtags)
	mux.HandleFunc("GET /analytics/overview/instagram/competitor/individualHashtagData", igCompetitorHandler.HandleIndividualHashtagData)
	mux.HandleFunc("GET /analytics/overview/instagram/competitor/biographyData", igCompetitorHandler.HandleBiographyData)
}
