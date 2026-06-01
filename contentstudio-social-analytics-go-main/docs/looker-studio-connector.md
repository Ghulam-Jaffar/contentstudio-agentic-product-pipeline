# ContentStudio — Looker Studio Connector

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Connection Flow](#connection-flow)
4. [How Data Routing Works](#how-data-routing-works)
5. [Field Naming Convention](#field-naming-convention)
6. [Array Fields](#array-fields)
7. [Platform Reference](#platform-reference)
   - [Instagram](#instagram)
   - [Facebook](#facebook)
   - [LinkedIn](#linkedin)
   - [TikTok](#tiktok)
   - [YouTube](#youtube)
   - [Pinterest](#pinterest)
   - [X (Twitter)](#x-twitter)
   - [Google Business Profile (GMB)](#google-business-profile-gmb)
8. [Adding a New Platform](#adding-a-new-platform)
9. [Troubleshooting](#troubleshooting)

---

## Overview

The ContentStudio Looker Studio connector is a **Google Apps Script (GAS) Community Connector** that bridges ContentStudio's analytics API with Google Looker Studio. It lets users drag any analytics metric into a chart without writing SQL or building dashboards from scratch.

Each platform is a self-contained `.gs` file. All files share one global GAS scope, so helper functions defined in `main.gs` are available everywhere.

```
src/looker-studio/
├── main.gs        ← auth, config wizard, schema/data dispatch, shared utilities
├── facebook.gs    ← Facebook fields + fetchers
├── instagram.gs   ← Instagram fields + fetchers
├── linkedin.gs    ← LinkedIn fields + fetchers
├── tiktok.gs      ← TikTok fields + fetchers
├── youtube.gs     ← YouTube fields + fetchers
├── pinterest.gs   ← Pinterest fields + fetchers
├── twitter.gs     ← X (Twitter) fields + fetchers
└── gmb.gs         ← Google Business Profile fields + fetchers
```

---

## Architecture

```
Looker Studio chart request
        │
        ▼
   getData() [main.gs]
        │  resolves workspace timezone
        │  passes p.analytics_go, p.access_token, p.start_date, p.end_date
        ▼
  getData_{platform}(p)   [e.g. getData_tiktok]
        │  calls bestMatch() to choose which API endpoint to hit
        ▼
  analyticsGet(url, token)   [main.gs]
        │  GET request with Bearer token
        │  throws if HTTP ≠ 200 or response.status ≠ true
        ▼
  Go API server   (analytics/overview/{platform}/{endpoint})
        │  queries ClickHouse, returns JSON
        ▼
  fetch function maps JSON → array of row objects
        │  keys must match getFields_{platform} field IDs exactly
        ▼
  getData() in main.gs serialises rows → Looker Studio schema values
```

**Backend services** — the Go API (`cmd/api-server`) serves all analytics endpoints. Every response must include `"status": true` at the root level or the connector will throw an error.

---

## Connection Flow

### ContentStudio Modal

When a user clicks the **Looker Studio** button in the analytics tab bar, a modal opens showing:

1. **Auto-detected context** — Workspace, Platform, and Account pre-filled from the current view (with green checkmarks)
2. **How would you like to proceed?** (shown for Facebook, Instagram, LinkedIn only):
   - **Use Template** *(Recommended, pre-selected)* — copies a pre-built dashboard with all data wired in
   - **Start Fresh** — creates a new data source from scratch
3. **CTA button** — calls `GET /analytics/looker-studio/connect` and opens the returned URL in a new tab

### Template Copy Flow (Facebook / Instagram / LinkedIn)

When the user selects a template, the deep link format is:

```
https://lookerstudio.google.com/reporting/create
  ?c.reportId={templateId}
  &ds.ds0.connector=community
  &ds.ds0.connectorId={deploymentId}
  &ds.ds0.access_token={apiKey}
  &ds.ds0.workspace_id={workspaceId}
  &ds.ds0.platform={platform}
  &ds.ds0.account_id={accountId}
```

All params arrive in `request.configParams` on the first `getConfig()` call. The connector persists the token, pre-selects workspace/platform/account, calls `config.setIsSteppedConfig(false)`, and returns immediately — the user sees a pre-configured report copy with zero manual steps.

**Available templates:**

| Platform | Report ID |
|---|---|
| Facebook | `ff026271-696a-4bf2-8140-29115808d46e` |
| Instagram | `556f5ac6-f00d-40ad-8ff9-8bc5785fbab4` |
| LinkedIn | `5936a476-decc-4d3a-a933-bcebe31b4932` |

> `ds.ds0.connector` must always be the literal string `community`. The deployment ID goes in `ds.ds0.connectorId`. Putting the deployment ID in `connector` causes a "not a valid value" error in Looker Studio.

### Fresh Data Source Flow (all platforms)

When the user selects **Start Fresh** (or uses a platform with no template), the deep link is:

```
https://lookerstudio.google.com/datasources/create
  ?connectorId={deploymentId}
  &connectorConfig={urlEncodedJSON}
```

Where `connectorConfig` encodes `access_token`, `workspace_id`, `platform`, `account_id` as a JSON object.

### Manual Setup (no deep link)

Users who open the connector directly from Looker Studio see a **stepped config wizard**:

1. **API Key** — paste a ContentStudio API key
2. **Workspace** — dropdown populated by `GET /api/v1/workspaces`
3. **Platform** — static list (Instagram, Facebook, LinkedIn, TikTok, YouTube, Pinterest, X, GMB)
4. **Account** — dropdown populated by `GET /api/v1/workspaces/{id}/accounts`, filtered to selected platform

Once all steps are complete, Looker Studio calls `getSchema()` then `getData()` on every chart load.

---

## How Data Routing Works

Each platform exposes **one Looker Studio schema** (all possible fields) but has **multiple API endpoints** — one per chart type. The connector must figure out which endpoint to call for any given chart.

This is solved by `bestMatch()` in `main.gs`:

```javascript
function bestMatch(reqIds, fieldGroups) {
  // reqIds = the field IDs that the current chart is requesting
  // fieldGroups = { endpointKey: [field_id, field_id, ...], ... }
  // Returns the key whose field list overlaps reqIds the most
}
```

**How to design field groups correctly:**

- Every endpoint must have at least one **unique discriminating field** that no other endpoint shares. This guarantees unambiguous routing.
- Metrics that could appear in multiple endpoints (e.g. `date`) are not strong discriminators — they only break ties.
- `lp_` / `lt_` prefixes are used for "least posts / least tweets" metrics specifically so they don't collide with the same-named metrics in the "top posts" group.
- `ps_` / `tp_medass_` prefixes serve the same purpose in YouTube and Facebook respectively.

**Example** (TikTok):
```
Looker Studio chart requests: ['date', 'lp_like_count', 'lp_post_views']
bestMatch scores:
  topPosts   → 0 matches
  leastPosts → 2 matches  ← winner → calls tt_fetchLeastPosts
  summary    → 0 matches
  audience   → 1 match (date)
  daily      → 1 match (date)
```

---

## Field Naming Convention

All field IDs follow this pattern:

```
{api_response_root_key}_{exact_api_json_field}
```

**Examples:**

| API root key | API JSON field | GS field ID |
|---|---|---|
| `overview` (current) | `total_posts` | `sum_curr_total_posts` |
| `audience_growth` | `followers` | `audience_growth_followers` |
| `publishing_behaviour` | `total_posts` | `publishing_behaviour_total_posts` |
| `top_posts[]` | `created_time` | `top_posts_created_time` |

**Aliasing rule** — when the GS field name intentionally differs from the API field (due to legacy naming or clarity), the mapping is documented inline with a comment in the fetch function. Examples:

| GS field ID | API JSON field | Reason for alias |
|---|---|---|
| `phone_calls` (GMB) | `call_clicks` | Legacy naming from PHP API |
| `map_views` (GMB) | `maps_impressions` | Shorter alias |
| `video_avg_duration` (YouTube) | `average_view_duration` | Shorter alias |
| `video_title` (YouTube) | `title` | Namespaced to avoid collision |
| `like_count` (YouTube top posts) | `like` | PHP API uses singular form |

**Previous period** — for platforms that return current + previous period together, previous-period fields use a `prev_` prefix on the GS field ID.

---

## Array Fields

Some API responses include arrays per post (e.g. `media_assets`, `hashtags`, `product_tags`, `media_url`). Looker Studio fields are scalar, so arrays are **joined with a comma** into a single TEXT field.

**Example:**
```javascript
// API response: { "hashtags": ["#marketing", "#brand", "#social"] }
// GS field:
hashtags: (post.hashtags || []).join(',')
// Looker Studio value: "#marketing,#brand,#social"
```

**Facebook media assets** use the `tp_medass_` prefix (one field per asset property):

| GS field ID | API JSON field inside `media_assets[]` |
|---|---|
| `tp_medass_media_id` | `media_id` |
| `tp_medass_caption` | `caption` |
| `tp_medass_link` | `link` |
| `tp_medass_asset_type` | `assetType` |
| `tp_medass_call_to_action` | `callToAction` |
| `tp_medass_created_at` | `createdAt` |

Each field joins all assets in the post: `(post.media_assets || []).map(fn).join(',')`.

---

## Platform Reference

### Instagram

**API base:** `{analytics_go}instagram/`  
**Account param:** `instagram_id`

| Routing group | API endpoint | Discriminating fields | Returns |
|---|---|---|---|
| `summary` | `getPageAndPostsInsights` | `period`, `sum_curr_*`, `sum_prev_*` | One row with current + previous period metrics |
| `audience` | `getPageFollowersData` | `audience_growth_followers`, `audience_growth_followers_daily` | Daily follower count + daily change |
| `audienceRollup` | `getPageFollowersData` | `audience_growth_rollup_*` | Period rollup of follower count |
| `impTrend` | `getPageImpressions` | `impressions_impressions` | Daily impressions |
| `impRollup` | `getPageImpressions` | `impressions_rollup_*` | Period rollup of impressions |
| `engTrend` | `getPageEngagements` | `engagements_engagement` | Daily engagement + comments + reactions |
| `engRollup` | `getPageEngagements` | `engagements_rollup_*` | Period rollup of engagements |
| `pub` | `getPublishingBehavior` | `publishing_behaviour_total_posts` | Daily posts count + engagement metrics |
| `pubRollup` | `getPublishingBehavior` | `pub_beh_rlu_*` | Period rollup by media type |
| `stories` | `getStoriesPerformance` | `stories_performance_*` | Daily stories metrics |
| `storiesRollup` | `getStoriesPerformance` | `stories_rollup_*` | Period rollup of stories |
| `reels` | `getReelsPerformance` | `reels_*` | Daily reels metrics |
| `reelsRollup` | `getReelsPerformance` | `reels_rollup_*` | Period rollup of reels |
| `activeHours` | `getActiveUsers?type=activeHours` | `active_users_hours_*` | Hour-of-day activity heatmap |
| `activeDays` | `getActiveUsers?type=activeDays` | `active_users_days_*` | Day-of-week activity heatmap |
| `topPosts` | `getTopAndLeastPerformingMedia` | `top_posts_permalink`, `top_posts_like_count` | Top performing posts |
| `hashtags` | `getTopHashtags` | `top_hashtags_name` | Per-hashtag engagement |
| `hashtagsRollup` | `getTopHashtags` | `top_hashtags_rollup_*` | Period rollup of hashtag performance |
| `igLocCity` | `getDemographics?type=igLocCity` | `audience_city_city` | Audience city breakdown |
| `igLocCountry` | `getDemographics?type=igLocCountry` | `audience_country_country` | Audience country breakdown |
| `ageDemo` | `getDemographics?type=ageDemo` | `audience_age_bracket` | Audience age breakdown |
| `genderDemo` | `getDemographics?type=genderDemo` | `audience_gender_gender` | Audience gender breakdown |

**Field naming:** Instagram uses `{root}_{field}` throughout. Summary fields use `sum_curr_` and `sum_prev_` prefixes where root = `overview`.

---

### Facebook

**API base:** `{analytics_go}facebook/`  
**Account param:** `facebook_id`

| Routing group | API endpoint | Discriminating fields | Returns |
|---|---|---|---|
| `summary` | `getPageAndPostsInsights` | `period`, `sum_curr_*`, `sum_prev_*` | One row, current + previous period |
| `engagementTrend` | `getEngagement` | `engagement_page_engagements` | Daily page engagement |
| `impressionsTrend` | `getImpressions` | `impressions_*` | Daily impressions by type |
| `topPosts` | `topAndLeastPosts` | `top_posts_permalink`, `top_posts_like` | Top performing posts (50+ fields) |
| `reelsRollup` | `reelsOverview` | `reels_rollup_*` | Reels period comparison |
| `videoRollup` | `videoOverview` | `video_rollup_*` | Video period comparison |
| `videoDetails` | `videoDetails` | `video_details_*` | Per-video detail metrics |
| `fbLocCity` | `getDemographics?type=fbLocCity` | `audience_city_city` | Audience by city |
| `fbLocCountry` | `getDemographics?type=fbLocCountry` | `audience_country_country` | Audience by country |
| `ageDemo` | `getDemographics?type=ageDemo` | `audience_age_fans_age_bracket` | Audience by age |
| `genderDemo` | `getDemographics?type=genderDemo` | `audience_gender_gender` | Audience by gender |
| `active` | `getActiveUsersHours` / `getActiveUsersDays` | `active_users_days_*` | When fans are online |

**Top posts** — the `top_posts_` prefix is applied to all 50+ fields in the `top_posts[]` array. The `media_assets` array within each post is exposed as `tp_medass_*` fields (see [Array Fields](#array-fields)).

**Facebook summary API response structure:**

```json
{
  "status": true,
  "overview": {
    "current":  { "total_posts": 42, "post_likes": 1200, ... },
    "previous": { "total_posts": 38, "post_likes": 900,  ... }
  }
}
```

GS field `sum_curr_total_posts` ← `j.overview.current.total_posts`  
GS field `sum_prev_total_posts` ← `j.overview.previous.total_posts`

---

### LinkedIn

**API base:** `{analytics_go}linkedin/`  
**Account param:** `linkedin_id`

| Routing group | API endpoint | Discriminating fields | Returns |
|---|---|---|---|
| `summary` | `getPageAndPostsInsights` | `period`, `sum_curr_*`, `sum_prev_*` | One row, current + previous period |
| `audience` | `getPageFollowersData` | `audience_growth_total_follower_count` | Daily follower counts (total/organic/paid) |
| `audienceRollup` | `getPageFollowersData` | `audience_growth_rollup_*` | Period rollup of followers |
| `pub` | `getPublishingBehavior` | `publishing_behaviour_total_posts` | Daily posts + engagement |
| `pubRollup` | `getPublishingBehavior` | `pub_beh_rlu_*` | Period rollup by media type |
| `pageViews` | `getPageViews` | `page_views_total_page_views_daily` | Daily page views (desktop + mobile) |
| `pageViewsRollup` | `getPageViews` | `page_views_rollup_*` | Period rollup of page views |
| `postsPerDay` | `getPostsPerDay` | `day_of_week`, `posts_per_days_posts` | Posts count by day of week |
| `topPosts` | `getTopAndLeastPerformingPosts` | `top_posts_post_id`, `top_posts_favorites` | Top performing posts |
| `hashtags` | `getTopHashtags` | `hashtag`, `top_hashtags_posts` | Per-hashtag performance |
| `hashtagsRollup` | `getTopHashtags` | `top_hashtags_rollup_*` | Period rollup of hashtags |
| `demographics` | `getFollowerDemographics` | `follower_demographics_category` | Audience by function/industry/seniority/country/city |

**LinkedIn top posts** use `top_posts_` prefix. The `top_posts_poll_data` and `top_posts_media` fields contain JSON-stringified complex objects (arrays/objects not suitable for scalar exposure).

---

### TikTok

**API base:** `{analytics_go}tiktok/`  
**Account param:** `tiktok_id`

| Routing group | API endpoint | Discriminating fields | Returns |
|---|---|---|---|
| `summary` | `getPageAndPostsInsights` | `period`, `eng_rate`, `followers`, `total_followings` | One row, current + previous diffs + growth % |
| `audience` | `getPageFollowersAndViews` | `followers`, `followers_daily`, `profile_views` | Daily followers + profile views |
| `daily` | `getDailyEngagementsData` | `likes`, `comments`, `shares`, `engagement` | Daily engagement + cumulative totals |
| `postsEngagements` | `getPostsAndEngagements` | `video_views`, `total_posts` | Daily video views + post count |
| `topPosts` | `getTopAndLeastPerformingPosts` | `permalink`, `like_count`, `engagement_rate` | Top posts (full detail) |
| `leastPosts` | `getTopAndLeastPerformingPosts` | `lp_like_count`, `lp_engagement_rate` | Least performing posts |

**TikTok summary API response structure:**

```json
{
  "status": true,
  "data": {
    "total_followers": 5000,
    "total_followings": 120,
    "total_likes": 45000,
    "total_likes_diff": 500,
    "total_likes_growth": 0.012,
    ...
  }
}
```

**Field mapping for summary (all from `j.data`):**

| GS field ID | API field | Notes |
|---|---|---|
| `followers` | `total_followers` | |
| `total_followings` | `total_followings` | |
| `video_views` | `total_video_views` | |
| `engagement` | `total_engagements` | |
| `total_likes_diff` | `total_likes_diff` | Current minus previous |
| `total_likes_growth` | `total_likes_growth` | Decimal, e.g. 0.05 = 5% |

**TikTok top/least posts** — both call `getTopAndLeastPerformingPosts`. Top posts are at `j.data.top_posts[]`, least at `j.data.least_posts[]`. Both use the same `PostRow` shape:

| GS field (top) | GS field (least) | API JSON field |
|---|---|---|
| `like_count` | `lp_like_count` | `likes_count` |
| `comments_count` | `lp_comments_count` | `comments_count` |
| `post_views` | `lp_post_views` | `views_count` |
| `post_shares` | `lp_post_shares` | `shares_count` |
| `post_engagement` | `lp_post_engagement` | `engagements_count` |
| `engagement_rate` | `lp_engagement_rate` | `engagement_rate` |
| `duration` | `lp_duration` | `duration` |
| `height` | `lp_height` | `height` |
| `width` | `lp_width` | `width` |
| `hashtags` (joined) | `hashtags` (joined) | `hashtags[]` |

---

### YouTube

**API base:** `{analytics_go}youtube/`  
**Account param:** `youtube_id`

| Routing group | API endpoint | Discriminating fields | Returns |
|---|---|---|---|
| `summary` | `overviewSummary` | `period`, `subscribers`, `views`, `watch_time` | One row, current + previous period |
| `subscribers` | `overviewSubscriberTrend` | `subscribers_gained_daily`, `subscribers_total` | Daily subscriber gains |
| `engTrend` | `overviewEngagementTrend` | `like_daily`, `dislike_daily`, `engagement_daily` | Daily + cumulative totals |
| `views` | `overviewViewsTrend` | `subscriber_views_daily`, `video_views_daily` | Daily + cumulative view counts |
| `watchTime` | `overviewWatchTimeTrend` | `subscriber_watch_time_daily`, `average_watch_time` | Daily + cumulative watch time |
| `findVideo` | `overviewFindVideo` | `traffic_source`, `source_value` | Traffic source breakdown |
| `videoSharing` | `overviewVideoSharing` | `sharing_platform`, `source_value` | Sharing platform breakdown |
| `topPosts` | `getSortedTopPosts` | `video_id`, `like_count`, `video_views` | Top videos (full detail) |
| `leastPosts` | `overviewLeastPosts` | `lp_sort_by`, `lp_like_count`, `lp_video_views` | Least performing videos |
| `perfSchedule` | `overviewPerformanceAndVideoPostingSchedule` | `ps_count`, `ps_likes`, `ps_engagement` | Daily performance by publish date |

**YouTube summary API response structure:**

```json
{
  "status": true,
  "overview": {
    "current":  { "subscribers": 10000, "like": 500, "dislike": 5, "comment": 80, ... },
    "previous": { "subscribers": 9500, "like": 450, ... }
  }
}
```

**Important aliases** — YouTube API uses PHP-compatible singular forms:

| GS field ID | API JSON field | Notes |
|---|---|---|
| `likes` | `like` | Singular in API |
| `dislikes` | `dislike` | Singular in API |
| `comments` | `comment` | Singular in API |
| `shares` | `share` | Singular in API |
| `total_posts` | `videos` | Different name entirely |
| `video_avg_duration` | `average_view_duration` | Shortened alias |
| `video_title` | `title` | Namespaced to avoid collision |
| `permalink` | `share_url` | Different name entirely |

**videoSharing vs findVideo routing** — both have `source_value` and `source_perc`. `sharing_platform` appears only in `videoSharing`, so it always wins when requested alongside `source_value`. `traffic_source` appears only in `findVideo`. This guarantees unambiguous routing.

**leastPosts** — the API returns `j.least_posts_ordered_by_views[]` and `j.least_posts_ordered_by_engagement[]`. Both arrays are concatenated into one result set, with `lp_sort_by` set to `"views"` or `"engagement"` so charts can filter or distinguish them.

---

### Pinterest

**API base:** `{analytics_go}pinterest/`  
**Account param:** `pinterest_id`  
**Optional param:** `board_id` (when scoped to a board)

| Routing group | API endpoint | Discriminating fields | Returns |
|---|---|---|---|
| `summary` | `overviewSummary` | `period`, `follower_count`, `total_engagement` | One row, current + previous period |
| `followers` | `overviewFollowers` | `followers_daily`, `followers_gained` | Daily follower change |
| `impTrend` | `overviewImpressions` | `impressions_daily`, `impressions_total` | Daily + cumulative impressions |
| `engTrend` | `overviewEngagement` | `saves_daily`, `engagement_daily` | Daily + cumulative engagement |
| `rollup` | `overviewPinPostingRollup` | `total_pins`, `video_10s_view`, `quartile_95s_percent_view` | Pin aggregate for period |
| `topPins` | `overviewTopPins` | `pin_impressions`, `pin_engagement`, `pin_id` | Top pins (full detail) |
| `performance` | `overviewPinPostingPerformance` | `pin_performance_engagement`, `pin_performance_impressions` | Daily pin performance |
| `pinPosting` | `overviewPinPostingPerDay` | `pins_posting_count` | Daily pins posted count |

**Pinterest summary API response structure:**

```json
{
  "status": true,
  "overview": {
    "current":  { "follower_count": 2000, "impressions": 50000, "pin_clicks": 300, ... },
    "previous": { "follower_count": 1800, "impressions": 45000, ... }
  }
}
```

**Top pins** — fields come from `j.top[]` (a `PinItem` array). All fields use the `pin_` prefix to avoid collision with account-level metrics:

| GS field ID | API JSON field | Notes |
|---|---|---|
| `pin_impressions` | `impressions` | Pin-level, not account-level |
| `pin_saves` | `saves` | Pin-level |
| `pin_clicks_top` | `pin_clicks` | Suffix `_top` avoids collision with `pin_clicks` in other groups |
| `pin_outbound_clicks` | `outbound_clicks` | Pin-level |
| `pin_engagement` | `total_engagement` | Pin-level |
| `pin_engagement_rate` | `engagement_rate` | |
| `pin_product_tags` | `product_tags[]` | Array joined with comma |

**Rollup additional fields:**

| GS field ID | API JSON field | Notes |
|---|---|---|
| `quartile_95s_percent_view` | `quartile_95s_percent_view` | Video quartile metric (PERCENT type) |
| `video_10s_view` | `video_10s_view` | 10-second video views |

**Pin performance vs pin posting** — `performance` has `pin_performance_engagement` and `pin_performance_impressions` (unique to that group). `pinPosting` has `pins_posting_count` (unique). Both avoid collision with `pins_count` (used in the `performance` group itself).

---

### X (Twitter)

**API base:** `{analytics}twitter/`  
**Note:** Twitter uses the Node.js pipeline URL (`p.analytics`), not the Go service URL (`p.analytics_go`)  
**Account param:** `twitter_id`

| Routing group | API endpoint | Discriminating fields | Returns |
|---|---|---|---|
| `summary` | `getPageAndPostsInsights` | `period`, `followers_count`, `eng_rate` | One row with diffs + growth % |
| `audience` | `getFollowersTrendData` | `followers`, `followers_daily`, `following_count_daily` | Daily follower counts |
| `engagement` | `getEngagementImpressionData` | `engagement`, `impressions`, `tweets_daily` | Daily engagement + impressions |
| `topTweets` | `getTopTweets` | `permalink`, `like_count`, `tweet_id` | Top tweets (full detail) |
| `leastTweets` | `getLeastTweets` | `lt_permalink`, `lt_like_count`, `lt_tweet_id` | Least performing tweets |
| `credits` | `getCreditsUsedCount` | `credits_used` | Credits consumed in date range |

**Twitter summary** — `j.data` is a flat object. All fields are mapped directly:

| GS field ID | API JSON field | Notes |
|---|---|---|
| `followers_count` | `followers_count` | Total followers at period end |
| `eng_rate` | `eng_rate` | Engagement rate (PERCENT) |
| `followers_count_diff` | `followers_count_diff` | Diff vs previous period |
| `followers_count_growth` | `followers_count_growth` | Growth % |

**Top/Least tweets** — both come from a `Tweet` struct. Top uses `j.top_tweets[]`, least uses `j.least_tweets[]`. Least uses `lt_` prefix on all metrics:

| GS field (top) | GS field (least) | API JSON field |
|---|---|---|
| `permalink` | `lt_permalink` | `permalink` |
| `tweet_id` | `lt_tweet_id` | `id` |
| `tweet_text` | `lt_tweet_text` | `tweet_text` |
| `tweet_type` | `lt_tweet_type` | `tweet_type` |
| `tweet_media_url` (joined) | `lt_tweet_media_url` (joined) | `media_url[]` |
| `like_count` | `lt_like_count` | `like_count` |
| `impressions` | `lt_impressions` | `impression_count` |
| `listed_count` | `lt_listed_count` | `listed_count` |

**Engagement trend** — `j.tweeted_at_date[]` is the date bucket array. Daily counts map directly:

| GS field ID | API JSON field |
|---|---|
| `engagement` | `total_engagement[]` |
| `impressions` | `impression_count[]` |
| `tweets_daily` | `tweet_count[]` |
| `bookmarks_daily` | `bookmark_count[]` |
| `quotes_daily` | `quote_count[]` |

---

### Google Business Profile (GMB)

**API base:** `{analytics_go}gmb/`  
**Account param:** `gmb_id`

| Routing group | API endpoint | Discriminating fields | Returns |
|---|---|---|---|
| `summary` | `summary` | `period`, `review_count`, `avg_rating` | One row, current + previous period |
| `daily` | `impressions` | `search_views`, `map_views`, `total_views` | Daily impression counts |
| `actions` | `actions` | `direction_requests`, `phone_calls`, `website_clicks` | Daily action counts |
| `impRollup` | `impressions` | `imp_total`, `imp_desktop_maps` | Period rollup of impressions |
| `actRollup` | `actions` | `act_call_clicks`, `act_website_clicks` | Period rollup of actions |
| `media` | `mediaActivity` | `photo_views`, `video_count` | Daily media activity |
| `mediaRollup` | `mediaActivity` | `ma_total_photos`, `ma_total_videos` | Period rollup of media |
| `reviews` | `reviews` | `new_reviews`, `prev_review_count` | Review rollup |
| `keywords` | `searchKeywords` | `keyword`, `keyword_impressions` | Top search keywords |
| `topPosts` | `topPosts` | `post_name`, `topic_type`, `state` | Recent GMB posts |
| `pubBehav` | `publishingBehavior` | `total_posts` (with `date`) | Daily posts count |

**GMB summary API response structure:**

```json
{
  "status": true,
  "overview": {
    "current":  { "search_impressions": 500, "call_clicks": 30, "total_posts": 5, ... },
    "previous": { "search_impressions": 450, "call_clicks": 25, ... }
  }
}
```

**Important aliases** (GS field ← API field):

| GS field ID | API JSON field | Reason |
|---|---|---|
| `search_views` | `search_impressions` | Legacy PHP name |
| `map_views` | `maps_impressions` | Legacy PHP name |
| `total_views` | `total_impressions` | Legacy PHP name |
| `phone_calls` | `call_clicks` | Legacy PHP name |
| `review_count` | `total_reviews` | Shorter alias |
| `avg_rating` | `average_rating` | Shorter alias |

**GMB top posts** — `j.posts[]` is a `TopPost` array. Array fields are joined with comma:

| GS field ID | API JSON field | Notes |
|---|---|---|
| `tp_media_names` | `media_names[]` | Joined |
| `tp_media_formats` | `media_formats[]` | Joined |
| `tp_media_google_urls` | `media_google_urls[]` | Joined |

**Keywords** — `j.keywords[]` is a `SearchKeyword` array:

| GS field ID | API JSON field |
|---|---|
| `keyword_impressions` | `impressions_value` |
| `impressions_threshold` | `impressions_threshold` |
| `keyword_month` | `keyword_month` |

---

## Adding a New Platform

1. **Create `{platform}.gs`** in `src/looker-studio/`:
   - `getFields_{platform}(fields, types)` — declare all dimensions and metrics using `{api_root}_{api_field}` naming
   - `buildUrl_{abbr}(endpoint, p, extra)` — constructs `{analytics_go}{platform}/{endpoint}` + common params
   - `getData_{platform}(p)` — calls `bestMatch` and dispatches to fetch functions
   - One `{abbr}_fetch*()` function per API endpoint

2. **Register in `main.gs`**:
   - Add `case '{platform}': return getFields_{platform}(fields, types);` in `buildFields()`
   - Add `case '{platform}': rows = getData_{platform}(p); break;` in `getData()`
   - Add the platform option in `getConfig()` step 3
   - Add the platform entry in `getAccountsForPlatform()` with the correct `idField`

3. **Ensure the Go backend** returns `"status": true` in every response — `analyticsGet()` throws on missing or false status.

4. **Design discriminating fields** — each routing group in `bestMatch` must have at least one unique field ID not shared by any other group.

---

## Troubleshooting

### "Bad API response" error in Looker Studio

The connector checks `if (!json.status) throw ...` on every API response. Causes:
- Go handler does not set `"status": true` in its response struct
- API returned an error payload (`{"status": false, "message": "..."}`)
- Date range has no data and the handler returns an empty object without `status`

**Fix:** Ensure every Go response struct has `Status bool \`json:"status"\`` and every return path sets `Status: true`.

### Chart shows no data / zeros

- The `bestMatch` routing picked the wrong endpoint. Add a more distinctive field to the intended group.
- The API field name in the fetch function doesn't match the actual JSON key. Check the Go response struct's `json:""` tags.
- The date range has no data in ClickHouse for the account.

### Wrong endpoint called

Run in GAS console with `console.log('best:', best)` inside `getData_{platform}` to see which group won. Increase overlap of the correct group or add a unique field that only that chart would request.

### Auth errors

- Token may have expired. User should regenerate from ContentStudio's Data Studio page and reconnect.
- Token stored in `UserProperties` may be stale. In GAS, clear with `PropertiesService.getUserProperties().deleteProperty('cs_token')`.

### Timezone issues

The connector resolves workspace timezone on every `getData()` call via `POST /fetchUserWorkspaces`. If that call fails, it silently falls back to UTC. All date fields in responses must be in `YYYYMMDD` format (no separators) for Looker Studio's `YEAR_MONTH_DAY` type.
