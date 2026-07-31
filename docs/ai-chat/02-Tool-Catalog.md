# ContentStudio AI Chat — Tool Catalog (v1.1)

**Purpose:** the definitive list of tools the AI can call, and exactly which API endpoint each one maps to. This is the backend handoff document.

**Source:** ContentStudio API v1 — **159 endpoints**, verified against `contentstudio-backend/routes/api/v1.php` · Base URL `https://api.contentstudio.io` · Auth `X-API-Key`

**Total tools:** 27 in Phases 1–2, plus Inbox tools in Phase 3 = **~33**

> ### ⚠️ Read [As-Built Baseline](https://app.helpin.ai/share/fff3d91449a0c4f64f84ac5eadc98f9d) first
> **14 of these tools already exist and are live in the chat** as `mcp_*` tools on the Operations agent. This catalog was written as a greenfield spec; it is not. Before implementing anything here:
> - **8 Phase 1 read tools already ship** — see the "Exists" column in §9.
> - **The response envelope below is not what ships.** The live envelope is `{mcp_operation, operation, success, message, workflow_state, next_action, data, followup_actions, error, streamed_chunks}` from `mcp_tools.py::_envelope()`. Pick one deliberately.
> - **Safety tiers are already a state machine**, not per-tool flags. Map 🟢/🟡/🔴 onto the existing `resolving → awaiting_* → confirmation → complete` steps.
> - **New tools need "hallucination shim" kwargs and `strict=False`**, or Groq's router will fail them. And adding an operation touches 9 files — checklist in `contentstudio-ai-agents/CLAUDE.md`.
> - **There is a second, hosted MCP server** in the backend at `/api/mcp` with 8 tools. Decide whether tools land there too.

---

## 0. Conventions

**Naming:** `domain_action` in snake_case. Domain first so related tools cluster in the model's attention (`posts_*`, `analytics_*`, `media_*`). Note the live tools use an `mcp_` prefix instead (`mcp_fetch_posts`) — reconcile the convention before adding more.

**Proposed tool response envelope** (⚠️ differs from what ships today — see the banner above):
```json
{
  "ok": true,
  "summary": "Short natural-language result the model can quote directly",
  "data": { },
  "render": { "component": "post_preview", "props": { } },
  "next_cursor": "opaque-string-or-null",
  "warnings": ["Instagram token expires in 3 days"]
}
```

**Error envelope:**
```json
{
  "ok": false,
  "error_code": "ACCOUNT_TOKEN_EXPIRED",
  "message": "The Instagram account @bloomville needs to be reconnected.",
  "recoverable": true,
  "suggested_tool": "accounts_connect",
  "suggested_args": { "platform": "instagram", "process": "reconnect", "account_id": "..." }
}
```

`suggested_tool` matters. It lets the model recover in one turn instead of apologising and stopping.

**Safety tier** on every tool: 🟢 Green (auto-run) · 🟡 Amber (preview + confirm) · 🔴 Red (explicit confirm).

**`workspace_id`** is never a model-supplied argument. It is injected from session context. The model must not be able to act on the wrong workspace by hallucinating an ID.

---

## 1. Context & Session

### `context_get` 🟢
Bootstrap. Called automatically at session start and after any tool that changes accounts, labels, campaigns or categories. Not usually chosen by the model.

| Field | Source |
|---|---|
| user | `GET /api/v1/me` |
| workspaces | `GET /api/v1/workspaces` |
| accounts | `GET /api/v1/workspaces/{id}/accounts` |
| labels | `GET /api/v1/workspaces/{id}/labels` |
| campaigns | `GET /api/v1/workspaces/{id}/campaigns` |
| categories | `GET /api/v1/workspaces/{id}/content-categories` |

Returns a compact block (~500 tokens) injected into the system prompt: workspace name/id/timezone/first_day, account list (`id`, `name`, `platform`, `status`), label/campaign/category names + ids, user role + permissions, remaining AI credits, today's date in workspace timezone.

**Backend note:** cache for the session. Six calls on every message is not acceptable.

---

## 2. Social Accounts & Connections

### `accounts_list` 🟢
`GET /api/v1/workspaces/{workspace_id}/accounts`

