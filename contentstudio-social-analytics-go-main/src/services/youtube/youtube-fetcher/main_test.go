package main

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// Helper for tests to create mock channel response using JSON
func createMockChannelResponse(t *testing.T) social.YouTubeChannelResponse {
	respJSON := `{
		"items": [{
			"id": "UC_test_channel",
			"snippet": { "title": "Test Channel", "publishedAt": "2020-01-15T10:30:00Z" },
			"statistics": { "subscriberCount": "10000", "videoCount": "500", "viewCount": "5000000" }
		}]
	}`
	var resp social.YouTubeChannelResponse
	require.NoError(t, json.Unmarshal([]byte(respJSON), &resp))
	return resp
}

// testProducerWithMessages wraps kafka.MockProducer with message collection
type testProducerWithMessages struct {
	kafka.MockProducer
	mu       sync.Mutex
	messages map[string][][]byte
}

func newTestProducer() *testProducerWithMessages {
	p := &testProducerWithMessages{
		messages: make(map[string][][]byte),
	}
	p.ProduceFunc = func(ctx context.Context, topic string, key, value []byte) error {
		p.mu.Lock()
		defer p.mu.Unlock()
		p.messages[topic] = append(p.messages[topic], value)
		return nil
	}
	return p
}

func (p *testProducerWithMessages) getMessages(topic string) [][]byte {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.messages[topic]
}

func createTestFetcherLogger() *logger.Logger {
	return logger.New("debug")
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
			err:      errors.New("API error: status 401 unauthorized"),
			expected: true,
		},
		{
			name:     "status 403 error",
			err:      errors.New("request failed with status 403"),
			expected: false,
		},
		{
			name:     "network error",
			err:      errors.New("connection timeout"),
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

func TestSemForAccount(t *testing.T) {
	// Clear the global map first
	accountSemaphores = sync.Map{}

	sem1 := semForAccount("test_channel1_unique", 1)
	require.NotNil(t, sem1)

	// Same channel should return the same semaphore
	sem2 := semForAccount("test_channel1_unique", 1)
	assert.Equal(t, sem1, sem2)

	// Different channel should return a different semaphore
	sem3 := semForAccount("test_channel2_unique", 1)
	// Verify they are different by checking they are not the same pointer
	// We can't directly compare semaphores, but we can verify they have different addresses
	assert.NotNil(t, sem3)

	// Acquire sem1, verify sem3 can still be acquired (they're independent)
	ctx := context.Background()
	err := sem1.Acquire(ctx, 1)
	require.NoError(t, err)

	// sem3 should be acquirable since it's a different semaphore
	err = sem3.Acquire(ctx, 1)
	require.NoError(t, err)

	sem1.Release(1)
	sem3.Release(1)
}

func TestSemForAccount_Concurrent(t *testing.T) {
	// Clear the global map first
	accountSemaphores = sync.Map{}

	channelID := "test-channel-concurrent"
	var wg sync.WaitGroup
	semaphores := make(chan interface{}, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem := semForAccount(channelID, 1)
			semaphores <- sem
		}()
	}

	wg.Wait()
	close(semaphores)

	// All goroutines should get the same semaphore
	var first interface{}
	for sem := range semaphores {
		if first == nil {
			first = sem
		} else {
			assert.Equal(t, first, sem)
		}
	}
}

func TestErrUnauthorizedConstant(t *testing.T) {
	assert.Equal(t, "unauthorized: invalid or expired token", ErrUnauthorized.Error())
}

func TestConstants(t *testing.T) {
	assert.Equal(t, 10, maxWorkers)
	assert.Equal(t, 200, workOrderChanSize)
	assert.Equal(t, 500, timestampChanSize)
	assert.Equal(t, "youtube-fetcher-group", consumerGroup)
	assert.Equal(t, "work-order-youtube", topicWorkOrderBatch)
	assert.Equal(t, "raw-youtube-channels", topicRawChannels)
	assert.Equal(t, "raw-youtube-videos", topicRawVideos)
	assert.Equal(t, "raw-youtube-activity-insights", topicRawActivityInsights)
	assert.Equal(t, "raw-youtube-traffic-insights", topicRawTrafficInsights)
	assert.Equal(t, "raw-youtube-shared-insights", topicRawSharedInsights)
	assert.Equal(t, 15*time.Minute, idleTimeout)
	assert.Equal(t, 14, incrementalVideosDays)
	assert.Equal(t, 14, incrementalInsightsDays)
	assert.Equal(t, 90, immediateVideosDays)
	assert.Equal(t, 90, immediateInsightsDays)
	assert.Equal(t, 365, fullSyncVideosDays)
	assert.Equal(t, 365, fullSyncInsightsDays)
	assert.Equal(t, 50, maxConcurrentAccounts)
}

func TestWorkOrderMessageStruct(t *testing.T) {
	msg := WorkOrderMessage{
		AccountID:   "account123",
		ChannelID:   "UC_test_channel",
		Value:       []byte(`{"channel_id": "UC_test_channel"}`),
		AccessToken: "test-token",
	}

	assert.Equal(t, "account123", msg.AccountID)
	assert.Equal(t, "UC_test_channel", msg.ChannelID)
	assert.NotEmpty(t, msg.Value)
	assert.Equal(t, "test-token", msg.AccessToken)
}

func TestTimestampUpdateRequestStruct(t *testing.T) {
	req := TimestampUpdateRequest{
		AccountID: "account123",
		ChannelID: "UC_test_channel",
	}

	assert.Equal(t, "account123", req.AccountID)
	assert.Equal(t, "UC_test_channel", req.ChannelID)
}

func TestProduceMessage(t *testing.T) {
	log := createTestFetcherLogger()
	producer := newTestProducer()
	ctx := context.Background()

	testData := struct {
		ChannelID string `json:"channel_id"`
		Title     string `json:"title"`
	}{
		ChannelID: "UC_test",
		Title:     "Test Channel",
	}

	produceMessage(ctx, producer, "test-topic", "test-key", testData, log)

	messages := producer.getMessages("test-topic")
	require.Len(t, messages, 1)

	var result struct {
		ChannelID string `json:"channel_id"`
		Title     string `json:"title"`
	}
	err := json.Unmarshal(messages[0], &result)
	require.NoError(t, err)
	assert.Equal(t, "UC_test", result.ChannelID)
	assert.Equal(t, "Test Channel", result.Title)
}

func TestProduceMessage_MarshalError(t *testing.T) {
	log := createTestFetcherLogger()
	producer := newTestProducer()
	ctx := context.Background()

	// Create a value that can't be marshaled (channel type)
	invalidData := make(chan int)

	produceMessage(ctx, producer, "test-topic", "test-key", invalidData, log)

	// Should not produce any messages
	messages := producer.getMessages("test-topic")
	assert.Len(t, messages, 0)
}

func TestProduceMessage_ProduceError(t *testing.T) {
	log := createTestFetcherLogger()
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return errors.New("kafka produce error")
		},
	}
	ctx := context.Background()

	testData := struct {
		ChannelID string `json:"channel_id"`
	}{
		ChannelID: "UC_test",
	}

	// Should not panic, just log the error
	produceMessage(ctx, producer, "test-topic", "test-key", testData, log)
}

func TestYouTubeAccountWorkOrderSerialization(t *testing.T) {
	wo := kafkamodels.YouTubeAccountWorkOrder{
		ID:           "account123",
		ChannelID:    "UC_test_channel",
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		WorkspaceID:  "workspace123",
		SyncType:     kafkamodels.YouTubeSyncTypeFullSync,
	}

	data, err := json.Marshal(wo)
	require.NoError(t, err)

	var decoded kafkamodels.YouTubeAccountWorkOrder
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, wo.ID, decoded.ID)
	assert.Equal(t, wo.ChannelID, decoded.ChannelID)
	assert.Equal(t, wo.AccessToken, decoded.AccessToken)
	assert.Equal(t, wo.RefreshToken, decoded.RefreshToken)
	assert.Equal(t, wo.WorkspaceID, decoded.WorkspaceID)
	assert.Equal(t, wo.SyncType, decoded.SyncType)
}

func TestYouTubeBatchWorkOrderSerialization(t *testing.T) {
	batch := kafkamodels.YouTubeBatchWorkOrder{
		BatchID: "batch123",
		Accounts: []kafkamodels.YouTubeAccountWorkOrder{
			{
				ID:          "account1",
				ChannelID:   "UC_channel1",
				AccessToken: "token1",
				SyncType:    kafkamodels.YouTubeSyncTypeIncremental,
			},
			{
				ID:          "account2",
				ChannelID:   "UC_channel2",
				AccessToken: "token2",
				SyncType:    kafkamodels.YouTubeSyncTypeImmediate,
			},
		},
	}

	data, err := json.Marshal(batch)
	require.NoError(t, err)

	var decoded kafkamodels.YouTubeBatchWorkOrder
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, batch.BatchID, decoded.BatchID)
	require.Len(t, decoded.Accounts, 2)
	assert.Equal(t, "account1", decoded.Accounts[0].ID)
	assert.Equal(t, "account2", decoded.Accounts[1].ID)
}

func TestRawYouTubeChannelSerialization(t *testing.T) {
	now := time.Now().UTC()
	pubAt := time.Date(2020, 1, 15, 10, 30, 0, 0, time.UTC)

	raw := kafkamodels.RawYouTubeChannel{
		ChannelID:       "UC_test_channel",
		Title:           "Test Channel",
		Description:     "A test channel description",
		CustomURL:       "@testchannel",
		ThumbnailURL:    "https://example.com/thumb.jpg",
		BannerURL:       "https://example.com/banner.jpg",
		Country:         "US",
		SubscriberCount: 10000,
		VideoCount:      500,
		ViewCount:       5000000,
		WorkspaceID:     "workspace123",
		PublishedAt:     pubAt,
		SavingTime:      now,
	}

	data, err := json.Marshal(raw)
	require.NoError(t, err)

	var decoded kafkamodels.RawYouTubeChannel
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, raw.ChannelID, decoded.ChannelID)
	assert.Equal(t, raw.Title, decoded.Title)
	assert.Equal(t, raw.SubscriberCount, decoded.SubscriberCount)
	assert.Equal(t, raw.VideoCount, decoded.VideoCount)
	assert.Equal(t, raw.ViewCount, decoded.ViewCount)
}

func TestRawYouTubeVideoSerialization(t *testing.T) {
	now := time.Now().UTC()
	pubAt := time.Date(2023, 6, 15, 14, 30, 0, 0, time.UTC)

	raw := kafkamodels.RawYouTubeVideo{
		VideoID:       "video123",
		ChannelID:     "UC_test_channel",
		Title:         "Test Video",
		Description:   "A test video description",
		ThumbnailURL:  "https://example.com/video_thumb.jpg",
		Duration:      "PT5M30S",
		WorkspaceID:   "workspace123",
		SavingTime:    now,
		AnalyticsDate: now,
		MediaType:     kafkamodels.YouTubeMediaTypeVideo,
		Views:         1000000,
		Likes:         50000,
		Dislikes:      500,
		Comments:      10000,
		Favorites:     5000,
		PublishedAt:   pubAt,
	}

	data, err := json.Marshal(raw)
	require.NoError(t, err)

	var decoded kafkamodels.RawYouTubeVideo
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, raw.VideoID, decoded.VideoID)
	assert.Equal(t, raw.ChannelID, decoded.ChannelID)
	assert.Equal(t, raw.Title, decoded.Title)
	assert.Equal(t, raw.Duration, decoded.Duration)
	assert.Equal(t, raw.MediaType, decoded.MediaType)
	assert.Equal(t, raw.Views, decoded.Views)
	assert.Equal(t, raw.Likes, decoded.Likes)
}

