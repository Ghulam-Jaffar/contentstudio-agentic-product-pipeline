# Stories — Analytics UI Fixes (X duplicate button + Instagram Top Posts time)

One frontend-only, web-only story covering two independent display bugs in the Analytics module. Created under the current Miscellaneous epic. Structure in Shortcut using the **New Feature Template**.

---

## [FE] Fix two Analytics display bugs — duplicate "Auto sync schedule" on X and wrong time on Instagram Top Posts

### Description
As a ContentStudio user viewing my analytics, I want the analytics screens to display correctly so I can trust what I see and not be confused by repeated or wrong information. This story fixes two small, unrelated display bugs:

1. **X (Twitter):** the top-right toolbar shows the **"Auto sync schedule"** control **twice**. One is the intended control (it includes the gear icon to configure the sync schedule); the other is a redundant duplicate. This only happens on X — other platforms show their info control once, correctly.
2. **Instagram → Top Posts:** the time shown under each post is the **current (live) time**, not the time the post was actually created/published — which is misleading and makes the analytics look wrong.

---

### Workflow

**Part 1 — X (Twitter): single "Auto sync schedule" control**
1. User opens **Analytics** and selects an **X (Twitter)** account.
2. In the top-right of the analytics header, the user sees a **single** "Auto sync schedule" control — the one with the **gear icon** for configuring the sync schedule — alongside the **"Sync now"** manual-sync button.
3. The user no longer sees a second, duplicate "Auto sync schedule" label next to it.
4. On every other platform (Facebook, Instagram, LinkedIn, TikTok, Pinterest, YouTube, Google Business Profile) and on the Overview, the top-right platform-info control appears exactly as before — no change.

**Part 2 — Instagram: correct posted time on Top Posts**
1. User opens **Instagram Analytics** and views the **Top Posts** card on the overview.
2. Under each post's account name, the user sees the **date and time the post was published**, shown in the **workspace timezone**.
3. The displayed time reflects the original publish moment and does **not** change to "now" on reload.

---

### Acceptance criteria

**X (Twitter) — duplicate "Auto sync schedule" control**
- [ ] On the X (Twitter) analytics screen, the top-right toolbar shows the "Auto sync schedule" control **exactly once**.
- [ ] The control that remains is the one paired with the **gear icon** that opens the sync-configuration modal.
- [ ] The gear icon still opens the sync-configuration modal ("Configure sync") when clicked.
- [ ] The X-specific **"Sync now"** manual-sync button is still present and continues to trigger a manual sync.
- [ ] On all other platforms (Facebook, Instagram, LinkedIn, TikTok, Pinterest, YouTube, GMB) and the Overview, the top-right platform-info control still appears once and is unchanged.
- [ ] On shared / public analytics links, sync controls remain hidden (unchanged behavior).
- [ ] No visual regression in the top-right toolbar spacing/alignment on X after the duplicate is removed.

**Instagram — Top Posts time**
- [ ] Each Instagram **Top Posts** card shows the post's **original publish date/time** (in the workspace timezone), not the current time.
- [ ] Reloading or revisiting the page shows the **same** posted time for a given post — it never updates to the current time.
- [ ] The posted time shown on the Top Posts card **matches** the posted time shown for the same post elsewhere in Instagram analytics (e.g., the post preview / detail).
- [ ] When a post has **no available publish timestamp**, the card does **not** display the current time — the date/time line is **hidden** (shows nothing) instead of falling back to "now".
- [ ] The same corrected behavior applies in the **report/export view** of the Top Posts card.

---

### Mock-ups
N/A — no new UI. Part 1 removes a duplicate render of an existing control; Part 2 corrects the value shown in an existing date/time position (and hides it when unavailable).

---

### Impact on existing data
None. Both are display-only fixes — no data, schema, or API changes.

---

### Impact on other products
- **Web only.** Both the X analytics toolbar and the Instagram Top Posts card are web features — not present in the iOS/Android apps or the Chrome extension, so no mobile or extension stories are needed.
- **White-label:** no copy or theming changes; the X label reuses its existing i18n key and the Instagram date uses the shared date/time utility.

---

