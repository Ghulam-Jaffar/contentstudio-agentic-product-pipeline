// Package overview provides the ClickHouse repository layer for cross-platform Overview analytics.
// Queries are migrated from PHP OverviewV2Builder (contentstudio-backend).
// Key migration notes:
//   - PHP addDay() on end date is replicated in NewOverviewParams (currentEnd = rawEnd + 1 day)
//   - PHP DateFilter uses toDateTime(field) with no timezone; Go helpers.DateFilter adds timezone,
//     so a local dateFilter/secondaryDateFilter is used here instead.
//   - Secondary period: if full month → previous calendar month; otherwise → N days before start
//   - Empty account ID arrays produce (") which returns no rows (not an error)
package overview

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	ch "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
)

// OverviewParams holds pre-computed query parameters for all overview SQL queries.
// Mirrors PHP OverviewV2Builder constructor: end date gets +1 day, secondary period
// is computed using full-month detection and Carbon date arithmetic equivalents.
type OverviewParams struct {
	FacebookIDs  string
	InstagramIDs string
	LinkedInIDs  string
	TiktokIDs    string
	PinterestIDs string
	YouTubeIDs   string
	AllAccounts  string

	CurrentStart   string // "YYYY-MM-DD" raw start
	CurrentEnd     string // "YYYY-MM-DD" raw end + 1 day (PHP addDay())
	SecondaryStart string // "YYYY-MM-DD"
	SecondaryEnd   string // "YYYY-MM-DD"

	IncludeFacebook  bool
	IncludeInstagram bool
	IncludeLinkedIn  bool
	IncludeTiktok    bool
	IncludePinterest bool
	IncludeYouTube   bool

	Timezone string
	Type     string // ORDER BY column for getTopPostsQuery
	Limit    int
}

// Repository executes ClickHouse queries for Overview analytics.
type Repository struct {
	client *ch.Client
}

// NewRepository returns a new Repository backed by the given ClickHouse client.
func NewRepository(client *ch.Client) *Repository {
	return &Repository{client: client}
}

// NewOverviewParams parses the date range, applies PHP OverviewV2Builder date logic,
// and formats all account ID slices as SQL IN-clause strings.
func NewOverviewParams(
	startDateStr, endDateStr string,
	facebookIDs, instagramIDs, linkedInIDs, tiktokIDs, pinterestIDs, youtubeIDs []string,
	timezone string,
	sortType string,
	limit int,
) (*OverviewParams, error) {
	startDate, err := time.Parse("2006-01-02", strings.TrimSpace(startDateStr))
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}
	endDate, err := time.Parse("2006-01-02", strings.TrimSpace(endDateStr))
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	// PHP: $this->currentEndDate = Carbon::parse($date_array[1])->addDay()
	currentEnd := endDate.AddDate(0, 0, 1)

	// PHP: $isFullMonth = $currentStart->day === 1 && $currentEnd->day === $currentEnd->daysInMonth
	daysInEndMonth := time.Date(endDate.Year(), endDate.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
	isFullMonth := startDate.Day() == 1 && endDate.Day() == daysInEndMonth

	var secStart, secEnd time.Time
	if isFullMonth {
		// PHP: subMonthNoOverflow()->startOfMonth() and ->endOfMonth()->addDay()
		prevMonthStart := time.Date(startDate.Year(), startDate.Month()-1, 1, 0, 0, 0, 0, time.UTC)
		prevMonthEnd := time.Date(startDate.Year(), startDate.Month(), 0, 0, 0, 0, 0, time.UTC)
		secStart = prevMonthStart
		secEnd = prevMonthEnd.AddDate(0, 0, 1)
	} else {
		// PHP: $numberOfDays = date_diff($this->currentStartDate, $this->currentEndDate)->days
		// currentEndDate already has +1 day, so numberOfDays = range length in days
		numberOfDays := int(currentEnd.Sub(startDate).Hours() / 24)
		secEnd = startDate // PHP: $secondaryEndDate = $this->currentStartDate (no addDay)
		secStart = startDate.AddDate(0, 0, -numberOfDays)
	}

	if sortType == "" || sortType == "overview" {
		sortType = "total_engagement"
	}
	if limit <= 0 {
		limit = 20
	}
	if strings.EqualFold(timezone, "Europe/Kyiv") {
		timezone = "Europe/Riga"
	}
	if timezone == "" {
		timezone = "UTC"
	}

	allIDs := make([]string, 0, len(facebookIDs)+len(instagramIDs)+len(linkedInIDs)+len(tiktokIDs)+len(pinterestIDs)+len(youtubeIDs))
	allIDs = append(allIDs, facebookIDs...)
	allIDs = append(allIDs, instagramIDs...)
	allIDs = append(allIDs, linkedInIDs...)
	allIDs = append(allIDs, tiktokIDs...)
	allIDs = append(allIDs, pinterestIDs...)
	allIDs = append(allIDs, youtubeIDs...)

	return &OverviewParams{
		FacebookIDs:      ch.FormatAccountIDs(facebookIDs),
		InstagramIDs:     ch.FormatAccountIDs(instagramIDs),
		LinkedInIDs:      ch.FormatAccountIDs(linkedInIDs),
		TiktokIDs:        ch.FormatAccountIDs(tiktokIDs),
		PinterestIDs:     ch.FormatAccountIDs(pinterestIDs),
		YouTubeIDs:       ch.FormatAccountIDs(youtubeIDs),
		AllAccounts:      ch.FormatAccountIDs(allIDs),
		CurrentStart:     startDate.Format("2006-01-02"),
		CurrentEnd:       currentEnd.Format("2006-01-02"),
		SecondaryStart:   secStart.Format("2006-01-02"),
		SecondaryEnd:     secEnd.Format("2006-01-02"),
		IncludeFacebook:  len(facebookIDs) > 0,
		IncludeInstagram: len(instagramIDs) > 0,
		IncludeLinkedIn:  len(linkedInIDs) > 0,
		IncludeTiktok:    len(tiktokIDs) > 0,
		IncludePinterest: len(pinterestIDs) > 0,
		IncludeYouTube:   len(youtubeIDs) > 0,
		Timezone:         timezone,
		Type:             sortType,
		Limit:            limit,
	}, nil
}

// dateFilter matches PHP OverviewV2Builder.DateFilter(): toDateTime(field) BETWEEN start AND end.
// No timezone is applied to the field (unlike ch.DateFilter which adds timezone).
// End date already has +1 day baked into p.CurrentEnd.
func (p *OverviewParams) dateFilter(field string) string {
	return fmt.Sprintf(
		"%s >= toDateTime('%s',0) AND %s < toDateTime('%s',0)",
		field, p.CurrentStart, field, p.CurrentEnd,
	)
}

func (p *OverviewParams) secondaryDateFilter(field string) string {
	return fmt.Sprintf(
		"%s >= toDateTime('%s',0) AND %s < toDateTime('%s',0)",
		field, p.SecondaryStart, field, p.SecondaryEnd,
	)
}

// NewSecondaryParams returns a copy of p with the secondary period promoted to the current period.
func (p *OverviewParams) NewSecondaryParams() *OverviewParams {
	sec := *p
	sec.CurrentStart = p.SecondaryStart
	sec.CurrentEnd = p.SecondaryEnd
	return &sec
}

// buildFollowersSections mirrors PHP buildFollowersSections().
// Conditionally builds per-platform followers sub-queries UNION ALL'd together.
// When groupByAccount=true, includes the account_id column (for getAccountDataQuery).
func (p *OverviewParams) buildFollowersSections(groupByAccount bool) string {
	var sections []string

	accountCol := func(idCol string) string {
		if groupByAccount {
			return idCol + " as account_id,"
		}
		return ""
	}

	if p.IncludeFacebook {
		sections = append(sections, fmt.Sprintf(`SELECT
			toInt32(last_value(page_fans)) AS followers_count,
			%s
			'facebook' as platform_type
		FROM
			(SELECT page_fans, page_id, saving_time FROM facebook_insights
			WHERE page_id IN %s AND %s
			ORDER BY saving_time ASC
			)
		GROUP BY page_id`,
			accountCol("page_id"), p.FacebookIDs, p.dateFilter("saving_time"),
		))
	}

	if p.IncludeInstagram {
		igFilter := p.dateFilter("stored_event_at")
		sections = append(sections, fmt.Sprintf(`SELECT
			toInt32(if(count_in_range > 0, followers_in_range, followers_latest)) AS followers_count,
			%s
			'instagram' AS platform_type
		FROM (
			SELECT
				instagram_id,
				countIf(%s) AS count_in_range,
				argMaxIf(followers_count, stored_event_at, %s) AS followers_in_range,
				argMax(followers_count, stored_event_at) AS followers_latest
			FROM instagram_insights
			WHERE instagram_id IN %s
			GROUP BY instagram_id
		)`,
			accountCol("instagram_id"),
			igFilter, igFilter,
			p.InstagramIDs,
		))
	}

	if p.IncludeLinkedIn {
		sections = append(sections, fmt.Sprintf(`SELECT
			toInt32(argMax(totalFollowerCount, inserted_at)) AS followers_count,
			%s
			'linkedin' AS platform_type
		FROM linkedin_insights
		WHERE linkedin_id IN %s AND %s
		GROUP BY linkedin_id`,
			accountCol("linkedin_id"), p.LinkedInIDs, p.dateFilter("inserted_at"),
		))
	}

	if p.IncludeTiktok {
		sections = append(sections, fmt.Sprintf(`SELECT
			toInt32(argMax(total_follower_count, inserted_at)) AS followers_count,
			%s
			'tiktok' AS platform_type
		FROM tiktok_insights
		WHERE tiktok_id IN %s AND %s
		GROUP BY tiktok_id`,
			accountCol("tiktok_id"), p.TiktokIDs, p.dateFilter("inserted_at"),
		))
	}

	if p.IncludePinterest {
		sections = append(sections, fmt.Sprintf(`SELECT
			toInt32(argMax(follower_count, inserted_at)) AS followers_count,
			%s
			'pinterest' AS platform_type
		FROM pinterest_boards
		WHERE board_id IN %s AND %s
		GROUP BY board_id`,
			accountCol("board_id"), p.PinterestIDs, p.dateFilter("inserted_at"),
		))
	}

	if p.IncludeYouTube {
		sections = append(sections, fmt.Sprintf(`SELECT
			toInt32(if(count_in_range > 0, followers_in_range, followers_latest)) AS followers_count,
			%s
			'youtube' as platform_type
		FROM
			(SELECT
				channel_id,
				countIf(%s) as count_in_range,
				argMaxIf(subscriber_count, created_at, %s) as followers_in_range,
				argMax(subscriber_count, created_at) as followers_latest
			FROM youtube_channels
			WHERE channel_id IN %s
			GROUP BY channel_id)`,
			accountCol("channel_id"),
			p.dateFilter("created_at"), p.dateFilter("created_at"),
			p.YouTubeIDs,
		))
	}

	return strings.Join(sections, "\nUNION ALL\n")
}

