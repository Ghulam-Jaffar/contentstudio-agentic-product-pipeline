# Research: Public API: safe retries, account events, large uploads and usage history

**Date:** 2026-10-26
**Source analysis:** `docs/features/public-api-gap-analysis/01-research.md` (market-wide comparison, 18 vendors) plus a targeted codebase pass on `contentstudio-backend` `origin/master` (1de2b4bd6) and `contentstudio-frontend` `origin/master` (0524f0498d).

**Scope agreed with the PO (2026-10-25):**
1. Idempotency key on post creation and AI generation
2. Account connected, disconnected and needs-reconnect webhook events
3. Direct-to-storage and resumable large media upload on the public API
4. API usage history on the public API
5. Expose all of the above on every developer surface

**Out of scope (PO):**
- Webhook delivery metering changes
- Preflight validation endpoint
- Per-account publish results and the `post.partial` event
- Usage alerts. These will come later by email and mobile push.
- X wallet endpoints (parked)
- Multiple and scoped API keys

**Already in flight, not re-ticketed:**
- CONT-3684: webhook management on the public API
- CONT-3687: webhooks and limits on every surface
- CONT-4532: alt text in the Publishing API
- CONT-4619: request logs expose full bodies. This is a security bug and blocks usage history.

---

## 1. What this is and why

API-first competitors make their APIs safe to call from automations and agents. ContentStudio's API is broader than any incumbent's, but it is less safe to retry and less observable:

- **Retries can double-publish.** Zapier, Make, n8n and AI agents retry on timeouts by design. `POST /posts` has no idempotency, so a lost response followed by a retry creates a second post.
- **Dead accounts are discovered by failure.** No webhook tells an integrator that an account disconnected or needs reconnecting. They find out when a post fails.
- **Large videos go through our PHP servers.** The only public upload options are multipart through PHP (2 GB cap from `docker/php.ini`) or a URL we download ourselves.
- **Usage is a point-in-time number.** `GET /limits` shows used/limit. There is no history, so an integrator cannot see which calls consume their credits or when.

## 2. Competitor findings (relevant subset)

| Capability | Zernio | Ayrshare | Upload-Post | bundle.social | Outstand | Buffer |
|---|---|---|---|---|---|---|
| `Idempotency-Key` on create | Yes, 24h replay | Yes (`idempotencyKey`) | Yes | Unverified | Yes, plus `Idempotency-Replayed` header | Unverified |
| Account lifecycle webhooks | `account.connected`, `account.disconnected` | `social` (link, unlink, refresh) | `social_account_connected`, `_disconnected`, `_reauth_required` | `social-account.created`, `.updated`, `.deleted` | `account.connected` | No webhooks |
| Large or resumable upload | Presigned URL | Large-file upload | Async upload | S3-style multipart | Signed URL plus confirm | Unverified |
| Usage readable via API | Billing snapshot plus metering | Partial | `GET /me` returns plan plus usage | Usage API (posts, uploads, daily limits) | `get_account_usage` (MCP) | `RateLimit` headers plus usage UI |

**Common patterns:**
- **Idempotency:** a caller-supplied header with a 24-hour replay window. A conflict (same key, different body) returns 409.
- **Account events:** connected, disconnected, and needs reconnect are three distinct events. Each payload carries platform, account id, name and workspace.
- **Uploads:** a two- or three-step flow. Ask for an upload URL, upload directly to object storage, then confirm.
- **Usage:** a read endpoint, sometimes with a time series. Nobody sells upgrades through the API.

## 3. Codebase analysis

### 3.1 Idempotency

**Existing reference implementation:** the inbox, in social-inbox-manager, built by Hassan Saif.
- 2026-07-14 (`283f6a1`): v1 inbox API
- 2026-07-27 (`beabbde`): real idempotency semantics
- 2026-07-28 (`44fc80669`): backend gateway forwards the header

**How it works** (`social-inbox-manager/app/api/v1/services/idempotency.py`):
- Mongo collection `idempotency_keys`, with a unique index on `(key, scope)` and a 24h TTL
- The scope includes the API key id
- A fingerprint of the request; a mismatch returns 409
- State machine: `reserved` → `side_effect_started` → `completed` | `unknown`, with a 120s lease
- Fails closed when the store is unavailable (503), and fails open if writing the result fails after the side effect has happened
- The backend forwards the header verbatim: `BaseApiController::inboxForwardHeaders` (`app/Http/Controllers/Api/V1/BaseApiController.php:547-553`)

