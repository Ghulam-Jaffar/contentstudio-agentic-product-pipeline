# Epic: Standardize chart UI across Analytics

## Problem

Analytics has 36 chart components. They share an engine and nothing else. `useEcharts()` owns the instance lifecycle and `buildAxisDateTooltip()` gives one tooltip shape, but everything a user actually sees is built per component: legend presence and position, axis label formatting and rotation, grid padding, series colours, value abbreviation, and the empty, loading and error states.

The result is that two charts on the same screen can sit at different insets from their card, label their axes differently and legend themselves differently. Worse, the same chart exists as separate per-network files (`AudienceGrowthChart`, `HashtagsChart`, `PublishingBehaviourChart`, `DualChartComponent` all recur across networks), so fixing Instagram's version does not fix LinkedIn's, and they have drifted apart over time.

Adjacent to this, chart cards and report section headers carry hardcoded colours (report section headings use a fixed gradient, ads views have their own colour modules), so none of it follows a white-label customer's brand colour.

## Goal

Define the chart standard once and make it the default every chart gets, so a chart component supplies its data and its value formatting and nothing else. Same legend, same tooltip, same axes, same grid, same palette, same empty, loading and error treatment, everywhere in Analytics, and no fixed brandable colours.

## Scope

Three stories: one design, two frontend. The frontend work is split into two passes so each is reviewable rather than one 36-component diff.

## Sequencing

Design first. Then the shared layer plus the per-network charts, which is the largest and most-seen group. Then ads, competitor and overview. If **[FE] Migrate the analytics module to TypeScript & TanStack Query** is going to land in the same period, prefer landing it first: chart loading and error states come from query state, and doing this work on top of hand-rolled loading flags means redoing it.

## Out of scope

- Changing which charts exist, what they plot, or the metrics behind them. This is presentation only.
- The "no account connected" analytics screen. That is a whole-view empty state and is covered by **[FE] Redesign the Analytics empty-state screens with connect-account onboarding**. This epic covers a chart having no data for the selected range, which is a different state.
- Standardizing the API payloads behind the charts, covered by the analytics API consistency stories.
- Mobile. Analytics dashboards are web only.

## Stories

1. `[Design] Define the Analytics chart standard`
2. `[FE] Build the shared chart option layer and adopt it in the per-network charts`
3. `[FE] Adopt the shared chart layer in the ads, competitor and overview charts`

---
---

# 1. [Design] Define the Analytics chart standard

### Description

Analytics charts were each designed at the moment they were built, over a long period, so the module now reads as several products stacked together: legends sit in different places, axes format numbers differently, charts sit at different insets inside identical cards, and a chart with no data looks different depending on which network you are looking at. This story defines one chart standard so every chart in Analytics can be brought to it.

### Workflow

N/A. Design deliverable.

### Acceptance criteria

- [ ] A representative sample of existing charts across per-network, ads, competitor and overview views is audited, and the inconsistencies are listed explicitly rather than implied by the mockups.
- [ ] Legend treatment is specified: whether a legend appears, its position, orientation, item shape, text style, and behaviour when there are many series.
- [ ] Tooltip treatment is specified for axis-triggered and item-triggered charts, including the header, one row per series, and how values are formatted for counts, percentages, currency and durations.
- [ ] Axis treatment is specified: label style, rotation rules, tick density, how long labels are truncated, and how large numbers are abbreviated.
- [ ] Grid and card inset are specified, so every chart sits at the same distance from its card edges.
- [ ] A single categorical series palette is specified, ordered, with enough distinct colours for the busiest chart in the module, and accessible against the chart background.
- [ ] Colours are specified as theme tokens wherever they are brandable, and it is stated explicitly which chart colours are intentionally fixed data colours rather than brand colours.
- [ ] Chart loading state is specified.
- [ ] Chart empty state is specified for a chart with no data in the selected range, distinct from the whole-view no-account-connected screen.
- [ ] Chart error state is specified for a chart whose data failed to load, including a retry affordance.
- [ ] A partial-data state is specified for charts where some series have data and others do not.
- [ ] Chart card chrome is specified: title, subtitle, info affordance, and any per-chart control such as a metric selector.
- [ ] Responsive behaviour is specified, including what happens to the legend and axis labels at narrow widths, and how a wide chart scrolls without the page scrolling.
- [ ] The standard is specified for the chart types actually in use: line, area, bar, stacked bar, grouped bar, dual-axis, pie or donut, and any others found in the audit.
- [ ] The report-section header treatment is specified using theme tokens, replacing the current fixed gradient.
- [ ] The design names the design-library components to reuse and flags anything unavailable as a component gap.
- [ ] The standard is documented in a form a developer can implement from without further design questions, including redlines for spacing and type.

