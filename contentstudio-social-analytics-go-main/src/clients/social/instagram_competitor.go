package social

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	models "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

const DefaultIGCompLimit = 25

/* =========================
   Public API
========================= */

// GetBusinessDiscovery fetches Instagram business discovery data (posts and insights) for competitor analysis
func (c *InstagramClient) GetBusinessDiscovery(
	ctx context.Context,
	username string,
	limit int,
	cursor string,
	accessToken string,
	businessAccountID string,
) (*models.InstagramBusinessDiscoveryResponse, error) {

	if limit <= 0 {
		limit = DefaultIGCompLimit
	}

	mediaFields := "id,caption,comments_count,like_count,media_product_type,media_type," +
		"media_url,children{media_url},permalink,timestamp"

	// Build the business discovery query
	after := ""
	if cursor != "" {
		after = fmt.Sprintf(".after(%s)", cursor)
	}

	fields := fmt.Sprintf(
		"business_discovery.username(%s){id,profile_picture_url,ig_id,username,"+
			"biography,name,followers_count,follows_count,media_count,"+
			"media.limit(%d)%s{%s}}",
		username,
		limit,
		after,
		mediaFields,
	)

	params := map[string]string{
		"fields": fields,
	}

	fullURL := c.buildURL(fmt.Sprintf("/%s", businessAccountID), params, accessToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("InstagramClient.GetBusinessDiscovery: failed to create request: %w", err)
	}

	// Use shared doWithRetry for consistent retry logic, rate limiting, and error handling
	body, status, err := c.doWithRetry(ctx, businessAccountID, req, "GetBusinessDiscovery")
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, parseAPIErrorIG(body, status)
	}

	var response models.InstagramBusinessDiscoveryResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("InstagramClient.GetBusinessDiscovery: failed to decode response: %w", err)
	}

	return &response, nil
}

// GetCompetitorMediaURLs fetches only the competitor post media URLs and profile image.
func (c *InstagramClient) GetCompetitorMediaURLs(
	ctx context.Context,
	username string,
	posts []clickhousemodels.InstagramCompetitorMinimalPost,
	accessToken string,
	businessAccountID string,
) ([]clickhousemodels.InstagramCompetitorMinimalPost, string, error) {
	if strings.TrimSpace(username) == "" {
		return nil, "", fmt.Errorf("InstagramClient.GetCompetitorMediaURLs: username is required")
	}
	if strings.TrimSpace(accessToken) == "" {
		return nil, "", fmt.Errorf("InstagramClient.GetCompetitorMediaURLs: access token is required")
	}
	if strings.TrimSpace(businessAccountID) == "" {
		return nil, "", fmt.Errorf("InstagramClient.GetCompetitorMediaURLs: business account id is required")
	}
	if len(posts) == 0 {
		return nil, "", nil
	}

	// Find the oldest created_at among the stale posts. Business discovery returns
	// posts newest-first; once we see a post older than this we've gone far enough.
	// Truncate to second precision so the comparison is consistent with the
	// second-precision timestamps returned by the Instagram API.
	var oldestTarget time.Time
	for _, p := range posts {
		if !p.CreatedAt.IsZero() && (oldestTarget.IsZero() || p.CreatedAt.Before(oldestTarget)) {
			oldestTarget = p.CreatedAt
		}
	}
	oldestTargetSec := oldestTarget.Truncate(time.Second)

	targets := make(map[string]struct{}, len(posts))
	for _, post := range posts {
		if id := strings.TrimSpace(post.PostID); id != "" {
			targets[id] = struct{}{}
		}
	}
	if len(targets) == 0 {
		return nil, "", nil
	}

	limit := len(targets)
	if limit > 50 {
		limit = 50
	}
	if limit <= 0 {
		limit = DefaultIGCompLimit
	}

	cursor := ""
	profilePictureURL := ""
	seen := make(map[string]clickhousemodels.InstagramCompetitorMinimalPost, len(targets))

	for {
		resp, err := c.getBusinessDiscoveryURLs(ctx, username, limit, cursor, accessToken, businessAccountID)
		if err != nil {
			return nil, "", err
		}
		if strings.TrimSpace(profilePictureURL) == "" {
			profilePictureURL = strings.TrimSpace(resp.BusinessDiscovery.ProfilePictureURL)
		}

		if resp.BusinessDiscovery.Media != nil {
			for _, media := range resp.BusinessDiscovery.Media.Data {
				// Stop scanning this page and skip further pages if we've gone
				// past the oldest post we're looking for (feed is newest-first).
				if !oldestTargetSec.IsZero() && strings.TrimSpace(media.Timestamp) != "" {
					postTime, err := time.Parse("2006-01-02T15:04:05-0700", media.Timestamp)
					if err == nil && postTime.Before(oldestTargetSec) {
						goto done
					}
				}
				if _, ok := targets[media.ID]; !ok {
					continue
				}
				seen[media.ID] = clickhousemodels.InstagramCompetitorMinimalPost{
					InstagramID:       posts[0].InstagramID,
					PostID:            media.ID,
					MediaURL:          buildCompetitorInstagramMediaURL(media),
					ProfilePictureURL: profilePictureURL,
				}
			}
		}

		if len(seen) == len(targets) {
			break
		}
		if resp.BusinessDiscovery.Media == nil || resp.BusinessDiscovery.Media.Paging == nil ||
			resp.BusinessDiscovery.Media.Paging.Cursors == nil || strings.TrimSpace(resp.BusinessDiscovery.Media.Paging.Cursors.After) == "" {
			break
		}
		cursor = resp.BusinessDiscovery.Media.Paging.Cursors.After
	}
done:

	out := make([]clickhousemodels.InstagramCompetitorMinimalPost, 0, len(seen))
	for _, post := range posts {
		if refreshed, ok := seen[strings.TrimSpace(post.PostID)]; ok {
			refreshed.CreatedAt = post.CreatedAt
			out = append(out, refreshed)
		}
	}

	return out, profilePictureURL, nil
}

