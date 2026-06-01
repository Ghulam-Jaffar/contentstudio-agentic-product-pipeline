package social

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

// LinkedInClient handles communication with LinkedIn API v2.
// NOTE: This is an initial skeleton. Pagination, error handling, and rate-limit backoff
// will be added once real tokens / scopes are ready.

type LinkedInClient struct {
	HTTPClient *http.Client
	BaseURL    map[string]string // e.g. https://api.linkedin.com/v2/
}

const (
	defaultAPIVersion   = "202509" // keep same as python LINKEDIN_VERSION; update via config if needed
	restliHeaderVersion = "2.0.0"
)

// IsExpectedCompetitorErrorLI checks if an error is an expected competitor API error that should not be sent to Sentry
// These are permission/auth errors (400-403) that are expected when fetching competitor data
func IsExpectedCompetitorErrorLI(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Expected errors from LinkedIn API for competitors (produce 400, 401, 403)
	expectedPatterns := []string{
		"EXPIRED_ACCESS_TOKEN",
		"INVALID_POST_FINDER_AUTHOR_ENTITY_TYPE",
		"The token used in the request has expired",
		"token invalid or expired",
		"status 401",
		"status 403",
		"status 400",
		"unauthorized",
		"permission",
		"not authorized",
	}

	for _, pattern := range expectedPatterns {
		if strings.Contains(strings.ToLower(errStr), strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

// IsLinkedInAuthError returns true when err indicates an expired or invalid LinkedIn access token.
func IsLinkedInAuthError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	authPatterns := []string{
		"expired_access_token",
		"the token used in the request has expired",
		"token invalid or expired",
		"status 401",
		"unauthorized",
	}
	for _, p := range authPatterns {
		if strings.Contains(errStr, p) {
			return true
		}
	}
	return false
}

// NewLinkedInClient returns a new client with sane defaults.
func NewLinkedInClient() *LinkedInClient {
	return &LinkedInClient{
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
		BaseURL: map[string]string{
			"v1": "https://api.linkedin.com/v2/", // added both of the urls so that we can use it where we need it.
			"v2": "https://api.linkedin.com/rest/",
		},
	}
}

// these both struct is used to build the final followers object so that parser parse it once for total followers
//and other demographic use case as well

var size struct {
	FirstDegreeSize int `json:"firstDegreeSize"`
}

var out struct {
	Paging          json.RawMessage   `json:"paging"`
	FirstDegreeSize int               `json:"firstDegreeSize"`
	Elements        json.RawMessage   `json:"elements"`
	GeoNames        map[string]string `json:"geoNames,omitempty"` // Resolved geo ID -> name mapping
}

// makeRequest performs GET with retries & fetcher headers.
func (c *LinkedInClient) makeRequest(ctx context.Context, url string, accessToken string, extraHeaders map[string]string, follower bool) ([]byte, int, error) {
	const maxAttempts = 3

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, 0, err
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("LinkedIn-Version", defaultAPIVersion)
		for k, v := range extraHeaders {
			req.Header.Set(k, v)
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			if attempt == maxAttempts {
				return nil, 0, err
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

		// Never retry auth errors
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
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

	return nil, 0, fmt.Errorf("LinkedInClient.makeRequest: max retries exceeded for %s", url)
}

// FetchShares fetches recent organisation shares (posts) for a given organisation ID.
// Docs: https://learn.microsoft.com/en-us/linkedin/marketing/integrations/community-management/shares/ugly-url
func (c *LinkedInClient) FetchShares(ctx context.Context, organisationID, accessToken string) ([]json.RawMessage, error) {
	reqURL := c.BaseURL["v2"] + fmt.Sprintf("posts?q=author&author=urn:li:organization:%s&count=100", organisationID)

	body, status, err := c.makeRequest(ctx, reqURL, accessToken, nil, false)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("LinkedInClient.FetchShares: linkedin api-server error: %d - %s", status, string(body))
	}

	var wrapper struct {
		Elements []json.RawMessage `json:"elements"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, err
	}

	return wrapper.Elements, nil
}

// FetchPostsPaginated returns posts until cutoffTime is reached or no more posts.
// If cutoffTime is zero, fetches all available posts (up to 12 months API limit).
// Posts are fetched sorted by CREATED (newest first) and stops when createdAt < cutoffTime.
func (c *LinkedInClient) FetchPostsPaginated(ctx context.Context, linkedinID string, entityType string, accessToken string, cutoffTime time.Time) ([]json.RawMessage, error) {

	collected := make([]json.RawMessage, 0, 100)
	start := 0
	count := 100
	hasCutoff := !cutoffTime.IsZero()
	const maxLinkedInPages = 50
	pagesFetched := 0

	for {
		if pagesFetched >= maxLinkedInPages {
			return collected, fmt.Errorf("LinkedInClient.FetchPostsPaginated: reached max pages limit (%d); returning %d partial results", maxLinkedInPages, len(collected))
		}
		url := c.BaseURL["v2"] + fmt.Sprintf("posts?q=author&author=urn:li:%s:%s&count=%d&start=%d&sortBy=CREATED", entityType, linkedinID, count, start)

		body, status, err := c.makeRequest(ctx, url, accessToken, nil, false)
		if err != nil {
			return collected, err
		}
		if status != http.StatusOK {
			return collected, fmt.Errorf("LinkedInClient.FetchPostsPaginated: linkedin posts error: status %d body %s", status, string(body))
		}

		var resp struct {
			Elements []json.RawMessage `json:"elements"`
			Paging   struct {
				Start int `json:"start"`
				Total int `json:"total"`
				Count int `json:"count"`
			} `json:"paging"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return collected, err
		}
		if len(resp.Elements) == 0 {
			break
		}

		reachedCutoff := false
		for _, el := range resp.Elements {
			// Parse createdAt to check against cutoff
			if hasCutoff {
				var post struct {
					CreatedAt int64 `json:"createdAt"`
				}
				if err := json.Unmarshal(el, &post); err == nil && post.CreatedAt > 0 {
					postTime := time.UnixMilli(post.CreatedAt)
					if postTime.Before(cutoffTime) {
						reachedCutoff = true
						break
					}
				}
			}
			collected = append(collected, el)
		}

		if reachedCutoff {
			break
		}

		pagesFetched++
		start += count
		// Stop if we've fetched all available posts
		if len(resp.Elements) < count {
			break
		}

		// small delay to respect rate limits
		select {
		case <-ctx.Done():
			return collected, ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}

	return collected, nil
}

// joinIDs builds query string param ?ids=id1&ids=id2
func (c *LinkedInClient) joinIDs(param string, ids []string) string {
	return "?" + param + "=" + strings.Join(ids, "&"+param+"=")
}

// FetchImagesRaw fetches image assets for given ids (max 80 per call) and returns raw body.
func (c *LinkedInClient) FetchImagesRaw(ctx context.Context, ids []string, accessToken string) ([]byte, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	url := c.BaseURL["v2"] + "images" + c.joinIDs("ids", ids)
	body, status, err := c.makeRequest(ctx, url, accessToken, nil, false)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("LinkedInClient.FetchImagesRaw: linkedin images error status %d: %s", status, string(body))
	}
	return body, nil
}

// FetchVideosRaw fetches video asset metadata.
func (c *LinkedInClient) FetchVideosRaw(ctx context.Context, ids []string, accessToken string) ([]byte, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	url := c.BaseURL["v2"] + "videos" + c.joinIDs("ids", ids)
	body, status, err := c.makeRequest(ctx, url, accessToken, nil, false)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("LinkedInClient.FetchVideosRaw: linkedin videos error status %d: %s", status, string(body))
	}
	return body, nil
}

// FetchDocumentsRaw fetches document (carousel) metadata including page images.
func (c *LinkedInClient) FetchDocumentsRaw(ctx context.Context, ids []string, accessToken string) ([]byte, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	url := c.BaseURL["v2"] + "documents" + c.joinIDs("ids", ids)
	body, status, err := c.makeRequest(ctx, url, accessToken, nil, false)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("LinkedInClient.FetchDocumentsRaw: linkedin documents error status %d: %s", status, string(body))
	}
	return body, nil
}

