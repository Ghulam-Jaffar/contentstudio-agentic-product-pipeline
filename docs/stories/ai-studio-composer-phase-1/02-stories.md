# 02 — Stories: AI Studio Composer Phase 1

**Epic:** [AI Studio Composer: Presets & UI/UX Update](https://app.shortcut.com/contentstudio-team/epic/117353) (id `117353`)
**Project:** Web App (id `2554`)
**Group:** Frontend (`5fec5c8b-ca96-4126-a260-acb3f38fbcd7`)
**Skill set:** Frontend
**Product area:** Composer
**Workflow state:** Ready for Dev (`500000070`)
**Story template:** New Feature Template (`60cc481d-77f9-4f4b-92f0-f0fcc4eff65d`)
**Mockup:** https://ai-studio-cs.lovable.app/

All 10 stories are `[FE]` — AI features are web-only per `docs/story-guidelines.md` section 3.

---

## Story 1 — `[FE] Build collapsible AI Studio composer with Image/Video mode and redesigned model picker`

### Description

As a ContentStudio user generating AI content, I want a calm single-line composer that expands into a full creation tool when I focus it or pick a generation mode, so that low-effort chat stays simple while structured workflows are one tap away.

This story replaces today's always-expanded composer with a three-state composer (resting, mode-active, preset-active), introduces an explicit Image/Video Mode chip with an entry popover, and rebuilds the model picker as a search-first dropdown with a Featured section, capability badges, and a locked state. It also updates the existing empty-state quick-start buttons to use the new entry pattern and adds a "Browse presets" button.

### Workflow

1. User opens AI Studio. The chat area shows the empty state with a centered headline "What are we creating today?", subline "Type a message to start, or pick an intent below.", and three buttons: "🖼 Generate an image", "🎥 Generate a video", "✨ Browse presets".
2. The composer at the bottom is in its **resting** state — a single-line input with only a textarea, mic icon, and send icon visible. No top control row, no bottom row.
3. User clicks "🖼 Generate an image". The composer enters **mode-active** state. A top control row appears with the Mode chip ("Image"), Model pill (showing the default image model), Ratio pill, Style pill, and Quality pill. A bottom row appears with the standard utility icons, Brand chip, Presets button, and Enhance toggle.
4. User clicks the Model pill. A 380-pixel-wide dropdown opens beneath it. The dropdown header is a search input with the placeholder "Search models". Below the search, a "Featured" section shows three image-mode model tiles. Below that, an "All models" list shows every image model with provider abbreviation, model name, NEW or PRO badges where applicable, a one-line description, capability badges (Reference, Start/End, Multi-shot, Audio, 4K, Fast), and the credit cost on the right. The currently selected model has a blue check icon and a light-blue tinted background.
5. User types "fast" in the search. The list filters in place. User clicks a model. The dropdown closes and the Model pill updates to show the new model name.
6. User clicks the "×" on the Mode chip. The composer collapses back to the resting state. Prompt, ratio, style, and quality clear. Attachments persist.
7. User clicks the textarea in the resting state. The composer expands to mode-active immediately, defaulting to Image mode. User starts typing.
8. User clicks elsewhere on the page (textarea loses focus). After a 150-millisecond delay, if the prompt is empty, no mode is set, no slots filled, no attachments, no brand selected, and no popover is open, the composer collapses back to resting. Otherwise it stays expanded.
9. User clicks the Brand chip popover (a popover, not the textarea). Focus leaves the textarea but the composer stays expanded because the popover-open state keeps it expanded.

### Acceptance criteria

- [ ] On the empty chat state, three buttons render in a row centered above the composer: "🖼 Generate an image", "🎥 Generate a video", "✨ Browse presets".
- [ ] Clicking "🖼 Generate an image" enters Image mode and expands the composer.
- [ ] Clicking "🎥 Generate a video" enters Video mode and expands the composer.
- [ ] Clicking "✨ Browse presets" opens the Presets tray (built in `[FE] Add 16 preset workflows to AI Studio composer`).
- [ ] On first load, when there are no messages, the centered headline reads "What are we creating today?" and the subline reads "Type a message to start, or pick an intent below."
- [ ] When mode is null, prompt is empty, no slots are filled, no attachments exist, no brand is set, and no popover is open, the composer renders as a single-line input with a textarea, mic icon, and send icon (resting state).
- [ ] Resting state has no top control row visible and no bottom row visible.
- [ ] Focusing the textarea expands the composer to mode-active state with the default mode set to Image.
- [ ] Setting any of: prompt non-empty, mode non-null, preset non-null, slots filled, attachments present, brand set, or any popover open keeps the composer expanded regardless of focus.
- [ ] When all of the above are false and focus leaves the textarea, the composer collapses back to resting after a 150-millisecond delay.
- [ ] Collapse uses an ease-out animation on max-height and opacity of the top and bottom rows.
- [ ] Mode-active state shows a top control row with the Mode chip on the left, followed by Model, Ratio, Style, and Quality pills. Duration pill appears only in Video mode (built in `[FE] Add Ratio, Duration, and Quality pills to AI Studio composer top control row`).
- [ ] The Mode chip shows "Image" with an image icon for Image mode, and "Video" with a video icon for Video mode.
- [ ] The Mode chip has an "×" on the right that, when clicked, clears mode and returns the composer to resting state. Mode, model, preset, slots, ratio, duration, style, quality, and multi-shot all clear. Attachments persist.
- [ ] Clicking the Model pill opens a 380-pixel-wide dropdown, max-height 480 pixels, scrollable.
- [ ] The dropdown header is a single search input with the placeholder "Search models". No filter chips.
- [ ] The dropdown lists only models matching the current mode. In Image mode, only image models. In Video mode, only video models.
- [ ] When the search input is empty, a "Featured" section renders above the "All models" list, showing three model tiles for the current mode.
- [ ] Each Featured tile shows model name, NEW badge (if applicable), one-line description, and credit cost.
- [ ] Each "All models" row shows: provider logo (2-letter abbreviation in a zinc-100 box on the left), model name, NEW or PRO badges next to the name, a one-line description, capability badges (Reference, Start/End, Multi-shot, Audio, 4K, Fast — only those that apply), and the credit cost on the right.
- [ ] The currently selected model's row has a blue check icon and a `bg-primary-cs-50` (or equivalent theme-aware tinted) background.
- [ ] Typing in the search filters across model name and provider name, in place.
- [ ] Selecting a model in the dropdown closes the dropdown and updates the Model pill text.
- [ ] Switching to a model of a different mode (e.g., picking a video model while in Image mode) opens a confirm dialog (rendered by `[FE] Add validation banners, toasts, and confirm dialogs to AI Studio composer`). On confirm, mode switches and ratio/duration/style/quality reset to the new model's first values; slots not supported by the new state are dropped. On cancel, nothing changes. Prompt is preserved either way.
- [ ] Switching within the same mode applies silently. If filled slots aren't supported by the new model, they are dropped and a toast renders (built in the validation story). Ratio/duration/style/quality reset only when the current value isn't supported by the new model.
- [ ] In a State C (preset-active) context, the Model pill renders with a Lock icon prefix in amber styling and is non-clickable; the dropdown shows an amber banner at the top reading "🔒 Model is locked by the active preset.", with all rows at 50% opacity and `cursor-not-allowed` (rendered alongside `[FE] Add 16 preset workflows to AI Studio composer`).
- [ ] All popovers close on Escape key.
- [ ] Send button is disabled when mode is null and prompt is empty and no slots are filled and no attachments exist.
- [ ] Send button is disabled when mode is set and prompt is empty and no slots are filled and no attachments exist.
- [ ] When the user applies a preset, the metadata payload sent to the existing `sendChatMessage` flow includes a Usermaven event `ai_studio_preset_applied` with `{ preset_key, mode }` — *(this AC actually belongs to the preset story; flagged here because Story 1 owns the send-payload extension. If the preset story owns it instead, remove from here.)*

### Mock-ups

- Lovable prototype: https://ai-studio-cs.lovable.app/
- Spec doc (Shortcut): _will be linked from the epic; URL TBD when uploaded_

### Impact on existing data

- The state shape inside the AI Chat client store extends with new keys: `composerExpanded` (boolean derived from inputs), `mode` (existing), `model` (existing), `preset` (new — null in this story; populated by the presets story), `quality` (new — added here in case Quality pill ships before the top-row story), `popoverOpen` (registry of which popover is open, for the expansion-keeping rule).
- The existing `mediaDetails` payload sent to `sendChatMessage` is preserved; this story does not change the server contract for image/video generation.
- The existing empty-state buttons are replaced — "Write with AI", "Generate Hashtags", "Improve with AI", and "Caption Images" buttons are removed from the empty state; the three new entry buttons replace them. The corresponding handlers are no longer reachable from the empty state, but they remain on text-action paths (the `EventBus` listeners stay, since they're triggered from elsewhere in the product).

### Impact on other products

- AI features are web-only per `docs/story-guidelines.md` section 3 — no iOS/Android/Chrome extension work.
- White-label theming: all new color usage must use `text-primary-cs-*` / `bg-primary-cs-*` and equivalent theme-aware tokens. Amber (locked) and blue (selected) accents come from `@contentstudio/ui` tokens or theme-aware Tailwind classes — no hardcoded hex.
- Existing AI Studio entry points from the dashboard, header, sidebar, and "Smart Scheduling" tool are unchanged.

### Dependencies

- This story is the foundation for the rest of the epic. The following stories build into the shells this story creates:
  - `[FE] Add reference slot system to AI Studio composer with contextualized Add Media modal`
  - `[FE] Add @mention attachment autocomplete to AI Studio composer`
  - `[FE] Add 16 preset workflows to AI Studio composer`
  - `[FE] Add multi-shot video generation to AI Studio composer`
  - `[FE] Restyle Brand selector as prominent chip in AI Studio composer`
  - `[FE] Apply three-tier chip styling to AI Studio composer bottom row`
  - `[FE] Add Ratio, Duration, and Quality pills to AI Studio composer top control row`
  - `[FE] Add mobile-web responsive bottom sheets and compressed layout to AI Studio composer`
  - `[FE] Add validation banners, toasts, and confirm dialogs to AI Studio composer`

### Global quality & compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories)
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled)
- [ ] UI theming supported (default + white-label, design library components are being used)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension) — web-only feature, mobile/Chrome ext N/A

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- **Primary entry points:** `contentstudio-frontend/src/modules/AI-tools/ChatBox.vue` (empty-state buttons live here today), `contentstudio-frontend/src/components/dashboard/ChatInput.vue` (1483 lines — already over the 500-line cap; plan to split as part of this refactor), `contentstudio-frontend/src/composables/useAIChatMessage.ts` (state shape lives here).
- **Existing pattern to repurpose:** `ChatInput.vue` already implements an Image/Video generation toggle via `isImageGenerationActive` / `isVideoGenerationActive` and `toggleImageGeneration` / `toggleVideoGeneration`. The new Mode chip is a re-skin and re-anchor of this toggle.
- **Existing model picker to redesign:** `contentstudio-frontend/src/modules/AI-tools/components/MediaGenerationOptions.vue` has the provider-grouped dropdown with NEW badges, generation-time badges, image-to-image / image-to-video support badges, and selection check. Extract the data-fetching parts into the new picker; replace the UI with the search-first dropdown described in the spec.
- **Suggested new components:** `src/modules/AI-tools/components/composer/ComposerTopRow.vue`, `ComposerBottomRow.vue` (shells), `ModelPicker.vue`, `ModeChip.vue`. Plus a `useComposerState.ts` composable to centralize the `shouldExpand` logic + popover registry.
- **Existing send flow (preserve):** `props.sendChatMessage(content, images, mediaDetails, aiLibraryPosts, videos)` in `ChatInput.vue` and `ChatBox.vue` is unchanged. The `mediaDetails` payload extends transparently to carry future fields (preset, slots, multi-shot, quality).
- **i18n:** New keys go under `ai_tools.composer.*` — add to every locale directory under `src/locales/`. Empty-state copy goes under `ai_tools.composer.empty_state.*`.
- **CLAUDE.md compliance:** all new components in `<script setup lang="ts">`; use `@contentstudio/ui` `Dropdown` / `DropdownItem` / `Button` / `Icon` / `TextInput` / `SearchInput` / `Badge` / `ListItem`; never override their styles with Tailwind colors; chip styling is built inline with theme-aware Tailwind (`text-primary-cs-*`, `bg-primary-cs-50`).
- **Designer touchpoint:** chip styling for Mode chip and Model pill — pair with the Product designer on padding, border treatment, and amber/lock styling.

