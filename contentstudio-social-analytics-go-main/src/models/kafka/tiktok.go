package kafka

import (
	"encoding/json"
	"time"
)

// TikTokAccountWorkOrder represents a single TikTok account to process.
// Used both standalone and within batch messages.
type TikTokAccountWorkOrder struct {
	ID           string `json:"id"`            // MongoDB _id (hex)
	WorkspaceID  string `json:"workspace_id"`  // Workspace identifier
	TikTokID     string `json:"tiktok_id"`     // TikTok creator ID
	AccessToken  string `json:"access_token"`  // OAuth access token
	RefreshToken string `json:"refresh_token"` // OAuth refresh token
	Scope        string `json:"scope"`         // OAuth scopes
	SyncType     string `json:"sync_type"`     // "incremental" | "full_sync"
}

// TikTokBatchWorkOrder represents a batch of TikTok accounts to process.
// The scheduler produces batch messages to reduce Kafka overhead.
// The fetcher unpacks batches and distributes accounts to worker pools.
type TikTokBatchWorkOrder struct {
	BatchID   string                   `json:"batch_id"`   // Unique batch identifier (UUID)
	SyncType  string                   `json:"sync_type"`  // "incremental" | "full_sync"
	Accounts  []TikTokAccountWorkOrder `json:"accounts"`   // List of accounts in this batch (max 200)
	CreatedAt time.Time                `json:"created_at"` // Batch creation timestamp
}

// RawTikTokPost is the exact JSON blob returned by TikTok APIs. We wrap it in a
// struct to keep type-safety when sending through Kafka.
type RawTikTokPost struct {
	WorkspaceID string          `json:"workspace_id"`
	TikTokID    string          `json:"tiktok_id"`
	Data        json.RawMessage `json:"data"`
}

// ParsedTikTokPost is the normalised structure produced by the parser service
// and consumed downstream by the ClickHouse sink.
// All fields from the TikTok API video/list endpoint are included.
type ParsedTikTokPost struct {
	ID              string    `json:"id"`                  // Video ID from TikTok API
	TikTokID        string    `json:"tiktok_id,omitempty"` // TikTok account/creator ID or OpenID from MongoDB
	WorkspaceID     string    `json:"workspace_id"`
	DisplayName     string    `json:"display_name,omitempty"`
	ProfileLink     string    `json:"profile_link,omitempty"`
	CoverImageURL   string    `json:"cover_image_url,omitempty"`
	ShareURL        string    `json:"share_url,omitempty"`
	PostDescription string    `json:"post_description,omitempty"`
	Hashtags        []string  `json:"hashtags,omitempty"`
	Duration        int64     `json:"duration,omitempty"`
	Height          int64     `json:"height,omitempty"`
	Width           int64     `json:"width,omitempty"`
	Title           string    `json:"title,omitempty"`
	EmbedHTML       string    `json:"embed_html,omitempty"`
	EmbedLink       string    `json:"embed_link,omitempty"`
	LikeCount       int64     `json:"like_count"`
	CommentCount    int64     `json:"comment_count"`
	ShareCount      int64     `json:"share_count"`
	ViewCount       int64     `json:"view_count"`
	EngagementCount int64     `json:"engagement_count,omitempty"`
	EngagementRate  float64   `json:"engagement_rate,omitempty"`
	CreateTime      int64     `json:"create_time"`
	CreatedAt       time.Time `json:"created_at,omitempty"` // Convenience field for time.Time representation
}

// ParsedTikTokInsights represents aggregated account-level metrics.
type ParsedTikTokInsights struct {
	RecordID            string `json:"record_id"`
	TikTokID            string `json:"tiktok_id"` // TikTok account/creator ID or OpenID from MongoDB
	DisplayName         string `json:"display_name"`
	ProfileImage        string `json:"profile_image"`
	TotalFollowerCount  int64  `json:"total_follower_count"`
	TotalFollowingCount int64  `json:"total_following_count"`
	TotalLikeCount      int64  `json:"total_like_count"`
	TotalVideoCount     int64  `json:"total_video_count"`
	TotalVideoViews     int64  `json:"total_video_views"`
	TotalVideoLikes     int64  `json:"total_video_likes"`
	TotalVideoComments  int64  `json:"total_video_comments"`
	TotalVideoShares    int64  `json:"total_video_shares"`
	IsVerified          bool   `json:"is_verified"`
	Bio                 string `json:"bio"`
	ProfileLink         string `json:"profile_link"`
	InsertedAt          int64  `json:"inserted_at"`
}