type linkedinRefreshPost struct {
	activityID  string
	postID      string
	imageIDs    []string
	videoIDs    []string
	documentIDs []string
}

// GetPostURLs reuses the normal LinkedIn post->asset flow but only returns refreshed
// URL-bearing fields for stale posts already present in ClickHouse.
func (c *LinkedInClient) GetPostURLs(
	ctx context.Context,
	linkedinID, entityType, accessToken string,
	posts []clickhousemodels.LinkedInMinimalPost,
) ([]clickhousemodels.LinkedInMinimalPost, error) {
	if len(posts) == 0 {
		return nil, nil
	}

	rawPosts, err := c.FetchPostsPaginated(ctx, linkedinID, entityType, accessToken, time.Now().UTC().AddDate(0, -12, 0))
	if err != nil {
		return nil, err
	}

	targetPostIDs := make(map[string]clickhousemodels.LinkedInMinimalPost, len(posts))
	targetActivities := make(map[string]clickhousemodels.LinkedInMinimalPost, len(posts))
	for _, post := range posts {
		if post.PostID != "" {
			targetPostIDs[post.PostID] = post
		}
		if post.Activity != "" {
			targetActivities[post.Activity] = post
		}
	}

	matched := make(map[string]*linkedinRefreshPost, len(posts))
	imageIDs := make(map[string]struct{})
	videoIDs := make(map[string]struct{})
	documentIDs := make(map[string]struct{})

	for _, raw := range rawPosts {
		var post map[string]any
		if err := json.Unmarshal(raw, &post); err != nil {
			continue
		}

		activityID, _ := post["id"].(string)
		if activityID == "" {
			continue
		}
		postID := activityID[strings.LastIndex(activityID, ":")+1:]

		var target clickhousemodels.LinkedInMinimalPost
		var ok bool
		if target, ok = targetPostIDs[postID]; !ok {
			target, ok = targetActivities[activityID]
			if !ok {
				continue
			}
		}

		key := target.PostID
		if key == "" {
			key = postID
		}
		if key == "" {
			continue
		}

		ref := &linkedinRefreshPost{activityID: activityID, postID: key}
		collectLinkedInRefreshAssetIDs(post, ref, imageIDs, videoIDs, documentIDs)
		matched[key] = ref
	}

	if len(matched) == 0 {
		return nil, nil
	}

	imgMap, err := c.fetchLinkedInAssetMap(ctx, mapKeys(imageIDs), accessToken, c.FetchImagesRaw)
	if err != nil {
		return nil, err
	}
	vidMap, err := c.fetchLinkedInAssetMap(ctx, mapKeys(videoIDs), accessToken, c.FetchVideosRaw)
	if err != nil {
		return nil, err
	}
	docMap, err := c.fetchLinkedInAssetMap(ctx, mapKeys(documentIDs), accessToken, c.FetchDocumentsRaw)
	if err != nil {
		return nil, err
	}

	out := make([]clickhousemodels.LinkedInMinimalPost, 0, len(matched))
	for _, stale := range posts {
		ref := matched[stale.PostID]
		if ref == nil && stale.Activity != "" {
			postID := stale.Activity[strings.LastIndex(stale.Activity, ":")+1:]
			ref = matched[postID]
		}
		if ref == nil {
			continue
		}

		refreshed := clickhousemodels.LinkedInMinimalPost{
			LinkedinID: stale.LinkedinID,
			PostID:     stale.PostID,
			Activity:   stale.Activity,
		}
		appendURL := func(url string, preferPrimary bool) {
			url = strings.TrimSpace(url)
			if url == "" {
				return
			}
			if refreshed.Image == "" || preferPrimary {
				refreshed.Image = url
			}
			for _, existing := range refreshed.Media {
				if existing == url {
					return
				}
			}
			refreshed.Media = append(refreshed.Media, url)
		}

		for _, id := range ref.imageIDs {
			if asset := imgMap[id]; asset != nil {
				if url, _ := asset["downloadUrl"].(string); url != "" {
					appendURL(url, true)
				}
			}
		}
		for _, id := range ref.videoIDs {
			if asset := vidMap[id]; asset != nil {
				if url, _ := asset["thumbnail"].(string); url != "" {
					appendURL(url, true)
				}
			}
		}
		for _, id := range ref.documentIDs {
			if asset := docMap[id]; asset != nil {
				if url, _ := asset["downloadUrl"].(string); url != "" {
					appendURL(url, refreshed.Image == "")
				}
			}
		}

		if refreshed.Image == "" && len(refreshed.Media) == 0 {
			continue
		}
		out = append(out, refreshed)
	}

	return out, nil
}