---

## Story 2 — `[FE] Add reference slot system to AI Studio composer with contextualized Add Media modal`

### Description

As a ContentStudio user generating AI content, I want clearly-labeled reference slots that show only when they have a structured purpose, so that I know exactly which file goes where (e.g., "First frame", "Source face") and I'm not confused by always-empty optional reference boxes.

This story adds the slot row to the composer, the 12 slot types defined in the spec, slot UI states (empty / required / filled / coupled first-last / multi-slot wrap), and extends the existing Add Media modal with an optional slot-context banner and per-slot type filtering when the modal is invoked from a slot.

### Workflow

1. User has the composer in Image mode with no preset active. The slot row does not render — there are no slots in plain Image mode.
2. User switches to Video mode with a frame-capable model. A slot row appears below the top control row. It shows two coupled slots side by side: "First frame" and "Last frame", with a small "→" arrow between them. Each slot is a 160-pixel-wide tile with a 4:3 aspect ratio, a dashed border, a `+` icon, the slot label, and a hint subtitle.
3. User clicks the "First frame" slot. The Add Media modal opens. At the top of the modal, a blue context banner reads: "Adding to First frame · Accepts: Image · Max 20MB".
4. The media grid in the modal shows all available files. Files of incompatible types (e.g., videos) appear at 40% opacity, in grayscale, with `cursor-not-allowed`. Hovering one shows a tooltip: "Videos aren't supported in First frame".
5. User picks a JPEG. The "Add Media" button at the footer becomes enabled. User clicks it. The modal closes. The "First frame" slot now shows a solid border, a thumbnail preview of the picked image, and a filename overlay at the bottom with an "×" to remove.
6. User clicks the "×" on the filename overlay. The slot returns to its empty state.
7. User selects a preset like "Face swap". The slot row replaces with the preset's slots: "Face to use" (source_face) and "Image to edit" (target_image). Both have a blue dashed border, a blue dot in the top-right corner, and a "Required" subtitle in blue text.
8. User picks the source video model (no preset). They click "Last frame" and somehow a video file is selected (debug only — UI prevents this). A red banner renders above the slot row with the copy: "Videos aren't supported in Last frame." On the right, an "×" closes the banner. Send is disabled while the banner is shown.
9. User is in Image mode and tries to fill a slot with a video file. The same banner renders, but with an additional link on the right: "Switch to Video". Clicking it switches the composer to Video mode in one tap.

### Acceptance criteria

- [ ] In Image mode without a preset, no slot row renders.
- [ ] In Video mode with a frame-capable model, the slot row renders First frame + Last frame as a coupled pair with a "→" arrow between them, both 160 pixels wide, side by side.
- [ ] In Video mode with a non-frame model, no slot row renders.
- [ ] In State C (preset active), the slot row renders the preset's defined slots in the order specified.
- [ ] Empty slot UI: dashed border, plus icon centered, label below the icon, hint subtitle below the label, fixed width 160 pixels, aspect ratio 4:3.
- [ ] Required slot UI (preset only): blue dashed border, blue dot in top-right corner, "Required" subtitle in blue text.
- [ ] Filled slot UI: solid border, gradient thumbnail preview filling the tile, filename overlay at bottom with an "×" on the right to remove.
- [ ] Coupled first/last frame slots render side by side with a small "→" arrow between them, both at 160 pixels.
- [ ] Multi-slot presets (2+ slots) render as a `flex flex-wrap` row with each slot at 160 pixels, fixed-shrink. Wraps to next row when container narrows.
- [ ] All 12 slot keys are supported with their labels and hints exactly as specified:
  - `subject` — "Image to edit" / "The image you want to change" / image / 20MB
  - `style_ref` — "Style reference" / "Match this look or aesthetic" / image / 20MB
  - `mask` — "Area to edit" / "Mark the region to change" / image / 20MB
  - `character_ref` — "Character" / "Keep this person consistent" / image / 20MB
  - `first_frame` — "First frame" / "How the clip starts" / image / 20MB
  - `last_frame` — "Last frame" / "How the clip ends" / image / 20MB
  - `source_face` — "Face to use" / "The face you want in the result" / image / 20MB
  - `target_image` — "Image to edit" / "Where to put the new face" / image / 20MB
  - `base_video` — "Video to edit" / "The clip you want to change" / video / 200MB / ≤30s
  - `motion_video` — "Motion to copy" / "Video showing the movement" / video / 100MB / ≤30s
  - `motion_target` — "Subject" / "Apply the motion to this person" / image / 20MB
  - `audio` — "Audio clip" / "Voice or sound to sync to" / audio / 20MB
- [ ] Clicking any empty slot opens the Add Media modal with the slot context.
- [ ] When opened from a slot, the Add Media modal shows a blue context banner at the top with the copy: "Adding to {Slot label} · Accepts: {types} · Max {size}MB" and adds " · ≤{duration}s" only when the slot has a duration constraint.
- [ ] When opened from a slot, files of incompatible types in the media grid render at 40% opacity, grayscale, with `cursor-not-allowed`, and are non-clickable.
- [ ] When the slot has a max-size constraint, files exceeding the size also render at 40% opacity, grayscale, non-clickable, and a tooltip on hover reads "{Filename} exceeds {size}MB limit".
- [ ] When the slot has a duration constraint, videos exceeding the duration also render at 40% opacity, grayscale, non-clickable, with a tooltip "{Filename} exceeds {duration}s limit".
- [ ] Picking a compatible file enables the "Add Media" footer button.
- [ ] Clicking the "Add Media" button closes the modal and fills the slot.
- [ ] When opened from the paperclip (attachment context), the modal does NOT show the slot context banner; the header reads "Attach media" (no banner).
- [ ] When opened from the paperclip, no type filtering applies — all files are clickable.
- [ ] Clicking the "×" on a filled slot's filename overlay clears the slot.
- [ ] If an incompatible file ends up filling a slot (debug or programmatic), a red banner renders directly above the slot row with the copy: "{Videos|Images|Audio} aren't supported in {Slot label}." (rendering of this banner can be coordinated with `[FE] Add validation banners, toasts, and confirm dialogs to AI Studio composer`).
- [ ] When the user is in Image mode and the slot expects video, the red banner has an additional link aligned to the right: "Switch to Video". Clicking it switches to Video mode in one tap and clears the validation banner.
- [ ] Send button is disabled while a slot validation banner is shown.

