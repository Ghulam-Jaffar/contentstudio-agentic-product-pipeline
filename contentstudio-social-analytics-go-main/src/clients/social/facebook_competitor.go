package social

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	models "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

/* =========================
   Public API methods
========================= */

// GetCompetitorPageDetails fetches details of a Facebook page
func (c *FacebookClient) GetCompetitorPageDetails(
	ctx context.Context,
	pageID string,
	accessToken string,
) (*models.FacebookPageDetails, *models.Picture, error) {

	fields := "about,name,fan_count,talking_about_count,category,emails,followers_count," +
		"birthday,were_here_count,link,cover"

	// Fetch page details
	page := &models.FacebookPageDetails{}
	err := c.makeAPICall(ctx, "", fmt.Sprintf("/%s", pageID), map[string]string{
		"fields": fields,
	}, page, accessToken, "GetPageDetails")
	if err != nil {
		return nil, nil, err
	}

	// Fetch page picture
	picture := &models.Picture{}
	_ = c.makeAPICall(ctx, "", fmt.Sprintf("/%s/picture", pageID), map[string]string{
		"redirect": "0",
	}, picture, accessToken, "GetPageDetails")

	return page, picture, nil
}

// GetCompetitorPosts fetches posts from a Facebook page
func (c *FacebookClient) GetCompetitorPosts(
	ctx context.Context,
	pageID string,
	accessToken string,
	since, until time.Time,
	limit int,
) ([]*models.Post, string, error) {

	fields := "id,parent_id,created_time,from,message,status_type,permalink_url," +
		"full_picture,shares," +
		"attachments{title,description,unshimmed_url,target,media{source,image}," +
		"media_type,type,subattachments}," +
		"reactions.type(LIKE).limit(0).summary(1).as(like)," +
		"reactions.type(LOVE).limit(0).summary(1).as(love)," +
		"reactions.type(HAHA).limit(0).summary(1).as(haha)," +
		"reactions.type(WOW).limit(0).summary(1).as(wow)," +
		"reactions.type(SAD).limit(0).summary(1).as(sad)," +
		"reactions.type(ANGRY).limit(0).summary(1).as(angry)," +
		"comments.summary(1).as(comments)"

	params := map[string]string{
		"fields": fields,
		"since":  fmt.Sprintf("%d", since.Unix()),
		"until":  fmt.Sprintf("%d", until.Unix()),
		"limit":  fmt.Sprintf("%d", limit),
	}

	// Make the API call
	var response models.PagingResponse
	err := c.makeAPICall(ctx, "", fmt.Sprintf("/%s/posts", pageID), params, &response, accessToken, "GetPosts")
	if err != nil {
		return nil, "", err
	}

	// Parse posts
	posts := make([]*models.Post, len(response.Data))
	for i := range response.Data {
		posts[i] = &response.Data[i]
	}

	// Get next page URL
	nextURL := ""
	if response.Paging != nil {
		nextURL = response.Paging.Next
	}

	return posts, nextURL, nil
}

// GetPostsFromURL fetches posts using a next page URL
func (c *FacebookClient) GetCompetitorPostsFromURL(
	ctx context.Context,
	nextURL string,
	pageID string,
	accessToken string,
) ([]*models.Post, string, error) {

	// Ensure appsecret_proof is included
	if !strings.Contains(nextURL, "appsecret_proof=") {
		sep := "&"
		if !strings.Contains(nextURL, "?") {
			sep = "?"
		}
		nextURL = fmt.Sprintf("%s%sappsecret_proof=%s", nextURL, sep, c.generateAppSecretProof(accessToken))
	}

	// Make the API call
	var response models.PagingResponse
	err := c.makeAPICall(ctx, nextURL, "", nil, &response, accessToken, "GetPostsFromURL")
	if err != nil {
		return nil, "", err
	}

	// Parse posts
	posts := make([]*models.Post, len(response.Data))
	for i := range response.Data {
		posts[i] = &response.Data[i]
	}

	// Get next page URL
	next := ""
	if response.Paging != nil {
		next = response.Paging.Next
	}

	return posts, next, nil
}

