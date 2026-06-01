// Package social provides API clients for social networks.
package social

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

const (
	twitterBaseURL = "https://api.twitter.com/2/"
)

// TwitterClient implements OAuth 1.0a User Context authentication for the Twitter API v2.
// It replicates the Python Tweepy Client approach used in the Python analytics pipeline.
type TwitterClient struct {
	httpClient     *http.Client
	baseURL        string
	consumerKey    string
	consumerSecret string
	log            *logger.Logger
}

// NewTwitterClient returns a new Twitter API v2 client with OAuth 1.0a support.
func NewTwitterClient(consumerKey, consumerSecret string) *TwitterClient {
	return &TwitterClient{
		httpClient:     &http.Client{Timeout: 30 * time.Second},
		baseURL:        twitterBaseURL,
		consumerKey:    consumerKey,
		consumerSecret: consumerSecret,
		log:            logger.New("info"),
	}
}

// tweetFields are the fields requested for tweets (matches Python Tweepy v2 implementation).
var tweetFields = []string{
	"lang", "created_at", "author_id", "conversation_id",
	"public_metrics", "context_annotations", "attachments",
	"entities", "referenced_tweets", "edit_history_tweet_ids",
}

// userFields are the fields requested for user lookups (matches Python Tweepy v2 implementation).
var userFields = []string{
	"created_at", "description", "entities", "id", "location",
	"name", "pinned_tweet_id", "profile_image_url", "protected",
	"public_metrics", "url", "username", "verified", "verified_type", "withheld",
}

// mediaFields are the fields requested for media expansions.
var mediaFields = []string{
	"alt_text", "duration_ms", "height", "media_key",
	"preview_image_url", "type", "url", "variants", "width",
}

// expansions are the expansions requested for tweet lookups.
var expansions = []string{
	"author_id", "referenced_tweets.id", "attachments.media_keys",
	"in_reply_to_user_id", "edit_history_tweet_ids",
	"referenced_tweets.id.author_id", "entities.mentions.username",
}

// TwitterUserResponse represents the parsed user info from Twitter API v2.
type TwitterUserResponse struct {
	Data []TwitterUser `json:"data"`
}

// TwitterUserPublicMetrics represents the public metrics for a Twitter user.
type TwitterUserPublicMetrics struct {
	FollowersCount int64 `json:"followers_count"`
	FollowingCount int64 `json:"following_count"`
	TweetCount     int64 `json:"tweet_count"`
	ListedCount    int64 `json:"listed_count"`
	LikeCount      int64 `json:"like_count"`
}

// TwitterUser represents a Twitter user profile.
type TwitterUser struct {
	ID              string                   `json:"id"`
	Name            string                   `json:"name"`
	Username        string                   `json:"username"`
	Description     string                   `json:"description"`
	ProfileImageURL string                   `json:"profile_image_url"`
	Verified        bool                     `json:"verified"`
	VerifiedType    string                   `json:"verified_type"`
	Protected       bool                     `json:"protected"`
	Location        string                   `json:"location"`
	URL             string                   `json:"url"`
	CreatedAt       string                   `json:"created_at"`
	PublicMetrics   TwitterUserPublicMetrics `json:"public_metrics"`
	PinnedTweetID   string                   `json:"pinned_tweet_id"`
}

// TwitterTweetsResponse represents the response from the users/:id/tweets endpoint.
type TwitterTweetsResponse struct {
	Data     []TwitterTweet   `json:"data"`
	Includes *TwitterIncludes `json:"includes,omitempty"`
	Meta     *TwitterMeta     `json:"meta,omitempty"`
	Errors   []TwitterError   `json:"errors,omitempty"`
}

// TwitterTweet represents a single tweet from the API.
type TwitterTweet struct {
	ID                  string                   `json:"id"`
	Text                string                   `json:"text"`
	AuthorID            string                   `json:"author_id"`
	ConversationID      string                   `json:"conversation_id"`
	CreatedAt           string                   `json:"created_at"`
	Lang                string                   `json:"lang"`
	EditHistoryTweetIDs []string                 `json:"edit_history_tweet_ids"`
	PublicMetrics       TwitterTweetMetrics      `json:"public_metrics"`
	Entities            *TwitterEntities         `json:"entities,omitempty"`
	Attachments         *TwitterAttachments      `json:"attachments,omitempty"`
	ReferencedTweets    []TwitterReferenceTweet  `json:"referenced_tweets,omitempty"`
	ContextAnnotations  []map[string]interface{} `json:"context_annotations,omitempty"`
}

// TwitterTweetMetrics represents tweet engagement metrics.
type TwitterTweetMetrics struct {
	ImpressionCount int64 `json:"impression_count"`
	RetweetCount    int64 `json:"retweet_count"`
	ReplyCount      int64 `json:"reply_count"`
	LikeCount       int64 `json:"like_count"`
	BookmarkCount   int64 `json:"bookmark_count"`
	QuoteCount      int64 `json:"quote_count"`
}

