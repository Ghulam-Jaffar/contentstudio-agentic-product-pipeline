# Stories: Add New Image Generation Models

---

## Story 1: Add Nano Banana 2 image generation model (text-to-image + edit)

### Description:
As a ContentStudio user, I want to generate and edit images using Google's Nano Banana 2 model so that I can access faster, higher-quality image generation with Google's latest state-of-the-art model.

---

### Workflow:

1. User opens the AI Image Generator from the Composer
2. User clicks the model selector dropdown
3. Under the "Google" provider group, user sees "Nano Banana 2" listed alongside existing Google models
4. User selects "Nano Banana 2" and sees supported aspect ratios update accordingly
5. User enters a prompt and clicks generate — images are generated using the `fal-ai/nano-banana-2` endpoint
6. User can also use image-to-image editing with this model — when editing an existing image, the system routes to the `fal-ai/nano-banana-2/edit` endpoint

---

### Acceptance criteria:

- [ ] `fal-ai/nano-banana-2` is registered in `contentstudio-ai-agents/src/utils/model_registry.py` with `provider: "fal"`, `model_id: "fal-ai/nano-banana-2"`, and `edit_model_id: "fal-ai/nano-banana-2/edit"`
- [ ] `nano-banana-2` is added to the `FAL_MODELS` dict in `contentstudio-ai-agents/src/agents/image/image_generator.py` mapping to `"fal-ai/nano-banana-2"`
- [ ] Text-to-image generation works with the `nano-banana-2` model via the streaming API
- [ ] Image-to-image editing works with the `nano-banana-2` model (routes to `/edit` endpoint)
- [ ] Aspect ratio parameter is used (not `image_size`) — consistent with existing nano-banana/gemini models
- [ ] `nano-banana-2` is added to the Google provider group in `contentstudio-frontend/src/modules/composer_v2/composables/useImageGeneration.js` with value `nano-banana-2`, appropriate credits, maxImages, generationTime, and supportedAspectRatios (`21:9`, `16:9`, `3:2`, `4:3`, `5:4`, `1:1`, `4:5`, `3:4`, `2:3`, `9:16`)
- [ ] Translation keys added for model label (`ai_tools.image_generation.models.nano_banana_2`) and description (`ai_tools.image_generation.model_descriptions.nano_banana_2`) in all locale directories under `src/locales/`
- [ ] Model appears in the model selector dropdown under the Google provider
- [ ] Selecting the model updates the available aspect ratios in the UI
- [ ] User preference is saved when selecting this model (persisted via `ai_image_model` preference)

---

### Mock-ups:
N/A — follows existing model selector UI pattern. New model entry appears in the Google provider group of the nested model dropdown in `MediaGenerationOptions.vue`.

---

### Impact on existing data:
- No schema changes. New model is additive — registered in model_registry and FAL_MODELS dict.
- No migration needed. Existing user preferences for other models are unaffected.

---

### Impact on other products:
- **Mobile apps:** No impact — AI image generation is web-only.
- **Chrome extension:** No impact — image generation is not in the extension.
- **White-label:** No impact — model selection UI uses existing themed components.

---

### Dependencies:
None — the FAL.ai endpoint `fal-ai/nano-banana-2` must be available (it is a public FAL partner model).

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only) — N/A, image generation modal is already responsive
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---
---

## Story 2: Add Gemini 3.1 Flash Image Preview model (text-to-image + edit)

### Description:
As a ContentStudio user, I want to generate and edit images using Google's Gemini 3.1 Flash Image Preview model so that I can access the latest Gemini image generation capabilities with fast inference.

---

### Workflow:

1. User opens the AI Image Generator from the Composer
2. User clicks the model selector dropdown
3. Under the "Google" provider group, user sees "Gemini 3.1 Flash Image" listed alongside existing Google models
4. User selects "Gemini 3.1 Flash Image" and sees supported aspect ratios update accordingly
5. User enters a prompt and clicks generate — images are generated using the `fal-ai/gemini-3.1-flash-image-preview` endpoint
6. User can also use image-to-image editing with this model — when editing an existing image, the system routes to the `fal-ai/gemini-3.1-flash-image-preview/edit` endpoint

---

### Acceptance criteria:

- [ ] `fal-ai/gemini-3.1-flash-image-preview` is registered in `contentstudio-ai-agents/src/utils/model_registry.py` with `provider: "fal"`, `model_id: "fal-ai/gemini-3.1-flash-image-preview"`, and `edit_model_id: "fal-ai/gemini-3.1-flash-image-preview/edit"`
- [ ] `gemini-31-flash-image-preview` is added to the `FAL_MODELS` dict in `contentstudio-ai-agents/src/agents/image/image_generator.py` mapping to `"fal-ai/gemini-3.1-flash-image-preview"`
- [ ] Text-to-image generation works with the `gemini-31-flash-image-preview` model via the streaming API
- [ ] Image-to-image editing works with the `gemini-31-flash-image-preview` model (routes to `/edit` endpoint)
- [ ] Aspect ratio parameter is used (not `image_size`) — consistent with existing gemini models
- [ ] `gemini-31-flash-image-preview` is added to the Google provider group in `contentstudio-frontend/src/modules/composer_v2/composables/useImageGeneration.js` with value `gemini-31-flash-image-preview`, appropriate credits, maxImages, generationTime, and supportedAspectRatios (`1:1`, `3:4`, `4:3`, `9:16`, `16:9`)
- [ ] Translation keys added for model label (`ai_tools.image_generation.models.gemini_31_flash_image`) and description (`ai_tools.image_generation.model_descriptions.gemini_31_flash_image`) in all locale directories under `src/locales/`
- [ ] Model appears in the model selector dropdown under the Google provider
- [ ] Selecting the model updates the available aspect ratios in the UI
- [ ] User preference is saved when selecting this model (persisted via `ai_image_model` preference)

---

### Mock-ups:
N/A — follows existing model selector UI pattern. New model entry appears in the Google provider group of the nested model dropdown in `MediaGenerationOptions.vue`.

---

### Impact on existing data:
- No schema changes. New model is additive — registered in model_registry and FAL_MODELS dict.
- No migration needed. Existing user preferences for other models are unaffected.

---

### Impact on other products:
- **Mobile apps:** No impact — AI image generation is web-only.
- **Chrome extension:** No impact — image generation is not in the extension.
- **White-label:** No impact — model selection UI uses existing themed components.

---

### Dependencies:
None — the FAL.ai endpoint `fal-ai/gemini-3.1-flash-image-preview` must be available (it is a public FAL partner model).

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only) — N/A, image generation modal is already responsive
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
