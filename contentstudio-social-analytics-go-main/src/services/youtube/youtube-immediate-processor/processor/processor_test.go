package processor

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

func createTestLogger() *logger.Logger {
	return logger.New("debug")
}

func createTestConfig() *config.Config {
	return &config.Config{
		YouTube: config.YouTubeConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
		},
		DecryptionKey: "test-decryption-key",
	}
}

func createTestProcessor(mongoRepo mongodb.UnifiedSocialRepository, sink *conversions.ClickHouseSink) *Processor {
	log := createTestLogger()
	cfg := createTestConfig()
	return &Processor{
		MongoRepo:    mongoRepo,
		YTClient:     social.NewYouTubeClient(cfg.YouTube.ClientID, cfg.YouTube.ClientSecret),
		Sink:         sink,
		Notifier:     nil,
		PusherClient: nil,
		Logger:       log,
		Cfg:          cfg,
	}
}

func createTestProcessorWithMockYTClient(mongoRepo mongodb.UnifiedSocialRepository, ytClient social.YouTubeAPI, sink *conversions.ClickHouseSink) *Processor {
	log := createTestLogger()
	cfg := createTestConfig()
	return NewWithClient(mongoRepo, ytClient, sink, nil, nil, log, cfg)
}

func TestNew(t *testing.T) {
	log := createTestLogger()
	cfg := createTestConfig()
	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	mockClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockClient)

	processor := New(mockRepo, sink, nil, nil, log, cfg)

	assert.NotNil(t, processor)
	assert.NotNil(t, processor.MongoRepo)
	assert.NotNil(t, processor.YTClient)
	assert.NotNil(t, processor.Sink)
	assert.NotNil(t, processor.Logger)
	assert.NotNil(t, processor.Cfg)
}

func TestProcessAccount_MissingChannelID(t *testing.T) {
	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	log := createTestLogger()
	mockClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockClient)
	processor := createTestProcessor(mockRepo, sink)

	wo := WorkOrder{
		ID:          primitive.NewObjectID().Hex(),
		ChannelID:   "",
		AccessToken: "test-token",
	}

	err := processor.ProcessAccount(context.Background(), wo)
	assert.NoError(t, err)
}

func TestProcessAccount_MissingAccessToken(t *testing.T) {
	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	log := createTestLogger()
	mockClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockClient)
	processor := createTestProcessor(mockRepo, sink)

	wo := WorkOrder{
		ID:          primitive.NewObjectID().Hex(),
		ChannelID:   "UC_test_channel",
		AccessToken: "",
	}

	err := processor.ProcessAccount(context.Background(), wo)
	assert.NoError(t, err)
}

func TestProcessAccount_InvalidAccountID(t *testing.T) {
	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	log := createTestLogger()
	mockClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockClient)
	processor := createTestProcessor(mockRepo, sink)

	wo := WorkOrder{
		ID:          "invalid-object-id",
		ChannelID:   "UC_test_channel",
		AccessToken: "test-token",
	}

	err := processor.ProcessAccount(context.Background(), wo)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid account ID")
}

func TestProcessAccount_AccountNotFound(t *testing.T) {
	accountID := primitive.NewObjectID()
	mockRepo := &mongodb.MockUnifiedSocialRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return nil, nil
		},
	}
	log := createTestLogger()
	mockClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockClient)
	processor := createTestProcessor(mockRepo, sink)

	wo := WorkOrder{
		ID:          accountID.Hex(),
		ChannelID:   "UC_test_channel",
		AccessToken: "test-token",
	}

	err := processor.ProcessAccount(context.Background(), wo)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "account not found")
}

func TestProcessAccount_MongoDBError(t *testing.T) {
	accountID := primitive.NewObjectID()
	mockRepo := &mongodb.MockUnifiedSocialRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return nil, errors.New("mongodb connection error")
		},
	}
	log := createTestLogger()
	mockClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockClient)
	processor := createTestProcessor(mockRepo, sink)

	wo := WorkOrder{
		ID:          accountID.Hex(),
		ChannelID:   "UC_test_channel",
		AccessToken: "test-token",
	}

	err := processor.ProcessAccount(context.Background(), wo)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch account from MongoDB")
}

func TestProcessAccount_ClearsStaleProcessingErrorBeforeRetry(t *testing.T) {
	accountID := primitive.NewObjectID()
	workspaceID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	clearCalls := 0

	mockRepo := &mongodb.MockUnifiedSocialRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:                 accountID,
				UserID:             userID,
				WorkspaceID:        workspaceID,
				PlatformName:       "youtube",
				PlatformIdentifier: "UC_test_channel",
				MetaData:           map[string]interface{}{"last_processing_error": "token expired"},
			}, nil
		},
		ClearProcessingErrorFunc: func(ctx context.Context, id primitive.ObjectID) error {
			clearCalls++
			if id != accountID {
				t.Fatalf("cleared account %s, want %s", id.Hex(), accountID.Hex())
			}
			return nil
		},
	}

	mockYTClient := &social.MockYouTubeClient{
		RefreshTokenFunc: func(ctx context.Context, refreshToken string) (*social.YouTubeTokenResponse, error) {
			return &social.YouTubeTokenResponse{AccessToken: "new_access_token"}, nil
		},
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			return nil, errors.New("request failed with status 401: unauthorized")
		},
		FetchVideosFunc: func(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]social.YouTubeActivityItem, error) {
			return nil, errors.New("request failed with status 401: unauthorized")
		},
		FetchActivityInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, errors.New("request failed with status 401: unauthorized")
		},
		FetchTrafficInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, errors.New("request failed with status 401: unauthorized")
		},
		FetchSharedInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, errors.New("request failed with status 401: unauthorized")
		},
	}

	log := createTestLogger()
	mockClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockClient)
	processor := createTestProcessorWithMockYTClient(mockRepo, mockYTClient, sink)

	err := processor.ProcessAccount(context.Background(), WorkOrder{
		ID:           accountID.Hex(),
		ChannelID:    "UC_test_channel",
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		WorkspaceID:  workspaceID.Hex(),
		SyncType:     kafkamodels.YouTubeSyncTypeFullSync,
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnauthorized))
	assert.Equal(t, 1, clearCalls)
}

