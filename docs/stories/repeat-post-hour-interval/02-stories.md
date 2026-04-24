# Stories: Add "Hour(s)" Option to Repeat Post Interval Dropdown

---

## [FE] Add "Hour(s)" interval option to the Repeat Post dropdown in the Composer

### Description:

As a user scheduling a post in ContentStudio, I want to be able to repeat my post at an hourly interval so that I can re-share time-sensitive content multiple times within a single day — without manually duplicating posts.

The "Hour(s)" interval option already exists in the ContentStudio Chrome extension's Repeat Post feature, and the backend already processes it. On the web, the option was hidden (commented out in the UI). This story makes it visible.

**What needs to be built:**
- In `contentstudio-frontend/src/modules/composer_v2/components/PostingSchedule/PostingSchedule.vue` (line 632), uncomment the `<option value="Hour">` element and wire it to the i18n key `composer.posting_schedule.intervals.hours`
- Add the `hours` translation key to the `posting_schedule.intervals` object in all 8 locale files under `src/locales/` (`en`, `de`, `el`, `es`, `fr`, `it`, `pl`, `zh`)
- No backend changes are required — `PostingController.php` and `SocialPostingController.php` already handle `repeat_type: 'Hour'` via `addHours($repeat_gap)`

---

### Workflow:

1. User opens the Composer to create or edit a post
2. User sets scheduling type to "Schedule" (or "Post Now") — the "Repeat Post?" toggle appears below
3. User enables the "Repeat Post?" toggle
4. The repeat options row appears with:
   - A **Repeat** count input (1–30 times)
   - A **"With the interval of"** row containing a number input and an interval type dropdown
5. User clicks the interval type dropdown — they now see **four options**: Hour(s), Day(s), Week(s), Month(s)
6. User selects **"Hour(s)"** and types their desired interval number (e.g., `2` for "every 2 hours")
7. The post timings preview below the form immediately updates to show the repeated post times spaced N hours apart
8. User clicks **Schedule** — the post is saved and the repeated copies are queued at the correct hourly intervals

---

### Acceptance criteria:

- [ ] "Hour(s)" appears as the **first option** in the interval type dropdown in the Repeat Post section (before Day(s), Week(s), Month(s))
- [ ] Selecting "Hour(s)" and setting the interval number to N correctly previews repeated post times N hours apart in the timings preview grid
- [ ] The interval type dropdown renders "Hour(s)" in all 8 supported locales — no missing translation keys or fallback to key path
- [ ] No interval validation error (`intervalError`) appears when "Hour(s)" is selected, regardless of the gap value — the Day-specific minimum (3 days) does not apply to hours
- [ ] The gap number input accepts values from 1–99 when "Hour(s)" is selected (same behaviour as other types)
- [ ] Switching between Hour(s) and other interval types (Day, Week, Month) updates the timings preview correctly and clears/resets any active validation errors

---

### Mock-ups:

N/A — This is a one-line UI change (un-hiding an existing option). The visual pattern already exists in the Chrome extension and mirrors the Day(s)/Week(s)/Month(s) options.

---

### UI Copy:

**Interval type dropdown options (in order):**
| Value | Label |
|---|---|
| `Hour` | Hour(s) |
| `Day` | Day(s) |
| `Week` | Week(s) |
| `Month` | Month(s) |

**i18n key to add:** `composer.posting_schedule.intervals.hours`

Translations to add across all locale files:

| Locale | Translation |
|---|---|
| `en` | `Hour(s)` |
| `de` | `Stunde(n)` |
| `es` | `Hora(s)` |
| `fr` | `Heure(s)` |
| `it` | `Ora/e` |
| `pl` | `Godzina/y` |
| `el` | `Ώρα/ες` |
| `zh` | `小时` |

**Note on component:** The interval `<select>` in `PostingSchedule.vue` is a native HTML `<select>` (not a `@contentstudio/ui Dropdown`). This is an existing pattern in this Options API component — no component upgrade is in scope for this story.

---

### Impact on existing data:

None. The `repeat_type: 'Hour'` value is already stored and processed by the backend. It has been the default value for `repeat_type` in `planner.js` store and the `Planning.php` library throughout the product's history. This change only makes the existing option selectable in the web UI.

---

### Impact on other products:

- **Chrome extension:** Already supports "Hour(s)" — no change needed
- **Mobile apps (iOS / Android):** The Repeat Post feature is not available on mobile apps — no impact
- **White-label:** No impact — the change uses the existing select element styling and no hardcoded colours are introduced

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)