// buildCurrentDataQuery mirrors PHP getCurrentDataQuery().
// Returns per-account current-period followers + posts data as a CTE body.
func (p *OverviewParams) buildCurrentDataQuery() string {
	return fmt.Sprintf(`WITH fb_posts AS (
		SELECT post_id, max(saving_time)
		FROM facebook_posts
		WHERE page_id IN %s AND %s
		GROUP BY post_id
	),
	followers_current AS (
		SELECT argMax(fi.page_follows, fi.saving_time) AS followers_count,
			fi.page_id AS account_id, 'facebook' AS platform_type,
			any(fp.page_name) AS account_name
		FROM facebook_insights AS fi
		LEFT JOIN facebook_posts AS fp ON fp.page_id = fi.page_id
		WHERE fi.page_id IN %s
		AND toDateTime(fi.saving_time) BETWEEN toDateTime('%s',0) AND toDateTime('%s',0)
		GROUP BY fi.page_id

		UNION ALL

		SELECT argMax(followers_count, stored_event_at) AS followers_count,
			instagram_id AS account_id, 'instagram' AS platform_type,
			argMax(name, stored_event_at) AS account_name
		FROM instagram_insights
		WHERE instagram_id IN %s AND %s
		GROUP BY instagram_id

		UNION ALL

		SELECT argMax(totalFollowerCount, inserted_at) AS followers_count,
			linkedin_id AS account_id, 'linkedin' AS platform_type,
			argMax(organization_name, inserted_at) AS account_name
		FROM linkedin_insights
		WHERE linkedin_id IN %s AND %s
		GROUP BY linkedin_id

		UNION ALL

		SELECT argMax(total_follower_count, inserted_at) AS followers_count,
			tiktok_id AS account_id, 'tiktok' AS platform_type,
			argMax(display_name, inserted_at) AS account_name
		FROM tiktok_insights
		WHERE tiktok_id IN %s AND %s
		GROUP BY tiktok_id

		UNION ALL

		SELECT argMax(follower_count, inserted_at) AS followers_count,
			board_id AS account_id, 'pinterest' AS platform_type,
			argMax(name, inserted_at) AS account_name
		FROM pinterest_boards
		WHERE board_id IN %s AND %s
		GROUP BY board_id

		UNION ALL

		SELECT if(count_in_range > 0, followers_in_range, followers_latest) AS followers_count,
			channel_id AS account_id, 'youtube' AS platform_type,
			account_name
		FROM (
			SELECT channel_id,
				countIf(%s) AS count_in_range,
				argMaxIf(subscriber_count, created_at, %s) AS followers_in_range,
				argMax(subscriber_count, created_at) AS followers_latest,
				argMax(title, created_at) AS account_name
			FROM youtube_channels
			WHERE channel_id IN %s
			GROUP BY channel_id
		)
	),
	posts_current AS (
		SELECT
			toInt32(fb_posts_agg.total_posts) AS total_posts,
			toInt32(fb_posts_agg.total_engagements) AS total_engagements,
			toInt32(coalesce(fb_insights_agg.total_impression, 0)) AS total_impression,
			toInt32(coalesce(fb_insights_agg.total_reach, 0)) AS total_reach,
			'facebook' AS platform_type, fb_posts_agg.page_id AS account_id
		FROM (
			SELECT count() AS total_posts, sum(total_engagement) AS total_engagements, page_id
			FROM (
				SELECT argMax(total, saving_time) + argMax(comments, saving_time) + argMax(shares, saving_time) + argMax(post_clicks, saving_time) AS total_engagement, page_id
				FROM facebook_posts WHERE (post_id, saving_time) IN fb_posts
				GROUP BY post_id, page_id
			)
			GROUP BY page_id
		) AS fb_posts_agg
		LEFT JOIN (
			SELECT page_id, sum(page_impressions) AS total_impression, sum(page_impressions_unique) AS total_reach
			FROM (
				SELECT page_id, max(page_impressions) AS page_impressions, max(page_impressions_unique) AS page_impressions_unique
				FROM facebook_insights WHERE page_id IN %s AND %s
				GROUP BY page_id, toDate(created_time)
			)
			GROUP BY page_id
		) AS fb_insights_agg ON fb_posts_agg.page_id = fb_insights_agg.page_id

		UNION ALL

		SELECT toInt32(COUNT()) AS total_posts, toInt32(SUM(engagement)) AS total_engagements,
			toInt32(SUM(impressions)) AS total_impression, toInt32(SUM(reach)) AS total_reach,
			'instagram' AS platform_type, instagram_id AS account_id
		FROM (
			SELECT argMax(engagement, stored_event_at) AS engagement, argMax(views, stored_event_at) AS impressions,
				argMax(reach, stored_event_at) AS reach, instagram_id
			FROM instagram_posts
			WHERE instagram_id IN %s AND %s
			GROUP BY instagram_id, media_id
		)
		GROUP BY instagram_id

		UNION ALL

		SELECT toInt32(COUNT()) AS total_posts, toInt32(SUM(total_engagement)) AS total_engagements,
			toInt32(SUM(total_impressions)) AS total_impression, toInt32(SUM(total_reach)) AS total_reach,
			'linkedin' AS platform_type, linkedin_id AS account_id
		FROM (
			SELECT argMax(favorites, saving_time) + argMax(comments, saving_time) + argMax(repost, saving_time) AS total_engagement,
				argMax(impressions, saving_time) AS total_impressions, argMax(reach, saving_time) AS total_reach,
				linkedin_id
			FROM linkedin_posts WHERE linkedin_id IN %s AND %s
			GROUP BY linkedin_id, post_id
		)
		GROUP BY linkedin_id

		UNION ALL

		SELECT toInt32(COUNT()) AS total_posts, toInt32(SUM(total_engagement)) AS total_engagements,
			toInt32(SUM(view_count)) AS total_impression, toInt32(SUM(view_count)) AS total_reach,
			'tiktok' AS platform_type, tiktok_id AS account_id
		FROM (
			SELECT argMax(like_count, inserted_at) + argMax(comments_count, inserted_at) + argMax(share_count, inserted_at) AS total_engagement,
				argMax(view_count, inserted_at) AS view_count, tiktok_id
			FROM tiktok_posts WHERE tiktok_id IN %s AND %s
			GROUP BY tiktok_id, post_id
		)
		GROUP BY tiktok_id

		UNION ALL

		SELECT toInt32(COUNT()) AS total_posts, toInt32(SUM(total_engagement)) AS total_engagements,
			toInt32(SUM(views)) AS total_impression, 0 AS total_reach, -- reach metric not available for YouTube
			'youtube' AS platform_type, channel_id AS account_id
		FROM (
			SELECT argMax(likes, inserted_at) + argMax(comments, inserted_at) + argMax(shares, inserted_at) + argMax(dislikes, inserted_at) AS total_engagement,
				argMax(views, inserted_at) AS views, channel_id
			FROM youtube_videos WHERE channel_id IN %s AND %s
			GROUP BY channel_id, video_id
		)
		GROUP BY channel_id

		UNION ALL

		WITH pins AS (
			SELECT pin_id, board_id FROM pinterest_pins
			WHERE board_id IN %s AND %s
			GROUP BY pin_id, board_id
		)
		SELECT toInt32(COUNT(DISTINCT pins.pin_id)) AS total_posts, toInt32(SUM(coalesce(pinterest_insights.engagement, 0))) AS total_engagements,
			toInt32(SUM(coalesce(pinterest_insights.impression, 0))) AS total_impression, 0 AS total_reach, -- reach metric not available for Pinterest
			'pinterest' AS platform_type, pins.board_id AS account_id
		FROM pins
		LEFT JOIN (SELECT pin_id, engagement, impression FROM pinterest_pin_insights WHERE pin_id IN (SELECT pin_id FROM pins)) AS pinterest_insights
			ON pins.pin_id = pinterest_insights.pin_id
		GROUP BY board_id
	)
	SELECT
		toInt32(COALESCE(followers_current.followers_count, 0)) AS followers,
		toInt32(COALESCE(posts_current.total_posts, 0)) AS total_posts,
		toInt32(COALESCE(posts_current.total_engagements, 0)) AS engagement,
		toInt32(COALESCE(posts_current.total_impression, 0)) AS impressions,
		toInt32(COALESCE(posts_current.total_reach, 0)) AS reach,
		followers_current.platform_type AS platform_type,
		followers_current.account_id AS account_id,
		followers_current.account_name AS account_name
	FROM followers_current
	LEFT JOIN posts_current
		ON followers_current.account_id = posts_current.account_id
		AND followers_current.platform_type = posts_current.platform_type`,
		p.FacebookIDs, p.dateFilter("created_time"),
		p.FacebookIDs, p.CurrentStart, p.CurrentEnd,
		p.InstagramIDs, p.dateFilter("stored_event_at"),
		p.LinkedInIDs, p.dateFilter("inserted_at"),
		p.TiktokIDs, p.dateFilter("inserted_at"),
		p.PinterestIDs, p.dateFilter("inserted_at"),
		p.dateFilter("created_at"), p.dateFilter("created_at"), p.YouTubeIDs,
		p.FacebookIDs, p.dateFilter("created_time"), // facebook_insights for impressions/reach
		p.InstagramIDs, p.dateFilter("post_created_at"),
		p.LinkedInIDs, p.dateFilter("created_at"),
		p.TiktokIDs, p.dateFilter("created_at"),
		p.YouTubeIDs, p.dateFilter("published_at"),
		p.PinterestIDs, p.dateFilter("created_at"),
	)
}

