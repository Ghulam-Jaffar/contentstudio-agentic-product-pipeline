package conversions

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// mockClickHouseClient implements ClickHouseClientInterface for testing
type mockClickHouseClient struct {
	bulkInsertPostsFunc                    func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error
	bulkInsertMediaAssetsFunc              func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error
	bulkInsertVideoInsightsFunc            func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error
	bulkInsertReelsInsightsFunc            func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error
	bulkInsertInsightsFunc                 func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error
	bulkInsertInstagramPostsFunc           func(ctx context.Context, posts []*clickhousemodels.InstagramPost) error
	bulkInsertInstagramInsightsFunc        func(ctx context.Context, insights []*clickhousemodels.InstagramInsight) error
	bulkInsertLinkedInPostsFunc            func(ctx context.Context, posts []*clickhousemodels.LinkedInPosts) error
	bulkInsertLinkedInInsightsFunc         func(ctx context.Context, insights []*clickhousemodels.LinkedInInsights) error
	bulkInsertTikTokPostsFunc              func(ctx context.Context, posts []*clickhousemodels.TikTokPosts) error
	bulkInsertTikTokInsightsFunc           func(ctx context.Context, insights []*clickhousemodels.TikTokInsights) error
	bulkInsertGMBDailyMetricsFunc          func(ctx context.Context, metrics []*clickhousemodels.GMBDailyMetrics) error
	bulkInsertGMBMediaAssetsFunc           func(ctx context.Context, assets []*clickhousemodels.GMBMediaAssets) error
	bulkInsertGMBSearchKeywordsMonthlyFunc func(ctx context.Context, keywords []*clickhousemodels.GMBSearchKeywordsMonthly) error
	bulkInsertGMBLocalPostsFunc            func(ctx context.Context, posts []*clickhousemodels.GMBLocalPosts) error
	bulkInsertGMBReviewsFunc               func(ctx context.Context, reviews []*clickhousemodels.GMBReviews) error
	healthFunc                             func() error
	getMinimalOlderThan20DaysFunc          func(ctx context.Context, tableName, pageID string, limit, offset int) ([]clickhousemodels.MinimalPost, error)
	updateFullPicturesFunc                 func(ctx context.Context, tableName, pageID string, posts []clickhousemodels.MinimalPost) (int, error)
	getMinimalInstagramOlderThan20DaysFunc func(ctx context.Context, tableName, instagramID string, limit, offset int) ([]clickhousemodels.InstagramMinimalPost, error)
	updateInstagramMediaURLsFunc           func(ctx context.Context, tableName, instagramID string, posts []clickhousemodels.InstagramMinimalPost) (int, error)
	getMinimalLinkedInOlderThan7DaysFunc   func(ctx context.Context, tableName, linkedinID string, limit, offset int) ([]clickhousemodels.LinkedInMinimalPost, error)
	updateLinkedInPostURLsFunc             func(ctx context.Context, tableName, linkedinID string, posts []clickhousemodels.LinkedInMinimalPost) (int, error)
}

func (m *mockClickHouseClient) BulkInsertPosts(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
	if m.bulkInsertPostsFunc != nil {
		return m.bulkInsertPostsFunc(ctx, posts)
	}
	return nil
}

func (m *mockClickHouseClient) BulkInsertMediaAssets(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
	if m.bulkInsertMediaAssetsFunc != nil {
		return m.bulkInsertMediaAssetsFunc(ctx, assets)
	}
	return nil
}

func (m *mockClickHouseClient) BulkInsertVideoInsights(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
	if m.bulkInsertVideoInsightsFunc != nil {
		return m.bulkInsertVideoInsightsFunc(ctx, insights)
	}
	return nil
}

func (m *mockClickHouseClient) BulkInsertReelsInsights(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
	if m.bulkInsertReelsInsightsFunc != nil {
		return m.bulkInsertReelsInsightsFunc(ctx, insights)
	}
	return nil
}

func (m *mockClickHouseClient) BulkInsertInsights(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
	if m.bulkInsertInsightsFunc != nil {
		return m.bulkInsertInsightsFunc(ctx, insights)
	}
	return nil
}

func (m *mockClickHouseClient) BulkInsertInstagramPosts(ctx context.Context, posts []*clickhousemodels.InstagramPost) error {
	if m.bulkInsertInstagramPostsFunc != nil {
		return m.bulkInsertInstagramPostsFunc(ctx, posts)
	}
	return nil
}

