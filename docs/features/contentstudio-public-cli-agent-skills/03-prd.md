# PRD: ContentStudio Public CLI & Agent Skills

**Author:** Codex  
**Last Updated:** April 10, 2026  
**Status:** Draft  
**Target Release:** Q2 2026

---

## 1. Overview

ContentStudio Public CLI & Agent Skills packages ContentStudio's existing public publishing API into a public npm-installed CLI that developers, operators, and shell-capable AI agents can use for deterministic social-media workflows. V1 covers API-key authentication, workspace/account discovery, media upload, post creation and management, approvals, comments, JSON-safe output, a standalone public skill-install path, in-app setup entry points on the API Key and dashboard surfaces, and launch assets that position ContentStudio as a terminal-friendly automation surface.

---

## 2. Problem Statement

**What problem are we solving?**

ContentStudio already has public publishing API routes, API key management, and request logs, but it does not offer a public terminal product. That means developers and automation users must build directly against raw HTTP endpoints, and shell-capable AI agents have no clean, supported way to operate ContentStudio through a stable CLI contract.

**Who has this problem?**

- developers building internal publishing workflows
- technical marketing operators using scripts and cron jobs
- agencies or ops teams that want automation without a full custom integration
- AI-agent users who prefer a shell-executable tool over direct API orchestration

**What happens if we don't solve it?**

- ContentStudio remains harder to adopt for terminal-first and automation-first users
- the market narrative shifts to newer entrants that package API access into a clearer public developer surface
- internal teams and customers keep building one-off wrappers around the API, increasing support fragmentation

---

## 3. Goals & Success Metrics

| Goal | Metric | Target | How We'll Measure |
| --- | --- | --- | --- |
| Ship a credible public terminal surface | CLI MVP launched with auth, discovery, media, posting, approval, and comments commands | 100% of v1 command scope released in Q2 2026 | Release checklist and shipped package audit |
| Reduce time-to-first-success | New user completes install -> auth -> first successful data call in under 2 minutes | <= 2 minutes median in guided internal testing | QA runbook and pilot user walkthroughs |
| Make the CLI automation-safe | Every shipped command supports `--json` and structured non-zero failures | 100% command coverage | CLI contract test suite |
| Support agent adoption without custom glue | Bundled `SKILL.md` and docs are sufficient for shell-capable agent setup | 1 complete internal setup flow validated | Internal dogfooding with agent users |
| Avoid support regressions | No material increase in auth or API-contract confusion after launch | < 10% increase in related support tickets in first 30 days | Support ticket tagging and review |

---

## 4. Target Users

**Primary Persona:**  
Automation Developer / Technical Operator - Comfortable with terminals, scripts, API keys, and automation tools. Wants speed, predictability, and copy-pasteable commands more than UI guidance.

**Secondary Persona (if applicable):**  
AI-Agent Operator - Uses shell-capable agents to complete real tasks. Needs a binary with deterministic commands, stable JSON output, and explicit authentication/setup guidance.

**Non-Users (explicitly out of scope):**  
Non-technical social managers who only want a browser UI, mobile app users, and users who expect fully autonomous AI behavior with no operator review.

---

## 5. User Stories / Jobs to Be Done

| ID | As a... | I want to... | So that... | Priority |
| --- | --- | --- | --- | --- |
| US-1 | developer | install a ContentStudio CLI with npm and authenticate using my API key | I can start working without building a custom wrapper first | Must Have |
| US-2 | technical operator | list workspaces and connected social accounts from the terminal | I can discover the right publishing context before posting | Must Have |
| US-3 | automation user | upload media and create posts from the command line | I can script deterministic publishing workflows | Must Have |
| US-4 | reviewer or operator | approve, reject, comment on, or delete posts from the terminal | I can manage the full publishing lifecycle without leaving the shell | Must Have |
| US-5 | AI-agent operator | install ContentStudio into my agent environment with a one-line skill command and run the CLI with `--json` | my shell-capable agent can use ContentStudio safely with minimal setup | Must Have |
| US-6 | website visitor | understand what the CLI is and how to get started in minutes | I can evaluate the product quickly and convert | Should Have |
| US-7 | full-suite or API-plan user | discover the CLI from my dashboard and finish setup from the API Key page | I can go from in-app discovery to working terminal commands without searching external docs first | Must Have |

---

## 6. Requirements

### 6.1 Must Have (P0)

