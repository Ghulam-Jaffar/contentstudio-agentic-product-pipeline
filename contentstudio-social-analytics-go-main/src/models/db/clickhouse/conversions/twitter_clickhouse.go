package conversions

import (
	"context"
	"strconv"
	"time"

	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ConvertTwitterPost converts the parsed Kafka model to ClickHouse model.
func ConvertTwitterPost(p *kafkamodels.ParsedTwitterPost) *chmodels.TwitterPosts {
	if p == nil {
		return nil
	}

	now := time.Now().UTC()

	tweetedAt := parseTwitterTimeOrDefault(p.TweetedAt, now)
	idCreatedAt := parseTwitterTimeOrZero(p.IDCreatedAt)
	authorIDCreated := parseTwitterTimeOrZero(p.AuthorIDCreated)

	return &chmodels.TwitterPosts{
		TwitterID:           p.TwitterID,
		Name:                p.Name,
		Username:            p.Username,
		ProfileImageURL:     p.ProfileImageURL,
		FollowersCount:      p.FollowersCount,
		FollowingCount:      p.FollowingCount,
		TweetCount:          p.TweetCount,
		ListedCount:         p.ListedCount,
		TweetID:             p.TweetID,
		EditHistoryTweetIDs: p.EditHistoryTweetIDs,
		AuthorID:            p.AuthorID,
		AuthorUsername:      p.AuthorUsername,
		IDCreatedAt:         idCreatedAt,
		AuthorIDCreated:     authorIDCreated,
		TweetedAt:           tweetedAt,
		SavingTime:          now,
		Hashtags:            p.Hashtags,
		Permalink:           p.Permalink,
		TweetType:           p.TweetType,
		URLs:                p.URLs,
		MediaURL:            p.MediaURL,
		UsernameMentioned:   p.UsernameMentioned,
		UseridMentioned:     p.UseridMentioned,
		Lang:                p.Lang,
		TweetText:           p.TweetText,
		ImpressionCount:     p.ImpressionCount,
		RetweetCount:        p.RetweetCount,
		ReplyCount:          p.ReplyCount,
		LikeCount:           p.LikeCount,
		BookmarkCount:       p.BookmarkCount,
		QuoteCount:          p.QuoteCount,
		TotalEngagement:     p.TotalEngagement,
		DayOfWeek:           int64(tweetedAt.Weekday()),
		HourOfDay:           int64(tweetedAt.Hour()),
	}
}

// BulkInsertTwitterPosts is a thin wrapper delegating to the ClickHouse client.
func (s *ClickHouseSink) BulkInsertTwitterPosts(ctx context.Context, posts []*chmodels.TwitterPosts) error {
	return s.ClickhouseClient.BulkInsertTwitterPosts(ctx, posts)
}

// ConvertTwitterInsights converts parsed insights to ClickHouse model.
func ConvertTwitterInsights(p *kafkamodels.ParsedTwitterInsights) *chmodels.TwitterInsights {
	if p == nil {
		return nil
	}

	now := time.Now().UTC()

	accountCreatedDate := parseTwitterTimeOrZero(p.AccountCreatedDate)

	savingTime := now
	if p.InsertedAt > 0 {
		savingTime = time.Unix(p.InsertedAt, 0)
	}

	return &chmodels.TwitterInsights{
		TwitterID:          p.TwitterID,
		RecordID:           p.RecordID,
		Name:               p.Name,
		Username:           p.Username,
		ProfileImageURL:    p.ProfileImageURL,
		Description:        p.Description,
		Verified:           strconv.FormatBool(p.Verified),
		AccountCreatedDate: accountCreatedDate,
		FollowersCount:     p.FollowersCount,
		FollowingCount:     p.FollowingCount,
		TweetCount:         p.TweetCount,
		ListedCount:        p.ListedCount,
		LikeCount:          p.LikeCount,
		DayOfWeek:          int64(now.Weekday()),
		SavingTime:         savingTime,
	}
}

// BulkInsertTwitterInsights delegates to the ClickHouse client.
func (s *ClickHouseSink) BulkInsertTwitterInsights(ctx context.Context, insights []*chmodels.TwitterInsights) error {
	return s.ClickhouseClient.BulkInsertTwitterInsights(ctx, insights)
}

func parseTwitterTimeOrDefault(value string, fallback time.Time) time.Time {
	parsed := parseTwitterTimeOrZero(value)
	if parsed.IsZero() {
		return fallback
	}
	return parsed
}

func parseTwitterTimeOrZero(value string) time.Time {
	if value == "" {
		return time.Time{}
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.UTC()
		}
	}

	return time.Time{}
}