// buildOldDataQuery mirrors PHP getOldDataQuery().
// Note: fb_posts CTE intentionally uses current period (matching PHP behavior).
func (p *OverviewParams) buildOldDataQuery() string {
	return fmt.Sprintf(`WITH fb_posts AS (
		SELECT post_id, max(saving_time)
		FROM facebook_posts
		WHERE page_id IN %s AND %s
		GROUP BY post_id
	),
	followers_old AS (
		SELECT argMax(page_fans, saving_time) AS followers_count, page_id AS account_id, 'facebook' AS platform_type
		FROM facebook_insights
		WHERE page_id IN %s AND %s
		GROUP BY page_id

		UNION ALL

		SELECT argMax(followers_count, stored_event_at) AS followers_count, instagram_id AS account_id, 'instagram' AS platform_type
		FROM instagram_insights
		WHERE instagram_id IN %s AND %s
		GROUP BY instagram_id

		UNION ALL

		SELECT argMax(totalFollowerCount, inserted_at) AS followers_count, linkedin_id AS account_id, 'linkedin' AS platform_type
		FROM linkedin_insights
		WHERE linkedin_id IN %s AND %s
		GROUP BY linkedin_id

		UNION ALL

		SELECT argMax(total_follower_count, inserted_at) AS followers_count, tiktok_id AS account_id, 'tiktok' AS platform_type
		FROM tiktok_insights
		WHERE tiktok_id IN %s AND %s
		GROUP BY tiktok_id

		UNION ALL

		SELECT argMax(follower_count, inserted_at) AS followers_count, board_id AS account_id, 'pinterest' AS platform_type
		FROM pinterest_boards
		WHERE board_id IN %s AND %s
		GROUP BY board_id

		UNION ALL

		SELECT if(count_in_range > 0, followers_in_range, followers_latest) AS followers_count,
			channel_id AS account_id, 'youtube' AS platform_type
		FROM (
			SELECT channel_id,
				countIf(%s) AS count_in_range,
				argMaxIf(subscriber_count, created_at, %s) AS followers_in_range,
				argMax(subscriber_count, created_at) AS followers_latest
			FROM youtube_channels
			WHERE channel_id IN %s
			GROUP BY channel_id
		)
	),
	posts_old AS (
		SELECT
			toInt32(fb_posts_agg.total_posts) AS total_posts,
			toInt32(fb_posts_agg.total_engagements) AS total_engagements,
			toInt32(coalesce(fb_insights_agg.total_impression, 0)) AS total_impression,
			toInt32(coalesce(fb_insights_agg.total_reach, 0)) AS total_reach,
			'facebook' AS platform_type, fb_posts_agg.page_id AS account_id
		FROM (
			SELECT count() AS total_posts, sum(total_engagement) AS total_engagements, page_id
			FROM (
				SELECT argMax(total, saving_time) + argMax(comments, saving_time) + argMax(shares, saving_time) + argMax(post_clicks, saving_time) AS total_engagement, page_id
				FROM facebook_posts WHERE (post_id, saving_time) IN fb_posts
				GROUP BY post_id, page_id
			)
			GROUP BY page_id
		) AS fb_posts_agg
		LEFT JOIN (
			SELECT page_id, sum(page_impressions) AS total_impression, sum(page_impressions_unique) AS total_reach
			FROM (
				SELECT page_id, max(page_impressions) AS page_impressions, max(page_impressions_unique) AS page_impressions_unique
				FROM facebook_insights WHERE page_id IN %s AND %s
				GROUP BY page_id, toDate(created_time)
			)
			GROUP BY page_id
		) AS fb_insights_agg ON fb_posts_agg.page_id = fb_insights_agg.page_id

		UNION ALL

		SELECT toInt32(COUNT()) AS total_posts, toInt32(SUM(engagement)) AS total_engagements,
			toInt32(SUM(impressions)) AS total_impression, toInt32(SUM(reach)) AS total_reach,
			'instagram' AS platform_type, instagram_id AS account_id
		FROM (
			SELECT argMax(engagement, stored_event_at) AS engagement, argMax(views, stored_event_at) AS impressions, argMax(reach, stored_event_at) AS reach, instagram_id
			FROM instagram_posts
			WHERE instagram_id IN %s AND %s
			GROUP BY instagram_id, media_id
		)
		GROUP BY instagram_id

		UNION ALL

		SELECT toInt32(COUNT()) AS total_posts, toInt32(SUM(total_engagement)) AS total_engagements,
			toInt32(SUM(total_impressions)) AS total_impression, toInt32(SUM(total_reach)) AS total_reach,
			'linkedin' AS platform_type, linkedin_id AS account_id
		FROM (
			SELECT argMax(favorites, saving_time) + argMax(comments, saving_time) + argMax(repost, saving_time) AS total_engagement,
				argMax(impressions, saving_time) AS total_impressions, argMax(reach, saving_time) AS total_reach, linkedin_id
			FROM linkedin_posts WHERE linkedin_id IN %s AND %s
			GROUP BY linkedin_id, post_id
		)
		GROUP BY linkedin_id

		UNION ALL

		SELECT toInt32(COUNT()) AS total_posts, toInt32(SUM(total_engagement)) AS total_engagements,
			toInt32(SUM(view_count)) AS total_impression, toInt32(SUM(view_count)) AS total_reach,
			'tiktok' AS platform_type, tiktok_id AS account_id
		FROM (
			SELECT argMax(like_count, inserted_at) + argMax(comments_count, inserted_at) + argMax(share_count, inserted_at) AS total_engagement,
				argMax(view_count, inserted_at) AS view_count, tiktok_id
			FROM tiktok_posts WHERE tiktok_id IN %s AND %s
			GROUP BY tiktok_id, post_id
		)
		GROUP BY tiktok_id

		UNION ALL

		SELECT toInt32(COUNT()) AS total_posts, toInt32(SUM(total_engagement)) AS total_engagements,
			toInt32(SUM(views)) AS total_impression, 0 AS total_reach, -- reach metric not available for YouTube
			'youtube' AS platform_type, channel_id AS account_id
		FROM (
			SELECT argMax(likes, inserted_at) + argMax(comments, inserted_at) + argMax(shares, inserted_at) + argMax(dislikes, inserted_at) AS total_engagement,
				argMax(views, inserted_at) AS views, channel_id
			FROM youtube_videos WHERE channel_id IN %s AND %s
			GROUP BY channel_id, video_id
		)
		GROUP BY channel_id

		UNION ALL

		WITH pins AS (
			SELECT pin_id, board_id FROM pinterest_pins
			WHERE board_id IN %s AND %s
			GROUP BY pin_id, board_id
		)
		SELECT toInt32(COUNT(DISTINCT pins.pin_id)) AS total_posts, toInt32(SUM(coalesce(pinterest_insights.engagement, 0))) AS total_engagements,
			toInt32(SUM(coalesce(pinterest_insights.impression, 0))) AS total_impression, 0 AS total_reach, -- reach metric not available for Pinterest
			'pinterest' AS platform_type, pins.board_id AS account_id
		FROM pins
		LEFT JOIN (SELECT pin_id, engagement, impression FROM pinterest_pin_insights WHERE pin_id IN (SELECT pin_id FROM pins)) AS pinterest_insights
			ON pins.pin_id = pinterest_insights.pin_id
		GROUP BY board_id
	)
	SELECT
		toInt32(COALESCE(followers_old.followers_count, 0)) AS followers,
		toInt32(COALESCE(posts_old.total_posts, 0)) AS total_posts,
		toInt32(COALESCE(posts_old.total_engagements, 0)) AS engagement,
		toInt32(COALESCE(posts_old.total_impression, 0)) AS impressions,
		toInt32(COALESCE(posts_old.total_reach, 0)) AS reach,
		followers_old.platform_type AS platform_type,
		followers_old.account_id AS account_id
	FROM followers_old
	LEFT JOIN posts_old
		ON followers_old.account_id = posts_old.account_id
		AND followers_old.platform_type = posts_old.platform_type`,
		// fb_posts CTE uses CURRENT period (matching PHP)
		p.FacebookIDs, p.dateFilter("created_time"),
		// followers_old uses SECONDARY period
		p.FacebookIDs, p.secondaryDateFilter("saving_time"),
		p.InstagramIDs, p.secondaryDateFilter("stored_event_at"),
		p.LinkedInIDs, p.secondaryDateFilter("inserted_at"),
		p.TiktokIDs, p.secondaryDateFilter("inserted_at"),
		p.PinterestIDs, p.secondaryDateFilter("inserted_at"),
		p.secondaryDateFilter("created_at"), p.secondaryDateFilter("created_at"), p.YouTubeIDs,
		// posts_old uses SECONDARY period
		p.FacebookIDs, p.secondaryDateFilter("created_time"), // facebook_insights for impressions/reach
		p.InstagramIDs, p.secondaryDateFilter("post_created_at"),
		p.LinkedInIDs, p.secondaryDateFilter("created_at"),
		p.TiktokIDs, p.secondaryDateFilter("created_at"),
		p.YouTubeIDs, p.secondaryDateFilter("published_at"),
		p.PinterestIDs, p.secondaryDateFilter("created_at"),
	)
}

