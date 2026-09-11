# Research: Competitor analytics tools for the assistant, and the rest of the analytics tool gap

**Date:** 2026-09-11
**Codebase:** `contentstudio-ai-agents` (internal toolkit layer at `src/integrations/contentstudio/`)

---

## Current State

### The analytics tool surface today

`AnalyticsToolkit` (`src/integrations/contentstudio/toolkits/analytics.py`, 983 lines) registers
**6 tools** on the "Analytics Analyst" team member:

| Tool | What it reaches |
|---|---|
| `analytics_fetch_metrics` | `{platform}/summary` ×8, plus 18 date-bucketed series endpoints when `over_time=True` |
| `analytics_compare_accounts` | same summary endpoints, ranked |
| `analytics_top_content` | `top-posts` / `top-pins` / `top-tweets` ×8 |
| `analytics_insights` | `{platform}/ai-insights` ×7 (all but X) |
| `analytics_best_times` | `active-users` ×2 (Facebook, Instagram only) |
| `analytics_list_available_metrics` | the generated metric registry, no network call |

The metric registry (`analytics_registry.py`) is generated from the committed public API spec
(`api-docs.json`) and carries 96 metric descriptors plus 18 `SeriesDescriptor` entries.

### The committed spec is stale, and that is the headline finding

`api-docs.json` in the ai-agents repo has 148 paths / **104 analytics endpoints**, parsed
2026-07-31 and re-verified 2026-08-19. Counted against the live routes in
`contentstudio-backend/routes/api/v1.php` today, the public API carries **181 analytics
endpoints**, plus 6 competitor-report management routes:

| Group | Endpoints | In the committed spec? |
|---|---|---|
| facebook | 15 | yes |
| instagram | 15 | yes |
| youtube | 20 | 19 of 20 — `publishing-behaviour` is new |
| pinterest | 14 | yes |
| linkedin | 11 | yes |
| gmb | 10 | yes |
| tiktok | 8 | yes |
| twitter | 7 | yes |
| campaigns-labels | 5 | yes |
| **facebook/competitor** | **13** | **no** |
| **instagram/competitor** | **9** | **no** |
| **threads** | **16** | **no** |
| **google-ads** | **17** | **no** |
| **meta-ads** | **11** | **no** |
| **bluesky** | **10** | **no** |
| **TOTAL** | **181** | 104 known |

So **77 endpoints landed after the registry was last generated**, including two entire platforms
(Threads, Bluesky), both ads families, and all of competitor.

**Reachable from chat: 43 of 181, about a quarter.**

### Competitor analytics is live on the public API

Confirmed in `contentstudio-backend/routes/api/v1.php`:

- **Report management (6):** list, show, create, update, delete competitor reports, plus
  competitor search. A competitor report is the saved *set* of competitor pages, and its id is what
  every read below requires.
- **Facebook competitor reads (13):** `data-table-metrics`, `posting-activity-graph-by-types`,
  `posting-activity-by-specific-type`, `post-react-distribution`,
  `post-react-distribution-by-company`, `post-type-distribution`,
  `top-and-least-performing-posts`, `top-hashtags`, `individual-hashtag-data`, `biography-data`,
  `followers-growth-comparison`, `post-engagement-over-time`, `post-engagement-by-competitor`
- **Instagram competitor reads (9):** `data-table-metrics`, `posting-activity-graph-by-types`,
  `posting-activity-by-specific-type`, `posting-activity-table-by-type`,
  `top-and-least-performing-posts`, `top-hashtags`, `individual-hashtag-data`, `biography-data`,
  `followers-growth-comparison`

Both read groups sit behind a `CompetitorReportBelongsToWorkspace` middleware applied at the group
level, so tenant scoping on `competitor_report_id` is enforced for every endpoint including any
added later. The underlying data comes from the Go analytics service
(`contentstudio-social-analytics-go`, handlers at `src/api/analytics/fb_competitor/` and
`ig_competitor/`).

**There is no blocker.** The earlier reading that competitor was unavailable came from the stale
committed spec; the capability shipped since.

### The ads non-goal in the SOT is also stale

`ANALYTICS-SOT.md` §10 records: *"Does not build ads. Four ads agents exist against **zero** ads
endpoints; that gap is upstream."* That upstream gap is closed — **28 ads endpoints now exist**
(Meta Ads 11, Google Ads 17), both with their own `ai-insights` and `demographics`. The non-goal
needs rewriting or the four ads agents stay blind for no reason.

---

## The gap, grouped by what a user would ask

Nine endpoints are single-entity reads (`{platform}/post`, `twitter/tweet`, `pinterest/pin`,
`youtube/video`, `youtube/find-video`, and the new `bluesky/post`, `threads/post`). **Those are
already scoped** by `[BE] Add post detail and post performance reads for the assistant` in the AI
chat skills epic — excluded here to avoid duplicating it.