**There is no PHP idempotency store.** The only grep hits for "idempot" are comments and an unrelated analytics key.

**Post creation:**
- Route `routes/api/v1.php:320` → `PostController::store:908` → `storePost:913` → `SocialPostingController::processSocialShare:66`
- `publish_type=now` is queued, not synchronous: `dispatchNowJob:1275` → `PlanPostingJob`
- The slow part of the request is synchronous media ingestion (URL to GCS) in `PostMediaService`, wrapped by `withMediaRollback:2566`. A long request makes client timeouts, and so retries, likely.
- Gotcha: the OpenAPI enum at `PostController:571` omits `now`.

**AI generation retries pay twice.** None of the three endpoints deducts before the work is done, and none dedupes a client retry.
- **Images:** `AiImageSyncService::generate:131`. Synchronous. Checks credits up front, deducts after success (`:275`, last-writer-wins on the counter).
- **Videos:** `AiVideoSubmissionService::generate:41`. Async job, charged later via Kafka.
- **Carousels:** `AiCarouselService::generate:99`. Charged with `CarouselCreditUsage::chargeApi($rows, 'api-'.Str::uuid())`. The uuid is random per call, so it does not dedupe.

### 3.2 Account lifecycle events

**Connect (app):**
- `IntegrationController::connectAccounts:444` → `IntegrationBuilder::processIntegration:163` → per-platform connectors
- The user picks accounts → `POST /savePlatforms` → `Platforms::savePlatform:122` → per-platform repos
- Shared repo: `SocialRepo::savePlatform:313`. New accounts go through `addPlatform:537`; reconnect is the update branch at `:420` / `:491`.
- Separate repos: Facebook, LinkedIn, Twitter, Pinterest, Instagram and GMB each have their own.

**Connect (public API):**
- `ConnectController::connect:213` creates a `ConnectSession`
- The hosted picker calls `/api-connect/save-platforms` → `ApiConnectController::savePlatforms:156` → the same `Platforms::savePlatform`
- `ConnectSessionService::markCompleted:108`

**There is no single domain event on connect.** The closest common choke point is the `SocialIntegrations::boot()` model hook (`app/Models/Integrations/Platforms/SocialIntegrations.php:246-266`), which pushes to the Redis list `integration_observers` → `PlatformObserversJob`.
- Gotcha: "updated" also fires for token-refresh, logo and validity writes, so the hook must check which fields changed.
- Gotcha: `meta_ads` and `google_ads` use `quietUpdate`, which fires no events.

**Disconnect:**
- Single: `IntegrationController::removeIntegration:228-346`
- Bulk: `bulkRemoveIntegration:353-436` is a **different path** (`RemovedAccountCleanup`)
- The public API `DELETE accounts/{id}` (`AccountController::destroy:264`) proxies to `removeIntegration`

**Needs reconnect:**
- Fields: `social_integrations.validity` ∈ `valid | expiring_soon | expired | invalid`, plus `validity_error` and `token_expires_at`
- Set by the **external Platform Observer service**, not this repo. The backend only triggers it via Kafka `platform-observer.account-validation-trigger`.
- The Observer calls back `POST /notifications/accounts/expired` → `AccountsNotificationsController::accountExpired:104` → `AccountExpirationNotificationService::sendToWorkspaceAdmins`, which sends an in-app notice plus email to admins. **This callback is the natural place to emit `account.reconnect_required`.**
- `validity` is exposed on the public API in `AccountResource.php:29`
- Verify: only `InstagramLoginConnector.php:297` resets `validity` to `valid` on reconnect. `SocialRepo::getUpdatedData:720` does not appear to, which would leave reconnected accounts flagged.

**Webhook pattern:**
- New event = enum case + `label()` / `group()` / `description()` in `app/Data/Webhooks/Enums/WebhookEventType.php`
- A payload builder, like `PostWebhookPayload`
- `app(WebhookEmitter::class)->emit($workspaceId, $type, $data)` publishes to Kafka `contentstudio.api.webhooks.events`
- Delivery goes to endpoints of the workspace owner and admins (`WebhookRepo::enabledEndpointsForWorkspaceEvent:171`). Webhooks are owned per user.
- The FE event picker needs the new group. Prior art: CONT-3222 (inbox events in the picker).

### 3.3 Large media upload

