package parsing

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// InstagramParser converts raw Instagram Graph API payloads into strongly typed ParsedInstagramPost models
// used downstream by the ClickHouse sink. The logic is adapted from the legacy Python implementation
// (analytics/instagram_analytics_model.py) and mirrors the structure of FacebookParser already present
// in this codebase.
//
// NOTE: At this point the parser focuses on media (posts & stories). A separate function can be added
// later for account-level insights once the RawInstagramInsightsResponse is stored in Kafka.
type InstagramParser struct {
	hashtagRegex *regexp.Regexp
}

// NewInstagramParser returns a ready to use InstagramParser instance.
func NewInstagramParser() *InstagramParser {
	return &InstagramParser{
		hashtagRegex: regexp.MustCompile(`#(\w+)`),
	}
}

// ParseMediaWithInsights converts an enriched media record (with insights) into a ParsedInstagramPost.
func (ip *InstagramParser) ParseMediaWithInsights(
	enrichedData map[string]interface{},
	instagramID string,
) (*kafkamodels.ParsedInstagramPost, error) {
	// Extract base media data
	mediaData, _ := json.Marshal(enrichedData)
	var media kafkamodels.RawInstagramMedia
	if err := json.Unmarshal(mediaData, &media); err != nil {
		return nil, err
	}

	// Extract user info
	accountName := ""
	username := media.Username
	profilePictureURL := ""
	if userInfo, ok := enrichedData["user_info"].(map[string]interface{}); ok {
		if name, ok := userInfo["name"].(string); ok {
			accountName = name
		}
		if uname, ok := userInfo["username"].(string); ok {
			username = uname
		}
		if pic, ok := userInfo["profile_picture_url"].(string); ok {
			profilePictureURL = pic
		}
	}

	// Parse base media
	parsed := ip.ParseMedia(media, instagramID, accountName, username, profilePictureURL)
	if parsed == nil {
		return nil, nil
	}

	// Extract and apply media insights
	if insights, ok := enrichedData["insights"].(map[string]interface{}); ok {
		if data, ok := insights["data"].([]interface{}); ok {
			for _, metric := range data {
				if m, ok := metric.(map[string]interface{}); ok {
					nameVal, ok := m["name"]
					if !ok {
						continue
					}
					metricName, ok := nameVal.(string)
					if !ok || metricName == "" {
						continue
					}

					// Handle navigation metrics for stories
					if metricName == "navigation" {
						if totalValue, ok := m["total_value"].(map[string]interface{}); ok {
							if breakdowns, ok := totalValue["breakdowns"].([]interface{}); ok && len(breakdowns) > 0 {
								if breakdown, ok := breakdowns[0].(map[string]interface{}); ok {
									if results, ok := breakdown["results"].([]interface{}); ok {
										for _, result := range results {
											if r, ok := result.(map[string]interface{}); ok {
												if dims, ok := r["dimension_values"].([]interface{}); ok && len(dims) > 0 {
													dimension := dims[0].(string)
													value := ip.getInt64Value(r["value"])
													switch dimension {
													case "tap_back":
														parsed.TapsBack = value
													case "tap_forward":
														parsed.TapsForward = value
													case "tap_exit":
														parsed.Exits = value
													}
												}
											}
										}
									}
								}
							}
						}
						continue
					}

					// Handle regular metrics
					if values, ok := m["values"].([]interface{}); ok && len(values) > 0 {
						if v, ok := values[0].(map[string]interface{}); ok {
							value := ip.getInt64Value(v["value"])
							ip.applyMediaMetric(parsed, metricName, value)
						}
					}
				}
			}
		}
	}

	// Recalculate engagement with all metrics
	parsed.Engagement = parsed.LikeCount + parsed.CommentsCount + parsed.Saved

	return parsed, nil
}

