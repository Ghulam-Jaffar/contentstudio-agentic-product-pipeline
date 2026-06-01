// Package social contains clients for interacting with social media platform APIs.
package social

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/crypto"
	"golang.org/x/time/rate"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	fbmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/rs/zerolog/log"
)

const (
	// fbAPIVersion is the version of the Graph API we are targeting.
	fbAPIVersion = "v20.0"
	// fbBaseURL is the base URL for all Graph API calls.
	fbBaseURL = "https://graph.facebook.com/"
	// maxPagesToFetch is a safety brake to prevent infinite loops during pagination.
	maxPagesToFetch = 20

	// postFields is the comprehensive list of fields to request for each post,

	postFields = "from,child_attachments,parent_id,permalink_url,attachments{title.as(caption),description,unshimmed_url.as(link),target{id},media{source,image{src,height,width}},type,media_type,subattachments{data{type,media_type,media{source,image{src,height,width}}}}},post_photo_views,message,message_tags,full_picture,created_time,updated_time,shares,status_type,admin_creator,id,insights.metric(post_media_view,post_clicks,post_impressions_unique),insights.metric(post_media_view).period(lifetime).breakdown(is_from_ads).as(post_media_view_by_add),insights.metric(post_media_view).period(lifetime).breakdown(is_from_followers).as(post_media_view_by_followers),reactions.type(LIKE).limit(0).summary(1).as(like),reactions.type(LOVE).limit(0).summary(1).as(love),reactions.type(HAHA).limit(0).summary(1).as(haha),reactions.type(WOW).limit(0).summary(1).as(wow),reactions.type(SAD).limit(0).summary(1).as(sad),reactions.type(ANGRY).limit(0).summary(1).as(angry),reactions.type(THANKFUL).limit(0).summary(1).as(thankful),reactions.limit(0).summary(1).as(total),comments.limit(0).summary(true).filter(stream)"
	//postFields = "from,child_attachments,parent_id,permalink_url,attachments{title.as(caption),description,unshimmed_url.as(link),target{id},media{source,image{src,height,width}},type,media_type,subattachments{data{type,media_type,media{source,image{src,height,width}}}}},message,message_tags,full_picture,created_time,updated_time,shares,status_type,admin_creator,id,insights.metric(post_media_view),insights.metric(post_media_view,post_impressions_unique,post_impressions_unique,post_impressions_paid_unique,post_impressions_viral_unique).period(lifetime).breakdown(is_from_ads).as(post_media_view_by_add),insights.metric(post_media_view).period(lifetime).breakdown(is_from_followers).as(post_media_view_by_followers),reactions.type(LIKE).limit(0).summary(1).as(like),reactions.type(LOVE).limit(0).summary(1).as(love),reactions.type(HAHA).limit(0).summary(1).as(haha),reactions.type(WOW).limit(0).summary(1).as(wow),reactions.type(SAD).limit(0).summary(1).as(sad),reactions.type(ANGRY).limit(0).summary(1).as(angry),reactions.type(THANKFUL).limit(0).summary(1).as(thankful),comments.limit(0).summary(true).filter(stream)"
	/*
						These following fields are removed from the PostField and they have these alternate fields.
					    post_impressions* 							(Alternative: post_media_view)
						post_impressions_paid* 						(Alternative: post_media_view with is_from_ads breakdown)
						post_impressions_fan* 						(Alternative: post_media_view with is_from_followers breakdown)
						post_impressions_organic* 					(Alternative: post_media_view with is_from_ads breakdown)
						post_impressions_viral*                     Removed
						post_impressions_nonviral*                  Removed
			            post_impressions_unique                     Removed
		                post_impressions_unique 					Removed
		                post_impressions_paid_unique 				Removed
		                post_impressions_viral_unique 				Removed

				        for further, please take a look at the facebook documentation: https://developers.facebook.com/docs/platforminsights/page/deprecated-metrics

	*/

	//postFields = insights.metric(post_impressions_unique,post_impressions,post_impressions_paid,post_impressions_paid_unique, ,post_impressions_organic_unique,post_impressions_viral,post_impressions_viral_unique) these values are removed

	// removed insights.metric(post_impressions_unique,post_impressions,post_impressions_paid,post_impressions_paid_unique,post_impressions_organic,post_impressions_organic_unique,post_impressions_viral,post_impressions_viral_unique,post_clicks,post_video_views)
	// videoFields is the comprehensive list of fields to request for each video.
	//videoFields = "post_id,created_time,updated_time,description,video_insights.metric(post_video_avg_time_watched,blue_reels_play_count,post_video_view_time,total_video_views,total_video_views_unique,total_video_views_autoplayed,total_video_views_clicked_to_play,total_video_views_organic,total_video_views_organic_unique,total_video_views_paid,total_video_views_paid_unique,total_video_views_sound_on,total_video_complete_views,total_video_complete_views_unique,total_video_complete_views_auto_played,total_video_complete_views_clicked_to_play,total_video_complete_views_organic,total_video_complete_views_organic_unique,total_video_complete_views_paid,total_video_complete_views_paid_unique,total_video_10s_views,total_video_10s_views_unique,total_video_10s_views_auto_played,total_video_10s_views_clicked_to_play,total_video_10s_views_organic,total_video_10s_views_paid,total_video_10s_views_sound_on,total_video_15s_views,total_video_60s_excludes_shorter_views,total_video_retention_graph,total_video_retention_graph_autoplayed,total_video_retention_graph_clicked_to_play,total_video_view_total_time,total_video_view_total_time_organic,total_video_view_total_time_paid,total_video_impressions,total_video_impressions_unique,total_video_impressions_paid_unique,total_video_impressions_paid,total_video_impressions_organic_unique,total_video_impressions_organic,total_video_impressions_viral_unique,total_video_impressions_viral,total_video_impressions_fan_unique,total_video_impressions_fan,total_video_impressions_fan_paid_unique,total_video_impressions_fan_paid,total_video_stories_by_action_type,total_video_reactions_by_type_total,total_video_view_time_by_age_bucket_and_gender,total_video_view_time_by_region_id,total_video_views_by_distribution_type,total_video_view_time_by_distribution_type,total_video_view_total_time_live,total_video_views_live)"
	videoFields = "post_id,created_time,updated_time,description,video_insights.metric(total_video_impressions,total_video_impressions_unique,total_video_views,total_video_views_unique,total_video_views_autoplayed,total_video_views_clicked_to_play,post_video_avg_time_watched,total_video_complete_views,total_video_complete_views_unique,blue_reels_play_count,post_video_view_time,post_impressions_unique)"

	thumbFields = "id,attachments{media_type,type,target{id},media{image{src,height,width}},subattachments{data{media_type,type,target{id},media{image{src,height,width}}}}},full_picture"

	windowSize       = 20 // posts processed per outer "window"
	idsPerCall       = 10 // IDs per Graph multi-id call
	maxAttempts      = 5
	baseBackoff      = 300 * time.Millisecond
	maxBackoff       = 8 * time.Second
	interWindowPause = 250 * time.Millisecond
)

type fbImage struct {
	Src  string
	W, H int
}

// apiError represents the structure of an error response from the Facebook Graph API.
type apiError struct {
	Error struct {
		Message   string `json:"message"`
		Type      string `json:"type"`
		Code      int    `json:"code"`
		FBTraceID string `json:"fbtrace_id"`
	} `json:"error"`
}

// IsExpectedCompetitorError checks if an error is an expected competitor API error that should not be sent to Sentry
// These are permission/auth errors (400-404) that are expected when fetching competitor data
func IsExpectedCompetitorErrorFB(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()

	// Expected permission/auth errors from Facebook API for competitors (produce 400, 401, 403, 404)
	expectedPatterns := []string{
		"Error validating access token",
		"The session has been invalidated",
		"does not exist, cannot be loaded due to missing permissions",
		"GraphMethodException/100",
		"does not support this operation",
		"Tried accessing nonexisting field",
		"OAuthException/100",
		"OAuthException/2",
		"OAuthException/10",
		"OAuthException/190",
		"Code: 190",
		"Not enough viewers for the media to show insights",
	}

	for _, pattern := range expectedPatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	return strings.Contains(errMsg, "status 401") ||
		strings.Contains(errMsg, "status 403") ||
		strings.Contains(errMsg, "status 404")
}

// IsFacebookAuthError returns true when err indicates an expired or invalid Facebook access token.
func IsFacebookAuthError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	authPatterns := []string{
		"oauthexception/190",
		"error validating access token",
		"the session has been invalidated",
		"invalid oauth access token",
		"token has expired",
		"status 401",
	}
	for _, p := range authPatterns {
		if strings.Contains(errMsg, p) {
			return true
		}
	}
	return false
}

// ---------- RateManager (global + per-token) ----------

type RateLimits struct {
	PerTokenRPS   float64
	PerTokenBurst int
	GlobalRPS     float64
	GlobalBurst   int
}

