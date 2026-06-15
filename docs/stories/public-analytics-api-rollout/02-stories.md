# Public Analytics API Rollout — Story Titles + Generic Template

Exposing **analytics in the public-facing API** (customers call it with their own API keys), one platform at a time. **Facebook is already done** — these stories add the rest, mirroring the Facebook implementation. One story per platform; the body is generic — copy-paste as-is, it refers to "this platform" (whatever the title names).

> Suggested Shortcut fields: **Project:** Web App · **Group:** Backend · **Skill set:** Backend · **Product area:** Analytics · **Type:** Feature.
>
> Pairs with the analytics API consistency effort — the public endpoints should expose the **standardized** structure.

---

## Titles

1. `[BE] Add public Analytics API for Instagram`
2. `[BE] Add public Analytics API for LinkedIn`
3. `[BE] Add public Analytics API for X (Twitter)`
4. `[BE] Add public Analytics API for YouTube`
5. `[BE] Add public Analytics API for TikTok`
6. `[BE] Add public Analytics API for Google Business Profile (GMB)`
7. `[BE] Add public Analytics API for Pinterest`

*(Facebook already shipped — used as the reference implementation.)*

---

## Generic story body

> Copy-paste as-is. The body refers to "this platform" — i.e. whatever the story title names. It mirrors the already-shipped Facebook public analytics API.

### Description
As a ContentStudio customer using our Public API with my API key, I want to pull this platform's analytics programmatically — the same way Facebook analytics are already available — so that I can build my own reports, dashboards, and integrations on ContentStudio's analytics data.

### Workflow
*(Public API change; the "user" here is a developer calling the API with their API key.)*
1. A developer authenticates to the Public API with their API key.
2. They call this platform's analytics endpoints under the workspace analytics path, mirroring the existing Facebook endpoints.
3. They pass the same parameters Facebook uses (workspace, account id, date range, timezone).
4. They receive analytics in the same standardized response structure as Facebook.
5. The endpoints are discoverable in the Public API reference docs.

### Acceptance criteria
- [ ] This platform's analytics are exposed via the Public API under the workspace analytics path, mirroring the existing Facebook analytics endpoints (the same set of sections Facebook exposes — e.g. summary, audience growth, etc. — adapted to this platform's available metrics).
- [ ] Endpoints authenticate with the user's **API key**, using the same auth mechanism/header as the Facebook analytics endpoints.
- [ ] Request parameters follow the same conventions as Facebook (workspace id in the path, account id, `start_date`/`end_date` or `date` range, `timezone`).
- [ ] Responses follow the **standardized analytics structure and naming** (consistent with the analytics-consistency effort), matching the Facebook response shape.
- [ ] Standard error responses match Facebook: **401** unauthorized, **403** workspace permission denied, **422** validation error, **429** rate limited, **502** upstream analytics service error.
- [ ] Rate limiting is applied consistently with the other Public API analytics endpoints.
- [ ] Endpoints are documented in the Public API reference (OpenAPI/Swagger) with parameters, examples, and response schemas, like Facebook.
- [ ] Only metrics this platform actually supports are included; any differences from Facebook (metrics not available, platform-specific sections) are noted in the docs.
- [ ] Values returned by the Public API match what ContentStudio's own analytics show for this platform (no discrepancies).

### Mock-ups
N/A — Public API endpoints (documented via OpenAPI/Swagger).

### Impact on existing data
None — read-only API over existing analytics data. Proxies to the analytics pipeline like the Facebook endpoints.

### Impact on other products
- **Public API** — new endpoints for this platform (external consumers).
- Backend proxies to the analytics pipeline microservice.
- Frontend and Data Studio are unaffected (they use internal analytics APIs).
- API documentation is updated.

### Dependencies
- The **standardized analytics structure** for this platform (so the Public API exposes the consistent shape) — see the analytics API consistency work.
- The already-shipped **Facebook public Analytics API** is the reference implementation to mirror.

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*
- `contentstudio-backend/app/Http/Controllers/Api/V1/Analytics/FacebookAnalyticsController.php` — the reference controller to mirror: per-section endpoints (`summary`, `audience-growth`, …), `@OA` Swagger annotations, `ApiKeyHeader` security, and the 401/403/422/429/502 responses.
- `contentstudio-backend/routes/api/v1.php` — the Facebook route group (`workspaces/{workspace_id}/analytics/facebook`, name prefix `api.v1.workspace.analytics.facebook.`); add an equivalent group for this platform.
- `contentstudio-backend/app/Http/Controllers/Api/V1/Analytics/AnalyticsApiSchemas.php` — shared OpenAPI schemas (`AnalyticsUnauthorized`, `AnalyticsForbidden`, `AnalyticsValidationError`, `AnalyticsRateLimited`, `AnalyticsUpstreamError`); reuse them.
- `contentstudio-backend/app/Http/Controllers/Api/V1/BaseApiController.php` — base controller for API-key endpoints.
- Request DTOs like `FacebookSummaryRequest` — create platform equivalents for validation.
- `contentstudio-backend/app/Http/Controllers/Api/V1/SwaggerController.php` — make sure the new endpoints surface in the generated API docs.
- The endpoints proxy to the analytics pipeline microservice (`contentstudio-social-analytics-go`), same as Facebook.
