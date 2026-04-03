# Research: Calendar Post Status Border & Tooltip

## Current State

The calendar planner has 3 view modes controlled by two toggles: **media thumbnails** (show/hide) and **compact mode** (on/off):
1. **Normal mode** — full card with media thumbnail, post text, platforms, status (default)
2. **Media hidden mode** — same as normal but without the media thumbnail
3. **Compact mode** — minimal single-line row

**Status border:** Only compact mode has a left colored border (`border-l-[4px]`) using `getStatusColorCode()` from `src/modules/planner_v2/utils/index.js`. Normal and media-hidden modes have no status border.

**Tooltip:** All modes show a generic "Post Details" tooltip on hover (`planner.calendar_view.item_post.tooltips.post_details`). This tooltip doesn't indicate the post status.

**Status colors** are already mapped in `getStatusColorCode()`: published (#5EBC7E), draft (#76797C), scheduled (#F0BB52), processing (#9299F8), partially_failed (#B52D4A), failed (#EB554D), rejected (#EB516B), etc.

## What Needs to Change

1. **Left status border in all views** — Apply the same `border-l-[4px]` with `getStatusColorCode()` to normal mode and media-hidden mode (currently only compact mode has it)
2. **Status-aware tooltip** — Replace the generic "Post Details" tooltip with `"{Status} - Click to view post details"` (e.g., "Published - Click to view post details") across all three view modes

## Files Involved

- `src/modules/planner_v2/components/calendar-view/CalendarItemPost.vue` — the main calendar item component (template + `compactViewStatus` computed)
- `src/modules/planner_v2/utils/index.js` — `getStatusColorCode()` already exists, may need a `getStatusLabel()` for display text
- `src/locales/en/planner.json` — update `planner.calendar_view.item_post.tooltips.post_details` or add new key
- All other locale directories — mirror the new/updated tooltip key