type RateManager struct {
	global   *rate.Limiter
	mu       sync.Mutex
	perToken map[string]*rate.Limiter
	limits   RateLimits
}

func NewRateManager(lims RateLimits) *RateManager {
	if lims.PerTokenRPS <= 0 {
		lims.PerTokenRPS = 4.0
	}
	if lims.PerTokenBurst <= 0 {
		lims.PerTokenBurst = 4
	}
	if lims.GlobalRPS <= 0 {
		lims.GlobalRPS = 12.0
	}
	if lims.GlobalBurst <= 0 {
		lims.GlobalBurst = 12
	}
	return &RateManager{
		global:   rate.NewLimiter(rate.Limit(lims.GlobalRPS), lims.GlobalBurst),
		perToken: make(map[string]*rate.Limiter),
		limits:   lims,
	}
}

func (rm *RateManager) tokenLimiter(token string) *rate.Limiter {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if lim, ok := rm.perToken[token]; ok {
		return lim
	}
	lim := rate.NewLimiter(rate.Limit(rm.limits.PerTokenRPS), rm.limits.PerTokenBurst)
	rm.perToken[token] = lim
	return lim
}

func (rm *RateManager) Wait(ctx context.Context, token string) error {
	if err := rm.global.Wait(ctx); err != nil {
		return err
	}
	return rm.tokenLimiter(token).Wait(ctx)
}

// ---------- Facebook client ----------

// FacebookClient is a client for interacting with the Facebook Graph API.
type FacebookClient struct {
	httpClient *http.Client
	baseURL    string
	appSecret  string
	log        *logger.Logger
	rate       *RateManager
}

// NewFacebookClient creates a client with **default** rate limits (keeps backward compatibility).
func NewFacebookClient(appSecret string) *FacebookClient {
	return NewFacebookClientWithRates(appSecret, NewRateManager(RateLimits{}))
}

// NewFacebookClientWithRates creates a client wired to a shared RateManager (as used in main.go).
func NewFacebookClientWithRates(appSecret string, rm *RateManager) *FacebookClient {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   8 * time.Second,
		ResponseHeaderTimeout: 40 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConns:          200,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		ForceAttemptHTTP2:     true,
	}
	if rm == nil {
		rm = NewRateManager(RateLimits{})
	}
	return &FacebookClient{
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   65 * time.Second,
		},
		baseURL:   fbBaseURL,
		appSecret: appSecret,
		log:       logger.New("info"),
		rate:      rm,
	}
}

// waitRate applies global + per-token throttling in one place.
func (c *FacebookClient) waitRate(ctx context.Context, token string) error {
	if c.rate == nil {
		// extremely defensive; should not happen
		c.rate = NewRateManager(RateLimits{})
	}
	return c.rate.Wait(ctx, token)
}

// generateAppSecretProof creates a SHA256 HMAC of the access token using the app secret as the key.
func (c *FacebookClient) generateAppSecretProof(accessToken string) string {
	h := hmac.New(sha256.New, []byte(c.appSecret))
	h.Write([]byte(accessToken))
	return hex.EncodeToString(h.Sum(nil))
}

// FetchVideos retrieves all videos for a given Facebook Page (then enriches posts).
func (c *FacebookClient) FetchVideos(ctx context.Context, pageID, accessToken string, maxPages int) ([]kafkamodels.RawFacebookVideo, error) {
	//allVideos, err := c.FetchVideosWithLimit(ctx, pageID, accessToken, maxPages)
	//
	//if err != nil {
	//	return nil, err
	//}
	//
	//// Enrichment often outlives the page-listing call; give it a minimum budget.
	//windows := (len(allVideos) + 19) / 20
	//minBudget := 90 * time.Second
	//if w := time.Duration(windows) * 1200 * time.Millisecond; w > minBudget {
	//	minBudget = w
	//	if minBudget > 8*time.Minute {
	//		minBudget = 8 * time.Minute
	//	}
	//}
	//enrichCtx, cancel := withMinBudget(ctx, minBudget)
	//defer cancel()
	//
	//return c.FetchPostsByIDs(enrichCtx, pageID, accessToken, allVideos)

	return c.FetchVideosWithLimit(ctx, pageID, accessToken, maxPages)

}

//// FetchPostsByIDs does Graph multi-GET by windows of 20 videos (sub-batching 10 IDs/call),
//// with retry/backoff and shared rate-limiter.
//func (c *FacebookClient) FetchPostsByIDs(
//	ctx context.Context,
//	pageID, accessToken string,
//	allVideos []kafkamodels.RawFacebookVideo,
//) ([]kafkamodels.RawFacebookVideo, error) {
//
//	const (
//		windowSize       = 20 // 0:20, 20:40, ...
//		idsPerCall       = 10 // keep URLs small
//		maxAttempts      = 5
//		baseBackoff      = 300 * time.Millisecond
//		maxBackoff       = 8 * time.Second
//		interWindowPause = 250 * time.Millisecond
//	)
//
//	normalize := func(pid string) string {
//		pid = strings.TrimSpace(pid)
//		if pid == "" {
//			return ""
//		}
//		if !strings.Contains(pid, "_") && pageID != "" {
//			return pageID + "_" + pid
//		}
//		return pid
//	}
//
//	// simple backoff with jitter
//	backoff := func(try int) time.Duration {
//		d := baseBackoff << (try - 1)
//		if d > maxBackoff {
//			d = maxBackoff
//		}
//		j := time.Duration(int64(d) / 4) // ±25%
//		return d - j/2 + time.Duration(rand.Int63n(int64(j)))
//	}
//
//	n := len(allVideos)
//
//	for start := 0; start < n; start += windowSize {
//		end := start + windowSize
//		if end > n {
//			end = n
//		}
//
//		part := allVideos[start:end]
//
//		// Build unique list of IDs for this window
//		seen := make(map[string]struct{}, len(part))
//		ids := make([]string, 0, len(part))
//		for _, v := range part {
//			pid := normalize(v.PostID)
//			if pid == "" {
//				continue
//			}
//			if _, ok := seen[pid]; ok {
//				continue
//			}
//			seen[pid] = struct{}{}
//			ids = append(ids, pid)
//		}
//		if len(ids) == 0 {
//			continue
//		}
//
//		// Sub-batch to keep URL short
//		for i := 0; i < len(ids); i += idsPerCall {
//			j := i + idsPerCall
//			if j > len(ids) {
//				j = len(ids)
//			}
//			batch := ids[i:j]
//
//			u, _ := url.Parse(fmt.Sprintf("%s%s", c.baseURL, fbAPIVersion))
//			q := u.Query()
//			q.Set("ids", strings.Join(batch, ","))
//			q.Set("fields", postFields)
//			q.Set("access_token", accessToken)
//			u.RawQuery = q.Encode()
//
//			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
//			q2 := req.URL.Query()
//			q2.Set("appsecret_proof", c.generateAppSecretProof(accessToken))
//			req.URL.RawQuery = q2.Encode()
//
//			// Attempt loop with shared rate-limit + backoff
//			var (
//				resp *http.Response
//				body []byte
//				err  error
//			)
//			for attempt := 1; attempt <= maxAttempts; attempt++ {
//				if err = c.waitRate(ctx, accessToken); err != nil {
//					return nil, fmt.Errorf("rate limit wait failed (window %d:%d): %w", start, end, err)
//				}
//
//				resp, err = c.httpClient.Do(req)
//				if err != nil {
//					if attempt == maxAttempts {
//						return nil, fmt.Errorf("multi-GET posts failed (window %d:%d): %w", start, end, err)
//					}
//					time.Sleep(backoff(attempt))
//					continue
//				}
//
//				body, _ = io.ReadAll(resp.Body)
//				resp.Body.Close()
//
//				if resp.StatusCode == http.StatusOK {
//					break
//				}
//
//				var fbErr apiError
//				_ = json.Unmarshal(body, &fbErr)
//
//				// honor Retry-After
//				if ra := resp.Header.Get("Retry-After"); ra != "" {
//					if s, _ := strconv.Atoi(strings.TrimSpace(ra)); s > 0 {
//						time.Sleep(time.Duration(s) * time.Second)
//					}
//				} else {
//					time.Sleep(backoff(attempt))
//				}
//
//				if attempt == maxAttempts {
//					if fbErr.Error.Message != "" {
//						return nil, fmt.Errorf("facebook API error (window %d:%d): %s (%s/%d)",
//							start, end, fbErr.Error.Message, fbErr.Error.Type, fbErr.Error.Code)
//					}
//					return nil, fmt.Errorf("http %d (window %d:%d): %s", resp.StatusCode, start, end, string(body))
//				}
//			}
//
//			var byID map[string]kafkamodels.RawFacebookPost
//			if err := json.Unmarshal(body, &byID); err != nil {
//				return nil, fmt.Errorf("decode multi-GET response (window %d:%d): %w", start, end, err)
//			}
//			for k := start; k < end; k++ {
//				pid := normalize(allVideos[k].PostID)
//				if pid == "" {
//					continue
//				}
//				if p, ok := byID[pid]; ok && p.ID != "" {
//					cp := p
//					allVideos[k].FaceBookVideosPostInsights = &cp
//				}
//			}
//
//			// Gentle spacing between sub-batches to avoid spikes
//			time.Sleep(100 * time.Millisecond)
//		}
//
//		// Gentle spacing between windows
//		time.Sleep(interWindowPause)
//	}
//
//	return allVideos, nil
//}

