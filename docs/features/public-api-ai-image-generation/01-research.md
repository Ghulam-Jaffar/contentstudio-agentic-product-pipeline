# Research: AI image generation in the public API

**Date:** 2026-08-03
**Scope:** Expose ContentStudio's AI **image** generation tools across every public developer surface. Video is a separate epic.

---

## 1. Backlog check

Searched `docs/features/` (26 slugs) and `docs/stories/` (140+). **No existing work covers AI in the public API.**

Adjacent but distinct:

| Slug | What it actually is | Overlap |
|---|---|---|
| `api-centric-plan` | API-first onboarding, API key UI, nav gating for API-only plans | None. No AI scope. |
| `contentstudio-public-cli-agent-skills` | Ships the CLI, agent skill, MCP discovery | **Structural precedent.** This epic extends the surfaces it created. |
| `publishing-api-v15-v16` / `v17` / `v19-v22` | Posts, filters, labels, campaigns, platform overrides | None. Publishing only. |
| `public-analytics-api-rollout` | Analytics endpoints on the public API | None, but a useful precedent for adding a new domain. |
| `public-webhooks` | Webhook management API | Relevant later for video. Not needed for image, which is synchronous. |

---

## 2. The 8 image tools in scope

Confirmed against `contentstudio-ai-agents`. All are synchronous: one request, one result.

| # | Tool key | What it does | Service route |
|---|---|---|---|
| 1 | `text-to-image` | Generate an image from a prompt | `POST /api/v1/image/generate` |
| 2 | `image-to-image` | Transform an existing image from a prompt | `POST /api/v1/tools/image-to-image` |
| 3 | `product-image` | Product shot generation | `POST /api/v1/tools/product-image` |
| 4 | `headshot` | Professional headshot generation | `POST /api/v1/tools/headshot` |
| 5 | `face-swap` | Swap a face into an image | `POST /api/v1/tools/face-swap` |
| 6 | `outfit-swap` | Swap clothing in an image | `POST /api/v1/tools/outfit-swap` |
| 7 | `upscale` | Increase image resolution | `POST /api/v1/tools/upscale` |
| 8 | `remove-background` | Cut out the background | `POST /api/v1/tools/remove-background` |

Tools 3 to 8 are declared as `ToolSpec` entries in `contentstudio-ai-agents/src/tools/dedicated/specs.py` and wired with one line each in `src/api/routers/dedicated_tools_router.py` via `make_tool_route`. They share a uniform request/response contract.

**`image-to-image` is the exception.** Its route comment says it is *"bespoke — exposes model selection / free-text prompt / N attachments / mentions / logo, so it delegates to the chat core, not the ToolSpec engine."* Its request shape differs from the other five dedicated tools, so it cannot be mapped generically.

**Supporting service endpoints that already exist:**
- `GET /api/v1/image/models` — available image models
- `GET /api/v1/image/dimensions` — supported output dimensions
- `GET /api/v1/tools` and `GET /api/v1/tools/{key}` — live capability feed describing each tool's inputs and controls
- `POST /api/v1/image/batch` — batch generation
- `POST /api/v1/image/enhance-prompt` — prompt improvement
- `POST /api/v1/image/generate-alt-text`, `POST /api/v1/image/describe` — image understanding

**Models available:** 20 image models (`gpt-image-2`, `nano-banana-pro`, `flux_2_pro`, `imagen4-preview-ultra`, `kling-image-o3`, `qwen-image-max` and others), in `contentstudio-ai-agents/src/utils/model_registry.py`. Default is `nano-banana-pro`, set via `AI_AGENTS_DEFAULT_IMAGE_MODEL`.

---

## 3. Current public API

`contentstudio-backend/routes/api/v1.php`. Base `https://api.contentstudio.io/api/v1`, auth `X-API-Key: cs_...`.

Existing domains: `me`, workspaces CRUD, accounts, content categories, labels, campaigns, media, team members, posts, post approval, comments, and per-platform analytics (Facebook, Instagram, LinkedIn, YouTube, TikTok, Pinterest, GMB).

**There is no AI of any kind on the public API today.** Not generation, not analysis, not the capability feed.

**Middleware stack** on the whole v1 group:

```php
Route::middleware(['force.json', 'api.key', 'api.request.log', 'throttle:api-v1'])
```

---

## 4. How Laravel reaches the AI service

The AI agents service is **not publicly reachable** and must stay that way. Laravel calls it server to server.

