# ContentStudio AI Chat — As-Built Baseline (v1.0)

**Purpose:** what actually exists in the codebase today. The other planning docs in this collection were written from an exported OpenAPI spec without codebase access, and several of their premises are out of date. This document is the ground truth; where it disagrees with any of them, this wins.

**Verified:** 2026-07-28, by reading source in `contentstudio-backend/`, `contentstudio-ai-agents/`, `contentstudio-frontend/`.

---

## 1. Headline corrections

| The other docs said | Actually |
|---|---|
| 118 endpoints | **159 endpoints** in `routes/api/v1.php` |
| "No Inbox API" — P0 blocker, blocks Phase 3 | **Inbox is live: 28 public endpoints.** Phase 3 is not blocked |
| "Cannot reply to GMB reviews" | **`POST /inbox/reviews/{review_id}/reply` exists**, plus delete-reply |
| Tools to be built from zero | **14 platform tools already live** in the chat's Operations agent |
| (not mentioned) | A **second, hosted MCP server** exists in the backend at `/api/mcp` with 8 tools |
| (not mentioned) | **86 analytics narrative agents** already exist, dashboard-bound |
| Inbox bulk actions = L effort | Downstream service already takes `element_ids[]`. The public route exposes single-item only. **Route-signature change, not new plumbing** |

Two claims held up exactly as written, and both remain P0: **no update-post endpoint** and **no post validation endpoint**.

---

## 2. Public API — 159 endpoints

Source: [routes/api/v1.php](../../contentstudio-backend/routes/api/v1.php). Middleware on every route: `force.json`, `api.key`, `api.request.log`, `throttle:api-v1`. Workspace-scoped routes add `PermissionMiddleware`.

| Domain | Count | Coverage |
|---|---:|---|
| Analytics (8 platforms) | **99** | FB 15 · IG 15 · YT 19 · LI 11 · TikTok 8 · Pinterest 14 · X 7 · GMB 10 |
| Inbox | **28** | items 10 · conversations 6 · comments 5 · tags 5 · reviews 2 |
| Posts | 6 | list, create, delete, approval, notes list, notes add |
| Team members | 4 | list, invite, update, remove |
| Labels | 4 | full CRUD |
| Campaigns | 4 | full CRUD |
| Workspaces | 4 | list, create, update, delete |
| Account connection | 4 | `platforms`, `connect/{platform}`, `add/bluesky`, `add/facebook-group` |
| Media | 2 | list, upload |
| Content categories | 1 | list only |
| Accounts | 1 | list only |
| `me` | 1 | |
| `facebook/text-backgrounds` | 1 | |

