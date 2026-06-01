package clickhouse

import "time"

// TikTokPosts represents the ClickHouse table schema for TikTok posts analytics.
type TikTokPosts struct {
	TikTokID        string    `ch:"tiktok_id" json:"tiktok_id"` // Holds OpenID from MongoDB or TikTok account ID
	DisplayName     string    `ch:"display_name" json:"display_name"`
	ProfileLink     string    `ch:"profile_link" json:"profile_link"`
	PostID          string    `ch:"post_id" json:"post_id"`
	CoverImageURL   string    `ch:"cover_image_url" json:"cover_image_url"`
	ShareURL        string    `ch:"share_url" json:"share_url"`
	PostDescription string    `ch:"post_description" json:"post_description"`
	Hashtags        []string  `ch:"hashtags" json:"hashtags"`
	Duration        int64     `ch:"duration" json:"duration"`
	Height          int64     `ch:"height" json:"height"`
	Width           int64     `ch:"width" json:"width"`
	Title           string    `ch:"title" json:"title"`
	EmbedHTML       string    `ch:"embed_html" json:"embed_html"`
	EmbedLink       string    `ch:"embed_link" json:"embed_link"`
	LikeCount       int64     `ch:"like_count" json:"like_count"`
	CommentCount    int64     `ch:"comments_count" json:"comments_count"`
	ShareCount      int64     `ch:"share_count" json:"share_count"`
	ViewCount       int64     `ch:"view_count" json:"view_count"`
	EngagementCount int64     `ch:"engagement_count" json:"engagement_count"`
	EngagementRate  float64   `ch:"engagement_rate" json:"engagement_rate"`
	CreatedAt       time.Time `ch:"created_at" json:"created_at"`
	InsertedAt      time.Time `ch:"inserted_at" json:"inserted_at"`
}

// TikTokInsights mirrors the Python TiktokInsights table.
type TikTokInsights struct {
	RecordID            string    `ch:"record_id" json:"record_id"`
	TikTokID            string    `ch:"tiktok_id" json:"tiktok_id"` // Holds OpenID from MongoDB or TikTok account ID
	DisplayName         string    `ch:"display_name" json:"display_name"`
	ProfileImage        string    `ch:"profile_image" json:"profile_image"`
	TotalFollowerCount  int64     `ch:"total_follower_count" json:"total_follower_count"`
	TotalFollowingCount int64     `ch:"total_following_count" json:"total_following_count"`
	TotalLikeCount      int64     `ch:"total_like_count" json:"total_like_count"`
	TotalVideoCount     int64     `ch:"total_video_count" json:"total_video_count"`
	TotalVideoViews     int64     `ch:"total_video_views" json:"total_video_views"`
	TotalVideoLikes     int64     `ch:"total_video_likes" json:"total_video_likes"`
	TotalVideoComments  int64     `ch:"total_video_comments" json:"total_video_comments"`
	TotalVideoShares    int64     `ch:"total_video_shares" json:"total_video_shares"`
	IsVerified          bool      `ch:"is_verified" json:"is_verified"`
	Bio                 string    `ch:"bio" json:"bio"`
	ProfileLink         string    `ch:"profile_link" json:"profile_link"`
	InsertedAt          time.Time `ch:"inserted_at" json:"inserted_at"`
}