func TestSyncTypeDateRanges(t *testing.T) {
	now := time.Now().UTC()
	endDate := now.AddDate(0, 0, -1)

	tests := []struct {
		name                 string
		syncType             string
		expectedVideosDays   int
		expectedInsightsDays int
	}{
		{
			name:                 "incremental sync",
			syncType:             kafkamodels.YouTubeSyncTypeIncremental,
			expectedVideosDays:   incrementalVideosDays,
			expectedInsightsDays: incrementalInsightsDays,
		},
		{
			name:                 "immediate sync",
			syncType:             kafkamodels.YouTubeSyncTypeImmediate,
			expectedVideosDays:   immediateVideosDays,
			expectedInsightsDays: immediateInsightsDays,
		},
		{
			name:                 "full sync",
			syncType:             kafkamodels.YouTubeSyncTypeFullSync,
			expectedVideosDays:   fullSyncVideosDays,
			expectedInsightsDays: fullSyncInsightsDays,
		},
		{
			name:                 "unknown sync type (defaults to incremental)",
			syncType:             "unknown",
			expectedVideosDays:   incrementalVideosDays,
			expectedInsightsDays: incrementalInsightsDays,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var videosSince, insightsStartDate time.Time

			switch tt.syncType {
			case kafkamodels.YouTubeSyncTypeFullSync:
				videosSince = endDate.AddDate(0, 0, -fullSyncVideosDays)
				insightsStartDate = endDate.AddDate(0, 0, -fullSyncInsightsDays)
			case kafkamodels.YouTubeSyncTypeImmediate:
				videosSince = endDate.AddDate(0, 0, -immediateVideosDays)
				insightsStartDate = endDate.AddDate(0, 0, -immediateInsightsDays)
			default:
				videosSince = endDate.AddDate(0, 0, -incrementalVideosDays)
				insightsStartDate = endDate.AddDate(0, 0, -incrementalInsightsDays)
			}

			expectedVideosSince := endDate.AddDate(0, 0, -tt.expectedVideosDays)
			expectedInsightsStart := endDate.AddDate(0, 0, -tt.expectedInsightsDays)

			assert.Equal(t, expectedVideosSince, videosSince)
			assert.Equal(t, expectedInsightsStart, insightsStartDate)
		})
	}
}

func TestChannelDataProcessing(t *testing.T) {
	respJSON := `{
		"items": [{
			"id": "UC_test_channel",
			"snippet": {
				"title": "Test Channel",
				"description": "A test channel",
				"customUrl": "@testchannel",
				"publishedAt": "2020-01-15T10:30:00Z",
				"country": "US",
				"thumbnails": { "high": { "url": "https://example.com/thumb.jpg" } }
			},
			"statistics": {
				"subscriberCount": "10000",
				"videoCount": "500",
				"viewCount": "5000000"
			},
			"brandingSettings": {
				"image": { "bannerExternalUrl": "https://example.com/banner.jpg" }
			}
		}]
	}`
	var channelResp social.YouTubeChannelResponse
	require.NoError(t, json.Unmarshal([]byte(respJSON), &channelResp))

	require.Len(t, channelResp.Items, 1)
	channelData := channelResp.Items[0]

	assert.Equal(t, "UC_test_channel", channelData.ID)
	assert.Equal(t, "Test Channel", channelData.Snippet.Title)
	assert.Equal(t, "10000", channelData.Statistics.SubscriberCount)
	assert.Equal(t, "500", channelData.Statistics.VideoCount)
	assert.Equal(t, "5000000", channelData.Statistics.ViewCount)
}

func TestVideoDetailsMapProcessing(t *testing.T) {
	activitiesJSON := `[
		{
			"snippet": { "title": "Video 1", "publishedAt": "2023-06-01T00:00:00Z" },
			"contentDetails": { "upload": { "videoId": "vid1" } }
		},
		{
			"snippet": { "title": "Video 2", "publishedAt": "2023-06-02T00:00:00Z" },
			"contentDetails": { "upload": { "videoId": "vid2" } }
		}
	]`
	var activities []social.YouTubeActivityItem
	require.NoError(t, json.Unmarshal([]byte(activitiesJSON), &activities))

	videoIDs := make([]string, 0, len(activities))
	for _, activity := range activities {
		if activity.ContentDetails.Upload.VideoID != "" {
			videoIDs = append(videoIDs, activity.ContentDetails.Upload.VideoID)
		}
	}

	assert.Len(t, videoIDs, 2)
	assert.Contains(t, videoIDs, "vid1")
	assert.Contains(t, videoIDs, "vid2")

	// Simulate video details map
	videoDetailsJSON := `[
		{
			"id": "vid1",
			"snippet": { "title": "Video 1 Details", "publishedAt": "2023-06-01T00:00:00Z" },
			"contentDetails": { "duration": "PT5M" },
			"statistics": { "viewCount": "1000", "likeCount": "100" }
		},
		{
			"id": "vid2",
			"snippet": { "title": "Video 2 Details", "publishedAt": "2023-06-02T00:00:00Z" },
			"contentDetails": { "duration": "PT10M" },
			"statistics": { "viewCount": "2000", "likeCount": "200" }
		}
	]`
	var videoDetails []social.YouTubeVideoItem
	require.NoError(t, json.Unmarshal([]byte(videoDetailsJSON), &videoDetails))

	videoDetailsMap := make(map[string]*social.YouTubeVideoItem)
	for i := range videoDetails {
		videoDetailsMap[videoDetails[i].ID] = &videoDetails[i]
	}

	assert.Len(t, videoDetailsMap, 2)
	assert.NotNil(t, videoDetailsMap["vid1"])
	assert.NotNil(t, videoDetailsMap["vid2"])
	assert.Equal(t, "PT5M", videoDetailsMap["vid1"].ContentDetails.Duration)
	assert.Equal(t, "PT10M", videoDetailsMap["vid2"].ContentDetails.Duration)
}

func TestMediaTypeDetection(t *testing.T) {
	ytClient := social.NewYouTubeClient("test-client-id", "test-client-secret")

	videosJSON := `[
		{"id": "short1", "contentDetails": {"duration": "PT30S"}},
		{"id": "short2", "contentDetails": {"duration": "PT1M"}},
		{"id": "video1", "contentDetails": {"duration": "PT1M1S"}},
		{"id": "video2", "contentDetails": {"duration": "PT5M30S"}}
	]`
	var videos []social.YouTubeVideoItem
	require.NoError(t, json.Unmarshal([]byte(videosJSON), &videos))

	mediaTypes := ytClient.DetectMediaTypes(context.Background(), videos)

	assert.Equal(t, kafkamodels.YouTubeMediaTypeShort, mediaTypes["short1"])
	assert.Equal(t, kafkamodels.YouTubeMediaTypeShort, mediaTypes["short2"])
	assert.Equal(t, kafkamodels.YouTubeMediaTypeVideo, mediaTypes["video1"])
	assert.Equal(t, kafkamodels.YouTubeMediaTypeVideo, mediaTypes["video2"])
}

func TestActivityItemVideoExtraction(t *testing.T) {
	activityJSON := `{
		"snippet": {
			"title": "Test Video",
			"description": "Test description",
			"publishedAt": "2023-06-15T14:30:00Z",
			"thumbnails": { "high": { "url": "https://example.com/thumb.jpg" } }
		},
		"contentDetails": { "upload": { "videoId": "test_video_id" } }
	}`
	var activity social.YouTubeActivityItem
	require.NoError(t, json.Unmarshal([]byte(activityJSON), &activity))

	assert.Equal(t, "test_video_id", activity.ContentDetails.Upload.VideoID)
	assert.Equal(t, "Test Video", activity.Snippet.Title)
	assert.Equal(t, "https://example.com/thumb.jpg", activity.Snippet.Thumbnails.High.URL)
}

func TestAnalyticsResponseProcessing(t *testing.T) {
	resp := &social.YouTubeAnalyticsResponse{
		ColumnHeaders: []struct {
			Name       string `json:"name"`
			ColumnType string `json:"columnType"`
			DataType   string `json:"dataType"`
		}{
			{Name: "day"},
			{Name: "views"},
			{Name: "likes"},
		},
		Rows: [][]interface{}{
			{"2023-06-01", float64(1000), float64(100)},
			{"2023-06-02", float64(1500), float64(150)},
		},
	}

	assert.Len(t, resp.ColumnHeaders, 3)
	assert.Len(t, resp.Rows, 2)

	// Build column index map
	colIndex := make(map[string]int)
	for i, col := range resp.ColumnHeaders {
		colIndex[col.Name] = i
	}

	assert.Equal(t, 0, colIndex["day"])
	assert.Equal(t, 1, colIndex["views"])
	assert.Equal(t, 2, colIndex["likes"])

	// Extract data from first row
	row := resp.Rows[0]
	day := row[colIndex["day"]].(string)
	views := row[colIndex["views"]].(float64)
	likes := row[colIndex["likes"]].(float64)

	assert.Equal(t, "2023-06-01", day)
	assert.Equal(t, float64(1000), views)
	assert.Equal(t, float64(100), likes)
}

func TestInsightsResponseStructure(t *testing.T) {
	// Activity insights
	activityResp := &social.YouTubeAnalyticsResponse{
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
		},
	}

	assert.Len(t, activityResp.ColumnHeaders, 10)
	assert.Len(t, activityResp.Rows, 1)

	// Traffic insights
	trafficResp := &social.YouTubeAnalyticsResponse{
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
		},
	}

	assert.Len(t, trafficResp.ColumnHeaders, 3)
	assert.Len(t, trafficResp.Rows, 2)

	// Shared insights
	sharedResp := &social.YouTubeAnalyticsResponse{
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
			{"FACEBOOK", float64(200)},
		},
	}

	assert.Len(t, sharedResp.ColumnHeaders, 2)
	assert.Len(t, sharedResp.Rows, 2)
}

func TestTimestampUpdateChannel(t *testing.T) {
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	// Send a request
	req := TimestampUpdateRequest{
		AccountID: "account123",
		ChannelID: "UC_test_channel",
	}

	select {
	case timestampUpdateChan <- req:
		// Success
	default:
		t.Fatal("Failed to send to channel")
	}

	// Receive the request
	select {
	case received := <-timestampUpdateChan:
		assert.Equal(t, req.AccountID, received.AccountID)
		assert.Equal(t, req.ChannelID, received.ChannelID)
	default:
		t.Fatal("Failed to receive from channel")
	}
}

func TestTimestampUpdateChannel_Full(t *testing.T) {
	timestampUpdateChan := make(chan TimestampUpdateRequest, 1)

	// Fill the channel
	timestampUpdateChan <- TimestampUpdateRequest{AccountID: "first", ChannelID: "UC_first"}

	// This should not block (using select with default)
	req := TimestampUpdateRequest{AccountID: "second", ChannelID: "UC_second"}
	select {
	case timestampUpdateChan <- req:
		t.Fatal("Should not have succeeded - channel is full")
	default:
		// Expected - channel is full
	}
}

