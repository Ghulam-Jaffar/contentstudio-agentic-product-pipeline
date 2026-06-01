package kafka

type CompetitorWorkOrder struct {
	ReportID  string `json:"report_id"` // MongoDB report ID (for email notifications)
	PageID    string `json:"page_id"`   // Facebook/Instagram page ID (for data lookup)
	Channel   string `json:"channel"`   // "facebook" or "instagram"
	Mode      string `json:"mode"`      // Processing mode
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
}
