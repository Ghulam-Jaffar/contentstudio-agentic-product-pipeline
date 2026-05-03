# AI-Powered Composer Customization - Research

## Current State

- Epic `115059` exists in Shortcut as **AI-Powered Composer Customization** and currently has no stories. It is linked to objective `114402`, planned for May 3-10, 2026. Target iteration for the stories is `116060` (**04 May - 15 May - 2026**).
- The product handoff lives in `contentstudio-ai-customize-handoff/handoff/`. Source-of-truth files:
  - `PRD.md` - requirements, edge cases, platform limits, analytics events.
  - `prototype.jsx` - behavior/visual source of truth.
  - `SCREENSHOTS.md` and `screenshots/` - visual reference for control states, dropdowns, dialog, loading, edited dot, and mobile-width layout.
  - `VISUAL-SPEC.md` - tokens, spacing, animation, and icon references.
- Current web Composer Customize behavior:
  - The Customize switch is rendered in `contentstudio-frontend/src/modules/composer_v2/components/EditorBox/EditorBox.vue`.
  - The switch emits `isSeparateBoxes` through `handleCustomBoxToggle`.
  - `contentstudio-frontend/src/modules/composer_v2/views/SocialModal.vue` handles `isSeparateBoxes(status)` and `changeSharingBoxOption(...)`.
  - Current ON behavior copies `common_sharing_details.message` into empty platform-specific `*_sharing_details.message` fields, preserving manual customization.
  - Platform tabs are rendered in `contentstudio-frontend/src/modules/composer_v2/components/MainComposer.vue`.
- Current AI caption/credit behavior:
  - Existing AI caption frontend uses `fetchAiCaptionApi` in `contentstudio-frontend/src/api/composer.ts`, backed by `fetchAiCaption` from `contentstudio-frontend/src/modules/publish/config/api-utils.js`.
  - Backend route `POST /planner/fetchAiCaption` is defined in `contentstudio-backend/routes/web.php` and handled by `contentstudio-backend/app/Http/Controllers/Planner/HelperController.php::fetchAiCaption`.
  - Existing caption generation calls `env('LUMOTIVE_CAPTION_API') . 'caption_generation_gpt'`.
  - AI text credits use `caption_generation_credit`; backend checks/deducts via existing helper patterns and frontend updates `planStore.getPlan.used_limits.caption_generation_credit`.
- Brand voice / brand knowledge:
  - AI Content Library profile routes are under `aiContentLibrary/profile/get` in `contentstudio-backend/routes/web/ai.php`.
  - Frontend profile access uses `fetchAiProfileApi` and `useSetup()` in `contentstudio-frontend/src/modules/publisher/ai-content-library/`.
  - Existing Brand Knowledge UI route is `ai-content-library-profile` under Publisher / AI Studio.
- Analytics:
  - Usermaven is already used in Composer and AI Content Library.
  - Relevant existing event names include `ai_posts_generated`, `ai_post_regenerated`, and `brand_profile_created`.
  - The PRD specifies new Composer AI Customize events; they should be emitted from the FE generation story with small snake_case payloads.
- Native mobile:
  - iOS and Android Composer code exists, but story guidelines state ContentStudio mobile apps have no AI features. Native mobile AI stories should not be created.

## What Needs to Change

- Create 3 Shortcut stories under epic `115059`:
  - `[BE] Add AI Customize caption generation for Composer`
  - `[FE] Add AI Customize control and empty-caption dialog to Composer`
  - `[FE] Handle AI generation, regeneration, loading, errors, and edited states`
- Backend story must cover:
  - Batched generation from a base caption or empty-caption intent.
  - Single-platform regeneration.
  - Structured `{ platform: caption }` response.
  - Partial failure response shape.
  - 1 AI text credit per batched generation and 1 per single-platform regeneration.
  - Brand voice injection when `use_brand_voice` is true and a selected/default voice exists.
  - Full edge handling: insufficient credits, timeout, blocked content, malformed LLM output, over-limit output, full failure, partial platform failures.
  - Platform-specific caption limits from the PRD:
    - Facebook: 63,206, with practical prompt guidance of 500-1500 characters.
    - Instagram: 2,200.
    - Pinterest: 500.
    - X/Twitter: 280.
    - LinkedIn: 3,000.
    - Threads: 500.
    - TikTok: 2,200.
    - YouTube: 5,000.
    - Bluesky: 300.
- First FE story must cover:
  - Grouped Customize control with AI caret, divider, and existing toggle.
  - Dropdown menu states before/after generation.
  - Empty-caption modal with topic textarea, brand voice card, platform reminder strip, credit hint, and keyboard/dismissal behavior.
  - Screenshot/prototype-aligned UI copy and i18n readiness.
  - Brand voice setup CTA routing to the existing Brand Knowledge page.
  - Manual Customize behavior remains unchanged and does not auto-trigger AI.