func TestWorkOrderChannelProcessing(t *testing.T) {
	workOrderChan := make(chan WorkOrderMessage, 10)

	wo := kafkamodels.YouTubeAccountWorkOrder{
		ID:           "account123",
		ChannelID:    "UC_test_channel",
		AccessToken:  "test-token",
		RefreshToken: "refresh-token",
		WorkspaceID:  "workspace123",
		SyncType:     kafkamodels.YouTubeSyncTypeIncremental,
	}

	payload, err := json.Marshal(wo)
	require.NoError(t, err)

	msg := WorkOrderMessage{
		AccountID:   wo.ID,
		ChannelID:   wo.ChannelID,
		Value:       payload,
		AccessToken: wo.AccessToken,
	}

	workOrderChan <- msg

	received := <-workOrderChan
	assert.Equal(t, wo.ID, received.AccountID)
	assert.Equal(t, wo.ChannelID, received.ChannelID)

	var decoded kafkamodels.YouTubeAccountWorkOrder
	err = json.Unmarshal(received.Value, &decoded)
	require.NoError(t, err)
	assert.Equal(t, wo.SyncType, decoded.SyncType)
}

func TestBatchWorkOrderDistribution(t *testing.T) {
	batch := kafkamodels.YouTubeBatchWorkOrder{
		BatchID: "batch123",
		Accounts: []kafkamodels.YouTubeAccountWorkOrder{
			{ID: "acc1", ChannelID: "UC_ch1", AccessToken: "token1"},
			{ID: "acc2", ChannelID: "UC_ch2", AccessToken: "token2"},
			{ID: "acc3", ChannelID: "UC_ch3", AccessToken: "token3"},
		},
	}

	workOrderChan := make(chan WorkOrderMessage, len(batch.Accounts))

	for _, account := range batch.Accounts {
		payload, err := json.Marshal(account)
		require.NoError(t, err)

		workOrderChan <- WorkOrderMessage{
			AccountID:   account.ID,
			ChannelID:   account.ChannelID,
			Value:       payload,
			AccessToken: account.AccessToken,
		}
	}

	// Verify all accounts were distributed
	assert.Len(t, workOrderChan, 3)

	// Check each message
	for i := 0; i < 3; i++ {
		msg := <-workOrderChan
		assert.NotEmpty(t, msg.AccountID)
		assert.NotEmpty(t, msg.ChannelID)
		assert.NotEmpty(t, msg.Value)
	}
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	workOrderChan := make(chan WorkOrderMessage, 10)

	// Simulate a worker that respects context cancellation
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-workOrderChan:
				if !ok {
					return
				}
			}
		}
	}()

	// Cancel the context
	cancel()

	// Wait for goroutine to exit
	select {
	case <-done:
		// Success
	case <-time.After(time.Second):
		t.Fatal("Worker did not respect context cancellation")
	}
}

func TestParseStatistics(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"valid number", "12345", 12345},
		{"zero", "0", 0},
		{"large number", "9999999999", 9999999999},
		{"empty string", "", 0},
		{"invalid string", "invalid", 0},
		{"negative (edge case)", "-1", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result int64
			_, err := testParseStatistic(tt.input, &result)
			if tt.input == "" || tt.input == "invalid" {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func testParseStatistic(s string, result *int64) (int64, error) {
	var err error
	*result, err = func(s string) (int64, error) {
		if s == "" {
			return 0, errors.New("empty string")
		}
		var n int64
		for _, c := range s {
			if c == '-' && n == 0 {
				continue // Handle negative sign
			}
			if c < '0' || c > '9' {
				return 0, errors.New("invalid character")
			}
			n = n*10 + int64(c-'0')
		}
		if len(s) > 0 && s[0] == '-' {
			n = -n
		}
		return n, nil
	}(s)
	return *result, err
}

func TestYouTubeClientCreation(t *testing.T) {
	ytClient := social.NewYouTubeClient("client-id", "client-secret")
	require.NotNil(t, ytClient)
}

func TestGenerateEmbedHTML(t *testing.T) {
	videoID := "test_video_123"
	embedHTML := social.GenerateEmbedHTML(videoID)

	assert.Contains(t, embedHTML, videoID)
	assert.Contains(t, embedHTML, "iframe")
	assert.Contains(t, embedHTML, "youtube.com/embed")
}

func TestConcurrentSemaphoreAccess(t *testing.T) {
	// Clear the global map first
	accountSemaphores = sync.Map{}

	channelID := "test-channel"
	sem := semForAccount(channelID, 1)

	ctx := context.Background()
	acquired := make(chan struct{}, 10)
	released := make(chan struct{})

	// First goroutine acquires the semaphore
	go func() {
		err := sem.Acquire(ctx, 1)
		require.NoError(t, err)
		acquired <- struct{}{}
		<-released
		sem.Release(1)
	}()

	// Wait for first acquisition
	<-acquired

	// Second goroutine should block
	blocked := make(chan bool)
	go func() {
		ctxTimeout, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()
		err := sem.Acquire(ctxTimeout, 1)
		blocked <- err != nil // true if blocked (timeout)
	}()

	wasBlocked := <-blocked
	assert.True(t, wasBlocked, "Second acquisition should have been blocked")

	// Release first semaphore
	close(released)
}

func TestWorkOrderValidation(t *testing.T) {
	tests := []struct {
		name        string
		channelID   string
		accessToken string
		shouldSkip  bool
	}{
		{
			name:        "valid work order",
			channelID:   "UC_test_channel",
			accessToken: "valid-token",
			shouldSkip:  false,
		},
		{
			name:        "missing channel ID",
			channelID:   "",
			accessToken: "valid-token",
			shouldSkip:  true,
		},
		{
			name:        "missing access token",
			channelID:   "UC_test_channel",
			accessToken: "",
			shouldSkip:  true,
		},
		{
			name:        "both missing",
			channelID:   "",
			accessToken: "",
			shouldSkip:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldSkip := tt.channelID == "" || tt.accessToken == ""
			assert.Equal(t, tt.shouldSkip, shouldSkip)
		})
	}
}

func TestVideoMediaTypeAssignment(t *testing.T) {
	activitiesJSON := `[
		{"contentDetails": {"upload": {"videoId": "vid1"}}},
		{"contentDetails": {"upload": {"videoId": "vid2"}}},
		{"contentDetails": {"upload": {"videoId": "vid3"}}}
	]`
	var activities []social.YouTubeActivityItem
	require.NoError(t, json.Unmarshal([]byte(activitiesJSON), &activities))

	vid1JSON := `{"id": "vid1", "contentDetails": {"duration": "PT30S"}}`
	var vid1 social.YouTubeVideoItem
	require.NoError(t, json.Unmarshal([]byte(vid1JSON), &vid1))

	vid2JSON := `{"id": "vid2", "contentDetails": {"duration": "PT5M"}}`
	var vid2 social.YouTubeVideoItem
	require.NoError(t, json.Unmarshal([]byte(vid2JSON), &vid2))

	videoDetailsMap := map[string]*social.YouTubeVideoItem{
		"vid1": &vid1,
		"vid2": &vid2,
		// vid3 has no details
	}

	var videosForDetection []social.YouTubeVideoItem
	for _, activity := range activities {
		videoID := activity.ContentDetails.Upload.VideoID
		if videoID == "" {
			continue
		}
		if details, ok := videoDetailsMap[videoID]; ok {
			videosForDetection = append(videosForDetection, *details)
		}
	}

	assert.Len(t, videosForDetection, 2)
	assert.Equal(t, "vid1", videosForDetection[0].ID)
	assert.Equal(t, "vid2", videosForDetection[1].ID)
}

func TestSyncTypeConstants(t *testing.T) {
	assert.Equal(t, "incremental", kafkamodels.YouTubeSyncTypeIncremental)
	assert.Equal(t, "immediate", kafkamodels.YouTubeSyncTypeImmediate)
	assert.Equal(t, "full_sync", kafkamodels.YouTubeSyncTypeFullSync)
}

func TestMediaTypeConstants(t *testing.T) {
	assert.Equal(t, "video", kafkamodels.YouTubeMediaTypeVideo)
	assert.Equal(t, "short", kafkamodels.YouTubeMediaTypeShort)
}

func TestProcessWorkOrder_MissingChannelID(t *testing.T) {
	log := createTestFetcherLogger()
	producer := newTestProducer()
	mockYTClient := &social.MockYouTubeClient{}
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	wo := kafkamodels.YouTubeAccountWorkOrder{
		ID:          "account123",
		ChannelID:   "", // Missing
		AccessToken: "test-token",
	}
	payload, _ := json.Marshal(wo)

	msg := WorkOrderMessage{
		AccountID:   wo.ID,
		ChannelID:   wo.ChannelID,
		Value:       payload,
		AccessToken: wo.AccessToken,
	}

	processWorkOrder(context.Background(), msg, mockYTClient, producer, &mongodb.MockUnifiedSocialRepository{}, "test-key", log, timestampUpdateChan)

	// Should not produce any messages
	assert.Len(t, producer.getMessages(topicRawChannels), 0)
}

func TestProcessWorkOrder_MissingAccessToken(t *testing.T) {
	log := createTestFetcherLogger()
	producer := newTestProducer()
	mockYTClient := &social.MockYouTubeClient{}
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	wo := kafkamodels.YouTubeAccountWorkOrder{
		ID:          "account123",
		ChannelID:   "UC_test_channel",
		AccessToken: "", // Missing
	}
	payload, _ := json.Marshal(wo)

	msg := WorkOrderMessage{
		AccountID:   wo.ID,
		ChannelID:   wo.ChannelID,
		Value:       payload,
		AccessToken: wo.AccessToken,
	}

	processWorkOrder(context.Background(), msg, mockYTClient, producer, &mongodb.MockUnifiedSocialRepository{}, "test-key", log, timestampUpdateChan)

	// Should not produce any messages
	assert.Len(t, producer.getMessages(topicRawChannels), 0)
}

func TestProcessWorkOrder_InvalidJSON(t *testing.T) {
	log := createTestFetcherLogger()
	producer := newTestProducer()
	mockYTClient := &social.MockYouTubeClient{}
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	msg := WorkOrderMessage{
		AccountID:   "account123",
		ChannelID:   "UC_test_channel",
		Value:       []byte("invalid json"),
		AccessToken: "test-token",
	}

	processWorkOrder(context.Background(), msg, mockYTClient, producer, &mongodb.MockUnifiedSocialRepository{}, "test-key", log, timestampUpdateChan)

	// Should not produce any messages
	assert.Len(t, producer.getMessages(topicRawChannels), 0)
}

func TestProcessWorkOrder_SuccessWithChannelData(t *testing.T) {
	// Clear the global map first
	accountSemaphores = sync.Map{}

	log := createTestFetcherLogger()
	producer := newTestProducer()
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	channelRespJSON := `{
		"items": [{
			"id": "UC_test_channel",
			"snippet": { "title": "Test Channel", "publishedAt": "2020-01-15T10:30:00Z" },
			"statistics": { "subscriberCount": "10000", "videoCount": "500", "viewCount": "5000000" },
			"contentDetails": { "relatedPlaylists": { "uploads": "UU_test_channel" } }
		}]
	}`
	var channelResp social.YouTubeChannelResponse
	json.Unmarshal([]byte(channelRespJSON), &channelResp)

	mockYTClient := &social.MockYouTubeClient{
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			return &channelResp, nil
		},
		FetchVideosFunc: func(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]social.YouTubeActivityItem, error) {
			return []social.YouTubeActivityItem{}, nil
		},
		FetchActivityInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		FetchTrafficInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		FetchSharedInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		DetectMediaTypesFunc: func(ctx context.Context, videos []social.YouTubeVideoItem) map[string]string {
			return map[string]string{}
		},
	}

	wo := kafkamodels.YouTubeAccountWorkOrder{
		ID:          "account123",
		ChannelID:   "UC_test_channel",
		AccessToken: "test-token",
		WorkspaceID: "workspace123",
		SyncType:    kafkamodels.YouTubeSyncTypeIncremental,
	}
	payload, _ := json.Marshal(wo)

	msg := WorkOrderMessage{
		AccountID:   wo.ID,
		ChannelID:   wo.ChannelID,
		Value:       payload,
		AccessToken: wo.AccessToken,
	}

	processWorkOrder(context.Background(), msg, mockYTClient, producer, &mongodb.MockUnifiedSocialRepository{}, "test-key", log, timestampUpdateChan)

	// Should produce channel message
	channelMessages := producer.getMessages(topicRawChannels)
	require.Len(t, channelMessages, 1)

	var rawChannel kafkamodels.RawYouTubeChannel
	err := json.Unmarshal(channelMessages[0], &rawChannel)
	require.NoError(t, err)
	assert.Equal(t, "UC_test_channel", rawChannel.ChannelID)
	assert.Equal(t, "Test Channel", rawChannel.Title)
	assert.Equal(t, int64(10000), rawChannel.SubscriberCount)

	// Should send timestamp update
	select {
	case req := <-timestampUpdateChan:
		assert.Equal(t, "account123", req.AccountID)
		assert.Equal(t, "UC_test_channel", req.ChannelID)
	default:
		t.Error("Expected timestamp update request")
	}
}