// GetSharedPostDetails fetches details of the original shared post
func (c *FacebookClient) GetCompetitorSharedPostDetails(
	ctx context.Context,
	parentID string,
	accessToken string,
) (*models.Post, error) {

	fields := "id,parent_id,created_time,from,message,status_type,permalink_url," +
		"full_picture,shares," +
		"attachments{title,description,unshimmed_url,target,media{source,image}," +
		"media_type,type,subattachments}," +
		"reactions.type(LIKE).limit(0).summary(1).as(like)," +
		"reactions.type(LOVE).limit(0).summary(1).as(love)," +
		"comments.summary(1).as(comments)"

	params := map[string]string{
		"fields": fields,
	}

	// Make the API call
	var post models.Post
	err := c.makeAPICall(ctx, "", fmt.Sprintf("/%s", parentID), params, &post, accessToken, "GetSharedPostDetails")
	if err != nil {
		return nil, err
	}

	return &post, nil
}

// GetPagePicture fetches the profile picture of a page
func (c *FacebookClient) GetCompetitorPagePicture(
	ctx context.Context,
	pageID string,
	accessToken string,
) (*models.Picture, error) {

	// Fetch page picture
	picture := &models.Picture{}
	err := c.makeAPICall(ctx, "", fmt.Sprintf("/%s/picture", pageID), map[string]string{
		"redirect": "0",
	}, picture, accessToken, "GetPagePicture")
	if err != nil {
		return nil, err
	}

	return picture, nil
}

// GetCompetitorMediaAssetURLs fetches only the asset link fields needed for competitor URL refreshes.
func (c *FacebookClient) GetCompetitorMediaAssetURLs(
	ctx context.Context,
	facebookID string,
	accessToken string,
	assets []clickhousemodels.FacebookCompetitorMinimalMediaAsset,
) ([]clickhousemodels.FacebookCompetitorMinimalMediaAsset, error) {
	if strings.TrimSpace(accessToken) == "" {
		return nil, fmt.Errorf("FacebookClient.GetCompetitorMediaAssetURLs: access token is required")
	}
	if len(assets) == 0 {
		return nil, nil
	}

	postIDs := make([]string, 0, len(assets))
	seen := make(map[string]struct{}, len(assets))
	for _, asset := range assets {
		id := strings.TrimSpace(asset.PostID)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		postIDs = append(postIDs, id)
	}
	if len(postIDs) == 0 {
		return nil, nil
	}

	const batchSize = 50
	// child_attachments covers carousels that were ingested via the child_attachments API
	// field (MediaID stored as child.ID). The attachments+subattachments path covers the
	// rest (MediaID stored as subattachment.target.id or an md5 hash). full_picture
	// provides a fallback image URL for posts with no attachment or no media.image field.
	// Note: subattachments is a Graph API "connection" — the data[] wrapper is implicit
	// in the response so we must NOT nest fields inside data{} in the selector.
	fields := "id,full_picture,child_attachments{id,media{image{src,height,width}}},attachments{target{id},media{image{src,height,width}},subattachments{target{id},media{image{src,height,width}}}}"
	appsecretProof := c.generateAppSecretProof(accessToken)

	outByMediaID := make(map[string]clickhousemodels.FacebookCompetitorMinimalMediaAsset, len(assets))
	for i := 0; i < len(postIDs); i += batchSize {
		end := i + batchSize
		if end > len(postIDs) {
			end = len(postIDs)
		}
		batch := postIDs[i:end]

		u, _ := url.Parse(fmt.Sprintf("%s%s", c.baseURL, fbAPIVersion))
		q := u.Query()
		q.Set("ids", strings.Join(batch, ","))
		q.Set("fields", fields)
		q.Set("access_token", accessToken)
		q.Set("appsecret_proof", appsecretProof)
		u.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("FacebookClient.GetCompetitorMediaAssetURLs: failed to create request: %w", err)
		}

		body, status, err := c.doWithRetry(ctx, facebookID, req, "GetCompetitorMediaAssetURLs")
		if err != nil {
			return nil, err
		}
		if status != http.StatusOK {
			return nil, fmt.Errorf("FacebookClient.GetCompetitorMediaAssetURLs: unexpected http status after retry: %d", status)
		}

		var byID map[string]competitorPostURLsItem
		if err := json.Unmarshal(body, &byID); err != nil {
			return nil, fmt.Errorf("FacebookClient.GetCompetitorMediaAssetURLs: failed to decode response: %w", err)
		}

		for _, asset := range assets {
			item, ok := byID[strings.TrimSpace(asset.PostID)]
			if !ok || item.ID == "" {
				continue
			}
			if refreshed, ok := matchCompetitorMediaAsset(asset, item); ok {
				outByMediaID[refreshed.MediaID] = refreshed
			}
		}
	}

	out := make([]clickhousemodels.FacebookCompetitorMinimalMediaAsset, 0, len(outByMediaID))
	for _, asset := range assets {
		if refreshed, ok := outByMediaID[asset.MediaID]; ok {
			out = append(out, refreshed)
		}
	}

	return out, nil
}

