# Stories: Planner Card — Category Pill Width & Left Accent Border Fixes

---

## Story 1

### Title
`[FE] Fix content category pill width and left accent border on planner cards`

---

### Description

Planner cards in the calendar view have two visual bugs that make the card look unpolished and inconsistent.

1. **Category pill stretches too wide.** The Content Category pill (and the RSS/Evergreen/Repeat tag pill) takes up more horizontal space than its text needs, leaving a large empty gap on the right side of the pill. Other pills on the card render fit-to-content — this pill should too.

2. **Left accent border looks disconnected.** The accent bar on the left edge of each planner card runs the full height of the card, is clipped inside the card's rounded corners, and has no independent rounding of its own — making it look like an afterthought rather than an intentional design element. It also always shows the post status color; it should instead reflect the card's content category color so the accent is semantically meaningful.

Fixing both issues gives users a cleaner, more polished planner experience that visually communicates the card's content category at a glance.

---

### Workflow

This is a visual bug fix with no user-initiated action required. Users simply see the improved card rendering whenever they open the Planner in Calendar view.

1. User navigates to Publish → Planner and selects Calendar view.
2. Posts that have a Content Category assigned show the category pill at the top of the card — the pill is now only as wide as its label text with equal left and right padding.
3. The left accent bar now runs only from the top of the card down to the time row, sits flush against the card's left edge with no gap, and has rounded corners that flow naturally (top-left and bottom-right rounded; top-right and bottom-left sharp).
4. The accent bar color matches the color assigned to that post's content category — so a "Product Updates" category (e.g. green) shows a green accent, and a "Blog Posts" category (e.g. purple) shows a purple accent.
5. If a post has no content category assigned, no accent bar is shown.

---

### Acceptance criteria

**Category pill:**
- [ ] The Content Category pill on a planner card is only as wide as its label text plus its padding — it does not stretch to fill the available row width
- [ ] Left and right padding inside the pill are equal (matching the other pills on the card)
- [ ] The RSS, Evergreen, and Repeat tag pills (which share the same styling) are also fit-to-content after this fix
- [ ] A long category name truncates with an ellipsis before the pill grows beyond the card boundary (existing truncate behavior is preserved)

**Left accent border:**
- [ ] The left accent bar sits flush against the left edge of the card with zero gap between the accent and the card
- [ ] The accent bar spans only from the top of the card down to the time row — it does not run the full card height
- [ ] The accent bar has `border-radius: 8px 0 8px 0` — top-left and bottom-right corners are rounded, top-right and bottom-left corners are sharp
- [ ] The card's own top-left border radius is removed where the accent attaches, so there is no double curve at that corner
- [ ] The card's top-right, bottom-right, and bottom-left corners retain their existing rounding
- [ ] The accent bar color matches `item.content_category.color_code` for the post's assigned category
- [ ] When a post has no content category, no accent bar is rendered
- [ ] The accent bar color updates correctly across all content categories (not hardcoded to any single color)

---

### Mock-ups

N/A — no design files provided. Implement per the spec in Acceptance Criteria.

---

### Impact on existing data

None. This is a purely visual change to how existing `content_category.color_code` data is displayed. No schema or API changes.

---

### Impact on other products

The `CalendarItemPost` component is used in the Calendar view of the Planner (web only). This change does not affect:
- The Planner list view, feed view, or grid view
- Mobile apps
- Chrome extension
- The Composer's content category selector

---

### Dependencies

None.

---

### Global quality & compliance checklist

- [ ] Mobile responsiveness tested (frontend only, N/A for backend-only stories)
- [ ] Multilingual support verified (frontend + backend, translations available or fallback handled) — N/A, no new i18n strings introduced
- [ ] UI theming supported (default + white-label, design library components are being used)
- [ ] White-label domains impact reviewed
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)

---

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary file:**
- `contentstudio-frontend/src/modules/planner_v2/components/calendar-view/CalendarItemPost.vue`

**Pill width fix:**
- Pill spans are at lines 78–99. Both the `content_category` pill (line 82) and the `getItemTag` pill (line 93) share the class `inline-flex items-center max-w-full shrink content-category`. Adding `w-fit` (Tailwind v4: `width: fit-content`) to each pill span should resolve the stretching.
- The two wrapping `flex-1` parent divs (lines 76–77) cause the flex context that allows the pill to grow. Alternatively, `self-start` or `shrink-0` on the pill — or removing the inner `flex-1` from the direct pill parent — would also constrain it.

**Accent border fix:**
- The card's root `<div id="post-wrapper">` currently has `border-l-4! border-l-solid` + `:style="\`border-left-color: ${borderStatusColor}\`"` (lines 9 and 14). The accent should be moved out to a sibling element.
- Suggested wrapper approach: wrap the card `<div>` in a `<div class="flex flex-row">` and prepend a sibling `<div>` as the accent element before the card. The accent element's height should be constrained to match the category pill + time row section (approximately the `post-header` div height).
- `borderStatusColor` is computed at line 780 from `getStatusColorCode(props.item?.status)`. After the refactor, the right border (`border-r border-t border-b`) of the card can continue using this for status indication if desired, or be removed if the design replaces the concept entirely.
- `item?.content_category?.color_code` (line 81 shows this used for the pill background) is the correct source for the accent color.
- The CalenderView.vue scoped CSS at line 3138 has a `.content-category` font-size override for the popover view — check this still works after the template restructure.