| # | Capability gap | Eps | Where | The question nobody can ask today |
|---|---|---|---|---|
| A | **Competitor analytics** | 22 + 6 | FB, IG | "How do we compare with our competitors this month? Who posts more? What hashtags do they use?" |
| B | **Ads performance** | 28 | Meta Ads, Google Ads | "What did we spend and what did it return?" Four ads agents exist with nothing to call. |
| C | **Threads** | 16 | Threads | The whole platform. Includes its own `ai-insights`, `demographics`, `topic-tags`, `posts-per-hours`. |
| D | **Bluesky** | 10 | Bluesky | The whole platform. |
| E | **Audience composition** | 10 | FB, IG, LI, YT, Threads | "Who follows us — age, gender, which countries and cities?" |
| F | **Campaign & label performance** | 5 | campaigns-labels | "How did the summer campaign do?" The agent is handed the campaign and label list at bootstrap but cannot read a number against it. |
| G | **Format deep-dives** | 8 | FB, IG, YT, Pinterest | Reels, Stories, video insights, pin performance. |
| H | **Worst and re-sorted content** | 9 | FB, IG, LI, TikTok, YT, X, Bluesky, Threads | "What flopped?" and "top posts by comments, not engagement." |
| I | **Publishing cadence** | 10 | FB, IG, LI, TikTok, GMB, YT, Bluesky, Threads | "Are we posting enough, and which format pulls its weight?" |
| J | **GMB local business** | 4 | GMB | Reviews, search keywords, customer actions, media activity. A GMB customer's most distinctive data is entirely dark. |
| K | **Hashtags** | 5 | IG, LI, Bluesky, Threads (+ competitor hashtags in A) | "Which hashtags actually work for us?" |
| L | **Daily-delta series twins** | 8 | Pinterest, YouTube | Daily change vs the cumulative series already registered. Largely redundant — worth a deliberate "we do not register these" rather than silent absence. |
| M | **X metering + series** | 2 | X | `credits-used`, `engagement-impression`. |

Gap E has a documented failure mode behind it: SOT §11.3 records that given no vocabulary for a
request, **the model invents a product limitation** rather than admitting a tool gap. That is how
`analytics_top_content` came to tell users X had no top-content endpoint when it did (ANL-21).

---

## Tools versus arguments

The Analytics Analyst is at 6 tools, and the skills epic already carries the AC "Adding these reads
does not push any single assistant capability past its tool limit." One tool per gap above would
quadruple the count.

The precedent in this codebase runs the other way. `ANALYTICS-SOT.md` §3 is explicit that "the
composites are BEHAVIOUR not tools", and §11.4 resolved the charts gap as **parameters, not tools**
— `accounts` and `over_time` were added to existing tools rather than spawning `analytics_chart`.
Applying the same test:

| Gap | Shape |
|---|---|
| E (demographics), F (campaigns), G (formats), I (cadence), K (hashtags), L, M | **Arguments** — a new `metric_set`, a dimension argument, a scope argument for campaigns (which is not a platform), new registry entries |
| H (worst / re-sorted) | **Arguments** — `order_by` and `direction` on `analytics_top_content` |
| C, D (Threads, Bluesky) | **Registry entries only** — they are ordinary platforms; regenerating the registry should pick them up |
| A (competitor) | **Its own tool** — different entity (a competitor report, not a connected account), different required id, different answer shape. Plus a way to list reports. |
| B (ads) | **Its own tool or toolkit** — different entity (ad accounts, campaigns, ad sets, ads), spend-and-return metrics that do not fit the organic metric registry |
| J (GMB local) | **Likely its own tool** — reviews and search keywords are a different data shape from metric rows |

---

## Files Involved

| File | Change |
|---|---|
| `contentstudio-ai-agents/api-docs.json` | Refresh from the live spec — it is 77 endpoints behind |
| `contentstudio-ai-agents/scripts/generate_analytics_registry.py` | Regenerate; may need new descriptor shapes for competitor and ads |
| `contentstudio-ai-agents/src/integrations/contentstudio/analytics_registry.py` | Regenerated output |
| `contentstudio-ai-agents/src/integrations/contentstudio/analytics.py` | `PLATFORMS`, `INSIGHTS_PLATFORMS`, `BEST_TIMES_PLATFORMS`, endpoint maps |
| `contentstudio-ai-agents/src/integrations/contentstudio/toolkits/analytics.py` | New competitor tool, new arguments on existing tools |
| `contentstudio-ai-agents/tests/unit/integrations/test_analytics_spec_drift.py` | Drift guard must cover every newly registered family |
| `contentstudio-ai-agents/tasks/agentic-architecture/ANALYTICS-SOT.md` | §2.1 counts, §3 tool surface, and the §10 ads non-goal all move |

---

## Gotchas

- **`competitor_report_id` is required on every competitor read.** Listing the workspace's
  competitor reports has to come first, or the tool has nothing to pass.
- **Competitor is Facebook and Instagram only.** The tool must decline other networks rather than
  returning an empty comparison that reads as "no competitors".
- **Competitor data can be legitimately empty** — a report whose data has not been fetched yet.
  SOT §3.2 exists to force the answer to admit what is missing; the same discipline applies.
- **`campaigns-labels` is not a platform.** It shares the path segment but is excluded from
  `PLATFORMS`, which is derived from metric descriptors. It needs its own argument rather than
  being smuggled in as a ninth platform.
- **The registry is generated, the endpoint maps are hand-written.** That split already produced a
  false capability claim the model was instructed to repeat (ANL-21). Anything hand-added needs a
  drift test, and `test_analytics_spec_drift.py` only guards the three existing constants.
- **Series payload shapes are not derivable from the spec** — which key holds the rows and which is
  the axis has to be authored per endpoint. Every new series family inherits that cost.
- **Ads metrics are a different currency.** Spend, CPC, CPM, ROAS and conversions do not belong in
  a metric registry harvested from organic `summary` blocks.
- **Threads and Bluesky have `capabilities` endpoints**, which the other platforms do not. Worth
  checking whether they should drive what the assistant claims each platform supports rather than
  another hand-written map.
