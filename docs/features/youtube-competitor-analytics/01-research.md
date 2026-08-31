# 01 — Research: YouTube in Competitor Analytics

**Date:** 2026-08-20
**Feature:** Add YouTube as a platform in the Analytics → Competitors module, alongside Facebook, Instagram, and the in-flight X work.

---

## 0. What this feature is (and the one thing that makes it different)

Competitor analytics lets a workspace build a report of up to **5 competitors** compared against **the workspace's own connected profile**, which is included automatically and never occupies one of the 5 slots. A full report is therefore **6 profiles**. This feature adds **YouTube** as a platform in that same module.

YouTube is the opposite problem from X. X was a **money** problem: the data is rich but every read costs cash. YouTube is a **permission and policy** problem: the API is free, but almost everything that makes YouTube analytics valuable is visible **only to the channel owner**, and what we may store is capped by Google's developer policies.

| | Facebook / Instagram | X | **YouTube** |
|---|---|---|---|
| Data source | Meta Graph API | X API v2 | **YouTube Data API v3** |
| Cost per read | Free | $0.005 per post, $0.010 per user | **Free** |
| Hard limit | Rate limits | Your wallet balance | **10,000 quota units/day, cannot be bought** |
| Sync model | Daily cron | On-demand, metered | **Daily cron is viable** |
| Metering needed | No | Yes | **No** |
| Binding constraint | None | Cost | **Owner-only metrics + 30-day data retention policy** |

So the design question for YouTube is not "how do we avoid burning money", it is: **what is actually left once you remove everything that requires being the channel owner — and is that still worth shipping?** The answer is yes, but the gap between our own-channel YouTube report and a YouTube competitor report is the widest of any platform we support, and the UI has to be honest about it.

---

## 1. Competitor & industry research

### 1.1 The market context

YouTube competitor benchmarking is a mature category, and every serious tool ships it. The consistent pattern across Rival IQ, Sprout Social, Hootsuite, and Brandwatch is that they all show the **same narrow set** of channel-level public metrics — subscribers, views, uploads, likes, comments, engagement rate — because that is all the API gives anyone. Nobody shows competitor watch time or retention, because nobody can.

| Tool | What it shows for YouTube competitors | Entry price |
|---|---|---|
| **Rival IQ** | Subscriber growth, video views, engagement rate, post frequency, plus automated alerts when a competitor's video spikes | ~$239/mo |
| **Sprout Social** | Competitor channels tracked for post frequency, subscriber growth, engagement rate, views, likes, comments | ~$249/seat/mo |
| **Hootsuite** | Competitor metrics inside custom cross-network reports | ~$99/mo |
| **Brandwatch** | YouTube as one source in broad social listening, with AI trend detection | Enterprise |

### 1.2 Key findings