| Arg | Type | Notes |
|---|---|---|
| `platform` | enum, optional | facebook, twitter, linkedin, pinterest, instagram, gmb, youtube, tiktok, tumblr, threads, bluesky, telegram. Comma-separated for multiple. |
| `search` | string, optional | Name match |
| `page`, `per_page` | int | max 100 |

Mostly redundant with `context_get`, but needed for paging past the bootstrap limit and for freshness checks after a connect.

**Response must include token/connection status.** "Which accounts need reconnecting?" is a top-10 question and depends entirely on this field being present.

### `platforms_list` 🟢
`GET /api/v1/platforms` — what can be connected, and how (OAuth vs credentials).

### `accounts_connect` 🔴
`POST /api/v1/workspaces/{workspace_id}/connect/{platform}?process={connect|reconnect}`

| Arg | Type | Notes |
|---|---|---|
| `platform` | enum, required | facebook, facebook-profile, instagram, instagram-via-facebook, twitter, linkedin, pinterest, tiktok, youtube, threads, gmb, tumblr |
| `process` | enum, required | `connect` or `reconnect` |
| `account_id` | string | Required when `process=reconnect` |
| `return_url` | string, optional | Where the browser lands after the flow |

**Special handling:** this tool does not complete an action — it returns an **OAuth URL**. The chat must render a "Connect Instagram →" button, not a raw link, and must poll or listen for completion so it can confirm in-thread. `return_url` should point back to the chat session so the user lands where they left off.

### `accounts_add_credentials` 🔴
Credential-based connections that skip OAuth.

| Sub-case | Endpoint | Args |
|---|---|---|
| Bluesky | `POST /workspaces/{id}/add/bluesky` | `handle`, `app_password` |
| Facebook Group | `POST /workspaces/{id}/add/facebook-group` | per spec |

**Never let the model echo credentials back into the transcript.** These must be collected through a secure input component in the UI, not typed into chat. Design and Security both need to sign this off.

---

## 3. Publishing

### `posts_list` 🟢
`GET /api/v1/workspaces/{workspace_id}/posts`

| Arg | Type | Notes |
|---|---|---|
| `status[]` | array | draft, scheduled, published, … |
| `date_from`, `date_to` | date | YYYY-MM-DD |
| `labels[]`, `campaigns[]`, `content_category[]` | array of ids | Resolve names → ids from context |
| `created_by[]` | array | user ids |
| `approval_assigned_to[]`, `approval_requested_by[]` | array | user ids |
| `comment_status` | enum | resolved / unresolved / all |
| `page`, `per_page` | int | |

Powers calendar questions, approval queues, gap detection, "what's going out tomorrow."

**Shaping:** return `id`, truncated text (140 chars), accounts, status, scheduled time, labels, campaign. Never the full payload — 50 posts of full payload is the whole context window.

### `posts_create` 🟡
`POST /api/v1/workspaces/{workspace_id}/posts`

The most important and most dangerous tool we ship. The payload is large; the tool schema should expose it in full but with heavy description text so the model fills it correctly.

**Required:** `content.text`, `accounts[]`, `scheduling.publish_type`

| Group | Fields |
|---|---|
| `content` | `text` (≤5000), `media.images[]` (≤10 URLs), `media.video`, `media.media_ids[]` (≤10, from media library) |
| `accounts[]` | Account IDs. Platform auto-detected. Optional if `content_category_id` given. |
| `content_category_id` | Uses the category's accounts and schedule |
| `post_type` | feed, feed+reel, reel, carousel, story, feed+story, feed+reel+story, reel+story, carousel+story, video, shorts |
| `post_video_title` | YouTube / LinkedIn |
| `scheduling` | `publish_type`: scheduled / draft / queued / content_category · `scheduled_at` (`YYYY-MM-DD HH:mm:ss`) |
| `first_comment` | `message` (≤2000), `accounts[]` (must be subset of main accounts) |
| `labels[]` | ≤20 label ids |
| `campaign_id` | campaign (folder) id |
| `approval` | `approvers[]`, `approve_option`: anyone / everyone, `notes` |
| `hide_client` | Hide draft from client-role users |

**Platform option blocks:**