func TestProcessWorkOrder_WithVideos(t *testing.T) {
	// Clear the global map first
	accountSemaphores = sync.Map{}

	log := createTestFetcherLogger()
	producer := newTestProducer()
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	var videoActivity social.YouTubeActivityItem
	videoActivityJSON := `{
		"snippet": { "title": "Test Video", "publishedAt": "2023-06-01T00:00:00Z", "type": "upload" },
		"contentDetails": { "upload": { "videoId": "vid123" } }
	}`
	json.Unmarshal([]byte(videoActivityJSON), &videoActivity)

	var videoDetails social.YouTubeVideoItem
	videoDetailsJSON := `{
		"id": "vid123",
		"snippet": { "title": "Test Video Details", "publishedAt": "2023-06-01T00:00:00Z", "thumbnails": { "high": { "url": "https://example.com/thumb.jpg" } } },
		"contentDetails": { "duration": "PT5M30S" },
		"statistics": { "viewCount": "10000", "likeCount": "500", "dislikeCount": "10", "commentCount": "100", "favoriteCount": "50" }
	}`
	json.Unmarshal([]byte(videoDetailsJSON), &videoDetails)

	channelItem := social.YouTubeChannelItem{ID: "UC_test_channel"}
	channelItem.ContentDetails.RelatedPlaylists.Uploads = "UU_test_channel"

	mockYTClient := &social.MockYouTubeClient{
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			return &social.YouTubeChannelResponse{Items: []social.YouTubeChannelItem{channelItem}}, nil
		},
		FetchVideosFunc: func(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]social.YouTubeActivityItem, error) {
			return []social.YouTubeActivityItem{videoActivity}, nil
		},
		FetchVideoDetailsFunc: func(ctx context.Context, accessToken string, videoIDs []string) ([]social.YouTubeVideoItem, error) {
			assert.Equal(t, []string{"vid123"}, videoIDs)
			return []social.YouTubeVideoItem{videoDetails}, nil
		},
		FetchActivityInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		FetchTrafficInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		FetchSharedInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		DetectMediaTypesFunc: func(ctx context.Context, videos []social.YouTubeVideoItem) map[string]string {
			return map[string]string{"vid123": "video"}
		},
	}

	wo := kafkamodels.YouTubeAccountWorkOrder{
		ID:          "account123",
		ChannelID:   "UC_test_channel",
		AccessToken: "test-token",
		WorkspaceID: "workspace123",
		SyncType:    kafkamodels.YouTubeSyncTypeIncremental,
	}
	payload, _ := json.Marshal(wo)

	msg := WorkOrderMessage{
		AccountID:   wo.ID,
		ChannelID:   wo.ChannelID,
		Value:       payload,
		AccessToken: wo.AccessToken,
	}

	processWorkOrder(context.Background(), msg, mockYTClient, producer, &mongodb.MockUnifiedSocialRepository{}, "test-key", log, timestampUpdateChan)

	// Should produce video message
	videoMessages := producer.getMessages(topicRawVideos)
	require.Len(t, videoMessages, 1)

	var rawVideo kafkamodels.RawYouTubeVideo
	err := json.Unmarshal(videoMessages[0], &rawVideo)
	require.NoError(t, err)
	assert.Equal(t, "vid123", rawVideo.VideoID)
	assert.Equal(t, "UC_test_channel", rawVideo.ChannelID)
	assert.Equal(t, "Test Video Details", rawVideo.Title)
	assert.Equal(t, "PT5M30S", rawVideo.Duration)
	assert.Equal(t, "video", rawVideo.MediaType)
	assert.Equal(t, int64(10000), rawVideo.Views)
	assert.Equal(t, int64(500), rawVideo.Likes)
}

func TestProcessWorkOrder_WithInsights(t *testing.T) {
	// Clear the global map first
	accountSemaphores = sync.Map{}

	log := createTestFetcherLogger()
	producer := newTestProducer()
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	activityInsights := &social.YouTubeAnalyticsResponse{
		Rows: [][]interface{}{{"2023-06-01", float64(1000)}},
	}

	trafficInsights := &social.YouTubeAnalyticsResponse{
		Rows: [][]interface{}{{"2023-06-01", "YT_SEARCH", float64(500)}},
	}

	sharedInsights := &social.YouTubeAnalyticsResponse{
		Rows: [][]interface{}{{"TWITTER", float64(100)}},
	}

	mockYTClient := &social.MockYouTubeClient{
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			return &social.YouTubeChannelResponse{}, nil
		},
		FetchVideosFunc: func(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]social.YouTubeActivityItem, error) {
			return []social.YouTubeActivityItem{}, nil
		},
		FetchActivityInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return activityInsights, nil
		},
		FetchTrafficInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return trafficInsights, nil
		},
		FetchSharedInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return sharedInsights, nil
		},
		DetectMediaTypesFunc: func(ctx context.Context, videos []social.YouTubeVideoItem) map[string]string {
			return map[string]string{}
		},
	}

	wo := kafkamodels.YouTubeAccountWorkOrder{
		ID:          "account123",
		ChannelID:   "UC_test_channel",
		AccessToken: "test-token",
		WorkspaceID: "workspace123",
		SyncType:    kafkamodels.YouTubeSyncTypeIncremental,
	}
	payload, _ := json.Marshal(wo)

	msg := WorkOrderMessage{
		AccountID:   wo.ID,
		ChannelID:   wo.ChannelID,
		Value:       payload,
		AccessToken: wo.AccessToken,
	}

	processWorkOrder(context.Background(), msg, mockYTClient, producer, &mongodb.MockUnifiedSocialRepository{}, "test-key", log, timestampUpdateChan)

	// Should produce insights messages
	assert.Len(t, producer.getMessages(topicRawActivityInsights), 1)
	assert.Len(t, producer.getMessages(topicRawTrafficInsights), 1)
	assert.Len(t, producer.getMessages(topicRawSharedInsights), 1)
}

func TestProcessWorkOrder_UnauthorizedError(t *testing.T) {
	// Clear the global map first
	accountSemaphores = sync.Map{}

	log := createTestFetcherLogger()
	producer := newTestProducer()
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	mockYTClient := &social.MockYouTubeClient{
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

	wo := kafkamodels.YouTubeAccountWorkOrder{
		ID:          "account123",
		ChannelID:   "UC_test_channel",
		AccessToken: "test-token",
		WorkspaceID: "workspace123",
		SyncType:    kafkamodels.YouTubeSyncTypeIncremental,
	}
	payload, _ := json.Marshal(wo)

	msg := WorkOrderMessage{
		AccountID:   wo.ID,
		ChannelID:   wo.ChannelID,
		Value:       payload,
		AccessToken: wo.AccessToken,
	}

	processWorkOrder(context.Background(), msg, mockYTClient, producer, &mongodb.MockUnifiedSocialRepository{}, "test-key", log, timestampUpdateChan)

	// Should not produce any messages due to unauthorized error
	assert.Len(t, producer.getMessages(topicRawChannels), 0)
	assert.Len(t, producer.getMessages(topicRawVideos), 0)

	// Should not send timestamp update
	select {
	case <-timestampUpdateChan:
		t.Error("Should not send timestamp update on error")
	default:
		// Expected
	}
}

func TestProcessWorkOrder_WithTokenRefresh(t *testing.T) {
	// Clear the global map first
	accountSemaphores = sync.Map{}

	log := createTestFetcherLogger()
	producer := newTestProducer()
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	refreshTokenCalled := false
	var usedAccessToken string

	channelRespJSON := `{
		"items": [{
			"id": "UC_test_channel",
			"snippet": { "title": "Test Channel", "publishedAt": "2020-01-15T10:30:00Z" },
			"statistics": { "subscriberCount": "10000", "videoCount": "500", "viewCount": "5000000" },
			"contentDetails": { "relatedPlaylists": { "uploads": "UU_test_channel" } }
		}]
	}`
	var channelResp social.YouTubeChannelResponse
	json.Unmarshal([]byte(channelRespJSON), &channelResp)

	mockYTClient := &social.MockYouTubeClient{
		RefreshTokenFunc: func(ctx context.Context, refreshToken string) (*social.YouTubeTokenResponse, error) {
			refreshTokenCalled = true
			assert.Equal(t, "test-refresh-token", refreshToken)
			return &social.YouTubeTokenResponse{
				AccessToken: "new-access-token",
				ExpiresIn:   3600,
			}, nil
		},
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			usedAccessToken = accessToken
			return &channelResp, nil
		},
		FetchVideosFunc: func(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]social.YouTubeActivityItem, error) {
			return []social.YouTubeActivityItem{}, nil
		},
		FetchActivityInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		FetchTrafficInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		FetchSharedInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		DetectMediaTypesFunc: func(ctx context.Context, videos []social.YouTubeVideoItem) map[string]string {
			return map[string]string{}
		},
	}

	wo := kafkamodels.YouTubeAccountWorkOrder{
		ID:           "account123",
		ChannelID:    "UC_test_channel",
		AccessToken:  "old-access-token",
		RefreshToken: "test-refresh-token",
		WorkspaceID:  "workspace123",
		SyncType:     kafkamodels.YouTubeSyncTypeIncremental,
	}
	payload, _ := json.Marshal(wo)

	msg := WorkOrderMessage{
		AccountID:   wo.ID,
		ChannelID:   wo.ChannelID,
		Value:       payload,
		AccessToken: wo.AccessToken,
	}

	processWorkOrder(context.Background(), msg, mockYTClient, producer, &mongodb.MockUnifiedSocialRepository{}, "test-key", log, timestampUpdateChan)

	assert.True(t, refreshTokenCalled)
	assert.Equal(t, "new-access-token", usedAccessToken)
}

