# Public API gap analysis: ContentStudio vs the market

**Date:** 2026-10-21
**Scope:** public REST API v1, webhooks, OAuth, MCP, CLI. Compared against API-first vendors (Zernio, Ayrshare, Postiz, Upload-Post, bundle.social, Outstand, Postproxy) and incumbents (Buffer, Hootsuite, Sprout Social, Publer, Metricool, Agorapulse, Vista Social, Later, Sendible, SocialBee, Planable).
**Baseline:** `contentstudio-backend` `origin/master` (identical to `origin/develop` for `routes/api/v1.php`). Anything only on the `features` branch is marked as such.

---

## TL;DR

- **We are already ahead of every incumbent** on API breadth. Buffer, Hootsuite, Sprout, Publer, Metricool and Agorapulse all trail us on webhooks, approvals, inbox, automations, reports and OpenAPI. Only Buffer also ships an official CLI.
- **The real competition is the API-first vendors** (Zernio, Ayrshare, Upload-Post, bundle.social). The gap with them is not *what you can do*. It is **how safe and embeddable the API is**: idempotency, preflight validation, scoped keys, branded connect, account-health events, usage self-service, SDKs and a sandbox.
- **Three things are hurting us right now, and they are cheap to fix:**
  1. Webhook deliveries are charged against API credits, and when credits run out they are **skipped without retry**. Customers silently lose events, and no competitor meters webhooks.
  2. `POST /posts` has **no `Idempotency-Key`**. An agent or Zapier retry can double-post. Zernio, Ayrshare, Upload-Post and Outstand all support the header.
  3. Finished work is not reaching customers. The webhook-management API and Brand Knowledge API sit on the `features` branch only, and production MCP has none of AI image, AI video, AI carousel, webhooks or Brand Knowledge.

---

## 1. Where we stand today (verified in code)

| Area | ContentStudio today |
|---|---|
| Publishing | Full CRUD with `platform_overrides` for 12 networks, repeat posts, first comment, queue, content categories, approval routing, delete-from-network. No bulk create. |
| Media | Upload by URL or by multipart file, plus folders, archive, move, notes and storage. No presigned or resumable upload. |
| Approvals | Workflow CRUD plus approve and reject. **Only Hootsuite has anything comparable.** |
| Inbox | Search, read, reply, DM, hide, like, delete, reviews, tags, notes, bulk triage, `Idempotency-Key` |
| Analytics | 11 networks (including Threads and Bluesky), Meta Ads, Google Ads, campaign and label analytics, competitor reports (FB/IG), reports, schedules, share links |
| Automations | RSS, evergreen and bulk CSV over the API (only Ayrshare comes close) |
| AI | Image, video and carousel generation, with async jobs |
| Accounts | One-time OAuth connect URL with `return_url`, Bluesky credentials, disconnect |
| Webhooks | 15 events. HMAC-SHA256 signing with dual signatures during secret rotation, 7 retries with jitter, auto-disable, SSRF guard, preflight ping |
| Auth | One API key per user (`X-API-Key`, Bearer or `?api_key=`). OAuth server for MCP connectors, with redirect hosts limited to Claude and ChatGPT. Authorization follows the workspace role. |
| Limits | 100 req/min per key. 1 API credit per workspace-scoped call. `GET /limits` is read-only. |
| DX | OpenAPI at `/api-docs.json`, guide at `/guide`, CLI (`contentstudio-cli`), MCP (21 tools plus 305 long-tail ops), `llms.txt` on the website |

---

## 2. Gaps and recommended initiatives

Ranked by user value and effort. "Who has it" lists competitors that document the capability.

### P0: close the trust holes (small effort, high damage if left)

**1. Ship what is already built, and stop MCP lagging the API**
- Gap:
  - The webhook-management API (8 routes) and the Brand Knowledge API (10 routes) are on `features` only.
  - Production MCP has no AI image, video or carousel generation, no webhooks and no Brand Knowledge.
  - The planned parity guard (`public-api-parity-contract`) does not exist, so this drift will keep happening.
- Who has it: Zernio, Ayrshare and Buffer ship API, MCP and CLI together.
- Ask: merge to master, expose these in MCP and CLI, and build the parity guard.

