# Stories: Calendar View Header Spacing & Border

---

## [FE] Fix excess white space and missing border between planner header and calendar view

---

### Description:

As a ContentStudio user, I want the calendar view in the Planner to look clean and consistent with the list and feed views, so that there's a clear visual boundary between the header and the calendar content without a large white gap.

Currently, when a user switches to calendar view, there is noticeable white space between the planner header and the calendar grid below it, and no border or divider separating them. In contrast, the list view has a subtle shadow (`shadow-sm` on the `DataTable` container) and the feed view has card borders that provide natural visual separation. The calendar view has neither.

**Two issues to fix:**

1. **Missing border/divider** — The `PlannerHeader` component (`src/modules/planner_v2/components/PlannerHeader.vue`) has no `border-b` class. Add a bottom border (`border-b border-gray-200`) to the header so there's a clean visual separator between the header and the content below it in all views. This will also benefit the feed view which currently relies solely on card spacing.

2. **Excess white space** — The header wrapper in `MainPlanner.vue` (line 20) has `p-[0.156rem]` padding, and the calendar view (`CalenderView.vue`) has no tight coupling to the header, resulting in a visible white gap. Reduce or eliminate the excess spacing so the calendar grid sits closer to the header, consistent with how the list view table sits directly under the header.

**Key files:**
- `src/modules/planner_v2/components/PlannerHeader.vue` — add `border-b border-gray-200` to the header container
- `src/modules/planner_v2/views/MainPlanner.vue` — adjust padding/spacing on the header wrapper and/or `#mainView` div to reduce the gap for calendar view

---

### Workflow:

1. User opens the Planner and selects Calendar view.
2. User sees the planner header (filters, view toggles, search) with a clean bottom border separating it from the calendar grid below.
3. The calendar grid sits close to the header — no large white gap between them.
4. User switches to List view — the header still has the bottom border; the table sits close below it as before.
5. User switches to Feed view — the header still has the bottom border; feed cards appear below it as before.

---

### Acceptance criteria:

- [ ] The planner header has a visible bottom border (light gray) in all three views (calendar, list, feed)
- [ ] In calendar view, the white space between the header and the calendar grid is reduced to match the spacing in list view
- [ ] In list view, existing layout and spacing are unchanged or improved (no regression)
- [ ] In feed view, existing layout and spacing are unchanged or improved (no regression)
- [ ] The border uses a neutral gray color (`border-gray-200`) that works across default and white-label themes

---

### Mock-ups:

N/A — this is a spacing and border fix. The expected result is a thin gray line below the header and tighter spacing to the calendar grid.

---

### Impact on existing data:

None. This is a purely visual/CSS change.

---

### Impact on other products:

None. The planner is web-only. The border is applied using neutral gray (`border-gray-200`) which does not change per white-label theme, so no white-label impact.

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — the planner header is responsive; border should apply on all breakpoints
- [ ] Multilingual support — N/A, no text changes
- [ ] UI theming support (default + white-label, design library components are being used) — border uses neutral gray, theme-safe
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)