// GetTopPerformingGraph runs getTopPerformingGraphQuery — daily time-series from mv_social_daily_metrics.
func (r *Repository) GetTopPerformingGraph(ctx context.Context, params *OverviewParams) (*TopPerformingGraphResult, error) {
	window := fmt.Sprintf(
		"toDateTime(date) BETWEEN toDateTime('%s',0) AND toDateTime('%s',0)",
		params.CurrentStart, params.CurrentEnd,
	)

	query := fmt.Sprintf(`
		WITH date_range AS (
			SELECT toDate(addDays(toDate('%s'), number)) AS date
			FROM numbers(dateDiff('day', toDate('%s'), toDate('%s')) + 1)
		),
		platforms AS (
			SELECT 'facebook' AS platform UNION ALL SELECT 'instagram' UNION ALL SELECT 'linkedin'
			UNION ALL SELECT 'tiktok' UNION ALL SELECT 'youtube' UNION ALL SELECT 'pinterest'
		),
		date_platform_combinations AS (
			SELECT dr.date, p.platform FROM date_range dr CROSS JOIN platforms p
		),
		actual_data AS (
			SELECT date, toString(platform) AS platform,
				uniqMerge(posts_count) AS post_cnt,
				sumMerge(engagement_sum) AS eng_cnt,
				sumMerge(impressions_sum) AS impr_cnt,
				sumMerge(reach_sum) AS reach_cnt
			FROM mv_social_daily_metrics
			WHERE %s AND account_id IN %s
			GROUP BY date, platform
		),
		daily AS (
			SELECT dpc.date, dpc.platform,
				coalesce(ad.post_cnt, 0) AS post_cnt,
				coalesce(ad.eng_cnt, 0) AS eng_cnt,
				coalesce(ad.impr_cnt, 0) AS impr_cnt,
				coalesce(ad.reach_cnt, 0) AS reach_cnt
			FROM date_platform_combinations dpc
			LEFT JOIN actual_data ad ON dpc.date = ad.date AND dpc.platform = ad.platform
		)
		SELECT
			arraySort(arrayDistinct(groupArray(date))) AS buckets,
			groupArrayIf(post_cnt, platform = 'facebook') AS facebook_post_count,
			groupArrayIf(post_cnt, platform = 'instagram') AS instagram_post_count,
			groupArrayIf(post_cnt, platform = 'linkedin') AS linkedin_post_count,
			groupArrayIf(post_cnt, platform = 'tiktok') AS tiktok_post_count,
			groupArrayIf(post_cnt, platform = 'youtube') AS youtube_post_count,
			groupArrayIf(post_cnt, platform = 'pinterest') AS pinterest_post_count,
			groupArrayIf(eng_cnt, platform = 'facebook') AS facebook_engagement_count,
			groupArrayIf(eng_cnt, platform = 'instagram') AS instagram_engagement_count,
			groupArrayIf(eng_cnt, platform = 'linkedin') AS linkedin_engagement_count,
			groupArrayIf(eng_cnt, platform = 'tiktok') AS tiktok_engagement_count,
			groupArrayIf(eng_cnt, platform = 'youtube') AS youtube_engagement_count,
			groupArrayIf(eng_cnt, platform = 'pinterest') AS pinterest_engagement_count,
			groupArrayIf(impr_cnt, platform = 'facebook') AS facebook_impression_count,
			groupArrayIf(impr_cnt, platform = 'instagram') AS instagram_impression_count,
			groupArrayIf(impr_cnt, platform = 'linkedin') AS linkedin_impression_count,
			groupArrayIf(impr_cnt, platform = 'tiktok') AS tiktok_impression_count,
			groupArrayIf(impr_cnt, platform = 'youtube') AS youtube_impression_count,
			groupArrayIf(impr_cnt, platform = 'pinterest') AS pinterest_impression_count,
			groupArrayIf(reach_cnt, platform = 'facebook') AS facebook_reach_count,
			groupArrayIf(reach_cnt, platform = 'instagram') AS instagram_reach_count,
			groupArrayIf(reach_cnt, platform = 'linkedin') AS linkedin_reach_count,
			groupArrayIf(reach_cnt, platform = 'tiktok') AS tiktok_reach_count,
			groupArrayIf(0, platform = 'youtube') AS youtube_reach_count,    -- reach metric not available for YouTube
			groupArrayIf(0, platform = 'pinterest') AS pinterest_reach_count -- reach metric not available for Pinterest
		FROM daily`,
		params.CurrentStart, params.CurrentStart, params.CurrentEnd,
		window, params.AllAccounts,
	)

	var result TopPerformingGraphResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.Buckets,
		&result.FacebookPostCount, &result.InstagramPostCount, &result.LinkedInPostCount,
		&result.TiktokPostCount, &result.YouTubePostCount, &result.PinterestPostCount,
		&result.FacebookEngagementCount, &result.InstagramEngagementCount, &result.LinkedInEngagementCount,
		&result.TiktokEngagementCount, &result.YouTubeEngagementCount, &result.PinterestEngagementCount,
		&result.FacebookImpressionCount, &result.InstagramImpressionCount, &result.LinkedInImpressionCount,
		&result.TiktokImpressionCount, &result.YouTubeImpressionCount, &result.PinterestImpressionCount,
		&result.FacebookReachCount, &result.InstagramReachCount, &result.LinkedInReachCount,
		&result.TiktokReachCount, &result.YouTubeReachCount, &result.PinterestReachCount,
	)
	return &result, err
}

// buildPostsGroupedSections builds per-platform aggregated post sections (one row per platform) for GetPlatformData.
// All inner subqueries use argMax (not last_value/first_value) and every SELECT that uses aggregates has a GROUP BY,
// preventing ClickHouse 24+ new analyzer from inlining subqueries and seeing nested aggregate errors.
func (p *OverviewParams) buildPostsGroupedSections() string {
	var sections []string

	if p.IncludeFacebook {
		sections = append(sections, fmt.Sprintf(`
		SELECT
			toInt32(fb_post_data.total_posts) AS total_posts,
			toInt32(fb_post_data.total_engagements) AS total_engagements,
			toInt32(fb_insights_data.total_impression) AS total_impressions,
			toInt32(fb_insights_data.total_reach) AS total_reach,
			toInt32(fb_post_data.total_reactions) AS total_reactions,
			toInt32(fb_post_data.total_comments) AS total_comments,
			toInt32(fb_post_data.total_shares) AS total_shares,
			'facebook' AS platform_type
		FROM (
			SELECT count() AS total_posts,
				sum(total + comments + shares + post_clicks) AS total_engagements,
				sum(total) AS total_reactions,
				sum(comments) AS total_comments,
				sum(shares) AS total_shares
			FROM (
				SELECT
					argMax(total, saving_time) AS total,
					argMax(comments, saving_time) AS comments,
					argMax(shares, saving_time) AS shares,
					argMax(post_clicks, saving_time) AS post_clicks
				FROM facebook_posts
				WHERE (post_id, saving_time) IN (
					SELECT post_id, max(saving_time) FROM facebook_posts WHERE page_id IN %s AND %s GROUP BY post_id
				)
				GROUP BY post_id
			)
			GROUP BY tuple()
		) AS fb_post_data
		CROSS JOIN (
			SELECT sum(page_impressions) AS total_impression, sum(page_impressions_unique) AS total_reach
			FROM (
				SELECT max(page_impressions) AS page_impressions, max(page_impressions_unique) AS page_impressions_unique
				FROM facebook_insights WHERE page_id IN %s AND %s
				GROUP BY page_id, toDate(created_time)
			)
			GROUP BY tuple()
		) AS fb_insights_data`,
			p.FacebookIDs, p.dateFilter("created_time"),
			p.FacebookIDs, p.dateFilter("created_time"),
		))
	}

	if p.IncludeInstagram {
		sections = append(sections, fmt.Sprintf(`
		SELECT
			toInt32(count()) AS total_posts,
			toInt32(sum(engagement)) AS total_engagements,
			toInt32(sum(impressions)) AS total_impressions,
			toInt32(sum(reach)) AS total_reach,
			toInt32(sum(reactions)) AS total_reactions,
			toInt32(sum(comments)) AS total_comments,
			toInt32(sum(shares)) AS total_shares,
			'instagram' AS platform_type
		FROM (
			SELECT
				argMax(like_count, stored_event_at) + argMax(comments_count, stored_event_at) + argMax(saved, stored_event_at) AS engagement,
				argMax(views, stored_event_at) AS impressions,
				argMax(reach, stored_event_at) AS reach,
				argMax(like_count, stored_event_at) AS reactions,
				argMax(comments_count, stored_event_at) AS comments,
				argMax(saved, stored_event_at) AS shares
			FROM instagram_posts WHERE instagram_id IN %s AND %s
			GROUP BY media_id
		)
		GROUP BY platform_type`,
			p.InstagramIDs, p.dateFilter("post_created_at"),
		))
	}

	if p.IncludeLinkedIn {
		sections = append(sections, fmt.Sprintf(`
		SELECT
			toInt32(count()) AS total_posts,
			toInt32(sum(total_engagement)) AS total_engagements,
			toInt32(sum(impressions_v)) AS total_impressions,
			toInt32(sum(reach_v)) AS total_reach,
			toInt32(sum(reactions)) AS total_reactions,
			toInt32(sum(comments_v)) AS total_comments,
			toInt32(sum(shares)) AS total_shares,
			'linkedin' AS platform_type
		FROM (
			SELECT
				argMax(favorites, saving_time) + argMax(comments, saving_time) + argMax(repost, saving_time) AS total_engagement,
				argMax(impressions, saving_time) AS impressions_v,
				argMax(reach, saving_time) AS reach_v,
				argMax(favorites, saving_time) AS reactions,
				argMax(comments, saving_time) AS comments_v,
				argMax(repost, saving_time) AS shares
			FROM linkedin_posts WHERE linkedin_id IN %s AND %s
			GROUP BY post_id
		)
		GROUP BY platform_type`,
			p.LinkedInIDs, p.dateFilter("published_at"),
		))
	}

	if p.IncludeTiktok {
		sections = append(sections, fmt.Sprintf(`
		SELECT
			toInt32(count()) AS total_posts,
			toInt32(sum(like_count + comment_count + shares)) AS total_engagements,
			toInt32(sum(view_count)) AS total_impressions,
			toInt32(sum(view_count)) AS total_reach,
			toInt32(sum(like_count)) AS total_reactions,
			toInt32(sum(comment_count)) AS total_comments,
			toInt32(sum(shares)) AS total_shares,
			'tiktok' AS platform_type
		FROM (
			SELECT
				argMax(view_count, inserted_at) AS view_count,
				argMax(like_count, inserted_at) AS like_count,
				argMax(comments_count, inserted_at) AS comment_count,
				argMax(share_count, inserted_at) AS shares
			FROM tiktok_posts WHERE tiktok_id IN %s AND %s
			GROUP BY post_id
		)
		GROUP BY platform_type`,
			p.TiktokIDs, p.dateFilter("created_at"),
		))
	}

	if p.IncludeYouTube {
		sections = append(sections, fmt.Sprintf(`
		SELECT
			toInt32(count()) AS total_posts,
			toInt32(sum(total_engagement)) AS total_engagements,
			toInt32(sum(views_v)) AS total_impressions,
			0 AS total_reach, -- reach metric not available for YouTube
			toInt32(sum(likes_v)) AS total_reactions,
			toInt32(sum(comments_v)) AS total_comments,
			toInt32(sum(shares_v)) AS total_shares,
			'youtube' AS platform_type
		FROM (
			SELECT
				argMax(likes, inserted_at) + argMax(comments, inserted_at) + argMax(shares, inserted_at) + argMax(dislikes, inserted_at) AS total_engagement,
				argMax(views, inserted_at) AS views_v,
				argMax(likes, inserted_at) AS likes_v,
				argMax(comments, inserted_at) AS comments_v,
				argMax(shares, inserted_at) AS shares_v
			FROM youtube_videos WHERE channel_id IN %s AND %s
			GROUP BY video_id
		)
		GROUP BY platform_type`,
			p.YouTubeIDs, p.dateFilter("published_at"),
		))
	}

	if p.IncludePinterest {
		sections = append(sections, fmt.Sprintf(`
		SELECT
			toInt32(count(DISTINCT pins.pin_id)) AS total_posts,
			toInt32(sum(coalesce(pi.engagement, 0))) AS total_engagements,
			toInt32(sum(coalesce(pi.impression, 0))) AS total_impressions,
			0 AS total_reach, -- reach metric not available for Pinterest
			toInt32(sum(coalesce(pi.pin_clicks, 0))) AS total_reactions,
			toInt32(sum(coalesce(pi.outbound_click, 0))) AS total_comments,
			toInt32(sum(coalesce(pi.saves, 0))) AS total_shares,
			'pinterest' AS platform_type
		FROM (SELECT pin_id FROM pinterest_pins WHERE board_id IN %s AND %s GROUP BY pin_id) AS pins
		LEFT JOIN (
			SELECT pin_id, engagement, impression, pin_clicks, outbound_click, saves
			FROM pinterest_pin_insights
			WHERE pin_id IN (SELECT pin_id FROM pinterest_pins WHERE board_id IN %s AND %s GROUP BY pin_id)
		) AS pi ON pins.pin_id = pi.pin_id
		GROUP BY platform_type`,
			p.PinterestIDs, p.dateFilter("created_at"),
			p.PinterestIDs, p.dateFilter("created_at"),
		))
	}

	return strings.Join(sections, "\nUNION ALL\n")
}

