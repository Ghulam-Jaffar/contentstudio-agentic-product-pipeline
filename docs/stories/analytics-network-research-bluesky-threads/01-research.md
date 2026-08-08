# Research — Analytics network expansion: Bluesky and Threads

Items 23 and 24 of the 7 Aug 2026 backlog batch. Both were requested as **research tickets first**, with development tickets to be created after the scope is finalised. This document is the pre-work that scopes those research stories, not the research itself.

## Current state

### Both networks publish, neither has analytics

Publishing is live for both:

- **Bluesky** — `contentstudio-backend/app/Strategy/Integrations/BlueskyConnector.php`, `app/Strategy/Planner/BlueskyPosting.php`, and a `BlueskyPreview.vue` in the shared social-previews system.
- **Threads** — `contentstudio-backend/app/Strategy/Integrations/ThreadsConnector.php`, `app/Strategy/Planner/ThreadsPosting.php`, with `ThreadsPreview.vue` and `ThreadsMultiPreview.vue` in the shared previews.

Analytics is absent for both. `contentstudio-social-analytics-go/src/api/analytics/router.go` registers handlers for Facebook, Instagram, LinkedIn, X, TikTok, YouTube, Pinterest, Google Business Profile, overview, campaign and label, Looker Studio, Facebook and Instagram competitor, Meta Ads, Google Ads, and accounts. No Bluesky, no Threads.

### The comparison baseline the research needs

The research has to compare what each network's API offers against what we already deliver. The concrete shape of "what we already deliver" is the per-platform section list in `router.go`. LinkedIn, as a representative example, exposes: `summary`, `audienceGrowth`, `pageViews`, `publishingBehaviour`, `topPosts`, `postsPerDays`, `hashtags`, `getTopPosts`, `followersDemographics`, `ai_insights`.

Every platform has a variation of that set. A research doc that says "Bluesky offers follower counts" without mapping it against this list is not actionable, which is why each research story below requires the mapping explicitly.

### Where Bluesky and Threads already appear

Both are monitored platforms in Social Listening, which covers 18 sources including `bluesky` and `threads`. That is mention monitoring over public data, not account analytics, so it does not answer the analytics question. It is worth the researcher knowing about, since some of the public-data access patterns may overlap.

### Adjacent in-flight work the research must not contradict

- `docs/stories/analytics-api-consistency/` — standardises analytics API payload shape per platform. Any new platform should be specified against the standardised shape, not the current per-platform variation.
- `docs/stories/public-analytics-api-rollout/` — adds public analytics API coverage per platform, mirroring Facebook. A new network eventually needs a story there too.
- `docs/stories/analytics-php-to-golang-migration/` and `docs/stories/analytics-data-fetch-job-split/` — the ingestion architecture is moving. New platform ingestion should target where it is going, not where it was.
- `docs/stories/youtube-demographics-analytics/` — the most recent "add analytics coverage for a network" epic, and the closest template for what the dev stories after this research will look like.

## What each research story must produce

Both stories share the same deliverable shape, differing only in network. The output is a scope document that a PO can turn into dev stories without a second investigation:

1. **Metric inventory.** Everything the network's API exposes, split into account-level and post-level.
2. **Mapping against our existing sections.** For each section we deliver on other networks, whether it is achievable here, partially achievable, or not achievable, with the reason.
3. **Constraints.** Authentication and permission model, rate limits, historical-data availability and retention, granularity, latency between action and data availability, and any approval or review process required for the necessary access.
4. **Feasibility of our current connection.** Whether the existing publishing connection already carries the access analytics needs, or whether a new or expanded authorisation is required. This matters a lot: if existing connected accounts need to be re-authorised, that is a migration with user friction and it changes the shape of the dev work.
5. **Recommended scope.** What we should ship, what we should defer, and what we should tell users is not available and why.
6. **Not-enough-data behaviour.** What the API returns for a new or low-activity account, so the dev stories can specify that state rather than discovering it late.
7. **Ingestion fit.** Where the data would land in the analytics pipeline and whether the existing per-platform ingestion pattern fits.

## Why one story per network rather than one combined

Different APIs, different owners, different constraints. Threads is a Meta API with Meta's permission and review model; Bluesky is an open protocol with a very different access story. Combining them produces a document that is vague about both. They are grouped into one epic so they can be scheduled together, not merged.

## Files and systems the research will touch

- `contentstudio-backend/app/Strategy/Integrations/{BlueskyConnector,ThreadsConnector}.php` — what the current connection already grants
- `contentstudio-backend/app/Repository/Integrations/Platforms/Platforms.php` — how platforms are registered
- `contentstudio-social-analytics-go/src/api/analytics/router.go` — the section inventory to compare against
- `contentstudio-social-analytics-go/src/api/analytics/linkedin/`, `tiktok/`, `pinterest/` — the smaller, more recent platform implementations, which are the best templates for a new one
- `contentstudio-social-analytics-go/src/services/` — per-platform ingestion services

No code changes ship from either research story.

## Mobile

None. Both stories are research. Mobile analytics scope, if any, is decided after the scope document exists.
