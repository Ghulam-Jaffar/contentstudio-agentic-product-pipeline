# AI-Powered Composer Customization - Stories

## [BE] Add AI Customize caption generation for Composer

### Description
As a social media manager, I want ContentStudio to generate platform-specific captions from one base caption or post idea so that I can quickly publish tailored content across every selected network.

---

### Workflow

1. User selects multiple social accounts in Composer.
2. User either writes a base caption or opens the AI Customize dialog and describes what the post is about.
3. User chooses whether to use brand voice when a default brand voice is available.
4. User clicks Generate, Regenerate all platforms, or Regenerate just the active platform.
5. User receives one tailored caption per requested platform, with each caption kept within that platform's caption limit.
6. If some platforms fail, user still receives successful captions for the other platforms and can retry failed platforms.
7. If the request cannot be completed, user receives a recoverable error state instead of raw backend or LLM details.

---

### Acceptance criteria

- [ ] A backend endpoint accepts `workspace_id`, `selected_platforms`, `language`, `use_brand_voice`, optional `brand_voice_id`, optional `base_caption`, optional `intent`, and optional `target_platform`.
- [ ] Request validation requires at least one of `base_caption` or `intent`; if both are empty, the response tells the frontend that user intent is required.
- [ ] Batched generation returns a structured map of platform keys to captions for every successfully generated selected platform.
- [ ] Single-platform regeneration returns only the requested platform caption and does not require regenerating other platforms.
- [ ] Captions are generated intelligently for each platform's expected style, not by copying one generic caption into every platform.
- [ ] Captions respect these platform limits: Facebook 63,206 characters, Instagram 2,200, Pinterest 500, X/Twitter 280, LinkedIn 3,000, Threads 500, TikTok 2,200, YouTube 5,000, and Bluesky 300.
- [ ] Facebook generation uses practical guidance of 500-1,500 characters unless the input clearly needs a longer caption.
- [ ] If the LLM returns an over-limit caption, the backend truncates it to the platform limit and flags that platform as truncated so the frontend can show a notice.
- [ ] A batched generation consumes 1 AI text credit when at least one platform caption succeeds.
- [ ] A single-platform regeneration consumes 1 AI text credit when the requested platform caption succeeds.
- [ ] If the workspace has no AI text credits available, no LLM request is made and the response indicates the existing credit-exhausted frontend pattern should be shown.
- [ ] When `use_brand_voice` is true and a brand voice exists, the backend injects the selected/default brand voice into the AI generation context.
- [ ] When `use_brand_voice` is false, the backend generates without brand voice context.
- [ ] When `use_brand_voice` is true but no usable brand voice exists, generation still works without brand voice context and the response identifies that brand voice was unavailable.
- [ ] Total LLM failure returns no generated captions and includes a retryable user-facing error state.
- [ ] Partial platform failure returns successful platform captions plus a list of failed platforms and retryable user-facing error copy for those platforms.
- [ ] Blocked or inappropriate content returns user-facing copy equivalent to: "We couldn't generate captions. Try rephrasing your input."
- [ ] Timeout handling stops the request after 30 seconds and returns a retryable error state.
- [ ] Malformed LLM output is logged and converted into either partial failure or total failure, depending on whether any valid platform captions were recovered.
- [ ] Backend logs include workspace, requested platforms, generation mode, brand voice on/off, success/failure state, and credit deduction outcome, without logging full caption content as sensitive analytics payload.

---

### Mock-ups

N/A for backend. UI screenshots will be attached to the FE stories in Shortcut.

---

### Impact on existing data

No new persistent user-facing schema is required for v1. Existing workspace AI text credit usage is updated when generation succeeds. AI-generated caption history may reuse the existing AI generated content logging/history pattern if engineering chooses to keep these generations auditable.

---

### Impact on other products

Web Composer only. AI functionality is not added to iOS or Android because ContentStudio mobile apps do not support AI features yet. Chrome extension, Bulk Schedule, Recycle Posts, Campaigns, First Comment, carousel content, hashtags-as-separate-field, and AI media are out of scope.

---

### Dependencies

None.

---

### Global quality & compliance

- [ ] Mobile responsiveness (frontend only, N/A for backend-only story)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used; N/A for backend-only story)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Implementation references

*Pointers from research - not a contract. Engineering may choose a different approach.*