| Block | Key fields |
|---|---|
| `gmb_options` | `topic_type` STANDARD/EVENT/OFFER, `start_date`, `end_date`, `title` (≤100), `action_type` BOOK/ORDER/LEARN_MORE, `cta_link` |
| `youtube_options` | `title` (≤100), `privacy_status`, `category` (15 enum values), `tags[]` (≤30 chars each), `license`, `made_for_kids` |
| `tiktok_options` | `privacy_level`, `disable_comment/duet/stitch`, `auto_add_music`, `brand_content_toggle`, `brand_organic_toggle`, `disclose_commercial_content`, `is_aigc` |
| `twitter_options` | `has_threaded_tweets`, `threaded_tweets[]` = `{message, media[]}` |
| `threads_options` | `has_multi_threads`, `multi_threads[]` = `{message, media[]}` |
| `pinterest_options` | `title`, `link` |
| `facebook_options` | `facebook_background_id`, `collaborators[]`, `carousel{is_carousel_post, cards[2–10]{image,title,description,link}, accounts[], call_to_action (32 enum values), end_card, end_card_url}` |
| `instagram_options` | `collaborators[]` (≤3 public IG usernames) |

**Rules the model must be told explicitly in the tool description:**
1. `post_type: "carousel"` **requires** `facebook_options.carousel.is_carousel_post = true` with ≥2 cards, or the API returns 400.
2. `first_comment.accounts` must be a subset of `accounts`.
3. `publish_type: "scheduled"` requires `scheduled_at` in **workspace timezone**, formatted `YYYY-MM-DD HH:mm:ss`.
4. `tiktok_options.is_aigc` must be `true` whenever the media was AI-generated by our own image/video tools. This is a TikTok platform requirement, not a preference. **The backend should force this, not trust the model.**
5. When `content_category_id` is set with `publish_type: "content_category"`, the category supplies both accounts and timing.

**Preview card is mandatory.** Before this executes the user sees: rendered post per platform, media thumbnails, target accounts, scheduled time in their timezone, first comment, labels/campaign, approval chain. Confirm / Edit / Cancel.

**Idempotency-Key required.**

### `posts_update` 🟡 — ⚠️ ENDPOINT DOES NOT EXIST (P0 gap)
No `PUT /posts/{post_id}` in the spec. Needed for: change caption, change time, add/remove account, swap media, add label, move to campaign, convert draft → scheduled.

Delete-and-recreate is not an acceptable workaround — it destroys the post ID, approval history and internal comment thread. **See [API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168), gap #1.**

### `posts_validate` 🟢 — ⚠️ ENDPOINT DOES NOT EXIST (P0 gap)
Dry-run validation before create. AI-written content will break platform rules routinely (caption length per platform, image counts, video duration, aspect ratio, carousel minimums, YouTube title length, GMB title length). Without this we discover failures at publish time — hours after the user walked away. **See [API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168), gap #3.**

### `posts_delete` 🔴
`DELETE /api/v1/workspaces/{workspace_id}/posts/{post_id}`

Confirmation must name the post ("Delete the 3pm Instagram post about the summer sale?"), never just the ID. Batch deletes list every item.

### `posts_approve` 🟡
`POST /api/v1/workspaces/{workspace_id}/posts/{post_id}/approval` · body `{action, comment}`

Only offered when the user is an assigned approver. Filter at the registry, not in the prompt.

### `posts_internal_notes` 🟡
| Action | Endpoint |
|---|---|
| list | `GET /workspaces/{id}/posts/{post_id}/comments` |
| add | `POST /workspaces/{id}/posts/{post_id}/comments` · `{comment, is_note, mentioned_users}` |

> ### ⚠️ Naming warning
> These are **internal team notes on a draft** — the collaboration thread. They are **not** public comments from the audience on a published post.
>
> Replying to real audience comments belongs to **Inbox** (Phase 3) and does not exist yet.
>
> The tool is named `posts_internal_notes`, **not** `posts_comments`, specifically to stop the model reaching for it when a user says "reply to the comments on my post." Please keep this naming through implementation.

---

## 4. Media Library

### `media_list` 🟢
`GET /api/v1/workspaces/{workspace_id}/media`

| Arg | Notes |
|---|---|
| `type` | images / videos |
| `search` | name, partial, case-insensitive |
| `sort` | recent / oldest / size / a2z / z2a |
| `page`, `per_page` | max 100 |

**Shaping:** return `id`, `name`, `type`, `url`, `thumbnail`, dimensions, `created_at`. Render as a thumbnail grid, not a text list — "use the third one" only works if the user can see three.

### `media_upload` 🟡
`POST /api/v1/workspaces/{workspace_id}/media` · multipart · `file` | `url` | `folder_id`