func TestIsUnauthorizedError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "status 401 error",
			err:      errors.New("request failed with status 401"),
			expected: true,
		},
		{
			name:     "contains status 401",
			err:      errors.New("error: status 401 unauthorized"),
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("network timeout"),
			expected: false,
		},
		{
			name:     "status 500 error",
			err:      errors.New("request failed with status 500"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isUnauthorizedError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetInt64FromRow(t *testing.T) {
	tests := []struct {
		name     string
		row      []interface{}
		colIndex map[string]int
		colName  string
		expected int64
	}{
		{
			name:     "float64 value",
			row:      []interface{}{"2023-01-01", float64(12345)},
			colIndex: map[string]int{"day": 0, "views": 1},
			colName:  "views",
			expected: 12345,
		},
		{
			name:     "int64 value",
			row:      []interface{}{"2023-01-01", int64(54321)},
			colIndex: map[string]int{"day": 0, "views": 1},
			colName:  "views",
			expected: 54321,
		},
		{
			name:     "int value",
			row:      []interface{}{"2023-01-01", int(999)},
			colIndex: map[string]int{"day": 0, "views": 1},
			colName:  "views",
			expected: 999,
		},
		{
			name:     "column not found",
			row:      []interface{}{"2023-01-01", float64(12345)},
			colIndex: map[string]int{"day": 0, "views": 1},
			colName:  "likes",
			expected: 0,
		},
		{
			name:     "index out of range",
			row:      []interface{}{"2023-01-01"},
			colIndex: map[string]int{"day": 0, "views": 5},
			colName:  "views",
			expected: 0,
		},
		{
			name:     "unsupported type",
			row:      []interface{}{"2023-01-01", "not_a_number"},
			colIndex: map[string]int{"day": 0, "views": 1},
			colName:  "views",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getInt64FromRow(tt.row, tt.colIndex, tt.colName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFloat64FromRow(t *testing.T) {
	tests := []struct {
		name     string
		row      []interface{}
		colIndex map[string]int
		colName  string
		expected float64
	}{
		{
			name:     "float64 value",
			row:      []interface{}{"2023-01-01", float64(45.67)},
			colIndex: map[string]int{"day": 0, "percentage": 1},
			colName:  "percentage",
			expected: 45.67,
		},
		{
			name:     "int64 value",
			row:      []interface{}{"2023-01-01", int64(100)},
			colIndex: map[string]int{"day": 0, "percentage": 1},
			colName:  "percentage",
			expected: 100.0,
		},
		{
			name:     "int value",
			row:      []interface{}{"2023-01-01", int(50)},
			colIndex: map[string]int{"day": 0, "percentage": 1},
			colName:  "percentage",
			expected: 50.0,
		},
		{
			name:     "column not found",
			row:      []interface{}{"2023-01-01", float64(45.67)},
			colIndex: map[string]int{"day": 0, "percentage": 1},
			colName:  "ratio",
			expected: 0,
		},
		{
			name:     "unsupported type",
			row:      []interface{}{"2023-01-01", "not_a_float"},
			colIndex: map[string]int{"day": 0, "percentage": 1},
			colName:  "percentage",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFloat64FromRow(tt.row, tt.colIndex, tt.colName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseChannel(t *testing.T) {
	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	log := createTestLogger()
	mockClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockClient)
	processor := createTestProcessor(mockRepo, sink)

	channelID := "UC_test_channel"
	now := time.Now().UTC()

	// Create channel using JSON to populate inline structs
	channelJSON := `{
		"id": "UC_test_channel",
		"snippet": {
			"title": "Test Channel",
			"description": "A test channel description",
			"customUrl": "@testchannel",
			"publishedAt": "2020-01-15T10:30:00Z",
			"country": "US",
			"thumbnails": {
				"high": { "url": "https://example.com/thumbnail_high.jpg" }
			}
		},
		"statistics": {
			"subscriberCount": "10000",
			"videoCount": "500",
			"viewCount": "5000000"
		},
		"brandingSettings": {
			"image": { "bannerExternalUrl": "https://example.com/banner.jpg" }
		}
	}`
	var channel social.YouTubeChannelItem
	require.NoError(t, json.Unmarshal([]byte(channelJSON), &channel))

	parsed := processor.parseChannel(channelID, &channel, now)

	require.NotNil(t, parsed)
	assert.Equal(t, channelID, parsed.ChannelID)
	assert.Equal(t, "Test Channel", parsed.Title)
	assert.Equal(t, "A test channel description", parsed.Description)
	assert.Equal(t, "@testchannel", parsed.CustomURL)
	assert.Equal(t, "https://example.com/thumbnail_high.jpg", parsed.ThumbnailURL)
	assert.Equal(t, "https://example.com/banner.jpg", parsed.BannerURL)
	assert.Equal(t, "US", parsed.Country)
	assert.Equal(t, int64(10000), parsed.SubscriberCount)
	assert.Equal(t, int64(500), parsed.VideoCount)
	assert.Equal(t, int64(5000000), parsed.ViewCount)
}

func TestParseVideo(t *testing.T) {
	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	log := createTestLogger()
	mockClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockClient)
	processor := createTestProcessor(mockRepo, sink)

	channelID := "UC_test_channel"
	now := time.Now().UTC()

	videoJSON := `{
		"snippet": {
			"title": "Test Video Title",
			"description": "Test video description",
			"publishedAt": "2023-06-15T14:30:00Z",
			"thumbnails": { "high": { "url": "https://example.com/video_thumb_high.jpg" } }
		},
		"contentDetails": { "upload": { "videoId": "video123" } }
	}`
	var video social.YouTubeActivityItem
	require.NoError(t, json.Unmarshal([]byte(videoJSON), &video))

	detailsJSON := `{
		"id": "video123",
		"snippet": {
			"title": "Test Video Title (Details)",
			"description": "Test video description from details",
			"publishedAt": "2023-06-15T14:30:00Z",
			"thumbnails": { "high": { "url": "https://example.com/video_thumb_high_details.jpg" } }
		},
		"contentDetails": { "duration": "PT5M30S" },
		"statistics": {
			"viewCount": "1000000",
			"likeCount": "50000",
			"dislikeCount": "500",
			"commentCount": "10000",
			"favoriteCount": "5000"
		}
	}`
	var details social.YouTubeVideoItem
	require.NoError(t, json.Unmarshal([]byte(detailsJSON), &details))

	parsed := processor.parseVideo(channelID, &video, &details, kafkamodels.YouTubeMediaTypeVideo, now)

	require.NotNil(t, parsed)
	assert.Equal(t, "video123", parsed.VideoID)
	assert.Equal(t, channelID, parsed.ChannelID)
	assert.Equal(t, "Test Video Title (Details)", parsed.Title)
	assert.Equal(t, "Test video description from details", parsed.Description)
	assert.Equal(t, "PT5M30S", parsed.Duration)
	assert.Equal(t, kafkamodels.YouTubeMediaTypeVideo, parsed.MediaType)
	assert.Equal(t, int64(1000000), parsed.Views)
	assert.Equal(t, int64(50000), parsed.Likes)
	assert.Equal(t, int64(500), parsed.Dislikes)
	assert.Equal(t, int64(10000), parsed.Comments)
	assert.Equal(t, int64(5000), parsed.Favorites)
}

func TestParseVideo_WithoutDetails(t *testing.T) {
	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	log := createTestLogger()
	mockClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockClient)
	processor := createTestProcessor(mockRepo, sink)

	channelID := "UC_test_channel"
	now := time.Now().UTC()

	videoJSON := `{
		"snippet": {
			"title": "Test Video Title",
			"description": "Test video description",
			"publishedAt": "2023-06-15T14:30:00Z",
			"thumbnails": { "high": { "url": "https://example.com/video_thumb_high.jpg" } }
		},
		"contentDetails": { "upload": { "videoId": "video456" } }
	}`
	var video social.YouTubeActivityItem
	require.NoError(t, json.Unmarshal([]byte(videoJSON), &video))

	parsed := processor.parseVideo(channelID, &video, nil, kafkamodels.YouTubeMediaTypeShort, now)

	require.NotNil(t, parsed)
	assert.Equal(t, "video456", parsed.VideoID)
	assert.Equal(t, channelID, parsed.ChannelID)
	assert.Equal(t, "Test Video Title", parsed.Title)
	assert.Equal(t, "Test video description", parsed.Description)
	assert.Equal(t, kafkamodels.YouTubeMediaTypeShort, parsed.MediaType)
	assert.Equal(t, "https://example.com/video_thumb_high.jpg", parsed.ThumbnailURL)
}

func TestParseActivityInsights(t *testing.T) {
	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	log := createTestLogger()
	mockClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockClient)
	processor := createTestProcessor(mockRepo, sink)

	channelID := "UC_test_channel"

	resp := &social.YouTubeAnalyticsResponse{
		ColumnHeaders: []struct {
			Name       string `json:"name"`
			ColumnType string `json:"columnType"`
			DataType   string `json:"dataType"`
		}{
			{Name: "day"},
			{Name: "views"},
			{Name: "likes"},
			{Name: "dislikes"},
			{Name: "comments"},
			{Name: "shares"},
			{Name: "subscribersGained"},
			{Name: "estimatedMinutesWatched"},
			{Name: "averageViewDuration"},
			{Name: "averageViewPercentage"},
		},
		Rows: [][]interface{}{
			{"2023-06-01", float64(1000), float64(100), float64(10), float64(50), float64(25), float64(5), float64(5000), float64(300), float64(45.5)},
			{"2023-06-02", float64(1500), float64(150), float64(15), float64(75), float64(30), float64(8), float64(7500), float64(350), float64(50.2)},
		},
	}

	insights := processor.parseActivityInsights(channelID, resp)

	require.Len(t, insights, 2)

	assert.Equal(t, channelID, insights[0].ChannelID)
	assert.Equal(t, int64(1000), insights[0].Views)
	assert.Equal(t, int64(100), insights[0].Likes)
	assert.Equal(t, int64(10), insights[0].Dislikes)
	assert.Equal(t, int64(50), insights[0].Comments)
	assert.Equal(t, int64(25), insights[0].Shares)
	assert.Equal(t, int64(5), insights[0].SubscribersGained)
	assert.Equal(t, int64(5000), insights[0].EstimatedMinutesWatched)
	assert.Equal(t, int64(300), insights[0].AvgViewDuration)
	assert.Equal(t, float64(45.5), insights[0].AvgViewPercentage)

	assert.Equal(t, int64(1500), insights[1].Views)
}

func TestParseTrafficInsights(t *testing.T) {
	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	log := createTestLogger()
	mockClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockClient)
	processor := createTestProcessor(mockRepo, sink)

	channelID := "UC_test_channel"

	resp := &social.YouTubeAnalyticsResponse{
		ColumnHeaders: []struct {
			Name       string `json:"name"`
			ColumnType string `json:"columnType"`
			DataType   string `json:"dataType"`
		}{
			{Name: "day"},
			{Name: "insightTrafficSourceType"},
			{Name: "views"},
		},
		Rows: [][]interface{}{
			{"2023-06-01", "YT_SEARCH", float64(500)},
			{"2023-06-01", "RELATED_VIDEO", float64(300)},
			{"2023-06-01", "EXT_URL", float64(200)},
			{"2023-06-02", "YT_SEARCH", float64(600)},
		},
	}

	insights := processor.parseTrafficInsights(channelID, resp)

	require.Len(t, insights, 2)

	var day1 *kafkamodels.ParsedYouTubeTrafficInsights
	var day2 *kafkamodels.ParsedYouTubeTrafficInsights
	for i := range insights {
		if insights[i].CreatedAt.Format("2006-01-02") == "2023-06-01" {
			day1 = &insights[i]
		} else {
			day2 = &insights[i]
		}
	}

	require.NotNil(t, day1)
	assert.Equal(t, channelID, day1.ChannelID)
	assert.Equal(t, int64(500), day1.YTSearchViews)
	assert.Equal(t, int64(300), day1.RelatedVideoViews)
	assert.Equal(t, int64(200), day1.ExtURLViews)

	require.NotNil(t, day2)
	assert.Equal(t, int64(600), day2.YTSearchViews)
}

func TestParseSharedInsights(t *testing.T) {
	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	log := createTestLogger()
	mockClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockClient)
	processor := createTestProcessor(mockRepo, sink)

	channelID := "UC_test_channel"
	now := time.Now().UTC()

	resp := &social.YouTubeAnalyticsResponse{
		ColumnHeaders: []struct {
			Name       string `json:"name"`
			ColumnType string `json:"columnType"`
			DataType   string `json:"dataType"`
		}{
			{Name: "sharingService"},
			{Name: "shares"},
		},
		Rows: [][]interface{}{
			{"WHATS_APP", float64(100)},
			{"TWITTER", float64(200)},
			{"FACEBOOK", float64(150)},
			{"COPY_PASTE", float64(75)},
			{"TELEGRAM", float64(50)},
		},
	}

	insights := processor.parseSharedInsights(channelID, resp, now)

	require.NotNil(t, insights)
	assert.Equal(t, channelID, insights.ChannelID)
	assert.Equal(t, int64(100), insights.WhatsApp)
	assert.Equal(t, int64(200), insights.Twitter)
	assert.Equal(t, int64(150), insights.Facebook)
	assert.Equal(t, int64(75), insights.CopyPaste)
	assert.Equal(t, int64(50), insights.Telegram)
}

func TestParseAllData(t *testing.T) {
	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	log := createTestLogger()
	mockClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockClient)
	processor := createTestProcessor(mockRepo, sink)

	wo := WorkOrder{
		ID:          primitive.NewObjectID().Hex(),
		ChannelID:   "UC_test_channel",
		AccessToken: "test-token",
	}

	// Create channel using JSON
	channelJSON := `{
		"id": "UC_test_channel",
		"snippet": { "title": "Test Channel", "publishedAt": "2020-01-01T00:00:00Z" },
		"statistics": { "subscriberCount": "1000", "videoCount": "100", "viewCount": "100000" }
	}`
	var channel social.YouTubeChannelItem
	require.NoError(t, json.Unmarshal([]byte(channelJSON), &channel))

	// Create video activity using JSON
	videoActivityJSON := `{
		"snippet": { "title": "Video 1", "publishedAt": "2023-06-01T00:00:00Z" },
		"contentDetails": { "upload": { "videoId": "vid1" } }
	}`
	var videoActivity social.YouTubeActivityItem
	require.NoError(t, json.Unmarshal([]byte(videoActivityJSON), &videoActivity))

	// Create video details using JSON
	videoDetailsJSON := `{
		"id": "vid1",
		"snippet": { "title": "Video 1 Details", "publishedAt": "2023-06-01T00:00:00Z" },
		"contentDetails": { "duration": "PT3M" },
		"statistics": { "viewCount": "5000", "likeCount": "500" }
	}`
	var videoDetails social.YouTubeVideoItem
	require.NoError(t, json.Unmarshal([]byte(videoDetailsJSON), &videoDetails))

	fetchedData := &FetchedData{
		Channel: &channel,
		Videos:  []social.YouTubeActivityItem{videoActivity},
		VideoDetails: map[string]*social.YouTubeVideoItem{
			"vid1": &videoDetails,
		},
		ActivityInsights: &social.YouTubeAnalyticsResponse{
			ColumnHeaders: []struct {
				Name       string `json:"name"`
				ColumnType string `json:"columnType"`
				DataType   string `json:"dataType"`
			}{
				{Name: "day"},
				{Name: "views"},
			},
			Rows: [][]interface{}{
				{"2023-06-01", float64(1000)},
			},
		},
		TrafficInsights: &social.YouTubeAnalyticsResponse{
			ColumnHeaders: []struct {
				Name       string `json:"name"`
				ColumnType string `json:"columnType"`
				DataType   string `json:"dataType"`
			}{
				{Name: "day"},
				{Name: "insightTrafficSourceType"},
				{Name: "views"},
			},
			Rows: [][]interface{}{
				{"2023-06-01", "YT_SEARCH", float64(500)},
			},
		},
		SharedInsights: &social.YouTubeAnalyticsResponse{
			ColumnHeaders: []struct {
				Name       string `json:"name"`
				ColumnType string `json:"columnType"`
				DataType   string `json:"dataType"`
			}{
				{Name: "sharingService"},
				{Name: "shares"},
			},
			Rows: [][]interface{}{
				{"TWITTER", float64(100)},
			},
		},
	}

	parsed, err := processor.parseAllData(wo, fetchedData)

	require.NoError(t, err)
	require.NotNil(t, parsed)
	assert.NotNil(t, parsed.Channel)
	assert.Len(t, parsed.Videos, 1)
	assert.Len(t, parsed.ActivityInsights, 1)
	assert.Len(t, parsed.TrafficInsights, 1)
	assert.NotNil(t, parsed.SharedInsights)
}

func TestParseAllData_EmptyData(t *testing.T) {
	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	log := createTestLogger()
	mockClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockClient)
	processor := createTestProcessor(mockRepo, sink)

	wo := WorkOrder{
		ID:          primitive.NewObjectID().Hex(),
		ChannelID:   "UC_test_channel",
		AccessToken: "test-token",
	}

	fetchedData := &FetchedData{}

	parsed, err := processor.parseAllData(wo, fetchedData)

	require.NoError(t, err)
	require.NotNil(t, parsed)
	assert.Nil(t, parsed.Channel)
	assert.Len(t, parsed.Videos, 0)
	assert.Len(t, parsed.ActivityInsights, 0)
	assert.Len(t, parsed.TrafficInsights, 0)
	assert.Nil(t, parsed.SharedInsights)
}

func TestStoreInClickHouse(t *testing.T) {
	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	log := createTestLogger()

	channelsInserted := false
	videosInserted := false
	activityInserted := false
	trafficInserted := false
	sharedInserted := false

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertYouTubeChannelsFunc: func(ctx context.Context, channels []*clickhousemodels.YouTubeChannel) error {
			channelsInserted = true
			assert.Len(t, channels, 1)
			return nil
		},
		BulkInsertYouTubeVideosFunc: func(ctx context.Context, videos []*clickhousemodels.YouTubeVideo) error {
			videosInserted = true
			assert.Len(t, videos, 1)
			return nil
		},
		BulkInsertYouTubeActivityInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.YouTubeActivityInsights) error {
			activityInserted = true
			assert.Len(t, insights, 1)
			return nil
		},
		BulkInsertYouTubeTrafficInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.YouTubeTrafficInsights) error {
			trafficInserted = true
			assert.Len(t, insights, 1)
			return nil
		},
		BulkInsertYouTubeSharedInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.YouTubeSharedInsights) error {
			sharedInserted = true
			assert.Len(t, insights, 1)
			return nil
		},
	}

	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockClient)
	processor := createTestProcessor(mockRepo, sink)

	wo := WorkOrder{
		ID:        primitive.NewObjectID().Hex(),
		ChannelID: "UC_test_channel",
	}

	now := time.Now().UTC()
	parsedData := &ParsedData{
		Channel: &kafkamodels.ParsedYouTubeChannel{
			ChannelID: "UC_test_channel",
			Title:     "Test Channel",
		},
		Videos: []kafkamodels.ParsedYouTubeVideo{
			{
				VideoID:   "vid1",
				ChannelID: "UC_test_channel",
				Title:     "Test Video",
			},
		},
		ActivityInsights: []kafkamodels.ParsedYouTubeActivityInsights{
			{
				ChannelID: "UC_test_channel",
				Views:     1000,
				CreatedAt: now,
			},
		},
		TrafficInsights: []kafkamodels.ParsedYouTubeTrafficInsights{
			{
				ChannelID:     "UC_test_channel",
				YTSearchViews: 500,
				CreatedAt:     now,
			},
		},
		SharedInsights: &kafkamodels.ParsedYouTubeSharedInsights{
			ChannelID: "UC_test_channel",
			Twitter:   100,
		},
	}

	err := processor.storeInClickHouse(context.Background(), wo, parsedData)

	require.NoError(t, err)
	assert.True(t, channelsInserted)
	assert.True(t, videosInserted)
	assert.True(t, activityInserted)
	assert.True(t, trafficInserted)
	assert.True(t, sharedInserted)
}

