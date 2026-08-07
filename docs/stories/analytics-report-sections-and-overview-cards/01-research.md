# Research: Analytics report sections + overview card improvements

Batch of four requested improvements:

1. Bring the "pick the sections you want" report modal (built for workspace-level reports) to normal analytics: social analytics, competitor analytics, campaign & label, ad analytics.
2. Cap top posts and account insights at 5 in overview analytics and its reports.
3. Stop rendering metrics a platform does not report in overview top-post cards and account-insight cards (they currently render as 0).
4. Add the publishing behaviour type-wise breakdown table (exists on Instagram and LinkedIn) to Facebook and any other platform whose data supports it.

## Backlog check

Existing analytics story folders were checked for overlap. Closest neighbours:

- `docs/stories/analytics-improvements/` covers reports architecture, FE TypeScript migration, sync/token/thumbnail/Sentry work. No overlap with these four.
- `docs/stories/multi-workspace-analytics-report/` is the combined multi-workspace report that shipped the widget selector in the first place. It is the reference implementation for item 1, not a duplicate.
- `docs/stories/analytics-api-consistency/`, `analytics-report-email-copy-and-download/`, `analytics-sync-date-range-modal/` are unrelated.

No duplicates. This is net-new work.

---

## 1. Report section (widget) selection

### Current state

The render engine already supports per-platform section selection end to end. What is missing is the UI and the two create paths.

**Go reports service (`contentstudio-social-analytics-go/`)**

- `src/services/reports/catalog/` holds 15 platform catalogs: `overview`, `facebook`, `instagram`, `linkedin`, `twitter`, `pinterest`, `tiktok`, `youtube`, `gmb`, `meta_ads`, `google_ads`, `campaign_label`, `competitor` (registers `facebook_competitor` + `instagram_competitor`).
- `catalog.go` exposes `DefaultWidgets(platform)` (ordered full report), `List(platform)`, and `Resolve(ids, platform)` which validates ids against the platform and returns entries in the requested order. Unknown ids are returned separately and skipped gracefully for stored definitions.
- `src/models/reports/catalog.go` defines `WidgetCatalogEntry` with `id`, `title`, `type`, `platforms`, `data_deps`, `chart`. `data_deps` is what lets the builder fetch only the data selected sections need.
- `src/api/reports/handler.go` already serves `GET /reports/widgets?platform=…`, and the package doc states it exists "for the UI selector / MCP".
- `src/services/reports/builder/widget.go:117` calls `catalog.Resolve(def.Widgets, def.Platform)`, so an empty `Widgets` yields the platform default and a non-empty one yields exactly that selection.

**Laravel backend**

- `GoReportsClient::buildDefinition()` already forwards `'widgets' => $report->widgets ?? []`. A stale comment above it says "When the widget selector ships, pass `$report->widgets` here" but the line is live.
- `ReportsModel` documents `@property array $widgets`.
- `WorkspaceReportsController` validates `'widgets' => 'sometimes|array'` and stores `$request->get('widgets', [])`.
- `AnalyticsReports::store()` (one-off render + email) and `ScheduleReports::store()` (recurring) neither validate nor persist `widgets`. This is the gap.
- No route proxies the Go widget catalog to the SPA.

**Frontend**

- `contentstudio-frontend/src/modules/setting/components/WorkspaceReportModal.vue` is the only surface that lets a user pick sections. Its widget list is a hardcoded 9-entry `WIDGETS` const for overview only, labels under `settings.workspace_reports.widgets.*` in `src/locales/*/settings.json`. Pattern worth copying: `Checkbox` grid, "Sections to include" label, hint copy, submit blocked when zero selected, and all-selected collapses to `[]` so the backend renders the platform default.
- Normal analytics report entry points all live behind one button: `src/modules/analytics/views/common/ExportButton.vue`, mounted from `views/common/AnalyticsFilterBarWrapper.vue`, with three actions:
  - Export PDF, via `useRenderReportMutation()`
  - Email PDF, via the `send-report-by-email` event into `components/reports/modals/SendReportByEmailModal.vue`
  - Schedule PDF, via the `schedule-report` event into `components/reports/modals/ScheduleReportModal.vue`
  `ExportButton` receives a `type` prop that is the report platform (`facebook`, `instagram`, `linkedin`, `twitter`, `pinterest`, `tiktok`, `youtube`, `gmb`, `meta_ads`, `google_ads`, `group`/`individual` for overview, `campaign-and-label`).