// buildPostsPerAccountSections builds per-platform post sections with account_id for GetAccountData.
func (p *OverviewParams) buildPostsPerAccountSections() string {
	var sections []string

	if p.IncludeFacebook {
		sections = append(sections, fmt.Sprintf(`
		SELECT
			toInt32(count()) AS total_posts,
			toInt32(sum(total_engagement)) AS total_engagements,
			toInt32(sum(total_impressions)) AS total_impressions,
			toInt32(sum(reach)) AS total_reach,
			toInt32(sum(reactions)) AS total_reactions,
			toInt32(sum(comment_count)) AS total_comments,
			toInt32(sum(share_count)) AS total_shares,
			'facebook' AS platform_type,
			page_id AS account_id
		FROM (
			SELECT
				argMax(total, saving_time) + argMax(comments, saving_time) + argMax(shares, saving_time) + argMax(post_clicks, saving_time) AS total_engagement,
				argMax(post_impressions, saving_time) AS total_impressions,
				argMax(post_impressions_unique, saving_time) AS reach,
				argMax(total, saving_time) AS reactions, argMax(comments, saving_time) AS comment_count, argMax(shares, saving_time) AS share_count,
				page_id
			FROM facebook_posts
			WHERE (post_id, saving_time) IN (
				SELECT post_id, max(saving_time) FROM facebook_posts WHERE page_id IN %s AND %s GROUP BY post_id
			)
			GROUP BY page_id, post_id
		)
		GROUP BY page_id`,
			p.FacebookIDs, p.dateFilter("created_time"),
		))
	}

	if p.IncludeInstagram {
		sections = append(sections, fmt.Sprintf(`
		SELECT
			toInt32(count()) AS total_posts,
			toInt32(sum(engagement)) AS total_engagements,
			toInt32(sum(impressions)) AS total_impressions,
			toInt32(sum(reach)) AS total_reach,
			toInt32(sum(reactions)) AS total_reactions,
			toInt32(sum(comments)) AS total_comments,
			toInt32(sum(shares)) AS total_shares,
			'instagram' AS platform_type,
			instagram_id AS account_id
		FROM (
			SELECT argMax(engagement, stored_event_at) AS engagement, argMax(views, stored_event_at) AS impressions,
				argMax(reach, stored_event_at) AS reach, argMax(like_count, stored_event_at) AS reactions,
				argMax(comments_count, stored_event_at) AS comments, argMax(saved, stored_event_at) AS shares, instagram_id
			FROM instagram_posts WHERE instagram_id IN %s AND %s
			GROUP BY instagram_id, media_id
		)
		GROUP BY instagram_id`,
			p.InstagramIDs, p.dateFilter("post_created_at"),
		))
	}

	if p.IncludeLinkedIn {
		sections = append(sections, fmt.Sprintf(`
		SELECT
			toInt32(count()) AS total_posts,
			toInt32(sum(total_engagement)) AS total_engagements,
			toInt32(sum(impressions)) AS total_impressions,
			toInt32(sum(reach)) AS total_reach,
			toInt32(sum(reactions)) AS total_reactions,
			toInt32(sum(comment_count)) AS total_comments,
			toInt32(sum(shares)) AS total_shares,
			'linkedin' AS platform_type,
			linkedin_id AS account_id
		FROM (
			SELECT argMax(favorites, saving_time) + argMax(comments, saving_time) + argMax(repost, saving_time) AS total_engagement,
				argMax(impressions, saving_time) AS impressions, argMax(reach, saving_time) AS reach,
				argMax(favorites, saving_time) AS reactions, argMax(comments, saving_time) AS comment_count, argMax(repost, saving_time) AS shares,
				linkedin_id
			FROM linkedin_posts WHERE linkedin_id IN %s AND %s
			GROUP BY linkedin_id, post_id
		)
		GROUP BY linkedin_id`,
			p.LinkedInIDs, p.dateFilter("created_at"),
		))
	}

	if p.IncludeTiktok {
		sections = append(sections, fmt.Sprintf(`
		SELECT
			toInt32(count()) AS total_posts,
			toInt32(sum(like_count) + sum(comment_count) + sum(shares)) AS total_engagements,
			toInt32(sum(view_count)) AS total_impressions,
			toInt32(sum(view_count)) AS total_reach,
			toInt32(sum(like_count)) AS total_reactions,
			toInt32(sum(comment_count)) AS total_comments,
			toInt32(sum(shares)) AS total_shares,
			'tiktok' AS platform_type,
			tiktok_id AS account_id
		FROM (
			SELECT argMax(view_count, inserted_at) AS view_count, argMax(like_count, inserted_at) AS like_count,
				argMax(comments_count, inserted_at) AS comment_count, argMax(share_count, inserted_at) AS shares, tiktok_id
			FROM tiktok_posts WHERE tiktok_id IN %s AND %s
			GROUP BY tiktok_id, post_id
		)
		GROUP BY tiktok_id`,
			p.TiktokIDs, p.dateFilter("created_at"),
		))
	}

	if p.IncludeYouTube {
		sections = append(sections, fmt.Sprintf(`
		SELECT
			toInt32(count()) AS total_posts,
			toInt32(sum(total_engagement)) AS total_engagements,
			toInt32(sum(views)) AS total_impressions,
			0 AS total_reach, -- reach metric not available for YouTube
			toInt32(sum(reactions)) AS total_reactions,
			toInt32(sum(comment_count)) AS total_comments,
			toInt32(sum(share_count)) AS total_shares,
			'youtube' AS platform_type,
			channel_id AS account_id
		FROM (
			SELECT argMax(likes, inserted_at) + argMax(comments, inserted_at) + argMax(shares, inserted_at) + argMax(dislikes, inserted_at) AS total_engagement,
				argMax(views, inserted_at) AS views, argMax(likes, inserted_at) AS reactions,
				argMax(comments, inserted_at) AS comment_count, argMax(shares, inserted_at) AS share_count, channel_id
			FROM youtube_videos WHERE channel_id IN %s AND %s
			GROUP BY channel_id, video_id
		)
		GROUP BY channel_id`,
			p.YouTubeIDs, p.dateFilter("published_at"),
		))
	}

	if p.IncludePinterest {
		sections = append(sections, fmt.Sprintf(`
		SELECT
			toInt32(count(DISTINCT pins.pin_id)) AS total_posts,
			toInt32(sum(coalesce(pi.engagement, 0))) AS total_engagements,
			toInt32(sum(coalesce(pi.impression, 0))) AS total_impressions,
			0 AS total_reach, -- reach metric not available for Pinterest
			toInt32(sum(coalesce(pi.pin_clicks, 0))) AS total_reactions,
			toInt32(sum(coalesce(pi.outbound_click, 0))) AS total_comments,
			toInt32(sum(coalesce(pi.saves, 0))) AS total_shares,
			'pinterest' AS platform_type,
			pins.board_id AS account_id
		FROM (SELECT pin_id, board_id FROM pinterest_pins WHERE board_id IN %s AND %s GROUP BY pin_id, board_id) AS pins
		LEFT JOIN (
			SELECT pin_id, engagement, impression, pin_clicks, outbound_click, saves
			FROM pinterest_pin_insights
			WHERE pin_id IN (SELECT pin_id FROM pinterest_pins WHERE board_id IN %s AND %s GROUP BY pin_id)
		) AS pi ON pins.pin_id = pi.pin_id
		GROUP BY pins.board_id`,
			p.PinterestIDs, p.dateFilter("created_at"),
			p.PinterestIDs, p.dateFilter("created_at"),
		))
	}

	return strings.Join(sections, "\nUNION ALL\n")
}

