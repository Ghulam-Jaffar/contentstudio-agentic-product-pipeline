package parsing

import (
	"crypto/md5"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// FacebookParser handles parsing of raw Facebook posts into structured data
type FacebookParser struct {
	MediaTypeMapping map[string]string
	InsightFields    []string
	ReactionFields   []string
	ValueFields      []string
	log              *logger.Logger
}

// NewFacebookParser creates a new Facebook parser with predefined mappings
func NewFacebookParser() *FacebookParser {
	log := logger.New("info")
	log.Info().
		Str("module", "facebook_parser").
		Msg("Initializing FacebookParser")

	parser := &FacebookParser{
		MediaTypeMapping: map[string]string{
			"multi_share_no_end_card": "carousel",
			"photo":                   "images",
			"album":                   "images",
			"video_inline":            "videos",
			"link":                    "link",
			"share":                   "link",
		},
		InsightFields: []string{
			"post_impressions",
			"post_impressions_unique",
		},
		ReactionFields: []string{
			"total", "like", "love", "haha", "wow", "sad", "angry", "comments",
		},
		ValueFields: []string{
			"post_metadata", "message_tags", "full_picture", "updated_time",
			"created_time", "status_type",
		},
		log: log,
	}

	log.Info().
		Str("module", "facebook_parser").
		Int("media_type_mappings", len(parser.MediaTypeMapping)).
		Int("insight_fields", len(parser.InsightFields)).
		Int("reaction_fields", len(parser.ReactionFields)).
		Int("value_fields", len(parser.ValueFields)).
		Msg("FacebookParser initialized successfully")

	return parser
}

// ParsePost converts a raw Facebook post into a parsed post and media assets
func (fp *FacebookParser) ParsePost(rawPost kafkamodels.RawFacebookPost, pageID, pageName, workspaceID string) (*kafkamodels.ParsedFacebookPost, []kafkamodels.ParsedFacebookMediaAsset, error) {
	startTime := time.Now()

	fp.log.Info().
		Str("module", "facebook_parser").
		Str("function", "ParsePost").
		Str("post_id", rawPost.ID).
		Str("page_id", pageID).
		Str("page_name", pageName).
		Str("workspace_id", workspaceID).
		Bool("has_created_time", !rawPost.CreatedTime.Time.IsZero()).
		Bool("has_updated_time", !rawPost.UpdatedTime.Time.IsZero()).
		Str("status_type", rawPost.StatusType).
		Msg("Starting Facebook post parsing")

	// Initialize parsed post
	//_, parsePostID := splitComposite(rawPost.ID) // it returns page and post id
	//if parsePostID == "" {
	//	fp.log.Error().
	//		Str("module", "facebook_parser").
	//		Str("function", "ParsePost").
	//		Str("raw_post_id", rawPost.ID).
	//		Msg("Failed to extract post ID from composite ID")
	//	return nil, nil, fmt.Errorf("invalid post ID format: %s", rawPost.ID)
	//}

	fp.log.Debug().
		Str("module", "facebook_parser").
		Str("function", "ParsePost").
		Str("raw_post_id", rawPost.ID).
		//Str("extracted_post_id", parsePostID).
		Msg("Successfully extracted post ID from composite ID")

	post := &kafkamodels.ParsedFacebookPost{
		PageID:           pageID,
		PageName:         pageName,
		PostID:           rawPost.ID,
		Permalink:        rawPost.PermalinkURL,
		Caption:          rawPost.Message,
		SavingTime:       time.Now().UTC(),
		TotalEngagement:  0,
		TotalImpressions: 0,
		Total:            0,
		Comments:         0,
	}

	// Parse created time and set day/hour info
	if !rawPost.CreatedTime.Time.IsZero() {
		post.CreatedTime = rawPost.CreatedTime.Time
		post.DayOfWeek = rawPost.CreatedTime.Time.Weekday().String()
		post.HourOfDay = int32(rawPost.CreatedTime.Time.Hour())

		fp.log.Debug().
			Str("module", "facebook_parser").
			Str("function", "ParsePost").
			Str("post_id", rawPost.ID).
			Time("created_time", post.CreatedTime).
			Str("day_of_week", post.DayOfWeek).
			Int32("hour_of_day", post.HourOfDay).
			Msg("Parsed post created time and derived scheduling info")
	} else {
		fp.log.Warn().
			Str("module", "facebook_parser").
			Str("function", "ParsePost").
			Str("post_id", rawPost.ID).
			Msg("Post has no created time information")
	}

	// Parse updated time
	if !rawPost.UpdatedTime.Time.IsZero() {
		post.UpdatedTime = rawPost.UpdatedTime.Time
		fp.log.Debug().
			Str("module", "facebook_parser").
			Str("function", "ParsePost").
			Str("post_id", rawPost.ID).
			Time("updated_time", post.UpdatedTime).
			Msg("Parsed post updated time")
	}

	// Set published by information
	if rawPost.AdminCreator != nil {
		post.PublishedBy = rawPost.AdminCreator.Name
		post.PublishedByURL = fmt.Sprintf("https://facebook.com/%s", rawPost.AdminCreator.ID)
		post.Category = "" // Category would come from admin_creator if available

		fp.log.Debug().
			Str("module", "facebook_parser").
			Str("function", "ParsePost").
			Str("post_id", rawPost.ID).
			Str("published_by", post.PublishedBy).
			Str("published_by_id", rawPost.AdminCreator.ID).
			Str("published_by_url", post.PublishedByURL).
			Msg("Parsed admin creator information")
	}

	// Parse message tags
	post.MessageTags = fp.extractMessageTags(rawPost.MessageTags)
	fp.log.Debug().
		Str("module", "facebook_parser").
		Str("function", "ParsePost").
		Str("post_id", rawPost.ID).
		Int("message_tags_count", len(post.MessageTags)).
		Strs("message_tags", post.MessageTags).
		Msg("Parsed message tags")

	// Determine media type from status_type
	post.MediaType = fp.getMediaTypeFromStatus(rawPost.StatusType)
	post.StatusType = rawPost.StatusType

	fp.log.Debug().
		Str("module", "facebook_parser").
		Str("function", "ParsePost").
		Str("post_id", rawPost.ID).
		Str("raw_status_type", rawPost.StatusType).
		Str("mapped_media_type", post.MediaType).
		Msg("Determined media type from status")

	// Parse attachments and media assets
	var mediaAssets []kafkamodels.ParsedFacebookMediaAsset
	var isChild bool

	// Handle child attachments (carousel posts)
	if len(rawPost.ChildAttachments) > 0 {
		post.MediaType = "carousel"
		isChild = true

		fp.log.Info().
			Str("module", "facebook_parser").
			Str("function", "ParsePost").
			Str("post_id", rawPost.ID).
			Int("child_attachments_count", len(rawPost.ChildAttachments)).
			Msg("Processing carousel post with child attachments")

		mediaAssets = fp.parseChildAttachments(rawPost, pageID)

		fp.log.Info().
			Str("module", "facebook_parser").
			Str("function", "ParsePost").
			Str("post_id", rawPost.ID).
			Int("media_assets_from_children", len(mediaAssets)).
			Msg("Completed processing child attachments")
	}

	// Handle regular attachments
	if rawPost.Attachments != nil && len(rawPost.Attachments.Data) > 0 {
		attachment := rawPost.Attachments.Data[0]

		fp.log.Debug().
			Str("module", "facebook_parser").
			Str("function", "ParsePost").
			Str("post_id", rawPost.ID).
			Int("attachments_count", len(rawPost.Attachments.Data)).
			Str("first_attachment_type", attachment.Type).
			Bool("has_media", attachment.Media != nil).
			Msg("Processing regular attachments")

		// Set post metadata and link
		post.PostMetadata = attachment.Caption
		post.Link = attachment.Link

		// Set full picture based on media type
		if attachment.Media != nil {
			if attachment.Media.Image != nil {
				post.FullPicture = attachment.Media.Image.Src
				if post.FullPicture == "" {
					post.FullPicture = attachment.Media.Src
				}
			}

			fp.log.Debug().
				Str("module", "facebook_parser").
				Str("function", "ParsePost").
				Str("post_id", rawPost.ID).
				Str("full_picture", post.FullPicture).
				Bool("used_image_src", attachment.Media.Image != nil && attachment.Media.Image.Src != "").
				Msg("Set full picture from attachment media")
		}

		// Parse media assets if not already parsed from child attachments
		if !isChild {
			additionalAssets := fp.parseAttachments(rawPost, pageID)
			mediaAssets = append(mediaAssets, additionalAssets...)

			fp.log.Debug().
				Str("module", "facebook_parser").
				Str("function", "ParsePost").
				Str("post_id", rawPost.ID).
				Int("additional_assets", len(additionalAssets)).
				Int("total_assets", len(mediaAssets)).
				Msg("Added media assets from regular attachments")
		}

		// Handle description with truncation
		if attachment.Description != "" {
			if len(attachment.Description) > 25000 {
				post.Description = attachment.Description[:25000] + "..."
				fp.log.Warn().
					Str("module", "facebook_parser").
					Str("function", "ParsePost").
					Str("post_id", rawPost.ID).
					Int("original_length", len(attachment.Description)).
					Int("truncated_length", len(post.Description)).
					Msg("Truncated post description due to length limit")
			} else {
				post.Description = attachment.Description
			}
		}

		// Update status type for share posts
		if post.MediaType == "share" && attachment.Type != "" {
			if mappedType, exists := fp.MediaTypeMapping[attachment.Type]; exists {
				post.StatusType = mappedType
				fp.log.Debug().
					Str("module", "facebook_parser").
					Str("function", "ParsePost").
					Str("post_id", rawPost.ID).
					Str("original_attachment_type", attachment.Type).
					Str("mapped_status_type", mappedType).
					Msg("Updated status type for share post")
			}
		}
	}

	// Handle posts without attachments
	if rawPost.Attachments == nil && post.MediaType == "share" {
		post.MediaType = "text"
		fp.log.Debug().
			Str("module", "facebook_parser").
			Str("function", "ParsePost").
			Str("post_id", rawPost.ID).
			Msg("Changed media type from share to text for post without attachments")
	}

	// Extract video ID for video posts
	if rawPost.StatusType == "added_video" && post.Link != "" {
		post.VideoID = fp.extractVideoID(post.Link)
		fp.log.Debug().
			Str("module", "facebook_parser").
			Str("function", "ParsePost").
			Str("post_id", rawPost.ID).
			Str("link", post.Link).
			Str("video_id", post.VideoID).
			Msg("Extracted video ID from video post link")
	}

	// Parse insights
	fp.parsePostInsights(rawPost, post)

	// Parse reactions and engagement
	fp.ParsePostReactions(rawPost, post)

	// Parse shares
	if rawPost.Shares != nil {
		post.Shares = int32(rawPost.Shares.Count)
		post.TotalEngagement += int64(post.Shares)

		fp.log.Debug().
			Str("module", "facebook_parser").
			Str("function", "ParsePost").
			Str("post_id", rawPost.ID).
			Int32("shares_count", post.Shares).
			Int64("total_engagement", post.TotalEngagement).
			Msg("Parsed shares data")
	}

	// Handle shared posts
	if rawPost.ParentID != "" {
		fp.log.Info().
			Str("module", "facebook_parser").
			Str("function", "ParsePost").
			Str("post_id", rawPost.ID).
			Str("parent_id", rawPost.ParentID).
			Msg("Detected shared post - additional shared post processing would go here")
		// TODO: Implement shared post information retrieval
		// This would require additional API calls similar to get_shared_from in Python
	}

	if strings.Contains(post.Permalink, "reel") {
		post.MediaType = "reels"
		fp.log.Debug().
			Str("module", "facebook_parser").
			Str("function", "ParsePost").
			Str("post_id", rawPost.ID).
			Str("permalink", post.Permalink).
			Msg("Identified post as reel based on permalink")
	}

	if post.MediaType == "images" || post.MediaType == "videos" || post.MediaType == "reels" {
		mediaAsset := kafkamodels.ParsedFacebookMediaAsset{
			PostID:       rawPost.ID,
			PageID:       pageID,
			MediaID:      post.VideoID,
			AssetType:    post.MediaType,
			CallToAction: post.Permalink,
			Link:         post.FullPicture,
			Caption:      post.Caption,
			Description:  post.PostID,
			CreatedAt:    post.CreatedTime,
			InsertedAt:   post.SavingTime,
		}
		mediaAssets = append(mediaAssets, mediaAsset)

		fp.log.Debug().
			Str("module", "facebook_parser").
			Str("function", "ParsePost").
			Str("post_id", rawPost.ID).
			Str("media_type", post.MediaType).
			Str("media_id", mediaAsset.MediaID).
			Msg("Added default media asset for media post")
	}

	elapsed := time.Since(startTime)

	fp.log.Info().
		Str("module", "facebook_parser").
		Str("function", "ParsePost").
		Str("post_id", rawPost.ID).
		Str("page_id", pageID).
		Str("media_type", post.MediaType).
		Int("media_assets_count", len(mediaAssets)).
		Int64("total_engagement", post.TotalEngagement).
		Int64("total_impressions", post.TotalImpressions).
		Dur("processing_time", elapsed).
		Msg("Completed Facebook post parsing successfully")

	return post, mediaAssets, nil
}

// extractMessageTags converts message tags to a slice of strings
func (fp *FacebookParser) extractMessageTags(tags []struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
}) []string {
	var tagNames []string
	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}
	return tagNames
}