**2. `Idempotency-Key` on post create, media upload and AI generate**
- Gap: only the inbox routes accept it, so retries on `POST /posts` can double-publish. The risk is highest with agents and Zapier, which retry by design.
- Who has it: Zernio (24h replay), Ayrshare, Upload-Post, Outstand (`Idempotency-Replayed` header).
- Ask: reuse the inbox implementation pattern (`BaseApiController.php`).

**3. Stop dropping webhooks when credits run out**
- Gap: a delivery is charged against the API-credit pool. When the pool is empty the delivery is skipped with no retry. Each user is also capped at 5 endpoints.
- Who has it: no competitor meters webhook deliveries.
- Ask: stop metering deliveries, or at least queue them until credits return, and email the owner when that happens.

**4. Account health and lifecycle events**
- Gap: there is no `account.connected`, `account.disconnected` or `account.reauth_required` event and no account health endpoint. An integrator only finds a dead token when a post fails.
- Who has it: Zernio, Ayrshare, Upload-Post, bundle.social, Outstand.
- Ask: add a health endpoint per account, plus the three events. The in-app reconnect flow already knows the state.

**5. Usage and limits self-service (the original "auto-manage limits" question)**
- Gap: `GET /limits` is a point-in-time read. Missing:
  - usage history, broken down by day or endpoint
  - request logs over the API (they are viewable in the app only)
  - a threshold webhook (80% and 100% of credits)
  - a credits-remaining header on responses
  - any way to manage the X wallet (auto-recharge, spending cap, top-up) outside the logged-in app
- Who has it:
  - **Zernio is the only vendor that sets spending caps via API** (billing snapshot plus usage budgets).
  - Upload-Post, bundle.social and Outstand let you read usage.
  - Buffer has the best headers: per-window `RateLimit` and `RateLimit-Policy`, and one shared bucket for API, MCP and AI.
  - **No vendor lets you purchase through the API.**
- Ask:
  - `GET /usage` with history
  - `usage.threshold` webhook
  - credit headers on responses
  - X wallet read, auto-recharge and spending-cap endpoints gated by `billing_access`
- Leave purchasing and add-on buying out of the API. Nobody offers it.

### P1: make us embeddable (SaaS builders, agencies, AI agents)

**6. Scoped API keys**
- Gap:
  - Each user has one key, and it carries that user's full power across every workspace they belong to.
  - `contentstudio:read` and `contentstudio:write` are advertised in `config/oauth.php`, but nothing enforces them.
  - An agency cannot give a client, contractor or Zap a key limited to one workspace, or a read-only key.
- Who has it: Zernio (keys scoped to a profile), Postproxy (keys scoped to a profile group), Buffer (OAuth scopes), Publer (scoped keys).
- Ask: multiple named keys, each restricted to chosen workspaces and either read-only or read-write, with scope enforcement in `ApiKeyMiddleware`.

**7. Branded and headless account connect**
- Gap: the connect URL exists, but the end customer sees ContentStudio branding. There is also no headless mode, where the integrator picks which pages or organizations to add in their own UI.
- Who has it: Ayrshare (JWT connect plus branding), Upload-Post (logo, title, language, labels), bundle.social (portal link, "hide powered by"), Zernio (headless `pendingDataToken`).
- Ask: branding options on the connect link (logo and name, or white-label domain), plus a headless page-selection step.

**8. Third-party OAuth apps**
- Gap: the OAuth server exists, but redirect hosts are limited to the AI vendors. Partners cannot build "Connect your ContentStudio" into their own product.
- Who has it: Buffer, Sprout, Hootsuite and Postiz all run OAuth apps.
- Ask: partner app registration with redirect URIs and enforced scopes. This is also how we would get into app marketplaces.

**9. Preflight validation (dry run)**
- Gap: there is no way to check a post without creating it. The rules already exist (`PostStoreRequest.php`, 2,873 lines).
- Who has it: Zernio (`validate-post`, `validate-post-length`, `validate-media`), Ayrshare (validation endpoints).
- Ask: `POST /posts/validate` (or `?dry_run=true`) that returns per-platform errors, character counts and media problems.

**10. Partial success and per-platform results**
- Gap: `post.partial` is deferred, and a partial failure is folded into `post.published`. There is no retry for only the failed accounts.
- Who has it: Zernio (207 status plus `post.partial`), Ayrshare and bundle.social (retry).
- Ask:
  - a result array per account on the post and in webhooks
  - a `post.partial` event
  - `POST /posts/{id}/retry` for the failed accounts only