// ParseMedia converts a RawInstagramMedia record into a ParsedInstagramPost model. Additional context
// such as account name / profile-picture is provided by the caller (work-order processor or higher-level
// service) using the parameters. The function never panics – on invalid input it simply returns nil, err.
func (ip *InstagramParser) ParseMedia(
	media kafkamodels.RawInstagramMedia,
	instagramID string,
	accountName string,
	username string,
	profilePictureURL string,
) *kafkamodels.ParsedInstagramPost {

	if media.ID == "" || media.Timestamp == "" {
		return nil
	}

	// Parse timestamp – Instagram Graph API returns RFC3339 date-time string, e.g. "2024-01-02T15:04:05+0000"
	// The Go RFC3339 parser expects timezone format with a colon, so we attempt both.
	var createdAt time.Time
	var parseErr error
	// first try RFC3339 with timezone offset without colon
	createdAt, parseErr = time.Parse("2006-01-02T15:04:05-0700", media.Timestamp)
	if parseErr != nil {
		createdAt, parseErr = time.Parse(time.RFC3339, media.Timestamp)
	}
	if parseErr != nil {
		// fallback to zero value; downstream may handle missing timestamp
		createdAt = time.Time{}
	}

	parsed := &kafkamodels.ParsedInstagramPost{
		InstagramID:       instagramID,
		MediaID:           media.ID,
		Username:          username,
		Name:              accountName,
		ProfilePictureURL: profilePictureURL,
		Permalink:         media.Permalink,
		LikeCount:         int64(media.LikeCount),
		CommentsCount:     int64(media.CommentsCount),
		EntityType:        strings.ToUpper(media.MediaProductType),
		Caption:           media.Caption,
		StoredEventAt:     time.Now().UTC(),
		PostCreatedAt:     createdAt,
	}

	// Set MediaType from API response (IMAGE, VIDEO, CAROUSEL_ALBUM, etc.)
	parsed.MediaType = strings.ToUpper(media.MediaType)
	// Override for REELS to maintain consistency
	if media.MediaProductType == "REELS" {
		parsed.MediaType = "REELS"
	}

	// Derive date-time partitions if timestamp available
	if !createdAt.IsZero() {
		parsed.Timestamp = createdAt.Unix() * 1000 // milliseconds like Python version
		parsed.DayOfWeek = createdAt.Weekday().String()
		parsed.HourOfDay = int64(createdAt.Hour())
		parsed.Year = int64(createdAt.Year())
		parsed.Month = int64(createdAt.Month())
	}

	// Child assets
	var childMediaTypes []string
	var mediaURLs []string
	var videoURLs []string

	// Helper: add URLs for parent media
	appendMedia := func() {
		if media.MediaType == "VIDEO" {
			videoURLs = append(videoURLs, media.MediaURL)
			if media.ThumbnailURL != "" {
				mediaURLs = append(mediaURLs, media.ThumbnailURL)
			}
		} else {
			mediaURLs = append(mediaURLs, media.MediaURL)
		}
	}

	if len(media.Children.Data) == 0 {
		appendMedia()
	}

	for _, child := range media.Children.Data {
		childMediaTypes = append(childMediaTypes, strings.ToUpper(child.MediaType))
		if child.MediaType == "VIDEO" {
			videoURLs = append(videoURLs, child.MediaURL)
			if child.ThumbnailURL != "" {
				mediaURLs = append(mediaURLs, child.ThumbnailURL)
			}
		} else {
			mediaURLs = append(mediaURLs, child.MediaURL)
		}
	}

	parsed.ChildAssetsType = childMediaTypes
	parsed.MediaURL = mediaURLs
	parsed.VideoURL = videoURLs

	// Hashtags extraction from caption
	if parsed.Caption != "" {
		tags := ip.hashtagRegex.FindAllStringSubmatch(parsed.Caption, -1)
		if len(tags) > 0 {
			unique := make(map[string]struct{})
			for _, match := range tags {
				tag := match[1]
				if _, ok := unique[tag]; !ok {
					unique[tag] = struct{}{}
					parsed.Hashtags = append(parsed.Hashtags, tag)
				}
			}
		}
	}

	// Engagement & Impressions (initially zero – those are fetched through insights later)
	parsed.Engagement = parsed.LikeCount + parsed.CommentsCount + parsed.Saved

	return parsed
}

// applyMediaMetric applies a metric value to the appropriate field in ParsedInstagramPost
func (ip *InstagramParser) applyMediaMetric(parsed *kafkamodels.ParsedInstagramPost, name string, value int64) {
	switch name {
	case "impressions":
		parsed.Impressions = value
	case "reach":
		parsed.Reach = value
	case "saved":
		parsed.Saved = value
	case "video_views", "views":
		parsed.VideoViews = value
		parsed.Views = value
	case "exits":
		parsed.Exits = value
	case "replies":
		parsed.Replies = value
	case "taps_forward":
		parsed.TapsForward = value
	case "taps_back":
		parsed.TapsBack = value
	case "shares":
		parsed.Shares = value
	case "ig_reels_avg_watch_time":
		parsed.ReelsAvgWatchTime = value
	case "ig_reels_video_view_total_time":
		parsed.ReelsTotalWatchTime = value
	case "total_interactions":
		parsed.Engagement = value
	}
}

