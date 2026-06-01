package kafka

// ImmediateWorkOrder represents the message structure for Kafka topics
type ImmediateWorkOrder struct {
	ID                    string `json:"id"`
	AccountID             string `json:"account_id"` // Platform-specific ID (facebook_id, instagram_id, etc.)
	Type                  string `json:"type"`       // e.g., "Page", "Group", "Business", "Creator"
	AccessToken           string `json:"access_token"`
	WorkspaceID           string `json:"workspace_id"`
	LongAccessToken       string `json:"long_access_token,omitempty"`
	RefreshToken          string `json:"refresh_token,omitempty"`
	SyncType              string `json:"sync_type"` // "immediate"
	ConnectedViaInstagram bool   `json:"connected_via_instagram"`
	StartDate             string `json:"start_date,omitempty"`
	EndDate               string `json:"end_date,omitempty"`
}
