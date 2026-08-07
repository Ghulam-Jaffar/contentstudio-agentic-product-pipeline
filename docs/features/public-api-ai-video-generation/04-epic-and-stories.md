# Epic + Stories — AI Video Generation in the Public API

---

## Epic

**Title:** AI Video Generation in the Public API

**Description:**

Expose ContentStudio's 5 AI video tools through the public REST API and every downstream developer surface: the CLI and agent skill, the MCP server and Claude Desktop bundle, and the Zapier, Make, and n8n connectors.

Video generation takes minutes, so the real deliverable is an **asynchronous job contract**. Submit returns a job handle immediately, callers poll one status endpoint regardless of which tool they used, and integrators who prefer push receive signed `video.completed` and `video.failed` webhooks. Both mechanisms ship in this epic.

The epic also unifies two internal async engines behind one public job resource, and adds cost estimation, because video credit cost varies by model, duration, resolution, and audio rather than being a flat rate.

**Tools in scope (5):** `text-to-video`, `image-to-video`, `video-clips-highlights`, `motion-control`, `lip-sync`. Plus `reference-to-video` as a third mode on the generate endpoint.

**Out of scope:** image generation, which is the sibling epic and is synchronous. Also out: caption, hashtag, and post generation, bulk schedule, smart scheduling, and public SSE streaming.

| # | Story | Priority |
|---|---|---|
| S-1 | [BE] Add AI video generation endpoints and the async job contract to the public API | High |
| S-2 | [BE] Add video job events to the public webhooks system | High |
| S-3 | [BE] Add video generation commands to the CLI and agent skill | High |
| S-4 | [BE] Expose video generation tools in the MCP server and Claude Desktop bundle | High |
| S-5 | [BE] Add video generation actions to Zapier, Make, and n8n | Medium |
| S-6 | [BE] Publish public documentation and quickstarts for video generation | Medium |

S-1 gates everything. S-2 additionally depends on the public webhooks epic landing. S-3 through S-6 can run in parallel once S-1 is done.

---

## S-1 · [BE] Add AI video generation endpoints and the async job contract to the public API
**Project:** Web App · **Group:** Backend · **Skill:** Backend · **Product area:** Public API · **Priority:** High · **Type:** Feature

### Description

As a developer or AI agent using the ContentStudio API, I want to submit a video generation request, get a job handle back immediately, and poll one endpoint for the result, so that I can generate video programmatically without my request hanging for minutes.

Video is the highest-value AI capability we have and it is entirely unreachable outside the web UI. An API consumer can schedule a post but cannot produce the video for it.

The generation already works. This story builds the public contract in front of it: the async job resource, cost estimation, credit metering, storage checks, rate limiting, error normalisation, and persistence into the workspace media library.

**Two things make this materially different from the image epic:**

**It is asynchronous.** Generation takes minutes. Submit returns a handle, and the caller polls. An optional bounded wait lets fast jobs return inline so the documented quickstart is one call rather than three.

**There are two internal async engines.** Four tools run on Dramatiq (`job_id`, `GET /jobs/{id}`). `video-clips-highlights` runs on Temporal (`workflow_id`, `GET /workflows/{id}/status`, plus a separate result route and different status casing). The public API must present **one** job resource and normalise both. A caller must not be able to tell which queue ran their job.

