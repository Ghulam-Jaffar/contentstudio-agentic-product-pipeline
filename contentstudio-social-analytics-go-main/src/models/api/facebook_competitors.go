package api

// FacebookPageDetails represents the page details from Facebook API
type FacebookPageDetails struct {
	ID                string   `json:"id"`
	AccessToken       string   `json:"access_token"`
	About             string   `json:"about"`
	Name              string   `json:"name"`
	FanCount          int64    `json:"fan_count"`
	TalkingAboutCount int64    `json:"talking_about_count"`
	Bio               string   `json:"bio"`
	Category          string   `json:"category"`
	Checkins          int64    `json:"checkins"`
	Emails            []string `json:"emails"`
	FollowersCount    int64    `json:"followers_count"`
	Link              string   `json:"link"`
	WereHereCount     int64    `json:"were_here_count"`
	Birthday          string   `json:"birthday"`
	Cover             *Cover   `json:"cover"`
}

// Cover represents Facebook page cover photo
type Cover struct {
	ID     string `json:"id"`
	Source string `json:"source"`
}

// Picture represents Facebook page picture
type Picture struct {
	Data *PictureData `json:"data"`
}

// PictureData represents the picture data
type PictureData struct {
	URL string `json:"url"`
}

// Post represents a Facebook post
type Post struct {
	ID               string             `json:"id"`
	ParentID         string             `json:"parent_id,omitempty"`
	From             *From              `json:"from"`
	Message          string             `json:"message"`
	CreatedTime      string             `json:"created_time"`
	PermalinkURL     string             `json:"permalink_url"`
	StatusType       string             `json:"status_type"`
	FullPicture      string             `json:"full_picture"`
	Attachments      *Attachments       `json:"attachments"`
	ChildAttachments []*ChildAttachment `json:"child_attachments,omitempty"`
	// Reactions
	Like     *ReactionSummary `json:"like"`
	Love     *ReactionSummary `json:"love"`
	Haha     *ReactionSummary `json:"haha"`
	Wow      *ReactionSummary `json:"wow"`
	Sad      *ReactionSummary `json:"sad"`
	Angry    *ReactionSummary `json:"angry"`
	Comments *CommentSummary  `json:"comments"`
	Shares   *ShareCount      `json:"shares"`
}

// From represents the author of a post
type From struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// ReactionSummary represents reaction summary
type ReactionSummary struct {
	Summary *Summary `json:"summary"`
}

// CommentSummary represents comment summary
type CommentSummary struct {
	Summary *Summary `json:"summary"`
}

// Summary represents count summary
type Summary struct {
	TotalCount int64 `json:"total_count"`
}

// ShareCount represents share count
type ShareCount struct {
	Count int64 `json:"count"`
}

// Attachments represents post attachments
type Attachments struct {
	Data []*AttachmentData `json:"data"`
}

// AttachmentData represents attachment data
type AttachmentData struct {
	Type           string          `json:"type"`
	Title          string          `json:"title"`
	Description    string          `json:"description"`
	UnshimmedURL   string          `json:"unshimmed_url"`
	Target         *Target         `json:"target"`
	Media          *FacebookMedia  `json:"media"`
	MediaType      string          `json:"media_type"`
	Subattachments *Subattachments `json:"subattachments"`
}

// ChildAttachment represents child attachment for carousel posts
type ChildAttachment struct {
	ID           string        `json:"id"`
	Caption      string        `json:"caption"`
	Description  string        `json:"description"`
	Link         string        `json:"link"`
	Picture      string        `json:"picture"`
	CallToAction *CallToAction `json:"call_to_action"`
}

// CallToAction represents CTA button
type CallToAction struct {
	Type  string    `json:"type"`
	Value *CTAValue `json:"value"`
}

// CTAValue represents CTA value
type CTAValue struct {
	Link string `json:"link"`
}

// Target represents attachment target
type Target struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// Media represents media content
type FacebookMedia struct {
	Image  *MediaImage `json:"image"`
	Source string      `json:"source"`
}

// MediaImage represents media image
type MediaImage struct {
	Src string `json:"src"`
}

// Subattachments represents subattachments for carousel
type Subattachments struct {
	Data []*AttachmentData `json:"data"`
}

// PagingResponse represents paginated API response
type PagingResponse struct {
	Data   []Post          `json:"data"`
	Paging *FacebookPaging `json:"paging"`
}

// Paging represents pagination info
type FacebookPaging struct {
	Next     string `json:"next"`
	Previous string `json:"previous"`
}

// CompetitorPayload represents the payload for processing a competitor
type FacebookCompetitorPayload struct {
	AccessToken string   `json:"access_token"`
	ReportID    string   `json:"report_id"` // MongoDB report ID for notifications
	PageID      string   `json:"page_id"`
	PageName    string   `json:"page_name"`
	SyncStatus  SyncMode `json:"sync_status"`
	StartDate   string   `json:"start_date,omitempty"`
	EndDate     string   `json:"end_date,omitempty"`
}