func TestStoreInClickHouse_ChannelInsertError(t *testing.T) {
	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	log := createTestLogger()

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertYouTubeChannelsFunc: func(ctx context.Context, channels []*clickhousemodels.YouTubeChannel) error {
			return errors.New("channel insert error")
		},
	}

	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockClient)
	processor := createTestProcessor(mockRepo, sink)

	wo := WorkOrder{
		ID:        primitive.NewObjectID().Hex(),
		ChannelID: "UC_test_channel",
	}

	parsedData := &ParsedData{
		Channel: &kafkamodels.ParsedYouTubeChannel{
			ChannelID: "UC_test_channel",
			Title:     "Test Channel",
		},
	}

	err := processor.storeInClickHouse(context.Background(), wo, parsedData)
	assert.NoError(t, err) // Function logs error but doesn't return it
}

func TestStoreInClickHouse_EmptyData(t *testing.T) {
	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	log := createTestLogger()
	mockClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockClient)
	processor := createTestProcessor(mockRepo, sink)

	wo := WorkOrder{
		ID:        primitive.NewObjectID().Hex(),
		ChannelID: "UC_test_channel",
	}

	parsedData := &ParsedData{}

	err := processor.storeInClickHouse(context.Background(), wo, parsedData)
	assert.NoError(t, err)
}

