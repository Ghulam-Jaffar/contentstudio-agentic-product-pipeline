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

// mockClickHouseClientTwitter implements ClickHouseClientInterface for Twitter testing
type mockClickHouseClientTwitter struct {
	bulkInsertTwitterPostsFunc    func(ctx context.Context, posts []*clickhousemodels.TwitterPosts) error
	bulkInsertTwitterInsightsFunc func(ctx context.Context, insights []*clickhousemodels.TwitterInsights) error
	bulkInsertYouTubeChannelsFunc func(ctx context.Context, channels []*clickhousemodels.YouTubeChannel) error
	healthFunc                    func() error
}

func (m *mockClickHouseClientTwitter) BulkInsertPosts(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertMediaAssets(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertVideoInsights(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertReelsInsights(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertInsights(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertInstagramPosts(ctx context.Context, posts []*clickhousemodels.InstagramPost) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertInstagramInsights(ctx context.Context, insights []*clickhousemodels.InstagramInsight) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertLinkedInPosts(ctx context.Context, posts []*clickhousemodels.LinkedInPosts) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertLinkedInInsights(ctx context.Context, insights []*clickhousemodels.LinkedInInsights) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertTikTokPosts(ctx context.Context, posts []*clickhousemodels.TikTokPosts) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertTikTokInsights(ctx context.Context, insights []*clickhousemodels.TikTokInsights) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertTwitterPosts(ctx context.Context, posts []*clickhousemodels.TwitterPosts) error {
	if m.bulkInsertTwitterPostsFunc != nil {
		return m.bulkInsertTwitterPostsFunc(ctx, posts)
	}
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertTwitterInsights(ctx context.Context, insights []*clickhousemodels.TwitterInsights) error {
	if m.bulkInsertTwitterInsightsFunc != nil {
		return m.bulkInsertTwitterInsightsFunc(ctx, insights)
	}
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertYouTubeChannels(ctx context.Context, channels []*clickhousemodels.YouTubeChannel) error {
	if m.bulkInsertYouTubeChannelsFunc != nil {
		return m.bulkInsertYouTubeChannelsFunc(ctx, channels)
	}
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertYouTubeVideos(ctx context.Context, videos []*clickhousemodels.YouTubeVideo) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertYouTubeActivityInsights(ctx context.Context, insights []*clickhousemodels.YouTubeActivityInsights) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertYouTubeTrafficInsights(ctx context.Context, insights []*clickhousemodels.YouTubeTrafficInsights) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertYouTubeSharedInsights(ctx context.Context, insights []*clickhousemodels.YouTubeSharedInsights) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertPinterestUsers(ctx context.Context, users []clickhousemodels.PinterestUser) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertPinterestBoards(ctx context.Context, boards []clickhousemodels.PinterestBoard) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertPinterestPins(ctx context.Context, pins []clickhousemodels.PinterestPin) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertPinterestPinInsights(ctx context.Context, insights []clickhousemodels.PinterestPinInsight) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertPinterestUserInsights(ctx context.Context, insights []clickhousemodels.PinterestUserInsight) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertGMBDailyMetrics(ctx context.Context, metrics []*clickhousemodels.GMBDailyMetrics) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertGMBMediaAssets(ctx context.Context, assets []*clickhousemodels.GMBMediaAssets) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertGMBSearchKeywordsMonthly(ctx context.Context, keywords []*clickhousemodels.GMBSearchKeywordsMonthly) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertGMBLocalPosts(ctx context.Context, posts []*clickhousemodels.GMBLocalPosts) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertGMBReviews(ctx context.Context, reviews []*clickhousemodels.GMBReviews) error {
	return nil
}

func (m *mockClickHouseClientTwitter) GetMinimalOlderThan20DaysByPage(ctx context.Context, tableName, pageID string, limit, offset int) ([]clickhousemodels.MinimalPost, error) {
	return nil, nil
}

func (m *mockClickHouseClientTwitter) UpdateFullPictures(ctx context.Context, tableName, pageID string, posts []clickhousemodels.MinimalPost) (int, error) {
	return 0, nil
}

func (m *mockClickHouseClientTwitter) BulkUpdateFullPictures(ctx context.Context, tableName string, posts []clickhousemodels.MinimalPost) (int, error) {
	return 0, nil
}

func (m *mockClickHouseClientTwitter) GetMinimalInstagramOlderThan20DaysByAccount(ctx context.Context, tableName, instagramID string, limit, offset int) ([]clickhousemodels.InstagramMinimalPost, error) {
	return nil, nil
}

func (m *mockClickHouseClientTwitter) UpdateInstagramMediaURLs(ctx context.Context, tableName, instagramID string, posts []clickhousemodels.InstagramMinimalPost) (int, error) {
	return 0, nil
}

func (m *mockClickHouseClientTwitter) BulkUpdateInstagramMediaURLs(ctx context.Context, tableName string, posts []clickhousemodels.InstagramMinimalPost) (int, error) {
	return 0, nil
}

func (m *mockClickHouseClientTwitter) GetMinimalLinkedInOlderThan7DaysByAccount(ctx context.Context, tableName, linkedinID string, limit, offset int) ([]clickhousemodels.LinkedInMinimalPost, error) {
	return nil, nil
}

func (m *mockClickHouseClientTwitter) UpdateLinkedInPostURLs(ctx context.Context, tableName, linkedinID string, posts []clickhousemodels.LinkedInMinimalPost) (int, error) {
	return 0, nil
}

func (m *mockClickHouseClientTwitter) BulkUpdateLinkedInPostURLs(ctx context.Context, tableName string, posts []clickhousemodels.LinkedInMinimalPost) (int, error) {
	return 0, nil
}

func (m *mockClickHouseClientTwitter) GetDistinctFacebookPageIDsWithStaleURLs(ctx context.Context, tableName string, _ []string) ([]string, error) {
	return nil, nil
}

func (m *mockClickHouseClientTwitter) GetDistinctInstagramIDsWithStaleURLs(ctx context.Context, tableName string, _ []string) ([]string, error) {
	return nil, nil
}

func (m *mockClickHouseClientTwitter) GetDistinctLinkedInIDsWithStaleURLs(ctx context.Context, tableName string, _ []string) ([]string, error) {
	return nil, nil
}

func (m *mockClickHouseClientTwitter) GetMinimalFacebookCompetitorMediaAssetsOlderThan7DaysByAccount(ctx context.Context, tableName, facebookID string, limit, offset int) ([]clickhousemodels.FacebookCompetitorMinimalMediaAsset, error) {
	return nil, nil
}

func (m *mockClickHouseClientTwitter) UpdateFacebookCompetitorMediaAssetURLs(ctx context.Context, tableName, facebookID string, assets []clickhousemodels.FacebookCompetitorMinimalMediaAsset) (int, error) {
	return 0, nil
}

func (m *mockClickHouseClientTwitter) GetMinimalFacebookCompetitorSharedPostsOlderThan7DaysByAccount(ctx context.Context, tableName, facebookID string, limit, offset int) ([]clickhousemodels.FacebookCompetitorMinimalSharedPost, error) {
	return nil, nil
}

func (m *mockClickHouseClientTwitter) UpdateFacebookCompetitorSharedPictures(ctx context.Context, tableName, facebookID string, posts []clickhousemodels.FacebookCompetitorMinimalSharedPost) (int, error) {
	return 0, nil
}

func (m *mockClickHouseClientTwitter) GetMinimalInstagramCompetitorOlderThan7DaysByAccount(ctx context.Context, tableName string, instagramID int64, limit, offset int) ([]clickhousemodels.InstagramCompetitorMinimalPost, error) {
	return nil, nil
}

func (m *mockClickHouseClientTwitter) UpdateInstagramCompetitorMediaURLs(ctx context.Context, tableName string, instagramID int64, profilePictureURL string, posts []clickhousemodels.InstagramCompetitorMinimalPost) (int, error) {
	return 0, nil
}

func (m *mockClickHouseClientTwitter) BulkUpdateFacebookCompetitorMediaAssetURLs(ctx context.Context, tableName string, assets []clickhousemodels.FacebookCompetitorMinimalMediaAsset) (int, error) {
	return 0, nil
}

func (m *mockClickHouseClientTwitter) BulkUpdateFacebookCompetitorSharedPictures(ctx context.Context, tableName string, posts []clickhousemodels.FacebookCompetitorMinimalSharedPost) (int, error) {
	return 0, nil
}

func (m *mockClickHouseClientTwitter) BulkUpdateInstagramCompetitorMediaURLs(ctx context.Context, tableName string, posts []clickhousemodels.InstagramCompetitorMinimalPost, profilePics map[int64]string) (int, error) {
	return 0, nil
}

func (m *mockClickHouseClientTwitter) Health() error {
	if m.healthFunc != nil {
		return m.healthFunc()
	}
	return nil
}

// Meta Ads methods for interface compliance
func (m *mockClickHouseClientTwitter) BulkInsertMetaAdsAccountInfo(ctx context.Context, rows []*clickhousemodels.MetaAdsAccountInfo) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertMetaAdsCampaigns(ctx context.Context, rows []*clickhousemodels.MetaAdsCampaign) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertMetaAdsAdsets(ctx context.Context, rows []*clickhousemodels.MetaAdsAdset) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertMetaAdsAds(ctx context.Context, rows []*clickhousemodels.MetaAdsAd) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertMetaAdsCampaignInsights(ctx context.Context, rows []*clickhousemodels.MetaAdsCampaignInsights) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertMetaAdsAdsetInsights(ctx context.Context, rows []*clickhousemodels.MetaAdsAdsetInsights) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertMetaAdsAdInsights(ctx context.Context, rows []*clickhousemodels.MetaAdsAdInsights) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertMetaAdsDemographicsAgeGender(ctx context.Context, rows []*clickhousemodels.MetaAdsDemographicsAgeGender) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertMetaAdsDemographicsDevicePlatform(ctx context.Context, rows []*clickhousemodels.MetaAdsDemographicsDevicePlatform) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkInsertMetaAdsDemographicsRegionCountry(ctx context.Context, rows []*clickhousemodels.MetaAdsDemographicsRegionCountry) error {
	return nil
}

func (m *mockClickHouseClientTwitter) MarkFacebookPostsRefreshed(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkMarkFacebookPostsRefreshed(_ context.Context, _ string, _ []string) error {
	return nil
}

func (m *mockClickHouseClientTwitter) MarkInstagramPostsRefreshed(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkMarkInstagramPostsRefreshed(_ context.Context, _ string, _ []string) error {
	return nil
}

func (m *mockClickHouseClientTwitter) MarkLinkedInPostsRefreshed(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockClickHouseClientTwitter) BulkMarkLinkedInPostsRefreshed(_ context.Context, _ string, _ []string) error {
	return nil
}

func newTestTwitterSink(client ClickHouseClientInterface) *ClickHouseSink {
	logger := zerolog.Nop()
	return &ClickHouseSink{
		logger:           &logger,
		ClickhouseClient: client,
	}
}

// Tests for ConvertTwitterPost
func TestConvertTwitterPost_NilInput(t *testing.T) {
	result := ConvertTwitterPost(nil)
	if result != nil {
		t.Fatal("expected nil result for nil input")
	}
}

func TestConvertTwitterPost_ValidInput(t *testing.T) {
	now := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)
	createdAt := "2020-01-01T10:00:00Z"
	tweetedAt := "2025-01-15T14:30:00Z"

	input := &kafkamodels.ParsedTwitterPost{
		TwitterID:           "12345",
		Name:                "John Doe",
		Username:            "johndoe",
		ProfileImageURL:     "https://example.com/profile.jpg",
		FollowersCount:      1000,
		FollowingCount:      500,
		TweetCount:          250,
		ListedCount:         50,
		TweetID:             "tweet123",
		EditHistoryTweetIDs: []string{"tweet123", "tweet123_v1"},
		AuthorID:            "author123",
		AuthorUsername:      "author_user",
		IDCreatedAt:         createdAt,
		AuthorIDCreated:     createdAt,
		TweetedAt:           tweetedAt,
		Hashtags:            []string{"test", "golang"},
		Permalink:           "https://twitter.com/johndoe/status/tweet123",
		TweetType:           "original",
		URLs:                []string{"https://example.com"},
		MediaURL:            []string{"https://example.com/image.jpg"},
		UsernameMentioned:   []string{"user1", "user2"},
		UseridMentioned:     []string{"id1", "id2"},
		Lang:                "en",
		TweetText:           "Hello Twitter #test",
		ImpressionCount:     5000,
		RetweetCount:        100,
		ReplyCount:          50,
		LikeCount:           500,
		BookmarkCount:       50,
		QuoteCount:          10,
		TotalEngagement:     660,
	}

	result := ConvertTwitterPost(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.TwitterID != "12345" {
		t.Fatalf("expected TwitterID '12345', got %s", result.TwitterID)
	}
	if result.Name != "John Doe" {
		t.Fatalf("expected Name 'John Doe', got %s", result.Name)
	}
	if result.Username != "johndoe" {
		t.Fatalf("expected Username 'johndoe', got %s", result.Username)
	}
	if result.FollowersCount != 1000 {
		t.Fatalf("expected FollowersCount 1000, got %d", result.FollowersCount)
	}
	if result.LikeCount != 500 {
		t.Fatalf("expected LikeCount 500, got %d", result.LikeCount)
	}
	if result.TotalEngagement != 660 {
		t.Fatalf("expected TotalEngagement 660, got %d", result.TotalEngagement)
	}
	if len(result.Hashtags) != 2 {
		t.Fatalf("expected 2 hashtags, got %d", len(result.Hashtags))
	}
	if result.DayOfWeek != int64(now.Weekday()) {
		t.Fatalf("expected DayOfWeek %d, got %d", now.Weekday(), result.DayOfWeek)
	}
	if result.HourOfDay != int64(now.Hour()) {
		t.Fatalf("expected HourOfDay to be around %d", now.Hour())
	}
}

func TestConvertTwitterPost_EmptyStrings(t *testing.T) {
	input := &kafkamodels.ParsedTwitterPost{
		TwitterID:       "12345",
		TweetedAt:       "2025-01-15T14:30:00Z",
		IDCreatedAt:     "",
		AuthorIDCreated: "",
	}

	result := ConvertTwitterPost(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IDCreatedAt.IsZero() {
		t.Fatalf("expected zero IDCreatedAt, got %v", result.IDCreatedAt)
	}
	if !result.AuthorIDCreated.IsZero() {
		t.Fatalf("expected zero AuthorIDCreated, got %v", result.AuthorIDCreated)
	}
}

func TestConvertTwitterPost_DefaultTimeForEmptyTweetedAt(t *testing.T) {
	input := &kafkamodels.ParsedTwitterPost{
		TwitterID: "12345",
		TweetedAt: "", // Empty, should use current time (approximately)
	}

	result := ConvertTwitterPost(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// Should not be zero since it should use current time
	if result.TweetedAt.IsZero() {
		t.Fatal("expected non-zero TweetedAt for empty input")
	}
}

// Tests for ConvertTwitterInsights
func TestConvertTwitterInsights_NilInput(t *testing.T) {
	result := ConvertTwitterInsights(nil)
	if result != nil {
		t.Fatal("expected nil result for nil input")
	}
}

func TestConvertTwitterInsights_ValidInput(t *testing.T) {
	now := time.Now().UTC()
	insertedAt := now.Unix()
	createdDate := "2020-01-01T00:00:00Z"

	input := &kafkamodels.ParsedTwitterInsights{
		TwitterID:          "12345",
		RecordID:           "record123",
		Name:               "John Doe",
		Username:           "johndoe",
		ProfileImageURL:    "https://example.com/profile.jpg",
		Description:        "Test user",
		Verified:           true,
		AccountCreatedDate: createdDate,
		FollowersCount:     1000,
		FollowingCount:     500,
		TweetCount:         250,
		ListedCount:        50,
		LikeCount:          5000,
		InsertedAt:         insertedAt,
	}

	result := ConvertTwitterInsights(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.TwitterID != "12345" {
		t.Fatalf("expected TwitterID '12345', got %s", result.TwitterID)
	}
	if result.RecordID != "record123" {
		t.Fatalf("expected RecordID 'record123', got %s", result.RecordID)
	}
	if result.Verified != "true" {
		t.Fatalf("expected Verified 'true', got %s", result.Verified)
	}
	if result.FollowersCount != 1000 {
		t.Fatalf("expected FollowersCount 1000, got %d", result.FollowersCount)
	}
	if result.LikeCount != 5000 {
		t.Fatalf("expected LikeCount 5000, got %d", result.LikeCount)
	}
}

func TestConvertTwitterInsights_ZeroInsertedAt(t *testing.T) {
	input := &kafkamodels.ParsedTwitterInsights{
		TwitterID:  "12345",
		RecordID:   "record123",
		Verified:   false,
		InsertedAt: 0, // Should use current time
	}

	result := ConvertTwitterInsights(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Verified != "false" {
		t.Fatalf("expected Verified 'false', got %s", result.Verified)
	}
	// SavingTime should be approximately now, not zero
	if result.SavingTime.IsZero() {
		t.Fatal("expected non-zero SavingTime")
	}
}

// Tests for BulkInsertTwitterPosts
func TestBulkInsertTwitterPosts_EmptySlice(t *testing.T) {
	sink := newTestTwitterSink(&mockClickHouseClientTwitter{})
	err := sink.BulkInsertTwitterPosts(context.Background(), []*clickhousemodels.TwitterPosts{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func TestBulkInsertTwitterPosts_Success(t *testing.T) {
	called := false
	mock := &mockClickHouseClientTwitter{
		bulkInsertTwitterPostsFunc: func(ctx context.Context, posts []*clickhousemodels.TwitterPosts) error {
			called = true
			if len(posts) != 1 {
				t.Errorf("expected 1 post, got %d", len(posts))
			}
			return nil
		},
	}
	sink := newTestTwitterSink(mock)

	posts := []*clickhousemodels.TwitterPosts{
		{
			TwitterID: "12345",
			TweetID:   "tweet123",
			TweetText: "Hello Twitter",
		},
	}
	err := sink.BulkInsertTwitterPosts(context.Background(), posts)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertTwitterPosts to be called")
	}
}

func TestBulkInsertTwitterPosts_Error(t *testing.T) {
	expectedErr := errors.New("database error")
	mock := &mockClickHouseClientTwitter{
		bulkInsertTwitterPostsFunc: func(ctx context.Context, posts []*clickhousemodels.TwitterPosts) error {
			return expectedErr
		},
	}
	sink := newTestTwitterSink(mock)

	posts := []*clickhousemodels.TwitterPosts{
		{TwitterID: "12345", TweetID: "tweet123"},
	}
	err := sink.BulkInsertTwitterPosts(context.Background(), posts)

	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

// Tests for BulkInsertTwitterInsights
func TestBulkInsertTwitterInsights_EmptySlice(t *testing.T) {
	sink := newTestTwitterSink(&mockClickHouseClientTwitter{})
	err := sink.BulkInsertTwitterInsights(context.Background(), []*clickhousemodels.TwitterInsights{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func TestBulkInsertTwitterInsights_Success(t *testing.T) {
	called := false
	mock := &mockClickHouseClientTwitter{
		bulkInsertTwitterInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.TwitterInsights) error {
			called = true
			if len(insights) != 1 {
				t.Errorf("expected 1 insight, got %d", len(insights))
			}
			return nil
		},
	}
	sink := newTestTwitterSink(mock)

	insights := []*clickhousemodels.TwitterInsights{
		{
			TwitterID:      "12345",
			RecordID:       "record123",
			FollowersCount: 1000,
		},
	}
	err := sink.BulkInsertTwitterInsights(context.Background(), insights)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertTwitterInsights to be called")
	}
}

func TestBulkInsertTwitterInsights_Error(t *testing.T) {
	expectedErr := errors.New("database error")
	mock := &mockClickHouseClientTwitter{
		bulkInsertTwitterInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.TwitterInsights) error {
			return expectedErr
		},
	}
	sink := newTestTwitterSink(mock)

	insights := []*clickhousemodels.TwitterInsights{
		{TwitterID: "12345"},
	}
	err := sink.BulkInsertTwitterInsights(context.Background(), insights)

	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

// Tests for parseTwitterTimeOrDefault
func TestParseTwitterTimeOrDefault_ValidRFC3339(t *testing.T) {
	timeStr := "2025-01-15T14:30:00Z"
	fallback := time.Now()
	result := parseTwitterTimeOrDefault(timeStr, fallback)

	if result.IsZero() {
		t.Fatal("expected non-zero result for valid time")
	}
	if result.Year() != 2025 {
		t.Fatalf("expected year 2025, got %d", result.Year())
	}
}

func TestParseTwitterTimeOrDefault_EmptyString(t *testing.T) {
	fallback := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	result := parseTwitterTimeOrDefault("", fallback)

	if result != fallback {
		t.Fatalf("expected fallback time, got %v", result)
	}
}

func TestParseTwitterTimeOrDefault_InvalidFormat(t *testing.T) {
	fallback := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	result := parseTwitterTimeOrDefault("invalid-time", fallback)

	if result != fallback {
		t.Fatalf("expected fallback time for invalid format, got %v", result)
	}
}

// Tests for parseTwitterTimeOrZero
func TestParseTwitterTimeOrZero_ValidRFC3339(t *testing.T) {
	timeStr := "2025-01-15T14:30:00Z"
	result := parseTwitterTimeOrZero(timeStr)

	if result.IsZero() {
		t.Fatal("expected non-zero result for valid time")
	}
	if result.Year() != 2025 {
		t.Fatalf("expected year 2025, got %d", result.Year())
	}
}

func TestParseTwitterTimeOrZero_ValidRFC3339Nano(t *testing.T) {
	timeStr := "2025-01-15T14:30:00.123456789Z"
	result := parseTwitterTimeOrZero(timeStr)

	if result.IsZero() {
		t.Fatal("expected non-zero result for valid time with nanoseconds")
	}
	if result.Year() != 2025 {
		t.Fatalf("expected year 2025, got %d", result.Year())
	}
}

func TestParseTwitterTimeOrZero_EmptyString(t *testing.T) {
	result := parseTwitterTimeOrZero("")

	if !result.IsZero() {
		t.Fatalf("expected zero time for empty string, got %v", result)
	}
}

func TestParseTwitterTimeOrZero_InvalidFormat(t *testing.T) {
	result := parseTwitterTimeOrZero("not-a-date")

	if !result.IsZero() {
		t.Fatalf("expected zero time for invalid format, got %v", result)
	}
}

func TestParseTwitterTimeOrZero_CustomFormat(t *testing.T) {
	timeStr := "2025-01-15 14:30:05"
	result := parseTwitterTimeOrZero(timeStr)

	if result.IsZero() {
		t.Fatal("expected non-zero result for custom format")
	}
	if result.Year() != 2025 {
		t.Fatalf("expected year 2025, got %d", result.Year())
	}
}

func TestParseTwitterTimeOrZero_UTC(t *testing.T) {
	timeStr := "2025-01-15T14:30:00Z"
	result := parseTwitterTimeOrZero(timeStr)

	if result.Location() != time.UTC {
		t.Fatalf("expected UTC location, got %v", result.Location())
	}
}
