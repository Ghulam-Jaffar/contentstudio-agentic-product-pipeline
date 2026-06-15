# Stories — Move Planner Card Accent Border to the Top Edge

One FE story. Frontend-only CSS change in Publisher → Planner (Calendar view). The Product Owner creates this in Shortcut manually using the New Feature Template.

---

## [FE] Move the planner calendar card's colored status accent to the top edge

### Description

As a user viewing my scheduled posts in the **Planner calendar**, I want the colored status accent on each post card to run along the **top** edge of the card instead of the **left** edge, so the status color reads as a strip across the top of the card and the cards look cleaner and more scannable.

This is a positional change only — the accent keeps reflecting the post's publishing status (e.g., green = published, amber = scheduled, red = failed) exactly as it does today; it simply moves from the left edge to the top edge.

---

### Workflow

1. The user opens **Publisher → Planner** and switches to **Calendar view**.
2. Each post card shows a colored accent indicating its publishing status.
3. The accent now appears as a strip along the **top** edge of the card, rather than down the left edge.
4. The accent color still matches the post's status (unchanged), and hovering the card still shows the existing status tooltip.
5. The same top accent appears in the "+N more" day popover and in the shared/external (public) planner calendar.

---

### Acceptance criteria

- [ ] In **Planner → Calendar view**, each post card's colored status accent appears along the **top** edge of the card.
- [ ] The accent no longer appears on the card's **left** edge; the left edge shows the same thin border as the right and bottom edges.
- [ ] The accent color still reflects the post's status (published / scheduled / draft / under review / failed / partially failed / rejected / processing / etc.) — the color mapping is unchanged.
- [ ] The top accent spans the card's full width between its rounded top corners and looks intentional — corners stay rounded, with no clipping or doubled-up border.
- [ ] The accent renders the same way in both the **compact** and **standard** card layouts.
- [ ] The accent appears on the top edge in the **"+N more" day popover**.
- [ ] The accent appears on the top edge in the **shared / external (public) planner calendar**.
- [ ] The card's existing status tooltip and all other card behavior are unchanged.
- [ ] No visual change to the Feed view or List view (they do not use this accent).

---

### Mock-ups

N/A — positional CSS change. The PO may attach before/after screenshots when creating the story.

---

### Impact on existing data

None. No schema, model, or API change. The accent reads the post's existing status, same as today.

---

### Impact on other products

- **Web only.** The iOS and Android apps have their own planner UI and are not affected — no mobile stories.
- **White-label:** the accent color is the post's status color (a fixed status palette), not a brand/theme token — unchanged. No new theme tokens.
- **Chrome extension:** not affected.

---

### Dependencies

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend) — verify the top accent on narrow calendar cells / smaller widths
- [ ] Multilingual support — N/A: no strings added or changed
- [ ] UI theming support (default + white-label) — accent uses the status color palette (not a brand token); verify no white-label regression
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension) — web only; mobile apps and Chrome extension unaffected

---

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

**The only change — calendar post card:**
- `contentstudio-frontend/src/modules/planner_v2/components/calendar-view/CalendarItemPost.vue`
  - Border classes on the `post-wrapper` root (L9): today `… border-r border-t border-b border-l-4! border-l-solid …`. Swap the 4px accent from left to top — make the **top** edge the 4px accent and the **left** edge the thin border, e.g. `… border-r border-b border-l border-t-4! border-t-solid …` (keep `rounded-lg`). Note: the existing 1px `border-t` must become the 4px `border-t-4!` (don't stack both), and add a 1px `border-l` so the left edge keeps a thin border.
  - Inline accent color (L14): change `border-left-color: ${borderStatusColor}` → `border-top-color: ${borderStatusColor}`.
  - `borderStatusColor` computed (L780) is unchanged — it still returns the status color via `getStatusColorCode()`.
  - The border lives on the shared `post-wrapper` root, so both the compact and standard layouts update together.

**No change needed (inherits automatically):**
- `CalendarEvent.vue` (calendar grid) and `ExternalCalendarEvent.vue` (shared/external calendar) both render `CalendarItemPost.vue`, so the change propagates to those surfaces and the "+N more" popover without separate edits.

**Gotcha:**
- Tailwind v4 uses a **trailing** important (`border-t-4!`), matching the existing `border-l-4!`. Confirm no other rule sets a conflicting top border on `.post-wrapper`.

---

### Shortcut fields

- **Template:** New Feature Template (PO selects this so the standard sections + 5 quality-checklist tasks are pre-populated)
- **Story type:** chore (frontend styling change)
- **Project:** Web App
- **Group:** Frontend
- **Epic:** Q2 - 2026: Miscellaneous
- **Priority:** Low (cosmetic polish)
- **Product area:** Planner
- **Skill set:** Frontend
- **Estimate:** _(leave empty — devs estimate during sprint planning)_
- **Labels:** _(none — team manages labels)_
- **Workflow state:** Ready for Dev
- **Iteration:** _(PO assigns the current/target sprint at creation time)_
