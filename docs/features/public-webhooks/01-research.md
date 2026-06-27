# Research — Public / Outbound Webhooks

**Feature:** ContentStudio POSTs a signed JSON event to a customer-registered URL when publishing-lifecycle events occur (post published / failed / partial / scheduled / created). A developer/API feature analogous to the existing public API.

**Primary reference (cite for devs):** Zernio webhooks — https://docs.zernio.com/webhooks

> Note: the broad competitor web-sweep was interrupted; Part A is grounded on the Zernio reference the PO supplied + well-established developer-platform webhook standards + general product knowledge. Competitor specifics are best-effort and can be deep-dived on request. The codebase analysis (Part B) is fully verified against the repo.

---

## Part A — Competitor & Industry Research

### What is this feature?
An **outbound webhook** lets a customer register a URL; ContentStudio then sends an HTTP POST to that URL whenever a subscribed event happens. It's the **push** counterpart to the **pull** public API: instead of polling `GET /posts` to see if a scheduled post went out, the customer's system is notified the instant it does. Webhooks are a standard expectation for any product with a public API and a key enabler for automations, internal dashboards, and integrations (Zapier-style or bespoke).

### Competitor / landscape analysis
Outbound webhooks are a **developer/API-first** feature. The split is clear: API-centric social-posting platforms offer them; end-user-UI-first tools generally don't expose them for self-serve customers.

| Product | Outbound webhooks? | Events | Auth/Signing | Tier | Management UX | Notes |
|---|---|---|---|---|---|---|
| **Zernio** (the reference) | **Yes** (model to emulate) | Posts, accounts, inbox, comments, reviews, ads, WhatsApp | HMAC-SHA256, **optional** per-endpoint secret, `X-Zernio-Signature` | With API access | Dashboard page: create webhook (name, URL, secret, custom headers, event checkboxes), delivery logs, test event | At-least-once, exp-backoff retries (7 attempts → dead-letter), event-id dedup |
| **Ayrshare** | Yes | Post status, comments, messages | HMAC signing | Business/API plans | Dashboard + API | Developer-first social API |
| **Postiz** | Partial / emerging | Limited | — | API | — | Open-source, API-focused (flag: verify) |
| **Buffer** | No (general) | — | — | Legacy API | — | Public API largely deprecated for new devs |
| **Hootsuite** | Enterprise-only | Limited | — | Enterprise | — | Not a self-serve webhook product |
| **Publer / Later / Sendible / SocialBee / Agorapulse / Metricool** | Mostly no (self-serve) | — | — | — | — | End-user UI tools; outbound webhooks not a core self-serve offering |

**Key insight:** offering webhooks positions ContentStudio with the developer-platform segment it already targets (it has API-centric plans and an API landing page). Zernio — a direct analog (social posting API) — is the right model.

### Developer-platform gold standards (Stripe / GitHub / Shopify / Twilio / Svix) — consensus patterns
- **Event taxonomy:** `resource.action` naming, past tense (`post.published`). Customers subscribe only to events they handle.
- **Payload envelope:** stable **event id** (UUID, the dedup key) + event type + timestamp + a data object. Same id echoed in a header.
- **Security:** HMAC-SHA256 over the **raw request body**, keyed by a per-endpoint secret, sent in a signature header. Customer recomputes and compares. Reject on mismatch.
- **Delivery semantics:** **at-least-once**; consumers must be **idempotent** (dedupe on event id). Expect a fast `2xx` ack (~5s timeout), push heavy work to the consumer's own async queue.
- **Retries:** exponential backoff, capped, a handful of attempts (Zernio: 7, cap 24h), then **dead-letter**. Failures visible in logs. **Not auto-disabled** on failure — user pauses/removes manually.
- **Management UX:** register endpoint → pick events → reveal secret once → view delivery logs → send a **test event**.

### Billing norm (directly answers the PO's question)
**Webhooks are free and unmetered across the board.** Stripe, GitHub, Shopify, Twilio, Ayrshare, and Zernio do **not** charge per delivery or count deliveries against API rate limits. They're a retention/integration feature, and they *reduce* API load by replacing polling. Industry consensus strongly supports the locked decision: **free, no API-credit deduction.** Abuse is controlled structurally (plan-gating + an endpoint cap), not by metering.

### User expectations — table stakes vs delighters
- **Table stakes:** HMAC signing, retries + dead-letter, event-id idempotency, a management page, delivery logs, a test-event button, an optional secret.
- **Delighters:** custom headers per endpoint (Zernio has this), per-event subscription granularity, a clear payload doc, "resend delivery" from logs.

### Recommended approach for ContentStudio v1
Mirror Zernio, scoped to **publishing-lifecycle events only**: `post.published`, `post.failed`, `post.partial`, `post.scheduled`, `post.created`. Free, gated by the same plan flag that grants API access, with a small per-workspace endpoint cap. Reuse existing infra (Horizon delivery jobs, the API module UI shell). Inbox/comments/reviews/account events are explicitly a **future phase**.

---

## Part B — Codebase Analysis (verified)