func (m *mockClickHouseClient) BulkInsertInstagramInsights(ctx context.Context, insights []*clickhousemodels.InstagramInsight) error {
	if m.bulkInsertInstagramInsightsFunc != nil {
		return m.bulkInsertInstagramInsightsFunc(ctx, insights)
	}
	return nil
}

func (m *mockClickHouseClient) BulkInsertLinkedInPosts(ctx context.Context, posts []*clickhousemodels.LinkedInPosts) error {
	if m.bulkInsertLinkedInPostsFunc != nil {
		return m.bulkInsertLinkedInPostsFunc(ctx, posts)
	}
	return nil
}

func (m *mockClickHouseClient) BulkInsertLinkedInInsights(ctx context.Context, insights []*clickhousemodels.LinkedInInsights) error {
	if m.bulkInsertLinkedInInsightsFunc != nil {
		return m.bulkInsertLinkedInInsightsFunc(ctx, insights)
	}
	return nil
}

func (m *mockClickHouseClient) BulkInsertTikTokPosts(ctx context.Context, posts []*clickhousemodels.TikTokPosts) error {
	if m.bulkInsertTikTokPostsFunc != nil {
		return m.bulkInsertTikTokPostsFunc(ctx, posts)
	}
	return nil
}

func (m *mockClickHouseClient) BulkInsertTikTokInsights(ctx context.Context, insights []*clickhousemodels.TikTokInsights) error {
	if m.bulkInsertTikTokInsightsFunc != nil {
		return m.bulkInsertTikTokInsightsFunc(ctx, insights)
	}
	return nil
}

func (m *mockClickHouseClient) Health() error {
	if m.healthFunc != nil {
		return m.healthFunc()
	}
	return nil
}

func (m *mockClickHouseClient) GetMinimalOlderThan20DaysByPage(ctx context.Context, tableName, pageID string, limit, offset int) ([]clickhousemodels.MinimalPost, error) {
	if m.getMinimalOlderThan20DaysFunc != nil {
		return m.getMinimalOlderThan20DaysFunc(ctx, tableName, pageID, limit, offset)
	}
	return nil, nil
}

func (m *mockClickHouseClient) UpdateFullPictures(ctx context.Context, tableName, pageID string, posts []clickhousemodels.MinimalPost) (int, error) {
	if m.updateFullPicturesFunc != nil {
		return m.updateFullPicturesFunc(ctx, tableName, pageID, posts)
	}
	return 0, nil
}

func (m *mockClickHouseClient) BulkUpdateFullPictures(ctx context.Context, tableName string, posts []clickhousemodels.MinimalPost) (int, error) {
	return 0, nil
}

func (m *mockClickHouseClient) GetMinimalInstagramOlderThan20DaysByAccount(ctx context.Context, tableName, instagramID string, limit, offset int) ([]clickhousemodels.InstagramMinimalPost, error) {
	if m.getMinimalInstagramOlderThan20DaysFunc != nil {
		return m.getMinimalInstagramOlderThan20DaysFunc(ctx, tableName, instagramID, limit, offset)
	}
	return nil, nil
}

func (m *mockClickHouseClient) UpdateInstagramMediaURLs(ctx context.Context, tableName, instagramID string, posts []clickhousemodels.InstagramMinimalPost) (int, error) {
	if m.updateInstagramMediaURLsFunc != nil {
		return m.updateInstagramMediaURLsFunc(ctx, tableName, instagramID, posts)
	}
	return 0, nil
}

func (m *mockClickHouseClient) BulkUpdateInstagramMediaURLs(ctx context.Context, tableName string, posts []clickhousemodels.InstagramMinimalPost) (int, error) {
	return 0, nil
}

func (m *mockClickHouseClient) GetMinimalLinkedInOlderThan7DaysByAccount(ctx context.Context, tableName, linkedinID string, limit, offset int) ([]clickhousemodels.LinkedInMinimalPost, error) {
	if m.getMinimalLinkedInOlderThan7DaysFunc != nil {
		return m.getMinimalLinkedInOlderThan7DaysFunc(ctx, tableName, linkedinID, limit, offset)
	}
	return nil, nil
}

func (m *mockClickHouseClient) UpdateLinkedInPostURLs(ctx context.Context, tableName, linkedinID string, posts []clickhousemodels.LinkedInMinimalPost) (int, error) {
	if m.updateLinkedInPostURLsFunc != nil {
		return m.updateLinkedInPostURLsFunc(ctx, tableName, linkedinID, posts)
	}
	return 0, nil
}

func (m *mockClickHouseClient) BulkUpdateLinkedInPostURLs(ctx context.Context, tableName string, posts []clickhousemodels.LinkedInMinimalPost) (int, error) {
	return 0, nil
}

