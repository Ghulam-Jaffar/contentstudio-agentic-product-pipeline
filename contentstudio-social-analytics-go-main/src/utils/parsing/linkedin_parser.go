package parsing

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/common/constants"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ParsePost converts raw LinkedIn post JSON into ParsedLinkedinPost.
func ParsePost(raw json.RawMessage) (*kafkamodels.ParsedLinkedinPost, error) {
	var m map[string]interface{}

	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}

	p := &kafkamodels.ParsedLinkedinPost{}

	// Parse ID and activity
	if id, ok := m["id"].(string); ok {
		p.PostID = id[strings.LastIndex(id, ":")+1:]
		p.Activity = id
	}

	// Parse timestamps
	if created, ok := m["createdAt"].(float64); ok {
		p.CreatedAt = time.Unix(int64(created/1000), 0).UTC()
	}

	// Parse publishedAt - use this for DayOfWeek/HourOfDay since it's the actual publish time
	if published, ok := m["publishedAt"].(float64); ok {
		p.PublishedAt = time.Unix(int64(published/1000), 0).UTC()
		p.DayOfWeek = p.PublishedAt.Weekday().String()
		p.HourOfDay = int64(p.PublishedAt.Hour())
	} else {
		// Fallback to createdAt if publishedAt not available
		p.DayOfWeek = p.CreatedAt.Weekday().String()
		p.HourOfDay = int64(p.CreatedAt.Hour())
	}

	// Parse lastModifiedAt
	if lastModified, ok := m["lastModifiedAt"].(float64); ok {
		p.LastModifiedAt = time.Unix(int64(lastModified/1000), 0).UTC()
	}

	// Parse lifecycleState and visibility
	if lifecycleState, ok := m["lifecycleState"].(string); ok {
		p.LifecycleState = lifecycleState
	}
	if visibility, ok := m["visibility"].(string); ok {
		p.Visibility = visibility
	}

	// Parse isReshareDisabledByAuthor
	if isReshareDisabled, ok := m["isReshareDisabledByAuthor"].(bool); ok {
		p.IsReshareDisabled = isReshareDisabled
	}

	// Parse distribution fields
	if distribution, ok := m["distribution"].(map[string]interface{}); ok {
		if feedDist, ok := distribution["feedDistribution"].(string); ok {
			p.FeedDistribution = feedDist
		}
		if channels, ok := distribution["thirdPartyDistributionChannels"].([]interface{}); ok {
			for _, ch := range channels {
				if chStr, ok := ch.(string); ok {
					p.ThirdPartyChannels = append(p.ThirdPartyChannels, chStr)
				}
			}
		}
	}

	p.SavingTime = time.Now().UTC()

	// Parse commentary (title) and extract hashtags
	if commentary, ok := m["commentary"].(string); ok {
		// Replace LinkedIn's hashtag format {hashtag|#|tagname} with #tagname
		hashtagRegex := regexp.MustCompile(`\{hashtag\|\\#\|(.*?)\}`)
		p.Title = hashtagRegex.ReplaceAllString(commentary, "#$1")

		// Extract hashtags
		tagRegex := regexp.MustCompile(`#(\w+)`)
		matches := tagRegex.FindAllStringSubmatch(p.Title, -1)
		for _, match := range matches {
			if len(match) > 1 {
				p.Hashtags = append(p.Hashtags, match[1])
			}
		}
	}

	if meta, ok := m["meta"].(map[string]interface{}); ok {
		if status, ok := meta["stats"].(map[string]interface{}); ok {
			if v, ok := status["clickCount"].(float64); ok {
				p.PostClicks = int64(v)
			}
			if v, ok := status["commentCount"].(float64); ok {
				p.Comments = int64(v)
			}
			if v, ok := status["engagement"].(float64); ok {
				p.TotalEngagement = v
			}
			if v, ok := status["impressionCount"].(float64); ok {
				p.Impressions = int64(v)
			}
			if v, ok := status["uniqueImpressionsCount"].(float64); ok {
				p.Reach = int64(v)
			}
			if v, ok := status["shareCount"].(float64); ok {
				p.Repost = int64(v)
			}
			if v, ok := status["likeCount"].(float64); ok {
				p.Favorites = int64(v)
			}
		}

		if assets, ok := meta["assets"].(map[string]interface{}); ok {
			if images, ok := assets["images"].([]interface{}); ok {
				for _, image := range images {
					if imgMap, ok := image.(map[string]interface{}); ok {
						if url, ok := imgMap["downloadUrl"].(string); ok && url != "" {
							p.Image = url
							p.Media = append(p.Media, url)
						}
					}
				}
			}

			if videos, ok := assets["videos"].([]interface{}); ok {
				for _, video := range videos {
					if vidMap, ok := video.(map[string]interface{}); ok {
						if thumb, ok := vidMap["thumbnail"].(string); ok && thumb != "" {
							p.Image = thumb
							p.Media = append(p.Media, thumb)
						}
					}
				}
			}

			if documents, ok := assets["documents"].([]interface{}); ok {
				for _, doc := range documents {
					if docMap, ok := doc.(map[string]interface{}); ok {
						// Extract document downloadUrl (this is the PDF/document file URL)
						if url, ok := docMap["downloadUrl"].(string); ok && url != "" {
							if p.Image == "" {
								p.Image = url
							}
							p.Media = append(p.Media, url)
						}
					}
				}
			}
		}
	}

	// Build article URL
	p.ArticleURL = fmt.Sprintf("https://www.linkedin.com/feed/update/%s/", p.Activity)
	// Parse content and determine media type
	if content, ok := m["content"].(map[string]interface{}); ok && len(content) > 0 {
		contentType := ""
		for k := range content {
			contentType = k
			break
		}

		switch contentType {
		case "multiImage":
			p.MediaType = "images"
			if _, ok := content["multiImage"].(map[string]interface{}); ok {
				// Images will be populated during enrichment
				p.Media = []string{}
			}
		case "media":
			if media, ok := content["media"].(map[string]interface{}); ok {
				if id, ok := media["id"].(string); ok {
					if strings.Contains(id, "video") {
						p.MediaType = "videos"
					} else if strings.Contains(id, "document") {
						p.MediaType = "carousel"
					} else {
						p.MediaType = "images"
					}
				}
			}
		case "article":
			p.MediaType = "link"

			if article, ok := content["article"].(map[string]interface{}); ok {
				if title, ok := article["title"].(string); ok {
					p.ArticleTitle = title
				}
			}
		case "poll":
			p.MediaType = "poll"
			if poll, ok := content["poll"].(map[string]interface{}); ok {
				if source, ok := poll["source"].(string); ok {
					p.ArticleURL = source
				}
				if title, ok := poll["title"].(string); ok {
					p.ArticleTitle = title
				} else if question, ok := poll["question"].(string); ok {
					p.ArticleTitle = question
				}
				if options, ok := poll["options"]; ok {
					if data, err := json.Marshal(options); err == nil {
						p.PollData = string(data)
					}
				}
			}
		default:
			p.MediaType = "text"
		}
	} else {
		p.MediaType = "text"
	}

	return p, nil
}

