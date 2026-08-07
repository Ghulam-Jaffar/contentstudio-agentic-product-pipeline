# Research: AI video generation in the public API

**Date:** 2026-08-03
**Scope:** Expose ContentStudio's AI **video** generation tools across every public developer surface, with polling **and** webhooks. Image generation is a separate, already-specced epic.

**Decided going in:** polling is the baseline contract, webhooks ship in the same epic rather than a later phase. The reasoning is in the async pattern recommendation dated 2026-08-03.

---

## 1. Backlog check

No existing feature or story covers AI video in the public API.

| Slug | Relationship |
|---|---|
| `public-api-ai-image-generation` | **Sibling epic.** Same surfaces, same proxy architecture. Video diverges only on the async contract and dynamic pricing. |
| `public-webhooks` | **Shipped.** Verified in code: `WebhookEventType` enum, `WebhookEmitter`, `WebhookSigner`, `WebhookDeliverer`, `WebhookRepo`, Kafka ingress topic. Already carries 7 event types (`post.scheduled`, `post.published`, `post.failed`, `post.inreview`, `post.approved`, `post.rejected`, `post.comment`), not the 2 its PRD described. Video events are an additive extension, not a dependency. |
| `contentstudio-public-cli-agent-skills` | Structural precedent. Shipped the CLI, skill, and MCP server this epic extends. |
| `video-clips-drive-dropbox-upload`, `video-clips-respect-duration-and-count`, `runtime-bitrate-check` | Web-app clip work. No API scope. |

---

## 2. The 5 video tools in scope

| # | Tool key | What it does | Service entry point | Async engine |
|---|---|---|---|---|
| 1 | `text-to-video` | Generate video from a prompt | `POST /api/v1/jobs/video` with `generation_mode: text-to-video` | Dramatiq |
| 2 | `image-to-video` | Animate a still image | Same, `generation_mode: image-to-video` | Dramatiq |
| 3 | `video-clips-highlights` | Extract clips/highlights from a longer video | `POST /api/v1/workflows/generate-reel` | **Temporal** |
| 4 | `motion-control` | Apply motion direction to generation | `POST /api/v1/tools/motion-control` | Dramatiq |
| 5 | `lip-sync` | Sync lip movement to audio | `POST /api/v1/tools/lip-sync` | Dramatiq |

There is also a third generation mode, `reference-to-video`, available on the same job endpoint. It is a mode rather than a tool card in AI Studio, and should be exposed as a mode on the generate endpoint.

**Models:** 25 video models in `contentstudio-ai-agents/src/utils/model_registry.py` (`veo3.1`, `sora-2-pro`, `kling-video-v3-pro`, `seedance-2.0`, `wan-2.7`, `minimax-hailuo-2.3-pro` and others). Each model declares which of `text-to-video`, `image-to-video`, and `reference-to-video` it supports, so mode support is per model, not universal.

---

## 3. Finding: there are two async engines, not one

This is the most consequential structural finding.

**Dramatiq jobs** — `src/api/routers/jobs/jobs_router_dramatiq.py`:
- `POST /jobs/video`, `POST /jobs/video/batch`
- `GET /jobs/{job_id}`, `GET /jobs/{job_id}/events` (SSE), `DELETE /jobs/{job_id}`, `GET /jobs/`

**Temporal workflows** — `src/api/routers/workflows/workflow_router.py`:
- `POST /workflows/generate-reel`
- `GET /workflows/{workflow_id}/status`, `GET /workflows/{workflow_id}/result`
- `POST /workflows/{workflow_id}/cancel`, `POST /workflows/{workflow_id}/terminate`

Four of the five tools run on Dramatiq. `video-clips-highlights` runs on Temporal, with a different identifier (`workflow_id` not `job_id`), different status route, a separate result route, and different status values (`RUNNING`, `COMPLETED` in caps).

**Implication:** the public API must present **one** job resource and normalise both engines behind it. Exposing two async contracts because we happen to have two internal queues would be leaking implementation into a public contract that we then cannot change.

---

## 4. Finding: the enqueue response is already well shaped

`JobEnqueueResponse` in `jobs_router_dramatiq.py` already returns:

```python
success, job_id, status_url, events_url, message, estimated_seconds, status
```

`estimated_seconds` is particularly useful. It gives us a per-request completion estimate we can return to callers, which is exactly what the bounded-wait design and the connector guidance need. No new work required to obtain it.

`JobInfo` (the status response) returns `id`, `status`, `progress` (0 to 100), `message`, `error`, `result`.

---

## 5. Finding: video credit cost is dynamic, unlike image

Image generation is a flat one credit per image. Video is not.

From `contentstudio-frontend/src/composables/useVideoCreditCalculation.ts`:

```ts
const PRICING_MARKUP = 1.5           // 150% of API cost
const CREDIT_COST_PER_SECOND = 0.05  // dollars per credit-second

export const VIDEO_PRICING_MATRIX: Record<string, VideoModelConfig> = { ... }
```

Cost is computed from **model, duration, resolution, and whether audio is enabled**, against a per-model, per-resolution API cost matrix with a baseline.

**Implication:** the API cannot just say "video costs N credits". Callers need to know the cost **before** they submit, or an automated pipeline will burn a customer's balance unpredictably. This argues for a cost estimate endpoint, and for returning the final charged cost on the completed job.

---

## 6. Finding: the public job states must not mirror the internal ones

`src/api/models.py` declares a four-value `JobStatus` enum: `pending`, `processing`, `completed`, `failed`.

But the real stream emits seven active stages. From `ACTIVE_VIDEO_JOB_STATUSES` in `contentstudio-frontend/src/modules/AI-tools/composables/useVideoJobPolling.ts`:

```
unknown, queued, starting, enhancing, optimizing, generating, processing
```

And Temporal adds `RUNNING` / `COMPLETED` in a different case convention.

`enhancing` and `optimizing` are pipeline internals. Putting them in a public contract means every integrator's branching breaks when the pipeline changes. The recommendation is a stable coarse set (`queued`, `processing`, `completed`, `failed`, `cancelled`) with the granular value carried in a separate, explicitly non-contractual `stage` field for display.

---

## 7. What the web app does today

`useVideoJobPolling.ts` polls `/jobs/{id}`. It does **not** consume the SSE stream, even though `GET /jobs/{id}/events` exists.

That is the strongest available signal on what works in production. Polling is proven; SSE is built but unexercised. It argues for keeping SSE internal until an integrator actually asks for it, rather than committing it to a public contract we would then have to support.

The polling layer also reveals that dedicated-tool jobs (`motion-control`, `lip-sync`) return a **flat** result with a direct `url` and `tool` key, while chat-originated generation returns a `content[]` array. Per the code comment, the status endpoint carries no top-level `source`, so `tool` is the only signal distinguishing them. The public API needs to normalise these into one result shape.

---

## 8. Surfaces to cover

| Surface | This epic adds |
|---|---|
| REST API | Video domain, unified job resource, cost estimate |
| Webhooks | `video.completed` and `video.failed` events on the webhooks epic's delivery stack |
| CLI | `videos:*` commands, blocking with progress by default |
| Agent skill | Video generation and the generate-then-publish recipe |
| MCP server | Video tools with server-side bounded wait |
| Claude Desktop | Inherited via the MCPB bundle |
| Zapier / Make / n8n | Submit action plus a wait or check step, never blocking |

---

## 9. Constraints

**Latency.** Minutes, not seconds, and highly variable by model, duration, and resolution. No surface can block for the full duration. `estimated_seconds` from the enqueue response is the tool for managing expectations.

**File size.** Video assets are far larger than images. Landing them in the workspace media library, as the image epic does, interacts much harder with workspace storage limits.

**Connector timeouts.** Zapier, Make, and n8n all impose request timeouts far shorter than a video generation. Unlike image, where pinning a fast model was a mitigation, for video no model is fast enough. Those connectors must use their native resume mechanisms: n8n's `Wait` node, Make's incomplete-execution handling, Zapier's polling triggers.

**Cost of failure.** A failed video costs real money upstream. Credit behaviour on failure and cancellation needs an explicit decision, not a default.

---

---

## Brand knowledge

Every AI generation surface in the web app carries a brand on/off toggle. Generation conditioned on brand knowledge is the default experience, so an API that ignores brand would produce output visibly worse than the UI's.

**The Brand Knowledge Revamp has shipped.** Verified in `contentstudio-frontend/src/modules/publisher/ai-content-library/types/index.ts:133-150`:

```ts
// v2 on/off flag; backend defaults true and gates brand application in generation/chat.
brand_enabled?: boolean
brand_style?: BrandStyle
brand_profile?: BrandProfileInfo
brand_voice?: BrandVoiceV2
brand_topics?: Array<{ name: string; description: string }>
// Legacy multi-item arrays — removed by brand-knowledge:strip-legacy; optional during transition.
styles?: StyleOption[]
brand_voices?: BrandVoiceOption[]
```

Three consequences for the API:

1. **One brand per workspace.** All the brand fields are singular and the legacy arrays are marked removed. There is no brand for a caller to choose between, so the API needs no brand ID.
2. **`brand_enabled` defaults to true and already gates brand application in generation and chat.** So the correct API default for an omitted brand parameter is to honour that stored flag, which means a workspace with brand knowledge set up gets on-brand API output with no extra parameter. This is parity, not a new rule.
3. **The internal contract is already boolean-shaped.** `contentstudio-frontend/src/api/composer.ts:428-429` carries `use_brand_voice?: boolean` next to `brand_voice_id?: string | null` annotated *"unused — backend uses the workspace default"*, and returns `brand_voice_applied: boolean`. The public API mirrors this as `use_brand` and `brand_applied`.

**Existing brand endpoints** (`contentstudio-frontend/src/modules/publisher/config/api-utils.ts`), all internal: `profile/get`, `profile/setBrandEnabled`, `profile/updateBrandSection`, `profile/sources/{add,delete,sync,enrich}`, `profile/delete`, `analyzeBrand`.

**CRUD stays out of the public API by product decision.** Brand knowledge is created and maintained in the web app. The API consumes it. The only public brand surface is a read-only status endpoint so a caller can tell whether brand knowledge exists before relying on it.

`contentstudio-frontend/src/modules/publisher/ai-content-library/composables/useBrandKnowledgePresence.ts` already computes "does this workspace have brand knowledge" and should back that status endpoint rather than a second implementation.

**Video-specific trap:** `VideoJobRequest` accepts a free-form `brand_guidelines: dict[str, Any]`. This must **not** be exposed publicly. Letting callers hand-construct guidelines creates a second source of truth for the brand and produces API output that diverges from the app, which is the opposite of the parity requirement.

## 10. Recommended split

Six stories. S-1 deep, the rest thin.

| # | Story |
|---|---|
| S-1 | `[BE] Add AI video generation endpoints and the async job contract to the public API` |
| S-2 | `[BE] Add video job events to the public webhooks system` |
| S-3 | `[BE] Add video generation commands to the CLI and agent skill` |
| S-4 | `[BE] Expose video generation tools in the MCP server and Claude Desktop bundle` |
| S-5 | `[BE] Add video generation actions to Zapier, Make, and n8n` |
| S-6 | `[BE] Publish public documentation and quickstarts for video generation` |

No frontend, mobile, or design story. Nothing changes in the web app.