**Primary entry points:**
- `contentstudio-backend/routes/web.php` - existing planner/composer AI routes include `POST /planner/fetchAiCaption`.
- `contentstudio-backend/app/Http/Controllers/Planner/HelperController.php` - current `fetchAiCaption` handles AI caption requests, credit checks, LLM calls, history logging, and credit updates.
- `contentstudio-backend/app/Repository/Ai/AiContentLibrary/AiContentLibraryProfileRepo.php` - existing profile/brand voice lookup pattern.

**Existing patterns:**
- Existing AI caption generation posts to `env('LUMOTIVE_CAPTION_API') . 'caption_generation_gpt'`.
- Existing AI text credits use `caption_generation_credit`.
- Existing credit helpers include `PlanHelper::checkAvailableAICredits` and `PlanHelper::deductTextCredits`.

**Suggested names:**
- Endpoint: `POST /planner/aiCustomizeCaptions`.
- Request field names: `base_caption`, `intent`, `selected_platforms`, `language`, `use_brand_voice`, `brand_voice_id`, `target_platform`.
- Response fields: `captions`, `failed_platforms`, `truncated_platforms`, `limits`, `credit_full`, `brand_voice_applied`, `retryable`.

**Gotchas:**
- The current AI caption helper calculates text credit usage by word count. This feature requirement is explicit: 1 AI text credit per generation/regeneration.
- The PRD wins over the Shortcut epic description: AI media suggestions are out of scope for v1.

---

## [FE] Add AI Customize control and empty-caption dialog to Composer

### Description
As a social media manager, I want AI options attached to the existing Customize control so that I can choose between manual platform customization and AI-assisted caption generation without leaving Composer.

---

### Workflow

1. User opens Composer and selects more than one social account.
2. User sees a grouped Customize control in the existing bottom-right footer location.
3. User can turn Customize on manually and use the existing per-platform tabs without triggering AI.
4. User opens the AI dropdown from the left side of the grouped control.
5. If the caption box has text, user can start AI generation directly from the dropdown.
6. If the caption box is empty, user sees a focused dialog asking what the post is about.
7. If the workspace has a default brand voice, user sees the brand voice option enabled by default.
8. If no brand voice exists, user sees a Set up action that takes them to Brand Knowledge.
9. User can dismiss the dialog or submit it from keyboard or buttons.

---

### Acceptance criteria

- [ ] The current Customize switch remains in the same Composer footer position beside the existing editor toolbar controls.
- [ ] The Customize control is rendered as one grouped control: AI dropdown caret on the left, vertical divider in the middle, existing Customize toggle on the right.
- [ ] Clicking the Customize toggle ON still splits Composer into selected platform tabs and copies the base caption into empty platform captions exactly as it does today.
- [ ] Clicking the Customize toggle ON does not automatically trigger AI generation.
- [ ] When only one platform is selected, the AI dropdown is hidden or disabled and the existing Customize-disabled behavior is preserved.
- [ ] When Customize is OFF, the grouped control uses a neutral surface and the AI dropdown primary item says "Generate per-platform captions".
- [ ] When Customize is ON before AI generation, the grouped control uses the active tinted surface and the AI dropdown primary item says "Generate per-platform captions".
- [ ] After AI-generated content exists, the AI dropdown primary item says "Regenerate all platforms" and shows a secondary item "Regenerate just [Platform]".
- [ ] The active platform name in "Regenerate just [Platform]" visually uses that platform's brand color.
- [ ] Dropdown helper tip reads: "Toggle Customize without AI to manually write per-platform captions."
- [ ] The Customize hover tooltip reads: "Different captions per platform · AI available".
- [ ] The AI dropdown closes on outside click, Escape, or item selection.
- [ ] The AI dropdown is keyboard navigable with up/down arrows, Enter, and Escape.
- [ ] If the user clicks Generate with an empty caption, a centered modal dialog opens instead of sending a generation request.
- [ ] Dialog title reads: "Generate captions with AI".
- [ ] Dialog description reads: "Tell us what your post is about and we'll write a tailored caption for each selected platform."
- [ ] Topic field label reads: "What's this post about?"
- [ ] Topic placeholder reads: "e.g. Launching our new summer collection - bright colors, breezy fabrics, available from June 1st".
- [ ] Generate button is disabled while the topic field is empty or whitespace-only.
- [ ] Dialog footer hint reads: "Uses 1 AI text credit".
- [ ] Dialog footer buttons read "Cancel" and "Generate".
- [ ] Dialog can be dismissed by Cancel, Escape, or clicking the backdrop.
- [ ] Dialog can be submitted with Generate or Cmd/Ctrl + Enter.
- [ ] When brand voice exists, the dialog shows a "Use brand voice" option with its toggle defaulted ON.
- [ ] Brand voice info tooltip reads: "AI will use your default brand voice".
- [ ] Brand voice toggle state does not persist across generations; every new dialog defaults ON when a brand voice exists.
- [ ] When no brand voice exists, the dialog replaces the toggle with a "Set up" action.
- [ ] No-brand-voice helper text reads: "No brand voice set up yet. Add one to make AI captions match your tone consistently."
- [ ] Clicking "Set up" closes the dialog and navigates to the existing Brand Knowledge page.
- [ ] The platform reminder strip lists every selected platform that will receive captions.
- [ ] New copy is i18n-ready and not hardcoded in templates.
- [ ] The UI uses available `@contentstudio/ui` components where practical: `Button`, `Dropdown`, `DropdownItem`, `Modal` or `Dialog`, `Textarea`, `Switch`, `ActionIcon`, and `Icon`.
- [ ] The implementation uses theme-aware classes/CSS variables for white-label support and does not hardcode primary brand colors except platform identity colors where platform branding is required.
- [ ] The modal remains usable at mobile-width web viewports with readable copy and tappable controls.

