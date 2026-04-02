# Stories: Composer Right Sidebar Usermaven Analytics Tracking

---

## Story 1: [FE] Add Usermaven analytics tracking for composer right sidebar tab clicks

### Description:

As a product manager, I want to track which composer sidebar tabs users click so that we can measure usage of each tab and decide what to keep, improve, or remove.

The composer right sidebar (`ActionsAside.vue`) has 7 clickable items: AI Toolkit (with 4 sub-items), Preview, Tasks, Comments, Assistant, Activities, and Members. There's already a Usermaven tracking call at line 673 (`trackUserMaven('composer_right_sidebar_opened_' + status)`) but it concatenates the tab name into the event name, making cross-tab comparison difficult in dashboards. The AI Toolkit sub-items (AI Studio, Caption Generator, Image Generator, Hashtag Generator) are not tracked at all.

**Key files:**
- `contentstudio-frontend/src/modules/composer_v2/components/ActionsAside.vue` — sidebar component, `socialShareTab()` (line 672), `handleAIClick()` (line 633)
- `contentstudio-frontend/src/composables/useUserMaven.js` — Usermaven composable (already imported in ActionsAside.vue at line 496)

---

### Workflow:

1. User opens the Composer (social post modal)
2. User clicks any tab icon in the right sidebar (e.g., Preview, Tasks, Comments, Assistant, Activities, Members)
3. A `composer_sidebar_tab_clicked` event is sent to Usermaven with `{ tab_name: '<tab>' }`
4. User hovers over the AI Toolkit icon and clicks a sub-item (e.g., AI Studio, AI Caption Generator, AI Image Generator, AI Hashtag Generator)
5. A `composer_sidebar_tab_clicked` event is sent to Usermaven with `{ tab_name: '<ai_tool>' }`
6. Product team can now view a single event in the Usermaven dashboard, broken down by `tab_name`, to compare usage across all sidebar options

---

### Acceptance criteria:

- [ ] Clicking any sidebar tab fires a `composer_sidebar_tab_clicked` Usermaven event with a `tab_name` property
- [ ] The `tab_name` values are: `preview`, `task`, `comment`, `assistant`, `activity`, `members`, `ai_studio`, `ai_caption_generator`, `ai_image_generator`, `ai_hashtag_generator`
- [ ] The old concatenated events (`composer_right_sidebar_opened_*`) are removed and replaced by the new single event
- [ ] AI Toolkit sub-item clicks (AI Studio, Caption Generator, Image Generator, Hashtag Generator) are tracked — these were previously untracked
- [ ] Events fire on every click, including re-clicking the already-active tab
- [ ] No duplicate events fire for a single click
- [ ] No user-facing UI changes — this is purely analytics instrumentation

---

### Mock-ups:

N/A — no UI changes.

---

### Impact on existing data:

- **Usermaven event change:** The old events `composer_right_sidebar_opened_preview`, `composer_right_sidebar_opened_task`, etc. will stop firing. They are replaced by a single `composer_sidebar_tab_clicked` event with a `tab_name` property. The Usermaven dashboard may need to be updated to use the new event name.

---

### Impact on other products:

- **Mobile apps:** No impact — composer sidebar is web-only
- **Chrome extension:** No impact
- **White-label:** No impact — analytics tracking is invisible to users

---

### Dependencies:

None.

---

### Implementation guidance:

**1. Update `socialShareTab()` (line 672–673):**
```js
// Replace:
this.trackUserMaven('composer_right_sidebar_opened_' + status)
// With:
this.trackUserMaven('composer_sidebar_tab_clicked', { tab_name: status })
```

**2. Add tracking to `handleAIClick()` (line 633):**
Add at the top of the method, before the switch:
```js
const toolNameMap = {
  'writing-assistant': 'ai_studio',
  'caption': 'ai_caption_generator',
  'image': 'ai_image_generator',
  'hashtags': 'ai_hashtag_generator',
}
this.trackUserMaven('composer_sidebar_tab_clicked', { tab_name: toolNameMap[requestType] || requestType })
```

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only) — N/A, no UI changes
- [ ] Multilingual support (frontend + backend, translations available or fallback handled) — N/A, no user-facing strings
- [ ] UI theming support (default + white-label, design library components are being used) — N/A, no UI changes
- [ ] White-label domains impact review — N/A
- [ ] Cross-product impact assessed (web, mobile apps, Chrome extension) — N/A, web-only analytics instrumentation
