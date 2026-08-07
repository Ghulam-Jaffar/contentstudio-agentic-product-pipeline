# Research: Q3 misc product fixes

A batch of independent small fixes across the planner status model, the content library AI creations section, billing, and AI chat. These are not one feature, so this stays a lightweight `/story` batch (like the earlier misc cleanup batches) rather than a `/feature` epic.

Suggested story split (7 stories): see the reasoning under each item.

---

## 1. "Partially failed" should be a real post status, not a side flag

**Current state**
- Backend keeps the post `status` as `published` and carries partial failure in a separate boolean field `partially_failed`.
  - `contentstudio-backend/app/Services/PlannerViewService.php` returns both `status` and `partially_failed` side by side (lines 24-25, 169-175).
  - `contentstudio-backend/app/Repository/Publish/Planner/PlansRepository.php` filters and counts on the `partially_failed` boolean throughout.
- Frontend shows the "Partially failed" badge by checking the boolean before it looks at `status`.
  - `contentstudio-frontend/src/modules/planner_v2/composables/usePlannerStatusBadge.ts` `getPostStatusText()` checks `postData.partially_failed` first, then falls through to the `status` switch.
  - The status helpers already have a `partially_failed` branch: `contentstudio-frontend/src/modules/planner_v2/utils/index.ts` (`getStatusBgClass`, `getStatusColorClass`, `getStatusPillClasses`, `getStatusColorCode`).
- So the value is representable as a status on the frontend, but the API never sends `status = 'partially_failed'`; it sends `published` plus the boolean.

**What needs to change**
- Backend: expose partial failure as a first-class status value in the API (`status = 'partially_failed'`) so partially failed posts read consistently everywhere status is used (list, filters, counts, single post).
- Frontend: consume the status value directly for the badge, filters, and status counts instead of special-casing the `partially_failed` boolean.

**Why two stories:** backend status derivation and frontend consumption are different skill sets. `[BE]` + `[FE]`.

---

## 2. Let users delete AI posts (single and bulk) from the AI Library and the content library

**Current state**
- The AI content library already has delete and bulk-action logic: `useAILibraryOperations.js` and `useAIPostActions.ts` (post CRUD incl. delete, selection, bulk actions), and `AIPostCard.vue` supports selection (`selectable`, `selectionMode`, `selectedPosts`, select/deselect emits).
- But delete is not surfaced in the UI: the card overflow menu (`AIPostCardMenu.vue`) only offers Schedule with AI, Add to composer, Save as draft. No delete option, and no visible bulk-delete toolbar.

**What needs to change**
- Surface a Delete action on each AI post (single) and a bulk Delete for a multi-selection, in the AI Library (`ai-content-library`) and in the content library AI creations "AI posts" section.
- Confirmation before delete, matching the media library delete pattern.

**Why one story:** delete backend/composables exist; this is FE wiring/exposure. `[FE]`. Related to item 4 (once the content library AI posts render as standard media, standard media delete applies there too).

---

## 3. Content library AI creations "AI posts" should render as media items, not the full publisher AI Posts UI

**Current state**
- `contentstudio-frontend/src/modules/publish/components/media-library/components/ContentLibraryAiPosts.vue` renders the whole publisher AI Posts experience: `<AIPosts :hide-search :hide-header />` plus `CustomGenerateModal` and `PostSettingsModal`, and a "setup required" gate that pushes users into brand knowledge setup.
- The AI creations section is reached via the media library sidebar (`SideBar.vue`, type `ai_creations`) with pills `ai_studio | clips | ai_posts` (`FiltersBar.vue`).

**What needs to change**
- In the content library AI creations "AI posts" pill, render just the generated items as a media grid like the rest of the content library (`ai_studio` / `clips` pills), removing the embedded publisher AI Posts UI: no "generate new", no brand knowledge setup gate, no credits, no generation/settings modals.

**Why one story:** FE only. `[FE]`.

---

## 4. Stop auto-creating the "My AI creations" and "AI Video Clips" folders

**Current state**
- `contentstudio-backend/app/Repository/Storage/MediaLibraryFoldersRepo.php`:
  - `getOrCreateAiFolder()` creates a root folder named "My AI creations" (`is_ai_folder`).
  - `getOrCreateAiVideoClipsFolder()` creates a root folder "AI Video Clips" (`is_ai_video_folder`), and migrates a legacy "Reels" folder.
  - `getOrCreateReelsFolder()` is a deprecated alias.
