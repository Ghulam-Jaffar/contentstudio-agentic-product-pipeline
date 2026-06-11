# Research: Planner Card — Category Pill Width + Left Accent Border Fixes

## Current State

Both issues live in a single component:
**`contentstudio-frontend/src/modules/planner_v2/components/calendar-view/CalendarItemPost.vue`**

This is the card rendered inside the Calendar view of Planner v2. It displays a content category pill, a time row, post body text, labels, and action buttons.

### Issue 1 — Content Category Pill Width

The pill is a `<span>` with class:
```
leading-none px-2 py-1 text-white bg-[#2f8ae0] rounded-full content-category max-w-full shrink inline-flex items-center
```
Its parent `<div>` (line 77) has `flex items-center gap-2 min-w-0 flex-1`, where `flex-1` causes the container to grow to fill the available row width. The pill itself uses `inline-flex` which should fit its content, but lacks an explicit `w-fit` / `width: fit-content` constraint. The `shrink` class alone does not prevent it from stretching when the parent flex container distributes space. The result is the pill appears wider than its text on the right side.

The same pill pattern is used for the tag (Evergreen / RSS / Repeat) on line 93.

### Issue 2 — Left Accent Border

Currently the card's left border is implemented as a native CSS `border-l-4!` on the `#post-wrapper` div (line 9):
```
class="post-wrapper rounded-lg border-r border-t border-b border-l-4! border-l-solid bg-[#f2f7fa] transition-opacity"
:style="`border-left-color: ${borderStatusColor}`"
```

`borderStatusColor` is derived from `post_state` / `partially_failed` status (not from content category color). This produces a full-height border clipped inside the card, with a gap at the corner (the card's own `border-radius` curves away from the border), and no independent rounding on the accent itself.

The user wants the accent to:
- Be a sibling element outside the card, positioned via a flex row wrapper (no gap)
- Span only from the top of the card down to the time row
- Have `border-radius: 8px 0 8px 0` (top-left and bottom-right rounded)
- Pull its color from `item?.content_category?.color_code`
- Cause the card's own top-left border-radius to become 0 where it attaches

## What Needs to Change

- **`CalendarItemPost.vue`** — all changes are in this file
  1. **Pill width:** Add `w-fit` (or `self-start` / remove `flex-1` from inner container) so the category pill shrinks to fit its text content
  2. **Accent border:** Replace `border-l-4!` approach with a flex row wrapper + a sibling `<div>` accent element; adjust card's `rounded-lg` to `rounded-r-lg` (or remove top-left radius); set accent height to cover category + time row only; apply inline style for `border-radius: 8px 0 8px 0` and color from `content_category.color_code`

## UX Reference

Standard card accent bar patterns (e.g., Notion calendar, Google Calendar) use a sibling element sitting flush to the left with asymmetric rounding — top-left mirrors the card's corner, bottom fades into the card body.

## Files Involved

| File | Change |
|---|---|
| `contentstudio-frontend/src/modules/planner_v2/components/calendar-view/CalendarItemPost.vue` | Pill width fix + accent border restructure |
