# 01 — Research: X (Twitter) in Competitor Analytics

**Date:** 2026-08-18
**Feature:** Add X (Twitter) as a third platform in the Analytics → Competitors module, alongside Facebook and Instagram.
**Framing constraint (from the PO):** stay **in sync with the X platform-analytics work already in flight** — same X API surface, same metering model, same wallet/credit currency. **No new or additional data sources.**

---

## 0. What this feature is (and the one thing that makes it different)

Competitor analytics today lets a workspace build a *report* of up to **5 competitors** on **Facebook** or **Instagram**, compared against **the workspace's own connected profile**, which is included automatically and never occupies one of the 5 slots. A full report is therefore **6 profiles**: yours plus five rivals. It compares followers, posting activity, engagement, hashtags, bios and top/least posts side by side. This feature adds **X** as a third platform in the same module.

The important difference from Facebook and Instagram is not the UI — it's the **economics**:

| | Facebook / Instagram competitors | X competitors |
|---|---|---|
| Data source | Meta Graph API (page public data / IG business discovery) | X API v2 (pay-per-use) |
| Cost per read | **Free** | **$0.005 per post read, $0.010 per user read** |
| Sync model today | **Daily cron**, all competitors, always on | A daily cron would cost ~**$29–47/month per report** |
| Metering | None | **Must** meter, same as X platform analytics |

So the whole design question for X competitors is: *how do we give the same insight without an always-on cron burning real money?* The answer that keeps us "synced with X analytics" is to **copy the X platform-analytics sync model exactly** — count-based, on-demand, cost shown before you spend, charged to the same X wallet (or legacy X credits).

---

## 1. Competitor & industry research

### 1.1 The market context

X competitor tracking got **structurally more expensive for every vendor** in 2026. X killed the free tier, closed Basic/Pro to new signups in February 2026, auto-migrated remaining Basic subscribers on June 1 2026, and made **pay-per-use the default**. The market split into two camps: tools that pay X's official rates and pass the cost on, and tools that quietly scrape or buy third-party data.

ContentStudio is already in the first camp for X publishing and X analytics (the X wallet / X credits model), so the competitor module should stay there too.

### 1.2 Competitor analysis

| Competitor | X competitor tracking? | Key capabilities | Pricing / gating | UX approach | Differentiator |
|---|---|---|---|---|---|
| **Sprout Social** | ✅ Dedicated **X Competitors Report** | Followers + **net follower growth**, engagements broken down **by engagement type**, published content by type, posting volume; benchmarks you against the *average* of the tracked set | Premium Analytics tiers; enterprise pricing | A single side-by-side report; your profile is always one of the compared rows | "Average of profiles compared" benchmark row — turns 5 competitors into one industry baseline |
| **Rival IQ** | ✅ Core product | Cross-network competitive landscape (X, IG, FB, LinkedIn, TikTok, YouTube), engagement rate by follower, top posts, posting cadence | Paid tiers from ~$239/mo | "Landscape" = a saved set of brands, not a per-network report | Head-to-head landscapes + a public Social Insights API |
| **Socialinsider** | ✅ | Public-API-only competitor data across FB, IG, TikTok, LinkedIn, YouTube, X; benchmarks + AI-assisted takeaways | From ~$99/mo annual | Report-per-competitor-set | Explicit "only what public APIs allow" positioning — an honesty pattern worth copying |
| **Hootsuite** | ✅ via **Perch** | Competitor watchlist of up to **20 accounts** (plan-dependent), benchmarking, listening-backed context | Bundled into higher plans / add-on | Watchlist, not report-builder | Merges benchmarking with listening (share of voice) |
| **Metricool** | ✅, but X is an **add-on** | Up to **100 competitor profiles**, posting schedule + engagement-rate comparison, best-time heatmaps | **X is not in any plan by default — $5/month per connected X account add-on**; X analytics limited to premium users | Add competitors by handle, grid comparison | The closest precedent to what we're doing: **X is explicitly priced separately because X charges them** |
| **Buffer** | ❌ (no competitor module) | — | — | — | — |
| **Later / Publer / SocialBee / Loomly / Sendible** | Mostly ❌ or IG/FB-only | Some offer light IG competitor tracking | — | — | — |
| **Agorapulse** | Partial | Competitor benchmarking on FB/IG; X coverage reduced post-API-changes | Paid tiers | — | — |

