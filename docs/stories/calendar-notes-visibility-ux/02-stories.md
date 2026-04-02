# Stories: Calendar Notes Visibility UX Improvements

---

## Story 1: [FE] Improve calendar notes visibility UX with enhanced toast and settings indicator

### Description:

As a ContentStudio user, I want to know when my calendar notes are hidden so that I don't get confused when I create a note but can't see it on the calendar, and I want to see at a glance when calendar settings have been changed from their defaults.

Currently, when a user creates a note while the Notes toggle is off in the calendar settings dropdown, they see a generic "Note saved successfully." toast with no indication that notes are hidden. Additionally, the settings gear button in the calendar toolbar has no visual indicator when any setting (Notes, Holidays, Media Thumbnails, Compact View) is changed from its default state.

**Key files:**
- `contentstudio-frontend/src/modules/planner_v2/composables/usePlannerNotes.js` — `saveNote()` function (line 149), toast logic (lines 179–184)
- `contentstudio-frontend/src/modules/planner_v2/views/CalenderView.vue` — settings dropdown (lines 24–131), `settingsButton` custom button (lines 391–395), `toggleSettingsDropdown()` (line 2699)
- `contentstudio-frontend/src/modules/planner_v2/components/PlannerHeader.vue` — reference pattern for blue indicator (lines 302–311, `areFiltersApplied` prop)
- `contentstudio-frontend/src/stores/core/useProfileStore.ts` — `getCalendarNotesPreference` computed (line 154)

---

### Workflow:

1. User navigates to Planner → Calendar view
2. User opens the settings dropdown (gear icon in the calendar toolbar) and unchecks "Notes"
3. The settings gear button turns blue (primary color) with a small dot/badge indicator, signaling that a display setting has been modified
4. User clicks on a calendar date to create a new note
5. User fills in the note title and details, then clicks "Save"
6. User sees a success toast: **"Note saved successfully. Notes are currently hidden — enable them from Calendar settings to view."**
7. User opens the settings dropdown, re-enables "Notes"
8. The settings gear button returns to its default gray appearance (no indicator)
9. The newly created note now appears on the calendar

---

### Acceptance criteria:

- [ ] When a note is saved successfully and the Notes toggle is OFF (`getCalendarNotesPreference` is `false`), the success toast reads: "Note saved successfully. Notes are currently hidden — enable them from Calendar settings to view."
- [ ] When a note is saved successfully and the Notes toggle is ON, the existing toast is unchanged: "Note saved successfully."
- [ ] When a note is updated successfully while notes are hidden, the toast reads: "Note updated successfully. Notes are currently hidden — enable them from Calendar settings to view."
- [ ] The calendar settings button is replaced with a `Button` component (from `@contentstudio/ui`) styled identically to the Filters button in `PlannerHeader.vue`: `variant="outline"`, `color="primary"` with primary border/icon when any setting is off; `variant="text"`, `color="secondary"` with gray border when all settings are on
- [ ] The settings button shows a `Badge` component (from `@contentstudio/ui`) with a count of how many settings are turned off (e.g., "2" if Notes and Holidays are both unchecked)
- [ ] When all 4 settings toggles are checked/on, the settings button returns to its default gray appearance with no badge
- [ ] The blue indicator updates immediately when a user toggles any setting in the dropdown (no page refresh required)
- [ ] All new user-facing strings use `$t()` / `t()` with i18n keys and are added to all locale directories

---

### Mock-ups:

N/A — Follow the existing Filters button pattern from `PlannerHeader.vue` exactly:
- **Active state** (any setting off): `Button` with `variant="outline"`, `color="primary"` — gives primary-colored border and icon, matching the filters button when `areFiltersApplied` is true. Add a `Badge` showing the count of unchecked settings (e.g., "2").
- **Default state** (all settings on): `Button` with `variant="text"`, `color="secondary"` and gray border classes (`cstu-border cstu-border-gray-400/20`), no badge.
- Gear icon SVG fill: `rgb(var(--cstu-primary-500))` when active, `#4A4A4A` when inactive.

---

### Impact on existing data:

None. This is purely a frontend display change. No schema or API changes required.

---

### Impact on other products:

- **Mobile apps:** No impact — calendar settings are web-only
- **Chrome extension:** No impact
- **White-label:** No impact — uses CSS variable-based theming (`--cstu-primary-*`) for the blue indicator, compatible with white-label color overrides
- **Shared calendar links:** No impact — shared calendar view does not have the settings dropdown

---

### Dependencies:

None.

---

### UI Copy:

**Toast messages:**
- Hidden notes (new note): `"Note saved successfully. Notes are currently hidden — enable them from Calendar settings to view."`
  - i18n key: `planner.planner_notes.messages.note_saved_success_hidden`
- Hidden notes (updated note): `"Note updated successfully. Notes are currently hidden — enable them from Calendar settings to view."`
  - i18n key: `planner.planner_notes.messages.note_updated_success_hidden`

**Settings button tooltip** (unchanged): existing tooltip from `planner.calendar_view_main.calendar_controls.tooltips.calendar_preferences`

---

### Implementation guidance:

**1. Enhanced toast (`usePlannerNotes.js`):**
- Import `useProfileStore` in the composable
- In `saveNote()` success handler (line 179), check `profileStore.getCalendarNotesPreference`
- If `false`, use the `_hidden` variant i18n key instead of the standard one

**2. Settings button with badge (`CalenderView.vue`):**
- Replace the FullCalendar `settingsButton` custom button with a proper `Button` component (from `@contentstudio/ui`) rendered in the calendar header area (similar approach to how the settings dropdown is already positioned absolutely)
- Add a computed `disabledSettingsCount` that counts how many of the 4 toggles are off (`!defaultNotesToggle`, `!showHolidays`, `!showMediaPreview`, `!compactMode`)
- Add a computed `hasSettingsChanged` = `disabledSettingsCount > 0`
- When `hasSettingsChanged`: render `Button` with `variant="outline"`, `color="primary"` + a `Badge` component showing `disabledSettingsCount`
- When all on: render `Button` with `variant="text"`, `color="secondary"` and gray border, no badge
- Gear icon SVG fill toggles between `rgb(var(--cstu-primary-500))` and `#4A4A4A` based on `hasSettingsChanged`

**3. i18n keys:**
- Add `note_saved_success_hidden` and `note_updated_success_hidden` to `planner.planner_notes.messages` in all locale files under `src/locales/`

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only) — N/A, calendar settings gear button is desktop-only (calendar view has min-width 1200px)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
