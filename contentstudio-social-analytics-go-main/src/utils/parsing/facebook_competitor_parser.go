package parsing

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	apiModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	models "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

// FacebookClient interface for making API calls
type FacebookClient interface {
	GetCompetitorSharedPostDetails(ctx context.Context, parentID string, accessToken string) (*apiModels.Post, error)
	GetCompetitorPagePicture(ctx context.Context, pageID string, accessToken string) (*apiModels.Picture, error)
}

// FacebookCompetitorParser handles parsing of Facebook API responses to ClickHouse models
type FacebookCompetitorParser struct {
	pageID      string
	pageName    string
	fbClient    FacebookClient
	accessToken string
}

// NewFacebookCompetitorParser creates a new parser
func NewFacebookCompetitorParser(pageID, pageName string, fbClient FacebookClient, accessToken string) *FacebookCompetitorParser {
	return &FacebookCompetitorParser{
		pageID:      pageID,
		pageName:    pageName,
		fbClient:    fbClient,
		accessToken: accessToken,
	}
}

// ParsePageInsights parses page details into competitor insights
func (p *FacebookCompetitorParser) ParsePageInsights(pageDetails *apiModels.FacebookPageDetails, picture *apiModels.Picture) *models.FacebookCompetitorInsights {
	now := time.Now().UTC()
	recordID := generateRecordID(p.pageID, now)

	insight := &models.FacebookCompetitorInsights{
		RecordID:         recordID,
		PageID:           p.pageID,
		PageName:         p.pageName,
		TotalFanCount:    pageDetails.FanCount,
		TalkingAboutThis: pageDetails.TalkingAboutCount,
		Biography:        pageDetails.About,
		PageCategory:     pageDetails.Category,
		FollowersCount:   pageDetails.FollowersCount,
		Emails:           pageDetails.Emails,
		Birthday:         pageDetails.Birthday,
		WereHereCount:    pageDetails.WereHereCount,
		Permalink:        pageDetails.Link,
		InsertedAt:       now.Truncate(time.Hour).Add(0), // Set to start of hour
	}

	if picture != nil && picture.Data != nil {
		insight.ProfilePictureURL = picture.Data.URL
	}

	if pageDetails.Cover != nil {
		insight.CoverPhotoURL = pageDetails.Cover.Source
	}

	return insight
}

// ParsePosts parses Facebook posts into competitor posts and media assets
func (p *FacebookCompetitorParser) ParsePosts(ctx context.Context, posts []*apiModels.Post, pageDetails *apiModels.FacebookPageDetails) ([]*models.FacebookCompetitorPosts, []*models.FacebookCompetitorMediaAssets) {
	var competitorPosts []*models.FacebookCompetitorPosts
	var mediaAssets []*models.FacebookCompetitorMediaAssets

	for _, post := range posts {
		competitorPost := p.parsePost(ctx, post, pageDetails)
		if competitorPost == nil {
			continue
		}
		competitorPosts = append(competitorPosts, competitorPost)

		assets := p.parseMediaAssets(post)
		mediaAssets = append(mediaAssets, assets...)
	}

	return competitorPosts, mediaAssets
}

// parsePost parses a single Facebook post. Returns nil if created_time cannot be parsed,
// since a zero CreatedAt would write the row into the wrong partition and create
// a permanent cross-partition duplicate that FINAL cannot merge.
func (p *FacebookCompetitorParser) parsePost(ctx context.Context, post *apiModels.Post, pageDetails *apiModels.FacebookPageDetails) *models.FacebookCompetitorPosts {
	createdTime, err := time.Parse("2006-01-02T15:04:05-0700", post.CreatedTime)
	if err != nil {
		createdTime, err = time.Parse(time.RFC3339, post.CreatedTime)
	}
	if err != nil || createdTime.IsZero() {
		return nil
	}
	createdTime = createdTime.UTC()

	competitorPost := &models.FacebookCompetitorPosts{
		FacebookID: p.pageID,
		PostID:     post.ID,
		PageName:   p.pageName,
		CreatedAt:  createdTime,
		InsertedAt: time.Now().UTC(),
		DayOfWeek:  createdTime.Weekday().String(),
		HourOfDay:  int64(createdTime.Hour()),
		Caption:    post.Message,
		Permalink:  post.PermalinkURL,
		StatusType: post.StatusType,
	}

	// Set page details
	if pageDetails != nil {
		competitorPost.Biography = pageDetails.About
		competitorPost.FollowersCount = pageDetails.FollowersCount
		competitorPost.PageCategory = pageDetails.Category
		competitorPost.FanCount = pageDetails.FanCount
	}

	// Parse reactions
	engagements := p.parseEngagements(post)
	competitorPost.Like = engagements["like"]
	competitorPost.Love = engagements["love"]
	competitorPost.Haha = engagements["haha"]
	competitorPost.Wow = engagements["wow"]
	competitorPost.Sad = engagements["sad"]
	competitorPost.Angry = engagements["angry"]
	competitorPost.Comments = engagements["comments"]
	competitorPost.Shares = engagements["shares"]
	competitorPost.PostEngagement = engagements["post_engagement"]
	competitorPost.TotalPostReactions = engagements["total_reactions"]

	// Parse media type
	mediaType, statusType := p.determineMediaType(post)
	competitorPost.MediaType = mediaType
	if statusType != "" {
		competitorPost.StatusType = statusType
	}

	// Parse shared post information (with API calls if needed)
	p.parseSharedPostInfo(ctx, post, competitorPost)

	// Extract hashtags
	competitorPost.Hashtags = extractHashtags(post.Message)

	return competitorPost
}

