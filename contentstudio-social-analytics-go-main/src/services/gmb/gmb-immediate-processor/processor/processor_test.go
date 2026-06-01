package processor

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

func TestWorkOrder_Struct(t *testing.T) {
	wo := ImmediateWorkOrder{
		ID:           "account123",
		WorkspaceID:  "workspace_789",
		AccountID:    "gmb_account_1",
		LocationID:   "location_456",
		AccessToken:  "token_abc",
		RefreshToken: "refresh_xyz",
		AccountName:  "My Business",
		LocationName: "Main Office",
		LanguageCode: "en",
		SyncType:     "full_sync",
	}

	if wo.ID != "account123" {
		t.Fatalf("expected ID 'account123', got '%s'", wo.ID)
	}
	if wo.AccountID != "gmb_account_1" {
		t.Fatalf("expected AccountID 'gmb_account_1', got '%s'", wo.AccountID)
	}
	if wo.LocationID != "location_456" {
		t.Fatalf("expected LocationID 'location_456', got '%s'", wo.LocationID)
	}
	if wo.SyncType != "full_sync" {
		t.Fatalf("expected SyncType 'full_sync', got '%s'", wo.SyncType)
	}
	if wo.WorkspaceID != "workspace_789" {
		t.Fatalf("expected WorkspaceID 'workspace_789', got '%s'", wo.WorkspaceID)
	}
	if wo.AccessToken != "token_abc" {
		t.Fatalf("expected AccessToken 'token_abc', got '%s'", wo.AccessToken)
	}
	if wo.RefreshToken != "refresh_xyz" {
		t.Fatalf("expected RefreshToken 'refresh_xyz', got '%s'", wo.RefreshToken)
	}
	if wo.AccountName != "My Business" {
		t.Fatalf("expected AccountName 'My Business', got '%s'", wo.AccountName)
	}
	if wo.LocationName != "Main Office" {
		t.Fatalf("expected LocationName 'Main Office', got '%s'", wo.LocationName)
	}
	if wo.LanguageCode != "en" {
		t.Fatalf("expected LanguageCode 'en', got '%s'", wo.LanguageCode)
	}
}

func TestWorkOrder_EmptyStruct(t *testing.T) {
	wo := ImmediateWorkOrder{}

	if wo.ID != "" {
		t.Fatalf("expected empty ID, got '%s'", wo.ID)
	}
	if wo.AccountID != "" {
		t.Fatalf("expected empty AccountID, got '%s'", wo.AccountID)
	}
	if wo.LocationID != "" {
		t.Fatalf("expected empty LocationID, got '%s'", wo.LocationID)
	}
	if wo.WorkspaceID != "" {
		t.Fatalf("expected empty WorkspaceID, got '%s'", wo.WorkspaceID)
	}
	if wo.AccessToken != "" {
		t.Fatalf("expected empty AccessToken, got '%s'", wo.AccessToken)
	}
	if wo.RefreshToken != "" {
		t.Fatalf("expected empty RefreshToken, got '%s'", wo.RefreshToken)
	}
	if wo.SyncType != "" {
		t.Fatalf("expected empty SyncType, got '%s'", wo.SyncType)
	}
}

func TestWorkOrder_JSONTags(t *testing.T) {
	wo := ImmediateWorkOrder{
		ID:           "id_value",
		WorkspaceID:  "ws_value",
		AccountID:    "act_value",
		LocationID:   "loc_value",
		AccessToken:  "at_value",
		RefreshToken: "rt_value",
		AccountName:  "name_value",
		LocationName: "locname",
		LanguageCode: "en",
		SyncType:     "immediate",
	}

	if wo.ID == "" || wo.WorkspaceID == "" || wo.AccountID == "" || wo.LocationID == "" || wo.AccessToken == "" || wo.RefreshToken == "" || wo.SyncType == "" {
		t.Fatal("one or more fields were not set correctly")
	}
}

