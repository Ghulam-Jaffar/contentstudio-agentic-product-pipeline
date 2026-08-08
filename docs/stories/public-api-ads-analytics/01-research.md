# Research — Ads Analytics in the public API

Item 14 of the 7 Aug 2026 backlog batch.

## Current state

Meta Ads and Google Ads analytics are fully built in the product and entirely absent from the public API.

### Where the data is served from

Both are served by the Go analytics service, not by the Laravel backend. Routes are registered in `contentstudio-social-analytics-go/src/api/analytics/router.go` under `/analytics/overview/{meta-ads,google-ads}/`.

**Meta Ads** (`src/api/analytics/meta_ads/handler.go`), 15 sections:

`summary`, `resultsByObjective`, `impressionsVsSpend`, `clicksVsCtr`, `topCampaigns`, `performanceTrend`, `performanceByLevel`, `performanceByPlatform`, `campaignsList`, `adSetsList`, `adsList`, `demographicsAgeGender`, `demographicsRegionCountry`, `aiInsightsSummary`, `aiInsightsDetailed`

**Google Ads** (`src/api/analytics/google_ads/handler.go`), 21 sections:

`summary`, `conversionsByAction`, `impressionsVsSpend`, `clicksVsCtr`, `topCampaigns`, `performanceTrend`, `performanceByLevel`, `performanceByType`, `campaignsList`, `adGroupsList`, `adsList`, `keywordsList`, `searchTermsList`, `shoppingList`, `conversionFunnel`, `conversionsOverTime`, `conversionActionsList`, `campaignConversionActions`, `demographics`, `aiInsightsSummary`, `aiInsightsDetailed`

Request parsing is centralised per platform (`parseRequest` returning a `MetaAdsRequest` / equivalent), and every handler goes through a shared `handle` wrapper, so parameters are uniform across sections within a platform.

Storage is ClickHouse, with per-platform ingestion services (`src/services/meta-ads/meta-ads-sink`, `meta-ads-immediate-processor`, and the Google Ads equivalents).

### Account connection and reporting

- Connectors: `contentstudio-backend/app/Strategy/Integrations/MetaAdsConnector.php`, `GoogleAdsConnector.php`
- Ad accounts are stored as `ExternalLinkIntegration` records, and appear in `WorkspaceMemberPermissionsData` and `SocialAccountAccessRequest`, so per-member ad-account access control already exists.
- PDF reporting already treats them as first-class report types: `contentstudio-backend/app/Services/Analytics/GoReportsClient.php` lists `'meta_ads', 'google_ads'` among its supported report types.
- Sample-workspace seeders exist for both (`SampleMetaAdsClickhouseSeeder.php`, `SampleGoogleAdsClickhouseSeeder.php`), which is useful for API examples and for testing the public endpoints without a live ad account.

### Frontend consumers

- `contentstudio-frontend/src/modules/analytics/views/meta_ads/` — Overview, Campaigns, Ad Sets, Ads, Performance, Demographics tabs, plus Top Campaigns, Results by Objective, Impressions vs Spend and Clicks vs CTR
- `contentstudio-frontend/src/modules/analytics/views/google_ads/` — Overview, Campaigns, Ad Groups, Ads, Keywords, Conversions, Demographics tabs, plus Conversions by Action
- Report views: `views/reports/MetaAdsReport.vue`, `views/reports/GoogleAdsReport.vue`
- Both are gated by feature keys (`meta_ads_analytics`, and a note in `useAnalyticsRoutes.ts` that Google Ads is not currently flag-gated)

### Public API surface today

`contentstudio-backend/routes/api/v1.php` has analytics route groups for Facebook, Instagram, YouTube, LinkedIn, TikTok, Pinterest, Twitter, GMB and campaigns-labels. All organic. There is no ads group of any kind.

