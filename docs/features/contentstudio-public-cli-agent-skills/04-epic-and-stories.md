# Epic + Stories - ContentStudio Public CLI & Agent Skills

## Epic

**Title:** ContentStudio Public CLI & Agent Skills

**Description:**

Ship a public, npm-installed ContentStudio CLI that lets developers, operators, and AI agents authenticate with an API key, discover workspaces and accounts, upload media, and create or manage posts from the command line. V1 is built on the existing public REST API with structured JSON output, a bundled `SKILL.md` manifest, and a standalone public skill-install path for shell-capable agents.

The release also includes onboarding docs and launch assets so ContentStudio can be marketed as a safe, human-in-the-loop automation surface. The work covers public API hardening where needed, CLI foundations, auth/config flows, publishing commands, agent discovery packaging, public docs, and website launch support.

---

## Linked Docs

- Background strategy: https://app.shortcut.com/contentstudio-team/write/IkRvYyI6I3V1aWQgIjY5ZDc0MTFjLTU3ZDgtNDBmYy1hNTk5LWNmNTZiMjE0ZDRhYiI=
- Research: https://app.shortcut.com/contentstudio-team/write/IkRvYyI6I3V1aWQgIjY5ZDdhZTJjLWExMTgtNDdmNS1hMmE4LTFkNmRjNzVkNjI3NCI=
- PRD: https://app.shortcut.com/contentstudio-team/write/IkRvYyI6I3V1aWQgIjY5ZDdhZTQ5LTZkNmMtNDBhNS05NjFjLTAyYTljNjZjODBmYSI=

---

## Stories

---

### Story 1: [BE] Audit and harden the public publishing API for CLI reliability

**Description:**
As a developer using the ContentStudio CLI, I want the existing public publishing API to behave consistently across auth, discovery, media, posting, approvals, and comments so that terminal workflows and automation do not break on inconsistent payloads or unclear validation.

This story audits and hardens the existing public API routes in `contentstudio-backend/routes/api/v1.php` and the related controllers in `contentstudio-backend/app/Http/Controllers/Api/V1/`. The goal is not to create a new backend architecture. The goal is to close any contract gaps that would make the public CLI unreliable.

**Workflow:**
1. User runs `contentstudio auth login --api-key cs_xxx` and the CLI validates the key against `GET /api/v1/me`.
2. User runs `contentstudio workspaces list` and sees only workspaces they can access.
3. User runs `contentstudio accounts list --workspace <id>` and sees connected social accounts for that workspace.
4. User uploads media, creates posts, lists posts, approves or rejects posts, and adds comments from the CLI.
5. User gets consistent success and validation responses across all these flows, so the CLI can render readable output and structured JSON safely.

**Acceptance criteria:**
- [ ] Audit `GET /api/v1/me`, `GET /api/v1/workspaces`, `GET /api/v1/workspaces/{workspace_id}/accounts`, `GET|POST /api/v1/workspaces/{workspace_id}/media`, `GET|POST|DELETE /api/v1/workspaces/{workspace_id}/posts`, `POST /api/v1/workspaces/{workspace_id}/posts/{post_id}/approval`, and `GET|POST /api/v1/workspaces/{workspace_id}/posts/{post_id}/comments` for CLI readiness
- [ ] Success responses across these routes use a consistent top-level shape that the CLI can depend on without special-casing endpoint by endpoint
- [ ] Validation failures return clear field-level messages for missing workspace ids, invalid account ids, missing text content, invalid post state, invalid approval action, and invalid comment payloads
- [ ] `GET /api/v1/me` is sufficient for credential validation and returns the user identity details the CLI needs for `auth whoami`
- [ ] Media upload failures return clear validation messages for missing files, invalid file types, and rejected uploads
- [ ] Route-level permission failures return explicit authorization messages instead of generic server failures
- [ ] API request logging continues to work for these routes after any hardening changes
- [ ] No new auth model is introduced; all routes continue to use the existing `X-API-Key` contract

**Mock-ups:**
N/A - backend only

**Impact on existing data:**
- No new user-facing data model is required by this story
- Existing request-log behavior for API-key usage must remain intact