### 1.3 Key findings

1. **X competitor tracking is a paid-tier feature everywhere it exists.** Nobody gives it away on entry plans anymore.
2. **Metricool's $5/account X add-on is the strongest precedent for our model** — it proves users accept "X costs extra because X charges us," which is exactly the story the X wallet already tells.
3. **The metric set is remarkably consistent** across Sprout, Rival IQ and Socialinsider: followers + growth, posting volume, posting mix by type, engagement total + engagement rate, top posts, hashtags. That's *already* the widget set our FB/IG competitor reports ship.
4. **Nobody surfaces competitor impressions/views.** They can — `impression_count` is public on X — but the tools we surveyed lead with engagement. Showing **competitor impressions and view-based engagement rate** is a genuine, cheap differentiator for us (see §2.2).
5. **Share of voice / mention volume** is the one capability Sprout and Hootsuite have that we would *not* match in v1 — it needs X's search endpoints, which is net-new API surface and net-new spend.

### 1.4 Table stakes vs delighters

- **Table stakes:** add competitors by handle, followers + change, posting volume, engagement total + engagement rate, post type mix, top posts, side-by-side table, PDF export.
- **Delighters:** competitor **impressions/views** and view-based engagement rate (X-only, nobody does it), a benchmark "average of tracked profiles" row (Sprout's pattern), best-time-to-post heatmap (Metricool's pattern).
- **Explicitly not v1:** share of voice, sentiment, mention volume, ad/promoted detection, follower demographics — all either impossible on X's public API or net-new paid API surface.

---

## 2. X API deep dive — what is actually allowed

> This is the section the PO asked for in plain words: what we *can* build, what we *cannot*, and what it costs.

### 2.1 The two endpoints we already use (and nothing more)

X platform analytics in ContentStudio fetches through exactly two calls, confirmed in the Go fetcher:

- `GET /2/users` → one call per synced account (`src/clients/social/twitter.go:323`)
- `GET /2/users/:id/tweets` → paginated timeline, `max_results` + `pagination_token` only, **no date window** (`src/clients/social/twitter.go:256`)

Both work with an **app-only bearer token against any public account** — which is precisely why competitor analytics is possible at all. A competitor is just "a user we don't own", and public data is public.

**Staying inside these two endpoints satisfies the PO's "no new or additional information" constraint.** Adding one call — resolving a handle to an ID — is unavoidable, and that is still the same `users` endpoint family.

### 2.2 What we CAN get for a competitor (app-only bearer, public account)

**Profile level** — `GET /2/users/by/username/:handle` with `user.fields`:

| Field | Feeds which widget |
|---|---|
| `public_metrics.followers_count` | Followers, comparative table, followers chart |
| `public_metrics.following_count` | Bio analysis / comparative table |
| `public_metrics.tweet_count` | Lifetime post count |
| `public_metrics.listed_count` | Bio analysis (X-specific signal of authority) |
| `description`, `location`, `url`, `created_at`, `profile_image_url`, `verified` / `verified_type` | Bio Analysis widget, competitor tile avatar |

**Post level** — `GET /2/users/:id/tweets` with `tweet.fields` + `expansions`:

| Field | Feeds which widget |
|---|---|
| `public_metrics.like_count` | Engagement, top/least posts |
| `public_metrics.retweet_count` | Engagement (reposts) |
| `public_metrics.reply_count` | Engagement |
| `public_metrics.quote_count` | Engagement |
| `public_metrics.bookmark_count` | Engagement (X-only signal) |
| **`public_metrics.impression_count`** | **Impressions / views — available for ANY public post.** Enables view-based engagement rate |
| `created_at` | Posting activity over time, engagement over time |
| `entities.hashtags` | Most Engaged Hashtags |
| `attachments.media_keys` + `media.fields=type,public_metrics` | Post type mix (photo / video / GIF), **video `view_count` is public** |
| `attachments.poll_ids` | Poll detection (X-specific post type) |
| `referenced_tweets` | Distinguishes original posts vs replies vs reposts vs quotes |
| `text`, `id` | Post preview cards, deep link to the post |

**This is a richer competitor dataset than Meta gives us.** On Facebook and Instagram we cannot see a competitor's impressions at all; on X we can.

