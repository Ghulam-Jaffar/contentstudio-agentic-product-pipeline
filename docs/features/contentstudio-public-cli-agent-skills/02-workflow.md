# Workflow - ContentStudio Public CLI & Agent Skills

Date: April 9, 2026

## Feature Placement

### 1. Website

The public entry point should be a dedicated CLI landing page plus supporting CTAs from developer-facing pages. The website explains:

- what the CLI is
- how to install it
- where to get an API key
- how to complete the first publish flow
- how shell-capable AI agents can use the same binary safely

### 2. App

The app remains the place where users create, rotate, and inspect API keys.

Primary app entry point:

- `Settings -> API Key`

### 3. npm surface

Primary install methods:

- `npm install -g @contentstudio/cli`
- `npx @contentstudio/cli --help`

Primary binary:

- `contentstudio`

### 4. Agent surface

The CLI package should include a bundled `SKILL.md` manifest that declares:

- binary name: `contentstudio`
- required env var: `CONTENTSTUDIO_API_KEY`
- core supported commands

A separate public skill repo should also be published for direct registry-style installation:

- `npx skills add contentstudio/contentstudio-agent`

## Happy Path

1. User lands on the ContentStudio CLI website page and sees a clear terminal-first value proposition.
2. User opens `Settings -> API Key` in the app and generates or copies an existing API key.
3. User installs the CLI with npm or runs it with `npx`.
4. User authenticates with either:
   - `contentstudio auth login --api-key cs_xxx`
   - `export CONTENTSTUDIO_API_KEY=cs_xxx`
5. User runs `contentstudio workspaces list` to confirm access.
6. User runs `contentstudio accounts list --workspace <id>` to discover connected social accounts.
7. User uploads media with `contentstudio media upload --workspace <id> --file ./asset.png` when needed.
8. User creates a post with `contentstudio posts create --workspace <id> --accounts <id1,id2> --text "..."`.
9. User verifies or manages output with `posts list`, `posts delete`, `posts approve`, and `comments` commands.
10. Agent users either rely on the bundled `SKILL.md` in the CLI package or install the standalone skill with `npx skills add contentstudio/contentstudio-agent`.
11. The agent runs the same CLI commands with `--json` and the operator keeps human review over publish-triggering actions.

## Alternative Flows

### Invalid or revoked key

1. User runs an auth or data command with an invalid key.
2. CLI validates against `GET /api/v1/me`.
3. CLI returns a concise message in human mode and structured error JSON in `--json` mode.
4. User is directed back to `Settings -> API Key` to generate or rotate a key.

### Missing workspace selection

1. User runs a workspace-scoped command without `--workspace`.
2. CLI returns a validation error that explains the missing workspace id.
3. User can recover by running `contentstudio workspaces list`.

### Media upload failure

1. User uploads a missing, invalid, or oversized file.
2. CLI surfaces the backend validation error cleanly.
3. User corrects the file path or asset and retries.

### Post validation failure

1. User attempts to publish with invalid account ids, missing content, or a rejected scheduling payload.
2. CLI returns the backend validation error with no raw stack trace.
3. User edits the command and reruns it.

### Agent execution path

1. Operator installs the CLI and sets `CONTENTSTUDIO_API_KEY`.
2. Operator optionally installs the standalone skill with `npx skills add contentstudio/contentstudio-agent`.
3. The shell-capable agent discovers the `contentstudio` binary and reads either the bundled `SKILL.md` or the installed standalone skill metadata.
4. The agent runs `workspaces list --json` and `accounts list --json` before attempting any post action.
5. The agent executes publish commands through the CLI and receives parseable JSON back.
6. Human-in-the-loop messaging in docs and examples makes it clear that operators should review destructive or publish-triggering flows.

## Key Design Decisions

### 1. CLI-first product surface

Decision: ship the public CLI first.

Rationale:

- The public publishing API already exists.
- A CLI is easier to install, document, version, and market than a broader public developer platform.
- The CLI is sufficient for shell-capable AI agents.

### 2. Reusable client layer inside the CLI

Decision: create one reusable TypeScript client layer inside the CLI codebase instead of publishing a separate SDK in v1.

Rationale:

- This keeps implementation clean without forcing a second public package.
- If reuse expands later, the client layer can be extracted.

### 3. API key auth only in v1

Decision: support API-key login and env-var auth only in v1.

Rationale:

- ContentStudio already has API key management in the app.
- The CLI strategy doc and existing public API both align around this model.
- This avoids adding browser/device-auth work to the launch.

### 4. JSON as an explicit contract

Decision: default output is human-readable, but `--json` is a first-class supported mode on every command.

Rationale:

- Human users should get readable tables and messages.
- Agents and automation need stable parseable output.

### 5. Publishing-first scope

Decision: v1 covers auth, workspaces, accounts, media, posts, approvals, and comments only.

Rationale:

- These routes already exist publicly.
- This gives ContentStudio a credible CLI launch without dragging in analytics or deeper workflow surfaces.

## Integration With Existing Product Areas

- Settings: users get keys from the existing API Key page.
- Publishing: CLI operates on the same posting and approval primitives as the app.
- Media Library: upload flows reuse the existing media endpoint.
- Website: launch page and docs surface the new terminal workflow publicly.

## V1 vs V2 Scope Recommendation

### Include in v1

- npm package and binary
- auth login with API key
- env-var auth
- workspace and account discovery
- media upload
- post create/list/delete
- post approval actions
- comments list/add
- `--json` support on all commands
- bundled `SKILL.md`
- standalone public skill repo installable via `npx skills add contentstudio/contentstudio-agent`
- install docs, command reference, quickstart, agent setup guide
- website launch page and conversion CTAs

### Defer to v2

- separately published SDK
- analytics commands
- bulk campaign import/export flows
- inbox or CRM commands
- richer interactive prompts
- shell completion
- additional agent-specific helper templates beyond the core `SKILL.md`