// GetCompetitorSharedFromPictures fetches the source page pictures for shared posts.
func (c *FacebookClient) GetCompetitorSharedFromPictures(
	ctx context.Context,
	facebookID string,
	accessToken string,
	posts []clickhousemodels.FacebookCompetitorMinimalSharedPost,
) ([]clickhousemodels.FacebookCompetitorMinimalSharedPost, error) {
	if strings.TrimSpace(accessToken) == "" {
		return nil, fmt.Errorf("FacebookClient.GetCompetitorSharedFromPictures: access token is required")
	}
	if len(posts) == 0 {
		return nil, nil
	}

	ids := make([]string, 0, len(posts))
	seen := make(map[string]struct{}, len(posts))
	for _, post := range posts {
		id := strings.TrimSpace(post.SharedFromID)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, nil
	}

	const batchSize = 50
	appsecretProof := c.generateAppSecretProof(accessToken)
	byID := make(map[string]competitorPictureItem, len(ids))

	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		batch := ids[i:end]

		u, _ := url.Parse(fmt.Sprintf("%s%s", c.baseURL, fbAPIVersion))
		q := u.Query()
		q.Set("ids", strings.Join(batch, ","))
		q.Set("fields", "picture.type(large){url}")
		q.Set("access_token", accessToken)
		q.Set("appsecret_proof", appsecretProof)
		u.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("FacebookClient.GetCompetitorSharedFromPictures: failed to create request: %w", err)
		}

		body, status, err := c.doWithRetry(ctx, facebookID, req, "GetCompetitorSharedFromPictures")
		if err != nil {
			return nil, err
		}
		if status != http.StatusOK {
			return nil, fmt.Errorf("FacebookClient.GetCompetitorSharedFromPictures: unexpected http status after retry: %d", status)
		}

		var batchByID map[string]competitorPictureItem
		if err := json.Unmarshal(body, &batchByID); err != nil {
			return nil, fmt.Errorf("FacebookClient.GetCompetitorSharedFromPictures: failed to decode response: %w", err)
		}
		for k, v := range batchByID {
			byID[k] = v
		}
	}

	out := make([]clickhousemodels.FacebookCompetitorMinimalSharedPost, 0, len(posts))
	for _, post := range posts {
		item, ok := byID[strings.TrimSpace(post.SharedFromID)]
		if !ok || item.Picture == nil || item.Picture.Data == nil || strings.TrimSpace(item.Picture.Data.URL) == "" {
			continue
		}
		out = append(out, clickhousemodels.FacebookCompetitorMinimalSharedPost{
			FacebookID:    post.FacebookID,
			PostID:        post.PostID,
			SharedFromID:  post.SharedFromID,
			SharedFromPic: strings.TrimSpace(item.Picture.Data.URL),
			CreatedAt:     post.CreatedAt,
		})
	}

	return out, nil
}

/* =========================
   Core request logic
========================= */

// makeAPICall makes a Facebook API call and decodes the response
func (c *FacebookClient) makeAPICall(
	ctx context.Context,
	urlOrEndpoint string, // full URL (with http prefix) or empty string to use endpoint param
	endpoint string, // endpoint if urlOrEndpoint is empty
	params map[string]string,
	result interface{},
	accessToken string,
	caller string,
) error {

	var fullURL string

	// If urlOrEndpoint starts with http, use it as a full URL
	if strings.HasPrefix(urlOrEndpoint, "http") {
		fullURL = urlOrEndpoint
		// Add appsecret_proof if not already present
		if !strings.Contains(fullURL, "appsecret_proof=") {
			sep := "&"
			if !strings.Contains(fullURL, "?") {
				sep = "?"
			}
			fullURL = fmt.Sprintf("%s%sappsecret_proof=%s", fullURL, sep, c.generateAppSecretProof(accessToken))
		}
	} else {
		// Build URL from endpoint and params
		fullURL = buildCompURL(
			fmt.Sprintf("%s%s%s", c.baseURL, fbAPIVersion, endpoint),
			params,
			accessToken,
			c.generateAppSecretProof(accessToken),
		)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return fmt.Errorf("FacebookClient.makeAPICall: failed to create request: %w", err)
	}

	// Use shared doWithRetry for consistent retry logic, rate limiting, and error handling
	body, status, err := c.doWithRetry(ctx, "", req, caller)
	if err != nil {
		return err
	}

	if status != http.StatusOK {
		return parseAPIErrorFB(body, status)
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("FacebookClient.makeAPICall: failed to decode response: %w", err)
	}

	return nil
}

