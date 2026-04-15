# Stories — Schedule Time Manual Input
*Story Pipeline · Step 2 · Generated 2026-04-14*

---

## Story 1

**Title:** `[FE] Allow users to type hour and minute directly in time pickers`

---

### Description:

As a ContentStudio user, I want to be able to type the hour and minute directly into the time picker instead of being forced to scroll through long dropdown lists, so that I can set a precise time faster — especially when picking a minute value like 47 that would require scrolling almost to the bottom of a 60-item list.

---

### Workflow:

**Context:** This change affects two time picker components used across the product:

1. **Queue slot time picker** — Settings → Social Accounts → Account Details → Queue → add or edit a time slot for any day
2. **Evergreen Automation time picker** — Automation → Evergreen Automation → Custom Schedule → add or edit a time slot for any day
3. **Composer post scheduler time picker** — Composer → Post Scheduler modal → per-account date/time picker and same-time-for-all picker

**In the `TimePickerDropdown` component (queue slots & automation):**

1. User opens the time picker by clicking a time slot or the "+" add button on any day row.
2. A popover opens with the time picker UI.
3. At the top of the popover, the user now sees a **manual input row** with two small text fields side by side: one for the hour and one for the minute — both pre-filled with the currently selected time (e.g., `09` and `30`).
   - A colon `:` separator sits between them.
   - In 12h format, the AM/PM toggle buttons appear to the right of the minute input.
4. **Typing path:** The user clicks the hour input field, clears it, and types `14`. As they type, the hour column in the scrollable list below automatically scrolls to and highlights row `14`. When the user clicks into the minute field and types `05`, the minute column scrolls and highlights `05`. The user clicks **Update** (or **Add**) to confirm.
5. **Scrolling path:** The user ignores the input fields and scrolls the hour/minute columns as before, clicking a row to select it. The input fields at the top update to reflect the clicked value.
6. Both interactions are live-linked: editing either one updates the other in real time.

**In the `SelectTime` component (Composer):**

1. User opens the Composer and selects the date in the scheduler.
2. In the date picker footer, the time row shows the hour and minute as **clickable, editable input fields** instead of dropdown trigger buttons.
3. **Typing path:** The user clicks the hour field (e.g., showing `03`) and types `11`. The field accepts only the new typed value. Clicking the minute field and typing `45` updates the minute. The time is applied immediately on blur or Enter.
4. **Dropdown path:** The user clicks the chevron (▾) next to the hour or minute field to open the full scrollable list as before. Clicking a list item sets the value and closes the list.
5. Both paths remain available at all times.

---

### Acceptance criteria:

**TimePickerDropdown (Settings → Social Accounts queue slots & Automation → Evergreen schedule):**

- [ ] A manual input row with two text fields (hour and minute) appears at the top of the `TimePickerDropdown` popover, above the scrollable columns.
- [ ] The hour and minute input fields are pre-filled with the currently selected time when the popover opens.
- [ ] Typing a valid hour value into the hour input scrolls the hour column to that row and highlights it in real time.
- [ ] Typing a valid minute value into the minute input scrolls the minute column to that row and highlights it in real time.
- [ ] Clicking a row in the hour column updates the hour input field value.
- [ ] Clicking a row in the minute column updates the minute input field value.
- [ ] Hour input accepts values 0–23 in 24h format; 1–12 in 12h format.
- [ ] Minute input accepts values 0–59.
- [ ] On blur, single-digit values are padded to 2 digits (e.g., `9` → `09`, `5` → `05`).
- [ ] If an out-of-range value is typed (e.g., `99` for minute, `25` for hour in 24h mode), the input field border turns red and the Submit button is disabled with tooltip: "Please enter a valid time."
- [ ] In 12h format, the AM/PM toggle buttons appear to the right of the minute input in the manual entry row and stay in sync with the AM/PM column in the list below.
- [ ] The Cancel and Submit/Add buttons remain in place at the bottom of the popover; their behavior is unchanged.
- [ ] The scrollable hour and minute columns still work exactly as before for users who prefer clicking.

