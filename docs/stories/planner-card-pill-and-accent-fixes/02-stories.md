# Stories — Planner Card: Content Category Pill Width + Left Accent Border Fixes

Single combined `[FE]` story. Both fixes are CSS/markup-only changes in one component (`CalendarItemPost.vue`), pure frontend, no backend, no mobile, no API surface.

---

## Story 1 of 1

### Title

`[FE] Fix Content Category pill width and left accent border styling on calendar planner cards`

### Shortcut metadata (for the push step, not part of the story body)

| Field | Value |
|---|---|
| Story type | `feature` |
| Template | New Feature Template |
| Workflow state | Ready for dev |
| Project | Web App |
| Epic | Q2 - 2026: Miscellaneous |
| Group | Frontend |
| Iteration | (confirm with user before pushing) |
| Priority | Medium |
| Product area | Planner |
| Skill set | Frontend |
| Estimate | _(none)_ |
| Labels | _(none)_ |

### Story body (push this exact markdown as `description`)

```markdown
### Description:

As a ContentStudio user viewing my scheduled posts on the planner calendar, I want the Content Category pill on each card to fit snugly to its label text — and the left accent border on each card to look like an intentional, attached accent in my category's color — so the cards look polished, visually balanced, and consistent with the rest of the pills on the card.

Today the Content Category pill stretches wider than its text on calendar cards, while the sibling pill (hashtag / content type) on the same card sizes correctly — so two pills with the same styling look mismatched. And the left accent border runs the full height of the card, sits clipped inside the card with a visible gap from the edge, has no rounded corners, and is colored by post status instead of the card's content category — making it look disconnected and visually broken.

---

### Workflow:

1. User opens the planner and switches to **Calendar** view.
2. User sees their scheduled / published / draft posts laid out as cards on the calendar grid.
3. On each card that has a content category assigned:
   - The Content Category pill at the top of the card is **only as wide as its label text** — no extra space on the right — with equal padding on the left and right, matching the sibling pill (hashtag / content type) shown next to it.
   - A **left accent strip** in the card's content-category color sits flush against the card's left edge with zero gap, spans from the top of the card down to the end of the time row only (not full card height), and has rounded top-left + bottom-right corners that visually merge with the card.
4. On a card that has **no content category** assigned, the accent strip is hidden and the card keeps its existing fully-rounded shape — no flat-corner artifact remains.
5. User changes the color of a content category (Settings → Content Categories) and refreshes the planner; the accent strip on every card using that category updates to the new color.

---

### Acceptance criteria:

**Content Category pill width**

- [ ] On a calendar planner card, the Content Category pill width fits its label text — no excess space on the right.
- [ ] Left and right padding on the pill are visually equal and match the sibling pill (hashtag / content type) shown next to it on the same card.
- [ ] Pill behavior is consistent across short labels (e.g. `News`) and long labels (e.g. `Product Launch Announcements`); long labels truncate with an ellipsis instead of widening the card or wrapping.
- [ ] The same fit-to-content behavior applies to the sibling hashtag / content type pill — both pills use identical sizing rules.
- [ ] Pill rendering matches across the calendar view's normal cards and the FullCalendar "more events" popover.

**Left accent border**

- [ ] The left accent strip on each calendar planner card sits flush against the card's left edge with zero visible gap.
- [ ] The accent strip's height covers only the top of the card down to the bottom of the time row (the row containing the post-status icon and scheduled time) — it does **not** extend through the post body, media, or action buttons below.
- [ ] The accent strip's top-left corner and bottom-right corner are rounded (`8px`); top-right and bottom-left corners are sharp (`0px`).
- [ ] The card's top-left corner is flat (`0px`) where the accent attaches; top-right, bottom-right, and bottom-left corners are rounded (`12px`).
- [ ] The accent color matches the assigned content category's color (the same color used to fill the Content Category pill on the same card).
- [ ] When a card has **no content category** assigned, the accent strip is not rendered and the card keeps its existing fully-rounded shape (all four corners rounded the same as today).
- [ ] Post-status (scheduled / published / failed / partially failed / draft, etc.) is still visually identifiable on the card via the existing status icon + tooltip in the time row — losing the status-colored left border does not remove the user's ability to tell post state at a glance.
- [ ] Behavior matches in both **normal** calendar cards and **compact** calendar cards (when the calendar density preference is set to compact). In compact mode the accent still attaches flush, with the same rounded-corner rules; it spans the visible compact header rather than the full card.
- [ ] No visible regression in the FullCalendar "more events" popover — cards inside the popover render the accent and pill with the same rules.

**Theming / white-label**

- [ ] The accent color is taken directly from the content category's stored color and renders correctly under default ContentStudio theming and all white-label themes — no class hardcoded to a brand-blue palette.
- [ ] The card body, borders, and shadow do not change appearance — only the pill width and the left accent are affected.

---

### Mock-ups:

N/A — visual spec is in the AC. Recommend a quick Loom or before/after screenshot from design once the fix lands, attached to this story.

---

### Impact on existing data:

None. This is a pure presentational change. The `content_category.color_code` field is already loaded with each plan; no schema, API, or persisted-state changes.

---

### Impact on other products:

- **Mobile apps (iOS / Android):** None. The calendar card affected here is web-only — mobile uses its own planner UI.
- **Chrome extension:** None.
- **Backend:** None.
- **Other planner views (List, Feed, Grid, Mobile data card, Post preview drawer):** Out of scope. They render their own category chip / row treatments with different markup and don't currently use the left accent.

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support — N/A, no copy changes. The pill renders user-stored category names, which already pass through verbatim.
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension)

---

### Implementation references

*Pointers from research — not a contract. Engineering may choose a different approach.*

**Primary entry point:**

- `contentstudio-frontend/src/modules/planner_v2/components/calendar-view/CalendarItemPost.vue` — the only file that needs to change. Both fixes live here.
  - Template wrapper (lines 1–15): currently uses `border-l-4! border-l-solid` on the wrapper with `:style="border-left-color: ${borderStatusColor}"`. This whole approach is replaced by a sibling accent `div`.
  - Pill markup (lines 78–99): both the content-category pill and the sibling content-type / tag pill share identical classes — apply the same width fix to both so they stay visually consistent.
  - `borderStatusColor` computed (around line 780): becomes unused after the change and should be removed.

**Pill — suggested approach:**

- Current classes: `leading-none px-2 py-1 text-white bg-[#2f8ae0] rounded-full content-category max-w-full shrink inline-flex items-center`.
- The likely cause of the stretching is the combination of `max-w-full shrink` on a `truncate` child inside a `flex-1` parent — under FullCalendar's event sizing the pill is given more horizontal room and extends past content width.
- Removing `max-w-full shrink` and letting `inline-flex` size to content (or making the pill explicitly `w-fit`) is the minimal fix. Keep `truncate` on the inner span so long category names still ellipsize gracefully when the parent does run out of room.

**Accent border — suggested approach:**

- Restructure: wrap the post card and the accent in a `flex flex-row items-stretch` container so the accent is a true sibling of the card (no `border-left`, no absolute positioning, no gap).
- Render the accent as a sibling `div`:
  - Width: 4px (matching the current visual width).
  - Height: covers the header block (category-pill row + time row). Simplest path is to make the accent and the header share a flex column so they stretch together, with the post body in its own block below.
  - `border-radius: 8px 0 8px 0` — Tailwind: `rounded-tl-lg rounded-br-lg` (TL + BR = 8px) with the other corners default-sharp.
  - Background: bound via `:style="{ backgroundColor: item?.content_category?.color_code }"`. Hide the accent (`v-if="item?.content_category?.color_code"`) when no category.
- Card body radius: change wrapper's `rounded-lg` (8px on all corners) to `rounded-tl-none rounded-tr-xl rounded-br-xl rounded-bl-xl` (0, 12, 12, 12) when an accent is showing; fall back to the original `rounded-lg` (all four corners equal) when there is no category and no accent is rendered.

**Existing behavior to preserve:**

- Compact mode (`compactMode ? 'p-[2px]' : 'p-[8px]'`) — both the pill fix and the accent must work in normal *and* compact density.
- Post-status communication on the card stays via the status icon + tooltip in the time row (already present at lines 109–116). The colored left border is **not** the only status signal today, so removing it does not lose information.
- `.content-category` CSS rule in `CalenderView.vue:3138` (font-size bump inside the FullCalendar popover) — leave it alone, it's unrelated to width.

**Out of scope (confirmed via search):**

- `DataTable.vue` (list view category chip, line 606) — different markup, no accent border.
- `PlannerPostPreview_v2.vue` — detail drawer, separate layout.
- `DataCardMobile.vue` — different structure.
- All non-calendar planner views (Feed, Grid).

**Gotcha:**

- The `id="post-wrapper"` on the outer div is referenced by the `.planner-calender-main .fc-popover-body .post-wrapper#post-wrapper ...` selectors in `CalenderView.vue` (lines 3120+). If the restructure changes which element gets `id="post-wrapper"`, those popover-specific overrides may need a follow-up selector update to keep popover styling intact.

**Tailwind v4 reminder:**

- Project uses Tailwind v4 — important modifier is a trailing `!` (e.g. `p-0!`), not a leading `!`. Prefer named scale tokens (`rounded-xl` = 12px, `rounded-lg` = 8px) over arbitrary brackets.
```

---

## Notes for the dev picking this up

- **Estimate:** intentionally left empty — assign during sprint planning.
- **Labels:** none.
- **Single PR expected.** Branch from `develop`, name `feature/sc-{story-id}/fe-fix-planner-card-pill-and-accent`. Commit message format: `[sc-{story-id}] [FE] Fix Content Category pill width and left accent border on calendar planner cards`.
- **Manual QA matrix:** test both normal and compact calendar density, with and without a content category assigned, with short and long category names, and inside the FullCalendar "more events" popover.
```
