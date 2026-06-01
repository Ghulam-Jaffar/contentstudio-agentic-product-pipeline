package api

import (
	"encoding/json"
	"testing"
)

func TestInstagramBusinessDiscoveryResponse_Struct(t *testing.T) {
	response := InstagramBusinessDiscoveryResponse{
		ID: "ig123",
		BusinessDiscovery: BusinessDiscovery{
			ID:                "bd456",
			IgID:              123456789,
			Username:          "testuser",
			Name:              "Test User",
			Biography:         "Test bio",
			ProfilePictureURL: "https://example.com/profile.jpg",
			FollowersCount:    10000,
			FollowsCount:      500,
			MediaCount:        200,
		},
	}

	if response.ID != "ig123" {
		t.Fatalf("expected ID 'ig123', got %s", response.ID)
	}
	if response.BusinessDiscovery.Username != "testuser" {
		t.Fatalf("expected Username 'testuser', got %s", response.BusinessDiscovery.Username)
	}
	if response.BusinessDiscovery.FollowersCount != 10000 {
		t.Fatalf("expected FollowersCount 10000, got %d", response.BusinessDiscovery.FollowersCount)
	}
}

func TestBusinessDiscovery_WithMedia(t *testing.T) {
	discovery := BusinessDiscovery{
		ID:             "bd123",
		Username:       "testuser",
		FollowersCount: 5000,
		Media: &MediaPaging{
			Data: []InstagramMedia{
				{
					ID:               "media1",
					Caption:          "First post",
					CommentsCount:    50,
					LikeCount:        200,
					MediaProductType: "FEED",
					MediaType:        "IMAGE",
					MediaURL:         "https://example.com/image1.jpg",
					Permalink:        "https://instagram.com/p/abc1",
					Timestamp:        "2025-01-15T12:00:00+0000",
				},
				{
					ID:               "media2",
					Caption:          "Second post",
					CommentsCount:    30,
					LikeCount:        150,
					MediaProductType: "FEED",
					MediaType:        "VIDEO",
					MediaURL:         "https://example.com/video1.mp4",
					Permalink:        "https://instagram.com/p/abc2",
					Timestamp:        "2025-01-14T10:00:00+0000",
				},
			},
			Paging: &InstagramPaging{
				Cursors: &Cursors{
					Before: "cursor_before",
					After:  "cursor_after",
				},
				Next: "https://api.instagram.com/next",
			},
		},
	}

	if len(discovery.Media.Data) != 2 {
		t.Fatalf("expected 2 media items, got %d", len(discovery.Media.Data))
	}
	if discovery.Media.Paging.Next != "https://api.instagram.com/next" {
		t.Fatalf("expected Paging.Next, got %s", discovery.Media.Paging.Next)
	}
}

func TestInstagramMedia_Struct(t *testing.T) {
	media := InstagramMedia{
		ID:               "media123",
		Caption:          "Test caption #test",
		CommentsCount:    25,
		LikeCount:        100,
		MediaProductType: "FEED",
		MediaType:        "CAROUSEL_ALBUM",
		MediaURL:         "https://example.com/image.jpg",
		Permalink:        "https://instagram.com/p/abc123",
		Timestamp:        "2025-01-15T12:00:00+0000",
		Children: &Children{
			Data: []ChildMedia{
				{MediaURL: "https://example.com/child1.jpg", ID: "child1"},
				{MediaURL: "https://example.com/child2.jpg", ID: "child2"},
			},
		},
	}

	if media.ID != "media123" {
		t.Fatalf("expected ID 'media123', got %s", media.ID)
	}
	if media.MediaType != "CAROUSEL_ALBUM" {
		t.Fatalf("expected MediaType 'CAROUSEL_ALBUM', got %s", media.MediaType)
	}
	if len(media.Children.Data) != 2 {
		t.Fatalf("expected 2 children, got %d", len(media.Children.Data))
	}
}

func TestChildren_Struct(t *testing.T) {
	children := Children{
		Data: []ChildMedia{
			{MediaURL: "url1", ID: "1"},
			{MediaURL: "url2", ID: "2"},
			{MediaURL: "url3", ID: "3"},
		},
	}

	if len(children.Data) != 3 {
		t.Fatalf("expected 3 children, got %d", len(children.Data))
	}
}

func TestChildMedia_Struct(t *testing.T) {
	child := ChildMedia{
		MediaURL: "https://example.com/media.jpg",
		ID:       "child123",
	}

	if child.ID != "child123" {
		t.Fatalf("expected ID 'child123', got %s", child.ID)
	}
	if child.MediaURL != "https://example.com/media.jpg" {
		t.Fatalf("expected MediaURL, got %s", child.MediaURL)
	}
}