// TwitterEntities represents tweet entities (hashtags, mentions, urls).
type TwitterEntities struct {
	Hashtags []struct {
		Tag string `json:"tag"`
	} `json:"hashtags,omitempty"`
	Mentions []struct {
		Username string `json:"username"`
		ID       string `json:"id"`
	} `json:"mentions,omitempty"`
	URLs []struct {
		URL         string `json:"url"`
		ExpandedURL string `json:"expanded_url"`
		DisplayURL  string `json:"display_url"`
	} `json:"urls,omitempty"`
}

// TwitterAttachments represents tweet attachments.
type TwitterAttachments struct {
	MediaKeys []string `json:"media_keys,omitempty"`
	PollIDs   []string `json:"poll_ids,omitempty"`
}

// TwitterReferenceTweet represents a referenced tweet (retweet, quote, reply).
type TwitterReferenceTweet struct {
	Type string `json:"type"` // "retweeted", "quoted", "replied_to"
	ID   string `json:"id"`
}

// TwitterIncludes contains expanded objects from the API response.
type TwitterIncludes struct {
	Users  []TwitterUser  `json:"users,omitempty"`
	Media  []TwitterMedia `json:"media,omitempty"`
	Tweets []TwitterTweet `json:"tweets,omitempty"`
}

// TwitterMedia represents a media object from the API response.
type TwitterMedia struct {
	MediaKey        string `json:"media_key"`
	Type            string `json:"type"` // "photo", "video", "animated_gif"
	URL             string `json:"url,omitempty"`
	PreviewImageURL string `json:"preview_image_url,omitempty"`
	AltText         string `json:"alt_text,omitempty"`
	Height          int    `json:"height,omitempty"`
	Width           int    `json:"width,omitempty"`
	DurationMs      int    `json:"duration_ms,omitempty"`
}

// TwitterMeta holds pagination info from the API response.
type TwitterMeta struct {
	NextToken   string `json:"next_token,omitempty"`
	ResultCount int    `json:"result_count"`
	NewestID    string `json:"newest_id,omitempty"`
	OldestID    string `json:"oldest_id,omitempty"`
}

// TwitterError represents an error from the API response.
type TwitterError struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Type   string `json:"type"`
}

// doWithRetry executes a request built by makeReq with up to 3 attempts and exponential backoff.
// makeReq is called fresh on every attempt so OAuth 1.0a signatures (which embed a timestamp+nonce)
// are regenerated rather than reused.
func (c *TwitterClient) doWithRetry(ctx context.Context, caller string, makeReq func() (*http.Request, error)) ([]byte, int, error) {
	const maxAttempts = 3

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := makeReq()
		if err != nil {
			return nil, 0, err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if attempt == maxAttempts {
				return nil, 0, fmt.Errorf("TwitterClient.%s: request failed: %w", caller, err)
			}
			select {
			case <-ctx.Done():
				return nil, 0, ctx.Err()
			case <-time.After(time.Duration(attempt) * 500 * time.Millisecond):
			}
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// Never retry auth / rate-limit errors
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusTooManyRequests {
			return body, resp.StatusCode, nil
		}
		if resp.StatusCode == http.StatusOK {
			return body, resp.StatusCode, nil
		}
		// Retry server errors
		if resp.StatusCode >= 500 && attempt < maxAttempts {
			select {
			case <-ctx.Done():
				return nil, 0, ctx.Err()
			case <-time.After(time.Duration(attempt) * 500 * time.Millisecond):
			}
			continue
		}

		return body, resp.StatusCode, nil
	}

	return nil, 0, fmt.Errorf("TwitterClient.%s: max retries exceeded", caller)
}