func TestRefreshTokenIfNeeded_NoRefreshToken(t *testing.T) {
	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	log := createTestLogger()
	mockClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockClient)
	processor := createTestProcessor(mockRepo, sink)

	wo := WorkOrder{
		ID:           primitive.NewObjectID().Hex(),
		ChannelID:    "UC_test_channel",
		AccessToken:  "test-access-token",
		RefreshToken: "",
	}

	account := &mongomodels.SocialIntegration{
		ID:        primitive.NewObjectID(),
		ExtraData: nil,
	}

	token, err := processor.refreshTokenIfNeeded(context.Background(), wo, account)

	require.NoError(t, err)
	assert.Equal(t, "test-access-token", token)
}

func TestRefreshTokenIfNeeded_WithRefreshTokenFromExtraData(t *testing.T) {
	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	log := createTestLogger()
	mockClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockClient)
	processor := createTestProcessor(mockRepo, sink)

	wo := WorkOrder{
		ID:           primitive.NewObjectID().Hex(),
		ChannelID:    "UC_test_channel",
		AccessToken:  "test-access-token",
		RefreshToken: "",
	}

	account := &mongomodels.SocialIntegration{
		ID: primitive.NewObjectID(),
		ExtraData: map[string]interface{}{
			"refresh_token": "extra-data-refresh-token",
		},
	}

	// This will fail because we don't have a real YouTube API
	// but it tests that the refresh token is extracted from ExtraData
	_, err := processor.refreshTokenIfNeeded(context.Background(), wo, account)

	// We expect an error because the token refresh will fail with a fake token
	assert.Error(t, err)
}

func TestWorkOrderStruct(t *testing.T) {
	wo := WorkOrder{
		ID:           "test-id",
		ChannelID:    "UC_test_channel",
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		WorkspaceID:  "workspace-123",
		SyncType:     kafkamodels.YouTubeSyncTypeFullSync,
	}

	assert.Equal(t, "test-id", wo.ID)
	assert.Equal(t, "UC_test_channel", wo.ChannelID)
	assert.Equal(t, "access-token", wo.AccessToken)
	assert.Equal(t, "refresh-token", wo.RefreshToken)
	assert.Equal(t, "workspace-123", wo.WorkspaceID)
	assert.Equal(t, kafkamodels.YouTubeSyncTypeFullSync, wo.SyncType)
}

func TestFetchedDataStruct(t *testing.T) {
	channelJSON := `{"id": "UC_test"}`
	var channel social.YouTubeChannelItem
	require.NoError(t, json.Unmarshal([]byte(channelJSON), &channel))

	videoJSON := `{"snippet": {"title": "Video 1"}}`
	var video social.YouTubeActivityItem
	require.NoError(t, json.Unmarshal([]byte(videoJSON), &video))

	videoDetailsJSON := `{"id": "vid1"}`
	var videoDetails social.YouTubeVideoItem
	require.NoError(t, json.Unmarshal([]byte(videoDetailsJSON), &videoDetails))

	data := &FetchedData{
		Channel: &channel,
		Videos:  []social.YouTubeActivityItem{video},
		VideoDetails: map[string]*social.YouTubeVideoItem{
			"vid1": &videoDetails,
		},
		ActivityInsights: &social.YouTubeAnalyticsResponse{},
		TrafficInsights:  &social.YouTubeAnalyticsResponse{},
		SharedInsights:   &social.YouTubeAnalyticsResponse{},
	}

	assert.NotNil(t, data.Channel)
	assert.Len(t, data.Videos, 1)
	assert.Len(t, data.VideoDetails, 1)
	assert.NotNil(t, data.ActivityInsights)
	assert.NotNil(t, data.TrafficInsights)
	assert.NotNil(t, data.SharedInsights)
}

func TestParsedDataStruct(t *testing.T) {
	data := &ParsedData{
		Channel: &kafkamodels.ParsedYouTubeChannel{
			ChannelID: "UC_test",
		},
		Videos: []kafkamodels.ParsedYouTubeVideo{
			{VideoID: "vid1"},
		},
		ActivityInsights: []kafkamodels.ParsedYouTubeActivityInsights{
			{ChannelID: "UC_test"},
		},
		TrafficInsights: []kafkamodels.ParsedYouTubeTrafficInsights{
			{ChannelID: "UC_test"},
		},
		SharedInsights: &kafkamodels.ParsedYouTubeSharedInsights{
			ChannelID: "UC_test",
		},
	}

	assert.NotNil(t, data.Channel)
	assert.Len(t, data.Videos, 1)
	assert.Len(t, data.ActivityInsights, 1)
	assert.Len(t, data.TrafficInsights, 1)
	assert.NotNil(t, data.SharedInsights)
}

func TestConstants(t *testing.T) {
	assert.Equal(t, 90, DefaultInsightsDays)
	assert.Equal(t, 365, FullSyncInsightsDays)
	assert.Equal(t, 90, DefaultVideosDays)
	assert.Equal(t, 2020, FullSyncVideosStartYear)
}

func TestErrUnauthorized(t *testing.T) {
	assert.Equal(t, "unauthorized: invalid or expired token", ErrUnauthorized.Error())
}

func TestParseVideo_DefaultThumbnail(t *testing.T) {
	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	log := createTestLogger()
	mockClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockClient)
	processor := createTestProcessor(mockRepo, sink)

	channelID := "UC_test_channel"
	now := time.Now().UTC()

	videoJSON := `{
		"snippet": {
			"title": "Test Video",
			"publishedAt": "2023-06-15T14:30:00Z",
			"thumbnails": { "default": { "url": "https://example.com/default_thumb.jpg" } }
		},
		"contentDetails": { "upload": { "videoId": "video789" } }
	}`
	var video social.YouTubeActivityItem
	require.NoError(t, json.Unmarshal([]byte(videoJSON), &video))

	parsed := processor.parseVideo(channelID, &video, nil, kafkamodels.YouTubeMediaTypeVideo, now)

	require.NotNil(t, parsed)
	assert.Equal(t, "https://example.com/default_thumb.jpg", parsed.ThumbnailURL)
}

func TestParseVideo_DetailsDefaultThumbnail(t *testing.T) {
	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	log := createTestLogger()
	mockClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockClient)
	processor := createTestProcessor(mockRepo, sink)

	channelID := "UC_test_channel"
	now := time.Now().UTC()

	videoJSON := `{
		"snippet": { "title": "Test Video" },
		"contentDetails": { "upload": { "videoId": "video789" } }
	}`
	var video social.YouTubeActivityItem
	require.NoError(t, json.Unmarshal([]byte(videoJSON), &video))

	detailsJSON := `{
		"id": "video789",
		"snippet": {
			"title": "Test Video Details",
			"publishedAt": "2023-06-15T14:30:00Z",
			"thumbnails": { "default": { "url": "https://example.com/default_thumb_details.jpg" } }
		},
		"contentDetails": { "duration": "PT5M" },
		"statistics": { "viewCount": "1000" }
	}`
	var details social.YouTubeVideoItem
	require.NoError(t, json.Unmarshal([]byte(detailsJSON), &details))

	parsed := processor.parseVideo(channelID, &video, &details, kafkamodels.YouTubeMediaTypeVideo, now)

	require.NotNil(t, parsed)
	assert.Equal(t, "https://example.com/default_thumb_details.jpg", parsed.ThumbnailURL)
}

