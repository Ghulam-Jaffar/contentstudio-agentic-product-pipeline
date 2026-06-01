# Social Listening — Backend PRD

**Epic:** [Social Listening #112115](https://app.shortcut.com/contentstudio-team/epic/112115)
**Date:** March 11, 2026
**Service:** Extends `contentstudio-social-analytics-go`
**Data Provider:** Data365 API (6 platforms)

> This PRD is written for an AI agent to implement the full backend and pipeline. It contains every schema, model, endpoint, config, and architectural decision needed to build the feature end-to-end.

---

## Table of Contents

1. [Overview](#1-overview)
2. [Data365 API Integration](#2-data365-api-integration)
3. [MongoDB Schema — listening_topics](#3-mongodb-schema--listening_topics)
4. [ClickHouse Schema — listening_mentions](#4-clickhouse-schema--listening_mentions)
5. [Kafka Topics & Message Schemas](#5-kafka-topics--message-schemas)
6. [Pipeline Architecture](#6-pipeline-architecture)
7. [Deduplication Strategy](#7-deduplication-strategy)
8. [Sentiment & AI Enrichment](#8-sentiment--ai-enrichment)
9. [Alerts](#9-alerts)
10. [Usage Limits & Billing](#10-usage-limits--billing)
11. [API Endpoints](#11-api-endpoints)
12. [Directory Structure](#12-directory-structure)
13. [Configuration](#13-configuration)
14. [Resolved Decisions](#14-resolved-decisions)
15. [Open Questions](#15-open-questions)

---

## 1. Overview

### What We're Building

A social listening pipeline that monitors 6 social platforms for brand mentions, keywords, and competitor activity. Data is fetched daily (with on-demand triggers), normalized, enriched with AI sentiment analysis, and stored in ClickHouse for querying by the frontend.

> Scope note: This document is the current v1 source of truth. The 18-platform expansion in `social-listing-data-pipeline-prd.md` is roadmap material and does not change the 6-platform Data365 scope described here.

### Platforms (6)

| # | Platform | Data365 Search Endpoint | Max Docs/Query | Search Scope |
|---|---|---|---|---|
| 1 | TikTok | `data365_tiktok_keywords` | 500 | Video caption/description text |
| 2 | Instagram | Hashtag search only | varies | **No keyword search** — hashtag-based only (e.g., `#acmecorp`) |
| 3 | Facebook | `data365_facebook_posts_search` | 300 | Post text (Top/Latest/Hashtag modes) |
| 4 | Twitter/X | `data365_twitter_keywords` | 20,000 | Tweet body text (includes replies & quote tweets) |
| 5 | Reddit | keyword search | varies | Post title + body text |
| 6 | Threads | keyword search | varies | Post text (`parent_post_id` set for replies) |

### Architecture Pattern

Extends the existing Go analytics service (`contentstudio-social-analytics-go`). Same 5-stage Kafka pipeline pattern:

```
Scheduler → Kafka Work Order → Fetcher → Parser → Sentiment → Sink → ClickHouse
```

### Trigger Modes

| Mode | Trigger | Lookback |
|---|---|---|
| **Daily** | Cron job (scheduler) | Last 24 hours |
| **On-demand** | `POST /api/v1/listening-work` (user clicks refresh) | Last 24 hours |
| **Initial** | Auto-triggered on first topic creation | Last 30 days |

---

## 2. Data365 API Integration

### API Workflow

Data365 uses an **async POST-then-GET** pattern:

```
Step 1: POST /v1.1/{platform}/search?access_token=TOKEN
        Body: { "query": "keyword", "max_documents": 300 }
        → Returns: { "data": { "task_id": "abc123" }, "status": "ok" }

Step 2: Poll GET /v1.1/{platform}/search/{task_id}/status?access_token=TOKEN
        → Wait until status = "finished" (1-5 minutes)

Step 3: GET /v1.1/{platform}/search/{task_id}/posts?access_token=TOKEN
        → Returns paginated results (100 items/page, cursor-based)
```

### Rate Limits

- Up to **100 requests/second**
- Pagination returns max **100 elements per page**
- Async data collection latency: **1-5 minutes**

### Credit System

| Operation | Credits |
|---|---|
| Search request | 7 |
| Single post retrieval | 1 |
| Post with comments | 5 |
| Profile with info | 9 |

### Average Credits Per Topic (Daily Pipeline)

Assuming 2 keywords per topic, searched on all 6 platforms:

| Scenario | Mentions/Day | Daily Credits | Monthly (1 Topic) | Monthly (5 Topics) |
|---|---|---|---|---|
| **Low volume** | ~50 | ~334 | ~10,000 | ~50,000 |
| **Medium volume** | ~200 | ~1,100 | ~33,000 | ~165,000 |
| **High volume** | ~500 | ~2,640 | ~79,000 | ~396,000 |

Credit breakdown per daily run:
- Search: 2 keywords × 6 platforms × 7 credits = **84 credits** (fixed)
- Post retrieval: mentions × 1 credit each
- Pagination: negligible (<5% of total)

### Data365 Search Scope & Limitations

#### What Data365 searches

Data365 keyword search operates on **post-level content only**. It does NOT search within comments.

| Platform | What's Searched | What's Returned | Notes |
|---|---|---|---|
| **Twitter/X** | Tweet body text | Posts, replies, quote tweets | Replies and quote tweets are top-level tweets on Twitter — they appear as separate search results with `tweet_type` = `REPLY` or `QUOTE` |
| **Facebook** | Post text | Posts only | Supports `search_type`: Top, Latest, Hashtag |
| **TikTok** | Video caption/description | Posts only | Hashtag-based and music-based searches are separate endpoints |
| **Reddit** | Post title + body text | Posts only | Supports `subreddit:` filter, sort (new/top/hot/relevance) |
| **Threads** | Post text | Posts only | `parent_post_id` distinguishes top-level posts from replies |
| **Instagram** | **Hashtags only** | Posts matching hashtag | **No broad keyword search.** Only hashtag-based search and per-profile feed. Users must add hashtag keywords (e.g., `#acmecorp`) |

#### Why we skip comments (v1 decision)

**Decision: Fetch posts only (1 credit each). Do not fetch comments (5 credits each).**

Rationale:
1. **Data365 cannot search within comments** — keyword search matches post text only. Comments are a separate fetch endpoint (`/{platform}/post/{post_id}/comments`) that returns ALL comments on a post, not just ones matching the keyword
2. **5x credit cost** — fetching post with comments costs 5 credits vs 1 credit for post-only. At medium volume (200 mentions/day, 5 topics), this increases monthly credit usage from ~165K to ~825K — approaching plan limits
3. **Low signal-to-noise** — bulk-fetched comments are not filtered by keyword, so most would be irrelevant noise in the mention feed
4. **Industry standard** — Brand24, Mention, and Sprout Social primarily surface posts that match keywords, not all comments on those posts
5. **Can be added later** — a v2 enhancement could selectively fetch comments for high-engagement posts or let users expand comments on a specific mention in the UI

#### Instagram limitation

Instagram does NOT support broad keyword search through Data365. Only two search modes are available:
- **Hashtag search** — matches hashtags in captions (e.g., `#acmecorp`)
- **Profile feed** — fetches posts from a specific profile

**Implication for users:** When configuring a listening topic, Instagram keywords must be hashtags. The topic creation UI should guide users to add hashtag variants of their brand (e.g., `#acme`, `#acmecorp`, `#acmeproducts`). Non-hashtag keywords will not return Instagram results.

The parser should detect that a keyword has no `#` prefix when building the Data365 Instagram search request, and auto-prepend `#` or skip Instagram for that keyword.

#### Twitter/X replies and quote tweets

Data365 returns Twitter replies and quote tweets as **separate mention objects** in search results (each with its own unique `id`, `body`, `author`, and engagement metrics).

**Decision: Show each reply/quote tweet as a separate mention in the feed.**

Rationale:
- Each reply carries independent sentiment and context — a reply saying "Acme is terrible" is a different signal from the parent tweet "Acme launched today"
- Each has its own author, engagement metrics, and potential for influencer detection
- Data365 returns them as separate objects with unique IDs — no extra work to separate
- Sprout Social and Brand24 treat each reply/quote as a standalone mention
- The `content_type` field captures the distinction: `post`, `reply`, `quote` (mapped from Data365's `tweet_type`)

### Keyword Highlighting

The `matched_keywords` array stored per mention in ClickHouse enables frontend keyword highlighting.

**Backend responsibility:**
1. Parser stage detects which positive topic terms from `include_any` and `include_all` appear in the post text
2. Stores the matched keywords in `matched_keywords` Array(String) column
3. Laravel API returns `matched_keywords` with each mention in the response

**Frontend responsibility:**
1. Receives `matched_keywords` array with each mention
2. Wraps each keyword occurrence in the `text` field with a highlight element (e.g., `<mark>` or styled `<span>`)
3. Case-insensitive matching — "acme corp" highlights "Acme Corp", "ACME CORP", etc.

**Example API response:**
```json
{
  "mention_id": "twitter:123456",
  "text": "Just tried Acme Corp's new product and it's amazing! #acme",
  "matched_keywords": ["acme corp", "#acme"],
  "platform": "twitter",
  "content_type": "post"
}
```

**Frontend renders:**
```
Just tried <mark>Acme Corp</mark>'s new product and it's amazing! <mark>#acme</mark>
```

**Parser keyword matching logic:**
```go
func detectMatchedKeywords(text string, terms []string, caseSensitive bool) []string {
    candidate := text
    if !caseSensitive {
        candidate = strings.ToLower(text)
    }

    var matched []string
    for _, term := range terms {
        probe := term
        if !caseSensitive {
            probe = strings.ToLower(term)
        }
        if strings.Contains(candidate, probe) {
            matched = append(matched, term)
        }
    }
    return matched
}
```

### Data365 Response Fields Per Platform

#### Facebook

```json
{
  "id": "string",
  "text": "string",
  "created_time": "ISO 8601",
  "post_type": "image|video|link|status",
  "owner_id": "string",
  "owner_username": "string",
  "owner_full_name": "string",
  "attached_image_url": "string",
  "comments_count": 0,
  "shares_count": 0,
  "reactions_total_count": 0,
  "reactions_like_count": 0,
  "reactions_love_count": 0,
  "reactions_haha_count": 0,
  "reactions_wow_count": 0,
  "reactions_sad_count": 0,
  "reactions_angry_count": 0,
  "reactions_support_count": 0,
  "text_tagged_users": ["string"]
}
```

#### Instagram

```json
{
  "id": "string",
  "shortcode": "string",
  "created_time": "ISO 8601",
  "product_type": "feed|reels|stories",
  "text": "string (caption)",
  "text_tags": ["hashtags"],
  "text_tagged_users": ["string"],
  "attached_media_display_url": "string",
  "attached_video_url": "string",
  "attached_carousel_media_urls": ["string"],
  "is_video": true,
  "likes_count": 0,
  "comments_count": 0,
  "owner_id": "string",
  "owner_username": "string",
  "status": "EXIST"
}
```

#### Twitter/X

```json
{
  "id": "string",
  "body": "string",
  "published": "ISO 8601",
  "tweet_type": "POST|REPLY|QUOTE",
  "retweet_type": "NONE|RETWEET",
  "likes_count": 0,
  "favorites": 0,
  "quote_count": 0,
  "reply_count": 0,
  "retweet_count": 0,
  "links": ["string"],
  "language": "en",
  "link": "https://twitter.com/...",
  "author": { "handle": "string", "url": "string" },
  "user_id": "string"
}
```

#### Threads

```json
{
  "id": 123456789,
  "shortcode": "string",
  "post_url": "string",
  "text": "string",
  "created_time": "ISO 8601",
  "timestamp": 1234567890,
  "owner_id": 123456789,
  "owner_username": "string",
  "owner_full_name": "string",
  "owner_profile_pic_url": "string",
  "likes_count": 0,
  "comments_count": 0,
  "reposts_count": 0,
  "reshares_count": 0,
  "quotes_count": 0,
  "parent_post_id": null,
  "attached_medias": [],
  "text_tagged_users": []
}
```

#### TikTok (profile-level; post fields inferred)

```json
{
  "username": "string",
  "full_name": "string",
  "avatar_url": "string",
  "is_verified": true,
  "follower_count": 0,
  "heart_count": 0,
  "video_count": 0
}
```

#### Reddit (inferred)

```json
{
  "title": "string",
  "body": "string",
  "author": "string",
  "subreddit": "string",
  "score": 0,
  "upvotes": 0,
  "comment_count": 0,
  "permalink": "string",
  "created_utc": 1234567890,
  "is_nsfw": false,
  "media": null
}
```

---

## 3. MongoDB Schema — `listening_topics`

Collection: `listening_topics`

```javascript
{
  // Identity
  "_id": ObjectId,
  "workspace_id": ObjectId,
  "owner_user_id": ObjectId,                     // Billing owner/admin who owns this topic slot
  "created_by": ObjectId,
  "name": "Brand Monitoring - Acme Corp",

  // Topic Type
  "type": "brand" | "competitor" | "industry" | "custom",

  // Search Configuration
  "query": {
    "include_any": ["acme corp", "acmecorp", "#acme"],        // OR logic, also used as Data365 search seeds
    "include_all": ["launch"],                                 // AND logic, parser-stage post-filter
    "exclude_keywords": ["acme hardware", "acme roadrunner"], // NOT logic
    "include_authors": ["@acmecorp"],                         // Optional allowlist
    "exclude_authors": ["@acme_bot"],                         // Optional topic-level blocklist
    "language": ["en", "fr", "de"],                           // ISO 639-1, empty = all
    "regions": ["US", "GB"],                                  // ISO 3166-1, empty = all
    "exact_match": false,
    "case_sensitive": false
  },

  // Platform Filters (only Data365 platforms)
  "platforms": {
    "tiktok": true,
    "instagram": true,
    "facebook": true,
    "twitter": true,
    "reddit": true,
    "threads": true,
  },

  // State
  "status": "active" | "paused" | "deleted",
  "is_initial_sync_done": false,
  "mentions_limit_reached": false,

  // Pipeline Tracking
  "last_fetched_at": ISODate,
  "last_fetched_cursors": {
    "tiktok": { "cursor": "string", "last_post_time": ISODate },
    "instagram": { "cursor": "string", "last_post_time": ISODate },
    "facebook": { "cursor": "string", "last_post_time": ISODate },
    "twitter": { "cursor": "string", "last_post_time": ISODate },
    "reddit": { "cursor": "string", "last_post_time": ISODate },
    "threads": { "cursor": "string", "last_post_time": ISODate },
  },

  // Usage Tracking (operational topic counters; billing source of truth lives in Laravel)
  "usage": {
    "current_period_start": ISODate,
    "mentions_count": 0,
    "mentions_limit": 10000,
  },

  // Timestamps
  "created_at": ISODate,
  "updated_at": ISODate,
  "deleted_at": null,
}
```

### Indexes

```javascript
// Scheduler: fetch active topics needing update
{ "status": 1, "mentions_limit_reached": 1, "last_fetched_at": 1 }

// Topic slot enforcement per billing owner/admin
{ "owner_user_id": 1, "status": 1 }

// Topic listing/filtering per workspace
{ "workspace_id": 1, "status": 1 }

// Lookup by ID (default _id index covers this)
```

### Go Model

```go
// File: src/models/db/mongo/listening_topic.go

type ListeningTopic struct {
    ID                   primitive.ObjectID            `bson:"_id,omitempty"`
    WorkspaceID          primitive.ObjectID            `bson:"workspace_id"`
    OwnerUserID          primitive.ObjectID            `bson:"owner_user_id"`
    CreatedBy            primitive.ObjectID            `bson:"created_by"`
    Name                 string                        `bson:"name"`
    Type                 string                        `bson:"type"`
    Query                ListeningQuery                `bson:"query"`
    Platforms            map[string]bool               `bson:"platforms"`
    Status               string                        `bson:"status"`
    IsInitialSyncDone    bool                          `bson:"is_initial_sync_done"`
    MentionsLimitReached bool                          `bson:"mentions_limit_reached"`
    LastFetchedAt        *time.Time                    `bson:"last_fetched_at,omitempty"`
    LastFetchedCursors   map[string]PlatformCursor     `bson:"last_fetched_cursors"`
    Usage                TopicUsage                    `bson:"usage"`
    CreatedAt            time.Time                     `bson:"created_at"`
    UpdatedAt            time.Time                     `bson:"updated_at"`
    DeletedAt            *time.Time                    `bson:"deleted_at,omitempty"`
}

type ListeningQuery struct {
    IncludeAny      []string `bson:"include_any"`
    IncludeAll      []string `bson:"include_all,omitempty"`
    ExcludeKeywords []string `bson:"exclude_keywords"`
    IncludeAuthors  []string `bson:"include_authors,omitempty"`
    ExcludeAuthors  []string `bson:"exclude_authors,omitempty"`
    Language        []string `bson:"language,omitempty"`
    Regions         []string `bson:"regions,omitempty"`
    ExactMatch      bool     `bson:"exact_match"`
    CaseSensitive   bool     `bson:"case_sensitive"`
}

type PlatformCursor struct {
    Cursor       string    `bson:"cursor"`
    LastPostTime time.Time `bson:"last_post_time"`
}

type TopicUsage struct {
    CurrentPeriodStart time.Time `bson:"current_period_start"`
    MentionsCount      int       `bson:"mentions_count"`
    MentionsLimit      int       `bson:"mentions_limit"`
}
```

---

## 4. ClickHouse Schema — `listening_mentions`

Single unified table for all 6 platforms. Cross-platform queries are the primary use case.

### Table DDL

```sql
CREATE TABLE IF NOT EXISTS listening_mentions
(
    -- Identity
    mention_id        String,                    -- "{platform}:{native_id}"
    platform          LowCardinality(String),    -- tiktok, instagram, facebook, twitter, reddit, threads
    native_id         String,                    -- Platform's original post ID

    -- Topic Linkage
    workspace_id      String,                    -- MongoDB workspace ObjectId
    topic_id          String,                    -- MongoDB listening_topic ObjectId
    matched_keywords  Array(String),             -- Which keywords matched

    -- Content
    content_type      LowCardinality(String),    -- post, comment, reply, video, reel, thread
    text              String,                    -- Post text (truncated to 10KB)
    title             String,                    -- Reddit post title (empty for others)
    url               String,                    -- Permalink to original content
    language          LowCardinality(String),    -- ISO 639-1

    -- Author
    author_id         String,                    -- Platform user ID
    author_username   String,                    -- @handle
    author_name       String,                    -- Full display name
    author_avatar_url String,
    author_url        String,                    -- Profile URL
    author_followers  UInt64 DEFAULT 0,          -- Follower count (influencer scoring / reach proxy)
    author_verified   Bool DEFAULT false,

    -- Engagement Metrics (unified)
    likes_count       UInt64 DEFAULT 0,
    comments_count    UInt64 DEFAULT 0,
    shares_count      UInt64 DEFAULT 0,          -- Shares / retweets / reposts
    views_count       UInt64 DEFAULT 0,          -- Video views
    reactions_count   UInt64 DEFAULT 0,          -- Facebook total reactions
    quotes_count      UInt64 DEFAULT 0,          -- Twitter quotes / Threads quotes
    saves_count       UInt64 DEFAULT 0,          -- Bookmarks / saves
    score             Int64 DEFAULT 0,           -- Reddit score
    total_engagement  UInt64 DEFAULT 0,          -- Computed sum of all engagement

    -- Facebook Reaction Breakdown
    reaction_like     UInt32 DEFAULT 0,
    reaction_love     UInt32 DEFAULT 0,
    reaction_haha     UInt32 DEFAULT 0,
    reaction_wow      UInt32 DEFAULT 0,
    reaction_sad      UInt32 DEFAULT 0,
    reaction_angry    UInt32 DEFAULT 0,

    -- Media
    media_type        LowCardinality(String),    -- text, image, video, carousel, link
    media_urls        Array(String),             -- Attached image/video URLs
    thumbnail_url     String,

    -- Sentiment & AI (populated by sentiment service)
    sentiment_score   Float32 DEFAULT 0,         -- -1.0 to 1.0
    sentiment_label   LowCardinality(String),    -- positive, negative, neutral, mixed
    emotion           LowCardinality(String),    -- joy, anger, sadness, fear, surprise, disgust, neutral
    intent            LowCardinality(String),    -- complaint, praise, question, purchase, suggestion, general
    smart_tags        Array(String),             -- AI-generated topic labels
    is_spam           Bool DEFAULT false,

    -- Source Metadata
    source_name       String,                    -- Subreddit name, page name
    source_url        String,                    -- Subreddit URL, page URL

    -- User Actions (updated from Laravel API)
    is_bookmarked     Bool DEFAULT false,
    is_archived       Bool DEFAULT false,
    user_tags         Array(String),             -- Manual tags

    -- Timestamps
    published_at      DateTime64(3, 'UTC'),      -- Original publish time
    fetched_at        DateTime64(3, 'UTC'),      -- When we fetched it
    updated_at        DateTime64(3, 'UTC'),      -- ReplacingMergeTree version

    -- Dedup
    content_hash      String                     -- SHA256 for cross-topic dedup
)
ENGINE = ReplacingMergeTree(updated_at)
PARTITION BY toYYYYMM(published_at)
PRIMARY KEY (workspace_id, topic_id, mention_id)
ORDER BY (workspace_id, topic_id, mention_id)
SETTINGS index_granularity = 8192;
```

### Secondary Indexes

```sql
ALTER TABLE listening_mentions ADD INDEX idx_platform (platform) TYPE set(10) GRANULARITY 4;
ALTER TABLE listening_mentions ADD INDEX idx_sentiment (sentiment_label) TYPE set(5) GRANULARITY 4;
ALTER TABLE listening_mentions ADD INDEX idx_text_search (text) TYPE tokenbf_v1(30720, 2, 0) GRANULARITY 4;
ALTER TABLE listening_mentions ADD INDEX idx_published (published_at) TYPE minmax GRANULARITY 4;
ALTER TABLE listening_mentions ADD INDEX idx_keywords (matched_keywords) TYPE bloom_filter(0.01) GRANULARITY 4;
```

### Materialized View — Daily Aggregations

```sql
CREATE TABLE IF NOT EXISTS listening_daily_stats
(
    workspace_id      String,
    topic_id          String,
    platform          LowCardinality(String),
    day               Date,
    mention_count     AggregateFunction(count, UInt64),
    positive_count    AggregateFunction(sum, UInt64),
    negative_count    AggregateFunction(sum, UInt64),
    neutral_count     AggregateFunction(sum, UInt64),
    total_engagement  AggregateFunction(sum, UInt64),
    total_reach       AggregateFunction(sum, UInt64),
    avg_sentiment     AggregateFunction(avg, Float32)
)
ENGINE = AggregatingMergeTree()
PARTITION BY toYYYYMM(day)
ORDER BY (workspace_id, topic_id, platform, day);

CREATE MATERIALIZED VIEW IF NOT EXISTS mv_listening_daily_stats
TO listening_daily_stats
AS
SELECT
    workspace_id,
    topic_id,
    platform,
    toDate(published_at) AS day,
    countState() AS mention_count,
    sumState(if(sentiment_label = 'positive', 1, 0)::UInt64) AS positive_count,
    sumState(if(sentiment_label = 'negative', 1, 0)::UInt64) AS negative_count,
    sumState(if(sentiment_label = 'neutral', 1, 0)::UInt64) AS neutral_count,
    sumState(total_engagement) AS total_engagement,
    sumState(author_followers) AS total_reach,
    avgState(sentiment_score) AS avg_sentiment
FROM listening_mentions
GROUP BY workspace_id, topic_id, platform, day;
```

### Data365 → ClickHouse Field Mapping

| Data365 Field | Platform | ClickHouse Column |
|---|---|---|
| `id` | All | `native_id` |
| `text` | FB, IG, Threads | `text` |
| `body` | Twitter, Reddit | `text` |
| `title` | Reddit | `title` |
| `created_time` | FB, IG, Threads | `published_at` |
| `published` | Twitter | `published_at` |
| `created_utc` | Reddit | `published_at` |
| `owner_id` / `user_id` / `author` | All | `author_id` |
| `owner_username` / `author.handle` | All | `author_username` |
| `owner_full_name` | FB, Threads | `author_name` |
| `likes_count` / `favorites` | All | `likes_count` |
| `comments_count` / `reply_count` / `comment_count` | All | `comments_count` |
| `shares_count` | FB | `shares_count` |
| `retweet_count` | Twitter | `shares_count` |
| `reposts_count` | Threads | `shares_count` |
| `reactions_total_count` | FB | `reactions_count` |
| `reactions_like_count` ... `reactions_angry_count` | FB | `reaction_like` ... `reaction_angry` |
| `quote_count` / `quotes_count` | Twitter, Threads | `quotes_count` |
| `score` | Reddit | `score` |
| `post_type` / `product_type` / `tweet_type` | All | `content_type` (normalized) |
| `attached_image_url` / `attached_media_display_url` | All | `media_urls` (array) |
| `attached_carousel_media_urls` | IG | `media_urls` (array) |
| `link` / `post_url` / `permalink` | All | `url` |
| `language` | Twitter | `language` |
| `subreddit` | Reddit | `source_name` |
| `is_video` | IG | `media_type` = "video" |
| `owner_profile_pic_url` | Threads | `author_avatar_url` |

### Go Model

```go
// File: src/models/db/clickhouse/listening_mention.go

type ListeningMention struct {
    MentionID       string    `ch:"mention_id"`
    Platform        string    `ch:"platform"`
    NativeID        string    `ch:"native_id"`
    WorkspaceID     string    `ch:"workspace_id"`
    TopicID         string    `ch:"topic_id"`
    MatchedKeywords []string  `ch:"matched_keywords"`
    ContentType     string    `ch:"content_type"`
    Text            string    `ch:"text"`
    Title           string    `ch:"title"`
    URL             string    `ch:"url"`
    Language        string    `ch:"language"`
    AuthorID        string    `ch:"author_id"`
    AuthorUsername  string    `ch:"author_username"`
    AuthorName      string    `ch:"author_name"`
    AuthorAvatarURL string    `ch:"author_avatar_url"`
    AuthorURL       string    `ch:"author_url"`
    AuthorFollowers uint64    `ch:"author_followers"`
    AuthorVerified  bool      `ch:"author_verified"`
    LikesCount      uint64    `ch:"likes_count"`
    CommentsCount   uint64    `ch:"comments_count"`
    SharesCount     uint64    `ch:"shares_count"`
    ViewsCount      uint64    `ch:"views_count"`
    ReactionsCount  uint64    `ch:"reactions_count"`
    QuotesCount     uint64    `ch:"quotes_count"`
    SavesCount      uint64    `ch:"saves_count"`
    Score           int64     `ch:"score"`
    TotalEngagement uint64    `ch:"total_engagement"`
    ReactionLike    uint32    `ch:"reaction_like"`
    ReactionLove    uint32    `ch:"reaction_love"`
    ReactionHaha    uint32    `ch:"reaction_haha"`
    ReactionWow     uint32    `ch:"reaction_wow"`
    ReactionSad     uint32    `ch:"reaction_sad"`
    ReactionAngry   uint32    `ch:"reaction_angry"`
    MediaType       string    `ch:"media_type"`
    MediaURLs       []string  `ch:"media_urls"`
    ThumbnailURL    string    `ch:"thumbnail_url"`
    SentimentScore  float32   `ch:"sentiment_score"`
    SentimentLabel  string    `ch:"sentiment_label"`
    Emotion         string    `ch:"emotion"`
    Intent          string    `ch:"intent"`
    SmartTags       []string  `ch:"smart_tags"`
    IsSpam          bool      `ch:"is_spam"`
    SourceName      string    `ch:"source_name"`
    SourceURL       string    `ch:"source_url"`
    IsBookmarked    bool      `ch:"is_bookmarked"`
    IsArchived      bool      `ch:"is_archived"`
    UserTags        []string  `ch:"user_tags"`
    PublishedAt     time.Time `ch:"published_at"`
    FetchedAt       time.Time `ch:"fetched_at"`
    UpdatedAt       time.Time `ch:"updated_at"`
    ContentHash     string    `ch:"content_hash"`
}
```

---

## 5. Kafka Topics & Message Schemas

### Topics

Following existing convention (`{stage}-{domain}-{data_type}`):

```
# Work orders (Scheduler/API → Fetcher)
listening-work-order                  # Scheduled + immediate work orders (partitions: 6)

# Raw data (Fetcher → Parser)
listening-raw-mentions                # Raw Data365 responses, all platforms (partitions: 6)

# Parsed data (Parser → Sentiment)
listening-parsed-mentions             # Normalized mentions (partitions: 6)

# Enriched data (Sentiment → Sink)
listening-enriched-mentions           # Mentions with sentiment + tags (partitions: 6)
```

### Message Schemas

```go
// File: src/models/kafka/listening.go

// ListeningWorkOrder — produced by scheduler/API, consumed by fetcher
type ListeningWorkOrder struct {
    TopicID      string              `json:"topic_id"`
    WorkspaceID  string              `json:"workspace_id"`
    TopicName    string              `json:"topic_name"`
    Keywords     []string            `json:"keywords"`
    ExcludeWords []string            `json:"exclude_words"`
    Platforms    map[string]bool     `json:"platforms"`
    Language     []string            `json:"language"`
    Regions      []string            `json:"regions"`
    Cursors      map[string]CursorState `json:"cursors"`
    SyncType     string              `json:"sync_type"`    // "daily" | "immediate" | "initial"
    LookbackDays int                 `json:"lookback_days"` // 1 for daily, 30 for initial
    MentionsLeft int                 `json:"mentions_left"` // remaining quota
    CreatedAt    time.Time           `json:"created_at"`
}

type CursorState struct {
    Cursor       string    `json:"cursor"`
    LastPostTime time.Time `json:"last_post_time"`
}

// RawListeningMention — produced by fetcher, consumed by parser
type RawListeningMention struct {
    TopicID     string          `json:"topic_id"`
    WorkspaceID string          `json:"workspace_id"`
    Platform    string          `json:"platform"`
    Keywords    []string        `json:"keywords"`
    ExcludeWords []string       `json:"exclude_words"`
    RawData     json.RawMessage `json:"raw_data"`     // Data365 response as-is
    FetchedAt   time.Time       `json:"fetched_at"`
}

// ParsedListeningMention — produced by parser, consumed by sentiment service
type ParsedListeningMention struct {
    ListeningMention                    // Embedded ClickHouse model (all fields)
    // Sentiment fields will be zero-valued, filled by sentiment service
}

// EnrichedListeningMention — produced by sentiment, consumed by sink
type EnrichedListeningMention struct {
    ListeningMention                    // Fully populated including sentiment
}
```

---

## 6. Pipeline Architecture

### Flow

```
┌─────────────────────────────────────────────────────────────┐
│                      ENTRY POINTS                           │
│                                                             │
│  1. Daily Cron: ./jobs -mode listening -syncType daily      │
│  2. API: POST /api/v1/listening-work (on-demand/initial)    │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  SCHEDULER                                                  │
│                                                             │
│  • Query MongoDB: listening_topics                          │
│    WHERE status = "active"                                  │
│    AND mentions_limit_reached = false                       │
│    AND last_fetched_at < (now - 24h)  [daily mode]          │
│  • Batch: 50 topics per query, paginate via lastID          │
│  • Produce ListeningWorkOrder → Kafka                       │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  Kafka: listening-work-order                                │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  DATA365 FETCHER SERVICE                                    │
│                                                             │
│  For each work order:                                       │
│    For each keyword in topic:                               │
│      For each enabled platform:                             │
│        1. POST search request to Data365                    │
│        2. Poll status until "finished" (backoff: 5s→30s)    │
│        3. GET paginated results (100/page)                  │
│        4. Produce each page → Kafka raw-mentions            │
│                                                             │
│  Rate limiting: per-token + global (reuse RateManager)      │
│  Retry: 5 attempts, exponential backoff 300ms→8s            │
│  Error handling: log + skip on expected errors (404, 403)   │
│                                                             │
│  Concurrency:                                               │
│    • 6 platform goroutines per topic (parallel)             │
│    • 15 concurrent workers consuming from Kafka             │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  Kafka: listening-raw-mentions                              │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  PARSER SERVICE                                             │
│                                                             │
│  For each raw mention:                                      │
│    1. Detect platform from message                          │
│    2. Deserialize platform-specific JSON                    │
│    3. Map to unified ListeningMention struct                │
│    4. Generate mention_id = "{platform}:{native_id}"        │
│    5. Generate content_hash (SHA256 dedup key)              │
│    6. Check Redis dedup set — skip if exists                │
│    7. Apply exclude_keywords filter — skip if matched       │
│    8. Detect matched_keywords from text                     │
│    9. Compute total_engagement                              │
│   10. Normalize content_type and media_type                 │
│   11. Truncate text to 10KB                                 │
│   12. Produce → Kafka parsed-mentions                       │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  Kafka: listening-parsed-mentions                           │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  SENTIMENT SERVICE  (see Section 8 for details)             │
│                                                             │
│  • Route to AI Agents service (existing infra)              │
│  • Batch 50 mentions per request                            │
│  • Returns: sentiment, emotion, intent, smart_tags, is_spam │
│  • Produce → Kafka enriched-mentions                        │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  Kafka: listening-enriched-mentions                         │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  SINK SERVICE                                               │
│                                                             │
│  1. Batch collect 1000 mentions                             │
│  2. BulkInsert to ClickHouse (listening_mentions)           │
│  3. Group by topic_id, atomically increment                 │
│     usage.mentions_count in MongoDB                         │
│  4. If mentions_count >= mentions_limit:                    │
│       → Set mentions_limit_reached = true                   │
│       → Send limit-reached notification                     │
│  5. Update last_fetched_at and cursors on topic             │
│  6. Check alert triggers (see Section 9)                    │
└─────────────────────────────────────────────────────────────┘
```

### Data365 Fetcher — Polling Strategy

```go
// Polling for async Data365 search task completion
func (f *Data365Fetcher) pollTaskStatus(ctx context.Context, taskID string) error {
    backoff := []time.Duration{5*time.Second, 10*time.Second, 15*time.Second, 30*time.Second, 30*time.Second}
    for attempt, wait := range backoff {
        status, err := f.client.GetTaskStatus(ctx, taskID)
        if err != nil {
            return fmt.Errorf("poll attempt %d: %w", attempt, err)
        }
        if status == "finished" {
            return nil
        }
        if status == "error" {
            return fmt.Errorf("task %s failed", taskID)
        }
        time.Sleep(wait)
    }
    return fmt.Errorf("task %s timed out after polling", taskID)
}
```

---

## 7. Deduplication Strategy

### Level 1: ClickHouse ReplacingMergeTree

Same pattern as existing analytics pipeline:

```sql
ENGINE = ReplacingMergeTree(updated_at)
ORDER BY (workspace_id, topic_id, mention_id)
```

- `mention_id = "{platform}:{native_id}"` — unique per platform post
- `updated_at` version column — keeps latest engagement metrics
- Query with `FINAL` for deduped reads
- `insert_deduplicate = 0` in client config (dedup at query time, not insert time)

### Level 2: Redis Content Hash (parser stage)

Prevents processing the same mention across multiple search keywords within the same topic:

```go
func (p *Parser) isDuplicate(ctx context.Context, workspaceID, topicID, contentHash string) bool {
    key := fmt.Sprintf("listening:dedup:%s:%s", workspaceID, topicID)
    added, _ := p.redis.SAdd(ctx, key, contentHash).Result()
    if added == 0 {
        return true // duplicate
    }
    p.redis.Expire(ctx, key, 48*time.Hour)
    return false
}

func generateContentHash(platform, authorID, text string, publishedAt time.Time) string {
    normalized := strings.ToLower(strings.TrimSpace(text))
    if len(normalized) > 200 {
        normalized = normalized[:200]
    }
    input := fmt.Sprintf("%s:%s:%s:%s", platform, authorID, normalized, publishedAt.Format("2006-01-02"))
    hash := sha256.Sum256([]byte(input))
    return hex.EncodeToString(hash[:])
}
```

### Why Two Levels

| Scenario | Handled By |
|---|---|
| Same post fetched on consecutive days (updated engagement) | ClickHouse ReplacingMergeTree — keeps latest via `updated_at` |
| Same post matched by keyword "acme" and keyword "#acme" in same run | Redis content hash — skips second occurrence |
| Same post across different topics in same workspace | Allowed — different `topic_id` in ORDER BY = separate rows |

---

## 8. Sentiment & AI Enrichment

### Integration with Existing AI Agents Service

ContentStudio already has a centralized AI agents microservice:
- **URL:** Configured via `AI_AGENTS_BASE_URL` (e.g., `https://ai-agents.contentstudio.io/api/v1/`)
- **Auth:** `AI_AGENTS_API_KEY`
- **Framework:** Python Agno agents with Claude Sonnet 4.5 (primary) + GPT-4o (fallback)
- **Async jobs:** Dramatiq + Redis queues

### Sentiment Pipeline

The Go listening service calls the AI Agents service via HTTP, same pattern as `AiAgentService` in the Laravel backend.

```
┌──────────────────────┐     HTTP POST      ┌──────────────────────┐
│  Sentiment Service   │ ─────────────────→ │  AI Agents Service   │
│  (Go, Kafka consumer)│                    │  (Python, Agno)      │
│                      │ ←───────────────── │                      │
│  Batches 50 mentions │     JSON response  │  Claude/GPT model    │
└──────────────────────┘                    └──────────────────────┘
```

### New AI Agent: ListeningSentimentAgent

A new agent to be added to the Python AI agents service:

**Endpoint:** `POST /api/v1/listening/sentiment`

**Request:**
```json
{
  "mentions": [
    { "id": "twitter:123456", "text": "Acme Corp just launched...", "platform": "twitter" },
    { "id": "facebook:789", "text": "Terrible experience with...", "platform": "facebook" }
  ]
}
```

**Response:**
```json
{
  "results": [
    {
      "id": "twitter:123456",
      "sentiment_label": "positive",
      "sentiment_score": 0.82,
      "emotion": "joy",
      "intent": "praise",
      "smart_tags": ["product launch", "brand awareness"],
      "is_spam": false
    },
    {
      "id": "facebook:789",
      "sentiment_label": "negative",
      "sentiment_score": -0.75,
      "emotion": "anger",
      "intent": "complaint",
      "smart_tags": ["customer service", "bad experience"],
      "is_spam": false
    }
  ]
}
```

### Go Sentiment Service Flow

```go
// File: src/services/listening/sentiment/sentiment_service.go

// 1. Consume from listening-parsed-mentions
// 2. Buffer 50 mentions
// 3. POST batch to AI Agents service
// 4. Map results back to mentions
// 5. Produce to listening-enriched-mentions

func (s *SentimentService) processBatch(mentions []ParsedListeningMention) error {
    // Build request payload
    payload := buildSentimentRequest(mentions)

    // Call AI Agents service (same pattern as AiAgentService in Laravel)
    resp, err := s.aiClient.Post(ctx, "/listening/sentiment", payload)
    if err != nil {
        // Fallback: insert with empty sentiment rather than blocking pipeline
        for i := range mentions {
            mentions[i].SentimentLabel = "unknown"
        }
        return s.produceEnriched(mentions)
    }

    // Map results back
    resultMap := mapByID(resp.Results)
    for i := range mentions {
        if r, ok := resultMap[mentions[i].MentionID]; ok {
            mentions[i].SentimentScore = r.SentimentScore
            mentions[i].SentimentLabel = r.SentimentLabel
            mentions[i].Emotion = r.Emotion
            mentions[i].Intent = r.Intent
            mentions[i].SmartTags = r.SmartTags
            mentions[i].IsSpam = r.IsSpam
        }
    }

    return s.produceEnriched(mentions)
}
```

### Fallback Behavior

If the AI Agents service is down or slow:
- Mentions are inserted with `sentiment_label = "unknown"`
- A background retry job re-processes `unknown` sentiment mentions (query ClickHouse, re-call AI, update via insert with newer `updated_at`)

---

## 9. Alerts

### Alert Types

| Alert Type | Trigger Condition | Detection Point |
|---|---|---|
| **Volume Spike** | Mention count in last 1h exceeds 2× the 7-day hourly average | Sink service, after batch insert |
| **Sentiment Shift** | Average sentiment drops below -0.3 (from neutral/positive baseline) | Sink service, after batch insert |
| **Mentions Limit** | `usage.mentions_count >= usage.mentions_limit` | Sink service, after usage increment |

### Alert Detection (Sink Service)

```go
// After inserting a batch for a topic, check alert conditions
func (s *SinkService) checkAlerts(ctx context.Context, topicID, workspaceID string) {
    // Volume spike: query ClickHouse for last 1h count vs 7-day average
    // Sentiment shift: query listening_daily_stats for recent trend
    // If triggered, produce alert notification
}
```

### Alert Delivery

**Notification payload:**
```json
{
  "type": "listening_volume_spike" | "listening_sentiment_shift" | "listening_limit_reached",
  "workspace_id": "...",
  "topic_id": "...",
  "topic_name": "Brand Monitoring - Acme Corp",
  "data": {
    "current_value": 150,
    "baseline_value": 45,
    "platform": "twitter",
    "timestamp": "2026-03-11T10:00:00Z"
  }
}
```

### Open Questions — Alert Delivery Channel

| Option | Pros | Cons |
|---|---|---|
| **Pusher** (existing, 3 instances) | Already integrated, real-time, proven | Separate Pusher instance needed? Or reuse main? |
| **Redis pub/sub → Socket.io** (existing socket service) | Already handles 29+ channels, proven | Adds a new channel type to socket service |
| **Email** (Postmark, existing) | Users get alerts when offline | Delay, not real-time |
| **Webhook** (user-configured URL) | Power users, integrations | New infrastructure to build |

**Recommended:** Use the existing notification flow:
1. Go service pushes alert to **Redis channel** (e.g., `listening_alert`)
2. **Socket notification service** picks it up, routes to workspace/user
3. **Email** queued via Laravel `NotificationEmailsJob` (existing pattern)
4. **Webhook** delivery — TBD, may not be needed for v1

**Decision needed from team:**
- Do we create a new Pusher instance for listening (like `AnalyticsPusher`, `InboxPusher`)?
- Or reuse main Pusher + add `listening_alert` to socket service's Redis channel subscriptions?
- Should webhooks be in v1 scope?

---

## 10. Usage Limits & Billing

### Limits

| Limit | Default | Source of truth | Enforcement |
|---|---|---|---|
| Topic slots per owner/admin account | 5 | Laravel billing/allocation | Laravel validates entitlement before topic create/assignment |
| Mentions quota | Plan-defined separate quota | Laravel billing | Go enforces effective `mentions_limit_reached` flags and counters materialized into MongoDB |

### Flag-Based Mention Cap

```
1. Sink inserts batch to ClickHouse

2. Sink updates per-topic operational counters in MongoDB:
   db.listening_topics.findOneAndUpdate(
     { _id: topicId },
     { $inc: { "usage.mentions_count": batchSize } },
     { returnDocument: "after" }
   )

3. Laravel/billing sync materializes the effective quota state for the topic owner's plan:
   db.listening_topics.updateOne(
     { _id: topicId },
     { $set: { "mentions_limit_reached": true | false } }
   )

4. Scheduler skips flagged topics:
   filter: { status: "active", mentions_limit_reached: false }

5. On-demand API also checks the flag before queueing
```

### Monthly Reset (Billing-Cycle Aligned)

```go
// Laravel remains the billing source of truth.
// It synchronizes the owner's current billing window and effective
// mentions_limit_reached flag onto listening topic documents.
// Go consumes that state when scheduling and queueing work.
```

`current_period_start` is synchronized from the billing owner/account that owns the topic slot, not from the workspace alone.

### Upgrade Path

When an owner/admin upgrades their plan:
1. Laravel backend updates the effective quota metadata for all active topics owned by that account
2. Clears `mentions_limit_reached = false`
3. Topics resume in the next scheduler run or on-demand trigger

---

## 11. API Endpoints

### Data Flow Architecture

Social Listening is the first module to adopt the **Go-direct API pattern** for ClickHouse queries. Since the team is migrating analytics queries from Laravel to Go, building this new module directly in Go avoids building in Laravel first only to migrate later.

```
┌────────────────────────────────────────────────────────────────────────┐
│                         FRONTEND                                       │
│                                                                        │
│  Topic CRUD (MongoDB)          Mentions & Analytics (ClickHouse)       │
│  ─────────────────────         ──────────────────────────────────      │
│  POST/GET/PUT/DELETE           GET mentions, analytics, usage          │
│         │                                │                             │
│         ▼                                ▼                             │
│    Laravel Backend                  Go Service                         │
│    (api.contentstudio.io)           (analytics-go.contentstudio.io)    │
│         │                                │                             │
│         ▼                                ▼                             │
│      MongoDB                         ClickHouse                        │
│    (listening_topics)              (listening_mentions)                │
│                                    (listening_daily_stats)             │
└────────────────────────────────────────────────────────────────────────┘
```

**Split rationale:**
- **Laravel handles topic CRUD** — Topic management is a standard REST resource backed by MongoDB. Laravel already handles all MongoDB operations, JWT auth, Paddle billing, and validation. No ClickHouse involved.
- **Go handles mentions & analytics queries** — These are high-performance ClickHouse reads. Go is already the ClickHouse client in the pipeline. Building query endpoints here keeps read/write in the same service, avoids duplicating ClickHouse connection setup in Laravel, and sets the pattern for migrating existing analytics.
- **Go handles pipeline triggers** — `POST /api/v1/listening-work` stays in Go (Kafka producer).

### Go Service Endpoints — Pipeline Triggers

Added to `src/cmd/api-server/`:

#### POST /api/v1/listening-work

Triggers an on-demand or initial pipeline run for a topic.

```
POST /api/v1/listening-work    (JWT protected)

Request:
{
    "topic_id": "6601abc...",
    "workspace_id": "5f0c...",
    "sync_type": "immediate" | "initial"
}

Response (200):
{
    "status": "ok",
    "message": "Listening work order queued",
    "timestamp": "2026-03-11T10:00:00Z"
}

Response (429):
{
    "status": "error",
    "error": "mentions_limit_reached",
    "message": "Monthly mentions limit reached for this topic",
    "timestamp": "2026-03-11T10:00:00Z"
}

Response (404):
{
    "status": "error",
    "error": "topic_not_found",
    "message": "Topic not found or not active",
    "timestamp": "2026-03-11T10:00:00Z"
}
```

**Logic:**
1. Fetch topic from MongoDB by `topic_id`
2. Validate `workspace_id` matches
3. Check `status == "active"`
4. Check `mentions_limit_reached == false` — return 429 if true
5. Build `ListeningWorkOrder` with appropriate `lookback_days`:
   - `immediate`: 1 day
   - `initial`: 30 days
6. Produce to `listening-work-order` Kafka topic
7. For `initial`: set `is_initial_sync_done = true` after work order is queued

### Go Service Endpoints — ClickHouse Query APIs

These endpoints serve mention and analytics data directly from ClickHouse to the frontend. All are JWT-protected and require `workspace_id` validation.

#### GET /api/v1/listening/mentions

Returns paginated mentions for a topic with filters.

```
GET /api/v1/listening/mentions    (JWT protected)

Query Parameters:
  workspace_id   string   required   Workspace ID
  topic_id       string   required   Topic ID
  page           int      optional   Page number (default: 1)
  per_page       int      optional   Items per page (default: 20, max: 100)
  platform       string   optional   Filter: tiktok, instagram, facebook, twitter, reddit, threads
  sentiment      string   optional   Filter: positive, negative, neutral, mixed
  content_type   string   optional   Filter: post, reply, quote, video, reel, thread
  date_from      string   optional   ISO 8601 date (e.g., 2026-03-01)
  date_to        string   optional   ISO 8601 date
  search         string   optional   Full-text search in mention text
  sort_by        string   optional   published_at (default), total_engagement, sentiment_score
  sort_order     string   optional   desc (default), asc
  is_bookmarked  bool     optional   Filter bookmarked only
  smart_tags     string   optional   Comma-separated tag filter
  author_verified bool    optional   Filter verified authors only

Response (200):
{
    "status": "ok",
    "data": {
        "mentions": [
            {
                "mention_id": "twitter:123456",
                "platform": "twitter",
                "native_id": "123456",
                "content_type": "post",
                "text": "Just tried Acme Corp's new product...",
                "title": "",
                "url": "https://twitter.com/user/status/123456",
                "language": "en",
                "matched_keywords": ["acme corp"],
                "author": {
                    "id": "user123",
                    "username": "johndoe",
                    "name": "John Doe",
                    "avatar_url": "https://...",
                    "url": "https://twitter.com/johndoe",
                    "followers": 5200,
                    "verified": false
                },
                "engagement": {
                    "likes": 45,
                    "comments": 12,
                    "shares": 8,
                    "views": 0,
                    "reactions": 0,
                    "quotes": 3,
                    "saves": 0,
                    "score": 0,
                    "total": 68
                },
                "reactions": {
                    "like": 0, "love": 0, "haha": 0,
                    "wow": 0, "sad": 0, "angry": 0
                },
                "media": {
                    "type": "image",
                    "urls": ["https://..."],
                    "thumbnail_url": ""
                },
                "sentiment": {
                    "score": 0.82,
                    "label": "positive",
                    "emotion": "joy",
                    "intent": "praise"
                },
                "smart_tags": ["product launch", "brand awareness"],
                "is_spam": false,
                "source": {
                    "name": "",
                    "url": ""
                },
                "is_bookmarked": false,
                "is_archived": false,
                "user_tags": [],
                "published_at": "2026-03-10T14:30:00Z",
                "fetched_at": "2026-03-11T02:00:00Z"
            }
        ],
        "pagination": {
            "page": 1,
            "per_page": 20,
            "total": 1523,
            "total_pages": 77
        }
    },
    "timestamp": "2026-03-11T10:00:00Z"
}
```

**ClickHouse query pattern:**
```sql
SELECT *
FROM listening_mentions FINAL
WHERE workspace_id = {workspace_id:String}
  AND topic_id = {topic_id:String}
  AND published_at >= {date_from:DateTime64}
  AND published_at <= {date_to:DateTime64}
  -- Optional filters applied dynamically
  AND platform = {platform:String}
  AND sentiment_label = {sentiment:String}
ORDER BY published_at DESC
LIMIT {per_page:UInt32} OFFSET {offset:UInt32}
```

**Count query (for pagination total):**
```sql
SELECT count() as total
FROM listening_mentions FINAL
WHERE workspace_id = {workspace_id:String}
  AND topic_id = {topic_id:String}
  -- Same filters as above
```

#### GET /api/v1/listening/analytics/summary

Returns KPI cards and high-level stats for a topic.

```
GET /api/v1/listening/analytics/summary    (JWT protected)

Query Parameters:
  workspace_id   string   required
  topic_id       string   required
  date_from      string   required   ISO 8601
  date_to        string   required   ISO 8601

Response (200):
{
    "status": "ok",
    "data": {
        "total_mentions": 1523,
        "total_engagement": 89450,
        "total_reach": 2340000,
        "avg_sentiment": 0.34,
        "sentiment_breakdown": {
            "positive": 612,
            "negative": 234,
            "neutral": 651,
            "mixed": 26
        },
        "platform_breakdown": {
            "twitter": 542,
            "facebook": 312,
            "reddit": 245,
            "tiktok": 198,
            "instagram": 156,
            "threads": 70
        },
        "top_emotions": {
            "joy": 389,
            "neutral": 512,
            "anger": 134,
            "surprise": 98,
            "sadness": 67
        },
        "comparison": {
            "mentions_change_pct": 12.5,
            "sentiment_change": 0.08,
            "engagement_change_pct": -3.2
        }
    },
    "timestamp": "2026-03-11T10:00:00Z"
}
```

**ClickHouse queries (from `listening_daily_stats` materialized view):**
```sql
-- Current period
SELECT
    sum(countMerge(mention_count)) as total_mentions,
    sum(sumMerge(total_engagement)) as total_engagement,
    sum(sumMerge(total_reach)) as total_reach,
    avg(avgMerge(avg_sentiment)) as avg_sentiment,
    sum(sumMerge(positive_count)) as positive,
    sum(sumMerge(negative_count)) as negative,
    sum(sumMerge(neutral_count)) as neutral
FROM listening_daily_stats
WHERE workspace_id = {workspace_id:String}
  AND topic_id = {topic_id:String}
  AND day >= {date_from:Date}
  AND day <= {date_to:Date}

-- Platform breakdown
SELECT platform, sum(countMerge(mention_count)) as count
FROM listening_daily_stats
WHERE ...
GROUP BY platform

-- Comparison: previous period of same length
-- (date_from - period_length) to date_from
```

#### GET /api/v1/listening/analytics/timeline

Returns daily mention volume and sentiment trend for charts.

```
GET /api/v1/listening/analytics/timeline    (JWT protected)

Query Parameters:
  workspace_id   string   required
  topic_id       string   required
  date_from      string   required
  date_to        string   required
  group_by       string   optional   day (default), week, month
  platform       string   optional   Filter by platform

Response (200):
{
    "status": "ok",
    "data": {
        "timeline": [
            {
                "date": "2026-03-01",
                "mentions": 52,
                "positive": 22,
                "negative": 8,
                "neutral": 22,
                "engagement": 3200,
                "reach": 125000,
                "avg_sentiment": 0.41
            },
            {
                "date": "2026-03-02",
                "mentions": 67,
                "positive": 30,
                "negative": 15,
                "neutral": 22,
                "engagement": 4100,
                "reach": 189000,
                "avg_sentiment": 0.28
            }
        ]
    },
    "timestamp": "2026-03-11T10:00:00Z"
}
```

**ClickHouse query:**
```sql
SELECT
    day as date,
    countMerge(mention_count) as mentions,
    sumMerge(positive_count) as positive,
    sumMerge(negative_count) as negative,
    sumMerge(neutral_count) as neutral,
    sumMerge(total_engagement) as engagement,
    sumMerge(total_reach) as reach,
    avgMerge(avg_sentiment) as avg_sentiment
FROM listening_daily_stats
WHERE workspace_id = {workspace_id:String}
  AND topic_id = {topic_id:String}
  AND day >= {date_from:Date}
  AND day <= {date_to:Date}
GROUP BY day
ORDER BY day ASC
```

#### GET /api/v1/listening/analytics/top-authors

Returns top authors by mention count or engagement.

```
GET /api/v1/listening/analytics/top-authors    (JWT protected)

Query Parameters:
  workspace_id   string   required
  topic_id       string   required
  date_from      string   required
  date_to        string   required
  sort_by        string   optional   mentions (default), engagement, followers
  limit          int      optional   Default: 10

Response (200):
{
    "status": "ok",
    "data": {
        "authors": [
            {
                "author_id": "user123",
                "author_username": "techblogger",
                "author_name": "Tech Blogger",
                "author_avatar_url": "https://...",
                "author_url": "https://twitter.com/techblogger",
                "author_followers": 125000,
                "author_verified": true,
                "platform": "twitter",
                "mention_count": 8,
                "total_engagement": 12450,
                "avg_sentiment": 0.72
            }
        ]
    },
    "timestamp": "2026-03-11T10:00:00Z"
}
```

**ClickHouse query:**
```sql
SELECT
    author_id,
    any(author_username) as author_username,
    any(author_name) as author_name,
    any(author_avatar_url) as author_avatar_url,
    any(author_url) as author_url,
    max(author_followers) as author_followers,
    any(author_verified) as author_verified,
    any(platform) as platform,
    count() as mention_count,
    sum(total_engagement) as total_engagement,
    avg(sentiment_score) as avg_sentiment
FROM listening_mentions FINAL
WHERE workspace_id = {workspace_id:String}
  AND topic_id = {topic_id:String}
  AND published_at >= {date_from:DateTime64}
  AND published_at <= {date_to:DateTime64}
GROUP BY author_id
ORDER BY mention_count DESC
LIMIT {limit:UInt32}
```

#### GET /api/v1/listening/analytics/top-posts

Returns top mentions by engagement.

```
GET /api/v1/listening/analytics/top-posts    (JWT protected)

Query Parameters:
  workspace_id   string   required
  topic_id       string   required
  date_from      string   required
  date_to        string   required
  platform       string   optional
  limit          int      optional   Default: 10

Response (200):
{
    "status": "ok",
    "data": {
        "posts": [
            // Same structure as mentions in GET /mentions, sorted by total_engagement DESC
        ]
    },
    "timestamp": "2026-03-11T10:00:00Z"
}
```

#### POST /api/v1/listening/mentions/bookmark

Toggle bookmark on a mention. Updates ClickHouse via insert with newer `updated_at`.

```
POST /api/v1/listening/mentions/bookmark    (JWT protected)

Request:
{
    "workspace_id": "5f0c...",
    "topic_id": "6601abc...",
    "mention_id": "twitter:123456",
    "is_bookmarked": true
}

Response (200):
{
    "status": "ok",
    "message": "Mention bookmarked",
    "timestamp": "2026-03-11T10:00:00Z"
}
```

**Logic:** Read existing mention from ClickHouse, update `is_bookmarked`, re-insert with new `updated_at`. ReplacingMergeTree will deduplicate and keep the newer version.

#### POST /api/v1/listening/mentions/tag

Add/remove user tags on a mention.

```
POST /api/v1/listening/mentions/tag    (JWT protected)

Request:
{
    "workspace_id": "5f0c...",
    "topic_id": "6601abc...",
    "mention_id": "twitter:123456",
    "user_tags": ["important", "follow-up"]
}

Response (200):
{
    "status": "ok",
    "message": "Tags updated",
    "timestamp": "2026-03-11T10:00:00Z"
}
```

### Laravel Backend Endpoints (Topic CRUD — MongoDB)

These are handled by the Laravel backend. Not part of Go service scope, listed here for context so the implementing agent understands the full picture.

| Method | Endpoint | Purpose | DB |
|---|---|---|---|
| POST | `/api/listening/topics` | Create topic (enforce 5 limit, trigger initial sync via Go) | MongoDB |
| GET | `/api/listening/topics` | List topics for workspace | MongoDB |
| GET | `/api/listening/topics/{id}` | Get topic details + usage stats | MongoDB |
| PUT | `/api/listening/topics/{id}` | Update topic (keywords, platforms) | MongoDB |
| DELETE | `/api/listening/topics/{id}` | Soft delete (set status=deleted) | MongoDB |
| POST | `/api/listening/topics/{id}/pause` | Pause topic (set status=paused) | MongoDB |
| POST | `/api/listening/topics/{id}/resume` | Resume topic (set status=active) | MongoDB |
| GET | `/api/listening/usage` | Get workspace usage (topic count, mentions used) | MongoDB |

**On topic creation**, Laravel:
1. Validates workspace topic count < 5
2. Inserts into `listening_topics` MongoDB collection
3. Calls Go service `POST /api/v1/listening-work` with `sync_type: "initial"` to trigger first pipeline run

---

## 12. Directory Structure

New files within the existing Go service:

```
contentstudio-social-analytics-go/src/
│
├── cmd/
│   ├── api-server/
│   │   └── main.go                              # Register new routes
│   └── jobs/
│       └── main.go                              # Add listening scheduler branch following existing flag pattern
│
├── api/
│   ├── handle_listening_work.go                 # POST /api/v1/listening-work
│   ├── handle_listening_work_test.go
│   └── analytics/
│       └── listening/
│           ├── handler.go                       # GET /mentions, /summary, /timeline, /top-authors, /top-posts
│           └── handler_test.go
│
├── clients/
│   └── social/
│       └── data365.go                           # Data365 HTTP client
│                                                #   - POST search, GET status, GET results
│                                                #   - Rate limiting (RateManager)
│                                                #   - Retry with exponential backoff
│
├── services/
│   ├── analytics/
│   │   └── listening/
│   │       └── service.go                       # ClickHouse-backed read service
│   └── listening/
│       ├── listening-fetcher/                   # Data365 API fetcher (all 6 platforms)
│       ├── listening-parser/                    # Data365 JSON -> unified ListeningMention
│       ├── listening-sentiment/                 # Batch calls to AI Agents service
│       └── listening-clickhouse-sink/           # Batch insert + usage tracking + alerts
│
├── models/
│   ├── api/
│   │   └── listening.go                         # Request/response payloads for listening endpoints
│   ├── db/
│   │   ├── mongo/
│   │   │   ├── listening_topic.go               # ListeningTopic struct
│   │   │   └── listening_settings.go            # ListeningSettings struct
│   │   └── clickhouse/
│   │       └── listening.go                     # ListeningMention and related read models
│   └── kafka/
│       └── listening.go                         # WorkOrder, Raw, Parsed, Enriched messages
│
├── db/
│   ├── clickhouse/
│   │   ├── analytics-get-queries/
│   │   │   └── listening/
│   │   │       └── repository.go                # QueryMentions(), GetSummary(), GetTimeline(), GetTopAuthors(), GetTopPosts()
│   │   └── listening_write.go                   # BulkInsertMentions(), UpdateBookmark(), UpdateTags()
│   └── mongodb/
│       └── listening.go                         # CRUD for listening_topics/settings, usage increment, cursor updates
│
└── deployments/
    ├── clickhouse/schema/
    │   └── listening_schema.sql                 # DDL for listening_mentions + materialized views
    ├── docker-compose.local-infra.yml           # Add Redis service for dedup + locks
    └── .create-topics.sh                        # Add listening Kafka topics
```

### New Go Dependencies

```go
// No new dependencies required.
// Data365 client: standard net/http (already used)
// AI Agents client: standard net/http
// Redis dedup: github.com/redis/go-redis/v9 (already imported)
// Kafka: github.com/twmb/franz-go (already imported)
// ClickHouse: github.com/ClickHouse/clickhouse-go/v2 (already imported)
// MongoDB: go.mongodb.org/mongo-driver (already imported)
// Rate limiting: golang.org/x/time/rate (already imported)
```

---

## 13. Configuration

### New Environment Variables

```bash
# Data365 API
APP_DATA365_ACCESS_TOKEN=...                     # Data365 API token
APP_DATA365_BASE_URL=https://api.data365.co/v1.1 # API base URL
APP_DATA365_RPS=50                                # Requests per second
APP_DATA365_BURST=100                             # Burst limit
APP_DATA365_POLL_INTERVAL_SEC=5                   # Initial poll interval for async tasks
APP_DATA365_POLL_MAX_WAIT_SEC=300                 # Max wait for task completion (5 min)
APP_DATA365_MAX_DOCS_PER_SEARCH=300               # Default max_documents per search

# AI Agents (reuse existing config pattern from Laravel)
APP_AI_AGENTS_BASE_URL=https://ai-agents.contentstudio.io/api/v1
APP_AI_AGENTS_API_KEY=...
APP_AI_AGENTS_TIMEOUT_SEC=60
APP_AI_AGENTS_SENTIMENT_BATCH_SIZE=50

# Listening Pipeline
APP_LISTENING_ENABLED=true
APP_LISTENING_DEFAULT_MENTIONS_LIMIT=10000
APP_LISTENING_DEFAULT_TOPICS_LIMIT=5
APP_LISTENING_DAILY_LOOKBACK_DAYS=1
APP_LISTENING_INITIAL_LOOKBACK_DAYS=30
APP_LISTENING_DEDUP_TTL_HOURS=48
APP_LISTENING_SINK_BATCH_SIZE=1000
APP_LISTENING_FETCHER_WORKERS=15
APP_LISTENING_PARSER_WORKERS=10
APP_LISTENING_SENTIMENT_WORKERS=5
APP_LISTENING_SINK_WORKERS=10
```

### Viper Config Bindings

```go
// File: src/config/config.go (extend existing)

type ListeningConfig struct {
    Enabled              bool    `mapstructure:"listening_enabled"`
    DefaultMentionsLimit int     `mapstructure:"listening_default_mentions_limit"`
    DefaultTopicsLimit   int     `mapstructure:"listening_default_topics_limit"`
    DailyLookbackDays    int     `mapstructure:"listening_daily_lookback_days"`
    InitialLookbackDays  int     `mapstructure:"listening_initial_lookback_days"`
    DedupTTLHours        int     `mapstructure:"listening_dedup_ttl_hours"`
    SinkBatchSize        int     `mapstructure:"listening_sink_batch_size"`
    FetcherWorkers       int     `mapstructure:"listening_fetcher_workers"`
    ParserWorkers        int     `mapstructure:"listening_parser_workers"`
    SentimentWorkers     int     `mapstructure:"listening_sentiment_workers"`
    SinkWorkers          int     `mapstructure:"listening_sink_workers"`
}

type Data365Config struct {
    AccessToken       string  `mapstructure:"data365_access_token"`
    BaseURL           string  `mapstructure:"data365_base_url"`
    RPS               float64 `mapstructure:"data365_rps"`
    Burst             int     `mapstructure:"data365_burst"`
    PollIntervalSec   int     `mapstructure:"data365_poll_interval_sec"`
    PollMaxWaitSec    int     `mapstructure:"data365_poll_max_wait_sec"`
    MaxDocsPerSearch  int     `mapstructure:"data365_max_docs_per_search"`
}

type AIAgentsConfig struct {
    BaseURL            string `mapstructure:"ai_agents_base_url"`
    APIKey             string `mapstructure:"ai_agents_api_key"`
    TimeoutSec         int    `mapstructure:"ai_agents_timeout_sec"`
    SentimentBatchSize int    `mapstructure:"ai_agents_sentiment_batch_size"`
}
```

---

## 14. Saved Views

### What

Users can save filter combinations as named "views" for quick access (e.g., "Negative Twitter mentions", "High engagement posts", "Spanish mentions"). The frontend story "Views sidebar and custom view creation" requires backend support.

### MongoDB Schema — Embedded in `listening_topics`

Views are stored as a subdocument array on the topic, not a separate collection. A view is just a saved set of filters — it doesn't own any data. Embedding avoids extra lookups and keeps view access fast.

```javascript
// Add to listening_topics document
{
  // ... existing fields ...

  "views": [
    {
      "_id": ObjectId,                          // Auto-generated
      "name": "Negative Twitter Mentions",
      "is_default": false,                      // One view can be default (loads on page open)
      "filters": {
        "platform": "twitter",                  // null = all
        "sentiment": "negative",                // null = all
        "content_type": null,                   // null = all
        "date_range": "last_7_days",            // last_7_days, last_30_days, last_90_days, custom
        "date_from": null,                      // Only if date_range = "custom"
        "date_to": null,
        "search": "",                           // Text search query
        "is_bookmarked": false,
        "smart_tags": [],                       // Filter by AI tags
        "author_verified": false,
        "sort_by": "published_at",
        "sort_order": "desc"
      },
      "created_at": ISODate,
      "updated_at": ISODate
    }
  ],
}
```

### Laravel Endpoints

| Method | Endpoint | Purpose |
|---|---|---|
| POST | `/api/listening/topics/{id}/views` | Create a saved view |
| PUT | `/api/listening/topics/{id}/views/{view_id}` | Update view name/filters |
| DELETE | `/api/listening/topics/{id}/views/{view_id}` | Delete a view |

These use MongoDB `$push`, `$set`, and `$pull` on the `views` array. No Go service changes needed — views are just filter presets that the frontend sends as query params to the Go mentions endpoint.

### Go Service Impact

None. When the frontend loads a saved view, it simply passes the saved filters as query parameters to `GET /api/v1/listening/mentions`. The Go service doesn't need to know about views.

---

## 15. Global Settings

### What

Workspace-level listening configuration that applies across all topics. The epic story "Global settings and topic types API" requires a settings schema.

### MongoDB Schema — `listening_settings`

One document per workspace. Created with defaults when the workspace first enables listening.

```javascript
// Collection: listening_settings
{
  "_id": ObjectId,
  "workspace_id": ObjectId,                     // One-to-one with workspace

  // Global Exclude Keywords (applied to ALL topics in this workspace)
  "global_exclude_keywords": ["spam", "bot", "giveaway", "follow4follow"],
  "global_blocked_authors": ["@spam_account", "@bot_network"],
  "global_excluded_subreddits": ["test", "announcements"],   // Reddit-only

  // Default Language & Region (pre-filled when creating new topics)
  "default_language": ["en"],
  "default_regions": [],

  // Default Platforms (pre-filled when creating new topics)
  "default_platforms": {
    "tiktok": true,
    "instagram": true,
    "facebook": true,
    "twitter": true,
    "reddit": true,
    "threads": true
  },

  // Notification Preferences
  "notifications": {
    "volume_spike_enabled": true,
    "sentiment_shift_enabled": true,
    "limit_reached_enabled": true,
    "delivery_channels": ["in_app", "email"],   // in_app, email, webhook
    "webhook_url": null,                         // If webhook enabled
    "email_frequency": "always",                 // always, hourly_digest, daily_digest
  },

  // Custom Topic Types (beyond the 4 defaults)
  "custom_topic_types": [
    { "key": "campaign", "label": "Campaign Tracking" },
    { "key": "event", "label": "Event Monitoring" }
  ],

  // Timestamps
  "created_at": ISODate,
  "updated_at": ISODate
}
```

### Index

```javascript
{ "workspace_id": 1 }   // Unique index
```

### Laravel Endpoints

| Method | Endpoint | Purpose |
|---|---|---|
| GET | `/api/listening/settings` | Get workspace settings (create with defaults if not exists) |
| PUT | `/api/listening/settings` | Update settings |

### Pipeline Impact

The **scheduler** and **parser** need to apply workspace-level filters:
- **Scheduler:** Load settings and attach merged exclude keywords, blocked authors, and excluded subreddits to the work order
- **Parser:** Apply the merged filters during keyword matching and Reddit-specific screening

```go
// In scheduler, when building ListeningWorkOrder:
settings, _ := repo.GetListeningSettings(ctx, workspaceID)
workOrder.ExcludeWords = mergeUnique(topic.Query.ExcludeKeywords, settings.GlobalExcludeKeywords)
workOrder.BlockedAuthors = mergeUnique(topic.Query.ExcludeAuthors, settings.GlobalBlockedAuthors)
workOrder.ExcludedSubreddits = settings.GlobalExcludedSubreddits
```

---

## 16. Inline Mention Reply

### What

Frontend story "Inline mention reply with account selector and AI compose toolbar" allows users to reply to a mention directly from the listening feed using their connected social accounts.

### Architecture

Replying to a social post requires the platform's write API — this is **already handled by the Laravel backend** through existing publishing infrastructure. The Go service does NOT need write access to social platforms.

```
Frontend → Laravel (publish reply) → Platform API (Facebook Graph, Twitter API, etc.)
                                   ↑
                        Uses existing social_integrations tokens
```

### Laravel Endpoint

| Method | Endpoint | Purpose |
|---|---|---|
| POST | `/api/listening/mentions/{mention_id}/reply` | Reply to a mention |

**Request:**
```json
{
    "workspace_id": "5f0c...",
    "account_id": "6601...",            // social_integrations._id (connected account to reply from)
    "text": "Thanks for the feedback!",
    "mention_platform": "twitter",
    "mention_native_id": "123456"       // Platform's post ID
}
```

**Logic:**
1. Validate the workspace owns the `account_id`
2. Validate the account platform matches `mention_platform`
3. Route to the existing platform-specific reply/comment service:
   - Twitter: POST tweet as reply (using `in_reply_to_tweet_id`)
   - Facebook: POST comment on the post
   - Instagram: POST comment on the media
   - Reddit: POST comment on the post
   - TikTok: POST comment (if API supports)
   - Threads: POST reply
4. Return the reply status

### Go Service Impact

None. Reply is a write operation through existing Laravel publishing infrastructure. The Go service only reads data.

### AI Compose Integration

The "AI compose toolbar" reuses the existing AI Agents service. The frontend calls the existing caption/reply generation endpoint before submitting the reply — no new backend work needed.

---

## 17. Concurrency & Locking

### Problem

If a daily cron run and an on-demand trigger fire simultaneously for the same topic, two fetcher instances could fetch overlapping data, causing duplicate processing and wasted Data365 credits.

### Solution: Redis Distributed Lock

Use a Redis lock per topic when a pipeline run starts. Same pattern as distributed locks in the existing codebase.

```go
// In scheduler and on-demand handler, before producing work order:

func (s *Scheduler) acquireTopicLock(ctx context.Context, topicID string) (bool, error) {
    key := fmt.Sprintf("listening:lock:%s", topicID)
    // SET NX with TTL — only one process can acquire
    ok, err := s.redis.SetNX(ctx, key, "locked", 30*time.Minute).Result()
    return ok, err
}

func (s *Scheduler) releaseTopicLock(ctx context.Context, topicID string) {
    key := fmt.Sprintf("listening:lock:%s", topicID)
    s.redis.Del(ctx, key)
}
```

**TTL: 30 minutes** — auto-releases if the pipeline crashes (prevents deadlocks).

**Usage:**
```go
// Scheduler
if acquired, _ := s.acquireTopicLock(ctx, topic.ID); !acquired {
    logger.Info().Str("topic_id", topic.ID).Msg("topic already being processed, skipping")
    continue
}
// ... produce work order ...
// Lock is released by the sink service after pipeline completes
```

**Release point:** The **sink service** releases the lock after all mentions are inserted and cursors are updated. If the pipeline fails, the TTL auto-releases.

---

## 18. Data Retention

### Problem

At 10K mentions/month per active topic, 5 default topic slots per owner/admin account, and 1000 owner accounts:
- 50M mentions/month
- 600M mentions/year
- Storage grows indefinitely

### Solution: Partition-Based TTL

ClickHouse partitions by `toYYYYMM(published_at)`. Drop old partitions after retention period.

```sql
-- Automated cleanup: drop partitions older than 12 months
-- Run as a monthly cron job

ALTER TABLE listening_mentions DROP PARTITION '202503';  -- Drops March 2025 data
ALTER TABLE listening_daily_stats DROP PARTITION '202503';
```

### Implementation

```go
// Monthly cron job in scheduler
func (s *Scheduler) cleanupOldPartitions(ctx context.Context) {
    cutoff := time.Now().AddDate(-1, 0, 0) // 12 months ago
    partition := cutoff.Format("200601")    // e.g., "202503"

    queries := []string{
        fmt.Sprintf("ALTER TABLE listening_mentions DROP PARTITION '%s'", partition),
        fmt.Sprintf("ALTER TABLE listening_daily_stats DROP PARTITION '%s'", partition),
    }
    for _, q := range queries {
        if err := s.clickhouse.Exec(ctx, q); err != nil {
            logger.Error().Err(err).Str("query", q).Msg("failed to drop partition")
        }
    }
}
```

### Retention Policy

| Plan | Retention |
|---|---|
| Default | 12 months |
| Enterprise (future) | 24 months |

Add to `listening_settings`:
```javascript
"retention_months": 12
```

---

## 19. Error Handling & Retry Strategy

### Pipeline Error Scenarios

| Stage | Error | Handling |
|---|---|---|
| **Fetcher** | Data365 search task fails for one platform | Log + skip that platform, continue with others. Update cursor only for succeeded platforms. Sentry alert. |
| **Fetcher** | Data365 API is completely down (5xx) | Retry 5× with exponential backoff (300ms→8s). If all fail, mark work order as failed, don't update `last_fetched_at` so daily cron retries next run. |
| **Fetcher** | Data365 rate limit hit (429) | Respect `Retry-After` header. Pause fetcher goroutine. Auto-resume. |
| **Parser** | Malformed JSON from Data365 | Log + skip the mention. Increment `parse_errors` counter. Sentry if > 10% error rate. |
| **Sentiment** | AI Agents service timeout/error | Insert mentions with `sentiment_label = "pending"`. Background retry job re-processes pending mentions every hour. |
| **Sink** | ClickHouse insert failure | Retry 3× with backoff. If all fail, produce to dead letter topic `listening-dlq`. Manual intervention. |
| **Sink** | MongoDB usage increment failure | Retry 3×. Non-fatal — mention is already in ClickHouse. Usage count may drift slightly. Self-corrects on next successful increment. |

### Dead Letter Queue

```
listening-dlq     # Failed messages from any stage (partitions: 3)
```

Each DLQ message includes:
```go
type DLQMessage struct {
    OriginalTopic string          `json:"original_topic"`  // Which Kafka topic it came from
    Stage         string          `json:"stage"`           // fetcher, parser, sentiment, sink
    Error         string          `json:"error"`           // Error message
    Payload       json.RawMessage `json:"payload"`         // Original message
    Attempts      int             `json:"attempts"`        // How many times retried
    Timestamp     time.Time       `json:"timestamp"`
}
```

### Sentry Integration

Reuse existing Sentry setup (`APP_SENTRY_DSN`). Tag listening errors with:
```go
sentry.ConfigureScope(func(scope *sentry.Scope) {
    scope.SetTag("module", "listening")
    scope.SetTag("stage", "fetcher")         // fetcher, parser, sentiment, sink
    scope.SetTag("platform", "twitter")
    scope.SetTag("topic_id", topicID)
})
```

### Sentiment Retry Job

For mentions inserted with `sentiment_label = "pending"`:

```go
// Hourly cron job
func (s *Scheduler) retrySentiment(ctx context.Context) {
    // Query ClickHouse for pending mentions (max 500 per run)
    query := `
        SELECT mention_id, text, platform
        FROM listening_mentions FINAL
        WHERE sentiment_label = 'pending'
        AND fetched_at > now() - INTERVAL 48 HOUR
        LIMIT 500
    `
    // Send to AI Agents service in batches of 50
    // Re-insert with updated sentiment + new updated_at
}
```

---

## 20. Testing Strategy

### Unit Tests

| Component | What to Test | Mock |
|---|---|---|
| `data365.go` (client) | Request building, response parsing, error handling, rate limiting | HTTP mock server |
| `mention_parser.go` | All 6 platform response formats → unified struct, keyword matching, content hash, engagement calculation | Static JSON fixtures |
| `sentiment_service.go` | Batch building, AI response mapping, fallback on error | Mock AI client |
| `clickhouse_sink.go` | Batch grouping, usage increment logic, limit flag setting | Mock ClickHouse + MongoDB |
| `listening_scheduler.go` | Topic filtering, work order building, cursor handling, lock acquisition | Mock MongoDB + Redis |
| `listening_read.go` | Query building with filters, pagination, parameterized queries | Mock ClickHouse |

### Integration Tests

| Test | Scope | Environment |
|---|---|---|
| **Pipeline end-to-end** | Work order → Fetcher → Parser → Sentiment → Sink | Local Kafka + ClickHouse + MongoDB + Redis (Docker Compose) |
| **Data365 fetcher** | Async POST→poll→GET with real Data365 sandbox | Data365 trial account |
| **API endpoints** | All GET/POST endpoints with auth | Local Go server + test ClickHouse |

### Test Data Fixtures

Create fixtures in `src/services/listening/testdata/`:
```
testdata/
├── data365/
│   ├── facebook_search_response.json
│   ├── twitter_search_response.json
│   ├── instagram_hashtag_response.json
│   ├── tiktok_search_response.json
│   ├── reddit_search_response.json
│   └── threads_search_response.json
├── parsed/
│   └── unified_mentions.json
└── sentiment/
    └── ai_response.json
```

### Running Tests

```bash
# Unit tests
go test ./src/services/listening/... -v

# Integration tests (requires Docker stack)
docker-compose -f src/deployments/docker-compose.test.yml up -d
go test ./src/services/listening/... -v -tags=integration
```

---

## 21. Deployment

### Service Binaries

The existing Go service runs as **separate binaries per microservice**. Listening follows the same pattern:

| Binary | Kafka Consumer Group | Consumes | Produces |
|---|---|---|---|
| `listening-fetcher` | `listening-fetcher-group` | `listening-work-order` | `listening-raw-mentions` |
| `listening-parser` | `listening-parser-group` | `listening-raw-mentions` | `listening-parsed-mentions` |
| `listening-sentiment` | `listening-sentiment-group` | `listening-parsed-mentions` | `listening-enriched-mentions` |
| `listening-sink` | `listening-sink-group` | `listening-enriched-mentions` | (ClickHouse + MongoDB) |

The **API server** and **scheduler** are existing binaries extended with new handlers/flags.

### Dockerfile

Extend the existing multi-stage Dockerfile to build additional binaries:

```dockerfile
# Add to existing Dockerfile build targets
RUN go build -o /app/listening-fetcher ./src/services/listening/fetcher/cmd/
RUN go build -o /app/listening-parser ./src/services/listening/parser/cmd/
RUN go build -o /app/listening-sentiment ./src/services/listening/sentiment/cmd/
RUN go build -o /app/listening-sink ./src/services/listening/sink/cmd/
```

### Kubernetes / Docker Compose

Each binary runs as a separate deployment/service with its own scaling:

```yaml
# docker-compose.listening.yml (dev)
services:
  listening-fetcher:
    build: .
    command: /app/listening-fetcher
    env_file: .env
    depends_on: [kafka, clickhouse, mongodb, redis]
    deploy:
      replicas: 2

  listening-parser:
    build: .
    command: /app/listening-parser
    env_file: .env
    depends_on: [kafka, redis]
    deploy:
      replicas: 2

  listening-sentiment:
    build: .
    command: /app/listening-sentiment
    env_file: .env
    depends_on: [kafka]
    deploy:
      replicas: 1    # Lower — batches calls to AI service

  listening-sink:
    build: .
    command: /app/listening-sink
    env_file: .env
    depends_on: [kafka, clickhouse, mongodb]
    deploy:
      replicas: 2
```

### Scaling Guidelines

| Service | CPU-bound? | IO-bound? | Recommended Replicas |
|---|---|---|---|
| Fetcher | No | Yes (Data365 API calls, 1-5 min waits) | 2-4 |
| Parser | Yes (JSON parsing, hashing) | No | 2-3 |
| Sentiment | No | Yes (AI service HTTP calls) | 1-2 |
| Sink | No | Yes (ClickHouse batch inserts) | 2-3 |

---

## 22. Instagram Hashtag Auto-Handling

### Problem

Instagram only supports hashtag search via Data365. If a user configures keyword "acme corp" (no `#`), the Data365 Instagram search will fail or return nothing.

### Solution: Parser-Level Keyword Transformation

In the **fetcher**, when building Data365 search requests for Instagram:

```go
func (f *Data365Fetcher) buildInstagramKeywords(keywords []string) []string {
    var igKeywords []string
    for _, kw := range keywords {
        if strings.HasPrefix(kw, "#") {
            // Already a hashtag — use as-is
            igKeywords = append(igKeywords, kw)
        } else {
            // Convert to hashtag: "acme corp" → "#acmecorp" (remove spaces, lowercase, prepend #)
            normalized := strings.ToLower(strings.ReplaceAll(kw, " ", ""))
            igKeywords = append(igKeywords, "#"+normalized)
        }
    }
    return igKeywords
}
```

**Examples:**
| User Keyword | Instagram Search |
|---|---|
| `#acme` | `#acme` (unchanged) |
| `acme corp` | `#acmecorp` |
| `Acme Corp` | `#acmecorp` |
| `#AcmeCorp` | `#AcmeCorp` (unchanged) |

### Frontend Guidance

The topic creation UI should:
1. Show a notice next to the keyword input: "Instagram searches by hashtag only. Non-hashtag keywords will be auto-converted (e.g., 'acme corp' → '#acmecorp')."
2. Preview the Instagram-specific keywords so users can add explicit hashtag variants if needed.

---

## 23. Cross-Topic Bookmarks

### Problem

Frontend story "Bookmarks page with search and filters" implies a page showing all bookmarked mentions across all topics in a workspace. The current `GET /api/v1/listening/mentions?is_bookmarked=true` only works per-topic.

### Solution: New Go Endpoint

#### GET /api/v1/listening/bookmarks

```
GET /api/v1/listening/bookmarks    (JWT protected)

Query Parameters:
  workspace_id   string   required
  page           int      optional   Default: 1
  per_page       int      optional   Default: 20
  platform       string   optional
  sentiment      string   optional
  search         string   optional
  date_from      string   optional
  date_to        string   optional
  topic_id       string   optional   Filter by specific topic (optional)
  sort_by        string   optional   published_at, total_engagement
  sort_order     string   optional   desc, asc

Response: Same structure as GET /api/v1/listening/mentions, but with topic info added to each mention
```

**ClickHouse query:**
```sql
SELECT m.*, t.topic_name
FROM listening_mentions m FINAL
-- Topic name join handled in Go by fetching topics from MongoDB
WHERE m.workspace_id = {workspace_id:String}
  AND m.is_bookmarked = true
ORDER BY m.published_at DESC
LIMIT {per_page:UInt32} OFFSET {offset:UInt32}
```

Since ClickHouse doesn't join with MongoDB, the Go handler:
1. Queries ClickHouse for bookmarked mentions (gets `topic_id` per mention)
2. Batch-fetches topic names from MongoDB for the unique `topic_id`s in results
3. Merges topic name into response

### Directory Update

Add to `listening_mentions_handler.go`:
```go
// GET /api/v1/listening/bookmarks
func (h *ListeningHandler) GetBookmarks(w http.ResponseWriter, r *http.Request) { ... }
```

---

## 24. API Rate Limiting

### Problem

The new Go query endpoints (`GET /mentions`, `GET /analytics/*`) are exposed to the frontend. Without rate limiting, a single user or bot could overload ClickHouse with expensive queries.

### Solution: Per-Workspace Rate Limiter

Reuse the existing `golang.org/x/time/rate` package (already imported):

```go
// File: src/api/middleware/rate_limiter.go

type WorkspaceRateLimiter struct {
    limiters sync.Map   // map[workspaceID]*rate.Limiter
    rps      float64
    burst    int
}

func NewWorkspaceRateLimiter(rps float64, burst int) *WorkspaceRateLimiter {
    return &WorkspaceRateLimiter{rps: rps, burst: burst}
}

func (rl *WorkspaceRateLimiter) getLimiter(workspaceID string) *rate.Limiter {
    if limiter, ok := rl.limiters.Load(workspaceID); ok {
        return limiter.(*rate.Limiter)
    }
    limiter := rate.NewLimiter(rate.Limit(rl.rps), rl.burst)
    rl.limiters.Store(workspaceID, limiter)
    return limiter
}

func (rl *WorkspaceRateLimiter) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        workspaceID := r.URL.Query().Get("workspace_id")
        if workspaceID == "" {
            workspaceID = "unknown"
        }
        if !rl.getLimiter(workspaceID).Allow() {
            http.Error(w, `{"status":"error","error":"rate_limit_exceeded"}`, http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

### Rate Limits

| Endpoint Group | Rate | Burst |
|---|---|---|
| `GET /listening/mentions` | 10 req/sec per workspace | 20 |
| `GET /listening/analytics/*` | 5 req/sec per workspace | 10 |
| `POST /listening/*` (mutations) | 2 req/sec per workspace | 5 |

### Configuration

```bash
APP_LISTENING_API_MENTIONS_RPS=10
APP_LISTENING_API_MENTIONS_BURST=20
APP_LISTENING_API_ANALYTICS_RPS=5
APP_LISTENING_API_ANALYTICS_BURST=10
```

---

## 14. Resolved Decisions

| # | Decision | Resolution | Rationale |
|---|---|---|---|
| 1 | **Instagram search** | Hashtag-based only | Data365 has no broad keyword search for Instagram. Users must add hashtag keywords (e.g., `#acmecorp`). UI should guide this. |
| 2 | **Comments** | Skip in v1 (posts only, 1 credit each) | Data365 can't search within comments. Fetching all comments at 5x credit cost returns unfiltered noise. Can add selectively in v2. |
| 3 | **Twitter replies/quotes** | Show as separate mentions | Each reply/quote has independent sentiment, author, and engagement. Data365 returns them as separate objects. Industry standard (Brand24, Sprout Social). |
| 4 | **Cross-topic matches** | Duplicate rows per topic | The same post can appear once for each matching topic. This keeps query shapes simple and matches the current product decision. |
| 5 | **Topic slot ownership** | Owner/admin-level slots, allocatable across workspaces | Topic-slot billing is not per workspace. Laravel remains the source of truth for slot ownership and allocation. |
| 6 | **Authorization** | Trust Laravel | Laravel authorizes workspace/topic ownership. Go assumes authenticated requests have already passed product-layer access checks. |

---

## 15. Open Questions

| # | Question | Options | Impact |
|---|---|---|---|
| 1 | **Alert delivery channel** | a) Reuse main Pusher + add `listening_alert` Redis channel to socket service<br>b) Create new `ListeningPusher` instance (like `AnalyticsPusher`)<br>c) Skip real-time alerts in v1, email-only | Architecture of alert notifications |
| 2 | **Webhook alerts** | a) Include in v1 — user configures webhook URL per topic<br>b) Defer to v2 | Scope of v1 |
| 3 | **Export (PDF/CSV)** | a) Reuse existing Laravel export infra<br>b) Build in Go service<br>c) Defer | Story scope |
| 4 | **AI sentiment fallback** | a) Insert with `sentiment_label = "unknown"` and retry later<br>b) Block pipeline until AI responds<br>c) Use local VADER as fallback (add `govader` dependency) | Pipeline resilience |
| 5 | **Initial sync depth** | a) 30 days (current design)<br>b) 7 days (cheaper, faster)<br>c) User-configurable | First-time experience vs cost |