// FetchUserTweets fetches tweets for a Twitter user using OAuth 1.0a User Context.
// maxResults should be between 5 and 100 (API constraint).
// paginationToken is the cursor for pagination (empty for first page).
func (c *TwitterClient) FetchUserTweets(ctx context.Context, userID, oauthToken, oauthTokenSecret string, maxResults int, paginationToken string) (*TwitterTweetsResponse, error) {
	if maxResults <= 0 || maxResults > 100 {
		maxResults = 40 // default from Python
	}

	endpoint := fmt.Sprintf("%susers/%s/tweets", c.baseURL, userID)

	params := url.Values{}
	params.Set("tweet.fields", strings.Join(tweetFields, ","))
	params.Set("user.fields", strings.Join(userFields, ","))
	params.Set("media.fields", strings.Join(mediaFields, ","))
	params.Set("expansions", strings.Join(expansions, ","))
	params.Set("max_results", strconv.Itoa(maxResults))
	if paginationToken != "" {
		params.Set("pagination_token", paginationToken)
	}

	fullURL := endpoint + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("TwitterClient.FetchUserTweets: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Sign with OAuth 1.0a
	c.signRequest(req, oauthToken, oauthTokenSecret)

	start := time.Now()
	body, status, err := c.doWithRetry(ctx, "FetchUserTweets", func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
		if err != nil {
			return nil, fmt.Errorf("TwitterClient.FetchUserTweets: create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		c.signRequest(req, oauthToken, oauthTokenSecret)
		return req, nil
	})
	if err != nil {
		return nil, fmt.Errorf("TwitterClient.FetchUserTweets: http request: %w", err)
	}

	if status == http.StatusTooManyRequests {
		return nil, fmt.Errorf("TwitterClient.FetchUserTweets: twitter api rate limited (429)")
	}
	if status == http.StatusUnauthorized {
		return nil, fmt.Errorf("TwitterClient.FetchUserTweets: twitter api unauthorized (401): invalid or expired token")
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("TwitterClient.FetchUserTweets: twitter api error (status %d): %s", status, string(body))
	}

	var result TwitterTweetsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("TwitterClient.FetchUserTweets: decode response: %w", err)
	}

	elapsed := time.Since(start)
	c.log.Debug().Str("user", userID).Dur("elapsed", elapsed).Int("tweets", len(result.Data)).Msg("FetchUserTweets page")

	return &result, nil
}

// FetchUserInfo fetches user profile information for one or more Twitter user IDs.
// Uses the GET /2/users endpoint with OAuth 1.0a User Context.
func (c *TwitterClient) FetchUserInfo(ctx context.Context, userIDs []string, oauthToken, oauthTokenSecret string) (*TwitterUserResponse, error) {
	if len(userIDs) == 0 {
		return nil, fmt.Errorf("TwitterClient.FetchUserInfo: no user IDs provided")
	}

	endpoint := fmt.Sprintf("%susers", c.baseURL)

	params := url.Values{}
	params.Set("ids", strings.Join(userIDs, ","))
	params.Set("user.fields", strings.Join(userFields, ","))
	params.Set("tweet.fields", strings.Join(tweetFields, ","))

	fullURL := endpoint + "?" + params.Encode()

	body, status, err := c.doWithRetry(ctx, "FetchUserInfo", func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
		if err != nil {
			return nil, fmt.Errorf("TwitterClient.FetchUserInfo: create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		c.signRequest(req, oauthToken, oauthTokenSecret)
		return req, nil
	})
	if err != nil {
		return nil, fmt.Errorf("TwitterClient.FetchUserInfo: http request: %w", err)
	}

	if status == http.StatusTooManyRequests {
		return nil, fmt.Errorf("TwitterClient.FetchUserInfo: twitter api rate limited (429)")
	}
	if status == http.StatusUnauthorized {
		return nil, fmt.Errorf("TwitterClient.FetchUserInfo: twitter api unauthorized (401): invalid or expired token")
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("TwitterClient.FetchUserInfo: twitter api error (status %d): %s", status, string(body))
	}

	var result TwitterUserResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("TwitterClient.FetchUserInfo: decode response: %w", err)
	}

	return &result, nil
}

// signRequest applies OAuth 1.0a signature to an HTTP request.
// Implements HMAC-SHA1 signing per the OAuth 1.0a specification.
func (c *TwitterClient) signRequest(req *http.Request, oauthToken, oauthTokenSecret string) {
	nonce := generateNonce()
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// Collect OAuth params
	oauthParams := map[string]string{
		"oauth_consumer_key":     c.consumerKey,
		"oauth_nonce":            nonce,
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        timestamp,
		"oauth_token":            oauthToken,
		"oauth_version":          "1.0",
	}

	// Collect all params (query string + oauth)
	allParams := url.Values{}
	for k, v := range oauthParams {
		allParams.Set(k, v)
	}
	queryParams := req.URL.Query()
	for k, vs := range queryParams {
		for _, v := range vs {
			allParams.Add(k, v)
		}
	}

	// Sort and encode parameters
	paramStr := sortedParamString(allParams)

	// Build base string
	baseURL := req.URL.Scheme + "://" + req.URL.Host + req.URL.Path
	baseString := strings.ToUpper(req.Method) + "&" + url.QueryEscape(baseURL) + "&" + url.QueryEscape(paramStr)

	// Build signing key
	signingKey := url.QueryEscape(c.consumerSecret) + "&" + url.QueryEscape(oauthTokenSecret)

	// Generate signature
	mac := hmac.New(sha1.New, []byte(signingKey))
	mac.Write([]byte(baseString))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	oauthParams["oauth_signature"] = signature

	// Build Authorization header
	var headerParts []string
	for k, v := range oauthParams {
		headerParts = append(headerParts, fmt.Sprintf(`%s="%s"`, url.QueryEscape(k), url.QueryEscape(v)))
	}
	sort.Strings(headerParts)

	req.Header.Set("Authorization", "OAuth "+strings.Join(headerParts, ", "))
}

// sortedParamString sorts URL parameters alphabetically and encodes them.
func sortedParamString(params url.Values) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var pairs []string
	for _, k := range keys {
		values := params[k]
		sort.Strings(values)
		for _, v := range values {
			pairs = append(pairs, url.QueryEscape(k)+"="+url.QueryEscape(v))
		}
	}
	return strings.Join(pairs, "&")
}

// generateNonce creates a random string for OAuth nonce.
func generateNonce() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// IsExpectedTwitterError checks if the error is a known/expected error (rate limit, unauthorized)
func IsExpectedTwitterError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "rate limited") ||
		strings.Contains(errMsg, "unauthorized") ||
		strings.Contains(errMsg, "401") ||
		strings.Contains(errMsg, "429")
}