### Mock-ups

This story produces them.

### Impact on existing data

None.

### Impact on other products

- Web app only. Analytics dashboards are web only.
- The standard also governs charts as they appear in exported PDF reports, so the design has to hold up in a print context, not only on screen.
- White-label domains render all of this.

### Dependencies

None. Should start immediately.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories)
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled) — legend and axis labels must tolerate longer translated strings
- [ ] UI theming supported (default + white-label, design library components are being used)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- Every chart goes through `contentstudio-frontend/src/modules/analytics/composables/useEcharts.ts`, so there is one funnel the standard can be applied through. `useEcharts` currently contributes nothing visual; it takes an option object and passes it on.
- One piece of the standard already exists and is worth extending rather than replacing: `views/common/utils/echartsTooltip.ts` provides `buildAxisDateTooltip()` with a workspace-timezone date header and a `formatValue` callback per caller. Its docblock already articulates the "shared scaffolding, per-caller formatting" model this epic generalises.
- The per-network duplication is the main audit target: `AudienceGrowthChart`, `HashtagsChart`, `PublishingBehaviourChart` and `DualChartComponent` each exist several times under `views/*/components/graphs/`. Comparing two copies of the same chart is the fastest way to see the drift.
- Existing per-view colour modules such as `views/meta_ads/composables/metaAdsChartColors.ts` show what the single palette has to replace.
- The fixed gradient on report section headers is visible in `views/reports/OverviewReport.vue` (`from-[#01AAFF] to-[#0ACADD]`, repeated per section).

---
---

# 2. [FE] Build the shared chart option layer and adopt it in the per-network charts

### Description

Charts across the per-network analytics dashboards differ in legend, axes, grid inset, colours and how they behave with no data, because each was built with its own option object. This story builds the shared layer that contributes the agreed standard to any chart, then adopts it across the per-network charts, which is where users spend most of their time in Analytics.

After this story a chart component supplies its data and its value formatting. Everything else comes from the shared layer.

### Workflow

1. User opens a network's analytics dashboard.
2. Every chart on the screen shares the same legend placement, axis style, grid inset and palette.
3. A chart with no data for the selected range shows the standard empty treatment, not a blank canvas or a zero line.
4. A chart whose data failed to load shows the standard error treatment with a retry, rather than appearing empty.
5. While data is loading, every chart shows the standard loading treatment.
6. User switches to a different network. The charts there look and behave the same way.
7. User narrows the browser. Legends and axis labels follow the agreed responsive behaviour, and the page never scrolls sideways.

### Acceptance criteria