### Mock-ups

- Lovable prototype: https://ai-studio-cs.lovable.app/

### Impact on existing data

- AI Chat client state extends with `slots: { [slotKey]: { fileId, url, filename, type, size, duration } | null }` and `slotValidationError: { slotKey, message } | null`.
- The existing Add Media modal gains optional props: `kind` (`"slot" | "attachment"`, default `"attachment"`), `slot` (the slot definition object: key, label, hint, accepts, maxSize, maxDuration), and `onSelect` (a callback that receives the picked file object). When `kind` is undefined or `"attachment"`, the modal behaves exactly as today.
- When the user picks a different model within the same mode, slots that aren't supported by the new model are dropped from `slots` and a toast renders ("Dropped {N} reference{s} — not supported by {model name}", built in the validation story).

### Impact on other products

- AI features are web-only — no iOS/Android/Chrome extension work.
- The Add Media modal is also consumed by `composer_v2` (the publish composer) and the media library views. Existing consumers must continue working unchanged when the new optional props are absent.
- All copy goes through `$t()` and into every locale directory.

### Dependencies

- Depends on: `[FE] Build collapsible AI Studio composer with Image/Video mode and redesigned model picker` (slot row anchors into the top-row shell).
- Coordinates with: `[FE] Add validation banners, toasts, and confirm dialogs to AI Studio composer` for the file-mismatch banner rendering and the model-switch toast (this story emits the events; the validation story renders the UI).
- Coordinates with: `[FE] Add 16 preset workflows to AI Studio composer` (presets define the slot lists).

### Global quality & compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories)
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled)
- [ ] UI theming supported (default + white-label, design library components are being used)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension) — web-only

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- **Primary entry points:** `contentstudio-frontend/src/components/dashboard/ChatInput.vue` (slot row anchors above the textarea); the existing Add Media modal — search the codebase for usages, likely lives in `src/modules/publish/components/media-library/` based on the `Asset.vue` and `MediaLibraryTab.vue` patterns.
- **Suggested new components:** `src/modules/AI-tools/components/composer/ReferenceSlotRow.vue`, `ReferenceSlot.vue`. Optionally a `useReferenceSlots.ts` composable for the slot-state logic.
- **Existing behavior to preserve:** the existing inline paperclip attachment flow (drag-drop, image-required-but-missing amber banner for image-to-video models, image-to-image support badges, etc.) is unchanged. Slots are a *separate* concept that lives above the textarea, not a replacement for attachments.
- **Validation eventing:** consider exposing slot-validation state via the message store so the validation story can subscribe — e.g., `slotValidationError` ref, with `clearSlotValidation()` and `triggerSlotValidation(slotKey, message)` actions.
- **i18n:** new keys under `ai_tools.composer.slots.*` (slot labels, hints) and `ai_tools.composer.add_media.*` (banner copy: "Adding to {label} · Accepts: {types} · Max {size}MB · ≤{duration}s"). Add to every locale directory.
- **Care:** Add Media modal extension must keep existing `composer_v2` consumers working; introduce the new props as optional with sane defaults.
- **Designer touchpoint:** dashed-border treatment for required slots (blue dashed + dot), filled slot's filename overlay gradient, type-mismatch tooltip styling.

---

## Story 3 — `[FE] Add @mention attachment autocomplete to AI Studio composer`

### Description

As a ContentStudio user generating AI content with multiple attachments, I want to mention attachments inline in my prompt with `@filename`, so that I can describe how each attachment is used without losing flow ("animate @product_shot_1 with a slow zoom into @logo_overlay").

This story adds the `@`-trigger autocomplete popover to the composer's textarea. The popover lists the user's current attachments only. It supports the empty-state case (no attachments yet) with a helpful CTA.

### Workflow

1. User has the composer expanded in Image or Video mode with two attached images: `product_shot.jpg` and `logo_overlay.png`. They've added these via the paperclip.
2. User types a prompt: "Place the". They press space, then type "@". The autocomplete popover opens, anchored above and slightly to the left of the cursor. The popover lists the two attachments, each with a 24×24 thumbnail and the filename.
3. User types "log". The list filters to `logo_overlay.png`. User clicks the row. The popover closes. The textarea now reads: "Place the @logo_overlay" with the cursor right after the inserted text. Focus returns to the textarea.
4. User keeps typing: " on the @". The popover opens again, this time listing both attachments (since "@" was just typed without filter text). User picks `product_shot.jpg`. The textarea reads: "Place the @logo_overlay on the @product_shot".
5. User has no attachments and types "@". The popover opens with an empty state: a centered paperclip icon, the title "No attachments yet", a sub-text "Click 📎 below to attach files, then mention them with @", and a full-width black button labeled "Attach a file". Clicking the button opens the Add Media modal in attachment mode. After the user adds a file, focus returns to the textarea.
6. User types "@" mid-word (e.g., "user@email"). Because the "@" is not at the start and not preceded by whitespace, the popover does not open.
7. User has the popover open and presses Escape. The popover closes; focus stays in the textarea; the typed "@" remains in the textarea.

### Acceptance criteria

- [ ] Typing "@" at the start of the textarea opens the autocomplete popover.
- [ ] Typing "@" preceded by whitespace opens the autocomplete popover.
- [ ] Typing "@" mid-word (preceded by a non-whitespace character) does NOT open the popover.
- [ ] The popover is anchored above and slightly to the left of the cursor's screen position.
- [ ] The popover lists ONLY attachments — not slots, not modifiers, not anything else.
- [ ] When attachments exist, each item renders as a row with a 24×24 thumbnail on the left and the filename on the right.
- [ ] Typing characters after `@` filters the list in place by filename substring match (case-insensitive).
- [ ] Clicking a filtered item inserts `@{filename_without_extension}` at the cursor position, with spaces in the filename converted to underscores. Example: file `Product Shot.jpg` → inserts `@Product_Shot`.
- [ ] After insertion, the popover closes and focus returns to the textarea with the cursor positioned immediately after the inserted text.
- [ ] Pressing Up/Down arrow keys while the popover is open navigates the list. The highlighted item has a `bg-primary-cs-50` (or theme-equivalent) background.
- [ ] Pressing Enter while the popover is open selects the highlighted item (same behavior as click).
- [ ] Pressing Escape while the popover is open closes the popover and leaves the typed "@" in the textarea.
- [ ] Clicking outside the popover closes it.
- [ ] Blurring the textarea closes the popover.
- [ ] When no attachments exist, the popover opens to an empty state with:
  - A centered paperclip icon, 24×24, in zinc-300.
  - A title at 13 pixels, font-medium, zinc-700: "No attachments yet".
  - A sub-text at 11.5 pixels, zinc-500: "Click 📎 below to attach files, then mention them with @".
  - A full-width black button (zinc-900 background, 12 pixels white text) labeled "Attach a file".
- [ ] Clicking the "Attach a file" button in the empty state opens the Add Media modal in attachment context (no slot context banner).
- [ ] After the user adds a file via the modal, the modal closes and focus returns to the textarea (the popover does not re-open automatically).

### Mock-ups

- Lovable prototype: https://ai-studio-cs.lovable.app/

### Impact on existing data

- No changes to message state shape. Mention text in the prompt (e.g., `@product_shot_1`) is sent as-is in the `content` field of the existing send payload.
- The mapping between the mention token and the corresponding attachment is engineering-owned (engineering picks: pre-process tokens to structured references, send prompt + attachment list separately, or a structured prompt format). The user-facing behavior is unambiguous: if a user mentions an attachment, the AI treats that attachment as a reference for the corresponding part of the prompt.

### Impact on other products

- Web-only feature (AI features are web-only per `docs/story-guidelines.md` section 3).
- All copy goes through `$t()` and into every locale directory.

### Dependencies

- Depends on: `[FE] Build collapsible AI Studio composer with Image/Video mode and redesigned model picker` (popover anchors to the textarea built/refactored in that story).
- Soft dependency: the empty-state CTA opens the Add Media modal — works with whatever attachment-context behavior already exists today.

### Global quality & compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories)
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled)
- [ ] UI theming supported (default + white-label, design library components are being used)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension) — web-only

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- **Primary entry points:** `contentstudio-frontend/src/components/dashboard/ChatInput.vue` (textarea + cursor logic). The existing `currentMessage.images.urls` / `currentMessage.videos.items` arrays in `useAIChatMessage` are the source of attachments to list.
- **Suggested new component:** `src/modules/AI-tools/components/composer/MentionAutocomplete.vue`. A lightweight composable like `useMentionAutocomplete.ts` may help if cursor + popover coordination grows.
- **Cursor positioning:** the existing `ChatInput.vue` already exposes `setContent`, `focus`, and `positionCursorAtEnd` methods (called from `ChatBox.vue`). Extend with a `getCursorRect()` helper for popover anchoring.
- **i18n keys:** `ai_tools.composer.mentions.empty_title` ("No attachments yet"), `ai_tools.composer.mentions.empty_subtext`, `ai_tools.composer.mentions.attach_button` ("Attach a file"). Add to every locale directory.
- **Engineering note:** how mentions resolve to attachment references on send is engineering's call (pre-process tokens before sending, send prompt + attachment list separately, structured prompt format). The PRD spec calls this out explicitly in section 8.

