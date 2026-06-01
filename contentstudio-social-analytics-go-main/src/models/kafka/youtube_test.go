package kafka

import (
	"encoding/json"
	"testing"
	"time"
)

func TestYouTubeAccountWorkOrder_JSON(t *testing.T) {
	wo := YouTubeAccountWorkOrder{
		ID:           "123",
		ChannelID:    "UC123456",
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_123",
		WorkspaceID:  "workspace_456",
		SyncType:     YouTubeSyncTypeIncremental,
	}

	data, err := json.Marshal(wo)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded YouTubeAccountWorkOrder
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ID != wo.ID {
		t.Errorf("expected ID '%s', got '%s'", wo.ID, decoded.ID)
	}
	if decoded.ChannelID != wo.ChannelID {
		t.Errorf("expected ChannelID '%s', got '%s'", wo.ChannelID, decoded.ChannelID)
	}
	if decoded.SyncType != wo.SyncType {
		t.Errorf("expected SyncType '%s', got '%s'", wo.SyncType, decoded.SyncType)
	}
}

func TestYouTubeBatchWorkOrder_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	batch := YouTubeBatchWorkOrder{
		BatchID:   "batch_123",
		SyncType:  YouTubeSyncTypeFullSync,
		CreatedAt: now,
		Accounts: []YouTubeAccountWorkOrder{
			{ID: "1", ChannelID: "UC1", AccessToken: "token1", SyncType: YouTubeSyncTypeFullSync},
			{ID: "2", ChannelID: "UC2", AccessToken: "token2", SyncType: YouTubeSyncTypeFullSync},
		},
	}

	data, err := json.Marshal(batch)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded YouTubeBatchWorkOrder
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.BatchID != batch.BatchID {
		t.Errorf("expected BatchID '%s', got '%s'", batch.BatchID, decoded.BatchID)
	}
	if len(decoded.Accounts) != 2 {
		t.Errorf("expected 2 accounts, got %d", len(decoded.Accounts))
	}
	if decoded.Accounts[0].ChannelID != "UC1" {
		t.Errorf("expected first account ChannelID 'UC1', got '%s'", decoded.Accounts[0].ChannelID)
	}
}

func TestRawYouTubeChannel_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	raw := RawYouTubeChannel{
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
		WorkspaceID:     "workspace_123",
		SavingTime:      now,
	}

	data, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded RawYouTubeChannel
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ChannelID != raw.ChannelID {
		t.Errorf("expected ChannelID '%s', got '%s'", raw.ChannelID, decoded.ChannelID)
	}
	if decoded.SubscriberCount != raw.SubscriberCount {
		t.Errorf("expected SubscriberCount %d, got %d", raw.SubscriberCount, decoded.SubscriberCount)
	}
}

func TestRawYouTubeVideo_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	raw := RawYouTubeVideo{
		VideoID:                     "video123",
		ChannelID:                   "UC123456",
		Title:                       "Test Video",
		Description:                 "Test Description",
		ThumbnailURL:                "http://example.com/thumb.jpg",
		Duration:                    "PT5M30S",
		MediaType:                   YouTubeMediaTypeVideo,
		Views:                       10000,
		Likes:                       100,
		Dislikes:                    5,
		Comments:                    50,
		Shares:                      25,
		AvgViewPercentage:           75.5,
		ImpressionsClickThroughRate: 5.5,
		PublishedAt:                 now.AddDate(0, -1, 0),
		AnalyticsDate:               now,
		WorkspaceID:                 "workspace_123",
		SavingTime:                  now,
	}

	data, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded RawYouTubeVideo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.VideoID != raw.VideoID {
		t.Errorf("expected VideoID '%s', got '%s'", raw.VideoID, decoded.VideoID)
	}
	if decoded.MediaType != YouTubeMediaTypeVideo {
		t.Errorf("expected MediaType '%s', got '%s'", YouTubeMediaTypeVideo, decoded.MediaType)
	}
	if decoded.AvgViewPercentage != 75.5 {
		t.Errorf("expected AvgViewPercentage 75.5, got %f", decoded.AvgViewPercentage)
	}
}

func TestRawYouTubeActivityInsights_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	raw := RawYouTubeActivityInsights{
		ChannelID:                  "UC123456",
		Date:                       now,
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
		WorkspaceID:                "workspace_123",
		SavingTime:                 now,
	}

	data, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded RawYouTubeActivityInsights
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ChannelID != raw.ChannelID {
		t.Errorf("expected ChannelID '%s', got '%s'", raw.ChannelID, decoded.ChannelID)
	}
	if decoded.Views != raw.Views {
		t.Errorf("expected Views %d, got %d", raw.Views, decoded.Views)
	}
	if decoded.AvgViewPercentage != 65.5 {
		t.Errorf("expected AvgViewPercentage 65.5, got %f", decoded.AvgViewPercentage)
	}
}