// FetchVideosWithLimit retrieves videos for a given Facebook Page with a custom page limit.
func (c *FacebookClient) FetchVideosWithLimit(ctx context.Context, pageID, accessToken string, maxPages int) ([]kafkamodels.RawFacebookVideo, error) {
	var allVideos []kafkamodels.RawFacebookVideo

	c.log.Info().
		Str("page_id", pageID).
		Int("max_pages", maxPages).
		Str("module", "facebook_client").
		Msg("Starting to fetch Facebook videos")

	apiURL, err := url.Parse(fmt.Sprintf("%s%s/%s/videos", c.baseURL, fbAPIVersion, pageID))
	if err != nil {
		c.log.Warn().Err(err).Str("page_id", pageID).Str("module", "facebook_client").Msg("Failed to parse base URL for videos")
		return nil, fmt.Errorf("FacebookClient.FetchVideosWithLimit: failed to parse base URL for videos: %w", err)
	}

	query := apiURL.Query()
	query.Set("fields", videoFields)
	query.Set("access_token", accessToken)
	query.Set("limit", "25")
	apiURL.RawQuery = query.Encode()

	nextURL := apiURL.String()
	pagesFetched := 0
	startTime := time.Now()

	// Loop until there are no more pages or we hit our limit
	for nextURL != "" && pagesFetched < maxPages {
		select {
		case <-ctx.Done():
			c.log.Warn().Str("page_id", pageID).Int("pages_fetched", pagesFetched).Int("videos_collected", len(allVideos)).Str("module", "facebook_client").Msg("Context cancelled while fetching Facebook videos")
			return nil, ctx.Err()
		default:
		}

		c.log.Debug().Str("page_id", pageID).Int("current_page", pagesFetched+1).Int("max_pages", maxPages).Int("videos_so_far", len(allVideos)).Str("module", "facebook_client").Msg("Fetching Facebook videos page")

		req, err := http.NewRequestWithContext(ctx, "GET", nextURL, nil)
		if err != nil {
			c.log.Warn().Err(err).Str("page_id", pageID).Int("page_number", pagesFetched+1).Str("module", "facebook_client").Msg("Failed to create HTTP request for videos")
			return nil, fmt.Errorf("FacebookClient.FetchVideosWithLimit: failed to create HTTP request for videos %s: %w", nextURL, err)
		}

		// Add appsecret_proof to every request for security
		proof := c.generateAppSecretProof(accessToken)
		q := req.URL.Query()
		q.Set("appsecret_proof", proof)
		req.URL.RawQuery = q.Encode()

		body, status, err := c.doWithRetry(ctx, pageID, req, "FetchVideosWithLimit")
		if err != nil {
			c.log.Warn().Err(err).Str("page_id", pageID).Int("page_number", pagesFetched+1).Str("module", "facebook_client").Msg("Failed to execute HTTP request for videos")
			return nil, fmt.Errorf("FacebookClient.FetchVideosWithLimit: failed to execute request to %s: %w", req.URL.String(), err)
		}

		if status != http.StatusOK {
			var fbError apiError
			if json.Unmarshal(body, &fbError) == nil {
				c.log.Warn().
					Str("page_id", pageID).
					Int("page_number", pagesFetched+1).
					Int("status_code", status).
					Str("fb_error_message", fbError.Error.Message).
					Str("fb_error_type", fbError.Error.Type).
					Int("fb_error_code", fbError.Error.Code).
					Str("fb_trace_id", fbError.Error.FBTraceID).
					Str("module", "facebook_client").
					Msg("Facebook API returned error for videos")
				return nil, fmt.Errorf("FacebookClient.FetchVideosWithLimit: facebook API error for videos: %s (Type: %s, Code: %d)", fbError.Error.Message, fbError.Error.Type, fbError.Error.Code)
			}
			c.log.Warn().Str("page_id", pageID).Int("page_number", pagesFetched+1).Int("status_code", status).Str("module", "facebook_client").Msg("Facebook API returned non-200 status for videos")
			return nil, fmt.Errorf("FacebookClient.FetchVideosWithLimit: facebook API returned non-200 status for videos: %d", status)
		}

		var apiResponse struct {
			Data   []kafkamodels.RawFacebookVideo `json:"data"`
			Paging struct {
				Next string `json:"next"`
			} `json:"paging"`
		}

		if err := json.Unmarshal(body, &apiResponse); err != nil {
			c.log.Warn().Err(err).Str("page_id", pageID).Int("page_number", pagesFetched+1).Str("module", "facebook_client").Msg("Failed to decode Facebook JSON response for videos")
			return nil, fmt.Errorf("FacebookClient.FetchVideosWithLimit: failed to decode facebook JSON response for videos: %w", err)
		}

		c.log.Info().
			Str("page_id", pageID).
			Int("page_number", pagesFetched+1).
			Int("videos_in_page", len(apiResponse.Data)).
			Int("total_videos_so_far", len(allVideos)+len(apiResponse.Data)).
			Bool("has_next_page", apiResponse.Paging.Next != "").
			Str("module", "facebook_client").
			Msg("Successfully fetched Facebook videos page")

		allVideos = append(allVideos, apiResponse.Data...)
		nextURL = apiResponse.Paging.Next
		pagesFetched++
	}

	elapsed := time.Since(startTime)

	if pagesFetched >= maxPages {
		c.log.Warn().Str("page_id", pageID).Int("max_pages", maxPages).Int("pages_fetched", pagesFetched).Int("total_videos", len(allVideos)).Dur("elapsed_time", elapsed).Str("module", "facebook_client").Msg("Reached maximum page limit while fetching Facebook videos")
	}

	c.log.Info().
		Str("page_id", pageID).
		Int("pages_fetched", pagesFetched).
		Int("total_videos", len(allVideos)).
		Dur("elapsed_time", elapsed).
		Float64("videos_per_second", float64(len(allVideos))/elapsed.Seconds()).
		Str("module", "facebook_client").
		Msg("Completed fetching Facebook videos")

	return allVideos, nil
}