func TestParseActivityInsights_MissingDateColumn(t *testing.T) {
	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	log := createTestLogger()
	mockClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockClient)
	processor := createTestProcessor(mockRepo, sink)

	channelID := "UC_test_channel"

	// Row without proper date string
	resp := &social.YouTubeAnalyticsResponse{
		ColumnHeaders: []struct {
			Name       string `json:"name"`
			ColumnType string `json:"columnType"`
			DataType   string `json:"dataType"`
		}{
			{Name: "day"},
			{Name: "views"},
		},
		Rows: [][]interface{}{
			{float64(12345), float64(1000)}, // day is not a string
		},
	}

	insights := processor.parseActivityInsights(channelID, resp)

	// Should skip rows where date parsing fails
	assert.Len(t, insights, 0)
}

func TestParseTrafficInsights_AllTrafficSources(t *testing.T) {
	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	log := createTestLogger()
	mockClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockClient)
	processor := createTestProcessor(mockRepo, sink)

	channelID := "UC_test_channel"

	resp := &social.YouTubeAnalyticsResponse{
		ColumnHeaders: []struct {
			Name       string `json:"name"`
			ColumnType string `json:"columnType"`
			DataType   string `json:"dataType"`
		}{
			{Name: "day"},
			{Name: "insightTrafficSourceType"},
			{Name: "views"},
		},
		Rows: [][]interface{}{
			{"2023-06-01", "PAID", float64(100)},
			{"2023-06-01", "ANNOTATION", float64(200)},
			{"2023-06-01", "END_SCREEN", float64(300)},
			{"2023-06-01", "CAMPAIGN_CARD", float64(400)},
			{"2023-06-01", "SUBSCRIBER", float64(500)},
			{"2023-06-01", "NO_LINK_OTHER", float64(600)},
			{"2023-06-01", "YT_CHANNEL", float64(700)},
			{"2023-06-01", "YT_SEARCH", float64(800)},
			{"2023-06-01", "RELATED_VIDEO", float64(900)},
			{"2023-06-01", "YT_OTHER_PAGE", float64(1000)},
			{"2023-06-01", "EXT_URL", float64(1100)},
			{"2023-06-01", "PLAYLIST", float64(1200)},
			{"2023-06-01", "NOTIFICATION", float64(1300)},
			{"2023-06-01", "SHORTS", float64(1400)},
		},
	}

	insights := processor.parseTrafficInsights(channelID, resp)

	require.Len(t, insights, 1)
	day := insights[0]

	assert.Equal(t, int64(100), day.PaidViews)
	assert.Equal(t, int64(200), day.AnnotationViews)
	assert.Equal(t, int64(300), day.EndScreenViews)
	assert.Equal(t, int64(400), day.CampaignCardViews)
	assert.Equal(t, int64(500), day.SubscriberViews)
	assert.Equal(t, int64(600), day.NoLinkOtherViews)
	assert.Equal(t, int64(700), day.YTChannelViews)
	assert.Equal(t, int64(800), day.YTSearchViews)
	assert.Equal(t, int64(900), day.RelatedVideoViews)
	assert.Equal(t, int64(1000), day.YTOtherPageViews)
	assert.Equal(t, int64(1100), day.ExtURLViews)
	assert.Equal(t, int64(1200), day.PlaylistViews)
	assert.Equal(t, int64(1300), day.NotificationViews)
	assert.Equal(t, int64(1400), day.ShortsViews)
}

func TestParseSharedInsights_AllSharingServices(t *testing.T) {
	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	log := createTestLogger()
	mockClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockClient)
	processor := createTestProcessor(mockRepo, sink)

	channelID := "UC_test_channel"
	now := time.Now().UTC()

	resp := &social.YouTubeAnalyticsResponse{
		ColumnHeaders: []struct {
			Name       string `json:"name"`
			ColumnType string `json:"columnType"`
			DataType   string `json:"dataType"`
		}{
			{Name: "sharingService"},
			{Name: "shares"},
		},
		Rows: [][]interface{}{
			{"AMEBA", float64(10)},
			{"BLOGGER", float64(20)},
			{"COPY_PASTE", float64(30)},
			{"CYWORLD", float64(40)},
			{"DIGG", float64(50)},
			{"DROPBOX", float64(60)},
			{"EMBED", float64(70)},
			{"MAIL", float64(80)},
			{"WHATS_APP", float64(90)},
			{"OTHER", float64(100)},
			{"FACEBOOK_MESSENGER", float64(110)},
			{"FACEBOOK_PAGES", float64(120)},
			{"FACEBOOK", float64(130)},
			{"FOTKA", float64(140)},
			{"VKONTAKTE", float64(150)},
			{"DISCORD", float64(160)},
			{"GOOGLEPLUS", float64(170)},
			{"GOO", float64(180)},
			{"HANGOUTS", float64(190)},
			{"LINKEDIN", float64(200)},
			{"PINTEREST", float64(210)},
			{"MYSPACE", float64(220)},
			{"REDDIT", float64(230)},
			{"SKYPE", float64(240)},
			{"TELEGRAM", float64(250)},
			{"TWITTER", float64(260)},
			{"TUMBLR", float64(270)},
			{"VIBER", float64(280)},
			{"WEIBO", float64(290)},
			{"WECHAT", float64(300)},
			{"YOUTUBE", float64(310)},
			{"YOUTUBE_GAMING", float64(320)},
			{"YOUTUBE_KIDS", float64(330)},
			{"YOUTUBE_MUSIC", float64(340)},
			{"YOUTUBE_TV", float64(350)},
		},
	}

	insights := processor.parseSharedInsights(channelID, resp, now)

	require.NotNil(t, insights)
	assert.Equal(t, int64(10), insights.Ameba)
	assert.Equal(t, int64(20), insights.Blogger)
	assert.Equal(t, int64(30), insights.CopyPaste)
	assert.Equal(t, int64(40), insights.Cyworld)
	assert.Equal(t, int64(50), insights.Digg)
	assert.Equal(t, int64(60), insights.Dropbox)
	assert.Equal(t, int64(70), insights.Embed)
	assert.Equal(t, int64(80), insights.Mail)
	assert.Equal(t, int64(90), insights.WhatsApp)
	assert.Equal(t, int64(100), insights.Other)
	assert.Equal(t, int64(110), insights.FacebookMsgr)
	assert.Equal(t, int64(120), insights.FacebookPages)
	assert.Equal(t, int64(130), insights.Facebook)
	assert.Equal(t, int64(140), insights.Fotka)
	assert.Equal(t, int64(150), insights.VKontakte)
	assert.Equal(t, int64(160), insights.Discord)
	assert.Equal(t, int64(170), insights.GooglePlus)
	assert.Equal(t, int64(180), insights.Goo)
	assert.Equal(t, int64(190), insights.Hangouts)
	assert.Equal(t, int64(200), insights.LinkedIn)
	assert.Equal(t, int64(210), insights.Pinterest)
	assert.Equal(t, int64(220), insights.Myspace)
	assert.Equal(t, int64(230), insights.Reddit)
	assert.Equal(t, int64(240), insights.Skype)
	assert.Equal(t, int64(250), insights.Telegram)
	assert.Equal(t, int64(260), insights.Twitter)
	assert.Equal(t, int64(270), insights.Tumblr)
	assert.Equal(t, int64(280), insights.Viber)
	assert.Equal(t, int64(290), insights.Weibo)
	assert.Equal(t, int64(300), insights.WeChat)
	// YouTube shares are accumulated: 310 + 320 + 330 + 340 + 350 = 1650
	assert.Equal(t, int64(1650), insights.YouTube)
}

func TestProcessorStruct(t *testing.T) {
	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	log := createTestLogger()
	mockClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockClient)
	cfg := createTestConfig()
	ytClient := social.NewYouTubeClient(cfg.YouTube.ClientID, cfg.YouTube.ClientSecret)

	processor := &Processor{
		MongoRepo:    mockRepo,
		YTClient:     ytClient,
		Sink:         sink,
		Notifier:     nil,
		PusherClient: nil,
		Logger:       log,
		Cfg:          cfg,
	}

	assert.NotNil(t, processor.MongoRepo)
	assert.NotNil(t, processor.YTClient)
	assert.NotNil(t, processor.Sink)
	assert.Nil(t, processor.Notifier)
	assert.Nil(t, processor.PusherClient)
	assert.NotNil(t, processor.Logger)
	assert.NotNil(t, processor.Cfg)
}

func TestNewWithClient(t *testing.T) {
	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	mockYTClient := &social.MockYouTubeClient{}
	log := createTestLogger()
	mockCHClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockCHClient)
	cfg := createTestConfig()

	processor := NewWithClient(mockRepo, mockYTClient, sink, nil, nil, log, cfg)

	assert.NotNil(t, processor)
	assert.Equal(t, mockYTClient, processor.YTClient)
	assert.Equal(t, mockRepo, processor.MongoRepo)
}

