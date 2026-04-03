# Research: Composer Right Sidebar Usermaven Analytics Tracking

## Current State

The composer right sidebar (`ActionsAside.vue`) has **9 tab icons** that users can click:

| Tab | Status key | Condition |
|---|---|---|
| AI Toolkit (hover menu) | — | Always visible (hover flyout, not a tab click) |
| Preview | `preview` | Always visible |
| Tasks | `task` | Hidden on published posts |
| Comments | `comment` | Hidden on published posts |
| Assistant | `assistant` | Hidden on published posts |
| Labels | `labels` | Hidden on published posts |
| Campaigns | `campaigns` | Hidden on published posts |
| Activities | `activity` | Hidden on published posts |
| Members | `members` | Hidden on published posts |

**Existing tracking (line 673):** There's already a `trackUserMaven()` call inside `socialShareTab()`:
```js
this.trackUserMaven('composer_right_sidebar_opened_' + status)
```
This fires events like `composer_right_sidebar_opened_preview`, `composer_right_sidebar_opened_task`, etc.

**Problem:** The event naming is inconsistent with best practices — concatenating the tab name into the event name creates many separate events instead of a single event with properties. This makes it harder to compare tabs in analytics dashboards. Also, the AI Toolkit sub-items (AI Studio, Caption Generator, Image Generator, Hashtag Generator) are not tracked at all.

**Usermaven composable:** `src/composables/useUserMaven.js` — exposes `trackUserMaven(tagName, payload)` which calls `userMaven.track(tagName, payload)`. Already imported in `ActionsAside.vue` (line 496, 535).

## What Needs to Change

1. **Replace the current concatenated event** (`composer_right_sidebar_opened_` + status) with a single event name `composer_sidebar_tab_clicked` and pass `tab_name` as a property in the payload
2. **Add tracking for AI Toolkit sub-items** in `handleAIClick()` — fire `composer_sidebar_ai_tool_clicked` with `tool_name` property
3. **Add tracking for sidebar toggle** (expand/collapse) — fire `composer_sidebar_toggled` with `action: 'expanded' | 'collapsed'`

### Proposed Event Schema

| Event Name | Properties | Fires When |
|---|---|---|
| `composer_sidebar_tab_clicked` | `{ tab_name: 'preview' \| 'task' \| 'comment' \| 'assistant' \| 'labels' \| 'campaigns' \| 'activity' \| 'members' }` | User clicks any sidebar tab icon |
| `composer_sidebar_ai_tool_clicked` | `{ tool_name: 'ai_studio' \| 'ai_caption_generator' \| 'ai_image_generator' \| 'ai_hashtag_generator' }` | User clicks an AI Toolkit sub-item |
| `composer_sidebar_toggled` | `{ action: 'expanded' \| 'collapsed' }` | User expands or collapses the sidebar |

## Files Involved

| File | Change |
|---|---|
| `src/modules/composer_v2/components/ActionsAside.vue` | Update `socialShareTab()` event, add tracking to `handleAIClick()` and `toggleModalSidebar()` |
