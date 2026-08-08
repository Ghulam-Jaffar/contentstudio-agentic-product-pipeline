# Epic: Analytics network research, Bluesky and Threads

## Problem

We publish to Bluesky and Threads and give users no analytics for either. Both were requested as **research first**: find out what each network's API can actually tell us, compare it to the analytics we already deliver elsewhere, and decide what is realistic before anyone writes a dev story.

That order matters here more than usual. Bluesky is an open protocol with a very different access model from anything else we integrate, and Threads sits behind Meta's permission and review process. Guessing at either produces dev stories that get thrown away.

## Goal

Two scope documents, one per network, each concrete enough that a PO can write dev stories from it without commissioning a second investigation. Each answers: what data exists, how much of our existing analytics experience we can reproduce, what the constraints are, whether our current account connection is sufficient, and what we recommend shipping.

## Scope

Two research stories. No code ships from this epic. Development tickets are created after the scope documents are reviewed and signed off.

## Rules

- **Map against what we already deliver.** A metric list on its own is not a deliverable. Every section we ship on other networks has to be marked achievable, partly achievable, or not achievable, with the reason.
- **Answer the connection question explicitly.** If existing connected accounts would need re-authorising for analytics access, say so. That is a user-facing migration and it changes the dev work substantially.
- **Specify the not-enough-data case.** Every one of our analytics screens needs that state. Find out what the API returns for a new or quiet account rather than leaving it to be discovered during build.
- **Target where the pipeline is going.** Analytics ingestion is mid-migration. Specify against the standardised payload shape and the current ingestion direction, not the older per-platform variations.
- **Recommend, do not just report.** Each document ends with a recommended scope, including what to defer and what to tell users we cannot offer.

## Sequencing

Independent of each other, can run in parallel. Threads is the likelier one to be gated on an external approval process, so if only one can start, start Threads to get any Meta review moving early.

## Out of scope

- Any implementation. No endpoints, no ingestion, no UI.
- Social Listening coverage of these networks, which already exists and is a different thing (public mention monitoring, not account analytics).
- Public analytics API and Looker Studio exposure. Those follow after the internal scope is settled, as separate stories.

## Stories

1. `[Research] Bluesky analytics capabilities and recommended ContentStudio scope`
2. `[Research] Threads analytics capabilities and recommended ContentStudio scope`

---
---

# 1. [Research] Bluesky analytics capabilities and recommended ContentStudio scope

### Description

We publish to Bluesky but offer no analytics for it, and users who have connected a Bluesky account reasonably expect the same reporting they get for their other networks. Before committing to build anything, we need to know what Bluesky's APIs can actually tell us about an account and its posts, and how much of our existing analytics experience is reproducible. Bluesky is an open protocol rather than a conventional platform API, so the access and data-availability model is unlike anything else we integrate, and the answer may be materially better or worse than our other networks.

The deliverable is a scope document that a PO can write dev stories from.

### Workflow

*(Research story. The deliverable is a document, not a user-facing change.)*

1. Inventory what Bluesky's APIs expose, at account level and post level.
2. Compare that inventory against the analytics sections we already deliver on other networks.
3. Establish the constraints: authentication, rate limits, historical data, granularity, latency.
4. Determine whether our existing Bluesky publishing connection already grants what analytics needs.
5. Determine what the APIs return for a new or low-activity account.
6. Recommend a scope: ship, defer, or decline, with reasons.
7. Review the document with product before any dev story is written.

### Acceptance criteria