// getMediaTypeFromStatus maps Facebook status types to our media types
func (fp *FacebookParser) getMediaTypeFromStatus(statusType string) string {
	statusMapping := map[string]string{
		"added_photos":         "images",
		"added_video":          "videos",
		"shared_story":         "link",
		"published_story":      "link",
		"mobile_status_update": "text",
		"reels":                "reels",
	}

	if mapped, exists := statusMapping[statusType]; exists {
		return mapped
	}
	return "others"
}

// parsePostInsights extracts insights data from raw post
func (fp *FacebookParser) parsePostInsights(rawPost kafkamodels.RawFacebookPost, post *kafkamodels.ParsedFacebookPost) {
	if rawPost.Insights == nil || len(rawPost.Insights.Data) == 0 {
		return
	}

	for _, insight := range rawPost.Insights.Data {
		if len(insight.Values) == 0 {
			continue
		}

		value := fp.getInt64Value(insight.Values[0].Value)

		switch insight.Name {
		case "post_impressions":
			post.PostImpressions = value
			post.TotalImpressions += value
		case "post_impressions_unique":
			post.PostImpressionsUnique = value
			post.TotalImpressions += value
		case "post_impressions_paid":
			post.PostImpressionsPaid = value
			post.TotalImpressions += value
		case "post_impressions_paid_unique":
			post.PostImpressionsPaidUnique = value
			post.TotalImpressions += value
		case "post_impressions_organic":
			post.PostImpressionsOrganic = value
			post.TotalImpressions += value
		case "post_impressions_organic_unique":
			post.PostImpressionsOrganicUnique = value
			post.TotalImpressions += value
		case "post_impressions_viral":
			post.PostImpressionsViral = value
			post.TotalImpressions += value
		case "post_impressions_viral_unique":
			post.PostImpressionsViralUnique = value
			post.TotalImpressions += value
		case "post_clicks":
			post.PostClicks = value
			post.TotalEngagement += value
		case "post_video_views":
			post.PostVideoViews = value
		case "post_media_view":
			post.PostImpressions = value
			post.TotalImpressions += value
		}
	}

	if rawPost.PostMediaViewByFollowers != nil {
		post.PostImpressionsPaid = int64(rawPost.PostMediaViewByFollowers.Data[0].Values[0].Value)
	}

	if rawPost.PostMediaViewByAdd != nil {
		for _, add := range rawPost.PostMediaViewByAdd.Data[0].Values {
			if add.IsFromAds == "0" {
				post.PostImpressionsOrganic = int64(add.Value)
			}
			if add.IsFromAds == "1" {
				post.PostImpressionsPaid += int64(add.Value)
			}

		}
	}
}