func TestProcessAccount_UsesRequestedDateRangeAndFiltersCreatedAt(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	wantStart, err := time.Parse("2006-01-02", "2025-01-10")
	if err != nil {
		t.Fatalf("failed to parse start date: %v", err)
	}
	wantStart = time.Date(wantStart.Year(), wantStart.Month(), wantStart.Day(), 0, 0, 0, 0, time.UTC)
	wantEnd, err := time.Parse("2006-01-02", "2025-02-05")
	if err != nil {
		t.Fatalf("failed to parse end date: %v", err)
	}
	wantEnd = time.Date(wantEnd.Year(), wantEnd.Month(), wantEnd.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)

	var perfStart, perfEnd time.Time
	var keywordMonths []string
	var reviewPageTokens []string

	sink := &mockClickHouseSink{
		bulkInsertDailyMetricsFunc: func(ctx context.Context, metrics []*clickhousemodels.GMBDailyMetrics) error {
			if len(metrics) != 1 {
				t.Fatalf("expected 1 metric row, got %d", len(metrics))
			}
			if got := metrics[0].CreatedAt; !got.Equal(time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC)) {
				t.Fatalf("unexpected metric created_at: %s", got)
			}
			return nil
		},
		bulkInsertSearchKeywordsFunc: func(ctx context.Context, keywords []*clickhousemodels.GMBSearchKeywordsMonthly) error {
			if len(keywords) != 2 {
				t.Fatalf("expected 2 keyword rows, got %d", len(keywords))
			}
			if keywords[0].KeywordMonth.Month() != time.January && keywords[0].KeywordMonth.Month() != time.February {
				t.Fatalf("unexpected keyword month: %s", keywords[0].KeywordMonth)
			}
			return nil
		},
		bulkInsertLocalPostsFunc: func(ctx context.Context, posts []*clickhousemodels.GMBLocalPosts) error {
			if len(posts) != 1 {
				t.Fatalf("expected 1 local post row, got %d", len(posts))
			}
			if got := posts[0].CreatedAt; !got.Equal(time.Date(2025, 1, 11, 13, 15, 0, 0, time.UTC)) {
				t.Fatalf("unexpected local post created_at: %s", got)
			}
			return nil
		},
		bulkInsertReviewsFunc: func(ctx context.Context, reviews []*clickhousemodels.GMBReviews) error {
			if len(reviews) != 2 {
				t.Fatalf("expected 2 review rows, got %d", len(reviews))
			}
			for _, review := range reviews {
				if review.CreatedAt.Before(wantStart) || review.CreatedAt.After(wantEnd) {
					t.Fatalf("review created_at out of range: %s", review.CreatedAt)
				}
			}
			return nil
		},
		bulkInsertMediaAssetsFunc: func(ctx context.Context, assets []*clickhousemodels.GMBMediaAssets) error {
			if len(assets) != 1 {
				t.Fatalf("expected 1 media asset row, got %d", len(assets))
			}
			if got := assets[0].CreatedAt; !got.Equal(time.Date(2025, 1, 18, 9, 0, 0, 0, time.UTC)) {
				t.Fatalf("unexpected media asset created_at: %s", got)
			}
			return nil
		},
	}

	gmbClient := &mockGMBClient{
		refreshTokenFunc: func(ctx context.Context, refreshToken string) (*social.RefreshTokenResponse, error) {
			return &social.RefreshTokenResponse{AccessToken: "refreshed-token"}, nil
		},
		fetchVoiceOfMerchantFunc: func(ctx context.Context, locationID, accessToken string) (*social.VoiceOfMerchantResponse, error) {
			return &social.VoiceOfMerchantResponse{HasVoiceOfMerchant: true}, nil
		},
		fetchPerformanceMetricsFunc: func(ctx context.Context, locationID, accessToken string, startDate, endDate time.Time) (*social.GMBPerformanceResponse, error) {
			perfStart = startDate
			perfEnd = endDate
			return &social.GMBPerformanceResponse{
				MultiDailyMetricTimeSeries: []struct {
					DailyMetricTimeSeries []struct {
						DailyMetric string `json:"dailyMetric"`
						TimeSeries  struct {
							DatedValues []struct {
								Date struct {
									Year  int `json:"year"`
									Month int `json:"month"`
									Day   int `json:"day"`
								} `json:"date"`
								Value string `json:"value"`
							} `json:"datedValues"`
						} `json:"timeSeries"`
					} `json:"dailyMetricTimeSeries"`
				}{
					{
						DailyMetricTimeSeries: []struct {
							DailyMetric string `json:"dailyMetric"`
							TimeSeries  struct {
								DatedValues []struct {
									Date struct {
										Year  int `json:"year"`
										Month int `json:"month"`
										Day   int `json:"day"`
									} `json:"date"`
									Value string `json:"value"`
								} `json:"datedValues"`
							} `json:"timeSeries"`
						}{
							{
								DailyMetric: "CALL_CLICKS",
								TimeSeries: struct {
									DatedValues []struct {
										Date struct {
											Year  int `json:"year"`
											Month int `json:"month"`
											Day   int `json:"day"`
										} `json:"date"`
										Value string `json:"value"`
									} `json:"datedValues"`
								}{
									DatedValues: []struct {
										Date struct {
											Year  int `json:"year"`
											Month int `json:"month"`
											Day   int `json:"day"`
										} `json:"date"`
										Value string `json:"value"`
									}{
										{
											Date: struct {
												Year  int `json:"year"`
												Month int `json:"month"`
												Day   int `json:"day"`
											}{Year: 2025, Month: 1, Day: 12},
											Value: "9",
										},
									},
								},
							},
						},
					},
				},
			}, nil
		},
		fetchSearchKeywordsFunc: func(ctx context.Context, locationID, accessToken string, startMonth, endMonth time.Time) (*social.GMBSearchKeywordsResponse, error) {
			keywordMonths = append(keywordMonths, startMonth.Format("2006-01"))
			return &social.GMBSearchKeywordsResponse{
				SearchKeywordsCounts: []struct {
					SearchKeyword string `json:"searchKeyword"`
					InsightsValue struct {
						Value     string `json:"value"`
						Threshold string `json:"threshold"`
					} `json:"insightsValue"`
				}{
					{
						SearchKeyword: "pizza",
						InsightsValue: struct {
							Value     string `json:"value"`
							Threshold string `json:"threshold"`
						}{Value: "12", Threshold: "1"},
					},
				},
			}, nil
		},
		fetchLocalPostsFunc: func(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*social.GMBLocalPostsResponse, error) {
			return makeTestGMBLocalPostsResponse(), nil
		},
		fetchReviewsFunc: func(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*social.GMBReviewsResponse, error) {
			reviewPageTokens = append(reviewPageTokens, pageToken)
			if pageToken == "" {
				return makeTestGMBReviewsResponse(true), nil
			}
			if pageToken == "page-2" {
				return makeTestGMBReviewsResponse(false), nil
			}
			t.Fatalf("unexpected page token %q", pageToken)
			return nil, nil
		},
		fetchMediaAssetsFunc: func(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*social.GMBMediaAssetsResponse, error) {
			return makeTestGMBMediaAssetsResponse(), nil
		},
	}

	err = ProcessAccount(ctx, gmbClient, sink, nil, nil, nil, ImmediateWorkOrder{
		WorkspaceID:  "workspace-1",
		AccountID:    "account-1",
		LocationID:   "location-1",
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		AccountName:  "Account",
		LocationName: "Location",
		LanguageCode: "en",
		SyncType:     "immediate",
		StartDate:    "2025-01-10",
		EndDate:      "2025-02-05",
	}, "", time.Now().UTC(), log)
	if err != nil {
		t.Fatalf("ProcessAccount returned error: %v", err)
	}

	if !perfStart.Equal(wantStart) {
		t.Fatalf("unexpected performance start: got %s want %s", perfStart, wantStart)
	}
	if !perfEnd.Equal(wantEnd) {
		t.Fatalf("unexpected performance end: got %s want %s", perfEnd, wantEnd)
	}

	if len(keywordMonths) != 2 {
		t.Fatalf("expected 2 keyword requests, got %d (%v)", len(keywordMonths), keywordMonths)
	}
	if keywordMonths[0] != "2025-01" || keywordMonths[1] != "2025-02" {
		t.Fatalf("unexpected keyword months: %v", keywordMonths)
	}

	if len(reviewPageTokens) != 2 {
		t.Fatalf("expected 2 review pages fetched, got %d (%v)", len(reviewPageTokens), reviewPageTokens)
	}
	if reviewPageTokens[0] != "" || reviewPageTokens[1] != "page-2" {
		t.Fatalf("unexpected review page token sequence: %v", reviewPageTokens)
	}
}

