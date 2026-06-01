package api

type ImmediateWorkRequest struct {
	AccountID string `json:"account_id"`
	Channel   string `json:"channel"` // "facebook", "instagram", "linkedin", "tiktok"
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
	NTweets   int    `json:"n_tweets,omitempty"`
}