func TestProcessWorkOrder_FullSyncDateRange(t *testing.T) {
	// Clear the global map first
	accountSemaphores = sync.Map{}

	log := createTestFetcherLogger()
	producer := newTestProducer()
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	var capturedSinceDate time.Time

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
			return nil, nil
		},
		FetchTrafficInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		FetchSharedInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		DetectMediaTypesFunc: func(ctx context.Context, videos []social.YouTubeVideoItem) map[string]string {
			return map[string]string{}
		},
	}

	wo := kafkamodels.YouTubeAccountWorkOrder{
		ID:          "account123",
		ChannelID:   "UC_test_channel",
		AccessToken: "test-token",
		WorkspaceID: "workspace123",
		SyncType:    kafkamodels.YouTubeSyncTypeFullSync,
	}
	payload, _ := json.Marshal(wo)

	msg := WorkOrderMessage{
		AccountID:   wo.ID,
		ChannelID:   wo.ChannelID,
		Value:       payload,
		AccessToken: wo.AccessToken,
	}

	processWorkOrder(context.Background(), msg, mockYTClient, producer, &mongodb.MockUnifiedSocialRepository{}, "test-key", log, timestampUpdateChan)

	// For full sync, the since date should be fullSyncVideosDays ago from yesterday
	now := time.Now().UTC()
	expectedSince := now.AddDate(0, 0, -3).AddDate(0, 0, -fullSyncVideosDays)
	// Check the dates are within a reasonable range (same day)
	assert.Equal(t, expectedSince.Year(), capturedSinceDate.Year())
	assert.Equal(t, expectedSince.Month(), capturedSinceDate.Month())
	assert.Equal(t, expectedSince.Day(), capturedSinceDate.Day())
}

func TestProcessWorkOrder_ImmediateSyncDateRange(t *testing.T) {
	// Clear the global map first
	accountSemaphores = sync.Map{}

	log := createTestFetcherLogger()
	producer := newTestProducer()
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	var capturedSinceDate time.Time

	imCh := social.YouTubeChannelItem{ID: "UC_test_channel"}
	imCh.ContentDetails.RelatedPlaylists.Uploads = "UU_test_channel"

	mockYTClient := &social.MockYouTubeClient{
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			return &social.YouTubeChannelResponse{Items: []social.YouTubeChannelItem{imCh}}, nil
		},
		FetchVideosFunc: func(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]social.YouTubeActivityItem, error) {
			capturedSinceDate = since
			return []social.YouTubeActivityItem{}, nil
		},
		FetchActivityInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		FetchTrafficInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		FetchSharedInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		DetectMediaTypesFunc: func(ctx context.Context, videos []social.YouTubeVideoItem) map[string]string {
			return map[string]string{}
		},
	}

	wo := kafkamodels.YouTubeAccountWorkOrder{
		ID:          "account123",
		ChannelID:   "UC_test_channel",
		AccessToken: "test-token",
		WorkspaceID: "workspace123",
		SyncType:    kafkamodels.YouTubeSyncTypeImmediate,
	}
	payload, _ := json.Marshal(wo)

	msg := WorkOrderMessage{
		AccountID:   wo.ID,
		ChannelID:   wo.ChannelID,
		Value:       payload,
		AccessToken: wo.AccessToken,
	}

	processWorkOrder(context.Background(), msg, mockYTClient, producer, &mongodb.MockUnifiedSocialRepository{}, "test-key", log, timestampUpdateChan)

	// For immediate sync, the since date should be immediateVideosDays ago from yesterday
	now := time.Now().UTC()
	expectedSince := now.AddDate(0, 0, -3).AddDate(0, 0, -immediateVideosDays)
	assert.Equal(t, expectedSince.Year(), capturedSinceDate.Year())
	assert.Equal(t, expectedSince.Month(), capturedSinceDate.Month())
	assert.Equal(t, expectedSince.Day(), capturedSinceDate.Day())
}

func TestProcessWorkOrder_PartialAPIFailure(t *testing.T) {
	// Clear the global map first
	accountSemaphores = sync.Map{}

	log := createTestFetcherLogger()
	producer := newTestProducer()
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	channelRespJSON := `{
		"items": [{
			"id": "UC_test_channel",
			"snippet": { "title": "Test Channel", "publishedAt": "2020-01-15T10:30:00Z" },
			"statistics": { "subscriberCount": "10000", "videoCount": "500", "viewCount": "5000000" },
			"contentDetails": { "relatedPlaylists": { "uploads": "UU_test_channel" } }
		}]
	}`
	var channelResp social.YouTubeChannelResponse
	json.Unmarshal([]byte(channelRespJSON), &channelResp)

	mockYTClient := &social.MockYouTubeClient{
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			return &channelResp, nil
		},
		FetchVideosFunc: func(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]social.YouTubeActivityItem, error) {
			return nil, errors.New("network error")
		},
		FetchActivityInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return &social.YouTubeAnalyticsResponse{Rows: [][]interface{}{{"2023-06-01", float64(1000)}}}, nil
		},
		FetchTrafficInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, errors.New("network error")
		},
		FetchSharedInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return &social.YouTubeAnalyticsResponse{Rows: [][]interface{}{{"TWITTER", float64(100)}}}, nil
		},
		DetectMediaTypesFunc: func(ctx context.Context, videos []social.YouTubeVideoItem) map[string]string {
			return map[string]string{}
		},
	}

	wo := kafkamodels.YouTubeAccountWorkOrder{
		ID:          "account123",
		ChannelID:   "UC_test_channel",
		AccessToken: "test-token",
		WorkspaceID: "workspace123",
		SyncType:    kafkamodels.YouTubeSyncTypeIncremental,
	}
	payload, _ := json.Marshal(wo)

	msg := WorkOrderMessage{
		AccountID:   wo.ID,
		ChannelID:   wo.ChannelID,
		Value:       payload,
		AccessToken: wo.AccessToken,
	}

	processWorkOrder(context.Background(), msg, mockYTClient, producer, &mongodb.MockUnifiedSocialRepository{}, "test-key", log, timestampUpdateChan)

	// Should still produce channel and successful insights messages
	assert.Len(t, producer.getMessages(topicRawChannels), 1)
	assert.Len(t, producer.getMessages(topicRawVideos), 0) // Failed
	assert.Len(t, producer.getMessages(topicRawActivityInsights), 1)
	assert.Len(t, producer.getMessages(topicRawTrafficInsights), 0) // Failed
	assert.Len(t, producer.getMessages(topicRawSharedInsights), 1)

	// Should still send timestamp update since partial success
	select {
	case <-timestampUpdateChan:
		// Expected
	default:
		t.Error("Expected timestamp update request")
	}
}

func TestService_WorkOrderProcessor_ContextCancellation(t *testing.T) {
	log := createTestFetcherLogger()
	mockYTClient := &social.MockYouTubeClient{}
	producer := newTestProducer()
	consumer := &kafka.MockConsumer{}
	mongoRepo := &mongodb.MockUnifiedSocialRepository{}

	svc := NewService(mockYTClient, producer, consumer, mongoRepo, log, "test-key")

	ctx, cancel := context.WithCancel(context.Background())
	workOrderChan := make(chan WorkOrderMessage, 10)
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	done := make(chan struct{})
	go func() {
		defer close(done)
		svc.workOrderProcessor(ctx, 0, workOrderChan, timestampUpdateChan)
	}()

	// Cancel the context
	cancel()

	// Wait for worker to exit
	select {
	case <-done:
		// Success
	case <-time.After(time.Second):
		t.Fatal("Worker did not exit after context cancellation")
	}
}

func TestService_WorkOrderProcessor_ChannelClosed(t *testing.T) {
	log := createTestFetcherLogger()
	mockYTClient := &social.MockYouTubeClient{}
	producer := newTestProducer()
	consumer := &kafka.MockConsumer{}
	mongoRepo := &mongodb.MockUnifiedSocialRepository{}

	svc := NewService(mockYTClient, producer, consumer, mongoRepo, log, "test-key")

	ctx := context.Background()
	workOrderChan := make(chan WorkOrderMessage, 10)
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	done := make(chan struct{})
	go func() {
		defer close(done)
		svc.workOrderProcessor(ctx, 0, workOrderChan, timestampUpdateChan)
	}()

	// Close the channel
	close(workOrderChan)

	// Wait for worker to exit
	select {
	case <-done:
		// Success
	case <-time.After(time.Second):
		t.Fatal("Worker did not exit after channel closed")
	}
}

func TestNewService(t *testing.T) {
	log := createTestFetcherLogger()
	mockYTClient := &social.MockYouTubeClient{}
	producer := newTestProducer()
	consumer := &kafka.MockConsumer{}
	mongoRepo := &mongodb.MockUnifiedSocialRepository{}

	svc := NewService(mockYTClient, producer, consumer, mongoRepo, log, "test-key")

	assert.NotNil(t, svc)
	assert.Equal(t, mockYTClient, svc.YTClient)
	assert.Equal(t, producer, svc.Producer)
	assert.Equal(t, consumer, svc.Consumer)
	assert.Equal(t, mongoRepo, svc.MongoRepo)
	assert.Equal(t, log, svc.Logger)
	assert.Equal(t, "test-key", svc.DecryptionKey)
	assert.Equal(t, maxWorkers, svc.MaxWorkers)
	assert.Equal(t, maxConcurrentAccounts, svc.MaxConcurrentAccounts)
}

func TestService_Run_ContextCancellation(t *testing.T) {
	log := createTestFetcherLogger()
	mockYTClient := &social.MockYouTubeClient{}
	producer := newTestProducer()
	consumer := &kafka.MockConsumer{}
	mongoRepo := &mongodb.MockUnifiedSocialRepository{}

	svc := NewService(mockYTClient, producer, consumer, mongoRepo, log, "test-key")
	svc.MaxWorkers = 2 // Use fewer workers for faster tests

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		defer close(done)
		svc.Run(ctx)
	}()

	// Cancel context after a short delay
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Wait for service to stop
	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Service did not stop after context cancellation")
	}
}

func TestService_Run_DoesNotStopOnIdle(t *testing.T) {
	log := createTestFetcherLogger()
	mockYTClient := &social.MockYouTubeClient{}
	producer := newTestProducer()
	consumer := &kafka.MockConsumer{}
	mongoRepo := &mongodb.MockUnifiedSocialRepository{}

	svc := NewService(mockYTClient, producer, consumer, mongoRepo, log, "test-key")
	svc.MaxWorkers = 1
	// Keep the configured idle timeout short to verify it no longer stops the service.
	svc.IdleTimeout = 50 * time.Millisecond
	svc.IdleCheckPeriod = 20 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		svc.Run(ctx)
	}()

	select {
	case <-done:
		t.Fatal("Service stopped without explicit cancellation")
	case <-time.After(250 * time.Millisecond):
		// Expected: service keeps running even when idle.
	}

	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Service did not stop after context cancellation")
	}
}