// ParsePostReactions extracts reaction counts from raw post
func (fp *FacebookParser) ParsePostReactions(rawPost kafkamodels.RawFacebookPost, post *kafkamodels.ParsedFacebookPost) {
	// Parse individual reactions
	if rawPost.Total != nil && rawPost.Total.Summary != nil {
		post.Total = fp.getInt64Value(rawPost.Total.Summary.TotalCount)
		post.TotalEngagement += post.Total
	}

	if rawPost.Like != nil && rawPost.Like.Summary != nil {
		post.Like = int32(fp.getInt64Value(rawPost.Like.Summary.TotalCount))
	}

	if rawPost.Love != nil && rawPost.Love.Summary != nil {
		post.Love = int32(fp.getInt64Value(rawPost.Love.Summary.TotalCount))
	}

	if rawPost.Haha != nil && rawPost.Haha.Summary != nil {
		post.Haha = int32(fp.getInt64Value(rawPost.Haha.Summary.TotalCount))
	}

	if rawPost.Wow != nil && rawPost.Wow.Summary != nil {
		post.Wow = int32(fp.getInt64Value(rawPost.Wow.Summary.TotalCount))
	}

	if rawPost.Sad != nil && rawPost.Sad.Summary != nil {
		post.Sad = int32(fp.getInt64Value(rawPost.Sad.Summary.TotalCount))
	}

	if rawPost.Angry != nil && rawPost.Angry.Summary != nil {
		post.Angry = int32(fp.getInt64Value(rawPost.Angry.Summary.TotalCount))
	}
	if rawPost.Thankful != nil && rawPost.Thankful.Summary != nil {
		post.Thankful = int32(fp.getInt64Value(rawPost.Thankful.Summary.TotalCount))
	}

	// Parse comments
	if rawPost.Comments != nil && rawPost.Comments.Summary != nil {
		post.Comments = int32(fp.getInt64Value(rawPost.Comments.Summary.TotalCount))
		post.TotalEngagement += int64(post.Comments)
	}

}