- Second FE story must cover:
  - Base-caption generation flow.
  - Empty-caption generation flow after modal submit.
  - Full regeneration and active-platform-only regeneration.
  - Per-tab shimmer/loading state from screenshots.
  - Per-tab edited-dot baseline tracking.
  - Character-limit warnings after truncation.
  - Retry behavior for full and partial failures.
  - Customize OFF discard confirmation after AI-generated platform-specific captions exist.
  - Adding/removing platforms after generation.
  - Saving drafts/templates with final captions only, no AI metadata.
  - Usermaven events from the PRD.

## UX Reference

- Use `prototype.jsx` as the source of truth when PRD prose is ambiguous.
- Use `SCREENSHOTS.md` plus files in `contentstudio-ai-customize-handoff/handoff/screenshots/` as visual references for:
  - Customize OFF grouped control.
  - Customize ON tab layout.
  - AI dropdown before generation.
  - AI dropdown after generation.
  - Dialog with brand voice.
  - Dialog without brand voice.
  - Brand voice tooltip.
  - Generation loading state.
  - Post-generation edited dot.
  - Customize tooltip.
  - Mobile-width web layout.
- UI component catalog supports `Button`, `Dropdown`, `DropdownItem`, `Modal`/`Dialog`, `Textarea`, `Switch`, `Alert`, `ActionIcon`, `Icon`, and `Badge`. There is no dedicated Tooltip or Pill/Chip component; stories should flag use of existing tooltip pattern / `CstPopup` and `Badge` or theme-aware simple pills.

## Mobile Context

- Native iOS and Android AI Customize stories are out of scope because the guidelines state mobile apps have no AI features.
- The web FE stories should still include mobile-width responsive behavior for the Composer and modal, matching the optional mobile screenshot reference.
- Cross-product impact should explicitly state: web Composer only; no native mobile implementation; final caption data remains compatible with drafts/templates and planner preview.

## Files Involved

- `contentstudio-frontend/src/modules/composer_v2/components/EditorBox/EditorBox.vue` - current Customize switch and AI caption entry areas.
- `contentstudio-frontend/src/modules/composer_v2/components/MainComposer.vue` - platform tabs and platform-specific editor rendering.
- `contentstudio-frontend/src/modules/composer_v2/views/SocialModal.vue` - Composer state owner, Customize mode toggle/copy behavior, and sharing details mutations.
- `contentstudio-frontend/src/api/composer.ts` - existing Composer API helper pattern.
- `contentstudio-frontend/src/modules/publish/config/api-utils.js` - current `fetchAiCaption` URL constant location.
- `contentstudio-frontend/src/modules/publisher/ai-content-library/composables/useSetup.js` - current AI profile/brand voice frontend state pattern.
- `contentstudio-frontend/src/api/ai-content-library.ts` - current AI profile API helper.
- `contentstudio-backend/routes/web.php` - current planner/composer AI caption route area.
- `contentstudio-backend/app/Http/Controllers/Planner/HelperController.php` - existing AI caption generation and credit deduction behavior.
- `contentstudio-backend/routes/web/ai.php` - AI Content Library profile routes for brand voice.
- `contentstudio-backend/app/Repository/Ai/AiContentLibrary/AiContentLibraryProfileRepo.php` - brand voice/profile lookup.
- `contentstudio-backend/app/Helpers/Billing/PlanHelper.php` - AI text credit check/deduction pattern.

## Story Creation Notes

- Use `.claude/shortcut-config.json` as the active Shortcut config in this renamed repo.
- Use the New Feature Template `60cc481d-77f9-4f4b-92f0-f0fcc4eff65d`.
- Use project `2554` (Web App).
- Use iteration `116060` (04 May - 15 May - 2026).
- Use epic `115059`.
- Use Ready for Dev state `500000070`.
- No estimates and no labels.
- Backend story group: backend.
- FE story groups: frontend.
- Product area: composer.
- PRD wins over Shortcut epic description where they conflict; AI media suggestions are out of scope.
- Do not push local-only file paths into Shortcut story bodies. Shortcut readers cannot access this repository.
- Shortcut supports uploading files through `POST /api/v3/files` with `story_id`; after stories are created, upload relevant screenshots to the related FE stories' Files section.
- The JSX prototype should not be required to understand the story. Translate prototype behavior into acceptance criteria and implementation references. Optionally upload `prototype.jsx` to both FE stories as a supporting artifact, but treat it as secondary to the story body.
- In pushed stories, the Mock-ups section should say screenshots are attached to the story in Shortcut, not reference `contentstudio-ai-customize-handoff/...` local paths.
