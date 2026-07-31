# ContentStudio AI Chat — Agents & Skills (v1.1)

**Phases 4 and 5.** Both sit on top of the tool layer. Neither can start before Phase 2 is stable in production.

> **v1.1:** corrected against the codebase (2026-07-28). Review Watcher upgrades from notify to draft, all four inbox agents are unblocked, and Skills should build on the saved-prompts surface that already ships. See [As-Built-Baseline doc](https://app.helpin.ai/share/fff3d91449a0c4f64f84ac5eadc98f9d).

---

## 1. The distinction, stated plainly

| | **Chat** | **Skill** | **Agent** |
|---|---|---|---|
| Trigger | User types a request | User types `/name` | A schedule, or an event |
| Steps | AI decides on the fly | Fixed, pre-defined sequence | Fixed sequence |
| When | Now | Now | While you're asleep |
| Setup | None | Pick from library or build once | Configure once |
| Output | Conversation | A finished deliverable | A notification or deliverable |
| Best for | Anything unpredictable | Repeatable work you do often | Work you shouldn't have to remember |

**One-liner for marketing:** *Chat is asking. Skills are shortcuts. Agents are staff.*

**Engineering reality:** an Agent is a Skill on a schedule with a notification channel. Build the Skill runtime first and Agents become mostly scheduling and delivery infrastructure. This is the cheaper build order — but Agents are the stronger product story and the better retention driver, so **ship Agents first (Phase 4) using a simplified internal step-runner, then generalise that runner into user-facing Skills (Phase 5).**

---

# Part A — Agents (Phase 4)

## 2. Anatomy

```yaml
agent:
  id: string
  name: "Theo"                        # user-editable
  avatar: preset | generated
  role: "Morning briefer"             # one-line job title
  description: "Daily summary of your inbox, scheduled posts, and yesterday's performance."

  trigger:
    type: schedule | event
    cron: "0 8 * * 1-5"
    timezone: workspace_timezone       # NOT UTC
    event: post_failed | token_expired | sentiment_spike | approval_pending

  scope:
    workspace_id: string
    account_ids: [] | all

  instructions: "free text — what the user described"

  tools_allowed: [analytics_workspace_overview, posts_list, inbox_list_conversations]

  autonomy: notify_only | draft | auto_execute

  output:
    channels: [in_app, email, slack]
    format: digest | report | task_list
    recipients: [user_ids]

  limits:
    max_runs_per_month: int
    max_tool_calls_per_run: int
    credit_cap: int

  state:
    status: active | paused | error
    last_run: timestamp
    next_run: timestamp
    runs_completed: int
    hours_saved_estimate: float
```

### Autonomy levels — the most important field

| Level | Behaviour | Default for |
|---|---|---|
| `notify_only` | Reports findings. Takes no action. | Watchers, monitors |
| `draft` | Prepares work and waits for approval. | Content, replies, reports |
| `auto_execute` | Acts without asking. | Nothing, by default |

**`auto_execute` is off by default, admin-only to enable, and can never be granted to 🔴 tools.** An agent must never be able to delete a post, remove a team member or delete a workspace. Enforce in the registry — an agent's tool allowlist is intersected with a hard-coded non-destructive set before it ever runs.

Vista's answer to this question is *"Only if you let it."* Ours should be the same and should be said just as prominently in the UI.

## 3. Creation flow — describe → confirm → run

Copy Vista's three steps. It is the right pattern.

**Step 1 — Describe**
User types in plain language: *"Every morning at 8, tell me what happened yesterday."*
Also reachable from chat: *"Turn this into an agent"* after any useful conversation — the strongest possible conversion moment, because the value has already been proven.

**Step 2 — Confirm**
The AI parses that into a structured agent and shows it back for review:
- Name and role (AI-suggested, editable)
- Schedule in plain words: "Every weekday at 8:00 AM (Asia/Karachi)"
- Which tools it will use, in plain words: "Read your calendar, inbox and analytics"
- What it will and won't do: "It will send you a summary. It won't publish or reply to anything."
- Where results go
- Estimated cost per month in credits

Everything editable. Then **Confirm and proceed**.

**Step 3 — It runs**
Live status on the card. Run history. Pause, edit, duplicate, delete.

## 4. Default agent library

Pre-built, one-click deploy. Two enabled by default on Phase 4 launch (Morning Brief, Publishing Health) so the feature demonstrates itself.

| Agent | Job | Schedule | Tools | Autonomy | Phase |
|---|---|---|---|---|---|
| **Morning Brief** | Yesterday's performance + today's schedule + new messages + account health | Daily 8am | `analytics_workspace_overview`, `posts_list`, `inbox_list_conversations`, `accounts_list` | notify | 4 |
| **Publishing Health Monitor** ⭐ | Failed posts, expiring tokens, empty calendar days | Every 6h | `posts_list`, `accounts_list` | notify | 4 |
| **Weekly Performance Reporter** | Week in review, changes called out first | Mondays 9am | analytics suite | notify | 4 |
| **Approval Chaser** | Nudges pending approvals before they go stale | Daily 4pm | `posts_list` | notify | 4 |
| **Content Gap Watcher** | Flags days with nothing scheduled; offers drafts | Fridays | `posts_list`, `posts_create` (draft) | draft | 4 |
| **Client Report Drafter** | Monthly branded client report | Monthly, 1st | analytics suite | draft | 4 |
| **Inbox Triage** | Scores and ranks unread messages | Hourly | `inbox_*` | notify | 5 |
| **Reply-Worthy Finder** | Surfaces the comments actually worth a human reply | Every 4h | `inbox_*` | notify | 5 |
| **Sentiment Spike Watcher** | Alerts on unusual negative sentiment | Hourly | `inbox_*`, `analytics_summary` | notify | 5 |
| **Review Watcher** | New Google Business reviews **+ drafts replies** | Hourly | `reviews_list`, `inbox_reply` | draft | 5 |
| **Trend Scout** | Trending topics in your niche | Daily | Exa web research ✅ · own Discover index ⚠️ | notify | 5+ |

> **v1.1 corrections.**
> - **Review Watcher upgraded from notify to draft.** v1.0 assumed we could not reply to reviews; `POST /inbox/reviews/{review_id}/reply` is live, so this agent can draft responses for approval instead of just alerting. That closes the gap with Vista's review agent rather than shipping a weaker version.
> - **All four `inbox_*` agents are unblocked.** The Inbox API shipped (28 endpoints). They could move from Phase 5 to Phase 4 — or earlier — since Phase 2's write gaps are the slower path. **Sentiment Spike Watcher is the exception:** inbox items carry no sentiment field yet ([API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168) Gap 2a), so it still needs either that field or its own per-run analysis.
> - **Trend Scout is partly buildable today** via the existing Exa research tools. What is missing is ContentStudio's own Discover index — still the strategic gap ([API Gap Register](https://app.helpin.ai/share/fcecca058116a1870cb931c33d11e168) Gap 8).

⭐ **Publishing Health Monitor is our differentiator.** Vista has no equivalent. Expired tokens and silent post failures are among the top support-ticket drivers in this category — an agent that catches them before the customer notices is defensible, measurable value and an easy support-cost story for leadership.

## 5. Value display — copy this from Vista

Every agent card shows **"Saves ~X hrs/month."**

Estimate from: (manual minutes for the task) × (runs per month) ÷ 60. Publish the assumption in a tooltip so it stays honest. Roll up to a workspace-level **"Your agents saved you 47 hours this month"** — the single most renewal-relevant number in the product.

Also show live status ("Going through inbox…", "Next run in 2 hrs"). Visible automation is trusted automation.

## 6. Engineering requirements

| # | Requirement | Why |
|---|---|---|
| 1 | **Scheduler** — cron per workspace timezone, DST-aware, at-least-once with idempotency | An agent that posts twice on the DST boundary is a real incident |
| 2 | **Run isolation** | One agent failing must not stall the queue |
| 3 | **Run history** — inputs, tool calls, outputs, cost, duration, retained 90 days | Debugging and trust. Users will ask "what did it actually do?" |
| 4 | **Hard cost caps** per agent and per workspace | An hourly agent × 50 client workspaces = 36,000 runs/month |
| 5 | **Auto-pause on repeated failure** (3 consecutive) with a notification | Silent broken agents are worse than no agents |
| 6 | **Notification channels**: in-app feed, email, Slack | Slack is table stakes for agencies |
| 7 | **Tool allowlist intersected with a non-destructive set** | Hard guarantee, not a prompt instruction |
| 8 | **Agents run as a service identity with the creator's permissions**, revalidated per run | Creator leaves the workspace → agent must stop, not inherit orphaned access |

## 7. Agency consideration — decide before build

Multi-workspace agents: one agent covering 20 client workspaces, or 20 separate agents?

**Recommendation: one agent, multi-workspace scope, per-workspace output.** Agencies will not configure the same agent twenty times, and if we make them, they will conclude the feature isn't for them. This changes the data model materially — `scope.workspace_ids[]` rather than a single ID — so it must be decided before the schema is written, not after.

---

# Part B — Skills (Phase 5)

## 8. What a skill actually is

Vista's framing is precise and worth borrowing: *a skill is not a prompt.* It is a **multi-step expert workflow** — 4 to 8 steps of audit, analyse, compare, draft, format — that returns a finished deliverable.

The user-facing promise: **one command replaces thirty minutes.**

```yaml
skill:
  id: string
  command: "/client-report"
  name: "Client Report"
  description: "Full month report — performance, top posts, audience, recommendations."
  step_count: 6                       # advertised, builds confidence
  category: reporting | content | engagement | analysis
  inputs:
    - {name: workspace, type: workspace_ref, required: true}
    - {name: period, type: date_range, default: last_month}
  steps:
    - {n: 1, action: "Pull workspace summary", tool: analytics_workspace_overview}
    - {n: 2, action: "Rank accounts", tool: analytics_compare_accounts}
    - {n: 3, action: "Fetch top posts per account", tool: analytics_top_posts}
    - {n: 4, action: "Pull audience breakdown", tool: analytics_breakdown}
    - {n: 5, action: "Analyse patterns", tool: null}
    - {n: 6, action: "Format branded report", tool: null}
  output:
    format: report
    template: client_report_v1
    exportable: [pdf, link]
  visibility: system | workspace | private
  version: int
```

## 9. Invocation

`/` in the chat input opens a searchable palette. Fuzzy match on command and description. Show step count and a one-line description on each row — that is what makes it feel like depth rather than a macro.

Skills should also be callable **inside** a conversation: *"run /post-postmortem on yesterday's reel"* — the AI fills the inputs from context rather than prompting for them.

> **v1.1 — Skills have a head start in the frontend.** `contentstudio-frontend/src/modules/AI-tools/` already ships `SavedPrompts.vue`, `SavedPromptsModal.vue`, `AddCustomPromptModal.vue` and `PromptTemplate.vue`. Saved prompts are the closest existing surface to Skills: user-created, named, reusable, already discoverable in the chat UI.
>
> **Build Skills as an evolution of saved prompts, not a parallel feature.** Shipping both leaves users with two overlapping "saved thing I can run" concepts in the same input box, which is exactly the confusion §12 warns about for Skills vs Agents. The migration story also matters — existing saved prompts are single-step, so decide whether they become one-step Skills or stay a separate tier. Worth a Design review before Phase 5 scoping.
>
> Also check `src/orchestration/post_plan/` and the **Post Plan Builder** agent for overlap with any multi-post campaign skill before specifying a new one.

## 10. Default skill library

| Command | Job | Steps | Phase |
|---|---|---|---|
| `/client-report` | Full branded client report | 6 | 5 |
| `/weekly-review` | Week in review + next week's recommendations | 5 | 5 |
| `/post-postmortem` | Why one post over- or under-performed | 6 | 5 |
| `/content-plan` | A week or month of planned content | 7 | 5 |
| `/best-times` | Optimal posting windows per platform ⚠️ gap #5 | 4 | 5 |
| `/repurpose` | Turn a top performer into content for every other platform | 5 | 5 |
| `/hook-maker` | 8 opening-line options from your best-performing hooks | 4 | 5 |
| `/caption-remix` | 5 variants of a caption in different tones | 3 | 5 |
| `/audit-accounts` | Health check — tokens, gaps, cadence, dead accounts | 5 | 5 |
| `/triage-inbox` | Score, categorise and priority-rank the inbox | 4 | 5 |
| `/respond-to-review` | Brand-voice review reply, flagged if legal-sensitive ⚠️ gap #9 | 4 | 5+ |
| `/trend-ideas` | Trending topics → 10 ready concepts ⚠️ gap #8 | 5 | 5+ |

Seven of these have direct Vista equivalents. That is deliberate — parity on the obvious ones, then differentiate on `/audit-accounts`, `/content-plan` and anything Discover-powered.

## 11. Custom skills

**Build flow:** run something useful in chat → *"Save this as a skill"* → AI extracts the steps → user names it, edits steps, marks which inputs are variable → save.

This is the same conversion moment as agent creation and should use the same UI pattern. Users are far more likely to save a workflow they have just watched succeed than to build one from an empty form.

**Sharing:** private → workspace → organisation. Agency reuse across every client is the killer use case (Vista says it explicitly: *"reuse it across every client"*).

**Deferred:** a public marketplace. Sharing across organisations means moderation, quality control and support liability. Revisit after Phase 5 ships.

## 12. Skills vs Agents — keep the boundary clean

Users will confuse these. The product must not.

- A **skill** is a verb the user performs. It appears in the `/` palette.
- An **agent** is a noun that performs verbs. It appears in the agents list.
- **An agent can run a skill.** That is the correct composition, and it makes the "Weekly Reporter agent runs the /weekly-review skill" story clean.
- A skill cannot contain an agent.

Build the step-runner once. Agents and skills both call it. Do not build two.

---

## 13. Rollout

| Step | What | Gate |
|---|---|---|
| 4.1 | Agent runtime, scheduler, run history, cost caps | Phase 2 stable, error rate < 2% |
| 4.2 | 6 default notify-only agents, in-app + email delivery | 4.1 |
| 4.3 | Agent builder (describe → confirm → run), Slack delivery | 4.2 |
| 4.4 | `draft` autonomy, hours-saved rollup | 4.3 |
| 5.1 | Skill runtime (generalised step-runner), `/` palette | Phase 4 stable |
| 5.2 | 10 default skills | 5.1 |
| 5.3 | Custom skill builder, workspace sharing | 5.2 |
| 5.4 | Agents can invoke skills | 5.3 |

**Ship 4.2 with two agents already switched on.** A feature that requires configuration before it demonstrates value gets configured by very few people.