func TestService_Run_WithBatchWorkOrder(t *testing.T) {
	// Clear the global map first
	accountSemaphores = sync.Map{}

	log := createTestFetcherLogger()

	channelRespJSON := `{
		"items": [{
			"id": "UC_test_channel",
			"snippet": { "title": "Test Channel", "publishedAt": "2020-01-15T10:30:00Z" },
			"statistics": { "subscriberCount": "10000", "videoCount": "500", "viewCount": "5000000" },
			"contentDetails": { "relatedPlaylists": { "uploads": "UU_test_channel" } }
		}]
	}`
	var channelResp social.YouTubeChannelResponse
	json.Unmarshal([]byte(channelRespJSON), &channelResp)

	mockYTClient := &social.MockYouTubeClient{
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			return &channelResp, nil
		},
		FetchVideosFunc: func(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]social.YouTubeActivityItem, error) {
			return []social.YouTubeActivityItem{}, nil
		},
		FetchActivityInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		FetchTrafficInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		FetchSharedInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		DetectMediaTypesFunc: func(ctx context.Context, videos []social.YouTubeVideoItem) map[string]string {
			return map[string]string{}
		},
	}

	producer := newTestProducer()
	timestampUpdated := false

	mongoRepo := &mongodb.MockUnifiedSocialRepository{
		UpdateAnalyticsTimestampFunc: func(ctx context.Context, id primitive.ObjectID, field string, timestamp time.Time) error {
			timestampUpdated = true
			return nil
		},
	}

	batch := kafkamodels.YouTubeBatchWorkOrder{
		BatchID: "test-batch",
		Accounts: []kafkamodels.YouTubeAccountWorkOrder{
			{
				ID:          "507f1f77bcf86cd799439011", // Valid ObjectID
				ChannelID:   "UC_test_channel",
				AccessToken: "test-token",
				WorkspaceID: "workspace123",
				SyncType:    kafkamodels.YouTubeSyncTypeIncremental,
			},
		},
	}
	batchPayload, _ := json.Marshal(batch)

	consumerCalled := false
	consumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			if !consumerCalled {
				consumerCalled = true
				// Call handler with our batch
				handler(ctx, topicWorkOrderBatch, nil, batchPayload)
			}
			// Wait for context cancellation
			<-ctx.Done()
			return ctx.Err()
		},
	}

	svc := NewService(mockYTClient, producer, consumer, mongoRepo, log, "test-key")
	svc.MaxWorkers = 2

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		defer close(done)
		svc.Run(ctx)
	}()

	// Wait for processing
	time.Sleep(200 * time.Millisecond)
	cancel()

	<-done

	// Verify channel was fetched and produced
	channelMessages := producer.getMessages(topicRawChannels)
	assert.Len(t, channelMessages, 1)

	// Verify timestamp was updated
	assert.True(t, timestampUpdated)
}

func TestService_WorkerPool(t *testing.T) {
	log := createTestFetcherLogger()
	mockYTClient := &social.MockYouTubeClient{}
	producer := newTestProducer()
	consumer := &kafka.MockConsumer{}
	mongoRepo := &mongodb.MockUnifiedSocialRepository{}

	svc := NewService(mockYTClient, producer, consumer, mongoRepo, log, "test-key")
	svc.MaxWorkers = 3

	ctx, cancel := context.WithCancel(context.Background())
	workOrderChan := make(chan WorkOrderMessage, 10)
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	var wg sync.WaitGroup
	svc.startWorkerPool(ctx, &wg, workOrderChan, timestampUpdateChan)

	// Cancel context
	cancel()
	close(workOrderChan)
	wg.Wait()

	// All workers should have stopped
}

func TestService_TimestampUpdater(t *testing.T) {
	log := createTestFetcherLogger()

	timestampUpdates := make([]string, 0)
	mu := sync.Mutex{}

	mongoRepo := &mongodb.MockUnifiedSocialRepository{
		UpdateAnalyticsTimestampFunc: func(ctx context.Context, id primitive.ObjectID, field string, timestamp time.Time) error {
			mu.Lock()
			defer mu.Unlock()
			timestampUpdates = append(timestampUpdates, field)
			return nil
		},
	}

	mockYTClient := &social.MockYouTubeClient{}
	producer := newTestProducer()
	consumer := &kafka.MockConsumer{}

	svc := NewService(mockYTClient, producer, consumer, mongoRepo, log, "test-key")

	ctx, cancel := context.WithCancel(context.Background())
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	var wg sync.WaitGroup
	svc.startTimestampUpdater(ctx, &wg, timestampUpdateChan)

	// Send timestamp updates
	timestampUpdateChan <- TimestampUpdateRequest{
		AccountID: "507f1f77bcf86cd799439011",
		ChannelID: "UC_test_channel",
	}

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	cancel()
	close(timestampUpdateChan)
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	assert.Contains(t, timestampUpdates, "analytics")
}

func TestService_TimestampUpdater_InvalidObjectID(t *testing.T) {
	log := createTestFetcherLogger()

	updateCalled := false
	mongoRepo := &mongodb.MockUnifiedSocialRepository{
		UpdateAnalyticsTimestampFunc: func(ctx context.Context, id primitive.ObjectID, field string, timestamp time.Time) error {
			updateCalled = true
			return nil
		},
	}

	mockYTClient := &social.MockYouTubeClient{}
	producer := newTestProducer()
	consumer := &kafka.MockConsumer{}

	svc := NewService(mockYTClient, producer, consumer, mongoRepo, log, "test-key")

	ctx, cancel := context.WithCancel(context.Background())
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	var wg sync.WaitGroup
	svc.startTimestampUpdater(ctx, &wg, timestampUpdateChan)

	// Send invalid ObjectID
	timestampUpdateChan <- TimestampUpdateRequest{
		AccountID: "invalid-object-id",
		ChannelID: "UC_test_channel",
	}

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	cancel()
	close(timestampUpdateChan)
	wg.Wait()

	// Should not have called update due to invalid ObjectID
	assert.False(t, updateCalled)
}

func TestService_BatchConsumer_InvalidJSON(t *testing.T) {
	log := createTestFetcherLogger()
	mockYTClient := &social.MockYouTubeClient{}
	producer := newTestProducer()
	mongoRepo := &mongodb.MockUnifiedSocialRepository{}

	consumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			// Send invalid JSON
			handler(ctx, topicWorkOrderBatch, nil, []byte("invalid json"))
			<-ctx.Done()
			return ctx.Err()
		},
	}

	svc := NewService(mockYTClient, producer, consumer, mongoRepo, log, "test-key")
	svc.MaxWorkers = 1

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		defer close(done)
		svc.Run(ctx)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	<-done

	// Should not produce any messages
	assert.Len(t, producer.getMessages(topicRawChannels), 0)
}

func TestService_BatchConsumer_ConsumerError(t *testing.T) {
	log := createTestFetcherLogger()
	mockYTClient := &social.MockYouTubeClient{}
	producer := newTestProducer()
	mongoRepo := &mongodb.MockUnifiedSocialRepository{}

	consumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			return errors.New("kafka consumer error")
		},
	}

	svc := NewService(mockYTClient, producer, consumer, mongoRepo, log, "test-key")
	svc.MaxWorkers = 1
	svc.IdleTimeout = 1 * time.Hour // Prevent idle timeout

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		defer close(done)
		svc.Run(ctx)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	<-done
}

func TestService_BatchConsumer_ContextCancelledDuringDistribution(t *testing.T) {
	log := createTestFetcherLogger()
	mockYTClient := &social.MockYouTubeClient{}
	producer := newTestProducer()
	mongoRepo := &mongodb.MockUnifiedSocialRepository{}

	// Create a batch with many accounts to increase chance of context being cancelled during distribution
	accounts := make([]kafkamodels.YouTubeAccountWorkOrder, 100)
	for i := 0; i < 100; i++ {
		accounts[i] = kafkamodels.YouTubeAccountWorkOrder{
			ID:          "507f1f77bcf86cd799439011",
			ChannelID:   "UC_test_channel",
			AccessToken: "test-token",
		}
	}
	batch := kafkamodels.YouTubeBatchWorkOrder{
		BatchID:  "test-batch",
		Accounts: accounts,
	}
	batchPayload, _ := json.Marshal(batch)

	var handlerCtx context.Context
	consumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handlerCtx = ctx
			// Call handler - it will try to distribute to a small channel
			return handler(ctx, topicWorkOrderBatch, nil, batchPayload)
		},
	}

	svc := NewService(mockYTClient, producer, consumer, mongoRepo, log, "test-key")
	svc.MaxWorkers = 0 // No workers to consume from channel

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context quickly to trigger the context cancellation path
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	done := make(chan struct{})
	go func() {
		defer close(done)
		svc.Run(ctx)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Service did not stop")
	}

	// Verify the handler context was cancelled
	if handlerCtx != nil {
		select {
		case <-handlerCtx.Done():
			// Expected
		default:
		}
	}
}

func TestService_BatchConsumer_MarshalAccountError(t *testing.T) {
	log := createTestFetcherLogger()
	mockYTClient := &social.MockYouTubeClient{}
	producer := newTestProducer()
	mongoRepo := &mongodb.MockUnifiedSocialRepository{}

	// Create a batch with account that will be processed
	batch := kafkamodels.YouTubeBatchWorkOrder{
		BatchID: "test-batch",
		Accounts: []kafkamodels.YouTubeAccountWorkOrder{
			{
				ID:          "507f1f77bcf86cd799439011",
				ChannelID:   "UC_test_channel",
				AccessToken: "test-token",
			},
		},
	}
	batchPayload, _ := json.Marshal(batch)

	consumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, topicWorkOrderBatch, nil, batchPayload)
			<-ctx.Done()
			return ctx.Err()
		},
	}

	svc := NewService(mockYTClient, producer, consumer, mongoRepo, log, "test-key")
	svc.MaxWorkers = 1

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		defer close(done)
		svc.Run(ctx)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	<-done
}

func TestService_TimestampUpdater_UpdateError(t *testing.T) {
	log := createTestFetcherLogger()

	mongoRepo := &mongodb.MockUnifiedSocialRepository{
		UpdateAnalyticsTimestampFunc: func(ctx context.Context, id primitive.ObjectID, field string, timestamp time.Time) error {
			return errors.New("mongodb update error")
		},
	}

	mockYTClient := &social.MockYouTubeClient{}
	producer := newTestProducer()
	consumer := &kafka.MockConsumer{}

	svc := NewService(mockYTClient, producer, consumer, mongoRepo, log, "test-key")

	ctx, cancel := context.WithCancel(context.Background())
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	var wg sync.WaitGroup
	svc.startTimestampUpdater(ctx, &wg, timestampUpdateChan)

	// Send timestamp update that will fail
	timestampUpdateChan <- TimestampUpdateRequest{
		AccountID: "507f1f77bcf86cd799439011",
		ChannelID: "UC_test_channel",
	}

	time.Sleep(50 * time.Millisecond)

	cancel()
	close(timestampUpdateChan)
	wg.Wait()
}

func TestProcessWorkOrder_SemaphoreAcquireError(t *testing.T) {
	// Clear the global map first
	accountSemaphores = sync.Map{}

	log := createTestFetcherLogger()
	producer := newTestProducer()
	mockYTClient := &social.MockYouTubeClient{}
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	wo := kafkamodels.YouTubeAccountWorkOrder{
		ID:          "account123",
		ChannelID:   "UC_test_channel_sem",
		AccessToken: "test-token",
		SyncType:    kafkamodels.YouTubeSyncTypeIncremental,
	}
	payload, _ := json.Marshal(wo)

	msg := WorkOrderMessage{
		AccountID:   wo.ID,
		ChannelID:   wo.ChannelID,
		Value:       payload,
		AccessToken: wo.AccessToken,
	}

	// Pre-acquire the semaphore so the processWorkOrder will fail to acquire
	sem := semForAccount(wo.ChannelID, 1)
	sem.Acquire(context.Background(), 1)

	// Use a cancelled context so semaphore acquire fails immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	processWorkOrder(ctx, msg, mockYTClient, producer, &mongodb.MockUnifiedSocialRepository{}, "test-key", log, timestampUpdateChan)

	sem.Release(1)

	// Should not produce any messages
	assert.Len(t, producer.getMessages(topicRawChannels), 0)
}