Two paths: user attaches a file in chat, or the AI generates an image/video and pushes it to the library. **Every AI-generated asset should land in the media library automatically** — otherwise it exists only in the chat transcript and cannot be reused. This also gives us the audit trail for `is_aigc`.

**Gap:** no delete, no folder list/create endpoint (`folder_id` is accepted but not discoverable). [API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168), gap #11.

---

## 5. Organisation

### `campaigns_manage` 🟡 (🟢 for list)
| Action | Endpoint |
|---|---|
| list | `GET /workspaces/{id}/campaigns` |
| create | `POST /workspaces/{id}/campaigns` · `{name, color}` both required |
| update | `PUT /workspaces/{id}/campaigns/{campaign_id}` |
| delete | `DELETE /workspaces/{id}/campaigns/{campaign_id}` 🔴 |

`color` is required on create. The model will forget. **Default it server-side** from a palette rather than failing the call.

### `labels_manage` 🟡 (🟢 for list)
Same shape: `GET/POST /workspaces/{id}/labels`, `PUT/DELETE /workspaces/{id}/labels/{label_id}`, `{name, color}` both required, same colour defaulting.

### `content_categories_list` 🟢
`GET /workspaces/{id}/content-categories` · `search` (comma-separated terms), `page`, `per_page`

**Read-only in the current API.** The chat can schedule *into* a category but cannot create or edit one. [API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168), gap #6.

---

## 6. Analytics — 97 endpoints, 9 tools

The compression that makes this feasible. Every endpoint takes the same core params:

`workspace_id` (path) · `platform_id` (query, **required**, single account) · `start_date` · `end_date` · `timezone` (IANA, default UTC) · optionally `date` as `"YYYY-MM-DD - YYYY-MM-DD"`

### 6.1 `analytics_summary` 🟢
`GET /workspaces/{id}/analytics/{platform}/summary`

Args: `platform` (enum, 8 values), `account_id`, `start_date`, `end_date`, `timezone`, `board_id` (Pinterest only)

Returns `current` / `previous` / `difference` / `percentage` blocks. Built-in period comparison — use it rather than making two calls.

### 6.2 `analytics_trend` 🟢
One tool, `(platform, metric)` → endpoint. Backend owns the mapping table:

| `metric` | FB | IG | LI | X | TT | YT | PIN | GMB |
|---|:--:|:--:|:--:|:--:|:--:|:--:|:--:|:--:|
| `audience_growth` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | — |
| `engagement` | ✅ | ✅ | — | ✅¹ | ✅ | ✅ | ✅ | — |
| `impressions` | ✅ | ✅ | — | ✅¹ | — | — | ✅ | ✅² |
| `page_views` | — | — | ✅ | — | — | — | — | — |
| `views` | — | — | — | — | — | ✅ | — | — |
| `watch_time` | — | — | — | — | — | ✅ | — | — |
| `actions` | — | — | — | — | — | — | — | ✅ |
| `publishing_behaviour` | ✅ | ✅ | ✅ | — | ✅ | — | ✅³ | ✅ |
| `reels_performance` | ✅ | ✅ | — | — | — | — | — | — |
| `stories_performance` | — | ✅ | — | — | — | — | — | — |
| `video_insights` | ✅ | — | — | — | — | — | — | — |
| `performance_schedule` | — | — | — | — | — | ✅ | — | — |
| `media_activity` | — | — | — | — | — | — | — | ✅ |
| `pin_performance` | — | — | — | — | — | — | ✅ | — |

¹ X combines both in `/engagement-impression` · ² GMB `/impressions` splits by channel and device · ³ Pinterest calls it `pin-posting`

Extra arg: `granularity` = `cumulative` | `daily`. YouTube and Pinterest expose both as separate endpoints (`/engagement` vs `/engagement-daily`). The model must never have to know that — **default to `daily`**, since "how did we grow last week" almost always means daily deltas, and cumulative charts mislead.

Also accepts `media_type` where the endpoint supports it (FB, IG, LI publishing-behaviour).

**Unsupported combination → return a clean error, not an empty chart:**
`"LinkedIn does not report impressions over time. Available for LinkedIn: audience growth, page views, publishing behaviour, hashtags, top posts."` The model can then offer the alternative in the same turn.