**Endpoints:**

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/workspaces/{workspace_id}/ai/videos/tools` | Tools, supported modes, input schemas |
| `GET` | `/workspaces/{workspace_id}/ai/videos/models` | Models with supported modes, resolutions, durations |
| `POST` | `/workspaces/{workspace_id}/ai/videos/estimate` | Credit cost and time estimate, without submitting |
| `POST` | `/workspaces/{workspace_id}/ai/videos/generate` | `text-to-video`, `image-to-video`, `reference-to-video` |
| `POST` | `/workspaces/{workspace_id}/ai/videos/tools/{tool_key}` | `motion-control`, `lip-sync`, `video-clips-highlights` |
| `GET` | `/workspaces/{workspace_id}/ai/jobs/{job_id}` | Status, stage, progress, result |
| `GET` | `/workspaces/{workspace_id}/ai/jobs` | List jobs, filterable by status |
| `DELETE` | `/workspaces/{workspace_id}/ai/jobs/{job_id}` | Cancel |

The job resource is `/ai/jobs`, not `/ai/videos/jobs`, so any future async AI work reuses it rather than growing a parallel contract.

### Workflow

```mermaid
sequenceDiagram
    actor Dev as Developer or agent
    participant API as ContentStudio public API
    participant AI as AI agents service
    participant Media as Workspace media library

    Dev->>API: POST /ai/videos/estimate
    API-->>Dev: credit cost and estimated seconds

    Dev->>API: POST /ai/videos/generate with optional wait
    API->>API: validate key, workspace, credits, storage, rate limit
    API->>AI: enqueue on Dramatiq or Temporal
    AI-->>API: job id and estimated seconds
    API-->>Dev: job id, status URL, estimated cost

    loop Until terminal
        Dev->>API: GET /ai/jobs/{job_id}
        API-->>Dev: status, stage, progress
    end

    AI-->>API: generation complete
    API->>Media: persist video
    Media-->>API: media_id
    API->>API: charge video credits
    API-->>Dev: completed job carries media_id and final cost

    Dev->>API: POST /posts with that media_id
    API-->>Dev: scheduled post