- `contentstudio-backend/config/ai_agents.php` — `base_url` (`https://ai-agents-stage.contentstudio.io/api/v1/`), `api_key` from `AI_AGENTS_API_KEY`, `timeout` 300s, `default_image_model`, tool capability cache TTL 300s.
- `app/Services/AI/AiAgentService.php` — the shared client. `request()`, `requestRaw()`, `requestOrFail()`, `getRequest()`, `deleteRequest()`, `streamRequest()`, plus SSE helpers.
- `app/Services/AI/AiToolCapabilityService.php` — `list()` and `show($toolKey)` over the live capability feed, already cached.

**This means the public API work is a proxy layer, not new AI infrastructure.** The generation already works. What is missing is a public contract in front of it: public key auth, workspace scoping, credit metering, rate limiting, error normalisation, and persisting the output into the workspace.

---

## 5. Two separate credit systems, both relevant

This is the most consequential finding for the PRD.

**API credits** — `app/Http/Middleware/ApiKeyMiddleware.php:88-108`:
```php
$allowedCredits = (int)($workspace->user->subscription->limits['api_credit'] ?? 0)
                + (int)($workspace->user->addons['api_credit'] ?? 0);
if ($usedCredits >= $allowedCredits) { /* 403 */ }
```
Every public API request already burns one API credit. Line 108 notes the increment is skipped "for internal ai_creation keys", so an internal key type for AI already exists.

**Image generation credits** — `app/Libraries/Settings/SubscriptionLimits.php:49,75`:
```php
$imageGeneration = $workspace['image_generation_credit'] ?? 0;
```
This is the quota the web app's AI Studio spends against.

The two are independent today because they have never overlapped. An API-triggered image generation touches both, and the PRD has to say explicitly what gets charged. Getting this wrong means either giving away paid generation for free through the API, or double-charging customers who already bought image credits.

---

## 6. Constraints the API design has to absorb

**Latency.** Image generation takes roughly 5 to 60 seconds depending on model. The existing public API is built around sub-second responses. `AI_AGENTS_TIMEOUT` is already 300s server side, but the public-facing route needs its own timeout policy, and the docs need to tell integrators to raise their client timeouts. Zapier, Make, and n8n each have their own request timeouts that may be shorter than a slow model, which is a real integration risk.

**Rate limiting.** `throttle:api-v1` is a shared bucket. Image generation is orders of magnitude more expensive per call than `GET /posts`, so it needs a tighter dedicated throttle rather than inheriting the general one.

**Where the output goes.** The service returns generated image URLs. For the API to be useful the result should land in the workspace media library, so the returned `media_id` can be passed straight into `POST /workspaces/{ws}/posts`. Generate then publish in two calls is the flow that makes this epic worth shipping. `POST /workspaces/{workspace_id}/media` already exists for uploads and is the natural place for generated assets to appear.

**Safety and content policy.** `src/agents/image/safety.py` and `brand_intent_classifier.py` exist in the service, so generation can be refused. Refusals need a distinct, documented error code rather than a generic 500, or every integrator will treat a policy refusal as an outage.

---

## 7. Surfaces to cover

Per the current live inventory:

| Surface | Today | This epic |
|---|---|---|
| REST API | 5 domains, no AI | Add the image domain |
| CLI (`contentstudio-cli`) | `posts:*`, `accounts:*`, media | Add `images:*` commands |
| Agent skill (`contentstudio-agent`) | Publishing workflows | Document image generation |
| MCP server (`contentstudio-mcp`) | Publishing tools | Add image tools |
| Claude Desktop (MCPB bundle) | Ships the MCP server | Inherits the new MCP tools |
| Zapier / Make / n8n | Publishing actions | Add generation actions |

The CLI, MCP, and Claude Desktop paths all consume the REST API, so they are strictly downstream of the API story. Zapier, Make, and n8n are separate connector codebases with their own release cycles.

---

## 8. Recommended split

Five stories. One deep, four thin, as requested.

| # | Story | Depth |
|---|---|---|
| S-1 | `[BE] Add AI image generation endpoints to the public API` | Full |
| S-2 | `[BE] Add image generation commands to the CLI and agent skill` | Light |
| S-3 | `[BE] Expose image generation tools in the MCP server and Claude Desktop bundle` | Light |
| S-4 | `[BE] Add image generation actions to Zapier, Make, and n8n` | Light |
| S-5 | `[BE] Publish public documentation and quickstarts for image generation` | Light |

No frontend story. Nothing changes in the web app. No mobile story. No design story, since there is no UI.