func (m *mockClickHouseClient) GetDistinctFacebookPageIDsWithStaleURLs(ctx context.Context, tableName string, _ []string) ([]string, error) {
	return nil, nil
}

func (m *mockClickHouseClient) GetDistinctInstagramIDsWithStaleURLs(ctx context.Context, tableName string, _ []string) ([]string, error) {
	return nil, nil
}

func (m *mockClickHouseClient) GetDistinctLinkedInIDsWithStaleURLs(ctx context.Context, tableName string, _ []string) ([]string, error) {
	return nil, nil
}

func (m *mockClickHouseClient) GetMinimalFacebookCompetitorMediaAssetsOlderThan7DaysByAccount(ctx context.Context, tableName, facebookID string, limit, offset int) ([]clickhousemodels.FacebookCompetitorMinimalMediaAsset, error) {
	return nil, nil
}

func (m *mockClickHouseClient) UpdateFacebookCompetitorMediaAssetURLs(ctx context.Context, tableName, facebookID string, assets []clickhousemodels.FacebookCompetitorMinimalMediaAsset) (int, error) {
	return 0, nil
}

func (m *mockClickHouseClient) GetMinimalFacebookCompetitorSharedPostsOlderThan7DaysByAccount(ctx context.Context, tableName, facebookID string, limit, offset int) ([]clickhousemodels.FacebookCompetitorMinimalSharedPost, error) {
	return nil, nil
}

func (m *mockClickHouseClient) UpdateFacebookCompetitorSharedPictures(ctx context.Context, tableName, facebookID string, posts []clickhousemodels.FacebookCompetitorMinimalSharedPost) (int, error) {
	return 0, nil
}

func (m *mockClickHouseClient) GetMinimalInstagramCompetitorOlderThan7DaysByAccount(ctx context.Context, tableName string, instagramID int64, limit, offset int) ([]clickhousemodels.InstagramCompetitorMinimalPost, error) {
	return nil, nil
}

func (m *mockClickHouseClient) UpdateInstagramCompetitorMediaURLs(ctx context.Context, tableName string, instagramID int64, profilePictureURL string, posts []clickhousemodels.InstagramCompetitorMinimalPost) (int, error) {
	return 0, nil
}

func (m *mockClickHouseClient) BulkUpdateFacebookCompetitorMediaAssetURLs(ctx context.Context, tableName string, assets []clickhousemodels.FacebookCompetitorMinimalMediaAsset) (int, error) {
	return 0, nil
}

func (m *mockClickHouseClient) BulkUpdateFacebookCompetitorSharedPictures(ctx context.Context, tableName string, posts []clickhousemodels.FacebookCompetitorMinimalSharedPost) (int, error) {
	return 0, nil
}

func (m *mockClickHouseClient) BulkUpdateInstagramCompetitorMediaURLs(ctx context.Context, tableName string, posts []clickhousemodels.InstagramCompetitorMinimalPost, profilePics map[int64]string) (int, error) {
	return 0, nil
}

// YouTube methods for interface compliance
func (m *mockClickHouseClient) BulkInsertYouTubeChannels(ctx context.Context, channels []*clickhousemodels.YouTubeChannel) error {
	return nil
}

func (m *mockClickHouseClient) BulkInsertYouTubeVideos(ctx context.Context, videos []*clickhousemodels.YouTubeVideo) error {
	return nil
}

func (m *mockClickHouseClient) BulkInsertYouTubeActivityInsights(ctx context.Context, insights []*clickhousemodels.YouTubeActivityInsights) error {
	return nil
}

func (m *mockClickHouseClient) BulkInsertYouTubeTrafficInsights(ctx context.Context, insights []*clickhousemodels.YouTubeTrafficInsights) error {
	return nil
}

func (m *mockClickHouseClient) BulkInsertYouTubeSharedInsights(ctx context.Context, insights []*clickhousemodels.YouTubeSharedInsights) error {
	return nil
}

// Twitter methods for interface compliance
func (m *mockClickHouseClient) BulkInsertTwitterPosts(ctx context.Context, posts []*clickhousemodels.TwitterPosts) error {
	return nil
}

func (m *mockClickHouseClient) BulkInsertTwitterInsights(ctx context.Context, insights []*clickhousemodels.TwitterInsights) error {
	return nil
}

// Pinterest methods for interface compliance
func (m *mockClickHouseClient) BulkInsertPinterestUsers(ctx context.Context, users []clickhousemodels.PinterestUser) error {
	return nil
}

