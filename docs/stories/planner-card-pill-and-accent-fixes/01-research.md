# Research — Planner Card: Content Category Pill Width + Left Accent Border Fixes

## Scope

Two visual bugs on **calendar-view planner cards**:

1. The Content Category pill renders wider than its label text (excess space on the right, asymmetric padding).
2. The left accent border runs full card height, sits clipped inside the card with a gap, has no rounding, and is colored by post status — not by the card's content category.

Both fixes live in the same single-file component.

---

## Current State

### Component

- **`contentstudio-frontend/src/modules/planner_v2/components/calendar-view/CalendarItemPost.vue`** — single calendar-view card rendered inside FullCalendar (also reused inside `CalenderView.vue` popover).
- Options API; touches just the template (lines 1–15 for the wrapper, 73–102 for the pill area) and one computed (`borderStatusColor`, line 780).
- Other planner views (`DataTable.vue` list view, `FeedView`, `GridView`, `PlannerPostPreview_v2`) render their own category chips and do **not** use the same accent-border treatment, so they are out of scope. Mobile data card (`DataCardMobile.vue`) likewise unaffected.

### Current pill markup (lines 78–99)

```vue
<template v-if="item?.content_category">
  <span
    v-tooltip="t('planner.calendar_view.item_post.tooltips.content_category')"
    :style="{ 'background-color': item?.content_category?.color_code }"
    class="leading-none px-2 py-1 text-white bg-[#2f8ae0] rounded-full content-category max-w-full shrink inline-flex items-center"
  >
    <span class="truncate">{{ item?.content_category?.name }}</span>
  </span>
</template>

<template v-if="getItemTag">
  <span
    v-tooltip="t('planner.calendar_view.item_post.tooltips.content_type')"
    class="leading-none px-2 py-1 text-white bg-[#2f8ae0] rounded-full content-category max-w-full shrink inline-flex items-center"
  >
    <span class="truncate">{{ getItemTag }}</span>
  </span>
</template>
```

- The outer parent containers each carry `flex items-center gap-2 min-w-0 flex-1` (lines 76–77), pushing both pills to share the row.
- Only CSS rule for `.content-category` lives in `CalenderView.vue:3138` and just bumps font-size inside the FullCalendar popover — no width or padding overrides.
- Visible symptom: the **content category** pill stretches beyond the label while the **content type / hashtag** pill next to it (same classes) sits content-sized. The likely cause is the combination of `truncate` (`white-space:nowrap; overflow:hidden`) on the inner span and the `max-w-full shrink inline-flex` on the outer span, which under FullCalendar's event sizing extends the pill past content width when there is excess horizontal slack in the parent's `flex-1` child.

### Current accent border (lines 9–14)

```vue
<div
  id="post-wrapper"
  class="post-wrapper rounded-lg border-r border-t border-b border-l-4! border-l-solid bg-[#f2f7fa] transition-opacity"
  :class="[compactMode ? 'p-[2px]' : 'p-[8px]', shouldFadePublished ? 'opacity-70' : '']"
  :style="`border-left-color: ${borderStatusColor}`"
>
```

```js
// line 780
const borderStatusColor = computed(() => {
  return props.item?.partially_failed
    ? getStatusColorCode('partially_failed')
    : getStatusColorCode(props.item?.status)
})
```

Observations vs. expected behavior:

| Property | Current | Expected |
|---|---|---|
| Color source | Post status (`getStatusColorCode`) | `item.content_category.color_code` |
| Height | Full card height (CSS `border-left`) | Top of card to bottom of time row only |
| Position | Inside card (border, not a sibling) — causes the perceived "gap" and clipping | Direct sibling, flush against the left edge |
| Corners | All 4 corners follow card's `rounded-lg` | Accent: `8px 0 8px 0` (TL + BR rounded; TR + BL sharp); Card: `0 12px 12px 12px` (TL flat) |
| Width | 4px via `border-l-4!` | Same visual width, but rendered as its own block |

### Time row anchor (lines 104–121)

The "time row" referenced in the spec is the `flex items-center justify-between h-[19px] text-[10px]` block at line 105 — contains the status icon and `formatTime(item?.time)`. The accent must extend from the top of the wrapper down to the **bottom of this row**. Below it sits the post body, media preview, and action buttons, which the accent must not cover.