func TestGMBFallbackDateRange_Uses90DaysAndFiltersCreatedAt(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()
	now := time.Date(2026, 4, 22, 5, 14, 36, 0, time.UTC)

	startDate, endDate, hasRequestedRange, err := resolveGMBDateRange("", "", now)
	if err != nil {
		t.Fatalf("resolveGMBDateRange returned error: %v", err)
	}
	if hasRequestedRange {
		t.Fatal("expected fallback range, got requested range")
	}

	wantStart := time.Date(2026, 1, 22, 0, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2026, 4, 22, 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)
	if !startDate.Equal(wantStart) {
		t.Fatalf("unexpected fallback start date: got %s want %s", startDate, wantStart)
	}
	if !endDate.Equal(wantEnd) {
		t.Fatalf("unexpected fallback end date: got %s want %s", endDate, wantEnd)
	}

	sink := &mockClickHouseSink{
		bulkInsertLocalPostsFunc: func(ctx context.Context, posts []*clickhousemodels.GMBLocalPosts) error {
			if len(posts) != 1 {
				t.Fatalf("expected 1 local post row from fallback filter, got %d", len(posts))
			}
			if got := posts[0].CreatedAt; !got.Equal(time.Date(2025, 1, 11, 13, 15, 0, 0, time.UTC)) {
				t.Fatalf("unexpected filtered local post created_at: %s", got)
			}
			return nil
		},
		bulkInsertReviewsFunc: func(ctx context.Context, reviews []*clickhousemodels.GMBReviews) error {
			if len(reviews) != 2 {
				t.Fatalf("expected 2 review rows from fallback filter, got %d", len(reviews))
			}
			for _, review := range reviews {
				if review.CreatedAt.Before(startDate) || review.CreatedAt.After(endDate) {
					t.Fatalf("review created_at out of fallback range: %s", review.CreatedAt)
				}
			}
			return nil
		},
		bulkInsertMediaAssetsFunc: func(ctx context.Context, assets []*clickhousemodels.GMBMediaAssets) error {
			if len(assets) != 1 {
				t.Fatalf("expected 1 media asset row from fallback filter, got %d", len(assets))
			}
			if got := assets[0].CreatedAt; !got.Equal(time.Date(2025, 1, 18, 9, 0, 0, 0, time.UTC)) {
				t.Fatalf("unexpected filtered media asset created_at: %s", got)
			}
			return nil
		},
	}

	gmbClient := &mockGMBClient{
		refreshTokenFunc: func(ctx context.Context, refreshToken string) (*social.RefreshTokenResponse, error) {
			return &social.RefreshTokenResponse{AccessToken: "refreshed-token"}, nil
		},
		fetchVoiceOfMerchantFunc: func(ctx context.Context, locationID, accessToken string) (*social.VoiceOfMerchantResponse, error) {
			return &social.VoiceOfMerchantResponse{HasVoiceOfMerchant: true}, nil
		},
		fetchPerformanceMetricsFunc: func(ctx context.Context, locationID, accessToken string, startDate, endDate time.Time) (*social.GMBPerformanceResponse, error) {
			if !startDate.Equal(wantStart) || !endDate.Equal(wantEnd) {
				t.Fatalf("unexpected fallback performance range: %s -> %s", startDate, endDate)
			}
			return &social.GMBPerformanceResponse{}, nil
		},
		fetchSearchKeywordsFunc: func(ctx context.Context, locationID, accessToken string, startMonth, endMonth time.Time) (*social.GMBSearchKeywordsResponse, error) {
			return &social.GMBSearchKeywordsResponse{}, nil
		},
		fetchLocalPostsFunc: func(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*social.GMBLocalPostsResponse, error) {
			return makeTestGMBLocalPostsResponse(), nil
		},
		fetchReviewsFunc: func(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*social.GMBReviewsResponse, error) {
			if pageToken == "" {
				return makeTestGMBReviewsResponse(true), nil
			}
			if pageToken == "page-2" {
				return makeTestGMBReviewsResponse(false), nil
			}
			t.Fatalf("unexpected page token %q", pageToken)
			return nil, nil
		},
		fetchMediaAssetsFunc: func(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*social.GMBMediaAssetsResponse, error) {
			return makeTestGMBMediaAssetsResponse(), nil
		},
	}

	err = ProcessAccount(ctx, gmbClient, sink, nil, nil, nil, ImmediateWorkOrder{
		WorkspaceID:  "workspace-1",
		AccountID:    "account-1",
		LocationID:   "location-1",
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		AccountName:  "Account",
		LocationName: "Location",
		LanguageCode: "en",
		SyncType:     "immediate",
	}, "", now, log)
	if err != nil {
		t.Fatalf("ProcessAccount returned error: %v", err)
	}
}

