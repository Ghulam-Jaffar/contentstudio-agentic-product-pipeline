package clickhouse

import (
	"context"
	"errors"
	"testing"
	"time"

	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

// --- BulkInsertGMBDailyMetrics ---

func Test_BulkInsertGMBDailyMetrics_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.BulkInsertGMBDailyMetrics(context.Background(), []*chmodels.GMBDailyMetrics{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_BulkInsertGMBDailyMetrics_NilSlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.BulkInsertGMBDailyMetrics(context.Background(), nil)
	if err != nil {
		t.Fatalf("expected nil error for nil slice, got %v", err)
	}
}

func Test_BulkInsertGMBDailyMetrics_WithAllFields(t *testing.T) {
	now := time.Now()
	metrics := []*chmodels.GMBDailyMetrics{
		{
			AccountID: "acc_1", LocationID: "loc_1", AccountName: "Test Account",
			LocationName: "Test Location", PlatformName: "google",
			InsertedAt: now, CreatedAt: now,
			BusinessImpressionsDesktopMaps:   100,
			BusinessImpressionsDesktopSearch: 200,
			BusinessImpressionsMobileMaps:    300,
			BusinessImpressionsMobileSearch:  400,
			CallClicks: 50, WebsiteClicks: 60, BusinessDirectionRequests: 30,
			BusinessConversations: 10, BusinessBookings: 5,
			BusinessFoodOrders: 2, BusinessFoodMenuClicks: 3,
		},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertGMBDailyMetrics(context.Background(), metrics)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertGMBDailyMetrics_MultipleRecords(t *testing.T) {
	now := time.Now()
	metrics := make([]*chmodels.GMBDailyMetrics, 3)
	for i := range metrics {
		metrics[i] = &chmodels.GMBDailyMetrics{
			AccountID: "acc_1", LocationID: "loc_1",
			InsertedAt: now, CreatedAt: now.AddDate(0, 0, -i),
		}
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertGMBDailyMetrics(context.Background(), metrics)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertGMBDailyMetrics_PrepareBatchError(t *testing.T) {
	client := newTestClient(&mockConn{prepareBatchErr: errors.New("prepare failed")})
	metrics := []*chmodels.GMBDailyMetrics{{AccountID: "acc_1", LocationID: "loc_1"}}
	err := client.BulkInsertGMBDailyMetrics(context.Background(), metrics)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func Test_BulkInsertGMBDailyMetrics_AppendError(t *testing.T) {
	client := newTestClient(&mockConn{batchAppendErr: errors.New("append failed")})
	metrics := []*chmodels.GMBDailyMetrics{{AccountID: "acc_1", LocationID: "loc_1"}}
	err := client.BulkInsertGMBDailyMetrics(context.Background(), metrics)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func Test_BulkInsertGMBDailyMetrics_SendError(t *testing.T) {
	client := newTestClient(&mockConn{batchSendErr: errors.New("send failed")})
	metrics := []*chmodels.GMBDailyMetrics{{AccountID: "acc_1", LocationID: "loc_1"}}
	err := client.BulkInsertGMBDailyMetrics(context.Background(), metrics)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func Test_BulkInsertGMBDailyMetrics_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(&mockConn{})
	metrics := []*chmodels.GMBDailyMetrics{{AccountID: "acc_1", LocationID: "loc_1"}}
	_ = client.BulkInsertGMBDailyMetrics(ctx, metrics)
}

// --- BulkInsertGMBMediaAssets ---

func Test_BulkInsertGMBMediaAssets_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.BulkInsertGMBMediaAssets(context.Background(), []*chmodels.GMBMediaAssets{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_BulkInsertGMBMediaAssets_WithAllFields(t *testing.T) {
	now := time.Now()
	assets := []*chmodels.GMBMediaAssets{
		{
			AccountID: "acc_1", LocationID: "loc_1", AccountName: "Test Account",
			LocationName: "Test Location", PlatformName: "google",
			LanguageCode: "en", InsertedAt: now, CreatedAt: now,
			MediaName: "photo_1", SourceURL: "https://example.com/photo.jpg",
			MediaFormat: "PHOTO", LocationAssociationCategory: "EXTERIOR",
			GoogleURL: "https://lh3.google.com/photo", ThumbnailURL: "https://lh3.google.com/thumb",
			WidthPixels: 1920, HeightPixels: 1080,
		},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertGMBMediaAssets(context.Background(), assets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertGMBMediaAssets_PrepareBatchError(t *testing.T) {
	client := newTestClient(&mockConn{prepareBatchErr: errors.New("prepare failed")})
	assets := []*chmodels.GMBMediaAssets{{AccountID: "acc_1", LocationID: "loc_1"}}
	err := client.BulkInsertGMBMediaAssets(context.Background(), assets)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func Test_BulkInsertGMBMediaAssets_AppendError(t *testing.T) {
	client := newTestClient(&mockConn{batchAppendErr: errors.New("append failed")})
	assets := []*chmodels.GMBMediaAssets{{AccountID: "acc_1", LocationID: "loc_1"}}
	err := client.BulkInsertGMBMediaAssets(context.Background(), assets)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func Test_BulkInsertGMBMediaAssets_SendError(t *testing.T) {
	client := newTestClient(&mockConn{batchSendErr: errors.New("send failed")})
	assets := []*chmodels.GMBMediaAssets{{AccountID: "acc_1", LocationID: "loc_1"}}
	err := client.BulkInsertGMBMediaAssets(context.Background(), assets)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- BulkInsertGMBSearchKeywordsMonthly ---

func Test_BulkInsertGMBSearchKeywordsMonthly_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.BulkInsertGMBSearchKeywordsMonthly(context.Background(), []*chmodels.GMBSearchKeywordsMonthly{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_BulkInsertGMBSearchKeywordsMonthly_WithAllFields(t *testing.T) {
	now := time.Now()
	keywords := []*chmodels.GMBSearchKeywordsMonthly{
		{
			AccountID: "acc_1", LocationID: "loc_1", AccountName: "Test Account",
			LocationName: "Test Location", PlatformName: "google",
			InsertedAt: now, KeywordMonth: now,
			Keyword: "pizza near me", ImpressionsValue: 1500, ImpressionsThreshold: 0,
		},
		{
			AccountID: "acc_1", LocationID: "loc_1", AccountName: "Test Account",
			LocationName: "Test Location", PlatformName: "google",
			InsertedAt: now, KeywordMonth: now,
			Keyword: "best restaurant", ImpressionsValue: 800, ImpressionsThreshold: 1,
		},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertGMBSearchKeywordsMonthly(context.Background(), keywords)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertGMBSearchKeywordsMonthly_PrepareBatchError(t *testing.T) {
	client := newTestClient(&mockConn{prepareBatchErr: errors.New("prepare failed")})
	keywords := []*chmodels.GMBSearchKeywordsMonthly{{AccountID: "acc_1", LocationID: "loc_1"}}
	err := client.BulkInsertGMBSearchKeywordsMonthly(context.Background(), keywords)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func Test_BulkInsertGMBSearchKeywordsMonthly_AppendError(t *testing.T) {
	client := newTestClient(&mockConn{batchAppendErr: errors.New("append failed")})
	keywords := []*chmodels.GMBSearchKeywordsMonthly{{AccountID: "acc_1", LocationID: "loc_1"}}
	err := client.BulkInsertGMBSearchKeywordsMonthly(context.Background(), keywords)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func Test_BulkInsertGMBSearchKeywordsMonthly_SendError(t *testing.T) {
	client := newTestClient(&mockConn{batchSendErr: errors.New("send failed")})
	keywords := []*chmodels.GMBSearchKeywordsMonthly{{AccountID: "acc_1", LocationID: "loc_1"}}
	err := client.BulkInsertGMBSearchKeywordsMonthly(context.Background(), keywords)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- BulkInsertGMBLocalPosts ---

func Test_BulkInsertGMBLocalPosts_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.BulkInsertGMBLocalPosts(context.Background(), []*chmodels.GMBLocalPosts{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_BulkInsertGMBLocalPosts_WithAllFields(t *testing.T) {
	now := time.Now()
	posts := []*chmodels.GMBLocalPosts{
		{
			AccountID: "acc_1", LocationID: "loc_1", AccountName: "Test Account",
			LocationName: "Test Location", PlatformName: "google",
			LanguageCode: "en", InsertedAt: now, CreatedAt: now, UpdatedAt: now,
			PostName: "locations/loc_1/localPosts/post_1",
			State: "LIVE", TopicType: "STANDARD",
			SearchURL: "https://search.google.com/local/posts?q=test",
			MediaNames: []string{"media_1"}, MediaFormats: []string{"PHOTO"},
			MediaGoogleURLs: []string{"https://lh3.google.com/media_1"},
		},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertGMBLocalPosts(context.Background(), posts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertGMBLocalPosts_PrepareBatchError(t *testing.T) {
	client := newTestClient(&mockConn{prepareBatchErr: errors.New("prepare failed")})
	posts := []*chmodels.GMBLocalPosts{{AccountID: "acc_1", LocationID: "loc_1"}}
	err := client.BulkInsertGMBLocalPosts(context.Background(), posts)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func Test_BulkInsertGMBLocalPosts_AppendError(t *testing.T) {
	client := newTestClient(&mockConn{batchAppendErr: errors.New("append failed")})
	posts := []*chmodels.GMBLocalPosts{{AccountID: "acc_1", LocationID: "loc_1"}}
	err := client.BulkInsertGMBLocalPosts(context.Background(), posts)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func Test_BulkInsertGMBLocalPosts_SendError(t *testing.T) {
	client := newTestClient(&mockConn{batchSendErr: errors.New("send failed")})
	posts := []*chmodels.GMBLocalPosts{{AccountID: "acc_1", LocationID: "loc_1"}}
	err := client.BulkInsertGMBLocalPosts(context.Background(), posts)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- BulkInsertGMBReviews ---

func Test_BulkInsertGMBReviews_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.BulkInsertGMBReviews(context.Background(), []*chmodels.GMBReviews{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_BulkInsertGMBReviews_WithAllFields(t *testing.T) {
	now := time.Now()
	reviews := []*chmodels.GMBReviews{
		{
			AccountID: "acc_1", LocationID: "loc_1", AccountName: "Test Account",
			LocationName: "Test Location", PlatformName: "google",
			InsertedAt: now, CreatedAt: now, UpdatedAt: now,
			ReviewID: "review_1", ReviewName: "accounts/acc_1/locations/loc_1/reviews/review_1",
			ReviewerDisplayName: "John Doe",
			ReviewerProfilePhotoURL: "https://lh3.google.com/photo",
			StarRating: 5, Comment: "Great place!",
			ReplyComment: "Thank you!", ReplyUpdateTime: now,
		},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertGMBReviews(context.Background(), reviews)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertGMBReviews_PrepareBatchError(t *testing.T) {
	client := newTestClient(&mockConn{prepareBatchErr: errors.New("prepare failed")})
	reviews := []*chmodels.GMBReviews{{AccountID: "acc_1", LocationID: "loc_1"}}
	err := client.BulkInsertGMBReviews(context.Background(), reviews)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func Test_BulkInsertGMBReviews_AppendError(t *testing.T) {
	client := newTestClient(&mockConn{batchAppendErr: errors.New("append failed")})
	reviews := []*chmodels.GMBReviews{{AccountID: "acc_1", LocationID: "loc_1"}}
	err := client.BulkInsertGMBReviews(context.Background(), reviews)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func Test_BulkInsertGMBReviews_SendError(t *testing.T) {
	client := newTestClient(&mockConn{batchSendErr: errors.New("send failed")})
	reviews := []*chmodels.GMBReviews{{AccountID: "acc_1", LocationID: "loc_1"}}
	err := client.BulkInsertGMBReviews(context.Background(), reviews)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func Test_BulkInsertGMBReviews_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(&mockConn{})
	reviews := []*chmodels.GMBReviews{{AccountID: "acc_1", LocationID: "loc_1"}}
	_ = client.BulkInsertGMBReviews(ctx, reviews)
}
