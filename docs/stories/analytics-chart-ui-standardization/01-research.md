# Research — Standardize chart UI across Analytics

Item 19 of the 7 Aug 2026 backlog batch.

## Current state

`contentstudio-frontend/src/modules/analytics/` contains **36 chart components** spread across per-network views, ads views, competitor analytics and overview.

### What is already shared

- **`composables/useEcharts.ts`** — owns the ECharts instance. Lazy-loads `echarts/core`, charts, components, renderers and features; creates and binds the instance to a DOM ref; registers only the chart types a caller asks for; handles resize via a stored handler plus a `ResizeObserver`; cleans both up on unmount. Its public surface is small: `{ chartRef, setup(options, notMerge) }`.
- **`views/common/utils/echartsTooltip.ts`** — `buildAxisDateTooltip()`, a shared builder for `trigger: 'axis'` tooltips that render a workspace-timezone date header and one row per series. Its own docblock states the intent: "Charts only differ in how they format each series value (currency, %, abbreviated counts), so the date/HTML scaffolding lives here and the caller passes a `formatValue` callback."

So the instance lifecycle is shared, and one tooltip shape is shared.

### What is not shared

`useEcharts` takes an `EChartsOption` and hands it to ECharts. It does not contribute any of it. Everything visual is therefore built per component or per composable:

- Legend presence, position, orientation and item styling
- Axis labels, rotation, formatting, tick density
- Grid padding, so charts sit at different insets from their cards
- Series colors, with per-view color modules (`views/meta_ads/composables/metaAdsChartColors.ts`) rather than one palette
- Empty state when a series has no data
- Loading state while the query is pending
- Error state when a query fails
- Value formatting, abbreviation and currency handling

Option construction is scattered across at least:

`composables/analyticsDataHelpers.ts`, `components/common/composables/useAnalytics.ts`, `components/common/helper.ts`, `views/google_ads/composables/useGoogleAdsPerformanceCharts.ts`, `views/meta_ads/composables/useMetaAdsPerformanceCharts.ts`, `views/youtube/composables/useYoutubeAnalytics.ts`, `views/twitter/composables/useTwitterAnalytics.ts`, `views/overview/composables/useOverviewAnalytics.ts`, `views/pinterest/composables/usePinterestAnalytics.ts`, `views/tiktok/composables/useTiktokAnalytics.ts`, and per-view `types/` and `composables/` files.

### Chart component inventory by area

- **Competitor:** `SpecificPostTypeChart`, `FollowersComparisonBarChart`, `DoubleBarChart`, `PostEngagementChart`, `PostReactsDistributionChart`
- **Ads:** `ClicksVsCTRChart` and `ImpressionsVsSpendChart`, duplicated once for Meta and once for Google
- **Per network:** LinkedIn (`AudienceGrowthChart`, `PageViewsChart`, `HashtagsChart`, `PublishingBehaviourChart`), Instagram (`AudienceLocationChart`, `ReelsPerformanceChart`, `AudienceGrowthChart`, `StoriesPerformanceChart`, `HashtagsChart`, `PublishingBehaviourChart`), YouTube (`DeviceTypeChart`, `AgeGenderChart`, `TopGeographiesChart`, `DualChartComponent`), GMB and Pinterest each with their own `DualChartComponent`, Facebook (`VideoPerformanceChart`, `PublishingBehaviourChart`), and others
- Note the same chart name recurs across networks (`AudienceGrowthChart`, `HashtagsChart`, `PublishingBehaviourChart`, `DualChartComponent`) as separate per-network files, which is where much of the drift comes from: a fix in Instagram's `HashtagsChart` does not reach LinkedIn's.

### Related hardcoded-color finding

Report section headers use fixed gradients rather than theme tokens, for example `views/reports/OverviewReport.vue` uses `bg-linear-to-r from-[#01AAFF] to-[#0ACADD]` on every section heading. Ads views carry their own color modules. None of this follows a white-label customer's primary color. This is adjacent to the chart work and worth folding into the same standard, since a standardized chart in an unstandardized card still looks inconsistent.

## The gap

There is a shared chart *engine* and no shared chart *design*. Two charts on the same screen can differ in legend position, axis formatting, grid inset and empty-state treatment, and two versions of the same chart on different networks routinely do.

## What needs to change

- A design standard covering legend, tooltip, axis, grid, series colors, and the empty, loading and error states.
- A shared option layer that contributes the standard to every chart, leaving each component to supply only its data and its value formatting. Extend `useEcharts` or sit beside it, either is defensible.
- Adoption across all 36 chart components, in two passes so the change is reviewable.
- Series colors from one palette, replacing the per-view color modules.
- Chart empty, loading and error states from one treatment, replacing per-component handling.
- No hardcoded brandable colors in chart or chart-card chrome.

## Notes and constraints

- The repo's analytics module rule is that all charts go through `useEcharts()`, which is good news: there is one funnel to standardize through.
- `docs/stories/analytics-empty-state-screens/` covers the "no account connected" screen for a whole analytics view. That is a different thing from a chart having no data for the selected range. Both are needed and they should not be conflated.
- `docs/stories/analytics-improvements/` Story 2 migrates the analytics module to TypeScript and TanStack Query. If that lands first, chart loading and error states come from query state, which makes this work simpler. Worth sequencing.
- `docs/stories/analytics-api-consistency/` standardizes API payload shape per platform. Complementary: this story standardizes how the numbers look, that one standardizes how they arrive.

## Files involved

- `contentstudio-frontend/src/modules/analytics/composables/useEcharts.ts`
- `contentstudio-frontend/src/modules/analytics/views/common/utils/echartsTooltip.ts`
- `contentstudio-frontend/src/modules/analytics/composables/analyticsDataHelpers.ts`
- `contentstudio-frontend/src/modules/analytics/components/common/{composables/useAnalytics.ts,helper.ts,AnalyticsCardWrapper.vue}`
- All 36 `*Chart*.vue` components under `components/competitor/` and `views/*/components/graphs/`
- Per-view chart option composables listed above
- `views/meta_ads/composables/metaAdsChartColors.ts` and any per-view color equivalents
- `contentstudio-frontend/src/modules/analytics/CLAUDE.md` (module conventions, should record the new standard)

## Mobile

None. Analytics dashboards are web only.