func (m *mockClickHouseClient) BulkInsertPinterestBoards(ctx context.Context, boards []clickhousemodels.PinterestBoard) error {
	return nil
}

func (m *mockClickHouseClient) BulkInsertPinterestPins(ctx context.Context, pins []clickhousemodels.PinterestPin) error {
	return nil
}

func (m *mockClickHouseClient) BulkInsertPinterestPinInsights(ctx context.Context, insights []clickhousemodels.PinterestPinInsight) error {
	return nil
}

func (m *mockClickHouseClient) BulkInsertPinterestUserInsights(ctx context.Context, insights []clickhousemodels.PinterestUserInsight) error {
	return nil
}

// GMB methods for interface compliance
func (m *mockClickHouseClient) BulkInsertGMBDailyMetrics(ctx context.Context, metrics []*clickhousemodels.GMBDailyMetrics) error {
	if m.bulkInsertGMBDailyMetricsFunc != nil {
		return m.bulkInsertGMBDailyMetricsFunc(ctx, metrics)
	}
	return nil
}

func (m *mockClickHouseClient) BulkInsertGMBMediaAssets(ctx context.Context, assets []*clickhousemodels.GMBMediaAssets) error {
	if m.bulkInsertGMBMediaAssetsFunc != nil {
		return m.bulkInsertGMBMediaAssetsFunc(ctx, assets)
	}
	return nil
}

func (m *mockClickHouseClient) BulkInsertGMBSearchKeywordsMonthly(ctx context.Context, keywords []*clickhousemodels.GMBSearchKeywordsMonthly) error {
	if m.bulkInsertGMBSearchKeywordsMonthlyFunc != nil {
		return m.bulkInsertGMBSearchKeywordsMonthlyFunc(ctx, keywords)
	}
	return nil
}

func (m *mockClickHouseClient) BulkInsertGMBLocalPosts(ctx context.Context, posts []*clickhousemodels.GMBLocalPosts) error {
	if m.bulkInsertGMBLocalPostsFunc != nil {
		return m.bulkInsertGMBLocalPostsFunc(ctx, posts)
	}
	return nil
}

func (m *mockClickHouseClient) BulkInsertGMBReviews(ctx context.Context, reviews []*clickhousemodels.GMBReviews) error {
	if m.bulkInsertGMBReviewsFunc != nil {
		return m.bulkInsertGMBReviewsFunc(ctx, reviews)
	}
	return nil
}

// Meta Ads methods for interface compliance
func (m *mockClickHouseClient) BulkInsertMetaAdsAccountInfo(ctx context.Context, rows []*clickhousemodels.MetaAdsAccountInfo) error {
	return nil
}

func (m *mockClickHouseClient) BulkInsertMetaAdsCampaigns(ctx context.Context, rows []*clickhousemodels.MetaAdsCampaign) error {
	return nil
}

func (m *mockClickHouseClient) BulkInsertMetaAdsAdsets(ctx context.Context, rows []*clickhousemodels.MetaAdsAdset) error {
	return nil
}

func (m *mockClickHouseClient) BulkInsertMetaAdsAds(ctx context.Context, rows []*clickhousemodels.MetaAdsAd) error {
	return nil
}

func (m *mockClickHouseClient) BulkInsertMetaAdsCampaignInsights(ctx context.Context, rows []*clickhousemodels.MetaAdsCampaignInsights) error {
	return nil
}

func (m *mockClickHouseClient) BulkInsertMetaAdsAdsetInsights(ctx context.Context, rows []*clickhousemodels.MetaAdsAdsetInsights) error {
	return nil
}

func (m *mockClickHouseClient) BulkInsertMetaAdsAdInsights(ctx context.Context, rows []*clickhousemodels.MetaAdsAdInsights) error {
	return nil
}

func (m *mockClickHouseClient) BulkInsertMetaAdsDemographicsAgeGender(ctx context.Context, rows []*clickhousemodels.MetaAdsDemographicsAgeGender) error {
	return nil
}

func (m *mockClickHouseClient) BulkInsertMetaAdsDemographicsDevicePlatform(ctx context.Context, rows []*clickhousemodels.MetaAdsDemographicsDevicePlatform) error {
	return nil
}

func (m *mockClickHouseClient) BulkInsertMetaAdsDemographicsRegionCountry(ctx context.Context, rows []*clickhousemodels.MetaAdsDemographicsRegionCountry) error {
	return nil
}

func (m *mockClickHouseClient) MarkFacebookPostsRefreshed(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockClickHouseClient) BulkMarkFacebookPostsRefreshed(_ context.Context, _ string, _ []string) error {
	return nil
}