// buildCompURL constructs the full URL with access token, app secret proof and other params
func buildCompURL(
	base string,
	params map[string]string,
	token string,
	proof string,
) string {

	u, _ := url.Parse(base)
	q := u.Query()

	q.Set("access_token", token)
	q.Set("appsecret_proof", proof)

	for k, v := range params {
		q.Set(k, v)
	}

	u.RawQuery = q.Encode()
	return u.String()
}

// parseAPIErrorFB parses Facebook API error responses
func parseAPIErrorFB(body []byte, status int) error {
	var errResp models.ErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != nil {
		return fmt.Errorf(
			"parseAPIErrorFB: facebook api error (%d): %s - %s",
			errResp.Error.Code,
			errResp.Error.Type,
			errResp.Error.Message,
		)
	}
	return fmt.Errorf("parseAPIErrorFB: facebook api error (%d): %s", status, string(body))
}

func pickCompetitorFacebookImage(item competitorPostURLsItem) string {
	chosen := competitorFacebookImage{}
	if item.Attachments != nil {
		for _, att := range item.Attachments.Data {
			if att.Media != nil && att.Media.Image != nil && att.Media.Image.Src != "" {
				chosen = bestCompetitorFacebookImage(chosen, competitorFacebookImage{
					Src: att.Media.Image.Src,
					W:   att.Media.Image.Width,
					H:   att.Media.Image.Height,
				})
			}
			if att.Subattachments == nil {
				continue
			}
			for _, child := range att.Subattachments.Data {
				if child.Media != nil && child.Media.Image != nil && child.Media.Image.Src != "" {
					chosen = bestCompetitorFacebookImage(chosen, competitorFacebookImage{
						Src: child.Media.Image.Src,
						W:   child.Media.Image.Width,
						H:   child.Media.Image.Height,
					})
				}
			}
		}
	}

	if strings.TrimSpace(chosen.Src) != "" {
		return chosen.Src
	}
	return strings.TrimSpace(item.FullPicture)
}

type competitorFacebookImage struct {
	Src string
	W   int
	H   int
}

type competitorPostURLsItem struct {
	ID           string `json:"id"`
	PermalinkURL string `json:"permalink_url"`
	FullPicture  string `json:"full_picture"`
	// ChildAttachments covers carousel posts that were ingested via the
	// child_attachments API field (MediaID = child.ID).
	ChildAttachments []struct {
		ID    string `json:"id"`
		Media *struct {
			Image *struct {
				Src           string `json:"src"`
				Width, Height int
			} `json:"image"`
		} `json:"media"`
	} `json:"child_attachments"`
	Attachments *struct {
		Data []struct {
			Target *struct {
				ID string `json:"id"`
			} `json:"target"`
			Media *struct {
				Image *struct {
					Src           string `json:"src"`
					Width, Height int
				} `json:"image"`
			} `json:"media"`
			Subattachments *struct {
				Data []struct {
					Target *struct {
						ID string `json:"id"`
					} `json:"target"`
					Media *struct {
						Image *struct {
							Src           string `json:"src"`
							Width, Height int
						} `json:"image"`
					} `json:"media"`
				} `json:"data"`
			} `json:"subattachments"`
		} `json:"data"`
	} `json:"attachments"`
}

type competitorPictureItem struct {
	Picture *models.Picture `json:"picture"`
}

func bestCompetitorFacebookImage(a, b competitorFacebookImage) competitorFacebookImage {
	if a.W*a.H >= b.W*b.H {
		return a
	}
	return b
}

