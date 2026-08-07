# Research: FLUX.3 video models

Black Forest Labs' FLUX.3 video family on fal. Eleven endpoints covering five generation modes, each with a full-quality and a draft tier, plus a promotion step.

## Backlog check

- `docs/stories/video-model-reference-audio-video-inputs/` adds reference video and audio inputs to existing models. Different work, but it touches the same files, so the two will collide in `model_registry.py`, `parameter_optimizer_util.py`, and `useVideoGeneration.ts`. Worth sequencing rather than running both at once.
- `docs/stories/seedance-2-4k-resolution/` touches the same credit-cost surface.
- `docs/stories/ai-studio-image-audio-to-video-tool/` adds dedicated one-shot tools for other providers. Unrelated.
- No existing story covers FLUX.3 or any Black Forest Labs video model. Net-new.

Black Forest Labs is already a provider for us, but images only: `flux_pro`, `flux_2`, `flux_2_pro` in the model registry. FLUX.3 would be their first video model in our stack.

## Endpoint inventory

Eleven endpoints resolve to five modes at two tiers, plus one promotion step:

| Mode | Full | Draft |
|---|---|---|
| text-to-video | `blackforestlabs/flux-3/text-to-video` | `.../text-to-video/draft` |
| image-to-video | `blackforestlabs/flux-3/image-to-video` | `.../image-to-video/draft` |
| first-last-frame-to-video | `blackforestlabs/flux-3/first-last-frame-to-video` | `.../first-last-frame-to-video/draft` |
| keyframes-to-video | `blackforestlabs/flux-3/keyframes-to-video` | `.../keyframes-to-video/draft` |
| extend-video | `blackforestlabs/flux-3/extend-video` | `.../extend-video/draft` |

Plus `blackforestlabs/flux-3/draft-enhance`, which has no draft variant because it is the promotion step.

## Verified schemas

Pulled from the fal OpenAPI schemas, not from the marketing copy.

**Common to every full-quality generation mode:** `prompt` (required), `aspect_ratio` (`auto`, `21:9`, `2:1`, `16:9`, `4:3`, `1:1`, `3:4`, `9:16`, default `auto`), `resolution` (`720p`, `1080p`, default `720p`), `duration` (`auto` or 5 through 20 seconds), `generate_audio` (bool, default **true**), `safety_tolerance` (0 to 4, default 2). Output is `video` plus `seed`.

| Mode | Mode-specific inputs |
|---|---|
| text-to-video | none |
| image-to-video | `image_url` (required, PNG / JPEG / WebP) |
| first-last-frame-to-video | `start_image_url` and `end_image_url`, both required, PNG / JPEG / WebP. `duration` here is an integer 5 to 20 with default 5, not `auto` |
| keyframes-to-video | `keyframes`, a required array of 1 to 10 objects, each carrying `image_url` (PNG, JPEG, WebP) and `frame_index`. `frame_index` is described as the "frame position of this keyframe in the generated 24 fps video. Must be unique and at most `duration * 24`", minimum 0. `duration` is an integer 5 to 20, default 5. Sorting is not required. There is no fps input, 24 fps is fixed |
| extend-video | `video_url` (required). Accepted input formats mp4, mov, webm, m4v, gif. Max input length, max file size, and how much footage is added are all unstated |

**Draft tier:** identical inputs to its full counterpart **minus `resolution`**. Draft quality is fixed. Output is `video` plus `draft_cache`, described by fal as a "durable encrypted cache bundle" and returned as a file with a URL.

**draft-enhance:** takes `draft_cache_url` (required) and optional `safety_tolerance`. Nothing else. No prompt, no seed, no duration, no resolution. Output is `video` only. Always renders at full-quality 1080p, and fal states "synchronized audio is included at no extra cost".

## Pricing, per second of generated video

| Endpoint | 720p | 1080p |
|---|---|---|
| text-to-video | $0.17 | $0.29 |
| image-to-video | $0.17 | $0.29 |
| first-last-frame-to-video | $0.17 | $0.29 |
| keyframes-to-video (full) | assumed $0.17 / $0.29, **not verified** | |
| extend-video | **$0.41** | **$0.53** |
| any draft | **$0.06** (720p, fixed) | n/a |
| draft-enhance | n/a | **$0.29** (always 1080p, audio included) |

Two things stand out. Extend-video costs roughly 2.4x the other full modes on the same model. And drafts are cheap enough to change behavior, at about a third to a seventh of a full render.

## The draft economics, which are counter-intuitive

Draft at $0.06/s then enhance at $0.29/s totals **$0.35 per second**. Going straight to full 1080p is **$0.29 per second**.

So **enhancing a draft costs about 21 percent more than simply rendering at full quality in the first place.** The draft path is not a cheaper route to the same output.

Where drafts do pay off is iteration:

| Scenario | Cost per second |
|---|---|
| One full 1080p render, right first time | $0.29 |
| One draft, then enhance | $0.35 (21% more) |
| Three drafts, enhance the best one | $0.47 |
| Three full renders to get one you like | $0.87 (46% more than the draft route) |

The break-even is at roughly one discarded attempt. A user who nails it first time is better off going direct. A user who iterates at all is better off drafting.

This matters for how we present the feature. Framing drafts as "the cheap option" would be wrong and would cost users money. The honest framing is "try ideas cheaply, then commit", and the enhance action has to show its credit cost before the user commits to it.

It also matters because the draft path always ends at 1080p (enhance has no resolution parameter), so a user who wants 720p output cannot get there via a draft.

## What is precedented in our stack and what is not

