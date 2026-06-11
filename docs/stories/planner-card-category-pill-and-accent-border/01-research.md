# Research: Content Category Pill Width + Left Accent Border on Planner Cards

## Current State

Both bugs are in the same file:
`contentstudio-frontend/src/modules/planner_v2/components/calendar-view/CalendarItemPost.vue`

### Bug 1 — Category Pill Width (line 82)

```html
<div class="flex items-center gap-2 min-w-0 flex-1">   <!-- flex-1 stretches to full width -->
  <div class="flex items-center gap-2 min-w-0 flex-1"> <!-- flex-1 again -->
    <span class="... inline-flex items-center max-w-full shrink">
      <span class="truncate">{{ item?.content_category?.name }}</span>
    </span>
```

The pill span uses `inline-flex items-center` (which should be fit-to-content), but is nested inside two layers of `flex-1` divs. As a flex child, it can expand to fill the parent, causing excess horizontal space on the right. No `w-fit` or explicit constraint limits its width to its text content.

### Bug 2 — Left Accent Border (lines 9 + 14)

```html
<div
  class="post-wrapper rounded-lg border-r border-t border-b border-l-4! border-l-solid bg-[#f2f7fa]"
  :style="`border-left-color: ${borderStatusColor}`"
>
```

The accent is implemented as a CSS `border-l-4!` on the card itself. This means:
- It runs the **full card height** (not just down to the time row)
- It sits flush with the card's border (which clips it inside the card's rounded corners)
- The card already has `rounded-lg` on all corners — the top-left corner is doubly rounded
- **Color is `borderStatusColor`** (computed from post status, e.g. published/failed/scheduled) — NOT from the content category color

`borderStatusColor` is defined at lines 780–784:
```js
const borderStatusColor = computed(() => {
  return props.item?.partially_failed
    ? getStatusColorCode('partially_failed')
    : getStatusColorCode(props.item?.status)
})
```

## What Needs to Change

**Pill fix:**
- Add `w-fit` (Tailwind: `width: fit-content`) to the pill `<span>` so it doesn't expand beyond its text content
- Remove or balance `flex-1` from the parent wrappers if needed to prevent the pill from stretching
- Ensure left and right padding (`px-2`) are equal and match other pills on the card (e.g. hashtag pill)

**Accent border fix:**
- Remove `border-l-4!` from the card's class; instead add a separate sibling element to the LEFT of the card
- Wrap card + accent in a `flex flex-row` container
- Accent element height: covers only the category pill + time row area (not full card height)
- Accent border-radius: `rounded-tl-[8px] rounded-br-[8px]` — top-left and bottom-right rounded; top-right and bottom-left sharp
- Card border-radius: remove top-left rounding (`rounded-tl-none`) where the accent attaches; keep `rounded-tr-lg rounded-br-lg rounded-bl-lg`
- Accent color: use `item.content_category?.color_code` (category color) not post status color
- If no category is assigned to a post, hide the accent element

## Files Involved

- `contentstudio-frontend/src/modules/planner_v2/components/calendar-view/CalendarItemPost.vue` — only file affected