- [ ] A shared chart option layer exists that contributes the agreed legend, tooltip, axis, grid and palette defaults, and lets a caller override only what is genuinely chart-specific.
- [ ] A chart adopting the layer supplies its data and its value formatting, and does not restate legend, axis, grid or palette configuration.
- [ ] All per-network chart components adopt the layer: Facebook, Instagram, LinkedIn, X, YouTube, TikTok, Pinterest and Google Business Profile.
- [ ] Charts that exist once per network under the same name render identically across networks after adoption.
- [ ] Series colours come from the single agreed palette. No per-view colour module remains in use by an adopted chart.
- [ ] Chart loading state matches the agreed treatment on every adopted chart.
- [ ] Chart empty state matches the agreed treatment when a chart has no data for the selected range, and is distinguishable from the whole-view no-account-connected screen.
- [ ] Chart error state matches the agreed treatment when a chart's data fails to load, and offers a retry that re-requests only that chart.
- [ ] Partial data, where some series have values and others do not, renders per the agreed treatment rather than plotting a flat zero line for the missing series.
- [ ] Value formatting is consistent across adopted charts: the same number is abbreviated the same way, percentages and currency are formatted the same way, and durations are formatted the same way.
- [ ] Tooltips on adopted charts follow the agreed treatment, including the header and one row per series.
- [ ] No hardcoded brandable colour remains in any adopted chart or its card chrome.
- [ ] Adopted charts render correctly on a white-label domain, following the customer's primary colour where the standard says brandable.
- [ ] Adopted charts render correctly down to the smallest supported width, with the agreed legend and axis behaviour, and no horizontal page scroll.
- [ ] Adopted charts still render correctly inside exported PDF reports.
- [ ] Every chart still plots the same data it plotted before. No metric, series or date range changes as a result of this story.
- [ ] Chart instances are still torn down on unmount, with resize handling cleaned up, so the change does not introduce leaks.
- [ ] Every user-visible string introduced or changed is translated and present in every locale directory.

### UI copy

**Chart empty state** (no data for the selected range)
- Headline: `No data for this period`
- Subtext: `There is nothing to show for the dates you selected. Try a wider date range.`

**Chart error state**
- Headline: `We couldn't load this chart`
- Subtext: `Something went wrong fetching this data. Your other charts are unaffected.`
- Button: `Try again`

**Partial data note** (inline, when some series have no data)
- `Some data isn't available for this period.`

All strings go through translation and land in every locale directory in the same change. Note the deliberate absence of em dashes.

### Mock-ups

Provided by **[Design] Define the Analytics chart standard**.

### Impact on existing data

None. Presentation only.

### Impact on other products

- Web app only.
- Exported PDF reports render these charts, so the change is visible there too.
- White-label domains.

### Dependencies

- Depends on **[Design] Define the Analytics chart standard**.
- Should land after **[FE] Migrate the analytics module to TypeScript & TanStack Query** if that story is in the same period, so loading and error states come from query state rather than being hand-rolled and then redone.
- Independent of **[FE] Redesign the Analytics empty-state screens with connect-account onboarding**, which covers a different state.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories)
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled)
- [ ] UI theming supported (default + white-label, design library components are being used)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- The natural home for the shared layer is next to or inside `contentstudio-frontend/src/modules/analytics/composables/useEcharts.ts`, which every chart already calls. It currently exposes only `{ chartRef, setup(options, notMerge) }` and contributes no visual defaults; having `setup` merge the standard before handing the option to ECharts is one way in, and a separate option-builder the caller composes is another.
- `views/common/utils/echartsTooltip.ts` already implements the shared-scaffolding-plus-caller-formatting model with `buildAxisDateTooltip()`. Generalising it is likely cheaper than writing a new tooltip system.
- Existing option construction to consolidate: `composables/analyticsDataHelpers.ts`, `components/common/composables/useAnalytics.ts`, `components/common/helper.ts`, and per-network composables `views/youtube/composables/useYoutubeAnalytics.ts`, `views/twitter/composables/useTwitterAnalytics.ts`, `views/pinterest/composables/usePinterestAnalytics.ts`, `views/tiktok/composables/useTiktokAnalytics.ts`.
- The duplicated-per-network charts are the highest-value targets, since adopting the layer is also the moment to decide whether `AudienceGrowthChart`, `HashtagsChart`, `PublishingBehaviourChart` and `DualChartComponent` should collapse into one component per concept with a network prop. That consolidation is optional here and would be a large win; if it is not done, note it as follow-up rather than leaving it implied.
- `components/common/AnalyticsCardWrapper.vue` is the existing card chrome and is where the standard's card inset and title treatment most likely belong.
- The module's own conventions doc is `contentstudio-frontend/src/modules/analytics/CLAUDE.md`, which already states that all charts go through `useEcharts()` and that endpoints come from constant maps. Recording the new chart standard there is what stops the next chart from drifting again.