// ParseInsightsWithDemographics parses enriched insights data (with demographics and user info).
func (ip *InstagramParser) ParseInsightsWithDemographics(
	enrichedData map[string]interface{},
	instagramID string,
	recordID string,
) (*kafkamodels.ParsedInstagramInsight, error) {
	now := time.Now().UTC()
	parsed := &kafkamodels.ParsedInstagramInsight{
		InstagramID:         instagramID,
		RecordID:            recordID,
		StoredEventAt:       now,
		AudienceDatetime:    now,
		OnlineUsersDatetime: now.Add(-48 * time.Hour), // Default: 2 days ago (Instagram API returns data for 2 days ago)
		CreatedTime:         now,
		UpdatedTime:         now,
		Metadata: map[string]string{
			"source": "live_fetch",
		},
	}

	// Extract user info
	if userInfo, ok := enrichedData["user_info"].(map[string]interface{}); ok {

		if name, ok := userInfo["name"].(string); ok {
			parsed.Name = name
		}
		if username, ok := userInfo["username"].(string); ok {
			parsed.Username = username
		}
		if pic, ok := userInfo["profile_picture_url"].(string); ok {
			parsed.ProfilePictureURL = pic
		}
		fc, ok := userInfo["followers_count"]
		if ok {
			parsed.FollowersCount = ip.getInt64Value(fc)
		}
		if fc, ok := userInfo["follows_count"]; ok {
			parsed.FollowsCount = ip.getInt64Value(fc)
		}
		if mc, ok := userInfo["media_count"]; ok {
			parsed.MediaCount = ip.getInt64Value(mc)
		}
	}

	// Parse regular insights
	if insights, ok := enrichedData["insights"].(map[string]interface{}); ok {
		if data, ok := insights["data"].([]interface{}); ok {
			for _, metric := range data {
				if m, ok := metric.(map[string]interface{}); ok {
					metricName, _ := m["name"].(string)

					// Handle daily metrics (time_series - has "values" array)
					if values, ok := m["values"].([]interface{}); ok && len(values) > 0 {
						if v, ok := values[0].(map[string]interface{}); ok {
							value := ip.getInt64Value(v["value"])
							ip.applyMetric(parsed, metricName, value)
						}
					}

					// Handle total value metrics (aggregated - has "total_value")
					if totalValue, ok := m["total_value"].(map[string]interface{}); ok {
						if val, ok := totalValue["value"]; ok {
							ip.applyMetric(parsed, metricName, ip.getInt64Value(val))
						}
					}
				}
			}
		}
	}

	// Parse demographics
	if demographics, ok := enrichedData["demographics"].(map[string]interface{}); ok {
		if data, ok := demographics["data"].([]interface{}); ok {
			for _, metric := range data {
				if m, ok := metric.(map[string]interface{}); ok {
					metricName, _ := m["name"].(string)

					// Handle online followers
					if metricName == "online_followers" {
						if totalValue, ok := m["total_value"].(map[string]interface{}); ok {
							if val, ok := totalValue["value"].(map[string]interface{}); ok {
								// Convert hourly online followers to array format
								onlineFollowers := make([]string, 0, 24)
								for hour := 0; hour < 24; hour++ {
									hourStr := fmt.Sprintf("%d", hour)
									count := 0
									if v, ok := val[hourStr]; ok {
										count = int(ip.getInt64Value(v))
									}
									onlineFollowers = append(onlineFollowers, fmt.Sprintf("%s:%d", hourStr, count))
								}
								parsed.OnlineFollowers = onlineFollowers

								// Set online users datetime (2 days ago)
								parsed.OnlineUsersDatetime = time.Now().UTC().Add(-48 * time.Hour)
							}
						}
						continue
					}

					// Handle demographic breakdowns
					if totalValue, ok := m["total_value"].(map[string]interface{}); ok {
						if breakdowns, ok := totalValue["breakdowns"].([]interface{}); ok && len(breakdowns) > 0 {
							if breakdown, ok := breakdowns[0].(map[string]interface{}); ok {
								if results, ok := breakdown["results"].([]interface{}); ok {
									ip.parseDemographicBreakdown(parsed, metricName, results)
								}
							}
						}
					}
				}
			}
		}
	}

	// Set date fields based on record date (extracted from recordID: instagramID_YYYY-MM-DD)
	recordDate := now
	if len(recordID) > 11 {
		// Try to extract date from recordID format: {instagram_id}_{YYYY-MM-DD}
		parts := strings.Split(recordID, "_")
		if len(parts) >= 2 {
			if t, err := time.Parse("2006-01-02", parts[len(parts)-1]); err == nil {
				recordDate = t
			}
		}
	}
	parsed.DayOfWeek = recordDate.Weekday().String()
	parsed.Year = int64(recordDate.Year())
	parsed.Month = int64(recordDate.Month())

	// Calculate engagement if not set
	if parsed.Engagement == 0 {
		parsed.Engagement = parsed.Likes + parsed.Comments + parsed.Saves
	}

	return parsed, nil
}