---

## Story 4 — `[FE] Add 16 preset workflows to AI Studio composer`

### Description

As a ContentStudio user, I want pre-configured generation workflows for common tasks (face swap, style transfer, headshot, etc.) so that I don't have to know which model and which slot configuration are right for what I'm trying to do.

This story adds the Presets tray, 16 preset workflows across four categories, the preset chip in the top control row (replacing the Mode chip in State C), the preset application flow (locks model, reshapes slots, marks required slots), and the URL-input variant for the `click_to_ad` preset.

### Workflow

1. User clicks the "Presets" pill in the bottom row, or clicks "✨ Browse presets" on the empty state. The Presets tray slides in above the composer (desktop) or opens as a bottom sheet (mobile — built in `[FE] Add mobile-web responsive bottom sheets and compressed layout to AI Studio composer`).
2. The tray header reads "Presets" on the left, with a search input and a close "×" button on the right.
3. Below the header, category tabs render: "All", "Image edit", "Video tools", "Ads & products", "Audio". The active tab has a `bg-primary-cs-50` background and `text-primary-cs-700` text.
4. Below the tabs, a grid of preset cards renders. Each card shows: a 1.4:1 aspect-ratio thumbnail with the preset's emoji centered (large), the preset name (font-medium, 12.5 pixels), a mode badge ("IMAGE" or "VIDEO" — uppercase 9 pixels, zinc-200 background, zinc-600 text), and a truncated description.
5. Cards from a different mode than the current one render at 50% opacity (still clickable).
6. User clicks "Face swap". A confirm dialog renders if the preset's mode differs from the current mode (built in the validation story): "Switch mode? Face swap is an image preset. Switch from video to image?". On confirm, the composer applies the preset.
7. Applying the preset sets mode, sets and locks the model (engineering picks a model satisfying the preset's capability requirements), reshapes the slot row to the preset's slots, marks required slots, and shows a toast: "Applied preset: Face swap".
8. The Mode chip in the top control row is replaced by a Preset chip showing the preset's emoji and name. The Model pill renders with a Lock icon prefix and amber styling. The slot row shows "Face to use" (source_face) and "Image to edit" (target_image), both with the required treatment.
9. Send is disabled until both required slots are filled.
10. User clicks the "×" on the Preset chip. The composer drops back to State B (plain mode). The model becomes user-modifiable again. The slot row updates to plain-mode slots (none in plain Image mode; first/last frame in plain Video mode with frame-capable model). Filled slots not in the new state are dropped.
11. User clicks the "URL → Ad" preset. Instead of a slot row, a single labeled URL input renders with the placeholder "Paste a product URL (e.g., https://your-store.com/product/...)". The user pastes a URL and clicks send. The chat shows a "Generating ad…" loading state while the backend handles URL scraping and ad video generation.

### Acceptance criteria

- [ ] Clicking the "Presets" pill in the bottom row opens the Presets tray.
- [ ] Clicking "✨ Browse presets" on the empty state opens the Presets tray.
- [ ] The tray slides in above the composer with an ease-out animation on desktop.
- [ ] Tray header has the title "Presets" on the left, a search input in the middle, and an "×" close button on the right.
- [ ] Below the header, the category tabs render in this order: "All", "Image edit", "Video tools", "Ads & products", "Audio".
- [ ] The active tab has a `bg-primary-cs-50` background and `text-primary-cs-700` text. Inactive tabs are text-only.
- [ ] The grid is `auto-fill` with a minimum 140 pixels per card and an 8-pixel gap.
- [ ] Each preset card renders: a 1.4:1 aspect-ratio thumbnail with the emoji centered, the preset name (font-medium 12.5 pixels), a mode badge ("IMAGE" or "VIDEO" — uppercase 9 pixels, zinc-200 bg, zinc-600 text), and a truncated description.
- [ ] Cards from a non-current mode render at 50% opacity but remain clickable.
- [ ] All 16 presets are available, with the names, icons, slots, and required flags exactly as specified:
  - **Image edit:** Face swap (🔄, source_face + target_image, both required), Inpaint (✏️, subject + mask, both required), Outfit swap (👗, subject + style_ref, both required), Style transfer (🎨, subject + style_ref, both required), Remove background (🌫, subject required), Expand image (📐, subject required), Upscale (🔍, subject required).
  - **Ads & products:** Headshot (🪞, subject required), Product shot (📸, subject required), URL → Ad (🔗, URL input required, no slots).
  - **Video tools:** First/last frame (🎞, first_frame + last_frame, both required), Extend video (⏩, base_video required), Motion sync (🕺, motion_video + motion_target, both required), Replace character (🎭, base_video + character_ref, both required), Lip sync (💋, subject + audio, both required).
  - **Audio:** Add sound effect (🔊, base_video required).
- [ ] Each preset card's description matches the spec exactly (e.g., Face swap → "Replace a face in any image"; Inpaint → "Edit only a masked region"; etc.).
- [ ] The search input in the header filters preset cards across name and description in place.
- [ ] Clicking a preset card whose mode matches the current mode applies the preset silently.
- [ ] Clicking a preset card whose mode differs from the current mode opens a confirm dialog with the copy: "Switch mode? {preset name} is a {target mode} preset. Switch from {current mode} to {target mode}?". On confirm, the preset applies. On cancel, nothing changes.
- [ ] Applying a preset sets mode, sets the model (locked), reshapes slots, and shows a toast: "Applied preset: {preset name}".
- [ ] When a preset is active (State C), the Mode chip is replaced by a Preset chip showing the preset's emoji and name.
- [ ] The Preset chip has an "×" on the right that, when clicked, drops back to State B (plain mode) — model unlocks, slots update to plain-mode slots, filled slots not in the new state are dropped.
- [ ] When a preset is active, the Model pill renders with a Lock icon prefix and amber styling, and is non-clickable.
- [ ] When a preset is active, send is disabled until all required slots are filled.
- [ ] For the `URL → Ad` preset, instead of a slot row, a single full-width URL input renders below the top control row with the label "Product URL" and the placeholder "Paste a product URL (e.g., https://your-store.com/product/...)".
- [ ] For `URL → Ad`, send is disabled until the URL input has a non-empty value that begins with `http://` or `https://`.
- [ ] On send for `URL → Ad`, the chat shows a "Generating ad… this can take up to a minute" loading state while the backend handles URL scraping and generation.
- [ ] When the user applies any preset, a Usermaven event `ai_studio_preset_applied` fires with `{ preset_key, mode }`. Search `contentstudio-frontend/src/` for `userMaven.track(` first to confirm no equivalent event already exists; if it does, reuse that event name.

### Mock-ups

- Lovable prototype: https://ai-studio-cs.lovable.app/

### Impact on existing data

- AI Chat client state extends with `preset` (preset key string or `null`) and `presetUrlInput` (string, used by URL → Ad).
- The `mediaDetails` payload sent to `sendChatMessage` extends with `preset` (the preset key) and `presetUrlInput` (when applicable).

### Impact on other products

- Web-only feature.
- White-label theming: amber locked-state styling and blue active-tab styling must use theme-aware tokens; pair with the Product designer if alignment is needed.
- All copy goes through `$t()` and into every locale directory.

### Dependencies

- Depends on: `[FE] Build collapsible AI Studio composer with Image/Video mode and redesigned model picker` (Mode chip → Preset chip swap; Model pill locked state).
- Depends on: `[FE] Add reference slot system to AI Studio composer with contextualized Add Media modal` (slot row reshaping per preset).
- Coordinates with: `[FE] Add validation banners, toasts, and confirm dialogs to AI Studio composer` (cross-mode confirm dialog; preset-applied toast).
- Coordinates with: `[FE] Add mobile-web responsive bottom sheets and compressed layout to AI Studio composer` (Presets tray becomes a bottom sheet on mobile).

### Global quality & compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories)
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled)
- [ ] UI theming supported (default + white-label, design library components are being used)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension) — web-only

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- **Suggested new components:** `src/modules/AI-tools/components/composer/PresetTray.vue`, `PresetTrayContent.vue`, `PresetCard.vue`, `PresetChip.vue`, `PresetUrlInput.vue`. A `usePresets.ts` composable can centralize the preset catalog and the apply/clear logic.
- **Preset catalog source:** the 16 presets and their metadata (key, name, icon, mode, slots, required) are best stored as a typed const in `src/modules/AI-tools/components/composer/presets.ts`. Keep the model assignment OUT of this file — engineering owns model selection and the assignment can live in a separate config or be resolved at runtime.
- **Capability requirements (engineering reference):** Each preset has capability requirements that the chosen model must satisfy:
  - face_swap: Multi-image composition; facial identity preservation
  - inpaint: Mask-based localized editing
  - outfit_swap: Clothing-aware editing with reference image
  - style_transfer: Reference image + style transfer
  - bg_remove: Foreground/background segmentation
  - expand_image: Outpainting beyond original canvas
  - upscale: Super-resolution
  - headshot: Photorealistic portrait generation; identity preservation
  - product_shot: Photorealistic product photography
  - click_to_ad: URL scraping + product video generation
  - first_last: Start frame + end frame inputs
  - extend_video: Video continuation from existing clip
  - motion_sync: Motion transfer (extract from video, apply to image)
  - replace_char: Character replacement in video with reference
  - lip_sync: Audio-driven lip animation
  - add_sound: Audio generation conditioned on video