func collectLinkedInRefreshAssetIDs(
	post map[string]any,
	ref *linkedinRefreshPost,
	imageIDs, videoIDs, documentIDs map[string]struct{},
) {
	content, _ := post["content"].(map[string]any)
	if content == nil {
		return
	}

	for key, value := range content {
		switch key {
		case "multiImage":
			m, ok := value.(map[string]any)
			if !ok {
				continue
			}
			images, ok := m["images"].([]any)
			if !ok {
				continue
			}
			for _, item := range images {
				img, ok := item.(map[string]any)
				if !ok {
					continue
				}
				id, _ := img["id"].(string)
				if id == "" {
					continue
				}
				imageIDs[id] = struct{}{}
				ref.imageIDs = append(ref.imageIDs, id)
			}
		case "article":
			m, ok := value.(map[string]any)
			if !ok {
				continue
			}
			id, _ := m["thumbnail"].(string)
			if id == "" {
				continue
			}
			imageIDs[id] = struct{}{}
			ref.imageIDs = append(ref.imageIDs, id)
		case "media":
			m, ok := value.(map[string]any)
			if !ok {
				continue
			}
			id, _ := m["id"].(string)
			if id == "" {
				continue
			}
			switch {
			case strings.Contains(id, "video"):
				videoIDs[id] = struct{}{}
				ref.videoIDs = append(ref.videoIDs, id)
			case strings.Contains(id, "document"):
				documentIDs[id] = struct{}{}
				ref.documentIDs = append(ref.documentIDs, id)
			default:
				imageIDs[id] = struct{}{}
				ref.imageIDs = append(ref.imageIDs, id)
			}
		}
	}
}