// parseSharedPostInfo extracts shared post information, making API calls if needed
func (p *FacebookCompetitorParser) parseSharedPostInfo(ctx context.Context, post *apiModels.Post, competitorPost *models.FacebookCompetitorPosts) {
	// Check if this is a shared post (has ParentID)
	if post.ParentID == "" {
		return
	}

	// If we have fbClient, fetch the original post details
	if p.fbClient != nil && p.accessToken != "" {
		sharedPost, err := p.fbClient.GetCompetitorSharedPostDetails(ctx, post.ParentID, p.accessToken)
		if err == nil && sharedPost != nil {
			// Extract information from the original post
			if sharedPost.From != nil {
				competitorPost.SharedFromName = sharedPost.From.Name
				competitorPost.SharedFromID = sharedPost.From.ID

				// Fetch the original poster's profile picture
				picture, err := p.fbClient.GetCompetitorPagePicture(ctx, sharedPost.From.ID, p.accessToken)
				if err == nil && picture != nil && picture.Data != nil {
					competitorPost.SharedFromPic = picture.Data.URL
				}
			}

			// Get the original post's creation time
			if sharedPost.CreatedTime != "" {
				originalCreatedTime, err := time.Parse("2006-01-02T15:04:05-0700", sharedPost.CreatedTime)
				if err == nil {
					competitorPost.SharedCreatedAt = originalCreatedTime
				}
			}

			return
		}
	}

	// Fallback: Extract from attachments if API call failed or client not available
	if post.Attachments == nil || len(post.Attachments.Data) == 0 {
		return
	}

	attachment := post.Attachments.Data[0]

	// Get shared post creator information from attachment
	if attachment.UnshimmedURL != "" {
		competitorPost.SharedFromID = p.extractPageIDFromURL(attachment.UnshimmedURL)
	}

	if attachment.Title != "" {
		competitorPost.SharedFromName = attachment.Title
	}

	if attachment.Media != nil && attachment.Media.Image != nil {
		competitorPost.SharedFromPic = attachment.Media.Image.Src
	} else if post.FullPicture != "" {
		competitorPost.SharedFromPic = post.FullPicture
	}

	// Use share's creation time as fallback
	competitorPost.SharedCreatedAt = competitorPost.CreatedAt
}

// extractPageIDFromURL extracts page ID from a Facebook URL
func (p *FacebookCompetitorParser) extractPageIDFromURL(url string) string {
	parts := strings.Split(url, "/")
	for i, part := range parts {
		if part == "facebook.com" && i+1 < len(parts) {
			pageIdentifier := parts[i+1]
			if pageIdentifier != "photo.php" && pageIdentifier != "permalink.php" &&
				pageIdentifier != "posts" && pageIdentifier != "videos" {
				return pageIdentifier
			}
		}
	}

	if strings.Contains(url, "id=") {
		parts := strings.Split(url, "id=")
		if len(parts) > 1 {
			id := strings.Split(parts[1], "&")[0]
			return id
		}
	}

	return ""
}

// parseEngagements calculates engagement metrics from a post
func (p *FacebookCompetitorParser) parseEngagements(post *apiModels.Post) map[string]int64 {
	engagements := make(map[string]int64)

	if post.Like != nil && post.Like.Summary != nil {
		engagements["like"] = post.Like.Summary.TotalCount
	}
	if post.Love != nil && post.Love.Summary != nil {
		engagements["love"] = post.Love.Summary.TotalCount
	}
	if post.Haha != nil && post.Haha.Summary != nil {
		engagements["haha"] = post.Haha.Summary.TotalCount
	}
	if post.Wow != nil && post.Wow.Summary != nil {
		engagements["wow"] = post.Wow.Summary.TotalCount
	}
	if post.Sad != nil && post.Sad.Summary != nil {
		engagements["sad"] = post.Sad.Summary.TotalCount
	}
	if post.Angry != nil && post.Angry.Summary != nil {
		engagements["angry"] = post.Angry.Summary.TotalCount
	}
	if post.Comments != nil && post.Comments.Summary != nil {
		engagements["comments"] = post.Comments.Summary.TotalCount
	}
	if post.Shares != nil {
		engagements["shares"] = post.Shares.Count
	}

	// Calculate totals
	totalReactions := engagements["like"] + engagements["love"] + engagements["haha"] +
		engagements["wow"] + engagements["sad"] + engagements["angry"]
	engagements["total_reactions"] = totalReactions

	totalEngagement := totalReactions + engagements["comments"] + engagements["shares"]
	engagements["post_engagement"] = totalEngagement

	return engagements
}

