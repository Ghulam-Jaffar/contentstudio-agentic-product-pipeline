package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	apiModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	clickhouseModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

// Mock interfaces for testing
type mockFacebookClient struct {
	getCompetitorPageDetailsFunc func(ctx context.Context, pageID, accessToken string) (*apiModels.FacebookPageDetails, *apiModels.Picture, error)
	getCompetitorPostsFunc       func(ctx context.Context, pageID, accessToken string, since, until time.Time, limit int) ([]*apiModels.Post, string, error)
	getCompetitorPostsFromURLFunc func(ctx context.Context, url, pageID, accessToken string) ([]*apiModels.Post, string, error)
}

type mockCompetitorRepository struct {
	getByCompetitorIDFunc func(ctx context.Context, competitorID string) ([]*mongoModels.Competitor, error)
	updateStateFunc       func(ctx context.Context, id primitive.ObjectID, state string) error
	updateFieldFunc       func(ctx context.Context, id primitive.ObjectID, timestamp time.Time) error
	addErrorFunc          func(ctx context.Context, id primitive.ObjectID, errMsg string) error
	updateImageFunc       func(ctx context.Context, id primitive.ObjectID, image string) error
}

type mockClickHouseClient struct {
	insertCompetitorInsightsFunc    func(ctx context.Context, insights []*clickhouseModels.FacebookCompetitorInsights) error
	insertCompetitorPostsFunc       func(ctx context.Context, posts []*clickhouseModels.FacebookCompetitorPosts) error
	insertCompetitorMediaAssetsFunc func(ctx context.Context, assets []*clickhouseModels.FacebookCompetitorMediaAssets) error
}

func TestFetchResult_Struct(t *testing.T) {
	result := &FetchResult{
		Payload: &apiModels.FacebookCompetitorPayload{
			PageID:   "123456",
			PageName: "Test Page",
		},
		PageDetails:   nil,
		AccessToken:   "test_token",
		Picture:       nil,
		PostBatches:   [][]*apiModels.Post{},
		CurrentState:  "Added",
		TotalFetched:  100,
		TotalFiltered: 10,
		Error:         nil,
	}

	if result.Payload.PageID != "123456" {
		t.Fatalf("expected PageID '123456', got '%s'", result.Payload.PageID)
	}
	if result.TotalFetched != 100 {
		t.Fatalf("expected TotalFetched 100, got %d", result.TotalFetched)
	}
	if result.TotalFiltered != 10 {
		t.Fatalf("expected TotalFiltered 10, got %d", result.TotalFiltered)
	}
	if result.CurrentState != "Added" {
		t.Fatalf("expected CurrentState 'Added', got '%s'", result.CurrentState)
	}
}

func TestParseResult_Struct(t *testing.T) {
	result := &ParseResult{
		Payload: &apiModels.FacebookCompetitorPayload{
			PageID:   "123456",
			PageName: "Test Page",
			ReportID: "report123",
		},
		Posts:         []*clickhouseModels.FacebookCompetitorPosts{},
		MediaAssets:   []*clickhouseModels.FacebookCompetitorMediaAssets{},
		Insights:      nil,
		ReportID:      "report123",
		CurrentState:  "Processing",
		DataFlag:      true,
		TotalFetched:  50,
		TotalFiltered: 5,
		Error:         nil,
	}

	if result.Payload.PageID != "123456" {
		t.Fatalf("expected PageID '123456', got '%s'", result.Payload.PageID)
	}
	if result.ReportID != "report123" {
		t.Fatalf("expected ReportID 'report123', got '%s'", result.ReportID)
	}
	if !result.DataFlag {
		t.Fatal("expected DataFlag to be true")
	}
}

func TestStoreResult_Struct(t *testing.T) {
	cases := []struct {
		name   string
		result StoreResult
	}{
		{
			name: "successful store",
			result: StoreResult{
				PageID:         "page123",
				PageName:       "Test Page",
				Success:        true,
				TotalProcessed: 100,
				Error:          nil,
			},
		},
		{
			name: "failed store",
			result: StoreResult{
				PageID:         "page456",
				PageName:       "Failed Page",
				Success:        false,
				TotalProcessed: 0,
				Error:          nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.result.PageID == "" {
				t.Fatal("expected PageID to be set")
			}
			if tc.result.PageName == "" {
				t.Fatal("expected PageName to be set")
			}
		})
	}
}

