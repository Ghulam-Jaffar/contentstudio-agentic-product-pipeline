package clickhouse

import "time"

// InstagramCompetitorInsights represents competitor page insights
type InstagramCompetitorInsights struct {
	RecordID             string            `ch:"record_id"`
	InstagramAccountID   string            `ch:"instagram_account_id"`
	TotalFollowedByCount int64             `ch:"total_followed_by_count"`
	TotalFollowingCount  int64             `ch:"total_following_count"`
	ProfilePictureURL    string            `ch:"profile_picture_url"`
	PageName             string            `ch:"page_name"`
	Metadata             map[string]string `ch:"metadata"`
	InsertedAt           time.Time         `ch:"inserted_at"`
}

// InstagramCompetitorPosts represents competitor posts
type InstagramCompetitorPosts struct {
	InstagramID          int64     `ch:"instagram_id"`
	PostID               string    `ch:"post_id"`
	BusinessAccountID    string    `ch:"business_account_id"`
	TotalFollowedByCount int64     `ch:"total_followed_by_count"`
	TotalFollowingCount  int64     `ch:"total_following_count"`
	Username             string    `ch:"username"`
	Name                 string    `ch:"name"`
	PageCategory         string    `ch:"page_category"`
	ProfilePictureURL    string    `ch:"profile_picture_url"`
	Biography            string    `ch:"biography"`
	Engagement           int64     `ch:"engagement"`
	LikeCount            int64     `ch:"like_count"`
	CommentsCount        int64     `ch:"comments_count"`
	MediaCount           int64     `ch:"media_count"`
	Caption              string    `ch:"caption"`
	MediaType            string    `ch:"media_type"`
	MediaProductType     string    `ch:"media_product_type"`
	MediaURL             string    `ch:"media_url"`
	Permalink            string    `ch:"permalink"`
	Hashtags             []string  `ch:"hashtags"`
	CreatedAt            time.Time `ch:"created_at"`
	InsertedAt           time.Time `ch:"inserted_at"`
}

// InstagramCompetitorMinimalPost stores only the URL-bearing competitor post fields
// needed by the URL refresher job.
type InstagramCompetitorMinimalPost struct {
	InstagramID       int64     `ch:"instagram_id"`
	PostID            string    `ch:"post_id"`
	MediaURL          string    `ch:"media_url"`
	ProfilePictureURL string    `ch:"profile_picture_url"`
	CreatedAt         time.Time `ch:"created_at"`
}

// TableName returns the ClickHouse table name for insights
func (InstagramCompetitorInsights) TableName() string {
	return "instagram_competitor_insights"
}

// TableName returns the ClickHouse table name for posts
func (InstagramCompetitorPosts) TableName() string {
	return "instagram_competitor_posts"
}