// ParseMediaAsset decodes image/video asset raw JSON.
func ParseMediaAsset(raw json.RawMessage) (*kafkamodels.ParsedLinkedinMediaAsset, error) {
	var m struct {
		ID          string `json:"id"`
		DownloadURL string `json:"downloadUrl"`
		Thumbnail   string `json:"thumbnail,omitempty"`
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	t := "image"
	if strings.Contains(m.ID, "video") {
		t = "video"
	}
	return &kafkamodels.ParsedLinkedinMediaAsset{
		ID:          m.ID,
		DownloadURL: m.DownloadURL,
		Thumbnail:   m.Thumbnail,
		Type:        t,
	}, nil
}

// ParseStat decodes share statistics object.
// ParseStatBatch decodes a LinkedIn stats payload that may be either
// a batch (with "elements") or a single stats object.
func ParseStat(raw json.RawMessage) ([]*kafkamodels.ParsedLinkedinStat, error) {
	// Batch shape
	type elem struct {
		UGCPost string `json:"ugcPost"`
		Share   string `json:"share"`
		Total   struct {
			CommentCount           int64 `json:"commentCount"`
			LikeCount              int64 `json:"likeCount"`
			UniqueImpressionsCount int64 `json:"uniqueImpressionsCount"`
			ShareCount             int64 `json:"shareCount"`
			ClickCount             int64 `json:"clickCount"`
			ImpressionCount        int64 `json:"impressionCount"`
			// engagement can be float; we don't map it (no field in model)
			// Engagement            float64 `json:"engagement"`
		} `json:"totalShareStatistics"`
	}
	var batch struct {
		Elements []elem `json:"elements"`
	}

	// Try batch decode first
	if err := json.Unmarshal(raw, &batch); err == nil && len(batch.Elements) > 0 {
		out := make([]*kafkamodels.ParsedLinkedinStat, 0, len(batch.Elements))

		for _, e := range batch.Elements {
			id := e.UGCPost
			if id == "" {
				id = e.Share
			}
			if id == "" {
				continue
			}
			out = append(out, &kafkamodels.ParsedLinkedinStat{
				ActivityID:             id,
				CommentCount:           e.Total.CommentCount,
				LikeCount:              e.Total.LikeCount,
				UniqueImpressionsCount: e.Total.UniqueImpressionsCount,
				ShareCount:             e.Total.ShareCount,
				ClickCount:             e.Total.ClickCount,
				ImpressionCount:        e.Total.ImpressionCount,
			})
		}

		return out, nil
	}

	// Fallback: single object shape
	var single struct {
		UGCPost string `json:"ugcPost"`
		Share   string `json:"share"`
		Total   struct {
			CommentCount           int64 `json:"commentCount"`
			LikeCount              int64 `json:"likeCount"`
			UniqueImpressionsCount int64 `json:"uniqueImpressionsCount"`
			ShareCount             int64 `json:"shareCount"`
			ClickCount             int64 `json:"clickCount"`
			ImpressionCount        int64 `json:"impressionCount"`
		} `json:"totalShareStatistics"`
	}
	if err := json.Unmarshal(raw, &single); err != nil {
		return nil, err
	}
	id := single.UGCPost
	if id == "" {
		id = single.Share
	}
	if id == "" {
		// nothing usable
		return nil, nil
	}
	return []*kafkamodels.ParsedLinkedinStat{{

		ActivityID:             id,
		CommentCount:           single.Total.CommentCount,
		LikeCount:              single.Total.LikeCount,
		UniqueImpressionsCount: single.Total.UniqueImpressionsCount,
		ShareCount:             single.Total.ShareCount,
		ClickCount:             single.Total.ClickCount,
		ImpressionCount:        single.Total.ImpressionCount,
	}}, nil
}

// EnrichPostWithStats adds statistics to a parsed post.
func EnrichPostWithStats(post *kafkamodels.ParsedLinkedinPost, stats map[string]*kafkamodels.ParsedLinkedinStat) {
	if post == nil || stats == nil {
		return
	}

	if stat, ok := stats[post.Activity]; ok {
		post.Comments = stat.CommentCount
		post.Favorites = stat.LikeCount
		post.Reach = stat.UniqueImpressionsCount
		post.Repost = stat.ShareCount
		post.PostClicks = stat.ClickCount
		post.Impressions = stat.ImpressionCount
		// post.TotalEngagement = stat.CommentCount + stat.LikeCount + stat.ShareCount
	}
}

// EnrichPostWithMedia adds media URLs to a parsed post.
func EnrichPostWithMedia(post *kafkamodels.ParsedLinkedinPost, media map[string]*kafkamodels.ParsedLinkedinMediaAsset) {
	if post == nil || media == nil {
		return
	}

	// This would be called after parsing media assets to populate image/media fields
	// The logic depends on how media IDs are stored in the post content
}

// ParseInsights converts raw LinkedIn merged insights JSON into daily ParsedLinkedinInsights buckets.
// The input "raw" contains both followerData and pageStatistics merged by the fetcher.
// Returns one insight record per day based on the timeRange in the API response.
func ParseInsights(raw json.RawMessage) (*kafkamodels.ParsedLinkedinInsights, error) {
	results, err := ParseInsightsDaily(raw)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	// Return first one for backward compatibility
	return results[0], nil
}

// ParseInsightsDaily converts raw LinkedIn merged insights JSON into daily ParsedLinkedinInsights buckets.
// Returns one insight record per day based on the timeRange in the API response.
//
// Data sources for Pages:
//   - pageStatistics: Daily page views from organizationPageStatistics API (with time range)
//   - shareStatistics: Daily engagement metrics from organizationalEntityShareStatistics API (with time range)
//   - followerData: Snapshot follower demographics from organizationalEntityFollowerStatistics API (no time range)
//
// Data sources for Profiles:
//   - impressionData: Daily impression counts from memberCreatorPostAnalytics API (IMPRESSION)
//   - membersReachedData: Member reach data from memberCreatorPostAnalytics API (MEMBERS_REACHED) - no daily aggregation
//   - reshareData: Daily reshare counts from memberCreatorPostAnalytics API (RESHARE)
//   - reactionData: Daily reaction counts from memberCreatorPostAnalytics API (REACTION)
//   - commentData: Daily comment counts from memberCreatorPostAnalytics API (COMMENT)
//   - followerData: Daily follower counts from memberFollowersCount API
//
// Processing logic:
//  1. Page stats provide daily buckets (one record per day with page views)
//  2. Share stats are matched to page stats by date (timeRange.start) and merged
//  3. Follower data is a snapshot - same values are applied to ALL daily buckets
//  4. Each output record represents one day with combined page views, engagement, and follower data
func ParseInsightsDaily(raw json.RawMessage) ([]*kafkamodels.ParsedLinkedinInsights, error) {
	var entityCheck struct {
		EntityType string `json:"entityType"`
	}
	if err := json.Unmarshal(raw, &entityCheck); err == nil && entityCheck.EntityType == "profile" {
		return parseProfileInsightsDaily(raw)
	}

	var merged struct {
		FollowerData    json.RawMessage `json:"followerData"`
		PageStatistics  json.RawMessage `json:"pageStatistics"`
		ShareStatistics json.RawMessage `json:"shareStatistics"`
	}

	if err := json.Unmarshal(raw, &merged); err != nil {
		ins, err := parseFollowerDataLegacy(raw)
		if err != nil || ins == nil {
			return nil, err
		}
		return []*kafkamodels.ParsedLinkedinInsights{ins}, nil
	}

	// If no fields are present, this is likely old format - use legacy parser
	if len(merged.FollowerData) == 0 && len(merged.PageStatistics) == 0 && len(merged.ShareStatistics) == 0 {
		ins, err := parseFollowerDataLegacy(raw)
		if err != nil || ins == nil {
			return nil, err
		}
		return []*kafkamodels.ParsedLinkedinInsights{ins}, nil
	}

	// Parse follower data (single snapshot - will be applied to all daily buckets)
	var followerData *followerSnapshot
	if len(merged.FollowerData) > 0 && string(merged.FollowerData) != "null" {
		followerData = parseFollowerDataSnapshot(merged.FollowerData)
	}

	// Parse share statistics into daily buckets (engagement metrics)
	shareStatsByDate := make(map[int64]*shareStatsSnapshot)
	if len(merged.ShareStatistics) > 0 && string(merged.ShareStatistics) != "null" {
		shareStatsByDate = parseShareStatisticsDaily(merged.ShareStatistics)
	}

	// Parse page statistics into daily buckets (page views)
	var results []*kafkamodels.ParsedLinkedinInsights
	if len(merged.PageStatistics) > 0 && string(merged.PageStatistics) != "null" {
		results = parsePageStatisticsDaily(merged.PageStatistics)
	}

	// If we have share stats but no page stats, create records from share stats
	if len(results) == 0 && len(shareStatsByDate) > 0 {
		for startMs, stats := range shareStatsByDate {
			createdAt := time.UnixMilli(startMs).UTC()
			ins := &kafkamodels.ParsedLinkedinInsights{
				InsertedAt: time.Now().UTC(),
				CreatedAt:  createdAt,
			}
			applyShareSnapshot(ins, stats)
			results = append(results, ins)
		}
	}

	// If still no results, create a single record with follower data
	if len(results) == 0 {
		createdAt := time.Now().UTC().Truncate(24 * time.Hour)
		ins := &kafkamodels.ParsedLinkedinInsights{
			InsertedAt: time.Now().UTC(),
			CreatedAt:  createdAt,
		}
		if followerData != nil {
			applyFollowerSnapshot(ins, followerData)
		}
		return []*kafkamodels.ParsedLinkedinInsights{ins}, nil
	}

	// Apply follower data and share stats to all daily buckets
	for _, ins := range results {
		if followerData != nil {
			applyFollowerSnapshot(ins, followerData)
		}
		// Apply share stats for matching date
		startMs := ins.CreatedAt.UnixMilli()
		if stats, ok := shareStatsByDate[startMs]; ok {
			applyShareSnapshot(ins, stats)
		}
	}

	return results, nil
}

// parseProfileInsightsDaily parses profile insights from memberCreatorPostAnalytics and memberFollowersCount APIs.

// into daily insight records. Each daily record contains that day's metrics and calculates totals.
func parseProfileInsightsDaily(raw json.RawMessage) ([]*kafkamodels.ParsedLinkedinInsights, error) {
	var merged struct {
		EntityType         string          `json:"entityType"`
		ImpressionData     json.RawMessage `json:"impressionData"`
		MembersReachedData json.RawMessage `json:"membersReachedData"`
		ReshareData        json.RawMessage `json:"reshareData"`
		ReactionData       json.RawMessage `json:"reactionData"`
		CommentData        json.RawMessage `json:"commentData"`
		FollowerData       json.RawMessage `json:"followerData"`
		TotalFollowerData  json.RawMessage `json:"totalFollowerData"`
	}

	if err := json.Unmarshal(raw, &merged); err != nil {
		return nil, err
	}

	// Parse daily data from each analytics type
	impressionByDate := parseProfileAnalyticsDaily(merged.ImpressionData)
	reshareByDate := parseProfileAnalyticsDaily(merged.ReshareData)
	reactionByDate := parseProfileAnalyticsDaily(merged.ReactionData)
	commentByDate := parseProfileAnalyticsDaily(merged.CommentData)
	followerByDate := parseProfileFollowerDaily(merged.FollowerData)

	// Parse members reached (total value, no daily breakdown)
	var membersReached int64
	if len(merged.MembersReachedData) > 0 && string(merged.MembersReachedData) != "null" {
		var reachResp struct {
			Elements []struct {
				Count int64 `json:"count"`
			} `json:"elements"`
		}
		if err := json.Unmarshal(merged.MembersReachedData, &reachResp); err == nil && len(reachResp.Elements) > 0 {
			membersReached = reachResp.Elements[0].Count
		}
	}

	// Parse total follower count (from q=me API)
	var totalFollowerCount int64
	if len(merged.TotalFollowerData) > 0 && string(merged.TotalFollowerData) != "null" {
		var followerResp struct {
			Elements []struct {
				MemberFollowersCount int64 `json:"memberFollowersCount"`
			} `json:"elements"`
		}
		if err := json.Unmarshal(merged.TotalFollowerData, &followerResp); err == nil && len(followerResp.Elements) > 0 {
			totalFollowerCount = followerResp.Elements[0].MemberFollowersCount
		}
	}

	// Collect all unique dates from all sources
	dateSet := make(map[string]time.Time)
	for dateStr, data := range impressionByDate {
		dateSet[dateStr] = data.Date
	}
	for dateStr, data := range reshareByDate {
		dateSet[dateStr] = data.Date
	}
	for dateStr, data := range reactionByDate {
		dateSet[dateStr] = data.Date
	}
	for dateStr, data := range commentByDate {
		dateSet[dateStr] = data.Date
	}
	for dateStr, data := range followerByDate {
		dateSet[dateStr] = data.Date
	}

	if len(dateSet) == 0 {
		return nil, nil
	}

	// Create one insight record per day
	results := make([]*kafkamodels.ParsedLinkedinInsights, 0, len(dateSet))
	for dateStr, date := range dateSet {
		ins := &kafkamodels.ParsedLinkedinInsights{
			InsertedAt: time.Now().UTC(),
			CreatedAt:  date,
		}

		// Set daily values
		if data, ok := impressionByDate[dateStr]; ok {
			ins.ImpressionCount = data.Count
		}
		if data, ok := commentByDate[dateStr]; ok {
			ins.Comments = data.Count
		}
		if data, ok := reactionByDate[dateStr]; ok {
			ins.Reactions = data.Count
		}
		if data, ok := reshareByDate[dateStr]; ok {
			ins.Repost = data.Count
		}
		if data, ok := followerByDate[dateStr]; ok {
			ins.DailyFollowerCount = data.Count
		}

		// Members reached is total (not daily), duplicate across all days
		ins.Reach = membersReached

		// Total follower count from q=me API - same for all daily records
		ins.TotalFollowerCount = totalFollowerCount

		results = append(results, ins)
	}

	return results, nil
}

// profileDailyData holds daily count data from memberCreatorPostAnalytics or memberFollowersCount
type profileDailyData struct {
	Date  time.Time
	Count int64
}

// parseProfileAnalyticsDaily parses memberCreatorPostAnalytics response into a map of date -> count
// The response has elements with dateRange and count
func parseProfileAnalyticsDaily(raw json.RawMessage) map[string]*profileDailyData {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}

	var resp struct {
		Elements []struct {
			DateRange struct {
				Start struct {
					Day   int `json:"day"`
					Month int `json:"month"`
					Year  int `json:"year"`
				} `json:"start"`
			} `json:"dateRange"`
			Count int64 `json:"count"`
		} `json:"elements"`
	}

	if err := json.Unmarshal(raw, &resp); err != nil || len(resp.Elements) == 0 {
		return nil
	}

	result := make(map[string]*profileDailyData, len(resp.Elements))
	for _, el := range resp.Elements {
		date := time.Date(el.DateRange.Start.Year, time.Month(el.DateRange.Start.Month), el.DateRange.Start.Day, 0, 0, 0, 0, time.UTC)
		dateStr := date.Format("2006-01-02")
		result[dateStr] = &profileDailyData{
			Date:  date,
			Count: el.Count,
		}
	}
	return result
}

