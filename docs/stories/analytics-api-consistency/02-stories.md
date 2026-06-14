# Analytics API Consistency — Story Titles + Generic Template

Standardizing the **payload/response structure, naming, and conventions** of the analytics APIs so every social platform follows one consistent shape. One story per platform. The body is generic — copy-paste as-is; it refers to "this platform" (whatever the title names).

**Impacted surfaces (per platform):** Backend · Frontend · Data Studio · Public API.

> Written as backend-led API-contract stories (the consistent structure is backend-owned), with the consuming surfaces covered as cross-impact + verification. If a team prefers, any platform can be split per surface into `[BE]` / `[FE]` / `[Data Studio]` / `[Public API]` stories.
>
> Suggested Shortcut fields: **Project:** Web App · **Group:** Backend · **Skill set:** Backend · **Product area:** Analytics · **Type:** Chore.

---

## Titles

1. `[BE] Standardize analytics API payload & response structure for Facebook`
2. `[BE] Standardize analytics API payload & response structure for Instagram`
3. `[BE] Standardize analytics API payload & response structure for LinkedIn`
4. `[BE] Standardize analytics API payload & response structure for X (Twitter)`
5. `[BE] Standardize analytics API payload & response structure for YouTube`
6. `[BE] Standardize analytics API payload & response structure for TikTok`
7. `[BE] Standardize analytics API payload & response structure for Google Business Profile (GMB)`
8. `[BE] Standardize analytics API payload & response structure for Pinterest`

---

## Generic story body

> Copy-paste as-is. The body refers to "this platform" — i.e. whatever the story title names.

### Description
As a consumer of ContentStudio's analytics (our own frontend, Data Studio, and the Public API), I want this platform's analytics API to follow the same payload/response structure, field naming, and conventions as every other platform, so that analytics data is predictable to build on — with fewer per-platform special cases, less duplicated handling, and fewer bugs.

### Workflow
*(API-contract change; the "user" here is a developer consuming the analytics API.)*
1. A developer calls this platform's analytics endpoints.
2. Request parameters use the same naming/shape as the other platforms' analytics endpoints.
3. The response uses the same overall structure, field names, types, and metric/date conventions as the other platforms.
4. Our frontend, Data Studio, and the Public API all read this platform's analytics through the shared, consistent shape — no platform-specific branches.

### Acceptance criteria
- [ ] This platform's analytics request payloads and responses follow the standardized structure and naming used across all platforms (same field names, nesting, types, date and metric conventions).
- [ ] The same metric/concept uses the same name as on other platforms (no platform-specific aliases for identical concepts).
- [ ] **No metric values change** — only the structure and naming are standardized; values match what the platform reported before.
- [ ] **Frontend:** analytics views for this platform render correctly against the standardized structure with no regressions.
- [ ] **Data Studio:** reports/widgets for this platform consume the standardized structure correctly.
- [ ] **Public API:** this platform's analytics are exposed in the standardized structure, with versioning/backward-compatibility handled so existing external integrations don't break (any change is versioned and deprecation is communicated).
- [ ] Documentation (internal + Public API docs) reflects the standardized structure for this platform.
- [ ] Edge cases are consistent with other platforms (empty data, missing metrics, partial periods, rate-limit/error responses).

### Mock-ups
N/A — API structure/contract change.

### Impact on existing data
Request/response **structure and naming** change; underlying stored analytics data is unchanged. Likely needs a mapping/transform layer to the standardized shape — verify metric values are identical before/after.

### Impact on other products
- **Backend** — owns the standardized structure (this change).
- **Frontend** — analytics views must consume the standardized shape.
- **Data Studio** — reports must consume the standardized shape.
- **Public API** — external consumers; requires versioning/backward-compat so integrations don't break.
- *Mobile apps:* if they consume these analytics endpoints, confirm whether they're impacted and align them too.

### Dependencies
A single **canonical/standard analytics schema** (the shape all platforms conform to) should be agreed first — the first platform standardized typically establishes it; the rest conform to it.

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*
- `contentstudio-social-analytics-go/` — the analytics data pipeline (Scheduler → Fetcher → Parser → Processor → Sink); where per-platform analytics are produced/normalized.
- `contentstudio-backend/` — analytics controllers/services that serve the analytics APIs to the app and Public API.
- `contentstudio-frontend/src/modules/analytics_v3/` — the active analytics module that consumes these responses.
- Define the canonical schema once and conform each platform to it; keep a transform/adapter layer so stored data and values are untouched.