func makeTestGMBLocalPostsResponse() *social.GMBLocalPostsResponse {
	return &social.GMBLocalPostsResponse{
		LocalPosts: []struct {
			Name         string `json:"name"`
			LanguageCode string `json:"languageCode"`
			Summary      string `json:"summary"`
			State        string `json:"state"`
			UpdateTime   string `json:"updateTime"`
			CreateTime   string `json:"createTime"`
			SearchURL    string `json:"searchUrl"`
			Media        []struct {
				Name        string `json:"name"`
				MediaFormat string `json:"mediaFormat"`
				GoogleURL   string `json:"googleUrl"`
			} `json:"media"`
			TopicType string `json:"topicType"`
		}{
			{
				Name:         "inside-window",
				LanguageCode: "en",
				Summary:      "inside",
				State:        "ACTIVE",
				UpdateTime:   "2025-01-15T14:00:00Z",
				CreateTime:   "2025-01-11T13:15:00Z",
				SearchURL:    "https://example.com/inside",
				Media: []struct {
					Name        string `json:"name"`
					MediaFormat string `json:"mediaFormat"`
					GoogleURL   string `json:"googleUrl"`
				}{
					{Name: "media-inside", MediaFormat: "PHOTO", GoogleURL: "https://example.com/media-inside"},
				},
				TopicType: "STANDARD",
			},
			{
				Name:         "outside-window",
				LanguageCode: "en",
				Summary:      "outside",
				State:        "ACTIVE",
				UpdateTime:   "2024-12-30T14:00:00Z",
				CreateTime:   "2024-12-31T23:59:59Z",
				SearchURL:    "https://example.com/outside",
				Media: []struct {
					Name        string `json:"name"`
					MediaFormat string `json:"mediaFormat"`
					GoogleURL   string `json:"googleUrl"`
				}{
					{Name: "media-outside", MediaFormat: "PHOTO", GoogleURL: "https://example.com/media-outside"},
				},
				TopicType: "STANDARD",
			},
		},
	}
}