// parseDemographicBreakdown parses demographic breakdown data and applies to the appropriate fields
func (ip *InstagramParser) parseDemographicBreakdown(parsed *kafkamodels.ParsedInstagramInsight, metricName string, results []interface{}) {
	suffix := ""
	if strings.Contains(metricName, "engaged_audience") {
		suffix = "_by_engagement"
	} else if strings.Contains(metricName, "reached_audience") {
		suffix = "_by_reach"
	}

	// Determine breakdown type based on dimension values
	if len(results) == 0 {
		return
	}

	// Create formatted list of demographic data
	demographics := make([]string, 0, len(results))
	genderAges := make(map[string]int)

	for _, result := range results {
		if r, ok := result.(map[string]interface{}); ok {
			if dims, ok := r["dimension_values"].([]interface{}); ok {
				value := int(ip.getInt64Value(r["value"]))

				if len(dims) == 1 {
					// Single dimension (city, country, age, gender)
					dim := dims[0].(string)
					demographics = append(demographics, fmt.Sprintf("%s:%d", dim, value))
				} else if len(dims) == 2 {
					// Gender-age combination
					age := dims[0].(string)
					gender := dims[1].(string)
					key := fmt.Sprintf("%s.%s", gender, age)
					genderAges[key] = value
				}
			}
		}
	}

	// Apply to appropriate field based on first result's dimensions
	if len(results) > 0 {
		if r, ok := results[0].(map[string]interface{}); ok {
			if dims, ok := r["dimension_values"].([]interface{}); ok {
				if len(dims) == 1 {
					// Determine field type from value pattern
					sample := dims[0].(string)
					if strings.Contains(sample, "-") && len(sample) <= 7 {
						// Age range like "18-24"
						fieldName := "AudienceAge" + suffix
						ip.setDemographicField(parsed, fieldName, demographics)
					} else if sample == "M" || sample == "F" || sample == "U" {
						// Gender
						fieldName := "AudienceGender" + suffix
						ip.setDemographicField(parsed, fieldName, demographics)
					} else if len(sample) == 2 {
						// Country code
						fieldName := "AudienceCountry" + suffix
						ip.setDemographicField(parsed, fieldName, demographics)
					} else {
						// City or locale
						fieldName := "AudienceCity" + suffix
						ip.setDemographicField(parsed, fieldName, demographics)
					}
				} else if len(dims) == 2 {
					// Gender-age combination
					genderAgeList := make([]string, 0, len(genderAges))
					for k, v := range genderAges {
						genderAgeList = append(genderAgeList, fmt.Sprintf("%s:%d", k, v))
					}
					fieldName := "AudienceGenderAge" + suffix
					ip.setDemographicField(parsed, fieldName, genderAgeList)
				}
			}
		}
	}
}

