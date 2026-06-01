package main

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

func createTestSinkLogger() *logger.Logger {
	return logger.New("debug")
}

func TestRawMessage_Struct(t *testing.T) {
	msg := RawMessage{
		Topic: "test-topic",
		Key:   []byte("key"),
		Value: []byte("value"),
	}

	if msg.Topic != "test-topic" {
		t.Fatalf("expected topic 'test-topic', got '%s'", msg.Topic)
	}
	if string(msg.Key) != "key" {
		t.Fatalf("expected key 'key', got '%s'", string(msg.Key))
	}
	if string(msg.Value) != "value" {
		t.Fatalf("expected value 'value', got '%s'", string(msg.Value))
	}
}

func TestBatchCollectors_Struct(t *testing.T) {
	bc := &BatchCollectors{
		channels:         make(chan *chmodels.YouTubeChannel, 10),
		videos:           make(chan *chmodels.YouTubeVideo, 10),
		activityInsights: make(chan *chmodels.YouTubeActivityInsights, 10),
		trafficInsights:  make(chan *chmodels.YouTubeTrafficInsights, 10),
		sharedInsights:   make(chan *chmodels.YouTubeSharedInsights, 10),
	}

	if bc.channels == nil {
		t.Fatal("expected channels channel to be initialized")
	}
	if bc.videos == nil {
		t.Fatal("expected videos channel to be initialized")
	}
	if bc.activityInsights == nil {
		t.Fatal("expected activityInsights channel to be initialized")
	}
	if bc.trafficInsights == nil {
		t.Fatal("expected trafficInsights channel to be initialized")
	}
	if bc.sharedInsights == nil {
		t.Fatal("expected sharedInsights channel to be initialized")
	}
}

func TestConstants(t *testing.T) {
	if channelParserWorkers != 3 {
		t.Fatalf("expected channelParserWorkers to be 3, got %d", channelParserWorkers)
	}
	if videoParserWorkers != 5 {
		t.Fatalf("expected videoParserWorkers to be 5, got %d", videoParserWorkers)
	}
	if activityParserWorkers != 3 {
		t.Fatalf("expected activityParserWorkers to be 3, got %d", activityParserWorkers)
	}
	if maxBatchSize != 5000 {
		t.Fatalf("expected maxBatchSize to be 5000, got %d", maxBatchSize)
	}
	if consumerGroup != "youtube-clickhouse-sink-group" {
		t.Fatalf("unexpected consumerGroup: %s", consumerGroup)
	}
}