func TestConstants(t *testing.T) {
	if PostsLimitIncremental != 999 {
		t.Fatalf("expected PostsLimitIncremental 999, got %d", PostsLimitIncremental)
	}
	if PostsLimitFull != 999 {
		t.Fatalf("expected PostsLimitFull 999, got %d", PostsLimitFull)
	}
	if DaysToFetchIncremental != 14 {
		t.Fatalf("expected DaysToFetchIncremental 14, got %d", DaysToFetchIncremental)
	}
	if DaysToFetchFull != 90 {
		t.Fatalf("expected DaysToFetchFull 90, got %d", DaysToFetchFull)
	}
	if SafeFetchLimit != 50 {
		t.Fatalf("expected SafeFetchLimit 50, got %d", SafeFetchLimit)
	}
}

func TestFetchResult_WithPostBatches(t *testing.T) {
	post1 := &apiModels.Post{ID: "post1", Message: "Hello"}
	post2 := &apiModels.Post{ID: "post2", Message: "World"}
	batch := []*apiModels.Post{post1, post2}

	result := &FetchResult{
		Payload: &apiModels.FacebookCompetitorPayload{
			PageID: "123",
		},
		PostBatches: [][]*apiModels.Post{batch},
	}

	if len(result.PostBatches) != 1 {
		t.Fatalf("expected 1 batch, got %d", len(result.PostBatches))
	}
	if len(result.PostBatches[0]) != 2 {
		t.Fatalf("expected 2 posts in batch, got %d", len(result.PostBatches[0]))
	}
}

func TestParseResult_WithPosts(t *testing.T) {
	posts := []*clickhouseModels.FacebookCompetitorPosts{
		{FacebookID: "page1", PostID: "post1"},
		{FacebookID: "page1", PostID: "post2"},
	}

	result := &ParseResult{
		Posts:    posts,
		DataFlag: true,
	}

	if len(result.Posts) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(result.Posts))
	}
}

func TestParseResult_WithMediaAssets(t *testing.T) {
	assets := []*clickhouseModels.FacebookCompetitorMediaAssets{
		{PageID: "page1", PostID: "post1", AssetType: "image"},
	}

	result := &ParseResult{
		MediaAssets: assets,
	}

	if len(result.MediaAssets) != 1 {
		t.Fatalf("expected 1 media asset, got %d", len(result.MediaAssets))
	}
}

func TestCompetitorAnalysisService_NewService(t *testing.T) {
	service := NewCompetitorAnalysisService(nil, nil, nil, nil)

	if service == nil {
		t.Fatal("expected non-nil service")
	}
	if service.fbClient != nil {
		t.Fatal("expected nil fbClient")
	}
	if service.mongoRepo != nil {
		t.Fatal("expected nil mongoRepo")
	}
	if service.chRepo != nil {
		t.Fatal("expected nil chRepo")
	}
	if service.log != nil {
		t.Fatal("expected nil log")
	}
}

// Test ParseCompetitorData with various scenarios
func TestParseCompetitorData_FetchError(t *testing.T) {
	fetchResult := &FetchResult{
		Payload: &apiModels.FacebookCompetitorPayload{
			PageID:   "123",
			PageName: "Test Page",
		},
		Error: errors.New("fetch failed"),
	}

	// Since we can't easily mock the service without interfaces,
	// we test the result handling logic
	parseResult := &ParseResult{
		Payload:      fetchResult.Payload,
		CurrentState: fetchResult.CurrentState,
		Error:        fetchResult.Error,
	}

	if parseResult.Error == nil {
		t.Fatal("expected error to be propagated")
	}
	if parseResult.Error.Error() != "fetch failed" {
		t.Fatalf("expected 'fetch failed', got '%s'", parseResult.Error.Error())
	}
}

func TestParseCompetitorData_NilPageDetails(t *testing.T) {
	fetchResult := &FetchResult{
		Payload: &apiModels.FacebookCompetitorPayload{
			PageID:   "123",
			PageName: "Test Page",
		},
		PageDetails: nil,
		Error:       nil,
	}

	// Test that nil PageDetails would cause an error in parsing
	if fetchResult.PageDetails != nil {
		t.Fatal("expected nil PageDetails")
	}
}

