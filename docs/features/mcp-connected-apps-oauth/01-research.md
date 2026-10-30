# Research: Open the ContentStudio MCP server to any AI app, safely

**Date:** 2026-10-26
**Sources:**
- The MCP specification and security guidance
- Vendor documentation
- A codebase pass on `contentstudio-backend` `origin/master` (1de2b4bd6) and `contentstudio-frontend` `origin/master` (0524f0498d)
- Market context: `docs/features/public-api-gap-analysis/01-research.md`

## 1. What this is and why

**What works today:** our MCP server signs users in with OAuth, but only Claude, ChatGPT and local tools can connect. The redirect allowlist is `claude.ai, claude.com, chatgpt.com, openai.com`, plus loopback.

**What is blocked:** any web-based MCP client with its own domain is refused, including Gemini, Copilot Studio, Perplexity, Notion, Manus, n8n cloud and Make.

**Why the allowlist can't simply be opened:** the token we hand out is the user's single master API key. It never expires, it has full access to every workspace, and revoking it breaks every integration at once. The allowlist is the only thing stopping a fake app from obtaining that key. Opening it safely needs per-app tokens first.

## 2. The standard

### MCP security best practices (2025-11-25)
https://modelcontextprotocol.io/docs/2025-11-25/tutorials/security/security_best_practices.md

**The consent page MUST:**
- identify the client by name
- show the scopes requested
- **show the registered `redirect_uri` where tokens will be sent**
- have CSRF protection
- prevent framing (`frame-ancestors` / `X-Frame-Options: DENY`)

**Redirect URIs:** exact string matching against the registered URI, with no patterns or wildcards.

**Token passthrough is forbidden.** MCP servers MUST NOT accept tokens that were not explicitly issued for them.

**Scope minimization:**
- Start with minimal scopes, then elevate incrementally (a `WWW-Authenticate` scope challenge)
- Wildcard or omnibus scopes are named as a "common mistake"

### MCP specification 2026-07-28
https://modelcontextprotocol.io/specification/2026-07-28/changelog

- **Dynamic Client Registration (RFC 7591) is deprecated** in favour of **Client ID Metadata Documents (CIMD)**. DCR remains available for backwards compatibility.
- Authorization servers SHOULD include `iss` in authorization responses (RFC 9207).
- Clients MUST specify `application_type` in DCR.

### Client ID Metadata Documents (CIMD)
- The client's `client_id` is an HTTPS URL it controls.
- That URL serves a JSON document with the client name and allowed redirect URIs.
- The authorization server fetches the document, so a client can only claim the domain it controls.
- **Main server-side risk:** SSRF when fetching the URL.
- References: https://workos.com/blog/mcp-client-registration-cimd-vs-dcr, https://den.dev/blog/mcp-november-authorization-spec/

## 3. How others do it

| Vendor | Registration | Which clients | Token model | Notes |
|---|---|---|---|---|
| Linear | DCR, open | Any | Per-client OAuth token | No client id to manage (LiteLLM docs) |
| Notion | DCR, open | Any | Per-client OAuth token | Cited as a DCR example |
| Atlassian (Rovo MCP) | DCR, open | Any | OAuth 2.1 user tokens | Open to all Atlassian Cloud customers |
| Sentry | DCR, open | Any | OAuth 2.1 | MCP-level OAuth separate from Sentry's own |
| Buffer | OAuth apps | Registered apps | Scoped (`posts:read`, `posts:write`…), PKCE required, single-use refresh tokens | Reusing a refresh token revokes the grant |
| Hootsuite | OAuth (4 MCP servers) | Claude and ChatGPT connectors, remote URL for others | Follows the user's existing permissions | |

**Common pattern:** no vendor gives an AI app the user's master API key. Each connection gets its own token, and that token can be revoked on its own.

Not verified per vendor: the exact "Connected apps" settings page for each. Treat that as a typical pattern.

## 4. Codebase analysis

### 4.1 OAuth server

**Routes:**
- `routes/oauth.php:17-32` (metadata, register, authorize, token; `throttle:oauth`)
- Consent at `routes/api.php:159-161`
- Controller: `app/Http/Controllers/OAuth/OAuthController.php`

**Models** (`app/Models/OAuth/`):
- `OAuthClient` → `oauth_clients`: client_id, client_name, redirect_uris, token_endpoint_auth_method, last_used_at, expires_at. **It has no user_id.**
- `OAuthRequest` → `oauth_requests`
- `OAuthCode` → `oauth_codes`. Holds user_id and the **plaintext API key**, and is deleted on exchange.