// parseChildAttachments handles carousel post attachments
func (fp *FacebookParser) parseChildAttachments(rawPost kafkamodels.RawFacebookPost, pageID string) []kafkamodels.ParsedFacebookMediaAsset {
	var assets []kafkamodels.ParsedFacebookMediaAsset

	for i, attachment := range rawPost.ChildAttachments {
		//_, parsePostID := splitComposite(rawPost.ID)
		asset := kafkamodels.ParsedFacebookMediaAsset{
			PageID:      pageID,
			PostID:      rawPost.ID,
			MediaID:     GenerateMediaID(rawPost.ID, i),
			Caption:     rawPost.Message,
			Description: attachment.Description,
			CreatedAt:   rawPost.CreatedTime.Time,
			InsertedAt:  time.Now().UTC(),
		}

		// Determine asset type
		if attachment.Media != nil && attachment.Media.Source != "" {
			asset.AssetType = "video"
			asset.Link = attachment.Media.Source
		} else {
			asset.AssetType = "photo"
			if attachment.Media != nil && attachment.Media.Image != nil {
				asset.Link = attachment.Media.Image.Source
			}
		}

		assets = append(assets, asset)
	}

	return assets
}

// parseAttachments handles regular post attachments
func (fp *FacebookParser) parseAttachments(rawPost kafkamodels.RawFacebookPost, pageID string) []kafkamodels.ParsedFacebookMediaAsset {
	var assets []kafkamodels.ParsedFacebookMediaAsset

	if rawPost.Attachments == nil || len(rawPost.Attachments.Data) == 0 {
		return assets
	}

	for i, attachment := range rawPost.Attachments.Data {
		//_, parsePostID := splitComposite(rawPost.ID)
		asset := kafkamodels.ParsedFacebookMediaAsset{
			PageID:      pageID,
			PostID:      rawPost.ID,
			Caption:     rawPost.Message,
			Description: attachment.Description,
			CreatedAt:   rawPost.CreatedTime.Time,
			InsertedAt:  time.Now().UTC(),
		}

		// Set media ID
		if attachment.Target != nil && attachment.Target.ID != "" && attachment.Target.ID != pageID {
			asset.MediaID = attachment.Target.ID
		} else {
			asset.MediaID = GenerateMediaID(rawPost.ID, i)
		}

		// Set asset type
		if attachment.MediaType != "" {
			asset.AssetType = attachment.MediaType
		} else if attachment.Type == "multi_share_no_end_card" {
			asset.AssetType = "photo"
		} else {
			asset.AssetType = attachment.Type
		}

		// Set link/URL
		if attachment.Link != "" {
			asset.CallToAction = attachment.Link
		}

		// Set media link
		if attachment.Media != nil {

			if asset.AssetType == "video" {
				asset.Link = attachment.Media.Source
				if asset.Link == "" {
					asset.Link = attachment.Media.Src
				}
			} else if attachment.Media.Image != nil {
				asset.Link = attachment.Media.Image.Source
				if asset.Link == "" {
					asset.Link = attachment.Media.Image.Src
				}
			}
		}

		assets = append(assets, asset)

		// Handle subattachments

		if attachment.Subattachments != nil {
			for j, subattachment := range attachment.Subattachments.Data {
				subAsset := kafkamodels.ParsedFacebookMediaAsset{
					PageID:     pageID,
					PostID:     rawPost.ID,
					MediaID:    GenerateMediaID(rawPost.ID, i*1000+j), // Ensure unique ID
					Caption:    rawPost.Message,
					CreatedAt:  rawPost.CreatedTime.Time,
					InsertedAt: time.Now().UTC(),
					AssetType:  subattachment.Type,
				}

				if subattachment.Media != nil {
					if subattachment.Media.Source != "" {
						subAsset.AssetType = "video"
						subAsset.Link = subattachment.Media.Source
					} else if subattachment.Media.Image != nil {
						subAsset.AssetType = "photo"
						subAsset.Link = subattachment.Media.Image.Source
					}

				}

				assets = append(assets, subAsset)
			}
		}
	}

	return assets
}

// GenerateMediaID creates a unique media ID using MD5 hash
func GenerateMediaID(postID string, index int) string {
	input := fmt.Sprintf("%s_%d", postID, index)
	hash := md5.Sum([]byte(input))
	return fmt.Sprintf("%x", hash)
}

