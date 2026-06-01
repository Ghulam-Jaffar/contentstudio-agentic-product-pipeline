# ContentStudio Analytics API Endpoints Reference

Complete API reference with routes, payloads, and responses for every analytics endpoint.

---

## Table of Contents

1. [Common Request Parameters](#1-common-request-parameters)
2. [Facebook Analytics](#2-facebook-analytics)
3. [Facebook Competitor Analytics](#3-facebook-competitor-analytics)
4. [Instagram Analytics](#4-instagram-analytics)
5. [Instagram Competitor Analytics](#5-instagram-competitor-analytics)
6. [LinkedIn Analytics](#6-linkedin-analytics)
7. [YouTube Analytics](#7-youtube-analytics)
8. [TikTok Analytics](#8-tiktok-analytics)
9. [Pinterest Analytics](#9-pinterest-analytics)
10. [Twitter/X Analytics](#10-twitterx-analytics)
11. [Twitter/X Settings](#11-twitterx-settings)
12. [Cross-Platform Overview V1](#12-cross-platform-overview-v1)
13. [Cross-Platform Overview V2](#13-cross-platform-overview-v2)
14. [Campaign and Label Analytics](#14-campaign-and-label-analytics)
15. [Dashboard Analytics](#15-dashboard-analytics)
16. [Share Link Management](#16-share-link-management)
17. [Reports](#17-reports)
18. [Scheduled Reports](#18-scheduled-reports)
19. [Analytics Job Triggers](#19-analytics-job-triggers)
20. [AI Insights](#20-ai-insights)

---

## API Summary

| # | Section | Category | Endpoints |
|---|---------|----------|-----------|
| 2 | Facebook Analytics | Platform | 13 |
| 3 | Facebook Competitor Analytics | Platform | 13 |
| 4 | Instagram Analytics | Platform | 13 |
| 5 | Instagram Competitor Analytics | Platform | 11 |
| 6 | LinkedIn Analytics | Platform | 9 |
| 7 | YouTube Analytics | Platform | 11 |
| 8 | TikTok Analytics | Platform | 6 |
| 9 | Pinterest Analytics | Platform | 10 |
| 10 | Twitter/X Analytics | Platform | 6 |
| 11 | Twitter/X Settings | Settings | 5 |
| 12 | Cross-Platform Overview V1 | Overview | 6 |
| 13 | Cross-Platform Overview V2 | Overview | 6 |
| 14 | Campaign & Label Analytics | Analytics | 4 |
| 15 | Dashboard Analytics | Dashboard | 3 |
| 16 | Share Link Management | Management | 7 |
| 17 | Reports | Reports | 4 |
| 18 | Scheduled Reports | Reports | 4 |
| 19 | Analytics Job Triggers | Jobs | 2 |
| 20 | AI Insights | AI | 7 |
| | **Total** | | **140** |

---

## 1. Common Request Parameters

Most analytics endpoints accept these core parameters:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `workspace_id` | string | Yes | MongoDB workspace identifier |
| `date` | string | Yes | Format: `"YYYY-MM-DD - YYYY-MM-DD"` |
| `timezone` | string | Yes | IANA timezone string (e.g. `"UTC"`, `"America/New_York"`) |
| `{platform}_id` | string/array | Yes | Platform account identifier(s) |
| `media_type` | array | No | Filter by content type |
| `limit` | int | No | Max results (default varies 5-20) |
| `order_by` | string | No | Sort column for top posts |

**Previous Period:** Automatically computed as `previous_start = start - (end - start)`, `previous_end = start`.

**Growth Formula:** `round((current - previous) / max(previous, 1) * 100, 2)`. Returns `"N/A"` when previous is 0.

---

## 2. Facebook Analytics

### 2.1 Summary

**Route:** `GET /analytics/overview/facebook/summary`

**Payload:**
```json
{
  "workspace_id": "string (required)",
  "facebook_id": ["string"] ,
  "date": "YYYY-MM-DD - YYYY-MM-DD",
  "timezone": "string",
  "media_type": ["text", "link", "images", "videos", "carousel", "share", "reels", "others"],
  "limit": 15,
  "order_by": "total_engagement"
}
```

**Response:**
```json
{
  "overview": {
    "current": {
      "page_fans": 0,
      "page_follows": 0,
      "page_impressions": 0,
      "page_impressions_paid": 0,
      "page_impressions_organic": 0,
      "page_post_engagements": 0,
      "total_posts": 0,
      "total_engagement": 0,
      "post_clicks": 0,
      "reactions": 0,
      "comments": 0,
      "shares": 0,
      "engagement_rate": 0.0
    },
    "previous": { "...same keys..." }
  }
}
```

### 2.2 Audience Growth

**Route:** `GET /analytics/overview/facebook/overviewAudienceGrowth`

**Payload:** Same as 2.1

**Response:**
```json
{
  "audience_growth": {
    "fan_count": [0],
    "page_fans_by_like": [0],
    "page_fans_by_unlike": [0],
    "date_data": ["YYYY-MM-DD"]
  },
  "audience_growth_rollup": {
    "current": { "fan_count": 0, "page_fans_by_like": 0, "page_fans_by_unlike": 0 },
    "previous": { "...same keys..." }
  }
}
```

### 2.3 Publishing Behaviour

**Route:** `GET /analytics/overview/facebook/overviewPublishingBehaviour`

**Payload:** Same as 2.1

**Response:**
```json
{
  "publishing_behaviour": {
    "text": 0, "link": 0, "images": 0, "videos": 0,
    "carousel": 0, "share": 0, "reels": 0, "others": 0
  },
  "publishing_behaviour_rollup": {
    "current": { "total_posts": 0, "engagement": 0, "impressions": 0 },
    "previous": { "...same keys..." }
  }
}
```

### 2.4 Top Posts

**Route:** `GET /analytics/overview/facebook/overviewTopPosts`

**Payload:** Same as 2.1

**Response:**
```json
{
  "top_posts": [
    {
      "page_id": "string",
      "post_id": "string",
      "permalink": "string",
      "status_type": "string",
      "media_type": "string",
      "video_id": "string",
      "category": "string",
      "published_by": "string",
      "like": 0, "love": 0, "haha": 0, "wow": 0, "sad": 0, "angry": 0,
      "total": 0, "shares": 0, "comments": 0, "post_clicks": 0,
      "total_engagement": 0, "post_engaged_users": 0,
      "day_of_week": 0, "hour_of_day": 0,
      "created_time": "datetime", "saving_time": "datetime",
      "caption": "string", "description": "string",
      "full_picture": "string", "link": "string",
      "post_impressions": 0, "post_impressions_unique": 0,
      "post_impressions_paid": 0, "post_impressions_organic": 0,
      "post_video_views": 0,
      "media_assets": [
        { "media_id": "string", "caption": "string", "link": "string", "assetType": "string" }
      ]
    }
  ]
}
```

### 2.5 Get Top Posts (Paginated)

**Route:** `GET /analytics/overview/facebook/getTopPosts`

**Payload:** Same as 2.1

**Response:** Same structure as 2.4

### 2.6 Active Users

**Route:** `GET /analytics/overview/facebook/overviewActiveUsers`

**Payload:** Same as 2.1

**Response:**
```json
{
  "active_users": {
    "active_users_hours": {
      "buckets": [0, 1, 2, "...23"],
      "values": [0],
      "highest_hour": 0,
      "highest_value": 0
    },
    "active_users_days": {
      "date": [0, 1, 2, "...6"],
      "value": [0]
    }
  }
}
```

### 2.7 Page Impressions

**Route:** `GET /analytics/overview/facebook/overviewImpressions`

**Payload:** Same as 2.1

**Response:**
```json
{
  "impressions": {
    "date": ["YYYY-MM-DD"],
    "page_impressions": [0],
    "paid_impressions": [0],
    "organic_impressions": [0],
    "viral_impressions": [0]
  },
  "impressions_rollup": {
    "current": { "page_impressions": 0, "paid_impressions": 0, "organic_impressions": 0 },
    "previous": { "...same keys..." }
  }
}
```

### 2.8 Engagement

**Route:** `GET /analytics/overview/facebook/overviewEngagement`

**Payload:** Same as 2.1

**Response:**
```json
{
  "engagement": {
    "engagement": {
      "date": ["YYYY-MM-DD"],
      "engagement": [0],
      "comments": [0],
      "shares": [0],
      "reactions": [0]
    },
    "engagement_rollup": {
      "current": { "engagement": 0, "comments": 0, "shares": 0, "reactions": 0 },
      "previous": { "...same keys..." }
    }
  }
}
```

### 2.9 Reels Analytics

**Route:** `GET /analytics/overview/facebook/overviewReelsAnalytics`

**Payload:** Same as 2.1

**Response:**
```json
{
  "reels": {
    "date": ["YYYY-MM-DD"],
    "reels_plays": [0],
    "reels_watch_time": [0],
    "reels_engagement": [0]
  },
  "reels_rollup": {
    "current": { "reels_plays": 0, "reels_watch_time": 0, "reels_engagement": 0 },
    "previous": { "...same keys..." }
  }
}
```

### 2.10 Video Insights

**Route:** `GET /analytics/overview/facebook/overviewVideoInsights`

**Payload:** Same as 2.1

**Response:**
```json
{
  "video_insights": {
    "date": ["YYYY-MM-DD"],
    "video_views": [0],
    "video_watch_time": [0],
    "video_engagement": [0]
  },
  "video_rollup": {
    "current": { "video_views": 0, "video_watch_time": 0, "video_engagement": 0 },
    "previous": { "...same keys..." }
  }
}
```

### 2.11 Demographics

**Route:** `GET /analytics/overview/facebook/demographics`

**Payload:** Same as 2.1

**Response:**
```json
{
  "audience_gender": { "male": 0, "female": 0 },
  "fans": 0,
  "audience_age": [{ "age_range": "18-24", "value": 0 }],
  "max_gender_age": { "max_data": {} },
  "audience_country": { "US": 0, "UK": 0 },
  "audience_city": { "New York": 0 }
}
```

### 2.12 Demographics Overview

**Route:** `GET /analytics/overview/facebook/overviewDemographics`

**Payload:** Same as 2.1

**Response:** Same structure as 2.11

### 2.13 Audience Location

**Route:** `GET /analytics/overview/facebook/overviewAudienceLocation`

**Payload:** Same as 2.1

**Response:**
```json
{
  "audience_country": { "US": 0 },
  "audience_city": { "New York": 0 }
}
```

---

## 3. Facebook Competitor Analytics

### 3.1 Search Competitor

**Route:** `POST /analytics/overview/facebook/competitor/search`

**Payload:**
```json
{
  "search": "string (required - page name)"
}
```

**Response:**
```json
{
  "data": [
    {
      "competitor_id": "string",
      "name": "string",
      "image": "string (url)",
      "link": "string",
      "verification_status": "string",
      "location": {}
    }
  ]
}
```

### 3.2 Data Table Metrics

**Route:** `POST /analytics/overview/facebook/competitor/dataTableMetrics`

**Payload:**
```json
{
  "workspace_id": "string (required)",
  "facebook_id": ["string"],
  "date": "YYYY-MM-DD - YYYY-MM-DD",
  "timezone": "string",
  "sort_order": "followersCount",
  "accounts": []
}
```

**Response:**
```json
{
  "data_table_metrics": [
    {
      "facebook_id": "string",
      "page_name": "string",
      "image": "string",
      "followersCount": 0,
      "fanCount": 0,
      "averagePostsPerWeek": 0.0,
      "engagementRate": 0.0,
      "followersCountDiff": 0.0,
      "engagementRateDiff": 0.0
    }
  ]
}
```

### 3.3 Posting Activity by Types

**Route:** `POST /analytics/overview/facebook/competitor/postingActivityGraphByTypes`

**Payload:** Same as 3.2

**Response:**
```json
{
  "data": [
    {
      "facebook_id": "string",
      "page_name": "string",
      "image": "string",
      "mediaType": "text|link|image|video|carousel|share",
      "avgTotalEngagements": 0.0,
      "totalPosts": 0,
      "postsPerWeek": 0.0
    }
  ]
}
```

### 3.4 Posting Activity by Specific Type

**Route:** `POST /analytics/overview/facebook/competitor/postingActivityBySpecificType`

**Payload:**
```json
{
  "workspace_id": "string",
  "facebook_id": ["string"],
  "date": "YYYY-MM-DD - YYYY-MM-DD",
  "timezone": "string",
  "media_type": "text|link|image|video|carousel|share"
}
```

**Response:**
```json
{
  "data": [
    {
      "facebook_id": "string",
      "page_name": "string",
      "image": "string",
      "followers_count": 0,
      "average_engagement_per_post": 0.0,
      "average_posts_per_week": 0.0,
      "total_posts": 0
    }
  ]
}
```

### 3.5 Top and Least Performing Posts

**Route:** `POST /analytics/overview/facebook/competitor/topAndLeastPerformingPosts`

**Payload:** Same as 3.2

**Response:**
```json
{
  "data": [
    {
      "facebook_id": "string",
      "page_name": "string",
      "image": "string",
      "followers_count": 0,
      "top": [
        {
          "id": "string",
          "post_engagement": 0,
          "like": 0, "love": 0, "haha": 0, "wow": 0, "sad": 0, "angry": 0, "thankful": 0,
          "comments": 0, "shares": 0,
          "caption": "string", "media_type": "string", "permalink": "string",
          "hashtags": "string", "created_at": "datetime",
          "media": [{ "id": "string", "link": "string", "asset_type": "string" }]
        }
      ],
      "bottom": ["...same structure as top..."]
    }
  ]
}
```

### 3.6 Top Hashtags

**Route:** `POST /analytics/overview/facebook/competitor/topHashtags`

**Payload:** Same as 3.2 + `"limit": 7`

**Response:**
```json
{
  "data": [
    { "hashtag": "string", "hashtag_count": 0, "total_engagement": 0, "facebook_id": "string" }
  ]
}
```

### 3.7 Individual Hashtag Data

**Route:** `POST /analytics/overview/facebook/competitor/individualHashtagData`

**Payload:** Same as 3.2 + `"hashtag": "string"`

**Response:**
```json
{
  "data": [
    {
      "hashtag": "string", "post_id": "string", "caption": "string",
      "total_engagement": 0, "like": 0, "love": 0, "comments": 0, "shares": 0
    }
  ]
}
```

### 3.8 Biography Data

**Route:** `POST /analytics/overview/facebook/competitor/biographyData`

**Payload:** Same as 3.2

**Response:**
```json
{
  "data": [
    {
      "facebook_id": "string", "page_name": "string", "image": "string",
      "biography": "string", "biography_length": 0, "followers_count": 0
    }
  ]
}
```

### 3.9 Followers Growth Comparison

**Route:** `POST /analytics/overview/facebook/competitor/followersGrowthComparison`

**Payload:** Same as 3.2

**Response:**
```json
{
  "data": [
    {
      "facebook_id": "string", "page_name": "string", "image": "string",
      "date": ["YYYY-MM-DD"], "followers_count": [0]
    }
  ]
}
```

### 3.10 Post Engagement Over Time

**Route:** `POST /analytics/overview/facebook/competitor/postEngagementOverTime`

**Payload:** Same as 3.2

**Response:**
```json
{
  "data": [
    { "date": "YYYY-MM-DD", "total_engagement": 0, "avg_engagement_per_post": 0.0 }
  ]
}
```

### 3.11 Post Engagement by Competitor

**Route:** `POST /analytics/overview/facebook/competitor/postEngagementByCompetitor`

**Payload:** Same as 3.2

**Response:**
```json
{
  "data": [
    {
      "facebook_id": "string", "page_name": "string", "image": "string",
      "followers_count": 0, "total_engagement": 0, "total_posts": 0, "avg_engagement_per_post": 0.0
    }
  ]
}
```

### 3.12 Post React Distribution

**Route:** `POST /analytics/overview/facebook/competitor/postReactDistribution`

**Payload:** Same as 3.2

**Response:**
```json
{
  "data": [
    {
      "like": 0, "love": 0, "haha": 0, "wow": 0, "sad": 0, "angry": 0,
      "thankful": 0, "total_reactions": 0
    }
  ]
}
```

### 3.13 Post Type Distribution

**Route:** `POST /analytics/overview/facebook/competitor/postTypeDistribution`

**Payload:** Same as 3.2

**Response:**
```json
{
  "data": [
    {
      "facebook_id": "string", "page_name": "string",
      "mediaType": "string", "TotalEngagements": 0, "totalPosts": 0,
      "postsPerWeek": 0.0, "postsPerDay": 0.0
    }
  ]
}
```

---

## 4. Instagram Analytics

### 4.1 Summary

**Route:** `POST /analytics/overview/instagram/summary`

**Payload:**
```json
{
  "workspace_id": "string (required)",
  "instagram_id": ["string"],
  "date": "YYYY-MM-DD - YYYY-MM-DD",
  "timezone": "string"
}
```

**Response:**
```json
{
  "overview": {
    "summary": {
      "current": {
        "post_engagement": 0, "post_views": 0, "post_reach": 0,
        "post_saves": 0, "post_reactions": 0, "post_comments": 0,
        "profile_views": 0, "followers_count": 0, "follows_count": 0,
        "accounts_engaged": 0, "profile_engagement": 0, "profile_impressions": 0,
        "profile_reach": 0, "doc_count": 0, "total_stories": 0,
        "total_posts": 0, "eng_rate": 0.0
      },
      "previous": { "...same keys..." }
    }
  }
}
```

### 4.2 Audience Growth

**Route:** `POST /analytics/overview/instagram/audience_growth`

**Payload:** Same as 4.1

**Response:**
```json
{
  "overview": {
    "audience_growth": {
      "show_data": true,
      "followers": [0],
      "followers_daily": [0],
      "buckets": ["YYYY-MM-DD"]
    },
    "audience_growth_rollup": {
      "current": { "follower_count": 0, "follower_gained": 0, "dates": "YYYY-MM-DD" },
      "previous": { "...same keys..." }
    }
  }
}
```

### 4.3 Publishing Behaviour

**Route:** `POST /analytics/overview/instagram/publishing_behaviour`

**Payload:** Same as 4.1 + `"media_type": ["string"]`

**Response:**
```json
{
  "overview": {
    "publishing_behaviour": {
      "likes": [0], "comments": [0], "saved": [0], "engagement": [0],
      "reach": [0], "impressions": [0], "views": [0], "total_posts": [0],
      "buckets": ["YYYY-MM-DD"]
    },
    "publishing_behaviour_rollup": [{ "date": "YYYY-MM-DD", "value": 0 }]
  }
}
```

### 4.4 Top Posts

**Route:** `POST /analytics/overview/instagram/top_posts`

**Payload:** Same as 4.1 + `"limit": 5`

**Response:**
```json
{
  "top_posts": [
    {
      "media_id": "string", "like_count": 0, "comments_count": 0,
      "engagement": 0, "reach": 0, "impressions": 0,
      "media_type": "string", "post_created_at": "datetime"
    }
  ]
}
```

### 4.5 Get Top Posts (Paginated)

**Route:** `POST /analytics/overview/instagram/getTopPosts`

**Payload:** Same as 4.1 + `"limit": 15`

**Response:** Same structure as 4.4

### 4.6 Active Users

**Route:** `POST /analytics/overview/instagram/active_users`

**Payload:** Same as 4.1

**Response:**
```json
{
  "overview": {
    "active_users_hours": {
      "buckets": [0, 1, "...23"], "values": [0],
      "highest_value": 0, "highest_hour": 0
    },
    "active_users_days": {
      "buckets": ["Monday", "Tuesday", "..."],
      "values": [0], "highest_value": 0, "highest_day": "string"
    }
  }
}
```

### 4.7 Impressions

**Route:** `POST /analytics/overview/instagram/impressions`

**Payload:** Same as 4.1

**Response:**
```json
{
  "overview": {
    "impressions": { "impressions_array": [0], "dates": ["YYYY-MM-DD"] },
    "impressions_rollup": {
      "current": { "total_impressions": 0 },
      "previous": { "...same keys..." }
    }
  }
}
```

### 4.8 Engagement

**Route:** `POST /analytics/overview/instagram/engagement`

**Payload:** Same as 4.1

**Response:**
```json
{
  "overview": {
    "engagements": { "engagement_array": [0], "dates": ["YYYY-MM-DD"] },
    "engagements_rollup": {
      "current": { "total_engagement": 0 },
      "previous": { "...same keys..." }
    }
  }
}
```

### 4.9 Hashtags

**Route:** `POST /analytics/overview/instagram/hashtags`

**Payload:** Same as 4.1

**Response:**
```json
{
  "overview": {
    "hashtags": { "hashtags": ["string"], "counts": [0] },
    "hashtags_rollup": {
      "current": { "total_hashtags": 0 },
      "previous": { "...same keys..." }
    }
  }
}
```

### 4.10 Stories Performance

**Route:** `POST /analytics/overview/instagram/stories_performance`

**Payload:** Same as 4.1

**Response:**
```json
{
  "overview": {
    "stories_performance": {
      "likes": [0], "comments": [0], "exits": [0],
      "reach": [0], "impressions": [0], "replies": [0],
      "buckets": ["YYYY-MM-DD"]
    },
    "stories_rollup": {
      "current": {},
      "previous": {}
    }
  }
}
```

### 4.11 Reels Performance

**Route:** `POST /analytics/overview/instagram/reels_performance`

**Payload:** Same as 4.1

**Response:**
```json
{
  "overview": {
    "reels": {
      "likes": [0], "comments": [0], "saved": [0], "engagement": [0],
      "reach": [0], "impressions": [0], "views": [0], "plays": [0],
      "buckets": ["YYYY-MM-DD"]
    },
    "reels_rollup": { "current": {}, "previous": {} }
  }
}
```

### 4.12 Age/Gender Demographics

**Route:** `POST /analytics/overview/instagram/audience_age`

**Payload:** Same as 4.1

**Response:**
```json
{
  "audience_age": [{ "age_range": "18-24", "value": 0 }],
  "audience_gender": [{ "gender": "male", "value": 0 }],
  "max_audience_age": [{ "age": "25-34" }]
}
```

### 4.13 Location Demographics

**Route:** `POST /analytics/overview/instagram/country_city`

**Payload:** Same as 4.1

**Response:**
```json
{
  "audience_city": { "New York": 0 },
  "audience_country": { "US": 0 }
}
```

---

## 5. Instagram Competitor Analytics

### 5.1 Search Competitor

**Route:** `POST /analytics/overview/instagram/competitor/search`

**Payload:**
```json
{
  "search": "string (required - username or URL)"
}
```

**Response:**
```json
{
  "data": [
    {
      "biography": "string", "competitor_id": "string", "id": "string",
      "name": "string", "image": "string", "slug": "string"
    }
  ]
}
```

### 5.2 Add/Update Competitor Report

**Route:** `POST /analytics/overview/instagram/competitor/addUpdateCompetitorReport`

**Payload:**
```json
{
  "_id": "string (optional, for updates)",
  "platform_type": "INSTAGRAM",
  "workspace_id": "string",
  "name": "string",
  "created_by_user_id": "string",
  "updated_by_user_id": "string",
  "competitors": [
    { "competitor_id": "string", "name": "string", "image": "string", "slug": "string" }
  ]
}
```

**Response:**
```json
{
  "data": {
    "_id": "string", "workspace_id": "string", "platform_type": "INSTAGRAM",
    "name": "string", "competitors": [{ "competitor_id": "string", "name": "string" }],
    "created_at": "datetime", "updated_at": "datetime"
  }
}
```

### 5.3 Get Competitor Report

**Route:** `POST /analytics/overview/instagram/competitor/getCompetitorReport`

**Payload:** `{ "_id": "string" }`

**Response:** Same structure as 5.2 response

### 5.4 List Competitor Reports

**Route:** `POST /analytics/overview/instagram/competitor/getCompetitorReportsByWorkspace`

**Payload:** `{ "workspace_id": "string", "platform_type": "INSTAGRAM" }`

**Response:** `{ "data": [ ...array of reports same as 5.2... ] }`

### 5.5 Delete Competitor Report

**Route:** `POST /analytics/overview/instagram/competitor/deleteCompetitorReport`

**Payload:** `{ "_id": "string" }`

**Response:** `{ "res": true }`

### 5.6 Data Table Metrics

**Route:** `POST /analytics/overview/instagram/competitor/dataTableMetrics`

**Payload:**
```json
{
  "workspace_id": "string", "report_id": "string",
  "date": "YYYY-MM-DD - YYYY-MM-DD",
  "accounts": { "competitor_id": { "image": "", "name": "", "slug": "" } },
  "sort_order": "followersCount"
}
```

**Response:**
```json
{
  "data_table_metrics": [
    {
      "business_account_id": "string", "name": "string", "image": "string", "slug": "string",
      "followersCount": 0, "averagePostsPerWeek": 0.0, "engagementRate": 0.0,
      "followersCountDiff": 0, "averagePostsPerWeekDiff": 0, "engagementRateDiff": 0
    }
  ]
}
```

### 5.7 Posting Activity by Types

**Route:** `POST /analytics/overview/instagram/competitor/postingActivityGraphByTypes`

**Payload:** Same as 5.6

**Response:**
```json
{
  "data": [
    {
      "business_account_id": "string", "name": "string", "image": "string",
      "avgTotalEngagements": 0.0, "reelsCount": 0, "carouselCount": 0,
      "imageCount": 0, "videoCount": 0
    }
  ]
}
```

### 5.8 Followers Growth Comparison

**Route:** `POST /analytics/overview/instagram/competitor/followersGrowthComparison`

**Payload:** Same as 5.6

**Response:**
```json
{
  "data": [
    {
      "business_account_id": "string", "name": "string",
      "dates": ["YYYY-MM-DD"], "followers_growth": [0]
    }
  ]
}
```

### 5.9 Top and Least Performing Posts

**Route:** `POST /analytics/overview/instagram/competitor/topAndLeastPerformingPosts`

**Payload:** Same as 5.6

**Response:**
```json
{
  "data": [
    {
      "business_account_id": "string", "media_id": "string", "caption": "string",
      "media_type": "string", "like_count": 0, "comments_count": 0,
      "engagement": 0, "reach": 0, "saved": 0, "post_created_at": "datetime",
      "performance_rank": "top|least"
    }
  ]
}
```

### 5.10 Top Hashtags

**Route:** `POST /analytics/overview/instagram/competitor/topHashtags`

**Payload:** Same as 5.6 + `"limit": 7`

**Response:**
```json
{
  "data": [{ "hashtag": "string", "count": 0, "competitors_using": ["string"] }]
}
```

### 5.11 Biography Data

**Route:** `POST /analytics/overview/instagram/competitor/biographyData`

**Payload:** Same as 5.6

**Response:**
```json
{
  "data": [
    {
      "business_account_id": "string", "name": "string",
      "biography": "string", "biography_length": 0, "website": "string"
    }
  ]
}
```

---

## 6. LinkedIn Analytics

### 6.1 Summary

**Route:** `GET /analytics/overview/linkedin/summary`

**Payload:**
```json
{
  "workspace_id": "string (required)",
  "linkedin_id": "string|array",
  "date": "YYYY-MM-DD - YYYY-MM-DD",
  "timezone": "string"
}
```

**Response:**
```json
{
  "overview": {
    "current": {
      "post_comments": 0, "post_likes": 0, "total_engagement": 0,
      "total_posts": 0, "post_shares": 0, "post_clicks": 0,
      "followers": 0, "page_views": 0, "page_reach": 0,
      "page_impressions": 0, "page_unique_visitors": 0,
      "engagement_rate": 0.0
    },
    "previous": { "...same keys..." }
  }
}
```

### 6.2 Audience Growth

**Route:** `GET /analytics/overview/linkedin/audienceGrowth`

**Payload:** Same as 6.1

**Response:**
```json
{
  "audience_growth": {
    "show_data": true,
    "organic_follower_count": [0], "organic_followers_daily": [0],
    "paid_follower_count": [0], "paid_followers_daily": [0],
    "total_follower_count": [0], "total_followers_daily": [0],
    "buckets": ["YYYY-MM-DD"]
  },
  "audience_growth_rollup": {
    "current": { "organic_follower_count": 0, "paid_follower_count": 0, "total_follower_count": 0 },
    "previous": { "...same keys..." }
  }
}
```

### 6.3 Page Views

**Route:** `GET /analytics/overview/linkedin/pageViews`

**Payload:** Same as 6.1

**Response:**
```json
{
  "page_views": {
    "desktop_page_views": [0], "mobile_page_views": [0], "total_page_views": [0],
    "desktop_page_views_daily": [0], "mobile_page_views_daily": [0],
    "total_page_views_daily": [0], "show_data": 0, "buckets": ["YYYY-MM-DD"]
  },
  "page_views_rollup": {
    "current": { "sum_page_views": 0, "avg_page_views": 0.0 },
    "previous": { "...same keys..." }
  }
}
```

### 6.4 Publishing Behaviour

**Route:** `GET /analytics/overview/linkedin/publishingBehaviour`

**Payload:** Same as 6.1

**Response:**
```json
{
  "publishing_behaviour": {
    "likes": [0], "comments": [0], "shares": [0], "engagement": [0],
    "buckets": ["YYYY-MM-DD"]
  },
  "publishing_behaviour_rollup": {
    "current": [{ "day_of_week": "Monday", "post_count": 0, "total_engagement": 0 }],
    "previous": ["...same..."]
  }
}
```

### 6.5 Top Posts

**Route:** `GET /analytics/overview/linkedin/topPosts`

**Payload:** Same as 6.1 + `"limit": 3, "order_by": "total_engagement"`

**Response:**
```json
{
  "top_posts": [
    {
      "post_id": "string", "title": "string",
      "post_likes": 0, "post_comments": 0, "total_engagement": 0,
      "post_shares": 0, "impressions": 0, "engagement_rate": 0.0
    }
  ]
}
```

### 6.6 Get Top Posts (15)

**Route:** `GET /analytics/overview/linkedin/getTopPosts`

**Payload:** Same as 6.1 + `"limit": 15`

**Response:** Same structure as 6.5

### 6.7 Posts Per Day

**Route:** `GET /analytics/overview/linkedin/postsPerDays`

**Payload:** Same as 6.1

**Response:**
```json
{
  "posts_per_days": {
    "data": {
      "days": { "Monday": 0, "Tuesday": 0, "Wednesday": 0, "Thursday": 0, "Friday": 0, "Saturday": 0, "Sunday": 0 },
      "show_data": 0
    }
  }
}
```

### 6.8 Hashtags

**Route:** `GET /analytics/overview/linkedin/hashtags`

**Payload:** Same as 6.1

**Response:**
```json
{
  "top_hashtags": { "#tag": { "count": 0, "engagement": 0 } },
  "top_hashtags_rollup": {
    "current": { "total_hashtags": 0, "total_times_used": 0, "total_engagement": 0, "total_impressions": 0 },
    "previous": { "...same keys..." }
  }
}
```

### 6.9 Followers Demographics

**Route:** `GET /analytics/overview/linkedin/followersDemographics`

**Payload:** Same as 6.1

**Response:**
```json
{
  "follower_demographics": {
    "seniority": { "buckets": ["Director", "Manager"], "values": ["150", "300"] },
    "industry": { "buckets": ["Technology"], "values": ["400"] },
    "country": { "buckets": ["United States"], "values": ["400"] },
    "city": { "buckets": ["San Francisco"], "values": ["100"] }
  }
}
```

---

## 7. YouTube Analytics

### 7.1 Summary

**Route:** `POST /analytics/overview/youtube/overviewSummary`

**Payload:**
```json
{
  "workspace_id": "string (required)",
  "youtube_id": "string",
  "date": "YYYY-MM-DD - YYYY-MM-DD",
  "timezone": "string"
}
```

**Response:**
```json
{
  "overview": {
    "current": {
      "channel_id": "string",
      "watch_time": "0|N/A", "avg_view_duration": "0|N/A",
      "like": "0|N/A", "dislike": "0|N/A", "comment": "0|N/A",
      "share": "0|N/A", "engagement": "0|N/A",
      "subscribers": "0|N/A", "views": "0|N/A", "videos": "0|N/A"
    },
    "previous": { "...same keys..." },
    "difference": { "watch_time": 0, "...": 0 },
    "percentage": { "watch_time": 0.0, "...": 0.0 }
  }
}
```

### 7.2 Subscriber Trend

**Route:** `POST /analytics/overview/youtube/overviewSubscriberTrend`

**Payload:** Same as 7.1

**Response:**
```json
{
  "buckets": ["YYYY-MM-DD"],
  "subscribers_gained_daily": [0],
  "subscribers_total": [0]
}
```

### 7.3 Engagement Trend

**Route:** `POST /analytics/overview/youtube/overviewEngagementTrend`

**Payload:** Same as 7.1

**Response:**
```json
{
  "like_total": [0], "like_daily": [0],
  "dislike_total": [0], "dislike_daily": [0],
  "share_total": [0], "share_daily": [0],
  "comment_total": [0], "comment_daily": [0],
  "engagement_total": [0], "engagement_daily": [0],
  "bucket": ["YYYY-MM-DD"]
}
```

### 7.4 Views Trend

**Route:** `POST /analytics/overview/youtube/overviewViewsTrend`

**Payload:** Same as 7.1

**Response:**
```json
{
  "subscriber_views_total": [0], "subscriber_views_daily": [0],
  "non_subscriber_views_total": [0], "non_subscriber_views_daily": [0],
  "total_views_total": [0], "total_views_daily": [0],
  "buckets": ["YYYY-MM-DD"]
}
```

### 7.5 Watch Time Trend

**Route:** `POST /analytics/overview/youtube/overviewWatchTimeTrend`

**Payload:** Same as 7.1

**Response:**
```json
{
  "subscriber_watch_time_total": [0], "subscriber_watch_time_daily": [0],
  "non_subscriber_watch_time_total": [0], "non_subscriber_watch_time_daily": [0],
  "buckets": ["YYYY-MM-DD"]
}
```

### 7.6 Find Video (Traffic Sources)

**Route:** `POST /analytics/overview/youtube/overviewFindVideo`

**Payload:** Same as 7.1

**Response:**
```json
[
  { "name": "YouTube search", "value": 100, "perc_value": 15.5 },
  { "name": "Suggested videos", "value": 200, "perc_value": 30.0 }
]
```

### 7.7 Video Sharing

**Route:** `POST /analytics/overview/youtube/overviewVideoSharing`

**Payload:** Same as 7.1

**Response:**
```json
[
  { "name": "Facebook", "value": 250, "perc_value": 25.0 },
  { "name": "WhatsApp", "value": 150, "perc_value": 15.0 }
]
```

### 7.8 Top Posts

**Route:** `POST /analytics/overview/youtube/overviewTopPosts`

**Payload:** Same as 7.1 + `"limit": 5`

**Response:**
```json
{
  "top_posts_ordered_by_views": [
    {
      "video_id": "string", "title": "string", "description": "string",
      "duration": 600, "thumbnail_url": "string", "media_type": "video",
      "iframe_embed_url": "string", "share_url": "string",
      "engagement": 0, "like": 0, "dislike": 0, "views": 0, "comment": 0,
      "subscribers_gained": 0, "share": 0, "minutes_watched": 0,
      "average_view_duration": 0.0, "average_view_percentage": 0.0,
      "engagement_rate": 0.0, "published_at": "datetime"
    }
  ],
  "top_posts_ordered_by_engagement": ["...same structure..."]
}
```

### 7.9 Least Posts

**Route:** `POST /analytics/overview/youtube/overviewLeastPosts`

**Payload:** Same as 7.1 + `"limit": 5`

**Response:** Same structure as 7.8 with `least_posts_ordered_by_views` and `least_posts_ordered_by_engagement`

### 7.10 Performance and Posting Schedule

**Route:** `POST /analytics/overview/youtube/overviewPerformanceAndVideoPostingSchedule`

**Payload:** Same as 7.1

**Response:**
```json
{
  "engagement": {
    "count": [0], "likes": [0], "dislikes": [0], "comments": [0],
    "shares": [0], "engagement_total": [0], "hour": ["00:00"]
  },
  "video_views": { "views": [0], "hour": ["00:00"] }
}
```

### 7.11 Sorted Top Posts

**Route:** `POST /analytics/overview/youtube/getSortedTopPosts`

**Payload:** Same as 7.1 + `"limit": 5, "order_by": "engagement"`

**Response:**
```json
{
  "top_posts": ["...same video structure as 7.8..."]
}
```

### 7.12-7.15 Dynamic Trends

**Routes:**
- `POST /analytics/overview/youtube/overviewDynamicSubscriberTrend`
- `POST /analytics/overview/youtube/overviewDynamicEngagementTrend`
- `POST /analytics/overview/youtube/overviewDynamicViewsTrend`
- `POST /analytics/overview/youtube/overviewDynamicWatchTimeTrend`

**Payload:** Same as 7.1

**Response:** Same as their non-dynamic counterparts (7.2-7.5)

---

## 8. TikTok Analytics

### 8.1 Page and Posts Insights

**Route:** `POST /analytics/overview/tiktok/getPageAndPostsInsights`

**Payload:**
```json
{
  "tiktok_id": "string (required)",
  "workspace_id": "string (required)",
  "date": "YYYY-MM-DD - YYYY-MM-DD",
  "timezone": "UTC"
}
```

**Response:**
```json
{
  "data": {
    "tiktok_id": "string", "page_name": "string", "logo": "string",
    "total_likes": 0, "total_likes_growth": 0.0, "total_likes_diff": 0,
    "total_comments": 0, "total_comments_growth": 0.0, "total_comments_diff": 0,
    "total_shares": 0, "total_shares_growth": 0.0, "total_shares_diff": 0,
    "total_engagements": 0, "total_engagements_growth": 0.0, "total_engagements_diff": 0,
    "total_posts": 0, "total_posts_growth": 0.0, "total_posts_diff": 0,
    "total_followers": "0|N/A", "total_followers_growth": "0.0|N/A",
    "total_followings": "0|N/A", "total_video_views": "0|N/A"
  }
}
```

### 8.2 Page Followers and Views

**Route:** `POST /analytics/overview/tiktok/getPageFollowersAndViews`

**Payload:** Same as 8.1

**Response:**
```json
{
  "data": [
    {
      "platform_id": "string", "display_name": "string", "logo": "string",
      "followers_count": [0], "views_per_day": [0],
      "followers_count_diff": [0], "views_per_day_diff": [0],
      "day_bucket": ["datetime"]
    }
  ]
}
```

### 8.3 Posts and Engagements

**Route:** `POST /analytics/overview/tiktok/getPostsAndEngagements`

**Payload:** Same as 8.1

**Response:**
```json
{
  "data": [
    {
      "tiktok_id": "string", "page_name": "string",
      "days_bucket": ["date"],
      "sum_view_count": [0], "sum_like_count": [0], "sum_comments_count": [0],
      "sum_share_count": [0], "sum_engagement_count": [0],
      "avg_engagement_rate": [0.0], "post_count": [0]
    }
  ]
}
```

### 8.4 Daily Engagements

**Route:** `POST /analytics/overview/tiktok/getDailyEngagementsData`

**Payload:** Same as 8.1

**Response:**
```json
{
  "data": [
    {
      "tiktok_id": "string",
      "total_video_likes": [0], "total_video_comments": [0], "total_video_shares": [0],
      "daily_video_likes": [0], "daily_video_comments": [0], "daily_video_shares": [0],
      "total_engagement": [0], "daily_engagement": [0],
      "days_bucket": ["datetime"]
    }
  ]
}
```

### 8.5 Top and Least Performing Posts

**Route:** `POST /analytics/overview/tiktok/getTopAndLeastPerformingPosts`

**Payload:** Same as 8.1

**Response:**
```json
{
  "data": {
    "top_posts": [
      {
        "tiktok_id": "string", "post_id": "string",
        "cover_image_url": "string", "share_url": "string",
        "post_description": "string", "hashtags": "string",
        "embed_html": "string", "embed_link": "string",
        "likes_count": 0, "comments_count": 0, "shares_count": 0,
        "views_count": 0, "engagement_count": 0, "engagement_rate": 0.0,
        "created_time": "datetime"
      }
    ],
    "least_posts": ["...same structure..."]
  }
}
```

### 8.6 Posts Data (Paginated)

**Route:** `POST /analytics/overview/tiktok/getPostsData`

**Payload:** Same as 8.1 + `"limit": 5, "offset": 0, "sort_order": "total_engagement"`

**Response:**
```json
{
  "data": [
    {
      "tiktok_id": "string", "post_id": "string",
      "likes_count": 0, "comments_count": 0, "shares_count": 0,
      "views_count": 0, "total_engagement": 0, "engagement_rate": 0.0,
      "created_time": "datetime", "total": 0
    }
  ]
}
```

---

## 9. Pinterest Analytics

### 9.1 Summary

**Route:** `POST /analytics/overview/pinterest/overviewSummary`

**Payload:**
```json
{
  "workspace_id": "string (required)",
  "pinterest_id": "string (required)",
  "date": "YYYY-MM-DD - YYYY-MM-DD",
  "timezone": "string",
  "board_id": "string (optional)"
}
```

**Response:**
```json
{
  "overview": {
    "current": {
      "follower_count": "0|N/A", "impressions": "0|N/A",
      "engagement": "0|N/A", "pin_count": "0|N/A",
      "pin_clicks": "0|N/A", "saves": "0|N/A", "outbound_clicks": "0|N/A"
    },
    "previous": { "...same keys..." },
    "difference": { "...same keys..." },
    "percentage": { "...same keys..." }
  }
}
```

### 9.2 Followers

**Route:** `POST /analytics/overview/pinterest/overviewFollowers`

**Payload:** Same as 9.1

**Response:**
```json
{ "date_bucket": ["date"], "follower_count": [0], "follower_count_daily": [0] }
```

### 9.3 Impressions

**Route:** `POST /analytics/overview/pinterest/overviewImpressions`

**Payload:** Same as 9.1

**Response:**
```json
{ "date_bucket": ["date"], "impression_count": [0], "impression_count_daily": [0] }
```

### 9.4 Engagement

**Route:** `POST /analytics/overview/pinterest/overviewEngagement`

**Payload:** Same as 9.1

**Response:**
```json
{ "date_bucket": ["date"], "engagement": [0], "engagement_daily": [0] }
```

### 9.5 Pin Posting Per Day

**Route:** `POST /analytics/overview/pinterest/overviewPinPostingPerDay`

**Payload:** Same as 9.1

**Response:**
```json
{ "date_bucket": ["date"], "pin_count": [0] }
```

### 9.6 Pin Posting Rollup

**Route:** `POST /analytics/overview/pinterest/overviewPinPostingRollup`

**Payload:** Same as 9.1

**Response:**
```json
{
  "current": { "total_pins": 0 }, "previous": { "total_pins": 0 },
  "difference": { "total_pins": 0 }, "percentage": { "total_pins": 0.0 }
}
```

### 9.7 Top Pins

**Route:** `POST /analytics/overview/pinterest/overviewTopPins`

**Payload:** Same as 9.1 + `"limit": 5, "order_by": "impressions"`

**Response:**
```json
{
  "top": [
    {
      "pin_id": "string", "board_id": "string", "description": "string",
      "impressions": 0, "pin_clicks": 0, "outbound_clicks": 0,
      "saves": 0, "engagement": 0, "media_url": "string", "created_at": "datetime"
    }
  ]
}
```

### 9.8 Least Pins

**Route:** `POST /analytics/overview/pinterest/overviewLeastPins`

**Payload:** Same as 9.7

**Response:** Same structure as 9.7 with key `"least"` instead of `"top"`

### 9.9 Get Top Posts

**Route:** `POST /analytics/overview/pinterest/getTopPosts`

**Payload:** Same as 9.1

**Response:** Same structure as 9.7

### 9.10 Pin Posting Performance

**Route:** `POST /analytics/overview/pinterest/overviewPinPostingPerformance`

**Payload:** Same as 9.1

**Response:**
```json
{ "date_bucket": ["date"], "metric_data": [0] }
```

---

## 10. Twitter/X Analytics

### 10.1 Page and Posts Insights

**Route:** `POST /analytics/overview/twitter/getPageAndPostsInsights`

**Payload:**
```json
{
  "twitter_id": "string (required)",
  "workspace_id": "string (required)",
  "date": "YYYY-MM-DD - YYYY-MM-DD",
  "timezone": "UTC"
}
```

**Response:**
```json
{
  "data": {
    "twitter_id": "string", "page_name": "string", "logo": "string",
    "followers_count": 0, "followers_count_growth": "0.0|N/A", "followers_count_diff": 0,
    "following_count": 0, "following_count_growth": "0.0|N/A",
    "tweet_count": 0, "listed_count": 0,
    "total_engagement": 0, "total_engagement_growth": "0.0|N/A",
    "reply_count": 0, "retweet_count": 0, "bookmark_count": 0,
    "quote_count": 0, "like_count": 0
  }
}
```

### 10.2 Engagement Impression Data

**Route:** `POST /analytics/overview/twitter/getEngagementImpressionData`

**Payload:** Same as 10.1

**Response:**
```json
{
  "twitter_id": "string", "name": "string",
  "tweet_count": [0], "impression_count": [0], "total_engagement": [0],
  "retweet_count": [0], "reply_count": [0], "like_count": [0],
  "bookmark_count": [0], "quote_count": [0],
  "tweeted_at_date": ["date"]
}
```

### 10.3 Follower Trend

**Route:** `POST /analytics/overview/twitter/getFollowersTrendData`

**Payload:** Same as 10.1

**Response:**
```json
{
  "platform_id": "string", "name": "string",
  "follower_count": [0], "follower_count_daily": [0],
  "following_count": [0], "following_count_daily": [0],
  "buckets": ["date"]
}
```

### 10.4 Top Tweets

**Route:** `POST /analytics/overview/twitter/getTopTweets`

**Payload:** Same as 10.1 + `"order_by": "total_engagement", "limit": 5`

**Response:**
```json
{
  "top_tweets": [
    {
      "id": "string", "tweeted_at": "datetime", "tweet_text": "string",
      "tweet_type": "string", "permalink": "string",
      "retweet_count": 0, "like_count": 0, "reply_count": 0,
      "quote_count": 0, "bookmark_count": 0,
      "impression_count": 0, "total_engagement": 0
    }
  ]
}
```

### 10.5 Least Tweets

**Route:** `POST /analytics/overview/twitter/getLeastTweets`

**Payload:** Same as 10.4

**Response:** Same structure with key `"least_tweets"`

### 10.6 Credits Used Count

**Route:** `POST /analytics/overview/twitter/getCreditsUsedCount`

**Payload:** Same as 10.1

**Response:**
```json
{ "data": { "credits_used": 0 } }
```

---

## 11. Twitter/X Settings

### 11.1 Create Setting

**Route:** `POST /analytics/settings/twitter/createTwitterAnalyticsSetting`

**Payload:**
```json
{
  "platform_id": "string (required)",
  "workspace_id": "string (required)",
  "updated_by": "string (required)",
  "created_by": "string (required)",
  "job_type": "string (required)",
  "trigger_day": "string (required)",
  "platform_name": "string (required)",
  "post_count": 0
}
```

**Response:**
```json
{
  "_id": "ObjectId", "platform_id": "string", "workspace_id": "string",
  "job_type": "string", "trigger_day": "string", "platform_name": "string",
  "post_count": 0, "created_at": "datetime", "updated_at": "datetime"
}
```

### 11.2 Update Setting

**Route:** `POST /analytics/settings/twitter/updateTwitterAnalyticsSetting`

**Payload:** Same as 11.1

**Response:** Same as 11.1

### 11.3 Get Settings

**Route:** `POST /analytics/settings/twitter/getTwitterAnalyticsSetting`

**Payload:** `{ "workspace_id": "string (required)" }`

**Response:** `[ ...array of setting objects same as 11.1 response... ]`

### 11.4 Trigger Job

**Route:** `POST /analytics/settings/twitter/triggerTwitterAnalyticsJob`

**Payload:** `{ "platform_id": "string", "workspace_id": "string" }`

**Response:** `{ "status": true, "message": "string" }`

### 11.5 Create Job Logs

**Route:** `POST /analytics/settings/twitter/createTwitterJobLogs`

**Payload:**
```json
{
  "platform_id": "string", "workspace_id": "string",
  "platform_type": "string", "job_type": "string",
  "credits_used": 0, "executed_by": "string",
  "app_id": "string", "app_name": "string", "error": "string"
}
```

**Response:** `{ "status": true, "message": "string" }`

---

## 12. Cross-Platform Overview V1

### 12.1 Summary

**Route:** `POST /analytics/overview/summary`

**Payload:**
```json
{
  "workspace_id": "string (required)",
  "date": "YYYY-MM-DDTHH:mm:ss - YYYY-MM-DDTHH:mm:ss",
  "timezone": "string",
  "facebook_accounts": [], "instagram_accounts": [],
  "linkedin_accounts": [], "youtube_accounts": [],
  "tiktok_accounts": [], "pinterest_accounts": [],
  "previous_date": "string (optional)"
}
```

**Response:**
```json
{
  "ids": { "facebook": [], "instagram": [], "linkedin": [], "youtube": [], "tiktok": [], "pinterest": [] },
  "overview": {
    "current": { "total_posts": 0, "reposts": 0, "comments": 0, "reactions": 0 },
    "previous": { "...same keys..." }
  }
}
```

### 12.2 Engagement Rollup

**Route:** `POST /analytics/overview/engagementRollup`

**Payload:** Same as 12.1

**Response:**
```json
{
  "engagement_rollup": {
    "current": { "total_engagement": 0, "engagement_per_day": 0.0, "total_posts": 0, "posts_per_day": 0.0 },
    "previous": { "...same keys..." }
  }
}
```

### 12.3 Top Posts

**Route:** `POST /analytics/overview/topPosts`

**Payload:** Same as 12.1

**Response:**
```json
{
  "top_posts": {
    "overall": [{ "post_id": "", "total_engagement": 0, "network": "facebook" }],
    "facebook": [], "instagram": [], "linkedin": [], "youtube": [], "tiktok": []
  }
}
```

### 12.4 Posts Engagement

**Route:** `POST /analytics/overview/postsEngagement`

**Payload:** Same as 12.1

**Response:**
```json
{
  "posts_engagements": {
    "buckets": ["YYYY-MM-DD"],
    "data": { "total_engagement": [0], "comments": [0], "reactions": [0], "reposts": [0], "post_count": [0] },
    "show_data": 0
  }
}
```

### 12.5 Account Performance

**Route:** `POST /analytics/overview/accountPerformance`

**Payload:** Same as 12.1

**Response:**
```json
{
  "account_performance": {
    "facebook": { "comments": 0, "reactions": 0, "reposts": 0, "post_clicks": 0, "total_engagement": 0, "total_posts": 0 },
    "instagram": { "comments": 0, "reactions": 0, "reposts": 0, "saved": 0, "total_engagement": 0, "total_posts": 0 },
    "linkedin": {}, "youtube": {}, "tiktok": {}, "pinterest": {},
    "overall": [{ "platform": "facebook", "platform_id": "string", "...metrics...": 0 }]
  }
}
```

### 12.6 Time Recommendation

**Route:** `POST /analytics/overview/timeRecommendation`

**Payload:** Same as 12.1 + `"state": "merged", "accounts": [{ "facebook_id": "", "instagram_id": "", "linkedin_id": "" }]`

**Response:**
```json
{
  "data": [{
    "facebook": { "page_id": { "0": { "0": 0.0, "1": 0.0 } } },
    "instagram": {}, "linkedin": {},
    "merged": { "0": { "0": 0.0, "1": 0.0 } }
  }]
}
```

---

## 13. Cross-Platform Overview V2

### 13.1 Summary

**Route:** `POST /analytics/overview/getSummary`

**Payload:**
```json
{
  "workspace_id": "string (required)",
  "date": "YYYY-MM-DD - YYYY-MM-DD",
  "timezone": "string",
  "facebook_accounts": [], "instagram_accounts": [],
  "linkedin_accounts": [], "youtube_accounts": [],
  "tiktok_accounts": [], "pinterest_accounts": []
}
```

**Response:**
```json
{
  "summary": {
    "followers": 0, "posts": 0, "engagement": 0,
    "impressions": 0, "reach": 0, "engagement_rate": 0.0,
    "secondary_followers": 0, "secondary_posts": 0,
    "secondary_engagement": 0, "secondary_impressions": 0,
    "secondary_reach": 0, "secondary_engagement_rate": 0.0
  }
}
```

### 13.2 Top Performing Graph

**Route:** `POST /analytics/overview/getTopPerformingGraph`

**Payload:** Same as 13.1

**Response:**
```json
{
  "buckets": ["YYYY-MM-DD"],
  "facebook_post_count": [0], "facebook_engagement_count": [0], "facebook_impression_count": [0], "facebook_reach_count": [0],
  "instagram_post_count": [0], "instagram_engagement_count": [0], "instagram_impression_count": [0], "instagram_reach_count": [0],
  "linkedin_post_count": [0], "linkedin_engagement_count": [0],
  "tiktok_post_count": [0], "tiktok_engagement_count": [0],
  "youtube_post_count": [0], "youtube_engagement_count": [0],
  "pinterest_post_count": [0], "pinterest_engagement_count": [0]
}
```

### 13.3 Platform Data

**Route:** `POST /analytics/overview/getPlatformData`

**Payload:** Same as 13.1

**Response:**
```json
[
  {
    "platform": "facebook", "account_id": "string",
    "total_posts": 0, "total_engagement": 0, "total_impressions": 0
  }
]
```

### 13.4 Platform Data Detailed

**Route:** `POST /analytics/overview/getPlatformDataDetailed`

**Payload:** Same as 13.1

**Response:** Same as 13.3 with additional detail fields

### 13.5 Platform Data Graphs

**Route:** `POST /analytics/overview/getPlatformDataGraphs`

**Payload:** Same as 13.1

**Response:**
```json
[{ "date": "YYYY-MM-DD", "engagement": 0, "impressions": 0, "posts": 0 }]
```

### 13.6 Top Posts

**Route:** `POST /analytics/overview/getTopPosts`

**Payload:** Same as 13.1

**Response:**
```json
[
  {
    "post_id": "string", "platform": "facebook", "account_id": "string",
    "engagement": 0, "impressions": 0, "content": "string", "published_at": "datetime"
  }
]
```

---

## 14. Campaign and Label Analytics

### 14.1 Summary

**Route:** `POST /analytics/campaignLabelAnalytics/getSummaryAnalytics`

**Payload:**
```json
{
  "workspace_id": "string (required)",
  "campaigns": ["string"],
  "labels": ["string"],
  "date": "YYYY-MM-DD - YYYY-MM-DD",
  "facebook_accounts": [], "instagram_accounts": [],
  "linkedin_accounts": [], "youtube_accounts": [],
  "tiktok_accounts": [], "pinterest_accounts": []
}
```

**Response:**
```json
{
  "current": { "total_posts": 0, "total_engagement": 0, "total_impressions": 0, "total_engagement_rate_per_impression": 0.0 },
  "previous": { "...same keys..." },
  "difference": { "...same keys..." },
  "percentage": { "...same keys..." }
}
```

### 14.2 Breakdown Data

**Route:** `POST /analytics/campaignLabelAnalytics/getCampaignLabelBreakdownData`

**Payload:** Same as 14.1

**Response:**
```json
{
  "campaigns": {
    "campaign_id": [
      { "id": "string", "era": "current|previous", "total_posts": 0, "total_engagement": 0, "total_impressions": 0 }
    ]
  },
  "labels": { "...same structure..." }
}
```

### 14.3 Insights Breakdown

**Route:** `POST /analytics/campaignLabelAnalytics/getCampaignLabelInsightsBreakdown`

**Payload:** Same as 14.1

**Response:**
```json
{
  "campaign_id": {
    "total_engagement": [0], "total_impressions": [0],
    "total_posts": [0], "created_at": ["YYYY-MM-DD"]
  }
}
```

### 14.4 Planner Analytics

**Route:** `POST /analytics/campaignLabelAnalytics/getPlannerAnalytics`

**Payload:**
```json
{
  "platforms": "facebook|instagram|linkedin|tiktok|youtube|pinterest",
  "date": "YYYY-MM-DD - YYYY-MM-DD",
  "all_post_ids": ["string"]
}
```

**Response:**
```json
{ "total_posts": 0, "total_engagement": 0, "total_impressions": 0 }
```

---

## 15. Dashboard Analytics

### 15.1 Content Publishing Stats

**Route:** `POST /getContentPublishingStats`

**Payload:**
```json
{ "workspace_id": "string (required)", "date_range": "YYYY-MM-DD - YYYY-MM-DD" }
```

**Response:**
```json
{ "status": true, "stats": { "scheduled": 0, "published": 0, "partial": 0, "failed": 0 } }
```

### 15.2 Approval Publishing Stats

**Route:** `POST /getContentApprovalStats`

**Payload:** Same as 15.1

**Response:**
```json
{ "status": true, "stats": { "review": 0, "rejected": 0, "missed": 0 } }
```

### 15.3 Inbox Stats

**Route:** `POST /getInboxStats`

**Payload:** `{ "workspace_id": "string (required)" }`

**Response:**
```json
{ "status": true, "stats": { "UNASSIGNED": 0, "ASSIGNED": 0, "MARKED_AS_DONE": 0, "MINE": 0 } }
```

---

## 16. Share Link Management

### 16.1 Create Share Link

**Route:** `POST /analytics/share-link/create`

**Payload:**
```json
{
  "title": "string (required, max 255)",
  "workspace_id": "string (required)",
  "platform": "overview|facebook|instagram|linkedin|tiktok|youtube|pinterest|twitter",
  "is_date_range_fixed": false,
  "date_range": "string (required if is_date_range_fixed is true)",
  "is_account_switching_enabled": true,
  "is_password_protected": false,
  "password": "string (required if is_password_protected, min 4)",
  "account_id": "string (optional)",
  "overview_accounts": []
}
```

**Response:**
```json
{ "status": "success", "data": { "share_url": "https://app.example.com/share/analytics/{link_id}" } }
```

### 16.2 Update Share Link

**Route:** `POST /analytics/share-link/update/{id}`

**Payload:**
```json
{
  "title": "string", "is_date_range_fixed": false, "date_range": "string",
  "is_account_switching_enabled": true, "is_password_protected": false,
  "password": "string", "account_id": "string"
}
```

**Response:** `{ "status": "success", "message": "Share link updated successfully" }`

### 16.3 Toggle State

**Route:** `PUT /analytics/share-link/update/toggle-state/{id}`

**Payload:** `{ "is_disabled": true }`

**Response:** `{ "status": "success", "message": "Share link state toggled successfully" }`

### 16.4 List Share Links

**Route:** `GET /analytics/share-link/list/{workspace_id}`

**Response:**
```json
{
  "data": [
    {
      "link_id": "string", "title": "string", "platform": "string",
      "is_disabled": false, "is_password_protected": false,
      "user": { "firstname": "string", "lastname": "string" },
      "created_at": "datetime"
    }
  ]
}
```

### 16.5 Delete Share Link

**Route:** `DELETE /analytics/share-link/delete/{id}`

**Response:** `{ "status": "success", "data": "{id}", "message": "Share link deleted" }`

### 16.6 Get Details (Public)

**Route:** `GET /analytics/shared/{link_id}`

**Response:** Full share link document (all fields from 16.1 payload + metadata)

### 16.7 Verify Password

**Route:** `POST /analytics/shared/{link_id}/verify-password`

**Payload:** `{ "password": "string (required)" }`

**Response:** `{ "status": "success", "message": "Share link password verified" }`

---

## 17. Reports

### 17.1 Save/Render/Email Report

**Route:** `POST /analytics/reports/save`

**Payload:**
```json
{
  "type": "single-pdf-detailed|single-pdf-overview|multiple-pdf-overview|multiple-pdf-detailed|group|competitor",
  "action": "save|render|email",
  "workspace_id": "string (required)",
  "name": "string", "platform_type": "string",
  "accounts": [], "date": "YYYY-MM-DD - YYYY-MM-DD",
  "email_list": ["email@example.com"],
  "language": "en", "labels": [], "campaigns": [], "topPosts": 5
}
```

**Response (save):**
```json
{
  "status": true,
  "report": { "_id": "string", "workspace_id": "string", "type": "string", "progress": 0, "created_at": "datetime" }
}
```

**Response (render/email):** `{ "status": true, "message": "Report generation started" }`

### 17.2 Show Report

**Route:** `POST /analytics/reports/show`

**Payload:** `{ "_id": "string" }`

**Response:**
```json
{
  "status": true,
  "report": { "_id": "string", "status": "completed", "progress": 100, "export_url": "https://..." }
}
```

### 17.3 List Reports

**Route:** `POST /analytics/reports/list`

**Payload:** `{ "workspace_id": "string" }`

**Response:** `{ "status": true, "data": [ ...array of report objects... ] }`

### 17.4 Remove Report

**Route:** `POST /analytics/reports/remove`

**Payload:** `{ "report_id": "string" }`

**Response:** `{ "status": true }`

---

## 18. Scheduled Reports

### 18.1 List Schedules

**Route:** `POST /analytics/reports/schedule/show`

**Payload:** `{ "workspace_id": "string" }`

**Response:** `{ "status": true, "data": [ ...array of schedule objects... ] }`

### 18.2 Create/Update Schedule

**Route:** `POST /analytics/reports/schedule/save`

**Payload:**
```json
{
  "_id": "string (optional, for updates)",
  "workspace_id": "string", "name": "string",
  "type": "group|individual", "platform_type": "string",
  "accounts": [], "email_list": ["email@example.com"],
  "frequency": "daily|weekly|monthly|custom",
  "day_of_week": 0, "day_of_month": 1,
  "time": "09:00", "timezone": "UTC"
}
```

**Response:**
```json
{ "status": true, "report": { "_id": "string", "...fields..." }, "message": "Report has been scheduled" }
```

### 18.3 Remove Schedule

**Route:** `POST /analytics/reports/schedule/remove`

**Payload:** `{ "report_id": "string" }`

**Response:** `{ "status": true }`

### 18.4 Send (Cron)

**Route:** `POST /analytics/reports/schedule/send`

**Payload:** `{ "interval": "daily|weekly|monthly" }`

**Response:** No explicit response (internal cron endpoint)

---

## 19. Analytics Job Triggers

### 19.1 Trigger Analytics Job

**Route:** `POST /analytics/triggerJob`

**Payload:**
```json
{ "workspace_id": "string", "account_id": "string", "platform": "string" }
```

**Response:** `{ "status": true, "message": "string" }`

### 19.2 Trigger Competitor Job

**Route:** `POST /analytics/triggerCompetitorJob`

**Payload:**
```json
{ "workspace_id": "string", "report_id": "string", "platform": "string", "competitor_ids": ["string"] }
```

**Response:** `{ "status": true, "message": "string" }`

---

## 20. AI Insights

All AI insight endpoints follow the same pattern per platform.

### 20.1 Common Pattern

**Payload:**
```json
{
  "workspace_id": "string (required)",
  "date": "YYYY-MM-DD - YYYY-MM-DD",
  "{platform}_id": "string (required)",
  "type": "string (required - insight type)",
  "timezone": "string",
  "limit": 0
}
```

**Response (success):**
```json
{ "success": true, "data": { "...AI generated insight content..." } }
```

**Response (insufficient data):**
```json
{ "success": false, "message": "analytics.insufficient_data" }
```

### 20.2 Platform Routes and Types

| Route | Platform | Available Types |
|-------|----------|----------------|
| `GET /analytics/overview/facebook/ai_insights` | Facebook | `page_impressions`, `page_engagement`, `publishing_behaviour_impressions`, `publishing_behaviour_engagements`, `publishing_behaviour_reach`, `audience_growth`, `video_views`, `video_watch_time`, `video_engagements`, `reels_initial_plays`, `reels_watch_time`, `reels_engagement`, `top_posts`, `insights_summary` |
| `POST /analytics/overview/instagram/ai_insights` | Instagram | `impressions`, `engagement`, `publishing_behaviour_impressions`, `publishing_behaviour_engagements`, `publishing_behaviour_reach`, `audience_growth`, `reels_engagement`, `reels_watch_time`, `reels_shares`, `stories_interactions`, `stories_impressions`, `stories_reach`, `top_posts`, `top_hashtags`, `insights_summary` |
| `GET /analytics/overview/linkedin/ai_insights` | LinkedIn | `publishing_behaviour`, `publishing_behaviour_impressions`, `publishing_behaviour_reach`, `audience_growth`, `page_views`, `top_posts`, `top_hashtags`, `city_demographics`, `country_demographics`, `industry_demographics`, `post_density`, `seniority_demographics`, `insights_summary` |
| `POST /analytics/overview/youtube/ai_insights` | YouTube | `subscribers_trend`, `daily_views`, `daily_engagement`, `daily_watch_time`, `viewers_find_videos`, `engagement_vs_posting_pattern`, `sharing_services`, `top_and_least_posts`, `insights_summary` |
| `POST /analytics/overview/tiktok/ai_insights` | TikTok | `audience_growth`, `top_posts`, `daily_engagement`, `cumulative_engagement`, `daily_video_views`, `cumulative_video_views`, `engagement_vs_daily_posting`, `insights_summary` |
| `POST /analytics/overview/pinterest/ai_insights` | Pinterest | `daily_engagement`, `impressions_vs_posting_pattern`, `engagement_vs_posting_pattern`, `daily_pin_posting`, `daily_followers_trend`, `cumulative_followers_trend`, `impressions`, `engagement`, `top_and_least_posts`, `insights_summary` |
| `POST /analytics/overview/ai_insights` | Overview | `reach_across_platforms`, `engagement_across_platforms`, `impressions_across_platforms`, `platform_performance_comparison`, `overview_account_statistics`, `top_posts` |

**Cache:** All AI insights are cached for 24 hours. Cache key format: `{platform}_AI:{method}:{account_id}:{date}:{locale}`
