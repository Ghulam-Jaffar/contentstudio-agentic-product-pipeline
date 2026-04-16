# Stories: Social Account Connection via API

## Epic: Publishing API v1.13 — Social Account Connection

---

## Story 1: [BE] Add social account OAuth connection endpoints to Publishing API v1

**Epic:** Publishing API v1.13 — Social Account Connection

### Description:
As a developer using the ContentStudio API, I want to initiate a social account connection via API so that I can connect accounts programmatically from CLI tools, MCP agents, or custom integrations without manually navigating the web app.

This adds two new endpoints to the public API v1 (`contentstudio-backend/routes/api/v1.php`) that wrap the existing OAuth connection infrastructure in `IntegrationController::getAuthorizationUrl()` and `IntegrationController::checkConnectorState()`. No new OAuth flows are created — the API generates the same OAuth URL the web UI uses and exposes the same Redis-backed connection state for polling.

**Endpoints:**
- `POST /api/v1/workspaces/{workspace_id}/accounts/connect` — Accepts `{platform, callback_url?}`, saves connection details to cache, returns `{auth_url, connection_id}`. The caller opens `auth_url` in a browser to complete OAuth.
- `GET /api/v1/workspaces/{workspace_id}/accounts/connect/{connection_id}/status` — Returns the connection state: `pending`, `completed` (with connected account details), or `failed` (with error message). The caller polls this until the OAuth callback completes.

**Supported platforms:** facebook, instagram, twitter, linkedin, gmb, pinterest, threads, tumblr, youtube, tiktok

---

### Workflow:
1. Developer sends `POST /api/v1/workspaces/{workspace_id}/accounts/connect` with `{"platform": "facebook"}`
2. API returns `{"status": true, "auth_url": "https://facebook.com/oauth/...", "connection_id": "abc123"}`
3. Developer opens the `auth_url` in a browser (or presents it to the user)
4. User completes the OAuth flow on the platform's site — platform redirects back to the existing ContentStudio callback URL
5. Developer polls `GET /api/v1/workspaces/{workspace_id}/accounts/connect/abc123/status`
6. API returns `{"status": true, "state": "pending"}` while waiting
7. Once OAuth callback completes, API returns `{"status": true, "state": "completed", "accounts": [{...}]}` with the newly connected account(s)
8. If OAuth fails or the user denies access, API returns `{"status": true, "state": "failed", "error": "User denied access"}`

---

### Acceptance criteria:
- [ ] `POST /api/v1/workspaces/{workspace_id}/accounts/connect` accepts `platform` (required) and returns `{auth_url, connection_id}` for supported platforms
- [ ] The endpoint validates `platform` against the supported platforms list and returns a clear validation error for unsupported platforms
- [ ] The generated `auth_url` is identical to what the web UI generates via `IntegrationController::getAuthorizationUrl()` — same OAuth scopes, same callback URLs
- [ ] `connection_id` is a unique identifier that maps to the cached connection details (same Redis/cache mechanism used by `checkConnectorState`)
- [ ] `GET /api/v1/workspaces/{workspace_id}/accounts/connect/{connection_id}/status` returns `pending`, `completed`, or `failed` state
- [ ] When state is `completed`, the response includes the connected account(s) details (account ID, platform, account name, profile picture)
- [ ] When state is `failed`, the response includes an error message
- [ ] Connection state expires after 10 minutes (same TTL as existing cache entries) — polling after expiry returns a clear "connection expired" message
- [ ] Both endpoints require API key authentication (same `api.key` middleware as other v1 routes)
- [ ] Both endpoints require workspace permission (same `PermissionMiddleware` as other v1 routes)
- [ ] Rate limiting is applied consistently with other v1 endpoints
- [ ] API request logging works for both new endpoints

---

### Mock-ups:
N/A — backend-only, API endpoints.

---

### Impact on existing data:
- No schema changes — connection details use the same temporary Redis/cache mechanism as the web UI flow
- No changes to how connected accounts are stored in MongoDB after OAuth completes
- The existing OAuth callback routes (`/facebook/connect`, `/twitter/connect`, etc.) are unchanged

---

### Impact on other products:
- Mobile apps: Not affected — mobile apps do not use the public API v1
- Chrome extension: Not affected
- White-label: The API endpoints work for all white-label domains — the generated OAuth URLs will use the correct callback URLs per domain
- MCP server: Will consume these endpoints (see **[BE] Add connect_social_account tool to ContentStudio MCP server**)
- CLI: Will consume these endpoints (see **[BE] Add accounts connect command to ContentStudio CLI**)

---

### Dependencies:
None — builds on existing OAuth infrastructure in `IntegrationController`.

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

## Story 2: [BE] Add connect_social_account tool to ContentStudio MCP server

**Epic:** Publishing API v1.13 — Social Account Connection

### Description:
As an AI agent connected to ContentStudio via MCP, I want a `connect_social_account` tool so that I can help users connect their social media accounts without leaving the agent conversation.

This adds a new MCP tool in `contentstudio-backend/app/Mcp/Tools/ConnectSocialAccountTool.php` that calls the new public API endpoint (`POST /api/v1/workspaces/{workspace_id}/accounts/connect`) to generate an OAuth URL, and a companion `check_account_connection_status` tool to poll for completion.

The tool follows the same pattern as existing MCP tools (`FetchSocialAccountsTool`, `CreatePostTool`) — accepts an `auth_key` parameter, calls the internal API, and returns structured results.