// buildGraphSections builds per-platform daily time-series sections for GetAccountDataGraphs.
func (p *OverviewParams) buildGraphSections() string {
	var sections []string

	if p.IncludeFacebook {
		sections = append(sections, fmt.Sprintf(`
		SELECT toDate(created_time) AS date, page_id AS account_id,
			toFloat64(SUM(last_value_eng)) AS engagement,
			toFloat64(SUM(last_value_reach)) AS reach,
			toFloat64(SUM(last_value_impr)) AS impressions,
			toFloat64(COUNT(*)) AS posts
		FROM (
			SELECT last_value(total) + last_value(comments) + last_value(shares) + last_value(post_clicks) AS last_value_eng,
				last_value(post_impressions_unique) AS last_value_reach,
				last_value(post_impressions) AS last_value_impr,
				created_time, page_id, post_id
			FROM facebook_posts WHERE page_id IN %s AND %s
			GROUP BY page_id, post_id, created_time
		)
		GROUP BY date, account_id`,
			p.FacebookIDs, p.dateFilter("created_time"),
		))
	}

	if p.IncludeInstagram {
		sections = append(sections, fmt.Sprintf(`
		SELECT toDate(post_created_at) AS date, instagram_id AS account_id,
			toFloat64(SUM(last_value_eng)) AS engagement,
			toFloat64(SUM(last_value_reach)) AS reach,
			toFloat64(SUM(last_value_impr)) AS impressions,
			toFloat64(COUNT(*)) AS posts
		FROM (
			SELECT last_value(engagement) AS last_value_eng, last_value(reach) AS last_value_reach,
				last_value(views) AS last_value_impr, post_created_at, instagram_id, media_id
			FROM instagram_posts WHERE instagram_id IN %s AND %s
			GROUP BY media_id, instagram_id, post_created_at
		)
		GROUP BY date, account_id`,
			p.InstagramIDs, p.dateFilter("post_created_at"),
		))
	}

	if p.IncludeLinkedIn {
		sections = append(sections, fmt.Sprintf(`
		SELECT toDate(created_at) AS date, linkedin_id AS account_id,
			toFloat64(SUM(last_value_eng)) AS engagement,
			toFloat64(SUM(last_value_reach)) AS reach,
			toFloat64(SUM(last_value_impr)) AS impressions,
			toFloat64(COUNT(*)) AS posts
		FROM (
			SELECT first_value(favorites) + first_value(comments) + first_value(repost) AS last_value_eng,
				first_value(reach) AS last_value_reach, first_value(impressions) AS last_value_impr,
				post_id, created_at, linkedin_id
			FROM linkedin_posts WHERE linkedin_id IN %s AND %s
			GROUP BY post_id, linkedin_id, created_at
		)
		GROUP BY date, account_id`,
			p.LinkedInIDs, p.dateFilter("created_at"),
		))
	}

	if p.IncludeYouTube {
		sections = append(sections, fmt.Sprintf(`
		SELECT toDate(published_at) AS date, channel_id AS account_id,
			toFloat64(SUM(last_value_eng)) AS engagement,
			toFloat64(SUM(last_value_reach)) AS reach,
			toFloat64(SUM(last_value_impr)) AS impressions,
			toFloat64(COUNT(*)) AS posts
		FROM (
			SELECT last_value(likes) + last_value(comments) + last_value(shares) + last_value(dislikes) AS last_value_eng,
				last_value(views) AS last_value_impr, last_value(views) AS last_value_reach,
				video_id, channel_id, published_at
			FROM youtube_videos WHERE channel_id IN %s AND %s
			GROUP BY channel_id, video_id, published_at
		)
		GROUP BY date, account_id`,
			p.YouTubeIDs, p.dateFilter("published_at"),
		))
	}

	if p.IncludeTiktok {
		sections = append(sections, fmt.Sprintf(`
		SELECT toDate(created_at) AS date, tiktok_id AS account_id,
			toFloat64(SUM(last_value_eng)) AS engagement,
			toFloat64(0) AS reach,
			toFloat64(SUM(last_value_views)) AS impressions,
			toFloat64(COUNT(*)) AS posts
		FROM (
			SELECT last_value(like_count) + last_value(comments_count) + last_value(share_count) AS last_value_eng,
				last_value(view_count) AS last_value_views, post_id, tiktok_id, created_at
			FROM tiktok_posts WHERE tiktok_id IN %s AND %s
			GROUP BY post_id, tiktok_id, created_at
		)
		GROUP BY date, account_id`,
			p.TiktokIDs, p.dateFilter("created_at"),
		))
	}

	if p.IncludePinterest {
		sections = append(sections, fmt.Sprintf(`
		SELECT pins.date AS date, pins.account_id AS account_id,
			toFloat64(SUM(toInt32(pi.engagement))) AS engagement,
			toFloat64(0) AS reach, toFloat64(0) AS impressions,
			toFloat64(countDistinct(pins.pin_id)) AS posts
		FROM (
			SELECT board_id AS account_id, pin_id, toDate(created_at) AS date
			FROM pinterest_pins WHERE board_id IN %s AND %s
			GROUP BY board_id, pin_id, date
		) AS pins
		LEFT JOIN (
			SELECT pin_id, engagement FROM pinterest_pin_insights
			WHERE pin_id IN (SELECT pin_id FROM pinterest_pins WHERE board_id IN %s AND %s GROUP BY pin_id)
		) pi ON pins.pin_id = pi.pin_id
		GROUP BY pins.date, pins.account_id`,
			p.PinterestIDs, p.dateFilter("created_at"),
			p.PinterestIDs, p.dateFilter("created_at"),
		))
	}

	return strings.Join(sections, "\nUNION ALL\n")
}