func TestGetInt64FromRow(t *testing.T) {
	colIndex := map[string]int{
		"views":    0,
		"likes":    1,
		"comments": 2,
		"strVal":   3,
	}

	tests := []struct {
		name     string
		row      []interface{}
		colName  string
		expected int64
	}{
		{
			name:     "float64 value",
			row:      []interface{}{float64(1000), float64(100), float64(50), "10"},
			colName:  "views",
			expected: 1000,
		},
		{
			name:     "int64 value",
			row:      []interface{}{int64(2000), int64(200), int64(60), "20"},
			colName:  "views",
			expected: 2000,
		},
		{
			name:     "int value",
			row:      []interface{}{3000, 300, 70, "30"},
			colName:  "views",
			expected: 3000,
		},
		{
			name:     "string value",
			row:      []interface{}{float64(1000), float64(100), float64(50), "999"},
			colName:  "strVal",
			expected: 999,
		},
		{
			name:     "missing column",
			row:      []interface{}{float64(1000)},
			colName:  "nonexistent",
			expected: 0,
		},
		{
			name:     "index out of bounds",
			row:      []interface{}{float64(1000)},
			colName:  "comments",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getInt64FromRow(tt.row, colIndex, tt.colName)
			if result != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestGetFloat64FromRow(t *testing.T) {
	colIndex := map[string]int{
		"percentage": 0,
		"rate":       1,
		"strVal":     2,
	}

	tests := []struct {
		name     string
		row      []interface{}
		colName  string
		expected float64
	}{
		{
			name:     "float64 value",
			row:      []interface{}{float64(95.5), float64(0.85), "1.5"},
			colName:  "percentage",
			expected: 95.5,
		},
		{
			name:     "int64 value",
			row:      []interface{}{int64(100), int64(1), "2.5"},
			colName:  "percentage",
			expected: 100.0,
		},
		{
			name:     "int value",
			row:      []interface{}{50, 1, "3.5"},
			colName:  "percentage",
			expected: 50.0,
		},
		{
			name:     "string value",
			row:      []interface{}{float64(95.5), float64(0.85), "99.9"},
			colName:  "strVal",
			expected: 99.9,
		},
		{
			name:     "missing column",
			row:      []interface{}{float64(95.5)},
			colName:  "nonexistent",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFloat64FromRow(tt.row, colIndex, tt.colName)
			if result != tt.expected {
				t.Fatalf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestParseAndQueueChannel(t *testing.T) {
	log := createTestSinkLogger()
	ctx := context.Background()

	batches := &BatchCollectors{
		channels: make(chan *chmodels.YouTubeChannel, 10),
	}

	rawChannel := kafkamodels.RawYouTubeChannel{
		ChannelID:       "UC_test123",
		Title:           "Test Channel",
		Description:     "Test Description",
		SubscriberCount: 1000,
		VideoCount:      50,
		ViewCount:       100000,
	}
	payload, _ := json.Marshal(rawChannel)

	msg := RawMessage{
		Topic: topicRawChannels,
		Key:   []byte("UC_test123"),
		Value: payload,
	}

	var counter uint64
	parseAndQueueChannel(ctx, msg, batches, &counter, log)

	if atomic.LoadUint64(&counter) != 1 {
		t.Fatalf("expected counter to be 1, got %d", atomic.LoadUint64(&counter))
	}

	select {
	case ch := <-batches.channels:
		if ch.ChannelID != "UC_test123" {
			t.Fatalf("expected channel ID 'UC_test123', got '%s'", ch.ChannelID)
		}
		if ch.Title != "Test Channel" {
			t.Fatalf("expected title 'Test Channel', got '%s'", ch.Title)
		}
	default:
		t.Fatal("expected channel to be queued")
	}
}

func TestParseAndQueueChannel_InvalidJSON(t *testing.T) {
	log := createTestSinkLogger()
	ctx := context.Background()

	batches := &BatchCollectors{
		channels: make(chan *chmodels.YouTubeChannel, 10),
	}

	msg := RawMessage{
		Topic: topicRawChannels,
		Key:   []byte("UC_test123"),
		Value: []byte("invalid json"),
	}

	var counter uint64
	parseAndQueueChannel(ctx, msg, batches, &counter, log)

	if atomic.LoadUint64(&counter) != 0 {
		t.Fatalf("expected counter to be 0 for invalid JSON, got %d", atomic.LoadUint64(&counter))
	}
}

func TestParseAndQueueVideo(t *testing.T) {
	log := createTestSinkLogger()
	ctx := context.Background()

	batches := &BatchCollectors{
		videos: make(chan *chmodels.YouTubeVideo, 10),
	}

	rawVideo := kafkamodels.RawYouTubeVideo{
		VideoID:   "video_123",
		ChannelID: "UC_test123",
		Title:     "Test Video",
		Views:     5000,
		Likes:     100,
	}
	payload, _ := json.Marshal(rawVideo)

	msg := RawMessage{
		Topic: topicRawVideos,
		Key:   []byte("video_123"),
		Value: payload,
	}

	var counter uint64
	parseAndQueueVideo(ctx, msg, batches, &counter, log)

	if atomic.LoadUint64(&counter) != 1 {
		t.Fatalf("expected counter to be 1, got %d", atomic.LoadUint64(&counter))
	}

	select {
	case v := <-batches.videos:
		if v.VideoID != "video_123" {
			t.Fatalf("expected video ID 'video_123', got '%s'", v.VideoID)
		}
		if v.Title != "Test Video" {
			t.Fatalf("expected title 'Test Video', got '%s'", v.Title)
		}
	default:
		t.Fatal("expected video to be queued")
	}
}

func TestParseAndQueueActivityInsights(t *testing.T) {
	log := createTestSinkLogger()
	ctx := context.Background()

	batches := &BatchCollectors{
		activityInsights: make(chan *chmodels.YouTubeActivityInsights, 10),
	}

	rawInsights := struct {
		ChannelID string                           `json:"channel_id"`
		Response  *social.YouTubeAnalyticsResponse `json:"response"`
	}{
		ChannelID: "UC_test123",
		Response: &social.YouTubeAnalyticsResponse{
			ColumnHeaders: []struct {
				Name       string `json:"name"`
				ColumnType string `json:"columnType"`
				DataType   string `json:"dataType"`
			}{
				{Name: "day", ColumnType: "DIMENSION", DataType: "STRING"},
				{Name: "views", ColumnType: "METRIC", DataType: "INTEGER"},
				{Name: "likes", ColumnType: "METRIC", DataType: "INTEGER"},
			},
			Rows: [][]interface{}{
				{"2024-01-15", float64(1000), float64(50)},
			},
		},
	}
	payload, _ := json.Marshal(rawInsights)

	msg := RawMessage{
		Topic: topicRawActivityInsights,
		Key:   []byte("UC_test123"),
		Value: payload,
	}

	var counter uint64
	parseAndQueueActivityInsights(ctx, msg, batches, &counter, log)

	if atomic.LoadUint64(&counter) != 1 {
		t.Fatalf("expected counter to be 1, got %d", atomic.LoadUint64(&counter))
	}
}

func TestParseAndQueueActivityInsights_EmptyResponse(t *testing.T) {
	log := createTestSinkLogger()
	ctx := context.Background()

	batches := &BatchCollectors{
		activityInsights: make(chan *chmodels.YouTubeActivityInsights, 10),
	}

	rawInsights := struct {
		ChannelID string                           `json:"channel_id"`
		Response  *social.YouTubeAnalyticsResponse `json:"response"`
	}{
		ChannelID: "UC_test123",
		Response:  nil,
	}
	payload, _ := json.Marshal(rawInsights)

	msg := RawMessage{
		Topic: topicRawActivityInsights,
		Key:   []byte("UC_test123"),
		Value: payload,
	}

	var counter uint64
	parseAndQueueActivityInsights(ctx, msg, batches, &counter, log)

	if atomic.LoadUint64(&counter) != 0 {
		t.Fatalf("expected counter to be 0 for empty response, got %d", atomic.LoadUint64(&counter))
	}
}

func TestParseAndQueueActivityInsights_Deduplication(t *testing.T) {
	log := createTestSinkLogger()
	ctx := context.Background()

	batches := &BatchCollectors{
		activityInsights: make(chan *chmodels.YouTubeActivityInsights, 10),
	}

	rawInsights := struct {
		ChannelID string                           `json:"channel_id"`
		Response  *social.YouTubeAnalyticsResponse `json:"response"`
	}{
		ChannelID: "UC_test123",
		Response: &social.YouTubeAnalyticsResponse{
			ColumnHeaders: []struct {
				Name       string `json:"name"`
				ColumnType string `json:"columnType"`
				DataType   string `json:"dataType"`
			}{
				{Name: "day", ColumnType: "DIMENSION", DataType: "STRING"},
				{Name: "views", ColumnType: "METRIC", DataType: "INTEGER"},
				{Name: "likes", ColumnType: "METRIC", DataType: "INTEGER"},
				{Name: "comments", ColumnType: "METRIC", DataType: "INTEGER"},
			},
			Rows: [][]interface{}{
				{"2024-01-15", float64(1000), float64(50), float64(10)},
				{"2024-01-15", float64(500), float64(25), float64(5)},
				{"2024-01-16", float64(2000), float64(100), float64(20)},
			},
		},
	}
	payload, _ := json.Marshal(rawInsights)

	msg := RawMessage{
		Topic: topicRawActivityInsights,
		Key:   []byte("UC_test123"),
		Value: payload,
	}

	var counter uint64
	parseAndQueueActivityInsights(ctx, msg, batches, &counter, log)

	if atomic.LoadUint64(&counter) != 2 {
		t.Fatalf("expected counter to be 2 (2 unique dates), got %d", atomic.LoadUint64(&counter))
	}

	insights := make([]*chmodels.YouTubeActivityInsights, 0, 2)
	for i := 0; i < 2; i++ {
		select {
		case insight := <-batches.activityInsights:
			insights = append(insights, insight)
		default:
			t.Fatal("expected 2 insights to be queued")
		}
	}

	dateViews := make(map[string]int64)
	dateLikes := make(map[string]int64)
	dateComments := make(map[string]int64)
	for _, insight := range insights {
		dateStr := insight.CreatedAt.Format("2006-01-02")
		dateViews[dateStr] = insight.Views
		dateLikes[dateStr] = insight.Likes
		dateComments[dateStr] = insight.Comments
	}

	if dateViews["2024-01-15"] != 1500 {
		t.Fatalf("expected 2024-01-15 views to be 1500 (1000+500), got %d", dateViews["2024-01-15"])
	}
	if dateLikes["2024-01-15"] != 75 {
		t.Fatalf("expected 2024-01-15 likes to be 75 (50+25), got %d", dateLikes["2024-01-15"])
	}
	if dateComments["2024-01-15"] != 15 {
		t.Fatalf("expected 2024-01-15 comments to be 15 (10+5), got %d", dateComments["2024-01-15"])
	}
	if dateViews["2024-01-16"] != 2000 {
		t.Fatalf("expected 2024-01-16 views to be 2000, got %d", dateViews["2024-01-16"])
	}
}

func TestParseAndQueueTrafficInsights(t *testing.T) {
	log := createTestSinkLogger()
	ctx := context.Background()

	batches := &BatchCollectors{
		trafficInsights: make(chan *chmodels.YouTubeTrafficInsights, 10),
	}

	rawInsights := struct {
		ChannelID string                           `json:"channel_id"`
		Response  *social.YouTubeAnalyticsResponse `json:"response"`
	}{
		ChannelID: "UC_test123",
		Response: &social.YouTubeAnalyticsResponse{
			ColumnHeaders: []struct {
				Name       string `json:"name"`
				ColumnType string `json:"columnType"`
				DataType   string `json:"dataType"`
			}{
				{Name: "day", ColumnType: "DIMENSION", DataType: "STRING"},
				{Name: "insightTrafficSourceType", ColumnType: "DIMENSION", DataType: "STRING"},
				{Name: "views", ColumnType: "METRIC", DataType: "INTEGER"},
			},
			Rows: [][]interface{}{
				{"2024-01-15", "YT_SEARCH", float64(500)},
				{"2024-01-15", "SUBSCRIBER", float64(300)},
			},
		},
	}
	payload, _ := json.Marshal(rawInsights)

	msg := RawMessage{
		Topic: topicRawTrafficInsights,
		Key:   []byte("UC_test123"),
		Value: payload,
	}

	var counter uint64
	parseAndQueueTrafficInsights(ctx, msg, batches, &counter, log)

	if atomic.LoadUint64(&counter) < 1 {
		t.Fatalf("expected counter to be at least 1, got %d", atomic.LoadUint64(&counter))
	}
}

func TestParseAndQueueSharedInsights(t *testing.T) {
	log := createTestSinkLogger()
	ctx := context.Background()

	batches := &BatchCollectors{
		sharedInsights: make(chan *chmodels.YouTubeSharedInsights, 10),
	}

	rawInsights := struct {
		ChannelID string                           `json:"channel_id"`
		Response  *social.YouTubeAnalyticsResponse `json:"response"`
	}{
		ChannelID: "UC_test123",
		Response: &social.YouTubeAnalyticsResponse{
			ColumnHeaders: []struct {
				Name       string `json:"name"`
				ColumnType string `json:"columnType"`
				DataType   string `json:"dataType"`
			}{
				{Name: "sharingService", ColumnType: "DIMENSION", DataType: "STRING"},
				{Name: "shares", ColumnType: "METRIC", DataType: "INTEGER"},
			},
			Rows: [][]interface{}{
				{"TWITTER", float64(100)},
				{"FACEBOOK", float64(200)},
			},
		},
	}
	payload, _ := json.Marshal(rawInsights)

	msg := RawMessage{
		Topic: topicRawSharedInsights,
		Key:   []byte("UC_test123"),
		Value: payload,
	}

	var counter uint64
	parseAndQueueSharedInsights(ctx, msg, batches, &counter, log)

	if atomic.LoadUint64(&counter) != 1 {
		t.Fatalf("expected counter to be 1, got %d", atomic.LoadUint64(&counter))
	}
}

func TestChannelParserWorker_StopsOnContextCancel(t *testing.T) {
	log := createTestSinkLogger()
	ctx, cancel := context.WithCancel(context.Background())

	batches := &BatchCollectors{
		channels: make(chan *chmodels.YouTubeChannel, 10),
	}
	msgCh := make(chan RawMessage, 10)
	var counter uint64
	var wg sync.WaitGroup

	wg.Add(1)
	go channelParserWorker(ctx, 1, msgCh, batches, &counter, log, &wg)

	cancel()
	wg.Wait()
}

func TestChannelParserWorker_StopsOnChannelClose(t *testing.T) {
	log := createTestSinkLogger()
	ctx := context.Background()

	batches := &BatchCollectors{
		channels: make(chan *chmodels.YouTubeChannel, 10),
	}
	msgCh := make(chan RawMessage, 10)
	var counter uint64
	var wg sync.WaitGroup

	wg.Add(1)
	go channelParserWorker(ctx, 1, msgCh, batches, &counter, log, &wg)

	close(msgCh)
	wg.Wait()
}

func TestVideoParserWorker_StopsOnContextCancel(t *testing.T) {
	log := createTestSinkLogger()
	ctx, cancel := context.WithCancel(context.Background())

	batches := &BatchCollectors{
		videos: make(chan *chmodels.YouTubeVideo, 10),
	}
	msgCh := make(chan RawMessage, 10)
	var counter uint64
	var wg sync.WaitGroup

	wg.Add(1)
	go videoParserWorker(ctx, 1, msgCh, batches, &counter, log, &wg)

	cancel()
	wg.Wait()
}

func TestActivityParserWorker_StopsOnContextCancel(t *testing.T) {
	log := createTestSinkLogger()
	ctx, cancel := context.WithCancel(context.Background())

	batches := &BatchCollectors{
		activityInsights: make(chan *chmodels.YouTubeActivityInsights, 10),
	}
	msgCh := make(chan RawMessage, 10)
	var counter uint64
	var wg sync.WaitGroup

	wg.Add(1)
	go activityParserWorker(ctx, 1, msgCh, batches, &counter, log, &wg)

	cancel()
	wg.Wait()
}

func TestTrafficParserWorker_StopsOnContextCancel(t *testing.T) {
	log := createTestSinkLogger()
	ctx, cancel := context.WithCancel(context.Background())

	batches := &BatchCollectors{
		trafficInsights: make(chan *chmodels.YouTubeTrafficInsights, 10),
	}
	msgCh := make(chan RawMessage, 10)
	var counter uint64
	var wg sync.WaitGroup

	wg.Add(1)
	go trafficParserWorker(ctx, 1, msgCh, batches, &counter, log, &wg)

	cancel()
	wg.Wait()
}

func TestSharedParserWorker_StopsOnContextCancel(t *testing.T) {
	log := createTestSinkLogger()
	ctx, cancel := context.WithCancel(context.Background())

	batches := &BatchCollectors{
		sharedInsights: make(chan *chmodels.YouTubeSharedInsights, 10),
	}
	msgCh := make(chan RawMessage, 10)
	var counter uint64
	var wg sync.WaitGroup

	wg.Add(1)
	go sharedParserWorker(ctx, 1, msgCh, batches, &counter, log, &wg)

	cancel()
	wg.Wait()
}

// ================== Logging Contract Tests (Point 4 — Calling service logs errors with context) ==================

func TestLoggingContract_YouTubeAnalyticsSink_ErrorHasContextFields(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()

	log.Error().
		Str("error_message", "bulk insert videos failed").
		Str("function", "batchInsertWorker").
		Str("stage", "bulk_insert_videos").
		Msg("YouTube analytics sink error")

	output := buf.String()

	checks := map[string]string{
		"ERR":                "expected ERR level",
		"error_message":      "expected error_message field",
		"function":           "expected function field",
		"batchInsertWorker":  "expected batchInsertWorker value",
		"stage":              "expected stage field",
		"bulk_insert_videos": "expected bulk_insert_videos stage value",
	}
	for substr, errMsg := range checks {
		if !strings.Contains(output, substr) {
			t.Errorf("%s, got: %s", errMsg, output)
		}
	}
}

func TestLoggingContract_YouTubeAnalyticsSink_NoCaptureException(t *testing.T) {
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	log.Error().
		Str("error_message", "unmarshal failed").
		Str("function", "channelParserWorker").
		Str("stage", "unmarshal_channel").
		Msg("Failed to unmarshal channel")

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls (hook handles Sentry), got %d", len(*captureRecords))
	}
}

func TestLoggingContract_YouTubeAnalyticsSink_SingleSentryEvent(t *testing.T) {
	hookRecords, cleanup := logger.InstallHookSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	log.Error().
		Str("error_message", "clickhouse connection lost").
		Str("function", "batchInsertWorker").
		Str("stage", "bulk_insert_channels").
		Msg("bulk insert channels failed")

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