func TestProcessAccount_FullFlowWithMock(t *testing.T) {
	accountID := primitive.NewObjectID()
	workspaceID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	mockRepo := &mongodb.MockUnifiedSocialRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:                 accountID,
				UserID:             userID,
				WorkspaceID:        workspaceID,
				PlatformName:       "youtube",
				PlatformIdentifier: "UC_test_channel",
				State:              mongomodels.StateProcessed,
			}, nil
		},
	}

	mockYTClient := &social.MockYouTubeClient{
		RefreshTokenFunc: func(ctx context.Context, refreshToken string) (*social.YouTubeTokenResponse, error) {
			return &social.YouTubeTokenResponse{AccessToken: "new_access_token"}, nil
		},
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			ch := social.YouTubeChannelItem{ID: "UC_test_channel"}
			ch.ContentDetails.RelatedPlaylists.Uploads = "UU_test_channel"
			return &social.YouTubeChannelResponse{
				Items: []social.YouTubeChannelItem{ch},
			}, nil
		},
		FetchVideosFunc: func(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]social.YouTubeActivityItem, error) {
			return []social.YouTubeActivityItem{}, nil
		},
		FetchVideoDetailsFunc: func(ctx context.Context, accessToken string, videoIDs []string) ([]social.YouTubeVideoItem, error) {
			return []social.YouTubeVideoItem{}, nil
		},
		FetchActivityInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return &social.YouTubeAnalyticsResponse{}, nil
		},
		FetchTrafficInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return &social.YouTubeAnalyticsResponse{}, nil
		},
		FetchSharedInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return &social.YouTubeAnalyticsResponse{}, nil
		},
		DetectMediaTypesFunc: func(ctx context.Context, videos []social.YouTubeVideoItem) map[string]string {
			return map[string]string{}
		},
	}

	log := createTestLogger()
	channelsInserted := false
	mockCHClient := &conversions.MockClickHouseClient{
		BulkInsertYouTubeChannelsFunc: func(ctx context.Context, channels []*clickhousemodels.YouTubeChannel) error {
			channelsInserted = true
			return nil
		},
	}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockCHClient)

	processor := createTestProcessorWithMockYTClient(mockRepo, mockYTClient, sink)

	wo := WorkOrder{
		ID:           accountID.Hex(),
		ChannelID:    "UC_test_channel",
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		WorkspaceID:  workspaceID.Hex(),
		SyncType:     kafkamodels.YouTubeSyncTypeIncremental,
	}

	err := processor.ProcessAccount(context.Background(), wo)

	require.NoError(t, err)
	assert.True(t, channelsInserted)
}

func TestProcessAccount_UnauthorizedError(t *testing.T) {
	accountID := primitive.NewObjectID()
	workspaceID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	mockRepo := &mongodb.MockUnifiedSocialRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:                 accountID,
				UserID:             userID,
				WorkspaceID:        workspaceID,
				PlatformName:       "youtube",
				PlatformIdentifier: "UC_test_channel",
				State:              mongomodels.StateProcessed,
			}, nil
		},
	}

	mockYTClient := &social.MockYouTubeClient{
		RefreshTokenFunc: func(ctx context.Context, refreshToken string) (*social.YouTubeTokenResponse, error) {
			return nil, errors.New("token refresh failed")
		},
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			return nil, errors.New("request failed with status 401: unauthorized")
		},
		FetchVideosFunc: func(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]social.YouTubeActivityItem, error) {
			return nil, errors.New("request failed with status 401: unauthorized")
		},
		FetchActivityInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, errors.New("request failed with status 401: unauthorized")
		},
		FetchTrafficInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, errors.New("request failed with status 401: unauthorized")
		},
		FetchSharedInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, errors.New("request failed with status 401: unauthorized")
		},
	}

	log := createTestLogger()
	mockCHClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockCHClient)

	processor := createTestProcessorWithMockYTClient(mockRepo, mockYTClient, sink)

	wo := WorkOrder{
		ID:           accountID.Hex(),
		ChannelID:    "UC_test_channel",
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		WorkspaceID:  workspaceID.Hex(),
	}

	err := processor.ProcessAccount(context.Background(), wo)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnauthorized))
}

func TestProcessAccount_FullSyncDateRanges(t *testing.T) {
	accountID := primitive.NewObjectID()
	workspaceID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	var capturedSinceDate time.Time

	mockRepo := &mongodb.MockUnifiedSocialRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:                 accountID,
				UserID:             userID,
				WorkspaceID:        workspaceID,
				PlatformName:       "youtube",
				PlatformIdentifier: "UC_test_channel",
				State:              mongomodels.StateProcessed,
			}, nil
		},
	}

	fsCh := social.YouTubeChannelItem{ID: "UC_test_channel"}
	fsCh.ContentDetails.RelatedPlaylists.Uploads = "UU_test_channel"

	mockYTClient := &social.MockYouTubeClient{
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			return &social.YouTubeChannelResponse{Items: []social.YouTubeChannelItem{fsCh}}, nil
		},
		FetchVideosFunc: func(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]social.YouTubeActivityItem, error) {
			capturedSinceDate = since
			return []social.YouTubeActivityItem{}, nil
		},
		FetchActivityInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return &social.YouTubeAnalyticsResponse{}, nil
		},
		FetchTrafficInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return &social.YouTubeAnalyticsResponse{}, nil
		},
		FetchSharedInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return &social.YouTubeAnalyticsResponse{}, nil
		},
		DetectMediaTypesFunc: func(ctx context.Context, videos []social.YouTubeVideoItem) map[string]string {
			return map[string]string{}
		},
	}

	log := createTestLogger()
	mockCHClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockCHClient)

	processor := createTestProcessorWithMockYTClient(mockRepo, mockYTClient, sink)

	wo := WorkOrder{
		ID:          accountID.Hex(),
		ChannelID:   "UC_test_channel",
		AccessToken: "test-access-token",
		SyncType:    kafkamodels.YouTubeSyncTypeFullSync,
	}

	err := processor.ProcessAccount(context.Background(), wo)

	require.NoError(t, err)
	// For full sync, videos since date should be from FullSyncVideosStartYear (2020)
	assert.Equal(t, FullSyncVideosStartYear, capturedSinceDate.Year())
	assert.Equal(t, time.January, capturedSinceDate.Month())
	assert.Equal(t, 1, capturedSinceDate.Day())
}

func TestProcessAccount_WithVideos(t *testing.T) {
	accountID := primitive.NewObjectID()
	workspaceID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	mockRepo := &mongodb.MockUnifiedSocialRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:                 accountID,
				UserID:             userID,
				WorkspaceID:        workspaceID,
				PlatformName:       "youtube",
				PlatformIdentifier: "UC_test_channel",
				State:              mongomodels.StateProcessed,
			}, nil
		},
	}

	// Create video activity items using JSON
	var videoActivity social.YouTubeActivityItem
	videoActivityJSON := `{
		"snippet": { "title": "Test Video", "publishedAt": "2023-06-01T00:00:00Z", "type": "upload" },
		"contentDetails": { "upload": { "videoId": "vid123" } }
	}`
	json.Unmarshal([]byte(videoActivityJSON), &videoActivity)

	var videoDetails social.YouTubeVideoItem
	videoDetailsJSON := `{
		"id": "vid123",
		"snippet": { "title": "Test Video Details", "publishedAt": "2023-06-01T00:00:00Z" },
		"contentDetails": { "duration": "PT5M30S" },
		"statistics": { "viewCount": "10000", "likeCount": "500" }
	}`
	json.Unmarshal([]byte(videoDetailsJSON), &videoDetails)

	chItem := social.YouTubeChannelItem{ID: "UC_test_channel"}
	chItem.ContentDetails.RelatedPlaylists.Uploads = "UU_test_channel"

	mockYTClient := &social.MockYouTubeClient{
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			return &social.YouTubeChannelResponse{
				Items: []social.YouTubeChannelItem{chItem},
			}, nil
		},
		FetchVideosFunc: func(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]social.YouTubeActivityItem, error) {
			return []social.YouTubeActivityItem{videoActivity}, nil
		},
		FetchVideoDetailsFunc: func(ctx context.Context, accessToken string, videoIDs []string) ([]social.YouTubeVideoItem, error) {
			assert.Equal(t, []string{"vid123"}, videoIDs)
			return []social.YouTubeVideoItem{videoDetails}, nil
		},
		FetchActivityInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return &social.YouTubeAnalyticsResponse{}, nil
		},
		FetchTrafficInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return &social.YouTubeAnalyticsResponse{}, nil
		},
		FetchSharedInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return &social.YouTubeAnalyticsResponse{}, nil
		},
		DetectMediaTypesFunc: func(ctx context.Context, videos []social.YouTubeVideoItem) map[string]string {
			return map[string]string{"vid123": "video"}
		},
	}

	log := createTestLogger()
	channelsInserted := false
	videosInserted := false
	var insertedVideos []*clickhousemodels.YouTubeVideo

	mockCHClient := &conversions.MockClickHouseClient{
		BulkInsertYouTubeChannelsFunc: func(ctx context.Context, channels []*clickhousemodels.YouTubeChannel) error {
			channelsInserted = true
			return nil
		},
		BulkInsertYouTubeVideosFunc: func(ctx context.Context, videos []*clickhousemodels.YouTubeVideo) error {
			videosInserted = true
			insertedVideos = videos
			return nil
		},
	}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockCHClient)

	processor := createTestProcessorWithMockYTClient(mockRepo, mockYTClient, sink)

	wo := WorkOrder{
		ID:          accountID.Hex(),
		ChannelID:   "UC_test_channel",
		AccessToken: "test-access-token",
		SyncType:    kafkamodels.YouTubeSyncTypeIncremental,
	}

	err := processor.ProcessAccount(context.Background(), wo)

	require.NoError(t, err)
	assert.True(t, channelsInserted)
	assert.True(t, videosInserted)
	assert.Len(t, insertedVideos, 1)
	assert.Equal(t, "vid123", insertedVideos[0].VideoID)
}