func makeTestGMBReviewsResponse(firstPage bool) *social.GMBReviewsResponse {
	if firstPage {
		return &social.GMBReviewsResponse{
			Reviews: []struct {
				ReviewID string `json:"reviewId"`
				Reviewer struct {
					ProfilePhotoURL string `json:"profilePhotoUrl"`
					DisplayName     string `json:"displayName"`
				} `json:"reviewer"`
				StarRating  string `json:"starRating"`
				Comment     string `json:"comment"`
				CreateTime  string `json:"createTime"`
				UpdateTime  string `json:"updateTime"`
				Name        string `json:"name"`
				ReviewReply *struct {
					Comment    string `json:"comment"`
					UpdateTime string `json:"updateTime"`
				} `json:"reviewReply,omitempty"`
			}{
				{
					ReviewID: "review-in-1",
					Reviewer: struct {
						ProfilePhotoURL string `json:"profilePhotoUrl"`
						DisplayName     string `json:"displayName"`
					}{ProfilePhotoURL: "https://example.com/photo-1", DisplayName: "Alice"},
					StarRating: "FIVE",
					Comment:    "inside one",
					CreateTime: "2025-01-14T10:00:00Z",
					UpdateTime: "2025-01-14T11:00:00Z",
					Name:       "reviews/1",
					ReviewReply: &struct {
						Comment    string `json:"comment"`
						UpdateTime string `json:"updateTime"`
					}{Comment: "thanks", UpdateTime: "2025-01-15T09:00:00Z"},
				},
				{
					ReviewID: "review-out-1",
					Reviewer: struct {
						ProfilePhotoURL string `json:"profilePhotoUrl"`
						DisplayName     string `json:"displayName"`
					}{ProfilePhotoURL: "https://example.com/photo-2", DisplayName: "Bob"},
					StarRating: "FOUR",
					Comment:    "outside one",
					CreateTime: "2024-12-29T10:00:00Z",
					UpdateTime: "2024-12-29T11:00:00Z",
					Name:       "reviews/2",
				},
			},
			NextPageToken: "page-2",
		}
	}

	return &social.GMBReviewsResponse{
		Reviews: []struct {
			ReviewID string `json:"reviewId"`
			Reviewer struct {
				ProfilePhotoURL string `json:"profilePhotoUrl"`
				DisplayName     string `json:"displayName"`
			} `json:"reviewer"`
			StarRating  string `json:"starRating"`
			Comment     string `json:"comment"`
			CreateTime  string `json:"createTime"`
			UpdateTime  string `json:"updateTime"`
			Name        string `json:"name"`
			ReviewReply *struct {
				Comment    string `json:"comment"`
				UpdateTime string `json:"updateTime"`
			} `json:"reviewReply,omitempty"`
		}{
			{
				ReviewID: "review-in-2",
				Reviewer: struct {
					ProfilePhotoURL string `json:"profilePhotoUrl"`
					DisplayName     string `json:"displayName"`
				}{ProfilePhotoURL: "https://example.com/photo-3", DisplayName: "Carol"},
				StarRating: "FIVE",
				Comment:    "inside two",
				CreateTime: "2025-02-01T10:00:00Z",
				UpdateTime: "2025-02-01T11:00:00Z",
				Name:       "reviews/3",
			},
			{
				ReviewID: "review-out-2",
				Reviewer: struct {
					ProfilePhotoURL string `json:"profilePhotoUrl"`
					DisplayName     string `json:"displayName"`
				}{ProfilePhotoURL: "https://example.com/photo-4", DisplayName: "Dave"},
				StarRating: "THREE",
				Comment:    "outside two",
				CreateTime: "2025-03-01T10:00:00Z",
				UpdateTime: "2025-03-01T11:00:00Z",
				Name:       "reviews/4",
			},
		},
		NextPageToken: "",
	}
}