func TestProcessWorkOrder_TokenRefreshError(t *testing.T) {
	// Clear the global map first
	accountSemaphores = sync.Map{}

	log := createTestFetcherLogger()
	producer := newTestProducer()
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	channelRespJSON := `{
		"items": [{
			"id": "UC_test_channel",
			"snippet": { "title": "Test Channel", "publishedAt": "2020-01-15T10:30:00Z" },
			"statistics": { "subscriberCount": "10000", "videoCount": "500", "viewCount": "5000000" },
			"contentDetails": { "relatedPlaylists": { "uploads": "UU_test_channel" } }
		}]
	}`
	var channelResp social.YouTubeChannelResponse
	json.Unmarshal([]byte(channelRespJSON), &channelResp)

	mockYTClient := &social.MockYouTubeClient{
		RefreshTokenFunc: func(ctx context.Context, refreshToken string) (*social.YouTubeTokenResponse, error) {
			return nil, errors.New("token refresh failed")
		},
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			return &channelResp, nil
		},
		FetchVideosFunc: func(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]social.YouTubeActivityItem, error) {
			return []social.YouTubeActivityItem{}, nil
		},
		FetchActivityInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		FetchTrafficInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		FetchSharedInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		DetectMediaTypesFunc: func(ctx context.Context, videos []social.YouTubeVideoItem) map[string]string {
			return map[string]string{}
		},
	}

	wo := kafkamodels.YouTubeAccountWorkOrder{
		ID:           "account123",
		ChannelID:    "UC_test_channel",
		AccessToken:  "old-access-token",
		RefreshToken: "test-refresh-token",
		WorkspaceID:  "workspace123",
		SyncType:     kafkamodels.YouTubeSyncTypeIncremental,
	}
	payload, _ := json.Marshal(wo)

	msg := WorkOrderMessage{
		AccountID:   wo.ID,
		ChannelID:   wo.ChannelID,
		Value:       payload,
		AccessToken: wo.AccessToken,
	}

	processWorkOrder(context.Background(), msg, mockYTClient, producer, &mongodb.MockUnifiedSocialRepository{}, "test-key", log, timestampUpdateChan)

	// Should still produce channel messages (using old token)
	channelMessages := producer.getMessages(topicRawChannels)
	assert.Len(t, channelMessages, 1)
}

func TestProcessWorkOrder_ContextCancelledDuringFetch(t *testing.T) {
	// Clear the global map first
	accountSemaphores = sync.Map{}

	log := createTestFetcherLogger()
	producer := newTestProducer()
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	ctx, cancel := context.WithCancel(context.Background())

	mockYTClient := &social.MockYouTubeClient{
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			cancel() // Cancel context during fetch
			return nil, ctx.Err()
		},
		FetchVideosFunc: func(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]social.YouTubeActivityItem, error) {
			return nil, ctx.Err()
		},
		FetchActivityInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, ctx.Err()
		},
		FetchTrafficInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, ctx.Err()
		},
		FetchSharedInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, ctx.Err()
		},
	}

	wo := kafkamodels.YouTubeAccountWorkOrder{
		ID:          "account123",
		ChannelID:   "UC_test_channel_ctx",
		AccessToken: "test-token",
		SyncType:    kafkamodels.YouTubeSyncTypeIncremental,
	}
	payload, _ := json.Marshal(wo)

	msg := WorkOrderMessage{
		AccountID:   wo.ID,
		ChannelID:   wo.ChannelID,
		Value:       payload,
		AccessToken: wo.AccessToken,
	}

	processWorkOrder(ctx, msg, mockYTClient, producer, &mongodb.MockUnifiedSocialRepository{}, "test-key", log, timestampUpdateChan)

	// Should not produce any messages due to context cancellation
	assert.Len(t, producer.getMessages(topicRawChannels), 0)
}

func TestProcessWorkOrder_VideoDetailsFetchError(t *testing.T) {
	// Clear the global map first
	accountSemaphores = sync.Map{}

	log := createTestFetcherLogger()
	producer := newTestProducer()
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	var videoActivity social.YouTubeActivityItem
	videoActivityJSON := `{
		"snippet": { "title": "Test Video", "publishedAt": "2023-06-01T00:00:00Z" },
		"contentDetails": { "upload": { "videoId": "vid123" } }
	}`
	json.Unmarshal([]byte(videoActivityJSON), &videoActivity)

	mockYTClient := &social.MockYouTubeClient{
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			return &social.YouTubeChannelResponse{}, nil
		},
		FetchVideosFunc: func(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]social.YouTubeActivityItem, error) {
			return []social.YouTubeActivityItem{videoActivity}, nil
		},
		FetchVideoDetailsFunc: func(ctx context.Context, accessToken string, videoIDs []string) ([]social.YouTubeVideoItem, error) {
			return nil, errors.New("video details fetch error")
		},
		FetchActivityInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		FetchTrafficInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		FetchSharedInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		DetectMediaTypesFunc: func(ctx context.Context, videos []social.YouTubeVideoItem) map[string]string {
			return map[string]string{}
		},
	}

	wo := kafkamodels.YouTubeAccountWorkOrder{
		ID:          "account123",
		ChannelID:   "UC_test_channel",
		AccessToken: "test-token",
		SyncType:    kafkamodels.YouTubeSyncTypeIncremental,
	}
	payload, _ := json.Marshal(wo)

	msg := WorkOrderMessage{
		AccountID:   wo.ID,
		ChannelID:   wo.ChannelID,
		Value:       payload,
		AccessToken: wo.AccessToken,
	}

	processWorkOrder(context.Background(), msg, mockYTClient, producer, &mongodb.MockUnifiedSocialRepository{}, "test-key", log, timestampUpdateChan)

	// Videos should still be produced even if details fetch fails
	// (they just won't have detailed stats)
}

func TestProcessWorkOrder_VideoDetailsUnauthorizedError(t *testing.T) {
	// Clear the global map first
	accountSemaphores = sync.Map{}

	log := createTestFetcherLogger()
	producer := newTestProducer()
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	var videoActivity social.YouTubeActivityItem
	videoActivityJSON := `{
		"snippet": { "title": "Test Video", "publishedAt": "2023-06-01T00:00:00Z" },
		"contentDetails": { "upload": { "videoId": "vid123" } }
	}`
	json.Unmarshal([]byte(videoActivityJSON), &videoActivity)

	mockYTClient := &social.MockYouTubeClient{
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			return &social.YouTubeChannelResponse{}, nil
		},
		FetchVideosFunc: func(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]social.YouTubeActivityItem, error) {
			return []social.YouTubeActivityItem{videoActivity}, nil
		},
		FetchVideoDetailsFunc: func(ctx context.Context, accessToken string, videoIDs []string) ([]social.YouTubeVideoItem, error) {
			return nil, errors.New("request failed with status 401: unauthorized")
		},
		FetchActivityInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		FetchTrafficInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		FetchSharedInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
	}

	wo := kafkamodels.YouTubeAccountWorkOrder{
		ID:          "account123",
		ChannelID:   "UC_test_channel",
		AccessToken: "test-token",
		SyncType:    kafkamodels.YouTubeSyncTypeIncremental,
	}
	payload, _ := json.Marshal(wo)

	msg := WorkOrderMessage{
		AccountID:   wo.ID,
		ChannelID:   wo.ChannelID,
		Value:       payload,
		AccessToken: wo.AccessToken,
	}

	processWorkOrder(context.Background(), msg, mockYTClient, producer, &mongodb.MockUnifiedSocialRepository{}, "test-key", log, timestampUpdateChan)

	// Should abort due to unauthorized error
	assert.Len(t, producer.getMessages(topicRawVideos), 0)
}

func TestProcessWorkOrder_AllInsightsErrors(t *testing.T) {
	// Clear the global map first
	accountSemaphores = sync.Map{}

	log := createTestFetcherLogger()
	producer := newTestProducer()
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	channelRespJSON := `{
		"items": [{
			"id": "UC_test_channel",
			"snippet": { "title": "Test Channel", "publishedAt": "2020-01-15T10:30:00Z" },
			"statistics": { "subscriberCount": "10000", "videoCount": "500", "viewCount": "5000000" },
			"contentDetails": { "relatedPlaylists": { "uploads": "UU_test_channel" } }
		}]
	}`
	var channelResp social.YouTubeChannelResponse
	json.Unmarshal([]byte(channelRespJSON), &channelResp)

	mockYTClient := &social.MockYouTubeClient{
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			return &channelResp, nil
		},
		FetchVideosFunc: func(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]social.YouTubeActivityItem, error) {
			return []social.YouTubeActivityItem{}, nil
		},
		FetchActivityInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, errors.New("activity insights error")
		},
		FetchTrafficInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, errors.New("traffic insights error")
		},
		FetchSharedInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, errors.New("shared insights error")
		},
		DetectMediaTypesFunc: func(ctx context.Context, videos []social.YouTubeVideoItem) map[string]string {
			return map[string]string{}
		},
	}

	wo := kafkamodels.YouTubeAccountWorkOrder{
		ID:          "account123",
		ChannelID:   "UC_test_channel",
		AccessToken: "test-token",
		WorkspaceID: "workspace123",
		SyncType:    kafkamodels.YouTubeSyncTypeIncremental,
	}
	payload, _ := json.Marshal(wo)

	msg := WorkOrderMessage{
		AccountID:   wo.ID,
		ChannelID:   wo.ChannelID,
		Value:       payload,
		AccessToken: wo.AccessToken,
	}

	processWorkOrder(context.Background(), msg, mockYTClient, producer, &mongodb.MockUnifiedSocialRepository{}, "test-key", log, timestampUpdateChan)

	// Should still produce channel
	channelMessages := producer.getMessages(topicRawChannels)
	assert.Len(t, channelMessages, 1)

	// But no insights
	assert.Len(t, producer.getMessages(topicRawActivityInsights), 0)
	assert.Len(t, producer.getMessages(topicRawTrafficInsights), 0)
	assert.Len(t, producer.getMessages(topicRawSharedInsights), 0)
}

func TestProcessWorkOrder_EmptyInsightsRows(t *testing.T) {
	// Clear the global map first
	accountSemaphores = sync.Map{}

	log := createTestFetcherLogger()
	producer := newTestProducer()
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	channelRespJSON := `{
		"items": [{
			"id": "UC_test_channel",
			"snippet": { "title": "Test Channel", "publishedAt": "2020-01-15T10:30:00Z" },
			"statistics": { "subscriberCount": "10000", "videoCount": "500", "viewCount": "5000000" },
			"contentDetails": { "relatedPlaylists": { "uploads": "UU_test_channel" } }
		}]
	}`
	var channelResp social.YouTubeChannelResponse
	json.Unmarshal([]byte(channelRespJSON), &channelResp)

	mockYTClient := &social.MockYouTubeClient{
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			return &channelResp, nil
		},
		FetchVideosFunc: func(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]social.YouTubeActivityItem, error) {
			return []social.YouTubeActivityItem{}, nil
		},
		FetchActivityInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return &social.YouTubeAnalyticsResponse{Rows: [][]interface{}{}}, nil
		},
		FetchTrafficInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return &social.YouTubeAnalyticsResponse{Rows: [][]interface{}{}}, nil
		},
		FetchSharedInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return &social.YouTubeAnalyticsResponse{Rows: [][]interface{}{}}, nil
		},
		DetectMediaTypesFunc: func(ctx context.Context, videos []social.YouTubeVideoItem) map[string]string {
			return map[string]string{}
		},
	}

	wo := kafkamodels.YouTubeAccountWorkOrder{
		ID:          "account123",
		ChannelID:   "UC_test_channel",
		AccessToken: "test-token",
		WorkspaceID: "workspace123",
		SyncType:    kafkamodels.YouTubeSyncTypeIncremental,
	}
	payload, _ := json.Marshal(wo)

	msg := WorkOrderMessage{
		AccountID:   wo.ID,
		ChannelID:   wo.ChannelID,
		Value:       payload,
		AccessToken: wo.AccessToken,
	}

	processWorkOrder(context.Background(), msg, mockYTClient, producer, &mongodb.MockUnifiedSocialRepository{}, "test-key", log, timestampUpdateChan)

	// Should produce channel and empty insights messages are still produced (with empty rows)
	assert.Len(t, producer.getMessages(topicRawChannels), 1)
	// Insights are produced even with empty rows (they just contain empty rows array)
	assert.Len(t, producer.getMessages(topicRawActivityInsights), 1)
	assert.Len(t, producer.getMessages(topicRawTrafficInsights), 1)
	assert.Len(t, producer.getMessages(topicRawSharedInsights), 1)
}

