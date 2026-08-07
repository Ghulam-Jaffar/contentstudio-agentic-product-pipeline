# Research: Seedance 2.5 via direct API

Adding ByteDance's Seedance 2.5, accessed directly through BytePlus ModelArk rather than through fal. The API was officially released on 7 August 2026.

## Read this first

**1. Every video model we have is fal, and there is no provider seam.** This is the real cost of "not from fal", and it is far bigger than adding a model entry.

**2. BytePlus bills video by tokens, not by the second.** Our entire cost model is per-second USD. This is a structural mismatch, not a new number to paste in.

**3. The official option matrix still needs to be read by someone with access.** The official documentation links are below, but `docs.byteplus.com` is fully client-side rendered, so its content cannot be extracted programmatically. The Lark documents are permissioned. Everything in the "still to confirm" section needs a human with the docs open.

## Official sources

- Prompt guide: https://bytedance.larkoffice.com/docx/A88jd0B47oAd8zxWp5ycZFMfnxh
- Model one-pager: [External] Dreamina-Seedance-2.5 Brochure, https://bytedance.larkoffice.com/file/D6j4b38i3odYhcxRfa3cSNVxnIh
- Tutorial series: https://docs.byteplus.com/en/docs/ModelArk/2607688
- Pricing: https://docs.byteplus.com/en/docs/ModelArk/1544106
- Create a video generation task: https://docs.byteplus.com/en/docs/ModelArk/1520757
- Video Generation API index: https://docs.byteplus.com/en/docs/ModelArk/Video_Generation_API
- Model list: https://docs.byteplus.com/en/docs/ModelArk/1330310

## Confirmed: commercial terms

From the official release note:

- **Activating Seedance 2.5 requires the account to hold an available $30 resource package.** Resource packages already bought for the 2.0 series are sufficient to activate 2.5. This is an account prerequisite, not a code dependency, and it blocks any testing until it is in place.
- **Seedance 2.0 mini and Seedance 2.0 fast are discounted from 14:00 UTC+8 on 7 August 2026 until 7 September 2026.** Discounts apply automatically.
  - Seedance 2.0 mini at 40 percent of list, from about **$0.032 per second** at 720p
  - Seedance 2.0 fast at 75 percent of list, from about **$0.09 per second** at 720p

### An observation worth acting on separately

We currently buy Seedance 2.0 Fast through fal at **$0.2419 per second at 720p** (`api_cost` in `src/utils/model_registry.py`). BytePlus direct is about **$0.09 per second** discounted, implying roughly **$0.12 at list**. That is around half the fal rate for what is very likely the same model.

Seedance 2.0 mini is not in our registry at all, and at roughly $0.032 per second discounted it would be by a distance the cheapest video model we could offer.

Neither belongs in this story. Both are a strong argument for building the provider seam properly rather than as a one-off for 2.5, and worth their own ticket once the seam exists. Confirm the comparison is like for like first, since platform, quotas and output specs may differ.

## Confirmed: the ModelArk platform

Two facts established from the official documentation's own navigation and search summaries.

**Video is billed per token, not per second.** ModelArk quotes video models in USD per million output tokens. For reference, the Seedance 1.0 family lists at 2.5 USD per million tokens for Pro, 1.0 for Pro Fast and 1.8 for Lite. This lines up with the token formula already recorded in our registry comment for the fal Seedance entries, `(h * w * duration * 24) / 1024` tokens, which is how fal derives its per-second rates in the first place.

Consequence for us: `api_cost` in `src/utils/model_registry.py` and `VIDEO_PRICING_MATRIX` in `contentstudio-frontend/src/composables/useVideoCreditCalculation.ts` are both per-second-USD-by-resolution. A direct BytePlus model either needs its per-second equivalents derived from the token rate at each resolution and duration, or the cost model needs to understand token pricing. Deriving per-second rates per resolution is the smaller change and matches how the fal entries were built.

**Task lifecycle operations exist, including cancellation.** ModelArk's video generation API documents four operations: create a video generation task, retrieve a video generation task, list video generation tasks, and cancel or delete a video generation task. That answers the open question from the first pass: in-flight cancellation is supported, so Seedance 2.5 does not need special-case cancel behavior. The exact contract still needs reading.

## Verified: the codebase side

### The video path is hardwired to fal

