# Story: Harden validation on legacy API endpoints

## [BE] Audit and add request validation to legacy API endpoints

### Description

The newer public API v1 endpoints validate their request payloads with dedicated FormRequests and return structured errors. Several older API endpoints predate that and either validate loosely or inconsistently. Audit the legacy endpoints and add consistent request validation with clear, structured error responses, so bad input is rejected predictably instead of causing unexpected behavior.

> This may be split into per-area stories after the audit if the surface is large.

### Acceptance criteria

- [ ] The legacy endpoints in scope validate required fields, types, and allowed values before processing.
- [ ] Invalid requests return a consistent, structured validation error with field-level messages and the correct HTTP status.
- [ ] Validation rules match the actual behavior the endpoint supports (no rejecting previously valid requests, no accepting clearly invalid ones).
- [ ] Error response shape is consistent with the pattern used by the v1 endpoints.
- [ ] Existing valid client calls continue to work unchanged (no breaking changes for well-formed requests).

### Impact on existing data

None. Validation only. Watch for over-tightening that could reject requests real clients currently send.

### Impact on other products

Any client of these endpoints (web app, mobile apps, Chrome extension, integrations) is affected if it sends malformed requests today. Confirm no well-formed client call breaks.

### Dependencies

An audit of the legacy endpoints is a prerequisite to finalize the exact list and rules.

### Global quality and compliance checklist

- [ ] Mobile responsiveness (N/A, backend-only story)
- [ ] Multilingual support verified (validation messages localized or fallback handled)
- [ ] UI theming supported (N/A, backend-only story)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

### Implementation references

*Pointers from research, not a contract. Engineering may choose a different approach.*

- Public API v1 validation pattern to mirror: FormRequests under `contentstudio-backend/app/Http/Requests/Api/V1/` (for example `PostStoreRequest.php`) and the structured error responses used by the v1 controllers.
- Legacy endpoints live under `contentstudio-backend/routes/api.php` (and related controllers). Start with an audit of which of these lack a FormRequest or robust inline validation.
- **Open item:** confirm the exact list of "old endpoints" in scope so rules can be written precisely and the story split if needed.