**Impact on other products:**
- Improves reliability for any existing or future tool using the public API
- No mobile app impact
- No Chrome extension impact

**Dependencies:**
- None - this story should land before the CLI command contract is considered stable

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness - N/A, backend-only story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support - N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 2: [BE] Build the ContentStudio CLI foundation and reusable client layer

**Description:**
As a developer building the ContentStudio CLI, I want a clean package foundation and one reusable TypeScript client layer so that all CLI commands share the same auth, request, timeout, and error-handling behavior.

This story sets up the CLI package boundary, command framework, reusable HTTP client layer, config loading, and shared output primitives. It is the foundation story for all command work that follows.

**Workflow:**
1. User installs the CLI with npm.
2. User runs `contentstudio --help` and sees a discoverable command tree.
3. Every CLI command uses the same client layer for base URL resolution, auth headers, retries, timeout handling, and response normalization.
4. Future commands plug into the same framework instead of implementing raw fetch logic independently.

**Acceptance criteria:**
- [ ] A dedicated CLI package is created for `@contentstudio/cli` with binary name `contentstudio`
- [ ] The package has a shared TypeScript client layer for base URL handling, auth header injection, retries, timeouts, and normalized errors
- [ ] The command framework supports subcommands for `auth`, `workspaces`, `accounts`, `media`, `posts`, and `comments`
- [ ] Human-readable output helpers and JSON output helpers are implemented centrally instead of per-command
- [ ] The package exposes a single executable entry point and supports both global npm install and `npx`
- [ ] The CLI can read configuration from local config plus environment variables without duplicating logic across commands
- [ ] The foundation does not require a separately published SDK package in v1
- [ ] The foundation includes basic automated tests for command bootstrapping and shared client behavior

**Mock-ups:**
N/A - backend/tooling story

**Impact on existing data:**
- No product database changes
- New local CLI config file will be created on user machines during auth flows

**Impact on other products:**
- Creates a new public developer surface for ContentStudio
- No mobile app impact
- No Chrome extension impact

**Dependencies:**
- Depends on: **[BE] Audit and harden the public publishing API for CLI reliability**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness - N/A, backend/tooling story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support - N/A, backend/tooling story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 3: [BE] Implement CLI authentication, local config persistence, and structured errors

**Description:**
As a CLI user, I want a simple login command, environment-variable auth, and predictable structured failures so that I can get into a working state quickly and debug auth problems without reading raw backend payloads.

This story covers `contentstudio auth login`, `contentstudio auth whoami`, `contentstudio auth logout`, config persistence, environment-variable override behavior, and the CLI's shared error contract.

**Workflow:**
1. User copies an API key from `Settings -> API Key` in the app.
2. User runs `contentstudio auth login --api-key cs_xxx`.
3. CLI validates the key with `GET /api/v1/me` and saves local config if valid.
4. User runs `contentstudio auth whoami` and sees their authenticated identity.
5. User can alternatively set `CONTENTSTUDIO_API_KEY` and skip local login.
6. If the key is invalid, revoked, or missing, the CLI returns a concise error in human mode and structured JSON in `--json` mode.

**Acceptance criteria:**
- [ ] `contentstudio auth login --api-key <key>` validates credentials through `GET /api/v1/me` before saving local config
- [ ] `contentstudio auth whoami` returns authenticated user info using the same public API contract
- [ ] `contentstudio auth logout` removes locally stored credentials cleanly
- [ ] `CONTENTSTUDIO_API_KEY` is supported as an auth source and overrides local config when present
- [ ] Local config stores base URL, API key, and optional default workspace id in a user-scoped config path
- [ ] Invalid, revoked, or missing credentials produce concise human-readable errors
- [ ] The same auth failures produce structured JSON errors when `--json` is used
- [ ] Non-zero exit codes are returned for auth and validation failures
- [ ] Auth command help text tells users to get keys from `Settings -> API Key`

**Mock-ups:**
N/A - backend/tooling story

**Impact on existing data:**
- No server-side data model changes
- Local user config is created and removed by CLI auth commands