func (m *mockClickHouseClient) MarkInstagramPostsRefreshed(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockClickHouseClient) BulkMarkInstagramPostsRefreshed(_ context.Context, _ string, _ []string) error {
	return nil
}

func (m *mockClickHouseClient) MarkLinkedInPostsRefreshed(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockClickHouseClient) BulkMarkLinkedInPostsRefreshed(_ context.Context, _ string, _ []string) error {
	return nil
}

func newTestLogger() *zerolog.Logger {
	logger := zerolog.Nop()
	return &logger
}

func newTestSink(mockClient *mockClickHouseClient) *ClickHouseSink {
	logger := newTestLogger()
	return &ClickHouseSink{
		logger:           logger,
		ClickhouseClient: mockClient,
	}
}

// Tests for BulkInsert methods

func TestBulkInsertPosts_EmptySlice(t *testing.T) {
	sink := newTestSink(&mockClickHouseClient{})
	err := sink.BulkInsertPosts(context.Background(), []*clickhousemodels.FacebookPosts{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func TestBulkInsertPosts_Success(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			called = true
			if len(posts) != 1 {
				t.Fatalf("expected 1 post, got %d", len(posts))
			}
			return nil
		},
	}
	sink := newTestSink(mock)

	posts := []*clickhousemodels.FacebookPosts{{PageID: "page123", PostID: "post456"}}
	err := sink.BulkInsertPosts(context.Background(), posts)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertPosts to be called")
	}
}

func TestBulkInsertMediaAssets_EmptySlice(t *testing.T) {
	sink := newTestSink(&mockClickHouseClient{})
	err := sink.BulkInsertMediaAssets(context.Background(), []*clickhousemodels.FacebookMediaAssets{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func TestBulkInsertMediaAssets_Success(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertMediaAssetsFunc: func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
			called = true
			return nil
		},
	}
	sink := newTestSink(mock)

	assets := []*clickhousemodels.FacebookMediaAssets{{PageID: "page123"}}
	err := sink.BulkInsertMediaAssets(context.Background(), assets)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertMediaAssets to be called")
	}
}

func TestBulkInsertVideoInsights_EmptySlice(t *testing.T) {
	sink := newTestSink(&mockClickHouseClient{})
	err := sink.BulkInsertVideoInsights(context.Background(), []*clickhousemodels.FacebookVideoInsights{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func TestBulkInsertVideoInsights_Success(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertVideoInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
			called = true
			return nil
		},
	}
	sink := newTestSink(mock)

	insights := []*clickhousemodels.FacebookVideoInsights{{VideoID: "video123"}}
	err := sink.BulkInsertVideoInsights(context.Background(), insights)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertVideoInsights to be called")
	}
}

func TestBulkInsertReelsInsights_EmptySlice(t *testing.T) {
	sink := newTestSink(&mockClickHouseClient{})
	err := sink.BulkInsertReelsInsights(context.Background(), []*clickhousemodels.FacebookReelsInsights{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func TestBulkInsertReelsInsights_Success(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertReelsInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
			called = true
			return nil
		},
	}
	sink := newTestSink(mock)

	insights := []*clickhousemodels.FacebookReelsInsights{{PostID: "post123"}}
	err := sink.BulkInsertReelsInsights(context.Background(), insights)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertReelsInsights to be called")
	}
}

func TestBulkInsertInsights_EmptySlice(t *testing.T) {
	sink := newTestSink(&mockClickHouseClient{})
	err := sink.BulkInsertInsights(context.Background(), []*clickhousemodels.FacebookInsights{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func TestBulkInsertInsights_Success(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
			called = true
			return nil
		},
	}
	sink := newTestSink(mock)

	insights := []*clickhousemodels.FacebookInsights{{PageID: "page123"}}
	err := sink.BulkInsertInsights(context.Background(), insights)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertInsights to be called")
	}
}

func TestHealth_Success(t *testing.T) {
	mock := &mockClickHouseClient{
		healthFunc: func() error {
			return nil
		},
	}
	sink := newTestSink(mock)

	err := sink.Health()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestHealth_Error(t *testing.T) {
	expectedErr := errors.New("connection failed")
	mock := &mockClickHouseClient{
		healthFunc: func() error {
			return expectedErr
		},
	}
	sink := newTestSink(mock)

	err := sink.Health()
	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestProcessParsedData_SingleFacebookPost(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			called = true
			return nil
		},
	}
	sink := newTestSink(mock)

	post := &kafkamodels.ParsedFacebookPost{PageID: "page123", PostID: "post456"}
	err := sink.ProcessParsedData(context.Background(), post)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertPosts to be called")
	}
}

