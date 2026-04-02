# Research: Analytics Post Thumbnail Refresh — All Platforms

## Current State

A background job (`url-refresher`) exists in `contentstudio-social-analytics-go` that periodically re-fetches post thumbnails for Facebook posts in analytics. It finds posts older than 20 days whose thumbnail URLs may have expired, calls the Facebook Graph API to get fresh URLs, and updates the stored records in ClickHouse.

- **Job location:** `contentstudio-social-analytics-go/src/cmd/jobs/url-refresher/`
- **How it works (Facebook):**
  1. Fetches all valid Facebook accounts from MongoDB
  2. Loads posts older than 20 days from ClickHouse that need a thumbnail refresh
  3. Calls the Facebook Graph API to get updated `full_picture` / attachment URLs
  4. Writes the refreshed thumbnail URLs back to ClickHouse

## What Needs to Change

- The thumbnail refresh job is Facebook-only. It needs to be extended to all other social platforms that are part of the analytics pipeline.

**Platforms supported in analytics (ClickHouse schemas confirmed):**
- Instagram
- LinkedIn
- Pinterest
- TikTok
- Twitter
- YouTube

Each platform stores post thumbnail/image URLs in ClickHouse (confirmed via schema files in `src/deployments/clickhouse/schema/`). These URLs expire over time and need periodic refreshing from each platform's API, the same way Facebook does it.

## Files Involved

- `contentstudio-social-analytics-go/src/cmd/jobs/url-refresher/main.go` — current FB-only job entry point
- `contentstudio-social-analytics-go/src/cmd/jobs/url-refresher/helper.go` — FB worker pool + account processing logic
- `contentstudio-social-analytics-go/src/deployments/clickhouse/schema/instagram_schema.sql`
- `contentstudio-social-analytics-go/src/deployments/clickhouse/schema/linkedin_schema.sql`
- `contentstudio-social-analytics-go/src/deployments/clickhouse/schema/pinterest_schema.sql`
- `contentstudio-social-analytics-go/src/deployments/clickhouse/schema/tiktok_schema.sql`
- `contentstudio-social-analytics-go/src/deployments/clickhouse/schema/twitter_schema.sql`
- `contentstudio-social-analytics-go/src/deployments/clickhouse/schema/youtube_schema.sql`
