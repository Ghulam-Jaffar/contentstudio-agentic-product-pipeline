# Research: Add Nano Banana 2 & Gemini 3.1 Flash Image Models

## Current State

ContentStudio AI agents already support similar predecessor models:
- **`nano-banana-pro`** → `fal-ai/nano-banana-pro` (with edit: `fal-ai/nano-banana-pro/edit`)
- **`gemini-25-flash-image`** → `fal-ai/gemini-25-flash-image` (with edit: `fal-ai/gemini-25-flash-image/edit`)

Both are classified as **aspect_ratio models** (use `aspect_ratio` param, not `image_size`). The frontend lists them under the "Google" provider group with full aspect ratio support.

## New Models to Add

| Model Key | FAL Endpoint | Edit Endpoint | Notes |
|---|---|---|---|
| `nano-banana-2` | `fal-ai/nano-banana-2` | `fal-ai/nano-banana-2/edit` | Google's next-gen fast image gen/edit |
| `gemini-3.1-flash-image-preview` | `fal-ai/gemini-3.1-flash-image-preview` | `fal-ai/gemini-3.1-flash-image-preview/edit` | Same model, alternate endpoint name |

**Parameters (both models):** aspect_ratio, num_images, seed, output_format (png), safety_tolerance (1-4), enable_web_search, thinking_level — all follow the same aspect_ratio pattern as existing Gemini/Nano-Banana models.

## What Needs to Change

### AI Agents Backend (`contentstudio-ai-agents/`)

1. **`src/utils/model_registry.py`** — Add both models with `provider: "fal"`, `model_id`, and `edit_model_id`
2. **`src/agents/image/image_generator.py`** — Add both to `FAL_MODELS` dict (~line 138-144). The `uses_aspect_ratio` detection (~line 1108) already matches on `"nano-banana"` and `"gemini"` substrings, so both new models are auto-detected.
3. **`src/api/models.py`** — `ImageModel` enum is stale (only has old models) and is only used by the legacy REST endpoint, NOT the streaming API. The streaming API passes model strings directly. **No change needed** here unless we want to update the legacy enum.
4. **Edit flow** (~line 2332-2334) — Already matches on `"gemini"` and `"nano-banana"` substrings, so edit routing is auto-handled.

### Frontend (`contentstudio-frontend/`)

5. **`src/modules/composer_v2/composables/useImageGeneration.js`** — Add both models to the Google provider group (~line 357), following the same structure as `nano-banana-pro` and `gemini-25-flash-image` (value, label, credits, maxImages, description, generationTime, supportedAspectRatios).
6. **`src/locales/{lang}/ai_tools.json`** — Add translation keys for model labels and descriptions in all locale directories (en, ar, de, es, fr, it, ja, ko, nl, pt, ru, tr, zh).

## Files Involved

| File | Change |
|---|---|
| `contentstudio-ai-agents/src/utils/model_registry.py` | Add 2 model entries |
| `contentstudio-ai-agents/src/agents/image/image_generator.py` | Add 2 entries to FAL_MODELS dict |
| `contentstudio-frontend/src/modules/composer_v2/composables/useImageGeneration.js` | Add 2 models to Google provider |
| `contentstudio-frontend/src/locales/*/ai_tools.json` | Add translation keys for model labels/descriptions |
