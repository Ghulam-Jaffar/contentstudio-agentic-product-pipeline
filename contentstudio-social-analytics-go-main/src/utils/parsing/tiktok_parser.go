package parsing

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// TikTokParser handles parsing of TikTok API responses
type TikTokParser struct {
	hashtagRegex *regexp.Regexp
}

// NewTikTokParser creates a new TikTok parser instance
func NewTikTokParser() *TikTokParser {
	return &TikTokParser{
		hashtagRegex: regexp.MustCompile(`#(\w+)`),
	}
}

// ParseUserInfo parses TikTok user information
func (p *TikTokParser) ParseUserInfo(raw json.RawMessage) (*TikTokUserInfo, error) {
	var user TikTokUserInfo
	if err := json.Unmarshal(raw, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// TikTokUserInfo represents user profile data
type TikTokUserInfo struct {
	OpenID          string `json:"open_id"`
	UnionID         string `json:"union_id"`
	AvatarURL       string `json:"avatar_url"`
	AvatarURL100    string `json:"avatar_url_100"`
	AvatarLargeURL  string `json:"avatar_large_url"`
	DisplayName     string `json:"display_name"`
	BioDescription  string `json:"bio_description"`
	ProfileDeepLink string `json:"profile_deep_link"`
	IsVerified      bool   `json:"is_verified"`
	FollowerCount   int64  `json:"follower_count"`
	FollowingCount  int64  `json:"following_count"`
	LikesCount      int64  `json:"likes_count"`
	VideoCount      int64  `json:"video_count"`
}

// ParseVideo parses a single TikTok video
func (p *TikTokParser) ParseVideo(raw json.RawMessage, userInfo *TikTokUserInfo, tiktokID string) (*kafkamodels.ParsedTikTokPost, error) {
	var video struct {
		ID               string `json:"id"`
		CreateTime       int64  `json:"create_time"`
		CoverImageURL    string `json:"cover_image_url"`
		ShareURL         string `json:"share_url"`
		VideoDescription string `json:"video_description"`
		Duration         int64  `json:"duration"`
		Height           int64  `json:"height"`
		Width            int64  `json:"width"`
		Title            string `json:"title"`
		EmbedHTML        string `json:"embed_html"`
		EmbedLink        string `json:"embed_link"`
		LikeCount        int64  `json:"like_count"`
		CommentCount     int64  `json:"comment_count"`
		ShareCount       int64  `json:"share_count"`
		ViewCount        int64  `json:"view_count"`
	}

	if err := json.Unmarshal(raw, &video); err != nil {
		return nil, err
	}

	post := &kafkamodels.ParsedTikTokPost{
		ID:              video.ID,
		TikTokID:        tiktokID,
		DisplayName:     userInfo.DisplayName,
		ProfileLink:     userInfo.ProfileDeepLink,
		CoverImageURL:   video.CoverImageURL,
		ShareURL:        video.ShareURL,
		PostDescription: video.VideoDescription,
		Duration:        video.Duration,
		Height:          video.Height,
		Width:           video.Width,
		Title:           video.Title,
		EmbedHTML:       video.EmbedHTML,
		EmbedLink:       video.EmbedLink,
		LikeCount:       video.LikeCount,
		CommentCount:    video.CommentCount,
		ShareCount:      video.ShareCount,
		ViewCount:       video.ViewCount,
		CreateTime:      video.CreateTime,
	}

	// Extract hashtags from description or title
	text := video.VideoDescription
	if text == "" {
		text = video.Title
	}

	// Extract hashtags
	hashtags := []string{}
	words := strings.Fields(text)
	for _, word := range words {
		if strings.HasPrefix(word, "#") {
			tag := strings.TrimPrefix(word, "#")
			if tag != "" {
				hashtags = append(hashtags, tag)
			}
		}
	}
	post.Hashtags = hashtags

	// Calculate engagement
	post.EngagementCount = video.LikeCount + video.CommentCount + video.ShareCount
	if video.ViewCount > 0 {
		post.EngagementRate = float64(post.EngagementCount) / float64(video.ViewCount)
	}

	return post, nil
}

// GenerateInsights generates account-level insights from user info and aggregated video stats
func (p *TikTokParser) GenerateInsights(userInfo *TikTokUserInfo, tiktokID string, openID string,
	totalViews, totalLikes, totalComments, totalShares int64) *kafkamodels.ParsedTikTokInsights {

	// Generate record ID exactly as Python: md5("{tiktok_id}_{date}")
	now := time.Now().UTC()
	dateStr := fmt.Sprintf("%s_%s", tiktokID, now.Format("2006-01-02"))
	hash := md5.Sum([]byte(dateStr))
	recordID := fmt.Sprintf("%x", hash)

	// Set inserted_at to midnight UTC like Python: pendulum.now("UTC").set(hour=0, minute=0, second=0)
	insertedAt := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	return &kafkamodels.ParsedTikTokInsights{
		RecordID:            recordID,
		TikTokID:            tiktokID, // Uses tiktokID (can be account ID or OpenID from MongoDB)
		DisplayName:         userInfo.DisplayName,
		ProfileImage:        userInfo.AvatarLargeURL,
		TotalFollowerCount:  userInfo.FollowerCount,
		TotalFollowingCount: userInfo.FollowingCount,
		TotalLikeCount:      userInfo.LikesCount,
		TotalVideoCount:     userInfo.VideoCount,
		TotalVideoViews:     totalViews,
		TotalVideoLikes:     totalLikes,
		TotalVideoComments:  totalComments,
		TotalVideoShares:    totalShares,
		IsVerified:          userInfo.IsVerified,
		Bio:                 userInfo.BioDescription,
		ProfileLink:         userInfo.ProfileDeepLink,
		InsertedAt:          insertedAt.Unix(),
	}
}

// ValidateScopes checks if the provided scopes are sufficient for analytics
func ValidateScopes(scopeString string) bool {
	if scopeString == "" {
		return false
	}

	scopes := strings.Split(scopeString, ",")
	requiredScopes := map[string]bool{
		"user.info.basic":   false,
		"user.info.profile": false,
		"user.info.stats":   false,
		"video.list":        false,
	}

	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if _, ok := requiredScopes[scope]; ok {
			requiredScopes[scope] = true
		}
	}

	// Check if all required scopes are present
	for _, present := range requiredScopes {
		if !present {
			return false
		}
	}

	return true
}