The reference implementation for a public analytics endpoint is `contentstudio-backend/app/Http/Controllers/Api/V1/Analytics/FacebookAnalyticsController.php`: per-section endpoints, `ApiKeyHeader` security, `@OA` annotations, shared error schemas in `AnalyticsApiSchemas.php`, and a proxy to the Go analytics service.

Related in-flight work: the `public-analytics-api-rollout` stories add public organic analytics for the remaining networks by mirroring Facebook. Paid analytics is a separate axis and is not covered there.

## The gap

Customers with paid campaigns can see their Meta Ads and Google Ads performance in ContentStudio but cannot pull it programmatically. Agencies in particular want spend, campaign and conversion data alongside organic in their own dashboards and client reports, and today they have to either export PDFs or go back to the ad platforms directly, which defeats the point of consolidating in ContentStudio.

## What needs to change

- Public API endpoints for Meta Ads and Google Ads analytics, mirroring the organic analytics pattern: API key auth, workspace-scoped path, shared error schemas, documented in the reference.
- Expose the connected ad accounts so a caller can discover what to query before querying it. Organic callers already list social accounts; there is no equivalent listing for ad accounts.
- Decide the section granularity for the public API. The internal surface has 15 Meta and 21 Google sections, several of which are chart-shaped rather than data-shaped (`impressionsVsSpend`, `clicksVsCtr`). A public API probably wants fewer, better-named resources rather than a one-to-one mirror of internal chart endpoints.
- Decide whether AI insights sections are in the public contract at all. They are generated narrative, not metrics, and exposing them publicly is a product decision rather than a technical one.
- Preserve the existing per-member ad-account access control when answering API-key requests.
- Currency needs an explicit contract: spend is money, ad accounts have their own currency, and a public API must state which currency a figure is in.
- Then propagate to every developer surface: CLI plus agent skills, MCP, Claude extension, GPT app, Zapier, Make.com, n8n, ClawHub.

## Open questions for engineering and product

- **Section shape.** Mirror all 36 internal sections, or define a smaller public resource model (accounts, campaigns, ad sets or ad groups, ads, keywords, conversions, demographics, summary) with metrics and breakdowns as parameters? The second is a better public API and more work.
- **AI insights in or out.** Recommend out for v1.
- **Filters.** The internal handlers accept whatever `parseRequest` supports. The public contract needs an explicit, documented filter set (date range, campaign status, objective, level) rather than inheriting an internal one.
- **Pagination.** The list endpoints (`campaignsList`, `adsList`, `keywordsList`) are already paginated internally, per a note in `usePDFReports.ts` about paginated campaign rows. The public contract should use the same pagination style as the rest of the public API.
- **Credit accounting.** Ads endpoints can return large result sets. Confirm how they count against the per-plan API credit quota.

## Files involved

Backend:
- `contentstudio-backend/routes/api/v1.php` — new route groups
- `contentstudio-backend/app/Http/Controllers/Api/V1/Analytics/` — new controllers alongside the organic ones
- `contentstudio-backend/app/Http/Controllers/Api/V1/Analytics/AnalyticsApiSchemas.php` — reuse shared error schemas
- `contentstudio-backend/app/Http/Controllers/Api/V1/SwaggerController.php` — docs generation
- `contentstudio-backend/app/Models/Integrations/Platforms/ExternalLinkIntegration.php` — ad account records
- `contentstudio-backend/app/Data/Workspace/WorkspaceMemberPermissionsData.php` — ad-account access control

Go analytics service:
- `contentstudio-social-analytics-go/src/api/analytics/router.go`
- `contentstudio-social-analytics-go/src/api/analytics/meta_ads/handler.go`
- `contentstudio-social-analytics-go/src/api/analytics/google_ads/handler.go`

External surfaces: `contentstudio-cli`, `contentstudio-mcp`, the Claude extension bundle, the GPT app action schema, the Zapier / Make / n8n apps, the ClawHub listing.

## Mobile

None. Ads analytics is web only and this is a developer-facing surface.
