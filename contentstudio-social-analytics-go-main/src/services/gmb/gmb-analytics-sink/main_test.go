package main

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

func newTestLogger() *logger.Logger {
	log, _ := logger.NewTestLoggerWithHook()
	return log
}

// ================== Constants Tests ==================

func TestConstants(t *testing.T) {
	if maxBatchSize <= 0 {
		t.Fatalf("maxBatchSize must be positive, got %d", maxBatchSize)
	}
	if batchTimeout <= 0 {
		t.Fatalf("batchTimeout must be positive, got %v", batchTimeout)
	}
	if batchProcessorsPerType <= 0 {
		t.Fatalf("batchProcessorsPerType must be positive, got %d", batchProcessorsPerType)
	}
	if messageChanSize <= 0 {
		t.Fatalf("messageChanSize must be positive, got %d", messageChanSize)
	}
	if rawDataTopic == "" {
		t.Fatal("rawDataTopic must not be empty")
	}
	if consumerGroup == "" {
		t.Fatal("consumerGroup must not be empty")
	}
	if idleTimeout <= 0 {
		t.Fatalf("idleTimeout must be positive, got %v", idleTimeout)
	}
}

func TestTopicNames(t *testing.T) {
	if rawDataTopic != "raw-gmb-data" {
		t.Fatalf("expected rawDataTopic 'raw-gmb-data', got %s", rawDataTopic)
	}
	if consumerGroup != "gmb-analytics-sink-group" {
		t.Fatalf("expected consumerGroup 'gmb-analytics-sink-group', got %s", consumerGroup)
	}
}

// ================== Struct Tests ==================

func TestRawMessageStruct(t *testing.T) {
	msg := RawMessage{
		Topic: "test-topic",
		Key:   []byte("key"),
		Value: []byte(`{"data_type":"test"}`),
	}
	if msg.Topic != "test-topic" {
		t.Fatal("unexpected topic")
	}
}

func TestBatchCollectorsStruct(t *testing.T) {
	b := &BatchCollectors{
		dailyMetrics:   make(chan *chmodels.GMBDailyMetrics, 10),
		mediaAssets:    make(chan *chmodels.GMBMediaAssets, 10),
		searchKeywords: make(chan *chmodels.GMBSearchKeywordsMonthly, 10),
		localPosts:     make(chan *chmodels.GMBLocalPosts, 10),
		reviews:        make(chan *chmodels.GMBReviews, 10),
	}
	if cap(b.dailyMetrics) != 10 {
		t.Fatal("unexpected channel capacity")
	}
}

// ================== Data Parser Tests ==================

