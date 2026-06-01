package parsing

import (
	"strconv"
	"strings"

	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// NormalizeFacebookPostID ensures video post IDs match the composite pageID_postID shape used by posts.
func NormalizeFacebookPostID(pageID, postID string) string {
	postID = strings.TrimSpace(postID)
	if postID == "" {
		return ""
	}
	if pageID == "" {
		return postID
	}
	if strings.HasPrefix(postID, pageID+"_") {
		return postID
	}
	return pageID + "_" + postID
}

// HasFacebookReelsMetric returns true when the raw video payload explicitly contains reels insights.
func HasFacebookReelsMetric(video kafkamodels.RawFacebookVideo) bool {
	for _, insight := range video.VideoInsights.Data {
		if insight.Name == "blue_reels_play_count" {
			return true
		}
	}
	return false
}

// IsFacebookVideoLikePost reports whether a raw Facebook post is actually a video or reel post.
func IsFacebookVideoLikePost(post kafkamodels.RawFacebookPost) bool {
	switch post.StatusType {
	case "added_video", "reels":
		return true
	}

	if post.Attachments != nil {
		for _, attachment := range post.Attachments.Data {
			if looksLikeFacebookVideoAttachment(attachment.Type) || looksLikeFacebookVideoAttachment(attachment.MediaType) {
				return true
			}
			if attachment.Subattachments != nil {
				for _, sub := range attachment.Subattachments.Data {
					if looksLikeFacebookVideoAttachment(sub.Type) || looksLikeFacebookVideoAttachment(sub.MediaType) {
						return true
					}
				}
			}
		}
	}

	for _, attachment := range post.ChildAttachments {
		if looksLikeFacebookVideoAttachment(attachment.Type) || looksLikeFacebookVideoAttachment(attachment.MediaType) {
			return true
		}
	}

	return false
}

// HasMeaningfulFacebookVideoMetrics returns true when the payload carries non-zero video-specific metrics.
// This is used only as a fallback when posts are unavailable for cross-validation.
func HasMeaningfulFacebookVideoMetrics(video kafkamodels.RawFacebookVideo) bool {
	for _, insight := range video.VideoInsights.Data {
		if len(insight.Values) == 0 {
			continue
		}
		if !isMeaningfulFacebookVideoMetric(insight.Name) {
			continue
		}
		if facebookMetricValueAsFloat64(insight.Values[0].Value) > 0 {
			return true
		}
	}
	return false
}

// FilterFacebookVideos keeps only reels or videos that can be validated against fetched posts.
// If posts are unavailable, it falls back to non-zero video metrics to avoid persisting obviously bad rows.
func FilterFacebookVideos(pageID string, posts []kafkamodels.RawFacebookPost, videos []kafkamodels.RawFacebookVideo) ([]kafkamodels.RawFacebookVideo, int) {
	if len(videos) == 0 {
		return nil, 0
	}

	eligiblePostIDs := make(map[string]struct{}, len(posts))
	for _, post := range posts {
		if IsFacebookVideoLikePost(post) {
			eligiblePostIDs[post.ID] = struct{}{}
		}
	}

	filtered := make([]kafkamodels.RawFacebookVideo, 0, len(videos))
	skipped := 0

	for _, video := range videos {
		if HasFacebookReelsMetric(video) {
			filtered = append(filtered, video)
			continue
		}

		normalizedPostID := NormalizeFacebookPostID(pageID, video.PostID)
		if len(posts) > 0 {
			if normalizedPostID != "" {
				if _, ok := eligiblePostIDs[normalizedPostID]; ok {
					filtered = append(filtered, video)
					continue
				}
			}
			skipped++
			continue
		}

		if HasMeaningfulFacebookVideoMetrics(video) {
			filtered = append(filtered, video)
			continue
		}

		skipped++
	}

	return filtered, skipped
}

func looksLikeFacebookVideoAttachment(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "video", "video_inline", "video_autoplay", "reels":
		return true
	default:
		return false
	}
}

func isMeaningfulFacebookVideoMetric(name string) bool {
	switch name {
	case "blue_reels_play_count",
		"total_video_impressions",
		"total_video_impressions_unique",
		"total_video_views",
		"total_video_views_unique",
		"total_video_views_autoplayed",
		"total_video_views_clicked_to_play",
		"total_video_complete_views",
		"total_video_complete_views_unique",
		"post_video_avg_time_watched",
		"post_video_view_time",
		"total_video_view_total_time",
		"total_video_view_total_time_organic",
		"total_video_view_total_time_paid":
		return true
	default:
		return false
	}
}

func facebookMetricValueAsFloat64(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}