---
---

# 3. [FE] Adopt the shared chart layer in the ads, competitor and overview charts

### Description

The ads, competitor and overview charts are the remaining group still building their own option objects, with their own colour modules and their own empty and error handling. This story brings them onto the shared standard so Analytics is consistent end to end, and removes the last of the per-view chart colour modules.

### Workflow

1. User opens Meta Ads or Google Ads analytics. The charts there look like the charts on the network dashboards.
2. User opens competitor analytics. Same.
3. User opens the Analytics overview. Same.
4. Charts with no data, failing data or partial data behave per the standard in all three areas.
5. Exported reports for these areas render the standardized charts correctly.

### Acceptance criteria

- [ ] The ads charts adopt the shared layer: the Meta and Google `ClicksVsCTRChart` and `ImpressionsVsSpendChart`, and any other chart in the two ads views.
- [ ] The competitor charts adopt the shared layer: `SpecificPostTypeChart`, `FollowersComparisonBarChart`, `DoubleBarChart`, `PostEngagementChart`, `PostReactsDistributionChart`.
- [ ] The overview charts adopt the shared layer.
- [ ] The duplicated Meta and Google versions of the same chart render identically after adoption.
- [ ] Per-view chart colour modules are removed from use, and series colours come from the single agreed palette.
- [ ] Currency values in the ads charts are formatted per the standard and always carry their currency, so a chart across accounts in different currencies is not misleading.
- [ ] Chart loading, empty, error and partial-data states match the agreed treatment in all three areas.
- [ ] Report section headers use theme tokens, replacing the current fixed gradient, across every report view.
- [ ] No hardcoded brandable colour remains in these charts or their card chrome.
- [ ] The charts render correctly on a white-label domain.
- [ ] The charts render correctly down to the smallest supported width, with no horizontal page scroll.
- [ ] The charts render correctly inside exported PDF reports for these areas.
- [ ] Every chart still plots the same data it plotted before.
- [ ] Every user-visible string introduced or changed is translated and present in every locale directory.
- [ ] After this story, no chart component in the analytics module builds its own legend, axis, grid or palette configuration. Any that cannot adopt the layer is documented with the reason.

### UI copy

Reuses the empty, error and partial-data copy introduced by **[FE] Build the shared chart option layer and adopt it in the per-network charts**. No new strings expected beyond area-specific empty-state wording if the design calls for it.

### Mock-ups

Provided by **[Design] Define the Analytics chart standard**.

### Impact on existing data

None. Presentation only.

### Impact on other products

- Web app only.
- Exported PDF reports for ads, competitor and overview.
- White-label domains.

### Dependencies

- Depends on **[FE] Build the shared chart option layer and adopt it in the per-network charts** for the layer itself.
- Touches the same ads views as the public Ads Analytics API epic, but that epic is backend-only, so there is no conflict.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories)
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled)
- [ ] UI theming supported (default + white-label, design library components are being used)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- Ads chart option construction lives in `views/meta_ads/composables/useMetaAdsPerformanceCharts.ts` and `views/google_ads/composables/useGoogleAdsPerformanceCharts.ts`, with colours in `views/meta_ads/composables/metaAdsChartColors.ts`.
- The Meta and Google `ClicksVsCTRChart` and `ImpressionsVsSpendChart` are duplicated files with the same names under each ads view. Adopting the layer is the natural moment to consider collapsing them into one component with a platform prop.
- Overview option construction is in `views/overview/composables/useOverviewAnalytics.ts`. Competitor charts sit under `components/competitor/`.
- The report section gradient appears per section heading in `views/reports/OverviewReport.vue` and likely in the sibling report views (`FacebookReport_v2`, `InstagramReport_v2`, `LinkedinReport_v2`, `TwitterReport`, `YoutubeReport`, `TiktokReport`, `GmbReport`, `MetaAdsReport`, `GoogleAdsReport`, `FbCompetitorReport`, `IgCompetitorReport`). Worth grepping for the hex values rather than fixing them one file at a time.