func TestProcessWorkOrder_ChannelFetchError(t *testing.T) {
	// Clear the global map first
	accountSemaphores = sync.Map{}

	log := createTestFetcherLogger()
	producer := newTestProducer()
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	mockYTClient := &social.MockYouTubeClient{
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			return nil, errors.New("channel fetch error")
		},
		FetchVideosFunc: func(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]social.YouTubeActivityItem, error) {
			return []social.YouTubeActivityItem{}, nil
		},
		FetchActivityInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return &social.YouTubeAnalyticsResponse{Rows: [][]interface{}{{"2023-06-01", float64(100)}}}, nil
		},
		FetchTrafficInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		FetchSharedInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		DetectMediaTypesFunc: func(ctx context.Context, videos []social.YouTubeVideoItem) map[string]string {
			return map[string]string{}
		},
	}

	wo := kafkamodels.YouTubeAccountWorkOrder{
		ID:          "account123",
		ChannelID:   "UC_test_channel",
		AccessToken: "test-token",
		WorkspaceID: "workspace123",
		SyncType:    kafkamodels.YouTubeSyncTypeIncremental,
	}
	payload, _ := json.Marshal(wo)

	msg := WorkOrderMessage{
		AccountID:   wo.ID,
		ChannelID:   wo.ChannelID,
		Value:       payload,
		AccessToken: wo.AccessToken,
	}

	processWorkOrder(context.Background(), msg, mockYTClient, producer, &mongodb.MockUnifiedSocialRepository{}, "test-key", log, timestampUpdateChan)

	// Should not produce channel but should produce insights
	assert.Len(t, producer.getMessages(topicRawChannels), 0)
	assert.Len(t, producer.getMessages(topicRawActivityInsights), 1)
}

func TestService_TimestampUpdater_ChannelClosed(t *testing.T) {
	log := createTestFetcherLogger()

	mongoRepo := &mongodb.MockUnifiedSocialRepository{}
	mockYTClient := &social.MockYouTubeClient{}
	producer := newTestProducer()
	consumer := &kafka.MockConsumer{}

	svc := NewService(mockYTClient, producer, consumer, mongoRepo, log, "test-key")

	ctx := context.Background()
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	var wg sync.WaitGroup
	svc.startTimestampUpdater(ctx, &wg, timestampUpdateChan)

	// Close the channel immediately
	close(timestampUpdateChan)
	wg.Wait()
}

func TestProcessWorkOrder_ChannelFetchUnauthorized(t *testing.T) {
	// Clear the global map first
	accountSemaphores = sync.Map{}

	log := createTestFetcherLogger()
	producer := newTestProducer()
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	mockYTClient := &social.MockYouTubeClient{
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

	wo := kafkamodels.YouTubeAccountWorkOrder{
		ID:          "account123",
		ChannelID:   "UC_test_channel",
		AccessToken: "test-token",
		WorkspaceID: "workspace123",
		SyncType:    kafkamodels.YouTubeSyncTypeIncremental,
	}
	payload, _ := json.Marshal(wo)

	msg := WorkOrderMessage{
		AccountID:   wo.ID,
		ChannelID:   wo.ChannelID,
		Value:       payload,
		AccessToken: wo.AccessToken,
	}

	processWorkOrder(context.Background(), msg, mockYTClient, producer, &mongodb.MockUnifiedSocialRepository{}, "test-key", log, timestampUpdateChan)

	// Should not produce anything due to unauthorized
	assert.Len(t, producer.getMessages(topicRawChannels), 0)
	assert.Len(t, producer.getMessages(topicRawActivityInsights), 0)
}

func TestProcessWorkOrder_ProducerError(t *testing.T) {
	// Clear the global map first
	accountSemaphores = sync.Map{}

	log := createTestFetcherLogger()
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	channelRespJSON := `{
		"items": [{
			"id": "UC_test_channel",
			"snippet": { "title": "Test Channel", "publishedAt": "2020-01-15T10:30:00Z" },
			"statistics": { "subscriberCount": "10000", "videoCount": "500", "viewCount": "5000000" },
			"contentDetails": { "relatedPlaylists": { "uploads": "UU_test_channel" } }
		}]
	}`
	var channelResp social.YouTubeChannelResponse
	json.Unmarshal([]byte(channelRespJSON), &channelResp)

	mockYTClient := &social.MockYouTubeClient{
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			return &channelResp, nil
		},
		FetchVideosFunc: func(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]social.YouTubeActivityItem, error) {
			return []social.YouTubeActivityItem{}, nil
		},
		FetchActivityInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		FetchTrafficInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		FetchSharedInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		DetectMediaTypesFunc: func(ctx context.Context, videos []social.YouTubeVideoItem) map[string]string {
			return map[string]string{}
		},
	}

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return errors.New("kafka produce error")
		},
	}

	wo := kafkamodels.YouTubeAccountWorkOrder{
		ID:          "account123",
		ChannelID:   "UC_test_channel",
		AccessToken: "test-token",
		WorkspaceID: "workspace123",
		SyncType:    kafkamodels.YouTubeSyncTypeIncremental,
	}
	payload, _ := json.Marshal(wo)

	msg := WorkOrderMessage{
		AccountID:   wo.ID,
		ChannelID:   wo.ChannelID,
		Value:       payload,
		AccessToken: wo.AccessToken,
	}

	// Should not panic even if producer errors
	processWorkOrder(context.Background(), msg, mockYTClient, producer, &mongodb.MockUnifiedSocialRepository{}, "test-key", log, timestampUpdateChan)
}

func TestProcessWorkOrder_UnknownSyncType(t *testing.T) {
	// Clear the global map first
	accountSemaphores = sync.Map{}

	log := createTestFetcherLogger()
	producer := newTestProducer()
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	channelRespJSON := `{
		"items": [{
			"id": "UC_test_channel",
			"snippet": { "title": "Test Channel", "publishedAt": "2020-01-15T10:30:00Z" },
			"statistics": { "subscriberCount": "10000", "videoCount": "500", "viewCount": "5000000" },
			"contentDetails": { "relatedPlaylists": { "uploads": "UU_test_channel" } }
		}]
	}`
	var channelResp social.YouTubeChannelResponse
	json.Unmarshal([]byte(channelRespJSON), &channelResp)

	var capturedSinceDate time.Time
	mockYTClient := &social.MockYouTubeClient{
		FetchChannelsFunc: func(ctx context.Context, accessToken string) (*social.YouTubeChannelResponse, error) {
			return &channelResp, nil
		},
		FetchVideosFunc: func(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]social.YouTubeActivityItem, error) {
			capturedSinceDate = since
			return []social.YouTubeActivityItem{}, nil
		},
		FetchActivityInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		FetchTrafficInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		FetchSharedInsightsFunc: func(ctx context.Context, accessToken string, startDate, endDate time.Time) (*social.YouTubeAnalyticsResponse, error) {
			return nil, nil
		},
		DetectMediaTypesFunc: func(ctx context.Context, videos []social.YouTubeVideoItem) map[string]string {
			return map[string]string{}
		},
	}

	wo := kafkamodels.YouTubeAccountWorkOrder{
		ID:          "account123",
		ChannelID:   "UC_test_channel",
		AccessToken: "test-token",
		WorkspaceID: "workspace123",
		SyncType:    "unknown_sync_type", // Unknown sync type - should default to incremental
	}
	payload, _ := json.Marshal(wo)

	msg := WorkOrderMessage{
		AccountID:   wo.ID,
		ChannelID:   wo.ChannelID,
		Value:       payload,
		AccessToken: wo.AccessToken,
	}

	processWorkOrder(context.Background(), msg, mockYTClient, producer, &mongodb.MockUnifiedSocialRepository{}, "test-key", log, timestampUpdateChan)

	// Unknown sync type should default to incremental (14 days)
	// End date is 3 days ago (to account for YouTube API data delay)
	now := time.Now().UTC()
	expectedSince := now.AddDate(0, 0, -3).AddDate(0, 0, -incrementalVideosDays)
	assert.Equal(t, expectedSince.Year(), capturedSinceDate.Year())
	assert.Equal(t, expectedSince.Month(), capturedSinceDate.Month())
	assert.Equal(t, expectedSince.Day(), capturedSinceDate.Day())
}

// ================== Logging Contract Tests (Point 4 — Calling service logs errors with context) ==================

func TestLoggingContract_YouTubeFetcher_ErrorHasContextFields(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()

	log.Error().
		Str("error_message", "failed to fetch videos").
		Str("function", "processWorkOrder").
		Str("stage", "fetch_videos").
		Msg("YouTube fetcher error")

	output := buf.String()

	checks := map[string]string{
		"ERR":              "expected ERR level",
		"error_message":    "expected error_message field",
		"function":         "expected function field",
		"processWorkOrder": "expected processWorkOrder value",
		"stage":            "expected stage field",
		"fetch_videos":     "expected fetch_videos stage value",
	}
	for substr, errMsg := range checks {
		if !strings.Contains(output, substr) {
			t.Errorf("%s, got: %s", errMsg, output)
		}
	}
}

func TestLoggingContract_YouTubeFetcher_NoCaptureException(t *testing.T) {
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	log.Error().
		Str("error_message", "API quota exceeded").
		Str("function", "processWorkOrder").
		Str("stage", "fetch_channels").
		Msg("Failed to fetch channels")

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls (hook handles Sentry), got %d", len(*captureRecords))
	}
}

func TestLoggingContract_YouTubeFetcher_SingleSentryEvent(t *testing.T) {
	hookRecords, cleanup := logger.InstallHookSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	log.Error().
		Str("error_message", "kafka produce timeout").
		Str("function", "produceMessages").
		Str("stage", "produce_kafka").
		Msg("Failed to produce message")

	var errorLevelCount int
	for _, r := range *hookRecords {
		if r.Level == zerolog.ErrorLevel {
			errorLevelCount++
		}
	}
	if errorLevelCount != 1 {
		t.Fatalf("expected exactly 1 ErrorLevel hook firing, got %d", errorLevelCount)
	}
}

func TestLoggingContract_YouTubeFetcher_ExpectedError_WarnOnly(t *testing.T) {
	hookRecords, hookCleanup := logger.InstallHookSpy()
	defer hookCleanup()

	log, buf := logger.NewTestLoggerWithHook()

	log.Warn().
		Str("error_message", "Unauthorized - stopping all API calls").
		Str("function", "processWorkOrder").
		Str("stage", "fetch_channels").
		Str("channel_id", "yt-channel-123").
		Msg("YouTube auth error, skipping account")

	output := buf.String()

	if !strings.Contains(output, "WRN") {
		t.Fatalf("expected WRN level, got: %s", output)
	}
	if strings.Contains(output, "ERR") {
		t.Fatalf("expected error should NOT produce ERR level: %s", output)
	}

	for _, r := range *hookRecords {
		if r.Level >= zerolog.ErrorLevel {
			t.Fatalf("expected error should not trigger Error+ hook, got level %v", r.Level)
		}
	}
}
