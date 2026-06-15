# Research — Move Planner Post Card Accent Border from Left Edge to Top Edge

**Type:** Frontend-only CSS fix
**Area:** Publisher → Planner (planner_v2), Calendar view
**Date:** 2026-06-11

---

## Current State

The planner **calendar** post card renders a colored accent as a **4px border on its left edge**, with thin (1px) borders on the other three sides and `rounded-lg` corners on a light (`#f2f7fa`) background.

- Component: `contentstudio-frontend/src/modules/planner_v2/components/calendar-view/CalendarItemPost.vue`
  - Border classes (L9): `rounded-lg border-r border-t border-b border-l-4! border-l-solid …`
  - Accent color (L14): `:style="border-left-color: ${borderStatusColor}"`
  - `borderStatusColor` (L780) → `getStatusColorCode(item.status)` (uses `partially_failed` when applicable)
- The same card is reused in:
  - Calendar grid — via `CalendarEvent.vue`
  - The "+N more" day popover — same card inside `.fc-popover-body`
  - Shared / external calendar — via `ExternalCalendarEvent.vue`
- The **feed** view card (`FeedViewCard.vue`) and **list** view (`DataTable.vue`) do **not** use this left accent border — they convey status differently (a status dot/label and a status column). So this change is scoped to the calendar post card only.

### ⚠️ Important: the accent reflects post STATUS, not content category

In the current code the accent color comes from `getStatusColorCode(status)` — the post's **publishing status**, not its content category:

| Status | Accent color |
|---|---|
| Published | `#5EBC7E` (green) |
| Scheduled | `#F0BB52` (amber) |
| Draft | `#76797C` (gray) |
| Under review / review | `#5FB6F9` (blue) |
| Failed | `#EB554D` (red) |
| Partially failed | `#B52D4A` (dark red) |
| Rejected | `#EB516B` (pink-red) |
| Processing | `#9299F8` (indigo) |
| … | … |

The **content category** color (`content_category.color_code`) is used only for the category **pill's** background (L81) — not for the accent border. So the request's phrase "colored category accent border" does not match the code, where it's a **status** accent. This needs confirmation before writing the story (see open question).

## What Needs to Change

- Move the 4px colored accent from the card's **left** edge to its **top** edge: `border-l-4` → `border-t-4` (keep the other three sides as the thin 1px border).
- Keep the accent dynamic/inline-colored, just bound to the top edge (`border-top-color` instead of `border-left-color`).
- Keep `rounded-lg` so the top accent runs between the card's rounded top corners and looks intentional.
- Applies automatically to both the compact and normal card layouts (the border is on the shared `post-wrapper` root) and to every surface that renders this card (calendar grid, "+N more" popover, shared/external calendar).
- No API, data-model, or backend change. No new component, no copy change.

## Decision (resolved with user)

**Move only — keep the status color.** Move the existing **status**-colored accent from the left edge to the top edge with no color-source change. The accent continues to reflect the post's publishing status (the "category" wording in the request was a misnomer); the at-a-glance status signal is preserved, just relocated to the top.

(The alternative — recoloring the accent to the content category color — was declined; it would need a fallback for posts with no category and would drop the status signal.)

## Files Involved

- `contentstudio-frontend/src/modules/planner_v2/components/calendar-view/CalendarItemPost.vue` — accent border (L9, L14; `borderStatusColor` L780) — **the only change**
- (reference) `contentstudio-frontend/src/modules/planner_v2/utils/index.js` — `getStatusColorCode()` color source
- (verify, no change expected) `CalendarEvent.vue` / `ExternalCalendarEvent.vue` — render the card; no separate border logic

## Notes / Scope Boundaries

- **Web-only.** iOS/Android apps have their own planner UI — unaffected; no mobile stories.
- **No analytics event** — pure styling change.
- **Theming:** the accent color is a status/category data color, not a white-label brand token — leave it as a dynamic inline color.