func (c *LinkedInClient) fetchLinkedInAssetMap(
	ctx context.Context,
	ids []string,
	accessToken string,
	fetch func(context.Context, []string, string) ([]byte, error),
) (map[string]map[string]any, error) {
	if len(ids) == 0 {
		return map[string]map[string]any{}, nil
	}

	out := make(map[string]map[string]any, len(ids))
	for _, chunkIDs := range chunkStrings(ids, 80) {
		body, err := fetch(ctx, chunkIDs, accessToken)
		if err != nil {
			return nil, err
		}
		for id, asset := range parseLinkedInAssetBatch(body) {
			out[id] = asset
		}
	}
	return out, nil
}

func parseLinkedInAssetBatch(body []byte) map[string]map[string]any {
	var resp struct {
		Results map[string]map[string]any `json:"results"`
	}
	_ = json.Unmarshal(body, &resp)

	out := make(map[string]map[string]any, len(resp.Results))
	for _, asset := range resp.Results {
		if id, _ := asset["id"].(string); id != "" {
			out[id] = asset
			continue
		}
		if id, _ := asset["asset"].(string); id != "" {
			out[id] = asset
		}
	}
	return out
}

func chunkStrings(in []string, size int) [][]string {
	if size <= 0 || len(in) == 0 {
		return nil
	}
	out := make([][]string, 0, (len(in)+size-1)/size)
	for len(in) > size {
		out = append(out, in[:size])
		in = in[size:]
	}
	return append(out, in)
}

func mapKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for key := range m {
		out = append(out, key)
	}
	return out
}

// FetchStatsRaw fetches share statistics for ugcPosts and shares urn lists.
func (c *LinkedInClient) FetchStatsRaw(ctx context.Context, linkedinID string, ugcPosts []string, shares []string, accessToken string) ([]byte, error) {
	if len(ugcPosts) == 0 && len(shares) == 0 {
		return nil, nil
	}
	urn := "?"
	if len(ugcPosts) > 0 {
		urn += "ugcPosts=" + strings.Join(ugcPosts, "&ugcPosts=")
	}
	if len(shares) > 0 {
		if len(ugcPosts) > 0 {
			urn += "&"
		}
		urn += "shares=" + strings.Join(shares, "&shares=")
	}
	url := c.BaseURL["v2"] + fmt.Sprintf("organizationalEntityShareStatistics%s&q=organizationalEntity&organizationalEntity=urn:li:organization:%s", urn, linkedinID)
	body, status, err := c.makeRequest(ctx, url, accessToken, nil, false)

	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("LinkedInClient.FetchStatsRaw: linkedin stats error status %d: %s", status, string(body))
	}
	return body, nil
}

// FetchFollowerData fetches follower demographic statistics for an organisation.
// Returns follower stats elements merged with total follower count (snapshot, not daily).
// The follower data is duplicated across all daily insight buckets.
// Note: This method returns raw data without geo name resolution.
// Use FetchFollowerDataWithGeoNames if you need geo IDs resolved to names.
//
// Docs: https://learn.microsoft.com/en-us/linkedin/marketing/community-management/organizations/follower-statistics
func (c *LinkedInClient) FetchFollowerData(ctx context.Context, linkedinID string, accessToken string) ([]byte, error) {
	return c.FetchFollowerDataWithGeoNames(ctx, linkedinID, accessToken, nil)
}

