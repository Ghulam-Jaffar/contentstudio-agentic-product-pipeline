package clickhouse

import "time"

// TwitterPosts represents the ClickHouse table schema for Twitter posts analytics.
// Matches the existing twitter_posts ClickHouse table schema.
type TwitterPosts struct {
	TwitterID           string    `ch:"twitter_id" json:"twitter_id"`
	Name                string    `ch:"name" json:"name"`
	Username            string    `ch:"username" json:"username"`
	ProfileImageURL     string    `ch:"profile_image_url" json:"profile_image_url"`
	FollowersCount      int64     `ch:"followers_count" json:"followers_count"`
	FollowingCount      int64     `ch:"following_count" json:"following_count"`
	TweetCount          int64     `ch:"tweet_count" json:"tweet_count"`
	ListedCount         int64     `ch:"listed_count" json:"listed_count"`
	TweetID             string    `ch:"tweet_id" json:"tweet_id"`
	EditHistoryTweetIDs []string  `ch:"edit_history_tweet_ids" json:"edit_history_tweet_ids"`
	AuthorID            string    `ch:"author_id" json:"author_id"`
	AuthorUsername      string    `ch:"author_username" json:"author_username"`
	IDCreatedAt         time.Time `ch:"id_created_at" json:"id_created_at"`
	AuthorIDCreated     time.Time `ch:"author_id_created" json:"author_id_created"`
	TweetedAt           time.Time `ch:"tweeted_at" json:"tweeted_at"`
	SavingTime          time.Time `ch:"saving_time" json:"saving_time"`
	Hashtags            []string  `ch:"hashtags" json:"hashtags"`
	Permalink           string    `ch:"permalink" json:"permalink"`
	TweetType           string    `ch:"tweet_type" json:"tweet_type"`
	URLs                []string  `ch:"urls" json:"urls"`
	MediaURL            []string  `ch:"media_url" json:"media_url"`
	UsernameMentioned   []string  `ch:"username_mentioned" json:"username_mentioned"`
	UseridMentioned     []string  `ch:"userid_mentioned" json:"userid_mentioned"`
	Lang                string    `ch:"lang" json:"lang"`
	TweetText           string    `ch:"tweet_text" json:"tweet_text"`
	ImpressionCount     int64     `ch:"impression_count" json:"impression_count"`
	RetweetCount        int64     `ch:"retweet_count" json:"retweet_count"`
	ReplyCount          int64     `ch:"reply_count" json:"reply_count"`
	LikeCount           int64     `ch:"like_count" json:"like_count"`
	BookmarkCount       int64     `ch:"bookmark_count" json:"bookmark_count"`
	QuoteCount          int64     `ch:"quote_count" json:"quote_count"`
	TotalEngagement     int64     `ch:"total_engagement" json:"total_engagement"`
	DayOfWeek           int64     `ch:"day_of_week" json:"day_of_week"`
	HourOfDay           int64     `ch:"hour_of_day" json:"hour_of_day"`
}

// TwitterInsights represents the ClickHouse table schema for Twitter insights analytics.
// Matches the existing twitter_insights ClickHouse table schema.
type TwitterInsights struct {
	TwitterID          string    `ch:"twitter_id" json:"twitter_id"`
	RecordID           string    `ch:"record_id" json:"record_id"`
	Name               string    `ch:"name" json:"name"`
	Username           string    `ch:"username" json:"username"`
	ProfileImageURL    string    `ch:"profile_image_url" json:"profile_image_url"`
	Description        string    `ch:"description" json:"description"`
	Verified           string    `ch:"verified" json:"verified"`
	AccountCreatedDate time.Time `ch:"account_created_date" json:"account_created_date"`
	FollowersCount     int64     `ch:"followers_count" json:"followers_count"`
	FollowingCount     int64     `ch:"following_count" json:"following_count"`
	TweetCount         int64     `ch:"tweet_count" json:"tweet_count"`
	ListedCount        int64     `ch:"listed_count" json:"listed_count"`
	LikeCount          int64     `ch:"like_count" json:"like_count"`
	DayOfWeek          int64     `ch:"day_of_week" json:"day_of_week"`
	SavingTime         time.Time `ch:"saving_time" json:"saving_time"`
}