// determineMediaType determines the media type of a post
func (p *FacebookCompetitorParser) determineMediaType(post *apiModels.Post) (string, string) {
	if len(post.ChildAttachments) > 0 {
		if post.ParentID != "" {
			return "share", ""
		}
		return "carousel", ""
	}

	if post.Attachments == nil || len(post.Attachments.Data) == 0 {
		if post.ParentID != "" {
			return "share", "text"
		}
		return "text", ""
	}

	attachment := post.Attachments.Data[0]

	switch post.StatusType {
	case "added_video":
		return "videos", ""
	case "added_photos":
		return "image", ""
	case "shared_story":
		return "link", ""
	case "mobile_status_update":
		if post.ParentID == "" {
			return "others", ""
		}
	}

	// Check for subattachments (carousel)
	if attachment.Subattachments != nil && len(attachment.Subattachments.Data) > 0 {
		return "carousel", ""
	}

	if post.ParentID != "" {
		return "share", p.mapMediaType(attachment.Type)
	}

	// Map by attachment type
	return p.mapMediaType(attachment.Type), ""
}

// mapMediaType maps Facebook attachment types to our media types
func (p *FacebookCompetitorParser) mapMediaType(attachmentType string) string {
	switch attachmentType {
	case "multi_share_no_end_card", "album":
		return "carousel"
	case "photo":
		return "image"
	case "video", "video_inline":
		return "videos"
	case "link", "share":
		return "link"
	default:
		return "others"
	}
}

// determineAssetMediaType determines media type for individual assets
func (p *FacebookCompetitorParser) determineAssetMediaType(post *apiModels.Post, attachmentData *apiModels.AttachmentData, subAttachment *apiModels.AttachmentData) string {
	// Check status type first for videos
	if post.StatusType == "added_video" {
		return "video"
	}

	// Check media_type field in attachment
	if attachmentData != nil && attachmentData.MediaType != "" {
		switch attachmentData.MediaType {
		case "photo", "album":
			return "image"
		case "video", "video_inline":
			return "video"
		case "link":
			return "link"
		}
	}

	// Check type field
	var typeToCheck string
	if subAttachment != nil {
		typeToCheck = subAttachment.Type
	} else if attachmentData != nil {
		typeToCheck = attachmentData.Type
	}

	switch typeToCheck {
	case "photo":
		return "image"
	case "album":
		return "image"
	case "video", "video_inline":
		return "video"
	case "link", "share":
		return "link"
	}

	// Default to image if media exists
	if (subAttachment != nil && subAttachment.Media != nil && subAttachment.Media.Image != nil) ||
		(attachmentData != nil && attachmentData.Media != nil && attachmentData.Media.Image != nil) {
		return "image"
	}

	return "others"
}

// determineChildAttachmentMediaType determines media type for a specific child attachment
// This considers the attachment data from the parent post to determine the correct type
func (p *FacebookCompetitorParser) determineChildAttachmentMediaType(post *apiModels.Post, attachmentData *apiModels.AttachmentData) string {
	// For child attachments, we need to check the parent attachment's media type
	// since individual child attachments don't have their own media type field

	// Check post status type first
	if post.StatusType == "added_video" {
		return "video"
	}

	// Check parent attachment's media_type
	if attachmentData != nil && attachmentData.MediaType != "" {
		switch attachmentData.MediaType {
		case "photo", "album":
			return "image"
		case "video", "video_inline":
			return "video"
		case "link":
			return "link"
		}
	}

	// Check parent attachment's type
	if attachmentData != nil && attachmentData.Type != "" {
		switch attachmentData.Type {
		case "photo", "album":
			return "image"
		case "video", "video_inline":
			return "video"
		case "link", "share":
			return "link"
		}
	}

	// Try to infer from the parent attachment's media structure
	if attachmentData != nil && attachmentData.Media != nil {
		if attachmentData.Media.Image != nil {
			return "image"
		}
		if attachmentData.Media.Source != "" {
			// Source typically indicates video
			return "video"
		}
	}

	// Default to image for carousel items
	return "image"
}

