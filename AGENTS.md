# AGENTS.md

This file provides guidance to Codex (Codex.ai/code) when working with code in this repository.

## What This Repo Is

A **Codex-powered product development pipeline** for [ContentStudio](https://contentstudio.io), a social media management platform. It automates the workflow from feature idea → research → PRD → ready-to-create stories, with review gates at every step. The pipeline produces local markdown deliverables and nothing else — it never calls a project-tracker API. A Product Owner takes the approved markdown and creates the epics and stories in the tracker manually.

This is **not** a code project. There's no package.json or composer.json at root. The `contentstudio-backend/`, `contentstudio-frontend/`, `contentstudio-flutter/`, `social-inbox-manager/`, and other service directories are **gitignored separate repos** mounted here so the pipeline can analyze the actual codebase when writing stories.

## Two Pipeline Commands

### `/feature` — Full Feature Pipeline (4+1 steps)
For major features requiring PRDs and dedicated epics.

**Research → Workflow Design → PRD → Epic + Stories → [Optional] Implement FE**

- Runs parallel competitor research (WebSearch) + codebase analysis (Explore agent) in Step 1
- Produces research, workflow, and PRD as local markdown — **nothing is pushed anywhere**
- Authors a dedicated epic + stories as markdown for the PO to create in the tracker manually
- Outputs saved to `docs/features/<slug>/` (01-research.md through 04-epic-and-stories.md, optionally 05-implementation.md)
- Optionally implements `[FE]` stories: branches from `develop` in `contentstudio-frontend/`, one descriptive commit per story, creates PR
- Review gate after every step — never proceed without explicit user approval

### `/story` — Quick Story Pipeline (2+1 steps)
For small improvements that don't need a full PRD. Max 4 stories; if 5+, redirect to `/feature`.

**Research → Stories → [Optional] Implement FE**

- Lean codebase research using direct Grep/Read (not Explore agents)
- Produces stories as local markdown for the PO to create in the tracker manually — **nothing is pushed anywhere**
- Outputs saved to `docs/stories/<slug>/` (01-research.md and 02-stories.md, optionally 03-implementation.md)
- Optionally implements `[FE]` stories: branches from `develop` in `contentstudio-frontend/`, one descriptive commit per story, creates PR

## Key Files

| File | Purpose |
|---|---|
| `.Codex/commands/feature.md` | `/feature` pipeline definition |
| `.Codex/commands/story.md` | `/story` pipeline definition |
| `docs/story-guidelines.md` | **Mandatory** 20-section rulebook — read before writing any story |
| `docs/ui-components.md` | **Mandatory** catalog of available UI components — read before writing FE stories. Update when `@contentstudio/ui` changes. |
| `docs/PRD Feature Template.md` | 12-section PRD template used by `/feature` Step 3 |
| `docs/story-template.md` | Story body structure (Description, Workflow, AC, Mock-ups, Impact, Dependencies, Quality checklist) |

## Story Rules (Summary)

The full rules are in `docs/story-guidelines.md`. Key points:

- **Structure stories with the standard sections** from `docs/story-template.md` — Description, Workflow, Acceptance criteria, Mock-ups, Impact on existing data, Impact on other products, Dependencies, Global quality checklist
- **Titles:** `[BE]` / `[FE]` / `[Flutter]` / `[Design]` prefix + action-oriented title (`[Flutter]` is the only mobile prefix — no more `[iOS]` / `[Android]`)
- **Workflow sections:** Written from user's POV, never developer POV
- **FE stories must include all UI copy:** modal titles, labels, tooltips, placeholders, validation errors, empty/error/loading states — written for non-technical users with concrete examples
- **No estimates, no labels** — devs handle these during sprint planning
- **No trailing metadata block** — a story ends at the global quality checklist. No project/group/epic/priority fields block of any kind
- **No dark mode, no RTL** — ContentStudio doesn't support either
- **AI generation features are web-only** — no mobile stories for AI image/video/caption generation. **Exception:** AI chat/assistant exists in the Flutter app (`lib/features/ai_assistant/`) and is in scope for mobile.
- **UI components:** FE stories must reference components from `docs/ui-components.md` by name. Prefer `@contentstudio/ui` components over legacy `Cst*`. Flag any component gaps explicitly.
- **Color theming:** Use `text-primary-cs-500`, `bg-primary-cs-50`, etc. (CSS variable-backed) — never hardcode colors like `text-blue-600`
- **Reference stories by full title**, never by number
- **Create one `[Flutter]` story** when the mobile app is impacted — a single cross-platform story, never a separate iOS one and Android one. Ground it in `contentstudio-flutter/`
- **No local pipeline file references in stories** — never put `docs/features/...` or `docs/stories/...` paths in story content (the PO will recreate these stories in the tracker, where local paths don't resolve). Reference other stories by full title. Codebase paths (e.g., `contentstudio-frontend/src/...`) are fine.

## Deliverables Are Local Markdown Only

**The pipeline has no project-tracker integration.** It never creates epics, stories, docs, tasks, or iterations through any API, and it holds no tracker credentials, tokens, or field-ID config. Both pipelines produce local markdown that the Product Owner reviews and then recreates in whatever tracker the team uses.

Because nothing is pushed, stories carry **no metadata block** — no template ID, story type, project, group, epic, priority, product area, skill set, estimate, labels, or iteration. A story ends at the global quality checklist. The PO sets all of that when creating the work by hand.

The one thing the story body does carry is the standard 5-item global quality checklist, left unchecked for devs:

1. `Mobile responsiveness (frontend only, N/A for backend-only stories)`
2. `Multilingual support (frontend + backend, translations available or fallback handled)`
3. `UI theming support (default + white-label, design library components are being used)`
4. `White-label domains impact review`
5. `Cross-product impact assessment (web, mobile apps, Chrome extension)`

## ContentStudio Product Context

The pipeline analyzes and writes stories for these codebases (mounted but gitignored):

- **`contentstudio-backend/`** — Laravel 10 API (PHP 8.3, MongoDB, Redis, Kafka). Has its own `AGENTS.md` with project-specific rules.
- **`contentstudio-frontend/`** — Vue 3 SPA (Composition API, Vuex → Pinia). Has docs in `contentstudio-frontend/docs/`.
- **`contentstudio-flutter/`** — **the mobile app, and the single source of truth for all mobile work** (Flutter/Dart, one codebase for iOS + Android). Has its own `AGENTS.md`, `AGENTS.md`, `WORKFLOW.md`, `REWRITE_PLAN.md`, and feature docs in `docs/features/`. Code lives in `lib/` split into `app/`, `core/`, `shared/`, and `features/<feature>/` (composer, planner, inbox, ai_assistant, approval_workflows, billing, entitlements, media_library, social_channels, workspaces, auth, settings, notifications, and more). Analyzed only when the feature/story involves mobile.
- **`contentstudio-ai-agents/`** — Python 3.13 multi-agent platform (Agno framework, FastAPI, Dramatiq + Redis, Kafka, PostgreSQL). Handles AI content generation (captions, images, videos, analytics). Has its own `AGENTS.md`. Analyzed only when the feature/story involves AI generation or the AI agent pipeline.
- **`contentstudio-social-analytics-go/`** — Go microservices analytics pipeline (Kafka, ClickHouse, MongoDB). 5-stage pipeline: Scheduler → Fetcher → Parser → Processor → Sink. Analyzed only when the feature/story involves social media analytics data processing.
- **`social-inbox-manager/`** — Python social inbox service (FastAPI, Kafka, MongoDB, Redis, Pusher). Orchestrates ingestion, sync, and management of social media inbox data across platforms (Facebook, Instagram, LinkedIn, YouTube, GMB). Per-platform workers and strategies, webhook handling with Kafka fan-out, real-time UI updates via Pusher. Analyzed only when the feature/story involves social inbox, conversations, messages, comments, or reviews.

When the pipeline does codebase analysis, it searches these directories for relevant models, controllers, services, components, routes, and composables to ground stories in the actual implementation. `contentstudio-flutter/` is included **only when the feature description or request mentions mobile, the app, iOS, or Android**. AI agents and analytics Go codebases are included only when the feature description explicitly involves AI generation or analytics data pipelines.

### Mobile: Flutter only

The native **iOS** and **Android** apps have been replaced by the single Flutter app. Their repos are **retired and no longer mounted here** — `contentstudio-ios-v2/` and `contentstudio-android-v2/` were deleted on 2026-08-17. All mobile research, epics, and stories are grounded in `contentstudio-flutter/` and use the `[Flutter]` prefix.

If a deliberate parity check is ever needed on native behavior that hasn't been ported yet, the archives are still on GitHub (`d4interactive/contentstudio-ios-v2`, `d4interactive/contentstudio-android-v2`) — ask the user before cloning either back, and even then the resulting story still describes Flutter work.

Older deliverables under `docs/features/` and `docs/stories/` still contain `[iOS]` / `[Android]` stories. Those are historical records of work already created — leave them as-is, and don't use them as a pattern for new stories.

## Branch & PR Conventions (for code implementation)

When implementing stories in the sub-project codebases:

- **Branch from:** `develop` (in `contentstudio-flutter/` the integration branch is **`develop-cs`**, and branches are named `feat/<task>` — see its `WORKFLOW.md`)
- **Branch naming:** `feature/{story-title-slug}` (or a feature slug when multiple FE stories share one branch)
- **Commit format:** `{description}` — a plain descriptive message. Stories don't exist in a tracker at implementation time, so there is no ticket ID to prefix
- **PR base:** `develop` (`develop-cs` for `contentstudio-flutter/`) — never push directly to `develop-cs` or `main`