// parseProfileFollowerDaily parses memberFollowersCount response into a map of date -> count
// The response has elements with dateRange and memberFollowersCount
func parseProfileFollowerDaily(raw json.RawMessage) map[string]*profileDailyData {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}

	var resp struct {
		Elements []struct {
			DateRange struct {
				Start struct {
					Day   int `json:"day"`
					Month int `json:"month"`
					Year  int `json:"year"`
				} `json:"start"`
			} `json:"dateRange"`
			MemberFollowersCount int64 `json:"memberFollowersCount"`
		} `json:"elements"`
	}

	if err := json.Unmarshal(raw, &resp); err != nil || len(resp.Elements) == 0 {
		return nil
	}

	result := make(map[string]*profileDailyData, len(resp.Elements))
	for _, el := range resp.Elements {
		date := time.Date(el.DateRange.Start.Year, time.Month(el.DateRange.Start.Month), el.DateRange.Start.Day, 0, 0, 0, 0, time.UTC)
		dateStr := date.Format("2006-01-02")
		result[dateStr] = &profileDailyData{
			Date:  date,
			Count: el.MemberFollowersCount,
		}
	}
	return result
}

// followerSnapshot holds parsed follower data from organizationalEntityFollowerStatistics API (no time range).
// This is a snapshot of current follower demographics - same values are applied to ALL daily insight buckets.
// Contains total/organic/paid follower counts and demographic breakdowns (seniority, industry, country, city).
type followerSnapshot struct {
	TotalFollowerCount   int64
	OrganicFollowerCount int64
	PaidFollowerCount    int64
	FollowersBySeniority string
	FollowersByIndustry  string
	FollowersByCountry   string
	FollowersByCity      string
}

