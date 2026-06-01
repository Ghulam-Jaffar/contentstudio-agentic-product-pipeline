package service

import (
	"testing"

	apiModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	clickhouseModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

func TestFetchResult_Struct(t *testing.T) {
	result := &FetchResult{
		Payload: &apiModels.InstagramCompetitorPayload{
			PageID:   "123456",
			PageName: "testuser",
		},
		BusinessDiscovery: nil,
		MediaBatches:      [][]apiModels.InstagramMedia{},
		CurrentState:      "Added",
		TotalFetched:      50,
		TotalFiltered:     5,
		Error:             nil,
	}

	if result.Payload.PageID != "123456" {
		t.Fatalf("expected PageID '123456', got '%s'", result.Payload.PageID)
	}
	if result.TotalFetched != 50 {
		t.Fatalf("expected TotalFetched 50, got %d", result.TotalFetched)
	}
	if result.TotalFiltered != 5 {
		t.Fatalf("expected TotalFiltered 5, got %d", result.TotalFiltered)
	}
	if result.CurrentState != "Added" {
		t.Fatalf("expected CurrentState 'Added', got '%s'", result.CurrentState)
	}
}

func TestParseResult_Struct(t *testing.T) {
	result := &ParseResult{
		Payload: &apiModels.InstagramCompetitorPayload{
			PageID:   "123456",
			PageName: "testuser",
			ReportID: "report123",
		},
		Posts:         []*clickhouseModels.InstagramCompetitorPosts{},
		Insights:      nil,
		ReportID:      "report123",
		CurrentState:  "Processing",
		ProfileImage:  "https://example.com/profile.jpg",
		DataFlag:      true,
		TotalFetched:  25,
		TotalFiltered: 2,
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
	if result.ProfileImage != "https://example.com/profile.jpg" {
		t.Fatalf("unexpected ProfileImage: %s", result.ProfileImage)
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
				PageName:       "testuser",
				Success:        true,
				TotalProcessed: 50,
				Error:          nil,
			},
		},
		{
			name: "failed store",
			result: StoreResult{
				PageID:         "page456",
				PageName:       "faileduser",
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
	if SafeFetchLimit != 25 {
		t.Fatalf("expected SafeFetchLimit 25, got %d", SafeFetchLimit)
	}
}

func TestFetchResult_WithMediaBatches(t *testing.T) {
	media1 := apiModels.InstagramMedia{ID: "media1", MediaType: "IMAGE"}
	media2 := apiModels.InstagramMedia{ID: "media2", MediaType: "VIDEO"}
	batch := []apiModels.InstagramMedia{media1, media2}

	result := &FetchResult{
		Payload: &apiModels.InstagramCompetitorPayload{
			PageID: "123",
		},
		MediaBatches: [][]apiModels.InstagramMedia{batch},
	}

	if len(result.MediaBatches) != 1 {
		t.Fatalf("expected 1 batch, got %d", len(result.MediaBatches))
	}
	if len(result.MediaBatches[0]) != 2 {
		t.Fatalf("expected 2 media items in batch, got %d", len(result.MediaBatches[0]))
	}
}

func TestParseResult_WithPosts(t *testing.T) {
	posts := []*clickhouseModels.InstagramCompetitorPosts{
		{InstagramID: 12345, PostID: "post1"},
		{InstagramID: 12345, PostID: "post2"},
	}

	result := &ParseResult{
		Posts:    posts,
		DataFlag: true,
	}

	if len(result.Posts) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(result.Posts))
	}
}

func TestCompetitorAnalysisService_NewService(t *testing.T) {
	service := NewCompetitorAnalysisService(nil, nil, nil, nil)

	if service == nil {
		t.Fatal("expected non-nil service")
	}
	if service.igClient != nil {
		t.Fatal("expected nil igClient")
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

func TestFetchResult_EmptyMediaBatches(t *testing.T) {
	result := &FetchResult{
		Payload: &apiModels.InstagramCompetitorPayload{
			PageID: "123",
		},
		MediaBatches: [][]apiModels.InstagramMedia{},
	}

	if len(result.MediaBatches) != 0 {
		t.Fatalf("expected 0 batches, got %d", len(result.MediaBatches))
	}
}

func TestParseResult_EmptyPosts(t *testing.T) {
	result := &ParseResult{
		Posts:    []*clickhouseModels.InstagramCompetitorPosts{},
		DataFlag: false,
	}

	if len(result.Posts) != 0 {
		t.Fatalf("expected 0 posts, got %d", len(result.Posts))
	}
	if result.DataFlag {
		t.Fatal("expected DataFlag to be false")
	}
}
