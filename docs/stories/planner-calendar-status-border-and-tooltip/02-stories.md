# Stories: Calendar Post Status Border & Tooltip

---

## [FE] Show post status border and status-aware tooltip in all calendar view modes

---

### Description:

As a ContentStudio user, I want to see the post status color on every post card in the calendar — not just in compact mode — so that I can quickly scan the calendar and understand the state of each post at a glance, regardless of which view mode I'm using.

Currently, the calendar's compact mode shows a colored left border that matches the post status (e.g., green for published, yellow for scheduled, red for failed). The normal mode and media-hidden mode do not show this border. Additionally, hovering over any post in any mode shows a generic "Post Details" tooltip — it should instead show the post's actual status so users know what state the post is in before clicking.

**Two changes:**

1. **Status border in all views** — Apply the same left colored border (`border-l-[4px]` with status color from `getStatusColorCode()`) to normal mode and media-hidden mode. The border is already implemented for compact mode in `CalendarItemPost.vue` — extend the same logic to the other two modes.

2. **Status-aware tooltip** — Replace the generic "Post Details" tooltip with a dynamic tooltip that shows the post status. Format: `"{Status} — Click to view post details"` (e.g., "Published — Click to view post details", "Scheduled — Click to view post details").

**Key files:**
- `src/modules/planner_v2/components/calendar-view/CalendarItemPost.vue` — extend left border styling to all modes, update tooltip to use status text
- `src/modules/planner_v2/utils/index.js` — `getStatusColorCode()` already exists and is reused; may add a `getStatusDisplayLabel()` helper
- `src/locales/en/planner.json` and all other locale directories — add/update tooltip key for the status-aware format

---

### Workflow:

1. User opens the Planner and selects Calendar view.
2. In **normal mode** (default), each post card on the calendar shows a colored left border matching its status — green for published, yellow for scheduled, gray for draft, red for failed, etc.
3. User switches to **media-hidden mode** (media thumbnails toggled off) — post cards still show the same colored left border.
4. User switches to **compact mode** — post rows continue showing the colored left border (existing behavior, unchanged).
5. In any of the three modes, user hovers over a post — the tooltip reads the post's current status followed by a prompt to click, e.g., "Published — Click to view post details" or "Scheduled — Click to view post details".

---

### Acceptance criteria:

- [ ] In normal mode, every post card shows a left colored border matching the post status
- [ ] In media-hidden mode, every post card shows a left colored border matching the post status
- [ ] In compact mode, the existing left colored border continues to work as before
- [ ] The border color matches the status: published (green), scheduled (yellow), draft (gray), failed (red), partially failed (dark red), rejected (pink), processing (purple), under review (orange), missed review (red)
- [ ] Hovering over a post in any view mode shows a tooltip in the format: "{Status} — Click to view post details"
- [ ] The tooltip dynamically reflects the post's current status (e.g., "Published", "Scheduled", "Draft", "Failed", etc.)
- [ ] The old generic "Post Details" tooltip no longer appears
- [ ] Tooltip text is translatable (uses `$t()` keys, not hardcoded strings)

---

### UI Copy:

**Tooltip format (per status):**
- `"Published — Click to view post details"`
- `"Scheduled — Click to view post details"`
- `"Draft — Click to view post details"`
- `"Failed — Click to view post details"`
- `"Partially Failed — Click to view post details"`
- `"Rejected — Click to view post details"`
- `"Processing — Click to view post details"`
- `"Under Review — Click to view post details"`
- `"Missed Review — Click to view post details"`
- `"Notification Sent — Click to view post details"`
- `"Notification Declined — Click to view post details"`

**i18n key:** `planner.calendar_view.item_post.tooltips.status_click_details` with interpolation: `"{status} — Click to view post details"` (where `{status}` is the translated status label).

---

### Mock-ups:

N/A — the left border already exists in compact mode and is being extended to other modes. The tooltip is text-only.

---

### Impact on existing data:

None. This is a purely visual change — no data model or API changes.

---

### Impact on other products:

None. The calendar planner is web-only. No impact on mobile apps, Chrome extension, or white-label (border uses existing status colors which are already theme-compatible).

---

### Dependencies:

None.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness — N/A, calendar planner is not used on mobile viewports
- [ ] Multilingual support — tooltip uses `$t()` interpolation; status labels and click prompt translated in all locales
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension)
