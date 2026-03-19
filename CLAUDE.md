# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Repo Is

A **Claude Code-powered product development pipeline** for [ContentStudio](https://contentstudio.io), a social media management platform. It automates the workflow from feature idea → research → PRD → Shortcut stories, with review gates at every step.

This is **not** a code project. There's no package.json or composer.json at root. The `contentstudio-backend/`, `contentstudio-frontend/`, `contentstudio-ios-v2/`, and `contentstudio-android-v2/` directories are **gitignored separate repos** mounted here so the pipeline can analyze the actual codebase when writing stories.

## Two Pipeline Commands

### `/feature` — Full Feature Pipeline (5+1 steps)
For major features requiring PRDs and dedicated epics.

**Research → Workflow Design → PRD → Epic + Stories → Push to Shortcut → [Optional] Implement FE**

- Runs parallel competitor research (WebSearch) + codebase analysis (Explore agent) in Step 1
- Creates **Shortcut Docs** for both research (Step 1) and PRD (Step 3), then links them to the epic in Step 5
- Creates a dedicated Shortcut epic for the feature
- Outputs saved to `docs/features/<slug>/` (01-research.md through 05-shortcut-links.md, optionally 06-implementation.md)
- After Shortcut push, optionally implements `[FE]` stories: branches from `develop` in `contentstudio-frontend/`, one commit per story with `[sc-{id}]`, creates PR
- Review gate after every step — never proceed without explicit user approval

### `/story` — Quick Story Pipeline (3+1 steps)
For small improvements that don't need a full PRD. Max 4 stories; if 5+, redirect to `/feature`.

**Research → Stories → Push to Shortcut → [Optional] Implement FE**

- Lean codebase research using direct Grep/Read (not Explore agents)
- Stories link to the miscellaneous quarterly epic (currently Q1 2026, id: 107163)
- Outputs saved to `docs/stories/<slug>/` (01-research.md through 03-shortcut-links.md, optionally 04-implementation.md)
- After Shortcut push, optionally implements `[FE]` stories: branches from `develop` in `contentstudio-frontend/`, one commit per story with `[sc-{id}]`, creates PR

## Key Files

| File | Purpose |
|---|---|
| `.claude/shortcut-config.json` | All Shortcut API IDs: workflows, states, groups, custom fields, projects, epics, competitors |
| `.claude/commands/feature.md` | `/feature` pipeline definition |
| `.claude/commands/story.md` | `/story` pipeline definition |
| `docs/story-guidelines.md` | **Mandatory** 15-section rulebook — read before writing any story |
| `docs/ui-components.md` | **Mandatory** catalog of available UI components — read before writing FE stories. Update when `@contentstudio/ui` changes. |
| `docs/PRD Feature Template.md` | 12-section PRD template used by `/feature` Step 3 |
| `docs/Shortcut story template.md` | Story body structure (Description, Workflow, AC, Mock-ups, Impact, Dependencies, Quality checklist) |

## Story Rules (Summary)

The full rules are in `docs/story-guidelines.md`. Key points:

- **Always use the "New Feature Template"** (under no team) when creating stories — pass `story_template_id` in the API payload
- **Titles:** `[BE]` / `[FE]` / `[iOS]` / `[Android]` / `[Design]` prefix + action-oriented title
- **Workflow sections:** Written from user's POV, never developer POV
- **FE stories must include all UI copy:** modal titles, labels, tooltips, placeholders, validation errors, empty/error/loading states — written for non-technical users with concrete examples
- **No estimates, no labels** — devs handle these during sprint planning
- **Always assign a Shortcut project** (Web App, Mobile, Chrome App, etc.)
- **Always assign custom fields:** `priority`, `product_area`, and `skill_set` — IDs in `.claude/shortcut-config.json`
- **No dark mode, no RTL** — ContentStudio doesn't support either
- **AI features are web-only** — no mobile AI stories
- **UI components:** FE stories must reference components from `docs/ui-components.md` by name. Prefer `@contentstudio/ui` components over legacy `Cst*`. Flag any component gaps explicitly.
- **Color theming:** Use `text-primary-cs-500`, `bg-primary-cs-50`, etc. (CSS variable-backed) — never hardcode colors like `text-blue-600`
- **Reference stories by full title**, never by number
- **Create separate iOS/Android stories** when mobile apps are impacted

## Shortcut API Integration

- API base: `https://api.app.shortcut.com/api/v3`
- Auth token: `SHORTCUT_API_TOKEN` env var (loaded from `.env`)
- Workspace: `contentstudio-team`
- Stories default to `ready_for_dev` workflow state (id: 500000070)
- Epics default to `to_do` state (id: 500000002)

### Windows curl requirement
Write JSON payloads to a project-local temp file (e.g., `D:/code/office/contentstudio-agentic-product-pipeline/tmp-payload.json`), then pass with `curl --data @filename`. Save API responses to a file with `-o`, then parse with `node` (not `python`). Clean up temp files after. Never use `/tmp/` or `/dev/stdin` — they don't work on Windows.

### Checklist tasks (critical gotcha)
`story_template_id` does **not** auto-create checklist tasks via the API. After creating each story, manually POST all 5 tasks to `POST /stories/{story_id}/tasks`:
1. `Mobile responsiveness tested (frontend only, N/A for backend-only stories)`
2. `Multilingual support verified (frontend + backend, translations available or fallback handled)`
3. `UI theming supported (default + white-label, design library components are being used)`
4. `White-label domains impact reviewed`
5. `Cross-product impact assessed (web, mobile apps, Chrome extension)`

### Shortcut Docs (feature pipeline only)
Create docs via `POST /documents` with `{"title": "...", "content": "...", "content_format": "markdown"}`. Link a doc to an epic via `PUT /documents/{doc_id}/epics/{epic_id}` — returns 204 No Content on success.

### Iteration confirmation
Before pushing stories, fetch active iterations via `GET /iterations`, find the next upcoming one (status: `unstarted`, nearest `start_date`), and confirm with the user which to use before setting `iteration_id`.

## ContentStudio Product Context

The pipeline analyzes and writes stories for these codebases (mounted but gitignored):

- **`contentstudio-backend/`** — Laravel 10 API (PHP 8.3, MongoDB, Redis, Kafka). Has its own `CLAUDE.md` with project-specific rules.
- **`contentstudio-frontend/`** — Vue 3 SPA (Composition API, Vuex → Pinia). Has docs in `contentstudio-frontend/docs/`.
- **`contentstudio-ios-v2/`** — iOS app (Swift, Xcode, CocoaPods). Analyzed only when the feature/story involves mobile.
- **`contentstudio-android-v2/`** — Android app (Kotlin, Gradle). Analyzed only when the feature/story involves mobile.
- **`contentstudio-ai-agents/`** — Python 3.13 multi-agent platform (Agno framework, FastAPI, Dramatiq + Redis, Kafka, PostgreSQL). Handles AI content generation (captions, images, videos, analytics). Has its own `CLAUDE.md`. Analyzed only when the feature/story involves AI generation or the AI agent pipeline.
- **`contentstudio-social-analytics-go/`** — Go microservices analytics pipeline (Kafka, ClickHouse, MongoDB). 5-stage pipeline: Scheduler → Fetcher → Parser → Processor → Sink. Analyzed only when the feature/story involves social media analytics data processing.

When the pipeline does codebase analysis, it searches these directories for relevant models, controllers, services, components, routes, and composables to ground stories in the actual implementation. iOS/Android codebases are included **only when the feature description or request mentions mobile, iOS, or Android**. AI agents and analytics Go codebases are included only when the feature description explicitly involves AI generation or analytics data pipelines.

## Branch & PR Conventions (for code implementation)

When implementing stories in the sub-project codebases:

- **Branch from:** `develop`
- **Branch naming:** `feature/sc-{story-id}/{story-title-slug}`
- **Commit format:** `[sc-{id}] {description}`
- **PR base:** `develop`
- Shortcut auto-links PRs via the `sc-XXXXX` in the branch name