func applyFollowerSnapshot(ins *kafkamodels.ParsedLinkedinInsights, snap *followerSnapshot) {
	ins.TotalFollowerCount = snap.TotalFollowerCount
	ins.OrganicFollowerCount = snap.OrganicFollowerCount
	ins.PaidFollowerCount = snap.PaidFollowerCount
	ins.FollowersBySeniority = snap.FollowersBySeniority
	ins.FollowersByIndustry = snap.FollowersByIndustry
	ins.FollowersByCountry = snap.FollowersByCountry
	ins.FollowersByCity = snap.FollowersByCity
}

// shareStatsSnapshot holds daily share statistics (engagement metrics) from organizationalEntityShareStatistics API.
// Each day has its own engagement metrics - impressions, clicks, likes, comments, shares, and engagement rate.
type shareStatsSnapshot struct {
	ImpressionCount        int64
	UniqueImpressionsCount int64
	ClickCount             int64
	LikeCount              int64
	CommentCount           int64
	ShareCount             int64
	Engagement             float64
}

func applyShareSnapshot(ins *kafkamodels.ParsedLinkedinInsights, snap *shareStatsSnapshot) {
	ins.ImpressionCount = snap.ImpressionCount
	ins.Reach = snap.UniqueImpressionsCount
	ins.PostClicks = snap.ClickCount
	ins.Comments = snap.CommentCount
	ins.Repost = snap.ShareCount
	ins.Reactions = snap.LikeCount
	ins.Engagement = snap.Engagement
}

// parseShareStatisticsDaily parses organizationalEntityShareStatistics API response and returns a map of date -> stats.
// The API returns daily elements with timeRange.start as the day's timestamp (milliseconds).
// Each element contains totalShareStatistics with engagement metrics for that day.
func parseShareStatisticsDaily(raw json.RawMessage) map[int64]*shareStatsSnapshot {
	var resp struct {
		Elements []struct {
			TotalShareStatistics struct {
				UniqueImpressionsCount int64   `json:"uniqueImpressionsCount"`
				ShareCount             int64   `json:"shareCount"`
				Engagement             float64 `json:"engagement"`
				ClickCount             int64   `json:"clickCount"`
				LikeCount              int64   `json:"likeCount"`
				ImpressionCount        int64   `json:"impressionCount"`
				CommentCount           int64   `json:"commentCount"`
			} `json:"totalShareStatistics"`
			TimeRange struct {
				Start int64 `json:"start"`
				End   int64 `json:"end"`
			} `json:"timeRange"`
		} `json:"elements"`
	}

	if err := json.Unmarshal(raw, &resp); err != nil || len(resp.Elements) == 0 {
		return nil
	}

	result := make(map[int64]*shareStatsSnapshot, len(resp.Elements))
	for _, el := range resp.Elements {
		result[el.TimeRange.Start] = &shareStatsSnapshot{
			ImpressionCount:        el.TotalShareStatistics.ImpressionCount,
			UniqueImpressionsCount: el.TotalShareStatistics.UniqueImpressionsCount,
			ClickCount:             el.TotalShareStatistics.ClickCount,
			LikeCount:              el.TotalShareStatistics.LikeCount,
			CommentCount:           el.TotalShareStatistics.CommentCount,
			ShareCount:             el.TotalShareStatistics.ShareCount,
			Engagement:             el.TotalShareStatistics.Engagement,
		}
	}
	return result
}