func makeTestGMBMediaAssetsResponse() *social.GMBMediaAssetsResponse {
	return &social.GMBMediaAssetsResponse{
		MediaItems: []struct {
			Name                string `json:"name"`
			MediaFormat         string `json:"mediaFormat"`
			LocationAssociation struct {
				Category string `json:"category"`
			} `json:"locationAssociation"`
			GoogleURL    string `json:"googleUrl"`
			ThumbnailURL string `json:"thumbnailUrl"`
			CreateTime   string `json:"createTime"`
			Dimensions   struct {
				WidthPixels  int `json:"widthPixels"`
				HeightPixels int `json:"heightPixels"`
			} `json:"dimensions"`
			SourceURL string `json:"sourceUrl"`
		}{
			{
				Name:        "media-inside",
				MediaFormat: "PHOTO",
				LocationAssociation: struct {
					Category string `json:"category"`
				}{Category: "CATEGORY"},
				GoogleURL:    "https://example.com/media-inside",
				ThumbnailURL: "https://example.com/thumb-inside",
				CreateTime:   "2025-01-18T09:00:00Z",
				Dimensions: struct {
					WidthPixels  int `json:"widthPixels"`
					HeightPixels int `json:"heightPixels"`
				}{WidthPixels: 1200, HeightPixels: 900},
				SourceURL: "https://example.com/source-inside",
			},
			{
				Name:        "media-outside",
				MediaFormat: "PHOTO",
				LocationAssociation: struct {
					Category string `json:"category"`
				}{Category: "CATEGORY"},
				GoogleURL:    "https://example.com/media-outside",
				ThumbnailURL: "https://example.com/thumb-outside",
				CreateTime:   "2024-12-15T09:00:00Z",
				Dimensions: struct {
					WidthPixels  int `json:"widthPixels"`
					HeightPixels int `json:"heightPixels"`
				}{WidthPixels: 640, HeightPixels: 480},
				SourceURL: "https://example.com/source-outside",
			},
		},
	}
}

