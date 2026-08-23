# Research — Bluesky analytics integration (surfacing work)

**Requested as:** three stories covering Overview analytics, the dedicated Bluesky analytics page, and report creation and management. Explicitly scoped as **generic stories, no codebase investigation** — this doc records the assumptions the stories rest on rather than verified code pointers.

## Starting position

The dev team has already integrated Bluesky on the data side. These three stories cover the **product surfacing** of that integration: making Bluesky accounts selectable, showing their metrics, exposing AI insights, and letting users generate and manage Bluesky reports.

All three are written as `[FE]` stories on that basis. If any analytics section, AI insights endpoint, or report generator does **not** yet return Bluesky data, a companion `[BE]` story is needed per gap before the matching FE story can start. That is the single biggest assumption in this batch and the first thing the PO should confirm with the dev who did the integration.

## Prior work in this repo

- `docs/stories/analytics-network-research-bluesky-threads/` — the research spikes for Bluesky and Threads analytics, raised as items 23 and 24 of the 7 Aug 2026 backlog. Those stories scoped *whether and what* to build. This batch is the *build* side, so there is no duplication, but the scope document those spikes produce should be reconciled with these three stories before they are pointed.
- `docs/features/gmb-analytics/` — the closest precedent for adding a network to analytics end to end. Its story split (dedicated endpoints, Overview and Campaign integration, export reports, AI insights, then the FE pages) is the shape this batch compresses into three.
- `docs/stories/analytics-api-consistency/` — standardises analytics payload shape per platform. Bluesky should be specified against the standardised shape, not a bespoke one.
- `docs/stories/analytics-account-support-tooltips/` and `docs/stories/analytics-supported-account-type-messaging/` — the established pattern for telling users why a metric is missing for a given network. Reused in all three stories here instead of inventing new messaging.
- `docs/stories/analytics-empty-state-screens/` — the empty-state pattern the new Bluesky page should follow.

## Metric availability assumption

Bluesky is an open protocol with public engagement counts and no advertising or delivery telemetry. The stories assume:

**Available:** followers, follows, post count, per-post likes, reposts, replies, quotes, and anything derivable from those (engagement totals, engagement rate against followers, top and least performing posts, posting frequency and cadence).

**Not available:** impressions, reach, video views, profile visits, link or button clicks, and audience demographics such as age, gender, country, or city.

Every unsupported metric is handled with the existing "network does not provide this" tooltip pattern rather than a zero, because a zero reads as "no engagement" instead of "not measurable here". **Confirm this split against what the dev's integration actually ingests before the stories are estimated** — if impressions turn out to be available, the unsupported-metric copy in each story drops out.

## Deliberately out of scope for this batch

- **Campaign and Label analytics.** GMB got a story for it. Bluesky should too, but it was not requested here. Flag it to the PO as a likely fourth story.
- **Public analytics API coverage.** `docs/stories/public-analytics-api-rollout/` is where a Bluesky endpoint would land.
- **Competitor analytics.** Only Facebook and Instagram have it.
- **Mobile.** Whether the Flutter app surfaces per-network analytics pages was not checked. If it does, one `[Flutter]` story is needed; if analytics is web-only in the app, nothing is required. Confirm before closing the batch.
- **A `[Design]` companion story.** All three stories are frontend work, so per the standing rule one `[Design]` story should accompany the batch. Not written here because the request was for exactly three stories.

## Notes for whoever implements

- Bluesky's brand colour is a fixed blue that is close to ContentStudio's default primary. Platform colour must stay a literal platform constant and never borrow the theme primary, or Bluesky rows become invisible on a white-label domain whose primary is also blue. Worth an explicit check in review.
- Bluesky handles are domain-shaped, so account labels in selectors and report headers need to tolerate long strings and truncate rather than wrap.