- **i18n:** preset names and descriptions go under `ai_tools.composer.presets.{preset_key}.name` and `.description`. Category tab labels under `ai_tools.composer.presets.categories.*`. Add to every locale directory.
- **Existing pattern:** the existing `ActionSpecificPrompts` map in `src/modules/AI-tools/Prompts.js` is conceptually similar (key → metadata). Reusing the same shape may help.
- **Designer touchpoint:** preset card thumbnail (emoji size, gradient bg, rounded corners), tray slide-in animation, mobile bottom-sheet treatment (handled in the mobile story).

---

## Story 5 — `[FE] Add multi-shot video generation to AI Studio composer`

### Description

As a ContentStudio user generating video content, I want to define a video as multiple sequential shots (each with its own prompt and duration), so that I can craft narrative video content rather than being limited to single-prompt clips.

This story adds a visible Multi-shot toggle in the composer (placement to be aligned with the Product designer — surfaced directly, not buried inside any menu) that appears only when a multi-shot-capable video model is selected. When the toggle is ON, the single textarea is replaced with a vertical stack of "Shot N" cards, each with a per-shot duration dropdown and a textarea.

### Workflow

1. User has the composer in Video mode with a multi-shot-capable model selected. A labeled Multi-shot Switch appears visibly in the top control row area (next to other controls like Style and Quality), labeled "Multi-shot".
2. User toggles the Switch ON. The single textarea is replaced with a "Multi-shot" section. The section header reads "Multi-shot · 1 shot" on the left and "Total: 5s" on the right.
3. Below the header, one shot card renders: "Shot 1" label on the left, a duration dropdown showing "5s" (the default), a textarea with the placeholder "Describe the first scene…". No trash icon (only one shot).
4. User types into the Shot 1 textarea. Then clicks "+ Add shot" below the card. A second card appears: "Shot 2", duration dropdown "5s", textarea placeholder "Describe scene 2…". The header now reads "Multi-shot · 2 shots", "Total: 10s". Each card now shows a trash icon on the right.
5. User changes Shot 1 duration to "8s". The header total updates to "Total: 13s".
6. User keeps adding shots until they have 6. The "+ Add shot" button becomes disabled. Hovering it shows a tooltip: "Maximum 6 shots".
7. User clicks the trash icon on Shot 3. Shot 3 is removed. Shots 4, 5, 6 renumber to 3, 4, 5. The header total recalculates.
8. User toggles the Multi-shot Switch OFF. The textarea returns. All shot data is discarded silently (no confirm dialog — toggling off is intentional).
9. User switches to a model that doesn't support multi-shot. The Multi-shot Switch disappears entirely. If multi-shot was ON, the textarea returns and shot data is discarded.

### Acceptance criteria

- [ ] In Video mode with a multi-shot-capable model, a labeled Multi-shot `Switch` (from `@contentstudio/ui`) is visible in the composer's top control row area. Exact placement to be aligned with the Product designer.
- [ ] The Switch is labeled "Multi-shot".
- [ ] In Image mode, the Switch is hidden.
- [ ] In Video mode with a non-multi-shot model, the Switch is hidden.
- [ ] Switch defaults to OFF.
- [ ] When the Switch is OFF, the composer renders a single textarea (existing behavior).
- [ ] When the Switch is ON, the single textarea is replaced with a "Multi-shot" section.
- [ ] Section header reads "Multi-shot · {N} shot" when N is 1, "Multi-shot · {N} shots" when N is 2 or more, on the left.
- [ ] Section header shows "Total: {sum}s" on the right, where sum is the total of all shot durations.
- [ ] Section has a max-height of 280 pixels with internal scroll when shots overflow.
- [ ] Each shot card renders: shot number label ("Shot 1", "Shot 2", …) on the left, a duration dropdown, a trash icon on the far right (only when there is more than one shot), and a textarea below.
- [ ] The first shot's textarea placeholder reads "Describe the first scene…".
- [ ] Subsequent shot textareas have the placeholder "Describe scene {N}…".
- [ ] The duration dropdown options come from the active model's supported per-shot durations (engineering owns the catalog).
- [ ] A "+ Add shot" button renders below the last shot card.
- [ ] Clicking "+ Add shot" appends a new shot card with the model's default per-shot duration.
- [ ] When the shot count reaches 6, the "+ Add shot" button is disabled and shows a tooltip on hover: "Maximum 6 shots".
- [ ] Clicking the trash icon on a shot card removes it. Remaining shots renumber sequentially. Total duration recalculates.
- [ ] Toggling the Switch OFF discards all shot data silently and returns to the single textarea.
- [ ] Switching from a multi-shot-capable model to a non-multi-shot model while the Switch is ON automatically resets to single-textarea mode and discards shot data.
- [ ] Send disabled rules apply: if Multi-shot is ON, send is disabled until at least Shot 1's textarea has non-empty content (other shots may be empty — engineering may tighten this rule).
- [ ] On send when Multi-shot is ON, the metadata payload includes `multiShot: { shots: [{ duration: number, prompt: string }, ...], totalDuration: number }`.
- [ ] When the user toggles Multi-shot ON for the first time per session, a Usermaven event `ai_studio_multishot_enabled` fires with `{ model }` — confirm by searching `contentstudio-frontend/src/` for `userMaven.track(` first; reuse if equivalent exists.

### Mock-ups

- Lovable prototype: https://ai-studio-cs.lovable.app/

### Impact on existing data

- AI Chat client state extends with `multiShot: { enabled: boolean, shots: Array<{ duration: number, prompt: string }> }`.
- The `mediaDetails` payload sent to `sendChatMessage` extends to include `multiShot` when enabled. Backend integration approach (single multi-prompt call vs. one call per shot + stitch) is engineering's call.

### Impact on other products

- Web-only.
- Style/copy: all strings via `$t()`, all locales.

### Dependencies

- Depends on: `[FE] Build collapsible AI Studio composer with Image/Video mode and redesigned model picker` (Switch is anchored in the top-row shell built there).
- Depends on: `[FE] Add Ratio, Duration, and Quality pills to AI Studio composer top control row` (the top-row layout — Multi-shot Switch sits alongside Ratio / Duration / Style / Quality pills).

### Global quality & compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories)
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled)
- [ ] UI theming supported (default + white-label, design library components are being used)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension) — web-only

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- **Suggested new components:** `src/modules/AI-tools/components/composer/MultiShotEditor.vue`, `ShotCard.vue`, `MultiShotToggle.vue`. State lives in `useAIChatMessage.ts` (extend with `multiShot` ref).
- **`@contentstudio/ui`:** use the `Switch` component for the toggle. Use `Dropdown` / `DropdownItem` for the per-shot duration. Use `Button` (variant `ghost`) with a `Trash` icon for the per-shot remove.
- **Designer touchpoint:** placement of the Multi-shot Switch relative to the Style and Quality pills; whether it sits as a labeled Switch in the row or above the textarea as a section banner.
- **i18n:** new keys under `ai_tools.composer.multishot.*` (label "Multi-shot", section header singular/plural, Total prefix, "+ Add shot" button, "Maximum 6 shots" tooltip, shot 1 / shot N placeholders). Add to every locale directory.

---

## Story 6 — `[FE] Restyle Brand selector as prominent chip in AI Studio composer`

### Description

As a ContentStudio user, I want the Brand selector in the AI Studio composer to look and feel like a deliberate, weighty selection (rather than a thin utility button), so that I'm reminded which brand context my generation is being created for.

This story is a **visual restyle only** of the existing `BrandVoiceSelector` component as it appears in the AI Studio composer's bottom row. The popover behavior, brand kit integration, brand list, "None" option, and brand selection logic are unchanged.

### Workflow

1. User has the composer expanded in Image or Video mode. The bottom row shows a "+ Brand" chip with a dashed border (empty state).
2. User clicks the chip. The existing brand popover opens with "None" at the top and connected brand options below, each with a colored dot.
3. User selects a brand. The popover closes. The chip now renders as a solid white card with a subtle shadow, the brand's color dot on the left, and the brand name text. The chip's appearance is visually weightier than other bottom-row controls.
4. User clicks the chip again. The popover opens, showing the same options. The selected brand has a check next to it.
5. User selects "None". The popover closes. The chip returns to its "+ Brand" dashed-border empty state.

### Acceptance criteria