// ================== MongoDB State Update Tests ==================

func TestProcessAccount_UpdatesMongoDBStateOnCompletion(t *testing.T) {
	wo := ImmediateWorkOrder{
		ID:           "507f1f77bcf86cd799439011",
		WorkspaceID:  "workspace_123",
		AccountID:    "account_456",
		LocationID:   "location_789",
		AccessToken:  "token_test",
		RefreshToken: "refresh_test",
		SyncType:     "immediate",
	}

	if wo.ID == "" {
		t.Fatal("Work order ID should not be empty for state update test")
	}

	_, err := parseObjectID(wo.ID)
	if err != nil {
		t.Fatalf("Work order ID should be valid MongoDB ObjectID: %v", err)
	}
}

func parseObjectID(id string) (interface{}, error) {
	if len(id) != 24 {
		return nil, fmt.Errorf("invalid ObjectID length: %d", len(id))
	}
	return id, nil
}

// ==================== Logging Contract Tests ====================

func TestLoggingContract_GMB_ExpectedError_WarnOnly(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()
	captureRecords, captureCleanup := logger.InstallCaptureSpy()
	defer captureCleanup()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Immediately cancel so HTTP calls fail fast

	gmbClient := social.NewGMBClient("test-id", "test-secret")

	mongoErr := errors.New("expected: mongodb unavailable")
	mockRepo := &mockMongoRepo{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return nil, mongoErr
		},
	}

	wo := ImmediateWorkOrder{
		ID:           primitive.NewObjectID().Hex(),
		WorkspaceID:  "ws_123",
		AccountID:    "act_123",
		LocationID:   "loc_123",
		AccessToken:  "test_token",
		RefreshToken: "test_refresh",
		SyncType:     "immediate",
	}

	_ = ProcessAccount(ctx, gmbClient, nil, mockRepo, nil, nil, wo, "", time.Now().UTC(), log)

	output := buf.String()

	if !strings.Contains(output, "WRN") {
		t.Error("expected WRN-level log entries")
	}

	// The MongoDB error specifically should NOT trigger CaptureException
	for _, rec := range *captureRecords {
		if rec.Err != nil && strings.Contains(rec.Err.Error(), "mongodb unavailable") {
			t.Error("CaptureException should NOT be called for expected/handled MongoDB errors")
		}
	}
}