// setDemographicField sets the appropriate demographic field based on field name
func (ip *InstagramParser) setDemographicField(parsed *kafkamodels.ParsedInstagramInsight, fieldName string, values []string) {
	switch fieldName {
	case "AudienceAge":
		parsed.AudienceAge = values
	case "AudienceGender":
		parsed.AudienceGender = values
	case "AudienceGenderAge":
		parsed.AudienceGenderAge = values
	case "AudienceCity":
		parsed.AudienceCity = values
	case "AudienceCountry":
		parsed.AudienceCountry = values
	case "AudienceLocale":
		parsed.AudienceLocale = values
	case "AudienceAge_by_engagement":
		parsed.AudienceAgeByEngagement = values
	case "AudienceGender_by_engagement":
		parsed.AudienceGenderByEngagement = values
	case "AudienceGenderAge_by_engagement":
		parsed.AudienceGenderAgeByEngagement = values
	case "AudienceCity_by_engagement":
		parsed.AudienceCityByEngagement = values
	case "AudienceCountry_by_engagement":
		parsed.AudienceCountryByEngagement = values
	case "AudienceAge_by_reach":
		parsed.AudienceAgeByReach = values
	case "AudienceGender_by_reach":
		parsed.AudienceGenderByReach = values
	case "AudienceGenderAge_by_reach":
		parsed.AudienceGenderAgeByReach = values
	case "AudienceCity_by_reach":
		parsed.AudienceCityByReach = values
	case "AudienceCountry_by_reach":
		parsed.AudienceCountryByReach = values
	}
}

// ParseInsightsDaily converts enriched insights data into multiple ParsedInstagramInsight records (one per day).
// This follows the Facebook pattern where:
// - Time series metrics (reach): Have daily values in "values" array - we extract value for each date
// - Total value metrics (likes, comments, etc.): Have ONE aggregated value - we apply to all dates
func (ip *InstagramParser) ParseInsightsDaily(
	enrichedData map[string]interface{},
	instagramID string,
) ([]*kafkamodels.ParsedInstagramInsight, error) {
	// First, collect all unique dates from time series metrics (reach has "values" array)
	dateSet := make(map[string]time.Time)

	if insights, ok := enrichedData["insights"].(map[string]interface{}); ok {
		if data, ok := insights["data"].([]interface{}); ok {
			for _, metric := range data {
				if m, ok := metric.(map[string]interface{}); ok {
					// Check if this metric has a "values" array (time series)
					if values, ok := m["values"].([]interface{}); ok {
						for _, val := range values {
							if v, ok := val.(map[string]interface{}); ok {
								endTime, _ := v["end_time"].(string)
								if endTime != "" {
									t, err := ip.parseEndTime(endTime)
									if err != nil {
										continue
									}
									dateStr := t.Format("2006-01-02")
									dateSet[dateStr] = t
								}
							}
						}
					}
				}
			}
		}
	}

	// Extract user info once (shared across all days)
	var name, username, profilePic string
	var followersCount, followsCount, mediaCount int64
	if userInfo, ok := enrichedData["user_info"].(map[string]interface{}); ok {
		if n, ok := userInfo["name"].(string); ok {
			name = n
		}
		if u, ok := userInfo["username"].(string); ok {
			username = u
		}
		if p, ok := userInfo["profile_picture_url"].(string); ok {
			profilePic = p
		}
		if fc, ok := userInfo["followers_count"]; ok {
			followersCount = ip.getInt64Value(fc)
		}
		if fc, ok := userInfo["follows_count"]; ok {
			followsCount = ip.getInt64Value(fc)
		}
		if mc, ok := userInfo["media_count"]; ok {
			mediaCount = ip.getInt64Value(mc)
		}
	}

	// Collect total_value metrics (these are aggregated, applied to all dates)
	// IMPORTANT: Skip metrics that have "values" array (time series) - they're handled separately per date
	totalValueMetrics := make(map[string]int64)
	if insights, ok := enrichedData["insights"].(map[string]interface{}); ok {
		if data, ok := insights["data"].([]interface{}); ok {
			for _, metric := range data {
				if m, ok := metric.(map[string]interface{}); ok {
					metricName, _ := m["name"].(string)
					// Skip if this metric has a "values" array (time series metric)
					if _, hasValues := m["values"].([]interface{}); hasValues {
						continue
					}
					if totalValue, ok := m["total_value"].(map[string]interface{}); ok {
						if val, ok := totalValue["value"]; ok {
							totalValueMetrics[metricName] = ip.getInt64Value(val)
						}
					}
				}
			}
		}
	}

	// If no dates found from time series AND no total_value metrics, skip processing
	// This prevents creating empty records that overwrite good data
	if len(dateSet) == 0 && len(totalValueMetrics) == 0 {
		return nil, nil
	}

	// If no dates found from time series but we have total_value metrics, use today
	if len(dateSet) == 0 {
		now := time.Now().UTC()
		dateSet[now.Format("2006-01-02")] = now
	}

	// Create a parsed insight record for each date
	now := time.Now().UTC()
	var results []*kafkamodels.ParsedInstagramInsight

	for dateStr, endTime := range dateSet {
		recordID := fmt.Sprintf("%s_%s", instagramID, dateStr)

		parsed := &kafkamodels.ParsedInstagramInsight{
			InstagramID:         instagramID,
			RecordID:            recordID,
			Username:            username,
			Name:                name,
			ProfilePictureURL:   profilePic,
			FollowersCount:      followersCount,
			FollowsCount:        followsCount,
			MediaCount:          mediaCount,
			Year:                int64(endTime.Year()),
			Month:               int64(endTime.Month()),
			DayOfWeek:           endTime.Weekday().String(),
			CreatedTime:         endTime, // The actual date the data belongs to (from API end_time)
			StoredEventAt:       now,     // When the record was saved (today's date)
			AudienceDatetime:    now,
			OnlineUsersDatetime: now.Add(-48 * time.Hour),
		}

		// Apply time series metrics (reach) for this specific date
		if insights, ok := enrichedData["insights"].(map[string]interface{}); ok {
			if data, ok := insights["data"].([]interface{}); ok {
				for _, metric := range data {
					if m, ok := metric.(map[string]interface{}); ok {
						metricName, _ := m["name"].(string)
						if values, ok := m["values"].([]interface{}); ok {
							value := ip.getValueForDate(values, endTime)
							if value > 0 {
								ip.applyMetric(parsed, metricName, value)
							}
						}
					}
				}
			}
		}

		// Apply total_value metrics (same for all dates)
		for metricName, value := range totalValueMetrics {
			ip.applyMetric(parsed, metricName, value)
		}

		// Parse demographics (same for all dates)
		ip.parseDemographics(parsed, enrichedData)

		// Calculate engagement if not set
		if parsed.Engagement == 0 {
			parsed.Engagement = parsed.Likes + parsed.Comments + parsed.Saves
		}

		results = append(results, parsed)
	}

	return results, nil
}