// extractVideoID extracts video ID from Facebook video link
func (fp *FacebookParser) extractVideoID(link string) string {
	// Extract video ID from Facebook video URL
	re := regexp.MustCompile(`/videos/(\d+)`)
	matches := re.FindStringSubmatch(link)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// ParseRawFacebookPost is a convenience function to parse a raw Facebook post with default parameters
func ParseRawFacebookPost(rawPost kafkamodels.RawFacebookPost) (*kafkamodels.ParsedFacebookPost, []kafkamodels.ParsedFacebookMediaAsset, error) {
	parser := NewFacebookParser()

	// Extract page info from the post
	pageID := ""
	pageName := ""
	if rawPost.From != nil {
		pageID = rawPost.From.ID
		pageName = rawPost.From.Name
	}

	return parser.ParsePost(rawPost, pageID, pageName, "")
}

// ParseVideo parses a RawFacebookVideo into ParsedFacebookVideoInsights
func (fp *FacebookParser) ParseVideo(rawVideo kafkamodels.RawFacebookVideo, pageID, pageName string) (kafkamodels.ParsedFacebookVideoInsights, error) {
	// Get current time for saving_time
	savingTime := time.Now().UTC()
	// Initialize parsed video insights
	parsed := kafkamodels.ParsedFacebookVideoInsights{
		VideoID: rawVideo.ID,
		PostID:  fmt.Sprintf("%s_%s", pageID, rawVideo.PostID),
		PageID:  pageID,

		// Timestamps
		SavingTime: savingTime,
	}

	// Parse created time and set day/hour info
	if !rawVideo.CreatedTime.Time.IsZero() {
		parsed.CreatedTime = rawVideo.CreatedTime.Time
	}

	// Parse updated time
	if !rawVideo.UpdatedTime.Time.IsZero() {
		parsed.UpdatedTime = rawVideo.CreatedTime.Time
	}
	// Parse video insights if available
	if len(rawVideo.VideoInsights.Data) > 0 {
		fp.parseVideoInsights(&parsed, rawVideo.VideoInsights.Data)
	}

	return parsed, nil
}

// ParseVideoPostInsights is used to parse post insights related to videos and reels
func (fp *FacebookParser) ParseVideoPostInsights(videoAsPost kafkamodels.RawFacebookPost) *kafkamodels.ParsedFacebookPost {
	parsePost := &kafkamodels.ParsedFacebookPost{}
	fp.parsePostInsights(videoAsPost, parsePost)

	// Parse reactions and engagement
	fp.ParsePostReactions(videoAsPost, parsePost)

	if !videoAsPost.CreatedTime.Time.IsZero() {
		parsePost.CreatedTime = videoAsPost.CreatedTime.Time
		parsePost.DayOfWeek = videoAsPost.CreatedTime.Time.Weekday().String()
		parsePost.HourOfDay = int32(videoAsPost.CreatedTime.Time.Hour())
	}

	// Parse updated time
	if !videoAsPost.UpdatedTime.Time.IsZero() {
		parsePost.UpdatedTime = videoAsPost.UpdatedTime.Time
	}

	return parsePost
}

// parseVideoInsights processes video insights data and populates the parsed video struct
func (fp *FacebookParser) parseVideoInsights(parsed *kafkamodels.ParsedFacebookVideoInsights, insights []struct {
	Name   string `json:"name"`
	Period string `json:"period"`
	Values []struct {
		Value   interface{} `json:"value"`
		EndTime string      `json:"end_time"`
	} `json:"values"`
	Title       string `json:"title"`
	Description string `json:"description"`
}) {
	for _, insight := range insights {
		if len(insight.Values) == 0 {
			continue
		}

		// Get the latest value (usually index 0)
		value := insight.Values[0].Value
		switch insight.Name {
		// View metrics
		case "total_video_views":
			parsed.TotalVideoViews = fp.getInt64Value(value)
		case "total_video_views_unique":
			parsed.TotalVideoViewsUnique = fp.getInt64Value(value)
		case "total_video_views_autoplayed":
			parsed.TotalVideoViewsAutoplayed = fp.getInt64Value(value)
		case "total_video_views_clicked_to_play":
			parsed.TotalVideoViewsClickedToPlay = fp.getInt64Value(value)
		case "total_video_views_organic":
			parsed.TotalVideoViewsOrganic = fp.getInt64Value(value)
		case "total_video_views_organic_unique":
			parsed.TotalVideoViewsOrganicUnique = fp.getInt64Value(value)
		case "total_video_views_paid":
			parsed.TotalVideoViewsPaid = fp.getInt64Value(value)
		case "total_video_views_paid_unique":
			parsed.TotalVideoViewsPaidUnique = fp.getInt64Value(value)
		case "total_video_views_sound_on":
			parsed.TotalVideoViewsSoundOn = fp.getInt64Value(value)

		// Complete view metrics
		case "total_video_complete_views":
			parsed.TotalVideoCompleteViews = fp.getInt64Value(value)
		case "total_video_complete_views_unique":
			parsed.TotalVideoCompleteViewsUnique = fp.getInt64Value(value)
		case "total_video_complete_views_auto_played":
			parsed.TotalVideoCompleteViewsAutoplayed = fp.getInt64Value(value)
		case "total_video_complete_views_clicked_to_play":
			parsed.TotalVideoCompleteViewsClickedToPlay = fp.getInt64Value(value)
		case "total_video_complete_views_organic":
			parsed.TotalVideoCompleteViewsOrganic = fp.getInt64Value(value)
		case "total_video_complete_views_organic_unique":
			parsed.TotalVideoCompleteViewsOrganicUnique = fp.getInt64Value(value)
		case "total_video_complete_views_paid":
			parsed.TotalVideoCompleteViewsPaid = fp.getInt64Value(value)
		case "total_video_complete_views_paid_unique":
			parsed.TotalVideoCompleteViewsPaidUnique = fp.getInt64Value(value)

		// Time-based view metrics
		case "total_video_10s_views":
			parsed.TotalVideo10sViews = fp.getInt64Value(value)
		case "total_video_10s_views_unique":
			parsed.TotalVideo10sViewsUnique = fp.getInt64Value(value)
		case "total_video_10s_views_auto_played":
			parsed.TotalVideo10sViewsAutoplayed = fp.getInt64Value(value)
		case "total_video_10s_views_clicked_to_play":
			parsed.TotalVideo10sViewsClickedToPlay = fp.getInt64Value(value)
		case "total_video_10s_views_organic":
			parsed.TotalVideo10sViewsOrganic = fp.getInt64Value(value)
		case "total_video_10s_views_paid":
			parsed.TotalVideo10sViewsPaid = fp.getInt64Value(value)
		case "total_video_10s_views_sound_on":
			parsed.TotalVideo10sViewsSoundOn = fp.getInt64Value(value)
		case "total_video_15s_views":
			parsed.TotalVideo15sViews = fp.getInt64Value(value)
		case "total_video_60s_excludes_shorter_views":
			parsed.TotalVideo60sExcludesShorterViews = fp.getInt64Value(value)

		// Time metrics
		case "total_video_avg_time_watched":
			parsed.TotalVideoAvgTimeWatched = fp.getInt64Value(value)
		case "post_video_avg_time_watched":
			parsed.PostVideoAvgTimeWatched = fp.getInt64Value(value)
		case "total_video_view_total_time":
			parsed.TotalVideoViewTotalTime = fp.getInt64Value(value)
		case "total_video_view_total_time_organic":
			parsed.TotalVideoViewTotalTimeOrganic = fp.getInt64Value(value)
		case "total_video_view_total_time_paid":
			parsed.TotalVideoViewTotalTimePaid = fp.getInt64Value(value)
		case "post_video_view_time":
			parsed.PostVideoViewTime = fp.getInt64Value(value)

		// Impression metrics
		case "total_video_impressions":
			parsed.TotalVideoImpressions = fp.getInt64Value(value)
		case "total_video_impressions_unique":
			parsed.TotalVideoImpressionsUnique = fp.getInt64Value(value)
		case "total_video_impressions_paid":
			parsed.TotalVideoImpressionsPaid = fp.getInt64Value(value)
		case "total_video_impressions_paid_unique":
			parsed.TotalVideoImpressionsPaidUnique = fp.getInt64Value(value)
		case "total_video_impressions_organic":
			parsed.TotalVideoImpressionsOrganic = fp.getInt64Value(value)
		case "total_video_impressions_organic_unique":
			parsed.TotalVideoImpressionsOrganicUnique = fp.getInt64Value(value)
		case "total_video_impressions_viral":
			parsed.TotalVideoImpressionsViral = fp.getInt64Value(value)
		case "total_video_impressions_viral_unique":
			parsed.TotalVideoImpressionsViralUnique = fp.getInt64Value(value)
		case "total_video_impressions_fan":
			parsed.TotalVideoImpressionsFan = fp.getInt64Value(value)
		case "total_video_impressions_fan_unique":
			parsed.TotalVideoImpressionsFanUnique = fp.getInt64Value(value)
		case "total_video_impressions_fan_paid":
			parsed.TotalVideoImpressionsFanPaid = fp.getInt64Value(value)
		case "total_video_impressions_fan_paid_unique":
			parsed.TotalVideoImpressionsFanPaidUnique = fp.getInt64Value(value)
		case "post_impressions_unique":
			parsed.PostImpressionsUnique = fp.getInt64Value(value)

		// Reels-specific metrics
		case "blue_reels_play_count":
			parsed.BlueReelsPlayCount = fp.getInt64Value(value)

		// Retention and other string metrics
		case "total_video_retention_graph":
			parsed.TotalVideoRetentionGraph = fp.getStringValue(value)
		case "total_video_retention_graph_autoplayed":
			parsed.TotalVideoRetentionGraphAutoplayed = fp.getStringArrayValue(value)
		case "total_video_retention_graph_clicked_to_play":
			parsed.TotalVideoRetentionGraphClickedToPlay = fp.getStringArrayValue(value)
		case "total_video_stories_by_action_type":
			parsed.TotalVideoStoriesByActionType = fp.getStringArrayValue(value)
		case "total_video_reactions_by_type_total":
			parsed.TotalVideoReactionsByTypeTotal = fp.getStringArrayValue(value)
		case "total_video_view_time_by_age_bucket_and_gender":
			parsed.TotalVideoViewTimeByAgeBucketAndGender = fp.getStringArrayValue(value)
		case "total_video_view_time_by_region_id":
			parsed.TotalVideoViewTimeByRegionID = fp.getStringArrayValue(value)
		case "total_video_views_by_distribution_type":
			parsed.TotalVideoViewsByDistributionType = fp.getStringArrayValue(value)
		case "total_video_view_time_by_distribution_type":
			parsed.TotalVideoViewTimeByDistributionType = fp.getStringArrayValue(value)
		case "total_video_view_total_time_live":
			parsed.TotalVideoViewTotalTimeLive = fp.getInt64Value(value)
		case "total_video_views_live":
			parsed.TotalVideoViewsLive = fp.getInt64Value(value)
		}
	}
}

// getInt64Value safely extracts an int64 value from an interface{}
func (fp *FacebookParser) getInt64Value(value interface{}) int64 {
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case int:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(v)
	case string:
		// Try to parse string as int64
		if v == "" {
			return 0
		}
		// Handle string numbers if needed
		return 0
	default:
		return 0
	}
}

