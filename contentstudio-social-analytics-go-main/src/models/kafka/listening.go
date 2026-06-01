package kafka

import (
	"encoding/json"
	"time"
)

// Kafka topic names for the listening pipeline.
const (
	TopicListeningWork     = "listening-work"
	TopicListeningRaw      = "listening-raw"
	TopicListeningParsed   = "listening-parsed"
	TopicListeningEnriched = "listening-enriched"
	TopicListeningDLQ      = "listening-dlq"
)

// ListeningWorkOrder is the trigger message that initiates a fetch cycle for a topic.
type ListeningWorkOrder struct {
	TopicID                  string            `json:"topic_id"`
	WorkspaceID              string            `json:"workspace_id"`
	SuperAdminID             string            `json:"super_admin_id"`
	IncludeKeywords          []string          `json:"include_keywords"`
	ExcludeKeywords          []string          `json:"exclude_keywords"`
	IncludeAny               []string          `json:"include_any"`
	IncludeAll               []string          `json:"include_all"`
	ExactMatch               bool              `json:"exact_match"`
	CaseSensitive            bool              `json:"case_sensitive"`
	IncludeAuthors           []string          `json:"include_authors"`
	ExcludeAuthors           []string          `json:"exclude_authors"`
	Languages                []string          `json:"language"`
	Regions                  []string          `json:"regions"`
	EnabledPlatforms         []string          `json:"enabled_platforms"`
	GlobalExcludedSubreddits []string          `json:"global_excluded_subreddits,omitempty"`
	MentionsLimit            int               `json:"mentions_limit"`
	Cursors                  map[string]string `json:"cursors"`
	// FromDate restricts the search to posts on or after this date.
	// Zero value means no lower bound (fetch all available history).
	// Set to topic.CreatedAt - 90 days for the initial crawl window.
	FromDate time.Time `json:"from_date,omitempty"`
	// ToDate restricts the search to posts on or before this date.
	// Zero value means no upper bound (up to now).
	ToDate time.Time `json:"to_date,omitempty"`
	// SyncType indicates if this is an 'initial' or 'incremental' sync.
	SyncType string `json:"sync_type,omitempty"`
}

// ListeningRawPayload is emitted by the fetcher with raw Data365 API responses.
type ListeningRawPayload struct {
	TopicID   string             `json:"topic_id"`
	Platform  string             `json:"platform"`
	Keyword   string             `json:"keyword"`
	RawData   json.RawMessage    `json:"raw_data"`
	WorkOrder ListeningWorkOrder `json:"work_order"`
}

// ListeningMention is the normalized mention shape used across parser, sentiment, and sink stages.
type ListeningMention struct {
	MentionID       string    `json:"mention_id"`
	TopicID         string    `json:"topic_id"`
	WorkspaceID     string    `json:"workspace_id"`
	SuperAdminID    string    `json:"super_admin_id"`
	MentionsLimit   int       `json:"mentions_limit"`
	Platform        string    `json:"platform"`
	NativeID        string    `json:"native_id"`
	ContentHash     string    `json:"content_hash"`
	AuthorID        string    `json:"author_id"`
	AuthorName      string    `json:"author_name"`
	AuthorHandle    string    `json:"author_handle,omitempty"`
	AuthorImageURL  string    `json:"author_image_url,omitempty"`
	AuthorURL       string    `json:"author_url,omitempty"`
	AuthorFollowers int64     `json:"author_followers,omitempty"`
	PostText        string    `json:"post_text"`
	Language        string    `json:"language,omitempty"`
	AuthorCountry   string    `json:"author_country,omitempty"`
	AITags          []string  `json:"ai_tags,omitempty"`
	PostedAt        time.Time `json:"posted_at"`
	MatchedKeywords []string  `json:"matched_keywords"`
	TotalEngagement int64     `json:"total_engagement"`
	LikesCount      int64     `json:"likes_count"`
	CommentsCount   int64     `json:"comments_count"`
	SharesCount     int64     `json:"shares_count"`
	ContentType     string    `json:"content_type"`
	MediaType       string    `json:"media_type"`
	URL             string    `json:"url"`
	MediaURLs       []string  `json:"media_urls,omitempty"`
	SentimentLabel  string    `json:"sentiment_label"`
	SentimentScore  float64   `json:"sentiment_score"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	// User interaction flags — default false on ingest; updated via API when users act on mentions.
	PostRead          bool   `json:"post_read"`
	PostIrrelevant    bool   `json:"post_irrelevant"`
	Bookmark          bool   `json:"bookmark"`
	SentimentOverride string `json:"sentiment_override,omitempty"`
}

// ListeningDLQMessage wraps a failed pipeline message for the dead-letter queue.
type ListeningDLQMessage struct {
	OriginalTopic string          `json:"original_topic"`
	Stage         string          `json:"stage"`
	Error         string          `json:"error"`
	Payload       json.RawMessage `json:"payload"`
	AttemptCount  int             `json:"attempt_count"`
	Timestamp     time.Time       `json:"timestamp"`
}
