# Research: audio and video reference inputs for video models

MiniMax H3 was recently added. Its reference-to-video endpoint accepts reference **videos** and reference **audio** alongside reference images, and we handle none of that. The same gap exists on other models that support the same inputs.

## Backlog check

- `docs/stories/ai-studio-image-audio-to-video-tool/` is adjacent but different: a new dedicated one-shot tool using `ltx-2.3/audio-to-video` and `longcat-single-avatar/image-audio-to-video`. That story adds a new tool, this one adds inputs to existing video models. No overlap.
- `docs/stories/ai-studio-image-to-video-one-shot/`, `ai-studio-seedream-5-models/`, `video-clips-*` are unrelated.

Nothing in the backlog covers reference audio or video inputs. Net-new.

## Current state

### The registry declares these inputs and nothing reads them

`contentstudio-ai-agents/src/utils/model_registry.py` already carries capability keys for audio and video inputs on four models. Every one of them is commented as scaffolding:

| Model (registry key) | Mode | Declared keys | Comment in code |
|---|---|---|---|
| `wan-2.7` (line 302) | image-to-video | `supports_video_input` / `video_input_param_name: video_url`, `supports_audio_input` / `audio_input_param_name: audio_url` | "scaffold only (declarative; not yet sent to fal)" |
| `wan-2.7` (line 329) | reference-to-video | `supports_reference_video`, `reference_video_param_name: reference_video_urls` | "scaffold only (reference_video_urls, MP4/MOV, max 100 MB)" |
| `wan-2.6` (line 376) | image-to-video | `supports_audio_input` / `audio_input_param_name: audio_url` | "scaffold only" |
| `wan-2.6` (line 387) | reference-to-video | `requires_reference_videos`, `video_refs_param_name: video_urls`, `max_videos: 3` | "Video-ref input path not wired yet (structure only)" |
| `seedance-2.0` (line 1595) | reference-to-video | `supports_video_input: video_urls`, `supports_audio_input: audio_urls` | none |
| `seedance-2.0-fast` (line 1676) | reference-to-video | `supports_video_input: video_urls`, `supports_audio_input: audio_urls` | none |
| `minimax-h3` (line 553) | reference-to-video | **nothing** | fal supports both, we declare neither |

A grep for every one of those keys across `src/` returns hits only in `model_registry.py` itself. Nothing consumes them. They are documentation, not behavior.

### The parameter mapper has no audio or video path at all

`contentstudio-ai-agents/src/utils/video/parameter_optimizer_util.py` is where user inputs become fal params. Its signature accepts only `source_image_url`, `reference_image_urls`, `front_image_url`, and `end_image_url`. There is no argument for an audio or video input, so even a correctly declared registry key has nothing to flow through.

Two specific consequences:

- **Line 122**: `if not reference_image_urls or len(reference_image_urls) < 2: raise ValueError("reference-to-video mode requires at least 2 reference images")`. Per the fal schema, reference images are not mandatory when reference videos are supplied. This hard gate would reject a valid video-only reference request.
- **Lines 147 to 152**: the image list param name is discovered by probing `mode_reqs` for `image_urls`, `reference_image_urls`, or `input_image_urls`. Any audio or video equivalent needs the same treatment or it silently lands nowhere.

### MiniMax H3 registry entry versus the real fal schema

Fetched from the fal OpenAPI schema for `minimax/h3/reference-to-video`:

| Parameter | fal schema | In our registry? |
|---|---|---|
| `prompt` | required, 1 to 7000 chars. Assets are referenced in the prompt as "Image 1", "Video 1", "Audio 1" | model-level `max_prompt_chars: 2000`, see the note below |
| `reference_image_urls` | array, max 9 | yes, `max_images: 9` |
| `reference_video_urls` | array, max 3 videos, each 2 to 15 seconds, combined 15 seconds max | **no** |
| `reference_audio_urls` | array, max 3 clips, each 2 to 15 seconds, combined 15 seconds max, and audio cannot be the only reference input | **no** |
| all reference types combined | max 12 files total | **no** |
| `duration` | int, default 5, range 5 to 15 | yes |
| `resolution` | `768P` or `2K`, default `2K` | yes |
| `aspect_ratio` | `adaptive`, `21:9`, `16:9`, `4:3`, `1:1`, `3:4`, `9:16`, default `adaptive` | yes |