func parseFollowerDataSnapshot(raw json.RawMessage) *followerSnapshot {
	var resp struct {
		FirstDegreeSize int64             `json:"firstDegreeSize"`
		GeoNames        map[string]string `json:"geoNames"` // Resolved geo ID -> name mapping from LinkedIn Geo API
		Elements        []struct {
			FollowerCountsBySeniority []struct {
				Seniority      string `json:"seniority"`
				FollowerCounts struct {
					Organic int64 `json:"organicFollowerCount"`
					Paid    int64 `json:"paidFollowerCount"`
				} `json:"followerCounts"`
			} `json:"followerCountsBySeniority"`
			FollowerCountsByIndustry []struct {
				Industry       string `json:"industry"`
				FollowerCounts struct {
					Organic int64 `json:"organicFollowerCount"`
					Paid    int64 `json:"paidFollowerCount"`
				} `json:"followerCounts"`
			} `json:"followerCountsByIndustry"`
			FollowerCountsByGeoCountry []struct {
				Geo            string `json:"geo"`
				FollowerCounts struct {
					Organic int64 `json:"organicFollowerCount"`
					Paid    int64 `json:"paidFollowerCount"`
				} `json:"followerCounts"`
			} `json:"followerCountsByGeoCountry"`
			FollowerCountsByGeo []struct {
				Geo            string `json:"geo"`
				FollowerCounts struct {
					Organic int64 `json:"organicFollowerCount"`
					Paid    int64 `json:"paidFollowerCount"`
				} `json:"followerCounts"`
			} `json:"followerCountsByGeo"`
		} `json:"elements"`
	}

	if err := json.Unmarshal(raw, &resp); err != nil || len(resp.Elements) == 0 {
		return nil
	}

	bySeniority := map[string]int64{}
	byIndustry := map[string]int64{}
	byCountry := map[string]int64{}
	byCity := map[string]int64{}

	var totalPaid, totalOrganic int64

	seniors := constants.GetSeniorities()
	inds := constants.GetIndustries()
	countries := constants.GetCountries()

	// Use resolved geo names if available, otherwise fall back to constants or numeric IDs
	geoNames := resp.GeoNames
	if geoNames == nil {
		geoNames = map[string]string{}
	}

	for _, el := range resp.Elements {
		for _, s := range el.FollowerCountsBySeniority {
			name := seniors[s.Seniority]
			if name == "" {
				name = s.Seniority
			}
			bySeniority[name] += s.FollowerCounts.Organic + s.FollowerCounts.Paid
		}
		for _, ind := range el.FollowerCountsByIndustry {
			name := inds[ind.Industry]
			if name == "" {
				name = ind.Industry
			}
			byIndustry[name] += ind.FollowerCounts.Organic + ind.FollowerCounts.Paid
		}
		for _, c := range el.FollowerCountsByGeoCountry {
			urn := c.Geo
			code := urn[strings.LastIndex(urn, ":")+1:]
			// Priority: 1. Resolved geo name, 2. Constants lookup, 3. Numeric ID
			name := geoNames[code]
			if name == "" {
				name = countries[urn]
			}
			if name == "" {
				name = code
			}
			byCountry[name] += c.FollowerCounts.Organic + c.FollowerCounts.Paid
			totalOrganic += c.FollowerCounts.Organic
			totalPaid += c.FollowerCounts.Paid
		}
		for _, g := range el.FollowerCountsByGeo {
			urn := g.Geo
			code := urn[strings.LastIndex(urn, ":")+1:]
			// Priority: 1. Resolved geo name, 2. Numeric ID (cities don't have constants)
			name := geoNames[code]
			if name == "" {
				name = code
			}
			byCity[name] += g.FollowerCounts.Organic + g.FollowerCounts.Paid
		}
	}

	senBytes, _ := json.Marshal(bySeniority)
	indBytes, _ := json.Marshal(byIndustry)
	couBytes, _ := json.Marshal(byCountry)
	cityBytes, _ := json.Marshal(byCity)

	return &followerSnapshot{
		TotalFollowerCount:   resp.FirstDegreeSize,
		OrganicFollowerCount: resp.FirstDegreeSize - totalPaid,
		PaidFollowerCount:    totalPaid,
		FollowersBySeniority: string(senBytes),
		FollowersByIndustry:  string(indBytes),
		FollowersByCountry:   string(couBytes),
		FollowersByCity:      string(cityBytes),
	}
}

func parseFollowerDataLegacy(raw json.RawMessage) (*kafkamodels.ParsedLinkedinInsights, error) {
	var resp struct {
		FirstDegreeSize int64             `json:"firstDegreeSize"`
		GeoNames        map[string]string `json:"geoNames"` // Resolved geo ID -> name mapping from LinkedIn Geo API
		Elements        []struct {
			FollowerCountsBySeniority []struct {
				Seniority      string `json:"seniority"`
				FollowerCounts struct {
					Organic int64 `json:"organicFollowerCount"`
					Paid    int64 `json:"paidFollowerCount"`
				} `json:"followerCounts"`
			} `json:"followerCountsBySeniority"`
			FollowerCountsByIndustry []struct {
				Industry       string `json:"industry"`
				FollowerCounts struct {
					Organic int64 `json:"organicFollowerCount"`
					Paid    int64 `json:"paidFollowerCount"`
				} `json:"followerCounts"`
			} `json:"followerCountsByIndustry"`
			FollowerCountsByGeoCountry []struct {
				Geo            string `json:"geo"`
				FollowerCounts struct {
					Organic int64 `json:"organicFollowerCount"`
					Paid    int64 `json:"paidFollowerCount"`
				} `json:"followerCounts"`
			} `json:"followerCountsByGeoCountry"`
			FollowerCountsByGeo []struct {
				Geo            string `json:"geo"`
				FollowerCounts struct {
					Organic int64 `json:"organicFollowerCount"`
					Paid    int64 `json:"paidFollowerCount"`
				} `json:"followerCounts"`
			} `json:"followerCountsByGeo"`
		} `json:"elements"`
	}

	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	if len(resp.Elements) == 0 {
		return nil, nil
	}

	bySeniority := map[string]int64{}
	byIndustry := map[string]int64{}
	byCountry := map[string]int64{}
	byCity := map[string]int64{}

	var totalPaid, totalOrganic int64

	seniors := constants.GetSeniorities()
	inds := constants.GetIndustries()
	countries := constants.GetCountries()

	// Use resolved geo names if available, otherwise fall back to constants or numeric IDs
	geoNames := resp.GeoNames
	if geoNames == nil {
		geoNames = map[string]string{}
	}

	for _, el := range resp.Elements {
		for _, s := range el.FollowerCountsBySeniority {
			name := seniors[s.Seniority]
			if name == "" {
				name = s.Seniority
			}
			bySeniority[name] += s.FollowerCounts.Organic + s.FollowerCounts.Paid
		}
		for _, ind := range el.FollowerCountsByIndustry {
			name := inds[ind.Industry]
			if name == "" {
				name = ind.Industry
			}
			byIndustry[name] += ind.FollowerCounts.Organic + ind.FollowerCounts.Paid
		}
		for _, c := range el.FollowerCountsByGeoCountry {
			urn := c.Geo
			code := urn[strings.LastIndex(urn, ":")+1:]
			// Priority: 1. Resolved geo name, 2. Constants lookup, 3. Numeric ID
			name := geoNames[code]
			if name == "" {
				name = countries[urn]
			}
			if name == "" {
				name = code
			}
			byCountry[name] += c.FollowerCounts.Organic + c.FollowerCounts.Paid
			totalOrganic += totalOrganic + c.FollowerCounts.Organic
			totalPaid += totalPaid + c.FollowerCounts.Paid
		}
		for _, g := range el.FollowerCountsByGeo {
			urn := g.Geo
			code := urn[strings.LastIndex(urn, ":")+1:]
			// Priority: 1. Resolved geo name, 2. Numeric ID (cities don't have constants)
			name := geoNames[code]
			if name == "" {
				name = code
			}
			byCity[name] += g.FollowerCounts.Organic + g.FollowerCounts.Paid
		}
	}

	senBytes, _ := json.Marshal(bySeniority)
	indBytes, _ := json.Marshal(byIndustry)
	couBytes, _ := json.Marshal(byCountry)
	cityBytes, _ := json.Marshal(byCity)

	ins := &kafkamodels.ParsedLinkedinInsights{
		FollowersBySeniority: string(senBytes),
		FollowersByIndustry:  string(indBytes),
		FollowersByCountry:   string(couBytes),
		FollowersByCity:      string(cityBytes),
		PaidFollowerCount:    totalPaid,
		OrganicFollowerCount: resp.FirstDegreeSize - totalPaid,
		TotalFollowerCount:   resp.FirstDegreeSize,
		InsertedAt:           time.Now().UTC().Truncate(24 * time.Hour),
	}

	return ins, nil
}