Analytics is 62% of the surface, which is what makes [Tool Catalog](https://app.helpin.ai/share/9a2e81a1a3ff56c267ca61d0c50e5217)'s collapse-to-6-tools decision correct. That call stands.

### 2.1 Inbox, in detail

Live at `workspaces/{workspace_id}/inbox/*`, proxied to `social-inbox-manager`. Controllers: `app/Http/Controllers/Api/V1/Inbox/`.

- **Items (unified triage):** `GET items`, `GET items/summary`, then `POST items/{id}/` → `done`, `pending`, `read`, `archive`, `assignee`, `tags`; `GET`/`POST items/{id}/notes`
- **Conversations (DMs):** `GET`/`POST messages`, `POST messages/media`, `GET bookmarks`, `POST messages/{id}/bookmark`, `DELETE messages/{id}`
- **Post comments:** `GET`/`POST posts/{post_id}/comments`, `DELETE comments/{id}`, `POST comments/{id}/visibility`, `POST comments/{id}/like`
- **Reviews:** `POST reviews/{review_id}/reply`, `DELETE reviews/{review_id}/reply`
- **Tags:** list, create, merge, update, delete

Available filters on `GET items`: `inbox_type`/`inbox_types`, `platform_id`, `platform_type`, `element_type`, `assigned`, `assigned_to`, `marked_done`, `archived`, `tags`, `search_term`, `conversation_id`, `page`, `limit`.

**This satisfies [API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168)'s requirement #4 (one unified conversation model) already.** `items` is the single triage surface across DMs, comments and reviews. The 6-tool Inbox design in [Tool Catalog](https://app.helpin.ai/share/9a2e81a1a3ff56c267ca61d0c50e5217) §8 should be re-cut against this shape rather than the assumed one.

**Still genuinely missing for AI use:** no sentiment or intent field on inbox items (sentiment exists only in `app/Data/Listening/*` and Facebook analytics), and no server-side "needs reply" filter. Both remain real gaps.

---

## 3. Three tool surfaces, not one

Any new tool has up to three places to land. The other planning docs only account for the first.

### 3.1 AI chat — Operations agent (the live tool column)

`contentstudio-ai-agents/src/orchestration/mcp_tools.py` → `MCP_TOOLS`, registered on the `Operations` agent in `src/orchestration/agents.py:2494`.

| # | Tool | Maps to |
|---|---|---|
| 1 | `mcp_fetch_workspaces` | `GET /workspaces` |
| 2 | `mcp_fetch_accounts` | `GET /workspaces/{id}/accounts` |
| 3 | `mcp_fetch_content_categories` | `GET /content-categories` |
| 4 | `mcp_fetch_team_members` | `GET /team-members` |
| 5 | `mcp_fetch_labels` | `GET /labels` |
| 6 | `mcp_fetch_campaigns` | `GET /campaigns` |
| 7 | `mcp_fetch_posts` | `GET /posts` |
| 8 | `mcp_fetch_comments` | `GET /posts/{id}/comments` |
| 9 | `mcp_schedule_post` | `POST /posts` (multi-step workflow) |
| 10 | `mcp_delete_post` | `DELETE /posts/{id}` (confirmation workflow) |
| 11 | `mcp_approve_post` | `POST /posts/{id}/approval` |
| 12 | `mcp_add_comment` | `POST /posts/{id}/comments` |
| 13 | `mcp_select_workspace` | session state |
| 14 | `mcp_validate_token` | `GET /me` |

Plus `improve_ai_library_post` in `AI_LIBRARY_TOOLS`.

Full team (`src/orchestration/agents.py`): Caption Writer, Image Generator, Video Generator, Image Analyzer, General Assistant, **Operations**, Post Improver, Post Plan Builder — coordinated by the `ContentStudio AI` team in `src/orchestration/team.py`.

**Existing conventions any new tool must follow** — these already answer several open questions in [Tool Catalog](https://app.helpin.ai/share/9a2e81a1a3ff56c267ca61d0c50e5217) §0:

- **Response envelope is already fixed.** `_envelope()` emits `{mcp_operation, operation, success, message, workflow_state, next_action, data, followup_actions, error, streamed_chunks}`. the Tool Catalog proposes a different envelope (`{ok, summary, data, render, next_cursor, warnings}`). **the Tool Catalog's envelope is not what ships today** — either migrate deliberately or adopt the existing one.
- **`workspace_id` is already never model-supplied**, resolved by `_mcp_resolve_workspace()`. the Tool Catalog §0's rule is already enforced.
- **Confirmation flows already exist** as a state machine, not a per-tool flag. Steps: `resolving` → `awaiting_accounts` → `awaiting_content` → `content_confirmation` → `awaiting_approvers` → `confirmation` → `complete`/`cancelled`. the Tool Catalog's 🟢/🟡/🔴 tiers need mapping onto this, not building fresh.
- **Tools take "hallucination shim" kwargs** (`platform`, `query`, `workspace`, `filter`, …) accepted and ignored, with `strict=False`. Real parameter extraction happens in `UnifiedIntentDetector`, not from the tool signature. New tools must follow this or Groq's router will fail them.
- **Adding an operation touches 9 places** — the checklist is in `contentstudio-ai-agents/CLAUDE.md` under "Adding New MCP Operations". Budget for it.

### 3.2 Backend hosted MCP server

`routes/mcp.php` → `/api/mcp` (base, `/sse`, `/message`) via `CustomMcpController`, plus `php-mcp/laravel` discovery over `app/Mcp/`.

8 tools: `fetch_workspaces`, `fetch_social_accounts`, `fetch_posts`, `create_post`, `delete_post`, `validate_token`, `help`, `ping`. One resource (`WorkspacesResource`), one prompt (`GettingStartedPrompt`).

This is a **narrower, separate implementation** of the same capability as §3.1. It is not mentioned anywhere in the other planning docs. Decide explicitly whether new tools land here too, or whether this server is deprecated in favour of the npm one.

### 3.3 npm MCP server + CLI

`contentstudio-mcp` and `contentstudio-cli` on npm — external repos, not mounted here. See the public-dev-surfaces reference for what is GA.

**Net: the same tool can need implementing up to three times.** [Tool Catalog](https://app.helpin.ai/share/9a2e81a1a3ff56c267ca61d0c50e5217)'s per-tool effort estimates assume one. Worth an explicit decision on which surfaces are in scope before sizing the work.

---

## 4. Analytics AI already exists — 86 agents

`contentstudio-ai-agents/src/agents/analytics/`:

| Platform | Agents | | Platform | Agents |
|---|---:|---|---|---:|
| Facebook | 14 | | Pinterest | 10 |
| LinkedIn | 14 | | TikTok | 10 |
| Instagram | 13 | | YouTube | 10 |
| GMB | 5 | | **Overview (cross-account)** | **6** |
| Meta Ads | 2 | | Google Ads | 2 |

These generate the narrative insight text behind the 7 `analytics/{platform}/ai-insights` endpoints. They are **dashboard-bound**: reached through `src/api/routers/analytics/*`, not exposed as chat tools.

Two consequences the docs miss:

1. **[Tool Catalog](https://app.helpin.ai/share/9a2e81a1a3ff56c267ca61d0c50e5217) §6.6's "design question" is already decided by the architecture.** The insight engine exists and is per-metric. The chat should call these rather than re-deriving narratives from raw numbers, or the product will give two different answers to the same question.
2. **[API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168) Gap 4's hardest part is partly built.** `src/agents/analytics/overview/` already does cross-account aggregation — `overview_account_performance`, `overview_account_statistics`, `overview_engagement_performance`, `overview_impressions_performance`, `overview_reach_performance`, `overview_top_posts`. The cross-platform metric normalisation Gap 4 calls for exists in some form here and in the internal API. **Re-scope Gap 4 from "build" to "expose and reconcile"** before sizing it.

### 4.1 New gap the register misses: Ads analytics

Meta Ads and Google Ads analytics agents exist (4 agents, with `schemas.py` and dedicated routers), but there are **zero ads endpoints in the public API**. So ads is unreachable from chat, CLI or MCP. Not in [API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168)'s register at all. Suggest adding as P2 — it is a paid-media surface competitors mostly lack.

---

## 5. Other existing pieces worth knowing

- **Content generation tools** (already in chat, as the user noted): `tool_generate_image`, `tool_generate_image_set`, `tool_transform_image`, `tool_generate_video`, `tool_enhance_image_prompt`, `tool_analyze_image`, `tool_describe_image`, `tool_generate_alt_text`. Plus a dedicated-tools engine (`src/tools/dedicated/`) with `creative_image_tool` and `video_tool` specs.
- **Web/research tools:** `tool_exa_answer`, `tool_find_similar`, `tool_get_contents` — Exa-backed. Relevant to [API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168) Gap 8: the chat can already research the open web, so "find trending topics" is partly reachable today. What is missing is ContentStudio's own **Discover** index, which remains genuinely absent from the public API. **Gap 8 stands, and the pull-forward recommendation stands.**
- **Discovery + listening agents:** `src/agents/discovery/` (brand_context, source_candidates, topic_tagger, topic_taxonomy) and `src/agents/listening/` (brand_discovery, mention_analyzer). Internal only.
- **Frontend chat UI:** `contentstudio-frontend/src/modules/AI-tools/` — `AIChatMain.vue`, `AIChatModal.vue`, `AIChatWidget.vue`, `ChatBox.vue`, plus `SavedPrompts*` and `AddCustomPromptModal.vue`. **Saved prompts already ship**, which is the nearest existing thing to [Agents & Skills](https://app.helpin.ai/share/f4cb43a256cbad3ee529632028d38f30)'s Skills concept — build Skills on top of it rather than beside it.
- **Post Plan Builder** agent + `src/orchestration/post_plan/` — an existing multi-post planning flow that overlaps [Use Case Catalog](https://app.helpin.ai/share/03a7cf23c93af70ebde7a7410d4e6e5e)'s campaign workflows. Check for overlap before specifying new ones.

---

## 6. What this changes about the plan

Ordered by how much it moves the work.

1. **Phase 1 is largely already shipped.** 8 of [Tool Catalog](https://app.helpin.ai/share/9a2e81a1a3ff56c267ca61d0c50e5217)'s 14 Phase 1 read tools exist as `mcp_fetch_*`. Phase 1 should be re-scoped from "build 14 read tools" to "add the missing reads — analytics and inbox — and harden what's there." The suggested beta-cohort gate is still right, just cheaper than costed.
2. **Phase 3 (Inbox) is unblocked and could move ahead of Phase 2.** 28 endpoints are live. Phase 2's two hardest dependencies (update-post, validate) do not exist yet, so Inbox is now the shortest path to new user-visible capability.
3. **Analytics is the real Phase 1 gap, not posts.** 99 endpoints, 86 narrative agents, zero chat tools. This is the largest single block of unexposed capability in the product and it maps directly to the user's "how did my profiles do last week" examples.
4. **Re-cut the 6 Inbox tools** in [Tool Catalog](https://app.helpin.ai/share/9a2e81a1a3ff56c267ca61d0c50e5217) §8 against the actual items/conversations/comments/reviews/tags shape.
5. **Reconcile the response envelope** — [Tool Catalog](https://app.helpin.ai/share/9a2e81a1a3ff56c267ca61d0c50e5217)'s proposal vs the shipping `_envelope()`. Pick one before any tool is written.
6. **Decide the surface question** — chat only, or chat + hosted MCP + npm MCP/CLI. This is the biggest uncosted variable in the plan.

Unchanged and still correct: the 32-tools-not-159 compression, `posts_internal_notes` naming (the trap is real and now sharper, since Inbox comment-reply genuinely exists and the two will collide), no-update-post as P0, no-validate as P0, Discover as the strategic gap versus Vista.
