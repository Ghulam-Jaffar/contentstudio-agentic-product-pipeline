package twitter

import (
	"fmt"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
)

const maxLimit = 100

// TwitterRequest is the base request for Twitter analytics endpoints.
type TwitterRequest struct {
	WorkspaceID string `json:"workspace_id"`
	TwitterID   string `json:"twitter_id"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Timezone    string `json:"timezone"`
}

func (r *TwitterRequest) Validate() error {
	if r.WorkspaceID == "" {
		return httputil.NewValidationError("workspace_id is required")
	}
	if r.TwitterID == "" {
		return httputil.NewValidationError("twitter_id is required")
	}
	if r.StartDate == "" {
		return httputil.NewValidationError("start_date is required")
	}
	if r.EndDate == "" {
		return httputil.NewValidationError("end_date is required")
	}

	startDate, err := time.Parse("2006-01-02", r.StartDate)
	if err != nil {
		return httputil.NewValidationError("start_date must be in YYYY-MM-DD format")
	}
	endDate, err := time.Parse("2006-01-02", r.EndDate)
	if err != nil {
		return httputil.NewValidationError("end_date must be in YYYY-MM-DD format")
	}
	if endDate.Before(startDate) {
		return httputil.NewValidationError("end_date cannot be before start_date")
	}
	if r.Timezone != "" {
		if _, err := time.LoadLocation(r.Timezone); err != nil {
			return httputil.NewValidationError("invalid timezone: " + r.Timezone)
		}
	}
	return nil
}

func (r *TwitterRequest) GetTimezone() string {
	if r.Timezone == "" {
		return "UTC"
	}
	return r.Timezone
}

func (r *TwitterRequest) ToQueryParams() (*clickhouse.QueryParams, error) {
	params, err := clickhouse.ParseStartEndDate(r.StartDate, r.EndDate, r.GetTimezone())
	if err != nil {
		return nil, err
	}
	params.AccountIDs = []string{r.TwitterID}
	return params, nil
}

type TweetsRequest struct {
	TwitterRequest
	Limit   int    `json:"limit"`
	OrderBy string `json:"order_by"`
}

func (r *TweetsRequest) GetLimit() int {
	if r.Limit <= 0 {
		return 5
	}
	if r.Limit > maxLimit {
		return maxLimit
	}
	return r.Limit
}

var validTweetOrderFields = map[string]bool{
	"total_engagement": true,
	"like_count":       true,
	"reply_count":      true,
	"quote_count":      true,
	"retweet_count":    true,
	"impression_count": true,
	"bookmark_count":   true,
}

func (r *TweetsRequest) GetOrderBy() string {
	if r.OrderBy == "" || !validTweetOrderFields[r.OrderBy] {
		return "total_engagement"
	}
	return r.OrderBy
}

type MetricsResponse struct {
	Status bool                   `json:"status"`
	Data   map[string]interface{} `json:"data"`
}

type EngagementImpressionResponse struct {
	Status          bool     `json:"status"`
	TwitterID       string   `json:"twitter_id"`
	TweetCount      []int64  `json:"tweet_count"`
	ImpressionCount []int64  `json:"impression_count"`
	TotalEngagement []int64  `json:"total_engagement"`
	RetweetCount    []int64  `json:"retweet_count"`
	ReplyCount      []int64  `json:"reply_count"`
	LikeCount       []int64  `json:"like_count"`
	BookmarkCount   []int64  `json:"bookmark_count"`
	QuoteCount      []int64  `json:"quote_count"`
	TweetedAtDate   []string `json:"tweeted_at_date"`
}

type FollowersTrendResponse struct {
	Status              bool     `json:"status"`
	PlatformID          string   `json:"platform_id"`
	Name                string   `json:"name"`
	Username            string   `json:"username"`
	FollowerCount       []int64  `json:"follower_count"`
	FollowerCountDaily  []int64  `json:"follower_count_daily"`
	FollowingCount      []int64  `json:"following_count"`
	FollowingCountDaily []int64  `json:"following_count_daily"`
	Buckets             []string `json:"buckets"`
}

type Tweet struct {
	ID              string   `json:"id"`
	TweetedAt       string   `json:"tweeted_at"`
	TweetText       string   `json:"tweet_text"`
	TweetType       string   `json:"tweet_type"`
	Permalink       string   `json:"permalink"`
	MediaURL        []string `json:"media_url"`
	ListedCount     int32    `json:"listed_count"`
	RetweetCount    int32    `json:"retweet_count"`
	LikeCount       int32    `json:"like_count"`
	ReplyCount      int32    `json:"reply_count"`
	QuoteCount      int32    `json:"quote_count"`
	BookmarkCount   int32    `json:"bookmark_count"`
	ImpressionCount int32    `json:"impression_count"`
	TotalEngagement int32    `json:"total_engagement"`
}

type TopTweetsResponse struct {
	Status    bool    `json:"status"`
	TopTweets []Tweet `json:"top_tweets"`
}

type LeastTweetsResponse struct {
	Status      bool    `json:"status"`
	LeastTweets []Tweet `json:"least_tweets"`
}

type CreditsUsedData struct {
	CreditsUsed int64  `json:"credits_used"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	WorkspaceID string `json:"workspace_id"`
	TwitterID   string `json:"twitter_id"`
}

type CreditsUsedResponse struct {
	Status bool            `json:"status"`
	Data   CreditsUsedData `json:"data"`
}

func BuildDateTimeRange(startDate, endDate string) (time.Time, time.Time, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start_date: %w", err)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end_date: %w", err)
	}

	return start.UTC(), end.Add(24*time.Hour - time.Second).UTC(), nil
}