**SelectTime (Composer post scheduler):**

- [ ] The hour and minute displayed in the `SelectTime` component are rendered as editable input fields (not static text), each with a dropdown chevron (▾) to open the scrollable list.
- [ ] Clicking the hour input field allows the user to type a new hour value directly.
- [ ] Clicking the minute input field allows the user to type a new minute value directly.
- [ ] Typing a valid value and pressing Enter or blurring the field applies the value immediately.
- [ ] The hour/minute dropdown lists still open on chevron click and work as before.
- [ ] Selecting a value from the dropdown list updates the input field value.
- [ ] Hour input accepts 0–23 in 24h, 1–12 in 12h; minute input accepts 0–59.
- [ ] On blur, single-digit values are padded to 2 digits.
- [ ] Invalid values show a red border on the input; the selected time is not updated until corrected.
- [ ] AM/PM dropdown (12h mode) is unaffected.

---

### Mock-ups:

N/A — no mock-ups provided. UI must follow the existing time picker layout; the manual input row in `TimePickerDropdown` sits above the scrollable columns. In `SelectTime`, the existing CstDropdown trigger text becomes an editable input.

**`TimePickerDropdown` layout (updated):**

```
┌──────────────────────────────────┐
│  [ 09 ] : [ 30 ]    [ AM | PM ]  │  ← new manual input row (TextInput fields)
│  ────────────────────────────── │
│   Hour     │    Minute           │  ← scrollable columns (unchanged)
│  ┄┄┄┄┄┄┄   │  ┄┄┄┄┄┄┄           │
│    07       │    28               │
│  ▶ 09 ◀    │    29               │
│    10       │  ▶ 30 ◀            │
│    11       │    31               │
│  ────────────────────────────── │
│  [ Cancel ]        [ Update ]    │
└──────────────────────────────────┘
```

---

### Impact on existing data:

None. This is a UI-only change. No data model or API changes.

---

### Impact on other products:

- **Mobile apps (iOS / Android):** Not affected. These components are web-only.
- **Chrome extension:** Not affected.
- **White-label:** Uses `TextInput` from `@contentstudio/ui` and `text-primary-cs-500` / `border-primary-cs-200` theme classes — fully white-label compatible.

---

### Dependencies:

None.

---

### UI Copy

**Manual input row — `TimePickerDropdown`:**
- Hour input placeholder: `HH`
- Minute input placeholder: `MM`
- Colon separator between inputs: `:`
- In 12h mode, AM/PM toggle labels: `AM` / `PM` (unchanged)

**Validation messages (inline, below the input row):**
- Hour out of range (24h): `"Hour must be between 0 and 23"`
- Hour out of range (12h): `"Hour must be between 1 and 12"`
- Minute out of range: `"Minute must be between 0 and 59"`

**Submit button tooltip when disabled due to invalid input:**
- `"Please enter a valid time before saving."`

**`SelectTime` — no new labels needed.** The input placeholder is the current padded value (e.g., `09`, `30`). Error state uses a red border only (no inline message, to avoid layout shift in the compact footer).

---

### Component references:

- Use `TextInput` from `@contentstudio/ui` for the manual input fields in both components.
- In `TimePickerDropdown`, add `size="sm"` (or equivalent size prop) to keep the inputs compact in the popover header.
- Use `border-red-500` for the error border state on invalid input (neutral/error red — acceptable, not a theme color).
- Keep `DropdownItem` (from `@contentstudio/ui`) for the scrollable list rows — no changes there.
- `SelectTime.vue` uses legacy `CstDropdown` / `CstDropdownItem` — keep those for the list; replace only the trigger display text with a `TextInput`.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only)
- [ ] Multilingual support (add i18n keys for validation messages to all locale files under `src/locales/`)
- [ ] UI theming supported (default + white-label, design library components are being used)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)
