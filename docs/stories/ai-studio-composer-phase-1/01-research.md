# 01 — Research: AI Studio Composer Phase 1

**Epic:** [AI Studio Composer: Presets & UI/UX Update](https://app.shortcut.com/contentstudio-team/epic/117353) (id `117353`, state `to_do`)
**Phase:** 1 of 4 (UX overhaul + structured workflows; account/platform targeting and AI account intelligence are deferred to Phases 2–4)
**Source-of-truth spec:** `composer-context.md` (29 sections) at the project root
**Prototype:** `Composer.jsx` at the project root (Phase 2/3 UI — `AccountsDropdown`, `AccountsPickerContent`, ratio mismatch hint — is to be ignored)
**Reference mockup:** https://ai-studio-cs.lovable.app/

---

## Current State (existing AI Studio composer in `contentstudio-frontend`)

The product already has a working "AI Studio" chat composer. It lives in `src/modules/AI-tools/`, mounted from `AIChatMain.vue` and rendered through `ChatBox.vue`. The composer input itself is `src/components/dashboard/ChatInput.vue` (≈1480 lines), and the existing model/ratio/style/enhance controls live inside `src/modules/AI-tools/components/MediaGenerationOptions.vue` (≈990 lines). State is centralized in the `useAIChat` facade (`src/composables/useAIChat.ts`), which fans out to `useAIChatMessage`, `useAIChatActions`, and `useAIChatMedia`. Pinia singleton state lives in `useAIChatStore`.

What already works today (Phase 1 should preserve this and not re-build it):

- **Image/Video generation toggle** — exposed as a "Generate Media" dropdown in `ChatInput.vue` with Image/Video items and `X` to clear; backed by `isImageGenerationActive` / `isVideoGenerationActive` and `toggleImageGeneration` / `toggleVideoGeneration` in `useAIChatMessage`. This is the foundation we'll re-skin into the new "Mode chip."
- **Inline paperclip attachments** — file upload + drag/drop already attach images/videos as inline chips above the textarea. Image/video attachments go through `addImagesToMessage` / `addVideosToMessage`, with previews and an `X` to remove. The existing flow validates max-image limits and surfaces an "image required for image-to-video" amber banner. **Per user direction, this stays — Phase 1 does not rebuild the paperclip flow itself**, only the slot-context banner and type-filter behavior of the Add Media modal when invoked from a slot.
- **Empty-state quick-start buttons** — the empty chat state in `ChatBox.vue` already shows dashed-bordered buttons: "Write with AI", "Generate Image", "Generate Video", "Generate Hashtags", "Improve with AI". They wire into `handleGenerateImage` / `handleGenerateVideo` etc. **Per user direction, we update existing buttons rather than rebuilding the empty state.** "Browse presets" needs to be added.
- **Model picker** — `MediaGenerationOptions.vue` already has a provider-grouped model dropdown with NEW badges, model descriptions, generation-time badges, image-to-image / image-to-video support badges, and a checkmark for the active model. Phase 1 redesigns this into a search-first dropdown with featured section + capability badges + locked state.
- **Aspect ratio + style + enhance toggle** — already implemented (style chip is in the spec as preserved-as-is).
- **Brand selector** — exists as `BrandVoiceSelector.vue` in `src/modules/AI-tools/components/`. Phase 1 only restyles it to the new Tier 1 prominent-chip system; behavior is unchanged.
- **Saved Prompts** + **Audio mic / recording** + **Audio transcription** — already present as utility icons (`AudioInputButton.vue`, `AudioRecordingInterface.vue`, `WaveformVisualizer.vue`). Mic stays a placeholder per spec; saved prompts stays as-is.
- **Send button + chat message rendering** — `handleSend` in `ChatInput.vue` builds a `messageData` object via `getCurrentMessageForSending()` and passes it to `props.sendChatMessage(content, images, mediaDetails, aiLibraryPosts, videos)`. User messages render via `UserChatTemplate.vue`, AI responses via `BotChatTemplate.vue`. **Per user direction, send/generation flow is already handled — Phase 1 only extends the metadata payload to carry preset/slot/multi-shot data; no separate sending story.**
- **Drag-and-drop** — full drop-target on the chat area; preserved as-is.
- **Image lightbox** + **video preview tile** — preserved.
- **Video credits preview** — preserved.

What does **not** exist today and is Phase 1 scope:

1. **3-state collapsible composer** — composer is always expanded today. Spec calls for State A (single-line resting), State B (mode active, top + bottom rows visible), State C (preset active, model locked).
2. **Reference slots** — no concept of structured first/last frame, source face, mask, style ref, character ref, base video, motion video, audio slots.
3. **Add Media modal contextualization** — the existing Add Media modal (used by `composer_v2` and elsewhere) does not accept a slot-context banner or per-slot type filtering. A new optional `kind` / `slot` / `onSelect` API needs to be added.
4. **16 preset workflows** — no preset concept at all. Spec defines four categories (Image edit, Ads & products, Video tools, Audio) totaling 16 presets. Each preset locks a model and reshapes the slot row.
5. **@mention autocomplete** — no mention popover. Today users describe attachments inline in prose.
6. **Multi-shot video** — no shot stack UI. Only single-prompt video.
7. **Quality control (1K / 2K / 4K)** — confirmed by user as not present today. Variations stepper is explicitly deferred ("we don't need to do variations at the moment").
8. **Top control row chip system** — chips (mode chip, model pill, ratio, duration, style, More) need a unified pill styling with `ChevronDown` indicators, lock prefix on model when preset is active, and `X` clear on mode/preset chips.
9. **Bottom row three-tier chip system** — Tier 1 (Brand, prominent), Tier 2 (Presets, Enhance, subtle), Tier 3 (Attach, Saved, Mic, Send, bare icons), with vertical dividers separating the tiers.
10. **Mobile bottom sheets** — Options sheet (replaces the top control row's pills on mobile) and Presets sheet (replaces the slide-in tray on mobile). Existing dropdowns are desktop-style only.
11. **Validation banners + cross-mode confirm dialogs + toasts** — file-type mismatch banner with "Switch to Video" link; cross-mode confirm ("Switch mode? This is a video model…"); toast on dropped slots when switching models within a mode; toast on preset application.

---

## Prototype Component Inventory (`Composer.jsx`)

The prototype is in vanilla React + Tailwind. Component names below are **referenced in stories so engineering knows what each story builds** (the Vue equivalents will live under `contentstudio-frontend/src/modules/AI-tools/`):

| Prototype component | Phase 1 story | Vue target (suggested) |
|---|---|---|
| `Pill`, `IconButton`, `TextButton`, `Popover`, `PopoverItem`, `SectionLabel` | Story 1 (architecture) + Story 7 (bottom row) + Story 8 (top row) | Shared primitives in `src/modules/AI-tools/components/composer/` |
| `ComposerTop` | Story 1 + Story 8 | `ComposerTopRow.vue` |
| `ComposerBottom` | Story 1 + Story 7 | `ComposerBottomRow.vue` |
| `InputArea` | Story 1 | Refactor of `ChatInput.vue` with collapsible logic |
| `SlotRow`, `Slot` | Story 2 | `ReferenceSlotRow.vue`, `ReferenceSlot.vue` |
| `AddMediaModal` (slot-context props) | Story 2 | Existing Add Media modal extended with `kind` / `slot` / `onSelect` props |
| `MentionAutocomplete` | Story 3 | `MentionAutocomplete.vue` |
| `PresetTray`, `PresetTrayContent` | Story 4 | `PresetTray.vue`, `PresetTrayContent.vue` |
| `MultiShotEditor` | Story 5 | `MultiShotEditor.vue` |
| Brand chip styling (within `ComposerBottom`) | Story 6 | Restyle of existing `BrandVoiceSelector.vue` |
| Bottom row dividers + tiered styling | Story 7 | Within `ComposerBottomRow.vue` |
| `ModelPickerDropdown` | Story 1 (model picker is part of architecture story per user) | `ModelPicker.vue` (replaces existing `MediaGenerationOptions` model dropdown) |
| `BottomSheet`, `OptionsSheet`, `SheetRow` | Story 9 | `MobileBottomSheet.vue`, `OptionsBottomSheet.vue`, `PresetsBottomSheet.vue` |
| `ConfirmDialog`, `useToast` | Story 10 | Use existing `Dialog` from `@contentstudio/ui` + `useAlertStore` toast pattern |
| `AccountsDropdown`, `AccountsPickerContent` | **NOT PHASE 1** — Phase 2/3 |  — |
| Ratio mismatch hint UI | **NOT PHASE 1** — Phase 2 |  — |

Constants in the prototype (`MODELS`, `PRESETS`, `BRANDS`, `PLATFORMS`, `ACCOUNTS`) are placeholder data — engineering owns the real model catalog, and `PLATFORMS` / `ACCOUNTS` belong to deferred phases.

---

## What Needs to Change (mapped to the 10 stories)

### Story 1 — `[FE] Composer architecture (3-state collapsible) + Mode selection + Model picker`

Refactor `ChatInput.vue` to support three visual states; replace the existing "Generate Media" dropdown (`isImageGenerationActive` / `isVideoGenerationActive`) with the new Mode chip and Image/Video popover; rebuild model picker as a 380px-wide search-first dropdown with featured section, capability badges, and locked state. Update existing empty-state quick-start buttons (`Write with AI`, `Generate Image`, `Generate Video`) and add a new "Browse presets" button.

**Touches:** `ChatBox.vue`, `ChatInput.vue`, `MediaGenerationOptions.vue` (model dropdown portion), `useAIChatMessage.ts`.
**New components:** `ModelPicker.vue`, `ComposerTopRow.vue` (shell), `ComposerBottomRow.vue` (shell). Architecture story owns the *shells*; Stories 7/8 own the *visual styling system* of those rows.
**Edge case:** Blur delay (150ms) before collapsing must not collapse the composer when user clicks a popover — keep `shouldExpand` true while any popover is open.

### Story 2 — `[FE] Reference slots + Add Media modal contextualization`

Add the slot row that conditionally renders First/Last frame in plain Video mode (frame-capable models only) and preset slots in State C. Slot UI: empty (dashed), required (blue dashed + dot), filled (solid + thumbnail + filename + `X`). Coupled first/last frame with `→` arrow between them. Extend the Add Media modal to accept `kind: "slot"` with a blue context banner ("Adding to {Slot label} · Accepts: {types} · Max {size}MB · ≤{duration}s") and to dim incompatible files (40% opacity, grayscale, cursor-not-allowed).

**Touches:** Existing Add Media modal (consumed by `composer_v2` — care needed not to break that consumer). Likely lives in `src/modules/publish/components/media-library/` based on the Grep results.
**New components:** `ReferenceSlotRow.vue`, `ReferenceSlot.vue`.
**Edge case:** When switching models within a mode, drop filled slots not supported by the new model and show a toast (toast itself is in Story 10; this story emits the event).
**Slot definitions:** all 12 slot keys from spec section 6 (`subject`, `style_ref`, `mask`, `character_ref`, `first_frame`, `last_frame`, `source_face`, `target_image`, `base_video`, `motion_video`, `motion_target`, `audio`) with their max-size and accept-type constraints.

### Story 3 — `[FE] @mention autocomplete`

Add the `@`-trigger popover to the textarea (only fires when `@` is at start or preceded by whitespace). Lists current attachments only. Empty state has "No attachments yet" + "Click 📎 below to attach files, then mention them with @" + "Attach a file" button (opens Add Media in attachment mode). Insertion: `@filename_without_extension` (spaces → underscores) at cursor. Closes on Escape, click-outside, blur, or insert.

**Touches:** `ChatInput.vue` textarea logic.
**New components:** `MentionAutocomplete.vue`.
**Engineering note:** Mention-to-attachment linking format on send is engineering's call (pre-process tokens, separate fields, structured prompt — open to the team).

### Story 4 — `[FE] 16 preset workflows`

Build the Presets tray (slide-in above composer on desktop; bottom sheet on mobile per Story 9). Header: "Presets" + search + close. Category tabs: All · Image edit · Video tools · Ads & products · Audio. Grid of preset cards with emoji thumbnail, name, mode badge ("IMAGE" / "VIDEO"), description; cards from non-current mode at 50% opacity. Clicking a preset locks the model (amber banner + lock icon prefix on model pill), reshapes the slot row to the preset's slot list with required-marked slots, and disables send until required slots are filled. Cross-mode preset application shows a confirm dialog (dialog itself in Story 10; this story emits the trigger). Click `×` on the preset chip to drop back to plain mode.

**Touches:** `useAIChatMessage.ts` (add preset state), Mode chip → Preset chip swap in `ComposerTopRow.vue`.
**New components:** `PresetTray.vue`, `PresetTrayContent.vue`, `PresetCard.vue`, `PresetChip.vue`.
**16 presets:** as defined in spec section 9.1, all in scope. The `click_to_ad` preset (URL → Ad) replaces the slot row with a single URL input field; user pastes a product URL, clicks send, and the chat surfaces a loading state while the backend handles URL scraping and ad video generation. No special preset-time UI beyond the URL input.
**Capability requirements table:** preserved verbatim from spec section 9.2 in the story body so engineering can map presets to models.

### Story 5 — `[FE] Multi-shot video`

Add Multi-shot as a **visible top-row toggle** (Switch component, surfaced directly — not buried inside any More menu) that appears only when in Video mode + a multi-shot-capable model is selected. Exact placement to be aligned with the Product designer (likely sits next to the Style / Quality pills in the top control row, or as a labeled switch directly above the textarea). When ON, replace the single textarea with a vertical stack of "Shot N" cards: shot number label, duration dropdown, trash icon (when more than 1 shot), textarea. Section header: "Multi-shot · {N} shot{s}" + "Total: {sum}s". `+ Add shot` button (disabled at 6 shots max). When OFF, return to single textarea (discard shot data — no warning).

**Touches:** `useAIChatMessage.ts` (shot array state), `ComposerTopRow.vue` (toggle placement).
**New components:** `MultiShotEditor.vue`, `ShotCard.vue`, `MultiShotToggle.vue`.
**Send-flow extension:** The metadata payload on send needs to include `multiShot: { shots: [{ duration, prompt }], totalDuration }` — the existing `mediaDetails` shape extends.
**Designer touchpoint:** placement of the visible toggle relative to the other top-row controls.

### Story 6 — `[FE] Brand chip restyle`

Visual restyle only — empty state: dashed border + `+ Brand`; selected state: solid white card with subtle shadow + colored dot + brand name. Behavior, brand kit integration, and brand selection logic are unchanged.

**Touches:** `src/modules/AI-tools/components/BrandVoiceSelector.vue`.

### Story 7 — `[FE] Bottom row chip styling system`

Implement the three-tier chip styling system inside `ComposerBottomRow.vue`:
- Tier 1 prominent (Brand only in Phase 1, dashed→solid)
- Tier 2 subtle pills (Presets, Enhance — text-only inactive, blue-50 / blue-700 active)
- Tier 3 bare icons (Attach, Saved, Mic, Send — 32×32 rounded-md hover zinc-100)
- Vertical dividers (`w-px h-5 bg-zinc-200 mx-0.5`) separating the three groups

**Touches:** `ComposerBottomRow.vue` (shell built in Story 1).
**Note:** Brand styling specifically is also covered in Story 6; both stories will touch the Brand chip but with different concerns (Story 6 = the chip itself, Story 7 = its placement and the tier system around it). Coordinate via shared composable / props.

### Story 8 — `[FE] Top control row (Ratio + Duration + Quality)`

Implement the rest of the top control row: Ratio pill (icon-aware: `Square` / `Smartphone` / `Monitor`), Duration pill (video only), and a Quality pill (1K / 2K / 4K, supported values come from the active model). The Style pill is preserved as-is from the existing `MediaGenerationOptions` (user confirmed). The Mode chip and Model pill are built in Story 1. The Multi-shot toggle is built and placed in Story 5.

**Touches:** `ComposerTopRow.vue` (shell built in Story 1), `MediaGenerationOptions.vue` (existing aspect-ratio dropdown — refactor into the new chip styling).
**New components:** `RatioPill.vue`, `DurationPill.vue`, `QualityPill.vue`.
**Variations:** explicitly excluded — composer always generates one output per send.
**No More popover in Phase 1:** with Variations excluded and Multi-shot moved to a visible toggle (Story 5), Quality is the only remaining setting that would have lived under More — it's better placed directly in the row as its own pill.

### Story 9 — `[FE] Mobile adaptations`

On mobile the desktop top row collapses to a Mode chip + an "⚙ Options" pill that opens an `OptionsBottomSheet` (rounded-t-3xl, drag handle, scrollable, max-height 80vh) containing all the row's controls as `SheetRow` items. Presets tray becomes a `PresetsBottomSheet`. Bottom row compresses to `[📎] [Brand] ........ [credits] [🎤] [➤]` (Saved / Presets / Enhance move into the Options sheet). The model picker may remain a popover or become a full-screen sheet — engineering's call.

**Touches:** `ComposerTopRow.vue`, `ComposerBottomRow.vue`, `PresetTray.vue`.
**New components:** `MobileBottomSheet.vue` (shared shell), `OptionsBottomSheet.vue`, `PresetsBottomSheet.vue`.
**Edge case:** Mobile gestures (swipe to dismiss, long-press) **OPEN QUESTION** — flag in story.

### Story 10 — `[FE] Validation, error handling, toasts, confirm dialogs`

Centralize the validation + feedback layer:
- Red banner above slot row on file-type mismatch with "Switch to Video" link (only when in Image mode and a video was tried) and `X` to dismiss; send disabled while banner is up.
- Send-disabled tooltip listing missing required slots ("Fill required slots: Source face, Target image").
- Toasts: "Dropped {N} reference{s} — not supported by {model name}" on within-mode model switch; "Applied preset: {name}" on preset application.
- Confirm dialogs: cross-mode model switch ("Switch mode? {Model} is a video model. Switch from image to video? Your prompt will be preserved."); cross-mode preset application ("Switch mode? {preset} is a {mode} preset. Switch from {current} to {new}?").
- Generation-failure error states + insufficient-credit warning are spec-flagged as **OPEN QUESTIONS** — flag both in story; do not over-spec.

**Touches:** Wires into existing `useAlertStore` toast pattern + `Dialog` from `@contentstudio/ui` per CLAUDE.md conventions.

---

## UX Reference

Mockup: https://ai-studio-cs.lovable.app/ — the working Lovable prototype. Source-of-truth for exact visuals, transitions, and chip styling. Spec at `composer-context.md` is authoritative for behavior, copy, and edge cases. Where they diverge, the spec wins, except for visual details where the mockup is the latest source.

---

## Files Involved

### Existing files that will be touched

- `contentstudio-frontend/src/modules/AI-tools/ChatBox.vue` — empty-state buttons update (Story 1), conversation container untouched.
- `contentstudio-frontend/src/components/dashboard/ChatInput.vue` — major refactor for 3-state collapse, Mode chip, model picker, top/bottom row shells (Stories 1, 2, 3, 5, 7, 8). Likely needs to be split into smaller components since current size (1483 lines) exceeds the 500-line limit per `contentstudio-frontend/CLAUDE.md`.
- `contentstudio-frontend/src/modules/AI-tools/components/MediaGenerationOptions.vue` — model picker portion replaced (Story 1); aspect-ratio dropdown replaced (Story 8); style preserved as-is.
- `contentstudio-frontend/src/modules/AI-tools/components/BrandVoiceSelector.vue` — restyle only (Story 6).
- `contentstudio-frontend/src/composables/useAIChatMessage.ts` — extend message state for `preset`, `slots: { [key]: { type, file, url } }`, `multiShot: { enabled, shots: [{ duration, prompt }] }`, `quality`, slot-validation flag.
- The existing **Add Media modal** consumer in `src/modules/publish/components/media-library/` and `composer_v2/` — extend with optional `kind` / `slot` / `onSelect` props (Story 2). Care: do not break existing consumers.

### New files (suggested; engineering may choose otherwise)

```
contentstudio-frontend/src/modules/AI-tools/components/composer/
  ComposerTopRow.vue
  ComposerBottomRow.vue
  ModelPicker.vue
  ReferenceSlotRow.vue
  ReferenceSlot.vue
  PresetTray.vue
  PresetTrayContent.vue
  PresetCard.vue
  PresetChip.vue
  MentionAutocomplete.vue
  MultiShotEditor.vue
  ShotCard.vue
  RatioPill.vue
  DurationPill.vue
  MorePopover.vue
  QualitySelector.vue
  MobileBottomSheet.vue
  OptionsBottomSheet.vue
  PresetsBottomSheet.vue
```

(Component file count is intentional — `contentstudio-frontend/CLAUDE.md` mandates components stay under 500 lines, with logic extracted into composables. The architecture story will likely also produce a `useComposerState.ts` composable.)

---

## i18n Keys

Per `contentstudio-frontend/CLAUDE.md`, all user-facing strings go through `$t()` / `t()` and must be added to **every** locale directory under `src/locales/`. The existing AI Studio uses the `ai_tools` and `dashboard.chat_input` namespaces. New keys for Phase 1 should land under `ai_tools.composer.*` (preset names, slot labels, validation messages, dialog copy, toast copy, empty-state buttons).

---

## Engineering Ownership (PRD references — not specced in stories)

Per spec section 2 and per user direction:

- **Model catalog** (which models exist, credit costs, capability flags, max output, supported ratios/durations/styles, plan tier gating) — engineering picks.
- **Model selection per preset** — engineering picks the model that fulfills the spec section 9.2 capability requirements.
- **Default model per mode** — engineering picks.
- **Featured models per mode (3 in picker)** — engineering picks.
- **Multi-shot integration approach** (one call per shot + stitch, or multi-prompt single call) — engineering picks.
- **Backend payload shape for slots, multi-shot, attachments, @mention resolution** — engineering picks. Stories spec the user-facing behavior only.
- **Generation-failure error taxonomy** — open question per spec section 21; engineering proposes.
- **Plan tier gating + insufficient-credit warning UX** — open question per spec section 21.

---

## Open Questions (carried into stories)

The spec already calls these out (section 28). Carrying the still-relevant ones into the relevant stories:

1. Engineering effort estimate.
2. Preset model assignments + default models per mode + featured models — engineering's call but blocks story-level testing.
3. Plan tier gating per model.
4. Generation queueing (cancel, queue display, multiple in flight).
5. Accessibility audit — needs design + eng review.
6. Generation-failure error taxonomy + retry copy.
7. Quota / credit warning UX.

**Resolved:**
- `click_to_ad` preset is in scope — URL input + send + chat handles loading.
- AI-message action buttons (Save / Variations / Edit / Download) are already handled today; no Phase 1 story needed.
- Variations are explicitly out — composer generates one output per send.
- Pill/Chip component gap → engineering + Product designer pair on chip styling per story; no separate `[Design]` story.
- Mobile gestures (swipe-down to dismiss, long-press) are out of Phase 1 — classic mobile-responsive only; sheets close via "×" or backdrop tap.

---

## Mobile Context

**Not a mobile-app feature.** AI features are web-only per `docs/story-guidelines.md` section 3 ("Mobile apps have no AI features"). The "mobile" scope in this work is **mobile-web responsive layout** — i.e., the composer must work on phone-sized browsers. No iOS / Android stories.

---

## Cross-Product Impact

- **White-label theming:** must use `text-primary-cs-*` / `bg-primary-cs-*` (CSS-variable-backed) per `docs/story-guidelines.md` section 5. The prototype uses Tailwind zinc/blue/amber directly; stories will translate to ContentStudio's theme system.
- **i18n:** all new strings into every locale directory.
- **`@contentstudio/ui` components only:** `Dropdown`, `DropdownItem`, `ListItem`, `Modal`, `Dialog`, `Switch`, `Button`, `Icon`, `TextInput`, `SearchInput`, `Textarea`, `Badge`, `ActionIcon`, `Tabs`, `SegmentedControl`. There is no dedicated Pill/Chip component yet; chip primitives will be built inline with Tailwind. Stories will note "consult the Product designer where chip styling needs alignment" — engineering and design will pair on the visual treatment as needed.
- **Existing `Add Media` modal consumers** (`composer_v2`, media library views) must continue working unchanged when the new optional `kind` / `slot` props are absent.

---

## Risks / Things to Watch

1. **`ChatInput.vue` is already 1483 lines** — well over the 500-line cap in `contentstudio-frontend/CLAUDE.md`. Story 1 should plan for splitting it as part of the refactor, not after.
2. **Slot validation cross-cuts Stories 2 + 10.** Story 2 emits the validation event; Story 10 renders the banner. Coordinate via `useAIChatMessage` state.
3. **Brand chip cross-cuts Stories 6 + 7.** Story 6 owns the chip; Story 7 owns the row's tier system. Low risk if both stories ship together but worth noting.
4. **Multi-shot toggle placement (Story 5)** lives in the top control row built in Story 8. Story 5 must coordinate with Story 8 on placement; recommend Story 8 lands first or both ship together.
5. **Add Media modal extension** is a touch outside the AI Studio module — modify with care; the modal is consumed by `composer_v2` and the media library.
