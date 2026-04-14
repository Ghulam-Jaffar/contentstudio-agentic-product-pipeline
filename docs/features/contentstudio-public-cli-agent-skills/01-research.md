# Research - ContentStudio Public CLI & Agent Skills

Date: April 9, 2026

## What Is This Feature?

ContentStudio Public CLI & Agent Skills is a new public developer surface for ContentStudio.

The feature packages ContentStudio's existing public publishing API into a terminal-first product that users can install with npm, authenticate with an API key, and use for deterministic workflows such as listing workspaces, discovering accounts, uploading media, creating posts, reviewing approvals, and managing comments. The same CLI also becomes the agent-facing surface through a bundled `SKILL.md` manifest and machine-safe `--json` output.

The public launch should not stop at the website. The app should also close the loop by exposing first-party CLI setup inside `Settings -> API Key` and a lightweight discovery entry inside the dashboard integrations surfaces.

This is not "AI orchestration inside the CLI." The CLI should stay operational, script-safe, and predictable. The agent story comes from making the CLI easy for shell-capable agents to discover and execute.

## Why This Matters Now

- The market is moving toward agent-friendly product surfaces, not just browser dashboards.
- ContentStudio already has public API routes for publishing workflows, which lowers the cost of shipping a CLI now.
- The product already exposes API-key management and request-log experiences in the app, so users already have a credential and observability model the CLI can build on.
- ContentStudio currently lacks a public npm package, a terminal workflow, and an installable agent-skill surface. That is now a visible gap against newer automation-first entrants.

## Competitor Analysis

### Deep benchmarks

| Product | What they ship publicly | Auth model | Agent-facing pattern | Key takeaway |
| --- | --- | --- | --- | --- |
| Postiz | Public social media CLI plus skills-based install flow | API key env var (`POSTIZ_API_KEY`) | `npx skills add`, structured JSON output, install docs framed for agent use | Strong proof that a terminal product can double as an agent product when commands are deterministic and packaging is lightweight |
| Zernio | Public npm CLI, SDKs, API docs, account-connect guides | Browser login or manual API key | JSON by default, explicit CLI docs, AI-agent positioning in public docs | Strong proof that a serious developer surface can be marketed as "built for developers and AI agents" without collapsing into vague AI marketing |

### Traditional suite competitor scan

The default competitor set in `.claude/shortcut-config.json` was reviewed at a public-positioning level. The broad pattern is consistent: these products market dashboards, scheduling, analytics, and automation integrations, but do not prominently surface a public npm CLI or skills-based agent install flow comparable to Postiz or Zernio.

| Competitor | Public positioning in scan | Public CLI / agent-skill surface found? | Takeaway |
| --- | --- | --- | --- |
| Buffer | Dashboard + API/integrations | No public CLI surfaced in scan | API exists, but terminal-first packaging is not the story |
| Hootsuite | Enterprise suite + dev/integration ecosystem | No public CLI surfaced in scan | Integration-first, not agent-first |
| Publer | Publishing and automation UI | No public CLI surfaced in scan | Workflow product, not terminal product |
| Later | Creator/social planning UI | No public CLI surfaced in scan | Not marketed to developer workflows |
| Sprout Social | Enterprise suite | No public CLI surfaced in scan | High-end suite, not developer-first distribution |
| Loomly | Planning and approvals UI | No public CLI surfaced in scan | Team workflow focus |
| Sendible | Agency dashboard | No public CLI surfaced in scan | Agency UI, not terminal UI |
| SocialBee | Scheduling and automations UI | No public CLI surfaced in scan | Dashboard-first |
| Agorapulse | Inbox/publishing suite | No public CLI surfaced in scan | Ops workflow, not developer workflow |
| Metricool | Analytics + publishing UI | No public CLI surfaced in scan | Reporting/product UI focus |

Inference: the whitespace is not "another API." It is a usable terminal surface that is easy to install, easy to script, and easy for AI agents to operate safely.

## What Users Will Expect

### Table stakes

- npm install and `npx` entry points
- clear API key authentication
- list workspaces and accounts before attempting a post
- media upload before post creation
- stable JSON output for automation
- predictable exit codes and concise error messages
- an in-app quickstart section next to API key management so users can move from key creation to CLI usage immediately

### Competitive delighters

- a single binary that covers discovery, upload, publishing, approvals, and comments
- a bundled `SKILL.md` manifest so no extra glue code is needed for shell-capable agents
- a standalone public skill-install path such as `npx skills add contentstudio/contentstudio-agent`
- launch docs with copy-pasteable workflows for humans and agents
- first-party in-app discovery on the API Key page and dashboard, without burying the CLI inside third-party automation cards
- marketing that speaks to developers and automation operators without overpromising autonomy

## Recommended Product Shape For ContentStudio

1. Ship `@contentstudio/cli` as the primary public surface.
2. Keep the CLI thin over the existing public REST API in `contentstudio-backend/routes/api/v1.php`.
3. Implement one reusable TypeScript client layer inside the CLI codebase instead of publishing a standalone SDK on day 1.
4. Make `--json` a first-class contract, not an afterthought.
5. Bundle a `SKILL.md` manifest so shell-capable agents can discover the CLI and required env vars.
6. Publish a standalone public skill repo so users can install ContentStudio into supported agent registries with a one-liner such as `npx skills add contentstudio/contentstudio-agent`.
7. Add a first-party `ContentStudio CLI & Agent Access` section to `Settings -> API Key`, above Zapier/Make.com/n8n, so the app teaches the official install flow where users already generate keys.
8. Add a `ContentStudio CLI` discovery card to the shared dashboard integrations carousel used by both standard and API-centric dashboards.
9. Keep v1 publishing-first: auth, workspaces, accounts, media, posts, approvals, comments.
10. Include launch assets in the same epic: install guide, command reference, agent setup guide, website landing/supporting CTAs, and in-app setup entry points.