// FetchVideosSince retrieves videos for a given Facebook Page since a specific date.
// The since parameter filters videos created after the specified time.
// Uses same incremental logic as posts (last 14 days).
func (c *FacebookClient) FetchVideosSince(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookVideo, error) {
	var allVideos []kafkamodels.RawFacebookVideo

	c.log.Info().
		Str("page_id", pageID).
		Time("since", since).
		Time("until", until).
		Str("module", "facebook_client").
		Msg("Starting to fetch Facebook videos with date filter")

	apiURL, err := url.Parse(fmt.Sprintf("%s%s/%s/videos", c.baseURL, fbAPIVersion, pageID))
	if err != nil {
		c.log.Warn().Err(err).Str("page_id", pageID).Str("module", "facebook_client").Msg("Failed to parse base URL for videos")
		return nil, fmt.Errorf("FacebookClient.FetchVideosSince: failed to parse base URL for videos: %w", err)
	}

	query := apiURL.Query()
	query.Set("fields", videoFields)
	query.Set("access_token", accessToken)
	query.Set("limit", "25")
	query.Set("since", fmt.Sprintf("%d", since.Unix()))
	query.Set("until", fmt.Sprintf("%d", until.Unix()))
	apiURL.RawQuery = query.Encode()

	nextURL := apiURL.String()
	pagesFetched := 0
	startTime := time.Now()

	for nextURL != "" {
		select {
		case <-ctx.Done():
			c.log.Warn().Str("page_id", pageID).Int("pages_fetched", pagesFetched).Int("videos_collected", len(allVideos)).Str("module", "facebook_client").Msg("Context cancelled while fetching Facebook videos")
			return allVideos, ctx.Err()
		default:
		}

		if pagesFetched >= maxPagesToFetch {
			c.log.Warn().Str("page_id", pageID).Int("pages_fetched", pagesFetched).Int("videos_collected", len(allVideos)).Str("module", "facebook_client").Msg("Reached max pages limit; returning partial results")
			break
		}

		c.log.Debug().Str("page_id", pageID).Int("current_page", pagesFetched+1).Int("videos_so_far", len(allVideos)).Str("module", "facebook_client").Msg("Fetching Facebook videos page")

		req, err := http.NewRequestWithContext(ctx, "GET", nextURL, nil)
		if err != nil {
			c.log.Warn().Err(err).Str("page_id", pageID).Int("page_number", pagesFetched+1).Str("module", "facebook_client").Msg("Failed to create HTTP request for videos")
			return allVideos, fmt.Errorf("FacebookClient.FetchVideosSince: failed to create HTTP request for videos %s: %w", nextURL, err)
		}

		proof := c.generateAppSecretProof(accessToken)
		q := req.URL.Query()
		q.Set("appsecret_proof", proof)
		req.URL.RawQuery = q.Encode()

		body, status, err := c.doWithRetry(ctx, pageID, req, "FetchVideosSince")
		if err != nil {
			c.log.Warn().Err(err).Str("page_id", pageID).Int("page_number", pagesFetched+1).Int("videos_collected", len(allVideos)).Str("module", "facebook_client").Msg("Failed to execute HTTP request for videos; returning partial results")
			return allVideos, fmt.Errorf("FacebookClient.FetchVideosSince: failed to execute request: %w", err)
		}

		if status != http.StatusOK {
			var fbError apiError
			if json.Unmarshal(body, &fbError) == nil {
				c.log.Warn().
					Str("page_id", pageID).
					Int("page_number", pagesFetched+1).
					Int("videos_collected", len(allVideos)).
					Int("status_code", status).
					Str("fb_error_message", fbError.Error.Message).
					Str("fb_error_type", fbError.Error.Type).
					Int("fb_error_code", fbError.Error.Code).
					Str("fb_trace_id", fbError.Error.FBTraceID).
					Str("module", "facebook_client").
					Msg("Facebook API returned error for videos; returning partial results")
				return allVideos, fmt.Errorf("FacebookClient.FetchVideosSince: facebook API error for videos: %s (Type: %s, Code: %d)", fbError.Error.Message, fbError.Error.Type, fbError.Error.Code)
			}
			c.log.Warn().Str("page_id", pageID).Int("page_number", pagesFetched+1).Int("videos_collected", len(allVideos)).Int("status_code", status).Str("module", "facebook_client").Msg("Facebook API returned non-200 status for videos; returning partial results")
			return allVideos, fmt.Errorf("FacebookClient.FetchVideosSince: facebook API returned non-200 status for videos: %d", status)
		}

		var apiResponse struct {
			Data   []kafkamodels.RawFacebookVideo `json:"data"`
			Paging struct {
				Next string `json:"next"`
			} `json:"paging"`
		}

		if err := json.Unmarshal(body, &apiResponse); err != nil {
			c.log.Warn().Err(err).Str("page_id", pageID).Int("page_number", pagesFetched+1).Int("videos_collected", len(allVideos)).Str("module", "facebook_client").Msg("Failed to decode Facebook JSON response for videos; returning partial results")
			return allVideos, fmt.Errorf("FacebookClient.FetchVideosSince: failed to decode facebook JSON response for videos: %w", err)
		}

		c.log.Info().
			Str("page_id", pageID).
			Int("page_number", pagesFetched+1).
			Int("videos_in_page", len(apiResponse.Data)).
			Int("total_videos_so_far", len(allVideos)+len(apiResponse.Data)).
			Bool("has_next_page", apiResponse.Paging.Next != "").
			Str("module", "facebook_client").
			Msg("Successfully fetched Facebook videos page")

		allVideos = append(allVideos, apiResponse.Data...)
		nextURL = apiResponse.Paging.Next
		pagesFetched++
	}

	elapsed := time.Since(startTime)

	c.log.Info().
		Str("page_id", pageID).
		Int("pages_fetched", pagesFetched).
		Int("total_videos", len(allVideos)).
		Dur("elapsed_time", elapsed).
		Float64("videos_per_second", float64(len(allVideos))/elapsed.Seconds()).
		Str("module", "facebook_client").
		Msg("Completed fetching Facebook videos with date filter")

	return allVideos, nil
}

// FetchPosts retrieves all posts for a given Facebook Page.
func (c *FacebookClient) FetchPosts(ctx context.Context, pageID, accessToken string) ([]kafkamodels.RawFacebookPost, error) {
	return c.FetchPostsWithLimit(ctx, pageID, accessToken, maxPagesToFetch)
}

// FetchPostsWithLimit retrieves posts for a given Facebook Page with a custom page limit.
func (c *FacebookClient) FetchPostsWithLimit(ctx context.Context, pageID, accessToken string, maxPages int) ([]kafkamodels.RawFacebookPost, error) {
	var allPosts []kafkamodels.RawFacebookPost

	c.log.Info().Str("page_id", pageID).Int("max_pages", maxPages).Str("module", "facebook_client").Msg("Starting to fetch Facebook posts")

	apiURL, err := url.Parse(fmt.Sprintf("%s%s/%s/posts", c.baseURL, fbAPIVersion, pageID))
	if err != nil {
		c.log.Warn().Err(err).Str("page_id", pageID).Str("module", "facebook_client").Msg("Failed to parse base URL")
		return nil, fmt.Errorf("FacebookClient.FetchPostsWithLimit: failed to parse base URL: %w", err)
	}

	query := apiURL.Query()
	query.Set("fields", postFields)
	query.Set("access_token", accessToken)
	query.Set("limit", "50")
	apiURL.RawQuery = query.Encode()

	nextURL := apiURL.String()
	pagesFetched := 0
	startTime := time.Now()

	for nextURL != "" && pagesFetched < maxPages {
		select {
		case <-ctx.Done():
			c.log.Warn().Str("page_id", pageID).Int("pages_fetched", pagesFetched).Int("posts_collected", len(allPosts)).Str("module", "facebook_client").Msg("Context cancelled while fetching Facebook posts")
			return nil, ctx.Err()
		default:
		}

		c.log.Debug().Str("page_id", pageID).Int("current_page", pagesFetched+1).Int("max_pages", maxPages).Int("posts_so_far", len(allPosts)).Str("module", "facebook_client").Msg("Fetching Facebook posts page")

		req, err := http.NewRequestWithContext(ctx, "GET", nextURL, nil)
		if err != nil {
			c.log.Warn().Err(err).Str("page_id", pageID).Int("page_number", pagesFetched+1).Str("module", "facebook_client").Msg("Failed to create HTTP request")
			return nil, fmt.Errorf("FacebookClient.FetchPostsWithLimit: failed to create HTTP request for %s: %w", nextURL, err)
		}

		proof := c.generateAppSecretProof(accessToken)
		q := req.URL.Query()
		q.Set("appsecret_proof", proof)
		req.URL.RawQuery = q.Encode()

		body, status, err := c.doWithRetry(ctx, pageID, req, "FetchPostsWithLimit")
		if err != nil {
			c.log.Warn().Err(err).Str("page_id", pageID).Int("page_number", pagesFetched+1).Str("module", "facebook_client").Msg("Failed to execute HTTP request")
			return nil, fmt.Errorf("FacebookClient.FetchPostsWithLimit: failed to execute request to %s: %w", req.URL.String(), err)
		}

		if status != http.StatusOK {
			var fbError apiError
			if json.Unmarshal(body, &fbError) == nil {
				err := fmt.Errorf("FacebookClient.FetchPostsWithLimit: facebook API error: %s (Type: %s, Code: %d)", fbError.Error.Message, fbError.Error.Type, fbError.Error.Code)
				c.log.Warn().
					Str("page_id", pageID).
					Int("page_number", pagesFetched+1).
					Int("status_code", status).
					Str("fb_error_message", fbError.Error.Message).
					Str("fb_error_type", fbError.Error.Type).
					Int("fb_error_code", fbError.Error.Code).
					Str("fb_trace_id", fbError.Error.FBTraceID).
					Str("module", "facebook_client").
					Msg("Facebook API returned error")
				return nil, err
			}
			c.log.Warn().Str("page_id", pageID).Int("page_number", pagesFetched+1).Int("status_code", status).Str("module", "facebook_client").Msg("Facebook API returned non-200 status")
			return nil, fmt.Errorf("FacebookClient.FetchPostsWithLimit: facebook API returned non-200 status: %d", status)
		}

		var apiResponse struct {
			Data   []kafkamodels.RawFacebookPost `json:"data"`
			Paging struct {
				Next string `json:"next"`
			} `json:"paging"`
		}

		if err := json.Unmarshal(body, &apiResponse); err != nil {
			c.log.Warn().Err(err).Str("page_id", pageID).Int("page_number", pagesFetched+1).Str("module", "facebook_client").Msg("Failed to decode Facebook JSON response")
			return nil, fmt.Errorf("FacebookClient.FetchPostsWithLimit: failed to decode facebook JSON response: %w", err)
		}

		c.log.Info().
			Str("page_id", pageID).
			Int("page_number", pagesFetched+1).
			Int("posts_in_page", len(apiResponse.Data)).
			Int("total_posts_so_far", len(allPosts)+len(apiResponse.Data)).
			Bool("has_next_page", apiResponse.Paging.Next != "").
			Str("module", "facebook_client").
			Msg("Successfully fetched Facebook posts page")

		allPosts = append(allPosts, apiResponse.Data...)
		nextURL = apiResponse.Paging.Next
		pagesFetched++
	}

	elapsed := time.Since(startTime)

	if pagesFetched >= maxPages {
		c.log.Warn().Str("page_id", pageID).Int("max_pages", maxPages).Int("total_posts", len(allPosts)).Dur("elapsed_time", elapsed).Str("module", "facebook_client").Msg("Reached maximum page limit - more data may be available")
	}

	c.log.Info().
		Str("page_id", pageID).
		Int("pages_fetched", pagesFetched).
		Int("total_posts", len(allPosts)).
		Dur("elapsed_time", elapsed).
		Float64("posts_per_second", float64(len(allPosts))/elapsed.Seconds()).
		Str("module", "facebook_client").
		Msg("Completed fetching Facebook posts")

	return allPosts, nil
}