- [ ] A document exists inventorying every analytics-relevant data point Bluesky's APIs expose, separated into account-level and post-level.
- [ ] For each analytics section we deliver on other networks, the document states whether it is achievable on Bluesky, partly achievable, or not achievable, with the reason. Sections to cover at minimum: summary, audience growth, publishing behaviour, top posts, posts per day, hashtags, and follower demographics.
- [ ] The document states which of our standard engagement metrics are available, and names any Bluesky-specific interaction type that has no equivalent on our other networks.
- [ ] Authentication and authorisation requirements are documented, including whether analytics access needs anything beyond what publishing already uses.
- [ ] The document answers explicitly whether existing connected Bluesky accounts would need to be re-authorised, and if so, what that means for users.
- [ ] Rate limits are documented, with an assessment of whether they support the fetch volume our sync would need across our connected account base.
- [ ] Historical data availability is documented: how far back data goes, whether there is backfill, and what a newly connected account can show on day one.
- [ ] Data granularity and freshness are documented, including the delay between a post being published and its metrics being available.
- [ ] Any protocol-level consideration specific to Bluesky is documented, including how self-hosted instances or alternative providers affect what we can read and whether all connected accounts are equally readable.
- [ ] The document states what the APIs return for a new or low-activity account, so the not-enough-data state can be specified in the dev stories.
- [ ] Where the data would land in the analytics ingestion pipeline is assessed, against the current direction of the pipeline rather than its older per-platform shape.
- [ ] The document ends with a recommended scope: what to ship first, what to defer, and what to tell users is not available and why.
- [ ] Gaps against our other networks are called out plainly, so we do not ship a Bluesky analytics screen that quietly looks impoverished next to the others without that being a known decision.
- [ ] The document is reviewed with product, and the review outcome is recorded, before any development story is created.

### Mock-ups

N/A. Research deliverable. If the recommended scope implies a screen materially different from our existing per-network dashboards, the document should say so, so a design story can be raised.

### Impact on existing data

None. Research only.

### Impact on other products

- None directly. The document informs future work in the analytics pipeline, the web app, and eventually the public analytics API and the Looker Studio connector.
- If the connection question comes back as "re-authorisation needed", that has a user-facing consequence which the document must flag for the PO.

### Dependencies

None. Can start immediately.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A, research
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled) — N/A, research
- [ ] UI theming supported (default + white-label, design library components are being used) — N/A, research
- [ ] White-label domains impact reviewed — N/A, research
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension) — the document should note whether the recommended scope has any mobile implication

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- What we already have for Bluesky: `contentstudio-backend/app/Strategy/Integrations/BlueskyConnector.php` shows exactly what the current connection obtains, and `app/Strategy/Planner/BlueskyPosting.php` shows what we do with it. Reading those first answers the connection-sufficiency question faster than reading Bluesky's docs.
- The comparison baseline is `contentstudio-social-analytics-go/src/api/analytics/router.go`, which lists every section we serve per platform. LinkedIn's set is representative: `summary`, `audienceGrowth`, `pageViews`, `publishingBehaviour`, `topPosts`, `postsPerDays`, `hashtags`, `getTopPosts`, `followersDemographics`, `ai_insights`.
- The best templates for what a new platform implementation looks like are the smaller, more recent packages under `contentstudio-social-analytics-go/src/api/analytics/`: `pinterest/`, `tiktok/`, `linkedin/`.
- Social Listening already monitors Bluesky as one of its sources. That is public mention monitoring rather than account analytics, but the researcher should check whether any of that access is reusable before concluding something is unavailable.
- The YouTube demographics analytics epic is the most recent add-analytics-coverage effort and the closest model for the dev stories that will follow this research.
- Specify against the standardised analytics payload shape rather than the current per-platform variation, since the per-platform analytics API standardisation work is moving every platform toward it.

---
---

# 2. [Research] Threads analytics capabilities and recommended ContentStudio scope

### Description

We publish to Threads and offer no analytics for it. Threads sits behind Meta's API surface, which means its permission model, review process and metric availability are all things we need to establish before committing to build, and Meta's approval timelines can themselves shape when this is deliverable. We also already integrate Facebook and Instagram analytics through Meta APIs, so there is existing knowledge and existing infrastructure to check against before treating Threads as a new problem.

The deliverable is a scope document that a PO can write dev stories from.

### Workflow

*(Research story. The deliverable is a document, not a user-facing change.)*