- Public npm package `@contentstudio/cli` with binary name `contentstudio`
- Reusable TypeScript client layer inside the CLI codebase
- API key authentication through `contentstudio auth login --api-key <key>`
- Environment-variable auth through `CONTENTSTUDIO_API_KEY`
- Credential validation through `GET /api/v1/me`
- Commands for:
  - `auth login`
  - `auth whoami`
  - `auth logout`
  - `workspaces list`
  - `accounts list --workspace <id>`
  - `media list --workspace <id>`
  - `media upload --workspace <id> --file <path>`
  - `posts list --workspace <id>`
  - `posts create --workspace <id> --accounts <ids> --text "..."`
  - `posts delete --workspace <id> --post <id>`
  - `posts approve --workspace <id> --post <id> --action approve|reject`
  - `comments list --workspace <id> --post <id>`
  - `comments add --workspace <id> --post <id> --text "..."`
- Human-readable output by default
- Valid structured JSON output with `--json` on every shipped command
- Structured non-zero failures in `--json` mode
- Public docs for install, auth, quickstart, command reference, troubleshooting, and agent setup
- Bundled `SKILL.md` manifest declaring `contentstudio` and `CONTENTSTUDIO_API_KEY`
- Standalone public skill repo installable via `npx skills add contentstudio/contentstudio-agent`
- Website launch page and developer-facing CTAs
- Web-app `ContentStudio CLI & Agent Access` quickstart section on `Settings -> API Key`
- Web-app `ContentStudio CLI` discovery card in the shared dashboard integrations carousel used by both the standard dashboard and API-centric dashboard
- Public API audit/gap-closure before CLI contract freeze

### 6.2 Should Have (P1)

- Default workspace support in local config
- Friendly table formatting in human mode
- Examples for multi-account posting and media upload
- Explicit guidance that API keys are created in `Settings -> API Key`
- Website copy that distinguishes CLI users from full-suite browser users cleanly
- Dashboard and API Key page copy that distinguishes the first-party CLI from third-party automation integrations clearly

### 6.3 Nice to Have (P2)

- Interactive prompts when required flags are missing
- shell completion
- extracted `@contentstudio/sdk` package if reuse justifies it
- additional prebuilt examples for cron jobs and agent workflows

### 6.4 Explicitly Out of Scope

- analytics commands
- inbox, CRM, or broadcast command groups
- mobile CLI clients
- browser/device auth
- a separately marketed standalone SDK in v1
- fully autonomous AI workflows without operator review

---

## 7. User Flow (High Level)

1. User opens the CLI landing page or discovers `ContentStudio CLI` from the dashboard integrations carousel.
2. User generates or copies an API key from `Settings -> API Key`.
3. User uses the in-app `ContentStudio CLI & Agent Access` quickstart to copy install commands or open the setup docs.
4. User installs the CLI via npm or runs it with `npx`.
5. User authenticates with an API key or environment variable.
6. User runs `workspaces list` and `accounts list` to identify context.
7. User optionally uploads media.
8. User creates, lists, approves, comments on, or deletes posts from the CLI.
9. Agent users repeat the same flow with `--json` and the bundled `SKILL.md`.
10. Agent users can also install the standalone skill with `npx skills add contentstudio/contentstudio-agent` and use the same CLI contract.

---

## 8. Business Rules & Constraints

| Rule ID | Rule | Rationale |
| --- | --- | --- |
| BR-1 | The CLI must only use the existing public REST API as its source-of-truth backend contract | Keeps the launch grounded in a stable supported surface |
| BR-2 | Every shipped command must support `--json` | Required for automation and shell-capable agent use |
| BR-3 | `contentstudio auth login --api-key <key>` must validate credentials through `GET /api/v1/me` | Prevents saving bad credentials locally |
| BR-4 | The app remains the source of API key creation and rotation | Avoids adding new credential-management work to the CLI launch |
| BR-5 | V1 command scope is limited to routes already available in `routes/api/v1.php` | Prevents scope creep into unsupported surfaces |
| BR-6 | Error output in human mode must be concise and readable; error output in `--json` mode must be structured and script-safe | Supports both interactive and automated use |
| BR-7 | Website and docs must position the CLI as a safe automation surface with human-in-the-loop usage guidance | Avoids misleading AI/autonomy messaging |
| BR-8 | The standalone skill-install path must stay in sync with the published CLI command surface and auth contract | Prevents agent install paths from drifting away from the actual product |
| BR-9 | The app must treat the CLI as a first-party ContentStudio capability, not bury it inside third-party automation content | Preserves the correct product hierarchy and messaging |
| BR-10 | The detailed quickstart lives on `Settings -> API Key`, while dashboard surfaces act as discovery and routing entry points | Keeps the install/setup flow close to API key generation without overloading the dashboard |