func TestFetchAllData_ParallelFetching(t *testing.T) {
	log := createTestLogger()

	fetchChannelsCalled := false
	fetchVideosCalled := false
	fetchActivityInsightsCalled := false
	fetchTrafficInsightsCalled := false
	fetchSharedInsightsCalled := false

	testCh := social.YouTubeChannelItem{ID: "UC_test"}
	testCh.ContentDetails.RelatedPlaylists.Uploads = "UU_test"

	mockYTClient := &social.MockYouTubeClient{
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			fetchChannelsCalled = true
			return &social.YouTubeChannelResponse{
				Items: []social.YouTubeChannelItem{testCh},
			}, nil
		},
		FetchVideosFunc: func(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]social.YouTubeActivityItem, error) {
			fetchVideosCalled = true
			return []social.YouTubeActivityItem{}, nil
		},
		FetchActivityInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			fetchActivityInsightsCalled = true
			return &social.YouTubeAnalyticsResponse{Rows: [][]interface{}{{"2023-06-01", float64(100)}}}, nil
		},
		FetchTrafficInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			fetchTrafficInsightsCalled = true
			return &social.YouTubeAnalyticsResponse{}, nil
		},
		FetchSharedInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			fetchSharedInsightsCalled = true
			return &social.YouTubeAnalyticsResponse{}, nil
		},
	}

	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	mockCHClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockCHClient)
	processor := createTestProcessorWithMockYTClient(mockRepo, mockYTClient, sink)

	wo := WorkOrder{
		ID:          primitive.NewObjectID().Hex(),
		ChannelID:   "UC_test",
		AccessToken: "test-token",
		SyncType:    kafkamodels.YouTubeSyncTypeIncremental,
	}

	data, err := processor.fetchAllData(context.Background(), wo)

	require.NoError(t, err)
	assert.True(t, fetchChannelsCalled)
	assert.True(t, fetchVideosCalled)
	assert.True(t, fetchActivityInsightsCalled)
	assert.True(t, fetchTrafficInsightsCalled)
	assert.True(t, fetchSharedInsightsCalled)
	assert.NotNil(t, data.Channel)
}

func TestFetchAllData_PartialFailure(t *testing.T) {
	log := createTestLogger()

	pfCh := social.YouTubeChannelItem{ID: "UC_test"}
	pfCh.ContentDetails.RelatedPlaylists.Uploads = "UU_test"

	mockYTClient := &social.MockYouTubeClient{
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			return &social.YouTubeChannelResponse{
				Items: []social.YouTubeChannelItem{pfCh},
			}, nil
		},
		FetchVideosFunc: func(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]social.YouTubeActivityItem, error) {
			return nil, errors.New("network error")
		},
		FetchActivityInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return &social.YouTubeAnalyticsResponse{}, nil
		},
		FetchTrafficInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return &social.YouTubeAnalyticsResponse{}, nil
		},
		FetchSharedInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return &social.YouTubeAnalyticsResponse{}, nil
		},
	}

	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	mockCHClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockCHClient)
	processor := createTestProcessorWithMockYTClient(mockRepo, mockYTClient, sink)

	wo := WorkOrder{
		ID:          primitive.NewObjectID().Hex(),
		ChannelID:   "UC_test",
		AccessToken: "test-token",
		SyncType:    kafkamodels.YouTubeSyncTypeIncremental,
	}

	data, err := processor.fetchAllData(context.Background(), wo)

	require.NoError(t, err)
	assert.NotNil(t, data.Channel)
	assert.Len(t, data.Videos, 0)
}

func TestFetchAllData_UsesRequestedDateRange(t *testing.T) {
	log := createTestLogger()

	startDate := "2025-01-10"
	endDate := "2025-01-20"
	parsedStart, err := time.Parse("2006-01-02", startDate)
	require.NoError(t, err)
	parsedEnd, err := time.Parse("2006-01-02", endDate)
	require.NoError(t, err)
	endExclusive := parsedEnd.AddDate(0, 0, 1)

	channel := social.YouTubeChannelItem{ID: "UC_test"}
	channel.ContentDetails.RelatedPlaylists.Uploads = "UU_test"

	videos := []social.YouTubeActivityItem{
		{
			ID: "playlist-new",
		},
		{
			ID: "playlist-keep",
		},
	}
	videos[0].ContentDetails.Upload.VideoID = "video-new"
	videos[0].Snippet.PublishedAt = "2025-01-21T00:00:00Z"
	videos[1].ContentDetails.Upload.VideoID = "video-keep"
	videos[1].Snippet.PublishedAt = "2025-01-15T12:00:00Z"

	var gotVideosSince time.Time
	var gotInsightsStart time.Time
	var gotInsightsEnd time.Time
	var gotVideoDetailsIDs []string

	mockYTClient := &social.MockYouTubeClient{
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			return &social.YouTubeChannelResponse{Items: []social.YouTubeChannelItem{channel}}, nil
		},
		FetchVideosFunc: func(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]social.YouTubeActivityItem, error) {
			gotVideosSince = since
			return videos, nil
		},
		FetchVideoDetailsFunc: func(ctx context.Context, accessToken string, videoIDs []string) ([]social.YouTubeVideoItem, error) {
			gotVideoDetailsIDs = append([]string{}, videoIDs...)
			return nil, nil
		},
		FetchActivityInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			gotInsightsStart = startDate
			gotInsightsEnd = endDate
			return &social.YouTubeAnalyticsResponse{}, nil
		},
		FetchTrafficInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return &social.YouTubeAnalyticsResponse{}, nil
		},
		FetchSharedInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return &social.YouTubeAnalyticsResponse{}, nil
		},
	}

	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	mockCHClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockCHClient)
	processor := createTestProcessorWithMockYTClient(mockRepo, mockYTClient, sink)

	wo := WorkOrder{
		ID:          primitive.NewObjectID().Hex(),
		ChannelID:   "UC_test",
		AccessToken: "test-token",
		SyncType:    kafkamodels.YouTubeSyncTypeIncremental,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	data, err := processor.fetchAllData(context.Background(), wo)
	require.NoError(t, err)
	require.NotNil(t, data.Channel)
	require.Len(t, data.Videos, 1)
	assert.Equal(t, "video-keep", data.Videos[0].ContentDetails.Upload.VideoID)
	assert.True(t, gotVideosSince.Equal(parsedStart))
	assert.True(t, gotInsightsStart.Equal(parsedStart))
	assert.True(t, gotInsightsEnd.Equal(parsedEnd))
	assert.ElementsMatch(t, []string{"video-keep"}, gotVideoDetailsIDs)
	assert.True(t, gotVideosSince.Before(endExclusive))
}

func TestRefreshTokenIfNeeded_WithMockYTClient(t *testing.T) {
	log := createTestLogger()

	mockYTClient := &social.MockYouTubeClient{
		RefreshTokenFunc: func(ctx context.Context, refreshToken string) (*social.YouTubeTokenResponse, error) {
			assert.Equal(t, "test-refresh-token", refreshToken)
			return &social.YouTubeTokenResponse{
				AccessToken: "new-access-token-from-refresh",
				ExpiresIn:   3600,
			}, nil
		},
	}

	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	mockCHClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockCHClient)
	processor := createTestProcessorWithMockYTClient(mockRepo, mockYTClient, sink)

	wo := WorkOrder{
		ID:           primitive.NewObjectID().Hex(),
		ChannelID:    "UC_test",
		AccessToken:  "old-access-token",
		RefreshToken: "test-refresh-token",
	}

	account := &mongomodels.SocialIntegration{
		ID: primitive.NewObjectID(),
	}

	token, err := processor.refreshTokenIfNeeded(context.Background(), wo, account)

	require.NoError(t, err)
	assert.Equal(t, "new-access-token-from-refresh", token)
}

