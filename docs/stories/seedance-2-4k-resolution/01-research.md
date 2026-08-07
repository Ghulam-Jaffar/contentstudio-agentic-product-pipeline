# Research: Seedance 2.0 4K resolution support

Seedance 2.0 supports 4K output on fal. We only expose 480p, 720p, and 1080p. This is about adding 4K and charging for it correctly.

Kept separate from `docs/stories/video-model-reference-audio-video-inputs/` on purpose. Same model family, different capability, no shared code path beyond the model definition file.

## Backlog check

- `docs/stories/runtime-bitrate-check/` touches video output settings but is about bitrate, not resolution.
- `docs/stories/ai-studio-seedream-5-models/` is image models, unrelated despite the similar name.
- No existing story covers Seedance resolutions or 4K. Net-new.

## Current state

### What fal supports versus what we expose

Fetched from the fal OpenAPI schema for `bytedance/seedance-2.0/text-to-video`:

| Parameter | fal schema | Our registry |
|---|---|---|
| `resolution` | `480p`, `720p`, `1080p`, **`4k`**, default `720p` | `["480p", "720p", "1080p"]`, default `720p` |
| `duration` | `auto`, then `4` through `15`, default `auto` | `["4"` through `"15"]`, no `auto` |
| `aspect_ratio` | `auto`, `21:9`, `16:9`, `4:3`, `1:1`, `3:4`, `9:16`, default `auto` | the six explicit ratios, no `auto`, default `16:9` |
| `generate_audio` | bool, default true | `supports_audio: True` |
| `bitrate_mode` | `standard`, `high`, default `standard` | not present |

So `4k` is the gap in scope here. The missing `auto` values for duration and aspect ratio, and `bitrate_mode`, are separate observations noted at the bottom, not part of this work.

Note the resolution string fal uses is lowercase **`4k`**. Seedance's other keys are lowercase (`480p`, `720p`, `1080p`) so it fits the existing pattern, but this is per-model: MiniMax H3 uses uppercase `2K` and `768P`. The key has to match fal exactly, and the reason that matters is in the next section.

`contentstudio-ai-agents/src/utils/model_registry.py`, `seedance-2.0` entry from line 1547:

- `supported_resolutions: ["480p", "720p", "1080p"]`
- `api_cost: { "480p": 0.1346, "720p": 0.3034, "1080p": 0.6804 }` (USD per second)
- `default_resolution: "720p"`
- The `mode_requirements` blocks for image-to-video and reference-to-video carry their own `supported_resolutions` overrides with the same three values, so mode-level lists need updating too, not just the model level.

`seedance-2.0-fast` is capped at 480p and 720p with a code comment saying "1080p dropped, confirmed", so it is out of scope for 4K.

### The credit path will silently undercharge if 4K is exposed before the cost is added

This is the part worth attention. `contentstudio-frontend/src/composables/useVideoCreditCalculation.ts` resolves a per-second cost like this (around line 388):

```
} else if (apiCost[resolution] !== undefined) {
    costPerSecond = apiCost[resolution] as number
} else {
    // Fixed pricing (single resolution model)
    const availableResolution = Object.keys(apiCost)[0]
    costPerSecond = apiCost[availableResolution] as number
}
```

If the requested resolution has no entry in `api_cost`, it falls back to **the first key in the object**. For `seedance-2.0` that first key is `480p` at $0.1346 per second. So a 4K generation with no `4k` cost entry would be priced at 480p rates, silently, with no error and no log. Given 4K's real cost is roughly $2.72 per second (derivation below), that is around a **20x undercharge per second**.

Credits themselves derive automatically from that per-second figure: markup is applied, multiplied by duration, divided by the credit unit cost, then `Math.ceil`. So the fix is genuinely just "put the right number in both places", but the failure mode if that is missed is expensive and invisible.

The same silent fallback applies to any future resolution added to any model, which is why the second story exists.