- Competitor analytics exports through its own path: `src/modules/analytics/composables/competitor/useReportExport.ts` posts `{ type: 'competitor', platform_type: facebook|instagram, … }` via `exportReportApi`.
- Export/schedule is gated by `canAccess('exports_schedule_reports')`.

### What needs to change

- Expose the Go widget catalog to the SPA so the selector can never offer a section the engine cannot render.
- Accept, validate, and persist `widgets` on the one-off report path and the scheduled report path (and on the competitor export payload).
- Add a section picker to the three report actions on every analytics dashboard in scope, plus the competitor export.
- Keep "all selected" equivalent to today's behaviour so existing reports and existing scheduled reports are unaffected.

### Platform section counts (from the Go catalogs, for sizing the picker)

Overview 9, Facebook ~19, Instagram ~22, LinkedIn ~16, Twitter 5, Pinterest ~11, TikTok ~11, YouTube ~17, GMB ~13, Meta Ads ~15, Google Ads ~19, Campaign & Label ~13, Facebook/Instagram competitor ~14. Several platforms have enough sections that a flat checkbox grid needs grouping or scrolling, unlike the 9-item overview grid in the workspace modal.

---

## 2 & 3. Overview top posts and account insights

### Current state

**Top posts** (`src/modules/analytics/views/overview/components/TopPostsComponent.vue`)

- `postsPerView` prop defaults to 5, `visiblePostsCount` starts there, and a "load more" control appears whenever `filteredPosts.length > visiblePostsCount`. So 5 is the initial page size, not a cap.
- Per-platform tabs (`overall`, instagram, facebook, linkedin, pinterest, tiktok, youtube) filter the same list.
- Card metrics come from `getPlatformMetrics(post)` in `views/overview/composables/useOverviewAnalytics.ts` (around line 1345):
  - Two fixed rows: post type, then the primary metric driven by the sort dropdown (reach / engagement / impressions), then the opposite of reach/engagement.
  - Then a `platformMetrics` map with explicit entries for `pinterest` and `instagram` only. Every other platform falls through to `default`: Likes, Comments, Shares. Shares is the visible offender for platforms that do not report it, and the impressions primary metric falls back to `post.views > 0 ? post.views : post.reach` regardless of platform support.

**Account insights** (`src/modules/analytics/components/competitor/AccountStatisticsWrapper.vue` + `AccountStatistics.vue`)

- Despite the `competitor/` folder, this is the overview "Accounts insights" section. i18n lives under `analytics.overview.account_statistics.*` and `analytics.overview.account_statistics_card.*`.
- `visibleAccountsCount` is responsive: 8 at 2xl, 6 at xl, 4 at lg, 2 below, with "Load more" expanding to the full list. Report view hides the button but keeps the viewport-derived count.
- `AccountStatistics.vue` `statistics` computed always emits the same five rows for every platform: Followers, Posts, Engagement, Impressions, Reach. The sparkline metric dropdown (`metrics` computed) always offers engagement, reach, impressions, posts. Nothing is filtered by platform, so unsupported metrics render 0 with a 0% trend.

**Reports**

- `src/services/reports/builder/widgets_overview.go` registers `accounts_insights` and `overview_top_posts` (built by the shared `registerAccountInsightsWidget` / `registerTopPostCardsWidget` helpers). No count cap and no per-platform metric filtering there either.
- The overview report page (`views/reports/OverviewReport.vue`) reuses the same two FE components with `isReportView`.

### What needs to change

- Hard cap both sections at 5 items in the dashboard and in generated reports, and drop the load-more affordance for them.
- Build the metric list per platform instead of a fixed list, in both the top-post card and the account-insight card, on the dashboard and in the report engine.

### The metric-support matrix, verified from the query layer

No guesswork needed. `src/db/clickhouse/analytics-get-queries/overview/repository.go` zero-fills unsupported metrics per platform, with comments saying so. Only the six platforms the overview tabs expose are listed (Twitter and GMB are not in overview).