**Impact on other products:**
- Leverages the existing in-app API key page without changing it
- No mobile app impact
- No Chrome extension impact

**Dependencies:**
- Depends on: **[BE] Build the ContentStudio CLI foundation and reusable client layer**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness - N/A, backend/tooling story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support - N/A, backend/tooling story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 4: [BE] Add publishing, media, approval, and comment commands to the CLI

**Description:**
As a technical operator, I want the ContentStudio CLI to cover the core publish lifecycle so that I can discover context, upload media, create posts, review approvals, and manage comments from one terminal workflow.

This story implements the first functional command set for the public CLI on top of the existing public API.

**Workflow:**
1. User runs `contentstudio workspaces list` to choose a workspace.
2. User runs `contentstudio accounts list --workspace <id>` to identify connected accounts.
3. User runs `contentstudio media upload --workspace <id> --file ./asset.png` when media is needed.
4. User runs `contentstudio posts create --workspace <id> --accounts <id1,id2> --text "..."` to create a draft or scheduled post.
5. User runs `contentstudio posts list --workspace <id>` to verify the result.
6. User runs `contentstudio posts approve`, `contentstudio comments list`, `contentstudio comments add`, or `contentstudio posts delete` as needed to complete the workflow.

**Acceptance criteria:**
- [ ] `workspaces list` returns accessible workspaces in human mode and `--json`
- [ ] `accounts list --workspace <id>` returns connected social accounts for the chosen workspace in human mode and `--json`
- [ ] `media list --workspace <id>` and `media upload --workspace <id> --file <path>` work against the existing media endpoints
- [ ] `posts list --workspace <id>` works in human mode and `--json`
- [ ] `posts create --workspace <id> --accounts <ids> --text "..."` supports the existing public post-creation contract and returns the created post data
- [ ] `posts delete --workspace <id> --post <id>` deletes a post through the existing public API route
- [ ] `posts approve --workspace <id> --post <id> --action approve|reject` works against the existing approval route and supports rejection comments where the API allows them
- [ ] `comments list --workspace <id> --post <id>` and `comments add --workspace <id> --post <id> --text "..."` work in human mode and `--json`
- [ ] All commands return parseable structured output in `--json` mode
- [ ] Validation and permission failures are surfaced without raw stack traces

**Mock-ups:**
N/A - backend/tooling story

**Impact on existing data:**
- No schema changes required by the CLI itself
- Commands operate on existing workspace, media, post, approval, and comment records

**Impact on other products:**
- Reuses existing public API functionality
- No mobile app impact
- No Chrome extension impact

**Dependencies:**
- Depends on: **[BE] Audit and harden the public publishing API for CLI reliability**
- Depends on: **[BE] Build the ContentStudio CLI foundation and reusable client layer**
- Depends on: **[BE] Implement CLI authentication, local config persistence, and structured errors**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness - N/A, backend/tooling story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support - N/A, backend/tooling story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 5: [BE] Package the CLI for AI-agent discovery with bundled SKILL.md and JSON-safe execution

**Description:**
As an AI-agent operator, I want the ContentStudio CLI package to declare its binary, required environment variable, and supported commands clearly so that shell-capable agents can discover and execute ContentStudio workflows without custom reverse engineering.

This story adds the bundled `SKILL.md` manifest, ensures JSON-safe execution is documented and tested, and frames the CLI as an agent-usable shell tool without changing the underlying command behavior.

**Workflow:**
1. Operator installs the ContentStudio CLI.
2. Operator sets `CONTENTSTUDIO_API_KEY`.
3. A shell-capable AI agent reads the bundled `SKILL.md` and sees the `contentstudio` binary plus the required environment variable.
4. The agent runs `contentstudio workspaces list --json`, `accounts list --json`, and publishing commands through the same public CLI used by human operators.
5. The operator keeps review authority over publish-triggering workflows by following the documented human-in-the-loop usage guidance.