func parseFollowerDataIntoInsights(raw json.RawMessage, ins *kafkamodels.ParsedLinkedinInsights) {
	var resp struct {
		FirstDegreeSize int64             `json:"firstDegreeSize"`
		GeoNames        map[string]string `json:"geoNames"` // Resolved geo ID -> name mapping from LinkedIn Geo API
		Elements        []struct {
			FollowerCountsBySeniority []struct {
				Seniority      string `json:"seniority"`
				FollowerCounts struct {
					Organic int64 `json:"organicFollowerCount"`
					Paid    int64 `json:"paidFollowerCount"`
				} `json:"followerCounts"`
			} `json:"followerCountsBySeniority"`
			FollowerCountsByIndustry []struct {
				Industry       string `json:"industry"`
				FollowerCounts struct {
					Organic int64 `json:"organicFollowerCount"`
					Paid    int64 `json:"paidFollowerCount"`
				} `json:"followerCounts"`
			} `json:"followerCountsByIndustry"`
			FollowerCountsByGeoCountry []struct {
				Geo            string `json:"geo"`
				FollowerCounts struct {
					Organic int64 `json:"organicFollowerCount"`
					Paid    int64 `json:"paidFollowerCount"`
				} `json:"followerCounts"`
			} `json:"followerCountsByGeoCountry"`
			FollowerCountsByGeo []struct {
				Geo            string `json:"geo"`
				FollowerCounts struct {
					Organic int64 `json:"organicFollowerCount"`
					Paid    int64 `json:"paidFollowerCount"`
				} `json:"followerCounts"`
			} `json:"followerCountsByGeo"`
		} `json:"elements"`
	}

	if err := json.Unmarshal(raw, &resp); err != nil || len(resp.Elements) == 0 {
		return
	}

	bySeniority := map[string]int64{}
	byIndustry := map[string]int64{}
	byCountry := map[string]int64{}
	byCity := map[string]int64{}

	var totalPaid, totalOrganic int64

	seniors := constants.GetSeniorities()
	inds := constants.GetIndustries()
	countries := constants.GetCountries()

	// Use resolved geo names if available, otherwise fall back to constants or numeric IDs
	geoNames := resp.GeoNames
	if geoNames == nil {
		geoNames = map[string]string{}
	}

	for _, el := range resp.Elements {
		for _, s := range el.FollowerCountsBySeniority {
			name := seniors[s.Seniority]
			if name == "" {
				name = s.Seniority
			}
			bySeniority[name] += s.FollowerCounts.Organic + s.FollowerCounts.Paid
		}
		for _, ind := range el.FollowerCountsByIndustry {
			name := inds[ind.Industry]
			if name == "" {
				name = ind.Industry
			}
			byIndustry[name] += ind.FollowerCounts.Organic + ind.FollowerCounts.Paid
		}
		for _, c := range el.FollowerCountsByGeoCountry {
			urn := c.Geo
			code := urn[strings.LastIndex(urn, ":")+1:]
			// Priority: 1. Resolved geo name, 2. Constants lookup, 3. Numeric ID
			name := geoNames[code]
			if name == "" {
				name = countries[urn]
			}
			if name == "" {
				name = code
			}
			byCountry[name] += c.FollowerCounts.Organic + c.FollowerCounts.Paid
			totalOrganic += totalOrganic + c.FollowerCounts.Organic
			totalPaid += totalPaid + c.FollowerCounts.Paid
		}
		for _, g := range el.FollowerCountsByGeo {
			urn := g.Geo
			code := urn[strings.LastIndex(urn, ":")+1:]
			// Priority: 1. Resolved geo name, 2. Numeric ID (cities don't have constants)
			name := geoNames[code]
			if name == "" {
				name = code
			}
			byCity[name] += g.FollowerCounts.Organic + g.FollowerCounts.Paid
		}
	}

	senBytes, _ := json.Marshal(bySeniority)
	indBytes, _ := json.Marshal(byIndustry)
	couBytes, _ := json.Marshal(byCountry)
	cityBytes, _ := json.Marshal(byCity)

	ins.FollowersBySeniority = string(senBytes)
	ins.FollowersByIndustry = string(indBytes)
	ins.FollowersByCountry = string(couBytes)
	ins.FollowersByCity = string(cityBytes)
	ins.PaidFollowerCount = totalPaid
	ins.OrganicFollowerCount = resp.FirstDegreeSize - totalPaid
	ins.TotalFollowerCount = resp.FirstDegreeSize
}