func TestInstagramPaging_Struct(t *testing.T) {
	paging := InstagramPaging{
		Cursors: &Cursors{
			Before: "before_cursor",
			After:  "after_cursor",
		},
		Next: "https://api.instagram.com/next",
	}

	if paging.Cursors.Before != "before_cursor" {
		t.Fatalf("expected Before 'before_cursor', got %s", paging.Cursors.Before)
	}
	if paging.Next != "https://api.instagram.com/next" {
		t.Fatalf("expected Next, got %s", paging.Next)
	}
}

func TestCursors_Struct(t *testing.T) {
	cursors := Cursors{
		Before: "cursor_before",
		After:  "cursor_after",
	}

	if cursors.Before != "cursor_before" {
		t.Fatalf("expected Before 'cursor_before', got %s", cursors.Before)
	}
	if cursors.After != "cursor_after" {
		t.Fatalf("expected After 'cursor_after', got %s", cursors.After)
	}
}

func TestErrorResponse_Struct(t *testing.T) {
	response := ErrorResponse{
		Error: &ErrorDetail{
			Message:      "Invalid access token",
			Type:         "OAuthException",
			Code:         190,
			ErrorSubcode: 460,
			FBTraceID:    "trace123",
		},
	}

	if response.Error.Message != "Invalid access token" {
		t.Fatalf("expected Message, got %s", response.Error.Message)
	}
	if response.Error.Code != 190 {
		t.Fatalf("expected Code 190, got %d", response.Error.Code)
	}
}

func TestErrorDetail_Struct(t *testing.T) {
	detail := ErrorDetail{
		Message:      "Rate limit exceeded",
		Type:         "OAuthException",
		Code:         4,
		ErrorSubcode: 0,
		FBTraceID:    "trace456",
	}

	if detail.Type != "OAuthException" {
		t.Fatalf("expected Type 'OAuthException', got %s", detail.Type)
	}
}

func TestSyncMode_Constants(t *testing.T) {
	if SyncModeIncremental != "incremental" {
		t.Fatalf("expected SyncModeIncremental 'incremental', got %s", SyncModeIncremental)
	}
	if SyncModeFullRefresh != "full_refresh" {
		t.Fatalf("expected SyncModeFullRefresh 'full_refresh', got %s", SyncModeFullRefresh)
	}
}

func TestInstagramCompetitorPayload_Struct(t *testing.T) {
	payload := InstagramCompetitorPayload{
		AccessToken: "token123",
		ReportID:    "report456",
		PageID:      "page789",
		PageName:    "Test Page",
		DisplayName: "Test Display Name",
		BusinessID:  "business123",
		SyncStatus:  SyncModeIncremental,
	}

	if payload.AccessToken != "token123" {
		t.Fatalf("expected AccessToken 'token123', got %s", payload.AccessToken)
	}
	if payload.BusinessID != "business123" {
		t.Fatalf("expected BusinessID 'business123', got %s", payload.BusinessID)
	}
	if payload.SyncStatus != SyncModeIncremental {
		t.Fatalf("expected SyncStatus 'incremental', got %s", payload.SyncStatus)
	}
}

func TestInstagramBusinessDiscoveryResponse_JSON_Unmarshal(t *testing.T) {
	jsonData := `{
		"business_discovery": {
			"id": "bd123",
			"ig_id": 123456789,
			"username": "testuser",
			"name": "Test User",
			"biography": "Test bio",
			"profile_picture_url": "https://example.com/profile.jpg",
			"followers_count": 10000,
			"follows_count": 500,
			"media_count": 200
		},
		"id": "ig123"
	}`

	var response InstagramBusinessDiscoveryResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if response.ID != "ig123" {
		t.Fatalf("expected ID 'ig123', got %s", response.ID)
	}
	if response.BusinessDiscovery.IgID != 123456789 {
		t.Fatalf("expected IgID 123456789, got %d", response.BusinessDiscovery.IgID)
	}
}

func TestInstagramCompetitorPayload_JSON_Unmarshal(t *testing.T) {
	jsonData := `{
		"access_token": "token123",
		"report_id": "report456",
		"page_id": "page789",
		"page_name": "Test Page",
		"display_name": "Display Name",
		"business_id": "business123",
		"sync_status": "full_refresh"
	}`

	var payload InstagramCompetitorPayload
	err := json.Unmarshal([]byte(jsonData), &payload)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if payload.SyncStatus != SyncModeFullRefresh {
		t.Fatalf("expected SyncStatus 'full_refresh', got %s", payload.SyncStatus)
	}
}
