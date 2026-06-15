# Research — Content Category Pill Width in Planner

**Type:** Frontend-only CSS/markup bug fix
**Area:** Publisher → Planner (planner_v2)
**Date:** 2026-06-11

---

## Current State

The **Content Category pill** is the small colored chip showing a post's content category (its background comes from the category's `color_code`). It renders on planner post "cards" across several planner views. The bug: on at least one view the pill stretches the full available width with empty space on its right, instead of hugging its label text like the other chips on the card (content type, labels, campaign).

How each surface renders the pill today:

| Surface | Component | Pill markup | Sizing today |
|---|---|---|---|
| Calendar view card | `contentstudio-frontend/src/modules/planner_v2/components/calendar-view/CalendarItemPost.vue` (L78–88) | `<span class="… px-2 py-1 rounded-full content-category max-w-full shrink inline-flex items-center">` | `inline-flex` → already content-width |
| Calendar "+N more" popover | same `CalendarItemPost.vue` rendered inside `.fc-popover-body` | `.content-category` rule in `CalenderView.vue` only overrides `font-size: 13px` | inherits `inline-flex` → content-width |
| Shared / external calendar | `ExternalCalendarEvent.vue` → renders `CalendarItemPost.vue` | same as calendar view | content-width |
| **Feed view card** | `contentstudio-frontend/src/modules/planner/components/view/feed-view/FeedViewCard.vue` (L170–182) — legacy component reused by `planner_v2/views/FeedView.vue` | `<div class="top-category" :style="{ background-color }">` | **block-level `<div>` → stretches full width** |
| List view (table cell) | `contentstudio-frontend/src/modules/planner_v2/components/DataTable.vue` (L606–616) | `<span class="inline-flex p-2 … rounded-full">` | `inline-flex` → content-width (already correct) |

**Most likely culprit — the feed view card.** `FeedViewCard.vue` styles the pill with the class `top-category`, whose padding, border-radius, white text, and crucially `display: inline-block` all come from legacy selectors scoped under `.planner_component .planner_feed_view .feed_box …` (both in the component's own `<style scoped>` and in the global `src/assets/styles/legacy/modules/composer/planner/_planner.css` ~L1963–1981). In `planner_v2`'s FeedView the card's root element is just `.feed_box` — the `.planner_component .planner_feed_view` ancestors don't exist — so those rules **don't match**. The `.top-category` element falls back to a plain block `<div>`, which fills the row width (the elongated pill), and only the inline `background-color` survives.

The calendar-view and list-view pills already use `inline-flex` and size to content, so they are visually correct today; they are in scope only to verify symmetric left/right padding and consistency with the other chips.

## What Needs to Change

- **Feed view card (`FeedViewCard.vue`):** make the Content Category pill size to its content again — give `.top-category` an explicit fit-to-content display (`inline-flex` / `inline-block` / `w-fit`) plus the padding and `rounded-full` it should always have had, instead of depending on legacy `.planner_component .planner_feed_view` selectors that never match in planner_v2. Keep the dynamic `background-color` from `content_category.color_code`.
- **Calendar / shared / external view (`CalendarItemPost.vue`):** verify the pill remains content-width and that left/right padding is symmetric (`px-2`) and matches the content-type chip beside it. No change expected beyond confirmation; tighten only if a visual gap is found.
- **Consistency:** the Content Category pill's horizontal padding should match the sibling chips on the same card (content type / labels / campaign) so the row reads as one consistent chip style.
- No API, data-model, or backend change. No new component. No copy change (label text is the category name; tooltip already exists).

## UX Reference

N/A — purely an internal CSS/layout fix; no new UX pattern.

## Files Involved

- `contentstudio-frontend/src/modules/planner/components/view/feed-view/FeedViewCard.vue` — feed view card pill (`.top-category`, L170–182; legacy scoped styles ~L1649+) — **primary fix**
- `contentstudio-frontend/src/modules/planner_v2/components/calendar-view/CalendarItemPost.vue` — calendar/shared/external card pill (L78–88) — verify content-width + symmetric padding
- `contentstudio-frontend/src/modules/planner_v2/views/CalenderView.vue` — `.content-category` popover rule (L3138) — font-size only; confirm no width override needed
- `contentstudio-frontend/src/modules/planner_v2/components/DataTable.vue` — list-view pill (L606–616) — already `inline-flex`; reference for the correct pattern
- (reference only) `contentstudio-frontend/src/assets/styles/legacy/modules/composer/planner/_planner.css` (~L1963–1981) — legacy `.top-category` rules that no longer match in planner_v2

## Notes / Scope Boundaries

- **Web-only.** The planner calendar/feed is a web surface; the iOS/Android apps have their own planner UI and are unaffected by this CSS change — no mobile stories.
- **No analytics event** — pure styling fix, no new trackable user action.
- **Theming:** the pill background is a per-category color from `color_code` (not a brand/theme token), so it is intentionally dynamic and not a white-label `primary-cs` color — leave the inline `background-color` as-is.