func matchCompetitorMediaAsset(
	asset clickhousemodels.FacebookCompetitorMinimalMediaAsset,
	item competitorPostURLsItem,
) (clickhousemodels.FacebookCompetitorMinimalMediaAsset, bool) {
	// Path A: carousel posts ingested via child_attachments (MediaID = child.ID).
	for _, child := range item.ChildAttachments {
		if strings.TrimSpace(child.ID) != asset.MediaID {
			continue
		}
		if child.Media == nil || child.Media.Image == nil || strings.TrimSpace(child.Media.Image.Src) == "" {
			return clickhousemodels.FacebookCompetitorMinimalMediaAsset{}, false
		}
		return clickhousemodels.FacebookCompetitorMinimalMediaAsset{
			PageID:    asset.PageID,
			PostID:    asset.PostID,
			MediaID:   asset.MediaID,
			Link:      strings.TrimSpace(child.Media.Image.Src),
			CreatedAt: asset.CreatedAt,
		}, true
	}

	// Path B: attachments / subattachments (MediaID = target.id or md5 hash).
	if item.Attachments == nil || len(item.Attachments.Data) == 0 {
		// Fallback: use full_picture for posts that have no separate attachment object
		// (e.g. plain photo or video posts where Facebook omits the attachments field).
		if fp := strings.TrimSpace(item.FullPicture); fp != "" {
			if asset.MediaID == generateCompetitorMediaID(item.ID, 0) {
				return clickhousemodels.FacebookCompetitorMinimalMediaAsset{
					PageID:    asset.PageID,
					PostID:    asset.PostID,
					MediaID:   asset.MediaID,
					Link:      fp,
					CreatedAt: asset.CreatedAt,
				}, true
			}
		}
		return clickhousemodels.FacebookCompetitorMinimalMediaAsset{}, false
	}

	attachment := item.Attachments.Data[0]
	if attachment.Subattachments != nil && len(attachment.Subattachments.Data) > 0 {
		for i, child := range attachment.Subattachments.Data {
			// Try both strategies: the API may return target.id during refresh but not
			// during ingestion (or vice versa), causing systematic ID mismatches if we
			// only use one approach. Matching either means we found the right asset.
			targetID := ""
			if child.Target != nil {
				targetID = strings.TrimSpace(child.Target.ID)
			}
			hashID := generateCompetitorMediaID(item.ID, i)
			if targetID != asset.MediaID && hashID != asset.MediaID {
				continue
			}
			if child.Media == nil || child.Media.Image == nil || strings.TrimSpace(child.Media.Image.Src) == "" {
				return clickhousemodels.FacebookCompetitorMinimalMediaAsset{}, false
			}
			return clickhousemodels.FacebookCompetitorMinimalMediaAsset{
				PageID:    asset.PageID,
				PostID:    asset.PostID,
				MediaID:   asset.MediaID,
				Link:      strings.TrimSpace(child.Media.Image.Src),
				CreatedAt: asset.CreatedAt,
			}, true
		}
		return clickhousemodels.FacebookCompetitorMinimalMediaAsset{}, false
	}

	// Single attachment (no subattachments). Use media.image.src if present, else fall
	// back to full_picture for video posts and other types that lack a media.image field.
	imageURL := ""
	if attachment.Media != nil && attachment.Media.Image != nil {
		imageURL = strings.TrimSpace(attachment.Media.Image.Src)
	}
	if imageURL == "" {
		imageURL = strings.TrimSpace(item.FullPicture)
	}
	if imageURL == "" {
		return clickhousemodels.FacebookCompetitorMinimalMediaAsset{}, false
	}
	if asset.MediaID != generateCompetitorMediaID(item.ID, 0) {
		return clickhousemodels.FacebookCompetitorMinimalMediaAsset{}, false
	}
	return clickhousemodels.FacebookCompetitorMinimalMediaAsset{
		PageID:    asset.PageID,
		PostID:    asset.PostID,
		MediaID:   asset.MediaID,
		Link:      imageURL,
		CreatedAt: asset.CreatedAt,
	}, true
}

func generateCompetitorMediaID(postID string, index int) string {
	hash := md5.Sum([]byte(fmt.Sprintf("%s_%d", postID, index)))
	return hex.EncodeToString(hash[:])
}
