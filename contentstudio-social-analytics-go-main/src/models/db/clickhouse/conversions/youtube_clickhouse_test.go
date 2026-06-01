package conversions

import (
	"testing"
	"time"

	parsed "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

func TestConvertYouTubeChannel(t *testing.T) {
	now := time.Now().UTC()
	input := &parsed.ParsedYouTubeChannel{
		RecordID:        "record123",
		ChannelID:       "UC123456",
		Title:           "Test Channel",
		Description:     "Test Description",
		CustomURL:       "@testchannel",
		ThumbnailURL:    "http://example.com/thumb.jpg",
		BannerURL:       "http://example.com/banner.jpg",
		Country:         "US",
		SubscriberCount: 1000,
		VideoCount:      50,
		ViewCount:       100000,
		PublishedAt:     now.AddDate(-1, 0, 0),
		CreatedAt:       now,
		InsertedAt:      now,
	}

	result := ConvertYouTubeChannel(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.RecordID != "record123" {
		t.Errorf("expected RecordID 'record123', got '%s'", result.RecordID)
	}
	if result.ChannelID != "UC123456" {
		t.Errorf("expected ChannelID 'UC123456', got '%s'", result.ChannelID)
	}
	if result.Title != "Test Channel" {
		t.Errorf("expected Title 'Test Channel', got '%s'", result.Title)
	}
	if result.SubscriberCount != 1000 {
		t.Errorf("expected SubscriberCount 1000, got %d", result.SubscriberCount)
	}
	if result.VideoCount != 50 {
		t.Errorf("expected VideoCount 50, got %d", result.VideoCount)
	}
	if result.ViewCount != 100000 {
		t.Errorf("expected ViewCount 100000, got %d", result.ViewCount)
	}
}

func TestConvertYouTubeChannel_Nil(t *testing.T) {
	result := ConvertYouTubeChannel(nil)
	if result != nil {
		t.Error("expected nil result for nil input")
	}
}

func TestConvertYouTubeVideo(t *testing.T) {
	now := time.Now().UTC()
	input := &parsed.ParsedYouTubeVideo{
		VideoID:                     "video123",
		ChannelID:                   "UC123456",
		Title:                       "Test Video",
		Description:                 "Test Description",
		Duration:                    "PT5M30S",
		ThumbnailURL:                "http://example.com/thumb.jpg",
		IframeEmbedHTML:             "<iframe>...</iframe>",
		Likes:                       100,
		Dislikes:                    5,
		Views:                       10000,
		Comments:                    50,
		Shares:                      25,
		Favorites:                   10,
		Saved:                       15,
		SubscribersGained:           5,
		RedViews:                    200,
		MinutesWatched:              5000,
		RedMinutesWatched:           100,
		AvgViewDuration:             180,
		AvgViewPercentage:           75.5,
		Impressions:                 50000,
		ImpressionsClickThroughRate: 5.5,
		PublishedAt:                 now.AddDate(0, -1, 0),
		CreatedAt:                   now,
		InsertedAt:                  now,
		MediaType:                   "video",
	}

	result := ConvertYouTubeVideo(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.VideoID != "video123" {
		t.Errorf("expected VideoID 'video123', got '%s'", result.VideoID)
	}
	if result.Views != 10000 {
		t.Errorf("expected Views 10000, got %d", result.Views)
	}
	if result.Likes != 100 {
		t.Errorf("expected Likes 100, got %d", result.Likes)
	}
	if result.MediaType != "video" {
		t.Errorf("expected MediaType 'video', got '%s'", result.MediaType)
	}
	if result.AvgViewPercentage != 75.5 {
		t.Errorf("expected AvgViewPercentage 75.5, got %f", result.AvgViewPercentage)
	}
}

func TestConvertYouTubeVideo_Nil(t *testing.T) {
	result := ConvertYouTubeVideo(nil)
	if result != nil {
		t.Error("expected nil result for nil input")
	}
}

func TestConvertYouTubeActivityInsights(t *testing.T) {
	now := time.Now().UTC()
	input := &parsed.ParsedYouTubeActivityInsights{
		RecordID:                   "record123",
		ChannelID:                  "UC123456",
		RedViews:                   500,
		Views:                      10000,
		Likes:                      200,
		Dislikes:                   10,
		Comments:                   50,
		Shares:                     30,
		SubscribersGained:          25,
		EstimatedMinutesWatched:    50000,
		EstimatedRedMinutesWatched: 1000,
		AvgViewDuration:            300,
		AvgViewPercentage:          65.5,
		CreatedAt:                  now,
	}

	result := ConvertYouTubeActivityInsights(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.RecordID != "record123" {
		t.Errorf("expected RecordID 'record123', got '%s'", result.RecordID)
	}
	if result.Views != 10000 {
		t.Errorf("expected Views 10000, got %d", result.Views)
	}
	if result.SubscribersGained != 25 {
		t.Errorf("expected SubscribersGained 25, got %d", result.SubscribersGained)
	}
	if result.EstimatedMinutesWatched != 50000 {
		t.Errorf("expected EstimatedMinutesWatched 50000, got %d", result.EstimatedMinutesWatched)
	}
}

func TestConvertYouTubeActivityInsights_Nil(t *testing.T) {
	result := ConvertYouTubeActivityInsights(nil)
	if result != nil {
		t.Error("expected nil result for nil input")
	}
}

func TestConvertYouTubeTrafficInsights(t *testing.T) {
	now := time.Now().UTC()
	input := &parsed.ParsedYouTubeTrafficInsights{
		RecordID:               "record123",
		ChannelID:              "UC123456",
		PaidViews:              100,
		AnnotationViews:        50,
		EndScreenViews:         75,
		CampaignCardViews:      25,
		SubscriberViews:        500,
		NoLinkOtherViews:       200,
		YTChannelViews:         300,
		YTSearchViews:          1000,
		RelatedVideoViews:      800,
		YTOtherPageViews:       150,
		ExtURLViews:            400,
		PlaylistViews:          250,
		NotificationViews:      100,
		SubscriberWatchTime:    5000,
		NonSubscriberWatchTime: 10000,
		ShortsViews:            2000,
		CreatedAt:              now,
	}

	result := ConvertYouTubeTrafficInsights(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.RecordID != "record123" {
		t.Errorf("expected RecordID 'record123', got '%s'", result.RecordID)
	}
	if result.YTSearchViews != 1000 {
		t.Errorf("expected YTSearchViews 1000, got %d", result.YTSearchViews)
	}
	if result.ShortsViews != 2000 {
		t.Errorf("expected ShortsViews 2000, got %d", result.ShortsViews)
	}
	if result.SubscriberWatchTime != 5000 {
		t.Errorf("expected SubscriberWatchTime 5000, got %d", result.SubscriberWatchTime)
	}
}

func TestConvertYouTubeTrafficInsights_Nil(t *testing.T) {
	result := ConvertYouTubeTrafficInsights(nil)
	if result != nil {
		t.Error("expected nil result for nil input")
	}
}

func TestConvertYouTubeSharedInsights(t *testing.T) {
	now := time.Now().UTC()
	input := &parsed.ParsedYouTubeSharedInsights{
		RecordID:      "record123",
		ChannelID:     "UC123456",
		Ameba:         10,
		Blogger:       20,
		CopyPaste:     100,
		Cyworld:       5,
		Digg:          15,
		Dropbox:       25,
		Embed:         200,
		Mail:          50,
		WhatsApp:      300,
		Other:         75,
		FacebookMsgr:  150,
		FacebookPages: 100,
		Facebook:      500,
		Fotka:         10,
		VKontakte:     30,
		Discord:       80,
		GooglePlus:    20,
		Goo:           5,
		Hangouts:      10,
		LinkedIn:      100,
		Pinterest:     75,
		Myspace:       5,
		Reddit:        200,
		Skype:         30,
		Telegram:      150,
		Twitter:       400,
		Tumblr:        50,
		Viber:         25,
		Weibo:         40,
		WeChat:        60,
		YouTube:       100,
		InsertedAt:    now,
	}

	result := ConvertYouTubeSharedInsights(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.RecordID != "record123" {
		t.Errorf("expected RecordID 'record123', got '%s'", result.RecordID)
	}
	if result.WhatsApp != 300 {
		t.Errorf("expected WhatsApp 300, got %d", result.WhatsApp)
	}
	if result.Twitter != 400 {
		t.Errorf("expected Twitter 400, got %d", result.Twitter)
	}
	if result.Facebook != 500 {
		t.Errorf("expected Facebook 500, got %d", result.Facebook)
	}
}

func TestConvertYouTubeSharedInsights_Nil(t *testing.T) {
	result := ConvertYouTubeSharedInsights(nil)
	if result != nil {
		t.Error("expected nil result for nil input")
	}
}

func TestGenerateYouTubeRecordID(t *testing.T) {
	date := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	channelID := "UC123456"

	result := GenerateYouTubeRecordID(channelID, date)

	if result == "" {
		t.Error("expected non-empty record ID")
	}

	// Same inputs should produce same output
	result2 := GenerateYouTubeRecordID(channelID, date)
	if result != result2 {
		t.Error("expected same record ID for same inputs")
	}

	// Different channel should produce different output
	result3 := GenerateYouTubeRecordID("UC789", date)
	if result == result3 {
		t.Error("expected different record ID for different channel")
	}

	// Different date should produce different output
	result4 := GenerateYouTubeRecordID(channelID, date.AddDate(0, 0, 1))
	if result == result4 {
		t.Error("expected different record ID for different date")
	}
}

func TestConvertRawToChannel(t *testing.T) {
	now := time.Now().UTC()
	raw := &parsed.RawYouTubeChannel{
		ChannelID:       "UC123456",
		Title:           "Test Channel",
		Description:     "Test Description",
		CustomURL:       "@testchannel",
		ThumbnailURL:    "http://example.com/thumb.jpg",
		BannerURL:       "http://example.com/banner.jpg",
		Country:         "US",
		SubscriberCount: 1000,
		VideoCount:      50,
		ViewCount:       100000,
		PublishedAt:     now.AddDate(-1, 0, 0),
	}

	result := ConvertRawToChannel(raw)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ChannelID != "UC123456" {
		t.Errorf("expected ChannelID 'UC123456', got '%s'", result.ChannelID)
	}
	if result.RecordID == "" {
		t.Error("expected non-empty RecordID")
	}
	if result.InsertedAt.IsZero() {
		t.Error("expected non-zero InsertedAt")
	}
}

func TestConvertRawToChannel_Nil(t *testing.T) {
	result := ConvertRawToChannel(nil)
	if result != nil {
		t.Error("expected nil result for nil input")
	}
}

func TestConvertRawToVideo(t *testing.T) {
	now := time.Now().UTC()
	analyticsDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	raw := &parsed.RawYouTubeVideo{
		VideoID:       "video123",
		ChannelID:     "UC123456",
		Title:         "Test Video",
		Description:   "Test Description",
		Duration:      "PT5M30S",
		ThumbnailURL:  "http://example.com/thumb.jpg",
		MediaType:     "video",
		Views:         10000,
		Likes:         100,
		Dislikes:      5,
		Comments:      50,
		PublishedAt:   now.AddDate(0, -1, 0),
		AnalyticsDate: analyticsDate,
	}

	result := ConvertRawToVideo(raw)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.VideoID != "video123" {
		t.Errorf("expected VideoID 'video123', got '%s'", result.VideoID)
	}
	if result.Views != 10000 {
		t.Errorf("expected Views 10000, got %d", result.Views)
	}
	if result.CreatedAt != analyticsDate {
		t.Errorf("expected CreatedAt to be analytics date, got %v", result.CreatedAt)
	}
}

func TestConvertRawToVideo_EmptyAnalyticsDate(t *testing.T) {
	raw := &parsed.RawYouTubeVideo{
		VideoID:       "video123",
		ChannelID:     "UC123456",
		AnalyticsDate: time.Time{}, // Zero value
	}

	result := ConvertRawToVideo(raw)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// Should fall back to today's date
	if result.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestConvertRawToVideo_Nil(t *testing.T) {
	result := ConvertRawToVideo(nil)
	if result != nil {
		t.Error("expected nil result for nil input")
	}
}

func TestConvertRawToActivityInsights(t *testing.T) {
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	raw := &parsed.RawYouTubeActivityInsights{
		ChannelID:                  "UC123456",
		Date:                       date,
		Views:                      10000,
		RedViews:                   500,
		Likes:                      200,
		Dislikes:                   10,
		Comments:                   50,
		Shares:                     30,
		SubscribersGained:          25,
		EstimatedMinutesWatched:    50000,
		EstimatedRedMinutesWatched: 1000,
		AvgViewDuration:            300,
		AvgViewPercentage:          65.5,
	}

	result := ConvertRawToActivityInsights(raw)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ChannelID != "UC123456" {
		t.Errorf("expected ChannelID 'UC123456', got '%s'", result.ChannelID)
	}
	if result.RecordID == "" {
		t.Error("expected non-empty RecordID")
	}
	if result.Views != 10000 {
		t.Errorf("expected Views 10000, got %d", result.Views)
	}
	if result.CreatedAt != date {
		t.Errorf("expected CreatedAt to be %v, got %v", date, result.CreatedAt)
	}
}

func TestConvertRawToActivityInsights_Nil(t *testing.T) {
	result := ConvertRawToActivityInsights(nil)
	if result != nil {
		t.Error("expected nil result for nil input")
	}
}

func TestAggregateTrafficInsights(t *testing.T) {
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	channelID := "UC123456"

	rawInsights := []*parsed.RawYouTubeTrafficInsights{
		{TrafficSource: parsed.TrafficSourceYTSearch, Views: 1000, MinutesWatched: 5000},
		{TrafficSource: parsed.TrafficSourceExtURL, Views: 500, MinutesWatched: 2500},
		{TrafficSource: parsed.TrafficSourceSubscriber, Views: 300, MinutesWatched: 1500},
		{TrafficSource: parsed.TrafficSourceNoLinkOther, Views: 200, MinutesWatched: 1000},
		{TrafficSource: parsed.TrafficSourceShorts, Views: 2000, MinutesWatched: 3000},
		{TrafficSource: parsed.TrafficSourceRelatedVideo, Views: 800, MinutesWatched: 4000},
		{TrafficSource: parsed.TrafficSourceYTChannel, Views: 150, MinutesWatched: 750},
		{TrafficSource: parsed.TrafficSourcePlaylist, Views: 100, MinutesWatched: 500},
		{TrafficSource: parsed.TrafficSourceNotification, Views: 75, MinutesWatched: 375},
	}

	result := AggregateTrafficInsights(channelID, date, rawInsights)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ChannelID != channelID {
		t.Errorf("expected ChannelID '%s', got '%s'", channelID, result.ChannelID)
	}
	if result.YTSearchViews != 1000 {
		t.Errorf("expected YTSearchViews 1000, got %d", result.YTSearchViews)
	}
	if result.ExtURLViews != 500 {
		t.Errorf("expected ExtURLViews 500, got %d", result.ExtURLViews)
	}
	if result.SubscriberViews != 300 {
		t.Errorf("expected SubscriberViews 300, got %d", result.SubscriberViews)
	}
	if result.SubscriberWatchTime != 1500 {
		t.Errorf("expected SubscriberWatchTime 1500, got %d", result.SubscriberWatchTime)
	}
	// NonSubscriberWatchTime = sum of MinutesWatched from all non-subscriber sources:
	// YTSearch(5000) + ExtURL(2500) + NoLinkOther(1000) + Shorts(3000) + RelatedVideo(4000) + YTChannel(750) + Playlist(500) + Notification(375) = 17125
	if result.NonSubscriberWatchTime != 17125 {
		t.Errorf("expected NonSubscriberWatchTime 17125, got %d", result.NonSubscriberWatchTime)
	}
	if result.ShortsViews != 2000 {
		t.Errorf("expected ShortsViews 2000, got %d", result.ShortsViews)
	}
}

func TestAggregateTrafficInsights_Empty(t *testing.T) {
	result := AggregateTrafficInsights("UC123", time.Now(), []*parsed.RawYouTubeTrafficInsights{})
	if result != nil {
		t.Error("expected nil result for empty input")
	}
}

func TestAggregateTrafficInsights_AllSources(t *testing.T) {
	date := time.Now().UTC()
	channelID := "UC123456"

	rawInsights := []*parsed.RawYouTubeTrafficInsights{
		{TrafficSource: parsed.TrafficSourcePaid, Views: 100},
		{TrafficSource: parsed.TrafficSourceAnnotation, Views: 50},
		{TrafficSource: parsed.TrafficSourceEndScreen, Views: 75},
		{TrafficSource: parsed.TrafficSourceCampaignCard, Views: 25},
		{TrafficSource: parsed.TrafficSourceYTOtherPage, Views: 60},
	}

	result := AggregateTrafficInsights(channelID, date, rawInsights)

	if result.PaidViews != 100 {
		t.Errorf("expected PaidViews 100, got %d", result.PaidViews)
	}
	if result.AnnotationViews != 50 {
		t.Errorf("expected AnnotationViews 50, got %d", result.AnnotationViews)
	}
	if result.EndScreenViews != 75 {
		t.Errorf("expected EndScreenViews 75, got %d", result.EndScreenViews)
	}
	if result.CampaignCardViews != 25 {
		t.Errorf("expected CampaignCardViews 25, got %d", result.CampaignCardViews)
	}
	if result.YTOtherPageViews != 60 {
		t.Errorf("expected YTOtherPageViews 60, got %d", result.YTOtherPageViews)
	}
}

func TestAggregateSharedInsights(t *testing.T) {
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	channelID := "UC123456"

	rawInsights := []*parsed.RawYouTubeSharedInsights{
		{SharingService: parsed.SharingServiceWhatsApp, Shares: 300},
		{SharingService: parsed.SharingServiceTwitter, Shares: 400},
		{SharingService: parsed.SharingServiceFacebook, Shares: 500},
		{SharingService: parsed.SharingServiceTelegram, Shares: 150},
		{SharingService: parsed.SharingServiceReddit, Shares: 200},
		{SharingService: parsed.SharingServiceLinkedIn, Shares: 100},
		{SharingService: parsed.SharingServiceCopyPaste, Shares: 250},
		{SharingService: parsed.SharingServiceEmbed, Shares: 175},
	}

	result := AggregateSharedInsights(channelID, date, rawInsights)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ChannelID != channelID {
		t.Errorf("expected ChannelID '%s', got '%s'", channelID, result.ChannelID)
	}
	if result.WhatsApp != 300 {
		t.Errorf("expected WhatsApp 300, got %d", result.WhatsApp)
	}
	if result.Twitter != 400 {
		t.Errorf("expected Twitter 400, got %d", result.Twitter)
	}
	if result.Facebook != 500 {
		t.Errorf("expected Facebook 500, got %d", result.Facebook)
	}
	if result.Telegram != 150 {
		t.Errorf("expected Telegram 150, got %d", result.Telegram)
	}
	if result.Reddit != 200 {
		t.Errorf("expected Reddit 200, got %d", result.Reddit)
	}
}

func TestAggregateSharedInsights_Empty(t *testing.T) {
	result := AggregateSharedInsights("UC123", time.Now(), []*parsed.RawYouTubeSharedInsights{})
	if result != nil {
		t.Error("expected nil result for empty input")
	}
}

func TestAggregateSharedInsights_AllServices(t *testing.T) {
	date := time.Now().UTC()
	channelID := "UC123456"

	rawInsights := []*parsed.RawYouTubeSharedInsights{
		{SharingService: parsed.SharingServiceAmeba, Shares: 10},
		{SharingService: parsed.SharingServiceBlogger, Shares: 20},
		{SharingService: parsed.SharingServiceCyworld, Shares: 5},
		{SharingService: parsed.SharingServiceDigg, Shares: 15},
		{SharingService: parsed.SharingServiceDropbox, Shares: 25},
		{SharingService: parsed.SharingServiceMail, Shares: 50},
		{SharingService: parsed.SharingServiceOther, Shares: 75},
		{SharingService: parsed.SharingServiceFacebookMsgr, Shares: 100},
		{SharingService: parsed.SharingServiceFacebookPages, Shares: 80},
		{SharingService: parsed.SharingServiceFotka, Shares: 10},
		{SharingService: parsed.SharingServiceVKontakte, Shares: 30},
		{SharingService: parsed.SharingServiceDiscord, Shares: 60},
		{SharingService: parsed.SharingServiceGooglePlus, Shares: 5},
		{SharingService: parsed.SharingServiceGoo, Shares: 3},
		{SharingService: parsed.SharingServiceHangouts, Shares: 8},
		{SharingService: parsed.SharingServicePinterest, Shares: 40},
		{SharingService: parsed.SharingServiceMyspace, Shares: 2},
		{SharingService: parsed.SharingServiceSkype, Shares: 15},
		{SharingService: parsed.SharingServiceTumblr, Shares: 25},
		{SharingService: parsed.SharingServiceViber, Shares: 35},
		{SharingService: parsed.SharingServiceWeibo, Shares: 20},
		{SharingService: parsed.SharingServiceWeChat, Shares: 45},
		{SharingService: parsed.SharingServiceYouTube, Shares: 100},
		{SharingService: parsed.SharingServiceYouTubeGaming, Shares: 50},
		{SharingService: parsed.SharingServiceYouTubeKids, Shares: 30},
		{SharingService: parsed.SharingServiceYouTubeMusic, Shares: 40},
		{SharingService: parsed.SharingServiceYouTubeTV, Shares: 20},
	}

	result := AggregateSharedInsights(channelID, date, rawInsights)

	if result.Ameba != 10 {
		t.Errorf("expected Ameba 10, got %d", result.Ameba)
	}
	if result.Discord != 60 {
		t.Errorf("expected Discord 60, got %d", result.Discord)
	}
	// YouTube variants should be consolidated
	expectedYouTube := int64(100 + 50 + 30 + 40 + 20) // All YouTube variants
	if result.YouTube != expectedYouTube {
		t.Errorf("expected YouTube %d (consolidated), got %d", expectedYouTube, result.YouTube)
	}
}