func TestRawYouTubeTrafficInsights_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	raw := RawYouTubeTrafficInsights{
		ChannelID:      "UC123456",
		Date:           now,
		TrafficSource:  TrafficSourceYTSearch,
		Views:          1000,
		MinutesWatched: 5000,
		WorkspaceID:    "workspace_123",
		SavingTime:     now,
	}

	data, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded RawYouTubeTrafficInsights
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.TrafficSource != TrafficSourceYTSearch {
		t.Errorf("expected TrafficSource '%s', got '%s'", TrafficSourceYTSearch, decoded.TrafficSource)
	}
	if decoded.Views != 1000 {
		t.Errorf("expected Views 1000, got %d", decoded.Views)
	}
}

func TestRawYouTubeSharedInsights_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	raw := RawYouTubeSharedInsights{
		ChannelID:      "UC123456",
		SharingService: SharingServiceWhatsApp,
		Shares:         300,
		WorkspaceID:    "workspace_123",
		SavingTime:     now,
	}

	data, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded RawYouTubeSharedInsights
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.SharingService != SharingServiceWhatsApp {
		t.Errorf("expected SharingService '%s', got '%s'", SharingServiceWhatsApp, decoded.SharingService)
	}
	if decoded.Shares != 300 {
		t.Errorf("expected Shares 300, got %d", decoded.Shares)
	}
}

func TestParsedYouTubeChannel_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	parsed := ParsedYouTubeChannel{
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

	data, err := json.Marshal(parsed)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ParsedYouTubeChannel
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.RecordID != parsed.RecordID {
		t.Errorf("expected RecordID '%s', got '%s'", parsed.RecordID, decoded.RecordID)
	}
}

func TestParsedYouTubeVideo_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	parsed := ParsedYouTubeVideo{
		VideoID:                     "video123",
		ChannelID:                   "UC123456",
		Title:                       "Test Video",
		Description:                 "Test Description",
		Duration:                    "PT5M30S",
		ThumbnailURL:                "http://example.com/thumb.jpg",
		IframeEmbedHTML:             "<iframe>...</iframe>",
		MediaType:                   YouTubeMediaTypeShort,
		Views:                       10000,
		Likes:                       100,
		Dislikes:                    5,
		Comments:                    50,
		ImpressionsClickThroughRate: 5.5,
		PublishedAt:                 now.AddDate(0, -1, 0),
		CreatedAt:                   now,
		InsertedAt:                  now,
	}

	data, err := json.Marshal(parsed)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ParsedYouTubeVideo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.VideoID != parsed.VideoID {
		t.Errorf("expected VideoID '%s', got '%s'", parsed.VideoID, decoded.VideoID)
	}
	if decoded.MediaType != YouTubeMediaTypeShort {
		t.Errorf("expected MediaType '%s', got '%s'", YouTubeMediaTypeShort, decoded.MediaType)
	}
}

func TestParsedYouTubeActivityInsights_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	parsed := ParsedYouTubeActivityInsights{
		RecordID:                   "record123",
		ChannelID:                  "UC123456",
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
		CreatedAt:                  now,
	}

	data, err := json.Marshal(parsed)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ParsedYouTubeActivityInsights
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.RecordID != parsed.RecordID {
		t.Errorf("expected RecordID '%s', got '%s'", parsed.RecordID, decoded.RecordID)
	}
}

func TestParsedYouTubeTrafficInsights_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	parsed := ParsedYouTubeTrafficInsights{
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
		ShortsViews:            2000,
		SubscriberWatchTime:    5000,
		NonSubscriberWatchTime: 10000,
		CreatedAt:              now,
	}

	data, err := json.Marshal(parsed)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ParsedYouTubeTrafficInsights
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.YTSearchViews != 1000 {
		t.Errorf("expected YTSearchViews 1000, got %d", decoded.YTSearchViews)
	}
	if decoded.ShortsViews != 2000 {
		t.Errorf("expected ShortsViews 2000, got %d", decoded.ShortsViews)
	}
}

func TestParsedYouTubeSharedInsights_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	parsed := ParsedYouTubeSharedInsights{
		RecordID:      "record123",
		ChannelID:     "UC123456",
		WhatsApp:      300,
		Twitter:       400,
		Facebook:      500,
		Telegram:      150,
		Reddit:        200,
		LinkedIn:      100,
		Discord:       80,
		YouTube:       250,
		InsertedAt:    now,
	}

	data, err := json.Marshal(parsed)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ParsedYouTubeSharedInsights
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.WhatsApp != 300 {
		t.Errorf("expected WhatsApp 300, got %d", decoded.WhatsApp)
	}
	if decoded.Discord != 80 {
		t.Errorf("expected Discord 80, got %d", decoded.Discord)
	}
}