// getFloat64Value safely extracts a float64 value from an interface{}
func (fp *FacebookParser) getFloat64Value(value interface{}) float64 {
	if value == nil {
		return 0.0
	}

	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		// Try to parse string as float64
		if v == "" {
			return 0.0
		}
		// Handle string numbers if needed
		return 0.0
	default:
		return 0.0
	}
}

// getStringValue safely extracts a string value from an interface{}
func (fp *FacebookParser) getStringValue(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	default:
		return ""
	}
}

// getStringArrayValue safely extracts a string array value from an interface{}
func (fp *FacebookParser) getStringArrayValue(value interface{}) []string {
	if value == nil {
		return []string{}
	}

	switch v := value.(type) {
	case []string:
		return v
	case string:
		return []string{v}
	default:
		return []string{}
	}
}

// ParseVideoFromJSON is a convenience function that parses a RawFacebookVideo from JSON
// into ParsedFacebookVideoInsights
func ParseVideoFromJSON(rawVideo kafkamodels.RawFacebookVideo, pageID, pageName string) (kafkamodels.ParsedFacebookVideoInsights, error) {
	parser := NewFacebookParser()
	return parser.ParseVideo(rawVideo, pageID, pageName)
}

// ParseInsights converts RawFacebookInsights to ParsedFacebookInsights (single record - backward compatibility)
func (p *FacebookParser) ParseInsights(rawInsights kafkamodels.RawFacebookInsights, pageID, workspaceID string) (*kafkamodels.ParsedFacebookInsights, error) {
	results, err := p.ParseInsightsDaily(rawInsights, pageID, workspaceID)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	// Return first one for backward compatibility
	return results[0], nil
}

// ParseInsightsDaily converts RawFacebookInsights to multiple ParsedFacebookInsights records (one per day).
// Each day in the API response values array becomes a separate record with its own created_time.
func (p *FacebookParser) ParseInsightsDaily(rawInsights kafkamodels.RawFacebookInsights, pageID, workspaceID string) ([]*kafkamodels.ParsedFacebookInsights, error) {
	// First, collect all unique dates from the insight values
	dateSet := make(map[string]time.Time)
	for _, insight := range rawInsights.Data {
		for _, val := range insight.Values {
			if val.EndTime == "" {
				continue
			}
			endTime, err := time.Parse(time.RFC3339, val.EndTime)
			if err != nil {
				// Try alternative format
				endTime, err = time.Parse("2006-01-02T15:04:05-0700", val.EndTime)
				if err != nil {
					continue
				}
			}
			dateStr := endTime.Format("2006-01-02")
			dateSet[dateStr] = endTime
		}
	}

	// If no dates found, return nil
	if len(dateSet) == 0 {
		return nil, nil
	}

	// Create a parsed insight record for each date
	now := time.Now().UTC()
	var results []*kafkamodels.ParsedFacebookInsights
	for dateStr, endTime := range dateSet {
		// Generate hash ID for this date's insights record
		nameID := fmt.Sprintf("%s_%s", pageID, dateStr)
		hash := md5.Sum([]byte(nameID))
		hashID := fmt.Sprintf("%x", hash)

		parsed := &kafkamodels.ParsedFacebookInsights{
			HashID:      hashID,
			PageID:      pageID,
			WorkspaceID: workspaceID,
			Year:        endTime.Year(),
			Month:       int(endTime.Month()),
			DayOfWeek:   endTime.Weekday().String(),
			CreatedTime: endTime, // The actual date the data belongs to (from API end_time)
			SavingTime:  now,     // When the record was saved (today's date)

			// Initialize counters
			PositiveSentiment:    0,
			NegativeSentiment:    0,
			PageFansByLike:       0,
			PageFansByUnlike:     0,
			PagePositiveFeedback: 0,
			ActiveUsers:          0,
		}

		// Process each insight metric for this specific date
		for _, insight := range rawInsights.Data {
			p.processInsightMetricForDate(parsed, insight, endTime)
		}

		// Calculate derived metrics
		p.calculateDerivedMetrics(parsed)

		results = append(results, parsed)
	}

	p.log.Info().
		Str("page_id", pageID).
		Int("daily_records", len(results)).
		Msg("Parsed Facebook insights into daily records")

	return results, nil
}

