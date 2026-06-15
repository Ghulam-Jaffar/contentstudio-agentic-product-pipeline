# Stories — Content Category Pill Width in Planner

One FE story. Frontend-only CSS/markup bug fix in Publisher → Planner. The Product Owner creates this in Shortcut manually using the New Feature Template.

---

## [FE] Fix Content Category pill width on planner cards (fit-to-content)

### Description

As a user planning content in the Planner, I want the **Content Category** chip on each post card to be only as wide as its label text — just like the other chips on the card — so my cards look clean and consistent instead of showing a stretched, half-empty colored bar.

Today, on the planner post cards the Content Category chip can render the full width of the card with empty space on its right, making it look elongated and visually unbalanced compared to the content type, label, and campaign chips next to it. This story makes the chip hug its label with even padding on both sides, matching the other chips.

---

### Workflow

1. The user opens **Publisher → Planner** and views their posts in **Feed view** (the issue is also checked in Calendar and List views).
2. For any post that has a Content Category assigned, the card shows a small rounded chip with the category name, colored with that category's assigned color.
3. The chip is only as wide as its text, with equal padding on the left and right — it sits flush next to the other chips on the card (content type, labels, campaign) and reads as the same chip style.
4. If a category name is long, the chip truncates the text with an ellipsis within a sensible maximum width instead of overflowing the card or stretching it wider.

---

### Acceptance criteria

- [ ] In **Feed view**, the Content Category chip is only as wide as its label text (fit-to-content) — there is no empty trailing space stretching it across the card.
- [ ] The chip keeps its rounded-pill shape, white text, and the category's assigned background color.
- [ ] The chip's left and right padding are equal (symmetric).
- [ ] The chip's padding and pill style match the other chips on the same card (content type, labels, campaign), so the card reads as one consistent chip style.
- [ ] In **Calendar view** — including the "+N more" day popover — the Content Category chip is fit-to-content with symmetric padding.
- [ ] In the **shared / external planner calendar view**, the Content Category chip renders with the same fit-to-content style.
- [ ] In **List view**, the Content Category cell chip remains fit-to-content (no regression).
- [ ] A very long category name truncates with an ellipsis inside the chip's maximum width — it does not overflow the chip or force the card wider.
- [ ] The fix works for any category color (the chip background still reflects the category's assigned color).
- [ ] No change to which posts show the chip, the chip's tooltip text ("Content Category"), or any other card behavior.

---

### Mock-ups

N/A — visual fix; the reporter described the expected behavior in words. The PO may attach before/after screenshots when creating the story.

---

### Impact on existing data

None. No schema, model, or API change. The chip reads the existing `content_category` name and color exactly as it does today.

---

### Impact on other products

- **Web only.** The iOS and Android apps have their own planner UI and are not affected by this CSS change — no mobile stories.
- **White-label:** the chip background is the per-category color, not a brand/theme token, so it stays dynamic and unchanged. No new theme tokens introduced.
- **Chrome extension:** not affected.

---

### Dependencies

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend) — verify the chip on narrow calendar cells / smaller screen widths
- [ ] Multilingual support — N/A: no new strings (the chip label is the user's category name; the existing "Content Category" tooltip is unchanged)
- [ ] UI theming support (default + white-label) — chip uses the dynamic per-category color (not a brand token); verify no white-label regression
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension) — web only; mobile apps and Chrome extension unaffected

---

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary fix — feed view card:**
- `contentstudio-frontend/src/modules/planner/components/view/feed-view/FeedViewCard.vue` (L170–182) — the Content Category chip uses the class `top-category`. Its padding, `border-radius`, white text, and `display: inline-block` all come from legacy selectors scoped under `.planner_component .planner_feed_view .feed_box .top-category` (in this file's own `<style scoped>` ~L1649+ and the global `src/assets/styles/legacy/modules/composer/planner/_planner.css` ~L1963–1981). `planner_v2`'s Feed view reuses this legacy card with a root of just `.feed_box`, so the `.planner_component .planner_feed_view` ancestors never exist and those rules **don't match** — the chip falls back to a plain block `<div>` that fills the row (the elongated pill), keeping only the inline `background-color`. Give the chip a fit-to-content display, symmetric padding, and the pill shape directly on the element (e.g. Tailwind `inline-flex items-center w-fit px-2 py-1 rounded-full text-white`) and keep the inline `background-color`.

**Reference pattern (already correct — match it for consistency):**
- Calendar card chip: `contentstudio-frontend/src/modules/planner_v2/components/calendar-view/CalendarItemPost.vue` (L78–88) — `inline-flex items-center … px-2 py-1 rounded-full max-w-full shrink`, with the label wrapped in `<span class="truncate">`.
- List cell chip: `contentstudio-frontend/src/modules/planner_v2/components/DataTable.vue` (L606–616) — `inline-flex p-2 … rounded-full`.

**Verify-only (likely no change):**
- `CalendarItemPost.vue` (calendar view, and the shared view via `ExternalCalendarEvent.vue`) — confirm the chip is still content-width with symmetric `px-2` padding next to the content-type chip.
- `CalenderView.vue` (L3138) — the `.content-category` rule only sets `font-size: 13px` for the "+N more" popover; confirm no width override is needed there.

**Truncation:**
- Mirror the calendar chip's pattern on the feed chip — wrap the label in a truncating span inside a `max-w-full` chip — so long category names ellipsis instead of stretching the card.

**Gotcha:**
- `FeedViewCard.vue` is a large (1700+ line) legacy Options API component. Per `contentstudio-frontend/CLAUDE.md`, a small bug fix here does not require a full Composition API conversion — touch only the chip's markup/classes.

---

### Shortcut fields

- **Template:** New Feature Template (PO selects this so the standard sections + 5 quality-checklist tasks are pre-populated)
- **Story type:** chore (frontend styling / bug fix)
- **Project:** Web App
- **Group:** Frontend
- **Epic:** Q2 - 2026: Miscellaneous
- **Priority:** Low (cosmetic polish — adjust if bundled with other planner work)
- **Product area:** Planner
- **Skill set:** Frontend
- **Estimate:** _(leave empty — devs estimate during sprint planning)_
- **Labels:** _(none — team manages labels)_
- **Workflow state:** Ready for Dev
- **Iteration:** _(PO assigns the current/target sprint at creation time)_