func TestProcessParsedData_MultipleFacebookPosts(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			called = true
			if len(posts) != 2 {
				t.Fatalf("expected 2 posts, got %d", len(posts))
			}
			return nil
		},
	}
	sink := newTestSink(mock)

	posts := []*kafkamodels.ParsedFacebookPost{
		{PageID: "page1", PostID: "post1"},
		{PageID: "page2", PostID: "post2"},
	}
	err := sink.ProcessParsedData(context.Background(), posts)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertPosts to be called")
	}
}

func TestProcessParsedData_SingleVideoInsights(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertVideoInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
			called = true
			return nil
		},
	}
	sink := newTestSink(mock)

	insights := &kafkamodels.ParsedFacebookVideoInsights{VideoID: "video123"}
	err := sink.ProcessParsedData(context.Background(), insights)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertVideoInsights to be called")
	}
}

func TestProcessParsedData_MultipleVideoInsights(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertVideoInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
			called = true
			return nil
		},
	}
	sink := newTestSink(mock)

	insights := []*kafkamodels.ParsedFacebookVideoInsights{
		{VideoID: "video1"},
		{VideoID: "video2"},
	}
	err := sink.ProcessParsedData(context.Background(), insights)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertVideoInsights to be called")
	}
}

func TestProcessParsedData_SingleReelsInsights(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertReelsInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
			called = true
			return nil
		},
	}
	sink := newTestSink(mock)

	insights := &kafkamodels.ParsedFacebookReelsInsights{PostID: "post123"}
	err := sink.ProcessParsedData(context.Background(), insights)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertReelsInsights to be called")
	}
}

func TestProcessParsedData_MultipleReelsInsights(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertReelsInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
			called = true
			return nil
		},
	}
	sink := newTestSink(mock)

	insights := []*kafkamodels.ParsedFacebookReelsInsights{
		{PostID: "post1"},
		{PostID: "post2"},
	}
	err := sink.ProcessParsedData(context.Background(), insights)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertReelsInsights to be called")
	}
}

func TestProcessParsedData_SinglePageInsights(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
			called = true
			return nil
		},
	}
	sink := newTestSink(mock)

	insights := &kafkamodels.ParsedFacebookInsights{PageID: "page123"}
	err := sink.ProcessParsedData(context.Background(), insights)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertInsights to be called")
	}
}

func TestProcessParsedData_MultiplePageInsights(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
			called = true
			return nil
		},
	}
	sink := newTestSink(mock)

	insights := []*kafkamodels.ParsedFacebookInsights{
		{PageID: "page1"},
		{PageID: "page2"},
	}
	err := sink.ProcessParsedData(context.Background(), insights)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertInsights to be called")
	}
}

func TestProcessParsedData_SingleMediaAsset(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertMediaAssetsFunc: func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
			called = true
			return nil
		},
	}
	sink := newTestSink(mock)

	asset := &kafkamodels.ParsedFacebookMediaAsset{MediaID: "media123"}
	err := sink.ProcessParsedData(context.Background(), asset)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertMediaAssets to be called")
	}
}

func TestProcessParsedData_MultipleMediaAssets(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertMediaAssetsFunc: func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
			called = true
			return nil
		},
	}
	sink := newTestSink(mock)

	assets := []*kafkamodels.ParsedFacebookMediaAsset{
		{MediaID: "media1"},
		{MediaID: "media2"},
	}
	err := sink.ProcessParsedData(context.Background(), assets)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertMediaAssets to be called")
	}
}

func TestProcessParsedData_UnsupportedType(t *testing.T) {
	sink := newTestSink(&mockClickHouseClient{})

	err := sink.ProcessParsedData(context.Background(), "unsupported string type")
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

func TestProcessParsedData_NilPostsFiltered(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			called = true
			if len(posts) != 1 {
				t.Fatalf("expected 1 non-nil post, got %d", len(posts))
			}
			return nil
		},
	}
	sink := newTestSink(mock)

	posts := []*kafkamodels.ParsedFacebookPost{
		{PageID: "page1"},
		nil,
	}
	err := sink.ProcessParsedData(context.Background(), posts)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertPosts to be called")
	}
}

func TestHandleFailedInsert(t *testing.T) {
	sink := newTestSink(&mockClickHouseClient{})
	testErr := errors.New("insert failed")

	// Just verify it doesn't panic
	sink.HandleFailedInsert(context.Background(), "test data", testErr)
}