**Account insight cards** (per-account totals query, around lines 390 to 690):

| Platform | Followers | Posts | Engagement | Impressions | Reach |
|---|---|---|---|---|---|
| Facebook | yes | yes | yes | yes | yes |
| Instagram | yes | yes | yes | yes | yes |
| LinkedIn | yes | yes | yes | yes | yes |
| TikTok | yes | yes | yes | yes (views) | **copy of views** |
| YouTube | yes | yes | yes | yes (views) | **hardcoded 0**, comment: "reach metric not available for YouTube" |
| Pinterest | yes | yes | yes | yes | **hardcoded 0**, comment: "reach metric not available for Pinterest" |

So the confirmed defects are: Reach renders 0 for YouTube and Pinterest, and Reach for TikTok and YouTube is the views figure duplicated into the reach column, which makes it a copy of Impressions rather than a real reach number. The same zero-fill exists in the graph series builder (lines 968, 1001, 1005), so the sparkline metric dropdown has the same problem.

**Top post cards** (top-posts union query, around lines 1805 to 1950):

| Platform | Likes | Comments | Shares | Saves | Pin clicks | Outbound clicks | Views | Reach |
|---|---|---|---|---|---|---|---|---|
| Facebook | yes | yes | yes | **0** | 0 | 0 | yes | yes |
| Instagram | yes | yes | **0** | yes | 0 | 0 | yes | yes |
| LinkedIn | yes | yes | yes | **0** | 0 | 0 | **0** | yes |
| TikTok | yes | yes | yes | **0** | 0 | 0 | yes | copy of views |
| YouTube | yes | yes | yes | **0** | 0 | 0 | yes | copy of views |
| Pinterest | **0** | **0** | **0** | yes | yes | yes | **0** | yes (impressions) |

The frontend `platformMetrics` map already gets Instagram and Pinterest right, which is why those two platforms have explicit entries. The gap is that Facebook, LinkedIn, TikTok and YouTube all share the `default` block, and the reach/views primary metric is picked from the sort dropdown without checking platform support. LinkedIn is the clearest case: it has no views figure at all.

---

## 4. Publishing behaviour breakdown table

### Current state

- `PublishingBehaviourBreakdownTable.vue` exists twice, once per platform: `views/instagram/components/` and `views/linkedin/components/`. Rendered on the dashboard from `views/<platform>/components/OverviewSection.vue` and in the report layouts `views/reports/InstagramReport_v2.vue` and `views/reports/LinkedinReport_v2.vue`.
- Data source is the existing `publishingBehaviour` section response, field `publishing_behaviour_rollup.current`, an array of one row per media type plus a `TOTAL` row. Instagram columns: Media Type, Total Posts, Engagement, Likes, Comments, Saved, Reach, Views. LinkedIn adds Shares, Clicks, Impressions.
- Media types are rendered with coloured icons and translated labels (`REELS`, `IMAGE`, `CAROUSEL_ALBUM`, `VIDEO`, `TOTAL`).
- Go models:
  - `instagram/types.go` and `linkedin/types.go`: `PublishingBehaviourRollup map[string][]PublishingBehaviourMediaType` (per media type).
  - `facebook/types.go`: `PublishingBehaviourRollup map[string]*PublishingRollup` with `post_count`, `post_engagement`, `post_reactions`, `post_comments`, `post_clicks`, `post_impressions`, `post_shares`. Aggregate only, no type dimension.
- Report catalog: Instagram has `ig_publishing_breakdown`, LinkedIn has `publishing_breakdown`. Facebook has publishing-versus-metric charts but no breakdown widget. Twitter, Pinterest, TikTok, YouTube, GMB have none.

### Platform feasibility (ClickHouse columns + the values actually stored)