// parsePageStatisticsDaily parses organizationPageStatistics API response and returns one insight per day.
// The API returns daily elements with timeRange.start as the day's timestamp (milliseconds).
// Each element contains pageViews and uniquePageViews by device type and other dimensions.
// The CreatedAt field is set from timeRange.start to identify which day the data represents.
func parsePageStatisticsDaily(raw json.RawMessage) []*kafkamodels.ParsedLinkedinInsights {
	type pageViewsData struct {
		PageViews       int64 `json:"pageViews"`
		UniquePageViews int64 `json:"uniquePageViews"`
	}
	type viewsObj struct {
		AllPageViews        pageViewsData `json:"allPageViews"`
		AllDesktopPageViews pageViewsData `json:"allDesktopPageViews"`
		AllMobilePageViews  pageViewsData `json:"allMobilePageViews"`
		OverviewPageViews   pageViewsData `json:"overviewPageViews"`
		AboutPageViews      pageViewsData `json:"aboutPageViews"`
		JobsPageViews       pageViewsData `json:"jobsPageViews"`
		PeoplePageViews     pageViewsData `json:"peoplePageViews"`
		CareersPageViews    pageViewsData `json:"careersPageViews"`
		LifeAtPageViews     pageViewsData `json:"lifeAtPageViews"`
		InsightsPageViews   pageViewsData `json:"insightsPageViews"`
		ProductsPageViews   pageViewsData `json:"productsPageViews"`
	}

	var resp struct {
		Elements []struct {
			TimeRange struct {
				Start int64 `json:"start"`
				End   int64 `json:"end"`
			} `json:"timeRange"`
			TotalPageStatistics struct {
				Views viewsObj `json:"views"`
			} `json:"totalPageStatistics"`
			PageStatisticsByGeoCountry []struct {
				Geo            string `json:"geo"`
				PageStatistics struct {
					Views struct {
						AllPageViews pageViewsData `json:"allPageViews"`
					} `json:"views"`
				} `json:"pageStatistics"`
			} `json:"pageStatisticsByGeoCountry"`
			PageStatisticsByGeo []struct {
				Geo            string `json:"geo"`
				PageStatistics struct {
					Views struct {
						AllPageViews pageViewsData `json:"allPageViews"`
					} `json:"views"`
				} `json:"pageStatistics"`
			} `json:"pageStatisticsByGeo"`
			PageStatisticsByIndustryV2 []struct {
				IndustryV2     string `json:"industryV2"`
				PageStatistics struct {
					Views struct {
						AllPageViews pageViewsData `json:"allPageViews"`
					} `json:"views"`
				} `json:"pageStatistics"`
			} `json:"pageStatisticsByIndustryV2"`
			PageStatisticsBySeniority []struct {
				Seniority      string `json:"seniority"`
				PageStatistics struct {
					Views struct {
						AllPageViews pageViewsData `json:"allPageViews"`
					} `json:"views"`
				} `json:"pageStatistics"`
			} `json:"pageStatisticsBySeniority"`
			PageStatisticsByFunction []struct {
				Function       string `json:"function"`
				PageStatistics struct {
					Views struct {
						AllPageViews pageViewsData `json:"allPageViews"`
					} `json:"views"`
				} `json:"pageStatistics"`
			} `json:"pageStatisticsByFunction"`
			PageStatisticsByStaffCountRange []struct {
				StaffCountRange string `json:"staffCountRange"`
				PageStatistics  struct {
					Views struct {
						AllPageViews pageViewsData `json:"allPageViews"`
					} `json:"views"`
				} `json:"pageStatistics"`
			} `json:"pageStatisticsByStaffCountRange"`
		} `json:"elements"`
	}

	if err := json.Unmarshal(raw, &resp); err != nil || len(resp.Elements) == 0 {
		return nil
	}

	countries := constants.GetCountries()
	seniors := constants.GetSeniorities()
	inds := constants.GetIndustries()
	funcs := constants.GetFunctions()

	results := make([]*kafkamodels.ParsedLinkedinInsights, 0, len(resp.Elements))

	for _, el := range resp.Elements {
		// Extract date from timeRange.start (milliseconds)
		createdAt := time.UnixMilli(el.TimeRange.Start).UTC()

		ins := &kafkamodels.ParsedLinkedinInsights{
			InsertedAt: time.Now().UTC(),
			CreatedAt:  createdAt,
		}

		views := el.TotalPageStatistics.Views
		ins.PageViews = views.AllPageViews.PageViews
		ins.UniqueVisitors = views.AllPageViews.UniquePageViews
		ins.DesktopPageViews = views.AllDesktopPageViews.PageViews
		ins.MobilePageViews = views.AllMobilePageViews.PageViews
		ins.OverviewPageViews = views.OverviewPageViews.PageViews
		ins.AboutPageViews = views.AboutPageViews.PageViews
		ins.JobsPageViews = views.JobsPageViews.PageViews
		ins.PeoplePageViews = views.PeoplePageViews.PageViews
		ins.CareersPageViews = views.CareersPageViews.PageViews
		ins.LifeAtPageViews = views.LifeAtPageViews.PageViews
		ins.InsightsPageViews = views.InsightsPageViews.PageViews
		ins.ProductsPageViews = views.ProductsPageViews.PageViews

		// Parse demographics for this day
		byCountry := map[string]int64{}
		byRegion := map[string]int64{}
		byIndustry := map[string]int64{}
		bySeniority := map[string]int64{}
		byFunction := map[string]int64{}
		byStaffCount := map[string]int64{}

		for _, c := range el.PageStatisticsByGeoCountry {
			name := countries[c.Geo]
			if name == "" {
				name = c.Geo[strings.LastIndex(c.Geo, ":")+1:]
			}
			byCountry[name] += c.PageStatistics.Views.AllPageViews.PageViews
		}

		for _, g := range el.PageStatisticsByGeo {
			regionID := g.Geo[strings.LastIndex(g.Geo, ":")+1:]
			byRegion[regionID] += g.PageStatistics.Views.AllPageViews.PageViews
		}

		for _, ind := range el.PageStatisticsByIndustryV2 {
			name := inds[ind.IndustryV2]
			if name == "" {
				name = ind.IndustryV2[strings.LastIndex(ind.IndustryV2, ":")+1:]
			}
			byIndustry[name] += ind.PageStatistics.Views.AllPageViews.PageViews
		}

		for _, s := range el.PageStatisticsBySeniority {
			name := seniors[s.Seniority]
			if name == "" {
				name = s.Seniority[strings.LastIndex(s.Seniority, ":")+1:]
			}
			bySeniority[name] += s.PageStatistics.Views.AllPageViews.PageViews
		}

		for _, f := range el.PageStatisticsByFunction {
			name := funcs[f.Function]
			if name == "" {
				name = f.Function[strings.LastIndex(f.Function, ":")+1:]
			}
			byFunction[name] += f.PageStatistics.Views.AllPageViews.PageViews
		}

		for _, sc := range el.PageStatisticsByStaffCountRange {
			byStaffCount[sc.StaffCountRange] += sc.PageStatistics.Views.AllPageViews.PageViews
		}

		countryBytes, _ := json.Marshal(byCountry)
		ins.PageViewsByCountry = string(countryBytes)

		regionBytes, _ := json.Marshal(byRegion)
		ins.PageViewsByRegion = string(regionBytes)

		industryBytes, _ := json.Marshal(byIndustry)
		ins.PageViewsByIndustry = string(industryBytes)

		seniorityBytes, _ := json.Marshal(bySeniority)
		ins.PageViewsBySeniority = string(seniorityBytes)

		functionBytes, _ := json.Marshal(byFunction)
		ins.PageViewsByFunction = string(functionBytes)

		staffCountBytes, _ := json.Marshal(byStaffCount)
		ins.PageViewsByStaffCount = string(staffCountBytes)

		results = append(results, ins)
	}

	return results
}

