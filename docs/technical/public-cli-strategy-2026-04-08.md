# Public CLI Strategy

Date: April 8, 2026

Audience: Product, Engineering, Platform

Scope: Define how ContentStudio should ship a public npm-installed CLI for end users and developers, what it should cover in v1, and how it should be implemented on top of the existing public API.

## Executive Summary

Yes, shipping a public CLI is a good direction.

The CLI should be:

- installed with npm
- authenticated with a ContentStudio API key
- built on top of the existing public REST API
- implemented through a reusable TypeScript client layer
- focused on deterministic operational workflows, not internal AI orchestration

The right architecture is:

1. Public REST API as the source of truth
2. Reusable TypeScript client layer
3. Public CLI on top of that client layer

The wrong architecture is:

- making the CLI call internal agent code
- making the CLI depend on MCP
- building ad hoc raw HTTP calls directly into each CLI command

## Goal

We want a user to be able to do this:

```bash
npm install -g @contentstudio/cli
contentstudio auth login --api-key cs_xxx
contentstudio workspaces list
contentstudio accounts list --workspace <id>
contentstudio posts create --workspace <id> --accounts <id1,id2> --text "Hello"
```

And we want that experience to be:

- easy to install
- easy to understand
- safe for automation
- stable across versions

## Current Foundation

The backend already exposes API-key-protected public API routes in `contentstudio-backend/routes/api/v1.php`.

Useful v1 CLI-compatible endpoints already exist for:

- user identity
- workspaces
- accounts
- content categories
- labels
- campaigns
- media
- team members
- posts
- approvals
- comments

That means the CLI does not need a new backend architecture to start. It needs a clean product layer on top of the current public API.

## Product Positioning

The CLI is for:

- developers
- technical operators
- automation users
- support/internal teams
- advanced customers who prefer terminal workflows

The CLI is not:

- a replacement for the web app
- a wrapper over internal AI-agent workflows
- a substitute for a public MCP integration

## Recommended Architecture

## 1. Public API

The CLI should treat the public REST API as the canonical backend contract.

Responsibilities of the API:

- API key authentication
- workspace permission checks
- request validation
- rate limiting
- stable request/response schemas
- logging and auditability

The CLI should never bypass this layer.

## 2. Reusable TypeScript Client Layer

Before building many CLI commands, create a reusable client layer.

This does **not** have to be a separately published package on day 1.

Recommended MVP approach:

- start with an internal client layer inside the CLI codebase
- keep it clean, typed, and reusable
- extract it into `@contentstudio/sdk` later if reuse justifies it

Possible future extracted package:

- `@contentstudio/sdk`

Responsibilities:

- base URL configuration
- API key auth header injection
- retries and timeout policy
- typed request/response models
- pagination helpers
- normalized errors
- upload helpers

The CLI should depend on this client layer instead of implementing raw fetch logic command-by-command.

For v1, the key requirement is **one reusable client implementation**, not necessarily one separately published SDK package.

## 3. CLI package

Publish a separate package:

- `@contentstudio/cli`

Install options:

```bash
npm install -g @contentstudio/cli
```

or

```bash
npx @contentstudio/cli --help
```

Recommended binary name:

```bash
contentstudio
```

## CLI Principles

The CLI should be:

- human-readable by default
- machine-friendly with `--json`
- script-safe with predictable exit codes
- explicit rather than magical
- thin over the reusable client layer

The CLI should not:

- leak raw backend stack traces
- force users through browser-only auth for v1
- expose unstable internal concepts like workflow session blobs

## V1 Scope

Recommended v1 commands:

### Auth

```bash
contentstudio auth login --api-key <key>
contentstudio auth whoami
contentstudio auth logout
```

### Workspaces

```bash
contentstudio workspaces list
```

### Accounts

```bash
contentstudio accounts list --workspace <id>
```

### Posts

```bash
contentstudio posts list --workspace <id>
contentstudio posts create --workspace <id> --accounts <id1,id2> --text "Hello"
contentstudio posts delete --workspace <id> --post <id>
contentstudio posts approve --workspace <id> --post <id> --action approve
contentstudio posts approve --workspace <id> --post <id> --action reject --comment "Needs changes"
```

### Media

```bash
contentstudio media list --workspace <id>
contentstudio media upload --workspace <id> --file ./asset.png
```

### Comments

```bash
contentstudio comments list --workspace <id> --post <id>
contentstudio comments add --workspace <id> --post <id> --text "Looks good"
```

### Discovery helpers

```bash
contentstudio categories list --workspace <id>
contentstudio labels list --workspace <id>
contentstudio campaigns list --workspace <id>
contentstudio team-members list --workspace <id>
```