**Acceptance criteria:**
- [ ] The published CLI package contains a bundled `SKILL.md`
- [ ] `SKILL.md` declares `contentstudio` as the binary and `CONTENTSTUDIO_API_KEY` as the required environment variable
- [ ] `SKILL.md` lists the core supported commands for discovery, media, publishing, approvals, and comments
- [ ] Every shipped command behaves correctly with `--json` so shell-capable agents can parse command output safely
- [ ] CLI examples for agent use avoid raw HTTP examples and use CLI commands directly
- [ ] Human-in-the-loop guidance is included in the packaged docs/examples for publish-triggering flows
- [ ] Package publish contents are reviewed to ensure the manifest ships with the public npm package

**Mock-ups:**
N/A - backend/tooling story

**Impact on existing data:**
- No server-side data changes
- No app-side data changes

**Impact on other products:**
- Creates a discoverable shell interface for AI-agent users
- No mobile app impact
- No Chrome extension impact

**Dependencies:**
- Depends on: **[BE] Build the ContentStudio CLI foundation and reusable client layer**
- Depends on: **[BE] Implement CLI authentication, local config persistence, and structured errors**
- Depends on: **[BE] Add publishing, media, approval, and comment commands to the CLI**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness - N/A, backend/tooling story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support - N/A, backend/tooling story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 6: [BE] Publish public CLI docs, quickstart examples, and agent setup guides

**Description:**
As a technical user evaluating the ContentStudio CLI, I want clear install, auth, quickstart, troubleshooting, and agent-setup docs so that I can get from zero to a working publish flow without trial-and-error.

This story produces the public documentation set that supports the CLI launch and keeps support load low.

**Workflow:**
1. User lands on the docs entry point for the CLI.
2. User sees installation instructions for npm and `npx`.
3. User sees where to get an API key in ContentStudio and how to run `auth login`.
4. User follows a first-success walkthrough: `workspaces list` -> `accounts list` -> optional `media upload` -> `posts create`.
5. User reads troubleshooting for invalid keys, missing workspace ids, and validation failures.
6. Agent users follow a separate setup guide for `CONTENTSTUDIO_API_KEY`, `--json`, and bundled `SKILL.md` usage.

**Acceptance criteria:**
- [ ] Public docs include installation, authentication, quickstart, command reference, troubleshooting, and FAQ sections
- [ ] Docs explicitly tell users to get API keys from `Settings -> API Key`
- [ ] Quickstart includes a full first-success flow using the CLI, not raw API requests
- [ ] Command reference covers all shipped v1 commands and flags
- [ ] Troubleshooting covers invalid API key, missing workspace id, missing accounts, upload failures, and validation failures
- [ ] Agent setup guide covers `CONTENTSTUDIO_API_KEY`, `--json`, and the bundled `SKILL.md`
- [ ] Docs copy avoids promising unsupported analytics, inbox, or broader automation features in v1
- [ ] Examples are copy-pasteable and aligned with shipped command names

**Mock-ups:**
N/A - documentation story

**Impact on existing data:**
- None - documentation deliverable only

**Impact on other products:**
- Reduces support burden for the CLI launch
- Supports website conversion and onboarding flows

**Dependencies:**
- Depends on: **[BE] Implement CLI authentication, local config persistence, and structured errors**
- Depends on: **[BE] Add publishing, media, approval, and comment commands to the CLI**
- Depends on: **[BE] Package the CLI for AI-agent discovery with bundled SKILL.md and JSON-safe execution**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness - N/A, documentation story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support - N/A, documentation story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 7: [FE] Launch the public CLI website page and conversion CTAs

**Description:**
As a website visitor interested in automating ContentStudio, I want a clear CLI landing page and strong conversion CTAs so that I can understand what the CLI does, how to get started, and why it is useful for developer and AI-agent workflows.

This story adds the website surface for the launch. Website uses its own CMS/component system rather than `@contentstudio/ui`, so implementation should follow the website stack and patterns already used by the marketing site.

**Workflow:**
1. User visits the CLI landing page from the website navigation, developer page, or launch CTA.
2. User sees a hero section that explains the CLI in plain language and shows the install command immediately.
3. User scrolls through capability cards for discovery, media upload, publishing, approvals/comments, and JSON-safe automation.
4. User sees a quickstart section that shows how to install the CLI, get an API key, list workspaces, and create a post.
5. User sees a human-in-the-loop note so the positioning stays responsible and concrete.
6. User clicks either the docs CTA or the app CTA to get their API key and start using the product.