**Registration** (`ClientRegistrationService::register:25`):
- Public clients, no secret
- Redirects are checked by `RedirectUriValidator::isAllowed:15-47`: https on the allowlist, subdomains included, or any loopback

**Authorize** (`AuthorizationServerService::createAuthorizationRequest:30`):
- Redirect is matched exactly against the registered list (`:43`) ✅
- PKCE S256 required (`:48`) ✅
- Client TTL 90 days, sliding

**Token** (`TokenService::exchange:19`):
- `authorization_code` grant only
- **The access token is the user's `cs_` API key** (`:58`). `approve:109-121` creates it if missing.
- Response: `expires_in: 31536000`, advertised only. No refresh token.
- `redirect_uri` and `client_id` are only checked when sent (`:33`, `:38`). They should be required.

**Scopes:**
- `contentstudio:read` and `contentstudio:write`, both granted by default (`config/oauth.php:55-56`)
- Stored and echoed back, but **never enforced**: `ApiKeyMiddleware` ignores them

**Minor:** request and code consumption are read-then-update, not atomic.

### 4.2 API keys
- **Model:** `app/Models/ApiKey.php` (collection `api_keys`; fields key, user_id, revoked, last_used_at, ai_creation)
- **Repo:** `app/Repository/ApiKeyRepo.php`. One non-AI key per user (`getUserApiKey:77`).
  - `revoke:133` and `regenerate:155` disconnect every MCP client at once
  - Keys are stored in plaintext (`findValidByKey:192`)
- **Auth resolution:** `RequestApiKey::extract` (X-API-Key, Bearer, or `?api_key`) → `ApiKeyMiddleware::findApiKey:159`. **That is the place to accept a new connected-app token type.** It is already a test seam.
- **Metering:** a connected-app token should be charged to the same owner credit pool as the API key, so usage stays one number.

### 4.3 Frontend

**Consent page:** `src/modules/integration/components/oauth/OAuthConsent.vue`, route `/oauth/consent` (`src/router.ts:354`).
- Shows the client name, "Signed in as {email}", and the scope list
- Shows the disclaimer "Your API key is provisioned automatically. You can revoke access anytime from your ContentStudio settings." This becomes untrue once tokens are per-app.
- **Does not show the redirect domain**
- Strings: `src/locales/en/integration.json:2-27`. API: `src/api/oauth-consent.ts`.

**API settings:** route `/:workspace/api` (`src/router.ts:620`) → `src/modules/setting/components/api/ApiModule.vue`.
- Top tabs: `api` / `webhooks` (`:44-46`)
- Inner tabs: `api_key` / `request_logs` (`:54-55`)
- Uses `ApiKeyCredentials.vue`, `ApiRequestLogs.vue`, `CliAgentSection.vue`, `useApiKeys.ts`
- **"Connected apps" fits as a third top-level tab beside API and Webhooks** (the inner tabs are API Docs and Request Logs, which are about the API key; PO screenshot 2026-10-26)

**No record of connected clients exists:**
- `OAuthClient` has no user link
- `OAuthCode` is deleted after exchange
- A grants record is needed: user, client id, client name, redirect host, scopes, created, last used, revoked

### 4.4 MCP server
- **Hosted connector:** `claude.ai ContentStudio`. Repo `d4interactive/contentstudio-mcp`, not mounted.
- **Accepts:** API key via env var (`CONTENTSTUDIO_API_KEY`) for the npm package, or the OAuth token for the hosted connector.
- **Both must keep working:**
  - API-key users (CLI, npm MCP, Zapier, Make, n8n) are unaffected
  - OAuth users get per-app tokens

## 5. Recommended approach (agreed in principle with the PO, 2026-10-25)
1. **Per-app tokens.** Issue a token per approved connection, separate from the API key:
   - short-lived access token (hours) plus a rotating refresh token
   - stored hashed
   - recorded as a grant: user, app name, redirect host, scopes, created, last used
   - accepted by the API alongside API keys and charged to the same credits
2. **Enforce scopes.** A read-only grant cannot call write endpoints.
3. **Consent screen:**
   - show the destination domain
   - choose read-only or read-and-write
   - label unknown apps as unverified
   - stop saying the API key is provisioned
4. **Connected apps page** in Settings → API:
   - list grants with name, domain, access level, connected date and last used
   - revoke one app without touching the API key
5. **CIMD support** alongside DCR, with an SSRF-safe metadata fetch.
6. **Open the redirect allowlist** to any HTTPS host. This lands last, once steps 1-4 are live.
7. **Existing connections:** Claude and ChatGPT users currently hold API-key tokens.
   - Keep accepting them until they reconnect, or for a set grace period
   - The PO decides whether to force a reconnect
