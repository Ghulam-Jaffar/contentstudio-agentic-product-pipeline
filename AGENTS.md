# AGENTS.md

This file provides guidance to Codex when working with code in this repository. It mirrors `CLAUDE.md`; keep the two in sync, changing only the agent-specific paths.

## What This Repo Is

A **Codex-powered product development pipeline** for [ContentStudio](https://contentstudio.io), a social media management platform. It automates the workflow from feature idea → research → PRD → stories → epics and stories created in **Helpin**, the team's tracker, with review gates at every step. Deliverables are authored as local markdown first; once the Product Owner approves them, the pipeline pushes them to Helpin over the **Helpin MCP server** (`https://mcp.helpin.ai/mcp`, configured in `.mcp.json`).

This is **not** a code project. There's no package.json or composer.json at root. The `contentstudio-backend/`, `contentstudio-frontend/`, `contentstudio-flutter/`, `social-inbox-manager/`, and other service directories are **gitignored separate repos** mounted here so the pipeline can analyze the actual codebase when writing stories.

## Three Pipeline Commands

### `/feature` — Full Feature Pipeline (5+1 steps)
For major features requiring PRDs and dedicated epics.

**Research → Workflow Design → PRD → Epic + Stories → Push to Helpin → [Optional] Implement FE**

- Runs parallel competitor research (WebSearch) + codebase analysis (Explore agent) in Step 1
- Produces research, workflow, and PRD as local markdown, reviewed before anything is pushed
- Authors a dedicated epic + stories as markdown, then creates them in Helpin once approved
- Research, workflow and PRD go into Helpin as **Docs linked to the epic**; authored stories all stay stories
- Outputs saved to `docs/features/<slug>/` (01-research.md through 04-epic-and-stories.md, then 05-helpin-links.md, optionally 06-implementation.md)
- Optionally implements `[FE]` stories: branches from `develop` in `contentstudio-frontend/`, one descriptive commit per story, creates PR
- Review gate after every step — never proceed without explicit user approval

### `/story` — Quick Story Pipeline (3+1 steps)
For small improvements that don't need a full PRD. Max 4 stories; if 5+, redirect to `/feature`.

**Research → Stories → Push to Helpin → [Optional] Implement FE**

- Lean codebase research using direct Grep/Read (not Explore agents)
- Produces stories as local markdown, reviewed before anything is pushed, then creates them in Helpin
- Outputs saved to `docs/stories/<slug>/` (01-research.md, 02-stories.md, 03-helpin-links.md, optionally 04-implementation.md)
- Optionally implements `[FE]` stories: branches from `develop` in `contentstudio-frontend/`, one descriptive commit per story, creates PR

### `/frill` — Customer Feedback Intake (4 steps)
Front-end to the other two pipelines. Pulls feature requests off the public Frill board (https://contentstudio.frill.co), triages them, and hands a brief to `/feature` or `/story` in-session — no copy-paste.

**Sync → Triage → Brief → [/feature | /story]**

- **Near read-only against Frill *ideas*.** Never creates an idea, edits its body, deletes it or comments on it. The **one** permitted write is a **status change**, and only on the PO's explicit approval (see **Closing the loop on Frill requests**). Announcements are the other write, see **Changelog Publishing**
- All Frill I/O goes through `agents/scripts/frill-sync.sh` (retry/backoff, throttling, ordering assertions, atomic snapshot writes)
- Triage clusters duplicates (5 WhatsApp ideas, 3 Dark Mode), dedupes against the ~238 existing dirs in `docs/features/` and `docs/stories/`, and rejects out-of-scope asks (dark mode, RTL, blog publishing)
- **Always enriches before routing** — median idea body is ~200 characters, so a raw idea is a seed, not a brief. Requirements get mined from the idea's comment thread and its duplicate cluster
- Full provenance (votes, cluster, comment mining) goes in `01-research.md` only. The **Frill idea link itself does travel with the story to Helpin** — it is what lets a shipped ticket find the customers who asked for it
- Every processed idea gets a `docs/frill/ledger.json` entry (including rejections) so it never resurfaces
- When a Frill idea is routed onward, the resulting Helpin story carries the **Frill idea link** so the loop can be closed later

## Key Files

| File | Purpose |
|---|---|
| `agents/commands/feature.md` | `/feature` pipeline definition (Codex skill: `.codex/skills/feature/SKILL.md` symlinks here) |
| `agents/commands/story.md` | `/story` pipeline definition (Codex skill: `.codex/skills/story/SKILL.md` symlinks here) |
| `agents/commands/frill.md` | `/frill` customer-feedback intake pipeline definition (Codex skill: `.codex/skills/frill/SKILL.md` symlinks here) |
| `agents/scripts/frill-sync.sh` | Frill API sync/triage script — all Frill *read* I/O goes through it |
| `.mcp.json` | Helpin MCP server registration for Claude Code. Codex does not read it; register the same URL in `~/.codex/config.toml` |
| `docs/stories/HELPIN-PUSH-PLAN.md` | Helpin push execution spec + the open backlog batch |
| `docs/frill/ledger.json` | Committed provenance ledger: Frill idea → decision + deliverable |
| `docs/story-guidelines.md` | **Mandatory** 20-section rulebook — read before writing any story |
| `docs/ui-components.md` | **Mandatory** catalog of available UI components — read before writing FE stories. Update when `@contentstudio/ui` changes. |
| `docs/PRD Feature Template.md` | 12-section PRD template used by `/feature` Step 3 |
| `docs/story-template.md` | Story body structure (Description, Workflow, AC, Mock-ups, Impact, Dependencies, Quality checklist) |
| `videos/README.md` | Where generated videos live, how they are built and rendered, naming rules |

## Story Rules (Summary)

The full rules are in `docs/story-guidelines.md`. Key points:

- **Structure stories with the standard sections** from `docs/story-template.md` — Description, Workflow, Acceptance criteria, Mock-ups, Impact on existing data, Impact on other products, Dependencies, Global quality checklist
- **Titles:** `[BE]` / `[FE]` / `[Flutter]` / `[Design]` prefix + action-oriented title (`[Flutter]` is the only mobile prefix — no more `[iOS]` / `[Android]`)
- **Workflow sections:** Written from user's POV, never developer POV
- **FE stories must include all UI copy:** modal titles, labels, tooltips, placeholders, validation errors, empty/error/loading states — written for non-technical users with concrete examples
- **No estimates, and no labels written into the story body** — estimates are set by devs during sprint planning; labels, priority, state, sprint and assignee are supplied by the PO and set as **Helpin fields at push time**
- **No trailing metadata block** — a story ends at the global quality checklist. No project/group/epic/priority fields block of any kind
- **No dark mode, no RTL** — ContentStudio doesn't support either
- **AI generation features are web-only** — no mobile stories for AI image/video/caption generation. **Exception:** AI chat/assistant exists in the Flutter app (`lib/features/ai_assistant/`) and is in scope for mobile.
- **UI components:** FE stories must reference components from `docs/ui-components.md` by name. Prefer `@contentstudio/ui` components over legacy `Cst*`. Flag any component gaps explicitly.
- **Color theming:** Use `text-primary-cs-500`, `bg-primary-cs-50`, etc. (CSS variable-backed) — never hardcode colors like `text-blue-600`
- **Reference stories by full title**, never by number
- **Create one `[Flutter]` story** when the mobile app is impacted — a single cross-platform story, never a separate iOS one and Android one. Ground it in `contentstudio-flutter/`
- **No local pipeline file references in stories** — never put `docs/features/...` or `docs/stories/...` paths in story content (the PO will recreate these stories in the tracker, where local paths don't resolve). Reference other stories by full title. Codebase paths (e.g., `contentstudio-frontend/src/...`) are fine.

## Deliverables: Markdown First, Then Helpin

**Every deliverable is authored as local markdown and reviewed there.** Nothing is created in Helpin until the PO approves the markdown. The push is the final pipeline step, not a side effect of authoring.

All tracker I/O goes through the **Helpin MCP server** (`https://mcp.helpin.ai/mcp`, OAuth as the signed-in user). The repo's `.mcp.json` registers it for Claude Code only; Codex needs the same server added to its own MCP config (`~/.codex/config.toml`) before any push. Granted scopes: `context.read`, `pm.read`, `pm.write`, `docs.read`, `docs.write`, `agents.read`, `agents.run`. There are no API tokens or field-ID config files in this repo — the MCP connection carries the credential, and it is stored outside the repo.

Stories still carry **no metadata block in the body** — no template ID, story type, project, group, epic, priority, product area, skill set, estimate, labels, or iteration. A story ends at the global quality checklist. That rule survives the integration for a better reason than before: **that metadata now goes in as Helpin fields at push time**, supplied by the PO, so duplicating it as text in the body would just be a second copy that drifts.

### Rules for every push

- **Only newly authored work is pushed.** The existing backlog under `docs/features/` and `docs/stories/` is **not** being backfilled into Helpin — those folders stay as local markdown records. Never offer to push an old folder or treat one as pending work. If the PO wants a specific old folder pushed, they will say so explicitly.
- **Every push needs the PO's explicit approval in the moment.** Approving the markdown approves *the authoring* — it is not consent to create anything in Helpin. Show exactly what will be created (titles, epic, space, state, sprint, assignee, priority, labels) and wait for a clear yes. A yes covers the batch in front of them, never the next one.
- **Never invent an identifier.** Space, state, sprint, assignee, label, priority and epic references are resolved from the live Helpin workspace or the PO. If one can't be resolved, stop and ask — do not guess and do not create a placeholder.
- **The PO specifies** the sprint (current or next), workflow state, labels, assignee and priority. Ask; don't assume a default.
- **Story bodies go in as authored.** Do not re-summarize to fit a field.
- **Leave every checklist box unchecked.** They are for devs during implementation.
- **Write the returned URLs back** to `<slug>/0N-helpin-links.md` so the local docs and Helpin stay cross-referenced.
- **On partial failure, record what was created and stop.** Never blindly retry a half-finished batch — that is how duplicate epics happen.
- **Docs are filed by document type, never left uncategorized** — `01-research.md` to the Research collection, workflow and PRD to PRDs & Feature Specs, competitor research to Competitor Analysis. Product-area collections are curated by humans; the pipeline does not file into them.
- **Nothing in Helpin is ever deleted — archive is the only removal, and the pipeline cannot even do that.** The MCP surface has no delete or archive tool, and the Helpin UI offers only Archive. So a mistaken push is not undoable, just archivable, and cleanup is always a manual UI job for the PO. Never promise to delete or archive something from Helpin — hand over the link. This is the real weight behind the approval gate: there is no clean way back.
- **The 6-item quality checklist stays in the story body as content** — written as markdown in the description, never created as native Helpin `checklist_items`. PO decision, 2026-09-23: the checklist is part of what the story *says*, not a set of sub-tasks to track. Helpin strips `- [ ]` syntax, so it renders as a plain bullet list; that is accepted and expected, so do not "fix" it by converting to checklist items.

Execution spec: `docs/stories/HELPIN-PUSH-PLAN.md`.

The one thing the story body does carry is the standard 6-item global quality checklist, left unchecked for devs:

1. `Mobile responsiveness (frontend only, N/A for backend-only stories)`
2. `Multilingual support (frontend + backend, translations available or fallback handled)`
3. `UI theming support (default + white-label, design library components are being used)`
4. `White-label domains impact review`
5. `Cross-product impact assessment (web, mobile apps, Chrome extension)`
6. `Developer surfaces coverage (any new or changed API is reflected in the public API, CLI, MCP server and automation apps, N/A when nothing API-facing changes)`

**Any API a story creates must reach the public developer surfaces** — public REST API, `contentstudio` CLI, MCP server, and the automation apps (Zapier, Make, n8n) that sit on top of them. The checklist item is the prompt; when a surface needs real work, write a story for it. See `docs/story-guidelines.md` section 8.

## Changelog Publishing (Helpin → Frill)

Shipped work gets announced back to the customers who asked for it. The flow: fetch the release ticket from Helpin → extract the main points → create the changelog on Frill.

- **Always created as a draft** (`published_at: null`). The PO publishes from Frill's admin UI. Nothing reaches the public board without a human.
- Uses `POST /v1/announcements`, linking source ideas via `idea_idxs` so the Frill → Helpin → Frill loop closes.
- This is one of only two Frill writes. The other is an idea **status change**, below. Never create an idea, edit its body, delete it or comment on it.
- `agents/scripts/frill-sync.sh` is read-only by construction. The announcement write is a separate, explicit path — do not loosen `api_get` to accommodate it.
- Frill's published API docs have already proven unreliable for this board. Probe the live contract before relying on any documented field. Two confirmed traps: `author_idx` is **required** on announcement create despite the docs, and unknown `sortBy` values are silently ignored.
- All Frill writing goes through `agents/scripts/frill-announce.sh`, which hardcodes `published_at: null`, has no publish flag, and refuses a body containing an em dash.

## Closing the loop on Frill requests

**Whenever a Frill-originated request ships, say so without being asked.** Surfacing it is not optional and does not wait for a prompt: check `docs/frill/ledger.json` against what shipped and bring every completed request to the PO.

- **Always surface, never auto-move.** Report that the request is complete and wait. The PO decides.
- **On approval, move the idea to `Shipped 🚀`** with `agents/scripts/frill-announce.sh status <idea_idx> <status_idx>`.
- **This emails everyone who voted.** That is the point, and it is why it never happens without approval.
- Record it in the ledger's `frill_status` block, alongside `shipped` and `announced`.
- A status change is the **only** edit permitted on an idea. Never touch its body, never comment, never delete.

## ContentStudio Product Context

The pipeline analyzes and writes stories for these codebases (mounted but gitignored):

- **`contentstudio-backend/`** — Laravel 10 API (PHP 8.3, MongoDB, Redis, Kafka). Has its own `AGENTS.md` with project-specific rules.
- **`contentstudio-frontend/`** — Vue 3 SPA (Composition API, Vuex → Pinia). Has docs in `contentstudio-frontend/docs/`.
- **`contentstudio-flutter/`** — **the mobile app, and the single source of truth for all mobile work** (Flutter/Dart, one codebase for iOS + Android). Has its own `CLAUDE.md`, `AGENTS.md`, `WORKFLOW.md`, `REWRITE_PLAN.md`, and feature docs in `docs/features/`. Code lives in `lib/` split into `app/`, `core/`, `shared/`, and `features/<feature>/` (composer, planner, inbox, ai_assistant, approval_workflows, billing, entitlements, media_library, social_channels, workspaces, auth, settings, notifications, and more). Analyzed only when the feature/story involves mobile.
- **`contentstudio-ai-agents/`** — Python 3.13 multi-agent platform (Agno framework, FastAPI, Dramatiq + Redis, Kafka, PostgreSQL). Handles AI content generation (captions, images, videos, analytics). Has its own `CLAUDE.md`. Analyzed only when the feature/story involves AI generation or the AI agent pipeline.
- **`contentstudio-social-analytics-go/`** — Go microservices analytics pipeline (Kafka, ClickHouse, MongoDB). 5-stage pipeline: Scheduler → Fetcher → Parser → Processor → Sink. Analyzed only when the feature/story involves social media analytics data processing.
- **`social-inbox-manager/`** — Python social inbox service (FastAPI, Kafka, MongoDB, Redis, Pusher). Orchestrates ingestion, sync, and management of social media inbox data across platforms (Facebook, Instagram, LinkedIn, YouTube, GMB). Per-platform workers and strategies, webhook handling with Kafka fan-out, real-time UI updates via Pusher. Analyzed only when the feature/story involves social inbox, conversations, messages, comments, or reviews.

When the pipeline does codebase analysis, it searches these directories for relevant models, controllers, services, components, routes, and composables to ground stories in the actual implementation. `contentstudio-flutter/` is included **only when the feature description or request mentions mobile, the app, iOS, or Android**. AI agents and analytics Go codebases are included only when the feature description explicitly involves AI generation or analytics data pipelines.

### Mobile: Flutter only

The native **iOS** and **Android** apps have been replaced by the single Flutter app. Their repos are **retired and no longer mounted here** — `contentstudio-ios-v2/` and `contentstudio-android-v2/` were deleted on 2026-08-17. All mobile research, epics, and stories are grounded in `contentstudio-flutter/` and use the `[Flutter]` prefix.

If a deliberate parity check is ever needed on native behavior that hasn't been ported yet, the archives are still on GitHub (`d4interactive/contentstudio-ios-v2`, `d4interactive/contentstudio-android-v2`) — ask the user before cloning either back, and even then the resulting story still describes Flutter work.

Older deliverables under `docs/features/` and `docs/stories/` still contain `[iOS]` / `[Android]` stories. Those are historical records of work already created — leave them as-is, and don't use them as a pattern for new stories.

## Generated Videos

**Every video the pipeline generates goes in `videos/<slug>/`, never under `docs/`.** One folder per video keeps every video in one place, whatever feature it belongs to.

- **Folder:** `videos/<feature-or-story-slug>-<kind>/` (kind is `promo`, `demo`, `teaser` or `explainer`), holding `README.md`, `source.html` and `<slug>-<ratio>.mp4` (for example `-16x9`, `-9x16`)
- **Docs keep only a pointer:** add `promo-video.md` (or `demo-video.md`) to the feature or story folder, linking to the video folder. No MP4, HTML or render script inside `docs/`
- **Shared tooling:** `videos/_tooling/render.js` renders any `source.html` to MP4 with headless Chrome and ffmpeg. Don't copy a render script into a video folder
- **Add each new video to the index** in `videos/README.md`
- **Only claim what ships.** Check the feature's research and PRD before putting a capability on screen. No em dashes in on-screen copy

Full contract (the deterministic `window.render(t)` page, render commands, aspect ratios): `videos/README.md`.

## Branch & PR Conventions (for code implementation)

When implementing stories in the sub-project codebases:

- **Branch from:** `develop` (in `contentstudio-flutter/` the integration branch is **`develop-cs`**, and branches are named `feat/<task>` — see its `WORKFLOW.md`)
- **Branch naming:** `feature/{story-title-slug}` (or a feature slug when multiple FE stories share one branch)
- **Commit format:** `{description}` — a plain descriptive message. Stories don't exist in a tracker at implementation time, so there is no ticket ID to prefix
- **PR base:** `develop` (`develop-cs` for `contentstudio-flutter/`) — never push directly to `develop-cs` or `main`
