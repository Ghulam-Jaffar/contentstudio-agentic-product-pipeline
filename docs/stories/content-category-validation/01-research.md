# Research: Content Category Validation in Post Composer

## Current State

When a user selects "Add to Content Category" in the Posting Schedule section of the Post Composer but has not yet picked a category from the top-left sidebar dropdown, the current UX is:

- A tooltip-based error icon (`error-icon.svg`) appears **to the right of the "Posting Schedule" heading** inside the collapsible section header, powered by a `v-menu` hover popover.
- No red border appears on the Content Category dropdown in the sidebar.
- The "Add to Category" main action button in the footer is **not visually disabled** for the "no category selected" case (it's only disabled when a category is selected but has no available slots — `isNextSlotAvailable`).
- No error pill/badge appears in the footer bar.

### Error source
`postingScheduleErrors` is a computed property in `SocialModal.vue` (line ~1633):
```js
if (publishTimeOptions.time_type === 'content_category' && selectedContentCategory === null) {
  errors.push({ text: t('composer.errors.content_category_required') })
}
```

This is passed as `:posting-schedule-errors="postingScheduleErrors"` into `MainComposer.vue` (line 130 in SocialModal.vue), which uses it to conditionally show the error icon next to the heading (line 1280).

### Button disabled logic
`MainComposerFooter.vue` has:
- `isButtonDisabled`: checks `isDisabledPostingBtn || isNextSlotAvailable || isImageUploading || disableSchedule`
- `isDisabledPostingBtn`: checks `postHasErrors || processPlanLoader || postStatus === 'published'`
- `postHasErrors`: aggregates media/social errors — does **not** include `postingScheduleErrors`
- The "Add to Category" button button visual classes (`opacity-50`) are only driven by `isDisabledPostingBtn`, not `isButtonDisabled` — so even when `isNextSlotAvailable` is true, the button doesn't look visually dimmed.

## What Needs to Change

1. **Content Category dropdown (AccountSelectionAside.vue):** Apply a red border to the `CstDropdown` when `time_type === 'content_category'` and no category is selected. Requires a new computed error state to be passed from `MainComposer.vue` (which receives it from `SocialModal.vue`).

2. **Error banner below Posting Schedule (MainComposer.vue):** Replace the existing `v-menu` error icon tooltip next to the "Posting Schedule" heading with a full-width inline error banner placed immediately below the `<PostingSchedule>` component slot inside the `CstCollapsible`. Use `Alert` from `@contentstudio/ui`.

3. **Error badge/pill in footer (MainComposerFooter.vue):** Add a clickable red error pill `!(1)` to the left of the "Send for approval" button. Clicking it opens a `Modal` or `Dialog` showing the error message. Pass `postingScheduleErrors` as a new prop to `MainComposerFooter` via `footerData` in `SocialModal.vue`.

4. **Dim "Add to Category" button (MainComposerFooter.vue):** When `time_type === 'content_category'` and `postingScheduleErrors.length > 0`, apply `opacity-50 cursor-not-allowed` to the main action button and set `:disabled="true"`.

5. **Auto-clear:** Since errors are driven by `selectedContentCategory === null`, they clear automatically when the user picks a category — no separate clear logic needed.

## Files Involved

| File | Change |
|---|---|
| `contentstudio-frontend/src/modules/composer_v2/views/SocialModal.vue` | Add `postingScheduleErrors` to `footerData` computed (line ~2970); pass category error state to `MainComposer` |
| `contentstudio-frontend/src/modules/composer_v2/components/MainComposer.vue` | Remove old error icon tooltip (~line 1279–1304); add inline `Alert` banner below `PostingSchedule`; pass error state as prop to `AccountSelectionAside` |
| `contentstudio-frontend/src/modules/composer_v2/components/AccountSelectionAside.vue` | Apply `border border-red-500` to `CstDropdown` when category error is active |
| `contentstudio-frontend/src/modules/composer_v2/components/MainComposerFooter.vue` | Add error pill badge left of "Send for approval"; dim "Add to Category" button using `postingScheduleErrors` prop |