// parseEndTime parses Instagram's end_time format
func (ip *InstagramParser) parseEndTime(endTime string) (time.Time, error) {
	// Try multiple formats
	formats := []string{
		"2006-01-02T15:04:05-0700",
		"2006-01-02T15:04:05+0000",
		time.RFC3339,
	}
	for _, format := range formats {
		if t, err := time.Parse(format, endTime); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("InstagramParser.parseEndTime: unable to parse end_time: %s", endTime)
}

// getValueForDate finds the insight value for a specific date from values array
func (ip *InstagramParser) getValueForDate(values []interface{}, targetDate time.Time) int64 {
	targetDateStr := targetDate.Format("2006-01-02")
	for _, val := range values {
		if v, ok := val.(map[string]interface{}); ok {
			if endTime, ok := v["end_time"].(string); ok {
				t, err := ip.parseEndTime(endTime)
				if err != nil {
					continue
				}
				if t.Format("2006-01-02") == targetDateStr {
					if value, ok := v["value"]; ok {
						return ip.getInt64Value(value)
					}
				}
			}
		}
	}
	return 0
}

// parseDemographics extracts demographic data and applies to parsed insight
func (ip *InstagramParser) parseDemographics(parsed *kafkamodels.ParsedInstagramInsight, enrichedData map[string]interface{}) {
	if demographics, ok := enrichedData["demographics"].(map[string]interface{}); ok {
		if data, ok := demographics["data"].([]interface{}); ok {
			for _, metric := range data {
				if m, ok := metric.(map[string]interface{}); ok {
					metricName, _ := m["name"].(string)

					// Handle online followers
					if metricName == "online_followers" {
						if totalValue, ok := m["total_value"].(map[string]interface{}); ok {
							if val, ok := totalValue["value"].(map[string]interface{}); ok {
								onlineFollowers := make([]string, 0, 24)
								for hour := 0; hour < 24; hour++ {
									hourStr := fmt.Sprintf("%d", hour)
									if count, ok := val[hourStr]; ok {
										onlineFollowers = append(onlineFollowers, fmt.Sprintf("%s:%v", hourStr, count))
									}
								}
								parsed.OnlineFollowers = onlineFollowers
							}
						}
						continue
					}

					// Handle breakdown demographics
					if totalValue, ok := m["total_value"].(map[string]interface{}); ok {
						if breakdowns, ok := totalValue["breakdowns"].([]interface{}); ok {
							for _, bd := range breakdowns {
								if breakdown, ok := bd.(map[string]interface{}); ok {
									if results, ok := breakdown["results"].([]interface{}); ok {
										var demographicList []string
										for _, r := range results {
											if result, ok := r.(map[string]interface{}); ok {
												if dimVals, ok := result["dimension_values"].([]interface{}); ok && len(dimVals) > 0 {
													key := fmt.Sprintf("%v", dimVals[0])
													if val, ok := result["value"]; ok {
														demographicList = append(demographicList, fmt.Sprintf("%s:%v", key, val))
													}
												}
											}
										}
										// Map metric name to field
										switch metricName {
										case "engaged_audience_demographics":
											if strings.Contains(strings.Join(demographicList, ""), "M:") || strings.Contains(strings.Join(demographicList, ""), "F:") {
												parsed.AudienceGenderByEngagement = demographicList
											} else {
												parsed.AudienceAgeByEngagement = demographicList
											}
										case "reached_audience_demographics":
											if strings.Contains(strings.Join(demographicList, ""), "M:") || strings.Contains(strings.Join(demographicList, ""), "F:") {
												parsed.AudienceGenderByReach = demographicList
											} else {
												parsed.AudienceAgeByReach = demographicList
											}
										case "follower_demographics":
											if strings.Contains(strings.Join(demographicList, ""), "M:") || strings.Contains(strings.Join(demographicList, ""), "F:") {
												parsed.AudienceGender = demographicList
											} else {
												parsed.AudienceAge = demographicList
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

// ParseInsights converts a RawInstagramInsightsResponse into ParsedInstagramInsight.
// Only commonly used metrics are mapped initially – extend as new requirements emerge.
func (ip *InstagramParser) ParseInsights(raw *kafkamodels.RawInstagramInsightsResponse, instagramID, username, name, profilePic string, recordID string) (*kafkamodels.ParsedInstagramInsight, error) {
	if raw == nil || len(raw.Data) == 0 {
		return nil, nil
	}

	now := time.Now().UTC()
	parsed := &kafkamodels.ParsedInstagramInsight{
		InstagramID:         instagramID,
		RecordID:            recordID,
		Username:            username,
		Name:                name,
		ProfilePictureURL:   profilePic,
		StoredEventAt:       now,
		AudienceDatetime:    now,
		OnlineUsersDatetime: now.Add(-48 * time.Hour), // Default: 2 days ago
	}

	// default day/year/month fields to today – can be overwritten when metric has date info
	parsed.DayOfWeek = now.Weekday().String()
	parsed.Year = int64(now.Year())
	parsed.Month = int64(now.Month())
	for _, metric := range raw.Data {

		// daily metric – first element usually has EndTime date string
		if len(metric.Values) > 0 {
			v := metric.Values[0]
			ip.applyMetric(parsed, metric.Name, v.Value)
		}
		// total_value style metrics
		if metric.TotalValue.Value != 0 {
			ip.applyMetric(parsed, metric.Name, metric.TotalValue.Value)
		}
	}

	// Engagement derived as likes + comments + saves if not provided
	if parsed.Engagement == 0 {
		parsed.Engagement = parsed.Likes + parsed.Comments + parsed.Saves
	}

	return parsed, nil
}

func (ip *InstagramParser) applyMetric(parsed *kafkamodels.ParsedInstagramInsight, name string, val interface{}) {
	iv := ip.getInt64Value(val)
	switch name {
	case "follows":
		parsed.FollowsCount = iv
	case "follower_count":
		parsed.FollowerCount = iv
	case "media_count":
		parsed.MediaCount = iv
	case "tags":
		parsed.Tags = iv
	case "impressions":
		parsed.Impressions = iv
	case "profile_views":
		parsed.ProfileViews = iv
	case "shares":
		parsed.Shares = iv
	case "accounts_engaged":
		parsed.AccountsEngaged = iv
	case "reach":
		parsed.Reach = iv
	case "views":
		parsed.Views = iv
	case "saves":
		parsed.Saves = iv
	case "likes":
		parsed.Likes = iv
	case "comments":
		parsed.Comments = iv
	default:
		// ignore other metrics for now
	}
}

// getInt64Value extracts int64 from interface{}
func (ip *InstagramParser) getInt64Value(value interface{}) int64 {
	switch v := value.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	case json.Number:
		i, _ := v.Int64()
		return i
	default:
		return 0
	}
}