### 2.3 What we CANNOT get for a competitor

| Not available | Why |
|---|---|
| **Follower growth history** | X returns a *current* follower count only. Any growth chart must be built from **our own snapshots**, starting the day we begin tracking. No backfill, ever. |
| **Impressions, profile clicks, link clicks beyond `impression_count`** | `non_public_metrics` and `organic_metrics` require **OAuth user context as the post author** and only cover the last 30 days. Owner-only, by design. |
| **Audience demographics** (age, gender, country, interests) | Not exposed by the X API for any account, including your own. |
| **Video completion / playback funnel** for competitors | `playback_0_count` … `playback_100_count` are owner-only. Only total `view_count` is public. |
| **Protected / private accounts** | Invisible to app-only auth. Must be rejected at add-competitor time with a clear message. |
| **Reaction-type breakdown** | X has no love/haha/wow reactions. The Facebook "Post Reactions Distribution" widget has **no X equivalent** — the honest analog is an engagement-type split (likes / reposts / replies / quotes / bookmarks). |
| **Posts older than the 3,200 most recent** | Hard timeline ceiling, and in practice some high-volume accounts stop paginating far earlier (~100 results). |
| **Mentions of a competitor / share of voice / sentiment** | Needs the search endpoints: recent search is **7 days only**, full-archive is **Enterprise ($42k+/mo)**. Net-new API surface and net-new spend → out of v1. |
| **Ads / promoted posts** | Not in the public API. |
| **Stories/Reels-style ephemeral content, Spaces, Communities** | No equivalent public analytics surface. |

### 2.4 Pricing — what a sync actually costs

**X's 2026 pay-per-use rates (verified July 2026):**

| Operation | X's price |
|---|---|
| Post read | **$0.005** |
| User read / lookup | **$0.010** |
| Post created | $0.015 ($0.200 with a link) |
| Like / mute / block | $0.001 |
| Monthly post-read cap | **~2M posts per billing cycle** (one source says 3M — **verify in our developer portal**) |

**Critical billing mechanic:** X bills **per resource returned, not per request**. One call returning 100 posts is **100 reads**, not 1. There is no bulk discount for pagination.

**What that means for one competitor report sync.** Every figure below is for a **full report — 5 competitors plus the workspace's own profile, so 6 profile lookups**, not 5. The own profile is a billed lookup like any other unless we can reuse it from X platform analytics (see §7, open question 8).

| Report shape (5 competitors + you) | Raw X cost | Wallet price (+20% markup) | Legacy X credits (@ $0.0166/credit, rounded up) |
|---|---|---|---|
| 6 profiles × 20 posts (120 posts) | $0.66 | **$0.79** | **40 credits** |
| 6 profiles × 30 posts (180 posts) | $0.96 | **$1.15** | **58 credits** |
| 6 profiles × 50 posts (300 posts) | $1.56 | **$1.87** | **94 credits** |
| 6 profiles × 100 posts (600 posts) | $3.06 | **$3.67** | **185 credits** |

