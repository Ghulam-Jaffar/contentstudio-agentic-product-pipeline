# Research: Analytics Period Comparison ("Compare to")

**Date:** 2026-10-21
**Feature:** Let users choose what their analytics are compared against (previous period, same period last year, a custom range, or no comparison), and carry that choice into exported, emailed and scheduled reports.

**Prototype (canvas):** https://claude.ai/artifact/GkPcaAzYtjeL5jaedTAeyp

---

## 1. Problem

Every analytics card, rollup and report shows a change against a "previous period" that users cannot see or choose. That period is always the same number of days immediately before the selected range. The result:

- A calendar quarter or month is not compared with the previous calendar quarter or month, so the direction of the change can disagree with what the user sees by hand.
- The comparison dates and the previous value are not shown anywhere, so users can't check the numbers.
- When the comparison period has no data, cards show "0%", which reads as "no change".

Customers have hit this repeatedly:

| Ticket | Customer | What happened |
|---|---|---|
| CONT-3152 | Business & Agency - Small | Asked what "previous time period" means and for a custom comparison range, to compare month over month |
| CONT-2392 | Agency Unlimited (Elev8 Fun) | Engagement fell from 6.96K to 5.11K month to month, but the report showed +27.55%. The comparison window was hidden. Support asked for comparison dates and previous values in the report |
| CONT-4294 | MP Consulting, report #2335 | The Q3 report showed reach up when Q3 vs Q2 was down. The rolling window was compared instead of calendar Q2, and Facebook and Instagram expose only about 90 days of history, so the baseline was partial |
| CONT-4343 | Output-Films (workspace NATUS) | Questioned the Instagram Q3 vs Q2 engagement rate |

Abdul Wasay owned all four tickets.

## 2. Competitor research

"Comparison report" covers five types: period vs period, profile vs profile, brand vs competitors, brand vs industry, and tag or campaign vs tag or campaign. Period vs period is the most common.

| Tool | Period comparison | UI | Export/schedule |
|---|---|---|---|
| Sprout Social | Previous period by default. Custom ranges of any length, with a warning when lengths differ | "Compare to" section in every report's date picker. Report Builder widgets can show Chart, Table or %Change | PDF, send, schedule |
| Metricool | Previous equal-length period only. Cannot be changed | Arrows on each metric, with the absolute and % change on hover | PDF/PPT, monthly email |
| Agorapulse | Custom comparison period, defaulting to the period before. Custom needs Advanced/Custom plans | Colour-coded % in section headers, e.g. "-21.4% Compared to Mar 8 - Mar 15" | PDF, scheduled email |
| Hootsuite | Any other period | "Compare with another period" in the date picker | Unverified |

Recurring patterns: a comparison range inside the date picker, change badges on metrics, and custom comparison held back for higher plans. A dashed previous-period line overlaid on charts was not confirmed in any official doc.

Sources: support.sproutsocial.com (360027546792, 115005750066), help.metricool.com (how-comparison-periods-work-in-your-metrics), support.agorapulse.com (11066881), help.hootsuite.com (6442372179867).

## 3. How it works today (codebase)

**The comparison window is calculated in five places, and they don't agree:**

- **Go analytics:** `contentstudio-social-analytics-go/src/db/clickhouse/helpers.go:29-110` (`ParseDateRange`, `ParseStartEndDate`) sets `prevDateTo = dateFrom - 1 day` and `prevDateFrom = dateFrom - (dateTo - dateFrom)`. `PreviousPeriod()` in `db/clickhouse/paired.go:121` notes the previous window comes out one day shorter. Each platform service also has its own `prevPeriodParams` copy (e.g. `facebook/service.go:638`). `PairPeriods` / `ScanPairedRows` query both periods in one UNION ALL.
- **Campaign & Label:** `services/analytics/campaign_label/service.go:1820` `prevDate()` ends the previous window on the start date, a different formula.
- **Laravel (older paths):** `getPreviousDate()` in `Analyze/YoutubeController.php:36` and `PinterestController.php:42`, inlined in `FacebookController.php:42,58`.
- **Frontend (PDF render):** `usePDFReports.ts:771` builds `previous_date` from `previousPeriodStart()` in `src/composables/useDateTime.ts:214`.

**There's no request parameter for a different comparison range.** The one exception: `Analyze/OverviewController.php:211,279` and `CampaignLabelAnalyticsBuilder.php:31` already accept an optional `previous_date`. Nothing sends it.

**Response shapes**

- Platform summaries return `{ current, previous }`. The frontend computes growth in `analyticsDataHelpers.ts` `calculateGrowthAndDiff`, which returns 0 when previous is 0 or missing. That is today's misleading "0%".
- Overview returns `*_change_pct`, `secondary_*` and `diff_*` (Go `models/analytics/overview/types.go`).
- Meta and Google Ads use `metric.change`.

**UI**

- Date picker: `AnalyticsFilterBarWrapper.vue:299` → `CstDateRangePicker` with `preset="analytics"` (presets in `usePickerPresets.ts:54-183`).
- The range is saved as the `analytics_date_range` user preference.
- The only explanation is a tooltip: "Previous period means the same number of days immediately before your selected date range" (`locales/en/analytics.json:540`).
- No chart shows a previous-period overlay.

**Reports**

- Saved reports store only `date`.
- Scheduled runs recalculate it in `ReportsHelper::resolveReportingPeriod` (weekly: the last 7 days; monthly: the previous month, `ReportsHelper.php:131`).
- No comparison information is stored.
- PDF templates show changes through the same stat cards.
- The PDF is produced by either the Go report engine or the Vue report page through Gotenberg.

**Public API:** `app/Http/Controllers/Api/V1/Analytics/*` proxy to Go. Facebook's summary is documented as "current vs previous period" with `current` and `previous` objects. `AnalyticsQueryRequest` has no comparison parameter.

## 4. Proposed product direction (from the prototype, PO discussion 2026-10-02)

- **A "Compare to" control next to the date range**, with four options:
  - Previous period (default, today's behaviour)
  - Same period last year
  - Custom range
  - No comparison
- **The comparison is saved as a rule, not as dates.** Exports and emails turn the rule into dates when they run. Schedules move the rule forward with each run. The PDF cover prints both ranges.
- **Each widget declares a treatment:**

  | Widget | Treatment | Phase |
  |---|---|---|
  | Cards and rollups | Change plus previous value | 1 |
  | Bar and stacked charts | Change in the header | 1 |
  | AI Insights | Names the comparison | 1 |
  | Line charts | Dashed overlay | 2 |
  | Tables | Change column | 2 |
  | Top posts | Not compared, because posts use lifetime numbers | n/a |
  | Demographics | Not compared, because they are a snapshot | n/a |

- **Data gaps:** "No data for [dates]" instead of 0%, and "Partial: X of Y days have data".

**Open decisions:**

1. Custom range on schedules: not allowed, or allowed as a fixed baseline?
2. Monthly "previous period": the previous calendar month, or the same number of days? CONT-4294 points to calendar periods.
3. Weekly "last year": the same dates, or the same weekdays 52 weeks back?
4. Different lengths: compare totals with a warning, or rely on per-day averages (rollups already do this)?
5. Partial data inflates changes. In the prototype, 14 of 31 days of data showed ↑112%. Hide the change below a threshold?
6. Plan gating for custom ranges.
7. Group composition changes, raised in CONT-2392: when accounts are added to a group, should both periods use the same account set?

## 5. Related work

- **Analytics report preview epic** (CONT-4249 to CONT-4255): the preview should show the comparison once this lands.
- **CONT-3522:** analytics data retention policy, which affects how far back "last year" can go.