## V1 Features

### Must-have

- API key login
- config persistence
- workspace/account/post/media/comment commands
- `--json` output
- useful help text
- predictable error handling

### Nice-to-have

- interactive prompts when flags are omitted
- shell completion
- aliases
- CSV export

### Out of scope for v1

- full AI chat workflows
- streaming AI generation UX
- internal workflow state management
- MCP/server behavior

## Auth Model

Use API key auth for v1.

### Login flow

```bash
contentstudio auth login --api-key cs_xxx
```

The CLI should validate the key by calling `GET /api/v1/me`.

### Config storage

Store config in a user-local config file:

- macOS/Linux: `~/.config/contentstudio/config.json`
- Windows: equivalent platform config directory

Recommended config shape:

```json
{
  "baseUrl": "https://app.contentstudio.io",
  "apiKey": "cs_xxx",
  "defaultWorkspaceId": "workspace_id"
}
```

Later enhancement:

- OS keychain integration

But for v1, simple config storage is acceptable if documented clearly.

## Output Design

### Default mode

Readable tables or concise summaries.

Example:

```bash
contentstudio workspaces list
```

Output:

```text
ID           Name              Role
ws_123       Marketing Team    owner
ws_456       Client Alpha      admin
```

### JSON mode

```bash
contentstudio workspaces list --json
```

Output should be raw structured JSON suitable for scripting.

### Exit codes

- `0` success
- non-zero for validation/auth/network/server failures

## Error Handling

The CLI should map backend failures into clean user-facing messages.

Examples:

- missing API key
- invalid API key
- workspace permission denied
- validation error on post creation
- rate limit hit
- network timeout

The CLI should show:

- one concise message by default
- optional verbose/debug mode for raw details

Example:

```bash
contentstudio posts create ...
Error: API key is invalid or expired.
```

Not:

- raw stack traces
- unformatted backend payloads

## Package Structure

Recommended:

```text
packages/
  cli/
```

Suggested internal structure:

```text
packages/
  cli/
    src/
      client/
      commands/
      config/
      output/
```

### `packages/cli`

- internal API client layer
- command parser
- interactive prompts
- output formatting
- config file handling
- command implementations using the internal client layer

Possible future extraction:

```text
packages/
  sdk/
  cli/
```

Only do that once another consumer actually needs the same client code.

## Suggested Tech Stack

Recommended:

- TypeScript
- Node.js
- `commander` or `yargs` for command parsing
- `chalk` or minimal color support
- `ora` only if progress spinners add real value
- `undici`/native fetch or a small HTTP abstraction through the client layer

Avoid:

- heavy framework choices for a small CLI
- command implementations tightly coupled to UI-specific assumptions

## Release Strategy

### Phase 1: Client foundation

- build typed API client
- cover current public API v1 endpoints used by CLI
- normalize auth/errors

### Phase 2: CLI MVP

- ship `auth`, `workspaces`, `accounts`, `posts`, `media`, `comments`
- support npm global install and `npx`
- document quickstart

### Phase 3: CLI hardening

- interactive prompts
- default workspace support
- completions
- better bulk operations

### Phase 4: Optional SDK extraction

- extract the internal client layer into `@contentstudio/sdk` only if another consumer needs it
- keep CLI as the first shipping product surface

## Documentation Requirements

To make the CLI successful, we need:

- installation docs
- auth docs
- command reference
- examples
- automation examples using `--json`
- troubleshooting guide

Minimum onboarding target:

1. user installs the package
2. user pastes API key
3. user runs `contentstudio workspaces list`
4. user successfully creates or lists content in under 2 minutes

## Risks

### 1. Weak public API contract

If the API responses are inconsistent, the CLI will feel unstable even if the terminal UX is good.

### 2. No reusable client layer

If commands call raw endpoints independently, auth, retries, and error handling will diverge quickly.

### 3. Over-scoping v1

If we try to include AI workflows or deeply interactive planning flows too early, the CLI will become brittle and hard to support.

### 4. Poor install/onboarding

If npm install works but auth/setup is confusing, users will abandon the CLI immediately.

## Recommendation

Proceed with a CLI initiative, but keep it narrow and productized:

1. Public API remains the source of truth
2. Build a reusable client layer inside the CLI first
3. Build `@contentstudio/cli`
4. Extract `@contentstudio/sdk` later only if reuse demands it
5. Ship a focused operational CLI first

This is the right direction because it gives ContentStudio:

- a developer-friendly terminal surface
- a better automation story
- a foundation for future MCP or other integrations

But the CLI should be treated as its own product surface, not as a side effect of internal agent tooling.