// FetchFollowerDataWithGeoNames fetches follower data and includes geo name mappings.
// If geoNames is nil, it will try to resolve geo IDs using LinkedIn API.
// If geoNames is provided (e.g., from ClickHouse cache), it will use those mappings.
func (c *LinkedInClient) FetchFollowerDataWithGeoNames(ctx context.Context, linkedinID string, accessToken string, geoNames map[string]string) ([]byte, error) {
	bodySR, err := c.fetchFollowerStatsRaw(ctx, linkedinID, accessToken)
	if err != nil {
		return nil, err
	}

	bodyS, err := c.fetchFollowers(ctx, linkedinID, accessToken)
	if err != nil {
		return nil, err
	}

	var base struct {
		Paging   json.RawMessage `json:"paging"`
		Elements json.RawMessage `json:"elements"`
	}
	if err := json.Unmarshal(bodySR, &base); err != nil {
		return nil, fmt.Errorf("LinkedInClient.FetchFollowerDataWithGeoNames: unmarshal big payload: %w", err)
	}

	if err := json.Unmarshal(bodyS, &size); err != nil {
		return nil, fmt.Errorf("LinkedInClient.FetchFollowerDataWithGeoNames: unmarshal size payload: %w", err)
	}

	out.Paging = base.Paging
	out.Elements = base.Elements
	out.FirstDegreeSize = size.FirstDegreeSize

	// If geoNames provided externally (from cache), use them
	if geoNames != nil {
		out.GeoNames = geoNames
	} else {
		// Extract geo IDs from elements and resolve to names via LinkedIn API
		geoIDs := c.extractGeoIDsFromElements(base.Elements)
		if len(geoIDs) > 0 {
			resolvedNames, err := c.ResolveGeoIDs(ctx, geoIDs, accessToken)
			if err != nil {
				// Failed to resolve geo IDs - will fall back to numeric IDs in parser
				out.GeoNames = nil
			} else {
				out.GeoNames = resolvedNames
			}
		}
	}

	// Pretty output (or use json.Marshal for compact)
	body, _ := json.MarshalIndent(out, "", "  ")
	return body, nil
}

// GeoIDWithType represents a geo ID with its type (country or city).
type GeoIDWithType struct {
	ID   string
	Type string // "country" or "city"
}

// FollowerStatsWithGeoIDs contains follower stats raw data and extracted geo IDs.
// This allows fetching follower stats once and reusing for both geo extraction and final data.
type FollowerStatsWithGeoIDs struct {
	RawStats   []byte          // Raw follower stats response
	GeoIDs     []GeoIDWithType // Extracted geo IDs with types
	TotalCount int             // Total follower count
}

// GetGeoIDsFromFollowerStatsRaw fetches raw follower stats and extracts geo IDs without resolving them.
// Use this to get IDs for cache lookup before calling FetchFollowerDataWithGeoNames.
func (c *LinkedInClient) GetGeoIDsFromFollowerStatsRaw(ctx context.Context, linkedinID string, accessToken string) ([]string, error) {
	bodySR, err := c.fetchFollowerStatsRaw(ctx, linkedinID, accessToken)
	if err != nil {
		return nil, err
	}

	var base struct {
		Elements json.RawMessage `json:"elements"`
	}
	if err := json.Unmarshal(bodySR, &base); err != nil {
		return nil, fmt.Errorf("LinkedInClient.GetGeoIDsFromFollowerStatsRaw: unmarshal payload: %w", err)
	}

	return c.extractGeoIDsFromElements(base.Elements), nil
}

// GetGeoIDsWithTypeFromFollowerStatsRaw fetches raw follower stats and extracts geo IDs with their types.
// Returns geo IDs with type information (country or city).
func (c *LinkedInClient) GetGeoIDsWithTypeFromFollowerStatsRaw(ctx context.Context, linkedinID string, accessToken string) ([]GeoIDWithType, error) {
	bodySR, err := c.fetchFollowerStatsRaw(ctx, linkedinID, accessToken)
	if err != nil {
		return nil, err
	}

	var base struct {
		Elements json.RawMessage `json:"elements"`
	}
	if err := json.Unmarshal(bodySR, &base); err != nil {
		return nil, fmt.Errorf("LinkedInClient.GetGeoIDsWithTypeFromFollowerStatsRaw: unmarshal payload: %w", err)
	}

	return c.extractGeoIDsWithTypeFromElements(base.Elements), nil
}

// FetchFollowerStatsWithGeoIDs fetches follower stats and extracts geo IDs in a single API call.
// Returns both raw stats and geo IDs, avoiding duplicate API calls.
// Use this when you need to resolve geo IDs before building the final response.
func (c *LinkedInClient) FetchFollowerStatsWithGeoIDs(ctx context.Context, linkedinID string, accessToken string) (*FollowerStatsWithGeoIDs, error) {
	// Fetch follower stats (single API call)
	bodySR, err := c.fetchFollowerStatsRaw(ctx, linkedinID, accessToken)
	if err != nil {
		return nil, err
	}

	// Fetch total follower count
	bodyS, err := c.fetchFollowers(ctx, linkedinID, accessToken)
	if err != nil {
		return nil, err
	}

	var base struct {
		Elements json.RawMessage `json:"elements"`
	}
	if err := json.Unmarshal(bodySR, &base); err != nil {
		return nil, fmt.Errorf("LinkedInClient.FetchFollowerStatsWithGeoIDs: unmarshal payload: %w", err)
	}

	var sizeData struct {
		FirstDegreeSize int `json:"firstDegreeSize"`
	}
	if err := json.Unmarshal(bodyS, &sizeData); err != nil {
		return nil, fmt.Errorf("LinkedInClient.FetchFollowerStatsWithGeoIDs: unmarshal size payload: %w", err)
	}

	// Extract geo IDs with types
	geoIDs := c.extractGeoIDsWithTypeFromElements(base.Elements)

	return &FollowerStatsWithGeoIDs{
		RawStats:   bodySR,
		GeoIDs:     geoIDs,
		TotalCount: sizeData.FirstDegreeSize,
	}, nil
}

