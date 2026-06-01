package api

// CompetitorWorkRequest represents a request to process competitor analysis for a specific competitor and channel.
type CompetitorWorkRequest struct {
	ReportID  string `json:"report_id"` // MongoDB report ID (for email notifications)
	PageID    string `json:"page_id"`   // Facebook/Instagram page ID (for data lookup)
	Channel   string `json:"channel"`   // "facebook" or "instagram"
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
}