### Existing related code
- **Public API:** `contentstudio-backend/routes/api/v1.php` under `['api.key', 'api.request.log', 'throttle:api-v1']`. Auth via `app/Http/Middleware/ApiKeyMiddleware.php` (`X-API-Key`, `cs_`-prefixed, SHA-256 hashed, `api_keys` collection). Management: `ApiKeyController` (`/api/api-keys`).
- **API access gating (webhooks mirror this):** plan flag **`features.api_access`** (boolean) on the `subscription_plans` collection (added in `database/migrations/2025_01_23_120000_add_api_key_feature_to_plans.php`). Frontend gate: `useFeatures().canAccess(...)`.
- **API credit model (NOT used by webhooks — for contrast):** `ApiKeyMiddleware` increments per-workspace `used_api_credit`; allowance from `SubscriptionLimits` (`api_credit` + addons). Webhooks bypass this entirely.
- **Post-lifecycle hooks to fire from:**
  - `app/Jobs/PlanFinalizerJob.php` — writes the final `published` / `failed` / `partially_failed` status → the source for `post.published` / `post.failed` / `post.partial`.
  - `app/Observers/Publish/PlanObserver.php` (`plan_created`) → `post.created`; the scheduled state on the `Plans` model → `post.scheduled`.
  - `Plans` model (`app/Models/Publish/Planner/Plans.php`); statuses in `app/Data/Enums/PostStatus.php`. A plan fans out to multiple social accounts → maps to the payload `platforms[]` array.

### Reusable components / services
- **Delivery engine:** Horizon queued-job pattern with `$tries` / `$backoff` / `failed()` — `app/Jobs/UpdateWebhookForPlatform.php` (uses the `Http` facade for outbound calls), `GenerateReportJob` (exponential `$backoff`). The natural template for a `DeliverWebhookJob`.
- **Event IDs:** `(string) Str::uuid()` (Laravel built-in) is the established pattern across the codebase — use for the webhook event id / dedup key.
- **Kafka (optional buffer):** topics + SASL credentials in `config/kafka.php` (`backend_sasl` for backend topics). A `webhook.delivery.event` topic could decouple event capture from HTTP delivery if needed — not required for v1.
- **Frontend shell:** the API module already exists — `src/modules/setting/components/api/ApiModule.vue` (+ `ApiOverviewHeader.vue`), mounted at the top-level route `/:workspace/api` (name `api`), reached from the **desktop rail "API" entry**. `ApiRequestLogs.vue` is a near-identical template for the delivery-logs view. `useApiKeys` / `useApiRequestLogs` composables + `settingKeys` query factory show the data-fetching pattern.

### Integration points
- **Backend:** new `webhook_endpoints` (customer URLs, events, secret, custom headers, active/paused) + `webhook_deliveries` (attempts, status, response) collections; a dispatcher that hooks into `PlanFinalizerJob` + `PlanObserver`; a `DeliverWebhookJob`; management endpoints backing the UI (CRUD + regenerate secret + test + list deliveries) shaped like `ApiKeyController`.
- **Frontend:** a new **Webhooks** section/route **inside the API module** (`src/modules/setting/components/api/`), reached from the desktop rail API area — **not** Settings. Empty state, create/edit form, delivery logs, test event.
- **Availability:** gate the Webhooks UI + endpoint creation with the same `features.api_access` flag.
- **Docs:** add `contentstudio-backend/docs/api/webhooks-endpoint.md` (+ quick-reference) following `posts-endpoint.md`, with `@OA\*` annotations and L5-Swagger regeneration. Cite https://docs.zernio.com/webhooks.

### Technical considerations
- **No existing outbound-webhook scaffolding** — build fresh. Do **not** reuse the inbound social/billing webhook handlers (`FacebookWebhooksController`, `PaddleBillingController`, etc.) — they have inverse security assumptions (receiving, not sending).
- **Signing:** HMAC-SHA256 over the raw body, per-endpoint secret (optional, per Zernio), header `X-ContentStudio-Signature`; also send `X-ContentStudio-Event-Id`.
- **Reliability:** at-least-once, exponential backoff → dead-letter, idempotency via event id, fast-ack expectation. Don't auto-disable endpoints on failure.
- **Payload includes post content** (locked decision) — flag a guard for very large posts (truncate/omit?) as an open question.

### Mobile impact
**None.** This is a developer/API, web-only feature. No iOS/Android stories.

---

## Locked decisions (carried from PO)
1. **Free** — no API-credit deduction. Protect via `features.api_access` gating + a per-workspace endpoint cap.
2. **v1 = publishing-lifecycle events only:** `post.published`, `post.failed`, `post.partial`, `post.scheduled`, `post.created`. Inbox/comments/reviews/account = future phase.
3. **Zernio-style payload, including post content.** Envelope: `{ id, event, timestamp, workspace_id, post: { id, status, scheduledFor, publishedAt, content, platforms[]: { platform, status, platformPostId, publishedUrl, error } } }`. Headers: `X-ContentStudio-Signature` (HMAC-SHA256, raw body, per-endpoint secret) + `X-ContentStudio-Event-Id`.
4. **Available to anyone with API access** (`features.api_access`). No per-delivery credit.

## Open questions (resolve during PRD/workflow)
- Exact retry schedule + max attempts + dead-letter behavior (Zernio: 7 attempts, exp backoff capped at 24h — adopt as default?).
- Per-workspace endpoint cap value (e.g. 5–10).
- Expose webhook management via the public `/api/v1` too (Zernio does), or in-app management only for v1?
- Payload `content` for very large posts — truncate/omit, or always full?
- Adopt Zernio's **custom headers per endpoint** and a **Name** field in v1? (Both are in Zernio's create form and cheap to include.)