**Internal flow already used by the web app.** No Google SDK is involved: `contentstudio-frontend/package.json` has no Google package, and the upload is a plain `XMLHttpRequest` PUT plus `fetch` in `src/composables/useGCSUpload.ts`.
1. `POST media_library/assets/generate-signed-url` → `MediaLibraryAssetsController::generateSignedUrls:1988`
   - Storage-limit check against the declared size
   - A 60-minute single PUT URL
   - For files of 10 MiB or more (`config/media.php:96`), a 15-minute `resumable_init_url` via `GcsSignedUrlGenerator::signResumableInit:49`
   - The FE uploads resumable files in 8 MiB chunks
2. `POST media_library/assets/process-uploaded-gcs-media` → `processUploadedGCSMedia:2056`
   - Thumbnails or dimensions, then `MediaRepository::saveMedia`, then image processing

**Gotchas in the internal flow, which the public version must not copy:**
- No workspace-membership check (`workspace_id` comes from the body)
- The process step does not re-check storage or verify that the GCS object exists or matches the declared size

**Public API today:**
- `POST {ws}/media` → `MediaController::upload:310`
- `MediaUploadRequest` accepts a `url` (string max 2048 chars) **or** a `file`
- No size or type limits in the request; the cap is PHP `upload_max_filesize` = 2048M

### 3.4 API usage

**Metering:**
- `ApiKeyMiddleware::handle:27` charges 1 credit per workspace-scoped call (`:49`)
- Allowance = `subscription.limits.api_credit` + `addons.api_credit`
- Used = max `used_api_credit` across the owner's workspaces (`WorkspaceRepo:551`), written to all of them (`:542`). This is a non-atomic read-then-write.

**Reset:** `usedCredits:reset` on the 1st of each month (`config/billing.php:25`) zeroes `used_api_credit` and `used_webhook_delivery_credit`.

**Request logs** (the data source for history):
- Pipeline: `ApiRequestLogMiddleware` → Kafka `contentstudio.api.request-logs` → ClickHouse `api_request_logs`, with a 365-day TTL
- Fields: request_id, api_key_id, user_id, workspace_id, method, path, full_url, headers, bodies, status, duration, IP, user agent

**CONT-4619:**
- `full_url` keeps `?api_key=` unmasked
- Request and response bodies are stored raw
- Usage history must expose only aggregates (counts by day, endpoint, status, workspace), never bodies, and should land after CONT-4619

**Read side today (in the app only):**
- `RequestLogController` (`routes/api.php:164-166`)
- `ApiRequestLogBuilder`

**`GET {ws}/limits`:**
- `UsageLimitsService::forWorkspace:71`
- `api_credits` row: `{used, limit, remaining}`, where used = calls + webhook deliveries
- `usage_reset {last_reset_at, next_reset_at}`

### 3.5 Developer surfaces
- **CLI and MCP repos are not mounted locally.** Per memory: npm `contentstudio-cli` and `contentstudio-mcp`.
- **Production MCP:** 21 tools plus 305 long-tail ops across 10 toolsets.
- **Large upload for agents:** an agent in Claude or ChatGPT cannot stream a local multi-GB file through MCP. The MCP tool returns the upload URL and instructions; the CLI performs the actual upload.

## 4. Recommended approach
1. **Idempotency:**
   - Port the inbox design to PHP for `POST /posts` and the three AI generate endpoints: unique key per API key, fingerprint, 24h replay, and 409 on a body mismatch
   - Return an `Idempotency-Replayed: true` header on a replayed response
   - The CLI and MCP send a generated key on every create automatically
2. **Account events:**
   - `account.connected` (new and reconnected, with a `reconnected` flag), `account.disconnected`, `account.reconnect_required`
   - Emit on the save, remove (single and bulk) and expired-callback paths
   - Add an "Accounts" group to the in-app webhook event picker
3. **Large upload:**
   - Public three-step flow: start (signed single and resumable URLs), upload directly to storage, complete (verify the object, check storage, add to the library, return a media id)
   - Membership and storage checks on both steps
   - The CLI does the upload end to end
4. **Usage history:**
   - A read endpoint with daily totals for a date range, broken down by endpoint group and response status
   - Aggregates only
   - Credits charged by calls vs webhook deliveries shown separately
5. **Surfaces:** CLI commands, MCP tools or ops, docs, and confirmation that the features work in Zapier, Make and n8n.