### 6.3 `analytics_breakdown` 🟢
Non-time-series slices.

| `dimension` | Endpoint | Platforms |
|---|---|---|
| `demographics` | `/{platform}/demographics` | FB (age/gender/country/city), IG (age/gender), LI (industry/country) |
| `location` | `/{platform}/audience-location` | FB, IG |
| `active_hours` | `/{platform}/active-users` | **FB, IG only** |
| `hashtags` | `/{platform}/hashtags` | IG, LI |
| `search_keywords` | `/gmb/search-keywords` | GMB |
| `traffic_sources` | `/youtube/find-video` | YT |
| `sharing_platforms` | `/youtube/video-sharing` | YT |
| `posts_per_day` | `/linkedin/posts-per-days` | LI |
| `demographics_overview` | `/facebook/overview-demographics` | FB |

`active_hours` existing only for Facebook and Instagram is the reason we cannot honestly answer "best time to post" today. [API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168), gap #5.

### 6.4 `analytics_top_posts` 🟢

| Arg | Notes |
|---|---|
| `platform`, `account_id`, dates | standard |
| `direction` | `top` \| `least` — TikTok and Pinterest return both in one call; X and YouTube have separate `least-*` endpoints; **FB, IG, LI, GMB have no "least" endpoint** |
| `order_by` | FB, IG, LI, X, YT, TikTok (`sort_order`) |
| `media_type` | FB, IG, LI, YT, PIN |
| `hashtags` | IG, LI |
| `entity_type` | IG |
| `tweet_type` | X |
| `topic_type` | GMB |
| `limit` | default 15, max 100 — **cap at 10 for chat** |
| `offset` | pagination |

Endpoint selection: `/top-posts` or `/sorted-top-posts` (use sorted when `order_by`/filters are supplied) · X → `/top-tweets` or `/least-tweets` · Pinterest → `/top-pins` · YouTube → `/top-posts` or `/least-posts`.

### 6.5 `analytics_get_post` 🟢
Single-item deep dive. `/facebook/post` · `/instagram/post` · `/linkedin/post` · `/tiktok/post` · `/gmb/post` · `/twitter/tweet` · `/youtube/video` · `/pinterest/pin` — all take `post_id`.

Powers "why did this one do so well?"

### 6.6 `analytics_ai_insights` 🟢
`GET /{platform}/ai-insights` · `type`, `limit` (default 5), `language` (ISO 639-1, default `en`)

Available: FB, IG, LI, TikTok, YouTube, Pinterest, GMB. **Not X.**

**Design question for the team:** this is our existing AI insight engine. When the chat asks for insights, do we surface these, or does the chat model generate its own from raw data? Doing both produces two different answers to the same question in the same product. Recommendation: use these as *input* to the chat model, and never present them as a separate competing opinion.

### 6.7 `analytics_workspace_overview` 🟢 — ⚠️ COMPOSITE, MUST BE BUILT
No endpoint exists. Backend fans out `/summary` across every connected account, normalises metric names across platforms, and returns one shaped result.

Answers: *"How did we do last week?"* · *"Give me a snapshot across everything."*

Without it this question is 12+ sequential calls and ~30 seconds. **This is the difference between a demo and a product.** [API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168), gap #4.

Normalisation is the hard part: `fan_count` (FB) / followers (IG) / subscribers (YT) / connections (LI) all become `audience`. The mapping table needs product sign-off, not just engineering judgement — leadership will read these numbers.

### 6.8 `analytics_compare_accounts` 🟢 — ⚠️ COMPOSITE, MUST BE BUILT
Fans out and **ranks**. Args: `account_ids[]` (default all), `metric`, `start_date`, `end_date`.

Answers: *"Which account is doing best?"* · *"Which is underperforming?"* · *"Rank my accounts by engagement rate."*

Must return normalised **rates**, not just raw totals. An account with 500k followers will always win on raw engagement. Engagement *rate* is the question actually being asked.

### 6.9 `reviews_list` 🟢
`GET /workspaces/{id}/analytics/gmb/reviews` — ratings, distribution, daily activity.

