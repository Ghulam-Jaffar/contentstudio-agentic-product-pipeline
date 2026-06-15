# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Repo Is

A **Claude Code-powered product development pipeline** for [ContentStudio](https://contentstudio.io), a social media management platform. It automates the workflow from feature idea → research → PRD → ready-to-create stories, with review gates at every step. The pipeline produces local markdown deliverables; it does **not** push to Shortcut — a Product Owner creates the epics and stories in Shortcut manually from the markdown.

This is **not** a code project. There's no package.json or composer.json at root. The `contentstudio-backend/`, `contentstudio-frontend/`, `contentstudio-ios-v2/`, `contentstudio-android-v2/`, `social-inbox-manager/`, and other service directories are **gitignored separate repos** mounted here so the pipeline can analyze the actual codebase when writing stories.

## Two Pipeline Commands

### `/feature` — Full Feature Pipeline (4+1 steps)
For major features requiring PRDs and dedicated epics.

**Research → Workflow Design → PRD → Epic + Stories → [Optional] Implement FE**

- Runs parallel competitor research (WebSearch) + codebase analysis (Explore agent) in Step 1
- Produces research, workflow, and PRD as local markdown — **nothing is pushed to Shortcut**
- Authors a dedicated epic + stories as markdown for the PO to create in Shortcut manually; each story carries a **Shortcut fields** block
- Outputs saved to `docs/features/<slug>/` (01-research.md through 04-epic-and-stories.md, optionally 05-implementation.md)
- Optionally implements `[FE]` stories: branches from `develop` in `contentstudio-frontend/`, one descriptive commit per story, creates PR
- Review gate after every step — never proceed without explicit user approval

### `/story` — Quick Story Pipeline (2+1 steps)
For small improvements that don't need a full PRD. Max 4 stories; if 5+, redirect to `/feature`.

**Research → Stories → [Optional] Implement FE**

- Lean codebase research using direct Grep/Read (not Explore agents)
- Produces stories as local markdown for the PO to create in Shortcut manually — **nothing is pushed to Shortcut**
- Each story's **Shortcut fields** block references the current Miscellaneous quarterly epic (see `.claude/shortcut-config.json` → `miscellaneous_epics`)
- Outputs saved to `docs/stories/<slug>/` (01-research.md and 02-stories.md, optionally 03-implementation.md)
- Optionally implements `[FE]` stories: branches from `develop` in `contentstudio-frontend/`, one descriptive commit per story, creates PR

## Key Files

| File | Purpose |
|---|---|
| `.claude/shortcut-config.json` | Shortcut **field reference** (no API push): workflow states, groups, custom fields, projects, epics, competitors — used to fill each story's Shortcut fields block |
| `.claude/commands/feature.md` | `/feature` pipeline definition |
| `.claude/commands/story.md` | `/story` pipeline definition |
| `docs/story-guidelines.md` | **Mandatory** 20-section rulebook — read before writing any story |
| `docs/ui-components.md` | **Mandatory** catalog of available UI components — read before writing FE stories. Update when `@contentstudio/ui` changes. |
| `docs/PRD Feature Template.md` | 12-section PRD template used by `/feature` Step 3 |
| `docs/Shortcut story template.md` | Story body structure (Description, Workflow, AC, Mock-ups, Impact, Dependencies, Quality checklist) |

## Story Rules (Summary)

The full rules are in `docs/story-guidelines.md`. Key points:

- **Structure stories with the "New Feature Template"** sections (under no team) — the PO selects this template when creating the story in Shortcut so the sections + quality-checklist tasks are pre-populated (the pipeline doesn't push, so there's no `story_template_id` API payload)
- **Titles:** `[BE]` / `[FE]` / `[iOS]` / `[Android]` / `[Design]` prefix + action-oriented title
- **Workflow sections:** Written from user's POV, never developer POV
- **FE stories must include all UI copy:** modal titles, labels, tooltips, placeholders, validation errors, empty/error/loading states — written for non-technical users with concrete examples
- **No estimates, no labels** — devs handle these during sprint planning
- **Always note a Shortcut project** (Web App, Mobile, Chrome App, etc.) in the story's Shortcut fields block
- **Always note the custom fields:** `priority`, `product_area`, and `skill_set` — options in `.claude/shortcut-config.json`
- **No dark mode, no RTL** — ContentStudio doesn't support either
- **AI features are web-only** — no mobile AI stories
- **UI components:** FE stories must reference components from `docs/ui-components.md` by name. Prefer `@contentstudio/ui` components over legacy `Cst*`. Flag any component gaps explicitly.
- **Color theming:** Use `text-primary-cs-500`, `bg-primary-cs-50`, etc. (CSS variable-backed) — never hardcode colors like `text-blue-600`
- **Reference stories by full title**, never by number
- **Create separate iOS/Android stories** when mobile apps are impacted
- **No local pipeline file references in stories** — never put `docs/features/...` or `docs/stories/...` paths in story content (the PO will create these stories in Shortcut, where local paths don't resolve). Reference other stories by full title. Codebase paths (e.g., `contentstudio-frontend/src/...`) are fine.

## Shortcut Integration (no API push)

**The pipeline does not write to Shortcut.** It never creates epics, stories, docs, tasks, or iterations via the Shortcut API. Both pipelines produce local markdown that the Product Owner reviews and then creates in Shortcut manually.

`.claude/shortcut-config.json` is used only as a **field reference** — the canonical names and options for workflow states, groups, custom fields (`priority`, `product_area`, `skill_set`), projects, and the miscellaneous epic. Each story's markdown ends with a **Shortcut fields** block built from these so the PO has everything ready when creating the story by hand:

- **Template:** New Feature Template — the PO selects it when creating the story so the standard sections + the 5 quality-checklist tasks are pre-populated. The 5 tasks are:
  1. `Mobile responsiveness tested (frontend only, N/A for backend-only stories)`
  2. `Multilingual support verified (frontend + backend, translations available or fallback handled)`
  3. `UI theming supported (default + white-label, design library components are being used)`
  4. `White-label domains impact reviewed`
  5. `Cross-product impact assessed (web, mobile apps, Chrome extension)`
- **Story type, project, group, epic, priority, product area, skill set** — mapped from config
- **No estimate, no labels** — devs estimate during sprint planning and the team manages labels
- **Iteration:** the PO assigns the current/target sprint at creation time

Workspace (for reference when creating work manually): `contentstudio-team`. Stories are typically created in the `ready_for_dev` workflow state and epics in the `to_do` state.

## ContentStudio Product Context

The pipeline analyzes and writes stories for these codebases (mounted but gitignored):

- **`contentstudio-backend/`** — Laravel 10 API (PHP 8.3, MongoDB, Redis, Kafka). Has its own `CLAUDE.md` with project-specific rules.
- **`contentstudio-frontend/`** — Vue 3 SPA (Composition API, Vuex → Pinia). Has docs in `contentstudio-frontend/docs/`.
- **`contentstudio-ios-v2/`** — iOS app (Swift, Xcode, CocoaPods). Analyzed only when the feature/story involves mobile.
- **`contentstudio-android-v2/`** — Android app (Kotlin, Gradle). Analyzed only when the feature/story involves mobile.
- **`contentstudio-ai-agents/`** — Python 3.13 multi-agent platform (Agno framework, FastAPI, Dramatiq + Redis, Kafka, PostgreSQL). Handles AI content generation (captions, images, videos, analytics). Has its own `CLAUDE.md`. Analyzed only when the feature/story involves AI generation or the AI agent pipeline.
- **`contentstudio-social-analytics-go/`** — Go microservices analytics pipeline (Kafka, ClickHouse, MongoDB). 5-stage pipeline: Scheduler → Fetcher → Parser → Processor → Sink. Analyzed only when the feature/story involves social media analytics data processing.
- **`social-inbox-manager/`** — Python social inbox service (FastAPI, Kafka, MongoDB, Redis, Pusher). Orchestrates ingestion, sync, and management of social media inbox data across platforms (Facebook, Instagram, LinkedIn, YouTube, GMB). Per-platform workers and strategies, webhook handling with Kafka fan-out, real-time UI updates via Pusher. Analyzed only when the feature/story involves social inbox, conversations, messages, comments, or reviews.

When the pipeline does codebase analysis, it searches these directories for relevant models, controllers, services, components, routes, and composables to ground stories in the actual implementation. iOS/Android codebases are included **only when the feature description or request mentions mobile, iOS, or Android**. AI agents and analytics Go codebases are included only when the feature description explicitly involves AI generation or analytics data pipelines.

## Branch & PR Conventions (for code implementation)

When implementing stories in the sub-project codebases:

- **Branch from:** `develop`
- **Branch naming:** `feature/{story-title-slug}` (or a feature slug when multiple FE stories share one branch)
- **Commit format:** `{description}` — prefix `[sc-{id}] ` only if the PO has already created the story in Shortcut and supplied its ID (then Shortcut auto-links the commit/PR)
- **PR base:** `develop`