**Acceptance criteria:**
- [ ] Website includes a dedicated CLI landing page with hero, capability overview, quickstart, FAQ, and CTA sections
- [ ] Main hero copy uses:
  - Eyebrow: "New for developers and automation teams"
  - Heading: "Run ContentStudio from the terminal"
  - Subtext: "Install the ContentStudio CLI with npm, authenticate with your API key, and manage publishing workflows from scripts, servers, and shell-capable AI agents."
  - Primary CTA: "View CLI Docs"
  - Secondary CTA: "Get Your API Key"
- [ ] Install section includes:
  - Heading: "Get started in minutes"
  - Step 1 label: "Install the CLI"
  - Step 2 label: "Connect your ContentStudio account"
  - Step 3 label: "Run your first command"
  - Install command example: `npm install -g @contentstudio/cli`
  - Auth command example: `contentstudio auth login --api-key cs_xxx`
  - First command example: `contentstudio workspaces list`
- [ ] Capability cards include the following titles and descriptions:
  - "Workspace discovery" - "List workspaces and connected social accounts before you publish."
  - "Media upload" - "Upload images and other assets from the terminal, then reuse them in posts."
  - "Publishing workflows" - "Create, review, approve, reject, and delete posts with explicit commands."
  - "Automation-safe JSON" - "Use `--json` for parseable output in scripts and agent workflows."
  - "Agent-ready packaging" - "Use the same CLI from shell-capable AI agents through the bundled SKILL.md manifest."
- [ ] Human-in-the-loop note uses:
  - Heading: "Built for automation, designed for operator review"
  - Body: "Use the CLI to speed up publishing workflows, but keep a human in the loop for live posting, approvals, and destructive actions."
- [ ] FAQ section includes:
  - "Where do I get my API key?" -> "Generate or copy it from Settings -> API Key in your ContentStudio account."
  - "Do I need to call the API directly?" -> "No. The CLI is the supported terminal interface for the v1 launch."
  - "Can AI agents use this?" -> "Yes. Shell-capable agents can run the same CLI commands and parse `--json` output."
  - "What is included in v1?" -> "Auth, workspace and account discovery, media upload, post creation and management, approvals, comments, and setup docs."
- [ ] Website CTAs route users to the docs surface and the in-app API key flow correctly
- [ ] Page is responsive across desktop and mobile website breakpoints

**Mock-ups:**
Design team to provide page mockups and any supporting launch assets if net-new website visuals are required

**Impact on existing data:**
- None - website content and navigation updates only

**Impact on other products:**
- Feeds traffic into the CLI docs and the app's existing API key page
- No mobile app impact
- No Chrome extension impact

**Dependencies:**
- Depends on: **[BE] Publish public CLI docs, quickstart examples, and agent setup guides**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 8: [Design] Create launch assets for the public CLI website and docs

**Description:**
As a designer supporting the CLI launch, I want a cohesive visual package for the website page and supporting docs assets so that ContentStudio's terminal story looks intentional, credible, and easy to understand.

This story covers the design deliverables needed for the website and documentation launch. It should stay grounded in ContentStudio's existing website visual language rather than inventing a separate product identity.

**Workflow:**
1. Designer reviews the Research doc, PRD, and website story.
2. Designer creates the CLI landing-page layout, responsive variants, quickstart code-block treatment, and capability-card visuals.
3. Designer provides any diagrams or supporting assets needed for docs and launch materials.
4. Final assets are linked to the website and docs implementation stories.

**Acceptance criteria:**
- [ ] Desktop and mobile mockups are created for the CLI landing page
- [ ] Hero section, quickstart section, capability cards, FAQ section, and CTA areas are designed
- [ ] Code-block styling for install/auth/first-command examples is defined
- [ ] Any required supporting visual assets for docs and launch materials are provided
- [ ] Designs align with ContentStudio's current website visual system and messaging tone
- [ ] Final design links are attached to the related implementation stories before development starts