---

### Mock-ups

Attach the relevant Composer control and dialog screenshots to this story in Shortcut after story creation. Required visual references include Customize OFF, Customize ON, AI dropdown before generation, AI dropdown after generation, dialog with brand voice, dialog without brand voice, brand voice tooltip, Customize tooltip, and mobile-width dialog if available.

### Attached visual references

These screenshots are uploaded to Shortcut and should be embedded with image syntax so Shortcut renders previews:

![ai-customize-screenshot-01.png](https://media.app.shortcut.com/api/attachments/files/clubhouse-assets/5e0c5625-83f1-4c4f-b9a3-ac79e02e1f07/69f66fc0-19c8-43c8-857a-2908769e4c7c/ai-customize-screenshot-01.png)

![ai-customize-screenshot-02.png](https://media.app.shortcut.com/api/attachments/files/clubhouse-assets/5e0c5625-83f1-4c4f-b9a3-ac79e02e1f07/69f66fc1-67b4-4817-9ef3-fb724ab58840/ai-customize-screenshot-02.png)

![ai-customize-screenshot-03.png](https://media.app.shortcut.com/api/attachments/files/clubhouse-assets/5e0c5625-83f1-4c4f-b9a3-ac79e02e1f07/69f66fbf-8238-4e97-b51e-e877b66222c2/ai-customize-screenshot-03.png)

![ai-customize-screenshot-04.png](https://media.app.shortcut.com/api/attachments/files/clubhouse-assets/5e0c5625-83f1-4c4f-b9a3-ac79e02e1f07/69f66fc0-3abf-4b81-acb3-44d547e4bdcf/ai-customize-screenshot-04.png)

![ai-customize-screenshot-05.png](https://media.app.shortcut.com/api/attachments/files/clubhouse-assets/5e0c5625-83f1-4c4f-b9a3-ac79e02e1f07/69f66fc0-1bbf-44e3-83d6-8b204c41c8ec/ai-customize-screenshot-05.png)

![ai-customize-screenshot-06.png](https://media.app.shortcut.com/api/attachments/files/clubhouse-assets/5e0c5625-83f1-4c4f-b9a3-ac79e02e1f07/69f66fc0-021d-40b3-b86b-a82edba1153d/ai-customize-screenshot-06.png)

![ai-customize-screenshot-07.png](https://media.app.shortcut.com/api/attachments/files/clubhouse-assets/5e0c5625-83f1-4c4f-b9a3-ac79e02e1f07/69f66fc0-e70d-460a-ad26-6f5bac5d14c5/ai-customize-screenshot-07.png)

[ai-customize-prototype.jsx](https://media.app.shortcut.com/api/attachments/files/clubhouse-assets/5e0c5625-83f1-4c4f-b9a3-ac79e02e1f07/69f66fc0-3edd-40b8-a4a6-8458e0a58084/ai-customize-prototype.jsx)

---

### Impact on existing data

No persistent data changes. This story introduces UI state for the dropdown, dialog, selected brand voice usage, and pending generation intent only.

---

### Impact on other products

Web Composer only. AI functionality is web-only; native iOS and Android receive no AI Customize UI. White-label domains are impacted because the grouped control, dialog, buttons, and active states must use ContentStudio theme-aware styling. Chrome extension is not impacted.

---

### Dependencies

Depends on: **[BE] Add AI Customize caption generation for Composer** for final connected generation, but the control and dialog can be built with mocked response handling until backend is ready.

---

### Global quality & compliance

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Implementation references

*Pointers from research - not a contract. Engineering may choose a different approach.*

**Primary entry points:**
- `contentstudio-frontend/src/modules/composer_v2/components/EditorBox/EditorBox.vue` - current Customize switch and AI caption controls live here.
- `contentstudio-frontend/src/modules/composer_v2/views/SocialModal.vue` - owns `isSeparateBoxes(status)` and sharing details state.
- `contentstudio-frontend/src/modules/publisher/config/routes/publisher.js` - Brand Knowledge route is `ai-content-library-profile`.

**Existing behavior to preserve:**
- `EditorBox.vue` emits `isSeparateBoxes` through `handleCustomBoxToggle`.
- `SocialModal.vue` copies `common_sharing_details.message` into empty platform-specific message fields when Customize turns ON.

**Component gaps:**
- `docs/ui-components.md` lists no dedicated Tooltip or Pill/Chip component. Use the existing tooltip pattern / `CstPopup` for tooltips and `Badge` or simple theme-aware pills for platform reminders.

---

## [FE] Handle AI generation, regeneration, loading, errors, and edited states

### Description
As a social media manager, I want AI-generated captions to populate each selected platform tab with clear loading, retry, editing, and limit feedback so that I can confidently review and publish tailored content.

---

### Workflow

1. User writes a base caption and selects multiple platforms, or submits an empty-caption topic from the AI dialog.
2. User clicks "Generate per-platform captions".
3. If Customize is OFF, Composer turns Customize ON and shows platform tabs.
4. Each selected platform tab shows generation progress while AI writes for that platform.
5. Generated captions populate the relevant platform tabs as they complete.
6. User can switch tabs during generation to inspect platform progress.
7. User can edit any generated caption after generation completes.
8. User can regenerate all platforms or only the active platform from the AI dropdown.
9. User sees clear recovery options if all or some platforms fail.
10. User is warned before turning Customize OFF when AI-generated platform captions would be discarded.

---

### Acceptance criteria

- [ ] When the user generates from a non-empty base caption, no intent dialog is shown.
- [ ] When the user generates from an empty caption, generation uses the submitted topic from the dialog.
- [ ] If Customize is OFF when generation starts, Composer turns Customize ON before platform captions populate.
- [ ] Generation sends the selected platform list, base caption or intent, language, brand voice on/off state, and selected/default brand voice identifier when available.
- [ ] Each selected platform tab shows a loading label in the format "Writing for {Platform Name}..." while that platform is generating.
- [ ] Loading state includes shimmer skeleton lines matching the screenshot reference pattern.
- [ ] User can switch between platform tabs while generation is in progress.
- [ ] Menu actions that would conflict with an in-progress generation are disabled until the relevant generation finishes.
- [ ] Generated captions populate only their corresponding platform tabs.
- [ ] Generated captions are not copied into First Comment, carousel content, hashtags-as-separate-field, AI media, Bulk Schedule, Recycle Posts, or Campaigns.
- [ ] Generated captions respect platform limits: Facebook 63,206 characters, Instagram 2,200, Pinterest 500, X/Twitter 280, LinkedIn 3,000, Threads 500, TikTok 2,200, YouTube 5,000, and Bluesky 300.
- [ ] If the backend flags a caption as truncated, that platform tab shows a small notice: "Trimmed to fit {Platform Name}'s caption limit."
- [ ] After generation, every platform tab remains fully editable.
- [ ] The generated baseline is tracked per platform.
- [ ] If the user edits a generated tab so it differs from the generated baseline, that tab shows an "Edited" dot.
- [ ] The edited dot has accessible text/title "Edited".
- [ ] If the user reverts a tab to the generated baseline, the edited dot disappears.
- [ ] If the user regenerates a platform, that platform's generated baseline resets and its edited dot disappears.
- [ ] "Regenerate all platforms" regenerates all currently selected platforms and leaves removed platforms untouched.
- [ ] "Regenerate just [Platform]" regenerates only the active platform and leaves every other platform caption unchanged.
- [ ] If the user adds a new platform after generation, the new platform tab starts empty and can be filled by single-platform regeneration.
- [ ] If the user removes a platform after generation, that platform's generated caption state is discarded silently.
- [ ] If the base caption changes after generation, no automatic regeneration happens.
- [ ] If all platforms fail, tabs revert to their pre-generation captions and an inline error banner appears with a Retry action.
- [ ] Total failure error copy reads: "We couldn't generate captions. Try again."
- [ ] If some platforms fail, successful tabs remain populated and failed tabs show a retry action.
- [ ] Partial failure copy reads: "We couldn't generate captions for {Platform Names}. Try again."
- [ ] Blocked-content error copy reads: "We couldn't generate captions. Try rephrasing your input."
- [ ] Timeout error copy reads: "Caption generation is taking longer than expected. Try again."
- [ ] If AI text credits are exhausted, the existing credit-exhausted pattern is shown and no platform captions are changed.
- [ ] Toggling Customize OFF after AI generation opens a confirmation dialog before discarding platform-specific captions.
- [ ] Customize OFF confirmation title reads: "Discard platform-specific captions?"
- [ ] Customize OFF confirmation body reads: "Turning off Customize will discard your platform-specific captions. Continue?"
- [ ] Customize OFF confirmation buttons read "Keep Customize on" and "Discard captions".
- [ ] Confirming Customize OFF discards platform-specific captions and preserves the base/common caption.
- [ ] Cancelling Customize OFF keeps Customize ON and preserves all platform captions.
- [ ] Saving a draft or template stores the final caption text only and does not store generated baseline, edited-dot state, or other AI metadata.
- [ ] Existing Composer validation and post preview continue to use the final platform-specific captions.
- [ ] When the AI menu opens, a `customize_ai_menu_opened` Usermaven event fires with `{ selected_platforms, platform_count, customize_on }`.
- [ ] When the user starts generation, a `customize_ai_generate_clicked` Usermaven event fires with `{ selected_platforms, platform_count, source, brand_voice_on }`, where source is `base_caption` or `intent`.
- [ ] When generation succeeds, a `customize_ai_generate_success` Usermaven event fires with `{ selected_platforms, platform_count, generated_platforms, truncated_platforms, brand_voice_on }`.
- [ ] When generation fails, a `customize_ai_generate_failed` Usermaven event fires with `{ selected_platforms, platform_count, failed_platforms, error_type, brand_voice_on }`.
- [ ] When a user regenerates one platform, a `customize_ai_per_platform_regenerate` Usermaven event fires with `{ platform, brand_voice_on }`.
- [ ] When a user edits a generated caption for the first time per generation, a `customize_ai_caption_edited_after_generation` Usermaven event fires with `{ platform, character_count }`.
- [ ] When the empty-caption dialog opens, an `intent_dialog_shown` Usermaven event fires with `{ selected_platforms, platform_count, brand_voice_available }`.
- [ ] When the empty-caption dialog is submitted, an `intent_dialog_submitted` Usermaven event fires with `{ selected_platforms, platform_count, brand_voice_on }`.
- [ ] When the brand voice toggle changes, a `brand_voice_toggle_state` Usermaven event fires with `{ enabled }`.
- [ ] When the no-brand-voice setup CTA is clicked, a `brand_voice_setup_cta_clicked` Usermaven event fires with `{ source: 'composer_ai_customize' }`.
- [ ] Analytics payloads do not include caption content, topic text, user email, or other PII.

---

### Mock-ups

Attach the relevant generation-state screenshots to this story in Shortcut after story creation. Required visual references include loading state, post-generation edited dot, dropdown after generation, and mobile-width Composer behavior if available.

### Attached visual references

These screenshots are uploaded to Shortcut and should be embedded with image syntax so Shortcut renders previews:

![ai-customize-screenshot-01.png](https://media.app.shortcut.com/api/attachments/files/clubhouse-assets/5e0c5625-83f1-4c4f-b9a3-ac79e02e1f07/69f67045-ce10-4be2-80cf-34ee5ed887ff/ai-customize-screenshot-01.png)

![ai-customize-screenshot-02.png](https://media.app.shortcut.com/api/attachments/files/clubhouse-assets/5e0c5625-83f1-4c4f-b9a3-ac79e02e1f07/69f67046-0750-4a1e-879f-3893f12d7cf4/ai-customize-screenshot-02.png)

![ai-customize-screenshot-03.png](https://media.app.shortcut.com/api/attachments/files/clubhouse-assets/5e0c5625-83f1-4c4f-b9a3-ac79e02e1f07/69f67045-574c-480f-9308-08b4bccb17fc/ai-customize-screenshot-03.png)

![ai-customize-screenshot-04.png](https://media.app.shortcut.com/api/attachments/files/clubhouse-assets/5e0c5625-83f1-4c4f-b9a3-ac79e02e1f07/69f67046-195d-49db-811c-bb101543f5ec/ai-customize-screenshot-04.png)

![ai-customize-screenshot-05.png](https://media.app.shortcut.com/api/attachments/files/clubhouse-assets/5e0c5625-83f1-4c4f-b9a3-ac79e02e1f07/69f67046-6937-43fe-95b6-303f5fd5bf4c/ai-customize-screenshot-05.png)

![ai-customize-screenshot-06.png](https://media.app.shortcut.com/api/attachments/files/clubhouse-assets/5e0c5625-83f1-4c4f-b9a3-ac79e02e1f07/69f67046-f363-4304-8b8e-1f2772d5ddd2/ai-customize-screenshot-06.png)

![ai-customize-screenshot-07.png](https://media.app.shortcut.com/api/attachments/files/clubhouse-assets/5e0c5625-83f1-4c4f-b9a3-ac79e02e1f07/69f67045-a782-4fa9-b82b-628027bfb0e1/ai-customize-screenshot-07.png)

[ai-customize-prototype.jsx](https://media.app.shortcut.com/api/attachments/files/clubhouse-assets/5e0c5625-83f1-4c4f-b9a3-ac79e02e1f07/69f67045-fae7-45b6-831b-a8881c484dfb/ai-customize-prototype.jsx)

---

### Impact on existing data

Drafts, templates, and final Composer payloads should save only the final caption text. AI generation metadata, generated baselines, loading state, failed platform lists, and edited-dot state are transient Composer UI state only.

---

### Impact on other products

Web Composer only. Native mobile apps do not receive AI Customize functionality. Planner previews and saved drafts/templates are impacted only by receiving final platform-specific caption text. Chrome extension and AI Studio are not changed except for existing Brand Knowledge routing.

---

### Dependencies

Depends on: **[BE] Add AI Customize caption generation for Composer** and **[FE] Add AI Customize control and empty-caption dialog to Composer**.

---

### Global quality & compliance

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Implementation references

*Pointers from research - not a contract. Engineering may choose a different approach.*

**Primary entry points:**
- `contentstudio-frontend/src/modules/composer_v2/components/MainComposer.vue` - platform tabs and platform-specific editor rendering.
- `contentstudio-frontend/src/modules/composer_v2/views/SocialModal.vue` - sharing details state mutations and `setSharingMessage` behavior.
- `contentstudio-frontend/src/api/composer.ts` - existing Composer API helper pattern.

**Existing patterns:**
- `contentstudio-frontend/src/modules/composer_v2/components/AiCaptionModal.vue` updates `planStore.getPlan.used_limits.caption_generation_credit` after AI caption success.
- `contentstudio-frontend/src/modules/composer_v2/components/ActionsAside.vue` and AI Content Library composables show existing `userMaven.track(...)` usage.

**Suggested names:**
- API helper: `fetchAiCustomizeCaptionsApi`.
- UI state keys: `aiCustomizeLoadingPlatforms`, `aiCustomizeGeneratedBaseline`, `aiCustomizeFailedPlatforms`, `aiCustomizeTruncatedPlatforms`.

**Gotchas:**
- Existing per-platform sharing details use keys such as `facebook_sharing_details`, `instagram_sharing_details`, and `pinterest_sharing_details`.
- The current manual Customize copy behavior should remain the baseline behavior when users ignore AI.
