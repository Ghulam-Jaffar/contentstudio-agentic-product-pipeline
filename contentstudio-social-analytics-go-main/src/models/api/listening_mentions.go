package api

import "time"

type MentionFilter struct {
	WorkspaceID        string   `json:"workspace_id"`
	TopicIDs           []string `json:"topic_ids"`
	Platforms          []string `json:"platforms"`
	Sentiments         []string `json:"sentiments"`
	AITags             []string `json:"ai_tags"`
	ExcludeAITags      []string `json:"exclude_ai_tags"`
	Language           []string `json:"language"`
	MinFollowers       int      `json:"min_followers"`
	MinTotalEngagement int      `json:"min_total_engagement"`
	DateFrom           string   `json:"date_from"`
	DateTo             string   `json:"date_to"`
	Sort               string   `json:"sort"`
	Cursor             string   `json:"cursor"`
	Limit              int      `json:"limit"`
	IsBookmarked       *bool    `json:"is_bookmarked"`
	IsRead             *bool    `json:"is_read"`
	IncludeIrrelevant  bool     `json:"include_irrelevant"`
	ViewID             string   `json:"view_id"`
	Search             string   `json:"search"`
}

type MentionResponse struct {
	ID                string           `json:"id"`
	TopicID           string           `json:"topic_id"`
	Platform          string           `json:"platform"`
	AuthorID          string           `json:"author_id"`
	AuthorName        string           `json:"author_name"`
	AuthorHandle      string           `json:"author_handle"`
	AuthorImageURL    string           `json:"author_image_url"`
	AuthorURL         string           `json:"author_url,omitempty"`
	AuthorFollowers   int64            `json:"author_followers"`
	Content           string           `json:"content"`
	URL               string           `json:"url"`
	MediaURLs         []string         `json:"media_urls"`
	Language          string           `json:"language"`
	PublishedAt       time.Time        `json:"published_at"`
	Sentiment         string           `json:"sentiment"`
	SentimentOverride string           `json:"sentiment_override,omitempty"`
	AITags            []string         `json:"ai_tags"`
	Engagement        MentionEngagment `json:"engagement"`
	IsRead            bool             `json:"is_read"`
	IsBookmarked      bool             `json:"is_bookmarked"`
	IsIrrelevant      bool             `json:"is_irrelevant"`
	KeywordMatches    []string         `json:"keyword_matches"`
	ContentType       string           `json:"content_type"`
	MediaType         string           `json:"media_type"`
}

type MentionEngagment struct {
	Total    int64 `json:"total"`
	Likes    int64 `json:"likes"`
	Comments int64 `json:"comments"`
	Shares   int64 `json:"shares"`
}

type MentionListResponse struct {
	Status      bool              `json:"status"`
	Data        []MentionResponse `json:"data"`
	NextCursor  string            `json:"next_cursor"`
	HasMore     bool              `json:"has_more"`
	TotalUnread int               `json:"total_unread"`
}

type UnreadCountResponse struct {
	Status      bool `json:"status"`
	UnreadCount int  `json:"unread_count"`
}

type MentionPatchRequest struct {
	IsRead            *bool  `json:"is_read,omitempty"`
	IsBookmarked      *bool  `json:"is_bookmarked,omitempty"`
	IsIrrelevant      *bool  `json:"is_irrelevant,omitempty"`
	SentimentOverride string `json:"sentiment_override,omitempty"`
	TopicID           string `json:"topic_id"`
}

type MarkAllReadRequest struct {
	WorkspaceID string   `json:"workspace_id"`
	TopicIDs    []string `json:"topic_ids"`
	Platforms   []string `json:"platforms"`
	Sentiments  []string `json:"sentiments"`
}

type ViewFilterPreset struct {
	TopicIDs           []string `json:"topic_ids"             bson:"topic_ids"`
	Platforms          []string `json:"platforms"            bson:"platforms"`
	Sentiments         []string `json:"sentiments"           bson:"sentiments"`
	AITags             []string `json:"ai_tags"              bson:"ai_tags"`
	MinFollowers       int      `json:"min_followers"        bson:"min_followers"`
	MinTotalEngagement int      `json:"min_total_engagement" bson:"min_total_engagement"`
	Language           []string `json:"language"             bson:"language"`
}

type ViewRequest struct {
	WorkspaceID  string           `json:"workspace_id"`
	Name         string           `json:"name"`
	Icon         string           `json:"icon"`
	FilterPreset ViewFilterPreset `json:"filter_preset"`
}

type ViewResponse struct {
	ID           string           `json:"id"`
	WorkspaceID  string           `json:"workspace_id"`
	Name         string           `json:"name"`
	Icon         string           `json:"icon"`
	Type         string           `json:"type"`
	SystemKey    string           `json:"system_key,omitempty"`
	Count        int              `json:"count"`
	FilterPreset ViewFilterPreset `json:"filter_preset"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}
