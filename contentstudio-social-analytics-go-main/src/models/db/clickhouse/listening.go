package clickhouse

import "time"

// ListeningMentionRow maps to the listening_mentions ClickHouse table.
type ListeningMentionRow struct {
	MentionID       string    `ch:"mention_id"`
	TopicID         string    `ch:"topic_id"`
	Platform        string    `ch:"platform"`
	NativeID        string    `ch:"native_id"`
	ContentHash     string    `ch:"content_hash"`
	AuthorID        string    `ch:"author_id"`
	AuthorName      string    `ch:"author_name"`
	AuthorHandle    string    `ch:"author_handle"`
	AuthorImageURL  string    `ch:"author_image_url"`
	AuthorURL       string    `ch:"author_url"`
	AuthorFollowers int64     `ch:"author_followers"`
	PostText        string    `ch:"post_text"`
	Language        string    `ch:"language"`
	PostedAt        time.Time `ch:"posted_at"`
	MatchedKeywords []string  `ch:"matched_keywords"`
	TotalEngagement int64     `ch:"total_engagement"`
	LikesCount      int64     `ch:"likes_count"`
	CommentsCount   int64     `ch:"comments_count"`
	SharesCount     int64     `ch:"shares_count"`
	ContentType     string    `ch:"content_type"`
	MediaType       string    `ch:"media_type"`
	URL             string    `ch:"url"`
	MediaURLs       []string  `ch:"media_urls"`
	AITags          []string  `ch:"ai_tags"`
	SentimentLabel  string    `ch:"sentiment_label"`
	SentimentScore  float64   `ch:"sentiment_score"`
	CreatedAt       time.Time `ch:"created_at"`
	UpdatedAt       time.Time `ch:"updated_at"`
	// User interaction flags — stored as Bool (UInt8 internally in ClickHouse).
	// Updated via a new insert with the same sort key + newer updated_at (ReplacingMergeTree).
	PostRead          bool   `ch:"post_read"`
	PostIrrelevant    bool   `ch:"post_irrelevant"`
	Bookmark          bool   `ch:"bookmark"`
	SentimentOverride string `ch:"sentiment_override"`
}