```

1. A developer calls the models endpoint to see which models exist and which modes, resolutions, and durations each supports.
2. They call the estimate endpoint with their intended request and get back what it will cost in credits and roughly how long it will take. Nothing is submitted and nothing is charged.
3. They submit the generation request and immediately receive a job ID, a status URL, and the estimated cost.
4. They poll the job status endpoint and see the state move from queued to processing, with a progress percentage and a human-readable stage.
5. When the job completes, the response carries the finished video's media ID and the final charged credit cost.
6. They pass that media ID into the create-post endpoint and the video is attached to a scheduled post.
7. If they had asked for a bounded wait and the job finished inside it, they get the finished result on the submit call and never poll at all.
8. If the job fails, they are charged nothing and the response explains why.
9. If they cancel a running job, they are charged nothing, including for partial work.

### Acceptance criteria

**Tool coverage and discovery**
- [ ] `GET /ai/videos/tools` returns the available video tools with, for each, its key, name, description, supported modes, and input schema.
- [ ] `GET /ai/videos/models` returns the available models, and for each one the generation modes, resolutions, and durations it supports.
- [ ] Mode support is reported **per model**, not as a global list, because models differ in which modes they support.
- [ ] Both discovery endpoints derive from the AI service's live capability feed, so a tool or model added to the service appears without a Laravel release.
- [ ] `POST /ai/videos/generate` supports `text-to-video`, `image-to-video`, and `reference-to-video`.
- [ ] `POST /ai/videos/tools/{tool_key}` executes `motion-control`, `lip-sync`, and `video-clips-highlights`.
- [ ] Requesting a mode, resolution, or duration the chosen model does not support returns `422` naming the unsupported combination.
- [ ] An unknown `tool_key` returns `404`.

**One job contract over two engines**
- [ ] All five tools return a job handle in the **same** response shape, whichever internal engine runs them.
- [ ] `GET /ai/jobs/{job_id}` returns the same response shape for all five tools.
- [ ] A caller cannot determine from any public response or identifier which internal async engine ran their job.
- [ ] `video-clips-highlights`, which runs on Temporal, is reachable through the same job ID, status endpoint, and cancel endpoint as the four Dramatiq tools.
- [ ] The result shape is normalised across tools. Dedicated-tool jobs and generation jobs return the finished video in the same field, despite differing internally.

**Job state machine**
- [ ] The public `status` field only ever returns `queued`, `processing`, `completed`, `failed`, or `cancelled`.
- [ ] Granular pipeline stages are exposed in a separate `stage` field, never in `status`.
- [ ] `progress` is returned as an integer 0 to 100.
- [ ] `message` carries human-readable status text.
- [ ] The documentation marks `stage`, `progress`, and `message` as informational and subject to change without notice, and marks `status` as the stable contract.

**Bounded wait**
- [ ] Submit accepts an optional `wait` parameter, capped at a configurable maximum.
- [ ] If the job completes inside the window, the submit response returns the finished result including the media ID.
- [ ] If it does not complete, the submit response returns the job handle exactly as if `wait` had been omitted.
- [ ] A request exceeding the cap is rejected with `422` rather than being silently clamped.
- [ ] Polling and cancellation work identically whether or not `wait` was used.

**Cost estimation and credits**
- [ ] `POST /ai/videos/estimate` returns the credit cost and estimated completion time for a request without submitting it or spending anything.
- [ ] The estimate accounts for model, duration, resolution, and whether audio is enabled.
- [ ] The submit response includes the estimated credit cost and the estimated seconds to completion.
- [ ] A completed job carries the **final charged** credit cost, which is authoritative over the estimate.
- [ ] Video credits are checked at submit and charged on completion.
- [ ] A **failed** job charges no video credits. Any credits reserved at submit are released.
- [ ] A **cancelled** job charges no video credits, including for partial work already done.
- [ ] Submitting with insufficient video credits returns `403` with a distinct code, stating the required and available amounts.
- [ ] One API credit is spent at submit, consistent with every other v1 endpoint.

**Output and storage**
- [ ] Every completed job persists the video into the workspace media library.
- [ ] The completed job result includes a media ID, URL, duration, resolution, and MIME type.
- [ ] The returned media ID is accepted by `POST /workspaces/{workspace_id}/posts` and the video attaches to the created post.
- [ ] Generated video appears in `GET /workspaces/{workspace_id}/media` alongside uploaded media.
- [ ] Workspace storage limits are checked **at submit**. A workspace at its limit is rejected before generation starts, never after minutes of work.

**Rate limiting and concurrency**
- [ ] Video submit endpoints use their own throttle bucket, separate from the general v1 throttle and from image.
- [ ] A per-workspace cap limits concurrently running jobs, and exceeding it returns `429`.
- [ ] Job status polling has a more permissive limit than submit, since polling is the mechanism we document.
- [ ] `429` responses include `Retry-After`.

**Job lifecycle**
- [ ] `DELETE /ai/jobs/{job_id}` cancels a queued or running job and moves it to `cancelled`.
- [ ] Cancelling an already-terminal job returns `409`.
- [ ] `GET /ai/jobs` lists the workspace's jobs and can be filtered by status.
- [ ] Completed and failed jobs remain queryable for a documented retention window.
- [ ] A job pruned past retention returns `410`, distinct from `404` for an ID that never existed.
- [ ] Pruning a job does not affect the generated media in the library.

**Errors**
- [ ] A content policy refusal returns `422` with a distinct code and a reason, never a server error.
- [ ] An upstream model or provider failure returns `502` with a code indicating the request is safe to retry.
- [ ] Every error response follows the same top-level shape as the rest of the v1 API.

**Isolation and security**
- [ ] The AI agents service is never reachable from the public internet. All access is proxied through Laravel with the internal service key.
- [ ] The internal service key never appears in any public response, error body, or log line.
- [ ] Job endpoints are workspace-scoped and reject keys without access to that workspace.
- [ ] A caller cannot read or cancel a job belonging to a workspace they do not have access to.
- [ ] An AI service outage causes video endpoints to fail with `502` and leaves every other v1 domain unaffected.

**Not in scope**
- [ ] The SSE event stream is **not** exposed publicly.

**Documentation**
- [ ] The OpenAPI specification covers all eight endpoints, all five tools, the job state machine, and the full error code table.
- [ ] The interactive API reference at `/guide` renders the new endpoints.

### Impact on existing data

Adds generated videos to the workspace media library, where they are indistinguishable from uploads to downstream consumers. Introduces a public job record with a retention policy. Spends against the video credit balance, which no API path has touched before. Video files are much larger than images and will grow workspace storage consumption faster.

### Impact on other products

None directly. Web app, mobile apps, and Chrome extension unchanged. Everything else in this epic depends on these endpoints.

### Dependencies

- None blocking, but shipping the sibling image epic first establishes the `/ai/*` namespace, error shapes, and credit-metering pattern this extends.
- PRD open questions 1, 2, 3, 4, 6, 7, and 8 need answers before build: the bounded-wait cap, whether credits are reserved or only checked, the concurrency cap, the retention window, who absorbs upstream cost on late failure, whether batch is in v1, and whether generated media needs a provenance marker.

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support — N/A, no interface
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Dramatiq job routes to proxy** (`contentstudio-ai-agents/src/api/routers/jobs/jobs_router_dramatiq.py`):
- `POST /api/v1/jobs/video`, `POST /api/v1/jobs/video/batch`
- `GET /api/v1/jobs/{job_id}`, `DELETE /api/v1/jobs/{job_id}`, `GET /api/v1/jobs/`
- `VideoJobRequest` already carries `model_preference`, `concept`, `generation_mode`, `source_image_url`, `reference_image_urls`, `style`, `aspect_ratio`, `duration_seconds`, `resolution`, `brand_guidelines`, `enhance_prompt`, `enable_audio`, `seed`, `video_context`.
- `JobEnqueueResponse` already returns `job_id`, `status_url`, `events_url`, `estimated_seconds`, `status`. **`estimated_seconds` is exactly what the bounded wait and the submit response need, and it is free.**
- `JobInfo` returns `id`, `status`, `progress`, `message`, `error`, `result`.

**Temporal workflow routes** (`contentstudio-ai-agents/src/api/routers/workflows/workflow_router.py`):
- `POST /api/v1/workflows/generate-reel` for `video-clips-highlights`
- `GET /api/v1/workflows/{workflow_id}/status`, `GET /api/v1/workflows/{workflow_id}/result`
- `POST /api/v1/workflows/{workflow_id}/cancel`, `POST /api/v1/workflows/{workflow_id}/terminate`

**Laravel pieces, all live:**
- `contentstudio-backend/routes/api/v1.php` — the v1 group and its `['force.json', 'api.key', 'api.request.log', 'throttle:api-v1']` stack.
- `contentstudio-backend/app/Services/AI/AiAgentService.php` — the server-to-server client.
- `contentstudio-backend/app/Services/AI/AiToolCapabilityService.php` — capability feed, cached at 300s.
- `contentstudio-backend/app/Http/Middleware/ApiKeyMiddleware.php:88-108` — API credit check and increment.
- `contentstudio-backend/config/kafka.php:67-70` — `ai-agents.video.job.lifecycle`, `.batch.`, `.transform.`, `.reels.` topics. These are how completion is learned without polling the service ourselves, and are the natural trigger for the webhook events in S-2.

**Gotchas:**
- **Two async engines.** Four tools on Dramatiq, `video-clips-highlights` on Temporal. Different identifier names, different status routes, a separate result route, and status values in caps (`RUNNING`, `COMPLETED`). Normalising these is the core of this story, not an afterthought.
- **Result shapes differ by origin.** Per the comment in `contentstudio-frontend/src/modules/AI-tools/composables/useVideoJobPolling.ts`, dedicated-tool jobs (`motion-control`, `lip-sync`) return a **flat** result with a direct `url` and `tool` key, while chat-originated generation returns a `content[]` array. The status endpoint carries no top-level `source`, so `tool` is the only signal telling them apart. The public API has to normalise this.
- **Seven granular stages, not four.** `src/api/models.py` declares a four-value `JobStatus`, but `ACTIVE_VIDEO_JOB_STATUSES` in the frontend polling composable lists `unknown`, `queued`, `starting`, `enhancing`, `optimizing`, `generating`, `processing`. Do not put these in the public `status`.
- **Credit cost is dynamic.** `contentstudio-frontend/src/composables/useVideoCreditCalculation.ts` holds `VIDEO_PRICING_MATRIX` with per-model, per-resolution API costs, `PRICING_MARKUP = 1.5`, and `CREDIT_COST_PER_SECOND = 0.05`. The estimate endpoint should share this logic rather than reimplementing it, or the two will drift and customers will be quoted one price and charged another.
- The AI service base URL in `config/ai_agents.php` currently points at staging. Confirm production config before launch.

---

## S-2 · [BE] Add video job events to the public webhooks system
**Project:** Web App · **Group:** Backend · **Skill:** Backend · **Product area:** Public API · **Priority:** High · **Type:** Feature

### Description
As an API developer, I want a signed webhook when my video finishes so that I do not have to poll at all.

Adds `video.completed` and `video.failed` to the existing webhooks system. This builds entirely on the public webhooks epic's delivery stack. No new delivery machinery.

### Workflow
1. The user registers a webhook and subscribes to the video events.
2. They send a test event to confirm their endpoint before going live.
3. They submit a video generation request through the API.
4. When it finishes, ContentStudio posts a signed payload to their URL carrying the media ID and final credit cost.
5. If their endpoint is down, delivery retries with backoff and the attempts appear in the delivery log.

### Acceptance criteria
- [ ] `video.completed` and `video.failed` are available as subscribable events.
- [ ] `video.completed` carries the job ID, media ID, media URL, duration, resolution, and the final charged credit cost.
- [ ] `video.failed` carries the job ID and the failure reason.
- [ ] Cancellation emits no event, since the caller initiated it.
- [ ] Both events inherit HMAC signing, at-least-once delivery, exponential-backoff retries into the dead-letter queue, and delivery logs from the existing webhooks system.
- [ ] Both events are available through the existing test-event flow.
- [ ] Delivery metering follows the existing rule of one API request per successful delivery.
- [ ] A workspace with no webhook registered is unaffected, and polling continues to work identically.
- [ ] Webhook delivery is triggered from the existing video lifecycle Kafka events rather than by polling the AI service.

### Impact on existing data
Adds two event types to the webhooks system. No schema change beyond whatever the events table already supports.

### Impact on other products
None.

### Dependencies
- Depends on: **[BE] Add AI video generation endpoints and the async job contract to the public API**
- **Hard dependency on the public webhooks epic**, currently In Review. That epic must land first. Its PRD already scopes non-publishing events as a future phase, which this story fills.

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support — N/A, no interface
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## S-3 · [BE] Add video generation commands to the CLI and agent skill
**Project:** Web App · **Group:** Backend · **Skill:** Backend · **Product area:** Public API · **Priority:** High · **Type:** Feature

### Description
As a terminal user or shell-capable agent, I want video generation commands in the CLI, so that I can generate and publish without leaving the shell.

Terminal users expect to wait and watch, so the CLI blocks with a progress indicator by default rather than handing back a job ID.

### Workflow
1. The user lists available video tools and models.
2. They run an estimate command to see what a generation will cost before committing.
3. They run the generate command. It blocks, showing progress, and prints the media ID when done.
4. Or they pass a no-wait flag, get a job ID immediately, and check it later.
5. They pass the media ID into the existing post creation command.

### Acceptance criteria
- [ ] A `videos` command group exists with `videos:generate`, `videos:tools`, `videos:models`, `videos:estimate`, and `videos:job`.
- [ ] One command exists per dedicated tool: `motion-control`, `lip-sync`, and clips extraction.
- [ ] All 5 tools from S-1 are reachable from the CLI. A tool in the API but missing from the CLI fails this story.
- [ ] Generate blocks by default, rendering progress and stage from the polling response.
- [ ] `--no-wait` returns the job ID immediately without polling.
- [ ] `videos:job <id>` reports status and, when complete, the media ID.
- [ ] Ctrl-C during a blocking generate leaves the job running and prints how to check it later, rather than silently orphaning it.
- [ ] `--json` output includes the media ID in a form that pipes into the post creation command.
- [ ] Commands follow the existing colon syntax, `--json` envelope, and `--dry-run` conventions.
- [ ] Insufficient credits, policy refusal, and rate limit errors each render a readable message rather than a raw HTTP error.
- [ ] `SKILL.md` documents video generation, the estimate-then-generate habit, and the generate-then-publish recipe.

### Impact on existing data
None.

### Impact on other products
CLI package release.

### Dependencies
- Depends on: **[BE] Add AI video generation endpoints and the async job contract to the public API**

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A, CLI story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support — N/A, no interface
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## S-4 · [BE] Expose video generation tools in the MCP server and Claude Desktop bundle
**Project:** Web App · **Group:** Backend · **Skill:** Backend · **Product area:** Public API · **Priority:** High · **Type:** Feature

### Description
As an AI agent operator, I want ContentStudio's video tools available as callable MCP tools that do not stall the conversation, so the agent can produce a video and schedule the post in one exchange.

**The key design decision:** the MCP server waits server-side rather than making the agent poll. Making an agent poll burns context on every status call and makes the conversation look stalled to the user. Short jobs return inline and feel instant; long ones degrade to a handle the agent can check later. This is the deliberate difference from Higgsfield, whose agent-facing video MCP has agents poll.

### Workflow
1. The user connects the ContentStudio MCP server in Claude or another MCP-capable app.
2. They ask for a post with a video.
3. The agent calls the generation tool. The server waits internally.
4. If the video finishes inside the wait, the agent gets it directly and continues.
5. If not, the agent gets a job handle and checks it with a separate tool once the estimated time has passed.
6. The agent calls the existing post tool with the media ID.

### Acceptance criteria
- [ ] One MCP tool exists per video tool from S-1, all 5.
- [ ] A discovery tool exposes available models with their supported modes, resolutions, and durations.
- [ ] An estimate tool returns credit cost before generation, so an agent can check cost before spending.
- [ ] A check-status tool retrieves a job by ID.
- [ ] Generation tools wait server-side up to the bounded-wait cap and return the finished video directly when it completes in time.
- [ ] When the wait elapses, the tool returns a job handle plus the estimated remaining time, so the agent knows when to check rather than polling blindly.
- [ ] Tool descriptions state the credit cost model, so an agent can reason about spend before invoking.
- [ ] Tool schemas derive from the API's capability feed rather than being hardcoded.
- [ ] Generation tools return a media ID in a form the existing post creation tool accepts.
- [ ] Insufficient credits and policy refusals return messages an agent can act on, not raw HTTP errors.
- [ ] The Claude Desktop MCPB bundle exposes the new tools with no separate build step.
- [ ] Verified end to end in Claude Desktop and at least one other MCP-capable client.

### Impact on existing data
None.

### Impact on other products
MCP package release, propagating to Claude Desktop and anywhere else the MCP server is consumed.

### Dependencies
- Depends on: **[BE] Add AI video generation endpoints and the async job contract to the public API**

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support — N/A, no interface
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## S-5 · [BE] Add video generation actions to Zapier, Make, and n8n
**Project:** Web App · **Group:** Backend · **Skill:** Backend · **Product area:** Public API · **Priority:** Medium · **Type:** Feature

### Description
As a no-code automator, I want video generation steps in Zapier, Make, and n8n so that I can wire generation between a trigger and a ContentStudio post action.

**These connectors must never block.** Unlike image, where pinning a fast model kept calls inside platform timeouts, no video model is fast enough. Each connector uses its platform's native resume mechanism instead.

### Workflow
1. The user adds a ContentStudio "Generate video" action to a scenario.
2. They pick a tool and model, enter a prompt, and optionally map an input image.
3. The action returns a job ID immediately.
4. The scenario resumes via the platform's own mechanism, then a check step returns the media ID.
5. They map that media ID into the existing post action.

### Acceptance criteria
- [ ] A "Generate video" submit action and a "Check video job" action exist in all three connectors.
- [ ] No connector action blocks waiting for generation.
- [ ] Each connector uses its platform's idiomatic resume mechanism: n8n's `Wait` node, Make's incomplete-execution handling, and Zapier's polling trigger.
- [ ] Documented example scenarios show the full submit, wait, check, publish chain on each platform.
- [ ] Tool and model selection are pickers rather than free-text fields.
- [ ] Input images can be mapped from a previous step.
- [ ] The check action output includes a media ID that maps into the existing post creation action.
- [ ] The submit action surfaces the estimated credit cost so users see cost before running a scenario at volume.
- [ ] Insufficient credits, policy refusal, and rate limit errors surface as readable messages in each platform's error UI.
- [ ] Each connector passes its platform's review and is published.

### Impact on existing data
None.

### Impact on other products
Three independent connector releases.

### Dependencies
- Depends on: **[BE] Add AI video generation endpoints and the async job contract to the public API**
- Benefits from **[BE] Add video job events to the public webhooks system**, since a webhook-driven trigger is a cleaner resume path than polling on platforms that support it.
- Marketplace review for n8n and Make is the schedule long pole for this epic. Submit early.

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A, backend-only story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support — N/A, no interface
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## S-6 · [BE] Publish public documentation and quickstarts for video generation
**Project:** Web App · **Group:** Backend · **Skill:** Backend · **Product area:** Public API · **Priority:** Medium · **Type:** Chore

### Description
As a developer integrating ContentStudio, I want documentation for the async video contract on every surface, so that I can get from API key to a published post with a generated video without reading source code.

The async contract is the part most likely to be misused, so the documentation carries more weight here than it did for image.

### Acceptance criteria
- [ ] The API reference documents all eight endpoints, all five tools, request and response schemas, and the full error code table.
- [ ] The job state machine is documented with the five stable statuses, and `stage`, `progress`, and `message` are explicitly marked informational and subject to change.
- [ ] A quickstart shows the full loop: estimate, submit, poll, publish.
- [ ] A second quickstart shows the webhook path as an alternative to polling.
- [ ] Recommended polling intervals are documented, with guidance to use `estimated_seconds` rather than a fixed interval.
- [ ] Credit behaviour is documented explicitly: cost varies by model, duration, resolution, and audio; credits are checked at submit and charged on completion; failed and cancelled jobs are not charged.
- [ ] The available models are documented with supported modes, resolutions, durations, and relative cost and speed.
- [ ] Rate limits and the concurrent job cap are documented.
- [ ] The job retention window is documented, including the distinction between a pruned job and an unknown one.
- [ ] It is documented that generated video lands in the workspace media library and counts against storage limits, and that storage is checked at submit.
- [ ] CLI, MCP, Zapier, Make, and n8n each have setup and usage docs for video generation.
- [ ] Connector docs state plainly that video generation cannot complete inside a single blocking step and show the platform's resume pattern.
- [ ] A help centre article covers video generation for non-developer connector users.

### Impact on existing data
None.

### Impact on other products
Public docs site and help centre.

### Dependencies
- Depends on: **[BE] Add AI video generation endpoints and the async job contract to the public API**
- Best written alongside S-2 through S-5 so each surface's docs land with its release.

### Global quality & compliance (wherever applicable)
- [ ] Mobile responsiveness — N/A, documentation story
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support — N/A, no interface
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