Two things the schema does **not** document: accepted container formats and maximum file size for reference video and audio. Those need confirming against fal at build time. Our own registry records wan-2.7 reference video as MP4 or MOV up to 100 MB, which is a reasonable starting assumption but is not the same model.

**Secondary finding worth checking:** `minimax-h3` carries `max_prompt_chars: 2000` at model level, while the fal r2v schema allows 7000. If the 2000 was taken from the text-to-video endpoint, reference-to-video prompts are being truncated or rejected roughly 3.5x earlier than fal requires. This matters more than usual here, because H3 expects the prompt to describe how each numbered asset is used, so reference prompts are naturally longer than text-to-video prompts.

### The prompt is part of the contract

This is the part most likely to be missed. H3 does not infer what to do with each file. The prompt has to name them: "Image 1 walks into the room while Audio 1 plays". So shipping the upload slots without surfacing which file is Image 1 versus Video 1 versus Audio 1 gives the user no way to write a working prompt. Any UI has to label the slots.

### Frontend

- Reference-to-video is never named in the frontend. A grep for `reference-to-video`, `reference_image_urls`, `reference_video`, or `reference_audio` across `contentstudio-frontend/src/` returns nothing.
- The mode is reached implicitly: `useVideoGeneration.ts` declares `minimax-h3` with `imageToVideoSupport.maxImages: 9` and `startEndFrameSupport: { bothRequired: false, allowReferences: true }` (lines 1184 to 1213). Attaching several images is what routes a request to the r2v endpoint.
- Attachments are images only. The chat attachment path in `components/dashboard/ChatInput.vue` hardcodes `type: 'image/jpeg'` at three separate points (lines 1616, 1632, 1753). There is no audio or video attach affordance for video generation.
- `getAudioButtonState()` in `useVideoGeneration.ts` (line 1539) is about generated **output** audio, not audio input. For `minimax-h3` and the `wan` family it returns `{ show: true, toggleable: false }` because fal gives no switch to turn generated audio off. It is unrelated to this work, and the naming is a trap: do not extend it for input audio.
- `MediaGenerationOptions.vue` lists `minimax-h3` (line 550) in the model picker.

### Laravel backend

No involvement. A grep for `minimax` across `contentstudio-backend/` hits only an unrelated sample-data seeding command. Video generation goes from the frontend to the AI agents service, and media URLs come from the existing upload path, so this work does not touch Laravel.

### Existing test coverage

`contentstudio-ai-agents/tests/unit/test_minimax_h3_capabilities.py` has 6 tests, including `test_h3_r2v_uses_reference_image_urls_no_prompt_optimizer` and `test_registry_h3_declares_capabilities`. They lock in the image-only behavior, so they will need extending rather than replacing. `tests/unit/test_reference_to_video_validation.py` covers the existing r2v validation path.

## Which models accept audio or video input

The answer is not "only WAN, Seedance, and MiniMax", because there are two separate paths through this codebase and only one of them is broken.

### Path A: dedicated one-shot tools, where audio and video inputs already work

`contentstudio-ai-agents/src/tools/dedicated/` has a full input-kind system. `specs.py:20` declares `InputKind = Literal["image_url", "video_url", "audio_url", "text", "enum", "number", "bool"]`, and `models.py` carries per-tool request schemas with real `video_url` and `audio_url` fields, validation, and input aliases (for example `video_url` also accepts `driving_video_url`, `motion_video_url`; `audio_url` also accepts `voice_url`, `speech_url`).

Three shipped models use it:

| Model | Mode | Inputs |
|---|---|---|
| `kling-v3-motion-control` | motion-control | image plus a driving video |
| `sync-lipsync-v3` | lip-sync | video plus audio |
| `omnihuman-v1.5` | audio-to-video | image plus audio |

Two more are planned in `docs/stories/ai-studio-image-audio-to-video-tool/`: `ltx-2.3/audio-to-video` and `longcat-single-avatar/image-audio-to-video`. Neither is in the model registry yet.

So the product already accepts audio and video inputs today. Just not on the general video generation path, and not through the registry's `mode_requirements` declarations. Note that `requires_video` and `requires_audio` in the registry are declared on all three of these models and are read by nothing either. The dedicated tools work because their own request schemas carry the fields explicitly, bypassing the registry.

