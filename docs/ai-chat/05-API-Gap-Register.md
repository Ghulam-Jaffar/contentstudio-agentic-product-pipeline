# ContentStudio API — Gap Register for AI Chat (v2.0)

**Purpose:** everything the AI chat needs that the current API cannot do. Sized for backend planning.

**Reviewed against:** live source in `contentstudio-backend/routes/api/v1.php` — **159 endpoints**. (v1.0 of this doc was written against an exported spec showing 118 and is superseded. See [As-Built Baseline](https://app.helpin.ai/share/fff3d91449a0c4f64f84ac5eadc98f9d).)

**Summary:** 3 P0 blockers · 5 P1 · 7 P2. **Nothing in Phase 1 (read-only) is blocked. Phase 3 (Inbox) is no longer blocked either.**

> ### What changed from v1.0
> - **Gap 2 (No Inbox API) — RESOLVED.** 28 inbox endpoints are live. Replaced by three narrower gaps: 2a sentiment/intent fields, 2b needs-reply filter, 2c bulk route exposure.
> - **Gap 9 (No GMB review replies) — RESOLVED.** `POST`/`DELETE /inbox/reviews/{review_id}/reply` both exist.
> - **Gap 4 re-scoped** from "build" to "expose and reconcile" — cross-account aggregation already exists in `src/agents/analytics/overview/`.
> - **Gap 16 added** — no ads analytics in the public API.
> - Gaps 1 and 3 verified unchanged and still P0.

---

## P0 — Blocks Phase 2

### Gap 1 · No update-post endpoint
**Missing:** `PUT /api/v1/workspaces/{workspace_id}/posts/{post_id}`
**Have:** create, delete, approve, notes only. **Verified against source — still absent.**

**Why it blocks:** "Change the caption," "move it to 5pm," "add LinkedIn to that post," "swap the image," "add a label" — this is the second thing every single user will ask after creating a post. Delete-and-recreate destroys the post ID, approval history and internal comment thread, and is not something we can ship.

**Needs to support:** content text and media · accounts (add/remove) · scheduled_at · publish_type (draft → scheduled) · labels · campaign_id · first_comment · platform option blocks

**Constraint:** must reject edits to already-published posts, or define exactly what happens (platform edit vs no-op).
**Blocks:** UC-P09, UC-C06, UC-C08
**Effort:** M–L

---

### Gap 2 · Inbox API — RESOLVED, with three narrower gaps remaining

~~No Inbox API~~ — **28 endpoints are live** at `workspaces/{workspace_id}/inbox/*`, proxied to `social-inbox-manager`. Items (unified triage), conversations/DMs, post comments, reviews, tags. Full list in [As-Built Baseline](https://app.helpin.ai/share/fff3d91449a0c4f64f84ac5eadc98f9d) §2.1.

Requirement #4 from v1.0 (one unified conversation model) **is already satisfied** — `GET items` is the single triage surface across DMs, comments and reviews, filterable by `inbox_type`, `platform_id`, `element_type`, `assigned`, `marked_done`, `archived`, `tags`, `search_term`.

Three real gaps remain:

#### Gap 2a · No sentiment or intent field on inbox items — **P1**
Sentiment exists in `app/Data/Listening/*` and Facebook analytics, but not on inbox items. Without it every triage agent re-analyses raw text on every run — slow, expensive, and inconsistent between runs for the same message.
**Blocks:** quality of the Inbox Triage agent
**Effort:** M

#### Gap 2b · No server-side "needs reply" filter — **P1**
`marked_done` and `assigned` are the closest available. Computing "needs reply" client-side means fetching everything first, which defeats the point for a busy inbox.
**Effort:** S

#### Gap 2c · Bulk actions exist downstream but are not exposed — **P2, cheap**
The SIM service filters on `element_ids[]` (a list). The public routes take a single `{element_id}` in the path and wrap it as a one-element list — see the comment at `InboxItemController::markDone`. **So bulk is a route-signature change, not new plumbing.** v1.0 costed this as L; it is S.
**Effort:** S

**Net:** Phase 3 can start now. Idempotency on reply is still required — see Infra A.

---

### Gap 3 · No post validation / dry-run
**Missing:** `POST /api/v1/workspaces/{workspace_id}/posts/validate` — **verified absent.**

**Why it blocks:** AI-generated content will break platform rules routinely — caption length per platform, image counts, video duration and aspect ratio, carousel minimums, YouTube title ≤100, GMB title ≤100, TikTok requirements. Without a validator we find out at publish time, hours after the user has closed the tab.

**Should return:** per-account pass/fail, error list with the specific rule broken, warnings (e.g. "no hashtags for Instagram"), and **suggested fixes** the AI can apply automatically.

**Secondary benefit:** this is the highest-leverage single endpoint for chat quality. It converts a class of silent failures into an in-conversation self-correction loop — the AI fixes its own mistake before the user ever sees it.

**Blocks:** quality of every publishing use case
**Effort:** M

---

## P1 — Degrades the launch

### Gap 4 · No cross-account analytics rollup in the public API — **re-scoped**
**Missing from the public API:** `GET /analytics/overview` and `/analytics/compare`. Every one of the 99 analytics endpoints takes a single `platform_id`. "How did we do last week?" across 12 accounts is 12+ sequential calls.

**Re-scoped from v1.0.** This is no longer a from-scratch build. `contentstudio-ai-agents/src/agents/analytics/overview/` already does cross-account aggregation across 6 agents (`overview_account_performance`, `overview_account_statistics`, `overview_engagement_performance`, `overview_impressions_performance`, `overview_reach_performance`, `overview_top_posts`), and the internal API already serves the overview dashboard. **The work is exposing and reconciling what exists, not inventing normalisation.**

**Still needs product sign-off:** the metric-normalisation mapping. `fan_count` (FB) / followers (IG, X, TikTok, Pinterest) / subscribers (YT) / connections (LI) all becoming `audience`; same for impressions vs reach vs views. Leadership and clients read these numbers and they must mean the same thing across platforms. Confirm the existing overview agents' mapping is the one we standardise on, rather than defining a second.

**`/analytics/compare` must return normalised rates, not totals** — a 500k-follower account always wins on raw engagement, which is never the question being asked.

**Blocks:** UC-A01, UC-A03, W3, W6, W7, Morning Brief agent, Weekly Reporter agent, `/client-report`
**Effort:** S–M (down from M)

---

### Gap 5 · Limited best-time-to-post data
**Have:** `/analytics/{platform}/active-users` — **Facebook and Instagram only.** Partial substitutes exist: `publishing-behaviour` (FB, IG, LI, TikTok, Pinterest, GMB) and `performance-schedule` (YouTube).
**Missing:** a first-class best-times answer on X, LinkedIn, TikTok, YouTube, Pinterest, GMB.

Ideally: `GET /analytics/best-times?account_ids[]=&lookback=90d` returning recommended windows per account, derived from our own published-post performance where the platform gives us nothing.

**Why it matters:** "schedule at the best times" appears in the flagship workflow, in Vista's marketing, and in almost every competitor's. Answering it properly for two platforms out of eight is a visible hole.
**Interim:** derive from `publishing-behaviour` plus our own post performance and **label it clearly as an estimate.** Never present a guess as platform data.
**Blocks:** UC-A07, W1, W2, `/best-times`
**Effort:** M

---

### Gap 6 · Content Categories are read-only
**Have:** `GET /content-categories` — the only endpoint. **Verified.**
**Missing:** create, update, delete, and slot/schedule management.

The chat can schedule into a category but cannot create one. Given the recent Content Categories redesign with one-click templates, exposing create through the API would make "set me up with a category structure" work during onboarding (W5) — a strong activation moment.
**Blocks:** part of W5
**Effort:** S–M

---

### Gap 7 · No queue / slot visibility
**Have:** `publish_type: "queued"` is accepted.
**Missing:** anything that reads queue configuration or returns the next available slots.

The AI can queue a post but cannot tell the user when it will go out — which makes the confirmation card unable to state the one fact the user cares about.
**Needs:** `GET /workspaces/{id}/queue` (config) and `GET /workspaces/{id}/queue/next-slots?count=10`
**Blocks:** UC-P12
**Effort:** S

---

### Gap 8 · No Discover / trending endpoint
**Missing:** the entire Discover module is absent from the public API. **Verified.**

**Partial mitigation that already exists:** the chat has Exa-backed web research (`tool_exa_answer`, `tool_find_similar`, `tool_get_contents`), so open-web trend research is reachable today. What is unreachable is **ContentStudio's own Discover index** — our curated, per-topic content engine.

**Why it matters more than its priority suggests:** the flagship workflow (W1: niche → trends → campaign → posts) starts here, and **Discover is our clearest genuine advantage over Vista Social.** They have no content discovery engine. We have a Feedly-class one and it is currently unreachable from the chat.

**Minimum viable:** `GET /discover/trending?topics[]=&region=&window=7d` returning trending articles/topics with engagement signals.
**Blocks:** W1, `/trend-ideas`, Trend Scout agent
**Effort:** M
**Recommendation:** pull forward. This is the difference between matching Vista and beating them. Note the existing `src/agents/discovery/` agents (topic_taxonomy, topic_tagger, source_candidates, brand_context) as likely building blocks.

---

### ~~Gap 9 · No Google Business review replies~~ — **RESOLVED**
`POST /workspaces/{id}/inbox/reviews/{review_id}/reply` and `DELETE .../reply` are both live. Review reading is available via `GET /analytics/gmb/reviews` and via the inbox items surface.

**The Review Watcher agent in [Agents & Skills](https://app.helpin.ai/share/f4cb43a256cbad3ee529632028d38f30) is buildable end-to-end today** — watch and answer. No longer strictly worse than Vista's.

---

## P2 — Fast follow

| # | Gap | Detail | Blocks | Effort |
|---|---|---|---|---|
| 10 | **No bulk post operations** | No bulk delete, reschedule, label or campaign-assign on posts. "Push everything next week forward a day" is N calls. (Distinct from inbox bulk — see Gap 2c.) | UC-C06 | M |
| 11 | **Media library incomplete** | No delete. No folder list/create — `folder_id` is accepted on upload but not discoverable. No metadata/tagging. **Verified: list + upload only.** | UC-M04 | S |
| 12 | **No account disconnect** | `connect`, `add/bluesky`, `add/facebook-group` exist; nothing removes an account. | Cleanup workflows | S |
| 13 | **No post-level publish diagnostics** | No structured "why did this fail" beyond status. Needed for the Publishing Health Monitor agent to be genuinely useful. | UC-N03, health agent | M |
| 14 | **No cross-workspace analytics** | "Which client had the best month?" is impossible. Major agency ask; also blocks multi-workspace agents. | UC-G05 | L |
| 15 | **No AI generation endpoints in the public API** | Image/video/caption generation exists internally but is not exposed. Fine for chat (already wired); blocks parity for MCP and CLI. | MCP/CLI parity | M |
| 16 | **No ads analytics in the public API** ⭐ new | Meta Ads and Google Ads narrative agents exist in `src/agents/analytics/{meta_ads,google_ads}/`, but there are zero ads endpoints in v1. Unreachable from chat, CLI and MCP. A paid-media surface most competitors lack. | paid-media use cases | M |

---

## Infrastructure requirements (not endpoints, but required)

| # | Requirement | Priority | Note |
|---|---|---|---|
| A | **Idempotency-Key support on all write endpoints** | **P0** | Chat retries. Without this, a network blip publishes twice, or double-sends a DM. Must land before Phase 2 **and** before inbox reply ships in Phase 3. |
| B | **Per-key rate limits with clear 429 headers** | P0 | Exists as `throttle:api-v1` (`config/api_rate_limits.php`). **Verify the 429 response exposes retry headers** so the AI backs off instead of hammering. |
| C | **Structured error codes** (not just messages) | P0 | Partly in place — `BaseApiController::localizedFlatError` already carries a stable `error_code`. **Audit for coverage**, then wire to `suggested_tool` recovery. |
| D | **Token/connection status on `accounts_list`** | P0 | Expired-token detection depends on it entirely. Verify the field is present in the accounts response. |
| E | **Consistent pagination** across all list endpoints | P1 | Currently mixed — inbox uses `limit`, most others `per_page`. Worth normalising or documenting per-tool. |
| F | **Webhook: post published / post failed** | P1 | Event-triggered agents need it; polling does not scale. |
| G | **Unified request metering** | P1 | Already in progress for the API plan — chat and agents must count against the same meter. |

---

## Recommended build order for backend

Revised: Inbox moves up (unblocked and cheap), Gap 4 moves up (smaller than thought), the old Inbox sprint disappears.

| Sprint | Items | Unblocks |
|---|---|---|
| 1 | Infra A–D (audit C and D first — both may be partly done) | Phase 2 can start safely |
| 2 | Gap 4 (expose overview/compare) + analytics chat tools | Phase 1's headline capability actually works |
| 3 | Gap 2a/2b/2c (inbox sentiment, needs-reply, bulk routes) | **Phase 3 — shortest path to new capability** |
| 4 | Gap 1 (update post) + Gap 3 (validate) | Phase 2 |
| 5 | Gap 5 (best times) + Gap 7 (queue) | Flagship workflows complete |
| 6 | Gap 8 (Discover) + Gap 6 (categories write) | W1 flagship + onboarding |
| 7 | Gaps 10–16 + Infra E–G | Phase 4/5 polish |

**Two deliberate orderings.** Gap 4 stays ahead of Gap 1: Phase 1 is read-only and ships first, and without the analytics composites Phase 1 answers its most common question in 30 seconds. And Inbox (sprint 3) now sits ahead of the Phase 2 write gaps — it is live, the remaining work is small, and it delivers user-visible capability sooner than update-post can.