// FetchPostsSince retrieves posts for a given Facebook Page within a date range.
// The since and until parameters filter posts by created time.
func (c *FacebookClient) FetchPostsSince(ctx context.Context, pageID, accessToken string, since, until time.Time) ([]kafkamodels.RawFacebookPost, error) {
	var allPosts []kafkamodels.RawFacebookPost

	c.log.Info().
		Str("page_id", pageID).
		Time("since", since).
		Time("until", until).
		Str("module", "facebook_client").
		Msg("Starting to fetch Facebook posts with date filter")

	apiURL, err := url.Parse(fmt.Sprintf("%s%s/%s/feed", c.baseURL, fbAPIVersion, pageID))
	if err != nil {
		c.log.Warn().Err(err).Str("page_id", pageID).Str("module", "facebook_client").Msg("Failed to parse base URL")
		return nil, fmt.Errorf("FacebookClient.FetchPostsSince: failed to parse base URL: %w", err)
	}

	query := apiURL.Query()
	query.Set("fields", postFields)
	query.Set("access_token", accessToken)
	query.Set("limit", "50")
	query.Set("since", fmt.Sprintf("%d", since.Unix()))
	query.Set("until", fmt.Sprintf("%d", until.Unix()))
	apiURL.RawQuery = query.Encode()

	nextURL := apiURL.String()
	pagesFetched := 0
	startTime := time.Now()

	for nextURL != "" {
		select {
		case <-ctx.Done():
			c.log.Warn().Str("page_id", pageID).Int("pages_fetched", pagesFetched).Int("posts_collected", len(allPosts)).Str("module", "facebook_client").Msg("Context cancelled while fetching Facebook posts")
			return allPosts, ctx.Err()
		default:
		}

		if pagesFetched >= maxPagesToFetch {
			c.log.Warn().Str("page_id", pageID).Int("pages_fetched", pagesFetched).Int("posts_collected", len(allPosts)).Str("module", "facebook_client").Msg("Reached max pages limit; returning partial results")
			break
		}

		c.log.Debug().Str("page_id", pageID).Int("current_page", pagesFetched+1).Int("posts_so_far", len(allPosts)).Str("module", "facebook_client").Msg("Fetching Facebook posts page")

		req, err := http.NewRequestWithContext(ctx, "GET", nextURL, nil)
		if err != nil {
			c.log.Warn().Err(err).Str("page_id", pageID).Int("page_number", pagesFetched+1).Str("module", "facebook_client").Msg("Failed to create HTTP request")
			return allPosts, fmt.Errorf("FacebookClient.FetchPostsSince: failed to create HTTP request for %s: %w", nextURL, err)
		}

		proof := c.generateAppSecretProof(accessToken)
		q := req.URL.Query()
		q.Set("appsecret_proof", proof)
		req.URL.RawQuery = q.Encode()

		body, status, err := c.doWithRetry(ctx, pageID, req, "FetchPostsSince")
		if err != nil {
			c.log.Warn().Err(err).Str("page_id", pageID).Int("page_number", pagesFetched+1).Int("posts_collected", len(allPosts)).Str("module", "facebook_client").Msg("Failed to execute HTTP request; returning partial results")
			return allPosts, fmt.Errorf("FacebookClient.FetchPostsSince: failed to execute request: %w", err)
		}

		if status != http.StatusOK {
			var fbError apiError
			if json.Unmarshal(body, &fbError) == nil {
				err := fmt.Errorf("FacebookClient.FetchPostsSince: facebook API error: %s (Type: %s, Code: %d)", fbError.Error.Message, fbError.Error.Type, fbError.Error.Code)
				c.log.Warn().
					Str("page_id", pageID).
					Int("page_number", pagesFetched+1).
					Int("posts_collected", len(allPosts)).
					Int("status_code", status).
					Str("fb_error_message", fbError.Error.Message).
					Str("fb_error_type", fbError.Error.Type).
					Int("fb_error_code", fbError.Error.Code).
					Str("fb_trace_id", fbError.Error.FBTraceID).
					Str("module", "facebook_client").
					Msg("Facebook API returned error; returning partial results")
				return allPosts, err
			}
			c.log.Warn().Str("page_id", pageID).Int("page_number", pagesFetched+1).Int("posts_collected", len(allPosts)).Int("status_code", status).Str("module", "facebook_client").Msg("Facebook API returned non-200 status; returning partial results")
			return allPosts, fmt.Errorf("FacebookClient.FetchPostsSince: facebook API returned non-200 status: %d", status)
		}

		var apiResponse struct {
			Data   []kafkamodels.RawFacebookPost `json:"data"`
			Paging struct {
				Next string `json:"next"`
			} `json:"paging"`
		}

		if err := json.Unmarshal(body, &apiResponse); err != nil {
			c.log.Warn().Err(err).Str("page_id", pageID).Int("page_number", pagesFetched+1).Int("posts_collected", len(allPosts)).Str("module", "facebook_client").Msg("Failed to decode Facebook JSON response; returning partial results")
			return allPosts, fmt.Errorf("FacebookClient.FetchPostsSince: failed to decode facebook JSON response: %w", err)
		}

		c.log.Info().
			Str("page_id", pageID).
			Int("page_number", pagesFetched+1).
			Int("posts_in_page", len(apiResponse.Data)).
			Int("total_posts_so_far", len(allPosts)+len(apiResponse.Data)).
			Bool("has_next_page", apiResponse.Paging.Next != "").
			Str("module", "facebook_client").
			Msg("Successfully fetched Facebook posts page")

		allPosts = append(allPosts, apiResponse.Data...)
		nextURL = apiResponse.Paging.Next
		pagesFetched++
	}

	elapsed := time.Since(startTime)

	c.log.Info().
		Str("page_id", pageID).
		Int("pages_fetched", pagesFetched).
		Int("total_posts", len(allPosts)).
		Time("since", since).
		Time("until", until).
		Dur("elapsed_time", elapsed).
		Float64("posts_per_second", float64(len(allPosts))/elapsed.Seconds()).
		Str("module", "facebook_client").
		Msg("Completed fetching Facebook posts with date filter")

	return allPosts, nil
}