func TestDataParser_PerformanceMetrics(t *testing.T) {
	raw := kafkamodels.RawGMBData{
		AccountID:  "acc-1",
		LocationID: "loc-1",
		DataType:   "performance_metrics",
		Data: map[string]interface{}{
			"multiDailyMetricTimeSeries": []interface{}{
				map[string]interface{}{
					"dailyMetricTimeSeries": []interface{}{
						map[string]interface{}{
							"dailyMetric": "CALL_CLICKS",
							"timeSeries": map[string]interface{}{
								"datedValues": []interface{}{
									map[string]interface{}{
										"date":  map[string]interface{}{"year": float64(2024), "month": float64(1), "day": float64(15)},
										"value": "42",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	msgValue, _ := json.Marshal(raw)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	batches := &BatchCollectors{
		dailyMetrics:   make(chan *chmodels.GMBDailyMetrics, 100),
		mediaAssets:    make(chan *chmodels.GMBMediaAssets, 100),
		searchKeywords: make(chan *chmodels.GMBSearchKeywordsMonthly, 100),
		localPosts:     make(chan *chmodels.GMBLocalPosts, 100),
		reviews:        make(chan *chmodels.GMBReviews, 100),
	}

	jobs := make(chan RawMessage, 10)
	jobs <- RawMessage{Topic: rawDataTopic, Key: []byte("key"), Value: msgValue}
	close(jobs)

	var wg sync.WaitGroup
	var parsedDM, parsedMA, parsedSK, parsedLP, parsedRV uint64
	wg.Add(1)

	log := newTestLogger()
	go dataParser(ctx, &wg, 0, jobs, batches, log, &parsedDM, &parsedMA, &parsedSK, &parsedLP, &parsedRV)
	wg.Wait()

	if atomic.LoadUint64(&parsedDM) == 0 {
		t.Fatal("expected at least one daily metric to be parsed")
	}
}

func TestDataParser_Reviews(t *testing.T) {
	raw := kafkamodels.RawGMBData{
		AccountID:  "acc-1",
		LocationID: "loc-1",
		DataType:   "reviews",
		Data: map[string]interface{}{
			"reviews": []interface{}{
				map[string]interface{}{
					"name":       "accounts/123/locations/456/reviews/789",
					"reviewId":   "review-001",
					"starRating": "FIVE",
					"comment":    "Great place!",
					"createTime": "2024-01-15T10:00:00Z",
					"updateTime": "2024-01-15T10:00:00Z",
					"reviewer": map[string]interface{}{
						"displayName":     "John Doe",
						"profilePhotoUrl": "https://example.com/photo.jpg",
					},
				},
			},
		},
	}

	msgValue, _ := json.Marshal(raw)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	batches := &BatchCollectors{
		dailyMetrics:   make(chan *chmodels.GMBDailyMetrics, 100),
		mediaAssets:    make(chan *chmodels.GMBMediaAssets, 100),
		searchKeywords: make(chan *chmodels.GMBSearchKeywordsMonthly, 100),
		localPosts:     make(chan *chmodels.GMBLocalPosts, 100),
		reviews:        make(chan *chmodels.GMBReviews, 100),
	}

	jobs := make(chan RawMessage, 10)
	jobs <- RawMessage{Topic: rawDataTopic, Key: []byte("key"), Value: msgValue}
	close(jobs)

	var wg sync.WaitGroup
	var parsedDM, parsedMA, parsedSK, parsedLP, parsedRV uint64
	wg.Add(1)

	log := newTestLogger()
	go dataParser(ctx, &wg, 0, jobs, batches, log, &parsedDM, &parsedMA, &parsedSK, &parsedLP, &parsedRV)
	wg.Wait()

	if atomic.LoadUint64(&parsedRV) == 0 {
		t.Fatal("expected at least one review to be parsed")
	}
}

func TestDataParser_LocalPosts(t *testing.T) {
	raw := kafkamodels.RawGMBData{
		AccountID:  "acc-1",
		LocationID: "loc-1",
		DataType:   "local_posts",
		Data: map[string]interface{}{
			"localPosts": []interface{}{
				map[string]interface{}{
					"name":       "accounts/123/locations/456/localPosts/789",
					"state":      "LIVE",
					"topicType":  "STANDARD",
					"searchUrl":  "https://search.google.com/local/posts?q=test",
					"createTime": "2024-01-15T10:00:00Z",
					"updateTime": "2024-01-15T12:00:00Z",
					"media":      []interface{}{},
				},
			},
		},
	}

	msgValue, _ := json.Marshal(raw)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	batches := &BatchCollectors{
		dailyMetrics:   make(chan *chmodels.GMBDailyMetrics, 100),
		mediaAssets:    make(chan *chmodels.GMBMediaAssets, 100),
		searchKeywords: make(chan *chmodels.GMBSearchKeywordsMonthly, 100),
		localPosts:     make(chan *chmodels.GMBLocalPosts, 100),
		reviews:        make(chan *chmodels.GMBReviews, 100),
	}

	jobs := make(chan RawMessage, 10)
	jobs <- RawMessage{Topic: rawDataTopic, Key: []byte("key"), Value: msgValue}
	close(jobs)

	var wg sync.WaitGroup
	var parsedDM, parsedMA, parsedSK, parsedLP, parsedRV uint64
	wg.Add(1)

	log := newTestLogger()
	go dataParser(ctx, &wg, 0, jobs, batches, log, &parsedDM, &parsedMA, &parsedSK, &parsedLP, &parsedRV)
	wg.Wait()

	if atomic.LoadUint64(&parsedLP) == 0 {
		t.Fatal("expected at least one local post to be parsed")
	}
}

func TestDataParser_MediaAssets(t *testing.T) {
	raw := kafkamodels.RawGMBData{
		AccountID:  "acc-1",
		LocationID: "loc-1",
		DataType:   "media_assets",
		Data: map[string]interface{}{
			"mediaItems": []interface{}{
				map[string]interface{}{
					"name":        "accounts/123/locations/456/media/789",
					"mediaFormat": "PHOTO",
					"locationAssociation": map[string]interface{}{
						"category": "INTERIOR",
					},
					"googleUrl":    "https://lh3.googleusercontent.com/photo",
					"thumbnailUrl": "https://lh3.googleusercontent.com/thumb",
					"sourceUrl":    "https://example.com/photo.jpg",
					"dimensions": map[string]interface{}{
						"widthPixels":  float64(1920),
						"heightPixels": float64(1080),
					},
					"createTime": "2024-01-15T10:00:00Z",
				},
			},
		},
	}

	msgValue, _ := json.Marshal(raw)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	batches := &BatchCollectors{
		dailyMetrics:   make(chan *chmodels.GMBDailyMetrics, 100),
		mediaAssets:    make(chan *chmodels.GMBMediaAssets, 100),
		searchKeywords: make(chan *chmodels.GMBSearchKeywordsMonthly, 100),
		localPosts:     make(chan *chmodels.GMBLocalPosts, 100),
		reviews:        make(chan *chmodels.GMBReviews, 100),
	}

	jobs := make(chan RawMessage, 10)
	jobs <- RawMessage{Topic: rawDataTopic, Key: []byte("key"), Value: msgValue}
	close(jobs)

	var wg sync.WaitGroup
	var parsedDM, parsedMA, parsedSK, parsedLP, parsedRV uint64
	wg.Add(1)

	log := newTestLogger()
	go dataParser(ctx, &wg, 0, jobs, batches, log, &parsedDM, &parsedMA, &parsedSK, &parsedLP, &parsedRV)
	wg.Wait()

	if atomic.LoadUint64(&parsedMA) == 0 {
		t.Fatal("expected at least one media asset to be parsed")
	}
}

func TestDataParser_SearchKeywords(t *testing.T) {
	raw := kafkamodels.RawGMBData{
		AccountID:  "acc-1",
		LocationID: "loc-1",
		DataType:   "search_keywords",
		Data: map[string]interface{}{
			"searchKeywordsCounts": []interface{}{
				map[string]interface{}{
					"searchKeyword":     "pizza near me",
					"insideSearchCount": "42",
				},
			},
		},
	}

	msgValue, _ := json.Marshal(raw)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	batches := &BatchCollectors{
		dailyMetrics:   make(chan *chmodels.GMBDailyMetrics, 100),
		mediaAssets:    make(chan *chmodels.GMBMediaAssets, 100),
		searchKeywords: make(chan *chmodels.GMBSearchKeywordsMonthly, 100),
		localPosts:     make(chan *chmodels.GMBLocalPosts, 100),
		reviews:        make(chan *chmodels.GMBReviews, 100),
	}

	jobs := make(chan RawMessage, 10)
	jobs <- RawMessage{Topic: rawDataTopic, Key: []byte("key"), Value: msgValue}
	close(jobs)

	var wg sync.WaitGroup
	var parsedDM, parsedMA, parsedSK, parsedLP, parsedRV uint64
	wg.Add(1)

	log := newTestLogger()
	go dataParser(ctx, &wg, 0, jobs, batches, log, &parsedDM, &parsedMA, &parsedSK, &parsedLP, &parsedRV)
	wg.Wait()

	if atomic.LoadUint64(&parsedSK) == 0 {
		t.Fatal("expected at least one search keyword to be parsed")
	}
}

func TestDataParser_UnknownDataType(t *testing.T) {
	raw := kafkamodels.RawGMBData{
		AccountID:  "acc-1",
		LocationID: "loc-1",
		DataType:   "unknown_type",
		Data:       map[string]interface{}{},
	}

	msgValue, _ := json.Marshal(raw)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	batches := &BatchCollectors{
		dailyMetrics:   make(chan *chmodels.GMBDailyMetrics, 100),
		mediaAssets:    make(chan *chmodels.GMBMediaAssets, 100),
		searchKeywords: make(chan *chmodels.GMBSearchKeywordsMonthly, 100),
		localPosts:     make(chan *chmodels.GMBLocalPosts, 100),
		reviews:        make(chan *chmodels.GMBReviews, 100),
	}

	jobs := make(chan RawMessage, 10)
	jobs <- RawMessage{Topic: rawDataTopic, Key: []byte("key"), Value: msgValue}
	close(jobs)

	var wg sync.WaitGroup
	var parsedDM, parsedMA, parsedSK, parsedLP, parsedRV uint64
	wg.Add(1)

	log := newTestLogger()
	go dataParser(ctx, &wg, 0, jobs, batches, log, &parsedDM, &parsedMA, &parsedSK, &parsedLP, &parsedRV)
	wg.Wait()

	total := atomic.LoadUint64(&parsedDM) + atomic.LoadUint64(&parsedMA) +
		atomic.LoadUint64(&parsedSK) + atomic.LoadUint64(&parsedLP) + atomic.LoadUint64(&parsedRV)
	if total != 0 {
		t.Fatalf("expected 0 parsed items for unknown data type, got %d", total)
	}
}

func TestDataParser_InvalidJSON(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	batches := &BatchCollectors{
		dailyMetrics:   make(chan *chmodels.GMBDailyMetrics, 100),
		mediaAssets:    make(chan *chmodels.GMBMediaAssets, 100),
		searchKeywords: make(chan *chmodels.GMBSearchKeywordsMonthly, 100),
		localPosts:     make(chan *chmodels.GMBLocalPosts, 100),
		reviews:        make(chan *chmodels.GMBReviews, 100),
	}

	jobs := make(chan RawMessage, 10)
	jobs <- RawMessage{Topic: rawDataTopic, Key: []byte("key"), Value: []byte("invalid json")}
	close(jobs)

	var wg sync.WaitGroup
	var parsedDM, parsedMA, parsedSK, parsedLP, parsedRV uint64
	wg.Add(1)

	log := newTestLogger()
	go dataParser(ctx, &wg, 0, jobs, batches, log, &parsedDM, &parsedMA, &parsedSK, &parsedLP, &parsedRV)
	wg.Wait()

	total := atomic.LoadUint64(&parsedDM) + atomic.LoadUint64(&parsedMA) +
		atomic.LoadUint64(&parsedSK) + atomic.LoadUint64(&parsedLP) + atomic.LoadUint64(&parsedRV)
	if total != 0 {
		t.Fatalf("expected 0 parsed items for invalid JSON, got %d", total)
	}
}

func TestDataParser_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	batches := &BatchCollectors{
		dailyMetrics:   make(chan *chmodels.GMBDailyMetrics, 100),
		mediaAssets:    make(chan *chmodels.GMBMediaAssets, 100),
		searchKeywords: make(chan *chmodels.GMBSearchKeywordsMonthly, 100),
		localPosts:     make(chan *chmodels.GMBLocalPosts, 100),
		reviews:        make(chan *chmodels.GMBReviews, 100),
	}

	jobs := make(chan RawMessage, 10)

	var wg sync.WaitGroup
	var parsedDM, parsedMA, parsedSK, parsedLP, parsedRV uint64
	wg.Add(1)

	log := newTestLogger()
	go dataParser(ctx, &wg, 0, jobs, batches, log, &parsedDM, &parsedMA, &parsedSK, &parsedLP, &parsedRV)

	cancel()
	wg.Wait()
}

// ================== Batch Processor Tests ==================

func TestDailyMetricsBatchProcessor_FlushOnClose(t *testing.T) {
	ctx := context.Background()
	sink := NewMockClickHouseSink()

	ch := make(chan *chmodels.GMBDailyMetrics, 100)
	ch <- &chmodels.GMBDailyMetrics{AccountID: "acc-1"}
	ch <- &chmodels.GMBDailyMetrics{AccountID: "acc-2"}
	close(ch)

	var wg sync.WaitGroup
	var insertedCounter uint64
	wg.Add(1)

	log := newTestLogger()
	go dailyMetricsBatchProcessor(ctx, &wg, 0, ch, sink, log, &insertedCounter)
	wg.Wait()

	if atomic.LoadUint64(&insertedCounter) != 2 {
		t.Fatalf("expected 2 inserted, got %d", atomic.LoadUint64(&insertedCounter))
	}
}

func TestReviewsBatchProcessor_FlushOnClose(t *testing.T) {
	ctx := context.Background()
	sink := NewMockClickHouseSink()

	ch := make(chan *chmodels.GMBReviews, 100)
	ch <- &chmodels.GMBReviews{AccountID: "acc-1"}
	ch <- &chmodels.GMBReviews{AccountID: "acc-2"}
	ch <- &chmodels.GMBReviews{AccountID: "acc-3"}
	close(ch)

	var wg sync.WaitGroup
	var insertedCounter uint64
	wg.Add(1)

	log := newTestLogger()
	go reviewsBatchProcessor(ctx, &wg, 0, ch, sink, log, &insertedCounter)
	wg.Wait()

	if atomic.LoadUint64(&insertedCounter) != 3 {
		t.Fatalf("expected 3 inserted, got %d", atomic.LoadUint64(&insertedCounter))
	}
}

func TestMediaAssetsBatchProcessor_FlushOnClose(t *testing.T) {
	ctx := context.Background()
	sink := NewMockClickHouseSink()

	ch := make(chan *chmodels.GMBMediaAssets, 100)
	ch <- &chmodels.GMBMediaAssets{AccountID: "acc-1"}
	close(ch)

	var wg sync.WaitGroup
	var insertedCounter uint64
	wg.Add(1)

	log := newTestLogger()
	go mediaAssetsBatchProcessor(ctx, &wg, 0, ch, sink, log, &insertedCounter)
	wg.Wait()

	if atomic.LoadUint64(&insertedCounter) != 1 {
		t.Fatalf("expected 1 inserted, got %d", atomic.LoadUint64(&insertedCounter))
	}
}

func TestSearchKeywordsBatchProcessor_FlushOnClose(t *testing.T) {
	ctx := context.Background()
	sink := NewMockClickHouseSink()

	ch := make(chan *chmodels.GMBSearchKeywordsMonthly, 100)
	ch <- &chmodels.GMBSearchKeywordsMonthly{AccountID: "acc-1"}
	ch <- &chmodels.GMBSearchKeywordsMonthly{AccountID: "acc-2"}
	close(ch)

	var wg sync.WaitGroup
	var insertedCounter uint64
	wg.Add(1)

	log := newTestLogger()
	go searchKeywordsBatchProcessor(ctx, &wg, 0, ch, sink, log, &insertedCounter)
	wg.Wait()

	if atomic.LoadUint64(&insertedCounter) != 2 {
		t.Fatalf("expected 2 inserted, got %d", atomic.LoadUint64(&insertedCounter))
	}
}

func TestLocalPostsBatchProcessor_FlushOnClose(t *testing.T) {
	ctx := context.Background()
	sink := NewMockClickHouseSink()

	ch := make(chan *chmodels.GMBLocalPosts, 100)
	ch <- &chmodels.GMBLocalPosts{AccountID: "acc-1"}
	close(ch)

	var wg sync.WaitGroup
	var insertedCounter uint64
	wg.Add(1)

	log := newTestLogger()
	go localPostsBatchProcessor(ctx, &wg, 0, ch, sink, log, &insertedCounter)
	wg.Wait()

	if atomic.LoadUint64(&insertedCounter) != 1 {
		t.Fatalf("expected 1 inserted, got %d", atomic.LoadUint64(&insertedCounter))
	}
}

// ================== GMBDailyMetricsBuilder Tests ==================

func TestGMBDailyMetricsBuilder_SetMetric(t *testing.T) {
	b := &conversions.GMBDailyMetricsBuilder{
		AccountID:  "acc-1",
		LocationID: "loc-1",
		Date:       "2024-01-15",
	}
	b.SetMetric("CALL_CLICKS", "42")
	b.SetMetric("WEBSITE_CLICKS", "100")

	result := b.Build()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.CallClicks != 42 {
		t.Fatalf("expected CallClicks 42, got %d", result.CallClicks)
	}
	if result.WebsiteClicks != 100 {
		t.Fatalf("expected WebsiteClicks 100, got %d", result.WebsiteClicks)
	}
	if result.PlatformName != "gmb" {
		t.Fatalf("expected PlatformName 'gmb', got %s", result.PlatformName)
	}
}

func TestGMBDailyMetricsBuilder_AllMetrics(t *testing.T) {
	b := &conversions.GMBDailyMetricsBuilder{
		AccountID:  "acc-1",
		LocationID: "loc-1",
		Date:       "2024-01-15",
	}
	b.SetMetric("BUSINESS_IMPRESSIONS_DESKTOP_MAPS", "100")
	b.SetMetric("BUSINESS_IMPRESSIONS_DESKTOP_SEARCH", "200")
	b.SetMetric("BUSINESS_IMPRESSIONS_MOBILE_MAPS", "300")
	b.SetMetric("BUSINESS_IMPRESSIONS_MOBILE_SEARCH", "400")
	b.SetMetric("CALL_CLICKS", "50")
	b.SetMetric("WEBSITE_CLICKS", "60")
	b.SetMetric("BUSINESS_DIRECTION_REQUESTS", "70")
	b.SetMetric("BUSINESS_CONVERSATIONS", "10")
	b.SetMetric("BUSINESS_BOOKINGS", "5")
	b.SetMetric("BUSINESS_FOOD_ORDERS", "3")
	b.SetMetric("BUSINESS_FOOD_MENU_CLICKS", "8")

	result := b.Build()
	if result.BusinessImpressionsDesktopMaps != 100 {
		t.Fatalf("expected 100, got %d", result.BusinessImpressionsDesktopMaps)
	}
	if result.BusinessFoodMenuClicks != 8 {
		t.Fatalf("expected 8, got %d", result.BusinessFoodMenuClicks)
	}
}

// ================== Logging Contract Tests ==================

func TestParsePerformanceMetrics_EmptyResponse(t *testing.T) {
	ctx := context.Background()
	out := make(chan *chmodels.GMBDailyMetrics, 100)
	var counter uint64
	log := newTestLogger()

	dataBytes := []byte(`{"multiDailyMetricTimeSeries":[]}`)
	raw := kafkamodels.RawGMBData{AccountID: "acc-1", LocationID: "loc-1"}
	parsePerformanceMetrics(ctx, dataBytes, raw, out, &counter, log)

	if counter != 0 {
		t.Fatalf("expected 0 parsed, got %d", counter)
	}
}

func TestParseReviews_EmptyResponse(t *testing.T) {
	ctx := context.Background()
	out := make(chan *chmodels.GMBReviews, 100)
	var counter uint64
	log := newTestLogger()

	dataBytes := []byte(`{"reviews":[]}`)
	raw := kafkamodels.RawGMBData{AccountID: "acc-1", LocationID: "loc-1"}
	parseReviews(ctx, dataBytes, raw, out, &counter, log)

	if counter != 0 {
		t.Fatalf("expected 0 parsed, got %d", counter)
	}
}

func TestParseMediaAssets_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	out := make(chan *chmodels.GMBMediaAssets, 100)
	var counter uint64
	log := newTestLogger()

	dataBytes := []byte(`invalid json`)
	raw := kafkamodels.RawGMBData{AccountID: "acc-1", LocationID: "loc-1"}
	parseMediaAssets(ctx, dataBytes, raw, out, &counter, log)

	if counter != 0 {
		t.Fatalf("expected 0 parsed on invalid JSON, got %d", counter)
	}
}

func TestParseSearchKeywords_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	out := make(chan *chmodels.GMBSearchKeywordsMonthly, 100)
	var counter uint64
	log := newTestLogger()

	dataBytes := []byte(`not json`)
	raw := kafkamodels.RawGMBData{AccountID: "acc-1", LocationID: "loc-1"}
	parseSearchKeywords(ctx, dataBytes, raw, out, &counter, log)

	if counter != 0 {
		t.Fatalf("expected 0 parsed on invalid JSON, got %d", counter)
	}
}

func TestParseLocalPosts_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	out := make(chan *chmodels.GMBLocalPosts, 100)
	var counter uint64
	log := newTestLogger()

	dataBytes := []byte(`{broken}`)
	raw := kafkamodels.RawGMBData{AccountID: "acc-1", LocationID: "loc-1"}
	parseLocalPosts(ctx, dataBytes, raw, out, &counter, log)

	if counter != 0 {
		t.Fatalf("expected 0 parsed on invalid JSON, got %d", counter)
	}
}

// ================== Idle Timeout Tests ==================

func TestIdleTimeout_Values(t *testing.T) {
	if idleTimeout != 5*time.Minute {
		t.Fatalf("expected 5m idle timeout, got %v", idleTimeout)
	}
	if idleCheckInterval != 30*time.Second {
		t.Fatalf("expected 30s idle check interval, got %v", idleCheckInterval)
	}
}

// ================== Logging Contract Tests ==================

func TestLoggingContract_GMBSink_NoCaptureException(t *testing.T) {
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	log.Error().
		Str("error_message", "ClickHouse batch insert failed").
		Str("function", "dailyMetricsBatchProcessor").
		Str("stage", "bulk_insert").
		Msg("Failed to insert daily metrics batch")

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls (hook handles Sentry), got %d", len(*captureRecords))
	}
}
