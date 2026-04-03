# Stories: Filter Button Applied Count Badge

---

## Story 1: [FE] Show applied filter count badge on planner header filter button

---

### Description:

As a ContentStudio user, I want to see how many filters I've applied at a glance on the Planner's filter button, so that I know whether I'm viewing a heavily filtered or lightly filtered set of posts without opening the filter sidebar.

Currently, when filters are applied in the Planner, the filter button turns primary-colored but shows no count. The user has no way of knowing whether 1 filter or 5 filters are active without opening the sidebar.

Add a count badge to the filter button that shows the number of active filter categories. For example, if the user has filtered by statuses and labels, the badge shows "2".

**Key files:**
- `src/modules/planner_v2/views/MainPlanner.vue` — `areFiltersApplied` computed (line 633) already checks 8 filter query params. Add a new computed `appliedFiltersCount` that counts how many of those params have non-empty values, and pass it as a prop to `PlannerHeader`.
- `src/modules/planner_v2/components/PlannerHeader.vue` — add the count badge to the filter button in both mobile (line ~7) and desktop (line ~303) layouts.

**Filter categories counted** (from the existing `areFiltersApplied` logic):
- `statuses`
- `members`
- `created_by`
- `type`
- `labels`
- `campaigns`
- `content_category`
- `date`

---

### Workflow:

1. User opens the Planner — the filter button shows no badge (no filters applied).
2. User opens the filter sidebar and applies a status filter (e.g., "Scheduled") — the filter button turns primary-colored and a small badge appears showing "1".
3. User also applies a label filter — the badge updates to "2".
4. User also filters by content category — the badge updates to "3".
5. User clears one filter — the badge decrements accordingly.
6. User clears all filters — the badge disappears and the button returns to its default secondary style.

---

### Acceptance criteria:

- [ ] When one or more filter categories are active, a count badge appears on or next to the filter button in the planner header
- [ ] The badge shows the number of active filter categories (not the number of individual selected values — e.g., selecting 3 statuses counts as 1)
- [ ] The badge appears on both mobile and desktop layouts of the filter button
- [ ] When no filters are applied, no badge is shown
- [ ] The badge updates in real time as filters are added or removed (without needing to close/reopen the sidebar)
- [ ] The badge uses the `Badge` component from `@contentstudio/ui` with a primary color variant, or a small inline counter styled consistently with the existing button

---

### UI Copy:

**Badge content:** Just the number (e.g., "1", "2", "3"). No text label needed.

**Tooltip** (existing): The existing filter button tooltip ("Open filters") remains unchanged.

---

### Mock-ups:

N/A

---

### Impact on existing data:

None. This is a purely frontend visual change — no API or data model changes.

---

### Impact on other products:

None. Planner is web-only.

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — badge must appear on the mobile filter button as well
- [ ] Multilingual support — N/A, badge shows only a number
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---
---

## Story 2: [FE] Show applied filter count badge on inbox filter button

---

### Description:

As a ContentStudio user, I want to see how many filters I've applied on the Inbox filter button, so that I know at a glance whether my conversation list is filtered and by how many categories.

Currently, the filter button in the Inbox left sidebar (`FilterDrawer.vue`) turns primary-colored when filters are active via `hasFilterSelected`, but shows no count. The user has to open the filter drawer to see what's applied.

Add a count badge to the filter button that shows the number of active filter categories. For example, if the user has filtered by channel and tags, the badge shows "2".

**Key files:**
- `src/modules/inbox-revamp/components/FilterDrawer.vue` — compute a count from the selected filter state (channels, tags, assigned members, read/unread status, date, etc.) and display a badge on the filter `Button` (line ~3).

---

### Workflow:

1. User opens the Inbox — the filter button shows no badge (no filters applied).
2. User opens the filter drawer and selects a channel filter — the filter button turns primary-colored and a small badge appears showing "1".
3. User also applies a tag filter — the badge updates to "2".
4. User clears all filters via the "Clear" button in the drawer — the badge disappears and the button returns to its default style.

---

### Acceptance criteria:

- [ ] When one or more filter categories are active in the inbox, a count badge appears on or next to the filter button
- [ ] The badge shows the number of active filter categories (not individual values)
- [ ] When no filters are applied, no badge is shown
- [ ] The badge updates in real time as filters are added or removed
- [ ] The badge styling is consistent with the planner filter count badge (see **[FE] Show applied filter count badge on planner header filter button**)

---

### UI Copy:

**Badge content:** Just the number (e.g., "1", "2"). No text label needed.

---

### Mock-ups:

N/A

---

### Impact on existing data:

None. Purely frontend visual change.

---

### Impact on other products:

None. Inbox is web-only.

---

### Dependencies:

Depends on: **[FE] Show applied filter count badge on planner header filter button** — for consistent badge styling/pattern. Can be implemented in parallel but should match visually.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, inbox is not used on mobile viewports
- [ ] Multilingual support — N/A, badge shows only a number
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)