// BuildFollowerDataWithGeoNames builds the final follower data response using pre-fetched stats.
// This avoids duplicate API calls when geo IDs were resolved separately.
func (c *LinkedInClient) BuildFollowerDataWithGeoNames(stats *FollowerStatsWithGeoIDs, geoNames map[string]string) ([]byte, error) {
	var base struct {
		Paging   json.RawMessage `json:"paging"`
		Elements json.RawMessage `json:"elements"`
	}
	if err := json.Unmarshal(stats.RawStats, &base); err != nil {
		return nil, fmt.Errorf("LinkedInClient.BuildFollowerDataWithGeoNames: unmarshal payload: %w", err)
	}

	out.Paging = base.Paging
	out.Elements = base.Elements
	out.FirstDegreeSize = stats.TotalCount
	out.GeoNames = geoNames

	body, _ := json.MarshalIndent(out, "", "  ")
	return body, nil
}

// extractGeoIDsFromElements parses the elements to extract all unique geo IDs with their types.
// Geo IDs are found in followerCountsByGeo (city) and followerCountsByGeoCountry (country) fields.
func (c *LinkedInClient) extractGeoIDsFromElements(elements json.RawMessage) []string {
	geoIDsWithType := c.extractGeoIDsWithTypeFromElements(elements)
	geoIDs := make([]string, 0, len(geoIDsWithType))
	for _, g := range geoIDsWithType {
		geoIDs = append(geoIDs, g.ID)
	}
	return geoIDs
}

// extractGeoIDsWithTypeFromElements parses the elements to extract all unique geo IDs with their types.
// followerCountsByGeoCountry -> type "country"
// followerCountsByGeo -> type "city"
func (c *LinkedInClient) extractGeoIDsWithTypeFromElements(elements json.RawMessage) []GeoIDWithType {
	var parsed []struct {
		FollowerCountsByGeoCountry []struct {
			Geo string `json:"geo"`
		} `json:"followerCountsByGeoCountry"`
		FollowerCountsByGeo []struct {
			Geo string `json:"geo"`
		} `json:"followerCountsByGeo"`
	}

	if err := json.Unmarshal(elements, &parsed); err != nil {
		return nil
	}

	// Use map to deduplicate, storing type
	geoIDMap := make(map[string]string) // id -> type
	for _, el := range parsed {
		for _, gc := range el.FollowerCountsByGeoCountry {
			// Extract ID from URN like "urn:li:geo:101022442"
			if idx := strings.LastIndex(gc.Geo, ":"); idx >= 0 {
				geoIDMap[gc.Geo[idx+1:]] = "country"
			}
		}
		for _, g := range el.FollowerCountsByGeo {
			if idx := strings.LastIndex(g.Geo, ":"); idx >= 0 {
				geoIDMap[g.Geo[idx+1:]] = "city"
			}
		}
	}

	// Convert to slice
	result := make([]GeoIDWithType, 0, len(geoIDMap))
	for id, geoType := range geoIDMap {
		result = append(result, GeoIDWithType{ID: id, Type: geoType})
	}
	return result
}
func (c *LinkedInClient) fetchFollowerStatsRaw(ctx context.Context, linkedinID string, accessToken string) ([]byte, error) {
	url := c.BaseURL["v2"] + fmt.Sprintf(
		"organizationalEntityFollowerStatistics?q=organizationalEntity&organizationalEntity=urn:li:organization:%s",
		linkedinID,
	)
	body, status, err := c.makeRequest(ctx, url, accessToken, nil, false)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("LinkedInClient.fetchFollowerStatsRaw: linkedin follower stats raw error status %d: %s", status, string(body))
	}
	return body, nil
}

func (c *LinkedInClient) fetchFollowers(ctx context.Context, linkedinID string, accessToken string) ([]byte, error) {
	url := c.BaseURL["v2"] + fmt.Sprintf("networkSizes/urn:li:organization:%s?edgeType=COMPANY_FOLLOWED_BY_MEMBER", linkedinID)
	body, status, err := c.makeRequest(ctx, url, accessToken, nil, true)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("LinkedInClient.fetchFollowers: linkedin followers error status %d: %s, id:%s", status, string(body), linkedinID)
	}

	return body, nil
}