| Platform | Type column | Distinct values in practice | Verdict |
|---|---|---|---|
| Facebook | `facebook_posts.media_type`, `status_type` | photo/image, video, reel, album/carousel, link, share, text (see `utils/parsing/facebook_competitor_parser*.go`) | **In scope.** Rich type dimension, clear value |
| YouTube | `youtube_videos.media_type` | `"video"` and `"short"` only (`models/kafka/youtube.go:57`) | **In scope but thin.** 2 rows plus total. Shorts versus long form is a real creator decision, so still worth it |
| Pinterest | `pinterest_pins.media_type`, `creative_type` | `"image"` and `"video"` only (`analytics-get-queries/pinterest/repository.go:455` restricts to exactly those two) | **Out of scope.** Not literally one type, but a 2-row table for image versus video is not worth the data plumbing |
| TikTok | none | video only, `'video' AS media_type` is hardcoded in the query | Out |
| Twitter | none | n/a | Out |
| GMB | none | n/a | Out |

Facebook's analytics API already accepts a `media_type` filter param (`src/api/analytics/facebook/handler.go:46`), so the dimension is already queryable on that platform.

### What needs to change

- Add a per-type rollup to the Facebook and YouTube publishing behaviour responses in the Go analytics service, shaped like the Instagram/LinkedIn one so the table component can be shared.
- Register a breakdown widget in the Facebook and YouTube report catalogs and wire the adapters.
- Render the table on both dashboards and in their report layouts. The two existing components are near-duplicates and are good candidates for one shared component with a per-platform column config.

---

## Files involved

**Frontend (`contentstudio-frontend/`)**

- `src/modules/analytics/views/common/ExportButton.vue`
- `src/modules/analytics/views/common/AnalyticsFilterBarWrapper.vue`
- `src/modules/analytics/components/reports/modals/ScheduleReportModal.vue`, `SendReportByEmailModal.vue`, `composables/useScheduleReportModal.ts`
- `src/modules/analytics/composables/competitor/useReportExport.ts`
- `src/modules/analytics/queries/useReportsQueries.ts`, `reportsQueryOptions.ts`
- `src/modules/setting/components/WorkspaceReportModal.vue` (reference pattern)
- `src/modules/analytics/views/overview/components/TopPostsComponent.vue`
- `src/modules/analytics/views/overview/composables/useOverviewAnalytics.ts`
- `src/modules/analytics/components/competitor/AccountStatisticsWrapper.vue`, `AccountStatistics.vue`
- `src/modules/analytics/views/instagram/components/PublishingBehaviourBreakdownTable.vue`, `src/modules/analytics/views/linkedin/components/PublishingBehaviourBreakdownTable.vue`
- `src/modules/analytics/views/facebook/components/OverviewSection.vue`, and the pinterest/youtube equivalents
- `src/modules/analytics/views/reports/FacebookReport_v2.vue`, `PinterestReport.vue`, `YoutubeReport.vue`, `OverviewReport.vue`
- `src/api/analytics-store.ts`, `src/api/analytics.ts`, `src/config/api-utils.ts`
- `src/locales/*/analytics.json` (8 locales)

**Laravel backend (`contentstudio-backend/`)**

- `app/Http/Controllers/Analytics/Analytics/AnalyticsReports.php`
- `app/Http/Controllers/Analytics/Analytics/ScheduleReports.php`
- `app/Services/Analytics/GoReportsClient.php`
- `app/Models/Analytics/ReportsModel.php`
- `routes/api.php`

**Go analytics + reports service (`contentstudio-social-analytics-go/`)**

- `src/api/reports/handler.go`
- `src/services/reports/catalog/facebook.go`, `pinterest.go`, `youtube.go`
- `src/services/reports/builder/widgets_overview.go`, `widget.go`
- `src/services/reports/adapters/facebook/adapter.go`, `pinterest/adapter.go`, `youtube/adapter.go`, `overview/adapter.go`
- `src/models/analytics/facebook/types.go`, `pinterest/types.go`, `youtube/types.go`
- `src/db/clickhouse/analytics-get-queries/` (facebook, pinterest, youtube publishing behaviour queries)

## Note on the metric approach

The responses cannot distinguish "absent" from "zero" today, because the query layer zero-fills unsupported metrics on purpose. So the stories use a **declared per-platform metric allowlist**, seeded from the verified matrix above, rather than an "is it null" check. Fixing the zero-fill upstream so the service omits unsupported fields would be the cleaner long-term shape, but it is a breaking response change across every overview consumer and is deliberately out of scope here. Flagging it in case engineering prefers to fix that upstream first.