- [ ] In the empty state (no brand selected), the Brand chip renders with a dashed border (`border-dashed`), lighter text color, and the text "+ Brand". No fill, no shadow.
- [ ] In the selected state, the Brand chip renders as a solid white card (`bg-white`), with a subtle shadow (`shadow-sm` or equivalent), a colored dot on the left (using the brand's color), and the brand name text. Bigger visual weight than the Tier 2 subtle pills (Presets, Enhance) and Tier 3 bare icons (Attach, Saved, Mic, Send) — see `[FE] Apply three-tier chip styling to AI Studio composer bottom row`.
- [ ] Empty state has padding visually consistent with the other prominent chips (see Tier 1 spec in the bottom-row story).
- [ ] Selected state padding visually consistent with empty state.
- [ ] Brand color dot uses the brand's color from the existing brand kit data (preserved as-is).
- [ ] Clicking the chip in either state opens the existing brand popover. The popover content, the "None" option, the connected brand options, and selection logic are unchanged.
- [ ] When the user selects a brand from the popover, the chip transitions from empty to selected state.
- [ ] When the user selects "None", the chip transitions from selected to empty state.
- [ ] Hover states: empty-state chip on hover gets `bg-zinc-50` (or equivalent neutral). Selected-state chip on hover gets a slightly stronger shadow.
- [ ] Existing `BrandVoiceSelector` props, emits, and integration with `useAIChat` / brand kit composables are unchanged.

### Mock-ups

- Lovable prototype: https://ai-studio-cs.lovable.app/

### Impact on existing data

- No state shape changes. This is purely a visual restyle of the existing `BrandVoiceSelector` component.

### Impact on other products

- Web-only.
- This is the visual styling for the AI Studio composer's instance of the brand selector. Other consumers of `BrandVoiceSelector` (if any) should not be affected — wrap the new styling so it only applies in the AI Studio composer context, or apply via wrapper props.

### Dependencies

- Depends on: `[FE] Build collapsible AI Studio composer with Image/Video mode and redesigned model picker` (chip is anchored in the bottom-row shell).
- Coordinates with: `[FE] Apply three-tier chip styling to AI Studio composer bottom row` (Tier 1 prominent chip styling system; this story implements the chip itself).

### Global quality & compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories)
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled)
- [ ] UI theming supported (default + white-label, design library components are being used)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension) — web-only

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- **Primary entry point:** `contentstudio-frontend/src/modules/AI-tools/components/BrandVoiceSelector.vue`. Restyle the trigger element only; popover is unchanged.
- **Designer touchpoint:** exact shadow weight, padding, dot size, and selected-vs-empty contrast — pair with the Product designer to align with Tier 1 styling.
- **Shared with bottom-row story:** the Tier 1 prominent-chip token (padding, shadow, border treatment) should ideally be extracted into a shared CSS class or component slot used by both this story and the bottom-row story.

---

## Story 7 — `[FE] Apply three-tier chip styling system to AI Studio composer bottom row`

### Description

As a ContentStudio user, I want the AI Studio composer's bottom row to clearly distinguish between weighty selectors that affect generation (Brand), utility toggles (Presets, Enhance), and bare utility icons (Attach, Saved, Mic, Send), so that I can scan and act on the row instinctively.

This story applies the three-tier chip styling system (prominent / subtle / bare-icon) and the vertical-divider grouping defined in the spec to the bottom row built in Story 1.

### Workflow

1. User has the composer expanded in any mode. They look at the bottom row.
2. From left to right they see, separated by vertical dividers:
   - **Group 1 (bare icons):** the paperclip (📎) and the saved-prompts bookmark (🔖). Both render as 32×32 standard icon buttons with hover `bg-zinc-100`.
   - **Vertical divider** (1 pixel wide, 20 pixels tall, zinc-200, with horizontal margin).
   - **Group 2 (prominent chip):** the Brand chip (built and visually defined in `[FE] Restyle Brand selector as prominent chip in AI Studio composer`). Empty: dashed border + "+ Brand". Selected: solid white card + shadow + dot + name.
   - **Vertical divider.**
   - **Group 3 (subtle pills):** the Presets pill ("Presets" text-only inactive; `bg-primary-cs-50` and `text-primary-cs-700` when the tray is open) and the Enhance pill ("Enhance" with the same styling pattern; active when the toggle is ON).
   - **Spacer (flex-1).**
   - **Group 4 (bare icons, right-aligned):** the mic (🎤) and the send arrow (➤). 32×32 standard icon buttons.
3. User hovers each control. Tier 1 (Brand) gets a stronger shadow on hover. Tier 2 pills get `bg-zinc-100` on hover when inactive. Tier 3 icons get `bg-zinc-100` on hover.

### Acceptance criteria

- [ ] The bottom row renders, from left to right, the following groups separated by vertical dividers:
  - Group 1 (Tier 3 — bare icons): paperclip (📎), saved prompts (🔖).
  - Vertical divider.
  - Group 2 (Tier 1 — prominent chip): Brand chip.
  - Vertical divider.
  - Group 3 (Tier 2 — subtle pills): Presets pill, Enhance pill.
  - Spacer (flex-1).
  - Group 4 (Tier 3 — bare icons, right-aligned): mic (🎤), send (➤).
- [ ] Vertical dividers are `w-px h-5 bg-zinc-200 mx-0.5`.
- [ ] Tier 1 (Brand): empty-state styling = `border border-dashed text-zinc-500 px-3 py-1.5 rounded-md text-sm`; selected-state styling = `bg-white shadow-sm border border-zinc-200 px-3 py-1.5 rounded-md text-sm flex items-center gap-2`. Hover (selected): stronger shadow.
- [ ] Tier 2 (Presets, Enhance): inactive = text-only `text-zinc-700 hover:bg-zinc-100 px-3 py-1.5 rounded-md text-sm`; active = `bg-primary-cs-50 text-primary-cs-700 px-3 py-1.5 rounded-md text-sm`.
- [ ] Tier 3 (bare icons): `w-8 h-8 rounded-md hover:bg-zinc-100 flex items-center justify-center`.
- [ ] The Presets pill becomes "active" styling when the Presets tray is open (`bg-primary-cs-50` + `text-primary-cs-700`).
- [ ] The Enhance pill becomes "active" styling when the Enhance toggle is ON.
- [ ] All chip and pill styling uses theme-aware Tailwind classes — no hardcoded hex.
- [ ] Tier 1 (Brand) is visually weightier than Tier 2 pills, which are visually weightier than Tier 3 bare icons. Designer signs off on the relative weighting.
- [ ] Existing chip/pill behaviors (popover open/close, click handlers, tooltips on hover) are unchanged.

### Mock-ups

- Lovable prototype: https://ai-studio-cs.lovable.app/

### Impact on existing data

- No state shape changes.

### Impact on other products

- Web-only.

### Dependencies

- Depends on: `[FE] Build collapsible AI Studio composer with Image/Video mode and redesigned model picker` (the bottom-row shell is built there).
- Depends on: `[FE] Restyle Brand selector as prominent chip in AI Studio composer` (the Brand chip itself).
- Coordinates with: `[FE] Add 16 preset workflows to AI Studio composer` (Presets pill active state ties to tray open state).

### Global quality & compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — desktop styling here; mobile layout in `[FE] Add mobile-web responsive bottom sheets and compressed layout to AI Studio composer`
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled)
- [ ] UI theming supported (default + white-label, design library components are being used)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension) — web-only

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- **Primary entry point:** the bottom-row shell component built in Story 1, suggested name `src/modules/AI-tools/components/composer/ComposerBottomRow.vue`.
- **Suggestion:** extract the three tier styles into a small CSS-in-JS helper or shared composable (e.g., `useComposerChipStyles.ts`) so Tier 1, Tier 2, Tier 3 are defined once and reused. Keeps Story 6 (Brand chip) and this story consistent.
- **i18n:** "Presets" and "Enhance" labels already exist or go under `ai_tools.composer.bottom_row.*`. Add to every locale directory.
- **Designer touchpoint:** relative visual weighting between tiers, hover states, divider treatment.

---

## Story 8 — `[FE] Add Ratio, Duration, and Quality pills to AI Studio composer top control row`

### Description

As a ContentStudio user generating AI content, I want quick access to aspect ratio, video duration, and output quality directly in the top control row, so that I can dial in those parameters without opening sub-menus.

This story adds the Ratio pill, Duration pill (Video mode only), and Quality pill (1K / 2K / 4K) to the top control row. The Mode chip and Model pill are built in Story 1; the existing Style pill is preserved as-is. There is no More menu in Phase 1 — Variations is excluded entirely, and Multi-shot is built as a visible Switch in Story 5.

### Workflow

1. User is in Image mode. The top control row from left to right shows: Mode chip ("Image"), Model pill, Ratio pill, Style pill, Quality pill.
2. User clicks the Ratio pill. A small popover opens showing all aspect ratios supported by the current model (e.g., 1:1, 4:5, 9:16, 16:9, 3:4, 4:3). The pill's icon updates based on the active ratio: `Square` for 1:1, `Smartphone` for 9:16 or 3:4, `Monitor` for 16:9 or 4:3.
3. User picks 9:16. The popover closes. The Ratio pill text and icon update.
4. User switches to Video mode. The top control row now shows: Mode chip ("Video"), Model pill, Ratio pill, Duration pill (new), Style pill, Quality pill.
5. User clicks the Duration pill (currently "5s"). A popover opens with the durations supported by the current video model (e.g., 5s, 10s). User picks 10s. The popover closes; pill updates.
6. User clicks the Quality pill. A popover opens with the quality options the current model supports (1K, 2K, 4K). The currently selected one has a check. User picks 4K. Pill updates.
7. User switches to a model that doesn't support 4K. The Quality pill auto-resets to the highest quality the new model does support (e.g., 2K). A toast renders ("Quality changed to 2K — new model doesn't support 4K", rendered by the validation story).

### Acceptance criteria

