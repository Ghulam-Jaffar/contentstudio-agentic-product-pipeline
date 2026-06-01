package parsing

import (
	"fmt"
	"strings"
	"time"

	apiModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	models "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

// InstagramCompetitorParser handles parsing of Instagram API responses to ClickHouse models
type InstagramCompetitorParser struct {
	pageID      string
	pageName    string
	displayName string
}

// NewInstagramCompetitorParser creates a new parser
func NewInstagramCompetitorParser(pageID, pageName, displayName string) *InstagramCompetitorParser {
	return &InstagramCompetitorParser{
		pageID:      pageID,
		pageName:    pageName,
		displayName: displayName,
	}
}

// ParsePageInsights parses business discovery data into competitor insights
func (p *InstagramCompetitorParser) ParsePageInsights(businessDiscovery *apiModels.BusinessDiscovery) *models.InstagramCompetitorInsights {
	now := time.Now().UTC()
	recordID := generateRecordID(fmt.Sprintf("%d", businessDiscovery.IgID), now)

	insight := &models.InstagramCompetitorInsights{
		RecordID:             recordID,
		InstagramAccountID:   businessDiscovery.ID,
		TotalFollowedByCount: businessDiscovery.FollowersCount,
		TotalFollowingCount:  businessDiscovery.FollowsCount,
		ProfilePictureURL:    businessDiscovery.ProfilePictureURL,
		PageName:             p.displayName,
		InsertedAt:           now.Truncate(time.Hour).Add(0),
	}

	return insight
}

// ParsePosts parses Instagram media items into competitor posts
func (p *InstagramCompetitorParser) ParsePosts(media []apiModels.InstagramMedia, businessDiscovery *apiModels.BusinessDiscovery, profileImage string) []*models.InstagramCompetitorPosts {
	var posts []*models.InstagramCompetitorPosts

	for _, item := range media {
		post := p.parsePost(item, businessDiscovery, profileImage)
		if post == nil {
			continue
		}
		posts = append(posts, post)
	}

	return posts
}

// parsePost parses a single Instagram media item. Returns nil if the timestamp cannot be
// parsed, since a zero CreatedAt would land in the wrong partition and create a permanent
// cross-partition duplicate that FINAL cannot merge.
func (p *InstagramCompetitorParser) parsePost(media apiModels.InstagramMedia, businessDiscovery *apiModels.BusinessDiscovery, profileImage string) *models.InstagramCompetitorPosts {
	createdTime, err := time.Parse("2006-01-02T15:04:05-0700", media.Timestamp)
	if err != nil {
		createdTime, err = time.Parse(time.RFC3339, media.Timestamp)
	}
	if err != nil || createdTime.IsZero() {
		return nil
	}
	createdTime = createdTime.UTC()
	now := time.Now().UTC()

	post := &models.InstagramCompetitorPosts{
		InstagramID:          businessDiscovery.IgID,
		PostID:               media.ID,
		BusinessAccountID:    businessDiscovery.ID,
		TotalFollowedByCount: businessDiscovery.FollowersCount,
		TotalFollowingCount:  businessDiscovery.FollowsCount,
		Username:             businessDiscovery.Username,
		Name:                 businessDiscovery.Name,
		Biography:            businessDiscovery.Biography,
		ProfilePictureURL:    profileImage,
		MediaCount:           businessDiscovery.MediaCount,
		LikeCount:            media.LikeCount,
		CommentsCount:        media.CommentsCount,
		Caption:              media.Caption,
		MediaType:            media.MediaType,
		MediaProductType:     media.MediaProductType,
		MediaURL:             media.MediaURL,
		Permalink:            media.Permalink,
		CreatedAt:            createdTime,
		InsertedAt:           now,
	}

	// Calculate engagement
	post.Engagement = post.LikeCount + post.CommentsCount

	// Handle carousel albums - concatenate child media URLs
	if media.MediaType == "CAROUSEL_ALBUM" && media.Children != nil && len(media.Children.Data) > 0 {
		var mediaURLs []string
		for _, child := range media.Children.Data {
			if child.MediaURL != "" {
				mediaURLs = append(mediaURLs, child.MediaURL)
			}
		}
		if len(mediaURLs) > 0 {
			post.MediaURL = strings.Join(mediaURLs, ",")
		}
	}

	// Extract hashtags
	post.Hashtags = extractHashtags(media.Caption)

	return post
}