func TestYouTubeConstants(t *testing.T) {
	// Test traffic source constants
	if TrafficSourceYTSearch != "YT_SEARCH" {
		t.Errorf("expected TrafficSourceYTSearch 'YT_SEARCH', got '%s'", TrafficSourceYTSearch)
	}
	if TrafficSourceShorts != "SHORTS" {
		t.Errorf("expected TrafficSourceShorts 'SHORTS', got '%s'", TrafficSourceShorts)
	}

	// Test sharing service constants
	if SharingServiceWhatsApp != "WHATS_APP" {
		t.Errorf("expected SharingServiceWhatsApp 'WHATS_APP', got '%s'", SharingServiceWhatsApp)
	}
	if SharingServiceTwitter != "TWITTER" {
		t.Errorf("expected SharingServiceTwitter 'TWITTER', got '%s'", SharingServiceTwitter)
	}
	if SharingServiceDiscord != "DISCORD" {
		t.Errorf("expected SharingServiceDiscord 'DISCORD', got '%s'", SharingServiceDiscord)
	}

	// Test media type constants
	if YouTubeMediaTypeVideo != "video" {
		t.Errorf("expected YouTubeMediaTypeVideo 'video', got '%s'", YouTubeMediaTypeVideo)
	}
	if YouTubeMediaTypeShort != "short" {
		t.Errorf("expected YouTubeMediaTypeShort 'short', got '%s'", YouTubeMediaTypeShort)
	}

	// Test sync type constants
	if YouTubeSyncTypeIncremental != "incremental" {
		t.Errorf("expected YouTubeSyncTypeIncremental 'incremental', got '%s'", YouTubeSyncTypeIncremental)
	}
	if YouTubeSyncTypeImmediate != "immediate" {
		t.Errorf("expected YouTubeSyncTypeImmediate 'immediate', got '%s'", YouTubeSyncTypeImmediate)
	}
	if YouTubeSyncTypeFullSync != "full_sync" {
		t.Errorf("expected YouTubeSyncTypeFullSync 'full_sync', got '%s'", YouTubeSyncTypeFullSync)
	}
}

func TestYouTubeKafkaTopics(t *testing.T) {
	// Verify all topics are defined
	if YouTubeKafkaTopics.WorkOrder != "work-order-youtube" {
		t.Errorf("expected WorkOrder topic 'work-order-youtube', got '%s'", YouTubeKafkaTopics.WorkOrder)
	}
	if YouTubeKafkaTopics.ImmediateWorkOrder != "immediate-work-order-youtube" {
		t.Errorf("expected ImmediateWorkOrder topic 'immediate-work-order-youtube', got '%s'", YouTubeKafkaTopics.ImmediateWorkOrder)
	}
	if YouTubeKafkaTopics.RawChannels != "raw-youtube-channels" {
		t.Errorf("expected RawChannels topic 'raw-youtube-channels', got '%s'", YouTubeKafkaTopics.RawChannels)
	}
	if YouTubeKafkaTopics.RawVideos != "raw-youtube-videos" {
		t.Errorf("expected RawVideos topic 'raw-youtube-videos', got '%s'", YouTubeKafkaTopics.RawVideos)
	}
	if YouTubeKafkaTopics.ParsedVideos != "parsed-youtube-videos" {
		t.Errorf("expected ParsedVideos topic 'parsed-youtube-videos', got '%s'", YouTubeKafkaTopics.ParsedVideos)
	}
}

func TestYouTubeTrafficSourceConstants(t *testing.T) {
	sources := []string{
		TrafficSourcePaid,
		TrafficSourceAnnotation,
		TrafficSourceEndScreen,
		TrafficSourceCampaignCard,
		TrafficSourceSubscriber,
		TrafficSourceNoLinkOther,
		TrafficSourceYTChannel,
		TrafficSourceYTSearch,
		TrafficSourceRelatedVideo,
		TrafficSourceYTOtherPage,
		TrafficSourceExtURL,
		TrafficSourcePlaylist,
		TrafficSourceNotification,
		TrafficSourceShorts,
	}

	for _, source := range sources {
		if source == "" {
			t.Error("traffic source constant should not be empty")
		}
	}

	if len(sources) != 14 {
		t.Errorf("expected 14 traffic sources, got %d", len(sources))
	}
}

func TestYouTubeSharingServiceConstants(t *testing.T) {
	services := []string{
		SharingServiceAmeba,
		SharingServiceBlogger,
		SharingServiceCopyPaste,
		SharingServiceCyworld,
		SharingServiceDigg,
		SharingServiceDropbox,
		SharingServiceEmbed,
		SharingServiceMail,
		SharingServiceWhatsApp,
		SharingServiceOther,
		SharingServiceFacebookMsgr,
		SharingServiceFacebookPages,
		SharingServiceFacebook,
		SharingServiceFotka,
		SharingServiceVKontakte,
		SharingServiceDiscord,
		SharingServiceGooglePlus,
		SharingServiceGoo,
		SharingServiceHangouts,
		SharingServiceLinkedIn,
		SharingServicePinterest,
		SharingServiceMyspace,
		SharingServiceReddit,
		SharingServiceSkype,
		SharingServiceTelegram,
		SharingServiceTwitter,
		SharingServiceTumblr,
		SharingServiceViber,
		SharingServiceWeibo,
		SharingServiceWeChat,
		SharingServiceYouTube,
		SharingServiceYouTubeGaming,
		SharingServiceYouTubeKids,
		SharingServiceYouTubeMusic,
		SharingServiceYouTubeTV,
	}

	for _, service := range services {
		if service == "" {
			t.Error("sharing service constant should not be empty")
		}
	}

	if len(services) != 35 {
		t.Errorf("expected 35 sharing services, got %d", len(services))
	}
}