| Mode | Precedent |
|---|---|
| text-to-video | Standard, every video model has it |
| image-to-video | Standard |
| first-last-frame-to-video | **Already a first-class mode.** Veo 3.1 / Fast / Lite use it, with `requires_first_frame`, `requires_last_frame`, `first_frame_param_name`, `last_frame_param_name` in `model_registry.py` and `startEndFrameSupport: { bothRequired: true }` in `useVideoGeneration.ts`. FLUX.3 names its params `start_image_url` and `end_image_url` rather than Veo's `first_frame_url` and `last_frame_url`, but since the param names are already configurable per model that is a declaration, not a code change |
| keyframes-to-video | **No precedent.** An ordered array of objects each with an image and a frame index is unlike anything we send today. Reference images are an unordered flat list, and first-last-frame is two fixed slots |
| extend-video | **No precedent in the general video path.** Video input exists today only in the dedicated one-shot tools (motion-control, lip-sync). The general generation path has no video input at all |
| draft and draft-enhance | **No precedent anywhere.** Two-phase generation, a persisted cache handle, and a second billable step against one logical output |

## Two structural problems this model creates

**1. Per-mode pricing.** `api_cost` in `model_registry.py` is keyed by resolution at **model** level, and `VIDEO_PRICING_MATRIX` in `useVideoCreditCalculation.ts` mirrors that shape. FLUX.3 breaks the assumption: extend-video is $0.41 / $0.53 while text-to-video is $0.17 / $0.29 on the same model, and drafts are $0.06 flat. Either the cost lookup gains a mode dimension, or FLUX.3 is registered as several registry keys. Registering it as several keys is the smaller change but fragments the model in the picker, where users expect one FLUX.3 entry.

**2. The silent undercharge fallback.** Documented in `docs/stories/seedance-2-4k-resolution/`: `useVideoCreditCalculation.ts` around line 388 falls back to the first key in `api_cost` when the requested resolution has no entry, with no error. With FLUX.3's spread of rates on one model, a missed key means charging draft rates for an extend-video render, a roughly sevenfold undercharge. This is not being fixed as its own work, so FLUX.3's cost entries need verifying against real generations.

## Other gaps FLUX.3 exposes

- **`auto` for duration and aspect ratio.** We send explicit values everywhere and support `auto` nowhere. Same gap noted on Seedance 2.0. Not blocking, but we lose the model's own judgement.
- **Durations up to 20 seconds.** Everything else in our stack caps at 15. At $0.29/s a 20 second 1080p render is $5.80 in API cost, and a 20 second extend-video is $10.60. These will be the most expensive single generations we offer.
- **`safety_tolerance`.** A 0 to 4 parameter we do not model at all. Sending the default of 2 is the obvious starting point.
- **Audio is generated by default and does not affect price.** Unlike Veo, which has audio-on and audio-off pricing tiers, FLUX.3 bundles it. Simplifies the cost matrix.

## Keyframes are positional, not just ordered

Worth calling out separately, because it changes the interface a lot. `frame_index` is an absolute frame number in a fixed 24 fps timeline, not an ordinal. Frame 0 is the first frame, frame `duration * 24` is the last. A 5 second clip has frames 0 to 120, a 20 second clip has 0 to 480.

Three consequences:

- **The user is placing images in time, not sorting a list.** Drag-to-reorder is the wrong metaphor. The natural control is a timeline where an image is dropped at a moment.
- **Frame numbers should never reach the user.** They think in seconds. The interface should work in seconds and convert, since seconds x 24 is the frame index.
- **Changing duration invalidates keyframes.** The ceiling is `duration * 24`, so shortening a clip from 20 seconds to 10 leaves any keyframe past the 10 second mark out of range. This is the duration-change case the story has to handle, and now it is concrete rather than hypothetical.

Also note the minimum is 1 keyframe, not 2. A single keyframe pinned at frame 0 is effectively image-to-video, and pinned anywhere else it is a shot that has to arrive at a given image at a given moment. Whether to allow 1 is a product call rather than a technical limit.

Uniqueness is enforced per keyframe, so no two images can share a frame index. If the interface snaps placement to whole seconds, that caps the usable keyframes at one per second, which only bites on clips shorter than 10 seconds.

## Still unverified

- Full-quality keyframes-to-video pricing
- Extend-video input constraints: max clip length, max file size, and how much footage is appended
- Draft cache lifetime. fal calls it "durable" without stating a TTL
- Whether the enhanced output's duration and aspect ratio always match the draft

## Files involved

**AI agents (`contentstudio-ai-agents/`)**

- `src/utils/model_registry.py`
- `src/utils/video/parameter_optimizer_util.py`
- `src/agents/video/video_parameter_optimizer.py`
- `src/utils/video/cost_calculator.py`
- `src/utils/credit_validation/validation.py`
- `src/api/routers/jobs/jobs_router_dramatiq.py`
- `src/api/routers/streaming_router.py`
- `src/locales/en/en.json` plus the other 5 locales

**Frontend (`contentstudio-frontend/`)**

- `src/composables/useVideoGeneration.ts`
- `src/composables/useVideoCreditCalculation.ts`
- `src/components/dashboard/ChatInput.vue`
- `src/modules/AI-tools/components/MediaGenerationOptions.vue`
- `src/modules/AI-tools/utils/buildChatStreamPayload.ts`
- `src/locales/*/ai_tools.json` (8 locales)

## Reference

- https://fal.ai/models/blackforestlabs/flux-3/text-to-video
- https://fal.ai/models/blackforestlabs/flux-3/image-to-video
- https://fal.ai/models/blackforestlabs/flux-3/first-last-frame-to-video
- https://fal.ai/models/blackforestlabs/flux-3/keyframes-to-video
- https://fal.ai/models/blackforestlabs/flux-3/extend-video
- https://fal.ai/models/blackforestlabs/flux-3/draft-enhance