// FetchInsights retrieves Facebook page insights for a given page and date range.
func (c *FacebookClient) FetchInsights(ctx context.Context, pageID, accessToken string, since, until time.Time) (*kafkamodels.RawFacebookInsights, error) {
	startTime := time.Now()

	c.log.Info().
		Str("page_id", pageID).
		Str("since", since.Format("2006-01-02")).
		Str("until", until.Format("2006-01-02")).
		Str("module", "facebook_client").
		Msg("Starting to fetch Facebook page insights")

	metrics := []string{

		/*

				    for further, please take a look at the facebook documentation: https://developers.facebook.com/docs/platforminsights/page/deprecated-metrics
					page_fans 							(Alternative: page_follows)
					Page_fans_city 						(Alternative: page_follows_city)
					Page_fans_country 					(Alternative: page_follows_country)
					page_impressions* 					(Alternative: page_media_view)
					page_impressions_paid* 				(Alternative: page_media_view with is_from_ads breakdown)
					page_impressions_viral*             Removed
					page_impressions_nonviral*			Removed
			        Page_fan_adds						Removed
					Page_fans_locale					Removed
					Page_fan_adds_unique				Removed
					Page_fan_removes					Removed
					page_fan_removes_unique*			Removed

		*/
		"page_follows",
		"page_views_total",
		"page_media_view",
		"page_post_engagements",
		"page_total_actions",
		"page_fan_adds_by_paid_non_paid_unique",
		"page_video_views",
		"page_video_views_paid",
		"page_video_views_organic",
		"page_actions_post_reactions_like_total",
		"page_actions_post_reactions_love_total",
		"page_actions_post_reactions_anger_total",
		"page_video_views_autoplayed",
		"page_impressions_unique",
	}

	apiURL, err := url.Parse(fmt.Sprintf("%s%s/%s/insights", c.baseURL, fbAPIVersion, pageID))

	if err != nil {
		c.log.Warn().Err(err).Str("page_id", pageID).Str("module", "facebook_client").Msg("Failed to parse base URL for insights")
		return nil, fmt.Errorf("FacebookClient.FetchInsights: failed to parse base URL for insights: %w", err)
	}

	query := apiURL.Query()
	query.Set("metric", strings.Join(metrics, ","))
	query.Set("period", "day")
	query.Set("since", fmt.Sprintf("%d", since.Unix()))
	query.Set("until", fmt.Sprintf("%d", until.Unix()))
	query.Set("access_token", accessToken)
	apiURL.RawQuery = query.Encode()

	proof := c.generateAppSecretProof(accessToken)
	query.Set("appsecret_proof", proof)
	apiURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL.String(), nil)
	if err != nil {
		c.log.Warn().Err(err).Str("page_id", pageID).Str("module", "facebook_client").Msg("Failed to create HTTP request for insights")
		return nil, fmt.Errorf("FacebookClient.FetchInsights: failed to create HTTP request for insights: %w", err)
	}

	c.log.Debug().
		Str("url", req.URL.String()).
		Str("module", "facebook_client").
		Msg("Requesting Facebook insights")

	body, status, err := c.doWithRetry(ctx, pageID, req, "FetchInsights")
	if err != nil {
		c.log.Warn().Err(err).Str("page_id", pageID).Str("module", "facebook_client").Msg("Failed to execute HTTP request for insights")
		return nil, fmt.Errorf("FacebookClient.FetchInsights: failed to execute request for insights: %w", err)
	}

	if status != http.StatusOK {
		var fbError apiError
		if json.Unmarshal(body, &fbError) == nil {
			err := fmt.Errorf("FacebookClient.FetchInsights: facebook API error for insights: %s (Type: %s, Code: %d)", fbError.Error.Message, fbError.Error.Type, fbError.Error.Code)
			c.log.Warn().
				Str("page_id", pageID).
				Int("status_code", status).
				Str("fb_error_message", fbError.Error.Message).
				Str("fb_error_type", fbError.Error.Type).
				Int("fb_error_code", fbError.Error.Code).
				Str("fb_trace_id", fbError.Error.FBTraceID).
				Str("module", "facebook_client").
				Msg("Facebook API returned error for insights")
			return nil, err
		}
		c.log.Warn().Str("page_id", pageID).Int("status_code", status).Str("module", "facebook_client").Msg("Facebook API returned non-200 status for insights")
		return nil, fmt.Errorf("FacebookClient.FetchInsights: facebook API returned non-200 status for insights: %d", status)
	}

	var insightsResponse struct {
		Data   []kafkamodels.FacebookInsightData `json:"data"`
		Paging map[string]interface{}            `json:"paging,omitempty"`
	}
	if err := json.Unmarshal(body, &insightsResponse); err != nil {
		c.log.Warn().Err(err).Str("page_id", pageID).Str("module", "facebook_client").Msg("Failed to decode Facebook JSON response for insights")
		return nil, fmt.Errorf("FacebookClient.FetchInsights: failed to decode facebook JSON response for insights: %w", err)
	}

	// Demographic (lifetime) metrics
	demographicMetrics := []string{
		"page_fans",
		"page_fans_locale",
		"page_fans_country",
		"page_fans_city",
		"page_fans_gender_age",
	}

	demographicURL, err := url.Parse(fmt.Sprintf("%s%s/%s/insights", c.baseURL, fbAPIVersion, pageID))
	if err != nil {
		c.log.Warn().Err(err).Str("page_id", pageID).Msg("Failed to parse URL for demographic insights, continuing with basic metrics")
	} else {
		q := demographicURL.Query()
		q.Set("metric", strings.Join(demographicMetrics, ","))
		q.Set("period", "lifetime")
		q.Set("access_token", accessToken)
		q.Set("appsecret_proof", proof)
		demographicURL.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(ctx, "GET", demographicURL.String(), nil)

		if err == nil {
			if err := c.waitRate(ctx, accessToken); err != nil {
				return nil, fmt.Errorf("FacebookClient.FetchInsights: rate limit wait failed: %w", err)
			}
			if resp, err := c.httpClient.Do(req); err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					var demographicResponse struct {
						Data []kafkamodels.FacebookInsightData `json:"data"`
					}
					if json.NewDecoder(resp.Body).Decode(&demographicResponse) == nil {
						insightsResponse.Data = append(insightsResponse.Data, demographicResponse.Data...)
					}
				}
			}
		} else {
			c.log.Warn().Err(err).Str("page_id", pageID).Msg("Failed to build demographic request, continuing")
		}
	}

	// Additional page info
	pageInfoURL, err := url.Parse(fmt.Sprintf("%s%s/%s", c.baseURL, fbAPIVersion, pageID))
	if err != nil {
		c.log.Warn().Err(err).Str("page_id", pageID).Msg("Failed to parse URL for page info")
	} else {
		q := pageInfoURL.Query()
		q.Set("fields", "talking_about_count,category,fan_count")
		q.Set("access_token", accessToken)
		q.Set("appsecret_proof", proof)
		pageInfoURL.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(ctx, "GET", pageInfoURL.String(), nil)
		if err == nil {
			if err := c.waitRate(ctx, accessToken); err != nil {
				return nil, fmt.Errorf("FacebookClient.FetchInsights: rate limit wait failed: %w", err)
			}
			if resp, err := c.httpClient.Do(req); err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					var pageInfo struct {
						TalkingAboutCount int    `json:"talking_about_count"`
						Category          string `json:"category"`
						FanCount          int    `json:"fan_count"`
					}
					if json.NewDecoder(resp.Body).Decode(&pageInfo) == nil {
						// Collect all unique dates from the existing insights response.
						// This ensures page info values match the exact dates Facebook returned.
						dateSet := make(map[string]string)
						for _, insight := range insightsResponse.Data {
							for _, val := range insight.Values {
								if val.EndTime != "" {
									// Use the date part as key, but store the full EndTime string
									endTime, err := time.Parse(time.RFC3339, val.EndTime)
									if err != nil {
										endTime, err = time.Parse("2006-01-02T15:04:05-0700", val.EndTime)
										if err != nil {
											continue
										}
									}
									dateStr := endTime.Format("2006-01-02")
									if _, exists := dateSet[dateStr]; !exists {
										dateSet[dateStr] = val.EndTime
									}
								}
							}
						}

						// Generate values for each date found in the insights response.
						var talkingAboutValues []kafkamodels.FacebookInsightValue
						var categoryValues []kafkamodels.FacebookInsightValue
						var fanCountValues []kafkamodels.FacebookInsightValue

						for _, endTimeStr := range dateSet {
							talkingAboutValues = append(talkingAboutValues, kafkamodels.FacebookInsightValue{
								Value:   pageInfo.TalkingAboutCount,
								EndTime: endTimeStr,
							})
							categoryValues = append(categoryValues, kafkamodels.FacebookInsightValue{
								Value:   pageInfo.Category,
								EndTime: endTimeStr,
							})
							fanCountValues = append(fanCountValues, kafkamodels.FacebookInsightValue{
								Value:   pageInfo.FanCount,
								EndTime: endTimeStr,
							})
						}

						pageInfoInsights := []kafkamodels.FacebookInsightData{
							{
								Name:   "talking_about_count",
								Period: "day",
								Values: talkingAboutValues,
							},
							{
								Name:   "page_category",
								Period: "day",
								Values: categoryValues,
							},
							{
								Name:   "page_fans",
								Period: "day",
								Values: fanCountValues,
							},
						}
						insightsResponse.Data = append(insightsResponse.Data, pageInfoInsights...)
					}
				}
			}
		} else {
			c.log.Warn().Err(err).Str("page_id", pageID).Msg("Failed to build page info request, continuing")
		}
	}

	rawInsights := &kafkamodels.RawFacebookInsights{
		PageID:     pageID,
		Data:       insightsResponse.Data,
		Paging:     insightsResponse.Paging,
		SavingTime: time.Now().UTC(),
	}

	elapsed := time.Since(startTime)

	c.log.Info().
		Str("page_id", pageID).
		Int("insights_count", len(insightsResponse.Data)).
		Dur("elapsed_time", elapsed).
		Str("module", "facebook_client").
		Msg("Successfully fetched Facebook page insights")

	return rawInsights, nil
}

// ---- Public Orchestrator -----------------------------------------------------

