package api

import (
	"encoding/json"
	"testing"
)

func TestPostThumbs_Struct(t *testing.T) {
	thumbs := PostThumbs{
		PostID:       "page123_post456",
		PostThumbURL: "https://example.com/thumb.jpg",
		Children: []ChildThumb{
			{MediaID: "media1", ThumbURL: "https://example.com/child1.jpg", MediaType: "photo", Type: "photo"},
			{MediaID: "media2", ThumbURL: "https://example.com/child2.jpg", MediaType: "video", Type: "video_inline"},
		},
	}

	if thumbs.PostID != "page123_post456" {
		t.Fatalf("expected PostID 'page123_post456', got %s", thumbs.PostID)
	}
	if thumbs.PostThumbURL != "https://example.com/thumb.jpg" {
		t.Fatalf("expected PostThumbURL, got %s", thumbs.PostThumbURL)
	}
	if len(thumbs.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(thumbs.Children))
	}
}

func TestChildThumb_Struct(t *testing.T) {
	child := ChildThumb{
		MediaID:   "media123",
		ThumbURL:  "https://example.com/thumb.jpg",
		MediaType: "photo",
		Type:      "photo",
	}

	if child.MediaID != "media123" {
		t.Fatalf("expected MediaID 'media123', got %s", child.MediaID)
	}
	if child.MediaType != "photo" {
		t.Fatalf("expected MediaType 'photo', got %s", child.MediaType)
	}
}

func TestPostThumbs_EmptyChildren(t *testing.T) {
	thumbs := PostThumbs{
		PostID:       "page123_post456",
		PostThumbURL: "https://example.com/thumb.jpg",
		Children:     nil,
	}

	if thumbs.Children != nil {
		t.Fatalf("expected nil Children, got %v", thumbs.Children)
	}
}

func TestPostThumbs_MultipleChildren(t *testing.T) {
	children := []ChildThumb{
		{MediaID: "1", ThumbURL: "url1", MediaType: "photo", Type: "photo"},
		{MediaID: "2", ThumbURL: "url2", MediaType: "photo", Type: "photo"},
		{MediaID: "3", ThumbURL: "url3", MediaType: "video", Type: "video_inline"},
		{MediaID: "4", ThumbURL: "url4", MediaType: "photo", Type: "photo"},
	}

	thumbs := PostThumbs{
		PostID:       "carousel_post",
		PostThumbURL: "main_thumb.jpg",
		Children:     children,
	}

	if len(thumbs.Children) != 4 {
		t.Fatalf("expected 4 children, got %d", len(thumbs.Children))
	}

	photoCount := 0
	videoCount := 0
	for _, child := range thumbs.Children {
		if child.MediaType == "photo" {
			photoCount++
		} else if child.MediaType == "video" {
			videoCount++
		}
	}

	if photoCount != 3 {
		t.Fatalf("expected 3 photos, got %d", photoCount)
	}
	if videoCount != 1 {
		t.Fatalf("expected 1 video, got %d", videoCount)
	}
}

func TestFacebookPageDetails_Struct(t *testing.T) {
	page := FacebookPageDetails{
		ID:                "page123",
		AccessToken:       "token456",
		About:             "About this page",
		Name:              "Test Page",
		FanCount:          10000,
		TalkingAboutCount: 500,
		Bio:               "Page bio",
		Category:          "Business",
		Checkins:          1000,
		Emails:            []string{"contact@example.com", "support@example.com"},
		FollowersCount:    9500,
		Link:              "https://facebook.com/testpage",
		WereHereCount:     800,
		Birthday:          "01/15/2020",
		Cover: &Cover{
			ID:     "cover123",
			Source: "https://example.com/cover.jpg",
		},
	}

	if page.ID != "page123" {
		t.Fatalf("expected ID 'page123', got %s", page.ID)
	}
	if page.FanCount != 10000 {
		t.Fatalf("expected FanCount 10000, got %d", page.FanCount)
	}
	if len(page.Emails) != 2 {
		t.Fatalf("expected 2 emails, got %d", len(page.Emails))
	}
	if page.Cover.Source != "https://example.com/cover.jpg" {
		t.Fatalf("expected Cover.Source, got %s", page.Cover.Source)
	}
}

func TestFacebookPageDetails_JSON_Unmarshal(t *testing.T) {
	jsonData := `{
		"id": "page123",
		"name": "Test Page",
		"fan_count": 10000,
		"followers_count": 9500,
		"cover": {
			"id": "cover123",
			"source": "https://example.com/cover.jpg"
		}
	}`

	var page FacebookPageDetails
	err := json.Unmarshal([]byte(jsonData), &page)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if page.ID != "page123" {
		t.Fatalf("expected ID 'page123', got %s", page.ID)
	}
	if page.FanCount != 10000 {
		t.Fatalf("expected FanCount 10000, got %d", page.FanCount)
	}
}

func TestCover_Struct(t *testing.T) {
	cover := Cover{
		ID:     "cover123",
		Source: "https://example.com/cover.jpg",
	}

	if cover.ID != "cover123" {
		t.Fatalf("expected ID 'cover123', got %s", cover.ID)
	}
	if cover.Source != "https://example.com/cover.jpg" {
		t.Fatalf("expected Source, got %s", cover.Source)
	}
}

func TestPicture_Struct(t *testing.T) {
	picture := Picture{
		Data: &PictureData{
			URL: "https://example.com/picture.jpg",
		},
	}

	if picture.Data.URL != "https://example.com/picture.jpg" {
		t.Fatalf("expected URL, got %s", picture.Data.URL)
	}
}