// processInsightMetricForDate processes individual insight metrics for a specific date
func (p *FacebookParser) processInsightMetricForDate(parsed *kafkamodels.ParsedFacebookInsights, insight kafkamodels.FacebookInsightData, targetDate time.Time) {
	name := insight.Name
	value := p.getValueForDate(insight.Values, targetDate)

	if value == nil {
		return
	}

	p.applyInsightValue(parsed, name, value)
}

// applyInsightValue applies an insight value to a parsed insights record
func (p *FacebookParser) applyInsightValue(parsed *kafkamodels.ParsedFacebookInsights, name string, value interface{}) {
	switch name {
	// Basic page metrics
	case "page_follows":
		parsed.PageFollows = p.getInt64Value(value)
	case "page_views_total":
		parsed.PageViews = p.getInt64Value(value)
	case "page_fans":
		parsed.PageFans = p.getInt64Value(value)
	case "page_total_actions":
		parsed.PageTotalActions = p.getInt64Value(value)
	case "page_post_engagements":
		parsed.PagePostEngagements = p.getInt64Value(value)

	// Impression metrics
	case "page_impressions_unique":
		parsed.PageImpressionsUnique = p.getInt64Value(value)
	case "page_media_view":
		parsed.PageImpressions = p.getInt64Value(value) // alternate for page impressions
		parsed.PageMediaView = p.getInt64Value(value)
	case "page_impressions_organic_v2", "page_impressions_organic":
		parsed.PageImpressionsOrganic = p.getInt64Value(value)
	case "page_impressions_paid":
		parsed.PageImpressionsPaid = p.getInt64Value(value)

	// Fan metrics
	case "page_fan_adds_unique":
		parsed.PageFanAddsUnique = p.getInt64Value(value)
	case "page_fan_removes_unique":
		parsed.PageFanRemovesUnique = p.getInt64Value(value)
	case "page_fan_adds_by_paid_non_paid_unique":
		parsed.PageFanAddsByPaidNonPaidUnique = p.convertMapToStringSlice(value)
		parsed.PageFansByLike += p.sumMapValues(value)

	// Video metrics
	case "page_video_views":
		parsed.PageVideoViews = p.getInt64Value(value)
	case "page_video_views_paid":
		parsed.PageVideoViewsPaid = p.getInt64Value(value)
	case "page_video_views_organic":
		parsed.PageVideoViewsOrganic = p.getInt64Value(value)
	case "page_video_views_autoplayed":
		parsed.PageVideoViewsAutoplayed = p.getInt64Value(value)
	case "page_video_views_click_to_play":
		parsed.PageVideoViewsClickToPlay = p.getInt64Value(value)
	case "page_video_repeat_views":
		parsed.PageVideoRepeatViews = p.getInt64Value(value)

	// Reaction metrics (positive sentiment)
	case "page_actions_post_reactions_like_total":
		reactionValue := p.getInt64Value(value)
		parsed.PageActionsPostReactionsLikeTotal = reactionValue
		parsed.PositiveSentiment += reactionValue
	case "page_actions_post_reactions_love_total":
		reactionValue := p.getInt64Value(value)
		parsed.PageActionsPostReactionsLoveTotal = reactionValue
		parsed.PositiveSentiment += reactionValue
	case "page_actions_post_reactions_anger_total":
		reactionValue := p.getInt64Value(value)
		parsed.PageActionsPostReactionsAngerTotal = reactionValue
		parsed.NegativeSentiment = reactionValue

	// Feedback metrics
	case "page_negative_feedback":
		parsed.PageNegativeFeedback = p.getInt64Value(value)
	case "page_negative_feedback_by_type":
		parsed.PageNegativeFeedbackByType = p.convertMapToStringSlice(value)
	case "page_positive_feedback_by_type":
		parsed.PagePositiveFeedbackByType = p.convertMapToStringSlice(value)
		parsed.PagePositiveFeedback += p.sumMapValues(value)

	// Fan activity and demographics
	case "page_fans_online":
		onlineFans := p.getOnlineFans(value)
		parsed.PageFansOnline = p.convertMapToStringSlice(onlineFans)
		parsed.PrimeTime = p.getPrimeTime(onlineFans, parsed.SavingTime)
		parsed.ActiveUsers = p.calculateAverageOnlineFans(onlineFans)
	case "page_fans_locale":
		parsed.PageFansLocale = p.convertMapToStringSlice(value)
	case "page_fans_country":
		parsed.PageFansCountry = p.convertMapToStringSlice(value)
	case "page_fans_city":
		parsed.PageFansCity = p.convertMapToStringSlice(value)
	case "page_fans_gender_age":
		parsed.PageFansGenderAge = p.convertMapToStringSlice(value)
		parsed.PageFansGender = p.extractGenderFromGenderAge(value)
		parsed.PageFansAge = p.extractAgeFromGenderAge(value)

	// Page info
	case "talking_about_count":
		parsed.TalkingAboutCount = p.getInt64Value(value)
	case "page_category":
		parsed.PageCategory = p.getStringValue(value)

	// Source tracking
	case "page_fans_by_like_source_unique":
		parsed.PageFansByLikeSourceUnique = p.convertMapToStringSlice(value)
		parsed.PageFansByLike += p.sumMapValues(value)
	case "page_fans_by_unlike_source_unique":
		parsed.PageFansByUnlikeSourceUnique = p.convertMapToStringSlice(value)
		parsed.PageFansByUnlike += p.sumMapValues(value)
	}
}