- [ ] In Image mode, the top control row from left to right is: Mode chip, Model pill, Ratio pill, Style pill, Quality pill (and the Multi-shot Switch from Story 5 is hidden in Image mode).
- [ ] In Video mode, the top control row from left to right is: Mode chip, Model pill, Ratio pill, Duration pill, Style pill, Quality pill, Multi-shot Switch (when supported, per Story 5).
- [ ] Ratio pill shows the current ratio (e.g., "1:1", "9:16") with a `ChevronDown` indicator and a leading icon: `Square` for 1:1, `Smartphone` for 9:16 or 3:4, `Monitor` for 16:9 or 4:3.
- [ ] Clicking the Ratio pill opens a popover listing all ratios supported by the current model.
- [ ] The currently selected ratio in the popover has a check icon and `bg-primary-cs-50` background.
- [ ] Selecting a ratio closes the popover and updates the pill.
- [ ] When the active model changes and the current ratio isn't supported, the ratio resets to the new model's first supported ratio.
- [ ] In Video mode, the Duration pill renders to the right of Ratio. It shows the current duration (e.g., "5s") with a `ChevronDown` and a `Clock` leading icon.
- [ ] Clicking the Duration pill opens a popover with all durations supported by the current video model.
- [ ] Selecting a duration closes the popover and updates the pill. Reset behavior on model change matches the Ratio pill.
- [ ] The Style pill is preserved as-is from the existing `MediaGenerationOptions` component — no behavior or styling changes from this story (visual chip styling alignment is allowed if needed for consistency, but functionality is unchanged).
- [ ] The Quality pill shows the current quality (e.g., "1K", "2K", "4K") with a `ChevronDown`.
- [ ] Clicking the Quality pill opens a popover listing 1K / 2K / 4K (limited to qualities supported by the current model — engineering owns the model catalog).
- [ ] Selecting a quality closes the popover and updates the pill.
- [ ] When the active model changes and the current quality isn't supported, the quality auto-resets to the highest the new model does support.
- [ ] All pill chrome (padding, border, hover, ChevronDown placement) is consistent with the Mode chip and Model pill from Story 1.
- [ ] All popovers close on Escape.
- [ ] No More popover renders in Phase 1.
- [ ] Variations control is not present anywhere in the composer.

### Mock-ups

- Lovable prototype: https://ai-studio-cs.lovable.app/

### Impact on existing data

- AI Chat client state extends with `quality` (string: "1K" | "2K" | "4K"). `ratio` and `duration` already exist in the current `mediaDetails` and are preserved.
- The `mediaDetails` payload extends with `quality`.

### Impact on other products

- Web-only.
- The existing `MediaGenerationOptions` component currently houses the Aspect Ratio dropdown — that dropdown is replaced by the new Ratio pill in this story. Other consumers of `MediaGenerationOptions` (if any) must continue working — confirm during implementation.

### Dependencies

- Depends on: `[FE] Build collapsible AI Studio composer with Image/Video mode and redesigned model picker` (Mode chip, Model pill, top-row shell).
- Coordinates with: `[FE] Add multi-shot video generation to AI Studio composer` (Multi-shot Switch placement next to Style/Quality).
- Coordinates with: `[FE] Add validation banners, toasts, and confirm dialogs to AI Studio composer` (auto-reset toast for quality on model change).

### Global quality & compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — mobile layout collapses these pills into the Options bottom sheet (see `[FE] Add mobile-web responsive bottom sheets and compressed layout to AI Studio composer`)
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled)
- [ ] UI theming supported (default + white-label, design library components are being used)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension) — web-only

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- **Primary entry points:** `contentstudio-frontend/src/modules/AI-tools/components/MediaGenerationOptions.vue` — the existing aspect-ratio dropdown lives here. Refactor it into the new Ratio pill, or extract just the ratio data-fetching logic and use it in the new pill.
- **Suggested new components:** `src/modules/AI-tools/components/composer/RatioPill.vue`, `DurationPill.vue`, `QualityPill.vue`.
- **`@contentstudio/ui`:** use `Dropdown` + `DropdownItem` for each popover. Use `Icon` for ChevronDown / Square / Smartphone / Monitor / Clock. Pill chrome is inline Tailwind, theme-aware.
- **i18n:** `ai_tools.composer.ratio.*`, `ai_tools.composer.duration.*`, `ai_tools.composer.quality.*`. Add to every locale directory.
- **Designer touchpoint:** consistency of pill chrome with Mode chip and Model pill from Story 1.

---

## Story 9 — `[FE] Add mobile-web responsive bottom sheets and compressed layout to AI Studio composer`

### Description

As a ContentStudio user accessing AI Studio on a phone-sized browser, I want a layout that collapses the top control row's pills into a single Options sheet and the Presets tray into a bottom sheet, so that the composer is usable without horizontal scroll on small viewports.

This story is **mobile-web only** (AI features are web-only per `docs/story-guidelines.md` section 3 — there are no iOS/Android stories). It defines the responsive breakpoint behavior, the Options bottom sheet, the Presets bottom sheet, and the compressed bottom row.

### Workflow

1. User opens AI Studio in a phone-sized browser. The composer is in the resting state — a single-line input.
2. User taps the textarea. The composer expands to mode-active.
3. The top control row on mobile collapses to: Mode chip (with "×") + a "⚙ Options" pill. All other top-row pills (Ratio, Duration, Style, Quality, and the Multi-shot Switch) are NOT inline.
4. User taps the "⚙ Options" pill. A bottom sheet slides up from the bottom of the screen (`rounded-t-3xl`, drag handle at the top, scrollable content, max-height 80vh). The sheet is titled "Options" with an "×" close button.
5. The sheet content lists the row controls as `SheetRow` items: Aspect ratio, Duration (Video mode only), Style, Quality, Multi-shot toggle (when supported). Each row shows the current value on the right.
6. User taps "Aspect ratio". An inline picker expands within the sheet (or a sub-popover, engineering's call). User picks 9:16 and dismisses.
7. User taps the close "×". The sheet slides down and closes.
8. User taps the "Presets" pill in the bottom row. The Presets tray opens as a bottom sheet (same `rounded-t-3xl` styling, drag handle, close button) instead of sliding above the composer.
9. The bottom row on mobile is compressed to: paperclip + Brand chip + spacer + credits indicator + mic + send. Saved prompts, Presets pill, and Enhance pill are NOT inline — they live inside the Options bottom sheet.
10. User rotates the device to landscape (≥640 pixels wide). The composer transitions to the desktop layout.

### Acceptance criteria

- [ ] At a viewport width less than 640 pixels (`sm` breakpoint per Tailwind), the composer renders in mobile mode.
- [ ] At a viewport width 640 pixels or more, the composer renders in desktop mode.
- [ ] In mobile mode, the top control row renders only: Mode chip (with "×") and an "⚙ Options" pill on the right.
- [ ] Tapping the "⚙ Options" pill opens an `OptionsBottomSheet`.
- [ ] The bottom sheet has `rounded-t-3xl`, a drag handle at the top, a title bar with "Options" on the left and an "×" close button on the right, and scrollable content.
- [ ] The bottom sheet has `max-h-[80vh]`.
- [ ] The bottom sheet content is a stack of `SheetRow` items in this order: "Aspect ratio", "Duration" (only when Video mode), "Style", "Quality", "Multi-shot" toggle (only when Video mode + multi-shot-capable model).
- [ ] Each `SheetRow` shows the row's current value on the right.
- [ ] Tapping a `SheetRow` opens an inline sub-picker (or sub-popover) for that control. Engineering picks the exact mechanism.
- [ ] Tapping the "×" closes the sheet.
- [ ] Tapping the backdrop closes the sheet.
- [ ] Pressing Escape closes the sheet (where applicable on mobile-web with a connected keyboard).
- [ ] No swipe-down-to-dismiss gesture and no long-press on chips in Phase 1 — sheets close via "×" or backdrop only.
- [ ] In mobile mode, the Presets tray opens as a `PresetsBottomSheet` with the same chrome (rounded top, drag handle, title, close).
- [ ] The Presets bottom sheet's content is the same as the desktop tray content (header, search, category tabs, preset grid).
- [ ] In mobile mode, the bottom row from left to right is: paperclip (📎), Brand chip, spacer, credits indicator (when in Video mode with credits-tracked models), mic (🎤), send (➤).
- [ ] In mobile mode, the Saved prompts, Presets pill, and Enhance pill are NOT inline — Saved prompts is accessible via the existing global Saved Prompts modal trigger; Presets is accessible via the empty-state "Browse presets" button or the new Options sheet (engineering's call) or — preferred — added to the Options sheet as an additional `SheetRow` "Presets". Confirm with the Product designer.
- [ ] In mobile mode, the slot row uses the same `flex flex-wrap` layout but slot tiles wrap naturally to 2-up at narrower widths.
- [ ] The Add Media modal renders full-screen on mobile (existing component is already mobile-responsive; preserve).
- [ ] The composer's expansion logic from Story 1 still applies on mobile — bottom sheets keep the composer expanded while open.
- [ ] The model picker on mobile may render as a popover or as a full-screen sheet — engineering's call.

### Mock-ups

- Lovable prototype: https://ai-studio-cs.lovable.app/ — open in a phone-sized viewport for the mobile mockup.

### Impact on existing data

- No state shape changes. The Options sheet reads and writes the same state as the desktop top-row pills (ratio, duration, style, quality, multi-shot).

### Impact on other products

- Mobile-web only. No iOS or Android stories — AI features are web-only per `docs/story-guidelines.md` section 3.
- White-label theming applies to all sheet chrome.

### Dependencies

- Depends on: `[FE] Build collapsible AI Studio composer with Image/Video mode and redesigned model picker` (composer shell + expansion logic).
- Depends on: `[FE] Add Ratio, Duration, and Quality pills to AI Studio composer top control row` (the controls' state and behavior).
- Depends on: `[FE] Add 16 preset workflows to AI Studio composer` (Presets tray content, repackaged as a bottom sheet on mobile).
- Depends on: `[FE] Add multi-shot video generation to AI Studio composer` (Multi-shot toggle in the Options sheet).
- Depends on: `[FE] Apply three-tier chip styling system to AI Studio composer bottom row` (compressed bottom row reuses the chip styles).

### Global quality & compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — this story IS the mobile responsiveness work
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled)
- [ ] UI theming supported (default + white-label, design library components are being used)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension) — mobile-web only; native iOS/Android N/A

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- **Suggested new components:** `src/modules/AI-tools/components/composer/MobileBottomSheet.vue` (shared chrome), `OptionsBottomSheet.vue`, `PresetsBottomSheet.vue`, `SheetRow.vue`. The model picker on mobile could reuse `MobileBottomSheet` if a sheet is preferred over the desktop popover.
- **Existing pattern:** ContentStudio doesn't have a clear bottom-sheet primitive yet — pair with the Product designer on the chrome (drag handle, rounded radius, backdrop opacity) and confirm whether a new `BottomSheet` component is worth contributing to `@contentstudio/ui` (out-of-scope here; flag separately if it makes sense).
- **No gesture support in Phase 1** — no swipe-down to dismiss, no long-press. The bottom sheets close via the "×" button or the backdrop tap only. This is a classic mobile-responsive layout, not a gesture-rich experience.
- **Breakpoint:** Tailwind `sm` (640px) is the suggested cutover; engineering may prefer a different breakpoint after validating with the design.
- **i18n:** `ai_tools.composer.options_sheet.*` (sheet title "Options", row labels). Add to every locale directory.

---

## Story 10 — `[FE] Add validation banners, toasts, and confirm dialogs to AI Studio composer`

### Description

As a ContentStudio user, I want clear, immediate feedback when the composer drops a slot, blocks send, or is about to switch modes, so that I trust the composer not to silently lose my work or generate something I didn't want.

This story centralizes the cross-cutting validation, feedback, and confirmation layer used across the rest of the epic: the slot-type-mismatch banner, the send-disabled tooltip, the model-switch toast, the preset-applied toast, the cross-mode model-switch confirm dialog, and the cross-mode preset-application confirm dialog.

### Workflow

1. User somehow ends up with a video file in an image-only slot (debug only — UI prevents this normally). A red banner appears above the slot row: "Videos aren't supported in First frame." On the right, an "×" closes the banner.
2. User is in Image mode and tries to fill a slot that expects a video. The same banner appears, but with an extra link on the right: "Switch to Video". Tapping it switches the composer to Video mode and clears the banner.
3. User has the composer in State C (Face swap preset active) but hasn't filled the required slots yet. They hover the (disabled) send button. A tooltip shows: "Fill required slots: Face to use, Image to edit".
4. User changes the model to a different image model that doesn't support the currently filled `style_ref` slot. The slot is dropped silently. A toast slides in from the bottom: "Dropped 1 reference — not supported by {new model name}". The toast auto-dismisses after 3 seconds.
5. User clicks an Image preset while in Video mode. A confirm dialog appears: "Switch mode? Face swap is an image preset. Switch from video to image?". Buttons: "Cancel" (ghost) and "Continue" (zinc-900 primary). On Continue, the preset applies. On Cancel, nothing changes.
6. User picks a video model from the model picker while in Image mode. A confirm dialog appears: "Switch mode? {Model} is a video model. Switch from image to video? Your prompt will be preserved." Buttons: "Cancel" / "Continue". On Continue, mode switches, ratio/duration/style/quality reset to the new model's first values, slots not in the new state are dropped, prompt persists.
7. User applies a preset. A toast slides in: "Applied preset: Face swap". Auto-dismiss after 3 seconds.
8. User's auto-quality reset (from Story 8 — model doesn't support current quality) shows a toast: "Quality changed to 2K — new model doesn't support 4K". Auto-dismiss after 3 seconds.

