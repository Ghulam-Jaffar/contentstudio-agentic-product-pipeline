# Research — Best Time to Post in the public API

Item 13 of the 7 Aug 2026 backlog batch.

## Current state

Best Time to Post exists and works, but only for the in-app Composer. Nothing about it is reachable from the public API.

**How it is computed.** Per-platform analytics builders hold the query:

- `contentstudio-backend/app/Builders/Analytics/Analyze/FacebookBuilder.php` — `getTimeRecommendationQuery($period)`, documented as "Get the data for best time to post for the given page"
- `contentstudio-backend/app/Builders/Analytics/Analyze/InstagramBuilder.php` — same
- `contentstudio-backend/app/Builders/Analytics/Analyze/LinkedinBuilder.php` — same

Those are the only three builders carrying best-time queries. The controllers that call them:

- `contentstudio-backend/app/Http/Controllers/Analytics/Analyze/FacebookController.php` — `getTimeRecommendation($period)` runs the builder query against the analytics ClickHouse cluster and returns the rows.
- `contentstudio-backend/app/Http/Controllers/Analytics/Analyze/OverviewController.php` — the cross-platform path. Has per-platform distribution methods for Facebook, Instagram and LinkedIn, then `mergeTimeResponseData()` reshapes the hour-by-day matrix, applies the workspace timezone offset, and derives "top 3 best times to post" plus a flag indicating whether there was enough data to answer.

So the shape available today is: a 24-hour by day-of-week score matrix, a derived top-three list, and a has-enough-data flag, computed per account, in the workspace timezone, over a period.

**Who consumes it.** Only the Composer scheduling UI:

- `contentstudio-frontend/src/modules/composer/composables/useTimeRecommendation.ts`
- `contentstudio-frontend/src/modules/publish/components/posting/BestTimeToPost.vue`, `TimeRecommendation.vue`
- `contentstudio-frontend/src/modules/composer/features/scheduling/posting-schedule/components/TimeRecommendationModal.vue`, `BestTimeToPostRow.vue`
- `contentstudio-frontend/src/modules/composer/features/scheduling/PostSchedulerModule/modal/useApplyBestTimes.ts`

**Public API surface today.** `contentstudio-backend/routes/api/v1.php` exposes analytics per network under `workspaces/{workspace_id}/analytics/{network}` for Facebook, Instagram, YouTube, LinkedIn, TikTok, Pinterest, Twitter, GMB, plus campaigns-labels. There is no best-time route in any of those groups, and no `best-times` route anywhere in the API v1 file. Grepping the routes directory for "best" returns nothing.

The public API pattern for analytics is already established by the Facebook analytics controller: `Api/V1/Analytics/FacebookAnalyticsController.php` with per-section endpoints, `ApiKeyHeader` security, OpenAPI annotations, and shared error schemas in `AnalyticsApiSchemas.php`.

## The gap

A customer using the public API can create and schedule posts, and can read analytics, but cannot ask ContentStudio "when should I post on this account?" They have to reimplement the recommendation themselves from raw analytics, which is both wasted effort and a different answer than the one the app shows.

## What needs to change

- A public endpoint returning Best Time to Post per connected social account, authenticated by API key, consistent with the existing public analytics endpoints.
- Decide and document the response structure. The internal shape (hour-by-day matrix, top three, has-enough-data flag) is a reasonable basis, but the public contract should be explicit about the score's meaning rather than exposing an internal metric name.
- Decide and document supported networks. Best-time queries exist for Facebook, Instagram and LinkedIn only. Networks without a builder query need either a clear "not supported for this network" response or exclusion from the endpoint, not a silently empty result.
- Parameters need settling: account, period or date range, and timezone. The internal path already applies a workspace timezone offset in `mergeTimeResponseData()`, so timezone handling has to be explicit in the public contract.
- Then propagate to every developer surface: CLI plus agent skills, MCP, Claude extension, GPT app, Zapier, Make.com, n8n, ClawHub.

## Open questions for engineering

- Is the recommendation per account, or also available aggregated across a set of accounts? `OverviewController` computes a cross-platform merged view, so both are technically available.
- Does the public contract expose the raw score, a normalized 0 to 100 confidence, or only the ranked slots? Exposing an internal ClickHouse score directly makes it hard to change later.
- What is the minimum data threshold, and what does the API return below it? The internal path has a has-enough-data flag; the public API should return something explicit rather than an empty array.

## Files involved

- `contentstudio-backend/routes/api/v1.php` — new route group
- `contentstudio-backend/app/Http/Controllers/Api/V1/Analytics/` — new controller alongside the per-network analytics controllers
- `contentstudio-backend/app/Http/Controllers/Api/V1/Analytics/AnalyticsApiSchemas.php` — reuse the shared error schemas
- `contentstudio-backend/app/Builders/Analytics/Analyze/{Facebook,Instagram,Linkedin}Builder.php` — the existing queries
- `contentstudio-backend/app/Http/Controllers/Analytics/Analyze/OverviewController.php` — the existing reshaping and top-three derivation
- `contentstudio-backend/app/Http/Controllers/Api/V1/SwaggerController.php` — docs generation

External surfaces: `contentstudio-cli`, `contentstudio-mcp`, the Claude extension bundle, the GPT app action schema, the Zapier / Make / n8n apps, and the ClawHub listing.

## Mobile

None. This is a developer-facing API surface. If the recommendation is wanted inside the mobile app, that is a separate `[Flutter]` story.