- These get called when AI content is generated/persisted:
  - `contentstudio-backend/app/Services/ChatMessageService.php` (lines 123, 237).
  - `contentstudio-backend/app/Services/AI/AiToolMediaPersistenceService.php` (line 34).
- The product now has a dedicated AI creations section in the content library (sidebar type `ai_creations` with `ai_studio` / `clips` / `ai_posts` pills), so these auto-created folders are redundant.

**What needs to change**
- Stop auto-creating the "My AI creations" and "AI Video Clips" (and legacy "Reels") folders when AI content is generated. AI content should still be persisted and still appear in the dedicated AI creations section.

**Why one story:** BE only. `[BE]`.

---

## 5. Billing: clear message when no card is attached (instead of "support has been notified")

**Current state**
- When a plan change or transaction fails, the billing dialogs fall back to a generic error: `UNKNOWN_ERROR()` = "Uh-oh! An unknown error occurred, support has been notified."
  - `contentstudio-frontend/src/modules/setting/components/billing/dialogs/UpgradePlanConfirmation.vue` catch blocks (lines 110, 166).
  - `contentstudio-frontend/src/modules/setting/components/billing/dialogs/UpgradePlanComponent.vue` (line 212).
- A specific "credit card issue" case already exists in the codebase: the string "Could not upgrade/downgrade due to credit card issue" is used in `CancelRecurringPlanDialog.vue` / `CancelPlanDialog.vue`.
- So when a user with no card on file tries to switch plans, they hit the generic "support has been notified" error rather than being told they need to add a card.

**What needs to change**
- Detect the "no payment method / card required" failure and show a clear, human message that explains what happened and what to do (add a card), instead of the generic unknown-error toast.

**Why one story:** FE handling of the error. `[FE]`.

---

## 6. AI chat: rename the bottom link "Explore AI Studio" to "Learn more"

**Current state**
- `contentstudio-frontend/src/modules/AI-tools/ChatBox.vue` (line 374) renders a hardcoded docs link with the text "Explore AI Studio" at the bottom of the normal chat view.

**What needs to change**
- Change the link text to "Learn more" (link target unchanged).

**Why one story:** trivial FE copy change. `[FE]`.

---

## 7. AI chat: show bulk-generated images in a carousel

**Current state**
- AI chat can generate multiple images at once. They render in a wrapping grid: `contentstudio-frontend/src/modules/AI-tools/components/BotMessageImages.vue` uses `flex flex-wrap gap-2` over `images`, each with a hover toolbar (preview, info, download, add to media library, add to composer).
- With several images the message gets tall and awkward.

**What needs to change**
- When a chat image response has more than one image, display them in a carousel inside the message, keeping each image's existing actions. Single-image responses stay as they are.

**Why one story:** FE only. `[FE]`. An existing `contentstudio-frontend/src/components/dashboard/HorizontalCarousel.vue` may be reusable; the UI catalog lists no carousel, so the component choice should be flagged.

---

## Files involved

Backend
- `contentstudio-backend/app/Services/PlannerViewService.php`
- `contentstudio-backend/app/Repository/Publish/Planner/PlansRepository.php`
- `contentstudio-backend/app/Repository/Storage/MediaLibraryFoldersRepo.php`
- `contentstudio-backend/app/Services/ChatMessageService.php`
- `contentstudio-backend/app/Services/AI/AiToolMediaPersistenceService.php`

Frontend
- `contentstudio-frontend/src/modules/planner_v2/composables/usePlannerStatusBadge.ts`
- `contentstudio-frontend/src/modules/planner_v2/utils/index.ts`
- `contentstudio-frontend/src/modules/publisher/ai-content-library/` (AIPostCardMenu.vue, AIPostCard.vue, AIPostSection.vue, useAILibraryOperations.js, useAIPostActions.ts)
- `contentstudio-frontend/src/modules/publish/components/media-library/components/ContentLibraryAiPosts.vue`
- `contentstudio-frontend/src/modules/publish/components/media-library/components/FiltersBar.vue`, `SideBar.vue`
- `contentstudio-frontend/src/modules/setting/components/billing/dialogs/UpgradePlanConfirmation.vue`, `UpgradePlanComponent.vue`
- `contentstudio-frontend/src/modules/AI-tools/ChatBox.vue`
- `contentstudio-frontend/src/modules/AI-tools/components/BotMessageImages.vue`
- `contentstudio-frontend/src/components/dashboard/HorizontalCarousel.vue`