**11. Large and resumable media upload**
- Gap: upload is limited to a URL or a single multipart request. There is no presigned or chunked upload for large video.
- Who has it: Zernio (presigned), bundle.social (S3 multipart), Sprout (5 MiB parts), Ayrshare (large-file upload).
- Ask: a presigned upload URL plus a confirm step.

### P2: depth and breadth

**12. Publishing payload gaps**
- Not in the API payload today:
  - alt text
  - Instagram user tags and product tags
  - location
  - YouTube playlist
  - Pinterest title
  - Instagram share-to-feed
  - bulk post create
  - an external reference ID (bundle.social's `reference key`)
  - account and platform filter on the post list (already planned in `public-api-publishing-gaps`)
- Confidence is medium: these are absent from the request rules and the transformer, but I did not trace the composer path end to end. Each one needs checking against the app before a story is written.

**13. Webhook event breadth**
- Gap: we have 15 event types against about 60 at Zernio.
- Highest-value additions:
  - `account.*` (initiative 4)
  - `post.partial`, `post.created`, `post.updated`, `post.deleted`
  - `ai.job.completed` (planned with the video API)
  - `analytics.synced` and `report.ready` (planned in `analytics-webhook-events`)
  - `usage.threshold` (initiative 5)

**14. AI text generation on the API**
- Gap: captions, hashtags, alt text, rewrite and translate are not on the API.
- Who has it: Ayrshare `/generate` (text, rewrite, translate, alt text, hashtags), Upload-Post (titles and hashtags).
- **Differentiator:** nobody exposes a full agent. "Run an AI Chat Skill over the API" would be unique to us, since Skills are already live in the product.

**15. Inbox automation rules over the API**
- Gap: the auto-reply engine (keyword and AI rules) exists in `social-inbox-manager/app/services/auto_reply/`, but its rules cannot be managed through the API.
- Comment-to-DM is the 2026 trend: Zernio, Ayrshare, Upload-Post, bundle.social and PostPeer all added it. Whether our engine can send a DM rather than a public reply is unverified.

### P3: developer experience and commercial model

**16. Official SDKs.** We have none. Zernio has 8 languages, Ayrshare has Node and Python, and Postiz, bundle.social and Upload-Post each have one or two. Generate Node and Python SDKs from our OpenAPI spec.

**17. Sandbox and test keys.** There is no test mode. Postproxy offers a sandbox on every plan. A test key that validates and simulates without publishing removes the biggest fear in onboarding.

**18. API hygiene:**
- cursor pagination (today it is offset only)
- a public API changelog and deprecation policy (Zernio even fires an `api.changelog.published` webhook)
- a status page
- a reference page of character and media limits per platform

**19. Pricing and metering model (strategic, not engineering)**
- We charge 1 credit per call, including reads and webhook deliveries.
- Every API-first vendor charges per connected account, profile or post, with reads and webhooks free.
- Per-call metering punishes exactly what we want integrators to do: polling, reading limits before a batch, and receiving webhooks.
- This is worth a pricing review alongside the `api-centric` plan.

---

## 3. Where we are ahead, worth marketing

- **Approval workflows over the API.** Among incumbents only Hootsuite has this, and no API-first vendor has more than an `requiresApproval` flag.
- **Webhook engineering** matches the best of the market: dual-signature rotation, jitter, auto-disable, SSRF guard and preflight.
- **Analytics depth:** 11 networks, Meta and Google Ads, competitor reports, scheduled reports and share links. Only Metricool is comparable.
- **A full inbox API with idempotency.** Sprout and Hootsuite expose parts of the inbox, and Agorapulse only via MCP.
- **Automations (RSS, evergreen, bulk CSV)** over the API.
- **API, CLI and MCP together.** Among incumbents only Buffer matches this.

---

## 4. Competitor snapshot

| Capability | CS | Zernio | Ayrshare | Upload-Post | bundle.social | Postiz | Buffer | Hootsuite | Sprout |
|---|---|---|---|---|---|---|---|---|---|
| Publish CRUD | Y | Y | Y | Y | Y | Y | Y | Y | drafts |
| Bulk post create | CSV only | Y | Y | ? | Y | N | N | ? | N |
| Idempotency on create | **N** | Y | Y | Y | ? | N | ? | ? | ? |
| Preflight validation | **N** | Y | Y | partial | partial | partial | N | N | N |
| Presigned or multipart upload | **N** | Y | Y | async | Y | partial | ? | ? | Y |
| Approvals API | **Y** | N | flag | N | N | preview | ? | Y | N |
| Inbox (comments, DMs, reviews) | **Y** | Y | Y | partial | Y | N | emerging | Y | export |
| Comment-to-DM automation | ? | Y | Y | Y | Y | N | N | N | N |
| Analytics | **Y (deep)** | Y (deep) | Y | Y | Y | Y | Y | MCP | Y |
| Webhook event types | 15 | ~60 | ~11 | 4 | 8 | 1 | 0 | Y | 0 |
| Account health / reauth events | **N** | Y | Y | Y | Y | Y | N | ? | N |
| Branded connect page | **N** | partial | Y | Y | Y | N | n/a | n/a | n/a |
| Scoped or per-tenant keys | **N** | Y | Y | Y | Y | OAuth | OAuth | OAuth | OAuth |
| Usage read via API | Y (`/limits`) | Y | partial | Y | Y | N | headers | N | N |
| Set spending caps via API | **N** | Y | N | N | N | N | N | N | N |
| SDKs | **N** | 8 | 2 | 2 | 1 | 1 | ? | ? | ? |
| Sandbox | **N** | messaging | test posts | N | N | self-host | N | N | N |
| MCP | Y | Y | Y | Y | Y | Y | Y | Y (4) | N |
| CLI | Y | Y | N | Y | Y | Y | Y | N | N |
| OpenAPI | Y | Y | ? | Y | Y | Y | explorer | ? | N |
| Billing unit | API call | account | profile | profile | post | channel | plan | plan | plan |

Key: Y = documented, N = not offered or not found, ? = unverified, partial = partly offered, n/a = not applicable.

---

## 5. Evidence pointers (ContentStudio code)

- Routes: `contentstudio-backend/routes/api/v1.php`, `routes/oauth.php`, `routes/billing.php`
- Auth and credits: `app/Http/Middleware/ApiKeyMiddleware.php`, `app/Models/ApiKey.php`, `config/oauth.php:55` (scopes advertised, not enforced)
- Rate limits: `config/api_rate_limits.php`
- Idempotency (inbox only): `app/Http/Controllers/Api/V1/Inbox/InboxActionController.php:45`
- Webhooks: `app/Data/Webhooks/Enums/WebhookEventType.php`, `config/webhooks.php`, `app/Services/Webhooks/WebhookSigner.php`
- Post payload: `app/Http/Requests/Api/V1/PostStoreRequest.php`
- Media upload: `app/Http/Requests/Api/V1/MediaUploadRequest.php` (URL or file only)
- Limits: `app/Http/Controllers/Api/V1/LimitController.php`
- Inbox auto-reply engine: `social-inbox-manager/app/services/auto_reply/`

## 6. Competitor sources

- Zernio: https://docs.zernio.com/llms.txt, https://zernio.com/pricing, https://docs.zernio.com/webhooks.mdx, https://docs.zernio.com/guides/rate-limits.mdx
- Ayrshare: https://www.ayrshare.com/docs/_llms/en/documentation.txt, https://www.ayrshare.com/pricing
- Postiz: https://docs.postiz.com/llms.txt, https://postiz.com/pricing
- Upload-Post: https://docs.upload-post.com/llms.txt
- bundle.social: https://info.bundle.social/llms.txt, https://bundle.social/pricing
- Outstand: https://www.outstand.so/llms.txt
- Postproxy: https://postproxy.dev/llms.txt
- Buffer: https://developers.buffer.com (authentication, api-limits, integrations guides)
- Hootsuite: https://developer.hootsuite.com/docs/api-overview, https://www.hootsuite.com/integrations/mcp
- Sprout Social: https://api.sproutsocial.com/docs/
- Publer: https://publer.com/docs
- Metricool: https://help.metricool.com/mcp-vs-api-access-what-is-the-difference-5y3ib
- Agorapulse: https://support.agorapulse.com/en/articles/9155447-analytics-open-api
- Vista Social: https://vistasocial.com/insights/social-media-management-api-pricing/

**Unverified:**
- Hootsuite's and Agorapulse's API references are rendered in JavaScript, so their endpoint details are partly secondhand.
- Field-level options at Zernio, Ayrshare and bundle.social (alt text, user tags, Stories, polls) were not checked.
- Zernio's pricing page says analytics is included, but its docs call follower stats an add-on.