// GetPlatformData executes a single UNION ALL query (followers LEFT JOIN posts) grouped by platform_type.
func (r *Repository) GetPlatformData(ctx context.Context, params *OverviewParams) ([]PlatformDataRow, error) {
	followersSections := params.buildFollowersSections(false)
	if followersSections == "" {
		return []PlatformDataRow{}, nil
	}
	postsSections := params.buildPostsGroupedSections()

	var query string
	if postsSections == "" {
		query = fmt.Sprintf(`
		SELECT toInt32(sum(followers_count)) AS followers,
			toInt32(0) AS total_posts, toInt32(0) AS engagement, toInt32(0) AS impressions,
			toInt32(0) AS reach, toInt32(0) AS reactions, toInt32(0) AS comments, toInt32(0) AS shares,
			platform_type
		FROM (%s)
		GROUP BY platform_type
		ORDER BY platform_type DESC`, followersSections)
	} else {
		query = fmt.Sprintf(`
		SELECT
			toInt32(f.followers_count) AS followers,
			toInt32(coalesce(p.total_posts, 0)) AS total_posts,
			toInt32(coalesce(p.total_engagements, 0)) AS engagement,
			toInt32(coalesce(p.total_impressions, 0)) AS impressions,
			toInt32(coalesce(p.total_reach, 0)) AS reach,
			toInt32(coalesce(p.total_reactions, 0)) AS reactions,
			toInt32(coalesce(p.total_comments, 0)) AS comments,
			toInt32(coalesce(p.total_shares, 0)) AS shares,
			f.platform_type
		FROM (
			SELECT toInt32(sum(followers_count)) AS followers_count, platform_type
			FROM (%s)
			GROUP BY platform_type
		) AS f
		LEFT JOIN (
			%s
		) AS p ON f.platform_type = p.platform_type
		ORDER BY f.platform_type DESC`,
			followersSections, postsSections)
	}

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []PlatformDataRow
	for rows.Next() {
		var row PlatformDataRow
		if err := rows.Scan(
			&row.Followers, &row.TotalPosts, &row.Engagement,
			&row.Impressions, &row.Reach, &row.Reactions, &row.Comments, &row.Shares,
			&row.PlatformType,
		); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	return results, rows.Err()
}

// GetAccountData executes a single UNION ALL query (followers LEFT JOIN posts) grouped by (account_id, platform_type).
func (r *Repository) GetAccountData(ctx context.Context, params *OverviewParams) ([]AccountDataRow, error) {
	followersSections := params.buildFollowersSections(true)
	if followersSections == "" {
		return []AccountDataRow{}, nil
	}
	postsSections := params.buildPostsPerAccountSections()

	var query string
	if postsSections == "" {
		query = fmt.Sprintf(`
		SELECT toInt32(followers_count) AS followers,
			toInt32(0) AS total_posts, toInt32(0) AS engagement, toInt32(0) AS impressions,
			toInt32(0) AS reach, toInt32(0) AS reactions, toInt32(0) AS comments, toInt32(0) AS shares,
			platform_type, account_id
		FROM (%s)
		ORDER BY platform_type DESC, account_id DESC`, followersSections)
	} else {
		query = fmt.Sprintf(`
		SELECT
			toInt32(f.followers_count) AS followers,
			toInt32(coalesce(p.total_posts, 0)) AS total_posts,
			toInt32(coalesce(p.total_engagements, 0)) AS engagement,
			toInt32(coalesce(p.total_impressions, 0)) AS impressions,
			toInt32(coalesce(p.total_reach, 0)) AS reach,
			toInt32(coalesce(p.total_reactions, 0)) AS reactions,
			toInt32(coalesce(p.total_comments, 0)) AS comments,
			toInt32(coalesce(p.total_shares, 0)) AS shares,
			f.platform_type,
			f.account_id
		FROM (%s) AS f
		LEFT JOIN (
			%s
		) AS p ON f.platform_type = p.platform_type AND f.account_id = p.account_id
		ORDER BY f.platform_type DESC, f.account_id DESC`,
			followersSections, postsSections)
	}

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []AccountDataRow
	for rows.Next() {
		var row AccountDataRow
		if err := rows.Scan(
			&row.Followers, &row.TotalPosts, &row.Engagement,
			&row.Impressions, &row.Reach, &row.Reactions, &row.Comments, &row.Shares,
			&row.PlatformType, &row.AccountID,
		); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	return results, rows.Err()
}

type accountPeriodData struct {
	Followers    int32
	TotalPosts   int32
	Engagement   int32
	Impressions  int32
	Reach        int32
	PlatformType string
	AccountID    string
	AccountName  string
}

// GetAccountDataDetailed runs current and old data queries concurrently, then performs a
// full outer join in Go to compute per-account current vs previous period metrics with pct changes.
func (r *Repository) GetAccountDataDetailed(ctx context.Context, params *OverviewParams) ([]AccountDataDetailedRow, error) {
	type periodResult struct {
		rows []accountPeriodData
		err  error
	}

	currentCh := make(chan periodResult, 1)
	oldCh := make(chan periodResult, 1)

	go func() {
		rows, err := r.queryCurrentData(ctx, params)
		currentCh <- periodResult{rows, err}
	}()
	go func() {
		rows, err := r.queryOldData(ctx, params)
		oldCh <- periodResult{rows, err}
	}()

	currentRes := <-currentCh
	oldRes := <-oldCh

	if currentRes.err != nil {
		return nil, currentRes.err
	}
	if oldRes.err != nil {
		return nil, oldRes.err
	}

	type detailedKey struct {
		AccountID    string
		PlatformType string
	}

	currentMap := make(map[detailedKey]accountPeriodData, len(currentRes.rows))
	for _, d := range currentRes.rows {
		currentMap[detailedKey{d.AccountID, d.PlatformType}] = d
	}
	oldMap := make(map[detailedKey]accountPeriodData, len(oldRes.rows))
	for _, d := range oldRes.rows {
		oldMap[detailedKey{d.AccountID, d.PlatformType}] = d
	}

	allKeys := make(map[detailedKey]struct{}, len(currentMap)+len(oldMap))
	for k := range currentMap {
		allKeys[k] = struct{}{}
	}
	for k := range oldMap {
		allKeys[k] = struct{}{}
	}

	results := make([]AccountDataDetailedRow, 0, len(allKeys))
	for key := range allKeys {
		cd := currentMap[key]
		od := oldMap[key]
		results = append(results, AccountDataDetailedRow{
			PlatformType:         key.PlatformType,
			AccountID:            key.AccountID,
			AccountName:          cd.AccountName,
			CurrentFollowers:     cd.Followers,
			OldFollowers:         od.Followers,
			CurrentPosts:         cd.TotalPosts,
			OldPosts:             od.TotalPosts,
			CurrentEngagement:    cd.Engagement,
			OldEngagement:        od.Engagement,
			CurrentImpressions:   cd.Impressions,
			OldImpressions:       od.Impressions,
			CurrentReach:         cd.Reach,
			OldReach:             od.Reach,
			FollowersChangePct:   detailedPctChange(cd.Followers, od.Followers),
			PostsChangePct:       detailedPctChange(cd.TotalPosts, od.TotalPosts),
			EngagementChangePct:  detailedPctChange(cd.Engagement, od.Engagement),
			ImpressionsChangePct: detailedPctChange(cd.Impressions, od.Impressions),
			ReachChangePct:       detailedPctChange(cd.Reach, od.Reach),
		})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].PlatformType > results[j].PlatformType
	})
	return results, nil
}

