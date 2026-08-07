# Research — Top Posts sort dropdown + Least Posts consistency across platform analytics

Make the Overview "Top Posts" section consistent across platform analytics reports: add the **sort dropdown** (heading becomes a "Top posts by …" dropdown) and the **Least Posts** section to the platforms that lack them, matching Pinterest/YouTube.

## Current State (verified — refines the original ask)

| Platform | Top Posts sort dropdown | Least Posts section |
|---|---|---|
| Pinterest | ✅ has it | ✅ has it |
| YouTube | ✅ has it | ✅ has it |
| **TikTok** | ❌ **missing** | ✅ **already has it** |
| **Facebook** | ❌ missing | ❌ missing |
| **Instagram** | ❌ missing | ❌ missing |
| **LinkedIn** | ❌ missing | ❌ missing |

> Note: TikTok already shows Least Posts — it only needs the **dropdown**. Facebook / Instagram / LinkedIn need **both**.

**Reference implementation (the consistency target):** `views/pinterest/components/TopAndLeastEngagingPosts.vue` (and the YouTube equivalent). It uses the shared `views/common/AnalyticsDropdown.vue` + `AnalyticsDropdownItem.vue`; the selected sort type (`selectedTopLeastSortType` / `TopLeastEngagementDropdown` from the platform composable) drives **server-side** queries for both top and least — `usePinterestOverviewSectionQuery` with route keys `TOP_PERFORMING_PINS` and `LEAST_PERFORMING_PINS`. So one dropdown controls **both** the Top and Least lists (top = highest by the chosen metric, least = lowest), each showing 5 with a "show more" modal.

**Facebook / Instagram / LinkedIn today:** only `views/<platform>/components/TopPosts.vue`. Facebook reads `overviewTopPostsData` (route `OVERVIEW_TOP_POSTS` → the response's `top_posts`), shows top 5 + a `TopPostsModal`. **No sort dropdown, no metric param, and no least-performing list.** Instagram and LinkedIn mirror this structure.

**TikTok today:** `views/tiktok/components/TopAndLeastEngagingPosts.vue` renders top **and** least, but **without** the `AnalyticsDropdown` — so it needs only the dropdown added.

## What Needs to Change
1. **Add the sort dropdown** to the Top Posts heading for **Facebook, Instagram, LinkedIn, TikTok** — the heading becomes a "Top posts by …" dropdown (mirroring Pinterest/YouTube). Per-platform options:

   | Platform | Dropdown options |
   |---|---|
   | Facebook | Engagement, Impressions, Reach |
   | Instagram | Engagement, Views, Reach |
   | LinkedIn | Engagement, Impressions, Reach |
   | TikTok | Engagement, Views |

2. **Add the Least Posts section** to **Facebook, Instagram, LinkedIn** (TikTok already has it), in the Overview — same 5-posts + show-more pattern as the reference.
3. The dropdown controls **both** Top and Least (pick a metric → top 5 highest + least 5 lowest by that metric), matching Pinterest/YouTube.

## Backend is the **Go analytics service** (`contentstudio-social-analytics-go`) — verified in-code
Analytics is served by the Go service (frontend calls `analyticsGoBaseUrl` / `VITE_ANALYTICS_GO_URL`), **not** the Laravel `contentstudio-backend`. Per-platform HTTP handlers live in `src/api/analytics/<platform>/handler.go`. What I found there sharpens the scope considerably:

| Platform | Top-posts `order_by` (metric sort) | Least-performing endpoint | Go work needed |
|---|---|---|---|
| Pinterest / YouTube | ✅ | ✅ | — (reference) |
| **TikTok** | ✅ combined `HandleTopAndLeastPerformingPosts` (`sort_order`) returns **both top + least** | ✅ (same endpoint) | **None — FE-only** |
| **Facebook** | ✅ `HandleOverviewTopPosts` already parses `order_by` | ❌ **no least endpoint** | Add least endpoint |
| **Instagram** | ✅ `HandleTopPosts` sets `topReq.OrderBy` | ❌ **no least endpoint** | Add least endpoint |
| **LinkedIn** | ✅ `handleTopPostsWithDefault` ("accepts order_by") | ❌ **no least endpoint** | Add least endpoint |

So the Go work is **narrower than first thought**:
- **Metric sorting of the top list already exists** for FB/IG/LinkedIn via `order_by` — the FE dropdown just needs to pass it. *(Confirm the accepted `order_by` values cover the requested metrics — engagement/impressions/reach for FB & LinkedIn, engagement/views/reach for IG; extend if a value is missing.)*
- **Least-performing endpoints must be added** for FB/IG/LinkedIn (mirroring Pinterest's `HandleTopPins` order_by+limit and YouTube's dual top/least `parseTopAndLeastVideosRequest`). Least lists exist today only for Pinterest/YouTube/TikTok/competitors.
- **TikTok needs no backend work** — `HandleTopAndLeastPerformingPosts` already returns top **and** least with a sort param; TikTok is purely the FE dropdown.

This is **Full Stack** (Go least-endpoints for FB/IG/LinkedIn + Vue frontend for all four dropdowns and the FB/IG/LinkedIn least sections).

## Mobile Context
Web only — analytics reports are not in the mobile apps.

## Files Involved
- **Reference to copy:** `views/pinterest/components/TopAndLeastEngagingPosts.vue`, `views/youtube/components/TopAndLeastEngagingPosts.vue`; shared `views/common/AnalyticsDropdown.vue` + `AnalyticsDropdownItem.vue`.
- **FE targets:** `views/facebook/components/TopPosts.vue` (+ `useFacebookAnalytics.ts`), `views/instagram/components/TopPosts.vue`, `views/linkedin/components/TopPosts.vue`, `views/tiktok/components/TopAndLeastEngagingPosts.vue`; each platform's Overview wiring and `use<Platform>Analytics` composable (`TopLeastEngagementDropdown` / `selectedTopLeastSortType` scaffold).
- **Go analytics service (`contentstudio-social-analytics-go`):** `src/api/analytics/{facebook,instagram,linkedin}/handler.go` (+ their `service`/types) — **add a least-performing posts endpoint** for each, modeled on `pinterest/handler.go` `HandleTopPins` and `youtube` `parseTopAndLeastVideosRequest`. Top-posts `order_by` already exists on all three (confirm it accepts the requested metrics). **TikTok (`tiktok/handler.go` `HandleTopAndLeastPerformingPosts`) needs no change** — it already returns top+least with a sort param. This is Go, not the Laravel `contentstudio-backend`.
- i18n: `analytics.<platform>.*` for the dropdown option labels + the Least Posts heading (all 8 locale dirs).

## Assumptions to confirm (mirroring Pinterest/YouTube)
- The dropdown drives **both** Top and Least lists.
- **Default sort = Engagement** for every platform.
- Least Posts uses the **same 5-posts + "show more" modal** pattern as Top Posts.
- One **[Full Stack]** story (per your "one story" ask), covering all four platforms.
