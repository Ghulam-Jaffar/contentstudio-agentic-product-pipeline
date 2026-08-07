# Research — Composer "Customize" control → single button with dropdown

Product-team feedback on the shipped AI Customize feature. Frontend-only change; the AI caption
API and generation logic already exist and are untouched.

## Current State

The customize control is a **grouped control** in the editor toolbar — an AI wand icon + caret on
the left, a vertical divider, then the word "Customize" with a `Switch` on the right.

- `contentstudio-frontend/src/modules/composer/components/EditorBox/editor-box/CustomizeAiControl.vue`
  — the grouped control (wand dropdown + divider + Switch).
- `.../editor-box/EditorBoxToolbar.vue` (~line 248) renders it when `toolbarConfig.customize !== false`,
  passing `separate-boxes`, `disabled`, `disabled-tooltip`, `on-toggle`.
- `.../editor-box/useEditorBoxDropdowns.ts` — `handleCustomBoxToggle(val)` emits `isSeparateBoxes`;
  `getCustomizeTooltip()` builds the toggle tooltip.
- `src/modules/composer/composables/useComposerAiCustomize.ts` — exposes `onGenerateClick()`
  (opens the intent dialog when the base caption is empty, otherwise generates straight away),
  `regenerateAll()`, `regeneratePlatform()`, `canGenerate` (≥ 2 platforms), `hasAiContent`,
  `isGenerating`, `trackMenuOpened()`.
- `src/stores/composer/useComposerAiCustomizeStore.ts` — `customizeOn`, `hasAiContent`,
  `statusByPlatform`, intent-dialog state.
- `src/modules/composer/components/AiCustomizeDialog.vue` — the empty-caption "Generate captions
  with AI" modal. **No change needed.**
- Copy lives in `src/locales/*/composer.json` under `composer.ai_customize.*` and
  `composer.editor_box.customize*` (8 locales).

**Today's wand menu:** section header "AI for Customize", a primary item that is
"Generate per-platform captions" before generation and "Regenerate captions for all platforms"
after, plus a secondary "Regenerate caption for {platform} only", plus a footer tip line.

**Where it renders:** the common editor box *and* every per-platform tab — the per-platform toolbar
configs spread the same defaults and don't set `customize: false`. `features/post-editor/types.ts`
does set `customize: false`, so the reusable Post Editor Surface (Evergreen modal) never shows it.

## What Needs to Change

- Replace the wand-icon + toggle group with **one "Customize" button that opens a dropdown**.
- **Customization off** → menu offers `Customize manually` and `Customize with AI`.
- `Customize manually` → the existing platform split, identical whether the composer is empty or
  has text (today's toggle-on behaviour).
- `Customize with AI` → empty composer opens the existing modal unchanged; composer with text
  splits and generates per-platform captions (today's behaviour).
- **Customization on** → menu offers `Regenerate captions with AI` and `Turn off customization`.
- No backend change, no change to the generation modal or the generated-caption states.

## Decisions confirmed with the PO

1. **Manual-on menu:** after a *manual* customize, the menu shows `Customize with AI` +
   `Turn off customization`, so a user who split by hand can still hand it over to AI. **Confirmed.**
2. **Per-platform regenerate:** the shipped "Regenerate caption for {platform} only" item is **not**
   part of the new menu and is out of scope for this change. The on-state menu stays at two items.
3. **Turn off customization** keeps today's behaviour — per-platform captions are discarded with no
   confirmation step. A confirm dialog would be a separate ask.

## Files Involved

- `src/modules/composer/components/EditorBox/editor-box/CustomizeAiControl.vue` (rewrite)
- `src/modules/composer/components/EditorBox/editor-box/EditorBoxToolbar.vue` (props/usage)
- `src/modules/composer/components/EditorBox/editor-box/useEditorBoxDropdowns.ts` (tooltip/toggle handler)
- `src/locales/*/composer.json` — all 8 locales (new menu labels; some old keys go unused)