// GetAccountDataGraphs executes a single UNION ALL query and uses groupArray() to assemble
// per-account time-series arrays in ClickHouse — PHP getAccountDataGraphsQuery() approach.
func (r *Repository) GetAccountDataGraphs(ctx context.Context, params *OverviewParams) ([]AccountDataGraphsRow, error) {
	graphSections := params.buildGraphSections()
	if graphSections == "" {
		return []AccountDataGraphsRow{}, nil
	}

	query := fmt.Sprintf(`
	SELECT account_id,
		groupArray(engagement) AS engagement,
		groupArray(reach) AS reach,
		groupArray(impressions) AS impressions,
		groupArray(posts) AS posts,
		groupArray(date) AS buckets
	FROM (
		SELECT account_id, date,
			toFloat64(sum(engagement)) AS engagement,
			toFloat64(sum(reach)) AS reach,
			toFloat64(sum(impressions)) AS impressions,
			toFloat64(sum(posts)) AS posts
		FROM (
			%s
		)
		GROUP BY account_id, date
		ORDER BY account_id ASC, date ASC
	)
	GROUP BY account_id`, graphSections)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []AccountDataGraphsRow
	for rows.Next() {
		var row AccountDataGraphsRow
		if err := rows.Scan(
			&row.AccountID, &row.Engagement, &row.Reach, &row.Impressions, &row.Posts, &row.Buckets,
		); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sort.Slice(results, func(i, j int) bool {
		var ei, ej float64
		for _, v := range results[i].Engagement {
			ei += v
		}
		for _, v := range results[j].Engagement {
			ej += v
		}
		return ei > ej
	})
	return results, nil
}

// topPostsGlobalSortCol maps the sort type to a SQL column name usable across all platform sections.
func topPostsGlobalSortCol(sortType string) string {
	switch strings.ToLower(strings.TrimSpace(sortType)) {
	case "likes":
		return "likes"
	case "comments":
		return "comments"
	case "shares":
		return "shares"
	case "saves":
		return "saves"
	case "pin_clicks":
		return "pin_clicks"
	case "outbound_clicks":
		return "outbound_clicks"
	case "dislikes_count":
		return "dislikes_count"
	case "views":
		return "views"
	case "reach":
		return "reach"
	case "impressions", "total_impressions":
		return "greatest(views, reach)"
	default:
		return "total_engagement"
	}
}

// GetTopPosts executes a UNION ALL query across included platforms while applying
// the requested limit per platform section, then globally ordering merged results.
func (r *Repository) GetTopPosts(ctx context.Context, params *OverviewParams) ([]TopPostRow, error) {
	var withParts []string
	var sections []string
	sortCol := topPostsGlobalSortCol(params.Type)
	addSection := func(sectionSQL string) {
		sections = append(sections, fmt.Sprintf(`
		SELECT * FROM (
			%s
		) ORDER BY %s DESC LIMIT %d`,
			sectionSQL, sortCol, params.Limit,
		))
	}

	if params.IncludeFacebook {
		withParts = append(withParts, fmt.Sprintf(`facebook_latest_posts AS (
			SELECT post_id, max(saving_time) AS max_saving_time
			FROM facebook_posts WHERE page_id IN %s AND %s
			GROUP BY post_id
		)`, params.FacebookIDs, params.dateFilter("created_time")))
		addSection(fmt.Sprintf(`
		SELECT 'facebook' AS platform_type, page_id AS account_id, post_id,
			toInt32(total_v) AS likes, toInt32(comments_v) AS comments,
			toInt32(shares_v) AS shares, toInt32(0) AS saves,
			toInt32(0) AS pin_clicks, toInt32(0) AS outbound_clicks, toInt32(0) AS dislikes_count,
			permalink_v AS permalink, media_type_v AS media_type,
			thumbnail_v AS thumbnail, caption_v AS category,
			created_time,
			toInt32(total_v + comments_v + shares_v + post_clicks_v) AS total_engagement,
			toInt32(post_impressions_v) AS views, toInt32(post_impressions_unique_v) AS reach
		FROM (
			SELECT page_id, post_id, created_time,
				argMax(total, saving_time) AS total_v,
				argMax(comments, saving_time) AS comments_v,
				argMax(shares, saving_time) AS shares_v,
				argMax(post_clicks, saving_time) AS post_clicks_v,
				argMax(post_impressions, saving_time) AS post_impressions_v,
				argMax(post_impressions_unique, saving_time) AS post_impressions_unique_v,
				argMax(permalink, saving_time) AS permalink_v,
				argMax(media_type, saving_time) AS media_type_v,
				argMax(full_picture, saving_time) AS thumbnail_v,
				argMax(caption, saving_time) AS caption_v
			FROM facebook_posts
			WHERE page_id IN %s AND %s
			AND (post_id, saving_time) IN (SELECT post_id, max_saving_time FROM facebook_latest_posts)
			GROUP BY page_id, post_id, created_time
		)`,
			params.FacebookIDs, params.dateFilter("created_time"),
		))
	}

	if params.IncludeInstagram {
		addSection(fmt.Sprintf(`
		SELECT 'instagram' AS platform_type, instagram_id AS account_id, media_id AS post_id,
			toInt32(argMax(like_count, stored_event_at)) AS likes, toInt32(argMax(comments_count, stored_event_at)) AS comments,
			toInt32(0) AS shares, toInt32(argMax(saved, stored_event_at)) AS saves,
			toInt32(0) AS pin_clicks, toInt32(0) AS outbound_clicks, toInt32(0) AS dislikes_count,
			argMax(permalink, stored_event_at) AS permalink, argMax(media_type, stored_event_at) AS media_type,
			if(length(argMax(media_url, stored_event_at)) > 0, arrayElement(argMax(media_url, stored_event_at), 1), '') AS thumbnail,
			argMax(caption, stored_event_at) AS category,
			post_created_at AS created_time,
			toInt32(argMax(like_count + comments_count + saved, stored_event_at)) AS total_engagement,
			toInt32(argMax(views, stored_event_at)) AS views, toInt32(argMax(reach, stored_event_at)) AS reach
		FROM (SELECT * FROM instagram_posts WHERE instagram_id IN %s AND %s AND media_type != 'STORY')
		GROUP BY instagram_id, media_id, post_created_at`,
			params.InstagramIDs, params.dateFilter("post_created_at"),
		))
	}

	if params.IncludeLinkedIn {
		addSection(fmt.Sprintf(`
		SELECT 'linkedin' AS platform_type, linkedin_id AS account_id, post_id,
			toInt32(fav_v) AS likes, toInt32(comments_v) AS comments,
			toInt32(repost_v) AS shares, toInt32(0) AS saves,
			toInt32(0) AS pin_clicks, toInt32(0) AS outbound_clicks, toInt32(0) AS dislikes_count,
			permalink_v AS permalink, media_type_v AS media_type,
			thumbnail_v AS thumbnail, title_v AS category,
			created_at AS created_time,
			toInt32(fav_v + comments_v + repost_v) AS total_engagement,
			toInt32(0) AS views, toInt32(reach_v) AS reach
		FROM (
			SELECT linkedin_id, post_id, created_at,
				argMax(favorites, saving_time) AS fav_v,
				argMax(comments, saving_time) AS comments_v,
				argMax(repost, saving_time) AS repost_v,
				argMax(reach, saving_time) AS reach_v,
				argMax(article_url, saving_time) AS permalink_v,
				argMax(media_type, saving_time) AS media_type_v,
				argMax(image, saving_time) AS thumbnail_v,
				argMax(title, saving_time) AS title_v
			FROM linkedin_posts WHERE linkedin_id IN %s AND %s
			GROUP BY linkedin_id, post_id, created_at
		)`,
			params.LinkedInIDs, params.dateFilter("created_at"),
		))
	}

	if params.IncludeTiktok {
		addSection(fmt.Sprintf(`
		SELECT 'tiktok' AS platform_type, tiktok_id AS account_id, post_id,
			toInt32(argMax(like_count, inserted_at)) AS likes, toInt32(argMax(comments_count, inserted_at)) AS comments,
			toInt32(argMax(share_count, inserted_at)) AS shares, toInt32(0) AS saves,
			toInt32(0) AS pin_clicks, toInt32(0) AS outbound_clicks, toInt32(0) AS dislikes_count,
			argMax(share_url, inserted_at) AS permalink, 'video' AS media_type,
			argMax(embed_link, inserted_at) AS thumbnail, argMax(post_description, inserted_at) AS category,
			created_at AS created_time,
			toInt32(argMax(like_count + comments_count + share_count, inserted_at)) AS total_engagement,
			toInt32(argMax(view_count, inserted_at)) AS views, toInt32(argMax(view_count, inserted_at)) AS reach
		FROM tiktok_posts WHERE tiktok_id IN %s AND %s
		GROUP BY tiktok_id, post_id, created_at`,
			params.TiktokIDs, params.dateFilter("created_at"),
		))
	}

	if params.IncludeYouTube {
		addSection(fmt.Sprintf(`
		SELECT 'youtube' AS platform_type, channel_id AS account_id, video_id AS post_id,
			toInt32(likes_v) AS likes, toInt32(comments_v) AS comments,
			toInt32(shares_v) AS shares, toInt32(0) AS saves,
			toInt32(0) AS pin_clicks, toInt32(0) AS outbound_clicks, toInt32(dislikes_v) AS dislikes_count,
			permalink_v AS permalink, media_type_v AS media_type, thumbnail_v AS thumbnail,
			description_v AS category,
			published_at AS created_time,
			toInt32(likes_v + comments_v + shares_v + dislikes_v) AS total_engagement,
			toInt32(views_v) AS views, toInt32(views_v) AS reach
		FROM (
			SELECT channel_id, video_id, published_at,
				argMax(likes, inserted_at) AS likes_v,
				argMax(comments, inserted_at) AS comments_v,
				argMax(shares, inserted_at) AS shares_v,
				argMax(dislikes, inserted_at) AS dislikes_v,
				argMax(views, inserted_at) AS views_v,
				REPLACE(concat('https://', argMax(substring(iframe_embed_html, position('//' IN iframe_embed_html) + length('//'), position('"' IN substring(iframe_embed_html, position('//' IN iframe_embed_html) + length('//'))) - 1), inserted_at)), 'embed/', 'watch?v=') AS permalink_v,
				argMax(media_type, inserted_at) AS media_type_v,
				argMax(thumbnail_url, inserted_at) AS thumbnail_v,
				argMax(description, inserted_at) AS description_v
			FROM youtube_videos WHERE channel_id IN %s AND %s
			GROUP BY channel_id, video_id, published_at
		)`,
			params.YouTubeIDs, params.dateFilter("published_at"),
		))
	}

	if params.IncludePinterest {
		addSection(fmt.Sprintf(`
		SELECT 'pinterest' AS platform_type, pins.board_id AS account_id, pins.pin_id AS post_id,
			toInt32(0) AS likes, toInt32(0) AS comments, toInt32(0) AS shares,
			toInt32(SUM(pin_ins.saves)) AS saves,
			toInt32(SUM(pin_ins.pin_clicks)) AS pin_clicks,
			toInt32(SUM(pin_ins.outbound_click)) AS outbound_clicks,
			toInt32(0) AS dislikes_count,
			any(pins.permalink) AS permalink, any(pins.media_type) AS media_type,
			any(pins.thumbnail) AS thumbnail, any(pins.category) AS category,
			any(pins.created_time) AS created_time,
			toInt32(SUM(pin_ins.saves + pin_ins.pin_clicks + pin_ins.outbound_click)) AS total_engagement,
			toInt32(0) AS views, toInt32(SUM(pin_ins.impression)) AS reach
		FROM (
			SELECT pin_id, board_id,
				format('{}{}', 'https://www.pinterest.com/pin/', pin_id) AS permalink,
				last_value(media_type) AS media_type, last_value(cover_image_url) AS thumbnail,
				last_value(description) AS category, last_value(created_at) AS created_time
			FROM pinterest_pins WHERE board_id IN %s AND %s
			GROUP BY pin_id, board_id
		) AS pins
		LEFT JOIN pinterest_pin_insights AS pin_ins ON pins.pin_id = pin_ins.pin_id
		GROUP BY pins.pin_id, pins.board_id`,
			params.PinterestIDs, params.dateFilter("created_at"),
		))
	}

	if len(sections) == 0 {
		return []TopPostRow{}, nil
	}

	withClause := ""
	if len(withParts) > 0 {
		withClause = "WITH " + strings.Join(withParts, ",\n") + "\n"
	}
	query := fmt.Sprintf(`%sSELECT platform_type, account_id, post_id,
		likes, comments, shares, saves, pin_clicks, outbound_clicks, dislikes_count,
		permalink, media_type, thumbnail, category, created_time,
		total_engagement, views, reach
	FROM (
		%s
	) ORDER BY %s DESC`,
		withClause, strings.Join(sections, "\nUNION ALL\n"), sortCol)

	return r.scanTopPostRows(ctx, query)
}

func (r *Repository) queryCurrentData(ctx context.Context, params *OverviewParams) ([]accountPeriodData, error) {
	rows, err := r.client.Conn.Query(ctx, params.buildCurrentDataQuery())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []accountPeriodData
	for rows.Next() {
		var d accountPeriodData
		if err := rows.Scan(
			&d.Followers, &d.TotalPosts, &d.Engagement, &d.Impressions, &d.Reach,
			&d.PlatformType, &d.AccountID, &d.AccountName,
		); err != nil {
			return nil, err
		}
		results = append(results, d)
	}
	return results, rows.Err()
}

func (r *Repository) queryOldData(ctx context.Context, params *OverviewParams) ([]accountPeriodData, error) {
	rows, err := r.client.Conn.Query(ctx, params.buildOldDataQuery())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []accountPeriodData
	for rows.Next() {
		var d accountPeriodData
		if err := rows.Scan(
			&d.Followers, &d.TotalPosts, &d.Engagement, &d.Impressions, &d.Reach,
			&d.PlatformType, &d.AccountID,
		); err != nil {
			return nil, err
		}
		results = append(results, d)
	}
	return results, rows.Err()
}

func detailedPctChange(current, old int32) float64 {
	if old == 0 {
		return 0
	}
	return math.Round(float64(current-old)*100.0/float64(old)*100) / 100
}

func (r *Repository) scanTopPostRows(ctx context.Context, query string) ([]TopPostRow, error) {
	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []TopPostRow
	for rows.Next() {
		var row TopPostRow
		if err := rows.Scan(
			&row.PlatformType, &row.AccountID, &row.PostID,
			&row.Likes, &row.Comments, &row.Shares, &row.Saves,
			&row.PinClicks, &row.OutboundClicks, &row.DislikesCount,
			&row.Permalink, &row.MediaType, &row.Thumbnail, &row.Category, &row.CreatedTime,
			&row.TotalEngagement, &row.Views, &row.Reach,
		); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	return results, rows.Err()
}