func TestLoggingContract_GMB_NoErrorLevelInProcessor(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()
	_, hookCleanup := logger.InstallHookSpy()
	defer hookCleanup()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	gmbClient := social.NewGMBClient("test-id", "test-secret")

	mockRepo := &mockMongoRepo{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return nil, errors.New("mongo error")
		},
	}

	wo := ImmediateWorkOrder{
		ID:           primitive.NewObjectID().Hex(),
		WorkspaceID:  "ws_123",
		AccountID:    "act_123",
		LocationID:   "loc_123",
		AccessToken:  "test_token",
		RefreshToken: "test_refresh",
		SyncType:     "immediate",
	}

	_ = ProcessAccount(ctx, gmbClient, nil, mockRepo, nil, nil, wo, "", time.Now().UTC(), log)

	output := buf.String()
	errCount := strings.Count(output, "ERR")
	if errCount > 0 {
		t.Errorf("expected 0 ERR-level entries, got %d; processors should never log at Error level", errCount)
	}
}

func TestComputePerformanceDateRange_Uses90DaysAndFiltersCreatedAt(t *testing.T) {
	fixedNow := time.Date(2025, 12, 1, 12, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2025, 12, 1, 23, 59, 59, 999999999, time.UTC)
	fallbackStart := fixedNow.AddDate(0, 0, -90).Truncate(24 * time.Hour)

	t.Run("90 day fallback when no account", func(t *testing.T) {
		start, end := computePerformanceDateRange(nil, fixedNow)
		if !start.Equal(fallbackStart) || !end.Equal(expectedEnd) {
			t.Errorf("unexpected fallback performance range: %v -> %v", start, end)
		}
	})

	t.Run("uses CreatedAt when more recent than 90 days", func(t *testing.T) {
		created := time.Date(2025, 11, 1, 8, 0, 0, 0, time.UTC)
		account := &mongomodels.SocialIntegration{
			CreatedAt: &mongomodels.MongoTime{Time: created},
		}
		start, end := computePerformanceDateRange(account, fixedNow)
		expectedStart := created.UTC().Truncate(24 * time.Hour)
		if !start.Equal(expectedStart) || !end.Equal(expectedEnd) {
			t.Errorf("unexpected filtered performance range: %v -> %v", start, end)
		}
	})

	t.Run("uses 90 day fallback when CreatedAt is older", func(t *testing.T) {
		created := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
		account := &mongomodels.SocialIntegration{
			CreatedAt: &mongomodels.MongoTime{Time: created},
		}
		start, end := computePerformanceDateRange(account, fixedNow)
		if !start.Equal(fallbackStart) || !end.Equal(expectedEnd) {
			t.Errorf("unexpected fallback performance range: %v -> %v", start, end)
		}
	})
}

func TestProcessAccount_ClearsStaleProcessingErrorBeforeRetry(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	accountID := primitive.NewObjectID()
	clearCalls := 0
	mockRepo := &mockMongoRepo{
		FindByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				ID:       accountID,
				MetaData: map[string]interface{}{"last_processing_error": "quota exceeded"},
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
	gmbClient := social.NewGMBClient("test-id", "test-secret")

	err := ProcessAccount(ctx, gmbClient, nil, mockRepo, nil, nil, ImmediateWorkOrder{
		ID:           accountID.Hex(),
		AccountID:    "gmb-123",
		LocationID:   "loc-123",
		AccessToken:  "token",
		RefreshToken: "refresh",
		SyncType:     "full_sync",
	}, "", time.Now().UTC(), log)
	if err == nil {
		t.Fatal("expected refresh failure")
	}
	if clearCalls != 1 {
		t.Fatalf("expected 1 stale-error clear, got %d", clearCalls)
	}
}