// GetPostThumbnails fetches best thumbnails for input posts using FB Graph multi-id calls.
// It processes posts in windows, batches each window, retries with backoff, and logs progress.
func (c *FacebookClient) GetPostThumbnails(
	ctx context.Context,
	pageID, accessToken, longAccessToken, decryptionKey string,
	posts []clickhouse.MinimalPost,
) ([]clickhouse.MinimalPost, error) {

	startTS := time.Now()
	c.log.Info().
		Str("facebook_id", pageID).
		Int("posts_in", len(posts)).
		Int("window_size", windowSize).
		Int("ids_per_call", idsPerCall).
		Str("component", "facebook_client").
		Msg("GetPostThumbnails: starting")

	// Resolve the token (prefer long-lived if available)
	at, err := c.resolveToken(pageID, accessToken, longAccessToken, decryptionKey)
	if err != nil {
		return nil, err
	}

	// Early exit
	if len(posts) == 0 {
		c.log.Info().Str("facebook_id", pageID).Str("component", "facebook_client").Msg("GetPostThumbnails: no posts provided; returning empty result")
		return nil, nil
	}

	// Main accumulation maps: normalized post_id -> best image URL, or permanently inaccessible
	outThumbs := make(map[string]string, len(posts))
	outErrors := make(map[string]struct{}, 16)

	// Process posts in windows
	for start := 0; start < len(posts); start += windowSize {
		if err := ctx.Err(); err != nil {
			c.log.Warn().Str("facebook_id", pageID).Str("component", "facebook_client").Err(err).Msg("GetPostThumbnails: context canceled")
			return nil, err
		}
		end := min(start+windowSize, len(posts))
		window := posts[start:end]

		if err := c.processWindow(ctx, pageID, at, window, outThumbs, outErrors, start, end); err != nil {
			return nil, err
		}
		time.Sleep(interWindowPause)
	}

	// Assemble results in caller order; also include posts to clear (empty FullPicture)
	// so BulkUpdateFullPictures can zero-out their full_picture in ClickHouse, preventing
	// them from being retried on every future run.
	result := c.assembleResults(pageID, posts, outThumbs)
	result = append(result, c.assembleClearedPosts(pageID, posts, outErrors)...)

	c.log.Info().
		Str("facebook_id", pageID).
		Int("posts_in", len(posts)).
		Int("posts_with_thumbs", len(result)).
		Int("unique_ids_resolved", len(outThumbs)).
		Int("permanently_inaccessible", len(outErrors)).
		Dur("duration", time.Since(startTS)).
		Str("component", "facebook_client").
		Msg("GetPostThumbnails: completed")

	return result, nil
}

// ---- Helpers: Orchestration --------------------------------------------------

// processWindow normalizes post IDs in a window, then fetches thumbnails in batches.
func (c *FacebookClient) processWindow(
	ctx context.Context,
	pageID, accessToken string,
	window []clickhouse.MinimalPost,
	outThumbs map[string]string,
	outErrors map[string]struct{},
	windowStart, windowEnd int,
) error {

	c.log.Debug().
		Str("facebook_id", pageID).
		Int("window_start", windowStart).
		Int("window_end", windowEnd).
		Int("window_size_actual", len(window)).
		Str("component", "facebook_client").
		Msg("GetPostThumbnails: processing window")

	ids := buildUniqueNormalizedIDs(window, pageID)
	if len(ids) == 0 {
		c.log.Debug().
			Str("facebook_id", pageID).
			Int("window_start", windowStart).
			Int("window_end", windowEnd).
			Str("component", "facebook_client").
			Msg("GetPostThumbnails: window has no valid normalized IDs; skipping")
		return nil
	}

	c.log.Debug().
		Str("facebook_id", pageID).
		Int("ids_in_window", len(ids)).
		Str("component", "facebook_client").
		Msg("GetPostThumbnails: issuing multi-id requests")

	for i := 0; i < len(ids); i += idsPerCall {
		if err := ctx.Err(); err != nil {
			c.log.Warn().Str("facebook_id", pageID).Str("component", "facebook_client").Err(err).Msg("GetPostThumbnails: context canceled during batch loop")
			return err
		}

		batch := ids[i:min(i+idsPerCall, len(ids))]
		if err := c.processBatch(ctx, pageID, accessToken, batch, outThumbs, outErrors); err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond) // smooth bursts
	}

	return nil
}

// processBatch performs one multi-id GET, handles retries/backoff, decodes, and extracts best URLs.
func (c *FacebookClient) processBatch(
	ctx context.Context,
	pageID, accessToken string,
	batch []string,
	outThumbs map[string]string,
	outErrors map[string]struct{},
) error {

	u, _ := url.Parse(fmt.Sprintf("%s%s", c.baseURL, fbAPIVersion))
	q := u.Query()
	q.Set("ids", strings.Join(batch, ","))
	q.Set("fields", thumbFields)
	q.Set("access_token", accessToken)
	u.RawQuery = q.Encode()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if c.appSecret != "" {
		q2 := req.URL.Query()
		q2.Set("appsecret_proof", c.generateAppSecretProof(accessToken))
		req.URL.RawQuery = q2.Encode()
	}

	c.log.Debug().
		Str("facebook_id", pageID).
		Int("batch_size", len(batch)).
		Str("url", u.String()).
		Str("component", "facebook_client").
		Msg("GetPostThumbnails: sending request")

	body, status, err := c.doWithRetry(ctx, pageID, req, "GetPostThumbnails")
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		// This should be unreachable because doWithRetry exits non-200 with error after max attempts,
		// but keep a guard here anyway.
		return fmt.Errorf("FacebookClient.processBatch: unexpected http status after retry: %d", status)
	}

	byID, err := decodeByID(body)
	if err != nil {
		c.log.Warn().
			Str("facebook_id", pageID).
			Int("batch_size", len(batch)).
			Err(err).
			Str("component", "facebook_client").
			Msg("GetPostThumbnails: JSON decode failed")
		return fmt.Errorf("FacebookClient.processBatch: decode thumbs response: %w", err)
	}

	extracted := extractBestPerPost(batch, byID, outThumbs, outErrors)
	log.Debug().
		Str("facebook_id", pageID).
		Int("batch_size", len(batch)).
		Int("extracted", extracted).
		Str("component", "facebook_client").
		Msg("GetPostThumbnails: batch processed")

	return nil
}

// doWithRetry executes the HTTP request with rate-limit waits and exponential backoff + jitter.
func (c *FacebookClient) doWithRetry(
	ctx context.Context,
	pageID string,
	req *http.Request,
	caller string,
) (body []byte, status int, err error) {

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err = c.waitRate(ctx, req.URL.Query().Get("access_token")); err != nil {
			c.log.Warn().
				Str("facebook_id", pageID).
				Str("caller", caller).
				Err(err).
				Msg("Rate limit wait failed")
			return nil, 0, fmt.Errorf("FacebookClient.doWithRetry: rate limit wait failed: %w", err)
		}

		resp, httpErr := c.httpClient.Do(req)
		if httpErr != nil {
			if attempt == maxAttempts {
				c.log.Warn().
					Str("facebook_id", pageID).
					Str("caller", caller).
					Int("attempt", attempt).
					Err(httpErr).
					Msg("HTTP request failed; max attempts reached")
				return nil, 0, fmt.Errorf("FacebookClient.doWithRetry: request failed: %w", httpErr)
			}
			delay := computeBackoff(attempt)
			c.log.Warn().
				Str("facebook_id", pageID).
				Str("caller", caller).
				Int("attempt", attempt).
				Dur("backoff", delay).
				Err(httpErr).
				Msg("HTTP request failed; retrying with backoff")
			time.Sleep(delay)
			continue
		}

		body, _ = io.ReadAll(resp.Body)
		resp.Body.Close()
		status = resp.StatusCode

		if status == http.StatusOK {
			return body, status, nil
		}

		// Non-200: parse FB error (if present), then honor Retry-After or backoff.
		var fbErr apiError
		_ = json.Unmarshal(body, &fbErr)

		// Construct error for checking if it's expected
		var currentErr error
		if fbErr.Error.Message != "" {
			currentErr = fmt.Errorf("FacebookClient.doWithRetry: facebook API error: %s (%s/%d)", fbErr.Error.Message, fbErr.Error.Type, fbErr.Error.Code)
		} else {
			currentErr = fmt.Errorf("FacebookClient.doWithRetry: http %d: %s", status, string(body))
		}

		// Check if this is an expected client error (4xx permission/auth issues)
		isExpected := status >= 400 && status < 500 && IsExpectedCompetitorErrorFB(currentErr)

		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if s, _ := strconv.Atoi(strings.TrimSpace(ra)); s > 0 {
				c.log.Warn().
					Str("facebook_id", pageID).
					Str("caller", caller).
					Int("attempt", attempt).
					Int("retry_after_sec", s).
					Int("status_code", status).
					Str("fb_error_type", fbErr.Error.Type).
					Int("fb_error_code", fbErr.Error.Code).
					Str("fb_error_message", fbErr.Error.Message).
					Msg("Non-200 response; respecting Retry-After")
				time.Sleep(time.Duration(s) * time.Second)
			}
		} else {
			// For expected errors on first attempt, don't retry
			if isExpected && attempt == 1 {
				c.log.Warn().
					Str("facebook_id", pageID).
					Str("caller", caller).
					Int("status_code", status).
					Str("fb_error_type", fbErr.Error.Type).
					Int("fb_error_code", fbErr.Error.Code).
					Str("fb_error_message", fbErr.Error.Message).
					Msg("Facebook API: expected client error (permission/auth); not retrying")
				return nil, status, currentErr
			}

			delay := computeBackoff(attempt)
			c.log.Warn().
				Str("facebook_id", pageID).
				Str("caller", caller).
				Int("attempt", attempt).
				Int("status_code", status).
				Dur("backoff", delay).
				Str("fb_error_type", fbErr.Error.Type).
				Int("fb_error_code", fbErr.Error.Code).
				Str("fb_error_message", fbErr.Error.Message).
				Msg("Non-200 response; retrying with backoff")
			time.Sleep(delay)
		}

		if attempt == maxAttempts {
			// Final error with payload details if available
			if fbErr.Error.Message != "" {
				// Use Warn for expected errors, Error for unexpected
				if isExpected {
					c.log.Warn().
						Str("facebook_id", pageID).
						Str("caller", caller).
						Int("status_code", status).
						Str("fb_error_type", fbErr.Error.Type).
						Int("fb_error_code", fbErr.Error.Code).
						Str("fb_error_message", fbErr.Error.Message).
						Str("fb_trace_id", fbErr.Error.FBTraceID).
						Msg("Facebook API: expected client error after max attempts")
				} else {
					c.log.Warn().
						Str("facebook_id", pageID).
						Str("caller", caller).
						Int("status_code", status).
						Str("fb_error_type", fbErr.Error.Type).
						Int("fb_error_code", fbErr.Error.Code).
						Str("fb_error_message", fbErr.Error.Message).
						Str("fb_trace_id", fbErr.Error.FBTraceID).
						Msg("Giving up after max attempts (Facebook error)")
				}
				return nil, status, currentErr
			}
			if isExpected {
				c.log.Warn().
					Str("facebook_id", pageID).
					Str("caller", caller).
					Int("status_code", status).
					Msg("Facebook API: expected client error after max attempts (HTTP)")
			} else {
				c.log.Warn().
					Str("facebook_id", pageID).
					Str("caller", caller).
					Int("status_code", status).
					Msg("Giving up after max attempts (HTTP error)")
			}
			return nil, status, currentErr
		}
	}
	return nil, 0, fmt.Errorf("FacebookClient.doWithRetry: unreachable")
}