1. **The public metric ceiling is the same for everyone.** Our competitive risk is not data access, it is presentation and integration — a YouTube competitor report sitting next to Facebook, Instagram, and X reports in one module, with one export.
2. **Everyone leans on subscriber growth, and everyone is quietly imprecise about it,** because the API rounds subscriber counts. Rival IQ publishes a help article explaining the rounding to customers. We should be upfront in the UI rather than in a help doc.
3. **Nobody exploits video tags.** `snippet.tags` is public, and it is the direct analog of the hashtag widget we already ship for Facebook, Instagram, and X. A "most engaged tags" comparison across six channels is a real differentiator that costs nothing extra.
4. **Nobody compares video length or Shorts mix well.** Video duration is public, so Shorts versus long-form strategy comparison is available and genuinely actionable.
5. **Alerting is the recurring premium feature** (Rival IQ's differentiator). Out of scope for v1 but worth noting.

### 1.3 Table stakes vs delighters

- **Table stakes:** subscriber comparison, view comparison, upload cadence, engagement rate, top and least performing videos, channel profile comparison.
- **Delighters:** most engaged tags, Shorts versus long-form mix, video length distribution, content category mix.
- **Cannot be built at any price:** competitor watch time, retention, traffic sources, demographics, impressions, CTR.

---

## 2. YouTube API deep dive — what is actually allowed

> This is the section the PO asked for in plain words: what we *can* build, what we *cannot*, and what the limits are.

### 2.1 The three endpoints we already use — and they already work for competitors

ContentStudio's YouTube analytics fetches through the Data API v3 for content and the **YouTube Analytics API v2** for performance. Confirmed in the Go client (`src/clients/social/youtube.go`):

| Call | Current params | Works for a channel we don't own? |
|---|---|---|
| `GET /youtube/v3/channels` (`part=id,snippet,statistics,brandingSettings,contentDetails`) | `mine=true` | **Yes** — swap `mine=true` for `id=` or `forHandle=` |
| `GET /youtube/v3/playlistItems` (`part=snippet,contentDetails`, the channel's uploads playlist) | `playlistId=<uploads>` | **Yes** — every public channel exposes an uploads playlist |
| `GET /youtube/v3/videos` (`part=id,snippet,contentDetails,statistics`, batched 50 IDs) | `id=<batch>` | **Yes** |
| `GET /youtubeanalytics/v2/reports` | `ids=channel==MINE` | **No. Owner only, and this is where all the good metrics live.** |

**The competitor pipeline is the first three calls with a different filter.** No new API surface, no new credentials, no new client. The fourth call is simply absent from a competitor sync, and that absence defines the whole feature.

`channels.list` also accepts **`forHandle`** (e.g. `@GoogleDevelopers`), which is exactly the add-a-competitor path — and unlike X, it is free.

### 2.2 What we CAN get for a competitor (public, any channel)

**Channel level** — `channels.list`:

| Field | Feeds which widget |
|---|---|
| `statistics.subscriberCount` | Subscribers, comparative table — **rounded, see §2.4** |
| `statistics.videoCount` | Lifetime video count, channel profile |
| `statistics.viewCount` | Lifetime views, channel profile |
| `snippet.title`, `description`, `publishedAt`, `country`, `thumbnails`, `customUrl` | Channel profile analysis, competitor tile |
| `brandingSettings.channel.keywords` | Channel keyword comparison |
| `contentDetails.relatedPlaylists.uploads` | The uploads playlist we page for videos |

**Video level** — `videos.list`:

| Field | Feeds which widget |
|---|---|
| `statistics.viewCount` | Views comparison, top/least videos |
| `statistics.likeCount` | Engagement, engagement split |
| `statistics.commentCount` | Engagement, engagement split |
| `snippet.publishedAt` | Upload cadence, engagement over time |
| `snippet.tags` | **Most engaged tags — public, confirmed** |
| `snippet.categoryId` | Content category mix |
| `snippet.title`, `description`, `thumbnails` | Video cards — **30-day retention, see §2.4** |
| `snippet.liveBroadcastContent` | Live and premiere detection |
| `contentDetails.duration` | **Shorts vs long-form, video length distribution** |
| `contentDetails.definition`, `caption` | Quality and captions availability |

**Derived from the above:** engagement rate by views, engagement rate by subscribers, average views per video, likes per view, comments per view, uploads per week, video length distribution, Shorts share, category mix, day-of-week and time-of-day upload pattern, tag overlap between channels.

### 2.3 What we CANNOT get for a competitor

Everything below comes from the YouTube Analytics API, which only ever answers for `channel==MINE`.

| Not available | Why |
|---|---|
| **Watch time / estimated minutes watched** | Owner only. Our own-channel report shows this; a competitor report cannot. |
| **Average view duration, average view percentage** | Owner only. |
| **Audience retention curves** | Owner only. |
| **Impressions and click-through rate** | Owner only, and not in the Data API at all. |
| **Traffic sources** (search, suggested, browse, external) | Owner only. |
| **Demographics** — age, gender, country, device | Owner only. |
| **Subscribers gained / lost per day** | Owner only. Only a current, rounded total is public. |
| **Shares** | Owner only. |
| **Dislikes** | **Made private on 13 December 2021.** Returned only to the video owner. Our own-channel report still shows a Dislikes card, so this must be hidden in competitor comparisons to avoid an asymmetric table. |
| **Revenue, Premium (`redViews`) metrics** | Owner only. |
| **Exact subscriber count** | Rounded for everyone but the owner in YouTube Studio. |
| **Unlisted and private videos** | Not returned publicly, so a competitor's upload count can undercount deliberately unlisted content. |
| **Custom thumbnail flag** | Owner only. |
| Comment *text* and sentiment | Technically fetchable via `commentThreads` at 1 unit per call, but comment text is capped at **30-day retention** and it is a net-new data category. Out of v1. |

**The honest framing for the PO:** our own YouTube channel report has roughly 17 widget groups including watch time, traffic sources, and full demographics. A YouTube competitor report can support around 15 — but a *different* 15, with none of the retention or discovery insight. It compares **what a channel published and how audiences reacted publicly**, not **how YouTube distributed it**.

### 2.4 The three real constraints

There is no per-read price to model. Instead there are three limits, and each one shapes the UI.

#### (a) Quota — free, finite, and not purchasable

Every read costs quota units against a **default 10,000 units per day per Google Cloud project**, resetting at midnight Pacific. Costs are per call, not per item returned:

| Call | Units |
|---|---|
| `channels.list` | **1** |
| `playlistItems.list` | **1** |
| `videos.list` | **1** |
| `search.list` | **100 — avoid entirely** |

**Cost of one competitor report sync** (6 channels, 50 videos each):

| Step | Calls | Units |
|---|---|---|
| Channel stats for all 6 channels (batched, up to 50 IDs per call) | 1 | 1 |
| Uploads playlist pages, 50 videos per call per channel | 6 | 6 |
| Video details, 300 IDs batched 50 per call | 6 | 6 |
| **Total** | **13** | **13** |

At 13 units a sync, the default allowance supports roughly **770 report syncs a day platform-wide** — and that allowance is **shared with YouTube publishing and our own-channel YouTube analytics**. A daily cron across, say, 500 workspaces with two reports each would need ~13,000 units/day and would blow the quota.

**There is no paid tier.** More quota requires the **YouTube API Services Audit and Quota Extension Form**, a manual compliance review that takes weeks to months. Splitting load across multiple Cloud projects violates the Terms of Service.

⚠️ **Risk to raise now:** public reporting is that quota-extension requests framed as scrapers, bulk harvesting, or **competitive analytics** are frequently rejected. However, the **derived-metrics policy explicitly permits** ranking channel performance and tracking views between different channels for comparison and historical analysis. So the application must be made as an **Analytics & Reporting creator tool**, which is what we are, and not as a competitive-intelligence scraper. **This application is the critical path item for the whole feature and should start before engineering does.**

#### (b) Data retention — 30 days by default, 36 months only with an amendment

Google's developer policies require an API client to **delete or refresh stored YouTube data after 30 calendar days**. That alone would cap every trend chart at a 30-day window.

The **additional policies for derived metrics and data storage** amendment changes this, and it is the difference between a real product and a toy:

| Data | Retention with the amendment |
|---|---|
| Metrics from statistical endpoints — **views, likes, subscriber counts, comment counts** | **Up to 36 calendar months** |
| Derived metrics, including sentiment | Up to 36 months |
| **Titles, channel names, descriptions, comment text** | **Still 30 days** |

Consequences for the design:
- Long-run subscriber and view trends are allowed — but **only after the amendment is accepted**, via the same form as the quota extension (Section 5 → Analytics & Reporting).
- **Top and least performing video cards must refresh their titles, descriptions, and thumbnails within 30 days**, or re-fetch them on view. Metrics may persist; the words may not.
- Derived metrics such as engagement rate **must be visually distinguished from API-sourced metrics**. Our charts should mark computed columns.
- Comparing and ranking channels is explicitly allowed, but the presentation must not frame it as a rivalry or invite brigading. "Competitors" framing is fine; leaderboard-style taunting is not.

#### (c) Subscriber counts are rounded down to three significant figures

A channel with 123,456 subscribers reports **123,000**. Above 1,000 subscribers, nobody but the owner sees a precise number.

This breaks the naive "net change in subscribers" widget. For a 123,000-subscriber channel, the smallest change the API can express is **1,000 subscribers** — so day-to-day movement reads as a flat line punctuated by sudden jumps. Two design consequences:

1. Subscriber change must be presented over ranges long enough to exceed the rounding step, and labelled as approximate.
2. Small channels are the exception: under 1,000 subscribers the count is exact, so the widget behaves differently at the bottom of the range. The UI must not look broken in either case.

---

## 3. Codebase analysis

> Frontend paths relative to `contentstudio-frontend/` (read from `origin/develop`). Go paths relative to `contentstudio-social-analytics-go/`. Always check the mounted checkout is not behind: `git log HEAD..origin/develop | wc -l`.

### 3.1 What YouTube analytics already ships

**Go — the platform report.** `src/services/reports/catalog/youtube.go` registers a rich own-channel report: `overview_cards`, `yt_subscriber_trend`, `yt_subscriber_daily`, `yt_views_trend`, `yt_watch_time_trend`, `yt_engagement_trend`, `yt_engagement_vs_schedule`, `yt_views_vs_schedule`, `yt_traffic_sources`, `yt_video_sharing`, `yt_demographics`, four top/least video widgets split by engagement and views, and three AI insight widgets.

Overview cards, from `src/services/reports/adapters/youtube/adapter.go`: Subscribers, Videos, Video Views, Watch Time, Average View Duration, Engagements, Likes, **Dislikes**, Shares, Comments. **Five of those ten cards are owner-only** and cannot appear in a competitor report.

**Go — the client.** `src/clients/social/youtube.go` already has everything the competitor pipeline needs: `FetchChannels`, `FetchVideos` (uploads playlist paging), `FetchVideoDetails` (50-ID batching), plus `IsShortByDuration` and `ParseISO8601Duration` — meaning **Shorts detection already exists and works off public data**. The owner-only calls (`FetchActivityInsights`, `FetchTrafficInsights`, `FetchSharedInsights`, `FetchDemographics`, `FetchAllVideosAnalytics`) are simply not used by a competitor sync.

**Frontend.** `src/modules/analytics/views/youtube/` holds the full platform dashboard — cards, overview, posts, demographics, AI insights tab, and a Top/Least × Engagement/Views toggle on top videos worth mirroring for competitors.

### 3.2 What competitor analytics is built from

Unchanged from the X research, and it generalises the same way:

- **Frontend:** `views/competitor/{facebook,instagram}/` report screens are siblings — a YouTube screen is the next one. `composables/competitor/constants.ts` holds `PLATFORMS = { INSTAGRAM, FACEBOOK, LABELSANDCAMPAIGNS }` and the endpoint maps; **adding YouTube starts here**. `ManageCompetitorsModal.vue` hard-caps 5 competitors and branches on `platform === 'instagram'` in several helpers — each is a seam.
- **Laravel:** `Analytics/Analyze/{Facebook,Instagram}CompetitorController.php` plus matching builders and repos. A YouTube controller and builder are the next siblings. `CompetitorsRepo::addUpdateCompetitor` upserts by `competitor_id` + `platform_type`.
- **Go:** `src/models/db/mongo/competitors.go` already carries `platform_type`, so **storage generalises with no schema change**. `src/cmd/competitor-jobs/main.go` takes `-socialNetwork=facebook|instagram` and fans out Kafka work orders — YouTube is a third value, and unlike X it **can** be scheduled on. Clients `{facebook,instagram}_competitor.go` → `youtube_competitor.go`, reusing the existing YouTube client.
- **Reports:** `src/services/reports/catalog/competitor.go` is the widget source of truth, keyed by `facebook_competitor` / `instagram_competitor` platform strings. `utils/reportSections.ts` on the frontend maps export sections to those widget IDs.

### 3.3 Widget parity against the existing competitor catalog

| Widget | FB | IG | X | **YouTube** | Note |
|---|---|---|---|---|---|
| `competitor_overview` | ✅ | ✅ | ✅ | ✅ | Subscribers, views, engagement rate |
| `competitor_followers_table` | ✅ | ✅ | ✅ | ✅ | Subscribers rounded; adds views columns |
| `competitor_followers_insights` | ✅ | ✅ | ⚠️ | ⚠️ **Degraded** | Rounding, not missing history, is the problem here |
| `competitor_engagement_rate` | ✅ | ✅ | ✅ | ✅ | By subscribers **and** by views |
| `competitor_post_type` | ✅ | ✅ | ✅ | ✅ | Types are Shorts / long-form / live |
| `competitor_hashtags` | ✅ | ✅ | ✅ | ✅ | Repurposed for **video tags** |
| `competitor_page_bio` | ✅ | ✅ | ✅ | ✅ | Channel profile + keywords |
| `competitor_performance_summary` | ✅ | ✅ | ✅ | ✅ | |
| `competitor_post_engagement` | ✅ | ❌ | ✅ | ✅ | |
| `competitor_post_type_grid` | ✅ | ✅ | ✅ | ✅ | |
| `competitor_activity_table_1..7` | ❌ | ✅ | ✅ | ✅ | |
| `competitor_engagement_time_grid` | ✅ | ❌ | ✅ | ✅ | |
| `competitor_reactions_1..8` | ✅ | ❌ | ❌ | ❌ | No reactions on YouTube either |
| `competitor_top_posts` / `least_posts` | ✅ | ✅ | ✅ | ✅ | Should split by views **and** engagement |
| **Views comparison** | ❌ | ❌ | ❌ | 🆕 | YouTube's analog of X impressions |
| **Video length distribution** | ❌ | ❌ | ❌ | 🆕 | YouTube only |
| **Content category mix** | ❌ | ❌ | ❌ | 🆕 | YouTube only |

### 3.4 Technical considerations & gotchas

- **`competitor_id` is a string-or-int union.** YouTube channel IDs are strings starting `UC…` — no int branch risk, unlike X, but store as string.
- **Handles are mutable, channel IDs are not.** Key on the `UC…` ID, display the handle, re-resolve on each sync (free).
- **Shorts definition drifts.** The Shorts duration ceiling has moved before; put the threshold in config rather than hardcoding it.
- **`likeCount` can be hidden by a creator.** Treat a missing like count as unknown, not zero.
- **`commentCount` is absent when comments are disabled.** Same rule.
- **Terminated, private, or renamed channels** must set the existing `Competitor.State` / `Error` fields and surface a per-channel state rather than failing the report.
- **Thumbnail URLs are stable but titles are not, and the 30-day text-retention rule applies.** The existing `url-refresher` thumbnail jobs are the right hook.
- **Quota exhaustion is a 403, not a 429.** The existing YouTube client already conflates the two; a competitor sync must distinguish "we are out of quota today" from "this channel is gone", because the first is our fault and should not mark a competitor broken.
- **Public share links** render competitor reports unauthenticated and must not trigger a sync.
- **AI insights** exist on competitor reports and on the YouTube platform report; the YouTube competitor prompt needs vocabulary that never references watch time or retention, since it has neither.
- **i18n:** 8 locales, all keys in one commit.

---

## 4. Recommended approach for ContentStudio

### 4.1 The core recommendation

**Ship YouTube competitor analytics on the Facebook and Instagram model, not the X model.** The data is free, so it gets a **daily cron**, no wallet, no cost preview, no metering, and no per-sync gating. That makes it substantially cheaper to build than X competitor analytics.

Concretely:
1. **Add competitors by handle or channel URL**, resolved for free, validated on add.
2. **Daily incremental sync via the existing competitor cron**, with a manual refresh available.
3. **Widget parity with Facebook**, minus reaction distribution, plus three YouTube-only widgets: views comparison, video length distribution, and content category mix.
4. **Repurpose the hashtag widget for video tags** — public, free, and nobody else does it.
5. **Be honest in the UI** about rounded subscriber counts and about the metrics that exist on your own channel report but cannot exist here.
6. **Apply for the quota extension and the derived-metrics amendment before engineering starts.** Without the amendment, no chart can look back further than 30 days.

### 4.2 Suggested v1 scope

**In:** YouTube as a competitor platform with report CRUD and the 5-competitor cap; add by handle or channel URL; daily cron plus manual refresh; the widget set in §3.3 including the three YouTube-only widgets and tags; top and least videos split by views and engagement; channel profile comparison; PDF export with a `youtube_competitor` section list; public share links; AI insights.

**Deferred to v2:** competitor spike alerting (Rival IQ's differentiator); comment volume and sentiment; upload time-of-day heatmap; playlist strategy comparison; cross-network competitor landscape.

**Out:** everything owner-gated — watch time, retention, traffic sources, demographics, impressions, CTR, dislikes, revenue — plus historical backfill before the first sync.

---

## 5. Plain-English summary: what we can and can't do

**We can show, for any public YouTube channel:** subscribers (rounded to three significant figures), lifetime views, lifetime video count, channel description, country, join date, channel keywords, and then for each recent video its views, likes, comments, publish time, tags, category, duration, whether it is a Short or long-form or a live broadcast, and whether it has captions. From those we can derive engagement rate by views and by subscribers, average views per video, upload cadence, video length strategy, category mix, tag overlap, and top and least performing videos by either views or engagement.

**We cannot show:** watch time, average view duration, audience retention, impressions, click-through rate, how viewers found the videos, who the audience is, subscribers gained and lost day by day, shares, dislikes, revenue, exact subscriber counts, or anything about unlisted and private videos. All of it is restricted to the channel owner, and no amount of quota or money unlocks it.

**It costs:** nothing in cash. One report sync is about **13 quota units** out of a **10,000-per-day** allowance shared with YouTube publishing and our own-channel analytics — roughly 770 syncs a day platform-wide before we need Google's approval for more. That approval, and the separate permission to keep competitor metrics longer than 30 days, are both applications to Google that take weeks to months and gate the feature.

---

## 6. Dependencies & sequencing

1. **Hard, external, and slow — Google approvals.** The quota extension and the derived-metrics storage amendment are submitted on the same form. Without the amendment, trend charts are capped at 30 days. **Start this first; it is the critical path.**
2. **Soft — Competitor Analytics Revamp (Facebook + Instagram).** If it lands first, YouTube should be built on the revamped empty state, Add Competitor modal, report screen, and export modal.
3. **Soft — X Competitor Analytics.** X introduces the third-platform seams in the competitor module (platform switcher, per-platform report screen, per-platform export section list). YouTube built after X is materially cheaper, because it becomes the fourth platform through doors X already opened. Sequencing decision needed from the PO.
4. **None on the X wallet or metering epics.** YouTube needs no billing integration at all.

---

## 7. Open questions for the review gate

| # | Question | Why it matters |
|---|---|---|
| 1 | **Do we submit the Google quota and derived-metrics application now, before scoping stories?** | It takes weeks to months and gates the trend charts. Everything else is buildable in parallel, but shipping is blocked on it. |
| 2 | **If the derived-metrics amendment is refused, do we still ship with 30-day windows only?** | A 30-day-capped competitor report is still useful but much weaker. Worth deciding the fallback now. |
| 3 | **How many recent videos per channel do we sync?** | No cash cost, but it drives quota and storage. 50 per channel is 13 units a sync; 200 per channel is about 30. |
| 4 | **How do we present rounded subscriber counts?** | Options: show the rounded number plainly, show an approximate indicator, or suppress net-change entirely below a threshold. This is the single most visible honesty decision in the feature. |
| 5 | **Do we hide the Dislikes card on our own channel inside competitor comparisons?** | Leaving it in produces a table where one column only ever has one value filled. |
| 6 | **Build after X competitor analytics, or in parallel?** | X opens the third-platform seams. Building in parallel means two teams touching the same modal and switcher. |
| 7 | **Do we add a manual refresh button, or is the daily cron enough?** | Free data means a refresh button is harmless to the customer but does consume shared quota. |
| 8 | **Is YouTube competitor analytics gated by the same plan gate as Facebook and Instagram competitors?** | No billing gate is needed, so the plan gate is the only lever. |

---

## 8. Sources

- [Channels: list — YouTube Data API](https://developers.google.com/youtube/v3/docs/channels/list)
- [Videos — YouTube Data API resource reference](https://developers.google.com/youtube/v3/docs/videos)
- [Revision History — YouTube Data API](https://developers.google.com/youtube/v3/revision_history)
- [Quota and Compliance Audits — YouTube Data API](https://developers.google.com/youtube/v3/guides/quota_and_compliance_audits)
- [Additional policies for derived metrics and data storage — YouTube](https://developers.google.com/youtube/terms/derived-metrics-policy)
- [YouTube API Services — Developer Policies](https://developers.google.com/youtube/terms/developer-policies)
- [YouTube API Quota Limits 2026: 10,000 Units, Costs & How to Get More — Phyllo](https://www.getphyllo.com/post/youtube-api-limits-how-to-calculate-api-usage-cost-and-fix-exceeded-api-quota)
- [YouTube API Quota Increase: How The Audit Works (2026)](https://singhamandeep.com/youtube-data-api-quota-increase-audit/)
- [Why YouTube Subscriber Counts are Rounded — Rival IQ](https://help.rivaliq.com/en/articles/9788197-why-youtube-subscriber-counts-are-rounded)
- [13 YouTube Analytics Tools You Need in 2026 — Sprout Social](https://sproutsocial.com/insights/youtube-analytics-tools/)