func parsePageStatisticsIntoInsights(raw json.RawMessage, ins *kafkamodels.ParsedLinkedinInsights) {
	type pageViewsData struct {
		PageViews       int64 `json:"pageViews"`
		UniquePageViews int64 `json:"uniquePageViews"`
	}
	type viewsObj struct {
		AllPageViews        pageViewsData `json:"allPageViews"`
		AllDesktopPageViews pageViewsData `json:"allDesktopPageViews"`
		AllMobilePageViews  pageViewsData `json:"allMobilePageViews"`
		OverviewPageViews   pageViewsData `json:"overviewPageViews"`
		AboutPageViews      pageViewsData `json:"aboutPageViews"`
		JobsPageViews       pageViewsData `json:"jobsPageViews"`
		PeoplePageViews     pageViewsData `json:"peoplePageViews"`
		CareersPageViews    pageViewsData `json:"careersPageViews"`
		LifeAtPageViews     pageViewsData `json:"lifeAtPageViews"`
		InsightsPageViews   pageViewsData `json:"insightsPageViews"`
		ProductsPageViews   pageViewsData `json:"productsPageViews"`
	}

	var resp struct {
		Elements []struct {
			TotalPageStatistics struct {
				Views viewsObj `json:"views"`
			} `json:"totalPageStatistics"`
			PageStatisticsByGeoCountry []struct {
				Geo            string `json:"geo"`
				PageStatistics struct {
					Views struct {
						AllPageViews pageViewsData `json:"allPageViews"`
					} `json:"views"`
				} `json:"pageStatistics"`
			} `json:"pageStatisticsByGeoCountry"`
			PageStatisticsByGeo []struct {
				Geo            string `json:"geo"`
				PageStatistics struct {
					Views struct {
						AllPageViews pageViewsData `json:"allPageViews"`
					} `json:"views"`
				} `json:"pageStatistics"`
			} `json:"pageStatisticsByGeo"`
			PageStatisticsByIndustryV2 []struct {
				IndustryV2     string `json:"industryV2"`
				PageStatistics struct {
					Views struct {
						AllPageViews pageViewsData `json:"allPageViews"`
					} `json:"views"`
				} `json:"pageStatistics"`
			} `json:"pageStatisticsByIndustryV2"`
			PageStatisticsBySeniority []struct {
				Seniority      string `json:"seniority"`
				PageStatistics struct {
					Views struct {
						AllPageViews pageViewsData `json:"allPageViews"`
					} `json:"views"`
				} `json:"pageStatistics"`
			} `json:"pageStatisticsBySeniority"`
			PageStatisticsByFunction []struct {
				Function       string `json:"function"`
				PageStatistics struct {
					Views struct {
						AllPageViews pageViewsData `json:"allPageViews"`
					} `json:"views"`
				} `json:"pageStatistics"`
			} `json:"pageStatisticsByFunction"`
			PageStatisticsByStaffCountRange []struct {
				StaffCountRange string `json:"staffCountRange"`
				PageStatistics  struct {
					Views struct {
						AllPageViews pageViewsData `json:"allPageViews"`
					} `json:"views"`
				} `json:"pageStatistics"`
			} `json:"pageStatisticsByStaffCountRange"`
		} `json:"elements"`
	}

	if err := json.Unmarshal(raw, &resp); err != nil || len(resp.Elements) == 0 {
		return
	}

	// Aggregate all elements (each element represents a day when using time intervals)
	// For page views: SUM across all days
	// For unique visitors: SUM across all days (LinkedIn's uniquePageViews per day)
	for _, el := range resp.Elements {
		views := el.TotalPageStatistics.Views

		ins.PageViews += views.AllPageViews.PageViews
		ins.UniqueVisitors += views.AllPageViews.UniquePageViews
		ins.DesktopPageViews += views.AllDesktopPageViews.PageViews
		ins.MobilePageViews += views.AllMobilePageViews.PageViews
		ins.OverviewPageViews += views.OverviewPageViews.PageViews
		ins.AboutPageViews += views.AboutPageViews.PageViews
		ins.JobsPageViews += views.JobsPageViews.PageViews
		ins.PeoplePageViews += views.PeoplePageViews.PageViews
		ins.CareersPageViews += views.CareersPageViews.PageViews
		ins.LifeAtPageViews += views.LifeAtPageViews.PageViews
		ins.InsightsPageViews += views.InsightsPageViews.PageViews
		ins.ProductsPageViews += views.ProductsPageViews.PageViews
	}

	// Aggregate demographic breakdowns across all elements (days)
	countries := constants.GetCountries()
	seniors := constants.GetSeniorities()
	inds := constants.GetIndustries()
	funcs := constants.GetFunctions()

	byCountry := map[string]int64{}
	byRegion := map[string]int64{}
	byIndustry := map[string]int64{}
	bySeniority := map[string]int64{}
	byFunction := map[string]int64{}
	byStaffCount := map[string]int64{}

	for _, el := range resp.Elements {
		for _, c := range el.PageStatisticsByGeoCountry {
			name := countries[c.Geo]
			if name == "" {
				name = c.Geo[strings.LastIndex(c.Geo, ":")+1:]
			}
			byCountry[name] += c.PageStatistics.Views.AllPageViews.PageViews
		}

		for _, g := range el.PageStatisticsByGeo {
			regionID := g.Geo[strings.LastIndex(g.Geo, ":")+1:]
			byRegion[regionID] += g.PageStatistics.Views.AllPageViews.PageViews
		}

		for _, ind := range el.PageStatisticsByIndustryV2 {
			name := inds[ind.IndustryV2]
			if name == "" {
				name = ind.IndustryV2[strings.LastIndex(ind.IndustryV2, ":")+1:]
			}
			byIndustry[name] += ind.PageStatistics.Views.AllPageViews.PageViews
		}

		for _, s := range el.PageStatisticsBySeniority {
			name := seniors[s.Seniority]
			if name == "" {
				name = s.Seniority[strings.LastIndex(s.Seniority, ":")+1:]
			}
			bySeniority[name] += s.PageStatistics.Views.AllPageViews.PageViews
		}

		for _, f := range el.PageStatisticsByFunction {
			name := funcs[f.Function]
			if name == "" {
				name = f.Function[strings.LastIndex(f.Function, ":")+1:]
			}
			byFunction[name] += f.PageStatistics.Views.AllPageViews.PageViews
		}

		for _, sc := range el.PageStatisticsByStaffCountRange {
			byStaffCount[sc.StaffCountRange] += sc.PageStatistics.Views.AllPageViews.PageViews
		}
	}

	countryBytes, _ := json.Marshal(byCountry)
	ins.PageViewsByCountry = string(countryBytes)

	regionBytes, _ := json.Marshal(byRegion)
	ins.PageViewsByRegion = string(regionBytes)

	industryBytes, _ := json.Marshal(byIndustry)
	ins.PageViewsByIndustry = string(industryBytes)

	seniorityBytes, _ := json.Marshal(bySeniority)
	ins.PageViewsBySeniority = string(seniorityBytes)

	functionBytes, _ := json.Marshal(byFunction)
	ins.PageViewsByFunction = string(functionBytes)

	staffCountBytes, _ := json.Marshal(byStaffCount)
	ins.PageViewsByStaffCount = string(staffCountBytes)
}

//func StringToInt64(s string) int64 {
//	val, err := strconv.ParseInt(s, 10, 64)
//	if err != nil {
//		log.Error().Err(err).Str("string", s).Msg("string conversion failed")
//		return 0
//	}
//	return val
//}
