# Stories: Planner Card — Category Pill Width + Left Accent Border Fixes

---

## Story 1 of 1

**Title:** `[FE] Fix content category pill width and left accent border on planner calendar cards`

---

### Description:

As a ContentStudio user viewing my content calendar, I want the content category pill on planner cards to be sized to its text and the left accent border to sit flush against the card edge with the correct height, rounding, and category color — so that the cards look clean, intentional, and visually consistent.

---

### Workflow:

1. User opens the Planner and navigates to Calendar view.
2. User sees planner cards rendered on the calendar grid.
3. Each card with a content category assigned displays a colored pill (e.g., "Blog", "Campaign") at the top of the card.
4. The pill's width matches the length of the category name text — equal padding on both sides, no trailing space on the right.
5. To the immediate left of the card sits a vertical accent bar:
   - It is flush against the card's left edge with no gap.
   - Its color matches the content category assigned to that card (e.g., the same color shown in the pill).
   - Its height spans from the top of the card down to the time row only — it does not run the full card height.
   - The top-left and bottom-right corners of the accent bar are rounded; the top-right and bottom-left corners are sharp, so the bar connects cleanly to the card and tapers away at the bottom.
   - The card's top-left corner is flat where the accent attaches, eliminating any double-curve between the accent and card corner.
6. Cards without a content category assigned show no accent bar and retain standard card styling.

---

### Acceptance criteria:

- [ ] Content category pill width fits the text content — no excess horizontal space on the right side
- [ ] Left and right padding on the content category pill are equal and visually consistent with the Hashtag and Phase pills on the same card
- [ ] The left accent bar sits flush against the card's left edge with zero gap between the accent and the card
- [ ] The accent bar height covers only the category pill row and the time row — it does not extend to the full card height
- [ ] The accent bar's top-left and bottom-right corners are rounded; top-right and bottom-left corners are sharp
- [ ] The card's top-left border radius is removed at the point where the accent bar attaches, so there is no double-rounding between the accent and the card
- [ ] The accent bar color is driven by the content category's assigned color — it matches the color used in the category pill on the same card
- [ ] Cards with no content category assigned display no accent bar and render with their normal rounded card styling
- [ ] The same pill and accent fixes apply in the expanded popover view (when clicking "more" to see all cards in a day), not only in the main calendar grid cells

---

### Mock-ups:

N/A — fixes are described by the acceptance criteria above. Reference the current pill (e.g., "Blog" pill with trailing whitespace) and accent border (full-height, gapped, unrounded) for before/after comparison during QA.

---

### Impact on existing data:

None. No schema changes. The accent color is read from the existing `content_category.color_code` field already stored on the post object.

---

### Impact on other products:

- **Mobile apps:** Not affected — the Calendar view is web-only.
- **Chrome extension:** Not affected.
- **White-label:** The accent color comes from `content_category.color_code` (user-defined), not from CSS variables, so it adapts automatically. No white-label impact.

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only) — N/A; Calendar view is desktop-only
- [ ] Multilingual support (frontend + backend, translations available or fallback handled) — N/A; no new user-facing strings introduced
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry point:**
- `contentstudio-frontend/src/modules/planner_v2/components/calendar-view/CalendarItemPost.vue` — all changes are in this single file

**Fix 1 — Category pill width:**
- The pill `<span>` at line 82 uses `inline-flex items-center max-w-full shrink` but no `w-fit`. Its parent `<div>` at line 76–77 has `flex-1`, which causes the flex container to grow and the pill to stretch. Adding `w-fit` (Tailwind v4) to the pill span, or removing `flex-1` from the inner wrapper, constrains it to its text content.
- The same tag span (line 93) used for Evergreen/RSS/Repeat has the same class set and needs the same fix.

**Fix 2 — Left accent bar:**
- Current approach: `border-l-4! border-l-solid` on `#post-wrapper` (line 9), colored via `borderStatusColor` (computed from post status, not content category).
- Suggested approach: wrap the entire card in a `flex flex-row` container; insert a sibling `<div>` to the left of the card as the accent bar. Set its height to cover only the category pill + time row area. Apply `border-radius: 8px 0 8px 0` inline (top-left + bottom-right rounded, top-right + bottom-left sharp). Set its background color from `item?.content_category?.color_code`. Show it only when `item?.content_category` is present.
- Remove `border-l-4! border-l-solid` from `#post-wrapper`; change card's `rounded-lg` to `rounded-r-lg` (or adjust to `0 12px 12px 12px`) so the top-left corner is flat where the accent attaches.
- Popover override in `CalenderView.vue` at line 3116–3121 sets `border-left: 1px solid rgb(236, 238, 245) !important` on `#post-wrapper` inside `.fc-popover-body` — this will need updating when the border-left approach is removed from the card.
- `borderStatusColor` computed (line 780) can remain for any future use; it just no longer drives the left visual indicator.
