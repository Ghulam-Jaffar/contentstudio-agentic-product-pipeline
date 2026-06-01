package api

// InstagramBusinessDiscoveryResponse represents the Instagram Business Discovery API response
type InstagramBusinessDiscoveryResponse struct {
	BusinessDiscovery BusinessDiscovery `json:"business_discovery"`
	ID                string            `json:"id"`
}

// BusinessDiscovery represents the business discovery data
type BusinessDiscovery struct {
	ID                string       `json:"id"`
	IgID              int64        `json:"ig_id"`
	Username          string       `json:"username"`
	Name              string       `json:"name"`
	Biography         string       `json:"biography"`
	ProfilePictureURL string       `json:"profile_picture_url"`
	FollowersCount    int64        `json:"followers_count"`
	FollowsCount      int64        `json:"follows_count"`
	MediaCount        int64        `json:"media_count"`
	Media             *MediaPaging `json:"media,omitempty"`
}

// MediaPaging represents paginated media response
type MediaPaging struct {
	Data   []InstagramMedia `json:"data"`
	Paging *InstagramPaging `json:"paging,omitempty"`
}

// Media represents an Instagram media item (post)
type InstagramMedia struct {
	ID               string    `json:"id"`
	Caption          string    `json:"caption"`
	CommentsCount    int64     `json:"comments_count"`
	LikeCount        int64     `json:"like_count"`
	MediaProductType string    `json:"media_product_type"`
	MediaType        string    `json:"media_type"`
	MediaURL         string    `json:"media_url"`
	Permalink        string    `json:"permalink"`
	Timestamp        string    `json:"timestamp"`
	Children         *Children `json:"children,omitempty"`
}

// Children represents child media items (for carousels)
type Children struct {
	Data []ChildMedia `json:"data"`
}

// ChildMedia represents a child media item
type ChildMedia struct {
	MediaURL string `json:"media_url"`
	ID       string `json:"id"`
}

// Paging represents pagination cursors
type InstagramPaging struct {
	Cursors *Cursors `json:"cursors,omitempty"`
	Next    string   `json:"next,omitempty"`
}

// Cursors represents pagination cursor tokens
type Cursors struct {
	Before string `json:"before,omitempty"`
	After  string `json:"after,omitempty"`
}

// ErrorResponse represents an Instagram API error response
type ErrorResponse struct {
	Error *ErrorDetail `json:"error,omitempty"`
}

// ErrorDetail represents error details
type ErrorDetail struct {
	Message      string `json:"message"`
	Type         string `json:"type"`
	Code         int    `json:"code"`
	ErrorSubcode int    `json:"error_subcode"`
	FBTraceID    string `json:"fbtrace_id"`
}

// SyncMode represents the sync mode type
type SyncMode string

const (
	// SyncModeIncremental for incremental sync
	SyncModeIncremental SyncMode = "incremental"
	// SyncModeFullRefresh for full refresh sync
	SyncModeFullRefresh SyncMode = "full_refresh"
)

// CompetitorPayload represents the payload for processing a competitor
type InstagramCompetitorPayload struct {
	AccessToken string   `json:"access_token"`
	ReportID    string   `json:"report_id"` // MongoDB report ID for notifications
	PageID      string   `json:"page_id"`
	PageName    string   `json:"page_name"`
	DisplayName string   `json:"display_name"`
	BusinessID  string   `json:"business_id"` // Facebook Page ID linked to Instagram account (used for API calls)
	SyncStatus  SyncMode `json:"sync_status"`
	StartDate   string   `json:"start_date,omitempty"`
	EndDate     string   `json:"end_date,omitempty"`
}