/* =========================
   Helpers
========================= */

// buildURL constructs the full URL with access token and app secret proof
func (c *InstagramClient) buildURL(endpoint string, params map[string]string, accessToken string) string {
	u, _ := url.Parse(c.baseURL + igAPIVersion + endpoint)
	q := u.Query()

	q.Set("access_token", accessToken)
	q.Set("appsecret_proof", c.generateAppSecretProof(accessToken))

	for k, v := range params {
		q.Set(k, v)
	}

	u.RawQuery = q.Encode()
	return u.String()
}

// parseAPIErrorIG parses Instagram API error responses
func parseAPIErrorIG(body []byte, status int) error {
	var errResp models.ErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != nil {
		return fmt.Errorf(
			"parseAPIErrorIG: instagram api error (%d): %s - %s",
			errResp.Error.Code,
			errResp.Error.Type,
			errResp.Error.Message,
		)
	}
	return fmt.Errorf("parseAPIErrorIG: instagram api error (%d): %s", status, string(body))
}

func (c *InstagramClient) getBusinessDiscoveryURLs(
	ctx context.Context,
	username string,
	limit int,
	cursor string,
	accessToken string,
	businessAccountID string,
) (*models.InstagramBusinessDiscoveryResponse, error) {
	if limit <= 0 {
		limit = DefaultIGCompLimit
	}

	mediaFields := "id,media_type,media_url,timestamp,children{media_url}"
	after := ""
	if cursor != "" {
		after = fmt.Sprintf(".after(%s)", cursor)
	}

	fields := fmt.Sprintf(
		"business_discovery.username(%s){profile_picture_url,media.limit(%d)%s{%s}}",
		username,
		limit,
		after,
		mediaFields,
	)

	params := map[string]string{"fields": fields}
	fullURL := c.buildURL(fmt.Sprintf("/%s", businessAccountID), params, accessToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("InstagramClient.getBusinessDiscoveryURLs: failed to create request: %w", err)
	}

	body, status, err := c.doWithRetry(ctx, businessAccountID, req, "GetCompetitorMediaURLs")
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, parseAPIErrorIG(body, status)
	}

	var response models.InstagramBusinessDiscoveryResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("InstagramClient.getBusinessDiscoveryURLs: failed to decode response: %w", err)
	}

	return &response, nil
}

func buildCompetitorInstagramMediaURL(media models.InstagramMedia) string {
	if strings.EqualFold(media.MediaType, "CAROUSEL_ALBUM") && media.Children != nil && len(media.Children.Data) > 0 {
		urls := make([]string, 0, len(media.Children.Data))
		for _, child := range media.Children.Data {
			if strings.TrimSpace(child.MediaURL) != "" {
				urls = append(urls, child.MediaURL)
			}
		}
		if len(urls) > 0 {
			return strings.Join(urls, ",")
		}
	}
	return strings.TrimSpace(media.MediaURL)
}