// ResolveGeoIDs resolves LinkedIn geo IDs to human-readable location names.
// Takes a slice of geo IDs (e.g., ["101022442", "102713980"]) and returns a map of ID to name.
// Uses the LinkedIn Geo API: https://api.linkedin.com/v2/geo?ids=List(id1,id2,...)&locale=(language:en,country:US)
// LinkedIn API has a limit of 150 IDs per request, so we batch in chunks of 100.
func (c *LinkedInClient) ResolveGeoIDs(ctx context.Context, geoIDs []string, accessToken string) (map[string]string, error) {
	if len(geoIDs) == 0 {
		return map[string]string{}, nil
	}

	result := make(map[string]string)
	const batchSize = 100

	// Process geo IDs in batches of 100 (LinkedIn limit is 150)
	for i := 0; i < len(geoIDs); i += batchSize {
		end := i + batchSize
		if end > len(geoIDs) {
			end = len(geoIDs)
		}
		batch := geoIDs[i:end]

		batchResult, err := c.resolveGeoIDsBatch(ctx, batch, accessToken)
		if err != nil {
			return nil, err
		}

		for id, name := range batchResult {
			result[id] = name
		}
	}

	return result, nil
}

// resolveGeoIDsBatch resolves a single batch of geo IDs (max 100).
func (c *LinkedInClient) resolveGeoIDsBatch(ctx context.Context, geoIDs []string, accessToken string) (map[string]string, error) {
	// Build the IDs list for the API: List(id1,id2,id3...)
	idsList := "List(" + strings.Join(geoIDs, ",") + ")"
	url := c.BaseURL["v1"] + fmt.Sprintf("geo?ids=%s&locale=(language:en,country:US)", idsList)

	// Geo API requires X-Restli-Protocol-Version and LinkedIn-Version headers
	headers := map[string]string{
		"X-Restli-Protocol-Version": "2.0.0",
		"LinkedIn-Version":          defaultAPIVersion,
	}

	body, status, err := c.makeRequest(ctx, url, accessToken, headers, false)
	if err != nil {
		return nil, fmt.Errorf("LinkedInClient.resolveGeoIDsBatch: geo resolution request failed: %w", err)
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("LinkedInClient.resolveGeoIDsBatch: linkedin geo API error status %d: %s", status, string(body))
	}

	// Parse the response
	var response struct {
		Results map[string]struct {
			DefaultLocalizedName struct {
				Value string `json:"value"`
			} `json:"defaultLocalizedName"`
			ID int64 `json:"id"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("LinkedInClient.resolveGeoIDsBatch: failed to parse geo response: %w", err)
	}

	// Build the result map
	result := make(map[string]string, len(response.Results))
	for id, data := range response.Results {
		result[id] = data.DefaultLocalizedName.Value
	}

	return result, nil
}

// FetchOrganizationDetailsRaw fetches organization details including name.
// It returns the raw JSON response which the caller can forward to Kafka for later parsing.
func (c *LinkedInClient) FetchOrganizationDetailsRaw(ctx context.Context, linkedinID string, accessToken string) ([]byte, error) {
	url := c.BaseURL["v2"] + fmt.Sprintf("organizations/%s", linkedinID)
	body, status, err := c.makeRequest(ctx, url, accessToken, nil, false)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("LinkedInClient.FetchOrganizationDetailsRaw: linkedin organization details error status %d: %s", status, string(body))
	}
	return body, nil
}

// FetchPageStatisticsRaw fetches organization page view statistics (views by device, section, demographics).
// Parameters:
//   - startMs: start time in milliseconds (Unix epoch)
//   - endMs: end time in milliseconds (Unix epoch)
//
// Callers should pass:
//   - Incremental sync: last 10 days
//   - Full sync / Immediate: last 365 days
//
// Docs: https://learn.microsoft.com/en-us/linkedin/marketing/community-management/organizations/page-statistics
func (c *LinkedInClient) FetchPageStatisticsRaw(ctx context.Context, linkedinID string, accessToken string, startMs, endMs int64) ([]byte, error) {
	// LinkedIn API returns daily page view stats when timeIntervals is specified
	// Each element contains pageViews, uniquePageViews, and demographic breakdowns for that day
	url := c.BaseURL["v2"] + fmt.Sprintf(
		"organizationPageStatistics?q=organization&organization=urn:li:organization:%s&timeIntervals.timeGranularityType=DAY&timeIntervals.timeRange.start=%d&timeIntervals.timeRange.end=%d",
		linkedinID, startMs, endMs,
	)
	body, status, err := c.makeRequest(ctx, url, accessToken, nil, false)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("LinkedInClient.FetchPageStatisticsRaw: linkedin page statistics error status %d: %s", status, string(body))
	}
	return body, nil
}

// FetchShareStatisticsRaw fetches organization-level share statistics (clicks, likes, comments, shares, impressions, reach).
// Parameters:
//   - startMs: start time in milliseconds (Unix epoch)
//   - endMs: end time in milliseconds (Unix epoch)
//
// Returns daily engagement metrics (impressions, clicks, likes, comments, shares) when time range is provided.
// Each element contains totalShareStatistics with engagement data for that day.
// Docs: https://learn.microsoft.com/en-us/linkedin/marketing/community-management/organizations/share-statistics
func (c *LinkedInClient) FetchShareStatisticsRaw(ctx context.Context, linkedinID string, accessToken string, startMs, endMs int64) ([]byte, error) {
	url := c.BaseURL["v2"] + fmt.Sprintf(
		"organizationalEntityShareStatistics?q=organizationalEntity&organizationalEntity=urn:li:organization:%s&timeIntervals.timeGranularityType=DAY&timeIntervals.timeRange.start=%d&timeIntervals.timeRange.end=%d",
		linkedinID, startMs, endMs,
	)

	body, status, err := c.makeRequest(ctx, url, accessToken, nil, false)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("LinkedInClient.FetchShareStatisticsRaw: linkedin share statistics error status %d: %s", status, string(body))
	}
	return body, nil
}

// FetchMemberCreatorPostAnalyticsRaw fetches aggregated post statistics for a member creator.
// Parameters:
//   - queryType: IMPRESSION, MEMBERS_REACHED, RESHARE, REACTION, or COMMENT
//   - startDate: start date (inclusive) - can be nil for lifetime stats
//   - endDate: end date (exclusive) - can be nil for lifetime stats
//
// Note: MEMBERS_REACHED does not support DAILY aggregation, only TOTAL.
// For IMPRESSION, RESHARE, REACTION, COMMENT - daily breakdown is returned when dates are provided.
//
// API: GET https://api.linkedin.com/rest/memberCreatorPostAnalytics?q=me&queryType={type}&aggregation=DAILY&dateRange=(start:(day:D,month:M,year:Y),end:(day:D,month:M,year:Y))
func (c *LinkedInClient) FetchMemberCreatorPostAnalyticsRaw(ctx context.Context, accessToken string, queryType string, startDate, endDate *time.Time) ([]byte, error) {
	url := c.BaseURL["v2"] + "memberCreatorPostAnalytics?q=me&queryType=" + queryType

	// MEMBERS_REACHED doesn't support DAILY aggregation
	if queryType != "MEMBERS_REACHED" && startDate != nil && endDate != nil {
		url += "&aggregation=DAILY"
		url += fmt.Sprintf("&dateRange=(start:(day:%d,month:%d,year:%d),end:(day:%d,month:%d,year:%d))",
			startDate.Day(), int(startDate.Month()), startDate.Year(),
			endDate.Day(), int(endDate.Month()), endDate.Year())
	}

	headers := map[string]string{"X-Restli-Protocol-Version": restliHeaderVersion}
	body, status, err := c.makeRequest(ctx, url, accessToken, headers, false)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("LinkedInClient.FetchMemberCreatorPostAnalyticsRaw: linkedin member creator post analytics error status %d: %s", status, string(body))
	}

	return body, nil
}

// FetchMemberFollowersCountRaw fetches total follower count and daily follower count changes for a member.
// When startDate and endDate are provided, fetches daily breakdown using q=dateRange.
// When dates are nil, fetches total follower count using q=me.
//
// API (total): GET https://api.linkedin.com/rest/memberFollowersCount?q=me
// API (daily): GET https://api.linkedin.com/rest/memberFollowersCount?q=dateRange&dateRange=(start:(day:D,month:M,year:Y),end:(day:D,month:M,year:Y))
func (c *LinkedInClient) FetchMemberFollowersCountRaw(ctx context.Context, accessToken string, startDate, endDate *time.Time) ([]byte, error) {
	var url string
	if startDate != nil && endDate != nil {
		url = c.BaseURL["v2"] + "memberFollowersCount?q=dateRange"
		url += fmt.Sprintf("&dateRange=(start:(day:%d,month:%d,year:%d),end:(day:%d,month:%d,year:%d))",
			startDate.Day(), int(startDate.Month()), startDate.Year(),
			endDate.Day(), int(endDate.Month()), endDate.Year())
	} else {
		url = c.BaseURL["v2"] + "memberFollowersCount?q=me"
	}

	headers := map[string]string{"X-Restli-Protocol-Version": restliHeaderVersion}
	body, status, err := c.makeRequest(ctx, url, accessToken, headers, false)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("LinkedInClient.FetchMemberFollowersCountRaw: linkedin member followers count error status %d: %s", status, string(body))
	}

	return body, nil
}
