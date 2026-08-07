# Bundle Context — ContentStudio AI Chat Tool Layer

> ### ⚠️ Not the project CLAUDE.md, and partly superseded
> This was the `CLAUDE.md` from the claude.ai session that produced docs 01–05. It is kept here as a record of that session's decisions. **It is not loaded as project instructions** — the repo root `CLAUDE.md` governs, and where the two disagree, the repo wins.
>
> **Two kinds of correction apply:**
>
> **1. Facts (see [06-As-Built-Baseline.md](06-As-Built-Baseline.md)):**
> - 118 endpoints → **159**
> - "No Inbox API" → **live, 28 endpoints.** Phase 3 unblocked
> - "33 tools to build" → **14 already ship** in the chat's Operations agent
> - The success/error envelopes below are **a proposal, not what ships.** Live shape is `_envelope()` in `contentstudio-ai-agents/src/orchestration/mcp_tools.py`
> - Tool naming: live tools use `mcp_*`, not `domain_action`
>
> **2. House conventions — this repo differs:**
> - **Not Jira.** Work is tracked in **Helpin** (helpin.ai). No Shortcut or Jira references in stories
> - **Story rules live in `docs/story-guidelines.md`** — read that, not the 9-section list below (the two mostly agree; the guidelines are authoritative)
> - **No trailing "fields"/metadata block** on stories
> - **No em dashes in user-facing UI copy**
> - **`[Flutter]`** prefix for mobile, not `[iOS]`/`[Android]` — the native apps are merging into one Flutter app
> - A **`[Design]`** story is required for any FE work

## What this project is

Turning ContentStudio's AI chat from a talk-only assistant into an operator that can
act on the user's workspace. The chat calls tools; tools are thin wrappers over the
existing public REST API. Later the same tool layer powers Agents (scheduled jobs)
and Skills (saved multi-step playbooks).

**Core principle: no duplicated business logic.** Chat, Agents, Skills, MCP and CLI
are all thin wrappers over the same REST API. If the API can't do it, the chat can't
either — see `docs/ai-chat/05-API-Gap-Register.md`.

## Read these first

| File | What's in it |
|---|---|
| `docs/ai-chat/01-AI-Chat-Master-Plan.md` | Strategy, architecture decisions, phasing, competitive read |
| `docs/ai-chat/02-Tool-Catalog.md` | **The spec.** ~33 tools, each mapped to exact API endpoints |
| `docs/ai-chat/03-Use-Case-Catalog.md` | ~70 user scenarios → tool chains → edge cases |
| `docs/ai-chat/04-Agents-and-Skills.md` | Phase 4 & 5 |
| `docs/ai-chat/05-API-Gap-Register.md` | What the API can't do yet, prioritised |
| **`docs/ai-chat/06-As-Built-Baseline.md`** | **Read first.** What actually exists in the codebase today |
| `docs/ai-chat/contentstudio-api-v1.json` | OpenAPI spec, 118 endpoints — **stale; the live API has 159.** Prefer `contentstudio-backend/routes/api/v1.php` |

## Non-negotiable decisions (do not relitigate without asking)

1. **~32 tools, not 159.** 99 of the 159 endpoints are analytics with an identical
   shape. They collapse into 6 parameterised tools via a `(platform, metric)` mapping
   table owned by the backend. See `02-Tool-Catalog.md` §6. *(Decision holds — and the
   86 per-metric analytics agents already in `contentstudio-ai-agents` are the strongest
   argument for it.)*

2. **`workspace_id` is never a model argument.** Always injected from session context.
   The model must not be able to act on the wrong workspace by hallucinating an ID.

3. **Permissions are enforced at the tool registry, not in the prompt.** Filter the
   tool list per user, per workspace, at session start. Never rely on system-prompt
   instructions to withhold a capability.

4. **Safety tiers on every tool:** 🟢 auto-run (reads) · 🟡 preview + confirm (writes)
   · 🔴 explicit confirm (destructive). Batch operations always confirm, once, listing
   every affected item.

5. **Agents can never call 🔴 tools.** Intersect the agent's allowlist with a
   hard-coded non-destructive set before execution.

6. **Idempotency-Key on every write.** Chat retries; without it a network blip
   publishes twice.

7. **Tool responses are shaped, not proxied.** Cap ~2000 tokens, top-N defaults,
   round numbers, strip nulls, attach a `render` directive for the UI component.

8. **Ship Phase 1 (read-only tools) alone first.** It can't break anything and it's
   the cheapest way to measure tool-selection accuracy before trusting writes. *(Still
   right, and cheaper than costed — 8 of the 14 reads already ship. Phase 1's real work
   is the 6 analytics tools plus the 2 composites.)*

## Naming trap

`/api/v1/workspaces/{id}/posts/{post_id}/comments` is **internal team notes on a
draft** (has `is_note`, `mentioned_users`). It is **NOT** public audience comments on
a published post.

- Tool name is `posts_internal_notes` — keep it.
- Audience comments belong to Inbox, and **that API is live**:
  `/api/v1/workspaces/{id}/inbox/posts/{post_id}/comments` → `inbox_reply_to_comment`.
- Users will say "reply to the comments on my post" and mean the second one.

**This trap is now sharper.** Both endpoints exist, and their paths differ by a single
`inbox/` segment while meaning completely different things. Getting the tool names
wrong will send internal draft notes where a customer reply was intended.

## Response envelopes

Success:
```json
{
  "ok": true,
  "summary": "Short natural-language result the model can quote",
  "data": {},
  "render": { "component": "post_preview", "props": {} },
  "next_cursor": null,
  "warnings": []
}
```

Error — `suggested_tool` lets the model self-recover in one turn:
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

## Conventions

- Tool names: `domain_action`, snake_case, domain first (`posts_create`,
  `analytics_trend`) so related tools cluster in the model's attention.
- Dates: resolve in **workspace timezone**, never UTC. `scheduled_at` format is
  `YYYY-MM-DD HH:mm:ss`.
- API auth: `X-API-Key` header. Base URL `https://api.contentstudio.io`.

## Known blockers (P0 — don't design around them silently, flag them)

| Gap | Impact | v1.1 status |
|---|---|---|
| No `PUT /posts/{id}` | "Change the caption" / "move it to 5pm" impossible without destroying post ID, approval state and comment thread | **Confirmed P0** |
| No `POST /posts/validate` | AI-written posts break platform rules; we find out at publish time | **Confirmed P0** |
| No cross-account analytics rollup | "How did we do last week?" = 12+ sequential calls, ~30s | **P1** — `src/agents/analytics/overview/` already aggregates; expose, don't build |
| ~~No Inbox API~~ | ~~Blocks all of Phase 3~~ | **RESOLVED** — 28 endpoints live |
| **No analytics or inbox chat tools** ⭐ | 127 of 159 endpoints unreachable from chat. The actual biggest gap, and it is on our side of the line, not the API's | **P0 for Phase 1** |

## User story format (if generating Jira stories)

Sections, in this order. **No Gherkin / Given-When-Then.**

1. Description
2. Workflow
3. Acceptance Criteria
4. Mock-ups
5. Impact on existing data
6. Impact on other products
7. Dependencies
8. Global quality & compliance checklist
9. Implementation references *(optional)*

## Style

Minimal, clean UI over elaborate solutions. Plain language in docs — anyone from any
department should be able to read them. Structured formats: tables over prose walls.
