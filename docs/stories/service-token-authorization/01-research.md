# 01 Research — Service token authorization for browser-reachable services

Date: 2026-08-23
Scope: how `contentstudio-social-analytics-go` and `social-inbox-manager` authenticate and authorize requests that come straight from the browser, and what stops one user's token from reading another user's data.

This doc is the local grounding for the epic. Nothing here goes into the story body.

---

## 1. The shape of the problem

The frontend talks to three separate HTTP surfaces, not one:

| Surface | Frontend env var | Who authenticates |
|---|---|---|
| Laravel API | `VITE_CS_BACKEND_URL` / `VITE_BASE_URL` | Laravel auth + `PermissionMiddleware` |
| Go analytics API | `VITE_ANALYTICS_GO_URL` | own JWT middleware |
| Inbox service API | `VITE_INBOX_API_URL` (`app-inbox-api.contentstudio.io`) | own JWT dependency |

The Laravel app mints a JWT. The browser then sends that same JWT to the Go service and to the inbox service. Each of those services validates the signature itself. Neither of them calls back to Laravel to ask "is this user allowed to do this?".

That is the gap. **A valid signature proves the token was issued by us. It does not prove the caller may read the workspace they asked for.** Because `workspace_id` travels in the request body or query string, a caller with any valid token can change it.

---

## 2. Go analytics API: signature only, no ownership check

`contentstudio-social-analytics-go/src/api/middleware/auth.go` and `jwt.go`:

- `AuthMiddleware.Authenticate` accepts, in priority order:
  1. `X-Shareable-ID` header on `/analytics/*` paths. If the link id resolves to an active user, the request is authenticated **as that user and JWT validation is skipped entirely**. `TestAuthMiddleware_ShareableOverridesInvalidJWT` confirms a valid shareable id wins even when the Bearer token is invalid.
  2. `Authorization: Bearer` — HMAC-SHA256, verified against `JWTConfig.Secret`, then optionally against `AdminSecret` and `InternalAdminSecret`. A token that verifies against either admin secret is marked `isAdmin` and **skips the issuer check**.
  3. `X-API-Key` — looked up in the `api_keys` Mongo collection.
- On success the middleware stores `&JWTClaims{RegisteredClaims{Subject: userID}}` in the request context. `JWTClaims` carries nothing but the standard registered claims: no workspace list, no role, no scope.
- Downstream: `grep` for `ClaimsContextKey` / `Subject` across `src/api` returns hits in exactly **one** handler, `src/api/analytics/looker_studio/handler.go`. Every other handler ignores the authenticated identity.
- Meanwhile `workspace_id` is read from the request in `src/api/analytics/facebook/handler.go`, `src/api/analytics/accounts/handler.go`, `src/api/analytics/tiktok/handler.go`, `src/api/analytics/meta_ads/handler.go`, `src/api/listening/mentions_handler.go`, `views_handler.go`, `credits_handler.go`, `filter_resolver.go`, `src/api/handle_listening_work.go`, `src/api/immediate_work_apis.go`.

Net effect: any authenticated caller can pass any `workspace_id` and the service will serve that workspace's analytics. Listening credits handlers are in the same position, which makes it a billing surface too, not only a data surface.

## 3. Inbox service: the check exists, but it is behind a flag and only partly wired

`social-inbox-manager` has already done the first pass of this work. `app/api/permissions.py` documents the exact problem in its own module docstring, and calls it an IDOR:

> Every `/inbox/*` endpoint authenticates the caller (`require_auth`) but, historically, took `workspace_id` straight from the request body and trusted it, with no check that the user actually belongs to that workspace. Because SIM is internet-facing [...] that is an IDOR.

What exists:

- `WorkspaceMembership.is_member()` queries the shared `workspace_team` Mongo collection, the same collection Laravel's `PermissionHelper::isValidWorkspaceMember()` uses. Fails closed on missing ids or a datastore error.
- `WorkspaceMembership.can_access()` = membership **or** the globally readable sample workspace, mirroring Laravel's virtual-membership behaviour for sample workspaces. Reads only; mutations are separately rejected for sample workspaces via `ensure_workspace_writable` (`/v1`) or `_guard_workspace_mutation` (legacy).
- `app/api/auth.py::_enforce_workspace_membership()` is called from `require_auth`, the single chokepoint for `/inbox/*`.
- `app/api/v1/dependencies.py` re-expresses the same logic as first-class FastAPI dependencies for the `/v1` API.

