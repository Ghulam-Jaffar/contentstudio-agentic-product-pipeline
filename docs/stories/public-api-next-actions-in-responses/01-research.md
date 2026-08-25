# Public API: Next Actions in Responses · Research

## Answer to the question: does anything like this exist today?

**No.** The public API has no next-action, hint, suggestion, links, meta or actions concept anywhere. Grepped `next_action`, `next_steps`, `suggested_action`, `remediation`, `hint`, `_links`, `links`, `meta`, `actions` and `HATEOAS` across `app/Http/Controllers/Api/V1/` and the wider backend. Nothing.

## Current State

### The response envelope

Every public API response is built through `contentstudio-backend/app/Http/Controllers/Api/V1/BaseApiController.php`. Errors carry exactly four keys:

```
{ status: false, message: "...", error_code: "VALIDATION_ERROR", errors: {...} }
```

`localizedFlatError()` (line 143) does accept an `$extra` array that gets merged into the envelope, so there is an existing seam for adding a field without reworking the base controller. It is currently used for one-off additions only.

### Error codes are patchy, and the exact example given has none

22 distinct error codes exist across the whole public API:

`API_ACCESS_NOT_ALLOWED`, `CAMPAIGN_CREATE_FAILED`, `CAMPAIGN_DELETE_FAILED`, `CAMPAIGN_NOT_FOUND`, `CAMPAIGN_UPDATE_FAILED`, `INBOX_CONTRACT_VIOLATION`, `INBOX_UPSTREAM_ERROR`, `LABEL_CREATE_FAILED`, `LABEL_DELETE_FAILED`, `LABEL_NOT_FOUND`, `LABEL_UPDATE_FAILED`, `PROFILE_FETCH_FAILED`, `REQUIRES_REMOVAL_CONFIRMATION`, `SCHEDULING_UPSTREAM_ERROR`, `TEAM_MEMBER_ADD_FAILED`, `TEAM_MEMBER_LIMIT_REACHED`, `TEAM_MEMBER_NOT_FOUND`, `TEAM_MEMBER_REMOVE_FAILED`, `TEAM_MEMBER_UPDATE_FAILED`, `VALIDATION_ERROR`, `WORKSPACE_DELETE_FAILED`, `WORKSPACE_SAVE_FAILED`.

**There is no error code for an invalid or unknown workspace**, which is the exact case raised. That path lives in `contentstudio-backend/app/Http/Middleware/ApiKeyMiddleware.php:69-74` and returns a bare message with **no `error_code` field at all**:

```
{ status: false, message: "<workspace not found message>" }  // 404
```

So an agent hitting the most common publishing mistake gets an untyped, localized sentence and nothing machine-readable. It cannot even reliably detect *which* error it hit, let alone what to do about it.

The same middleware returns `API_ACCESS_NOT_ALLOWED` (403) when the workspace has no API credits, and a bare message for a locked workspace (line 77-82) — again with no code.

### Publishing endpoints are the weakest

`contentstudio-backend/app/Http/Controllers/Api/V1/PostController.php` is the largest controller in the API. Its error paths frequently:
- pass through whatever upstream produced, e.g. `'error_code' => $result['error_code'] ?? null` (line 3476), so the code can be `null`
- call `responseError($message, null, 422)` (line 3481) with an explicit `null` code

So publishing, the surface an agent uses most, is the least typed part of the API.

### The closest existing thing

`contentstudio-backend/app/Mcp/Tools/HelpTool.php` is a static guidance document exposed as an MCP tool. Its overview (lines 214-245) literally spells out the sequence an agent should follow:

> 1. Validate token
> 2. Fetch workspaces → ASK user to select
> 3. Fetch accounts → ASK user to select
> 4. Ask for content details
> 5. Show preview
> 6. Create post → only after user confirms

That is exactly the knowledge this ticket wants to put into responses, but today it is:
- a **separate tool the agent must know to call**, not something the failing response tells it
- **static** — it does not know what actually went wrong on a given call
- **MCP only** — the public API, CLI and any direct integrator get none of it

The backend MCP server exposes 8 tools total (`CreatePostTool`, `DeletePostTool`, `FetchPostsTool`, `FetchSocialAccountsTool`, `FetchWorkspacesTool`, `HelpTool`, `PingTool`, `ValidateTokenTool`).

## Why this matters now

Two things in flight make this more valuable than it would have been a year ago:

- The AI agents platform is replacing its MCP tools with an internal toolkit layer built **on the public API**. Those tools will consume these exact responses. Untyped errors mean each tool re-implements its own guesswork from message strings.
- The **developer surface parity contract** work in this sprint is about making MCP, CLI, API and other apps behave consistently. Next actions belong in that contract, otherwise only one surface will ever have them.

## Design Traps

- **`message` is localized.** `LocalizationHelper::apiResponse()` translates it. If next actions are written as prose they will be translated too, and an agent keying off English text breaks the moment a user's locale changes. The machine-readable part must be stable identifiers that are never translated, with any human sentence kept separate.
- **The gap is bigger than adding a field.** Roughly half the failure paths that matter for publishing return `null` or absent error codes. Next actions cannot be attached to an error that has no identity, so typing the errors is the prerequisite, not a nice-to-have alongside.
- **Additive only.** Existing integrators parse the current four keys. A new field is safe; renaming or restructuring is not.
- **Do not leak internals.** Upstream error codes from the scheduling and inbox microservices (`SCHEDULING_UPSTREAM_ERROR`, `INBOX_UPSTREAM_ERROR`) are passed through today. Next actions should be expressed in terms of the public API's own capabilities, not internal service names.

## Scope Note

Publishing is the right place to start and where the pain is, but the value only lands if the shape is defined once centrally rather than hand-rolled per endpoint. The base controller is the natural home.

## Mobile Context

Not impacted. The public API is a developer surface.

## Files Involved

- `contentstudio-backend/app/Http/Controllers/Api/V1/BaseApiController.php` — envelope helpers, the `$extra` seam
- `contentstudio-backend/app/Http/Middleware/ApiKeyMiddleware.php` — untyped workspace, locked-workspace and credit errors
- `contentstudio-backend/app/Http/Controllers/Api/V1/PostController.php` — publishing error paths, null codes
- `contentstudio-backend/app/Mcp/Tools/HelpTool.php` — the static guidance that should become dynamic
- Public API reference docs, wherever the error catalog is published
