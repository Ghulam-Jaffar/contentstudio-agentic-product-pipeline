# Research: Social Account Connection via API

## Current State

### Public API v1 (`contentstudio-backend/routes/api/v1.php`)
- `GET /api/v1/workspaces/{workspace_id}/accounts` — Lists connected social accounts (read-only)
- No endpoint exists to **initiate** a social account connection via API
- All account connection is browser-only via OAuth redirects

### Account Connection Flow (Web UI)
The web UI connection flow is in `contentstudio-backend/app/Http/Controllers/Integrations/IntegrationController.php`:

1. **`POST /getAuthorizationUrl`** — Frontend sends `{platform, workspace_id, connection_details}`. Backend saves connection details to cache, generates an OAuth URL with encrypted state, returns `{auth_url}`.
2. **User is redirected to platform OAuth page** (Facebook, Instagram, Twitter, LinkedIn, etc.)
3. **Platform redirects back** to callback route (e.g., `/facebook/connect`) handled by `connectAccounts()`
4. **`POST /checkConnectorState`** — Frontend polls this to check if the OAuth callback completed. Returns the connection result from Redis.

Key classes:
- `IntegrationController::getAuthorizationUrl()` — generates OAuth URL
- `IntegrationController::checkConnectorState()` — checks connection result from Redis
- `Connector` / `IntegrationConnector` — platform-specific OAuth URL builders
- `IntegrationBuilder::processIntegration()` — processes the OAuth callback

### MCP Server (`contentstudio-backend/app/Mcp/Tools/`)
Existing tools: `fetch_social_accounts`, `fetch_workspaces`, `create_post`, `delete_post`, `fetch_posts`, `validate_token`, `ping`, `help`
- No `connect_social_account` tool exists

### CLI Epic (115952 — ContentStudio Public CLI & Agent Skills)
- CLI has `accounts list` command planned
- No `accounts connect` command exists
- CLI is built on the Public API v1 endpoints

### Supported Platforms for OAuth Connection
From `routes/web/integrations.php` callback routes: Facebook, Instagram, Twitter/X, LinkedIn, GMB, Pinterest, Threads, Tumblr, YouTube, TikTok, Medium, Webflow, Shopify, WhatsApp, Bluesky (non-OAuth)

## What Needs to Change

### 1. Public API v1 — New endpoints
- `POST /api/v1/workspaces/{workspace_id}/accounts/connect` — Generate OAuth URL for a platform, return `{auth_url, connection_id}`
- `GET /api/v1/workspaces/{workspace_id}/accounts/connect/{connection_id}/status` — Poll connection status (pending → completed/failed)

The API wraps the existing `getAuthorizationUrl` + `checkConnectorState` logic — no new OAuth infrastructure needed.

### 2. MCP Server — New tool
- `connect_social_account` tool that calls the new API endpoint and returns the `auth_url` for the agent/user to open

### 3. CLI — New command
- `contentstudio accounts connect --workspace <id> --platform facebook` — generates OAuth URL, opens browser, polls for completion

## Files Involved

### API endpoint
- `contentstudio-backend/routes/api/v1.php` — add new routes
- `contentstudio-backend/app/Http/Controllers/Api/V1/AccountController.php` — add `connect()` and `connectStatus()` methods
- New Form Request class for validation

### MCP tool
- `contentstudio-backend/app/Mcp/Tools/ConnectSocialAccountTool.php` — new MCP tool
- `contentstudio-backend/config/mcp.php` — register new tool

### CLI command (in CLI package, not backend)
- New `accounts connect` command in the CLI package