---

## 9. Open Questions

| Question | Options | Owner | Due Date | Decision |
| --- | --- | --- | --- | --- |
| Should the CLI repo/package live in the existing mono-repo or a separate public package repo? | mono-repo package / separate repo | Engineering | Sprint start | Pending |
| Do we want `auth whoami` to show only user info or user info plus default workspace summary? | user only / user + workspace | Product + Engineering | Sprint 1 | Pending |
| Should launch docs live under the marketing website docs area, a dedicated developer docs area, or both? | website only / developer docs only / both | Product Marketing | Sprint 1 | Pending |
| Do we want a dedicated default workspace config command in v1 or only explicit `--workspace` flags? | explicit flags only / add default workspace | Product | Sprint 1 | Pending |

---

## 10. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
| --- | --- | --- | --- |
| Public API response shapes are inconsistent for CLI usage | Medium | High | Start with an API audit/hardening story before CLI command rollout |
| CLI contract sprawls into unsupported surfaces | Medium | High | Freeze v1 to publishing-first commands only |
| Docs and website ship late, weakening the launch | Medium | High | Keep docs and website work in the same epic, not as post-launch polish |
| Standalone skill install drifts from the CLI release | Medium | Medium | Add a dedicated distribution story and keep docs/website examples aligned with the same release |
| Positioning becomes vague "AI" language with no concrete workflow | Medium | Medium | Keep marketing anchored in install, auth, JSON, and command examples |
| The package boundary becomes a delivery blocker because the repo has no root Node workspace | Medium | Medium | Make package-boundary setup part of the CLI foundation story |

---

## 11. Dependencies

- **Internal:** public API controllers and middleware in `contentstudio-backend`; API key management flows already exposed in settings; website/marketing capacity for launch assets; web-app dashboard and settings surfaces in `contentstudio-frontend`
- **External:** npm publishing access; user API keys; supported public API uptime and rate limits
- **Blockers:** API audit must happen before command contract freeze; package-boundary decision must happen before CLI implementation work starts

---

## 12. Appendix

- Background strategy doc: `Public CLI Strategy` - https://app.shortcut.com/contentstudio-team/write/IkRvYyI6I3V1aWQgIjY5ZDc0MTFjLTU3ZDgtNDBmYy1hNTk5LWNmNTZiMjE0ZDRhYiI=
- Competitive research: `ContentStudio Public CLI & Agent Skills - Research` - https://app.shortcut.com/contentstudio-team/write/IkRvYyI6I3V1aWQgIjY5ZDdhZTJjLWExMTgtNDdmNS1hMmE4LTFkNmRjNzVkNjI3NCI=
- Shortcut PRD doc: `PRD: ContentStudio Public CLI & Agent Skills` - https://app.shortcut.com/contentstudio-team/write/IkRvYyI6I3V1aWQgIjY5ZDdhZTQ5LTZkNmMtNDBhNS05NjFjLTAyYTljNjZjODBmYSI=
- Relevant backend files:
  - `contentstudio-backend/routes/api/v1.php`
  - `contentstudio-backend/routes/api.php`
  - `contentstudio-backend/app/Http/Controllers/Api/V1/`
  - `contentstudio-backend/app/Http/Controllers/ApiKeyController.php`
  - `contentstudio-backend/app/Repository/ApiKeyRepo.php`
- Relevant frontend files:
  - `contentstudio-frontend/src/modules/setting/components/ApiKeysPage.vue`
  - `contentstudio-frontend/src/modules/setting/config/routes/setting.js`
  - `contentstudio-frontend/src/modules/dashboard/components/ApiCentricDashboard.vue`
  - `contentstudio-frontend/src/views/DashboardNew.vue`
  - `contentstudio-frontend/src/components/dashboard/IntegrationsCard.vue`
  - `contentstudio-frontend/src/components/dashboard/IntegrationCard.vue`

---

## Changelog

| Date | Author | Changes |
| --- | --- | --- |
| April 9, 2026 | Codex | Initial draft |
| April 10, 2026 | Codex | Added in-app CLI placement across API Key and dashboard surfaces |