func TestRefreshTokenIfNeeded_FromExtraDataRefreshToken(t *testing.T) {
	log := createTestLogger()

	mockYTClient := &social.MockYouTubeClient{
		RefreshTokenFunc: func(ctx context.Context, refreshToken string) (*social.YouTubeTokenResponse, error) {
			assert.Equal(t, "extra-data-refresh", refreshToken)
			return &social.YouTubeTokenResponse{
				AccessToken: "refreshed-from-extra-data",
			}, nil
		},
	}

	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	mockCHClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockCHClient)
	processor := createTestProcessorWithMockYTClient(mockRepo, mockYTClient, sink)

	wo := WorkOrder{
		ID:           primitive.NewObjectID().Hex(),
		ChannelID:    "UC_test",
		AccessToken:  "old-access-token",
		RefreshToken: "", // No refresh token in work order
	}

	account := &mongomodels.SocialIntegration{
		ID: primitive.NewObjectID(),
		ExtraData: map[string]interface{}{
			"refresh_token": "extra-data-refresh",
		},
	}

	token, err := processor.refreshTokenIfNeeded(context.Background(), wo, account)

	require.NoError(t, err)
	assert.Equal(t, "refreshed-from-extra-data", token)
}

func TestRefreshTokenIfNeeded_RefreshTokenCamelCase(t *testing.T) {
	log := createTestLogger()

	mockYTClient := &social.MockYouTubeClient{
		RefreshTokenFunc: func(ctx context.Context, refreshToken string) (*social.YouTubeTokenResponse, error) {
			assert.Equal(t, "camel-case-refresh", refreshToken)
			return &social.YouTubeTokenResponse{
				AccessToken: "refreshed-from-camel-case",
			}, nil
		},
	}

	mockRepo := &mongodb.MockUnifiedSocialRepository{}
	mockCHClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockCHClient)
	processor := createTestProcessorWithMockYTClient(mockRepo, mockYTClient, sink)

	wo := WorkOrder{
		ID:           primitive.NewObjectID().Hex(),
		ChannelID:    "UC_test",
		AccessToken:  "old-access-token",
		RefreshToken: "",
	}

	account := &mongomodels.SocialIntegration{
		ID: primitive.NewObjectID(),
		ExtraData: map[string]interface{}{
			"refreshToken": "camel-case-refresh", // Note: camelCase key
		},
	}

	token, err := processor.refreshTokenIfNeeded(context.Background(), wo, account)

	require.NoError(t, err)
	assert.Equal(t, "refreshed-from-camel-case", token)
}

// ==================== Logging Contract Tests ====================

func TestLoggingContract_YouTube_AuthError_WarnOnly(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()
	captureRecords, captureCleanup := logger.InstallCaptureSpy()
	defer captureCleanup()

	accountID := primitive.NewObjectID()
	mockRepo := &mongodb.MockUnifiedSocialRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:           accountID,
				PlatformName: "youtube",
				State:        mongomodels.StateProcessed,
			}, nil
		},
	}

	mockYTClient := &social.MockYouTubeClient{
		RefreshTokenFunc: func(ctx context.Context, refreshToken string) (*social.YouTubeTokenResponse, error) {
			return &social.YouTubeTokenResponse{AccessToken: "new_token"}, nil
		},
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			return nil, errors.New("request failed with status 401: unauthorized")
		},
	}

	cfg := createTestConfig()
	mockCHClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockCHClient)
	processor := NewWithClient(mockRepo, mockYTClient, sink, nil, nil, log, cfg)

	wo := WorkOrder{
		ID:           accountID.Hex(),
		ChannelID:    "UC_test_channel",
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
	}

	err := processor.ProcessAccount(context.Background(), wo)

	// Auth error should be returned (not swallowed)
	if err == nil {
		t.Fatal("expected error for auth failure")
	}
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}

	output := buf.String()

	// Should NOT have ERR level
	if strings.Contains(output, "ERR") {
		t.Error("unexpected ERR-level log; auth errors should not produce Error-level logs")
	}

	// CaptureException should NOT be called for auth errors
	for _, rec := range *captureRecords {
		if rec.Err != nil && strings.Contains(rec.Err.Error(), "status 401") {
			t.Error("CaptureException should NOT be called for auth errors")
		}
	}
}

func TestLoggingContract_YouTube_NonAuthSwallowed_UsesCaptureException(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()
	captureRecords, captureCleanup := logger.InstallCaptureSpy()
	defer captureCleanup()

	accountID := primitive.NewObjectID()
	mockRepo := &mongodb.MockUnifiedSocialRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:           accountID,
				PlatformName: "youtube",
				State:        mongomodels.StateProcessed,
			}, nil
		},
		UpdateStateFunc: func(ctx context.Context, id primitive.ObjectID, state string) error {
			return nil
		},
		UpdateAnalyticsTimestampFunc: func(ctx context.Context, id primitive.ObjectID, field string, timestamp time.Time) error {
			return nil
		},
	}

	refreshErr := errors.New("token refresh service unavailable")
	lcCh := social.YouTubeChannelItem{ID: "UC_test"}
	lcCh.ContentDetails.RelatedPlaylists.Uploads = "UU_test"

	mockYTClient := &social.MockYouTubeClient{
		RefreshTokenFunc: func(ctx context.Context, refreshToken string) (*social.YouTubeTokenResponse, error) {
			return nil, refreshErr
		},
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			return &social.YouTubeChannelResponse{
				Items: []social.YouTubeChannelItem{lcCh},
			}, nil
		},
		FetchVideosFunc: func(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]social.YouTubeActivityItem, error) {
			return []social.YouTubeActivityItem{}, nil
		},
		FetchActivityInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return &social.YouTubeAnalyticsResponse{}, nil
		},
		FetchTrafficInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return &social.YouTubeAnalyticsResponse{}, nil
		},
		FetchSharedInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return &social.YouTubeAnalyticsResponse{}, nil
		},
		DetectMediaTypesFunc: func(ctx context.Context, videos []social.YouTubeVideoItem) map[string]string {
			return map[string]string{}
		},
	}

	cfg := createTestConfig()
	mockCHClient := &conversions.MockClickHouseClient{
		BulkInsertYouTubeChannelsFunc: func(ctx context.Context, channels []*clickhousemodels.YouTubeChannel) error {
			return nil
		},
	}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockCHClient)
	processor := NewWithClient(mockRepo, mockYTClient, sink, nil, nil, log, cfg)

	wo := WorkOrder{
		ID:           accountID.Hex(),
		ChannelID:    "UC_test",
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		SyncType:     kafkamodels.YouTubeSyncTypeIncremental,
	}

	err := processor.ProcessAccount(context.Background(), wo)

	// Function should succeed (refresh error is swallowed, continues with original token)
	if err != nil {
		t.Fatalf("expected nil error (refresh error swallowed), got: %v", err)
	}

	// CaptureException SHOULD have been called for the swallowed refresh token error
	// (the error is wrapped by refreshTokenIfNeeded, so use errors.Is)
	found := false
	for _, rec := range *captureRecords {
		if errors.Is(rec.Err, refreshErr) {
			found = true
			break
		}
	}
	if !found {
		t.Error("CaptureException should be called for unexpected swallowed refresh token error")
	}

	output := buf.String()

	if !strings.Contains(output, "WRN") {
		t.Error("expected WRN-level log for token refresh failure")
	}
	if strings.Contains(output, "ERR") {
		t.Error("unexpected ERR-level log; processors should not log at Error level")
	}
}

func TestLoggingContract_YouTube_NoErrorLevelInProcessor(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()
	_, hookCleanup := logger.InstallHookSpy()
	defer hookCleanup()

	accountID := primitive.NewObjectID()
	mockRepo := &mongodb.MockUnifiedSocialRepository{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:           accountID,
				PlatformName: "youtube",
				State:        mongomodels.StateProcessed,
			}, nil
		},
	}

	// Multiple errors from different YouTube API endpoints
	mockYTClient := &social.MockYouTubeClient{
		RefreshTokenFunc: func(ctx context.Context, refreshToken string) (*social.YouTubeTokenResponse, error) {
			return nil, errors.New("refresh service down")
		},
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			return nil, errors.New("channels service unavailable")
		},
	}

	cfg := createTestConfig()
	mockCHClient := &conversions.MockClickHouseClient{}
	sink := conversions.NewClickHouseSinkWithClient(&log.Logger, mockCHClient)
	processor := NewWithClient(mockRepo, mockYTClient, sink, nil, nil, log, cfg)

	wo := WorkOrder{
		ID:           accountID.Hex(),
		ChannelID:    "UC_test",
		AccessToken:  "test-token",
		RefreshToken: "test-refresh",
	}

	_ = processor.ProcessAccount(context.Background(), wo)

	output := buf.String()
	errCount := strings.Count(output, "ERR")
	if errCount > 0 {
		t.Errorf("expected 0 ERR-level entries, got %d; processors should never log at Error level", errCount)
	}
}