*(Formula: **(competitors + 1)** × [$0.010 user read + posts × $0.005]. The +1 is the workspace's own profile, which every widget compares against. Credit conversion follows the X analytics metering rule — cost basis converted at 1 credit = $0.0166, rounded up.)*

**Two findings that must shape the design:**

1. **Legacy-credit workspaces hit the wall on their first sync — even at a conservative default.** A full report at just **20 posts each costs 40 credits**, against a Standard plan's **45 X credits a month**: one sync consumes almost the whole allowance. At 30 posts it is **58 credits**, more than the allowance outright. Even an Agency plan (150/month) affords only two syncs at 30 posts. This is worse than a 5-profile estimate suggests, because the workspace's own profile is a sixth billed lookup.
2. **A daily cron is financially out of the question.** A full report synced daily = **~$35/month per report** at 30 posts, **~$56/month** at 50 posts — per report, per workspace, at the price the customer pays. Facebook and Instagram get a free daily cron; X cannot.

**Platform-level capacity risk:** the ~2M posts/month read cap is **per app** — it is ContentStudio's ceiling shared across every customer. At 250 post reads per report sync, that is roughly **8,000 competitor-report syncs per month platform-wide**, and X *publishing* and X *platform analytics* draw from the same cap. This needs a spending limit and a monitoring alarm, not just per-workspace metering.

---

## 3. Codebase analysis

> Frontend paths relative to `contentstudio-frontend/` (read from `origin/develop`; the mounted checkout is on `features`, **197 commits behind** — always read disputed files with `git show origin/develop:<path>`). Backend paths relative to `contentstudio-backend/` (91 commits behind). Go paths relative to `contentstudio-social-analytics-go/` (199 commits behind `origin/main`).

### 3.1 How competitor analytics is built today (three services)

**Frontend — `src/modules/analytics/`**
- Views: `views/competitor/MainAnalytics.vue`, `ShareAnalytics.vue`, `common/ReportsList.vue` (report tiles + the real empty state), `common/CompetitorAnalyticsLanding.vue` (the **plan gate**, not the empty state), `CompetitorUpgradeModal.vue`, `CompetitorDummyGraphs.vue`
- Per-platform report screens: `views/competitor/facebook/FacebookCompetitorReport.vue`, `views/competitor/instagram/InstagramCompetitorReport.vue` → **an X report screen is the third sibling**
- ~34 shared widget components in `components/competitor/`: `ManageCompetitorsModal.vue`, `CompetitorsTable.vue`, `ComparativeTableHelper.vue`, `FollowersGrowthComparison.vue`, `FollowersComparisonBarChart.vue`, `PerformanceComparison.vue`, `TopAndLeastPerformingPosts.vue`, `CompetitorTopLeastPosts.vue`, `HashtagTableHelper.vue`, `BioAnalysisTableHelper.vue`, `PostingActivityTableHelper.vue`, `SpecificPostTypeChart.vue`, `PostEngagementChart.vue`, `PostEngagementOverTime.vue`, `PostReactsDistributionChart.vue`, `AiInsightsCard.vue`, `ShareAnalyticsModal.vue`, `ManageSharedLinksDrawer.vue`, plus per-network post previews (`FacebookPublishedPostPreview.vue`, `InstagramPublishedPostPreview.vue`, **and an existing `LinkedinPublishedPostPreview.vue` — evidence a third network was anticipated**)
- Config: `composables/competitor/constants.ts` — `PLATFORMS = { INSTAGRAM, FACEBOOK, LABELSANDCAMPAIGNS }`, endpoint maps, response-key map. **Adding X starts here.**
- `ManageCompetitorsModal.vue` hard-caps a report at **5 competitors** (`competitors.length === 5`) and picks its search placeholder with a two-branch ternary on `platform === 'instagram'` — that ternary appears in several helpers (`useCompetitorHelper.ts` lines ~207, ~320) and each is a **third-platform seam**.

**Laravel — `contentstudio-backend/`**
- Controllers: `app/Http/Controllers/Analytics/Analyze/{Facebook,Instagram}CompetitorController.php` — CRUD for reports plus all the ClickHouse-backed read endpoints (`dataTableMetrics`, `followersGrowthComparison`, `topHashtags`, `biographyData`, `topAndLeastPerformingPosts`, `postingActivity*`)
- Routes: `routes/web/analytics.php:298-340`; the FB controller owns `POST search` → `searchCompetitor()`, which hits Facebook page search and returns `{competitor_id, name, image}`
- Repos/models: `CompetitorsRepo.php` (`addUpdateCompetitor` upserts by `competitor_id` + `platform_type`, then `triggerIgCompetitorJob` fires an immediate fetch), `CompetitorsReportsRepo.php`, `FbCompetitorRepo.php`, `IGCompetitorRepository.php`, models `CompetitorsModel`, `CompetitorsReportsModel`, `FbCompetitorModel`, `IgCompetitorsModel`
- Builders: `app/Builders/Analytics/Analyze/{Facebook,Instagram}CompetitorBuilder.php` (ClickHouse query construction) → **an X builder is the third sibling**

**Go — `contentstudio-social-analytics-go/`**
- Mongo: `src/models/db/mongo/competitors.go` — `Competitor` carries `platform_type`, so **the storage model already generalises**; `CompetitorReport` carries `platform_type` + `competitors[]`
- Cron: `src/cmd/competitor-jobs/main.go` supports `-socialNetwork=facebook|instagram` and `-syncType=incremental|full_refresh`, fans out Kafka work orders per competitor (`fetcher/facebook.go`, `fetcher/instagram.go`, topics `competitor-work-order-<network>-batch`)
- Work order: `src/models/kafka/competitor_work_order.go` — `{ReportID, PageID, Channel, Mode, StartDate, EndDate}`; `Channel` is documented as *"facebook or instagram"*
- Clients: `src/clients/social/{facebook,instagram}_competitor.go` → **`x_competitor.go` is the third sibling**, and it can reuse the existing `twitter.go` client's auth and pagination
- Analysis services: `src/services/{facebook,instagram}/*-competitor-analysis/`
- Parsers: `src/utils/parsing/{facebook_competitor_parser,instagram_competitor_parsing}.go`
- ClickHouse: `src/db/clickhouse/{facebook,instagram}_competitor.go`, models in `src/models/db/clickhouse/`, read queries in `src/db/clickhouse/analytics-get-queries/{fb,ig}_competitor/`
- API handlers: `src/api/analytics/{fb,ig}_competitor/handler.go` + `router.go`
- Thumbnails: `src/services/thumbnails/{facebook,instagram}_competitor*.go` + `url-refresher` jobs (media URLs expire; X media URLs are stable-ish but this path still applies to avatars)
- Reports/PDF: `src/services/reports/adapters/competitor/{adapter,facebook,instagram}.go`, `builder/widgets_competitor.go`, and the widget catalog `catalog/competitor.go`

### 3.2 The existing widget catalog — and X parity

From `src/services/reports/catalog/competitor.go`, the shipped competitor widget set and what X can support:

| Widget | FB | IG | **X — possible?** | Notes |
|---|---|---|---|---|
| `competitor_overview` (summary cards) | ✅ | ✅ | ✅ | Followers, posts, engagement, ER |
| `competitor_followers_table` (comparative table) | ✅ | ✅ | ✅ | Plus X-only columns: impressions, bookmarks, listed count |
| `competitor_followers_insights` (followers vs net change) | ✅ | ✅ | ⚠️ **Partial** | Net change only between our own snapshots — no history before tracking starts |
| `competitor_engagement_rate` | ✅ | ✅ | ✅ | ER by followers; **X bonus: ER by impressions** |
| `competitor_post_type` (posting activity by type) | ✅ | ✅ | ✅ | X types: text, image, video, GIF, link, poll, reply, repost, quote |
| `competitor_hashtags` (most engaged hashtags) | ✅ | ✅ | ✅ | From `entities.hashtags`, no extra cost |
| `competitor_page_bio` (bio analysis) | ✅ | ✅ | ✅ | Rich on X: bio, location, url, joined date, verified type, listed count |
| `competitor_performance_summary` (scatter) | ✅ | ✅ | ✅ | |
| `competitor_post_engagement` | ✅ | ❌ | ✅ | X *can* do what IG can't |
| `competitor_post_type_grid` | ✅ | ✅ | ✅ | |
| `competitor_activity_table_1..7` | ❌ | ✅ | ✅ | |
| `competitor_engagement_time_grid` | ✅ | ❌ | ✅ | X has per-post `created_at` + metrics |
| `competitor_reactions_1..8` (reaction distribution) | ✅ | ❌ | ❌ **No equivalent** | X has no reaction types. Analog: engagement-type split |
| `competitor_top_posts_1..6` / `competitor_least_posts_1..6` | ✅ | ✅ | ✅ | |

**Read:** X reaches parity on everything except reaction distribution, is *stronger* than Instagram on several widgets, and adds two X-only capabilities (impressions, bookmarks).

### 3.3 The X metering machinery this must plug into

From `docs/features/x-analytics-metering/` and `docs/features/x-pay-per-use-credits/`:

- **Two cohorts, decided by `WalletService::isEnabledForUser`:** dollar **X wallet** vs legacy **`x_posting_credits`**. Never mix currencies for one account.
- **Cost basis:** read from the wallet pricing config (`config/x_pay_per_use.php`), never hardcoded. Analytics **read rates** (`post_read`, `user_read`) are still flagged as **net-new config that must be added** — competitor analytics needs the same rates, so this is a shared dependency, not a second one.
- **Charge on success, on actual volume** (BR-2), atomic + idempotent (BR-4), insufficient balance blocks the sync rather than partial-syncing (BR-5, BR-14).
- **Opt-in per account, unlocked per workspace** (BR-6, BR-8), billing-capable users only via `can_see_subscription` / super admin (BR-7).
- **Never expose X's raw cost or our markup** (BR-10) — show only the price the user pays.
- The X sync UI precedent is **count-based, not date-based**: `SyncDateRangeModal.vue` in `mode="twitter"` shows a tweet-count dropdown `[10, 20, 30, 50, 80, 100, 120, 150]` (default 30), and `TwitterJobSettingsModal.vue` handles the recurring schedule. **Competitor sync should reuse both patterns verbatim.**

### 3.4 Integration points (where X plugs in)

| Layer | Change |
|---|---|
| FE constants | `PLATFORMS.X` in `composables/competitor/constants.ts` + an `X_API_ENDPOINTS` map |
| FE report screen | `views/competitor/x/XCompetitorReport.vue`, third sibling to FB/IG |
| FE add-competitor | `ManageCompetitorsModal.vue` — X branch: handle entry + validate, not typeahead search |
| FE sync UI | New for competitors: cost preview + sync trigger, reusing the X platform-analytics count dropdown and wallet-gate patterns |
| FE export | `utils/reportSections.ts` needs an `x_competitor` section list |
| Laravel | `XCompetitorController` + `XCompetitorBuilder` + routes under `overview/x/competitor`; handle-resolution endpoint replacing `searchCompetitor` |
| Laravel billing | Reuse the X wallet deduct/consume + ledger from the metering epic; add an `x_competitor_sync` consumption type |
| Go cron | `-socialNetwork=x` in `competitor-jobs`, `fetcher/x.go`, Kafka topic `competitor-work-order-x-batch` — **but scheduled off by default** |
| Go client | `src/clients/social/x_competitor.go`, reusing `twitter.go` auth/pagination |
| Go storage | ClickHouse `x_competitor` table + models + get-queries package |
| Go reports | `adapters/competitor/x.go` + `x_competitor` added to the catalog platform lists |
| Mongo | No schema change — `platform_type: "x"` already fits |

### 3.5 Technical considerations & gotchas

- **`competitor_id` is a string-or-int union** (`Competitor.CompetitorID interface{}`). X user IDs are large numeric strings — store as **string** and don't let them fall into the int64 branch.
- **Handle changes.** X handles are mutable; the numeric user ID is not. Key everything on the ID, display the handle, and re-resolve the handle on each sync (it comes free in the user read we already pay for).
- **Deleted/suspended/protected accounts** must set the existing `Competitor.State` / `Error` fields and surface a clear per-competitor error state rather than failing the whole report.
- **`impression_count` is 0 for posts published before December 2022** and occasionally missing — charts must handle nulls without showing a misleading zero.
- **Reposts inflate counts.** `GET /2/users/:id/tweets` returns retweets and replies unless excluded. Whether "posts" means original posts only is a **product decision with a direct cost consequence** — every returned repost is a billed read.
- **The 2M/month read cap is shared** with X publishing and X platform analytics. Needs a platform-level guard.
- **AI Insights** (`AiInsightsCard.vue`, `useAiInsights.ts`) exists on competitor reports — X should inherit it, and the prompt needs X-appropriate vocabulary (reposts, not shares).
- **Public share links** (`ShareAnalytics.vue`, `ManageSharedLinksDrawer.vue`, the `freetool.php` routes) also render competitor reports — an X report must work unauthenticated, and a shared link must **never** trigger a billed sync.
- **i18n:** 8 locales (`en, de, el, es, fr, it, pl, zh`), all keys in one commit.

---

## 4. Recommended approach for ContentStudio

### 4.1 The core recommendation

**Ship X competitor analytics as an on-demand, metered, count-based report that mirrors X platform analytics exactly.** Not a daily cron.

Concretely:
1. **Add competitors by handle**, validated on add (one $0.010 user read), not by typeahead search — a search-as-you-type box would bill a user read per keystroke-ish request.
2. **Sync is explicit.** A "Sync competitors" action on the X report, with the **same count dropdown** the X platform sync uses (posts *per competitor*), and the **estimated cost shown before confirming** — reusing the metering epic's preview + wallet-gate components.
3. **Optional scheduled sync**, off by default, with the projected recurring cost shown at setup and auto-pause when unfunded — the exact pattern `TwitterJobSettingsModal.vue` establishes.
4. **Charge the same wallet, same ledger, same cohort rule**, as a distinct `x_competitor_sync` consumption entry.
5. **Widget parity with Facebook**, minus reaction distribution, plus impressions and bookmarks.
6. **Be honest about follower history:** growth is measured from the first sync forward, stated plainly in the UI.

### 4.2 Why not a daily cron like FB/IG

At a full report (5 competitors + your own profile) × 30 posts, a daily cron costs ~$35/month per report; at 50 posts, ~$56. A workspace with three X reports would burn **$105–170/month** — likely more than their ContentStudio subscription. On-demand syncing puts the spend decision where the value is: the moment someone actually opens the report.

### 4.3 Suggested v1 scope

**In:**
- X as a platform option in competitor analytics (report create/edit/delete, up to 5 competitors compared against the workspace's own X profile — same cap and same shape as FB/IG)
- Add competitor by @handle with validation (exists, public, not suspended)
- On-demand metered sync with pre-sync cost preview and wallet/credit gating
- Report screen with parity widgets + X-only impressions & bookmarks
- Bio analysis, hashtags, top/least posts, posting activity by type, engagement over time, performance comparison
- PDF export with an `x_competitor` section list
- Public share links (read-only, never triggers a sync)
- AI insights on the X report

**Deferred to v2:**
- Scheduled/recurring competitor sync (if v1 ships on-demand only)
- Share of voice / mention volume (needs search endpoints + new spend)
- Best-time-to-post heatmap
- "Average of tracked profiles" benchmark row (Sprout's pattern)
- Cross-network competitor landscape (one report spanning FB + IG + X)

**Out:**
- Follower demographics, sentiment, ad detection, historical backfill before first sync

---

## 5. Plain-English summary: what we can and can't do

**We can show, for any public X competitor:** followers, following, listed count, lifetime post count, bio/location/join date/verified status, every post's likes, reposts, replies, quotes, bookmarks, **impressions/views**, hashtags used, post type (text/image/video/GIF/poll/link), when they post, top and worst performing posts, engagement rate by followers **and by impressions**.

**We cannot show:** how their followers grew before we started tracking, who their followers are (age/gender/country), their profile clicks or link clicks, video completion rates, anything about private accounts, any reaction-type breakdown (X has none), posts older than their most recent ~3,200, who mentions them, or anything about their ads.

**It costs:** $0.005 per post we read, $0.010 per profile we look up — X bills per *item returned*, so 100 posts = 100 charges. A full report is **6 profiles**: the five competitors plus the workspace's own profile, which every comparison is measured against. At 30 posts each that costs us **$0.96**, which is **$1.15** to the customer at our standard markup, or **~58 X credits** for legacy-credit workspaces — **more than a Standard plan's entire monthly allowance of 45**. Even the gentler 20-post default is **$0.79 / 40 credits**.

---

## 6. Dependencies & sequencing

1. **Hard dependency — X Pay-Per-Use Credit Wallet** (`docs/features/x-pay-per-use-credits/`): balance, atomic deduction, ledger, pricing config, top-up, permission model.
2. **Hard dependency — X Analytics Metering** (`docs/features/x-analytics-metering/`): the `post_read` / `user_read` rates in the pricing config, the cost-preview and wallet-gate UI components, the cohort currency rule, the unlock/consent pattern. Competitor analytics is the **second consumer** of all of it — the goal is to add zero new billing concepts.
3. **Soft dependency — Competitor Analytics Revamp (FB + IG)** (`docs/features/competitor-analytics-revamp/`): the revamped empty state, Add Competitor modal, report screen and export modal. If that lands first, X should be built on the revamped components, not the old ones. **Sequencing decision needed from the PO.**
4. **External:** X API pay-per-use rates and the monthly read cap; our developer-portal spending limit.

---

## 7. Open questions for the review gate

| # | Question | Why it matters |
|---|---|---|
| 1 | **On-demand only in v1, or on-demand + optional schedule?** | Scheduling roughly doubles the BE/FE work and adds the auto-pause path. On-demand alone is shippable and safe. |
| 2 | **Do we count reposts and replies as "posts"?** | Every repost returned is a billed read. Excluding them is cheaper and arguably more meaningful, but changes what "posting activity" means vs FB/IG. |
| 3 | **Default post count per competitor** — 30 like platform analytics, or lower? | At 6 profiles, 30 posts each = **58 credits**, well over a Standard plan's 45/month allowance; even 20 posts is **40 credits**. A default of 10 (**22 credits**) is the only setting that leaves a legacy-credit workspace room to sync twice. |
| 4 | **Do legacy X-credit workspaces get competitor analytics at all in v1?** | Their allowances (45–150/month) barely cover one or two syncs. Options: allow with clear warnings, or wallet-only for competitor analytics. |
| 5 | **Is X competitor analytics gated by the same plan gate as FB/IG competitors, or does it also need the X analytics unlock?** | Two gates stacked could be confusing; one gate could mean unmetered surprise spend. |
| 6 | **Build on the revamped FB/IG components or the current ones?** | Depends on the revamp's landing date. |
| 7 | **Show competitor impressions prominently?** | It's a real differentiator, but it invites "why can't I see this for Facebook competitors?" |
| 8 | **Can we reuse the workspace's own X profile from platform analytics instead of looking it up again on every competitor sync?** | The own profile is the 6th billed lookup in every sync. If its follower count can be read from the platform-analytics sync we already run, a full report drops from **$0.79 to $0.66** (40 credits to 34) and the baseline row becomes free. Needs a check on how fresh the platform-analytics snapshot is. |
| 8 | **Per-workspace or per-account monthly spend cap for competitor syncs?** | Protects both the customer and our shared 2M-read platform cap. |

---

## 8. Sources

- [X API pay-per-usage pricing and credits — X Developer Docs](https://docs.x.com/x-api/getting-started/pricing)
- [Metrics — X Developer Docs](https://docs.x.com/x-api/fundamentals/metrics)
- [X API Rate Limits — X Developer Docs](https://docs.x.com/x-api/fundamentals/rate-limits)
- [Get Posts (`/2/users/:id/tweets`) — X Developer Docs](https://docs.x.com/x-api/users/get-posts)
- [How Much Does the X (Twitter) API Cost in 2026? — twitterapi.io](https://twitterapi.io/blog/x-api-cost-breakdown-2026)
- [X (Twitter) API Pricing in 2026: All Tiers — Postproxy](https://postproxy.dev/blog/x-api-pricing-2026/)
- [X (Twitter) API Pricing 2026: Tiers, Free Tier & Real Costs — Sorsa](https://api.sorsa.io/blog/twitter-api-pricing-2026)
- [The X (Twitter) API in 2026: Pricing, Rate Limits & What Still Works — SocialCrawl](https://www.socialcrawl.dev/blog/x-twitter-api-2026)
- [X (Twitter) API Rate Limits in 2026: Every Endpoint, Explained — Sorsa](https://api.sorsa.io/blog/twitter-api-rate-limits-2026)
- [What's included in the X (formerly Twitter) Competitors Report? — Sprout Social Support](https://support.sproutsocial.com/hc/en-us/articles/202604313-What-s-included-in-the-X-formerly-Twitter-Competitors-Report)
- [How to Do a Twitter Competitor Analysis — Sprout Social](https://sproutsocial.com/insights/twitter-competitor-analysis/)
- [Twitter (X) Analytics in 2026 — Metricool](https://metricool.com/analytics-twitter/)
- [12 social media competitor analysis tools — Hootsuite Blog](https://blog.hootsuite.com/social-media-competitor-analysis-tools/)
- [Rival IQ — Competitive Social Media Analytics](https://www.rivaliq.com/)
- [Socialinsider — Sprinklr alternatives / competitor analytics coverage](https://www.socialinsider.io/blog/sprinklr-alternatives/)
- Internal: `docs/features/x-analytics-metering/` (01-research, 03-prd), `docs/features/x-pay-per-use-credits/`, `docs/features/competitor-analytics-revamp/01-research.md`
- Codebase: `contentstudio-frontend@origin/develop`, `contentstudio-backend@origin/develop`, `contentstudio-social-analytics-go@origin/main`