## Codebase Analysis

### Existing related code

#### Public publishing API already exists

The backend already exposes API-key-protected public API routes in `contentstudio-backend/routes/api/v1.php`, including:

- `GET /api/v1/me`
- `GET /api/v1/workspaces`
- `GET /api/v1/workspaces/{workspace_id}/accounts`
- `GET|POST /api/v1/workspaces/{workspace_id}/media`
- `GET|POST|DELETE /api/v1/workspaces/{workspace_id}/posts`
- `POST /api/v1/workspaces/{workspace_id}/posts/{post_id}/approval`
- `GET|POST /api/v1/workspaces/{workspace_id}/posts/{post_id}/comments`

Relevant backend controllers already exist in `contentstudio-backend/app/Http/Controllers/Api/V1/`, including:

- `UserController.php`
- `WorkspaceController.php`
- `AccountController.php`
- `MediaController.php`
- `PostController.php`
- `CommentController.php`
- `CampaignController.php`
- `LabelController.php`
- `TeamMemberController.php`

This means the CLI can be built as a product layer over an existing contract instead of requiring a new backend surface first.

#### API key management already exists in the product

The app already has session-authenticated API key management and request-log experiences:

- `contentstudio-backend/routes/api.php` exposes `GET /api-keys`, `POST /api-keys`, `POST /api-keys/{id}/revoke`, `POST /api-keys/{id}/regenerate`
- `contentstudio-backend/app/Http/Controllers/ApiKeyController.php` handles read/create/revoke/regenerate
- `contentstudio-backend/app/Repository/ApiKeyRepo.php` standardizes key generation with the `cs_` prefix
- `contentstudio-frontend/src/modules/setting/components/ApiKeysPage.vue` already presents API key credentials, request usage, and request logs
- `contentstudio-frontend/src/modules/setting/config/routes/setting.js` already routes users to `Settings -> API Key`
- `contentstudio-frontend/src/modules/dashboard/components/ApiCentricDashboard.vue` already includes API-key CTAs and the shared integrations carousel
- `contentstudio-frontend/src/views/DashboardNew.vue` already renders the same `IntegrationsCard.vue` component on the standard dashboard
- `contentstudio-frontend/src/components/dashboard/IntegrationsCard.vue` already powers the "Content Creation & Automation Tools" carousel used for product discovery

This is important because the CLI launch does not need to invent a credential-management experience. Users can keep generating and rotating keys in the app.

#### Request logs and usage already exist

The app already exposes request-log endpoints for API key usage in `contentstudio-backend/routes/api.php` and the frontend already renders request-log UI in `ApiKeysPage.vue`. That reduces the support burden for a CLI launch because users already have a place to inspect what their key is doing.

### Reusable foundations

- Existing REST controllers and validation requests provide the data plane for the CLI.
- Existing API key creation and rotation flows provide the auth plane.
- Existing request logs provide an observability plane for support and debugging.
- Existing public docs copy in `docs/technical/public-cli-strategy-2026-04-08.md` provides strong product framing that can be converted into a feature-specific Research doc, PRD, and website copy.

### Gaps

- No public npm package currently exists in this repo.
- No root JavaScript workspace exists today other than `cs-prototypes/package.json`, so the CLI will likely need a fresh package boundary.
- No public `SKILL.md` manifest exists for ContentStudio.
- No standalone public skill repo exists for direct agent-registry installation.
- No install guide, command reference, or agent setup guide exists for a public CLI.
- No website page currently markets ContentStudio as a terminal-first or agent-friendly product surface.

## Technical Considerations

- The CLI should use the existing `X-API-Key` contract rather than inventing a new auth scheme.
- The CLI should validate credentials by calling `GET /api/v1/me`.
- The public API audit story should explicitly verify response shapes, validation behavior, and upload flows against CLI needs before the CLI contract is frozen.
- Because the root repo is not already organized as a Node workspace, the CLI foundation story should decide and implement a clean package boundary before command work starts.

## Risks

| Risk | Why it matters | Mitigation |
| --- | --- | --- |
| Weak API consistency | CLI users will feel every inconsistency immediately | Start with an API audit/gap-closure story before command expansion |
| Over-scoping v1 | Inbox, analytics, or advanced automation would slow launch materially | Keep v1 limited to publishing-first public API routes |
| Weak onboarding | A great CLI with poor docs will still fail | Launch docs and website work are part of the same epic, not follow-up polish |
| AI-agent positioning gets vague | "Agentic" language can become fluff fast | Anchor all messaging in explicit install, auth, JSON, and command workflows |

## Sources

### External

- Postiz public agent page: https://postiz.com/hermes-agent
- Zernio CLI docs: https://docs.zernio.com/resources/cli
- Zernio connecting accounts guide: https://docs.zernio.com/guides/connecting-accounts

### Internal

- `contentstudio-backend/routes/api/v1.php`
- `contentstudio-backend/routes/api.php`
- `contentstudio-backend/app/Http/Controllers/Api/V1/`
- `contentstudio-backend/app/Http/Controllers/ApiKeyController.php`
- `contentstudio-backend/app/Repository/ApiKeyRepo.php`
- `contentstudio-frontend/src/modules/setting/components/ApiKeysPage.vue`
- `contentstudio-frontend/src/modules/setting/config/routes/setting.js`
- `docs/technical/public-cli-strategy-2026-04-08.md`