// parseMediaAssets parses media assets from a post
func (p *FacebookCompetitorParser) parseMediaAssets(post *apiModels.Post) []*models.FacebookCompetitorMediaAssets {
	var assets []*models.FacebookCompetitorMediaAssets
	createdTime, err := time.Parse("2006-01-02T15:04:05-0700", post.CreatedTime)
	if err != nil {
		createdTime, err = time.Parse(time.RFC3339, post.CreatedTime)
	}
	if err != nil || createdTime.IsZero() {
		return nil
	}
	createdTime = createdTime.UTC()

	// Handle child attachments (carousel)
	if len(post.ChildAttachments) > 0 {
		var attachmentData *apiModels.AttachmentData
		if post.Attachments != nil && len(post.Attachments.Data) > 0 {
			attachmentData = post.Attachments.Data[0]
		}

		for _, child := range post.ChildAttachments {
			asset := &models.FacebookCompetitorMediaAssets{
				MediaID:     child.ID,
				PostID:      post.ID,
				PageID:      p.pageID,
				Caption:     child.Caption,
				Description: child.Description,
				Link:        child.Picture,
				CreatedAt:   createdTime,
				InsertedAt:  time.Now().UTC(),
				AssetType:   p.determineChildAttachmentMediaType(post, attachmentData),
			}

			// Set CTA information
			if child.CallToAction != nil {
				asset.CTAType = child.CallToAction.Type
				if child.CallToAction.Value != nil && child.CallToAction.Value.Link != "" {
					asset.CallToAction = child.CallToAction.Value.Link
				}
			}
			// Fallback to child.Link if no CTA was set
			if asset.CallToAction == "" && child.Link != "" {
				asset.CallToAction = child.Link
			}

			assets = append(assets, asset)
		}
		return assets
	}

	// Handle regular attachments
	if post.Attachments == nil || len(post.Attachments.Data) == 0 {
		return assets
	}

	attachmentData := post.Attachments.Data[0]

	// Handle subattachments (carousel without child_attachments)
	if attachmentData.Subattachments != nil && len(attachmentData.Subattachments.Data) > 0 {
		for i, subAttachment := range attachmentData.Subattachments.Data {
			mediaID := ""
			if subAttachment.Target != nil && subAttachment.Target.ID != "" {
				mediaID = subAttachment.Target.ID
			} else {
				// Generate hash for media ID
				mediaID = p.generateMediaID(post.ID, i)
			}

			asset := &models.FacebookCompetitorMediaAssets{
				MediaID:     mediaID,
				PostID:      post.ID,
				PageID:      p.pageID,
				Description: subAttachment.Description,
				CreatedAt:   createdTime,
				InsertedAt:  time.Now().UTC(),
				AssetType:   p.determineAssetMediaType(post, attachmentData, subAttachment),
			}

			if subAttachment.Media != nil && subAttachment.Media.Image != nil {
				asset.Link = subAttachment.Media.Image.Src
			}

			if attachmentData.Title != "" {
				asset.Caption = attachmentData.Title
			}

			assets = append(assets, asset)
		}
	} else {
		// Single attachment
		mediaID := p.generateMediaID(post.ID, 0)

		asset := &models.FacebookCompetitorMediaAssets{
			MediaID:     mediaID,
			PostID:      post.ID,
			PageID:      p.pageID,
			Description: attachmentData.Description,
			CreatedAt:   createdTime,
			InsertedAt:  time.Now().UTC(),
			AssetType:   p.determineAssetMediaType(post, attachmentData, nil),
		}

		if attachmentData.Media != nil && attachmentData.Media.Image != nil {
			asset.Link = attachmentData.Media.Image.Src
		}

		if attachmentData.Title != "" {
			asset.Caption = attachmentData.Title
		}

		assets = append(assets, asset)
	}

	return assets
}

// extractHashtags extracts hashtags from text
func extractHashtags(text string) []string {
	if text == "" {
		return []string{}
	}

	pattern := regexp.MustCompile(`(?i)#([a-z0-9_]+)`)
	matches := pattern.FindAllStringSubmatch(text, -1)

	hashtags := make([]string, 0, len(matches))
	for _, match := range matches {
		hashtags = append(hashtags, match[1]) // without #
	}

	return hashtags
}

// generateRecordID generates a record ID for insights
func generateRecordID(pageID string, timestamp time.Time) string {
	date := timestamp.Format("2006-01-02")
	str := fmt.Sprintf("%s_%s", pageID, date)
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}

// generateMediaID generates a media ID for assets
func (p *FacebookCompetitorParser) generateMediaID(postID string, index int) string {
	str := fmt.Sprintf("%s_%d", postID, index)
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}