func TestPost_Struct(t *testing.T) {
	post := Post{
		ID:           "post123",
		ParentID:     "parent456",
		Message:      "Test post message",
		CreatedTime:  "2025-01-15T12:00:00+0000",
		PermalinkURL: "https://facebook.com/post/123",
		StatusType:   "added_photos",
		FullPicture:  "https://example.com/picture.jpg",
		From: &From{
			Name: "Test User",
			ID:   "user123",
		},
		Like: &ReactionSummary{
			Summary: &Summary{TotalCount: 100},
		},
		Love: &ReactionSummary{
			Summary: &Summary{TotalCount: 50},
		},
		Haha: &ReactionSummary{
			Summary: &Summary{TotalCount: 25},
		},
		Wow: &ReactionSummary{
			Summary: &Summary{TotalCount: 10},
		},
		Sad: &ReactionSummary{
			Summary: &Summary{TotalCount: 5},
		},
		Angry: &ReactionSummary{
			Summary: &Summary{TotalCount: 2},
		},
		Comments: &CommentSummary{
			Summary: &Summary{TotalCount: 30},
		},
		Shares: &ShareCount{
			Count: 15,
		},
	}

	if post.ID != "post123" {
		t.Fatalf("expected ID 'post123', got %s", post.ID)
	}
	if post.From.Name != "Test User" {
		t.Fatalf("expected From.Name 'Test User', got %s", post.From.Name)
	}
	if post.Like.Summary.TotalCount != 100 {
		t.Fatalf("expected Like count 100, got %d", post.Like.Summary.TotalCount)
	}
	if post.Shares.Count != 15 {
		t.Fatalf("expected Shares count 15, got %d", post.Shares.Count)
	}
}

func TestPost_WithAttachments(t *testing.T) {
	post := Post{
		ID: "post123",
		Attachments: &Attachments{
			Data: []*AttachmentData{
				{
					Type:         "photo",
					Title:        "Photo title",
					Description:  "Photo description",
					UnshimmedURL: "https://example.com/photo",
					Target: &Target{
						ID:  "target123",
						URL: "https://example.com/target",
					},
					Media: &FacebookMedia{
						Image: &MediaImage{
							Src: "https://example.com/image.jpg",
						},
						Source: "https://example.com/source",
					},
					MediaType: "photo",
				},
			},
		},
	}

	if len(post.Attachments.Data) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(post.Attachments.Data))
	}
	if post.Attachments.Data[0].Type != "photo" {
		t.Fatalf("expected attachment type 'photo', got %s", post.Attachments.Data[0].Type)
	}
}

func TestPost_WithChildAttachments(t *testing.T) {
	post := Post{
		ID: "carousel_post",
		ChildAttachments: []*ChildAttachment{
			{
				ID:          "child1",
				Caption:     "Child 1 caption",
				Description: "Child 1 description",
				Link:        "https://example.com/child1",
				Picture:     "https://example.com/child1.jpg",
				CallToAction: &CallToAction{
					Type: "LEARN_MORE",
					Value: &CTAValue{
						Link: "https://example.com/learn",
					},
				},
			},
			{
				ID:      "child2",
				Caption: "Child 2 caption",
				Picture: "https://example.com/child2.jpg",
			},
		},
	}

	if len(post.ChildAttachments) != 2 {
		t.Fatalf("expected 2 child attachments, got %d", len(post.ChildAttachments))
	}
	if post.ChildAttachments[0].CallToAction.Type != "LEARN_MORE" {
		t.Fatalf("expected CTA type 'LEARN_MORE', got %s", post.ChildAttachments[0].CallToAction.Type)
	}
}

func TestSubattachments_Struct(t *testing.T) {
	subattachments := Subattachments{
		Data: []*AttachmentData{
			{Type: "photo", MediaType: "photo"},
			{Type: "video", MediaType: "video"},
		},
	}

	if len(subattachments.Data) != 2 {
		t.Fatalf("expected 2 subattachments, got %d", len(subattachments.Data))
	}
}

func TestPagingResponse_Struct(t *testing.T) {
	response := PagingResponse{
		Data: []Post{
			{ID: "post1", Message: "Message 1"},
			{ID: "post2", Message: "Message 2"},
		},
		Paging: &FacebookPaging{
			Next:     "https://api.facebook.com/next",
			Previous: "https://api.facebook.com/previous",
		},
	}

	if len(response.Data) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(response.Data))
	}
	if response.Paging.Next != "https://api.facebook.com/next" {
		t.Fatalf("expected Paging.Next, got %s", response.Paging.Next)
	}
}

func TestFacebookCompetitorPayload_Struct(t *testing.T) {
	payload := FacebookCompetitorPayload{
		AccessToken: "token123",
		ReportID:    "report456",
		PageID:      "page789",
		PageName:    "Test Page",
		SyncStatus:  SyncModeIncremental,
	}

	if payload.AccessToken != "token123" {
		t.Fatalf("expected AccessToken 'token123', got %s", payload.AccessToken)
	}
	if payload.SyncStatus != SyncModeIncremental {
		t.Fatalf("expected SyncStatus 'incremental', got %s", payload.SyncStatus)
	}
}

func TestFacebookCompetitorPayload_JSON_Unmarshal(t *testing.T) {
	jsonData := `{
		"access_token": "token123",
		"report_id": "report456",
		"page_id": "page789",
		"page_name": "Test Page",
		"sync_status": "full_refresh"
	}`

	var payload FacebookCompetitorPayload
	err := json.Unmarshal([]byte(jsonData), &payload)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if payload.SyncStatus != SyncModeFullRefresh {
		t.Fatalf("expected SyncStatus 'full_refresh', got %s", payload.SyncStatus)
	}
}