// processInsightMetric processes individual insight metrics (uses applyInsightValue)
func (p *FacebookParser) processInsightMetric(parsed *kafkamodels.ParsedFacebookInsights, insight kafkamodels.FacebookInsightData, targetDate time.Time) {
	p.processInsightMetricForDate(parsed, insight, targetDate)
}

// getValueForDate finds the insight value for a specific date
func (p *FacebookParser) getValueForDate(values []kafkamodels.FacebookInsightValue, targetDate time.Time) interface{} {
	targetDateStr := targetDate.Format("2006-01-02")

	for _, val := range values {
		// Try RFC3339 format first
		endTime, err := time.Parse(time.RFC3339, val.EndTime)
		if err != nil {
			// Try Facebook's format: 2025-12-06T08:00:00+0000
			endTime, err = time.Parse("2006-01-02T15:04:05-0700", val.EndTime)
			if err != nil {
				continue
			}
		}
		if endTime.Format("2006-01-02") == targetDateStr {
			return val.Value
		}
	}

	// Return nil if no exact date match (don't fallback to last value)
	return nil
}

// convertMapToStringSlice converts map values to string slice format
func (p *FacebookParser) convertMapToStringSlice(value interface{}) []string {
	if value == nil {
		return []string{}
	}

	switch v := value.(type) {
	case map[string]interface{}:
		var result []string
		for key, val := range v {
			result = append(result, fmt.Sprintf("%s:%v", key, val))
		}
		return result
	default:
		return []string{fmt.Sprintf("%v", value)}
	}
}

// sumMapValues sums all values in a map
func (p *FacebookParser) sumMapValues(value interface{}) int64 {
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case map[string]interface{}:
		var sum int64
		for _, val := range v {
			sum += p.getInt64Value(val)
		}
		return sum
	default:
		return p.getInt64Value(value)
	}
}

// getOnlineFans processes online fans data to ensure all 24 hours are represented
func (p *FacebookParser) getOnlineFans(value interface{}) map[string]interface{} {
	online := make(map[string]interface{})

	// Initialize all 24 hours
	for hr := 0; hr < 24; hr++ {
		online[fmt.Sprintf("%d", hr)] = 0
	}

	// Fill in actual data
	if fansMap, ok := value.(map[string]interface{}); ok {
		for hour, count := range fansMap {
			online[hour] = count
		}
	}

	return online
}

// getPrimeTime calculates the hour with most online fans
func (p *FacebookParser) getPrimeTime(activity map[string]interface{}, baseDate time.Time) time.Time {
	if len(activity) == 0 {
		return baseDate.Truncate(24 * time.Hour) // Start of day if no data
	}

	maxHour := "0"
	maxCount := int64(0)

	for hour, count := range activity {
		if hourCount := p.getInt64Value(count); hourCount > maxCount {
			maxCount = hourCount
			maxHour = hour
		}
	}

	primeHour := 0
	if hour, err := fmt.Sscanf(maxHour, "%d", &primeHour); err == nil && hour == 1 {
		return baseDate.Truncate(24 * time.Hour).Add(time.Duration(primeHour) * time.Hour)
	}

	return baseDate.Truncate(24 * time.Hour)
}

// calculateAverageOnlineFans calculates average fans online per hour
func (p *FacebookParser) calculateAverageOnlineFans(activity map[string]interface{}) int64 {
	if len(activity) == 0 {
		return 0
	}

	var total int64
	count := 0

	for _, val := range activity {
		total += p.getInt64Value(val)
		count++
	}

	if count > 0 {
		return total / int64(count)
	}

	return 0
}

// extractGenderFromGenderAge extracts gender distribution from gender_age data
func (p *FacebookParser) extractGenderFromGenderAge(value interface{}) []string {
	genderCounts := map[string]int64{
		"U": 0, "M": 0, "F": 0,
	}

	if genderAgeMap, ok := value.(map[string]interface{}); ok {
		for genderAge, count := range genderAgeMap {
			if len(genderAge) > 0 {
				gender := string(genderAge[0])
				if _, exists := genderCounts[gender]; exists {
					genderCounts[gender] += p.getInt64Value(count)
				}
			}
		}
	}

	return p.convertMapToStringSlice(genderCounts)
}

// extractAgeFromGenderAge extracts age distribution from gender_age data
func (p *FacebookParser) extractAgeFromGenderAge(value interface{}) []string {
	ageCounts := map[string]int64{
		"13-17": 0, "18-24": 0, "25-34": 0, "35-44": 0,
		"45-54": 0, "55-64": 0, "65+": 0,
	}

	if genderAgeMap, ok := value.(map[string]interface{}); ok {
		for genderAge, count := range genderAgeMap {
			// Extract age part (everything after first underscore or dot)
			parts := strings.Split(genderAge, "_")
			if len(parts) < 2 {
				parts = strings.Split(genderAge, ".")
			}

			if len(parts) >= 2 {
				age := parts[1]
				if _, exists := ageCounts[age]; exists {
					ageCounts[age] += p.getInt64Value(count)
				}
			}
		}
	}

	return p.convertMapToStringSlice(ageCounts)
}

// calculateDerivedMetrics calculates any remaining derived metrics
func (p *FacebookParser) calculateDerivedMetrics(parsed *kafkamodels.ParsedFacebookInsights) {
	// Calculate average active users (fans online per hour / 24)
	if parsed.ActiveUsers > 0 {
		parsed.ActiveUsers = parsed.ActiveUsers / 24
	}
}

// ParseInsightsFromJSON is a convenience function for parsing insights
func ParseInsightsFromJSON(rawInsights kafkamodels.RawFacebookInsights, pageID, workspaceID string) (*kafkamodels.ParsedFacebookInsights, error) {
	parser := NewFacebookParser()
	return parser.ParseInsights(rawInsights, pageID, workspaceID)
}

func splitComposite(id string) (pageID, postID string) {
	parts := strings.SplitN(id, "_", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	// If unexpected format, return all in postID
	return "", id
}