What is still open:

1. **It is off by default.** `_enforce_workspace_membership` returns immediately unless `settings.ENFORCE_WORKSPACE_MEMBERSHIP` is truthy, read dynamically so it can be toggled. The rollout state per environment needs confirming.
2. **Requests with no `workspace_id` are not checked.** `_extract_workspace_id` is a best-effort read of the body or query string. If it finds nothing, the code logs a warning and lets the request through. The docstring notes these endpoints are tracked separately.
3. **`workspace_id` still comes from the caller.** Membership is verified, but the request still names its own tenant.
4. **Three signing secrets are accepted.** `verify_token` tries `JWT_SECRET_KEY`, then `IMPERSONATION_JWT_SECRET`, then `PUBLIC_API_JWT_SECRET` in turn, and returns the first that decodes. A token minted for one purpose is indistinguishable from another downstream, and the impersonation secret in particular deserves a narrower blast radius.
5. **Role and permission are not checked at all**, only membership. Laravel distinguishes roles per workspace, for example approver-only access. The inbox service does not.

## 4. What Laravel does that the other two services do not

- `app/Http/Middleware/PermissionMiddleware.php`, `SubscriptionMiddleware.php`, `ShareableAwareAuthMiddleware.php`, `ShareableOptionalMiddleware.php`, `InternalApiMiddleware.php`, `DemoWorkspaceAccessModeMiddleware.php`, `MergeWorkspaceIdFromRoute.php`.
- `PermissionHelper::isValidWorkspaceMember()` plus per-role permission checks and the sample-workspace virtual membership.

So the product already has an authorization model. It simply does not reach the two services the browser also talks to.

## 5. Token issuance and lifetime

- Frontend sends the same JWT to all three surfaces, so the token's blast radius is every service, not one.
- The Go service accepts two admin secrets (`AdminSecret`, `InternalAdminSecret`) which bypass the issuer check. Any leak of an admin-signed token is effectively unrestricted.
- No evidence of token revocation, rotation, or per-audience (`aud`) scoping across the three surfaces. Existing related work: `realtime-token-expiry-detection`, `sign-out-other-devices-on-password-reset`.
- Shareable analytics links are a deliberate anonymous path. Any design has to keep them working: the shareable branch must stay a distinct, read-only, link-scoped identity rather than a hole in the membership check.

## 6. Design options the research ticket should compare

1. **Membership check in each service against the shared `workspace_team` collection.** What the inbox service already did. Cheapest, no new service call, but duplicates the authorization rules in every language and drifts from Laravel's role model.
2. **Scoped tokens.** Laravel mints a short-lived token whose claims carry the audience and the workspaces or accounts the bearer may read. Services verify claims and stop trusting the body. Removes the cross-service blast radius, but needs token minting, refresh, and a story for shareable links and the public API.
3. **Central authorization call.** Services ask Laravel per request. Single source of truth, worst latency and a hard availability coupling. Needs caching.
4. **Gateway in front of both services.** Terminates auth once and injects a verified identity plus workspace scope. Combines well with moving internal traffic off public hostnames, which the internal-DNS epic covers.

Whichever is chosen, the plan must cover: role and permission checks not just membership, endpoints that carry no `workspace_id`, per-account scoping inside a workspace (a member may see the workspace but not every connected account), the three-secret decode path, admin-signed tokens, shareable links, public API keys, and the rollout order with the enforcement flag.

## 7. Adjacent existing deliverables (do not duplicate)

- `inbox-stability-hardening`, `inbox-module-architecture-sync-findings`
- `analytics-api-consistency`, `analytics-php-to-golang-migration`
- `realtime-token-expiry-detection`, `sign-out-other-devices-on-password-reset`
- `public-analytics-api-rollout`, `api-legacy-endpoint-validations`
