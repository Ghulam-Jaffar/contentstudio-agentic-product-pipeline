# Research: Calendar Notes Visibility UX Improvements

## Current State

The calendar view has a **settings dropdown** (gear icon in the FullCalendar header toolbar) that controls visibility of: Notes, Holidays, Media Thumbnails, and Compact View. Each is a checkbox toggle.

**Notes visibility** is controlled by `defaultNotesToggle` (backed by user preference `planner_default_notes_view` in the profile store). When unchecked, notes are hidden from the calendar — but there is **no indication anywhere** that notes are being hidden.

**Current toast on note creation** (from `usePlannerNotes.js:179-182`):
- Success: `"Note saved successfully."` (i18n key: `planner.planner_notes.messages.note_saved_success`)
- No mention of visibility status — if notes are hidden, user creates a note and sees the success toast but the note doesn't appear on the calendar.

**Settings dropdown button**: Custom FullCalendar button (`settingsButton`) with class `calendar-settings-button`. Currently has **no visual indicator** when any setting is changed from default. The button always looks the same regardless of toggle states.

**Existing blue indicator pattern** (Filters button in `PlannerHeader.vue:302-311`):
- When `areFiltersApplied` is true: `variant="outline"`, `color="primary"` — renders with primary (blue) outline
- When false: `variant="text"`, `color="secondary"` with gray border classes
- The SVG fill also changes: `rgb(var(--cstu-primary-500))` when active vs `#4A4A4A` when inactive

## What Needs to Change

### 1. Enhanced toast when notes are hidden
- When a note is saved successfully AND `defaultNotesToggle` is `false`, the toast should include guidance: something like "Note saved successfully. Notes are currently hidden — enable them from Calendar settings to view."
- The current toast message key `planner.planner_notes.messages.note_saved_success` needs a companion key for the hidden-notes variant.
- The `saveNote()` function in `usePlannerNotes.js` needs to check the notes visibility preference from `useProfileStore.getCalendarNotesPreference`.

### 2. Settings dropdown button blue indicator
- The `settingsButton` (FullCalendar custom button `.fc-settingsButton-button`) should visually change when **any** option in the dropdown is unchecked (i.e., differs from "all enabled" default).
- Need to apply primary color styling (like the filters button pattern) — blue icon/border when any setting is off.
- The settings dropdown has 4 toggles: Notes, Holidays, Media Thumbnails, Compact View. If any is unchecked, the button should turn blue with a visual indicator.

## Files Involved

| File | Change |
|---|---|
| `src/modules/planner_v2/composables/usePlannerNotes.js` | Check notes visibility pref, show enhanced toast when hidden |
| `src/modules/planner_v2/views/CalenderView.vue` | Add computed for "any setting changed", apply blue styling to settings button |
| `src/locales/en/planner.json` | Add new i18n key for hidden-notes toast message |
| `src/locales/{all-others}/planner.json` | Mirror new i18n keys |