**Mock-ups:**
This story is the mockup deliverable

**Impact on existing data:**
- None

**Impact on other products:**
- Supports website launch and public docs presentation
- No mobile app impact
- No Chrome extension impact

**Dependencies:**
- None - can start from the Research doc and PRD

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Story 9: [BE] Publish a standalone ContentStudio skill repo for direct agent installation

**Description:**
As an AI-agent user, I want to install ContentStudio into a supported agent registry with a one-line command so that I can add the skill quickly without manually extracting metadata from the CLI package.

This story creates a dedicated public skill repo under the ContentStudio GitHub organization for direct installation through agent-skill tooling. The CLI remains the executable product surface. The standalone skill repo is the distribution layer that points agents to the `contentstudio` binary and required environment variable.

**Workflow:**
1. User installs the public CLI with npm so the `contentstudio` binary is available.
2. User runs `npx skills add contentstudio/contentstudio-agent`.
3. The skills tool installs the standalone ContentStudio skill from the public repo.
4. The agent reads the installed `SKILL.md` and sees the `contentstudio` binary, required environment variable, and supported commands.
5. User sets `CONTENTSTUDIO_API_KEY`.
6. The agent uses the installed skill to run `contentstudio ... --json` commands through the public CLI.

**Acceptance criteria:**
- [ ] A dedicated public repo exists under the ContentStudio GitHub org for the standalone skill install path
- [ ] The repo is installable with `npx skills add contentstudio/contentstudio-agent`
- [ ] The repo contains a top-level `SKILL.md` that declares `contentstudio` as the binary and `CONTENTSTUDIO_API_KEY` as the required environment variable
- [ ] The standalone `SKILL.md` documents the same core command surface as the CLI launch: workspaces, accounts, media, posts, approvals, and comments
- [ ] The standalone skill repo README tells users to install `@contentstudio/cli` first so the binary is present
- [ ] Public docs and website copy mention the standalone skill-install path alongside the npm CLI install path
- [ ] A release/update process is documented so the standalone skill repo stays in sync with CLI command changes

**Mock-ups:**
N/A - backend/tooling story

**Impact on existing data:**
- No server-side data changes
- No app-side data changes

**Impact on other products:**
- Adds a second, lower-friction install path for agent ecosystems that prefer skill registries
- No mobile app impact
- No Chrome extension impact

**Dependencies:**
- Depends on: **[BE] Package the CLI for AI-agent discovery with bundled SKILL.md and JSON-safe execution**
- Depends on: **[BE] Publish public CLI docs, quickstart examples, and agent setup guides**

**Global quality & compliance (wherever applicable)**
- [ ] Mobile responsiveness - N/A, backend/tooling story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support - N/A, backend/tooling story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## Story Summary

| # | Title | Group | Project | Priority | Skill Set | Product Area | Story Type |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | [BE] Audit and harden the public publishing API for CLI reliability | Backend | Web App | High | Backend | Publishing | feature |
| 2 | [BE] Build the ContentStudio CLI foundation and reusable client layer | Backend | Web App | High | Backend | Throughout Product | feature |
| 3 | [BE] Implement CLI authentication, local config persistence, and structured errors | Backend | Web App | High | Backend | Settings | feature |
| 4 | [BE] Add publishing, media, approval, and comment commands to the CLI | Backend | Web App | High | Backend | Publishing | feature |
| 5 | [BE] Package the CLI for AI-agent discovery with bundled SKILL.md and JSON-safe execution | Backend | Web App | High | Backend | Throughout Product | feature |
| 6 | [BE] Publish public CLI docs, quickstart examples, and agent setup guides | Technical Writing | Website | Medium | Product | Throughout Product | chore |
| 7 | [FE] Launch the public CLI website page and conversion CTAs | Frontend | Website | High | Frontend | Throughout Product | feature |
| 8 | [Design] Create launch assets for the public CLI website and docs | Design | Website | Medium | Design | Throughout Product | chore |
| 9 | [BE] Publish a standalone ContentStudio skill repo for direct agent installation | Backend | Web App | High | Backend | Throughout Product | feature |
