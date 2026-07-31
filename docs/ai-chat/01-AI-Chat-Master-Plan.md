# ContentStudio AI Chat — Tools, Agents & Skills
## Master Plan (v1.1)

**Owner:** Ghulam Jaffar (Product)
**Status:** Draft for review — Engineering, Design, Leadership
**Related docs:** [Tool Catalog](https://app.helpin.ai/share/9a2e81a1a3ff56c267ca61d0c50e5217) · [Use Case Catalog](https://app.helpin.ai/share/03a7cf23c93af70ebde7a7410d4e6e5e) · [Agents & Skills](https://app.helpin.ai/share/f4cb43a256cbad3ee529632028d38f30) · [API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168) · [PRD](https://app.helpin.ai/share/c49875c01ee91933257622bac2311813) · [As-Built Baseline](https://app.helpin.ai/share/fff3d91449a0c4f64f84ac5eadc98f9d)

> **v1.1 corrections after codebase verification (2026-07-28).** Read [As-Built Baseline](https://app.helpin.ai/share/fff3d91449a0c4f64f84ac5eadc98f9d) alongside this. Three premises changed:
> 1. The API has **159 endpoints**, not 118.
> 2. **The Inbox API is live** (28 endpoints). Phase 3 is not blocked, and is now the shortest path to new capability.
> 3. **The tool layer is not new.** 14 platform tools already ship in the chat's Operations agent, plus a second hosted MCP server in the backend with 8 more.

---

## 1. What we are building, in one paragraph

Today ContentStudio's AI chat can talk, generate images and video, and run **14 platform operations** — list workspaces, accounts, posts, labels, campaigns, categories and team members; schedule, delete, approve and comment on posts. What it cannot do is the other two thirds of the product: **analytics (99 endpoints, zero chat tools)** and **inbox (28 endpoints, zero chat tools)**, plus media, connections and most write paths on the organisation objects. This plan completes that action layer, then reuses it to power **Agents** (things that run on a schedule without being asked) and **Skills** (saved multi-step playbooks the user fires with a single slash command).

The whole thing rests on one foundation we already own: **the public API**. No new business logic. The chat is a thin conversational wrapper over the same REST endpoints, exactly the pattern already agreed for MCP and CLI.

```
                    ┌─────────────────────────────┐
   Chat  ───────────┤                             │
   Agents ──────────┤   TOOL LAYER (14 live)      ├──── ContentStudio REST API
   Skills ──────────┤   ~32 tools, one registry   │      (existing, 159 endpoints)
   MCP / CLI ───────┤                             │
                    └─────────────────────────────┘
```

Everything in this plan is a wrapper. If a capability does not exist in the API, it does not exist in the chat — which is why **Section 7 (gaps)** matters as much as the tool list.

**One structural caveat the original plan missed.** "One registry" is the goal, not the current state: the chat's Operations agent and the backend's hosted `/api/mcp` server are two separate implementations of overlapping tools, and the npm MCP server and CLI are a third. Consolidating them, or explicitly scoping three of them, is a prerequisite for the diagram above being true.

---

## 2. The three layers (and why they are different products)

Vista Social has shipped exactly this three-layer model. It is the right model. But the three layers solve different jobs and should not be built at once.

| Layer | What it is | When it runs | User effort | Our phase |
|---|---|---|---|---|
| **Chat + Tools** | Ask for anything, AI does it now | On demand | High — user must ask each time | Phase 1–3 |
| **Agents** | A saved job that runs by itself | On a schedule / on a trigger | Set up once, then zero | Phase 4 |
| **Skills** | A saved multi-step recipe | On demand, one command | One command instead of ten | Phase 5 |

**The dependency is strict.** Agents are "a chat prompt on a cron with a tool allowlist." Skills are "a chat prompt with a fixed step sequence." Neither can exist before the tool layer is solid and safe. Building agents on a shaky tool layer means an agent silently publishing the wrong thing at 3am — the single worst outcome available to us.

**Sequence: Tools → prove reliability → Agents → Skills.**

---

## 3. Competitive read — Ask Vista

What they shipped, and what it tells us.

**What is genuinely good:**
- The chat lives **inside the composer**, not in a separate corner. It is where the work already happens.
- Agents are given **names and faces** (Theo, Maeve, Leo, Arjun, Lucia, Mira). This is not decoration — it turns an automation into "a teammate," which is a much easier thing to sell and to trust.
- Every agent shows **"Saves ~X hrs/month."** They quantified the value on the card itself.
- Agents show **live status** ("Going through inbox…", "Next run in 2 hrs"). The automation is visible, not a black box.
- Skills are explicitly positioned as **"not a prompt"** — they advertise the step count (4–8 steps) to signal depth.
- Agent creation is **3 steps: describe → confirm → runs.** Natural language in, structured job out.
- Default answer on safety: **"Only if you let it."** Nothing publishes without approval unless you opt in.

**What we can beat them on:**

| Advantage | Why it is ours |
|---|---|
| **Discover** | We have a real content discovery engine (Feedly/Inoreader class). Vista has nothing comparable. "Find trending topics in my niche and build a campaign from them" is a workflow only we can run end to end. |
| **Approval workflows** | Our post approval model (approvers, anyone/everyone, notes) is richer. An AI that respects a real approval chain is an agency selling point. |
| **API-first / MCP / CLI** | Already on our roadmap. Same tool layer serves all four surfaces. Vista lists MCP as an integration; ours would be native. |
| **Content Categories & evergreen** | Category-driven scheduling is a differentiator once the AI can drive it. |
| **Blog / long-form** | Outside Vista's scope entirely. |

**What we should copy without apology:** named agents, the hours-saved metric, live run status, describe→confirm→run creation, approval-by-default.

---

## 4. Where the chat lives

One chat service. Multiple entry points.

| Surface | Behaviour | Phase |
|---|---|---|
| **AI Studio (full page)** | The home of chat. Full history, agents, skills, long workflows. | 1 |
| **Composer side panel** | Chat scoped to the post being written. Rewrite, adapt per platform, suggest times. | 2 |
| **Global command bar (⌘K)** | Quick asks from anywhere. Opens into AI Studio if the answer is long. | 3 |
| **Analytics side panel** | Chat pre-scoped to the current report and date range. | 3 |
| **Mobile** | Read + light write. No destructive tools. | 4 |

The chat is a service with a session context, not a screen. Surface only changes (a) which tools are pre-scoped and (b) how results render.

---

## 5. Architecture decisions (the ones that matter)

### 5.1 We do NOT expose 159 endpoints as 159 tools

This is the single most important call in this document, and codebase verification strengthened it.

**99 of the 159 endpoints are analytics** (62%), and they are all the same shape: `workspace_id` + `platform_id` + `start_date` + `end_date` + `timezone`. They differ only by platform and by which metric comes back. Exposing them one-to-one would:

- Blow up the system prompt with near-identical tool descriptions
- Wreck tool-selection accuracy (models degrade badly past ~40–50 tools)
- Make every analytics question a guessing game between eight nearly identical options

**Instead: ~32 tools.** Analytics collapses from 99 endpoints into **6 parameterised tools** using `platform` and `metric` enums. The backend maps `(platform, metric)` to the correct endpoint. Full mapping in [Tool Catalog](https://app.helpin.ai/share/9a2e81a1a3ff56c267ca61d0c50e5217).

```
 99 analytics endpoints  →   6 tools
 28 inbox endpoints      →   9 tools   (API live, re-cut in [Tool Catalog](https://app.helpin.ai/share/9a2e81a1a3ff56c267ca61d0c50e5217) §8)
 32 other endpoints      →  ~17 tools  (14 already shipping)
                            ─────────
                            ~32 tools
```

**Supporting evidence found in the codebase:** there are already **86 per-metric analytics narrative agents** in `contentstudio-ai-agents/src/agents/analytics/`, one per platform per metric. That is the one-to-one explosion, already built, on the dashboard side. It works there because the dashboard selects the agent by UI context. A chat model choosing between 86 near-identical options would not work — which is exactly why the collapse to 6 parameterised tools is right.

### 5.2 Composite tools — the API is per-account, users think per-workspace

Every analytics endpoint takes **one** `platform_id`. But nobody asks "how did Instagram account 17841453340834745 do." They ask **"how did we do last week?"** — which across 12 connected accounts on 5 platforms is 12+ sequential API calls, each a separate LLM round trip. That is 30+ seconds and a wall of tokens for one simple question.

**Decision:** the tool layer needs **server-side aggregation tools** that fan out internally and return one shaped result.

- `analytics_workspace_overview` — one call, all accounts, all platforms, normalised KPIs
- `analytics_compare_accounts` — ranks accounts on a chosen metric

These are **not** raw API passthroughs. They are backend composites, and they are the difference between a demo and a product.

**v1.1 correction — cheaper than originally flagged.** `contentstudio-ai-agents/src/agents/analytics/overview/` already does cross-account aggregation across 6 agents, and the internal API already serves the overview dashboard from it. The work is **exposing and reconciling what exists**, not inventing cross-platform metric normalisation from scratch. What still needs product sign-off is confirming those existing mappings (`fan_count`/followers/subscribers/connections → `audience`) are the ones we standardise on, rather than defining a competing second set. [API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168) Gap 4, re-scoped to S–M.

### 5.3 Context bootstrap — stop the AI wasting turns on lookups

Users say *"post to our Instagram."* The API needs `17841453340834745`.

If the model has to call `accounts_list` before every action, every request costs an extra round trip. **Fix:** at session start, inject a compact context block (~500 tokens) into the system prompt:

- Workspace ID, name, timezone, first day of week
- Connected accounts: `id`, `name`, `platform`, `status` (active / expired)
- Labels, campaigns, content categories (id + name only)
- Current user's role and permissions
- Plan limits and remaining AI credits
- Today's date in workspace timezone

Refresh on workspace switch, or when a write tool changes the list. Everything else stays lazy.

### 5.4 Tool results are shaped for the model, not copied from the API

A Facebook summary response is 19 metrics × 4 blocks. Multiply by accounts and the context window is gone before the model has said a word.

**Every tool response passes through a shaping layer:**
- Round numbers, strip nulls and empty arrays
- Top-N by default (5 posts, not 100)
- Hard cap ~2,000 tokens per tool result; paginate past that
- Attach a **render directive** so the UI draws a component instead of printing text

```json
{
  "summary": "Instagram grew 4.2% (+320 followers) this week",
  "data": { "...": "compact" },
  "render": { "component": "kpi_grid", "props": { "...": "..." } },
  "next_cursor": null
}
```

**Render directives are how the chat stops looking like a terminal.** Post previews, KPI grids, charts, account pickers, calendars, confirmation cards. Vista does this well and it is most of why their chat feels like a product.

### 5.5 Permissions are enforced at the tool registry, not in the prompt

The API already has workspace roles and per-member permissions. **The tool list is filtered per user, per workspace, at session start.** A client-role user is never *offered* `posts_delete` or `team_invite`.

Never rely on the system prompt to say "don't do X for this user." Prompts get talked around. Registries do not.

### 5.6 Safety tiers

| Tier | Tools | Behaviour |
|---|---|---|
| **Green — auto** | All reads (list, get, analytics) | Runs immediately, no confirmation |
| **Amber — preview + confirm** | Create post, upload media, create campaign/label, add comment, approve/reject | Renders a preview card; user clicks Confirm |
| **Red — explicit confirm** | Delete post, remove team member, delete workspace/campaign/label, connect account | Typed or two-step confirmation; names the exact object |

**Batch operations always confirm**, regardless of tier. "Schedule 12 posts" gets one confirmation showing all 12, not twelve confirmations and not zero.

**Auto-approve mode** (Phase 4, per workspace, off by default, admin-only) lets trusted workflows skip Amber. Red never auto-approves.

### 5.7 Idempotency

Chat retries. Networks fail mid-call. Without protection, a retried `posts_create` publishes twice.

Every write tool sends a client-generated `Idempotency-Key`. The API returns the original result on replay. **This must exist before Phase 2 ships.**

### 5.8 Metering

Each tool call costs an LLM round trip plus an API request. Both need to count against the unified request metering already being built for the API plan. Agents make this urgent — an hourly agent across 50 client workspaces is 36,000 runs a month.

Per workspace: monthly AI action budget, visible remaining balance, soft warning at 80%, hard stop at 100% with an upgrade path. Agents get their own per-agent cap so one runaway agent cannot drain the workspace.

---

## 6. Phasing

| Phase | Scope | Blocked by | Outcome |
|---|---|---|---|
| **0 — Foundation** | Registry consolidation (see caveat in §1), envelope reconciliation, context bootstrap, permission filtering, tier→state-machine mapping, render components, idempotency, metering hooks | — | Nothing user-visible; everything depends on it |
| **1 — Read tools** | **All 6 analytics tools + 2 composites** (the bulk of the work), media list, platforms list, reviews list, context bootstrap. Accounts/posts/labels/campaigns/categories reads already ship | Phase 0 | "Ask anything about your workspace." Zero risk, immediate value, best possible way to test tool-selection accuracy |
| **2 — Write tools** | Create post, upload media, campaigns/labels/team/workspace CRUD, connections | Phase 1 + idempotency + **update-post and validate endpoints (both missing)** | The flagship: "create and schedule this for me" |
| **3 — Inbox tools** | Items triage, threads, reply (DM/comment/review), assign, tags, notes, comment moderation | ~~Inbox API~~ **unblocked — API is live** | Engagement workflows |
| **4 — Agents** | Agent builder, scheduler, run history, default agent library, notification channels | Phases 1–3 stable | Recurring value without user effort |
| **5 — Skills** | Skill runtime, `/` invocation, default skill library, custom skill builder, workspace sharing | Phase 4 | Packaged expertise, agency reuse |

**Recommendation on sequencing:** ship Phase 1 to a beta cohort on its own. Read-only chat is genuinely useful, carries no risk of destroying a customer's calendar, and is the cheapest possible way to find out whether the model picks the right tool. Every tool-selection bug found in Phase 1 is a bug that would have published something wrong in Phase 2.

**v1.1 — consider swapping Phases 2 and 3.** Phase 2's two hardest dependencies (update-post, post validation) do not exist and are M–L builds. Phase 3's API is live, and the remaining gaps are small (sentiment fields, a needs-reply filter, exposing bulk on existing routes). Inbox is now the shortest path from Phase 1 to new user-visible capability, and it is where Vista's agent story is strongest. Phase 2 stays the flagship, but it no longer has to be next.

**Also note:** Phase 5 (Skills) has a head start. `contentstudio-frontend/src/modules/AI-tools/` already ships `SavedPrompts.vue`, `SavedPromptsModal.vue` and `AddCustomPromptModal.vue`. Saved prompts are the nearest existing thing to Skills — build on that surface rather than beside it.

---

## 7. What the API cannot do yet

Full detail with priorities in **[API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168) — API Gap Register**. The headline blockers:

Priorities revised in v1.1 after codebase verification.

| # | Gap | Why it hurts | Priority |
|---|---|---|---|
| 1 | **No update-post endpoint** — only create, delete, approve | "Change the caption" / "move it to 5pm" is the second thing every user will say. Right now the only answer is delete-and-recreate, which loses the post ID, approval state and comments. **Verified still absent.** | **P0** |
| 3 | **No post validation / dry-run** | AI-generated posts will break platform rules (length, media counts, aspect ratios, carousel minimums). Without a validator we find out at publish time. **Verified still absent.** | **P0** |
| 4 | **No cross-account analytics rollup in the public API** | Section 5.2. Every "how are we doing" question is unusably slow. Now **expose-and-reconcile**, not build. | **P1** |
| 2a | **No sentiment / intent field on inbox items** | Every triage agent re-analyses raw text on every run — slow, costly, inconsistent between runs. | **P1** |
| 2b | **No server-side "needs reply" inbox filter** | Computing it client-side means fetching the whole inbox first. | **P1** |
| 5 | **Limited best-time-to-post** | `active-users` is Facebook and Instagram only. `publishing-behaviour` and `performance-schedule` are partial substitutes on 6 more platforms. The flagship "schedule at the best times" workflow needs a first-class answer. | **P1** |
| 6 | **Content Categories are read-only** | Chat can use categories but cannot create or edit them. **Verified: list is the only endpoint.** | **P1** |
| 7 | **No queue / slot read** | `publish_type: "queued"` is accepted but nothing exposes the queue or the next free slot. | **P1** |
| 8 | **No Discover / trending endpoint** | The flagship niche→trends→campaign workflow needs it. Our biggest differentiator is unreachable from the API. Exa-backed open-web research already works in chat; our own curated index does not. | **P1** |
| 2c | **Inbox bulk actions not exposed** | Downstream already accepts `element_ids[]`; the public route is single-item only. A route-signature change, not new plumbing. | **P2** |
| 10 | **No bulk post operations** | "Reschedule everything next week by a day" is N calls. | **P2** |
| 16 | **No ads analytics in the public API** ⭐ new | Meta Ads and Google Ads narrative agents exist internally; zero ads endpoints in v1. A paid-media surface most competitors lack. | **P2** |
| ~~2~~ | ~~No Inbox API~~ | **RESOLVED — 28 endpoints live.** | — |
| ~~9~~ | ~~No GMB review replies~~ | **RESOLVED — `POST`/`DELETE /inbox/reviews/{id}/reply` live.** Review Watcher is buildable end-to-end. | — |

### One naming trap engineering must not fall into

`/api/v1/workspaces/{id}/posts/{post_id}/comments` is **internal team collaboration comments** — notes on a draft, with `is_note` and `mentioned_users`. It is **not** comments from the public on a published social post.

Replying to real audience comments is an **Inbox** capability, at `/api/v1/workspaces/{id}/inbox/posts/{post_id}/comments`.

**This trap got sharper in v1.1, not safer.** Both endpoints are now live, and their paths differ by one segment while their meaning differs completely. Two clearly different tool names (`posts_internal_notes` vs `inbox_reply_to_comment`) are mandatory, or the model — and the user — will confuse them constantly. When a user says "reply to the comments on my post," they mean the inbox one, every time.

---

## 8. Success measures

| Metric | Target |
|---|---|
| Tool-selection accuracy (correct tool on first attempt) | > 92% |
| Task completion without user rephrasing | > 80% |
| Confirmation rejection rate (user says no to a preview) | < 10% — higher means the AI is guessing wrong |
| Median time-to-result, read queries | < 4s |
| Median time-to-result, publish workflow | < 20s |
| % of active workspaces using chat weekly | 25% by end of Phase 2 |
| % of chat users who create an agent | 30% within 30 days of Phase 4 |
| Failed / errored tool calls | < 2% |

Track **confirmation rejection rate** most closely. It is the only metric that tells us whether the AI understood the user, as opposed to whether it managed to call something.

---

## 9. Open decisions

0. **Which tool surfaces are in scope?** ⭐ *added in v1.1, and the biggest uncosted variable in this plan.* The same tool may need implementing in up to three places: the chat's Operations agent, the backend's hosted `/api/mcp` server (8 tools today), and the npm `contentstudio-mcp` / `contentstudio-cli` packages. Every per-tool estimate in [Tool Catalog](https://app.helpin.ai/share/9a2e81a1a3ff56c267ca61d0c50e5217) assumes one. Decide whether the hosted MCP server is deprecated in favour of the npm one, or kept in sync, before sizing anything.
0b. **Which response envelope?** The live `_envelope()` shape versus [Tool Catalog](https://app.helpin.ai/share/9a2e81a1a3ff56c267ca61d0c50e5217)'s proposed `{ok, summary, data, render, …}`. Must be settled in Phase 0 — every tool depends on it.
1. **Model choice and cost ceiling** per tool-calling turn — affects the metering model directly.
2. **Chat home:** dedicated AI Studio page first, or composer panel first? (Recommendation: AI Studio first, composer in Phase 2 — Vista's composer-first placement is good but only after the tool layer is trusted.)
3. **Agent autonomy default:** notify-only, or draft-and-wait? (Recommendation: draft-and-wait. Notify-only is too weak to justify setup.)
4. **Multi-workspace agents** for agencies — one agent across 20 clients, or one per workspace? Materially different data model. Decide before Phase 4 build starts.
5. **Skills marketplace** — user-shareable across workspaces/accounts, or private only? Affects moderation load.
6. **Do we support undo?** For post creation, a 60-second undo window is cheaper than a confirmation dialog and feels better. Worth prototyping.

---

## 10. Immediate next steps

Revised in v1.1 — two items are obsolete, three are new.

| # | Action | Owner |
|---|---|---|
| 1 | **Decide the surface question** (§9 decision 0) and the envelope question (0b) — both block Phase 0 | Product + Tech Lead |
| 2 | Review and sign off tool grouping ([Tool Catalog](https://app.helpin.ai/share/9a2e81a1a3ff56c267ca61d0c50e5217)), now with the "Exists today" column | Product + Tech Lead |
| 3 | Size the **two remaining P0 gaps** — update-post and validate ([API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168)) | Backend |
| 4 | ⭐ **Scope the analytics tool set** — 99 endpoints, 86 existing narrative agents, zero chat exposure. The largest single block of unreachable capability, and the direct answer to the "how did my profiles do" use cases | Product + Backend |
| 5 | ⭐ **Confirm the overview agents' metric normalisation** is the standard we expose, rather than defining a second one ([API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168) Gap 4) | Product + Backend |
| 6 | ⭐ **Re-plan Inbox as near-term, not blocked** — 28 live endpoints; remaining gaps are 2a/2b/2c and all small | Product + Backend |
| 7 | Design the render-component set: post preview, KPI grid, chart, confirmation card, account picker | Design (Fasih) |
| 8 | Audit Infra C and D (structured error codes, account token status) — both may be partly done | Backend |
| 9 | Write the Phase 1 epic from [Use Case Catalog](https://app.helpin.ai/share/03a7cf23c93af70ebde7a7410d4e6e5e) use cases | Product |
| ~~—~~ | ~~Confirm Inbox API delivery date~~ — shipped | — |
| ~~—~~ | ~~Build Phase 0 registry with 3 read tools as a spike~~ — 14 tools already live; measure selection accuracy on those instead | — |