### Dependencies
None on the frontend. **Note (Instagram part):** the fix assumes the overview Top Posts data already includes a publish timestamp. If verification shows the overview `top_posts` response contains **no** publish-time field at all, a small backend follow-up is required to include it (flagged in Implementation references); until then, the card hides the date line rather than showing the current time.

---

### Global quality & compliance (wherever applicable)

- [ ] Mobile responsiveness (frontend only, N/A for backend-only stories)
- [ ] Multilingual support (frontend + backend, translations available or fallback handled)
- [ ] UI theming support (default + white-label, design library components are being used)
- [ ] White-label domains impact review
- [ ] Cross-product impact assessment (web, mobile apps, Chrome extension) — N/A for mobile/Chrome, web-only features

---

### Implementation references
*Pointers from research — not a contract. Engineering may choose a different approach.*

**Part 1 — X (Twitter) duplicate control**
- `contentstudio-frontend/src/modules/analytics/views/common/TabsComponent.vue` — the top-right toolbar renders `PlatformTooltip` **twice** for X:
  - the **dedicated X block** (`v-if="type === 'twitter' && !isSharedAnalytics"`, ~lines 40-63) — this is the one that also contains the **gear icon** that opens the sync-config modal; **keep this one**.
  - the **generic** `PlatformTooltip` (`v-if="!isSharedAnalytics"`, ~lines 143-148) — rendered for every platform; for X it *also* prints "Auto sync schedule", producing the duplicate. Excluding X here (e.g. add `&& type !== 'twitter'` to its `v-if`) removes the duplicate while leaving every other platform untouched.
  - `contentstudio-frontend/src/modules/analytics/components/common/PlatformTooltip.vue` — the "Auto sync schedule" label is rendered by this component only when `type === 'twitter'` (~lines 170-172) via i18n key `analytics.platform_tooltip.auto_sync_schedule`. No copy change. The dedicated block uses `:enable-peek="true"` — confirm the peek/chevron preview still renders as intended.

**Part 2 — Instagram Top Posts time**
- `contentstudio-frontend/src/modules/analytics/views/instagram_v2/components/TopPostCard.vue` (~line 78) — the date is rendered with `createDate(post.post_created_at, false, true).inWorkspaceTimezone().formatDateTime()`.
- **Root cause / gotcha:** `contentstudio-frontend/src/composables/useDateTime.ts` — `createDate` returns `dayjs()` (the **current time**) whenever its `input` is falsy (`return input ? … : dayjs()`). So an empty/missing `post_created_at` silently renders "now". The fix must **guard the empty case** and not call `createDate` unguarded on a possibly-empty value.
- **Data source:** `contentstudio-frontend/src/modules/analytics/views/instagram_v2/composables/useInstagramAnalytics.js` (~line 217, `OVERVIEW_TOP_POSTS` branch) — the overview Top Posts card is fed **raw, untransformed** data: `overviewTopPostsData.value = data?.top_posts`. Contrast the detailed `TOP_POSTS` branch (~line 240) which runs `transformObject` and maps `postCreatedAt`, `timestamp`, and `stored_event_at`. Confirm which timestamp key the **overview** `top_posts` payload actually carries and bind the card to it (candidates: `post_created_at`, `timestamp`, `stored_event_at`).
- **Suggested approach:** `post_created_at` is the canonical Instagram field used by `getPostDate` (`useAnalyticsUtils.js`) and `AnalyticPreview.vue`. If the overview payload uses a different key, prefer normalizing the timestamp once at the data layer (in `useInstagramAnalytics.js`) so all consumers stay consistent, rather than special-casing this one card. If the endpoint returns no publish timestamp at all, that's a backend follow-up; hide the date line in the meantime.

---

### Shortcut fields
- **Template:** New Feature Template
- **Story type:** bug
- **Project:** Web App
- **Group:** Frontend
- **Epic:** Q2 - 2026: Miscellaneous
- **Priority:** Medium
- **Product Area:** Analytics
- **Skill Set:** Frontend
- **Estimate:** _(empty — devs estimate during sprint planning)_
- **Labels:** _(none)_
- **Iteration:** _(PO assigns current/target sprint at creation)_