---

### Workflow:
1. User tells the AI agent: "Connect my Facebook account to ContentStudio"
2. Agent calls `connect_social_account` tool with `{workspace_id, auth_key, platform: "facebook"}`
3. Tool returns `{auth_url: "https://...", connection_id: "abc123", message: "Open this URL to connect your Facebook account"}`
4. Agent presents the URL to the user — user clicks it, completes OAuth in their browser
5. Agent calls `check_account_connection_status` tool with `{workspace_id, auth_key, connection_id: "abc123"}`
6. Tool returns the connection state — agent reports success or failure to the user

---

### Acceptance criteria:
- [ ] New `connect_social_account` MCP tool exists at `contentstudio-backend/app/Mcp/Tools/ConnectSocialAccountTool.php`
- [ ] Tool accepts parameters: `workspace_id` (required), `auth_key` (required), `platform` (required)
- [ ] Tool returns `{status, auth_url, connection_id, message}` on success
- [ ] Tool returns a clear error for invalid/unsupported platforms
- [ ] New `check_account_connection_status` MCP tool exists at `contentstudio-backend/app/Mcp/Tools/CheckAccountConnectionStatusTool.php`
- [ ] Status tool accepts: `workspace_id` (required), `auth_key` (required), `connection_id` (required)
- [ ] Status tool returns `{status, state, accounts?, error?}` matching the API endpoint response
- [ ] Both tools are registered in `contentstudio-backend/config/mcp.php`
- [ ] Both tools follow the same error handling pattern as existing MCP tools (auth validation, try/catch with Sentry)
- [ ] Tool descriptions are clear enough for an AI agent to understand when and how to use them

---

### Mock-ups:
N/A — backend/MCP tooling.

---

### Impact on existing data:
- No schema changes — tools call the API endpoints which use existing cache/Redis mechanisms

---

### Impact on other products:
- MCP clients (AI agents) gain the ability to initiate social account connections
- No impact on web app, mobile, or Chrome extension

---

### Dependencies:
Depends on **[BE] Add social account OAuth connection endpoints to Publishing API v1** — the MCP tools call the new API endpoints.

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support — N/A, MCP tool responses are English-only
- [ ] UI theming support — N/A, backend-only story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

## Story 3: [BE] Add accounts connect command to ContentStudio CLI

**Epic:** ContentStudio Public CLI & Agent Skills (epic 115952)

### Description:
As a developer using the ContentStudio CLI, I want an `accounts connect` command so that I can connect social media accounts from the terminal without opening the web app.

This adds a new `accounts connect` subcommand to the CLI package that calls the public API endpoint (`POST /api/v1/workspaces/{workspace_id}/accounts/connect`) to generate an OAuth URL, opens the user's default browser, and polls the status endpoint until the connection completes or times out.

The command follows the same patterns as existing CLI commands (`accounts list`, `posts create`) — supports both human-readable and `--json` output modes.

---

### Workflow:
1. User runs `contentstudio accounts connect --workspace <id> --platform facebook`
2. CLI calls the API to generate the OAuth URL
3. CLI opens the user's default browser to the OAuth URL and prints: "Opening your browser to connect Facebook... Complete the authorization in your browser."
4. CLI polls the status endpoint every 3 seconds
5. On success: CLI prints "Successfully connected: My Business Page (facebook)" with the account details
6. On failure: CLI prints the error message from the API
7. On timeout (5 minutes): CLI prints "Connection timed out. Please try again."
8. With `--json` flag: CLI outputs structured JSON at each stage instead of human-readable text
9. With `--no-browser` flag: CLI prints the URL without opening the browser (useful for headless environments / SSH sessions)

---

### Acceptance criteria:
- [ ] `contentstudio accounts connect --workspace <id> --platform <name>` initiates the OAuth connection flow
- [ ] CLI opens the user's default browser to the OAuth URL automatically
- [ ] `--no-browser` flag prints the URL to stdout without opening a browser
- [ ] CLI polls the connection status every 3 seconds with a spinner/progress indicator in human mode
- [ ] On successful connection, CLI displays the connected account name, platform, and account ID
- [ ] On failure, CLI displays the error message and exits with a non-zero code
- [ ] After 5 minutes without completion, CLI times out with a clear message and non-zero exit code
- [ ] `--json` flag outputs structured JSON for each state (url_generated, polling, completed, failed, timeout)
- [ ] Running `contentstudio accounts connect` without `--platform` shows a list of supported platforms to choose from
- [ ] The command is documented in the CLI `--help` output and in the bundled `SKILL.md`
- [ ] The command works with the existing `CONTENTSTUDIO_API_KEY` auth mechanism

---

### Mock-ups:
N/A — CLI command, terminal output only.

---

### Impact on existing data:
- No data changes — the CLI calls the API which uses existing OAuth infrastructure

---

### Impact on other products:
- Extends the CLI capabilities for developers and AI agents using the shell skill
- The `SKILL.md` bundled with the CLI package should be updated to include the `accounts connect` command
- No impact on web app, mobile, or Chrome extension

---

### Dependencies:
Depends on **[BE] Add social account OAuth connection endpoints to Publishing API v1** — the CLI calls the new API endpoints.

---

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A, CLI tooling story
- [ ] Multilingual support — N/A, CLI output is English-only
- [ ] UI theming support — N/A, CLI tooling story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)