func TestConvertStringArrayFields(t *testing.T) {
	cases := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "nil input returns empty slice",
			input:    nil,
			expected: []string{},
		},
		{
			name:     "empty input returns empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "filters out empty strings",
			input:    []string{"a", "", "b", "  ", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "keeps non-empty strings",
			input:    []string{"US:1000", "UK:500", "CA:300"},
			expected: []string{"US:1000", "UK:500", "CA:300"},
		},
		{
			name:     "filters whitespace-only strings",
			input:    []string{"data1", "   ", "\t", "data2"},
			expected: []string{"data1", "data2"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := convertStringArrayFields(tc.input)
			if len(result) != len(tc.expected) {
				t.Fatalf("expected length %d, got %d", len(tc.expected), len(result))
			}
			for i, v := range result {
				if v != tc.expected[i] {
					t.Fatalf("at index %d: expected %q, got %q", i, tc.expected[i], v)
				}
			}
		})
	}
}

func TestConvertFacebookPost_NilInput(t *testing.T) {
	sink := &ClickHouseSink{}
	result := sink.ConvertFacebookPost(nil)
	if result != nil {
		t.Fatal("expected nil result for nil input")
	}
}

func TestConvertFacebookPost_ValidInput(t *testing.T) {
	sink := &ClickHouseSink{}
	now := time.Now()

	input := &kafkamodels.ParsedFacebookPost{
		PageName:        "Test Page",
		PageID:          "page123",
		MediaType:       "photo",
		PostID:          "post456",
		Permalink:       "https://facebook.com/post/456",
		StatusType:      "added_photos",
		VideoID:         "video789",
		Category:        "Business",
		PublishedBy:     "Admin",
		Like:            100,
		Love:            50,
		Haha:            25,
		Wow:             10,
		Sad:             5,
		Angry:           2,
		Thankful:        1,
		Total:           193,
		Shares:          30,
		Comments:        45,
		TotalEngagement: 268,
		DayOfWeek:       "Monday",
		HourOfDay:       14,
		CreatedTime:     now,
		MessageTags:     []string{"#test"},
		Caption:         "Test caption",
		PostImpressions: 10000,
	}

	result := sink.ConvertFacebookPost(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.PageID != "page123" {
		t.Fatalf("expected PageID 'page123', got %s", result.PageID)
	}
	if result.PostID != "post456" {
		t.Fatalf("expected PostID 'post456', got %s", result.PostID)
	}
	if result.Like != 100 {
		t.Fatalf("expected Like 100, got %d", result.Like)
	}
	if result.Total != 193 {
		t.Fatalf("expected Total 193, got %d", result.Total)
	}
	if result.TotalEngagement != 268 {
		t.Fatalf("expected TotalEngagement 268, got %d", result.TotalEngagement)
	}
}

func TestConvertFacebookMediaAssets_NilInput(t *testing.T) {
	sink := &ClickHouseSink{}
	result := sink.ConvertFacebookMediaAssets(nil)
	if result != nil {
		t.Fatal("expected nil result for nil input")
	}
}

func TestConvertFacebookMediaAssets_ValidInput(t *testing.T) {
	sink := &ClickHouseSink{}
	now := time.Now()

	input := &kafkamodels.ParsedFacebookMediaAsset{
		PageID:       "page123",
		MediaID:      "media456",
		PostID:       "post789",
		AssetType:    "image",
		Link:         "https://example.com/image.jpg",
		CallToAction: "Learn More",
		CTAType:      "LEARN_MORE",
		Caption:      "Test caption",
		Description:  "Test description",
		CreatedAt:    now,
		InsertedAt:   now,
	}

	result := sink.ConvertFacebookMediaAssets(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.PageID != "page123" {
		t.Fatalf("expected PageID 'page123', got %s", result.PageID)
	}
	if result.MediaID != "media456" {
		t.Fatalf("expected MediaID 'media456', got %s", result.MediaID)
	}
	if result.AssetType != "image" {
		t.Fatalf("expected AssetType 'image', got %s", result.AssetType)
	}
}

func TestConvertFacebookVideoInsights_NilInput(t *testing.T) {
	sink := &ClickHouseSink{}
	result := sink.ConvertFacebookVideoInsights(nil)
	if result != nil {
		t.Fatal("expected nil result for nil input")
	}
}

func TestConvertFacebookVideoInsights_ValidInput(t *testing.T) {
	sink := &ClickHouseSink{}
	now := time.Now()

	input := &kafkamodels.ParsedFacebookVideoInsights{
		PostID:                            "post123",
		PageID:                            "page456",
		VideoID:                           "video789",
		CreatedTime:                       now,
		TotalVideoViews:                   10000,
		TotalVideoViewsUnique:             8000,
		TotalVideoViewsOrganic:            7000,
		TotalVideoViewsPaid:               3000,
		TotalVideoCompleteViews:           5000,
		TotalVideoImpressions:             50000,
		TotalEngagement:                   15000,
		TotalVideoConsumptionRate:         0.75,
		TotalVideoAdBreakEarnings:         150.50,
		TotalVideoAdBreakAdCPM:            5.25,
		TotalVideoViewsByDistributionType: []string{"organic:7000", "paid:3000"},
	}

	result := sink.ConvertFacebookVideoInsights(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.VideoID != "video789" {
		t.Fatalf("expected VideoID 'video789', got %s", result.VideoID)
	}
	if result.TotalVideoViews != 10000 {
		t.Fatalf("expected TotalVideoViews 10000, got %d", result.TotalVideoViews)
	}
	if result.TotalVideoConsumptionRate != 0.75 {
		t.Fatalf("expected TotalVideoConsumptionRate 0.75, got %f", result.TotalVideoConsumptionRate)
	}
}

func TestConvertFacebookReelsInsights_NilInput(t *testing.T) {
	sink := &ClickHouseSink{}
	result := sink.ConvertFacebookReelsInsights(nil)
	if result != nil {
		t.Fatal("expected nil result for nil input")
	}
}

func TestConvertFacebookReelsInsights_ValidInput(t *testing.T) {
	sink := &ClickHouseSink{}
	now := time.Now()

	input := &kafkamodels.ParsedFacebookReelsInsights{
		PageID:               "page123",
		PostID:               "post456",
		AverageTimeWatched:   15000,
		TotalTimeWatchedInMs: 500000,
		PlayCount:            1000,
		ImpressionsUnique:    800,
		ReelFollowers:        50,
		CreatedAt:            now,
		SavingTime:           now,
	}

	result := sink.ConvertFacebookReelsInsights(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.PageID != "page123" {
		t.Fatalf("expected PageID 'page123', got %s", result.PageID)
	}
	if result.PlayCount != 1000 {
		t.Fatalf("expected PlayCount 1000, got %d", result.PlayCount)
	}
}

func TestConvertFacebookInsights_NilInput(t *testing.T) {
	sink := &ClickHouseSink{}
	result := sink.ConvertFacebookInsights(nil)
	if result != nil {
		t.Fatal("expected nil result for nil input")
	}
}

func TestConvertFacebookInsights_ValidInput(t *testing.T) {
	sink := &ClickHouseSink{}
	now := time.Now()

	input := &kafkamodels.ParsedFacebookInsights{
		HashID:          "hash123",
		PageID:          "page456",
		PageCategory:    "Business",
		DayOfWeek:       "Monday",
		Year:            2025,
		Month:           1,
		CreatedTime:     now,
		PageFans:        10000,
		PageFollows:     9500,
		PageViews:       50000,
		PageImpressions: 100000,
		PageVideoViews:  25000,
		PageFansCity:    []string{"NYC:1000", "LA:800"},
		PageFansCountry: []string{"US:5000", "UK:2000"},
		PageFansAge:     []string{"25-34:3000"},
	}

	result := sink.ConvertFacebookInsights(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.HashID != "hash123" {
		t.Fatalf("expected HashID 'hash123', got %s", result.HashID)
	}
	if result.PageFans != 10000 {
		t.Fatalf("expected PageFans 10000, got %d", result.PageFans)
	}
	if result.Year != 2025 {
		t.Fatalf("expected Year 2025, got %d", result.Year)
	}
	if len(result.PageFansCity) != 2 {
		t.Fatalf("expected 2 cities, got %d", len(result.PageFansCity))
	}
}

func TestConvertFacebookInsights_WithEmptyArrays(t *testing.T) {
	sink := &ClickHouseSink{}

	input := &kafkamodels.ParsedFacebookInsights{
		HashID:          "hash123",
		PageID:          "page456",
		PageFansCity:    []string{"NYC:1000", "", "LA:800", "  "},
		PageFansCountry: nil,
	}

	result := sink.ConvertFacebookInsights(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.PageFansCity) != 2 {
		t.Fatalf("expected 2 non-empty cities, got %d", len(result.PageFansCity))
	}
	if len(result.PageFansCountry) != 0 {
		t.Fatalf("expected 0 countries (empty slice), got %d", len(result.PageFansCountry))
	}
}
