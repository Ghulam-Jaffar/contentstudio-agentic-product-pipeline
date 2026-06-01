package parsing

import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// TwitterParser handles parsing of Twitter API v2 responses into normalised structures.
type TwitterParser struct{}

// NewTwitterParser creates a new Twitter parser instance.
func NewTwitterParser() *TwitterParser {
	return &TwitterParser{}
}

// ParseTweet converts a raw TwitterTweet (from API response) + user info into a ParsedTwitterPost.
// The userInfo provides account-level fields that are stored alongside each tweet.
// The includes object provides media expansion data.
func (p *TwitterParser) ParseTweet(
	tweet social.TwitterTweet,
	userInfo *social.TwitterUser,
	includes *social.TwitterIncludes,
) *kafkamodels.ParsedTwitterPost {
	if tweet.ID == "" {
		return nil
	}

	post := &kafkamodels.ParsedTwitterPost{
		TweetID:             tweet.ID,
		AuthorID:            tweet.AuthorID,
		TweetText:           tweet.Text,
		Lang:                tweet.Lang,
		EditHistoryTweetIDs: tweet.EditHistoryTweetIDs,
		TweetedAt:           tweet.CreatedAt,
		ImpressionCount:     tweet.PublicMetrics.ImpressionCount,
		RetweetCount:        tweet.PublicMetrics.RetweetCount,
		ReplyCount:          tweet.PublicMetrics.ReplyCount,
		LikeCount:           tweet.PublicMetrics.LikeCount,
		BookmarkCount:       tweet.PublicMetrics.BookmarkCount,
		QuoteCount:          tweet.PublicMetrics.QuoteCount,
	}

	// Calculate total engagement
	post.TotalEngagement = post.RetweetCount + post.ReplyCount + post.LikeCount + post.BookmarkCount + post.QuoteCount

	// Populate user info if provided
	if userInfo != nil {
		post.TwitterID = userInfo.ID
		post.Name = userInfo.Name
		post.Username = userInfo.Username
		post.ProfileImageURL = userInfo.ProfileImageURL
		post.FollowersCount = userInfo.PublicMetrics.FollowersCount
		post.FollowingCount = userInfo.PublicMetrics.FollowingCount
		post.TweetCount = userInfo.PublicMetrics.TweetCount
		post.ListedCount = userInfo.PublicMetrics.ListedCount
		post.AuthorUsername = userInfo.Username
		post.IDCreatedAt = userInfo.CreatedAt
		post.AuthorIDCreated = userInfo.CreatedAt
		post.Permalink = fmt.Sprintf("https://twitter.com/%s/status/%s", userInfo.Username, tweet.ID)
	}

	// Determine tweet type from referenced tweets
	post.TweetType = determineTweetType(tweet.ReferencedTweets)

	// Extract hashtags from entities
	if tweet.Entities != nil {
		for _, ht := range tweet.Entities.Hashtags {
			if ht.Tag != "" {
				post.Hashtags = append(post.Hashtags, ht.Tag)
			}
		}
		// Extract mentions
		for _, m := range tweet.Entities.Mentions {
			if m.Username != "" {
				post.UsernameMentioned = append(post.UsernameMentioned, m.Username)
			}
			if m.ID != "" {
				post.UseridMentioned = append(post.UseridMentioned, m.ID)
			}
		}
		// Extract URLs
		for _, u := range tweet.Entities.URLs {
			expanded := u.ExpandedURL
			if expanded == "" {
				expanded = u.URL
			}
			if expanded != "" {
				post.URLs = append(post.URLs, expanded)
			}
		}
	}

	// Extract media URLs from includes
	if includes != nil && tweet.Attachments != nil {
		mediaKeyMap := buildMediaKeyMap(includes.Media)
		for _, mk := range tweet.Attachments.MediaKeys {
			if media, ok := mediaKeyMap[mk]; ok {
				mediaURL := media.URL
				if mediaURL == "" {
					mediaURL = media.PreviewImageURL
				}
				if mediaURL != "" {
					post.MediaURL = append(post.MediaURL, mediaURL)
				}
			}
		}
	}

	return post
}

// GenerateInsights creates account-level insights from a TwitterUser profile.
// The recordID is generated as MD5("{twitter_id}_{date}") matching the Python implementation.
func (p *TwitterParser) GenerateInsights(userInfo *social.TwitterUser) *kafkamodels.ParsedTwitterInsights {
	if userInfo == nil {
		return nil
	}

	now := time.Now().UTC()
	dateStr := fmt.Sprintf("%s_%s", userInfo.ID, now.Format("2006-01-02"))
	hash := md5.Sum([]byte(dateStr))
	recordID := fmt.Sprintf("%x", hash)

	// Set inserted_at to midnight UTC like Python: pendulum.now("UTC").set(hour=0, minute=0, second=0)
	insertedAt := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	return &kafkamodels.ParsedTwitterInsights{
		TwitterID:          userInfo.ID,
		RecordID:           recordID,
		Name:               userInfo.Name,
		Username:           userInfo.Username,
		ProfileImageURL:    userInfo.ProfileImageURL,
		Description:        userInfo.Description,
		Verified:           userInfo.Verified,
		AccountCreatedDate: userInfo.CreatedAt,
		FollowersCount:     userInfo.PublicMetrics.FollowersCount,
		FollowingCount:     userInfo.PublicMetrics.FollowingCount,
		TweetCount:         userInfo.PublicMetrics.TweetCount,
		ListedCount:        userInfo.PublicMetrics.ListedCount,
		LikeCount:          userInfo.PublicMetrics.LikeCount,
		InsertedAt:         insertedAt.Unix(),
	}
}

// determineTweetType returns the tweet type based on referenced tweets.
func determineTweetType(refs []social.TwitterReferenceTweet) string {
	if len(refs) == 0 {
		return "tweet"
	}
	for _, ref := range refs {
		switch strings.ToLower(ref.Type) {
		case "retweeted":
			return "retweet"
		case "quoted":
			return "quote"
		case "replied_to":
			return "reply"
		}
	}
	return "tweet"
}

// buildMediaKeyMap creates a lookup map from media_key to TwitterMedia.
func buildMediaKeyMap(media []social.TwitterMedia) map[string]social.TwitterMedia {
	m := make(map[string]social.TwitterMedia, len(media))
	for _, item := range media {
		m[item.MediaKey] = item
	}
	return m
}
