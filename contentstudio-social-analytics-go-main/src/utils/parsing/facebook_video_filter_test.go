package parsing

import (
	"testing"

	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

func TestNormalizeFacebookPostID(t *testing.T) {
	if got := NormalizeFacebookPostID("1182", "12345"); got != "1182_12345" {
		t.Fatalf("NormalizeFacebookPostID returned %q", got)
	}
	if got := NormalizeFacebookPostID("1182", "1182_12345"); got != "1182_12345" {
		t.Fatalf("NormalizeFacebookPostID preserved composite ID incorrectly: %q", got)
	}
}

func TestIsFacebookVideoLikePost(t *testing.T) {
	videoPost := kafkamodels.RawFacebookPost{
		ID:         "page_post_1",
		StatusType: "added_video",
	}
	if !IsFacebookVideoLikePost(videoPost) {
		t.Fatal("expected added_video post to be treated as video-like")
	}

	imagePost := kafkamodels.RawFacebookPost{
		ID:         "page_post_2",
		StatusType: "added_photos",
		Attachments: &struct {
			Data []struct {
				Type        string `json:"type"`
				MediaType   string `json:"media_type"`
				Caption     string `json:"caption"`
				Description string `json:"description"`
				Link        string `json:"link"`
				Target      *struct {
					ID string `json:"id"`
				} `json:"target"`
				Media *struct {
					Src    string `json:"src,omitempty"`
					Source string `json:"source"`
					Image  *struct {
						Height int    `json:"height"`
						Width  int    `json:"width"`
						Src    string `json:"src"`
						Source string `json:"source"`
					} `json:"image"`
				} `json:"media"`
				Subattachments *struct {
					Data []struct {
						Type      string `json:"type"`
						MediaType string `json:"media_type"`
						Media     *struct {
							Src    string `json:"src"`
							Source string `json:"source"`
							Image  *struct {
								Height int    `json:"height"`
								Width  int    `json:"width"`
								Src    string `json:"src"`
								Source string `json:"source"`
							} `json:"image"`
						} `json:"media"`
					} `json:"data"`
				} `json:"subattachments"`
			} `json:"data"`
		}{
			Data: []struct {
				Type        string `json:"type"`
				MediaType   string `json:"media_type"`
				Caption     string `json:"caption"`
				Description string `json:"description"`
				Link        string `json:"link"`
				Target      *struct {
					ID string `json:"id"`
				} `json:"target"`
				Media *struct {
					Src    string `json:"src,omitempty"`
					Source string `json:"source"`
					Image  *struct {
						Height int    `json:"height"`
						Width  int    `json:"width"`
						Src    string `json:"src"`
						Source string `json:"source"`
					} `json:"image"`
				} `json:"media"`
				Subattachments *struct {
					Data []struct {
						Type      string `json:"type"`
						MediaType string `json:"media_type"`
						Media     *struct {
							Src    string `json:"src"`
							Source string `json:"source"`
							Image  *struct {
								Height int    `json:"height"`
								Width  int    `json:"width"`
								Src    string `json:"src"`
								Source string `json:"source"`
							} `json:"image"`
						} `json:"media"`
					} `json:"data"`
				} `json:"subattachments"`
			}{
				{
					Type:      "photo",
					MediaType: "photo",
				},
			},
		},
	}
	if IsFacebookVideoLikePost(imagePost) {
		t.Fatal("expected added_photos post to be treated as non-video")
	}
}

func TestFilterFacebookVideos_UsesVideoLikePosts(t *testing.T) {
	posts := []kafkamodels.RawFacebookPost{
		{ID: "1182824828435079_1432746388863659", StatusType: "added_photos"},
		{ID: "1182824828435079_2000000000000000", StatusType: "added_video"},
	}
	videos := []kafkamodels.RawFacebookVideo{
		{ID: "video-images", PostID: "1432746388863659"},
		{ID: "video-real", PostID: "2000000000000000"},
	}

	filtered, skipped := FilterFacebookVideos("1182824828435079", posts, videos)
	if skipped != 1 {
		t.Fatalf("expected 1 skipped video, got %d", skipped)
	}
	if len(filtered) != 1 {
		t.Fatalf("expected 1 filtered video, got %d", len(filtered))
	}
	if filtered[0].ID != "video-real" {
		t.Fatalf("expected real video to remain, got %q", filtered[0].ID)
	}
}

func TestFilterFacebookVideos_FallsBackToMetricsWithoutPosts(t *testing.T) {
	videos := []kafkamodels.RawFacebookVideo{
		{
			ID:     "zero-video",
			PostID: "post1",
		},
		{
			ID:     "real-video",
			PostID: "post2",
			VideoInsights: struct {
				Data []struct {
					Name   string `json:"name"`
					Period string `json:"period"`
					Values []struct {
						Value   interface{} `json:"value"`
						EndTime string      `json:"end_time"`
					} `json:"values"`
					Title       string `json:"title"`
					Description string `json:"description"`
				} `json:"data"`
				Paging struct {
					Previous string `json:"previous"`
					Next     string `json:"next"`
				} `json:"paging"`
			}{
				Data: []struct {
					Name   string `json:"name"`
					Period string `json:"period"`
					Values []struct {
						Value   interface{} `json:"value"`
						EndTime string      `json:"end_time"`
					} `json:"values"`
					Title       string `json:"title"`
					Description string `json:"description"`
				}{
					{
						Name: "total_video_views",
						Values: []struct {
							Value   interface{} `json:"value"`
							EndTime string      `json:"end_time"`
						}{
							{Value: 42},
						},
					},
				},
			},
		},
	}

	filtered, skipped := FilterFacebookVideos("1182824828435079", nil, videos)
	if skipped != 1 {
		t.Fatalf("expected 1 skipped video without posts, got %d", skipped)
	}
	if len(filtered) != 1 || filtered[0].ID != "real-video" {
		t.Fatalf("unexpected filtered videos: %+v", filtered)
	}
}