`ModelApiCostResolution` (line 39) declares `'480p'`, `'720p'`, `'1080p'` explicitly plus an index signature `[key: string]: ResolutionCost | undefined`. So adding `'4k'` will not be a type error, and TypeScript will not catch a typo or a casing mismatch either.

### Deriving the 4K cost

The registry documents Seedance's pricing formula itself, in the comment above `seedance-2.0-fast`:

> Pricing from the same token formula: `(h * w * duration * 24) / 1024` tokens at $0.0112/1000.

Applying that to the standard tier and solving for the rate from the known 1080p figure:

- 1080p is 1920 x 1080, so `(1920 * 1080 * 24) / 1024` = 48,600 tokens per second
- $0.6804 / 48.6 = **$0.014 per 1000 tokens** for the standard tier, exact
- Sanity check at 480p (854 x 480): 9,608 tokens/s x $0.014 = $0.1345, registry says $0.1346, matches within rounding

For 4K at 3840 x 2160:

- `(3840 * 2160 * 24) / 1024` = 194,400 tokens per second
- 194.4 x $0.014 = **$2.7216 per second**, which is exactly 4x the 1080p rate
- At the 5 second default that is about $13.61 in API cost for one video

This is a derivation from our own documented formula, not a quoted figure. It should be confirmed against fal's published 4K rate before shipping, and the assumed 3840 x 2160 pixel dimensions should be confirmed too since fal only says `4k`.

The magnitude matters for product, not just engineering: at 4x the 1080p rate a single 15 second 4K video costs roughly $41 in API spend. Whether 4K should be available on all plans, or gated, is a product decision this research cannot make.

### Drift found while checking

`useVideoCreditCalculation.ts` prices `seedance-2.0-fast` at `1080p: 0.5443`, but the model registry lists that model's `supported_resolutions` as `["480p", "720p"]` with a comment saying 1080p was dropped and confirmed. The frontend carries a price for a resolution the backend says does not exist. Harmless today because the resolution cannot be selected, but the two lists have drifted and one of them is wrong.

## What needs to change

1. Add `4k` to Seedance 2.0's supported resolutions at model level and in the image-to-video and reference-to-video mode overrides, after confirming fal accepts `4k` on those endpoints too. The schema fetched was text-to-video only.
2. Add the 4K per-second cost to `api_cost` in the AI agents registry and to the pricing matrix in the frontend, matching fal's resolution key casing exactly.
3. Confirm the 4K cost against fal's published rate rather than shipping the derived figure.
4. Expose 4K in the resolution selector for Seedance 2.0 only, leaving other models untouched.

The silent fallback described above is not being fixed as its own piece of work. It stays a known trap, which is why the story asks for the charged credits to be verified on a real 4K generation rather than assumed from the config.

## Files involved

**AI agents (`contentstudio-ai-agents/`)**

- `src/utils/model_registry.py`, `seedance-2.0` entry from line 1547
- `src/utils/video/cost_calculator.py`
- `src/utils/credit_validation/validation.py`
- `tests/unit/` video cost and capability tests

**Frontend (`contentstudio-frontend/`)**

- `src/composables/useVideoCreditCalculation.ts`, `seedance-2.0` at line 289, `ModelApiCostResolution` at line 39, resolution lookup around line 388
- `src/composables/useVideoGeneration.ts`, Seedance model definitions and `supportedResolutions`
- `src/modules/AI-tools/components/MediaGenerationOptions.vue`, resolution selector
- `src/locales/*/ai_tools.json` if any new copy is needed

## Out of scope, noted for later

- Seedance 2.0 accepts `auto` for both duration and aspect ratio. We send explicit values only, so users cannot let the model choose.
- Seedance 2.0 accepts `bitrate_mode` of `standard` or `high`. We do not send it. Adjacent to `docs/stories/runtime-bitrate-check/`.
- The `seedance-2.0-fast` 1080p price drift described above.

## Reference

fal model page: https://fal.ai/models/bytedance/seedance-2.0
