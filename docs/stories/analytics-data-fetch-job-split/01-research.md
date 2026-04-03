# Research: Analytics Data Fetch Job Split (Daily + Bi-weekly)

## Current State

The analytics pipeline has a scheduled job (`src/cmd/jobs/main.go`) that runs for each platform (Facebook, Instagram, LinkedIn, TikTok). Each run sends all connected accounts through the data fetch pipeline.

The job already supports a `syncType` parameter (`incremental` vs `full_sync`), but currently it always fetches **2 weeks of historical data** regardless of how frequently it runs. The job is triggered daily for all platforms.

**Key file:** `contentstudio-social-analytics-go/src/cmd/jobs/main.go`

## What Needs to Change

- **Daily job:** Instead of fetching 2 weeks of data, it should only fetch **today's data** (current day only)
- **Bi-weekly job:** A new separate schedule that runs every 2 weeks and fetches the full 2-week window of data for all platforms

All platforms are affected: Facebook, Instagram, LinkedIn, TikTok.

## Files Involved

- `contentstudio-social-analytics-go/src/cmd/jobs/main.go` — job entry point with `syncType` flag
- `contentstudio-social-analytics-go/src/cmd/jobs/fetcher/facebook.go`
- `contentstudio-social-analytics-go/src/cmd/jobs/fetcher/instagram.go`
- `contentstudio-social-analytics-go/src/cmd/jobs/fetcher/linkedin.go`
- Cron/scheduling configuration (infra-level)