1. Inventory what the Threads APIs expose, at account level and post level.
2. Compare that inventory against the analytics sections we already deliver, especially against Instagram, as the closest existing Meta implementation.
3. Establish the permission and review requirements, and how long approval realistically takes.
4. Establish rate limits, historical data availability, granularity and latency.
5. Determine whether our existing Threads publishing connection already grants what analytics needs.
6. Determine what the APIs return for a new or low-activity account.
7. Recommend a scope, then review with product before any dev story is written.

### Acceptance criteria

- [ ] A document exists inventorying every analytics-relevant metric the Threads APIs expose, separated into account-level and post-level.
- [ ] For each analytics section we deliver on other networks, the document states whether it is achievable on Threads, partly achievable, or not achievable, with the reason. Sections to cover at minimum: summary, audience growth, publishing behaviour, top posts, posts per day, hashtags, and follower demographics.
- [ ] The document compares Threads metric availability specifically against our existing Instagram analytics, since that is the closest Meta implementation we already run, and notes what is reusable.
- [ ] Required permissions and scopes are documented, along with whether they need Meta app review, what that review requires from us, and a realistic expectation of how long it takes.
- [ ] Authentication requirements are documented, including whether the existing Threads publishing connection is sufficient or a new or expanded authorisation is needed.
- [ ] The document answers explicitly whether existing connected Threads accounts would need to be re-authorised, and what that means for users.
- [ ] Rate limits are documented, with an assessment of whether they support the fetch volume our sync would need across our connected account base.
- [ ] Historical data availability is documented: how far back data goes, whether backfill is possible, and what a newly connected account can show on day one.
- [ ] Data granularity and freshness are documented, including the delay between publishing and metrics availability.
- [ ] Account-type constraints are documented: whether analytics is available for all Threads account types we can connect, or only some.
- [ ] The document states what the APIs return for a new or low-activity account, so the not-enough-data state can be specified in the dev stories.
- [ ] Where the data would land in the analytics ingestion pipeline is assessed, including whether the existing Meta ingestion path can be extended rather than duplicated.
- [ ] The document ends with a recommended scope: what to ship first, what to defer, and what to tell users is not available and why.
- [ ] Gaps against our other networks are called out plainly.
- [ ] Any dependency on Meta approval that could block delivery is flagged as a schedule risk, not buried in the constraints section.
- [ ] The document is reviewed with product, and the review outcome is recorded, before any development story is created.

### Mock-ups

N/A. Research deliverable. If the recommended scope implies a screen materially different from our existing per-network dashboards, the document should say so, so a design story can be raised.

### Impact on existing data

None. Research only.

### Impact on other products

- None directly. The document informs future work in the analytics pipeline, the web app, and eventually the public analytics API and the Looker Studio connector.
- If Meta app review is required, that has a delivery-timeline consequence the PO needs to know early.

### Dependencies

None. Can start immediately, and should start early if Meta review turns out to be on the critical path.

### Global quality and compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — N/A, research
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled) — N/A, research
- [ ] UI theming supported (default + white-label, design library components are being used) — N/A, research
- [ ] White-label domains impact reviewed — N/A, research
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension) — the document should note whether the recommended scope has any mobile implication

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- What we already have for Threads: `contentstudio-backend/app/Strategy/Integrations/ThreadsConnector.php` shows what the current connection obtains, and `app/Strategy/Planner/ThreadsPosting.php` shows what we do with it. Start there for the connection-sufficiency question.
- The closest existing Meta analytics implementation is Instagram: `contentstudio-social-analytics-go/src/api/analytics/instagram/` and the Instagram sections in `router.go`. If Threads can extend that ingestion path rather than needing its own, that materially shrinks the dev work and is worth establishing during research.
- The comparison baseline for sections is `contentstudio-social-analytics-go/src/api/analytics/router.go`.
- Our existing Meta integrations already carry accumulated knowledge of Meta's review process and permission naming. Checking what we already hold before assuming a new review is needed may shorten this considerably.
- Social Listening already monitors Threads as one of its sources, using public data. Worth checking for reusable access.
- Specify against the standardised analytics payload shape that the per-platform analytics API standardisation work is moving every platform toward.
- The YouTube demographics analytics epic is the closest model for the dev stories that follow.