| Where | What it does |
|---|---|
| `src/agents/video/video_generator.py:23` | `import fal_client` at module level |
| `src/agents/video/video_generator.py:96` | `handler = fal_client.submit(model_endpoint, arguments=parameters)`, then `handler.get()` blocks until done |
| `src/agents/video/video_generator.py:107-118` | Pulls `handler.cancel_url` and `handler.request_id` and stores them via `store_fal_cancel_url` / `store_fal_request_id` |
| `src/jobs/cancellation.py:235-262` | `fal_client.status_async` and `fal_client.cancel_async` |
| `src/jobs/cancellation.py:381-387` | Maps fal status objects (`InProgress`, `Completed`, `InQueue`) onto our status strings |
| `src/utils/video/media_probe.py:7` | `ensure_fal_credentials` |

The `provider` field exists on every registry entry, but for video it is metadata only. Nothing branches on it. The only place `provider` selects behavior is `src/agents/image/models_registry.py:18`, which filters image models to `provider == "fal"`.

Registry provider counts today: 47 fal, 4 openai, 2 anthropic, 2 groq, 1 google. Every non-fal entry is a text or image model. **Seedance 2.5 would be the first non-fal video model.**

### What that means concretely

Submitting, polling, cancelling and probing media all assume fal's client and its handler object. A direct provider needs its own path for each, plus something to route between them. Whether that is a proper provider interface or a branch is an engineering call, but the seam does not exist and has to be created.

ModelArk's create / retrieve / list / cancel shape is a reasonable fit for the existing Dramatiq flow in `src/jobs/`, which publishes progress to Redis and lifecycle events to Kafka via `src/events/publisher.py`. Polling retrieve keeps the current shape. If ModelArk also offers callbacks, using them would mean a public endpoint and reconciliation with that flow, which is more work for a first integration.

## Still to confirm from the official docs

Nobody should build against guesses here. Someone with the docs open needs to fill in:

1. Exact model IDs for Seedance 2.5 across its modes
2. Endpoint URL, HTTP method and auth scheme
3. Supported durations, resolutions and aspect ratios, with exact parameter names and allowed values
4. Whether audio is generated, and whether it affects price
5. Reference input types and limits for reference-to-video
6. The token rate for Seedance 2.5 at each resolution, so per-second equivalents can be derived
7. Rate limits and concurrency caps, which fal previously absorbed for us
8. The exact cancel semantics, and what happens to billing for a cancelled task

An earlier pass collected figures from third-party aggregator and press pages. Those have been removed rather than left in the document, because having plausible unofficial numbers sitting next to official links is how wrong numbers get built. The one thing worth carrying forward from them is a warning: two third-party sources disagreed on resolution, one reporting 480p and 720p and another claiming native 4K, so do not assume the resolution ladder without checking.

## Why this model stresses our existing assumptions

**Long durations.** Seedance 2.5 is reported to generate substantially longer clips than anything we carry. Our longest today is FLUX.3 at 20 seconds and everything else caps at 15. Duration selectors and any 15 or 20 second assumptions need checking against the confirmed maximum.

**Token billing versus per-second costing.** Covered above. This is the main cost-model impact.

**The silent undercharge fallback.** `useVideoCreditCalculation.ts` around line 388 falls back to the first key in `api_cost` when the requested resolution has no entry, with no error and nothing logged. Documented in `docs/stories/seedance-2-4k-resolution/`. Any new model with several rates needs its charges verified against real generations.

**Reference video and audio inputs.** Seedance 2.5's reference-to-video mode accepts reference videos and audio. `docs/stories/video-model-reference-audio-video-inputs/` is already scoped to build exactly that mechanism for MiniMax H3 and roll it out to WAN and Seedance 2.0. Seedance 2.5 should consume it rather than build a second one.

## Related backlog

- `docs/stories/video-model-reference-audio-video-inputs/` builds the reference video and audio mechanism this model will need.
- `docs/stories/flux-3-video-models/` hits the same per-mode pricing limitation and also adds a long-duration model.
- `docs/stories/seedance-2-4k-resolution/` covers Seedance 2.0 on fal. Same model family, different provider.

All four touch `model_registry.py`, `useVideoGeneration.ts` and `useVideoCreditCalculation.ts`. They will conflict if run concurrently.

## Files involved

**AI agents (`contentstudio-ai-agents/`)**

- `src/agents/video/video_generator.py`
- `src/jobs/cancellation.py`
- `src/utils/video/media_probe.py`
- `src/utils/model_registry.py`
- `src/utils/video/parameter_optimizer_util.py`
- `src/utils/video/cost_calculator.py`
- `src/utils/credit_validation/validation.py`
- `src/api/routers/jobs/jobs_router_dramatiq.py`
- `src/utils/config.py` for the new credentials
- `src/agents/image/providers/` as the closest existing example of a provider module

**Frontend (`contentstudio-frontend/`)**

- `src/composables/useVideoGeneration.ts`
- `src/composables/useVideoCreditCalculation.ts`
- `src/modules/AI-tools/components/MediaGenerationOptions.vue`
