package conversions

import (
	"context"
	"testing"

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

func TestStarRatingToUint64(t *testing.T) {
	tests := []struct {
		input    string
		expected uint64
	}{
		{"ONE", 1},
		{"TWO", 2},
		{"THREE", 3},
		{"FOUR", 4},
		{"FIVE", 5},
		{"one", 1},
		{"Five", 5},
		{"3", 3},
		{"UNKNOWN", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := starRatingToUint64(tt.input)
			if got != tt.expected {
				t.Errorf("starRatingToUint64(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestConvertGMBDailyMetrics_NilInput(t *testing.T) {
	result := ConvertGMBDailyMetrics(nil)
	if result != nil {
		t.Fatal("expected nil result for nil input")
	}
}

func TestConvertGMBDailyMetrics_ValidInput(t *testing.T) {
	input := &kafkamodels.ParsedGMBDailyMetrics{
		LocationID:                       "loc-1",
		AccountID:                        "acc-1",
		AccountName:                      "Test Account",
		LocationName:                     "Test Location",
		Date:                             "2024-01-15",
		BusinessImpressionsDesktopMaps:   100,
		BusinessImpressionsDesktopSearch: 200,
		BusinessImpressionsMobileMaps:    300,
		BusinessImpressionsMobileSearch:  400,
		CallClicks:                       50,
		WebsiteClicks:                    60,
		BusinessDirectionRequests:        70,
		BusinessConversations:            10,
		BusinessBookings:                 5,
		BusinessFoodOrders:               3,
		BusinessFoodMenuClicks:           8,
	}

	result := ConvertGMBDailyMetrics(input)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.CallClicks != 50 {
		t.Fatalf("expected CallClicks 50, got %d", result.CallClicks)
	}
	if result.InsertedAt.IsZero() {
		t.Fatal("expected InsertedAt to be set")
	}
}

func TestConvertGMBMediaAsset_NilInput(t *testing.T) {
	result := ConvertGMBMediaAsset(nil)
	if result != nil {
		t.Fatal("expected nil result for nil input")
	}
}

func TestConvertGMBMediaAsset_ValidInput(t *testing.T) {
	input := &kafkamodels.ParsedGMBMediaAsset{
		LocationID:                  "loc-1",
		AccountID:                   "acc-1",
		AccountName:                 "Test Account",
		LocationName:                "Test Location",
		LanguageCode:                "en",
		MediaName:                   "accounts/123/locations/456/media/789",
		MediaFormat:                 "PHOTO",
		LocationAssociationCategory: "INTERIOR",
		GoogleURL:                   "https://lh3.googleusercontent.com/photo",
		ThumbnailURL:                "https://lh3.googleusercontent.com/thumb",
		SourceURL:                   "https://example.com/photo.jpg",
		WidthPixels:                 1920,
		HeightPixels:                1080,
		CreateTime:                  "2024-01-15T10:00:00Z",
	}

	result := ConvertGMBMediaAsset(input)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.MediaFormat != "PHOTO" {
		t.Fatalf("expected MediaFormat 'PHOTO', got %s", result.MediaFormat)
	}
	if result.WidthPixels != 1920 {
		t.Fatalf("expected WidthPixels 1920, got %d", result.WidthPixels)
	}
}

func TestConvertGMBSearchKeyword_NilInput(t *testing.T) {
	result := ConvertGMBSearchKeyword(nil)
	if result != nil {
		t.Fatal("expected nil result for nil input")
	}
}

func TestConvertGMBSearchKeyword_ValidInput(t *testing.T) {
	input := &kafkamodels.ParsedGMBSearchKeyword{
		LocationID:           "loc-1",
		AccountID:            "acc-1",
		AccountName:          "Test Account",
		LocationName:         "Test Location",
		KeywordMonth:         "2024-01",
		Keyword:              "pizza near me",
		ImpressionsValue:     150,
		ImpressionsThreshold: 21,
	}

	result := ConvertGMBSearchKeyword(input)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Keyword != "pizza near me" {
		t.Fatalf("expected Keyword 'pizza near me', got %s", result.Keyword)
	}
	if result.ImpressionsValue != 150 {
		t.Fatalf("expected ImpressionsValue 150, got %d", result.ImpressionsValue)
	}
}

func TestConvertGMBLocalPost_NilInput(t *testing.T) {
	result := ConvertGMBLocalPost(nil)
	if result != nil {
		t.Fatal("expected nil result for nil input")
	}
}

func TestConvertGMBLocalPost_ValidInput(t *testing.T) {
	input := &kafkamodels.ParsedGMBLocalPost{
		LocationID:      "loc-1",
		AccountID:       "acc-1",
		AccountName:     "Test Account",
		LocationName:    "Test Location",
		PostName:        "accounts/123/locations/456/localPosts/789",
		LanguageCode:    "en",
		State:           "LIVE",
		TopicType:       "STANDARD",
		SearchURL:       "https://search.google.com/local/posts?q=test",
		MediaNames:      []string{"media-1"},
		MediaFormats:    []string{"PHOTO"},
		MediaGoogleURLs: []string{"https://lh3.googleusercontent.com/photo1"},
		CreateTime:      "2024-01-15T10:00:00Z",
		UpdateTime:      "2024-01-15T12:00:00Z",
	}

	result := ConvertGMBLocalPost(input)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.State != "LIVE" {
		t.Fatalf("expected State 'LIVE', got %s", result.State)
	}
	if len(result.MediaNames) != 1 {
		t.Fatalf("expected 1 media name, got %d", len(result.MediaNames))
	}
}

func TestConvertGMBReview_NilInput(t *testing.T) {
	result := ConvertGMBReview(nil)
	if result != nil {
		t.Fatal("expected nil result for nil input")
	}
}

func TestConvertGMBReview_ValidInput(t *testing.T) {
	input := &kafkamodels.ParsedGMBReview{
		LocationID:              "loc-1",
		AccountID:               "acc-1",
		AccountName:             "Test Account",
		LocationName:            "Test Location",
		ReviewID:                "review-001",
		ReviewerDisplayName:     "John Doe",
		ReviewerProfilePhotoURL: "https://lh3.googleusercontent.com/photo",
		StarRating:              "FIVE",
		Comment:                 "Great food!",
		CreateTime:              "2024-01-15T10:00:00Z",
		UpdateTime:              "2024-01-15T10:00:00Z",
		ReplyComment:    "Thanks for your review!",
		ReplyUpdateTime: "2024-01-16T08:00:00Z",
	}

	result := ConvertGMBReview(input)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.StarRating != 5 {
		t.Fatalf("expected StarRating 5, got %d", result.StarRating)
	}
	if result.ReviewID != "review-001" {
		t.Fatalf("expected ReviewID 'review-001', got %s", result.ReviewID)
	}
	if result.ReplyComment != "Thanks for your review!" {
		t.Fatalf("expected reply comment 'Thanks for your review!', got %s", result.ReplyComment)
	}
}

func TestConvertGMBReview_NoReply(t *testing.T) {
	input := &kafkamodels.ParsedGMBReview{
		LocationID:              "loc-1",
		AccountID:               "acc-1",
		ReviewID:                "review-002",
		ReviewerDisplayName:     "Jane Doe",
		ReviewerProfilePhotoURL: "",
		StarRating:              "THREE",
		Comment:                 "Average experience",
		CreateTime:              "2024-01-15T10:00:00Z",
		UpdateTime:              "2024-01-15T10:00:00Z",
	}

	result := ConvertGMBReview(input)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.StarRating != 3 {
		t.Fatalf("expected StarRating 3, got %d", result.StarRating)
	}
	if result.ReplyComment != "" {
		t.Fatalf("expected empty reply comment, got %s", result.ReplyComment)
	}
}

// BulkInsert wrapper tests

func TestBulkInsertGMBDailyMetrics_Success(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertGMBDailyMetricsFunc: func(ctx context.Context, metrics []*clickhousemodels.GMBDailyMetrics) error {
			called = true
			return nil
		},
	}
	sink := newTestSink(mock)

	metrics := []*clickhousemodels.GMBDailyMetrics{{AccountID: "acc-1"}}
	err := sink.BulkInsertGMBDailyMetrics(context.Background(), metrics)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertGMBDailyMetrics to be called")
	}
}

func TestBulkInsertGMBMediaAssets_Success(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertGMBMediaAssetsFunc: func(ctx context.Context, assets []*clickhousemodels.GMBMediaAssets) error {
			called = true
			return nil
		},
	}
	sink := newTestSink(mock)

	assets := []*clickhousemodels.GMBMediaAssets{{AccountID: "acc-1"}}
	err := sink.BulkInsertGMBMediaAssets(context.Background(), assets)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertGMBMediaAssets to be called")
	}
}

func TestBulkInsertGMBSearchKeywordsMonthly_Success(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertGMBSearchKeywordsMonthlyFunc: func(ctx context.Context, keywords []*clickhousemodels.GMBSearchKeywordsMonthly) error {
			called = true
			return nil
		},
	}
	sink := newTestSink(mock)

	keywords := []*clickhousemodels.GMBSearchKeywordsMonthly{{AccountID: "acc-1"}}
	err := sink.BulkInsertGMBSearchKeywordsMonthly(context.Background(), keywords)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertGMBSearchKeywordsMonthly to be called")
	}
}

func TestBulkInsertGMBLocalPosts_Success(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertGMBLocalPostsFunc: func(ctx context.Context, posts []*clickhousemodels.GMBLocalPosts) error {
			called = true
			return nil
		},
	}
	sink := newTestSink(mock)

	posts := []*clickhousemodels.GMBLocalPosts{{AccountID: "acc-1"}}
	err := sink.BulkInsertGMBLocalPosts(context.Background(), posts)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertGMBLocalPosts to be called")
	}
}

func TestBulkInsertGMBReviews_Success(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertGMBReviewsFunc: func(ctx context.Context, reviews []*clickhousemodels.GMBReviews) error {
			called = true
			return nil
		},
	}
	sink := newTestSink(mock)

	reviews := []*clickhousemodels.GMBReviews{{AccountID: "acc-1"}}
	err := sink.BulkInsertGMBReviews(context.Background(), reviews)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertGMBReviews to be called")
	}
}
