# Research — Analytics metric consistency audit

Item 20 of the 7 Aug 2026 backlog batch.

## Two of the four requested items are already authored

The request had four parts. Two already have stories written and should not be duplicated:

| Requested | Already covered by |
|---|---|
| "Finalize Analytics report widget selection" | `docs/stories/analytics-report-sections-and-overview-cards/` — Story 1, *[Full Stack] Let users pick which sections to include in analytics reports*. Brings the section picker that workspace-level reports already have to Overview, the per-network dashboards, competitor analytics, Campaign and Label, and ads analytics. Every section stays selected by default. |
| "Standardize the number/count of Top Posts and Account Insights" | `docs/stories/analytics-report-sections-and-overview-cards/` — Story 2, *[Full Stack] Cap overview top posts and account insights at five and hide metrics a platform does not report*. Also related: `docs/stories/analytics-top-least-posts-consistency/` adds the Top Posts sort dropdown and a Least Posts section to Facebook, Instagram, LinkedIn and TikTok, matching Pinterest and YouTube. |

The counts finding that motivated the request is real and visible in the code:

- `views/facebook/components/OverviewSection.vue:69` — `const TOP_POSTS_LIMIT = 5`
- `views/instagram/components/OverviewSection.vue:99` — `const TOP_POSTS_LIMIT = 5`
- `views/youtube/composables/useYoutubeAnalytics.ts:170` — `const DEFAULT_TOP_AND_LEAST_POSTS_LIMIT = 10`
- `views/performance-report/label-and-campaign/composables/useLabelAndCampaignPosts.ts:130` — `export const DEFAULT_POSTS_LIMIT = 10`

Four constants, two values, four files. The existing story caps overview cards at five, which settles the overview surface. Whether YouTube's ten and the label-and-campaign ten should also become five is a question the audit story below should answer rather than assume, since they are different surfaces.

## What is genuinely not covered

The remaining two parts of the request:

- "Audit metrics across Analytics to ensure the same metric is named, calculated, formatted, and displayed consistently across different reports/screens."
- "Identify and resolve inconsistencies between widgets, reports, exports, and other Analytics surfaces."

Nothing in the backlog covers this. The closest is `docs/stories/analytics-api-consistency/`, which standardizes the **API payload and response shape** per platform. That is the transport layer. It does not address whether a metric called Engagement on the Facebook dashboard is the same number as Engagement in the exported PDF, or whether Reach is computed the same way in the Overview as on the per-network screen.

## Current state relevant to the audit

**Surfaces a metric can appear on**, each with its own rendering path:

1. Per-network dashboard sections (`views/{facebook,instagram,linkedin,twitter,youtube,tiktok,pinterest,gmb}/components/`)
2. The Overview dashboard (`views/overview/`)
3. Competitor analytics (`components/competitor/`)
4. Campaign and Label performance reports (`views/performance-report/label-and-campaign/`)
5. Ads analytics (`views/meta_ads/`, `views/google_ads/`)
6. Exported and emailed PDF reports (`views/reports/*.vue`, driven by `composables/usePDFReports.ts`)
7. The public analytics API (Facebook shipped, the rest in flight via `public-analytics-api-rollout`)
8. The Looker Studio connector

**Where formatting and derivation live today**, none of it single-sourced:

- `components/common/helper.ts` — includes per-platform payload builders with their own `accounts_limit` values (10 in several places, 0 in another)
- `components/common/composables/useAnalytics.ts`
- `composables/analyticsDataHelpers.ts`
- Per-network composables: `useYoutubeAnalytics.ts`, `useTwitterAnalytics.ts`, `usePinterestAnalytics.ts`, `useTiktokAnalytics.ts`, `useOverviewAnalytics.ts`
- Per-view formatting helpers: `views/meta_ads/composables/useMetaAdsTableFormatting.ts`
- Report-specific composables: `useMetaAdsReport.ts`, `useGoogleAdsReport.ts`, `usePDFReports.ts`
- Metric labels live in `src/locales/*/analytics.json`, so the same underlying metric can carry different label keys on different screens with nothing enforcing they match

**One observed naming inconsistency**: `views/overview/composables/useOverviewAnalytics.ts:995` refers to `'Platform-Specific Account Insights'` as a literal string key for an icon map, while the section is called Account Insights elsewhere. Small, but indicative: metric and section names are strings scattered across composables rather than a defined set.

**A known metric-truthfulness issue already being fixed**: the existing overview-cards story stops showing 0 for metrics a platform does not actually report, so a zero on screen means a real zero. The audit should treat that as the standard to generalise, not re-solve.

## What the audit story needs to produce

- One canonical dictionary of Analytics metrics: for each, its display name, its definition in plain language, its unit, its formatting rule, its abbreviation rule, and which platforms report it.
- For each metric, the list of surfaces it appears on and confirmation that each surface uses the same name, derivation and formatting.
- A list of every divergence found, each classified as a naming difference, a derivation difference, a formatting difference or a genuine per-platform difference that should be documented rather than removed.
- A decision per divergence, so the follow-up work is a list of fixes rather than a second investigation.

The audit is deliberately its own story. Doing the fixes without first agreeing what each metric means produces a second set of inconsistencies.

## Dependencies worth noting

- `analytics-api-consistency` is upstream. If the API shape is being standardized per platform, the metric dictionary should be the same dictionary those stories standardize toward, not a parallel one.
- The public analytics API rollout and the Looker Studio connector both expose metrics externally, so a rename decided in the audit has an external contract cost. The audit must flag which decisions are breaking for external consumers.
- The public Ads Analytics API epic's documentation story needs a metric glossary. It should cite this dictionary rather than writing a second one.

## Files involved

- `contentstudio-frontend/src/modules/analytics/components/common/helper.ts`
- `contentstudio-frontend/src/modules/analytics/components/common/composables/useAnalytics.ts`
- `contentstudio-frontend/src/modules/analytics/composables/analyticsDataHelpers.ts`
- `contentstudio-frontend/src/modules/analytics/composables/usePDFReports.ts`
- Per-network composables under `views/*/composables/`
- `views/meta_ads/composables/useMetaAdsTableFormatting.ts` and the Google Ads equivalent
- `views/reports/*.vue`
- `contentstudio-frontend/src/locales/*/analytics.json`
- `contentstudio-social-analytics-go/src/api/analytics/` — where metrics are computed
- `contentstudio-frontend/src/modules/analytics/CLAUDE.md` — should record the dictionary's location

## Mobile

None. Analytics is web only.