### Acceptance criteria

- [ ] When a slot has a type-mismatch validation error, a red banner renders directly above the slot row.
- [ ] Banner background uses `bg-red-50` (or theme-equivalent), text `text-red-700`, border `border-red-200`, rounded-md.
- [ ] Banner copy: "Videos aren't supported in {Slot label}." for video-in-image-slot. "Images aren't supported in {Slot label}." for image-in-video-slot. "Audio isn't supported in {Slot label}." for audio mismatch.
- [ ] When the user is in Image mode and the slot expects video, the banner shows an additional `text-red-700 underline` link on the right reading "Switch to Video".
- [ ] Clicking the "Switch to Video" link switches the composer to Video mode and clears the validation banner.
- [ ] The banner has an "×" on the far right that closes the banner without changing mode.
- [ ] While the validation banner is shown, the send button is disabled.
- [ ] When send is disabled because of unfilled required slots, hovering the send button shows a tooltip: "Fill required slots: {comma-separated list of missing slot labels}". Example: "Fill required slots: Face to use, Image to edit".
- [ ] Toasts are positioned `fixed bottom-center` with `bg-zinc-900 text-white rounded-lg px-4 py-2` (or equivalent theme-aware tokens), and slide-in-from-bottom animation.
- [ ] Toasts auto-dismiss after 3000 milliseconds.
- [ ] Toast: "Dropped {N} reference{s} — not supported by {model name}" — `s` only when N is 2 or more — fires when the user changes models within the same mode and one or more filled slots are dropped.
- [ ] Toast: "Applied preset: {preset name}" — fires when the user applies a preset.
- [ ] Toast: "Quality changed to {new quality} — new model doesn't support {old quality}" — fires when the user changes models and the previous quality isn't supported.
- [ ] Confirm dialog has a black/40 backdrop and a centered white card (`max-w-sm rounded-2xl p-5`).
- [ ] Confirm dialog has a font-semibold title, body text, "Cancel" ghost button, and "Continue" zinc-900 primary button (using `Dialog` from `@contentstudio/ui`, or `Modal` if `Dialog` doesn't fit the chrome).
- [ ] Confirm dialog: cross-mode model switch — title "Switch mode?", body "{Model} is a {target mode} model. Switch from {current mode} to {target mode}? Your prompt will be preserved."
- [ ] Confirm dialog: cross-mode preset application — title "Switch mode?", body "{Preset name} is a {target mode} preset. Switch from {current mode} to {target mode}?"
- [ ] On Continue, the underlying state transition runs as specified by the originating story.
- [ ] On Cancel, no state changes.
- [ ] All toast/dialog/banner copy goes through `$t()` and is added to every locale directory.

### Mock-ups

- Lovable prototype: https://ai-studio-cs.lovable.app/

### Impact on existing data

- No new client state. This story consumes events/state from other stories (slot validation flag, model-switch deltas, preset-applied event, quality-reset event) and renders the corresponding UI. Implementation may centralize via a small `useComposerFeedback.ts` composable.

### Impact on other products

- Web-only.
- Use the existing `useAlertStore` toast pattern from `contentstudio-frontend/CLAUDE.md` where it fits; otherwise build a small bespoke toast for the visual treatment specified above. Engineering decides whether to extend `useAlertStore` styling or use a new wrapper.

### Dependencies

- Depends on: `[FE] Build collapsible AI Studio composer with Image/Video mode and redesigned model picker` (cross-mode model-switch dialog).
- Depends on: `[FE] Add reference slot system to AI Studio composer with contextualized Add Media modal` (slot-mismatch banner).
- Depends on: `[FE] Add 16 preset workflows to AI Studio composer` (cross-mode preset confirm dialog; preset-applied toast).
- Depends on: `[FE] Add Ratio, Duration, and Quality pills to AI Studio composer top control row` (quality-reset toast).
- Coordinates with: Story 1 for send-disabled tooltip on the send button.

### Global quality & compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories) — toasts, banners, and dialogs must render correctly on mobile-web
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled)
- [ ] UI theming supported (default + white-label, design library components are being used)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension) — web-only

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

- **Existing patterns:** `useAlertStore` (`src/stores/core/useAlertStore.ts`) is the standard ContentStudio toast mechanism — `.alert({ message, type })`. The visual treatment specified here (zinc-900 bg, white text, slide-from-bottom) may diverge from the default; pair with the Product designer to decide whether to extend `useAlertStore` or build a small bespoke `ComposerToast` for this surface.
- **Confirm dialog:** prefer `Dialog` from `@contentstudio/ui` if its chrome matches the spec; otherwise use `Modal`. Avoid `getCurrentInstance()` per CLAUDE.md — use `inject('root') as ModalPlugin` to access `$cstuModal`.
- **OPEN QUESTION (carry from spec):** generation-failure error taxonomy and retry copy. Spec section 21 calls this out as engineering-proposed. This story does NOT spec the failure surface; flag for follow-up.
- **OPEN QUESTION (carry from spec):** quota / credit warning UX (low credits, insufficient credits). Spec section 21. Out of this story; flag for follow-up.
- **i18n:** all banner / toast / dialog copy under `ai_tools.composer.feedback.*`. Add to every locale directory.
- **Designer touchpoint:** banner styling, toast styling vs. the default `useAlertStore` look, dialog button hierarchy.