// assembleResults preserves the caller’s order and returns only resolved posts.
func (c *FacebookClient) assembleResults(
	pageID string,
	posts []clickhouse.MinimalPost,
	outThumbs map[string]string,
) []clickhouse.MinimalPost {
	res := make([]clickhouse.MinimalPost, 0, len(outThumbs))
	for _, in := range posts {
		norm := normalizePostID(in.PostID, pageID)
		if url := strings.TrimSpace(outThumbs[norm]); url != "" {
			res = append(res, clickhouse.MinimalPost{
				PageID:      in.PageID,
				PostID:      in.PostID,
				FullPicture: url,
			})
		}
	}
	return res
}

// assembleClearedPosts returns MinimalPost entries (with empty FullPicture) for every input post
// whose normalized ID appears in outErrors. These will be used to zero-out the full_picture field
// in ClickHouse so the post is excluded from future refresh runs.
func (c *FacebookClient) assembleClearedPosts(
	pageID string,
	posts []clickhouse.MinimalPost,
	outErrors map[string]struct{},
) []clickhouse.MinimalPost {
	if len(outErrors) == 0 {
		return nil
	}
	res := make([]clickhouse.MinimalPost, 0, len(outErrors))
	for _, in := range posts {
		norm := normalizePostID(in.PostID, pageID)
		if _, ok := outErrors[norm]; ok {
			res = append(res, clickhouse.MinimalPost{
				PageID:      in.PageID,
				PostID:      in.PostID,
				FullPicture: "",
			})
		}
	}
	return res
}

// ---- Helpers: Token, IDs, Decode, Selection ---------------------------------

// resolveToken picks the decrypted long token if available; logs outcomes.
func (c *FacebookClient) resolveToken(
	pageID, accessToken, longAccessToken, decryptionKey string,
) (string, error) {
	at := accessToken
	if longAccessToken != "" {
		if decrypted, err := crypto.DecryptToken(longAccessToken, decryptionKey); err != nil {
			c.log.Warn().Err(err).Str("facebook_id", pageID).Str("component", "facebook_client").Msg("GetPostThumbnails: failed to decrypt long access token; using provided token")
		} else {
			at = decrypted
			c.log.Info().Str("facebook_id", pageID).Str("component", "facebook_client").Msg("GetPostThumbnails: successfully decrypted long access token")
		}
	}
	if strings.TrimSpace(at) == "" {
		c.log.Warn().Str("facebook_id", pageID).Str("component", "facebook_client").Msg("GetPostThumbnails: no valid access token available")
		return "", fmt.Errorf("FacebookClient.resolveToken: no valid access token")
	}
	return at, nil
}

// buildUniqueNormalizedIDs returns de-duplicated normalized IDs for a window.
func buildUniqueNormalizedIDs(window []clickhouse.MinimalPost, pageID string) []string {
	seen := make(map[string]struct{}, len(window))
	out := make([]string, 0, len(window))
	for _, p := range window {
		np := normalizePostID(p.PostID, pageID)
		if np == "" {
			continue
		}
		if _, ok := seen[np]; ok {
			continue
		}
		seen[np] = struct{}{}
		out = append(out, np)
	}
	return out
}

// decodeByID parses the minimal map[id]payload response for the multi-id query.
func decodeByID(body []byte) (map[string]clickhouse.FbItem, error) {
	var byID map[string]clickhouse.FbItem
	if err := json.Unmarshal(body, &byID); err != nil {
		return nil, err
	}
	return byID, nil
}

// extractBestPerPost picks the best image for each id in batch and fills outThumbs.
// Posts that the API explicitly returns as permanently inaccessible (GraphMethodException/100)
// are recorded in outErrors so their URLs can be cleared from ClickHouse.
func extractBestPerPost(
	batch []string,
	byID map[string]clickhouse.FbItem,
	outThumbs map[string]string,
	outErrors map[string]struct{},
) int {
	extracted := 0
	for _, pid := range batch {
		p, ok := byID[pid]
		if !ok {
			continue // not in response — could be transient, don't clear
		}
		if p.ID == "" {
			// API returned an error node for this ID.
			if p.Error != nil && p.Error.Code == 100 {
				outErrors[pid] = struct{}{} // GraphMethodException/100 — permanently gone
			}
			continue
		}

		chosen := fbImage{}
		if p.Attachments != nil {
			for _, att := range p.Attachments.Data {
				if att.Media != nil && att.Media.Image != nil && att.Media.Image.Src != "" {
					chosen = bestImage(chosen, fbImage{
						Src: att.Media.Image.Src,
						W:   att.Media.Image.Width,
						H:   att.Media.Image.Height,
					})
				}
				// NOTE: subattachments ignored here; MinimalPost stores a single full_picture.
			}
		}

		url := strings.TrimSpace(chosen.Src)
		if url == "" {
			url = strings.TrimSpace(p.FullPicture)
		}
		if url != "" {
			outThumbs[pid] = url
			extracted++
		}
	}
	return extracted
}

// ---- Small Utilities ---------------------------------------------------------

func normalizePostID(pid, pageID string) string {
	pid = strings.TrimSpace(pid)
	if pid == "" {
		return ""
	}
	if !strings.Contains(pid, "_") && pageID != "" {
		return pageID + "_" + pid
	}
	return pid
}

func bestImage(a, b fbImage) fbImage {
	if a.W*a.H >= b.W*b.H {
		return a
	}
	return b
}

func computeBackoff(try int) time.Duration {
	d := baseBackoff << (try - 1)
	if d > maxBackoff {
		d = maxBackoff
	}
	j := time.Duration(int64(d) / 4) // ±25% jitter window
	return d - j/2 + time.Duration(rand.Int63n(int64(j)))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// withMinBudget ensures at least d of time; it still cancels if parent is canceled.
func withMinBudget(parent context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if dl, ok := parent.Deadline(); ok && time.Until(dl) >= d {
		ctx, cancel := context.WithCancel(parent)
		return ctx, cancel
	}
	ctx, cancel := context.WithTimeout(context.Background(), d)
	go func() {
		select {
		case <-parent.Done():
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx, cancel
}

// helpers
func ternaryID(t *struct {
	ID string `json:"id"`
}) string {
	if t == nil {
		return ""
	}
	return t.ID
}
func dedupeChildren(in []fbmodels.ChildThumb) []fbmodels.ChildThumb {
	seen := map[string]struct{}{}
	out := make([]fbmodels.ChildThumb, 0, len(in))
	for _, c := range in {
		key := c.MediaID + "|" + c.ThumbURL
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, c)
	}
	return out
}