**Correction (v1.1): replies are supported.** `POST /workspaces/{id}/inbox/reviews/{review_id}/reply` and `DELETE .../reply` are both live. Pair this read tool with an `inbox_reply_to_review` write tool — the Review Watcher agent in [Agents & Skills](https://app.helpin.ai/share/f4cb43a256cbad3ee529632028d38f30) is buildable end-to-end today. [API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168), gap #9 is resolved.

---

## 7. Team & Workspace

### `team_manage` 🟡 (🔴 for remove)
| Action | Endpoint |
|---|---|
| list | `GET /workspaces/{id}/team-members` |
| invite | `POST /workspaces/{id}/team-members` · `{role*, email*, membership, permissions}` |
| update | `PUT /workspaces/{id}/team-members/{member_id}` |
| remove | `DELETE /workspaces/{id}/team-members/{member_id}` 🔴 |

Admin-only. Invites and removals always confirm — an accidental removal is not recoverable through the chat.

### `workspaces_manage` 🟡 (🔴 for delete)
| Action | Endpoint |
|---|---|
| list | `GET /workspaces` |
| create | `POST /workspaces` · `{name*, logo*, timezone*, super_admin_id, note, instagram_posting_method, first_day}` |
| update | `PUT /workspaces/{workspace_id}` |
| delete | `DELETE /workspaces/{workspace_id}` 🔴 |

`logo` is required on create — the model cannot invent one. Either default a generated placeholder server-side or have the UI prompt for upload. Do not let this fail silently.

**Workspace delete should arguably not be a chat tool at all.** It is irreversible, destroys everything, and nobody needs conversational convenience for it. Recommend excluding from the registry entirely in v1.

---

## 8. Inbox — Phase 3, **NOT blocked** (28 endpoints are live)

**Corrected in v1.1.** The Inbox API shipped. It lives at `workspaces/{workspace_id}/inbox/*`, proxied to `social-inbox-manager`. Phase 3 can start now — and given Phase 2's two P0 gaps (update-post, validate) do not exist yet, Inbox is currently the shortest path to new user-visible capability.

**The v1.0 six-tool design assumed the wrong shape.** The API is organised as one unified `items` triage surface plus per-type action groups. Re-cut to match:

| Tool | Maps to | Tier |
|---|---|---|
| `inbox_list_items` 🟢 | `GET inbox/items` — filters: `inbox_type`/`inbox_types`, `platform_id`, `platform_type`, `element_type`, `assigned`, `assigned_to`, `marked_done`, `archived`, `tags`, `search_term`, `conversation_id`, `page`, `limit` | 🟢 |
| `inbox_summary` 🟢 | `GET inbox/items/summary` — counts for "what's waiting on me" | 🟢 |
| `inbox_get_thread` 🟢 | `GET inbox/conversations/{id}/messages` (DMs) · `GET inbox/posts/{post_id}/comments` (comments) | 🟢 |
| `inbox_reply` 🟡 | `POST inbox/conversations/{id}/messages` · `POST inbox/conversations/{id}/messages/media` · `POST inbox/posts/{post_id}/comments` · `POST inbox/reviews/{review_id}/reply` | 🟡 |
| `inbox_triage` 🟡 | `POST inbox/items/{id}/` → `done` · `pending` · `read` · `archive` | 🟡 |
| `inbox_assign` 🟡 | `POST inbox/items/{id}/assignee` | 🟡 |
| `inbox_tags` 🟡 | `POST inbox/items/{id}/tags` · `GET`/`POST`/`PATCH`/`DELETE inbox/tags` · `POST inbox/tags/merge` | 🟡 |
| `inbox_notes` 🟡 | `GET`/`POST inbox/items/{id}/notes` | 🟡 |
| `inbox_moderate_comment` 🔴 | `DELETE inbox/comments/{id}` · `POST inbox/comments/{id}/visibility` · `POST inbox/comments/{id}/like` | 🔴 |

**Naming collision to watch.** `inbox_get_thread`/`inbox_reply` on `posts/{post_id}/comments` handle **audience** comments on published posts. `posts_internal_notes` (§3) handles **internal team notes on drafts**. Both are now real, live, and one endpoint path away from each other. The §3 naming warning is more important than when it was written, not less.

**Requirements status** (was a wishlist, now audited — see [API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168) Gap 2):
1. ~~Sentiment and intent fields per message~~ — **still missing.** Gap 2a, P1.
2. ~~Bulk endpoints~~ — **downstream already accepts `element_ids[]`; only the public route is single-item.** Route-signature change, not new plumbing. Gap 2c, S effort.
3. ~~Filter by "needs reply" server-side~~ — **still missing.** Gap 2b, P1.
4. ~~A single conversation model~~ — **satisfied.** `items` is exactly this.
5. **Idempotency on reply** — still required. Infra A. A double-sent DM is a visible, embarrassing failure.

---

## 9. Summary table

**Exists today** column added in v1.1 — verified against `contentstudio-ai-agents/src/orchestration/mcp_tools.py`.

| # | Tool | Tier | Phase | Endpoints wrapped | Exists today |
|---|---|---|---|---|---|
| 1 | `context_get` | 🟢 | 1 | 6 | partial — via `mcp_select_workspace` + fetches |
| 2 | `accounts_list` | 🟢 | 1 | 1 | ✅ `mcp_fetch_accounts` |
| 3 | `platforms_list` | 🟢 | 1 | 1 | — |
| 4 | `accounts_connect` | 🔴 | 2 | 1 | — |
| 5 | `accounts_add_credentials` | 🔴 | 2 | 2 | — |
| 6 | `posts_list` | 🟢 | 1 | 1 | ✅ `mcp_fetch_posts` |
| 7 | `posts_create` | 🟡 | 2 | 1 | ✅ `mcp_schedule_post` (full workflow) |
| 8 | `posts_update` | 🟡 | 2 | **0 — build** | — |
| 9 | `posts_validate` | 🟢 | 2 | **0 — build** | — |
| 10 | `posts_delete` | 🔴 | 2 | 1 | ✅ `mcp_delete_post` (confirmation flow) |
| 11 | `posts_approve` | 🟡 | 2 | 1 | ✅ `mcp_approve_post` |
| 12 | `posts_internal_notes` | 🟡 | 2 | 2 | ✅ `mcp_fetch_comments` + `mcp_add_comment` |
| 13 | `media_list` | 🟢 | 1 | 1 | — |
| 14 | `media_upload` | 🟡 | 2 | 1 | — |
| 15 | `campaigns_manage` | 🟡 | 2 | 4 | partial — `mcp_fetch_campaigns` (read only) |
| 16 | `labels_manage` | 🟡 | 2 | 4 | partial — `mcp_fetch_labels` (read only) |
| 17 | `content_categories_list` | 🟢 | 1 | 1 | ✅ `mcp_fetch_content_categories` |
| 18 | `analytics_summary` | 🟢 | 1 | 8 | — |
| 19 | `analytics_trend` | 🟢 | 1 | ~45 | — |
| 20 | `analytics_breakdown` | 🟢 | 1 | ~15 | — |
| 21 | `analytics_top_posts` | 🟢 | 1 | ~18 | — |
| 22 | `analytics_get_post` | 🟢 | 1 | 8 | — |
| 23 | `analytics_ai_insights` | 🟢 | 1 | 7 | backend agents exist, not exposed to chat |
| 24 | `analytics_workspace_overview` | 🟢 | 1 | **0 — expose** | `overview/*` agents exist (6) |
| 25 | `analytics_compare_accounts` | 🟢 | 1 | **0 — expose** | partial, same agents |
| 26 | `reviews_list` | 🟢 | 1 | 1 | — |
| 27 | `team_manage` | 🟡 | 2 | 4 | partial — `mcp_fetch_team_members` (read only) |
| 28 | `workspaces_manage` | 🟡 | 2 | 4 | partial — `mcp_fetch_workspaces`, `mcp_select_workspace` |
| 29–37 | `inbox_*` (9 tools, re-cut in §8) | 🟢🟡🔴 | 3 | 28 — **API is live** | — |
| — | `mcp_validate_token` | 🟢 | — | 1 | ✅ live, not in the original catalog |

**Revised Phase 1.** The original "14 read tools, ship to a beta cohort" framing is still the right gate, but 8 of those reads already ship. The real Phase 1 work is:

1. **Analytics chat tools (6)** — 99 endpoints and 86 narrative agents, zero chat exposure. This is the single largest block of unreachable capability in the product, and it maps directly to the "how did my profiles do last week" questions driving this project.
2. **`media_list`, `platforms_list`, `reviews_list`** — small, missing.
3. **`context_get`** — consolidate the existing scattered fetches into one cached bootstrap.
4. **Harden what exists** — envelope reconciliation, tier/state-machine mapping.

The beta-cohort recommendation stands and is now cheaper than originally costed.