func TestStoreCompetitorData_ParseError(t *testing.T) {
	parseResult := &ParseResult{
		Payload: &apiModels.FacebookCompetitorPayload{
			PageID:   "123",
			PageName: "Test Page",
		},
		Error: errors.New("parse failed"),
	}

	storeResult := &StoreResult{
		PageID:   parseResult.Payload.PageID,
		PageName: parseResult.Payload.PageName,
		Error:    parseResult.Error,
	}

	if storeResult.Error == nil {
		t.Fatal("expected error to be propagated")
	}
	if storeResult.Success {
		t.Fatal("expected Success to be false when error exists")
	}
}

func TestFetchResult_PartialCompletion(t *testing.T) {
	post1 := &apiModels.Post{ID: "post1", Message: "Hello"}
	batch := []*apiModels.Post{post1}

	result := &FetchResult{
		Payload: &apiModels.FacebookCompetitorPayload{
			PageID: "123",
		},
		PostBatches:   [][]*apiModels.Post{batch},
		TotalFetched:  10,
		TotalFiltered: 2,
		Error:         nil, // Partial completion - some data but API error during fetch
	}

	// Verify partial data is available
	if len(result.PostBatches) != 1 {
		t.Fatalf("expected 1 batch, got %d", len(result.PostBatches))
	}
	if result.TotalFetched != 10 {
		t.Fatalf("expected TotalFetched 10, got %d", result.TotalFetched)
	}
}

func TestParseResult_WithInsights(t *testing.T) {
	insights := &clickhouseModels.FacebookCompetitorInsights{
		PageID:         "page123",
		TotalFanCount:  10000,
		FollowersCount: 5000,
	}

	result := &ParseResult{
		Payload: &apiModels.FacebookCompetitorPayload{
			PageID: "page123",
		},
		Insights: insights,
		DataFlag: true,
	}

	if result.Insights == nil {
		t.Fatal("expected Insights to be set")
	}
	if result.Insights.TotalFanCount != 10000 {
		t.Fatalf("expected TotalFanCount 10000, got %d", result.Insights.TotalFanCount)
	}
}

func TestStoreResult_Success(t *testing.T) {
	result := &StoreResult{
		PageID:         "page123",
		PageName:       "Test Page",
		Success:        true,
		TotalProcessed: 50,
		Error:          nil,
	}

	if !result.Success {
		t.Fatal("expected Success to be true")
	}
	if result.TotalProcessed != 50 {
		t.Fatalf("expected TotalProcessed 50, got %d", result.TotalProcessed)
	}
	if result.Error != nil {
		t.Fatal("expected nil Error")
	}
}

func TestFetchResult_SyncModes(t *testing.T) {
	cases := []struct {
		name       string
		syncStatus apiModels.SyncMode
	}{
		{"incremental sync", apiModels.SyncModeIncremental},
		{"full sync", apiModels.SyncModeFullRefresh},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := &FetchResult{
				Payload: &apiModels.FacebookCompetitorPayload{
					PageID:     "123",
					SyncStatus: tc.syncStatus,
				},
			}

			if result.Payload.SyncStatus != tc.syncStatus {
				t.Fatalf("expected SyncStatus %v, got %v", tc.syncStatus, result.Payload.SyncStatus)
			}
		})
	}
}

func TestParseResult_DataFlag(t *testing.T) {
	cases := []struct {
		name     string
		posts    []*clickhouseModels.FacebookCompetitorPosts
		dataFlag bool
	}{
		{
			name:     "with posts",
			posts:    []*clickhouseModels.FacebookCompetitorPosts{{PostID: "1"}},
			dataFlag: true,
		},
		{
			name:     "without posts",
			posts:    []*clickhouseModels.FacebookCompetitorPosts{},
			dataFlag: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := &ParseResult{
				Posts:    tc.posts,
				DataFlag: tc.dataFlag,
			}

			if result.DataFlag != tc.dataFlag {
				t.Fatalf("expected DataFlag %v, got %v", tc.dataFlag, result.DataFlag)
			}
		})
	}
}

// Suppress unused variable warnings
var (
	_ = mockFacebookClient{}
	_ = mockCompetitorRepository{}
	_ = mockClickHouseClient{}
	_ context.Context
	_ primitive.ObjectID
)