### Status-color usage (don't lose it)

`borderStatusColor` is currently the only carrier of post-status visual signal on the card body. Reassigning the accent to category color removes that signal. The post-status icon at line 109–116 (with tooltip) already conveys status, so removing the status-tinted border is acceptable — but worth flagging for design sign-off in the story.

---

## What Needs to Change (in `CalendarItemPost.vue`)

**Pill (apply to both the content-category pill and the matching content-type/tag pill so they stay consistent):**

- Drop `max-w-full shrink` and ensure the span is content-sized — `inline-flex` alone with `w-fit` (or no width hint and no `flex-1` ancestor that forces stretch) is the simplest fix.
- Keep `px-2 py-1` (already symmetric) — confirm padding matches the sibling hashtag/phase pills exactly.
- Keep `truncate` on the inner text span only as a safety net for over-long names, not as the cause of the stretch.

**Accent border:**

- Stop using `border-left` on the wrapper. Wrap the card and the accent in a new `flex flex-row` container so the accent is a sibling, not a border.
- Render the accent as a `div` with:
  - `width: 4px` (or current visual width)
  - `height` covering the category-pill row + time row (use a wrapping div for those two rows whose height the accent matches, or position the accent absolutely against a header sub-section)
  - `border-radius: 8px 0 8px 0` (TL + BR rounded; TR + BL sharp)
  - `background-color` bound to `item?.content_category?.color_code` with a sensible fallback when the post has no category
- Change card's outer `rounded-lg` (= 8px) to `rounded-tl-none rounded-tr-xl rounded-br-xl rounded-bl-xl` — equivalent to the spec's `0 12px 12px 12px`. (Current card uses `rounded-lg` = 8px; spec asks for 12px on the three remaining corners — design clarification flagged in the story.)
- Remove the `border-l-4! border-l-solid` and dynamic `border-left-color` style from the wrapper.
- `borderStatusColor` computed becomes unused — remove it (status is still conveyed by the icon + tooltip at line 109).

**Fallback when a card has no content category:**

The accent should be hidden (or use a neutral color) when `item?.content_category?.color_code` is missing — otherwise the corner-shape change applied to the card would leave a sharp top-left with no accent attached. Flag this for the spec.

---

## Mobile Context

N/A — this is a calendar-view component in the web app only. Mobile apps (`contentstudio-ios-v2`, `contentstudio-android-v2`) are not impacted.

---

## Files Involved

| Path | Why |
|---|---|
| `contentstudio-frontend/src/modules/planner_v2/components/calendar-view/CalendarItemPost.vue` | Template (lines 1–15, 73–102) and `borderStatusColor` computed (line 780). Only file that needs to change. |

## Files Confirmed Out of Scope

| Path | Reason |
|---|---|
| `contentstudio-frontend/src/modules/planner_v2/components/DataTable.vue` (line 606) | List view category chip — different markup, no accent border, not described in the spec |
| `contentstudio-frontend/src/modules/planner_v2/components/PlannerPostPreview_v2.vue` | Detail drawer — different layout |
| `contentstudio-frontend/src/modules/planner_v2/components/DataCardMobile.vue` | Mobile data card — different structure |
| `contentstudio-frontend/src/modules/planner_v2/views/CalenderView.vue:3138` | Only sets `.content-category` font-size inside FullCalendar popover; no width/padding to change |

---

## Open Questions for Confirmation

1. The spec asks for **`border-radius: 0 12px 12px 12px`** on the card but the card currently uses `rounded-lg` (= 8px) on all corners. Should the three non-attached corners go from 8px → 12px, or stay 8px and only the top-left flatten? (Defaulting to "follow the spec exactly: bump to 12px on the three remaining corners".)
2. When a card has **no content category**, should the accent + flat-corner treatment disappear entirely (fall back to the existing all-rounded card), or render with a neutral grey accent? (Defaulting to "hide accent + restore the existing all-rounded card".)
3. The current border color signals **post status**. Reassigning to category color removes that signal — the status icon + tooltip remain. Is design OK losing the colored border-as-status cue? (Flagging in story, no AC change needed if confirmed.)
