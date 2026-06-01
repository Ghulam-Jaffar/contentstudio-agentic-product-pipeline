# ContentStudio Analytics API Reference

This document is the complete reference for every analytics API endpoint, ClickHouse SQL query, and business logic rule implemented in the ContentStudio backend. It is intended to be sufficient for a Go developer to reimplement every endpoint by reading only this document.

---

## Table of Contents

1. [Common Patterns and Conventions](#1-common-patterns-and-conventions)
2. [ClickHouse Tables Reference](#2-clickhouse-tables-reference)
3. [Facebook Analytics](#3-facebook-analytics)
4. [Facebook Competitor Analytics](#4-facebook-competitor-analytics)
5. [Instagram Analytics](#5-instagram-analytics)
6. [Instagram Competitor Analytics](#6-instagram-competitor-analytics)
7. [LinkedIn Analytics](#7-linkedin-analytics)
8. [YouTube Analytics](#8-youtube-analytics)
9. [TikTok Analytics](#9-tiktok-analytics)
10. [Pinterest Analytics](#10-pinterest-analytics)
11. [Twitter/X Analytics](#11-twitterx-analytics)
12. [Cross-Platform Overview V2](#12-cross-platform-overview-v2)
13. [Campaign and Label Analytics](#13-campaign-and-label-analytics)
14. [Date Handling Patterns](#14-date-handling-patterns)
15. [Deduplication Strategies](#15-deduplication-strategies)
16. [Zero-Fill Patterns](#16-zero-fill-patterns)
17. [Dynamic Aggregation Patterns](#17-dynamic-aggregation-patterns)
18. [Known Quirks and Bugs](#18-known-quirks-and-bugs)
19. [Cross-Platform Overview V1](#19-cross-platform-overview-v1)
20. [Dashboard Analytics](#20-dashboard-analytics)
21. [Analytics Share Link Management](#21-analytics-share-link-management)
22. [Reports and Scheduled Reports](#22-reports-and-scheduled-reports)
23. [Analytics Job Triggers](#23-analytics-job-triggers)
24. [Twitter/X Settings Management](#24-twitterx-settings-management)
25. [Competitor Management CRUD](#25-competitor-management-crud)
26. [AI Insights](#26-ai-insights)

---

## 1. Common Patterns and Conventions

### 1.1 Request Parameters (Universal)

Every analytics endpoint accepts these core parameters:

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `workspace_id` | string | required | MongoDB workspace identifier |
| `date` | string | required | Format: `"YYYY-MM-DD - YYYY-MM-DD"` |
| `timezone` | string | `"UTC"` | IANA timezone string |
| `{platform}_id` | string/array | required | Platform account identifier(s) |
| `media_type` | array | platform defaults | Filter by content type |
| `limit` | int | varies (5-20) | Max results for top posts |
| `order_by` | string | varies | Sort column for top posts |

### 1.2 Previous Period Calculation

Every endpoint automatically computes a comparison period of equal length:

```
previous_start = start - (end - start)
previous_end = start
```

Example: current range `2024-02-01 - 2024-02-29` (28 days) produces previous `2024-01-04 - 2024-02-01`.

### 1.3 ClickHouse Execution Pattern

All queries run via `Clickhouse::getInstance('analytics')->getClient()` with `->onCluster(env('CLICKHOUSE_ANALYTICS_CLUSTER'))`.

### 1.4 Engagement Formulas by Platform

| Platform | Engagement Formula |
|----------|-------------------|
| **Facebook** | reactions + comments + shares |
| **Instagram** | likes + comments + saved |
| **LinkedIn** | comments + favorites + shares (NOT including clicks) |
| **YouTube** | likes + dislikes + comments + shares |
| **TikTok** | `engagement_count` column (pre-computed) |
| **Pinterest** | saves + pin_clicks + outbound_clicks |
| **Twitter/X** | `total_engagement` column (pre-computed) |

### 1.5 'N/A' Sentinel Value

When a metric has no data (count is zero), the string `'N/A'` is returned instead of null or 0. The controller's difference/percentage calculations must handle this explicitly by checking for `'N/A'` before arithmetic.

### 1.6 Growth/Percentage Formula

```
growth = round((current - previous) / max(previous, 1) * 100, 2)
```

Returns `"N/A"` when previous is 0 or either value is `"N/A"`.

---

## 2. ClickHouse Tables Reference

### Facebook Tables

| Table | Key Columns |
|-------|-------------|
| `facebook_posts` | `post_id`, `page_id`, `saving_time`, `created_time`, `media_type`, `day_of_week`, `hour_of_day`, `total` (reactions), `comments`, `shares`, `post_clicks`, `total_engagement`, `post_impressions`, `post_impressions_unique`, `post_impressions_paid`, `post_impressions_organic`, `post_impressions_viral`, `post_impressions_paid_unique`, `post_impressions_organic_unique`, `post_impressions_viral_unique`, `post_video_views`, `permalink`, `full_picture`, `caption`, `description`, `link`, `status_type`, `video_id`, `category`, `published_by`, `message_tags`, `post_metadata` |
| `facebook_insights` | `page_id`, `hash_id`, `saving_time`, `created_time`, `page_fans`, `page_follows`, `page_impressions`, `page_impressions_paid`, `page_impressions_organic`, `page_post_engagements`, `page_fans_by_like`, `page_fans_by_unlike`, `talking_about_count`, `positive_sentiment`, `negative_sentiment`, `page_positive_feedback`, `page_negative_feedback`, `page_fans_online`, `page_fans_gender`, `page_fans_age`, `page_fans_gender_age`, `page_fans_country`, `page_fans_city`, `day_of_week` |
| `facebook_media_assets` | `post_id`, `page_id`, `media_id`, `caption`, `link`, `assetType`, `callToAction`, `createdAt`, `inserted_at` |
| `facebook_video_insights` | `post_id`, `page_id`, `video_id`, `total_video_views`, `total_video_views_organic`, `total_video_views_paid`, `total_video_view_total_time`, `total_video_view_total_time_organic`, `total_video_view_total_time_paid` |
| `facebook_reels_insights` | `post_id`, `page_id`, `play_count`, `total_time_watched_in_ms`, `average_time_watched`, `impressions_unique`, `created_at` |
| `facebook_competitor_posts` | `facebook_id`, `post_id`, `like`, `haha`, `angry`, `sad`, `love`, `thankful`, `wow`, `total_post_reactions`, `comments`, `shares`, `post_engagement`, `media_type`, `status_type`, `permalink`, `caption`, `hashtags`, `day_of_week`, `hour_of_day`, `created_at`, `inserted_at` |
| `facebook_competitor_insights` | `facebook_id`, `inserted_at`, `followersCount`, `fanCount`, `page_name`, `page_category`, `biography`, `slug`, `image` |
| `facebook_competitor_media_assets` | `post_id`, `facebook_id`, `media_id`, `caption`, `link`, `asset_type`, `call_to_action`, `created_at` |

### Instagram Tables

| Table | Key Columns |
|-------|-------------|
| `instagram_posts` | `media_id`, `instagram_id`, `stored_event_at`, `post_created_at`, `media_type`, `entity_type`, `engagement`, `like_count`, `comments_count`, `saved`, `reach`, `impressions`, `views`, `shares`, `hashtags`, `reels_avg_watch_time`, `reels_total_watch_time`, `replies`, `exits`, `taps_forward`, `taps_back`, `permalink`, `caption`, `media_url` |
| `instagram_insights` | `instagram_id`, `record_id`, `stored_event_at`, `created_time`, `online_users_datetime`, `followers_count`, `follows_count`, `profile_views`, `engagement`, `impressions`, `reach`, `accounts_engaged`, `online_followers`, `day_of_week`, `audience_age`, `audience_gender`, `audience_city`, `audience_country` |

### LinkedIn Tables

| Table | Key Columns |
|-------|-------------|
| `linkedin_posts` | `post_id`, `linkedin_id`, `saving_time`, `published_at`, `created_at`, `media_type`, `day_of_week`, `favorites`, `comments`, `repost`, `post_clicks`, `total_engagement`, `impressions`, `reach`, `hashtags`, `title`, `image`, `article_url` |
| `linkedin_insights` | `linkedin_id`, `record_id`, `inserted_at`, `created_at`, `totalFollowerCount`, `organicFollowerCount`, `paidFollowerCount`, `page_views`, `desktop_page_views`, `mobile_page_views`, `impressionCount`, `engagement`, `reach`, `repost`, `comments`, `reactions`, `unique_visitors`, `followers_by_seniority`, `followers_by_industry`, `followers_by_country`, `followers_by_city` |

### YouTube Tables

| Table | Key Columns |
|-------|-------------|
| `youtube_activity_insights` | `record_id`, `channel_id`, `created_at`, `estimated_minutes_watched`, `average_view_duration`, `views`, `likes`, `dislikes`, `comments`, `shares` |
| `youtube_channels` | `record_id`, `channel_id`, `subscriber_count`, `inserted_at`, `created_at`, `title` |
| `youtube_videos` | `video_id`, `channel_id`, `published_at`, `inserted_at`, `title`, `description`, `duration`, `thumbnail_url`, `media_type`, `iframe_embed_html`, `likes`, `dislikes`, `views`, `red_views`, `favorites`, `comments`, `subscribers_gained`, `shares`, `minutes_watched`, `red_minutes_watched`, `average_view_duration`, `average_view_percentage` |
| `youtube_traffic_insights` | `record_id`, `channel_id`, `created_at`, `subscriber_views`, `subscriber_watch_time`, `non_subsciber_watch_time` (typo), `paid_views`, `annotation_views`, `end_screen_views`, `campaign_card_view`, `no_link_other_views`, `yt_channel_views`, `yt_search_views`, `related_video_views`, `yt_other_page_views`, `ext_url_views`, `playlist_views`, `notification_views`, `shorts_views` |
| `youtube_shared_insights` | `channel_id`, `inserted_at`, and 31 sharing platform columns (ameba, blogger, copy_paste, cyworld, digg, dropbox, embed, mail, whats_app, other, facebook_messenger, facebook_pages, facebook, fotka, vkontakte, google_plus, discord, linkedin, goo, hangouts, pinterest, myspace, reddit, skype, telegram, tumblr, twitter, viber, weibo, wechat, youtube) |

### TikTok Tables

| Table | Key Columns |
|-------|-------------|
| `tiktok_posts` | `tiktok_id`, `post_id`, `display_name`, `like_count`, `comments_count`, `share_count`, `engagement_count`, `engagement_rate`, `view_count`, `created_at`, `inserted_at`, `profile_link`, `cover_image_url`, `share_url`, `post_description`, `hashtags`, `duration`, `height`, `width`, `title`, `embed_html`, `embed_link` |
| `tiktok_insights` | `tiktok_id`, `record_id`, `display_name`, `total_follower_count`, `total_following_count`, `total_video_views`, `total_video_likes`, `total_video_comments`, `total_video_shares`, `inserted_at` |

### Pinterest Tables

| Table | Key Columns |
|-------|-------------|
| `pinterest_pins` | `pin_id`, `board_id`, `user_id`, `created_at`, `media_type`, `is_owner`, `title`, `description`, `board_owner`, `cover_image_url`, `dominant_color`, `creative_type`, `product_tags`, `height`, `width` |
| `pinterest_pin_insights` | `pin_id`, `record_id`, `user_id`, `created_at`, `saving_time`, `inserted_at`, `impression`, `pin_clicks`, `outbound_click`, `saves`, `engagement`, `quartile_95s_percent_view`, `closeup`, `video_start`, `video_10s_view`, `video_avg_watch_time` |
| `pinterest_user_insights` | `user_id`, `created_at`, `impression`, `pin_clicks`, `outbound_click`, `saves`, `engagement` |
| `pinterest_users` | `user_id`, `inserted_at`, `follower_count` |
| `pinterest_boards` | `board_id`, `user_id`, `inserted_at`, `follower_count`, `name` |

### Twitter/X Tables

| Table | Key Columns |
|-------|-------------|
| `twitter_posts` | `twitter_id`, `post_id`, `created_at`, `inserted_at`, `tweet_type`, `total_engagement`, `impressions`, `like_count`, `reply_count`, `retweet_count`, `quote_count`, `bookmark_count`, `url_link_clicks`, `user_profile_clicks`, `impression_count`, `hashtags`, `permalink`, `full_text` |
| `twitter_insights` | `twitter_id`, `record_id`, `inserted_at`, `followers_count`, `following_count`, `tweet_count`, `listed_count` |

### Cross-Platform Views

| Table | Key Columns |
|-------|-------------|
| `mv_social_daily_metrics` | Materialized view. `date`, `platform`, `account_id`, `posts_count` (AggregateFunction(uniq)), `engagement_sum` (AggregateFunction(sum)), `impressions_sum` (AggregateFunction(sum)), `reach_sum` (AggregateFunction(sum)). Read with `uniqMerge(posts_count)`, `sumMerge(engagement_sum)`, etc. |

---

## 3. Facebook Analytics

### 3.1 Request Parameters

| Parameter | Type | Default | Notes |
|-----------|------|---------|-------|
| `workspace_id` | string | required | |
| `date` | string | required | `"YYYY-MM-DD - YYYY-MM-DD"` |
| `facebook_id` | string/array | required | Single ID or array |
| `timezone` | string | required | IANA timezone |
| `media_type` | array | `['text','link','images','videos','carousel','share','reels','others']` | |
| `limit` | int | 15 | Top posts limit |
| `order_by` | string | `'total_engagement'` | Sort column |

### 3.2 Date Filter Helper

Two modes:

```sql
-- posts mode (insights=false): converts column to user timezone
toDateTime({date_column}, 0, '{timezone}') BETWEEN toDateTime('{startDate}',0) AND toDateTime('{endDate+1day}',0)

-- insights mode (insights=true): plain date comparison
toDate({date_column}) BETWEEN toDate('{startDate}') AND toDate('{endDate}')
```

**CRITICAL side-effect:** After building the filter, `currentEndDate` is mutated via `subDay()`. This means `getDateFilters` must only be called once per query or the date window shrinks.

### 3.3 Endpoints and SQL Queries

#### 3.3.1 Overview Summary (`getSummaryQuery`)

**Route:** `POST overview/facebook/summary`

**Tables:** `facebook_posts`, `facebook_insights`

**Response keys:** `overview.current`, `overview.previous`

```sql
WITH posts as (
    SELECT post_id, max(saving_time)
    FROM facebook_posts
    WHERE page_id in {facebookId}
      AND toDateTime(created_time, 0, '{timezone}') BETWEEN toDateTime('{startDate}',0) AND toDateTime('{endDate+1day}',0)
    group by post_id
)
SELECT *
from (
    SELECT *
    from(
        SELECT c1 as page_id
        FROM VALUES {facebookId}
    ) as page_ids
    LEFT JOIN (
      SELECT toInt32(count()) as doc_count,
        toUInt64(page_id) as page_id,
        toInt32(reactions + comments + repost + posts_clicks) as total_engagement,
        toInt32(sum(total)) as reactions,
        toInt32(sum(comments)) as comments,
        toInt32(sum(post_clicks)) as posts_clicks,
        toInt32(sum(post_impressions)) as impressions,
        toInt32(sum(post_impressions_unique)) as reach,
        toInt32(sum(shares)) as repost
      FROM facebook_posts
      WHERE (post_id, saving_time) IN (posts)
      group by page_id
    ) AS posts_summary using page_id
) as posts_data
LEFT JOIN (
  SELECT toUInt64(page_id) as page_id,
    toInt32(sum(positive_sentiment)) as positive_sentiment,
    toInt32(sum(negative_sentiment)) as negative_sentiment,
    toInt32(sum(page_impressions)) as page_impressions,
    toInt32(sum(page_impressions_paid)) as page_impressions_paid,
    toInt32(sum(page_impressions_organic)) as page_impressions_organic,
    toInt32(sum(page_engagements)) as page_engagements,
    toInt32(sum(page_positive_feedback)) as page_positive_feedback,
    toInt32(sum(page_negative_feedback)) as page_negative_feedback,
    toInt32(max(fan_count)) as fan_count,
    toInt32(sum(talking_about_count)) as talking_about_count,
    toInt32(max(page_follows)) as page_follows
  FROM (
      SELECT page_id,
        toDate(created_time) as created_date,
        argMin(page_fans, saving_time) as fan_count,
        max(page_impressions) as page_impressions,
        max(page_impressions_paid) as page_impressions_paid,
        max(page_impressions_organic) as page_impressions_organic,
        max(page_post_engagements) as page_engagements,
        max(positive_sentiment) as positive_sentiment,
        max(negative_sentiment) as negative_sentiment,
        max(page_positive_feedback) as page_positive_feedback,
        max(page_negative_feedback) as page_negative_feedback,
        max(talking_about_count) as talking_about_count,
        argMin(page_follows, saving_time) as page_follows
      FROM facebook_insights
      WHERE page_id in {facebookId}
            AND toDate(created_time) BETWEEN toDate('{startDate}') AND toDate('{endDate}')
      GROUP BY page_id, created_date
  )
  group by page_id
) AS insights_summary USING page_id
```

**Output columns:** `doc_count`, `total_engagement`, `reactions`, `comments`, `posts_clicks`, `impressions`, `reach`, `repost`, `positive_sentiment`, `negative_sentiment`, `page_impressions`, `page_impressions_paid`, `page_impressions_organic`, `page_engagements`, `fan_count`, `talking_about_count`, `page_follows`

---

#### 3.3.2 Last Follower Counts (Fallback) (`getLastFollowerCounts`)

**Table:** `facebook_insights`

Used when fan_count time-series leads with zeros. Looks back 2 years.

```sql
SELECT
    arrayFirst(x -> x != 0, groupArray(fans)) AS page_fans,
    arrayFirst(x -> x != 0, groupArray(page_fans_by_like)) AS page_fans_by_like,
    arrayFirst(x -> x != 0, groupArray(page_fans_by_unlike)) AS page_fans_by_unlike
FROM (
    SELECT
        last_value(created_time) as inserted_time,
        toInt32(last_value(page_fans)) as fans,
        toInt32(last_value(page_fans_by_like)) as page_fans_by_like,
        toInt32(last_value(page_fans_by_unlike)) as page_fans_by_unlike
    FROM facebook_insights
    WHERE page_id in {facebookId}
        AND toDate(created_time) BETWEEN toDate('{startDate-2years}') AND toDate('{startDate}')
        AND page_fans != 0
    GROUP BY hash_id
    ORDER BY inserted_time DESC
)
```

---

#### 3.3.3 Audience Growth Time-Series (`getOverviewAudienceGrowthQuery`)

**Route:** `POST overview/facebook/audienceGrowth`

**Table:** `facebook_insights`

**Response keys:** `audience_growth` (time-series), `audience_growth_rollup.current`, `audience_growth_rollup.previous`

```sql
SELECT notEmpty(fan_count_temp) as show_data,
    arrayFill(x -> not x==0, fan_count_temp) as fan_count,
    page_fans_daily,
    page_fans_by_like_temp as page_fans_by_like,
    page_fans_by_unlike_temp as page_fans_by_unlike,
    page_impressions as page_impressions,
    page_engagements as page_engagements,
    date as buckets
FROM (
    SELECT groupArray(page_fans_total) as fan_count_temp,
            groupArray(page_fans_daily) as page_fans_daily,
            groupArray(page_fans_by_like) as page_fans_by_like_temp,
            groupArray(page_fans_by_unlike) as page_fans_by_unlike_temp,
            groupArray(page_impressions) as page_impressions,
            groupArray(page_engagements) as page_engagements,
            groupArray(created_date) as date
    from(
        SELECT toInt32(page_fans_total) as page_fans_total,
                toInt32(if(page_fans_total > 0, page_fans_total - lagInFrame(page_fans_total, 1, page_fans_total) OVER (ORDER BY created_date ASC), 0)) as page_fans_daily,
                toInt32(page_fans_by_like) as page_fans_by_like,
                toInt32(page_fans_by_unlike) as page_fans_by_unlike,
                toInt32(page_impressions) as page_impressions,
                toInt32(page_engagements) as page_engagements,
                created_date
        FROM (
            SELECT argMin(page_follows, saving_time) as page_fans_total,
                    argMin(page_fans_by_like, saving_time) as page_fans_by_like,
                    argMin(page_fans_by_unlike, saving_time) as page_fans_by_unlike,
                    max(page_impressions) as page_impressions,
                    max(page_post_engagements) as page_engagements,
                    toDate(created_time) as created_date
            FROM facebook_insights
            WHERE page_id in {facebookId} AND {dateFilter}
            GROUP BY created_date
            ORDER BY created_date ASC
            WITH FILL FROM toDate('{startDate}') TO toDate('{endDate}') + 1 STEP 1
        )
    )
)
```

**Key logic:**
- `argMin(page_follows, saving_time)` takes earliest daily record for fan count
- `lagInFrame(...) OVER (ORDER BY created_date)` computes daily net change
- `arrayFill(x -> not x==0, ...)` forward-fills zeros in the fan_count array
- `WITH FILL` zero-fills missing dates

**Dynamic version:** Switches daily/monthly at 60-day threshold. Returns additional `aggregation_level` field.

**Rollup query:** Returns `avg_page_fans_by_like`, `avg_page_fans_by_unlike`, `fan_count`, `talking_about_count`, `doc_count`, `page_id`.

---

#### 3.3.4 Publishing Behaviour (`getOverviewPublishingBehaviourByMediaTypeQuery`)

**Route:** `POST overview/facebook/publishingBehaviour`

**Table:** `facebook_posts`

```sql
WITH posts as (
    SELECT post_id, max(saving_time) as saving_time
    FROM facebook_posts
    WHERE page_id in {facebookId}
        AND media_type in ({media_types})
        AND {dateFilter_created_time}
    GROUP BY post_id
)
SELECT
    groupArray(reactions) as reactions_engagement,
    groupArray(comments) as comments_engagement,
    groupArray(shares) as shares_engagement,
    groupArray(paid_impressions) as paid_impressions,
    groupArray(organic_impressions) as organic_impressions,
    groupArray(viral_impressions) as viral_impressions,
    groupArray(paid_reach) as paid_reach,
    groupArray(organic_reach) as organic_reach,
    groupArray(viral_reach) as viral_reach,
    groupArray(created_date) as buckets,
    groupArray(post_count) as post_count
FROM (
    SELECT
        count() as post_count,
        sum(total) as reactions,
        sum(comments) as comments,
        sum(shares) as shares,
        sum(post_impressions_paid) as paid_impressions,
        sum(post_impressions_organic) as organic_impressions,
        sum(post_impressions_viral) as viral_impressions,
        sum(post_impressions_paid_unique) as paid_reach,
        sum(post_impressions_organic_unique) as organic_reach,
        sum(post_impressions_viral_unique) as viral_reach,
        toDate(created_time) as created_date
    FROM facebook_posts
    WHERE (post_id, saving_time) IN (posts)
    GROUP BY created_date
    ORDER BY created_date ASC
    WITH FILL TO toDate('{endDate+1day}')
)
```

---

#### 3.3.5 Top Posts (`getTop15PostsQuery`)

**Route:** `POST overview/facebook/getTopPosts`

**Tables:** `facebook_posts`, `facebook_media_assets`

```sql
WITH posts as (
    SELECT post_id, max(saving_time)
    FROM facebook_posts
    WHERE page_id in {facebookId}
    AND media_type in ({media_types})
        AND {dateFilter_created_time}
    group by post_id
)
SELECT *
FROM (
    SELECT *, 'top_posts' as post_category
    FROM (
        SELECT *
        FROM facebook_posts
        WHERE (post_id, saving_time) IN (posts)
        ORDER BY {order_by} desc, created_time
        LIMIT {limit} BY page_id
    ) as post_data
    LEFT JOIN (
        SELECT *
        FROM facebook_media_assets
        WHERE page_id in {facebookId}
            AND {dateFilter_created_at}
        order by inserted_at
    ) as media_assets using post_id
)
```

**Key:** `LIMIT {limit} BY page_id` returns top N posts per page for multi-account queries. Multiple rows per post_id from the media_assets JOIN are collapsed in the controller.

---

#### 3.3.6 Active Users by Hour (`getOverviewActiveUsersQuery`)

**Route:** `POST overview/facebook/activeUsers`

**Table:** `facebook_insights` (column: `page_fans_online`)

```sql
SELECT max(buckets) AS buckets,
    max(value) as values,
    arrayMax(values) as highest_value,
    buckets[indexOf(values, highest_value)] as highest_hour
FROM (
    SELECT max(buckets) AS buckets,
            arrayMap((x)->toInt32(x)/count(), sumForEach(values)) AS value
        FROM (
            SELECT arrayMap((x)->x[1], active_users_per_hour) AS buckets,
                    arrayMap((x)->toInt32(x[2]), active_users_per_hour) AS values
                FROM (
                    SELECT arrayMap((x)->(splitByChar('$', x)), arr) AS active_users_per_hour
                    FROM (
                        SELECT max(page_fans_online) AS arr
                        FROM facebook_insights
                        WHERE page_id in {facebookId}
                        AND toDate(created_time) BETWEEN toDate('{startDate}') AND toDate('{endDate}')
                        GROUP BY hash_id
                    )
                ))
)
```

**Data format:** `page_fans_online` is stored as a `$`-delimited packed string: `"0$123$1$456$..."` where pairs are `[hour, value]`.

**Controller timezone adjustment:**
```
timezoneInterval = round(Carbon::now(timezone)->offsetHours) + 8
```
The +8 accounts for data being stored in UTC-8. Buckets wrap around 0-23 after shifting.

---

#### 3.3.7 Active Users by Day (`getOverviewActiveUsersPerDayQuery`)

**Table:** `facebook_insights`

```sql
SELECT groupArray(day_name) as buckets,
        groupArray(active_users) as values,
        max(active_users) as highest_value,
        buckets[indexOf(values, highest_value)] as highest_day
FROM (
    SELECT ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'][day_num] as day_name,
            ifNull(active_users, 0) as active_users
    FROM (
        SELECT day_num, toInt32(count()) as active_users
        FROM (
            SELECT
                indexOf(['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'], last_value(day_of_week)) as day_num,
                last_value(created_time) as inserted_at
            FROM facebook_insights
            WHERE page_id in {facebookId}
            AND toDate(created_time) BETWEEN toDate('{startDate}') AND toDate('{endDate}')
            GROUP BY hash_id
        )
        GROUP BY day_num
        ORDER BY day_num ASC
        WITH FILL FROM 1 TO 8 STEP 1
    )
)
```

---

#### 3.3.8 Page Impressions (`getOverviewImpressionsQuery`)

**Route:** `POST overview/facebook/pageImpressions`

**Table:** `facebook_insights`

```sql
SELECT
    groupArray(page_impressions) as page_impressions,
    groupArray(created_date) as buckets
FROM (
    SELECT toInt32(max(page_impressions)) as page_impressions,
        toDate(created_time) as created_date
    FROM facebook_insights
    WHERE page_id = '{facebookId}'
    AND {dateFilter}
    GROUP BY created_date
    ORDER BY created_date ASC
    WITH FILL FROM toDate('{startDate}') TO toDate('{endDate}') + 1 STEP 1
)
```

**Rollup:** `total_impressions`, `avg_impressions_per_day` (sum/totalDays), `avg_impressions_per_week` (sum/totalWeeks).

---

#### 3.3.9 Page Engagements (`getOverviewEngagementsQuery`)

**Route:** `POST overview/facebook/engagement`

**Table:** `facebook_insights`

```sql
SELECT
    groupArray(page_engagements) as page_engagements,
    groupArray(created_date) as buckets
FROM (
    SELECT toInt32(max(page_post_engagements)) as page_engagements,
        toDate(created_time) as created_date
    FROM facebook_insights
    WHERE page_id = '{facebookId}'
    AND {dateFilter}
    GROUP BY created_date
    ORDER BY created_date ASC
    WITH FILL FROM toDate('{startDate}') TO toDate('{endDate}') + 1 STEP 1
)
```

---

#### 3.3.10 Reels Analytics (`getOverviewReelsAnalyticsQuery`)

**Route:** `POST overview/facebook/reelsAnalytics`

**Tables:** `facebook_posts`, `facebook_reels_insights`

```sql
WITH facebook_post_data AS (
    SELECT post_id,
         last_value(total) as reactions,
         last_value(comments) as comments,
         last_value(shares) as repost,
         reactions+comments+repost as total_engagement,
         toDate(created_time) as created_at
    FROM facebook_posts
    WHERE page_id in ({facebookId}) AND {dateFilter_created_at}
    GROUP BY post_id, created_at
)
SELECT
    groupArray(created_at) as buckets,
    groupArray(total_reels_count) as total_reels,
    groupArray(total_seconds_watched) as total_seconds_watched,
    groupArray(initial_plays) as initial_plays,
    groupArray(total_engagement) as engagement,
    groupArray(reactions) as reactions,
    groupArray(comments) as comments,
    groupArray(shares) as shares,
    toInt32(sum(total_reels_count)) as show_data
FROM (
    SELECT created_at,
        toInt32(count()) as total_reels_count,
        round(sum(total_time_watched_in_ms) / 1000, 2) as total_seconds_watched,
        toInt32(sum(play_count)) as initial_plays,
        toInt32(sum(total_engagement)) as total_engagement,
        toInt32(sum(reactions)) as reactions,
        toInt32(sum(comments)) as comments,
        toInt32(sum(repost)) as shares
    FROM (
        SELECT post_id, toDate(created_at) as created_at,
            last_value(average_time_watched) as average_time_watched,
            last_value(total_time_watched_in_ms) as total_time_watched_in_ms,
            last_value(play_count) as play_count,
            last_value(impressions_unique) as reach
        FROM facebook_reels_insights
        WHERE page_id in ({facebookId}) AND {dateFilter_created_at}
        GROUP BY post_id, created_at
    ) AS reels
    LEFT JOIN facebook_post_data ON facebook_post_data.post_id = reels.post_id
    GROUP BY reels.created_at
    ORDER BY created_at ASC
    WITH FILL TO (toDate('{endDate+1day}'))
)
```

**Note:** `total_time_watched_in_ms` is divided by 1000 for seconds.

---

#### 3.3.11 Video Insights (`getVideoInsightsQuery`)

**Route:** `POST overview/facebook/videoInsights`

**Tables:** `facebook_posts`, `facebook_video_insights`

Joins `facebook_posts` (filtered to `media_type='videos'`) with `facebook_video_insights` on `post_id`. View times are in ms, divided by 1000 for seconds.

---

#### 3.3.12 Time Recommendation (`getTimeRecommendationQuery`)

**Table:** `facebook_posts`

```sql
SELECT
    max(page_id) as facebookId,
    day_of_week,
    hour_of_day,
    sum(post_impressions) as post_impressions,
    sum(total_engagement) as total_engagement
FROM facebook_posts
WHERE page_id in {facebookId} AND {dateFilter}
GROUP BY day_of_week, hour_of_day
ORDER BY day_of_week, hour_of_day
```

---

#### 3.3.13 Demographics

**Audience Gender:** Reads `page_fans_gender` blob from `facebook_insights`. Parsed from `$`-delimited pairs. Known quirk: M/U/F labels mapped to wrong aliases.

**Audience Age:** Reads `page_fans_age` from most recent record. Age buckets: `65+`, `55-64`, `45-54`, `35-44`, `25-34`, `18-34`, `13-17`. Date boundary: if range crosses 2024-03-14, end date is clamped.

**Audience Country/City:** Reads `page_fans_country`/`page_fans_city` from most recent record. `$`-delimited parsing.

**Max Gender+Age:** Reads `page_fans_gender_age` blob. Finds the single bucket with highest value.

---

## 4. Facebook Competitor Analytics

### 4.1 Constructor and Timezone Handling

**CRITICAL difference from own analytics:** Competitor dates are parsed in user timezone, converted to UTC before embedding in SQL. Database timestamps are stored in UTC.

```
startDate = Carbon::parse(date[0], timezone)->setTimezone('UTC')
endDate = Carbon::parse(date[1], timezone)->setTime(23,59,59)->setTimezone('UTC')
```

### 4.2 Tables

`facebook_competitor_posts`, `facebook_competitor_insights`, `facebook_competitor_media_assets`

### 4.3 Helper Methods

#### `getPageFilters($insightsFlag, $filterAdded)`
Builds `facebook_id IN ('id1','id2',...)` for posts or `page_id IN (...)` for insights. If `$filterAdded = true`, excludes accounts with `state == "Added"`.

#### `getDateFilters($insightsFlag, $dateColumn)`
```sql
-- posts: created_at BETWEEN toDateTime('{startUTC}',0) AND toDateTime('{endUTC}',0)
-- insights: inserted_at BETWEEN toDateTime('{startUTC}',0) AND toDateTime('{endUTC}',0)
```

#### `getConstantConditions($field, $idField)`
Builds `multiIf(id='x','val', id='y','val', 'Random')` to map competitor IDs to metadata (name, image, state, slug).

### 4.4 Controller Methods

| Method | Description | Sort Default |
|--------|-------------|--------------|
| `searchCompetitor` | Graph API page search with appsecret_proof | N/A |
| `dataTableMetrics` | Current/previous period KPIs with growth % | `followersCount` |
| `postingActivityGraphByTypes` | Aggregate by media_type | `avgTotalEngagements` |
| `postingActivityBySpecificType` | Per-competitor for one media_type | N/A |
| `topAndLeastPerformingPosts` | Top-5/bottom-5 per competitor with media assets | engagement DESC/ASC |
| `topHashtags` | Top hashtags by engagement (limit default: 7) | count DESC |
| `individualHashtagData` | Per-competitor stats for one hashtag | engagement DESC |
| `biographyData` | Biography + length per competitor | `biography_length` |
| `followersGrowthComparison` | Per-competitor followers time-series | `followers_count` |
| `postReactDistribution` | Engagement totals for single page | N/A |
| `postReactDistributionByCompany` | Per-reaction breakdown for single page | N/A |
| `postTypeDistribution` | Per-media-type breakdown per competitor | N/A |
| `postEngagementOverTime` | Daily time-series per competitor (single page) | N/A |
| `postEngagementByCompetitor` | Total engagement per competitor | engagement DESC |

### 4.5 SQL Queries

#### 4.5.1 `getDataTableDataMetricsQuery($sortOrder)` -- Main Data Table

**Tables:** `facebook_competitor_posts`, `facebook_competitor_insights`

```sql
WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM facebook_competitor_posts
    WHERE {pageFilters} AND {dateFilters}
    group by post_id
)
SELECT
    multiIf(facebook_id='x','imgUrl',...) as image,
    multiIf(facebook_id='x','name',...) as name,
    multiIf(facebook_id='x','state',...) as state,
    * FROM (
    SELECT pages_ids.facebook_id as facebook_id,
        week_metrics.averageEngagement,
        week_metrics.averagePostsPerWeek,
        week_metrics.engagementRate,
        days_metrics.dayOfWeek,
        days_metrics.hourOfDay,
        days_metrics.averagePostsPerDay,
        days_metrics.averagePostsPerDayEngagement
    FROM (SELECT CAST(c1 AS String) as facebook_id FROM VALUES {accountIds}) as pages_ids
    LEFT JOIN (
        SELECT facebook_id,
            round(avg(total_engagement), 2) as averageEngagement,
            if(dateDiff('week', ...) != 0,
               round(sum(posts_in_a_week)/dateDiff('week',...), 2), 0) as averagePostsPerWeek,
            round((sum(total_engagement)/sum(posts_in_a_week)) / max(followers_count) * 100, 2) as engagementRate
        FROM (
            SELECT CAST(facebook_id AS String) AS facebook_id,
                sum(post_engagement) as total_engagement,
                toStartOfWeek(created_at) as week,
                count() as posts_in_a_week,
                followers_count
            FROM facebook_competitor_posts
            WHERE (post_id, inserted_at) IN (posts)
            group by facebook_id, week, followers_count
            order by week ASC
            WITH FILL FROM toStartOfWeek(toDate('{startDate}'))
                       TO toStartOfWeek(toDate('{endDate}'))
                       STEP INTERVAL 1 WEEK
        )
        group by facebook_id
    ) as week_metrics ON ...
    LEFT JOIN (
        SELECT facebook_id,
            argMax(multiIf(day_of_week=1,'Monday',...), total_posts) as dayOfWeek,
            argMax(multiIf(hour_of_day=0,'12:00 AM',...), total_posts) as hourOfDay,
            if(dateDiff('day',...) != 0,
               round(sum(total_posts)/dateDiff('day',...), 2), 0) as averagePostsPerDay,
            avg(total_engagement) as averagePostsPerDayEngagement,
            argMax(hour_of_day, total_posts) as maximumPostsHour,
            argMax(hour_of_day, total_engagement) as mostEngagementHour
        from (
            select facebook_id, ...,
                sum(post_engagement) as total_engagement,
                toDayOfWeek(created_at) as day_of_week,
                toHour(created_at) as hour_of_day,
                uniq(post_id) as total_posts,
                toDate(created_at) as date_c
            from facebook_competitor_posts
            WHERE (post_id, inserted_at) IN (posts)
            group by facebook_id, day_of_week, hour_of_day, date_c
            order by date_c ASC
            WITH FILL FROM toDate('{startDate}') TO toDate('{endDate}') STEP 1
        )
        group by facebook_id
    ) as days_metrics ON ...
) as page_metrics
LEFT JOIN (
    SELECT max(followers_count) as followersCount,
           max(total_fan_count) as fanCount,
           page_id
    FROM facebook_competitor_insights
    WHERE {pageFilters_insights} AND {dateFilters_insights}
    group by page_id
) as page_metadata ON page_metrics.facebook_id = page_metadata.page_id
ORDER BY {sortOrder} DESC
```

**Returns per competitor:** `facebook_id`, `image`, `name`, `state`, `averageEngagement`, `averagePostsPerWeek`, `engagementRate`, `dayOfWeek` (best day), `hourOfDay` (best hour), `averagePostsPerDay`, `averagePostsPerDayEngagement`, `followersCount`, `fanCount`.

---

#### 4.5.2 `getPostingActivityGraphByTypesQuery($sortOrder)` -- Activity by Media Type (Aggregated)

**Table:** `facebook_competitor_posts`

```sql
WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM facebook_competitor_posts
    WHERE {pageFilters} AND {dateFilters}
    group by post_id
)
SELECT
    media_type as mediaType,
    round(sum(avg_engagement),2) as avgTotalEngagements,
    sum(page_post_count) as totalPosts,
    toInt32(sum(total_engagement)) as total_engagement,
    round(sum(er),2) as avgEngagementRate,
    if(dateDiff('week',...) = 0, 0, round(totalPosts/dateDiff('week',...),2)) as postsPerWeek,
    if(dateDiff('day',...) = 0, 0, round(totalPosts/dateDiff('day',...),2)) as postsPerDay,
    if(dateDiff('hour',...) = 0, 0, round(totalPosts/dateDiff('hour',...),2)) as postsPerHour,
    dateDiff('week',...) as weekCount,
    dateDiff('day',...) as dayCount,
    dateDiff('hour',...) as hourCount
FROM (
    SELECT facebook_id, page_name, media_type,
        sum(count) as page_post_count,
        sum(total_post_engagement) as total_engagement,
        if(page_post_count<=0,0, round(sum(total_post_engagement)/page_post_count,2)) as avg_engagement,
        if(page_post_count<=0,0, round(((sum(er)/page_post_count)/max(followers_count))*100,2)) as er
    FROM (
        SELECT facebook_id, page_name, media_type,
            count() as count,
            sum(post_engagement) as total_post_engagement,
            total_post_engagement as er,
            argMax(followers_count, created_at) as followers_count,
            toStartOfWeek(created_at) as week
        from facebook_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        group by facebook_id, page_name, media_type, week
        ORDER by week asc
        WITH FILL FROM toStartOfWeek(toDate('{startDate}'))
                   TO toStartOfWeek(toDate('{endDate}'))
                   STEP INTERVAL 1 WEEK
    )
    WHERE media_type!=''
    GROUP BY facebook_id, page_name, media_type
)
group by media_type
```

Aggregates **all competitors together** by `media_type`.

---

#### 4.5.3 `getPostingActivityBySpecificTypeQuery($mediaType, $sortOrder)`

Same structure as above but WHERE clause adds `media_type = '$mediaType'` and groups per `facebook_id`. Adds per-competitor metadata from `facebook_competitor_insights`. Returns one row per competitor for that media type.

---

#### 4.5.4 `getTopPerformingAndLeastPerformingPostsQuery()` -- Top+Least Posts

**Tables:** `facebook_competitor_posts`, `facebook_competitor_media_assets`

```sql
WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM facebook_competitor_posts
    WHERE {pageFilters} AND {dateFilters}
    group by post_id
)
SELECT * FROM (
    SELECT *, 'top_5_posts' as category,
        'https://graph.facebook.com/' || facebook_id || '/picture?type=large' as image
    FROM (
        SELECT * FROM facebook_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        ORDER BY post_engagement desc
        LIMIT 5 BY facebook_id
    ) as post_data
    LEFT JOIN (
        SELECT * FROM facebook_competitor_media_assets
        WHERE {pageFilters_insights} AND created_at BETWEEN ...
        order by inserted_at desc
    ) as media_assets using post_id
UNION ALL
    SELECT *, 'least_5_posts' as category, ...
    FROM (
        SELECT * FROM facebook_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        ORDER BY post_engagement asc
        LIMIT 5 BY facebook_id
    ) as post_data
    LEFT JOIN (...) as media_assets using post_id
)
```

**Key:** `LIMIT 5 BY facebook_id` -- returns top/bottom 5 per competitor.

---

#### 4.5.5 `getTopHashtagsQuery($limit)` -- Top Hashtags

**Tables:** `facebook_competitor_posts`, `facebook_competitor_insights`

```sql
WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM facebook_competitor_posts
    WHERE {pageFilters} AND {dateFilters}
    group by post_id
)
SELECT
    groupUniqArray(hashtags_business_ids) as companies_using,
    groupUniqArray(hashtags_name) as companies_name,
    sum(hashtag_count) as count,
    sum(hashtags_total_engagement) as total_engagement,
    groupUniqArray(followers_total_count) as total_followers,
    round(sum(engagement_per_follower),2) as engagement_per_follower,
    round(sum(engagement_rate_by_follower),2) as engagement_rate_by_follower,
    round(sum(engagement_per_post),2) as engagement_per_post,
    tag
from (
    SELECT
        hashtags.facebook_id as hashtags_business_ids,
        hashtags.name as hashtags_name,
        hashtags.count as hashtag_count,
        hashtags.total_engagement as hashtags_total_engagement,
        round((hashtags.total_engagement / hashtags.count),2) as engagement_per_post,
        followers.total_followers as followers_total_count,
        round((hashtags.total_engagement / followers.total_followers),2) as engagement_per_follower,
        round((hashtags.total_engagement / followers.total_followers / hashtags.count) * 100,2) as engagement_rate_by_follower,
        hashtags.tag as tag
    from (
        SELECT max(followers_count) as total_followers, page_id
        FROM facebook_competitor_insights
        WHERE page_id IN {accountIds}
        group by page_id
    ) as followers
    left join (
        SELECT page_name as name, arrayJoin(hashtags) as tag,
            uniq(post_id) as count, facebook_id as facebook_id,
            sum(post_engagement) as total_engagement
        from facebook_competitor_posts
        where length(hashtags) > 0 AND (post_id, inserted_at) IN (posts)
        group by tag, facebook_id, page_name
        order by count desc
    ) as hashtags on followers.page_id = hashtags.facebook_id
)
where tag != ''
group by tag
order by count desc
limit {limit}
```

**Key:** `arrayJoin(hashtags)` -- hashtags is a native ClickHouse array in `facebook_competitor_posts`.

---

#### 4.5.6 `getIndividualHashtagQuery($hashtag)` -- Single Hashtag Detail

**Tables:** `facebook_competitor_posts`, `facebook_competitor_insights`

```sql
WITH posts as (...)
SELECT * FROM (
    SELECT arrayJoin(hashtags) as tag,
        uniq(post_id) as count,
        sum(post_engagement) as total_engagement,
        max(followers_count) as total_followers,
        round((total_engagement / count), 2) as engagement_per_post,
        round((total_engagement / total_followers), 2) as engagement_per_follower,
        round((total_engagement / total_followers / count) * 100, 2) as engagement_rate_by_follower,
        facebook_id
    FROM facebook_competitor_posts
    where length(hashtags) > 0 AND tag = '{hashtag}' AND (post_id, inserted_at) IN (posts)
    group by tag, facebook_id
    order by total_engagement desc
) as hashtags_statistics
LEFT JOIN (
    select argMax(profile_picture_url, inserted_at) as image,
           argMax(page_name, inserted_at) as name,
           multiIf(...) as slug,
           argMax(followers_count, inserted_at) as followersCount,
           page_id
    from facebook_competitor_insights
    group by page_id
) as page_constants ON page_id = hashtags_statistics.facebook_id
```

---

#### 4.5.7 `getBiographyQuery($sortOrder)` -- Biography Data

**Tables:** `facebook_competitor_posts`, `facebook_competitor_insights`

```sql
SELECT * FROM (
    select last_value(biography) as biography,
           lengthUTF8(biography) as biography_length,
           facebook_id,
           multiIf(...) as state,
           multiIf(...) as slug
    from facebook_competitor_posts
    WHERE {pageFilters}
    group by facebook_id
    ORDER BY max(inserted_at) DESC
) as biography_statistics
LEFT JOIN (
    select argMax(profile_picture_url, inserted_at) as image,
           argMax(page_name, inserted_at) as name,
           argMax(followers_count, inserted_at) as followersCount,
           page_id
    from facebook_competitor_insights
    group by page_id
) as page_constants ON page_id = biography_statistics.facebook_id
ORDER BY {sortOrder} DESC
```

**Note:** No date filter on biography -- always shows latest value.

---

#### 4.5.8 `getFollowersGrowthComparisonQuery($sortOrder)` -- Followers Growth Over Time

**Table:** `facebook_competitor_insights`

```sql
SELECT facebook_id,
    multiIf(...) as name,
    multiIf(...) as image,
    multiIf(...) as state,
    multiIf(...) as slug,
    dates, followers_count, dates_with_followers_count
FROM (SELECT c1 as facebook_id FROM VALUES {accountIds}) as page_ids
LEFT JOIN (
    SELECT facebook_id,
        groupArray(date) as dates,
        groupArray(followers_count) as followers_count,
        arrayZip(dates, followers_count) as dates_with_followers_count
    FROM (
        select page_name, profile_picture_url,
               page_id as facebook_id,
               toDate(inserted_at) as date,
               toInt32(followers_count) as followers_count
        from facebook_competitor_insights
        WHERE {pageFilters_insights} AND {dateFilters_insights}
        order by date ASC
    )
    group by facebook_id
) as page_insights on page_ids.facebook_id = page_insights.facebook_id
ORDER BY {sortOrder} ASC
```

Returns `dates_with_followers_count` as a zipped array of `[date, followers_count]` tuples.

---

#### 4.5.9 `getPostReactDistributionByCompany($sortOrder, $facebook_id)` -- Per-Reaction Breakdown

**Table:** `facebook_competitor_posts`

```sql
SELECT facebook_id, page_name,
    'https://graph.facebook.com/' || facebook_id || '/picture' as image,
    toInt32(sum(like)) as total_likes,
    toInt32(sum(haha)) as total_hahas,
    toInt32(sum(angry)) as total_angry,
    toInt32(sum(sad)) as total_sad,
    toInt32(sum(thankful)) as total_thankful,
    toInt32(sum(love)) as total_love,
    toInt32(sum(wow)) as total_wow,
    toInt32(sum(total_post_reactions)) as total_post_reactions,
    toInt32(sum(comments)) as comments,
    toInt32(sum(shares)) as shares,
    toInt32(count()) as total_posts
FROM facebook_competitor_posts
WHERE facebook_id = '{facebook_id}' AND {dateFilters}
GROUP BY facebook_id, page_name
```

**Note:** No deduplication CTE -- reads all rows directly.

---

#### 4.5.10 `getPostEngagementOverTime($sortOrder, $facebook_id)` -- Daily Engagement Time Series

**Table:** `facebook_competitor_posts`

```sql
SELECT
    toInt32(sum(post_engagement)) as total_engagements,
    toInt32(count()) as total_posts,
    toDate(created_at) as date
FROM facebook_competitor_posts
WHERE facebook_id = '{facebook_id}' AND {dateFilters}
GROUP BY toDate(created_at)
ORDER BY toDate(created_at) ASC
WITH FILL FROM toDate('{startDate}') TO toDate('{endDate}') STEP 1
```

---

#### 4.5.11 `getPostEngagementByCompetitor($sortOrder)` -- Total Engagement Per Competitor

**Table:** `facebook_competitor_posts`

```sql
select facebook_id, page_name,
       'https://graph.facebook.com/' || facebook_id || '/picture' as image,
       toInt32(count()) as total_posts,
       toInt32(sum(post_engagement)) as total_engagements
from (
    select facebook_id, page_name, post_id,
           max(post_engagement) as post_engagement
    FROM facebook_competitor_posts
    WHERE {pageFilters} AND {dateFilters}
    group by facebook_id, page_name, post_id
)
group by facebook_id, page_name
ORDER BY total_engagements DESC
```

Deduplication is done inline: `max(post_engagement)` per `post_id` before aggregating.

---

## 5. Instagram Analytics

### 5.1 Request Parameters

Same as Facebook but uses `instagram_id` instead of `facebook_id`. Media types default to `['REELS','IMAGE','VIDEO','CAROUSEL_ALBUM']`.

### 5.2 Endpoints and SQL Queries

#### 5.2.1 Summary (`summaryQuery`)

**Route:** `POST overview/instagram/summary`

**Tables:** `instagram_posts`, `instagram_insights`

```sql
WITH posts AS (
    SELECT media_id, max(stored_event_at)
    FROM instagram_posts
    WHERE instagram_id in {instagram_id}
      AND toDateTime(post_created_at,0,'{timezone}') BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
    GROUP BY media_id
)
SELECT
    toInt32(sum(total_engagement))   as post_engagement,
    toInt32(sum(post_views))         as post_views,
    toInt32(sum(total_reach))        as post_reach,
    toInt32(sum(total_saves))        as post_saves,
    toInt32(sum(reactions))          as post_reactions,
    toInt32(sum(comments))           as post_comments,
    toInt32(sum(profile_views))      as profile_views,
    toInt32(first_value(followers_count))  as followers_count,
    toInt32(first_value(follows_count))    as follows_count,
    toInt32(sum(accounts_engaged))   as accounts_engaged,
    toInt32(sum(engagement))         as profile_engagement,
    toInt32(sum(impressions))        as profile_impressions,
    toInt32(sum(reach))              as profile_reach,
    toInt32(sum(doc_count))          as doc_count,
    toInt32(sum(total_stories))      as total_stories,
    toInt32(sum(total_posts))        as total_posts,
    round(if(doc_count > 0, toFloat32(post_engagement/doc_count), 0), 2) as eng_rate
FROM (
    -- posts subquery with LEFT JOIN insights subquery
    -- Posts: count, SUM(comments_count), SUM(engagement), SUM(impressions), SUM(reach), SUM(saved), SUM(like_count), SUM(views)
    -- Stories counted via: countIf(entity_type = 'STORY' OR media_type = 'STORY')
    -- Insights: SUM(daily_profile_views), argMaxIf(daily_followers_count, date_bucket, daily_followers_count > 0), etc.
)
```

---

#### 5.2.2 Audience Growth (`audienceQuery`)
#### 3.3.8 Page Impressions (`getOverviewImpressionsQuery`)

**Route:** `POST overview/facebook/pageImpressions`

**Table:** `facebook_insights`

```sql
SELECT
    groupArray(page_impressions) as page_impressions,
    groupArray(created_date) as buckets
FROM (
    SELECT toInt32(max(page_impressions)) as page_impressions,
        toDate(created_time) as created_date
    FROM facebook_insights
    WHERE page_id = '{facebookId}'
    AND {dateFilter}
    GROUP BY created_date
    ORDER BY created_date ASC
    WITH FILL FROM toDate('{startDate}') TO toDate('{endDate}') + 1 STEP 1
)
```

**Rollup:** `total_impressions`, `avg_impressions_per_day` (sum/totalDays), `avg_impressions_per_week` (sum/totalWeeks).

---

#### 3.3.9 Page Engagements (`getOverviewEngagementsQuery`)

**Route:** `POST overview/facebook/engagement`

**Table:** `facebook_insights`

```sql
SELECT
    groupArray(page_engagements) as page_engagements,
    groupArray(created_date) as buckets
FROM (
    SELECT toInt32(max(page_post_engagements)) as page_engagements,
        toDate(created_time) as created_date
    FROM facebook_insights
    WHERE page_id = '{facebookId}'
    AND {dateFilter}
    GROUP BY created_date
    ORDER BY created_date ASC
    WITH FILL FROM toDate('{startDate}') TO toDate('{endDate}') + 1 STEP 1
)
```

---

#### 3.3.10 Reels Analytics (`getOverviewReelsAnalyticsQuery`)

**Route:** `POST overview/facebook/reelsAnalytics`

**Tables:** `facebook_posts`, `facebook_reels_insights`

```sql
WITH facebook_post_data AS (
    SELECT post_id,
         last_value(total) as reactions,
         last_value(comments) as comments,
         last_value(shares) as repost,
         reactions+comments+repost as total_engagement,
         toDate(created_time) as created_at
    FROM facebook_posts
    WHERE page_id in ({facebookId}) AND {dateFilter_created_at}
    GROUP BY post_id, created_at
)
SELECT
    groupArray(created_at) as buckets,
    groupArray(total_reels_count) as total_reels,
    groupArray(total_seconds_watched) as total_seconds_watched,
    groupArray(initial_plays) as initial_plays,
    groupArray(total_engagement) as engagement,
    groupArray(reactions) as reactions,
    groupArray(comments) as comments,
    groupArray(shares) as shares,
    toInt32(sum(total_reels_count)) as show_data
FROM (
    SELECT created_at,
        toInt32(count()) as total_reels_count,
        round(sum(total_time_watched_in_ms) / 1000, 2) as total_seconds_watched,
        toInt32(sum(play_count)) as initial_plays,
        toInt32(sum(total_engagement)) as total_engagement,
        toInt32(sum(reactions)) as reactions,
        toInt32(sum(comments)) as comments,
        toInt32(sum(repost)) as shares
    FROM (
        SELECT post_id, toDate(created_at) as created_at,
            last_value(average_time_watched) as average_time_watched,
            last_value(total_time_watched_in_ms) as total_time_watched_in_ms,
            last_value(play_count) as play_count,
            last_value(impressions_unique) as reach
        FROM facebook_reels_insights
        WHERE page_id in ({facebookId}) AND {dateFilter_created_at}
        GROUP BY post_id, created_at
    ) AS reels
    LEFT JOIN facebook_post_data ON facebook_post_data.post_id = reels.post_id
    GROUP BY reels.created_at
    ORDER BY created_at ASC
    WITH FILL TO (toDate('{endDate+1day}'))
)
```

**Note:** `total_time_watched_in_ms` is divided by 1000 for seconds.

---

#### 3.3.11 Video Insights (`getVideoInsightsQuery`)

**Route:** `POST overview/facebook/videoInsights`

**Tables:** `facebook_posts`, `facebook_video_insights`

Joins `facebook_posts` (filtered to `media_type='videos'`) with `facebook_video_insights` on `post_id`. View times are in ms, divided by 1000 for seconds.

---

#### 3.3.12 Time Recommendation (`getTimeRecommendationQuery`)

**Table:** `facebook_posts`

```sql
SELECT
    max(page_id) as facebookId,
    day_of_week,
    hour_of_day,
    sum(post_impressions) as post_impressions,
    sum(total_engagement) as total_engagement
FROM facebook_posts
WHERE page_id in {facebookId} AND {dateFilter}
GROUP BY day_of_week, hour_of_day
ORDER BY day_of_week, hour_of_day
```

---

#### 3.3.13 Demographics

**Audience Gender:** Reads `page_fans_gender` blob from `facebook_insights`. Parsed from `$`-delimited pairs. Known quirk: M/U/F labels mapped to wrong aliases.

**Audience Age:** Reads `page_fans_age` from most recent record. Age buckets: `65+`, `55-64`, `45-54`, `35-44`, `25-34`, `18-34`, `13-17`. Date boundary: if range crosses 2024-03-14, end date is clamped.

**Audience Country/City:** Reads `page_fans_country`/`page_fans_city` from most recent record. `$`-delimited parsing.

**Max Gender+Age:** Reads `page_fans_gender_age` blob. Finds the single bucket with highest value.

---

## 4. Facebook Competitor Analytics

### 4.1 Constructor and Timezone Handling

**CRITICAL difference from own analytics:** Competitor dates are parsed in user timezone, converted to UTC before embedding in SQL. Database timestamps are stored in UTC.

```
startDate = Carbon::parse(date[0], timezone)->setTimezone('UTC')
endDate = Carbon::parse(date[1], timezone)->setTime(23,59,59)->setTimezone('UTC')
```

### 4.2 Tables

`facebook_competitor_posts`, `facebook_competitor_insights`, `facebook_competitor_media_assets`

### 4.3 Helper Methods

#### `getPageFilters($insightsFlag, $filterAdded)`
Builds `facebook_id IN ('id1','id2',...)` for posts or `page_id IN (...)` for insights. If `$filterAdded = true`, excludes accounts with `state == "Added"`.

#### `getDateFilters($insightsFlag, $dateColumn)`
```sql
-- posts: created_at BETWEEN toDateTime('{startUTC}',0) AND toDateTime('{endUTC}',0)
-- insights: inserted_at BETWEEN toDateTime('{startUTC}',0) AND toDateTime('{endUTC}',0)
```

#### `getConstantConditions($field, $idField)`
Builds `multiIf(id='x','val', id='y','val', 'Random')` to map competitor IDs to metadata (name, image, state, slug).

### 4.4 Controller Methods

| Method | Description | Sort Default |
|--------|-------------|--------------|
| `searchCompetitor` | Graph API page search with appsecret_proof | N/A |
| `dataTableMetrics` | Current/previous period KPIs with growth % | `followersCount` |
| `postingActivityGraphByTypes` | Aggregate by media_type | `avgTotalEngagements` |
| `postingActivityBySpecificType` | Per-competitor for one media_type | N/A |
| `topAndLeastPerformingPosts` | Top-5/bottom-5 per competitor with media assets | engagement DESC/ASC |
| `topHashtags` | Top hashtags by engagement (limit default: 7) | count DESC |
| `individualHashtagData` | Per-competitor stats for one hashtag | engagement DESC |
| `biographyData` | Biography + length per competitor | `biography_length` |
| `followersGrowthComparison` | Per-competitor followers time-series | `followers_count` |
| `postReactDistribution` | Engagement totals for single page | N/A |
| `postReactDistributionByCompany` | Per-reaction breakdown for single page | N/A |
| `postTypeDistribution` | Per-media-type breakdown per competitor | N/A |
| `postEngagementOverTime` | Daily time-series per competitor (single page) | N/A |
| `postEngagementByCompetitor` | Total engagement per competitor | engagement DESC |

### 4.5 SQL Queries

#### 4.5.1 `getDataTableDataMetricsQuery($sortOrder)` -- Main Data Table

**Tables:** `facebook_competitor_posts`, `facebook_competitor_insights`

```sql
WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM facebook_competitor_posts
    WHERE {pageFilters} AND {dateFilters}
    group by post_id
)
SELECT
    multiIf(facebook_id='x','imgUrl',...) as image,
    multiIf(facebook_id='x','name',...) as name,
    multiIf(facebook_id='x','state',...) as state,
    * FROM (
    SELECT pages_ids.facebook_id as facebook_id,
        week_metrics.averageEngagement,
        week_metrics.averagePostsPerWeek,
        week_metrics.engagementRate,
        days_metrics.dayOfWeek,
        days_metrics.hourOfDay,
        days_metrics.averagePostsPerDay,
        days_metrics.averagePostsPerDayEngagement
    FROM (SELECT CAST(c1 AS String) as facebook_id FROM VALUES {accountIds}) as pages_ids
    LEFT JOIN (
        SELECT facebook_id,
            round(avg(total_engagement), 2) as averageEngagement,
            if(dateDiff('week', ...) != 0,
               round(sum(posts_in_a_week)/dateDiff('week',...), 2), 0) as averagePostsPerWeek,
            round((sum(total_engagement)/sum(posts_in_a_week)) / max(followers_count) * 100, 2) as engagementRate
        FROM (
            SELECT CAST(facebook_id AS String) AS facebook_id,
                sum(post_engagement) as total_engagement,
                toStartOfWeek(created_at) as week,
                count() as posts_in_a_week,
                followers_count
            FROM facebook_competitor_posts
            WHERE (post_id, inserted_at) IN (posts)
            group by facebook_id, week, followers_count
            order by week ASC
            WITH FILL FROM toStartOfWeek(toDate('{startDate}'))
                       TO toStartOfWeek(toDate('{endDate}'))
                       STEP INTERVAL 1 WEEK
        )
        group by facebook_id
    ) as week_metrics ON ...
    LEFT JOIN (
        SELECT facebook_id,
            argMax(multiIf(day_of_week=1,'Monday',...), total_posts) as dayOfWeek,
            argMax(multiIf(hour_of_day=0,'12:00 AM',...), total_posts) as hourOfDay,
            if(dateDiff('day',...) != 0,
               round(sum(total_posts)/dateDiff('day',...), 2), 0) as averagePostsPerDay,
            avg(total_engagement) as averagePostsPerDayEngagement,
            argMax(hour_of_day, total_posts) as maximumPostsHour,
            argMax(hour_of_day, total_engagement) as mostEngagementHour
        from (
            select facebook_id, ...,
                sum(post_engagement) as total_engagement,
                toDayOfWeek(created_at) as day_of_week,
                toHour(created_at) as hour_of_day,
                uniq(post_id) as total_posts,
                toDate(created_at) as date_c
            from facebook_competitor_posts
            WHERE (post_id, inserted_at) IN (posts)
            group by facebook_id, day_of_week, hour_of_day, date_c
            order by date_c ASC
            WITH FILL FROM toDate('{startDate}') TO toDate('{endDate}') STEP 1
        )
        group by facebook_id
    ) as days_metrics ON ...
) as page_metrics
LEFT JOIN (
    SELECT max(followers_count) as followersCount,
           max(total_fan_count) as fanCount,
           page_id
    FROM facebook_competitor_insights
    WHERE {pageFilters_insights} AND {dateFilters_insights}
    group by page_id
) as page_metadata ON page_metrics.facebook_id = page_metadata.page_id
ORDER BY {sortOrder} DESC
```

**Returns per competitor:** `facebook_id`, `image`, `name`, `state`, `averageEngagement`, `averagePostsPerWeek`, `engagementRate`, `dayOfWeek` (best day), `hourOfDay` (best hour), `averagePostsPerDay`, `averagePostsPerDayEngagement`, `followersCount`, `fanCount`.

---

#### 4.5.2 `getPostingActivityGraphByTypesQuery($sortOrder)` -- Activity by Media Type (Aggregated)

**Table:** `facebook_competitor_posts`

```sql
WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM facebook_competitor_posts
    WHERE {pageFilters} AND {dateFilters}
    group by post_id
)
SELECT
    media_type as mediaType,
    round(sum(avg_engagement),2) as avgTotalEngagements,
    sum(page_post_count) as totalPosts,
    toInt32(sum(total_engagement)) as total_engagement,
    round(sum(er),2) as avgEngagementRate,
    if(dateDiff('week',...) = 0, 0, round(totalPosts/dateDiff('week',...),2)) as postsPerWeek,
    if(dateDiff('day',...) = 0, 0, round(totalPosts/dateDiff('day',...),2)) as postsPerDay,
    if(dateDiff('hour',...) = 0, 0, round(totalPosts/dateDiff('hour',...),2)) as postsPerHour,
    dateDiff('week',...) as weekCount,
    dateDiff('day',...) as dayCount,
    dateDiff('hour',...) as hourCount
FROM (
    SELECT facebook_id, page_name, media_type,
        sum(count) as page_post_count,
        sum(total_post_engagement) as total_engagement,
        if(page_post_count<=0,0, round(sum(total_post_engagement)/page_post_count,2)) as avg_engagement,
        if(page_post_count<=0,0, round(((sum(er)/page_post_count)/max(followers_count))*100,2)) as er
    FROM (
        SELECT facebook_id, page_name, media_type,
            count() as count,
            sum(post_engagement) as total_post_engagement,
            total_post_engagement as er,
            argMax(followers_count, created_at) as followers_count,
            toStartOfWeek(created_at) as week
        from facebook_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        group by facebook_id, page_name, media_type, week
        ORDER by week asc
        WITH FILL FROM toStartOfWeek(toDate('{startDate}'))
                   TO toStartOfWeek(toDate('{endDate}'))
                   STEP INTERVAL 1 WEEK
    )
    WHERE media_type!=''
    GROUP BY facebook_id, page_name, media_type
)
group by media_type
```

Aggregates **all competitors together** by `media_type`.

---

#### 4.5.3 `getPostingActivityBySpecificTypeQuery($mediaType, $sortOrder)`

Same structure as above but WHERE clause adds `media_type = '$mediaType'` and groups per `facebook_id`. Adds per-competitor metadata from `facebook_competitor_insights`. Returns one row per competitor for that media type.

---

#### 4.5.4 `getTopPerformingAndLeastPerformingPostsQuery()` -- Top+Least Posts

**Tables:** `facebook_competitor_posts`, `facebook_competitor_media_assets`

```sql
WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM facebook_competitor_posts
    WHERE {pageFilters} AND {dateFilters}
    group by post_id
)
SELECT * FROM (
    SELECT *, 'top_5_posts' as category,
        'https://graph.facebook.com/' || facebook_id || '/picture?type=large' as image
    FROM (
        SELECT * FROM facebook_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        ORDER BY post_engagement desc
        LIMIT 5 BY facebook_id
    ) as post_data
    LEFT JOIN (
        SELECT * FROM facebook_competitor_media_assets
        WHERE {pageFilters_insights} AND created_at BETWEEN ...
        order by inserted_at desc
    ) as media_assets using post_id
UNION ALL
    SELECT *, 'least_5_posts' as category, ...
    FROM (
        SELECT * FROM facebook_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        ORDER BY post_engagement asc
        LIMIT 5 BY facebook_id
    ) as post_data
    LEFT JOIN (...) as media_assets using post_id
)
```

**Key:** `LIMIT 5 BY facebook_id` -- returns top/bottom 5 per competitor.

---

#### 4.5.5 `getTopHashtagsQuery($limit)` -- Top Hashtags

**Tables:** `facebook_competitor_posts`, `facebook_competitor_insights`

```sql
WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM facebook_competitor_posts
    WHERE {pageFilters} AND {dateFilters}
    group by post_id
)
SELECT
    groupUniqArray(hashtags_business_ids) as companies_using,
    groupUniqArray(hashtags_name) as companies_name,
    sum(hashtag_count) as count,
    sum(hashtags_total_engagement) as total_engagement,
    groupUniqArray(followers_total_count) as total_followers,
    round(sum(engagement_per_follower),2) as engagement_per_follower,
    round(sum(engagement_rate_by_follower),2) as engagement_rate_by_follower,
    round(sum(engagement_per_post),2) as engagement_per_post,
    tag
from (
    SELECT
        hashtags.facebook_id as hashtags_business_ids,
        hashtags.name as hashtags_name,
        hashtags.count as hashtag_count,
        hashtags.total_engagement as hashtags_total_engagement,
        round((hashtags.total_engagement / hashtags.count),2) as engagement_per_post,
        followers.total_followers as followers_total_count,
        round((hashtags.total_engagement / followers.total_followers),2) as engagement_per_follower,
        round((hashtags.total_engagement / followers.total_followers / hashtags.count) * 100,2) as engagement_rate_by_follower,
        hashtags.tag as tag
    from (
        SELECT max(followers_count) as total_followers, page_id
        FROM facebook_competitor_insights
        WHERE page_id IN {accountIds}
        group by page_id
    ) as followers
    left join (
        SELECT page_name as name, arrayJoin(hashtags) as tag,
            uniq(post_id) as count, facebook_id as facebook_id,
            sum(post_engagement) as total_engagement
        from facebook_competitor_posts
        where length(hashtags) > 0 AND (post_id, inserted_at) IN (posts)
        group by tag, facebook_id, page_name
        order by count desc
    ) as hashtags on followers.page_id = hashtags.facebook_id
)
where tag != ''
group by tag
order by count desc
limit {limit}
```

**Key:** `arrayJoin(hashtags)` -- hashtags is a native ClickHouse array in `facebook_competitor_posts`.

---

#### 4.5.6 `getIndividualHashtagQuery($hashtag)` -- Single Hashtag Detail

**Tables:** `facebook_competitor_posts`, `facebook_competitor_insights`

```sql
WITH posts as (...)
SELECT * FROM (
    SELECT arrayJoin(hashtags) as tag,
        uniq(post_id) as count,
        sum(post_engagement) as total_engagement,
        max(followers_count) as total_followers,
        round((total_engagement / count), 2) as engagement_per_post,
        round((total_engagement / total_followers), 2) as engagement_per_follower,
        round((total_engagement / total_followers / count) * 100, 2) as engagement_rate_by_follower,
        facebook_id
    FROM facebook_competitor_posts
    where length(hashtags) > 0 AND tag = '{hashtag}' AND (post_id, inserted_at) IN (posts)
    group by tag, facebook_id
    order by total_engagement desc
) as hashtags_statistics
LEFT JOIN (
    select argMax(profile_picture_url, inserted_at) as image,
           argMax(page_name, inserted_at) as name,
           multiIf(...) as slug,
           argMax(followers_count, inserted_at) as followersCount,
           page_id
    from facebook_competitor_insights
    group by page_id
) as page_constants ON page_id = hashtags_statistics.facebook_id
```

---

#### 4.5.7 `getBiographyQuery($sortOrder)` -- Biography Data

**Tables:** `facebook_competitor_posts`, `facebook_competitor_insights`

```sql
SELECT * FROM (
    select last_value(biography) as biography,
           lengthUTF8(biography) as biography_length,
           facebook_id,
           multiIf(...) as state,
           multiIf(...) as slug
    from facebook_competitor_posts
    WHERE {pageFilters}
    group by facebook_id
    ORDER BY max(inserted_at) DESC
) as biography_statistics
LEFT JOIN (
    select argMax(profile_picture_url, inserted_at) as image,
           argMax(page_name, inserted_at) as name,
           argMax(followers_count, inserted_at) as followersCount,
           page_id
    from facebook_competitor_insights
    group by page_id
) as page_constants ON page_id = biography_statistics.facebook_id
ORDER BY {sortOrder} DESC
```

**Note:** No date filter on biography -- always shows latest value.

---

#### 4.5.8 `getFollowersGrowthComparisonQuery($sortOrder)` -- Followers Growth Over Time

**Table:** `facebook_competitor_insights`

```sql
SELECT facebook_id,
    multiIf(...) as name,
    multiIf(...) as image,
    multiIf(...) as state,
    multiIf(...) as slug,
    dates, followers_count, dates_with_followers_count
FROM (SELECT c1 as facebook_id FROM VALUES {accountIds}) as page_ids
LEFT JOIN (
    SELECT facebook_id,
        groupArray(date) as dates,
        groupArray(followers_count) as followers_count,
        arrayZip(dates, followers_count) as dates_with_followers_count
    FROM (
        select page_name, profile_picture_url,
               page_id as facebook_id,
               toDate(inserted_at) as date,
               toInt32(followers_count) as followers_count
        from facebook_competitor_insights
        WHERE {pageFilters_insights} AND {dateFilters_insights}
        order by date ASC
    )
    group by facebook_id
) as page_insights on page_ids.facebook_id = page_insights.facebook_id
ORDER BY {sortOrder} ASC
```

Returns `dates_with_followers_count` as a zipped array of `[date, followers_count]` tuples.

---

#### 4.5.9 `getPostReactDistributionByCompany($sortOrder, $facebook_id)` -- Per-Reaction Breakdown

**Table:** `facebook_competitor_posts`

```sql
SELECT facebook_id, page_name,
    'https://graph.facebook.com/' || facebook_id || '/picture' as image,
    toInt32(sum(like)) as total_likes,
    toInt32(sum(haha)) as total_hahas,
    toInt32(sum(angry)) as total_angry,
    toInt32(sum(sad)) as total_sad,
    toInt32(sum(thankful)) as total_thankful,
    toInt32(sum(love)) as total_love,
    toInt32(sum(wow)) as total_wow,
    toInt32(sum(total_post_reactions)) as total_post_reactions,
    toInt32(sum(comments)) as comments,
    toInt32(sum(shares)) as shares,
    toInt32(count()) as total_posts
FROM facebook_competitor_posts
WHERE facebook_id = '{facebook_id}' AND {dateFilters}
GROUP BY facebook_id, page_name
```

**Note:** No deduplication CTE -- reads all rows directly.

---

#### 4.5.10 `getPostEngagementOverTime($sortOrder, $facebook_id)` -- Daily Engagement Time Series

**Table:** `facebook_competitor_posts`

```sql
SELECT
    toInt32(sum(post_engagement)) as total_engagements,
    toInt32(count()) as total_posts,
    toDate(created_at) as date
FROM facebook_competitor_posts
WHERE facebook_id = '{facebook_id}' AND {dateFilters}
GROUP BY toDate(created_at)
ORDER BY toDate(created_at) ASC
WITH FILL FROM toDate('{startDate}') TO toDate('{endDate}') STEP 1
```

---

#### 4.5.11 `getPostEngagementByCompetitor($sortOrder)` -- Total Engagement Per Competitor

**Table:** `facebook_competitor_posts`

```sql
select facebook_id, page_name,
       'https://graph.facebook.com/' || facebook_id || '/picture' as image,
       toInt32(count()) as total_posts,
       toInt32(sum(post_engagement)) as total_engagements
from (
    select facebook_id, page_name, post_id,
           max(post_engagement) as post_engagement
    FROM facebook_competitor_posts
    WHERE {pageFilters} AND {dateFilters}
    group by facebook_id, page_name, post_id
)
group by facebook_id, page_name
ORDER BY total_engagements DESC
```

Deduplication is done inline: `max(post_engagement)` per `post_id` before aggregating.

---

## 5. Instagram Analytics

### 5.1 Request Parameters

Same as Facebook but uses `instagram_id` instead of `facebook_id`. Media types default to `['REELS','IMAGE','VIDEO','CAROUSEL_ALBUM']`.

### 5.2 Endpoints and SQL Queries

#### 5.2.1 Summary (`summaryQuery`)

**Route:** `POST overview/instagram/summary`

**Tables:** `instagram_posts`, `instagram_insights`

```sql
WITH posts AS (
    SELECT media_id, max(stored_event_at)
    FROM instagram_posts
    WHERE instagram_id in {instagram_id}
      AND toDateTime(post_created_at,0,'{timezone}') BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
    GROUP BY media_id
)
SELECT
    toInt32(sum(total_engagement))   as post_engagement,
    toInt32(sum(post_views))         as post_views,
    toInt32(sum(total_reach))        as post_reach,
    toInt32(sum(total_saves))        as post_saves,
    toInt32(sum(reactions))          as post_reactions,
    toInt32(sum(comments))           as post_comments,
    toInt32(sum(profile_views))      as profile_views,
    toInt32(first_value(followers_count))  as followers_count,
    toInt32(first_value(follows_count))    as follows_count,
    toInt32(sum(accounts_engaged))   as accounts_engaged,
    toInt32(sum(engagement))         as profile_engagement,
    toInt32(sum(impressions))        as profile_impressions,
    toInt32(sum(reach))              as profile_reach,
    toInt32(sum(doc_count))          as doc_count,
    toInt32(sum(total_stories))      as total_stories,
    toInt32(sum(total_posts))        as total_posts,
    round(if(doc_count > 0, toFloat32(post_engagement/doc_count), 0), 2) as eng_rate
FROM (
    -- posts subquery with LEFT JOIN insights subquery
    -- Posts: count, SUM(comments_count), SUM(engagement), SUM(impressions), SUM(reach), SUM(saved), SUM(like_count), SUM(views)
    -- Stories counted via: countIf(entity_type = 'STORY' OR media_type = 'STORY')
    -- Insights: SUM(daily_profile_views), argMaxIf(daily_followers_count, date_bucket, daily_followers_count > 0), etc.
)
```

---
**Table:** `instagram_insights`

```sql
SELECT
    notEmpty(follower_count_temp) AS show_data,
    arrayFill(x -> not x == 0, follower_count_temp) AS followers,
    followers_daily,
    dates AS buckets
FROM (
    SELECT
        groupArray(follower_count) AS follower_count_temp,
        arrayConcat(
            [toInt32(0)],
            arrayMap(x -> toInt32(x),
                arraySlice(
                    arrayDifference(arrayFill(x -> not (x == 0), follower_count_temp)),
                    2
                )
            )
        ) AS followers_daily,
        groupArray(dates) AS dates
    FROM (
        SELECT
            toDate(created_time) AS dates,
            toInt32(argMin(followers_count, stored_event_at)) AS follower_count
        FROM instagram_insights
        WHERE instagram_id IN {instagram_id}
          AND toDateTime(created_time,0,'{timezone}') BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
        GROUP BY dates
        ORDER BY dates ASC
        WITH FILL FROM toDate('{currentStartDate}') TO toDate('{currentEndDate}') + 1 STEP 1
    )
)
```

**Zero-fill fallback:** If `followers[0] == 0`, looks back 2 years via `getLastFollowersCount()`.

---

#### 5.2.3 Last Followers Count (Fallback) (`getLastFollowersCount`)

**Table:** `instagram_insights`

Used when `followers[0] == 0`. Looks back 2 years.

```sql
SELECT arrayFirst(x -> x != 0, groupArray(followers)) AS followers
FROM (
    SELECT
        last_value(created_time) as inserted_time,
        toInt32(last_value(followers_count)) as followers
    FROM instagram_insights
    WHERE instagram_id in {instagram_id}
      AND toDateTime(created_time,0,'{timezone}') BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
      AND followers_count != 0
    GROUP BY record_id
    ORDER BY inserted_time DESC
)
```

---

#### 5.2.4 Audience Rollup (`audienceRollupQuery`)

**Table:** `instagram_insights`

```sql
SELECT
    toInt32(argMaxIf(followers, date_bucket, followers > 0)) as follower_count,
    toInt32(last_value(followers) - first_value(followers))  as follower_gained,
    max(date_bucket) as dates
FROM (
    SELECT
        toInt32(argMin(followers_count, stored_event_at)) AS followers,
        toDate(created_time) AS date_bucket
    FROM instagram_insights
    WHERE instagram_id in {instagram_id}
      AND toDateTime(created_time,0,'{timezone}') BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
      AND followers_count > 0
    GROUP BY date_bucket
    ORDER BY date_bucket ASC
)
```

---

#### 5.2.5 Publishing Behaviour (`getOverviewPublishingBehaviourByMediaTypeQuery`)

**Route:** `POST overview/instagram/publish`

**Table:** `instagram_posts`

```sql
WITH posts AS (
    SELECT media_id, last_value(stored_event_at)
    FROM instagram_posts
    WHERE instagram_id in {instagram_id}
      AND toDateTime(post_created_at,0,'{timezone}') BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
      AND media_type in ('{media_type}')
    GROUP BY media_id
)
SELECT
    groupArray(likes)       as likes,
    groupArray(comments)    as comments,
    groupArray(saved)       as saved,
    groupArray(engagement)  as engagement,
    groupArray(reach)       as reach,
    groupArray(impressions) as impressions,
    groupArray(views)       as views,
    groupArray(total_posts) as total_posts,
    groupArray(created_at)  as buckets
FROM (
    SELECT
        toInt32(sum(like_count))      as likes,
        toInt32(sum(comments_count))  as comments,
        toInt32(sum(saved))           as saved,
        toInt32(sum(engagement))      as engagement,
        toInt32(sum(reach))           as reach,
        toInt32(sum(impressions))     as impressions,
        toInt32(count(*))             as total_posts,
        toInt32(sum(views))           as views,
        toDate(post_created_at)       as created_at
    FROM (
        SELECT
            media_id,
            last_value(like_count)      as like_count,
            last_value(comments_count)  as comments_count,
            last_value(saved)           as saved,
            last_value(engagement)      as engagement,
            last_value(reach)           as reach,
            last_value(impressions)     as impressions,
            last_value(views)           as views,
            post_created_at
        FROM instagram_posts
        WHERE (media_id, stored_event_at) in posts
        GROUP BY media_id, post_created_at
    )
    GROUP BY created_at
    ORDER BY created_at ASC
    WITH FILL FROM toDate('{currentStartDate}') TO toDate('{currentEndDate}') + 1
)
```

---

#### 5.2.6 Publish Rollup (`publishRollupQuery`)

**Table:** `instagram_posts`

Returns one row per media type plus a TOTAL row.

```sql
WITH
    posts AS (
        SELECT media_id, last_value(stored_event_at)
        FROM instagram_posts
        WHERE instagram_id IN {instagram_id}
          AND {DateFilter('post_created_at')}
          AND media_type != ''
        GROUP BY media_id
    ),
    media_types AS (
        SELECT arrayJoin(['REELS', 'CAROUSEL_ALBUM', 'IMAGE', 'VIDEO']) AS media_type
    ),
    metrics AS (
        SELECT
            mt.media_type,
            toInt32(COALESCE(total_posts, 0)) AS total_posts,
            toInt32(COALESCE(likes, 0))       AS likes,
            toInt32(COALESCE(comments, 0))    AS comments,
            toInt32(COALESCE(saved, 0))       AS saved,
            toInt32(COALESCE(engagement, 0))  AS engagement,
            toInt32(COALESCE(reach, 0))       AS reach,
            toInt32(COALESCE(views, 0))       AS views
        FROM media_types mt
        LEFT JOIN (
            SELECT
                media_type,
                COUNT(*) AS total_posts, SUM(like_count) AS likes,
                SUM(comments_count) AS comments, SUM(saved) AS saved,
                SUM(engagement) AS engagement, SUM(reach) AS reach,
                SUM(views) AS views
            FROM (
                SELECT
                    media_id,
                    last_value(media_type)        AS media_type,
                    last_value(like_count)        AS like_count,
                    last_value(comments_count)    AS comments_count,
                    last_value(saved)             AS saved,
                    last_value(engagement)        AS engagement,
                    last_value(reach)             AS reach,
                    last_value(views)             AS views,
                    post_created_at
                FROM instagram_posts
                WHERE (media_id, stored_event_at) IN posts
                GROUP BY media_id, post_created_at
            )
            GROUP BY media_type
        ) t ON mt.media_type = t.media_type
    )
SELECT * FROM (
    SELECT * FROM metrics
    UNION ALL
    SELECT 'TOTAL' AS media_type,
        SUM(total_posts) AS total_posts, SUM(likes) AS likes,
        SUM(comments) AS comments,       SUM(saved) AS saved,
        SUM(engagement) AS engagement,   SUM(reach) AS reach,
        sum(views) AS views
    FROM metrics
)
ORDER BY CASE WHEN media_type = 'TOTAL' THEN 1 ELSE 0 END, media_type
```

---

#### 5.2.7 Top Posts (`topPostQuery`)

**Table:** `instagram_posts`

```sql
WITH posts AS (
    SELECT media_id, max(stored_event_at)
    FROM instagram_posts
    WHERE instagram_id in {instagram_id}
      AND {DateFilter('post_created_at')}
      AND media_type in ('{media_type}')
      [AND has(hashtags, '{hashtag}')]    -- only if hashtags filter present
    GROUP BY media_id
)
SELECT * EXCEPT (taps_back, taps_forward), engagement as total_engagement
FROM instagram_posts
WHERE (media_id, stored_event_at) in posts
  AND entity_type != 'STORY'
ORDER BY {order_by} DESC
LIMIT {limit}
```

**Parameters:** `limit` default 15 (top posts page) or 5 (overview), `order_by` default `'engagement'`, `media_type` from request, `hashtags` optional.

---

#### 5.2.8 Active Users by Hour (`activeUsersHourQuery`)

**Route:** `POST overview/instagram/activeUsers`

**Table:** `instagram_insights`

```sql
SELECT
    max(buckets)     AS buckets,
    max(values)      as values,
    arrayMax(values) as highest_value,
    buckets[indexOf(values, highest_value)] as highest_hour
FROM (
    SELECT
        first_value(buckets) as buckets,
        arrayMap((x)->toInt32(x)/count(), sumForEach(values)) AS values
    FROM (
        SELECT
            arrayMap((x)->toInt32(x[1]), active_users_per_hour) as buckets,
            arrayMap((x)->toInt32(x[2]), active_users_per_hour) as values,
            arrayMax(values)                AS highest_value,
            toInt32(indexOf(values, highest_value)) AS highest_hour
        FROM (
            SELECT arrayMap((x)->(splitByChar('$', x)), online_followers) as active_users_per_hour
            FROM (
                SELECT
                    max(instagram_id)            AS insta_id,
                    max(online_followers)        AS online_followers,
                    max(online_users_datetime)   AS latest_record
                FROM instagram_insights
                WHERE instagram_id in {instagram_id}
                  AND toDateTime(online_users_datetime,0,'{timezone}') BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
                GROUP BY record_id
                ORDER BY latest_record DESC
            )
        )
    )
)
ORDER BY arrayMax(buckets) DESC
```

**Data format:** `online_followers` stores `"0$42$1$18$2$5..."` where odd positions are hour numbers and even positions are user counts.

**Controller timezone adjustment:** `round(Carbon::now($timezone)->offsetHours) + 8` (+8 accounts for UTC+8 storage), wraps modulo 24.

---

#### 5.2.9 Active Users by Day (`activeUsersDayQuery`)

**Table:** `instagram_insights`

```sql
SELECT
    groupArray(value)    AS values,
    groupArray(day)      AS buckets,
    arrayMax(values)     AS highest_value,
    buckets[indexOf(values, highest_value)] AS highest_day
FROM (
    SELECT
        max(toInt32(arraySum(arrayMap((x)->toInt32(x[2]),active_users_per_hour)))) AS value,
        day
    FROM (
        SELECT
            arrayMap((x)->(splitByChar('$', x)), online_followers) as active_users_per_hour,
            day
        FROM (
            SELECT
                max(day_of_week)         AS day,
                max(online_followers)    AS online_followers,
                record_id
            FROM instagram_insights
            WHERE instagram_id in {instagram_id}
              AND toDateTime(created_time,0,'{timezone}') BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
            GROUP BY record_id
        )
    )
    WHERE day != ''
    GROUP BY day
    ORDER BY
        CASE
            WHEN day='Monday'    THEN 1
            WHEN day='Tuesday'   THEN 2
            WHEN day='Wednesday' THEN 3
            WHEN day='Thursday'  THEN 4
            WHEN day='Friday'    THEN 5
            WHEN day='Saturday'  THEN 6
            WHEN day='Sunday'    THEN 7
        END
)
```

---

#### 5.2.10 Impressions (`impressionsQuery`)

**Route:** `POST overview/instagram/impressions`

**Table:** `instagram_posts`

```sql
WITH posts AS (
    SELECT media_id, max(stored_event_at)
    FROM instagram_posts
    WHERE instagram_id in {instagram_id}
      AND {DateFilter('post_created_at')}
    GROUP BY media_id
)
SELECT
    groupArray(dates)   as buckets,
    groupArray(impress) as impressions,
    toInt32(sum(impress)) as show_data
FROM (
    SELECT
        toInt32(SUM(impressions)) as impress,
        toDate(post_created_at) as dates
    FROM instagram_posts
    WHERE (media_id, stored_event_at) in posts
    GROUP BY media_id, post_created_at
    ORDER BY toDate(post_created_at) ASC
    WITH FILL FROM toDate('{currentStartDate}') TO toDate('{currentEndDate}') + 1
)
```

**Rollup (`impressionsRollupQuery`):**

```sql
WITH posts AS (
    SELECT media_id, max(stored_event_at)
    FROM instagram_posts
    WHERE instagram_id in {instagram_id}
      AND {DateFilter('post_created_at')}
    GROUP BY media_id
)
SELECT
    toInt32(sum(impressions))   AS total_impressions,
    ROUND(avg(impressions), 3)  AS avg_impressions
FROM (
    SELECT
        toInt32(SUM(impressions)) as impressions,
        toDate(post_created_at) as dates
    FROM instagram_posts
    WHERE (media_id, stored_event_at) in posts
    GROUP BY media_id, post_created_at
    ORDER BY toDate(post_created_at) ASC
)
```

---

#### 5.2.11 Reels (`reelsQuery`)

**Route:** `POST overview/instagram/reels`

**Table:** `instagram_posts` (filtered `media_type='REELS'`)

```sql
SELECT
    groupArray(created_at)       as buckets,
    groupArray(total_post)       as total_posts,
    groupArray(engagement)       as engagement,
    groupArray(likes)            as likes,
    groupArray(comments)         as comments,
    groupArray(saves)            as saves,
    groupArray(shares)           as shares,
    groupArray(avg_watch_time)   as avg_watch_time,
    groupArray(total_watch_time) as total_watch_time,
    toInt32(sum(total_post))     as show_data
FROM (
    SELECT
        created_at,
        toInt32(count())              as total_post,
        toInt32(sum(engagement))      as engagement,
        toInt32(sum(likes))           as likes,
        toInt32(sum(comments))        as comments,
        toInt32(sum(saves))           as saves,
        toInt32(sum(shares))          as shares,
        if(count() != 0, round(avg(avg_watch_time)/1000, 2), 0) as avg_watch_time,
        toInt32(sum(total_watch_time)/1000) as total_watch_time
    FROM (
        SELECT
            toDate(last_value(post_created_at))         as created_at,
            last_value(engagement)                      as engagement,
            last_value(like_count)                      as likes,
            last_value(comments_count)                  as comments,
            last_value(saved)                           as saves,
            last_value(shares)                          as shares,
            last_value(reels_avg_watch_time)            as avg_watch_time,
            last_value(reels_total_watch_time)          as total_watch_time
        FROM instagram_posts
        WHERE instagram_id in {instagram_id}
          AND {DateFilter('post_created_at')}
          AND media_type='REELS'
        GROUP BY media_id
    )
    GROUP BY created_at
    ORDER BY created_at ASC
    WITH FILL FROM toDate('{currentStartDate}') TO toDate('{currentEndDate}') + 1
)
```

Watch times stored in ms, divided by 1000 for seconds output.

**Reels Rollup (`reelsRollupQuery`):**

```sql
SELECT
    toInt32(sum(engagement))    as engagement,
    toInt32(sum(likes))         as likes,
    toInt32(sum(comments))      as comments,
    toInt32(sum(saves))         as saves,
    toInt32(count())            as total_posts,
    toInt32(sum(shares))        as shares,
    if(count() != 0, round(avg(avg_watch_time)/1000, 2), 0) as avg_watch_time,
    round(sum(total_watch_time)/1000, 2)                    as total_watch_time
FROM (
    SELECT
        last_value(engagement)             as engagement,
        last_value(like_count)             as likes,
        last_value(comments_count)         as comments,
        last_value(saved)                  as saves,
        last_value(shares)                 as shares,
        last_value(reels_avg_watch_time)   as avg_watch_time,
        last_value(reels_total_watch_time) as total_watch_time
    FROM instagram_posts
    WHERE instagram_id in {instagram_id}
      AND {DateFilter('post_created_at')}
      AND media_type='REELS'
    GROUP BY media_id
)
```

---

#### 5.2.12 Engagements (`engagementsQuery`)

**Route:** `POST overview/instagram/engagement`

**Table:** `instagram_posts`

```sql
WITH posts AS (
    SELECT media_id, max(stored_event_at)
    FROM instagram_posts
    WHERE instagram_id in {instagram_id}
      AND {DateFilter('post_created_at')}
    GROUP BY media_id
)
SELECT
    groupArray(dates)     AS buckets,
    groupArray(engage)    AS engagement,
    groupArray(comments)  AS comments,
    groupArray(reactions) AS reactions,
    groupArray(doc_count) as doc_count,
    toInt32(sum(engage))  as show_data
FROM (
    SELECT
        toInt32(count())              as doc_count,
        toInt32(SUM(engagement))      as engage,
        toInt32(SUM(comments_count))  as comments,
        toInt32(SUM(like_count))      as reactions,
        toDate(post_created_at)       as dates
    FROM instagram_posts
    WHERE (media_id, stored_event_at) in posts
    GROUP BY toDate(post_created_at)
    ORDER BY toDate(post_created_at) ASC
    WITH FILL FROM toDate('{currentStartDate}') TO toDate('{currentEndDate}') + 1
)
```

**Engagements Rollup (`engagementsRollupQuery`):**

```sql
WITH posts AS (
    SELECT media_id, max(stored_event_at)
    FROM instagram_posts
    WHERE instagram_id in {instagram_id}
      AND {DateFilter('post_created_at')}
    GROUP BY media_id
)
SELECT
    toInt32(sum(engage))    AS engagement,
    ROUND(avg(engage), 3)   AS avg_engagement,
    toInt32(sum(comments))  as comments,
    toInt32(sum(reactions)) as reactions,
    toInt32(sum(saved))     as saved,
    toInt32(sum(doc_count)) as count
FROM (
    SELECT
        toInt32(count())              as doc_count,
        toInt32(sum(saved))           as saved,
        toInt32(SUM(engagement))      as engage,
        toInt32(SUM(comments_count))  as comments,
        toInt32(SUM(like_count))      as reactions,
        toDate(MAX(post_created_at))  as dates
    FROM instagram_posts
    WHERE (media_id, stored_event_at) in posts
    GROUP BY media_id
    ORDER BY toDate(MAX(post_created_at)) ASC
    WITH FILL FROM toDate('{currentStartDate}') TO toDate('{currentEndDate}') + 1
)
```

---

#### 5.2.13 Hashtags (`hashtagsEngagedQuery`)

**Route:** `POST overview/instagram/hashtags`

**Table:** `instagram_posts`

Returns top 30 hashtags by engagement.

```sql
WITH posts_data AS (
    SELECT media_id, max(stored_event_at)
    FROM instagram_posts
    WHERE instagram_id in {instagram_id}
      AND {DateFilter('post_created_at')}
    GROUP BY media_id
)
SELECT
    groupArray(name)        as name,
    groupArray(engagement)  as engagement,
    groupArray(likes)       as likes,
    groupArray(comments)    as comments,
    groupArray(saved)       as saved,
    groupArray(posts)       as posts
FROM (
    SELECT
        hashtag as name,
        toInt32(sum(engagement)) AS engagement,
        toInt32(sum(likes))      as likes,
        toInt32(sum(comments))   as comments,
        toInt32(sum(saved))      as saved,
        toInt32(sum(counts))     as posts
    FROM (
        SELECT
            arrayJoin(hashtags)   AS hashtag,
            sum(engagement)       as engagement,
            sum(like_count)       as likes,
            sum(comments_count)   as comments,
            sum(saved)            as saved,
            uniq(media_id)        AS counts
        FROM instagram_posts
        WHERE (media_id, stored_event_at) in posts_data
        GROUP BY hashtags
        ORDER BY engagement DESC
    )
    GROUP BY name
    ORDER BY engagement DESC
    LIMIT 30
)
```

**Hashtags Rollup (`hashtagsRollupQuery`):**

```sql
WITH posts AS (
    SELECT media_id, max(stored_event_at)
    FROM instagram_posts
    WHERE instagram_id in {instagram_id}
      AND {DateFilter('post_created_at')}
    GROUP BY media_id
)
SELECT
    toInt32(sum(engagement))                            AS total_engagement,
    toInt32(sum(like_count))                            AS total_likes,
    toInt32(sum(comments_count))                        AS total_comments,
    toInt32(sum(saved))                                 AS total_saves,
    toInt32(count(DISTINCT arrayJoin(hashtags)))        AS total_unique_hashtags,
    toInt32(sum(length(hashtags)))                      AS total_hashtag_uses
FROM (
    SELECT
        last_value(engagement)      as engagement,
        last_value(like_count)      as like_count,
        last_value(comments_count)  as comments_count,
        last_value(saved)           as saved,
        last_value(hashtags)        as hashtags,
        stored_event_at,
        media_id
    FROM instagram_posts
    WHERE (media_id, stored_event_at) in posts
    GROUP BY media_id, stored_event_at
)
```

---

#### 5.2.14 Stories (`storiesQuery`)

**Route:** `POST overview/instagram/stories`

**Table:** `instagram_posts` (filtered `entity_type = 'STORY' OR media_type = 'STORY'`)

```sql
WITH posts AS (
    SELECT media_id, max(stored_event_at)
    FROM instagram_posts
    WHERE instagram_id in {instagram_id}
      AND {DateFilter('post_created_at')}
    GROUP BY media_id
)
SELECT
    groupArray(dates)              AS buckets,
    groupArray(avg_impress)        AS avg_story_impressions,
    groupArray(impress)            AS story_impressions,
    groupArray(reach)              AS story_reach,
    groupArray(reply)              AS story_reply,
    groupArray(exit)               AS story_exits,
    groupArray(tap_forward)        AS story_taps_forward,
    groupArray(tap_back)           AS story_taps_back,
    groupArray(published_stories)  AS published_stories,
    toInt32(sum(reach) + sum(reply) + sum(exit) + sum(tap_forward) + sum(tap_back)) as show_data
FROM (
    SELECT
        toDate(post_created_at) AS dates,
        toInt32(SUM(replies))   AS reply,
        toInt32(SUM(exits))     AS exit,
        toInt32(SUM(taps_forward)) AS tap_forward,
        toInt32(SUM(taps_back))    AS tap_back,
        toInt32(SUM(CASE WHEN entity_type = 'STORY' OR media_type = 'STORY' THEN 1 ELSE 0 END)) AS published_stories,
        CASE WHEN published_stories = 0 THEN 0 ELSE toInt32(SUM(reach)) END       AS reach,
        CASE WHEN published_stories = 0 THEN 0 ELSE toInt32(SUM(impressions)) END AS impress,
        CASE WHEN published_stories = 0 THEN 0 ELSE toInt32(AVG(impressions)) END AS avg_impress
    FROM instagram_posts
    WHERE (media_id, stored_event_at) in posts
      AND (entity_type = 'STORY' OR media_type = 'STORY')
    GROUP BY media_id, post_created_at
    ORDER BY toDate(post_created_at) ASC
    WITH FILL FROM toDate('{currentStartDate}') TO toDate('{currentEndDate}') + 1
)
```

**Stories Rollup (`storiesRollupQuery`):**

```sql
WITH posts AS (
    SELECT media_id, max(stored_event_at)
    FROM instagram_posts
    WHERE instagram_id in {instagram_id}
      AND {DateFilter('post_created_at')}
    GROUP BY media_id
)
SELECT
    toInt32(sum(impress))       AS story_impressions,
    toInt32(avg(impress))       AS avg_story_impressions,
    toInt32(sum(reach))         AS story_reach,
    toInt32(sum(reply))         AS story_reply,
    toInt32(sum(exit))          AS story_exits,
    toInt32(sum(tap_forward))   AS story_taps_forward,
    toInt32(sum(tap_back))      AS story_taps_back,
    toInt32(sum(published_stories)) AS published_stories
FROM (
    SELECT
        toInt32(SUM(replies))   AS reply,
        toInt32(SUM(exits))     AS exit,
        toInt32(SUM(taps_forward)) AS tap_forward,
        toInt32(SUM(taps_back))    AS tap_back,
        toInt32(SUM(CASE WHEN entity_type = 'STORY' OR media_type = 'STORY' THEN 1 ELSE 0 END)) AS published_stories,
        CASE WHEN published_stories = 0 THEN 0 ELSE toInt32(SUM(reach)) END       AS reach,
        CASE WHEN published_stories = 0 THEN 0 ELSE toInt32(SUM(impressions)) END AS impress
    FROM instagram_posts
    WHERE (media_id, stored_event_at) in posts
      AND (entity_type = 'STORY' OR media_type = 'STORY')
    GROUP BY media_id, post_created_at
    ORDER BY toDate(post_created_at) ASC
    WITH FILL FROM toDate('{currentStartDate}') TO toDate('{currentEndDate}') + 1
)
```

---

#### 5.2.15 Demographics

**Audience Age (`audienceAgeQuery`):**

**Table:** `instagram_insights` -- demographics stored as `$`-delimited pairs in `audience_age`.

```sql
SELECT
    max(CASE WHEN ages = '65+'   THEN ages_count[7] ELSE 0 END) AS `65+`,
    max(CASE WHEN ages = '55-64' THEN ages_count[6] ELSE 0 END) AS `55-64`,
    max(CASE WHEN ages = '45-54' THEN ages_count[5] ELSE 0 END) AS `45-54`,
    max(CASE WHEN ages = '35-44' THEN ages_count[4] ELSE 0 END) AS `35-44`,
    max(CASE WHEN ages = '25-34' THEN ages_count[3] ELSE 0 END) AS `25-34`,
    max(CASE WHEN ages = '18-24' THEN ages_count[2] ELSE 0 END) AS `18-34`,
    max(CASE WHEN ages = '13-17' THEN ages_count[1] ELSE 0 END) AS `13-17`
FROM (
    SELECT
        arrayMap((x)->(x[1]), age_pair)          AS ages,
        arrayMap((x)->toInt32(x[2]), age_pair)   AS ages_count
    FROM (
        SELECT arrayMap((x)->(splitByChar('$',x)), audience_age) as age_pair
        FROM (
            SELECT instagram_id,
                   max(audience_age)    AS audience_age,
                   max(created_time)    AS latest_record
            FROM instagram_insights
            WHERE instagram_id in {instagram_id}
              AND {DateFilter('created_time')}
            GROUP BY instagram_id
        )
    )
) ARRAY JOIN ages
```

Known bug: `'18-24'` bucket is aliased as `'18-34'` in output.

**Audience Gender (`audienceGenderQuery`):**

```sql
SELECT
    MAX(CASE WHEN gender = 'U' THEN gender_count[1] ELSE 0 END) AS U,
    MAX(CASE WHEN gender = 'M' THEN gender_count[2] ELSE 0 END) AS M,
    MAX(CASE WHEN gender = 'F' THEN gender_count[3] ELSE 0 END) AS F
FROM (
    SELECT
        arrayMap((x)-> (x[1]), gender_pair)          As gender,
        arrayMap((x)-> toInt32(x[2]), gender_pair)   As gender_count,
        followers_count
    FROM (
        SELECT arrayMap((x)->(splitByChar('$',x)), audience_gender) as gender_pair, followers_count
        FROM (
            SELECT instagram_id,
                   max(audience_gender)            AS audience_gender,
                   first_value(followers_count)    AS followers_count,
                   max(created_time)               as latest_record
            FROM instagram_insights
            WHERE instagram_id in {instagram_id}
              AND {DateFilter('created_time')}
            GROUP BY instagram_id
        )
    )
) ARRAY JOIN gender
```

**Audience Max Gender+Age (`audienceMaxQuery`):**

```sql
SELECT
    genders_age[max_index][1] AS gender,
    genders_age[max_index][2] AS age,
    max_value as value
FROM (
    SELECT
        arrayMap((x)->(splitByChar('.',x)), arrayMap((x)->(x[1]), gender_age_pair)) AS genders_age,
        arrayMax(arrayMap((x)->toInt32(x[2]), gender_age_pair))    AS max_value,
        indexOf(arrayMap((x)->toInt32(x[2]), gender_age_pair), max_value) as max_index
    FROM (
        SELECT arrayMap((x)->(splitByChar('$',x)), audience_gender_age) as gender_age_pair
        FROM (
            SELECT instagram_id,
                   max(audience_gender_age) AS audience_gender_age,
                   max(created_time)        AS latest_record
            FROM instagram_insights
            WHERE instagram_id in {instagram_id}
              AND {DateFilter('created_time')}
            GROUP BY instagram_id
        )
    )
)
```

**Cities (`citiesQuery`):**

```sql
SELECT *
FROM (
    SELECT
        arrayMap((x)->x[1], audience_city_pair)          AS `city.cities`,
        arrayMap((x)->toInt32(x[2]), audience_city_pair) AS `city.city_values`
    FROM (
        SELECT arrayMap((x)->(splitByChar('$',x)), audience_city) as audience_city_pair
        FROM (
            SELECT instagram_id,
                   max(audience_city)   AS audience_city,
                   max(created_time)    AS latest_record
            FROM instagram_insights
            WHERE instagram_id in {instagram_id}
              AND {DateFilter('created_time')}
            GROUP BY instagram_id
        )
    )
) ARRAY JOIN city
ORDER BY city.city_values DESC
```

**Countries (`countriesQuery`):**

```sql
SELECT *
FROM (
    SELECT
        arrayMap((x)->x[1], audience_country_pair)          AS `country.countries`,
        arrayMap((x)->toInt32(x[2]), audience_country_pair) AS `country.country_values`
    FROM (
        SELECT arrayMap((x)->(splitByChar('$',x)), audience_country) as audience_country_pair
        FROM (
            SELECT instagram_id,
                   max(audience_country) AS audience_country,
                   max(created_time)     AS latest_record
            FROM instagram_insights
            WHERE instagram_id in {instagram_id}
              AND {DateFilter('created_time')}
            GROUP BY instagram_id
        )
    )
) ARRAY JOIN country
ORDER BY country.country_values DESC
```

---

#### 5.2.16 Time Recommendation (`timeRecommendationQuery`)

**Table:** `instagram_posts`

```sql
SELECT
    instagram_id,
    day_of_week,
    hour_of_day,
    sum(impressions)  as post_impressions,
    sum(engagement)   as total_engagement
FROM instagram_posts
WHERE instagram_id in {instagram_id}
  AND {DateFilter('post_created_at')}
GROUP BY instagram_id, day_of_week, hour_of_day
ORDER BY day_of_week, hour_of_day
```

---

## 6. Instagram Competitor Analytics

### 6.1 Constructor and Timezone Handling

Same pattern as Facebook Competitor: dates parsed in user timezone, converted to UTC before embedding in SQL.

```
startDate = Carbon::parse(date[0], timezone)->setTime(0,0,1)->setTimezone('UTC')
endDate = Carbon::parse(date[1], timezone)->setTime(23,59,59)->setTimezone('UTC')
```

**Key difference from own Instagram analytics:** The request key is `date_filter` (not `date`).

### 6.2 Tables

| Table | Used for |
|-------|---------|
| `instagram_competitor_posts` | All post-level competitor metrics |
| `instagram_competitor_insights` | Competitor account-level metrics (followers, following, profile picture) |

### 6.3 Helper Methods

#### `getPageFilters($insightsFlag, $filterAdded)`
```sql
-- posts: business_account_id in('id1','id2',...)
-- insights: instagram_account_id in('id1','id2',...)
-- filterAdded=true: excludes accounts with state='Added'
```

#### `getDateFilters($insightsFlag, $dateColumn)`
```sql
-- posts: created_at BETWEEN toDateTime('{startDate}',0) AND toDateTime('{endDate}',0)
-- insights: inserted_at BETWEEN toDateTime('{startDate}',0) AND toDateTime('{endDate}',0)
```

#### `getConstantConditions($field, $idField)`
Builds `multiIf(id='x','val',...,'Random')` for name/image/state/slug.

### 6.4 Controller Methods

| Method | Description | Sort Default |
|--------|-------------|--------------|
| `postingActivityGraphByTypes` | Aggregate by media_type across all competitors | `avgTotalEngagements` |
| `postingActivityBySpecificType` | Per-competitor for one media_type | `followersCount` |
| `postingActivityTableByType` | Table view per-competitor for one media_type | `followersCount` |
| `followersGrowthComparison` | Per-competitor followers time-series | `total_followed_by_count` |
| `dataTableMetrics` | Current/previous period KPIs with growth % | N/A |
| `topAndLeastPerformingPosts` | Top-5/bottom-5 per competitor | engagement DESC/ASC |
| `topHashtags` | Top hashtags (limit default: 7) | count DESC |
| `individualHashtagData` | Per-competitor stats for one hashtag | engagement DESC |
| `biographyData` | Biography + length per competitor | `biography_length` |

### 6.5 SQL Queries

#### 6.5.1 `getFollowersGrowthComparisonQuery($sortOrder)`

**Table:** `instagram_competitor_insights`

```sql
SELECT
    multiIf(page_ids.business_account_id='{id1}','{name1}',...,'Random') as name,
    multiIf(...) as image,
    multiIf(...) as state,
    multiIf(...) as slug,
    *
FROM (
    SELECT c1 as business_account_id
    FROM VALUES (('id1'),('id2'),...)
) as page_ids
LEFT JOIN (
    SELECT
        instagram_account_id,
        groupArray(date)                     as dates,
        groupArray(total_following_count)    as total_following_count,
        groupArray(total_followed_by_count)  as total_followed_by_count,
        arrayZip(dates, total_following_count)  as dates_with_following_count,
        arrayZip(dates, total_followed_by_count) as dates_with_followed_by_count
    FROM (
        SELECT
            page_name, profile_picture_url, instagram_account_id,
            toDate(inserted_at) as date,
            total_following_count, total_followed_by_count
        FROM instagram_competitor_insights
        WHERE instagram_account_id in('id1','id2',...)
          AND inserted_at BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
        GROUP BY inserted_at, instagram_account_id, profile_picture_url, page_name,
                 total_following_count, total_followed_by_count
        ORDER BY inserted_at
    )
    GROUP BY instagram_account_id, page_name
) as page_insights ON page_ids.business_account_id = page_insights.instagram_account_id
ORDER BY {sortOrder} DESC
```

---

#### 6.5.2 `getPostingActivityGraphByTypesQuery($sortOrder)`

Aggregated across ALL competitors by media type.

**Table:** `instagram_competitor_posts`

```sql
WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM instagram_competitor_posts
    WHERE business_account_id in('id1','id2',...)
      AND created_at BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
    GROUP BY post_id
)
SELECT
    post_type as mediaType,
    media_product_type as mediaProductType,
    round(sum(avg_engagement), 2) as avgTotalEngagements,
    sum(page_post_count) as totalPosts,
    round(sum(er), 2) as avgEngagementRate,
    if(dateDiff('week',toDate('{start}'),toDate('{end}'))=0, 0,
       round(totalPosts/dateDiff('week',toDate('{start}'),toDate('{end}')),2)) as postsPerWeek,
    if(dateDiff('day',toDate('{start}'),toDate('{end}'))=0, 0,
       round(totalPosts/dateDiff('day',toDate('{start}'),toDate('{end}')),2)) as postsPerDay,
    if(dateDiff('hour',toDate('{start}'),toDate('{end}'))=0, 0,
       round(totalPosts/dateDiff('hour',toDate('{start}'),toDate('{end}')),2)) as postsPerHour,
    dateDiff('week',toDate('{start}'),toDate('{end}')) as weekCount,
    dateDiff('day',toDate('{start}'),toDate('{end}'))  as dayCount,
    dateDiff('hour',toDate('{start}'),toDate('{end}')) as hourCount
FROM (
    SELECT
        business_account_id, name, post_type, media_product_type,
        sum(count) as page_post_count,
        groupArray(total_engagement),
        if(page_post_count <= 0, 0, round(sum(total_engagement)/page_post_count,2)) as avg_engagement,
        if(page_post_count <= 0, 0,
           round(((sum(er)/page_post_count)/max(followers_count))*100,2)) as er
    FROM (
        SELECT
            business_account_id, name,
            if(media_product_type=='REELS','VIDEO REEL',
               if(media_type=='CAROUSEL_ALBUM','CAROUSEL ALBUM', media_type)) as post_type,
            media_product_type,
            count() as count,
            sum(engagement) as total_engagement,
            total_engagement as er,
            argMax(total_followed_by_count, created_at) as followers_count,
            toStartOfWeek(created_at) as week
        FROM instagram_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        GROUP BY business_account_id, name, post_type, media_product_type, week
        ORDER BY week ASC
        WITH FILL FROM toStartOfWeek(toDate('{start}')) TO toStartOfWeek(toDate('{end}')) STEP INTERVAL 1 WEEK
    )
    WHERE business_account_id != ''
    GROUP BY business_account_id, name, post_type, media_product_type
)
GROUP BY post_type, media_product_type
ORDER BY {sortOrder} DESC
```

**Media type mapping:** `REELS` -> `'VIDEO REEL'`, `CAROUSEL_ALBUM` -> `'CAROUSEL ALBUM'`, others as-is.

---

#### 6.5.3 `getPostingActivityBySpecificTypeQuery($mediaType, $mediaProductType, $sortOrder)`

Per-competitor breakdown for a specific media type. Same structure as above but groups by `business_account_id` and adds metadata from `instagram_competitor_insights`.

```sql
-- Key additions:
LEFT JOIN (
    SELECT
        argMax(profile_picture_url, inserted_at) as image,
        argMax(page_name, inserted_at) as name,
        argMax(total_following_count, inserted_at) as followingCount,
        argMax(total_followed_by_count, inserted_at) as followersCount,
        instagram_account_id
    FROM instagram_competitor_insights
    GROUP BY instagram_account_id
) as page_constants ON page_constants.instagram_account_id = pages_ids.business_account_id
```

---

#### 6.5.4 `getPostingActivityTableByTypeQuery($mediaType, $mediaProductType, $sortOrder)`

```sql
WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM instagram_competitor_posts
    WHERE business_account_id in('id1','id2',...)
      AND created_at BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
    GROUP BY post_id
)
SELECT
    pages_ids.business_account_id as businessAccountId,
    multiIf(...) as name, multiIf(...) as image, multiIf(...) as state, multiIf(...) as slug,
    coalesce(page_insights.count, 0) as count,
    coalesce(page_insights.total_engagement, 0) as totalEngagement,
    round(if(
        coalesce(count,0)<=0 OR coalesce(followersCount,0)<=0, 0,
        ((coalesce(total_engagement,0)/coalesce(count,1))/coalesce(followersCount,1))*100
    ), 2) as engagementRate,
    coalesce(page_metadata.followingCount, 0) as followingCount,
    coalesce(page_metadata.followersCount, 0) as followersCount,
    page_insights.media_type as mediaType,
    page_insights.media_product_type as mediaProductType
FROM (
    SELECT c1 as business_account_id FROM VALUES (('id1'),...)
) as pages_ids
LEFT JOIN (
    SELECT business_account_id,
           if(media_product_type=='REELS','VIDEO REEL',if(media_type=='CAROUSEL_ALBUM','CAROUSEL ALBUM',media_type)) as media_type,
           media_product_type,
           max(count) as count,
           max(total_engagement) as total_engagement
    FROM (
        SELECT business_account_id, media_type, media_product_type,
               count() as count, sum(engagement) as total_engagement
        FROM instagram_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        GROUP BY business_account_id, media_type, media_product_type
    ) as page_metrics
    WHERE media_type = '{mediaType}' AND media_product_type = '{mediaProductType}'
    GROUP BY business_account_id, media_type, media_product_type
) as page_insights USING business_account_id
LEFT JOIN (
    SELECT
        argMax(total_following_count, inserted_at) as followingCount,
        argMax(total_followed_by_count, inserted_at) as followersCount,
        instagram_account_id
    FROM instagram_competitor_insights
    GROUP BY instagram_account_id
) as page_metadata ON page_metadata.instagram_account_id = pages_ids.business_account_id
ORDER BY {sortOrder} DESC
```

---

#### 6.5.5 `getDataTableDataMetricsQuery($sortOrder)`

Same weekly-fill pattern as Facebook competitor. Uses `instagram_competitor_posts` and `instagram_competitor_insights` tables with `business_account_id`/`instagram_account_id` join keys.

**Returns per competitor:** `businessAccountId`, `name`, `image`, `state`, `slug`, `averageEngagement`, `averagePostsPerWeek`, `engagementRate`, `dayOfWeek`, `hourOfDay`, `averagePostsPerDay`, `followersCount`, `followingCount`.

---

#### 6.5.6 `getTopPerformingAndLeastPerformingPostsQuery()`

```sql
WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM instagram_competitor_posts
    WHERE business_account_id in(...) AND created_at BETWEEN ...
    GROUP BY post_id
)
SELECT * FROM (
    SELECT *, 'top_5_posts' as category
    FROM (
        SELECT * FROM instagram_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        ORDER BY engagement desc
        LIMIT 5 BY business_account_id
    )
UNION ALL
    SELECT *, 'least_5_posts' as category
    FROM (
        SELECT * FROM instagram_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        ORDER BY engagement asc
        LIMIT 5 BY business_account_id
    )
)
```

---

#### 6.5.7 `getTopHashtagsQuery($limit)`

Same pattern as Facebook competitor hashtags but uses `instagram_competitor_posts` and `instagram_competitor_insights`. Uses `arrayJoin(hashtags)` to explode the hashtag array. Engagement rate formula: `(total_engagement / total_followers / count) * 100`.

---

#### 6.5.8 `getBiographyQuery($sortOrder)`

Uses `instagram_competitor_insights` (not posts) for biography data:

```sql
SELECT * FROM (
    select last_value(biography) as biography,
           lengthUTF8(biography) as biography_length,
           instagram_account_id as business_account_id,
           multiIf(...) as state,
           multiIf(...) as slug
    from instagram_competitor_insights
    group by instagram_account_id
    ORDER BY max(inserted_at) DESC
) as biography_statistics
LEFT JOIN (
    select argMax(profile_picture_url, inserted_at) as image,
           argMax(page_name, inserted_at) as name,
           argMax(total_followed_by_count, inserted_at) as followersCount,
           instagram_account_id
    from instagram_competitor_insights
    group by instagram_account_id
) as page_constants ON page_constants.instagram_account_id = biography_statistics.business_account_id
ORDER BY {sortOrder} DESC
```

---

## 7. LinkedIn Analytics

### 7.1 Route Map

| Route | Method |
|-------|--------|
| `POST overview/linkedin/summary` | `getOverviewSummary` |
| `POST overview/linkedin/audienceGrowth` | `getOverviewAudienceGrowth` |
| `POST overview/linkedin/pageViews` | `getOverviewPageViews` |
| `POST overview/linkedin/publishingBehaviour` | `getOverviewPublishingBehaviour` |
| `POST overview/linkedin/topPosts` | `getOverviewTopPosts` |
| `POST overview/linkedin/postsPerDays` | `getOverviewDailyPosts` |
| `POST overview/linkedin/hashtags` | `getOverviewHashtags` |
| `POST overview/linkedin/getTopPosts` | `TopPosts15` |
| `POST overview/linkedin/followersDemographics` | `getOverviewfollowersDemographics` |

### 7.2 Summary (`summaryQuery`)

**Tables:** `linkedin_posts`, `linkedin_insights`

```sql
WITH posts AS (
    SELECT
        post_id,
        max(saving_time)
    FROM linkedin_posts
    WHERE linkedin_id IN ('...')
      AND toDateTime(published_at, 0, '{timezone}') BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
    GROUP BY post_id
)
SELECT
    {liId}
    toInt32(coalesce(post.post_comments, 0))        AS post_comments,
    toInt32(coalesce(post.post_favorites, 0))        AS post_likes,
    toInt32(coalesce(post.total_engagement, 0))      AS total_engagement,
    toInt32(coalesce(post.total_posts, 0))           AS total_posts,
    toInt32(coalesce(post.post_shares, 0))           AS post_shares,
    toInt32(coalesce(post.post_clicks, 0))           AS post_clicks,
    toInt32(coalesce(insights.total_follower_count, 0)) AS followers,
    toInt32(coalesce(insights.page_views, 0))           AS page_views,
    toInt32(coalesce(insights.page_reach, 0))           AS page_reach,
    toInt32(coalesce(insights.page_shares, 0))          AS page_shares,
    toInt32(coalesce(insights.page_comments, 0))        AS page_comments,
    toInt32(coalesce(insights.page_reactions, 0))       AS page_reactions,
    toInt32(coalesce(insights.total_impressions, 0))    AS page_impressions,
    toInt32(coalesce(insights.unique_visitors, 0))      AS page_unique_visitors,
    round(coalesce(insights.engagement_rate, 0), 2)     AS engagement_rate,
    round(coalesce(post.post_engagement_rate, 0), 2)    AS post_engagement_rate
FROM
(
    SELECT toString(c1) AS linkedin_id
    FROM VALUES ('id1','id2',...)
) AS page_ids
LEFT JOIN
(
    SELECT
        linkedin_id,
        post_comments, post_favorites, post_clicks, post_shares,
        post_comments + post_favorites + post_shares AS total_engagement,
        post_engagement_rate, total_posts
    FROM
    (
        SELECT
            linkedin_id,
            SUM(comments)                           AS post_comments,
            SUM(favorites)                          AS post_favorites,
            SUM(post_clicks)                        AS post_clicks,
            SUM(repost)                             AS post_shares,
            sum(total_engagement * impressions) / nullIf(sum(impressions), 0) AS post_engagement_rate,
            COUNT(*)                                AS total_posts
        FROM linkedin_posts
        WHERE (post_id, saving_time) IN posts
        GROUP BY linkedin_id
    )
) AS post USING linkedin_id
LEFT JOIN
(
    SELECT
        linkedin_id AS platform_id,
        argMax(latest_follower_count, latest_date)      AS total_follower_count,
        sum(daily_page_views)                           AS page_views,
        sum(daily_reach)                                AS page_reach,
        sum(daily_repost)                               AS page_shares,
        sum(daily_comments)                             AS page_comments,
        sum(daily_reactions)                            AS page_reactions,
        sum(daily_impressions)                          AS total_impressions,
        sum(daily_engagement_impressions) / nullIf(sum(daily_impressions), 0) AS engagement_rate,
        sum(daily_unique_visitors)                      AS unique_visitors
    FROM (
        SELECT
            linkedin_id,
            toDate(created_at)                          AS latest_date,
            argMin(totalFollowerCount, inserted_at)     AS latest_follower_count,
            sum(page_views)                             AS daily_page_views,
            sum(reach)                                  AS daily_reach,
            sum(repost)                                 AS daily_repost,
            sum(comments)                               AS daily_comments,
            sum(reactions)                              AS daily_reactions,
            sum(impressionCount)                        AS daily_impressions,
            sum(engagement * impressionCount)           AS daily_engagement_impressions,
            sum(unique_visitors)                        AS daily_unique_visitors
        FROM linkedin_insights
        WHERE linkedin_id IN ('...')
          AND toDateTime(created_at, 0, '{timezone}') BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
        GROUP BY linkedin_id, toDate(created_at)
    )
    GROUP BY linkedin_id
) AS insights ON page_ids.linkedin_id = insights.platform_id
{group}
```

**Key business logic:**
- `total_engagement = comments + favorites + shares` (clicks excluded)
- Engagement rate is impression-weighted: `sum(engagement * impressionCount) / sum(impressionCount)`
- Follower count uses `argMax(latest_follower_count, latest_date)` from insights

---

### 7.3 Last Follower Counts (Fallback) (`GetLastFollowerCounts`)

**Table:** `linkedin_insights`

```sql
SELECT
    arrayFirst(x -> x != 0, groupArray(totalFollowerCount))   AS total_follower_count,
    arrayFirst(x -> x != 0, groupArray(organicFollowerCount)) AS organic_follower_count,
    arrayFirst(x -> x != 0, groupArray(paidFollowerCount))    AS paid_follower_count
FROM
(
    SELECT
        argMin(created_at, inserted_at)                   AS inserted_time,
        toInt32(argMin(totalFollowerCount, inserted_at))  AS totalFollowerCount,
        toInt32(argMin(organicFollowerCount, inserted_at)) AS organicFollowerCount,
        toInt32(argMin(paidFollowerCount, inserted_at))   AS paidFollowerCount
    FROM
        linkedin_insights
    WHERE linkedin_id IN ('...')
      AND toDateTime(created_at, 0, '{timezone}') BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
    GROUP BY record_id
    ORDER BY inserted_time DESC
)
```

---

### 7.4 Audience Growth (`audienceQuery`)

**Table:** `linkedin_insights`

```sql
SELECT
    notEmpty(total_follower_count_temp) AS show_data,
    arrayFill(x -> not x == 0, organic_follower_count_temp) AS organic_follower_count,
    arrayMap(x -> toInt32(x),
        arrayPushFront(
            arrayDifference(arrayFill(x -> not x == 0, organic_follower_count_temp)),
            toInt32(0)
        )
    ) AS organic_followers_daily,
    arrayFill(x -> not x == 0, paid_follower_count_temp) AS paid_follower_count,
    arrayMap(x -> toInt32(x),
        arrayPushFront(
            arrayDifference(arrayFill(x -> not x == 0, paid_follower_count_temp)),
            toInt32(0)
        )
    ) AS paid_followers_daily,
    arrayFill(x -> not x == 0, total_follower_count_temp) AS total_follower_count,
    arrayMap(x -> toInt32(x),
        arrayPushFront(
            arrayDifference(arrayFill(x -> not x == 0, total_follower_count_temp)),
            toInt32(0)
        )
    ) AS total_followers_daily,
    buckets
FROM
(
    SELECT
        groupArray(dates)               AS buckets,
        groupArray(organicFollowerCount) AS organic_follower_count_temp,
        groupArray(paidFollowerCount)    AS paid_follower_count_temp,
        groupArray(totalFollowerCount)   AS total_follower_count_temp
    FROM
    (
        SELECT
            toDate(created_at)                              AS dates,
            toInt32(argMin(organicFollowerCount, inserted_at)) AS organicFollowerCount,
            toInt32(argMin(paidFollowerCount, inserted_at))    AS paidFollowerCount,
            toInt32(argMin(totalFollowerCount, inserted_at))   AS totalFollowerCount
        FROM linkedin_insights
        WHERE linkedin_id IN ('...')
            AND toDateTime(created_at, 0, '{timezone}') BETWEEN ...
        GROUP BY dates
        ORDER BY dates ASC
        WITH FILL TO toDate('{currentEndDate}') STEP 1
    )
)
```

**Audience Rollup (`audienceQueryRollup`):**

```sql
SELECT
    toInt32(last_value(organicFollowerCount)) AS organic_follower_count,
    toInt32(last_value(paidFollowerCount))    AS paid_follower_count,
    toInt32(last_value(totalFollowerCount))   AS total_follower_count,
    round(AVG(totalFollowerCount), 2)          AS avg_follower_count
FROM
(
    SELECT
        argMin(organicFollowerCount, inserted_at) AS organicFollowerCount,
        argMin(paidFollowerCount, inserted_at)    AS paidFollowerCount,
        argMin(totalFollowerCount, inserted_at)   AS totalFollowerCount,
        toDate(created_at) AS dates
    FROM linkedin_insights
    WHERE linkedin_id IN ('...')
      AND toDateTime(created_at, 0, '{timezone}') BETWEEN ...
    GROUP BY dates
    ORDER BY dates ASC
)
```

---

### 7.5 Page Views (`pageViewsQuery`)

**Table:** `linkedin_insights`

```sql
SELECT
    arrayMap(x -> toInt32(x), arrayCumSum(desktop_views_daily))  AS desktop_page_views,
    arrayMap(x -> toInt32(x), arrayCumSum(mobile_views_daily))   AS mobile_page_views,
    arrayMap(x -> toInt32(x), arrayCumSum(total_views_daily))    AS total_page_views,
    arrayMap(x -> toInt32(x), desktop_views_daily)               AS desktop_page_views_daily,
    arrayMap(x -> toInt32(x), mobile_views_daily)                AS mobile_page_views_daily,
    arrayMap(x -> toInt32(x), total_views_daily)                 AS total_page_views_daily,
    toInt32(arraySum(total_views_daily)) AS show_data,
    buckets
FROM
(
    SELECT
        groupArray(dates)          AS buckets,
        groupArray(desktop_views)  AS desktop_views_daily,
        groupArray(mobile_views)   AS mobile_views_daily,
        groupArray(total_views)    AS total_views_daily
    FROM
    (
        SELECT
            toDate(created_at)                    AS dates,
            toInt32(SUM(desktop_page_views))      AS desktop_views,
            toInt32(SUM(mobile_page_views))       AS mobile_views,
            toInt32(SUM(page_views))              AS total_views
        FROM linkedin_insights
        WHERE linkedin_id IN ('...')
          AND toDateTime(created_at, 0, '{timezone}') BETWEEN ...
        GROUP BY dates
        ORDER BY dates ASC
        WITH FILL TO toDate('{currentEndDate}') STEP 1
    )
)
```

Returns both cumulative (`arrayCumSum`) and daily arrays. No forward-fill applied (zeros stay as zeros).

**Page Views Rollup (`pageViewsRollupQuery`):**

```sql
SELECT
    toInt32(SUM(total_views))    AS total_page_views,
    toInt32(SUM(desktop_views))  AS desktop_page_views,
    toInt32(SUM(mobile_views))   AS mobile_page_views,
    round(AVG(total_views), 2)   AS avg_page_views
FROM
(
    SELECT
        toInt32(SUM(page_views))          AS total_views,
        toInt32(SUM(desktop_page_views))  AS desktop_views,
        toInt32(SUM(mobile_page_views))   AS mobile_views
    FROM linkedin_insights
    WHERE linkedin_id IN ('...')
      AND toDateTime(created_at, 0, '{timezone}') BETWEEN ...
    GROUP BY toDate(created_at)
    ORDER BY toDate(created_at) ASC
)
```

---

### 7.6 Engagements (`engagementsQuery`)

**Table:** `linkedin_posts`

**Notable side-effect:** This method mutates `currentEndDate` via `subDay()` before building the query.

```sql
WITH posts AS (
    SELECT post_id, max(saving_time)
    FROM linkedin_posts
    WHERE linkedin_id IN ('...')
      AND toDateTime(published_at, 0, '{timezone}') BETWEEN ...
    GROUP BY post_id
)
SELECT
    groupArray(dates)           AS buckets,
    groupArray(comment)         AS comments,
    groupArray(favorite)        AS favorites,
    groupArray(totalEngagement) AS total_engagement,
    groupArray(count)           AS doc_count,
    toInt32(sum(totalEngagement)) AS show_data
FROM
(
    SELECT
        toInt32(count())              AS count,
        toInt32(SUM(comments))        AS comment,
        toInt32(SUM(favorites))       AS favorite,
        toInt32(SUM(total_engagement)) AS totalEngagement,
        toDate(created_at)            AS dates
    FROM linkedin_posts
    WHERE (post_id, saving_time) IN posts
    GROUP BY toDate(created_at)
    ORDER BY toDate(created_at) ASC
    WITH FILL TO toDate('{currentEndDate - 1 day}')
)
```

**Engagements Rollup (`engagementsRollupQuery`):**

```sql
WITH posts AS (
    SELECT post_id, max(saving_time)
    FROM linkedin_posts
    WHERE linkedin_id IN ('...')
      AND {DateFilter('published_at')}
    GROUP BY post_id
)
SELECT
    toInt32(SUM(comment))               AS comments,
    toInt32(SUM(favorite))              AS favorites,
    toInt32(SUM(repost))                AS shares,
    toInt32(favorites + comments + shares) AS total_engagement,
    avg(totalEngagement)                AS avg_engagement,
    toInt32(sum(count))                 AS doc_count
FROM (
    SELECT
        count()                       AS count,
        SUM(comments)                 AS comment,
        SUM(favorites)                AS favorite,
        SUM(repost)                   AS repost,
        SUM(total_engagement)         AS totalEngagement
    FROM linkedin_posts
    WHERE (post_id, saving_time) IN posts
    GROUP BY created_at
)
```

---

### 7.7 Publishing Behaviour (`publishingBehaviourQuery`)

**Table:** `linkedin_posts`

```sql
WITH posts AS (
    SELECT post_id, last_value(saving_time) AS saving_time
    FROM linkedin_posts
    WHERE linkedin_id IN ('...')
      AND {DateFilter('published_at')}
      AND media_type IN ('text','images','videos','carousel','link')
    GROUP BY post_id
)
SELECT
    groupArray(likes)          AS likes,
    groupArray(comments)       AS comments,
    groupArray(shares)         AS shares,
    groupArray(clicks)         AS clicks,
    groupArray(engagement_rate) AS engagement_rate,
    groupArray(impression)     AS impressions,
    groupArray(total_posts)    AS total_posts,
    groupArray(engagements)    AS engagement,
    groupArray(reach)          AS reach,
    groupArray(created_at)     AS buckets
FROM
(
    SELECT
        toInt32(SUM(like_count))        AS likes,
        toInt32(SUM(comments_count))    AS comments,
        toInt32(SUM(shares_count))      AS shares,
        toInt32(SUM(clicks_count))      AS clicks,
        toInt32(SUM(impressions))       AS impression,
        toInt32(COUNT(*))               AS total_posts,
        toInt32(likes + comments + shares) AS engagements,
        if(SUM(impressions) > 0,
            toFloat32(round(100 * engagements / impression, 2)),
            0
        ) AS engagement_rate,
        toInt32(SUM(reach))             AS reach,
        toDate(created_at)              AS created_at
    FROM
    (
        SELECT
            post_id,
            last_value(favorites)   AS like_count,
            last_value(comments)    AS comments_count,
            last_value(repost)      AS shares_count,
            last_value(post_clicks) AS clicks_count,
            last_value(impressions) AS impressions,
            last_value(reach)       AS reach,
            created_at
        FROM linkedin_posts
        WHERE (post_id, saving_time) IN posts
        GROUP BY post_id, created_at
    )
    GROUP BY created_at
    ORDER BY created_at ASC
    WITH FILL TO toDate('{currentEndDate}')
)
```

`engagement_rate = 100 * (likes + comments + shares) / impressions`

**Publishing Behaviour Rollup (`publishingBehaviourRollupQuery`):**

Uses virtual `media_types` table to ensure all 5 types always appear. UNION ALL adds a `'total'` row.

```sql
WITH
posts AS (
    SELECT post_id, max(saving_time)
    FROM linkedin_posts
    WHERE linkedin_id IN ('...')
      AND {DateFilter('published_at')}
      AND media_type != ''
    GROUP BY post_id
),
media_types AS (
    SELECT arrayJoin(['text', 'images', 'videos', 'link', 'carousel']) AS media_type
),
metrics AS (
    SELECT
        mt.media_type,
        toInt32(COALESCE(total_posts, 0))  AS total_posts,
        toInt32(COALESCE(likes, 0))        AS likes,
        toInt32(COALESCE(comments, 0))     AS comments,
        toInt32(COALESCE(shares, 0))       AS shares,
        toInt32(COALESCE(clicks, 0))       AS clicks,
        toInt32(COALESCE(engagements, 0))  AS engagements,
        toInt32(COALESCE(impressions, 0))  AS impressions,
        toInt32(COALESCE(reach, 0))        AS reach
    FROM media_types mt
    LEFT JOIN
    (
        SELECT
            media_type,
            toInt32(COUNT(*))        AS total_posts,
            toInt32(SUM(like_count)) AS likes,
            toInt32(SUM(comments_count)) AS comments,
            toInt32(SUM(shares_count)) AS shares,
            toInt32(SUM(clicks_count)) AS clicks,
            toInt32(likes + comments + shares) AS engagements,
            toInt32(SUM(impressions)) AS impressions,
            toInt32(SUM(reach))      AS reach
        FROM
        (
            SELECT
                post_id,
                last_value(media_type)      AS media_type,
                last_value(favorites)       AS like_count,
                last_value(comments)        AS comments_count,
                last_value(repost)          AS shares_count,
                last_value(post_clicks)     AS clicks_count,
                last_value(total_engagement) AS engagement,
                last_value(impressions)     AS impressions,
                last_value(reach)           AS reach,
                created_at
            FROM linkedin_posts
            WHERE (post_id, saving_time) IN posts
            GROUP BY post_id, created_at
        )
        GROUP BY media_type
    ) t ON mt.media_type = t.media_type
)
SELECT * FROM (
    SELECT * FROM metrics
    UNION ALL
    SELECT
        'total'                     AS media_type,
        toInt32(SUM(total_posts))   AS total_posts,
        toInt32(SUM(likes))         AS likes,
        toInt32(SUM(comments))      AS comments,
        toInt32(SUM(shares))        AS shares,
        toInt32(SUM(clicks))        AS clicks,
        toInt32(SUM(engagements))   AS engagements,
        toInt32(SUM(impressions))   AS impressions,
        toInt32(SUM(reach))         AS reach
    FROM metrics
)
ORDER BY CASE WHEN media_type = 'total' THEN 1 ELSE 0 END, media_type
```

---

### 7.8 Top Posts (`topPostQuery`)

**Table:** `linkedin_posts`

```sql
WITH posts AS (
    SELECT post_id, max(saving_time)
    FROM linkedin_posts
    WHERE linkedin_id IN ('...')
      AND {DateFilter('published_at')}
      [AND has(hashtags, 'tag1')]
    GROUP BY post_id
)
SELECT *
FROM linkedin_posts
WHERE (post_id, saving_time) IN posts
ORDER BY {order_by} DESC
LIMIT {limit}
```

Parameters: `limit` (3 for overview, 15 for full list), `order_by` (default `'total_engagement'`), `hashtags` (optional filter -- note: multi-hashtag filter appears buggy).

---

### 7.9 Posts Per Day (`postsPerDayQuery`)

**Table:** `linkedin_posts`

```sql
WITH posts AS (
    SELECT post_id, max(saving_time)
    FROM linkedin_posts
    WHERE linkedin_id IN ('...')
      AND {DateFilter('published_at')}
    GROUP BY post_id
)
SELECT
    toInt32(countIf(day_of_week == 'Monday'))    AS Monday,
    toInt32(countIf(day_of_week == 'Tuesday'))   AS Tuesday,
    toInt32(countIf(day_of_week == 'Wednesday')) AS Wednesday,
    toInt32(countIf(day_of_week == 'Thursday'))  AS Thursday,
    toInt32(countIf(day_of_week == 'Friday'))    AS Friday,
    toInt32(countIf(day_of_week == 'Saturday'))  AS Saturday,
    toInt32(countIf(day_of_week == 'Sunday'))    AS Sunday
FROM linkedin_posts
WHERE (post_id, saving_time) IN posts
```

---

### 7.10 Top Hashtags (`topHashtagsQuery`)

**Table:** `linkedin_posts`

Uses `arrayJoin(lp.hashtags)` with INNER JOIN on deduplicated posts. Top 30 by engagement.

```sql
WITH posts AS (
    SELECT post_id, max(saving_time) AS max_saving_time
    FROM linkedin_posts
    WHERE linkedin_id IN ('...')
      AND {DateFilter('published_at')}
    GROUP BY post_id
)
SELECT
    groupArray(name)        AS name,
    groupArray(engagements) AS engagements,
    groupArray(likes)       AS likes,
    groupArray(comments)    AS comments,
    groupArray(shares)      AS shares,
    groupArray(posts)       AS posts
FROM
(
    SELECT
        arrayJoin(lp.hashtags)                 AS name,
        toInt32(SUM(lp.favorites))             AS likes,
        toInt32(SUM(lp.comments))              AS comments,
        toInt32(SUM(lp.repost))                AS shares,
        toInt32(likes + comments + shares)     AS engagements,
        toInt32(count(DISTINCT lp.post_id))    AS posts
    FROM linkedin_posts AS lp
    INNER JOIN posts AS p
        ON lp.post_id = p.post_id AND lp.saving_time = p.max_saving_time
    GROUP BY name
    ORDER BY engagements DESC
    LIMIT 30
)
```

**Hashtags Rollup (`topHashtagsQueryRollup`):**

```sql
WITH posts AS (
    SELECT post_id, max(saving_time) AS max_saving_time
    FROM linkedin_posts
    WHERE linkedin_id IN ('...')
      AND {DateFilter('published_at')}
    GROUP BY post_id
),
post_data AS (
    SELECT lp.hashtags, lp.favorites, lp.comments, lp.repost, lp.impressions, lp.reach
    FROM linkedin_posts AS lp
    INNER JOIN posts AS p
        ON lp.post_id = p.post_id AND lp.saving_time = p.max_saving_time
    WHERE length(lp.hashtags) > 0
)
SELECT
    toInt32((SELECT COUNT(DISTINCT arrayJoin(hashtags)) FROM post_data)) AS total_hashtags,
    toInt32(SUM(length(hashtags)))   AS total_times_used,
    toInt32(SUM(favorites))          AS total_likes,
    toInt32(SUM(comments))           AS total_comments,
    toInt32(SUM(repost))             AS total_shares,
    toInt32(total_likes + total_comments + total_shares) AS total_engagement,
    toInt32(SUM(impressions))        AS total_impressions,
    toInt32(SUM(reach))              AS total_reach
FROM post_data
```

---

### 7.11 Followers Demographics (`followersDemographicQuery`)

**Table:** `contentstudiobackend.linkedin_insights` (fully qualified table name)

LinkedIn stores demographics as JSON. Uses `JSONExtractKeysAndValues(column, 'String')` to parse. Categories: seniority, industry, country, city.

```sql
WITH
latest_record AS (
    SELECT
        followers_by_seniority, followers_by_industry,
        followers_by_country, followers_by_city,
        totalFollowerCount
    FROM contentstudiobackend.linkedin_insights
    WHERE linkedin_id IN ('...')
      AND toDateTime(created_at, 0, '{timezone}') BETWEEN ...
    ORDER BY created_at DESC
    LIMIT 1
),
processed_data AS (
    SELECT
        arrayMap(
            x -> (x.1, toInt32(JSONExtractInt(followers_by_seniority, x.1))),
            JSONExtractKeysAndValues(followers_by_seniority, 'String')
        ) AS seniority_data,
        arrayMap(
            x -> (x.1, toInt32(JSONExtractInt(followers_by_industry, x.1))),
            JSONExtractKeysAndValues(followers_by_industry, 'String')
        ) AS industry_data,
        arrayMap(
            x -> (x.1, toInt32(JSONExtractInt(followers_by_country, x.1))),
            JSONExtractKeysAndValues(followers_by_country, 'String')
        ) AS country_data,
        arrayMap(
            x -> (x.1, toInt32(JSONExtractInt(followers_by_city, x.1))),
            JSONExtractKeysAndValues(followers_by_city, 'String')
        ) AS city_data,
        totalFollowerCount
    FROM latest_record
)
SELECT
    map(
        'buckets',
        arrayConcat(
            IF(arraySum(arrayMap(x -> x.2, seniority_data)) < toInt32(totalFollowerCount),
               ['Others'], []),
            arrayReverse(arrayMap(x -> x.1, seniority_data))
        ),
        'values',
        arrayConcat(
            IF(arraySum(arrayMap(x -> x.2, seniority_data)) < toInt32(totalFollowerCount),
               [toString(toInt32(totalFollowerCount) - arraySum(arrayMap(x -> x.2, seniority_data)))],
               []),
            arrayReverse(arrayMap(x -> toString(x.2), seniority_data))
        )
    ) AS seniority,
    map(
        'buckets',
        arrayConcat(
            IF(arraySum(arrayMap(x -> x.2, industry_data)) < toInt32(totalFollowerCount),
               ['Others'], []),
            arrayReverse(arrayMap(x -> x.1, industry_data))
        ),
        'values',
        arrayConcat(
            IF(arraySum(arrayMap(x -> x.2, industry_data)) < toInt32(totalFollowerCount),
               [toString(toInt32(totalFollowerCount) - arraySum(arrayMap(x -> x.2, industry_data)))], []),
            arrayReverse(arrayMap(x -> toString(x.2), industry_data))
        )
    ) AS industry,
    map(
        'buckets',
        arrayConcat(
            IF(arraySum(arrayMap(x -> x.2, country_data)) < toInt32(totalFollowerCount),
               ['Others'], []),
            arrayReverse(arrayMap(x -> x.1, country_data))
        ),
        'values',
        arrayConcat(
            IF(arraySum(arrayMap(x -> x.2, country_data)) < toInt32(totalFollowerCount),
               [toString(toInt32(totalFollowerCount) - arraySum(arrayMap(x -> x.2, country_data)))], []),
            arrayReverse(arrayMap(x -> toString(x.2), country_data))
        )
    ) AS country,
    map(
        'buckets',
        arrayConcat(
            IF(arraySum(arrayMap(x -> x.2, city_data)) < toInt32(totalFollowerCount),
               ['Others'], []),
            arrayReverse(arrayMap(x -> x.1, city_data))
        ),
        'values',
        arrayConcat(
            IF(arraySum(arrayMap(x -> x.2, city_data)) < toInt32(totalFollowerCount),
               [toString(toInt32(totalFollowerCount) - arraySum(arrayMap(x -> x.2, city_data)))], []),
            arrayReverse(arrayMap(x -> toString(x.2), city_data))
        )
    ) AS city
FROM processed_data
```

Each category includes an "Others" entry if sum of known categories is less than totalFollowerCount.

---

### 7.12 Time Recommendation (`timeRecommendationQuery`)

**Table:** `linkedin_posts`

```sql
SELECT
    max(linkedin_id)         AS page_id,
    day_of_week,
    hour_of_day,
    0                        AS post_impressions,
    sum(total_engagement)    AS total_engagement
FROM linkedin_posts
WHERE linkedin_id IN ('...')
  AND toDateTime(published_at, 0, '{timezone}') BETWEEN ...
GROUP BY day_of_week, hour_of_day
ORDER BY day_of_week, hour_of_day
```

Note: `post_impressions` is hardcoded 0 (LinkedIn does not expose per-slot impression data).

---

## 8. YouTube Analytics

### 8.1 Route Map

| Route | Method |
|-------|--------|
| `POST overview/youtube/overviewSummary` | `overviewSummary` |
| `POST overview/youtube/overviewSubscriberTrend` | `overviewSubscriberTrend` |
| `POST overview/youtube/overviewEngagementTrend` | `overviewEngagementTrend` |
| `POST overview/youtube/overviewViewsTrend` | `overviewViewsTrend` |
| `POST overview/youtube/overviewWatchTimeTrend` | `overviewWatchTimeTrend` |
| `POST overview/youtube/overviewFindVideo` | `overviewFindVideo` |
| `POST overview/youtube/overviewVideoSharing` | `overviewVideoSharing` |
| `POST overview/youtube/overviewTopPosts` | `overviewTopPosts` |
| `POST overview/youtube/overviewLeastPosts` | `overviewLeastPosts` |
| `POST overview/youtube/overviewPerformanceAndVideoPostingSchedule` | `overviewPerformanceAndVideoPostingSchedule` |
| `POST overview/youtube/getSortedTopPosts` | `getSortedTopPosts` |

### 8.2 Summary (`summaryQuery`)

Complex 4-way LEFT JOIN across `youtube_activity_insights`, `youtube_channels`, `youtube_videos`, and views. Returns `'N/A'` string when metrics have no data. Subscriber count falls back to latest available if no data exists in the date range.

```sql
SELECT
    channel_id,
    if(count_stats > 0, toString(watch_time), 'N/A') as watch_time,
    if(count_stats > 0, toString(avg_view_duration), 'N/A') as avg_view_duration,
    if(count_stats > 0, toString(like), 'N/A') as like,
    if(count_stats > 0, toString(dislike), 'N/A') as dislike,
    if(count_stats > 0, toString(comment), 'N/A') as comment,
    if(count_stats > 0, toString(share), 'N/A') as share,
    if(count_stats > 0, toString(engagement), 'N/A') as engagement,
    if(count_subs > 0, toString(subscribers), 'N/A') as subscribers,
    if(count_views > 0, toString(video_views), 'N/A') as views,
    if(count_videos > 0, toString(videos), 'N/A') as videos
FROM
(
    SELECT
        channel_ids.channel_id as channel_id,
        coalesce(channel_stats.count_stats, 0) as count_stats,
        coalesce(channel_stats.watch_time, 0) as watch_time,
        coalesce(channel_stats.avg_view_duration, 0) as avg_view_duration,
        coalesce(channel_stats.like, 0) as like,
        coalesce(channel_stats.dislike, 0) as dislike,
        coalesce(channel_stats.comment, 0) as comment,
        coalesce(channel_stats.share, 0) as share,
        coalesce(channel_stats.engagement, 0) as engagement,
        coalesce(subs_stats.count_subs, 0) as count_subs,
        coalesce(subs_stats.subscribers, 0) as subscribers,
        coalesce(video_stats.count_videos, 0) as count_videos,
        coalesce(video_stats.videos, 0) as videos,
        coalesce(views_stats.count_views, 0) as count_views,
        coalesce(views_stats.video_views, 0) as video_views
    FROM
    (
        SELECT c1 as channel_id FROM VALUES {youtube_id_placeholder}
    ) AS channel_ids
    LEFT JOIN
    (
        -- channel_stats: activity insights aggregation
        SELECT
            channel_id,
            count_stats,
            watch_time,
            round(if(total_views > 0, total_est_watch_time/total_views, 0), 2) as avg_view_duration,
            like, dislike, comment, share, engagement
        FROM
        (
            SELECT
                channel_id,
                count() as count_stats,
                sum(estimated_minute_watched) as watch_time,
                sum(view_count) as total_views,
                sum(est_watch_time) as total_est_watch_time,
                sum(likes) as like,
                sum(dislikes) as dislike,
                sum(comments) as comment,
                sum(shares) as share,
                sum(likes + dislikes + comments + shares) as engagement
            FROM
            (
                SELECT
                    last_value(channel_id) as channel_id,
                    last_value(estimated_minutes_watched) as estimated_minute_watched,
                    last_value(estimated_minutes_watched)*60 as est_watch_time,
                    last_value(average_view_duration) as average_view_duration,
                    last_value(views) as view_count,
                    last_value(likes) as likes,
                    last_value(dislikes) as dislikes,
                    last_value(comments) as comments,
                    last_value(shares) as shares,
                    last_value(created_at) as created_at
                FROM youtube_activity_insights
                GROUP BY record_id
            )
            WHERE channel_id in {youtube_id_placeholder}
            AND {date_filter_created_at}
            GROUP BY channel_id
        )
    ) as channel_stats ON channel_ids.channel_id = channel_stats.channel_id
    LEFT JOIN
    (
        -- subs_stats: subscriber count with date-range fallback
        SELECT
            channel_id,
            if(count_in_range > 0, count_in_range, count_total) as count_subs,
            toInt32(if(count_in_range > 0, subscribers_in_range, subscribers_latest)) as subscribers
        FROM
        (
            SELECT
                channel_id,
                countIf(toDateTime(created_at) BETWEEN toDateTime('{start_date}', 0)
                    AND toDateTime('{end_date}', 0) + INTERVAL 1 DAY) as count_in_range,
                argMaxIf(subscriber_count, created_at,
                    toDateTime(created_at) BETWEEN toDateTime('{start_date}', 0)
                    AND toDateTime('{end_date}', 0) + INTERVAL 1 DAY) as subscribers_in_range,
                count(*) as count_total,
                argMax(subscriber_count, created_at) as subscribers_latest
            FROM youtube_channels
            WHERE channel_id in {youtube_id_placeholder}
            GROUP BY channel_id
        )
    ) as subs_stats ON channel_ids.channel_id = subs_stats.channel_id
    LEFT JOIN
    (
        -- video_stats: video count in date range
        SELECT
            channel_id,
            count(*) as count_videos,
            toInt32(count(*)) as videos
        FROM
        (
            SELECT channel_id, video_id
            FROM youtube_videos
            WHERE channel_id in {youtube_id_placeholder}
            AND {date_filter_published_at}
            GROUP BY channel_id, video_id
        )
        GROUP BY channel_id
    ) as video_stats ON channel_ids.channel_id = video_stats.channel_id
    LEFT JOIN
    (
        -- views_stats: total video views from activity insights
        SELECT
            channel_id,
            count(*) as count_views,
            toInt32(sum(view_count)) as video_views
        FROM
        (
            SELECT
                last_value(channel_id) as channel_id,
                last_value(views) as view_count,
                last_value(created_at) as created_at
            FROM youtube_activity_insights
            GROUP BY record_id
        )
        WHERE channel_id in {youtube_id_placeholder}
        AND {date_filter_created_at}
        GROUP BY channel_id
    ) as views_stats ON channel_ids.channel_id = views_stats.channel_id
)
```

---

### 8.3 Subscriber Trend (`subscriberTrendQuery`)

**Table:** `youtube_channels`

```sql
SELECT
    arrayDifference(subscribers_total) AS subscribers_gained_daily,
    subscribers_total,
    buckets
FROM
(
    SELECT
        arrayReverseFill(x -> not x==0, arrayFill(x -> not x==0, subscribers_gained)) AS subscribers_total,
        buckets
    FROM
    (
        SELECT
            groupArray(subscribers_gained_total) as subscribers_gained,
            groupArray(dates) as buckets
        FROM
        (
            SELECT toInt32(max(subscribers_gained)) as subscribers_gained_total, dates
            FROM
            (
                SELECT argMin(subscriber_count, inserted_at) as subscribers_gained,
                       toDate(inserted_at) as dates, channel_id
                FROM youtube_channels
                WHERE channel_id in ('...')
                AND toDateTime(inserted_at) >= toDateTime('START_DATE',0)
                  AND toDateTime(inserted_at) < toDateTime('END_DATE+1',0)
                GROUP BY dates, channel_id
            )
            GROUP BY dates
            ORDER BY dates ASC
            WITH FILL FROM toDate('START_DATE') TO toDate('END_DATE') + 1 STEP 1
        )
    )
)
```

Uses both `arrayFill` (forward-fill) and `arrayReverseFill` (backward-fill). Controller applies additional leading-zero replacement via `getLatestSubscriberCount()` fallback.

**Fallback (`getLatestSubscriberCount`):**

```sql
SELECT toInt32(subscriber_count) as subscriber_count
FROM youtube_channels
WHERE channel_id IN ('...')
AND subscriber_count > 0
ORDER BY inserted_at DESC
LIMIT 1
```

**Dynamic version (`subscriberDynamicTrendQuery`):** 180-day threshold. Uses `last_value(subscriber_count)` instead of `argMin`, groups by `record_id`.

---

### 8.4 Engagement Trend (`engagementTrendQuery`)

**Table:** `youtube_activity_insights`

```sql
SELECT
    arrayMap(i -> toInt32(i), arrayCumSum(like)) as like_total,
    like as like_daily,
    arrayMap(i -> toInt32(i), arrayCumSum(dislike)) as dislike_total,
    dislike as dislike_daily,
    arrayMap(i -> toInt32(i), arrayCumSum(share)) as share_total,
    share as share_daily,
    arrayMap(i -> toInt32(i), arrayCumSum(comment)) as comment_total,
    comment as comment_daily,
    arrayMap(i -> toInt32(i), arrayCumSum(engagement)) as engagement_total,
    engagement as engagement_daily,
    bucket
FROM
(
    SELECT
        groupArray(like) as like, groupArray(dislike) as dislike,
        groupArray(share) as share, groupArray(comment) as comment,
        groupArray(engagement) as engagement, groupArray(created_at) as bucket
    FROM
    (
        SELECT
            toInt32(sum(likes)) as like, toInt32(sum(dislikes)) as dislike,
            toInt32(sum(shares)) as share, toInt32(sum(comments)) as comment,
            toInt32(sum(likes + dislikes + shares + comments)) AS engagement,
            toDate(created_date) as created_at
        FROM
        (
            SELECT
                last_value(created_at) as created_date,
                last_value(likes) as likes, last_value(dislikes) as dislikes,
                last_value(shares) as shares, last_value(comments) as comments
            FROM youtube_activity_insights
            WHERE channel_id in ('...')
            AND [DateFilter on created_at]
            GROUP BY record_id
        )
        GROUP BY created_at
        ORDER BY created_at ASC
        WITH FILL FROM toDate('ACCOUNT_CREATED_DATE') STEP 1
    )
)
```

Fill starts from `account_created_date`. Uses `arrayCumSum` for running totals.

---

### 8.5 Video Views Trend (`videoViewsTrendQuery`)

**Table:** `youtube_traffic_insights`

```sql
SELECT
    subscriber_views_total, subscriber_views_daily,
    non_subscriber_views_total, non_subscriber_views_daily,
    arrayMap((i, value) -> toInt32(subscriber_views_daily[i] + non_subscriber_views_daily[i]),
             arrayEnumerate(subscriber_views_daily), subscriber_views_daily) as video_views_daily,
    arrayMap((i, value) -> toInt32(subscriber_views_total[i] + non_subscriber_views_total[i]),
             arrayEnumerate(subscriber_views_total), subscriber_views_total) as video_views_total,
    buckets
FROM
(
    SELECT
        arrayMap(i -> toInt32(i), arrayCumSum(subscriber_views)) as subscriber_views_total,
        subscriber_views as subscriber_views_daily,
        arrayMap(i -> toInt32(i), arrayCumSum(non_subscriber_views)) as non_subscriber_views_total,
        non_subscriber_views as non_subscriber_views_daily,
        created_date as buckets
    FROM
    (
        SELECT
            groupArray(subscriber_views) as subscriber_views,
            groupArray(non_subscriber_views) as non_subscriber_views,
            groupArray(created_date) as created_date
        FROM
        (
            SELECT
                toInt32(sum(subscriber_views)) as subscriber_views,
                toInt32(sum(
                    paid_views + annotation_views + end_screen_views + campaign_card_view +
                    no_link_other_views + yt_channel_views + yt_search_views + related_video_views +
                    yt_other_page_views + ext_url_views + playlist_views + notification_views + shorts_views
                )) AS non_subscriber_views,
                toDate(created_at) as created_date
            FROM
            (
                SELECT
                    last_value(channel_id) AS channel_id,
                    last_value(subscriber_views) AS subscriber_views,
                    last_value(paid_views) AS paid_views,
                    last_value(annotation_views) AS annotation_views,
                    last_value(end_screen_views) AS end_screen_views,
                    last_value(campaign_card_view) AS campaign_card_view,
                    last_value(no_link_other_views) AS no_link_other_views,
                    last_value(yt_channel_views) AS yt_channel_views,
                    last_value(yt_search_views) AS yt_search_views,
                    last_value(related_video_views) AS related_video_views,
                    last_value(yt_other_page_views) AS yt_other_page_views,
                    last_value(ext_url_views) AS ext_url_views,
                    last_value(playlist_views) AS playlist_views,
                    last_value(notification_views) AS notification_views,
                    last_value(shorts_views) AS shorts_views,
                    last_value(created_at) AS created_at
                FROM youtube_traffic_insights
                GROUP BY record_id
            )
            WHERE channel_id in ('...')
            AND [DateFilter on created_at]
            GROUP BY created_at
            ORDER BY toDate(created_at) ASC
            WITH FILL FROM toDate('ACCOUNT_CREATED_DATE') STEP 1
        )
    )
)
```

Non-subscriber views = sum of 13 traffic sources (plus `shorts_views` = 14 total).

---

### 8.6 Watch Time Trend (`watchTimeTrendQuery`)

**Table:** `youtube_traffic_insights`

```sql
SELECT
    arrayMap(x->toInt32(x), arrayCumSum(subscriber_watch_time)) as subscriber_watch_time_total,
    arrayMap(x->toInt32(x), arrayCumSum(non_subscriber_watch_time)) as non_subscriber_watch_time_total,
    subscriber_watch_time as subscriber_watch_time_daily,
    non_subscriber_watch_time as non_subscriber_watch_time_daily,
    average_watch_time,
    buckets
FROM
(
    SELECT
        groupArray(sub_watch_time) as subscriber_watch_time,
        groupArray(non_sub_watch_time) as non_subscriber_watch_time,
        avg(average_watch_time) as average_watch_time,
        groupArray(created_at) as buckets
    FROM
    (
        SELECT
            toInt32(sum(subscriber_watch_time)) as sub_watch_time,
            toInt32(sum(non_subscriber_watch_time)) as non_sub_watch_time,
            toInt32(sum(subscriber_watch_time + non_subscriber_watch_time)) as average_watch_time,
            toDate(created_at) as created_at
        FROM
        (
            SELECT
                last_value(subscriber_watch_time) as subscriber_watch_time,
                last_value(non_subsciber_watch_time) as non_subscriber_watch_time,
                last_value(created_at) as created_at,
                last_value(channel_id) as channel_id
            FROM youtube_traffic_insights
            GROUP BY record_id
        )
        WHERE channel_id in ('...')
        AND [DateFilter on created_at]
        GROUP BY created_at
        ORDER BY created_at ASC
        WITH FILL FROM toDate('START_DATE') TO toDate('END_DATE') + 1 STEP 1
    )
)
```

**CRITICAL:** Column name `non_subsciber_watch_time` has a typo (missing 's' in subscriber). Must use exact typo in Go implementation.

---

### 8.7 Find Video (`findVideoQuery`)

**Table:** `youtube_traffic_insights`

13 traffic sources (excludes `shorts_views`). Returns rows sorted ascending by value with name, value, and percentage.

```sql
SELECT
    (arrayJoin(arrayZip(names[1], values[1], perc_values)) AS t).1 as name,
    t.2 as value, t.3 as perc_value
FROM
(
    SELECT names, values, arrayMap(x -> x*100/total, values[1]) AS perc_values
    FROM
    (
        SELECT
            groupArray([*]) as values,
            groupArray(['Paid Views', 'Annotation Views', 'End Screen Views', 'Campaign Card Views',
                        'Subscriber Views', 'No Link Other views', 'YT Channel Views', 'YT Search Views',
                        'Related Video Views', 'YT Other Page Views', 'Ext URL Views',
                        'Playlist Views', 'Notification Views']) as names,
            arraySum([*]) as total
        FROM
        (
            SELECT
                toInt32(sum(paid_views)) as `Paid View`, toInt32(sum(annotation_views)) as `Annotation View`,
                toInt32(sum(end_screen_views)) as `End Screen View`, toInt32(sum(campaign_card_view)) as `Campaign Card Views`,
                toInt32(sum(subscriber_views)) as `Subscriber View`, toInt32(sum(no_link_other_views)) as `No Link Other View`,
                toInt32(sum(yt_channel_views)) as `YT Channel View`, toInt32(sum(yt_search_views)) as `YT Search View`,
                toInt32(sum(related_video_views)) as `Related Video View`, toInt32(sum(yt_other_page_views)) as `YT Other Page View`,
                toInt32(sum(ext_url_views)) as `Ext URL View`, toInt32(sum(playlist_views)) as `Playlist View`,
                toInt32(sum(notification_views)) as `Notification View`
            FROM youtube_traffic_insights
            WHERE channel_id in ('...')
            AND [DateFilter on created_at]
        )
        GROUP BY *
    )
)
ORDER BY value ASC
```

---

### 8.8 Video Sharing (`videoSharingQuery`)

**Table:** `youtube_shared_insights`

31 sharing platforms. Only takes the single most recent record. Filters out zero-value platforms.

```sql
SELECT
    (arrayJoin(arrayZip(names[1], values[1], perc_values)) AS t).1 as name,
    t.2 as value, t.3 as perc_value
FROM
(
    SELECT names, values, arrayMap(x -> x*100/total, values[1]) AS perc_values
    FROM
    (
        SELECT groupArray([*]) as values,
            groupArray(['Ameba', 'Blogger', 'Copy Paste', 'Cyworld', 'Digg', 'Dropbox', 'Embed',
                        'Mail', 'Whatsapp', 'Other', 'facebook Messenger', 'Facebook Pages', 'Facebook',
                        'Fotka', 'Vkontakte', 'Google Plus', 'Discord', 'Linkedin', 'Goo', 'Hangouts',
                        'Pinterest', 'Myspace', 'Reddit', 'Skype', 'Telegram', 'Tumblr', 'Twitter',
                        'Viber', 'Weibo', 'Wechat', 'Youtube']) as names,
            arraySum([*]) as total
        FROM
        (
            SELECT toInt32(ameba) as `Ameba`, toInt32(blogger) as `Blogger`, ...all 31 platforms...
            FROM youtube_shared_insights
            WHERE channel_id in ('...')
            AND [DateFilter on inserted_at]
            ORDER BY inserted_at DESC
            LIMIT 1
        )
        GROUP BY *
    )
)
WHERE value != 0
ORDER BY value DESC
```

---

### 8.9 Top/Least Posts (`topPostsQuery`/`leastPostsQuery`)

**Table:** `youtube_videos`

```sql
SELECT *, published_at_time as published_at FROM
(
    SELECT
        video_id,
        last_value(title) as title,
        last_value(description) as description,
        last_value(duration) as duration,
        last_value(thumbnail_url) as thumbnail_url,
        last_value(media_type) as media_type,
        concat('https://', substring(
            iframe_embed_html,
            position('//' IN iframe_embed_html) + length('//'),
            position('"' IN substring(iframe_embed_html, position('//' IN iframe_embed_html) + length('//'))) - 1
        )) AS iframe_embed_url,
        REPLACE(iframe_embed_url, 'embed/', 'watch?v=') as share_url,
        toInt32(argMax(likes, inserted_at) + argMax(dislikes, inserted_at)
                + argMax(comments, inserted_at) + argMax(shares, inserted_at)) as engagement,
        engagement as total_engagement,
        toInt32(argMax(likes, inserted_at)) as like,
        toInt32(argMax(dislikes, inserted_at)) as dislike,
        toInt32(argMax(views, inserted_at)) as views,
        toInt32(argMax(red_views, inserted_at)) as red_views,
        toInt32(argMax(favorites, inserted_at)) as favorites,
        toInt32(argMax(comments, inserted_at)) as comment,
        toInt32(argMax(subscribers_gained, inserted_at)) as subscribers_gained,
        toInt32(argMax(shares, inserted_at)) as share,
        toInt32(argMax(minutes_watched, inserted_at)) as minutes_watched,
        toInt32(argMax(red_minutes_watched, inserted_at)) as red_minutes_watched,
        toInt32(argMax(average_view_duration, inserted_at)) as average_view_duration,
        toInt32(argMax(average_view_percentage, inserted_at)) as average_view_percentage,
        if(count(*) != 0, round(engagement / toInt32(count(*)), 2), 0) as engagement_rate,
        max(published_at) as published_at_time
    FROM youtube_videos
    WHERE channel_id in ('...')
    AND [DateFilter on published_at]
    GROUP BY video_id
    ORDER BY {order_by} DESC  -- ASC for leastPostsQuery
    LIMIT {limit}  -- hardcoded 5 for leastPostsQuery
)
```

Uses `argMax(metric, inserted_at)` for deduplication. Permalink extracted from `iframe_embed_html` via `REPLACE(url, 'embed/', 'watch?v=')`.

---

### 8.10 Performance and Video Posting Schedule

**Engagement (`performanceEngagementAndVideoPostingScheduleQuery`):**

```sql
SELECT
    groupArray(count) as count, groupArray(likes) as likes,
    groupArray(dislikes) as dislikes, groupArray(shares) as shares,
    groupArray(comments) as comments, groupArray(engagement) as engagement,
    groupArray(published_at) as buckets
FROM
(
    SELECT
        toInt32(count()) as count, toInt32(sum(like)) as likes,
        toInt32(sum(dislike)) as dislikes, toInt32(sum(share)) as shares,
        toInt32(sum(comment)) as comments, toInt32(sum(engagement)) as engagement,
        toDate(published_at) as published_at
    FROM
    (
        SELECT
            last_value(likes) as like, last_value(dislikes) as dislike,
            last_value(shares) as share, last_value(comments) as comment,
            like + dislike + share + comment as engagement,
            published_at, channel_id
        FROM youtube_videos
        WHERE channel_id in ('...')
        AND [DateFilter on published_at]
        GROUP BY video_id, channel_id, published_at
    )
    GROUP BY published_at
    ORDER BY published_at ASC
    WITH FILL FROM toDate('ACCOUNT_CREATED_DATE') STEP 1
)
```

**Views (`performanceViewsAndVideoPostingScheduleQuery`):**

Uses CTE `videos` (post counts per day) LEFT JOINed with traffic insights aggregation. Same 13+1 non-subscriber view sources.

---

### 8.11 Engagement Rollup (`getEngagementRollup`)

Used by cross-platform overview page.

```sql
SELECT
    engagement_count, post_count,
    if(post_count!=0 AND {totalDays}!=0, (engagement_count/post_count)/{totalDays}, 0) as engagement_rate,
    if({totalDays}!=0, post_count/{totalDays}, 0) as post_rate,
    toInt32(reactions) as reactions, toInt32(comment) as comments, toInt32(share) as shares
FROM
(
    SELECT
        toInt32(sum(likes + dislikes + shares + comments)) as engagement_count,
        toInt32(sum(likes + dislikes)) as reactions,
        toInt32(sum(comments)) as comment, toInt32(sum(shares)) as share
    FROM
    (
        SELECT last_value(likes) as likes, last_value(dislikes) as dislikes,
            last_value(comments) as comments, last_value(shares) as shares
        FROM youtube_activity_insights
        WHERE channel_id in ('...')
        AND [DateFilter on created_at]
        AND [DateFilter on created_at]  -- NOTE: DateFilter appears TWICE (bug)
        GROUP BY record_id
    )
) AS engagement_stats
CROSS JOIN
(
    SELECT toInt32(count(*)) as post_count
    FROM (SELECT video_id FROM youtube_videos WHERE channel_id in ('...') AND [DateFilter on published_at] GROUP BY video_id)
) as post_count
```

Bug: `DateFilter('created_at')` is called twice.

---

## 9. TikTok Analytics

### 9.1 Constructor

```
timezone = payload['timezone'] ?? 'UTC'
date -> split on ' - ' -> startDate (00:00:00), endDate (23:59:59)
tiktok_id -> string or array, merged with optional tiktok_accounts
page_id = SQL IN-clause string e.g. ('id1','id2')
post_table = 'tiktok_posts'
insights_table = 'tiktok_insights'
```

### 9.2 Key Methods

| Method | Description |
|--------|-------------|
| `getPageAndPostsInsights` | Summary with current/previous comparison |
| `getPageFollowersAndViews` | Followers and views time-series |
| `getDynamicPageFollowersAndViews` | Dynamic daily/monthly version |
| `getPostsAndEngagements` | Post counts and engagement per day |
| `getDailyEngagementsData` | Cumulative engagement with daily deltas |
| `getDynamicDailyEngagementsData` | Dynamic daily/monthly version |
| `getTopAndLeastPerformingPosts` | UNION ALL of top-5 and bottom-5 |
| `getPostsData` | Paginated posts list |

### 9.3 `getPageAndPostsInsights()`

**Tables:** `tiktok_posts`, `tiktok_insights`

```sql
select
    tiktok_id as tiktok_id,
    if(posts_summary.display_name == '', '{platformName}', posts_summary.display_name) as page_name,
    '{platform_logo}' as logo,
    toInt32(posts_summary.total_likes) as total_likes,
    toInt32(posts_summary.total_comments) as total_comments,
    toInt32(posts_summary.total_shares) as total_shares,
    toInt32(posts_summary.total_engagements) as total_engagements,
    toInt32(posts_summary.total_posts) as total_posts,
    if(insights_summary.tiktok_id = '','N/A',toString(insights_summary.total_follower_count)) as total_follower_count,
    if(insights_summary.tiktok_id = '','N/A',toString(insights_summary.total_following_count)) as total_following_count,
    if(insights_summary.tiktok_id = '','N/A',toString(insights_summary.total_video_views)) as total_video_views
from
(
    select
        tiktok_id as tiktok_id,
        display_name as display_name,
        total_likes, total_comments, total_shares, total_engagements, total_posts
    from
    (
        SELECT c1 as tiktok_id
        FROM VALUES {page_id}
    ) as platform_id
    left join
    (
        select
            tiktok_id,
            last_value(display_name) as display_name,
            sum(like_count) as total_likes,
            sum(comments_count) as total_comments,
            sum(share_count) as total_shares,
            sum(engagement_count) as total_engagements,
            count() as total_posts
        from
        (
            select
                tiktok_id, post_id,
                last_value(display_name) as display_name,
                max(like_count) as like_count,
                max(comments_count) as comments_count,
                max(share_count) as share_count,
                max(engagement_count) as engagement_count
            from tiktok_posts
            where tiktok_id in {page_id}
              AND toDateTime(created_at, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
            group by tiktok_id, post_id
        )
        group by tiktok_id
    ) as post_data on platform_id.tiktok_id = post_data.tiktok_id
) as posts_summary
left join
(
    select
        tiktok_id, total_follower_count, total_following_count, total_video_views
    from tiktok_insights
    where tiktok_id in {page_id}
      AND toDateTime(inserted_at, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
    order by inserted_at desc
    limit 1
) as insights_summary on posts_summary.tiktok_id = insights_summary.tiktok_id
```

**Deduplication:** Inner subquery groups by `(tiktok_id, post_id)` taking `max()` of each metric. Insights uses `limit 1 order by inserted_at desc` for most recent row.

---

### 9.4 Followers and Views (`getPageFollowersAndViews`)

**Table:** `tiktok_insights`

```sql
select
    platform_id,
    max(display_name) as display_name,
    '{platform_logo}' as logo,
    groupArray(follower_count) as followers_count,
    groupArray(views_per_day) as views_per_day,
    groupArray(follower_count_diff) as followers_count_diff,
    groupArray(views_per_day_diff) as views_per_day_diff,
    groupArray(inserted_at) as day_bucket
from
(
    select
        max(tiktok_id) as platform_id,
        max(display_name) as display_name,
        max(total_follower_count) as follower_count,
        runningDifference(max(total_follower_count)) as follower_count_diff,
        if(runningDifference(max(total_video_views))<0, 0, runningDifference(max(total_video_views))) AS views_per_day_diff,
        max(total_video_views) AS views_per_day,
        toDateTime(inserted_at) as inserted_at
    from tiktok_insights
    where tiktok_id in {page_id}
      and toDateTime(inserted_at, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
    group by inserted_at
    order by inserted_at
    WITH FILL
        FROM toStartOfDay(toDate('{startDate}'))
        TO toStartOfDay(toDate('{endDate+1day}'))
        STEP INTERVAL 1 DAY
    INTERPOLATE (
        platform_id AS {page_id},
        display_name as '{platform_name}',
        follower_count as -1,
        views_per_day as -1
    )
)
group by platform_id
```

**Controller logic:** Subtracts 1 day from startDate before query, then shifts off first array element. Normalizes leading `-1` sentinels (trim + forward-fill).

---

### 9.5 Posts and Engagements (`getPostsAndEngagements`)

**Table:** `tiktok_posts`

```sql
select
    tiktok_id,
    '{platformName}' as page_name,
    '{platform_logo}' as logo,
    groupArray(posting_day) as days_bucket,
    groupArray(sum_view_count) as sum_view_count,
    groupArray(sum_like_count) as sum_like_count,
    groupArray(sum_comments_count) as sum_comments_count,
    groupArray(sum_share_count) as sum_share_count,
    groupArray(sum_engagement_count) as sum_engagement_count,
    groupArray(avg_engagement_rate) as avg_engagement_rate,
    groupArray(post_count) as post_count
from
(
    select
        tiktok_id,
        toDate(day_of_post) as posting_day,
        toInt32(sum(view_count)) as sum_view_count,
        toInt32(sum(like_count)) as sum_like_count,
        toInt32(sum(comments_count)) as sum_comments_count,
        toInt32(sum(share_count)) as sum_share_count,
        toInt32(sum(engagement_count)) as sum_engagement_count,
        toInt32(avg(round(engagement_rate, 2))) as avg_engagement_rate,
        toInt32(count()) as post_count
    from
    (
        select
            post_id, tiktok_id,
            max(view_count) as view_count,
            max(like_count) as like_count,
            max(comments_count) as comments_count,
            max(share_count) as share_count,
            max(engagement_count) as engagement_count,
            max(engagement_rate) as engagement_rate,
            toDate(max(created_at)) as day_of_post
        from tiktok_posts
        where tiktok_id in {page_id}
          AND toDateTime(created_at, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
        group by tiktok_id, post_id
    )
    group by tiktok_id, day_of_post
    order by toDate(day_of_post)
    WITH FILL
        FROM toDate('{startDate}')
        TO toDate('{endDate+1day}')
        STEP INTERVAL 1 DAY
    INTERPOLATE (tiktok_id AS {page_id})
)
group by tiktok_id
```

**Controller:** If all `post_count[]` elements are zero, empties all array columns.

---

### 9.6 Daily Engagements (`getDailyEngagementsData`)

**Table:** `tiktok_insights`

```sql
select
    platform_id as tiktok_id,
    '{platformName}' as page_name,
    '{platform_logo}' as logo,
    groupArray(total_video_likes) as total_video_likes,
    groupArray(total_video_comments) as total_video_comments,
    groupArray(total_video_shares) as total_video_shares,
    groupArray(daily_video_likes) as daily_video_likes,
    groupArray(daily_video_comments) as daily_video_comments,
    groupArray(daily_video_shares) as daily_video_shares,
    groupArray(total_engagement) as total_engagement,
    groupArray(daily_engagements) as daily_engagement,
    groupArray(day_of_metric) as days_bucket
from
(
    select
        MAX(record_id) as record_id,
        MAX(total_video_likes) as total_video_likes,
        MAX(total_video_comments) as total_video_comments,
        MAX(total_video_shares) as total_video_shares,
        toInt32(MAX(total_engagement)) as total_engagement,
        toInt32(if(MAX(daily_video_likes)<0, 0, MAX(daily_video_likes))) as daily_video_likes,
        toInt32(if(MAX(daily_video_comments)<0, 0, MAX(daily_video_comments))) as daily_video_comments,
        toInt32(if(MAX(daily_video_shares)<0, 0, MAX(daily_video_shares))) as daily_video_shares,
        toInt32(daily_video_likes + daily_video_comments + daily_video_shares) as daily_engagements,
        platform_id,
        toDateTime(metric_day) as day_of_metric
    from
    (
        select
            record_id,
            MAX(total_video_likes) as total_video_likes,
            MAX(total_video_comments) as total_video_comments,
            MAX(total_video_shares) as total_video_shares,
            total_video_likes + total_video_comments + total_video_shares as total_engagement,
            toInt32(runningDifference(total_video_likes)) as daily_video_likes,
            toInt32(runningDifference(total_video_comments)) as daily_video_comments,
            toInt32(runningDifference(total_video_shares)) as daily_video_shares,
            max(tiktok_id) as platform_id,
            toDate(max(inserted_at)) as metric_day
        from tiktok_insights
        where tiktok_id in {page_id}
          AND toDateTime(inserted_at, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
        group by record_id
    )
    group by platform_id, metric_day
    order by toDateTime(metric_day)
    WITH FILL
        FROM toStartOfDay(toDate('{startDate}'))
        TO toStartOfDay(toDate('{endDate+1day}'))
        STEP INTERVAL 1 DAY
    INTERPOLATE (
        platform_id AS {page_id},
        total_video_likes as -1,
        total_video_comments as -1,
        total_video_shares as -1
    )
)
group by platform_id
```

**Logic:** `runningDifference()` on cumulative totals produces daily deltas. Negative diffs clamped to 0. `-1` sentinel marks gap days.

---

### 9.7 Top and Least Performing Posts (`getTopAndLeastPerformingPosts`)

**Tables:** `tiktok_posts`, `tiktok_insights`

```sql
select * from
(
    select * from
    (
        (
            select
                'top_posts' as category,
                tiktok_id,
                '{platformName}' as page_name,
                '{platform_logo}' as logo,
                max(profile_link) as profile_link,
                post_id,
                max(cover_image_url) as cover_image_url,
                max(share_url) as share_url,
                max(post_description) as post_description,
                max(hashtags) as hashtags,
                max(duration) as duration,
                max(height) as height,
                max(width) as width,
                max(title) as title,
                max(embed_html) as embed_html,
                max(embed_link) as embed_link,
                max(like_count) as likes_count,
                max(comments_count) as comments_count,
                max(share_count) as shares_count,
                max(view_count) as views_count,
                max(engagement_count) as engagement_count,
                round(max(engagement_rate), 2) as engagement_rate,
                max(inserted_at) as inserted_at,
                max(created_at) as created_time
            from tiktok_posts
            WHERE tiktok_id in {page_id}
              AND toDateTime(created_at, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
            group by tiktok_id, post_id
            order by engagement_count desc
            limit 5
        )
        UNION ALL
        (
            select
                'least_posts' as category,
                -- same columns as above --
            from tiktok_posts
            WHERE tiktok_id in {page_id}
              AND toDateTime(created_at, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
            group by tiktok_id, post_id
            order by engagement_count asc
            limit 5
        )
    ) as post_data
    LEFT JOIN
    (
        select total_follower_count, tiktok_id
        from tiktok_insights
        WHERE tiktok_id in {page_id}
        order by inserted_at desc
        limit 1
    ) as insights_data on post_data.tiktok_id = insights_data.tiktok_id
)
```

**Note:** Insights join has NO date filter -- always uses latest `tiktok_insights` row.

---

### 9.8 Posts Data (`getPostsData($sortOrder, $limit, $offset)`)

**Tables:** `tiktok_posts`, `tiktok_insights`

```sql
select * from
(
    select * FROM
    (
        select
            tiktok_id,
            '{platformName}' as page_name,
            '{platform_logo}' as logo,
            max(profile_link) as profile_link,
            post_id,
            max(cover_image_url) as cover_image_url,
            max(share_url) as share_url,
            max(post_description) as post_description,
            max(hashtags) as hashtags,
            max(duration) as duration,
            max(height) as height,
            max(width) as width,
            max(title) as title,
            max(embed_html) as embed_html,
            max(embed_link) as embed_link,
            max(like_count) as likes_count,
            max(comments_count) as comments_count,
            max(share_count) as shares_count,
            max(view_count) as views_count,
            max(engagement_count) as engagements_count,
            engagements_count as total_engagement,
            round(max(engagement_rate), 2) as engagement_rate,
            max(inserted_at) as inserted_at,
            max(created_at) as created_time
        from tiktok_posts
        where tiktok_id in {page_id}
          and toDateTime(created_at, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
        group by tiktok_id, post_id
        order by {sortOrder} desc
        limit {limit}
        OFFSET {offset}
    ) as posts_data
    LEFT JOIN
    (
        SELECT tiktok_id, count() as total
        from
        (
            select tiktok_id, post_id, max(created_at) as created_time
            from tiktok_posts
            where tiktok_id in {page_id}
              and toDateTime(created_at, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
            group by tiktok_id, post_id
        )
        group by tiktok_id
    ) as post_count on post_count.tiktok_id = posts_data.tiktok_id
) as post_info
LEFT JOIN
(
    select total_follower_count, tiktok_id
    from tiktok_insights
    WHERE tiktok_id in {page_id}
    order by inserted_at desc
    limit 1
) as insights_data on post_info.tiktok_id = insights_data.tiktok_id
```

**Pagination:** `LIMIT {limit} OFFSET {offset}` on the posts subquery. Total count joined as `total`.

---

### 9.9 Engagement Rollup (`getEngagementRollup`)

**Table:** `tiktok_posts`

```sql
SELECT
    toInt32(sum(engagement_count)) as engagement_count,
    toInt32(sum(like_count)) as like_count,
    toInt32(sum(comments_count)) as comments_count,
    toInt32(sum(share_count)) as share_count,
    avg(engagement_rate) as engagement_rate,
    toInt32(count()) as post_count,
    count()/{numberOfDays} as post_rate
from
(
    SELECT
        last_value(like_count) as like_count,
        last_value(comments_count) as comments_count,
        last_value(share_count) as share_count,
        last_value(engagement_count) as engagement_count,
        last_value(engagement_rate) as engagement_rate
    FROM tiktok_posts
    where tiktok_id in {page_id}
      AND toDateTime(created_at, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
    GROUP BY post_id
)
```

### 9.10 Dynamic Aggregation

TikTok uses 180-day threshold:
- `dateDiff <= 180` -> daily (`INTERVAL 1 DAY`, `toStartOfDay`)
- `dateDiff > 180` -> monthly (`INTERVAL 1 MONTH`, `toStartOfMonth`)

Both `getDynamicPageFollowersAndViews()` and `getDynamicDailyEngagementsData()` return additional `aggregation_level` field.

---

## 10. Pinterest Analytics

### 10.1 Constructor

```
ids -> pinterest_id (the user_id)
board_id -> triggers board mode (user mode is default)
filter_by -> 'video' or 'image'; defaults to "'video', 'image'"
order_by -> column name for sorting pins
date -> defaults to last 30 days if not provided
timezone -> 'Europe/Kyiv' remapped to 'Europe/Riga'
```

### 10.2 User Mode vs Board Mode

Every Pinterest endpoint has two query variants:
- **User mode:** Filters by `user_id`, uses `pinterest_user_insights` and `pinterest_users`
- **Board mode:** Filters by `board_id`, uses `pinterest_pins`, `pinterest_pin_insights`, `pinterest_boards`

### 10.3 Date Handling

Uses `toDate()` not `toDateTime()` -- timezone-naive date-level comparison.

```sql
-- DateFilter:
toDate({date_field}) BETWEEN toDate('{currentStartDate}') AND toDate('{currentEndDate+1day}')

-- DateFill:
WITH FILL FROM toDate('{account_created_date}') TO toDate('{currentEndDate}') STEP 1
```# ContentStudio Analytics API Reference

This document is the complete reference for every analytics API endpoint, ClickHouse SQL query, and business logic rule implemented in the ContentStudio backend. It is intended to be sufficient for a Go developer to reimplement every endpoint by reading only this document.

---

## Table of Contents

1. [Common Patterns and Conventions](#1-common-patterns-and-conventions)
2. [ClickHouse Tables Reference](#2-clickhouse-tables-reference)
3. [Facebook Analytics](#3-facebook-analytics)
4. [Facebook Competitor Analytics](#4-facebook-competitor-analytics)
5. [Instagram Analytics](#5-instagram-analytics)
6. [Instagram Competitor Analytics](#6-instagram-competitor-analytics)
7. [LinkedIn Analytics](#7-linkedin-analytics)
8. [YouTube Analytics](#8-youtube-analytics)
9. [TikTok Analytics](#9-tiktok-analytics)
10. [Pinterest Analytics](#10-pinterest-analytics)
11. [Twitter/X Analytics](#11-twitterx-analytics)
12. [Cross-Platform Overview V2](#12-cross-platform-overview-v2)
13. [Campaign and Label Analytics](#13-campaign-and-label-analytics)
14. [Date Handling Patterns](#14-date-handling-patterns)
15. [Deduplication Strategies](#15-deduplication-strategies)
16. [Zero-Fill Patterns](#16-zero-fill-patterns)
17. [Dynamic Aggregation Patterns](#17-dynamic-aggregation-patterns)
18. [Known Quirks and Bugs](#18-known-quirks-and-bugs)
19. [Cross-Platform Overview V1](#19-cross-platform-overview-v1)
20. [Dashboard Analytics](#20-dashboard-analytics)
21. [Analytics Share Link Management](#21-analytics-share-link-management)
22. [Reports and Scheduled Reports](#22-reports-and-scheduled-reports)
23. [Analytics Job Triggers](#23-analytics-job-triggers)
24. [Twitter/X Settings Management](#24-twitterx-settings-management)
25. [Competitor Management CRUD](#25-competitor-management-crud)
26. [AI Insights](#26-ai-insights)

---

## 1. Common Patterns and Conventions

### 1.1 Request Parameters (Universal)

Every analytics endpoint accepts these core parameters:

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `workspace_id` | string | required | MongoDB workspace identifier |
| `date` | string | required | Format: `"YYYY-MM-DD - YYYY-MM-DD"` |
| `timezone` | string | `"UTC"` | IANA timezone string |
| `{platform}_id` | string/array | required | Platform account identifier(s) |
| `media_type` | array | platform defaults | Filter by content type |
| `limit` | int | varies (5-20) | Max results for top posts |
| `order_by` | string | varies | Sort column for top posts |

### 1.2 Previous Period Calculation

Every endpoint automatically computes a comparison period of equal length:

```
previous_start = start - (end - start)
previous_end = start
```

Example: current range `2024-02-01 - 2024-02-29` (28 days) produces previous `2024-01-04 - 2024-02-01`.

### 1.3 ClickHouse Execution Pattern

All queries run via `Clickhouse::getInstance('analytics')->getClient()` with `->onCluster(env('CLICKHOUSE_ANALYTICS_CLUSTER'))`.

### 1.4 Engagement Formulas by Platform

| Platform | Engagement Formula |
|----------|-------------------|
| **Facebook** | reactions + comments + shares |
| **Instagram** | likes + comments + saved |
| **LinkedIn** | comments + favorites + shares (NOT including clicks) |
| **YouTube** | likes + dislikes + comments + shares |
| **TikTok** | `engagement_count` column (pre-computed) |
| **Pinterest** | saves + pin_clicks + outbound_clicks |
| **Twitter/X** | `total_engagement` column (pre-computed) |

### 1.5 'N/A' Sentinel Value

When a metric has no data (count is zero), the string `'N/A'` is returned instead of null or 0. The controller's difference/percentage calculations must handle this explicitly by checking for `'N/A'` before arithmetic.

### 1.6 Growth/Percentage Formula

```
growth = round((current - previous) / max(previous, 1) * 100, 2)
```

Returns `"N/A"` when previous is 0 or either value is `"N/A"`.

---

## 2. ClickHouse Tables Reference

### Facebook Tables

| Table | Key Columns |
|-------|-------------|
| `facebook_posts` | `post_id`, `page_id`, `saving_time`, `created_time`, `media_type`, `day_of_week`, `hour_of_day`, `total` (reactions), `comments`, `shares`, `post_clicks`, `total_engagement`, `post_impressions`, `post_impressions_unique`, `post_impressions_paid`, `post_impressions_organic`, `post_impressions_viral`, `post_impressions_paid_unique`, `post_impressions_organic_unique`, `post_impressions_viral_unique`, `post_video_views`, `permalink`, `full_picture`, `caption`, `description`, `link`, `status_type`, `video_id`, `category`, `published_by`, `message_tags`, `post_metadata` |
| `facebook_insights` | `page_id`, `hash_id`, `saving_time`, `created_time`, `page_fans`, `page_follows`, `page_impressions`, `page_impressions_paid`, `page_impressions_organic`, `page_post_engagements`, `page_fans_by_like`, `page_fans_by_unlike`, `talking_about_count`, `positive_sentiment`, `negative_sentiment`, `page_positive_feedback`, `page_negative_feedback`, `page_fans_online`, `page_fans_gender`, `page_fans_age`, `page_fans_gender_age`, `page_fans_country`, `page_fans_city`, `day_of_week` |
| `facebook_media_assets` | `post_id`, `page_id`, `media_id`, `caption`, `link`, `assetType`, `callToAction`, `createdAt`, `inserted_at` |
| `facebook_video_insights` | `post_id`, `page_id`, `video_id`, `total_video_views`, `total_video_views_organic`, `total_video_views_paid`, `total_video_view_total_time`, `total_video_view_total_time_organic`, `total_video_view_total_time_paid` |
| `facebook_reels_insights` | `post_id`, `page_id`, `play_count`, `total_time_watched_in_ms`, `average_time_watched`, `impressions_unique`, `created_at` |
| `facebook_competitor_posts` | `facebook_id`, `post_id`, `like`, `haha`, `angry`, `sad`, `love`, `thankful`, `wow`, `total_post_reactions`, `comments`, `shares`, `post_engagement`, `media_type`, `status_type`, `permalink`, `caption`, `hashtags`, `day_of_week`, `hour_of_day`, `created_at`, `inserted_at` |
| `facebook_competitor_insights` | `facebook_id`, `inserted_at`, `followersCount`, `fanCount`, `page_name`, `page_category`, `biography`, `slug`, `image` |
| `facebook_competitor_media_assets` | `post_id`, `facebook_id`, `media_id`, `caption`, `link`, `asset_type`, `call_to_action`, `created_at` |

### Instagram Tables

| Table | Key Columns |
|-------|-------------|
| `instagram_posts` | `media_id`, `instagram_id`, `stored_event_at`, `post_created_at`, `media_type`, `entity_type`, `engagement`, `like_count`, `comments_count`, `saved`, `reach`, `impressions`, `views`, `shares`, `hashtags`, `reels_avg_watch_time`, `reels_total_watch_time`, `replies`, `exits`, `taps_forward`, `taps_back`, `permalink`, `caption`, `media_url` |
| `instagram_insights` | `instagram_id`, `record_id`, `stored_event_at`, `created_time`, `online_users_datetime`, `followers_count`, `follows_count`, `profile_views`, `engagement`, `impressions`, `reach`, `accounts_engaged`, `online_followers`, `day_of_week`, `audience_age`, `audience_gender`, `audience_city`, `audience_country` |

### LinkedIn Tables

| Table | Key Columns |
|-------|-------------|
| `linkedin_posts` | `post_id`, `linkedin_id`, `saving_time`, `published_at`, `created_at`, `media_type`, `day_of_week`, `favorites`, `comments`, `repost`, `post_clicks`, `total_engagement`, `impressions`, `reach`, `hashtags`, `title`, `image`, `article_url` |
| `linkedin_insights` | `linkedin_id`, `record_id`, `inserted_at`, `created_at`, `totalFollowerCount`, `organicFollowerCount`, `paidFollowerCount`, `page_views`, `desktop_page_views`, `mobile_page_views`, `impressionCount`, `engagement`, `reach`, `repost`, `comments`, `reactions`, `unique_visitors`, `followers_by_seniority`, `followers_by_industry`, `followers_by_country`, `followers_by_city` |

### YouTube Tables

| Table | Key Columns |
|-------|-------------|
| `youtube_activity_insights` | `record_id`, `channel_id`, `created_at`, `estimated_minutes_watched`, `average_view_duration`, `views`, `likes`, `dislikes`, `comments`, `shares` |
| `youtube_channels` | `record_id`, `channel_id`, `subscriber_count`, `inserted_at`, `created_at`, `title` |
| `youtube_videos` | `video_id`, `channel_id`, `published_at`, `inserted_at`, `title`, `description`, `duration`, `thumbnail_url`, `media_type`, `iframe_embed_html`, `likes`, `dislikes`, `views`, `red_views`, `favorites`, `comments`, `subscribers_gained`, `shares`, `minutes_watched`, `red_minutes_watched`, `average_view_duration`, `average_view_percentage` |
| `youtube_traffic_insights` | `record_id`, `channel_id`, `created_at`, `subscriber_views`, `subscriber_watch_time`, `non_subsciber_watch_time` (typo), `paid_views`, `annotation_views`, `end_screen_views`, `campaign_card_view`, `no_link_other_views`, `yt_channel_views`, `yt_search_views`, `related_video_views`, `yt_other_page_views`, `ext_url_views`, `playlist_views`, `notification_views`, `shorts_views` |
| `youtube_shared_insights` | `channel_id`, `inserted_at`, and 31 sharing platform columns (ameba, blogger, copy_paste, cyworld, digg, dropbox, embed, mail, whats_app, other, facebook_messenger, facebook_pages, facebook, fotka, vkontakte, google_plus, discord, linkedin, goo, hangouts, pinterest, myspace, reddit, skype, telegram, tumblr, twitter, viber, weibo, wechat, youtube) |

### TikTok Tables

| Table | Key Columns |
|-------|-------------|
| `tiktok_posts` | `tiktok_id`, `post_id`, `display_name`, `like_count`, `comments_count`, `share_count`, `engagement_count`, `engagement_rate`, `view_count`, `created_at`, `inserted_at`, `profile_link`, `cover_image_url`, `share_url`, `post_description`, `hashtags`, `duration`, `height`, `width`, `title`, `embed_html`, `embed_link` |
| `tiktok_insights` | `tiktok_id`, `record_id`, `display_name`, `total_follower_count`, `total_following_count`, `total_video_views`, `total_video_likes`, `total_video_comments`, `total_video_shares`, `inserted_at` |

### Pinterest Tables

| Table | Key Columns |
|-------|-------------|
| `pinterest_pins` | `pin_id`, `board_id`, `user_id`, `created_at`, `media_type`, `is_owner`, `title`, `description`, `board_owner`, `cover_image_url`, `dominant_color`, `creative_type`, `product_tags`, `height`, `width` |
| `pinterest_pin_insights` | `pin_id`, `record_id`, `user_id`, `created_at`, `saving_time`, `inserted_at`, `impression`, `pin_clicks`, `outbound_click`, `saves`, `engagement`, `quartile_95s_percent_view`, `closeup`, `video_start`, `video_10s_view`, `video_avg_watch_time` |
| `pinterest_user_insights` | `user_id`, `created_at`, `impression`, `pin_clicks`, `outbound_click`, `saves`, `engagement` |
| `pinterest_users` | `user_id`, `inserted_at`, `follower_count` |
| `pinterest_boards` | `board_id`, `user_id`, `inserted_at`, `follower_count`, `name` |

### Twitter/X Tables

| Table | Key Columns |
|-------|-------------|
| `twitter_posts` | `twitter_id`, `post_id`, `created_at`, `inserted_at`, `tweet_type`, `total_engagement`, `impressions`, `like_count`, `reply_count`, `retweet_count`, `quote_count`, `bookmark_count`, `url_link_clicks`, `user_profile_clicks`, `impression_count`, `hashtags`, `permalink`, `full_text` |
| `twitter_insights` | `twitter_id`, `record_id`, `inserted_at`, `followers_count`, `following_count`, `tweet_count`, `listed_count` |

### Cross-Platform Views

| Table | Key Columns |
|-------|-------------|
| `mv_social_daily_metrics` | Materialized view. `date`, `platform`, `account_id`, `posts_count` (AggregateFunction(uniq)), `engagement_sum` (AggregateFunction(sum)), `impressions_sum` (AggregateFunction(sum)), `reach_sum` (AggregateFunction(sum)). Read with `uniqMerge(posts_count)`, `sumMerge(engagement_sum)`, etc. |

---

## 3. Facebook Analytics

### 3.1 Request Parameters

| Parameter | Type | Default | Notes |
|-----------|------|---------|-------|
| `workspace_id` | string | required | |
| `date` | string | required | `"YYYY-MM-DD - YYYY-MM-DD"` |
| `facebook_id` | string/array | required | Single ID or array |
| `timezone` | string | required | IANA timezone |
| `media_type` | array | `['text','link','images','videos','carousel','share','reels','others']` | |
| `limit` | int | 15 | Top posts limit |
| `order_by` | string | `'total_engagement'` | Sort column |

### 3.2 Date Filter Helper

Two modes:

```sql
-- posts mode (insights=false): converts column to user timezone
toDateTime({date_column}, 0, '{timezone}') BETWEEN toDateTime('{startDate}',0) AND toDateTime('{endDate+1day}',0)

-- insights mode (insights=true): plain date comparison
toDate({date_column}) BETWEEN toDate('{startDate}') AND toDate('{endDate}')
```

**CRITICAL side-effect:** After building the filter, `currentEndDate` is mutated via `subDay()`. This means `getDateFilters` must only be called once per query or the date window shrinks.

### 3.3 Endpoints and SQL Queries

#### 3.3.1 Overview Summary (`getSummaryQuery`)

**Route:** `POST overview/facebook/summary`

**Tables:** `facebook_posts`, `facebook_insights`

**Response keys:** `overview.current`, `overview.previous`

```sql
WITH posts as (
    SELECT post_id, max(saving_time)
    FROM facebook_posts
    WHERE page_id in {facebookId}
      AND toDateTime(created_time, 0, '{timezone}') BETWEEN toDateTime('{startDate}',0) AND toDateTime('{endDate+1day}',0)
    group by post_id
)
SELECT *
from (
    SELECT *
    from(
        SELECT c1 as page_id
        FROM VALUES {facebookId}
    ) as page_ids
    LEFT JOIN (
      SELECT toInt32(count()) as doc_count,
        toUInt64(page_id) as page_id,
        toInt32(reactions + comments + repost + posts_clicks) as total_engagement,
        toInt32(sum(total)) as reactions,
        toInt32(sum(comments)) as comments,
        toInt32(sum(post_clicks)) as posts_clicks,
        toInt32(sum(post_impressions)) as impressions,
        toInt32(sum(post_impressions_unique)) as reach,
        toInt32(sum(shares)) as repost
      FROM facebook_posts
      WHERE (post_id, saving_time) IN (posts)
      group by page_id
    ) AS posts_summary using page_id
) as posts_data
LEFT JOIN (
  SELECT toUInt64(page_id) as page_id,
    toInt32(sum(positive_sentiment)) as positive_sentiment,
    toInt32(sum(negative_sentiment)) as negative_sentiment,
    toInt32(sum(page_impressions)) as page_impressions,
    toInt32(sum(page_impressions_paid)) as page_impressions_paid,
    toInt32(sum(page_impressions_organic)) as page_impressions_organic,
    toInt32(sum(page_engagements)) as page_engagements,
    toInt32(sum(page_positive_feedback)) as page_positive_feedback,
    toInt32(sum(page_negative_feedback)) as page_negative_feedback,
    toInt32(max(fan_count)) as fan_count,
    toInt32(sum(talking_about_count)) as talking_about_count,
    toInt32(max(page_follows)) as page_follows
  FROM (
      SELECT page_id,
        toDate(created_time) as created_date,
        argMin(page_fans, saving_time) as fan_count,
        max(page_impressions) as page_impressions,
        max(page_impressions_paid) as page_impressions_paid,
        max(page_impressions_organic) as page_impressions_organic,
        max(page_post_engagements) as page_engagements,
        max(positive_sentiment) as positive_sentiment,
        max(negative_sentiment) as negative_sentiment,
        max(page_positive_feedback) as page_positive_feedback,
        max(page_negative_feedback) as page_negative_feedback,
        max(talking_about_count) as talking_about_count,
        argMin(page_follows, saving_time) as page_follows
      FROM facebook_insights
      WHERE page_id in {facebookId}
            AND toDate(created_time) BETWEEN toDate('{startDate}') AND toDate('{endDate}')
      GROUP BY page_id, created_date
  )
  group by page_id
) AS insights_summary USING page_id
```

**Output columns:** `doc_count`, `total_engagement`, `reactions`, `comments`, `posts_clicks`, `impressions`, `reach`, `repost`, `positive_sentiment`, `negative_sentiment`, `page_impressions`, `page_impressions_paid`, `page_impressions_organic`, `page_engagements`, `fan_count`, `talking_about_count`, `page_follows`

---

#### 3.3.2 Last Follower Counts (Fallback) (`getLastFollowerCounts`)

**Table:** `facebook_insights`

Used when fan_count time-series leads with zeros. Looks back 2 years.

```sql
SELECT
    arrayFirst(x -> x != 0, groupArray(fans)) AS page_fans,
    arrayFirst(x -> x != 0, groupArray(page_fans_by_like)) AS page_fans_by_like,
    arrayFirst(x -> x != 0, groupArray(page_fans_by_unlike)) AS page_fans_by_unlike
FROM (
    SELECT
        last_value(created_time) as inserted_time,
        toInt32(last_value(page_fans)) as fans,
        toInt32(last_value(page_fans_by_like)) as page_fans_by_like,
        toInt32(last_value(page_fans_by_unlike)) as page_fans_by_unlike
    FROM facebook_insights
    WHERE page_id in {facebookId}
        AND toDate(created_time) BETWEEN toDate('{startDate-2years}') AND toDate('{startDate}')
        AND page_fans != 0
    GROUP BY hash_id
    ORDER BY inserted_time DESC
)
```

---

#### 3.3.3 Audience Growth Time-Series (`getOverviewAudienceGrowthQuery`)

**Route:** `POST overview/facebook/audienceGrowth`

**Table:** `facebook_insights`

**Response keys:** `audience_growth` (time-series), `audience_growth_rollup.current`, `audience_growth_rollup.previous`

```sql
SELECT notEmpty(fan_count_temp) as show_data,
    arrayFill(x -> not x==0, fan_count_temp) as fan_count,
    page_fans_daily,
    page_fans_by_like_temp as page_fans_by_like,
    page_fans_by_unlike_temp as page_fans_by_unlike,
    page_impressions as page_impressions,
    page_engagements as page_engagements,
    date as buckets
FROM (
    SELECT groupArray(page_fans_total) as fan_count_temp,
            groupArray(page_fans_daily) as page_fans_daily,
            groupArray(page_fans_by_like) as page_fans_by_like_temp,
            groupArray(page_fans_by_unlike) as page_fans_by_unlike_temp,
            groupArray(page_impressions) as page_impressions,
            groupArray(page_engagements) as page_engagements,
            groupArray(created_date) as date
    from(
        SELECT toInt32(page_fans_total) as page_fans_total,
                toInt32(if(page_fans_total > 0, page_fans_total - lagInFrame(page_fans_total, 1, page_fans_total) OVER (ORDER BY created_date ASC), 0)) as page_fans_daily,
                toInt32(page_fans_by_like) as page_fans_by_like,
                toInt32(page_fans_by_unlike) as page_fans_by_unlike,
                toInt32(page_impressions) as page_impressions,
                toInt32(page_engagements) as page_engagements,
                created_date
        FROM (
            SELECT argMin(page_follows, saving_time) as page_fans_total,
                    argMin(page_fans_by_like, saving_time) as page_fans_by_like,
                    argMin(page_fans_by_unlike, saving_time) as page_fans_by_unlike,
                    max(page_impressions) as page_impressions,
                    max(page_post_engagements) as page_engagements,
                    toDate(created_time) as created_date
            FROM facebook_insights
            WHERE page_id in {facebookId} AND {dateFilter}
            GROUP BY created_date
            ORDER BY created_date ASC
            WITH FILL FROM toDate('{startDate}') TO toDate('{endDate}') + 1 STEP 1
        )
    )
)
```

**Key logic:**
- `argMin(page_follows, saving_time)` takes earliest daily record for fan count
- `lagInFrame(...) OVER (ORDER BY created_date)` computes daily net change
- `arrayFill(x -> not x==0, ...)` forward-fills zeros in the fan_count array
- `WITH FILL` zero-fills missing dates

**Dynamic version:** Switches daily/monthly at 60-day threshold. Returns additional `aggregation_level` field.

**Rollup query:** Returns `avg_page_fans_by_like`, `avg_page_fans_by_unlike`, `fan_count`, `talking_about_count`, `doc_count`, `page_id`.

---

#### 3.3.4 Publishing Behaviour (`getOverviewPublishingBehaviourByMediaTypeQuery`)

**Route:** `POST overview/facebook/publishingBehaviour`

**Table:** `facebook_posts`

```sql
WITH posts as (
    SELECT post_id, max(saving_time) as saving_time
    FROM facebook_posts
    WHERE page_id in {facebookId}
        AND media_type in ({media_types})
        AND {dateFilter_created_time}
    GROUP BY post_id
)
SELECT
    groupArray(reactions) as reactions_engagement,
    groupArray(comments) as comments_engagement,
    groupArray(shares) as shares_engagement,
    groupArray(paid_impressions) as paid_impressions,
    groupArray(organic_impressions) as organic_impressions,
    groupArray(viral_impressions) as viral_impressions,
    groupArray(paid_reach) as paid_reach,
    groupArray(organic_reach) as organic_reach,
    groupArray(viral_reach) as viral_reach,
    groupArray(created_date) as buckets,
    groupArray(post_count) as post_count
FROM (
    SELECT
        count() as post_count,
        sum(total) as reactions,
        sum(comments) as comments,
        sum(shares) as shares,
        sum(post_impressions_paid) as paid_impressions,
        sum(post_impressions_organic) as organic_impressions,
        sum(post_impressions_viral) as viral_impressions,
        sum(post_impressions_paid_unique) as paid_reach,
        sum(post_impressions_organic_unique) as organic_reach,
        sum(post_impressions_viral_unique) as viral_reach,
        toDate(created_time) as created_date
    FROM facebook_posts
    WHERE (post_id, saving_time) IN (posts)
    GROUP BY created_date
    ORDER BY created_date ASC
    WITH FILL TO toDate('{endDate+1day}')
)
```

---

#### 3.3.5 Top Posts (`getTop15PostsQuery`)

**Route:** `POST overview/facebook/getTopPosts`

**Tables:** `facebook_posts`, `facebook_media_assets`

```sql
WITH posts as (
    SELECT post_id, max(saving_time)
    FROM facebook_posts
    WHERE page_id in {facebookId}
    AND media_type in ({media_types})
        AND {dateFilter_created_time}
    group by post_id
)
SELECT *
FROM (
    SELECT *, 'top_posts' as post_category
    FROM (
        SELECT *
        FROM facebook_posts
        WHERE (post_id, saving_time) IN (posts)
        ORDER BY {order_by} desc, created_time
        LIMIT {limit} BY page_id
    ) as post_data
    LEFT JOIN (
        SELECT *
        FROM facebook_media_assets
        WHERE page_id in {facebookId}
            AND {dateFilter_created_at}
        order by inserted_at
    ) as media_assets using post_id
)
```

**Key:** `LIMIT {limit} BY page_id` returns top N posts per page for multi-account queries. Multiple rows per post_id from the media_assets JOIN are collapsed in the controller.

---

#### 3.3.6 Active Users by Hour (`getOverviewActiveUsersQuery`)

**Route:** `POST overview/facebook/activeUsers`

**Table:** `facebook_insights` (column: `page_fans_online`)

```sql
SELECT max(buckets) AS buckets,
    max(value) as values,
    arrayMax(values) as highest_value,
    buckets[indexOf(values, highest_value)] as highest_hour
FROM (
    SELECT max(buckets) AS buckets,
            arrayMap((x)->toInt32(x)/count(), sumForEach(values)) AS value
        FROM (
            SELECT arrayMap((x)->x[1], active_users_per_hour) AS buckets,
                    arrayMap((x)->toInt32(x[2]), active_users_per_hour) AS values
                FROM (
                    SELECT arrayMap((x)->(splitByChar('$', x)), arr) AS active_users_per_hour
                    FROM (
                        SELECT max(page_fans_online) AS arr
                        FROM facebook_insights
                        WHERE page_id in {facebookId}
                        AND toDate(created_time) BETWEEN toDate('{startDate}') AND toDate('{endDate}')
                        GROUP BY hash_id
                    )
                ))
)
```

**Data format:** `page_fans_online` is stored as a `$`-delimited packed string: `"0$123$1$456$..."` where pairs are `[hour, value]`.

**Controller timezone adjustment:**
```
timezoneInterval = round(Carbon::now(timezone)->offsetHours) + 8
```
The +8 accounts for data being stored in UTC-8. Buckets wrap around 0-23 after shifting.

---

#### 3.3.7 Active Users by Day (`getOverviewActiveUsersPerDayQuery`)

**Table:** `facebook_insights`

```sql
SELECT groupArray(day_name) as buckets,
        groupArray(active_users) as values,
        max(active_users) as highest_value,
        buckets[indexOf(values, highest_value)] as highest_day
FROM (
    SELECT ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'][day_num] as day_name,
            ifNull(active_users, 0) as active_users
    FROM (
        SELECT day_num, toInt32(count()) as active_users
        FROM (
            SELECT
                indexOf(['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'], last_value(day_of_week)) as day_num,
                last_value(created_time) as inserted_at
            FROM facebook_insights
            WHERE page_id in {facebookId}
            AND toDate(created_time) BETWEEN toDate('{startDate}') AND toDate('{endDate}')
            GROUP BY hash_id
        )
        GROUP BY day_num
        ORDER BY day_num ASC
        WITH FILL FROM 1 TO 8 STEP 1
    )
)
```

---

#### 3.3.8 Page Impressions (`getOverviewImpressionsQuery`)

**Route:** `POST overview/facebook/pageImpressions`

**Table:** `facebook_insights`

```sql
SELECT
    groupArray(page_impressions) as page_impressions,
    groupArray(created_date) as buckets
FROM (
    SELECT toInt32(max(page_impressions)) as page_impressions,
        toDate(created_time) as created_date
    FROM facebook_insights
    WHERE page_id = '{facebookId}'
    AND {dateFilter}
    GROUP BY created_date
    ORDER BY created_date ASC
    WITH FILL FROM toDate('{startDate}') TO toDate('{endDate}') + 1 STEP 1
)
```

**Rollup:** `total_impressions`, `avg_impressions_per_day` (sum/totalDays), `avg_impressions_per_week` (sum/totalWeeks).

---

#### 3.3.9 Page Engagements (`getOverviewEngagementsQuery`)

**Route:** `POST overview/facebook/engagement`

**Table:** `facebook_insights`

```sql
SELECT
    groupArray(page_engagements) as page_engagements,
    groupArray(created_date) as buckets
FROM (
    SELECT toInt32(max(page_post_engagements)) as page_engagements,
        toDate(created_time) as created_date
    FROM facebook_insights
    WHERE page_id = '{facebookId}'
    AND {dateFilter}
    GROUP BY created_date
    ORDER BY created_date ASC
    WITH FILL FROM toDate('{startDate}') TO toDate('{endDate}') + 1 STEP 1
)
```

---

#### 3.3.10 Reels Analytics (`getOverviewReelsAnalyticsQuery`)

**Route:** `POST overview/facebook/reelsAnalytics`

**Tables:** `facebook_posts`, `facebook_reels_insights`

```sql
WITH facebook_post_data AS (
    SELECT post_id,
         last_value(total) as reactions,
         last_value(comments) as comments,
         last_value(shares) as repost,
         reactions+comments+repost as total_engagement,
         toDate(created_time) as created_at
    FROM facebook_posts
    WHERE page_id in ({facebookId}) AND {dateFilter_created_at}
    GROUP BY post_id, created_at
)
SELECT
    groupArray(created_at) as buckets,
    groupArray(total_reels_count) as total_reels,
    groupArray(total_seconds_watched) as total_seconds_watched,
    groupArray(initial_plays) as initial_plays,
    groupArray(total_engagement) as engagement,
    groupArray(reactions) as reactions,
    groupArray(comments) as comments,
    groupArray(shares) as shares,
    toInt32(sum(total_reels_count)) as show_data
FROM (
    SELECT created_at,
        toInt32(count()) as total_reels_count,
        round(sum(total_time_watched_in_ms) / 1000, 2) as total_seconds_watched,
        toInt32(sum(play_count)) as initial_plays,
        toInt32(sum(total_engagement)) as total_engagement,
        toInt32(sum(reactions)) as reactions,
        toInt32(sum(comments)) as comments,
        toInt32(sum(repost)) as shares
    FROM (
        SELECT post_id, toDate(created_at) as created_at,
            last_value(average_time_watched) as average_time_watched,
            last_value(total_time_watched_in_ms) as total_time_watched_in_ms,
            last_value(play_count) as play_count,
            last_value(impressions_unique) as reach
        FROM facebook_reels_insights
        WHERE page_id in ({facebookId}) AND {dateFilter_created_at}
        GROUP BY post_id, created_at
    ) AS reels
    LEFT JOIN facebook_post_data ON facebook_post_data.post_id = reels.post_id
    GROUP BY reels.created_at
    ORDER BY created_at ASC
    WITH FILL TO (toDate('{endDate+1day}'))
)
```

**Note:** `total_time_watched_in_ms` is divided by 1000 for seconds.

---

#### 3.3.11 Video Insights (`getVideoInsightsQuery`)

**Route:** `POST overview/facebook/videoInsights`

**Tables:** `facebook_posts`, `facebook_video_insights`

Joins `facebook_posts` (filtered to `media_type='videos'`) with `facebook_video_insights` on `post_id`. View times are in ms, divided by 1000 for seconds.

---

#### 3.3.12 Time Recommendation (`getTimeRecommendationQuery`)

**Table:** `facebook_posts`

```sql
SELECT
    max(page_id) as facebookId,
    day_of_week,
    hour_of_day,
    sum(post_impressions) as post_impressions,
    sum(total_engagement) as total_engagement
FROM facebook_posts
WHERE page_id in {facebookId} AND {dateFilter}
GROUP BY day_of_week, hour_of_day
ORDER BY day_of_week, hour_of_day
```

---

#### 3.3.13 Demographics

**Audience Gender:** Reads `page_fans_gender` blob from `facebook_insights`. Parsed from `$`-delimited pairs. Known quirk: M/U/F labels mapped to wrong aliases.

**Audience Age:** Reads `page_fans_age` from most recent record. Age buckets: `65+`, `55-64`, `45-54`, `35-44`, `25-34`, `18-34`, `13-17`. Date boundary: if range crosses 2024-03-14, end date is clamped.

**Audience Country/City:** Reads `page_fans_country`/`page_fans_city` from most recent record. `$`-delimited parsing.

**Max Gender+Age:** Reads `page_fans_gender_age` blob. Finds the single bucket with highest value.

---

## 4. Facebook Competitor Analytics

### 4.1 Constructor and Timezone Handling

**CRITICAL difference from own analytics:** Competitor dates are parsed in user timezone, converted to UTC before embedding in SQL. Database timestamps are stored in UTC.

```
startDate = Carbon::parse(date[0], timezone)->setTimezone('UTC')
endDate = Carbon::parse(date[1], timezone)->setTime(23,59,59)->setTimezone('UTC')
```

### 4.2 Tables

`facebook_competitor_posts`, `facebook_competitor_insights`, `facebook_competitor_media_assets`

### 4.3 Helper Methods

#### `getPageFilters($insightsFlag, $filterAdded)`
Builds `facebook_id IN ('id1','id2',...)` for posts or `page_id IN (...)` for insights. If `$filterAdded = true`, excludes accounts with `state == "Added"`.

#### `getDateFilters($insightsFlag, $dateColumn)`
```sql
-- posts: created_at BETWEEN toDateTime('{startUTC}',0) AND toDateTime('{endUTC}',0)
-- insights: inserted_at BETWEEN toDateTime('{startUTC}',0) AND toDateTime('{endUTC}',0)
```

#### `getConstantConditions($field, $idField)`
Builds `multiIf(id='x','val', id='y','val', 'Random')` to map competitor IDs to metadata (name, image, state, slug).

### 4.4 Controller Methods

| Method | Description | Sort Default |
|--------|-------------|--------------|
| `searchCompetitor` | Graph API page search with appsecret_proof | N/A |
| `dataTableMetrics` | Current/previous period KPIs with growth % | `followersCount` |
| `postingActivityGraphByTypes` | Aggregate by media_type | `avgTotalEngagements` |
| `postingActivityBySpecificType` | Per-competitor for one media_type | N/A |
| `topAndLeastPerformingPosts` | Top-5/bottom-5 per competitor with media assets | engagement DESC/ASC |
| `topHashtags` | Top hashtags by engagement (limit default: 7) | count DESC |
| `individualHashtagData` | Per-competitor stats for one hashtag | engagement DESC |
| `biographyData` | Biography + length per competitor | `biography_length` |
| `followersGrowthComparison` | Per-competitor followers time-series | `followers_count` |
| `postReactDistribution` | Engagement totals for single page | N/A |
| `postReactDistributionByCompany` | Per-reaction breakdown for single page | N/A |
| `postTypeDistribution` | Per-media-type breakdown per competitor | N/A |
| `postEngagementOverTime` | Daily time-series per competitor (single page) | N/A |
| `postEngagementByCompetitor` | Total engagement per competitor | engagement DESC |

### 4.5 SQL Queries

#### 4.5.1 `getDataTableDataMetricsQuery($sortOrder)` -- Main Data Table

**Tables:** `facebook_competitor_posts`, `facebook_competitor_insights`

```sql
WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM facebook_competitor_posts
    WHERE {pageFilters} AND {dateFilters}
    group by post_id
)
SELECT
    multiIf(facebook_id='x','imgUrl',...) as image,
    multiIf(facebook_id='x','name',...) as name,
    multiIf(facebook_id='x','state',...) as state,
    * FROM (
    SELECT pages_ids.facebook_id as facebook_id,
        week_metrics.averageEngagement,
        week_metrics.averagePostsPerWeek,
        week_metrics.engagementRate,
        days_metrics.dayOfWeek,
        days_metrics.hourOfDay,
        days_metrics.averagePostsPerDay,
        days_metrics.averagePostsPerDayEngagement
    FROM (SELECT CAST(c1 AS String) as facebook_id FROM VALUES {accountIds}) as pages_ids
    LEFT JOIN (
        SELECT facebook_id,
            round(avg(total_engagement), 2) as averageEngagement,
            if(dateDiff('week', ...) != 0,
               round(sum(posts_in_a_week)/dateDiff('week',...), 2), 0) as averagePostsPerWeek,
            round((sum(total_engagement)/sum(posts_in_a_week)) / max(followers_count) * 100, 2) as engagementRate
        FROM (
            SELECT CAST(facebook_id AS String) AS facebook_id,
                sum(post_engagement) as total_engagement,
                toStartOfWeek(created_at) as week,
                count() as posts_in_a_week,
                followers_count
            FROM facebook_competitor_posts
            WHERE (post_id, inserted_at) IN (posts)
            group by facebook_id, week, followers_count
            order by week ASC
            WITH FILL FROM toStartOfWeek(toDate('{startDate}'))
                       TO toStartOfWeek(toDate('{endDate}'))
                       STEP INTERVAL 1 WEEK
        )
        group by facebook_id
    ) as week_metrics ON ...
    LEFT JOIN (
        SELECT facebook_id,
            argMax(multiIf(day_of_week=1,'Monday',...), total_posts) as dayOfWeek,
            argMax(multiIf(hour_of_day=0,'12:00 AM',...), total_posts) as hourOfDay,
            if(dateDiff('day',...) != 0,
               round(sum(total_posts)/dateDiff('day',...), 2), 0) as averagePostsPerDay,
            avg(total_engagement) as averagePostsPerDayEngagement,
            argMax(hour_of_day, total_posts) as maximumPostsHour,
            argMax(hour_of_day, total_engagement) as mostEngagementHour
        from (
            select facebook_id, ...,
                sum(post_engagement) as total_engagement,
                toDayOfWeek(created_at) as day_of_week,
                toHour(created_at) as hour_of_day,
                uniq(post_id) as total_posts,
                toDate(created_at) as date_c
            from facebook_competitor_posts
            WHERE (post_id, inserted_at) IN (posts)
            group by facebook_id, day_of_week, hour_of_day, date_c
            order by date_c ASC
            WITH FILL FROM toDate('{startDate}') TO toDate('{endDate}') STEP 1
        )
        group by facebook_id
    ) as days_metrics ON ...
) as page_metrics
LEFT JOIN (
    SELECT max(followers_count) as followersCount,
           max(total_fan_count) as fanCount,
           page_id
    FROM facebook_competitor_insights
    WHERE {pageFilters_insights} AND {dateFilters_insights}
    group by page_id
) as page_metadata ON page_metrics.facebook_id = page_metadata.page_id
ORDER BY {sortOrder} DESC
```

**Returns per competitor:** `facebook_id`, `image`, `name`, `state`, `averageEngagement`, `averagePostsPerWeek`, `engagementRate`, `dayOfWeek` (best day), `hourOfDay` (best hour), `averagePostsPerDay`, `averagePostsPerDayEngagement`, `followersCount`, `fanCount`.

---

#### 4.5.2 `getPostingActivityGraphByTypesQuery($sortOrder)` -- Activity by Media Type (Aggregated)

**Table:** `facebook_competitor_posts`

```sql
WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM facebook_competitor_posts
    WHERE {pageFilters} AND {dateFilters}
    group by post_id
)
SELECT
    media_type as mediaType,
    round(sum(avg_engagement),2) as avgTotalEngagements,
    sum(page_post_count) as totalPosts,
    toInt32(sum(total_engagement)) as total_engagement,
    round(sum(er),2) as avgEngagementRate,
    if(dateDiff('week',...) = 0, 0, round(totalPosts/dateDiff('week',...),2)) as postsPerWeek,
    if(dateDiff('day',...) = 0, 0, round(totalPosts/dateDiff('day',...),2)) as postsPerDay,
    if(dateDiff('hour',...) = 0, 0, round(totalPosts/dateDiff('hour',...),2)) as postsPerHour,
    dateDiff('week',...) as weekCount,
    dateDiff('day',...) as dayCount,
    dateDiff('hour',...) as hourCount
FROM (
    SELECT facebook_id, page_name, media_type,
        sum(count) as page_post_count,
        sum(total_post_engagement) as total_engagement,
        if(page_post_count<=0,0, round(sum(total_post_engagement)/page_post_count,2)) as avg_engagement,
        if(page_post_count<=0,0, round(((sum(er)/page_post_count)/max(followers_count))*100,2)) as er
    FROM (
        SELECT facebook_id, page_name, media_type,
            count() as count,
            sum(post_engagement) as total_post_engagement,
            total_post_engagement as er,
            argMax(followers_count, created_at) as followers_count,
            toStartOfWeek(created_at) as week
        from facebook_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        group by facebook_id, page_name, media_type, week
        ORDER by week asc
        WITH FILL FROM toStartOfWeek(toDate('{startDate}'))
                   TO toStartOfWeek(toDate('{endDate}'))
                   STEP INTERVAL 1 WEEK
    )
    WHERE media_type!=''
    GROUP BY facebook_id, page_name, media_type
)
group by media_type
```

Aggregates **all competitors together** by `media_type`.

---

#### 4.5.3 `getPostingActivityBySpecificTypeQuery($mediaType, $sortOrder)`

Same structure as above but WHERE clause adds `media_type = '$mediaType'` and groups per `facebook_id`. Adds per-competitor metadata from `facebook_competitor_insights`. Returns one row per competitor for that media type.

---

#### 4.5.4 `getTopPerformingAndLeastPerformingPostsQuery()` -- Top+Least Posts

**Tables:** `facebook_competitor_posts`, `facebook_competitor_media_assets`

```sql
WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM facebook_competitor_posts
    WHERE {pageFilters} AND {dateFilters}
    group by post_id
)
SELECT * FROM (
    SELECT *, 'top_5_posts' as category,
        'https://graph.facebook.com/' || facebook_id || '/picture?type=large' as image
    FROM (
        SELECT * FROM facebook_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        ORDER BY post_engagement desc
        LIMIT 5 BY facebook_id
    ) as post_data
    LEFT JOIN (
        SELECT * FROM facebook_competitor_media_assets
        WHERE {pageFilters_insights} AND created_at BETWEEN ...
        order by inserted_at desc
    ) as media_assets using post_id
UNION ALL
    SELECT *, 'least_5_posts' as category, ...
    FROM (
        SELECT * FROM facebook_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        ORDER BY post_engagement asc
        LIMIT 5 BY facebook_id
    ) as post_data
    LEFT JOIN (...) as media_assets using post_id
)
```

**Key:** `LIMIT 5 BY facebook_id` -- returns top/bottom 5 per competitor.

---

#### 4.5.5 `getTopHashtagsQuery($limit)` -- Top Hashtags

**Tables:** `facebook_competitor_posts`, `facebook_competitor_insights`

```sql
WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM facebook_competitor_posts
    WHERE {pageFilters} AND {dateFilters}
    group by post_id
)
SELECT
    groupUniqArray(hashtags_business_ids) as companies_using,
    groupUniqArray(hashtags_name) as companies_name,
    sum(hashtag_count) as count,
    sum(hashtags_total_engagement) as total_engagement,
    groupUniqArray(followers_total_count) as total_followers,
    round(sum(engagement_per_follower),2) as engagement_per_follower,
    round(sum(engagement_rate_by_follower),2) as engagement_rate_by_follower,
    round(sum(engagement_per_post),2) as engagement_per_post,
    tag
from (
    SELECT
        hashtags.facebook_id as hashtags_business_ids,
        hashtags.name as hashtags_name,
        hashtags.count as hashtag_count,
        hashtags.total_engagement as hashtags_total_engagement,
        round((hashtags.total_engagement / hashtags.count),2) as engagement_per_post,
        followers.total_followers as followers_total_count,
        round((hashtags.total_engagement / followers.total_followers),2) as engagement_per_follower,
        round((hashtags.total_engagement / followers.total_followers / hashtags.count) * 100,2) as engagement_rate_by_follower,
        hashtags.tag as tag
    from (
        SELECT max(followers_count) as total_followers, page_id
        FROM facebook_competitor_insights
        WHERE page_id IN {accountIds}
        group by page_id
    ) as followers
    left join (
        SELECT page_name as name, arrayJoin(hashtags) as tag,
            uniq(post_id) as count, facebook_id as facebook_id,
            sum(post_engagement) as total_engagement
        from facebook_competitor_posts
        where length(hashtags) > 0 AND (post_id, inserted_at) IN (posts)
        group by tag, facebook_id, page_name
        order by count desc
    ) as hashtags on followers.page_id = hashtags.facebook_id
)
where tag != ''
group by tag
order by count desc
limit {limit}
```

**Key:** `arrayJoin(hashtags)` -- hashtags is a native ClickHouse array in `facebook_competitor_posts`.

---

#### 4.5.6 `getIndividualHashtagQuery($hashtag)` -- Single Hashtag Detail

**Tables:** `facebook_competitor_posts`, `facebook_competitor_insights`

```sql
WITH posts as (...)
SELECT * FROM (
    SELECT arrayJoin(hashtags) as tag,
        uniq(post_id) as count,
        sum(post_engagement) as total_engagement,
        max(followers_count) as total_followers,
        round((total_engagement / count), 2) as engagement_per_post,
        round((total_engagement / total_followers), 2) as engagement_per_follower,
        round((total_engagement / total_followers / count) * 100, 2) as engagement_rate_by_follower,
        facebook_id
    FROM facebook_competitor_posts
    where length(hashtags) > 0 AND tag = '{hashtag}' AND (post_id, inserted_at) IN (posts)
    group by tag, facebook_id
    order by total_engagement desc
) as hashtags_statistics
LEFT JOIN (
    select argMax(profile_picture_url, inserted_at) as image,
           argMax(page_name, inserted_at) as name,
           multiIf(...) as slug,
           argMax(followers_count, inserted_at) as followersCount,
           page_id
    from facebook_competitor_insights
    group by page_id
) as page_constants ON page_id = hashtags_statistics.facebook_id
```

---

#### 4.5.7 `getBiographyQuery($sortOrder)` -- Biography Data

**Tables:** `facebook_competitor_posts`, `facebook_competitor_insights`

```sql
SELECT * FROM (
    select last_value(biography) as biography,
           lengthUTF8(biography) as biography_length,
           facebook_id,
           multiIf(...) as state,
           multiIf(...) as slug
    from facebook_competitor_posts
    WHERE {pageFilters}
    group by facebook_id
    ORDER BY max(inserted_at) DESC
) as biography_statistics
LEFT JOIN (
    select argMax(profile_picture_url, inserted_at) as image,
           argMax(page_name, inserted_at) as name,
           argMax(followers_count, inserted_at) as followersCount,
           page_id
    from facebook_competitor_insights
    group by page_id
) as page_constants ON page_id = biography_statistics.facebook_id
ORDER BY {sortOrder} DESC
```

**Note:** No date filter on biography -- always shows latest value.

---

#### 4.5.8 `getFollowersGrowthComparisonQuery($sortOrder)` -- Followers Growth Over Time

**Table:** `facebook_competitor_insights`

```sql
SELECT facebook_id,
    multiIf(...) as name,
    multiIf(...) as image,
    multiIf(...) as state,
    multiIf(...) as slug,
    dates, followers_count, dates_with_followers_count
FROM (SELECT c1 as facebook_id FROM VALUES {accountIds}) as page_ids
LEFT JOIN (
    SELECT facebook_id,
        groupArray(date) as dates,
        groupArray(followers_count) as followers_count,
        arrayZip(dates, followers_count) as dates_with_followers_count
    FROM (
        select page_name, profile_picture_url,
               page_id as facebook_id,
               toDate(inserted_at) as date,
               toInt32(followers_count) as followers_count
        from facebook_competitor_insights
        WHERE {pageFilters_insights} AND {dateFilters_insights}
        order by date ASC
    )
    group by facebook_id
) as page_insights on page_ids.facebook_id = page_insights.facebook_id
ORDER BY {sortOrder} ASC
```

Returns `dates_with_followers_count` as a zipped array of `[date, followers_count]` tuples.

---

#### 4.5.9 `getPostReactDistributionByCompany($sortOrder, $facebook_id)` -- Per-Reaction Breakdown

**Table:** `facebook_competitor_posts`

```sql
SELECT facebook_id, page_name,
    'https://graph.facebook.com/' || facebook_id || '/picture' as image,
    toInt32(sum(like)) as total_likes,
    toInt32(sum(haha)) as total_hahas,
    toInt32(sum(angry)) as total_angry,
    toInt32(sum(sad)) as total_sad,
    toInt32(sum(thankful)) as total_thankful,
    toInt32(sum(love)) as total_love,
    toInt32(sum(wow)) as total_wow,
    toInt32(sum(total_post_reactions)) as total_post_reactions,
    toInt32(sum(comments)) as comments,
    toInt32(sum(shares)) as shares,
    toInt32(count()) as total_posts
FROM facebook_competitor_posts
WHERE facebook_id = '{facebook_id}' AND {dateFilters}
GROUP BY facebook_id, page_name
```

**Note:** No deduplication CTE -- reads all rows directly.

---

#### 4.5.10 `getPostEngagementOverTime($sortOrder, $facebook_id)` -- Daily Engagement Time Series

**Table:** `facebook_competitor_posts`

```sql
SELECT
    toInt32(sum(post_engagement)) as total_engagements,
    toInt32(count()) as total_posts,
    toDate(created_at) as date
FROM facebook_competitor_posts
WHERE facebook_id = '{facebook_id}' AND {dateFilters}
GROUP BY toDate(created_at)
ORDER BY toDate(created_at) ASC
WITH FILL FROM toDate('{startDate}') TO toDate('{endDate}') STEP 1
```

---

#### 4.5.11 `getPostEngagementByCompetitor($sortOrder)` -- Total Engagement Per Competitor

**Table:** `facebook_competitor_posts`

```sql
select facebook_id, page_name,
       'https://graph.facebook.com/' || facebook_id || '/picture' as image,
       toInt32(count()) as total_posts,
       toInt32(sum(post_engagement)) as total_engagements
from (
    select facebook_id, page_name, post_id,
           max(post_engagement) as post_engagement
    FROM facebook_competitor_posts
    WHERE {pageFilters} AND {dateFilters}
    group by facebook_id, page_name, post_id
)
group by facebook_id, page_name
ORDER BY total_engagements DESC
```

Deduplication is done inline: `max(post_engagement)` per `post_id` before aggregating.

---

## 5. Instagram Analytics

### 5.1 Request Parameters

Same as Facebook but uses `instagram_id` instead of `facebook_id`. Media types default to `['REELS','IMAGE','VIDEO','CAROUSEL_ALBUM']`.

### 5.2 Endpoints and SQL Queries

#### 5.2.1 Summary (`summaryQuery`)

**Route:** `POST overview/instagram/summary`

**Tables:** `instagram_posts`, `instagram_insights`

```sql
WITH posts AS (
    SELECT media_id, max(stored_event_at)
    FROM instagram_posts
    WHERE instagram_id in {instagram_id}
      AND toDateTime(post_created_at,0,'{timezone}') BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
    GROUP BY media_id
)
SELECT
    toInt32(sum(total_engagement))   as post_engagement,
    toInt32(sum(post_views))         as post_views,
    toInt32(sum(total_reach))        as post_reach,
    toInt32(sum(total_saves))        as post_saves,
    toInt32(sum(reactions))          as post_reactions,
    toInt32(sum(comments))           as post_comments,
    toInt32(sum(profile_views))      as profile_views,
    toInt32(first_value(followers_count))  as followers_count,
    toInt32(first_value(follows_count))    as follows_count,
    toInt32(sum(accounts_engaged))   as accounts_engaged,
    toInt32(sum(engagement))         as profile_engagement,
    toInt32(sum(impressions))        as profile_impressions,
    toInt32(sum(reach))              as profile_reach,
    toInt32(sum(doc_count))          as doc_count,
    toInt32(sum(total_stories))      as total_stories,
    toInt32(sum(total_posts))        as total_posts,
    round(if(doc_count > 0, toFloat32(post_engagement/doc_count), 0), 2) as eng_rate
FROM (
    -- posts subquery with LEFT JOIN insights subquery
    -- Posts: count, SUM(comments_count), SUM(engagement), SUM(impressions), SUM(reach), SUM(saved), SUM(like_count), SUM(views)
    -- Stories counted via: countIf(entity_type = 'STORY' OR media_type = 'STORY')
    -- Insights: SUM(daily_profile_views), argMaxIf(daily_followers_count, date_bucket, daily_followers_count > 0), etc.
)
```

---

#### 5.2.2 Audience Growth (`audienceQuery`)
#### 3.3.8 Page Impressions (`getOverviewImpressionsQuery`)

**Route:** `POST overview/facebook/pageImpressions`

**Table:** `facebook_insights`

```sql
SELECT
    groupArray(page_impressions) as page_impressions,
    groupArray(created_date) as buckets
FROM (
    SELECT toInt32(max(page_impressions)) as page_impressions,
        toDate(created_time) as created_date
    FROM facebook_insights
    WHERE page_id = '{facebookId}'
    AND {dateFilter}
    GROUP BY created_date
    ORDER BY created_date ASC
    WITH FILL FROM toDate('{startDate}') TO toDate('{endDate}') + 1 STEP 1
)
```

**Rollup:** `total_impressions`, `avg_impressions_per_day` (sum/totalDays), `avg_impressions_per_week` (sum/totalWeeks).

---

#### 3.3.9 Page Engagements (`getOverviewEngagementsQuery`)

**Route:** `POST overview/facebook/engagement`

**Table:** `facebook_insights`

```sql
SELECT
    groupArray(page_engagements) as page_engagements,
    groupArray(created_date) as buckets
FROM (
    SELECT toInt32(max(page_post_engagements)) as page_engagements,
        toDate(created_time) as created_date
    FROM facebook_insights
    WHERE page_id = '{facebookId}'
    AND {dateFilter}
    GROUP BY created_date
    ORDER BY created_date ASC
    WITH FILL FROM toDate('{startDate}') TO toDate('{endDate}') + 1 STEP 1
)
```

---

#### 3.3.10 Reels Analytics (`getOverviewReelsAnalyticsQuery`)

**Route:** `POST overview/facebook/reelsAnalytics`

**Tables:** `facebook_posts`, `facebook_reels_insights`

```sql
WITH facebook_post_data AS (
    SELECT post_id,
         last_value(total) as reactions,
         last_value(comments) as comments,
         last_value(shares) as repost,
         reactions+comments+repost as total_engagement,
         toDate(created_time) as created_at
    FROM facebook_posts
    WHERE page_id in ({facebookId}) AND {dateFilter_created_at}
    GROUP BY post_id, created_at
)
SELECT
    groupArray(created_at) as buckets,
    groupArray(total_reels_count) as total_reels,
    groupArray(total_seconds_watched) as total_seconds_watched,
    groupArray(initial_plays) as initial_plays,
    groupArray(total_engagement) as engagement,
    groupArray(reactions) as reactions,
    groupArray(comments) as comments,
    groupArray(shares) as shares,
    toInt32(sum(total_reels_count)) as show_data
FROM (
    SELECT created_at,
        toInt32(count()) as total_reels_count,
        round(sum(total_time_watched_in_ms) / 1000, 2) as total_seconds_watched,
        toInt32(sum(play_count)) as initial_plays,
        toInt32(sum(total_engagement)) as total_engagement,
        toInt32(sum(reactions)) as reactions,
        toInt32(sum(comments)) as comments,
        toInt32(sum(repost)) as shares
    FROM (
        SELECT post_id, toDate(created_at) as created_at,
            last_value(average_time_watched) as average_time_watched,
            last_value(total_time_watched_in_ms) as total_time_watched_in_ms,
            last_value(play_count) as play_count,
            last_value(impressions_unique) as reach
        FROM facebook_reels_insights
        WHERE page_id in ({facebookId}) AND {dateFilter_created_at}
        GROUP BY post_id, created_at
    ) AS reels
    LEFT JOIN facebook_post_data ON facebook_post_data.post_id = reels.post_id
    GROUP BY reels.created_at
    ORDER BY created_at ASC
    WITH FILL TO (toDate('{endDate+1day}'))
)
```

**Note:** `total_time_watched_in_ms` is divided by 1000 for seconds.

---

#### 3.3.11 Video Insights (`getVideoInsightsQuery`)

**Route:** `POST overview/facebook/videoInsights`

**Tables:** `facebook_posts`, `facebook_video_insights`

Joins `facebook_posts` (filtered to `media_type='videos'`) with `facebook_video_insights` on `post_id`. View times are in ms, divided by 1000 for seconds.

---

#### 3.3.12 Time Recommendation (`getTimeRecommendationQuery`)

**Table:** `facebook_posts`

```sql
SELECT
    max(page_id) as facebookId,
    day_of_week,
    hour_of_day,
    sum(post_impressions) as post_impressions,
    sum(total_engagement) as total_engagement
FROM facebook_posts
WHERE page_id in {facebookId} AND {dateFilter}
GROUP BY day_of_week, hour_of_day
ORDER BY day_of_week, hour_of_day
```

---

#### 3.3.13 Demographics

**Audience Gender:** Reads `page_fans_gender` blob from `facebook_insights`. Parsed from `$`-delimited pairs. Known quirk: M/U/F labels mapped to wrong aliases.

**Audience Age:** Reads `page_fans_age` from most recent record. Age buckets: `65+`, `55-64`, `45-54`, `35-44`, `25-34`, `18-34`, `13-17`. Date boundary: if range crosses 2024-03-14, end date is clamped.

**Audience Country/City:** Reads `page_fans_country`/`page_fans_city` from most recent record. `$`-delimited parsing.

**Max Gender+Age:** Reads `page_fans_gender_age` blob. Finds the single bucket with highest value.

---

## 4. Facebook Competitor Analytics

### 4.1 Constructor and Timezone Handling

**CRITICAL difference from own analytics:** Competitor dates are parsed in user timezone, converted to UTC before embedding in SQL. Database timestamps are stored in UTC.

```
startDate = Carbon::parse(date[0], timezone)->setTimezone('UTC')
endDate = Carbon::parse(date[1], timezone)->setTime(23,59,59)->setTimezone('UTC')
```

### 4.2 Tables

`facebook_competitor_posts`, `facebook_competitor_insights`, `facebook_competitor_media_assets`

### 4.3 Helper Methods

#### `getPageFilters($insightsFlag, $filterAdded)`
Builds `facebook_id IN ('id1','id2',...)` for posts or `page_id IN (...)` for insights. If `$filterAdded = true`, excludes accounts with `state == "Added"`.

#### `getDateFilters($insightsFlag, $dateColumn)`
```sql
-- posts: created_at BETWEEN toDateTime('{startUTC}',0) AND toDateTime('{endUTC}',0)
-- insights: inserted_at BETWEEN toDateTime('{startUTC}',0) AND toDateTime('{endUTC}',0)
```

#### `getConstantConditions($field, $idField)`
Builds `multiIf(id='x','val', id='y','val', 'Random')` to map competitor IDs to metadata (name, image, state, slug).

### 4.4 Controller Methods

| Method | Description | Sort Default |
|--------|-------------|--------------|
| `searchCompetitor` | Graph API page search with appsecret_proof | N/A |
| `dataTableMetrics` | Current/previous period KPIs with growth % | `followersCount` |
| `postingActivityGraphByTypes` | Aggregate by media_type | `avgTotalEngagements` |
| `postingActivityBySpecificType` | Per-competitor for one media_type | N/A |
| `topAndLeastPerformingPosts` | Top-5/bottom-5 per competitor with media assets | engagement DESC/ASC |
| `topHashtags` | Top hashtags by engagement (limit default: 7) | count DESC |
| `individualHashtagData` | Per-competitor stats for one hashtag | engagement DESC |
| `biographyData` | Biography + length per competitor | `biography_length` |
| `followersGrowthComparison` | Per-competitor followers time-series | `followers_count` |
| `postReactDistribution` | Engagement totals for single page | N/A |
| `postReactDistributionByCompany` | Per-reaction breakdown for single page | N/A |
| `postTypeDistribution` | Per-media-type breakdown per competitor | N/A |
| `postEngagementOverTime` | Daily time-series per competitor (single page) | N/A |
| `postEngagementByCompetitor` | Total engagement per competitor | engagement DESC |

### 4.5 SQL Queries

#### 4.5.1 `getDataTableDataMetricsQuery($sortOrder)` -- Main Data Table

**Tables:** `facebook_competitor_posts`, `facebook_competitor_insights`

```sql
WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM facebook_competitor_posts
    WHERE {pageFilters} AND {dateFilters}
    group by post_id
)
SELECT
    multiIf(facebook_id='x','imgUrl',...) as image,
    multiIf(facebook_id='x','name',...) as name,
    multiIf(facebook_id='x','state',...) as state,
    * FROM (
    SELECT pages_ids.facebook_id as facebook_id,
        week_metrics.averageEngagement,
        week_metrics.averagePostsPerWeek,
        week_metrics.engagementRate,
        days_metrics.dayOfWeek,
        days_metrics.hourOfDay,
        days_metrics.averagePostsPerDay,
        days_metrics.averagePostsPerDayEngagement
    FROM (SELECT CAST(c1 AS String) as facebook_id FROM VALUES {accountIds}) as pages_ids
    LEFT JOIN (
        SELECT facebook_id,
            round(avg(total_engagement), 2) as averageEngagement,
            if(dateDiff('week', ...) != 0,
               round(sum(posts_in_a_week)/dateDiff('week',...), 2), 0) as averagePostsPerWeek,
            round((sum(total_engagement)/sum(posts_in_a_week)) / max(followers_count) * 100, 2) as engagementRate
        FROM (
            SELECT CAST(facebook_id AS String) AS facebook_id,
                sum(post_engagement) as total_engagement,
                toStartOfWeek(created_at) as week,
                count() as posts_in_a_week,
                followers_count
            FROM facebook_competitor_posts
            WHERE (post_id, inserted_at) IN (posts)
            group by facebook_id, week, followers_count
            order by week ASC
            WITH FILL FROM toStartOfWeek(toDate('{startDate}'))
                       TO toStartOfWeek(toDate('{endDate}'))
                       STEP INTERVAL 1 WEEK
        )
        group by facebook_id
    ) as week_metrics ON ...
    LEFT JOIN (
        SELECT facebook_id,
            argMax(multiIf(day_of_week=1,'Monday',...), total_posts) as dayOfWeek,
            argMax(multiIf(hour_of_day=0,'12:00 AM',...), total_posts) as hourOfDay,
            if(dateDiff('day',...) != 0,
               round(sum(total_posts)/dateDiff('day',...), 2), 0) as averagePostsPerDay,
            avg(total_engagement) as averagePostsPerDayEngagement,
            argMax(hour_of_day, total_posts) as maximumPostsHour,
            argMax(hour_of_day, total_engagement) as mostEngagementHour
        from (
            select facebook_id, ...,
                sum(post_engagement) as total_engagement,
                toDayOfWeek(created_at) as day_of_week,
                toHour(created_at) as hour_of_day,
                uniq(post_id) as total_posts,
                toDate(created_at) as date_c
            from facebook_competitor_posts
            WHERE (post_id, inserted_at) IN (posts)
            group by facebook_id, day_of_week, hour_of_day, date_c
            order by date_c ASC
            WITH FILL FROM toDate('{startDate}') TO toDate('{endDate}') STEP 1
        )
        group by facebook_id
    ) as days_metrics ON ...
) as page_metrics
LEFT JOIN (
    SELECT max(followers_count) as followersCount,
           max(total_fan_count) as fanCount,
           page_id
    FROM facebook_competitor_insights
    WHERE {pageFilters_insights} AND {dateFilters_insights}
    group by page_id
) as page_metadata ON page_metrics.facebook_id = page_metadata.page_id
ORDER BY {sortOrder} DESC
```

**Returns per competitor:** `facebook_id`, `image`, `name`, `state`, `averageEngagement`, `averagePostsPerWeek`, `engagementRate`, `dayOfWeek` (best day), `hourOfDay` (best hour), `averagePostsPerDay`, `averagePostsPerDayEngagement`, `followersCount`, `fanCount`.

---

#### 4.5.2 `getPostingActivityGraphByTypesQuery($sortOrder)` -- Activity by Media Type (Aggregated)

**Table:** `facebook_competitor_posts`

```sql
WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM facebook_competitor_posts
    WHERE {pageFilters} AND {dateFilters}
    group by post_id
)
SELECT
    media_type as mediaType,
    round(sum(avg_engagement),2) as avgTotalEngagements,
    sum(page_post_count) as totalPosts,
    toInt32(sum(total_engagement)) as total_engagement,
    round(sum(er),2) as avgEngagementRate,
    if(dateDiff('week',...) = 0, 0, round(totalPosts/dateDiff('week',...),2)) as postsPerWeek,
    if(dateDiff('day',...) = 0, 0, round(totalPosts/dateDiff('day',...),2)) as postsPerDay,
    if(dateDiff('hour',...) = 0, 0, round(totalPosts/dateDiff('hour',...),2)) as postsPerHour,
    dateDiff('week',...) as weekCount,
    dateDiff('day',...) as dayCount,
    dateDiff('hour',...) as hourCount
FROM (
    SELECT facebook_id, page_name, media_type,
        sum(count) as page_post_count,
        sum(total_post_engagement) as total_engagement,
        if(page_post_count<=0,0, round(sum(total_post_engagement)/page_post_count,2)) as avg_engagement,
        if(page_post_count<=0,0, round(((sum(er)/page_post_count)/max(followers_count))*100,2)) as er
    FROM (
        SELECT facebook_id, page_name, media_type,
            count() as count,
            sum(post_engagement) as total_post_engagement,
            total_post_engagement as er,
            argMax(followers_count, created_at) as followers_count,
            toStartOfWeek(created_at) as week
        from facebook_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        group by facebook_id, page_name, media_type, week
        ORDER by week asc
        WITH FILL FROM toStartOfWeek(toDate('{startDate}'))
                   TO toStartOfWeek(toDate('{endDate}'))
                   STEP INTERVAL 1 WEEK
    )
    WHERE media_type!=''
    GROUP BY facebook_id, page_name, media_type
)
group by media_type
```

Aggregates **all competitors together** by `media_type`.

---

#### 4.5.3 `getPostingActivityBySpecificTypeQuery($mediaType, $sortOrder)`

Same structure as above but WHERE clause adds `media_type = '$mediaType'` and groups per `facebook_id`. Adds per-competitor metadata from `facebook_competitor_insights`. Returns one row per competitor for that media type.

---

#### 4.5.4 `getTopPerformingAndLeastPerformingPostsQuery()` -- Top+Least Posts

**Tables:** `facebook_competitor_posts`, `facebook_competitor_media_assets`

```sql
WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM facebook_competitor_posts
    WHERE {pageFilters} AND {dateFilters}
    group by post_id
)
SELECT * FROM (
    SELECT *, 'top_5_posts' as category,
        'https://graph.facebook.com/' || facebook_id || '/picture?type=large' as image
    FROM (
        SELECT * FROM facebook_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        ORDER BY post_engagement desc
        LIMIT 5 BY facebook_id
    ) as post_data
    LEFT JOIN (
        SELECT * FROM facebook_competitor_media_assets
        WHERE {pageFilters_insights} AND created_at BETWEEN ...
        order by inserted_at desc
    ) as media_assets using post_id
UNION ALL
    SELECT *, 'least_5_posts' as category, ...
    FROM (
        SELECT * FROM facebook_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        ORDER BY post_engagement asc
        LIMIT 5 BY facebook_id
    ) as post_data
    LEFT JOIN (...) as media_assets using post_id
)
```

**Key:** `LIMIT 5 BY facebook_id` -- returns top/bottom 5 per competitor.

---

#### 4.5.5 `getTopHashtagsQuery($limit)` -- Top Hashtags

**Tables:** `facebook_competitor_posts`, `facebook_competitor_insights`

```sql
WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM facebook_competitor_posts
    WHERE {pageFilters} AND {dateFilters}
    group by post_id
)
SELECT
    groupUniqArray(hashtags_business_ids) as companies_using,
    groupUniqArray(hashtags_name) as companies_name,
    sum(hashtag_count) as count,
    sum(hashtags_total_engagement) as total_engagement,
    groupUniqArray(followers_total_count) as total_followers,
    round(sum(engagement_per_follower),2) as engagement_per_follower,
    round(sum(engagement_rate_by_follower),2) as engagement_rate_by_follower,
    round(sum(engagement_per_post),2) as engagement_per_post,
    tag
from (
    SELECT
        hashtags.facebook_id as hashtags_business_ids,
        hashtags.name as hashtags_name,
        hashtags.count as hashtag_count,
        hashtags.total_engagement as hashtags_total_engagement,
        round((hashtags.total_engagement / hashtags.count),2) as engagement_per_post,
        followers.total_followers as followers_total_count,
        round((hashtags.total_engagement / followers.total_followers),2) as engagement_per_follower,
        round((hashtags.total_engagement / followers.total_followers / hashtags.count) * 100,2) as engagement_rate_by_follower,
        hashtags.tag as tag
    from (
        SELECT max(followers_count) as total_followers, page_id
        FROM facebook_competitor_insights
        WHERE page_id IN {accountIds}
        group by page_id
    ) as followers
    left join (
        SELECT page_name as name, arrayJoin(hashtags) as tag,
            uniq(post_id) as count, facebook_id as facebook_id,
            sum(post_engagement) as total_engagement
        from facebook_competitor_posts
        where length(hashtags) > 0 AND (post_id, inserted_at) IN (posts)
        group by tag, facebook_id, page_name
        order by count desc
    ) as hashtags on followers.page_id = hashtags.facebook_id
)
where tag != ''
group by tag
order by count desc
limit {limit}
```

**Key:** `arrayJoin(hashtags)` -- hashtags is a native ClickHouse array in `facebook_competitor_posts`.

---

#### 4.5.6 `getIndividualHashtagQuery($hashtag)` -- Single Hashtag Detail

**Tables:** `facebook_competitor_posts`, `facebook_competitor_insights`

```sql
WITH posts as (...)
SELECT * FROM (
    SELECT arrayJoin(hashtags) as tag,
        uniq(post_id) as count,
        sum(post_engagement) as total_engagement,
        max(followers_count) as total_followers,
        round((total_engagement / count), 2) as engagement_per_post,
        round((total_engagement / total_followers), 2) as engagement_per_follower,
        round((total_engagement / total_followers / count) * 100, 2) as engagement_rate_by_follower,
        facebook_id
    FROM facebook_competitor_posts
    where length(hashtags) > 0 AND tag = '{hashtag}' AND (post_id, inserted_at) IN (posts)
    group by tag, facebook_id
    order by total_engagement desc
) as hashtags_statistics
LEFT JOIN (
    select argMax(profile_picture_url, inserted_at) as image,
           argMax(page_name, inserted_at) as name,
           multiIf(...) as slug,
           argMax(followers_count, inserted_at) as followersCount,
           page_id
    from facebook_competitor_insights
    group by page_id
) as page_constants ON page_id = hashtags_statistics.facebook_id
```

---

#### 4.5.7 `getBiographyQuery($sortOrder)` -- Biography Data

**Tables:** `facebook_competitor_posts`, `facebook_competitor_insights`

```sql
SELECT * FROM (
    select last_value(biography) as biography,
           lengthUTF8(biography) as biography_length,
           facebook_id,
           multiIf(...) as state,
           multiIf(...) as slug
    from facebook_competitor_posts
    WHERE {pageFilters}
    group by facebook_id
    ORDER BY max(inserted_at) DESC
) as biography_statistics
LEFT JOIN (
    select argMax(profile_picture_url, inserted_at) as image,
           argMax(page_name, inserted_at) as name,
           argMax(followers_count, inserted_at) as followersCount,
           page_id
    from facebook_competitor_insights
    group by page_id
) as page_constants ON page_id = biography_statistics.facebook_id
ORDER BY {sortOrder} DESC
```

**Note:** No date filter on biography -- always shows latest value.

---

#### 4.5.8 `getFollowersGrowthComparisonQuery($sortOrder)` -- Followers Growth Over Time

**Table:** `facebook_competitor_insights`

```sql
SELECT facebook_id,
    multiIf(...) as name,
    multiIf(...) as image,
    multiIf(...) as state,
    multiIf(...) as slug,
    dates, followers_count, dates_with_followers_count
FROM (SELECT c1 as facebook_id FROM VALUES {accountIds}) as page_ids
LEFT JOIN (
    SELECT facebook_id,
        groupArray(date) as dates,
        groupArray(followers_count) as followers_count,
        arrayZip(dates, followers_count) as dates_with_followers_count
    FROM (
        select page_name, profile_picture_url,
               page_id as facebook_id,
               toDate(inserted_at) as date,
               toInt32(followers_count) as followers_count
        from facebook_competitor_insights
        WHERE {pageFilters_insights} AND {dateFilters_insights}
        order by date ASC
    )
    group by facebook_id
) as page_insights on page_ids.facebook_id = page_insights.facebook_id
ORDER BY {sortOrder} ASC
```

Returns `dates_with_followers_count` as a zipped array of `[date, followers_count]` tuples.

---

#### 4.5.9 `getPostReactDistributionByCompany($sortOrder, $facebook_id)` -- Per-Reaction Breakdown

**Table:** `facebook_competitor_posts`

```sql
SELECT facebook_id, page_name,
    'https://graph.facebook.com/' || facebook_id || '/picture' as image,
    toInt32(sum(like)) as total_likes,
    toInt32(sum(haha)) as total_hahas,
    toInt32(sum(angry)) as total_angry,
    toInt32(sum(sad)) as total_sad,
    toInt32(sum(thankful)) as total_thankful,
    toInt32(sum(love)) as total_love,
    toInt32(sum(wow)) as total_wow,
    toInt32(sum(total_post_reactions)) as total_post_reactions,
    toInt32(sum(comments)) as comments,
    toInt32(sum(shares)) as shares,
    toInt32(count()) as total_posts
FROM facebook_competitor_posts
WHERE facebook_id = '{facebook_id}' AND {dateFilters}
GROUP BY facebook_id, page_name
```

**Note:** No deduplication CTE -- reads all rows directly.

---

#### 4.5.10 `getPostEngagementOverTime($sortOrder, $facebook_id)` -- Daily Engagement Time Series

**Table:** `facebook_competitor_posts`

```sql
SELECT
    toInt32(sum(post_engagement)) as total_engagements,
    toInt32(count()) as total_posts,
    toDate(created_at) as date
FROM facebook_competitor_posts
WHERE facebook_id = '{facebook_id}' AND {dateFilters}
GROUP BY toDate(created_at)
ORDER BY toDate(created_at) ASC
WITH FILL FROM toDate('{startDate}') TO toDate('{endDate}') STEP 1
```

---

#### 4.5.11 `getPostEngagementByCompetitor($sortOrder)` -- Total Engagement Per Competitor

**Table:** `facebook_competitor_posts`

```sql
select facebook_id, page_name,
       'https://graph.facebook.com/' || facebook_id || '/picture' as image,
       toInt32(count()) as total_posts,
       toInt32(sum(post_engagement)) as total_engagements
from (
    select facebook_id, page_name, post_id,
           max(post_engagement) as post_engagement
    FROM facebook_competitor_posts
    WHERE {pageFilters} AND {dateFilters}
    group by facebook_id, page_name, post_id
)
group by facebook_id, page_name
ORDER BY total_engagements DESC
```

Deduplication is done inline: `max(post_engagement)` per `post_id` before aggregating.

---

## 5. Instagram Analytics

### 5.1 Request Parameters

Same as Facebook but uses `instagram_id` instead of `facebook_id`. Media types default to `['REELS','IMAGE','VIDEO','CAROUSEL_ALBUM']`.

### 5.2 Endpoints and SQL Queries

#### 5.2.1 Summary (`summaryQuery`)

**Route:** `POST overview/instagram/summary`

**Tables:** `instagram_posts`, `instagram_insights`

```sql
WITH posts AS (
    SELECT media_id, max(stored_event_at)
    FROM instagram_posts
    WHERE instagram_id in {instagram_id}
      AND toDateTime(post_created_at,0,'{timezone}') BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
    GROUP BY media_id
)
SELECT
    toInt32(sum(total_engagement))   as post_engagement,
    toInt32(sum(post_views))         as post_views,
    toInt32(sum(total_reach))        as post_reach,
    toInt32(sum(total_saves))        as post_saves,
    toInt32(sum(reactions))          as post_reactions,
    toInt32(sum(comments))           as post_comments,
    toInt32(sum(profile_views))      as profile_views,
    toInt32(first_value(followers_count))  as followers_count,
    toInt32(first_value(follows_count))    as follows_count,
    toInt32(sum(accounts_engaged))   as accounts_engaged,
    toInt32(sum(engagement))         as profile_engagement,
    toInt32(sum(impressions))        as profile_impressions,
    toInt32(sum(reach))              as profile_reach,
    toInt32(sum(doc_count))          as doc_count,
    toInt32(sum(total_stories))      as total_stories,
    toInt32(sum(total_posts))        as total_posts,
    round(if(doc_count > 0, toFloat32(post_engagement/doc_count), 0), 2) as eng_rate
FROM (
    -- posts subquery with LEFT JOIN insights subquery
    -- Posts: count, SUM(comments_count), SUM(engagement), SUM(impressions), SUM(reach), SUM(saved), SUM(like_count), SUM(views)
    -- Stories counted via: countIf(entity_type = 'STORY' OR media_type = 'STORY')
    -- Insights: SUM(daily_profile_views), argMaxIf(daily_followers_count, date_bucket, daily_followers_count > 0), etc.
)
```

---
**Table:** `instagram_insights`

```sql
SELECT
    notEmpty(follower_count_temp) AS show_data,
    arrayFill(x -> not x == 0, follower_count_temp) AS followers,
    followers_daily,
    dates AS buckets
FROM (
    SELECT
        groupArray(follower_count) AS follower_count_temp,
        arrayConcat(
            [toInt32(0)],
            arrayMap(x -> toInt32(x),
                arraySlice(
                    arrayDifference(arrayFill(x -> not (x == 0), follower_count_temp)),
                    2
                )
            )
        ) AS followers_daily,
        groupArray(dates) AS dates
    FROM (
        SELECT
            toDate(created_time) AS dates,
            toInt32(argMin(followers_count, stored_event_at)) AS follower_count
        FROM instagram_insights
        WHERE instagram_id IN {instagram_id}
          AND toDateTime(created_time,0,'{timezone}') BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
        GROUP BY dates
        ORDER BY dates ASC
        WITH FILL FROM toDate('{currentStartDate}') TO toDate('{currentEndDate}') + 1 STEP 1
    )
)
```

**Zero-fill fallback:** If `followers[0] == 0`, looks back 2 years via `getLastFollowersCount()`.

---

#### 5.2.3 Last Followers Count (Fallback) (`getLastFollowersCount`)

**Table:** `instagram_insights`

Used when `followers[0] == 0`. Looks back 2 years.

```sql
SELECT arrayFirst(x -> x != 0, groupArray(followers)) AS followers
FROM (
    SELECT
        last_value(created_time) as inserted_time,
        toInt32(last_value(followers_count)) as followers
    FROM instagram_insights
    WHERE instagram_id in {instagram_id}
      AND toDateTime(created_time,0,'{timezone}') BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
      AND followers_count != 0
    GROUP BY record_id
    ORDER BY inserted_time DESC
)
```

---

#### 5.2.4 Audience Rollup (`audienceRollupQuery`)

**Table:** `instagram_insights`

```sql
SELECT
    toInt32(argMaxIf(followers, date_bucket, followers > 0)) as follower_count,
    toInt32(last_value(followers) - first_value(followers))  as follower_gained,
    max(date_bucket) as dates
FROM (
    SELECT
        toInt32(argMin(followers_count, stored_event_at)) AS followers,
        toDate(created_time) AS date_bucket
    FROM instagram_insights
    WHERE instagram_id in {instagram_id}
      AND toDateTime(created_time,0,'{timezone}') BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
      AND followers_count > 0
    GROUP BY date_bucket
    ORDER BY date_bucket ASC
)
```

---

#### 5.2.5 Publishing Behaviour (`getOverviewPublishingBehaviourByMediaTypeQuery`)

**Route:** `POST overview/instagram/publish`

**Table:** `instagram_posts`

```sql
WITH posts AS (
    SELECT media_id, last_value(stored_event_at)
    FROM instagram_posts
    WHERE instagram_id in {instagram_id}
      AND toDateTime(post_created_at,0,'{timezone}') BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
      AND media_type in ('{media_type}')
    GROUP BY media_id
)
SELECT
    groupArray(likes)       as likes,
    groupArray(comments)    as comments,
    groupArray(saved)       as saved,
    groupArray(engagement)  as engagement,
    groupArray(reach)       as reach,
    groupArray(impressions) as impressions,
    groupArray(views)       as views,
    groupArray(total_posts) as total_posts,
    groupArray(created_at)  as buckets
FROM (
    SELECT
        toInt32(sum(like_count))      as likes,
        toInt32(sum(comments_count))  as comments,
        toInt32(sum(saved))           as saved,
        toInt32(sum(engagement))      as engagement,
        toInt32(sum(reach))           as reach,
        toInt32(sum(impressions))     as impressions,
        toInt32(count(*))             as total_posts,
        toInt32(sum(views))           as views,
        toDate(post_created_at)       as created_at
    FROM (
        SELECT
            media_id,
            last_value(like_count)      as like_count,
            last_value(comments_count)  as comments_count,
            last_value(saved)           as saved,
            last_value(engagement)      as engagement,
            last_value(reach)           as reach,
            last_value(impressions)     as impressions,
            last_value(views)           as views,
            post_created_at
        FROM instagram_posts
        WHERE (media_id, stored_event_at) in posts
        GROUP BY media_id, post_created_at
    )
    GROUP BY created_at
    ORDER BY created_at ASC
    WITH FILL FROM toDate('{currentStartDate}') TO toDate('{currentEndDate}') + 1
)
```

---

#### 5.2.6 Publish Rollup (`publishRollupQuery`)

**Table:** `instagram_posts`

Returns one row per media type plus a TOTAL row.

```sql
WITH
    posts AS (
        SELECT media_id, last_value(stored_event_at)
        FROM instagram_posts
        WHERE instagram_id IN {instagram_id}
          AND {DateFilter('post_created_at')}
          AND media_type != ''
        GROUP BY media_id
    ),
    media_types AS (
        SELECT arrayJoin(['REELS', 'CAROUSEL_ALBUM', 'IMAGE', 'VIDEO']) AS media_type
    ),
    metrics AS (
        SELECT
            mt.media_type,
            toInt32(COALESCE(total_posts, 0)) AS total_posts,
            toInt32(COALESCE(likes, 0))       AS likes,
            toInt32(COALESCE(comments, 0))    AS comments,
            toInt32(COALESCE(saved, 0))       AS saved,
            toInt32(COALESCE(engagement, 0))  AS engagement,
            toInt32(COALESCE(reach, 0))       AS reach,
            toInt32(COALESCE(views, 0))       AS views
        FROM media_types mt
        LEFT JOIN (
            SELECT
                media_type,
                COUNT(*) AS total_posts, SUM(like_count) AS likes,
                SUM(comments_count) AS comments, SUM(saved) AS saved,
                SUM(engagement) AS engagement, SUM(reach) AS reach,
                SUM(views) AS views
            FROM (
                SELECT
                    media_id,
                    last_value(media_type)        AS media_type,
                    last_value(like_count)        AS like_count,
                    last_value(comments_count)    AS comments_count,
                    last_value(saved)             AS saved,
                    last_value(engagement)        AS engagement,
                    last_value(reach)             AS reach,
                    last_value(views)             AS views,
                    post_created_at
                FROM instagram_posts
                WHERE (media_id, stored_event_at) IN posts
                GROUP BY media_id, post_created_at
            )
            GROUP BY media_type
        ) t ON mt.media_type = t.media_type
    )
SELECT * FROM (
    SELECT * FROM metrics
    UNION ALL
    SELECT 'TOTAL' AS media_type,
        SUM(total_posts) AS total_posts, SUM(likes) AS likes,
        SUM(comments) AS comments,       SUM(saved) AS saved,
        SUM(engagement) AS engagement,   SUM(reach) AS reach,
        sum(views) AS views
    FROM metrics
)
ORDER BY CASE WHEN media_type = 'TOTAL' THEN 1 ELSE 0 END, media_type
```

---

#### 5.2.7 Top Posts (`topPostQuery`)

**Table:** `instagram_posts`

```sql
WITH posts AS (
    SELECT media_id, max(stored_event_at)
    FROM instagram_posts
    WHERE instagram_id in {instagram_id}
      AND {DateFilter('post_created_at')}
      AND media_type in ('{media_type}')
      [AND has(hashtags, '{hashtag}')]    -- only if hashtags filter present
    GROUP BY media_id
)
SELECT * EXCEPT (taps_back, taps_forward), engagement as total_engagement
FROM instagram_posts
WHERE (media_id, stored_event_at) in posts
  AND entity_type != 'STORY'
ORDER BY {order_by} DESC
LIMIT {limit}
```

**Parameters:** `limit` default 15 (top posts page) or 5 (overview), `order_by` default `'engagement'`, `media_type` from request, `hashtags` optional.

---

#### 5.2.8 Active Users by Hour (`activeUsersHourQuery`)

**Route:** `POST overview/instagram/activeUsers`

**Table:** `instagram_insights`

```sql
SELECT
    max(buckets)     AS buckets,
    max(values)      as values,
    arrayMax(values) as highest_value,
    buckets[indexOf(values, highest_value)] as highest_hour
FROM (
    SELECT
        first_value(buckets) as buckets,
        arrayMap((x)->toInt32(x)/count(), sumForEach(values)) AS values
    FROM (
        SELECT
            arrayMap((x)->toInt32(x[1]), active_users_per_hour) as buckets,
            arrayMap((x)->toInt32(x[2]), active_users_per_hour) as values,
            arrayMax(values)                AS highest_value,
            toInt32(indexOf(values, highest_value)) AS highest_hour
        FROM (
            SELECT arrayMap((x)->(splitByChar('$', x)), online_followers) as active_users_per_hour
            FROM (
                SELECT
                    max(instagram_id)            AS insta_id,
                    max(online_followers)        AS online_followers,
                    max(online_users_datetime)   AS latest_record
                FROM instagram_insights
                WHERE instagram_id in {instagram_id}
                  AND toDateTime(online_users_datetime,0,'{timezone}') BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
                GROUP BY record_id
                ORDER BY latest_record DESC
            )
        )
    )
)
ORDER BY arrayMax(buckets) DESC
```

**Data format:** `online_followers` stores `"0$42$1$18$2$5..."` where odd positions are hour numbers and even positions are user counts.

**Controller timezone adjustment:** `round(Carbon::now($timezone)->offsetHours) + 8` (+8 accounts for UTC+8 storage), wraps modulo 24.

---

#### 5.2.9 Active Users by Day (`activeUsersDayQuery`)

**Table:** `instagram_insights`

```sql
SELECT
    groupArray(value)    AS values,
    groupArray(day)      AS buckets,
    arrayMax(values)     AS highest_value,
    buckets[indexOf(values, highest_value)] AS highest_day
FROM (
    SELECT
        max(toInt32(arraySum(arrayMap((x)->toInt32(x[2]),active_users_per_hour)))) AS value,
        day
    FROM (
        SELECT
            arrayMap((x)->(splitByChar('$', x)), online_followers) as active_users_per_hour,
            day
        FROM (
            SELECT
                max(day_of_week)         AS day,
                max(online_followers)    AS online_followers,
                record_id
            FROM instagram_insights
            WHERE instagram_id in {instagram_id}
              AND toDateTime(created_time,0,'{timezone}') BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
            GROUP BY record_id
        )
    )
    WHERE day != ''
    GROUP BY day
    ORDER BY
        CASE
            WHEN day='Monday'    THEN 1
            WHEN day='Tuesday'   THEN 2
            WHEN day='Wednesday' THEN 3
            WHEN day='Thursday'  THEN 4
            WHEN day='Friday'    THEN 5
            WHEN day='Saturday'  THEN 6
            WHEN day='Sunday'    THEN 7
        END
)
```

---

#### 5.2.10 Impressions (`impressionsQuery`)

**Route:** `POST overview/instagram/impressions`

**Table:** `instagram_posts`

```sql
WITH posts AS (
    SELECT media_id, max(stored_event_at)
    FROM instagram_posts
    WHERE instagram_id in {instagram_id}
      AND {DateFilter('post_created_at')}
    GROUP BY media_id
)
SELECT
    groupArray(dates)   as buckets,
    groupArray(impress) as impressions,
    toInt32(sum(impress)) as show_data
FROM (
    SELECT
        toInt32(SUM(impressions)) as impress,
        toDate(post_created_at) as dates
    FROM instagram_posts
    WHERE (media_id, stored_event_at) in posts
    GROUP BY media_id, post_created_at
    ORDER BY toDate(post_created_at) ASC
    WITH FILL FROM toDate('{currentStartDate}') TO toDate('{currentEndDate}') + 1
)
```

**Rollup (`impressionsRollupQuery`):**

```sql
WITH posts AS (
    SELECT media_id, max(stored_event_at)
    FROM instagram_posts
    WHERE instagram_id in {instagram_id}
      AND {DateFilter('post_created_at')}
    GROUP BY media_id
)
SELECT
    toInt32(sum(impressions))   AS total_impressions,
    ROUND(avg(impressions), 3)  AS avg_impressions
FROM (
    SELECT
        toInt32(SUM(impressions)) as impressions,
        toDate(post_created_at) as dates
    FROM instagram_posts
    WHERE (media_id, stored_event_at) in posts
    GROUP BY media_id, post_created_at
    ORDER BY toDate(post_created_at) ASC
)
```

---

#### 5.2.11 Reels (`reelsQuery`)

**Route:** `POST overview/instagram/reels`

**Table:** `instagram_posts` (filtered `media_type='REELS'`)

```sql
SELECT
    groupArray(created_at)       as buckets,
    groupArray(total_post)       as total_posts,
    groupArray(engagement)       as engagement,
    groupArray(likes)            as likes,
    groupArray(comments)         as comments,
    groupArray(saves)            as saves,
    groupArray(shares)           as shares,
    groupArray(avg_watch_time)   as avg_watch_time,
    groupArray(total_watch_time) as total_watch_time,
    toInt32(sum(total_post))     as show_data
FROM (
    SELECT
        created_at,
        toInt32(count())              as total_post,
        toInt32(sum(engagement))      as engagement,
        toInt32(sum(likes))           as likes,
        toInt32(sum(comments))        as comments,
        toInt32(sum(saves))           as saves,
        toInt32(sum(shares))          as shares,
        if(count() != 0, round(avg(avg_watch_time)/1000, 2), 0) as avg_watch_time,
        toInt32(sum(total_watch_time)/1000) as total_watch_time
    FROM (
        SELECT
            toDate(last_value(post_created_at))         as created_at,
            last_value(engagement)                      as engagement,
            last_value(like_count)                      as likes,
            last_value(comments_count)                  as comments,
            last_value(saved)                           as saves,
            last_value(shares)                          as shares,
            last_value(reels_avg_watch_time)            as avg_watch_time,
            last_value(reels_total_watch_time)          as total_watch_time
        FROM instagram_posts
        WHERE instagram_id in {instagram_id}
          AND {DateFilter('post_created_at')}
          AND media_type='REELS'
        GROUP BY media_id
    )
    GROUP BY created_at
    ORDER BY created_at ASC
    WITH FILL FROM toDate('{currentStartDate}') TO toDate('{currentEndDate}') + 1
)
```

Watch times stored in ms, divided by 1000 for seconds output.

**Reels Rollup (`reelsRollupQuery`):**

```sql
SELECT
    toInt32(sum(engagement))    as engagement,
    toInt32(sum(likes))         as likes,
    toInt32(sum(comments))      as comments,
    toInt32(sum(saves))         as saves,
    toInt32(count())            as total_posts,
    toInt32(sum(shares))        as shares,
    if(count() != 0, round(avg(avg_watch_time)/1000, 2), 0) as avg_watch_time,
    round(sum(total_watch_time)/1000, 2)                    as total_watch_time
FROM (
    SELECT
        last_value(engagement)             as engagement,
        last_value(like_count)             as likes,
        last_value(comments_count)         as comments,
        last_value(saved)                  as saves,
        last_value(shares)                 as shares,
        last_value(reels_avg_watch_time)   as avg_watch_time,
        last_value(reels_total_watch_time) as total_watch_time
    FROM instagram_posts
    WHERE instagram_id in {instagram_id}
      AND {DateFilter('post_created_at')}
      AND media_type='REELS'
    GROUP BY media_id
)
```

---

#### 5.2.12 Engagements (`engagementsQuery`)

**Route:** `POST overview/instagram/engagement`

**Table:** `instagram_posts`

```sql
WITH posts AS (
    SELECT media_id, max(stored_event_at)
    FROM instagram_posts
    WHERE instagram_id in {instagram_id}
      AND {DateFilter('post_created_at')}
    GROUP BY media_id
)
SELECT
    groupArray(dates)     AS buckets,
    groupArray(engage)    AS engagement,
    groupArray(comments)  AS comments,
    groupArray(reactions) AS reactions,
    groupArray(doc_count) as doc_count,
    toInt32(sum(engage))  as show_data
FROM (
    SELECT
        toInt32(count())              as doc_count,
        toInt32(SUM(engagement))      as engage,
        toInt32(SUM(comments_count))  as comments,
        toInt32(SUM(like_count))      as reactions,
        toDate(post_created_at)       as dates
    FROM instagram_posts
    WHERE (media_id, stored_event_at) in posts
    GROUP BY toDate(post_created_at)
    ORDER BY toDate(post_created_at) ASC
    WITH FILL FROM toDate('{currentStartDate}') TO toDate('{currentEndDate}') + 1
)
```

**Engagements Rollup (`engagementsRollupQuery`):**

```sql
WITH posts AS (
    SELECT media_id, max(stored_event_at)
    FROM instagram_posts
    WHERE instagram_id in {instagram_id}
      AND {DateFilter('post_created_at')}
    GROUP BY media_id
)
SELECT
    toInt32(sum(engage))    AS engagement,
    ROUND(avg(engage), 3)   AS avg_engagement,
    toInt32(sum(comments))  as comments,
    toInt32(sum(reactions)) as reactions,
    toInt32(sum(saved))     as saved,
    toInt32(sum(doc_count)) as count
FROM (
    SELECT
        toInt32(count())              as doc_count,
        toInt32(sum(saved))           as saved,
        toInt32(SUM(engagement))      as engage,
        toInt32(SUM(comments_count))  as comments,
        toInt32(SUM(like_count))      as reactions,
        toDate(MAX(post_created_at))  as dates
    FROM instagram_posts
    WHERE (media_id, stored_event_at) in posts
    GROUP BY media_id
    ORDER BY toDate(MAX(post_created_at)) ASC
    WITH FILL FROM toDate('{currentStartDate}') TO toDate('{currentEndDate}') + 1
)
```

---

#### 5.2.13 Hashtags (`hashtagsEngagedQuery`)

**Route:** `POST overview/instagram/hashtags`

**Table:** `instagram_posts`

Returns top 30 hashtags by engagement.

```sql
WITH posts_data AS (
    SELECT media_id, max(stored_event_at)
    FROM instagram_posts
    WHERE instagram_id in {instagram_id}
      AND {DateFilter('post_created_at')}
    GROUP BY media_id
)
SELECT
    groupArray(name)        as name,
    groupArray(engagement)  as engagement,
    groupArray(likes)       as likes,
    groupArray(comments)    as comments,
    groupArray(saved)       as saved,
    groupArray(posts)       as posts
FROM (
    SELECT
        hashtag as name,
        toInt32(sum(engagement)) AS engagement,
        toInt32(sum(likes))      as likes,
        toInt32(sum(comments))   as comments,
        toInt32(sum(saved))      as saved,
        toInt32(sum(counts))     as posts
    FROM (
        SELECT
            arrayJoin(hashtags)   AS hashtag,
            sum(engagement)       as engagement,
            sum(like_count)       as likes,
            sum(comments_count)   as comments,
            sum(saved)            as saved,
            uniq(media_id)        AS counts
        FROM instagram_posts
        WHERE (media_id, stored_event_at) in posts_data
        GROUP BY hashtags
        ORDER BY engagement DESC
    )
    GROUP BY name
    ORDER BY engagement DESC
    LIMIT 30
)
```

**Hashtags Rollup (`hashtagsRollupQuery`):**

```sql
WITH posts AS (
    SELECT media_id, max(stored_event_at)
    FROM instagram_posts
    WHERE instagram_id in {instagram_id}
      AND {DateFilter('post_created_at')}
    GROUP BY media_id
)
SELECT
    toInt32(sum(engagement))                            AS total_engagement,
    toInt32(sum(like_count))                            AS total_likes,
    toInt32(sum(comments_count))                        AS total_comments,
    toInt32(sum(saved))                                 AS total_saves,
    toInt32(count(DISTINCT arrayJoin(hashtags)))        AS total_unique_hashtags,
    toInt32(sum(length(hashtags)))                      AS total_hashtag_uses
FROM (
    SELECT
        last_value(engagement)      as engagement,
        last_value(like_count)      as like_count,
        last_value(comments_count)  as comments_count,
        last_value(saved)           as saved,
        last_value(hashtags)        as hashtags,
        stored_event_at,
        media_id
    FROM instagram_posts
    WHERE (media_id, stored_event_at) in posts
    GROUP BY media_id, stored_event_at
)
```

---

#### 5.2.14 Stories (`storiesQuery`)

**Route:** `POST overview/instagram/stories`

**Table:** `instagram_posts` (filtered `entity_type = 'STORY' OR media_type = 'STORY'`)

```sql
WITH posts AS (
    SELECT media_id, max(stored_event_at)
    FROM instagram_posts
    WHERE instagram_id in {instagram_id}
      AND {DateFilter('post_created_at')}
    GROUP BY media_id
)
SELECT
    groupArray(dates)              AS buckets,
    groupArray(avg_impress)        AS avg_story_impressions,
    groupArray(impress)            AS story_impressions,
    groupArray(reach)              AS story_reach,
    groupArray(reply)              AS story_reply,
    groupArray(exit)               AS story_exits,
    groupArray(tap_forward)        AS story_taps_forward,
    groupArray(tap_back)           AS story_taps_back,
    groupArray(published_stories)  AS published_stories,
    toInt32(sum(reach) + sum(reply) + sum(exit) + sum(tap_forward) + sum(tap_back)) as show_data
FROM (
    SELECT
        toDate(post_created_at) AS dates,
        toInt32(SUM(replies))   AS reply,
        toInt32(SUM(exits))     AS exit,
        toInt32(SUM(taps_forward)) AS tap_forward,
        toInt32(SUM(taps_back))    AS tap_back,
        toInt32(SUM(CASE WHEN entity_type = 'STORY' OR media_type = 'STORY' THEN 1 ELSE 0 END)) AS published_stories,
        CASE WHEN published_stories = 0 THEN 0 ELSE toInt32(SUM(reach)) END       AS reach,
        CASE WHEN published_stories = 0 THEN 0 ELSE toInt32(SUM(impressions)) END AS impress,
        CASE WHEN published_stories = 0 THEN 0 ELSE toInt32(AVG(impressions)) END AS avg_impress
    FROM instagram_posts
    WHERE (media_id, stored_event_at) in posts
      AND (entity_type = 'STORY' OR media_type = 'STORY')
    GROUP BY media_id, post_created_at
    ORDER BY toDate(post_created_at) ASC
    WITH FILL FROM toDate('{currentStartDate}') TO toDate('{currentEndDate}') + 1
)
```

**Stories Rollup (`storiesRollupQuery`):**

```sql
WITH posts AS (
    SELECT media_id, max(stored_event_at)
    FROM instagram_posts
    WHERE instagram_id in {instagram_id}
      AND {DateFilter('post_created_at')}
    GROUP BY media_id
)
SELECT
    toInt32(sum(impress))       AS story_impressions,
    toInt32(avg(impress))       AS avg_story_impressions,
    toInt32(sum(reach))         AS story_reach,
    toInt32(sum(reply))         AS story_reply,
    toInt32(sum(exit))          AS story_exits,
    toInt32(sum(tap_forward))   AS story_taps_forward,
    toInt32(sum(tap_back))      AS story_taps_back,
    toInt32(sum(published_stories)) AS published_stories
FROM (
    SELECT
        toInt32(SUM(replies))   AS reply,
        toInt32(SUM(exits))     AS exit,
        toInt32(SUM(taps_forward)) AS tap_forward,
        toInt32(SUM(taps_back))    AS tap_back,
        toInt32(SUM(CASE WHEN entity_type = 'STORY' OR media_type = 'STORY' THEN 1 ELSE 0 END)) AS published_stories,
        CASE WHEN published_stories = 0 THEN 0 ELSE toInt32(SUM(reach)) END       AS reach,
        CASE WHEN published_stories = 0 THEN 0 ELSE toInt32(SUM(impressions)) END AS impress
    FROM instagram_posts
    WHERE (media_id, stored_event_at) in posts
      AND (entity_type = 'STORY' OR media_type = 'STORY')
    GROUP BY media_id, post_created_at
    ORDER BY toDate(post_created_at) ASC
    WITH FILL FROM toDate('{currentStartDate}') TO toDate('{currentEndDate}') + 1
)
```

---

#### 5.2.15 Demographics

**Audience Age (`audienceAgeQuery`):**

**Table:** `instagram_insights` -- demographics stored as `$`-delimited pairs in `audience_age`.

```sql
SELECT
    max(CASE WHEN ages = '65+'   THEN ages_count[7] ELSE 0 END) AS `65+`,
    max(CASE WHEN ages = '55-64' THEN ages_count[6] ELSE 0 END) AS `55-64`,
    max(CASE WHEN ages = '45-54' THEN ages_count[5] ELSE 0 END) AS `45-54`,
    max(CASE WHEN ages = '35-44' THEN ages_count[4] ELSE 0 END) AS `35-44`,
    max(CASE WHEN ages = '25-34' THEN ages_count[3] ELSE 0 END) AS `25-34`,
    max(CASE WHEN ages = '18-24' THEN ages_count[2] ELSE 0 END) AS `18-34`,
    max(CASE WHEN ages = '13-17' THEN ages_count[1] ELSE 0 END) AS `13-17`
FROM (
    SELECT
        arrayMap((x)->(x[1]), age_pair)          AS ages,
        arrayMap((x)->toInt32(x[2]), age_pair)   AS ages_count
    FROM (
        SELECT arrayMap((x)->(splitByChar('$',x)), audience_age) as age_pair
        FROM (
            SELECT instagram_id,
                   max(audience_age)    AS audience_age,
                   max(created_time)    AS latest_record
            FROM instagram_insights
            WHERE instagram_id in {instagram_id}
              AND {DateFilter('created_time')}
            GROUP BY instagram_id
        )
    )
) ARRAY JOIN ages
```

Known bug: `'18-24'` bucket is aliased as `'18-34'` in output.

**Audience Gender (`audienceGenderQuery`):**

```sql
SELECT
    MAX(CASE WHEN gender = 'U' THEN gender_count[1] ELSE 0 END) AS U,
    MAX(CASE WHEN gender = 'M' THEN gender_count[2] ELSE 0 END) AS M,
    MAX(CASE WHEN gender = 'F' THEN gender_count[3] ELSE 0 END) AS F
FROM (
    SELECT
        arrayMap((x)-> (x[1]), gender_pair)          As gender,
        arrayMap((x)-> toInt32(x[2]), gender_pair)   As gender_count,
        followers_count
    FROM (
        SELECT arrayMap((x)->(splitByChar('$',x)), audience_gender) as gender_pair, followers_count
        FROM (
            SELECT instagram_id,
                   max(audience_gender)            AS audience_gender,
                   first_value(followers_count)    AS followers_count,
                   max(created_time)               as latest_record
            FROM instagram_insights
            WHERE instagram_id in {instagram_id}
              AND {DateFilter('created_time')}
            GROUP BY instagram_id
        )
    )
) ARRAY JOIN gender
```

**Audience Max Gender+Age (`audienceMaxQuery`):**

```sql
SELECT
    genders_age[max_index][1] AS gender,
    genders_age[max_index][2] AS age,
    max_value as value
FROM (
    SELECT
        arrayMap((x)->(splitByChar('.',x)), arrayMap((x)->(x[1]), gender_age_pair)) AS genders_age,
        arrayMax(arrayMap((x)->toInt32(x[2]), gender_age_pair))    AS max_value,
        indexOf(arrayMap((x)->toInt32(x[2]), gender_age_pair), max_value) as max_index
    FROM (
        SELECT arrayMap((x)->(splitByChar('$',x)), audience_gender_age) as gender_age_pair
        FROM (
            SELECT instagram_id,
                   max(audience_gender_age) AS audience_gender_age,
                   max(created_time)        AS latest_record
            FROM instagram_insights
            WHERE instagram_id in {instagram_id}
              AND {DateFilter('created_time')}
            GROUP BY instagram_id
        )
    )
)
```

**Cities (`citiesQuery`):**

```sql
SELECT *
FROM (
    SELECT
        arrayMap((x)->x[1], audience_city_pair)          AS `city.cities`,
        arrayMap((x)->toInt32(x[2]), audience_city_pair) AS `city.city_values`
    FROM (
        SELECT arrayMap((x)->(splitByChar('$',x)), audience_city) as audience_city_pair
        FROM (
            SELECT instagram_id,
                   max(audience_city)   AS audience_city,
                   max(created_time)    AS latest_record
            FROM instagram_insights
            WHERE instagram_id in {instagram_id}
              AND {DateFilter('created_time')}
            GROUP BY instagram_id
        )
    )
) ARRAY JOIN city
ORDER BY city.city_values DESC
```

**Countries (`countriesQuery`):**

```sql
SELECT *
FROM (
    SELECT
        arrayMap((x)->x[1], audience_country_pair)          AS `country.countries`,
        arrayMap((x)->toInt32(x[2]), audience_country_pair) AS `country.country_values`
    FROM (
        SELECT arrayMap((x)->(splitByChar('$',x)), audience_country) as audience_country_pair
        FROM (
            SELECT instagram_id,
                   max(audience_country) AS audience_country,
                   max(created_time)     AS latest_record
            FROM instagram_insights
            WHERE instagram_id in {instagram_id}
              AND {DateFilter('created_time')}
            GROUP BY instagram_id
        )
    )
) ARRAY JOIN country
ORDER BY country.country_values DESC
```

---

#### 5.2.16 Time Recommendation (`timeRecommendationQuery`)

**Table:** `instagram_posts`

```sql
SELECT
    instagram_id,
    day_of_week,
    hour_of_day,
    sum(impressions)  as post_impressions,
    sum(engagement)   as total_engagement
FROM instagram_posts
WHERE instagram_id in {instagram_id}
  AND {DateFilter('post_created_at')}
GROUP BY instagram_id, day_of_week, hour_of_day
ORDER BY day_of_week, hour_of_day
```

---

## 6. Instagram Competitor Analytics

### 6.1 Constructor and Timezone Handling

Same pattern as Facebook Competitor: dates parsed in user timezone, converted to UTC before embedding in SQL.

```
startDate = Carbon::parse(date[0], timezone)->setTime(0,0,1)->setTimezone('UTC')
endDate = Carbon::parse(date[1], timezone)->setTime(23,59,59)->setTimezone('UTC')
```

**Key difference from own Instagram analytics:** The request key is `date_filter` (not `date`).

### 6.2 Tables

| Table | Used for |
|-------|---------|
| `instagram_competitor_posts` | All post-level competitor metrics |
| `instagram_competitor_insights` | Competitor account-level metrics (followers, following, profile picture) |

### 6.3 Helper Methods

#### `getPageFilters($insightsFlag, $filterAdded)`
```sql
-- posts: business_account_id in('id1','id2',...)
-- insights: instagram_account_id in('id1','id2',...)
-- filterAdded=true: excludes accounts with state='Added'
```

#### `getDateFilters($insightsFlag, $dateColumn)`
```sql
-- posts: created_at BETWEEN toDateTime('{startDate}',0) AND toDateTime('{endDate}',0)
-- insights: inserted_at BETWEEN toDateTime('{startDate}',0) AND toDateTime('{endDate}',0)
```

#### `getConstantConditions($field, $idField)`
Builds `multiIf(id='x','val',...,'Random')` for name/image/state/slug.

### 6.4 Controller Methods

| Method | Description | Sort Default |
|--------|-------------|--------------|
| `postingActivityGraphByTypes` | Aggregate by media_type across all competitors | `avgTotalEngagements` |
| `postingActivityBySpecificType` | Per-competitor for one media_type | `followersCount` |
| `postingActivityTableByType` | Table view per-competitor for one media_type | `followersCount` |
| `followersGrowthComparison` | Per-competitor followers time-series | `total_followed_by_count` |
| `dataTableMetrics` | Current/previous period KPIs with growth % | N/A |
| `topAndLeastPerformingPosts` | Top-5/bottom-5 per competitor | engagement DESC/ASC |
| `topHashtags` | Top hashtags (limit default: 7) | count DESC |
| `individualHashtagData` | Per-competitor stats for one hashtag | engagement DESC |
| `biographyData` | Biography + length per competitor | `biography_length` |

### 6.5 SQL Queries

#### 6.5.1 `getFollowersGrowthComparisonQuery($sortOrder)`

**Table:** `instagram_competitor_insights`

```sql
SELECT
    multiIf(page_ids.business_account_id='{id1}','{name1}',...,'Random') as name,
    multiIf(...) as image,
    multiIf(...) as state,
    multiIf(...) as slug,
    *
FROM (
    SELECT c1 as business_account_id
    FROM VALUES (('id1'),('id2'),...)
) as page_ids
LEFT JOIN (
    SELECT
        instagram_account_id,
        groupArray(date)                     as dates,
        groupArray(total_following_count)    as total_following_count,
        groupArray(total_followed_by_count)  as total_followed_by_count,
        arrayZip(dates, total_following_count)  as dates_with_following_count,
        arrayZip(dates, total_followed_by_count) as dates_with_followed_by_count
    FROM (
        SELECT
            page_name, profile_picture_url, instagram_account_id,
            toDate(inserted_at) as date,
            total_following_count, total_followed_by_count
        FROM instagram_competitor_insights
        WHERE instagram_account_id in('id1','id2',...)
          AND inserted_at BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
        GROUP BY inserted_at, instagram_account_id, profile_picture_url, page_name,
                 total_following_count, total_followed_by_count
        ORDER BY inserted_at
    )
    GROUP BY instagram_account_id, page_name
) as page_insights ON page_ids.business_account_id = page_insights.instagram_account_id
ORDER BY {sortOrder} DESC
```

---

#### 6.5.2 `getPostingActivityGraphByTypesQuery($sortOrder)`

Aggregated across ALL competitors by media type.

**Table:** `instagram_competitor_posts`

```sql
WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM instagram_competitor_posts
    WHERE business_account_id in('id1','id2',...)
      AND created_at BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
    GROUP BY post_id
)
SELECT
    post_type as mediaType,
    media_product_type as mediaProductType,
    round(sum(avg_engagement), 2) as avgTotalEngagements,
    sum(page_post_count) as totalPosts,
    round(sum(er), 2) as avgEngagementRate,
    if(dateDiff('week',toDate('{start}'),toDate('{end}'))=0, 0,
       round(totalPosts/dateDiff('week',toDate('{start}'),toDate('{end}')),2)) as postsPerWeek,
    if(dateDiff('day',toDate('{start}'),toDate('{end}'))=0, 0,
       round(totalPosts/dateDiff('day',toDate('{start}'),toDate('{end}')),2)) as postsPerDay,
    if(dateDiff('hour',toDate('{start}'),toDate('{end}'))=0, 0,
       round(totalPosts/dateDiff('hour',toDate('{start}'),toDate('{end}')),2)) as postsPerHour,
    dateDiff('week',toDate('{start}'),toDate('{end}')) as weekCount,
    dateDiff('day',toDate('{start}'),toDate('{end}'))  as dayCount,
    dateDiff('hour',toDate('{start}'),toDate('{end}')) as hourCount
FROM (
    SELECT
        business_account_id, name, post_type, media_product_type,
        sum(count) as page_post_count,
        groupArray(total_engagement),
        if(page_post_count <= 0, 0, round(sum(total_engagement)/page_post_count,2)) as avg_engagement,
        if(page_post_count <= 0, 0,
           round(((sum(er)/page_post_count)/max(followers_count))*100,2)) as er
    FROM (
        SELECT
            business_account_id, name,
            if(media_product_type=='REELS','VIDEO REEL',
               if(media_type=='CAROUSEL_ALBUM','CAROUSEL ALBUM', media_type)) as post_type,
            media_product_type,
            count() as count,
            sum(engagement) as total_engagement,
            total_engagement as er,
            argMax(total_followed_by_count, created_at) as followers_count,
            toStartOfWeek(created_at) as week
        FROM instagram_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        GROUP BY business_account_id, name, post_type, media_product_type, week
        ORDER BY week ASC
        WITH FILL FROM toStartOfWeek(toDate('{start}')) TO toStartOfWeek(toDate('{end}')) STEP INTERVAL 1 WEEK
    )
    WHERE business_account_id != ''
    GROUP BY business_account_id, name, post_type, media_product_type
)
GROUP BY post_type, media_product_type
ORDER BY {sortOrder} DESC
```

**Media type mapping:** `REELS` -> `'VIDEO REEL'`, `CAROUSEL_ALBUM` -> `'CAROUSEL ALBUM'`, others as-is.

---

#### 6.5.3 `getPostingActivityBySpecificTypeQuery($mediaType, $mediaProductType, $sortOrder)`

Per-competitor breakdown for a specific media type. Same structure as above but groups by `business_account_id` and adds metadata from `instagram_competitor_insights`.

```sql
-- Key additions:
LEFT JOIN (
    SELECT
        argMax(profile_picture_url, inserted_at) as image,
        argMax(page_name, inserted_at) as name,
        argMax(total_following_count, inserted_at) as followingCount,
        argMax(total_followed_by_count, inserted_at) as followersCount,
        instagram_account_id
    FROM instagram_competitor_insights
    GROUP BY instagram_account_id
) as page_constants ON page_constants.instagram_account_id = pages_ids.business_account_id
```

---

#### 6.5.4 `getPostingActivityTableByTypeQuery($mediaType, $mediaProductType, $sortOrder)`

```sql
WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM instagram_competitor_posts
    WHERE business_account_id in('id1','id2',...)
      AND created_at BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
    GROUP BY post_id
)
SELECT
    pages_ids.business_account_id as businessAccountId,
    multiIf(...) as name, multiIf(...) as image, multiIf(...) as state, multiIf(...) as slug,
    coalesce(page_insights.count, 0) as count,
    coalesce(page_insights.total_engagement, 0) as totalEngagement,
    round(if(
        coalesce(count,0)<=0 OR coalesce(followersCount,0)<=0, 0,
        ((coalesce(total_engagement,0)/coalesce(count,1))/coalesce(followersCount,1))*100
    ), 2) as engagementRate,
    coalesce(page_metadata.followingCount, 0) as followingCount,
    coalesce(page_metadata.followersCount, 0) as followersCount,
    page_insights.media_type as mediaType,
    page_insights.media_product_type as mediaProductType
FROM (
    SELECT c1 as business_account_id FROM VALUES (('id1'),...)
) as pages_ids
LEFT JOIN (
    SELECT business_account_id,
           if(media_product_type=='REELS','VIDEO REEL',if(media_type=='CAROUSEL_ALBUM','CAROUSEL ALBUM',media_type)) as media_type,
           media_product_type,
           max(count) as count,
           max(total_engagement) as total_engagement
    FROM (
        SELECT business_account_id, media_type, media_product_type,
               count() as count, sum(engagement) as total_engagement
        FROM instagram_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        GROUP BY business_account_id, media_type, media_product_type
    ) as page_metrics
    WHERE media_type = '{mediaType}' AND media_product_type = '{mediaProductType}'
    GROUP BY business_account_id, media_type, media_product_type
) as page_insights USING business_account_id
LEFT JOIN (
    SELECT
        argMax(total_following_count, inserted_at) as followingCount,
        argMax(total_followed_by_count, inserted_at) as followersCount,
        instagram_account_id
    FROM instagram_competitor_insights
    GROUP BY instagram_account_id
) as page_metadata ON page_metadata.instagram_account_id = pages_ids.business_account_id
ORDER BY {sortOrder} DESC
```

---

#### 6.5.5 `getDataTableDataMetricsQuery($sortOrder)`

Same weekly-fill pattern as Facebook competitor. Uses `instagram_competitor_posts` and `instagram_competitor_insights` tables with `business_account_id`/`instagram_account_id` join keys.

**Returns per competitor:** `businessAccountId`, `name`, `image`, `state`, `slug`, `averageEngagement`, `averagePostsPerWeek`, `engagementRate`, `dayOfWeek`, `hourOfDay`, `averagePostsPerDay`, `followersCount`, `followingCount`.

---

#### 6.5.6 `getTopPerformingAndLeastPerformingPostsQuery()`

```sql
WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM instagram_competitor_posts
    WHERE business_account_id in(...) AND created_at BETWEEN ...
    GROUP BY post_id
)
SELECT * FROM (
    SELECT *, 'top_5_posts' as category
    FROM (
        SELECT * FROM instagram_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        ORDER BY engagement desc
        LIMIT 5 BY business_account_id
    )
UNION ALL
    SELECT *, 'least_5_posts' as category
    FROM (
        SELECT * FROM instagram_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        ORDER BY engagement asc
        LIMIT 5 BY business_account_id
    )
)
```

---

#### 6.5.7 `getTopHashtagsQuery($limit)`

Same pattern as Facebook competitor hashtags but uses `instagram_competitor_posts` and `instagram_competitor_insights`. Uses `arrayJoin(hashtags)` to explode the hashtag array. Engagement rate formula: `(total_engagement / total_followers / count) * 100`.

---

#### 6.5.8 `getBiographyQuery($sortOrder)`

Uses `instagram_competitor_insights` (not posts) for biography data:

```sql
SELECT * FROM (
    select last_value(biography) as biography,
           lengthUTF8(biography) as biography_length,
           instagram_account_id as business_account_id,
           multiIf(...) as state,
           multiIf(...) as slug
    from instagram_competitor_insights
    group by instagram_account_id
    ORDER BY max(inserted_at) DESC
) as biography_statistics
LEFT JOIN (
    select argMax(profile_picture_url, inserted_at) as image,
           argMax(page_name, inserted_at) as name,
           argMax(total_followed_by_count, inserted_at) as followersCount,
           instagram_account_id
    from instagram_competitor_insights
    group by instagram_account_id
) as page_constants ON page_constants.instagram_account_id = biography_statistics.business_account_id
ORDER BY {sortOrder} DESC
```

---

## 7. LinkedIn Analytics

### 7.1 Route Map

| Route | Method |
|-------|--------|
| `POST overview/linkedin/summary` | `getOverviewSummary` |
| `POST overview/linkedin/audienceGrowth` | `getOverviewAudienceGrowth` |
| `POST overview/linkedin/pageViews` | `getOverviewPageViews` |
| `POST overview/linkedin/publishingBehaviour` | `getOverviewPublishingBehaviour` |
| `POST overview/linkedin/topPosts` | `getOverviewTopPosts` |
| `POST overview/linkedin/postsPerDays` | `getOverviewDailyPosts` |
| `POST overview/linkedin/hashtags` | `getOverviewHashtags` |
| `POST overview/linkedin/getTopPosts` | `TopPosts15` |
| `POST overview/linkedin/followersDemographics` | `getOverviewfollowersDemographics` |

### 7.2 Summary (`summaryQuery`)

**Tables:** `linkedin_posts`, `linkedin_insights`

```sql
WITH posts AS (
    SELECT
        post_id,
        max(saving_time)
    FROM linkedin_posts
    WHERE linkedin_id IN ('...')
      AND toDateTime(published_at, 0, '{timezone}') BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
    GROUP BY post_id
)
SELECT
    {liId}
    toInt32(coalesce(post.post_comments, 0))        AS post_comments,
    toInt32(coalesce(post.post_favorites, 0))        AS post_likes,
    toInt32(coalesce(post.total_engagement, 0))      AS total_engagement,
    toInt32(coalesce(post.total_posts, 0))           AS total_posts,
    toInt32(coalesce(post.post_shares, 0))           AS post_shares,
    toInt32(coalesce(post.post_clicks, 0))           AS post_clicks,
    toInt32(coalesce(insights.total_follower_count, 0)) AS followers,
    toInt32(coalesce(insights.page_views, 0))           AS page_views,
    toInt32(coalesce(insights.page_reach, 0))           AS page_reach,
    toInt32(coalesce(insights.page_shares, 0))          AS page_shares,
    toInt32(coalesce(insights.page_comments, 0))        AS page_comments,
    toInt32(coalesce(insights.page_reactions, 0))       AS page_reactions,
    toInt32(coalesce(insights.total_impressions, 0))    AS page_impressions,
    toInt32(coalesce(insights.unique_visitors, 0))      AS page_unique_visitors,
    round(coalesce(insights.engagement_rate, 0), 2)     AS engagement_rate,
    round(coalesce(post.post_engagement_rate, 0), 2)    AS post_engagement_rate
FROM
(
    SELECT toString(c1) AS linkedin_id
    FROM VALUES ('id1','id2',...)
) AS page_ids
LEFT JOIN
(
    SELECT
        linkedin_id,
        post_comments, post_favorites, post_clicks, post_shares,
        post_comments + post_favorites + post_shares AS total_engagement,
        post_engagement_rate, total_posts
    FROM
    (
        SELECT
            linkedin_id,
            SUM(comments)                           AS post_comments,
            SUM(favorites)                          AS post_favorites,
            SUM(post_clicks)                        AS post_clicks,
            SUM(repost)                             AS post_shares,
            sum(total_engagement * impressions) / nullIf(sum(impressions), 0) AS post_engagement_rate,
            COUNT(*)                                AS total_posts
        FROM linkedin_posts
        WHERE (post_id, saving_time) IN posts
        GROUP BY linkedin_id
    )
) AS post USING linkedin_id
LEFT JOIN
(
    SELECT
        linkedin_id AS platform_id,
        argMax(latest_follower_count, latest_date)      AS total_follower_count,
        sum(daily_page_views)                           AS page_views,
        sum(daily_reach)                                AS page_reach,
        sum(daily_repost)                               AS page_shares,
        sum(daily_comments)                             AS page_comments,
        sum(daily_reactions)                            AS page_reactions,
        sum(daily_impressions)                          AS total_impressions,
        sum(daily_engagement_impressions) / nullIf(sum(daily_impressions), 0) AS engagement_rate,
        sum(daily_unique_visitors)                      AS unique_visitors
    FROM (
        SELECT
            linkedin_id,
            toDate(created_at)                          AS latest_date,
            argMin(totalFollowerCount, inserted_at)     AS latest_follower_count,
            sum(page_views)                             AS daily_page_views,
            sum(reach)                                  AS daily_reach,
            sum(repost)                                 AS daily_repost,
            sum(comments)                               AS daily_comments,
            sum(reactions)                              AS daily_reactions,
            sum(impressionCount)                        AS daily_impressions,
            sum(engagement * impressionCount)           AS daily_engagement_impressions,
            sum(unique_visitors)                        AS daily_unique_visitors
        FROM linkedin_insights
        WHERE linkedin_id IN ('...')
          AND toDateTime(created_at, 0, '{timezone}') BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
        GROUP BY linkedin_id, toDate(created_at)
    )
    GROUP BY linkedin_id
) AS insights ON page_ids.linkedin_id = insights.platform_id
{group}
```

**Key business logic:**
- `total_engagement = comments + favorites + shares` (clicks excluded)
- Engagement rate is impression-weighted: `sum(engagement * impressionCount) / sum(impressionCount)`
- Follower count uses `argMax(latest_follower_count, latest_date)` from insights

---

### 7.3 Last Follower Counts (Fallback) (`GetLastFollowerCounts`)

**Table:** `linkedin_insights`

```sql
SELECT
    arrayFirst(x -> x != 0, groupArray(totalFollowerCount))   AS total_follower_count,
    arrayFirst(x -> x != 0, groupArray(organicFollowerCount)) AS organic_follower_count,
    arrayFirst(x -> x != 0, groupArray(paidFollowerCount))    AS paid_follower_count
FROM
(
    SELECT
        argMin(created_at, inserted_at)                   AS inserted_time,
        toInt32(argMin(totalFollowerCount, inserted_at))  AS totalFollowerCount,
        toInt32(argMin(organicFollowerCount, inserted_at)) AS organicFollowerCount,
        toInt32(argMin(paidFollowerCount, inserted_at))   AS paidFollowerCount
    FROM
        linkedin_insights
    WHERE linkedin_id IN ('...')
      AND toDateTime(created_at, 0, '{timezone}') BETWEEN toDateTime('{start}',0) AND toDateTime('{end}',0)
    GROUP BY record_id
    ORDER BY inserted_time DESC
)
```

---

### 7.4 Audience Growth (`audienceQuery`)

**Table:** `linkedin_insights`

```sql
SELECT
    notEmpty(total_follower_count_temp) AS show_data,
    arrayFill(x -> not x == 0, organic_follower_count_temp) AS organic_follower_count,
    arrayMap(x -> toInt32(x),
        arrayPushFront(
            arrayDifference(arrayFill(x -> not x == 0, organic_follower_count_temp)),
            toInt32(0)
        )
    ) AS organic_followers_daily,
    arrayFill(x -> not x == 0, paid_follower_count_temp) AS paid_follower_count,
    arrayMap(x -> toInt32(x),
        arrayPushFront(
            arrayDifference(arrayFill(x -> not x == 0, paid_follower_count_temp)),
            toInt32(0)
        )
    ) AS paid_followers_daily,
    arrayFill(x -> not x == 0, total_follower_count_temp) AS total_follower_count,
    arrayMap(x -> toInt32(x),
        arrayPushFront(
            arrayDifference(arrayFill(x -> not x == 0, total_follower_count_temp)),
            toInt32(0)
        )
    ) AS total_followers_daily,
    buckets
FROM
(
    SELECT
        groupArray(dates)               AS buckets,
        groupArray(organicFollowerCount) AS organic_follower_count_temp,
        groupArray(paidFollowerCount)    AS paid_follower_count_temp,
        groupArray(totalFollowerCount)   AS total_follower_count_temp
    FROM
    (
        SELECT
            toDate(created_at)                              AS dates,
            toInt32(argMin(organicFollowerCount, inserted_at)) AS organicFollowerCount,
            toInt32(argMin(paidFollowerCount, inserted_at))    AS paidFollowerCount,
            toInt32(argMin(totalFollowerCount, inserted_at))   AS totalFollowerCount
        FROM linkedin_insights
        WHERE linkedin_id IN ('...')
            AND toDateTime(created_at, 0, '{timezone}') BETWEEN ...
        GROUP BY dates
        ORDER BY dates ASC
        WITH FILL TO toDate('{currentEndDate}') STEP 1
    )
)
```

**Audience Rollup (`audienceQueryRollup`):**

```sql
SELECT
    toInt32(last_value(organicFollowerCount)) AS organic_follower_count,
    toInt32(last_value(paidFollowerCount))    AS paid_follower_count,
    toInt32(last_value(totalFollowerCount))   AS total_follower_count,
    round(AVG(totalFollowerCount), 2)          AS avg_follower_count
FROM
(
    SELECT
        argMin(organicFollowerCount, inserted_at) AS organicFollowerCount,
        argMin(paidFollowerCount, inserted_at)    AS paidFollowerCount,
        argMin(totalFollowerCount, inserted_at)   AS totalFollowerCount,
        toDate(created_at) AS dates
    FROM linkedin_insights
    WHERE linkedin_id IN ('...')
      AND toDateTime(created_at, 0, '{timezone}') BETWEEN ...
    GROUP BY dates
    ORDER BY dates ASC
)
```

---

### 7.5 Page Views (`pageViewsQuery`)

**Table:** `linkedin_insights`

```sql
SELECT
    arrayMap(x -> toInt32(x), arrayCumSum(desktop_views_daily))  AS desktop_page_views,
    arrayMap(x -> toInt32(x), arrayCumSum(mobile_views_daily))   AS mobile_page_views,
    arrayMap(x -> toInt32(x), arrayCumSum(total_views_daily))    AS total_page_views,
    arrayMap(x -> toInt32(x), desktop_views_daily)               AS desktop_page_views_daily,
    arrayMap(x -> toInt32(x), mobile_views_daily)                AS mobile_page_views_daily,
    arrayMap(x -> toInt32(x), total_views_daily)                 AS total_page_views_daily,
    toInt32(arraySum(total_views_daily)) AS show_data,
    buckets
FROM
(
    SELECT
        groupArray(dates)          AS buckets,
        groupArray(desktop_views)  AS desktop_views_daily,
        groupArray(mobile_views)   AS mobile_views_daily,
        groupArray(total_views)    AS total_views_daily
    FROM
    (
        SELECT
            toDate(created_at)                    AS dates,
            toInt32(SUM(desktop_page_views))      AS desktop_views,
            toInt32(SUM(mobile_page_views))       AS mobile_views,
            toInt32(SUM(page_views))              AS total_views
        FROM linkedin_insights
        WHERE linkedin_id IN ('...')
          AND toDateTime(created_at, 0, '{timezone}') BETWEEN ...
        GROUP BY dates
        ORDER BY dates ASC
        WITH FILL TO toDate('{currentEndDate}') STEP 1
    )
)
```

Returns both cumulative (`arrayCumSum`) and daily arrays. No forward-fill applied (zeros stay as zeros).

**Page Views Rollup (`pageViewsRollupQuery`):**

```sql
SELECT
    toInt32(SUM(total_views))    AS total_page_views,
    toInt32(SUM(desktop_views))  AS desktop_page_views,
    toInt32(SUM(mobile_views))   AS mobile_page_views,
    round(AVG(total_views), 2)   AS avg_page_views
FROM
(
    SELECT
        toInt32(SUM(page_views))          AS total_views,
        toInt32(SUM(desktop_page_views))  AS desktop_views,
        toInt32(SUM(mobile_page_views))   AS mobile_views
    FROM linkedin_insights
    WHERE linkedin_id IN ('...')
      AND toDateTime(created_at, 0, '{timezone}') BETWEEN ...
    GROUP BY toDate(created_at)
    ORDER BY toDate(created_at) ASC
)
```

---

### 7.6 Engagements (`engagementsQuery`)

**Table:** `linkedin_posts`

**Notable side-effect:** This method mutates `currentEndDate` via `subDay()` before building the query.

```sql
WITH posts AS (
    SELECT post_id, max(saving_time)
    FROM linkedin_posts
    WHERE linkedin_id IN ('...')
      AND toDateTime(published_at, 0, '{timezone}') BETWEEN ...
    GROUP BY post_id
)
SELECT
    groupArray(dates)           AS buckets,
    groupArray(comment)         AS comments,
    groupArray(favorite)        AS favorites,
    groupArray(totalEngagement) AS total_engagement,
    groupArray(count)           AS doc_count,
    toInt32(sum(totalEngagement)) AS show_data
FROM
(
    SELECT
        toInt32(count())              AS count,
        toInt32(SUM(comments))        AS comment,
        toInt32(SUM(favorites))       AS favorite,
        toInt32(SUM(total_engagement)) AS totalEngagement,
        toDate(created_at)            AS dates
    FROM linkedin_posts
    WHERE (post_id, saving_time) IN posts
    GROUP BY toDate(created_at)
    ORDER BY toDate(created_at) ASC
    WITH FILL TO toDate('{currentEndDate - 1 day}')
)
```

**Engagements Rollup (`engagementsRollupQuery`):**

```sql
WITH posts AS (
    SELECT post_id, max(saving_time)
    FROM linkedin_posts
    WHERE linkedin_id IN ('...')
      AND {DateFilter('published_at')}
    GROUP BY post_id
)
SELECT
    toInt32(SUM(comment))               AS comments,
    toInt32(SUM(favorite))              AS favorites,
    toInt32(SUM(repost))                AS shares,
    toInt32(favorites + comments + shares) AS total_engagement,
    avg(totalEngagement)                AS avg_engagement,
    toInt32(sum(count))                 AS doc_count
FROM (
    SELECT
        count()                       AS count,
        SUM(comments)                 AS comment,
        SUM(favorites)                AS favorite,
        SUM(repost)                   AS repost,
        SUM(total_engagement)         AS totalEngagement
    FROM linkedin_posts
    WHERE (post_id, saving_time) IN posts
    GROUP BY created_at
)
```

---

### 7.7 Publishing Behaviour (`publishingBehaviourQuery`)

**Table:** `linkedin_posts`

```sql
WITH posts AS (
    SELECT post_id, last_value(saving_time) AS saving_time
    FROM linkedin_posts
    WHERE linkedin_id IN ('...')
      AND {DateFilter('published_at')}
      AND media_type IN ('text','images','videos','carousel','link')
    GROUP BY post_id
)
SELECT
    groupArray(likes)          AS likes,
    groupArray(comments)       AS comments,
    groupArray(shares)         AS shares,
    groupArray(clicks)         AS clicks,
    groupArray(engagement_rate) AS engagement_rate,
    groupArray(impression)     AS impressions,
    groupArray(total_posts)    AS total_posts,
    groupArray(engagements)    AS engagement,
    groupArray(reach)          AS reach,
    groupArray(created_at)     AS buckets
FROM
(
    SELECT
        toInt32(SUM(like_count))        AS likes,
        toInt32(SUM(comments_count))    AS comments,
        toInt32(SUM(shares_count))      AS shares,
        toInt32(SUM(clicks_count))      AS clicks,
        toInt32(SUM(impressions))       AS impression,
        toInt32(COUNT(*))               AS total_posts,
        toInt32(likes + comments + shares) AS engagements,
        if(SUM(impressions) > 0,
            toFloat32(round(100 * engagements / impression, 2)),
            0
        ) AS engagement_rate,
        toInt32(SUM(reach))             AS reach,
        toDate(created_at)              AS created_at
    FROM
    (
        SELECT
            post_id,
            last_value(favorites)   AS like_count,
            last_value(comments)    AS comments_count,
            last_value(repost)      AS shares_count,
            last_value(post_clicks) AS clicks_count,
            last_value(impressions) AS impressions,
            last_value(reach)       AS reach,
            created_at
        FROM linkedin_posts
        WHERE (post_id, saving_time) IN posts
        GROUP BY post_id, created_at
    )
    GROUP BY created_at
    ORDER BY created_at ASC
    WITH FILL TO toDate('{currentEndDate}')
)
```

`engagement_rate = 100 * (likes + comments + shares) / impressions`

**Publishing Behaviour Rollup (`publishingBehaviourRollupQuery`):**

Uses virtual `media_types` table to ensure all 5 types always appear. UNION ALL adds a `'total'` row.

```sql
WITH
posts AS (
    SELECT post_id, max(saving_time)
    FROM linkedin_posts
    WHERE linkedin_id IN ('...')
      AND {DateFilter('published_at')}
      AND media_type != ''
    GROUP BY post_id
),
media_types AS (
    SELECT arrayJoin(['text', 'images', 'videos', 'link', 'carousel']) AS media_type
),
metrics AS (
    SELECT
        mt.media_type,
        toInt32(COALESCE(total_posts, 0))  AS total_posts,
        toInt32(COALESCE(likes, 0))        AS likes,
        toInt32(COALESCE(comments, 0))     AS comments,
        toInt32(COALESCE(shares, 0))       AS shares,
        toInt32(COALESCE(clicks, 0))       AS clicks,
        toInt32(COALESCE(engagements, 0))  AS engagements,
        toInt32(COALESCE(impressions, 0))  AS impressions,
        toInt32(COALESCE(reach, 0))        AS reach
    FROM media_types mt
    LEFT JOIN
    (
        SELECT
            media_type,
            toInt32(COUNT(*))        AS total_posts,
            toInt32(SUM(like_count)) AS likes,
            toInt32(SUM(comments_count)) AS comments,
            toInt32(SUM(shares_count)) AS shares,
            toInt32(SUM(clicks_count)) AS clicks,
            toInt32(likes + comments + shares) AS engagements,
            toInt32(SUM(impressions)) AS impressions,
            toInt32(SUM(reach))      AS reach
        FROM
        (
            SELECT
                post_id,
                last_value(media_type)      AS media_type,
                last_value(favorites)       AS like_count,
                last_value(comments)        AS comments_count,
                last_value(repost)          AS shares_count,
                last_value(post_clicks)     AS clicks_count,
                last_value(total_engagement) AS engagement,
                last_value(impressions)     AS impressions,
                last_value(reach)           AS reach,
                created_at
            FROM linkedin_posts
            WHERE (post_id, saving_time) IN posts
            GROUP BY post_id, created_at
        )
        GROUP BY media_type
    ) t ON mt.media_type = t.media_type
)
SELECT * FROM (
    SELECT * FROM metrics
    UNION ALL
    SELECT
        'total'                     AS media_type,
        toInt32(SUM(total_posts))   AS total_posts,
        toInt32(SUM(likes))         AS likes,
        toInt32(SUM(comments))      AS comments,
        toInt32(SUM(shares))        AS shares,
        toInt32(SUM(clicks))        AS clicks,
        toInt32(SUM(engagements))   AS engagements,
        toInt32(SUM(impressions))   AS impressions,
        toInt32(SUM(reach))         AS reach
    FROM metrics
)
ORDER BY CASE WHEN media_type = 'total' THEN 1 ELSE 0 END, media_type
```

---

### 7.8 Top Posts (`topPostQuery`)

**Table:** `linkedin_posts`

```sql
WITH posts AS (
    SELECT post_id, max(saving_time)
    FROM linkedin_posts
    WHERE linkedin_id IN ('...')
      AND {DateFilter('published_at')}
      [AND has(hashtags, 'tag1')]
    GROUP BY post_id
)
SELECT *
FROM linkedin_posts
WHERE (post_id, saving_time) IN posts
ORDER BY {order_by} DESC
LIMIT {limit}
```

Parameters: `limit` (3 for overview, 15 for full list), `order_by` (default `'total_engagement'`), `hashtags` (optional filter -- note: multi-hashtag filter appears buggy).

---

### 7.9 Posts Per Day (`postsPerDayQuery`)

**Table:** `linkedin_posts`

```sql
WITH posts AS (
    SELECT post_id, max(saving_time)
    FROM linkedin_posts
    WHERE linkedin_id IN ('...')
      AND {DateFilter('published_at')}
    GROUP BY post_id
)
SELECT
    toInt32(countIf(day_of_week == 'Monday'))    AS Monday,
    toInt32(countIf(day_of_week == 'Tuesday'))   AS Tuesday,
    toInt32(countIf(day_of_week == 'Wednesday')) AS Wednesday,
    toInt32(countIf(day_of_week == 'Thursday'))  AS Thursday,
    toInt32(countIf(day_of_week == 'Friday'))    AS Friday,
    toInt32(countIf(day_of_week == 'Saturday'))  AS Saturday,
    toInt32(countIf(day_of_week == 'Sunday'))    AS Sunday
FROM linkedin_posts
WHERE (post_id, saving_time) IN posts
```

---

### 7.10 Top Hashtags (`topHashtagsQuery`)

**Table:** `linkedin_posts`

Uses `arrayJoin(lp.hashtags)` with INNER JOIN on deduplicated posts. Top 30 by engagement.

```sql
WITH posts AS (
    SELECT post_id, max(saving_time) AS max_saving_time
    FROM linkedin_posts
    WHERE linkedin_id IN ('...')
      AND {DateFilter('published_at')}
    GROUP BY post_id
)
SELECT
    groupArray(name)        AS name,
    groupArray(engagements) AS engagements,
    groupArray(likes)       AS likes,
    groupArray(comments)    AS comments,
    groupArray(shares)      AS shares,
    groupArray(posts)       AS posts
FROM
(
    SELECT
        arrayJoin(lp.hashtags)                 AS name,
        toInt32(SUM(lp.favorites))             AS likes,
        toInt32(SUM(lp.comments))              AS comments,
        toInt32(SUM(lp.repost))                AS shares,
        toInt32(likes + comments + shares)     AS engagements,
        toInt32(count(DISTINCT lp.post_id))    AS posts
    FROM linkedin_posts AS lp
    INNER JOIN posts AS p
        ON lp.post_id = p.post_id AND lp.saving_time = p.max_saving_time
    GROUP BY name
    ORDER BY engagements DESC
    LIMIT 30
)
```

**Hashtags Rollup (`topHashtagsQueryRollup`):**

```sql
WITH posts AS (
    SELECT post_id, max(saving_time) AS max_saving_time
    FROM linkedin_posts
    WHERE linkedin_id IN ('...')
      AND {DateFilter('published_at')}
    GROUP BY post_id
),
post_data AS (
    SELECT lp.hashtags, lp.favorites, lp.comments, lp.repost, lp.impressions, lp.reach
    FROM linkedin_posts AS lp
    INNER JOIN posts AS p
        ON lp.post_id = p.post_id AND lp.saving_time = p.max_saving_time
    WHERE length(lp.hashtags) > 0
)
SELECT
    toInt32((SELECT COUNT(DISTINCT arrayJoin(hashtags)) FROM post_data)) AS total_hashtags,
    toInt32(SUM(length(hashtags)))   AS total_times_used,
    toInt32(SUM(favorites))          AS total_likes,
    toInt32(SUM(comments))           AS total_comments,
    toInt32(SUM(repost))             AS total_shares,
    toInt32(total_likes + total_comments + total_shares) AS total_engagement,
    toInt32(SUM(impressions))        AS total_impressions,
    toInt32(SUM(reach))              AS total_reach
FROM post_data
```

---

### 7.11 Followers Demographics (`followersDemographicQuery`)

**Table:** `contentstudiobackend.linkedin_insights` (fully qualified table name)

LinkedIn stores demographics as JSON. Uses `JSONExtractKeysAndValues(column, 'String')` to parse. Categories: seniority, industry, country, city.

```sql
WITH
latest_record AS (
    SELECT
        followers_by_seniority, followers_by_industry,
        followers_by_country, followers_by_city,
        totalFollowerCount
    FROM contentstudiobackend.linkedin_insights
    WHERE linkedin_id IN ('...')
      AND toDateTime(created_at, 0, '{timezone}') BETWEEN ...
    ORDER BY created_at DESC
    LIMIT 1
),
processed_data AS (
    SELECT
        arrayMap(
            x -> (x.1, toInt32(JSONExtractInt(followers_by_seniority, x.1))),
            JSONExtractKeysAndValues(followers_by_seniority, 'String')
        ) AS seniority_data,
        arrayMap(
            x -> (x.1, toInt32(JSONExtractInt(followers_by_industry, x.1))),
            JSONExtractKeysAndValues(followers_by_industry, 'String')
        ) AS industry_data,
        arrayMap(
            x -> (x.1, toInt32(JSONExtractInt(followers_by_country, x.1))),
            JSONExtractKeysAndValues(followers_by_country, 'String')
        ) AS country_data,
        arrayMap(
            x -> (x.1, toInt32(JSONExtractInt(followers_by_city, x.1))),
            JSONExtractKeysAndValues(followers_by_city, 'String')
        ) AS city_data,
        totalFollowerCount
    FROM latest_record
)
SELECT
    map(
        'buckets',
        arrayConcat(
            IF(arraySum(arrayMap(x -> x.2, seniority_data)) < toInt32(totalFollowerCount),
               ['Others'], []),
            arrayReverse(arrayMap(x -> x.1, seniority_data))
        ),
        'values',
        arrayConcat(
            IF(arraySum(arrayMap(x -> x.2, seniority_data)) < toInt32(totalFollowerCount),
               [toString(toInt32(totalFollowerCount) - arraySum(arrayMap(x -> x.2, seniority_data)))],
               []),
            arrayReverse(arrayMap(x -> toString(x.2), seniority_data))
        )
    ) AS seniority,
    map(
        'buckets',
        arrayConcat(
            IF(arraySum(arrayMap(x -> x.2, industry_data)) < toInt32(totalFollowerCount),
               ['Others'], []),
            arrayReverse(arrayMap(x -> x.1, industry_data))
        ),
        'values',
        arrayConcat(
            IF(arraySum(arrayMap(x -> x.2, industry_data)) < toInt32(totalFollowerCount),
               [toString(toInt32(totalFollowerCount) - arraySum(arrayMap(x -> x.2, industry_data)))], []),
            arrayReverse(arrayMap(x -> toString(x.2), industry_data))
        )
    ) AS industry,
    map(
        'buckets',
        arrayConcat(
            IF(arraySum(arrayMap(x -> x.2, country_data)) < toInt32(totalFollowerCount),
               ['Others'], []),
            arrayReverse(arrayMap(x -> x.1, country_data))
        ),
        'values',
        arrayConcat(
            IF(arraySum(arrayMap(x -> x.2, country_data)) < toInt32(totalFollowerCount),
               [toString(toInt32(totalFollowerCount) - arraySum(arrayMap(x -> x.2, country_data)))], []),
            arrayReverse(arrayMap(x -> toString(x.2), country_data))
        )
    ) AS country,
    map(
        'buckets',
        arrayConcat(
            IF(arraySum(arrayMap(x -> x.2, city_data)) < toInt32(totalFollowerCount),
               ['Others'], []),
            arrayReverse(arrayMap(x -> x.1, city_data))
        ),
        'values',
        arrayConcat(
            IF(arraySum(arrayMap(x -> x.2, city_data)) < toInt32(totalFollowerCount),
               [toString(toInt32(totalFollowerCount) - arraySum(arrayMap(x -> x.2, city_data)))], []),
            arrayReverse(arrayMap(x -> toString(x.2), city_data))
        )
    ) AS city
FROM processed_data
```

Each category includes an "Others" entry if sum of known categories is less than totalFollowerCount.

---

### 7.12 Time Recommendation (`timeRecommendationQuery`)

**Table:** `linkedin_posts`

```sql
SELECT
    max(linkedin_id)         AS page_id,
    day_of_week,
    hour_of_day,
    0                        AS post_impressions,
    sum(total_engagement)    AS total_engagement
FROM linkedin_posts
WHERE linkedin_id IN ('...')
  AND toDateTime(published_at, 0, '{timezone}') BETWEEN ...
GROUP BY day_of_week, hour_of_day
ORDER BY day_of_week, hour_of_day
```

Note: `post_impressions` is hardcoded 0 (LinkedIn does not expose per-slot impression data).

---

## 8. YouTube Analytics

### 8.1 Route Map

| Route | Method |
|-------|--------|
| `POST overview/youtube/overviewSummary` | `overviewSummary` |
| `POST overview/youtube/overviewSubscriberTrend` | `overviewSubscriberTrend` |
| `POST overview/youtube/overviewEngagementTrend` | `overviewEngagementTrend` |
| `POST overview/youtube/overviewViewsTrend` | `overviewViewsTrend` |
| `POST overview/youtube/overviewWatchTimeTrend` | `overviewWatchTimeTrend` |
| `POST overview/youtube/overviewFindVideo` | `overviewFindVideo` |
| `POST overview/youtube/overviewVideoSharing` | `overviewVideoSharing` |
| `POST overview/youtube/overviewTopPosts` | `overviewTopPosts` |
| `POST overview/youtube/overviewLeastPosts` | `overviewLeastPosts` |
| `POST overview/youtube/overviewPerformanceAndVideoPostingSchedule` | `overviewPerformanceAndVideoPostingSchedule` |
| `POST overview/youtube/getSortedTopPosts` | `getSortedTopPosts` |

### 8.2 Summary (`summaryQuery`)

Complex 4-way LEFT JOIN across `youtube_activity_insights`, `youtube_channels`, `youtube_videos`, and views. Returns `'N/A'` string when metrics have no data. Subscriber count falls back to latest available if no data exists in the date range.

```sql
SELECT
    channel_id,
    if(count_stats > 0, toString(watch_time), 'N/A') as watch_time,
    if(count_stats > 0, toString(avg_view_duration), 'N/A') as avg_view_duration,
    if(count_stats > 0, toString(like), 'N/A') as like,
    if(count_stats > 0, toString(dislike), 'N/A') as dislike,
    if(count_stats > 0, toString(comment), 'N/A') as comment,
    if(count_stats > 0, toString(share), 'N/A') as share,
    if(count_stats > 0, toString(engagement), 'N/A') as engagement,
    if(count_subs > 0, toString(subscribers), 'N/A') as subscribers,
    if(count_views > 0, toString(video_views), 'N/A') as views,
    if(count_videos > 0, toString(videos), 'N/A') as videos
FROM
(
    SELECT
        channel_ids.channel_id as channel_id,
        coalesce(channel_stats.count_stats, 0) as count_stats,
        coalesce(channel_stats.watch_time, 0) as watch_time,
        coalesce(channel_stats.avg_view_duration, 0) as avg_view_duration,
        coalesce(channel_stats.like, 0) as like,
        coalesce(channel_stats.dislike, 0) as dislike,
        coalesce(channel_stats.comment, 0) as comment,
        coalesce(channel_stats.share, 0) as share,
        coalesce(channel_stats.engagement, 0) as engagement,
        coalesce(subs_stats.count_subs, 0) as count_subs,
        coalesce(subs_stats.subscribers, 0) as subscribers,
        coalesce(video_stats.count_videos, 0) as count_videos,
        coalesce(video_stats.videos, 0) as videos,
        coalesce(views_stats.count_views, 0) as count_views,
        coalesce(views_stats.video_views, 0) as video_views
    FROM
    (
        SELECT c1 as channel_id FROM VALUES {youtube_id_placeholder}
    ) AS channel_ids
    LEFT JOIN
    (
        -- channel_stats: activity insights aggregation
        SELECT
            channel_id,
            count_stats,
            watch_time,
            round(if(total_views > 0, total_est_watch_time/total_views, 0), 2) as avg_view_duration,
            like, dislike, comment, share, engagement
        FROM
        (
            SELECT
                channel_id,
                count() as count_stats,
                sum(estimated_minute_watched) as watch_time,
                sum(view_count) as total_views,
                sum(est_watch_time) as total_est_watch_time,
                sum(likes) as like,
                sum(dislikes) as dislike,
                sum(comments) as comment,
                sum(shares) as share,
                sum(likes + dislikes + comments + shares) as engagement
            FROM
            (
                SELECT
                    last_value(channel_id) as channel_id,
                    last_value(estimated_minutes_watched) as estimated_minute_watched,
                    last_value(estimated_minutes_watched)*60 as est_watch_time,
                    last_value(average_view_duration) as average_view_duration,
                    last_value(views) as view_count,
                    last_value(likes) as likes,
                    last_value(dislikes) as dislikes,
                    last_value(comments) as comments,
                    last_value(shares) as shares,
                    last_value(created_at) as created_at
                FROM youtube_activity_insights
                GROUP BY record_id
            )
            WHERE channel_id in {youtube_id_placeholder}
            AND {date_filter_created_at}
            GROUP BY channel_id
        )
    ) as channel_stats ON channel_ids.channel_id = channel_stats.channel_id
    LEFT JOIN
    (
        -- subs_stats: subscriber count with date-range fallback
        SELECT
            channel_id,
            if(count_in_range > 0, count_in_range, count_total) as count_subs,
            toInt32(if(count_in_range > 0, subscribers_in_range, subscribers_latest)) as subscribers
        FROM
        (
            SELECT
                channel_id,
                countIf(toDateTime(created_at) BETWEEN toDateTime('{start_date}', 0)
                    AND toDateTime('{end_date}', 0) + INTERVAL 1 DAY) as count_in_range,
                argMaxIf(subscriber_count, created_at,
                    toDateTime(created_at) BETWEEN toDateTime('{start_date}', 0)
                    AND toDateTime('{end_date}', 0) + INTERVAL 1 DAY) as subscribers_in_range,
                count(*) as count_total,
                argMax(subscriber_count, created_at) as subscribers_latest
            FROM youtube_channels
            WHERE channel_id in {youtube_id_placeholder}
            GROUP BY channel_id
        )
    ) as subs_stats ON channel_ids.channel_id = subs_stats.channel_id
    LEFT JOIN
    (
        -- video_stats: video count in date range
        SELECT
            channel_id,
            count(*) as count_videos,
            toInt32(count(*)) as videos
        FROM
        (
            SELECT channel_id, video_id
            FROM youtube_videos
            WHERE channel_id in {youtube_id_placeholder}
            AND {date_filter_published_at}
            GROUP BY channel_id, video_id
        )
        GROUP BY channel_id
    ) as video_stats ON channel_ids.channel_id = video_stats.channel_id
    LEFT JOIN
    (
        -- views_stats: total video views from activity insights
        SELECT
            channel_id,
            count(*) as count_views,
            toInt32(sum(view_count)) as video_views
        FROM
        (
            SELECT
                last_value(channel_id) as channel_id,
                last_value(views) as view_count,
                last_value(created_at) as created_at
            FROM youtube_activity_insights
            GROUP BY record_id
        )
        WHERE channel_id in {youtube_id_placeholder}
        AND {date_filter_created_at}
        GROUP BY channel_id
    ) as views_stats ON channel_ids.channel_id = views_stats.channel_id
)
```

---

### 8.3 Subscriber Trend (`subscriberTrendQuery`)

**Table:** `youtube_channels`

```sql
SELECT
    arrayDifference(subscribers_total) AS subscribers_gained_daily,
    subscribers_total,
    buckets
FROM
(
    SELECT
        arrayReverseFill(x -> not x==0, arrayFill(x -> not x==0, subscribers_gained)) AS subscribers_total,
        buckets
    FROM
    (
        SELECT
            groupArray(subscribers_gained_total) as subscribers_gained,
            groupArray(dates) as buckets
        FROM
        (
            SELECT toInt32(max(subscribers_gained)) as subscribers_gained_total, dates
            FROM
            (
                SELECT argMin(subscriber_count, inserted_at) as subscribers_gained,
                       toDate(inserted_at) as dates, channel_id
                FROM youtube_channels
                WHERE channel_id in ('...')
                AND toDateTime(inserted_at) >= toDateTime('START_DATE',0)
                  AND toDateTime(inserted_at) < toDateTime('END_DATE+1',0)
                GROUP BY dates, channel_id
            )
            GROUP BY dates
            ORDER BY dates ASC
            WITH FILL FROM toDate('START_DATE') TO toDate('END_DATE') + 1 STEP 1
        )
    )
)
```

Uses both `arrayFill` (forward-fill) and `arrayReverseFill` (backward-fill). Controller applies additional leading-zero replacement via `getLatestSubscriberCount()` fallback.

**Fallback (`getLatestSubscriberCount`):**

```sql
SELECT toInt32(subscriber_count) as subscriber_count
FROM youtube_channels
WHERE channel_id IN ('...')
AND subscriber_count > 0
ORDER BY inserted_at DESC
LIMIT 1
```

**Dynamic version (`subscriberDynamicTrendQuery`):** 180-day threshold. Uses `last_value(subscriber_count)` instead of `argMin`, groups by `record_id`.

---

### 8.4 Engagement Trend (`engagementTrendQuery`)

**Table:** `youtube_activity_insights`

```sql
SELECT
    arrayMap(i -> toInt32(i), arrayCumSum(like)) as like_total,
    like as like_daily,
    arrayMap(i -> toInt32(i), arrayCumSum(dislike)) as dislike_total,
    dislike as dislike_daily,
    arrayMap(i -> toInt32(i), arrayCumSum(share)) as share_total,
    share as share_daily,
    arrayMap(i -> toInt32(i), arrayCumSum(comment)) as comment_total,
    comment as comment_daily,
    arrayMap(i -> toInt32(i), arrayCumSum(engagement)) as engagement_total,
    engagement as engagement_daily,
    bucket
FROM
(
    SELECT
        groupArray(like) as like, groupArray(dislike) as dislike,
        groupArray(share) as share, groupArray(comment) as comment,
        groupArray(engagement) as engagement, groupArray(created_at) as bucket
    FROM
    (
        SELECT
            toInt32(sum(likes)) as like, toInt32(sum(dislikes)) as dislike,
            toInt32(sum(shares)) as share, toInt32(sum(comments)) as comment,
            toInt32(sum(likes + dislikes + shares + comments)) AS engagement,
            toDate(created_date) as created_at
        FROM
        (
            SELECT
                last_value(created_at) as created_date,
                last_value(likes) as likes, last_value(dislikes) as dislikes,
                last_value(shares) as shares, last_value(comments) as comments
            FROM youtube_activity_insights
            WHERE channel_id in ('...')
            AND [DateFilter on created_at]
            GROUP BY record_id
        )
        GROUP BY created_at
        ORDER BY created_at ASC
        WITH FILL FROM toDate('ACCOUNT_CREATED_DATE') STEP 1
    )
)
```

Fill starts from `account_created_date`. Uses `arrayCumSum` for running totals.

---

### 8.5 Video Views Trend (`videoViewsTrendQuery`)

**Table:** `youtube_traffic_insights`

```sql
SELECT
    subscriber_views_total, subscriber_views_daily,
    non_subscriber_views_total, non_subscriber_views_daily,
    arrayMap((i, value) -> toInt32(subscriber_views_daily[i] + non_subscriber_views_daily[i]),
             arrayEnumerate(subscriber_views_daily), subscriber_views_daily) as video_views_daily,
    arrayMap((i, value) -> toInt32(subscriber_views_total[i] + non_subscriber_views_total[i]),
             arrayEnumerate(subscriber_views_total), subscriber_views_total) as video_views_total,
    buckets
FROM
(
    SELECT
        arrayMap(i -> toInt32(i), arrayCumSum(subscriber_views)) as subscriber_views_total,
        subscriber_views as subscriber_views_daily,
        arrayMap(i -> toInt32(i), arrayCumSum(non_subscriber_views)) as non_subscriber_views_total,
        non_subscriber_views as non_subscriber_views_daily,
        created_date as buckets
    FROM
    (
        SELECT
            groupArray(subscriber_views) as subscriber_views,
            groupArray(non_subscriber_views) as non_subscriber_views,
            groupArray(created_date) as created_date
        FROM
        (
            SELECT
                toInt32(sum(subscriber_views)) as subscriber_views,
                toInt32(sum(
                    paid_views + annotation_views + end_screen_views + campaign_card_view +
                    no_link_other_views + yt_channel_views + yt_search_views + related_video_views +
                    yt_other_page_views + ext_url_views + playlist_views + notification_views + shorts_views
                )) AS non_subscriber_views,
                toDate(created_at) as created_date
            FROM
            (
                SELECT
                    last_value(channel_id) AS channel_id,
                    last_value(subscriber_views) AS subscriber_views,
                    last_value(paid_views) AS paid_views,
                    last_value(annotation_views) AS annotation_views,
                    last_value(end_screen_views) AS end_screen_views,
                    last_value(campaign_card_view) AS campaign_card_view,
                    last_value(no_link_other_views) AS no_link_other_views,
                    last_value(yt_channel_views) AS yt_channel_views,
                    last_value(yt_search_views) AS yt_search_views,
                    last_value(related_video_views) AS related_video_views,
                    last_value(yt_other_page_views) AS yt_other_page_views,
                    last_value(ext_url_views) AS ext_url_views,
                    last_value(playlist_views) AS playlist_views,
                    last_value(notification_views) AS notification_views,
                    last_value(shorts_views) AS shorts_views,
                    last_value(created_at) AS created_at
                FROM youtube_traffic_insights
                GROUP BY record_id
            )
            WHERE channel_id in ('...')
            AND [DateFilter on created_at]
            GROUP BY created_at
            ORDER BY toDate(created_at) ASC
            WITH FILL FROM toDate('ACCOUNT_CREATED_DATE') STEP 1
        )
    )
)
```

Non-subscriber views = sum of 13 traffic sources (plus `shorts_views` = 14 total).

---

### 8.6 Watch Time Trend (`watchTimeTrendQuery`)

**Table:** `youtube_traffic_insights`

```sql
SELECT
    arrayMap(x->toInt32(x), arrayCumSum(subscriber_watch_time)) as subscriber_watch_time_total,
    arrayMap(x->toInt32(x), arrayCumSum(non_subscriber_watch_time)) as non_subscriber_watch_time_total,
    subscriber_watch_time as subscriber_watch_time_daily,
    non_subscriber_watch_time as non_subscriber_watch_time_daily,
    average_watch_time,
    buckets
FROM
(
    SELECT
        groupArray(sub_watch_time) as subscriber_watch_time,
        groupArray(non_sub_watch_time) as non_subscriber_watch_time,
        avg(average_watch_time) as average_watch_time,
        groupArray(created_at) as buckets
    FROM
    (
        SELECT
            toInt32(sum(subscriber_watch_time)) as sub_watch_time,
            toInt32(sum(non_subscriber_watch_time)) as non_sub_watch_time,
            toInt32(sum(subscriber_watch_time + non_subscriber_watch_time)) as average_watch_time,
            toDate(created_at) as created_at
        FROM
        (
            SELECT
                last_value(subscriber_watch_time) as subscriber_watch_time,
                last_value(non_subsciber_watch_time) as non_subscriber_watch_time,
                last_value(created_at) as created_at,
                last_value(channel_id) as channel_id
            FROM youtube_traffic_insights
            GROUP BY record_id
        )
        WHERE channel_id in ('...')
        AND [DateFilter on created_at]
        GROUP BY created_at
        ORDER BY created_at ASC
        WITH FILL FROM toDate('START_DATE') TO toDate('END_DATE') + 1 STEP 1
    )
)
```

**CRITICAL:** Column name `non_subsciber_watch_time` has a typo (missing 's' in subscriber). Must use exact typo in Go implementation.

---

### 8.7 Find Video (`findVideoQuery`)

**Table:** `youtube_traffic_insights`

13 traffic sources (excludes `shorts_views`). Returns rows sorted ascending by value with name, value, and percentage.

```sql
SELECT
    (arrayJoin(arrayZip(names[1], values[1], perc_values)) AS t).1 as name,
    t.2 as value, t.3 as perc_value
FROM
(
    SELECT names, values, arrayMap(x -> x*100/total, values[1]) AS perc_values
    FROM
    (
        SELECT
            groupArray([*]) as values,
            groupArray(['Paid Views', 'Annotation Views', 'End Screen Views', 'Campaign Card Views',
                        'Subscriber Views', 'No Link Other views', 'YT Channel Views', 'YT Search Views',
                        'Related Video Views', 'YT Other Page Views', 'Ext URL Views',
                        'Playlist Views', 'Notification Views']) as names,
            arraySum([*]) as total
        FROM
        (
            SELECT
                toInt32(sum(paid_views)) as `Paid View`, toInt32(sum(annotation_views)) as `Annotation View`,
                toInt32(sum(end_screen_views)) as `End Screen View`, toInt32(sum(campaign_card_view)) as `Campaign Card Views`,
                toInt32(sum(subscriber_views)) as `Subscriber View`, toInt32(sum(no_link_other_views)) as `No Link Other View`,
                toInt32(sum(yt_channel_views)) as `YT Channel View`, toInt32(sum(yt_search_views)) as `YT Search View`,
                toInt32(sum(related_video_views)) as `Related Video View`, toInt32(sum(yt_other_page_views)) as `YT Other Page View`,
                toInt32(sum(ext_url_views)) as `Ext URL View`, toInt32(sum(playlist_views)) as `Playlist View`,
                toInt32(sum(notification_views)) as `Notification View`
            FROM youtube_traffic_insights
            WHERE channel_id in ('...')
            AND [DateFilter on created_at]
        )
        GROUP BY *
    )
)
ORDER BY value ASC
```

---

### 8.8 Video Sharing (`videoSharingQuery`)

**Table:** `youtube_shared_insights`

31 sharing platforms. Only takes the single most recent record. Filters out zero-value platforms.

```sql
SELECT
    (arrayJoin(arrayZip(names[1], values[1], perc_values)) AS t).1 as name,
    t.2 as value, t.3 as perc_value
FROM
(
    SELECT names, values, arrayMap(x -> x*100/total, values[1]) AS perc_values
    FROM
    (
        SELECT groupArray([*]) as values,
            groupArray(['Ameba', 'Blogger', 'Copy Paste', 'Cyworld', 'Digg', 'Dropbox', 'Embed',
                        'Mail', 'Whatsapp', 'Other', 'facebook Messenger', 'Facebook Pages', 'Facebook',
                        'Fotka', 'Vkontakte', 'Google Plus', 'Discord', 'Linkedin', 'Goo', 'Hangouts',
                        'Pinterest', 'Myspace', 'Reddit', 'Skype', 'Telegram', 'Tumblr', 'Twitter',
                        'Viber', 'Weibo', 'Wechat', 'Youtube']) as names,
            arraySum([*]) as total
        FROM
        (
            SELECT toInt32(ameba) as `Ameba`, toInt32(blogger) as `Blogger`, ...all 31 platforms...
            FROM youtube_shared_insights
            WHERE channel_id in ('...')
            AND [DateFilter on inserted_at]
            ORDER BY inserted_at DESC
            LIMIT 1
        )
        GROUP BY *
    )
)
WHERE value != 0
ORDER BY value DESC
```

---

### 8.9 Top/Least Posts (`topPostsQuery`/`leastPostsQuery`)

**Table:** `youtube_videos`

```sql
SELECT *, published_at_time as published_at FROM
(
    SELECT
        video_id,
        last_value(title) as title,
        last_value(description) as description,
        last_value(duration) as duration,
        last_value(thumbnail_url) as thumbnail_url,
        last_value(media_type) as media_type,
        concat('https://', substring(
            iframe_embed_html,
            position('//' IN iframe_embed_html) + length('//'),
            position('"' IN substring(iframe_embed_html, position('//' IN iframe_embed_html) + length('//'))) - 1
        )) AS iframe_embed_url,
        REPLACE(iframe_embed_url, 'embed/', 'watch?v=') as share_url,
        toInt32(argMax(likes, inserted_at) + argMax(dislikes, inserted_at)
                + argMax(comments, inserted_at) + argMax(shares, inserted_at)) as engagement,
        engagement as total_engagement,
        toInt32(argMax(likes, inserted_at)) as like,
        toInt32(argMax(dislikes, inserted_at)) as dislike,
        toInt32(argMax(views, inserted_at)) as views,
        toInt32(argMax(red_views, inserted_at)) as red_views,
        toInt32(argMax(favorites, inserted_at)) as favorites,
        toInt32(argMax(comments, inserted_at)) as comment,
        toInt32(argMax(subscribers_gained, inserted_at)) as subscribers_gained,
        toInt32(argMax(shares, inserted_at)) as share,
        toInt32(argMax(minutes_watched, inserted_at)) as minutes_watched,
        toInt32(argMax(red_minutes_watched, inserted_at)) as red_minutes_watched,
        toInt32(argMax(average_view_duration, inserted_at)) as average_view_duration,
        toInt32(argMax(average_view_percentage, inserted_at)) as average_view_percentage,
        if(count(*) != 0, round(engagement / toInt32(count(*)), 2), 0) as engagement_rate,
        max(published_at) as published_at_time
    FROM youtube_videos
    WHERE channel_id in ('...')
    AND [DateFilter on published_at]
    GROUP BY video_id
    ORDER BY {order_by} DESC  -- ASC for leastPostsQuery
    LIMIT {limit}  -- hardcoded 5 for leastPostsQuery
)
```

Uses `argMax(metric, inserted_at)` for deduplication. Permalink extracted from `iframe_embed_html` via `REPLACE(url, 'embed/', 'watch?v=')`.

---

### 8.10 Performance and Video Posting Schedule

**Engagement (`performanceEngagementAndVideoPostingScheduleQuery`):**

```sql
SELECT
    groupArray(count) as count, groupArray(likes) as likes,
    groupArray(dislikes) as dislikes, groupArray(shares) as shares,
    groupArray(comments) as comments, groupArray(engagement) as engagement,
    groupArray(published_at) as buckets
FROM
(
    SELECT
        toInt32(count()) as count, toInt32(sum(like)) as likes,
        toInt32(sum(dislike)) as dislikes, toInt32(sum(share)) as shares,
        toInt32(sum(comment)) as comments, toInt32(sum(engagement)) as engagement,
        toDate(published_at) as published_at
    FROM
    (
        SELECT
            last_value(likes) as like, last_value(dislikes) as dislike,
            last_value(shares) as share, last_value(comments) as comment,
            like + dislike + share + comment as engagement,
            published_at, channel_id
        FROM youtube_videos
        WHERE channel_id in ('...')
        AND [DateFilter on published_at]
        GROUP BY video_id, channel_id, published_at
    )
    GROUP BY published_at
    ORDER BY published_at ASC
    WITH FILL FROM toDate('ACCOUNT_CREATED_DATE') STEP 1
)
```

**Views (`performanceViewsAndVideoPostingScheduleQuery`):**

Uses CTE `videos` (post counts per day) LEFT JOINed with traffic insights aggregation. Same 13+1 non-subscriber view sources.

---

### 8.11 Engagement Rollup (`getEngagementRollup`)

Used by cross-platform overview page.

```sql
SELECT
    engagement_count, post_count,
    if(post_count!=0 AND {totalDays}!=0, (engagement_count/post_count)/{totalDays}, 0) as engagement_rate,
    if({totalDays}!=0, post_count/{totalDays}, 0) as post_rate,
    toInt32(reactions) as reactions, toInt32(comment) as comments, toInt32(share) as shares
FROM
(
    SELECT
        toInt32(sum(likes + dislikes + shares + comments)) as engagement_count,
        toInt32(sum(likes + dislikes)) as reactions,
        toInt32(sum(comments)) as comment, toInt32(sum(shares)) as share
    FROM
    (
        SELECT last_value(likes) as likes, last_value(dislikes) as dislikes,
            last_value(comments) as comments, last_value(shares) as shares
        FROM youtube_activity_insights
        WHERE channel_id in ('...')
        AND [DateFilter on created_at]
        AND [DateFilter on created_at]  -- NOTE: DateFilter appears TWICE (bug)
        GROUP BY record_id
    )
) AS engagement_stats
CROSS JOIN
(
    SELECT toInt32(count(*)) as post_count
    FROM (SELECT video_id FROM youtube_videos WHERE channel_id in ('...') AND [DateFilter on published_at] GROUP BY video_id)
) as post_count
```

Bug: `DateFilter('created_at')` is called twice.

---

## 9. TikTok Analytics

### 9.1 Constructor

```
timezone = payload['timezone'] ?? 'UTC'
date -> split on ' - ' -> startDate (00:00:00), endDate (23:59:59)
tiktok_id -> string or array, merged with optional tiktok_accounts
page_id = SQL IN-clause string e.g. ('id1','id2')
post_table = 'tiktok_posts'
insights_table = 'tiktok_insights'
```

### 9.2 Key Methods

| Method | Description |
|--------|-------------|
| `getPageAndPostsInsights` | Summary with current/previous comparison |
| `getPageFollowersAndViews` | Followers and views time-series |
| `getDynamicPageFollowersAndViews` | Dynamic daily/monthly version |
| `getPostsAndEngagements` | Post counts and engagement per day |
| `getDailyEngagementsData` | Cumulative engagement with daily deltas |
| `getDynamicDailyEngagementsData` | Dynamic daily/monthly version |
| `getTopAndLeastPerformingPosts` | UNION ALL of top-5 and bottom-5 |
| `getPostsData` | Paginated posts list |

### 9.3 `getPageAndPostsInsights()`

**Tables:** `tiktok_posts`, `tiktok_insights`

```sql
select
    tiktok_id as tiktok_id,
    if(posts_summary.display_name == '', '{platformName}', posts_summary.display_name) as page_name,
    '{platform_logo}' as logo,
    toInt32(posts_summary.total_likes) as total_likes,
    toInt32(posts_summary.total_comments) as total_comments,
    toInt32(posts_summary.total_shares) as total_shares,
    toInt32(posts_summary.total_engagements) as total_engagements,
    toInt32(posts_summary.total_posts) as total_posts,
    if(insights_summary.tiktok_id = '','N/A',toString(insights_summary.total_follower_count)) as total_follower_count,
    if(insights_summary.tiktok_id = '','N/A',toString(insights_summary.total_following_count)) as total_following_count,
    if(insights_summary.tiktok_id = '','N/A',toString(insights_summary.total_video_views)) as total_video_views
from
(
    select
        tiktok_id as tiktok_id,
        display_name as display_name,
        total_likes, total_comments, total_shares, total_engagements, total_posts
    from
    (
        SELECT c1 as tiktok_id
        FROM VALUES {page_id}
    ) as platform_id
    left join
    (
        select
            tiktok_id,
            last_value(display_name) as display_name,
            sum(like_count) as total_likes,
            sum(comments_count) as total_comments,
            sum(share_count) as total_shares,
            sum(engagement_count) as total_engagements,
            count() as total_posts
        from
        (
            select
                tiktok_id, post_id,
                last_value(display_name) as display_name,
                max(like_count) as like_count,
                max(comments_count) as comments_count,
                max(share_count) as share_count,
                max(engagement_count) as engagement_count
            from tiktok_posts
            where tiktok_id in {page_id}
              AND toDateTime(created_at, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
            group by tiktok_id, post_id
        )
        group by tiktok_id
    ) as post_data on platform_id.tiktok_id = post_data.tiktok_id
) as posts_summary
left join
(
    select
        tiktok_id, total_follower_count, total_following_count, total_video_views
    from tiktok_insights
    where tiktok_id in {page_id}
      AND toDateTime(inserted_at, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
    order by inserted_at desc
    limit 1
) as insights_summary on posts_summary.tiktok_id = insights_summary.tiktok_id
```

**Deduplication:** Inner subquery groups by `(tiktok_id, post_id)` taking `max()` of each metric. Insights uses `limit 1 order by inserted_at desc` for most recent row.

---

### 9.4 Followers and Views (`getPageFollowersAndViews`)

**Table:** `tiktok_insights`

```sql
select
    platform_id,
    max(display_name) as display_name,
    '{platform_logo}' as logo,
    groupArray(follower_count) as followers_count,
    groupArray(views_per_day) as views_per_day,
    groupArray(follower_count_diff) as followers_count_diff,
    groupArray(views_per_day_diff) as views_per_day_diff,
    groupArray(inserted_at) as day_bucket
from
(
    select
        max(tiktok_id) as platform_id,
        max(display_name) as display_name,
        max(total_follower_count) as follower_count,
        runningDifference(max(total_follower_count)) as follower_count_diff,
        if(runningDifference(max(total_video_views))<0, 0, runningDifference(max(total_video_views))) AS views_per_day_diff,
        max(total_video_views) AS views_per_day,
        toDateTime(inserted_at) as inserted_at
    from tiktok_insights
    where tiktok_id in {page_id}
      and toDateTime(inserted_at, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
    group by inserted_at
    order by inserted_at
    WITH FILL
        FROM toStartOfDay(toDate('{startDate}'))
        TO toStartOfDay(toDate('{endDate+1day}'))
        STEP INTERVAL 1 DAY
    INTERPOLATE (
        platform_id AS {page_id},
        display_name as '{platform_name}',
        follower_count as -1,
        views_per_day as -1
    )
)
group by platform_id
```

**Controller logic:** Subtracts 1 day from startDate before query, then shifts off first array element. Normalizes leading `-1` sentinels (trim + forward-fill).

---

### 9.5 Posts and Engagements (`getPostsAndEngagements`)

**Table:** `tiktok_posts`

```sql
select
    tiktok_id,
    '{platformName}' as page_name,
    '{platform_logo}' as logo,
    groupArray(posting_day) as days_bucket,
    groupArray(sum_view_count) as sum_view_count,
    groupArray(sum_like_count) as sum_like_count,
    groupArray(sum_comments_count) as sum_comments_count,
    groupArray(sum_share_count) as sum_share_count,
    groupArray(sum_engagement_count) as sum_engagement_count,
    groupArray(avg_engagement_rate) as avg_engagement_rate,
    groupArray(post_count) as post_count
from
(
    select
        tiktok_id,
        toDate(day_of_post) as posting_day,
        toInt32(sum(view_count)) as sum_view_count,
        toInt32(sum(like_count)) as sum_like_count,
        toInt32(sum(comments_count)) as sum_comments_count,
        toInt32(sum(share_count)) as sum_share_count,
        toInt32(sum(engagement_count)) as sum_engagement_count,
        toInt32(avg(round(engagement_rate, 2))) as avg_engagement_rate,
        toInt32(count()) as post_count
    from
    (
        select
            post_id, tiktok_id,
            max(view_count) as view_count,
            max(like_count) as like_count,
            max(comments_count) as comments_count,
            max(share_count) as share_count,
            max(engagement_count) as engagement_count,
            max(engagement_rate) as engagement_rate,
            toDate(max(created_at)) as day_of_post
        from tiktok_posts
        where tiktok_id in {page_id}
          AND toDateTime(created_at, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
        group by tiktok_id, post_id
    )
    group by tiktok_id, day_of_post
    order by toDate(day_of_post)
    WITH FILL
        FROM toDate('{startDate}')
        TO toDate('{endDate+1day}')
        STEP INTERVAL 1 DAY
    INTERPOLATE (tiktok_id AS {page_id})
)
group by tiktok_id
```

**Controller:** If all `post_count[]` elements are zero, empties all array columns.

---

### 9.6 Daily Engagements (`getDailyEngagementsData`)

**Table:** `tiktok_insights`

```sql
select
    platform_id as tiktok_id,
    '{platformName}' as page_name,
    '{platform_logo}' as logo,
    groupArray(total_video_likes) as total_video_likes,
    groupArray(total_video_comments) as total_video_comments,
    groupArray(total_video_shares) as total_video_shares,
    groupArray(daily_video_likes) as daily_video_likes,
    groupArray(daily_video_comments) as daily_video_comments,
    groupArray(daily_video_shares) as daily_video_shares,
    groupArray(total_engagement) as total_engagement,
    groupArray(daily_engagements) as daily_engagement,
    groupArray(day_of_metric) as days_bucket
from
(
    select
        MAX(record_id) as record_id,
        MAX(total_video_likes) as total_video_likes,
        MAX(total_video_comments) as total_video_comments,
        MAX(total_video_shares) as total_video_shares,
        toInt32(MAX(total_engagement)) as total_engagement,
        toInt32(if(MAX(daily_video_likes)<0, 0, MAX(daily_video_likes))) as daily_video_likes,
        toInt32(if(MAX(daily_video_comments)<0, 0, MAX(daily_video_comments))) as daily_video_comments,
        toInt32(if(MAX(daily_video_shares)<0, 0, MAX(daily_video_shares))) as daily_video_shares,
        toInt32(daily_video_likes + daily_video_comments + daily_video_shares) as daily_engagements,
        platform_id,
        toDateTime(metric_day) as day_of_metric
    from
    (
        select
            record_id,
            MAX(total_video_likes) as total_video_likes,
            MAX(total_video_comments) as total_video_comments,
            MAX(total_video_shares) as total_video_shares,
            total_video_likes + total_video_comments + total_video_shares as total_engagement,
            toInt32(runningDifference(total_video_likes)) as daily_video_likes,
            toInt32(runningDifference(total_video_comments)) as daily_video_comments,
            toInt32(runningDifference(total_video_shares)) as daily_video_shares,
            max(tiktok_id) as platform_id,
            toDate(max(inserted_at)) as metric_day
        from tiktok_insights
        where tiktok_id in {page_id}
          AND toDateTime(inserted_at, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
        group by record_id
    )
    group by platform_id, metric_day
    order by toDateTime(metric_day)
    WITH FILL
        FROM toStartOfDay(toDate('{startDate}'))
        TO toStartOfDay(toDate('{endDate+1day}'))
        STEP INTERVAL 1 DAY
    INTERPOLATE (
        platform_id AS {page_id},
        total_video_likes as -1,
        total_video_comments as -1,
        total_video_shares as -1
    )
)
group by platform_id
```

**Logic:** `runningDifference()` on cumulative totals produces daily deltas. Negative diffs clamped to 0. `-1` sentinel marks gap days.

---

### 9.7 Top and Least Performing Posts (`getTopAndLeastPerformingPosts`)

**Tables:** `tiktok_posts`, `tiktok_insights`

```sql
select * from
(
    select * from
    (
        (
            select
                'top_posts' as category,
                tiktok_id,
                '{platformName}' as page_name,
                '{platform_logo}' as logo,
                max(profile_link) as profile_link,
                post_id,
                max(cover_image_url) as cover_image_url,
                max(share_url) as share_url,
                max(post_description) as post_description,
                max(hashtags) as hashtags,
                max(duration) as duration,
                max(height) as height,
                max(width) as width,
                max(title) as title,
                max(embed_html) as embed_html,
                max(embed_link) as embed_link,
                max(like_count) as likes_count,
                max(comments_count) as comments_count,
                max(share_count) as shares_count,
                max(view_count) as views_count,
                max(engagement_count) as engagement_count,
                round(max(engagement_rate), 2) as engagement_rate,
                max(inserted_at) as inserted_at,
                max(created_at) as created_time
            from tiktok_posts
            WHERE tiktok_id in {page_id}
              AND toDateTime(created_at, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
            group by tiktok_id, post_id
            order by engagement_count desc
            limit 5
        )
        UNION ALL
        (
            select
                'least_posts' as category,
                -- same columns as above --
            from tiktok_posts
            WHERE tiktok_id in {page_id}
              AND toDateTime(created_at, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
            group by tiktok_id, post_id
            order by engagement_count asc
            limit 5
        )
    ) as post_data
    LEFT JOIN
    (
        select total_follower_count, tiktok_id
        from tiktok_insights
        WHERE tiktok_id in {page_id}
        order by inserted_at desc
        limit 1
    ) as insights_data on post_data.tiktok_id = insights_data.tiktok_id
)
```

**Note:** Insights join has NO date filter -- always uses latest `tiktok_insights` row.

---

### 9.8 Posts Data (`getPostsData($sortOrder, $limit, $offset)`)

**Tables:** `tiktok_posts`, `tiktok_insights`

```sql
select * from
(
    select * FROM
    (
        select
            tiktok_id,
            '{platformName}' as page_name,
            '{platform_logo}' as logo,
            max(profile_link) as profile_link,
            post_id,
            max(cover_image_url) as cover_image_url,
            max(share_url) as share_url,
            max(post_description) as post_description,
            max(hashtags) as hashtags,
            max(duration) as duration,
            max(height) as height,
            max(width) as width,
            max(title) as title,
            max(embed_html) as embed_html,
            max(embed_link) as embed_link,
            max(like_count) as likes_count,
            max(comments_count) as comments_count,
            max(share_count) as shares_count,
            max(view_count) as views_count,
            max(engagement_count) as engagements_count,
            engagements_count as total_engagement,
            round(max(engagement_rate), 2) as engagement_rate,
            max(inserted_at) as inserted_at,
            max(created_at) as created_time
        from tiktok_posts
        where tiktok_id in {page_id}
          and toDateTime(created_at, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
        group by tiktok_id, post_id
        order by {sortOrder} desc
        limit {limit}
        OFFSET {offset}
    ) as posts_data
    LEFT JOIN
    (
        SELECT tiktok_id, count() as total
        from
        (
            select tiktok_id, post_id, max(created_at) as created_time
            from tiktok_posts
            where tiktok_id in {page_id}
              and toDateTime(created_at, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
            group by tiktok_id, post_id
        )
        group by tiktok_id
    ) as post_count on post_count.tiktok_id = posts_data.tiktok_id
) as post_info
LEFT JOIN
(
    select total_follower_count, tiktok_id
    from tiktok_insights
    WHERE tiktok_id in {page_id}
    order by inserted_at desc
    limit 1
) as insights_data on post_info.tiktok_id = insights_data.tiktok_id
```

**Pagination:** `LIMIT {limit} OFFSET {offset}` on the posts subquery. Total count joined as `total`.

---

### 9.9 Engagement Rollup (`getEngagementRollup`)

**Table:** `tiktok_posts`

```sql
SELECT
    toInt32(sum(engagement_count)) as engagement_count,
    toInt32(sum(like_count)) as like_count,
    toInt32(sum(comments_count)) as comments_count,
    toInt32(sum(share_count)) as share_count,
    avg(engagement_rate) as engagement_rate,
    toInt32(count()) as post_count,
    count()/{numberOfDays} as post_rate
from
(
    SELECT
        last_value(like_count) as like_count,
        last_value(comments_count) as comments_count,
        last_value(share_count) as share_count,
        last_value(engagement_count) as engagement_count,
        last_value(engagement_rate) as engagement_rate
    FROM tiktok_posts
    where tiktok_id in {page_id}
      AND toDateTime(created_at, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
    GROUP BY post_id
)
```

### 9.10 Dynamic Aggregation

TikTok uses 180-day threshold:
- `dateDiff <= 180` -> daily (`INTERVAL 1 DAY`, `toStartOfDay`)
- `dateDiff > 180` -> monthly (`INTERVAL 1 MONTH`, `toStartOfMonth`)

Both `getDynamicPageFollowersAndViews()` and `getDynamicDailyEngagementsData()` return additional `aggregation_level` field.

---

## 10. Pinterest Analytics

### 10.1 Constructor

```
ids -> pinterest_id (the user_id)
board_id -> triggers board mode (user mode is default)
filter_by -> 'video' or 'image'; defaults to "'video', 'image'"
order_by -> column name for sorting pins
date -> defaults to last 30 days if not provided
timezone -> 'Europe/Kyiv' remapped to 'Europe/Riga'
```

### 10.2 User Mode vs Board Mode

Every Pinterest endpoint has two query variants:
- **User mode:** Filters by `user_id`, uses `pinterest_user_insights` and `pinterest_users`
- **Board mode:** Filters by `board_id`, uses `pinterest_pins`, `pinterest_pin_insights`, `pinterest_boards`

### 10.3 Date Handling

Uses `toDate()` not `toDateTime()` -- timezone-naive date-level comparison.

```sql
-- DateFilter:
toDate({date_field}) BETWEEN toDate('{currentStartDate}') AND toDate('{currentEndDate+1day}')

-- DateFill:
WITH FILL FROM toDate('{account_created_date}') TO toDate('{currentEndDate}') STEP 1
```

### 10.4 `summaryQueryForUser()`

**Tables:** `pinterest_user_insights`, `pinterest_users`

```sql
SELECT
    if(count > 0, toString(followers), 'N/A') as follower_count,
    if(count_insights > 0, toString(impressions), 'N/A') as impressions,
    if(count_insights > 0, toString(pin_clicks), 'N/A') as pin_clicks,
    if(count_insights > 0, toString(outbound_clicks), 'N/A') as outbound_clicks,
    if(count_insights > 0, toString(saves), 'N/A') as saves,
    if(count_insights > 0, toString(engagement), 'N/A') as total_engagement
FROM
(
    SELECT
        user_id,
        toInt32(count()) as count_insights,
        toInt32(sum(impressions)) as impressions,
        toInt32(sum(pin_clicks)) as pin_clicks,
        toInt32(sum(outbound_clicks)) as outbound_clicks,
        toInt32(sum(saves)) as saves,
        toInt32(sum(engagement)) as engagement
    FROM (
        SELECT
            user_id,
            max(impression) as impressions,
            max(pin_clicks) as pin_clicks,
            max(outbound_click) as outbound_clicks,
            max(saves) as saves,
            max(engagement) as engagement,
            toDate(created_at) as saving_time
        FROM pinterest_user_insights
        WHERE user_id in {pinterest_id}
          AND toDate(created_at) BETWEEN toDate('{startDate}') AND toDate('{endDate+1}')
        GROUP BY saving_time, user_id
    )
    GROUP BY user_id
) AS pin_insights
LEFT JOIN (
    SELECT first_value(follower_count) as followers,
           user_id,
           toInt32(count()) as count
    FROM pinterest_users
    WHERE user_id in {pinterest_id}
      AND toDate(inserted_at) BETWEEN toDate('{startDate}') AND toDate('{endDate+1}')
    GROUP BY user_id
) as pin_data USING user_id
```

---

### 10.5 `summaryQueryForBoard()`

**Tables:** `pinterest_pins`, `pinterest_boards`, `pinterest_pin_insights`

```sql
WITH pins AS (
    SELECT pin_id
    FROM pinterest_pins
    WHERE board_id IN {board_id}
    GROUP BY pin_id
)
SELECT board_id,
    if(count_board_data > 0, toString(follower_count), 'N/A') as follower_count,
    if(count > 0, toString(impressions), 'N/A') as impressions,
    if(count > 0, toString(pin_clicks), 'N/A') as pin_clicks,
    if(count > 0, toString(outbound_clicks), 'N/A') as outbound_clicks,
    if(count > 0, toString(saves), 'N/A') as saves,
    if(count > 0, toString(engagement), 'N/A') as total_engagement
FROM
(
    SELECT board_id,
           count() as count_board_data,
           toInt32(last_value(follower_count)) as follower_count
    FROM pinterest_boards
    WHERE board_id IN {board_id}
      AND toDate(inserted_at) BETWEEN toDate('{startDate}') AND toDate('{endDate+1}')
    GROUP BY board_id
) AS board_data
CROSS JOIN
(
    SELECT
        toInt32(count()) as count,
        toInt32(sum(impressions)) as impressions,
        toInt32(sum(pin_clicks)) as pin_clicks,
        toInt32(sum(outbound_clicks)) as outbound_clicks,
        toInt32(sum(saves)) as saves,
        toInt32(sum(engagement)) as engagement
    FROM (
        SELECT
            max(impression) as impressions,
            max(pin_clicks) as pin_clicks,
            max(outbound_click) as outbound_clicks,
            max(saves) as saves,
            max(engagement) as engagement,
            toDate(created_at) as saving_time
        FROM pinterest_pin_insights
        WHERE pin_id in (SELECT pin_id FROM pins)
          AND toDate(saving_time) BETWEEN toDate('{startDate}') AND toDate('{endDate+1}')
        GROUP BY pin_id, saving_time
    ) as pi
) as pin_metrics
```

**Note:** Uses `CROSS JOIN` -- valid because both sides produce exactly one row.

---

### 10.6 Followers Queries

#### `followersQueryForUser()`

**Table:** `pinterest_users`

```sql
SELECT
    arrayDifference(followers_total) as followers_daily,
    arrayFill((x) -> not x==0, followers_total) as followers_gained,
    show_data,
    buckets
FROM
(
    SELECT groupArray(followers) as followers_total,
           groupArray(saving_time) as buckets,
           toInt32(sum(followers)) as show_data
    FROM
    (
        SELECT toInt32(argMin(follower_count, inserted_at)) as followers,
               toDate(inserted_at) as saving_time
        FROM pinterest_users
        WHERE user_id in {pinterest_id}
          AND toDate(inserted_at) BETWEEN toDate('{startDate}') AND toDate('{endDate+1}')
        GROUP BY saving_time
        ORDER BY saving_time ASC
        WITH FILL
        TO toDate('{endDate}')
        STEP 1
    )
)
```

**Logic:** `argMin(follower_count, inserted_at)` picks earliest snapshot's follower count per day. `arrayDifference` computes daily gains. `arrayFill(x -> not x==0, ...)` forward-fills zeros.

#### `followersQueryForBoard()`

**Table:** `pinterest_boards`

```sql
-- Same structure, filters by user_id and board_id in pinterest_boards
SELECT toInt32(argMin(follower_count, inserted_at)) as followers,
       toDate(inserted_at) as saving_time
FROM pinterest_boards
WHERE user_id in {pinterest_id}
  AND board_id in {board_id}
  AND toDate(inserted_at) BETWEEN toDate('{startDate}') AND toDate('{endDate+1}')
GROUP BY saving_time
ORDER BY saving_time ASC
WITH FILL TO toDate('{endDate}') STEP 1
```

---

### 10.7 Impressions Queries

#### `impressionsQueryForUser()`

**Table:** `pinterest_user_insights`

```sql
SELECT
    arrayMap(x -> CAST(x AS Int32), arrayCumSum(impressions_daily)) as impressions_total,
    impressions_daily,
    show_data,
    buckets
FROM (
    SELECT groupArray(impressions) as impressions_daily,
           groupArray(saving_time) as buckets,
           toInt32(sum(impressions)) as show_data
    FROM (
        SELECT toInt32(sum(impression)) as impressions,
               toDate(created_at) as saving_time
        FROM pinterest_user_insights
        WHERE user_id in ({pinterest_id})
          AND toDate(created_at) BETWEEN toDate('{startDate}') AND toDate('{endDate+1}')
        GROUP BY saving_time
        ORDER BY saving_time ASC
        WITH FILL FROM toDate('{account_created_date}') TO toDate('{endDate}') STEP 1
    )
)
```

**Logic:** `arrayCumSum` gives cumulative total. Cast to `Int32` array.

#### `impressionsQueryForBoard()`

**Tables:** `pinterest_pins`, `pinterest_pin_insights`

```sql
WITH pins as (
    SELECT pin_id FROM pinterest_pins WHERE board_id in ({board_id})
)
-- Same structure but queries pinterest_pin_insights WHERE pin_id in pins
```

---

### 10.8 Engagement Queries

#### `engagementQueryForUser()`

**Table:** `pinterest_user_insights`

```sql
SELECT
    saves as saves_daily,
    arrayMap(x -> CAST(x AS Int32), arrayCumSum(saves)) as saves_total,
    outbound_clicks as outbound_clicks_daily,
    arrayMap(x -> CAST(x AS Int32), arrayCumSum(outbound_clicks)) as outbound_clicks_total,
    arrayMap(x -> CAST(x AS Int32), arrayCumSum(pin_clicks)) as pin_clicks_total,
    pin_clicks as pin_clicks_daily,
    engagements as engagement_daily,
    arrayMap(x -> CAST(x AS Int32), arrayCumSum(engagements)) as engagement_total,
    show_data,
    buckets
FROM (
    SELECT
        groupArray(pin_clicks) as pin_clicks,
        groupArray(outbound_clicks) as outbound_clicks,
        groupArray(saves) as saves,
        groupArray(engagement) as engagements,
        toInt32(sum(engagement)) as show_data,
        groupArray(saving_time) as buckets
    FROM (
        SELECT
            toInt32(sum(pin_clicks)) as pin_clicks,
            toInt32(sum(outbound_click)) as outbound_clicks,
            toInt32(sum(saves)) as saves,
            toInt32(sum(engagement)) as engagement,
            toDate(created_at) as saving_time
        FROM pinterest_user_insights
        WHERE user_id in ({pinterest_id})
          AND toDate(created_at) BETWEEN toDate('{startDate}') AND toDate('{endDate+1}')
        GROUP BY saving_time
        ORDER BY saving_time ASC
        WITH FILL FROM toDate('{account_created_date}') TO toDate('{endDate}') STEP 1
    )
)
```

#### `engagementQueryForBoard()`

**Tables:** `pinterest_pins`, `pinterest_pin_insights`

Same structure with CTE `pins` and filtering `pinterest_pin_insights WHERE pin_id in pins`.

---

### 10.9 Pin Posting Per Day

#### `pinPostingPerDayQueryForUser()`

**Table:** `pinterest_pins`

```sql
SELECT groupArray(pin_count) as pins_count,
       groupArray(created_at) as buckets,
       toInt32(sum(pin_count)) as show_data
FROM (
    SELECT toInt32(count(*)) as pin_count,
           toDate(created_at) as created_at
    FROM (
        SELECT pin_id, created_at as created_at
        FROM pinterest_pins
        WHERE user_id in ({pinterest_id})
          AND media_type in ({filter_by})
          AND is_owner = 1
          AND toDate(created_at) BETWEEN toDate('{startDate}') AND toDate('{endDate+1}')
        GROUP BY pin_id, created_at
    )
    GROUP BY created_at
    ORDER BY created_at ASC
    WITH FILL FROM toDate('{account_created_date}') TO toDate('{endDate}') STEP 1
)
```

**Filter:** `is_owner = 1` -- only counts pins the user published, not repins.

---

### 10.10 Pin Posting Rollup (`pinPostingRollupQueryForUser`)

**Tables:** `pinterest_pins`, `pinterest_pin_insights`

```sql
WITH pins as (
    SELECT pin_id
    FROM pinterest_pins
    WHERE user_id in ({pinterest_id})
      AND media_type in ({filter_by})
      AND is_owner = 1
      AND toDate(created_at) BETWEEN toDate('{startDate}') AND toDate('{endDate+1}')
    GROUP BY pin_id
),
insights AS (
    SELECT pin_id,
           SUM(impression) AS impression,
           SUM(pin_clicks) AS pin_clicks,
           SUM(outbound_click) AS outbound_click,
           SUM(saves) AS saves,
           SUM(quartile_95s_percent_view) AS quartile_95s_percent_view,
           SUM(closeup) AS closeup,
           SUM(video_start) AS video_start,
           SUM(video_10s_view) AS video_10s_view,
           SUM(video_avg_watch_time) AS video_avg_watch_time
    FROM pinterest_pin_insights
    WHERE pin_id in (SELECT pin_id FROM pins)
    GROUP BY pin_id
)
SELECT
    count(pin_id) as total_pins,
    if(total_pins > 0, toString(sum(impression)), 'N/A') as impressions,
    if(total_pins > 0, toString(sum(pin_clicks)), 'N/A') as pin_clicks,
    if(total_pins > 0, toString(sum(outbound_click)), 'N/A') as outbound_clicks,
    if(total_pins > 0, toString(sum(saves)), 'N/A') as saves,
    if(total_pins > 0, toString(sum(quartile_95s_percent_view)), 'N/A') as quartile_95s_percent_view,
    if(total_pins > 0, toString(sum(video_start)), 'N/A') as video_views,
    if(total_pins > 0, toString(sum(video_10s_view)), 'N/A') as video_10s_view,
    if(avg(avg_watch_time) > 0 and total_pins > 0, toString(round(avg(avg_watch_time), 2)), 'N/A') as avg_watch_time
FROM (
    SELECT
        p.pin_id as pin_id,
        sum(insights.impression) as impression,
        sum(insights.pin_clicks) as pin_clicks,
        sum(insights.outbound_click) as outbound_click,
        sum(insights.saves) as saves,
        sum(insights.quartile_95s_percent_view) as quartile_95s_percent_view,
        sum(insights.closeup) as closeup,
        sum(insights.video_start) as video_start,
        sum(insights.video_10s_view) as video_10s_view,
        sum(insights.video_avg_watch_time) as avg_watch_time
    FROM pins p
    LEFT JOIN insights ON p.pin_id = insights.pin_id
    GROUP BY p.pin_id
)
```

---

### 10.11 Pins List (`pinsQueryForUser($sort_by)`)

**Tables:** `pinterest_pins`, `pinterest_boards`, `pinterest_pin_insights`

```sql
SELECT
    pins.pin_id as pin_id,
    last_value(boards.name) as board_name,
    format('{}{}', 'https://www.pinterest.com/pin/', pins.pin_id) AS permalink,
    format('{}{}', 'https://assets.pinterest.com/ext/embed.html?id=', pins.pin_id) AS embed_link,
    last_value(pins.title) as title,
    last_value(pins.description) as description,
    last_value(pins.board_owner) as board_owner,
    replaceOne(last_value(pins.media_type), '_', ' ') as media_type,
    last_value(pins.cover_image_url) as cover_image_url,
    last_value(pins.dominant_color) as dominant_color,
    last_value(pins.creative_type) as creative_type,
    last_value(pins.product_tags) as product_tags,
    last_value(pins.height) as height,
    last_value(pins.width) as width,
    last_value(pins.created_at) as created_at,
    toInt32(sum(pinterest_pin_insights.impression)) AS impressions,
    toInt32(sum(pinterest_pin_insights.pin_clicks)) AS pin_clicks,
    toInt32(sum(pinterest_pin_insights.outbound_click)) AS outbound_clicks,
    toInt32(sum(pinterest_pin_insights.saves)) AS saves,
    toInt32(sum(pinterest_pin_insights.engagement)) as total_engagement,
    if(count(*) != 0,
       round(toInt32(pin_clicks + outbound_clicks + saves) / toInt32(count(*)), 2),
       0) as engagement_rate
FROM (
    SELECT pin_id, board_id,
           last_value(title) as title,
           last_value(description) as description,
           last_value(board_owner) as board_owner,
           last_value(media_type) as media_type,
           last_value(cover_image_url) as cover_image_url,
           last_value(dominant_color) as dominant_color,
           last_value(creative_type) as creative_type,
           last_value(product_tags) as product_tags,
           last_value(height) as height,
           last_value(width) as width,
           created_at
    FROM pinterest_pins
    WHERE user_id in {pinterest_id}
      AND toDate(created_at) BETWEEN toDate('{startDate}') AND toDate('{endDate+1}')
      AND is_owner = 1
    GROUP BY pin_id, board_id, created_at
) AS pins
LEFT JOIN (
    SELECT board_id, last_value(name) as name
    FROM pinterest_boards
    GROUP BY board_id
) AS boards ON pins.board_id = boards.board_id
LEFT JOIN (
    SELECT max(impression) as impression,
           max(pin_clicks) as pin_clicks,
           max(outbound_click) as outbound_click,
           max(saves) as saves,
           max(engagement) as engagement,
           pin_id
    FROM pinterest_pin_insights
    WHERE user_id in {pinterest_id}
    GROUP BY pin_id, created_at
    ORDER BY created_at DESC
) AS pinterest_pin_insights USING pin_id
GROUP BY pins.pin_id, pins.board_id
ORDER BY {order_by} {sort_by}
LIMIT {limit}
```

**Special:** `replaceOne(media_type, '_', ' ')` cleans up values. URLs built with `format()`. `engagement_rate = (pin_clicks + outbound_clicks + saves) / count(*)`.

---

### 10.12 Pin Posting Performance (`pinPostingPerformancePerDayQueryForUser`)

**Tables:** `pinterest_pins`, `pinterest_pin_insights`

```sql
WITH pins AS (
    SELECT pin_id, toDate(created_at) as saving_time
    FROM pinterest_pins
    WHERE user_id IN {pinterest_id}
      AND media_type IN ({filter_by})
      AND is_owner = 1
      AND toDate(created_at) BETWEEN toDate('{startDate}') AND toDate('{endDate+1}')
    GROUP BY pin_id, saving_time
    ORDER BY saving_time ASC
),
pin_metrics AS (
    SELECT
        pin_id,
        toInt32(SUM(impression)) AS impressions,
        toInt32(SUM(pin_clicks)) AS pin_clicks,
        toInt32(SUM(outbound_click)) AS outbound_clicks,
        toInt32(SUM(saves)) AS saves,
        toInt32(SUM(engagement)) AS engagement
    FROM pinterest_pin_insights
    WHERE pin_id IN (SELECT pin_id FROM pins)
    GROUP BY pin_id
),
pin_count AS (
    SELECT toDate(created_at) AS saving_time,
           toInt32(COUNT(*)) AS pin_count
    FROM pinterest_pins
    WHERE user_id IN {pinterest_id}
      AND toDate(created_at) BETWEEN toDate('{startDate}') AND toDate('{endDate+1}')
    GROUP BY saving_time
)
SELECT
    groupArray(pin_count) AS pins_count,
    groupArray(pin_clicks) AS pin_clicks,
    groupArray(outbound_clicks) AS outbound_clicks,
    groupArray(saves) AS saves,
    groupArray(engagement) AS engagements,
    groupArray(impressions) AS impressions,
    groupArray(saving_time) AS buckets,
    toInt32(arraySum(pin_clicks)) AS show_data
FROM (
    SELECT
        last_value(pin_id) as pin_id,
        last_value(impressions) as impressions,
        last_value(pin_clicks) as pin_clicks,
        last_value(pin_count) as pin_count,
        last_value(saves) as saves,
        last_value(engagement) as engagement,
        last_value(outbound_clicks) as outbound_clicks,
        saving_time
    FROM (SELECT * FROM pins LEFT JOIN pin_metrics USING pin_id) AS combined_metrics
    LEFT JOIN pin_count USING saving_time
    GROUP BY saving_time
    ORDER BY saving_time ASC
    WITH FILL FROM toDate('{account_created_date}') TO toDate('{endDate}') STEP 1
)
```

### 10.13 Dynamic Aggregation

180-day threshold (same as YouTube/TikTok, different from 60-day for Facebook/Instagram/LinkedIn):
- `dateDiff <= 180` -> daily (`toDate(inserted_at)` + `STEP 1`)
- `dateDiff > 180` -> monthly (`toStartOfMonth(toDate(inserted_at))` + `STEP INTERVAL 1 MONTH`)

---

## 11. Twitter/X Analytics

### 11.1 Constructor

```
timezone = payload['timezone'] ?? 'UTC'
date -> split on ' - ' -> startDate (00:00:00), endDate (23:59:59)
twitter_id -> string or array, merged with optional twitter_accounts
page_id = SQL IN-clause string
post_table = 'twitter_posts'
insights_table = 'twitter_insights'
```

### 11.2 Key Methods

| Method | Description |
|--------|-------------|
| `getPageAndPostsInsights` | Summary with current/previous |
| `getEngagementImpressionData` | 3-level aggregation, NO zero-fill |
| `getFollowerTrend` | Follower count with `runningDifference`, NO zero-fill |
| `getTopTweets` / `getLeastTweets` | Top/bottom tweets by sort column |
| `getCreditsUsedCount` | MongoDB job logs (not ClickHouse) |

### 11.3 Important Differences

- **No zero-fill in any query** (unlike all other platforms)
- Settings methods (add/remove/get accounts, hashtags) use MongoDB, not ClickHouse
- Copy-paste bug: key `tiktok_id` holds the Twitter account ID in some response structures
- `like_count` exists in **both** `twitter_insights` and `twitter_posts` tables

### 11.4 `getPageAndPostsInsights()`

**Tables:** `twitter_insights`, `twitter_posts`

```sql
select *
from
(
    select
        first_value(twitter_id) as twitter_id,
        first_value(name) as name,
        first_value(profile_image_url) as profile_image_url,
        first_value(followers_count) as followers_count,
        first_value(following_count) as following_count,
        first_value(tweet_count) as tweet_count,
        first_value(listed_count) as listed_count,
        first_value(like_count) as like_count,
        first_value(record_date) as record_date
    from
    (
        select
            twitter_id,
            max(name) as name,
            max(profile_image_url) as profile_image_url,
            argMin(followers_count, saving_time) as followers_count,
            argMin(following_count, saving_time) as following_count,
            max(tweet_count) as tweet_count,
            max(listed_count) as listed_count,
            max(like_count) as like_count,
            max(saving_time) as record_date
        from twitter_insights
        where twitter_id in {page_id}
          and toDateTime(saving_time, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
        group by twitter_id, record_id
        order by record_date desc
    )
) as insights_data
left join
(
    select
        twitter_id,
        max(name) as name,
        max(profile_image_url) as profile_image_url,
        sum(impression_count) as impression_count,
        sum(total_engagement) as total_engagement,
        groupUniqArray(tweet_type) as tweet_type,
        sum(reply_count) as reply_count,
        sum(retweet_count) as retweet_count,
        sum(bookmark_count) as bookmark_count,
        sum(like_count) as like_count,
        sum(quote_count) as quote_count,
        count() as tweet_count
    from
    (
        select
            twitter_id, tweet_id,
            max(name) as name,
            max(profile_image_url) as profile_image_url,
            max(impression_count) as impression_count,
            max(total_engagement) as total_engagement,
            max(tweet_type) as tweet_type,
            max(reply_count) as reply_count,
            max(retweet_count) as retweet_count,
            max(bookmark_count) as bookmark_count,
            max(like_count) as like_count,
            max(quote_count) as quote_count,
            max(tweeted_at) as tweeted_at_time
        from twitter_posts
        where twitter_id in {page_id}
          and toDateTime(tweeted_at, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
        group by twitter_id, tweet_id
        order by tweeted_at_time desc
    )
    group by twitter_id
) as posts_data
on insights_data.twitter_id == posts_data.twitter_id
```

**Key:** Insights uses `argMin(followers_count, saving_time)` -- earliest snapshot per record_id. `first_value()` from ordered result picks most recent record_id's data. Posts dedup via `max()` per `(twitter_id, tweet_id)`. `groupUniqArray(tweet_type)` collects unique tweet types.

---

### 11.5 Engagement and Impression Data (`getEngagementImpressionData`)

**Table:** `twitter_posts`

```sql
select
    twitter_id,
    groupArray(tweet_count) as tweet_count,
    max(name) as name,
    max(profile_picture_url) as profile_picture_url,
    groupArray(impression_count) as impression_count,
    groupArray(total_engagement) as total_engagement,
    groupArray(tweeted_at_date) as tweeted_at_date,
    groupArray(retweet_count) as retweet_count,
    groupArray(reply_count) as reply_count,
    groupArray(like_count) as like_count,
    groupArray(bookmark_count) as bookmark_count,
    groupArray(quote_count) as quote_count
from
(
    select
        twitter_id,
        count(tweet_id) as tweet_count,
        max(name) as name,
        max(profile_picture_url) as profile_picture_url,
        sum(impression_count) as impression_count,
        sum(total_engagement) as total_engagement,
        toDate(tweeted_at_time) as tweeted_at_date,
        sum(retweet_count) as retweet_count,
        sum(reply_count) as reply_count,
        sum(like_count) as like_count,
        sum(bookmark_count) as bookmark_count,
        sum(quote_count) as quote_count
    from
    (
        select
            twitter_id, tweet_id,
            max(name) as name,
            max(profile_image_url) as profile_picture_url,
            max(impression_count) as impression_count,
            max(total_engagement) as total_engagement,
            max(tweeted_at) as tweeted_at_time,
            max(retweet_count) as retweet_count,
            max(reply_count) as reply_count,
            max(like_count) as like_count,
            max(bookmark_count) as bookmark_count,
            max(quote_count) as quote_count
        from twitter_posts
        where twitter_id in {page_id}
          and toDateTime(tweeted_at, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
        group by twitter_id, tweet_id
        ORDER BY tweeted_at_time asc
    )
    group by twitter_id, tweeted_at_date
    order by tweeted_at_date asc
)
group by twitter_id
```

**Three-level aggregation:**
1. Innermost: deduplicate per `(twitter_id, tweet_id)` via `max()`
2. Middle: group by `(twitter_id, tweeted_at_date)` -- daily sums + tweet count
3. Outer: group by `twitter_id` -- `groupArray()` to collect daily arrays

**No zero-fill.** Missing days will not appear in the arrays.

---

### 11.6 Follower Trend (`getFollowerTrend`)

**Table:** `twitter_insights`

```sql
select
    platform_id,
    max(name) as name,
    max(username) as username,
    groupArray(follower_count) as follower_count,
    groupArray(follower_count_daily) as follower_count_daily,
    groupArray(following_count) as following_count,
    groupArray(following_count_daily) as following_count_daily,
    groupArray(saving_date) as buckets
from
(
    select
        record_id, platform_id, name, username,
        follower_count,
        if(runningDifference(follower_count) <= 0, 0, runningDifference(follower_count)) as follower_count_daily,
        following_count,
        if(runningDifference(following_count) <= 0, 0, runningDifference(following_count)) as following_count_daily,
        toDate(bucket_time) as saving_date
    FROM
    (
        select
            record_id,
            max(twitter_id) as platform_id,
            max(name) as name,
            max(username) as username,
            last_value(followers_count) as follower_count,
            last_value(following_count) as following_count,
            max(saving_time) as bucket_time
        from twitter_insights
        WHERE twitter_id in {page_id}
          and toDateTime(saving_time, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
        group by record_id
        order by bucket_time asc
    )
)
group by platform_id
```

**Logic:** Groups by `record_id` first (one record per observation period), `last_value()` gets final count. `runningDifference()` computes day-over-day gain. Negative diffs clamped to 0. **No zero-fill** -- gaps in data will cause missing dates.

---

### 11.7 Top/Least Tweets (`getTweetsData($order_by, $limit, $sort)`)

**Table:** `twitter_posts`

```sql
SELECT
    tweet_id as id,
    tweeted_at,
    last_value(tweet_text) as tweet_text,
    last_value(tweet_type) as tweet_type,
    last_value(permalink) as permalink,
    last_value(media_url) as media_url,
    toInt32(last_value(listed_count)) as listed_count,
    toInt32(last_value(retweet_count)) as retweet_count,
    toInt32(last_value(like_count)) as like_count,
    toInt32(last_value(reply_count)) as reply_count,
    toInt32(last_value(quote_count)) as quote_count,
    toInt32(last_value(bookmark_count)) as bookmark_count,
    toInt32(last_value(impression_count)) as impression_count,
    toInt32(last_value(total_engagement)) as total_engagement
FROM twitter_posts
WHERE twitter_id in {page_id}
  and toDateTime(tweeted_at, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
GROUP BY id, tweeted_at
ORDER BY {order_by} {sort}
LIMIT {limit}
```

**Deduplication:** Groups by `(tweet_id, tweeted_at)`. Uses `last_value()` -- picks most-recently-inserted snapshot's values. `sort` = `'DESC'` for top, `'ASC'` for least.

---

## 12. Cross-Platform Overview V2

### 12.1 Request Parameters

| Parameter | Rule |
|-----------|------|
| `workspace_id` | required |
| `date` | required |
| `timezone` | required (`'Europe/Kyiv'` remapped to `'Europe/Riga'`) |
| `facebook_accounts` | nullable array |
| `instagram_accounts` | nullable array |
| `linkedin_accounts` | nullable array |
| `pinterest_accounts` | nullable array |
| `youtube_accounts` | nullable array |
| `tiktok_accounts` | nullable array |
| `type` | nullable (metric selection) |
| `limit` | nullable (default 20) |

### 12.2 Date Logic

`currentEndDate = parsed_end + 1 day` (inclusive). Secondary period = equal span before current period.

### 12.3 Summary (`getSummaryQuery`)

The most complex query. Builds per-platform sub-queries for current and secondary periods via UNION ALL, then JOINs and computes percentage changes in pure SQL.

**Per-platform column mappings:**

| Platform | Followers Source | Impressions | Engagement | Reach |
|----------|----------------|-------------|------------|-------|
| Facebook | `page_follows` from insights | `avg(page_impressions)` | `sum(page_post_engagements)` | `avg(page_impressions_unique)` |
| Instagram | `followers_count` from insights | `avg(views)` from posts | `sum(engagement)` from posts | `avg(reach)` from posts |
| LinkedIn | `totalFollowerCount` from insights | `avg(impressions)` from posts | `sum(total_engagement)` from posts | `avg(reach)` from posts |
| TikTok | `total_follower_count` from insights | `sum(view_count)` from posts | `sum(engagement_count)` from posts | = impressions |
| Pinterest | `follower_count` from boards | `sum(impression)` from pin_insights | `sum(engagement)` from pin_insights | = impressions |
| YouTube | `subscriber_count` from channels (with fallback) | `sum(views)` from videos | `sum(likes+comments+shares+dislikes)` | = impressions |

#### Outer Shell

```sql
SELECT
    round(followers, 2) AS followers,
    round(posts, 2) AS posts,
    round(engagement, 2) AS engagement,
    round(impressions, 2) AS impressions,
    round(reach, 2) AS reach,
    round(engagement_rate, 2) AS engagement_rate,
    round(secondary_followers, 2) AS secondary_followers,
    round(secondary_posts, 2) AS secondary_posts,
    round(secondary_engagement, 2) AS secondary_engagement,
    round(secondary_impressions, 2) AS secondary_impressions,
    round(secondary_reach, 2) AS secondary_reach,
    round(secondary_engagement_rate, 2) AS secondary_engagement_rate,
    round(if(secondary_followers > 0, followers - secondary_followers, 0.0), 2) AS diff_followers,
    round(if(secondary_posts > 0, posts - secondary_posts, 0.0), 2) AS diff_posts,
    round(if(secondary_engagement > 0, engagement - secondary_engagement, 0.0), 2) AS diff_engagement,
    round(if(secondary_impressions > 0, impressions - secondary_impressions, 0.0), 2) AS diff_impressions,
    round(if(secondary_reach > 0, reach - secondary_reach, 0.0), 2) AS diff_reach,
    round(if(secondary_engagement_rate > 0, engagement_rate - secondary_engagement_rate, 0.0), 2) AS diff_engagement_rate,
    if(secondary_followers > 0, round((followers - secondary_followers) / secondary_followers * 100, 2), 0.0) AS followers_change_pct,
    if(secondary_posts > 0, round((posts - secondary_posts) / secondary_posts * 100, 2), 0.0) AS posts_change_pct,
    if(secondary_engagement > 0, round((engagement - secondary_engagement) / secondary_engagement * 100, 2), 0.0) AS engagement_change_pct,
    if(secondary_impressions > 0, round((impressions - secondary_impressions) / secondary_impressions * 100, 2), 0.0) AS impressions_change_pct,
    if(secondary_reach > 0, round((reach - secondary_reach) / secondary_reach * 100, 2), 0.0) AS reach_change_pct,
    if(secondary_engagement_rate > 0, round((engagement_rate - secondary_engagement_rate) / secondary_engagement_rate * 100, 2), 0.0) AS engagement_rate_change_pct
FROM (
    SELECT
        followers, posts, engagement, impressions, reach,
        if(impressions > 0, round((engagement / impressions) * 100, 2), 0.0) AS engagement_rate,
        secondary_followers, secondary_posts, secondary_engagement, secondary_impressions, secondary_reach,
        if(secondary_impressions > 0, round((secondary_engagement / secondary_impressions) * 100, 2), 0.0) AS secondary_engagement_rate
    FROM (
        SELECT
            sum(followers) AS followers, sum(total_posts) AS posts,
            sum(engagement) AS engagement, sum(impressions) AS impressions, sum(reach) AS reach,
            sum(secondary_followers) AS secondary_followers, sum(secondary_total_posts) AS secondary_posts,
            sum(secondary_engagement) AS secondary_engagement, sum(secondary_impressions) AS secondary_impressions,
            sum(secondary_reach) AS secondary_reach
        FROM (
            WITH
                current_period AS (
                    -- Facebook subquery
                    SELECT fb_insights.page_id,
                        toFloat64(ifNull(fb_insights.followers_count, 0)) AS followers_count,
                        toFloat64(ifNull(fb_posts.post_count, 0)) AS post_count,
                        toFloat64(ifNull(fb_insights.engagement, 0)) AS engagement,
                        toFloat64(ifNull(fb_insights.impressions, 0)) AS impressions,
                        toFloat64(ifNull(fb_insights.reach, 0)) AS reach
                    FROM (
                        SELECT page_id, first_value(followers_count) AS followers_count,
                            avg(page_impressions) AS impressions, avg(page_impressions_unique) AS reach,
                            sum(page_post_engagements) AS engagement
                        FROM (
                            SELECT page_id, max(page_follows) AS followers_count,
                                max(page_impressions) AS page_impressions,
                                max(page_impressions_unique) AS page_impressions_unique,
                                max(page_post_engagements) AS page_post_engagements
                            FROM facebook_insights
                            WHERE page_id IN {facebook_ids} AND {current_date_filter}
                            GROUP BY page_id, toDate(created_time)
                        ) GROUP BY page_id
                    ) AS fb_insights
                    LEFT JOIN (
                        SELECT page_id, count(post_id) AS post_count
                        FROM (SELECT page_id, post_id FROM facebook_posts
                              WHERE page_id IN {facebook_ids} AND {current_date_filter}
                              GROUP BY page_id, post_id)
                        GROUP BY page_id
                    ) AS fb_posts ON fb_insights.page_id = fb_posts.page_id

                    UNION ALL

                    -- Instagram subquery
                    SELECT ig_insights.page_id,
                        toFloat64(ifNull(ig_insights.followers_count, 0)) AS followers_count,
                        toFloat64(ifNull(ig_posts.post_count, 0)) AS post_count,
                        toFloat64(ifNull(ig_posts.engagement, 0)) AS engagement,
                        toFloat64(ifNull(ig_posts.impressions, 0)) AS impressions,
                        toFloat64(ifNull(ig_posts.reach, 0)) AS reach
                    FROM (
                        SELECT instagram_id AS page_id, first_value(followers_count) AS followers_count
                        FROM (SELECT instagram_id, followers_count FROM instagram_insights
                              WHERE instagram_id IN {instagram_ids} AND {current_date_filter}
                              ORDER BY created_time DESC)
                        GROUP BY instagram_id
                    ) AS ig_insights
                    LEFT JOIN (
                        SELECT page_id, count(post_id) AS post_count,
                            sum(engagement) AS engagement, avg(impressions) AS impressions, avg(reach) AS reach
                        FROM (
                            SELECT instagram_id AS page_id, media_id AS post_id,
                                first_value(engagement) AS engagement, first_value(views) AS impressions,
                                first_value(reach) AS reach
                            FROM (SELECT instagram_id, media_id, engagement, views, reach
                                  FROM instagram_posts WHERE instagram_id IN {instagram_ids}
                                  AND {current_date_filter} ORDER BY post_created_at DESC)
                            GROUP BY page_id, post_id
                        ) GROUP BY page_id
                    ) AS ig_posts ON ig_insights.page_id = ig_posts.page_id

                    UNION ALL

                    -- LinkedIn subquery
                    SELECT li_insights.linkedin_id AS page_id,
                        toFloat64(ifNull(li_insights.followers_count, 0)) AS followers_count,
                        toFloat64(ifNull(li_posts.post_count, 0)) AS post_count,
                        toFloat64(ifNull(li_posts.engagement, 0)) AS engagement,
                        toFloat64(ifNull(li_posts.impressions, 0)) AS impressions,
                        toFloat64(ifNull(li_posts.reach, 0)) AS reach
                    FROM (
                        SELECT linkedin_id, first_value(followers_count) AS followers_count
                        FROM (SELECT linkedin_id, totalFollowerCount AS followers_count
                              FROM linkedin_insights WHERE linkedin_id IN {linkedin_ids}
                              AND {current_date_filter} ORDER BY created_at DESC)
                        GROUP BY linkedin_id
                    ) AS li_insights
                    LEFT JOIN (
                        SELECT linkedin_id, count(post_id) AS post_count,
                            sum(total_engagement) AS engagement, avg(impressions) AS impressions,
                            avg(reach) AS reach
                        FROM (
                            SELECT linkedin_id, post_id, first_value(total_engagement) AS total_engagement,
                                first_value(impressions) AS impressions, first_value(reach) AS reach
                            FROM (SELECT linkedin_id, post_id, total_engagement, impressions, reach
                                  FROM linkedin_posts WHERE linkedin_id IN {linkedin_ids}
                                  AND {current_date_filter} ORDER BY published_at DESC)
                            GROUP BY linkedin_id, post_id
                        ) GROUP BY linkedin_id
                    ) AS li_posts ON li_insights.linkedin_id = li_posts.linkedin_id

                    UNION ALL

                    -- TikTok subquery
                    SELECT tt_insights.page_id,
                        toFloat64(ifNull(tt_insights.followers_count, 0)) AS followers_count,
                        toFloat64(ifNull(tt_posts.post_count, 0)) AS post_count,
                        toFloat64(ifNull(tt_posts.engagement, 0)) AS engagement,
                        toFloat64(ifNull(tt_posts.impressions, 0)) AS impressions,
                        toFloat64(ifNull(tt_posts.impressions, 0)) AS reach
                    FROM (
                        SELECT tiktok_id AS page_id, first_value(followers_count) AS followers_count
                        FROM (SELECT tiktok_id, total_follower_count AS followers_count
                              FROM tiktok_insights WHERE tiktok_id IN {tiktok_ids}
                              AND {current_date_filter} ORDER BY inserted_at DESC)
                        GROUP BY tiktok_id
                    ) AS tt_insights
                    LEFT JOIN (
                        SELECT page_id, count(post_id) AS post_count,
                            sum(engagement) AS engagement, sum(impressions) AS impressions
                        FROM (
                            SELECT tiktok_id AS page_id, post_id,
                                first_value(engagement_count) AS engagement,
                                first_value(view_count) AS impressions
                            FROM (SELECT tiktok_id, post_id, engagement_count, view_count
                                  FROM tiktok_posts WHERE tiktok_id IN {tiktok_ids}
                                  AND {current_date_filter} ORDER BY inserted_at DESC)
                            GROUP BY page_id, post_id
                        ) GROUP BY page_id
                    ) AS tt_posts ON tt_insights.page_id = tt_posts.page_id

                    UNION ALL

                    -- Pinterest subquery
                    SELECT pt_insights.board_id AS page_id,
                        toFloat64(ifNull(pt_insights.followers_count, 0)) AS followers_count,
                        toFloat64(ifNull(pt_posts.post_count, 0)) AS post_count,
                        toFloat64(ifNull(pt_posts.engagement, 0)) AS engagement,
                        toFloat64(ifNull(pt_posts.impressions, 0)) AS impressions,
                        toFloat64(ifNull(pt_posts.impressions, 0)) AS reach
                    FROM (
                        SELECT board_id, first_value(followers_count) AS followers_count
                        FROM (SELECT board_id, follower_count AS followers_count
                              FROM pinterest_boards WHERE board_id IN {pinterest_ids}
                              AND {current_date_filter} ORDER BY inserted_at DESC)
                        GROUP BY board_id
                    ) AS pt_insights
                    LEFT JOIN (
                        SELECT board_id, count(pin_id) AS post_count,
                            sum(engagement) AS engagement, sum(impression) AS impressions
                        FROM (
                            SELECT board_id, pin_id, sum(engagement) AS engagement, sum(impression) AS impression
                            FROM (
                                SELECT board_id, pin_id, record_id,
                                    first_value(engagement) AS engagement, first_value(impression) AS impression
                                FROM (
                                    WITH pins AS (
                                        SELECT pin_id, board_id FROM (
                                            SELECT board_id, pin_id FROM pinterest_pins
                                            WHERE board_id IN {pinterest_ids} AND {current_date_filter}
                                            ORDER BY inserted_at DESC
                                        ) GROUP BY board_id, pin_id
                                    )
                                    SELECT p.*, i.engagement, i.impression, i.record_id, i.inserted_at
                                    FROM pins AS p
                                    LEFT JOIN (
                                        SELECT pin_id, record_id, engagement, impression, inserted_at
                                        FROM pinterest_pin_insights
                                        WHERE pin_id IN (SELECT pin_id FROM pins)
                                        AND {current_date_filter}
                                    ) AS i ON i.pin_id = p.pin_id
                                    ORDER BY inserted_at DESC
                                ) GROUP BY board_id, pin_id, record_id
                            ) GROUP BY board_id, pin_id
                        ) GROUP BY board_id
                    ) AS pt_posts ON pt_insights.board_id = pt_posts.board_id

                    UNION ALL

                    -- YouTube subquery
                    SELECT yt_insights.channel_id AS page_id,
                        toFloat64(ifNull(yt_insights.followers_count, 0)) AS followers_count,
                        toFloat64(ifNull(yt_posts.post_count, 0)) AS post_count,
                        toFloat64(ifNull(yt_posts.engagement, 0)) AS engagement,
                        toFloat64(ifNull(yt_posts.impressions, 0)) AS impressions,
                        toFloat64(ifNull(yt_posts.impressions, 0)) AS reach
                    FROM (
                        SELECT channel_id,
                            if(count_in_range > 0, followers_in_range, followers_latest) AS followers_count
                        FROM (
                            SELECT channel_id,
                                countIf({current_date_filter}) AS count_in_range,
                                argMaxIf(subscriber_count, created_at, {current_date_filter}) AS followers_in_range,
                                argMax(subscriber_count, created_at) AS followers_latest
                            FROM youtube_channels
                            WHERE channel_id IN {youtube_ids}
                            GROUP BY channel_id
                        )
                    ) AS yt_insights
                    LEFT JOIN (
                        SELECT channel_id, count(post_id) AS post_count,
                            sum(likes)+sum(comments)+sum(shares)+sum(dislikes) AS engagement,
                            sum(impressions) AS impressions
                        FROM (
                            SELECT channel_id, video_id AS post_id,
                                first_value(likes) AS likes, first_value(comments) AS comments,
                                first_value(shares) AS shares, first_value(dislikes) AS dislikes,
                                first_value(views) AS impressions
                            FROM (SELECT channel_id, video_id, likes, comments, shares, dislikes, views
                                  FROM youtube_videos WHERE channel_id IN {youtube_ids}
                                  AND {current_date_filter} ORDER BY inserted_at DESC)
                            GROUP BY channel_id, post_id
                        ) GROUP BY channel_id
                    ) AS yt_posts ON yt_insights.channel_id = yt_posts.channel_id
                ),
                secondary_period AS (
                    -- Same 6 platform subqueries with {secondary_date_filter} instead of {current_date_filter}
                    -- Structure is identical, only date range changes
                )
            SELECT
                cp.page_id,
                ifNull(cp.followers_count, 0.0) AS followers,
                ifNull(cp.post_count, 0.0)      AS total_posts,
                ifNull(cp.engagement, 0.0)      AS engagement,
                ifNull(cp.impressions, 0.0)     AS impressions,
                ifNull(cp.reach, 0.0)           AS reach,
                ifNull(sp.followers_count, 0.0) AS secondary_followers,
                ifNull(sp.post_count, 0.0)      AS secondary_total_posts,
                ifNull(sp.engagement, 0.0)      AS secondary_engagement,
                ifNull(sp.impressions, 0.0)     AS secondary_impressions,
                ifNull(sp.reach, 0.0)           AS secondary_reach
            FROM current_period AS cp
            LEFT JOIN secondary_period AS sp ON cp.page_id = sp.page_id
        ) AS per_page
    ) AS totals
) AS aggregated
```

#### Per-Platform Sub-Queries (example: YouTube with follower fallback)

```sql
SELECT
    yt_insights.channel_id AS page_id,
    toFloat64(ifNull(yt_insights.followers_count, 0)) AS followers_count,
    toFloat64(ifNull(yt_posts.post_count, 0)) AS post_count,
    toFloat64(ifNull(yt_posts.engagement, 0)) AS engagement,
    toFloat64(ifNull(yt_posts.impressions, 0)) AS impressions,
    toFloat64(ifNull(yt_posts.impressions, 0)) AS reach,
    'youtube' AS platform
FROM (
    SELECT
        channel_id,
        if(count_in_range > 0, followers_in_range, followers_latest) AS followers_count
    FROM (
        SELECT
            channel_id,
            countIf(toDateTime(created_at) BETWEEN ...) as count_in_range,
            argMaxIf(subscriber_count, created_at, toDateTime(created_at) BETWEEN ...) as followers_in_range,
            argMax(subscriber_count, created_at) as followers_latest
        FROM youtube_channels
        WHERE channel_id IN {$youtube_ids}
        GROUP BY channel_id
    )
) AS yt_insights
LEFT JOIN (
    SELECT channel_id, count(post_id) AS post_count,
        sum(likes)+sum(comments)+sum(shares)+sum(dislikes) AS engagement,
        sum(impressions) AS impressions
    FROM (
        SELECT channel_id, video_id as post_id,
            first_value(likes) as likes, first_value(comments) as comments,
            first_value(shares) as shares, first_value(dislikes) as dislikes,
            first_value(views) AS impressions
        FROM (
            SELECT channel_id, video_id, likes, comments, shares, dislikes, views
            FROM youtube_videos
            WHERE channel_id IN {$youtube_ids}
            AND toDateTime(published_at) BETWEEN ...
            ORDER BY inserted_at DESC
        )
        GROUP BY channel_id, post_id
    )
    GROUP BY channel_id
) AS yt_posts ON yt_insights.channel_id = yt_posts.channel_id
```

**YouTube fallback:** If no data in range (`count_in_range = 0`), uses most recent subscriber count.

---

### 12.4 Top Performing Graph (`getTopPerformingGraphQuery`)

Uses materialized view `mv_social_daily_metrics` with `uniqMerge` and `sumMerge`.

```sql
WITH date_range AS (
    SELECT toDate(addDays(toDate('{currentStart}'), number)) AS date
    FROM numbers(dateDiff('day', toDate('{currentStart}'), toDate('{currentEnd}')) + 1)
),
platforms AS (
    SELECT platform FROM (
        SELECT 'facebook' AS platform UNION ALL SELECT 'instagram' UNION ALL SELECT 'linkedin'
        UNION ALL SELECT 'tiktok' UNION ALL SELECT 'youtube' UNION ALL SELECT 'pinterest'
    )
),
date_platform_combinations AS (
    SELECT dr.date, p.platform
    FROM date_range dr
    CROSS JOIN platforms p
),
actual_data AS (
    SELECT
        date,
        toString(platform) as platform,
        uniqMerge(posts_count) AS post_cnt,
        sumMerge(engagement_sum) AS eng_cnt,
        sumMerge(impressions_sum) AS impr_cnt,
        sumMerge(reach_sum) AS reach_cnt
    FROM mv_social_daily_metrics
    WHERE toDateTime(date) BETWEEN toDateTime('{currentStart}',0) AND toDateTime('{currentEnd}',0)
      AND account_id IN {all_accounts}
    GROUP BY date, platform
),
daily AS (
    SELECT
        dpc.date, dpc.platform,
        coalesce(ad.post_cnt, 0)  AS post_cnt,
        coalesce(ad.eng_cnt, 0)   AS eng_cnt,
        coalesce(ad.impr_cnt, 0)  AS impr_cnt,
        coalesce(ad.reach_cnt, 0) AS reach_cnt
    FROM date_platform_combinations dpc
    LEFT JOIN actual_data ad ON dpc.date = ad.date AND dpc.platform = ad.platform
)
SELECT
    arraySort(arrayDistinct(groupArray(date))) AS buckets,
    groupArrayIf(post_cnt,  platform = 'facebook')  AS facebook_post_count,
    groupArrayIf(post_cnt,  platform = 'instagram') AS instagram_post_count,
    groupArrayIf(post_cnt,  platform = 'linkedin')  AS linkedin_post_count,
    groupArrayIf(post_cnt,  platform = 'tiktok')    AS tiktok_post_count,
    groupArrayIf(post_cnt,  platform = 'youtube')   AS youtube_post_count,
    groupArrayIf(post_cnt,  platform = 'pinterest') AS pinterest_post_count,
    groupArrayIf(eng_cnt,   platform = 'facebook')  AS facebook_engagement_count,
    groupArrayIf(eng_cnt,   platform = 'instagram') AS instagram_engagement_count,
    groupArrayIf(eng_cnt,   platform = 'linkedin')  AS linkedin_engagement_count,
    groupArrayIf(eng_cnt,   platform = 'tiktok')    AS tiktok_engagement_count,
    groupArrayIf(eng_cnt,   platform = 'youtube')   AS youtube_engagement_count,
    groupArrayIf(eng_cnt,   platform = 'pinterest') AS pinterest_engagement_count,
    groupArrayIf(impr_cnt,  platform = 'facebook')  AS facebook_impression_count,
    -- ... same pattern for all platforms and metrics ...
    groupArrayIf(reach_cnt, platform = 'facebook')  AS facebook_reach_count
    -- ... same pattern for all platforms ...
FROM daily
```

**Zero-fill:** `CROSS JOIN` of date_range x platforms ensures every (date, platform) pair exists. `coalesce(..., 0)` fills missing values. `arraySort(arrayDistinct(...))` deduplicates and sorts buckets.

---

### 12.5 Platform Data (`getPlatformDataQuery`) -- Grouped by Platform

```sql
WITH fb_posts as (
    SELECT post_id, max(saving_time)
    FROM facebook_posts WHERE page_id in {$facebook_ids} AND date_filter
    group by post_id
),
followers AS (
    SELECT sum(followers_count) as followers_count, platform_type
    FROM (
        -- Facebook: last_value(page_fans) grouped by page_id
        -- Instagram: last_value(followers_count) from instagram_insights grouped by instagram_id
        -- LinkedIn: last_value(totalFollowerCount) from linkedin_insights grouped by linkedin_id
        -- TikTok: last_value(total_follower_count) from tiktok_insights grouped by tiktok_id
        -- Pinterest: last_value(follower_count) from pinterest_boards grouped by board_id
        -- YouTube: if(count_in_range>0, followers_in_range, followers_latest) from youtube_channels
    )
    GROUP BY platform_type
)
SELECT
    toInt32(sum(followers_count)) as followers,
    toInt32(sum(total_posts)) as total_posts,
    toInt32(sum(total_engagements)) as engagement,
    toInt32(sum(total_impression)) as impressions,
    toInt32(sum(total_reach)) as reach,
    toInt32(sum(total_reactions)) as reactions,
    toInt32(sum(total_comments)) as comments,
    toInt32(sum(total_shares)) as shares,
    platform_type
FROM followers
LEFT JOIN (
    -- Facebook: (reactions+comments+shares), post_impressions, post_impressions_unique
    -- Instagram: engagement, views(=impressions), reach, like_count(=reactions), comments_count, saved(=shares)
    -- LinkedIn: total_engagement, impressions, reach, favorites(=reactions), comments, repost(=shares)
    -- TikTok: engagement_count, view_count(=impressions and reach), like_count, comments_count, share_count
    -- YouTube: likes+comments+shares+dislikes(=engagement), views, likes(=reactions), comments, shares
    -- Pinterest: engagement, impression, pin_clicks(=reactions), outbound_click(=comments), saves(=shares)
) AS posts_data ON posts_data.platform_type = followers.platform_type
GROUP BY platform_type
ORDER BY platform_type DESC
```

**Column mapping quirks:**
- Instagram: `saved` maps to `shares` slot
- Pinterest: `pin_clicks` = reactions, `outbound_click` = comments, `saves` = shares
- TikTok: `reach` = 0 hardcoded in graph queries

---

### 12.6 Account Data Detailed (`getAccountDataDetailedQuery`)

FULL JOIN of current and old (secondary) period data with percentage changes:

```sql
WITH
    current_data AS ( [per-account metrics with current DateFilter] ),
    old_data AS ( [per-account metrics with SecondaryDateFilter] )
SELECT
    COALESCE(cd.platform_type, od.platform_type) AS platform_type,
    COALESCE(cd.account_id, od.account_id) AS account_id,
    cd.account_name AS account_name,
    toInt32(COALESCE(cd.followers, 0))   AS current_followers,
    toInt32(COALESCE(od.followers, 0))   AS old_followers,
    toInt32(COALESCE(cd.total_posts, 0)) AS current_posts,
    toInt32(COALESCE(od.total_posts, 0)) AS old_posts,
    toInt32(COALESCE(cd.engagement, 0))  AS current_engagement,
    toInt32(COALESCE(od.engagement, 0))  AS old_engagement,
    toInt32(COALESCE(cd.impressions, 0)) AS current_impressions,
    toInt32(COALESCE(od.impressions, 0)) AS old_impressions,
    toInt32(COALESCE(cd.reach, 0))       AS current_reach,
    toInt32(COALESCE(od.reach, 0))       AS old_reach,
    IF(od.followers=0, 0, round(((cd.followers-od.followers)*100.0)/od.followers,2)) AS followers_change_pct,
    IF(od.total_posts=0, 0, round(((cd.total_posts-od.total_posts)*100.0)/od.total_posts,2)) AS posts_change_pct,
    IF(od.engagement=0, 0, round(((cd.engagement-od.engagement)*100.0)/od.engagement,2)) AS engagement_change_pct,
    IF(od.impressions=0, 0, round(((cd.impressions-od.impressions)*100.0)/od.impressions,2)) AS impressions_change_pct,
    IF(od.reach=0, 0, round(((cd.reach-od.reach)*100.0)/od.reach,2)) AS reach_change_pct
FROM current_data AS cd
FULL JOIN old_data AS od ON cd.account_id = od.account_id AND cd.platform_type = od.platform_type
ORDER BY platform_type DESC
```

---

### 12.7 Account Data Graphs (`getAccountDataGraphsQuery`)

Per-account time-series arrays:

```sql
WITH aggregated_data AS (
    SELECT date, account_id,
           SUM(engagement) AS engagement, SUM(reach) AS reach,
           SUM(impressions) AS impressions, SUM(total_posts) AS total_posts
    FROM (
        -- Facebook: toDate(created_time), page_id as account_id, last_value(total+comments+shares)
        -- Instagram: toDate(post_created_at), instagram_id, last_value(engagement/reach/views)
        -- LinkedIn: toDate(created_at), linkedin_id, first_value(total_engagement/reach/impressions)
        -- YouTube: toDate(published_at), channel_id, last_value(likes+comments+shares+dislikes)
        -- TikTok: toDate(created_at), tiktok_id, last_value(engagement_count/view_count), reach=0
        -- Pinterest: pins CTE + LEFT JOIN pin_insights, reach=0, impressions=0
    )
    GROUP BY date, account_id
    ORDER BY date ASC
),
final_data AS (
    SELECT account_id, date,
           max(engagement) AS engagement, max(reach) AS reach,
           max(impressions) AS impressions, max(total_posts) AS total_posts
    FROM aggregated_data GROUP BY account_id, date
    ORDER BY account_id ASC, date ASC
)
SELECT
    account_id,
    groupArray(engagement)  AS engagement,
    groupArray(reach)       AS reach,
    groupArray(impressions) AS impressions,
    groupArray(total_posts) AS posts,
    groupArray(date)        AS buckets
FROM final_data
GROUP BY account_id
ORDER BY engagement DESC
```

**Note:** TikTok `reach` = 0 hardcoded, Pinterest `reach` and `impressions` = 0 hardcoded.

---

### 12.8 Top Posts (`getTopPostsQuery`)

Unified schema across all 6 platforms via UNION ALL:

```sql
SELECT
    platform_type, account_id, post_id,
    toInt32(reactions) AS likes,
    toInt32(comment_count) AS comments,
    toInt32(shares_count) AS shares,
    toInt32(saves) AS saves,
    toInt32(pin_clicks) AS pin_clicks,
    toInt32(outbound_clicks) AS outbound_clicks,
    toInt32(dislikes_count) AS dislikes_count,
    permalink, media_type, thumbnail, category,
    created_time,
    toInt32(total_engagement) AS total_engagement,
    toInt32(views) AS views,
    toInt32(total_impressions) AS reach
FROM all_posts
ORDER BY {$this->type} DESC
LIMIT {$this->limit}
```

**Per-platform column mapping:**

| Column | Facebook | Instagram | LinkedIn | TikTok | YouTube | Pinterest |
|---|---|---|---|---|---|---|
| reactions/likes | `total` | `like_count` | `favorites` | `like_count` | `likes` | 0 |
| comment_count | `comments` | `comments_count` | `comments` | `comments_count` | `comments` | 0 |
| shares_count | `shares` | 0 | `repost` | `share_count` | `shares` | 0 |
| saves | 0 | `saved` | 0 | 0 | 0 | `saves` |
| pin_clicks | 0 | 0 | 0 | 0 | 0 | `pin_clicks` |
| outbound_clicks | 0 | 0 | 0 | 0 | 0 | `outbound_click` |
| dislikes_count | 0 | 0 | 0 | 0 | `dislikes` | 0 |
| total_engagement | reactions+comments+shares | likes+comments+saved | favorites+comments+repost | likes+comments+shares | likes+comments+shares+dislikes | saves+pin_clicks+outbound_click |
| views | 0 | `views` | 0 | `view_count` | `views` | 0 |
| total_impressions/reach | `post_impressions` | `reach` | `reach` | `view_count` | `views` | `impression` |
| permalink | `permalink` | `permalink` | `article_url` | `share_url` | extracted from `iframe_embed_html` | `https://www.pinterest.com/pin/{pin_id}` |
| thumbnail | `full_picture` | `media_url[1]` | `image` | `embed_link` | `thumbnail_url` | `cover_image_url` |
| category | `caption` | `caption` | `title` | `post_description` | `description` | `description` |

**Instagram:** `media_type != 'STORY'` -- Stories excluded from top posts.

**YouTube permalink:**
```sql
REPLACE(
    concat('https://', substring(iframe_embed_html, position('//' IN iframe_embed_html)+2, ...)),
    'embed/', 'watch?v='
)
```

**Facebook Reels:** Separate CTE pre-selects `(post_id, max(saving_time))` to ensure latest snapshot.

---

## 13. Campaign and Label Analytics

### 13.1 Two-Phase Lookup Pattern

1. Resolve post IDs from MongoDB (`campaign_analytics`/`label_analytics` collections)
2. Query ClickHouse with those IDs

**MongoDB resolution flow:**
- `getCampaignPostIdsForSummary`: Queries `campaign_analytics` collection by campaign IDs
- `getLabelPostIdsForSummary`: Same pattern with `label_analytics`
- If campaign/label not found in MongoDB: fetches plan IDs from `PlansRepository`, resolves social post IDs, creates MongoDB records

### 13.2 Constructor and Date Handling

```php
$this->allPostedIds = "['" . implode("','", $payload['all_post_ids']) . "']";  // ClickHouse array literal
$this->currentEndDate = Carbon::parse($dateArray[1])->addDay();  // +1 day inclusive
// Previous period computed by setPreviousDate():
$previousDate = date_sub(date_create($date[0]), $dateDifference)->format('Y-m-d')
    . ' - ' . date_create($date[0])->format('Y-m-d');
```

### 13.3 Fully Qualified Table Names

This builder uses `contentstudiobackend.` database prefix for all tables, unlike other builders: `contentstudiobackend.facebook_posts`, `contentstudiobackend.instagram_posts`, etc.

### 13.4 Post ID CTE Methods

#### `inMemoryTableForOptimization()` -- Virtual in-memory lookup
```sql
SELECT 'campaign_id_1' AS id, 'post_id_a' AS post_id
UNION ALL SELECT 'campaign_id_1' AS id, 'post_id_b' AS post_id
...
```

#### `selectPostIdsForOptimization()` -- Array-based CTE (more efficient)
```sql
WITH pairs AS (
    select [
        ('campaign_id_1','post_id_a'),
        ('campaign_id_1','post_id_b'),
        ('campaign_id_2','post_id_c'),
        ...
    ] as pairs_array
),
postIds AS (
    SELECT pair.1 AS id, pair.2 AS post_id
    FROM pairs ARRAY JOIN pairs_array as pair
)
```

### 13.5 Summary (`getSummaryQuery`)

```sql
WITH postIds AS (
    SELECT arrayJoin(['post_id_1','post_id_2',...]) AS post_id
),
facebook_reels AS (
    SELECT last_value(post_id) as post_id
    FROM contentstudiobackend.facebook_video_insights
    WHERE video_id IN (SELECT post_id FROM postIds)
    GROUP BY video_id
)

SELECT
    toInt32(sum(total_posts)) as total_posts,
    toInt32(sum(total_engagements)) as total_engagement,
    toInt32(sum(total_impression)) as total_impressions,
    if(total_impressions!=0, round(total_engagement/total_impressions, 2), 0) as total_engagement_rate_per_impression
FROM (
    -- Facebook (if flagSetup['facebook'])
    SELECT
        toInt32(count()) as total_posts,
        toInt32(sum(total_engagement)) as total_engagements,
        toInt32(sum(total_impressions)) as total_impression
    FROM (
        SELECT last_value(total_engagement) as total_engagement,
               last_value(post_impressions) as total_impressions
        FROM contentstudiobackend.facebook_posts
        WHERE (post_id IN (SELECT post_id FROM postIds) OR post_id IN facebook_reels)
          AND toDateTime(created_time) BETWEEN ...
        GROUP BY post_id
    )

    UNION ALL

    -- Instagram (if flagSetup['instagram'])
    SELECT toInt32(count()), toInt32(sum(engagement)), toInt32(sum(views))
    FROM (
        SELECT last_value(engagement), last_value(views)
        FROM contentstudiobackend.instagram_posts
        WHERE media_id IN (SELECT post_id FROM postIds) AND date_filter
        GROUP BY media_id
    )

    UNION ALL

    -- LinkedIn (if flagSetup['linkedin'])
    SELECT toInt32(count()), toInt32(sum(total_engagement)), toInt32(sum(impressions))
    FROM (
        SELECT last_value(total_engagement), last_value(impressions)
        FROM contentstudiobackend.linkedin_posts
        WHERE activity IN (SELECT post_id FROM postIds) AND date_filter
        GROUP BY activity
    )

    UNION ALL

    -- TikTok (if flagSetup['tiktok'])
    SELECT toInt32(count()), toInt32(sum(total_engagement)), toInt32(sum(view_count))
    FROM (
        SELECT last_value(engagement_count) as total_engagement, last_value(view_count)
        FROM contentstudiobackend.tiktok_posts
        WHERE post_id IN (SELECT post_id FROM postIds) AND date_filter
        GROUP BY post_id
    )

    UNION ALL

    -- YouTube (if flagSetup['youtube'] -- hardcoded to true)
    SELECT toInt32(count()), toInt32(sum(total_engagement)), toInt32(sum(views))
    FROM (
        SELECT argMax(likes,inserted_at)+argMax(comments,inserted_at)
              +argMax(shares,inserted_at)+argMax(dislikes,inserted_at) as total_engagement,
               argMax(views,inserted_at) as views
        FROM contentstudiobackend.youtube_videos
        WHERE video_id IN (SELECT post_id FROM postIds) AND date_filter
        GROUP BY video_id
    )

    UNION ALL

    -- Pinterest (if flagSetup['pinterest'])
    WITH pins AS (
        SELECT pin_id FROM contentstudiobackend.pinterest_pins
        WHERE pin_id IN (SELECT post_id FROM postIds) AND date_filter
        GROUP BY pin_id
    ),
    pinterest_insights AS (
        SELECT pin_id, toInt32(SUM(engagement)) AS engagement, toInt32(SUM(impression)) AS impression
        FROM (
            SELECT pin_id, last_value(engagement), last_value(impression)
            FROM contentstudiobackend.pinterest_pin_insights
            WHERE pin_id in (SELECT post_id FROM postIds)
            GROUP BY record_id, pin_id
        )
        GROUP BY pin_id
    )
    SELECT
        toInt32(count()) as total_posts,
        toInt32(sum(total_engagement)) as total_engagements,
        toInt32(sum(total_impressions)) as total_impression
    FROM (
        SELECT toInt32(SUM(pinterest_insights.engagement)) AS total_engagement,
               toInt32(SUM(pinterest_insights.impression)) AS total_impressions
        FROM pins LEFT JOIN pinterest_insights ON pins.pin_id = pinterest_insights.pin_id
        GROUP BY pins.pin_id
    )
)
```

**Key:** Platform inclusion is conditional via `$flagSetup`. Trailing `UNION ALL` removed via `substr($query, 0, -10)`. Uses `arrayJoin()` on array literal for post IDs (different from breakdown which uses `ARRAY JOIN`).

---

### 13.6 Breakdown Data (`getBreakdownData($period)`)

Per-campaign/label totals for a given period:

```sql
WITH pairs AS (...), postIds AS (...),
facebook_reels AS (
    SELECT video_id, post_id
    FROM contentstudiobackend.facebook_video_insights
    WHERE video_id IN (SELECT post_id FROM postIds)
    GROUP BY video_id, post_id
),
pinterest_insights AS (
    SELECT pin_id,
        toInt32(SUM(engagement)) AS engagement,
        toInt32(SUM(impression)) AS impression
    FROM (
        SELECT pin_id,
               argMax(engagement, inserted_at) as engagement,
               argMax(impression, inserted_at) as impression
        FROM contentstudiobackend.pinterest_pin_insights
        WHERE pin_id in (SELECT post_id FROM postIds)
        GROUP BY record_id, pin_id
    )
    GROUP BY pin_id
)

SELECT
    id,
    COALESCE('{$period}', '{$period}') as era,
    COALESCE(toInt32(count()), 0) as total_posts,
    COALESCE(toInt32(sum(total_engagement)), 0) as total_engagement,
    COALESCE(toInt32(sum(total_impressions)), 0) as total_impressions
FROM (
    -- Facebook Reels (video posts via video_insights join)
    SELECT
        facebook_reels.video_id as post_id,
        sum(total_engagement) as total_engagement,
        sum(total_impressions) as total_impressions
    FROM (
        SELECT post_id,
            toFloat64(argMax(total_engagement, saving_time)) as total_engagement,
            toFloat64(argMax(post_impressions, saving_time)) as total_impressions
        FROM contentstudiobackend.facebook_posts
        WHERE post_id in (SELECT post_id FROM facebook_reels)
          AND toDateTime(created_time) BETWEEN toDateTime('{start}', 0) AND toDateTime('{end}', 0)
        GROUP BY post_id
    ) AS reels
    LEFT JOIN facebook_reels ON reels.post_id = facebook_reels.post_id
    GROUP BY post_id

    UNION ALL

    -- Facebook regular posts
    SELECT post_id,
        toFloat64(argMax(total_engagement, saving_time)) AS total_engagement,
        toFloat64(argMax(post_impressions, saving_time)) AS total_impressions
    FROM contentstudiobackend.facebook_posts
    WHERE post_id IN (SELECT post_id FROM postIds) AND date_filter
    GROUP BY post_id

    UNION ALL

    -- Instagram
    SELECT media_id AS post_id,
        toFloat64(argMax(engagement, stored_event_at)) AS total_engagement,
        toFloat64(argMax(views, stored_event_at)) AS total_impressions
    FROM contentstudiobackend.instagram_posts
    WHERE media_id IN (SELECT post_id FROM postIds) AND date_filter
    GROUP BY media_id

    UNION ALL

    -- LinkedIn
    SELECT activity AS post_id,
        toFloat64(argMax(total_engagement, saving_time)) AS total_engagement,
        toFloat64(argMax(impressions, saving_time)) AS total_impressions
    FROM contentstudiobackend.linkedin_posts
    WHERE activity IN (SELECT post_id FROM postIds) AND date_filter
    GROUP BY activity

    UNION ALL

    -- TikTok
    SELECT post_id,
        toFloat64(argMax(engagement_count, inserted_at)) AS total_engagement,
        toFloat64(argMax(view_count, inserted_at)) AS total_impressions
    FROM contentstudiobackend.tiktok_posts
    WHERE post_id IN (SELECT post_id FROM postIds) AND date_filter
    GROUP BY post_id

    UNION ALL

    -- YouTube
    SELECT video_id AS post_id,
        toFloat64(argMax(likes,inserted_at) + argMax(comments,inserted_at)
                + argMax(shares,inserted_at) + argMax(dislikes,inserted_at)) AS total_engagement,
        toFloat64(argMax(views, inserted_at)) AS total_impressions
    FROM contentstudiobackend.youtube_videos
    WHERE video_id IN (SELECT post_id FROM postIds) AND date_filter
    GROUP BY video_id

    UNION ALL

    -- Pinterest
    SELECT pins.pin_id as post_id,
        toFloat64(SUM(pinterest_insights.engagement)) AS total_engagement,
        toFloat64(SUM(pinterest_insights.impression)) AS total_impressions
    FROM (
        SELECT pin_id FROM contentstudiobackend.pinterest_pins
        WHERE pin_id IN (SELECT post_id FROM postIds) AND date_filter
        GROUP BY pin_id
    ) AS pins
    LEFT JOIN pinterest_insights ON pins.pin_id = pinterest_insights.pin_id
    GROUP BY pins.pin_id

) AS all_posts
LEFT JOIN postIds ON postIds.post_id = all_posts.post_id
GROUP BY postIds.id
```

**Key patterns:**
- Uses `argMax(value, timestamp)` for most metrics -- takes value at maximum timestamp
- `era` column is always `$period` ('current' or 'previous') -- allows merging in PHP
- Facebook Reels: resolved via `facebook_video_insights.video_id` -> `facebook_posts.post_id`
- Pinterest: pre-built CTE with `argMax(engagement/impression, inserted_at) GROUP BY record_id, pin_id`

---

### 13.7 Insights Time-Series (`getInsightsData`)

Per-campaign/label time-series:

```sql
WITH pairs AS (...), postIds AS (...),
facebook_reels AS (...),
pinterest_insights AS (...)

SELECT
    id,
    groupArray(total_engagement) as total_engagement,
    groupArray(total_impressions) as total_impressions,
    groupArray(total_posts) as total_posts,
    groupArray(created_at) as created_at
FROM (
    SELECT
        postIds.id as id,
        toInt32(sum(total_engagement)) as total_engagement,
        toInt32(sum(total_impressions)) as total_impressions,
        toInt32(count()) as total_posts,
        toDate(created_at) as created_at
    FROM (
        -- Facebook Reels
        SELECT facebook_reels.video_id as post_id,
            sum(total_engagement) as total_engagement,
            sum(total_impressions) as total_impressions,
            min(created_time) as created_at
        FROM (
            SELECT post_id,
                toFloat64(argMax(total_engagement, saving_time)) as total_engagement,
                toFloat64(argMax(post_impressions, saving_time)) as total_impressions,
                min(created_time) as created_time
            FROM contentstudiobackend.facebook_posts
            WHERE post_id in (SELECT post_id FROM facebook_reels)
              AND toDateTime(created_time) BETWEEN toDateTime('{start}', 0) AND toDateTime('{end}', 0)
            GROUP BY post_id
        ) AS reels
        LEFT JOIN facebook_reels ON reels.post_id = facebook_reels.post_id
        GROUP BY post_id

        UNION ALL

        -- Facebook regular posts
        SELECT post_id,
            toFloat64(argMax(total_engagement, saving_time)) AS total_engagement,
            toFloat64(argMax(post_impressions, saving_time)) AS total_impressions,
            min(created_time) as created_at
        FROM contentstudiobackend.facebook_posts
        WHERE post_id IN (SELECT post_id FROM postIds) AND date_filter
        GROUP BY post_id

        UNION ALL

        -- Instagram
        SELECT media_id AS post_id,
            toFloat64(argMax(engagement, stored_event_at)) AS total_engagement,
            toFloat64(argMax(views, stored_event_at)) AS total_impressions,
            min(post_created_at) as created_at
        FROM contentstudiobackend.instagram_posts
        WHERE media_id IN (SELECT post_id FROM postIds) AND date_filter
        GROUP BY media_id

        UNION ALL

        -- LinkedIn
        SELECT activity AS post_id,
            toFloat64(argMax(total_engagement, saving_time)) AS total_engagement,
            toFloat64(argMax(impressions, saving_time)) AS total_impressions,
            min(created_at) as created_at
        FROM contentstudiobackend.linkedin_posts
        WHERE activity IN (SELECT post_id FROM postIds) AND date_filter
        GROUP BY activity

        UNION ALL

        -- TikTok
        SELECT post_id,
            toFloat64(argMax(engagement_count, inserted_at)) AS total_engagement,
            toFloat64(argMax(view_count, inserted_at)) AS total_impressions,
            min(created_at) as created_at
        FROM contentstudiobackend.tiktok_posts
        WHERE post_id IN (SELECT post_id FROM postIds) AND date_filter
        GROUP BY post_id

        UNION ALL

        -- YouTube
        SELECT video_id AS post_id,
            toFloat64(argMax(likes, inserted_at) + argMax(comments, inserted_at)
                     + argMax(shares, inserted_at) + argMax(dislikes, inserted_at)) AS total_engagement,
            toFloat64(argMax(views, inserted_at)) AS total_impressions,
            min(published_at) as created_at
        FROM contentstudiobackend.youtube_videos
        WHERE video_id IN (SELECT post_id FROM postIds) AND date_filter
        GROUP BY video_id

        UNION ALL

        -- Pinterest
        SELECT pins.pin_id as post_id,
            toFloat64(SUM(pinterest_insights.engagement)) AS total_engagement,
            toFloat64(SUM(pinterest_insights.impression)) AS total_impressions,
            min(pins.created_at) as created_at
        FROM (
            SELECT pin_id, min(created_at) as created_at
            FROM contentstudiobackend.pinterest_pins
            WHERE pin_id IN (SELECT post_id FROM postIds) AND date_filter
            GROUP BY pin_id
        ) AS pins
        LEFT JOIN pinterest_insights ON pins.pin_id = pinterest_insights.pin_id
        GROUP BY pins.pin_id
    ) AS all_posts
    LEFT JOIN postIds ON all_posts.post_id = postIds.post_id
    GROUP BY created_at, postIds.id
    ORDER BY created_at DESC
)
GROUP BY id
```

Returns one row per campaign/label ID with arrays of `total_engagement[]`, `total_impressions[]`, `total_posts[]`, and `created_at[]`.

---

### 13.8 Planner Analytics (`getAnalyticsForPlannerQuery`)

Returns detailed per-post metrics for a single platform. Each platform returns metric-specific columns with tooltip strings:

**Facebook:** `engagement`, `impressions`, `reach`, `comments`, `repost` (shares), `post_clicks`, `reactions` (total), plus individual reaction types (`likes`, `love`, `wow`, `haha`, `sad`, `anger`), `media_type`

**Instagram:** `engagement`, `impressions`, `reach`, `likes` (like_count), `comments` (comments_count), `saves` (saved), `media_type` (CAROUSEL_ALBUM -> 'Carousel')

**LinkedIn:** `engagement`, `impressions`, `reach`, `comments`, `reactions` (favorites), `reposts` (repost), `post_clicks`, `media_type`

**TikTok:** `engagement` (engagement_count), `views` (view_count), `likes`, `comments`, `shares`, `engagement_rate`

**YouTube:** `engagement` (likes+comments+shares+dislikes), `views`, `likes`, `dislikes`, `comments`, `shares`, `red_views`, `subscribers_gained`, `minutes_watched`, `red_minutes_watched`, `average_view_duration`, `media_type`

**Pinterest:** `engagement`, `impressions`, `pin_clicks`, `outbound_clicks`, `saves`, `engagement_rate`

### 13.9 Summary Diff Calculation

```php
foreach ($current as $key => $value) {
    $prevValue = $previous[$key] ?? 'N/A';
    $res['difference'][$key] = ($value !== 'N/A' && $prevValue !== 'N/A')
        ? round(intval($value) - intval($prevValue), 2) : 'N/A';
    $res['percentage'][$key] = ($value !== 'N/A' && $prevValue !== 'N/A')
        ? round((intval($value) - intval($prevValue)) * 100 / max($prevValue, 1), 2) : 'N/A';
}
```

`max($prevValue, 1)` prevents divide-by-zero when previous = 0.

---

## 14. Date Handling Patterns

### 14.1 Platform-Specific Date Filters

| Platform | Posts Filter | Insights Filter |
|----------|-------------|----------------|
| **Facebook** | `toDateTime(created_time, 0, '{tz}') BETWEEN ...` | `toDate(created_time) BETWEEN ...` |
| **Instagram** | `toDateTime(post_created_at, 0, '{tz}') BETWEEN ...` | `toDateTime(created_time, 0, '{tz}') BETWEEN ...` |
| **LinkedIn** | `toDateTime(published_at, 0, '{tz}') BETWEEN ...` | `toDateTime(created_at, 0, '{tz}') BETWEEN ...` |
| **YouTube** | `toDateTime(published_at) >= ... AND < ...+1day` | `toDateTime(created_at) >= ... AND < ...+1day` |
| **TikTok** | `toDateTime(created_at, 0, '{tz}') BETWEEN ...` | `toDateTime(inserted_at, 0, '{tz}') BETWEEN ...` |
| **Pinterest** | `toDate(created_at) BETWEEN ...` (tz-naive) | `toDate(created_at) BETWEEN ...` (tz-naive) |
| **Twitter** | `toDateTime(created_at, 0, '{tz}') BETWEEN ...` | `toDateTime(inserted_at, 0, '{tz}') BETWEEN ...` |
| **Competitor** | Dates converted to UTC in PHP before SQL | Same (UTC-converted) |
| **Overview V2** | `toDateTime(field) BETWEEN ...` (no tz) | Same |
| **Campaign/Label** | `toDateTime(field) BETWEEN ...` (no tz) | Same |

### 14.2 DateFilter Side-Effect

Facebook, YouTube, Pinterest, and LinkedIn builders mutate `currentEndDate` within `DateFilter()`. Call only once per query. YouTube adds a day then subtracts back. LinkedIn's `engagementsQuery` calls `subDay()` before building.

---

## 15. Deduplication Strategies

### 15.1 CTE Deduplication Pattern (Facebook, Instagram, LinkedIn)

```sql
WITH posts AS (
    SELECT post_id, max(saving_time)
    FROM {platform}_posts
    WHERE {filters}
    GROUP BY post_id
)
SELECT ...
FROM {platform}_posts
WHERE (post_id, saving_time) IN (posts)
```

### 15.2 `last_value` Deduplication (YouTube, TikTok, Twitter)

```sql
SELECT
    last_value(field) as field
FROM {table}
GROUP BY record_id
```

### 15.3 `argMax`/`argMin` Deduplication

- `argMin(value, saving_time)` -- earliest record (Facebook insights fan count)
- `argMax(value, inserted_at)` -- latest record (YouTube videos, Campaign/Label)

---

## 16. Zero-Fill Patterns

### 16.1 `WITH FILL` (ClickHouse native)

```sql
ORDER BY date ASC
WITH FILL FROM toDate('{start}') TO toDate('{end}') + 1 STEP 1
```

### 16.2 `arrayFill` (Forward-fill)

```sql
arrayFill(x -> not x == 0, array)
```

Carries last non-zero value forward through zeros.

### 16.3 `arrayReverseFill` (Backward-fill)

```sql
arrayReverseFill(x -> not x == 0, array)
```

Used by YouTube subscriber trend in combination with `arrayFill`.

### 16.4 `INTERPOLATE AS -1` Sentinel (TikTok)

```sql
INTERPOLATE (
    follower_count as -1,
    views_per_day as -1
)
```

Marks gap days with -1. Controller trims leading -1 values, then forward-fills remaining gaps.

### 16.5 Controller-Level Leading-Zero Backfill

When `array[0] == 0`, extends date range 2 years back, fetches last known non-zero value, replaces all leading zeros. Used by: Facebook audience growth, Instagram audience growth, LinkedIn audience growth, YouTube subscriber trend.

### 16.6 No Zero-Fill (Twitter)

Twitter queries have no `WITH FILL` and no gap-filling logic.

---

## 17. Dynamic Aggregation Patterns

### 17.1 Thresholds

| Platform | Threshold | Daily | Monthly |
|----------|-----------|-------|---------|
| Facebook | 60 days | `toDate(field)`, STEP 1 | `toStartOfMonth(field)`, STEP INTERVAL 1 MONTH |
| Instagram | 60 days | Same | Same |
| LinkedIn | 60 days | Same | Same |
| YouTube | 180 days | Same | Same |
| TikTok | 180 days | Same | Same |
| Pinterest | 180 days | Same | Same |

### 17.2 Response Addition

Dynamic queries add `aggregation_level` field ('daily' or 'monthly') to the response.

### 17.3 AI Summary Variants

Facebook has additional AI summary methods that combine all metrics into a single dynamic query with `media_type_data` map array breakdown.

---

## 18. Known Quirks and Bugs

### 18.1 Facebook

- `getDateFilters` has side-effect mutating `currentEndDate` via `subDay()`
- Gender demographics: M/U/F labels mapped to wrong aliases in CASE expressions
- Age demographics: Date boundary clamped to 2024-03-14
- Active users timezone offset: `round(offsetHours) + 8` (UTC-8 storage assumption)

### 18.2 Instagram

- `'18-24'` age bucket aliased as `'18-34'` (copy-paste error)
- `overviewEngagement` reads `previous_date` from request instead of computing it
- `impressionsQuery` uses `views` column as impressions in Overview V2

### 18.3 LinkedIn

- `engagementsQuery` mutates `currentEndDate` via `subDay()` before building
- Demographics use `contentstudiobackend.linkedin_insights` (fully qualified) while all other queries use unqualified table names
- `total_engagement` excludes clicks (unlike the engagement_rate formula that works on impressions)
- Multi-hashtag filter appears buggy (comma-joined into single `has()` call)

### 18.4 YouTube

- Column typo: `non_subsciber_watch_time` (missing 's' in subscriber)
- `findVideoQuery` uses 13 traffic sources (excludes shorts_views) while `videoViewsTrendQuery` uses all 14
- `getEngagementRollup` has `DateFilter` called twice (duplicate filter bug)
- `avg_view_duration` computed differently in `summaryQuery` vs `sumSummaryQuery`
- `engagement_rate` in top/least posts = `engagement / count(*)` (records count, not views)
- Timezone parameter stored but never applied in any SQL query

### 18.5 TikTok

- Uses `INTERPOLATE AS -1` sentinel pattern (unique to TikTok)
- `runningDifference` can produce negative values; clamped to 0

### 18.6 Pinterest

- `'Europe/Kyiv'` remapped to `'Europe/Riga'` as ClickHouse workaround
- Date filtering is timezone-naive (`toDate()` not `toDateTime()`)
- Board mode uses `CROSS JOIN` (valid only because both sides produce one row)
- `engagement_rate` = `(pin_clicks + outbound_clicks + saves) / count(*)` (not impressions-based)

### 18.7 Twitter/X

- Copy-paste bug: response key `tiktok_id` holds Twitter account ID
- No zero-fill in any query
- Settings methods use MongoDB, not ClickHouse

### 18.8 Overview V2

- `currentEndDate = parsed_end + 1 day`
- TikTok/Pinterest/YouTube: `reach = impressions` (same field)
- Instagram uses `views` as impressions
- TikTok `reach` = 0 hardcoded in `getAccountDataGraphsQuery`
- Pinterest `reach` = 0 and `impressions` = 0 hardcoded in `getAccountDataGraphsQuery`

### 18.9 Campaign/Label

- Uses fully qualified `contentstudiobackend.` table names
- LinkedIn uses `activity` column as post_id (not `post_id`)
- `setPreviousDate` uses PHP `date_diff` which returns `d` component only (not total days) -- may be incorrect for multi-month ranges
- `getSummaryQuery` uses `arrayJoin` while `getBreakdownData` uses `ARRAY JOIN` with pairs -- different optimization approaches

---

## 19. Cross-Platform Overview V1

**Controller:** `OverviewController`
**File:** `app/Http/Controllers/Analytics/Analyze/OverviewController.php`

The V1 overview aggregates data across all platforms using per-platform builders and returns combined metrics.

### 19.1 Route Map

| Route | Method |
|-------|--------|
| `POST /analytics/overview/summary` | `overview` |
| `POST /analytics/overview/topPosts` | `topPosts` |
| `POST /analytics/overview/postsEngagement` | `postsEngagement` |
| `POST /analytics/overview/engagementRollup` | `engagementRollup` |
| `POST /analytics/overview/accountPerformance` | `accountPerformance` |
| `POST /analytics/overview/timeRecommendation` | `timeRecommendation` |

### 19.2 Request Parameters (Universal)

| Parameter | Type | Description |
|-----------|------|-------------|
| `workspace_id` | string | Required workspace ID |
| `date` | string | `"YYYY-MM-DDTHH:mm:ss - YYYY-MM-DDTHH:mm:ss"` (ISO 8601) |
| `timezone` | string | IANA timezone |
| `facebook_accounts` | array | Facebook page IDs |
| `instagram_accounts` | array | Instagram account IDs |
| `linkedin_accounts` | array | LinkedIn page IDs |
| `youtube_accounts` | array | YouTube channel IDs |
| `tiktok_accounts` | array | TikTok account IDs |
| `pinterest_accounts` | array | Pinterest board IDs |
| `previous_date` | string | Optional previous period for comparison |

### 19.3 Overview Summary (`overview`)

Aggregates summary metrics from all enabled platforms for current and optional previous period.

**Per-Platform Data Sources:**
- **Facebook**: `FacebookController` (Elasticsearch indices)
- **Instagram**: `InstagramBuilder.publishRollupQuery()` + `summaryQuery()` (ClickHouse)
- **LinkedIn**: `LinkedinBuilder.summaryQuery()` (ClickHouse)
- **YouTube**: `YoutubeBuilder.sumSummaryQuery()` + `summaryQuery()` (ClickHouse)
- **TikTok**: `TiktokBuilder.getPageAndPostsInsights()` (ClickHouse)

**Response:**
```json
{
  "status": true,
  "ids": {
    "facebook": [], "instagram": [], "linkedin": [],
    "youtube": [], "tiktok": [], "pinterest": [], "twitter": []
  },
  "overview": {
    "current": {
      "total_posts": 0, "reposts": 0, "comments": 0, "reactions": 0
    },
    "previous": {
      "total_posts": 0, "reposts": 0, "comments": 0, "reactions": 0
    }
  }
}
```

### 19.4 Engagement Rollup (`engagementRollup`)

Calculates per-day engagement and posting rates across all platforms.

**Calculation:**
```
total_days_current = (end_date - start_date).days + 1
total_days_previous = (end_date - start_date).days     // no +1 for previous
engagement_per_day = total_engagement / total_days
posts_per_day = total_posts / total_days
```

**Response:**
```json
{
  "engagement_rollup": {
    "current": {
      "total_engagement": 0, "engagement_per_day": 0.0,
      "total_posts": 0, "posts_per_day": 0.0
    },
    "previous": { "..." }
  }
}
```

### 19.5 Top Posts (`topPosts`)

Fetches top 15 posts per platform sorted by total engagement, then merges and sorts overall.

**Per-Platform Limits:** 15 posts each
**Overall Sort:** `array_multisort()` by engagement descending

**Response:**
```json
{
  "top_posts": {
    "overall": [{"post_id": "", "total_engagement": 0, "network": "facebook", "...": ""}],
    "facebook": [], "instagram": [], "linkedin": [],
    "youtube": [], "tiktok": [], "pinterest": [], "twitter": []
  }
}
```

### 19.6 Posts Engagement Time Series (`postsEngagement`)

Aggregates daily engagement across all platforms into time-series buckets.

**Response:**
```json
{
  "posts_engagements": {
    "buckets": ["YYYY-MM-DD"],
    "data": {
      "total_engagement": [], "comments": [],
      "reactions": [], "reposts": [], "post_count": []
    },
    "show_data": 0
  }
}
```

`show_data` = sum of all engagement + post counts. Used by frontend to decide whether to render the chart.

### 19.7 Account Performance (`accountPerformance`)

Returns per-platform and per-account engagement breakdown.

**Metrics by Platform:**

| Platform | Specific Metrics |
|----------|-----------------|
| Facebook | comments, reactions, reposts, post_clicks, total_engagement, total_posts |
| Instagram | comments, reactions, reposts, saved, total_engagement, total_posts |
| LinkedIn | comments, reactions, reposts, total_engagement, total_posts |
| YouTube | comments, reactions, reposts, total_engagement, total_posts |
| TikTok | comments, reactions, reposts, total_engagement, total_posts |
| Pinterest | pin_clicks, outbound_clicks, saves, total_engagement, followers |

### 19.8 Time Recommendation (`timeRecommendation`)

Best time to post heatmap using last 3 months of data.

**Request:**
```json
{
  "state": "merged",
  "accounts": [{"facebook_id": "", "instagram_id": "", "linkedin_id": ""}]
}
```

**OverviewBuilder Query Pattern:**
- Subtracts 3 months from start date for historical window
- Groups by `day_of_week` (0-6) and `hour_of_day` (0-23)
- Calculates engagement score: `(impressions + engagement) / post_count`
- Applies timezone offset to DateHistogram

**Per-Platform Fields:**

| Platform | ID Field | Date Field | Engagement Field | Impression Field |
|----------|----------|------------|-----------------|-----------------|
| Facebook | `page_id` | `created_time` | `total_engagement` | `post_impressions` |
| Instagram | `instagram_id` | `post_created_at` | `engagement` | `views` |
| LinkedIn | `linkedin_id` | `published_at` | `total_engagement` | (none, hardcoded 0) |

**Response:**
```json
{
  "data": [{
    "facebook": {"page_id": {"day_of_week": {"hour_of_day": 0.0}}},
    "instagram": {"...": "..."},
    "linkedin": {"...": "..."},
    "merged": {"day_of_week": {"hour_of_day": 0.0}}
  }]
}
```

When `state = "merged"`, scores are averaged across all platforms into a single heatmap.

---

## 20. Dashboard Analytics

**Controller:** `DashboardAnalytics`
**File:** `app/Http/Controllers/Analytics/Analytics/DashboardAnalytics.php`

Dashboard statistics for content publishing, approval workflows, and inbox management.

### 20.1 Content Publishing Stats

**Route:** `POST /getContentPublishingStats`

**Request:**
```json
{
  "workspace_id": "string (required)",
  "date_range": "YYYY-MM-DD - YYYY-MM-DD (required)"
}
```

**Business Logic:**
- Applies permission filtering based on user role
  - Collaborators: filtered to authorized accounts with `dashboard=true`
  - Admins: `dashboard_admin=true` for full access
- Calls `PlansRepository::fetchPlansCounts()` with workspace/date/permission filters

**Response:**
```json
{
  "status": true,
  "stats": {
    "scheduled": 0,
    "published": 0,
    "partial": 0,
    "failed": 0
  }
}
```

### 20.2 Approval Publishing Stats

**Route:** `POST /getContentApprovalStats`

**Request:** Same as content publishing (`workspace_id`, `date_range`)

**Response:**
```json
{
  "status": true,
  "stats": {
    "review": 0,
    "rejected": 0,
    "missed": 0
  }
}
```

### 20.3 Inbox Stats

**Route:** `POST /getInboxStats`

**Request:**
```json
{
  "workspace_id": "string (required)"
}
```

**Business Logic:**
- Retrieves conversation counts across 4 states
- Supports Facebook, Instagram, LinkedIn, GMB platforms
- Uses `InboxFiltersBuilder` with `ConversationFlag` filter per state

**Response:**
```json
{
  "status": true,
  "stats": {
    "UNASSIGNED": 0,
    "ASSIGNED": 0,
    "MARKED_AS_DONE": 0,
    "MINE": 0
  }
}
```

---

## 21. Analytics Share Link Management

**Controller:** `AnalyticsShareLinkController`
**File:** `app/Http/Controllers/Analytics/AnalyticsShareLinkController.php`
**MongoDB Collection:** `analytics_share_links`

Enables creating, managing, and sharing analytics dashboards via generated links with controlled access.

### 21.1 Data Model

```javascript
{
  "_id": ObjectId,
  "link_id": "unique_generated_uid",
  "title": "string (max 255)",
  "workspace_id": "string",
  "user_id": "string",
  "platform": "overview|facebook|instagram|linkedin|tiktok|youtube|pinterest|twitter",
  "is_date_range_fixed": boolean,
  "date_range": { "from": "YYYY-MM-DD", "to": "YYYY-MM-DD" } || null,
  "is_account_switching_enabled": boolean,
  "is_password_protected": boolean,
  "password": "string" || null,
  "account_id": "string" || null,
  "overview_accounts": array || null,
  "is_disabled": boolean,
  "created_by": "string",
  "updated_by": "string",
  "created_at": ISODate,
  "updated_at": ISODate
}
```

**Relationships:** `belongsTo User` via `created_by` field

### 21.2 Create Share Link

**Route:** `POST /analytics/share-link/create`

**Request:**
```json
{
  "title": "string (required, max 255)",
  "workspace_id": "string (required)",
  "platform": "string (required, one of: overview|facebook|instagram|linkedin|tiktok|youtube|pinterest|twitter)",
  "is_date_range_fixed": "boolean (optional, default: false)",
  "date_range": "string (required if is_date_range_fixed is true)",
  "is_account_switching_enabled": "boolean (optional, default: true)",
  "is_password_protected": "boolean (optional, default: false)",
  "password": "string (required if is_password_protected, min: 4 chars)",
  "account_id": "string (optional)",
  "overview_accounts": "array (optional)"
}
```

**Business Logic:**
1. Generates unique `link_id` via `Helper::generateUid()`
2. Stores `user_id` from authenticated user
3. Sets `is_disabled = false`
4. Constructs share URL: `{app_url}share/analytics/{link_id}` (WhiteLabelHelper)

**Response:**
```json
{
  "status": "success",
  "data": { "share_url": "https://app.example.com/share/analytics/{link_id}" },
  "message": "Share link created successfully"
}
```

### 21.3 Update Share Link

**Route:** `POST /analytics/share-link/update/{id}`

Updates title, date range settings, account switching, password settings. Sets `updated_by` to current user.

### 21.4 Toggle Share Link State

**Route:** `PUT /analytics/share-link/update/toggle-state/{id}`

**Request:** `{ "is_disabled": boolean }`

Enables/disables a share link without deleting it.

### 21.5 Fetch Share Links

**Route:** `GET /analytics/share-link/list/{workspace_id}`

Returns all share links for workspace, sorted by `created_at DESC`, with eager-loaded user relationship (`firstname`, `lastname`).

### 21.6 Delete Share Link

**Route:** `DELETE /analytics/share-link/delete/{id}`

Hard deletes the share link document from MongoDB.

### 21.7 Get Share Link Details (Public)

**Route:** `GET /analytics/shared/{link_id}`

Retrieves share link by `link_id` (not `_id`). Returns full document. Used by public share page.

### 21.8 Verify Share Link Password

**Route:** `POST /analytics/shared/{link_id}/verify-password`

**Request:** `{ "password": "string (required)" }`

**Business Logic:**
1. If share link is NOT password protected: returns success without verification
2. If password protected: direct string comparison (`$shareLink->password === $request->input('password')`)

**Note:** Passwords are stored in plain text (not hashed).

---

## 22. Reports and Scheduled Reports

### 22.1 Analytics Reports

**Controller:** `AnalyticsReports`
**File:** `app/Http/Controllers/Analytics/Analytics/AnalyticsReports.php`
**MongoDB Collection:** `reports`

#### Data Model

```javascript
{
  "_id": ObjectId,
  "workspace_id": "string",
  "user_id": "string",
  "name": "string",
  "type": "single-pdf-detailed|single-pdf-overview|multiple-pdf-overview|multiple-pdf-detailed|group|competitor",
  "platform_type": "facebook|instagram|linkedin|...",
  "accounts": [],
  "date": "YYYY-MM-DD - YYYY-MM-DD",
  "source": "scheduled|manual|export" || null,
  "status": "pending|processing|completed|failed",
  "progress": 0-100,
  "export_url": "https://s3-bucket.../report.pdf" || null,
  "expire_time": datetime || null,
  "execution_id": "string" || null,
  "schedule_id": ObjectId || null,
  "language": "en|es|fr|...",
  "labels": [],
  "campaigns": [],
  "topPosts": integer,
  "error": boolean || null,
  "error_message": "string" || null,
  "created_at": datetime,
  "updated_at": datetime
}
```

#### Store Report

**Route:** `POST /analytics/reports/save`

**Request:**
```json
{
  "type": "string (required: single-pdf-detailed|single-pdf-overview|multiple-pdf-overview|multiple-pdf-detailed|group|competitor)",
  "action": "save|render|email (optional, default: save)",
  "workspace_id": "string (required)",
  "name": "string (optional)",
  "platform_type": "string (optional)",
  "accounts": "array (optional)",
  "date": "string (optional)",
  "email_list": "array (optional)",
  "language": "string (optional, default: en)",
  "labels": "array (optional)",
  "campaigns": "array (optional)",
  "topPosts": "integer (optional)"
}
```

**Actions:**
- `save`: Persists report config to MongoDB, returns report object
- `render`: Creates `ExportReportJob` and dispatches to queue
- `email`: Same as render but with `send_email: true` and `email_list`

**ExportReportJob Categories:**
- `single`: Combined report for selected accounts
- `multiple`: Individual account PDFs batched together
- `grouped`: Grouped report across accounts
- `competitor`: Competitor analysis report

#### Show Report

**Route:** `POST /analytics/reports/show`

**Request:** `{ "_id": "ObjectId (required)" }`

Returns single report by ID. Used to check generation progress and get export URL.

#### List Reports

**Route:** `POST /analytics/reports/list`

**Request:** `{ "workspace_id": "string (required)" }`

**Query Filters:**
- Excludes scheduled reports (`source` must be null)
- Excludes `single-pdf-detailed` type
- Limited to last 30 reports, sorted `created_at DESC`
- Competitor reports enriched via MongoDB aggregation pipeline joining `competitors_reports`

#### Remove Report

**Route:** `POST /analytics/reports/remove`

**Request:** `{ "report_id": "string (required)" }`

Hard deletes from MongoDB. No cascading operations.

### 22.2 Scheduled Reports

**Controller:** `ScheduleReports`
**File:** `app/Http/Controllers/Analytics/Analytics/ScheduleReports.php`
**MongoDB Collection:** `schedule_reports`

#### Data Model

```javascript
{
  "_id": ObjectId,
  "workspace_id": "string",
  "user_id": "string",
  "name": "string",
  "type": "group|individual|...",
  "platform_type": "string",
  "accounts": [],
  "email_list": [],
  "report_type": "string",
  "frequency": "daily|weekly|monthly|custom",
  "cron": "string" || null,
  "day_of_week": 0-6 || null,
  "day_of_month": 1-31 || null,
  "time": "HH:MM" || null,
  "timezone": "IANA timezone",
  "next_run_at": datetime,
  "last_run_at": datetime || null,
  "last_execution_id": "string" || null,
  "consecutive_failures": 0,
  "max_attempts": 3,
  "active": true,
  "paused_at": datetime || null,
  "paused_reason": "string" || null,
  "created_at": datetime,
  "updated_at": datetime
}
```

#### Show Scheduled Reports

**Route:** `POST /analytics/reports/schedule/show`

**Request:** `{ "workspace_id": "string (required)" }`

Returns all scheduled reports for workspace.

#### Create/Update Schedule

**Route:** `POST /analytics/reports/schedule/save`

**Request:**
```json
{
  "_id": "string (optional, for updates)",
  "workspace_id": "string (required)",
  "name": "string (required)",
  "type": "string (required)",
  "platform_type": "string (required)",
  "accounts": "array (required)",
  "email_list": "array (required)",
  "frequency": "daily|weekly|monthly|custom (required)",
  "day_of_week": "integer (optional, 0-6)",
  "day_of_month": "integer (optional, 1-31)",
  "time": "HH:MM (optional)",
  "timezone": "string (optional)",
  "cron": "string (optional, overrides other scheduling fields)"
}
```

If `_id` provided, updates existing; otherwise creates new.

**Auto-Pause:** After `max_attempts` consecutive failures, sets `active=false`, records `paused_at` and `paused_reason`.

#### Remove Schedule

**Route:** `POST /analytics/reports/schedule/remove`

**Request:** `{ "report_id": "string (required)" }`

Hard deletes schedule. Execution history is NOT deleted.

#### Send Scheduled Reports (Cron)

**Route:** `POST /analytics/reports/schedule/send`

**Request:** `{ "interval": "daily|weekly|monthly" }`

Called by Laravel scheduler. For each due schedule: creates report record, creates execution history, dispatches `ExportReportJob`, updates `next_run_at` and `last_run_at`.

---

## 23. Analytics Job Triggers

**Controller:** `AnalyticsJobController`
**File:** `app/Http/Controllers/Analytics/Analyze/AnalyticsJobController.php`

### 23.1 Trigger Analytics Job

**Route:** `POST /analytics/triggerJob`

**Request:**
```json
{
  "workspace_id": "string (required)",
  "account_id": "string (required)",
  "platform": "string (required)"
}
```

**Business Logic:**
1. Validates account exists in `SocialIntegrations` by `_id`, `workspace_id`, and `platform_type`
2. Checks cache key `analytics:immediate-job-{platform_identifier}-{platform_type}` to prevent duplicate triggers
3. Calls `AnalyticsHelper::triggerAnalyticsPipeline($platform_type, $account)`
4. Caches trigger for 1 hour (60-minute TTL)

**Response:**
```json
{ "status": true, "message": "Localized message" }
```

### 23.2 Trigger Competitor Job

**Route:** `POST /analytics/triggerCompetitorJob`

**Request:**
```json
{
  "workspace_id": "string (required)",
  "report_id": "string (required)",
  "platform": "string (required)",
  "competitor_ids": "array (required)"
}
```

**Business Logic:**
1. Checks cache key `analytics:competitor-report-job-{report_id}`
2. Fetches competitors from `CompetitorsModel` matching IDs and platform
3. Iterates and calls `CompetitorsRepo::triggerIgCompetitorJob()` for each
4. Caches for 1 hour if all successful

---

## 24. Twitter/X Settings Management

**Controller:** `TwitterController` (settings methods)
**File:** `app/Http/Controllers/Analytics/Analyze/TwitterController.php`
**MongoDB Collection:** `twitter_job_settings`

### 24.1 Create Setting

**Route:** `POST /analytics/settings/twitter/createTwitterAnalyticsSetting`

**Request:**
```json
{
  "platform_id": "string (required, must exist in TwitterAccounts)",
  "workspace_id": "string (required, must exist in Workspace)",
  "updated_by": "string (required, must exist in User)",
  "created_by": "string (required, must exist in User)",
  "job_type": "string (required)",
  "trigger_day": "string (required)",
  "platform_name": "string (required)",
  "post_count": "integer (required)"
}
```

Creates job configuration for automatic Twitter analytics fetching.

### 24.2 Update Setting

**Route:** `POST /analytics/settings/twitter/updateTwitterAnalyticsSetting`

Same parameters as create. Finds by composite key (`platform_id`, `workspace_id`) and updates.

### 24.3 Get Settings

**Route:** `POST /analytics/settings/twitter/getTwitterAnalyticsSetting`

**Request:** `{ "workspace_id": "string (required)" }`

Returns all Twitter job settings for workspace.

### 24.4 Fetch Single Setting

**Route:** `POST /analytics/settings/twitter/fetchTwitterAnalyticsSetting`

Returns a single setting record.

### 24.5 Fetch All Settings

**Route:** `POST /analytics/settings/twitter/fetchTwitterAnalyticsSettings`

Returns all settings across workspaces (admin endpoint).

### 24.6 Delete Setting

**Route:** `POST /analytics/settings/twitter/deleteTwitterAnalyticsSetting`

Removes setting by ID.

### 24.7 Trigger Twitter Analytics Job

**Route:** `POST /analytics/settings/twitter/triggerTwitterAnalyticsJob`

**Request:**
```json
{
  "platform_id": "string (required, must exist in TwitterAccounts)",
  "workspace_id": "string (required)"
}
```

**Business Logic:**
1. Retrieves Twitter account details
2. Constructs Argo workflow API URL: `{ARGO_BASE_URL}/social-account-added-{ARGO_ENV}`
3. Sends payload: `{"channel": "twitter", "account-id": "{account_id}"}`
4. Triggers asynchronous data fetching via Argo Workflows

### 24.8 Create Job Logs

**Route:** `POST /analytics/settings/twitter/createTwitterJobLogs`

**Request:**
```json
{
  "platform_id": "string (required)",
  "workspace_id": "string (required)",
  "platform_type": "string (required)",
  "job_type": "string (required)",
  "credits_used": "integer (required)",
  "executed_by": "string (required)",
  "app_id": "string (required)",
  "app_name": "string (required)",
  "error": "string (required)"
}
```

**MongoDB Collection:** `twitter_jobs_metadata`

Automatically calculates: `job_executed_at` (UTC ISO), `day_of_week` (0-6), `hour_of_day` (0-23). Used for API quota tracking and usage analytics.

### 24.9 Get Credits Used Count

**Route:** `POST /analytics/overview/twitter/getCreditsUsedCount`

Returns total API credits used by the workspace's Twitter analytics jobs.

---

## 25. Competitor Management CRUD

### 25.1 Facebook Competitor Search

**Controller:** `FacebookCompetitorController`
**Route:** `POST /analytics/overview/facebook/competitor/search`

**Request:** `{ "search": "string (required, page name)" }`

**Business Logic:**
1. Retrieves valid Facebook token from Redis: `redis-data-science:facebook_valid_token_set`
2. Calls Facebook Graph API `/pages/search` with fields: `id, name, link, verification_status, location`
3. Implements 3-attempt retry logic
4. Transforms response: renames `id` to `competitor_id`, generates profile picture URL

**Response:**
```json
{
  "data": [{
    "competitor_id": "string",
    "name": "string",
    "link": "string",
    "verification_status": "string",
    "location": {},
    "image": "https://graph.facebook.com/{id}/picture?type=large"
  }]
}
```

### 25.2 Instagram Competitor Search

**Controller:** `InstagramCompetitorController`
**Route:** `POST /analytics/overview/instagram/competitor/search`

**Request:** `{ "search": "string (required, username or URL)" }`

**Business Logic:**
1. Extracts slug from URL if necessary
2. Retrieves valid Instagram token from Redis: `redis-data-science:instagram_valid_token_set`
3. Calls Facebook Graph API `business_discovery` endpoint
4. Decrypts token via `SocialHelper::decryptToken()`
5. Generates `appsecret_proof` for API security

**Response:**
```json
{
  "data": [{
    "biography": "string",
    "competitor_id": "string",
    "id": "string",
    "name": "string",
    "image": "string (profile picture URL)",
    "slug": "string (username)"
  }]
}
```

### 25.3 Add/Update Competitor Report

**Route:** `POST /analytics/overview/instagram/competitor/addUpdateCompetitorReport`

**Request:**
```json
{
  "_id": "ObjectId (optional, for updates)",
  "platform_type": "facebook|instagram",
  "workspace_id": "string",
  "name": "string",
  "created_by_user_id": "string",
  "updated_by_user_id": "string",
  "competitors": [{"competitor_id": "", "name": "", "image": "", "slug": ""}]
}
```

**Business Logic:**
1. Creates/updates report in `competitors_reports` collection
2. Creates/updates individual competitor records in `competitors` collection
3. Logs operation via `LogsBuilder`
4. Triggers background Argo job to fetch competitor analytics data

### 25.4 Get Competitor Report

**Route:** `POST /analytics/overview/instagram/competitor/getCompetitorReport`

**Request:** `{ "_id": "ObjectId (required)" }`

Returns single report with all competitor documents populated.

### 25.5 List Competitor Reports

**Route:** `POST /analytics/overview/instagram/competitor/getCompetitorReportsByWorkspace`

**Request:** `{ "workspace_id": "string", "platform_type": "string" }`

Returns all competitor reports for workspace and platform, each with populated competitor details.

### 25.6 Delete Competitor Report

**Route:** `POST /analytics/overview/instagram/competitor/deleteCompetitorReport`

**Request:** `{ "_id": "ObjectId (required)" }`

Hard deletes report document. Individual competitor records remain (reusable across reports).

---

## 26. AI Insights

**Controllers:** Platform-specific in `app/Http/Controllers/Analytics/AI/`
**AI Service:** `AiAgentService` at `app/Services/AI/AiAgentService.php`

### 26.1 Architecture

Each platform has a dedicated AI Insights controller that:
1. Receives a `type` parameter identifying which insight to generate
2. Checks cache (24-hour TTL) by key: `{platform}_AI:{method}:{account_id}:{date}:{locale}`
3. Calls the corresponding platform controller/builder to gather data
4. Validates data availability (checks array sums for zero values)
5. Sends dataset to external AI Agent service
6. Caches and returns response

**AI Agent Service:**
- Base URL from config `ai_agents.base_url`
- Bearer token auth from config `ai_agents.api_key`
- 120-second timeout
- Injects current locale as `language` parameter
- Endpoint pattern: `{platform}/{insight-type}`

### 26.2 Facebook AI Insights

**Route:** `GET /analytics/overview/facebook/ai_insights`

**Request:**
```json
{
  "workspace_id": "string (required)",
  "date": "YYYY-MM-DD - YYYY-MM-DD (required)",
  "facebook_id": "string (required)",
  "type": "string (required)",
  "timezone": "string (required)",
  "limit": "integer (required)"
}
```

**Insight Types:**

| Type | Builder Method | AI Endpoint | Data Validation |
|------|---------------|-------------|-----------------|
| `page_impressions` | `getImpressionsAIInsights()` | `facebook/page-impressions` | `array_sum(page_impressions) === 0` |
| `page_engagement` | `getEngagementAIInsights()` | `facebook/page-engagement` | `array_sum(page_engagements) === 0` |
| `publishing_behaviour_impressions` | `getPublishingBehaviourAIInsights()` | `facebook/publishing-impressions` | `array_sum(post_count) === 0` |
| `publishing_behaviour_engagements` | `getPublishingBehaviourAIInsights(type)` | `facebook/publishing-engagement` | post count check |
| `publishing_behaviour_reach` | `getPublishingBehaviourAIInsights(type)` | `facebook/publishing-reach` | post count check |
| `audience_growth` | `getAudienceGrowthAIInsights()` | `facebook/audience-growth` | `array_sum(fan_count) === 0` |
| `video_views` | `getVideoAIInsights()` | `facebook/video-views` | `array_sum(total_posts) === 0` |
| `video_watch_time` | `getVideoAIInsights(type)` | `facebook/video-watch-time` | total_posts check |
| `video_engagements` | `getVideoAIInsights(type)` | `facebook/video-engagement` | total_posts check |
| `reels_initial_plays` | `getReelsAIInsights(type)` | `facebook/reels-plays` | `array_sum(total_posts) === 0` |
| `reels_watch_time` | `getReelsAIInsights(type)` | `facebook/reels-watch-time` | total_posts check |
| `reels_engagement` | `getReelsAIInsights(type)` | `facebook/reels-engagement` | total_posts check |
| `top_posts` | `getTopPosts('current')` | `facebook/top-posts` | `count(top_posts) === 0` |
| `insights_summary` | `getCombinedAISummaryData()` + `getTopPosts()` | `facebook/insights-summary` | `array_sum(post_count) === 0` |

**Summary Payload (insights_summary):**
```json
{
  "facebook_page_data": {},
  "reels_data": {},
  "publishing_behaviour": {},
  "facebook_video_data": {},
  "top_posts": [],
  "language": "en"
}
```

### 26.3 Instagram AI Insights

**Route:** `POST /analytics/overview/instagram/ai_insights`

**Insight Types:**

| Type | Builder Method | AI Endpoint |
|------|---------------|-------------|
| `impressions` | `getImpressions()` | `instagram/impressions` |
| `engagement` | `getEngagement()` | `instagram/engagement` |
| `publishing_behaviour_impressions` | `getDynamicPublish()` | `instagram/publishing-impressions` |
| `publishing_behaviour_engagements` | `getDynamicPublish('engagement')` | `instagram/publishing-engagement` |
| `publishing_behaviour_reach` | `getDynamicPublish('reach')` | `instagram/publishing-reach` |
| `audience_growth` | `getAudienceGrowth()` | `instagram/audience-growth` |
| `reels_engagement` | `getReelsDynamic('engagement')` | `instagram/reels-engagement` |
| `reels_watch_time` | `getReelsDynamic('watch_time')` | `instagram/reels-watch-time` |
| `reels_shares` | `getReelsDynamic('shares')` | `instagram/reels-shares` |
| `stories_interactions` | `getStoriesDynamic()` | `instagram/stories-interactions` |
| `stories_impressions` | `getStoriesDynamic()` | `instagram/stories-impressions` |
| `stories_reach` | `getStoriesDynamic()` | `instagram/stories-reach` |
| `top_posts` | `getTopPosts($request)` | `instagram/top-posts` |
| `top_hashtags` | `getHashtags()` | `instagram/hashtags` |
| `insights_summary` | `getSummary()` + `getReels()` + `getStories()` + `getPublish()` + `getTopPosts()` + `getHashtags()` | `instagram/insights-summary` |

### 26.4 YouTube AI Insights

**Route:** `POST /analytics/overview/youtube/ai_insights`

**Additional Request Parameter:** `trend_type` (nullable string, for subscriber and engagement trends)

**Insight Types:**

| Type | Builder Method | AI Endpoint |
|------|---------------|-------------|
| `subscribers_trend` | `overviewDynamicSubscriberTrend()` | `youtube/cumulative-subscribers-trend` |
| `daily_views` | `overviewDynamicViewsTrend()` | `youtube/daily-views` |
| `daily_engagement` | `overviewDynamicEngagementTrend()` | `youtube/daily-engagement` |
| `daily_watch_time` | `overviewDynamicWatchTimeTrend()` | `youtube/daily-watch-time` |
| `viewers_find_videos` | `overviewFindVideo()` | `youtube/traffic-sources` |
| `engagement_vs_posting_pattern` | `overviewPerformanceAndVideoPostingSchedule()` | `youtube/posting-patterns` |
| `sharing_services` | `overviewVideoSharing()` | `youtube/sharing-trends` |
| `top_and_least_posts` | `overviewLeastPosts()` + `overviewTopPosts()` | `youtube/top-least-performing-posts` |
| `insights_summary` | All trend methods + `overviewFindVideo()` + `overviewVideoSharing()` + `overviewLeastPosts()` | `youtube/overview-summary` |

### 26.5 LinkedIn AI Insights

**Route:** `GET /analytics/overview/linkedin/ai_insights`

**Insight Types:** `publishing_behaviour`, `publishing_behaviour_impressions`, `publishing_behaviour_reach`, `audience_growth`, `page_views`, `top_posts`, `top_hashtags`, `city_demographics`, `country_demographics`, `industry_demographics`, `post_density`, `seniority_demographics`, `insights_summary`

### 26.6 TikTok AI Insights

**Route:** `POST /analytics/overview/tiktok/ai_insights`

**Insight Types:** `audience_growth`, `top_posts`, `daily_engagement`, `cumulative_engagement`, `daily_video_views`, `cumulative_video_views`, `engagement_vs_daily_posting`, `insights_summary`

### 26.7 Pinterest AI Insights

**Route:** `POST /analytics/overview/pinterest/ai_insights`

**Insight Types:** `daily_engagement`, `impressions_vs_posting_pattern`, `engagement_vs_posting_pattern`, `daily_pin_posting`, `daily_followers_trend`, `cumulative_followers_trend`, `impressions`, `engagement`, `top_and_least_posts`, `insights_summary`

### 26.8 Overview AI Insights

**Route:** `POST /analytics/overview/ai_insights`

**Insight Types:** `reach_across_platforms`, `engagement_across_platforms`, `impressions_across_platforms`, `platform_performance_comparison`, `overview_account_statistics`, `top_posts`

### 26.9 Error Handling

When data validation fails (zero values):
```json
{ "success": false, "message": "analytics.insufficient_data" }
```

Error message keys: `analytics.insufficient_data`, `analytics.insufficient_data_posts`, `analytics.insufficient_data_videos`, `analytics.insufficient_data_reels`, `analytics.invalid_insight_type`

---

## Appendix: Complete Column Reference by Table

### facebook_posts
`post_id`, `page_id`, `saving_time`, `created_time`, `media_type`, `status_type`, `video_id`, `category`, `published_by`, `published_by_url`, `shared_from_name`, `shared_from_id`, `shared_from_link`, `like`, `love`, `haha`, `wow`, `sad`, `angry`, `total`, `shares`, `comments`, `post_clicks`, `total_engagement`, `post_engaged_users`, `day_of_week`, `hour_of_day`, `updated_time`, `message_tags`, `post_metadata`, `caption`, `description`, `full_picture`, `link`, `permalink`, `post_impressions`, `post_impressions_unique`, `post_impressions_paid`, `post_impressions_paid_unique`, `post_impressions_organic`, `post_impressions_organic_unique`, `post_impressions_viral`, `post_impressions_viral_unique`, `post_video_views`, `total_impressions`

### facebook_insights
`page_id`, `hash_id`, `saving_time`, `created_time`, `page_fans`, `page_follows`, `page_impressions`, `page_impressions_paid`, `page_impressions_organic`, `page_post_engagements`, `page_fans_by_like`, `page_fans_by_unlike`, `talking_about_count`, `positive_sentiment`, `negative_sentiment`, `page_positive_feedback`, `page_negative_feedback`, `page_fans_online`, `page_fans_gender`, `page_fans_age`, `page_fans_gender_age`, `page_fans_country`, `page_fans_city`, `day_of_week`

### instagram_posts
`media_id`, `instagram_id`, `stored_event_at`, `post_created_at`, `media_type`, `entity_type`, `engagement`, `like_count`, `comments_count`, `saved`, `reach`, `impressions`, `views`, `shares`, `hashtags`, `reels_avg_watch_time`, `reels_total_watch_time`, `replies`, `exits`, `taps_forward`, `taps_back`, `permalink`, `caption`, `media_url`

### instagram_insights
`instagram_id`, `record_id`, `stored_event_at`, `created_time`, `online_users_datetime`, `followers_count`, `follows_count`, `profile_views`, `engagement`, `impressions`, `reach`, `accounts_engaged`, `online_followers`, `day_of_week`, `audience_age`, `audience_gender`, `audience_city`, `audience_country`

### linkedin_posts
`post_id`, `activity`, `linkedin_id`, `saving_time`, `published_at`, `created_at`, `media_type`, `day_of_week`, `favorites`, `comments`, `repost`, `post_clicks`, `total_engagement`, `impressions`, `reach`, `hashtags`, `title`, `image`, `article_url`

### linkedin_insights
`linkedin_id`, `record_id`, `inserted_at`, `created_at`, `totalFollowerCount`, `organicFollowerCount`, `paidFollowerCount`, `page_views`, `desktop_page_views`, `mobile_page_views`, `impressionCount`, `engagement`, `reach`, `repost`, `comments`, `reactions`, `unique_visitors`, `followers_by_seniority` (JSON), `followers_by_industry` (JSON), `followers_by_country` (JSON), `followers_by_city` (JSON)

### youtube_activity_insights
`record_id`, `channel_id`, `created_at`, `estimated_minutes_watched`, `average_view_duration`, `views`, `likes`, `dislikes`, `comments`, `shares`

### youtube_channels
`record_id`, `channel_id`, `subscriber_count`, `inserted_at`, `created_at`, `title`

### youtube_videos
`video_id`, `channel_id`, `published_at`, `inserted_at`, `title`, `description`, `duration`, `thumbnail_url`, `media_type`, `iframe_embed_html`, `likes`, `dislikes`, `views`, `red_views`, `favorites`, `comments`, `subscribers_gained`, `shares`, `minutes_watched`, `red_minutes_watched`, `average_view_duration`, `average_view_percentage`

### youtube_traffic_insights
`record_id`, `channel_id`, `created_at`, `subscriber_views`, `subscriber_watch_time`, `non_subsciber_watch_time`, `paid_views`, `annotation_views`, `end_screen_views`, `campaign_card_view`, `no_link_other_views`, `yt_channel_views`, `yt_search_views`, `related_video_views`, `yt_other_page_views`, `ext_url_views`, `playlist_views`, `notification_views`, `shorts_views`

### youtube_shared_insights
`channel_id`, `inserted_at`, `ameba`, `blogger`, `copy_paste`, `cyworld`, `digg`, `dropbox`, `embed`, `mail`, `whats_app`, `other`, `facebook_messenger`, `facebook_pages`, `facebook`, `fotka`, `vkontakte`, `google_plus`, `discord`, `linkedin`, `goo`, `hangouts`, `pinterest`, `myspace`, `reddit`, `skype`, `telegram`, `tumblr`, `twitter`, `viber`, `weibo`, `wechat`, `youtube`

### tiktok_posts
`tiktok_id`, `post_id`, `display_name`, `like_count`, `comments_count`, `share_count`, `engagement_count`, `engagement_rate`, `view_count`, `created_at`, `inserted_at`, `profile_link`, `cover_image_url`, `share_url`, `post_description`, `hashtags`, `duration`, `height`, `width`, `title`, `embed_html`, `embed_link`

### tiktok_insights
`tiktok_id`, `record_id`, `display_name`, `total_follower_count`, `total_following_count`, `total_video_views`, `total_video_likes`, `total_video_comments`, `total_video_shares`, `inserted_at`

### pinterest_pins
`pin_id`, `board_id`, `user_id`, `created_at`, `inserted_at`, `media_type`, `is_owner`, `title`, `description`, `board_owner`, `cover_image_url`, `dominant_color`, `creative_type`, `product_tags`, `height`, `width`

### pinterest_pin_insights
`pin_id`, `record_id`, `user_id`, `created_at`, `saving_time`, `inserted_at`, `impression`, `pin_clicks`, `outbound_click`, `saves`, `engagement`, `quartile_95s_percent_view`, `closeup`, `video_start`, `video_10s_view`, `video_avg_watch_time`

### pinterest_user_insights
`user_id`, `created_at`, `impression`, `pin_clicks`, `outbound_click`, `saves`, `engagement`

### pinterest_users
`user_id`, `inserted_at`, `follower_count`

### pinterest_boards
`board_id`, `user_id`, `inserted_at`, `follower_count`, `name`

### twitter_posts
`twitter_id`, `post_id`, `created_at`, `inserted_at`, `tweet_type`, `total_engagement`, `impressions`, `like_count`, `reply_count`, `retweet_count`, `quote_count`, `bookmark_count`, `url_link_clicks`, `user_profile_clicks`, `impression_count`, `hashtags`, `permalink`, `full_text`

### twitter_insights
`twitter_id`, `record_id`, `inserted_at`, `followers_count`, `following_count`, `tweet_count`, `listed_count`


### 10.4 `summaryQueryForUser()`

**Tables:** `pinterest_user_insights`, `pinterest_users`

```sql
SELECT
    if(count > 0, toString(followers), 'N/A') as follower_count,
    if(count_insights > 0, toString(impressions), 'N/A') as impressions,
    if(count_insights > 0, toString(pin_clicks), 'N/A') as pin_clicks,
    if(count_insights > 0, toString(outbound_clicks), 'N/A') as outbound_clicks,
    if(count_insights > 0, toString(saves), 'N/A') as saves,
    if(count_insights > 0, toString(engagement), 'N/A') as total_engagement
FROM
(
    SELECT
        user_id,
        toInt32(count()) as count_insights,
        toInt32(sum(impressions)) as impressions,
        toInt32(sum(pin_clicks)) as pin_clicks,
        toInt32(sum(outbound_clicks)) as outbound_clicks,
        toInt32(sum(saves)) as saves,
        toInt32(sum(engagement)) as engagement
    FROM (
        SELECT
            user_id,
            max(impression) as impressions,
            max(pin_clicks) as pin_clicks,
            max(outbound_click) as outbound_clicks,
            max(saves) as saves,
            max(engagement) as engagement,
            toDate(created_at) as saving_time
        FROM pinterest_user_insights
        WHERE user_id in {pinterest_id}
          AND toDate(created_at) BETWEEN toDate('{startDate}') AND toDate('{endDate+1}')
        GROUP BY saving_time, user_id
    )
    GROUP BY user_id
) AS pin_insights
LEFT JOIN (
    SELECT first_value(follower_count) as followers,
           user_id,
           toInt32(count()) as count
    FROM pinterest_users
    WHERE user_id in {pinterest_id}
      AND toDate(inserted_at) BETWEEN toDate('{startDate}') AND toDate('{endDate+1}')
    GROUP BY user_id
) as pin_data USING user_id
```

---

### 10.5 `summaryQueryForBoard()`

**Tables:** `pinterest_pins`, `pinterest_boards`, `pinterest_pin_insights`

```sql
WITH pins AS (
    SELECT pin_id
    FROM pinterest_pins
    WHERE board_id IN {board_id}
    GROUP BY pin_id
)
SELECT board_id,
    if(count_board_data > 0, toString(follower_count), 'N/A') as follower_count,
    if(count > 0, toString(impressions), 'N/A') as impressions,
    if(count > 0, toString(pin_clicks), 'N/A') as pin_clicks,
    if(count > 0, toString(outbound_clicks), 'N/A') as outbound_clicks,
    if(count > 0, toString(saves), 'N/A') as saves,
    if(count > 0, toString(engagement), 'N/A') as total_engagement
FROM
(
    SELECT board_id,
           count() as count_board_data,
           toInt32(last_value(follower_count)) as follower_count
    FROM pinterest_boards
    WHERE board_id IN {board_id}
      AND toDate(inserted_at) BETWEEN toDate('{startDate}') AND toDate('{endDate+1}')
    GROUP BY board_id
) AS board_data
CROSS JOIN
(
    SELECT
        toInt32(count()) as count,
        toInt32(sum(impressions)) as impressions,
        toInt32(sum(pin_clicks)) as pin_clicks,
        toInt32(sum(outbound_clicks)) as outbound_clicks,
        toInt32(sum(saves)) as saves,
        toInt32(sum(engagement)) as engagement
    FROM (
        SELECT
            max(impression) as impressions,
            max(pin_clicks) as pin_clicks,
            max(outbound_click) as outbound_clicks,
            max(saves) as saves,
            max(engagement) as engagement,
            toDate(created_at) as saving_time
        FROM pinterest_pin_insights
        WHERE pin_id in (SELECT pin_id FROM pins)
          AND toDate(saving_time) BETWEEN toDate('{startDate}') AND toDate('{endDate+1}')
        GROUP BY pin_id, saving_time
    ) as pi
) as pin_metrics
```

**Note:** Uses `CROSS JOIN` -- valid because both sides produce exactly one row.

---

### 10.6 Followers Queries

#### `followersQueryForUser()`

**Table:** `pinterest_users`

```sql
SELECT
    arrayDifference(followers_total) as followers_daily,
    arrayFill((x) -> not x==0, followers_total) as followers_gained,
    show_data,
    buckets
FROM
(
    SELECT groupArray(followers) as followers_total,
           groupArray(saving_time) as buckets,
           toInt32(sum(followers)) as show_data
    FROM
    (
        SELECT toInt32(argMin(follower_count, inserted_at)) as followers,
               toDate(inserted_at) as saving_time
        FROM pinterest_users
        WHERE user_id in {pinterest_id}
          AND toDate(inserted_at) BETWEEN toDate('{startDate}') AND toDate('{endDate+1}')
        GROUP BY saving_time
        ORDER BY saving_time ASC
        WITH FILL
        TO toDate('{endDate}')
        STEP 1
    )
)
```

**Logic:** `argMin(follower_count, inserted_at)` picks earliest snapshot's follower count per day. `arrayDifference` computes daily gains. `arrayFill(x -> not x==0, ...)` forward-fills zeros.

#### `followersQueryForBoard()`

**Table:** `pinterest_boards`

```sql
-- Same structure, filters by user_id and board_id in pinterest_boards
SELECT toInt32(argMin(follower_count, inserted_at)) as followers,
       toDate(inserted_at) as saving_time
FROM pinterest_boards
WHERE user_id in {pinterest_id}
  AND board_id in {board_id}
  AND toDate(inserted_at) BETWEEN toDate('{startDate}') AND toDate('{endDate+1}')
GROUP BY saving_time
ORDER BY saving_time ASC
WITH FILL TO toDate('{endDate}') STEP 1
```

---

### 10.7 Impressions Queries

#### `impressionsQueryForUser()`

**Table:** `pinterest_user_insights`

```sql
SELECT
    arrayMap(x -> CAST(x AS Int32), arrayCumSum(impressions_daily)) as impressions_total,
    impressions_daily,
    show_data,
    buckets
FROM (
    SELECT groupArray(impressions) as impressions_daily,
           groupArray(saving_time) as buckets,
           toInt32(sum(impressions)) as show_data
    FROM (
        SELECT toInt32(sum(impression)) as impressions,
               toDate(created_at) as saving_time
        FROM pinterest_user_insights
        WHERE user_id in ({pinterest_id})
          AND toDate(created_at) BETWEEN toDate('{startDate}') AND toDate('{endDate+1}')
        GROUP BY saving_time
        ORDER BY saving_time ASC
        WITH FILL FROM toDate('{account_created_date}') TO toDate('{endDate}') STEP 1
    )
)
```

**Logic:** `arrayCumSum` gives cumulative total. Cast to `Int32` array.

#### `impressionsQueryForBoard()`

**Tables:** `pinterest_pins`, `pinterest_pin_insights`

```sql
WITH pins as (
    SELECT pin_id FROM pinterest_pins WHERE board_id in ({board_id})
)
-- Same structure but queries pinterest_pin_insights WHERE pin_id in pins
```

---

### 10.8 Engagement Queries

#### `engagementQueryForUser()`

**Table:** `pinterest_user_insights`

```sql
SELECT
    saves as saves_daily,
    arrayMap(x -> CAST(x AS Int32), arrayCumSum(saves)) as saves_total,
    outbound_clicks as outbound_clicks_daily,
    arrayMap(x -> CAST(x AS Int32), arrayCumSum(outbound_clicks)) as outbound_clicks_total,
    arrayMap(x -> CAST(x AS Int32), arrayCumSum(pin_clicks)) as pin_clicks_total,
    pin_clicks as pin_clicks_daily,
    engagements as engagement_daily,
    arrayMap(x -> CAST(x AS Int32), arrayCumSum(engagements)) as engagement_total,
    show_data,
    buckets
FROM (
    SELECT
        groupArray(pin_clicks) as pin_clicks,
        groupArray(outbound_clicks) as outbound_clicks,
        groupArray(saves) as saves,
        groupArray(engagement) as engagements,
        toInt32(sum(engagement)) as show_data,
        groupArray(saving_time) as buckets
    FROM (
        SELECT
            toInt32(sum(pin_clicks)) as pin_clicks,
            toInt32(sum(outbound_click)) as outbound_clicks,
            toInt32(sum(saves)) as saves,
            toInt32(sum(engagement)) as engagement,
            toDate(created_at) as saving_time
        FROM pinterest_user_insights
        WHERE user_id in ({pinterest_id})
          AND toDate(created_at) BETWEEN toDate('{startDate}') AND toDate('{endDate+1}')
        GROUP BY saving_time
        ORDER BY saving_time ASC
        WITH FILL FROM toDate('{account_created_date}') TO toDate('{endDate}') STEP 1
    )
)
```

#### `engagementQueryForBoard()`

**Tables:** `pinterest_pins`, `pinterest_pin_insights`

Same structure with CTE `pins` and filtering `pinterest_pin_insights WHERE pin_id in pins`.

---

### 10.9 Pin Posting Per Day

#### `pinPostingPerDayQueryForUser()`

**Table:** `pinterest_pins`

```sql
SELECT groupArray(pin_count) as pins_count,
       groupArray(created_at) as buckets,
       toInt32(sum(pin_count)) as show_data
FROM (
    SELECT toInt32(count(*)) as pin_count,
           toDate(created_at) as created_at
    FROM (
        SELECT pin_id, created_at as created_at
        FROM pinterest_pins
        WHERE user_id in ({pinterest_id})
          AND media_type in ({filter_by})
          AND is_owner = 1
          AND toDate(created_at) BETWEEN toDate('{startDate}') AND toDate('{endDate+1}')
        GROUP BY pin_id, created_at
    )
    GROUP BY created_at
    ORDER BY created_at ASC
    WITH FILL FROM toDate('{account_created_date}') TO toDate('{endDate}') STEP 1
)
```

**Filter:** `is_owner = 1` -- only counts pins the user published, not repins.

---

### 10.10 Pin Posting Rollup (`pinPostingRollupQueryForUser`)

**Tables:** `pinterest_pins`, `pinterest_pin_insights`

```sql
WITH pins as (
    SELECT pin_id
    FROM pinterest_pins
    WHERE user_id in ({pinterest_id})
      AND media_type in ({filter_by})
      AND is_owner = 1
      AND toDate(created_at) BETWEEN toDate('{startDate}') AND toDate('{endDate+1}')
    GROUP BY pin_id
),
insights AS (
    SELECT pin_id,
           SUM(impression) AS impression,
           SUM(pin_clicks) AS pin_clicks,
           SUM(outbound_click) AS outbound_click,
           SUM(saves) AS saves,
           SUM(quartile_95s_percent_view) AS quartile_95s_percent_view,
           SUM(closeup) AS closeup,
           SUM(video_start) AS video_start,
           SUM(video_10s_view) AS video_10s_view,
           SUM(video_avg_watch_time) AS video_avg_watch_time
    FROM pinterest_pin_insights
    WHERE pin_id in (SELECT pin_id FROM pins)
    GROUP BY pin_id
)
SELECT
    count(pin_id) as total_pins,
    if(total_pins > 0, toString(sum(impression)), 'N/A') as impressions,
    if(total_pins > 0, toString(sum(pin_clicks)), 'N/A') as pin_clicks,
    if(total_pins > 0, toString(sum(outbound_click)), 'N/A') as outbound_clicks,
    if(total_pins > 0, toString(sum(saves)), 'N/A') as saves,
    if(total_pins > 0, toString(sum(quartile_95s_percent_view)), 'N/A') as quartile_95s_percent_view,
    if(total_pins > 0, toString(sum(video_start)), 'N/A') as video_views,
    if(total_pins > 0, toString(sum(video_10s_view)), 'N/A') as video_10s_view,
    if(avg(avg_watch_time) > 0 and total_pins > 0, toString(round(avg(avg_watch_time), 2)), 'N/A') as avg_watch_time
FROM (
    SELECT
        p.pin_id as pin_id,
        sum(insights.impression) as impression,
        sum(insights.pin_clicks) as pin_clicks,
        sum(insights.outbound_click) as outbound_click,
        sum(insights.saves) as saves,
        sum(insights.quartile_95s_percent_view) as quartile_95s_percent_view,
        sum(insights.closeup) as closeup,
        sum(insights.video_start) as video_start,
        sum(insights.video_10s_view) as video_10s_view,
        sum(insights.video_avg_watch_time) as avg_watch_time
    FROM pins p
    LEFT JOIN insights ON p.pin_id = insights.pin_id
    GROUP BY p.pin_id
)
```

---

### 10.11 Pins List (`pinsQueryForUser($sort_by)`)

**Tables:** `pinterest_pins`, `pinterest_boards`, `pinterest_pin_insights`

```sql
SELECT
    pins.pin_id as pin_id,
    last_value(boards.name) as board_name,
    format('{}{}', 'https://www.pinterest.com/pin/', pins.pin_id) AS permalink,
    format('{}{}', 'https://assets.pinterest.com/ext/embed.html?id=', pins.pin_id) AS embed_link,
    last_value(pins.title) as title,
    last_value(pins.description) as description,
    last_value(pins.board_owner) as board_owner,
    replaceOne(last_value(pins.media_type), '_', ' ') as media_type,
    last_value(pins.cover_image_url) as cover_image_url,
    last_value(pins.dominant_color) as dominant_color,
    last_value(pins.creative_type) as creative_type,
    last_value(pins.product_tags) as product_tags,
    last_value(pins.height) as height,
    last_value(pins.width) as width,
    last_value(pins.created_at) as created_at,
    toInt32(sum(pinterest_pin_insights.impression)) AS impressions,
    toInt32(sum(pinterest_pin_insights.pin_clicks)) AS pin_clicks,
    toInt32(sum(pinterest_pin_insights.outbound_click)) AS outbound_clicks,
    toInt32(sum(pinterest_pin_insights.saves)) AS saves,
    toInt32(sum(pinterest_pin_insights.engagement)) as total_engagement,
    if(count(*) != 0,
       round(toInt32(pin_clicks + outbound_clicks + saves) / toInt32(count(*)), 2),
       0) as engagement_rate
FROM (
    SELECT pin_id, board_id,
           last_value(title) as title,
           last_value(description) as description,
           last_value(board_owner) as board_owner,
           last_value(media_type) as media_type,
           last_value(cover_image_url) as cover_image_url,
           last_value(dominant_color) as dominant_color,
           last_value(creative_type) as creative_type,
           last_value(product_tags) as product_tags,
           last_value(height) as height,
           last_value(width) as width,
           created_at
    FROM pinterest_pins
    WHERE user_id in {pinterest_id}
      AND toDate(created_at) BETWEEN toDate('{startDate}') AND toDate('{endDate+1}')
      AND is_owner = 1
    GROUP BY pin_id, board_id, created_at
) AS pins
LEFT JOIN (
    SELECT board_id, last_value(name) as name
    FROM pinterest_boards
    GROUP BY board_id
) AS boards ON pins.board_id = boards.board_id
LEFT JOIN (
    SELECT max(impression) as impression,
           max(pin_clicks) as pin_clicks,
           max(outbound_click) as outbound_click,
           max(saves) as saves,
           max(engagement) as engagement,
           pin_id
    FROM pinterest_pin_insights
    WHERE user_id in {pinterest_id}
    GROUP BY pin_id, created_at
    ORDER BY created_at DESC
) AS pinterest_pin_insights USING pin_id
GROUP BY pins.pin_id, pins.board_id
ORDER BY {order_by} {sort_by}
LIMIT {limit}
```

**Special:** `replaceOne(media_type, '_', ' ')` cleans up values. URLs built with `format()`. `engagement_rate = (pin_clicks + outbound_clicks + saves) / count(*)`.

---

### 10.12 Pin Posting Performance (`pinPostingPerformancePerDayQueryForUser`)

**Tables:** `pinterest_pins`, `pinterest_pin_insights`

```sql
WITH pins AS (
    SELECT pin_id, toDate(created_at) as saving_time
    FROM pinterest_pins
    WHERE user_id IN {pinterest_id}
      AND media_type IN ({filter_by})
      AND is_owner = 1
      AND toDate(created_at) BETWEEN toDate('{startDate}') AND toDate('{endDate+1}')
    GROUP BY pin_id, saving_time
    ORDER BY saving_time ASC
),
pin_metrics AS (
    SELECT
        pin_id,
        toInt32(SUM(impression)) AS impressions,
        toInt32(SUM(pin_clicks)) AS pin_clicks,
        toInt32(SUM(outbound_click)) AS outbound_clicks,
        toInt32(SUM(saves)) AS saves,
        toInt32(SUM(engagement)) AS engagement
    FROM pinterest_pin_insights
    WHERE pin_id IN (SELECT pin_id FROM pins)
    GROUP BY pin_id
),
pin_count AS (
    SELECT toDate(created_at) AS saving_time,
           toInt32(COUNT(*)) AS pin_count
    FROM pinterest_pins
    WHERE user_id IN {pinterest_id}
      AND toDate(created_at) BETWEEN toDate('{startDate}') AND toDate('{endDate+1}')
    GROUP BY saving_time
)
SELECT
    groupArray(pin_count) AS pins_count,
    groupArray(pin_clicks) AS pin_clicks,
    groupArray(outbound_clicks) AS outbound_clicks,
    groupArray(saves) AS saves,
    groupArray(engagement) AS engagements,
    groupArray(impressions) AS impressions,
    groupArray(saving_time) AS buckets,
    toInt32(arraySum(pin_clicks)) AS show_data
FROM (
    SELECT
        last_value(pin_id) as pin_id,
        last_value(impressions) as impressions,
        last_value(pin_clicks) as pin_clicks,
        last_value(pin_count) as pin_count,
        last_value(saves) as saves,
        last_value(engagement) as engagement,
        last_value(outbound_clicks) as outbound_clicks,
        saving_time
    FROM (SELECT * FROM pins LEFT JOIN pin_metrics USING pin_id) AS combined_metrics
    LEFT JOIN pin_count USING saving_time
    GROUP BY saving_time
    ORDER BY saving_time ASC
    WITH FILL FROM toDate('{account_created_date}') TO toDate('{endDate}') STEP 1
)
```

### 10.13 Dynamic Aggregation

180-day threshold (same as YouTube/TikTok, different from 60-day for Facebook/Instagram/LinkedIn):
- `dateDiff <= 180` -> daily (`toDate(inserted_at)` + `STEP 1`)
- `dateDiff > 180` -> monthly (`toStartOfMonth(toDate(inserted_at))` + `STEP INTERVAL 1 MONTH`)

---

## 11. Twitter/X Analytics

### 11.1 Constructor

```
timezone = payload['timezone'] ?? 'UTC'
date -> split on ' - ' -> startDate (00:00:00), endDate (23:59:59)
twitter_id -> string or array, merged with optional twitter_accounts
page_id = SQL IN-clause string
post_table = 'twitter_posts'
insights_table = 'twitter_insights'
```

### 11.2 Key Methods

| Method | Description |
|--------|-------------|
| `getPageAndPostsInsights` | Summary with current/previous |
| `getEngagementImpressionData` | 3-level aggregation, NO zero-fill |
| `getFollowerTrend` | Follower count with `runningDifference`, NO zero-fill |
| `getTopTweets` / `getLeastTweets` | Top/bottom tweets by sort column |
| `getCreditsUsedCount` | MongoDB job logs (not ClickHouse) |

### 11.3 Important Differences

- **No zero-fill in any query** (unlike all other platforms)
- Settings methods (add/remove/get accounts, hashtags) use MongoDB, not ClickHouse
- Copy-paste bug: key `tiktok_id` holds the Twitter account ID in some response structures
- `like_count` exists in **both** `twitter_insights` and `twitter_posts` tables

### 11.4 `getPageAndPostsInsights()`

**Tables:** `twitter_insights`, `twitter_posts`

```sql
select *
from
(
    select
        first_value(twitter_id) as twitter_id,
        first_value(name) as name,
        first_value(profile_image_url) as profile_image_url,
        first_value(followers_count) as followers_count,
        first_value(following_count) as following_count,
        first_value(tweet_count) as tweet_count,
        first_value(listed_count) as listed_count,
        first_value(like_count) as like_count,
        first_value(record_date) as record_date
    from
    (
        select
            twitter_id,
            max(name) as name,
            max(profile_image_url) as profile_image_url,
            argMin(followers_count, saving_time) as followers_count,
            argMin(following_count, saving_time) as following_count,
            max(tweet_count) as tweet_count,
            max(listed_count) as listed_count,
            max(like_count) as like_count,
            max(saving_time) as record_date
        from twitter_insights
        where twitter_id in {page_id}
          and toDateTime(saving_time, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
        group by twitter_id, record_id
        order by record_date desc
    )
) as insights_data
left join
(
    select
        twitter_id,
        max(name) as name,
        max(profile_image_url) as profile_image_url,
        sum(impression_count) as impression_count,
        sum(total_engagement) as total_engagement,
        groupUniqArray(tweet_type) as tweet_type,
        sum(reply_count) as reply_count,
        sum(retweet_count) as retweet_count,
        sum(bookmark_count) as bookmark_count,
        sum(like_count) as like_count,
        sum(quote_count) as quote_count,
        count() as tweet_count
    from
    (
        select
            twitter_id, tweet_id,
            max(name) as name,
            max(profile_image_url) as profile_image_url,
            max(impression_count) as impression_count,
            max(total_engagement) as total_engagement,
            max(tweet_type) as tweet_type,
            max(reply_count) as reply_count,
            max(retweet_count) as retweet_count,
            max(bookmark_count) as bookmark_count,
            max(like_count) as like_count,
            max(quote_count) as quote_count,
            max(tweeted_at) as tweeted_at_time
        from twitter_posts
        where twitter_id in {page_id}
          and toDateTime(tweeted_at, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
        group by twitter_id, tweet_id
        order by tweeted_at_time desc
    )
    group by twitter_id
) as posts_data
on insights_data.twitter_id == posts_data.twitter_id
```

**Key:** Insights uses `argMin(followers_count, saving_time)` -- earliest snapshot per record_id. `first_value()` from ordered result picks most recent record_id's data. Posts dedup via `max()` per `(twitter_id, tweet_id)`. `groupUniqArray(tweet_type)` collects unique tweet types.

---

### 11.5 Engagement and Impression Data (`getEngagementImpressionData`)

**Table:** `twitter_posts`

```sql
select
    twitter_id,
    groupArray(tweet_count) as tweet_count,
    max(name) as name,
    max(profile_picture_url) as profile_picture_url,
    groupArray(impression_count) as impression_count,
    groupArray(total_engagement) as total_engagement,
    groupArray(tweeted_at_date) as tweeted_at_date,
    groupArray(retweet_count) as retweet_count,
    groupArray(reply_count) as reply_count,
    groupArray(like_count) as like_count,
    groupArray(bookmark_count) as bookmark_count,
    groupArray(quote_count) as quote_count
from
(
    select
        twitter_id,
        count(tweet_id) as tweet_count,
        max(name) as name,
        max(profile_picture_url) as profile_picture_url,
        sum(impression_count) as impression_count,
        sum(total_engagement) as total_engagement,
        toDate(tweeted_at_time) as tweeted_at_date,
        sum(retweet_count) as retweet_count,
        sum(reply_count) as reply_count,
        sum(like_count) as like_count,
        sum(bookmark_count) as bookmark_count,
        sum(quote_count) as quote_count
    from
    (
        select
            twitter_id, tweet_id,
            max(name) as name,
            max(profile_image_url) as profile_picture_url,
            max(impression_count) as impression_count,
            max(total_engagement) as total_engagement,
            max(tweeted_at) as tweeted_at_time,
            max(retweet_count) as retweet_count,
            max(reply_count) as reply_count,
            max(like_count) as like_count,
            max(bookmark_count) as bookmark_count,
            max(quote_count) as quote_count
        from twitter_posts
        where twitter_id in {page_id}
          and toDateTime(tweeted_at, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
        group by twitter_id, tweet_id
        ORDER BY tweeted_at_time asc
    )
    group by twitter_id, tweeted_at_date
    order by tweeted_at_date asc
)
group by twitter_id
```

**Three-level aggregation:**
1. Innermost: deduplicate per `(twitter_id, tweet_id)` via `max()`
2. Middle: group by `(twitter_id, tweeted_at_date)` -- daily sums + tweet count
3. Outer: group by `twitter_id` -- `groupArray()` to collect daily arrays

**No zero-fill.** Missing days will not appear in the arrays.

---

### 11.6 Follower Trend (`getFollowerTrend`)

**Table:** `twitter_insights`

```sql
select
    platform_id,
    max(name) as name,
    max(username) as username,
    groupArray(follower_count) as follower_count,
    groupArray(follower_count_daily) as follower_count_daily,
    groupArray(following_count) as following_count,
    groupArray(following_count_daily) as following_count_daily,
    groupArray(saving_date) as buckets
from
(
    select
        record_id, platform_id, name, username,
        follower_count,
        if(runningDifference(follower_count) <= 0, 0, runningDifference(follower_count)) as follower_count_daily,
        following_count,
        if(runningDifference(following_count) <= 0, 0, runningDifference(following_count)) as following_count_daily,
        toDate(bucket_time) as saving_date
    FROM
    (
        select
            record_id,
            max(twitter_id) as platform_id,
            max(name) as name,
            max(username) as username,
            last_value(followers_count) as follower_count,
            last_value(following_count) as following_count,
            max(saving_time) as bucket_time
        from twitter_insights
        WHERE twitter_id in {page_id}
          and toDateTime(saving_time, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
        group by record_id
        order by bucket_time asc
    )
)
group by platform_id
```

**Logic:** Groups by `record_id` first (one record per observation period), `last_value()` gets final count. `runningDifference()` computes day-over-day gain. Negative diffs clamped to 0. **No zero-fill** -- gaps in data will cause missing dates.

---

### 11.7 Top/Least Tweets (`getTweetsData($order_by, $limit, $sort)`)

**Table:** `twitter_posts`

```sql
SELECT
    tweet_id as id,
    tweeted_at,
    last_value(tweet_text) as tweet_text,
    last_value(tweet_type) as tweet_type,
    last_value(permalink) as permalink,
    last_value(media_url) as media_url,
    toInt32(last_value(listed_count)) as listed_count,
    toInt32(last_value(retweet_count)) as retweet_count,
    toInt32(last_value(like_count)) as like_count,
    toInt32(last_value(reply_count)) as reply_count,
    toInt32(last_value(quote_count)) as quote_count,
    toInt32(last_value(bookmark_count)) as bookmark_count,
    toInt32(last_value(impression_count)) as impression_count,
    toInt32(last_value(total_engagement)) as total_engagement
FROM twitter_posts
WHERE twitter_id in {page_id}
  and toDateTime(tweeted_at, 0, '{timezone}') BETWEEN toDateTime('{startDate}', 0) AND toDateTime('{endDate}', 0)
GROUP BY id, tweeted_at
ORDER BY {order_by} {sort}
LIMIT {limit}
```

**Deduplication:** Groups by `(tweet_id, tweeted_at)`. Uses `last_value()` -- picks most-recently-inserted snapshot's values. `sort` = `'DESC'` for top, `'ASC'` for least.

---

## 12. Cross-Platform Overview V2

### 12.1 Request Parameters

| Parameter | Rule |
|-----------|------|
| `workspace_id` | required |
| `date` | required |
| `timezone` | required (`'Europe/Kyiv'` remapped to `'Europe/Riga'`) |
| `facebook_accounts` | nullable array |
| `instagram_accounts` | nullable array |
| `linkedin_accounts` | nullable array |
| `pinterest_accounts` | nullable array |
| `youtube_accounts` | nullable array |
| `tiktok_accounts` | nullable array |
| `type` | nullable (metric selection) |
| `limit` | nullable (default 20) |

### 12.2 Date Logic

`currentEndDate = parsed_end + 1 day` (inclusive). Secondary period = equal span before current period.

### 12.3 Summary (`getSummaryQuery`)

The most complex query. Builds per-platform sub-queries for current and secondary periods via UNION ALL, then JOINs and computes percentage changes in pure SQL.

**Per-platform column mappings:**

| Platform | Followers Source | Impressions | Engagement | Reach |
|----------|----------------|-------------|------------|-------|
| Facebook | `page_follows` from insights | `avg(page_impressions)` | `sum(page_post_engagements)` | `avg(page_impressions_unique)` |
| Instagram | `followers_count` from insights | `avg(views)` from posts | `sum(engagement)` from posts | `avg(reach)` from posts |
| LinkedIn | `totalFollowerCount` from insights | `avg(impressions)` from posts | `sum(total_engagement)` from posts | `avg(reach)` from posts |
| TikTok | `total_follower_count` from insights | `sum(view_count)` from posts | `sum(engagement_count)` from posts | = impressions |
| Pinterest | `follower_count` from boards | `sum(impression)` from pin_insights | `sum(engagement)` from pin_insights | = impressions |
| YouTube | `subscriber_count` from channels (with fallback) | `sum(views)` from videos | `sum(likes+comments+shares+dislikes)` | = impressions |

#### Outer Shell

```sql
SELECT
    round(followers, 2) AS followers,
    round(posts, 2) AS posts,
    round(engagement, 2) AS engagement,
    round(impressions, 2) AS impressions,
    round(reach, 2) AS reach,
    round(engagement_rate, 2) AS engagement_rate,
    round(secondary_followers, 2) AS secondary_followers,
    round(secondary_posts, 2) AS secondary_posts,
    round(secondary_engagement, 2) AS secondary_engagement,
    round(secondary_impressions, 2) AS secondary_impressions,
    round(secondary_reach, 2) AS secondary_reach,
    round(secondary_engagement_rate, 2) AS secondary_engagement_rate,
    round(if(secondary_followers > 0, followers - secondary_followers, 0.0), 2) AS diff_followers,
    round(if(secondary_posts > 0, posts - secondary_posts, 0.0), 2) AS diff_posts,
    round(if(secondary_engagement > 0, engagement - secondary_engagement, 0.0), 2) AS diff_engagement,
    round(if(secondary_impressions > 0, impressions - secondary_impressions, 0.0), 2) AS diff_impressions,
    round(if(secondary_reach > 0, reach - secondary_reach, 0.0), 2) AS diff_reach,
    round(if(secondary_engagement_rate > 0, engagement_rate - secondary_engagement_rate, 0.0), 2) AS diff_engagement_rate,
    if(secondary_followers > 0, round((followers - secondary_followers) / secondary_followers * 100, 2), 0.0) AS followers_change_pct,
    if(secondary_posts > 0, round((posts - secondary_posts) / secondary_posts * 100, 2), 0.0) AS posts_change_pct,
    if(secondary_engagement > 0, round((engagement - secondary_engagement) / secondary_engagement * 100, 2), 0.0) AS engagement_change_pct,
    if(secondary_impressions > 0, round((impressions - secondary_impressions) / secondary_impressions * 100, 2), 0.0) AS impressions_change_pct,
    if(secondary_reach > 0, round((reach - secondary_reach) / secondary_reach * 100, 2), 0.0) AS reach_change_pct,
    if(secondary_engagement_rate > 0, round((engagement_rate - secondary_engagement_rate) / secondary_engagement_rate * 100, 2), 0.0) AS engagement_rate_change_pct
FROM (
    SELECT
        followers, posts, engagement, impressions, reach,
        if(impressions > 0, round((engagement / impressions) * 100, 2), 0.0) AS engagement_rate,
        secondary_followers, secondary_posts, secondary_engagement, secondary_impressions, secondary_reach,
        if(secondary_impressions > 0, round((secondary_engagement / secondary_impressions) * 100, 2), 0.0) AS secondary_engagement_rate
    FROM (
        SELECT
            sum(followers) AS followers, sum(total_posts) AS posts,
            sum(engagement) AS engagement, sum(impressions) AS impressions, sum(reach) AS reach,
            sum(secondary_followers) AS secondary_followers, sum(secondary_total_posts) AS secondary_posts,
            sum(secondary_engagement) AS secondary_engagement, sum(secondary_impressions) AS secondary_impressions,
            sum(secondary_reach) AS secondary_reach
        FROM (
            WITH
                current_period AS (
                    -- Facebook subquery
                    SELECT fb_insights.page_id,
                        toFloat64(ifNull(fb_insights.followers_count, 0)) AS followers_count,
                        toFloat64(ifNull(fb_posts.post_count, 0)) AS post_count,
                        toFloat64(ifNull(fb_insights.engagement, 0)) AS engagement,
                        toFloat64(ifNull(fb_insights.impressions, 0)) AS impressions,
                        toFloat64(ifNull(fb_insights.reach, 0)) AS reach
                    FROM (
                        SELECT page_id, first_value(followers_count) AS followers_count,
                            avg(page_impressions) AS impressions, avg(page_impressions_unique) AS reach,
                            sum(page_post_engagements) AS engagement
                        FROM (
                            SELECT page_id, max(page_follows) AS followers_count,
                                max(page_impressions) AS page_impressions,
                                max(page_impressions_unique) AS page_impressions_unique,
                                max(page_post_engagements) AS page_post_engagements
                            FROM facebook_insights
                            WHERE page_id IN {facebook_ids} AND {current_date_filter}
                            GROUP BY page_id, toDate(created_time)
                        ) GROUP BY page_id
                    ) AS fb_insights
                    LEFT JOIN (
                        SELECT page_id, count(post_id) AS post_count
                        FROM (SELECT page_id, post_id FROM facebook_posts
                              WHERE page_id IN {facebook_ids} AND {current_date_filter}
                              GROUP BY page_id, post_id)
                        GROUP BY page_id
                    ) AS fb_posts ON fb_insights.page_id = fb_posts.page_id

                    UNION ALL

                    -- Instagram subquery
                    SELECT ig_insights.page_id,
                        toFloat64(ifNull(ig_insights.followers_count, 0)) AS followers_count,
                        toFloat64(ifNull(ig_posts.post_count, 0)) AS post_count,
                        toFloat64(ifNull(ig_posts.engagement, 0)) AS engagement,
                        toFloat64(ifNull(ig_posts.impressions, 0)) AS impressions,
                        toFloat64(ifNull(ig_posts.reach, 0)) AS reach
                    FROM (
                        SELECT instagram_id AS page_id, first_value(followers_count) AS followers_count
                        FROM (SELECT instagram_id, followers_count FROM instagram_insights
                              WHERE instagram_id IN {instagram_ids} AND {current_date_filter}
                              ORDER BY created_time DESC)
                        GROUP BY instagram_id
                    ) AS ig_insights
                    LEFT JOIN (
                        SELECT page_id, count(post_id) AS post_count,
                            sum(engagement) AS engagement, avg(impressions) AS impressions, avg(reach) AS reach
                        FROM (
                            SELECT instagram_id AS page_id, media_id AS post_id,
                                first_value(engagement) AS engagement, first_value(views) AS impressions,
                                first_value(reach) AS reach
                            FROM (SELECT instagram_id, media_id, engagement, views, reach
                                  FROM instagram_posts WHERE instagram_id IN {instagram_ids}
                                  AND {current_date_filter} ORDER BY post_created_at DESC)
                            GROUP BY page_id, post_id
                        ) GROUP BY page_id
                    ) AS ig_posts ON ig_insights.page_id = ig_posts.page_id

                    UNION ALL

                    -- LinkedIn subquery
                    SELECT li_insights.linkedin_id AS page_id,
                        toFloat64(ifNull(li_insights.followers_count, 0)) AS followers_count,
                        toFloat64(ifNull(li_posts.post_count, 0)) AS post_count,
                        toFloat64(ifNull(li_posts.engagement, 0)) AS engagement,
                        toFloat64(ifNull(li_posts.impressions, 0)) AS impressions,
                        toFloat64(ifNull(li_posts.reach, 0)) AS reach
                    FROM (
                        SELECT linkedin_id, first_value(followers_count) AS followers_count
                        FROM (SELECT linkedin_id, totalFollowerCount AS followers_count
                              FROM linkedin_insights WHERE linkedin_id IN {linkedin_ids}
                              AND {current_date_filter} ORDER BY created_at DESC)
                        GROUP BY linkedin_id
                    ) AS li_insights
                    LEFT JOIN (
                        SELECT linkedin_id, count(post_id) AS post_count,
                            sum(total_engagement) AS engagement, avg(impressions) AS impressions,
                            avg(reach) AS reach
                        FROM (
                            SELECT linkedin_id, post_id, first_value(total_engagement) AS total_engagement,
                                first_value(impressions) AS impressions, first_value(reach) AS reach
                            FROM (SELECT linkedin_id, post_id, total_engagement, impressions, reach
                                  FROM linkedin_posts WHERE linkedin_id IN {linkedin_ids}
                                  AND {current_date_filter} ORDER BY published_at DESC)
                            GROUP BY linkedin_id, post_id
                        ) GROUP BY linkedin_id
                    ) AS li_posts ON li_insights.linkedin_id = li_posts.linkedin_id

                    UNION ALL

                    -- TikTok subquery
                    SELECT tt_insights.page_id,
                        toFloat64(ifNull(tt_insights.followers_count, 0)) AS followers_count,
                        toFloat64(ifNull(tt_posts.post_count, 0)) AS post_count,
                        toFloat64(ifNull(tt_posts.engagement, 0)) AS engagement,
                        toFloat64(ifNull(tt_posts.impressions, 0)) AS impressions,
                        toFloat64(ifNull(tt_posts.impressions, 0)) AS reach
                    FROM (
                        SELECT tiktok_id AS page_id, first_value(followers_count) AS followers_count
                        FROM (SELECT tiktok_id, total_follower_count AS followers_count
                              FROM tiktok_insights WHERE tiktok_id IN {tiktok_ids}
                              AND {current_date_filter} ORDER BY inserted_at DESC)
                        GROUP BY tiktok_id
                    ) AS tt_insights
                    LEFT JOIN (
                        SELECT page_id, count(post_id) AS post_count,
                            sum(engagement) AS engagement, sum(impressions) AS impressions
                        FROM (
                            SELECT tiktok_id AS page_id, post_id,
                                first_value(engagement_count) AS engagement,
                                first_value(view_count) AS impressions
                            FROM (SELECT tiktok_id, post_id, engagement_count, view_count
                                  FROM tiktok_posts WHERE tiktok_id IN {tiktok_ids}
                                  AND {current_date_filter} ORDER BY inserted_at DESC)
                            GROUP BY page_id, post_id
                        ) GROUP BY page_id
                    ) AS tt_posts ON tt_insights.page_id = tt_posts.page_id

                    UNION ALL

                    -- Pinterest subquery
                    SELECT pt_insights.board_id AS page_id,
                        toFloat64(ifNull(pt_insights.followers_count, 0)) AS followers_count,
                        toFloat64(ifNull(pt_posts.post_count, 0)) AS post_count,
                        toFloat64(ifNull(pt_posts.engagement, 0)) AS engagement,
                        toFloat64(ifNull(pt_posts.impressions, 0)) AS impressions,
                        toFloat64(ifNull(pt_posts.impressions, 0)) AS reach
                    FROM (
                        SELECT board_id, first_value(followers_count) AS followers_count
                        FROM (SELECT board_id, follower_count AS followers_count
                              FROM pinterest_boards WHERE board_id IN {pinterest_ids}
                              AND {current_date_filter} ORDER BY inserted_at DESC)
                        GROUP BY board_id
                    ) AS pt_insights
                    LEFT JOIN (
                        SELECT board_id, count(pin_id) AS post_count,
                            sum(engagement) AS engagement, sum(impression) AS impressions
                        FROM (
                            SELECT board_id, pin_id, sum(engagement) AS engagement, sum(impression) AS impression
                            FROM (
                                SELECT board_id, pin_id, record_id,
                                    first_value(engagement) AS engagement, first_value(impression) AS impression
                                FROM (
                                    WITH pins AS (
                                        SELECT pin_id, board_id FROM (
                                            SELECT board_id, pin_id FROM pinterest_pins
                                            WHERE board_id IN {pinterest_ids} AND {current_date_filter}
                                            ORDER BY inserted_at DESC
                                        ) GROUP BY board_id, pin_id
                                    )
                                    SELECT p.*, i.engagement, i.impression, i.record_id, i.inserted_at
                                    FROM pins AS p
                                    LEFT JOIN (
                                        SELECT pin_id, record_id, engagement, impression, inserted_at
                                        FROM pinterest_pin_insights
                                        WHERE pin_id IN (SELECT pin_id FROM pins)
                                        AND {current_date_filter}
                                    ) AS i ON i.pin_id = p.pin_id
                                    ORDER BY inserted_at DESC
                                ) GROUP BY board_id, pin_id, record_id
                            ) GROUP BY board_id, pin_id
                        ) GROUP BY board_id
                    ) AS pt_posts ON pt_insights.board_id = pt_posts.board_id

                    UNION ALL

                    -- YouTube subquery
                    SELECT yt_insights.channel_id AS page_id,
                        toFloat64(ifNull(yt_insights.followers_count, 0)) AS followers_count,
                        toFloat64(ifNull(yt_posts.post_count, 0)) AS post_count,
                        toFloat64(ifNull(yt_posts.engagement, 0)) AS engagement,
                        toFloat64(ifNull(yt_posts.impressions, 0)) AS impressions,
                        toFloat64(ifNull(yt_posts.impressions, 0)) AS reach
                    FROM (
                        SELECT channel_id,
                            if(count_in_range > 0, followers_in_range, followers_latest) AS followers_count
                        FROM (
                            SELECT channel_id,
                                countIf({current_date_filter}) AS count_in_range,
                                argMaxIf(subscriber_count, created_at, {current_date_filter}) AS followers_in_range,
                                argMax(subscriber_count, created_at) AS followers_latest
                            FROM youtube_channels
                            WHERE channel_id IN {youtube_ids}
                            GROUP BY channel_id
                        )
                    ) AS yt_insights
                    LEFT JOIN (
                        SELECT channel_id, count(post_id) AS post_count,
                            sum(likes)+sum(comments)+sum(shares)+sum(dislikes) AS engagement,
                            sum(impressions) AS impressions
                        FROM (
                            SELECT channel_id, video_id AS post_id,
                                first_value(likes) AS likes, first_value(comments) AS comments,
                                first_value(shares) AS shares, first_value(dislikes) AS dislikes,
                                first_value(views) AS impressions
                            FROM (SELECT channel_id, video_id, likes, comments, shares, dislikes, views
                                  FROM youtube_videos WHERE channel_id IN {youtube_ids}
                                  AND {current_date_filter} ORDER BY inserted_at DESC)
                            GROUP BY channel_id, post_id
                        ) GROUP BY channel_id
                    ) AS yt_posts ON yt_insights.channel_id = yt_posts.channel_id
                ),
                secondary_period AS (
                    -- Same 6 platform subqueries with {secondary_date_filter} instead of {current_date_filter}
                    -- Structure is identical, only date range changes
                )
            SELECT
                cp.page_id,
                ifNull(cp.followers_count, 0.0) AS followers,
                ifNull(cp.post_count, 0.0)      AS total_posts,
                ifNull(cp.engagement, 0.0)      AS engagement,
                ifNull(cp.impressions, 0.0)     AS impressions,
                ifNull(cp.reach, 0.0)           AS reach,
                ifNull(sp.followers_count, 0.0) AS secondary_followers,
                ifNull(sp.post_count, 0.0)      AS secondary_total_posts,
                ifNull(sp.engagement, 0.0)      AS secondary_engagement,
                ifNull(sp.impressions, 0.0)     AS secondary_impressions,
                ifNull(sp.reach, 0.0)           AS secondary_reach
            FROM current_period AS cp
            LEFT JOIN secondary_period AS sp ON cp.page_id = sp.page_id
        ) AS per_page
    ) AS totals
) AS aggregated
```

#### Per-Platform Sub-Queries (example: YouTube with follower fallback)

```sql
SELECT
    yt_insights.channel_id AS page_id,
    toFloat64(ifNull(yt_insights.followers_count, 0)) AS followers_count,
    toFloat64(ifNull(yt_posts.post_count, 0)) AS post_count,
    toFloat64(ifNull(yt_posts.engagement, 0)) AS engagement,
    toFloat64(ifNull(yt_posts.impressions, 0)) AS impressions,
    toFloat64(ifNull(yt_posts.impressions, 0)) AS reach,
    'youtube' AS platform
FROM (
    SELECT
        channel_id,
        if(count_in_range > 0, followers_in_range, followers_latest) AS followers_count
    FROM (
        SELECT
            channel_id,
            countIf(toDateTime(created_at) BETWEEN ...) as count_in_range,
            argMaxIf(subscriber_count, created_at, toDateTime(created_at) BETWEEN ...) as followers_in_range,
            argMax(subscriber_count, created_at) as followers_latest
        FROM youtube_channels
        WHERE channel_id IN {$youtube_ids}
        GROUP BY channel_id
    )
) AS yt_insights
LEFT JOIN (
    SELECT channel_id, count(post_id) AS post_count,
        sum(likes)+sum(comments)+sum(shares)+sum(dislikes) AS engagement,
        sum(impressions) AS impressions
    FROM (
        SELECT channel_id, video_id as post_id,
            first_value(likes) as likes, first_value(comments) as comments,
            first_value(shares) as shares, first_value(dislikes) as dislikes,
            first_value(views) AS impressions
        FROM (
            SELECT channel_id, video_id, likes, comments, shares, dislikes, views
            FROM youtube_videos
            WHERE channel_id IN {$youtube_ids}
            AND toDateTime(published_at) BETWEEN ...
            ORDER BY inserted_at DESC
        )
        GROUP BY channel_id, post_id
    )
    GROUP BY channel_id
) AS yt_posts ON yt_insights.channel_id = yt_posts.channel_id
```

**YouTube fallback:** If no data in range (`count_in_range = 0`), uses most recent subscriber count.

---

### 12.4 Top Performing Graph (`getTopPerformingGraphQuery`)

Uses materialized view `mv_social_daily_metrics` with `uniqMerge` and `sumMerge`.

```sql
WITH date_range AS (
    SELECT toDate(addDays(toDate('{currentStart}'), number)) AS date
    FROM numbers(dateDiff('day', toDate('{currentStart}'), toDate('{currentEnd}')) + 1)
),
platforms AS (
    SELECT platform FROM (
        SELECT 'facebook' AS platform UNION ALL SELECT 'instagram' UNION ALL SELECT 'linkedin'
        UNION ALL SELECT 'tiktok' UNION ALL SELECT 'youtube' UNION ALL SELECT 'pinterest'
    )
),
date_platform_combinations AS (
    SELECT dr.date, p.platform
    FROM date_range dr
    CROSS JOIN platforms p
),
actual_data AS (
    SELECT
        date,
        toString(platform) as platform,
        uniqMerge(posts_count) AS post_cnt,
        sumMerge(engagement_sum) AS eng_cnt,
        sumMerge(impressions_sum) AS impr_cnt,
        sumMerge(reach_sum) AS reach_cnt
    FROM mv_social_daily_metrics
    WHERE toDateTime(date) BETWEEN toDateTime('{currentStart}',0) AND toDateTime('{currentEnd}',0)
      AND account_id IN {all_accounts}
    GROUP BY date, platform
),
daily AS (
    SELECT
        dpc.date, dpc.platform,
        coalesce(ad.post_cnt, 0)  AS post_cnt,
        coalesce(ad.eng_cnt, 0)   AS eng_cnt,
        coalesce(ad.impr_cnt, 0)  AS impr_cnt,
        coalesce(ad.reach_cnt, 0) AS reach_cnt
    FROM date_platform_combinations dpc
    LEFT JOIN actual_data ad ON dpc.date = ad.date AND dpc.platform = ad.platform
)
SELECT
    arraySort(arrayDistinct(groupArray(date))) AS buckets,
    groupArrayIf(post_cnt,  platform = 'facebook')  AS facebook_post_count,
    groupArrayIf(post_cnt,  platform = 'instagram') AS instagram_post_count,
    groupArrayIf(post_cnt,  platform = 'linkedin')  AS linkedin_post_count,
    groupArrayIf(post_cnt,  platform = 'tiktok')    AS tiktok_post_count,
    groupArrayIf(post_cnt,  platform = 'youtube')   AS youtube_post_count,
    groupArrayIf(post_cnt,  platform = 'pinterest') AS pinterest_post_count,
    groupArrayIf(eng_cnt,   platform = 'facebook')  AS facebook_engagement_count,
    groupArrayIf(eng_cnt,   platform = 'instagram') AS instagram_engagement_count,
    groupArrayIf(eng_cnt,   platform = 'linkedin')  AS linkedin_engagement_count,
    groupArrayIf(eng_cnt,   platform = 'tiktok')    AS tiktok_engagement_count,
    groupArrayIf(eng_cnt,   platform = 'youtube')   AS youtube_engagement_count,
    groupArrayIf(eng_cnt,   platform = 'pinterest') AS pinterest_engagement_count,
    groupArrayIf(impr_cnt,  platform = 'facebook')  AS facebook_impression_count,
    -- ... same pattern for all platforms and metrics ...
    groupArrayIf(reach_cnt, platform = 'facebook')  AS facebook_reach_count
    -- ... same pattern for all platforms ...
FROM daily
```

**Zero-fill:** `CROSS JOIN` of date_range x platforms ensures every (date, platform) pair exists. `coalesce(..., 0)` fills missing values. `arraySort(arrayDistinct(...))` deduplicates and sorts buckets.

---

### 12.5 Platform Data (`getPlatformDataQuery`) -- Grouped by Platform

```sql
WITH fb_posts as (
    SELECT post_id, max(saving_time)
    FROM facebook_posts WHERE page_id in {$facebook_ids} AND date_filter
    group by post_id
),
followers AS (
    SELECT sum(followers_count) as followers_count, platform_type
    FROM (
        -- Facebook: last_value(page_fans) grouped by page_id
        -- Instagram: last_value(followers_count) from instagram_insights grouped by instagram_id
        -- LinkedIn: last_value(totalFollowerCount) from linkedin_insights grouped by linkedin_id
        -- TikTok: last_value(total_follower_count) from tiktok_insights grouped by tiktok_id
        -- Pinterest: last_value(follower_count) from pinterest_boards grouped by board_id
        -- YouTube: if(count_in_range>0, followers_in_range, followers_latest) from youtube_channels
    )
    GROUP BY platform_type
)
SELECT
    toInt32(sum(followers_count)) as followers,
    toInt32(sum(total_posts)) as total_posts,
    toInt32(sum(total_engagements)) as engagement,
    toInt32(sum(total_impression)) as impressions,
    toInt32(sum(total_reach)) as reach,
    toInt32(sum(total_reactions)) as reactions,
    toInt32(sum(total_comments)) as comments,
    toInt32(sum(total_shares)) as shares,
    platform_type
FROM followers
LEFT JOIN (
    -- Facebook: (reactions+comments+shares), post_impressions, post_impressions_unique
    -- Instagram: engagement, views(=impressions), reach, like_count(=reactions), comments_count, saved(=shares)
    -- LinkedIn: total_engagement, impressions, reach, favorites(=reactions), comments, repost(=shares)
    -- TikTok: engagement_count, view_count(=impressions and reach), like_count, comments_count, share_count
    -- YouTube: likes+comments+shares+dislikes(=engagement), views, likes(=reactions), comments, shares
    -- Pinterest: engagement, impression, pin_clicks(=reactions), outbound_click(=comments), saves(=shares)
) AS posts_data ON posts_data.platform_type = followers.platform_type
GROUP BY platform_type
ORDER BY platform_type DESC
```

**Column mapping quirks:**
- Instagram: `saved` maps to `shares` slot
- Pinterest: `pin_clicks` = reactions, `outbound_click` = comments, `saves` = shares
- TikTok: `reach` = 0 hardcoded in graph queries

---

### 12.6 Account Data Detailed (`getAccountDataDetailedQuery`)

FULL JOIN of current and old (secondary) period data with percentage changes:

```sql
WITH
    current_data AS ( [per-account metrics with current DateFilter] ),
    old_data AS ( [per-account metrics with SecondaryDateFilter] )
SELECT
    COALESCE(cd.platform_type, od.platform_type) AS platform_type,
    COALESCE(cd.account_id, od.account_id) AS account_id,
    cd.account_name AS account_name,
    toInt32(COALESCE(cd.followers, 0))   AS current_followers,
    toInt32(COALESCE(od.followers, 0))   AS old_followers,
    toInt32(COALESCE(cd.total_posts, 0)) AS current_posts,
    toInt32(COALESCE(od.total_posts, 0)) AS old_posts,
    toInt32(COALESCE(cd.engagement, 0))  AS current_engagement,
    toInt32(COALESCE(od.engagement, 0))  AS old_engagement,
    toInt32(COALESCE(cd.impressions, 0)) AS current_impressions,
    toInt32(COALESCE(od.impressions, 0)) AS old_impressions,
    toInt32(COALESCE(cd.reach, 0))       AS current_reach,
    toInt32(COALESCE(od.reach, 0))       AS old_reach,
    IF(od.followers=0, 0, round(((cd.followers-od.followers)*100.0)/od.followers,2)) AS followers_change_pct,
    IF(od.total_posts=0, 0, round(((cd.total_posts-od.total_posts)*100.0)/od.total_posts,2)) AS posts_change_pct,
    IF(od.engagement=0, 0, round(((cd.engagement-od.engagement)*100.0)/od.engagement,2)) AS engagement_change_pct,
    IF(od.impressions=0, 0, round(((cd.impressions-od.impressions)*100.0)/od.impressions,2)) AS impressions_change_pct,
    IF(od.reach=0, 0, round(((cd.reach-od.reach)*100.0)/od.reach,2)) AS reach_change_pct
FROM current_data AS cd
FULL JOIN old_data AS od ON cd.account_id = od.account_id AND cd.platform_type = od.platform_type
ORDER BY platform_type DESC
```

---

### 12.7 Account Data Graphs (`getAccountDataGraphsQuery`)

Per-account time-series arrays:

```sql
WITH aggregated_data AS (
    SELECT date, account_id,
           SUM(engagement) AS engagement, SUM(reach) AS reach,
           SUM(impressions) AS impressions, SUM(total_posts) AS total_posts
    FROM (
        -- Facebook: toDate(created_time), page_id as account_id, last_value(total+comments+shares)
        -- Instagram: toDate(post_created_at), instagram_id, last_value(engagement/reach/views)
        -- LinkedIn: toDate(created_at), linkedin_id, first_value(total_engagement/reach/impressions)
        -- YouTube: toDate(published_at), channel_id, last_value(likes+comments+shares+dislikes)
        -- TikTok: toDate(created_at), tiktok_id, last_value(engagement_count/view_count), reach=0
        -- Pinterest: pins CTE + LEFT JOIN pin_insights, reach=0, impressions=0
    )
    GROUP BY date, account_id
    ORDER BY date ASC
),
final_data AS (
    SELECT account_id, date,
           max(engagement) AS engagement, max(reach) AS reach,
           max(impressions) AS impressions, max(total_posts) AS total_posts
    FROM aggregated_data GROUP BY account_id, date
    ORDER BY account_id ASC, date ASC
)
SELECT
    account_id,
    groupArray(engagement)  AS engagement,
    groupArray(reach)       AS reach,
    groupArray(impressions) AS impressions,
    groupArray(total_posts) AS posts,
    groupArray(date)        AS buckets
FROM final_data
GROUP BY account_id
ORDER BY engagement DESC
```

**Note:** TikTok `reach` = 0 hardcoded, Pinterest `reach` and `impressions` = 0 hardcoded.

---

### 12.8 Top Posts (`getTopPostsQuery`)

Unified schema across all 6 platforms via UNION ALL:

```sql
SELECT
    platform_type, account_id, post_id,
    toInt32(reactions) AS likes,
    toInt32(comment_count) AS comments,
    toInt32(shares_count) AS shares,
    toInt32(saves) AS saves,
    toInt32(pin_clicks) AS pin_clicks,
    toInt32(outbound_clicks) AS outbound_clicks,
    toInt32(dislikes_count) AS dislikes_count,
    permalink, media_type, thumbnail, category,
    created_time,
    toInt32(total_engagement) AS total_engagement,
    toInt32(views) AS views,
    toInt32(total_impressions) AS reach
FROM all_posts
ORDER BY {$this->type} DESC
LIMIT {$this->limit}
```

**Per-platform column mapping:**

| Column | Facebook | Instagram | LinkedIn | TikTok | YouTube | Pinterest |
|---|---|---|---|---|---|---|
| reactions/likes | `total` | `like_count` | `favorites` | `like_count` | `likes` | 0 |
| comment_count | `comments` | `comments_count` | `comments` | `comments_count` | `comments` | 0 |
| shares_count | `shares` | 0 | `repost` | `share_count` | `shares` | 0 |
| saves | 0 | `saved` | 0 | 0 | 0 | `saves` |
| pin_clicks | 0 | 0 | 0 | 0 | 0 | `pin_clicks` |
| outbound_clicks | 0 | 0 | 0 | 0 | 0 | `outbound_click` |
| dislikes_count | 0 | 0 | 0 | 0 | `dislikes` | 0 |
| total_engagement | reactions+comments+shares | likes+comments+saved | favorites+comments+repost | likes+comments+shares | likes+comments+shares+dislikes | saves+pin_clicks+outbound_click |
| views | 0 | `views` | 0 | `view_count` | `views` | 0 |
| total_impressions/reach | `post_impressions` | `reach` | `reach` | `view_count` | `views` | `impression` |
| permalink | `permalink` | `permalink` | `article_url` | `share_url` | extracted from `iframe_embed_html` | `https://www.pinterest.com/pin/{pin_id}` |
| thumbnail | `full_picture` | `media_url[1]` | `image` | `embed_link` | `thumbnail_url` | `cover_image_url` |
| category | `caption` | `caption` | `title` | `post_description` | `description` | `description` |

**Instagram:** `media_type != 'STORY'` -- Stories excluded from top posts.

**YouTube permalink:**
```sql
REPLACE(
    concat('https://', substring(iframe_embed_html, position('//' IN iframe_embed_html)+2, ...)),
    'embed/', 'watch?v='
)
```

**Facebook Reels:** Separate CTE pre-selects `(post_id, max(saving_time))` to ensure latest snapshot.

---

## 13. Campaign and Label Analytics

### 13.1 Two-Phase Lookup Pattern

1. Resolve post IDs from MongoDB (`campaign_analytics`/`label_analytics` collections)
2. Query ClickHouse with those IDs

**MongoDB resolution flow:**
- `getCampaignPostIdsForSummary`: Queries `campaign_analytics` collection by campaign IDs
- `getLabelPostIdsForSummary`: Same pattern with `label_analytics`
- If campaign/label not found in MongoDB: fetches plan IDs from `PlansRepository`, resolves social post IDs, creates MongoDB records

### 13.2 Constructor and Date Handling

```php
$this->allPostedIds = "['" . implode("','", $payload['all_post_ids']) . "']";  // ClickHouse array literal
$this->currentEndDate = Carbon::parse($dateArray[1])->addDay();  // +1 day inclusive
// Previous period computed by setPreviousDate():
$previousDate = date_sub(date_create($date[0]), $dateDifference)->format('Y-m-d')
    . ' - ' . date_create($date[0])->format('Y-m-d');
```

### 13.3 Fully Qualified Table Names

This builder uses `contentstudiobackend.` database prefix for all tables, unlike other builders: `contentstudiobackend.facebook_posts`, `contentstudiobackend.instagram_posts`, etc.

### 13.4 Post ID CTE Methods

#### `inMemoryTableForOptimization()` -- Virtual in-memory lookup
```sql
SELECT 'campaign_id_1' AS id, 'post_id_a' AS post_id
UNION ALL SELECT 'campaign_id_1' AS id, 'post_id_b' AS post_id
...
```

#### `selectPostIdsForOptimization()` -- Array-based CTE (more efficient)
```sql
WITH pairs AS (
    select [
        ('campaign_id_1','post_id_a'),
        ('campaign_id_1','post_id_b'),
        ('campaign_id_2','post_id_c'),
        ...
    ] as pairs_array
),
postIds AS (
    SELECT pair.1 AS id, pair.2 AS post_id
    FROM pairs ARRAY JOIN pairs_array as pair
)
```

### 13.5 Summary (`getSummaryQuery`)

```sql
WITH postIds AS (
    SELECT arrayJoin(['post_id_1','post_id_2',...]) AS post_id
),
facebook_reels AS (
    SELECT last_value(post_id) as post_id
    FROM contentstudiobackend.facebook_video_insights
    WHERE video_id IN (SELECT post_id FROM postIds)
    GROUP BY video_id
)

SELECT
    toInt32(sum(total_posts)) as total_posts,
    toInt32(sum(total_engagements)) as total_engagement,
    toInt32(sum(total_impression)) as total_impressions,
    if(total_impressions!=0, round(total_engagement/total_impressions, 2), 0) as total_engagement_rate_per_impression
FROM (
    -- Facebook (if flagSetup['facebook'])
    SELECT
        toInt32(count()) as total_posts,
        toInt32(sum(total_engagement)) as total_engagements,
        toInt32(sum(total_impressions)) as total_impression
    FROM (
        SELECT last_value(total_engagement) as total_engagement,
               last_value(post_impressions) as total_impressions
        FROM contentstudiobackend.facebook_posts
        WHERE (post_id IN (SELECT post_id FROM postIds) OR post_id IN facebook_reels)
          AND toDateTime(created_time) BETWEEN ...
        GROUP BY post_id
    )

    UNION ALL

    -- Instagram (if flagSetup['instagram'])
    SELECT toInt32(count()), toInt32(sum(engagement)), toInt32(sum(views))
    FROM (
        SELECT last_value(engagement), last_value(views)
        FROM contentstudiobackend.instagram_posts
        WHERE media_id IN (SELECT post_id FROM postIds) AND date_filter
        GROUP BY media_id
    )

    UNION ALL

    -- LinkedIn (if flagSetup['linkedin'])
    SELECT toInt32(count()), toInt32(sum(total_engagement)), toInt32(sum(impressions))
    FROM (
        SELECT last_value(total_engagement), last_value(impressions)
        FROM contentstudiobackend.linkedin_posts
        WHERE activity IN (SELECT post_id FROM postIds) AND date_filter
        GROUP BY activity
    )

    UNION ALL

    -- TikTok (if flagSetup['tiktok'])
    SELECT toInt32(count()), toInt32(sum(total_engagement)), toInt32(sum(view_count))
    FROM (
        SELECT last_value(engagement_count) as total_engagement, last_value(view_count)
        FROM contentstudiobackend.tiktok_posts
        WHERE post_id IN (SELECT post_id FROM postIds) AND date_filter
        GROUP BY post_id
    )

    UNION ALL

    -- YouTube (if flagSetup['youtube'] -- hardcoded to true)
    SELECT toInt32(count()), toInt32(sum(total_engagement)), toInt32(sum(views))
    FROM (
        SELECT argMax(likes,inserted_at)+argMax(comments,inserted_at)
              +argMax(shares,inserted_at)+argMax(dislikes,inserted_at) as total_engagement,
               argMax(views,inserted_at) as views
        FROM contentstudiobackend.youtube_videos
        WHERE video_id IN (SELECT post_id FROM postIds) AND date_filter
        GROUP BY video_id
    )

    UNION ALL

    -- Pinterest (if flagSetup['pinterest'])
    WITH pins AS (
        SELECT pin_id FROM contentstudiobackend.pinterest_pins
        WHERE pin_id IN (SELECT post_id FROM postIds) AND date_filter
        GROUP BY pin_id
    ),
    pinterest_insights AS (
        SELECT pin_id, toInt32(SUM(engagement)) AS engagement, toInt32(SUM(impression)) AS impression
        FROM (
            SELECT pin_id, last_value(engagement), last_value(impression)
            FROM contentstudiobackend.pinterest_pin_insights
            WHERE pin_id in (SELECT post_id FROM postIds)
            GROUP BY record_id, pin_id
        )
        GROUP BY pin_id
    )
    SELECT
        toInt32(count()) as total_posts,
        toInt32(sum(total_engagement)) as total_engagements,
        toInt32(sum(total_impressions)) as total_impression
    FROM (
        SELECT toInt32(SUM(pinterest_insights.engagement)) AS total_engagement,
               toInt32(SUM(pinterest_insights.impression)) AS total_impressions
        FROM pins LEFT JOIN pinterest_insights ON pins.pin_id = pinterest_insights.pin_id
        GROUP BY pins.pin_id
    )
)
```

**Key:** Platform inclusion is conditional via `$flagSetup`. Trailing `UNION ALL` removed via `substr($query, 0, -10)`. Uses `arrayJoin()` on array literal for post IDs (different from breakdown which uses `ARRAY JOIN`).

---

### 13.6 Breakdown Data (`getBreakdownData($period)`)

Per-campaign/label totals for a given period:

```sql
WITH pairs AS (...), postIds AS (...),
facebook_reels AS (
    SELECT video_id, post_id
    FROM contentstudiobackend.facebook_video_insights
    WHERE video_id IN (SELECT post_id FROM postIds)
    GROUP BY video_id, post_id
),
pinterest_insights AS (
    SELECT pin_id,
        toInt32(SUM(engagement)) AS engagement,
        toInt32(SUM(impression)) AS impression
    FROM (
        SELECT pin_id,
               argMax(engagement, inserted_at) as engagement,
               argMax(impression, inserted_at) as impression
        FROM contentstudiobackend.pinterest_pin_insights
        WHERE pin_id in (SELECT post_id FROM postIds)
        GROUP BY record_id, pin_id
    )
    GROUP BY pin_id
)

SELECT
    id,
    COALESCE('{$period}', '{$period}') as era,
    COALESCE(toInt32(count()), 0) as total_posts,
    COALESCE(toInt32(sum(total_engagement)), 0) as total_engagement,
    COALESCE(toInt32(sum(total_impressions)), 0) as total_impressions
FROM (
    -- Facebook Reels (video posts via video_insights join)
    SELECT
        facebook_reels.video_id as post_id,
        sum(total_engagement) as total_engagement,
        sum(total_impressions) as total_impressions
    FROM (
        SELECT post_id,
            toFloat64(argMax(total_engagement, saving_time)) as total_engagement,
            toFloat64(argMax(post_impressions, saving_time)) as total_impressions
        FROM contentstudiobackend.facebook_posts
        WHERE post_id in (SELECT post_id FROM facebook_reels)
          AND toDateTime(created_time) BETWEEN toDateTime('{start}', 0) AND toDateTime('{end}', 0)
        GROUP BY post_id
    ) AS reels
    LEFT JOIN facebook_reels ON reels.post_id = facebook_reels.post_id
    GROUP BY post_id

    UNION ALL

    -- Facebook regular posts
    SELECT post_id,
        toFloat64(argMax(total_engagement, saving_time)) AS total_engagement,
        toFloat64(argMax(post_impressions, saving_time)) AS total_impressions
    FROM contentstudiobackend.facebook_posts
    WHERE post_id IN (SELECT post_id FROM postIds) AND date_filter
    GROUP BY post_id

    UNION ALL

    -- Instagram
    SELECT media_id AS post_id,
        toFloat64(argMax(engagement, stored_event_at)) AS total_engagement,
        toFloat64(argMax(views, stored_event_at)) AS total_impressions
    FROM contentstudiobackend.instagram_posts
    WHERE media_id IN (SELECT post_id FROM postIds) AND date_filter
    GROUP BY media_id

    UNION ALL

    -- LinkedIn
    SELECT activity AS post_id,
        toFloat64(argMax(total_engagement, saving_time)) AS total_engagement,
        toFloat64(argMax(impressions, saving_time)) AS total_impressions
    FROM contentstudiobackend.linkedin_posts
    WHERE activity IN (SELECT post_id FROM postIds) AND date_filter
    GROUP BY activity

    UNION ALL

    -- TikTok
    SELECT post_id,
        toFloat64(argMax(engagement_count, inserted_at)) AS total_engagement,
        toFloat64(argMax(view_count, inserted_at)) AS total_impressions
    FROM contentstudiobackend.tiktok_posts
    WHERE post_id IN (SELECT post_id FROM postIds) AND date_filter
    GROUP BY post_id

    UNION ALL

    -- YouTube
    SELECT video_id AS post_id,
        toFloat64(argMax(likes,inserted_at) + argMax(comments,inserted_at)
                + argMax(shares,inserted_at) + argMax(dislikes,inserted_at)) AS total_engagement,
        toFloat64(argMax(views, inserted_at)) AS total_impressions
    FROM contentstudiobackend.youtube_videos
    WHERE video_id IN (SELECT post_id FROM postIds) AND date_filter
    GROUP BY video_id

    UNION ALL

    -- Pinterest
    SELECT pins.pin_id as post_id,
        toFloat64(SUM(pinterest_insights.engagement)) AS total_engagement,
        toFloat64(SUM(pinterest_insights.impression)) AS total_impressions
    FROM (
        SELECT pin_id FROM contentstudiobackend.pinterest_pins
        WHERE pin_id IN (SELECT post_id FROM postIds) AND date_filter
        GROUP BY pin_id
    ) AS pins
    LEFT JOIN pinterest_insights ON pins.pin_id = pinterest_insights.pin_id
    GROUP BY pins.pin_id

) AS all_posts
LEFT JOIN postIds ON postIds.post_id = all_posts.post_id
GROUP BY postIds.id
```

**Key patterns:**
- Uses `argMax(value, timestamp)` for most metrics -- takes value at maximum timestamp
- `era` column is always `$period` ('current' or 'previous') -- allows merging in PHP
- Facebook Reels: resolved via `facebook_video_insights.video_id` -> `facebook_posts.post_id`
- Pinterest: pre-built CTE with `argMax(engagement/impression, inserted_at) GROUP BY record_id, pin_id`

---

### 13.7 Insights Time-Series (`getInsightsData`)

Per-campaign/label time-series:

```sql
WITH pairs AS (...), postIds AS (...),
facebook_reels AS (...),
pinterest_insights AS (...)

SELECT
    id,
    groupArray(total_engagement) as total_engagement,
    groupArray(total_impressions) as total_impressions,
    groupArray(total_posts) as total_posts,
    groupArray(created_at) as created_at
FROM (
    SELECT
        postIds.id as id,
        toInt32(sum(total_engagement)) as total_engagement,
        toInt32(sum(total_impressions)) as total_impressions,
        toInt32(count()) as total_posts,
        toDate(created_at) as created_at
    FROM (
        -- Facebook Reels
        SELECT facebook_reels.video_id as post_id,
            sum(total_engagement) as total_engagement,
            sum(total_impressions) as total_impressions,
            min(created_time) as created_at
        FROM (
            SELECT post_id,
                toFloat64(argMax(total_engagement, saving_time)) as total_engagement,
                toFloat64(argMax(post_impressions, saving_time)) as total_impressions,
                min(created_time) as created_time
            FROM contentstudiobackend.facebook_posts
            WHERE post_id in (SELECT post_id FROM facebook_reels)
              AND toDateTime(created_time) BETWEEN toDateTime('{start}', 0) AND toDateTime('{end}', 0)
            GROUP BY post_id
        ) AS reels
        LEFT JOIN facebook_reels ON reels.post_id = facebook_reels.post_id
        GROUP BY post_id

        UNION ALL

        -- Facebook regular posts
        SELECT post_id,
            toFloat64(argMax(total_engagement, saving_time)) AS total_engagement,
            toFloat64(argMax(post_impressions, saving_time)) AS total_impressions,
            min(created_time) as created_at
        FROM contentstudiobackend.facebook_posts
        WHERE post_id IN (SELECT post_id FROM postIds) AND date_filter
        GROUP BY post_id

        UNION ALL

        -- Instagram
        SELECT media_id AS post_id,
            toFloat64(argMax(engagement, stored_event_at)) AS total_engagement,
            toFloat64(argMax(views, stored_event_at)) AS total_impressions,
            min(post_created_at) as created_at
        FROM contentstudiobackend.instagram_posts
        WHERE media_id IN (SELECT post_id FROM postIds) AND date_filter
        GROUP BY media_id

        UNION ALL

        -- LinkedIn
        SELECT activity AS post_id,
            toFloat64(argMax(total_engagement, saving_time)) AS total_engagement,
            toFloat64(argMax(impressions, saving_time)) AS total_impressions,
            min(created_at) as created_at
        FROM contentstudiobackend.linkedin_posts
        WHERE activity IN (SELECT post_id FROM postIds) AND date_filter
        GROUP BY activity

        UNION ALL

        -- TikTok
        SELECT post_id,
            toFloat64(argMax(engagement_count, inserted_at)) AS total_engagement,
            toFloat64(argMax(view_count, inserted_at)) AS total_impressions,
            min(created_at) as created_at
        FROM contentstudiobackend.tiktok_posts
        WHERE post_id IN (SELECT post_id FROM postIds) AND date_filter
        GROUP BY post_id

        UNION ALL

        -- YouTube
        SELECT video_id AS post_id,
            toFloat64(argMax(likes, inserted_at) + argMax(comments, inserted_at)
                     + argMax(shares, inserted_at) + argMax(dislikes, inserted_at)) AS total_engagement,
            toFloat64(argMax(views, inserted_at)) AS total_impressions,
            min(published_at) as created_at
        FROM contentstudiobackend.youtube_videos
        WHERE video_id IN (SELECT post_id FROM postIds) AND date_filter
        GROUP BY video_id

        UNION ALL

        -- Pinterest
        SELECT pins.pin_id as post_id,
            toFloat64(SUM(pinterest_insights.engagement)) AS total_engagement,
            toFloat64(SUM(pinterest_insights.impression)) AS total_impressions,
            min(pins.created_at) as created_at
        FROM (
            SELECT pin_id, min(created_at) as created_at
            FROM contentstudiobackend.pinterest_pins
            WHERE pin_id IN (SELECT post_id FROM postIds) AND date_filter
            GROUP BY pin_id
        ) AS pins
        LEFT JOIN pinterest_insights ON pins.pin_id = pinterest_insights.pin_id
        GROUP BY pins.pin_id
    ) AS all_posts
    LEFT JOIN postIds ON all_posts.post_id = postIds.post_id
    GROUP BY created_at, postIds.id
    ORDER BY created_at DESC
)
GROUP BY id
```

Returns one row per campaign/label ID with arrays of `total_engagement[]`, `total_impressions[]`, `total_posts[]`, and `created_at[]`.

---

### 13.8 Planner Analytics (`getAnalyticsForPlannerQuery`)

Returns detailed per-post metrics for a single platform. Each platform returns metric-specific columns with tooltip strings:

**Facebook:** `engagement`, `impressions`, `reach`, `comments`, `repost` (shares), `post_clicks`, `reactions` (total), plus individual reaction types (`likes`, `love`, `wow`, `haha`, `sad`, `anger`), `media_type`

**Instagram:** `engagement`, `impressions`, `reach`, `likes` (like_count), `comments` (comments_count), `saves` (saved), `media_type` (CAROUSEL_ALBUM -> 'Carousel')

**LinkedIn:** `engagement`, `impressions`, `reach`, `comments`, `reactions` (favorites), `reposts` (repost), `post_clicks`, `media_type`

**TikTok:** `engagement` (engagement_count), `views` (view_count), `likes`, `comments`, `shares`, `engagement_rate`

**YouTube:** `engagement` (likes+comments+shares+dislikes), `views`, `likes`, `dislikes`, `comments`, `shares`, `red_views`, `subscribers_gained`, `minutes_watched`, `red_minutes_watched`, `average_view_duration`, `media_type`

**Pinterest:** `engagement`, `impressions`, `pin_clicks`, `outbound_clicks`, `saves`, `engagement_rate`

### 13.9 Summary Diff Calculation

```php
foreach ($current as $key => $value) {
    $prevValue = $previous[$key] ?? 'N/A';
    $res['difference'][$key] = ($value !== 'N/A' && $prevValue !== 'N/A')
        ? round(intval($value) - intval($prevValue), 2) : 'N/A';
    $res['percentage'][$key] = ($value !== 'N/A' && $prevValue !== 'N/A')
        ? round((intval($value) - intval($prevValue)) * 100 / max($prevValue, 1), 2) : 'N/A';
}
```

`max($prevValue, 1)` prevents divide-by-zero when previous = 0.

---

## 14. Date Handling Patterns

### 14.1 Platform-Specific Date Filters

| Platform | Posts Filter | Insights Filter |
|----------|-------------|----------------|
| **Facebook** | `toDateTime(created_time, 0, '{tz}') BETWEEN ...` | `toDate(created_time) BETWEEN ...` |
| **Instagram** | `toDateTime(post_created_at, 0, '{tz}') BETWEEN ...` | `toDateTime(created_time, 0, '{tz}') BETWEEN ...` |
| **LinkedIn** | `toDateTime(published_at, 0, '{tz}') BETWEEN ...` | `toDateTime(created_at, 0, '{tz}') BETWEEN ...` |
| **YouTube** | `toDateTime(published_at) >= ... AND < ...+1day` | `toDateTime(created_at) >= ... AND < ...+1day` |
| **TikTok** | `toDateTime(created_at, 0, '{tz}') BETWEEN ...` | `toDateTime(inserted_at, 0, '{tz}') BETWEEN ...` |
| **Pinterest** | `toDate(created_at) BETWEEN ...` (tz-naive) | `toDate(created_at) BETWEEN ...` (tz-naive) |
| **Twitter** | `toDateTime(created_at, 0, '{tz}') BETWEEN ...` | `toDateTime(inserted_at, 0, '{tz}') BETWEEN ...` |
| **Competitor** | Dates converted to UTC in PHP before SQL | Same (UTC-converted) |
| **Overview V2** | `toDateTime(field) BETWEEN ...` (no tz) | Same |
| **Campaign/Label** | `toDateTime(field) BETWEEN ...` (no tz) | Same |

### 14.2 DateFilter Side-Effect

Facebook, YouTube, Pinterest, and LinkedIn builders mutate `currentEndDate` within `DateFilter()`. Call only once per query. YouTube adds a day then subtracts back. LinkedIn's `engagementsQuery` calls `subDay()` before building.

---

## 15. Deduplication Strategies

### 15.1 CTE Deduplication Pattern (Facebook, Instagram, LinkedIn)

```sql
WITH posts AS (
    SELECT post_id, max(saving_time)
    FROM {platform}_posts
    WHERE {filters}
    GROUP BY post_id
)
SELECT ...
FROM {platform}_posts
WHERE (post_id, saving_time) IN (posts)
```

### 15.2 `last_value` Deduplication (YouTube, TikTok, Twitter)

```sql
SELECT
    last_value(field) as field
FROM {table}
GROUP BY record_id
```

### 15.3 `argMax`/`argMin` Deduplication

- `argMin(value, saving_time)` -- earliest record (Facebook insights fan count)
- `argMax(value, inserted_at)` -- latest record (YouTube videos, Campaign/Label)

---

## 16. Zero-Fill Patterns

### 16.1 `WITH FILL` (ClickHouse native)

```sql
ORDER BY date ASC
WITH FILL FROM toDate('{start}') TO toDate('{end}') + 1 STEP 1
```

### 16.2 `arrayFill` (Forward-fill)

```sql
arrayFill(x -> not x == 0, array)
```

Carries last non-zero value forward through zeros.

### 16.3 `arrayReverseFill` (Backward-fill)

```sql
arrayReverseFill(x -> not x == 0, array)
```

Used by YouTube subscriber trend in combination with `arrayFill`.

### 16.4 `INTERPOLATE AS -1` Sentinel (TikTok)

```sql
INTERPOLATE (
    follower_count as -1,
    views_per_day as -1
)
```

Marks gap days with -1. Controller trims leading -1 values, then forward-fills remaining gaps.

### 16.5 Controller-Level Leading-Zero Backfill

When `array[0] == 0`, extends date range 2 years back, fetches last known non-zero value, replaces all leading zeros. Used by: Facebook audience growth, Instagram audience growth, LinkedIn audience growth, YouTube subscriber trend.

### 16.6 No Zero-Fill (Twitter)

Twitter queries have no `WITH FILL` and no gap-filling logic.

---

## 17. Dynamic Aggregation Patterns

### 17.1 Thresholds

| Platform | Threshold | Daily | Monthly |
|----------|-----------|-------|---------|
| Facebook | 60 days | `toDate(field)`, STEP 1 | `toStartOfMonth(field)`, STEP INTERVAL 1 MONTH |
| Instagram | 60 days | Same | Same |
| LinkedIn | 60 days | Same | Same |
| YouTube | 180 days | Same | Same |
| TikTok | 180 days | Same | Same |
| Pinterest | 180 days | Same | Same |

### 17.2 Response Addition

Dynamic queries add `aggregation_level` field ('daily' or 'monthly') to the response.

### 17.3 AI Summary Variants

Facebook has additional AI summary methods that combine all metrics into a single dynamic query with `media_type_data` map array breakdown.

---

## 18. Known Quirks and Bugs

### 18.1 Facebook

- `getDateFilters` has side-effect mutating `currentEndDate` via `subDay()`
- Gender demographics: M/U/F labels mapped to wrong aliases in CASE expressions
- Age demographics: Date boundary clamped to 2024-03-14
- Active users timezone offset: `round(offsetHours) + 8` (UTC-8 storage assumption)

### 18.2 Instagram

- `'18-24'` age bucket aliased as `'18-34'` (copy-paste error)
- `overviewEngagement` reads `previous_date` from request instead of computing it
- `impressionsQuery` uses `views` column as impressions in Overview V2

### 18.3 LinkedIn

- `engagementsQuery` mutates `currentEndDate` via `subDay()` before building
- Demographics use `contentstudiobackend.linkedin_insights` (fully qualified) while all other queries use unqualified table names
- `total_engagement` excludes clicks (unlike the engagement_rate formula that works on impressions)
- Multi-hashtag filter appears buggy (comma-joined into single `has()` call)

### 18.4 YouTube

- Column typo: `non_subsciber_watch_time` (missing 's' in subscriber)
- `findVideoQuery` uses 13 traffic sources (excludes shorts_views) while `videoViewsTrendQuery` uses all 14
- `getEngagementRollup` has `DateFilter` called twice (duplicate filter bug)
- `avg_view_duration` computed differently in `summaryQuery` vs `sumSummaryQuery`
- `engagement_rate` in top/least posts = `engagement / count(*)` (records count, not views)
- Timezone parameter stored but never applied in any SQL query

### 18.5 TikTok

- Uses `INTERPOLATE AS -1` sentinel pattern (unique to TikTok)
- `runningDifference` can produce negative values; clamped to 0

### 18.6 Pinterest

- `'Europe/Kyiv'` remapped to `'Europe/Riga'` as ClickHouse workaround
- Date filtering is timezone-naive (`toDate()` not `toDateTime()`)
- Board mode uses `CROSS JOIN` (valid only because both sides produce one row)
- `engagement_rate` = `(pin_clicks + outbound_clicks + saves) / count(*)` (not impressions-based)

### 18.7 Twitter/X

- Copy-paste bug: response key `tiktok_id` holds Twitter account ID
- No zero-fill in any query
- Settings methods use MongoDB, not ClickHouse

### 18.8 Overview V2

- `currentEndDate = parsed_end + 1 day`
- TikTok/Pinterest/YouTube: `reach = impressions` (same field)
- Instagram uses `views` as impressions
- TikTok `reach` = 0 hardcoded in `getAccountDataGraphsQuery`
- Pinterest `reach` = 0 and `impressions` = 0 hardcoded in `getAccountDataGraphsQuery`

### 18.9 Campaign/Label

- Uses fully qualified `contentstudiobackend.` table names
- LinkedIn uses `activity` column as post_id (not `post_id`)
- `setPreviousDate` uses PHP `date_diff` which returns `d` component only (not total days) -- may be incorrect for multi-month ranges
- `getSummaryQuery` uses `arrayJoin` while `getBreakdownData` uses `ARRAY JOIN` with pairs -- different optimization approaches

---

## 19. Cross-Platform Overview V1

**Controller:** `OverviewController`
**File:** `app/Http/Controllers/Analytics/Analyze/OverviewController.php`

The V1 overview aggregates data across all platforms using per-platform builders and returns combined metrics.

### 19.1 Route Map

| Route | Method |
|-------|--------|
| `POST /analytics/overview/summary` | `overview` |
| `POST /analytics/overview/topPosts` | `topPosts` |
| `POST /analytics/overview/postsEngagement` | `postsEngagement` |
| `POST /analytics/overview/engagementRollup` | `engagementRollup` |
| `POST /analytics/overview/accountPerformance` | `accountPerformance` |
| `POST /analytics/overview/timeRecommendation` | `timeRecommendation` |

### 19.2 Request Parameters (Universal)

| Parameter | Type | Description |
|-----------|------|-------------|
| `workspace_id` | string | Required workspace ID |
| `date` | string | `"YYYY-MM-DDTHH:mm:ss - YYYY-MM-DDTHH:mm:ss"` (ISO 8601) |
| `timezone` | string | IANA timezone |
| `facebook_accounts` | array | Facebook page IDs |
| `instagram_accounts` | array | Instagram account IDs |
| `linkedin_accounts` | array | LinkedIn page IDs |
| `youtube_accounts` | array | YouTube channel IDs |
| `tiktok_accounts` | array | TikTok account IDs |
| `pinterest_accounts` | array | Pinterest board IDs |
| `previous_date` | string | Optional previous period for comparison |

### 19.3 Overview Summary (`overview`)

Aggregates summary metrics from all enabled platforms for current and optional previous period.

**Per-Platform Data Sources:**
- **Facebook**: `FacebookController` (Elasticsearch indices)
- **Instagram**: `InstagramBuilder.publishRollupQuery()` + `summaryQuery()` (ClickHouse)
- **LinkedIn**: `LinkedinBuilder.summaryQuery()` (ClickHouse)
- **YouTube**: `YoutubeBuilder.sumSummaryQuery()` + `summaryQuery()` (ClickHouse)
- **TikTok**: `TiktokBuilder.getPageAndPostsInsights()` (ClickHouse)

**Response:**
```json
{
  "status": true,
  "ids": {
    "facebook": [], "instagram": [], "linkedin": [],
    "youtube": [], "tiktok": [], "pinterest": [], "twitter": []
  },
  "overview": {
    "current": {
      "total_posts": 0, "reposts": 0, "comments": 0, "reactions": 0
    },
    "previous": {
      "total_posts": 0, "reposts": 0, "comments": 0, "reactions": 0
    }
  }
}
```

### 19.4 Engagement Rollup (`engagementRollup`)

Calculates per-day engagement and posting rates across all platforms.

**Calculation:**
```
total_days_current = (end_date - start_date).days + 1
total_days_previous = (end_date - start_date).days     // no +1 for previous
engagement_per_day = total_engagement / total_days
posts_per_day = total_posts / total_days
```

**Response:**
```json
{
  "engagement_rollup": {
    "current": {
      "total_engagement": 0, "engagement_per_day": 0.0,
      "total_posts": 0, "posts_per_day": 0.0
    },
    "previous": { "..." }
  }
}
```

### 19.5 Top Posts (`topPosts`)

Fetches top 15 posts per platform sorted by total engagement, then merges and sorts overall.

**Per-Platform Limits:** 15 posts each
**Overall Sort:** `array_multisort()` by engagement descending

**Response:**
```json
{
  "top_posts": {
    "overall": [{"post_id": "", "total_engagement": 0, "network": "facebook", "...": ""}],
    "facebook": [], "instagram": [], "linkedin": [],
    "youtube": [], "tiktok": [], "pinterest": [], "twitter": []
  }
}
```

### 19.6 Posts Engagement Time Series (`postsEngagement`)

Aggregates daily engagement across all platforms into time-series buckets.

**Response:**
```json
{
  "posts_engagements": {
    "buckets": ["YYYY-MM-DD"],
    "data": {
      "total_engagement": [], "comments": [],
      "reactions": [], "reposts": [], "post_count": []
    },
    "show_data": 0
  }
}
```

`show_data` = sum of all engagement + post counts. Used by frontend to decide whether to render the chart.

### 19.7 Account Performance (`accountPerformance`)

Returns per-platform and per-account engagement breakdown.

**Metrics by Platform:**

| Platform | Specific Metrics |
|----------|-----------------|
| Facebook | comments, reactions, reposts, post_clicks, total_engagement, total_posts |
| Instagram | comments, reactions, reposts, saved, total_engagement, total_posts |
| LinkedIn | comments, reactions, reposts, total_engagement, total_posts |
| YouTube | comments, reactions, reposts, total_engagement, total_posts |
| TikTok | comments, reactions, reposts, total_engagement, total_posts |
| Pinterest | pin_clicks, outbound_clicks, saves, total_engagement, followers |

### 19.8 Time Recommendation (`timeRecommendation`)

Best time to post heatmap using last 3 months of data.

**Request:**
```json
{
  "state": "merged",
  "accounts": [{"facebook_id": "", "instagram_id": "", "linkedin_id": ""}]
}
```

**OverviewBuilder Query Pattern:**
- Subtracts 3 months from start date for historical window
- Groups by `day_of_week` (0-6) and `hour_of_day` (0-23)
- Calculates engagement score: `(impressions + engagement) / post_count`
- Applies timezone offset to DateHistogram

**Per-Platform Fields:**

| Platform | ID Field | Date Field | Engagement Field | Impression Field |
|----------|----------|------------|-----------------|-----------------|
| Facebook | `page_id` | `created_time` | `total_engagement` | `post_impressions` |
| Instagram | `instagram_id` | `post_created_at` | `engagement` | `views` |
| LinkedIn | `linkedin_id` | `published_at` | `total_engagement` | (none, hardcoded 0) |

**Response:**
```json
{
  "data": [{
    "facebook": {"page_id": {"day_of_week": {"hour_of_day": 0.0}}},
    "instagram": {"...": "..."},
    "linkedin": {"...": "..."},
    "merged": {"day_of_week": {"hour_of_day": 0.0}}
  }]
}
```

When `state = "merged"`, scores are averaged across all platforms into a single heatmap.

---

## 20. Dashboard Analytics

**Controller:** `DashboardAnalytics`
**File:** `app/Http/Controllers/Analytics/Analytics/DashboardAnalytics.php`

Dashboard statistics for content publishing, approval workflows, and inbox management.

### 20.1 Content Publishing Stats

**Route:** `POST /getContentPublishingStats`

**Request:**
```json
{
  "workspace_id": "string (required)",
  "date_range": "YYYY-MM-DD - YYYY-MM-DD (required)"
}
```

**Business Logic:**
- Applies permission filtering based on user role
  - Collaborators: filtered to authorized accounts with `dashboard=true`
  - Admins: `dashboard_admin=true` for full access
- Calls `PlansRepository::fetchPlansCounts()` with workspace/date/permission filters

**Response:**
```json
{
  "status": true,
  "stats": {
    "scheduled": 0,
    "published": 0,
    "partial": 0,
    "failed": 0
  }
}
```

### 20.2 Approval Publishing Stats

**Route:** `POST /getContentApprovalStats`

**Request:** Same as content publishing (`workspace_id`, `date_range`)

**Response:**
```json
{
  "status": true,
  "stats": {
    "review": 0,
    "rejected": 0,
    "missed": 0
  }
}
```

### 20.3 Inbox Stats

**Route:** `POST /getInboxStats`

**Request:**
```json
{
  "workspace_id": "string (required)"
}
```

**Business Logic:**
- Retrieves conversation counts across 4 states
- Supports Facebook, Instagram, LinkedIn, GMB platforms
- Uses `InboxFiltersBuilder` with `ConversationFlag` filter per state

**Response:**
```json
{
  "status": true,
  "stats": {
    "UNASSIGNED": 0,
    "ASSIGNED": 0,
    "MARKED_AS_DONE": 0,
    "MINE": 0
  }
}
```

---

## 21. Analytics Share Link Management

**Controller:** `AnalyticsShareLinkController`
**File:** `app/Http/Controllers/Analytics/AnalyticsShareLinkController.php`
**MongoDB Collection:** `analytics_share_links`

Enables creating, managing, and sharing analytics dashboards via generated links with controlled access.

### 21.1 Data Model

```javascript
{
  "_id": ObjectId,
  "link_id": "unique_generated_uid",
  "title": "string (max 255)",
  "workspace_id": "string",
  "user_id": "string",
  "platform": "overview|facebook|instagram|linkedin|tiktok|youtube|pinterest|twitter",
  "is_date_range_fixed": boolean,
  "date_range": { "from": "YYYY-MM-DD", "to": "YYYY-MM-DD" } || null,
  "is_account_switching_enabled": boolean,
  "is_password_protected": boolean,
  "password": "string" || null,
  "account_id": "string" || null,
  "overview_accounts": array || null,
  "is_disabled": boolean,
  "created_by": "string",
  "updated_by": "string",
  "created_at": ISODate,
  "updated_at": ISODate
}
```

**Relationships:** `belongsTo User` via `created_by` field

### 21.2 Create Share Link

**Route:** `POST /analytics/share-link/create`

**Request:**
```json
{
  "title": "string (required, max 255)",
  "workspace_id": "string (required)",
  "platform": "string (required, one of: overview|facebook|instagram|linkedin|tiktok|youtube|pinterest|twitter)",
  "is_date_range_fixed": "boolean (optional, default: false)",
  "date_range": "string (required if is_date_range_fixed is true)",
  "is_account_switching_enabled": "boolean (optional, default: true)",
  "is_password_protected": "boolean (optional, default: false)",
  "password": "string (required if is_password_protected, min: 4 chars)",
  "account_id": "string (optional)",
  "overview_accounts": "array (optional)"
}
```

**Business Logic:**
1. Generates unique `link_id` via `Helper::generateUid()`
2. Stores `user_id` from authenticated user
3. Sets `is_disabled = false`
4. Constructs share URL: `{app_url}share/analytics/{link_id}` (WhiteLabelHelper)

**Response:**
```json
{
  "status": "success",
  "data": { "share_url": "https://app.example.com/share/analytics/{link_id}" },
  "message": "Share link created successfully"
}
```

### 21.3 Update Share Link

**Route:** `POST /analytics/share-link/update/{id}`

Updates title, date range settings, account switching, password settings. Sets `updated_by` to current user.

### 21.4 Toggle Share Link State

**Route:** `PUT /analytics/share-link/update/toggle-state/{id}`

**Request:** `{ "is_disabled": boolean }`

Enables/disables a share link without deleting it.

### 21.5 Fetch Share Links

**Route:** `GET /analytics/share-link/list/{workspace_id}`

Returns all share links for workspace, sorted by `created_at DESC`, with eager-loaded user relationship (`firstname`, `lastname`).

### 21.6 Delete Share Link

**Route:** `DELETE /analytics/share-link/delete/{id}`

Hard deletes the share link document from MongoDB.

### 21.7 Get Share Link Details (Public)

**Route:** `GET /analytics/shared/{link_id}`

Retrieves share link by `link_id` (not `_id`). Returns full document. Used by public share page.

### 21.8 Verify Share Link Password

**Route:** `POST /analytics/shared/{link_id}/verify-password`

**Request:** `{ "password": "string (required)" }`

**Business Logic:**
1. If share link is NOT password protected: returns success without verification
2. If password protected: direct string comparison (`$shareLink->password === $request->input('password')`)

**Note:** Passwords are stored in plain text (not hashed).

---

## 22. Reports and Scheduled Reports

### 22.1 Analytics Reports

**Controller:** `AnalyticsReports`
**File:** `app/Http/Controllers/Analytics/Analytics/AnalyticsReports.php`
**MongoDB Collection:** `reports`

#### Data Model

```javascript
{
  "_id": ObjectId,
  "workspace_id": "string",
  "user_id": "string",
  "name": "string",
  "type": "single-pdf-detailed|single-pdf-overview|multiple-pdf-overview|multiple-pdf-detailed|group|competitor",
  "platform_type": "facebook|instagram|linkedin|...",
  "accounts": [],
  "date": "YYYY-MM-DD - YYYY-MM-DD",
  "source": "scheduled|manual|export" || null,
  "status": "pending|processing|completed|failed",
  "progress": 0-100,
  "export_url": "https://s3-bucket.../report.pdf" || null,
  "expire_time": datetime || null,
  "execution_id": "string" || null,
  "schedule_id": ObjectId || null,
  "language": "en|es|fr|...",
  "labels": [],
  "campaigns": [],
  "topPosts": integer,
  "error": boolean || null,
  "error_message": "string" || null,
  "created_at": datetime,
  "updated_at": datetime
}
```

#### Store Report

**Route:** `POST /analytics/reports/save`

**Request:**
```json
{
  "type": "string (required: single-pdf-detailed|single-pdf-overview|multiple-pdf-overview|multiple-pdf-detailed|group|competitor)",
  "action": "save|render|email (optional, default: save)",
  "workspace_id": "string (required)",
  "name": "string (optional)",
  "platform_type": "string (optional)",
  "accounts": "array (optional)",
  "date": "string (optional)",
  "email_list": "array (optional)",
  "language": "string (optional, default: en)",
  "labels": "array (optional)",
  "campaigns": "array (optional)",
  "topPosts": "integer (optional)"
}
```

**Actions:**
- `save`: Persists report config to MongoDB, returns report object
- `render`: Creates `ExportReportJob` and dispatches to queue
- `email`: Same as render but with `send_email: true` and `email_list`

**ExportReportJob Categories:**
- `single`: Combined report for selected accounts
- `multiple`: Individual account PDFs batched together
- `grouped`: Grouped report across accounts
- `competitor`: Competitor analysis report

#### Show Report

**Route:** `POST /analytics/reports/show`

**Request:** `{ "_id": "ObjectId (required)" }`

Returns single report by ID. Used to check generation progress and get export URL.

#### List Reports

**Route:** `POST /analytics/reports/list`

**Request:** `{ "workspace_id": "string (required)" }`

**Query Filters:**
- Excludes scheduled reports (`source` must be null)
- Excludes `single-pdf-detailed` type
- Limited to last 30 reports, sorted `created_at DESC`
- Competitor reports enriched via MongoDB aggregation pipeline joining `competitors_reports`

#### Remove Report

**Route:** `POST /analytics/reports/remove`

**Request:** `{ "report_id": "string (required)" }`

Hard deletes from MongoDB. No cascading operations.

### 22.2 Scheduled Reports

**Controller:** `ScheduleReports`
**File:** `app/Http/Controllers/Analytics/Analytics/ScheduleReports.php`
**MongoDB Collection:** `schedule_reports`

#### Data Model

```javascript
{
  "_id": ObjectId,
  "workspace_id": "string",
  "user_id": "string",
  "name": "string",
  "type": "group|individual|...",
  "platform_type": "string",
  "accounts": [],
  "email_list": [],
  "report_type": "string",
  "frequency": "daily|weekly|monthly|custom",
  "cron": "string" || null,
  "day_of_week": 0-6 || null,
  "day_of_month": 1-31 || null,
  "time": "HH:MM" || null,
  "timezone": "IANA timezone",
  "next_run_at": datetime,
  "last_run_at": datetime || null,
  "last_execution_id": "string" || null,
  "consecutive_failures": 0,
  "max_attempts": 3,
  "active": true,
  "paused_at": datetime || null,
  "paused_reason": "string" || null,
  "created_at": datetime,
  "updated_at": datetime
}
```

#### Show Scheduled Reports

**Route:** `POST /analytics/reports/schedule/show`

**Request:** `{ "workspace_id": "string (required)" }`

Returns all scheduled reports for workspace.

#### Create/Update Schedule

**Route:** `POST /analytics/reports/schedule/save`

**Request:**
```json
{
  "_id": "string (optional, for updates)",
  "workspace_id": "string (required)",
  "name": "string (required)",
  "type": "string (required)",
  "platform_type": "string (required)",
  "accounts": "array (required)",
  "email_list": "array (required)",
  "frequency": "daily|weekly|monthly|custom (required)",
  "day_of_week": "integer (optional, 0-6)",
  "day_of_month": "integer (optional, 1-31)",
  "time": "HH:MM (optional)",
  "timezone": "string (optional)",
  "cron": "string (optional, overrides other scheduling fields)"
}
```

If `_id` provided, updates existing; otherwise creates new.

**Auto-Pause:** After `max_attempts` consecutive failures, sets `active=false`, records `paused_at` and `paused_reason`.

#### Remove Schedule

**Route:** `POST /analytics/reports/schedule/remove`

**Request:** `{ "report_id": "string (required)" }`

Hard deletes schedule. Execution history is NOT deleted.

#### Send Scheduled Reports (Cron)

**Route:** `POST /analytics/reports/schedule/send`

**Request:** `{ "interval": "daily|weekly|monthly" }`

Called by Laravel scheduler. For each due schedule: creates report record, creates execution history, dispatches `ExportReportJob`, updates `next_run_at` and `last_run_at`.

---

## 23. Analytics Job Triggers

**Controller:** `AnalyticsJobController`
**File:** `app/Http/Controllers/Analytics/Analyze/AnalyticsJobController.php`

### 23.1 Trigger Analytics Job

**Route:** `POST /analytics/triggerJob`

**Request:**
```json
{
  "workspace_id": "string (required)",
  "account_id": "string (required)",
  "platform": "string (required)"
}
```

**Business Logic:**
1. Validates account exists in `SocialIntegrations` by `_id`, `workspace_id`, and `platform_type`
2. Checks cache key `analytics:immediate-job-{platform_identifier}-{platform_type}` to prevent duplicate triggers
3. Calls `AnalyticsHelper::triggerAnalyticsPipeline($platform_type, $account)`
4. Caches trigger for 1 hour (60-minute TTL)

**Response:**
```json
{ "status": true, "message": "Localized message" }
```

### 23.2 Trigger Competitor Job

**Route:** `POST /analytics/triggerCompetitorJob`

**Request:**
```json
{
  "workspace_id": "string (required)",
  "report_id": "string (required)",
  "platform": "string (required)",
  "competitor_ids": "array (required)"
}
```

**Business Logic:**
1. Checks cache key `analytics:competitor-report-job-{report_id}`
2. Fetches competitors from `CompetitorsModel` matching IDs and platform
3. Iterates and calls `CompetitorsRepo::triggerIgCompetitorJob()` for each
4. Caches for 1 hour if all successful

---

## 24. Twitter/X Settings Management

**Controller:** `TwitterController` (settings methods)
**File:** `app/Http/Controllers/Analytics/Analyze/TwitterController.php`
**MongoDB Collection:** `twitter_job_settings`

### 24.1 Create Setting

**Route:** `POST /analytics/settings/twitter/createTwitterAnalyticsSetting`

**Request:**
```json
{
  "platform_id": "string (required, must exist in TwitterAccounts)",
  "workspace_id": "string (required, must exist in Workspace)",
  "updated_by": "string (required, must exist in User)",
  "created_by": "string (required, must exist in User)",
  "job_type": "string (required)",
  "trigger_day": "string (required)",
  "platform_name": "string (required)",
  "post_count": "integer (required)"
}
```

Creates job configuration for automatic Twitter analytics fetching.

### 24.2 Update Setting

**Route:** `POST /analytics/settings/twitter/updateTwitterAnalyticsSetting`

Same parameters as create. Finds by composite key (`platform_id`, `workspace_id`) and updates.

### 24.3 Get Settings

**Route:** `POST /analytics/settings/twitter/getTwitterAnalyticsSetting`

**Request:** `{ "workspace_id": "string (required)" }`

Returns all Twitter job settings for workspace.

### 24.4 Fetch Single Setting

**Route:** `POST /analytics/settings/twitter/fetchTwitterAnalyticsSetting`

Returns a single setting record.

### 24.5 Fetch All Settings

**Route:** `POST /analytics/settings/twitter/fetchTwitterAnalyticsSettings`

Returns all settings across workspaces (admin endpoint).

### 24.6 Delete Setting

**Route:** `POST /analytics/settings/twitter/deleteTwitterAnalyticsSetting`

Removes setting by ID.

### 24.7 Trigger Twitter Analytics Job

**Route:** `POST /analytics/settings/twitter/triggerTwitterAnalyticsJob`

**Request:**
```json
{
  "platform_id": "string (required, must exist in TwitterAccounts)",
  "workspace_id": "string (required)"
}
```

**Business Logic:**
1. Retrieves Twitter account details
2. Constructs Argo workflow API URL: `{ARGO_BASE_URL}/social-account-added-{ARGO_ENV}`
3. Sends payload: `{"channel": "twitter", "account-id": "{account_id}"}`
4. Triggers asynchronous data fetching via Argo Workflows

### 24.8 Create Job Logs

**Route:** `POST /analytics/settings/twitter/createTwitterJobLogs`

**Request:**
```json
{
  "platform_id": "string (required)",
  "workspace_id": "string (required)",
  "platform_type": "string (required)",
  "job_type": "string (required)",
  "credits_used": "integer (required)",
  "executed_by": "string (required)",
  "app_id": "string (required)",
  "app_name": "string (required)",
  "error": "string (required)"
}
```

**MongoDB Collection:** `twitter_jobs_metadata`

Automatically calculates: `job_executed_at` (UTC ISO), `day_of_week` (0-6), `hour_of_day` (0-23). Used for API quota tracking and usage analytics.

### 24.9 Get Credits Used Count

**Route:** `POST /analytics/overview/twitter/getCreditsUsedCount`

Returns total API credits used by the workspace's Twitter analytics jobs.

---

## 25. Competitor Management CRUD

### 25.1 Facebook Competitor Search

**Controller:** `FacebookCompetitorController`
**Route:** `POST /analytics/overview/facebook/competitor/search`

**Request:** `{ "search": "string (required, page name)" }`

**Business Logic:**
1. Retrieves valid Facebook token from Redis: `redis-data-science:facebook_valid_token_set`
2. Calls Facebook Graph API `/pages/search` with fields: `id, name, link, verification_status, location`
3. Implements 3-attempt retry logic
4. Transforms response: renames `id` to `competitor_id`, generates profile picture URL

**Response:**
```json
{
  "data": [{
    "competitor_id": "string",
    "name": "string",
    "link": "string",
    "verification_status": "string",
    "location": {},
    "image": "https://graph.facebook.com/{id}/picture?type=large"
  }]
}
```

### 25.2 Instagram Competitor Search

**Controller:** `InstagramCompetitorController`
**Route:** `POST /analytics/overview/instagram/competitor/search`

**Request:** `{ "search": "string (required, username or URL)" }`

**Business Logic:**
1. Extracts slug from URL if necessary
2. Retrieves valid Instagram token from Redis: `redis-data-science:instagram_valid_token_set`
3. Calls Facebook Graph API `business_discovery` endpoint
4. Decrypts token via `SocialHelper::decryptToken()`
5. Generates `appsecret_proof` for API security

**Response:**
```json
{
  "data": [{
    "biography": "string",
    "competitor_id": "string",
    "id": "string",
    "name": "string",
    "image": "string (profile picture URL)",
    "slug": "string (username)"
  }]
}
```

### 25.3 Add/Update Competitor Report

**Route:** `POST /analytics/overview/instagram/competitor/addUpdateCompetitorReport`

**Request:**
```json
{
  "_id": "ObjectId (optional, for updates)",
  "platform_type": "facebook|instagram",
  "workspace_id": "string",
  "name": "string",
  "created_by_user_id": "string",
  "updated_by_user_id": "string",
  "competitors": [{"competitor_id": "", "name": "", "image": "", "slug": ""}]
}
```

**Business Logic:**
1. Creates/updates report in `competitors_reports` collection
2. Creates/updates individual competitor records in `competitors` collection
3. Logs operation via `LogsBuilder`
4. Triggers background Argo job to fetch competitor analytics data

### 25.4 Get Competitor Report

**Route:** `POST /analytics/overview/instagram/competitor/getCompetitorReport`

**Request:** `{ "_id": "ObjectId (required)" }`

Returns single report with all competitor documents populated.

### 25.5 List Competitor Reports

**Route:** `POST /analytics/overview/instagram/competitor/getCompetitorReportsByWorkspace`

**Request:** `{ "workspace_id": "string", "platform_type": "string" }`

Returns all competitor reports for workspace and platform, each with populated competitor details.

### 25.6 Delete Competitor Report

**Route:** `POST /analytics/overview/instagram/competitor/deleteCompetitorReport`

**Request:** `{ "_id": "ObjectId (required)" }`

Hard deletes report document. Individual competitor records remain (reusable across reports).

---

## 26. AI Insights

**Controllers:** Platform-specific in `app/Http/Controllers/Analytics/AI/`
**AI Service:** `AiAgentService` at `app/Services/AI/AiAgentService.php`

### 26.1 Architecture

Each platform has a dedicated AI Insights controller that:
1. Receives a `type` parameter identifying which insight to generate
2. Checks cache (24-hour TTL) by key: `{platform}_AI:{method}:{account_id}:{date}:{locale}`
3. Calls the corresponding platform controller/builder to gather data
4. Validates data availability (checks array sums for zero values)
5. Sends dataset to external AI Agent service
6. Caches and returns response

**AI Agent Service:**
- Base URL from config `ai_agents.base_url`
- Bearer token auth from config `ai_agents.api_key`
- 120-second timeout
- Injects current locale as `language` parameter
- Endpoint pattern: `{platform}/{insight-type}`

### 26.2 Facebook AI Insights

**Route:** `GET /analytics/overview/facebook/ai_insights`

**Request:**
```json
{
  "workspace_id": "string (required)",
  "date": "YYYY-MM-DD - YYYY-MM-DD (required)",
  "facebook_id": "string (required)",
  "type": "string (required)",
  "timezone": "string (required)",
  "limit": "integer (required)"
}
```

**Insight Types:**

| Type | Builder Method | AI Endpoint | Data Validation |
|------|---------------|-------------|-----------------|
| `page_impressions` | `getImpressionsAIInsights()` | `facebook/page-impressions` | `array_sum(page_impressions) === 0` |
| `page_engagement` | `getEngagementAIInsights()` | `facebook/page-engagement` | `array_sum(page_engagements) === 0` |
| `publishing_behaviour_impressions` | `getPublishingBehaviourAIInsights()` | `facebook/publishing-impressions` | `array_sum(post_count) === 0` |
| `publishing_behaviour_engagements` | `getPublishingBehaviourAIInsights(type)` | `facebook/publishing-engagement` | post count check |
| `publishing_behaviour_reach` | `getPublishingBehaviourAIInsights(type)` | `facebook/publishing-reach` | post count check |
| `audience_growth` | `getAudienceGrowthAIInsights()` | `facebook/audience-growth` | `array_sum(fan_count) === 0` |
| `video_views` | `getVideoAIInsights()` | `facebook/video-views` | `array_sum(total_posts) === 0` |
| `video_watch_time` | `getVideoAIInsights(type)` | `facebook/video-watch-time` | total_posts check |
| `video_engagements` | `getVideoAIInsights(type)` | `facebook/video-engagement` | total_posts check |
| `reels_initial_plays` | `getReelsAIInsights(type)` | `facebook/reels-plays` | `array_sum(total_posts) === 0` |
| `reels_watch_time` | `getReelsAIInsights(type)` | `facebook/reels-watch-time` | total_posts check |
| `reels_engagement` | `getReelsAIInsights(type)` | `facebook/reels-engagement` | total_posts check |
| `top_posts` | `getTopPosts('current')` | `facebook/top-posts` | `count(top_posts) === 0` |
| `insights_summary` | `getCombinedAISummaryData()` + `getTopPosts()` | `facebook/insights-summary` | `array_sum(post_count) === 0` |

**Summary Payload (insights_summary):**
```json
{
  "facebook_page_data": {},
  "reels_data": {},
  "publishing_behaviour": {},
  "facebook_video_data": {},
  "top_posts": [],
  "language": "en"
}
```

### 26.3 Instagram AI Insights

**Route:** `POST /analytics/overview/instagram/ai_insights`

**Insight Types:**

| Type | Builder Method | AI Endpoint |
|------|---------------|-------------|
| `impressions` | `getImpressions()` | `instagram/impressions` |
| `engagement` | `getEngagement()` | `instagram/engagement` |
| `publishing_behaviour_impressions` | `getDynamicPublish()` | `instagram/publishing-impressions` |
| `publishing_behaviour_engagements` | `getDynamicPublish('engagement')` | `instagram/publishing-engagement` |
| `publishing_behaviour_reach` | `getDynamicPublish('reach')` | `instagram/publishing-reach` |
| `audience_growth` | `getAudienceGrowth()` | `instagram/audience-growth` |
| `reels_engagement` | `getReelsDynamic('engagement')` | `instagram/reels-engagement` |
| `reels_watch_time` | `getReelsDynamic('watch_time')` | `instagram/reels-watch-time` |
| `reels_shares` | `getReelsDynamic('shares')` | `instagram/reels-shares` |
| `stories_interactions` | `getStoriesDynamic()` | `instagram/stories-interactions` |
| `stories_impressions` | `getStoriesDynamic()` | `instagram/stories-impressions` |
| `stories_reach` | `getStoriesDynamic()` | `instagram/stories-reach` |
| `top_posts` | `getTopPosts($request)` | `instagram/top-posts` |
| `top_hashtags` | `getHashtags()` | `instagram/hashtags` |
| `insights_summary` | `getSummary()` + `getReels()` + `getStories()` + `getPublish()` + `getTopPosts()` + `getHashtags()` | `instagram/insights-summary` |

### 26.4 YouTube AI Insights

**Route:** `POST /analytics/overview/youtube/ai_insights`

**Additional Request Parameter:** `trend_type` (nullable string, for subscriber and engagement trends)

**Insight Types:**

| Type | Builder Method | AI Endpoint |
|------|---------------|-------------|
| `subscribers_trend` | `overviewDynamicSubscriberTrend()` | `youtube/cumulative-subscribers-trend` |
| `daily_views` | `overviewDynamicViewsTrend()` | `youtube/daily-views` |
| `daily_engagement` | `overviewDynamicEngagementTrend()` | `youtube/daily-engagement` |
| `daily_watch_time` | `overviewDynamicWatchTimeTrend()` | `youtube/daily-watch-time` |
| `viewers_find_videos` | `overviewFindVideo()` | `youtube/traffic-sources` |
| `engagement_vs_posting_pattern` | `overviewPerformanceAndVideoPostingSchedule()` | `youtube/posting-patterns` |
| `sharing_services` | `overviewVideoSharing()` | `youtube/sharing-trends` |
| `top_and_least_posts` | `overviewLeastPosts()` + `overviewTopPosts()` | `youtube/top-least-performing-posts` |
| `insights_summary` | All trend methods + `overviewFindVideo()` + `overviewVideoSharing()` + `overviewLeastPosts()` | `youtube/overview-summary` |

### 26.5 LinkedIn AI Insights

**Route:** `GET /analytics/overview/linkedin/ai_insights`

**Insight Types:** `publishing_behaviour`, `publishing_behaviour_impressions`, `publishing_behaviour_reach`, `audience_growth`, `page_views`, `top_posts`, `top_hashtags`, `city_demographics`, `country_demographics`, `industry_demographics`, `post_density`, `seniority_demographics`, `insights_summary`

### 26.6 TikTok AI Insights

**Route:** `POST /analytics/overview/tiktok/ai_insights`

**Insight Types:** `audience_growth`, `top_posts`, `daily_engagement`, `cumulative_engagement`, `daily_video_views`, `cumulative_video_views`, `engagement_vs_daily_posting`, `insights_summary`

### 26.7 Pinterest AI Insights

**Route:** `POST /analytics/overview/pinterest/ai_insights`

**Insight Types:** `daily_engagement`, `impressions_vs_posting_pattern`, `engagement_vs_posting_pattern`, `daily_pin_posting`, `daily_followers_trend`, `cumulative_followers_trend`, `impressions`, `engagement`, `top_and_least_posts`, `insights_summary`

### 26.8 Overview AI Insights

**Route:** `POST /analytics/overview/ai_insights`

**Insight Types:** `reach_across_platforms`, `engagement_across_platforms`, `impressions_across_platforms`, `platform_performance_comparison`, `overview_account_statistics`, `top_posts`

### 26.9 Error Handling

When data validation fails (zero values):
```json
{ "success": false, "message": "analytics.insufficient_data" }
```

Error message keys: `analytics.insufficient_data`, `analytics.insufficient_data_posts`, `analytics.insufficient_data_videos`, `analytics.insufficient_data_reels`, `analytics.invalid_insight_type`

---

## Appendix: Complete Column Reference by Table

### facebook_posts
`post_id`, `page_id`, `saving_time`, `created_time`, `media_type`, `status_type`, `video_id`, `category`, `published_by`, `published_by_url`, `shared_from_name`, `shared_from_id`, `shared_from_link`, `like`, `love`, `haha`, `wow`, `sad`, `angry`, `total`, `shares`, `comments`, `post_clicks`, `total_engagement`, `post_engaged_users`, `day_of_week`, `hour_of_day`, `updated_time`, `message_tags`, `post_metadata`, `caption`, `description`, `full_picture`, `link`, `permalink`, `post_impressions`, `post_impressions_unique`, `post_impressions_paid`, `post_impressions_paid_unique`, `post_impressions_organic`, `post_impressions_organic_unique`, `post_impressions_viral`, `post_impressions_viral_unique`, `post_video_views`, `total_impressions`

### facebook_insights
`page_id`, `hash_id`, `saving_time`, `created_time`, `page_fans`, `page_follows`, `page_impressions`, `page_impressions_paid`, `page_impressions_organic`, `page_post_engagements`, `page_fans_by_like`, `page_fans_by_unlike`, `talking_about_count`, `positive_sentiment`, `negative_sentiment`, `page_positive_feedback`, `page_negative_feedback`, `page_fans_online`, `page_fans_gender`, `page_fans_age`, `page_fans_gender_age`, `page_fans_country`, `page_fans_city`, `day_of_week`

### instagram_posts
`media_id`, `instagram_id`, `stored_event_at`, `post_created_at`, `media_type`, `entity_type`, `engagement`, `like_count`, `comments_count`, `saved`, `reach`, `impressions`, `views`, `shares`, `hashtags`, `reels_avg_watch_time`, `reels_total_watch_time`, `replies`, `exits`, `taps_forward`, `taps_back`, `permalink`, `caption`, `media_url`

### instagram_insights
`instagram_id`, `record_id`, `stored_event_at`, `created_time`, `online_users_datetime`, `followers_count`, `follows_count`, `profile_views`, `engagement`, `impressions`, `reach`, `accounts_engaged`, `online_followers`, `day_of_week`, `audience_age`, `audience_gender`, `audience_city`, `audience_country`

### linkedin_posts
`post_id`, `activity`, `linkedin_id`, `saving_time`, `published_at`, `created_at`, `media_type`, `day_of_week`, `favorites`, `comments`, `repost`, `post_clicks`, `total_engagement`, `impressions`, `reach`, `hashtags`, `title`, `image`, `article_url`

### linkedin_insights
`linkedin_id`, `record_id`, `inserted_at`, `created_at`, `totalFollowerCount`, `organicFollowerCount`, `paidFollowerCount`, `page_views`, `desktop_page_views`, `mobile_page_views`, `impressionCount`, `engagement`, `reach`, `repost`, `comments`, `reactions`, `unique_visitors`, `followers_by_seniority` (JSON), `followers_by_industry` (JSON), `followers_by_country` (JSON), `followers_by_city` (JSON)

### youtube_activity_insights
`record_id`, `channel_id`, `created_at`, `estimated_minutes_watched`, `average_view_duration`, `views`, `likes`, `dislikes`, `comments`, `shares`

### youtube_channels
`record_id`, `channel_id`, `subscriber_count`, `inserted_at`, `created_at`, `title`

### youtube_videos
`video_id`, `channel_id`, `published_at`, `inserted_at`, `title`, `description`, `duration`, `thumbnail_url`, `media_type`, `iframe_embed_html`, `likes`, `dislikes`, `views`, `red_views`, `favorites`, `comments`, `subscribers_gained`, `shares`, `minutes_watched`, `red_minutes_watched`, `average_view_duration`, `average_view_percentage`

### youtube_traffic_insights
`record_id`, `channel_id`, `created_at`, `subscriber_views`, `subscriber_watch_time`, `non_subsciber_watch_time`, `paid_views`, `annotation_views`, `end_screen_views`, `campaign_card_view`, `no_link_other_views`, `yt_channel_views`, `yt_search_views`, `related_video_views`, `yt_other_page_views`, `ext_url_views`, `playlist_views`, `notification_views`, `shorts_views`

### youtube_shared_insights
`channel_id`, `inserted_at`, `ameba`, `blogger`, `copy_paste`, `cyworld`, `digg`, `dropbox`, `embed`, `mail`, `whats_app`, `other`, `facebook_messenger`, `facebook_pages`, `facebook`, `fotka`, `vkontakte`, `google_plus`, `discord`, `linkedin`, `goo`, `hangouts`, `pinterest`, `myspace`, `reddit`, `skype`, `telegram`, `tumblr`, `twitter`, `viber`, `weibo`, `wechat`, `youtube`

### tiktok_posts
`tiktok_id`, `post_id`, `display_name`, `like_count`, `comments_count`, `share_count`, `engagement_count`, `engagement_rate`, `view_count`, `created_at`, `inserted_at`, `profile_link`, `cover_image_url`, `share_url`, `post_description`, `hashtags`, `duration`, `height`, `width`, `title`, `embed_html`, `embed_link`

### tiktok_insights
`tiktok_id`, `record_id`, `display_name`, `total_follower_count`, `total_following_count`, `total_video_views`, `total_video_likes`, `total_video_comments`, `total_video_shares`, `inserted_at`

### pinterest_pins
`pin_id`, `board_id`, `user_id`, `created_at`, `inserted_at`, `media_type`, `is_owner`, `title`, `description`, `board_owner`, `cover_image_url`, `dominant_color`, `creative_type`, `product_tags`, `height`, `width`

### pinterest_pin_insights
`pin_id`, `record_id`, `user_id`, `created_at`, `saving_time`, `inserted_at`, `impression`, `pin_clicks`, `outbound_click`, `saves`, `engagement`, `quartile_95s_percent_view`, `closeup`, `video_start`, `video_10s_view`, `video_avg_watch_time`

### pinterest_user_insights
`user_id`, `created_at`, `impression`, `pin_clicks`, `outbound_click`, `saves`, `engagement`

### pinterest_users
`user_id`, `inserted_at`, `follower_count`

### pinterest_boards
`board_id`, `user_id`, `inserted_at`, `follower_count`, `name`

### twitter_posts
`twitter_id`, `post_id`, `created_at`, `inserted_at`, `tweet_type`, `total_engagement`, `impressions`, `like_count`, `reply_count`, `retweet_count`, `quote_count`, `bookmark_count`, `url_link_clicks`, `user_profile_clicks`, `impression_count`, `hashtags`, `permalink`, `full_text`

### twitter_insights
`twitter_id`, `record_id`, `inserted_at`, `followers_count`, `following_count`, `tweet_count`, `listed_count`
