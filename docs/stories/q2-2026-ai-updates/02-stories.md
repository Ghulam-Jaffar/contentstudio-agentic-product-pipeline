# Stories: Q2 2026 AI Updates

---

## Story 1: [BE] Refactor AI agents architecture and platform boundaries

### Description:

As the ContentStudio engineering team, we want the AI agents codebase refactored into clearer architectural boundaries so that AI Studio work is easier to extend, debug, and ship without regressions.

This story covers a structural refactor of the AI agents layer based on the architecture findings captured here:
- Shortcut doc: https://app.shortcut.com/contentstudio-team/write/IkRvYyI6I3V1aWQgIjY5Y2UxYTUwLTgyYTEtNDRlYy1iZDg2LTA1YzY5ZTdhYzY5YyI=
- Local reference: `docs/technical/ai-agents-architecture-and-platform-findings-2026-04-02.md`

The goal is not a product-visible feature. The goal is to reduce coupling between orchestration, model/provider integrations, media generation flows, and shared platform concerns so new AI Studio work can land faster and with less duplication.

---

### Workflow:

1. Developer picks up a new AI Studio change in chat, image, or video generation
2. Developer can locate the correct orchestration, provider, shared utility, and capability modules without tracing through unrelated flows
3. Shared concerns such as request state, provider capability mapping, media settings, and asset consistency are implemented once and reused
4. Future AI Studio changes ship against a cleaner platform structure instead of adding more one-off logic

---

### Acceptance criteria:

- [ ] The refactor follows the architecture directions captured in the linked findings doc
- [ ] AI agent orchestration logic, provider/model capability logic, and shared media generation utilities are separated into clearer modules
- [ ] Duplicated logic across chat, image, and video generation flows is reduced in the touched areas
- [ ] Existing AI Studio chat, image generation, image editing, and video generation flows continue to work after the refactor
- [ ] Logging and debugging paths remain intact or improve for the touched flows
- [ ] The refactor does not introduce user-facing behavior regressions in existing AI Studio tools
- [ ] A short technical note or migration summary is added for the team covering the new structure and extension points

---

### Mock-ups:

N/A - architecture refactor only.

---

### Impact on existing data:

No schema change expected. This story is focused on code structure and shared platform behavior.

---

### Impact on other products:

- Web App: Affects AI Studio internals used by chat, image, and video generation
- Mobile apps: No direct impact
- Chrome extension: No direct impact

---

### Dependencies:

The linked architecture findings doc is the source context for this story.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness - N/A, backend/platform story
- [ ] Multilingual support - N/A, no UI copy in scope
- [ ] UI theming support - N/A, no UI changes in scope
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## Story 2: [FE] Add cancel control for in-progress AI chat generation

### Description:

As an AI Studio user, I want to stop an in-progress AI response so that I stay in control when I submitted the wrong prompt, the response is going in the wrong direction, or I simply want to end generation and try again.

This story adds an industry-standard cancel interaction to AI chat. While a response is streaming, the send button should switch to a stop-style control and cancel the in-flight generation cleanly across frontend and backend.

---

### Workflow:

1. User enters a prompt in AI Studio chat and clicks send
2. While the response is streaming, the send button changes into a stop control with a square/end icon
3. User clicks the stop control
4. Streaming stops immediately and the UI returns to an idle state without hanging loaders
5. User can edit the prompt, send a new message, or continue from the partial response if one already appeared

---

### Acceptance criteria:

- [ ] When a chat response is in progress, the send button is replaced by a stop control
- [ ] The stop control uses a standard stop/end visual pattern rather than a second text CTA
- [ ] Clicking stop cancels the in-flight request or stream without requiring a page refresh
- [ ] The loader state ends immediately after cancel is acknowledged
- [ ] Any partial content already received remains visible instead of disappearing
- [ ] The user can send a new message immediately after cancellation
- [ ] Backend streaming/generation work is terminated or ignored cleanly so it does not continue pushing orphaned events
- [ ] Canceling one response does not break later responses in the same chat session

---

### Mock-ups:

N/A - extend the existing AI Studio chat composer controls using the current send button pattern.

---

### UI Copy

- Stop button tooltip: `Stop generating`
- Optional transient status text after cancel: `Generation stopped`

---

### Impact on existing data:

No schema change expected. This story affects request lifecycle and UI state only.

---

### Impact on other products:

- Web App: AI Studio chat only
- Mobile apps: No impact
- Chrome extension: No impact

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (chat composer controls remain usable on smaller widths)
- [ ] Multilingual support (new stop tooltip and any status copy use i18n)
- [ ] UI theming support (reuse existing themed button/icon patterns)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## Story 3: [FE] Preserve image generation settings when editing AI chat-generated assets

### Description:

As an AI Studio user, I want image edits to keep the generation settings from the original image so that editing a generated asset does not unexpectedly change its output format.

Right now a user can generate an image in chat using a non-square aspect ratio such as `16:9`, ask AI Studio to edit it, and receive an edited image that falls back to `1:1`. This story keeps the original image settings consistent across the edit flow, including aspect ratio and other relevant generation preferences unless the user explicitly changes them.

---

### Workflow:

1. User generates an image in AI Studio chat using a selected model and settings such as `16:9`
2. User asks AI Studio to edit that image
3. The edit flow opens or runs with the original generation settings already preserved
4. The edited result keeps the same aspect ratio and applicable preferences unless the user changes them
5. User receives a consistent edited output instead of an unexpected square image

---

### Acceptance criteria:

- [ ] Editing a generated image preserves the original aspect ratio by default
- [ ] A `16:9` generated image edited through chat remains `16:9` unless the user explicitly changes it
- [ ] The edit flow preserves other applicable generation preferences from the source image when supported by the selected model
- [ ] The UI shows the retained settings correctly before or during edit execution
- [ ] The edit request sent to the backend uses the preserved settings instead of falling back to default values
- [ ] Unsupported settings are ignored gracefully rather than causing the whole edit request to fail
- [ ] Existing image edit flows continue to work for square and non-square images

---

### Mock-ups:

N/A - extends the current AI Studio image edit flow and settings handling.

---

### Impact on existing data:

No schema change expected. This story affects request construction and edit defaults.

---

### Impact on other products:

- Web App: AI Studio image generation/edit flows
- Mobile apps: No impact
- Chrome extension: No impact

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (settings state remains usable in responsive layouts)
- [ ] Multilingual support (no new user-visible copy unless needed)
- [ ] UI theming support (reuse existing themed controls)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## Story 4: [BE] Harden AI chat streaming so responses do not stall after loader starts

### Description:

As an AI Studio user, I want every submitted prompt to either return a streamed response or a clear error so that chat never gets stuck in a silent loading state.

There is an intermittent issue where the user sends text, sees the loader, and then receives no response and no visible failure. This story hardens the event and streaming lifecycle so stalled generations are detected, surfaced, and recoverable instead of silently failing.

---

### Workflow:

1. User sends a prompt in AI Studio chat
2. The UI enters loading state and waits for stream events
3. If the stream succeeds, the response renders normally
4. If the stream breaks, times out, or loses its event chain, the request is marked failed instead of hanging forever
5. User sees a clear retry path and can continue using chat without refreshing the page

---

### Acceptance criteria:

- [ ] Chat requests no longer remain indefinitely in a loader-only state when the response stream breaks
- [ ] Missing, broken, or out-of-order stream events are handled defensively in the affected chat flow
- [ ] A failed stream resolves into a visible error state or retry prompt instead of silence
- [ ] Frontend request state is cleaned up correctly after stream failure
- [ ] Backend logging captures enough context to diagnose why a response did not complete
- [ ] A later prompt in the same chat session still works after one failed response
- [ ] Normal successful streaming behavior remains unchanged

---

### Mock-ups:

N/A - reliability fix in existing AI Studio chat flow.

---

### UI Copy

- Error state: `We couldn't complete this response. Please try again.`
- Retry CTA: `Retry`

---

### Impact on existing data:

No schema change expected. This story affects stream lifecycle handling and observability.

---

### Impact on other products:

- Web App: AI Studio chat
- Mobile apps: No impact
- Chrome extension: No impact

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness - N/A, backend/reliability-focused story with minor existing UI state handling
- [ ] Multilingual support (new failure copy uses i18n if introduced)
- [ ] UI theming support (reuse current error/retry patterns)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## Story 5: [BE] Apply brand fonts in AI chat media generation

### Description:

As an AI Studio user, I want brand fonts from Brand Style to be used consistently in generated chat media so that outputs reflect the same brand identity already applied through logos and colors.

Today brand logos and colors are already making their way into generation, but brand fonts are not being used properly. This story fixes the brand style pipeline so font selections are carried into generation behavior wherever the selected model and workflow support branded text styling.

---

### Workflow:

1. User configures Brand Style with fonts, logo, and colors
2. User generates branded media in AI Studio chat
3. The generation pipeline includes the selected brand font information together with the already-used brand colors and logo context
4. Generated outputs better reflect the configured brand style instead of ignoring the font selection

---

### Acceptance criteria:

- [ ] Brand font settings from the user's brand style are included in the relevant generation context or prompt-building path
- [ ] Font data is passed consistently alongside brand logo and color context in the affected AI chat generation flows
- [ ] Supported generation flows no longer ignore the configured brand font selection
- [ ] If a selected model cannot honor font guidance directly, the system falls back gracefully without breaking generation
- [ ] Existing logo and color behavior continues to work as before
- [ ] The fix applies to the targeted AI chat media generation flows that already consume brand styling

---

### Mock-ups:

N/A - backend behavior fix.

---

### Impact on existing data:

No schema change expected. Reuses existing brand style data.

---

### Impact on other products:

- Web App: AI Studio branded media generation
- Mobile apps: No impact
- Chrome extension: No impact

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness - N/A, backend story
- [ ] Multilingual support - N/A, no new UI copy in scope
- [ ] UI theming support - N/A, no UI changes in scope
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## Story 6: [BE] Bring video generation consistency controls in line with image generation

### Description:

As an AI Studio user, I want video generation to preserve character, product, and attachment consistency just like image generation does so that branded or reference-based videos stay visually aligned across outputs.

The image generation flow already has fixes for character/product/attachment consistency. This story brings the same consistency handling to video generation so reference-aware behavior is not limited to images only.

---

### Workflow:

1. User generates video content in AI Studio using a character, product, or attachment as a reference
2. The video generation pipeline applies the same consistency rules already available in image generation
3. The generated video stays closer to the supplied references across reruns and related outputs
4. Users get more predictable multi-asset creative results between image and video workflows

---

### Acceptance criteria:

- [ ] The video generation flow supports the same character consistency handling already implemented for image generation
- [ ] The video generation flow supports the same product consistency handling already implemented for image generation
- [ ] The video generation flow supports the same attachment/reference consistency handling already implemented for image generation
- [ ] Shared consistency logic is reused where practical instead of reimplemented as a separate one-off path
- [ ] Existing video generation behavior continues to work when no reference assets are supplied
- [ ] The fix applies to AI Studio video generation flows that support referenced creative inputs

---

### Mock-ups:

N/A - backend/platform behavior story.

---

### Impact on existing data:

No schema change expected.

---

### Impact on other products:

- Web App: AI Studio video generation
- Mobile apps: No impact
- Chrome extension: No impact

---

### Dependencies:

May depend on the underlying consistency utilities already used by image generation remaining reusable after the current refactor work.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness - N/A, backend story
- [ ] Multilingual support - N/A, no new UI copy in scope
- [ ] UI theming support - N/A, no UI changes in scope
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## Story 7: [BE] Update existing AI image model option mappings for newly available settings

### Description:

As an AI Studio user, I want existing image models to expose their newly supported options so that I can use the latest provider capabilities without waiting for separate one-off fixes per model.

Several existing image models now support additional settings/options that are not yet reflected properly in our current model mappings. This story updates the model capability layer and request-building paths so available options stay aligned with current provider support.

---

### Workflow:

1. User opens AI image generation in AI Studio
2. User selects an existing image model
3. The system surfaces the newly supported options for that model where applicable
4. The selected settings are passed through correctly during generation and edit flows

---

### Acceptance criteria:

- [ ] Existing image model capability mappings are updated for the newly available provider settings
- [ ] Supported options are exposed only for models that actually support them
- [ ] Unsupported options are not shown or sent for models that do not support them
- [ ] Request payloads use the updated option mappings when users generate or edit images
- [ ] Existing supported options continue to work without regression
- [ ] The update covers the current set of image models whose provider capabilities changed

---

### Mock-ups:

N/A - capability and request-mapping update.

---

### Impact on existing data:

No schema change expected.

---

### Impact on other products:

- Web App: AI image generation
- Mobile apps: No impact
- Chrome extension: No impact

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness - N/A, backend/configuration story
- [ ] Multilingual support - N/A unless new option labels are introduced separately
- [ ] UI theming support - N/A, no design system change in this story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## Story 8: [BE] Update existing AI video model option mappings for newly available settings

### Description:

As an AI Studio user, I want existing video models to expose their newly supported options so that I can use updated provider capabilities without switching to the wrong defaults or missing important controls.

Several existing video models now support additional settings/options that are not yet reflected properly in our current model mappings. This story updates the model capability layer and request-building paths so available options stay aligned with current provider support.

---

### Workflow:

1. User opens AI video generation in AI Studio
2. User selects an existing video model
3. The system surfaces the newly supported options for that model where applicable
4. The selected settings are passed through correctly during video generation

---

### Acceptance criteria:

- [ ] Existing video model capability mappings are updated for the newly available provider settings
- [ ] Supported options are exposed only for models that actually support them
- [ ] Unsupported options are not shown or sent for models that do not support them
- [ ] Request payloads use the updated option mappings during video generation
- [ ] Existing supported options continue to work without regression
- [ ] The update covers the current set of video models whose provider capabilities changed

---

### Mock-ups:

N/A - capability and request-mapping update.

---

### Impact on existing data:

No schema change expected.

---

### Impact on other products:

- Web App: AI video generation
- Mobile apps: No impact
- Chrome extension: No impact

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness - N/A, backend/configuration story
- [ ] Multilingual support - N/A unless new option labels are introduced separately
- [ ] UI theming support - N/A, no design system change in this story
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

## Story 9: [FE] Add resolution and variations controls for supported AI image models

### Description:

As an AI Studio user, I want image resolution and variations controls for models that support them so that I can get the output quality and number of creative options I need without using hidden defaults.

This story adds explicit resolution and variations controls in the AI image generation experience, but only for models that actually support those capabilities.

---

### Workflow:

1. User opens AI image generation
2. User selects an image model
3. If the selected model supports output resolution and/or variations, those controls appear
4. User chooses the desired settings and generates images
5. The output matches the selected resolution and requested variation count where supported

---

### Acceptance criteria:

- [ ] Resolution control is available for image models that support explicit resolution selection
- [ ] Variations control is available for image models that support explicit variations/count selection
- [ ] Models that do not support one or both settings do not show unusable controls
- [ ] Selected resolution and variations values are passed through correctly in generation requests
- [ ] Default values are sensible for supported models when the user does not change them
- [ ] The controls work with existing image generation flows without breaking current settings handling

---

### Mock-ups:

N/A - extend the current AI image generation settings panel using existing controls and patterns.

---

### UI Copy

- Resolution label: `Resolution`
- Resolution helper text: `Choose the output quality supported by this model.`
- Variations label: `Variations`
- Variations helper text: `Choose how many image options to generate in one run.`

---

### Impact on existing data:

No schema change expected. This story adds supported request options only.

---

### Impact on other products:

- Web App: AI image generation
- Mobile apps: No impact
- Chrome extension: No impact

---

### Dependencies:

Depends on the underlying model capability mappings staying current for supported image models.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (controls remain usable in compact settings layouts)
- [ ] Multilingual support (new control labels/helper text use i18n)
- [ ] UI theming support (reuse current themed form controls)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