This is worth reusing rather than reinventing: the input-kind vocabulary, the alias handling, and the validation patterns in `src/tools/dedicated/models.py` are the closest existing precedent for what the general path needs.

### Path B: general video generation, where they do not

Twelve models declare a `reference-to-video` mode: `wan-2.7`, `wan-2.6`, `minimax-h3`, `kling-video-o1`, `kling-video-v1.6-standard`, `kling-video-v1.6-pro`, `kling-video-v3-standard`, `kling-video-v3-pro`, `veo3.1-fast`, `veo3.1`, `seedance-v1-lite`, `seedance-2.0`, `seedance-2.0-fast`.

Of those, only the WAN, Seedance 2.0, and MiniMax families accept reference video or audio as far as we can tell:

- **MiniMax H3**: verified against the fal schema. Accepts `reference_video_urls` and `reference_audio_urls`.
- **WAN 2.7, WAN 2.6, Seedance 2.0, Seedance 2.0 Fast**: declared in our own registry, unverified against fal schemas, unwired.
- **Kling Video O1**: verified against the fal schema for `fal-ai/kling-video/o1/reference-to-video`. **Images only.** Accepts `prompt`, `image_urls`, `elements`, `duration`, `aspect_ratio`, and nothing else. No video or audio input of any kind.
- **Kling v1.6 Standard and Pro, Kling v3 Standard and Pro, Veo 3.1, Veo 3.1 Fast, Seedance v1 Lite**: not verified. Kling O1 is the newest Kling reference endpoint and it is images-only, which makes it unlikely the older Kling reference modes accept more. Veo 3.1's reference support is image based. Still worth confirming rather than assuming, which is why the rollout story carries an explicit audit item.

Also worth noting: Veo 3.1, Veo 3.1 Fast, and Veo 3.1 Lite have a `first-last-frame-to-video` mode, which is image based and unrelated to audio or video input.

## What needs to change

1. Declare reference video and reference audio support for `minimax-h3` in the registry, with the real fal limits (3 each, 2 to 15 seconds per file, 15 seconds combined per type, 12 files total across all reference types, audio never alone).
2. Give the parameter mapper actual audio and video input arguments, and make it read the capability keys that already exist rather than ignoring them.
3. Relax the "at least 2 reference images" gate so a video-driven reference request is valid.
4. Validate before spending credits: counts, per-file duration, combined duration, total file count, the audio-not-alone rule, format, and size. A rejected fal call still costs latency and a confusing error.
5. Do the same for the other models that declare these inputs: `wan-2.7`, `wan-2.6`, `seedance-2.0`, `seedance-2.0-fast`. The registry already says what each one accepts and under which param name, and the names differ per model (`audio_url` versus `audio_urls`, `video_url` versus `video_urls` versus `reference_video_urls`), so the mapper must be param-name driven, not hardcoded.
6. Frontend: let users attach reference videos and audio for models that support them, label the slots so prompts can reference them, and validate client-side with clear messages.
7. Confirm the accepted formats and max file size per model against the fal schema, since the OpenAPI schema does not state them.

## Files involved

**AI agents (`contentstudio-ai-agents/`)**

- `src/utils/model_registry.py`
- `src/utils/video/parameter_optimizer_util.py`
- `src/agents/video/video_parameter_optimizer.py`
- `src/utils/credit_validation/validation.py`
- `src/utils/video/cost_calculator.py`
- `src/api/routers/jobs/jobs_router_dramatiq.py`
- `src/api/routers/streaming_router.py`
- `src/locales/en/en.json` and the other 5 locales
- `tests/unit/test_minimax_h3_capabilities.py`, `tests/unit/test_reference_to_video_validation.py`

**Frontend (`contentstudio-frontend/`)**

- `src/composables/useVideoGeneration.ts`
- `src/composables/useVideoCreditCalculation.ts`
- `src/components/dashboard/ChatInput.vue`
- `src/modules/AI-tools/components/MediaGenerationOptions.vue`
- `src/modules/AI-tools/utils/buildChatStreamPayload.ts`
- `src/locales/*/ai_tools.json` (8 locales)

## Reference

fal model page: https://fal.ai/models/minimax/h3/reference-to-video
